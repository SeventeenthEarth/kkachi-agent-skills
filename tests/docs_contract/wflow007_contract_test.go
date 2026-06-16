package docscontract

import "testing"

func TestWFLOW007WorkflowRouteContractIsDocumentedAndBounded(t *testing.T) {
	cases := map[string][]string{
		"skills/kkachi-workflow-route/SKILL.md": {
			"workflow-route",
			"already-classified metadata",
			"--classification-reason <reason>",
			"bundle_route_matched",
			"classification_reason_missing",
			"bundle_ambiguous",
			"bundle_selected_mismatch",
			"does not infer task class",
			"workflow create/show/ready/node APIs",
			"render",
			"dispatch packets",
			"materialize",
			"`.kkachi/runs/<run_id>/workflow.yaml`",
			"direct_kah_state_write:false",
		},
		"docs/sot/kas-cli-contract.md": {
			"WFLOW-007",
			"workflow-route",
			"already-classified metadata only",
			"registries/task-taxonomy.yaml",
			"bundle_route_matched",
			"classification_required_input_missing",
			"classification_reason_missing",
			"classification_class_unsupported",
			"bundle_no_match",
			"bundle_ambiguous",
			"bundle_selected_mismatch",
			"workflow create/show/ready/node APIs",
			"render",
			"dispatch packets",
			"materialize run-local workflows",
			"use an LLM tie-break",
		},
		"docs/sot/task-dag-workflow-contract.md": {
			"WFLOW-007",
			"workflow-route",
			"already-classified",
			"metadata",
			"exactly one bundle",
			"taxonomy checksum",
			"registry checksum",
			"direct_kah_state_write:false",
			"bundle_no_match",
			"bundle_ambiguous",
			"bundle_selected_mismatch",
			"workflow create/show/ready/node APIs",
			"render",
			"dispatch packets",
			"WFLOW-008 owns run-local workflow materialization",
			"WFLOW-009",
			"DAGSM-006",
			"promotion/apply",
		},
		"docs/roadmap.md": {
			"| WFLOW-007 | Implement task classification to bundle routing | Completed |",
			"route-only `workflow-route`",
			"no KAH workflow create/resume/ready-node or dispatch path",
			"run-20260616T083404Z-0626230bc56c",
			"52436478-0857-4452-9daa-2c00e14045f6",
			"evt-003666",
		},
		"skill-pack.yaml": {
			"kkachi-workflow-route",
		},
	}
	for rel, needles := range cases {
		requireContainsAll(t, rel, needles)
	}
}
