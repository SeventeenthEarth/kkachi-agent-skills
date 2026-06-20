package docscontract

import (
	"path/filepath"
	"strings"
	"testing"
)

func requireNotContains(t *testing.T, rel string, needles []string) {
	t.Helper()
	content := readRepoFile(t, rel)
	for _, needle := range needles {
		if strings.Contains(content, needle) {
			t.Fatalf("%s still contains stale or forbidden phrase %q", rel, needle)
		}
	}
}

func TestSTRICT007StatusAndStaleNextActionAreCurrent(t *testing.T) {
	requireContainsAll(t, "docs/sot/strict-workflow-execution-contract.md", []string{
		"| `STRICT-006` | KAH | Phase-plan projection and workflow consistency gate | Source-side complete |",
		"KAH source commit `f9d82b7`",
		"| `STRICT-007` | KAS | Strict orchestration skill/templates/e2e adoption | Completed |",
		"The shared STRICT epic is source-side complete across KAS and KAH",
		"route -> trigger -> ready -> start -> work -> required outputs -> complete",
		"effective KAH capability evidence is missing any of `task_dag_schema_validation`, `workflow_instance_state`, `workflow_strict_transition_ledger`, `workflow_transition_order_verification`, or `workflow_phase_projection_validation`",
	})
	requireNotContains(t, "docs/sot/strict-workflow-execution-contract.md", []string{
		"Next, advance to KAH `STRICT-006`",
		"| `STRICT-006` | KAH | Phase-plan projection and workflow consistency gate | Planned |",
		"Current KAS work is `STRICT-007`",
		"| `STRICT-007` | KAS | Strict orchestration skill/templates/e2e adoption | In progress |",
	})

	requireContainsAll(t, "docs/roadmap.md", []string{
		"| STRICT-007 | Adopt strict flow in active KAS orchestration skills/templates/e2e | Completed |",
		"Completed source-side after KAH `STRICT-006` source completion at `f9d82b7`",
		"workflow_phase_projection_validation",
		"run-local `required_outputs` materialization",
		"Red/Orange/Gray MAR PASS",
	})
	requireNotContains(t, "docs/roadmap.md", []string{
		"Depends on KAH `STRICT-006`",
		"| STRICT-007 | Adopt strict flow in active KAS orchestration skills/templates/e2e | In progress |",
	})
}

func TestSTRICT007WorkflowManagedAuthoritySurfacesAreStrict(t *testing.T) {
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"kah_workflow_instance_and_transition_ledger",
		"phase_projection_required: true",
		"phase-plan/checklist evidence against workflow instance state",
		"KAS must follow KAH ready/start/complete admission",
	})
	requireNotContains(t, "registries/phase-contracts.yaml", []string{
		"KHS phase_plan is the workflow SOT",
		"sot: phase_plan\n",
	})
	requireContainsAll(t, "templates/run-artifacts/phase-plan.yaml.tmpl", []string{
		"run_local_phase_plan_with_kah_workflow_projection_when_workflow_managed",
	})
}

func TestSTRICT007ActiveSkillAndTemplateAdoptionIsDocumented(t *testing.T) {
	requireContainsAll(t, "skills/kkachi-orchestrate/SKILL.md", []string{
		"route -> trigger -> ready -> start -> work -> required outputs -> complete",
		"stale `expected_start_revision`",
		"missing `workflow_phase_projection_validation`",
		"fallback workflow, fallback backend, or fallback agent",
	})
	requireContainsAll(t, "skills/kkachi-implement/SKILL.md", []string{
		"workflow node start --expect-revision <expected_start_revision>",
		"stale expected revisions",
		"missing required outputs",
		"KAH complete failures are blockers",
	})
	requireContainsAll(t, "skills/kkachi-final-verify/SKILL.md", []string{
		"route/materialization/dispatch evidence",
		"KAH node start and complete evidence",
		"transition ledger verification",
		"workflow_phase_projection_validation",
	})
	requireContainsAll(t, "skills/kkachi-workflow-trigger/SKILL.md", []string{
		"workflow_phase_projection_validation",
		"stale revision, KAH start rejection, missing required outputs, or KAH complete rejection is a fail-closed blocker",
	})
	requireContainsAll(t, "templates/run-artifacts/workflow-dispatch-packet.yaml.tmpl", []string{
		"required_outputs:",
		"kah_node_start_succeeded_at_expected_start_revision",
		"required_outputs_exist_before_complete",
		"workflow_phase_projection_validation_before_phase_completion_claim",
		"direct_kah_state_write: false",
	})
	requireContainsAll(t, "templates/runners/direct-codex-sdk-appserver-runner.py.tmpl", []string{
		"`kah_completion_claimed` false",
		"expected_start_revision` is stale",
		"required outputs are missing",
		"workflow_phase_projection_validation",
	})

	promptPaths := []string{
		"templates/prompts/claude/path-a-implement.md.tmpl",
		"templates/prompts/codex/path-a-implement.md.tmpl",
		"templates/prompts/gemini/path-a-implement.md.tmpl",
		"templates/prompts/glm/path-a-implement.md.tmpl",
		"templates/prompts/opencode/path-a-implement.md.tmpl",
	}
	for _, rel := range promptPaths {
		requireContainsAll(t, rel, []string{
			"Strict workflow-managed execution:",
			"do not begin backend/agent work until KAH node start succeeds at `expected_start_revision`",
			"missing required outputs",
			"KAH node complete failure",
			"Do not claim completion",
		})
		if filepath.Base(filepath.Dir(rel)) == "gemini" {
			requireContainsAll(t, rel, []string{"explicit post-approval start evidence"})
		}
	}
}
