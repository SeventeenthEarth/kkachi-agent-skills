package workflowmaterializer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowcreator"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowregistry"
)

func TestMaterializeFromRouteWritesOnlyRunLocalWorkflowArtifacts(t *testing.T) {
	project := t.TempDir()
	registryPath := writeMaterializerRegistry(t, project, "demo")
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routePath := writeRouteResult(t, project, registry, "demo")

	result, err := MaterializeFromRoute(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", RouteResult: routePath})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Status != "materialized" || result.DirectKAHStateWrite {
		t.Fatalf("unexpected result: %+v", result)
	}
	for _, path := range result.WrittenPaths {
		if !strings.HasPrefix(path, ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/") {
			t.Fatalf("write escaped run-local workflow dir: %+v", result.WrittenPaths)
		}
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(path))); err != nil {
			t.Fatalf("missing written path %s: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi-workflow.yaml")); !os.IsNotExist(err) {
		t.Fatalf("materializer touched project graph: %v", err)
	}
	for _, forbidden := range []string{".kkachi/workflows", ".kkachi/workflow-catalog.yaml"} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(forbidden))); !os.IsNotExist(err) {
			t.Fatalf("materializer touched persistent workflow path %s: %v", forbidden, err)
		}
	}
	workflowBytes, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(result.WorkflowFile)))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(workflowBytes)
	for _, want := range []string{"workflow_id: demo", "schema_version: task-dag/v1", "  - id: setup", "    depends_on: []", "  - id: verify", "    depends_on: [setup]"} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing %q:\n%s", want, workflow)
		}
	}
	contractsBytes, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(result.NodeContractSource)))
	if err != nil {
		t.Fatal(err)
	}
	var contracts struct {
		SchemaVersion string                          `json:"schema_version"`
		Ref           string                          `json:"ref"`
		Contracts     []workflowregistry.NodeContract `json:"contracts"`
	}
	if err := json.Unmarshal(contractsBytes, &contracts); err != nil {
		t.Fatal(err)
	}
	if contracts.SchemaVersion != workflowregistry.NodeContractsVersion || contracts.Ref == "" || len(contracts.Contracts) != 2 {
		t.Fatalf("unexpected contract bundle: %+v", contracts)
	}
	for _, contract := range contracts.Contracts {
		if contract.CompletionAuthority != workflowregistry.KAHOnlyAuthority || contract.DirectKAHStateWrite == nil || *contract.DirectKAHStateWrite {
			t.Fatalf("contract authority drift: %+v", contract)
		}
	}
}

func TestMaterializeFromRouteFailsClosedForUnmatchedRouteAndSymlinkParent(t *testing.T) {
	project := t.TempDir()
	registryPath := writeMaterializerRegistry(t, project, "demo")
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routePath := writeRouteResult(t, project, registry, "demo")
	data, err := os.ReadFile(routePath)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), `"status":"bundle_route_matched"`, `"status":"bundle_no_match"`, 1))
	blockedRoute := filepath.Join(project, "blocked-route.json")
	if err := os.WriteFile(blockedRoute, data, 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := MaterializeFromRoute(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", RouteResult: blockedRoute})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "route_result_not_matched" || len(result.WrittenPaths) != 0 {
		t.Fatalf("expected unmatched route blocker, got %+v", result)
	}

	linkProject := t.TempDir()
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(linkProject, ".kkachi", "runs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(linkProject, ".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d")); err != nil {
		t.Fatal(err)
	}
	result, err = MaterializeFromRoute(Options{Project: linkProject, RunID: "run-20260616T105614Z-4b0ebe11b67d", RouteResult: routePath})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "unsafe_run_local_path" {
		t.Fatalf("expected symlink parent blocker, got %+v", result)
	}
}

