package workflowmaterializer

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

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowcreator"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowregistry"
)

const (
	Command       = "workflow-materialize"
	SchemaVersion = "kas-run-local-workflow-materialization/v1"
)

type Options struct {
	Project              string
	RunID                string
	RouteResult          string
	CustomWorkflowPacket string
	Approval             string
}

type Result struct {
	OK                       bool              `json:"ok"`
	Command                  string            `json:"command"`
	SchemaVersion            string            `json:"schema_version"`
	Status                   string            `json:"status"`
	Project                  string            `json:"project"`
	RunID                    string            `json:"run_id"`
	WorkflowID               string            `json:"workflow_id,omitempty"`
	WorkflowFile             string            `json:"workflow_file,omitempty"`
	NodeContractSource       string            `json:"node_contract_source,omitempty"`
	RouteResultCopy          string            `json:"route_result_copy,omitempty"`
	CustomWorkflowPacketCopy string            `json:"custom_workflow_packet_copy,omitempty"`
	MaterializationFile      string            `json:"materialization_file,omitempty"`
	ChecksumsFile            string            `json:"checksums_file,omitempty"`
	SelectedBundle           string            `json:"selected_bundle,omitempty"`
	TaskClass                string            `json:"task_class,omitempty"`
	ClassificationReason     string            `json:"classification_reason,omitempty"`
	RunLocalPosture          string            `json:"run_local_posture"`
	NoPromotion              bool              `json:"no_promotion"`
	PersistentPromotion      bool              `json:"persistent_promotion"`
	SourceEvidence           SourceEvidence    `json:"source_evidence"`
	Checksums                map[string]string `json:"checksums,omitempty"`
	WrittenPaths             []string          `json:"written_paths"`
	Diagnostics              []Diagnostic      `json:"diagnostics"`
	ReasonCodes              []string          `json:"reason_codes"`
	DirectKAHStateWrite      bool              `json:"direct_kah_state_write"`
}

type SourceEvidence struct {
	RouteResultPath              string `json:"route_result_path,omitempty"`
	RouteResultChecksum          string `json:"route_result_checksum,omitempty"`
	CustomWorkflowPacketPath     string `json:"custom_workflow_packet_path,omitempty"`
	CustomWorkflowPacketChecksum string `json:"custom_workflow_packet_checksum,omitempty"`
	ApprovalEvidence             string `json:"approval_evidence,omitempty"`
	DryRunPlanHash               string `json:"dry_run_plan_hash,omitempty"`
	ApprovedPlanHash             string `json:"approved_plan_hash,omitempty"`
	TaxonomyPath                 string `json:"taxonomy_path,omitempty"`
	TaxonomyChecksum             string `json:"taxonomy_checksum,omitempty"`
	SelectorRegistryPath         string `json:"selector_registry_path,omitempty"`
	SelectorRegistryChecksum     string `json:"selector_registry_checksum,omitempty"`
	RegistryVersion              string `json:"registry_version,omitempty"`
}

type Diagnostic struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
	Field   string `json:"field,omitempty"`
}

type routeResult struct {
	OK                   bool                 `json:"ok"`
	Command              string               `json:"command"`
	Status               string               `json:"status"`
	TaskClass            string               `json:"task_class"`
	ClassificationReason string               `json:"classification_reason"`
	SelectedBundle       string               `json:"selected_bundle"`
	WorkflowID           string               `json:"workflow_id"`
	WorkflowPath         string               `json:"workflow_path"`
	SelectedSpine        string               `json:"selected_spine"`
	SkippedPhaseReasons  map[string]string    `json:"skipped_phase_reasons"`
	Taxonomy             sourceResultEvidence `json:"taxonomy"`
	SelectorRegistry     sourceResultEvidence `json:"selector_registry"`
	DirectKAHStateWrite  bool                 `json:"direct_kah_state_write"`
}

type sourceResultEvidence struct {
	Path     string `json:"path"`
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}

