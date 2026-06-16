package workflowpromoter

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowregistry"
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

func TestDryRunBuildsHashBoundPromotionPacketWithoutWrites(t *testing.T) {
	project := t.TempDir()
	writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")

	result, err := BuildDryRun(Options{
		Project:          project,
		RunID:            "run-20260616T105614Z-4b0ebe11b67d",
		TargetWorkflowID: "promoted-flow",
		ReuseReason:      "release workflow proved reusable across two production hotfixes",
		ThinTrigger:      true,
		Runner:           workflowCapableRunner().Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Status != "dry_run_ready" || result.DirectKAHStateWrite {
		t.Fatalf("unexpected dry-run result: %+v", result)
	}
	if result.Command != Command || result.Workflow.SourceID != "source-flow" || result.Workflow.TargetID != "promoted-flow" {
		t.Fatalf("unexpected workflow evidence: %+v", result.Workflow)
	}
	if result.MachinePacket.SourceProvenance.RunID != "run-20260616T105614Z-4b0ebe11b67d" ||
		result.MachinePacket.SourceProvenance.WorkflowChecksum == "" ||
		result.MachinePacket.SourceProvenance.NodeContractsChecksum == "" ||
		result.MachinePacket.SourceProvenance.MaterializationChecksum == "" {
		t.Fatalf("missing source provenance: %+v", result.MachinePacket.SourceProvenance)
	}
	if result.ApprovalRequest.EvidenceRef != "dry-run:"+result.MachinePacket.ApprovalHash ||
		!result.ApprovalRequest.HashIncludesSourceRunMaterialization ||
		!result.ApprovalRequest.HashIncludesSourceChecksums ||
		!result.ApprovalRequest.HashIncludesGeneratedContent ||
		!result.ApprovalRequest.HashIncludesNoWriteEvidence {
		t.Fatalf("approval hash scope incomplete: %+v", result.ApprovalRequest)
	}
	if result.MachinePacket.CandidatePaths.WorkflowDAG != ".kkachi/workflows/promoted-flow.yaml" ||
		result.MachinePacket.CandidatePaths.Catalog != ".kkachi/workflow-catalog.yaml" ||
		result.MachinePacket.CandidatePaths.NodeContractRegistry != ".kkachi/workflows/promoted-flow-node-contracts.yaml" ||
		result.MachinePacket.CandidatePaths.TriggerSkill != ".kkachi/workflow-triggers/promoted-flow-trigger/SKILL.md" {
		t.Fatalf("unexpected target paths: %+v", result.MachinePacket.CandidatePaths)
	}
	if len(result.MachinePacket.GeneratedContent) != 4 {
		t.Fatalf("generated content count = %d, want 4", len(result.MachinePacket.GeneratedContent))
	}
	generated := generatedContentByKind(t, result, "workflow_dag")
	if !strings.Contains(generated, "workflow_id: promoted-flow") || strings.Contains(generated, "workflow_id: source-flow") {
		t.Fatalf("workflow was not retargeted:\n%s", generated)
	}
	registry := generatedContentByKind(t, result, "node_contract_registry")
	for _, want := range []string{
		"workflow_id: promoted-flow",
		"completion_authority: kah_only",
		"direct_kah_state_write: false",
		"fallback_policy: none_fail_closed",
	} {
		if !strings.Contains(registry, want) {
			t.Fatalf("generated registry missing %q:\n%s", want, registry)
		}
	}
	if result.MachinePacket.NoWrite.ProjectWriteCount != 0 ||
		result.MachinePacket.NoWrite.KAHStateWriteCount != 0 ||
		result.MachinePacket.NoWrite.KABRuntimeMutationCount != 0 ||
		result.MachinePacket.NoWrite.AuthProviderConfigWriteCount != 0 {
		t.Fatalf("dry-run no-write evidence drifted: %+v", result.MachinePacket.NoWrite)
	}
	for _, forbidden := range []string{".kkachi/workflows", ".kkachi/workflow-catalog.yaml", ".kkachi-workflow.yaml"} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(forbidden))); !os.IsNotExist(err) {
			t.Fatalf("dry-run wrote forbidden project-local path %s: %v", forbidden, err)
		}
	}
}

