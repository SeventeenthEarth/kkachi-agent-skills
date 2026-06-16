package workflowcreator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
)

const (
	Command              = "workflow-create"
	RequestSchemaVersion = "kas-workflow-create-request/v1"
	PacketSchemaVersion  = "kas-workflow-create-packet/v1"
	ApprovalSchema       = "kas-workflow-create-approval/v1"
	Canonicalization     = "utf8-json-sorted-keys-normalized-relative-paths/v1"

	ModeDAGOnly     = "dag_only"
	ModeThinTrigger = "thin_trigger"
	ModeFullTrigger = "full_trigger"
)

var approvedEvidencePattern = regexp.MustCompile(`^dry-run:sha256:[0-9a-f]{64}$`)

type Runner func(workDir string, args ...string) CommandResult

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type Options struct {
	Project           string
	WorkflowID        string
	Mode              string
	RequestPath       string
	FullTriggerReason string
	Approval          string
	Runner            Runner
}

type Request struct {
	SchemaVersion     string            `json:"schema_version"`
	SelectorMetadata  map[string]any    `json:"selector_metadata"`
	BaseChecksums     map[string]string `json:"base_checksums,omitempty"`
	Nodes             []RequestNode     `json:"nodes"`
	Trigger           TriggerRequest    `json:"trigger,omitempty"`
	FullTriggerReason string            `json:"full_trigger_reason,omitempty"`
}

type RequestNode struct {
	NodeID            string   `json:"node_id"`
	TaskClass         string   `json:"task_class"`
	DependsOn         []string `json:"depends_on"`
	RequiredOutputs   []string `json:"required_outputs"`
	OwnerRole         string   `json:"owner_role"`
	ExecutionLane     string   `json:"execution_lane"`
	RequiredInputs    []string `json:"required_inputs"`
	ExpectedArtifacts []string `json:"expected_artifacts"`
	PromptRef         string   `json:"prompt_ref"`
	ApprovalRequired  bool     `json:"approval_required"`
	FallbackPolicy    string   `json:"fallback_policy"`
	VerificationGate  string   `json:"verification_gate"`
}

type TriggerRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type Result struct {
	OK                  bool             `json:"ok"`
	Command             string           `json:"command"`
	Mode                string           `json:"mode"`
	Status              string           `json:"status"`
	Project             ProjectEvidence  `json:"project"`
	Workflow            WorkflowEvidence `json:"workflow"`
	Summary             Summary          `json:"summary"`
	MachinePacket       MachinePacket    `json:"machine_packet"`
	ApprovalRequest     ApprovalRequest  `json:"approval_request,omitempty"`
	Approval            ApprovalEvidence `json:"approval,omitempty"`
	KAHCapability       KAHCapability    `json:"kah_capability"`
	Diagnostics         []Diagnostic     `json:"diagnostics"`
	ReasonCodes         []string         `json:"reason_codes"`
	NextAction          string           `json:"next_action"`
	DirectKAHStateWrite bool             `json:"direct_kah_state_write"`
}

type ProjectEvidence struct {
	Path string `json:"path"`
}

type WorkflowEvidence struct {
	ID            string `json:"id"`
	SchemaVersion string `json:"schema_version"`
}

type Summary struct {
	Mode             string         `json:"mode"`
	WorkflowID       string         `json:"workflow_id"`
	GeneratedCount   int            `json:"generated_count"`
	ChangedPathCount int            `json:"changed_path_count"`
	CountsByKind     map[string]int `json:"counts_by_kind"`
	ApproverReview   string         `json:"approver_review"`
}

type MachinePacket struct {
	SchemaVersion      string             `json:"schema_version"`
	ApprovalSchema     string             `json:"approval_schema"`
	Canonicalization   string             `json:"canonicalization"`
	TargetPaths        []string           `json:"target_paths"`
	CandidatePaths     CandidatePaths     `json:"candidate_paths"`
	GeneratedContent   []GeneratedContent `json:"generated_content"`
	SelectorMetadata   map[string]any     `json:"selector_metadata"`
	NodeContracts      []NodeContract     `json:"node_contracts"`
	TriggerPlan        TriggerPlan        `json:"trigger_plan"`
	CapabilityEvidence CapabilityEvidence `json:"capability_evidence"`
	BaseChecksums      map[string]string  `json:"base_checksums"`
	ChangedPaths       []ChangedPath      `json:"changed_paths"`
	Conflicts          []Conflict         `json:"conflicts"`
	Diagnostics        []Diagnostic       `json:"diagnostics"`
	NoWrite            NoWriteEvidence    `json:"no_write"`
	ApprovalHash       string             `json:"approval_hash"`
}