type contractBundle struct {
	SchemaVersion string                          `json:"schema_version"`
	Ref           string                          `json:"ref"`
	Contracts     []workflowregistry.NodeContract `json:"contracts"`
}

func MaterializeFromRoute(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	result := newResult(opts)
	if opts.Project == "" {
		return fail(result, "project_required", "workflow materialization requires --project <path>.", "", "project"), nil
	}
	if opts.RunID == "" {
		return fail(result, "run_id_required", "workflow materialization requires --run <run-id>.", "", "run"), nil
	}
	if opts.RouteResult == "" {
		return fail(result, "route_result_required", "workflow materialization requires --route-result <json-path>.", "", "route_result"), nil
	}
	if strings.ContainsAny(opts.RunID, "/\\\n\r\t") || opts.RunID == "." || opts.RunID == ".." {
		return fail(result, "run_id_invalid", "run id must be printable and path-safe.", "", "run"), nil
	}

	routeBytes, err := os.ReadFile(opts.RouteResult)
	if err != nil {
		return fail(result, "route_result_unreadable", err.Error(), opts.RouteResult, "route_result"), nil
	}
	var route routeResult
	if err := json.Unmarshal(routeBytes, &route); err != nil {
		return fail(result, "route_result_invalid", "route result JSON is not parseable.", opts.RouteResult, "route_result"), nil
	}
	if !route.OK || route.Status != "bundle_route_matched" {
		return fail(result, "route_result_not_matched", "route result must be ok:true with status bundle_route_matched.", opts.RouteResult, "route_result"), nil
	}
	if route.DirectKAHStateWrite {
		return fail(result, "route_result_direct_kah_state_write_true", "route result must preserve direct_kah_state_write:false.", opts.RouteResult, "direct_kah_state_write"), nil
	}
	workflowID := firstNonEmpty(route.WorkflowID, route.SelectedBundle, route.SelectedSpine)
	if !safeID(workflowID) {
		return fail(result, "workflow_id_invalid", "route result selected workflow id is not file-safe.", opts.RouteResult, "workflow_id"), nil
	}
	if route.SelectorRegistry.Path == "" {
		return fail(result, "selector_registry_required", "route result must include selector_registry.path.", opts.RouteResult, "selector_registry.path"), nil
	}

	registry, err := workflowregistry.Load(route.SelectorRegistry.Path)
	if err != nil {
		return fail(result, "selector_registry_unreadable", err.Error(), route.SelectorRegistry.Path, "selector_registry.path"), nil
	}
	if route.SelectorRegistry.Checksum != "" && route.SelectorRegistry.Checksum != registry.Checksum {
		return fail(result, "selector_registry_checksum_mismatch", "route result selector registry checksum does not match current source.", registry.Path, "selector_registry.checksum"), nil
	}
	if _, ok := findWorkflow(registry.Workflows, workflowID); !ok {
		return fail(result, "bundle_no_match", "selected bundle is absent from selector registry.", registry.Path, "workflow_id"), nil
	}
	contracts := workflowregistry.ContractsForWorkflow(registry.NodeContracts, workflowID)
	if len(contracts) == 0 {
		return fail(result, "blocked_missing_ready_node_contract", "selected bundle has no node contracts.", registry.Path, "node_contracts"), nil
	}

	base := filepath.ToSlash(filepath.Join(".kkachi", "runs", opts.RunID, "workflow"))
	paths := map[string]string{
		"materialization": filepath.ToSlash(filepath.Join(base, "materialization.json")),
		"workflow":        filepath.ToSlash(filepath.Join(base, "workflow.yaml")),
		"node_contracts":  filepath.ToSlash(filepath.Join(base, "node-contracts.json")),
		"route_result":    filepath.ToSlash(filepath.Join(base, "route-result.json")),
		"checksums":       filepath.ToSlash(filepath.Join(base, "checksums.json")),
	}
	for _, path := range paths {
		if !safeRunLocalPath(opts.RunID, path) {
			return fail(result, "unsafe_run_local_path", "materialization path is outside the run-local workflow directory.", path, "path"), nil
		}
		if err := ensureSafeTarget(opts.Project, path); err != nil {
			return fail(result, "unsafe_run_local_path", err.Error(), path, "path"), nil
		}
	}

	workflowContent := renderWorkflowDAG(workflowID, contracts)
	bundle := contractBundle{SchemaVersion: workflowregistry.NodeContractsVersion, Ref: "run-local:" + opts.RunID + ":" + workflowID, Contracts: contracts}
	contractBytes, _ := json.MarshalIndent(bundle, "", "  ")
	contractBytes = append(contractBytes, '\n')

	result.OK = true
	result.Status = "materialized"
	result.WorkflowID = workflowID
	result.WorkflowFile = paths["workflow"]
	result.NodeContractSource = paths["node_contracts"]
	result.RouteResultCopy = paths["route_result"]
	result.MaterializationFile = paths["materialization"]
	result.ChecksumsFile = paths["checksums"]
	result.SelectedBundle = route.SelectedBundle
	result.TaskClass = route.TaskClass
	result.ClassificationReason = route.ClassificationReason
	result.SourceEvidence = SourceEvidence{
		RouteResultPath:          filepath.ToSlash(opts.RouteResult),
		RouteResultChecksum:      checksumBytes(routeBytes),
		TaxonomyPath:             route.Taxonomy.Path,
		TaxonomyChecksum:         route.Taxonomy.Checksum,
		SelectorRegistryPath:     registry.Path,
		SelectorRegistryChecksum: registry.Checksum,
		RegistryVersion:          registry.Version,
	}
	result.Checksums = map[string]string{
		paths["workflow"]:       checksumBytes([]byte(workflowContent)),
		paths["node_contracts"]: checksumBytes(contractBytes),
		paths["route_result"]:   checksumBytes(routeBytes),
	}
	result.WrittenPaths = []string{paths["materialization"], paths["workflow"], paths["node_contracts"], paths["route_result"], paths["checksums"]}
	sort.Strings(result.WrittenPaths)

	writes := map[string][]byte{
		paths["workflow"]:        []byte(workflowContent),
		paths["node_contracts"]:  contractBytes,
		paths["route_result"]:    routeBytes,
		paths["checksums"]:       jsonWithNewline(result.Checksums),
		paths["materialization"]: jsonWithNewline(result),
	}
	for _, path := range result.WrittenPaths {
		data := writes[path]
		if len(data) == 0 {
			continue
		}
		if err := writeProjectFile(opts.Project, path, data); err != nil {
			return fail(result, "materialization_write_failed", err.Error(), path, "path"), nil
		}
	}
	return result, nil
}

