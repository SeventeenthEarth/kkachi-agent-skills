package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func buildBinary(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	binary := filepath.Join(t.TempDir(), "kkachi-hermes-skills")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/kkachi-hermes-skills")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}
	return binary
}

func TestRealRepoListAndInstallDryRunDoNotWriteProfile(t *testing.T) {
	root := repoRoot(t)
	binary := buildBinary(t)
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	list := exec.Command(binary, "list", "--repo", root, "--profile", "e2e", "--profile-root", profileRoot, "--json")
	list.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	listOut, err := list.CombinedOutput()
	if err != nil {
		t.Fatalf("list failed: %v\n%s", err, listOut)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("list created profile root: %v", err)
	}
	var listPayload map[string]any
	if err := json.Unmarshal(listOut, &listPayload); err != nil {
		t.Fatal(err)
	}
	if listPayload["ok"] != true || len(listPayload["packs"].([]any)) == 0 {
		t.Fatalf("unexpected list payload: %+v", listPayload)
	}

	install := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--dry-run", "--profile-root", profileRoot, "--json")
	install.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	installOut, err := install.CombinedOutput()
	if err != nil {
		t.Fatalf("install failed: %v\n%s", err, installOut)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("install dry-run created profile root: %v", err)
	}
	var installPayload map[string]any
	if err := json.Unmarshal(installOut, &installPayload); err != nil {
		t.Fatal(err)
	}
	summary := installPayload["summary"].(map[string]any)
	counts := summary["counts_by_action"].(map[string]any)
	if installPayload["ok"] != true || installPayload["mode"] != "dry_run" || counts["create"].(float64) != summary["total_files"].(float64) {
		t.Fatalf("unexpected install payload: %+v", installPayload)
	}
}

func TestRealRepoApprovedCopyWritesTempProfileOnly(t *testing.T) {
	root := repoRoot(t)
	binary := buildBinary(t)
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	dryRun := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--dry-run", "--profile-root", profileRoot, "--json")
	dryRun.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	dryRunOut, err := dryRun.CombinedOutput()
	if err != nil {
		t.Fatalf("dry-run failed: %v\n%s", err, dryRunOut)
	}
	var dryRunPayload map[string]any
	if err := json.Unmarshal(dryRunOut, &dryRunPayload); err != nil {
		t.Fatal(err)
	}
	evidence := dryRunPayload["approval_request"].(map[string]any)["evidence_ref"].(string)

	approve := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--approve", evidence, "--profile-root", profileRoot, "--json")
	approve.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	approveOut, err := approve.CombinedOutput()
	if err != nil {
		t.Fatalf("approve failed: %v\n%s", err, approveOut)
	}
	var approvePayload map[string]any
	if err := json.Unmarshal(approveOut, &approvePayload); err != nil {
		t.Fatal(err)
	}
	if approvePayload["ok"] != true || approvePayload["mode"] != "approved_copy" {
		t.Fatalf("unexpected approve payload: %+v", approvePayload)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "kkachi-plan", "SKILL.md")); err != nil {
		t.Fatalf("approved copy did not write expected skill: %v", err)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")); err != nil {
		t.Fatalf("approved copy did not write manifest: %v", err)
	}
}

func TestRealRepoDoctorReportsApprovedTempProfileReadOnly(t *testing.T) {
	root := repoRoot(t)
	binary := buildBinary(t)
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	dryRun := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--dry-run", "--profile-root", profileRoot, "--json")
	dryRun.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	dryRunOut, err := dryRun.CombinedOutput()
	if err != nil {
		t.Fatalf("dry-run failed: %v\n%s", err, dryRunOut)
	}
	var dryRunPayload map[string]any
	if err := json.Unmarshal(dryRunOut, &dryRunPayload); err != nil {
		t.Fatal(err)
	}
	evidence := dryRunPayload["approval_request"].(map[string]any)["evidence_ref"].(string)

	approve := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--approve", evidence, "--profile-root", profileRoot, "--json")
	approve.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	approveOut, err := approve.CombinedOutput()
	if err != nil {
		t.Fatalf("approve failed: %v\n%s", err, approveOut)
	}
	before, err := os.ReadFile(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	doctorJSON := exec.Command(binary, "doctor", "--repo", root, "--profile", "e2e", "--profile-root", profileRoot, "--json")
	doctorJSON.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	doctorJSONOut, err := doctorJSON.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor json failed: %v\n%s", err, doctorJSONOut)
	}
	var payload map[string]any
	if err := json.Unmarshal(doctorJSONOut, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "doctor" || payload["manifest"].(map[string]any)["state"] != "ok" {
		t.Fatalf("unexpected doctor payload: %+v", payload)
	}
	kab := payload["kab"].(map[string]any)
	if kab["required_for_minimum_cli"] != false || kab["required_for_execution_runtime"] != true {
		t.Fatalf("unexpected KAB payload: %+v", payload)
	}

	doctorHuman := exec.Command(binary, "doctor", "--repo", root, "--profile", "e2e", "--profile-root", profileRoot, "--no-color")
	doctorHuman.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	doctorHumanOut, err := doctorHuman.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor human failed: %v\n%s", err, doctorHumanOut)
	}
	if !strings.Contains(string(doctorHumanOut), "상태:") || !strings.Contains(string(doctorHumanOut), "건강") {
		t.Fatalf("unexpected doctor human output: %s", doctorHumanOut)
	}
	after, err := os.ReadFile(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("doctor rewrote the install manifest")
	}
}

func TestSyncProjectKASValidateStateReadOnly(t *testing.T) {
	binary := buildBinary(t)
	dir := t.TempDir()
	statePath := filepath.Join(dir, "kas-project-state.yaml")
	if err := os.WriteFile(statePath, []byte(e2eValidKASState()), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binary, "sync-project-kas", "--profile", "hwangchung", "--project", "kan-plugin", "--state", statePath, "--dry-run", "--json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync-project-kas failed: %v\n%s", err, out)
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["yaml_state_path"] != statePath || payload["legacy_marker_path"] != filepath.Join(dir, "kab-adoption-stage.md") {
		t.Fatalf("unexpected sync-project-kas payload: %+v", payload)
	}
	after, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("sync-project-kas rewrote state file")
	}

	missingDryRun := exec.Command(binary, "sync-project-kas", "--profile", "hwangchung", "--project", "kan-plugin", "--state", statePath, "--json")
	missingOut, err := missingDryRun.CombinedOutput()
	if err == nil {
		t.Fatalf("expected missing --dry-run failure, got success: %s", missingOut)
	}
	var missingPayload map[string]any
	if err := json.Unmarshal(missingOut, &missingPayload); err != nil {
		t.Fatal(err)
	}
	if missingPayload["ok"] != false || missingPayload["diagnostics"].([]any)[0].(map[string]any)["code"] != "sync_project_kas_requires_dry_run" {
		t.Fatalf("unexpected missing dry-run payload: %+v", missingPayload)
	}
}

func e2eValidKASState() string {
	checksum := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	return `version: "0.1"
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
    source_checksum: "` + checksum + `"
    project_checksum: "` + checksum + `"
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
  require_llm_merge_when:
    - "project_skill_mapping_exists"
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
}
