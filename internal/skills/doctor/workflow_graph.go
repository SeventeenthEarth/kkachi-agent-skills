package doctor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
)

type WorkflowGraphOptions struct {
	Project string
	Runner  CommandRunner
}

type WorkflowGraphProject struct {
	Path         string `json:"path"`
	GraphPath    string `json:"graph_path"`
	GraphPresent bool   `json:"graph_present"`
}

type WorkflowGraphKAS struct {
	CLIVersion            string                     `json:"cli_version"`
	CompatibilityRegistry WorkflowGraphCompatibility `json:"compatibility_registry"`
}

type WorkflowGraphCommandEvidence struct {
	Command string `json:"command"`
	State   string `json:"state"`
}

type WorkflowGraphFlagEvidence struct {
	Flag  string `json:"flag"`
	State string `json:"state"`
}

type WorkflowGraphKAH struct {
	Available          bool                           `json:"available"`
	Version            string                         `json:"version,omitempty"`
	VersionState       string                         `json:"version_state"`
	CapabilitiesState  string                         `json:"capabilities_state"`
	GraphHelpState     string                         `json:"graph_help_state"`
	RequiredCommands   []WorkflowGraphCommandEvidence `json:"required_commands"`
	CompatibilityFlags []WorkflowGraphFlagEvidence    `json:"compatibility_flags"`
}

type WorkflowGraphEvidence struct {
	ValidateState    string `json:"validate_state,omitempty"`
	ExplainState     string `json:"explain_state,omitempty"`
	DiagnosticsState string `json:"diagnostics_state,omitempty"`
	SchemaVersion    string `json:"schema_version,omitempty"`
	SourceTemplate   string `json:"source_template,omitempty"`
	TemplateVersion  string `json:"template_version,omitempty"`
	Custom           bool   `json:"custom"`
	Checksum         string `json:"checksum,omitempty"`
}

type WorkflowGraphResult struct {
	OK          bool                   `json:"ok"`
	Command     string                 `json:"command"`
	Mode        string                 `json:"mode"`
	NoWrite     bool                   `json:"no_write"`
	Status      string                 `json:"status"`
	Project     WorkflowGraphProject   `json:"project"`
	KAS         WorkflowGraphKAS       `json:"kas"`
	KAH         WorkflowGraphKAH       `json:"kah"`
	Graph       WorkflowGraphEvidence  `json:"graph"`
	Diagnostics []discovery.Diagnostic `json:"diagnostics"`
	ReasonCodes []string               `json:"reason_codes"`
	Remediation string                 `json:"remediation"`
	NextAction  string                 `json:"next_action"`
}

