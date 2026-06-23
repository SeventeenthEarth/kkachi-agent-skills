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
	output := runMARCommand(t, "--help")
	help := string(output)
	for _, want := range []string{"doctor", "render", "validate", "merge-pack", "role-lanes", "provider-lanes", "provider-preflight", "provider-attempt"} {
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

func TestMAR005ValidateParsesKimiStreamJSONAssistantContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kimi-stream.jsonl")
	raw := "{\"role\":\"assistant\",\"content\":\"```json\\n{\\\"status\\\":\\\"PASS\\\",\\\"summary\\\":\\\"stream parsed\\\",\\\"confidence\\\":0.9,\\\"findings\\\":[],\\\"role_scoped_acceptance_criteria_verdicts\\\":[]}\\n```\"}\n" +
		"{\"role\":\"meta\",\"type\":\"session.resume_hint\",\"content\":\"resume\"}"
	if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
		t.Fatalf("write kimi stream fixture: %v", err)
	}

	output := runMARCommand(t,
		"validate",
		"--input", path,
	)
	var result struct {
		Status  string `json:"status"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse validate output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.Summary != "stream parsed" {
		t.Fatalf("kimi stream-json assistant content must parse as review JSON\n%s", output)
	}
}

func TestMAR005ValidateParsesDirectFencedJSONProviderOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fenced-provider-output.txt")
	raw := "```json\n{\"status\":\"PASS\",\"summary\":\"direct fenced parsed\",\"confidence\":\"high\",\"findings\":[],\"role_scoped_acceptance_criteria_verdicts\":[]}\n```\n"
	if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
		t.Fatalf("write direct fenced fixture: %v", err)
	}

	output := runMARCommand(t,
		"validate",
		"--input", path,
	)
	var result struct {
		Status  string `json:"status"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse validate output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.Summary != "direct fenced parsed" {
		t.Fatalf("direct fenced provider output must parse as review JSON\n%s", output)
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
	if !result.NoProviderExecution || result.FixtureEvidence.Status != "DEGRADED" || !result.FixtureEvidence.NoProviderExecution {
		t.Fatalf("doctor fixture evidence mismatch\n%s", output)
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
		{"request changes beats pass", []string{"merge-pack", "tests/fixtures/mar/pass-review.json", "tests/fixtures/mar/request-changes-review.json"}, "REQUEST_CHANGES"},
		{"all failed remains failed", []string{"merge-pack", "tests/fixtures/mar/all-failed.json"}, "FAILED"},
		{"insufficient coverage blocks", []string{"merge-pack", "--required-reviewers", "3", "tests/fixtures/mar/pass-review.json"}, "BLOCKED"},
		{"provider unavailable degrades", []string{"merge-pack", "tests/fixtures/mar/provider-unavailable.json"}, "DEGRADED"},
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
			if result.Status != tc.want || !result.NoProviderExecution {
				t.Fatalf("merge-pack status/no-provider = %q/%v, want %q/true\n%s", result.Status, result.NoProviderExecution, tc.want, output)
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
	if result.Status != "PASS" || !result.NoProviderExecution {
		t.Fatalf("render status/no-provider = %q/%v\n%s", result.Status, result.NoProviderExecution, output)
	}
	for _, want := range []string{"run-local", "MAR-003", "contract text", "diff text"} {
		if !strings.Contains(result.RenderedPrompt, want) {
			t.Fatalf("rendered prompt missing %q\n%s", want, result.RenderedPrompt)
		}
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

func TestMAR005RoleLaneRegistryReadbackIsFailClosed(t *testing.T) {
	output := runMARCommand(t, "role-lanes")
	var result struct {
		Status        string   `json:"status"`
		SchemaVersion string   `json:"schema_version"`
		RequiredRoles []string `json:"required_roles"`
		OptionalRoles []string `json:"optional_roles"`
		Roles         map[string]struct {
			PrimaryProvider   string `json:"primary_provider"`
			SecondaryProvider string `json:"secondary_provider"`
			Description       string `json:"description"`
		} `json:"roles"`
		Providers map[string]struct {
			SelectedModel         *string  `json:"selected_model"`
			SelectedModelRequired *bool    `json:"selected_model_required"`
			ModelSelection        string   `json:"model_selection"`
			DefaultRequired       bool     `json:"default_required"`
			CommandArgs           []string `json:"command_args"`
			TimeoutSeconds        int      `json:"timeout_seconds"`
			DegradedPosture       string   `json:"degraded_posture"`
		} `json:"providers"`
		ProviderFailureReasons []string `json:"provider_failure_reasons"`
		NoProviderExecution    bool     `json:"no_provider_execution"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse role-lanes output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.SchemaVersion != "mar.role_lanes.v1" {
		t.Fatalf("role-lanes status/schema = %q/%q\n%s", result.Status, result.SchemaVersion, output)
	}
	wantRoles := []string{"logic", "security", "arch", "cve", "test_adequacy"}
	if strings.Join(result.RequiredRoles, ",") != strings.Join(wantRoles, ",") {
		t.Fatalf("required roles = %v, want %v", result.RequiredRoles, wantRoles)
	}
	if strings.Join(result.OptionalRoles, ",") != "vision" {
		t.Fatalf("optional roles = %v, want [vision]", result.OptionalRoles)
	}
	wantMatrix := map[string][2]string{
		"logic":         {"zcode_glm_5_2", "kimi_default"},
		"security":      {"antigravity_gemini", "zcode_glm_5_2"},
		"arch":          {"zcode_glm_5_2", "kimi_default"},
		"cve":           {"antigravity_gemini", "zcode_glm_5_2"},
		"test_adequacy": {"kimi_default", "zcode_glm_5_2"},
	}
	for _, roleID := range wantRoles {
		role := result.Roles[roleID]
		want := wantMatrix[roleID]
		if role.PrimaryProvider != want[0] || role.SecondaryProvider != want[1] || role.Description == "" {
			t.Fatalf("role %s has wrong provider candidates: %+v", roleID, role)
		}
	}
	zcode := result.Providers["zcode_glm_5_2"]
	zcodeArgs := strings.Join(zcode.CommandArgs, "\x00")
	if !zcode.DefaultRequired || !containsString(zcode.CommandArgs, "--prompt-file") || !containsString(zcode.CommandArgs, "{prompt_path}") {
		t.Fatalf("zcode_glm_5_2 must use adapter prompt-file mode: %+v", zcode)
	}
	if zcode.TimeoutSeconds < 1800 {
		t.Fatalf("zcode_glm_5_2 must allow a 30 minute MAR review timeout, got %d", zcode.TimeoutSeconds)
	}
	for _, forbidden := range []string{"--model", "{prompt_text}"} {
		if strings.Contains(zcodeArgs, forbidden) {
			t.Fatalf("zcode_glm_5_2 command_args use unsupported option/data %q: %v", forbidden, zcode.CommandArgs)
		}
	}
	kimi := result.Providers["kimi_default"]
	kimiArgs := strings.Join(kimi.CommandArgs, "\x00")
	if !kimi.DefaultRequired || kimi.SelectedModel != nil || kimi.SelectedModelRequired == nil || *kimi.SelectedModelRequired || kimi.ModelSelection != "cli_default_latest" {
		t.Fatalf("kimi_default must use CLI default/latest without explicit selected_model: %+v", kimi)
	}
	if !containsString(kimi.CommandArgs, "--prompt-file") || !containsString(kimi.CommandArgs, "{prompt_path}") {
		t.Fatalf("kimi_default must pass prompt by adapter prompt-file mode: %+v", kimi)
	}
	for _, forbidden := range []string{"--model", "{prompt_text}"} {
		if strings.Contains(kimiArgs, forbidden) {
			t.Fatalf("kimi_default command_args use unsupported explicit model/text option %q: %v", forbidden, kimi.CommandArgs)
		}
	}
	agy := result.Providers["antigravity_gemini"]
	if agy.DefaultRequired || agy.SelectedModel == nil || *agy.SelectedModel != "Gemini 3.5 Flash (High)" || agy.SelectedModelRequired == nil || *agy.SelectedModelRequired || agy.DegradedPosture == "" {
		t.Fatalf("antigravity_gemini must pin selected model and remain explicit non-default metadata: %+v", agy)
	}
	if strings.Contains(strings.Join(agy.CommandArgs, "\x00"), "read-only") || !containsString(agy.CommandArgs, "--prompt-file") {
		t.Fatalf("antigravity command_args must use adapter prompt-file and must not pass a value to --sandbox: %v", agy.CommandArgs)
	}
	for _, reason := range []string{"cli_missing", "model_unavailable", "mutation_detected", "unknown_provider_failure"} {
		if !containsString(result.ProviderFailureReasons, reason) {
			t.Fatalf("role-lanes missing failure reason %q in %v", reason, result.ProviderFailureReasons)
		}
	}
	if !result.NoProviderExecution {
		t.Fatalf("role-lanes must not execute providers\n%s", output)
	}
}

func TestMAR005ProviderToolchainOverlayResolvesExplicitArgv(t *testing.T) {
	registry := writeTempProviderRegistry(t, roleRegistryPayload([]string{"logic"}, map[string]any{
		"alias_backed":     providerPayload("zcode", "definitely-missing-shell-alias", "glm-5.2", []string{"--version"}, false),
		"secondary_backed": providerPayload("kimi", "python3", "fixture-kimi-default", []string{"--version"}, true),
	}, map[string]any{"logic": rolePayload("alias_backed", "secondary_backed")}))
	toolchain := writeTempToolchain(t, `kas_cli: v0.1.8
kah_cli: v0.1.14
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
			RoleID                string   `json:"role_id"`
			ProviderID            string   `json:"provider_id"`
			ProviderCandidate     string   `json:"provider_candidate"`
			TerminalStatus        string   `json:"terminal_status"`
			ProviderFailureReason *string  `json:"provider_failure_reason"`
			RedactedCommand       []string `json:"redacted_command"`
			NoProviderExecution   bool     `json:"no_provider_execution"`
		} `json:"attempts"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse provider-preflight output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.Reason != "provider_preflight_passed" || len(result.Attempts) != 1 {
		t.Fatalf("toolchain-resolved provider preflight mismatch\n%s", output)
	}
	attempt := result.Attempts[0]
	if attempt.RoleID != "logic" || attempt.ProviderID != "alias_backed" || attempt.ProviderCandidate != "primary" {
		t.Fatalf("unexpected preflight attempt identity: %+v\n%s", attempt, output)
	}
	if attempt.TerminalStatus != "PASS" || attempt.ProviderFailureReason != nil || !attempt.NoProviderExecution {
		t.Fatalf("toolchain-resolved preflight attempt = %+v\n%s", attempt, output)
	}
	if strings.Join(attempt.RedactedCommand, " ") != "python3 --version" {
		t.Fatalf("toolchain resolved_argv must replace missing registry executable, got %v\n%s", attempt.RedactedCommand, output)
	}
}

func TestMAR005ProviderToolchainOverlayAcceptsNestedMARProviderTools(t *testing.T) {
	registry := writeTempProviderRegistry(t, roleRegistryPayload([]string{"logic"}, map[string]any{
		"alias_backed":     providerPayload("zcode", "definitely-missing-shell-alias", "glm-5.2", []string{"--version"}, false),
		"secondary_backed": providerPayload("kimi", "python3", "fixture-kimi-default", []string{"--version"}, true),
	}, map[string]any{"logic": rolePayload("alias_backed", "secondary_backed")}))
	toolchain := writeTempToolchain(t, `schema_version: "kkachi.toolchain.v1"
mar:
  provider_tools:
    schema_version: mar.provider_tools.v1
    providers:
      alias_backed:
        command_lane: zcode
        resolved_argv:
          - python3
        selected_model: glm-5.2
        version: 3.x
        validated: true
        adapter_proof_evidence: .kkachi/runs/run-test/mar/provider-proof-alias_backed.json
`)

	output := runMARCommand(t, "provider-preflight", "--registry", registry, "--toolchain", toolchain)
	var result struct {
		Status   string `json:"status"`
		Reason   string `json:"reason"`
		Attempts []struct {
			ProviderID            string   `json:"provider_id"`
			TerminalStatus        string   `json:"terminal_status"`
			ProviderFailureReason *string  `json:"provider_failure_reason"`
			RedactedCommand       []string `json:"redacted_command"`
		} `json:"attempts"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse nested provider-preflight output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.Reason != "provider_preflight_passed" || len(result.Attempts) != 1 {
		t.Fatalf("nested mar.provider_tools overlay did not pass\n%s", output)
	}
	attempt := result.Attempts[0]
	if attempt.ProviderID != "alias_backed" || attempt.TerminalStatus != "PASS" || attempt.ProviderFailureReason != nil {
		t.Fatalf("nested overlay attempt mismatch: %+v\n%s", attempt, output)
	}
	if strings.Join(attempt.RedactedCommand, " ") != "python3 --version" {
		t.Fatalf("nested overlay resolved_argv must replace missing registry executable, got %v\n%s", attempt.RedactedCommand, output)
	}
}

func TestMAR005MaterializedProjectScriptResolvesProjectRootRegistry(t *testing.T) {
	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".kkachi", "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, ".kkachi", "registries"), 0o755); err != nil {
		t.Fatal(err)
	}
	scriptData, err := os.ReadFile(filepath.Join(repoRoot(t), "scripts", "mar.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ".kkachi", "scripts", "mar.py"), scriptData, 0o755); err != nil {
		t.Fatal(err)
	}
	registryPayload := roleRegistryPayload([]string{"logic"}, map[string]any{
		"alias_backed": providerPayload("zcode", "definitely-missing-shell-alias", "glm-5.2", []string{"--version"}, false),
	}, map[string]any{"logic": rolePayload("alias_backed", "alias_backed")})
	registryData, err := json.MarshalIndent(registryPayload, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, ".kkachi", "registries", "mar-provider-lanes.json"), registryData, 0o644); err != nil {
		t.Fatal(err)
	}
	toolchain := `schema_version: "kkachi.toolchain.v1"
mar:
  provider_tools:
    schema_version: mar.provider_tools.v1
    providers:
      alias_backed:
        command_lane: zcode
        resolved_argv:
          - python3
        selected_model: glm-5.2
        version: 3.x
        validated: true
        adapter_proof_evidence: .kkachi/runs/run-test/mar/provider-proof-alias_backed.json
`
	if err := os.WriteFile(filepath.Join(projectRoot, ".kkachi", "toolchain.yaml"), []byte(toolchain), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("python3", ".kkachi/scripts/mar.py", "provider-preflight", "--registry", ".kkachi/registries/mar-provider-lanes.json", "--toolchain", ".kkachi/toolchain.yaml")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("materialized project MAR preflight failed: %v\n%s", err, output)
	}
	var result struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse materialized project MAR preflight output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.Reason != "provider_preflight_passed" || strings.Contains(result.Detail, ".kkachi/.kkachi") {
		t.Fatalf("materialized MAR script must resolve registry/toolchain from project root, got %+v\n%s", result, output)
	}
}

func TestMAR005ProviderToolchainOverlayAcceptsEmptyNestedProviderMap(t *testing.T) {
	registry := writeTempProviderRegistry(t, roleRegistryPayload([]string{"logic"}, map[string]any{
		"unvalidated": providerPayload("zcode", "python3", "glm-5.2", []string{"--version"}, false),
	}, map[string]any{"logic": rolePayload("unvalidated", "unvalidated")}))
	toolchain := writeTempToolchain(t, `schema_version: "kkachi.toolchain.v1"
mar:
  provider_tools:
    schema_version: mar.provider_tools.v1
    providers: {}
`)

	output := runMARCommand(t, "provider-preflight", "--registry", registry, "--toolchain", toolchain)
	var result struct {
		Status   string `json:"status"`
		Reason   string `json:"reason"`
		Detail   string `json:"detail"`
		Attempts []struct {
			ProviderID            string  `json:"provider_id"`
			TerminalStatus        string  `json:"terminal_status"`
			ProviderFailureReason *string `json:"provider_failure_reason"`
			PreflightEvidencePath *string `json:"preflight_evidence_path"`
		} `json:"attempts"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse empty nested provider map preflight output: %v\n%s", err, output)
	}
	if result.Reason == "provider_registry_failure" || result.Detail != "" {
		t.Fatalf("empty nested providers map must not be treated as a registry parse failure\n%s", output)
	}
	if result.Status != "FAILED" || result.Reason != "all_provider_preflight_failed" || len(result.Attempts) == 0 {
		t.Fatalf("empty nested providers map should preserve fail-closed provider preflight semantics\n%s", output)
	}
	attempt := result.Attempts[0]
	if attempt.ProviderID != "unvalidated" || attempt.TerminalStatus != "BLOCKED" || attempt.ProviderFailureReason == nil || *attempt.ProviderFailureReason != "adapter_proof_required" {
		t.Fatalf("empty nested providers map attempt mismatch: %+v\n%s", attempt, output)
	}
}

func TestMAR005ProviderSafetyAndPreflightFailures(t *testing.T) {
	registry := writeTempProviderRegistry(t, roleRegistryPayload([]string{"logic"}, map[string]any{
		"missing_cli":       providerPayload("missing", "definitely-missing-mar-cli", "glm-5.2", []string{"--model", "{model}"}, false),
		"missing_model":     providerPayload("agy", "agy", nil, []string{"--model", "{model}"}, false),
		"unvalidated":       providerPayload("python3", "python3", "fixture-model", []string{"--version"}, false),
		"unvalidated_other": providerPayload("python3", "python3", "fixture-model", []string{"--version"}, false),
	}, map[string]any{"logic": rolePayload("missing_cli", "missing_model")}))

	output := runMARCommand(t, "provider-preflight", "--registry", registry)
	var result struct {
		Status   string          `json:"status"`
		Reason   string          `json:"reason"`
		Attempts []attemptReason `json:"attempts"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse provider-preflight output: %v\n%s", err, output)
	}
	if result.Status == "PASS" {
		t.Fatalf("provider-preflight must fail closed when CLI/model evidence is absent\n%s", output)
	}
	requireAttemptReason(t, result.Attempts, "missing_cli", "cli_missing")
	requireAttemptReason(t, result.Attempts, "missing_model", "model_unavailable")

	blocked := runMARCommand(t, "provider-preflight", "--roles", "logic")
	var blockedResult struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(blocked, &blockedResult); err != nil {
		t.Fatalf("parse narrower preflight output: %v\n%s", err, blocked)
	}
	if blockedResult.Status != "BLOCKED" || blockedResult.Reason != "pre_scoped_evidence_required" {
		t.Fatalf("narrower role set without evidence = %q/%q\n%s", blockedResult.Status, blockedResult.Reason, blocked)
	}

	unvalidatedRegistry := writeTempProviderRegistry(t, roleRegistryPayload([]string{"logic"}, map[string]any{
		"unvalidated":       providerPayload("python3", "python3", "fixture-model", []string{"--version"}, false),
		"unvalidated_other": providerPayload("python3", "python3", "fixture-model", []string{"--version"}, false),
	}, map[string]any{"logic": rolePayload("unvalidated", "unvalidated_other")}))
	attemptOutput := runMARCommand(t, "provider-attempt", "--registry", unvalidatedRegistry, "--provider", "unvalidated")
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

func TestMAR005ProviderPreflightAllowsExplicitDefaultModelLane(t *testing.T) {
	defaultModel := providerPayload("kimi", "python3", nil, []string{"--version"}, true)
	defaultModel["selected_model_required"] = false
	defaultModel["model_selection"] = "cli_default_latest"
	registry := writeTempProviderRegistry(t, roleRegistryPayload([]string{"logic"}, map[string]any{
		"default_model": defaultModel,
	}, map[string]any{"logic": rolePayload("default_model", "default_model")}))

	output := runMARCommand(t, "provider-preflight", "--registry", registry)
	var result struct {
		Status   string          `json:"status"`
		Reason   string          `json:"reason"`
		Attempts []attemptReason `json:"attempts"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse default-model provider-preflight output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || result.Reason != "provider_preflight_passed" || len(result.Attempts) != 1 {
		t.Fatalf("default-model provider-preflight mismatch\n%s", output)
	}
	attempt := result.Attempts[0]
	if attempt.ProviderID != "default_model" || attempt.SelectedModel != nil || attempt.TerminalStatus != "PASS" || attempt.ProviderFailureReason != "" {
		t.Fatalf("default-model preflight attempt should pass without selected_model: %+v\n%s", attempt, output)
	}
}

func TestMAR005ProviderAttemptEmitsFailClosedMissingCLIArtifact(t *testing.T) {
	registry := writeTempProviderRegistry(t, roleRegistryPayload([]string{"logic"}, map[string]any{
		"missing_cli":       providerPayload("missing", "definitely-missing-mar-cli", "glm-5.2", []string{"--model", "{model}"}, false),
		"secondary_missing": providerPayload("missing", "definitely-missing-secondary-cli", "fixture-kimi-default", []string{"--model", "{model}"}, false),
	}, map[string]any{"logic": rolePayload("missing_cli", "secondary_missing")}))

	output := runMARCommand(t,
		"provider-attempt",
		"--registry", registry,
		"--provider", "missing_cli",
		"--run-id", "run-test",
		"--task-id", "MAR-005",
	)
	var attempt struct {
		SchemaVersion         string   `json:"schema_version"`
		ProviderID            string   `json:"provider_id"`
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
	if attempt.SchemaVersion != "mar.provider_attempt.v1" || attempt.ProviderID != "missing_cli" {
		t.Fatalf("provider-attempt identity mismatch: %+v\n%s", attempt, output)
	}
	if attempt.TerminalStatus == "PASS" || attempt.ProviderFailureReason != "cli_missing" || attempt.ParserStatus != "not_run" {
		t.Fatalf("provider-attempt missing CLI metadata mismatch\n%s", output)
	}
	if len(attempt.RedactedCommand) == 0 || !attempt.MutationCheck.Checked || attempt.MutationCheck.Detected || !attempt.NoProviderExecution {
		t.Fatalf("provider-attempt safety evidence mismatch\n%s", output)
	}
}

func TestMAR005RoleProviderAttemptMergesParsedReviewerPayload(t *testing.T) {
	provider := writeTempExecutable(t, "mar-provider-reviewer", `#!/bin/sh
cat <<'JSON'
{"status":"PASS_WITH_FINDINGS","summary":"reviewed role","confidence":0.6,"findings":[{"id":"LIVE-001","severity":"blocker","message":"synthetic blocker"}],"role_scoped_acceptance_criteria_verdicts":[{"criterion":"role/provider separation","verdict":"pass","evidence":"fixture"}],"no_provider_execution":false}
JSON
`)
	secondary := writeTempExecutable(t, "mar-provider-secondary", `#!/bin/sh
echo '{"status":"PASS","summary":"secondary should not run"}'
`)
	registry := writeTempProviderRegistry(t, roleRegistryPayload([]string{"test_adequacy"}, map[string]any{
		"primary":   providerPayload("fixture", provider, "fixture-model", []string{}, true),
		"secondary": providerPayload("fixture", secondary, "fixture-model", []string{}, true),
	}, map[string]any{"test_adequacy": rolePayload("primary", "secondary")}))

	output := runMARCommand(t,
		"provider-attempt",
		"--registry", registry,
		"--role", "test_adequacy",
		"--run-id", "run-test",
		"--task-id", "MAR-005",
		"--pre-scoped-evidence", "live-provider-test",
	)
	var result struct {
		Status       string `json:"status"`
		RoleCoverage struct {
			RedTriggerSummary struct {
				RedAdjudicationRequired bool     `json:"red_adjudication_required"`
				Triggers                []string `json:"triggers"`
			} `json:"red_trigger_summary"`
			BlueMatrixInputs struct {
				CoveredRoles             []string                    `json:"covered_roles"`
				AcceptanceCriteriaMatrix map[string][]map[string]any `json:"acceptance_criteria_matrix"`
			} `json:"blue_matrix_inputs"`
		} `json:"role_coverage"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse role provider-attempt output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || !containsString(result.RoleCoverage.BlueMatrixInputs.CoveredRoles, "test_adequacy") {
		t.Fatalf("role provider-attempt must cover test_adequacy when primary provider succeeds\n%s", output)
	}
	if len(result.RoleCoverage.BlueMatrixInputs.AcceptanceCriteriaMatrix["test_adequacy"]) != 1 {
		t.Fatalf("parsed provider AC verdicts must feed Blue matrix\n%s", output)
	}
	if !result.RoleCoverage.RedTriggerSummary.RedAdjudicationRequired || !containsString(result.RoleCoverage.RedTriggerSummary.Triggers, "low_confidence") || !containsString(result.RoleCoverage.RedTriggerSummary.Triggers, "blocker") {
		t.Fatalf("parsed confidence/finding severity must feed Red trigger summary\n%s", output)
	}
}

func TestMAR005RoleMergePackCoversPrimarySecondaryAndFailClosedPaths(t *testing.T) {
	cases := []struct {
		name       string
		attempts   []map[string]any
		want       string
		wantReason string
		notWant    string
	}{
		{
			name:       "primary success covers role",
			attempts:   []map[string]any{providerAttemptFixture("logic-primary", "logic", "zcode_glm_5_2", "primary", "PASS", nil)},
			want:       "PASS",
			wantReason: "all_required_role_coverage_resolved",
		},
		{
			name: "primary failure secondary success covers role",
			attempts: []map[string]any{
				providerAttemptFixture("logic-primary", "logic", "zcode_glm_5_2", "primary", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("logic-secondary", "logic", "kimi_default", "secondary", "PASS", map[string]any{"acceptance_criteria_verdicts": []map[string]any{{"id": "AC-1", "verdict": "covered"}}}),
			},
			want:       "PASS",
			wantReason: "all_required_role_coverage_resolved",
		},
		{
			name: "primary and secondary failure fails closed",
			attempts: []map[string]any{
				providerAttemptFixture("logic-primary", "logic", "zcode_glm_5_2", "primary", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("logic-secondary", "logic", "kimi_default", "secondary", "BLOCKED", map[string]any{"provider_failure_reason": "model_unavailable"}),
			},
			want:       "FAILED",
			wantReason: "all_required_roles_unresolved",
			notWant:    "PASS",
		},
		{
			name: "undeclared tertiary cannot cover role",
			attempts: []map[string]any{
				providerAttemptFixture("logic-primary", "logic", "zcode_glm_5_2", "primary", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
				providerAttemptFixture("logic-secondary", "logic", "kimi_default", "secondary", "DEGRADED", map[string]any{"provider_failure_reason": "nonzero_exit"}),
				providerAttemptFixture("logic-tertiary", "logic", "antigravity_gemini", "tertiary", "PASS", nil),
			},
			want:    "FAILED",
			notWant: "PASS",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			paths := writeTempAttemptFixtures(t, tc.attempts)
			args := append([]string{"merge-pack", "--provider-coverage", "--roles", "logic", "--pre-scoped-evidence", "test"}, paths...)
			output := runMARCommand(t, args...)
			var result roleMergeResult
			if err := json.Unmarshal(output, &result); err != nil {
				t.Fatalf("parse role merge-pack output: %v\n%s", err, output)
			}
			if result.Status != tc.want {
				t.Fatalf("role merge-pack status = %q, want %q (%s)\n%s", result.Status, tc.want, result.Reason, output)
			}
			if tc.wantReason != "" && result.Reason != tc.wantReason {
				t.Fatalf("role merge-pack reason = %q, want %q\n%s", result.Reason, tc.wantReason, output)
			}
			if tc.notWant != "" && result.Status == tc.notWant {
				t.Fatalf("role merge-pack status = forbidden %q\n%s", tc.notWant, output)
			}
			if result.Status == "FAILED" && (!containsString(result.UnresolvedRequiredRoles, "logic") || !strings.Contains(result.OperatorReportText, "주군/operator report")) {
				t.Fatalf("failed role coverage must report unresolved role to operator\n%s", output)
			}
			if tc.name == "primary failure secondary success covers role" {
				role := result.Coverage.ByRole["logic"]
				if role.ProviderID != "kimi_default" || role.ProviderCandidate != "secondary" || role.FallbackReason != "cli_missing" {
					t.Fatalf("secondary coverage metadata wrong: %+v\n%s", role, output)
				}
				if len(result.BlueMatrixInputs.AcceptanceCriteriaMatrix["logic"]) != 1 {
					t.Fatalf("secondary success must feed role-scoped AC verdicts into Blue matrix\n%s", output)
				}
			}
		})
	}
}

func TestMAR005RedTriggerDetectionAndTestAdequacyCoverage(t *testing.T) {
	paths := writeTempAttemptFixtures(t, []map[string]any{
		providerAttemptFixture("test-primary", "test_adequacy", "zcode_glm_5_2", "primary", "PASS_WITH_FINDINGS", map[string]any{
			"confidence": 0.6,
			"acceptance_criteria_verdicts": []map[string]any{
				{"id": "AC-9", "verdict": "test_gap"},
			},
		}),
	})
	args := append([]string{"merge-pack", "--provider-coverage", "--roles", "test_adequacy", "--pre-scoped-evidence", "test"}, paths...)
	output := runMARCommand(t, args...)
	var result roleMergeResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse test_adequacy merge-pack output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || !containsString(result.BlueMatrixInputs.CoveredRoles, "test_adequacy") {
		t.Fatalf("test_adequacy role must produce clean role coverage when provider succeeds\n%s", output)
	}
	if !result.RedTriggerSummary.RedAdjudicationRequired || !containsString(result.RedTriggerSummary.Triggers, "low_confidence") {
		t.Fatalf("low confidence role output must trigger Red summary\n%s", output)
	}
	if len(result.BlueMatrixInputs.AcceptanceCriteriaMatrix["test_adequacy"]) != 1 {
		t.Fatalf("test_adequacy AC verdict must feed Blue matrix\n%s", output)
	}
}

func TestMARConfidenceStringsDoNotCrashRedTriggerSummary(t *testing.T) {
	paths := writeTempAttemptFixtures(t, []map[string]any{
		providerAttemptFixture("logic-primary", "logic", "zcode_glm_5_2", "primary", "PASS", map[string]any{
			"confidence": "high",
		}),
		providerAttemptFixture("security-primary", "security", "zcode_glm_5_2", "primary", "PASS", map[string]any{
			"confidence": "low",
		}),
	})
	args := append([]string{"merge-pack", "--provider-coverage", "--roles", "logic,security", "--pre-scoped-evidence", "string-confidence-test"}, paths...)
	output := runMARCommand(t, args...)
	var result roleMergeResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse string confidence merge-pack output: %v\n%s", err, output)
	}
	if result.Status != "PASS" || !containsString(result.BlueMatrixInputs.CoveredRoles, "logic") || !containsString(result.BlueMatrixInputs.CoveredRoles, "security") {
		t.Fatalf("string confidence role attempts should still cover roles\n%s", output)
	}
	if !result.RedTriggerSummary.RedAdjudicationRequired || !containsString(result.RedTriggerSummary.Triggers, "low_confidence") {
		t.Fatalf("string low confidence must trigger Red summary without crashing\n%s", output)
	}
}

func TestMAR005RoleMergePackDegradesPartialRequiredCoverage(t *testing.T) {
	paths := writeTempAttemptFixtures(t, []map[string]any{
		providerAttemptFixture("logic-primary", "logic", "zcode_glm_5_2", "primary", "PASS", nil),
		providerAttemptFixture("security-primary", "security", "zcode_glm_5_2", "primary", "DEGRADED", map[string]any{"provider_failure_reason": "cli_missing"}),
		providerAttemptFixture("security-secondary", "security", "kimi_default", "secondary", "BLOCKED", map[string]any{"provider_failure_reason": "model_unavailable"}),
	})
	args := append([]string{"merge-pack", "--provider-coverage", "--roles", "logic,security", "--pre-scoped-evidence", "partial-test"}, paths...)
	output := runMARCommand(t, args...)
	var result roleMergeResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse partial role merge-pack output: %v\n%s", err, output)
	}
	if result.Status != "DEGRADED" || !containsString(result.BlueMatrixInputs.CoveredRoles, "logic") || !containsString(result.UnresolvedRequiredRoles, "security") {
		t.Fatalf("partial role coverage must degrade with unresolved security role\n%s", output)
	}
	if !strings.Contains(result.OperatorReportText, "security") {
		t.Fatalf("partial role coverage must report unresolved security role\n%s", output)
	}
}

func TestMARTL002ProviderAdaptersRoleMatrixAndVisionMetadata(t *testing.T) {
	output := runMARCommand(t, "role-lanes")
	var result struct {
		Status        string   `json:"status"`
		RequiredRoles []string `json:"required_roles"`
		OptionalRoles []string `json:"optional_roles"`
		Roles         map[string]struct {
			Required                 bool     `json:"required"`
			PrimaryProvider          string   `json:"primary_provider"`
			SecondaryProvider        string   `json:"secondary_provider"`
			TriggeredRequiredWhen    []string `json:"triggered_required_when"`
			VisionCoverageLimitation string   `json:"vision_coverage_limitation"`
		} `json:"roles"`
		Providers map[string]struct {
			SelectedModel         *string  `json:"selected_model"`
			SelectedModelRequired *bool    `json:"selected_model_required"`
			CommandArgs           []string `json:"command_args"`
			AdapterScript         string   `json:"adapter_script"`
			Validated             bool     `json:"validated"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("parse role-lanes output: %v\n%s", err, output)
	}
	if result.Status != "PASS" {
		t.Fatalf("role-lanes status = %q\n%s", result.Status, output)
	}
	if strings.Join(result.RequiredRoles, ",") != "logic,security,arch,cve,test_adequacy" {
		t.Fatalf("required roles changed unexpectedly: %v", result.RequiredRoles)
	}
	if strings.Join(result.OptionalRoles, ",") != "vision" {
		t.Fatalf("optional roles = %v, want [vision]", result.OptionalRoles)
	}
	wantMatrix := map[string][2]string{
		"logic":         {"zcode_glm_5_2", "kimi_default"},
		"security":      {"antigravity_gemini", "zcode_glm_5_2"},
		"arch":          {"zcode_glm_5_2", "kimi_default"},
		"cve":           {"antigravity_gemini", "zcode_glm_5_2"},
		"test_adequacy": {"kimi_default", "zcode_glm_5_2"},
		"vision":        {"antigravity_gemini", "zcode_glm_5_2"},
	}
	for roleID, want := range wantMatrix {
		role, ok := result.Roles[roleID]
		if !ok {
			t.Fatalf("missing role %s in role-lanes output", roleID)
		}
		if role.PrimaryProvider != want[0] || role.SecondaryProvider != want[1] {
			t.Fatalf("role %s providers = %s/%s, want %s/%s", roleID, role.PrimaryProvider, role.SecondaryProvider, want[0], want[1])
		}
	}
	vision := result.Roles["vision"]
	if vision.Required || len(vision.TriggeredRequiredWhen) == 0 || !containsString(vision.TriggeredRequiredWhen, "ui") || vision.VisionCoverageLimitation == "" {
		t.Fatalf("vision role must be optional with trigger metadata and fallback limitation: %+v", vision)
	}
	for providerID, script := range map[string]string{
		"zcode_glm_5_2":      "scripts/mar_adapters/mar-zcode.sh",
		"kimi_default":       "scripts/mar_adapters/mar-kimi.sh",
		"antigravity_gemini": "scripts/mar_adapters/mar-agy.sh",
	} {
		provider := result.Providers[providerID]
		if provider.AdapterScript != script {
			t.Fatalf("provider %s adapter_script = %q, want %q", providerID, provider.AdapterScript, script)
		}
		if !containsString(provider.CommandArgs, "--prompt-file") || !containsString(provider.CommandArgs, "{prompt_path}") {
			t.Fatalf("provider %s must use prompt-file adapter transport: %+v", providerID, provider.CommandArgs)
		}
		if containsString(provider.CommandArgs, "{prompt_text}") || containsString(provider.CommandArgs, "--prompt") {
			t.Fatalf("provider %s must not pass giant prompt text argv: %+v", providerID, provider.CommandArgs)
		}
		scriptPath := filepath.Join(repoRoot(t), script)
		if _, err := os.Stat(scriptPath); err != nil {
			t.Fatalf("provider %s adapter script missing at %s: %v", providerID, script, err)
		}
		scriptBytes, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatalf("read provider %s adapter script %s: %v", providerID, script, err)
		}
		if !strings.Contains(string(scriptBytes), "HOME=/Users/draccoon") {
			t.Fatalf("provider %s adapter script must pin Hermes auth HOME before CLI execution", providerID)
		}
		scriptText := string(scriptBytes)
		switch providerID {
		case "zcode_glm_5_2":
			if !strings.Contains(scriptText, "cd /Applications/ZCode.app/Contents/Resources/glm") || !strings.Contains(scriptText, "--prompt") {
				t.Fatalf("zcode adapter must execute from packaged glm cwd and use prompt mode")
			}
		case "kimi_default":
			if !strings.Contains(scriptText, "--prompt") {
				t.Fatalf("kimi adapter must use prompt mode when requesting stream-json output")
			}
		}
	}
	agy := result.Providers["antigravity_gemini"]
	if agy.SelectedModel == nil || *agy.SelectedModel != "Gemini 3.5 Flash (High)" || agy.SelectedModelRequired == nil || *agy.SelectedModelRequired {
		t.Fatalf("agy provider must pin Gemini 3.5 Flash (High) without selected_model_required: %+v", agy)
	}

	// This check validates the portable source registry contract. Disable the
	// local .kkachi toolchain overlay so developer-local resolved_argv does not
	// rewrite adapter heads during docs-contract verification.
	preflightOutput := runMARCommand(t, "provider-preflight", "--toolchain", "")
	var preflight struct {
		Attempts []struct {
			ProviderID      string   `json:"provider_id"`
			RedactedCommand []string `json:"redacted_command"`
		} `json:"attempts"`
	}
	if err := json.Unmarshal(preflightOutput, &preflight); err != nil {
		t.Fatalf("parse provider-preflight output: %v\n%s", err, preflightOutput)
	}
	adapterHeadByProvider := map[string]string{
		"zcode_glm_5_2":      "scripts/mar_adapters/mar-zcode.sh",
		"kimi_default":       "scripts/mar_adapters/mar-kimi.sh",
		"antigravity_gemini": "scripts/mar_adapters/mar-agy.sh",
	}
	for _, attempt := range preflight.Attempts {
		wantHead, ok := adapterHeadByProvider[attempt.ProviderID]
		if !ok || len(attempt.RedactedCommand) == 0 {
			t.Fatalf("unexpected preflight attempt command metadata: %+v", attempt)
		}
		if attempt.RedactedCommand[0] != wantHead {
			t.Fatalf("provider %s must execute adapter head %q, got %v", attempt.ProviderID, wantHead, attempt.RedactedCommand)
		}
	}
}

func TestMARTL002VisionTriggerSelectionIsExplicitAndFailClosed(t *testing.T) {
	defaultOutput := runMARCommand(t, "provider-preflight")
	var defaultResult struct {
		RequestedRoles []string `json:"requested_roles"`
	}
	if err := json.Unmarshal(defaultOutput, &defaultResult); err != nil {
		t.Fatalf("parse default provider-preflight output: %v\n%s", err, defaultOutput)
	}
	if containsString(defaultResult.RequestedRoles, "vision") {
		t.Fatalf("default MAR must not include optional vision role\n%s", defaultOutput)
	}

	triggeredOutput := runMARCommand(t, "provider-preflight", "--changed-surfaces", "ui,css", "--pre-scoped-evidence", "ui-change-evidence")
	var triggeredResult struct {
		RequestedRoles []string `json:"requested_roles"`
	}
	if err := json.Unmarshal(triggeredOutput, &triggeredResult); err != nil {
		t.Fatalf("parse triggered provider-preflight output: %v\n%s", err, triggeredOutput)
	}
	if !containsString(triggeredResult.RequestedRoles, "vision") {
		t.Fatalf("UI/visual trigger must include optional vision role\n%s", triggeredOutput)
	}

	paths := writeTempAttemptFixtures(t, []map[string]any{
		providerAttemptFixture("logic-primary", "logic", "zcode_glm_5_2", "primary", "PASS", nil),
		providerAttemptFixture("security-primary", "security", "antigravity_gemini", "primary", "PASS", nil),
		providerAttemptFixture("arch-primary", "arch", "zcode_glm_5_2", "primary", "PASS", nil),
		providerAttemptFixture("cve-primary", "cve", "antigravity_gemini", "primary", "PASS", nil),
		providerAttemptFixture("test-primary", "test_adequacy", "kimi_default", "primary", "PASS", nil),
	})
	args := append([]string{"merge-pack", "--provider-coverage", "--changed-surfaces", "ui", "--pre-scoped-evidence", "ui-change-evidence"}, paths...)
	mergeOutput := runMARCommand(t, args...)
	var mergeResult roleMergeResult
	if err := json.Unmarshal(mergeOutput, &mergeResult); err != nil {
		t.Fatalf("parse triggered merge-pack output: %v\n%s", err, mergeOutput)
	}
	if mergeResult.Status != "DEGRADED" || !containsString(mergeResult.UnresolvedRequiredRoles, "vision") {
		t.Fatalf("triggered vision must be required for that run and fail closed when uncovered\n%s", mergeOutput)
	}
}

type attemptReason struct {
	ProviderID            string  `json:"provider_id"`
	TerminalStatus        string  `json:"terminal_status"`
	ProviderFailureReason string  `json:"provider_failure_reason"`
	SelectedModel         *string `json:"selected_model"`
	ExitCode              *int    `json:"exit_code"`
}

type roleMergeResult struct {
	Status                  string   `json:"status"`
	Reason                  string   `json:"reason"`
	UnresolvedRequiredRoles []string `json:"unresolved_required_roles"`
	OperatorReportText      string   `json:"operator_report_text"`
	RedTriggerSummary       struct {
		RedAdjudicationRequired bool     `json:"red_adjudication_required"`
		Triggers                []string `json:"triggers"`
	} `json:"red_trigger_summary"`
	Coverage struct {
		ByRole map[string]struct {
			ProviderID        string `json:"provider_id"`
			ProviderCandidate string `json:"provider_candidate"`
			FallbackReason    string `json:"fallback_reason"`
		} `json:"by_role"`
	} `json:"coverage"`
	BlueMatrixInputs struct {
		CoveredRoles             []string                    `json:"covered_roles"`
		AcceptanceCriteriaMatrix map[string][]map[string]any `json:"acceptance_criteria_matrix"`
	} `json:"blue_matrix_inputs"`
}

func roleRegistryPayload(requiredRoles []string, providers map[string]any, roles map[string]any) map[string]any {
	return map[string]any{
		"schema_version": "mar.role_lanes.v1",
		"required_roles": requiredRoles,
		"providers":      providers,
		"roles":          roles,
	}
}

func rolePayload(primary, secondary string) map[string]any {
	return map[string]any{
		"required":                     true,
		"description":                  "test role",
		"primary_provider":             primary,
		"secondary_provider":           secondary,
		"provider_selection_rationale": "test",
	}
}

func providerPayload(commandLane, executable string, model any, args []string, validated bool) map[string]any {
	return map[string]any{
		"command_lane":    commandLane,
		"executable":      executable,
		"selected_model":  model,
		"prompt_template": "templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl",
		"command_args":    args,
		"validated":       validated,
		"validation_required_before_success_coverage": true,
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

func writeTempExecutable(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("write temp executable %s: %v", name, err)
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

func providerAttemptFixture(attemptID, roleID, providerID, candidate, terminalStatus string, extra map[string]any) map[string]any {
	attempt := map[string]any{
		"schema_version":            "mar.provider_attempt.v1",
		"run_id":                    "run-test",
		"task_id":                   "MAR-005",
		"attempt_id":                attemptID,
		"role_id":                   roleID,
		"provider_id":               providerID,
		"provider_candidate":        candidate,
		"reviewer_id":               providerID,
		"command_lane":              providerID,
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

func requireAttemptReason(t *testing.T, attempts []attemptReason, providerID, reason string) {
	t.Helper()
	for _, attempt := range attempts {
		if attempt.ProviderID == providerID {
			if attempt.TerminalStatus == "PASS" || attempt.ProviderFailureReason != reason {
				t.Fatalf("attempt %s status/reason = %q/%q, want non-PASS/%s", providerID, attempt.TerminalStatus, attempt.ProviderFailureReason, reason)
			}
			return
		}
	}
	t.Fatalf("missing attempt for provider %q in %+v", providerID, attempts)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
