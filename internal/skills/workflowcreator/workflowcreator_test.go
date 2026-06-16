package workflowcreator

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	responses map[string]CommandResult
	calls     []string
}

func (r *fakeRunner) Run(workDir string, args ...string) CommandResult {
	key := strings.Join(args, " ")
	r.calls = append(r.calls, key)
	if result, ok := r.responses[key]; ok {
		return result
	}
	return CommandResult{Stderr: []byte("unexpected call: " + key), Err: errors.New("unexpected call")}
}

func TestDryRunBuildsHashBoundPacketsForAllModes(t *testing.T) {
	for _, mode := range []string{ModeDAGOnly, ModeThinTrigger, ModeFullTrigger} {
		t.Run(mode, func(t *testing.T) {
			project := t.TempDir()
			request := writeRequest(t, project, validRequest(mode))
			result, err := BuildDryRun(Options{
				Project: project, WorkflowID: "release-flow", Mode: mode, RequestPath: request, FullTriggerReason: "scenario-specific operator input", Runner: workflowCapableRunner().Run,
			})
			if err != nil {
				t.Fatal(err)
			}
			if !result.OK || result.Status != "dry_run_ready" || result.DirectKAHStateWrite {
				t.Fatalf("unexpected dry-run result: %+v", result)
			}
			if result.ApprovalRequest.EvidenceRef != "dry-run:"+result.MachinePacket.ApprovalHash || !strings.HasPrefix(result.MachinePacket.ApprovalHash, "sha256:") {
				t.Fatalf("approval evidence not bound to packet hash: %+v", result.ApprovalRequest)
			}
			if !result.ApprovalRequest.HashIncludesTargetPaths || !result.ApprovalRequest.HashIncludesGeneratedContent || !result.ApprovalRequest.HashIncludesCapabilityEvidence || !result.ApprovalRequest.HashIncludesNoWriteEvidence {
				t.Fatalf("approval request missing hash-scope fields: %+v", result.ApprovalRequest)
			}
			if result.MachinePacket.NoWrite.ProjectWriteCount != 0 || result.MachinePacket.NoWrite.KAHStateWriteCount != 0 || result.MachinePacket.NoWrite.ProfileWriteCount != 0 {
				t.Fatalf("dry-run no-write evidence is not zeroed: %+v", result.MachinePacket.NoWrite)
			}
			if result.MachinePacket.CandidatePaths.WorkflowDAG != ".kkachi/workflows/release-flow.yaml" ||
				result.MachinePacket.CandidatePaths.Catalog != ".kkachi/workflow-catalog.yaml" ||
				result.MachinePacket.CandidatePaths.NodeContractRegistry != ".kkachi/workflows/release-flow-node-contracts.yaml" {
				t.Fatalf("unexpected candidate paths: %+v", result.MachinePacket.CandidatePaths)
			}
			wantGenerated := 3
			if mode != ModeDAGOnly {
				wantGenerated = 4
				if result.MachinePacket.TriggerPlan.Path == "" || result.MachinePacket.TriggerPlan.DelegatesTo != "kkachi-agent-skills workflow-trigger" {
					t.Fatalf("trigger mode missing bounded trigger plan: %+v", result.MachinePacket.TriggerPlan)
				}
			}
			if len(result.MachinePacket.GeneratedContent) != wantGenerated {
				t.Fatalf("generated content count = %d, want %d", len(result.MachinePacket.GeneratedContent), wantGenerated)
			}
			for _, item := range result.MachinePacket.GeneratedContent {
				if strings.HasPrefix(item.Path, "/") || strings.Contains(item.Path, "\\") || item.SHA256 == "" {
					t.Fatalf("generated content path/checksum not normalized: %+v", item)
				}
			}
		})
	}
}

func TestDryRunApprovalHashIsStableAndBindsSelectorMetadata(t *testing.T) {
	project := t.TempDir()
	request := writeRequest(t, project, validRequest(ModeDAGOnly))
	runner := workflowCapableRunner()
	first, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: request, Runner: runner.Run})
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: request, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if first.MachinePacket.ApprovalHash != second.MachinePacket.ApprovalHash {
		t.Fatalf("hash not stable: %s != %s", first.MachinePacket.ApprovalHash, second.MachinePacket.ApprovalHash)
	}
	changed := writeRequest(t, project, strings.Replace(validRequest(ModeDAGOnly), `"release"`, `"hotfix"`, 1))
	third, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: changed, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if first.MachinePacket.ApprovalHash == third.MachinePacket.ApprovalHash {
		t.Fatalf("hash did not bind selector metadata: %s", first.MachinePacket.ApprovalHash)
	}
}

