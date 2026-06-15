package workflowtrigger

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowregistry"
)

const (
	Command = "workflow-trigger"
	Mode    = "explicit_workflow_trigger"
)

type Runner func(workDir string, args ...string) CommandResult

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type Options struct {
	Project              string
	WorkflowID           string
	NodeContractSource   string
	NodeContractRef      string
	SelectorRegistry     string
	TaskClass            string
	Labels               []string
	ChangedSurfaces      []string
	Risk                 string
	RequiredAgents       []string
	RequiredCapabilities []string
	RunID                string
	InstanceID           string
	Runner               Runner
}

type Result struct {
	OK                   bool             `json:"ok"`
	Command              string           `json:"command"`
	Mode                 string           `json:"mode"`
	Status               string           `json:"status"`
	Project              string           `json:"project"`
	WorkflowID           string           `json:"workflow_id"`
	Workflow             WorkflowEvidence `json:"workflow"`
	NodeContractSource   string           `json:"node_contract_source"`
	NodeContractRef      string           `json:"node_contract_ref,omitempty"`
	SelectorRegistry     RegistryEvidence `json:"selector_registry,omitempty"`
	SelectorMatch        SelectorEvidence `json:"selector_match,omitempty"`
	TaskClass            string           `json:"task_class,omitempty"`
	Labels               []string         `json:"labels,omitempty"`
	ChangedSurfaces      []string         `json:"changed_surfaces,omitempty"`
	RequiredCapabilities []string         `json:"required_capabilities,omitempty"`
	KAHCapability        KAHCapability    `json:"kah_capability"`
	Instance             InstanceEvidence `json:"instance"`
	ReadyNodes           []ReadyNode      `json:"ready_nodes"`
	DispatchPackets      []DispatchPacket `json:"dispatch_packets"`
	Diagnostics          []Diagnostic     `json:"diagnostics"`
	ReasonCodes          []string         `json:"reason_codes"`
	NextAction           string           `json:"next_action"`
	DirectKAHStateWrite  bool             `json:"direct_kah_state_write"`
}

type WorkflowEvidence struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type RegistryEvidence struct {
	Path     string `json:"path,omitempty"`
	Version  string `json:"version,omitempty"`
	Checksum string `json:"checksum,omitempty"`
}

type SelectorEvidence struct {
	Status       string   `json:"status,omitempty"`
	WorkflowID   string   `json:"workflow_id,omitempty"`
	CandidateIDs []string `json:"candidate_ids,omitempty"`
}

type KAHCapability struct {
	Available        bool     `json:"available"`
	Version          string   `json:"version,omitempty"`
	WorkflowCommands []string `json:"workflow_commands"`
	Reason           string   `json:"reason,omitempty"`
}

type InstanceEvidence struct {
	ID       string `json:"id,omitempty"`
	Source   string `json:"source,omitempty"`
	Revision int    `json:"revision,omitempty"`
}

type ReadyNode struct {
	ID      string   `json:"id"`
	Reasons []string `json:"reasons,omitempty"`
}

type DispatchPacket struct {
	WorkflowID                        string   `json:"workflow_id"`
	InstanceID                        string   `json:"instance_id"`
	NodeID                            string   `json:"node_id"`
	OwnerRole                         string   `json:"owner_role"`
	ExecutionLane                     string   `json:"execution_lane"`
	RequiredInputs                    []string `json:"required_inputs"`
	ExpectedArtifacts                 []string `json:"expected_artifacts"`
	PromptRef                         string   `json:"prompt_ref"`
	ApprovalRequired                  bool     `json:"approval_required"`
	FallbackPolicy                    string   `json:"fallback_policy"`
	VerificationGate                  string   `json:"verification_gate"`
	NodeContractSource                string   `json:"node_contract_source"`
	NodeContractRef                   string   `json:"node_contract_ref,omitempty"`
	SelectorRegistrySource            string   `json:"selector_registry_source,omitempty"`
	SelectorRegistryChecksum          string   `json:"selector_registry_checksum,omitempty"`
	SelectorMatch                     string   `json:"selector_match,omitempty"`
	TaskClass                         string   `json:"task_class,omitempty"`
	Labels                            []string `json:"labels,omitempty"`
	ChangedSurfaces                   []string `json:"changed_surfaces,omitempty"`
	RequiredCapabilities              []string `json:"required_capabilities,omitempty"`
	CompletionAuthority               string   `json:"completion_authority"`
	Stage1DirectCodexIsKABNativeCodex bool     `json:"stage1_direct_codex_is_kab_native_codex"`
	SourceChecksum                    string   `json:"source_checksum"`
	Status                            string   `json:"status"`
}