type CandidatePaths struct {
	WorkflowDAG          string `json:"workflow_dag"`
	Catalog              string `json:"catalog"`
	NodeContractRegistry string `json:"node_contract_registry"`
	TriggerSkill         string `json:"trigger_skill,omitempty"`
}

type GeneratedContent struct {
	Path    string `json:"path"`
	Kind    string `json:"kind"`
	Content string `json:"content"`
	SHA256  string `json:"sha256"`
}

type NodeContract struct {
	WorkflowID          string   `json:"workflow_id"`
	NodeID              string   `json:"node_id"`
	TaskClass           string   `json:"task_class"`
	OwnerRole           string   `json:"owner_role"`
	ExecutionLane       string   `json:"execution_lane"`
	RequiredInputs      []string `json:"required_inputs"`
	ExpectedArtifacts   []string `json:"expected_artifacts"`
	PromptRef           string   `json:"prompt_ref"`
	ApprovalRequired    bool     `json:"approval_required"`
	FallbackPolicy      string   `json:"fallback_policy"`
	VerificationGate    string   `json:"verification_gate"`
	CompletionAuthority string   `json:"completion_authority"`
	DirectKAHStateWrite bool     `json:"direct_kah_state_write"`
}

type TriggerPlan struct {
	Mode             string `json:"mode"`
	Path             string `json:"path,omitempty"`
	Generated        bool   `json:"generated"`
	Reason           string `json:"reason,omitempty"`
	DelegatesTo      string `json:"delegates_to,omitempty"`
	CustomLogic      bool   `json:"custom_logic"`
	ValidationStatus string `json:"validation_status,omitempty"`
}

type CapabilityEvidence struct {
	KAS map[string]any `json:"kas"`
	KAH KAHCapability  `json:"kah"`
}

type KAHCapability struct {
	Available          bool     `json:"available"`
	Version            string   `json:"version,omitempty"`
	WorkflowCommands   []string `json:"workflow_commands"`
	CompatibilityFlags []string `json:"compatibility_flags"`
	HelpState          string   `json:"help_state"`
	ApplySurface       string   `json:"apply_surface,omitempty"`
	Reason             string   `json:"reason,omitempty"`
}

type ChangedPath struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	Kind   string `json:"kind"`
}

type Conflict struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type Diagnostic struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
	Field   string `json:"field,omitempty"`
}

type NoWriteEvidence struct {
	Guaranteed                   bool `json:"guaranteed"`
	ProjectWriteCount            int  `json:"project_write_count"`
	CandidateFileWriteCount      int  `json:"candidate_file_write_count"`
	KAHStateWriteCount           int  `json:"kah_state_write_count"`
	KABRuntimeMutationCount      int  `json:"kab_runtime_mutation_count"`
	HermesRuntimeMutationCount   int  `json:"hermes_runtime_mutation_count"`
	ProfileWriteCount            int  `json:"profile_write_count"`
	AuthProviderConfigWriteCount int  `json:"auth_provider_config_write_count"`
}

type ApprovalRequest struct {
	Required                            bool   `json:"required"`
	EvidenceRef                         string `json:"evidence_ref"`
	DryRunPlanHash                      string `json:"dry_run_plan_hash"`
	HashIncludesTargetPaths             bool   `json:"hash_includes_target_paths"`
	HashIncludesGeneratedContent        bool   `json:"hash_includes_generated_content"`
	HashIncludesSelectorMetadata        bool   `json:"hash_includes_selector_metadata"`
	HashIncludesCapabilityEvidence      bool   `json:"hash_includes_capability_evidence"`
	HashIncludesBaseChecksums           bool   `json:"hash_includes_base_checksums"`
	HashIncludesChangedPaths            bool   `json:"hash_includes_changed_paths"`
	HashIncludesConflictsAndDiagnostics bool   `json:"hash_includes_conflicts_and_diagnostics"`
	HashIncludesNoWriteEvidence         bool   `json:"hash_includes_no_write_evidence"`
}

type ApprovalEvidence struct {
	EvidenceRef        string `json:"evidence_ref,omitempty"`
	DryRunPlanHash     string `json:"dry_run_plan_hash,omitempty"`
	ApprovedPlanHash   string `json:"approved_plan_hash,omitempty"`
	MatchedCurrentPlan bool   `json:"matched_current_plan"`
}

type commandRunner struct{}

func (commandRunner) Run(workDir string, args ...string) CommandResult {
	cmd := exec.Command("kkachi-agent-helper", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return CommandResult{Stdout: out, Stderr: exitErr.Stderr, Err: err}
		}
		return CommandResult{Stdout: out, Err: err}
	}
	return CommandResult{Stdout: out}
}

