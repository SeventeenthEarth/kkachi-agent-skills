package projectinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
)

const modeUninstallDryRun = "project_uninstall_dry_run"

type PlannedRemoval struct {
	Action         string  `json:"action"`
	Project        string  `json:"project"`
	InstalledSkill string  `json:"installed_skill"`
	SourcePackID   string  `json:"source_pack_id"`
	TargetPath     string  `json:"target_path"`
	ManifestSHA256 string  `json:"manifest_sha256"`
	CurrentSHA256  *string `json:"current_sha256,omitempty"`
	Bytes          int     `json:"bytes,omitempty"`
	Reason         string  `json:"reason"`
}

type SkippedLocalFile struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	SHA256 string `json:"sha256,omitempty"`
	Bytes  int    `json:"bytes,omitempty"`
}

type UninstallBackupRecovery struct {
	BackupRequired      bool     `json:"backup_required"`
	BackupDestination   string   `json:"backup_destination"`
	BackupWriteDeferred bool     `json:"backup_write_deferred"`
	ApplyDeferred       bool     `json:"apply_deferred"`
	BackupManifestPath  string   `json:"backup_manifest_path,omitempty"`
	BackupEvidencePath  string   `json:"backup_evidence_path,omitempty"`
	BackupSHA256        string   `json:"backup_sha256,omitempty"`
	BackupVerified      bool     `json:"backup_verified,omitempty"`
	Instructions        []string `json:"instructions"`
}

type UninstallChecksums struct {
	Manifest      string `json:"manifest"`
	PlannedRemove string `json:"planned_removals"`
	ChangedPaths  string `json:"changed_paths"`
	SkippedFiles  string `json:"skipped_files"`
}

type ProjectUninstallResult struct {
	OK                      bool                     `json:"ok"`
	Command                 string                   `json:"command"`
	Mode                    string                   `json:"mode"`
	CLIVersion              string                   `json:"cli_version"`
	DryRun                  bool                     `json:"dry_run"`
	NoWrite                 NoWriteEvidence          `json:"no_write"`
	TargetProfile           discovery.TargetProfile  `json:"target_profile"`
	Project                 Project                  `json:"project"`
	SourcePack              SourcePack               `json:"source_pack"`
	ManifestPath            string                   `json:"manifest_path"`
	ManifestSHA256          *string                  `json:"manifest_sha256,omitempty"`
	PlannedRemovals         []PlannedRemoval         `json:"planned_removals"`
	SkippedLocalFiles       []SkippedLocalFile       `json:"skipped_local_files"`
	ChangedPaths            []ChangedPath            `json:"changed_paths"`
	BackupRecovery          UninstallBackupRecovery  `json:"backup_recovery"`
	Checksums               UninstallChecksums       `json:"checksums"`
	PlanHash                string                   `json:"plan_hash"`
	FutureApplyCommand      string                   `json:"future_apply_command"`
	ApprovalRequest         ApprovalRequest          `json:"approval_request,omitempty"`
	Approval                ApprovalEvidence         `json:"approval,omitempty"`
	UninstallID             string                   `json:"uninstall_id,omitempty"`
	BackupPath              string                   `json:"backup_path,omitempty"`
	Diagnostics             []discovery.Diagnostic   `json:"diagnostics"`
	ProjectSuiteDiagnostics []ProjectSuiteDiagnostic `json:"project_suite_diagnostics"`
	NextAction              string                   `json:"next_action"`
}

