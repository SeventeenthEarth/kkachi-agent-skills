package toolchain

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/install"
)

func fixedNow() time.Time {
	return time.Date(2026, 6, 20, 12, 34, 56, 0, time.UTC)
}

func fakeProbeRunner(t *testing.T, calls *[][]string, version string) Runner {
	t.Helper()
	return func(workDir string, args ...string) CommandResult {
		*calls = append(*calls, append([]string{workDir}, args...))
		if strings.Join(args, " ") != "project probe-toolchain --json --project-root "+workDir {
			return CommandResult{Stderr: []byte("unexpected command"), Err: errors.New("unexpected command")}
		}
		payload := fakeProbePayload(workDir, version)
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		return CommandResult{Stdout: data}
	}
}

func fakeProbeRunnerWith(t *testing.T, calls *[][]string, mutate func(map[string]any)) Runner {
	t.Helper()
	return func(workDir string, args ...string) CommandResult {
		*calls = append(*calls, append([]string{workDir}, args...))
		payload := fakeProbePayload(workDir, "0.2.0")
		if mutate != nil {
			mutate(payload)
		}
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		return CommandResult{Stdout: data}
	}
}

func fakeProbePayload(workDir string, version string) map[string]any {
	return map[string]any{
		"ok":             true,
		"schema_version": "kah.toolchain_probe.v1",
		"no_write": map[string]any{
			"guaranteed":  true,
			"write_count": float64(0),
		},
		"kah": map[string]any{
			"version":     version,
			"binary_path": filepath.Join(workDir, ".kkachi", "bin", "kkachi-agent-helper"),
		},
		"project": map[string]any{
			"root":                   workDir,
			"kkachi_dir":             filepath.Join(workDir, ".kkachi"),
			"kkachi_dir_present":     true,
			"project_initialized":    true,
			"workflow_graph_present": false,
		},
		"doctor": map[string]any{
			"status":       "PASS",
			"reason_codes": []any{},
		},
		"diagnostics": []any{},
	}
}

func TestDoctorFailsClosedWhenToolchainMissingInvalidOrUnreadableAndDoesNotWrite(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T, root string)
		code  string
	}{
		{
			name: "missing",
			code: "toolchain_missing",
		},
		{
			name: "invalid_yaml",
			setup: func(t *testing.T, root string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Join(root, ".kkachi"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, ".kkachi", "toolchain.yaml"), []byte("schema_version: [unterminated\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			code: "toolchain_invalid",
		},
		{
			name: "secret",
			setup: func(t *testing.T, root string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Join(root, ".kkachi"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, ".kkachi", "toolchain.yaml"), []byte("schema_version: \"kkachi.toolchain.v1\"\nprovider_api_key: \"sk-live-secret\"\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			code: "toolchain_secret_detected",
		},
		{
			name: "unreadable",
			setup: func(t *testing.T, root string) {
				t.Helper()
				if runtime.GOOS == "windows" {
					t.Skip("chmod unreadable metadata fixture is POSIX-only")
				}
				path := filepath.Join(root, ".kkachi", "toolchain.yaml")
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(`schema_version: "kkachi.toolchain.v1"`+"\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(path, 0o644)
				})
				if err := os.Chmod(path, 0o000); err != nil {
					t.Fatal(err)
				}
			},
			code: "toolchain_unreadable",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.setup != nil {
				tc.setup(t, root)
			}
			before := snapshotFiles(t, root)
			result := Doctor(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &[][]string{}, "0.2.0"), Now: fixedNow})
			after := snapshotFiles(t, root)
			if !reflect.DeepEqual(before, after) {
				t.Fatalf("doctor wrote files: before=%v after=%v", before, after)
			}
			if result.OK {
				t.Fatalf("expected fail-closed doctor result: %+v", result)
			}
			if firstCode(result.Diagnostics) != tc.code {
				t.Fatalf("expected %s, got %+v", tc.code, result.Diagnostics)
			}
		})
	}
}

