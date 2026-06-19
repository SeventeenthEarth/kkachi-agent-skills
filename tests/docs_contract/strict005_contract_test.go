package docscontract

import "testing"

func TestSTRICT005DispatchAndRunnerGuardContractIsDocumented(t *testing.T) {
	cases := map[string][]string{
		"templates/run-artifacts/workflow-dispatch-packet.yaml.tmpl": {
			"strict_order: true",
			"run_id",
			"instance_revision",
			"expected_start_revision",
			"ready_node_reasons",
			"completion_authority: kah_only",
			"direct_kah_state_write: false",
		},
		"templates/runners/direct-codex-sdk-appserver-runner.py.tmpl": {
			"--dispatch-packet",
			"strict_order must be true",
			"instance_revision must equal expected_start_revision",
			"expected_start_revision",
			"workflow node start",
			"thread.run",
			"workflow node complete",
			"completion evidence does not exist",
			"kah_completion_claimed",
		},
		"skills/kkachi-workflow-trigger/SKILL.md": {
			"workflow_strict_transition_ledger",
			"workflow_transition_order_verification",
			"expected_start_revision",
			"strict_workflow_expected_start_revision_missing",
		},
		"docs/sot/strict-workflow-execution-contract.md": {
			"strict_order",
			"instance_revision",
			"expected_start_revision",
			"workflow_strict_transition_ledger",
			"workflow_transition_order_verification",
			"no `thread.run` before KAH start succeeds",
			"kah_completion_claimed:false",
		},
		"docs/sot/stage1-direct-codex-sdk-appserver-runner.md": {
			"--expect-revision",
			"KAH start succeeds",
			"post-start revision",
			"kah_completion_claimed",
		},
	}
	for rel, needles := range cases {
		requireContainsAll(t, rel, needles)
	}
}
