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
		"project tailoring checksum drift | `warning`",
		"project_tailoring_checksum_drift",
		"tailoring_mode: profile_local_repo_semantic_tailoring",
		"unknown profile skill dir | `warning`",
		"profile/source language drift | `warning`",
		"KASPROJ-002 originally introduced the read-only dry-run planner surface",
		"KASPROJ-003 approved install surface",
		"KASPROJ-004` doctor/repair/migrate",
		"dry-run evidence, require explicit approval before profile mutation",
		"never mutate auth, tokens, secrets, gateway, provider/model config, KAH state, or KAB runtime state",
		"install-project-kas --profile <profile> --project <project> --suite-role blue_commander --source-pack kas-default-project-suite --dry-run",
		"install-project-kas --profile <profile> --project <project> --suite-role blue_commander --source-pack kas-default-project-suite --approve dry-run:<plan_hash>",
		"install --profile <profile> --project <project> --suite-role <role> --apply dry-run:sha256:<hash>",
		"formal source suite `kas-default-project-suite`",
		"prefix-render only",
		"semantic_adaptation_claimed:false",
		"drift_policy: manual_review_required",
		"approval evidence to the recomputed `plan_hash`",
		"backup/recovery evidence",
		"project-suite doctor should be rerun",
		"Repair and migration default `source_pack` to `kas-default-project-suite`",
		"manual_semantic_port_tasks[]",
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
		"KASPROJ-002 read-only dry-run planning, KASPROJ-003 approval-hash-bound install, KASPROJ-004 doctor/repair/migrate, and KASPROJ-005 project-tailored doctor policy are implemented/in-review",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"docs/sot/project-specific-kas-install-contract.md",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"### EPIC: KASPROJ",
		"| KASPROJ-001 | Specify project-specific KAS install layout SOT | In Review |",
		"KASPROJ-002 | Implement project-specific dry-run planner | In Review",
		"KASPROJ-003 | Implement approved project-specific install | In Review",
		"KASPROJ-004 | Implement doctor/repair/migrate for project suites | In Review",
		"KASPROJ-005 | Fix project-tailored doctor checksum policy | In Review",
		"KASPROJ-006 | Apply project-specific KAS install to one approved operational profile/project set | Planned",
		"In review pending post-implementation gates",
	})
	requireRoadmapTaskStatus(t, "KASPROJ-001", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-002", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-003", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-004", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-005", "In Review")
	requireRoadmapTaskStatus(t, "KASPROJ-006", "Planned")
}

func TestProjectSpecificKASCLIContractDefinesApprovedInstallWithoutRolloutClaim(t *testing.T) {
	requireContainsAll(t, "docs/sot/kas-cli-contract.md", []string{
		"docs/sot/project-specific-kas-install-contract.md",
		"Project-specific KAS install vocabulary",
		"skills/<project>/<project>-<phase-or-skill>/SKILL.md",
		"Installed skill",
		"Target path",
		"Drift policy",
		"The current CLI implements KASPROJ-002 read-only project-suite dry-run planning, KASPROJ-003 approval-hash-bound project-suite install, and KASPROJ-004 project-suite doctor/repair/migrate",
		"must not be described as performing semantic-port completion, operational rollout, generic fallback, or KAB runtime activation",
		"install-project-kas --profile <profile> --project <project> --suite-role <role> --source-pack <source_pack> --dry-run",
		"install-project-kas --profile <profile> --project <project> --suite-role <role> --source-pack kas-default-project-suite --approve dry-run:<hash>",
		"`kas-default-project-suite`, resolved from repository `skill-pack.yaml`",
		"Missing both `--dry-run`/`--approve`, using both together, malformed `--approve`",
		"`plan_hash` binds CLI version, target profile/root, manifest path/state/previous manifest sha",
		"backup/recovery evidence",
		"semantic_adaptation_claimed:false",
		"\"condition\": \"generic_installed_skill_name\"",
		"doctor --profile <profile> --project <project> --project-suite",
		"repair-project-kas --profile <profile> --project <project> --dry-run",
		"repair-project-kas --profile <profile> --project <project> --approve dry-run:<hash>",
		"migrate-project-kas --profile <profile> --project <project> --from-generic --dry-run",
		"migrate-project-kas --profile <profile> --project <project> --from-generic --approve dry-run:<hash>",
		"update --dry-run` is the public project KAS sync command",
		"roles/profiles, source packs, skill ids, target paths, checksums",
		"`sync-project-kas`",
		"compatibility command",
		"install --profile <profile> --project <project> --dry-run",
		"install --project --apply",
		"install --profile <profile> --project <project> --from-generic --dry-run",
		"install --from-generic --apply",
		"repair --profile <profile> --project <project> --dry-run",
		"repair --apply",
		"uninstall --profile <profile> --project <project> --dry-run",
		"uninstall --apply",
		"these public project forms were read-only",
		"Public write/apply forms failed closed before TOKEN-005",
		"`uninstall --dry-run` is planner-only in TOKEN-004",
		"reports manifest-tracked planned removals",
		"skipped local-only or unmanifested files",
		"Removal and backup/evidence writing are TOKEN-005 behavior",
		"TOKEN-005 public lifecycle write forms use `--apply dry-run:sha256:<hash>`",
		"`update --apply`",
		"writes only hash-bound auto-copy candidates",
		"`--backup-vault-root <abs-path>`",
		"profile manifest is updated last",
		"Repair and migration default `source_pack` to `kas-default-project-suite`",
		"KASPROJ-005 refines doctor semantics so project-local semantic tailoring is a warning condition",
		"project_tailoring_checksum_drift",
		"umbrella-only installs are incomplete/invalid",
		"duplicate generic installed skill names such as `kkachi-plan` are invalid",
	})
}
