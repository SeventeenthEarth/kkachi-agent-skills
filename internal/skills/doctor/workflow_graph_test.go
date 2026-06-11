package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type workflowGraphFakeRunner struct {
	results map[string]CommandResult
	calls   []string
}

func (runner *workflowGraphFakeRunner) Run(workDir string, args ...string) CommandResult {
	key := strings.Join(args, " ")
	runner.calls = append(runner.calls, key)
	if result, ok := runner.results[key]; ok {
		return result
	}
	return CommandResult{Err: ErrExit}
}

func doctorRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func writeWorkflowGraphFixture(t *testing.T, project string, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(project, ".kkachi-workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func baseWorkflowGraphRunner() *workflowGraphFakeRunner {
	return &workflowGraphFakeRunner{results: map[string]CommandResult{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.9\n")},
		"capabilities --json": {Stdout: []byte(`{"commands":["graph validate","graph explain"],"flags":["workflow_graph_readonly","workflow_graph_diagnostics","workflow_graph_no_direct_yaml_fallback","workflow_graph_configurable_feedback_intake"]}`)},
		"graph --help":        {Stdout: []byte("Usage: graph\n  validate\n  explain\n")},
		"graph validate --file .kkachi-workflow.yaml --json": {Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0","checksum":"sha256:ok"}`)},
		"graph explain --file .kkachi-workflow.yaml --json":  {Stdout: []byte(`{"ok":true,"graph":{"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0","checksum":"sha256:ok"}}`)},
	}}
}

func TestBuildWorkflowGraphClassifications(t *testing.T) {
	repo := doctorRepoRoot(t)
	tests := []struct {
		name        string
		graph       string
		mutate      func(*workflowGraphFakeRunner)
		wantStatus  string
		wantOK      bool
		versionGate bool
	}{
		{
			name:       "pass",
			graph:      "version: workflow-graph/v1\nsource_template: kas-default\n",
			wantStatus: "pass",
			wantOK:     true,
		},
		{
			name:  "old-kah",
			graph: "version: workflow-graph/v1\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["--version"] = CommandResult{Stdout: []byte("kkachi-agent-helper 0.1.8\n")}
			},
			wantStatus:  "update_kah_required",
			versionGate: true,
		},
		{
			name:  "malformed-kah-version",
			graph: "version: workflow-graph/v1\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["--version"] = CommandResult{Stdout: []byte("kkachi-agent-helper dev build\n")}
			},
			wantStatus:  "update_kah_required",
			versionGate: true,
		},
		{
			name:  "stale",
			graph: "version: workflow-graph/v1\nsource_template: kas-default\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.0.9"}`)}
			},
			wantStatus: "graph_stale",
		},
		{
			name:  "broken",
			graph: "broken: true\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["graph validate --file .kkachi-workflow.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":false,"diagnostics":[{"code":"parse"}]}`), Err: ErrExit}
			},
			wantStatus: "graph_broken",
		},
		{
			name:  "custom-supported",
			graph: "version: workflow-graph/v1\nsource_template: custom\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"team-custom","custom":true}`)}
			},
			wantStatus: "custom_supported",
			wantOK:     true,
		},
		{
			name:  "unsupported-schema",
			graph: "version: ???\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"not-workflow"}`)}
			},
			wantStatus: "unsupported",
		},
		{
			name:  "newer-schema",
			graph: "version: workflow-graph/v2\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v2"}`)}
			},
			wantStatus: "update_kas_recommended",
			wantOK:     true,
		},
		{
			name:  "newer-template-version",
			graph: "version: workflow-graph/v1\nsource_template: kas-default\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.2.0"}`)}
			},
			wantStatus: "update_kas_recommended",
			wantOK:     true,
		},
		{
			name:  "conflict",
			graph: "version: workflow-graph/v1\n",
			mutate: func(r *workflowGraphFakeRunner) {
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = CommandResult{Err: ErrExit}
			},
			wantStatus: "graph_conflict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := t.TempDir()
			writeWorkflowGraphFixture(t, project, tt.graph)
			runner := baseWorkflowGraphRunner()
			if tt.mutate != nil {
				tt.mutate(runner)
			}
			before, err := os.ReadFile(filepath.Join(project, ".kkachi-workflow.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			result, err := BuildWorkflowGraph(repo, WorkflowGraphOptions{Project: project, Runner: runner.Run})
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != tt.wantStatus || result.OK != tt.wantOK || !result.NoWrite {
				t.Fatalf("unexpected result: %+v", result)
			}
			after, err := os.ReadFile(filepath.Join(project, ".kkachi-workflow.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatalf("workflow graph doctor modified graph file")
			}
			for _, call := range runner.calls {
				if strings.Contains(call, " init") || strings.Contains(call, " diff") || strings.Contains(call, " propose") || strings.Contains(call, " apply") || strings.Contains(call, " export") {
					t.Fatalf("workflow graph doctor called mutating/out-of-scope command %q", call)
				}
				if tt.versionGate && call != "--version" {
					t.Fatalf("version-gated KAH should stop before capabilities/help/graph commands, got %q", call)
				}
			}
		})
	}
}

func TestBuildWorkflowGraphMissingGraphDoesNotCreateFiles(t *testing.T) {
	project := t.TempDir()
	runner := baseWorkflowGraphRunner()
	result, err := BuildWorkflowGraph(doctorRepoRoot(t), WorkflowGraphOptions{Project: project, Runner: runner.Run})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "graph_missing" || result.Project.GraphPresent {
		t.Fatalf("unexpected missing graph result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi-workflow.yaml")); !os.IsNotExist(err) {
		t.Fatalf("doctor created graph file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".kkachi")); !os.IsNotExist(err) {
		t.Fatalf("doctor created .kkachi directory: %v", err)
	}
	for _, call := range runner.calls {
		if strings.Contains(call, "validate") || strings.Contains(call, "explain") || strings.Contains(call, "init") {
			t.Fatalf("missing graph should not run graph validate/explain/init, got %q", call)
		}
	}
}