func BuildProjectUninstallDryRun(opts ProjectSuiteOptions) ProjectUninstallResult {
	profileRoot := resolveProfileRoot(opts.Profile, opts.ProfileRoot)
	sourcePackID := opts.SourcePack
	if sourcePackID == "" {
		sourcePackID = VirtualSourcePackID
	}
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	result := ProjectUninstallResult{
		OK:                true,
		Command:           "uninstall",
		Mode:              modeUninstallDryRun,
		CLIVersion:        CLIVersionForKASPROJ004(),
		DryRun:            true,
		NoWrite:           NoWriteEvidence{Guaranteed: true},
		TargetProfile:     discovery.TargetProfile{Name: opts.Profile, Root: profileRoot, ManifestPath: manifestPath, ManifestState: manifestState(manifestPath)},
		Project:           Project{ID: opts.Project, TargetSuitePath: "skills/" + opts.Project},
		SourcePack:        SourcePack{ID: sourcePackID, Source: "manifest", ResolvedFrom: "profile_manifest", PackChecksums: map[string]string{}, FormalRegistry: "profile_manifest"},
		ManifestPath:      manifestPath,
		PlannedRemovals:   []PlannedRemoval{},
		SkippedLocalFiles: []SkippedLocalFile{},
		ChangedPaths:      []ChangedPath{},
		BackupRecovery: UninstallBackupRecovery{
			BackupRequired:      false,
			BackupDestination:   "approved Obsidian vault backup area (TOKEN-005 apply only)",
			BackupWriteDeferred: true,
			ApplyDeferred:       true,
			Instructions: []string{
				"TOKEN-004 reads only the manifest and filesystem.",
				"TOKEN-005 must write backup evidence before any manifest-tracked removal.",
				"Local-only and unmanifested files are skipped by default.",
			},
		},
		FutureApplyCommand: fmt.Sprintf("kkachi-agent-skills uninstall --profile %s --project %s --apply dry-run:<hash> --backup-vault-root <abs-path>", opts.Profile, opts.Project),
		NextAction:         "Review uninstall dry-run evidence; removal and backup writes are TOKEN-005 behavior.",
	}
	if opts.Profile == "" {
		addUninstallError(&result, "profile_required", "uninstall requires --profile <profile>.")
	}
	if opts.Project == "" || !validProjectID.MatchString(opts.Project) {
		addUninstallError(&result, "invalid_project_id", fmt.Sprintf("project id %q is not a safe project suite id", opts.Project))
	}
	if sourcePackID != VirtualSourcePackID {
		addUninstallError(&result, "unknown_source_pack", fmt.Sprintf("unsupported project source pack %q", sourcePackID))
		finalizeUninstall(&result)
		return result
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			addUninstallError(&result, "manifest_missing", "profile manifest is missing; uninstall cannot infer managed project-suite files.")
		} else {
			addUninstallError(&result, "manifest_unreadable", "profile manifest is unreadable: "+err.Error())
		}
		finalizeUninstall(&result)
		return result
	}
	manifestSHA := sha256Bytes(data)
	result.ManifestSHA256 = &manifestSHA
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		addUninstallError(&result, "manifest_parse_error", "cannot parse KAS manifest: "+err.Error())
		finalizeUninstall(&result)
		return result
	}
	if version, _ := manifest["version"].(string); version != ManifestVersion {
		addUninstallError(&result, "unsupported_manifest_version", fmt.Sprintf("unsupported KAS manifest version: %q", manifest["version"]))
		finalizeUninstall(&result)
		return result
	}
	if kind, _ := manifest["kind"].(string); kind != ProfileManifestKind {
		addUninstallError(&result, "unsupported_manifest_kind", fmt.Sprintf("unsupported KAS manifest kind: %q", manifest["kind"]))
		finalizeUninstall(&result)
		return result
	}
	suite, ok := manifestProjectSuite(manifest, opts.Project, sourcePackID)
	if !ok {
		finalizeUninstall(&result)
		return result
	}
	result.SourcePack.ID = sourcePackIDFromProjectSuite(suite)
	if result.SourcePack.ID == "" {
		result.SourcePack.ID = sourcePackID
	}
	rawSkills, ok := suite["installed_skills"].([]any)
	if !ok {
		addUninstallError(&result, "manifest_project_skills_invalid", "project suite installed_skills must be an array")
		finalizeUninstall(&result)
		return result
	}
	manifested := map[string]bool{}
	for _, raw := range rawSkills {
		skill, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		installed, _ := skill["installed_skill"].(string)
		target, _ := skill["target_path"].(string)
		checksum, _ := skill["checksum"].(string)
		manifested[target] = true
		if unsafeManifestProjectTarget(opts.Project, installed, target) {
			addUninstallSuiteDiag(&result, suiteDiag(opts.Project, installed, target, "error", "unsafe_manifest_target_path", "project suite manifest contains an unsafe target_path", "Repair the manifest before uninstall."))
			continue
		}
		if checksum == "" {
			addUninstallSuiteDiag(&result, suiteDiag(opts.Project, installed, target, "error", "checksum_evidence_missing", "manifest target lacks checksum evidence", "Repair manifest checksum evidence before uninstall."))
			continue
		}
		abs := filepath.Join(profileRoot, filepath.FromSlash(target))
		current, size, readErr := existingFileChecksumAndSize(abs)
		action := "remove"
		reason := "manifest_tracked_project_suite_file"
		if readErr != nil {
			if errors.Is(readErr, os.ErrNotExist) {
				action = "no_change"
				reason = "manifest_tracked_file_already_missing"
			} else {
				code := "target_read_failed"
				if strings.Contains(readErr.Error(), "symlink") {
					code = "target_symlink_rejected"
				}
				addUninstallSuiteDiag(&result, suiteDiag(opts.Project, installed, target, "error", code, "manifest target_path cannot be read safely: "+readErr.Error(), "Fix file permissions before uninstall."))
				action = "error"
			}
		} else if current != checksum {
			addUninstallSuiteDiag(&result, suiteDiag(opts.Project, installed, target, "error", "checksum_mismatch", "installed file checksum differs from manifest evidence", "Run doctor/repair or review local edits before uninstall."))
			action = "conflict"
			reason = "manifest_checksum_mismatch"
		}
		removal := PlannedRemoval{Action: action, Project: opts.Project, InstalledSkill: installed, SourcePackID: result.SourcePack.ID, TargetPath: target, ManifestSHA256: checksum, CurrentSHA256: stringPtrIfNotEmpty(current), Bytes: size, Reason: reason}
		result.PlannedRemovals = append(result.PlannedRemovals, removal)
		if action == "remove" {
			result.ChangedPaths = append(result.ChangedPaths, ChangedPath{Path: target, Action: "remove", InstalledSkill: installed, SourcePackID: result.SourcePack.ID, PreviousSHA256: &current, Bytes: size})
			result.BackupRecovery.BackupRequired = true
		}
	}
	if rawComponents, ok := suite["composition_files"].([]any); ok {
		for _, raw := range rawComponents {
			component, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			name, _ := component["name"].(string)
			target, _ := component["target_path"].(string)
			checksum, _ := component["checksum"].(string)
			manifested[target] = true
			if discovery.IsInvalidRelativePath(target) || !strings.HasPrefix(target, "skills/"+opts.Project+"/") {
				addUninstallSuiteDiag(&result, suiteDiag(opts.Project, name, target, "error", "unsafe_manifest_target_path", "project suite manifest contains an unsafe composition target_path", "Repair the manifest before uninstall."))
				continue
			}
			if checksum == "" {
				addUninstallSuiteDiag(&result, suiteDiag(opts.Project, name, target, "error", "checksum_evidence_missing", "manifest composition target lacks checksum evidence", "Repair manifest checksum evidence before uninstall."))
				continue
			}
			abs := filepath.Join(profileRoot, filepath.FromSlash(target))
			current, size, readErr := existingFileChecksumAndSize(abs)
			action := "remove"
			reason := "manifest_tracked_project_composition_file"
			if readErr != nil {
				if errors.Is(readErr, os.ErrNotExist) {
					action = "no_change"
					reason = "manifest_tracked_file_already_missing"
				} else {
					code := "target_read_failed"
					if strings.Contains(readErr.Error(), "symlink") {
						code = "target_symlink_rejected"
					}
					addUninstallSuiteDiag(&result, suiteDiag(opts.Project, name, target, "error", code, "manifest composition target_path cannot be read safely: "+readErr.Error(), "Fix file permissions before uninstall."))
					action = "error"
				}
			} else if current != checksum {
				addUninstallSuiteDiag(&result, suiteDiag(opts.Project, name, target, "error", "checksum_mismatch", "installed composition file checksum differs from manifest evidence", "Run doctor/repair or review local edits before uninstall."))
				action = "conflict"
				reason = "manifest_checksum_mismatch"
			}
			removal := PlannedRemoval{Action: action, Project: opts.Project, InstalledSkill: name, SourcePackID: result.SourcePack.ID, TargetPath: target, ManifestSHA256: checksum, CurrentSHA256: stringPtrIfNotEmpty(current), Bytes: size, Reason: reason}
			result.PlannedRemovals = append(result.PlannedRemovals, removal)
			if action == "remove" {
				result.ChangedPaths = append(result.ChangedPaths, ChangedPath{Path: target, Action: "remove", InstalledSkill: name, SourcePackID: result.SourcePack.ID, PreviousSHA256: &current, Bytes: size})
				result.BackupRecovery.BackupRequired = true
			}
		}
	}
	result.SkippedLocalFiles = scanSkippedProjectFiles(profileRoot, opts.Project, manifested)
	finalizeUninstall(&result)
	return result
}

