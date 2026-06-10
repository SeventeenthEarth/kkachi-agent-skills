package docscontract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const tokenEconomyAgentInstructionSOT = "docs/sot/token-economy-and-agent-instruction-contract.md"

var token003AgentInstructionTemplates = []string{
	"templates/agent-instructions/AGENTS.md.tmpl",
	"templates/agent-instructions/CLAUDE.md.tmpl",
}

const token003DryRunExamples = "docs/examples/agent-instruction-dry-run-merge-plans.md"

func TestTokenEconomyAgentInstructionSOTDefinesCoreContract(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"KAS PR 1: token-economy, English product-output, and lifecycle UX SOT/docs-contract",
		"KAS PR 2: English compact artifact-first backend prompt templates and phase guidance",
		"KAS PR 3: English `AGENTS.md` / `CLAUDE.md` template and management workflow",
		"KAS PR 4: project KAS lifecycle UX and read-only planner surface",
		"KAS PR 5: approved lifecycle writes, update apply, and uninstall with vault backup",
		"KAS PR 6: skill slimming and reference split for high-token KAS guidance surfaces",
		"KAH PR 1: mechanical token-economy / English-output / project KAS lifecycle evidence gates",
		"The same KAS contract must apply to direct Codex app-server lanes and KAB-mediated backend lanes",
		"KAB must stay a connection interface",
		"KAS owns prompt/policy semantics",
		"KAH validates only mechanically checkable evidence",
	})
}

func TestTokenEconomyAgentInstructionSOTRequiresEnglishKASSurfaces(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"All KAS-generated prompt templates, backend prompts, CLI help text, human CLI output, console summaries, report schemas, and artifact templates must be English by default.",
		"KAS must not generate Korean prose in prompts or console output.",
		"This language rule applies to direct Codex app-server lanes and KAB-mediated backend lanes.",
		"Chat reports from a Hermes team member to 주군 are outside this product-output contract.",
	})
}

func TestTokenEconomyAgentInstructionSOTRequiresCompactArtifactFirstOutput(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"English compact console output contract",
		"Status: <pass|fail|blocked|in_progress|not_applicable>",
		"Summary: <1-5 bullets>",
		"Files: <changed/inspected paths only>",
		"Verification: <commands/checks and short result>",
		"Risks/blockers: <only actionable remaining issues>",
		"Detailed artifact: <path or not_applicable reason>",
		"Next action requested: <approval/review/none>",
		"Do not paste long plans, full diffs, full logs, full file contents, long reviews, or exhaustive checklist text into the console",
		"Artifact-first detail policy",
		".kkachi/runs/<run_id>/artifacts/<phase>/backend-<phase>.md",
		"A backend that cannot write or update the requested artifact must report that blocker instead of dumping the full detail into chat.",
	})
}

func TestTokenEconomyAgentInstructionSOTDefinesTaskClassesAndAgentFiles(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"Task class gating",
		"`simple_report`",
		"`investigation`",
		"`docs_only`",
		"`development`",
		"`review`",
		"`epic_closure`",
		"`AGENTS.md`: generic repo-local agent instruction",
		"`CLAUDE.md`: Claude-compatible repo-local instruction",
		"forrestchang/andrej-karpathy-skills",
		"think before coding",
		"simplicity first",
		"surgical changes",
		"goal-driven execution and verification",
		"<!-- KAS:MANAGED:BEGIN core-behavior -->",
		"<!-- PROJECT:LOCAL:BEGIN -->",
		"dry-run merge plan and require approval before rewriting or inserting managed blocks",
	})
}

func TestTokenEconomyAgentInstructionSOTDefinesColorTeamLifecycleContract(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"Color-team project KAS lifecycle contract",
		"Project-scoped KAS lifecycle support must not be blue-only",
		"`blue`, `red`, `orange`, and `gray` role profiles",
		"This is orchestration authority only",
		"The public operator-facing project KAS lifecycle verbs are:",
		"`install`: create a project-scoped KAS suite",
		"`update`: compare the installed suite with the current upstream source",
		"`doctor`: inspect installed state without writing",
		"`repair`: restore missing or damaged KAS-managed project-suite files",
		"`uninstall`: remove KAS-managed project-suite files with manifest-bound backup and evidence",
		"Operators should not need to choose between `sync`, `migrate`, and approved write modes for normal use.",
	})
}

