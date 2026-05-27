package docscontract

import (
	"strings"
	"testing"
)

var staleClean004ActiveGuidanceSurfaces = []string{
	"docs/README.md",
	"docs/sot/concept.md",
	"docs/sot/skill-template.md",
	"docs/sot/external-feedback-intake-policy.md",
	"docs/sot/khs-architecture-and-integration.md",
}

func TestStaleClean004KAHCapabilityPostureIsCurrent(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"KAH 0.1.4 graph and configurable-feedback capability evidence is recognized",
		"CLIMVP and GRAPHMVP-001..004 KAS guidance surfaces are implemented",
		"Do not apply this label to KAH 0.1.4 `kkachi-agent-helper graph` or configurable-feedback surfaces",
		"KAS registries/templates/skills are updated, while operator report/e2e adoption remains `kas-integration-pending` until verified.",
	})

	requireContainsAll(t, "docs/sot/concept.md", []string{
		"KAB is required when a run needs backend execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence.",
		"Feedback may run up to five rounds: round 1 is required when external feedback is in the run contract, and rounds 2..5 are optional continuation rounds.",
		"optional continuation rounds 2..5 may be skipped with an explicit reason.",
	})

	requireContainsAll(t, "docs/sot/skill-template.md", []string{
		"request-feedback round 1 always, optional continuation rounds 2..5",
		"feedback round 1 is mandatory; optional continuation rounds 2..5 may be skipped with explicit reason.",
	})

	requireContainsAll(t, "docs/sot/khs-architecture-and-integration.md", []string{
		"updated by STALECLEAN-004 to optional continuation rounds 2..5",
		"operator report/e2e adoption remains integration-pending until verified",
	})
}

func TestStaleClean004ActiveGuidanceAvoidsAbsentKAHStaleClaims(t *testing.T) {
	for _, rel := range staleClean004ActiveGuidanceSurfaces {
		content := readRepoFile(t, rel)
		forbidden := []string{
			"Feedback may run up to three rounds.",
			"rounds 2-3 may be skipped",
			"additional rounds 2-3",
			"Up to two additional rounds, maximum three pairs | Supersede or amend",
			"Still pending docs cleanup",
			"update skill only through skill/change approval gates",
			"KAS registry/template/skill/report adoption exists",
			"GRAPHMVP artifact/report mapping, and KAB runtime alignment still proceed",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Fatalf("%s retains stale KAH 0.1.4 absent/incomplete guidance %q", rel, phrase)
			}
		}
	}
}