func BuildDryRun(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	result := newResult("workflow_create_dry_run", opts)
	request, err := loadRequest(opts.RequestPath)
	if err != nil {
		return fail(result, mapRequestError(err), err.Error()), nil
	}
	if err := validateInputs(opts, request); err != nil {
		return fail(result, codeForValidationError(err), err.Error()), nil
	}
	fillGeneratedPlan(&result, opts, request)
	result.KAHCapability = preflightKAH(opts)
	result.MachinePacket.CapabilityEvidence = capabilityEvidence(result.KAHCapability)
	if !result.KAHCapability.Available {
		result = addDiagnostic(result, "error", "blocked_missing_kah_workflow_capability", "effective kkachi-agent-helper does not advertise workflow catalog diagnostics; dry-run is non-approvable.", "", "")
	}
	finalize(&result)
	return result, nil
}

func Apply(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	dryRun, err := BuildDryRun(opts)
	if err != nil {
		return dryRun, err
	}
	dryRun.Mode = "workflow_create_apply"
	if !approvedEvidencePattern.MatchString(opts.Approval) {
		return approvalFailure(dryRun, opts.Approval, "", "approval_evidence_malformed", "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>."), nil
	}
	approvedHash := strings.TrimPrefix(opts.Approval, "dry-run:")
	dryRun.Approval = ApprovalEvidence{EvidenceRef: opts.Approval, DryRunPlanHash: dryRun.MachinePacket.ApprovalHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: approvedHash == dryRun.MachinePacket.ApprovalHash}
	if approvedHash != dryRun.MachinePacket.ApprovalHash {
		return approvalFailure(dryRun, opts.Approval, approvedHash, "approval_plan_hash_mismatch", "approval evidence does not match the current workflow-create dry-run packet; no files were written and KAH was not asked to apply."), nil
	}
	if !dryRun.OK || !dryRun.ApprovalRequest.Required {
		return approvalFailure(dryRun, opts.Approval, approvedHash, "workflow_create_plan_not_approvable", "current workflow-create plan is blocked or non-approvable; no files were written."), nil
	}
	if dryRun.KAHCapability.ApplySurface == "" {
		return approvalFailure(dryRun, opts.Approval, approvedHash, "blocked_missing_kah_workflow_catalog_capability", "effective KAH help/capabilities do not advertise a workflow catalog proposal/apply surface; KAS will not direct-write candidate workflow state."), nil
	}
	return approvalFailure(dryRun, opts.Approval, approvedHash, "workflow_create_apply_refused", "KAS detected an advertised apply surface but no reviewed command mapping is implemented for WFLOW-004; stop and update the KAH integration contract before applying."), nil
}

func normalizeOptions(opts Options) Options {
	if opts.Runner == nil {
		opts.Runner = commandRunner{}.Run
	}
	opts.WorkflowID = strings.TrimSpace(opts.WorkflowID)
	opts.Mode = strings.TrimSpace(opts.Mode)
	opts.RequestPath = strings.TrimSpace(opts.RequestPath)
	if opts.Project != "" {
		if abs, err := filepath.Abs(opts.Project); err == nil {
			opts.Project = abs
		}
	}
	return opts
}

func newResult(mode string, opts Options) Result {
	return Result{
		Command:             Command,
		Mode:                mode,
		Status:              "blocked",
		Project:             ProjectEvidence{Path: opts.Project},
		Workflow:            WorkflowEvidence{ID: opts.WorkflowID, SchemaVersion: "task-dag/v1"},
		Diagnostics:         []Diagnostic{},
		ReasonCodes:         []string{},
		DirectKAHStateWrite: false,
	}
}

func loadRequest(path string) (Request, error) {
	if strings.TrimSpace(path) == "" {
		return Request{}, errors.New("workflow-create requires --request <json-path>")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Request{}, fmt.Errorf("request is unreadable: %w", err)
	}
	var request Request
	if err := json.Unmarshal(data, &request); err != nil {
		return Request{}, fmt.Errorf("request must be deterministic JSON: %w", err)
	}
	if request.SchemaVersion != RequestSchemaVersion {
		return Request{}, fmt.Errorf("request must use schema_version %s", RequestSchemaVersion)
	}
	return request, nil
}

