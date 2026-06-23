package docscontract

import "testing"

const skillPluginWrapperOverlaySOT = "docs/sot/skill-plugin-wrapper-overlay-contract.md"

func TestSKILLPluginWrapperOverlaySOTDefinesArchitecture(t *testing.T) {
	requireContainsAll(t, skillPluginWrapperOverlaySOT, []string{
		"Status: candidate SOT for `SKILL` epic",
		"KAS plugin base",
		"Color role manifests",
		"Thin profile wrappers",
		"Project overlays",
		"Guide skills",
		"Plugin update lifecycle",
		"Doctor and migration",
		"plugin-qualified names such as `kkachi-agent-skills:plan`",
		"kkachi-agent-skills update plugin --dry-run",
		"KAH has no implementation responsibility for `update plugin`",
		"Profile-local full KAS base copy present | warning during migration, error after cutover",
	})
}

func TestSKILLProjectOverlayContractDefinesLayoutAndMergePolicy(t *testing.T) {
	requireContainsAll(t, skillPluginWrapperOverlaySOT, []string{
		"skills/<project>/kas-overlays/<project>-<role>-<base-skill>-overlay/SKILL.md",
		"overlay_for: kkachi-agent-skills:plan",
		"merge_mode: additive_constraints",
		"replacement_candidate",
		"replacement_approved",
		"If an overlay conflicts with plugin base safety, role authority, or a current SOT, the agent must stop",
		"copied full base skill bodies",
		"hidden fallback when the plugin base is missing",
	})
}

func TestSKILLDocsRegistrationAndRoadmap(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"sot/skill-plugin-wrapper-overlay-contract.md",
		"SKILL plugin-wrapper-overlay SOT",
		"plugin base + thin color wrapper + profile-local project overlay architecture",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"docs/sot/skill-plugin-wrapper-overlay-contract.md",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"### EPIC: SKILL",
		"| SKILL-001 | Accept plugin-wrapper-overlay SOT | Completed |",
		"Implement Hermes plugin package, role manifests, and plugin update surface",
		"plugin update readiness, ambiguous update command surfaces, KAH-boundary violations",
		"| SKILL-006 | Pilot one approved profile/project migration | Planned |",
		"no profile mutation or cleanup",
	})
}
