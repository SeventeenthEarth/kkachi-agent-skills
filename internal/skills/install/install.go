package install

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
)

const (
	CLIVersion         = "0.1.0"
	ManifestVersion    = "0.1"
	ManifestKind       = "kas_profile_skill_manifest"
	KABBoundaryMessage = "KAB is not required for the minimum install dry-run; execution-runtime work remains KAB-gated."
)

type Options struct {
	Profile     string
	PackIDs     []string
	ProfileRoot string
}

type ChangedPath struct {
	Path           string  `json:"path"`
	Action         string  `json:"action"`
	PackID         string  `json:"pack_id,omitempty"`
	PreviousSHA256 *string `json:"previous_sha256"`
	NewSHA256      string  `json:"new_sha256,omitempty"`
	Bytes          *int    `json:"bytes,omitempty"`
	ErrorCode      string  `json:"error_code,omitempty"`
	ErrorMessage   string  `json:"error_message,omitempty"`
	BackupPath     string  `json:"backup_path,omitempty"`
}

type BackupEntry struct {
	Path           string `json:"path"`
	BackupPath     string `json:"backup_path"`
	PreviousSHA256 string `json:"previous_sha256"`
	Bytes          int    `json:"bytes"`
}

type Requested struct {
	PackIDs            []string         `json:"pack_ids"`
	CategoryExpansions map[string][]any `json:"category_expansions"`
}

type Summary struct {
	TotalPacks     int            `json:"total_packs"`
	TotalFiles     int            `json:"total_files"`
	CountsByAction map[string]int `json:"counts_by_action"`
	ConflictCount  int            `json:"conflict_count"`
}

type ApprovalRequest struct {
	Required       bool   `json:"required"`
	Summary        string `json:"summary"`
	EvidenceRef    string `json:"evidence_ref"`
	DryRunPlanHash string `json:"dry_run_plan_hash"`
}

type ApprovalEvidence struct {
	EvidenceRef        string `json:"evidence_ref"`
	DryRunPlanHash     string `json:"dry_run_plan_hash"`
	ApprovedPlanHash   string `json:"approved_plan_hash"`
	MatchedCurrentPlan bool   `json:"matched_current_plan"`
}

type Recovery struct {
	RollbackSupported      bool     `json:"rollback_supported"`
	BackupPath             string   `json:"backup_path"`
	PreviousManifestSHA256 *string  `json:"previous_manifest_sha256"`
	Instructions           []string `json:"instructions"`
}

type Result struct {
	OK              bool                    `json:"ok"`
	Command         string                  `json:"command"`
	Mode            string                  `json:"mode"`
	CLIVersion      string                  `json:"cli_version"`
	SourceRepo      discovery.SourceRepo    `json:"source_repo"`
	TargetProfile   discovery.TargetProfile `json:"target_profile"`
	Requested       Requested               `json:"requested"`
	Summary         Summary                 `json:"summary"`
	DryRunPlanHash  string                  `json:"dry_run_plan_hash"`
	CanonicalPlan   map[string]any          `json:"canonical_plan"`
	Packs           []map[string]any        `json:"packs"`
	ChangedPaths    []ChangedPath           `json:"changed_paths"`
	BackupPlan      []BackupEntry           `json:"backup_plan"`
	ApprovalRequest ApprovalRequest         `json:"approval_request"`
	Approval        ApprovalEvidence        `json:"approval,omitempty"`
	InstallID       string                  `json:"install_id,omitempty"`
	ManifestPath    string                  `json:"manifest_path,omitempty"`
	BackupPath      string                  `json:"backup_path,omitempty"`
	Recovery        *Recovery               `json:"recovery,omitempty"`
	Diagnostics     []discovery.Diagnostic  `json:"diagnostics"`
	NextAction      string                  `json:"next_action"`
}

type sourceFile struct {
	RelativePath string
	SHA256       string
	Bytes        int
	Mode         string
}

