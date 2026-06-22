package workflowmaterializer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
	for _, forbidden := range []string{"      - artifacts/setup.md\n", "      - artifacts/verify.md\n"} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("workflow required_outputs must not target project-root artifacts, found %q:\n%s", forbidden, workflow)
		}
	}
	for _, want := range []string{
		"      - .kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/artifacts/setup.md\n",
		"      - .kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/artifacts/verify.md\n",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing run-local required output %q:\n%s", want, workflow)
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
		for _, artifact := range contract.ExpectedArtifacts {
			if !strings.HasPrefix(artifact, ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/") {
				t.Fatalf("run-local node contract expected_artifacts must not target project root: %+v", contract.ExpectedArtifacts)
			}
		}
	}
}

func TestMaterializeFromRouteInjectsTealNodesOnlyWhenRequired(t *testing.T) {
	project := t.TempDir()
	registryPath := writeDesignMaterializerRegistry(t, project)
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routePath := writeRouteResultWithTeal(t, project, registry, "development_full", map[string]any{
		"contract_version":               "design003.v1",
		"project_has_teal_lane":          true,
		"ui_ux_change":                   true,
		"teal_required":                  true,
		"derivation":                     "project_has_teal_lane && ui_ux_change",
		"teal_skip_reason":               "",
		"teal_waiver_approved":           false,
		"teal_waiver_approval_ref":       "",
		"teal_waiver_scope":              "",
		"teal_waiver_expires_at":         "",
		"required_when_teal_required":    []string{"DESIGN_PLAN_GATE", "DESIGN_FIDELITY_REVIEW"},
		"missing_required_status":        "required_teal_verdict_missing",
		"ordinary_review_is_substitute":  false,
		"mar_review_is_substitute":       false,
		"backend_evidence_is_substitute": false,
		"helper_notes_are_substitute":    false,
	})

	result, err := MaterializeFromRoute(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", RouteResult: routePath})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || !result.TealApplicability.TealRequired {
		t.Fatalf("expected Teal-required materialization, got %+v", result)
	}
	contractsBytes, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(result.NodeContractSource)))
	if err != nil {
		t.Fatal(err)
	}
	var contracts struct {
		Contracts []workflowregistry.NodeContract `json:"contracts"`
	}
	if err := json.Unmarshal(contractsBytes, &contracts); err != nil {
		t.Fatal(err)
	}
	nodeIDs := []string{}
	for _, contract := range contracts.Contracts {
		nodeIDs = append(nodeIDs, contract.NodeID)
		if strings.HasPrefix(contract.NodeID, "design_") {
			if contract.FallbackPolicy != workflowregistry.NoFallbackPolicy || contract.CompletionAuthority != workflowregistry.KAHOnlyAuthority || contract.DirectKAHStateWrite == nil || *contract.DirectKAHStateWrite {
				t.Fatalf("design node authority drift: %+v", contract)
			}
		}
	}
	wantOrder := []string{"plan", "design_plan_gate", "implement", "design_fidelity_review", "final_verify"}
	if !reflect.DeepEqual(nodeIDs, wantOrder) {
		t.Fatalf("unexpected Teal node order: got %v want %v", nodeIDs, wantOrder)
	}
	if !strings.Contains(string(contractsBytes), "DESIGN_PLAN_GATE") || !strings.Contains(string(contractsBytes), "DESIGN_FIDELITY_REVIEW") {
		t.Fatalf("materialized contracts missing Teal verdict semantics:\n%s", string(contractsBytes))
	}

	workflowBytes, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(result.WorkflowFile)))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(workflowBytes)
	if !strings.Contains(workflow, "  - id: design_plan_gate") || !strings.Contains(workflow, "    depends_on: [plan]") ||
		!strings.Contains(workflow, "  - id: design_fidelity_review") || !strings.Contains(workflow, "    depends_on: [implement]") {
		t.Fatalf("workflow.yaml missing Teal boundary nodes:\n%s", workflow)
	}
}

