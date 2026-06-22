package docscontract

import (
	"strings"
	"testing"
)

func TestDesign003RouteAndMaterializerGuidanceIsDocumented(t *testing.T) {
	cases := map[string][]string{
		"skills/kkachi-workflow-route/SKILL.md": {
			"--project-has-teal-lane true|false",
			"--ui-ux-change true|false",
			"--teal-skip-reason <reason>",
			"teal_required = project_has_teal_lane && ui_ux_change",
			"teal_applicability",
			"teal_applicability_required",
			"teal_skip_reason_required",
			"does not materialize design nodes",
		},
		"skills/kkachi-workflow-trigger/SKILL.md": {
			"teal_applicability",
			"DESIGN_PLAN_GATE",
			"DESIGN_FIDELITY_REVIEW",
			"route_result_teal_applicability_required",
			"only when `teal_required=true`",
			"does not inject Teal nodes when `teal_required=false`",
		},
		"docs/sot/task-dag-workflow-contract.md": {
			"DESIGN-003",
			"`workflow-route` records `teal_applicability`",
			"`workflow-trigger` route-backed materialization injects Teal nodes only when `teal_required=true`",
			"route_result_teal_applicability_required",
			"not universal nodes in `registries/task-dag-workflow-registry.yaml`",
		},
		"docs/sot/teal-ui-workflow-policy.md": {
			"DESIGN-003 implements the KAS selector/materializer portion",
			"`workflow-route` derives and records `teal_required`",
			"`workflow-trigger` route-backed materialization inserts design gates only for `teal_required=true`",
			"KAH schema and gate enforcement remain DESIGN-004 and DESIGN-005",
		},
	}
	for rel, needles := range cases {
		requireContainsAll(t, rel, needles)
	}
}

func TestDesign003GuidancePreservesNoSubstitutionReadback(t *testing.T) {
	for _, rel := range []string{
		"skills/kkachi-workflow-route/SKILL.md",
		"skills/kkachi-workflow-trigger/SKILL.md",
		"skills/kkachi-task-contract/SKILL.md",
		"skills/kkachi-plan/SKILL.md",
		"skills/kkachi-review/SKILL.md",
		"skills/kkachi-final-verify/SKILL.md",
		"templates/run-artifacts/task-contract.yaml.tmpl",
		"templates/run-artifacts/phase-plan.yaml.tmpl",
	} {
		requireContainsAll(t, rel, []string{
			"DESIGN_PLAN_GATE",
			"DESIGN_FIDELITY_REVIEW",
			"required Teal verdict",
		})
	}
	route := readRepoFile(t, "skills/kkachi-workflow-route/SKILL.md")
	for _, forbidden := range []string{
		"assume teal_required=false",
		"fallback to ordinary color review",
		"fallback to MAR",
		"backend evidence may substitute",
	} {
		if strings.Contains(route, forbidden) {
			t.Fatalf("workflow route guidance contains forbidden substitution/fallback wording %q", forbidden)
		}
	}
}
