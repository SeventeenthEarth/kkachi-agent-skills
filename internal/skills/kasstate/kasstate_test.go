package kasstate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
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
	repo, projectRoot, path := setupUnchangedClassificationFixture(t, validStateYAML)
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: path, DryRun: true, RepoPath: repo, ProjectRoot: projectRoot})
	if !result.OK || result.StateSource != "yaml" || result.ReadSurfaces.YAML.State != "valid" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Mode != "dry_run_classification" || result.Summary.NoActionCount != 1 || len(result.UnchangedMappings) != 1 {
		t.Fatalf("expected unchanged dry-run classification: %+v", result)
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
	repo, projectRoot, path := setupUnchangedClassificationFixture(t, func(overrides ...string) string {
		return validStateYAML(append([]string{
			"numeric: 1=>numeric: 2",
			"canonical: \"stage1_direct_codex_app_server_baseline\"=>canonical: \"stage2_kab_codex_first\"",
		}, overrides...)...)
	})
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: path, DryRun: true, RepoPath: repo, ProjectRoot: projectRoot})
	if !result.OK {
		t.Fatalf("unexpected invalid result: %+v", result)
	}
	if result.EffectiveStageClaim.Numeric != 2 || result.EffectiveStageClaim.KABExecutionClaimAllowed {
		t.Fatalf("Stage 2 YAML must not become KAB execution evidence: %+v", result.EffectiveStageClaim)
	}
}

func TestDryRunClassificationMatrixAndSemanticPackets(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "upstream")
	projectRoot := filepath.Join(dir, "project")
	gitInit(t, repo)
	for _, pack := range []string{"auto-pack", "local-pack", "semantic-pack", "removed-pack", "unchanged-pack"} {
		writeSkillPack(t, filepath.Join(repo, "skills", pack), pack, "baseline")
		writeSkillPack(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-"+pack), pack, "baseline")
	}
	gitCommitAll(t, repo, "baseline")
	baselineCommit := gitRevParse(t, repo)
	baselines := []map[string]string{}
	for _, pack := range []string{"auto-pack", "local-pack", "semantic-pack", "removed-pack", "unchanged-pack"} {
		sourceChecksum := checksumFor(t, filepath.Join(repo, "skills", pack))
		projectChecksum := checksumFor(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-"+pack))
		baselines = append(baselines, map[string]string{
			"upstream_pack":    pack,
			"project_skill":    "kan-plugin-" + pack,
			"source_checksum":  sourceChecksum,
			"project_checksum": projectChecksum,
			"merge_mode":       "semantic_port",
		})
	}
	writeSkillPack(t, filepath.Join(repo, "skills", "auto-pack"), "auto-pack", "upstream changed")
	writeSkillPack(t, filepath.Join(repo, "skills", "semantic-pack"), "semantic-pack", "upstream changed")
	if err := os.RemoveAll(filepath.Join(repo, "skills", "removed-pack")); err != nil {
		t.Fatal(err)
	}
	writeSkillPack(t, filepath.Join(repo, "skills", "new-pack"), "new-pack", "new upstream")
	gitCommitAll(t, repo, "current")
	writeSkillPack(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-local-pack"), "local-pack", "local changed")
	writeSkillPack(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-semantic-pack"), "semantic-pack", "local changed")
	statePath := writeProjectState(t, projectRoot, stateYAMLForBaselines(baselineCommit, baselines))

	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: statePath, DryRun: true, RepoPath: repo, ProjectRoot: projectRoot})
	if !result.OK {
		t.Fatalf("unexpected fail-closed result: %+v", result)
	}
	counts := result.Summary.CountsByClassification
	for _, want := range []string{"auto_copy_candidate", "local_only", "semantic_merge_required", "new_upstream_candidate", "removed_or_renamed_upstream"} {
		if counts[want] != 1 {
			t.Fatalf("expected one %s, got counts %+v classifications %+v", want, counts, result.Classifications)
		}
	}
	if result.Summary.NoActionCount != 1 || len(result.UnchangedMappings) != 1 {
		t.Fatalf("expected one unchanged mapping: %+v", result)
	}
	if result.Summary.SemanticPortPacketCount != 2 || len(result.SemanticPortPackets) != 2 {
		t.Fatalf("expected semantic packets for semantic/new candidates: %+v", result.SemanticPortPackets)
	}
	for _, packet := range result.SemanticPortPackets {
		for _, want := range []string{"no_write_statement", "state_sha256:", "Preservation constraints", "approved sync/apply/write-capable behavior is out of scope"} {
			if !strings.Contains(packet.Content, want) {
				t.Fatalf("packet missing %q:\n%s", want, packet.Content)
			}
		}
	}
}

func TestDryRunFailClosedForDirtySourceAndBaselineMismatch(t *testing.T) {
	repo, projectRoot, statePath := setupUnchangedClassificationFixture(t, validStateYAML)
	writeSkillPack(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "dirty")
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: statePath, DryRun: true, RepoPath: repo, ProjectRoot: projectRoot})
	if result.OK || result.Summary.CountsByClassification["fail_closed_conflict"] != 1 || !containsCode(result, "source_repo_dirty_requires_review") {
		t.Fatalf("dirty source must fail closed: %+v", result)
	}

	repo, projectRoot, statePath = setupUnchangedClassificationFixture(t, func(overrides ...string) string {
		return validStateYAML(append([]string{goodChecksum + "\"=>sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\""}, overrides...)...)
	})
	result = Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: statePath, DryRun: true, RepoPath: repo, ProjectRoot: projectRoot})
	if result.OK || result.Summary.CountsByClassification["fail_closed_conflict"] != 1 {
		t.Fatalf("baseline checksum mismatch must fail closed: %+v", result)
	}
	if !classificationHasCode(result, "baseline_source_checksum_mismatch") {
		t.Fatalf("missing baseline mismatch diagnostic: %+v", result.Classifications)
	}
}

