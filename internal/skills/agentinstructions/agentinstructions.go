package agentinstructions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ManagedBegin = "<!-- KAS:MANAGED:BEGIN core-behavior -->"
	ManagedEnd   = "<!-- KAS:MANAGED:END core-behavior -->"
	LocalBegin   = "<!-- PROJECT:LOCAL:BEGIN -->"
	LocalEnd     = "<!-- PROJECT:LOCAL:END -->"
)

var managedFiles = []managedFile{
	{Path: "AGENTS.md", TemplatePath: "templates/agent-instructions/AGENTS.md.tmpl"},
	{Path: "CLAUDE.md", TemplatePath: "templates/agent-instructions/CLAUDE.md.tmpl"},
}

type Options struct {
	RepoPath            string
	SourceRepoPath      string
	Project             string
	RepositoryRole      string
	ProjectSuiteID      string
	KABAdoptionStage    string
	UpstreamKASBaseline string
	LocalAuthorityNotes string
	NotApplicable       []string
	NotApplicableReason string
}

type Result struct {
	OK              bool            `json:"ok"`
	Command         string          `json:"command"`
	Mode            string          `json:"mode"`
	RepoPath        string          `json:"repo_path"`
	SourceRepoPath  string          `json:"source_repo_path"`
	Project         string          `json:"project"`
	Summary         Summary         `json:"summary"`
	FilePlans       []FilePlan      `json:"file_plans"`
	ChangedPaths    []ChangedPath   `json:"changed_paths"`
	NoWrite         NoWriteEvidence `json:"no_write"`
	PlanHash        string          `json:"plan_hash"`
	ApprovalRequest ApprovalRequest `json:"approval_request"`
	Approval        *Approval       `json:"approval,omitempty"`
	Diagnostics     []Diagnostic    `json:"diagnostics"`
	NextAction      string          `json:"next_action"`
}

type Summary struct {
	CountsByOutcome map[string]int `json:"counts_by_outcome"`
	BlockedCount    int            `json:"blocked_count"`
	WritableCount   int            `json:"writable_count"`
}

type FilePlan struct {
	Path                string               `json:"path"`
	TemplatePath        string               `json:"template_path,omitempty"`
	Outcome             string               `json:"outcome"`
	PreviousSHA256      string               `json:"previous_sha256,omitempty"`
	NewSHA256           string               `json:"new_sha256,omitempty"`
	RenderedSHA256      string               `json:"rendered_sha256,omitempty"`
	PreservationActions []PreservationAction `json:"preservation_actions"`
	Reason              string               `json:"reason,omitempty"`
}

type ChangedPath struct {
	Path           string `json:"path"`
	Action         string `json:"action"`
	PreviousSHA256 string `json:"previous_sha256,omitempty"`
	NewSHA256      string `json:"new_sha256,omitempty"`
}

type PreservationAction struct {
	Action string `json:"action"`
	Detail string `json:"detail"`
}

type NoWriteEvidence struct {
	Guaranteed                     bool `json:"guaranteed"`
	RepoWriteCount                 int  `json:"repo_write_count"`
	ProfileWriteCount              int  `json:"profile_write_count"`
	KAHStateWriteCount             int  `json:"kah_state_write_count"`
	KABRuntimeMutationCount        int  `json:"kab_runtime_mutation_count"`
	AuthProviderConfigWriteCount   int  `json:"auth_provider_config_write_count"`
	RuntimeProviderModelWriteCount int  `json:"runtime_provider_model_write_count"`
}

type ApprovalRequest struct {
	Required       bool   `json:"required"`
	EvidenceRef    string `json:"evidence_ref,omitempty"`
	PlanHash       string `json:"plan_hash,omitempty"`
	HashDiscipline string `json:"hash_discipline"`
}

type Approval struct {
	EvidenceRef        string `json:"evidence_ref"`
	ApprovedPlanHash   string `json:"approved_plan_hash"`
	MatchedCurrentPlan bool   `json:"matched_current_plan"`
}

type Diagnostic struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

type managedFile struct {
	Path         string
	TemplatePath string
}

