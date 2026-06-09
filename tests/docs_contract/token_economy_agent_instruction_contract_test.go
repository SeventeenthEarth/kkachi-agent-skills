package docscontract

import (
	"strings"
	"testing"
)

const tokenEconomyAgentInstructionSOT = "docs/sot/token-economy-and-agent-instruction-contract.md"

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
		"Candidate token-economy and agent-instruction SOT",
		"selected candidate 6 KAS PR + 1 dependent KAH PR workstream pending responsible-approver confirmation",
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
