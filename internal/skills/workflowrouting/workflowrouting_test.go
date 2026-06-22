package workflowrouting

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRouteShippedDevelopmentClassificationToBundle(t *testing.T) {
	result, err := Route(Options{
		TaxonomyPath:         shippedTaxonomyPath(t),
		SelectorRegistryPath: shippedRegistryPath(t),
		TaskClass:            "development",
		ClassificationReason: "KAH classified WFLOW-007 as development.",
		SelectedSpine:        "development_full",
		ProjectHasTealLane:   boolPtr(false),
		UIUXChange:           boolPtr(false),
		TealSkipReason:       "No UI/UX surface in this project/task.",
		RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Status != "bundle_route_matched" || result.SelectedBundle != "development_full" || result.WorkflowID != "development_full" {
		t.Fatalf("unexpected route result: %+v", result)
	}
	if result.WorkflowPath != ".kkachi/workflows/development_full.yaml" || result.DirectKAHStateWrite {
		t.Fatalf("unexpected workflow/no-write evidence: %+v", result)
	}
	if result.WorkPath != "A_development_execution" || result.WorkMode != "standard" || result.ExecutionMode != "production_write" {
		t.Fatalf("missing taxonomy defaults: %+v", result)
	}
	if result.Taxonomy.Checksum == "" || result.SelectorRegistry.Checksum == "" {
		t.Fatalf("missing source checksums: %+v", result)
	}
	if !reflect.DeepEqual(result.RequiredCapabilities, []string{"task_dag_schema_validation", "workflow_instance_state"}) {
		t.Fatalf("required capabilities = %+v", result.RequiredCapabilities)
	}
	if len(result.SkippedPhaseReasons) != 0 {
		t.Fatalf("development should not inherit skipped phases: %+v", result.SkippedPhaseReasons)
	}
}

func TestRouteAliasAndSkippedPhaseReasons(t *testing.T) {
	result, err := Route(Options{
		TaxonomyPath:         shippedTaxonomyPath(t),
		SelectorRegistryPath: shippedRegistryPath(t),
		TaskClass:            "investigation",
		ClassificationReason: "Read-only evidence collection.",
		ProjectHasTealLane:   boolPtr(false),
		UIUXChange:           boolPtr(false),
		TealSkipReason:       "No UI/UX surface in this project/task.",
		RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.TaskClass != "research_evidence" || result.SelectedBundle != "research_evidence_light" {
		t.Fatalf("unexpected alias route: %+v", result)
	}
	if result.SkippedPhaseReasons["implement"] != "skipped_by_default_for_research_evidence" {
		t.Fatalf("missing skipped phase reason: %+v", result.SkippedPhaseReasons)
	}
}

func TestRouteUsesStandardBundleCapabilityDefaults(t *testing.T) {
	result, err := Route(Options{
		TaxonomyPath:         shippedTaxonomyPath(t),
		SelectorRegistryPath: shippedRegistryPath(t),
		TaskClass:            "development",
		ClassificationReason: "Classified metadata omitted explicit capability flags.",
		ProjectHasTealLane:   boolPtr(false),
		UIUXChange:           boolPtr(false),
		TealSkipReason:       "No UI/UX surface in this project/task.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.SelectedBundle != "development_full" {
		t.Fatalf("expected standard capability defaults to route development, got %+v", result)
	}
	if !reflect.DeepEqual(result.RequiredCapabilities, []string{"task_dag_schema_validation", "workflow_instance_state"}) {
		t.Fatalf("required capability defaults = %+v", result.RequiredCapabilities)
	}
}

func TestRouteDerivesTealApplicabilityFromExplicitFacts(t *testing.T) {
	result, err := Route(Options{
		TaxonomyPath:         shippedTaxonomyPath(t),
		SelectorRegistryPath: shippedRegistryPath(t),
		TaskClass:            "development",
		ClassificationReason: "UI-bearing project work changes a registered Teal surface.",
		SelectedSpine:        "development_full",
		ProjectHasTealLane:   boolPtr(true),
		UIUXChange:           boolPtr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || !result.TealApplicability.TealRequired {
		t.Fatalf("expected derived teal_required=true route, got %+v", result)
	}
	if !result.TealApplicability.ProjectHasTealLane || !result.TealApplicability.UIUXChange {
		t.Fatalf("route did not preserve Teal input facts: %+v", result.TealApplicability)
	}
	if result.TealApplicability.Derivation != "project_has_teal_lane && ui_ux_change" {
		t.Fatalf("missing derivation evidence: %+v", result.TealApplicability)
	}
	if len(result.TealApplicability.RequiredWhenTealRequired) != 2 ||
		result.TealApplicability.RequiredWhenTealRequired[0] != "DESIGN_PLAN_GATE" ||
		result.TealApplicability.RequiredWhenTealRequired[1] != "DESIGN_FIDELITY_REVIEW" {
		t.Fatalf("missing required Teal verdict evidence: %+v", result.TealApplicability)
	}
}

func TestRouteRecordsNonUITealSkipAndFailsClosedForMissingFacts(t *testing.T) {
	result, err := Route(Options{
		TaxonomyPath:         shippedTaxonomyPath(t),
		SelectorRegistryPath: shippedRegistryPath(t),
		TaskClass:            "development",
		ClassificationReason: "Kkachi source work has no UI surface.",
		SelectedSpine:        "development_full",
		ProjectHasTealLane:   boolPtr(false),
		UIUXChange:           boolPtr(false),
		TealSkipReason:       "No UI/UX surface in this project/task.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.TealApplicability.TealRequired || result.TealApplicability.TealSkipReason == "" {
		t.Fatalf("expected non-UI skip route, got %+v", result)
	}

	missingFacts, err := Route(Options{
		TaxonomyPath:         shippedTaxonomyPath(t),
		SelectorRegistryPath: shippedRegistryPath(t),
		TaskClass:            "development",
		ClassificationReason: "missing Teal facts must fail closed.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if missingFacts.OK || missingFacts.Status != "teal_applicability_required" {
		t.Fatalf("expected missing Teal facts blocker, got %+v", missingFacts)
	}

	missingSkip, err := Route(Options{
		TaxonomyPath:         shippedTaxonomyPath(t),
		SelectorRegistryPath: shippedRegistryPath(t),
		TaskClass:            "development",
		ClassificationReason: "false Teal input without skip reason must fail closed.",
		ProjectHasTealLane:   boolPtr(false),
		UIUXChange:           boolPtr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if missingSkip.OK || missingSkip.Status != "teal_skip_reason_required" {
		t.Fatalf("expected missing skip reason blocker, got %+v", missingSkip)
	}
}

func TestRouteFailsClosedForMissingAndUnsupportedClassification(t *testing.T) {
	cases := map[string]Options{
		"missing class": {
			TaxonomyPath:         shippedTaxonomyPath(t),
			SelectorRegistryPath: shippedRegistryPath(t),
			ClassificationReason: "reason",
		},
		"missing reason": {
			TaxonomyPath:         shippedTaxonomyPath(t),
			SelectorRegistryPath: shippedRegistryPath(t),
			TaskClass:            "development",
		},
		"unsupported class": {
			TaxonomyPath:         shippedTaxonomyPath(t),
			SelectorRegistryPath: shippedRegistryPath(t),
			TaskClass:            "security",
			ClassificationReason: "reason",
			ProjectHasTealLane:   boolPtr(false),
			UIUXChange:           boolPtr(false),
			TealSkipReason:       "No UI/UX surface in this project/task.",
		},
	}
	want := map[string]string{
		"missing class":     "classification_required_input_missing",
		"missing reason":    "classification_reason_missing",
		"unsupported class": "classification_class_unsupported",
	}
	for name, opts := range cases {
		t.Run(name, func(t *testing.T) {
			result, err := Route(opts)
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Status != want[name] || result.DirectKAHStateWrite {
				t.Fatalf("expected fail-closed %s, got %+v", want[name], result)
			}
			if result.WorkflowID != "" || result.WorkflowPath != "" {
				t.Fatalf("failed route must not select workflow: %+v", result)
			}
		})
	}
}

func TestRouteFailsClosedForUnreadableAndUnsupportedSources(t *testing.T) {
	dir := t.TempDir()
	unsupportedTaxonomy := writeFile(t, dir, "unsupported-taxonomy.yaml", `version: 9.9.9
task_classes:
  development:
    default_work_path: A_development_execution
    default_work_mode: standard
    default_execution_mode: production_write
    default_spine: development_full
    required_phases: [task_contract]
`)
	unsupportedRegistry := writeFile(t, dir, "unsupported-registry.yaml", strings.Replace(validRouteRegistry("development_full", "development"), "kas-task-dag-workflow-registry/v1", "unsupported/v1", 1))

	cases := map[string]struct {
		opts Options
		want string
	}{
		"taxonomy unreadable": {
			opts: withNonUITeal(Options{TaxonomyPath: filepath.Join(dir, "missing-taxonomy.yaml"), SelectorRegistryPath: shippedRegistryPath(t), TaskClass: "development", ClassificationReason: "reason"}),
			want: "taxonomy_unreadable",
		},
		"taxonomy unsupported": {
			opts: withNonUITeal(Options{TaxonomyPath: unsupportedTaxonomy, SelectorRegistryPath: shippedRegistryPath(t), TaskClass: "development", ClassificationReason: "reason"}),
			want: "taxonomy_schema_unsupported",
		},
		"registry unreadable": {
			opts: withNonUITeal(Options{TaxonomyPath: shippedTaxonomyPath(t), SelectorRegistryPath: filepath.Join(dir, "missing-registry.yaml"), TaskClass: "development", ClassificationReason: "reason"}),
			want: "bundle_registry_unreadable",
		},
		"registry unsupported": {
			opts: withNonUITeal(Options{TaxonomyPath: shippedTaxonomyPath(t), SelectorRegistryPath: unsupportedRegistry, TaskClass: "development", ClassificationReason: "reason"}),
			want: "bundle_registry_schema_unsupported",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, err := Route(tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Status != tc.want || result.DirectKAHStateWrite {
				t.Fatalf("expected %s, got %+v", tc.want, result)
			}
			if result.WorkflowID != "" || result.SelectedBundle != "" {
				t.Fatalf("failed source route must not select bundle: %+v", result)
			}
		})
	}
}

func TestRouteFailsClosedForNoMatchAmbiguousAndSelectedSpineMismatch(t *testing.T) {
	taxonomy := writeFile(t, t.TempDir(), "taxonomy.yaml", `version: 0.1.0
task_classes:
  development:
    default_work_path: A_development_execution
    default_work_mode: standard
    default_execution_mode: production_write
    default_spine: development_full
    required_phases: [task_contract]
  docs_only:
    default_work_path: B_discovery_shaping
    default_work_mode: light
    default_execution_mode: docs_only
    default_spine: docs_only_light
    required_phases: [task_contract]
`)
	registryDir := t.TempDir()
	noMatchRegistry := writeFile(t, registryDir, "no-match.yaml", validRouteRegistry("docs_only_light", "docs_only"))
	result, err := Route(withNonUITeal(Options{TaxonomyPath: taxonomy, SelectorRegistryPath: noMatchRegistry, TaskClass: "development", ClassificationReason: "reason", RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}}))
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "bundle_no_match" {
		t.Fatalf("expected no match, got %+v", result)
	}

	ambiguousRegistry := writeFile(t, registryDir, "ambiguous.yaml", validRouteRegistry("development_full", "development")+strings.ReplaceAll(validRouteRegistry("development_alt", "development"), "version: kas-task-dag-workflow-registry/v1\n", ""))
	result, err = Route(withNonUITeal(Options{TaxonomyPath: taxonomy, SelectorRegistryPath: ambiguousRegistry, TaskClass: "development", ClassificationReason: "reason", RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}}))
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "bundle_ambiguous" || len(result.SelectorMatch.CandidateIDs) != 2 {
		t.Fatalf("expected ambiguity, got %+v", result)
	}

	result, err = Route(withNonUITeal(Options{TaxonomyPath: taxonomy, SelectorRegistryPath: ambiguousRegistry, TaskClass: "docs_only", ClassificationReason: "reason", SelectedSpine: "development_full", RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}}))
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "bundle_no_match" {
		t.Fatalf("expected no match before mismatch for docs_only, got %+v", result)
	}

	singleRegistry := writeFile(t, registryDir, "single.yaml", validRouteRegistry("development_full", "development"))
	result, err = Route(withNonUITeal(Options{TaxonomyPath: taxonomy, SelectorRegistryPath: singleRegistry, TaskClass: "development", ClassificationReason: "reason", SelectedSpine: "other_bundle", RequiredCapabilities: []string{"task_dag_schema_validation", "workflow_instance_state"}}))
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Status != "bundle_selected_mismatch" {
		t.Fatalf("expected selected-spine mismatch, got %+v", result)
	}
}

func shippedTaxonomyPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "registries", "task-taxonomy.yaml")
}

func shippedRegistryPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "registries", "task-dag-workflow-registry.yaml")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func writeFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func validRouteRegistry(workflowID string, taskClass string) string {
	return `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: ` + workflowID + `
    workflow_path: .kkachi/workflows/` + workflowID + `.yaml
    selector:
      task_classes: [` + taskClass + `]
      labels_any: []
      labels_all: []
      changed_surfaces_any: []
      risk_levels: []
      required_agents_all: []
      required_capabilities_all: [task_dag_schema_validation, workflow_instance_state]
    fallback_policy: none_fail_closed
node_contracts:
  - workflow_id: ` + workflowID + `
    node_id: setup
    task_class: ` + taskClass + `
    owner_role: implementer_backend
    execution_lane: direct_kas_skill
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/setup.md]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: kah_workflow_node_evidence
    completion_authority: kah_only
    direct_kah_state_write: false
`
}

func boolPtr(value bool) *bool {
	return &value
}

func withNonUITeal(opts Options) Options {
	opts.ProjectHasTealLane = boolPtr(false)
	opts.UIUXChange = boolPtr(false)
	opts.TealSkipReason = "No UI/UX surface in this project/task."
	return opts
}