func TestDryRunFailsClosedForMissingNilOrEmptySelectorMetadata(t *testing.T) {
	project := t.TempDir()
	for name, content := range map[string]string{
		"missing": strings.Replace(validRequest(ModeDAGOnly), validSelectorMetadataBlock(), "", 1),
		"nil":     strings.Replace(validRequest(ModeDAGOnly), validSelectorMetadataBlock(), `  "selector_metadata": null,`, 1),
		"empty":   strings.Replace(validRequest(ModeDAGOnly), validSelectorMetadataBlock(), `  "selector_metadata": {},`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			request := writeRequest(t, project, content)
			result, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: request, Runner: workflowCapableRunner().Run})
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Status != "selector_metadata_invalid" || result.ApprovalRequest.Required {
				t.Fatalf("expected selector metadata blocker, got %+v", result)
			}
			if len(result.MachinePacket.GeneratedContent) != 0 || len(result.MachinePacket.ChangedPaths) != 0 {
				t.Fatalf("selector metadata blocker generated candidate content: %+v", result.MachinePacket)
			}
		})
	}
}

func TestDryRunFailsClosedForUnsafePathUnsupportedModeAndMissingCapability(t *testing.T) {
	project := t.TempDir()
	unsafe := writeRequest(t, project, strings.Replace(validRequest(ModeDAGOnly), "artifacts/plan.md", "../outside.md", 1))
	result, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: unsafe, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || firstCode(result.Diagnostics) != "unsafe_target_path" {
		t.Fatalf("expected unsafe path blocker, got %+v", result)
	}
	valid := writeRequest(t, project, validRequest(ModeDAGOnly))
	result, err = BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: "webhook", RequestPath: valid, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || firstCode(result.Diagnostics) != "workflow_creator_mode_unsupported" {
		t.Fatalf("expected unsupported mode blocker, got %+v", result)
	}
	result, err = BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: valid, Runner: missingWorkflowRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || firstCode(result.Diagnostics) != "blocked_missing_kah_workflow_capability" || result.ApprovalRequest.Required {
		t.Fatalf("expected missing KAH capability non-approvable blocker, got %+v", result)
	}
}

func TestDryRunFailsClosedForInvalidGeneratedTriggerSkill(t *testing.T) {
	project := t.TempDir()
	request := writeRequest(t, project, strings.Replace(validRequest(ModeThinTrigger), `"release-flow-trigger"`, `"bad trigger name"`, 1))
	result, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeThinTrigger, RequestPath: request, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "generated_skill_validation_failed" || result.ApprovalRequest.Required {
		t.Fatalf("expected generated trigger validation blocker, got %+v", result)
	}
	if result.MachinePacket.TriggerPlan.ValidationStatus != "generated_skill_validation_failed" {
		t.Fatalf("expected trigger validation status to be bound in packet: %+v", result.MachinePacket.TriggerPlan)
	}
}