func TestInitRejectsUnsafeProbeFactsBeforeWriting(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
		code   string
	}{
		{
			name: "project_root_mismatch",
			mutate: func(payload map[string]any) {
				project := payload["project"].(map[string]any)
				project["root"] = filepath.Join(os.TempDir(), "other-project")
			},
			code: "kah_probe_project_root_mismatch",
		},
		{
			name: "secret_like_probe_value",
			mutate: func(payload map[string]any) {
				kah := payload["kah"].(map[string]any)
				kah["version"] = "sk-live-from-probe"
			},
			code: "kah_probe_secret_detected",
		},
		{
			name: "relative_binary_path",
			mutate: func(payload map[string]any) {
				kah := payload["kah"].(map[string]any)
				kah["binary_path"] = "bin/kkachi-agent-helper"
			},
			code: "kah_probe_facts_missing",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var calls [][]string
			result := Init(Options{ProjectRoot: root, Runner: fakeProbeRunnerWith(t, &calls, tc.mutate), Now: fixedNow})
			if result.OK || firstCode(result.Diagnostics) != tc.code {
				t.Fatalf("expected %s, got %+v", tc.code, result)
			}
			if _, err := os.Stat(filepath.Join(root, ".kkachi", "toolchain.yaml")); !os.IsNotExist(err) {
				t.Fatalf("unsafe init must not write toolchain.yaml, stat err=%v", err)
			}
		})
	}
}

func TestInitRejectsUnhealthyKAHProbeBeforeWriting(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
		code   string
	}{
		{
			name: "uninitialized_project",
			mutate: func(payload map[string]any) {
				project := payload["project"].(map[string]any)
				project["project_initialized"] = false
			},
			code: "kah_probe_project_uninitialized",
		},
		{
			name: "doctor_fail",
			mutate: func(payload map[string]any) {
				doctor := payload["doctor"].(map[string]any)
				doctor["status"] = "FAIL"
			},
			code: "kah_probe_doctor_not_pass",
		},
		{
			name: "doctor_unknown",
			mutate: func(payload map[string]any) {
				doctor := payload["doctor"].(map[string]any)
				doctor["status"] = "UNKNOWN"
			},
			code: "kah_probe_doctor_not_pass",
		},
		{
			name: "doctor_warn",
			mutate: func(payload map[string]any) {
				doctor := payload["doctor"].(map[string]any)
				doctor["status"] = "WARN"
			},
			code: "kah_probe_doctor_not_pass",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var calls [][]string
			result := Init(Options{ProjectRoot: root, Runner: fakeProbeRunnerWith(t, &calls, tc.mutate), Now: fixedNow})
			if result.OK || firstCode(result.Diagnostics) != tc.code {
				t.Fatalf("expected %s, got %+v", tc.code, result)
			}
			if _, err := os.Stat(filepath.Join(root, ".kkachi", "toolchain.yaml")); !os.IsNotExist(err) {
				t.Fatalf("unhealthy init must not write toolchain.yaml, stat err=%v", err)
			}
		})
	}
}

