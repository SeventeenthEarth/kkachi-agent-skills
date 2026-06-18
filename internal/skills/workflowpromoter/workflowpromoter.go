package workflowpromoter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/kahrunner"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowregistry"
)

const (
	Command               = "workflow-promote"
	PacketSchemaVersion   = "kas-workflow-promote-packet/v1"
	ApprovalSchema        = "kas-workflow-promote-approval/v1"
	Canonicalization      = "utf8-json-sorted-keys-normalized-relative-paths/v1"
	MaterializationSchema = "kas-run-local-workflow-materialization/v1"
)

var approvedEvidencePattern = regexp.MustCompile(`^dry-run:sha256:[0-9a-f]{64}$`)

type Runner func(workDir string, args ...string) CommandResult

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type Options struct {
	Project          string
	RunID            string
	Materialization  string
	TargetWorkflowID string
	ReuseReason      string
	ThinTrigger      bool
	Approval         string
	Runner           Runner
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
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

type Summary struct {
	SourceRunID      string         `json:"source_run_id"`
	SourceWorkflowID string         `json:"source_workflow_id"`
	TargetWorkflowID string         `json:"target_workflow_id"`
	GeneratedCount   int            `json:"generated_count"`
	ChangedPathCount int            `json:"changed_path_count"`
	CountsByKind     map[string]int `json:"counts_by_kind"`
	ApproverReview   string         `json:"approver_review"`
}

type MachinePacket struct {
	SchemaVersion      string             `json:"schema_version"`
	ApprovalSchema     string             `json:"approval_schema"`
	Canonicalization   string             `json:"canonicalization"`
	ReuseReason        string             `json:"reuse_reason"`
	SourceProvenance   SourceProvenance   `json:"source_provenance"`
	TargetPaths        []string           `json:"target_paths"`
	CandidatePaths     CandidatePaths     `json:"candidate_paths"`
	GeneratedContent   []GeneratedContent `json:"generated_content"`
	TriggerPlan        TriggerPlan        `json:"trigger_plan"`
	CapabilityEvidence CapabilityEvidence `json:"capability_evidence"`
	BaseChecksums      map[string]string  `json:"base_checksums"`
	ChangedPaths       []ChangedPath      `json:"changed_paths"`
	Conflicts          []Conflict         `json:"conflicts"`
	Diagnostics        []Diagnostic       `json:"diagnostics"`
	NoWrite            NoWriteEvidence    `json:"no_write"`
	ApprovalHash       string             `json:"approval_hash"`
}

type SourceProvenance struct {
	RunID                   string            `json:"run_id"`
	MaterializationPath     string            `json:"materialization_path"`
	MaterializationChecksum string            `json:"materialization_checksum"`
	WorkflowID              string            `json:"workflow_id"`
	WorkflowPath            string            `json:"workflow_path"`
	WorkflowChecksum        string            `json:"workflow_checksum"`
	NodeContractsPath       string            `json:"node_contracts_path"`
	NodeContractsChecksum   string            `json:"node_contracts_checksum"`
	ChecksumsPath           string            `json:"checksums_path"`
	ChecksumsChecksum       string            `json:"checksums_checksum"`
	RecordedChecksums       map[string]string `json:"recorded_checksums"`
	SourceEvidence          map[string]any    `json:"source_evidence,omitempty"`
	RunLocalPosture         string            `json:"run_local_posture"`
	NoPromotion             bool              `json:"no_promotion"`
	PersistentPromotion     bool              `json:"persistent_promotion"`
	DirectKAHStateWrite     bool              `json:"direct_kah_state_write"`
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
	Required                             bool   `json:"required"`
	EvidenceRef                          string `json:"evidence_ref"`
	DryRunPlanHash                       string `json:"dry_run_plan_hash"`
	HashIncludesSourceRunMaterialization bool   `json:"hash_includes_source_run_materialization"`
	HashIncludesSourceChecksums          bool   `json:"hash_includes_source_checksums"`
	HashIncludesTargetPaths              bool   `json:"hash_includes_target_paths"`
	HashIncludesGeneratedContent         bool   `json:"hash_includes_generated_content"`
	HashIncludesTriggerPlan              bool   `json:"hash_includes_trigger_plan"`
	HashIncludesCapabilityEvidence       bool   `json:"hash_includes_capability_evidence"`
	HashIncludesBaseChecksums            bool   `json:"hash_includes_base_checksums"`
	HashIncludesChangedPaths             bool   `json:"hash_includes_changed_paths"`
	HashIncludesConflictsAndDiagnostics  bool   `json:"hash_includes_conflicts_and_diagnostics"`
	HashIncludesNoWriteEvidence          bool   `json:"hash_includes_no_write_evidence"`
}

type ApprovalEvidence struct {
	EvidenceRef        string `json:"evidence_ref,omitempty"`
	DryRunPlanHash     string `json:"dry_run_plan_hash,omitempty"`
	ApprovedPlanHash   string `json:"approved_plan_hash,omitempty"`
	MatchedCurrentPlan bool   `json:"matched_current_plan"`
}

type materializationRecord struct {
	SchemaVersion       string         `json:"schema_version"`
	RunID               string         `json:"run_id"`
	WorkflowID          string         `json:"workflow_id"`
	WorkflowFile        string         `json:"workflow_file"`
	NodeContractSource  string         `json:"node_contract_source"`
	RunLocalPosture     string         `json:"run_local_posture"`
	NoPromotion         bool           `json:"no_promotion"`
	PersistentPromotion bool           `json:"persistent_promotion"`
	SourceEvidence      map[string]any `json:"source_evidence"`
	DirectKAHStateWrite bool           `json:"direct_kah_state_write"`
}

type nodeContractBundle struct {
	SchemaVersion string                          `json:"schema_version"`
	Ref           string                          `json:"ref"`
	Contracts     []workflowregistry.NodeContract `json:"contracts"`
}

type commandRunner struct{}

func (commandRunner) Run(workDir string, args ...string) CommandResult {
	result := kahrunner.Runner{}.Run(workDir, args...)
	return CommandResult{Stdout: result.Stdout, Stderr: result.Stderr, Err: result.Err}
}

func BuildDryRun(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	result := newResult("workflow_promote_dry_run", opts)
	if err := validateOptions(opts); err != nil {
		return fail(result, codeForValidationError(err), err.Error(), "", ""), nil
	}
	source, workflowContent, contracts, err := loadSourceBundle(opts)
	if err != nil {
		return fail(result, sourceLoadCode(err), err.Error(), "", ""), nil
	}
	result.Workflow.SourceID = source.WorkflowID
	fillGeneratedPlan(&result, opts, source, workflowContent, contracts)
	result.KAHCapability = preflightKAH(opts)
	result.MachinePacket.CapabilityEvidence = capabilityEvidence(result.KAHCapability)
	if !result.KAHCapability.Available {
		result = addDiagnostic(result, "error", "blocked_missing_kah_workflow_capability", "effective kkachi-agent-helper does not advertise workflow catalog diagnostics; promotion dry-run is non-approvable.", "", "")
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
	dryRun.Mode = "workflow_promote_apply"
	if !approvedEvidencePattern.MatchString(opts.Approval) {
		return approvalFailure(dryRun, opts.Approval, "", "approval_evidence_malformed", "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>."), nil
	}
	approvedHash := strings.TrimPrefix(opts.Approval, "dry-run:")
	dryRun.Approval = ApprovalEvidence{EvidenceRef: opts.Approval, DryRunPlanHash: dryRun.MachinePacket.ApprovalHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: approvedHash == dryRun.MachinePacket.ApprovalHash}
	if approvedHash != dryRun.MachinePacket.ApprovalHash {
		return approvalFailure(dryRun, opts.Approval, approvedHash, "approval_plan_hash_mismatch", "approval evidence does not match the current workflow-promote dry-run packet; no files were written and KAH was not asked to apply."), nil
	}
	if !dryRun.OK || !dryRun.ApprovalRequest.Required {
		return approvalFailure(dryRun, opts.Approval, approvedHash, "workflow_promote_plan_not_approvable", "current workflow-promote plan is blocked or non-approvable; no files were written."), nil
	}
	if dryRun.KAHCapability.ApplySurface == "" {
		return approvalFailure(dryRun, opts.Approval, approvedHash, "blocked_missing_kah_workflow_catalog_capability", "DAGSM-006 reviewed workflow catalog proposal/apply support is absent; KAS will not direct-write project-local workflow state."), nil
	}
	return approvalFailure(dryRun, opts.Approval, approvedHash, "workflow_promote_apply_refused", "KAS detected an advertised apply surface but no reviewed DAGSM-006 command mapping is implemented for WFLOW-009; stop and update the KAH integration contract before applying."), nil
}

func normalizeOptions(opts Options) Options {
	if opts.Runner == nil {
		opts.Runner = commandRunner{}.Run
	}
	opts.Project = strings.TrimSpace(opts.Project)
	if opts.Project != "" {
		if abs, err := filepath.Abs(opts.Project); err == nil {
			opts.Project = abs
		}
	}
	opts.RunID = strings.TrimSpace(opts.RunID)
	opts.Materialization = strings.TrimSpace(opts.Materialization)
	opts.TargetWorkflowID = strings.TrimSpace(opts.TargetWorkflowID)
	opts.ReuseReason = strings.TrimSpace(opts.ReuseReason)
	opts.Approval = strings.TrimSpace(opts.Approval)
	return opts
}

func newResult(mode string, opts Options) Result {
	return Result{
		Command:             Command,
		Mode:                mode,
		Status:              "blocked",
		Project:             ProjectEvidence{Path: opts.Project},
		Workflow:            WorkflowEvidence{TargetID: opts.TargetWorkflowID},
		Diagnostics:         []Diagnostic{},
		ReasonCodes:         []string{},
		DirectKAHStateWrite: false,
	}
}

func validateOptions(opts Options) error {
	if opts.Project == "" {
		return errors.New("workflow-promote requires --project <path>")
	}
	if opts.RunID == "" {
		return errors.New("workflow-promote requires --run <run-id>")
	}
	if strings.ContainsAny(opts.RunID, "/\\\n\r\t") || opts.RunID == "." || opts.RunID == ".." {
		return errors.New("run id must be printable and path-safe")
	}
	if opts.TargetWorkflowID == "" {
		return errors.New("workflow-promote requires --target-workflow-id <id>")
	}
	if !safeID(opts.TargetWorkflowID) {
		return errors.New("target workflow id must be a simple file-safe id")
	}
	if opts.ReuseReason == "" {
		return errors.New("workflow-promote requires --reuse-reason <reason>")
	}
	return nil
}

func loadSourceBundle(opts Options) (SourceProvenance, string, []workflowregistry.NodeContract, error) {
	materializationPath := opts.Materialization
	if materializationPath == "" {
		materializationPath = filepath.Join(opts.Project, ".kkachi", "runs", opts.RunID, "workflow", "materialization.json")
	}
	materializationBytes, err := os.ReadFile(materializationPath)
	if err != nil {
		return SourceProvenance{}, "", nil, fmt.Errorf("materialization_unreadable: %w", err)
	}
	var materialization materializationRecord
	if err := json.Unmarshal(materializationBytes, &materialization); err != nil {
		return SourceProvenance{}, "", nil, errors.New("materialization_invalid: materialization.json is not parseable JSON")
	}
	if materialization.SchemaVersion != MaterializationSchema {
		return SourceProvenance{}, "", nil, fmt.Errorf("materialization_schema_unsupported: materialization.json must use schema_version %s", MaterializationSchema)
	}
	if materialization.RunID != opts.RunID {
		return SourceProvenance{}, "", nil, errors.New("materialization_run_mismatch: materialization run_id does not match --run")
	}
	if !materialization.NoPromotion || materialization.PersistentPromotion || materialization.DirectKAHStateWrite {
		return SourceProvenance{}, "", nil, errors.New("materialization_authority_drift: source materialization must preserve no_promotion:true, persistent_promotion:false, and direct_kah_state_write:false")
	}
	workflowPath := firstNonEmpty(materialization.WorkflowFile, filepath.ToSlash(filepath.Join(".kkachi", "runs", opts.RunID, "workflow", "workflow.yaml")))
	nodeContractsPath := firstNonEmpty(materialization.NodeContractSource, filepath.ToSlash(filepath.Join(".kkachi", "runs", opts.RunID, "workflow", "node-contracts.json")))
	if !safeRunLocalPath(opts.RunID, workflowPath) || !safeRunLocalPath(opts.RunID, nodeContractsPath) {
		return SourceProvenance{}, "", nil, errors.New("unsafe_run_local_path: source workflow and node-contract paths must stay under the run-local workflow directory")
	}
	workflowBytes, err := os.ReadFile(filepath.Join(opts.Project, filepath.FromSlash(workflowPath)))
	if err != nil {
		return SourceProvenance{}, "", nil, fmt.Errorf("workflow_unreadable: %w", err)
	}
	nodeContractBytes, err := os.ReadFile(filepath.Join(opts.Project, filepath.FromSlash(nodeContractsPath)))
	if err != nil {
		return SourceProvenance{}, "", nil, fmt.Errorf("node_contracts_unreadable: %w", err)
	}
	var bundle nodeContractBundle
	if err := json.Unmarshal(nodeContractBytes, &bundle); err != nil {
		return SourceProvenance{}, "", nil, errors.New("node_contracts_invalid: node-contracts.json is not parseable JSON")
	}
	if bundle.SchemaVersion != workflowregistry.NodeContractsVersion {
		return SourceProvenance{}, "", nil, fmt.Errorf("node_contracts_schema_unsupported: node-contracts.json must use schema_version %s", workflowregistry.NodeContractsVersion)
	}
	if len(bundle.Contracts) == 0 {
		return SourceProvenance{}, "", nil, errors.New("node_contracts_empty: node-contracts.json must include at least one contract")
	}
	for _, contract := range bundle.Contracts {
		if contract.CompletionAuthority != workflowregistry.KAHOnlyAuthority {
			return SourceProvenance{}, "", nil, fmt.Errorf("node_contracts_invalid: node contract %s/%s requires completion_authority %s", contract.WorkflowID, contract.NodeID, workflowregistry.KAHOnlyAuthority)
		}
		if contract.DirectKAHStateWrite == nil || *contract.DirectKAHStateWrite {
			return SourceProvenance{}, "", nil, fmt.Errorf("node_contracts_invalid: node contract %s/%s must set direct_kah_state_write false", contract.WorkflowID, contract.NodeID)
		}
		if contract.FallbackPolicy != workflowregistry.NoFallbackPolicy {
			return SourceProvenance{}, "", nil, fmt.Errorf("node_contracts_invalid: node contract %s/%s uses unsupported fallback_policy %q", contract.WorkflowID, contract.NodeID, contract.FallbackPolicy)
		}
	}
	if err := workflowregistry.ValidateNodeContracts(bundle.Contracts); err != nil {
		return SourceProvenance{}, "", nil, fmt.Errorf("node_contracts_invalid: %w", err)
	}
	sourceWorkflowID := firstNonEmpty(materialization.WorkflowID, bundle.Contracts[0].WorkflowID)
	if !safeID(sourceWorkflowID) {
		return SourceProvenance{}, "", nil, errors.New("workflow_id_invalid: source workflow id is not file-safe")
	}
	checksumsPath := filepath.ToSlash(filepath.Join(".kkachi", "runs", opts.RunID, "workflow", "checksums.json"))
	checksumsAbs := filepath.Join(opts.Project, filepath.FromSlash(checksumsPath))
	checksumBytesRaw, err := os.ReadFile(checksumsAbs)
	if err != nil {
		return SourceProvenance{}, "", nil, fmt.Errorf("checksums_unreadable: %w", err)
	}
	recorded := map[string]string{}
	if err := json.Unmarshal(checksumBytesRaw, &recorded); err != nil {
		return SourceProvenance{}, "", nil, errors.New("checksums_invalid: checksums.json is not parseable JSON")
	}
	workflowChecksum := checksumBytes(workflowBytes)
	nodeContractsChecksum := checksumBytes(nodeContractBytes)
	if err := verifyRecordedChecksum(recorded, workflowPath, workflowChecksum); err != nil {
		return SourceProvenance{}, "", nil, err
	}
	if err := verifyRecordedChecksum(recorded, nodeContractsPath, nodeContractsChecksum); err != nil {
		return SourceProvenance{}, "", nil, err
	}
	return SourceProvenance{
		RunID:                   opts.RunID,
		MaterializationPath:     filepath.ToSlash(materializationPath),
		MaterializationChecksum: checksumBytes(materializationBytes),
		WorkflowID:              sourceWorkflowID,
		WorkflowPath:            workflowPath,
		WorkflowChecksum:        workflowChecksum,
		NodeContractsPath:       nodeContractsPath,
		NodeContractsChecksum:   nodeContractsChecksum,
		ChecksumsPath:           checksumsPath,
		ChecksumsChecksum:       checksumBytes(checksumBytesRaw),
		RecordedChecksums:       normalizeStringMap(recorded),
		SourceEvidence:          normalizeAnyMap(materialization.SourceEvidence),
		RunLocalPosture:         materialization.RunLocalPosture,
		NoPromotion:             materialization.NoPromotion,
		PersistentPromotion:     materialization.PersistentPromotion,
		DirectKAHStateWrite:     materialization.DirectKAHStateWrite,
	}, string(workflowBytes), bundle.Contracts, nil
}

func fillGeneratedPlan(result *Result, opts Options, source SourceProvenance, workflowContent string, sourceContracts []workflowregistry.NodeContract) {
	paths := candidatePaths(opts.TargetWorkflowID, opts.ThinTrigger)
	targetWorkflow, err := retargetWorkflow(workflowContent, source.WorkflowID, opts.TargetWorkflowID)
	if err != nil {
		*result = addDiagnostic(*result, "error", "source_workflow_invalid", err.Error(), source.WorkflowPath, "workflow_id")
		targetWorkflow = workflowContent
	}
	contracts := retargetContracts(sourceContracts, opts.TargetWorkflowID)
	registry := renderNodeContractRegistry(contracts)
	catalog := renderCatalog(opts.TargetWorkflowID, paths)
	result.Workflow.SourceID = source.WorkflowID
	result.MachinePacket = MachinePacket{
		SchemaVersion:    PacketSchemaVersion,
		ApprovalSchema:   ApprovalSchema,
		Canonicalization: Canonicalization,
		ReuseReason:      opts.ReuseReason,
		SourceProvenance: source,
		CandidatePaths:   paths,
		GeneratedContent: []GeneratedContent{
			generated(paths.WorkflowDAG, "workflow_dag", targetWorkflow),
			generated(paths.Catalog, "workflow_catalog", catalog),
			generated(paths.NodeContractRegistry, "node_contract_registry", registry),
		},
		Conflicts:   []Conflict{},
		Diagnostics: []Diagnostic{},
		NoWrite:     NoWriteEvidence{Guaranteed: true},
	}
	if opts.ThinTrigger {
		trigger := renderThinTriggerSkill(opts.TargetWorkflowID, opts.ReuseReason, paths.NodeContractRegistry)
		result.MachinePacket.GeneratedContent = append(result.MachinePacket.GeneratedContent, generated(paths.TriggerSkill, "trigger_skill", trigger))
		result.MachinePacket.TriggerPlan = TriggerPlan{Mode: "thin_trigger", Path: paths.TriggerSkill, Generated: true, Reason: opts.ReuseReason, DelegatesTo: "kkachi-agent-skills workflow-trigger", CustomLogic: false, ValidationStatus: "generated_skill_validated"}
	} else {
		result.MachinePacket.TriggerPlan = TriggerPlan{Mode: "none", Generated: false, DelegatesTo: "kkachi-agent-skills workflow-trigger", CustomLogic: false}
	}
	for _, item := range result.MachinePacket.GeneratedContent {
		result.MachinePacket.TargetPaths = append(result.MachinePacket.TargetPaths, item.Path)
		result.MachinePacket.ChangedPaths = append(result.MachinePacket.ChangedPaths, ChangedPath{Path: item.Path, Action: plannedAction(opts.Project, item.Path), Kind: item.Kind})
	}
	sort.Strings(result.MachinePacket.TargetPaths)
	sort.Slice(result.MachinePacket.ChangedPaths, func(i, j int) bool {
		return result.MachinePacket.ChangedPaths[i].Path < result.MachinePacket.ChangedPaths[j].Path
	})
	result.MachinePacket.BaseChecksums = baseChecksums(opts.Project, result.MachinePacket.TargetPaths, &result.Diagnostics)
	result.MachinePacket.Diagnostics = append([]Diagnostic{}, result.Diagnostics...)
}

func finalize(result *Result) {
	result.MachinePacket.Diagnostics = append([]Diagnostic{}, result.Diagnostics...)
	result.MachinePacket.ApprovalHash = checksumAny(canonicalApproval(*result))
	result.Summary = summarize(*result)
	result.OK = noErrorDiagnostics(result.Diagnostics)
	result.Status = "dry_run_ready"
	if !result.OK {
		result.Status = primaryStatus(result.Diagnostics)
	}
	result.ReasonCodes = reasonCodes(result.Diagnostics)
	result.ApprovalRequest = ApprovalRequest{
		Required:                             result.OK && len(result.MachinePacket.ChangedPaths) > 0,
		EvidenceRef:                          "dry-run:" + result.MachinePacket.ApprovalHash,
		DryRunPlanHash:                       result.MachinePacket.ApprovalHash,
		HashIncludesSourceRunMaterialization: true,
		HashIncludesSourceChecksums:          true,
		HashIncludesTargetPaths:              true,
		HashIncludesGeneratedContent:         true,
		HashIncludesTriggerPlan:              true,
		HashIncludesCapabilityEvidence:       true,
		HashIncludesBaseChecksums:            true,
		HashIncludesChangedPaths:             true,
		HashIncludesConflictsAndDiagnostics:  true,
		HashIncludesNoWriteEvidence:          true,
	}
	if !result.OK {
		result.ApprovalRequest.Required = false
		result.NextAction = "Resolve workflow-promote diagnostics and rerun dry-run; apply remains fail-closed."
	} else {
		result.NextAction = "Review workflow-promote dry-run packet and approve with workflow-promote --apply " + result.ApprovalRequest.EvidenceRef + "; KAH remains authoritative for project-local workflow/catalog proposal/apply."
	}
}

func canonicalApproval(result Result) map[string]any {
	return map[string]any{
		"schema_version":      ApprovalSchema,
		"canonicalization":    Canonicalization,
		"command":             result.Command,
		"cli_version":         version.CLIVersion,
		"project":             map[string]any{"relative_root": "."},
		"workflow":            result.Workflow,
		"reuse_reason":        result.MachinePacket.ReuseReason,
		"source_provenance":   result.MachinePacket.SourceProvenance,
		"target_paths":        result.MachinePacket.TargetPaths,
		"generated_content":   result.MachinePacket.GeneratedContent,
		"trigger_plan":        result.MachinePacket.TriggerPlan,
		"capability_evidence": result.MachinePacket.CapabilityEvidence,
		"base_checksums":      result.MachinePacket.BaseChecksums,
		"changed_paths":       result.MachinePacket.ChangedPaths,
		"conflicts":           result.MachinePacket.Conflicts,
		"diagnostics":         result.MachinePacket.Diagnostics,
		"no_write":            result.MachinePacket.NoWrite,
	}
}

func RecomputeApprovalHash(result Result) string {
	return checksumAny(canonicalApproval(result))
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

func candidatePaths(workflowID string, thinTrigger bool) CandidatePaths {
	paths := CandidatePaths{
		WorkflowDAG:          filepath.ToSlash(filepath.Join(".kkachi", "workflows", workflowID+".yaml")),
		Catalog:              filepath.ToSlash(filepath.Join(".kkachi", "workflow-catalog.yaml")),
		NodeContractRegistry: filepath.ToSlash(filepath.Join(".kkachi", "workflows", workflowID+"-node-contracts.yaml")),
	}
	if thinTrigger {
		paths.TriggerSkill = filepath.ToSlash(filepath.Join(".kkachi", "workflow-triggers", workflowID+"-trigger", "SKILL.md"))
	}
	return paths
}

func retargetWorkflow(content string, sourceID string, targetID string) (string, error) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "workflow_id:") {
			prefix := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines[i] = prefix + "workflow_id: " + targetID
			out := strings.Join(lines, "\n")
			if !strings.HasSuffix(out, "\n") {
				out += "\n"
			}
			return out, nil
		}
	}
	return "", fmt.Errorf("source workflow %s does not declare workflow_id", sourceID)
}

