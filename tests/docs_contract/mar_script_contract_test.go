package docscontract

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var marStatusVocabulary = []string{
	"PASS",
	"PASS_WITH_FINDINGS",
	"REQUEST_CHANGES",
	"BLOCKED",
	"DEGRADED",
	"FAILED",
}

func TestMAR003ScriptHelpExposesLocalMVPSurfaces(t *testing.T) {
	root := repoRoot(t)
	cmd := exec.Command("python3", "scripts/mar.py", "--help")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python3 scripts/mar.py --help failed: %v\n%s", err, output)
	}

	help := string(output)
	for _, want := range []string{"doctor", "render", "validate", "merge-pack"} {
		if !strings.Contains(help, want) {
			t.Fatalf("scripts/mar.py --help missing %q\n%s", want, help)
		}
	}
}

func TestMAR003ScriptReadbackEnforcesStatusVocabulary(t *testing.T) {
	source := readRepoFile(t, "scripts/mar.py")
	for _, status := range marStatusVocabulary {
		if !strings.Contains(source, status) {
			t.Fatalf("scripts/mar.py missing status vocabulary value %q", status)
		}
	}

	for _, forbidden := range []string{
		"PASS_WITH_WARNINGS",
		"NEEDS_WORK",
		"ERROR",
		"OK",
		"SUCCESS",
		"WARNING",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("scripts/mar.py contains non-contract status vocabulary value %q", forbidden)
		}
	}
}

func TestMAR003FixturesCoverLocalMockPaths(t *testing.T) {
	jsonFixtures := map[string]string{
		"pass-review.json":            "PASS",
		"request-changes-review.json": "REQUEST_CHANGES",
		"provider-unavailable.json":   "DEGRADED",
		"all-failed.json":             "FAILED",
		"insufficient-coverage.json":  "BLOCKED",
		"mutation-detected.json":      "REQUEST_CHANGES",
	}

	for name, wantStatus := range jsonFixtures {
		var fixture struct {
			Case   string `json:"case"`
			Status string `json:"status"`
			Mock   bool   `json:"mock_provider_path"`
		}
		data := readMARFixture(t, name)
		if err := json.Unmarshal(data, &fixture); err != nil {
			t.Fatalf("parse fixture %s: %v", name, err)
		}
		if fixture.Case == "" {
			t.Fatalf("fixture %s missing case", name)
		}
		if fixture.Status != wantStatus {
			t.Fatalf("fixture %s status = %q, want %q", name, fixture.Status, wantStatus)
		}
		if !fixture.Mock {
			t.Fatalf("fixture %s must mark mock_provider_path=true", name)
		}
	}

	for _, name := range []string{"parse-failure-raw.txt", "raw-output-over-cap.txt"} {
		data := strings.TrimSpace(string(readMARFixture(t, name)))
		if data == "" {
			t.Fatalf("fixture %s is empty", name)
		}
	}
}

