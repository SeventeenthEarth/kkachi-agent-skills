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
	"ralplan",
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
	{"plan", "ralplan"},
	{"ralplan", "implement"},
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

func requireContainsAllInText(t *testing.T, label string, content string, needles []string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(content, needle) {
			t.Fatalf("%s missing %q", label, needle)
		}
	}
}

func extractYAMLMapEntryBlock(t *testing.T, content string, marker string) string {
	t.Helper()
	start := strings.Index(content, marker)
	if start == -1 {
		t.Fatalf("missing YAML map entry marker %q", marker)
	}
	lines := strings.Split(content[start:], "\n")
	markerIndent := len(marker) - len(strings.TrimLeft(marker, " "))
	block := []string{lines[0]}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			block = append(block, line)
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent <= markerIndent {
			break
		}
		block = append(block, line)
	}
	return strings.Join(block, "\n")
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
			"ralplan",
			"enhance-test",
			"ai-slop-cleaner",
			"mar-review",
		})
	}

	requireContainsAll(t, "registries/graph-template-registry.yaml", []string{
		"expected_phase_count: 18",
		"expected_edge_count: 17",
		"custom project workflows remain supportable",
		"explicit approval/question evidence is a boundary, not a default ask phase",
		"phase_id_translation:",
		"older phase-contract/skill activity aliases remain translation-only",
		"run-local phase-plan applicability may still be skipped or not_applicable",
	})
	requireContainsAll(t, "docs/sot/phase-orchestration-policy.md", []string{
		"plan -> ralplan -> implement",
		"plan-vet remains an acceptance/review gate term",
		"not a universal forced",
		"Project-specific `.kkachi-workflow.yaml` composition remains supported",
		"`final` graph phase",
		"`final-verify` is only the skill/activity alias",
	})
	requireContainsAll(t, "skills/kkachi-orchestrate/SKILL.md", []string{
		"plan -> ralplan -> implement",
		"plan-vet remains an acceptance/review gate term",
		"must not be forced onto custom project workflows",
		"non-development task classes",
		"older activity aliases such as `update_docs` and `final_verify` are translation-only compatibility names",
	})
	for _, rel := range []string{"docs/sot/phase-orchestration-policy.md", "skills/kkachi-orchestrate/SKILL.md"} {
		content := readRepoFile(t, rel)
		for _, stale := range []string{"plan -> vet -> implement", "vet -> implement"} {
			if strings.Contains(content, stale) {
				t.Fatalf("%s retains stale default-spine phase ordering %q", rel, stale)
			}
		}
	}
}

func TestPOLPR004RunArtifactTemplatesDoNotRequireDefaultAsk(t *testing.T) {
	targets := []string{
		"templates/run-artifacts/checklist.md.tmpl",
		"templates/run-artifacts/phase-plan.yaml.tmpl",
		"templates/run-artifacts/plan.md.tmpl",
		"templates/run-artifacts/task-contract.yaml.tmpl",
	}
	staleRequiredAsk := []string{
		"phase: ask",
		"| ask |",
		"ask.md",
		"ask_phase_required: true",
		"plan -> ask",
		"vet -> ask",
	}
	for _, rel := range targets {
		content := readRepoFile(t, rel)
		for _, stale := range staleRequiredAsk {
			if strings.Contains(content, stale) {
				t.Fatalf("%s still contains required/default ask wording %q", rel, stale)
			}
		}
		requireContainsAll(t, rel, []string{
			"approval",
		})
	}
}

func TestPOLPR004PhasePlanRoleMappingUsesCorrectedV02Train(t *testing.T) {
	content := readRepoFile(t, "templates/run-artifacts/phase-plan.yaml.tmpl")
	plannerBlock := extractYAMLMapEntryBlock(t, content, "  planner_backend:")
	for _, stale := range []string{"      - ask", "      - request_feedback"} {
		if strings.Contains(plannerBlock, stale) {
			t.Fatalf("planner_backend phases still contain default question/feedback phase %q", stale)
		}
	}
	requireContainsAllInText(t, "planner_backend role mapping", plannerBlock, []string{
		"      - plan",
		"      - ralplan",
	})

	implementerBlock := extractYAMLMapEntryBlock(t, content, "  implementer_backend:")
	requireContainsAllInText(t, "implementer_backend role mapping", implementerBlock, []string{
		"      - implement",
		"      - enhance_test",
		"      - ai_slop_cleaner",
		"      - optimize",
		"      - update_docs",
	})
	for _, stale := range []string{"      - docs_update", "      - docs", "      - final_verify"} {
		if strings.Contains(implementerBlock, stale) {
			t.Fatalf("implementer_backend phases contain stale/non-canonical phase alias %q", stale)
		}
	}
}

func TestPOLPR004PhaseContractsIncludeRalplanInDefaultDevelopmentSurfaces(t *testing.T) {
	content := readRepoFile(t, "registries/phase-contracts.yaml")
	requireContainsAllInText(t, "phase-contracts ralplan defaults", content, []string{
		"canonical_spine:\n  - plan\n  - ralplan\n  - implement",
		"phases: [codegraph_refresh, plan, ralplan, implement",
		"  planner_backend:\n    phases:\n      - plan\n      - ralplan",
		"gjc_ralplan_status_json",
		"substantive_ralplan_candidate_evidence",
		"proceed from ralplan to implementation",
	})
	for _, stale := range []string{
		"phases: [codegraph_refresh, plan, implement",
		"proceed from vet to implementation",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("phase-contracts retains stale/default-bypassing wording %q", stale)
		}
	}
}

func TestPOLPR004ActiveV02SurfacesDoNotExposeDirectCodexDefaultLane(t *testing.T) {
	phasePlan := readRepoFile(t, "templates/run-artifacts/phase-plan.yaml.tmpl")
	for _, stale := range []string{
		"\nstage1_direct_codex_runner:",
		"    - direct_codex_app_server",
	} {
		if strings.Contains(phasePlan, stale) {
			t.Fatalf("phase-plan template exposes active/default direct-Codex lane surface %q", stale)
		}
	}
	requireContainsAllInText(t, "phase-plan historical direct-Codex compatibility", phasePlan, []string{
		"historical_compatibility_lanes:",
		"direct_codex_app_server:",
		"disabled_by_default: true",
		"explicit_selection_required: true",
		"GJC ralplan/ultragoal is the active v0.2 default lane",
	})

	backendProfiles := readRepoFile(t, "registries/backend-prompt-profiles.yaml")
	if strings.Contains(extractYAMLMapEntryBlock(t, backendProfiles, "output_policy:"), "    - direct_codex_app_server") {
		t.Fatalf("backend prompt output_policy still applies to direct_codex_app_server as an active/default lane")
	}
	requireContainsAllInText(t, "backend prompt profile historical direct-Codex compatibility", backendProfiles, []string{
		"historical_compatibility_lanes:",
		"direct_codex_app_server:",
		"disabled_by_default: true",
		"explicit_selection_required: true",
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