func RenderHumanProjectUninstall(result ProjectUninstallResult) string {
	status := "ready"
	if !result.OK {
		status = "blocked"
	} else if !result.DryRun {
		status = "complete"
	}
	lines := []string{
		fmt.Sprintf("Status: uninstall %s - profile %s / project %s.", status, result.TargetProfile.Name, result.Project.ID),
		fmt.Sprintf("Manifest: %s (%s)", result.TargetProfile.ManifestState, result.ManifestPath),
		fmt.Sprintf("Planned removals: %d; skipped local-only/unmanifested files: %d.", len(result.PlannedRemovals), len(result.SkippedLocalFiles)),
	}
	if result.DryRun {
		lines = append(lines,
			"Writes: dry-run only; profile/manifest/KAH/KAB/auth/provider/runtime/profile activation writes 0.",
			"Approval evidence: "+result.ApprovalRequest.EvidenceRef,
			"Backup: deferred to TOKEN-005 apply path; destination posture: "+result.BackupRecovery.BackupDestination,
		)
	} else {
		lines = append(lines,
			"Approval evidence: "+result.Approval.EvidenceRef,
			"Backup: "+result.BackupPath,
			fmt.Sprintf("backup_verified:%t backup_sha256:%s", result.BackupRecovery.BackupVerified, result.BackupRecovery.BackupSHA256),
		)
	}
	for _, removal := range result.PlannedRemovals {
		lines = append(lines, fmt.Sprintf("Plan: %s %s (%s)", removal.Action, removal.TargetPath, removal.Reason))
	}
	for _, skipped := range result.SkippedLocalFiles {
		lines = append(lines, fmt.Sprintf("Skip: %s - %s", skipped.Path, skipped.Reason))
	}
	for _, diagnostic := range result.ProjectSuiteDiagnostics {
		lines = append(lines, humanProjectSuiteDiagnosticLine(diagnostic))
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, "Diagnostic: "+diagnostic.Code+" - "+diagnostic.Message)
	}
	lines = append(lines, "Future apply: "+result.FutureApplyCommand)
	lines = append(lines, "Next: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func ApplyProjectUninstall(opts ProjectSuiteOptions, evidenceRef string, backupVaultRoot string) (ProjectUninstallResult, error) {
	dryRun := BuildProjectUninstallDryRun(opts)
	approvedHash, evidenceOK := approvedHashFromEvidence(evidenceRef)
	if !evidenceOK {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "approval_evidence_malformed", "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>."), nil
	}
	if approvedHash != dryRun.PlanHash {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "approval_plan_hash_mismatch", "approval evidence does not match the current dry-run plan hash; no files were written."), nil
	}
	if !dryRun.OK {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "project_uninstall_plan_not_approvable", "current dry-run plan contains conflict or error entries; no files were written."), nil
	}
	removals := plannedManifestRemovals(dryRun)
	if len(removals) == 0 {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "project_uninstall_no_manifest_tracked_removals", "current dry-run plan has no manifest-tracked files to remove; no files were written."), nil
	}
	uninstallID := makeUninstallID(dryRun.PlanHash)
	backupRoot, err := validateBackupVaultRoot(backupVaultRoot, dryRun.TargetProfile.Root, dryRun.TargetProfile.Name, dryRun.Project.ID, uninstallID)
	if err != nil {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "backup_vault_root_rejected", err.Error()), nil
	}
	if err := preflightApprovedUninstall(dryRun, removals); err != nil {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "project_uninstall_preflight_failed", err.Error()), nil
	}
	backupManifestPath, backupEvidencePath, backupSHA, err := writeUninstallBackup(dryRun, removals, evidenceRef, approvedHash, uninstallID, backupRoot)
	if err != nil {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "backup_write_failed", err.Error()), nil
	}
	changed := []ChangedPath{}
	for _, removal := range removals {
		target, err := safeProfileWritePath(dryRun.TargetProfile.Root, removal.TargetPath)
		if err != nil {
			return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "unsafe_target_path", err.Error()), nil
		}
		if err := os.Remove(target); err != nil {
			return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "remove_failed", err.Error()), nil
		}
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "remove_verify_failed", "removed file is still present: "+removal.TargetPath), nil
		}
		changed = append(changed, ChangedPath{Path: removal.TargetPath, Action: "remove", InstalledSkill: removal.InstalledSkill, SourcePackID: removal.SourcePackID, PreviousSHA256: removal.CurrentSHA256, Bytes: removal.Bytes})
	}
	manifest, err := buildUninstalledProjectManifest(dryRun, evidenceRef, approvedHash, uninstallID, backupRoot, backupManifestPath, backupEvidencePath, backupSHA)
	if err != nil {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "manifest_build_failed", err.Error()), nil
	}
	manifestTarget, err := safeProfileWritePath(dryRun.TargetProfile.Root, ".kas/skill-pack-manifest.json")
	if err != nil {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "manifest_write_failed", err.Error()), nil
	}
	if err := writeJSONFile(manifestTarget, manifest); err != nil {
		return uninstallApprovalFailure(dryRun, evidenceRef, approvedHash, "manifest_write_failed", err.Error()), nil
	}
	changed = append(changed, ChangedPath{Path: ".kas/skill-pack-manifest.json", Action: "manifest_update", PreviousSHA256: dryRun.ManifestSHA256})
	result := dryRun
	result.OK = true
	result.Mode = "project_uninstall_approved"
	result.DryRun = false
	result.NoWrite = NoWriteEvidence{Guaranteed: false, ProfileWriteCount: len(removals), SkillWriteCount: len(removals), ManifestWriteCount: 1, KASDirectoryWriteCount: 1}
	result.ChangedPaths = changed
	result.Approval = ApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: true}
	result.UninstallID = uninstallID
	result.BackupPath = backupRoot
	result.BackupRecovery.BackupDestination = backupRoot
	result.BackupRecovery.BackupWriteDeferred = false
	result.BackupRecovery.ApplyDeferred = false
	result.BackupRecovery.BackupManifestPath = backupManifestPath
	result.BackupRecovery.BackupEvidencePath = backupEvidencePath
	result.BackupRecovery.BackupSHA256 = backupSHA
	result.BackupRecovery.BackupVerified = true
	result.Diagnostics = nil
	result.NextAction = "Uninstall complete. Review backup evidence before deleting vault recovery files."
	return result, nil
}

