package workflowtrigger

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
	if packet.WorkflowID != "demo" || packet.InstanceID != "run-20260615T010203Z-abcdef123456" || packet.NodeID != "setup" {
		t.Fatalf("unexpected packet identity: %+v", packet)
	}
	if packet.OwnerRole != "implementer_backend" || packet.ExecutionLane != "direct_kas_skill" || !packet.ApprovalRequired {
		t.Fatalf("unexpected packet contract fields: %+v", packet)
	}
	if packet.FallbackPolicy != "none_fail_closed" || packet.NodeContractRef != "contracts-a" || packet.SourceChecksum == "" {
		t.Fatalf("packet missing source/fallback evidence: %+v", packet)
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

func writeNodeContractBundle(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "node-contracts.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
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

func newWorkflowFakeRunner(workflowID string, runID string, readyNodes []string) *fakeKAHRunner {
	ready := make([]map[string]any, 0, len(readyNodes))
	nodes := make([]map[string]any, 0, len(readyNodes))
	for _, id := range readyNodes {
		ready = append(ready, map[string]any{"id": id, "reasons": []string{"dependencies_satisfied", "state_pending"}})
		nodes = append(nodes, map[string]any{"id": id, "state": "pending", "depends_on": []string{}, "join": "all_of", "required_outputs": []string{"artifacts/" + id + ".md"}})
	}
	instance := map[string]any{
		"version":        "workflow-instance/v1",
		"run_id":         runID,
		"workflow_id":    workflowID,
		"schema_version": "task-dag/v1",
		"source_path":    ".kkachi/workflows/" + workflowID + ".yaml",
		"revision":       1,
		"nodes":          nodes,
	}
	instanceResult := map[string]any{"ok": true, "status": "pass", "reason": "workflow_instance_loaded", "run_id": runID, "instance": instance, "ready": ready}
	instanceJSON, _ := json.Marshal(instanceResult)
	readyResult := map[string]any{"ok": true, "status": "pass", "reason": "workflow_ready_nodes_computed", "run_id": runID, "instance": instance, "ready": ready}
	readyJSON, _ := json.Marshal(readyResult)

	return &fakeKAHRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.10\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true}}`)},
		"workflow --help":     {Stdout: []byte("Subcommands:\n  validate\n  explain\n  create\n  show\n  ready\n  node\n")},
		"workflow validate --file .kkachi/workflows/" + workflowID + ".yaml --json": {Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"` + workflowID + `","schema_version":"task-dag/v1"}`)},
		"workflow explain --file .kkachi/workflows/" + workflowID + ".yaml --json":  {Stdout: []byte(`{"ok":true,"status":"valid","workflow_id":"` + workflowID + `","schema_version":"task-dag/v1"}`)},
		"workflow create --run " + runID + " --file .kkachi/workflows/" + workflowID + ".yaml --json": {
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
