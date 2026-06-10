package kasstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/install"
)

const (
	Command         = "sync-project-kas"
	CLIVersion      = install.CLIVersion
	SchemaVersion   = "0.1"
	stage3Canonical = "stage3_kab_backend_selected"
)

type Options struct {
	Profile          string
	Project          string
	StatePath        string
	LegacyMarkerPath string
	DryRun           bool
	RepoPath         string
	ProjectRoot      string
}

type ReadSurface struct {
	State   string `json:"state"`
	Path    string `json:"path"`
	SHA256  string `json:"sha256,omitempty"`
	Message string `json:"message,omitempty"`
}

type ReadSurfaces struct {
	YAML         ReadSurface `json:"yaml"`
	LegacyMarker ReadSurface `json:"legacy_marker"`
}

type EffectiveStageClaim struct {
	Numeric                  int    `json:"numeric"`
	Canonical                string `json:"canonical"`
	Source                   string `json:"source"`
	KABExecutionClaimAllowed bool   `json:"kab_execution_claim_allowed"`
	FailClosedToStage1       bool   `json:"fail_closed_to_stage1"`
}

type Validation struct {
	SchemaVersion     string                 `json:"schema_version,omitempty"`
	PackBaselineCount int                    `json:"pack_baseline_count"`
	Diagnostics       []discovery.Diagnostic `json:"diagnostics"`
}

type BaselineRepo struct {
	GitCommit        string `json:"git_commit,omitempty"`
	DirtyRecorded    *bool  `json:"dirty_recorded,omitempty"`
	BaselineVerified bool   `json:"baseline_verified"`
}

type ProjectRoot struct {
	Path       string `json:"path,omitempty"`
	Resolution string `json:"resolution"`
	State      string `json:"state"`
}

type Summary struct {
	TotalMappings           int            `json:"total_mappings"`
	CountsByClassification  map[string]int `json:"counts_by_classification"`
	NoActionCount           int            `json:"no_action_count"`
	SemanticPortPacketCount int            `json:"semantic_port_packet_count"`
	WriteCount              int            `json:"write_count"`
}

type LifecycleNoWriteEvidence struct {
	Guaranteed                   bool `json:"guaranteed"`
	ProfileWriteCount            int  `json:"profile_write_count"`
	SkillWriteCount              int  `json:"skill_write_count"`
	ManifestWriteCount           int  `json:"manifest_write_count"`
	KAHStateWriteCount           int  `json:"kah_state_write_count"`
	KABRuntimeMutationCount      int  `json:"kab_runtime_mutation_count"`
	HermesRuntimeMutationCount   int  `json:"hermes_runtime_mutation_count"`
	AuthProviderConfigWriteCount int  `json:"auth_provider_config_write_count"`
	ProfileActivationCount       int  `json:"profile_activation_count"`
}

type LifecycleTargetProfile struct {
	Name          string `json:"name"`
	Role          string `json:"role"`
	StatePath     string `json:"state_path"`
	ProjectRoot   string `json:"project_root,omitempty"`
	LegacyMarker  string `json:"legacy_marker_path,omitempty"`
	PlannedState  string `json:"planned_state"`
	DoctorCommand string `json:"doctor_command"`
}

type LifecycleSourcePack struct {
	ID        string            `json:"id"`
	Paths     map[string]string `json:"paths"`
	Checksums map[string]string `json:"checksums"`
	State     string            `json:"planned_state"`
}

type LifecyclePlannedState struct {
	ID             string            `json:"id"`
	SkillID        string            `json:"skill_id"`
	SourcePackID   string            `json:"source_pack_id,omitempty"`
	TargetPath     string            `json:"target_path,omitempty"`
	PlannedState   string            `json:"planned_state"`
	Classification string            `json:"classification"`
	Checksums      map[string]string `json:"checksums"`
	ChangedPaths   []string          `json:"changed_paths"`
	DoctorCommand  string            `json:"doctor_command"`
}

type LifecycleBackupRecovery struct {
	BackupRequired        bool     `json:"backup_required"`
	BackupWriteDeferred   bool     `json:"backup_write_deferred"`
	RecoveryWriteDeferred bool     `json:"recovery_write_deferred"`
	Instructions          []string `json:"instructions"`
}

type LifecycleApprovalRequest struct {
	Required                    bool   `json:"required"`
	EvidenceRef                 string `json:"evidence_ref"`
	HashIncludesProfile         bool   `json:"hash_includes_profile"`
	HashIncludesChangedPaths    bool   `json:"hash_includes_changed_paths"`
	HashIncludesNoWriteEvidence bool   `json:"hash_includes_no_write_evidence"`
	HashIncludesBackupPlan      bool   `json:"hash_includes_backup_plan"`
}

type LifecycleApprovalEvidence struct {
	EvidenceRef        string `json:"evidence_ref"`
	DryRunPlanHash     string `json:"dry_run_plan_hash"`
	ApprovedPlanHash   string `json:"approved_plan_hash"`
	MatchedCurrentPlan bool   `json:"matched_current_plan"`
}

type LifecycleUpdateResult struct {
	OK                 bool                      `json:"ok"`
	Command            string                    `json:"command"`
	Mode               string                    `json:"mode"`
	CLIVersion         string                    `json:"cli_version"`
	DryRun             bool                      `json:"dry_run"`
	TargetRoles        []string                  `json:"target_roles"`
	TargetProfiles     []LifecycleTargetProfile  `json:"target_profiles"`
	SourcePacks        []LifecycleSourcePack     `json:"source_packs"`
	SkillIDs           []string                  `json:"skill_ids"`
	PlannedStates      []LifecyclePlannedState   `json:"planned_states"`
	ChangedPaths       []string                  `json:"changed_paths"`
	BackupRecovery     LifecycleBackupRecovery   `json:"backup_recovery"`
	DoctorCommands     []string                  `json:"doctor_commands"`
	NoWrite            LifecycleNoWriteEvidence  `json:"no_write"`
	SyncClassification Result                    `json:"sync_classification"`
	PlanHash           string                    `json:"plan_hash"`
	ApprovalRequest    LifecycleApprovalRequest  `json:"approval_request,omitempty"`
	Approval           LifecycleApprovalEvidence `json:"approval,omitempty"`
	Diagnostics        []discovery.Diagnostic    `json:"diagnostics"`
	NextAction         string                    `json:"next_action"`
}

type Classification struct {
	ID                   string                 `json:"id"`
	UpstreamPack         string                 `json:"upstream_pack,omitempty"`
	ProjectSkill         string                 `json:"project_skill,omitempty"`
	Classification       string                 `json:"classification"`
	Basis                []string               `json:"basis"`
	Paths                map[string]string      `json:"paths"`
	Checksums            map[string]string      `json:"checksums"`
	RequiresSemanticPort bool                   `json:"requires_semantic_port"`
	Diagnostics          []discovery.Diagnostic `json:"diagnostics"`
}