func plannedManifestRemovals(result ProjectUninstallResult) []PlannedRemoval {
	removals := []PlannedRemoval{}
	for _, removal := range result.PlannedRemovals {
		if removal.Action == "remove" {
			removals = append(removals, removal)
		}
	}
	return removals
}

func uninstallApprovalFailure(dryRun ProjectUninstallResult, evidenceRef string, approvedHash string, code string, message string) ProjectUninstallResult {
	result := dryRun
	result.OK = false
	result.Mode = "project_uninstall_approved"
	result.DryRun = false
	result.NoWrite = NoWriteEvidence{Guaranteed: true}
	result.Approval = ApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: approvedHash == dryRun.PlanHash}
	result.Diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: code, Message: message}}, dryRun.Diagnostics...)
	result.NextAction = "No files were removed. Fix the reported issue, rerun dry-run, and apply only the matching plan hash."
	return result
}

func makeUninstallID(planHash string) string {
	short := strings.TrimPrefix(planHash, "sha256:")
	if len(short) > 12 {
		short = short[:12]
	}
	return "kas-project-uninstall-" + short
}

func validateBackupVaultRoot(root string, profileRoot string, profile string, project string, uninstallID string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("uninstall --apply requires --backup-vault-root <abs-path>")
	}
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("--backup-vault-root must be an absolute path")
	}
	cleanRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("--backup-vault-root must resolve without symlink errors: %w", err)
	}
	info, err := os.Stat(cleanRoot)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("--backup-vault-root must be an existing directory")
	}
	cleanProfile, err := filepath.EvalSymlinks(profileRoot)
	if err != nil {
		cleanProfile, _ = filepath.Abs(profileRoot)
	}
	if sameOrInside(cleanRoot, cleanProfile) {
		return "", fmt.Errorf("--backup-vault-root must not be inside the target profile")
	}
	probe, err := os.CreateTemp(cleanRoot, ".kas-vault-probe-*")
	if err != nil {
		return "", fmt.Errorf("--backup-vault-root is not writable/verifiable: %w", err)
	}
	probePath := probe.Name()
	if _, err := probe.Write([]byte("kas backup probe\n")); err != nil {
		_ = probe.Close()
		_ = os.Remove(probePath)
		return "", fmt.Errorf("--backup-vault-root is not writable/verifiable: %w", err)
	}
	if err := probe.Close(); err != nil {
		_ = os.Remove(probePath)
		return "", fmt.Errorf("--backup-vault-root is not writable/verifiable: %w", err)
	}
	_ = os.Remove(probePath)
	return filepath.Join(cleanRoot, "kas-backups", safePathComponent(profile), safePathComponent(project), uninstallID), nil
}