func TestLegacyKAHCLIKeysAreCompatibleAndInvalidValuesFailClosed(t *testing.T) {
	root := t.TempDir()
	helper := filepath.Join(root, ".kkachi", "bin", "kkachi-agent-helper")
	if err := os.MkdirAll(filepath.Dir(helper), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(helper, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".kkachi", "toolchain.yaml"), []byte("kah_cli: \"v0.2.0\"\nkah_cli_path: \""+helper+"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var calls [][]string
	ok := Doctor(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !ok.OK {
		t.Fatalf("expected compatibility keys to validate with matching probe: %+v", ok.Diagnostics)
	}

	if err := os.WriteFile(filepath.Join(root, ".kkachi", "toolchain.yaml"), []byte("kah_cli: \"not-a-version\"\nkah_cli_path: \""+helper+"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bad := Doctor(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if bad.OK || firstCode(bad.Diagnostics) != "kah_cli_invalid" {
		t.Fatalf("expected invalid legacy version to fail closed, got %+v", bad)
	}
}

func TestDoctorFailsClosedForStage2ActivationInStage1Metadata(t *testing.T) {
	root := t.TempDir()
	var calls [][]string
	init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !init.OK {
		t.Fatalf("init failed: %+v", init.Diagnostics)
	}
	path := filepath.Join(root, ".kkachi", "toolchain.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(data), `stage2_activation: false`, `stage2_activation: true`)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	result := Doctor(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if result.OK || firstCode(result.Diagnostics) != "toolchain_stage2_activation_unsupported" {
		t.Fatalf("expected Stage 2 fail-closed diagnostic, got %+v", result)
	}
}

func TestDoctorFailsClosedForUnhealthyStoredV1KAHMetadataAndDoesNotWrite(t *testing.T) {
	for _, tc := range []struct {
		name    string
		replace func(string) string
		code    string
	}{
		{
			name: "project_uninitialized",
			replace: func(text string) string {
				return strings.ReplaceAll(text, `project_initialized: true`, `project_initialized: false`)
			},
			code: "toolchain_kah_project_uninitialized",
		},
		{
			name: "doctor_not_pass",
			replace: func(text string) string {
				return strings.ReplaceAll(text, `doctor_status: "PASS"`, `doctor_status: "WARN"`)
			},
			code: "toolchain_kah_doctor_not_pass",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var calls [][]string
			init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
			if !init.OK {
				t.Fatalf("init failed: %+v", init.Diagnostics)
			}
			path := filepath.Join(root, ".kkachi", "toolchain.yaml")
			original := readFile(t, path)
			modified := tc.replace(original)
			if modified == original {
				t.Fatalf("test fixture did not modify toolchain metadata")
			}
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				t.Fatal(err)
			}

			result := Doctor(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
			if result.OK || firstCode(result.Diagnostics) != tc.code {
				t.Fatalf("expected %s, got %+v", tc.code, result)
			}
			if after := readFile(t, path); after != modified {
				t.Fatalf("doctor wrote toolchain metadata:\nbefore:\n%s\nafter:\n%s", modified, after)
			}
		})
	}
}

func TestInitWritesToolchainAtomicallyAfterDeterministicKAHProbe(t *testing.T) {
	root := t.TempDir()
	var calls [][]string
	result := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !result.OK {
		t.Fatalf("init failed: %+v", result.Diagnostics)
	}
	if len(calls) != 1 || strings.Join(calls[0][1:], " ") != "project probe-toolchain --json --project-root "+result.ProjectRoot {
		t.Fatalf("unexpected KAH probe calls: %+v", calls)
	}
	toolchainPath := filepath.Join(root, ".kkachi", "toolchain.yaml")
	data, err := os.ReadFile(toolchainPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		`schema_version: "kkachi.toolchain.v1"`,
		`generator: "kkachi-agent-skills toolchain init"`,
		`canonical: "stage1_direct_codex_app_server_baseline"`,
		`stage2_activation: false`,
		`providers: {}`,
		`no_secrets: true`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated toolchain missing %q:\n%s", want, text)
		}
	}
	for _, legacy := range []string{"kah_cli:", "kah_cli_path:"} {
		if strings.Contains(text, legacy) {
			t.Fatalf("generated v1 toolchain must not be legacy-only; found %q in:\n%s", legacy, text)
		}
	}
	if leftovers, _ := filepath.Glob(filepath.Join(root, ".kkachi", ".toolchain.yaml.tmp-*")); len(leftovers) != 0 {
		t.Fatalf("atomic temp files left behind: %v", leftovers)
	}
}

func TestRefreshRejectsUnhealthyCurrentProbeAndPreservesToolchain(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
		code   string
	}{
		{
			name: "uninitialized_project",
			mutate: func(payload map[string]any) {
				project := payload["project"].(map[string]any)
				project["project_initialized"] = false
			},
			code: "kah_probe_project_uninitialized",
		},
		{
			name: "doctor_fail",
			mutate: func(payload map[string]any) {
				doctor := payload["doctor"].(map[string]any)
				doctor["status"] = "FAIL"
			},
			code: "kah_probe_doctor_not_pass",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var calls [][]string
			init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
			if !init.OK {
				t.Fatalf("init failed: %+v", init.Diagnostics)
			}
			path := filepath.Join(root, ".kkachi", "toolchain.yaml")
			before := readFile(t, path)
			result := Refresh(Options{ProjectRoot: root, Runner: fakeProbeRunnerWith(t, &calls, tc.mutate), Now: fixedNow})
			if result.OK || firstCode(result.Diagnostics) != tc.code {
				t.Fatalf("expected %s, got %+v", tc.code, result)
			}
			if after := readFile(t, path); after != before {
				t.Fatalf("refresh changed toolchain metadata on unhealthy probe:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}

func TestRefreshUpdatesObservedFactsAndPreservesPolicyFields(t *testing.T) {
	root := t.TempDir()
	var calls [][]string
	init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !init.OK {
		t.Fatalf("init failed: %+v", init.Diagnostics)
	}
	path := filepath.Join(root, ".kkachi", "toolchain.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(data), `selected_at: "2026-06-20T12:34:56Z"`, `selected_at: "approved-time"`)
	text = strings.ReplaceAll(text, `approval_evidence: "not_applicable"`, `approval_evidence: "ticket-123"`)
	text = strings.ReplaceAll(text, `required_roles: ["logic", "security", "arch", "cve", "test_adequacy"]`, `required_roles: ["logic", "test_adequacy"]`)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	refreshed := Refresh(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.3.0"), Now: func() time.Time {
		return fixedNow().Add(time.Hour)
	}})
	if !refreshed.OK {
		t.Fatalf("refresh failed: %+v", refreshed.Diagnostics)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updatedText := string(updated)
	for _, want := range []string{
		`generator: "kkachi-agent-skills toolchain refresh"`,
		`kah_cli_version: "0.3.0"`,
		`selected_at: "approved-time"`,
		`approval_evidence: "ticket-123"`,
		`schema_version: "mar.role_lanes.v1"`,
		`required_roles: ["logic", "test_adequacy"]`,
		`schema_version: "mar.provider_tools.v1"`,
	} {
		if !strings.Contains(updatedText, want) {
			t.Fatalf("refreshed toolchain missing %q:\n%s", want, updatedText)
		}
	}
}

func TestImportLegacyWritesStage1ToolchainFromExplicitProfileProject(t *testing.T) {
	root := t.TempDir()
	profileRoot := t.TempDir()
	writeLegacyState(t, profileRoot, stage1LegacyFixture("kan-plugin"))
	var calls [][]string
	result := ImportLegacy(Options{ProjectRoot: root, Profile: "hwangchung", Project: "kan-plugin", ProfileRoot: profileRoot, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !result.OK || !result.Wrote {
		t.Fatalf("import-legacy failed: %+v", result)
	}
	text := readFile(t, filepath.Join(root, ".kkachi", "toolchain.yaml"))
	for _, want := range []string{
		`generator: "kkachi-agent-skills toolchain import-legacy"`,
		`id: "kan-plugin"`,
		`canonical: "stage1_direct_codex_app_server_baseline"`,
		`selected_at: "2026-06-06T00:00:00Z"`,
		`approval_evidence: "legacy-ticket-1"`,
		`stage2_activation: false`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("imported toolchain missing %q:\n%s", want, text)
		}
	}
}

func TestImportLegacyFailsClosedOnExistingToolchainConflict(t *testing.T) {
	root := t.TempDir()
	profileRoot := t.TempDir()
	writeLegacyState(t, profileRoot, stage1LegacyFixture("kan-plugin"))
	var calls [][]string
	init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !init.OK {
		t.Fatalf("init failed: %+v", init.Diagnostics)
	}
	before := readFile(t, filepath.Join(root, ".kkachi", "toolchain.yaml"))
	result := ImportLegacy(Options{ProjectRoot: root, Profile: "hwangchung", Project: "kan-plugin", ProfileRoot: profileRoot, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if result.OK || firstCode(result.Diagnostics) != "toolchain_import_existing_toolchain_conflict" {
		t.Fatalf("expected existing toolchain conflict, got %+v", result)
	}
	if after := readFile(t, filepath.Join(root, ".kkachi", "toolchain.yaml")); after != before {
		t.Fatalf("import-legacy overwrote existing toolchain:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestImportLegacyFailsClosedWhenLegacyStateMissing(t *testing.T) {
	root := t.TempDir()
	result := ImportLegacy(Options{ProjectRoot: root, Profile: "hwangchung", Project: "kan-plugin", ProfileRoot: t.TempDir(), Runner: fakeProbeRunner(t, &[][]string{}, "0.2.0"), Now: fixedNow})
	if result.OK || firstCode(result.Diagnostics) != "legacy_state_missing" {
		t.Fatalf("expected missing legacy state, got %+v", result)
	}
	if _, err := os.Stat(filepath.Join(root, ".kkachi", "toolchain.yaml")); !os.IsNotExist(err) {
		t.Fatalf("missing legacy state must not write toolchain.yaml, stat err=%v", err)
	}
}

func TestImportLegacyFailsClosedOnConflictingLegacyFacts(t *testing.T) {
	for _, tc := range []struct {
		name     string
		setup    func(t *testing.T, profileRoot string)
		profile  string
		project  string
		wantCode string
	}{
		{
			name: "explicit_project_mismatch",
			setup: func(t *testing.T, profileRoot string) {
				fixture := stage1LegacyFixture("other-project")
				fixture.StateProjectID = "kan-plugin"
				fixture.StateKASSuite = "kan-plugin"
				writeLegacyState(t, profileRoot, fixture)
			},
			profile:  "hwangchung",
			project:  "other-project",
			wantCode: "legacy_state_conflict",
		},
		{
			name: "state_profile_mismatch",
			setup: func(t *testing.T, profileRoot string) {
				writeLegacyState(t, profileRoot, stage1LegacyFixture("kan-plugin"))
			},
			profile:  "other-profile",
			project:  "kan-plugin",
			wantCode: "legacy_state_conflict",
		},
		{
			name: "marker_stage_conflict",
			setup: func(t *testing.T, profileRoot string) {
				writeLegacyState(t, profileRoot, stage1LegacyFixture("kan-plugin"))
				stage2, err := install.ResolveKABAdoptionStage(install.StageSelectionInput{Numeric: "2"})
				if err != nil {
					t.Fatal(err)
				}
				markerPath := filepath.Join(profileRoot, "skills", "kan-plugin", "kan-plugin-kas", "references", "kab-adoption-stage.md")
				if err := os.WriteFile(markerPath, []byte(install.KABAdoptionStageMarkerContent(stage2)), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			profile:  "hwangchung",
			project:  "kan-plugin",
			wantCode: "legacy_stage_marker_conflict",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			profileRoot := t.TempDir()
			tc.setup(t, profileRoot)
			result := ImportLegacy(Options{ProjectRoot: root, Profile: tc.profile, Project: tc.project, ProfileRoot: profileRoot, Runner: fakeProbeRunner(t, &[][]string{}, "0.2.0"), Now: fixedNow})
			if result.OK || firstCode(result.Diagnostics) != tc.wantCode {
				t.Fatalf("expected %s, got %+v", tc.wantCode, result)
			}
			if _, err := os.Stat(filepath.Join(root, ".kkachi", "toolchain.yaml")); !os.IsNotExist(err) {
				t.Fatalf("conflicting legacy facts must not write toolchain.yaml, stat err=%v", err)
			}
		})
	}
}

func TestSetStageStage1RecordsApprovalWithoutKABClaim(t *testing.T) {
	root := t.TempDir()
	var calls [][]string
	init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !init.OK {
		t.Fatalf("init failed: %+v", init.Diagnostics)
	}
	result := SetStage(Options{ProjectRoot: root, Stage: "1", ApprovalEvidence: "approval:t-123", Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: func() time.Time {
		return fixedNow().Add(time.Hour)
	}})
	if !result.OK || !result.Wrote {
		t.Fatalf("set-stage Stage 1 failed: %+v", result)
	}
	text := readFile(t, filepath.Join(root, ".kkachi", "toolchain.yaml"))
	for _, want := range []string{
		`generator: "kkachi-agent-skills toolchain set-stage"`,
		`numeric: 1`,
		`approval_evidence: "approval:t-123"`,
		`stage2_activation: false`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("Stage 1 set-stage output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "native_codex") || strings.Contains(text, `stage2_activation: true`) {
		t.Fatalf("Stage 1 metadata must not claim KAB native_codex execution:\n%s", text)
	}
}

func TestSetStageStage2AndStage3FailClosed(t *testing.T) {
	for _, tc := range []struct {
		name             string
		stage            string
		approvalEvidence string
		code             string
	}{
		{name: "stage2_missing_approval", stage: "2", code: "stage_approval_evidence_required"},
		{name: "stage2_missing_capability", stage: "2", approvalEvidence: "approval:t-2", code: "toolchain_stage2_capability_proof_missing"},
		{name: "stage3_reserved", stage: "3", approvalEvidence: "approval:t-3", code: "toolchain_stage3_reserved"},
		{name: "stage1_secret_approval", stage: "1", approvalEvidence: "sk-live-secret", code: "toolchain_secret_detected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var calls [][]string
			init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
			if !init.OK {
				t.Fatalf("init failed: %+v", init.Diagnostics)
			}
			path := filepath.Join(root, ".kkachi", "toolchain.yaml")
			before := readFile(t, path)
			result := SetStage(Options{ProjectRoot: root, Stage: tc.stage, ApprovalEvidence: tc.approvalEvidence, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
			if result.OK || firstCode(result.Diagnostics) != tc.code {
				t.Fatalf("expected %s, got %+v", tc.code, result)
			}
			if after := readFile(t, path); after != before {
				t.Fatalf("failed set-stage changed toolchain:\nbefore:\n%s\nafter:\n%s", before, after)
			}
		})
	}
}

func TestDoctorAcceptsBoundedNonSecretMARProviderProof(t *testing.T) {
	root := t.TempDir()
	var calls [][]string
	init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !init.OK {
		t.Fatalf("init failed: %+v", init.Diagnostics)
	}
	path := filepath.Join(root, ".kkachi", "toolchain.yaml")
	text := strings.ReplaceAll(readFile(t, path), "    providers: {}", `    providers:
      zcode_glm_5_2:
        resolved_argv: ["scripts/mar_adapters/mar-zcode.sh", "--prompt-file"]
        command_lane: "zcode"
        selected_model: "glm-5.2"
        validated: true
        adapter_proof_evidence: "local-proof:t-123"`)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	result := Doctor(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
	if !result.OK {
		t.Fatalf("expected bounded non-secret MAR provider proof to validate, got %+v", result.Diagnostics)
	}
}

func TestDoctorFailsClosedForMalformedOrSecretMARProviderProof(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(string) string
		code   string
	}{
		{
			name: "malformed_boolean",
			mutate: func(text string) string {
				return strings.ReplaceAll(text, "    providers: {}", "    providers:\n      zcode_glm_5_2:\n        validated: maybe")
			},
			code: "toolchain_mar_provider_tools_invalid",
		},
		{
			name: "secret_like_value",
			mutate: func(text string) string {
				return strings.ReplaceAll(text, "    providers: {}", "    providers:\n      zcode_glm_5_2:\n        adapter_proof_evidence: \"sk-live-secret\"")
			},
			code: "toolchain_secret_detected",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			var calls [][]string
			init := Init(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
			if !init.OK {
				t.Fatalf("init failed: %+v", init.Diagnostics)
			}
			path := filepath.Join(root, ".kkachi", "toolchain.yaml")
			if err := os.WriteFile(path, []byte(tc.mutate(readFile(t, path))), 0o644); err != nil {
				t.Fatal(err)
			}
			result := Doctor(Options{ProjectRoot: root, Runner: fakeProbeRunner(t, &calls, "0.2.0"), Now: fixedNow})
			if result.OK || firstCode(result.Diagnostics) != tc.code {
				t.Fatalf("expected %s, got %+v", tc.code, result)
			}
		})
	}
}

type legacyStateFixture struct {
	Profile        string
	PathProject    string
	StateProjectID string
	StateKASSuite  string
	StageNumeric   int
	StageCanonical string
}

func stage1LegacyFixture(project string) legacyStateFixture {
	return legacyStateFixture{
		Profile:        "hwangchung",
		PathProject:    project,
		StateProjectID: project,
		StateKASSuite:  project,
		StageNumeric:   1,
		StageCanonical: "stage1_direct_codex_app_server_baseline",
	}
}

func writeLegacyState(t *testing.T, profileRoot string, fixture legacyStateFixture) {
	t.Helper()
	dir := filepath.Join(profileRoot, "skills", fixture.PathProject, fixture.PathProject+"-kas", "references")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	state := `version: "0.1"
project:
  id: "` + fixture.StateProjectID + `"
  repo: "` + fixture.StateProjectID + `"
  kas_suite: "` + fixture.StateKASSuite + `"
  profile: "` + fixture.Profile + `"
kab_adoption_stage:
  numeric: ` + strconv.Itoa(fixture.StageNumeric) + `
  canonical: "` + fixture.StageCanonical + `"
  selection_source: "approved_project_policy"
  selected_at: "2026-06-06T00:00:00Z"
  approval_evidence: "legacy-ticket-1"
  stage2_activation: false
`
	if err := os.WriteFile(filepath.Join(dir, "kas-project-state.yaml"), []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}
	stage, err := install.ResolveKABAdoptionStage(install.StageSelectionInput{Numeric: strconv.Itoa(fixture.StageNumeric), Canonical: fixture.StageCanonical})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "kab-adoption-stage.md"), []byte(install.KABAdoptionStageMarkerContent(stage)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func snapshotFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return files
}

func firstCode(diagnostics []Diagnostic) string {
	if len(diagnostics) == 0 {
		return ""
	}
	return diagnostics[0].Code
}