func MaterializeFromCustomPacket(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	result := newResult(opts)
	if opts.Project == "" {
		return fail(result, "project_required", "workflow materialization requires --project <path>.", "", "project"), nil
	}
	if opts.RunID == "" {
		return fail(result, "run_id_required", "workflow materialization requires --run <run-id>.", "", "run"), nil
	}
	if opts.CustomWorkflowPacket == "" {
		return fail(result, "custom_workflow_packet_required", "custom workflow materialization requires --custom-workflow-packet <json-path>.", "", "custom_workflow_packet"), nil
	}
	if opts.Approval == "" {
		return fail(result, "approval_evidence_required", "custom workflow materialization requires --approval dry-run:sha256:<hash>.", "", "approval"), nil
	}
	approvedHash, ok := approvedPlanHash(opts.Approval)
	if !ok {
		return fail(result, "approval_evidence_malformed", "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>.", "", "approval"), nil
	}
	if strings.ContainsAny(opts.RunID, "/\\\n\r\t") || opts.RunID == "." || opts.RunID == ".." {
		return fail(result, "run_id_invalid", "run id must be printable and path-safe.", "", "run"), nil
	}

	packetBytes, err := os.ReadFile(opts.CustomWorkflowPacket)
	if err != nil {
		return fail(result, "custom_workflow_packet_unreadable", err.Error(), opts.CustomWorkflowPacket, "custom_workflow_packet"), nil
	}
	var packet workflowcreator.Result
	if err := json.Unmarshal(packetBytes, &packet); err != nil {
		return fail(result, "custom_workflow_packet_invalid", "custom workflow packet JSON is not parseable.", opts.CustomWorkflowPacket, "custom_workflow_packet"), nil
	}
	if !packet.OK || packet.Command != workflowcreator.Command || packet.Status != "dry_run_ready" {
		return fail(result, "custom_workflow_packet_not_approvable", "custom workflow packet must be an ok:true workflow-create dry-run result.", opts.CustomWorkflowPacket, "custom_workflow_packet"), nil
	}
	if packet.DirectKAHStateWrite {
		return fail(result, "custom_workflow_packet_direct_kah_state_write_true", "custom workflow packet must preserve direct_kah_state_write:false.", opts.CustomWorkflowPacket, "direct_kah_state_write"), nil
	}
	if packet.MachinePacket.SchemaVersion != workflowcreator.PacketSchemaVersion || packet.MachinePacket.ApprovalSchema != workflowcreator.ApprovalSchema || packet.MachinePacket.Canonicalization != workflowcreator.Canonicalization {
		return fail(result, "custom_workflow_packet_schema_unsupported", "custom workflow packet schema/canonicalization is unsupported.", opts.CustomWorkflowPacket, "machine_packet.schema_version"), nil
	}
	recomputedHash := workflowcreator.RecomputeApprovalHash(packet)
	if recomputedHash != packet.MachinePacket.ApprovalHash {
		return fail(result, "custom_workflow_packet_hash_mismatch", "custom workflow packet approval_hash does not match the WFLOW-004 canonical packet hash.", opts.CustomWorkflowPacket, "machine_packet.approval_hash"), nil
	}
	if approvedHash != recomputedHash {
		return fail(result, "approval_plan_hash_mismatch", "approval evidence does not match the custom workflow dry-run packet; no files were written.", opts.CustomWorkflowPacket, "approval"), nil
	}

	workflowID := firstNonEmpty(packet.Workflow.ID, workflowIDFromContracts(packet.MachinePacket.NodeContracts))
	if !safeID(workflowID) {
		return fail(result, "workflow_id_invalid", "custom workflow packet workflow id is not file-safe.", opts.CustomWorkflowPacket, "workflow.id"), nil
	}
	dag, ok := generatedContent(packet.MachinePacket.GeneratedContent, packet.MachinePacket.CandidatePaths.WorkflowDAG, "workflow_dag")
	if !ok || strings.TrimSpace(dag.Content) == "" {
		return fail(result, "custom_workflow_dag_missing", "custom workflow packet must include generated workflow_dag content.", opts.CustomWorkflowPacket, "machine_packet.generated_content"), nil
	}
	if dag.SHA256 != "" && dag.SHA256 != checksumBytes([]byte(dag.Content)) {
		return fail(result, "custom_workflow_dag_checksum_mismatch", "custom workflow DAG content checksum does not match packet evidence.", opts.CustomWorkflowPacket, "machine_packet.generated_content.sha256"), nil
	}
	contracts, err := contractsFromCustomPacket(packet.MachinePacket.NodeContracts)
	if err != nil {
		return fail(result, "custom_node_contracts_invalid", err.Error(), opts.CustomWorkflowPacket, "machine_packet.node_contracts"), nil
	}
	for _, contract := range contracts {
		if contract.WorkflowID != workflowID {
			return fail(result, "custom_node_contract_workflow_mismatch", "custom node contract workflow_id does not match approved workflow id.", opts.CustomWorkflowPacket, "machine_packet.node_contracts.workflow_id"), nil
		}
	}

	base := filepath.ToSlash(filepath.Join(".kkachi", "runs", opts.RunID, "workflow"))
	paths := map[string]string{
		"materialization":        filepath.ToSlash(filepath.Join(base, "materialization.json")),
		"workflow":               filepath.ToSlash(filepath.Join(base, "workflow.yaml")),
		"node_contracts":         filepath.ToSlash(filepath.Join(base, "node-contracts.json")),
		"custom_workflow_packet": filepath.ToSlash(filepath.Join(base, "custom-workflow-packet.json")),
		"checksums":              filepath.ToSlash(filepath.Join(base, "checksums.json")),
	}
	for _, path := range paths {
		if !safeRunLocalPath(opts.RunID, path) {
			return fail(result, "unsafe_run_local_path", "materialization path is outside the run-local workflow directory.", path, "path"), nil
		}
		if err := ensureSafeTarget(opts.Project, path); err != nil {
			return fail(result, "unsafe_run_local_path", err.Error(), path, "path"), nil
		}
	}

	bundle := contractBundle{SchemaVersion: workflowregistry.NodeContractsVersion, Ref: "run-local:" + opts.RunID + ":" + workflowID, Contracts: contracts}
	contractBytes, _ := json.MarshalIndent(bundle, "", "  ")
	contractBytes = append(contractBytes, '\n')

	result.OK = true
	result.Status = "materialized"
	result.WorkflowID = workflowID
	result.WorkflowFile = paths["workflow"]
	result.NodeContractSource = paths["node_contracts"]
	result.CustomWorkflowPacketCopy = paths["custom_workflow_packet"]
	result.MaterializationFile = paths["materialization"]
	result.ChecksumsFile = paths["checksums"]
	result.SelectedBundle = workflowID
	result.TaskClass = taskClassFromPacket(packet)
	result.ClassificationReason = "approved one-off custom workflow packet"
	result.SourceEvidence = SourceEvidence{
		CustomWorkflowPacketPath:     filepath.ToSlash(opts.CustomWorkflowPacket),
		CustomWorkflowPacketChecksum: checksumBytes(packetBytes),
		ApprovalEvidence:             opts.Approval,
		DryRunPlanHash:               recomputedHash,
		ApprovedPlanHash:             approvedHash,
	}
	result.Checksums = map[string]string{
		paths["workflow"]:                 checksumBytes([]byte(dag.Content)),
		paths["node_contracts"]:           checksumBytes(contractBytes),
		paths["custom_workflow_packet"]:   checksumBytes(packetBytes),
		"approval_hash":                   recomputedHash,
		"approved_plan_hash":              approvedHash,
		"custom_workflow_packet_checksum": checksumBytes(packetBytes),
	}
	result.WrittenPaths = []string{paths["materialization"], paths["workflow"], paths["node_contracts"], paths["custom_workflow_packet"], paths["checksums"]}
	sort.Strings(result.WrittenPaths)

	writes := map[string][]byte{
		paths["workflow"]:               []byte(dag.Content),
		paths["node_contracts"]:         contractBytes,
		paths["custom_workflow_packet"]: packetBytes,
		paths["checksums"]:              jsonWithNewline(result.Checksums),
		paths["materialization"]:        jsonWithNewline(result),
	}
	for _, path := range result.WrittenPaths {
		data := writes[path]
		if len(data) == 0 {
			continue
		}
		if err := writeProjectFile(opts.Project, path, data); err != nil {
			return fail(result, "materialization_write_failed", err.Error(), path, "path"), nil
		}
	}
	return result, nil
}