func BuildLifecycleUpdate(opts Options) LifecycleUpdateResult {
	sync := Build(opts)
	doctorCommand := fmt.Sprintf("kkachi-hermes-skills doctor --profile %s --project %s --project-suite", opts.Profile, opts.Project)
	result := LifecycleUpdateResult{
		OK:             sync.OK,
		Command:        "update",
		Mode:           "project_update_dry_run",
		CLIVersion:     CLIVersion,
		DryRun:         true,
		TargetRoles:    []string{"project_suite"},
		TargetProfiles: []LifecycleTargetProfile{{Name: opts.Profile, Role: "project_suite", StatePath: opts.StatePath, LegacyMarker: sync.LegacyMarkerPath, PlannedState: lifecycleOverallState(sync), DoctorCommand: doctorCommand}},
		SourcePacks:    []LifecycleSourcePack{},
		SkillIDs:       []string{},
		PlannedStates:  []LifecyclePlannedState{},
		ChangedPaths:   []string{},
		BackupRecovery: LifecycleBackupRecovery{
			BackupRequired:        lifecycleBackupRequired(sync),
			BackupWriteDeferred:   true,
			RecoveryWriteDeferred: true,
			Instructions:          []string{"TOKEN-004 is read-only; review update classifications before TOKEN-005 apply.", "Run the listed doctor command after future approved writes."},
		},
		DoctorCommands:     []string{doctorCommand},
		NoWrite:            LifecycleNoWriteEvidence{Guaranteed: true},
		SyncClassification: sync,
		Diagnostics:        append([]discovery.Diagnostic(nil), sync.Validation.Diagnostics...),
		NextAction:         sync.NextAction,
	}
	if sync.ProjectRoot != nil {
		result.TargetProfiles[0].ProjectRoot = sync.ProjectRoot.Path
	}
	for _, classification := range sync.Classifications {
		result.Diagnostics = append(result.Diagnostics, classification.Diagnostics...)
		state := lifecycleStateForClassification(classification.Classification)
		changed := lifecycleChangedPaths(classification.Paths)
		result.PlannedStates = append(result.PlannedStates, LifecyclePlannedState{
			ID:             classification.ID,
			SkillID:        classification.ProjectSkill,
			SourcePackID:   classification.UpstreamPack,
			TargetPath:     classification.Paths["project_skill_path"],
			PlannedState:   state,
			Classification: classification.Classification,
			Checksums:      classification.Checksums,
			ChangedPaths:   changed,
			DoctorCommand:  doctorCommand,
		})
		result.SourcePacks = append(result.SourcePacks, LifecycleSourcePack{ID: classification.UpstreamPack, Paths: classification.Paths, Checksums: classification.Checksums, State: state})
		if classification.ProjectSkill != "" {
			result.SkillIDs = append(result.SkillIDs, classification.ProjectSkill)
		}
		result.ChangedPaths = append(result.ChangedPaths, changed...)
	}
	for _, unchanged := range sync.UnchangedMappings {
		result.PlannedStates = append(result.PlannedStates, LifecyclePlannedState{
			ID:             unchanged.ID,
			SkillID:        unchanged.ProjectSkill,
			SourcePackID:   unchanged.UpstreamPack,
			TargetPath:     unchanged.Paths["project_skill_path"],
			PlannedState:   "no_change",
			Classification: "unchanged",
			Checksums:      unchanged.Checksums,
			DoctorCommand:  doctorCommand,
		})
		result.SourcePacks = append(result.SourcePacks, LifecycleSourcePack{ID: unchanged.UpstreamPack, Paths: unchanged.Paths, Checksums: unchanged.Checksums, State: "no_change"})
		result.SkillIDs = append(result.SkillIDs, unchanged.ProjectSkill)
	}
	result.SkillIDs = uniqueStrings(result.SkillIDs)
	result.ChangedPaths = uniqueStrings(result.ChangedPaths)
	sort.Slice(result.SourcePacks, func(i, j int) bool {
		if result.SourcePacks[i].ID == result.SourcePacks[j].ID {
			return result.SourcePacks[i].State < result.SourcePacks[j].State
		}
		return result.SourcePacks[i].ID < result.SourcePacks[j].ID
	})
	if !sync.OK {
		result.NextAction = "Resolve update dry-run diagnostics before TOKEN-005 apply; no files were written."
	} else {
		result.NextAction = "Review update --dry-run planned_states and run doctor after future approved writes; no files were written."
	}
	finalizeLifecycleUpdateHash(&result)
	return result
}

func ApplyLifecycleUpdate(opts Options, evidenceRef string) LifecycleUpdateResult {
	opts.DryRun = true
	dryRun := BuildLifecycleUpdate(opts)
	approvedHash, evidenceOK := lifecycleApprovedHashFromEvidence(evidenceRef)
	if !evidenceOK {
		return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "approval_evidence_malformed", "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>.")
	}
	if approvedHash != dryRun.PlanHash {
		return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "approval_plan_hash_mismatch", "approval evidence does not match the current dry-run plan hash; no files were written.")
	}
	if !dryRun.OK {
		return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "project_update_plan_not_approvable", "current dry-run plan contains conflict or error entries; no files were written.")
	}
	writable := []Classification{}
	for _, classification := range dryRun.SyncClassification.Classifications {
		switch classification.Classification {
		case "auto_copy_candidate":
			writable = append(writable, classification)
		case "new_upstream_candidate", "semantic_merge_required", "removed_or_renamed_upstream", "fail_closed_conflict":
			return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "project_update_manual_review_required", "update --apply writes only hash-bound auto-copy candidates; semantic, new, removed, or conflict classifications require manual review.")
		case "local_only":
			// no write
		default:
			return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "project_update_unknown_classification", "update --apply cannot write unknown classification "+classification.Classification+".")
		}
	}
	if len(writable) == 0 {
		return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "project_update_no_auto_copy_candidates", "current dry-run plan has no auto-copy update candidates; no files were written.")
	}
	for _, classification := range writable {
		if err := preflightLifecycleAutoCopy(dryRun, classification); err != nil {
			return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "project_update_preflight_failed", err.Error())
		}
	}
	for _, classification := range writable {
		if err := copyLifecyclePack(classification.Paths["current_upstream_path"], classification.Paths["project_skill_path"], dryRun.SyncClassification.SourceRepo.Path, dryRun.SyncClassification.ProjectRoot.Path); err != nil {
			return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "project_update_write_failed", err.Error())
		}
	}
	stateBytes, err := os.ReadFile(opts.StatePath)
	if err != nil {
		return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "state_read_failed", err.Error())
	}
	updatedState, err := rewriteLifecycleStateBaselines(string(stateBytes), writable)
	if err != nil {
		return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "state_update_failed", err.Error())
	}
	if err := writeLifecycleFileAtomic(opts.StatePath, []byte(updatedState), 0o644); err != nil {
		return lifecycleApprovalFailure(dryRun, evidenceRef, approvedHash, "state_write_failed", err.Error())
	}
	result := dryRun
	result.OK = true
	result.Mode = "project_update_approved"
	result.DryRun = false
	result.NoWrite = LifecycleNoWriteEvidence{Guaranteed: false, ProfileWriteCount: len(writable), SkillWriteCount: len(writable)}
	result.Approval = LifecycleApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: true}
	result.NextAction = "Update apply complete. Run the listed doctor command before claiming profile readiness."
	return result
}

func RenderHumanLifecycleUpdate(result LifecycleUpdateResult) string {
	status := "ready"
	if !result.OK {
		status = "blocked"
	} else if !result.DryRun {
		status = "complete"
	}
	lines := []string{
		fmt.Sprintf("Status: update %s for profile %s.", status, firstTargetProfileName(result.TargetProfiles)),
		fmt.Sprintf("Targets: roles=%s profiles=%d source_packs=%d skills=%d.", strings.Join(result.TargetRoles, ","), len(result.TargetProfiles), len(result.SourcePacks), len(result.SkillIDs)),
		fmt.Sprintf("Planned states: %s", lifecycleStateCounts(result.PlannedStates)),
	}
	if result.DryRun {
		lines = append(lines, "Writes: dry-run only; profile/auth/token/gateway/provider/model/KAB/KAH/Hermes runtime/profile activation writes 0.")
		lines = append(lines, "Approval evidence: "+result.ApprovalRequest.EvidenceRef)
	} else {
		lines = append(lines, "Approval evidence: "+result.Approval.EvidenceRef)
	}
	for _, state := range result.PlannedStates {
		lines = append(lines, fmt.Sprintf("Plan: %s %s -> %s (%s)", state.SkillID, state.TargetPath, state.PlannedState, state.Classification))
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, fmt.Sprintf("%s: %s - %s", diagnostic.Level, diagnostic.Code, diagnostic.Message))
	}
	for _, command := range result.DoctorCommands {
		lines = append(lines, "Doctor: "+command)
	}
	lines = append(lines, "Next: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func lifecycleStateForClassification(classification string) string {
	switch classification {
	case "auto_copy_candidate":
		return "update"
	case "new_upstream_candidate":
		return "create"
	case "local_only":
		return "no_change"
	case "semantic_merge_required", "removed_or_renamed_upstream":
		return "blocked"
	case "fail_closed_conflict":
		return "error"
	default:
		return "no_change"
	}
}

