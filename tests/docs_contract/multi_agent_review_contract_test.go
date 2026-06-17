package docscontract

import (
	"testing"
)

var marPromptTemplates = []string{
	"templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl",
	"templates/prompts/mar/kimi-k2-6-reviewer-request.md.tmpl",
	"templates/prompts/mar/antigravity-gemini-reviewer-request.md.tmpl",
	"templates/prompts/mar/premium-reviewer-request.md.tmpl",
}

func TestMultiAgentReviewSkillScaffoldDefinesReadOnlyReviewContract(t *testing.T) {
	requireContainsAll(t, "skills/kkachi-multi-agent-review/SKILL.md", []string{
		"kkachi-multi-agent-review",
		"Multi-agent review",
		"MAR-002 does not execute providers",
		"fail-closed",
		"DEGRADED",
		"BLOCKED",
		"Red adjudication",
		"Codex/Claude premium reviewers require explicit approval",
	})
}

func TestMultiAgentReviewReferencesDefineReviewerPolicy(t *testing.T) {
	requireContainsAll(t, "skills/kkachi-multi-agent-review/references/reviewer-role-matrix.md", []string{
		"zcode_glm_5_2",
		"kimi_k2_6",
		"antigravity_gemini",
		"premium_approval_required",
	})

	requireContainsAll(t, "skills/kkachi-multi-agent-review/references/premium-escalation-guide.md", []string{
		"Codex",
		"Claude",
		"explicit 주군 approval",
		"no automatic premium fallback",
	})
}

func TestMultiAgentReviewPromptTemplatesPreserveReadOnlyBoundaries(t *testing.T) {
	for _, rel := range marPromptTemplates {
		requireContainsAll(t, rel, []string{
			"read-only review",
			"Do not edit files",
			"Do not run tests",
			"Do not run linters",
			"Do not run builds",
			"Do not run installs",
			"Do not run package managers",
			"Do not start services",
			"Do not run network probes",
			"Do not run runtime verification commands",
			"Do not mutate auth",
			"provider execution is not completion evidence",
		})
	}
}

func TestMultiAgentReviewRunArtifactsCaptureBlueAndRedDisposition(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/mar-blue-disposition.md.tmpl", []string{
		"Blue disposition",
		"coverage_state",
		"red_adjudication_trigger",
		"no_provider_execution_claim",
	})

	requireContainsAll(t, "templates/run-artifacts/mar-red-adjudication-handoff.md.tmpl", []string{
		"Red adjudication handoff",
		"blocker",
		"high_risk",
		"disagreement",
		"degraded",
		"premium_escalation",
	})
}

func TestMultiAgentReviewSkillPackRegistrationExists(t *testing.T) {
	requireContainsAll(t, "skill-pack.yaml", []string{
		"kkachi-multi-agent-review",
	})
}