func validateInputs(opts Options, request Request) error {
	if opts.Project == "" {
		return errors.New("workflow-create requires --project <path>")
	}
	if opts.WorkflowID == "" {
		return errors.New("workflow-create requires --workflow-id <id>")
	}
	if !safeID(opts.WorkflowID) {
		return errors.New("workflow id must be a simple file-safe id")
	}
	if opts.Mode == "" {
		return errors.New("workflow-create requires --mode dag_only|thin_trigger|full_trigger")
	}
	if !modeAllowed(opts.Mode) {
		return fmt.Errorf("unsupported workflow creator mode %q", opts.Mode)
	}
	if opts.Mode == ModeFullTrigger && strings.TrimSpace(firstNonEmpty(opts.FullTriggerReason, request.FullTriggerReason)) == "" {
		return errors.New("full_trigger mode requires an explicit full_trigger_reason")
	}
	if len(request.SelectorMetadata) == 0 {
		return errors.New("selector_metadata must include deterministic workflow selector metadata")
	}
	if len(request.Nodes) == 0 {
		return errors.New("request must include at least one node")
	}
	for _, node := range request.Nodes {
		if node.NodeID == "" || !safeID(node.NodeID) || node.TaskClass == "" || node.OwnerRole == "" || node.ExecutionLane == "" || node.PromptRef == "" || node.VerificationGate == "" {
			return fmt.Errorf("node contracts require node_id, task_class, owner_role, execution_lane, prompt_ref, and verification_gate")
		}
		if len(node.RequiredOutputs) == 0 || len(node.RequiredInputs) == 0 || len(node.ExpectedArtifacts) == 0 {
			return fmt.Errorf("node %s requires required_outputs, required_inputs, and expected_artifacts", node.NodeID)
		}
		if node.FallbackPolicy != "" && node.FallbackPolicy != "none_fail_closed" {
			return fmt.Errorf("node %s uses unsupported fallback_policy %q", node.NodeID, node.FallbackPolicy)
		}
		for _, path := range append(append([]string{}, node.RequiredOutputs...), node.RequiredInputs...) {
			if !safeRelPath(path) {
				return fmt.Errorf("node %s uses unsafe path %q", node.NodeID, path)
			}
		}
	}
	return nil
}

func fillGeneratedPlan(result *Result, opts Options, request Request) {
	paths := candidatePaths(opts.WorkflowID, opts.Mode)
	result.MachinePacket = MachinePacket{
		SchemaVersion:    PacketSchemaVersion,
		ApprovalSchema:   ApprovalSchema,
		Canonicalization: Canonicalization,
		CandidatePaths:   paths,
		SelectorMetadata: normalizeMap(request.SelectorMetadata),
		BaseChecksums:    map[string]string{},
		Conflicts:        []Conflict{},
		Diagnostics:      []Diagnostic{},
		NoWrite:          NoWriteEvidence{Guaranteed: true},
	}
	dag := renderDAG(opts.WorkflowID, request.Nodes)
	catalog := renderCatalog(opts.WorkflowID, paths)
	contracts := contractsFromRequest(opts.WorkflowID, request.Nodes)
	registry := renderNodeContractRegistry(contracts)
	result.MachinePacket.NodeContracts = contracts
	result.MachinePacket.GeneratedContent = append(result.MachinePacket.GeneratedContent,
		generated(paths.WorkflowDAG, "workflow_dag", dag),
		generated(paths.Catalog, "workflow_catalog", catalog),
		generated(paths.NodeContractRegistry, "node_contract_registry", registry),
	)
	if opts.Mode == ModeThinTrigger || opts.Mode == ModeFullTrigger {
		trigger := renderTriggerSkill(opts.WorkflowID, opts.Mode, firstNonEmpty(opts.FullTriggerReason, request.FullTriggerReason), request.Trigger)
		result.MachinePacket.GeneratedContent = append(result.MachinePacket.GeneratedContent, generated(paths.TriggerSkill, "trigger_skill", trigger))
		result.MachinePacket.TriggerPlan = TriggerPlan{Mode: opts.Mode, Path: paths.TriggerSkill, Generated: true, Reason: firstNonEmpty(opts.FullTriggerReason, request.FullTriggerReason), DelegatesTo: "kkachi-agent-skills workflow-trigger", CustomLogic: opts.Mode == ModeFullTrigger, ValidationStatus: "generated_skill_validated"}
		if err := validateGeneratedTriggerSkill(trigger); err != nil {
			*result = addDiagnostic(*result, "error", "generated_skill_validation_failed", err.Error(), paths.TriggerSkill, "trigger")
			result.MachinePacket.TriggerPlan.ValidationStatus = "generated_skill_validation_failed"
		}
	} else {
		result.MachinePacket.TriggerPlan = TriggerPlan{Mode: opts.Mode, Generated: false, DelegatesTo: "kkachi-agent-skills workflow-trigger", CustomLogic: false}
	}
	for _, item := range result.MachinePacket.GeneratedContent {
		result.MachinePacket.TargetPaths = append(result.MachinePacket.TargetPaths, item.Path)
		result.MachinePacket.ChangedPaths = append(result.MachinePacket.ChangedPaths, ChangedPath{Path: item.Path, Action: plannedAction(opts.Project, item.Path), Kind: item.Kind})
	}
	sort.Strings(result.MachinePacket.TargetPaths)
	sort.Slice(result.MachinePacket.ChangedPaths, func(i, j int) bool {
		return result.MachinePacket.ChangedPaths[i].Path < result.MachinePacket.ChangedPaths[j].Path
	})
	result.MachinePacket.BaseChecksums = baseChecksums(opts.Project, request.BaseChecksums, result.MachinePacket.TargetPaths, &result.Diagnostics)
	result.MachinePacket.Diagnostics = append([]Diagnostic{}, result.Diagnostics...)
}

