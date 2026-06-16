package docscontract

import "testing"

const roleAwareProjectSuiteSOT = "docs/sot/role-aware-project-suite-contract.md"

func TestRoleAwareProjectSuiteSOTDefinesRolesAndRegistry(t *testing.T) {
	requireContainsAll(t, roleAwareProjectSuiteSOT, []string{
		"Status: KASROLE-001..004 source behavior completed through commits",
		"Red `t_bfad026c`, Orange `t_21f593c9`, Gray `t_acc68322`, Blue synthesis `t_a971aa92`",
		"`blue_commander`",
		"`red_reviewer`",
		"`orange_pm_reviewer`",
		"`gray_scribe`",
		"registries/project-suite-roles.yaml",
		"Unknown, misspelled, or unregistered `suite_role` values must fail closed",
		`display_label: "Blue commander / full project suite"`,
		`display_label: "Red safety/fail-closed reviewer subset"`,
		`display_label: "Orange operator-value reviewer subset"`,
		`display_label: "Gray evidence/scribe reviewer subset"`,
		"forbidden_source_skills",
		"kkachi-docs-update",
	})
}

func TestRoleAwareProjectSuiteSOTDefinesCLIManifestDoctorAndRepairBoundaries(t *testing.T) {
	requireContainsAll(t, roleAwareProjectSuiteSOT, []string{
		"`--suite-role` is mandatory for role-aware project suite install/repair",
		"KAS must not silently default Red/Orange/Gray profiles to a full source suite",
		"selected `suite_role` plus operator-readable `display_label`",
		"selected/excluded skill counts in the compact human summary",
		`"suite_mode": "role_subset"`,
		`"suite_role": "red_reviewer"`,
		"unknown or unregistered `suite_role` requested or recorded | error",
		"legacy full suite in Red/Orange/Gray profile | error or blocked diagnostic",
		"no-spillover scan evidence",
		"Apply requires explicit 주군 operational approval for the exact target profiles/projects",
		"never mutate profile config, SOUL, gateway, auth, tokens, providers, models, KAH state, or KAB runtime state",
	})
}

func TestRoleAwareProjectSuiteDocsRegistrationAndRoadmap(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"sot/role-aware-project-suite-contract.md",
		"canonical role registry path `registries/project-suite-roles.yaml`",
		"unknown-role fail-closed behavior",
		"KASROLE-001..004 source work is implemented and committed through `7324d9c`",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"docs/sot/role-aware-project-suite-contract.md",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"### EPIC: KASROLE",
		"| KASROLE-001 | Accept role-aware project suite SOT and registry plan | Completed |",
		"Initial color evidence: Red `t_bfad026c`, Orange `t_21f593c9`, Gray `t_acc68322`, Blue synthesis `t_a971aa92`",
		"unknown-role fail-closed fixtures",
		"missing/unknown `suite_role`",
		"KASROLE is the KAS v0.1.3 release baseline",
		"WFLOW epic completion is the KAS v0.1.4 release baseline",
		"WFLOW startup must still run effective installed-binary/profile doctor evidence",
	})
	requireRoadmapTaskStatus(t, "KASROLE-001", "Completed")
	requireRoadmapTaskStatus(t, "KASROLE-002", "Completed")
	requireRoadmapTaskStatus(t, "KASROLE-003", "Completed")
	requireRoadmapTaskStatus(t, "KASROLE-004", "Completed")
}