func BuildDryRun(repo string, opts Options) (Result, error) {
	sourceRepo, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return Result{}, err
	}
	sourcePacks, err := discovery.DiscoverSourcePacks(sourceRepo)
	if err != nil {
		return Result{}, err
	}
	allPacks := map[string]discovery.SourcePack{}
	for _, pack := range sourcePacks {
		allPacks[pack.PackID] = pack
	}

	root := resolveProfileRoot(opts.Profile, opts.ProfileRoot)
	manifestPath := filepath.Join(root, ".kas", "skill-pack-manifest.json")
	target := discovery.TargetProfile{
		Name:          opts.Profile,
		Root:          root,
		ManifestPath:  manifestPath,
		ManifestState: "manifest_missing",
	}
	sourceInfo := discovery.SourceRepoInfo(sourceRepo)
	diagnostics := []discovery.Diagnostic{{Level: "info", Code: "kab_not_required_for_minimum_dry_run", Message: KABBoundaryMessage}}
	changedPaths := []ChangedPath{}
	packPayloads := []map[string]any{}
	backupPlan := []BackupEntry{}
	requestedPackIDs := append([]string{}, opts.PackIDs...)

	if opts.ProfileRoot == "" {
		if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
			message := "unknown Hermes profile: " + opts.Profile
			diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: "unknown_profile", Message: message}}, diagnostics...)
			changedPaths = append(changedPaths, changedPath("error", ".", "", nil, "", nil, "unknown_profile", message))
			return buildResult(false, sourceInfo, target, requestedPackIDs, packPayloads, changedPaths, backupPlan, diagnostics), nil
		}
	}

	manifestEntries := map[string]map[string]any{}
	state, manifest, manifestBytes, manifestErr := loadManifest(manifestPath)
	switch state {
	case "present":
		target.ManifestState = "manifest_present"
		sum := sha256.Sum256(manifestBytes)
		previous := hex.EncodeToString(sum[:])
		target.PreviousManifestSHA256 = &previous
		manifestEntries = manifestInstalls(manifest)
	case "error":
		target.ManifestState = "manifest_unreadable"
		if manifestErr.SHA256 != "" {
			target.PreviousManifestSHA256 = &manifestErr.SHA256
		}
		diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: manifestErr.Code, Message: manifestErr.Message}}, diagnostics...)
		changedPaths = append(changedPaths, changedPath("error", ".kas/skill-pack-manifest.json", "", nil, "", nil, manifestErr.Code, manifestErr.Message))
		return buildResult(false, sourceInfo, target, requestedPackIDs, packPayloads, changedPaths, backupPlan, diagnostics), nil
	}

	for _, packID := range requestedPackIDs {
		pack, ok := allPacks[packID]
		if !ok {
			message := "unknown KAS pack id: " + packID
			diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: "unknown_pack_id", Message: message}}, diagnostics...)
			changedPaths = append(changedPaths, changedPath("error", "skills/"+packID, packID, nil, "", nil, "unknown_pack_id", message))
			continue
		}
		payload, paths, backups := planPack(sourceRepo, root, pack, manifestEntries[packID])
		packPayloads = append(packPayloads, payload)
		changedPaths = append(changedPaths, paths...)
		backupPlan = append(backupPlan, backups...)
		for _, entry := range paths {
			if entry.Action == "error" {
				code := entry.ErrorCode
				if code == "" {
					code = "install_plan_error"
				}
				message := entry.ErrorMessage
				if message == "" {
					message = "install dry-run failed"
				}
				diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: code, Message: message}}, diagnostics...)
			}
		}
	}

	ok := true
	for _, entry := range changedPaths {
		if entry.Action == "conflict" || entry.Action == "error" {
			ok = false
			break
		}
	}
	return buildResult(ok, sourceInfo, target, requestedPackIDs, packPayloads, changedPaths, backupPlan, diagnostics), nil
}

func RenderHumanDryRun(result Result) string {
	state := "완료"
	if !result.OK {
		state = "실패"
	}
	lines := []string{
		fmt.Sprintf("상태: install dry-run %s — 프로필 %s에는 아무것도 쓰지 않았습니다.", state, result.TargetProfile.Name),
		"대상: " + result.TargetProfile.Root,
		fmt.Sprintf(
			"변경 계획: create %d, update %d, skip %d, conflict %d, error %d.",
			result.Summary.CountsByAction["create"],
			result.Summary.CountsByAction["update"],
			result.Summary.CountsByAction["skip"],
			result.Summary.CountsByAction["conflict"],
			result.Summary.CountsByAction["error"],
		),
		"plan hash: " + result.DryRunPlanHash,
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, "진단: "+diagnostic.Message)
	}
	lines = append(lines, result.NextAction)
	return strings.Join(lines, "\n")
}

