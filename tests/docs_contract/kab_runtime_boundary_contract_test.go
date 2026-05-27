package docscontract

import (
	"strings"
	"testing"
)

var kabBoundaryActiveSurfaces = []string{
	"README.md",
	"registries/phase-contracts.yaml",
	"skills/kkachi-orchestrate/SKILL.md",
	"skills/kkachi-implement/SKILL.md",
	"skills/kkachi-plan/SKILL.md",
	"docs/sot/phase-orchestration-policy.md",
	"templates/run-artifacts/plan.md.tmpl",
	"templates/run-artifacts/checklist.md.tmpl",
}

func TestKABRuntimeBoundaryPreservesRequiredKabLaterSurfaces(t *testing.T) {
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"backend_execution: required_when_backend_or_bridge_evidence_needed",
		"automated_review_by_different_tool: kab_later_until_transport_evidenced",
		"plan_lifecycle: required_when_kab_plan_surface_is_used",
		"bridge_evidence: required_when_runtime_completion_or_backend_identity_is_claimed",
		"kas_kah_local_exception:",
		"kab_later_required_for:",
		"- backend_execution",
		"- automated_review_by_different_tool",
		"- kab_plan_lifecycle",
		"- bridge_evidence",
	})

	requireContainsAll(t, "README.md", []string{
		"KAB remains",
		"required for backend execution, automated review-by-different-tool transport,",
		"KAB plan lifecycle, and bridge evidence when those surfaces are in scope.",
		"Scoped KAS/KAH-local CLIMVP, GRAPHMVP, and docs-only maintenance may proceed without KAB",
		"must not claim backend execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence.",
	})

	requireContainsAll(t, "templates/run-artifacts/checklist.md.tmpl", []string{
		"KAB evidence when KAB-backed",
		"Feedback rounds are between 1 and 5.",
		"KAB evidence when the run is KAB-backed or claims backend runtime evidence.",
	})
}

func TestKABRuntimeBoundaryAvoidsOverbroadKabPrerequisiteClaims(t *testing.T) {
	for _, rel := range kabBoundaryActiveSurfaces {
		content := readRepoFile(t, rel)
		forbidden := []string{
			"KHS code-change and development runs use KAB as the rule",
			"Docs-only KHS runs also use KAB by default",
			"Do not bypass KAB for KHS code-change or development runs",
			"KHS must copy `plan.plan_text` from KAB into this artifact before implementation starts.",
			"Code-change KHS runs use KAB; if the user forbids KAB for code changes",
			"local + KAH/KAB evidence",
			"Feedback rounds are between 1 and 3.",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Fatalf("%s retains overbroad KAB prerequisite or stale boundary claim %q", rel, phrase)
			}
		}
	}
}
