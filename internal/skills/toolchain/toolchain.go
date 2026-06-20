package toolchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/kahrunner"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
)

const (
	Command        = "toolchain"
	SchemaVersion  = "kkachi.toolchain.v1"
	kahProbeSchema = "kah.toolchain_probe.v1"
)

var semverPattern = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+$`)

type Runner func(workDir string, args ...string) CommandResult

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type Options struct {
	ProjectRoot string
	Runner      Runner
	Now         func() time.Time
}

type Result struct {
	OK            bool         `json:"ok"`
	Command       string       `json:"command"`
	ProjectRoot   string       `json:"project_root"`
	ToolchainPath string       `json:"toolchain_path"`
	Wrote         bool         `json:"wrote"`
	Diagnostics   []Diagnostic `json:"diagnostics"`
	NextAction    string       `json:"next_action,omitempty"`
}

type Diagnostic struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
	Field   string `json:"field,omitempty"`
}

type probePayload struct {
	OK            bool   `json:"ok"`
	SchemaVersion string `json:"schema_version"`
	NoWrite       struct {
		Guaranteed bool `json:"guaranteed"`
		WriteCount int  `json:"write_count"`
	} `json:"no_write"`
	KAH struct {
		Version    string `json:"version"`
		BinaryPath string `json:"binary_path"`
	} `json:"kah"`
	Project struct {
		Root               string `json:"root"`
		KKachiDir          string `json:"kkachi_dir"`
		KKachiDirPresent   bool   `json:"kkachi_dir_present"`
		ProjectInitialized bool   `json:"project_initialized"`
	} `json:"project"`
	Doctor struct {
		Status      string   `json:"status"`
		ReasonCodes []string `json:"reason_codes"`
	} `json:"doctor"`
}

type documentPolicy struct {
	stageSelectedAt     string
	approvalEvidence    string
	stage2Activation    bool
	requiredRoles       []string
	providerToolsSchema string
}

func Init(opts Options) Result {
	opts = normalizeOptions(opts)
	result := baseResult("toolchain init", opts.ProjectRoot)
	probe, ok := runProbe(opts, &result)
	if !ok {
		return failResult(result)
	}
	policy := defaultPolicy(opts.nowString())
	data := renderDocument("init", opts, probe, policy)
	if secret := secretDiagnostic([]byte(data), result.ToolchainPath); secret != nil {
		secret.Code = "toolchain_generated_secret_detected"
		result.Diagnostics = append(result.Diagnostics, *secret)
		return failResult(result)
	}
	if err := os.MkdirAll(filepath.Join(opts.ProjectRoot, ".kkachi"), 0o755); err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kkachi_dir_create_failed", err.Error(), filepath.Join(opts.ProjectRoot, ".kkachi"), ""))
		return failResult(result)
	}
	if err := writeAtomic(filepath.Join(opts.ProjectRoot, ".kkachi", "toolchain.yaml"), []byte(data)); err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_write_failed", err.Error(), result.ToolchainPath, ""))
		return failResult(result)
	}
	result.Wrote = true
	return result
}

func Doctor(opts Options) Result {
	opts = normalizeOptions(opts)
	result := baseResult("toolchain doctor", opts.ProjectRoot)
	doc, ok := readAndValidate(result.ToolchainPath, &result)
	if !ok {
		return failResult(result)
	}
	if ok := validateLegacyKAHSelection(doc, &result); !ok {
		return failResult(result)
	}
	if _, ok := runProbe(opts, &result); !ok {
		return failResult(result)
	}
	return result
}

func Refresh(opts Options) Result {
	opts = normalizeOptions(opts)
	result := baseResult("toolchain refresh", opts.ProjectRoot)
	doc, ok := readAndValidate(result.ToolchainPath, &result)
	if !ok {
		return failResult(result)
	}
	if ok := validateLegacyKAHSelection(doc, &result); !ok {
		return failResult(result)
	}
	probe, ok := runProbe(opts, &result)
	if !ok {
		return failResult(result)
	}
	policy := policyFromDocument(doc, opts.nowString())
	data := renderDocument("refresh", opts, probe, policy)
	if secret := secretDiagnostic([]byte(data), result.ToolchainPath); secret != nil {
		secret.Code = "toolchain_generated_secret_detected"
		result.Diagnostics = append(result.Diagnostics, *secret)
		return failResult(result)
	}
	if err := writeAtomic(result.ToolchainPath, []byte(data)); err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_write_failed", err.Error(), result.ToolchainPath, ""))
		return failResult(result)
	}
	result.Wrote = true
	return result
}

func RenderHuman(result Result) string {
	status := "PASS"
	if !result.OK {
		status = "FAIL"
	}
	verb := result.Command
	if verb == "" {
		verb = Command
	}
	lines := []string{
		fmt.Sprintf("%s %s", verb, status),
		"Project root: " + result.ProjectRoot,
		"Toolchain: " + result.ToolchainPath,
	}
	if result.Wrote {
		lines = append(lines, "Write: toolchain.yaml updated")
	} else {
		lines = append(lines, "Write: none")
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, fmt.Sprintf("%s: %s", diagnostic.Code, diagnostic.Message))
	}
	if result.NextAction != "" {
		lines = append(lines, "Next: "+result.NextAction)
	}
	return strings.Join(lines, "\n")
}

func normalizeOptions(opts Options) Options {
	if opts.ProjectRoot == "" {
		if cwd, err := os.Getwd(); err == nil {
			opts.ProjectRoot = cwd
		}
	}
	abs, err := filepath.Abs(opts.ProjectRoot)
	if err == nil {
		opts.ProjectRoot = filepath.Clean(abs)
	}
	if realPath, err := filepath.EvalSymlinks(opts.ProjectRoot); err == nil {
		opts.ProjectRoot = filepath.Clean(realPath)
	}
	if opts.Runner == nil {
		opts.Runner = func(workDir string, args ...string) CommandResult {
			result := kahrunner.Runner{}.Run(workDir, args...)
			return CommandResult{Stdout: result.Stdout, Stderr: result.Stderr, Err: result.Err}
		}
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return opts
}

func (opts Options) nowString() string {
	return opts.Now().UTC().Format(time.RFC3339)
}

func baseResult(command string, projectRoot string) Result {
	return Result{
		OK:            true,
		Command:       command,
		ProjectRoot:   projectRoot,
		ToolchainPath: filepath.Join(projectRoot, ".kkachi", "toolchain.yaml"),
		Diagnostics:   []Diagnostic{},
	}
}

func failResult(result Result) Result {
	result.OK = false
	if result.NextAction == "" {
		result.NextAction = "Resolve toolchain diagnostics and rerun; missing or invalid local toolchain state fails closed."
	}
	return result
}

func runProbe(opts Options, result *Result) (probePayload, bool) {
	command := opts.Runner(opts.ProjectRoot, "project", "probe-toolchain", "--json", "--project-root", opts.ProjectRoot)
	if command.Err != nil {
		message := strings.TrimSpace(string(command.Stderr))
		if message == "" {
			message = command.Err.Error()
		}
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_unavailable", message, "", "kah"))
		return probePayload{}, false
	}
	var payload probePayload
	if err := json.Unmarshal(command.Stdout, &payload); err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_invalid_json", err.Error(), "", "kah"))
		return probePayload{}, false
	}
	if !payload.OK {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_failed", "KAH toolchain probe returned ok=false", "", "kah"))
		return probePayload{}, false
	}
	if payload.SchemaVersion != kahProbeSchema {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_schema_invalid", "KAH toolchain probe schema_version is unsupported", "", "kah.schema_version"))
		return probePayload{}, false
	}
	if !payload.NoWrite.Guaranteed || payload.NoWrite.WriteCount != 0 {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_no_write_invalid", "KAH toolchain probe did not provide zero-write evidence", "", "kah.no_write"))
		return probePayload{}, false
	}
	if strings.TrimSpace(payload.KAH.Version) == "" || strings.TrimSpace(payload.KAH.BinaryPath) == "" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_facts_missing", "KAH probe did not include required helper version and binary_path facts", "", "kah"))
		return probePayload{}, false
	}
	if !filepath.IsAbs(payload.KAH.BinaryPath) {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_facts_missing", "KAH probe binary_path must be absolute", "", "kah.binary_path"))
		return probePayload{}, false
	}
	if filepath.Clean(payload.Project.Root) != filepath.Clean(opts.ProjectRoot) {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_project_root_mismatch", "KAH probe project.root does not match requested project root", "", "project.root"))
		return probePayload{}, false
	}
	if !payload.Project.ProjectInitialized {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_project_uninitialized", "KAH probe reports project.project_initialized=false", "", "project.project_initialized"))
		return probePayload{}, false
	}
	if !isPassStatus(payload.Doctor.Status) {
		status := strings.TrimSpace(payload.Doctor.Status)
		if status == "" {
			status = "<empty>"
		}
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_doctor_not_pass", "KAH probe doctor.status is not PASS: "+status, "", "doctor.status"))
		return probePayload{}, false
	}
	for _, fact := range []struct {
		field string
		value string
	}{
		{field: "kah.version", value: payload.KAH.Version},
		{field: "kah.binary_path", value: payload.KAH.BinaryPath},
		{field: "project.root", value: payload.Project.Root},
		{field: "project.kkachi_dir", value: payload.Project.KKachiDir},
	} {
		if containsSecretLike(fact.value) {
			result.Diagnostics = append(result.Diagnostics, diag("error", "kah_probe_secret_detected", "KAH probe included a secret-like fact value", "", fact.field))
			return probePayload{}, false
		}
	}
	return payload, true
}

func readAndValidate(path string, result *Result) (map[string]string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		code := "toolchain_missing"
		if !errors.Is(err, os.ErrNotExist) {
			code = "toolchain_unreadable"
		}
		result.Diagnostics = append(result.Diagnostics, diag("error", code, err.Error(), path, ""))
		return nil, false
	}
	if secret := secretDiagnostic(data, path); secret != nil {
		result.Diagnostics = append(result.Diagnostics, *secret)
		return nil, false
	}
	doc, err := parseYAMLScalars(data)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_invalid", err.Error(), path, ""))
		return nil, false
	}
	if schema := doc["schema_version"]; schema != "" && schema != SchemaVersion {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_schema_version_invalid", "toolchain schema_version is unsupported", path, "schema_version"))
		return nil, false
	}
	if schema := doc["schema_version"]; schema == "" && (doc["kah_cli"] == "" && doc["kah_cli_path"] == "") {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_invalid", "toolchain.yaml must include schema_version or compatibility KAH selection keys", path, "schema_version"))
		return nil, false
	}
	for _, required := range []string{"generated_by", "project", "kas", "kah", "kab", "mar", "evidence_posture"} {
		if doc["schema_version"] == SchemaVersion && !hasSection(data, required) {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_required_group_missing", "toolchain.yaml is missing required group "+required, path, required))
			return nil, false
		}
	}
	if doc["schema_version"] == SchemaVersion {
		for _, required := range []string{
			"generated_by.kas_cli_version",
			"generated_by.kah_cli_version",
			"generated_by.generated_at",
			"generated_by.generator",
			"project.root",
			"kas.cli_version",
			"kah.cli_version",
			"kah.binary_path",
			"kah.project_initialized",
			"kah.doctor_status",
			"kab.adoption_stage.numeric",
			"kab.adoption_stage.canonical",
			"kab.adoption_stage.stage2_activation",
			"mar.role_policy.schema_version",
			"mar.provider_tools.schema_version",
			"evidence_posture.no_secrets",
			"evidence_posture.missing_or_invalid_fails_closed",
		} {
			if strings.TrimSpace(doc[required]) == "" {
				result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_required_field_missing", "toolchain.yaml is missing required field "+required, path, required))
				return nil, false
			}
		}
		if !strings.EqualFold(strings.TrimSpace(doc["kah.project_initialized"]), "true") {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_kah_project_uninitialized", "stored toolchain metadata reports kah.project_initialized is not true", path, "kah.project_initialized"))
			return nil, false
		}
		if !isPassStatus(doc["kah.doctor_status"]) {
			status := strings.TrimSpace(doc["kah.doctor_status"])
			if status == "" {
				status = "<empty>"
			}
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_kah_doctor_not_pass", "stored toolchain metadata reports kah.doctor_status is not PASS: "+status, path, "kah.doctor_status"))
			return nil, false
		}
		if doc["kab.adoption_stage.numeric"] != "1" || doc["kab.adoption_stage.canonical"] != "stage1_direct_codex_app_server_baseline" {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_stage_invalid", "toolchain metadata supports only Stage 1 direct Codex app-server metadata", path, "kab.adoption_stage"))
			return nil, false
		}
		if strings.EqualFold(doc["kab.adoption_stage.stage2_activation"], "true") {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_stage2_activation_unsupported", "toolchain metadata does not activate Stage 2", path, "kab.adoption_stage.stage2_activation"))
			return nil, false
		}
		if doc["evidence_posture.no_secrets"] != "true" || doc["evidence_posture.missing_or_invalid_fails_closed"] != "true" {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_evidence_posture_invalid", "toolchain evidence_posture must preserve no_secrets and fail-closed assertions", path, "evidence_posture"))
			return nil, false
		}
	}
	return doc, true
}

func validateLegacyKAHSelection(doc map[string]string, result *Result) bool {
	rawVersion, hasVersion := doc["kah_cli"]
	rawPath := strings.TrimSpace(doc["kah_cli_path"])
	if hasVersion && strings.TrimSpace(rawVersion) == "" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_cli_invalid", "top-level kah_cli must be an exact semver token when present", result.ToolchainPath, "kah_cli"))
		return false
	}
	if hasVersion && !semverPattern.MatchString(strings.TrimSpace(rawVersion)) {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kah_cli_invalid", "top-level kah_cli must be an exact semver token when present", result.ToolchainPath, "kah_cli"))
		return false
	}
	if rawPath != "" {
		if !filepath.IsAbs(rawPath) {
			result.Diagnostics = append(result.Diagnostics, diag("error", "kah_cli_path_invalid", "top-level kah_cli_path must be absolute when present", result.ToolchainPath, "kah_cli_path"))
			return false
		}
		st, err := os.Stat(rawPath)
		if err != nil || st.IsDir() || st.Mode().Perm()&0111 == 0 {
			result.Diagnostics = append(result.Diagnostics, diag("error", "kah_cli_path_invalid", "top-level kah_cli_path must point to an executable file when present", result.ToolchainPath, "kah_cli_path"))
			return false
		}
	}
	return true
}

func renderDocument(action string, opts Options, probe probePayload, policy documentPolicy) string {
	info := version.Current()
	commit := info.GitCommit
	if commit == "" {
		commit = "unknown"
	}
	kahBinaryPath := probe.KAH.BinaryPath
	selectionSource := selectionSource(opts.ProjectRoot, kahBinaryPath)
	projectID := filepath.Base(opts.ProjectRoot)
	lines := []string{
		`schema_version: "` + SchemaVersion + `"`,
		"generated_by:",
		`  kas_cli_version: "` + version.CLIVersion + `"`,
		`  kah_cli_version: "` + probe.KAH.Version + `"`,
		`  generated_at: "` + opts.nowString() + `"`,
		`  generator: "kkachi-agent-skills toolchain ` + action + `"`,
		"project:",
		`  id: "` + yamlEscape(projectID) + `"`,
		`  root: "` + yamlEscape(opts.ProjectRoot) + `"`,
		`  kkachi_dir: ".kkachi"`,
		"kas:",
		`  cli_version: "` + version.CLIVersion + `"`,
		"  source:",
		`    module: "` + yamlEscape(info.ModulePath) + `"`,
		`    commit: "` + yamlEscape(commit) + `"`,
		`    dirty: ` + strconv.FormatBool(info.Dirty),
		"kah:",
		`  cli_version: "` + yamlEscape(probe.KAH.Version) + `"`,
		`  binary_path: "` + yamlEscape(kahBinaryPath) + `"`,
		`  selection_source: "` + selectionSource + `"`,
		`  project_initialized: ` + strconv.FormatBool(probe.Project.ProjectInitialized),
		`  doctor_status: "` + yamlEscape(statusOrUnknown(probe.Doctor.Status)) + `"`,
		"kab:",
		"  adoption_stage:",
		"    numeric: 1",
		`    canonical: "stage1_direct_codex_app_server_baseline"`,
		`    selected_at: "` + yamlEscape(policy.stageSelectedAt) + `"`,
		`    approval_evidence: "` + yamlEscape(policy.approvalEvidence) + `"`,
		`    stage2_activation: ` + strconv.FormatBool(policy.stage2Activation),
		"mar:",
		"  role_policy:",
		`    schema_version: "mar.role_lanes.v1"`,
		`    required_roles: [` + quoteList(policy.requiredRoles) + `]`,
		"  provider_tools:",
		`    schema_version: "` + yamlEscape(policy.providerToolsSchema) + `"`,
		"    providers: {}",
		"evidence_posture:",
		"  generated_local_toolchain_state: true",
		"  no_secrets: true",
		"  not_runtime_session_state: true",
		"  missing_or_invalid_fails_closed: true",
	}
	return strings.Join(lines, "\n") + "\n"
}

func defaultPolicy(now string) documentPolicy {
	return documentPolicy{
		stageSelectedAt:     now,
		approvalEvidence:    "not_applicable",
		stage2Activation:    false,
		requiredRoles:       []string{"logic", "security", "arch", "cve", "test_adequacy"},
		providerToolsSchema: "mar.provider_tools.v1",
	}
}

func policyFromDocument(doc map[string]string, fallbackNow string) documentPolicy {
	policy := defaultPolicy(fallbackNow)
	if value := doc["kab.adoption_stage.selected_at"]; value != "" {
		policy.stageSelectedAt = value
	}
	if value := doc["kab.adoption_stage.approval_evidence"]; value != "" {
		policy.approvalEvidence = value
	}
	if value := doc["kab.adoption_stage.stage2_activation"]; value != "" {
		policy.stage2Activation = strings.EqualFold(value, "true")
	}
	if value := doc["mar.role_policy.required_roles"]; value != "" {
		policy.requiredRoles = splitInlineList(value)
	}
	if len(policy.requiredRoles) == 0 {
		policy.requiredRoles = defaultPolicy(fallbackNow).requiredRoles
	}
	if value := doc["mar.provider_tools.schema_version"]; value != "" {
		policy.providerToolsSchema = value
	}
	return policy
}

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".toolchain.yaml.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	if dirHandle, err := os.Open(dir); err == nil {
		_ = dirHandle.Sync()
		_ = dirHandle.Close()
	}
	return nil
}

func parseYAMLScalars(data []byte) (map[string]string, error) {
	values := map[string]string{}
	var stack []string
	lines := bytes.Split(data, []byte("\n"))
	for lineNo, raw := range lines {
		if len(bytes.TrimSpace(raw)) == 0 {
			continue
		}
		line := string(raw)
		if strings.TrimSpace(line) == "---" {
			continue
		}
		if strings.Contains(line, "\t") {
			return nil, fmt.Errorf("line %d uses tabs", lineNo+1)
		}
		indent := countLeadingSpaces(line)
		if indent%2 != 0 {
			return nil, fmt.Errorf("line %d has unsupported indentation", lineNo+1)
		}
		level := indent / 2
		if level > len(stack) {
			return nil, fmt.Errorf("line %d skips indentation level", lineNo+1)
		}
		stack = stack[:level]
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "- ") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("line %d is not a key/value entry", lineNo+1)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(stripComment(value))
		path := strings.Join(append(stack, key), ".")
		if value == "" {
			stack = append(stack, key)
			continue
		}
		if strings.HasPrefix(value, "[") && !strings.HasSuffix(value, "]") {
			return nil, fmt.Errorf("line %d has invalid inline list", lineNo+1)
		}
		if strings.Count(value, `"`)%2 != 0 || strings.Count(value, `'`)%2 != 0 {
			return nil, fmt.Errorf("line %d has unbalanced quotes", lineNo+1)
		}
		values[path] = strings.Trim(strings.TrimSpace(value), `"'`)
		if level == 0 {
			values[key] = values[path]
		}
	}
	return values, nil
}