func TestMaterializeFromRouteFailsClosedWhenTealRequiredAnchorsAreMissing(t *testing.T) {
	for _, tc := range []struct {
		name     string
		missing  string
		anchorID string
	}{
		{name: "missing implement", missing: "implement", anchorID: "implement"},
		{name: "missing final verify", missing: "final_verify", anchorID: "final_verify"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project := t.TempDir()
			registryPath := writeDesignMaterializerRegistry(t, project)
			removeDesignRegistryNode(t, registryPath, tc.missing)
			registry, err := workflowregistry.Load(registryPath)
			if err != nil {
				t.Fatal(err)
			}
			routePath := writeRouteResultWithTeal(t, project, registry, "development_full", map[string]any{
				"contract_version":               "design003.v1",
				"project_has_teal_lane":          true,
				"ui_ux_change":                   true,
				"teal_required":                  true,
				"derivation":                     "project_has_teal_lane && ui_ux_change",
				"teal_skip_reason":               "",
				"required_when_teal_required":    []string{"DESIGN_PLAN_GATE", "DESIGN_FIDELITY_REVIEW"},
				"missing_required_status":        "required_teal_verdict_missing",
				"ordinary_review_is_substitute":  false,
				"mar_review_is_substitute":       false,
				"backend_evidence_is_substitute": false,
				"helper_notes_are_substitute":    false,
			})

			result, err := MaterializeFromRoute(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", RouteResult: routePath})
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Status != "route_result_teal_anchor_missing" || len(result.WrittenPaths) != 0 {
				t.Fatalf("expected missing %s anchor blocker with no writes, got %+v", tc.anchorID, result)
			}
			if len(result.Diagnostics) == 0 || !strings.Contains(result.Diagnostics[0].Message, tc.anchorID) {
				t.Fatalf("expected diagnostic to name missing %s anchor, got %+v", tc.anchorID, result.Diagnostics)
			}
			if _, err := os.Stat(filepath.Join(project, ".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow")); !os.IsNotExist(err) {
				t.Fatalf("missing anchor path must not write workflow artifacts: %v", err)
			}
		})
	}
}