func TestTokenEconomyAgentInstructionSOTDefinesAdvancedLifecycleAndApplyGate(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"`sync` is the read-only planner/classifier behind `update --dry-run`",
		"`migrate` is a one-time compatibility path",
		"`install --from-generic`",
		"Approved writes are not a separate operator concept.",
		"Dry-run and hash-bound apply gate",
		"planned state: `create`, `update`, `remove`, `no_change`, `conflict`, `blocked`, or `error`",
		"<command> --apply dry-run:sha256:<hash>",
		"`--apply` is the preferred operator-facing spelling",
		"`--approve dry-run:sha256:<hash>` behavior may remain as a compatibility alias",
		"separate `approved sync/write` concept",
	})
}

func TestTokenEconomyAgentInstructionSOTDefinesUninstallVaultBackupPolicy(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"Uninstall and backup policy",
		"`uninstall` must remove only manifest-tracked KAS-managed project-suite artifacts by default.",
		"It must not remove local-only files, unmanifested project instructions, credentials, profile configuration, gateway/model/provider/auth settings, or runtime state.",
		"approved long-lived Obsidian vault backup area",
		"not only inside the active Hermes profile",
		"files skipped because they are local-only or unmanifested",
		"exact apply command",
	})
}

func TestTokenEconomyAgentInstructionSOTPreservesLayerAndMutationBoundaries(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"Hermes runtime fork/patch work",
		"KAB policy, fallback, or selection-judgment changes",
		"auth, token, gateway, provider, or model configuration mutation",
		"unapproved profile skill installation or operational rollout",
		"removal of existing project-local agent instructions, local-only files, or unmanifested project-suite files without approval",
		"KAH must expose only mechanically checkable results: `pass`, `fail`, or `not_applicable`",
		"KAH must not judge prose quality or rewrite instructions",
	})

	content := readRepoFile(t, tokenEconomyAgentInstructionSOT)
	forbidden := []string{
		"Hermes runtime fork is required",
		"KAB decides token-economy policy",
		"KAH should summarize the output quality",
		"warnings are sufficient for KAH gates",
		"unapproved profile mutation is authorized by this document",
		"auth/token changes are authorized by this document",
		"KAS may generate Korean prose in prompts",
		"operators must choose between sync and update",
	}
	for _, phrase := range forbidden {
		if strings.Contains(content, phrase) {
			t.Fatalf("%s contains forbidden token-economy boundary claim %q", tokenEconomyAgentInstructionSOT, phrase)
		}
	}
}

func TestTokenEconomyAgentInstructionDocsRegistrationAndRoadmap(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"sot/token-economy-and-agent-instruction-contract.md",
		"Accepted token-economy and agent-instruction SOT",
		"selected 10 KAS PR + 2 dependent KAH PR workstream",
		"English KAS-generated prompt/CLI/console/artifact-template output",
		"compact console output",
		"artifact-first details",
		"`AGENTS.md` / `CLAUDE.md` management",
		"color-team project KAS lifecycle gates",
		"uninstall/vault-backup policy",
		"no implementation, unapproved profile mutation, KAB activation, or Hermes runtime fork is authorized by the SOT alone",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"docs/sot/token-economy-and-agent-instruction-contract.md",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"sot/token-economy-and-agent-instruction-contract.md",
		"EPIC: TOKEN — KAS token economy and English operator surfaces",
		"TOKEN-001 | Accept token-economy, English output, and lifecycle UX SOT",
		"10 KAS PRs plus 2 dependent KAH PR",
		"English KAS-generated prompt/CLI/console/artifact-template output",
		"does not authorize implementation, unapproved profile mutation, KAB activation, Hermes runtime changes, or auth/token/provider/gateway/model mutation by itself",
	})
}

