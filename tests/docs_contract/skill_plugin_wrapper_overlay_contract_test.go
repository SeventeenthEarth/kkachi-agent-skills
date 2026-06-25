package docscontract

import (
	"strings"
	"testing"
)

const skillPluginWrapperOverlaySOT = "docs/sot/skill-plugin-wrapper-overlay-contract.md"

func TestSKILLPluginWrapperOverlaySOTDefinesArchitecture(t *testing.T) {
	requireContainsAll(t, skillPluginWrapperOverlaySOT, []string{
		"Status: accepted SOT for `SKILL` epic",
		"SKILL-002 implements the plugin package and source role manifests without any public `update` or `migrate` CLI surface",
		"KAS plugin base",
		"Color role manifests",
		"Thin project wrappers",
		"Project overlays",
		"Guide skills",
		"No CLI update/migrate lifecycle",
		"Doctor and overlay refresh",
		"plugin-qualified names such as `kkachi-agent-skills:plan`",
		"Short names such as `kkachi-agent-skills:plan` are in-plugin canonicalization to registered base IDs such as `kkachi-plan`",
		"must not expose public `update` or `migrate` commands",
		"`kkachi-agent-skills-overlay-refresh` skill",
		"mark the temporary legacy archive with `metadata.kas.kind: project_overlay_legacy`",
		"after successful refresh and review, remove `skills/<project>/<project>-overlay-legacy/`",
		"kkachi-agent-skills doctor --plugin --repo <kas-repo> --json",
		"`--repo` is required for plugin doctor mode",
		"`doctor_mode_ambiguous`",
		"The `kkachi-agent-skills` CLI must not expose public `update` or `migrate` commands",
		"KAH has no implementation responsibility for overlay refresh",
		"Profile-local full KAS base copy present | warning during migration, error after cutover",
	})
}

func TestSKILLProjectOverlayContractDefinesLayoutAndMergePolicy(t *testing.T) {
	requireContainsAll(t, skillPluginWrapperOverlaySOT, []string{
		"skills/<project>/<project>-overlay/SKILL.md",
		"skills/<project>/<project>-overlay/references/*.md",
		"applies_to:",
		"kkachi-agent-skills:plan",
		"merge_mode: additive_constraints",
		"refresh_skill: kkachi-agent-skills:kkachi-agent-skills-overlay-refresh",
		"generated wrappers must not hard-code Blue metadata",
		"role-aware `applies_to` lists that match the selected suite role subset",
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
		"| SKILL-002 | Implement Hermes plugin package and role manifests without update/migrate CLI | In Review |",
		"| SKILL-004 | Implement read-only SKILL doctor diagnostics | In Review |",
		"`doctor --plugin --repo <kkachi-agent-skills-repo>` distinguishes `plugin_base`, `project_wrapper`, `project_overlay`, `project_overlay_legacy`, `legacy_copied_base_suite`, `personal_skill`, and `unknown_source`",
		"The public CLI must not expose `update` or `migrate` commands",
		"Do not mark Completed until color review and final KAH gate pass",
		"public update/migrate command exposure, KAH-boundary violations",
		"| SKILL-006 | Pilot one approved profile/project overlay refresh | Completed |",
		"single project overlay layout",
		"legacy archive inactive/readback then removed",
	})
}

func TestSKILL003WrapperTemplatesAndGuidesAreThinAndConcrete(t *testing.T) {
	wrappers := []string{
		"templates/profile-wrappers/kkachi-blue-wrapper/SKILL.md.tmpl",
		"templates/profile-wrappers/kkachi-red-wrapper/SKILL.md.tmpl",
		"templates/profile-wrappers/kkachi-orange-wrapper/SKILL.md.tmpl",
		"templates/profile-wrappers/kkachi-gray-wrapper/SKILL.md.tmpl",
	}
	for _, wrapper := range wrappers {
		requireContainsAll(t, wrapper, []string{
			"kind: project_wrapper",
			"role_manifest: kkachi-agent-skills:roles/",
			"plugin_namespace: kkachi-agent-skills",
			"overlay_skill:",
			"refresh_skill: kkachi-agent-skills:kkachi-agent-skills-overlay-refresh",
			`skill_view("kkachi-agent-skills:<base>")`,
			`skill_view("kkachi-agent-skills:kkachi-agent-skills-overlay-refresh")`,
			"LLM-assisted semantic porting, not CLI migration",
			"Allowed base subset:",
			"stop, report",
			"do not use profile-local copied skills as fallback",
			"This wrapper does not authorize profile cleanup",
		})
		content := readRepoFile(t, wrapper)
		if strings.Count(content, "# ") > 1 {
			t.Fatalf("%s looks like a copied base body, heading count too high", wrapper)
		}
		if len(content) > 2400 {
			t.Fatalf("%s is too large for a thin wrapper template: %d bytes", wrapper, len(content))
		}
	}

	requireContainsAll(t, "skills/kas-project-overlay-guide/SKILL.md", []string{
		"skills/<project>/<project>-overlay/SKILL.md",
		"skills/<project>/<project>-overlay/references/*.md",
		"applies_to:",
		"kkachi-agent-skills:plan",
		"kkachi-plan",
		"Stop and request review",
	})
	requireContainsAll(t, "skills/kas-overlay-compose-guide/SKILL.md", []string{
		"current context:",
		"project overlay:",
		"wrapper:",
		"plugin base:",
		`skill_view("kkachi-agent-skills:plan")`,
		"Conflict Handling",
	})
	requireContainsAll(t, "skills/kas-overlay-doctor-guide/SKILL.md", []string{
		"non-executing guidance only",
		"SKILL-004 owns the",
		"implemented read-only doctor diagnostics",
		"required plugin base skill missing",
		"applies_to contains `kkachi-plan`",
		"profile-local full KAS base copy present",
	})
	requireContainsAll(t, "skills/kkachi-agent-skills-overlay-refresh/SKILL.md", []string{
		"name: kkachi-agent-skills-overlay-refresh",
		"<project>-overlay-legacy",
		"kind: project_overlay_legacy",
		"active: false",
		"remove `<project>-overlay-legacy`",
		"not a CLI migration",
	})
}

