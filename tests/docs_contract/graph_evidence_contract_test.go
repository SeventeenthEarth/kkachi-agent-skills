package docscontract

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var graphEvidenceFields = []string{
	"template_id",
	"template_path",
	"template_version",
	"proposal_id",
	"proposal_path",
	"semantic_diff_output_path",
	"validation_report_path",
	"explain_report_path",
	"approval_evidence_ref",
	"audit_evidence_path",
	"graph_checksum",
	"graph_version",
	"kah_graph_audit_event_ids",
	"capability_check_evidence",
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func requireContainsAll(t *testing.T, rel string, needles []string) {
	t.Helper()
	content := readRepoFile(t, rel)
	for _, needle := range needles {
		if !strings.Contains(content, needle) {
			t.Fatalf("%s missing %q", rel, needle)
		}
	}
}

func TestGraphEvidenceArtifactPreservesCanonicalFields(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/graph-evidence.md.tmpl", append(graphEvidenceFields,
		"kkachi-agent-helper graph diff",
		"capability-check.md",
		"`kkachi-agent-helper graph`",
		"do not directly edit `.kkachi-workflow.yaml`",
	))
}

func TestGraphEvidenceContractSurfacesStayAligned(t *testing.T) {
	contractSurfaces := []string{
		"templates/run-artifacts/task-contract.yaml.tmpl",
		"registries/graph-template-registry.yaml",
		"docs/sot/workflow-graph-integration.md",
		"docs/sot/graph-template-registry.md",
		"docs/sot/khs-architecture-and-integration.md",
		"skills/kkachi-final-verify/SKILL.md",
	}
	for _, rel := range contractSurfaces {
		requireContainsAll(t, rel, graphEvidenceFields)
		requireContainsAll(t, rel, []string{"graph-evidence.md"})
	}
}

func TestGraphEvidenceContractAvoidsForbiddenFallbackClaims(t *testing.T) {
	content := strings.Join([]string{
		readRepoFile(t, "templates/run-artifacts/graph-evidence.md.tmpl"),
		readRepoFile(t, "docs/sot/workflow-graph-integration.md"),
		readRepoFile(t, "docs/sot/graph-template-registry.md"),
		readRepoFile(t, "skills/kkachi-final-verify/SKILL.md"),
	}, "\n")

	forbidden := []string{
		"kah graph apply",
		"kah graph validate",
		"direct `.kkachi-workflow.yaml` fallback",
		"use `.kkachi-workflow.yaml` as fallback",
		"repair `.kkachi-workflow.yaml` through direct",
	}
	for _, phrase := range forbidden {
		if strings.Contains(content, phrase) {
			t.Fatalf("GRAPHMVP-004 contract contains forbidden fallback or alias claim %q", phrase)
		}
	}
}