func ApplyApprovedInstall(repo string, opts Options, evidenceRef string) (Result, error) {
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		return Result{}, err
	}
	approvedHash, ok := approvedHashFromEvidence(evidenceRef)
	if !ok || approvedHash != dryRun.DryRunPlanHash {
		result := dryRun
		result.OK = false
		result.Mode = "approved_copy"
		result.Approval = ApprovalEvidence{
			EvidenceRef:        evidenceRef,
			DryRunPlanHash:     dryRun.DryRunPlanHash,
			ApprovedPlanHash:   approvedHash,
			MatchedCurrentPlan: false,
		}
		result.Diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: "approval_plan_hash_mismatch", Message: "approval evidence does not match the current dry-run plan hash; rerun dry-run and approve the current plan."}}, dryRun.Diagnostics...)
		result.NextAction = "설치하지 않았습니다. 현재 dry-run plan hash를 다시 확인한 뒤 --approve dry-run:<hash>로 재실행하세요."
		return result, nil
	}
	if !dryRun.OK {
		result := dryRun
		result.Mode = "approved_copy"
		result.Approval = ApprovalEvidence{
			EvidenceRef:        evidenceRef,
			DryRunPlanHash:     dryRun.DryRunPlanHash,
			ApprovedPlanHash:   approvedHash,
			MatchedCurrentPlan: true,
		}
		result.Diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: "install_plan_not_approvable", Message: "current dry-run plan contains conflict or error entries; approved install is closed."}}, dryRun.Diagnostics...)
		result.NextAction = "설치하지 않았습니다. conflict/error를 해결하고 새 dry-run을 승인하세요."
		return result, nil
	}

	installID := makeInstallID(dryRun.DryRunPlanHash, time.Now().UTC())
	backupRoot := filepath.Join(dryRun.TargetProfile.Root, ".kas", "backups", installID)
	relativeBackupRoot := filepath.ToSlash(filepath.Join(".kas", "backups", installID))
	actualChanged := []ChangedPath{}
	actualBackupByPath := map[string]string{}

	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "update" {
			continue
		}
		backupRel := filepath.ToSlash(filepath.Join(relativeBackupRoot, filepath.FromSlash(entry.Path)))
		if err := copyProfileFile(dryRun.TargetProfile.Root, entry.Path, backupRel); err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "backup_failed", err.Error()), nil
		}
		backup := changedPath("backup", entry.Path, entry.PackID, entry.PreviousSHA256, "", entry.Bytes, "", "")
		backup.BackupPath = backupRel
		actualChanged = append(actualChanged, backup)
		actualBackupByPath[entry.Path] = backupRel
	}

	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "create" && entry.Action != "update" {
			continue
		}
		sourceRel, ok := sourceRelativeForChangedPath(dryRun.Packs, entry)
		if !ok {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "source_mapping_failed", "could not map target path to source file: "+entry.Path), nil
		}
		sourcePath := filepath.Join(dryRun.SourceRepo.Path, filepath.FromSlash(sourceRel))
		targetPath, err := safeProfileWritePath(dryRun.TargetProfile.Root, entry.Path)
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "unsafe_target_path", err.Error()), nil
		}
		if err := copyFile(sourcePath, targetPath); err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "copy_failed", err.Error()), nil
		}
		sum, err := checksumFile(targetPath)
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "checksum_verify_failed", err.Error()), nil
		}
		if sum != entry.NewSHA256 {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "checksum_mismatch_after_copy", "copied file checksum does not match source checksum: "+entry.Path), nil
		}
		if backupRel := actualBackupByPath[entry.Path]; backupRel != "" {
			entry.BackupPath = backupRel
		}
		actualChanged = append(actualChanged, entry)
	}

	previousManifestPath := ""
	if dryRun.TargetProfile.PreviousManifestSHA256 != nil {
		previousManifestPath = filepath.Join(backupRoot, "skill-pack-manifest.json.previous")
		previousManifestRel, err := filepath.Rel(dryRun.TargetProfile.Root, previousManifestPath)
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "previous_manifest_backup_failed", err.Error()), nil
		}
		previousManifestTarget, err := safeProfileWritePath(dryRun.TargetProfile.Root, filepath.ToSlash(previousManifestRel))
		if err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "previous_manifest_backup_failed", err.Error()), nil
		}
		if err := copyAbsoluteFile(dryRun.TargetProfile.ManifestPath, previousManifestTarget); err != nil {
			return approvedFailure(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, "previous_manifest_backup_failed", err.Error()), nil
		}
	}
	manifest, err := buildUpdatedManifest(dryRun, evidenceRef, approvedHash, installID, backupRoot, previousManifestPath, actualChanged)
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
	manifestChanged := changedPath("manifest_update", ".kas/skill-pack-manifest.json", "", dryRun.TargetProfile.PreviousManifestSHA256, "", nil, "", "")
	actualChanged = append(actualChanged, manifestChanged)

	result := buildApprovedResult(dryRun, evidenceRef, approvedHash, installID, backupRoot, actualChanged, true, nil)
	result.NextAction = "설치 완료. 다음: kkachi-hermes-skills doctor --profile " + dryRun.TargetProfile.Name + " (CLIMVP-005에서 구현 예정)."
	return result, nil
}

func RenderHumanApproved(result Result) string {
	state := "완료"
	if !result.OK {
		state = "실패"
	}
	lines := []string{
		fmt.Sprintf("상태: approved copy install %s — 프로필 %s.", state, result.TargetProfile.Name),
		"대상: " + result.TargetProfile.Root,
		fmt.Sprintf(
			"변경: create %d, update %d, backup %d, manifest_update %d, conflict %d, error %d.",
			result.Summary.CountsByAction["create"],
			result.Summary.CountsByAction["update"],
			result.Summary.CountsByAction["backup"],
			result.Summary.CountsByAction["manifest_update"],
			result.Summary.CountsByAction["conflict"],
			result.Summary.CountsByAction["error"],
		),
		"manifest: " + result.ManifestPath,
		"복구: " + result.BackupPath,
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, "진단: "+diagnostic.Message)
	}
	lines = append(lines, result.NextAction)
	return strings.Join(lines, "\n")
}

