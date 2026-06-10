package projectinstall

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
)

func ApplyApprovedInstall(repo string, opts Options, evidenceRef string) (Result, error) {
	opts.DryRun = true
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		return Result{}, err
	}
	approvedHash, evidenceOK := approvedHashFromEvidence(evidenceRef)
	if !evidenceOK {
		result := dryRun
		result.OK = false
		result.Mode = "project_approved_copy"
		result.DryRun = false
		result.Approval = ApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: false}
		result.Diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: "approval_evidence_malformed", Message: "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>."}}, dryRun.Diagnostics...)
		result.NextAction = "No files were written. Use the exact approval_request.evidence_ref value from the matching dry-run JSON."
		return result, nil
	}
	if approvedHash != dryRun.PlanHash {
		result := dryRun
		result.OK = false
		result.Mode = "project_approved_copy"
		result.DryRun = false
		result.Approval = ApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: false}
		result.Diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: "approval_plan_hash_mismatch", Message: "approval evidence does not match the current dry-run plan hash; no files were written."}}, dryRun.Diagnostics...)
		result.NextAction = "No files were written. Re-run dry-run and approve only the current plan hash."
		return result, nil
	}
	if !dryRun.OK {
		result := dryRun
		result.OK = false
		result.Mode = "project_approved_copy"
		result.DryRun = false
		result.Approval = ApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: true}
		result.Diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: "project_install_plan_not_approvable", Message: "current dry-run plan contains conflict or error entries; no files were written."}}, dryRun.Diagnostics...)
		result.NextAction = "No files were written. Resolve conflicts/errors and approve a fresh dry-run."
		return result, nil
	}
	installID := makeInstallID(dryRun.PlanHash, time.Now().UTC())
	backupRoot := filepath.Join(dryRun.TargetProfile.Root, ".kas", "backups", installID)
	relativeBackupRoot := filepath.ToSlash(filepath.Join(".kas", "backups", installID))
	if err := preflightApprovedInstall(dryRun, relativeBackupRoot); err != nil {
		return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, nil, "project_install_preflight_failed", err.Error()), nil
	}
	actualChanged := []ChangedPath{}
	backupByPath := map[string]string{}

	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "update" {
			continue
		}
		backupRel := filepath.ToSlash(filepath.Join(relativeBackupRoot, filepath.FromSlash(entry.Path)))
		if err := copyProfileFile(dryRun.TargetProfile.Root, entry.Path, backupRel); err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "backup_failed", err.Error()), nil
		}
		backup := entry
		backup.Action = "backup"
		backup.NewSHA256 = ""
		backup.BackupPath = backupRel
		actualChanged = append(actualChanged, backup)
		backupByPath[entry.Path] = backupRel
	}

	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "create" && entry.Action != "update" {
			continue
		}
		content, err := renderedContentForEntry(dryRun, entry)
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "source_render_failed", err.Error()), nil
		}
		target, err := safeProfileWritePath(dryRun.TargetProfile.Root, entry.Path)
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "unsafe_target_path", err.Error()), nil
		}
		if err := writeFileAtomic(target, content, 0o644); err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "skill_write_failed", err.Error()), nil
		}
		sum, err := checksumFile(target)
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "checksum_verify_failed", err.Error()), nil
		}
		if sum != entry.NewSHA256 {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "checksum_mismatch_after_write", "written file checksum does not match approved plan: "+entry.Path), nil
		}
		if backupRel := backupByPath[entry.Path]; backupRel != "" {
			entry.BackupPath = backupRel
		}
		actualChanged = append(actualChanged, entry)
	}

	previousManifestPath := ""
	if dryRun.TargetProfile.PreviousManifestSHA256 != nil {
		previousManifestPath = filepath.Join(backupRoot, "skill-pack-manifest.json.previous")
		previousRel, err := filepath.Rel(dryRun.TargetProfile.Root, previousManifestPath)
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "previous_manifest_backup_failed", err.Error()), nil
		}
		previousTarget, err := safeProfileWritePath(dryRun.TargetProfile.Root, filepath.ToSlash(previousRel))
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "previous_manifest_backup_failed", err.Error()), nil
		}
		if err := copyAbsoluteFile(dryRun.TargetProfile.ManifestPath, previousTarget); err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "previous_manifest_backup_failed", err.Error()), nil
		}
	}
	manifest, err := buildUpdatedProjectManifest(dryRun, evidenceRef, approvedHash, installID, backupRoot, previousManifestPath, actualChanged)
	if err != nil {
		return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "manifest_build_failed", err.Error()), nil
	}
	manifestTarget, err := safeProfileWritePath(dryRun.TargetProfile.Root, ".kas/skill-pack-manifest.json")
	if err != nil {
		return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "manifest_write_failed", err.Error()), nil
	}
	if err := writeJSONFile(manifestTarget, manifest); err != nil {
		return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "manifest_write_failed", err.Error()), nil
	}
	actualChanged = append(actualChanged, ChangedPath{Path: ".kas/skill-pack-manifest.json", Action: "manifest_update", PreviousSHA256: dryRun.TargetProfile.PreviousManifestSHA256})

	result := buildApprovedResult(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, true, nil)
	result.NextAction = "Install complete. Verify with doctor --project-suite; do not claim operational rollout from KASPROJ-003."
	return result, nil
}