type Diagnostic struct {
	Level    string `json:"level"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	NodeID   string `json:"node_id,omitempty"`
	Field    string `json:"field,omitempty"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
}

type NodeContract = workflowregistry.NodeContract

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

func Trigger(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	result := newResult(opts)
	if opts.Project == "" {
		return fail(result, "project_required", "workflow-trigger requires --project <path>."), nil
	}
	selectorMode := hasSelectorInput(opts)
	explicitMode := strings.TrimSpace(opts.WorkflowID) != "" || strings.TrimSpace(opts.NodeContractSource) != ""
	var contracts []NodeContract
	checksum := ""
	if selectorMode && explicitMode {
		return fail(result, "selector_explicit_mode_conflict", "workflow-trigger accepts either explicit workflow/node-contract inputs or selector registry inputs, not both."), nil
	}
	if selectorMode {
		selectedContracts, registryChecksum, ok := resolveSelectorMode(&result, opts)
		if !ok {
			return result, nil
		}
		contracts = selectedContracts
		checksum = registryChecksum
		opts.WorkflowID = result.WorkflowID
		opts.NodeContractSource = result.NodeContractSource
	} else {
		if strings.TrimSpace(opts.WorkflowID) == "" {
			return fail(result, "workflow_id_required", "workflow-trigger requires --workflow-id <id>."), nil
		}
		if !safeWorkflowID(opts.WorkflowID) {
			return fail(result, "workflow_id_invalid", "workflow id must be a simple file-safe id."), nil
		}
		result.WorkflowID = opts.WorkflowID
		result.Workflow = WorkflowEvidence{ID: opts.WorkflowID, Path: workflowPath(opts.WorkflowID)}
		if strings.TrimSpace(opts.NodeContractSource) == "" {
			return fail(result, "node_contract_source_required", "workflow-trigger requires --node-contract-source <path>."), nil
		}
		if strings.TrimSpace(opts.RunID) == "" && strings.TrimSpace(opts.InstanceID) == "" {
			return fail(result, "run_or_instance_required", "workflow-trigger requires --run <run-id> for create or --instance-id <id> for resume."), nil
		}
		loadedContracts, loadedChecksum, loadErr := loadNodeContracts(opts.NodeContractSource, opts.NodeContractRef)
		if loadErr != nil {
			return fail(result, loadErr.Code, loadErr.Message), nil
		}
		contracts = loadedContracts
		checksum = loadedChecksum
		result.NodeContractSource = opts.NodeContractSource
		result.NodeContractRef = opts.NodeContractRef
	}
	if strings.TrimSpace(opts.RunID) == "" && strings.TrimSpace(opts.InstanceID) == "" {
		return fail(result, "run_or_instance_required", "workflow-trigger requires --run <run-id> for create or --instance-id <id> for resume."), nil
	}

	preflight := preflightKAH(opts)
	result.KAHCapability = preflight.capability
	if preflight.err != nil {
		return fail(result, "blocked_missing_kah_workflow_capability", preflight.err.Error()), nil
	}

	explainNodeIDs, explainHadNodeIDs, ok := runWorkflowValidateExplain(&result, opts)
	if !ok {
		return result, nil
	}
	if selectorMode {
		if explainHadNodeIDs {
			if err := workflowregistry.ValidateContractsCoverNodeIDs(contracts, result.WorkflowID, explainNodeIDs); err != nil {
				return fail(result, "blocked_registry_node_contract_mismatch", err.Error()), nil
			}
		} else {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Level: "info", Code: "kah_explain_node_id_readback_unavailable", Message: "KAH workflow explain did not expose reliable node ids; ready-node contract checks remain enforced."})
		}
	}

	runID := strings.TrimSpace(opts.InstanceID)
	source := "existing"
	if runID == "" {
		runID = strings.TrimSpace(opts.RunID)
		source = "created"
		if !runWorkflowCreate(&result, opts, runID) {
			return result, nil
		}
	} else if !runWorkflowShow(&result, opts, runID) {
		return result, nil
	}
	result.Instance.ID = runID
	result.Instance.Source = source

	ready, ok := runWorkflowReady(&result, opts, runID)
	if !ok {
		return result, nil
	}
	result.ReadyNodes = ready
	if len(ready) == 0 {
		result.OK = true
		result.Status = "no_ready_nodes"
		result.NextAction = "No KAH-ready workflow nodes are available; wait for dependencies or inspect KAH workflow state."
		return result, nil
	}

	packets := []DispatchPacket{}
	for _, node := range ready {
		contract, ok := findContract(contracts, opts.WorkflowID, node.ID)
		if !ok {
			result = fail(result, "blocked_missing_ready_node_contract", "ready node has no matching explicit node contract.")
			result.Diagnostics[len(result.Diagnostics)-1].NodeID = node.ID
			return result, nil
		}
		packets = append(packets, packetFromContract(contract, runID, opts.NodeContractSource, opts.NodeContractRef, checksum, result))
	}
	result.DispatchPackets = packets
	result.OK = true
	result.Status = "dispatch_packets_rendered"
	result.NextAction = "Dispatch packets rendered only; execute through the declared lane and let KAH advance node state after evidence exists."
	return result, nil
}