func TestDryRunHashBindsSourceAndTargetInputs(t *testing.T) {
	project := t.TempDir()
	writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")
	first, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if first.MachinePacket.ApprovalHash != second.MachinePacket.ApprovalHash {
		t.Fatalf("approval hash is not stable: %s != %s", first.MachinePacket.ApprovalHash, second.MachinePacket.ApprovalHash)
	}
	changedReason, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "different operator justification", Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if first.MachinePacket.ApprovalHash == changedReason.MachinePacket.ApprovalHash {
		t.Fatalf("approval hash did not bind reuse reason: %s", first.MachinePacket.ApprovalHash)
	}
	workflowPath := filepath.Join(project, ".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow", "workflow.yaml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workflowPath, []byte(strings.Replace(string(data), "artifacts/setup.md", "artifacts/changed.md", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	tampered, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if tampered.OK || tampered.Status != "source_checksum_mismatch" || tampered.ApprovalRequest.Required {
		t.Fatalf("expected source checksum mismatch blocker, got %+v", tampered)
	}
}

func TestDryRunRejectsMissingRecordedSourceChecksums(t *testing.T) {
	for name, missingPath := range map[string]string{
		"workflow checksum missing":       filepath.ToSlash(filepath.Join(".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow", "workflow.yaml")),
		"node contracts checksum missing": filepath.ToSlash(filepath.Join(".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow", "node-contracts.json")),
	} {
		t.Run(name, func(t *testing.T) {
			project := t.TempDir()
			writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")
			mutateRunLocalChecksums(t, project, "run-20260616T105614Z-4b0ebe11b67d", func(checksums map[string]string) {
				delete(checksums, missingPath)
			})
			result, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run})
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Status != "source_checksum_missing" || result.ApprovalRequest.Required {
				t.Fatalf("expected missing source checksum to fail closed, got %+v", result)
			}
			assertNoProjectLocalWrites(t, project)
		})
	}
}

func TestApprovalHashBindsPromotionPacketFields(t *testing.T) {
	project := t.TempDir()
	writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")
	result, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("test setup produced blocked dry-run: %+v", result)
	}
	base := result.MachinePacket.ApprovalHash
	for name, mutate := range map[string]func(*Result){
		"source provenance": func(r *Result) {
			r.MachinePacket.SourceProvenance.MaterializationChecksum = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
		"target paths": func(r *Result) {
			r.MachinePacket.TargetPaths = append(r.MachinePacket.TargetPaths, ".kkachi/workflows/extra.yaml")
		},
		"generated content": func(r *Result) {
			r.MachinePacket.GeneratedContent[0].Content += "# changed\n"
			r.MachinePacket.GeneratedContent[0].SHA256 = checksumBytes([]byte(r.MachinePacket.GeneratedContent[0].Content))
		},
		"capability evidence": func(r *Result) {
			r.MachinePacket.CapabilityEvidence.KAH.Version = "kkachi-agent-helper changed"
		},
		"diagnostics": func(r *Result) {
			r.MachinePacket.Diagnostics = append(r.MachinePacket.Diagnostics, Diagnostic{Level: "warning", Code: "changed_diagnostic", Message: "diagnostic changed"})
		},
		"conflicts": func(r *Result) {
			r.MachinePacket.Conflicts = append(r.MachinePacket.Conflicts, Conflict{Code: "changed_conflict", Message: "conflict changed"})
		},
		"no-write evidence": func(r *Result) {
			r.MachinePacket.NoWrite.KAHStateWriteCount = 1
		},
	} {
		t.Run(name, func(t *testing.T) {
			changed := result
			mutate(&changed)
			if hash := RecomputeApprovalHash(changed); hash == base {
				t.Fatalf("approval hash did not bind %s", name)
			}
		})
	}
}

func TestDryRunRejectsMaterializationAuthorityDrift(t *testing.T) {
	for name, mutate := range map[string]func(map[string]any){
		"automatic promotion enabled": func(materialization map[string]any) {
			materialization["no_promotion"] = false
		},
		"persistent promotion enabled": func(materialization map[string]any) {
			materialization["persistent_promotion"] = true
		},
		"direct KAH write enabled": func(materialization map[string]any) {
			materialization["direct_kah_state_write"] = true
		},
	} {
		t.Run(name, func(t *testing.T) {
			project := t.TempDir()
			writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")
			mutateRunLocalMaterialization(t, project, "run-20260616T105614Z-4b0ebe11b67d", mutate)
			result, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run})
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Status != "materialization_authority_drift" || result.ApprovalRequest.Required {
				t.Fatalf("expected materialization authority drift to fail closed, got %+v", result)
			}
			assertNoProjectLocalWrites(t, project)
		})
	}
}