func planPack(sourceRepo string, profileRoot string, pack discovery.SourcePack, manifestEntry map[string]any) (map[string]any, []ChangedPath, []BackupEntry) {
	sourcePackPath := filepath.Join(sourceRepo, filepath.FromSlash(pack.SourcePath))
	targetPackPath := filepath.ToSlash(filepath.Join("skills", filepath.FromSlash(pack.PackID)))
	changedPaths := []ChangedPath{}
	backupPlan := []BackupEntry{}
	filesPayload := []map[string]any{}
	payload := map[string]any{
		"pack_id":                 pack.PackID,
		"category":                pack.Category,
		"name":                    pack.Name,
		"source_path":             pack.SourcePath,
		"target_path":             targetPackPath,
		"pack_checksum":           pack.Checksum,
		"pack_checksum_algorithm": "sha256",
		"installed_state":         "not_installed",
		"files":                   filesPayload,
	}

	if errors := manifestPathErrors(manifestEntry, targetPackPath, pack.PackID); len(errors) > 0 {
		payload["installed_state"] = "error"
		return payload, errors, backupPlan
	}

	sourceFiles, sourceErr := sourceFiles(sourcePackPath)
	if sourceErr != nil {
		errorPath := filepath.ToSlash(filepath.Join(targetPackPath, sourceErr.RelativePath))
		changedPaths = append(changedPaths, changedPath("error", errorPath, pack.PackID, nil, "", nil, sourceErr.Code, sourceErr.Message))
		payload["installed_state"] = "error"
		return payload, changedPaths, backupPlan
	}

	manifestFiles := manifestFiles(manifestEntry)
	packHasUpdate := false
	packHasConflict := false
	allSkips := len(sourceFiles) > 0
	for _, sourceFile := range sourceFiles {
		relativeTarget := filepath.ToSlash(filepath.Join(targetPackPath, filepath.FromSlash(sourceFile.RelativePath)))
		targetFile := filepath.Join(profileRoot, filepath.FromSlash(relativeTarget))
		previousSHA, previousErr := existingFileSHA(targetFile)
		var changed ChangedPath
		switch {
		case previousErr != nil:
			bytes := sourceFile.Bytes
			changed = changedPath("error", relativeTarget, pack.PackID, nil, sourceFile.SHA256, &bytes, previousErr.Code, previousErr.Message)
		case previousSHA == nil:
			bytes := sourceFile.Bytes
			changed = changedPath("create", relativeTarget, pack.PackID, nil, sourceFile.SHA256, &bytes, "", "")
			allSkips = false
		default:
			manifestSHA := manifestFileSHA(manifestFiles[sourceFile.RelativePath])
			bytes := sourceFile.Bytes
			if manifestSHA == "" {
				changed = changedPath("conflict", relativeTarget, pack.PackID, previousSHA, sourceFile.SHA256, &bytes, "existing_target_not_manifested", "target file exists but is not recorded in a trusted KAS manifest")
				packHasConflict = true
				allSkips = false
			} else if *previousSHA != manifestSHA {
				changed = changedPath("conflict", relativeTarget, pack.PackID, previousSHA, sourceFile.SHA256, &bytes, "profile_local_modification", "target file differs from the trusted manifest checksum")
				packHasConflict = true
				allSkips = false
			} else if *previousSHA == sourceFile.SHA256 {
				changed = changedPath("skip", relativeTarget, pack.PackID, previousSHA, sourceFile.SHA256, &bytes, "", "")
			} else {
				changed = changedPath("update", relativeTarget, pack.PackID, previousSHA, sourceFile.SHA256, &bytes, "", "")
				backup := BackupEntry{
					Path:           relativeTarget,
					BackupPath:     ".kas/backups/dry-run/" + relativeTarget,
					PreviousSHA256: *previousSHA,
					Bytes:          sourceFile.Bytes,
				}
				changed.BackupPath = backup.BackupPath
				backupPlan = append(backupPlan, backup)
				packHasUpdate = true
				allSkips = false
			}
		}
		changedPaths = append(changedPaths, changed)
		filesPayload = append(filesPayload, map[string]any{
			"relative_path":   sourceFile.RelativePath,
			"action":          changed.Action,
			"bytes":           sourceFile.Bytes,
			"previous_sha256": changed.PreviousSHA256,
			"new_sha256":      sourceFile.SHA256,
			"sha256":          sourceFile.SHA256,
			"mode":            sourceFile.Mode,
		})
	}
	payload["files"] = filesPayload
	if packHasConflict {
		payload["installed_state"] = "conflict"
	} else if packHasUpdate {
		payload["installed_state"] = "installed_drifted"
	} else if allSkips {
		payload["installed_state"] = "installed_current"
	} else if manifestEntry == nil {
		payload["installed_state"] = "not_installed"
	} else {
		payload["installed_state"] = "installed_drifted"
	}
	return payload, changedPaths, backupPlan
}

