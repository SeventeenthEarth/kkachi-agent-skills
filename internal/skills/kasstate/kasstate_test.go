package kasstate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/install"
)

const goodChecksum = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func validStateYAML(overrides ...string) string {
	body := `version: "0.1"

project:
  id: "kan-plugin"
  repo: "kkachi-agent-network-plugin"
  kas_suite: "kan-plugin"
  profile: "hwangchung"

kab_adoption_stage:
  numeric: 1
  canonical: "stage1_direct_codex_app_server_baseline"
  selection_source: "approved_project_policy"
  selected_at: "2026-06-06T00:00:00Z"
  approval_evidence: "not_applicable"
  stage2_activation: false

upstream_kas:
  repo: "kkachi-hermes-skills"
  remote: "github.com/SeventeenthEarth/kkachi-hermes-skills"
  commit: "0123456789abcdef0123456789abcdef01234567"
  dirty: false
  synced_at: "2026-06-06T00:00:00Z"
  sync_task: "KASUPD-001"

pack_baselines:
  - upstream_pack: "kkachi-plan"
    project_skill: "kan-plugin-plan"
    source_checksum: "` + goodChecksum + `"
    project_checksum: "` + goodChecksum + `"
    merge_mode: "semantic_port"

overlay_policy:
  local_overlay_allowed: true
  preserve_project_authority: true
  preserve_project_roadmap_ids: true
  preserve_project_test_commands: true
  preserve_role_labels: true
  overwrite_mode: "never_without_review"

update_policy:
  default_mode: "dry_run_then_semantic_merge"
  auto_apply_when:
    - "target_file_missing"
    - "local_unchanged_since_baseline_and_upstream_changed"
  require_llm_merge_when:
    - "local_changed_since_baseline_and_upstream_changed"
    - "project_skill_mapping_exists"
    - "policy_text_requires_project_specific_translation"
  fail_closed_when:
    - "state_file_missing"
    - "state_schema_invalid"
    - "stage_unsupported"
    - "upstream_commit_unknown"
    - "checksum_mismatch_without_baseline"
    - "auth_token_gateway_or_provider_mutation_detected"

evidence_posture:
  not_kab_runtime_evidence: true
  not_stage2_activation_by_itself: true
  missing_or_unreadable_fails_to_stage1_claims: true
`
	for _, override := range overrides {
		oldNew := strings.SplitN(override, "=>", 2)
		if len(oldNew) != 2 {
			panic("override must use old=>new")
		}
		body = strings.Replace(body, oldNew[0], oldNew[1], 1)
	}
	return body
}

func writeState(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "kas-project-state.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBuildValidYAMLRecognizesStaticStage(t *testing.T) {
	path := writeState(t, validStateYAML())
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: path, DryRun: true})
	if !result.OK || result.StateSource != "yaml" || result.ReadSurfaces.YAML.State != "valid" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.EffectiveStageClaim.Numeric != 1 || result.EffectiveStageClaim.Canonical != install.KABStage1Canonical {
		t.Fatalf("unexpected stage claim: %+v", result.EffectiveStageClaim)
	}
	if result.EffectiveStageClaim.KABExecutionClaimAllowed || result.EffectiveStageClaim.FailClosedToStage1 {
		t.Fatalf("unexpected execution claim posture: %+v", result.EffectiveStageClaim)
	}
	if result.Validation.PackBaselineCount != 1 || result.WriteTargetAfterApprovedSync != "yaml_state_path" {
		t.Fatalf("unexpected validation/write target: %+v", result)
	}
}

func TestBuildValidStage2YAMLDoesNotClaimKABExecution(t *testing.T) {
	path := writeState(t, validStateYAML(
		"numeric: 1=>numeric: 2",
		"canonical: \"stage1_direct_codex_app_server_baseline\"=>canonical: \"stage2_kab_codex_first\"",
	))
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: path, DryRun: true})
	if !result.OK {
		t.Fatalf("unexpected invalid result: %+v", result)
	}
	if result.EffectiveStageClaim.Numeric != 2 || result.EffectiveStageClaim.KABExecutionClaimAllowed {
		t.Fatalf("Stage 2 YAML must not become KAB execution evidence: %+v", result.EffectiveStageClaim)
	}
}

