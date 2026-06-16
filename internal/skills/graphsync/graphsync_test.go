package graphsync

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/doctor"
)

var errFakeExit = errors.New("fake exit")

type fakeRunner struct {
	project string
	results map[string]doctor.CommandResult
	calls   []string
	apply   func()
}

func (runner *fakeRunner) Run(workDir string, args ...string) doctor.CommandResult {
	key := strings.Join(args, " ")
	runner.calls = append(runner.calls, key)
	if key == "graph apply --proposal prop-1 --approval approved:1 --json" && runner.apply != nil {
		runner.apply()
	}
	if result, ok := runner.results[key]; ok {
		return result
	}
	return doctor.CommandResult{Err: errFakeExit}
}

func graphsyncRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func writeGraph(t *testing.T, project string, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(project, ".kkachi-workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readGraph(t *testing.T, project string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(project, ".kkachi-workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func newFakeRunner(project string) *fakeRunner {
	return &fakeRunner{
		project: project,
		results: map[string]doctor.CommandResult{
			"--version":           {Stdout: []byte("kkachi-agent-helper 0.1.9\n")},
			"capabilities --json": {Stdout: []byte(`{"commands":["graph validate","graph explain","graph diff","graph propose","graph apply"],"flags":["workflow_graph_readonly","workflow_graph_diagnostics","workflow_graph_no_direct_yaml_fallback","workflow_graph_configurable_feedback_intake","workflow_graph_apply"]}`)},
			"graph --help":        {Stdout: []byte("Usage: graph\n  validate\n  explain\n  diff\n  propose\n  apply\n")},
			"graph validate --file .kkachi-workflow.yaml --json":                                                                        {Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.0.9","checksum":"sha256:old"}`)},
			"graph explain --file .kkachi-workflow.yaml --json":                                                                         {Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.0.9","checksum":"sha256:old"}`)},
			"graph diff --from .kkachi-workflow.yaml --to .kkachi/graph/candidates/kas-default-96af9b5b030fc4ca.yaml --semantic --json": {Stdout: []byte(`{"ok":true,"summary":"semantic diff ready","risk_flags":["phase_path_change"],"reason_codes":["graph_stale"]}`)},
			"graph propose --candidate-file .kkachi/graph/candidates/kas-default-96af9b5b030fc4ca.yaml --reason repair --json":          {Stdout: []byte(`{"ok":true,"proposal_id":"prop-1","proposal_path":".kkachi/graph/proposals/prop-1.yaml","approval_required":true,"risk_flags":["phase_path_change"],"reason_codes":["proposal_recorded"]}`)},
			"graph apply --proposal prop-1 --approval approved:1 --json":                                                                {Stdout: []byte(`{"ok":true,"proposal_id":"prop-1","approval_ref":"approved:1","audit_event_ids":["evt-1"],"audit_path":".kkachi/events/evt-1.json","backup_path":".kkachi/graph/backups/old.yaml","recovery_path":".kkachi/graph/recovery/prop-1.md"}`)},
		},
	}
}

func TestProposeRefusesUnsafeDoctorStates(t *testing.T) {
	tests := []struct {
		name   string
		graph  string
		mutate func(*fakeRunner)
	}{
		{
			name:  "old-kah",
			graph: "version: workflow-graph/v1\n",
			mutate: func(r *fakeRunner) {
				r.results["--version"] = doctor.CommandResult{Stdout: []byte("kkachi-agent-helper 0.1.8\n")}
			},
		},
		{
			name:  "custom-supported",
			graph: "version: workflow-graph/v1\n",
			mutate: func(r *fakeRunner) {
				r.results["graph validate --file .kkachi-workflow.yaml --json"] = doctor.CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"team-custom","custom":true}`)}
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = doctor.CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"team-custom","custom":true}`)}
			},
		},
		{
			name:  "pass",
			graph: "version: workflow-graph/v1\n",
			mutate: func(r *fakeRunner) {
				r.results["graph validate --file .kkachi-workflow.yaml --json"] = doctor.CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0"}`)}
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = doctor.CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0"}`)}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := t.TempDir()
			writeGraph(t, project, tt.graph)
			runner := newFakeRunner(project)
			tt.mutate(runner)
			before := readGraph(t, project)
			result, err := Propose(Options{Repo: graphsyncRepoRoot(t), Project: project, Reason: "repair", Runner: runner.Run})
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Status != "blocked_for_approval" {
				t.Fatalf("expected refused proposal, got %+v", result)
			}
			if after := readGraph(t, project); after != before {
				t.Fatalf("KAS changed graph directly in refused proposal")
			}
			for _, call := range runner.calls {
				if strings.Contains(call, " diff ") || strings.Contains(call, " propose ") || strings.Contains(call, " apply ") {
					t.Fatalf("refused proposal called mutating repair command %q", call)
				}
			}
		})
	}
}

func TestProposalAllowedRefusesUnsafeStatuses(t *testing.T) {
	for _, status := range []string{
		"pass",
		"custom_supported",
		"update_kah_required",
		"update_kah_recommended",
		"update_kas_recommended",
		"graph_conflict",
		"unsupported",
	} {
		t.Run(status, func(t *testing.T) {
			if proposalAllowed(status) {
				t.Fatalf("proposalAllowed(%q)=true, want false", status)
			}
		})
	}
}

func TestProposeMissingStaleBrokenStates(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(string, *fakeRunner)
		wantDiffState string
	}{
		{
			name: "missing",
			setup: func(project string, r *fakeRunner) {
				delete(r.results, "graph validate --file .kkachi-workflow.yaml --json")
				delete(r.results, "graph explain --file .kkachi-workflow.yaml --json")
			},
			wantDiffState: "not_applicable_missing_or_invalid_base",
		},
		{
			name: "stale",
			setup: func(project string, r *fakeRunner) {
				writeGraph(t, project, "version: workflow-graph/v1\n")
			},
			wantDiffState: "completed",
		},
		{
			name: "broken",
			setup: func(project string, r *fakeRunner) {
				writeGraph(t, project, "broken: true\n")
				r.results["graph validate --file .kkachi-workflow.yaml --json"] = doctor.CommandResult{Stdout: []byte(`{"ok":false}`), Err: errFakeExit}
				r.results["graph explain --file .kkachi-workflow.yaml --json"] = doctor.CommandResult{Stdout: []byte(`{"ok":false}`), Err: errFakeExit}
			},
			wantDiffState: "not_applicable_missing_or_invalid_base",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := t.TempDir()
			runner := newFakeRunner(project)
			tt.setup(project, runner)
			before := ""
			if _, err := os.Stat(filepath.Join(project, ".kkachi-workflow.yaml")); err == nil {
				before = readGraph(t, project)
			}
			result, err := Propose(Options{Repo: graphsyncRepoRoot(t), Project: project, Reason: "repair", Runner: runner.Run})
			if err != nil {
				t.Fatal(err)
			}
			if !result.OK || result.Status != "proposal_available" || result.Proposal.ID != "prop-1" {
				t.Fatalf("unexpected proposal result: %+v", result)
			}
			if result.SemanticDiff.State != tt.wantDiffState {
				t.Fatalf("diff state=%s want %s", result.SemanticDiff.State, tt.wantDiffState)
			}
			if tt.wantDiffState == "not_applicable_missing_or_invalid_base" {
				for _, call := range runner.calls {
					if strings.Contains(call, "graph diff") {
						t.Fatalf("invalid/missing base should not run diff, got %q", call)
					}
				}
			}
			if before != "" && readGraph(t, project) != before {
				t.Fatalf("proposal changed target graph directly")
			}
			if result.Candidate.Path == "" || result.Candidate.Checksum == "" {
				t.Fatalf("missing candidate evidence: %+v", result)
			}
		})
	}
}