func finalize(result *Result) {
	result.MachinePacket.Diagnostics = append([]Diagnostic{}, result.Diagnostics...)
	canonical := canonicalApproval(*result)
	result.MachinePacket.ApprovalHash = checksumAny(canonical)
	result.Summary = summarize(*result)
	result.OK = noErrorDiagnostics(result.Diagnostics)
	result.Status = "dry_run_ready"
	if !result.OK {
		result.Status = primaryStatus(result.Diagnostics)
	}
	result.ReasonCodes = reasonCodes(result.Diagnostics)
	result.ApprovalRequest = ApprovalRequest{
		Required:                            result.OK && len(result.MachinePacket.ChangedPaths) > 0,
		EvidenceRef:                         "dry-run:" + result.MachinePacket.ApprovalHash,
		DryRunPlanHash:                      result.MachinePacket.ApprovalHash,
		HashIncludesTargetPaths:             true,
		HashIncludesGeneratedContent:        true,
		HashIncludesSelectorMetadata:        true,
		HashIncludesCapabilityEvidence:      true,
		HashIncludesBaseChecksums:           true,
		HashIncludesChangedPaths:            true,
		HashIncludesConflictsAndDiagnostics: true,
		HashIncludesNoWriteEvidence:         true,
	}
	if !result.OK {
		result.ApprovalRequest.Required = false
		result.NextAction = "Resolve workflow-create diagnostics and rerun dry-run; apply remains fail-closed."
	} else {
		result.NextAction = "Review workflow-create dry-run packet and approve with workflow-create --apply " + result.ApprovalRequest.EvidenceRef + "; KAH remains authoritative for validation/proposal/apply."
	}
}

func canonicalApproval(result Result) map[string]any {
	return map[string]any{
		"schema_version":      ApprovalSchema,
		"canonicalization":    Canonicalization,
		"command":             result.Command,
		"mode":                result.MachinePacket.TriggerPlan.Mode,
		"cli_version":         version.CLIVersion,
		"project":             map[string]any{"relative_root": "."},
		"workflow":            result.Workflow,
		"target_paths":        result.MachinePacket.TargetPaths,
		"generated_content":   result.MachinePacket.GeneratedContent,
		"selector_metadata":   result.MachinePacket.SelectorMetadata,
		"node_contracts":      result.MachinePacket.NodeContracts,
		"trigger_plan":        result.MachinePacket.TriggerPlan,
		"capability_evidence": result.MachinePacket.CapabilityEvidence,
		"base_checksums":      result.MachinePacket.BaseChecksums,
		"changed_paths":       result.MachinePacket.ChangedPaths,
		"conflicts":           result.MachinePacket.Conflicts,
		"diagnostics":         result.MachinePacket.Diagnostics,
		"no_write":            result.MachinePacket.NoWrite,
	}
}

