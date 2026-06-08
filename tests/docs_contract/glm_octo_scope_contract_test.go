package docscontract

import "testing"

func TestGLMOctoReviewScopeForbidsTestExecution(t *testing.T) {
	coreScopeFiles := []string{
		"skills/kkachi-request-feedback/SKILL.md",
		"skills/kkachi-prompt-compose/SKILL.md",
		"skills/kkachi-final-verify/SKILL.md",
		"skills/kkachi-implement/SKILL.md",
		"skills/kkachi-review/SKILL.md",
		"skills/kkachi-orchestrate/SKILL.md",
		"docs/sot/khs-architecture-and-integration.md",
		"README.md",
		"templates/project-overlay/references/backend-policy.md.tmpl",
	}

	for _, rel := range coreScopeFiles {
		requireContainsAll(t, rel, []string{
			"implemented code",
			"tests",
			"linters",
			"builds",
			"installs",
			"package managers",
			"network probes",
			"service starts",
			"runtime verification",
		})
	}

	requireContainsAll(t, "skills/kkachi-request-feedback/SKILL.md", []string{
		"requirements-and-implemented-code-only review lane",
		"must explicitly forbid running tests",
		"must not create new verification by executing commands",
		"Hermes/KAB must reject any Octo permission request outside that read-only inspection scope",
		"if a forbidden command is approved or executed, the official Octo gate fails closed",
		"During permission handling, approve only read-only inspection commands",
	})

	requireContainsAll(t, "skills/kkachi-prompt-compose/SKILL.md", []string{
		"followed immediately by an explicit requirements-and-implemented-code-only review scope",
		"It must forbid running tests",
		"permission rejection evidence for any out-of-scope command request",
	})

	requireContainsAll(t, "skills/kkachi-final-verify/SKILL.md", []string{
		"if the prompt omits the no-test/code-only scope",
		"if Octo executed a forbidden command",
		"show test/lint/build/install/network/service/runtime execution during Octo",
	})
}
