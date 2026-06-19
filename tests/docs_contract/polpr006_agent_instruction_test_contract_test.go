package docscontract

import (
	"strings"
	"testing"
)

func TestPOLPR006RepoLocalAgentInstructionLifecycleContract(t *testing.T) {
	requireContainsAll(t, "docs/sot/kas-cli-contract.md", []string{
		"update agent-instructions",
		"--repo-path <path>",
		"--source-repo <kas-source-repo>",
		"templates/agent-instructions/*.tmpl",
		"`AGENTS.md` / `CLAUDE.md` repo-local lifecycle",
		"profile-local skill installation",
		"`create`",
		"`update_managed_block`",
		"`no_change`",
		"`not_applicable`",
		"`blocked_unmarked_existing_file`",
		"`error`",
		"`source_template_missing_managed_block`",
		"`existing_managed_block_malformed`",
		"preservation action",
		"`preserve_project_local_block`, not as a top-level file-plan outcome",
		"`--apply dry-run:sha256:<hash>`",
		"fail closed",
	})
	requireContainsAll(t, "docs/sot/token-economy-and-agent-instruction-contract.md", []string{
		"repo-local agent-instruction lifecycle",
		"distinct from profile-local skill installation",
		"blocked_unmarked_existing_file",
		"preservation action",
		"not as a top-level file-plan outcome",
		"source_template_missing_managed_block",
		"existing_managed_block_malformed",
		"dry-run hash",
	})
}

func TestPOLPR006TestLayerContractAndMakeTargets(t *testing.T) {
	requireContainsAll(t, "docs/sot/phase-orchestration-policy.md", []string{
		"unit tests",
		"integration tests",
		"e2e tests",
		"`test-prepare`",
		"`test-unit`",
		"`test-int`",
		"`test-e2e`",
		"`test`",
		"disposable",
		"production/shared",
	})
	requireContainsAll(t, "skills/kkachi-enhance-test/SKILL.md", []string{
		"unit",
		"integration",
		"e2e",
		"test-prepare",
		"test-unit",
		"test-int",
		"test-e2e",
	})
	requireContainsAll(t, "Makefile", []string{
		"test-prepare:",
		"test-unit:",
		"test-int:",
		"test-e2e:",
		"test: test-prepare test-unit test-int test-e2e",
	})
}

func TestPOLPR006FailedTestOwnershipSplitIsSurfaced(t *testing.T) {
	requireContainsAll(t, "docs/sot/phase-orchestration-policy.md", []string{
		"Blue owns compact triage",
		"selected implementer owns detailed RCA",
		"code/docs mutation",
	})
	requireContainsAll(t, "skills/kkachi-handle-feedback/SKILL.md", []string{
		"Blue owns reproduction/classification/routing/acceptance",
		"selected implementer owns detailed RCA",
	})
}

func TestPOLPR006AvoidsFallbackWidening(t *testing.T) {
	for _, rel := range []string{
		"docs/sot/kas-cli-contract.md",
		"docs/sot/token-economy-and-agent-instruction-contract.md",
	} {
		content := readRepoFile(t, rel)
		for _, forbidden := range []string{
			"fall back to profile skill install",
			"automatically migrate unmarked AGENTS.md",
			"production e2e by default",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s contains forbidden fallback wording %q", rel, forbidden)
			}
		}
	}
}
