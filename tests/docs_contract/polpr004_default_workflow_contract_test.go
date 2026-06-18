package docscontract

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var polpr004DefaultPhases = []string{
	"intake",
	"sot",
	"roadmap",
	"task-classification",
	"plan",
	"vet",
	"ask",
	"implement",
	"enhance-test",
	"ai-slop-cleaner",
	"optimize",
	"docs",
	"verify",
	"review",
	"request-feedback-1",
	"handle-feedback-1",
	"mar-review",
	"second-color-review",
	"final",
}

var polpr004DefaultEdges = []phaseEdge{
	{"intake", "sot"},
	{"sot", "roadmap"},
	{"roadmap", "task-classification"},
	{"task-classification", "plan"},
	{"plan", "vet"},
	{"vet", "ask"},
	{"ask", "implement"},
	{"implement", "enhance-test"},
	{"enhance-test", "ai-slop-cleaner"},
	{"ai-slop-cleaner", "optimize"},
	{"optimize", "docs"},
	{"docs", "verify"},
	{"verify", "review"},
	{"review", "request-feedback-1"},
	{"request-feedback-1", "handle-feedback-1"},
	{"handle-feedback-1", "mar-review"},
	{"mar-review", "second-color-review"},
	{"second-color-review", "final"},
}

var polpr004DeferredResiduals = []string{
	".kkachi-workflow.yaml",
	"feedback.md",
}

type phaseEdge struct {
	from string
	to   string
}

func TestPOLPR004KASDefaultTemplateAndRegistryStayAligned(t *testing.T) {
	template := readRepoFile(t, "templates/workflow-graphs/kas-default.yaml")
	registry := readRepoFile(t, "registries/graph-template-registry.yaml")

	templatePhases := extractIDsFromBlock(t, template, "phases:", "edges:", "template phases")
	templateEdges := extractEdgesFromBlock(t, template, "edges:", "proposals:", "template edges")
	registryEntry := extractIndentedListEntry(t, registry, "templates:", "  - id: kas-default")
	registryPhases := extractIDsFromBlock(t, registryEntry, "    phases:", "    edges:", "registry phases")
	registryEdges := extractEdgesFromBlock(t, registryEntry, "    edges:", "    gates:", "registry edges")

	requireStringSlicesEqual(t, "template phases", templatePhases, polpr004DefaultPhases)
	requireStringSlicesEqual(t, "registry phases", registryPhases, polpr004DefaultPhases)
	requireEdgesEqual(t, "template edges", templateEdges, polpr004DefaultEdges)
	requireEdgesEqual(t, "registry edges", registryEdges, polpr004DefaultEdges)
	if got := extractIntField(t, registryEntry, "expected_phase_count"); got != len(polpr004DefaultPhases) {
		t.Fatalf("expected_phase_count = %d, want %d", got, len(polpr004DefaultPhases))
	}
	if got := extractIntField(t, registryEntry, "expected_edge_count"); got != len(polpr004DefaultEdges) {
		t.Fatalf("expected_edge_count = %d, want %d", got, len(polpr004DefaultEdges))
	}
}

func TestPOLPR004ActiveDefaultWorkflowSurfacesUseMARAndPreserveCustomGraphs(t *testing.T) {
	for _, rel := range []string{
		"registries/graph-template-registry.yaml",
		"templates/workflow-graphs/kas-default.yaml",
		"docs/sot/phase-orchestration-policy.md",
		"skills/kkachi-orchestrate/SKILL.md",
	} {
		requireContainsAll(t, rel, []string{
			"enhance-test",
			"ai-slop-cleaner",
			"mar-review",
		})
	}

	requireContainsAll(t, "registries/graph-template-registry.yaml", []string{
		"expected_phase_count: 19",
		"expected_edge_count: 18",
		"custom project workflows remain supportable",
		"phase_id_translation:",
		"older phase-contract/skill activity aliases remain translation-only",
		"run-local phase-plan applicability may still be skipped or not_applicable",
	})
	requireContainsAll(t, "docs/sot/phase-orchestration-policy.md", []string{
		"not a universal forced",
		"Project-specific `.kkachi-workflow.yaml` composition remains supported",
		"`final` graph phase",
		"`final-verify` is only the skill/activity alias",
	})
	requireContainsAll(t, "skills/kkachi-orchestrate/SKILL.md", []string{
		"must not be forced onto custom project workflows",
		"non-development task classes",
		"older activity aliases such as `update_docs` and `final_verify` are translation-only compatibility names",
	})
}

func TestPOLPR004ActiveDefaultGraphSurfacesDoNotRegressToOctoReview(t *testing.T) {
	activeDefaultSurfaces := []string{
		"registries/graph-template-registry.yaml",
		"templates/workflow-graphs/kas-default.yaml",
		"docs/sot/phase-orchestration-policy.md",
		"skills/kkachi-orchestrate/SKILL.md",
		"skills/kkachi-orchestrate/references/run-operating-policy.md",
	}
	for _, rel := range activeDefaultSurfaces {
		content := readRepoFile(t, rel)
		for _, stale := range []string{"octo-review", "GLM Octo", "KAH 0.1.4 built-in khs-default"} {
			if strings.Contains(content, stale) {
				t.Fatalf("%s contains stale active default workflow wording %q", rel, stale)
			}
		}
	}
}