func hasSection(data []byte, section string) bool {
	for _, raw := range bytes.Split(data, []byte("\n")) {
		if string(raw) == section+":" {
			return true
		}
	}
	return false
}

func stripComment(value string) string {
	inSingle := false
	inDouble := false
	for i, r := range value {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return strings.TrimSpace(value[:i])
			}
		}
	}
	return value
}

func countLeadingSpaces(value string) int {
	count := 0
	for _, r := range value {
		if r != ' ' {
			return count
		}
		count++
	}
	return count
}

func secretDiagnostic(data []byte, path string) *Diagnostic {
	for lineNo, raw := range bytes.Split(data, []byte("\n")) {
		line := strings.TrimSpace(string(raw))
		lower := strings.ToLower(line)
		for _, marker := range []string{"api_key", "apikey", "auth_token", "bearer", "password", "provider_cookie", "gateway_credential", "ssh_key", "private_key", "access_token", "refresh_token"} {
			if strings.Contains(lower, marker) {
				diag := diag("error", "toolchain_secret_detected", fmt.Sprintf("secret-like key or value found on line %d", lineNo+1), path, "")
				return &diag
			}
		}
		for _, marker := range []string{"sk-", "ghp_", "xoxb-", "-----begin private key-----"} {
			if strings.Contains(lower, marker) {
				diag := diag("error", "toolchain_secret_detected", fmt.Sprintf("secret-like token value found on line %d", lineNo+1), path, "")
				return &diag
			}
		}
	}
	return nil
}