func RenderHumanApproved(result Result) string {
	state := "complete"
	if !result.OK {
		state = "blocked"
	}
	lines := []string{
		fmt.Sprintf("Status: project KAS approved install %s - profile %s / project %s.", state, result.TargetProfile.Name, result.Project.ID),
		"Approval evidence: " + result.Approval.EvidenceRef,
		"install_id: " + result.InstallID,
		"manifest: " + result.ManifestPath,
		fmt.Sprintf("Changes: create %d, update %d, backup %d, manifest_update %d, conflict %d, error %d.",
			result.Summary.CountsByAction["create"],
			result.Summary.CountsByAction["update"],
			result.Summary.CountsByAction["backup"],
			result.Summary.CountsByAction["manifest_update"],
			result.Summary.CountsByAction["conflict"],
			result.Summary.CountsByAction["error"],
		),
		"Recovery: " + result.BackupPath,
		"semantic_adaptation_claimed:false; drift_policy: manual_review_required.",
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, "Diagnostic: "+diagnostic.Message)
	}
	lines = append(lines, "Next: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func preflightApprovedInstall(dryRun Result, relativeBackupRoot string) error {
	seenSkills := map[string]bool{}
	seenPaths := map[string]bool{}
	for _, skill := range dryRun.PlannedSkills {
		if seenSkills[skill.InstalledSkill] {
			return fmt.Errorf("duplicate installed skill: %s", skill.InstalledSkill)
		}
		if seenPaths[skill.TargetPath] {
			return fmt.Errorf("duplicate target path: %s", skill.TargetPath)
		}
		seenSkills[skill.InstalledSkill] = true
		seenPaths[skill.TargetPath] = true
		if unsafeTargetPath(dryRun.Project.ID, skill.InstalledSkill, skill.TargetPath) {
			return fmt.Errorf("unsafe target path: %s", skill.TargetPath)
		}
		if _, err := safeProfileWritePath(dryRun.TargetProfile.Root, skill.TargetPath); err != nil {
			return fmt.Errorf("skill target preflight failed for %s: %w", skill.TargetPath, err)
		}
	}
	if _, err := safeProfileWritePath(dryRun.TargetProfile.Root, ".kas/skill-pack-manifest.json"); err != nil {
		return fmt.Errorf("manifest write preflight failed for .kas/skill-pack-manifest.json: %w", err)
	}
	if dryRun.TargetProfile.PreviousManifestSHA256 != nil {
		previousRel := filepath.ToSlash(filepath.Join(relativeBackupRoot, "skill-pack-manifest.json.previous"))
		if _, err := safeProfileWritePath(dryRun.TargetProfile.Root, previousRel); err != nil {
			return fmt.Errorf("previous manifest backup preflight failed for %s: %w", previousRel, err)
		}
	}
	needsBackupRoot := false
	for _, entry := range dryRun.ChangedPaths {
		if entry.Action == "update" {
			needsBackupRoot = true
			break
		}
	}
	if needsBackupRoot {
		if _, err := safeProfileWritePath(dryRun.TargetProfile.Root, relativeBackupRoot); err != nil {
			return fmt.Errorf("backup root preflight failed for %s: %w", relativeBackupRoot, err)
		}
	}
	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "create" && entry.Action != "update" && entry.Action != "skip" {
			return fmt.Errorf("plan contains non-writable action: %s %s", entry.Action, entry.Path)
		}
		target, err := safeProfileWritePath(dryRun.TargetProfile.Root, entry.Path)
		if err != nil {
			return err
		}
		actual, readErr := existingFileSHA(target)
		if readErr != nil {
			return fmt.Errorf("%s: %s", readErr.Code, readErr.Message)
		}
		if entry.Action == "update" {
			backupRel := filepath.ToSlash(filepath.Join(relativeBackupRoot, filepath.FromSlash(entry.Path)))
			if _, err := safeProfileWritePath(dryRun.TargetProfile.Root, backupRel); err != nil {
				return fmt.Errorf("backup target preflight failed for %s: %w", backupRel, err)
			}
		}
		switch entry.Action {
		case "create":
			if actual != nil {
				return fmt.Errorf("target appeared after dry-run: %s", entry.Path)
			}
		case "update", "skip":
			if actual == nil {
				return fmt.Errorf("target disappeared after dry-run: %s", entry.Path)
			}
			if entry.PreviousSHA256 == nil || *actual != *entry.PreviousSHA256 {
				return fmt.Errorf("target checksum changed after dry-run: %s", entry.Path)
			}
		}
		content, err := renderedContentForEntry(dryRun, entry)
		if err != nil {
			return err
		}
		if sha256Bytes(content) != entry.NewSHA256 {
			return fmt.Errorf("source checksum changed after dry-run: %s", entry.Path)
		}
	}
	return nil
}