func retargetContracts(source []workflowregistry.NodeContract, targetID string) []workflowregistry.NodeContract {
	contracts := make([]workflowregistry.NodeContract, 0, len(source))
	falseValue := false
	for _, item := range source {
		contract := item
		contract.WorkflowID = targetID
		contract.DirectKAHStateWrite = &falseValue
		contracts = append(contracts, contract)
	}
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].NodeID < contracts[j].NodeID })
	return contracts
}

func renderCatalog(workflowID string, paths CandidatePaths) string {
	return fmt.Sprintf("schema_version: workflow-catalog/v1\ncatalog_id: kas-promoted-workflows\nworkflows:\n  - workflow_id: %s\n    path: %s\n    schema_version: task-dag/v1\n    node_contract_registry: %s\n", workflowID, paths.WorkflowDAG, paths.NodeContractRegistry)
}

func renderNodeContractRegistry(contracts []workflowregistry.NodeContract) string {
	var b strings.Builder
	b.WriteString("schema_version: kas-task-dag-workflow-registry/v1\nnode_contracts:\n")
	for _, contract := range contracts {
		fmt.Fprintf(&b, "  - workflow_id: %s\n", contract.WorkflowID)
		fmt.Fprintf(&b, "    node_id: %s\n", contract.NodeID)
		fmt.Fprintf(&b, "    task_class: %s\n", contract.TaskClass)
		fmt.Fprintf(&b, "    owner_role: %s\n", contract.OwnerRole)
		fmt.Fprintf(&b, "    execution_lane: %s\n", contract.ExecutionLane)
		b.WriteString("    required_inputs:\n")
		for _, input := range normalized(contract.RequiredInputs) {
			fmt.Fprintf(&b, "      - %s\n", input)
		}
		b.WriteString("    expected_artifacts:\n")
		for _, artifact := range normalized(contract.ExpectedArtifacts) {
			fmt.Fprintf(&b, "      - %s\n", artifact)
		}
		fmt.Fprintf(&b, "    prompt_ref: %s\n", contract.PromptRef)
		fmt.Fprintf(&b, "    approval_required: %t\n", contract.ApprovalRequired)
		fmt.Fprintf(&b, "    fallback_policy: %s\n", contract.FallbackPolicy)
		fmt.Fprintf(&b, "    verification_gate: %s\n", contract.VerificationGate)
		fmt.Fprintf(&b, "    completion_authority: %s\n", contract.CompletionAuthority)
		b.WriteString("    direct_kah_state_write: false\n")
	}
	return b.String()
}