func preflightKAH(opts Options) KAHCapability {
	capability := KAHCapability{WorkflowCommands: []string{}, CompatibilityFlags: []string{}, HelpState: "unavailable"}
	versionResult := opts.Runner(opts.Project, "--version")
	if versionResult.Err != nil {
		capability.Reason = "kah_version_unavailable"
		return capability
	}
	capability.Version = strings.TrimSpace(string(versionResult.Stdout))
	caps := opts.Runner(opts.Project, "capabilities", "--json")
	if caps.Err == nil {
		capability.WorkflowCommands, capability.CompatibilityFlags = parseCapabilityPayload(caps.Stdout)
	}
	help := opts.Runner(opts.Project, "workflow", "--help")
	if help.Err != nil {
		capability.Reason = "workflow_help_unavailable"
		return capability
	}
	helpText := string(help.Stdout) + "\n" + string(help.Stderr)
	capability.HelpState = "ok"
	for _, want := range []string{"validate", "explain", "catalog"} {
		if !contains(capability.WorkflowCommands, want) && !strings.Contains(helpText, want) {
			capability.Reason = "workflow_catalog_diagnostics_missing"
			return capability
		}
	}
	if !(contains(capability.CompatibilityFlags, "task_dag_schema_validation") && contains(capability.CompatibilityFlags, "workflow_catalog_diagnostics") && contains(capability.CompatibilityFlags, "workflow_node_contract_registry_evidence")) {
		capability.Reason = "workflow_catalog_flags_missing"
		return capability
	}
	if strings.Contains(helpText, "workflow catalog apply") && contains(capability.CompatibilityFlags, "workflow_catalog_apply") {
		capability.ApplySurface = "workflow catalog apply"
	}
	capability.Available = true
	return capability
}

