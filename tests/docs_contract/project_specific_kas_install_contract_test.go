package docscontract

import (
	"strings"
	"testing"
)

const projectSpecificKASSOT = "docs/sot/project-specific-kas-install-contract.md"

func TestProjectSpecificKASInstallSOTDefinesCanonicalLayoutAndExamples(t *testing.T) {
	requireContainsAll(t, projectSpecificKASSOT, []string{
		"Status: canonical SOT for `KASPROJ-001`; `KASPROJ-002` read-only dry-run planner is implemented/pre-commit-ready; `KASPROJ-003` approved install is implemented/in-review pending gates",
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
		"kas_profile_skill_manifest",
		"kas_project_skill_manifest",
		"project_suites",
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
		"KASPROJ-002 implements the read-only dry-run planner surface",
		"KASPROJ-003 approved install surface",
		"KASPROJ-004` doctor/repair/migrate",
		"dry-run evidence, require explicit approval before profile mutation",
		"never mutate auth, tokens, secrets, gateway, provider/model config, KAH state, or KAB runtime state",
		"install-project-kas --profile <profile> --project <project> --source-pack kas-default-project-suite --dry-run",
		"install-project-kas --profile <profile> --project <project> --source-pack kas-default-project-suite --approve dry-run:<plan_hash>",
		"formal source suite `kas-default-project-suite`",
		"prefix-render only",
		"semantic_adaptation_claimed:false",
		"drift_policy: manual_review_required",
		"approval evidence to the recomputed `plan_hash`",
		"backup/recovery evidence",
		"project doctor/repair remains KASPROJ-004",
		"JSON output is `ok:false` and the CLI exits 2",
		"\"condition\": \"umbrella_only\"",
		"\"plan_hash\": \"sha256:<hex>\"",
	})
}

func TestProjectSpecificKASInstallDocsRegistrationAndRoadmap(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"sot/project-specific-kas-install-contract.md",
		"KASPROJ-001 project-specific KAS install layout SOT",
		"requires `skills/<project>/<project>-<phase-or-skill>/SKILL.md`",
		"generic duplicate skill names such as `kkachi-plan` are invalid",
		"umbrella-only installs are incomplete",
		"KASPROJ-002 read-only dry-run planning and KASPROJ-003 approval-hash-bound install are implemented/in-review",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"docs/sot/project-specific-kas-install-contract.md",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"### EPIC: KASPROJ",
		"| KASPROJ-001 | Specify project-specific KAS install layout SOT | In Review |",
		"KASPROJ-002 | Implement project-specific dry-run planner | In Review",
		"KASPROJ-003 | Implement approved project-specific install | In Review",
		"KASPROJ-004 | Implement doctor/repair/migrate for project suites | Planned",
		"KASPROJ-005 | Apply project-specific KAS install to one approved operational profile/project set | Planned",
		"In review pending post-implementation gates",
	})
	requireRoadmapTaskStatus(t, "KASPROJ-001", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-002", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-003", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-004", "Planned")
	requireRoadmapTaskStatus(t, "KASPROJ-005", "Planned")
}

func TestProjectSpecificKASCLIContractDefinesApprovedInstallWithoutRolloutClaim(t *testing.T) {
	requireContainsAll(t, "docs/sot/kas-cli-contract.md", []string{
		"docs/sot/project-specific-kas-install-contract.md",
		"Project-specific KAS install vocabulary",
		"skills/<project>/<project>-<phase-or-skill>/SKILL.md",
		"Installed skill",
		"Target path",
		"Drift policy",
		"The current CLI implements KASPROJ-002 read-only project-suite dry-run planning and KASPROJ-003 approval-hash-bound project-suite install",
		"must not be described as performing project-suite doctor, repair, migration, semantic-port completion, operational rollout, or KAB runtime activation",
		"install-project-kas --profile <profile> --project <project> --source-pack <source_pack> --dry-run",
		"install-project-kas --profile <profile> --project <project> --source-pack kas-default-project-suite --approve dry-run:<hash>",
		"`kas-default-project-suite`, resolved from repository `skill-pack.yaml`",
		"Missing both `--dry-run`/`--approve`, using both together, malformed `--approve`",
		"`plan_hash` binds CLI version, target profile/root, manifest path/state/previous manifest sha",
		"backup/recovery evidence",
		"semantic_adaptation_claimed:false",
		"\"condition\": \"generic_installed_skill_name\"",
		"repair-project-kas --profile <profile> --project <project> --dry-run",
		"migrate-project-kas --profile <profile> --project <project> --from-generic --dry-run",
		"umbrella-only installs are incomplete/invalid",
		"duplicate generic installed skill names such as `kkachi-plan` are invalid",
	})
}