func buildResult(ok bool, sourceRepo discovery.SourceRepo, targetProfile discovery.TargetProfile, requestedPackIDs []string, packs []map[string]any, changedPaths []ChangedPath, backupPlan []BackupEntry, diagnostics []discovery.Diagnostic) Result {
	sort.Slice(changedPaths, func(i, j int) bool {
		if changedPaths[i].Action == changedPaths[j].Action {
			return changedPaths[i].Path < changedPaths[j].Path
		}
		return changedPaths[i].Action < changedPaths[j].Action
	})
	sort.Slice(backupPlan, func(i, j int) bool { return backupPlan[i].Path < backupPlan[j].Path })
	counts := map[string]int{"create": 0, "update": 0, "skip": 0, "conflict": 0, "error": 0, "backup": 0, "manifest_update": 0}
	for _, entry := range changedPaths {
		counts[entry.Action]++
	}
	sourceChecksums := map[string]any{}
	normalizedPacks := []map[string]any{}
	sortedPacks := append([]map[string]any{}, packs...)
	sort.Slice(sortedPacks, func(i, j int) bool { return sortedPacks[i]["pack_id"].(string) < sortedPacks[j]["pack_id"].(string) })
	for _, pack := range sortedPacks {
		packID := pack["pack_id"].(string)
		sourceChecksums[packID] = pack["pack_checksum"]
		normalizedPacks = append(normalizedPacks, map[string]any{
			"pack_id":     packID,
			"source_path": pack["source_path"],
			"target_path": pack["target_path"],
		})
	}
	canonicalChanged := canonicalChangedPaths(changedPaths)
	conflicts := []map[string]any{}
	errors := []map[string]any{}
	for _, entry := range canonicalChanged {
		switch entry["action"] {
		case "conflict":
			conflicts = append(conflicts, entry)
		case "error":
			errors = append(errors, entry)
		}
	}
	canonicalPlan := map[string]any{
		"command_mode":        "install:dry_run",
		"cli_version":         CLIVersion,
		"requested_pack_ids":  requestedPackIDs,
		"category_expansions": map[string]any{},
		"source_repo": map[string]any{
			"path":       sourceRepo.Path,
			"git_commit": sourceRepo.GitCommit,
			"dirty":      sourceRepo.Dirty,
		},
		"source_pack_checksums":    sourceChecksums,
		"target_profile":           map[string]any{"name": targetProfile.Name, "root": targetProfile.Root, "manifest_path": targetProfile.ManifestPath},
		"normalized_packs":         normalizedPacks,
		"changed_paths":            canonicalChanged,
		"conflicts":                conflicts,
		"errors":                   errors,
		"backup_plan":              backupPlan,
		"manifest_path":            targetProfile.ManifestPath,
		"previous_manifest_sha256": targetProfile.PreviousManifestSHA256,
	}
	planHash := "sha256:" + canonicalHash(canonicalPlan)
	totalFiles := 0
	for _, entry := range changedPaths {
		if entry.Action != "error" {
			totalFiles++
		}
	}
	return Result{
		OK:             ok,
		Command:        "install",
		Mode:           "dry_run",
		CLIVersion:     CLIVersion,
		SourceRepo:     sourceRepo,
		TargetProfile:  targetProfile,
		Requested:      Requested{PackIDs: requestedPackIDs, CategoryExpansions: map[string][]any{}},
		Summary:        Summary{TotalPacks: len(packs), TotalFiles: totalFiles, CountsByAction: counts, ConflictCount: counts["conflict"]},
		DryRunPlanHash: planHash,
		CanonicalPlan:  canonicalPlan,
		Packs:          packs,
		ChangedPaths:   changedPaths,
		BackupPlan:     backupPlan,
		ApprovalRequest: ApprovalRequest{
			Required:       ok,
			Summary:        fmt.Sprintf("Approve copying %d pack / %d file changes into profile %s.", len(packs), counts["create"]+counts["update"], targetProfile.Name),
			EvidenceRef:    "dry-run:" + planHash,
			DryRunPlanHash: planHash,
		},
		Diagnostics: diagnostics,
		NextAction:  "Review changed_paths. KAB is not required for minimum dry-run; execution-runtime remains KAB-gated. Approved copy install is deferred to CLIMVP-004.",
	}
}

type manifestLoadError struct {
	Code    string
	Message string
	SHA256  string
}