func BuildDryRun(opts Options) (Result, error) {
	normalized, err := normalizeOptions(opts)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		OK:             true,
		Command:        "update agent-instructions",
		Mode:           "dry_run",
		RepoPath:       normalized.RepoPath,
		SourceRepoPath: normalized.SourceRepoPath,
		Project:        normalized.Project,
		NoWrite:        noWriteEvidence(true),
	}
	notApplicable := notApplicableSet(normalized.NotApplicable)
	for _, file := range managedFiles {
		plan, diagnostic, err := planFile(normalized, file, notApplicable)
		if err != nil {
			return Result{}, err
		}
		result.FilePlans = append(result.FilePlans, plan)
		if diagnostic != nil {
			result.Diagnostics = append(result.Diagnostics, *diagnostic)
		}
		if plan.Outcome != "no_change" && plan.Outcome != "preserve_project_local_block" {
			result.ChangedPaths = append(result.ChangedPaths, ChangedPath{
				Path:           plan.Path,
				Action:         plan.Outcome,
				PreviousSHA256: plan.PreviousSHA256,
				NewSHA256:      plan.NewSHA256,
			})
		}
		if plan.Outcome == "blocked_unmarked_existing_file" {
			result.OK = false
		}
		if plan.Outcome == "error" {
			result.OK = false
		}
	}
	finalize(&result)
	return result, nil
}

func Apply(opts Options, evidenceRef string) (Result, error) {
	dryRun, err := BuildDryRun(opts)
	if err != nil {
		return Result{}, err
	}
	dryRun.Mode = "apply"
	approvedHash, ok := parseEvidenceRef(evidenceRef)
	dryRun.Approval = &Approval{EvidenceRef: evidenceRef, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: ok && approvedHash == dryRun.PlanHash}
	if !ok {
		dryRun.OK = false
		dryRun.Diagnostics = append(dryRun.Diagnostics, Diagnostic{Level: "error", Code: "approval_evidence_malformed", Message: "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>."})
		dryRun.NextAction = "No files were written. Re-run dry-run and apply only the emitted approval_request.evidence_ref."
		return dryRun, nil
	}
	if approvedHash != dryRun.PlanHash {
		dryRun.OK = false
		dryRun.Diagnostics = append(dryRun.Diagnostics, Diagnostic{Level: "error", Code: "approval_hash_mismatch", Message: "approved hash does not match the current agent-instructions dry-run plan."})
		dryRun.NextAction = "No files were written. Re-run dry-run and approve only the current plan hash."
		return dryRun, nil
	}
	if !dryRun.OK {
		dryRun.Diagnostics = append([]Diagnostic{{Level: "error", Code: "agent_instructions_plan_not_approvable", Message: "current dry-run contains blocked agent-instruction files; no files were written."}}, dryRun.Diagnostics...)
		dryRun.NextAction = "Resolve blocked files or mark them not_applicable through an approved policy, then rerun dry-run."
		return dryRun, nil
	}
	writes := 0
	for _, plan := range dryRun.FilePlans {
		if plan.Outcome != "create" && plan.Outcome != "update_managed_block" {
			continue
		}
		content, err := plannedContentFor(opts, plan.Path)
		if err != nil {
			return applyFailureResult(dryRun, writes, Diagnostic{Level: "error", Code: "planned_content_unavailable", Message: err.Error(), Path: plan.Path}, "No complete apply claim is available. Re-run dry-run after fixing diagnostics."), nil
		}
		target := filepath.Join(dryRun.RepoPath, filepath.FromSlash(plan.Path))
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return applyFailureResult(dryRun, writes, Diagnostic{Level: "error", Code: "agent_instruction_write_failed", Message: err.Error(), Path: plan.Path}, "Apply stopped on a write failure. Inspect file state and restart from a fresh dry-run."), nil
		}
		written, err := os.ReadFile(target)
		if err != nil {
			return applyFailureResult(dryRun, writes+1, Diagnostic{Level: "error", Code: "agent_instruction_write_verification_failed", Message: err.Error(), Path: plan.Path}, "Apply wrote a file but could not verify the result. Inspect file state and restart from a fresh dry-run."), nil
		}
		if sha256String(string(written)) != plan.NewSHA256 {
			return applyFailureResult(dryRun, writes+1, Diagnostic{Level: "error", Code: "agent_instruction_write_verification_failed", Message: "written bytes do not match the approved plan hash.", Path: plan.Path}, "Apply wrote unexpected bytes. Inspect file state and restart from a fresh dry-run."), nil
		}
		writes++
	}
	dryRun.NoWrite = noWriteEvidence(false)
	dryRun.NoWrite.RepoWriteCount = writes
	dryRun.NextAction = "Agent-instruction apply complete. Re-run update agent-instructions --dry-run or project verification as needed."
	return dryRun, nil
}

