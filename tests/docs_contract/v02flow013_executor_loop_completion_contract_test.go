package docscontract

import (
	"strings"
	"testing"
)

func TestV02FLOW013StatusHierarchyRequiresExecutorLoopEvidence(t *testing.T) {
	requireContainsAll(t, "docs/sot/v020-gjc-workflow-train-corrections.md", []string{
		"V02FLOW-013 executor-loop completion status hierarchy",
		"implementation_goal_bundle_ready: goal bundle only; never sufficient for implementation completion",
		"implementation_diff_ready: executor-loop source diff, checkpoint, and checksum evidence is present",
		"implementation_verified: executor-loop diff, checkpoint, HOME, checksum, termination, and verification-output evidence passed",
		"Required executor-loop evidence fields: changed_source_refs, diff_refs, checkpoint_ref, checkpoint_status, verification_output_refs, checksums, termination_reason, HOME, no_authority_boundaries",
	})
}

func TestV02FLOW013ImplementerOwnedPhasesRequireExecutorLoopEvidence(t *testing.T) {
	for _, rel := range []string{
		"registries/task-dag-workflow-registry.yaml",
		".kkachi-workflow.yaml",
		"skills/kkachi-implement/SKILL.md",
		"skills/kkachi-enhance-test/SKILL.md",
		"skills/kkachi-optimize/SKILL.md",
		"skills/kkachi-docs-update/SKILL.md",
		"skills/kkachi-handle-feedback/SKILL.md",
		"skills/kkachi-review/SKILL.md",
	} {
		requireContainsAll(t, rel, []string{
			"V02FLOW-013",
			"implementation_goal_bundle_ready is goal-bundle-only and never sufficient for implementation completion",
			"implementation_diff_ready",
			"implementation_verified",
			"changed_source_refs",
			"diff_refs",
			"checkpoint_ref",
			"verification_output_refs",
			"no_authority_boundaries",
		})
	}
}

func TestV02FLOW013RegistryRequiresExecutorLoopArtifactForMutationPhases(t *testing.T) {
	registry := readRepoFile(t, "registries/task-dag-workflow-registry.yaml")
	for _, node := range []string{"implement", "enhance_test", "ai_slop_cleaner", "optimize", "update_docs", "handle_feedback"} {
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
			"ultragoal-executor-loop-evidence.md",
			"completion_authority: kah_only",
			"fallback_policy: none_fail_closed",
		})
	}
}

func TestV02FLOW013RejectsGoalBundleOnlyCompletionLanguage(t *testing.T) {
	active := strings.Join([]string{
		readRepoFile(t, "docs/sot/v020-gjc-workflow-train-corrections.md"),
		readRepoFile(t, "docs/roadmap.md"),
		readRepoFile(t, "registries/task-dag-workflow-registry.yaml"),
		readRepoFile(t, ".kkachi-workflow.yaml"),
		readRepoFile(t, "skills/kkachi-plan/SKILL.md"),
		readRepoFile(t, "skills/kkachi-implement/SKILL.md"),
		readRepoFile(t, "skills/kkachi-enhance-test/SKILL.md"),
		readRepoFile(t, "skills/kkachi-optimize/SKILL.md"),
		readRepoFile(t, "skills/kkachi-docs-update/SKILL.md"),
		readRepoFile(t, "skills/kkachi-handle-feedback/SKILL.md"),
		readRepoFile(t, "skills/kkachi-review/SKILL.md"),
	}, "\n")

	for _, forbidden := range []string{
		"implementation_goal_bundle_ready satisfies implementation completion",
		"implementation_goal_bundle_ready is sufficient for implementation completion",
		"goal-bundle-only evidence satisfies implementation completion",
		"ultragoal_goals_ready satisfies implementation completion",
		"create-goals completion is implementation completion",
	} {
		if strings.Contains(active, forbidden) {
			t.Fatalf("active V02FLOW-013 surfaces contain forbidden goal-bundle-only completion language %q", forbidden)
		}
	}
}
