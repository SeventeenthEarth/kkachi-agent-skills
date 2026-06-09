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
		"selected 6 KAS PR + 1 dependent KAH PR workstream",
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
		"6 KAS PRs plus 1 dependent KAH PR",
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
	requireRoadmapTaskStatus(t, "TOKEN-005", "Planned")
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
		"TOKEN-005 | Implement approved lifecycle writes and uninstall vault backup | Planned",
	})
	requireContainsAll(t, tokenEconomyAgentInstructionSOT, []string{
		"KAS PR 3: English agent instruction templates and management workflow",
		"CLI or documented dry-run workflow reports create/update/no-change/not_applicable posture without blind overwrite.",
		"KAS PR 4: project KAS lifecycle UX and read-only planner surface",
		"KAS PR 5: approved lifecycle writes, update apply, and uninstall with vault backup",
	})
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
