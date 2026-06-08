package docscontract

import (
	"strings"
	"testing"
)

const projectSpecificKASSOT = "docs/sot/project-specific-kas-install-contract.md"

func TestProjectSpecificKASInstallSOTDefinesCanonicalLayoutAndExamples(t *testing.T) {
	requireContainsAll(t, projectSpecificKASSOT, []string{
		"Status: canonical SOT for `KASPROJ-001`",
		"~/.hermes/profiles/<profile>/skills/<project>/<project>-<phase-or-skill>/SKILL.md",
		"skills/<project>/<project>-<phase-or-skill>/SKILL.md",
		"/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-server/doksuri-server-plan/SKILL.md",
		"/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-daemon/doksuri-daemon-plan/SKILL.md",
		"/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-client/doksuri-client-plan/SKILL.md",
		"/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-cli/doksuri-cli-plan/SKILL.md",
		"project's language, runtime, repository layout, test commands, docs map",
	})
}

func TestProjectSpecificKASInstallSOTRequiresProjectPrefixedNames(t *testing.T) {
	requireContainsAll(t, projectSpecificKASSOT, []string{
		"Project-prefix skill names are mandatory.",
		"doksuri-server-plan",
		"doksuri-daemon-plan",
		"doksuri-client-implement",
		"doksuri-cli-final-verify",
		"kkachi-plan",
		"Duplicate generic names are invalid in one Hermes profile",
		"stable unambiguous `installed_skill` id",
	})

	content := readRepoFile(t, projectSpecificKASSOT)
	forbidden := []string{
		"generic names are allowed for project suites",
		"kkachi-plan is valid for all project suites",
		"fallback to global generic skills",
	}
	for _, phrase := range forbidden {
		if strings.Contains(content, phrase) {
			t.Fatalf("project-specific install SOT contains forbidden generic/fallback claim %q", phrase)
		}
	}
}

func TestProjectSpecificKASInstallSOTPreservesUmbrellaAndManifestVocabulary(t *testing.T) {
	requireContainsAll(t, projectSpecificKASSOT, []string{
		"umbrella-only installation is incomplete and invalid",
		"skills/<project>/<project>-kas/SKILL.md",
		"must be reported as `umbrella_only`, not healthy",
		"project",
		"source_pack",
		"installed_skill",
		"target_path",
		"drift_policy",
		"kas_project_skill_manifest",
		"skills/doksuri-server/doksuri-server-plan/SKILL.md",
	})
}

func TestProjectSpecificKASInstallSOTDefinesDoctorSeveritiesAndApprovalBoundary(t *testing.T) {
	requireContainsAll(t, projectSpecificKASSOT, []string{
		"missing project suite | `error`",
		"umbrella-only | `error`",
		"missing file | `error`",
		"checksum mismatch | `error`",
		"unknown profile skill dir | `warning`",
		"profile/source language drift | `warning`",
		"KASPROJ-002` dry-run planner",
		"KASPROJ-003` approved install",
		"KASPROJ-004` doctor/repair/migrate",
		"dry-run evidence, require explicit approval before profile mutation",
		"never mutate auth, tokens, secrets, gateway, provider/model config, KAH state, or KAB runtime state",
		"must not be marked `Completed`",
	})
}

func TestProjectSpecificKASInstallDocsRegistrationAndRoadmap(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"sot/project-specific-kas-install-contract.md",
		"KASPROJ-001 project-specific KAS install layout SOT",
		"requires `skills/<project>/<project>-<phase-or-skill>/SKILL.md`",
		"generic duplicate skill names such as `kkachi-plan` are invalid",
		"umbrella-only installs are incomplete",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"docs/sot/project-specific-kas-install-contract.md",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"### EPIC: KASPROJ",
		"| KASPROJ-001 | Specify project-specific KAS install layout SOT | In Review |",
		"KASPROJ-002 | Implement project-specific dry-run planner | Planned",
		"KASPROJ-003 | Implement approved project-specific install | Planned",
		"KASPROJ-004 | Implement doctor/repair/migrate for project suites | Planned",
		"KASPROJ-005 | Apply project-specific KAS install to one approved operational profile/project set | Planned",
		"Do not mark Completed until the docs/spec PR receives the responsible review/commit gate.",
	})
	requireRoadmapTaskStatus(t, "KASPROJ-001", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-002", "Planned")
	requireRoadmapTaskStatus(t, "KASPROJ-003", "Planned")
	requireRoadmapTaskStatus(t, "KASPROJ-004", "Planned")
	requireRoadmapTaskStatus(t, "KASPROJ-005", "Planned")
}

func TestProjectSpecificKASCLIContractExtendsVocabularyWithoutImplementationClaim(t *testing.T) {
	requireContainsAll(t, "docs/sot/kas-cli-contract.md", []string{
		"docs/sot/project-specific-kas-install-contract.md",
		"Project-specific KAS install vocabulary",
		"skills/<project>/<project>-<phase-or-skill>/SKILL.md",
		"Installed skill",
		"Target path",
		"Drift policy",
		"must not be described as already installing, repairing, or migrating project-specific suites",
		"install-project-kas --profile <profile> --project <project> --source-pack <source_pack> --dry-run",
		"repair-project-kas --profile <profile> --project <project> --dry-run",
		"migrate-project-kas --profile <profile> --project <project> --from-generic --dry-run",
		"umbrella-only installs are incomplete/invalid",
		"duplicate generic installed skill names such as `kkachi-plan` are invalid",
	})
}
