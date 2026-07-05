package docscontract

import (
	"strings"
	"testing"
)

func TestV02FLOW005UltragoalExecutorLoopPolicyIsDocumented(t *testing.T) {
	requireContainsAll(t, "docs/sot/v020-gjc-workflow-train-corrections.md", []string{
		"V02FLOW-005 executor-loop policy",
		"create-goals -> complete-goals -> execute-goal -> checkpoint -> verify -> repeat-or-terminate",
		"`complete-goals` freezes the selected goal bundle only; it is not implementation completion, implementation acceptance, verification, MAR, second-color, final, or commit readiness.",
		"`goal_bundle.status`, `goal_bundle.goals[].status`, and `executor_candidate.candidate_status` must reject `accepted`.",
		"A top-level packet `status: accepted`, if retained, is a derived/read-only post-final-gate alias of the dedicated gated `acceptance` object only.",
		"Blue direct patch exception records must include reason, scope, why executor loop was not used, changed-surface refs, verification required, review required, MAR required, final gate required, and waiver ref.",
	})
}

func TestV02FLOW005UltragoalPacketTemplateEncodesExecutorLoopAndAcceptanceBoundary(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/gjc-ultragoal-packet.yaml.tmpl", []string{
		"executor_loop_contract:",
		"loop_sequence:",
		"create-goals",
		"complete-goals",
		"execute-goal",
		"checkpoint",
		"verify",
		"repeat-or-terminate",
		"accepted_forbidden_in_pre_final_statuses:",
		"goal_bundle.status",
		"goal_bundle.goals[].status",
		"executor_candidate.candidate_status",
		"top_level_accepted_rule: derived_read_only_post_final_gate_alias_only",
		"blue_direct_patch_exception_record:",
		"why_executor_loop_was_not_used",
		"review_required",
		"mar_required",
		"final_gate_required",
	})
}

func TestV02FLOW005PhaseContractRegistersExecutorLoopStatusBoundaries(t *testing.T) {
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"V02FLOW-005 ultragoal executor-loop policy",
		"create-goals -> complete-goals -> execute-goal -> checkpoint -> verify -> repeat-or-terminate",
		"goal_bundle.goals[].status accepted is forbidden before final acceptance",
		"executor_candidate.candidate_status accepted is forbidden before final acceptance",
		"top-level accepted is derived/read-only after final gate only",
	})
}

func TestV02FLOW005ActiveGuidanceRejectsStaleExecutorLoopSpellings(t *testing.T) {
	activeSurfaces := []string{
		"docs/sot/v020-gjc-workflow-train-corrections.md",
		"docs/roadmap.md",
		"skills/kkachi-implement/SKILL.md",
	}
	for _, rel := range activeSurfaces {
		content := readRepoFile(t, rel)
		for _, line := range strings.Split(content, "\n") {
			if strings.Contains(line, "create-goals -> complete-goals -> execute -> checkpoint -> verify -> repeat") {
				t.Fatalf("%s has stale execute loop spelling in line %q", rel, line)
			}
			if strings.Contains(line, "create-goals -> complete-goals -> execute-goal -> checkpoint -> verify -> repeat`") {
				t.Fatalf("%s has stale repeat-only executor loop spelling in line %q", rel, line)
			}
			if strings.Contains(line, "create-goals -> complete-goals -> execute-goal -> checkpoint -> verify -> repeat,") {
				t.Fatalf("%s has stale repeat-only executor loop spelling in line %q", rel, line)
			}
		}
	}
}