func normalizeOptions(opts Options) Options {
	opts.Project = strings.TrimSpace(opts.Project)
	if opts.Project != "" {
		if abs, err := filepath.Abs(opts.Project); err == nil {
			opts.Project = abs
		}
	}
	opts.RunID = strings.TrimSpace(opts.RunID)
	opts.RouteResult = strings.TrimSpace(opts.RouteResult)
	opts.CustomWorkflowPacket = strings.TrimSpace(opts.CustomWorkflowPacket)
	opts.Approval = strings.TrimSpace(opts.Approval)
	return opts
}

func newResult(opts Options) Result {
	return Result{
		Command:             Command,
		SchemaVersion:       SchemaVersion,
		Status:              "blocked",
		Project:             opts.Project,
		RunID:               opts.RunID,
		RunLocalPosture:     "ephemeral_run_local_only",
		NoPromotion:         true,
		PersistentPromotion: false,
		Checksums:           map[string]string{},
		WrittenPaths:        []string{},
		Diagnostics:         []Diagnostic{},
		ReasonCodes:         []string{},
		DirectKAHStateWrite: false,
	}
}

func fail(result Result, code string, message string, path string, field string) Result {
	result.OK = false
	result.Status = code
	result.ReasonCodes = appendUnique(append(result.ReasonCodes, code)...)
	result.Diagnostics = append(result.Diagnostics, Diagnostic{Level: "error", Code: code, Message: message, Path: path, Field: field})
	result.WrittenPaths = []string{}
	result.Checksums = map[string]string{}
	result.DirectKAHStateWrite = false
	return result
}

