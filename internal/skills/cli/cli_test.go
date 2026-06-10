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
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/projectinstall"
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

func assertNoHangul(t *testing.T, out string) {
	t.Helper()
	for _, r := range out {
		if r >= 0xAC00 && r <= 0xD7AF {
			t.Fatalf("expected no Korean prose in human output, got %q", out)
		}
	}
}

func writeCLITestSkillPackYAML(t *testing.T, repo string, skills ...string) {
	t.Helper()
	content := "name: fixture\nversion: 0.1.0\nskills:\n"
	for _, skill := range skills {
		content += "  - " + skill + "\n"
	}
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertCLIErrorCode(t *testing.T, data []byte, code string) {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != false {
		t.Fatalf("expected ok:false payload: %+v", payload)
	}
	diagnostics := payload["diagnostics"].([]any)
	if diagnostics[0].(map[string]any)["code"] != code {
		t.Fatalf("expected diagnostic %s, got %+v", code, payload)
	}
}

func assertEnglishCLIError(t *testing.T, data []byte) {
	t.Helper()
	out := string(data)
	if !strings.HasPrefix(out, "Error: ") {
		t.Fatalf("expected English error prefix, got %q", out)
	}
	for _, r := range out {
		if r >= 0xAC00 && r <= 0xD7AF {
			t.Fatalf("expected no Korean prose in CLI error output, got %q", out)
		}
	}
}

func assertEnglishHumanOutput(t *testing.T, out string, labels ...string) {
	t.Helper()
	for _, r := range out {
		if r >= 0xAC00 && r <= 0xD7AF {
			t.Fatalf("expected no Korean prose in human output, got %q", out)
		}
	}
	for _, label := range labels {
		if !strings.Contains(out, label) {
			t.Fatalf("expected human output to contain %q, got %q", label, out)
		}
	}
}

func assertNoWriteEvidence(t *testing.T, payload map[string]any) {
	t.Helper()
	noWrite, ok := payload["no_write"].(map[string]any)
	if !ok {
		t.Fatalf("payload missing no_write evidence: %+v", payload)
	}
	if noWrite["guaranteed"] != true {
		t.Fatalf("expected guaranteed no-write evidence: %+v", noWrite)
	}
	for _, key := range []string{
		"profile_write_count",
		"skill_write_count",
		"manifest_write_count",
		"kas_directory_write_count",
		"kah_state_write_count",
		"kab_runtime_mutation_count",
		"hermes_runtime_mutation_count",
		"auth_provider_config_write_count",
		"profile_activation_count",
	} {
		if value, ok := noWrite[key]; ok && value != float64(0) {
			t.Fatalf("expected %s=0 in no-write evidence: %+v", key, noWrite)
		}
	}
}

func TestPublicLifecycleFailClosedHumanErrorsAreEnglish(t *testing.T) {
	for name, args := range map[string][]string{
		"update-missing-dry-run": {
			"update",
			"--profile", "kwanwoo",
			"--project", "doksuri-server",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Main(args, &stdout, &stderr, nil)
			if code != 2 {
				t.Fatalf("expected fail-closed exit 2, got %d", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("expected human error on stderr only, got stdout=%q", stdout.String())
			}
			assertEnglishCLIError(t, stderr.Bytes())
		})
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
			for _, want := range []string{"list", "install", "update", "doctor", "repair", "uninstall"} {
				if !strings.Contains(out, want) {
					t.Fatalf("root help did not list public command %s: %q", want, out)
				}
			}
			if !strings.Contains(out, "Compatibility commands:") || !strings.Contains(out, "sync-project-kas") || !strings.Contains(out, "migrate-project-kas") {
				t.Fatalf("root help did not list available commands: %q", out)
			}
			if !strings.Contains(out, "update   Classify project KAS updates without writing") || !strings.Contains(out, "Compatibility commands:") {
				t.Fatalf("root help did not prioritize public dry-run lifecycle UX over legacy sync/migrate verbs: %q", out)
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected root help on stdout only, got stderr=%q", stderr.String())
			}
		})
	}
}