func TestToken003AgentInstructionTemplatesCarryManagedEnglishPolicy(t *testing.T) {
	for _, rel := range token003AgentInstructionTemplates {
		requireContainsAll(t, rel, []string{
			"<!-- KAS:MANAGED:BEGIN core-behavior -->",
			"<!-- KAS:MANAGED:END core-behavior -->",
			"<!-- PROJECT:LOCAL:BEGIN -->",
			"<!-- PROJECT:LOCAL:END -->",
			"{{project_name}}",
			"{{repository_role}}",
			"{{project_suite_id}}",
			"{{kab_adoption_stage}}",
			"{{upstream_kas_baseline}}",
			"{{local_authority_notes}}",
			"KAS",
			"KAH",
			"KAB",
			"English",
			"artifact",
			"auth",
			"token",
			"provider/model/gateway",
		})
	}

	requireContainsAll(t, "templates/agent-instructions/AGENTS.md.tmpl", []string{
		"Think before coding",
		"Simplicity first",
		"Surgical changes",
		"Goal-driven execution",
		"Status: pass, fail, blocked, in_progress, or not_applicable.",
	})
	requireContainsAll(t, "templates/agent-instructions/CLAUDE.md.tmpl", []string{
		"Think before coding.",
		"Keep the implementation simple.",
		"Make surgical changes.",
		"Work toward verified goals.",
	})
}

func TestToken003DryRunExamplesCoverNoWriteMergeCases(t *testing.T) {
	requireContainsAll(t, token003DryRunExamples, []string{
		"Existing Managed File",
		"Existing Unmarked AGENTS.md",
		"Missing CLAUDE.md",
		"Malformed Or Conflicting Markers",
		"mode: dry_run_only",
		"target_file_writes: []",
		"planned_state: update",
		"planned_state: blocked",
		"planned_state: not_applicable",
		"planned_state: conflict",
		"preserve_project_local_block",
		"preserve_entire_existing_file",
		"unmarked_existing_instruction_file",
		"missing_optional_claude_instruction",
		"malformed_or_conflicting_markers",
		"TOKEN-003 must not blind-overwrite or insert managed blocks into an",
		"TOKEN-003",
		"must not create root `CLAUDE.md`.",
	})

	content := readRepoFile(t, token003DryRunExamples)
	if count := strings.Count(content, "target_file_writes: []"); count != 4 {
		t.Fatalf("%s should keep every documented merge case no-write; got %d target_file_writes entries", token003DryRunExamples, count)
	}
}

func TestToken003DryRunSurfacesRejectWriteCapableWording(t *testing.T) {
	relPaths := append([]string{token003DryRunExamples}, token003AgentInstructionTemplates...)
	for _, rel := range relPaths {
		requireContainsNone(t, rel, []string{
			"--apply",
			"--approve",
			"apply dry-run:sha256",
			"approve dry-run:",
			"may write target files",
			"will write target files",
			"rewrite root `AGENTS.md` now",
			"create root `CLAUDE.md` now",
			"blind-overwrite is allowed",
			"profile mutation is authorized",
			"auth/token/provider/gateway/model mutation is authorized",
		}, "contains TOKEN-003 forbidden write-capable wording")
	}
}

