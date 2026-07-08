package docscontract

import (
	"strings"
	"testing"
)

func TestV02FLOW015ExecutorLoopPhaseProofContract(t *testing.T) {
	requireContainsAll(t, "docs/sot/v020-gjc-workflow-train-corrections.md", []string{
		"V02FLOW-015 mutation-phase fixture/e2e proof",
		"selected executor lane `gjc_ultragoal_executor_loop_candidate`",
		"requested phase id",
		"canonical phase id",
		"argv/process refs",
		"real-user `HOME`",
		"non-empty diff refs",
		"checkpoint ref/status/timestamp",
		"verification command/status/output refs",
		"repeat/termination reason",
		"candidate-only authority boundaries",
	})
}

func TestV02FLOW015PhaseAliasesCannotBypassExecutorLoop(t *testing.T) {
	requireContainsAll(t, "docs/sot/v020-gjc-workflow-train-corrections.md", []string{
		"Known workflow-projection aliases must map to the same executor-loop requirement instead of bypassing it",
		"`impl`/`implement`",
		"`test-enhance`/`test_enhance`/`enhance-test`/`enhance_test`",
		"`ai-slop-cleaner`/`ai_slop_cleaner`/`slop_cleanup`",
		"`docs-update`/`docs_update`/`update_docs`/`docs`",
	})
}

func TestV02FLOW015SkillPromptsRequireCompleteExecutorLoopEvidence(t *testing.T) {
	for _, rel := range []string{
		"skills/kkachi-implement/SKILL.md",
		"skills/kkachi-enhance-test/SKILL.md",
		"skills/kkachi-optimize/SKILL.md",
		"skills/kkachi-docs-update/SKILL.md",
		"skills/kkachi-review/SKILL.md",
	} {
		requireContainsAll(t, rel, []string{
			"V02FLOW-015",
			"gjc_ultragoal_executor_loop_candidate",
			"phase/canonical phase",
			"argv/process refs",
			"real-user `HOME`",
			"cwd",
			"timestamps",
			"exit code",
			"non-empty `diff_refs`",
			"repeat/termination reason",
			"authority overclaim",
		})
	}
}

func TestV02FLOW015RegistryKeepsMutationPhasesOnExecutorLoopLane(t *testing.T) {
	registry := readRepoFile(t, "registries/task-dag-workflow-registry.yaml")
	for _, node := range []string{"implement", "enhance_test", "ai_slop_cleaner", "optimize", "update_docs"} {
		marker := "node_id: " + node
		start := strings.Index(registry, marker)
		if start < 0 {
			t.Fatalf("registry missing node %s", node)
		}
		next := strings.Index(registry[start+len(marker):], "\n  - workflow_id:")
		section := registry[start:]
		if next >= 0 {
			section = registry[start : start+len(marker)+next]
		}
		requireContainsAllInText(t, "registry node "+node, section, []string{
			"execution_lane: gjc_ultragoal_executor_loop_candidate",
			"ultragoal-executor-loop-evidence.md",
			"completion_authority: kah_only",
			"fallback_policy: none_fail_closed",
		})
	}
}