func RenderHuman(result Result) string {
	state := "pass"
	if !result.OK {
		state = "blocked"
	}
	lines := []string{
		"Status: " + state,
		"Command: update agent-instructions",
		fmt.Sprintf("Plan: create %d, update_managed_block %d, no_change %d, not_applicable %d, blocked_unmarked_existing_file %d.",
			result.Summary.CountsByOutcome["create"],
			result.Summary.CountsByOutcome["update_managed_block"],
			result.Summary.CountsByOutcome["no_change"],
			result.Summary.CountsByOutcome["not_applicable"],
			result.Summary.CountsByOutcome["blocked_unmarked_existing_file"],
		),
		"Plan hash: " + result.PlanHash,
		"Next: " + result.NextAction,
	}
	if result.Summary.CountsByOutcome["preserve_project_local_block"] > 0 {
		lines = append(lines, fmt.Sprintf("Preserved project-local blocks: %d.", result.Summary.CountsByOutcome["preserve_project_local_block"]))
	}
	if result.Summary.CountsByOutcome["error"] > 0 {
		lines = append(lines, fmt.Sprintf("Errors: %d.", result.Summary.CountsByOutcome["error"]))
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, "Diagnostic: "+diagnostic.Code+" "+diagnostic.Message)
	}
	return strings.Join(lines, "\n")
}

func applyFailureResult(result Result, writes int, diagnostic Diagnostic, nextAction string) Result {
	result.OK = false
	result.Diagnostics = append(result.Diagnostics, diagnostic)
	result.NextAction = nextAction
	if writes > 0 {
		result.NoWrite = noWriteEvidence(false)
		result.NoWrite.RepoWriteCount = writes
	}
	return result
}

func normalizeOptions(opts Options) (Options, error) {
	if strings.TrimSpace(opts.RepoPath) == "" {
		return Options{}, fmt.Errorf("repo-path is required")
	}
	repoPath, err := filepath.Abs(opts.RepoPath)
	if err != nil {
		return Options{}, err
	}
	sourceRepo := strings.TrimSpace(opts.SourceRepoPath)
	if sourceRepo == "" {
		sourceRepo = "."
	}
	sourceRepo, err = filepath.Abs(sourceRepo)
	if err != nil {
		return Options{}, err
	}
	opts.RepoPath = filepath.Clean(repoPath)
	opts.SourceRepoPath = filepath.Clean(sourceRepo)
	if opts.Project == "" {
		opts.Project = filepath.Base(opts.RepoPath)
	}
	if opts.RepositoryRole == "" {
		opts.RepositoryRole = "project-specific suite"
	}
	if opts.ProjectSuiteID == "" {
		opts.ProjectSuiteID = opts.Project
	}
	if opts.KABAdoptionStage == "" {
		opts.KABAdoptionStage = "stage1_direct_codex_app_server_baseline"
	}
	if opts.UpstreamKASBaseline == "" {
		opts.UpstreamKASBaseline = "current KAS source"
	}
	if opts.LocalAuthorityNotes == "" {
		opts.LocalAuthorityNotes = "Follow repository SOT and nearest AGENTS.md."
	}
	return opts, nil
}