func loadManifest(path string) (string, map[string]any, []byte, manifestLoadError) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "missing", nil, nil, manifestLoadError{}
		}
		return "error", nil, nil, manifestLoadError{Code: "manifest_parse_error", Message: "cannot parse KAS manifest: " + err.Error(), SHA256: fileSHAIfPossible(path)}
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "error", nil, nil, manifestLoadError{Code: "manifest_parse_error", Message: "cannot parse KAS manifest: " + err.Error(), SHA256: fileSHAIfPossible(path)}
	}
	if version, _ := manifest["version"].(string); version != ManifestVersion {
		return "error", nil, nil, manifestLoadError{Code: "unsupported_manifest_version", Message: fmt.Sprintf("unsupported KAS manifest version: %q", manifest["version"]), SHA256: shaBytes(data)}
	}
	if kind, _ := manifest["kind"].(string); kind != ManifestKind {
		return "error", nil, nil, manifestLoadError{Code: "unsupported_manifest_kind", Message: fmt.Sprintf("unsupported KAS manifest kind: %q", manifest["kind"]), SHA256: shaBytes(data)}
	}
	return "present", manifest, data, manifestLoadError{}
}

type sourceFileError struct {
	Code         string
	RelativePath string
	Message      string
}

func sourceFiles(packDir string) ([]sourceFile, *sourceFileError) {
	files := []sourceFile{}
	var sourceErr *sourceFileError
	err := filepath.WalkDir(packDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(packDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if excluded(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			sourceErr = &sourceFileError{Code: "source_symlink_rejected", RelativePath: rel, Message: "source pack contains an unsupported symlink: " + rel}
			return filepath.SkipAll
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if discovery.IsInvalidRelativePath(rel) {
			sourceErr = &sourceFileError{Code: "unsafe_source_path", RelativePath: rel, Message: "source pack contains an unsafe relative path: " + rel}
			return filepath.SkipAll
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, sourceFile{RelativePath: rel, SHA256: shaBytes(data), Bytes: len(data), Mode: discovery.ModeString(info.Mode().Perm())})
		return nil
	})
	if sourceErr != nil {
		return nil, sourceErr
	}
	if err != nil {
		return nil, &sourceFileError{Code: "source_read_error", RelativePath: ".", Message: err.Error()}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].RelativePath < files[j].RelativePath })
	return files, nil
}

func manifestInstalls(manifest map[string]any) map[string]map[string]any {
	entries := map[string]map[string]any{}
	installs, _ := manifest["installs"].([]any)
	for _, raw := range installs {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		packID, _ := entry["pack_id"].(string)
		if packID != "" {
			entries[packID] = entry
		}
	}
	return entries
}

func manifestFiles(manifestEntry map[string]any) map[string]map[string]any {
	entries := map[string]map[string]any{}
	if manifestEntry == nil {
		return entries
	}
	files, _ := manifestEntry["files"].([]any)
	for _, raw := range files {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		rel, _ := entry["relative_path"].(string)
		if rel != "" {
			entries[rel] = entry
		}
	}
	return entries
}

func manifestPathErrors(manifestEntry map[string]any, targetPackPath string, packID string) []ChangedPath {
	if manifestEntry == nil {
		return nil
	}
	errors := []ChangedPath{}
	targetPath, ok := manifestEntry["target_path"].(string)
	if !ok || discovery.IsInvalidRelativePath(targetPath) {
		message := fmt.Sprintf("KAS manifest target_path is unsafe for pack %s: %v", packID, manifestEntry["target_path"])
		errors = append(errors, changedPath("error", targetPackPath, packID, nil, "", nil, "unsafe_manifest_target_path", message))
	} else if targetPath != targetPackPath {
		message := fmt.Sprintf("KAS manifest target_path for pack %s is %q, expected %q", packID, targetPath, targetPackPath)
		errors = append(errors, changedPath("error", targetPackPath, packID, nil, "", nil, "manifest_target_path_mismatch", message))
	}
	files, ok := manifestEntry["files"].([]any)
	if !ok {
		return errors
	}
	for _, raw := range files {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		rel, ok := entry["relative_path"].(string)
		if !ok || discovery.IsInvalidRelativePath(rel) {
			message := fmt.Sprintf("KAS manifest file relative_path is unsafe for pack %s: %v", packID, entry["relative_path"])
			errors = append(errors, changedPath("error", filepath.ToSlash(filepath.Join(targetPackPath, "<manifest-file>")), packID, nil, "", nil, "unsafe_manifest_file_path", message))
		}
	}
	return errors
}

func manifestFileSHA(file map[string]any) string {
	if file == nil {
		return ""
	}
	for _, key := range []string{"new_sha256", "sha256"} {
		value, _ := file[key].(string)
		if value != "" {
			return value
		}
	}
	return ""
}

type pathReadError struct {
	Code    string
	Message string
}

func existingFileSHA(path string) (*string, *pathReadError) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
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
	sum := shaBytes(data)
	return &sum, nil
}

func changedPath(action, path, packID string, previousSHA *string, newSHA string, bytes *int, errorCode, errorMessage string) ChangedPath {
	entry := ChangedPath{Path: path, Action: action}
	if packID != "" {
		entry.PackID = packID
	}
	if previousSHA != nil || action == "create" || action == "skip" || action == "update" || action == "conflict" {
		entry.PreviousSHA256 = previousSHA
	}
	if newSHA != "" {
		entry.NewSHA256 = newSHA
	}
	if bytes != nil {
		entry.Bytes = bytes
	}
	entry.ErrorCode = errorCode
	entry.ErrorMessage = errorMessage
	return entry
}