func TestSubcommandHelpExitsZero(t *testing.T) {
	for _, command := range []string{"list", "install", "update", "doctor", "repair", "uninstall", "sync-project-kas", "install-project-kas"} {
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

func TestPublicLifecycleWrappersJSONDryRunOnly(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profile")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("public install dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("public install dry-run created profile root: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["command"] != "install" || payload["mode"] != "project_dry_run" || payload["dry_run"] != true {
		t.Fatalf("unexpected public install payload: %+v", payload)
	}
	assertNoWriteEvidence(t, payload)

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--apply", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected public install ambiguous-mode rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "project_install_mode_ambiguous")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("public repair dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["command"] != "repair" || payload["mode"] != "project_repair_dry_run" || payload["no_write"].(map[string]any)["guaranteed"] != true {
		t.Fatalf("unexpected public repair payload: %+v", payload)
	}
	assertNoWriteEvidence(t, payload)

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--approve", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected public repair ambiguous-mode rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "project_repair_mode_ambiguous")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--from-generic", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("install --from-generic dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["command"] != "install" || payload["mode"] != "project_migration_dry_run" || payload["dry_run"] != true {
		t.Fatalf("unexpected public migration payload: %+v", payload)
	}
	assertNoWriteEvidence(t, payload)
}

func TestPublicLifecycleProjectHumanOutputIsEnglish(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profile")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	for name, tc := range map[string]struct {
		args       []string
		wantCode   int
		wantStdout bool
		labels     []string
	}{
		"install-project-dry-run": {
			args:       []string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot},
			wantCode:   0,
			wantStdout: true,
			labels:     []string{"Status:", "Source pack:", "Plan:", "Writes:", "Approval evidence:", "Next:"},
		},
		"install-from-generic-dry-run": {
			args:       []string{"install", "--from-generic", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot},
			wantCode:   0,
			wantStdout: true,
			labels:     []string{"Status:", "Source pack:", "Plan:", "Writes:", "Approval required:", "Approval evidence:", "Manual semantic-port:", "Next:"},
		},
		"repair-dry-run": {
			args:       []string{"repair", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot},
			wantCode:   0,
			wantStdout: true,
			labels:     []string{"Status:", "Source pack:", "Plan:", "Writes:", "Approval required:", "Approval evidence:", "project-suite diagnostic:", "Action:", "Next:"},
		},
		"doctor-project-suite": {
			args:     []string{"doctor", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--project-suite", "--profile-root", profileRoot},
			wantCode: 2,
			labels:   []string{"Status:", "manifest:", "suite:", "source_pack:", "error:", "Next:"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Main(tc.args, &stdout, &stderr, env)
			if code != tc.wantCode {
				t.Fatalf("expected code %d, got %d stdout=%s stderr=%s", tc.wantCode, code, stdout.String(), stderr.String())
			}
			out := stderr.String()
			if tc.wantStdout {
				out = stdout.String()
				if stderr.Len() != 0 {
					t.Fatalf("expected human output on stdout only, got stderr=%q", stderr.String())
				}
			} else if stdout.Len() != 0 {
				t.Fatalf("expected failed doctor human output on stderr only, got stdout=%q", stdout.String())
			}
			assertEnglishHumanOutput(t, out, tc.labels...)
		})
	}
}

func TestInstallProjectKASJSONShapeAndNoWrite(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-final-verify"), "kkachi-final-verify")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan", "kkachi-final-verify")
	profileRoot := filepath.Join(dir, "profiles", "kwanwoo")

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("install-project-kas dry-run created profile root: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "install-project-kas" || payload["mode"] != "project_dry_run" || payload["dry_run"] != true {
		t.Fatalf("unexpected project dry-run payload: %+v", payload)
	}
	if payload["cli_version"] == "" || payload["planned_manifest"] == nil || payload["planned_skills"] == nil || payload["changed_paths"] == nil || payload["checksums"] == nil || payload["plan_hash"] == "" {
		t.Fatalf("missing required JSON evidence: %+v", payload)
	}
	noWrite := payload["no_write"].(map[string]any)
	if noWrite["guaranteed"] != true || noWrite["profile_write_count"] != float64(0) || noWrite["manifest_write_count"] != float64(0) || noWrite["kah_state_write_count"] != float64(0) || noWrite["kab_runtime_mutation_count"] != float64(0) {
		t.Fatalf("unexpected no-write evidence: %+v", noWrite)
	}
	project := payload["project"].(map[string]any)
	if project["id"] != "doksuri-server" || project["target_suite_path"] != "skills/doksuri-server" {
		t.Fatalf("unexpected project evidence: %+v", project)
	}
	sourcePack := payload["source_pack"].(map[string]any)
	if sourcePack["id"] != projectinstall.VirtualSourcePackID || sourcePack["formal_registry"] != "skill-pack.yaml" {
		t.Fatalf("unexpected source pack evidence: %+v", sourcePack)
	}
}

