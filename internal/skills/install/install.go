package install

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