func RenderHuman(result Result) string {
	if result.OK {
		return fmt.Sprintf("Status: %s\nWorkflow: %s\nReady nodes: %d\nDispatch packets: %d\nDirect KAH state write: false\nNext: %s", result.Status, result.WorkflowID, len(result.ReadyNodes), len(result.DispatchPackets), result.NextAction)
	}
	return fmt.Sprintf("Status: %s\nWorkflow: %s\nDirect KAH state write: false\nNext: %s", result.Status, result.WorkflowID, result.NextAction)
}

func normalizeOptions(opts Options) Options {
	if opts.Runner == nil {
		opts.Runner = commandRunner{}.Run
	}
	if opts.Project != "" {
		if abs, err := filepath.Abs(opts.Project); err == nil {
			opts.Project = abs
		}
	}
	opts.Labels = workflowregistry.NormalizeSelectorValues(opts.Labels)
	opts.ChangedSurfaces = workflowregistry.NormalizeSelectorValues(opts.ChangedSurfaces)
	opts.RequiredAgents = workflowregistry.NormalizeSelectorValues(opts.RequiredAgents)
	opts.RequiredCapabilities = workflowregistry.NormalizeSelectorValues(opts.RequiredCapabilities)
	opts.TaskClass = strings.TrimSpace(opts.TaskClass)
	opts.Risk = strings.TrimSpace(opts.Risk)
	opts.SelectorRegistry = strings.TrimSpace(opts.SelectorRegistry)
	return opts
}

func newResult(opts Options) Result {
	return Result{
		Command:              Command,
		Mode:                 Mode,
		Status:               "blocked",
		Project:              opts.Project,
		WorkflowID:           opts.WorkflowID,
		NodeContractSource:   opts.NodeContractSource,
		NodeContractRef:      opts.NodeContractRef,
		TaskClass:            opts.TaskClass,
		Labels:               opts.Labels,
		ChangedSurfaces:      opts.ChangedSurfaces,
		RequiredCapabilities: opts.RequiredCapabilities,
		ReadyNodes:           []ReadyNode{},
		DispatchPackets:      []DispatchPacket{},
		Diagnostics:          []Diagnostic{},
		ReasonCodes:          []string{},
		DirectKAHStateWrite:  false,
	}
}

func fail(result Result, code string, message string) Result {
	result.OK = false
	if result.Status == "" || result.Status == "blocked" {
		result.Status = code
	}
	result.ReasonCodes = appendUnique(append(result.ReasonCodes, code)...)
	result.Diagnostics = append(result.Diagnostics, Diagnostic{Level: "error", Code: code, Message: message})
	result.NextAction = "Fix the reported workflow trigger blocker and rerun with explicit workflow and node-contract inputs."
	result.DispatchPackets = []DispatchPacket{}
	return result
}

type codedError struct {
	Code    string
	Message string
}

func (e codedError) Error() string {
	return e.Message
}

func loadNodeContracts(path string, expectedRef string) ([]NodeContract, string, *codedError) {
	contracts, checksum, err := workflowregistry.LoadJSONContracts(path, expectedRef)
	if err != nil {
		return nil, "", mapNodeContractError(err)
	}
	return contracts, checksum, nil
}

