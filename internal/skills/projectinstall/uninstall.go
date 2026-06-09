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

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
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
		FutureApplyCommand: fmt.Sprintf("kkachi-hermes-skills uninstall --profile %s --project %s --apply dry-run:<hash>", opts.Profile, opts.Project),
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
				addUninstallSuiteDiag(&result, suiteDiag(opts.Project, installed, target, "error", "target_read_failed", "manifest target_path cannot be read: "+readErr.Error(), "Fix file permissions before uninstall."))
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
	result.SkippedLocalFiles = scanSkippedProjectFiles(profileRoot, opts.Project, manifested)
	finalizeUninstall(&result)
	return result
}

func RenderHumanProjectUninstall(result ProjectUninstallResult) string {
	status := "ready"
	if !result.OK {
		status = "blocked"
	}
	lines := []string{
		fmt.Sprintf("Status: uninstall dry-run %s - profile %s / project %s.", status, result.TargetProfile.Name, result.Project.ID),
		fmt.Sprintf("Manifest: %s (%s)", result.TargetProfile.ManifestState, result.ManifestPath),
		fmt.Sprintf("Planned removals: %d; skipped local-only/unmanifested files: %d.", len(result.PlannedRemovals), len(result.SkippedLocalFiles)),
		"Writes: dry-run only; profile/manifest/KAH/KAB/auth/provider/runtime/profile activation writes 0.",
		"Backup: deferred to TOKEN-005 apply path; destination posture: " + result.BackupRecovery.BackupDestination,
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
	result.FutureApplyCommand = fmt.Sprintf("kkachi-hermes-skills uninstall --profile %s --project %s --apply dry-run:%s", result.TargetProfile.Name, result.Project.ID, result.PlanHash)
	result.OK = noErrorDiagnostics(result.Diagnostics)
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
