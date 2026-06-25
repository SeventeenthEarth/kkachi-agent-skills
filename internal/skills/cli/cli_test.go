package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/projectinstall"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowregistry"
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

func writeCLITestSkillWithBody(t *testing.T, dir string, name string, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\n---\n# "+name+"\n"+body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func cliRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
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
	writeCLITestRoleRegistry(t, repo)
}

func writeCLITestRoleRegistry(t *testing.T, repo string) {
	t.Helper()
	content := `version: role-aware-project-suite/v1
roles:
  blue_commander:
    display_label: "Blue commander / full project suite"
    selection_mode: full_source_suite
    required_source_skills: "*"
    optional_source_skills: []
    forbidden_source_skills: []
  red_reviewer:
    display_label: "Red safety/fail-closed reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
      - kkachi-verify
    optional_source_skills: []
    forbidden_source_skills:
      - kkachi-implement
  orange_pm_reviewer:
    display_label: "Orange operator-value reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
    optional_source_skills:
      - kkachi-final-verify
    optional_selection_default: unselected
    forbidden_source_skills:
      - kkachi-implement
  gray_scribe:
    display_label: "Gray evidence/scribe reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
      - kkachi-final-verify
    optional_source_skills: []
    forbidden_source_skills:
      - kkachi-implement
      - kkachi-docs-update
`
	if err := os.MkdirAll(filepath.Join(repo, "registries"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, filepath.FromSlash(projectinstall.RoleRegistryPath)), []byte(content), 0o644); err != nil {
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

func assertPluginNoWriteEvidence(t *testing.T, payload map[string]any) {
	t.Helper()
	noWrite, ok := payload["no_write_evidence"].(map[string]any)
	if !ok {
		t.Fatalf("payload missing plugin no_write_evidence: %+v", payload)
	}
	if noWrite["guaranteed"] != true {
		t.Fatalf("expected guaranteed plugin no-write evidence: %+v", noWrite)
	}
	for _, key := range []string{
		"profile_wrapper_write_count",
		"project_overlay_write_count",
		"copied_legacy_suite_write_count",
		"kah_state_write_count",
		"kab_runtime_mutation_count",
		"hermes_runtime_mutation_count",
		"auth_provider_config_write_count",
		"profile_activation_count",
	} {
		if noWrite[key] != float64(0) {
			t.Fatalf("expected %s=0 in plugin no-write evidence: %+v", key, noWrite)
		}
	}
}

func installCLITestFakeKAH(t *testing.T) {
	t.Helper()
	binDir := t.TempDir()
	helper := filepath.Join(binDir, "kkachi-agent-helper")
	script := `#!/bin/sh
case "$*" in
  "--version")
    echo "kkachi-agent-helper 0.1.9"
    ;;
  "capabilities --json")
    echo '{"commands":["graph validate","graph explain"],"flags":["workflow_graph_readonly","workflow_graph_diagnostics","workflow_graph_no_direct_yaml_fallback","workflow_graph_configurable_feedback_intake"]}'
    ;;
  "graph --help")
    echo "Usage: graph"
    echo "  validate"
    echo "  explain"
    ;;
  "graph validate --file .kkachi-workflow.yaml --json")
    echo '{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0"}'
    ;;
  "graph explain --file .kkachi-workflow.yaml --json")
    echo '{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0"}'
    ;;
  *)
    echo "unexpected kkachi-agent-helper call: $*" >&2
    exit 9
    ;;
esac
`
	if err := os.WriteFile(helper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func installCLIWorkflowGraphRepairFakeKAH(t *testing.T) {
	t.Helper()
	binDir := t.TempDir()
	helper := filepath.Join(binDir, "kkachi-agent-helper")
	script := `#!/bin/sh
case "$*" in
  "--version")
    echo "kkachi-agent-helper 0.1.9"
    ;;
  "capabilities --json")
    echo '{"commands":["graph validate","graph explain","graph diff","graph propose","graph apply"],"flags":["workflow_graph_readonly","workflow_graph_diagnostics","workflow_graph_no_direct_yaml_fallback","workflow_graph_configurable_feedback_intake","workflow_graph_apply"]}'
    ;;
  "graph --help")
    echo "Usage: graph"
    echo "  validate"
    echo "  explain"
    echo "  diff"
    echo "  propose"
    echo "  apply"
    ;;
  "graph validate --file .kkachi-workflow.yaml --json")
    if grep -q applied .kkachi-workflow.yaml 2>/dev/null; then
      echo '{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0","checksum":"sha256:new"}'
    else
      echo '{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.0.9","checksum":"sha256:old"}'
    fi
    ;;
  "graph explain --file .kkachi-workflow.yaml --json")
    if grep -q applied .kkachi-workflow.yaml 2>/dev/null; then
      echo '{"ok":true,"graph":{"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0","checksum":"sha256:new"}}'
    else
      echo '{"ok":true,"graph":{"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.0.9","checksum":"sha256:old"}}'
    fi
    ;;
  "graph diff --from .kkachi-workflow.yaml --to .kkachi/graph/candidates/kas-default-3dcade9fff1844f8.yaml --semantic --json")
    echo '{"ok":true,"summary":"semantic diff ready","risk_flags":["phase_path_change"],"reason_codes":["graph_stale"]}'
    ;;
  "graph propose --candidate-file .kkachi/graph/candidates/kas-default-3dcade9fff1844f8.yaml --reason repair --json")
    echo '{"ok":true,"proposal_id":"prop-1","proposal_path":".kkachi/graph/proposals/prop-1.yaml","approval_required":true,"risk_flags":["phase_path_change"],"reason_codes":["proposal_recorded"]}'
    ;;
  "graph apply --proposal prop-1 --approval approved:1 --json")
    printf 'version: workflow-graph/v1\napplied: true\n' > .kkachi-workflow.yaml
    echo '{"ok":true,"proposal_id":"prop-1","approval_ref":"approved:1","audit_event_ids":["evt-1"],"audit_path":".kkachi/events/evt-1.json","backup_path":".kkachi/graph/backups/old.yaml","recovery_path":".kkachi/graph/recovery/prop-1.md"}'
    ;;
  *)
    echo "unexpected kkachi-agent-helper call: $*" >&2
    exit 9
    ;;
esac
`
	if err := os.WriteFile(helper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func installCLIWorkflowTriggerFakeKAH(t *testing.T, workflowSupported bool) {
	t.Helper()
	binDir := t.TempDir()
	helper := filepath.Join(binDir, "kkachi-agent-helper")
	script := `#!/bin/sh
case "$*" in
  "--version")
    echo "kkachi-agent-helper 0.1.10"
    ;;
  "capabilities --json")
    if [ "` + boolShellValue(workflowSupported) + `" = "1" ]; then
      echo '{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true,"workflow_strict_transition_ledger":true,"workflow_transition_order_verification":true,"workflow_phase_projection_validation":true}}'
    else
      echo '{"command_groups":[{"name":"graph","status":"supported","subcommands":["validate"]}],"compatibility_flags":{"workflow_instance_state":false}}'
    fi
    ;;
  "workflow --help")
    if [ "` + boolShellValue(workflowSupported) + `" = "1" ]; then
      echo "Subcommands:"
      echo "  validate"
      echo "  explain"
      echo "  create"
      echo "  show"
      echo "  ready"
      echo "  node"
    else
      echo "unknown help topic" >&2
      exit 2
    fi
    ;;
  "workflow validate --file .kkachi/workflows/demo.yaml --json")
    echo '{"ok":true,"status":"valid","workflow_id":"demo","schema_version":"task-dag/v1"}'
    ;;
  "workflow explain --file .kkachi/workflows/demo.yaml --json")
    echo '{"ok":true,"status":"valid","workflow_id":"demo","schema_version":"task-dag/v1"}'
    ;;
  "workflow create --run run-20260615T010203Z-abcdef123456 --file .kkachi/workflows/demo.yaml --json")
    echo '{"ok":true,"status":"pass","reason":"workflow_instance_created","run_id":"run-20260615T010203Z-abcdef123456","instance":{"version":"workflow-instance/v1","run_id":"run-20260615T010203Z-abcdef123456","workflow_id":"demo","schema_version":"task-dag/v1","source_path":".kkachi/workflows/demo.yaml","revision":1,"nodes":[{"id":"setup","depends_on":[],"join":"all_of","required_outputs":["artifacts/setup.md"],"state":"pending"}]},"ready":[{"id":"setup","reasons":["dependencies_satisfied","state_pending"]}]}'
    ;;
  "workflow ready --run run-20260615T010203Z-abcdef123456 --json")
    echo '{"ok":true,"status":"pass","reason":"workflow_ready_nodes_computed","run_id":"run-20260615T010203Z-abcdef123456","instance":{"version":"workflow-instance/v1","run_id":"run-20260615T010203Z-abcdef123456","workflow_id":"demo","schema_version":"task-dag/v1","source_path":".kkachi/workflows/demo.yaml","revision":1,"nodes":[{"id":"setup","depends_on":[],"join":"all_of","required_outputs":["artifacts/setup.md"],"state":"pending"}]},"ready":[{"id":"setup","reasons":["dependencies_satisfied","state_pending"]}]}'
    ;;
  *)
    echo "unexpected kkachi-agent-helper call: $*" >&2
    exit 9
    ;;
esac
`
	if err := os.WriteFile(helper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func installCLIWorkflowTriggerRunLocalFakeKAH(t *testing.T, workflowSupported bool) {
	t.Helper()
	binDir := t.TempDir()
	helper := filepath.Join(binDir, "kkachi-agent-helper")
	workflowFile := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml"
	requiredOutput := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/artifacts/setup.md"
	script := `#!/bin/sh
case "$*" in
  "--version")
    if [ "` + boolShellValue(workflowSupported) + `" = "1" ]; then
      echo "kkachi-agent-helper 0.1.10-source"
    else
      echo "kkachi-agent-helper 0.1.9"
    fi
    ;;
  "capabilities --json")
    if [ "` + boolShellValue(workflowSupported) + `" = "1" ]; then
      echo '{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true,"workflow_strict_transition_ledger":true,"workflow_transition_order_verification":true,"workflow_phase_projection_validation":true}}'
    else
      echo '{"command_groups":[{"name":"graph","status":"supported","subcommands":["validate"]}],"compatibility_flags":{"workflow_instance_state":false}}'
    fi
    ;;
  "workflow --help")
    if [ "` + boolShellValue(workflowSupported) + `" = "1" ]; then
      echo "Subcommands:"
      echo "  validate"
      echo "  explain"
      echo "  create"
      echo "  show"
      echo "  ready"
      echo "  node"
    else
      echo "unknown help topic" >&2
      exit 2
    fi
    ;;
  "workflow validate --file ` + workflowFile + ` --json")
    echo '{"ok":true,"status":"valid","workflow_id":"demo","schema_version":"task-dag/v1"}'
    ;;
  "workflow explain --file ` + workflowFile + ` --json")
    echo '{"ok":true,"status":"valid","workflow_id":"demo","schema_version":"task-dag/v1","nodes":[{"id":"setup"}]}'
    ;;
  "workflow create --run run-20260616T105614Z-4b0ebe11b67d --file ` + workflowFile + ` --json")
    echo '{"ok":true,"status":"pass","reason":"workflow_instance_created","run_id":"run-20260616T105614Z-4b0ebe11b67d","instance":{"version":"workflow-instance/v1","run_id":"run-20260616T105614Z-4b0ebe11b67d","workflow_id":"demo","schema_version":"task-dag/v1","source_path":"` + workflowFile + `","revision":1,"nodes":[{"id":"setup","depends_on":[],"join":"all_of","required_outputs":["` + requiredOutput + `"],"state":"pending"}]},"ready":[{"id":"setup","reasons":["dependencies_satisfied","state_pending"]}]}'
    ;;
  "workflow ready --run run-20260616T105614Z-4b0ebe11b67d --json")
    echo '{"ok":true,"status":"pass","reason":"workflow_ready_nodes_computed","run_id":"run-20260616T105614Z-4b0ebe11b67d","instance":{"version":"workflow-instance/v1","run_id":"run-20260616T105614Z-4b0ebe11b67d","workflow_id":"demo","schema_version":"task-dag/v1","source_path":"` + workflowFile + `","revision":1,"nodes":[{"id":"setup","depends_on":[],"join":"all_of","required_outputs":["` + requiredOutput + `"],"state":"pending"}]},"ready":[{"id":"setup","reasons":["dependencies_satisfied","state_pending"]}]}'
    ;;
  *)
    echo "unexpected kkachi-agent-helper call: $*" >&2
    exit 9
    ;;
esac
`
	if err := os.WriteFile(helper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func installCLIWorkflowCreateFakeKAH(t *testing.T, workflowCatalogSupported bool) {
	t.Helper()
	binDir := t.TempDir()
	helper := filepath.Join(binDir, "kkachi-agent-helper")
	script := `#!/bin/sh
case "$*" in
  "--version")
    if [ "` + boolShellValue(workflowCatalogSupported) + `" = "1" ]; then
      echo "kkachi-agent-helper 0.1.10-source"
    else
      echo "kkachi-agent-helper 0.1.9"
    fi
    ;;
  "capabilities --json")
    if [ "` + boolShellValue(workflowCatalogSupported) + `" = "1" ]; then
      echo '{"command_groups":[{"name":"workflow","status":"supported","subcommands":["validate","explain","catalog","create","show","ready","node"]}],"compatibility_flags":{"task_dag_schema_validation":true,"workflow_instance_state":true,"workflow_catalog_diagnostics":true,"workflow_final_gate_integration":true,"workflow_node_contract_registry_evidence":true}}'
    else
      echo '{"command_groups":[{"name":"graph","status":"supported","subcommands":["validate"]}],"compatibility_flags":{"workflow_instance_state":false}}'
    fi
    ;;
  "workflow --help")
    if [ "` + boolShellValue(workflowCatalogSupported) + `" = "1" ]; then
      echo "workflow validate"
      echo "workflow explain"
      echo "workflow catalog validate"
      echo "workflow catalog explain"
      echo "workflow create"
      echo "workflow show"
      echo "workflow ready"
      echo "workflow node"
    else
      echo "unknown help topic" >&2
      exit 2
    fi
    ;;
  *)
    echo "unexpected kkachi-agent-helper call: $*" >&2
    exit 9
    ;;
esac
`
	if err := os.WriteFile(helper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func cliWorkflowRegistry(workflowID string) string {
	return `version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: ` + workflowID + `
    workflow_path: .kkachi/workflows/` + workflowID + `.yaml
    selector:
      task_classes: [development]
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
    task_class: development
    owner_role: planner_backend
    execution_lane: stage1_direct_codex_app_server
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/setup.md]
    prompt_ref: skills/kkachi-plan/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: kah_workflow_node_evidence
    completion_authority: kah_only
    direct_kah_state_write: false
`
}

func writeCLIWorkflowRouteResult(t *testing.T, dir string, registry workflowregistry.Registry, workflowID string) string {
	t.Helper()
	path := filepath.Join(dir, "route-result.json")
	payload := map[string]any{
		"ok":                    true,
		"command":               "workflow-route",
		"status":                "bundle_route_matched",
		"task_class":            "development",
		"classification_reason": "test route",
		"selected_bundle":       workflowID,
		"workflow_id":           workflowID,
		"workflow_path":         ".kkachi/workflows/" + workflowID + ".yaml",
		"selected_spine":        workflowID,
		"taxonomy": map[string]any{
			"path":     "registries/task-taxonomy.yaml",
			"checksum": "sha256:taxonomy",
		},
		"selector_registry": map[string]any{
			"path":     registry.Path,
			"version":  registry.Version,
			"checksum": registry.Checksum,
		},
		"teal_applicability": map[string]any{
			"contract_version":               "design003.v1",
			"project_has_teal_lane":          false,
			"ui_ux_change":                   false,
			"teal_required":                  false,
			"derivation":                     "project_has_teal_lane && ui_ux_change",
			"teal_skip_reason":               "No UI/UX surface in this project/task.",
			"required_when_teal_required":    []string{"DESIGN_PLAN_GATE", "DESIGN_FIDELITY_REVIEW"},
			"missing_required_status":        "required_teal_verdict_missing",
			"ordinary_review_is_substitute":  false,
			"mar_review_is_substitute":       false,
			"backend_evidence_is_substitute": false,
			"helper_notes_are_substitute":    false,
		},
		"direct_kah_state_write": false,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func boolShellValue(value bool) string {
	if value {
		return "1"
	}
	return "0"
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
			for _, want := range []string{"list", "install", "doctor", "repair", "toolchain", "uninstall", "version"} {
				if !strings.Contains(out, want) {
					t.Fatalf("root help did not list public command %s: %q", want, out)
				}
			}
			if !strings.Contains(out, "Compatibility commands:") || !strings.Contains(out, "sync-project-kas") || strings.Contains(out, "migrate-project-kas") {
				t.Fatalf("root help did not list only supported compatibility commands: %q", out)
			}
			if strings.Contains(out, "update") || strings.Contains(out, "migrate-profile-skills") || !strings.Contains(out, "version  Print CLI version information") {
				t.Fatalf("root help must not expose removed update/migrate commands: %q", out)
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected root help on stdout only, got stderr=%q", stderr.String())
			}
		})
	}
}

func TestRemovedUpdateAndMigrateCommandsFailClosed(t *testing.T) {
	for _, args := range [][]string{
		{"update", "--help"},
		{"update", "plugin", "--dry-run"},
		{"migrate-profile-skills", "--dry-run"},
		{"migrate-project-kas", "--dry-run"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Main(args, &stdout, &stderr, nil)
			if code != 2 {
				t.Fatalf("expected removed command to fail closed, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			if stdout.Len() != 0 {
				t.Fatalf("expected removed command failure on stderr only, stdout=%s", stdout.String())
			}
			if !strings.Contains(stderr.String(), "unknown_command") {
				t.Fatalf("expected unknown_command evidence for removed command, got %s", stderr.String())
			}
		})
	}
}

func TestToolchainHelpExposesTOLMRCommandSurface(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Main([]string{"toolchain", "--help"}, &stdout, &stderr, map[string]string{})
	if code != 0 {
		t.Fatalf("toolchain help failed: code=%d stderr=%s", code, stderr.String())
	}
	for _, want := range []string{"Usage of toolchain:", "init", "doctor", "refresh"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("expected toolchain help to expose %q, got %q", want, stdout.String())
		}
	}
	for _, want := range []string{"import-legacy", "set-stage", "--approval-evidence", "install-launchers", "--bin-dir"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("expected TOLMR-003/toolchain launcher help to expose %q, got %q", want, stdout.String())
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected help on stdout only, got stderr=%q", stderr.String())
	}
}

func TestToolchainInstallLaunchersCLIWritesEmbeddedWrappers(t *testing.T) {
	binDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Main([]string{"toolchain", "install-launchers", "--bin-dir", binDir, "--json"}, &stdout, &stderr, map[string]string{})
	if code != 0 {
		t.Fatalf("install-launchers failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected JSON success on stdout only, got stderr=%q", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "toolchain install-launchers" || payload["wrote"] != true {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	launchers, ok := payload["launchers"].([]any)
	if !ok || len(launchers) != 2 {
		t.Fatalf("expected two launchers: %+v", payload)
	}
	for _, name := range []string{"kkachi-agent-skills-toolchain", "kkachi-agent-helper-toolchain"} {
		path := filepath.Join(binDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("missing installed launcher %s: %v", name, err)
		}
		if !strings.Contains(string(data), "kkachi.toolchain.v1") || !strings.Contains(string(data), "--toolchain-status") {
			t.Fatalf("launcher %s missing v1/status support:\n%s", name, string(data))
		}
	}
}

func TestToolchainDoctorCLIJSONFailsClosedWhenMissing(t *testing.T) {
	projectRoot := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Main([]string{"toolchain", "doctor", "--project-root", projectRoot, "--json"}, &stdout, &stderr, map[string]string{})
	if code == 0 {
		t.Fatalf("expected missing toolchain failure, stdout=%s", stdout.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected failure JSON on stderr only, stdout=%s", stdout.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "toolchain_missing")
	if _, err := os.Stat(filepath.Join(projectRoot, ".kkachi", "toolchain.yaml")); !os.IsNotExist(err) {
		t.Fatalf("doctor must not create toolchain.yaml, stat err=%v", err)
	}
}

func TestToolchainTOLMR003CommandsDispatchAndFailClosed(t *testing.T) {
	t.Run("import_legacy_requires_profile", func(t *testing.T) {
		projectRoot := t.TempDir()
		var stdout, stderr bytes.Buffer
		code := Main([]string{"toolchain", "import-legacy", "--project-root", projectRoot, "--project", "kan-plugin", "--json"}, &stdout, &stderr, map[string]string{})
		if code == 0 {
			t.Fatalf("expected import-legacy validation failure, stdout=%s", stdout.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("expected import-legacy failure on stderr only, stdout=%s", stdout.String())
		}
		assertCLIErrorCode(t, stderr.Bytes(), "profile_required")
		if _, err := os.Stat(filepath.Join(projectRoot, ".kkachi", "toolchain.yaml")); !os.IsNotExist(err) {
			t.Fatalf("import-legacy validation failure must not create toolchain.yaml, stat err=%v", err)
		}
	})

	t.Run("set_stage_requires_existing_toolchain", func(t *testing.T) {
		projectRoot := t.TempDir()
		var stdout, stderr bytes.Buffer
		code := Main([]string{"toolchain", "set-stage", "--project-root", projectRoot, "--stage", "1", "--approval-evidence", "approval:t-1", "--json"}, &stdout, &stderr, map[string]string{})
		if code == 0 {
			t.Fatalf("expected set-stage fail-closed result, stdout=%s", stdout.String())
		}
		if stdout.Len() != 0 {
			t.Fatalf("expected set-stage failure on stderr only, stdout=%s", stdout.String())
		}
		assertCLIErrorCode(t, stderr.Bytes(), "toolchain_missing")
		if _, err := os.Stat(filepath.Join(projectRoot, ".kkachi", "toolchain.yaml")); !os.IsNotExist(err) {
			t.Fatalf("set-stage missing-toolchain failure must not create toolchain.yaml, stat err=%v", err)
		}
	})
}

func TestRootVersionExitsZeroAndPrintsVersion(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"-version"}, {"version"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Main(args, &stdout, &stderr, nil)
			if code != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			out := strings.TrimSpace(stdout.String())
			if out != "kkachi-agent-skills 0.1.9" {
				t.Fatalf("unexpected version output: %q", out)
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected version on stdout only, got stderr=%q", stderr.String())
			}
		})
	}
}

func TestVersionJSONIncludesBuildMetadata(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Main([]string{"version", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "version" || payload["cli_version"] != "0.1.9" {
		t.Fatalf("unexpected version payload: %+v", payload)
	}
	if payload["module_path"] == "" || payload["module_version"] == "" || payload["git_commit"] == nil || payload["dirty"] == nil {
		t.Fatalf("version payload missing build metadata: %+v", payload)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected version JSON on stdout only, got stderr=%q", stderr.String())
	}
}

func TestSubcommandHelpExitsZero(t *testing.T) {
	for _, command := range []string{"list", "install", "doctor", "repair", "uninstall", "sync-project-kas", "install-project-kas"} {
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

func TestRepairWorkflowGraphProposeAndApplyJSON(t *testing.T) {
	installCLIWorkflowGraphRepairFakeKAH(t)
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, ".kkachi-workflow.yaml"), []byte("version: workflow-graph/v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"repair", "--repo", cliRepoRoot(t), "--project", project, "--workflow-graph", "--propose", "--reason", "repair", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow graph propose failed: code=%d stderr=%s", code, stderr.String())
	}
	var proposal map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &proposal); err != nil {
		t.Fatal(err)
	}
	if proposal["ok"] != true || proposal["mode"] != "workflow_graph_repair_propose" || proposal["status"] != "proposal_available" {
		t.Fatalf("unexpected proposal payload: %+v", proposal)
	}
	if proposal["proposal"].(map[string]any)["id"] != "prop-1" || proposal["semantic_diff"].(map[string]any)["state"] != "completed" {
		t.Fatalf("proposal missing normalized evidence: %+v", proposal)
	}
	if !strings.Contains(proposal["next_command"].(string), "--apply-proposal prop-1 --approval <approval-ref>") {
		t.Fatalf("proposal missing safe next command: %+v", proposal)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair", "--repo", cliRepoRoot(t), "--project", project, "--workflow-graph", "--apply-proposal", "prop-1", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing approval to fail closed, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var missingApproval map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &missingApproval); err != nil {
		t.Fatal(err)
	}
	if missingApproval["status"] != "blocked_for_approval" {
		t.Fatalf("unexpected missing approval payload: %+v", missingApproval)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair", "--repo", cliRepoRoot(t), "--project", project, "--workflow-graph", "--apply-proposal", "prop-1", "--approval", "approved:1", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow graph apply failed: code=%d stderr=%s", code, stderr.String())
	}
	var applied map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil {
		t.Fatal(err)
	}
	if applied["ok"] != true || applied["mode"] != "workflow_graph_repair_apply" || applied["status"] != "applied" {
		t.Fatalf("unexpected apply payload: %+v", applied)
	}
	if applied["post_apply"].(map[string]any)["graph_checksum"] != "sha256:new" {
		t.Fatalf("apply missing post-apply checksum: %+v", applied)
	}
}

func TestRepairWorkflowGraphRejectsProjectSuiteFlags(t *testing.T) {
	installCLIWorkflowGraphRepairFakeKAH(t)
	var stdout, stderr bytes.Buffer
	code := Main([]string{"repair", "--repo", cliRepoRoot(t), "--project", t.TempDir(), "--workflow-graph", "--propose", "--reason", "repair", "--profile", "kwanwoo", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected mixed workflow/project repair flags to fail, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "workflow_graph_repair_mode_ambiguous")
}

func TestWorkflowTriggerCLIRequiresExplicitInputs(t *testing.T) {
	for name, tc := range map[string]struct {
		args     []string
		wantCode string
	}{
		"workflow-id": {
			args:     []string{"workflow-trigger", "--project", t.TempDir(), "--node-contract-source", "contracts.json", "--run", "run-20260615T010203Z-abcdef123456", "--json"},
			wantCode: "workflow_id_required",
		},
		"node-contract-source": {
			args:     []string{"workflow-trigger", "--project", t.TempDir(), "--workflow-id", "demo", "--run", "run-20260615T010203Z-abcdef123456", "--json"},
			wantCode: "node_contract_source_required",
		},
		"run-or-instance": {
			args:     []string{"workflow-trigger", "--project", t.TempDir(), "--workflow-id", "demo", "--node-contract-source", "contracts.json", "--json"},
			wantCode: "run_or_instance_required",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Main(tc.args, &stdout, &stderr, nil)
			if code != 2 {
				t.Fatalf("expected exit 2, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			assertCLIErrorCode(t, stderr.Bytes(), tc.wantCode)
		})
	}
}

func TestWorkflowTriggerCLIJSONSuccessAndBlockerRouting(t *testing.T) {
	project := t.TempDir()
	source := filepath.Join(project, "node-contracts.json")
	if err := os.WriteFile(source, []byte(`{
  "schema_version": "kas-node-contracts/v1",
  "contracts": [
    {
      "workflow_id": "demo",
      "node_id": "setup",
      "owner_role": "implementer_backend",
      "execution_lane": "direct_kas_skill",
      "required_inputs": ["task-contract.yaml"],
      "expected_artifacts": ["artifacts/setup.md"],
      "prompt_ref": "skills/kkachi-implement/SKILL.md",
      "approval_required": false,
      "fallback_policy": "none_fail_closed",
      "verification_gate": "make test"
    }
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	installCLIWorkflowTriggerFakeKAH(t, true)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-trigger", "--project", project, "--workflow-id", "demo", "--node-contract-source", source, "--run", "run-20260615T010203Z-abcdef123456", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow trigger failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected success JSON on stdout only, got stderr=%q", stderr.String())
	}
	var success map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &success); err != nil {
		t.Fatal(err)
	}
	if success["ok"] != true || success["command"] != "workflow-trigger" || success["status"] != "dispatch_packets_rendered" || success["direct_kah_state_write"] != false {
		t.Fatalf("unexpected success payload: %+v", success)
	}
	if packets := success["dispatch_packets"].([]any); len(packets) != 1 || packets[0].(map[string]any)["node_id"] != "setup" {
		t.Fatalf("missing dispatch packet: %+v", success)
	}

	installCLIWorkflowTriggerFakeKAH(t, false)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-trigger", "--project", project, "--workflow-id", "demo", "--node-contract-source", source, "--run", "run-20260615T010203Z-abcdef123456", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected capability blocker exit 2, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected blocker JSON on stderr only, got stdout=%q", stdout.String())
	}
	var blocked map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &blocked); err != nil {
		t.Fatal(err)
	}
	if blocked["ok"] != false || blocked["status"] != "blocked_missing_kah_workflow_capability" || blocked["direct_kah_state_write"] != false {
		t.Fatalf("unexpected blocker payload: %+v", blocked)
	}
	if packets, ok := blocked["dispatch_packets"].([]any); !ok || len(packets) != 0 {
		t.Fatalf("blocker must not include dispatch packets: %+v", blocked)
	}
}

func TestWorkflowTriggerCLISelectorRegistrySuccess(t *testing.T) {
	project := t.TempDir()
	registry := filepath.Join(project, "workflow-registry.yaml")
	if err := os.WriteFile(registry, []byte(`version: kas-task-dag-workflow-registry/v1
workflows:
  - workflow_id: demo
    workflow_path: .kkachi/workflows/demo.yaml
    selector:
      task_classes: [development]
      labels_any: []
      labels_all: []
      changed_surfaces_any: [code]
      risk_levels: []
      required_agents_all: []
      required_capabilities_all: [task_dag_schema_validation, workflow_instance_state]
    fallback_policy: none_fail_closed
node_contracts:
  - workflow_id: demo
    node_id: setup
    task_class: development
    owner_role: implementer_backend
    execution_lane: direct_kas_skill
    required_inputs: [task-contract.yaml]
    expected_artifacts: [artifacts/setup.md]
    prompt_ref: skills/kkachi-implement/SKILL.md
    approval_required: false
    fallback_policy: none_fail_closed
    verification_gate: make test
    completion_authority: kah_only
    direct_kah_state_write: false
`), 0o644); err != nil {
		t.Fatal(err)
	}
	installCLIWorkflowTriggerFakeKAH(t, true)
	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-trigger", "--project", project, "--selector-registry", registry, "--task-class", "development", "--changed-surfaces", "code", "--required-capability", "task_dag_schema_validation,workflow_instance_state", "--run", "run-20260615T010203Z-abcdef123456", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("selector workflow trigger failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["workflow_id"] != "demo" || payload["direct_kah_state_write"] != false {
		t.Fatalf("unexpected selector payload: %+v", payload)
	}
	selectorMatch := payload["selector_match"].(map[string]any)
	if selectorMatch["status"] != "selector_matched" || selectorMatch["workflow_id"] != "demo" {
		t.Fatalf("missing selector match readback: %+v", payload)
	}
	packet := payload["dispatch_packets"].([]any)[0].(map[string]any)
	if packet["selector_registry_checksum"] == "" || packet["completion_authority"] != "kah_only" || packet["stage1_direct_codex_is_kab_native_codex"] != false {
		t.Fatalf("missing packet registry/authority readback: %+v", packet)
	}
}

func TestWorkflowTriggerCLIWorkflowManagedRejectsBypass(t *testing.T) {
	project := t.TempDir()
	source := filepath.Join(project, "node-contracts.json")
	if err := os.WriteFile(source, []byte(`{
  "schema_version": "kas-node-contracts/v1",
  "contracts": [
    {
      "workflow_id": "demo",
      "node_id": "setup",
      "owner_role": "implementer_backend",
      "execution_lane": "direct_kas_skill",
      "required_inputs": ["task-contract.yaml"],
      "expected_artifacts": ["artifacts/setup.md"],
      "prompt_ref": "skills/kkachi-implement/SKILL.md",
      "approval_required": false,
      "fallback_policy": "none_fail_closed",
      "verification_gate": "make test"
    }
  ]
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	installCLIWorkflowTriggerFakeKAH(t, true)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-trigger", "--project", project, "--workflow-id", "demo", "--node-contract-source", source, "--run", "run-20260615T010203Z-abcdef123456", "--workflow-managed", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected strict workflow-managed blocker, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "strict_workflow_route_result_required")
	var blocked map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &blocked); err != nil {
		t.Fatal(err)
	}
	if packets, ok := blocked["dispatch_packets"].([]any); !ok || len(packets) != 0 {
		t.Fatalf("strict blocker must not render packets: %+v", blocked)
	}
	nextAction, _ := blocked["next_action"].(string)
	for _, want := range []string{"workflow-route", "--route-result", "--materialize-run-local", "--workflow-managed", "route-backed run-local materialization evidence"} {
		if !strings.Contains(nextAction, want) {
			t.Fatalf("strict next_action missing %q: %q", want, nextAction)
		}
	}
	if strings.Contains(nextAction, "explicit workflow and node-contract inputs") {
		t.Fatalf("strict next_action must not steer operators to explicit workflow fallback: %q", nextAction)
	}
}

func TestWorkflowRouteCLIRoutesClassifiedTaskWithoutKAH(t *testing.T) {
	root := cliRepoRoot(t)
	taxonomy := filepath.Join(root, "registries", "task-taxonomy.yaml")
	registry := filepath.Join(root, "registries", "task-dag-workflow-registry.yaml")

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-route", "--taxonomy", taxonomy, "--selector-registry", registry, "--task-class", "development", "--classification-reason", "KAH classified WFLOW-007 as development.", "--selected-spine", "development_full", "--project-has-teal-lane", "false", "--ui-ux-change", "false", "--teal-skip-reason", "No UI/UX surface in this project/task.", "--required-capability", "task_dag_schema_validation,workflow_instance_state", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow-route failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected route success on stdout only, got stderr=%q", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "workflow-route" || payload["status"] != "bundle_route_matched" || payload["selected_bundle"] != "development_full" || payload["direct_kah_state_write"] != false {
		t.Fatalf("unexpected workflow-route payload: %+v", payload)
	}
	if _, ok := payload["dispatch_packets"]; ok {
		t.Fatalf("workflow-route must not render dispatch packets: %+v", payload)
	}
	teal := payload["teal_applicability"].(map[string]any)
	if teal["teal_required"] != false || teal["teal_skip_reason"] != "No UI/UX surface in this project/task." {
		t.Fatalf("workflow-route must preserve non-UI Teal skip evidence: %+v", teal)
	}
}

func TestWorkflowRouteCLIMatchesDesign006GoldenScenarios(t *testing.T) {
	root := cliRepoRoot(t)
	taxonomy := filepath.Join(root, "registries", "task-taxonomy.yaml")
	registry := filepath.Join(root, "registries", "task-dag-workflow-registry.yaml")
	scenarios := readDesign006CLIScenarios(t, filepath.Join(root, "docs", "examples", "design006-teal-compatibility-scenarios.json"))

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.ID, func(t *testing.T) {
			args := []string{
				"workflow-route",
				"--taxonomy", taxonomy,
				"--selector-registry", registry,
				"--task-class", "development",
				"--classification-reason", "DESIGN-006 golden compatibility scenario.",
				"--selected-spine", "development_full",
				"--project-has-teal-lane", boolArg(scenario.ProjectHasTealLane),
				"--ui-ux-change", boolArg(scenario.UIUXChange),
				"--required-capability", "task_dag_schema_validation,workflow_instance_state",
				"--json",
			}
			if scenario.TealSkipReason != "" {
				args = append(args[:len(args)-1], "--teal-skip-reason", scenario.TealSkipReason, "--json")
			}

			var stdout, stderr bytes.Buffer
			code := Main(args, &stdout, &stderr, nil)
			if code != 0 {
				t.Fatalf("workflow-route failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			var payload map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			teal := payload["teal_applicability"].(map[string]any)
			if teal["teal_required"] != scenario.TealRequired {
				t.Fatalf("teal_applicability = %+v, want teal_required=%t", teal, scenario.TealRequired)
			}
			if !scenario.TealRequired && teal["teal_skip_reason"] != scenario.TealSkipReason {
				t.Fatalf("teal_applicability = %+v, want skip reason %q", teal, scenario.TealSkipReason)
			}
			for _, field := range []string{"ordinary_review_is_substitute", "mar_review_is_substitute", "backend_evidence_is_substitute", "helper_notes_are_substitute"} {
				if teal[field] != false {
					t.Fatalf("teal_applicability = %+v, want %s=false", teal, field)
				}
			}
		})
	}
}

func TestWorkflowRouteCLIFailsClosed(t *testing.T) {
	root := cliRepoRoot(t)
	taxonomy := filepath.Join(root, "registries", "task-taxonomy.yaml")
	registry := filepath.Join(root, "registries", "task-dag-workflow-registry.yaml")

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-route", "--taxonomy", taxonomy, "--selector-registry", registry, "--task-class", "development", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected fail-closed exit 2, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected failure JSON on stderr only, got stdout=%q", stdout.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != false || payload["status"] != "classification_reason_missing" || payload["direct_kah_state_write"] != false {
		t.Fatalf("unexpected workflow-route failure payload: %+v", payload)
	}
	if _, ok := payload["dispatch_packets"]; ok {
		t.Fatalf("workflow-route failure must not include dispatch packets: %+v", payload)
	}
}

type design006CLIScenario struct {
	ID                 string `json:"id"`
	ProjectHasTealLane bool   `json:"project_has_teal_lane"`
	UIUXChange         bool   `json:"ui_ux_change"`
	TealRequired       bool   `json:"teal_required"`
	TealSkipReason     string `json:"teal_skip_reason"`
}

func readDesign006CLIScenarios(t *testing.T, path string) []design006CLIScenario {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read DESIGN-006 scenarios: %v", err)
	}
	var payload struct {
		Version   string                 `json:"version"`
		Scenarios []design006CLIScenario `json:"scenarios"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode DESIGN-006 scenarios: %v", err)
	}
	if payload.Version != "design006.v1" {
		t.Fatalf("version = %q, want design006.v1", payload.Version)
	}
	return payload.Scenarios
}

func boolArg(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func TestWorkflowTriggerCLIRunLocalMaterialization(t *testing.T) {
	project := t.TempDir()
	registryPath := filepath.Join(project, "workflow-registry.yaml")
	if err := os.WriteFile(registryPath, []byte(cliWorkflowRegistry("demo")), 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeCLIWorkflowRouteResult(t, project, registry, "demo")
	installCLIWorkflowTriggerRunLocalFakeKAH(t, true)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-trigger", "--project", project, "--route-result", routeResult, "--materialize-run-local", "--run", "run-20260616T105614Z-4b0ebe11b67d", "--workflow-managed", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("run-local workflow trigger failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["mode"] != "run_local_materialized_trigger" || payload["workflow_id"] != "demo" || payload["direct_kah_state_write"] != false {
		t.Fatalf("unexpected run-local payload: %+v", payload)
	}
	workflow := payload["workflow"].(map[string]any)
	if workflow["path"] != ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml" {
		t.Fatalf("workflow path did not use run-local file: %+v", payload)
	}
	materialization := payload["materialization"].(map[string]any)
	if materialization["no_promotion"] != true || materialization["persistent_promotion"] != false || materialization["selected_bundle"] != "demo" {
		t.Fatalf("missing materialization evidence: %+v", materialization)
	}
	packet := payload["dispatch_packets"].([]any)[0].(map[string]any)
	if packet["workflow_file"] != workflow["path"] || packet["direct_kah_state_write"] != false || packet["completion_authority"] != "kah_only" || packet["fallback_policy"] != "none_fail_closed" {
		t.Fatalf("dispatch packet authority drift: %+v", packet)
	}
	for _, rel := range []string{"workflow.yaml", "node-contracts.json", "route-result.json", "materialization.json", "checksums.json"} {
		if _, err := os.Stat(filepath.Join(project, ".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow", rel)); err != nil {
			t.Fatalf("missing materialized %s: %v", rel, err)
		}
	}
}

func TestWorkflowTriggerCLIRunLocalUnsupportedKAHNoWrite(t *testing.T) {
	project := t.TempDir()
	registryPath := filepath.Join(project, "workflow-registry.yaml")
	if err := os.WriteFile(registryPath, []byte(cliWorkflowRegistry("demo")), 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeCLIWorkflowRouteResult(t, project, registry, "demo")
	installCLIWorkflowTriggerRunLocalFakeKAH(t, false)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-trigger", "--project", project, "--route-result", routeResult, "--materialize-run-local", "--run", "run-20260616T105614Z-4b0ebe11b67d", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected unsupported KAH blocker, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["status"] != "blocked_missing_kah_workflow_capability" || payload["direct_kah_state_write"] != false {
		t.Fatalf("unexpected unsupported payload: %+v", payload)
	}
	if _, ok := payload["materialization"]; ok {
		t.Fatalf("unsupported KAH payload must omit materialization: %+v", payload["materialization"])
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("unsupported KAH path wrote .kkachi before preflight passed: %v", err)
	}
}

func TestWorkflowTriggerCLICustomWorkflowPacketMaterialization(t *testing.T) {
	project := t.TempDir()
	request := writeCLIWorkflowCreateRequest(t, project)
	requestBytes, err := os.ReadFile(request)
	if err != nil {
		t.Fatal(err)
	}
	runLocalSetup := ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/artifacts/setup.md"
	requestText := strings.ReplaceAll(string(requestBytes), `"node_id": "plan"`, `"node_id": "setup"`)
	requestText = strings.ReplaceAll(requestText, `["artifacts/plan.md"]`, `["`+runLocalSetup+`"]`)
	requestText = strings.ReplaceAll(requestText, `["plan.md"]`, `["`+runLocalSetup+`"]`)
	requestText = strings.ReplaceAll(requestText, `["task-contract.yaml"]`, `[".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/task-contract.yaml"]`)
	if err := os.WriteFile(request, []byte(requestText), 0o644); err != nil {
		t.Fatal(err)
	}
	installCLIWorkflowCreateFakeKAH(t, true)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-create", "--project", project, "--workflow-id", "demo", "--mode", "dag_only", "--request", request, "--dry-run", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow-create dry-run failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	packetPath := filepath.Join(project, "workflow-create-dry-run.json")
	if err := os.WriteFile(packetPath, stdout.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}
	approval := dryRun["approval_request"].(map[string]any)["evidence_ref"].(string)

	installCLIWorkflowTriggerRunLocalFakeKAH(t, true)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-trigger", "--project", project, "--run", "run-20260616T105614Z-4b0ebe11b67d", "--custom-workflow-packet", packetPath, "--approval", approval, "--materialize-run-local", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("custom run-local workflow trigger failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["mode"] != "run_local_materialized_trigger" || payload["workflow_id"] != "demo" || payload["direct_kah_state_write"] != false {
		t.Fatalf("unexpected custom trigger payload: %+v", payload)
	}
	materialization := payload["materialization"].(map[string]any)
	if materialization["custom_workflow_packet_copy"] != ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/custom-workflow-packet.json" || materialization["approval_evidence"] != approval || materialization["no_promotion"] != true || materialization["persistent_promotion"] != false {
		t.Fatalf("missing custom materialization evidence: %+v", materialization)
	}
	packet := payload["dispatch_packets"].([]any)[0].(map[string]any)
	if packet["workflow_file"] != ".kkachi/runs/run-20260616T105614Z-4b0ebe11b67d/workflow/workflow.yaml" || packet["completion_authority"] != "kah_only" || packet["fallback_policy"] != "none_fail_closed" || packet["direct_kah_state_write"] != false {
		t.Fatalf("custom dispatch packet authority drift: %+v", packet)
	}
	for _, rel := range []string{"workflow.yaml", "node-contracts.json", "custom-workflow-packet.json", "materialization.json", "checksums.json"} {
		if _, err := os.Stat(filepath.Join(project, ".kkachi", "runs", "run-20260616T105614Z-4b0ebe11b67d", "workflow", rel)); err != nil {
			t.Fatalf("missing custom materialized %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi", "workflows")); !os.IsNotExist(err) {
		t.Fatalf("custom trigger promoted persistent workflow state: %v", err)
	}
}

func TestWorkflowTriggerCLICustomWorkflowPacketRequiresApprovalAndExclusiveSource(t *testing.T) {
	project := t.TempDir()
	packetPath := filepath.Join(project, "workflow-create-dry-run.json")
	if err := os.WriteFile(packetPath, []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	installCLIWorkflowTriggerRunLocalFakeKAH(t, true)
	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-trigger", "--project", project, "--run", "run-20260616T105614Z-4b0ebe11b67d", "--custom-workflow-packet", packetPath, "--materialize-run-local", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing approval failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "approval_evidence_required")
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("missing approval wrote .kkachi: %v", err)
	}

	registryPath := filepath.Join(project, "workflow-registry.yaml")
	if err := os.WriteFile(registryPath, []byte(cliWorkflowRegistry("demo")), 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeCLIWorkflowRouteResult(t, project, registry, "demo")
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-trigger", "--project", project, "--run", "run-20260616T105614Z-4b0ebe11b67d", "--route-result", routeResult, "--custom-workflow-packet", packetPath, "--approval", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--materialize-run-local", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected source conflict, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "run_local_materialization_source_required")
}

func TestWorkflowCreateCLIDryRunAndFailClosedApply(t *testing.T) {
	project := t.TempDir()
	request := writeCLIWorkflowCreateRequest(t, project)
	installCLIWorkflowCreateFakeKAH(t, true)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-create", "--project", project, "--workflow-id", "release-flow", "--mode", "thin_trigger", "--request", request, "--dry-run", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow-create dry-run failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected dry-run success on stdout only, got stderr=%q", stderr.String())
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}
	if dryRun["ok"] != true || dryRun["command"] != "workflow-create" || dryRun["status"] != "dry_run_ready" || dryRun["direct_kah_state_write"] != false {
		t.Fatalf("unexpected workflow-create dry-run payload: %+v", dryRun)
	}
	packet := dryRun["machine_packet"].(map[string]any)
	if packet["approval_hash"] == "" || packet["generated_content"] == nil || packet["candidate_paths"] == nil {
		t.Fatalf("dry-run missing machine packet evidence: %+v", dryRun)
	}
	noWrite := packet["no_write"].(map[string]any)
	if noWrite["guaranteed"] != true || noWrite["project_write_count"] != float64(0) || noWrite["kah_state_write_count"] != float64(0) || noWrite["profile_write_count"] != float64(0) {
		t.Fatalf("unexpected no-write evidence: %+v", noWrite)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("workflow-create dry-run created project state: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-create", "--project", project, "--workflow-id", "release-flow", "--mode", "thin_trigger", "--request", request, "--dry-run", "--apply", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected ambiguous mode failure, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "workflow_create_mode_ambiguous")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-create", "--project", project, "--workflow-id", "release-flow", "--mode", "thin_trigger", "--request", request, "--apply", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected wrong hash failure, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var wrongHash map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &wrongHash); err != nil {
		t.Fatal(err)
	}
	if wrongHash["status"] != "approval_plan_hash_mismatch" || wrongHash["approval"].(map[string]any)["matched_current_plan"] != false {
		t.Fatalf("unexpected wrong-hash payload: %+v", wrongHash)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("wrong hash created project state: %v", err)
	}
}

func TestWorkflowCreateCLIInstalledKAHCaveatIsNonApprovable(t *testing.T) {
	project := t.TempDir()
	request := writeCLIWorkflowCreateRequest(t, project)
	installCLIWorkflowCreateFakeKAH(t, false)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-create", "--project", project, "--workflow-id", "release-flow", "--mode", "dag_only", "--request", request, "--dry-run", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing installed KAH workflow capability blocker, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var blocked map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &blocked); err != nil {
		t.Fatal(err)
	}
	if blocked["status"] != "blocked_missing_kah_workflow_capability" || blocked["approval_request"].(map[string]any)["required"] != false {
		t.Fatalf("unexpected installed-KAH caveat payload: %+v", blocked)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("blocked dry-run created project state: %v", err)
	}
}

func TestWorkflowPromoteCLIDryRunAndFailClosedApply(t *testing.T) {
	project := t.TempDir()
	registryPath := filepath.Join(project, "workflow-registry.yaml")
	if err := os.WriteFile(registryPath, []byte(cliWorkflowRegistry("demo")), 0o644); err != nil {
		t.Fatal(err)
	}
	registry, err := workflowregistry.Load(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	routeResult := writeCLIWorkflowRouteResult(t, project, registry, "demo")
	installCLIWorkflowTriggerRunLocalFakeKAH(t, true)

	var stdout, stderr bytes.Buffer
	code := Main([]string{"workflow-trigger", "--project", project, "--route-result", routeResult, "--materialize-run-local", "--run", "run-20260616T105614Z-4b0ebe11b67d", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow-trigger materialization failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	installCLIWorkflowCreateFakeKAH(t, true)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-promote", "--project", project, "--run", "run-20260616T105614Z-4b0ebe11b67d", "--target-workflow-id", "promoted-demo", "--reuse-reason", "demo workflow proved reusable for repeat release tasks", "--thin-trigger", "--dry-run", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow-promote dry-run failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}
	if dryRun["ok"] != true || dryRun["command"] != "workflow-promote" || dryRun["status"] != "dry_run_ready" || dryRun["direct_kah_state_write"] != false {
		t.Fatalf("unexpected workflow-promote dry-run payload: %+v", dryRun)
	}
	packet := dryRun["machine_packet"].(map[string]any)
	if packet["approval_hash"] == "" || packet["source_provenance"] == nil || packet["generated_content"] == nil || packet["target_paths"] == nil {
		t.Fatalf("dry-run missing promotion machine packet evidence: %+v", packet)
	}
	source := packet["source_provenance"].(map[string]any)
	if source["run_id"] != "run-20260616T105614Z-4b0ebe11b67d" || source["workflow_checksum"] == "" || source["node_contracts_checksum"] == "" || source["materialization_checksum"] == "" {
		t.Fatalf("missing source provenance: %+v", source)
	}
	noWrite := packet["no_write"].(map[string]any)
	if noWrite["guaranteed"] != true || noWrite["project_write_count"] != float64(0) || noWrite["kah_state_write_count"] != float64(0) || noWrite["kab_runtime_mutation_count"] != float64(0) {
		t.Fatalf("unexpected no-write evidence: %+v", noWrite)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi", "workflows")); !os.IsNotExist(err) {
		t.Fatalf("workflow-promote dry-run created project workflow state: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-promote", "--project", project, "--run", "run-20260616T105614Z-4b0ebe11b67d", "--target-workflow-id", "promoted-demo", "--reuse-reason", "demo workflow proved reusable for repeat release tasks", "--dry-run", "--apply", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected ambiguous mode failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "workflow_promote_mode_ambiguous")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-promote", "--project", project, "--run", "run-20260616T105614Z-4b0ebe11b67d", "--target-workflow-id", "promoted-demo", "--reuse-reason", "demo workflow proved reusable for repeat release tasks", "--apply", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected wrong hash failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var wrongHash map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &wrongHash); err != nil {
		t.Fatal(err)
	}
	if wrongHash["status"] != "approval_plan_hash_mismatch" || wrongHash["approval"].(map[string]any)["matched_current_plan"] != false {
		t.Fatalf("unexpected wrong-hash payload: %+v", wrongHash)
	}

	approval := dryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"workflow-promote", "--project", project, "--run", "run-20260616T105614Z-4b0ebe11b67d", "--target-workflow-id", "promoted-demo", "--reuse-reason", "demo workflow proved reusable for repeat release tasks", "--thin-trigger", "--apply", approval, "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected correct-hash fail-closed apply, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var blocked map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &blocked); err != nil {
		t.Fatal(err)
	}
	if blocked["status"] != "blocked_missing_kah_workflow_catalog_capability" || blocked["approval"].(map[string]any)["matched_current_plan"] != true {
		t.Fatalf("unexpected correct-hash blocked payload: %+v", blocked)
	}
	for _, forbidden := range []string{".kkachi/workflows", ".kkachi/workflow-catalog.yaml", ".kkachi-workflow.yaml"} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(forbidden))); !os.IsNotExist(err) {
			t.Fatalf("workflow-promote apply wrote forbidden path %s: %v", forbidden, err)
		}
	}
}

func writeCLIWorkflowCreateRequest(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "workflow-create-request.json")
	content := `{
  "schema_version": "kas-workflow-create-request/v1",
  "selector_metadata": {
    "task_class": "development",
    "labels": ["release"],
    "changed_surfaces": ["code"],
    "required_capabilities": ["task_dag_schema_validation", "workflow_instance_state"]
  },
  "nodes": [
    {
      "node_id": "plan",
      "task_class": "development",
      "depends_on": [],
      "required_outputs": ["artifacts/plan.md"],
      "owner_role": "planner_backend",
      "execution_lane": "stage1_direct_codex_app_server",
      "required_inputs": ["task-contract.yaml"],
      "expected_artifacts": ["plan.md"],
      "prompt_ref": "skills/kkachi-plan/SKILL.md",
      "approval_required": false,
      "fallback_policy": "none_fail_closed",
      "verification_gate": "kah_workflow_node_evidence"
    }
  ],
  "trigger": {
    "name": "release-flow-trigger",
    "description": "Release flow trigger"
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDoctorWorkflowGraphJSONAndFlagValidation(t *testing.T) {
	installCLITestFakeKAH(t)
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, ".kkachi-workflow.yaml"), []byte("version: workflow-graph/v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Main([]string{"doctor", "--repo", cliRepoRoot(t), "--project", project, "--workflow-graph", "--json"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("workflow graph doctor code=%d stderr=%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected workflow graph doctor on stdout only, got stderr=%q", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "doctor" || payload["mode"] != "workflow_graph_doctor" || payload["no_write"] != true || payload["status"] != "pass" {
		t.Fatalf("unexpected workflow graph doctor payload: %+v", payload)
	}
	if payload["kas"].(map[string]any)["cli_version"] != "0.1.9" || payload["kah"].(map[string]any)["graph_help_state"] != "ok" {
		t.Fatalf("missing KAS/KAH evidence: %+v", payload)
	}
	for _, raw := range payload["kah"].(map[string]any)["compatibility_flags"].([]any) {
		flag := raw.(map[string]any)["flag"]
		if flag == "graph_validate_json" || flag == "graph_explain_json" {
			t.Fatalf("validate/explain JSON support should be command evidence, not pseudo-flag evidence: %+v", payload)
		}
	}
	if payload["graph"].(map[string]any)["schema_version"] != "workflow-graph/v1" {
		t.Fatalf("missing graph evidence: %+v", payload)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"doctor", "--repo", cliRepoRoot(t), "--workflow-graph", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing project failure, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "project_required")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"doctor", "--repo", cliRepoRoot(t), "--project", project, "--workflow-graph", "--project-suite", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected ambiguous doctor mode failure, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "doctor_mode_ambiguous")
}

func TestPublicLifecycleWrappersJSONDryRunOnly(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profile")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
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
	code = Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "blue_commander", "--dry-run", "--apply", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
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
			args:       []string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot},
			wantCode:   0,
			wantStdout: true,
			labels:     []string{"Status:", "Source pack:", "Role:", "Plan:", "Writes:", "Approval evidence:", "Next:"},
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
			labels:   []string{"Status:", "manifest:", "suite:", "source_pack:", "project-suite diagnostic:", "Next:"},
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
	code := Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
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

func TestInstallProjectPublicSuiteRoleSurfaceAndMissingRegistry(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profiles", "kwanwoo")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "blue_commander", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected missing mode exit 2, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "project_install_requires_dry_run_or_apply")
	var errPayload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &errPayload); err != nil {
		t.Fatal(err)
	}
	if next, _ := errPayload["next_action"].(string); !strings.Contains(next, "--suite-role <role>") {
		t.Fatalf("missing mode guidance should preserve suite-role placeholder: %q", next)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("public project install dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "install" || payload["suite_role"] != "blue_commander" {
		t.Fatalf("unexpected public install payload: %+v", payload)
	}
	next, _ := payload["next_action"].(string)
	if !strings.Contains(next, "--suite-role blue_commander") || !strings.Contains(next, "--apply dry-run:sha256:") {
		t.Fatalf("next action must preserve --suite-role and approval evidence: %q", next)
	}
	approval := payload["approval_request"].(map[string]any)
	if approval["hash_includes_role_fields"] != true {
		t.Fatalf("approval hash must advertise role-field composition: %+v", approval)
	}

	if err := os.Remove(filepath.Join(repo, filepath.FromSlash(projectinstall.RoleRegistryPath))); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected missing registry exit 2, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "role_registry_unreadable")
}

func TestInstallProjectKASApprovedInstallAndFailClosedGuards(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profile")
	base := []string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--profile-root", profileRoot, "--json"}

	var stdout, stderr bytes.Buffer
	code := Main(base, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected missing dry-run exit 2, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "project_install_requires_dry_run_or_approve")
	if !strings.Contains(stderr.String(), "--suite-role") {
		t.Fatalf("missing mode next action should preserve suite-role guidance: %s", stderr.String())
	}

	for name, args := range map[string][]string{
		"profile":     {"install-project-kas", "--repo", repo, "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--dry-run", "--json"},
		"project":     {"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--dry-run", "--json"},
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
	code = Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected missing suite-role rejection, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "suite_role_required")

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
	if !strings.Contains(stderr.String(), "--suite-role") {
		t.Fatalf("unsupported write next action should preserve suite-role guidance: %s", stderr.String())
	}

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
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-wrapper", "SKILL.md")); err != nil {
		t.Fatalf("approved install did not write project wrapper: %v", err)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("plugin-only approved install wrote copied project skill or unexpected stat error: %v", err)
	}
}

func TestInstallProjectKASConflictJSONExitsTwo(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-kas"), "kkachi-kas")
	writeCLITestSkillPackYAML(t, repo, "kkachi-kas")
	profileRoot := filepath.Join(dir, "profile")
	var stdout, stderr bytes.Buffer
	code := Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 0 {
		t.Fatalf("plugin-only umbrella source dry-run should not copy project skills, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || len(payload["planned_skills"].([]any)) != 0 || len(payload["composition_files"].([]any)) != 3 {
		t.Fatalf("plugin-only payload should plan only composition files: %+v", payload)
	}
}

func TestPublicRepairRoleAwarePrune(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-review"), "kkachi-review")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-verify"), "kkachi-verify")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-implement"), "kkachi-implement")
	writeCLITestSkillPackYAML(t, repo, "kkachi-review", "kkachi-verify", "kkachi-implement")
	profileRoot := filepath.Join(dir, "profile")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("install dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var installDryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &installDryRun); err != nil {
		t.Fatal(err)
	}
	evidence := installDryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "blue_commander", "--apply", evidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("install apply failed: code=%d stderr=%s", code, stderr.String())
	}
	setCLIManifestSuiteRole(t, profileRoot, "doksuri-server", "red_reviewer")
	addCLILegacyCopiedProjectSkill(t, profileRoot, "doksuri-server", "kkachi-implement")
	writeCLITestSkill(t, filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-local-note"), "doksuri-server-local-note")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--prune-extra", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected missing suite-role fail closed, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "suite_role_required")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "red_reviewer", "--prune-extra", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("repair prune dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}
	if dryRun["command"] != "repair" || dryRun["suite_role"] != "red_reviewer" || dryRun["prune_extra"] != true {
		t.Fatalf("unexpected repair dry-run payload: %+v", dryRun)
	}
	summary := dryRun["summary"].(map[string]any)
	counts := summary["counts_by_action"].(map[string]any)
	if counts["remove"] != float64(1) {
		t.Fatalf("unexpected repair prune counts: %+v", summary)
	}
	noSpillover := dryRun["no_spillover"].(map[string]any)
	if len(noSpillover["unknown_personal_skills_preserved"].([]any)) != 1 {
		t.Fatalf("missing no-spillover evidence: %+v", noSpillover)
	}
	assertNoWriteEvidence(t, dryRun)
	repairEvidence := dryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	nextAction := dryRun["next_action"].(string)
	for _, want := range []string{"repair --profile kwanwoo --project doksuri-server --suite-role red_reviewer --prune-extra --apply " + repairEvidence, "--backup-vault-root <approved-abs-vault-root>"} {
		if !strings.Contains(nextAction, want) {
			t.Fatalf("public repair next_action missing %q: %s", want, nextAction)
		}
	}
	if strings.Contains(nextAction, "repair-project-kas") || strings.Contains(nextAction, "--approve") {
		t.Fatalf("public repair next_action must prefer public --apply flow, got: %s", nextAction)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "red_reviewer", "--prune-extra", "--apply", repairEvidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 2 {
		t.Fatalf("expected repair apply without backup vault root to fail, got %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "backup_vault_root_rejected")

	vault := filepath.Join(dir, "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--suite-role", "red_reviewer", "--prune-extra", "--apply", repairEvidence, "--backup-vault-root", vault, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("repair prune apply failed: code=%d stderr=%s", code, stderr.String())
	}
	var applied map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil {
		t.Fatal(err)
	}
	if applied["ok"] != true || applied["mode"] != "project_repair_approved" || applied["repair_id"] == "" {
		t.Fatalf("unexpected repair apply payload: %+v", applied)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-implement", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("repair apply did not prune implement skill or unexpected stat error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-local-note", "SKILL.md")); err != nil {
		t.Fatalf("repair apply touched unknown personal skill: %v", err)
	}
}

func TestRepairProjectKASRoleAwareCompatibility(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-review"), "kkachi-review")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-verify"), "kkachi-verify")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-implement"), "kkachi-implement")
	writeCLITestSkillPackYAML(t, repo, "kkachi-review", "kkachi-verify", "kkachi-implement")
	profileRoot := filepath.Join(dir, "profile")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("install-project-kas dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var installDryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &installDryRun); err != nil {
		t.Fatal(err)
	}
	evidence := installDryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--approve", evidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("install-project-kas approve failed: code=%d stderr=%s", code, stderr.String())
	}
	setCLIManifestSuiteRole(t, profileRoot, "doksuri-server", "red_reviewer")
	addCLILegacyCopiedProjectSkill(t, profileRoot, "doksuri-server", "kkachi-implement")

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "red_reviewer", "--prune-extra", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("repair-project-kas prune dry-run failed: code=%d stderr=%s", code, stderr.String())
	}
	var dryRun map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &dryRun); err != nil {
		t.Fatal(err)
	}
	if dryRun["command"] != "repair-project-kas" || dryRun["suite_role"] != "red_reviewer" || dryRun["prune_extra"] != true {
		t.Fatalf("compat dry-run missing role/prune fields: %+v", dryRun)
	}
	repairEvidence := dryRun["approval_request"].(map[string]any)["evidence_ref"].(string)
	vault := filepath.Join(dir, "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"repair-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "red_reviewer", "--prune-extra", "--approve", repairEvidence, "--backup-vault-root", vault, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("repair-project-kas approve failed: code=%d stderr=%s", code, stderr.String())
	}
	var applied map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil {
		t.Fatal(err)
	}
	if applied["ok"] != true || applied["mode"] != "project_repair_approved" {
		t.Fatalf("unexpected compat approved payload: %+v", applied)
	}
}

func setCLIManifestSuiteRole(t *testing.T, profileRoot string, project string, suiteRole string) {
	t.Helper()
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, rawSuite := range manifest["project_suites"].([]any) {
		suite := rawSuite.(map[string]any)
		if suite["project"] == project {
			suite["suite_role"] = suiteRole
			suite["suite_mode"] = "role_subset"
		}
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(encoded, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func addCLILegacyCopiedProjectSkill(t *testing.T, profileRoot string, project string, sourceSkill string) {
	t.Helper()
	installed := strings.TrimPrefix(sourceSkill, "kkachi-")
	installed = project + "-" + installed
	content := "---\nname: " + installed + "\n---\n# " + installed + "\n"
	target := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
	writeCLITestSkill(t, filepath.Join(profileRoot, "skills", project, installed), installed)
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, rawSuite := range manifest["project_suites"].([]any) {
		suite := rawSuite.(map[string]any)
		if suite["project"] != project {
			continue
		}
		rawSkills := suite["installed_skills"].([]any)
		sum := sha256.Sum256([]byte(content))
		suite["installed_skills"] = append(rawSkills, map[string]any{
			"source_skill":    sourceSkill,
			"source_pack_id":  sourceSkill,
			"installed_skill": installed,
			"target_path":     target,
			"checksum":        fmt.Sprintf("sha256:%x", sum),
			"drift_policy":    "manual_review_required",
			"tailoring_mode":  "prefix_render_only",
		})
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(encoded, '\n'), 0o644); err != nil {
		t.Fatal(err)
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

func TestDoctorPluginModeJSONHumanAndAmbiguousFlags(t *testing.T) {
	repo := t.TempDir()
	writeCLIPluginUpdateFixture(t, repo)
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeCLISkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"))
	writeCLISkillDoctorOverlay(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "doksuri-blue-plan-overlay"), "kkachi-agent-skills:plan")
	before := countRegularFiles(t, profileRoot)

	var stdout, stderr bytes.Buffer
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}
	code := Main([]string{"doctor", "--plugin", "--repo", repo, "--profile", "demo", "--profile-root", profileRoot, "--project", "doksuri", "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("doctor --plugin failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["command"] != "doctor" || payload["mode"] != "skill_plugin_overlay_doctor" || payload["ok"] != true {
		t.Fatalf("unexpected plugin doctor payload: %+v", payload)
	}
	if payload["no_write"].(map[string]any)["guaranteed"] != true {
		t.Fatalf("missing no-write evidence: %+v", payload)
	}
	for _, field := range []string{"provenance_contract_version", "source_class_evidence", "dependency_audit", "skill_dependencies", "command_surface_dependencies", "deleted_bundle_reference", "deleted_bundle_diagnostics"} {
		if _, ok := payload[field]; !ok {
			t.Fatalf("missing KASREL field %s in %+v", field, payload)
		}
	}
	for _, field := range []string{"source_class_evidence", "skill_dependencies", "command_surface_dependencies", "deleted_bundle_diagnostics"} {
		if _, ok := payload[field].([]any); !ok {
			t.Fatalf("KASREL field %s should be an explicit array: %+v", field, payload[field])
		}
	}
	counts := payload["summary"].(map[string]any)["counts_by_source_class"].(map[string]any)
	for _, sourceClass := range []string{"plugin_base", "color_wrapper", "project_overlay"} {
		if counts[sourceClass].(float64) == 0 {
			t.Fatalf("missing source class %s in %+v", sourceClass, counts)
		}
	}
	if got := countRegularFiles(t, profileRoot); got != before {
		t.Fatalf("doctor --plugin wrote profile files: before=%d after=%d", before, got)
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"doctor", "--plugin", "--repo", repo, "--profile", "demo", "--profile-root", profileRoot, "--no-color"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("doctor --plugin human failed: code=%d stderr=%s", code, stderr.String())
	}
	for _, want := range []string{"Status:", "SKILL plugin/wrapper/overlay doctor", "Writes:", "Next:"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("human output missing %q: %s", want, stdout.String())
		}
	}

	stdout.Reset()
	stderr.Reset()
	code = Main([]string{"doctor", "--plugin", "--workflow-graph", "--repo", repo, "--project", t.TempDir(), "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected ambiguous mode failure, got %d", code)
	}
	assertCLIErrorCode(t, stderr.Bytes(), "doctor_mode_ambiguous")
}

func TestDoctorPluginModeFailsClosedOnCopiedBaseFallback(t *testing.T) {
	repo := t.TempDir()
	writeCLIPluginUpdateFixture(t, repo)
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeCLITestSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan")

	var stdout, stderr bytes.Buffer
	code := Main([]string{"doctor", "--plugin", "--repo", repo, "--profile", "demo", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected copied base fallback failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != false || payload["mode"] != "skill_plugin_overlay_doctor" {
		t.Fatalf("unexpected failure payload: %+v", payload)
	}
	reasons := map[string]bool{}
	for _, raw := range payload["reason_codes"].([]any) {
		reasons[raw.(string)] = true
	}
	for _, want := range []string{"legacy_copied_base_suite_present", "profile_skill_shadows_plugin_base", "missing_wrapper_evidence"} {
		if !reasons[want] {
			t.Fatalf("missing reason code %s in %+v", want, reasons)
		}
	}
	if !strings.Contains(payload["next_action"].(string), "Do not fall back") {
		t.Fatalf("next_action should forbid fallback: %+v", payload)
	}
}

func TestDoctorPluginModeRequiresExplicitRepo(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Main([]string{"doctor", "--plugin", "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing repo failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "repo_required_for_plugin_doctor")
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout on missing repo failure, got %q", stdout.String())
	}
}

func TestDoctorPluginModeRejectsProfileRootOverrideWithoutHarnessGuard(t *testing.T) {
	repo := t.TempDir()
	writeCLIPluginUpdateFixture(t, repo)
	var stdout, stderr bytes.Buffer
	code := Main([]string{"doctor", "--plugin", "--repo", repo, "--profile", "demo", "--profile-root", filepath.Join(t.TempDir(), "profile"), "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected profile-root guard failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	assertCLIErrorCode(t, stderr.Bytes(), "profile_root_override_rejected")
}

func TestDoctorPluginModeFailsClosedOnMissingRegisteredBaseFile(t *testing.T) {
	repo := t.TempDir()
	writeCLIPluginUpdateFixture(t, repo)
	if err := os.Remove(filepath.Join(repo, "skills", "kkachi-plan", "SKILL.md")); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"doctor", "--plugin", "--repo", repo, "--json"}, &stdout, &stderr, nil)
	if code != 2 {
		t.Fatalf("expected missing base file failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	reasons := map[string]bool{}
	for _, raw := range payload["reason_codes"].([]any) {
		reasons[raw.(string)] = true
	}
	if !reasons["missing_plugin_base_skill"] {
		t.Fatalf("missing reason code in %+v", payload)
	}
}

func TestDoctorPluginModeFailsClosedOnWrapperMismatchAndOverlayShadowing(t *testing.T) {
	repo := t.TempDir()
	writeCLIPluginUpdateFixture(t, repo)
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeCLISkillDoctorMismatchedWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"))
	writeCLISkillDoctorOverlay(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "kkachi-plan"), "kkachi-agent-skills:plan")

	var stdout, stderr bytes.Buffer
	code := Main([]string{"doctor", "--plugin", "--repo", repo, "--profile", "demo", "--profile-root", profileRoot, "--project", "doksuri", "--json"}, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected wrapper/overlay failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	reasons := map[string]bool{}
	for _, raw := range payload["reason_codes"].([]any) {
		reasons[raw.(string)] = true
	}
	for _, want := range []string{"wrapper_role_manifest_mismatch", "project_overlay_shadows_plugin_base"} {
		if !reasons[want] {
			t.Fatalf("missing reason code %s in %+v", want, payload)
		}
	}
}

func TestDoctorPluginModeFailsClosedOnProviderModelOverlayKeys(t *testing.T) {
	repo := t.TempDir()
	writeCLIPluginUpdateFixture(t, repo)
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeCLISkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"))
	writeCLISkillDoctorOverlayWithBody(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "provider-model-overlay"), "kkachi-agent-skills:plan", "provider: openai\nmodel: gpt-5\n")

	var stdout, stderr bytes.Buffer
	code := Main([]string{"doctor", "--plugin", "--repo", repo, "--profile", "demo", "--profile-root", profileRoot, "--project", "doksuri", "--json"}, &stdout, &stderr, map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"})
	if code != 2 {
		t.Fatalf("expected provider/model overlay failure, got code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	reasons := map[string]bool{}
	for _, raw := range payload["reason_codes"].([]any) {
		reasons[raw.(string)] = true
	}
	if !reasons["overlay_runtime_config_boundary_violation"] {
		t.Fatalf("missing runtime boundary reason in %+v", payload)
	}
}

func writeCLIPluginUpdateFixture(t *testing.T, repo string) {
	t.Helper()
	skills := []string{"kkachi-final-verify", "kkachi-plan", "kkachi-review", "kkachi-verify"}
	for _, skill := range skills {
		writeCLITestSkill(t, filepath.Join(repo, "skills", skill), skill)
	}
	for _, guide := range []string{"kas-project-overlay-guide", "kas-overlay-compose-guide", "kas-overlay-doctor-guide"} {
		writeCLITestSkill(t, filepath.Join(repo, "skills", guide), guide)
	}
	if err := os.MkdirAll(filepath.Join(repo, "roles"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "name: kkachi-agent-skills\nversion: 0.1.0\nplugin:\n  namespace: kkachi-agent-skills\n  package_manifest: skill-pack.yaml\n  load_policy: plugin_qualified_fail_closed\nroles:\n  blue: roles/blue.yaml\n  red: roles/red.yaml\n  orange: roles/orange.yaml\n  gray: roles/gray.yaml\nguides:\n  - kas-project-overlay-guide\n  - kas-overlay-compose-guide\n  - kas-overlay-doctor-guide\nskills:\n"
	for _, skill := range skills {
		content += "  - " + skill + "\n"
	}
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, role := range []struct {
		name   string
		skills []string
	}{
		{name: "blue", skills: skills},
		{name: "red", skills: []string{"kkachi-review", "kkachi-verify"}},
		{name: "orange", skills: []string{"kkachi-review"}},
		{name: "gray", skills: []string{"kkachi-final-verify", "kkachi-review"}},
	} {
		roleContent := "version: kas-plugin-role-manifest/v1\nrole: " + role.name + "\nskills:\n"
		for _, skill := range role.skills {
			roleContent += "  - " + skill + "\n"
		}
		if err := os.WriteFile(filepath.Join(repo, "roles", role.name+".yaml"), []byte(roleContent), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeCLISkillDoctorWrapper(t *testing.T, dir string) {
	t.Helper()
	content := `---
name: kkachi-blue-wrapper
metadata:
  kas:
    kind: color_wrapper
    role: blue_commander
    role_manifest: kkachi-agent-skills:roles/blue.yaml
    plugin_namespace: kkachi-agent-skills
    overlay_root: skills/<project>/kas-overlays
---
# Wrapper
`
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCLISkillDoctorMismatchedWrapper(t *testing.T, dir string) {
	t.Helper()
	content := `---
name: kkachi-blue-wrapper
metadata:
  kas:
    kind: color_wrapper
    role: red_reviewer
    role_manifest: kkachi-agent-skills:roles/blue.yaml
    plugin_namespace: kkachi-agent-skills
    overlay_root: skills/<project>/kas-overlays
---
# Wrapper
`
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeCLISkillDoctorOverlay(t *testing.T, dir string, overlayFor string) {
	t.Helper()
	writeCLISkillDoctorOverlayWithBody(t, dir, overlayFor, "")
}

func writeCLISkillDoctorOverlayWithBody(t *testing.T, dir string, overlayFor string, body string) {
	t.Helper()
	content := `---
name: doksuri-blue-plan-overlay
metadata:
  kas:
    kind: project_overlay
    project: doksuri
    role: blue_commander
    overlay_for: ` + overlayFor + `
    merge_mode: additive_constraints
    base_version: "0.1.0"
---
# Overlay
` + body
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func countRegularFiles(t *testing.T, root string) int {
	t.Helper()
	count := 0
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestUninstallDryRunPlansManifestTrackedOnly(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	writeCLITestSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan")
	writeCLITestSkillPackYAML(t, repo, "kkachi-plan")
	profileRoot := filepath.Join(dir, "profile")
	env := map[string]string{"KAS_ALLOW_PROFILE_ROOT_OVERRIDE": "1"}

	var stdout, stderr bytes.Buffer
	code := Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--dry-run", "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
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
	code = Main([]string{"install-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--source-pack", projectinstall.VirtualSourcePackID, "--suite-role", "blue_commander", "--approve", evidence, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("approved install failed: code=%d stderr=%s", code, stderr.String())
	}
	localOnly := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-local", "SKILL.md")
	writeCLITestSkill(t, filepath.Dir(localOnly), "doksuri-server-local")
	target := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-wrapper", "SKILL.md")
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
	if len(removals) != 3 {
		t.Fatalf("unexpected planned removals: %+v", removals)
	}
	seenRemovals := map[string]bool{}
	for _, raw := range removals {
		removal := raw.(map[string]any)
		if removal["action"] != "remove" {
			t.Fatalf("unexpected planned removal action: %+v", removal)
		}
		seenRemovals[removal["target_path"].(string)] = true
	}
	for _, want := range []string{
		"skills/doksuri-server/doksuri-server-wrapper/SKILL.md",
		"skills/doksuri-server/doksuri-server-overlay/SKILL.md",
		"skills/doksuri-server/doksuri-server-overlay/references/legacy-delta-extract.md",
	} {
		if !seenRemovals[want] {
			t.Fatalf("missing planned removal %s in %+v", want, removals)
		}
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
	if _, err := os.Stat(filepath.Join(applied["backup_path"].(string), "files", "skills", "doksuri-server", "doksuri-server-wrapper", "SKILL.md")); err != nil {
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
  repo: "kkachi-agent-skills"
  remote: "github.com/SeventeenthEarth/kkachi-agent-skills"
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

func TestNormalizeInstallArgsTreatsSuiteRoleAsValueFlag(t *testing.T) {
	args := []string{"alpha", "--suite-role", "blue_commander", "--dry-run", "--profile=kwanwoo", "--suite-role=red_reviewer"}
	got := normalizeInstallArgs(args)
	want := []string{"--suite-role", "blue_commander", "--dry-run", "--profile=kwanwoo", "--suite-role=red_reviewer", "alpha"}
	if len(got) != len(want) {
		t.Fatalf("normalizeInstallArgs() length = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalizeInstallArgs()[%d] = %q, want %q; got %v", i, got[i], want[i], got)
		}
	}
	if !hasInstallFlagValue("--suite-role=gray_scribe") {
		t.Fatal("hasInstallFlagValue should recognize --suite-role=<value>")
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
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	code = Main([]string{"repair-project-kas", "--repo", repo, "--profile", "kwanwoo", "--project", "doksuri-server", "--approve", evidence, "--backup-vault-root", vault, "--profile-root", profileRoot, "--json"}, &stdout, &stderr, env)
	if code != 0 {
		t.Fatalf("approved repair failed: code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-wrapper", "SKILL.md")); err != nil {
		t.Fatalf("approved repair did not write project wrapper under temp profile: %v", err)
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
		t.Fatalf("expected removed migrate-project-kas command to fail closed, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown_command") {
		t.Fatalf("expected unknown_command for removed migrate-project-kas, got %s", stderr.String())
	}

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
