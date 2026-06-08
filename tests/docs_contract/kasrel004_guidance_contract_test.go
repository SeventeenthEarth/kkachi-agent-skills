package docscontract

import (
	"strings"
	"testing"
)

var kasrel004ActiveGuidanceSurfaces = []string{
	"skills/kkachi-install-guide/SKILL.md",
	"skills/kkachi-orchestrate/SKILL.md",
	"skills/kkachi-plan/SKILL.md",
	"skills/kkachi-review/SKILL.md",
	"skills/kkachi-verify/SKILL.md",
	"skills/kkachi-final-verify/SKILL.md",
	"skills/kkachi-request-feedback/SKILL.md",
}

var kasrel004RequiredEvidenceFields = []string{
	"provenance_contract_version",
	"source_class_evidence",
	"dependency_audit",
	"skill_dependencies",
	"command_surface_dependencies",
	"deleted_bundle_reference",
	"deleted_bundle_diagnostics",
}

func TestKASREL004GuidanceRequiresProvenanceDependencyEvidence(t *testing.T) {
	section := kasrel004SharedGuidanceGateSection(t)
	requireTextContainsAll(t, kasrel001SOT, section, []string{
		"Before KAS guidance claims",
		"install health",
		"readiness",
		"release compatibility",
		"orchestration safety",
		"review PASS",
		"verification PASS",
		"final completion",
		"current non-secret KASREL provenance/dependency evidence",
		"list --json",
		"install --dry-run --json",
		"doctor --json",
		"Missing, ambiguous, or stale evidence fails closed",
		"do not upgrade it to confidence by assumption",
	})
	requireTextContainsAll(t, kasrel001SOT, section, kasrel004RequiredEvidenceFields)

	for _, rel := range kasrel004ActiveGuidanceSurfaces {
		section := kasrel004GuidanceGateSection(t, rel)
		requireTextContainsAll(t, rel, section, []string{
			"Apply the shared KASREL-004 evidence gate",
			"docs/sot/kasrel-hermes-v016-provenance-contract.md",
			"install health",
			"readiness",
			"release compatibility",
			"orchestration safety",
			"review PASS",
			"verification PASS",
			"final completion",
			"Missing, ambiguous, or stale provenance/dependency evidence fails closed",
		})
		requireTextContainsAll(t, rel, section, kasrel004RequiredEvidenceFields)
	}
}

func TestKASREL004GuidancePreservesDeletedBundleFailClosedBoundary(t *testing.T) {
	section := kasrel004SharedGuidanceGateSection(t)
	requireTextContainsAll(t, kasrel001SOT, section, []string{
		"Deleted-bundle references are cleanup/fail-closed diagnostics only.",
		"Do not look up stale bundle paths",
		"substitute another bundle/hub/external/profile/KAS-managed skill",
		"invent fallback candidates",
		"downgrade a required missing deleted-bundle skill to a warning",
		"treat it as blocking readiness when required",
	})

	for _, rel := range kasrel004ActiveGuidanceSurfaces {
		section := kasrel004GuidanceGateSection(t, rel)
		requireTextContainsAll(t, rel, section, []string{
			"deleted_bundle_reference",
			"deleted_bundle_diagnostics",
			"deleted-bundle references remain cleanup/blocking diagnostics",
			"not fallback lookup or substitution authority",
		})
	}
}

func TestKASREL004GuidanceKeepsLayerAndDependencyTaxonomySplit(t *testing.T) {
	section := kasrel004SharedGuidanceGateSection(t)
	requireTextContainsAll(t, kasrel001SOT, section, []string{
		"`skill_dependencies` are named skills only",
		"`command_surface_dependencies`, not fake skill dependencies",
		"KAS owns policy, skill/process guidance, skill-pack provenance/dependency evidence",
		"KAH owns deterministic project state, run artifacts, schemas, events, locks, diagnostics, gates",
		"KAB owns backend runtime/session control and bridge execution evidence",
		"KASREL evidence by itself must not start KAB",
		"Active KAS guidance surfaces may reference this shared gate instead of duplicating the full text",
	})

	for _, rel := range kasrel004ActiveGuidanceSurfaces {
		section := kasrel004GuidanceGateSection(t, rel)
		requireTextContainsAll(t, rel, section, []string{
			"shared KASREL-004 evidence gate",
			"docs/sot/kasrel-hermes-v016-provenance-contract.md",
			"skill_dependencies",
			"command_surface_dependencies",
		})
	}
}

func kasrel004GuidanceGateSection(t *testing.T, rel string) string {
	t.Helper()
	content := readRepoFile(t, rel)
	heading := "## KASREL provenance/dependency evidence gate"
	start := strings.Index(content, heading)
	if start < 0 {
		t.Fatalf("%s missing %q", rel, heading)
	}

	section := content[start:]
	if next := strings.Index(section[len(heading):], "\n## "); next >= 0 {
		section = section[:len(heading)+next]
	}
	return section
}

func kasrel004SharedGuidanceGateSection(t *testing.T) string {
	t.Helper()
	content := readRepoFile(t, kasrel001SOT)
	heading := "## 9. KASREL-004 guidance evidence gate"
	start := strings.Index(content, heading)
	if start < 0 {
		t.Fatalf("%s missing %q", kasrel001SOT, heading)
	}

	section := content[start:]
	if next := strings.Index(section[len(heading):], "\n## "); next >= 0 {
		section = section[:len(heading)+next]
	}
	return section
}

func requireTextContainsAll(t *testing.T, label, content string, needles []string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(content, needle) {
			t.Fatalf("%s missing %q", label, needle)
		}
	}
}

func TestKASREL004RoadmapStatusBoundary(t *testing.T) {
	requireRoadmapTaskStatus(t, "KASREL-003", "Completed")
	requireRoadmapTaskStatus(t, "KASREL-004", "In Progress")
	requireContainsAll(t, "docs/roadmap.md", []string{
		"commit `75d0361`",
		"evt-001339",
		"evt-001340",
		"Do not mark Completed until those KASREL-004 gates are complete.",
	})
}

func TestKASREL004ActiveSurfacesAvoidForbiddenClaims(t *testing.T) {
	for _, rel := range append(kasrel004ActiveGuidanceSurfaces,
		"docs/README.md",
		"docs/roadmap.md",
		"docs/sot/kasrel-hermes-v016-provenance-contract.md",
	) {
		content := readRepoFile(t, rel)
		forbidden := []string{
			"fallback lookup paths for deleted Hermes bundle skills are authorized",
			"substitute another skill for a deleted bundle skill",
			"command surfaces are skill dependencies",
			"KASREL activates KAB",
			"KAH owns skill provenance",
			"KAB owns skill provenance",
			"deleted-bundle fallback lookup path",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Fatalf("%s contains forbidden KASREL-004 claim %q", rel, phrase)
			}
		}
	}
}
