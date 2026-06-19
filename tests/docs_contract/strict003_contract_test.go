package docscontract

import "testing"

func TestSTRICT003WorkflowManagedRouteTriggerContract(t *testing.T) {
	cases := map[string][]string{
		"skills/kkachi-workflow-trigger/SKILL.md": {
			"--workflow-managed",
			"workflow-route",
			"ok:true/status:bundle_route_matched",
			"route-backed run-local materialization evidence",
			"Direct explicit or selector dispatch is rejected before KAH calls",
			"no fallback to the legacy development spine",
		},
		"docs/sot/kas-cli-contract.md": {
			"STRICT-003 adds `--workflow-managed`",
			"`ok:true`, `status: bundle_route_matched`",
			"selected workflow id/bundle trace",
			"`.kkachi/runs/<run_id>/workflow/materialization.json`",
			"must be route-result backed, not a custom one-off packet",
			"Selector mode, explicit workflow mode, custom packet mode",
			"legacy default development spine",
		},
		"docs/sot/strict-workflow-execution-contract.md": {
			"`STRICT-003`",
			"`workflow-trigger --workflow-managed`",
			"preserved `workflow-route` evidence",
			"route-backed safe resume evidence",
			"KAH ready-node-derived dispatch packets",
			"no selector/explicit/custom fallback",
		},
		"docs/roadmap.md": {
			"| STRICT-003 | Make classification route/trigger mandatory before KAS dispatch | Completed |",
			"`workflow-trigger --workflow-managed`",
			"route-backed safe resume readback",
			"selector/explicit/custom bypass failures",
			"not KAB `native_codex`",
		},
	}
	for rel, needles := range cases {
		requireContainsAll(t, rel, needles)
	}
}