func TestSKILL003GuidePackageConventionAndReadbackContract(t *testing.T) {
	requireContainsAll(t, "skill-pack.yaml", []string{
		"guides:",
		"kas-project-overlay-guide",
		"kas-overlay-compose-guide",
		"kas-overlay-doctor-guide",
		"kkachi-agent-skills-overlay-refresh",
	})
	requireContainsAll(t, skillPluginWrapperOverlaySOT, []string{
		"source package convention stores guide bodies at `skills/<guide-id>/SKILL.md`",
		"`guides:` manifest metadata maps those source skill directories into the plugin",
		"guide readback surface",
	})
	requireContainsAll(t, "internal/skills/discovery/plugin.go", []string{
		"official_plugin_guide",
		"BuildSourceGuideSkillReadback",
	})
	requireContainsAll(t, "internal/skills/pluginupdate/pluginupdate.go", []string{
		"GuideSkillImpact",
		"SourceClass",
	})
}

func TestSKILL004DoctorCommandContract(t *testing.T) {
	requireContainsAll(t, skillPluginWrapperOverlaySOT, []string{
		"KAS doctor has a read-only SKILL mode exposed as `kkachi-agent-skills doctor --plugin --repo <kas-repo> --json`",
		"diagnostic-only",
		"`--repo` is required for plugin doctor mode",
		"`provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`",
		"does not authorize profile mutation, KAB activation, provider/model configuration, deleted-bundle fallback lookup, or install readiness by assumption",
		"without writing",
		"incompatible with `--workflow-graph` and `--project-suite`",
		"mixed doctor modes fail closed with `doctor_mode_ambiguous`",
		"`plugin_base`, `project_wrapper`, `project_overlay`, `legacy_copied_base_suite`, `personal_skill`, and `unknown_source`",
		"profile-local copied suites are never a fallback",
	})
	requireContainsAll(t, "internal/skills/doctor/skill.go", []string{
		"skill_plugin_overlay_doctor",
		"legacy_copied_base_suite",
		"missing_plugin_base_skill",
		"Do not fall back to copied profile-local KAS suites",
	})
	requireContainsAll(t, "internal/skills/cli/cli.go", []string{
		"doctor --plugin cannot be combined with --workflow-graph or --project-suite",
		"repo_required_for_plugin_doctor",
	})
}

func TestSKILL005RemovedUpdateMigrateCommandContract(t *testing.T) {
	requireContainsAll(t, skillPluginWrapperOverlaySOT, []string{
		"The `kkachi-agent-skills` CLI must not expose public `update` or `migrate` commands",
		"semantic overlay refresh belongs in the `kkachi-agent-skills-overlay-refresh` skill",
		"Legacy material is handled through review, doctor evidence",
		"This SOT does not authorize broad copied-suite cleanup, automatic deletion, CLI migration",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"| SKILL-005 | Remove update/migrate CLI and keep legacy handling review-only | In Review |",
		"Public `update`, `migrate-profile-skills`, and `migrate-project-kas` CLI surfaces are removed",
		"Do not mark Completed until focused tests, aggregate feasible tests, `git diff --check`, KAH gates, color review/MAR as required, and final KAH gate pass.",
	})
	requireContainsAll(t, "internal/skills/cli/cli.go", []string{
		"only the list, install, doctor, repair, toolchain",
		"sync-project-kas, install-project-kas, repair-project-kas",
	})
	cli := readRepoFile(t, "internal/skills/cli/cli.go")
	for _, forbidden := range []string{
		"case \"update\"",
		"case \"migrate-profile-skills\"",
		"case \"migrate-project-kas\"",
		"runMigrateProfileSkills",
	} {
		if strings.Contains(cli, forbidden) {
			t.Fatalf("removed CLI surface still present: %s", forbidden)
		}
	}
}