func TestMaterializeFromCustomPacketRequiresApprovalAndMatchesHashBeforeWrite(t *testing.T) {
	project := t.TempDir()
	packetPath, approval := writeCustomWorkflowPacket(t, project, "demo")

	result, err := MaterializeFromCustomPacket(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", CustomWorkflowPacket: packetPath})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "approval_evidence_required" {
		t.Fatalf("expected approval blocker, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("missing approval wrote run-local state: %v", err)
	}

	wrong := strings.Replace(approval, approval[len(approval)-1:], "0", 1)
	if wrong == approval {
		wrong = strings.Replace(approval, approval[len(approval)-1:], "1", 1)
	}
	result, err = MaterializeFromCustomPacket(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", CustomWorkflowPacket: packetPath, Approval: wrong})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "approval_plan_hash_mismatch" {
		t.Fatalf("expected hash mismatch blocker, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("hash mismatch wrote run-local state: %v", err)
	}
}

func TestMaterializeFromCustomPacketRecomputesApprovalHashBeforeWrite(t *testing.T) {
	project := t.TempDir()
	packetPath, approval := writeCustomWorkflowPacket(t, project, "demo")
	data, err := os.ReadFile(packetPath)
	if err != nil {
		t.Fatal(err)
	}
	var packet workflowcreator.Result
	if err := json.Unmarshal(data, &packet); err != nil {
		t.Fatal(err)
	}
	packet.MachinePacket.SelectorMetadata["task_class"] = "tampered-after-approval"
	tampered, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(packetPath, append(tampered, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := MaterializeFromCustomPacket(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", CustomWorkflowPacket: packetPath, Approval: approval})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "custom_workflow_packet_hash_mismatch" {
		t.Fatalf("expected canonical packet hash mismatch, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("tampered packet wrote run-local state: %v", err)
	}
}

func TestMaterializeFromCustomPacketRequiresDeclaredDAGPath(t *testing.T) {
	project := t.TempDir()
	packetPath, _ := writeCustomWorkflowPacket(t, project, "demo")
	data, err := os.ReadFile(packetPath)
	if err != nil {
		t.Fatal(err)
	}
	var packet workflowcreator.Result
	if err := json.Unmarshal(data, &packet); err != nil {
		t.Fatal(err)
	}
	packet.MachinePacket.GeneratedContent[0].Path = ".kkachi/workflows/other.yaml"
	packet.MachinePacket.ApprovalHash = workflowcreator.RecomputeApprovalHash(packet)
	packet.ApprovalRequest.EvidenceRef = "dry-run:" + packet.MachinePacket.ApprovalHash
	packet.ApprovalRequest.DryRunPlanHash = packet.MachinePacket.ApprovalHash
	updated, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(packetPath, append(updated, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := MaterializeFromCustomPacket(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", CustomWorkflowPacket: packetPath, Approval: packet.ApprovalRequest.EvidenceRef})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "custom_workflow_dag_missing" {
		t.Fatalf("expected declared DAG path blocker, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("mismatched DAG path wrote run-local state: %v", err)
	}
}

func TestMaterializeFromCustomPacketWritesApprovedRunLocalArtifacts(t *testing.T) {
	project := t.TempDir()
	packetPath, approval := writeCustomWorkflowPacket(t, project, "demo")

	result, err := MaterializeFromCustomPacket(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", CustomWorkflowPacket: packetPath, Approval: approval})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Status != "materialized" || result.DirectKAHStateWrite || !result.NoPromotion || result.PersistentPromotion {
		t.Fatalf("unexpected custom materialization result: %+v", result)
	}
	if result.SourceEvidence.ApprovalEvidence != approval || result.SourceEvidence.DryRunPlanHash == "" || result.SourceEvidence.CustomWorkflowPacketChecksum == "" {
		t.Fatalf("missing approval/source evidence: %+v", result.SourceEvidence)
	}
	if result.CustomWorkflowPacketCopy != ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/custom-workflow-packet.json" {
		t.Fatalf("unexpected custom packet copy path: %+v", result)
	}
	for _, rel := range []string{"workflow.yaml", "node-contracts.json", "custom-workflow-packet.json", "materialization.json", "checksums.json"} {
		if _, err := os.Stat(filepath.Join(project, ".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow", rel)); err != nil {
			t.Fatalf("missing materialized %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi", "workflows")); !os.IsNotExist(err) {
		t.Fatalf("custom materialization promoted persistent workflow state: %v", err)
	}
	workflowBytes, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(result.WorkflowFile)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(workflowBytes), "workflow_id: demo") || !strings.Contains(string(workflowBytes), "  - id: plan") {
		t.Fatalf("workflow.yaml was not written from approved DAG content:\n%s", string(workflowBytes))
	}
	contractsBytes, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(result.NodeContractSource)))
	if err != nil {
		t.Fatal(err)
	}
	var contracts struct {
		SchemaVersion string                          `json:"schema_version"`
		Ref           string                          `json:"ref"`
		Contracts     []workflowregistry.NodeContract `json:"contracts"`
	}
	if err := json.Unmarshal(contractsBytes, &contracts); err != nil {
		t.Fatal(err)
	}
	if contracts.Ref != "run-local:run-20260616T105614Z-4b0ebe11b67d:demo" || len(contracts.Contracts) != 1 {
		t.Fatalf("unexpected custom contract bundle: %+v", contracts)
	}
	contract := contracts.Contracts[0]
	if contract.CompletionAuthority != workflowregistry.KAHOnlyAuthority || contract.DirectKAHStateWrite == nil || *contract.DirectKAHStateWrite || contract.FallbackPolicy != workflowregistry.NoFallbackPolicy {
		t.Fatalf("custom contract authority drift: %+v", contract)
	}
	checksumsBytes, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(result.ChecksumsFile)))
	if err != nil {
		t.Fatal(err)
	}
	var checksums map[string]string
	if err := json.Unmarshal(checksumsBytes, &checksums); err != nil {
		t.Fatal(err)
	}
	if checksums["approval_hash"] == "" || checksums["approved_plan_hash"] == "" || checksums["custom_workflow_packet_checksum"] == "" {
		t.Fatalf("checksums missing approval packet evidence: %+v", checksums)
	}
}

func writeMaterializerRegistry(t *testing.T, dir string, workflowID string) string {
	t.Helper()
	path := filepath.Join(dir, "workflow-registry.yaml")
	content := `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: ` + workflowID + `
    workflow_path: .kkachi/workflows/` + workflowID + `.yaml
    selector:
      task_classes: [development]
      labels_any: []
      labels_all: []
      changed_surfaces_any: []
      risk_levels: []
      required_agents_all: []
      required_capabilities_all: [task_dag_schema_validation, workflow_instance_state]
    fallback_policy: none_fail_closed
node_contracts:
  - workflow_id: ` + workflowID + `
    node_id: setup
    task_class: development
    owner_role: planner_backend
    execution_lane: stage1_direct_codex_app_server
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/setup.md]
    prompt_ref: skills/kkachi-plan/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: kah_workflow_node_evidence
    completion_authority: kah_only
    direct_kah_state_write: false
  - workflow_id: ` + workflowID + `
    node_id: verify
    task_class: development
    owner_role: hermes
    execution_lane: direct_kas_skill
    required_inputs: [artifacts/setup.md]
    expected_artifacts: [artifacts/verify.md]
    prompt_ref: skills/kkachi-final-verify/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: kah_workflow_node_evidence
    completion_authority: kah_only
    direct_kah_state_write: false
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeCustomWorkflowPacket(t *testing.T, dir string, workflowID string) (string, string) {
	t.Helper()
	dag := "workflow_id: " + workflowID + "\nschema_version: task-dag/v1\nnodes:\n  - id: plan\n    depends_on: []\n    join: all_of\n    required_outputs:\n      - artifacts/plan.md\n"
	packet := workflowcreator.Result{
		OK:      true,
		Command: workflowcreator.Command,
		Mode:    workflowcreator.ModeDAGOnly,
		Status:  "dry_run_ready",
		Project: workflowcreator.ProjectEvidence{Path: dir},
		Workflow: workflowcreator.WorkflowEvidence{
			ID:            workflowID,
			SchemaVersion: "task-dag/v1",
		},
		MachinePacket: workflowcreator.MachinePacket{
			SchemaVersion:    workflowcreator.PacketSchemaVersion,
			ApprovalSchema:   workflowcreator.ApprovalSchema,
			Canonicalization: workflowcreator.Canonicalization,
			TargetPaths:      []string{".kkachi/workflows/" + workflowID + ".yaml", ".kkachi/workflows/" + workflowID + "-node-contracts.yaml", ".kkachi/workflow-catalog.yaml"},
			CandidatePaths: workflowcreator.CandidatePaths{
				WorkflowDAG:          ".kkachi/workflows/" + workflowID + ".yaml",
				Catalog:              ".kkachi/workflow-catalog.yaml",
				NodeContractRegistry: ".kkachi/workflows/" + workflowID + "-node-contracts.yaml",
			},
			GeneratedContent: []workflowcreator.GeneratedContent{
				{Path: ".kkachi/workflows/" + workflowID + ".yaml", Kind: "workflow_dag", Content: dag, SHA256: checksumBytes([]byte(dag))},
			},
			SelectorMetadata: map[string]any{"task_class": "development"},
			NodeContracts: []workflowcreator.NodeContract{
				{
					WorkflowID:          workflowID,
					NodeID:              "plan",
					TaskClass:           "development",
					OwnerRole:           "planner_backend",
					ExecutionLane:       "stage1_direct_codex_app_server",
					RequiredInputs:      []string{"task-contract.yaml"},
					ExpectedArtifacts:   []string{"artifacts/plan.md"},
					PromptRef:           "skills/kkachi-plan/SKILL.md",
					ApprovalRequired:    false,
					FallbackPolicy:      workflowregistry.NoFallbackPolicy,
					VerificationGate:    "kah_workflow_node_evidence",
					CompletionAuthority: workflowregistry.KAHOnlyAuthority,
					DirectKAHStateWrite: false,
				},
			},
			TriggerPlan: workflowcreator.TriggerPlan{Mode: workflowcreator.ModeDAGOnly, Generated: false, DelegatesTo: "kkachi-agent-skills workflow-trigger", CustomLogic: false},
			BaseChecksums: map[string]string{
				".kkachi/workflows/" + workflowID + ".yaml":                "missing",
				".kkachi/workflows/" + workflowID + "-node-contracts.yaml": "missing",
				".kkachi/workflow-catalog.yaml":                            "missing",
			},
			ChangedPaths: []workflowcreator.ChangedPath{
				{Path: ".kkachi/workflow-catalog.yaml", Action: "create", Kind: "workflow_catalog"},
				{Path: ".kkachi/workflows/" + workflowID + "-node-contracts.yaml", Action: "create", Kind: "node_contract_registry"},
				{Path: ".kkachi/workflows/" + workflowID + ".yaml", Action: "create", Kind: "workflow_dag"},
			},
			Conflicts:   []workflowcreator.Conflict{},
			Diagnostics: []workflowcreator.Diagnostic{},
			NoWrite:     workflowcreator.NoWriteEvidence{Guaranteed: true},
		},
		DirectKAHStateWrite: false,
	}
	packet.MachinePacket.ApprovalHash = workflowcreator.RecomputeApprovalHash(packet)
	packet.ApprovalRequest = workflowcreator.ApprovalRequest{
		Required:                     true,
		EvidenceRef:                  "dry-run:" + packet.MachinePacket.ApprovalHash,
		DryRunPlanHash:               packet.MachinePacket.ApprovalHash,
		HashIncludesTargetPaths:      true,
		HashIncludesGeneratedContent: true,
	}
	path := filepath.Join(dir, "workflow-create-dry-run.json")
	data, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, packet.ApprovalRequest.EvidenceRef
}

func writeRouteResult(t *testing.T, dir string, registry workflowregistry.Registry, workflowID string) string {
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