func sameOrInside(path string, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && (rel == "." || (!strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel)))
}

func safePathComponent(value string) string {
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", " ", "-")
	return replacer.Replace(value)
}

func preflightApprovedUninstall(dryRun ProjectUninstallResult, removals []PlannedRemoval) error {
	seen := map[string]bool{}
	for _, removal := range removals {
		if seen[removal.TargetPath] {
			return fmt.Errorf("duplicate removal target: %s", removal.TargetPath)
		}
		seen[removal.TargetPath] = true
		if removal.Reason == "manifest_tracked_project_composition_file" {
			if discovery.IsInvalidRelativePath(removal.TargetPath) || !strings.HasPrefix(removal.TargetPath, "skills/"+dryRun.Project.ID+"/") {
				return fmt.Errorf("unsafe manifest removal target: %s", removal.TargetPath)
			}
		} else if unsafeManifestProjectTarget(dryRun.Project.ID, removal.InstalledSkill, removal.TargetPath) {
			return fmt.Errorf("unsafe manifest removal target: %s", removal.TargetPath)
		}
		target, err := safeProfileWritePath(dryRun.TargetProfile.Root, removal.TargetPath)
		if err != nil {
			return err
		}
		actual, readErr := existingFileSHA(target)
		if readErr != nil {
			return fmt.Errorf("%s: %s", readErr.Code, readErr.Message)
		}
		if actual == nil {
			return fmt.Errorf("target disappeared after dry-run: %s", removal.TargetPath)
		}
		if removal.CurrentSHA256 == nil || *actual != *removal.CurrentSHA256 || *actual != removal.ManifestSHA256 {
			return fmt.Errorf("target checksum changed after dry-run: %s", removal.TargetPath)
		}
	}
	return nil
}

