package docscontract

import (
	"testing"
)

var marPromptTemplates = []string{
	"templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl",
	"templates/prompts/mar/kimi-default-reviewer-request.md.tmpl",
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
		"kimi_default",
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
			"Role ID",
			"Provider ID",
			"Role scope",
			"Acceptance criteria",
			"role-scoped acceptance-criteria verdicts",
		})
	}
}

func TestMultiAgentReviewRunArtifactsCaptureBlueAndRedDisposition(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/mar-blue-disposition.md.tmpl", []string{
		"Blue disposition",
		"coverage_state",
		"red_adjudication_trigger",
		"no_provider_execution_claim",
		"provider_attempt_paths",
		"required_roles",
		"covered_roles",
		"unresolved_required_roles",
		"operator_report_text",
		"role_provider_candidates",
		"role_fallback_reasons",
		"role_acceptance_criteria_matrix",
		"red_trigger_summary",
		"Provider Attempt Coverage",
		"Blue Matrix Inputs",
	})

	requireContainsAll(t, "templates/run-artifacts/mar-red-adjudication-handoff.md.tmpl", []string{
		"Red adjudication handoff",
		"blocker",
		"high_risk",
		"disagreement",
		"degraded",
		"unresolved_required_roles",
		"operator_report_text",
		"premium_escalation",
	})
}

func TestMultiAgentReviewSkillPackRegistrationExists(t *testing.T) {
	requireContainsAll(t, "skill-pack.yaml", []string{
		"kkachi-multi-agent-review",
	})
}

func TestMultiAgentReviewSOTDefinesRoleFirstFailureDecisionPolicy(t *testing.T) {
	requireContainsAll(t, "docs/sot/multi-agent-review-policy.md", []string{
		"MAR-005 makes review role coverage the default MAR completion unit",
		"logic",
		"security",
		"arch",
		"cve",
		"test_adequacy",
		"registries/mar-provider-lanes.json",
		"mar.role_lanes.v1",
		"scripts/mar.py role-lanes",
		"zcode_glm_5_2",
		"kimi_default",
		"unresolved_required_roles",
		"주군/operator report",
		"adapter_proof_required",
		"mutation_detected",
		"automatic alternate-provider substitution",
	})
}

func TestMultiAgentReviewSOTDefinesMAR006KAHRunnerOwnershipBoundary(t *testing.T) {
	requireContainsAll(t, "docs/sot/multi-agent-review-policy.md", []string{
		"MAR-006",
		"selecting which `kkachi-agent-helper` binary its own doctor",
		"workflow-create",
		"workflow-promote",
		"workflow-trigger",
		"graphsync /",
		"workflow-graph repair surfaces execute",
		"Resolver success is not KAH capability, evidence, project-state, or gate success.",
		"`mar_provider_tools.resolved_argv` is authority only for MAR provider command",
		"execution. It must not be used as KAH CLI selection authority.",
		"`KKACHI_KAH_BIN`",
		"`kah_cli_path`/`kah_cli`",
		"ambient PATH only when no",
		"explicit repo KAH selection exists",
	})
}

func TestRoadmapDefinesMAR006WithoutClaimingKAHGateBehavior(t *testing.T) {
	requireContainsAll(t, "docs/roadmap.md", []string{
		"MAR-006",
		"Resolve KAH CLI through repo toolchain-aware KAS runner",
		"doctor, workflow-create, workflow-promote, workflow-trigger, and graphsync/workflow-graph repair",
		"Explicit repo selection mismatch or missing binaries fail closed",
		"MAR provider `resolved_argv` overlay semantics remain distinct from KAH CLI selection",
		"KAS binary selection does not claim KAH deterministic evidence/gate behavior",
	})
}

func TestMultiAgentReviewProviderLaneRegistryIsRegistered(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"registries/mar-provider-lanes.json",
		"role lane readback",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"registries/mar-provider-lanes.json",
	})
}