func renderThinTriggerSkill(workflowID string, reason string, nodeContractRegistry string) string {
	name := workflowID + "-trigger"
	return fmt.Sprintf("---\nname: %s\ndescription: Thin trigger wrapper for promoted workflow %s.\nversion: 0.1.0\n---\n\n# %s\n\nPromoted workflow id: `%s`.\n\nReuse reason: %s\n\nThis generated trigger delegates to `kkachi-agent-skills workflow-trigger` and does not add custom dispatch logic.\n\nRun `kkachi-agent-skills workflow-trigger --project <path> --workflow-id %s --node-contract-source %s --run <run-id> --json` after KAH validates the workflow catalog. This scaffold must not select fallback agents, mutate KAH state directly, mutate KAB runtime state, or complete workflow nodes.\n", name, workflowID, name, workflowID, reason, workflowID, nodeContractRegistry)
}

func generated(path, kind, content string) GeneratedContent {
	return GeneratedContent{Path: path, Kind: kind, Content: content, SHA256: checksumBytes([]byte(content))}
}

func baseChecksums(project string, paths []string, diagnostics *[]Diagnostic) map[string]string {
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
	return Summary{
		SourceRunID:      result.MachinePacket.SourceProvenance.RunID,
		SourceWorkflowID: result.Workflow.SourceID,
		TargetWorkflowID: result.Workflow.TargetID,
		GeneratedCount:   len(result.MachinePacket.GeneratedContent),
		ChangedPathCount: len(result.MachinePacket.ChangedPaths),
		CountsByKind:     counts,
		ApproverReview:   "Review source_provenance, target_paths, generated_content, trigger_plan, capability_evidence, base_checksums, changed_paths, diagnostics, and no_write evidence before approval.",
	}
}

