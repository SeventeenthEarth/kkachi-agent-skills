package docscontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type design006Scenario struct {
	ID                         string   `json:"id"`
	Project                    string   `json:"project"`
	ProjectHasTealLane         bool     `json:"project_has_teal_lane"`
	UIUXChange                 bool     `json:"ui_ux_change"`
	TealRequired               bool     `json:"teal_required"`
	TealSkipReason             string   `json:"teal_skip_reason"`
	RequiredWhenTealRequired   []string `json:"required_when_teal_required"`
	ExpectedMaterializedNodes  []string `json:"expected_materialized_nodes"`
	OrdinaryReviewIsSubstitute bool     `json:"ordinary_review_is_substitute"`
	MARReviewIsSubstitute      bool     `json:"mar_review_is_substitute"`
	BackendEvidenceSubstitute  bool     `json:"backend_evidence_is_substitute"`
	HelperNotesAreSubstitute   bool     `json:"helper_notes_are_substitute"`
}

func TestDesign006GoldenCompatibilityScenarios(t *testing.T) {
	scenarios := readDesign006Scenarios(t)
	if len(scenarios) != 4 {
		t.Fatalf("scenario count = %d, want 4", len(scenarios))
	}
	seen := map[string]design006Scenario{}
	for _, scenario := range scenarios {
		seen[scenario.ID] = scenario
		if scenario.TealRequired != (scenario.ProjectHasTealLane && scenario.UIUXChange) {
			t.Fatalf("%s derives teal_required=%t from project_has_teal_lane=%t ui_ux_change=%t", scenario.ID, scenario.TealRequired, scenario.ProjectHasTealLane, scenario.UIUXChange)
		}
		if !scenario.TealRequired && scenario.TealSkipReason == "" {
			t.Fatalf("%s missing concrete skip reason", scenario.ID)
		}
		if scenario.TealRequired && scenario.TealSkipReason != "" {
			t.Fatalf("%s has teal_required=true with non-empty skip reason %q", scenario.ID, scenario.TealSkipReason)
		}
		requireStringSetExact(t, scenario.ID, scenario.RequiredWhenTealRequired, []string{"DESIGN_PLAN_GATE", "DESIGN_FIDELITY_REVIEW"})
		if scenario.TealRequired {
			requireStringSetExact(t, scenario.ID, scenario.ExpectedMaterializedNodes, []string{"DESIGN_PLAN_GATE", "DESIGN_FIDELITY_REVIEW"})
		} else {
			requireStringSetExact(t, scenario.ID, scenario.ExpectedMaterializedNodes, []string{})
		}
		if scenario.OrdinaryReviewIsSubstitute || scenario.MARReviewIsSubstitute || scenario.BackendEvidenceSubstitute || scenario.HelperNotesAreSubstitute {
			t.Fatalf("%s permits forbidden Teal verdict substitution: %+v", scenario.ID, scenario)
		}
	}
	for _, id := range []string{"kkachi_non_ui_skip", "kkachi_teal_lane_non_ui_skip", "sudal_ui_required", "doksuri_ui_required"} {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing DESIGN-006 scenario %s", id)
		}
	}
	if nodes := seen["kkachi_non_ui_skip"].ExpectedMaterializedNodes; len(nodes) != 0 {
		t.Fatalf("kkachi_non_ui_skip nodes = %#v, want no Teal materialized nodes", nodes)
	}
}

func readDesign006Scenarios(t *testing.T) []design006Scenario {
	t.Helper()
	path := filepath.Join(repoRoot(t), "docs", "examples", "design006-teal-compatibility-scenarios.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DESIGN-006 scenarios: %v", err)
	}
	var payload struct {
		Version   string              `json:"version"`
		Scenarios []design006Scenario `json:"scenarios"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode DESIGN-006 scenarios: %v", err)
	}
	if payload.Version != "design006.v1" {
		t.Fatalf("version = %q, want design006.v1", payload.Version)
	}
	return payload.Scenarios
}

func requireStringSetExact(t *testing.T, id string, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s values = %#v, want exactly %#v", id, got, want)
	}
	set := map[string]bool{}
	for _, value := range got {
		set[value] = true
	}
	for _, value := range want {
		if !set[value] {
			t.Fatalf("%s values = %#v, want exactly %#v", id, got, want)
		}
	}
}