func BuildWorkflowGraph(repo string, opts WorkflowGraphOptions) (WorkflowGraphResult, error) {
	if opts.Runner == nil {
		opts.Runner = defaultRunner
	}
	sourceRepoPath, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return WorkflowGraphResult{}, err
	}
	projectPath, err := filepath.Abs(opts.Project)
	if err != nil {
		projectPath = opts.Project
	}
	result := WorkflowGraphResult{
		Command: "doctor",
		Mode:    "workflow_graph_doctor",
		NoWrite: true,
		Status:  "unsupported",
		Project: WorkflowGraphProject{
			Path:      projectPath,
			GraphPath: filepath.Join(projectPath, ".kkachi-workflow.yaml"),
		},
		KAS: WorkflowGraphKAS{CLIVersion: version.CLIVersion},
		KAH: WorkflowGraphKAH{
			VersionState:      "not_run",
			CapabilitiesState: "not_run",
			GraphHelpState:    "not_run",
			RequiredCommands: []WorkflowGraphCommandEvidence{
				{Command: "graph validate --json", State: "not_run"},
				{Command: "graph explain --json", State: "not_run"},
			},
		},
		Diagnostics: []discovery.Diagnostic{},
		ReasonCodes: []string{},
	}

	compat, err := loadWorkflowGraphCompatibility(sourceRepoPath)
	result.KAS.CompatibilityRegistry = compat
	if err != nil {
		addWorkflowGraphReason(&result, "error", "compatibility_registry_unreadable", "workflow graph compatibility registry is unreadable or malformed: "+err.Error())
		return finalizeWorkflowGraph(&result, "unsupported")
	}

	if st, err := os.Stat(projectPath); err != nil {
		addWorkflowGraphReason(&result, "error", "project_unreadable", "project path is not readable: "+err.Error())
		return finalizeWorkflowGraph(&result, "unsupported")
	} else if !st.IsDir() {
		addWorkflowGraphReason(&result, "error", "project_not_directory", "project path is not a directory: "+projectPath)
		return finalizeWorkflowGraph(&result, "unsupported")
	}

	kahReady := probeWorkflowGraphKAH(&result, opts.Runner, compat)
	if !kahReady {
		return finalizeWorkflowGraph(&result, "update_kah_required")
	}

	if st, err := os.Stat(result.Project.GraphPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			addWorkflowGraphReason(&result, "warning", "graph_missing", ".kkachi-workflow.yaml is missing; read-only doctor did not initialize it.")
			return finalizeWorkflowGraph(&result, "graph_missing")
		}
		addWorkflowGraphReason(&result, "error", "graph_unreadable", ".kkachi-workflow.yaml is not readable: "+err.Error())
		return finalizeWorkflowGraph(&result, "unsupported")
	} else if st.IsDir() {
		addWorkflowGraphReason(&result, "error", "graph_not_file", ".kkachi-workflow.yaml is a directory, not a file.")
		return finalizeWorkflowGraph(&result, "unsupported")
	}
	result.Project.GraphPresent = true

	validatePayload, validateOK := runGraphJSONCommand(&result, opts.Runner, "validate", "graph", "validate", "--file", ".kkachi-workflow.yaml", "--json")
	explainPayload, explainOK := runGraphJSONCommand(&result, opts.Runner, "explain", "graph", "explain", "--file", ".kkachi-workflow.yaml", "--json")
	if validateOK && !explainOK {
		addWorkflowGraphReason(&result, "error", "graph_validate_explain_conflict", "KAH graph validate passed but graph explain failed.")
		return finalizeWorkflowGraph(&result, "graph_conflict")
	}
	if !validateOK {
		return finalizeWorkflowGraph(&result, "graph_broken")
	}
	if !explainOK {
		return finalizeWorkflowGraph(&result, "graph_conflict")
	}

	applyGraphPayload(&result, validatePayload)
	applyGraphPayload(&result, explainPayload)
	result.Graph.DiagnosticsState = "from_validate_explain"
	if result.Graph.SchemaVersion == "" {
		addWorkflowGraphReason(&result, "error", "graph_schema_missing", "KAH graph explain/validate did not report a schema version.")
		return finalizeWorkflowGraph(&result, "unsupported")
	}
	if !stringIn(result.Graph.SchemaVersion, compat.SupportedSchemaVersions) {
		if isClearlyNewerWorkflowSchema(result.Graph.SchemaVersion) {
			addWorkflowGraphReason(&result, "warning", "graph_schema_newer_than_kas", "workflow graph schema is newer than KAS 0.1.2 supports: "+result.Graph.SchemaVersion)
			return finalizeWorkflowGraph(&result, "update_kas_recommended")
		}
		addWorkflowGraphReason(&result, "error", "graph_schema_unsupported", "workflow graph schema is unsupported or ambiguous: "+result.Graph.SchemaVersion)
		return finalizeWorkflowGraph(&result, "unsupported")
	}
	if result.Graph.SourceTemplate != "" && result.Graph.SourceTemplate != "kas-default" {
		result.Graph.Custom = true
	}
	if result.Graph.SourceTemplate == "kas-default" && result.Graph.TemplateVersion != "" && result.Graph.TemplateVersion != "0.1.0" {
		if !isParseableVersion(result.Graph.TemplateVersion) {
			addWorkflowGraphReason(&result, "error", "graph_template_version_unsupported", "kas-default workflow graph template version is unsupported or ambiguous: "+result.Graph.TemplateVersion)
			return finalizeWorkflowGraph(&result, "unsupported")
		}
		if versionAtLeast(result.Graph.TemplateVersion, "0.1.0") {
			addWorkflowGraphReason(&result, "warning", "graph_template_newer_than_kas", "kas-default workflow graph template version is newer than KAS 0.1.2 supports: "+result.Graph.TemplateVersion)
			return finalizeWorkflowGraph(&result, "update_kas_recommended")
		}
		addWorkflowGraphReason(&result, "warning", "graph_template_stale", "kas-default workflow graph template version is older than the supported 0.1.0 template.")
		return finalizeWorkflowGraph(&result, "graph_stale")
	}
	if result.Graph.Custom {
		return finalizeWorkflowGraph(&result, "custom_supported")
	}
	return finalizeWorkflowGraph(&result, "pass")
}