func parseCapabilityPayload(data []byte) ([]string, []string) {
	var payload struct {
		CommandGroups []struct {
			Name        string   `json:"name"`
			Status      string   `json:"status"`
			Subcommands []string `json:"subcommands"`
		} `json:"command_groups"`
		Commands           []string        `json:"commands"`
		Flags              []string        `json:"flags"`
		CompatibilityFlags map[string]bool `json:"compatibility_flags"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, nil
	}
	commands := []string{}
	for _, group := range payload.CommandGroups {
		if group.Name == "workflow" && group.Status == "supported" {
			commands = append(commands, group.Subcommands...)
		}
	}
	for _, command := range payload.Commands {
		if strings.HasPrefix(command, "workflow ") {
			commands = append(commands, strings.TrimPrefix(command, "workflow "))
		}
	}
	flags := append([]string{}, payload.Flags...)
	for flag, ok := range payload.CompatibilityFlags {
		if ok {
			flags = append(flags, flag)
		}
	}
	return appendUnique(commands...), appendUnique(flags...)
}

func capabilityEvidence(kah KAHCapability) CapabilityEvidence {
	return CapabilityEvidence{
		KAS: map[string]any{"cli_version": version.CLIVersion, "command": Command},
		KAH: kah,
	}
}

func candidatePaths(workflowID string, mode string) CandidatePaths {
	paths := CandidatePaths{
		WorkflowDAG:          filepath.ToSlash(filepath.Join(".kkachi", "workflows", workflowID+".yaml")),
		Catalog:              filepath.ToSlash(filepath.Join(".kkachi", "workflow-catalog.yaml")),
		NodeContractRegistry: filepath.ToSlash(filepath.Join(".kkachi", "workflows", workflowID+"-node-contracts.yaml")),
	}
	if mode == ModeThinTrigger || mode == ModeFullTrigger {
		paths.TriggerSkill = filepath.ToSlash(filepath.Join(".kkachi", "workflow-triggers", workflowID+"-trigger", "SKILL.md"))
	}
	return paths
}

func renderDAG(workflowID string, nodes []RequestNode) string {
	var b strings.Builder
	fmt.Fprintf(&b, "schema_version: task-dag/v1\nworkflow_id: %s\nnodes:\n", workflowID)
	for _, node := range nodes {
		fmt.Fprintf(&b, "  - id: %s\n", node.NodeID)
		b.WriteString("    depends_on:\n")
		for _, dep := range normalizeStrings(node.DependsOn) {
			fmt.Fprintf(&b, "      - %s\n", dep)
		}
		if len(node.DependsOn) == 0 {
			b.WriteString("      []\n")
		}
		b.WriteString("    join: all_of\n")
		b.WriteString("    required_outputs:\n")
		for _, out := range normalizeStrings(node.RequiredOutputs) {
			fmt.Fprintf(&b, "      - %s\n", out)
		}
	}
	return b.String()
}

func renderCatalog(workflowID string, paths CandidatePaths) string {
	return fmt.Sprintf("schema_version: workflow-catalog/v1\ncatalog_id: kas-custom-workflows\nworkflows:\n  - workflow_id: %s\n    path: %s\n    schema_version: task-dag/v1\n    node_contract_registry: %s\n", workflowID, paths.WorkflowDAG, paths.NodeContractRegistry)
}

func contractsFromRequest(workflowID string, nodes []RequestNode) []NodeContract {
	contracts := []NodeContract{}
	for _, node := range nodes {
		fallback := node.FallbackPolicy
		if fallback == "" {
			fallback = "none_fail_closed"
		}
		contracts = append(contracts, NodeContract{
			WorkflowID: workflowID, NodeID: node.NodeID, TaskClass: node.TaskClass, OwnerRole: node.OwnerRole, ExecutionLane: node.ExecutionLane,
			RequiredInputs: normalizeStrings(node.RequiredInputs), ExpectedArtifacts: normalizeStrings(node.ExpectedArtifacts), PromptRef: node.PromptRef,
			ApprovalRequired: node.ApprovalRequired, FallbackPolicy: fallback, VerificationGate: node.VerificationGate, CompletionAuthority: "kah_only", DirectKAHStateWrite: false,
		})
	}
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].NodeID < contracts[j].NodeID })
	return contracts
}

func renderNodeContractRegistry(contracts []NodeContract) string {
	var b strings.Builder
	b.WriteString("schema_version: kas-task-dag-workflow-registry/v1\nnode_contracts:\n")
	for _, contract := range contracts {
		fmt.Fprintf(&b, "  - workflow_id: %s\n", contract.WorkflowID)
		fmt.Fprintf(&b, "    node_id: %s\n", contract.NodeID)
		fmt.Fprintf(&b, "    task_class: %s\n", contract.TaskClass)
		fmt.Fprintf(&b, "    completion_authority: kah_only\n")
		fmt.Fprintf(&b, "    direct_kah_state_write: false\n")
		fmt.Fprintf(&b, "    owner_role: %s\n", contract.OwnerRole)
		fmt.Fprintf(&b, "    execution_lane: %s\n", contract.ExecutionLane)
		fmt.Fprintf(&b, "    fallback_policy: none_fail_closed\n")
	}
	return b.String()
}

func renderTriggerSkill(workflowID string, mode string, reason string, trigger TriggerRequest) string {
	name := trigger.Name
	if name == "" {
		name = workflowID + "-trigger"
	}
	description := trigger.Description
	if description == "" {
		description = "Thin trigger wrapper for workflow " + workflowID + "."
	}
	customLine := "This generated trigger delegates to kkachi-workflow-trigger and does not add custom dispatch logic."
	if mode == ModeFullTrigger {
		customLine = "Full trigger mode is exceptional; custom input logic is limited to the approved reason: " + reason
	}
	return fmt.Sprintf("---\nname: %s\ndescription: %s\nversion: 0.1.0\n---\n\n# %s\n\nWorkflow id: `%s`.\n\n%s\n\nRun `kkachi-agent-skills workflow-trigger --project <path> --workflow-id %s --node-contract-source <path> --run <run-id> --json` after KAH validates the workflow catalog. This scaffold must not select fallback agents, mutate KAH state directly, or complete workflow nodes.\n", name, description, name, workflowID, customLine, workflowID)
}

func validateGeneratedTriggerSkill(content string) error {
	lines := strings.Split(content, "\n")
	if len(lines) < 5 || lines[0] != "---" {
		return errors.New("generated trigger SKILL.md must start with frontmatter")
	}
	name := strings.TrimPrefix(lines[1], "name: ")
	description := strings.TrimPrefix(lines[2], "description: ")
	if lines[1] == name || lines[2] == description || lines[3] != "version: 0.1.0" || lines[4] != "---" || !safeID(name) || strings.TrimSpace(description) == "" {
		return errors.New("generated trigger SKILL.md frontmatter is invalid")
	}
	for _, required := range []string{"kkachi-agent-skills workflow-trigger", "must not select fallback agents", "mutate KAH state directly"} {
		if !strings.Contains(content, required) {
			return fmt.Errorf("generated trigger SKILL.md missing required boundary text %q", required)
		}
	}
	return nil
}

func generated(path, kind, content string) GeneratedContent {
	return GeneratedContent{Path: path, Kind: kind, Content: content, SHA256: checksumBytes([]byte(content))}
}

func baseChecksums(project string, expected map[string]string, paths []string, diagnostics *[]Diagnostic) map[string]string {
	result := map[string]string{}
	for _, path := range paths {
		actual := "missing"
		data, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(path)))
		if err == nil {
			actual = checksumBytes(data)
		} else if !os.IsNotExist(err) {
			*diagnostics = append(*diagnostics, Diagnostic{Level: "error", Code: "base_catalog_unreadable", Message: err.Error(), Path: path})
		}
		result[path] = actual
		if expected != nil {
			if want, ok := expected[path]; ok && want != actual {
				*diagnostics = append(*diagnostics, Diagnostic{Level: "error", Code: "base_checksum_mismatch", Message: "base checksum does not match current project state", Path: path, Field: "base_checksums"})
			}
		}
	}
	return result
}

func plannedAction(project string, rel string) string {
	if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(rel))); err == nil {
		return "update"
	}
	return "create"
}

func summarize(result Result) Summary {
	counts := map[string]int{}
	for _, item := range result.MachinePacket.GeneratedContent {
		counts[item.Kind]++
	}
	return Summary{Mode: result.MachinePacket.TriggerPlan.Mode, WorkflowID: result.Workflow.ID, GeneratedCount: len(result.MachinePacket.GeneratedContent), ChangedPathCount: len(result.MachinePacket.ChangedPaths), CountsByKind: counts, ApproverReview: "Review target_paths, generated_content, selector_metadata, capability_evidence, base_checksums, changed_paths, diagnostics, and no_write evidence before approval."}
}

func RenderHuman(result Result) string {
	return fmt.Sprintf("Status: %s\nMode: %s\nWorkflow: %s\nGenerated candidates: %d\nChanged paths: %d\nApproval evidence: %s\nDirect KAH state write: false\nNext: %s", result.Status, result.Summary.Mode, result.Workflow.ID, result.Summary.GeneratedCount, result.Summary.ChangedPathCount, result.ApprovalRequest.EvidenceRef, result.NextAction)
}

func approvalFailure(result Result, evidenceRef string, approvedHash string, code string, message string) Result {
	result.OK = false
	result.Status = code
	result.Diagnostics = append(result.Diagnostics, Diagnostic{Level: "error", Code: code, Message: message})
	result.ReasonCodes = reasonCodes(result.Diagnostics)
	result.Approval = ApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: result.MachinePacket.ApprovalHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: approvedHash == result.MachinePacket.ApprovalHash}
	result.NextAction = message
	result.DirectKAHStateWrite = false
	return result
}

func fail(result Result, code string, message string) Result {
	result = addDiagnostic(result, "error", code, message, "", "")
	finalize(&result)
	return result
}

func addDiagnostic(result Result, level string, code string, message string, path string, field string) Result {
	result.Diagnostics = append(result.Diagnostics, Diagnostic{Level: level, Code: code, Message: message, Path: path, Field: field})
	return result
}

func mapRequestError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "--request"):
		return "request_required"
	case strings.Contains(msg, "unreadable"):
		return "request_unreadable"
	case strings.Contains(msg, "JSON"):
		return "request_invalid_json"
	case strings.Contains(msg, "schema_version"):
		return "request_schema_unsupported"
	default:
		return "request_required_field_missing"
	}
}

func codeForValidationError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "--project"):
		return "project_required"
	case strings.Contains(msg, "--workflow-id"):
		return "workflow_id_required"
	case strings.Contains(msg, "file-safe"):
		return "workflow_id_invalid"
	case strings.Contains(msg, "--mode"):
		return "workflow_creator_mode_required"
	case strings.Contains(msg, "unsupported workflow creator mode"):
		return "workflow_creator_mode_unsupported"
	case strings.Contains(msg, "full_trigger"):
		return "full_trigger_reason_required"
	case strings.Contains(msg, "selector_metadata"):
		return "selector_metadata_invalid"
	case strings.Contains(msg, "unsafe path"):
		return "unsafe_target_path"
	case strings.Contains(msg, "fallback_policy"):
		return "unsupported_mode"
	case strings.Contains(msg, "node"):
		return "node_contracts_invalid"
	default:
		return "request_required_field_missing"
	}
}

func modeAllowed(mode string) bool {
	return mode == ModeDAGOnly || mode == ModeThinTrigger || mode == ModeFullTrigger
}

func safeID(id string) bool {
	if id == "." || id == ".." || strings.ContainsAny(id, `/\`) {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func safeRelPath(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	return path != "" && path != "." && !strings.HasPrefix(path, "/") && !strings.Contains(path, "..") && !strings.Contains(path, "\\")
}

func normalizeStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = filepath.ToSlash(strings.TrimSpace(value))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func normalizeMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func checksumAny(value any) string {
	data, _ := json.Marshal(value)
	return checksumBytes(data)
}

func checksumBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func noErrorDiagnostics(diags []Diagnostic) bool {
	for _, diag := range diags {
		if diag.Level == "error" {
			return false
		}
	}
	return true
}

func primaryStatus(diags []Diagnostic) string {
	for _, diag := range diags {
		if diag.Level == "error" {
			return diag.Code
		}
	}
	return "dry_run_ready"
}

func reasonCodes(diags []Diagnostic) []string {
	codes := []string{}
	for _, diag := range diags {
		codes = append(codes, diag.Code)
	}
	return appendUnique(codes...)
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func appendUnique(values ...string) []string {
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