func finalizeLifecycleUpdateHash(result *LifecycleUpdateResult) {
	canonical := map[string]any{
		"command":             result.Command,
		"mode":                result.Mode,
		"dry_run":             result.DryRun,
		"target_roles":        result.TargetRoles,
		"target_profiles":     result.TargetProfiles,
		"source_packs":        result.SourcePacks,
		"skill_ids":           result.SkillIDs,
		"planned_states":      result.PlannedStates,
		"changed_paths":       result.ChangedPaths,
		"backup_recovery":     result.BackupRecovery,
		"doctor_commands":     result.DoctorCommands,
		"no_write":            result.NoWrite,
		"sync_classification": result.SyncClassification,
		"diagnostics":         result.Diagnostics,
	}
	result.PlanHash = checksumAny(canonical)
	result.ApprovalRequest = LifecycleApprovalRequest{Required: result.OK && len(result.ChangedPaths) > 0, EvidenceRef: "dry-run:" + result.PlanHash, HashIncludesProfile: true, HashIncludesChangedPaths: true, HashIncludesNoWriteEvidence: true, HashIncludesBackupPlan: true}
}

func lifecycleApprovedHashFromEvidence(evidenceRef string) (string, bool) {
	if !lifecycleApprovedEvidencePattern.MatchString(evidenceRef) {
		hash, _ := strings.CutPrefix(evidenceRef, "dry-run:")
		return hash, false
	}
	return strings.TrimPrefix(evidenceRef, "dry-run:"), true
}

func lifecycleApprovalFailure(dryRun LifecycleUpdateResult, evidenceRef string, approvedHash string, code string, message string) LifecycleUpdateResult {
	result := dryRun
	result.OK = false
	result.Mode = "project_update_approved"
	result.DryRun = false
	result.NoWrite = LifecycleNoWriteEvidence{Guaranteed: true}
	result.Approval = LifecycleApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: approvedHash == dryRun.PlanHash}
	result.Diagnostics = append([]discovery.Diagnostic{{Level: "error", Code: code, Message: message}}, dryRun.Diagnostics...)
	result.NextAction = "No files were written. Fix the reported issue, rerun dry-run, and apply only the matching plan hash."
	return result
}

func preflightLifecycleAutoCopy(result LifecycleUpdateResult, classification Classification) error {
	if result.SyncClassification.SourceRepo == nil || result.SyncClassification.ProjectRoot == nil {
		return fmt.Errorf("source repo and project root evidence are required")
	}
	sourceRel := classification.Paths["current_upstream_path"]
	targetRel := classification.Paths["project_skill_path"]
	if sourceRel == "" || targetRel == "" {
		return fmt.Errorf("auto-copy candidate lacks source or target path")
	}
	source, err := cleanExistingDirInside(result.SyncClassification.SourceRepo.Path, filepath.Join(result.SyncClassification.SourceRepo.Path, filepath.FromSlash(sourceRel)))
	if err != nil {
		return err
	}
	target, err := cleanExistingDirInside(result.SyncClassification.ProjectRoot.Path, filepath.Join(result.SyncClassification.ProjectRoot.Path, filepath.FromSlash(targetRel)))
	if err != nil {
		return err
	}
	sourceChecksum, err := discovery.ComputePackChecksum(source)
	if err != nil {
		return err
	}
	targetChecksum, err := discovery.ComputePackChecksum(target)
	if err != nil {
		return err
	}
	if ensureChecksumPrefix(sourceChecksum) != classification.Checksums["current_source_checksum"] {
		return fmt.Errorf("source checksum changed after dry-run: %s", sourceRel)
	}
	if ensureChecksumPrefix(targetChecksum) != classification.Checksums["current_project_checksum"] {
		return fmt.Errorf("target checksum changed after dry-run: %s", targetRel)
	}
	if classification.Checksums["current_project_checksum"] != classification.Checksums["recorded_project_checksum"] {
		return fmt.Errorf("target no longer matches recorded project baseline: %s", targetRel)
	}
	return nil
}

func copyLifecyclePack(sourceRel string, targetRel string, repoRoot string, projectRoot string) error {
	sourceRoot := filepath.Join(repoRoot, filepath.FromSlash(sourceRel))
	targetRoot := filepath.Join(projectRoot, filepath.FromSlash(targetRel))
	return filepath.WalkDir(sourceRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return fmt.Errorf("source path is not a regular file: %s", path)
		}
		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			return fmt.Errorf("source path escapes pack root: %s", path)
		}
		if excludedPackPath(filepath.ToSlash(rel)) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeLifecycleFileAtomic(filepath.Join(targetRoot, rel), data, info.Mode().Perm())
	})
}

func rewriteLifecycleStateBaselines(data string, classifications []Classification) (string, error) {
	updates := map[string]Classification{}
	for _, classification := range classifications {
		updates[classification.UpstreamPack+"|"+classification.ProjectSkill] = classification
	}
	lines := strings.Split(data, "\n")
	currentKey := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- upstream_pack:") {
			upstream := unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "- upstream_pack:")))
			currentKey = upstream + "|"
			continue
		}
		if currentKey != "" && strings.HasPrefix(trimmed, "project_skill:") {
			projectSkill := unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "project_skill:")))
			parts := strings.SplitN(currentKey, "|", 2)
			currentKey = parts[0] + "|" + projectSkill
			continue
		}
		classification, ok := updates[currentKey]
		if !ok {
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "source_checksum:"):
			lines[i] = linePrefixBeforeValue(line) + quoteYAML(classification.Checksums["current_source_checksum"])
		case strings.HasPrefix(trimmed, "project_checksum:"):
			lines[i] = linePrefixBeforeValue(line) + quoteYAML(classification.Checksums["current_source_checksum"])
		}
	}
	return strings.Join(lines, "\n"), nil
}

func linePrefixBeforeValue(line string) string {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return line
	}
	return line[:idx+1] + " "
}

