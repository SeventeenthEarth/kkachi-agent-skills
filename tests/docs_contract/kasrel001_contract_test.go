package docscontract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const kasrel001SOT = "docs/sot/kasrel-hermes-v016-provenance-contract.md"
const oldKasrel001PlanSOT = "docs/sot/TODO-kasrel-hermes-v016-provenance-plan.md"

func TestKASREL001AcceptedSOTDefinesProvenanceVocabulary(t *testing.T) {
	requireContainsAll(t, kasrel001SOT, []string{
		"Status: accepted SOT for KASREL-001",
		"skill provenance",
		"effective skill",
		"source class",
		"source evidence",
		"skill dependency",
		"command-surface dependency",
		"deleted-bundle reference",
		"Shadowing/conflict",
		"bundle_builtin",
		"hub_installed",
		"ops_external",
		"profile_personal",
		"kas_managed_profile",
		"unknown_or_unclassified",
	})
}

func TestKASREL001AcceptedSOTPreservesBoundariesAndNoFallbackPolicy(t *testing.T) {
	requireContainsAll(t, kasrel001SOT, []string{
		"Hermes CLI source labels are advisory only",
		"Deleted-bundle references are cleanup/fail-closed diagnostics only.",
		"look up stale bundle paths for deleted skills",
		"substitute a different bundle, hub, external, profile, or KAS-managed skill",
		"invent fallback candidates",
		"Command surfaces must not be represented as fake skill dependencies.",
		"KASREL does not activate KAB.",
		"no production CLI changes under `internal/skills/**` unless the specific later implementation task authorizes them",
	})

	content := readRepoFile(t, kasrel001SOT)
	forbidden := []string{
		"fallback lookup paths for deleted Hermes bundle skills are authorized",
		"substitute another skill for a deleted bundle skill",
		"command surfaces are skill dependencies",
		"KASREL activates KAB",
	}
	for _, phrase := range forbidden {
		if strings.Contains(content, phrase) {
			t.Fatalf("KASREL-001 contract contains forbidden boundary claim %q", phrase)
		}
	}
}

func TestKASREL001AcceptedSOTKeepsTaskStatusAndMutationBoundaries(t *testing.T) {
	requireContainsAll(t, kasrel001SOT, []string{
		"not implementation authorization and not installed-profile mutation approval",
		"profile install, profile repair, approved copy install, or other profile mutation",
		"binary rebuild/install or installed-binary parity repair",
		"auth, token, gateway, provider, model, credential, or profile configuration mutation",
		"KASREL-002 is the completed implementation task for skill inventory and provenance classification.",
		"KASREL-003 is the completed implementation task for dependency audit behavior",
		"KASREL-004 is the in-progress guidance-update task for install/readiness/orchestration/review/final-verification behavior",
		"no KASUPD write-capable sync behavior unless a separate KASUPD task authorizes it",
	})

	requireContainsAll(t, "docs/README.md", []string{
		"Accepted docs/spec-only contract for provenance/dependency vocabulary and release-compatibility boundaries",
		"use the SOT for detailed source-class, diagnostic, JSON-field, and follow-on task boundaries",
		"does not authorize profile mutation, production CLI implementation, or KAB activation",
	})

	requireContainsAll(t, "docs/roadmap.md", []string{
		"KASREL deferrals unless separately approved:",
		"automatic profile repair",
		"binary rebuild/install execution",
		"write-capable install approval",
		"KAB Stage 2/Stage 3 activation",
		"auth/token/gateway/provider/model mutation",
		"fallback path for removed Hermes bundle skills",
		"KASREL-001 is `Completed`",
		"Current task rows supersede that historical status only after their own implementation evidence and review gates",
	})
}

func TestKASREL001AcceptedSOTJSONFieldsRemainIllustrative(t *testing.T) {
	requireContainsAll(t, kasrel001SOT, []string{
		"KASREL contract fields implemented by KASREL-002 and KASREL-003",
		"Concrete JSON values and examples below are illustrative only.",
		"provenance_contract_version",
		"source_inventory_summary",
		"counts_by_source_class",
		"source_class_evidence",
		"provenance_state",
		"deleted_bundle_reference",
		"skill_dependencies",
		"command_surface_dependencies",
		"source_inventory_snapshot",
		"target_profile_inventory",
		"provenance_conflicts",
		"shadowing_conflicts",
		"dependency_audit",
		"deleted_bundle_diagnostics",
		"approval_request.hash_includes_provenance: true",
		"provenance_audit",
		"source_class_ambiguous",
		"command_surface_missing",
		"skill_dependency_missing",
	})
}

func TestKASREL001DocsIndexAndRoadmapPointToAcceptedSOT(t *testing.T) {
	oldRel := filepath.Join(repoRoot(t), oldKasrel001PlanSOT)
	if _, err := os.Stat(oldRel); !os.IsNotExist(err) {
		t.Fatalf("old TODO KASREL candidate SOT still exists or is not removable from active authority: %v", err)
	}

	for _, rel := range []string{"docs/README.md", "docs/roadmap.md"} {
		content := readRepoFile(t, rel)
		if strings.Contains(content, oldKasrel001PlanSOT) {
			t.Fatalf("%s still points to old TODO KASREL candidate SOT", rel)
		}
		if !strings.Contains(content, kasrel001SOT) && !strings.Contains(content, strings.TrimPrefix(kasrel001SOT, "docs/")) {
			t.Fatalf("%s does not point to accepted KASREL-001 SOT", rel)
		}
	}

	requireRoadmapTaskStatus(t, "KASREL-001", "Completed")
	requireRoadmapTaskStatus(t, "KASREL-002", "Completed")
	requireRoadmapTaskStatus(t, "KASREL-003", "Completed")
	requireRoadmapTaskStatus(t, "KASREL-004", "In Progress")
}

func TestKASREL002DocsReflectCompletedImplementationWithPreCommitBoundary(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"KASREL-002 implementation is completed/pre-commit-ready",
		"KASREL-003 dependency-audit implementation is completed",
		"KASREL-004 guidance work is in progress",
		"required review gates",
	})

	requireContainsAll(t, "docs/roadmap.md", []string{
		"| KASREL-002 | Implement skill inventory and provenance classification | Completed |",
		"no-write provenance inventory fields to `list --json`, `install --dry-run --json`, and `doctor --json`",
		"empty dependency fields reserved for KASREL-003",
		"no deleted-bundle fallback",
		"official KAB GLM Octo review passed with 0 blocking findings",
		"post-Octo Red/Orange/Gray re-review accepted with 0 blocking findings",
		"Final KAH gate and commit approval remain separate pre-commit gates.",
		"KASREL-003 is `Completed` by commit `75d0361`",
		"KASREL-004 is `In Progress` and must not be marked `Completed`",
	})
}

func requireRoadmapTaskStatus(t *testing.T, taskID, status string) {
	t.Helper()
	for _, line := range strings.Split(readRepoFile(t, "docs/roadmap.md"), "\n") {
		if !strings.HasPrefix(line, "| "+taskID+" |") {
			continue
		}
		if !strings.Contains(line, "| "+status+" |") {
			t.Fatalf("docs/roadmap.md has %s row without status %s: %s", taskID, status, line)
		}
		return
	}
	t.Fatalf("docs/roadmap.md missing %s row", taskID)
}
