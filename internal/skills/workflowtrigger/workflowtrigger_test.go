package workflowtrigger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowcreator"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowregistry"
)

type fakeKAHRunner struct {
	responses map[string]CommandResult
	calls     []string
}

func (r *fakeKAHRunner) Run(workDir string, args ...string) CommandResult {
	key := strings.Join(args, " ")
	r.calls = append(r.calls, key)
	if result, ok := r.responses[key]; ok {
		return result
	}
	return CommandResult{Stderr: []byte("unexpected call: " + key), Err: errors.New("unexpected call")}
}

func TestTriggerRendersDispatchPacketsFromExplicitWorkflowAndContract(t *testing.T) {
	project := t.TempDir()
	source := writeNodeContractBundle(t, project, `{
  "schema_version": "kas-node-contracts/v1",
  "ref": "contracts-a",
  "contracts": [
    {
      "workflow_id": "demo",
      "node_id": "setup",
      "owner_role": "implementer_backend",
      "execution_lane": "direct_kas_skill",
      "required_inputs": ["task-contract.yaml"],
      "expected_artifacts": ["artifacts/setup.md"],
      "prompt_ref": "skills/kkachi-implement/SKILL.md",
      "approval_required": true,
      "fallback_policy": "none_fail_closed",
      "verification_gate": "make test"
    }
  ]
}`)
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})

	result, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		NodeContractSource: source,
		NodeContractRef:    "contracts-a",
		RunID:              "run-20260615T010203Z-abcdef123456",
		Runner:             runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Status != "dispatch_packets_rendered" || result.DirectKAHStateWrite {
		t.Fatalf("unexpected result status/no-write: %+v", result)
	}
	if result.Workflow.Path != ".kkachi/workflows/demo.yaml" {
		t.Fatalf("workflow path = %q", result.Workflow.Path)
	}
	if len(result.DispatchPackets) != 1 {
		t.Fatalf("dispatch packets = %+v", result.DispatchPackets)
	}
	packet := result.DispatchPackets[0]
	if !packet.StrictOrder || packet.WorkflowID != "demo" || packet.RunID != "run-20260615T010203Z-abcdef123456" || packet.InstanceID != "run-20260615T010203Z-abcdef123456" || packet.NodeID != "setup" {
		t.Fatalf("unexpected packet identity: %+v", packet)
	}
	if packet.InstanceRevision != 1 {
		t.Fatalf("packet instance_revision = %d, want 1", packet.InstanceRevision)
	}
	if packet.OwnerRole != "implementer_backend" || packet.ExecutionLane != "direct_kas_skill" || !packet.ApprovalRequired {
		t.Fatalf("unexpected packet contract fields: %+v", packet)
	}
	if packet.FallbackPolicy != "none_fail_closed" || packet.NodeContractRef != "contracts-a" || packet.SourceChecksum == "" {
		t.Fatalf("packet missing source/fallback evidence: %+v", packet)
	}
	if packet.ExpectedStartRevision != 1 {
		t.Fatalf("packet expected_start_revision = %d, want ready instance revision 1", packet.ExpectedStartRevision)
	}
	if packet.InstanceRevision != packet.ExpectedStartRevision {
		t.Fatalf("packet instance_revision and expected_start_revision diverged: %+v", packet)
	}
	if !reflect.DeepEqual(packet.ReadyNodeReasons, []string{"dependencies_satisfied", "state_pending"}) {
		t.Fatalf("packet missing ready-node evidence: %+v", packet)
	}
	if !reflect.DeepEqual(packet.RequiredOutputs, []string{"artifacts/setup.md"}) {
		t.Fatalf("packet missing KAH required_outputs evidence: %+v", packet)
	}
	wantCalls := []string{
		"--version",
		"capabilities --json",
		"workflow --help",
		"workflow validate --file .kkachi/workflows/demo.yaml --json",
		"workflow explain --file .kkachi/workflows/demo.yaml --json",
		"workflow create --run run-20260615T010203Z-abcdef123456 --file .kkachi/workflows/demo.yaml --json",
		"workflow ready --run run-20260615T010203Z-abcdef123456 --json",
	}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestTriggerFailsClosedWhenWorkflowCapabilityMissing(t *testing.T) {
	project := t.TempDir()
	source := writeNodeContractBundle(t, project, validNodeContractBundle("demo", "setup"))
	runner := &fakeKAHRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.9\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"graph","status":"supported","subcommands":["validate"]}],"compatibility_flags":{"workflow_instance_state":false}}`)},
		"workflow --help":     {Stderr: []byte("unknown help topic\n"), Err: errors.New("exit 2")},
	}}

	result, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		NodeContractSource: source,
		RunID:              "run-20260615T010203Z-abcdef123456",
		Runner:             runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "blocked_missing_kah_workflow_capability" {
		t.Fatalf("expected capability blocker, got %+v", result)
	}
	if result.DirectKAHStateWrite {
		t.Fatalf("trigger must never report direct KAH writes: %+v", result)
	}
	if len(runner.calls) != 3 {
		t.Fatalf("expected preflight-only calls, got %#v", runner.calls)
	}
	for _, call := range runner.calls {
		if strings.Contains(call, "workflow create") || strings.Contains(call, "workflow show") || strings.Contains(call, "workflow ready") {
			t.Fatalf("capability blocker called workflow instance command: %#v", runner.calls)
		}
	}
}

func TestTriggerResumeUsesShowInsteadOfCreate(t *testing.T) {
	project := t.TempDir()
	source := writeNodeContractBundle(t, project, validNodeContractBundle("demo", "setup"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})

	result, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		NodeContractSource: source,
		InstanceID:         "run-20260615T010203Z-abcdef123456",
		Runner:             runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Instance.ID != "run-20260615T010203Z-abcdef123456" {
		t.Fatalf("unexpected resume result: %+v", result)
	}
	for _, call := range runner.calls {
		if strings.Contains(call, "workflow create") {
			t.Fatalf("resume must not create: %#v", runner.calls)
		}
	}
	if !containsCall(runner.calls, "workflow show --run run-20260615T010203Z-abcdef123456 --json") {
		t.Fatalf("resume did not call show: %#v", runner.calls)
	}
}

func TestTriggerNoReadyNodesIsSuccessfulNoPacketState(t *testing.T) {
	project := t.TempDir()
	source := writeNodeContractBundle(t, project, validNodeContractBundle("demo", "setup"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", nil)

	result, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		NodeContractSource: source,
		RunID:              "run-20260615T010203Z-abcdef123456",
		Runner:             runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Status != "no_ready_nodes" {
		t.Fatalf("expected no-ready success, got %+v", result)
	}
	if len(result.DispatchPackets) != 0 {
		t.Fatalf("no-ready state should emit no packets: %+v", result.DispatchPackets)
	}
}

func TestTriggerFailsClosedForReadyNodeWithoutMatchingContract(t *testing.T) {
	project := t.TempDir()
	source := writeNodeContractBundle(t, project, validNodeContractBundle("demo", "other"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})

	result, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		NodeContractSource: source,
		RunID:              "run-20260615T010203Z-abcdef123456",
		Runner:             runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "blocked_missing_ready_node_contract" {
		t.Fatalf("expected missing contract blocker, got %+v", result)
	}
	if len(result.DispatchPackets) != 0 {
		t.Fatalf("blocker must not render partial packets: %+v", result.DispatchPackets)
	}
}

func TestTriggerSelectorConflictFailsBeforeKAHCalls(t *testing.T) {
	project := t.TempDir()
	registry := writeRegistry(t, project, validRegistry("demo", "setup"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})

	result, err := Trigger(Options{
		Project:          project,
		WorkflowID:       "demo",
		SelectorRegistry: registry,
		TaskClass:        "development",
		RunID:            "run-20260615T010203Z-abcdef123456",
		Runner:           runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "selector_explicit_mode_conflict" {
		t.Fatalf("expected selector conflict, got %+v", result)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("conflict must not call KAH: %#v", runner.calls)
	}
}

func TestTriggerSelectorNoMatchAndAmbiguousDoNotCallKAH(t *testing.T) {
	project := t.TempDir()
	registry := writeRegistry(t, project, validRegistry("demo", "setup")+strings.ReplaceAll(validRegistry("other", "setup"), "version: kas-task-dag-workflow-registry/v1\n", ""))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})

	result, err := Trigger(Options{
		Project:          project,
		SelectorRegistry: registry,
		TaskClass:        "security",
		RunID:            "run-20260615T010203Z-abcdef123456",
		Runner:           runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "selector_no_match" {
		t.Fatalf("expected no match, got %+v", result)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("no-match must not call KAH: %#v", runner.calls)
	}

	result, err = Trigger(Options{
		Project:              project,
		SelectorRegistry:     registry,
		TaskClass:            "development",
		Labels:               []string{"stage1", "backend"},
		ChangedSurfaces:      []string{"code"},
		RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"},
		RunID:                "run-20260615T010203Z-abcdef123456",
		Runner:               runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "selector_ambiguous" || len(result.SelectorMatch.CandidateIDs) != 2 {
		t.Fatalf("expected ambiguity, got %+v", result)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("ambiguity must not call KAH: %#v", runner.calls)
	}
}

func TestTriggerSelectorMatchRendersRegistryReadbackPacket(t *testing.T) {
	project := t.TempDir()
	registry := writeRegistry(t, project, validRegistry("demo", "setup"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})
	runner.responses["workflow explain --file .kkachi/workflows/demo.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"demo","schema_version":"task-dag/v1","nodes":[{"id":"setup"}]}`)}

	result, err := Trigger(Options{
		Project:              project,
		SelectorRegistry:     registry,
		TaskClass:            "development",
		Labels:               []string{"stage1", "backend"},
		ChangedSurfaces:      []string{"code"},
		RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"},
		RunID:                "run-20260615T010203Z-abcdef123456",
		Runner:               runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.WorkflowID != "demo" || result.SelectorMatch.WorkflowID != "demo" {
		t.Fatalf("unexpected selector success: %+v", result)
	}
	if len(result.DispatchPackets) != 1 {
		t.Fatalf("dispatch packets = %+v", result.DispatchPackets)
	}
	packet := result.DispatchPackets[0]
	if packet.SelectorRegistrySource == "" || packet.SelectorRegistryChecksum == "" || packet.CompletionAuthority != "kah_only" || packet.Stage1DirectCodexIsKABNativeCodex {
		t.Fatalf("missing selector/readback authority fields: %+v", packet)
	}
	if packet.TaskClass != "development" || packet.SelectorMatch != "selector_matched" {
		t.Fatalf("missing selector packet metadata: %+v", packet)
	}
	if !reflect.DeepEqual(packet.Labels, []string{"backend", "stage1"}) {
		t.Fatalf("missing selector labels readback: %+v", packet)
	}
	if !reflect.DeepEqual(packet.ChangedSurfaces, []string{"code"}) {
		t.Fatalf("missing changed surfaces readback: %+v", packet)
	}
	if !reflect.DeepEqual(packet.RequiredCapabilities, []string{"task_dag_schema_validation", "workflow_instance_state"}) {
		t.Fatalf("missing required capabilities readback: %+v", packet)
	}
	if result.DirectKAHStateWrite {
		t.Fatalf("selector trigger must never report direct KAH state writes: %+v", result)
	}
}

func TestTriggerSelectorExplainNodeIDMismatchFailsClosed(t *testing.T) {
	project := t.TempDir()
	registry := writeRegistry(t, project, validRegistry("demo", "setup"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})
	runner.responses["workflow explain --file .kkachi/workflows/demo.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"demo","schema_version":"task-dag/v1","nodes":[{"id":"different"}]}`)}

	result, err := Trigger(Options{
		Project:              project,
		SelectorRegistry:     registry,
		TaskClass:            "development",
		ChangedSurfaces:      []string{"code"},
		RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"},
		RunID:                "run-20260615T010203Z-abcdef123456",
		Runner:               runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "blocked_registry_node_contract_mismatch" {
		t.Fatalf("expected node-id mismatch blocker, got %+v", result)
	}
	if containsCall(runner.calls, "workflow create --run run-20260615T010203Z-abcdef123456 --file .kkachi/workflows/demo.yaml --json") {
		t.Fatalf("node-id mismatch must not create workflow instance: %#v", runner.calls)
	}
}

func TestTriggerSelectorExplainNodeIDUnavailableIsInformational(t *testing.T) {
	project := t.TempDir()
	registry := writeRegistry(t, project, validRegistry("demo", "setup"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})
	runner.responses["workflow explain --file .kkachi/workflows/demo.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"demo","schema_version":"task-dag/v1","nodes":[{"name":"setup"}]}`)}

	result, err := Trigger(Options{
		Project:              project,
		SelectorRegistry:     registry,
		TaskClass:            "development",
		Labels:               []string{"stage1", "backend"},
		ChangedSurfaces:      []string{"code"},
		RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"},
		RunID:                "run-20260615T010203Z-abcdef123456",
		Runner:               runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Status != "dispatch_packets_rendered" {
		t.Fatalf("unavailable explain node ids should not block ready-node contract enforcement: %+v", result)
	}
	found := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Level == "info" && diagnostic.Code == "kah_explain_node_id_readback_unavailable" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing informational explain-node-id diagnostic: %+v", result.Diagnostics)
	}
}

func TestTriggerExplicitWorkflowFileUsesProvidedPath(t *testing.T) {
	project := t.TempDir()
	source := writeNodeContractBundle(t, project, validNodeContractBundle("demo", "setup"))
	workflowFile := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml"
	runner := newWorkflowFakeRunnerForFile("demo", "run-20260616T105614Z-4b0ebe11b67d", workflowFile, []string{"setup"})

	result, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		WorkflowFile:       workflowFile,
		NodeContractSource: source,
		RunID:              "run-20260616T105614Z-4b0ebe11b67d",
		Runner:             runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Workflow.Path != workflowFile {
		t.Fatalf("unexpected explicit workflow-file result: %+v", result)
	}
	if !containsCall(runner.calls, "workflow validate --file "+workflowFile+" --json") ||
		!containsCall(runner.calls, "workflow create --run run-20260616T105614Z-4b0ebe11b67d --file "+workflowFile+" --json") {
		t.Fatalf("KAH calls did not use explicit workflow file: %#v", runner.calls)
	}
	packet := result.DispatchPackets[0]
	if packet.WorkflowFile != workflowFile || packet.DirectKAHStateWrite {
		t.Fatalf("dispatch packet missing workflow file or no-write evidence: %+v", packet)
	}
}

func TestTriggerRunLocalMaterializationPreflightsBeforeWrite(t *testing.T) {
	project := t.TempDir()
	registryPath := writeRegistry(t, project, validRegistry("demo", "setup"))
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeTriggerRouteResult(t, project, registry, "demo")
	runner := &fakeKAHRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.9\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"graph","status":"supported","subcommands":["validate"]}],"compatibility_flags":{"workflow_instance_state":false}}`)},
		"workflow --help":     {Stderr: []byte("unknown help topic\n"), Err: errors.New("exit 2")},
	}}

	result, err := Trigger(Options{
		Project:             project,
		RouteResult:         routeResult,
		MaterializeRunLocal: true,
		WorkflowManaged:     true,
		RunID:               "run-20260616T105614Z-4b0ebe11b67d",
		Runner:              runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "strict_workflow_missing_kah_capability" {
		t.Fatalf("expected strict capability blocker, got %+v", result)
	}
	assertStrictWorkflowNextAction(t, result.NextAction)
	if result.Materialization != nil {
		t.Fatalf("preflight blocker exposed materialization evidence: %+v", result.Materialization)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["materialization"]; ok {
		t.Fatalf("preflight blocker JSON must omit materialization: %s", encoded)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("materialization wrote before KAH preflight passed: %v", err)
	}
	if len(runner.calls) != 3 {
		t.Fatalf("expected preflight-only calls, got %#v", runner.calls)
	}
}

func TestTriggerWorkflowManagedRequiresStrictKAHCapabilityFlags(t *testing.T) {
	project := t.TempDir()
	registryPath := writeRegistry(t, project, validRegistry("demo", "setup"))
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeTriggerRouteResult(t, project, registry, "demo")
	runner := &fakeKAHRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.10\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true}}`)},
		"workflow --help":     {Stdout: []byte("Subcommands:\n  validate\n  explain\n  create\n  show\n  ready\n  node\n")},
	}}

	result, err := Trigger(Options{
		Project:             project,
		RouteResult:         routeResult,
		MaterializeRunLocal: true,
		WorkflowManaged:     true,
		RunID:               "run-20260616T105614Z-4b0ebe11b67d",
		Runner:              runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "strict_workflow_missing_kah_capability" {
		t.Fatalf("expected strict capability blocker, got %+v", result)
	}
	if result.KAHCapability.CompatibilityFlags["workflow_strict_transition_ledger"] || result.KAHCapability.CompatibilityFlags["workflow_transition_order_verification"] || result.KAHCapability.CompatibilityFlags["workflow_phase_projection_validation"] {
		t.Fatalf("test fixture unexpectedly reports strict flags: %+v", result.KAHCapability.CompatibilityFlags)
	}
	assertStrictWorkflowNextAction(t, result.NextAction)
}

func TestTriggerWorkflowManagedRequiresPhaseProjectionCapabilityFlag(t *testing.T) {
	project := t.TempDir()
	registryPath := writeRegistry(t, project, validRegistry("demo", "setup"))
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeTriggerRouteResult(t, project, registry, "demo")
	runner := &fakeKAHRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.10\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true,"workflow_strict_transition_ledger":true,"workflow_transition_order_verification":true}}`)},
		"workflow --help":     {Stdout: []byte("Subcommands:\n  validate\n  explain\n  create\n  show\n  ready\n  node\n")},
	}}

	result, err := Trigger(Options{
		Project:             project,
		RouteResult:         routeResult,
		MaterializeRunLocal: true,
		WorkflowManaged:     true,
		RunID:               "run-20260616T105614Z-4b0ebe11b67d",
		Runner:              runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "strict_workflow_missing_kah_capability" {
		t.Fatalf("expected phase projection capability blocker, got %+v", result)
	}
	if len(result.DispatchPackets) != 0 {
		t.Fatalf("capability blocker must not render dispatch packets: %+v", result.DispatchPackets)
	}
	if result.KAHCapability.CompatibilityFlags["workflow_phase_projection_validation"] {
		t.Fatalf("test fixture unexpectedly reports phase projection flag: %+v", result.KAHCapability.CompatibilityFlags)
	}
	assertStrictWorkflowNextAction(t, result.NextAction)
}

func TestTriggerWorkflowManagedFailsClosedWhenReadyRevisionMissing(t *testing.T) {
	project := t.TempDir()
	registryPath := writeRegistry(t, project, validRegistry("demo", "setup"))
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeTriggerRouteResult(t, project, registry, "demo")
	workflowFile := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml"
	runner := newWorkflowFakeRunnerForFile("demo", "run-20260616T105614Z-4b0ebe11b67d", workflowFile, []string{"setup"})
	missingRevisionInstance := `{"ok":true,"status":"pass","run_id":"run-20260616T105614Z-4b0ebe11b67d","instance":{"run_id":"run-20260616T105614Z-4b0ebe11b67d","source_path":"` + workflowFile + `"}}`
	runner.responses["workflow create --run run-20260616T105614Z-4b0ebe11b67d --file "+workflowFile+" --json"] = CommandResult{Stdout: []byte(missingRevisionInstance)}
	runner.responses["workflow ready --run run-20260616T105614Z-4b0ebe11b67d --json"] = CommandResult{Stdout: []byte(`{"ok":true,"status":"pass","run_id":"run-20260616T105614Z-4b0ebe11b67d","ready":[{"id":"setup","reasons":["dependencies_satisfied","state_pending"]}]}`)}

	result, err := Trigger(Options{
		Project:             project,
		RouteResult:         routeResult,
		MaterializeRunLocal: true,
		WorkflowManaged:     true,
		RunID:               "run-20260616T105614Z-4b0ebe11b67d",
		Runner:              runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "strict_workflow_expected_start_revision_missing" {
		t.Fatalf("expected missing revision blocker, got %+v", result)
	}
	if len(result.DispatchPackets) != 0 {
		t.Fatalf("stale/missing revision blocker must not render packets: %+v", result.DispatchPackets)
	}
	assertStrictWorkflowNextAction(t, result.NextAction)
}

func TestTriggerWorkflowManagedFailsClosedWhenNoReadyRevisionMissing(t *testing.T) {
	project := t.TempDir()
	registryPath := writeRegistry(t, project, validRegistry("demo", "setup"))
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeTriggerRouteResult(t, project, registry, "demo")
	workflowFile := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml"
	runner := newWorkflowFakeRunnerForFile("demo", "run-20260616T105614Z-4b0ebe11b67d", workflowFile, nil)
	missingRevisionInstance := `{"ok":true,"status":"pass","run_id":"run-20260616T105614Z-4b0ebe11b67d","instance":{"run_id":"run-20260616T105614Z-4b0ebe11b67d","source_path":"` + workflowFile + `"}}`
	runner.responses["workflow create --run run-20260616T105614Z-4b0ebe11b67d --file "+workflowFile+" --json"] = CommandResult{Stdout: []byte(missingRevisionInstance)}
	runner.responses["workflow ready --run run-20260616T105614Z-4b0ebe11b67d --json"] = CommandResult{Stdout: []byte(`{"ok":true,"status":"pass","run_id":"run-20260616T105614Z-4b0ebe11b67d","ready":[]}`)}

	result, err := Trigger(Options{
		Project:             project,
		RouteResult:         routeResult,
		MaterializeRunLocal: true,
		WorkflowManaged:     true,
		RunID:               "run-20260616T105614Z-4b0ebe11b67d",
		Runner:              runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "strict_workflow_expected_start_revision_missing" {
		t.Fatalf("expected missing revision blocker before no-ready success, got %+v", result)
	}
}

func TestTriggerRunLocalMaterializationRendersDispatchPackets(t *testing.T) {
	project := t.TempDir()
	registryPath := writeRegistry(t, project, validRegistry("demo", "setup"))
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeTriggerRouteResult(t, project, registry, "demo")
	workflowFile := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml"
	runner := newWorkflowFakeRunnerForFile("demo", "run-20260616T105614Z-4b0ebe11b67d", workflowFile, []string{"setup"})

	result, err := Trigger(Options{
		Project:             project,
		RouteResult:         routeResult,
		MaterializeRunLocal: true,
		RunID:               "run-20260616T105614Z-4b0ebe11b67d",
		Runner:              runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Mode != "run_local_materialized_trigger" || result.Materialization == nil || result.Materialization.WorkflowFile != workflowFile {
		t.Fatalf("unexpected materialized trigger result: %+v", result)
	}
	if !result.Materialization.NoPromotion || result.Materialization.PersistentPromotion {
		t.Fatalf("materialization promotion posture drifted: %+v", result.Materialization)
	}
	if result.Workflow.Path != workflowFile || result.NodeContractSource != ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/node-contracts.json" {
		t.Fatalf("trigger did not consume run-local files: %+v", result)
	}
	if len(result.DispatchPackets) != 1 {
		t.Fatalf("dispatch packets = %+v", result.DispatchPackets)
	}
	packet := result.DispatchPackets[0]
	if packet.WorkflowFile != workflowFile || packet.CompletionAuthority != "kah_only" || packet.FallbackPolicy != "none_fail_closed" || packet.DirectKAHStateWrite {
		t.Fatalf("packet authority fields drifted: %+v", packet)
	}
	for _, rel := range []string{"workflow.yaml", "node-contracts.json", "route-result.json", "materialization.json", "checksums.json"} {
		if _, err := os.Stat(filepath.Join(project, ".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow", rel)); err != nil {
			t.Fatalf("missing materialized %s: %v", rel, err)
		}
	}
	for _, want := range []string{
		"workflow validate --file " + workflowFile + " --json",
		"workflow explain --file " + workflowFile + " --json",
		"workflow create --run run-20260616T105614Z-4b0ebe11b67d --file " + workflowFile + " --json",
	} {
		if !containsCall(runner.calls, want) {
			t.Fatalf("run-local trigger did not call KAH with explicit workflow file %q: %#v", want, runner.calls)
		}
	}
	for _, call := range runner.calls {
		if strings.Contains(call, ".kkachi/workflows/") {
			t.Fatalf("run-local trigger fell back to persistent workflow path: %#v", runner.calls)
		}
	}
}

func TestTriggerWorkflowManagedFailsClosedWithoutRouteOrMaterializationEvidence(t *testing.T) {
	project := t.TempDir()
	source := writeNodeContractBundle(t, project, validNodeContractBundle("demo", "setup"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})

	result, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		NodeContractSource: source,
		RunID:              "run-20260615T010203Z-abcdef123456",
		WorkflowManaged:    true,
		Runner:             runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "strict_workflow_route_result_required" {
		t.Fatalf("expected strict route-result blocker, got %+v", result)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("strict route-result blocker must not call KAH: %#v", runner.calls)
	}
	if len(result.DispatchPackets) != 0 {
		t.Fatalf("strict blocker must not render dispatch packets: %+v", result.DispatchPackets)
	}
	assertStrictWorkflowNextAction(t, result.NextAction)
}

func TestTriggerWorkflowManagedRejectsSelectorBypassBeforeKAH(t *testing.T) {
	project := t.TempDir()
	registry := writeRegistry(t, project, validRegistry("demo", "setup"))
	runner := newWorkflowFakeRunner("demo", "run-20260615T010203Z-abcdef123456", []string{"setup"})

	result, err := Trigger(Options{
		Project:              project,
		SelectorRegistry:     registry,
		TaskClass:            "development",
		ChangedSurfaces:      []string{"code"},
		RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"},
		RunID:                "run-20260615T010203Z-abcdef123456",
		WorkflowManaged:      true,
		Runner:               runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "strict_workflow_route_result_required" {
		t.Fatalf("expected selector bypass blocker, got %+v", result)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("selector bypass blocker must not call KAH: %#v", runner.calls)
	}
	assertStrictWorkflowNextAction(t, result.NextAction)
}

func TestTriggerWorkflowManagedRouteMaterializationRejectsFallbackInputsBeforeKAH(t *testing.T) {
	project := t.TempDir()
	registryPath := writeRegistry(t, project, validRegistry("demo", "setup"))
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeTriggerRouteResult(t, project, registry, "demo")
	customPacket, customApproval := writeTriggerCustomWorkflowPacket(t, project, "demo")

	tests := []struct {
		name       string
		mutateOpts func(*Options)
		wantStatus string
	}{
		{
			name: "approval conflict",
			mutateOpts: func(opts *Options) {
				opts.Approval = "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000"
			},
			wantStatus: "strict_workflow_materialization_mode_conflict",
		},
		{
			name: "custom packet without approval rejected",
			mutateOpts: func(opts *Options) {
				opts.CustomWorkflowPacket = customPacket
			},
			wantStatus: "strict_workflow_custom_packet_rejected",
		},
		{
			name: "custom packet with approval rejected",
			mutateOpts: func(opts *Options) {
				opts.CustomWorkflowPacket = customPacket
				opts.Approval = customApproval
			},
			wantStatus: "strict_workflow_custom_packet_rejected",
		},
		{
			name: "missing run id",
			mutateOpts: func(opts *Options) {
				opts.RunID = ""
			},
			wantStatus: "strict_workflow_run_id_required",
		},
		{
			name: "explicit workflow conflict",
			mutateOpts: func(opts *Options) {
				opts.WorkflowID = "demo"
			},
			wantStatus: "strict_workflow_materialization_mode_conflict",
		},
		{
			name: "selector conflict",
			mutateOpts: func(opts *Options) {
				opts.SelectorRegistry = registryPath
				opts.TaskClass = "development"
			},
			wantStatus: "strict_workflow_materialization_mode_conflict",
		},
		{
			name: "resume instance conflict",
			mutateOpts: func(opts *Options) {
				opts.InstanceID = "run-20260616T105614Z-4b0ebe11b67d"
			},
			wantStatus: "strict_workflow_materialization_mode_conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := newWorkflowFakeRunner("demo", "run-20260616T105614Z-4b0ebe11b67d", []string{"setup"})
			opts := Options{
				Project:             project,
				RouteResult:         routeResult,
				MaterializeRunLocal: true,
				RunID:               "run-20260616T105614Z-4b0ebe11b67d",
				WorkflowManaged:     true,
				Runner:              runner.Run,
			}
			tt.mutateOpts(&opts)

			result, err := Trigger(opts)
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Status != tt.wantStatus {
				t.Fatalf("expected %s, got %+v", tt.wantStatus, result)
			}
			if result.Status == "run_local_materialization_mode_conflict" {
				t.Fatalf("workflow-managed materialization returned non-strict mode conflict: %+v", result)
			}
			if !strings.HasPrefix(result.Status, "strict_workflow_") {
				t.Fatalf("workflow-managed materialization status must be strict-prefixed: %+v", result)
			}
			if len(runner.calls) != 0 {
				t.Fatalf("strict materialization blocker must not call KAH: %#v", runner.calls)
			}
			assertStrictWorkflowNextAction(t, result.NextAction)
		})
	}
}

func assertStrictWorkflowNextAction(t *testing.T, nextAction string) {
	t.Helper()
	for _, want := range []string{"rerun", "workflow"} {
		if !strings.Contains(nextAction, want) {
			t.Fatalf("strict next_action missing %q: %q", want, nextAction)
		}
	}
	if strings.Contains(nextAction, "explicit workflow and node-contract inputs") {
		t.Fatalf("strict next_action must not steer operators to explicit workflow fallback: %q", nextAction)
	}
}

func TestTriggerWorkflowManagedResumeRequiresRunLocalMaterializationEvidence(t *testing.T) {
	project := t.TempDir()
	registryPath := writeRegistry(t, project, validRegistry("demo", "setup"))
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeTriggerRouteResult(t, project, registry, "demo")
	workflowFile := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml"
	runner := newWorkflowFakeRunnerForFile("demo", "run-20260616T105614Z-4b0ebe11b67d", workflowFile, []string{"setup"})

	materialized, err := Trigger(Options{
		Project:             project,
		RouteResult:         routeResult,
		MaterializeRunLocal: true,
		WorkflowManaged:     true,
		RunID:               "run-20260616T105614Z-4b0ebe11b67d",
		Runner:              runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !materialized.OK {
		t.Fatalf("materialization setup failed: %+v", materialized)
	}

	resumeRunner := newWorkflowFakeRunnerForFile("demo", "run-20260616T105614Z-4b0ebe11b67d", workflowFile, []string{"setup"})
	resumed, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		WorkflowFile:       workflowFile,
		NodeContractSource: ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/node-contracts.json",
		InstanceID:         "run-20260616T105614Z-4b0ebe11b67d",
		WorkflowManaged:    true,
		Runner:             resumeRunner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resumed.OK || resumed.Mode != "workflow_managed_resume_trigger" || resumed.Materialization == nil || resumed.Materialization.RouteResultCopy == "" {
		t.Fatalf("unexpected strict resume result: %+v", resumed)
	}
	if containsCall(resumeRunner.calls, "workflow create --run run-20260616T105614Z-4b0ebe11b67d --file "+workflowFile+" --json") {
		t.Fatalf("strict resume must not create: %#v", resumeRunner.calls)
	}
	if !containsCall(resumeRunner.calls, "workflow show --run run-20260616T105614Z-4b0ebe11b67d --json") {
		t.Fatalf("strict resume did not read existing KAH instance: %#v", resumeRunner.calls)
	}

	mismatched, err := Trigger(Options{
		Project:            project,
		WorkflowID:         "demo",
		WorkflowFile:       workflowFile,
		NodeContractSource: ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/node-contracts.json",
		InstanceID:         "run-20260616T105614Z-missing",
		WorkflowManaged:    true,
		Runner:             resumeRunner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if mismatched.OK || mismatched.Status != "strict_workflow_materialization_evidence_required" {
		t.Fatalf("expected missing materialization blocker, got %+v", mismatched)
	}
}

func TestTriggerCustomMaterializationPreflightsBeforeWrite(t *testing.T) {
	project := t.TempDir()
	packetPath, approval := writeTriggerCustomWorkflowPacket(t, project, "demo")
	runner := &fakeKAHRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.9\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"graph","status":"supported","subcommands":["validate"]}],"compatibility_flags":{"workflow_instance_state":false}}`)},
		"workflow --help":     {Stderr: []byte("unknown help topic\n"), Err: errors.New("exit 2")},
	}}

	result, err := Trigger(Options{
		Project:              project,
		CustomWorkflowPacket: packetPath,
		Approval:             approval,
		MaterializeRunLocal:  true,
		RunID:                "run-20260616T105614Z-4b0ebe11b67d",
		Runner:               runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "blocked_missing_kah_workflow_capability" {
		t.Fatalf("expected capability blocker, got %+v", result)
	}
	if result.Materialization != nil {
		t.Fatalf("preflight blocker exposed custom materialization evidence: %+v", result.Materialization)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["materialization"]; ok {
		t.Fatalf("custom preflight blocker JSON must omit materialization: %s", encoded)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("custom materialization wrote before KAH preflight passed: %v", err)
	}
	if len(runner.calls) != 3 {
		t.Fatalf("expected preflight-only calls, got %#v", runner.calls)
	}
}

func TestTriggerCustomMaterializationRendersDispatchPackets(t *testing.T) {
	project := t.TempDir()
	packetPath, approval := writeTriggerCustomWorkflowPacket(t, project, "demo")
	workflowFile := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml"
	runner := newWorkflowFakeRunnerForFile("demo", "run-20260616T105614Z-4b0ebe11b67d", workflowFile, []string{"plan"})

	result, err := Trigger(Options{
		Project:              project,
		CustomWorkflowPacket: packetPath,
		Approval:             approval,
		MaterializeRunLocal:  true,
		RunID:                "run-20260616T105614Z-4b0ebe11b67d",
		Runner:               runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Mode != "run_local_materialized_trigger" || result.Materialization == nil || result.Materialization.CustomWorkflowPacketCopy == "" || result.Materialization.ApprovalEvidence != approval {
		t.Fatalf("unexpected custom materialized trigger result: %+v", result)
	}
	if !result.Materialization.NoPromotion || result.Materialization.PersistentPromotion {
		t.Fatalf("custom materialization promotion posture drifted: %+v", result.Materialization)
	}
	packet := result.DispatchPackets[0]
	if packet.WorkflowFile != workflowFile || packet.NodeID != "plan" || packet.CompletionAuthority != "kah_only" || packet.FallbackPolicy != "none_fail_closed" || packet.DirectKAHStateWrite {
		t.Fatalf("custom packet authority fields drifted: %+v", packet)
	}
	for _, rel := range []string{"workflow.yaml", "node-contracts.json", "custom-workflow-packet.json", "materialization.json", "checksums.json"} {
		if _, err := os.Stat(filepath.Join(project, ".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow", rel)); err != nil {
			t.Fatalf("missing custom materialized %s: %v", rel, err)
		}
	}
	for _, want := range []string{
		"workflow validate --file " + workflowFile + " --json",
		"workflow explain --file " + workflowFile + " --json",
		"workflow create --run run-20260616T105614Z-4b0ebe11b67d --file " + workflowFile + " --json",
	} {
		if !containsCall(runner.calls, want) {
			t.Fatalf("custom run-local trigger did not call KAH with explicit workflow file %q: %#v", want, runner.calls)
		}
	}
	for _, call := range runner.calls {
		if strings.Contains(call, ".kkachi/workflows/") {
			t.Fatalf("custom run-local trigger fell back to persistent workflow path: %#v", runner.calls)
		}
	}
}

func TestTriggerRunLocalMaterializationRejectsRootRequiredOutputsFromKAH(t *testing.T) {
	project := t.TempDir()
	packetPath, approval := writeTriggerCustomWorkflowPacket(t, project, "demo")
	workflowFile := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml"
	runner := newWorkflowFakeRunnerForFileWithRequiredOutputs("demo", "run-20260616T105614Z-4b0ebe11b67d", workflowFile, "plan", []string{"artifacts/plan.md"})

	result, err := Trigger(Options{
		Project:              project,
		CustomWorkflowPacket: packetPath,
		Approval:             approval,
		MaterializeRunLocal:  true,
		RunID:                "run-20260616T105614Z-4b0ebe11b67d",
		Runner:               runner.Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "strict_workflow_run_local_outputs_invalid" || len(result.DispatchPackets) != 0 {
		t.Fatalf("expected run-local KAH output path blocker, got %+v", result)
	}
}

func writeNodeContractBundle(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "node-contracts.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTriggerCustomWorkflowPacket(t *testing.T, dir string, workflowID string) (string, string) {
	t.Helper()
	runID := "run-20260616T105614Z-4b0ebe11b67d"
	artifactPath := ".kkachi/runs/" + runID + "/artifacts/plan.md"
	dag := "workflow_id: " + workflowID + "\nschema_version: task-dag/v1\nnodes:\n  - id: plan\n    depends_on: []\n    join: all_of\n    required_outputs:\n      - " + artifactPath + "\n"
	result := workflowcreator.Result{
		OK:                  true,
		Command:             workflowcreator.Command,
		Mode:                workflowcreator.ModeDAGOnly,
		Status:              "dry_run_ready",
		Project:             workflowcreator.ProjectEvidence{Path: dir},
		Workflow:            workflowcreator.WorkflowEvidence{ID: workflowID, SchemaVersion: "task-dag/v1"},
		DirectKAHStateWrite: false,
		MachinePacket: workflowcreator.MachinePacket{
			SchemaVersion:    workflowcreator.PacketSchemaVersion,
			ApprovalSchema:   workflowcreator.ApprovalSchema,
			Canonicalization: workflowcreator.Canonicalization,
			TargetPaths:      []string{".kkachi/workflows/" + workflowID + ".yaml"},
			CandidatePaths:   workflowcreator.CandidatePaths{WorkflowDAG: ".kkachi/workflows/" + workflowID + ".yaml"},
			GeneratedContent: []workflowcreator.GeneratedContent{{Path: ".kkachi/workflows/" + workflowID + ".yaml", Kind: "workflow_dag", Content: dag, SHA256: triggerChecksum([]byte(dag))}},
			SelectorMetadata: map[string]any{"task_class": "development"},
			NodeContracts: []workflowcreator.NodeContract{{
				WorkflowID:          workflowID,
				NodeID:              "plan",
				TaskClass:           "development",
				OwnerRole:           "planner_backend",
				ExecutionLane:       "stage1_direct_codex_app_server",
				RequiredInputs:      []string{".kkachi/runs/" + runID + "/task-contract.yaml"},
				ExpectedArtifacts:   []string{artifactPath},
				PromptRef:           "skills/kkachi-plan/SKILL.md",
				ApprovalRequired:    false,
				FallbackPolicy:      workflowregistry.NoFallbackPolicy,
				VerificationGate:    "kah_workflow_node_evidence",
				CompletionAuthority: workflowregistry.KAHOnlyAuthority,
				DirectKAHStateWrite: false,
			}},
			TriggerPlan:   workflowcreator.TriggerPlan{Mode: workflowcreator.ModeDAGOnly, Generated: false, DelegatesTo: "kkachi-agent-skills workflow-trigger", CustomLogic: false},
			BaseChecksums: map[string]string{".kkachi/workflows/" + workflowID + ".yaml": "missing"},
			ChangedPaths:  []workflowcreator.ChangedPath{{Path: ".kkachi/workflows/" + workflowID + ".yaml", Action: "create", Kind: "workflow_dag"}},
			Conflicts:     []workflowcreator.Conflict{},
			Diagnostics:   []workflowcreator.Diagnostic{},
			NoWrite:       workflowcreator.NoWriteEvidence{Guaranteed: true},
		},
	}
	result.MachinePacket.ApprovalHash = workflowcreator.RecomputeApprovalHash(result)
	result.ApprovalRequest = workflowcreator.ApprovalRequest{Required: true, EvidenceRef: "dry-run:" + result.MachinePacket.ApprovalHash, DryRunPlanHash: result.MachinePacket.ApprovalHash}
	path := filepath.Join(dir, "workflow-create-dry-run.json")
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, result.ApprovalRequest.EvidenceRef
}

func triggerChecksum(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func writeTriggerRouteResult(t *testing.T, dir string, registry workflowregistry.Registry, workflowID string) string {
	t.Helper()
	path := filepath.Join(dir, "route-result.json")
	payload := map[string]any{
		"ok":                    true,
		"command":               "workflow-route",
		"status":                "bundle_route_matched",
		"task_class":            "development",
		"classification_reason": "test route",
		"selected_bundle":       workflowID,
		"workflow_id":           workflowID,
		"workflow_path":         ".kkachi/workflows/" + workflowID + ".yaml",
		"selected_spine":        workflowID,
		"taxonomy": map[string]any{
			"path":     "registries/task-taxonomy.yaml",
			"checksum": "sha256:taxonomy",
		},
		"selector_registry": map[string]any{
			"path":     registry.Path,
			"version":  registry.Version,
			"checksum": registry.Checksum,
		},
		"direct_kah_state_write": false,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func validNodeContractBundle(workflowID string, nodeID string) string {
	return `{
  "schema_version": "kas-node-contracts/v1",
  "contracts": [
    {
      "workflow_id": "` + workflowID + `",
      "node_id": "` + nodeID + `",
      "owner_role": "implementer_backend",
      "execution_lane": "direct_kas_skill",
      "required_inputs": ["task-contract.yaml"],
      "expected_artifacts": ["artifacts/setup.md"],
      "prompt_ref": "skills/kkachi-implement/SKILL.md",
      "approval_required": false,
      "fallback_policy": "none_fail_closed",
      "verification_gate": "make test"
    }
  ]
}`
}

func writeRegistry(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "workflow-registry.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func validRegistry(workflowID string, nodeID string) string {
	return `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: ` + workflowID + `
    workflow_path: .kkachi/workflows/` + workflowID + `.yaml
    selector:
      task_classes: [development]
      labels_any: []
      labels_all: []
      changed_surfaces_any: [code, tests]
      risk_levels: []
      required_agents_all: []
      required_capabilities_all: [task_dag_schema_validation, workflow_instance_state]
    fallback_policy: none_fail_closed
node_contracts:
  - workflow_id: ` + workflowID + `
    node_id: ` + nodeID + `
    task_class: development
    owner_role: implementer_backend
    execution_lane: direct_kas_skill
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/` + nodeID + `.md]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: make test
    completion_authority: kah_only
    direct_kah_state_write: false
`
}

func newWorkflowFakeRunner(workflowID string, runID string, readyNodes []string) *fakeKAHRunner {
	return newWorkflowFakeRunnerForFile(workflowID, runID, ".kkachi/workflows/"+workflowID+".yaml", readyNodes)
}

func newWorkflowFakeRunnerForFile(workflowID string, runID string, workflowFile string, readyNodes []string) *fakeKAHRunner {
	ready := make([]map[string]any, 0, len(readyNodes))
	nodes := make([]map[string]any, 0, len(readyNodes))
	for _, id := range readyNodes {
		requiredOutputs := []string{"artifacts/" + id + ".md"}
		if strings.HasPrefix(workflowFile, ".kkachi/runs/"+runID+"/workflow/") {
			requiredOutputs = []string{".kkachi/runs/" + runID + "/artifacts/" + id + ".md"}
		}
		ready = append(ready, map[string]any{"id": id, "reasons": []string{"dependencies_satisfied", "state_pending"}})
		nodes = append(nodes, map[string]any{"id": id, "state": "pending", "depends_on": []string{}, "join": "all_of", "required_outputs": requiredOutputs})
	}
	instance := map[string]any{
		"version":        "workflow-instance/v1",
		"run_id":         runID,
		"workflow_id":    workflowID,
		"schema_version": "task-dag/v1",
		"source_path":    workflowFile,
		"revision":       1,
		"nodes":          nodes,
	}
	instanceResult := map[string]any{"ok": true, "status": "pass", "reason": "workflow_instance_loaded", "run_id": runID, "instance": instance, "ready": ready}
	instanceJSON, _ := json.Marshal(instanceResult)
	readyResult := map[string]any{"ok": true, "status": "pass", "reason": "workflow_ready_nodes_computed", "run_id": runID, "instance": instance, "ready": ready}
	readyJSON, _ := json.Marshal(readyResult)

	return &fakeKAHRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.10\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true,"workflow_strict_transition_ledger":true,"workflow_transition_order_verification":true,"workflow_phase_projection_validation":true}}`)},
		"workflow --help":     {Stdout: []byte("Subcommands:\n  validate\n  explain\n  create\n  show\n  ready\n  node\n")},
		"workflow validate --file " + workflowFile + " --json": {Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"` + workflowID + `","schema_version":"task-dag/v1"}`)},
		"workflow explain --file " + workflowFile + " --json":  {Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"` + workflowID + `","schema_version":"task-dag/v1"}`)},
		"workflow create --run " + runID + " --file " + workflowFile + " --json": {
			Stdout: instanceJSON,
		},
		"workflow show --run " + runID + " --json": {
			Stdout: instanceJSON,
		},
		"workflow ready --run " + runID + " --json": {
			Stdout: readyJSON,
		},
	}}
}

func newWorkflowFakeRunnerForFileWithRequiredOutputs(workflowID string, runID string, workflowFile string, nodeID string, requiredOutputs []string) *fakeKAHRunner {
	ready := []map[string]any{{"id": nodeID, "reasons": []string{"dependencies_satisfied", "state_pending"}, "required_outputs": requiredOutputs}}
	nodes := []map[string]any{{"id": nodeID, "state": "pending", "depends_on": []string{}, "join": "all_of", "required_outputs": requiredOutputs}}
	instance := map[string]any{
		"version":        "workflow-instance/v1",
		"run_id":         runID,
		"workflow_id":    workflowID,
		"schema_version": "task-dag/v1",
		"source_path":    workflowFile,
		"revision":       1,
		"nodes":          nodes,
	}
	instanceResult := map[string]any{"ok": true, "status": "pass", "reason": "workflow_instance_loaded", "run_id": runID, "instance": instance, "ready": ready}
	instanceJSON, _ := json.Marshal(instanceResult)
	readyResult := map[string]any{"ok": true, "status": "pass", "reason": "workflow_ready_nodes_computed", "run_id": runID, "instance": instance, "ready": ready}
	readyJSON, _ := json.Marshal(readyResult)
	return &fakeKAHRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.10\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true,"workflow_strict_transition_ledger":true,"workflow_transition_order_verification":true,"workflow_phase_projection_validation":true}}`)},
		"workflow --help":     {Stdout: []byte("Subcommands:\n  validate\n  explain\n  create\n  show\n  ready\n  node\n")},
		"workflow validate --file " + workflowFile + " --json": {Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"` + workflowID + `","schema_version":"task-dag/v1"}`)},
		"workflow explain --file " + workflowFile + " --json":  {Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"` + workflowID + `","schema_version":"task-dag/v1"}`)},
		"workflow create --run " + runID + " --file " + workflowFile + " --json": {
			Stdout: instanceJSON,
		},
		"workflow show --run " + runID + " --json": {
			Stdout: instanceJSON,
		},
		"workflow ready --run " + runID + " --json": {
			Stdout: readyJSON,
		},
	}}
}

func containsCall(calls []string, want string) bool {
	for _, call := range calls {
		if call == want {
			return true
		}
	}
	return false
}
