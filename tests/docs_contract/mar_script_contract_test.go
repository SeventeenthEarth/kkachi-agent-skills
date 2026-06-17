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
	for _, want := range []string{"doctor", "render", "validate", "merge-pack", "provider-lanes", "provider-preflight", "provider-attempt"} {
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

func TestMAR004ProviderLaneRegistryReadbackIsFailClosed(t *testing.T) {
	output := runMARCommand(t, "provider-lanes")
	var result struct {
		Status           string   `json:"status"`
		SchemaVersion    string   `json:"schema_version"`
		DefaultReviewers []string `json:"default_reviewers"`
		Reviewers        map[string]struct {
			CommandLane                     string   `json:"command_lane"`
			SelectedModel                   *string  `json:"selected_model"`
			ValidationRequiredBeforeSuccess bool     `json:"validation_required_before_success_coverage"`
			PromptTemplate                  string   `json:"prompt_template"`
			CommandArgs                     []string `json:"command_args"`
		} `json:"reviewers"`
		ProviderFailureReasons []string `json:"provider_failure_reasons"`
		NoProviderExecution    bool     `json:"no_provider_execution"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse provider-lanes output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.SchemaVersion != "mar.provider_lanes.v1" {
		t.Fatalf("provider-lanes status/schema = %q/%q\n%s", result.Status, result.SchemaVersion, output)
	}
	wantDefault := []string{"zcode_glm_5_2", "kimi_k2_7", "antigravity_gemini"}
	if strings.Join(result.DefaultReviewers, ",") != strings.Join(wantDefault, ",") {
		t.Fatalf("default reviewers = %v, want %v", result.DefaultReviewers, wantDefault)
	}
	for _, reviewer := range wantDefault {
		lane, ok := result.Reviewers[reviewer]
		if !ok {
			t.Fatalf("provider-lanes missing reviewer %q\n%s", reviewer, output)
		}
		if lane.CommandLane == "" || lane.PromptTemplate == "" || !lane.ValidationRequiredBeforeSuccess {
			t.Fatalf("provider-lanes has incomplete fail-closed lane for %q: %+v", reviewer, lane)
		}
	}
	if model := result.Reviewers["kimi_k2_7"].SelectedModel; model == nil || *model != "k2.7" {
		t.Fatalf("kimi_k2_7 selected_model = %v, want k2.7", model)
	}
	if model := result.Reviewers["antigravity_gemini"].SelectedModel; model != nil {
		t.Fatalf("antigravity_gemini selected_model = %v, want nil until explicitly selected", *model)
	}
	agyArgs := strings.Join(result.Reviewers["antigravity_gemini"].CommandArgs, "\x00")
	if strings.Contains(agyArgs, "read-only") {
		t.Fatalf("antigravity_gemini command_args must not pass a value to boolean --sandbox flag: %v", result.Reviewers["antigravity_gemini"].CommandArgs)
	}
	for _, reason := range []string{"cli_missing", "model_unavailable", "mutation_detected", "unknown_provider_failure"} {
		if !containsString(result.ProviderFailureReasons, reason) {
			t.Fatalf("provider-lanes missing failure reason %q in %v", reason, result.ProviderFailureReasons)
		}
	}
	if !result.NoProviderExecution {
		t.Fatalf("provider-lanes must not execute providers\n%s", output)
	}
}

func TestMAR004ProviderToolchainOverlayResolvesExplicitArgv(t *testing.T) {
	registry := writeTempProviderRegistry(t, map[string]any{
		"schema_version":    "mar.provider_lanes.v1",
		"default_reviewers": []string{"alias_backed"},
		"reviewers": map[string]any{
			"alias_backed": map[string]any{
				"command_lane":    "zcode",
				"executable":      "definitely-missing-shell-alias",
				"selected_model":  "glm-5.2",
				"prompt_template": "templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl",
				"command_args":    []string{"--version"},
				"validated":       false,
				"validation_required_before_success_coverage": true,
			},
		},
	})
	toolchain := writeTempToolchain(t, `kas_cli: v0.1.4
kah_cli: v0.1.10
mar_provider_tools:
  schema_version: mar.provider_tools.v1
  providers:
    alias_backed:
      command_lane: zcode
      resolved_argv:
        - python3
      selected_model: glm-5.2
      version: 3.x
      validated: true
      validation_evidence: .kkachi/runs/run-test/mar/provider-proof-alias_backed.json
`)

	output := runMARCommand(t, "provider-preflight", "--registry", registry, "--toolchain", toolchain)
	var result struct {
		Status   string `json:"status"`
		Reason   string `json:"reason"`
		Attempts []struct {
			ReviewerID            string   `json:"reviewer_id"`
			TerminalStatus        string   `json:"terminal_status"`
			ProviderFailureReason *string  `json:"provider_failure_reason"`
			RedactedCommand       []string `json:"redacted_command"`
			NoProviderExecution   bool     `json:"no_provider_execution"`
		} `json:"attempts"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse provider-preflight output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.Reason != "provider_preflight_passed" {
		t.Fatalf("toolchain-resolved provider preflight = %q/%q, want PASS/provider_preflight_passed\n%s", result.Status, result.Reason, output)
	}
	if len(result.Attempts) != 1 || result.Attempts[0].ReviewerID != "alias_backed" {
		t.Fatalf("unexpected preflight attempts: %+v\n%s", result.Attempts, output)
	}
	attempt := result.Attempts[0]
	if attempt.TerminalStatus != "PASS" || attempt.ProviderFailureReason != nil || !attempt.NoProviderExecution {
		t.Fatalf("toolchain-resolved preflight attempt = %+v\n%s", attempt, output)
	}
	if strings.Join(attempt.RedactedCommand, " ") != "python3 --version" {
		t.Fatalf("toolchain resolved_argv must replace missing registry executable, got %v\n%s", attempt.RedactedCommand, output)
	}
}

func TestMAR004ProviderValidationGateBlocksUnvalidatedExecutableLanes(t *testing.T) {
	registry := writeTempProviderRegistry(t, map[string]any{
		"schema_version":    "mar.provider_lanes.v1",
		"default_reviewers": []string{"unvalidated_available"},
		"reviewers": map[string]any{
			"unvalidated_available": map[string]any{
				"command_lane":    "python3",
				"executable":      "python3",
				"selected_model":  "fixture-model",
				"prompt_template": "templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl",
				"command_args":    []string{"--version"},
				"validated":       false,
				"validation_required_before_success_coverage": true,
			},
		},
	})

	preflightOutput := runMARCommand(t, "provider-preflight", "--registry", registry)
	var preflight struct {
		Status   string `json:"status"`
		Attempts []struct {
			ReviewerID            string  `json:"reviewer_id"`
			TerminalStatus        string  `json:"terminal_status"`
			ProviderFailureReason string  `json:"provider_failure_reason"`
			SelectedModel         *string `json:"selected_model"`
			ExitCode              *int    `json:"exit_code"`
		} `json:"attempts"`
	}
	if err := json.Unmarshal(preflightOutput, &preflight); err != nil {
		t.Fatalf("parse provider-preflight output: %v\n%s", err, preflightOutput)
	}
	if preflight.Status == "PASS" {
		t.Fatalf("unvalidated executable provider lane must not preflight PASS\n%s", preflightOutput)
	}
	requireAttemptReason(t, preflight.Attempts, "unvalidated_available", "adapter_proof_required")

	attemptOutput := runMARCommand(t, "provider-attempt", "--registry", registry, "--reviewer", "unvalidated_available")
	var attempt struct {
		TerminalStatus        string `json:"terminal_status"`
		ProviderFailureReason string `json:"provider_failure_reason"`
		NoProviderExecution   bool   `json:"no_provider_execution"`
		ExitCode              *int   `json:"exit_code"`
	}
	if err := json.Unmarshal(attemptOutput, &attempt); err != nil {
		t.Fatalf("parse provider-attempt output: %v\n%s", err, attemptOutput)
	}
	if attempt.TerminalStatus == "PASS" || attempt.ProviderFailureReason != "adapter_proof_required" || !attempt.NoProviderExecution || attempt.ExitCode != nil {
		t.Fatalf("unvalidated provider-attempt must block before subprocess execution: %+v\n%s", attempt, attemptOutput)
	}
}

func TestMAR004ProviderPreflightMapsMissingCLIAndMissingModel(t *testing.T) {
	registry := writeTempProviderRegistry(t, map[string]any{
		"schema_version":    "mar.provider_lanes.v1",
		"default_reviewers": []string{"missing_cli", "missing_model"},
		"reviewers": map[string]any{
			"missing_cli": map[string]any{
				"command_lane":    "definitely-missing-mar-cli",
				"executable":      "definitely-missing-mar-cli",
				"selected_model":  "glm-5.2",
				"prompt_template": "templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl",
				"command_args":    []string{"--model", "{model}"},
				"validated":       false,
				"validation_required_before_success_coverage": true,
			},
			"missing_model": map[string]any{
				"command_lane":    "agy",
				"executable":      "agy",
				"selected_model":  nil,
				"prompt_template": "templates/prompts/mar/antigravity-gemini-reviewer-request.md.tmpl",
				"command_args":    []string{"--model", "{model}"},
				"validated":       false,
				"validation_required_before_success_coverage": true,
			},
		},
	})

	output := runMARCommand(t, "provider-preflight", "--registry", registry)
	var result struct {
		Status   string `json:"status"`
		Reason   string `json:"reason"`
		Attempts []struct {
			ReviewerID            string  `json:"reviewer_id"`
			TerminalStatus        string  `json:"terminal_status"`
			ProviderFailureReason string  `json:"provider_failure_reason"`
			SelectedModel         *string `json:"selected_model"`
			ExitCode              *int    `json:"exit_code"`
		} `json:"attempts"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse provider-preflight output: %v\n%s", err, output)
	}
	if result.Status == "PASS" {
		t.Fatalf("provider-preflight must fail closed when CLI/model evidence is absent\n%s", output)
	}
	requireAttemptReason(t, result.Attempts, "missing_cli", "cli_missing")
	requireAttemptReason(t, result.Attempts, "missing_model", "model_unavailable")
}

func TestMAR004PreScopedNarrowerSetRequiresPriorEvidence(t *testing.T) {
	output := runMARCommand(t, "provider-preflight", "--reviewers", "zcode_glm_5_2")
	var result struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse provider-preflight output: %v\n%s", err, output)
	}
	if result.Status != "BLOCKED" || result.Reason != "pre_scoped_evidence_required" {
		t.Fatalf("narrower reviewer set without evidence = %q/%q, want BLOCKED/pre_scoped_evidence_required\n%s", result.Status, result.Reason, output)
	}
}

func TestMAR004ProviderAttemptEmitsFailClosedMissingCLIArtifact(t *testing.T) {
	registry := writeTempProviderRegistry(t, map[string]any{
		"schema_version":    "mar.provider_lanes.v1",
		"default_reviewers": []string{"missing_cli"},
		"reviewers": map[string]any{
			"missing_cli": map[string]any{
				"command_lane":    "definitely-missing-mar-cli",
				"executable":      "definitely-missing-mar-cli",
				"selected_model":  "glm-5.2",
				"prompt_template": "templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl",
				"command_args":    []string{"--model", "{model}"},
				"validated":       false,
				"validation_required_before_success_coverage": true,
			},
		},
	})

	output := runMARCommand(t,
		"provider-attempt",
		"--registry", registry,
		"--reviewer", "missing_cli",
		"--run-id", "run-test",
		"--task-id", "MAR-004",
	)
	var attempt struct {
		SchemaVersion         string   `json:"schema_version"`
		ReviewerID            string   `json:"reviewer_id"`
		CommandLane           string   `json:"command_lane"`
		SelectedModel         string   `json:"selected_model"`
		TerminalStatus        string   `json:"terminal_status"`
		ProviderFailureReason string   `json:"provider_failure_reason"`
		ParserStatus          string   `json:"parser_status"`
		RedactedCommand       []string `json:"redacted_command"`
		MutationCheck         struct {
			Checked  bool `json:"checked"`
			Detected bool `json:"detected"`
		} `json:"mutation_check"`
		NoProviderExecution bool `json:"no_provider_execution"`
	}
	if err := json.Unmarshal(output, &attempt); err != nil {
		t.Fatalf("parse provider-attempt output: %v\n%s", err, output)
	}
	if attempt.SchemaVersion != "mar.provider_attempt.v1" || attempt.ReviewerID != "missing_cli" {
		t.Fatalf("provider-attempt identity mismatch: %+v\n%s", attempt, output)
	}
	if attempt.TerminalStatus == "PASS" || attempt.ProviderFailureReason != "cli_missing" {
		t.Fatalf("provider-attempt missing CLI = %q/%q, want non-PASS/cli_missing\n%s", attempt.TerminalStatus, attempt.ProviderFailureReason, output)
	}
	if attempt.ParserStatus != "not_run" || len(attempt.RedactedCommand) == 0 {
		t.Fatalf("provider-attempt missing parser/redacted command evidence\n%s", output)
	}
	if !attempt.MutationCheck.Checked || attempt.MutationCheck.Detected {
		t.Fatalf("provider-attempt missing clean mutation check\n%s", output)
	}
	if !attempt.NoProviderExecution {
		t.Fatalf("missing CLI attempt must record no_provider_execution=true\n%s", output)
	}
}

func TestMAR004ProviderMergePackResolvesCoverageOnlyWithExplicitEvidence(t *testing.T) {
	cases := []struct {
		name     string
		attempts []map[string]any
		args     []string
		want     string
		notWant  string
	}{
		{
			name: "partial default failure cannot pass",
			attempts: []map[string]any{
				providerAttemptFixture("z-ok", "zcode_glm_5_2", "PASS", nil),
				providerAttemptFixture("k-fail", "kimi_k2_7", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("a-ok", "antigravity_gemini", "PASS", nil),
			},
			want:    "DEGRADED",
			notWant: "PASS",
		},
		{
			name: "all default failure cannot pass",
			attempts: []map[string]any{
				providerAttemptFixture("z-fail", "zcode_glm_5_2", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("k-fail", "kimi_k2_7", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("a-fail", "antigravity_gemini", "BLOCKED", map[string]any{"provider_failure_reason": "model_unavailable"}),
			},
			want:    "FAILED",
			notWant: "PASS",
		},
		{
			name: "same provider retry success resolves failed coverage",
			attempts: []map[string]any{
				providerAttemptFixture("z-ok", "zcode_glm_5_2", "PASS", nil),
				providerAttemptFixture("k-fail", "kimi_k2_7", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("k-retry", "kimi_k2_7", "PASS", map[string]any{"retry_of_attempt_id": "k-fail"}),
				providerAttemptFixture("a-ok", "antigravity_gemini", "PASS", nil),
			},
			want: "PASS",
		},
		{
			name: "same provider retry failure stays non clean",
			attempts: []map[string]any{
				providerAttemptFixture("z-ok", "zcode_glm_5_2", "PASS", nil),
				providerAttemptFixture("k-fail", "kimi_k2_7", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("k-retry-fail", "kimi_k2_7", "DEGRADED", map[string]any{
					"provider_failure_reason": "nonzero_exit",
					"retry_of_attempt_id":     "k-fail",
				}),
				providerAttemptFixture("a-ok", "antigravity_gemini", "PASS", nil),
			},
			want:    "DEGRADED",
			notWant: "PASS",
		},
		{
			name: "alternate success without approval stays non clean",
			attempts: []map[string]any{
				providerAttemptFixture("z-ok", "zcode_glm_5_2", "PASS", nil),
				providerAttemptFixture("k-fail", "kimi_k2_7", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("alt-ok", "alternate_reviewer", "PASS", map[string]any{"alternate_for_reviewer_id": "kimi_k2_7"}),
				providerAttemptFixture("a-ok", "antigravity_gemini", "PASS", nil),
			},
			want:    "DEGRADED",
			notWant: "PASS",
		},
		{
			name: "approved alternate success resolves coverage",
			attempts: []map[string]any{
				providerAttemptFixture("z-ok", "zcode_glm_5_2", "PASS", nil),
				providerAttemptFixture("k-fail", "kimi_k2_7", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("alt-ok", "alternate_reviewer", "PASS", map[string]any{
					"alternate_for_reviewer_id": "kimi_k2_7",
					"approval_evidence":         "주군-approved alternate alt-ok",
				}),
				providerAttemptFixture("a-ok", "antigravity_gemini", "PASS", nil),
			},
			want: "PASS",
		},
		{
			name: "explicit waiver resolves coverage",
			attempts: []map[string]any{
				providerAttemptFixture("z-ok", "zcode_glm_5_2", "PASS", nil),
				providerAttemptFixture("k-ok", "kimi_k2_7", "PASS", nil),
				providerAttemptFixture("a-fail", "antigravity_gemini", "BLOCKED", map[string]any{"provider_failure_reason": "model_unavailable"}),
			},
			args: []string{"--waiver", "antigravity_gemini=주군-waiver-model-unavailable"},
			want: "PASS",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			paths := writeTempAttemptFixtures(t, tc.attempts)
			args := append([]string{"merge-pack", "--provider-coverage"}, tc.args...)
			args = append(args, paths...)
			output := runMARCommand(t, args...)
			var result struct {
				Status   string `json:"status"`
				Reason   string `json:"reason"`
				Coverage struct {
					UnresolvedDefaultReviewers []string `json:"unresolved_default_reviewers"`
				} `json:"coverage"`
			}
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("parse provider merge-pack output: %v\n%s", err, output)
			}
			if result.Status != tc.want {
				t.Fatalf("provider merge-pack status = %q, want %q (%s)\n%s", result.Status, tc.want, result.Reason, output)
			}
			if tc.notWant != "" && result.Status == tc.notWant {
				t.Fatalf("provider merge-pack status = forbidden %q\n%s", tc.notWant, output)
			}
		})
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

func writeTempProviderRegistry(t *testing.T, payload map[string]any) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mar-provider-lanes.json")
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal temp provider registry: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp provider registry: %v", err)
	}
	return path
}

func writeTempToolchain(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "toolchain.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp toolchain: %v", err)
	}
	return path
}

func writeTempAttemptFixtures(t *testing.T, attempts []map[string]any) []string {
	t.Helper()
	dir := t.TempDir()
	paths := make([]string, 0, len(attempts))
	for _, attempt := range attempts {
		name, _ := attempt["attempt_id"].(string)
		if name == "" {
			name = "attempt"
		}
		path := filepath.Join(dir, name+".json")
		data, err := json.MarshalIndent(attempt, "", "  ")
		if err != nil {
			t.Fatalf("marshal temp attempt %s: %v", name, err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("write temp attempt %s: %v", name, err)
		}
		paths = append(paths, path)
	}
	return paths
}

func providerAttemptFixture(attemptID, reviewerID, terminalStatus string, extra map[string]any) map[string]any {
	attempt := map[string]any{
		"schema_version":            "mar.provider_attempt.v1",
		"run_id":                    "run-test",
		"task_id":                   "MAR-004",
		"attempt_id":                attemptID,
		"reviewer_id":               reviewerID,
		"command_lane":              reviewerID,
		"selected_model":            "fixture-model",
		"started_at":                "2026-06-17T00:00:00Z",
		"ended_at":                  "2026-06-17T00:00:01Z",
		"timeout_seconds":           1,
		"exit_code":                 0,
		"terminal_status":           terminalStatus,
		"provider_failure_reason":   nil,
		"parser_status":             "parsed",
		"mutation_check":            map[string]any{"checked": true, "detected": false},
		"redacted_command":          []string{"fixture-reviewer"},
		"raw_output_path":           "raw.txt",
		"parsed_finding_path":       "parsed.json",
		"capped_output_note":        map[string]any{"cap_bytes": 4096, "truncated": false},
		"retry_of_attempt_id":       nil,
		"alternate_for_reviewer_id": nil,
		"approval_evidence":         nil,
		"waiver_evidence":           nil,
	}
	for key, value := range extra {
		attempt[key] = value
	}
	return attempt
}

func requireAttemptReason(t *testing.T, attempts []struct {
	ReviewerID            string  `json:"reviewer_id"`
	TerminalStatus        string  `json:"terminal_status"`
	ProviderFailureReason string  `json:"provider_failure_reason"`
	SelectedModel         *string `json:"selected_model"`
	ExitCode              *int    `json:"exit_code"`
}, reviewerID, reason string) {
	t.Helper()
	for _, attempt := range attempts {
		if attempt.ReviewerID == reviewerID {
			if attempt.TerminalStatus == "PASS" || attempt.ProviderFailureReason != reason {
				t.Fatalf("attempt %s status/reason = %q/%q, want non-PASS/%s", reviewerID, attempt.TerminalStatus, attempt.ProviderFailureReason, reason)
			}
			return
		}
	}
	t.Fatalf("missing attempt for reviewer %q in %+v", reviewerID, attempts)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