func TestProposeCommandOrderForStaleGraph(t *testing.T) {
	project := t.TempDir()
	writeGraph(t, project, "version: workflow-graph/v1\n")
	runner := newFakeRunner(project)
	result, err := Propose(Options{Repo: graphsyncRepoRoot(t), Project: project, Reason: "repair", Runner: runner.Run})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("unexpected result: %+v", result)
	}
	want := []string{
		"--version",
		"capabilities --json",
		"graph --help",
		"graph validate --file .kkachi-workflow.yaml --json",
		"graph explain --file .kkachi-workflow.yaml --json",
		"capabilities --json",
		"graph --help",
		"graph diff --from .kkachi-workflow.yaml --to .kkachi/graph/candidates/kas-default-96af9b5b030fc4ca.yaml --semantic --json",
		"graph propose --candidate-file .kkachi/graph/candidates/kas-default-96af9b5b030fc4ca.yaml --reason repair --json",
	}
	if strings.Join(runner.calls, "\n") != strings.Join(want, "\n") {
		t.Fatalf("calls mismatch\ngot:\n%s\nwant:\n%s", strings.Join(runner.calls, "\n"), strings.Join(want, "\n"))
	}
}

func TestApplyRequiresApprovalAndAppliesThroughKAH(t *testing.T) {
	project := t.TempDir()
	writeGraph(t, project, "version: workflow-graph/v1\n")
	runner := newFakeRunner(project)
	missingApproval, err := Apply(Options{Repo: graphsyncRepoRoot(t), Project: project, Proposal: "prop-1", Runner: runner.Run})
	if err != nil {
		t.Fatal(err)
	}
	if missingApproval.OK || missingApproval.Status != "blocked_for_approval" {
		t.Fatalf("expected approval block, got %+v", missingApproval)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("missing approval should not call KAH, got calls: %s", strings.Join(runner.calls, "\n"))
	}

	runner = newFakeRunner(project)
	runner.apply = func() {
		writeGraph(t, project, "version: workflow-graph/v1\nmetadata:\n  source_template: kas-default\n")
	}
	runner.results["graph validate --file .kkachi-workflow.yaml --json"] = doctor.CommandResult{Stdout: []byte(`{"ok":true,"schema_version":"workflow-graph/v1","source_template":"kas-default","template_version":"0.1.0","checksum":"sha256:new"}`)}
	runner.results["graph explain --file .kkachi-workflow.yaml --json"] = doctor.CommandResult{Stdout: []byte(`{"ok":true,"graph":{"graph_version":"workflow-graph/v1","checksum":"sha256:new"}}`)}
	result, err := Apply(Options{Repo: graphsyncRepoRoot(t), Project: project, Proposal: "prop-1", Approval: "approved:1", Runner: runner.Run})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Status != "applied" || result.Apply.ProposalID != "prop-1" || result.PostApply.GraphChecksum != "sha256:new" {
		t.Fatalf("unexpected apply result: %+v", result)
	}
	wantTail := []string{
		"graph apply --proposal prop-1 --approval approved:1 --json",
		"graph validate --file .kkachi-workflow.yaml --json",
		"graph explain --file .kkachi-workflow.yaml --json",
	}
	gotTail := runner.calls[len(runner.calls)-3:]
	if strings.Join(gotTail, "\n") != strings.Join(wantTail, "\n") {
		t.Fatalf("apply tail mismatch\ngot:\n%s\nwant:\n%s", strings.Join(gotTail, "\n"), strings.Join(wantTail, "\n"))
	}
}
