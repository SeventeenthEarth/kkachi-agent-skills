package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
)

func writeCLITestSkill(t *testing.T, dir string, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRootHelpExitsZeroAndPrintsCommands(t *testing.T) {
	for _, arg := range []string{"--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Main([]string{arg}, &stdout, &stderr, nil)
			if code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			out := stdout.String()
			if !strings.Contains(out, "list") || !strings.Contains(out, "install") || !strings.Contains(out, "doctor") || !strings.Contains(out, "sync-project-kas") {
				t.Fatalf("root help did not list available commands: %q", out)
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected root help on stdout only, got stderr=%q", stderr.String())
			}
		})
	}
}

func TestSubcommandHelpExitsZero(t *testing.T) {
	for _, command := range []string{"list", "install", "doctor", "sync-project-kas"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Main([]string{command, "--help"}, &stdout, &stderr, nil)
			if code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			if !strings.Contains(stdout.String(), "Usage of "+command+":") {
				t.Fatalf("unexpected help output: %q", stdout.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected help on stdout only, got stderr=%q", stderr.String())
			}
		})
	}
}

func TestSyncProjectKASJSONAndFailClosedGuards(t *testing.T) {
	repo, projectRoot, statePath := setupCLISyncFixture(t)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"sync-project-kas", "--profile", "hwangchung", "--project", "kan-plugin", "--state", statePath, "--dry-run", "--json", "--repo", repo, "--project-root", projectRoot}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "sync-project-kas" || payload["yaml_state_path"] != statePath {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload["mode"] != "dry_run_classification" || payload["summary"].(map[string]any)["no_action_count"] != float64(1) {
		t.Fatalf("missing classification shape: %+v", payload)
	}
	if payload["legacy_marker_path"] != filepath.Join(filepath.Dir(statePath), "kab-adoption-stage.md") || payload["write_target_after_approved_sync"] != "yaml_state_path" {
		t.Fatalf("missing path/write target distinction: %+v", payload)
	}
	stage := payload["effective_stage_claim"].(map[string]any)
	if stage["kab_execution_claim_allowed"] != false || stage["source"] != "yaml" {
		t.Fatalf("unexpected stage claim: %+v", stage)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"sync-project-kas", "--profile", "hwangchung", "--project", "kan-plugin", "--state", statePath, "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing dry-run failure, got %d", code)
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["diagnostics"].([]any)[0].(map[string]any)["code"] != "sync_project_kas_requires_dry_run" {
		t.Fatalf("unexpected missing dry-run payload: %+v", payload)
	}

	before, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	dir := t.TempDir()
	badStatePath := filepath.Join(dir, "bad-state.yaml")
	if err := os.WriteFile(badStatePath, []byte(strings.Replace(cliValidKASState(), "commit: \"0123456789abcdef0123456789abcdef01234567\"", "commit: \"HEAD\"", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	code = Main([]string{"sync-project-kas", "--profile", "hwangchung", "--project", "kan-plugin", "--state", badStatePath, "--dry-run", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected invalid state failure, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["yaml_state_path"] != badStatePath || payload["legacy_marker_path"] != filepath.Join(dir, "kab-adoption-stage.md") {
		t.Fatalf("invalid JSON did not include both paths: %+v", payload)
	}
	stage = payload["effective_stage_claim"].(map[string]any)
	if stage["fail_closed_to_stage1"] != true || stage["source"] != "fail_closed" {
		t.Fatalf("invalid state did not fail closed to Stage 1: %+v", stage)
	}
	after, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("sync-project-kas dry-run rewrote state")
	}
}

func setupCLISyncFixture(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	repo := filepath.Join(dir, "upstream")
	projectRoot := filepath.Join(dir, "project")
	cliGitInit(t, repo)
	writeCLISkillPack(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "baseline")
	writeCLISkillPack(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-plan"), "kkachi-plan", "baseline")
	cliGitCommitAll(t, repo, "baseline")
	commit := strings.TrimSpace(cliRunGit(t, repo, "rev-parse", "HEAD"))
	sourceChecksum := cliChecksumFor(t, filepath.Join(repo, "skills", "kkachi-plan"))
	projectChecksum := cliChecksumFor(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-plan"))
	state := strings.Replace(cliValidKASState(), "commit: \"0123456789abcdef0123456789abcdef01234567\"", "commit: \""+commit+"\"", 1)
	state = strings.Replace(state, "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", sourceChecksum, 1)
	state = strings.Replace(state, "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", projectChecksum, 1)
	statePath := filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-kas", "references", "kas-project-state.yaml")
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}
	return repo, projectRoot, statePath
}

func writeCLISkillPack(t *testing.T, dir string, name string, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\n---\n# " + name + "\n\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func cliChecksumFor(t *testing.T, dir string) string {
	t.Helper()
	checksum, err := discovery.ComputePackChecksum(dir)
	if err != nil {
		t.Fatal(err)
	}
	return "sha256:" + checksum
}

func cliGitInit(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cliRunGit(t, dir, "init")
	cliRunGit(t, dir, "config", "user.email", "test@example.com")
	cliRunGit(t, dir, "config", "user.name", "Test User")
}

func cliGitCommitAll(t *testing.T, dir string, message string) {
	t.Helper()
	cliRunGit(t, dir, "add", "-A")
	cliRunGit(t, dir, "commit", "-m", message)
}

func cliRunGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func cliValidKASState() string {
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

func TestDoctorJSONHumanAndProfileRootGuard(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("dry-run code=%d stderr=%s", code, stderr.String())
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	evidence := dryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	code = Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--approve", evidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("approve code=%d stderr=%s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"doctor", "--repo", repo, "--profile", "demo", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("doctor code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["command"] != "doctor" || payload["ok"] != true || payload["manifest"].(map[string]any)["state"] != "ok" {
		t.Fatalf("unexpected doctor payload: %+v", payload)
	}
	if payload["provenance_contract_version"] != discovery.ProvenanceContractVersion || payload["provenance_audit"] == nil {
		t.Fatalf("missing doctor provenance payload: %+v", payload)
	}
	kab := payload["kab"].(map[string]any)
	if kab["required_for_minimum_cli"] != false || kab["required_for_execution_runtime"] != true {
		t.Fatalf("unexpected KAB payload: %+v", payload)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"doctor", "--repo", repo, "--profile", "demo", "--profile-root", profileRoot, "--no-color"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("doctor human code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "상태:") || !strings.Contains(stdout.String(), "건강") || !strings.Contains(stdout.String(), "KAB") {
		t.Fatalf("unexpected human output: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"doctor", "--repo", repo, "--profile", "demo", "--profile-root", filepath.Join(t.TempDir(), "profile"), "--json"}, &stdout, &stderr, map[string]string{})
	if code != 2 {
		t.Fatalf("expected guard failure, got %d", code)
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["diagnostics"].([]any)[0].(map[string]any)["code"] != "profile_root_override_rejected" {
		t.Fatalf("unexpected guard payload: %+v", payload)
	}
}

func TestListJSONShapeAndProfileRootGuard(t *testing.T) {
	repo := t.TempDir()
	writeCLITestSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha")

	var stdout, stderr bytes.Buffer
	code := Main([]string{"list", "--repo", repo, "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "list" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload["provenance_contract_version"] != discovery.ProvenanceContractVersion || payload["source_inventory_summary"] == nil {
		t.Fatalf("missing list provenance payload: %+v", payload)
	}
	packs := payload["packs"].([]any)
	if packs[0].(map[string]any)["pack_id"] != "alpha" {
		t.Fatalf("unexpected packs: %+v", packs)
	}
	if packs[0].(map[string]any)["source_class"] != "unknown_or_unclassified" || packs[0].(map[string]any)["provenance_state"] != "not_applicable" {
		t.Fatalf("source-only list row missing provenance defaults: %+v", packs[0])
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"list", "--repo", repo, "--profile", "demo", "--profile-root", filepath.Join(t.TempDir(), "profile"), "--json"}, &stdout, &stderr, map[string]string{})
	if code != 2 {
		t.Fatalf("expected guard failure, got %d", code)
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	diagnostics := payload["diagnostics"].([]any)
	if diagnostics[0].(map[string]any)["code"] != "profile_root_override_rejected" {
		t.Fatalf("unexpected guard payload: %+v", payload)
	}
}

func TestInstallDryRunJSONAndFailClosedGuards(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha")

	var stdout, stderr bytes.Buffer
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}
	code := Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("dry-run created profile root: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "install" || payload["mode"] != "dry_run" {
		t.Fatalf("unexpected dry-run payload: %+v", payload)
	}
	if payload["dry_run_plan_hash"] == "" || !strings.Contains(payload["next_action"].(string), "KAB is not required") {
		t.Fatalf("unexpected dry-run payload: %+v", payload)
	}
	if payload["provenance_contract_version"] != discovery.ProvenanceContractVersion || payload["source_inventory_snapshot"] == nil || payload["target_profile_inventory"] == nil {
		t.Fatalf("missing dry-run provenance payload: %+v", payload)
	}
	if payload["approval_request"].(map[string]any)["hash_includes_provenance"] != true {
		t.Fatalf("approval request does not hash-bind provenance: %+v", payload["approval_request"])
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing dry-run failure, got %d", code)
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["diagnostics"].([]any)[0].(map[string]any)["code"] != "install_requires_dry_run_or_approve" {
		t.Fatalf("unexpected missing dry-run payload: %+v", payload)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--approve", "dry-run:sha256:test", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected approve hash failure, got %d", code)
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["diagnostics"].([]any)[0].(map[string]any)["code"] != "approval_plan_hash_mismatch" {
		t.Fatalf("unexpected approve payload: %+v", payload)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("wrong approval wrote profile root: %v", err)
	}
}

func TestInstallApprovedCopyJSON(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("dry-run code=%d stderr=%s", code, stderr.String())
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}

	stdout.Reset()
	stderr.Reset()
	evidence := dryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	code = Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--approve", evidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("approve code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	approval := payload["approval"].(map[string]any)
	counts := payload["summary"].(map[string]any)["counts_by_action"].(map[string]any)
	if payload["mode"] != "approved_copy" || approval["dry_run_plan_hash"] != dryRun["dry_run_plan_hash"] || approval["approved_plan_hash"] != dryRun["dry_run_plan_hash"] {
		t.Fatalf("unexpected approved payload: %+v", payload)
	}
	if counts["manifest_update"].(float64) != 1 || payload["manifest_path"] == "" || payload["recovery"] == nil {
		t.Fatalf("missing manifest/recovery payload: %+v", payload)
	}
}

func TestInstallKABAdoptionStageFlagsDefaultsAndInteractive(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("default dry-run code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	stage := payload["kab_adoption_stage"].(map[string]any)
	if stage["numeric"].(float64) != 1 || stage["canonical"] != "stage1_direct_codex_app_server_baseline" || stage["source"] != "default_stage1" {
		t.Fatalf("unexpected default stage: %+v", stage)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot, "--kab-stage", "2", "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("stage2 dry-run code=%d stderr=%s", code, stderr.String())
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	stage = payload["kab_adoption_stage"].(map[string]any)
	if stage["numeric"].(float64) != 2 || stage["canonical"] != "stage2_kab_codex_first" || stage["source"] != "explicit_numeric" {
		t.Fatalf("unexpected numeric stage: %+v", stage)
	}

	for _, args := range [][]string{
		{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot, "--kab-stage", "1", "--kab-adoption-stage", "stage2_kab_codex_first", "--json"},
		{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot, "--kab-stage", "3", "--json"},
		{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot, "--kab-adoption-stage", "unknown", "--json"},
	} {
		stdout.Reset()
		stderr.Reset()
		code = Main(args, &stdout, &stderr, env)
		if code != 2 {
			t.Fatalf("expected selector failure for %v, got %d stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
		}
		payload = map[string]any{}
		if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if payload["diagnostics"].([]any)[0].(map[string]any)["code"] != "kab_adoption_stage_invalid" {
			t.Fatalf("unexpected selector failure payload: %+v", payload)
		}
	}

	oldInput := installPromptInput
	oldInteractive := installPromptInteractive
	defer func() {
		installPromptInput = oldInput
		installPromptInteractive = oldInteractive
	}()
	installPromptInput = strings.NewReader("\n")
	installPromptInteractive = func() bool { return true }
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--repo", repo, "--profile", "demo", "alpha", "--dry-run", "--profile-root", profileRoot}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("interactive dry-run code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Choice [1]:") || !strings.Contains(stdout.String(), "source interactive") {
		t.Fatalf("interactive output did not prompt/default: %q", stdout.String())
	}
}