func TestTokenRoadmapSeparatesCompletedToken004FromToken005Writes(t *testing.T) {
	requireRoadmapTaskStatus(t, "TOKEN-003", "Completed")
	requireRoadmapTaskStatus(t, "TOKEN-004", "Completed")
	requireRoadmapTaskStatus(t, "TOKEN-005", "Completed")
	requireRoadmapTaskStatus(t, "TOKEN-006", "Completed")
	requireRoadmapTaskStatus(t, "TOKEN-007", "Completed")
	requireContainsAll(t, "docs/roadmap.md", []string{
		"TOKEN-003 | Implement English repo-local agent instruction templates | Completed",
		"`AGENTS.md` and `CLAUDE.md` templates use English managed blocks, preserve project-local content, and encode KAS/KAH/KAB boundaries without blind overwrite.",
		"Verified dry-run examples, managed-marker tests, docs-contract, repo test gate, GLM Octo review, post-Octo second color review, and KAH final gate `evt-001886` in run `run-20260609T134803Z-fe1b071fd6d9`.",
		"TOKEN-004 | Implement public project KAS lifecycle UX and read-only planner | Completed",
		"Completed in KAH run `run-20260609T150446Z-ffd5705a418e`",
		"Codex implementation/fixes, enhance-test, AI slop cleanup, optimize, and docs-update evidence passed",
		"Official KAB GLM Octo session `872cd977-7b23-4b48-bbcc-886f2cf833b3` accepted with 0 blockers",
		"Post-fix second color re-review accepted: Red `t_61d7a108`, Orange `t_3c95fbed`, Gray `t_817322f8`",
		"KAH final verification gate passed as `evt-002022`",
		"TOKEN-005 | Implement approved lifecycle writes and uninstall vault backup | Completed",
		"Completed in KAH run `run-20260609T180803Z-8a90f04f1ab3`",
		"Codex implementation/fix stages, enhance-test, AI slop cleanup, optimize, docs, verification, `git diff --check`, targeted Go tests, and `make test` passed.",
		"Official KAB GLM Octo session `6f72523c-a0ef-41c4-9ec0-228dd1de8831` accepted as `GLM_OCTO_PASS` with 0 blockers.",
		"Second color review accepted: Red `t_093081cf`, Orange `t_3fc427aa`, Gray `t_a59bacce`.",
		"KAH review gate passed as `evt-002187`; KAH final gate remains separate/not yet known.",
		"TOKEN-006 | Slim high-frequency KAS skills and split references | Completed",
		"run `run-20260610T014133Z-ef48803be3de`",
		"direct Codex app-server implementation slimmed six high-frequency `SKILL.md` files by 20,578 bytes total",
		"Linked-reference docs-contract coverage and `make test` passed.",
		"First color review accepted: Red `t_d1aa939f`, Orange `t_33620d98`, Gray `t_5da311bc`.",
		"Official KAB GLM Octo session `0b104960-3729-43b5-b834-9bef3ded5745` accepted as `GLM_OCTO_PASS` with 0 blockers and required Codex rework `no`.",
		"Second color review accepted: Red `t_bf2a5dfd`, Orange `t_ec226e53`, Gray `t_1059dace`.",
		"Verification/review gates passed; KAH final gate remains separate/not yet known.",
	})
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"KAS PR 3: English agent instruction templates and management workflow",
		"CLI or documented dry-run workflow reports create/update/no-change/not_applicable posture without blind overwrite.",
		"KAS PR 4: project KAS lifecycle UX and read-only planner surface",
		"KAS PR 5: approved lifecycle writes, update apply, and uninstall with vault backup",
	})
}

func TestToken007SOTDefinesVerificationProfileAndRunnerEvidence(t *testing.T) {
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"KAS PR 7: project verification profiles and no-agent command runners",
		"avoids a global `make test` assumption",
		"aggregate, prepare, unit, integration, e2e, docs, or task-specific verification commands",
		"selected profile/gate id",
		"`selected_profile_id` plus `selected_gate_id`",
		"command, timeout, applicability",
		"Result vocabulary is exactly `pass`, `fail`, or `not_applicable`",
		"full stdout/stderr as an artifact",
		"full-log artifact preservation and compact console/model-visible summary contract",
		"compact console/model-visible summary",
		"exit code, duration, log path, log checksum",
		"bounded failure excerpt",
		"deterministic failure extractor posture",
		"Go, Python/pytest, JavaScript/Vitest/Jest, Playwright, and generic traceback/error patterns",
		"KAS owns policy/profile selection",
		"KAH may later validate only mechanical evidence fields",
		"TOKEN-007 does not define the TOKEN-010 changed-path skip matrix",
		"KAB/Hermes runtime behavior, profile/auth/token/provider/gateway/model configuration, headroom/proxy dependencies, and KAH gate implementation are out of scope",
	})
}