func probeWorkflowGraphKAH(result *WorkflowGraphResult, runner CommandRunner, compat WorkflowGraphCompatibility) bool {
	versionResult := runner("", "--version")
	if versionResult.Err != nil {
		result.KAH.VersionState = "missing"
		addWorkflowGraphReason(result, "error", "kah_missing", "kkachi-agent-helper is unavailable; workflow graph doctor cannot run.")
		return false
	}
	result.KAH.Available = true
	result.KAH.Version = strings.TrimSpace(string(versionResult.Stdout))
	if result.KAH.Version == "" {
		result.KAH.VersionState = "malformed"
		addWorkflowGraphReason(result, "error", "kah_version_malformed", "kkachi-agent-helper --version returned an empty version string.")
		return false
	}
	if !versionAtLeast(result.KAH.Version, compat.KAHMinRequired) {
		result.KAH.VersionState = "too_old_or_malformed"
		addWorkflowGraphReason(result, "error", "kah_version_below_minimum", "kkachi-agent-helper must be at least "+compat.KAHMinRequired+" for workflow graph doctor.")
		return false
	}
	result.KAH.VersionState = "ok"

	capabilities := runner("", "capabilities", "--json")
	if capabilities.Err != nil {
		result.KAH.CapabilitiesState = "unavailable"
		addWorkflowGraphReason(result, "error", "kah_capabilities_unavailable", "kkachi-agent-helper capabilities --json is unavailable.")
		return false
	}
	var capabilitiesPayload map[string]any
	if err := json.Unmarshal(capabilities.Stdout, &capabilitiesPayload); err != nil {
		result.KAH.CapabilitiesState = "degraded"
		addWorkflowGraphReason(result, "error", "kah_capabilities_degraded", "kkachi-agent-helper capabilities --json did not return parseable JSON.")
		return false
	}
	result.KAH.CapabilitiesState = "ok"
	result.KAH.CompatibilityFlags = collectWorkflowGraphFlags(capabilitiesPayload)

	help := runner("", "graph", "--help")
	if help.Err != nil {
		result.KAH.GraphHelpState = "unavailable"
		addWorkflowGraphReason(result, "error", "kah_graph_help_unavailable", "kkachi-agent-helper graph --help is unavailable.")
		return false
	}
	helpText := string(help.Stdout) + "\n" + string(help.Stderr)
	validatePresent := strings.Contains(helpText, "validate")
	explainPresent := strings.Contains(helpText, "explain")
	result.KAH.RequiredCommands = []WorkflowGraphCommandEvidence{
		{Command: "graph validate --json", State: presentState(validatePresent)},
		{Command: "graph explain --json", State: presentState(explainPresent)},
	}
	if !validatePresent || !explainPresent {
		result.KAH.GraphHelpState = "missing_readonly_commands"
		addWorkflowGraphReason(result, "error", "kah_graph_readonly_commands_missing", "kkachi-agent-helper graph --help does not advertise validate and explain.")
		return false
	}
	result.KAH.GraphHelpState = "ok"
	return true
}

func runGraphJSONCommand(result *WorkflowGraphResult, runner CommandRunner, state string, args ...string) (map[string]any, bool) {
	command := strings.Join(args, " ")
	commandResult := runner(result.Project.Path, args...)
	if state == "validate" {
		result.Graph.ValidateState = commandState(commandResult)
	} else {
		result.Graph.ExplainState = commandState(commandResult)
	}
	if commandResult.Err != nil {
		addWorkflowGraphReason(result, "error", "kah_graph_"+state+"_failed", "KAH "+command+" failed.")
		return nil, false
	}
	var payload map[string]any
	if err := json.Unmarshal(commandResult.Stdout, &payload); err != nil {
		addWorkflowGraphReason(result, "error", "kah_graph_"+state+"_degraded", "KAH "+command+" did not return parseable JSON.")
		return nil, false
	}
	if ok, hasOK := payload["ok"].(bool); hasOK && !ok {
		addWorkflowGraphReason(result, "error", "kah_graph_"+state+"_not_ok", "KAH "+command+" returned ok:false.")
		return payload, false
	}
	return payload, true
}