func TestMaterializeFromRouteSkipsTealNodesForNonUIAndRequiresApplicability(t *testing.T) {
	project := t.TempDir()
	registryPath := writeDesignMaterializerRegistry(t, project)
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routePath := writeRouteResultWithoutTeal(t, project, registry, "development_full")
	result, err := MaterializeFromRoute(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", RouteResult: routePath})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "route_result_teal_applicability_required" || len(result.WrittenPaths) != 0 {
		t.Fatalf("expected missing Teal applicability blocker with no writes, got %+v", result)
	}

	skipRoute := writeRouteResultWithTeal(t, project, registry, "development_full", map[string]any{
		"contract_version":               "design003.v1",
		"project_has_teal_lane":          false,
		"ui_ux_change":                   false,
		"teal_required":                  false,
		"derivation":                     "project_has_teal_lane && ui_ux_change",
		"teal_skip_reason":               "No UI/UX surface in this project/task.",
		"teal_waiver_approved":           false,
		"teal_waiver_approval_ref":       "",
		"teal_waiver_scope":              "",
		"teal_waiver_expires_at":         "",
		"required_when_teal_required":    []string{"DESIGN_PLAN_GATE", "DESIGN_FIDELITY_REVIEW"},
		"missing_required_status":        "required_teal_verdict_missing",
		"ordinary_review_is_substitute":  false,
		"mar_review_is_substitute":       false,
		"backend_evidence_is_substitute": false,
		"helper_notes_are_substitute":    false,
	})
	result, err = MaterializeFromRoute(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", RouteResult: skipRoute})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.TealApplicability.TealRequired || result.TealApplicability.TealSkipReason == "" {
		t.Fatalf("expected non-UI Teal skip evidence, got %+v", result)
	}
	contractsBytes, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(result.NodeContractSource)))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contractsBytes), "design_plan_gate") || strings.Contains(string(contractsBytes), "DESIGN_PLAN_GATE") || strings.Contains(string(contractsBytes), "design_fidelity_review") {
		t.Fatalf("non-UI source work must not receive Teal nodes:\n%s", string(contractsBytes))
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

func TestMaterializeFromRouteFailsClosedForUnsafeRunLocalArtifactPaths(t *testing.T) {
	project := t.TempDir()
	registryPath := writeMaterializerRegistry(t, project, "demo")
	registryBytes, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	registryBytes = []byte(strings.Replace(string(registryBytes), "expected_artifacts: [artifacts/setup.md]", "expected_artifacts: [../../../final-report.md]", 1))
	if err := os.WriteFile(registryPath, registryBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routePath := writeRouteResult(t, project, registry, "demo")

	result, err := MaterializeFromRoute(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", RouteResult: routePath})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "unsafe_run_local_artifact_path" || len(result.WrittenPaths) != 0 {
		t.Fatalf("expected unsafe artifact path blocker with no writes, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("unsafe artifact path wrote run-local state: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, "final-report.md")); !os.IsNotExist(err) {
		t.Fatalf("unsafe artifact path created project-root evidence: %v", err)
	}
}

func TestMaterializeFromCustomPacketRejectsProjectRootEvidencePaths(t *testing.T) {
	project := t.TempDir()
	packetPath, approval := writeCustomWorkflowPacketWithArtifact(t, project, "demo", "artifacts/plan.md")

	result, err := MaterializeFromCustomPacket(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", CustomWorkflowPacket: packetPath, Approval: approval})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "custom_workflow_run_local_path_required" || len(result.WrittenPaths) != 0 {
		t.Fatalf("expected custom packet run-local path blocker, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("root-directed custom packet wrote run-local state: %v", err)
	}
}

func TestMaterializeFromCustomPacketRejectsInlineProjectRootRequiredOutputs(t *testing.T) {
	project := t.TempDir()
	packetPath, approval := writeCustomWorkflowPacketWithInlineDAGOutput(t, project, "demo", "artifacts/plan.md", ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/artifacts/plan.md")

	result, err := MaterializeFromCustomPacket(Options{Project: project, RunID: "run-20260616T105614Z-4b0ebe11b67d", CustomWorkflowPacket: packetPath, Approval: approval})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "custom_workflow_run_local_path_required" || len(result.WrittenPaths) != 0 {
		t.Fatalf("expected inline DAG root path blocker, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("inline root-directed custom packet wrote run-local state: %v", err)
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

	wrong := approval[:len(approval)-1] + "0"
	if wrong == approval {
		wrong = approval[:len(approval)-1] + "1"
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

func writeDesignMaterializerRegistry(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "workflow-registry-design.yaml")
	content := `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: development_full
    workflow_path: .kkachi/workflows/development_full.yaml
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
  - workflow_id: development_full
    node_id: plan
    task_class: development
    owner_role: planner_backend
    execution_lane: stage1_direct_codex_app_server
    required_inputs: [task-contract.yaml]
    expected_artifacts: [plan.md]
    prompt_ref: skills/kkachi-plan/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: kah_workflow_node_evidence
    completion_authority: kah_only
    direct_kah_state_write: false
  - workflow_id: development_full
    node_id: implement
    task_class: development
    owner_role: implementer_backend
    execution_lane: stage1_direct_codex_app_server
    required_inputs: [plan.md]
    expected_artifacts: [diff.patch]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: true
    fallback_policy: none_fail_closed
    verification_gate: kah_workflow_node_evidence
    completion_authority: kah_only
    direct_kah_state_write: false
  - workflow_id: development_full
    node_id: final_verify
    task_class: development
    owner_role: hermes
    execution_lane: direct_kas_skill
    required_inputs: [diff.patch]
    expected_artifacts: [verification.md]
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

func removeDesignRegistryNode(t *testing.T, path string, nodeID string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	blocks := map[string]string{
		"implement": `  - workflow_id: development_full
    node_id: implement
    task_class: development
    owner_role: implementer_backend
    execution_lane: stage1_direct_codex_app_server
    required_inputs: [plan.md]
    expected_artifacts: [diff.patch]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: true
    fallback_policy: none_fail_closed
    verification_gate: kah_workflow_node_evidence
    completion_authority: kah_only
    direct_kah_state_write: false
`,
		"final_verify": `  - workflow_id: development_full
    node_id: final_verify
    task_class: development
    owner_role: hermes
    execution_lane: direct_kas_skill
    required_inputs: [diff.patch]
    expected_artifacts: [verification.md]
    prompt_ref: skills/kkachi-final-verify/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: kah_workflow_node_evidence
    completion_authority: kah_only
    direct_kah_state_write: false
`,
	}
	block := blocks[nodeID]
	if block == "" {
		t.Fatalf("unknown node id %q", nodeID)
	}
	updated := strings.Replace(string(data), block, "", 1)
	if updated == string(data) {
		t.Fatalf("node %s block was not removed", nodeID)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCustomWorkflowPacket(t *testing.T, dir string, workflowID string) (string, string) {
	return writeCustomWorkflowPacketWithArtifact(t, dir, workflowID, ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/artifacts/plan.md")
}

func writeCustomWorkflowPacketWithInlineDAGOutput(t *testing.T, dir string, workflowID string, dagArtifactPath string, contractArtifactPath string) (string, string) {
	packetPath, _ := writeCustomWorkflowPacketWithArtifact(t, dir, workflowID, contractArtifactPath)
	data, err := os.ReadFile(packetPath)
	if err != nil {
		t.Fatal(err)
	}
	var packet workflowcreator.Result
	if err := json.Unmarshal(data, &packet); err != nil {
		t.Fatal(err)
	}
	dag := "workflow_id: " + workflowID + "\nschema_version: task-dag/v1\nnodes:\n  - id: plan\n    depends_on: []\n    join: all_of\n    required_outputs: [" + dagArtifactPath + "]\n"
	for i := range packet.MachinePacket.GeneratedContent {
		if packet.MachinePacket.GeneratedContent[i].Kind == "workflow_dag" {
			packet.MachinePacket.GeneratedContent[i].Content = dag
			packet.MachinePacket.GeneratedContent[i].SHA256 = checksumBytes([]byte(dag))
		}
	}
	packet.MachinePacket.ApprovalHash = workflowcreator.RecomputeApprovalHash(packet)
	packet.ApprovalRequest.EvidenceRef = "dry-run:" + packet.MachinePacket.ApprovalHash
	packet.ApprovalRequest.DryRunPlanHash = packet.MachinePacket.ApprovalHash
	data, err = json.MarshalIndent(packet, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(packetPath, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return packetPath, packet.ApprovalRequest.EvidenceRef
}

func writeCustomWorkflowPacketWithArtifact(t *testing.T, dir string, workflowID string, artifactPath string) (string, string) {
	t.Helper()
	dag := "workflow_id: " + workflowID + "\nschema_version: task-dag/v1\nnodes:\n  - id: plan\n    depends_on: []\n    join: all_of\n    required_outputs:\n      - " + artifactPath + "\n"
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
					RequiredInputs:      []string{".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/task-contract.yaml"},
					ExpectedArtifacts:   []string{artifactPath},
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
	return writeRouteResultWithTeal(t, dir, registry, workflowID, map[string]any{
		"contract_version":               "design003.v1",
		"project_has_teal_lane":          false,
		"ui_ux_change":                   false,
		"teal_required":                  false,
		"derivation":                     "project_has_teal_lane && ui_ux_change",
		"teal_skip_reason":               "No UI/UX surface in this project/task.",
		"teal_waiver_approved":           false,
		"teal_waiver_approval_ref":       "",
		"teal_waiver_scope":              "",
		"teal_waiver_expires_at":         "",
		"required_when_teal_required":    []string{"DESIGN_PLAN_GATE", "DESIGN_FIDELITY_REVIEW"},
		"missing_required_status":        "required_teal_verdict_missing",
		"ordinary_review_is_substitute":  false,
		"mar_review_is_substitute":       false,
		"backend_evidence_is_substitute": false,
		"helper_notes_are_substitute":    false,
	})
}

func writeRouteResultWithoutTeal(t *testing.T, dir string, registry workflowregistry.Registry, workflowID string) string {
	return writeRouteResultWithTeal(t, dir, registry, workflowID, nil)
}

func writeRouteResultWithTeal(t *testing.T, dir string, registry workflowregistry.Registry, workflowID string, teal map[string]any) string {
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
	if teal != nil {
		payload["teal_applicability"] = teal
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
