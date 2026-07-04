package workflowregistry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var standardBundleNodeIDs = map[string][]string{
	"development_full":        {"codegraph_refresh", "plan", "ralplan", "implement", "enhance_test", "ai_slop_cleaner", "optimize", "update_docs", "request_feedback", "handle_feedback", "mar_review", "second_color_review", "final_verify", "improve"},
	"docs_only_light":         {"task_contract", "plan", "update_docs", "docs_validation", "final_verify"},
	"research_evidence_light": {"task_contract", "evidence_collection", "source_citation", "final_verify"},
	"review_light":            {"task_contract", "review_request", "feedback_evidence", "final_verify"},
	"bootstrap_config":        {"task_contract", "preflight", "configure", "verification", "final_verify"},
	"direct_report":           {"command_evidence", "final_report"},
}

var standardBundleTaskClasses = map[string]string{
	"development_full":        "development",
	"docs_only_light":         "docs_only",
	"research_evidence_light": "research_evidence",
	"review_light":            "collaboration_review",
	"bootstrap_config":        "bootstrap_config",
	"direct_report":           "simple_command_report",
}

func TestParseValidRegistryWithNestedSelectorAndContractLists(t *testing.T) {
	registry, err := Parse(validRegistryYAML("demo"))
	if err != nil {
		t.Fatal(err)
	}
	if registry.Version != RegistryVersion || len(registry.Workflows) != 1 || len(registry.NodeContracts) != 1 {
		t.Fatalf("unexpected registry: %+v", registry)
	}
	workflow := registry.Workflows[0]
	if workflow.Selector.TaskClasses[0] != "development" || workflow.Selector.RequiredCapabilitiesAll[1] != "workflow_instance_state" {
		t.Fatalf("selector not parsed: %+v", workflow.Selector)
	}
	contract := registry.NodeContracts[0]
	if contract.TaskClass != "development" || contract.RequiredInputs[0] != "task-contract.yaml" || contract.ExpectedArtifacts[0] != "artifacts/setup.md" {
		t.Fatalf("contract block lists not parsed: %+v", contract)
	}
	if contract.CompletionAuthority != KAHOnlyAuthority || contract.DirectKAHStateWrite == nil || *contract.DirectKAHStateWrite {
		t.Fatalf("contract authority fields not parsed: %+v", contract)
	}
}

func TestShippedRegistryContainsCanonicalStandardBundles(t *testing.T) {
	registry := loadShippedRegistry(t)
	if len(registry.Workflows) != len(standardBundleTaskClasses) {
		t.Fatalf("shipped registry workflow count = %d, want %d: %+v", len(registry.Workflows), len(standardBundleTaskClasses), registry.Workflows)
	}
	seen := map[string]bool{}
	for _, workflow := range registry.Workflows {
		seen[workflow.WorkflowID] = true
		if workflow.WorkflowID == "development-default" {
			t.Fatal("legacy development-default must not remain selectable in the shipped standard bundle registry")
		}
		if workflow.FallbackPolicy != NoFallbackPolicy {
			t.Fatalf("workflow %s fallback drifted: %s", workflow.WorkflowID, workflow.FallbackPolicy)
		}
	}
	for workflowID := range standardBundleTaskClasses {
		if !seen[workflowID] {
			t.Fatalf("shipped registry missing canonical bundle %s; seen=%v", workflowID, seen)
		}
	}
}