func applyGraphPayload(result *WorkflowGraphResult, payload map[string]any) {
	for _, key := range []string{"schema_version", "version"} {
		if value := stringField(payload, key); value != "" {
			result.Graph.SchemaVersion = value
		}
	}
	if value := stringField(payload, "source_template"); value != "" {
		result.Graph.SourceTemplate = value
	}
	if value := stringField(payload, "template_version"); value != "" {
		result.Graph.TemplateVersion = value
	}
	if value := stringField(payload, "checksum"); value != "" {
		result.Graph.Checksum = value
	}
	if value := stringField(payload, "graph_checksum"); value != "" {
		result.Graph.Checksum = value
	}
	if value, ok := payload["custom"].(bool); ok {
		result.Graph.Custom = value
	}
	for _, nestedKey := range []string{"graph", "explanation", "summary"} {
		if nested, ok := payload[nestedKey].(map[string]any); ok {
			applyGraphPayload(result, nested)
		}
	}
}

func collectWorkflowGraphFlags(payload map[string]any) []WorkflowGraphFlagEvidence {
	flags := map[string]bool{}
	collectStrings(payload, flags)
	wanted := []string{
		"workflow_graph_readonly",
		"workflow_graph_diagnostics",
		"workflow_graph_no_direct_yaml_fallback",
		"workflow_graph_configurable_feedback_intake",
	}
	evidence := make([]WorkflowGraphFlagEvidence, 0, len(wanted))
	for _, flag := range wanted {
		evidence = append(evidence, WorkflowGraphFlagEvidence{Flag: flag, State: presentState(flags[flag])})
	}
	return evidence
}

func collectStrings(value any, out map[string]bool) {
	switch typed := value.(type) {
	case string:
		out[typed] = true
	case []any:
		for _, item := range typed {
			collectStrings(item, out)
		}
	case map[string]any:
		for key, item := range typed {
			out[key] = true
			collectStrings(item, out)
		}
	}
}

func addWorkflowGraphReason(result *WorkflowGraphResult, level string, code string, message string) {
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: level, Code: code, Message: message})
	result.ReasonCodes = append(result.ReasonCodes, code)
}

func finalizeWorkflowGraph(result *WorkflowGraphResult, status string) (WorkflowGraphResult, error) {
	result.Status = status
	result.OK = status == "pass" || status == "custom_supported" || status == "update_kah_recommended" || status == "update_kas_recommended"
	result.Remediation = workflowGraphStatusRemediation(status)
	result.NextAction = result.Remediation
	if result.Diagnostics == nil {
		result.Diagnostics = []discovery.Diagnostic{}
	}
	if result.ReasonCodes == nil {
		result.ReasonCodes = []string{}
	}
	return *result, nil
}

func RenderHumanWorkflowGraph(result WorkflowGraphResult) string {
	state := "error"
	if result.OK {
		state = "healthy"
	}
	lines := []string{
		fmt.Sprintf("Status: %s; workflow graph doctor status %s.", state, result.Status),
		"project: " + result.Project.Path,
		fmt.Sprintf("graph: present=%t path=%s", result.Project.GraphPresent, result.Project.GraphPath),
		"KAH: " + result.KAH.VersionState + " " + result.KAH.Version,
		"Next: " + result.NextAction,
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, fmt.Sprintf("Diagnostic[%s]: %s", diagnostic.Level, diagnostic.Message))
	}
	return strings.Join(lines, "\n")
}

func stringField(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func presentState(present bool) string {
	if present {
		return "present"
	}
	return "missing"
}

func stringIn(value string, allowed []string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

var semverPattern = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

func versionAtLeast(raw string, minimum string) bool {
	actualParts, ok := parseVersionParts(raw)
	if !ok {
		return false
	}
	minParts, ok := parseVersionParts(minimum)
	if !ok {
		return false
	}
	for index := 0; index < 3; index++ {
		if actualParts[index] > minParts[index] {
			return true
		}
		if actualParts[index] < minParts[index] {
			return false
		}
	}
	return true
}

func parseVersionParts(raw string) ([3]int, bool) {
	var parts [3]int
	match := semverPattern.FindStringSubmatch(raw)
	if len(match) != 4 {
		return parts, false
	}
	for index := 0; index < 3; index++ {
		value, err := strconv.Atoi(match[index+1])
		if err != nil {
			return parts, false
		}
		parts[index] = value
	}
	return parts, true
}

func isParseableVersion(raw string) bool {
	_, ok := parseVersionParts(raw)
	return ok
}

func isClearlyNewerWorkflowSchema(schema string) bool {
	if !strings.HasPrefix(schema, "workflow-graph/v") {
		return false
	}
	raw := strings.TrimPrefix(schema, "workflow-graph/v")
	value, err := strconv.Atoi(raw)
	return err == nil && value > 1
}
