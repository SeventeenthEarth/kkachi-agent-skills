package docscontract

import (
	"strings"
	"testing"
)

func TestV02FLOW007ReviewTrainAndAggregateWatcherPolicyIsDocumented(t *testing.T) {
	requireContainsAll(t, "docs/sot/v020-gjc-workflow-train-corrections.md", []string{
		"V02FLOW-007 review-train and aggregate-watcher policy",
		"first official Red/Orange/Gray color review plus dependent Blue synthesis",
		"mandatory MAR for development work unless an explicit task-specific waiver is recorded before the MAR gate",
		"second official Red/Orange/Gray color adoption/review plus dependent Blue final disposition",
		"one aggregate watcher per color round is state-report-only",
		"must not perform Blue synthesis, inject or fake `진행해`, auto-trigger continuation, waive lanes, mutate source, substitute temporary subagents or self-approval for official review, commit, push, install, release, or change runtime/auth/provider/gateway/profile/model state",
	})
}

func TestV02FLOW007RoadmapMarksFinalGateCompleted(t *testing.T) {
	requireRoadmapTaskStatus(t, "V02FLOW-007", "Completed")
	requireContainsAll(t, "docs/roadmap.md", []string{
		"mandatory KAH-native MAR `PASS_WITH_FINDINGS`",
		"post-MAR second-color review",
		"final gate `evt-006977` passed",
		"Commit/push/install/release/runtime activation remain separate approvals",
	})
}

func TestV02FLOW007ActiveSkillGuidanceCarriesReviewTrainAndWatcherBoundaries(t *testing.T) {
	for _, rel := range []string{
		"skills/kkachi-review/SKILL.md",
		"skills/kkachi-request-feedback/SKILL.md",
		"skills/kkachi-final-verify/SKILL.md",
		"skills/kkachi-phase-state/SKILL.md",
		"skills/kkachi-orchestrate/SKILL.md",
	} {
		requireContainsAll(t, rel, []string{
			"first color review -> mandatory MAR -> second color adoption/review -> Blue disposition",
			"aggregate watcher",
			"state-report-only",
			"must not perform Blue synthesis",
			"fake `진행해`",
			"auto-continue",
			"temporary subagents and delegate_task do not count as official color review, MAR role coverage, or Blue synthesis",
		})
	}
}

func TestV02FLOW007PhaseContractsRejectWatcherAndTemporaryAgentSubstitution(t *testing.T) {
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"V02FLOW-007 review-train authority",
		"first_color_review -> request_feedback_1 -> handle_feedback_1 -> mar_review -> second_color_review -> final_verify",
		"aggregate watchers are state-report-only and cannot perform Blue synthesis, fake proceed, auto-continue, waive lanes, or mutate source",
		"temporary subagents and delegate_task do not count as official color review, MAR role coverage, project-Gray review, or Blue synthesis",
	})

	for _, rel := range []string{
		"registries/task-dag-workflow-registry.yaml",
		"registries/graph-template-registry.yaml",
		"templates/workflow-graphs/kas-default.yaml",
	} {
		content := readRepoFile(t, rel)
		for _, required := range []string{"mar-review", "second-color-review"} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s missing required V02FLOW-007 workflow phase %q", rel, required)
			}
		}
		for _, forbidden := range []string{"review -> final", "review->final", "review to final"} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s contains stale direct review-to-final projection %q", rel, forbidden)
			}
		}
	}
}