func planFile(opts Options, file managedFile, notApplicable map[string]bool) (FilePlan, *Diagnostic, error) {
	if notApplicable[file.Path] {
		return FilePlan{Path: file.Path, TemplatePath: file.TemplatePath, Outcome: "not_applicable", Reason: notApplicableReason(opts)}, nil, nil
	}
	rendered, err := renderTemplate(opts, file.TemplatePath)
	if err != nil {
		return FilePlan{}, nil, err
	}
	renderedBlock, err := validateManagedBlock(rendered)
	if err != nil {
		diagnostic := Diagnostic{Level: "error", Code: "source_template_missing_managed_block", Message: "source agent-instruction template lacks KAS managed markers; no repo-local instruction file can be written from it.", Path: file.TemplatePath}
		return FilePlan{Path: file.Path, TemplatePath: file.TemplatePath, Outcome: "error", Reason: "source template lacks KAS managed markers"}, &diagnostic, nil
	}
	target := filepath.Join(opts.RepoPath, filepath.FromSlash(file.Path))
	existing, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return FilePlan{Path: file.Path, TemplatePath: file.TemplatePath, Outcome: "create", NewSHA256: sha256String(rendered), RenderedSHA256: sha256String(rendered)}, nil, nil
	}
	if err != nil {
		return FilePlan{}, nil, err
	}
	current := string(existing)
	currentSHA := sha256String(current)
	currentBlock, err := validateManagedBlock(current)
	if err != nil {
		if strings.Count(current, ManagedBegin) == 0 && strings.Count(current, ManagedEnd) == 0 {
			diagnostic := Diagnostic{Level: "error", Code: "blocked_unmarked_existing_file", Message: "existing repo-local instruction file lacks KAS managed markers; automatic migration is not allowed.", Path: file.Path}
			return FilePlan{Path: file.Path, TemplatePath: file.TemplatePath, Outcome: "blocked_unmarked_existing_file", PreviousSHA256: currentSHA, Reason: "unmarked existing file requires an explicit migration decision"}, &diagnostic, nil
		}
		diagnostic := Diagnostic{Level: "error", Code: "existing_managed_block_malformed", Message: "existing repo-local instruction file has malformed KAS managed markers; automatic rewrite is not allowed.", Path: file.Path}
		return FilePlan{Path: file.Path, TemplatePath: file.TemplatePath, Outcome: "error", PreviousSHA256: currentSHA, Reason: "existing KAS managed markers are malformed"}, &diagnostic, nil
	}
	if currentBlock == "" {
		diagnostic := Diagnostic{Level: "error", Code: "blocked_unmarked_existing_file", Message: "existing repo-local instruction file lacks KAS managed markers; automatic migration is not allowed.", Path: file.Path}
		return FilePlan{Path: file.Path, TemplatePath: file.TemplatePath, Outcome: "blocked_unmarked_existing_file", PreviousSHA256: currentSHA, Reason: "unmarked existing file requires an explicit migration decision"}, &diagnostic, nil
	}
	merged, err := replaceManagedBlock(current, renderedBlock)
	if err != nil {
		return FilePlan{}, nil, err
	}
	preservation := []PreservationAction{}
	if hasLocalBlock(current) {
		preservation = append(preservation, PreservationAction{Action: "preserve_project_local_block", Detail: "PROJECT:LOCAL block is copied from the existing file."})
	}
	if merged == current {
		return FilePlan{Path: file.Path, TemplatePath: file.TemplatePath, Outcome: "no_change", PreviousSHA256: currentSHA, NewSHA256: currentSHA, RenderedSHA256: sha256String(rendered), PreservationActions: preservation}, nil, nil
	}
	return FilePlan{Path: file.Path, TemplatePath: file.TemplatePath, Outcome: "update_managed_block", PreviousSHA256: currentSHA, NewSHA256: sha256String(merged), RenderedSHA256: sha256String(rendered), PreservationActions: preservation}, nil, nil
}

func plannedContentFor(opts Options, rel string) (string, error) {
	normalized, err := normalizeOptions(opts)
	if err != nil {
		return "", err
	}
	for _, file := range managedFiles {
		if file.Path != rel {
			continue
		}
		rendered, err := renderTemplate(normalized, file.TemplatePath)
		if err != nil {
			return "", err
		}
		renderedBlock, err := validateManagedBlock(rendered)
		if err != nil {
			return "", fmt.Errorf("source template %s lacks KAS managed markers", file.TemplatePath)
		}
		target := filepath.Join(normalized.RepoPath, filepath.FromSlash(file.Path))
		existing, err := os.ReadFile(target)
		if os.IsNotExist(err) {
			return rendered, nil
		}
		if err != nil {
			return "", err
		}
		return replaceManagedBlock(string(existing), renderedBlock)
	}
	return "", fmt.Errorf("unknown managed instruction path %s", rel)
}