func TestMAR003ValidatePreservesParseFailureAsDegradedEvidence(t *testing.T) {
	output := runMARCommand(t,
		"validate",
		"--input", "tests/fixtures/mar/parse-failure-raw.txt",
		"--raw-cap", "32",
	)
	var result struct {
		Status              string `json:"status"`
		Reason              string `json:"reason"`
		NoProviderExecution bool   `json:"no_provider_execution"`
		ParseFailure        struct {
			Message string `json:"message"`
			Line    int    `json:"line"`
			Column  int    `json:"column"`
		} `json:"parse_failure"`
		RawOutput struct {
			BytesReturned int  `json:"bytes_returned"`
			CapBytes      int  `json:"cap_bytes"`
			Truncated     bool `json:"truncated"`
		} `json:"raw_output"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse validate output: %v\n%s", err, output)
	}
	if result.Status != "DEGRADED" || result.Reason != "parse_failure" {
		t.Fatalf("parse failure status/reason = %q/%q, want DEGRADED/parse_failure\n%s", result.Status, result.Reason, output)
	}
	if !result.NoProviderExecution {
		t.Fatalf("validate output must record no_provider_execution=true\n%s", output)
	}
	if result.ParseFailure.Message == "" || result.ParseFailure.Line == 0 || result.ParseFailure.Column == 0 {
		t.Fatalf("validate output missing parse failure details\n%s", output)
	}
	if !result.RawOutput.Truncated || result.RawOutput.CapBytes != 32 || result.RawOutput.BytesReturned > 32 {
		t.Fatalf("validate output missing deterministic truncation metadata\n%s", output)
	}
}

func TestMAR003DoctorAcceptsFixtureEvidenceWithoutProviderExecution(t *testing.T) {
	output := runMARCommand(t,
		"doctor",
		"--fixture", "tests/fixtures/mar/provider-unavailable.json",
	)
	var result struct {
		Status              string `json:"status"`
		CapabilityStatus    string `json:"capability_status"`
		NoProviderExecution bool   `json:"no_provider_execution"`
		FixtureEvidence     struct {
			Status              string `json:"status"`
			Reason              string `json:"reason"`
			NoProviderExecution bool   `json:"no_provider_execution"`
			Path                string `json:"path"`
		} `json:"fixture_evidence"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse doctor output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.CapabilityStatus != "PASS" {
		t.Fatalf("doctor capability status = %q/%q, want PASS/PASS\n%s", result.Status, result.CapabilityStatus, output)
	}
	if !result.NoProviderExecution {
		t.Fatalf("doctor output must record no_provider_execution=true\n%s", output)
	}
	if result.FixtureEvidence.Status != "DEGRADED" {
		t.Fatalf("doctor fixture status = %q, want DEGRADED\n%s", result.FixtureEvidence.Status, output)
	}
	if !result.FixtureEvidence.NoProviderExecution {
		t.Fatalf("doctor fixture evidence must record no_provider_execution=true\n%s", output)
	}
	if result.FixtureEvidence.Path != "tests/fixtures/mar/provider-unavailable.json" {
		t.Fatalf("doctor fixture path = %q, want fixture path\n%s", result.FixtureEvidence.Path, output)
	}
}

func TestMAR003MergePackAggregatesFailClosedLocalStatuses(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "request changes beats pass",
			args: []string{
				"merge-pack",
				"tests/fixtures/mar/pass-review.json",
				"tests/fixtures/mar/request-changes-review.json",
			},
			want: "REQUEST_CHANGES",
		},
		{
			name: "all failed remains failed",
			args: []string{
				"merge-pack",
				"tests/fixtures/mar/all-failed.json",
			},
			want: "FAILED",
		},
		{
			name: "insufficient coverage blocks",
			args: []string{
				"merge-pack",
				"--required-reviewers", "3",
				"tests/fixtures/mar/pass-review.json",
			},
			want: "BLOCKED",
		},
		{
			name: "provider unavailable degrades",
			args: []string{
				"merge-pack",
				"tests/fixtures/mar/provider-unavailable.json",
			},
			want: "DEGRADED",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runMARCommand(t, tc.args...)
			var result struct {
				Status              string `json:"status"`
				NoProviderExecution bool   `json:"no_provider_execution"`
			}
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("parse merge-pack output: %v\n%s", err, output)
			}
			if result.Status != tc.want {
				t.Fatalf("merge-pack status = %q, want %q\n%s", result.Status, tc.want, output)
			}
			if !result.NoProviderExecution {
				t.Fatalf("merge-pack output must record no_provider_execution=true\n%s", output)
			}
		})
	}
}

func TestMAR003RenderUsesLocalTemplateOnly(t *testing.T) {
	output := runMARCommand(t,
		"render",
		"--template", "templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl",
		"--set", "RunID=run-local",
		"--set", "TaskID=MAR-003",
		"--set", "TaskContract=contract text",
		"--set", "DiffSummary=diff text",
	)
	var result struct {
		Status              string `json:"status"`
		NoProviderExecution bool   `json:"no_provider_execution"`
		RenderedPrompt      string `json:"rendered_prompt"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse render output: %v\n%s", err, output)
	}
	if result.Status != "PASS" {
		t.Fatalf("render status = %q, want PASS\n%s", result.Status, output)
	}
	for _, want := range []string{"run-local", "MAR-003", "contract text", "diff text"} {
		if !strings.Contains(result.RenderedPrompt, want) {
			t.Fatalf("rendered prompt missing %q\n%s", want, result.RenderedPrompt)
		}
	}
	if !result.NoProviderExecution {
		t.Fatalf("render output must record no_provider_execution=true\n%s", output)
	}
}

func TestMAR003ScriptDoesNotClaimOutOfScopeExecution(t *testing.T) {
	surfaces := map[string]string{
		"scripts/mar.py":                            readRepoFile(t, "scripts/mar.py"),
		"docs/sot/multi-agent-review-policy.md":     readRepoFile(t, "docs/sot/multi-agent-review-policy.md"),
		"skills/kkachi-multi-agent-review/SKILL.md": readRepoFile(t, "skills/kkachi-multi-agent-review/SKILL.md"),
	}
	for rel, content := range surfaces {
		for _, forbidden := range []string{
			"MAR-003 implements live provider execution",
			"MAR-003 activates KAB",
			"MAR-003 implements KAH MAR gate behavior",
			"MAR-003 completion requires live provider execution",
			"MAR-003 completion requires KAB activation",
			"MAR-003 completion requires KAH gate behavior",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s contains out-of-scope MAR-003 claim %q", rel, forbidden)
			}
		}
	}
}

func runMARCommand(t *testing.T, args ...string) []byte {
	t.Helper()
	cmdArgs := append([]string{"scripts/mar.py"}, args...)
	cmd := exec.Command("python3", cmdArgs...)
	cmd.Dir = repoRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python3 %s failed: %v\n%s", strings.Join(cmdArgs, " "), err, output)
	}
	return output
}

func readMARFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "tests", "fixtures", "mar", name))
	if err != nil {
		t.Fatalf("read MAR fixture %s: %v", name, err)
	}
	return data
}
