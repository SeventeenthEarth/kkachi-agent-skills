package docscontract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToken006SlimmedSkillsKeepDirectReferenceLinks(t *testing.T) {
	t.Helper()

	cases := map[string][]string{
		"skills/kkachi-orchestrate/SKILL.md": {
			"references/kas-activation-scope.md",
			"references/kas-installed-profile-vs-source-promotion.md",
			"references/run-operating-policy.md",
			"references/orchestration-responsibilities.md",
		},
		"skills/kkachi-plan/SKILL.md": {
			"references/planner-lane-and-capture.md",
			"references/checklist-normalization.md",
		},
		"skills/kkachi-phase-state/SKILL.md": {
			"references/implemented-kah-command-surface.md",
			"references/run-state-flow.md",
		},
		"skills/kkachi-final-verify/SKILL.md": {
			"references/pre-commit-completion-report-template.md",
			"references/review-readiness-and-final-gate.md",
		},
		"skills/kkachi-request-feedback/SKILL.md": {
			"references/glm-octo-review-lane.md",
		},
		"skills/kkachi-implement/SKILL.md": {
			"references/bridge-observation-and-start-rules.md",
		},
	}

	for skillPath, refs := range cases {
		requireContainsAll(t, skillPath, refs)
		for _, ref := range refs {
			refPath := filepath.Join(repoRoot(t), filepath.Dir(skillPath), ref)
			if _, err := os.Stat(refPath); err != nil {
				t.Fatalf("direct reference %q from %s is missing or unreadable: %v", refPath, skillPath, err)
			}
		}
	}
}

func TestToken006SlimmedSkillsKeepCoreDiscoverabilityInMainSkillFiles(t *testing.T) {
	requireContainsAll(t, "skills/kkachi-plan/SKILL.md", []string{
		"fallback audit",
		"codegraph index <repo>",
		"codegraph init -i <repo>",
		"`checklist.md`",
	})
	requireContainsAll(t, "skills/kkachi-phase-state/SKILL.md", []string{
		"`phase-plan init`, `phase-plan show`, `phase-plan set`, `phase-plan validate`",
		"graph-evidence.md",
		".kkachi-workflow.yaml",
		"`artifact set-status`",
	})
}
