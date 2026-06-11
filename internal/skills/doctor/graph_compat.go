package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const graphWorkflowSyncRegistryPath = "registries/graph-workflow-sync-compatibility.yaml"

type WorkflowGraphCompatibility struct {
	Path                    string                  `json:"path"`
	KASVersion              string                  `json:"kas_version"`
	KAHMinRequired          string                  `json:"kah_min_required"`
	KAHRecommended          string                  `json:"kah_recommended"`
	KAHTested               []string                `json:"kah_tested"`
	SupportedSchemaVersions []string                `json:"supported_schema_versions"`
	SupportedTemplates      []WorkflowGraphTemplate `json:"supported_templates"`
}

type WorkflowGraphTemplate struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Path    string `json:"path,omitempty"`
}

func loadWorkflowGraphCompatibility(sourceRepo string) (WorkflowGraphCompatibility, error) {
	compat := WorkflowGraphCompatibility{Path: graphWorkflowSyncRegistryPath}
	data, err := os.ReadFile(filepath.Join(sourceRepo, graphWorkflowSyncRegistryPath))
	if err != nil {
		return compat, err
	}
	text := string(data)
	required := []string{
		"kind: kas_graph_workflow_sync_compatibility",
		"kas_version: 0.1.2",
		"min_required: 0.1.9",
		"recommended: 0.1.9",
		"- 0.1.9",
		"- workflow-graph/v1",
		"id: kas-default",
		"version: 0.1.0",
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			return compat, fmt.Errorf("workflow graph compatibility registry missing %q", want)
		}
	}
	compat.KASVersion = "0.1.2"
	compat.KAHMinRequired = "0.1.9"
	compat.KAHRecommended = "0.1.9"
	compat.KAHTested = []string{"0.1.9"}
	compat.SupportedSchemaVersions = []string{"workflow-graph/v1"}
	compat.SupportedTemplates = []WorkflowGraphTemplate{{ID: "kas-default", Version: "0.1.0", Path: "templates/workflow-graphs/kas-default.yaml"}}
	return compat, nil
}

func workflowGraphStatusRemediation(status string) string {
	switch status {
	case "pass":
		return "No graph workflow sync remediation required."
	case "custom_supported":
		return "Preserve the custom graph; do not force replacement with the default template."
	case "update_kah_required":
		return "Update kkachi-agent-helper to at least 0.1.9 before graph workflow sync repair."
	case "update_kah_recommended":
		return "Recommend kkachi-agent-helper 0.1.9 when the effective helper is older than the tested release."
	case "update_kas_recommended":
		return "Recommend newer KAS when the graph schema or policy is newer than KAS 0.1.2 understands."
	case "graph_missing":
		return "Graph is missing; proposal-first initialization belongs to GRSYNC-003 and was not attempted."
	case "graph_stale":
		return "Graph is stale; proposal-first repair belongs to GRSYNC-003 and was not attempted."
	case "graph_broken":
		return "Graph is broken; proposal-first repair is allowed only after KAH diagnostics identify a supported repair path."
	case "graph_conflict":
		return "Fail closed and record diagnostics because KAS metadata and KAH graph evidence disagree."
	default:
		return "Fail closed; require KAS/KAH support update or manual governance decision."
	}
}