func TestBuildInvalidYAMLAndPolicyFailuresFailClosed(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		code string
	}{
		{"syntax", strings.Replace(validStateYAML(), "project:", "project", 1), "state_schema_invalid"},
		{"stage3", validStateYAML("numeric: 1=>numeric: 3", "canonical: \"stage1_direct_codex_app_server_baseline\"=>canonical: \"stage3_kab_backend_selected\""), "stage_unsupported"},
		{"badCommit", validStateYAML("commit: \"0123456789abcdef0123456789abcdef01234567\"=>commit: \"HEAD\""), "upstream_commit_unknown"},
		{"badChecksum", validStateYAML(goodChecksum + "\"=>sha256:nothex\""), "checksum_mismatch_without_baseline"},
		{"overlayFalse", validStateYAML("preserve_project_authority: true=>preserve_project_authority: false"), "overlay_policy_invalid"},
		{"evidenceFalse", validStateYAML("not_stage2_activation_by_itself: true=>not_stage2_activation_by_itself: false"), "evidence_posture_invalid"},
		{"suspiciousValue", validStateYAML("approval_evidence: \"not_applicable\"=>approval_evidence: \"token=abc\""), "auth_token_gateway_or_provider_mutation_detected"},
		{"unsupportedYAMLFeature", strings.Replace(validStateYAML(), "repo: \"kkachi-agent-network-plugin\"", "repo: |\n    kkachi-agent-network-plugin", 1), "state_schema_invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: writeState(t, tt.yaml), DryRun: true})
			if result.OK || !result.EffectiveStageClaim.FailClosedToStage1 {
				t.Fatalf("expected fail-closed invalid result: %+v", result)
			}
			if !containsCode(result, tt.code) {
				t.Fatalf("missing diagnostic %s in %+v", tt.code, result.Validation.Diagnostics)
			}
		})
	}
}

func TestInvalidYAMLWithCompatibleLegacyMarkerStillFailsClosed(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "kas-project-state.yaml")
	if err := os.WriteFile(statePath, []byte(validStateYAML("commit: \"0123456789abcdef0123456789abcdef01234567\"=>commit: \"HEAD\"")), 0o644); err != nil {
		t.Fatal(err)
	}
	stage, err := install.ResolveKABAdoptionStage(install.StageSelectionInput{Numeric: "2"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kab-adoption-stage.md"), []byte(install.KABAdoptionStageMarkerContent(stage)), 0o644); err != nil {
		t.Fatal(err)
	}

	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: statePath, DryRun: true})
	if result.OK || result.StateSource != "legacy_marker_only" || !result.EffectiveStageClaim.FailClosedToStage1 {
		t.Fatalf("invalid YAML must remain fail-closed even when a legacy marker is readable: %+v", result)
	}
	if result.ReadSurfaces.YAML.State != "invalid" || result.ReadSurfaces.LegacyMarker.State != "present_compatible" {
		t.Fatalf("unexpected read-surface states: %+v", result.ReadSurfaces)
	}
	if result.EffectiveStageClaim.Numeric != 1 || result.EffectiveStageClaim.Source != "fail_closed" {
		t.Fatalf("legacy marker must not upgrade the effective stage claim: %+v", result.EffectiveStageClaim)
	}
}

func TestMissingYAMLWithCompatibleLegacyMarkerDoesNotUpgradeState(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "kas-project-state.yaml")
	stage, err := install.ResolveKABAdoptionStage(install.StageSelectionInput{Numeric: "2"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kab-adoption-stage.md"), []byte(install.KABAdoptionStageMarkerContent(stage)), 0o644); err != nil {
		t.Fatal(err)
	}
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: statePath, DryRun: true})
	if result.OK || result.StateSource != "legacy_marker_only" {
		t.Fatalf("legacy marker must not validate missing YAML: %+v", result)
	}
	if result.ReadSurfaces.LegacyMarker.State != "present_compatible" || result.EffectiveStageClaim.Numeric != 1 || !result.EffectiveStageClaim.FailClosedToStage1 {
		t.Fatalf("unexpected legacy compatibility result: %+v", result)
	}
}

func TestRenderHumanIncludesPathsAndDiagnostics(t *testing.T) {
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: filepath.Join(t.TempDir(), "kas-project-state.yaml"), DryRun: true})
	human := RenderHuman(result)
	for _, want := range []string{
		"검증 실패",
		"YAML: missing",
		"legacy marker: missing",
		"effective stage: 1",
		"state_file_missing",
		"write target after approved sync: yaml_state_path",
	} {
		if !strings.Contains(human, want) {
			t.Fatalf("human output missing %q:\n%s", want, human)
		}
	}
}

func containsCode(result Result, code string) bool {
	for _, diagnostic := range result.Validation.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}