func writeUninstallBackup(dryRun ProjectUninstallResult, removals []PlannedRemoval, evidenceRef string, approvedHash string, uninstallID string, backupRoot string) (string, string, string, error) {
	for _, removal := range removals {
		source, err := safeProfilePath(dryRun.TargetProfile.Root, removal.TargetPath)
		if err != nil {
			return "", "", "", err
		}
		dest := filepath.Join(backupRoot, "files", filepath.FromSlash(removal.TargetPath))
		if err := copyAbsoluteFile(source, dest); err != nil {
			return "", "", "", err
		}
		sum, err := checksumFile(dest)
		if err != nil {
			return "", "", "", err
		}
		if removal.CurrentSHA256 == nil || sum != *removal.CurrentSHA256 {
			return "", "", "", fmt.Errorf("backup checksum verification failed for %s", removal.TargetPath)
		}
	}
	manifestCopy := filepath.Join(backupRoot, "skill-pack-manifest.json.before")
	if err := copyAbsoluteFile(dryRun.ManifestPath, manifestCopy); err != nil {
		return "", "", "", err
	}
	evidencePath := filepath.Join(backupRoot, "uninstall-evidence.json")
	evidence := map[string]any{
		"uninstall_id":             uninstallID,
		"approval_evidence_ref":    evidenceRef,
		"dry_run_plan_hash":        dryRun.PlanHash,
		"approved_plan_hash":       approvedHash,
		"matched_current_plan":     approvedHash == dryRun.PlanHash,
		"profile":                  dryRun.TargetProfile,
		"project":                  dryRun.Project,
		"source_pack":              dryRun.SourcePack,
		"manifest_sha256":          dryRun.ManifestSHA256,
		"planned_removals":         removals,
		"skipped_local_files":      dryRun.SkippedLocalFiles,
		"backup_manifest_path":     manifestCopy,
		"backup_files_root":        filepath.Join(backupRoot, "files"),
		"forbidden_non_manifest":   "local-only and unmanifested files are preserved",
		"kah_kab_runtime_mutation": false,
	}
	if err := writeJSONFile(evidencePath, evidence); err != nil {
		return "", "", "", err
	}
	sum, err := checksumFile(evidencePath)
	if err != nil {
		return "", "", "", err
	}
	return manifestCopy, evidencePath, sum, nil
}

