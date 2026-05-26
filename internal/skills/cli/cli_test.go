package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			if !strings.Contains(out, "list") || !strings.Contains(out, "install") || !strings.Contains(out, "doctor") {
				t.Fatalf("root help did not list available commands: %q", out)
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected root help on stdout only, got stderr=%q", stderr.String())
			}
		})
	}
}

func TestSubcommandHelpExitsZero(t *testing.T) {
	for _, command := range []string{"list", "install", "doctor"} {
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
	packs := payload["packs"].([]any)
	if packs[0].(map[string]any)["pack_id"] != "alpha" {
		t.Fatalf("unexpected packs: %+v", packs)
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