func TestInstallProjectKASApprovedInstallAndFailClosedGuards(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profile")
	base := []string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--profile-root", profileRoot, "--json"}

	var stdout, stderr bytes.Buffer
	code := Main(base, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected missing dry-run exit 2, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "project_install_requires_dry_run_or_approve")

	for name, args := range map[string][]string{
		"profile":     {"install-project-kas", "--repo", repo, "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--dry-run", "--json"},
		"project":     {"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--source-pack", projectinstall.VirtualSourcePackID, "--dry-run", "--json"},
		"source-pack": {"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--json"},
	} {
		t.Run("missing-"+name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			code := Main(args, &stdout, &stderr, nil)
			if code != 2 {
				t.Fatalf("expected exit 2, got %d", code)
			}
			var payload map[string]any
			if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload["ok"] != false {
				t.Fatalf("expected ok:false error payload: %+v", payload)
			}
		})
	}

	stdout.Reset()
	stderr.Reset()
	code = Main(append(base, "--dry-run"), &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected profile-root guard failure, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "profile_root_override_rejected")

	stdout.Reset()
	stderr.Reset()
	code = Main(append(base, "--dry-run", "--approve", "dry-run:sha256:abc"), &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected ambiguous mode rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "project_install_mode_ambiguous")

	stdout.Reset()
	stderr.Reset()
	code = Main(append(base, "--approve", "not-a-dry-run-hash"), &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected malformed approve rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "approval_evidence_malformed")

	stdout.Reset()
	stderr.Reset()
	code = Main(append(base, "--approve", "--json"), &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected valueless approve rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "approval_evidence_malformed")

	stdout.Reset()
	stderr.Reset()
	code = Main(append(base, "--approve", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000"), &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected wrong hash rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "approval_plan_hash_mismatch")

	stdout.Reset()
	stderr.Reset()
	code = Main(append(base, "--dry-run", "--write"), &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected write form rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "project_install_write_form_unsupported")

	stdout.Reset()
	stderr.Reset()
	code = Main(append(base, "--dry-run"), &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 0 {
		t.Fatalf("dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}
	evidence := dryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	stdout.Reset()
	stderr.Reset()
	code = Main(append(base, "--approve", evidence), &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 0 {
		t.Fatalf("approved install failed: code=%d stderr=%s", code, stderr.String())
	}
	var approved map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &approved); err != nil {
		t.Fatal(err)
	}
	if approved["ok"] != true || approved["mode"] != "project_approved_copy" || approved["install_id"] == "" {
		t.Fatalf("unexpected approved payload: %+v", approved)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); err != nil {
		t.Fatalf("approved install did not write project skill: %v", err)
	}
}

func TestInstallProjectKASConflictJSONExitsTwo(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-kas"), "kkachi-kas")
	writeCLITestSkillPackYAML(t, repo, "kkachi-kas")
	profileRoot := filepath.Join(dir, "profile")
	var stdout, stderr bytes.Buffer
	code := Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected umbrella-only exit 2, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != false || payload["conflicts"] == nil || payload["plan_hash"] == "" {
		t.Fatalf("conflict payload did not include ok:false/hash-bound evidence: %+v", payload)
	}
	conflicts := payload["conflicts"].([]any)
	if conflicts[0].(map[string]any)["condition"] != "umbrella_only" {
		t.Fatalf("unexpected conflicts: %+v", conflicts)
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

func TestUpdateDryRunLifecycleJSONAndNoWrite(t *testing.T) {
	repo, projectRoot, statePath := setupCLISyncFixture(t)
	before, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"update", "--profile", "hwangchung", "--project", "kan-plugin", "--state", statePath, "--dry-run", "--json", "--repo", repo, "--project-root", projectRoot}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("update dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["command"] != "update" || payload["mode"] != "project_update_dry_run" || payload["dry_run"] != true {
		t.Fatalf("unexpected update payload: %+v", payload)
	}
	assertNoWriteEvidence(t, payload)
	if len(payload["target_profiles"].([]any)) != 1 || len(payload["planned_states"].([]any)) == 0 || len(payload["doctor_commands"].([]any)) != 1 {
		t.Fatalf("missing lifecycle planner fields: %+v", payload)
	}
	syncPayload := payload["sync_classification"].(map[string]any)
	if syncPayload["command"] != "sync-project-kas" || syncPayload["summary"].(map[string]any)["no_action_count"] != float64(1) {
		t.Fatalf("missing compatibility classification payload: %+v", syncPayload)
	}
	after, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("update dry-run rewrote state")
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"update", "--profile", "hwangchung", "--project", "kan-plugin", "--state", statePath, "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing dry-run failure, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "update_requires_dry_run_or_apply")
}

func TestUninstallDryRunPlansManifestTrackedOnly(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profile")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("install dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}
	evidence := dryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--approve", evidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("approved install failed: code=%d stderr=%s", code, stderr.String())
	}
	localOnly := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-local", "SKILL.md")
	writeCLITestSkill(t, filepath.Dir(localOnly), "doksuri-server-local")
	target := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")
	before, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"uninstall", "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("uninstall dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["command"] != "uninstall" || payload["mode"] != "project_uninstall_dry_run" || payload["dry_run"] != true {
		t.Fatalf("unexpected uninstall payload: %+v", payload)
	}
	futureApplyCommand, ok := payload["future_apply_command"].(string)
	if !ok || !strings.Contains(futureApplyCommand, "--apply dry-run:sha256:") ||
		!strings.Contains(futureApplyCommand, "--backup-vault-root <abs-path>") {
		t.Fatalf("future apply command must include approval evidence and backup vault placeholder, got %q", futureApplyCommand)
	}
	assertNoHangul(t, futureApplyCommand)
	removals := payload["planned_removals"].([]any)
	if len(removals) != 1 || removals[0].(map[string]any)["action"] != "remove" {
		t.Fatalf("unexpected planned removals: %+v", removals)
	}
	skipped := payload["skipped_local_files"].([]any)
	if len(skipped) != 1 || skipped[0].(map[string]any)["path"] != "skills/doksuri-server/doksuri-server-local/SKILL.md" {
		t.Fatalf("unexpected skipped files: %+v", skipped)
	}
	assertNoWriteEvidence(t, payload)
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("uninstall dry-run rewrote manifest-tracked file")
	}
	if _, err := os.Stat(localOnly); err != nil {
		t.Fatalf("uninstall dry-run touched skipped local-only file: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"uninstall", "--profile", "kwanwoo", "--project", "doksuri-server", "--apply", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected uninstall apply rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "approval_plan_hash_mismatch")

	evidence = payload["approval_request"].(map[string]any)["evidence_ref"].(string)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"uninstall", "--profile", "kwanwoo", "--project", "doksuri-server", "--apply", evidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected uninstall missing backup-vault-root rejection, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "backup_vault_root_rejected")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("missing vault root removed target: %v", err)
	}

	vaultRoot := filepath.Join(dir, "vault")
	if err := os.MkdirAll(vaultRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"uninstall", "--profile", "kwanwoo", "--project", "doksuri-server", "--apply", evidence, "--backup-vault-root", vaultRoot, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("approved uninstall failed: code=%d stderr=%s", code, stderr.String())
	}
	var applied map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil {
		t.Fatal(err)
	}
	if applied["ok"] != true || applied["mode"] != "project_uninstall_approved" || applied["dry_run"] != false {
		t.Fatalf("unexpected approved uninstall payload: %+v", applied)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("approved uninstall did not remove manifest target or unexpected stat error: %v", err)
	}
	if _, err := os.Stat(localOnly); err != nil {
		t.Fatalf("approved uninstall touched local-only file: %v", err)
	}
	backup := applied["backup_recovery"].(map[string]any)
	if backup["backup_verified"] != true || backup["backup_evidence_path"] == "" || backup["backup_sha256"] == "" {
		t.Fatalf("missing backup evidence: %+v", backup)
	}
	if _, err := os.Stat(backup["backup_evidence_path"].(string)); err != nil {
		t.Fatalf("backup evidence path missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(applied["backup_path"].(string), "files", "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); err != nil {
		t.Fatalf("backup file missing: %v", err)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest["project_suites"].([]any)) != 0 || manifest["last_uninstall"] == nil {
		t.Fatalf("manifest did not remove only project suite and preserve uninstall evidence: %+v", manifest)
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
	assertEnglishHumanOutput(t, stdout.String(), "Status:", "healthy", "warnings", "errors", "KAB")

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

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"list", "--repo", repo, "--no-color"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("list human code=%d stderr=%s", code, stderr.String())
	}
	assertEnglishHumanOutput(t, stdout.String(), "Status:", "Source:", "Next:")
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
	assertEnglishHumanOutput(t, stdout.String(), "Status:", "Target:", "Plan:", "Diagnostic:", "plan hash:")
	if !strings.Contains(stdout.String(), "Choice [1]:") || !strings.Contains(stdout.String(), "source interactive") {
		t.Fatalf("interactive output did not prompt/default: %q", stdout.String())
	}
}

func TestProjectSuiteDoctorRepairAndMigrateCLIForms(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profile")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"doctor", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--project-suite", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected missing project suite exit 2, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["mode"] != "project_suite_doctor" || payload["project"].(map[string]any)["id"] != "doksuri-server" {
		t.Fatalf("doctor --project-suite did not interpret --project as suite id: %+v", payload)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("repair dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["command"] != "repair-project-kas" || payload["mode"] != "project_repair_dry_run" || payload["source_pack"].(map[string]any)["id"] != projectinstall.VirtualSourcePackID || payload["source_pack"].(map[string]any)["source"] != "default" {
		t.Fatalf("unexpected repair dry-run payload: %+v", payload)
	}
	noWrite := payload["no_write"].(map[string]any)
	if noWrite["guaranteed"] != true || noWrite["kah_state_write_count"] != float64(0) || noWrite["kab_runtime_mutation_count"] != float64(0) || noWrite["auth_provider_config_write_count"] != float64(0) {
		t.Fatalf("unexpected repair no-write evidence: %+v", noWrite)
	}
	evidence := payload["approval_request"].(map[string]any)["evidence_ref"].(string)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--approve", evidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("approved repair failed: code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); err != nil {
		t.Fatalf("approved repair did not write project skill under temp profile: %v", err)
	}

	writeCLITestSkill(t, filepath.Join(profileRoot, "skills", "doksuri-server", "kkachi-plan"), "kkachi-plan")
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected rogue unknown project skill repair exit 2, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != false || payload["approval_request"].(map[string]any)["required"] != false || payload["diagnostics"].([]any)[0].(map[string]any)["code"] != "unknown_profile_skill_dir" {
		t.Fatalf("rogue unknown project skill must be non-approvable ok:false: %+v", payload)
	}
	if err := os.RemoveAll(filepath.Join(profileRoot, "skills", "doksuri-server", "kkachi-plan")); err != nil {
		t.Fatal(err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"migrate-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected --from-generic requirement, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "from_generic_required")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", "missing-suite", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected unknown explicit source-pack exit 2, got %d", code)
	}
	payload = map[string]any{}
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["diagnostics"].([]any)[0].(map[string]any)["code"] != "unknown_source_pack" {
		t.Fatalf("unexpected unknown source-pack payload: %+v", payload)
	}
}