func canonicalChangedPaths(changedPaths []ChangedPath) []map[string]any {
	canonical := []map[string]any{}
	for _, entry := range changedPaths {
		item := map[string]any{
			"action":          entry.Action,
			"path":            entry.Path,
			"previous_sha256": entry.PreviousSHA256,
			"new_sha256":      nil,
			"bytes":           nil,
		}
		if entry.NewSHA256 != "" {
			item["new_sha256"] = entry.NewSHA256
		}
		if entry.Bytes != nil {
			item["bytes"] = *entry.Bytes
		}
		if entry.ErrorCode != "" {
			item["error_code"] = entry.ErrorCode
		}
		if entry.ErrorMessage != "" {
			item["error_message"] = entry.ErrorMessage
		}
		canonical = append(canonical, item)
	}
	return canonical
}

func approvedHashFromEvidence(evidenceRef string) (string, bool) {
	hash, ok := strings.CutPrefix(evidenceRef, "dry-run:")
	if !ok || hash == "" {
		return hash, false
	}
	return hash, true
}

func makeInstallID(planHash string, now time.Time) string {
	short := strings.TrimPrefix(planHash, "sha256:")
	if len(short) > 12 {
		short = short[:12]
	}
	return "kas-install-" + now.UTC().Format("20060102T150405Z") + "-" + short
}

func approvedFailure(dryRun Result, evidenceRef string, approvedHash string, installID string, backupRoot string, changed []ChangedPath, code string, message string) Result {
	diagnostics := []discovery.Diagnostic{{Level: "error", Code: code, Message: message}}
	diagnostics = append(diagnostics, dryRun.Diagnostics...)
	result := buildApprovedResult(dryRun, evidenceRef, approvedHash, installID, backupRoot, changed, false, diagnostics)
	result.NextAction = "설치가 중단되었습니다. 파일과 manifest 상태를 확인하고 새 dry-run부터 다시 실행하세요."
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
		OK:             ok,
		Command:        "install",
		Mode:           "approved_copy",
		CLIVersion:     dryRun.CLIVersion,
		SourceRepo:     dryRun.SourceRepo,
		TargetProfile:  dryRun.TargetProfile,
		Requested:      dryRun.Requested,
		Summary:        Summary{TotalPacks: dryRun.Summary.TotalPacks, TotalFiles: counts["create"] + counts["update"], CountsByAction: counts, ConflictCount: counts["conflict"]},
		DryRunPlanHash: dryRun.DryRunPlanHash,
		CanonicalPlan:  dryRun.CanonicalPlan,
		Packs:          dryRun.Packs,
		ChangedPaths:   changed,
		BackupPlan:     dryRun.BackupPlan,
		Approval: ApprovalEvidence{
			EvidenceRef:        evidenceRef,
			DryRunPlanHash:     dryRun.DryRunPlanHash,
			ApprovedPlanHash:   approvedHash,
			MatchedCurrentPlan: approvedHash == dryRun.DryRunPlanHash,
		},
		InstallID:    installID,
		ManifestPath: dryRun.TargetProfile.ManifestPath,
		BackupPath:   backupRoot,
		Recovery: &Recovery{
			RollbackSupported:      true,
			BackupPath:             backupRoot,
			PreviousManifestSHA256: dryRun.TargetProfile.PreviousManifestSHA256,
			Instructions:           []string{"Restore files from backup_path", "Restore previous manifest snapshot when present"},
		},
		Diagnostics: diagnostics,
	}
}

func sourceRelativeForChangedPath(packs []map[string]any, entry ChangedPath) (string, bool) {
	for _, pack := range packs {
		if packID, _ := pack["pack_id"].(string); packID != entry.PackID {
			continue
		}
		targetPath, _ := pack["target_path"].(string)
		sourcePath, _ := pack["source_path"].(string)
		prefix := targetPath + "/"
		if !strings.HasPrefix(entry.Path, prefix) {
			return "", false
		}
		rel := strings.TrimPrefix(entry.Path, prefix)
		if discovery.IsInvalidRelativePath(rel) {
			return "", false
		}
		return filepath.ToSlash(filepath.Join(sourcePath, filepath.FromSlash(rel))), true
	}
	return "", false
}

