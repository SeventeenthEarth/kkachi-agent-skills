package docscontract

import (
	"strings"
	"testing"
)

var activeFeedbackIntakeSurfaces = []string{
	"registries/phase-contracts.yaml",
	"templates/run-artifacts/phase-plan.yaml.tmpl",
	"templates/run-artifacts/checklist.md.tmpl",
	"templates/run-artifacts/task-contract.yaml.tmpl",
	"skills/kkachi-plan/SKILL.md",
	"skills/kkachi-request-feedback/SKILL.md",
	"skills/kkachi-handle-feedback/SKILL.md",
	"skills/kkachi-orchestrate/SKILL.md",
}

func TestFeedbackIntakeActiveSurfacesUseConfigurableBounds(t *testing.T) {
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"minimum_rounds: 1",
		"maximum_rounds: 5",
		"round_1_required: true",
		"conditional_feedback_rounds_2_to_5",
		"feedback-5.md_when_round_5_runs",
		"handle-feedback-5.md_when_round_5_runs",
	})

	requireContainsAll(t, "templates/run-artifacts/phase-plan.yaml.tmpl", []string{
		"min_rounds: 1",
		"max_rounds: 5",
		"phase: request_feedback_4",
		"phase: handle_feedback_4",
		"phase: request_feedback_5",
		"phase: handle_feedback_5",
		"feedback-5.md",
		"handle-feedback-5.md",
	})

	requireContainsAll(t, "templates/run-artifacts/checklist.md.tmpl", []string{
		"request_feedback_4",
		"handle_feedback_4",
		"request_feedback_5",
		"handle_feedback_5",
		"feedback-5.md",
		"handle-feedback-5.md",
	})

	requireContainsAll(t, "templates/run-artifacts/task-contract.yaml.tmpl", []string{
		"feedback_rounds:",
		"min: 1",
		"max: 5",
	})

	requireContainsAll(t, "skills/kkachi-request-feedback/SKILL.md", []string{
		"optional continuation rounds 2..5",
		"never exceed five request-feedback/handle-feedback pairs",
		"feedback-2.md` through `feedback-5.md",
	})

	requireContainsAll(t, "skills/kkachi-handle-feedback/SKILL.md", []string{
		"feedback-triage-4.md",
		"handle-feedback-4.md",
		"feedback-triage-5.md",
		"handle-feedback-5.md",
	})
}

func TestFeedbackIntakeActiveSurfacesAvoidFixedThreeRoundClaims(t *testing.T) {
	for _, rel := range activeFeedbackIntakeSurfaces {
		content := readRepoFile(t, rel)
		forbidden := []string{
			"maximum_rounds: 3",
			"max_rounds: 3",
			"conditional_feedback_rounds_2_to_3",
			"must not exceed round 3",
			"at most three rounds",
			"up to three rounds",
			"never exceed three request-feedback/handle-feedback pairs",
			"rounds 2-3 as conditional",
			"Maximum feedback round.",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Fatalf("%s retains fixed-three feedback-intake claim %q", rel, phrase)
			}
		}
	}
}