func renderedContentForEntry(result Result, entry ChangedPath) ([]byte, error) {
	for _, skill := range result.PlannedSkills {
		if skill.TargetPath != entry.Path {
			continue
		}
		pack, ok := sourcePackForPlannedSkill(result.SourceRepo.Path, skill.SourcePackID)
		if !ok {
			return nil, fmt.Errorf("missing source pack for %s", skill.SourcePackID)
		}
		sourcePath := filepath.Join(result.SourceRepo.Path, filepath.FromSlash(pack.SourcePath), "SKILL.md")
		info, err := os.Lstat(sourcePath)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("source SKILL.md symlink rejected: %s", sourcePath)
		}
		return plannedSkillContent(result.SourceRepo.Path, pack, skill.SourceSkill, skill.InstalledSkill)
	}
	return nil, fmt.Errorf("no planned skill maps target path %s", entry.Path)
}

func sourcePackForPlannedSkill(sourceRepo string, packID string) (discovery.SourcePack, bool) {
	packs, _, err := discoverDefaultProjectSuite(sourceRepo)
	if err != nil {
		return discovery.SourcePack{}, false
	}
	for _, pack := range packs {
		if pack.PackID == packID {
			return pack, true
		}
	}
	return discovery.SourcePack{}, false
}

func approvedFailure(dryRun Result, evidenceRef string, approvedHash string, installID string, backupRoot string, changed []ChangedPath, code string, message string) Result {
	diagnostics := []discovery.Diagnostic{{Level: "error", Code: code, Message: message}}
	diagnostics = append(diagnostics, dryRun.Diagnostics...)
	result := buildApprovedResult(dryRun, evidenceRef, approvedHash, installID, backupRoot, changed, false, diagnostics)
	result.NextAction = "No files were written after the failed preflight/write step. Review file and manifest state, then start from a fresh dry-run."
	return result
}