func buildUninstalledProjectManifest(dryRun ProjectUninstallResult, evidenceRef string, approvedHash string, uninstallID string, backupRoot string, backupManifestPath string, backupEvidencePath string, backupSHA string) (map[string]any, error) {
	data, err := os.ReadFile(dryRun.ManifestPath)
	if err != nil {
		return nil, err
	}
	var existing map[string]any
	if err := json.Unmarshal(data, &existing); err != nil {
		return nil, err
	}
	installs := []any{}
	if raw, ok := existing["installs"].([]any); ok {
		installs = raw
	}
	projectSuites := []any{}
	if rawSuites, ok := existing["project_suites"].([]any); ok {
		for _, raw := range rawSuites {
			entry, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if entry["project"] == dryRun.Project.ID && sourcePackIDFromProjectSuite(entry) == dryRun.SourcePack.ID {
				continue
			}
			projectSuites = append(projectSuites, entry)
		}
	}
	sort.Slice(projectSuites, func(i, j int) bool {
		left := projectSuites[i].(map[string]any)
		right := projectSuites[j].(map[string]any)
		return fmt.Sprint(left["project"], sourcePackIDFromProjectSuite(left)) < fmt.Sprint(right["project"], sourcePackIDFromProjectSuite(right))
	})
	profile := existing["profile"]
	if profile == nil {
		profile = map[string]any{"name": dryRun.TargetProfile.Name, "root": dryRun.TargetProfile.Root}
	}
	sourceRepo := existing["source_repo"]
	return map[string]any{
		"version":        ManifestVersion,
		"kind":           ProfileManifestKind,
		"profile":        profile,
		"source_repo":    sourceRepo,
		"installs":       installs,
		"project_suites": projectSuites,
		"last_uninstall": map[string]any{
			"uninstall_id":          uninstallID,
			"approval_evidence_ref": evidenceRef,
			"dry_run_plan_hash":     dryRun.PlanHash,
			"approved_plan_hash":    approvedHash,
			"project":               dryRun.Project.ID,
			"source_pack_id":        dryRun.SourcePack.ID,
			"backup_path":           backupRoot,
			"backup_manifest_path":  backupManifestPath,
			"backup_evidence_path":  backupEvidencePath,
			"backup_sha256":         backupSHA,
		},
	}, nil
}