func RenderHuman(result Result) string {
	return fmt.Sprintf("Status: %s\nSource run: %s\nSource workflow: %s\nTarget workflow: %s\nGenerated candidates: %d\nChanged paths: %d\nApproval evidence: %s\nDirect KAH state write: false\nNext: %s", result.Status, result.Summary.SourceRunID, result.Workflow.SourceID, result.Workflow.TargetID, result.Summary.GeneratedCount, result.Summary.ChangedPathCount, result.ApprovalRequest.EvidenceRef, result.NextAction)
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

func fail(result Result, code string, message string, path string, field string) Result {
	result = addDiagnostic(result, "error", code, message, path, field)
	finalize(&result)
	return result
}

func addDiagnostic(result Result, level string, code string, message string, path string, field string) Result {
	result.Diagnostics = append(result.Diagnostics, Diagnostic{Level: level, Code: code, Message: message, Path: path, Field: field})
	return result
}

func codeForValidationError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "--project"):
		return "project_required"
	case strings.Contains(msg, "--run"):
		return "run_id_required"
	case strings.Contains(msg, "--target-workflow-id"):
		return "target_workflow_id_required"
	case strings.Contains(msg, "target workflow id"):
		return "target_workflow_id_invalid"
	case strings.Contains(msg, "path-safe"):
		return "run_id_invalid"
	case strings.Contains(msg, "--reuse-reason"):
		return "reuse_reason_required"
	default:
		return "workflow_promote_invalid"
	}
}

