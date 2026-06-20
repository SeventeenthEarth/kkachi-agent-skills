package docscontract

import "testing"

func TestMARTL002WorkflowSurfacesRepresentMARLoop(t *testing.T) {
	requireContainsAll(t, "docs/sot/mar-task-loop-contract.md", []string{
		"MARTL-002 should be limited to KAS alignment and adoption",
		"`REQUEST_CHANGES` | Non-terminal.",
		"Blue/Red disposition must route accepted changes to the selected implementer lane",
		"verification and refreshed MAR are required",
	})
	requireContainsAll(t, "registries/task-dag-workflow-registry.yaml", []string{
		"node_id: mar_review",
		"execution_lane: mar_role_first_review",
		"expected_artifacts: [mar-review.md, mar-merge-pack.md, mar-blue-disposition.md]",
		"node_id: second_color_review",
		"expected_artifacts: [second-color-review.md, post-mar-feedback-triage.md]",
		"required_inputs: [checklist.md, diff.patch, test-log.md, mar-review.md, mar-merge-pack.md, mar-blue-disposition.md, second-color-review.md]",
	})
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"mar_task_loop_policy:",
		"contract_version: \"martl.v1\"",
		"required_roles:",
		"- logic",
		"- test_adequacy",
		"non_terminal_statuses:",
		"- REQUEST_CHANGES",
		"request_changes_rule:",
		"Provider availability, prompt rendering, dispatch success, model wording, or model voting is not review completion authority.",
		"mar_review:",
		"second_color_review:",
	})
}

func TestMARTL002RunArtifactTemplatesAndSkillsEnforceReReview(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/checklist.md.tmpl", []string{
		"mar_review | yes for development",
		"second_color_review | yes for development",
		"REQUEST_CHANGES/BLOCKED/DEGRADED/FAILED are non-terminal",
		"Development runs have MARTL MAR evidence and post-MAR color review",
	})
	requireContainsAll(t, "templates/run-artifacts/phase-plan.yaml.tmpl", []string{
		"mar_task_loop:",
		"terminal_candidate_statuses: [PASS, PASS_WITH_FINDINGS]",
		"non_terminal_statuses: [REQUEST_CHANGES, BLOCKED, DEGRADED, FAILED]",
		"phase: mar_review",
		"phase: second_color_review",
		"Accepted REQUEST_CHANGES findings require selected implementer-lane mutation",
	})
	requireContainsAll(t, "skills/kkachi-review/SKILL.md", []string{
		"MARTL treats `REQUEST_CHANGES`, `BLOCKED`, `DEGRADED`, and `FAILED` as non-terminal",
		"refreshed MAR evidence must be captured",
		"post-MAR Red/Orange/Gray review must close before final/pre-commit reporting",
	})
	requireContainsAll(t, "skills/kkachi-final-verify/SKILL.md", []string{
		"MARTL closure requires `mar-review.md`, `mar-merge-pack.md`, `mar-blue-disposition.md`, and `second-color-review.md`",
		"accepted requested changes must be routed to the selected implementer lane",
		"refreshed MAR evidence plus post-MAR color review must exist before final completion",
	})
}
