package docscontract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const graphWorkflowSyncSOT = "docs/sot/graph-workflow-sync-compatibility.md"
const graphWorkflowSyncRegistry = "registries/graph-workflow-sync-compatibility.yaml"

var graphWorkflowSyncStatuses = []string{
	"pass",
	"custom_supported",
	"update_kah_required",
	"update_kah_recommended",
	"update_kas_recommended",
	"graph_missing",
	"graph_stale",
	"graph_broken",
	"graph_conflict",
	"proposal_available",
	"blocked_for_approval",
	"unsupported",
}

func TestGRSYNC001RegistryExistsAndDefinesCompatibilityEnvelope(t *testing.T) {
	if _, err := os.Stat(filepath.Join(repoRoot(t), graphWorkflowSyncRegistry)); err != nil {
		t.Fatalf("%s must exist for GRSYNC-001: %v", graphWorkflowSyncRegistry, err)
	}

	requireContainsAll(t, graphWorkflowSyncRegistry, []string{
		"kind: kas_graph_workflow_sync_compatibility",
		"status: implemented",
		"kas_version: 0.1.2",
		"package: kkachi-agent-helper",
		"min_required: 0.1.9",
		"recommended: 0.1.9",
		"- 0.1.9",
		"- workflow-graph/v1",
		"id: kas-default",
		"version: 0.1.0",
		"path: templates/workflow-graphs/kas-default.yaml",
		"source_registry: registries/graph-template-registry.yaml",
	})
}

func TestGRSYNC001RegistryDocumentsRequiredKAHGraphCapabilityEnvelope(t *testing.T) {
	requireContainsAll(t, graphWorkflowSyncRegistry, []string{
		"kkachi-agent-helper capabilities --json",
		"kkachi-agent-helper graph --help",
		"kkachi-agent-helper graph validate --json",
		"kkachi-agent-helper graph explain --json",
		"kkachi-agent-helper graph diff",
		"kkachi-agent-helper graph propose",
		"kkachi-agent-helper graph apply",
		"kkachi-agent-helper graph export",
		"init --from-template",
		"proposal_first_graph_repair",
		"approval_gated_graph_apply",
		"graph_audit_checksum_evidence",
	})
}

func TestGRSYNC001RegistryStatusVocabularyMatchesSOT(t *testing.T) {
	registry := readRepoFile(t, graphWorkflowSyncRegistry)
	for _, status := range graphWorkflowSyncStatuses {
		if !strings.Contains(registry, "status: "+status+"\n    remediation:") {
			t.Fatalf("%s missing remediation for status %q", graphWorkflowSyncRegistry, status)
		}
		requireContainsAll(t, graphWorkflowSyncSOT, []string{"`" + status + "`"})
	}
}

func TestGRSYNC001RegistryReleaseEvidencePreservesBoundaries(t *testing.T) {
	requireContainsAll(t, graphWorkflowSyncRegistry, []string{
		"kas_release: 0.1.2",
		"kah_release: 0.1.9",
		"KAS v0.1.2 graph workflow sync support is tied to KAH v0.1.9",
		"no KAH code change",
		"no KAB runtime state mutation",
		"no Hermes profile mutation",
		"no auth/token/provider/gateway/model mutation",
		"no automatic KAH binary update",
		"no automatic graph apply from cron or CI",
		"no direct .kkachi-workflow.yaml edit fallback",
		"no Kkachi v2 workflow fallback",
	})

	combined := strings.Join([]string{
		readRepoFile(t, graphWorkflowSyncRegistry),
		readRepoFile(t, graphWorkflowSyncSOT),
		readRepoFile(t, "docs/roadmap.md"),
	}, "\n")
	forbidden := []string{
		"automatic KAH binary update is authorized",
		"automatic graph apply from cron is authorized",
		"automatic graph apply from CI is authorized",
		"direct `.kkachi-workflow.yaml` edit fallback is allowed",
		"KAB has graph policy authority",
		"Kkachi v2 workflow fallback is supported",
		"Stage 2/3 KAB execution changes are authorized",
		"auth/token/gateway/provider/model mutation is authorized",
	}
	for _, phrase := range forbidden {
		if strings.Contains(combined, phrase) {
			t.Fatalf("GRSYNC-001 surfaces contain forbidden boundary claim %q", phrase)
		}
	}
}

func TestGRSYNC001DocsAndRoadmapRegisterCompatibilityRegistry(t *testing.T) {
	requireContainsAll(t, graphWorkflowSyncSOT, []string{
		"Status: accepted SOT for KAS v0.1.2 graph workflow sync support",
		graphWorkflowSyncRegistry,
		"GRSYNC-001 implements the KAS-side machine-readable compatibility source",
		"GRSYNC-002/003",
	})
	requireContainsAll(t, "docs/README.md", []string{
		"sot/graph-workflow-sync-compatibility.md",
		graphWorkflowSyncRegistry,
		"KAH v0.1.9 is the `min_required`, `recommended`, and `tested` dependency",
	})
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		graphWorkflowSyncSOT,
		"compatibility_registries:",
		graphWorkflowSyncRegistry,
	})
	requireRoadmapTaskStatus(t, "GRSYNC-001", "Completed")
	requireRoadmapTaskStatus(t, "GRSYNC-002", "Planned")
	requireRoadmapTaskStatus(t, "GRSYNC-003", "Planned")
	requireContainsAll(t, "docs/roadmap.md", []string{
		"docs-contract coverage validates registry existence/content",
		"Red `t_be0c21da`",
		"Orange `t_7c7854e7`",
		"Gray `t_5b798056`",
		"Blue synthesis `t_d2d19ec6`",
		"No KAH code, KAB runtime, profile mutation",
	})
}

func TestGRSYNC001RegistryAlignsWithExistingGraphTemplateRegistry(t *testing.T) {
	requireContainsAll(t, "registries/graph-template-registry.yaml", []string{
		"id: kas-default",
		"path: templates/workflow-graphs/kas-default.yaml",
		"version: 0.1.0",
		"status: implemented",
	})
	requireContainsAll(t, graphWorkflowSyncRegistry, []string{
		"id: kas-default",
		"version: 0.1.0",
		"path: templates/workflow-graphs/kas-default.yaml",
		"source_registry: registries/graph-template-registry.yaml",
	})
	requireContainsAll(t, "templates/workflow-graphs/kas-default.yaml", []string{
		"version: \"workflow-graph/v1\"",
		"source_template: \"kas-default\"",
	})
}