func TestShippedStandardBundlesSelectUniquelyAndCoverContracts(t *testing.T) {
	registry := loadShippedRegistry(t)
	for workflowID, taskClass := range standardBundleTaskClasses {
		t.Run(workflowID, func(t *testing.T) {
			match, err := Select(registry, Query{TaskClass: taskClass, RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}})
			if err != nil {
				t.Fatal(err)
			}
			if match.Status != "selector_matched" || match.Selected.WorkflowID != workflowID || len(match.Candidates) != 1 {
				t.Fatalf("expected unique %s match for task_class %s, got %+v", workflowID, taskClass, match)
			}
			contracts := ContractsForWorkflow(registry.NodeContracts, workflowID)
			if len(contracts) != len(standardBundleNodeIDs[workflowID]) {
				t.Fatalf("contract count for %s = %d, want %d", workflowID, len(contracts), len(standardBundleNodeIDs[workflowID]))
			}
			if err := ValidateContractsCoverNodeIDs(registry.NodeContracts, workflowID, standardBundleNodeIDs[workflowID]); err != nil {
				t.Fatalf("contracts do not cover expected nodes: %v", err)
			}
			for _, contract := range contracts {
				if contract.CompletionAuthority != KAHOnlyAuthority || contract.DirectKAHStateWrite == nil || *contract.DirectKAHStateWrite || contract.FallbackPolicy != NoFallbackPolicy {
					t.Fatalf("contract authority drift for %s/%s: %+v", workflowID, contract.NodeID, contract)
				}
			}
		})
	}
	missing, err := Select(registry, Query{RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}})
	if err == nil || missing.Status != "selector_required_input_missing" {
		t.Fatalf("expected missing task_class to fail closed, got %+v err=%v", missing, err)
	}
	noMatch, err := Select(registry, Query{TaskClass: "unknown", RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}})
	if err != nil || noMatch.Status != "selector_no_match" {
		t.Fatalf("expected unknown task_class to no-match, got %+v err=%v", noMatch, err)
	}
}

func TestParseFailsClosedForUnsupportedNestedShape(t *testing.T) {
	cases := map[string]string{
		"nested mapping under selector list": `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: demo
    workflow_path: .kkachi/workflows/demo.yaml
    selector:
      task_classes:
        - development
          nested: invalid
    fallback_policy: none_fail_closed
node_contracts:
  - workflow_id: demo
    node_id: setup
    task_class: development
    owner_role: implementer_backend
    execution_lane: direct_kas_skill
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/setup.md]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: make test
`,
		"mapping-style selector list item": `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: demo
    workflow_path: .kkachi/workflows/demo.yaml
    selector:
      task_classes:
        - name: development
    fallback_policy: none_fail_closed
node_contracts:
  - workflow_id: demo
    node_id: setup
    task_class: development
    owner_role: implementer_backend
    execution_lane: direct_kas_skill
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/setup.md]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: make test
`,
		"extra unsupported nesting": `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: demo
    workflow_path: .kkachi/workflows/demo.yaml
    selector:
      task_classes:
        - development
          - invalid
    fallback_policy: none_fail_closed
node_contracts:
  - workflow_id: demo
    node_id: setup
    task_class: development
    owner_role: implementer_backend
    execution_lane: direct_kas_skill
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/setup.md]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: make test
`,
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(input)
			if err == nil {
				t.Fatal("expected unsupported YAML shape to fail closed")
			}
			if !strings.Contains(err.Error(), "unsupported") && !strings.Contains(err.Error(), "expected") {
				t.Fatalf("expected parser shape failure, got %v", err)
			}
		})
	}
}

func TestParseFailsClosedForDuplicateNodeContracts(t *testing.T) {
	text := validRegistryYAML("demo") + `
  - workflow_id: demo
    node_id: setup
    task_class: development
    owner_role: implementer_backend
    execution_lane: direct_kas_skill
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/setup.md]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: make test
    completion_authority: kah_only
    direct_kah_state_write: false
`
	_, err := Parse(text)
	if err == nil || !strings.Contains(err.Error(), "duplicate node contract") {
		t.Fatalf("expected duplicate contract failure, got %v", err)
	}
}

func TestJSONAndYAMLNodeContractsShareCoreValidation(t *testing.T) {
	jsonContracts := []NodeContract{{
		WorkflowID:        "demo",
		NodeID:            "setup",
		OwnerRole:         "implementer_backend",
		ExecutionLane:     "direct_kas_skill",
		RequiredInputs:    []string{"task-contract.yaml"},
		ExpectedArtifacts: []string{"artifacts/setup.md"},
		PromptRef:         "skills/kkachi-implement/SKILL.md",
		ApprovalRequired:  false,
		FallbackPolicy:    NoFallbackPolicy,
		VerificationGate:  "make test",
	}}
	if err := ValidateNodeContracts(jsonContracts); err != nil {
		t.Fatalf("explicit JSON core contract should validate without task_class: %v", err)
	}
	registryWithoutTaskClass := strings.Replace(validRegistryYAML("demo"), "    task_class: development\n", "", 1)
	_, err := Parse(registryWithoutTaskClass)
	if err == nil || !strings.Contains(err.Error(), "registry node contracts require task_class") {
		t.Fatalf("registry YAML must add task_class invariant on top of shared core validation, got %v", err)
	}
	registryWithInvalidCore := strings.Replace(validRegistryYAML("demo"), "    fallback_policy: none_fail_closed\n", "    fallback_policy: retry\n", 1)
	_, err = Parse(registryWithInvalidCore)
	if err == nil || !strings.Contains(err.Error(), "unsupported fallback_policy") {
		t.Fatalf("registry YAML must share core node-contract validation, got %v", err)
	}
}