func mapNodeContractError(err error) *codedError {
	message := err.Error()
	switch {
	case strings.Contains(message, "unreadable"):
		return &codedError{Code: "node_contract_source_unreadable", Message: message}
	case strings.Contains(message, "must be JSON"):
		return &codedError{Code: "node_contract_source_invalid_json", Message: message}
	case strings.Contains(message, "schema_version"):
		return &codedError{Code: "node_contract_schema_unsupported", Message: message}
	case strings.Contains(message, "ref does not match"):
		return &codedError{Code: "node_contract_ref_mismatch", Message: message}
	case strings.Contains(message, "at least one contract"):
		return &codedError{Code: "node_contracts_empty", Message: message}
	default:
		return &codedError{Code: "node_contract_required_field_missing", Message: message}
	}
}

func hasSelectorInput(opts Options) bool {
	return opts.SelectorRegistry != "" || opts.TaskClass != "" || opts.Risk != "" || len(opts.Labels) > 0 || len(opts.ChangedSurfaces) > 0 || len(opts.RequiredAgents) > 0 || len(opts.RequiredCapabilities) > 0
}

func resolveSelectorMode(result *Result, opts Options) ([]NodeContract, string, bool) {
	if opts.SelectorRegistry == "" {
		*result = fail(*result, "selector_registry_required", "selector mode requires --selector-registry <path>.")
		return nil, "", false
	}
	registry, err := workflowregistry.Load(opts.SelectorRegistry)
	if err != nil {
		*result = fail(*result, "selector_registry_schema_unsupported", err.Error())
		if strings.Contains(err.Error(), "unreadable") {
			result.Status = "selector_registry_unreadable"
			result.ReasonCodes = []string{"selector_registry_unreadable"}
			result.Diagnostics = []Diagnostic{{Level: "error", Code: "selector_registry_unreadable", Message: err.Error()}}
		}
		return nil, "", false
	}
	result.SelectorRegistry = RegistryEvidence{Path: registry.Path, Version: registry.Version, Checksum: registry.Checksum}
	query := workflowregistry.Query{TaskClass: opts.TaskClass, Labels: opts.Labels, ChangedSurfaces: opts.ChangedSurfaces, Risk: opts.Risk, RequiredAgents: opts.RequiredAgents, RequiredCapabilities: opts.RequiredCapabilities}
	match, err := workflowregistry.Select(registry, query)
	result.TaskClass = match.Query.TaskClass
	result.Labels = match.Query.Labels
	result.ChangedSurfaces = match.Query.ChangedSurfaces
	result.RequiredCapabilities = match.Query.RequiredCapabilities
	result.SelectorMatch = SelectorEvidence{Status: match.Status, CandidateIDs: workflowCandidateIDs(match.Candidates)}
	if err != nil {
		*result = fail(*result, match.Status, err.Error())
		return nil, "", false
	}
	switch match.Status {
	case "selector_no_match":
		*result = fail(*result, "selector_no_match", "selector registry did not match any workflow.")
		return nil, "", false
	case "selector_ambiguous":
		*result = fail(*result, "selector_ambiguous", "selector registry matched multiple workflows; narrow selector metadata or provide explicit workflow inputs.")
		return nil, "", false
	case "selector_matched":
		result.WorkflowID = match.Selected.WorkflowID
		result.Workflow = WorkflowEvidence{ID: match.Selected.WorkflowID, Path: match.Selected.WorkflowPath}
		result.SelectorMatch.WorkflowID = match.Selected.WorkflowID
	default:
		*result = fail(*result, "selector_registry_schema_unsupported", "selector returned unsupported status.")
		return nil, "", false
	}
	contracts := workflowregistry.ContractsForWorkflow(registry.NodeContracts, match.Selected.WorkflowID)
	if len(contracts) == 0 {
		*result = fail(*result, "blocked_missing_ready_node_contract", "selected workflow has no node contracts in selector registry.")
		return nil, "", false
	}
	result.NodeContractSource = registry.Path
	opts.NodeContractSource = registry.Path
	return contracts, registry.Checksum, true
}

func workflowCandidateIDs(candidates []workflowregistry.Workflow) []string {
	ids := []string{}
	for _, candidate := range candidates {
		ids = append(ids, candidate.WorkflowID)
	}
	sort.Strings(ids)
	return ids
}

type preflightResult struct {
	capability KAHCapability
	err        error
}