func manifestProjectSuite(manifest map[string]any, project string, sourcePack string) (map[string]any, bool) {
	rawSuites, ok := manifest["project_suites"].([]any)
	if !ok {
		return nil, false
	}
	matches := []map[string]any{}
	for _, raw := range rawSuites {
		suite, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if suite["project"] == project && sourcePackIDFromProjectSuite(suite) == sourcePack {
			matches = append(matches, suite)
		}
	}
	if len(matches) != 1 {
		return nil, false
	}
	return matches[0], true
}

func scanSkippedProjectFiles(profileRoot string, project string, manifested map[string]bool) []SkippedLocalFile {
	root := filepath.Join(profileRoot, "skills", project)
	files := []SkippedLocalFile{}
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(profileRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if manifested[rel] {
			return nil
		}
		checksum, size, err := existingFileChecksumAndSize(path)
		skipped := SkippedLocalFile{Path: rel, Reason: "local_only_or_unmanifested_project_file"}
		if err == nil {
			skipped.SHA256 = checksum
			skipped.Bytes = size
		}
		files = append(files, skipped)
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

func existingFileChecksumAndSize(path string) (string, int, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", 0, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", 0, fmt.Errorf("target path is a symlink: %s", path)
	}
	if !info.Mode().IsRegular() {
		return "", 0, fmt.Errorf("target path is not a regular file: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", 0, err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), len(data), nil
}

func stringPtrIfNotEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func addUninstallSuiteDiag(result *ProjectUninstallResult, diag ProjectSuiteDiagnostic) {
	result.ProjectSuiteDiagnostics = append(result.ProjectSuiteDiagnostics, diag)
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: diag.Severity, Code: diag.Condition, Message: diag.Message})
}

func addUninstallError(result *ProjectUninstallResult, code string, message string) {
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: code, Message: message})
}

func finalizeUninstall(result *ProjectUninstallResult) {
	if len(result.PlannedRemovals) == 0 && noErrorDiagnostics(result.Diagnostics) {
		result.NextAction = "No manifest-tracked project-suite files are available for uninstall; no files were written."
	}
	sort.Slice(result.PlannedRemovals, func(i, j int) bool {
		return result.PlannedRemovals[i].TargetPath < result.PlannedRemovals[j].TargetPath
	})
	sort.Slice(result.ChangedPaths, func(i, j int) bool { return result.ChangedPaths[i].Path < result.ChangedPaths[j].Path })
	result.Checksums = UninstallChecksums{
		Manifest:      derefString(result.ManifestSHA256),
		PlannedRemove: checksumAny(result.PlannedRemovals),
		ChangedPaths:  checksumAny(result.ChangedPaths),
		SkippedFiles:  checksumAny(result.SkippedLocalFiles),
	}
	canonical := map[string]any{"command": result.Command, "mode": result.Mode, "dry_run": result.DryRun, "no_write": result.NoWrite, "target_profile": result.TargetProfile, "project": result.Project, "source_pack": result.SourcePack, "manifest_sha256": result.ManifestSHA256, "planned_removals": result.PlannedRemovals, "skipped_local_files": result.SkippedLocalFiles, "changed_paths": result.ChangedPaths, "backup_recovery": result.BackupRecovery, "checksums": result.Checksums, "diagnostics": result.Diagnostics}
	result.PlanHash = checksumAny(canonical)
	result.FutureApplyCommand = fmt.Sprintf("kkachi-agent-skills uninstall --profile %s --project %s --apply dry-run:%s --backup-vault-root <abs-path>", result.TargetProfile.Name, result.Project.ID, result.PlanHash)
	result.OK = noErrorDiagnostics(result.Diagnostics)
	result.ApprovalRequest = ApprovalRequest{Required: result.OK && len(result.ChangedPaths) > 0, EvidenceRef: "dry-run:" + result.PlanHash, DryRunPlanHash: result.PlanHash, HashIncludesProfile: true, HashIncludesManifestState: true, HashIncludesSourceSuite: true, HashIncludesNoWriteEvidence: true, HashIncludesBackupPlan: true, HashIncludesConflictsAndDiags: true}
	if !result.OK {
		result.NextAction = "Resolve uninstall dry-run diagnostics before TOKEN-005 removal; no files were written."
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