func TestProjectSkillMissingAndAmbiguousFailClosed(t *testing.T) {
	repo, projectRoot, statePath := setupUnchangedClassificationFixture(t, validStateYAML)
	if err := os.RemoveAll(filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-plan")); err != nil {
		t.Fatal(err)
	}
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: statePath, DryRun: true, RepoPath: repo, ProjectRoot: projectRoot})
	if result.OK || !classificationHasCode(result, "project_skill_missing") {
		t.Fatalf("missing project skill must fail closed: %+v", result)
	}

	repo, projectRoot, statePath = setupUnchangedClassificationFixture(t, validStateYAML)
	writeSkillPack(t, filepath.Join(projectRoot, "skills", "kan-plugin-plan"), "kkachi-plan", "baseline")
	result = Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: statePath, DryRun: true, RepoPath: repo, ProjectRoot: projectRoot})
	if result.OK || !classificationHasCode(result, "project_skill_path_ambiguous") {
		t.Fatalf("ambiguous project skill must fail closed: %+v", result)
	}
}

func TestProjectRootInferenceRequiresCanonicalStatePath(t *testing.T) {
	path := writeState(t, validStateYAML(
		"numeric: 1=>numeric: 2",
		"canonical: \"stage1_direct_codex_app_server_baseline\"=>canonical: \"stage2_kab_codex_first\"",
	))
	result := Build(Options{Profile: "hwangchung", Project: "kan-plugin", StatePath: path, DryRun: true})
	if result.OK || result.ProjectRoot == nil || result.ProjectRoot.State != "ambiguous" {
		t.Fatalf("non-canonical state path must fail closed on project root ambiguity: %+v", result)
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

func classificationHasCode(result Result, code string) bool {
	for _, classification := range result.Classifications {
		for _, diagnostic := range classification.Diagnostics {
			if diagnostic.Code == code {
				return true
			}
		}
	}
	return false
}

func setupUnchangedClassificationFixture(t *testing.T, stateBuilder func(...string) string) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	repo := filepath.Join(dir, "upstream")
	projectRoot := filepath.Join(dir, "project")
	gitInit(t, repo)
	writeSkillPack(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "baseline")
	writeSkillPack(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-plan"), "kkachi-plan", "baseline")
	gitCommitAll(t, repo, "baseline")
	commit := gitRevParse(t, repo)
	sourceChecksum := checksumFor(t, filepath.Join(repo, "skills", "kkachi-plan"))
	projectChecksum := checksumFor(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-plan"))
	state := stateBuilder(
		"commit: \"0123456789abcdef0123456789abcdef01234567\"=>commit: \""+commit+"\"",
		goodChecksum+"\"=>"+sourceChecksum+"\"",
		goodChecksum+"\"=>"+projectChecksum+"\"",
	)
	return repo, projectRoot, writeProjectState(t, projectRoot, state)
}

func stateYAMLForBaselines(commit string, baselines []map[string]string) string {
	lines := []string{
		`version: "0.1"`,
		``,
		`project:`,
		`  id: "kan-plugin"`,
		`  repo: "kkachi-agent-network-plugin"`,
		`  kas_suite: "kan-plugin"`,
		`  profile: "hwangchung"`,
		``,
		`kab_adoption_stage:`,
		`  numeric: 1`,
		`  canonical: "stage1_direct_codex_app_server_baseline"`,
		`  selection_source: "approved_project_policy"`,
		`  selected_at: "2026-06-06T00:00:00Z"`,
		`  approval_evidence: "not_applicable"`,
		`  stage2_activation: false`,
		``,
		`upstream_kas:`,
		`  repo: "kkachi-hermes-skills"`,
		`  remote: "github.com/SeventeenthEarth/kkachi-hermes-skills"`,
		`  commit: "` + commit + `"`,
		`  dirty: false`,
		`  synced_at: "2026-06-06T00:00:00Z"`,
		`  sync_task: "KASUPD-001"`,
		``,
		`pack_baselines:`,
	}
	for _, baseline := range baselines {
		lines = append(lines,
			`  - upstream_pack: "`+baseline["upstream_pack"]+`"`,
			`    project_skill: "`+baseline["project_skill"]+`"`,
			`    source_checksum: "`+baseline["source_checksum"]+`"`,
			`    project_checksum: "`+baseline["project_checksum"]+`"`,
			`    merge_mode: "semantic_port"`,
		)
	}
	lines = append(lines,
		``,
		`overlay_policy:`,
		`  local_overlay_allowed: true`,
		`  preserve_project_authority: true`,
		`  preserve_project_roadmap_ids: true`,
		`  preserve_project_test_commands: true`,
		`  preserve_role_labels: true`,
		`  overwrite_mode: "never_without_review"`,
		``,
		`update_policy:`,
		`  default_mode: "dry_run_then_semantic_merge"`,
		`  auto_apply_when:`,
		`    - "target_file_missing"`,
		`    - "local_unchanged_since_baseline_and_upstream_changed"`,
		`  require_llm_merge_when:`,
		`    - "local_changed_since_baseline_and_upstream_changed"`,
		`    - "project_skill_mapping_exists"`,
		`    - "policy_text_requires_project_specific_translation"`,
		`  fail_closed_when:`,
		`    - "state_file_missing"`,
		`    - "state_schema_invalid"`,
		`    - "stage_unsupported"`,
		`    - "upstream_commit_unknown"`,
		`    - "checksum_mismatch_without_baseline"`,
		`    - "auth_token_gateway_or_provider_mutation_detected"`,
		``,
		`evidence_posture:`,
		`  not_kab_runtime_evidence: true`,
		`  not_stage2_activation_by_itself: true`,
		`  missing_or_unreadable_fails_to_stage1_claims: true`,
		``,
	)
	return strings.Join(lines, "\n")
}

func writeProjectState(t *testing.T, projectRoot string, content string) string {
	t.Helper()
	path := filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-kas", "references", "kas-project-state.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeSkillPack(t *testing.T, dir string, name string, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\n---\n# " + name + "\n\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func checksumFor(t *testing.T, dir string) string {
	t.Helper()
	checksum, err := discovery.ComputePackChecksum(dir)
	if err != nil {
		t.Fatal(err)
	}
	return "sha256:" + checksum
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
}

func gitCommitAll(t *testing.T, dir string, message string) {
	t.Helper()
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", message)
}

func gitRevParse(t *testing.T, dir string) string {
	t.Helper()
	out := runGit(t, dir, "rev-parse", "HEAD")
	return strings.TrimSpace(out)
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}