func containsSecretLike(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{"api_key", "apikey", "auth_token", "bearer", "password", "provider_cookie", "gateway_credential", "ssh_key", "private_key", "access_token", "refresh_token", "sk-", "ghp_", "xoxb-", "-----begin private key-----"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func selectionSource(projectRoot string, kahBinaryPath string) string {
	if os.Getenv("KKACHI_KAH_BIN") != "" {
		return "KKACHI_KAH_BIN"
	}
	if strings.TrimSpace(kahBinaryPath) != "" && filepath.Clean(kahBinaryPath) == filepath.Join(projectRoot, ".kkachi", "bin", "kkachi-agent-helper") {
		return "repo_bin"
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".kkachi", "toolchain.yaml")); err == nil {
		return "toolchain"
	}
	return "PATH"
}

func statusOrUnknown(value string) string {
	if isPassStatus(value) {
		return "PASS"
	}
	return "UNKNOWN"
}

func isPassStatus(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "PASS")
}

func quoteList(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, `"`+yamlEscape(value)+`"`)
	}
	return strings.Join(parts, ", ")
}

func splitInlineList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.Trim(strings.TrimSpace(part), `"'`)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func yamlEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}

func diag(level, code, message, path, field string) Diagnostic {
	return Diagnostic{Level: level, Code: code, Message: message, Path: path, Field: field}
}

func SortedDiagnosticCodes(result Result) []string {
	codes := make([]string, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		codes = append(codes, diagnostic.Code)
	}
	sort.Strings(codes)
	return codes
}