func findWorkflow(workflows []workflowregistry.Workflow, workflowID string) (workflowregistry.Workflow, bool) {
	for _, workflow := range workflows {
		if workflow.WorkflowID == workflowID {
			return workflow, true
		}
	}
	return workflowregistry.Workflow{}, false
}

func renderWorkflowDAG(workflowID string, contracts []workflowregistry.NodeContract) string {
	var b strings.Builder
	fmt.Fprintf(&b, "workflow_id: %s\nschema_version: task-dag/v1\nnodes:\n", workflowID)
	previous := ""
	for _, contract := range contracts {
		fmt.Fprintf(&b, "  - id: %s\n", contract.NodeID)
		if previous == "" {
			b.WriteString("    depends_on: []\n")
		} else {
			fmt.Fprintf(&b, "    depends_on: [%s]\n", previous)
		}
		b.WriteString("    join: all_of\n")
		b.WriteString("    required_outputs:\n")
		for _, artifact := range normalized(contract.ExpectedArtifacts) {
			fmt.Fprintf(&b, "      - %s\n", artifact)
		}
		previous = contract.NodeID
	}
	return b.String()
}

func approvedPlanHash(approval string) (string, bool) {
	const prefix = "dry-run:"
	approval = strings.TrimSpace(approval)
	if !strings.HasPrefix(approval, prefix) {
		return "", false
	}
	hash := strings.TrimPrefix(approval, prefix)
	if !strings.HasPrefix(hash, "sha256:") || len(hash) != len("sha256:")+64 {
		return "", false
	}
	if !isLowerHex64(strings.TrimPrefix(hash, "sha256:")) {
		return "", false
	}
	return hash, true
}