func TestPOLPR004ActiveStaleScanOmissionsArePinnedAsDeferredResiduals(t *testing.T) {
	// These root residuals are intentionally deferred to POLPR-005/POLPR-008
	// and are not part of the active default-surface stale scan for POLPR-004.
	activeDefaultSurfaces := []string{
		"registries/graph-template-registry.yaml",
		"templates/workflow-graphs/kas-default.yaml",
		"docs/sot/phase-orchestration-policy.md",
		"skills/kkachi-orchestrate/SKILL.md",
		"skills/kkachi-orchestrate/references/run-operating-policy.md",
	}
	for _, residual := range polpr004DeferredResiduals {
		for _, active := range activeDefaultSurfaces {
			if active == residual {
				t.Fatalf("deferred residual %s must not be scanned as an active POLPR-004 default surface", residual)
			}
		}
	}
}

func extractIndentedListEntry(t *testing.T, content string, listHeader string, entryHeader string) string {
	t.Helper()
	start := strings.Index(content, listHeader)
	if start == -1 {
		t.Fatalf("missing list header %q", listHeader)
	}
	content = content[start:]
	entryStart := strings.Index(content, entryHeader)
	if entryStart == -1 {
		t.Fatalf("missing list entry %q", entryHeader)
	}
	entry := content[entryStart:]
	nextEntry := strings.Index(entry[len(entryHeader):], "\n  - id: ")
	if nextEntry == -1 {
		return entry
	}
	return entry[:len(entryHeader)+nextEntry]
}

func extractIDsFromBlock(t *testing.T, content string, startMarker string, endMarker string, label string) []string {
	t.Helper()
	block := extractBlock(t, content, startMarker, endMarker, label)
	idPattern := regexp.MustCompile(`(?m)^\s*-\s+id:\s+"?([^"\n]+)"?\s*$`)
	matches := idPattern.FindAllStringSubmatch(block, -1)
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match[1])
	}
	if len(ids) == 0 {
		t.Fatalf("%s has no phase ids", label)
	}
	return ids
}

func extractEdgesFromBlock(t *testing.T, content string, startMarker string, endMarker string, label string) []phaseEdge {
	t.Helper()
	block := extractBlock(t, content, startMarker, endMarker, label)
	fromPattern := regexp.MustCompile(`^\s*-\s+from:\s+"?([^"\n]+)"?\s*$`)
	toPattern := regexp.MustCompile(`^\s*to:\s+"?([^"\n]+)"?\s*$`)
	lines := strings.Split(block, "\n")
	edges := []phaseEdge{}
	for index := 0; index < len(lines); index++ {
		fromMatch := fromPattern.FindStringSubmatch(lines[index])
		if fromMatch == nil {
			continue
		}
		if index+1 >= len(lines) {
			t.Fatalf("%s has edge from %q without to", label, fromMatch[1])
		}
		toMatch := toPattern.FindStringSubmatch(lines[index+1])
		if toMatch == nil {
			t.Fatalf("%s has edge from %q with malformed to line %q", label, fromMatch[1], lines[index+1])
		}
		edges = append(edges, phaseEdge{from: fromMatch[1], to: toMatch[1]})
	}
	if len(edges) == 0 {
		t.Fatalf("%s has no edges", label)
	}
	return edges
}

func extractBlock(t *testing.T, content string, startMarker string, endMarker string, label string) string {
	t.Helper()
	start := strings.Index(content, startMarker)
	if start == -1 {
		t.Fatalf("%s missing start marker %q", label, startMarker)
	}
	content = content[start+len(startMarker):]
	end := strings.Index(content, endMarker)
	if end == -1 {
		t.Fatalf("%s missing end marker %q", label, endMarker)
	}
	return content[:end]
}

func extractIntField(t *testing.T, content string, field string) int {
	t.Helper()
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*%s:\s*([0-9]+)\s*$`, regexp.QuoteMeta(field)))
	match := pattern.FindStringSubmatch(content)
	if match == nil {
		t.Fatalf("missing integer field %q", field)
	}
	value, err := strconv.Atoi(match[1])
	if err != nil {
		t.Fatalf("parse %s: %v", field, err)
	}
	return value
}

func requireStringSlicesEqual(t *testing.T, label string, got []string, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s mismatch\ngot:  %v\nwant: %v", label, got, want)
	}
}

func requireEdgesEqual(t *testing.T, label string, got []phaseEdge, want []phaseEdge) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s mismatch\ngot:  %v\nwant: %v", label, got, want)
	}
}