func sourceLoadCode(err error) string {
	msg := err.Error()
	if i := strings.Index(msg, ":"); i > 0 {
		code := msg[:i]
		if strings.Contains(code, "_") {
			return code
		}
	}
	return "source_bundle_invalid"
}

func verifyRecordedChecksum(recorded map[string]string, path string, actual string) error {
	want, ok := recorded[path]
	if !ok || strings.TrimSpace(want) == "" {
		return fmt.Errorf("source_checksum_missing: recorded checksum for %s is required", path)
	}
	if want != actual {
		return fmt.Errorf("source_checksum_mismatch: recorded checksum for %s does not match current source", path)
	}
	return nil
}

func safeID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func safeRunLocalPath(runID string, path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	prefix := filepath.ToSlash(filepath.Join(".kkachi", "runs", runID, "workflow")) + "/"
	return safeRelPath(path) && strings.HasPrefix(path, prefix)
}

func safeRelPath(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	if path == "" || path == "." || strings.HasPrefix(path, "/") || strings.Contains(path, "\\") {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

func normalizeAnyMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}
	result := map[string]any{}
	for k, v := range input {
		result[k] = v
	}
	return result
}

func normalizeStringMap(input map[string]string) map[string]string {
	result := map[string]string{}
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result[k] = input[k]
	}
	return result
}

func normalized(values []string) []string {
	out := append([]string{}, values...)
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

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
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
	return "blocked"
}

func reasonCodes(diags []Diagnostic) []string {
	codes := []string{}
	for _, diag := range diags {
		codes = append(codes, diag.Code)
	}
	return appendUnique(codes...)
}

func checksumAny(value any) string {
	data, _ := json.Marshal(value)
	return "sha256:" + checksumHex(data)
}

func checksumBytes(data []byte) string {
	return "sha256:" + checksumHex(data)
}

func checksumHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
