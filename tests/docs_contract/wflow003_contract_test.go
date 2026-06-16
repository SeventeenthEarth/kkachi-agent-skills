package docscontract

import "testing"

func TestWFLOW003SelectorRegistryContractIsDocumentedAndBounded(t *testing.T) {
	cases := map[string][]string{
		"skills/kkachi-workflow-trigger/SKILL.md": {
			"WFLOW-003 selector mode",
			"--selector-registry <path>",
			"selector_no_match",
			"selector_ambiguous",
			"selector_explicit_mode_conflict",
			"does not auto-load",
			"completion_authority:kah_only",
			"stage1_direct_codex_is_kab_native_codex:false",
			"KAH completion authority",
			"agent fallback",
		},
		"docs/sot/kas-cli-contract.md": {
			"WFLOW-003",
			"kas-task-dag-workflow-registry/v1",
			"selector_no_match",
			"selector_ambiguous",
			"selector_explicit_mode_conflict",
			"Zero and multiple candidate",
			"must not call KAH workflow instance commands",
			"no choose-first",
			"direct KAH state",
		},
		"docs/sot/task-dag-workflow-contract.md": {
			"WFLOW-003",
			"kas-task-dag-workflow-registry/v1",
			"selector_no_match",
			"selector_ambiguous",
			"selector_explicit_mode_conflict",
			"Zero workflows matched",
			"Multiple workflows matched",
			"must not call KAH workflow instance commands",
			"no-write",
			"direct KAH state",
		},
	}
	for rel, needles := range cases {
		requireContainsAll(t, rel, needles)
	}
}

func TestWFLOW003DispatchPacketTemplatePreservesSelectorReadback(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/workflow-dispatch-packet.yaml.tmpl", []string{
		"selector_registry_source",
		"selector_registry_checksum",
		"selector_match",
		"task_class",
		"labels:",
		"changed_surfaces:",
		"required_capabilities:",
		"completion_authority: kah_only",
		"stage1_direct_codex_is_kab_native_codex: false",
		"direct_kah_state_write: false",
	})
}

func TestWFLOW003RegistryCarriesSharedNodeContractFields(t *testing.T) {
	requireContainsAll(t, "registries/task-dag-workflow-registry.yaml", []string{
		"version: kas-task-dag-workflow-registry/v1",
		"workflow_id: development_full",
		"workflow_id: docs_only_light",
		"workflow_id: research_evidence_light",
		"workflow_id: review_light",
		"workflow_id: bootstrap_config",
		"workflow_id: direct_report",
		"workflow_path: .kkachi/workflows/development_full.yaml",
		"required_capabilities_all: [task_dag_schema_validation, workflow_instance_state]",
		"node_contracts:",
		"task_class: development",
		"owner_role:",
		"execution_lane: stage1_direct_codex_app_server",
		"required_inputs:",
		"expected_artifacts:",
		"prompt_ref:",
		"approval_required:",
		"fallback_policy: none_fail_closed",
		"verification_gate: kah_workflow_node_evidence",
		"completion_authority: kah_only",
		"direct_kah_state_write: false",
	})
}