func preflightKAH(opts Options) preflightResult {
	capability := KAHCapability{WorkflowCommands: []string{}}
	version := opts.Runner(opts.Project, "--version")
	if version.Err != nil {
		return preflightResult{capability: capability, err: errors.New("kkachi-agent-helper --version failed")}
	}
	capability.Version = strings.TrimSpace(string(version.Stdout))

	caps := opts.Runner(opts.Project, "capabilities", "--json")
	if caps.Err != nil {
		return preflightResult{capability: capability, err: errors.New("kkachi-agent-helper capabilities --json failed")}
	}
	commands, flagsOK := parseWorkflowCapabilities(caps.Stdout)
	capability.WorkflowCommands = commands

	help := opts.Runner(opts.Project, "workflow", "--help")
	if help.Err != nil {
		capability.Reason = "workflow_help_unavailable"
		return preflightResult{capability: capability, err: errors.New("KAH workflow help is unavailable")}
	}
	helpText := string(help.Stdout)
	for _, required := range requiredWorkflowCommands() {
		if !contains(commands, required) || !strings.Contains(helpText, required) {
			capability.Reason = "workflow_subcommand_missing"
			return preflightResult{capability: capability, err: fmt.Errorf("KAH workflow subcommand %s is missing", required)}
		}
	}
	if !flagsOK {
		capability.Reason = "workflow_flags_missing"
		return preflightResult{capability: capability, err: errors.New("KAH workflow instance capability flags are missing")}
	}
	capability.Available = true
	return preflightResult{capability: capability}
}