func TestRegistryNodeContractAuthorityFieldsFailClosed(t *testing.T) {
	cases := map[string]string{
		"missing completion authority": strings.Replace(validRegistryYAML("demo"), "    completion_authority: kah_only\n", "", 1),
		"wrong completion authority":   strings.Replace(validRegistryYAML("demo"), "    completion_authority: kah_only\n", "    completion_authority: kas_local\n", 1),
		"missing direct write":         strings.Replace(validRegistryYAML("demo"), "    direct_kah_state_write: false\n", "", 1),
		"direct write true":            strings.Replace(validRegistryYAML("demo"), "    direct_kah_state_write: false\n", "    direct_kah_state_write: true\n", 1),
		"node fallback drift":          strings.Replace(validRegistryYAML("demo"), "    fallback_policy: none_fail_closed\n    verification_gate: make test\n", "    fallback_policy: retry\n    verification_gate: make test\n", 1),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(input)
			if err == nil {
				t.Fatal("expected authority-field drift to fail closed")
			}
			if !strings.Contains(err.Error(), "completion_authority") && !strings.Contains(err.Error(), "direct_kah_state_write") && !strings.Contains(err.Error(), "fallback_policy") {
				t.Fatalf("expected authority-field error, got %v", err)
			}
		})
	}
}

func TestSelectHandlesZeroOneMany(t *testing.T) {
	one, err := Parse(validRegistryYAML("demo"))
	if err != nil {
		t.Fatal(err)
	}
	match, err := Select(one, Query{TaskClass: "development", ChangedSurfaces: []string{"code"}, RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}})
	if err != nil || match.Status != "selector_matched" || match.Selected.WorkflowID != "demo" {
		t.Fatalf("unexpected one-match result: %+v err=%v", match, err)
	}
	zero, err := Select(one, Query{TaskClass: "security"})
	if err != nil || zero.Status != "selector_no_match" {
		t.Fatalf("unexpected zero-match result: %+v err=%v", zero, err)
	}
	many, err := Parse(validRegistryYAML("demo") + strings.ReplaceAll(validRegistryYAML("demo2"), "version: kas-task-dag-workflow-registry/v1\n", ""))
	if err != nil {
		t.Fatal(err)
	}
	match, err = Select(many, Query{TaskClass: "development", ChangedSurfaces: []string{"code"}, RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}})
	if err != nil || match.Status != "selector_ambiguous" || len(match.Candidates) != 2 {
		t.Fatalf("unexpected many-match result: %+v err=%v", match, err)
	}
}

func validRegistryYAML(workflowID string) string {
	return `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: ` + workflowID + `
    workflow_path: .kkachi/workflows/` + workflowID + `.yaml
    selector:
      task_classes:
        - development
      labels_any: []
      labels_all: []
      changed_surfaces_any:
        - code
        - tests
      risk_levels: []
      required_agents_all: []
      required_capabilities_all:
        - task_dag_schema_validation
        - workflow_instance_state
    fallback_policy: none_fail_closed
node_contracts:
  - workflow_id: ` + workflowID + `
    node_id: setup
    task_class: development
    owner_role: implementer_backend
    execution_lane: direct_kas_skill
    required_inputs:
      - task-contract.yaml
    expected_artifacts:
      - artifacts/setup.md
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: make test
    completion_authority: kah_only
    direct_kah_state_write: false
`
}

func loadShippedRegistry(t *testing.T) Registry {
	t.Helper()
	path := filepath.Join("..", "..", "..", "registries", "task-dag-workflow-registry.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := Parse(string(data))
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