func TestApplyRecomputesHashBeforeCapabilityApply(t *testing.T) {
	project := t.TempDir()
	request := writeRequest(t, project, validRequest(ModeDAGOnly))
	wrong, err := Apply(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: request, Approval: "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if wrong.OK || wrong.Status != "approval_plan_hash_mismatch" || wrong.Approval.MatchedCurrentPlan {
		t.Fatalf("expected hash mismatch before apply behavior: %+v", wrong)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("wrong hash created project state: %v", err)
	}
	dryRun, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: request, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	blocked, err := Apply(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: request, Approval: dryRun.ApprovalRequest.EvidenceRef, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if blocked.OK || blocked.Status != "blocked_missing_kah_workflow_catalog_capability" || !blocked.Approval.MatchedCurrentPlan {
		t.Fatalf("expected advertised-hash apply to fail closed on absent KAH catalog apply surface: %+v", blocked)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("blocked apply created project state: %v", err)
	}
}

func TestApplyRefusesAdvertisedButUnmappedKAHApplySurface(t *testing.T) {
	project := t.TempDir()
	request := writeRequest(t, project, validRequest(ModeDAGOnly))
	dryRun, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: request, Runner: workflowApplySurfaceRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK || dryRun.KAHCapability.ApplySurface == "" {
		t.Fatalf("test runner did not expose workflow apply surface: %+v", dryRun.KAHCapability)
	}
	blocked, err := Apply(Options{Project: project, WorkflowID: "demo", Mode: ModeDAGOnly, RequestPath: request, Approval: dryRun.ApprovalRequest.EvidenceRef, Runner: workflowApplySurfaceRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if blocked.OK || blocked.Status != "workflow_create_apply_refused" || !blocked.Approval.MatchedCurrentPlan {
		t.Fatalf("expected unmapped apply surface refusal, got %+v", blocked)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("unmapped apply surface created project state: %v", err)
	}
}

func writeRequest(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "request.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func validRequest(mode string) string {
	fullReason := ""
	if mode == ModeFullTrigger {
		fullReason = `,
  "full_trigger_reason": "scenario-specific operator input"`
	}
	return `{
  "schema_version": "kas-workflow-create-request/v1",
  "selector_metadata": {
    "task_class": "development",
    "labels": ["release"],
    "changed_surfaces": ["code"],
    "required_capabilities": ["task_dag_schema_validation", "workflow_instance_state"]
  },
  "nodes": [
    {
      "node_id": "plan",
      "task_class": "development",
      "depends_on": [],
      "required_outputs": ["artifacts/plan.md"],
      "owner_role": "planner_backend",
      "execution_lane": "stage1_direct_codex_app_server",
      "required_inputs": ["task-contract.yaml"],
      "expected_artifacts": ["plan.md"],
      "prompt_ref": "skills/kkachi-plan/SKILL.md",
      "approval_required": false,
      "fallback_policy": "none_fail_closed",
      "verification_gate": "kah_workflow_node_evidence"
    }
  ],
  "trigger": {
    "name": "release-flow-trigger",
    "description": "Release flow trigger"
  }` + fullReason + `
}`
}

func validSelectorMetadataBlock() string {
	return `  "selector_metadata": {
    "task_class": "development",
    "labels": ["release"],
    "changed_surfaces": ["code"],
    "required_capabilities": ["task_dag_schema_validation", "workflow_instance_state"]
  },`
}

func workflowCapableRunner() *fakeRunner {
	return &fakeRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.10-source\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","catalog","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true,"workflow_catalog_diagnostics":true,"workflow_final_gate_integration":true,"workflow_node_contract_registry_evidence":true}}`)},
		"workflow --help":     {Stdout: []byte("workflow validate\nworkflow explain\nworkflow catalog validate\nworkflow catalog explain\nworkflow create\nworkflow show\nworkflow ready\nworkflow node\n")},
	}}
}

func workflowApplySurfaceRunner() *fakeRunner {
	return &fakeRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.10-source\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","catalog","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true,"workflow_catalog_diagnostics":true,"workflow_final_gate_integration":true,"workflow_node_contract_registry_evidence":true,"workflow_catalog_apply":true}}`)},
		"workflow --help":     {Stdout: []byte("workflow validate\nworkflow explain\nworkflow catalog validate\nworkflow catalog explain\nworkflow catalog apply\nworkflow create\nworkflow show\nworkflow ready\nworkflow node\n")},
	}}
}

func missingWorkflowRunner() *fakeRunner {
	return &fakeRunner{responses: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.9\n")},
		"capabilities --json": {Stdout: []byte(`{"command_groups":[{"name":"graph","status":"supported","subcommands":["validate"]}],"compatibility_flags":{"workflow_instance_state":false}}`)},
		"workflow --help":     {Stderr: []byte("unknown help topic"), Err: errors.New("exit 2")},
	}}
}

func firstCode(diags []Diagnostic) string {
	if len(diags) == 0 {
		return ""
	}
	return diags[0].Code
}

func TestResultJSONCarriesMachinePacket(t *testing.T) {
	project := t.TempDir()
	request := writeRequest(t, project, validRequest(ModeThinTrigger))
	result, err := BuildDryRun(Options{Project: project, WorkflowID: "demo", Mode: ModeThinTrigger, RequestPath: request, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"machine_packet", "candidate_paths", "generated_content", "selector_metadata", "base_checksums", "no_write", "approval_hash"} {
		if !strings.Contains(text, want) {
			t.Fatalf("result JSON missing %s: %s", want, text)
		}
	}
}