func parseWorkflowCapabilities(data []byte) ([]string, bool) {
	var payload struct {
		CommandGroups []struct {
			Name        string   `json:"name"`
			Status      string   `json:"status"`
			Subcommands []string `json:"subcommands"`
		} `json:"command_groups"`
		Commands           []string        `json:"commands"`
		CompatibilityFlags map[string]bool `json:"compatibility_flags"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, false
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
	flagsOK := payload.CompatibilityFlags["task_dag_schema_validation"] && payload.CompatibilityFlags["workflow_instance_state"]
	return appendUnique(commands...), flagsOK
}

func runWorkflowValidateExplain(result *Result, opts Options) ([]string, bool, bool) {
	explainNodeIDs := []string{}
	explainHadNodeIDs := false
	for _, subcommand := range []string{"validate", "explain"} {
		command := opts.Runner(opts.Project, "workflow", subcommand, "--file", result.Workflow.Path, "--json")
		payload, ok := parseKAHJSON(result, command, "blocked_kah_workflow_"+subcommand+"_failed", "KAH workflow "+subcommand+" failed.")
		if !ok {
			return nil, false, false
		}
		if !truthy(payload["ok"]) {
			*result = fail(*result, "blocked_kah_workflow_"+subcommand+"_failed", "KAH workflow "+subcommand+" returned ok:false.")
			return nil, false, false
		}
		if id, _ := payload["workflow_id"].(string); id != "" && id != opts.WorkflowID {
			*result = fail(*result, "workflow_id_mismatch", "KAH workflow "+subcommand+" returned a different workflow_id.")
			return nil, false, false
		}
		if subcommand == "explain" {
			explainNodeIDs, explainHadNodeIDs = extractExplainNodeIDs(payload)
		}
	}
	return explainNodeIDs, explainHadNodeIDs, true
}

func extractExplainNodeIDs(payload map[string]any) ([]string, bool) {
	for _, source := range []any{payload["nodes"], nestedMapValue(payload, "workflow", "nodes")} {
		items := list(source)
		if len(items) == 0 {
			continue
		}
		ids := []string{}
		reliable := true
		for _, raw := range items {
			item, ok := raw.(map[string]any)
			if !ok {
				reliable = false
				break
			}
			id, _ := item["id"].(string)
			if id == "" {
				reliable = false
				break
			}
			ids = append(ids, id)
		}
		if reliable {
			sort.Strings(ids)
			return ids, true
		}
	}
	return nil, false
}

func nestedMapValue(payload map[string]any, parent string, child string) any {
	item, _ := payload[parent].(map[string]any)
	if item == nil {
		return nil
	}
	return item[child]
}

func runWorkflowCreate(result *Result, opts Options, runID string) bool {
	command := opts.Runner(opts.Project, "workflow", "create", "--run", runID, "--file", result.Workflow.Path, "--json")
	payload, ok := parseKAHJSON(result, command, "blocked_kah_workflow_create_failed", "KAH workflow create failed.")
	if !ok {
		return false
	}
	return normalizeInstanceResult(result, payload, "blocked_kah_workflow_create_failed")
}

func runWorkflowShow(result *Result, opts Options, runID string) bool {
	command := opts.Runner(opts.Project, "workflow", "show", "--run", runID, "--json")
	payload, ok := parseKAHJSON(result, command, "blocked_kah_workflow_show_failed", "KAH workflow show failed.")
	if !ok {
		return false
	}
	return normalizeInstanceResult(result, payload, "blocked_kah_workflow_show_failed")
}

func runWorkflowReady(result *Result, opts Options, runID string) ([]ReadyNode, bool) {
	command := opts.Runner(opts.Project, "workflow", "ready", "--run", runID, "--json")
	payload, ok := parseKAHJSON(result, command, "blocked_kah_workflow_ready_failed", "KAH workflow ready failed.")
	if !ok {
		return nil, false
	}
	if !truthy(payload["ok"]) {
		*result = fail(*result, "blocked_kah_workflow_ready_failed", "KAH workflow ready returned ok:false.")
		return nil, false
	}
	ready := []ReadyNode{}
	for _, raw := range list(payload["ready"]) {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id, _ := item["id"].(string)
		if id == "" {
			continue
		}
		ready = append(ready, ReadyNode{ID: id, Reasons: stringList(item["reasons"])})
	}
	return ready, true
}

func parseKAHJSON(result *Result, command CommandResult, code string, message string) (map[string]any, bool) {
	if command.Err != nil {
		*result = fail(*result, code, message)
		return nil, false
	}
	var payload map[string]any
	if err := json.Unmarshal(command.Stdout, &payload); err != nil {
		*result = fail(*result, code, message+" JSON output was not parseable.")
		return nil, false
	}
	return payload, true
}

func normalizeInstanceResult(result *Result, payload map[string]any, code string) bool {
	if !truthy(payload["ok"]) {
		*result = fail(*result, code, "KAH workflow instance command returned ok:false.")
		return false
	}
	instance, _ := payload["instance"].(map[string]any)
	if instance == nil {
		return true
	}
	if id, _ := instance["run_id"].(string); id != "" {
		result.Instance.ID = id
	}
	if source, _ := instance["source_path"].(string); source != "" {
		result.Instance.Source = source
	}
	if revision, ok := instance["revision"].(float64); ok {
		result.Instance.Revision = int(revision)
	}
	return true
}

func findContract(contracts []NodeContract, workflowID string, nodeID string) (NodeContract, bool) {
	for _, contract := range contracts {
		if contract.WorkflowID == workflowID && contract.NodeID == nodeID {
			return contract, true
		}
	}
	return NodeContract{}, false
}

func packetFromContract(contract NodeContract, instanceID string, source string, ref string, checksum string, result Result) DispatchPacket {
	return DispatchPacket{
		WorkflowID:                        contract.WorkflowID,
		InstanceID:                        instanceID,
		NodeID:                            contract.NodeID,
		OwnerRole:                         contract.OwnerRole,
		ExecutionLane:                     contract.ExecutionLane,
		RequiredInputs:                    contract.RequiredInputs,
		ExpectedArtifacts:                 contract.ExpectedArtifacts,
		PromptRef:                         contract.PromptRef,
		ApprovalRequired:                  contract.ApprovalRequired,
		FallbackPolicy:                    contract.FallbackPolicy,
		VerificationGate:                  contract.VerificationGate,
		NodeContractSource:                source,
		NodeContractRef:                   ref,
		SelectorRegistrySource:            result.SelectorRegistry.Path,
		SelectorRegistryChecksum:          result.SelectorRegistry.Checksum,
		SelectorMatch:                     result.SelectorMatch.Status,
		TaskClass:                         contract.TaskClass,
		Labels:                            result.Labels,
		ChangedSurfaces:                   result.ChangedSurfaces,
		RequiredCapabilities:              result.RequiredCapabilities,
		CompletionAuthority:               "kah_only",
		Stage1DirectCodexIsKABNativeCodex: false,
		SourceChecksum:                    checksum,
		Status:                            "ready_for_declared_lane",
	}
}

func workflowPath(workflowID string) string {
	return filepath.ToSlash(filepath.Join(".kkachi", "workflows", workflowID+".yaml"))
}

func safeWorkflowID(id string) bool {
	if id == "." || id == ".." || strings.Contains(id, "/") || strings.Contains(id, "\\") {
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

func requiredWorkflowCommands() []string {
	return []string{"validate", "explain", "create", "show", "ready", "node"}
}

func truthy(value any) bool {
	v, ok := value.(bool)
	return ok && v
}

func list(value any) []any {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	return items
}

func stringList(value any) []string {
	items := []string{}
	for _, raw := range list(value) {
		if item, ok := raw.(string); ok {
			items = append(items, item)
		}
	}
	return items
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