func buildApprovedResult(dryRun Result, evidenceRef string, approvedHash string, installID string, backupRoot string, changed []ChangedPath, ok bool, diagnostics []discovery.Diagnostic) Result {
	sort.Slice(changed, func(i, j int) bool {
		if changed[i].Action == changed[j].Action {
			return changed[i].Path < changed[j].Path
		}
		return changed[i].Action < changed[j].Action
	})
	counts := map[string]int{"create": 0, "update": 0, "skip": 0, "conflict": 0, "error": 0, "backup": 0, "manifest_update": 0}
	for _, entry := range changed {
		counts[entry.Action]++
	}
	if diagnostics == nil {
		diagnostics = dryRun.Diagnostics
	}
	return Result{
		OK:               ok,
		Command:          "install-project-kas",
		Mode:             "project_approved_copy",
		CLIVersion:       dryRun.CLIVersion,
		DryRun:           false,
		NoWrite:          NoWriteEvidence{Guaranteed: false, ProfileWriteCount: counts["create"] + counts["update"], SkillWriteCount: counts["create"] + counts["update"], ManifestWriteCount: counts["manifest_update"], KASDirectoryWriteCount: counts["manifest_update"]},
		SourceRepo:       dryRun.SourceRepo,
		TargetProfile:    dryRun.TargetProfile,
		Project:          dryRun.Project,
		SourcePack:       dryRun.SourcePack,
		ProjectTailoring: dryRun.ProjectTailoring,
		Summary:          Summary{TotalSkills: dryRun.Summary.TotalSkills, TotalFiles: counts["create"] + counts["update"] + counts["manifest_update"], CountsByAction: counts, ConflictCount: counts["conflict"], DiagnosticCount: len(diagnostics)},
		PlannedManifest:  dryRun.PlannedManifest,
		PlannedSkills:    dryRun.PlannedSkills,
		ChangedPaths:     changed,
		BackupPlan:       dryRun.BackupPlan,
		Checksums:        dryRun.Checksums,
		PlanHash:         dryRun.PlanHash,
		ApprovalRequest:  dryRun.ApprovalRequest,
		Approval:         ApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: approvedHash == dryRun.PlanHash},
		InstallID:        installID,
		ManifestPath:     dryRun.TargetProfile.ManifestPath,
		BackupPath:       backupRoot,
		Recovery:         &Recovery{RollbackSupported: true, BackupPath: backupRoot, PreviousManifestSHA256: dryRun.TargetProfile.PreviousManifestSHA256, Instructions: []string{"Restore updated files from backup_path when present.", "Restore previous manifest snapshot when present.", "For create-only installs, remove created paths and manifest entry if rollback is needed."}},
		Diagnostics:      diagnostics,
	}
}

func makeInstallID(planHash string, now time.Time) string {
	short := strings.TrimPrefix(planHash, "sha256:")
	if len(short) > 12 {
		short = short[:12]
	}
	return "kas-project-install-" + now.UTC().Format("20060102T150405Z") + "-" + short
}

func copyProfileFile(profileRoot string, sourceRel string, backupRel string) error {
	source, err := safeProfilePath(profileRoot, sourceRel)
	if err != nil {
		return err
	}
	backup, err := safeProfileWritePath(profileRoot, backupRel)
	if err != nil {
		return err
	}
	return copyFile(source, backup)
}

func copyAbsoluteFile(source string, target string) error {
	if !filepath.IsAbs(source) || !filepath.IsAbs(target) {
		return fmt.Errorf("copyAbsoluteFile requires absolute paths")
	}
	return copyFile(source, target)
}

func copyFile(source string, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".kas-copy-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, target); err != nil {
		return err
	}
	removeTmp = false
	return nil
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data, 0o644)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".kas-write-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	removeTmp = false
	return nil
}

func safeProfilePath(profileRoot string, rel string) (string, error) {
	if discovery.IsInvalidRelativePath(filepath.ToSlash(rel)) {
		return "", fmt.Errorf("unsafe profile relative path: %s", rel)
	}
	root, err := filepath.Abs(profileRoot)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("profile path escapes root: %s", rel)
	}
	return target, nil
}

func safeProfileWritePath(profileRoot string, rel string) (string, error) {
	target, err := safeProfilePath(profileRoot, rel)
	if err != nil {
		return "", err
	}
	root, err := filepath.Abs(profileRoot)
	if err != nil {
		return "", err
	}
	if err := rejectProfileSymlinkComponents(root, target); err != nil {
		return "", err
	}
	return target, nil
}

func rejectProfileSymlinkComponents(root string, target string) error {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	current := root
	for _, component := range strings.Split(rel, string(os.PathSeparator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("profile write path contains symlink below root: %s", current)
		}
	}
	return nil
}

func existingFileSHA(path string) (*string, *pathReadError) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, &pathReadError{Code: "target_read_error", Message: err.Error()}
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, &pathReadError{Code: "target_symlink_rejected", Message: "target path is a symlink: " + path}
	}
	if !info.Mode().IsRegular() {
		return nil, &pathReadError{Code: "target_path_not_file", Message: "target path is not a file: " + path}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &pathReadError{Code: "target_read_error", Message: err.Error()}
	}
	sum := sha256Bytes(data)
	return &sum, nil
}

type pathReadError struct {
	Code    string
	Message string
}

func checksumFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return sha256Bytes(data), nil
}
