package toolchain

import (
	"bytes"
	"embed"
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

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/install"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/kahrunner"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
)

const (
	Command        = "toolchain"
	SchemaVersion  = "kkachi.toolchain.v1"
	kahProbeSchema = "kah.toolchain_probe.v1"
)

//go:embed templates/kkachi-agent-toolchain.py.tmpl
var launcherTemplates embed.FS

var semverPattern = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+$`)

type Runner func(workDir string, args ...string) CommandResult

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type Options struct {
	ProjectRoot      string
	Profile          string
	Project          string
	ProfileRoot      string
	Stage            string
	ApprovalEvidence string
	LauncherBinDir   string
	Runner           Runner
	Now              func() time.Time
}

type Result struct {
	OK            bool             `json:"ok"`
	Command       string           `json:"command"`
	ProjectRoot   string           `json:"project_root"`
	ToolchainPath string           `json:"toolchain_path,omitempty"`
	Wrote         bool             `json:"wrote"`
	Launchers     []LauncherRecord `json:"launchers,omitempty"`
	Diagnostics   []Diagnostic     `json:"diagnostics"`
	NextAction    string           `json:"next_action,omitempty"`
}

type LauncherRecord struct {
	Name string `json:"name"`
	Path string `json:"path"`
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
	stageNumeric        int
	stageCanonical      string
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

func ImportLegacy(opts Options) Result {
	opts = normalizeOptions(opts)
	result := baseResult("toolchain import-legacy", opts.ProjectRoot)
	if strings.TrimSpace(opts.Profile) == "" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "profile_required", "toolchain import-legacy requires --profile <profile>", "", "profile"))
		return failResult(result)
	}
	if strings.TrimSpace(opts.Project) == "" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "project_required", "toolchain import-legacy requires --project <id>", "", "project"))
		return failResult(result)
	}
	if _, err := os.Stat(result.ToolchainPath); err == nil {
		if _, ok := readAndValidate(result.ToolchainPath, &result); ok {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_import_existing_toolchain_conflict", "existing .kkachi/toolchain.yaml is present; legacy import will not overwrite it", result.ToolchainPath, ""))
		}
		return failResult(result)
	} else if !errors.Is(err, os.ErrNotExist) {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_unreadable", err.Error(), result.ToolchainPath, ""))
		return failResult(result)
	}
	legacy, ok := readLegacyState(opts, &result)
	if !ok {
		return failResult(result)
	}
	probe, ok := runProbe(opts, &result)
	if !ok {
		return failResult(result)
	}
	policy := defaultPolicy(opts.nowString())
	policy.stageNumeric = legacy.stageNumeric
	policy.stageCanonical = legacy.stageCanonical
	policy.stageSelectedAt = legacy.selectedAt
	policy.approvalEvidence = legacy.approvalEvidence
	policy.stage2Activation = legacy.stage2Activation
	data := renderDocument("import-legacy", opts, probe, policy)
	if secret := secretDiagnostic([]byte(data), result.ToolchainPath); secret != nil {
		secret.Code = "toolchain_generated_secret_detected"
		result.Diagnostics = append(result.Diagnostics, *secret)
		return failResult(result)
	}
	if err := os.MkdirAll(filepath.Join(opts.ProjectRoot, ".kkachi"), 0o755); err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "kkachi_dir_create_failed", err.Error(), filepath.Join(opts.ProjectRoot, ".kkachi"), ""))
		return failResult(result)
	}
	if err := writeAtomic(result.ToolchainPath, []byte(data)); err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_write_failed", err.Error(), result.ToolchainPath, ""))
		return failResult(result)
	}
	result.Wrote = true
	return result
}

func SetStage(opts Options) Result {
	opts = normalizeOptions(opts)
	result := baseResult("toolchain set-stage", opts.ProjectRoot)
	doc, ok := readAndValidate(result.ToolchainPath, &result)
	if !ok {
		return failResult(result)
	}
	if strings.TrimSpace(opts.Stage) == "" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "stage_required", "toolchain set-stage requires --stage <stage>", "", "kab.adoption_stage"))
		return failResult(result)
	}
	stage, ok := resolveStage(opts.Stage, &result)
	if !ok {
		return failResult(result)
	}
	if strings.TrimSpace(opts.ApprovalEvidence) == "" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "stage_approval_evidence_required", "toolchain set-stage requires --approval-evidence <ref>", "", "kab.adoption_stage.approval_evidence"))
		return failResult(result)
	}
	if containsSecretLike(opts.ApprovalEvidence) {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_secret_detected", "approval evidence contains a secret-like value", "", "kab.adoption_stage.approval_evidence"))
		return failResult(result)
	}
	if stage.Numeric == 3 {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_stage3_reserved", "KAB Stage 3 is reserved and unauthorized for toolchain set-stage", result.ToolchainPath, "kab.adoption_stage"))
		return failResult(result)
	}
	if stage.Numeric == 2 {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_stage2_capability_proof_missing", "Stage 2 requires deterministic KAB native_codex capability proof; this repository does not expose that proof surface", result.ToolchainPath, "kab.adoption_stage"))
		return failResult(result)
	}
	probe, ok := runProbe(opts, &result)
	if !ok {
		return failResult(result)
	}
	policy := policyFromDocument(doc, opts.nowString())
	policy.stageNumeric = stage.Numeric
	policy.stageCanonical = stage.Canonical
	policy.stageSelectedAt = opts.nowString()
	policy.approvalEvidence = opts.ApprovalEvidence
	policy.stage2Activation = false
	data := renderDocument("set-stage", opts, probe, policy)
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

func InstallLaunchers(opts Options) Result {
	opts = normalizeOptions(opts)
	result := baseResult("toolchain install-launchers", opts.ProjectRoot)
	result.ToolchainPath = ""
	binDir := strings.TrimSpace(opts.LauncherBinDir)
	if binDir == "" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "launcher_bin_dir_required", "toolchain install-launchers requires --bin-dir <path>", "", "bin_dir"))
		return failResult(result)
	}
	absBinDir, err := filepath.Abs(binDir)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "launcher_bin_dir_invalid", err.Error(), binDir, "bin_dir"))
		return failResult(result)
	}
	absBinDir = filepath.Clean(absBinDir)
	if err := os.MkdirAll(absBinDir, 0o755); err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "launcher_bin_dir_create_failed", err.Error(), absBinDir, "bin_dir"))
		return failResult(result)
	}
	templateData, err := launcherTemplates.ReadFile("templates/kkachi-agent-toolchain.py.tmpl")
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "launcher_template_missing", err.Error(), "templates/kkachi-agent-toolchain.py.tmpl", ""))
		return failResult(result)
	}
	for _, launcher := range []struct {
		kind string
		name string
	}{
		{kind: "kas", name: "kkachi-agent-skills-toolchain"},
		{kind: "kah", name: "kkachi-agent-helper-toolchain"},
	} {
		content := strings.ReplaceAll(string(templateData), `{{KIND}}`, launcher.kind)
		path := filepath.Join(absBinDir, launcher.name)
		if err := writeExecutableAtomic(path, []byte(content)); err != nil {
			result.Diagnostics = append(result.Diagnostics, diag("error", "launcher_write_failed", err.Error(), path, ""))
			return failResult(result)
		}
		result.Launchers = append(result.Launchers, LauncherRecord{Name: launcher.name, Path: path})
	}
	result.Wrote = true
	result.NextAction = "Run kkachi-agent-skills-toolchain --toolchain-status from a project with kkachi.toolchain.v1 metadata."
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
		if !validateMARProviderTools(doc, path, result) {
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
	if strings.TrimSpace(opts.Project) != "" {
		projectID = strings.TrimSpace(opts.Project)
	}
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
		"    numeric: " + strconv.Itoa(policy.stageNumeric),
		`    canonical: "` + yamlEscape(policy.stageCanonical) + `"`,
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
		stageNumeric:        1,
		stageCanonical:      install.KABStage1Canonical,
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
	if value := doc["kab.adoption_stage.numeric"]; value != "" {
		if numeric, err := strconv.Atoi(value); err == nil {
			policy.stageNumeric = numeric
		}
	}
	if value := doc["kab.adoption_stage.canonical"]; value != "" {
		policy.stageCanonical = value
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

type legacyState struct {
	stageNumeric     int
	stageCanonical   string
	selectedAt       string
	approvalEvidence string
	stage2Activation bool
}

func readLegacyState(opts Options, result *Result) (legacyState, bool) {
	statePath := legacyStatePath(opts)
	stateData, err := os.ReadFile(statePath)
	if err != nil {
		code := "legacy_state_missing"
		if !errors.Is(err, os.ErrNotExist) {
			code = "legacy_state_unreadable"
		}
		result.Diagnostics = append(result.Diagnostics, diag("error", code, err.Error(), statePath, ""))
		return legacyState{}, false
	}
	if secret := secretDiagnostic(stateData, statePath); secret != nil {
		result.Diagnostics = append(result.Diagnostics, *secret)
		return legacyState{}, false
	}
	stateDoc, err := parseYAMLScalars(stateData)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "legacy_state_invalid", err.Error(), statePath, ""))
		return legacyState{}, false
	}
	if stateDoc["version"] != "0.1" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "legacy_state_invalid", "legacy kas-project-state.yaml version is unsupported", statePath, "version"))
		return legacyState{}, false
	}
	if stateDoc["project.id"] != opts.Project || stateDoc["project.kas_suite"] != opts.Project || stateDoc["project.profile"] != opts.Profile {
		result.Diagnostics = append(result.Diagnostics, diag("error", "legacy_state_conflict", "legacy project/profile facts do not match explicit import inputs", statePath, "project"))
		return legacyState{}, false
	}
	stage, ok := resolveStage(stateDoc["kab_adoption_stage.numeric"], result)
	if !ok {
		result.Diagnostics[len(result.Diagnostics)-1].Code = "legacy_state_invalid"
		result.Diagnostics[len(result.Diagnostics)-1].Path = statePath
		result.Diagnostics[len(result.Diagnostics)-1].Field = "kab_adoption_stage.numeric"
		return legacyState{}, false
	}
	if canonical := stateDoc["kab_adoption_stage.canonical"]; canonical == "" || canonical != stage.Canonical {
		result.Diagnostics = append(result.Diagnostics, diag("error", "legacy_state_conflict", "legacy KAB stage numeric/canonical facts conflict", statePath, "kab_adoption_stage"))
		return legacyState{}, false
	}
	legacy := legacyState{
		stageNumeric:     stage.Numeric,
		stageCanonical:   stage.Canonical,
		selectedAt:       stateDoc["kab_adoption_stage.selected_at"],
		approvalEvidence: stateDoc["kab_adoption_stage.approval_evidence"],
		stage2Activation: strings.EqualFold(stateDoc["kab_adoption_stage.stage2_activation"], "true"),
	}
	if legacy.selectedAt == "" || legacy.approvalEvidence == "" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "legacy_state_invalid", "legacy KAB stage selected_at and approval_evidence are required", statePath, "kab_adoption_stage"))
		return legacyState{}, false
	}
	if legacy.stageNumeric != 1 || legacy.stage2Activation {
		result.Diagnostics = append(result.Diagnostics, diag("error", "legacy_state_conflict", "legacy import only accepts non-activated Stage 1 metadata in Stage 1 direct implementation", statePath, "kab_adoption_stage"))
		return legacyState{}, false
	}
	markerPath := filepath.Join(filepath.Dir(statePath), "kab-adoption-stage.md")
	markerData, err := os.ReadFile(markerPath)
	if err != nil {
		code := "legacy_stage_marker_missing"
		if !errors.Is(err, os.ErrNotExist) {
			code = "legacy_stage_marker_unreadable"
		}
		result.Diagnostics = append(result.Diagnostics, diag("error", code, err.Error(), markerPath, ""))
		return legacyState{}, false
	}
	if secret := secretDiagnostic(markerData, markerPath); secret != nil {
		result.Diagnostics = append(result.Diagnostics, *secret)
		return legacyState{}, false
	}
	markerStage, parsed := install.ParseKABAdoptionStageMarker(markerData)
	if !parsed {
		result.Diagnostics = append(result.Diagnostics, diag("error", "legacy_stage_marker_invalid", "legacy kab-adoption-stage.md is missing parseable stage facts", markerPath, "kab_adoption_stage"))
		return legacyState{}, false
	}
	if markerStage.Numeric != legacy.stageNumeric || markerStage.Canonical != legacy.stageCanonical {
		result.Diagnostics = append(result.Diagnostics, diag("error", "legacy_stage_marker_conflict", "legacy marker stage conflicts with kas-project-state.yaml", markerPath, "kab_adoption_stage"))
		return legacyState{}, false
	}
	return legacy, true
}

func legacyStatePath(opts Options) string {
	root := strings.TrimSpace(opts.ProfileRoot)
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		root = filepath.Join(home, ".hermes", "profiles", opts.Profile)
	}
	return filepath.Join(root, "skills", opts.Project, opts.Project+"-kas", "references", "kas-project-state.yaml")
}

func resolveStage(raw string, result *Result) (install.KABAdoptionStage, bool) {
	raw = strings.TrimSpace(raw)
	canonical := ""
	numeric := raw
	switch raw {
	case install.KABStage1Canonical, install.KABStage2Canonical, "stage3_kab_backend_selected":
		canonical = raw
		numeric = ""
	}
	if raw == "3" || raw == "stage3_kab_backend_selected" {
		return install.KABAdoptionStage{Applicable: true, Numeric: 3, Canonical: "stage3_kab_backend_selected"}, true
	}
	stage, err := install.ResolveKABAdoptionStage(install.StageSelectionInput{Numeric: numeric, Canonical: canonical, Source: "toolchain_set_stage"})
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_stage_invalid", err.Error(), result.ToolchainPath, "kab.adoption_stage"))
		return install.KABAdoptionStage{}, false
	}
	return stage, true
}

func validateMARProviderTools(doc map[string]string, path string, result *Result) bool {
	if doc["mar.provider_tools.schema_version"] != "mar.provider_tools.v1" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_mar_provider_tools_invalid", "mar.provider_tools.schema_version must be mar.provider_tools.v1", path, "mar.provider_tools.schema_version"))
		return false
	}
	if value := doc["mar.provider_tools.providers"]; value != "" && value != "{}" {
		result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_mar_provider_tools_invalid", "mar.provider_tools.providers must be a mapping, not a scalar", path, "mar.provider_tools.providers"))
		return false
	}
	allowed := map[string]bool{
		"resolved_argv": true, "command_lane": true, "selected_model": true, "selected_model_required": true,
		"model_selection": true, "model_selection_note": true, "validated": true, "version": true,
		"reason": true, "validation_evidence": true, "adapter_proof_evidence": true,
	}
	for key, value := range doc {
		const prefix = "mar.provider_tools.providers."
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		rest := strings.TrimPrefix(key, prefix)
		parts := strings.Split(rest, ".")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || !allowed[parts[1]] {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_mar_provider_tools_invalid", "mar.provider_tools contains an unsupported provider proof field", path, key))
			return false
		}
		if containsSecretLike(value) {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_secret_detected", "MAR provider proof contains a secret-like value", path, key))
			return false
		}
		if (parts[1] == "validated" || parts[1] == "selected_model_required") && value != "true" && value != "false" {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_mar_provider_tools_invalid", "MAR provider proof boolean fields must be true or false", path, key))
			return false
		}
		if parts[1] == "resolved_argv" && len(splitInlineList(value)) == 0 {
			result.Diagnostics = append(result.Diagnostics, diag("error", "toolchain_mar_provider_tools_invalid", "MAR provider resolved_argv must be a non-empty inline string list", path, key))
			return false
		}
	}
	return true
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

func writeExecutableAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".kkachi-launcher.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
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
	if err := os.Chmod(path, 0o755); err != nil {
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
