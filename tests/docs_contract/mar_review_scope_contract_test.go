package docscontract

import "testing"

func TestMARReviewScopeIsRoleFirstAndFailClosed(t *testing.T) {
	coreScopeFiles := []string{
		"skills/kkachi-request-feedback/SKILL.md",
		"skills/kkachi-prompt-compose/SKILL.md",
		"skills/kkachi-final-verify/SKILL.md",
		"skills/kkachi-implement/SKILL.md",
		"skills/kkachi-review/SKILL.md",
		"skills/kkachi-orchestrate/SKILL.md",
		"README.md",
		"templates/project-overlay/references/backend-policy.md.tmpl",
	}

	for _, rel := range coreScopeFiles {
		requireContainsAll(t, rel, []string{
			"MAR",
			"role",
			"logic",
			"security",
			"arch",
			"cve",
			"test_adequacy",
		})
	}

	requireContainsAll(t, "skills/kkachi-request-feedback/SKILL.md", []string{
		"MAR is the default independent review lane",
		"Required role coverage is `logic`, `security`, `arch`, `cve`, and `test_adequacy`",
		"unresolved required role coverage fails closed",
		"must not turn provider availability, prompt rendering, or dispatch success into clean review coverage",
	})

	requireContainsAll(t, "skills/kkachi-prompt-compose/SKILL.md", []string{
		"role-first reviewer matrix",
		"raw-output cap",
		"dispatch success is not review completion",
	})

	requireContainsAll(t, "skills/kkachi-final-verify/SKILL.md", []string{
		"MAR review is role-first and fail-closed",
		"Provider availability, prompt rendering, or dispatch success alone is not review completion evidence",
		"degraded/failed/unresolved required roles as clean PASS",
	})
}