func quoteYAML(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

func writeLifecycleFileAtomic(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".kas-update-*")
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

func checksumAny(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		data = []byte(fmt.Sprint(value))
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func lifecycleOverallState(result Result) string {
	if !result.OK {
		return "error"
	}
	if result.Summary == nil {
		return "blocked"
	}
	if result.Summary.CountsByClassification["semantic_merge_required"] > 0 || result.Summary.CountsByClassification["removed_or_renamed_upstream"] > 0 {
		return "blocked"
	}
	if result.Summary.CountsByClassification["auto_copy_candidate"] > 0 {
		return "update"
	}
	if result.Summary.CountsByClassification["new_upstream_candidate"] > 0 {
		return "create"
	}
	return "no_change"
}

func lifecycleBackupRequired(result Result) bool {
	if result.Summary == nil {
		return false
	}
	return result.Summary.CountsByClassification["auto_copy_candidate"] > 0 || result.Summary.CountsByClassification["semantic_merge_required"] > 0
}

func lifecycleChangedPaths(paths map[string]string) []string {
	out := []string{}
	for _, key := range []string{"project_skill_path", "current_upstream_path"} {
		if paths[key] != "" {
			out = append(out, paths[key])
		}
	}
	return out
}

func lifecycleStateCounts(states []LifecyclePlannedState) string {
	counts := map[string]int{}
	for _, state := range states {
		counts[state.PlannedState]++
	}
	parts := []string{}
	for _, key := range []string{"create", "update", "remove", "no_change", "blocked", "conflict", "error"} {
		if counts[key] > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func firstTargetProfileName(profiles []LifecycleTargetProfile) string {
	if len(profiles) == 0 {
		return ""
	}
	return profiles[0].Name
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

type UnchangedMapping struct {
	ID           string            `json:"id"`
	UpstreamPack string            `json:"upstream_pack"`
	ProjectSkill string            `json:"project_skill"`
	Paths        map[string]string `json:"paths"`
	Checksums    map[string]string `json:"checksums"`
}

type SemanticPortPacket struct {
	PacketID                string `json:"packet_id"`
	ClassificationID        string `json:"classification_id"`
	RecommendedArtifactPath string `json:"recommended_artifact_path"`
	ContentSHA256           string `json:"content_sha256"`
	Content                 string `json:"content"`
}

type Result struct {
	OK                           bool                  `json:"ok"`
	Command                      string                `json:"command"`
	Mode                         string                `json:"mode"`
	CLIVersion                   string                `json:"cli_version"`
	DryRun                       bool                  `json:"dry_run"`
	TargetProfile                string                `json:"target_profile"`
	ProjectID                    string                `json:"project_id"`
	YAMLStatePath                string                `json:"yaml_state_path"`
	LegacyMarkerPath             string                `json:"legacy_marker_path"`
	StateSource                  string                `json:"state_source"`
	ReadSurfaces                 ReadSurfaces          `json:"read_surfaces"`
	EffectiveStageClaim          EffectiveStageClaim   `json:"effective_stage_claim"`
	WriteTargetAfterApprovedSync string                `json:"write_target_after_approved_sync"`
	Validation                   Validation            `json:"validation"`
	SourceRepo                   *discovery.SourceRepo `json:"source_repo,omitempty"`
	BaselineRepo                 *BaselineRepo         `json:"baseline_repo,omitempty"`
	ProjectRoot                  *ProjectRoot          `json:"project_root,omitempty"`
	Summary                      *Summary              `json:"summary,omitempty"`
	Classifications              []Classification      `json:"classifications,omitempty"`
	UnchangedMappings            []UnchangedMapping    `json:"unchanged_mappings,omitempty"`
	SemanticPortPackets          []SemanticPortPacket  `json:"semantic_port_packets,omitempty"`
	NextAction                   string                `json:"next_action"`
}

type stateFile struct {
	version         string
	project         map[string]string
	stage           map[string]string
	upstream        map[string]string
	packBaselines   []map[string]string
	overlayPolicy   map[string]string
	updatePolicy    map[string][]string
	updateScalars   map[string]string
	evidencePosture map[string]string
}

var (
	shaPattern                       = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	checksumPattern                  = regexp.MustCompile(`^sha256:[0-9a-fA-F]{64}$`)
	lifecycleApprovedEvidencePattern = regexp.MustCompile(`^dry-run:sha256:[0-9a-f]{64}$`)
)

func Build(opts Options) Result {
	legacyPath := opts.LegacyMarkerPath
	if legacyPath == "" && opts.StatePath != "" {
		legacyPath = filepath.Join(filepath.Dir(opts.StatePath), filepath.Base(install.KABAdoptionMarkerRelativePath))
	}
	result := Result{
		OK:                           false,
		Command:                      Command,
		Mode:                         "state_validate",
		CLIVersion:                   CLIVersion,
		DryRun:                       opts.DryRun,
		TargetProfile:                opts.Profile,
		ProjectID:                    opts.Project,
		YAMLStatePath:                opts.StatePath,
		LegacyMarkerPath:             legacyPath,
		StateSource:                  "fail_closed",
		WriteTargetAfterApprovedSync: "yaml_state_path",
		EffectiveStageClaim: EffectiveStageClaim{
			Numeric:                  1,
			Canonical:                install.KABStage1Canonical,
			Source:                   "fail_closed",
			KABExecutionClaimAllowed: false,
			FailClosedToStage1:       true,
		},
		Validation: Validation{Diagnostics: []discovery.Diagnostic{}},
	}

	result.ReadSurfaces.LegacyMarker = readLegacyMarker(legacyPath)
	result.ReadSurfaces.YAML = ReadSurface{State: "missing", Path: opts.StatePath}

	if opts.StatePath == "" {
		addDiag(&result, "error", "state_path_required", "sync-project-kas requires --state <path>.")
		result.NextAction = "Provide the project kas-project-state.yaml path and rerun with --dry-run."
		return result
	}
	data, err := os.ReadFile(opts.StatePath)
	if err != nil {
		result.ReadSurfaces.YAML.State = "missing"
		if !os.IsNotExist(err) {
			result.ReadSurfaces.YAML.State = "unreadable"
		}
		result.ReadSurfaces.YAML.Message = err.Error()
		addDiag(&result, "error", "state_file_missing", "kas-project-state.yaml is missing or unreadable; legacy marker read is reporting-only and cannot validate YAML state.")
		result.StateSource = stateSourceForInvalidYAML(result.ReadSurfaces.LegacyMarker.State)
		result.NextAction = "Create or repair kas-project-state.yaml before any project KAS sync; until then only Stage 1 static claims are allowed."
		return result
	}
	sum := sha256.Sum256(data)
	result.ReadSurfaces.YAML = ReadSurface{State: "read", Path: opts.StatePath, SHA256: hex.EncodeToString(sum[:])}

	parsed, parseDiagnostics := parseYAMLSubset(string(data))
	for _, diagnostic := range parseDiagnostics {
		result.Validation.Diagnostics = append(result.Validation.Diagnostics, diagnostic)
	}
	if len(parseDiagnostics) == 0 {
		validateState(&result, parsed, opts)
	}
	if hasErrors(result.Validation.Diagnostics) {
		result.ReadSurfaces.YAML.State = "invalid"
		result.StateSource = stateSourceForInvalidYAML(result.ReadSurfaces.LegacyMarker.State)
		result.NextAction = "Repair kas-project-state.yaml diagnostics before project KAS sync; legacy marker compatibility cannot upgrade invalid YAML."
		return result
	}

	result.OK = true
	result.ReadSurfaces.YAML.State = "valid"
	result.StateSource = "yaml"
	result.Validation.SchemaVersion = parsed.version
	result.Validation.PackBaselineCount = len(parsed.packBaselines)
	result.EffectiveStageClaim = EffectiveStageClaim{
		Numeric:                  mustAtoi(parsed.stage["numeric"]),
		Canonical:                parsed.stage["canonical"],
		Source:                   "yaml",
		KABExecutionClaimAllowed: false,
		FailClosedToStage1:       false,
	}
	result.NextAction = "State is valid for KASUPD-003 dry-run classification; no files were written."
	classifyDryRun(&result, parsed, opts)
	return result
}

func RenderHuman(result Result) string {
	status := "검증 실패"
	if result.OK {
		status = "검증 완료"
	}
	lines := []string{
		fmt.Sprintf("상태: %s — project %s / profile %s KAS state dry-run read.", status, result.ProjectID, result.TargetProfile),
		fmt.Sprintf("YAML: %s (%s)", result.ReadSurfaces.YAML.State, result.YAMLStatePath),
		fmt.Sprintf("legacy marker: %s (%s)", result.ReadSurfaces.LegacyMarker.State, result.LegacyMarkerPath),
		fmt.Sprintf("effective stage: %d %s (source %s, KAB execution claim allowed: %t)", result.EffectiveStageClaim.Numeric, result.EffectiveStageClaim.Canonical, result.EffectiveStageClaim.Source, result.EffectiveStageClaim.KABExecutionClaimAllowed),
		fmt.Sprintf("write target after approved sync: %s", result.WriteTargetAfterApprovedSync),
	}
	if result.SourceRepo != nil {
		commit := "unknown"
		if result.SourceRepo.GitCommit != nil {
			commit = *result.SourceRepo.GitCommit
		}
		dirty := "unknown"
		if result.SourceRepo.Dirty != nil {
			dirty = strconv.FormatBool(*result.SourceRepo.Dirty)
		}
		lines = append(lines, fmt.Sprintf("source repo: %s @ %s (dirty: %s)", result.SourceRepo.Path, commit, dirty))
	}
	if result.ProjectRoot != nil {
		lines = append(lines, fmt.Sprintf("project root: %s (%s, %s)", result.ProjectRoot.Path, result.ProjectRoot.Resolution, result.ProjectRoot.State))
	}
	if result.Summary != nil {
		lines = append(lines, fmt.Sprintf("classification counts: auto_copy_candidate=%d, local_only=%d, semantic_merge_required=%d, new_upstream_candidate=%d, removed_or_renamed_upstream=%d, fail_closed_conflict=%d, no_action=%d",
			result.Summary.CountsByClassification["auto_copy_candidate"],
			result.Summary.CountsByClassification["local_only"],
			result.Summary.CountsByClassification["semantic_merge_required"],
			result.Summary.CountsByClassification["new_upstream_candidate"],
			result.Summary.CountsByClassification["removed_or_renamed_upstream"],
			result.Summary.CountsByClassification["fail_closed_conflict"],
			result.Summary.NoActionCount,
		))
		lines = append(lines, fmt.Sprintf("semantic-port packets: %d; writes: %d", result.Summary.SemanticPortPacketCount, result.Summary.WriteCount))
	}
	for _, diagnostic := range result.Validation.Diagnostics {
		lines = append(lines, fmt.Sprintf("%s: %s — %s", diagnostic.Level, diagnostic.Code, diagnostic.Message))
	}
	for _, classification := range result.Classifications {
		for _, diagnostic := range classification.Diagnostics {
			lines = append(lines, fmt.Sprintf("%s: %s — %s", diagnostic.Level, diagnostic.Code, diagnostic.Message))
		}
	}
	lines = append(lines, "다음: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func classifyDryRun(result *Result, state stateFile, opts Options) {
	result.Mode = "dry_run_classification"
	result.Summary = &Summary{
		TotalMappings:          len(state.packBaselines),
		CountsByClassification: emptyClassificationCounts(),
		WriteCount:             0,
	}
	dirtyRecorded := state.upstream["dirty"] == "true"
	result.BaselineRepo = &BaselineRepo{
		GitCommit:        state.upstream["commit"],
		DirtyRecorded:    &dirtyRecorded,
		BaselineVerified: true,
	}

	repoPath, err := resolveSourceRepo(opts.RepoPath)
	if err != nil {
		addDiag(result, "error", "source_repo_unresolved", err.Error())
		addSyntheticConflict(result, "source_repo_unresolved", "cannot resolve current upstream KAS repo")
		finishClassification(result)
		return
	}
	info := discovery.SourceRepoInfo(repoPath)
	result.SourceRepo = &info

	root := resolveProjectRoot(opts.ProjectRoot, opts.StatePath, opts.Project)
	result.ProjectRoot = &root
	if root.State != "resolved" {
		addSyntheticConflict(result, "project_root_"+root.State, "cannot safely resolve project KAS root")
		finishClassification(result)
		return
	}

	if info.Dirty != nil && *info.Dirty {
		addDiag(result, "error", "source_repo_dirty_requires_review", "current upstream source repo is dirty; KASUPD-003 has no approval to classify against dirty source")
		for i, baseline := range sortedBaselines(state.packBaselines) {
			result.Classifications = append(result.Classifications, conflictForBaseline(i+1, baseline, []discovery.Diagnostic{
				diag("error", "source_repo_dirty_requires_review", "current upstream source repo is dirty and requires explicit review before dry-run classification"),
			}))
		}
		finishClassification(result)
		return
	}

	currentPacks, err := discovery.DiscoverSourcePacks(repoPath)
	if err != nil {
		addDiag(result, "error", "source_repo_unreadable", err.Error())
		addSyntheticConflict(result, "source_repo_unreadable", "cannot read current upstream KAS packs")
		finishClassification(result)
		return
	}
	currentByID := map[string]discovery.SourcePack{}
	for _, pack := range currentPacks {
		currentByID[pack.PackID] = pack
	}

	seen := map[string]bool{}
	for i, baseline := range sortedBaselines(state.packBaselines) {
		id := fmt.Sprintf("kas-sync-item-%04d", i+1)
		classification, unchanged := classifyBaseline(id, repoPath, root.Path, state, baseline, currentByID)
		seen[baseline["upstream_pack"]] = true
		if unchanged != nil {
			result.UnchangedMappings = append(result.UnchangedMappings, *unchanged)
			result.Summary.NoActionCount++
			continue
		}
		result.Classifications = append(result.Classifications, classification)
	}
	sort.Slice(currentPacks, func(i, j int) bool { return currentPacks[i].PackID < currentPacks[j].PackID })
	for _, pack := range currentPacks {
		if seen[pack.PackID] {
			continue
		}
		id := fmt.Sprintf("kas-sync-item-%04d", len(result.Classifications)+len(result.UnchangedMappings)+1)
		classification := Classification{
			ID:                   id,
			UpstreamPack:         pack.PackID,
			Classification:       "new_upstream_candidate",
			Basis:                []string{"current_upstream_pack_not_in_state_baselines"},
			Paths:                map[string]string{"current_upstream_path": pack.SourcePath},
			Checksums:            map[string]string{"current_source_checksum": ensureChecksumPrefix(pack.Checksum)},
			RequiresSemanticPort: true,
			Diagnostics:          []discovery.Diagnostic{},
		}
		result.Classifications = append(result.Classifications, classification)
	}
	for _, classification := range result.Classifications {
		if classification.RequiresSemanticPort {
			result.SemanticPortPackets = append(result.SemanticPortPackets, buildSemanticPacket(result, state, classification))
		}
	}
	finishClassification(result)
}

func classifyBaseline(id string, repoPath string, projectRoot string, state stateFile, baseline map[string]string, currentByID map[string]discovery.SourcePack) (Classification, *UnchangedMapping) {
	upstreamPack := baseline["upstream_pack"]
	projectSkill := baseline["project_skill"]
	classification := Classification{
		ID:             id,
		UpstreamPack:   upstreamPack,
		ProjectSkill:   projectSkill,
		Classification: "fail_closed_conflict",
		Basis:          []string{"project_skill_mapping_exists"},
		Paths: map[string]string{
			"baseline_upstream_path": packPathFromID(upstreamPack),
			"current_upstream_path":  packPathFromID(upstreamPack),
		},
		Checksums: map[string]string{
			"recorded_source_checksum":  baseline["source_checksum"],
			"recorded_project_checksum": baseline["project_checksum"],
		},
		Diagnostics: []discovery.Diagnostic{},
	}
	baselineChecksum, err := computeGitPackChecksum(repoPath, state.upstream["commit"], packPathFromID(upstreamPack))
	if err != nil {
		classification.Diagnostics = append(classification.Diagnostics, diag("error", "baseline_commit_unreadable", err.Error()))
		return classification, nil
	}
	classification.Checksums["computed_baseline_source_checksum"] = baselineChecksum
	if baselineChecksum != baseline["source_checksum"] {
		classification.Diagnostics = append(classification.Diagnostics, diag("error", "baseline_source_checksum_mismatch", "recorded source checksum does not match the pack at upstream_kas.commit"))
		return classification, nil
	}
	current, currentExists := currentByID[upstreamPack]
	if !currentExists {
		classification.Classification = "removed_or_renamed_upstream"
		classification.Basis = append(classification.Basis, "current_upstream_pack_missing")
		classification.Diagnostics = []discovery.Diagnostic{}
		return classification, nil
	}
	classification.Paths["current_upstream_path"] = current.SourcePath
	currentSourceChecksum := ensureChecksumPrefix(current.Checksum)
	classification.Checksums["current_source_checksum"] = currentSourceChecksum

	projectPath, projectRel, err := resolveProjectSkillPath(projectRoot, state.project["id"], projectSkill)
	if err != nil {
		code := "project_skill_missing"
		if strings.Contains(err.Error(), "ambiguous") {
			code = "project_skill_path_ambiguous"
		} else if strings.Contains(err.Error(), "escape") || strings.Contains(err.Error(), "invalid") {
			code = "project_skill_path_invalid"
		}
		classification.Diagnostics = append(classification.Diagnostics, diag("error", code, err.Error()))
		return classification, nil
	}
	classification.Paths["project_skill_path"] = projectRel
	projectChecksum, err := discovery.ComputePackChecksum(projectPath)
	if err != nil {
		classification.Diagnostics = append(classification.Diagnostics, diag("error", "project_skill_unreadable", err.Error()))
		return classification, nil
	}
	currentProjectChecksum := ensureChecksumPrefix(projectChecksum)
	classification.Checksums["current_project_checksum"] = currentProjectChecksum

	upstreamChanged := currentSourceChecksum != baseline["source_checksum"]
	localChanged := currentProjectChecksum != baseline["project_checksum"]
	switch {
	case upstreamChanged && localChanged:
		classification.Classification = "semantic_merge_required"
		classification.Basis = append(classification.Basis, "local_changed_since_baseline", "upstream_changed_since_baseline")
		classification.RequiresSemanticPort = true
	case upstreamChanged:
		classification.Classification = "auto_copy_candidate"
		classification.Basis = append(classification.Basis, "local_unchanged_since_baseline", "upstream_changed_since_baseline")
	case localChanged:
		classification.Classification = "local_only"
		classification.Basis = append(classification.Basis, "local_changed_since_baseline", "upstream_unchanged_since_baseline")
	default:
		unchanged := UnchangedMapping{
			ID:           id,
			UpstreamPack: upstreamPack,
			ProjectSkill: projectSkill,
			Paths:        classification.Paths,
			Checksums:    classification.Checksums,
		}
		return Classification{}, &unchanged
	}
	classification.Diagnostics = []discovery.Diagnostic{}
	return classification, nil
}

func finishClassification(result *Result) {
	if result.Summary == nil {
		return
	}
	for _, classification := range result.Classifications {
		if _, ok := result.Summary.CountsByClassification[classification.Classification]; ok {
			result.Summary.CountsByClassification[classification.Classification]++
		}
		if classification.Classification == "fail_closed_conflict" {
			result.OK = false
			result.BaselineRepo.BaselineVerified = false
		}
	}
	result.Summary.SemanticPortPacketCount = len(result.SemanticPortPackets)
	if result.OK {
		result.NextAction = "Review dry-run classifications and semantic-port packets; no files were written."
	} else {
		result.NextAction = "Resolve fail-closed diagnostics before any project KAS sync; no files were written."
	}
}

func emptyClassificationCounts() map[string]int {
	return map[string]int{
		"auto_copy_candidate":         0,
		"local_only":                  0,
		"semantic_merge_required":     0,
		"new_upstream_candidate":      0,
		"removed_or_renamed_upstream": 0,
		"fail_closed_conflict":        0,
	}
}

func addSyntheticConflict(result *Result, code string, message string) {
	if result.Summary == nil {
		result.Summary = &Summary{CountsByClassification: emptyClassificationCounts()}
	}
	result.Classifications = append(result.Classifications, Classification{
		ID:             "kas-sync-item-0001",
		Classification: "fail_closed_conflict",
		Basis:          []string{code},
		Paths:          map[string]string{},
		Checksums:      map[string]string{},
		Diagnostics:    []discovery.Diagnostic{diag("error", code, message)},
	})
}

func conflictForBaseline(n int, baseline map[string]string, diagnostics []discovery.Diagnostic) Classification {
	return Classification{
		ID:             fmt.Sprintf("kas-sync-item-%04d", n),
		UpstreamPack:   baseline["upstream_pack"],
		ProjectSkill:   baseline["project_skill"],
		Classification: "fail_closed_conflict",
		Basis:          []string{"project_skill_mapping_exists"},
		Paths: map[string]string{
			"baseline_upstream_path": packPathFromID(baseline["upstream_pack"]),
			"current_upstream_path":  packPathFromID(baseline["upstream_pack"]),
		},
		Checksums: map[string]string{
			"recorded_source_checksum":  baseline["source_checksum"],
			"recorded_project_checksum": baseline["project_checksum"],
		},
		Diagnostics: diagnostics,
	}
}

func sortedBaselines(baselines []map[string]string) []map[string]string {
	out := append([]map[string]string(nil), baselines...)
	sort.Slice(out, func(i, j int) bool {
		if out[i]["upstream_pack"] == out[j]["upstream_pack"] {
			return out[i]["project_skill"] < out[j]["project_skill"]
		}
		return out[i]["upstream_pack"] < out[j]["upstream_pack"]
	})
	return out
}

func buildSemanticPacket(result *Result, state stateFile, classification Classification) SemanticPortPacket {
	content := strings.Join([]string{
		"# Semantic Port Packet",
		"",
		"packet_scope: KASUPD-003 dry-run evidence only",
		"no_write_statement: sync-project-kas --dry-run wrote no project KAS files, profiles, manifests, KAH state, KAB state, or packet files.",
		fmt.Sprintf("state_path: %s", result.YAMLStatePath),
		fmt.Sprintf("state_sha256: %s", result.ReadSurfaces.YAML.SHA256),
		fmt.Sprintf("classification_id: %s", classification.ID),
		fmt.Sprintf("classification: %s", classification.Classification),
		fmt.Sprintf("upstream_pack: %s", classification.UpstreamPack),
		fmt.Sprintf("project_skill: %s", classification.ProjectSkill),
		fmt.Sprintf("baseline_commit: %s", state.upstream["commit"]),
		fmt.Sprintf("paths_json: %s", mustJSON(classification.Paths)),
		fmt.Sprintf("checksums_json: %s", mustJSON(classification.Checksums)),
		"",
		"Preservation constraints:",
		"- Preserve project authority, roadmap IDs, test commands, role labels, and selected KAB adoption stage.",
		"- Do not change auth, tokens, secrets, gateway/provider/model config, KAB runtime state, or installed Hermes profiles.",
		"- Treat this packet as review input only; approved sync/apply/write-capable behavior is out of scope.",
		"",
		"Dry-run evidence:",
		fmt.Sprintf("- basis: %s", strings.Join(classification.Basis, ", ")),
	}, "\n")
	sum := sha256.Sum256([]byte(content))
	packetID := "semantic-port-" + classification.ID
	return SemanticPortPacket{
		PacketID:                packetID,
		ClassificationID:        classification.ID,
		RecommendedArtifactPath: filepath.ToSlash(filepath.Join(".kas", "dry-runs", packetID+".md")),
		ContentSHA256:           "sha256:" + hex.EncodeToString(sum[:]),
		Content:                 content,
	}
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func resolveSourceRepo(value string) (string, error) {
	if value == "" {
		return discovery.FindSourceRepo("")
	}
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	return discovery.FindSourceRepo(abs)
}

func resolveProjectRoot(explicit string, statePath string, project string) ProjectRoot {
	if explicit != "" {
		root, err := cleanExistingDirInside("", explicit)
		if err != nil {
			return ProjectRoot{Path: explicit, Resolution: "explicit", State: "missing"}
		}
		return ProjectRoot{Path: root, Resolution: "explicit", State: "resolved"}
	}
	abs, err := filepath.Abs(statePath)
	if err != nil {
		return ProjectRoot{Path: statePath, Resolution: "inferred_from_state", State: "ambiguous"}
	}
	eval, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return ProjectRoot{Path: abs, Resolution: "inferred_from_state", State: "missing"}
	}
	parts := strings.Split(filepath.ToSlash(eval), "/")
	suffix := []string{"skills", project, project + "-kas", "references", "kas-project-state.yaml"}
	if len(parts) < len(suffix)+1 {
		return ProjectRoot{Path: eval, Resolution: "inferred_from_state", State: "ambiguous"}
	}
	start := len(parts) - len(suffix)
	for i, want := range suffix {
		if parts[start+i] != want {
			return ProjectRoot{Path: eval, Resolution: "inferred_from_state", State: "ambiguous"}
		}
	}
	rootSlash := strings.Join(parts[:start], "/")
	if rootSlash == "" {
		rootSlash = "/"
	}
	root := filepath.FromSlash(rootSlash)
	return ProjectRoot{Path: root, Resolution: "inferred_from_state", State: "resolved"}
}

func resolveProjectSkillPath(root string, project string, projectSkill string) (string, string, error) {
	if discovery.IsInvalidRelativePath(projectSkill) {
		return "", "", fmt.Errorf("invalid project_skill path %q", projectSkill)
	}
	candidates := []string{
		filepath.Join(root, "skills", project, projectSkill),
		filepath.Join(root, "skills", projectSkill),
	}
	found := []string{}
	for _, candidate := range candidates {
		resolved, err := cleanExistingDirInside(root, candidate)
		if err != nil {
			continue
		}
		if !isRegularFile(filepath.Join(resolved, "SKILL.md")) {
			continue
		}
		found = append(found, resolved)
	}
	if len(found) == 0 {
		return "", "", fmt.Errorf("project skill %q missing under expected mapped paths", projectSkill)
	}
	if len(found) > 1 {
		return "", "", fmt.Errorf("project skill path ambiguous for %q", projectSkill)
	}
	rel, err := filepath.Rel(root, found[0])
	if err != nil {
		return "", "", err
	}
	return found[0], filepath.ToSlash(rel), nil
}

func cleanExistingDirInside(root string, path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("missing directory %s", path)
	}
	if root != "" {
		rootResolved, err := filepath.EvalSymlinks(root)
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(rootResolved, resolved)
		if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			return "", fmt.Errorf("path %s would escape project root", path)
		}
	}
	return resolved, nil
}

func computeGitPackChecksum(repo string, commit string, packPath string) (string, error) {
	if discovery.IsInvalidRelativePath(packPath) {
		return "", fmt.Errorf("invalid baseline upstream path %q", packPath)
	}
	type entry struct {
		Path   string `json:"path"`
		Bytes  int    `json:"bytes"`
		Mode   string `json:"mode"`
		SHA256 string `json:"sha256"`
	}
	out, err := gitOutput(repo, "ls-tree", "-r", "-z", "-l", commit, "--", packPath)
	if err != nil {
		return "", err
	}
	entries := []entry{}
	for _, record := range strings.Split(string(out), "\x00") {
		if record == "" {
			continue
		}
		meta, path, ok := strings.Cut(record, "\t")
		if !ok {
			return "", fmt.Errorf("unexpected git ls-tree record for %s", packPath)
		}
		fields := strings.Fields(meta)
		if len(fields) < 4 || fields[1] != "blob" {
			continue
		}
		rel, err := filepath.Rel(filepath.FromSlash(packPath), filepath.FromSlash(path))
		if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			return "", fmt.Errorf("baseline path %s escapes pack %s", path, packPath)
		}
		rel = filepath.ToSlash(rel)
		if excludedPackPath(rel) {
			continue
		}
		data, err := gitOutput(repo, "show", commit+":"+path)
		if err != nil {
			return "", err
		}
		size, err := strconv.Atoi(fields[3])
		if err != nil {
			size = len(data)
		}
		sum := sha256.Sum256(data)
		entries = append(entries, entry{
			Path:   rel,
			Bytes:  size,
			Mode:   gitModeString(fields[0]),
			SHA256: hex.EncodeToString(sum[:]),
		})
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("baseline pack %s not found at commit %s", packPath, commit)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	payload, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func gitOutput(repo string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(string(exit.Stderr)))
		}
		return nil, err
	}
	return out, nil
}

func packPathFromID(packID string) string {
	if strings.Contains(packID, "/") {
		return "skills/" + packID
	}
	return "skills/" + packID
}

func ensureChecksumPrefix(value string) string {
	if strings.HasPrefix(value, "sha256:") {
		return value
	}
	return "sha256:" + value
}

func gitModeString(mode string) string {
	switch mode {
	case "100755":
		return "0755"
	default:
		return "0644"
	}
}

func excludedPackPath(relativePath string) bool {
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

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func readLegacyMarker(path string) ReadSurface {
	surface := ReadSurface{State: "not_requested", Path: path}
	if path == "" {
		return surface
	}
	data, err := os.ReadFile(path)
	if err != nil {
		surface.State = "missing"
		if !os.IsNotExist(err) {
			surface.State = "unreadable"
		}
		surface.Message = err.Error()
		return surface
	}
	sum := sha256.Sum256(data)
	surface.SHA256 = hex.EncodeToString(sum[:])
	if _, ok := install.ParseKABAdoptionStageMarker(data); ok {
		surface.State = "present_compatible"
	} else {
		surface.State = "present_invalid"
	}
	return surface
}

func stateSourceForInvalidYAML(legacyState string) string {
	if legacyState == "present_compatible" {
		return "legacy_marker_only"
	}
	return "fail_closed"
}

func parseYAMLSubset(data string) (stateFile, []discovery.Diagnostic) {
	state := stateFile{
		project:         map[string]string{},
		stage:           map[string]string{},
		upstream:        map[string]string{},
		overlayPolicy:   map[string]string{},
		updatePolicy:    map[string][]string{},
		updateScalars:   map[string]string{},
		evidencePosture: map[string]string{},
	}
	diagnostics := []discovery.Diagnostic{}
	section := ""
	listKey := ""
	currentPack := -1
	lines := strings.Split(data, "\n")
	for i, raw := range lines {
		lineNo := i + 1
		if strings.TrimSpace(raw) == "" || strings.HasPrefix(strings.TrimSpace(raw), "#") {
			continue
		}
		if strings.Contains(raw, "\t") {
			diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d uses tabs; only two-space indentation is supported", lineNo)))
			continue
		}
		if strings.Contains(raw, "&") || strings.Contains(raw, "*") || strings.Contains(raw, "|") || strings.Contains(raw, ">") {
			diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d uses unsupported YAML features; only the documented scalar/list subset is supported", lineNo)))
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		if indent%2 != 0 {
			diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d has unsupported indentation", lineNo)))
			continue
		}
		line := strings.TrimSpace(raw)
		if indent == 0 {
			section = ""
			listKey = ""
			currentPack = -1
			key, value, ok := splitYAMLKeyValue(line)
			if !ok {
				diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d is not a key/value entry", lineNo)))
				continue
			}
			if value == "" {
				section = key
			} else if key == "version" {
				state.version = unquote(value)
			}
			continue
		}
		switch section {
		case "project":
			parseScalarMapLine(&diagnostics, state.project, line, lineNo)
		case "kab_adoption_stage":
			parseScalarMapLine(&diagnostics, state.stage, line, lineNo)
		case "upstream_kas":
			parseScalarMapLine(&diagnostics, state.upstream, line, lineNo)
		case "overlay_policy":
			parseScalarMapLine(&diagnostics, state.overlayPolicy, line, lineNo)
		case "evidence_posture":
			parseScalarMapLine(&diagnostics, state.evidencePosture, line, lineNo)
		case "pack_baselines":
			if indent == 2 && strings.HasPrefix(line, "- ") {
				key, value, ok := splitYAMLKeyValue(strings.TrimSpace(strings.TrimPrefix(line, "- ")))
				if !ok || value == "" {
					diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d has unsupported pack_baselines list syntax", lineNo)))
					continue
				}
				state.packBaselines = append(state.packBaselines, map[string]string{key: unquote(value)})
				currentPack = len(state.packBaselines) - 1
			} else if indent == 4 && currentPack >= 0 {
				parseScalarMapLine(&diagnostics, state.packBaselines[currentPack], line, lineNo)
			} else {
				diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d has unsupported pack_baselines indentation", lineNo)))
			}
		case "update_policy":
			if indent == 2 {
				key, value, ok := splitYAMLKeyValue(line)
				if !ok {
					diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d is not a key/value entry", lineNo)))
					continue
				}
				listKey = ""
				if value == "" {
					listKey = key
					if state.updatePolicy[listKey] == nil {
						state.updatePolicy[listKey] = []string{}
					}
				} else {
					state.updateScalars[key] = unquote(value)
				}
			} else if indent == 4 && strings.HasPrefix(line, "- ") && listKey != "" {
				state.updatePolicy[listKey] = append(state.updatePolicy[listKey], unquote(strings.TrimSpace(strings.TrimPrefix(line, "- "))))
			} else {
				diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d has unsupported update_policy syntax", lineNo)))
			}
		default:
			diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d is under unsupported section %q", lineNo, section)))
		}
	}
	if hasSuspiciousInput(data) {
		diagnostics = append(diagnostics, diag("error", "auth_token_gateway_or_provider_mutation_detected", "state input contains secret/auth/token/gateway/provider/runtime-state-like material"))
	}
	return state, diagnostics
}

func parseScalarMapLine(diagnostics *[]discovery.Diagnostic, target map[string]string, line string, lineNo int) {
	key, value, ok := splitYAMLKeyValue(line)
	if !ok || value == "" {
		*diagnostics = append(*diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d is not a supported scalar key/value entry", lineNo)))
		return
	}
	target[key] = unquote(value)
}

func splitYAMLKeyValue(line string) (string, string, bool) {
	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" || strings.Contains(key, " ") {
		return "", "", false
	}
	return key, value, true
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func validateState(result *Result, state stateFile, opts Options) {
	if state.version != SchemaVersion {
		addDiag(result, "error", "state_schema_invalid", fmt.Sprintf("version must be %q", SchemaVersion))
	}
	requireMap(result, "project", state.project, []string{"id", "repo", "kas_suite", "profile"})
	requireMap(result, "kab_adoption_stage", state.stage, []string{"numeric", "canonical", "selection_source", "selected_at", "approval_evidence", "stage2_activation"})
	requireMap(result, "upstream_kas", state.upstream, []string{"repo", "remote", "commit", "dirty", "synced_at", "sync_task"})
	requireMap(result, "overlay_policy", state.overlayPolicy, []string{"local_overlay_allowed", "preserve_project_authority", "preserve_project_roadmap_ids", "preserve_project_test_commands", "preserve_role_labels", "overwrite_mode"})
	requireMap(result, "evidence_posture", state.evidencePosture, []string{"not_kab_runtime_evidence", "not_stage2_activation_by_itself", "missing_or_unreadable_fails_to_stage1_claims"})
	if opts.Profile != "" && state.project["profile"] != "" && opts.Profile != state.project["profile"] {
		addDiag(result, "error", "profile_mismatch", fmt.Sprintf("--profile %q does not match state project.profile %q", opts.Profile, state.project["profile"]))
	}
	if opts.Project != "" && state.project["id"] != "" && opts.Project != state.project["id"] {
		addDiag(result, "error", "project_mismatch", fmt.Sprintf("--project %q does not match state project.id %q", opts.Project, state.project["id"]))
	}
	validateStage(result, state.stage)
	validateUpstream(result, state.upstream)
	validatePackBaselines(result, state.packBaselines)
	validateOverlay(result, state.overlayPolicy)
	validateUpdatePolicy(result, state.updatePolicy, state.updateScalars)
	validateEvidencePosture(result, state.evidencePosture)
}

func validateStage(result *Result, stage map[string]string) {
	numeric := stage["numeric"]
	canonical := stage["canonical"]
	switch numeric {
	case "1":
		if canonical != install.KABStage1Canonical {
			addDiag(result, "error", "stage_schema_invalid", "Stage 1 numeric value must use the Stage 1 canonical value")
		}
	case "2":
		if canonical != install.KABStage2Canonical {
			addDiag(result, "error", "stage_schema_invalid", "Stage 2 numeric value must use the Stage 2 canonical value")
		}
	case "3":
		addDiag(result, "error", "stage_unsupported", "Stage 3 is reserved and unsupported for project KAS sync state")
	default:
		addDiag(result, "error", "stage_schema_invalid", "kab_adoption_stage.numeric must be 1 or 2")
	}
	if canonical == stage3Canonical {
		addDiag(result, "error", "stage_unsupported", "Stage 3 canonical value is reserved and unsupported")
	}
	if stage["stage2_activation"] != "false" {
		addDiag(result, "error", "stage2_activation_rejected", "stage2_activation must be false; YAML state is not Stage 2 activation by itself")
	}
}

func validateUpstream(result *Result, upstream map[string]string) {
	if commit := upstream["commit"]; commit == "" || !shaPattern.MatchString(commit) {
		addDiag(result, "error", "upstream_commit_unknown", "upstream_kas.commit must be a full 40-character git SHA")
	}
	if dirty := upstream["dirty"]; dirty != "true" && dirty != "false" {
		addDiag(result, "error", "state_schema_invalid", "upstream_kas.dirty must be a boolean")
	}
}

func validatePackBaselines(result *Result, baselines []map[string]string) {
	if len(baselines) == 0 {
		addDiag(result, "error", "checksum_mismatch_without_baseline", "pack_baselines must include at least one checksum baseline")
		return
	}
	required := []string{"upstream_pack", "project_skill", "source_checksum", "project_checksum", "merge_mode"}
	for i, baseline := range baselines {
		for _, key := range required {
			if strings.TrimSpace(baseline[key]) == "" {
				addDiag(result, "error", "state_schema_invalid", fmt.Sprintf("pack_baselines[%d].%s is required", i, key))
			}
		}
		if baseline["source_checksum"] != "" && !checksumPattern.MatchString(baseline["source_checksum"]) {
			addDiag(result, "error", "checksum_mismatch_without_baseline", fmt.Sprintf("pack_baselines[%d].source_checksum must be sha256:<64 hex>", i))
		}
		if baseline["project_checksum"] != "" && !checksumPattern.MatchString(baseline["project_checksum"]) {
			addDiag(result, "error", "checksum_mismatch_without_baseline", fmt.Sprintf("pack_baselines[%d].project_checksum must be sha256:<64 hex>", i))
		}
		if baseline["merge_mode"] != "" && baseline["merge_mode"] != "semantic_port" {
			addDiag(result, "error", "state_schema_invalid", fmt.Sprintf("pack_baselines[%d].merge_mode must be semantic_port for KASUPD-002", i))
		}
	}
}

func validateOverlay(result *Result, overlay map[string]string) {
	for _, key := range []string{"preserve_project_authority", "preserve_project_roadmap_ids", "preserve_project_test_commands", "preserve_role_labels"} {
		if overlay[key] != "true" {
			addDiag(result, "error", "overlay_policy_invalid", key+" must be true")
		}
	}
	if overlay["local_overlay_allowed"] != "true" && overlay["local_overlay_allowed"] != "false" {
		addDiag(result, "error", "overlay_policy_invalid", "local_overlay_allowed must be a boolean")
	}
	if overlay["overwrite_mode"] != "never_without_review" {
		addDiag(result, "error", "overlay_policy_invalid", "overwrite_mode must be never_without_review")
	}
}

func validateUpdatePolicy(result *Result, lists map[string][]string, scalars map[string]string) {
	if scalars["default_mode"] != "dry_run_then_semantic_merge" {
		addDiag(result, "error", "state_schema_invalid", "update_policy.default_mode must be dry_run_then_semantic_merge")
	}
	requiredFailClosed := []string{
		"state_file_missing",
		"state_schema_invalid",
		"stage_unsupported",
		"upstream_commit_unknown",
		"checksum_mismatch_without_baseline",
		"auth_token_gateway_or_provider_mutation_detected",
	}
	have := map[string]bool{}
	for _, value := range lists["fail_closed_when"] {
		have[value] = true
	}
	for _, value := range requiredFailClosed {
		if !have[value] {
			addDiag(result, "error", "state_schema_invalid", "update_policy.fail_closed_when must include "+value)
		}
	}
	for _, key := range []string{"auto_apply_when", "require_llm_merge_when", "fail_closed_when"} {
		if len(lists[key]) == 0 {
			addDiag(result, "error", "state_schema_invalid", "update_policy."+key+" must be non-empty")
		}
	}
}

func validateEvidencePosture(result *Result, posture map[string]string) {
	for _, key := range []string{"not_kab_runtime_evidence", "not_stage2_activation_by_itself", "missing_or_unreadable_fails_to_stage1_claims"} {
		if posture[key] != "true" {
			addDiag(result, "error", "evidence_posture_invalid", key+" must be true")
		}
	}
}

func requireMap(result *Result, section string, values map[string]string, keys []string) {
	for _, key := range keys {
		if strings.TrimSpace(values[key]) == "" {
			addDiag(result, "error", "state_schema_invalid", section+"."+key+" is required")
		}
	}
}

func hasSuspiciousInput(data string) bool {
	allowed := map[string]bool{
		"auth_token_gateway_or_provider_mutation_detected": true,
	}
	suspicious := []string{"token", "secret", "credential", "provider_key", "api_key", "password", "gateway", "session_id", "session_state", "runtime_state"}
	lines := strings.Split(data, "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.FieldsFunc(strings.ToLower(line), func(r rune) bool {
			return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_'
		})
		for _, part := range parts {
			if allowed[part] {
				continue
			}
			for _, word := range suspicious {
				if part == word || strings.Contains(part, word) {
					return true
				}
			}
		}
	}
	return false
}

func addDiag(result *Result, level string, code string, message string) {
	result.Validation.Diagnostics = append(result.Validation.Diagnostics, diag(level, code, message))
}

func diag(level string, code string, message string) discovery.Diagnostic {
	return discovery.Diagnostic{Level: level, Code: code, Message: message}
}

func hasErrors(diagnostics []discovery.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == "error" {
			return true
		}
	}
	return false
}

func mustAtoi(value string) int {
	n, _ := strconv.Atoi(value)
	return n
}