func buildUpdatedManifest(dryRun Result, evidenceRef string, approvedHash string, installID string, backupRoot string, previousManifestPath string, changed []ChangedPath) (map[string]any, error) {
	_, existing, _, _ := loadManifest(dryRun.TargetProfile.ManifestPath)
	installs := []any{}
	if existing != nil {
		if rawInstalls, ok := existing["installs"].([]any); ok {
			for _, raw := range rawInstalls {
				entry, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				packID, _ := entry["pack_id"].(string)
				if requestedPack(dryRun.Requested.PackIDs, packID) {
					continue
				}
				installs = append(installs, entry)
			}
		}
	}
	for _, pack := range dryRun.Packs {
		entry, err := manifestEntryForPack(dryRun, pack, evidenceRef, approvedHash, installID, backupRoot, previousManifestPath, changed)
		if err != nil {
			return nil, err
		}
		installs = append(installs, entry)
	}
	sort.Slice(installs, func(i, j int) bool {
		left := installs[i].(map[string]any)["pack_id"].(string)
		right := installs[j].(map[string]any)["pack_id"].(string)
		return left < right
	})
	return map[string]any{
		"version": ManifestVersion,
		"kind":    ManifestKind,
		"profile": map[string]any{
			"name": dryRun.TargetProfile.Name,
			"root": dryRun.TargetProfile.Root,
		},
		"source_repo": map[string]any{
			"path":       dryRun.SourceRepo.Path,
			"git_remote": nil,
			"git_commit": dryRun.SourceRepo.GitCommit,
			"dirty":      dryRun.SourceRepo.Dirty,
		},
		"installs": installs,
	}, nil
}

func manifestEntryForPack(dryRun Result, pack map[string]any, evidenceRef string, approvedHash string, installID string, backupRoot string, previousManifestPath string, changed []ChangedPath) (map[string]any, error) {
	packID := pack["pack_id"].(string)
	files := []any{}
	for _, raw := range pack["files"].([]map[string]any) {
		action, _ := raw["action"].(string)
		if action != "create" && action != "update" && action != "skip" {
			continue
		}
		rel := raw["relative_path"].(string)
		changedPath := changedEntryForPackFile(changed, packID, rel)
		backupRel := any(nil)
		if changedPath != nil && changedPath.BackupPath != "" {
			backupRel = changedPath.BackupPath
		}
		files = append(files, map[string]any{
			"relative_path":        rel,
			"action":               action,
			"bytes":                raw["bytes"],
			"previous_sha256":      raw["previous_sha256"],
			"new_sha256":           raw["new_sha256"],
			"sha256":               raw["sha256"],
			"backup_relative_path": backupRel,
			"mode":                 raw["mode"],
		})
	}
	category, _ := pack["category"].(string)
	name, _ := pack["name"].(string)
	backupRequired := false
	for _, entry := range changed {
		if entry.PackID == packID && entry.Action == "backup" {
			backupRequired = true
			break
		}
	}
	return map[string]any{
		"install_id":            installID,
		"installed_at":          time.Now().UTC().Format(time.RFC3339),
		"approval_evidence_ref": evidenceRef,
		"dry_run_plan_hash":     dryRun.DryRunPlanHash,
		"approved_plan_hash":    approvedHash,
		"pack_id":               packID,
		"category":              category,
		"name":                  name,
		"source_path":           pack["source_path"],
		"target_path":           pack["target_path"],
		"checksum_algorithm":    "sha256",
		"pack_checksum":         pack["pack_checksum"],
		"backup": map[string]any{
			"required": backupRequired,
			"path":     backupRoot,
			"created":  backupRequired,
		},
		"previous_manifest": map[string]any{
			"path":   previousManifestPath,
			"sha256": dryRun.TargetProfile.PreviousManifestSHA256,
		},
		"files": files,
	}, nil
}

func requestedPack(packIDs []string, packID string) bool {
	for _, requested := range packIDs {
		if requested == packID {
			return true
		}
	}
	return false
}

func changedEntryForPackFile(changed []ChangedPath, packID string, rel string) *ChangedPath {
	suffix := "/" + rel
	for i := range changed {
		if changed[i].PackID == packID && strings.HasSuffix(changed[i].Path, suffix) && (changed[i].Action == "create" || changed[i].Action == "update" || changed[i].Action == "skip") {
			return &changed[i]
		}
	}
	return nil
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".kas-manifest-*")
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
	if err := tmp.Chmod(0o644); err != nil {
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
			if errors.Is(err, os.ErrNotExist) {
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

func checksumFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return shaBytes(data), nil
}

func resolveProfileRoot(profile string, profileRoot string) string {
	if profileRoot != "" {
		abs, _ := filepath.Abs(profileRoot)
		return abs
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	root, _ := filepath.Abs(filepath.Join(home, ".hermes", "profiles", profile))
	return root
}

func excluded(relativePath string) bool {
	parts := strings.Split(relativePath, "/")
	for _, part := range parts {
		if part == ".git" || part == ".kkachi" || part == "__pycache__" {
			return true
		}
		if part == ".DS_Store" || strings.HasSuffix(part, ".swp") || strings.HasSuffix(part, ".swo") {
			return true
		}
	}
	return false
}

func canonicalHash(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func shaBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func fileSHAIfPossible(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return shaBytes(data)
}
