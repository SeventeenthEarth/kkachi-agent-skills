package docscontract

import "testing"

func TestPOLPR008AliasesAreExplicitCompatibilityTranslations(t *testing.T) {
	requireContainsAll(t, "registries/graph-template-registry.yaml", []string{
		"POLPR-008 closeout",
		"older phase-contract/skill activity aliases remain translation-only compatibility names",
		"explicit compatibility translations",
		"not pending phase-renaming work or fallback authority",
	})
	requireContainsAll(t, "docs/sot/graph-template-registry.md", []string{
		"POLPR-008 compatibility translation",
		"aliases are translation-only names",
		"not pending phase-renaming\nwork",
		"graph fallback authority",
	})
	requireContainsAll(t, "docs/sot/phase-orchestration-policy.md", []string{
		"explicit translation-only compatibility names",
		"not\npending reconciliation work",
		"graph fallback authority",
	})
	requireContainsAll(t, "skills/kkachi-orchestrate/SKILL.md", []string{
		"older activity aliases such as `update_docs` and `final_verify` are translation-only compatibility names",
	})
}

func TestPOLPR008ActiveGuidanceDoesNotLeaveAliasReconciliationPending(t *testing.T) {
	for _, rel := range []string{
		"registries/graph-template-registry.yaml",
		"docs/sot/graph-template-registry.md",
		"docs/sot/phase-orchestration-policy.md",
		"skills/kkachi-orchestrate/SKILL.md",
	} {
		requireContainsNone(t, rel, []string{
			"until POLPR-008/POLPR-005 reconciliation",
			"downstream reconciliation remains\nPOLPR-008/POLPR-005",
			"may still appear in compatibility surfaces until downstream",
		}, "active POLPR-008 mirror guidance must not describe alias translation as pending reconciliation")
	}
}