func isLowerHex64(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') {
			continue
		}
		return false
	}
	return true
}

func generatedContent(items []workflowcreator.GeneratedContent, path string, kind string) (workflowcreator.GeneratedContent, bool) {
	path = filepath.ToSlash(strings.TrimSpace(path))
	for _, item := range items {
		if filepath.ToSlash(strings.TrimSpace(item.Path)) == path && item.Kind == kind {
			return item, true
		}
	}
	return workflowcreator.GeneratedContent{}, false
}

func contractsFromCustomPacket(items []workflowcreator.NodeContract) ([]workflowregistry.NodeContract, error) {
	contracts := make([]workflowregistry.NodeContract, 0, len(items))
	falseValue := false
	for _, item := range items {
		if item.CompletionAuthority != workflowregistry.KAHOnlyAuthority {
			return nil, fmt.Errorf("node contract %s/%s requires completion_authority %s", item.WorkflowID, item.NodeID, workflowregistry.KAHOnlyAuthority)
		}
		if item.DirectKAHStateWrite {
			return nil, fmt.Errorf("node contract %s/%s must set direct_kah_state_write false", item.WorkflowID, item.NodeID)
		}
		if item.FallbackPolicy != workflowregistry.NoFallbackPolicy {
			return nil, fmt.Errorf("node contract %s/%s uses unsupported fallback_policy %q", item.WorkflowID, item.NodeID, item.FallbackPolicy)
		}
		contract := workflowregistry.NodeContract{
			WorkflowID:          item.WorkflowID,
			NodeID:              item.NodeID,
			TaskClass:           item.TaskClass,
			OwnerRole:           item.OwnerRole,
			ExecutionLane:       item.ExecutionLane,
			RequiredInputs:      normalized(item.RequiredInputs),
			ExpectedArtifacts:   normalized(item.ExpectedArtifacts),
			PromptRef:           item.PromptRef,
			ApprovalRequired:    item.ApprovalRequired,
			FallbackPolicy:      item.FallbackPolicy,
			VerificationGate:    item.VerificationGate,
			CompletionAuthority: item.CompletionAuthority,
			DirectKAHStateWrite: &falseValue,
		}
		contracts = append(contracts, contract)
	}
	if err := workflowregistry.ValidateNodeContracts(contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

func workflowIDFromContracts(contracts []workflowcreator.NodeContract) string {
	for _, contract := range contracts {
		if strings.TrimSpace(contract.WorkflowID) != "" {
			return strings.TrimSpace(contract.WorkflowID)
		}
	}
	return ""
}

func taskClassFromPacket(packet workflowcreator.Result) string {
	for _, contract := range packet.MachinePacket.NodeContracts {
		if strings.TrimSpace(contract.TaskClass) != "" {
			return strings.TrimSpace(contract.TaskClass)
		}
	}
	if value, ok := packet.MachinePacket.SelectorMetadata["task_class"].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
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
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	blocked := []string{
		".kkachi-workflow.yaml",
		".kkachi/workflow-catalog.yaml",
		".kkachi/workflows/",
		".hermes/",
		"auth/",
		"tokens/",
		"providers/",
		"gateway/",
		"models/",
		"kab/",
	}
	for _, prefix := range blocked {
		if path == strings.TrimSuffix(prefix, "/") || strings.HasPrefix(path, prefix) {
			return false
		}
	}
	return true
}

func ensureSafeTarget(project string, rel string) error {
	if !safeRelPath(rel) {
		return errors.New("path is not a safe repository-relative path")
	}
	root, err := filepath.Abs(project)
	if err != nil {
		return err
	}
	target := filepath.Join(root, filepath.FromSlash(rel))
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	if targetAbs != root && !strings.HasPrefix(targetAbs, root+string(os.PathSeparator)) {
		return errors.New("path escapes the project root")
	}
	current := root
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, part := range parts[:len(parts)-1] {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path parent %s is a symlink", filepath.ToSlash(strings.TrimPrefix(current, root+string(os.PathSeparator))))
		}
		if !info.IsDir() {
			return fmt.Errorf("path parent %s is not a directory", filepath.ToSlash(strings.TrimPrefix(current, root+string(os.PathSeparator))))
		}
	}
	if info, err := os.Lstat(targetAbs); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("target path is a symlink")
	}
	return nil
}

func writeProjectFile(project string, rel string, data []byte) error {
	if err := ensureSafeTarget(project, rel); err != nil {
		return err
	}
	path := filepath.Join(project, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func normalized(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
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

func safeID(id string) bool {
	if id == "" || id == "." || id == ".." || strings.ContainsAny(id, `/\`) {
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

func checksumBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func jsonWithNewline(value any) []byte {
	data, _ := json.MarshalIndent(value, "", "  ")
	return append(data, '\n')
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func appendUnique(values ...string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}