func renderTemplate(opts Options, rel string) (string, error) {
	data, err := os.ReadFile(filepath.Join(opts.SourceRepoPath, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	replacements := map[string]string{
		"{{project_name}}":          opts.Project,
		"{{repository_role}}":       opts.RepositoryRole,
		"{{project_suite_id}}":      opts.ProjectSuiteID,
		"{{kab_adoption_stage}}":    opts.KABAdoptionStage,
		"{{upstream_kas_baseline}}": opts.UpstreamKASBaseline,
		"{{local_authority_notes}}": opts.LocalAuthorityNotes,
	}
	rendered := string(data)
	for old, newValue := range replacements {
		rendered = strings.ReplaceAll(rendered, old, newValue)
	}
	return rendered, nil
}

func finalize(result *Result) {
	sort.Slice(result.FilePlans, func(i, j int) bool { return result.FilePlans[i].Path < result.FilePlans[j].Path })
	sort.Slice(result.ChangedPaths, func(i, j int) bool { return result.ChangedPaths[i].Path < result.ChangedPaths[j].Path })
	counts := map[string]int{
		"create":                         0,
		"update_managed_block":           0,
		"preserve_project_local_block":   0,
		"no_change":                      0,
		"not_applicable":                 0,
		"blocked_unmarked_existing_file": 0,
		"error":                          0,
	}
	for _, plan := range result.FilePlans {
		counts[plan.Outcome]++
		if plan.Outcome == "create" || plan.Outcome == "update_managed_block" {
			result.Summary.WritableCount++
		}
		if plan.Outcome == "blocked_unmarked_existing_file" {
			result.Summary.BlockedCount++
		}
		for _, action := range plan.PreservationActions {
			counts[action.Action]++
		}
	}
	result.Summary.CountsByOutcome = counts
	planHash := "sha256:" + hashCanonical(canonicalPlan(*result))
	result.PlanHash = planHash
	result.ApprovalRequest = ApprovalRequest{
		Required:       result.Summary.WritableCount > 0 && result.OK,
		EvidenceRef:    "dry-run:" + planHash,
		PlanHash:       planHash,
		HashDiscipline: "apply requires exact dry-run:sha256:<hash>; stale, malformed, blocked, or mismatched plans fail closed",
	}
	if !result.OK {
		result.NextAction = "Resolve blocked unmarked files or mark a file not_applicable with a reason; no automatic migration or profile install fallback is available."
	} else if result.Summary.WritableCount == 0 {
		result.NextAction = "No repo-local instruction writes are required."
	} else {
		result.NextAction = "Review file_plans and apply the exact approval_request.evidence_ref with update agent-instructions --apply dry-run:sha256:<hash>."
	}
}

func canonicalPlan(result Result) map[string]any {
	return map[string]any{
		"command":       result.Command,
		"mode":          "dry_run",
		"repo_path":     result.RepoPath,
		"source_repo":   result.SourceRepoPath,
		"project":       result.Project,
		"file_plans":    result.FilePlans,
		"changed_paths": result.ChangedPaths,
		"diagnostics":   result.Diagnostics,
		"no_write":      result.NoWrite,
	}
}

func hashCanonical(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func sha256String(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func noWriteEvidence(guaranteed bool) NoWriteEvidence {
	return NoWriteEvidence{
		Guaranteed:                     guaranteed,
		ProfileWriteCount:              0,
		KAHStateWriteCount:             0,
		KABRuntimeMutationCount:        0,
		AuthProviderConfigWriteCount:   0,
		RuntimeProviderModelWriteCount: 0,
	}
}

func notApplicableSet(values []string) map[string]bool {
	set := map[string]bool{}
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				set[item] = true
			}
		}
	}
	return set
}

func notApplicableReason(opts Options) string {
	if strings.TrimSpace(opts.NotApplicableReason) != "" {
		return opts.NotApplicableReason
	}
	return "project policy marks this repo-local instruction file not_applicable"
}

func hasLocalBlock(content string) bool {
	return strings.Contains(content, LocalBegin) && strings.Contains(content, LocalEnd)
}

func validateManagedBlock(content string) (string, error) {
	if strings.Count(content, ManagedBegin) != 1 || strings.Count(content, ManagedEnd) != 1 {
		return "", fmt.Errorf("expected exactly one KAS managed begin marker and one end marker")
	}
	start := strings.Index(content, ManagedBegin)
	end := strings.Index(content, ManagedEnd)
	if start == -1 || end == -1 || end < start {
		return "", fmt.Errorf("KAS managed begin marker must appear before end marker")
	}
	end += len(ManagedEnd)
	return content[start:end], nil
}

func replaceManagedBlock(content string, newBlock string) (string, error) {
	if _, err := validateManagedBlock(newBlock); err != nil {
		return "", err
	}
	oldBlock, err := validateManagedBlock(content)
	if err != nil {
		return "", err
	}
	start := strings.Index(content, oldBlock)
	return content[:start] + newBlock + content[start+len(oldBlock):], nil
}

func parseEvidenceRef(value string) (string, bool) {
	if !strings.HasPrefix(value, "dry-run:sha256:") {
		return "", false
	}
	hash := strings.TrimPrefix(value, "dry-run:")
	if len(hash) != len("sha256:")+64 {
		return hash, false
	}
	for _, r := range strings.TrimPrefix(hash, "sha256:") {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return hash, false
		}
	}
	return hash, true
}