func TestApplyRecomputesHashAndFailsClosedWithoutWriting(t *testing.T) {
	project := t.TempDir()
	writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")
	wrong, err := Apply(Options{
		Project:          project,
		RunID:            "run-20260616T105614Z-4b0ebe11b67d",
		TargetWorkflowID: "promoted-flow",
		ReuseReason:      "stable reuse evidence",
		Approval:         "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000",
		Runner:           workflowCapableRunner().Run,
	})
	if err != nil {
		t.Fatal(err)
	}
	if wrong.OK || wrong.Status != "approval_plan_hash_mismatch" || wrong.Approval.MatchedCurrentPlan {
		t.Fatalf("expected hash mismatch before apply behavior: %+v", wrong)
	}
	assertNoProjectLocalWrites(t, project)

	dryRun, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	blocked, err := Apply(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Approval: dryRun.ApprovalRequest.EvidenceRef, Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if blocked.OK || blocked.Status != "blocked_missing_kah_workflow_catalog_capability" || !blocked.Approval.MatchedCurrentPlan {
		t.Fatalf("expected correct-hash apply to fail closed while DAGSM-006 is absent: %+v", blocked)
	}
	assertNoProjectLocalWrites(t, project)
}

func TestApplyRefusesAdvertisedButUnmappedKAHApplySurface(t *testing.T) {
	project := t.TempDir()
	writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")
	dryRun, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowApplySurfaceRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK || dryRun.KAHCapability.ApplySurface == "" {
		t.Fatalf("test runner did not expose workflow apply surface: %+v", dryRun.KAHCapability)
	}
	blocked, err := Apply(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Approval: dryRun.ApprovalRequest.EvidenceRef, Runner: workflowApplySurfaceRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	if blocked.OK || blocked.Status != "workflow_promote_apply_refused" || !blocked.Approval.MatchedCurrentPlan {
		t.Fatalf("expected unmapped DAGSM-006 apply refusal, got %+v", blocked)
	}
	assertNoProjectLocalWrites(t, project)
}

func TestDryRunRequiresExplicitPromotionInputs(t *testing.T) {
	project := t.TempDir()
	writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")
	for name, opts := range map[string]Options{
		"missing target": {Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run},
		"missing reason": {Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", Runner: workflowCapableRunner().Run},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := BuildDryRun(opts)
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.ApprovalRequest.Required {
				t.Fatalf("expected fail-closed input validation, got %+v", result)
			}
		})
	}
}

func TestResultJSONCarriesPromotionMachinePacket(t *testing.T) {
	project := t.TempDir()
	writeRunLocalBundle(t, project, "run-20260616T105614Z-4b0ebe11b67d", "source-flow")
	result, err := BuildDryRun(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", TargetWorkflowID: "promoted-flow", ReuseReason: "stable reuse evidence", Runner: workflowCapableRunner().Run})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"machine_packet", "source_provenance", "target_paths", "generated_content", "base_checksums", "no_write", "approval_hash"} {
		if !strings.Contains(text, want) {
			t.Fatalf("result JSON missing %s: %s", want, text)
		}
	}
}

func writeRunLocalBundle(t *testing.T, project string, runID string, workflowID string) {
	t.Helper()
	base := filepath.Join(project, ".kkachi", "runs", runID, "workflow")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	workflowRel := filepath.ToSlash(filepath.Join(".kkachi", "runs", runID, "workflow", "workflow.yaml"))
	contractsRel := filepath.ToSlash(filepath.Join(".kkachi", "runs", runID, "workflow", "node-contracts.json"))
	materializationRel := filepath.ToSlash(filepath.Join(".kkachi", "runs", runID, "workflow", "materialization.json"))
	checksumsRel := filepath.ToSlash(filepath.Join(".kkachi", "runs", runID, "workflow", "checksums.json"))
	workflow := "workflow_id: " + workflowID + "\nschema_version: task-dag/v1\nnodes:\n  - id: setup\n    depends_on: []\n    join: all_of\n    required_outputs:\n      - artifacts/setup.md\n"
	falseValue := false
	contracts := nodeContractBundle{
		SchemaVersion: "kas-node-contracts/v1",
		Ref:           "run-local:" + runID + ":" + workflowID,
		Contracts: []workflowregistry.NodeContract{
			{
				WorkflowID: workflowID, NodeID: "setup", TaskClass: "development", OwnerRole: "planner_backend", ExecutionLane: "stage1_direct_codex_app_server",
				RequiredInputs: []string{"task-contract.yaml"}, ExpectedArtifacts: []string{"artifacts/setup.md"}, PromptRef: "skills/kkachi-plan/SKILL.md",
				ApprovalRequired: false, FallbackPolicy: "none_fail_closed", VerificationGate: "kah_workflow_node_evidence", CompletionAuthority: "kah_only", DirectKAHStateWrite: &falseValue,
			},
		},
	}
	contractBytes, err := json.MarshalIndent(contracts, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	contractBytes = append(contractBytes, '\n')
	if err := os.WriteFile(filepath.Join(project, filepath.FromSlash(workflowRel)), []byte(workflow), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, filepath.FromSlash(contractsRel)), contractBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	checksums := map[string]string{
		workflowRel:     checksumBytes([]byte(workflow)),
		contractsRel:    checksumBytes(contractBytes),
		"approval_hash": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
	}
	checksumBytesJSON, err := json.MarshalIndent(checksums, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	checksumBytesJSON = append(checksumBytesJSON, '\n')
	if err := os.WriteFile(filepath.Join(project, filepath.FromSlash(checksumsRel)), checksumBytesJSON, 0o644); err != nil {
		t.Fatal(err)
	}
	materialization := map[string]any{
		"ok":                     true,
		"command":                "workflow-materialize",
		"schema_version":         MaterializationSchema,
		"status":                 "materialized",
		"project":                project,
		"run_id":                 runID,
		"workflow_id":            workflowID,
		"workflow_file":          workflowRel,
		"node_contract_source":   contractsRel,
		"materialization_file":   materializationRel,
		"checksums_file":         checksumsRel,
		"run_local_posture":      "ephemeral_run_local_only",
		"no_promotion":           true,
		"persistent_promotion":   false,
		"source_evidence":        map[string]any{"selector_registry_checksum": "sha256:2222222222222222222222222222222222222222222222222222222222222222"},
		"direct_kah_state_write": false,
	}
	materializationBytes, err := json.MarshalIndent(materialization, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	materializationBytes = append(materializationBytes, '\n')
	if err := os.WriteFile(filepath.Join(project, filepath.FromSlash(materializationRel)), materializationBytes, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mutateRunLocalMaterialization(t *testing.T, project string, runID string, mutate func(map[string]any)) {
	t.Helper()
	path := filepath.Join(project, ".kkachi", "runs", runID, "workflow", "materialization.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var materialization map[string]any
	if err := json.Unmarshal(data, &materialization); err != nil {
		t.Fatal(err)
	}
	mutate(materialization)
	updated, err := json.MarshalIndent(materialization, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	updated = append(updated, '\n')
	if err := os.WriteFile(path, updated, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mutateRunLocalChecksums(t *testing.T, project string, runID string, mutate func(map[string]string)) {
	t.Helper()
	path := filepath.Join(project, ".kkachi", "runs", runID, "workflow", "checksums.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var checksums map[string]string
	if err := json.Unmarshal(data, &checksums); err != nil {
		t.Fatal(err)
	}
	mutate(checksums)
	updated, err := json.MarshalIndent(checksums, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	updated = append(updated, '\n')
	if err := os.WriteFile(path, updated, 0o644); err != nil {
		t.Fatal(err)
	}
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

func generatedContentByKind(t *testing.T, result Result, kind string) string {
	t.Helper()
	for _, item := range result.MachinePacket.GeneratedContent {
		if item.Kind == kind {
			return item.Content
		}
	}
	t.Fatalf("generated content kind %s not found: %+v", kind, result.MachinePacket.GeneratedContent)
	return ""
}

func assertNoProjectLocalWrites(t *testing.T, project string) {
	t.Helper()
	for _, forbidden := range []string{".kkachi/workflows", ".kkachi/workflow-catalog.yaml", ".kkachi-workflow.yaml", "providers", "tokens", "auth", "gateway", "models", "kab"} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(forbidden))); !os.IsNotExist(err) {
			t.Fatalf("apply wrote forbidden path %s: %v", forbidden, err)
		}
	}
}
