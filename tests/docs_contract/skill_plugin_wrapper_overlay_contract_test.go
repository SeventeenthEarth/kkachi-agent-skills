package docscontract

import "testing"

const skillPluginWrapperOverlaySOT = "docs/sot/skill-plugin-wrapper-overlay-contract.md"

func TestSKILLPluginWrapperOverlaySOTDefinesArchitecture(t *testing.T) {
	requireContainsAll(t, skillPluginWrapperOverlaySOT, []string{
		"Status: accepted SOT for `SKILL` epic",
		"SKILL-002 implements the plugin package, source role manifests, and dry-run update readback only",
		"KAS plugin base",
		"Color role manifests",
		"Thin profile wrappers",
		"Project overlays",
		"Guide skills",
		"Plugin update lifecycle",
		"Doctor and migration",
		"plugin-qualified names such as `kkachi-agent-skills:plan`",
		"Short names such as `kkachi-agent-skills:plan` are in-plugin canonicalization to registered base IDs such as `kkachi-plan`",
		"kkachi-agent-skills update plugin --dry-run",
		"`namespace`, `current_version`, `current_source`, `proposed_version`, `proposed_source`, `planned_changed_paths`, `role_manifest_impact`, `guide_skill_impact`, `no_write_evidence`, and `suggested_doctor_command`",
		"`suggested_doctor_command` is a non-executing recommendation string only",
		"`diagnostics` is intentionally empty until SKILL-004",
		"The explicit scoped surface is `update project-suite ...`; the bare `update ...` command remains a backward-compatible legacy alias",
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
		"Accepted SOT for moving KAS steady-state distribution",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"docs/sot/skill-plugin-wrapper-overlay-contract.md",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"### EPIC: SKILL",
		"| SKILL-001 | Accept plugin-wrapper-overlay SOT | Completed |",
		"| SKILL-002 | Implement Hermes plugin package, role manifests, and plugin update surface | In Review |",
		"`update project-suite` is the explicit scoped project-suite lifecycle alias",
		"Do not mark Completed until color review and final KAH gate pass",
		"plugin update readiness, ambiguous update command surfaces, KAH-boundary violations",
		"| SKILL-006 | Pilot one approved profile/project migration | Planned |",
		"no profile mutation or cleanup",
	})
}