func TestToken007GuidanceSurfacesUseSelectedVerificationProfile(t *testing.T) {
	verificationSurfaces := []string{
		"templates/run-artifacts/phase-plan.yaml.tmpl",
		"templates/run-artifacts/checklist.md.tmpl",
		"registries/phase-contracts.yaml",
		"skills/kkachi-implement/SKILL.md",
		"skills/kkachi-enhance-test/SKILL.md",
		"skills/kkachi-optimize/SKILL.md",
		"skills/kkachi-final-verify/SKILL.md",
	}
	for _, rel := range verificationSurfaces {
		requireContainsAll(t, rel, []string{
			"selected verification profile/gate",
		})
	}

	requireContainsAll(t, "templates/run-artifacts/phase-plan.yaml.tmpl", []string{
		"verification_profile:",
		"selected_profile_id:",
		"selected_gate_id:",
		"command:",
		"timeout:",
		"applicability:",
		"not_applicable_reason:",
		"not_applicable_note:",
		"When status is not_applicable, preserve selected_profile_id, selected_gate_id, command, timeout, applicability, and not_applicable_reason; leave execution evidence fields empty/null because no command ran.",
		"status:",
		"exit_code:",
		"duration:",
		"log_path:",
		"log_checksum:",
		"bounded_failure_excerpt:",
		"deterministic_failure_extractor_posture:",
		"pass_fail_or_not_applicable",
		"exit_code_or_null_when_command_ran_no_value_when_not_run",
		"duration_or_null_when_command_ran_no_value_when_not_run",
		"full_log_artifact_path_or_null_when_command_ran_no_value_when_not_run",
		"sha256_or_null_when_command_ran_no_value_when_not_run",
	})
	requireContainsNone(t, "templates/run-artifacts/phase-plan.yaml.tmpl", []string{
		"exit_code_or_not_applicable",
		"duration_or_not_applicable",
		"full_log_artifact_path_or_not_applicable",
		"sha256_or_not_applicable",
	}, "contains TOKEN-007 fake not_applicable execution-evidence placeholder")
	requireContainsNone(t, "registries/phase-contracts.yaml", []string{
		"re-run the aggregate verification command",
	}, "contains TOKEN-007 aggregate-only verification wording")
}

func TestToken007DoesNotHardcodeGlobalMakeTestOrForbiddenRuntimeScope(t *testing.T) {
	requireContainsNone(t, "templates/run-artifacts/phase-plan.yaml.tmpl", []string{
		"aggregate: \"make test\"",
		"prepare: \"make test-prepare\"",
		"unit: \"make test-unit\"",
		"integration: \"make test-int\"",
		"e2e: \"make test-e2e\"",
	}, "contains TOKEN-007 forbidden hard-coded verification command")

	for _, rel := range []string{
		tokenEconomyAgentInstructionSOT,
		"templates/run-artifacts/phase-plan.yaml.tmpl",
		"templates/run-artifacts/checklist.md.tmpl",
		"registries/phase-contracts.yaml",
	} {
		requireContainsNone(t, rel, []string{
			"KAH selects the verification profile",
			"KAH decides whether verification can be skipped",
			"KAB activates the no-agent runner",
			"Hermes runtime patch is required",
			"headroom is required",
			"profile/auth/token/provider/gateway/model configuration mutation is authorized",
			"warning-only verification state",
		}, "contains TOKEN-007 forbidden scope claim")
	}
}

func TestToken003DoesNotInstallRootInstructionFiles(t *testing.T) {
	if _, err := os.Stat(filepath.Join(repoRoot(t), "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("TOKEN-003 must not create root CLAUDE.md; stat error = %v", err)
	}

	requireContainsNone(t, "AGENTS.md", []string{
		"<!-- KAS:MANAGED:BEGIN core-behavior -->",
		"<!-- KAS:MANAGED:END core-behavior -->",
		"<!-- PROJECT:LOCAL:BEGIN -->",
		"<!-- PROJECT:LOCAL:END -->",
		"Repository agent instructions for {{project_name}}.",
		"Project-specific instructions belong here.",
	}, "must not install managed template content into root AGENTS.md; found")
}

func requireContainsNone(t *testing.T, rel string, forbidden []string, message string) {
	t.Helper()
	content := readRepoFile(t, rel)
	for _, phrase := range forbidden {
		if strings.Contains(content, phrase) {
			t.Fatalf("%s %s %q", rel, message, phrase)
		}
	}
}
