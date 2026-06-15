package workflowregistry

import (
	"strings"
	"testing"
)

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
`
}
