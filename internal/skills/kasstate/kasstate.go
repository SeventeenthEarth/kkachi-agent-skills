package kasstate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/install"
)

const (
	Command         = "sync-project-kas"
	CLIVersion      = install.CLIVersion
	SchemaVersion   = "0.1"
	stage3Canonical = "stage3_kab_backend_selected"
)

type Options struct {
	Profile          string
	Project          string
	StatePath        string
	LegacyMarkerPath string
	DryRun           bool
}

type ReadSurface struct {
	State   string `json:"state"`
	Path    string `json:"path"`
	SHA256  string `json:"sha256,omitempty"`
	Message string `json:"message,omitempty"`
}

type ReadSurfaces struct {
	YAML         ReadSurface `json:"yaml"`
	LegacyMarker ReadSurface `json:"legacy_marker"`
}

type EffectiveStageClaim struct {
	Numeric                  int    `json:"numeric"`
	Canonical                string `json:"canonical"`
	Source                   string `json:"source"`
	KABExecutionClaimAllowed bool   `json:"kab_execution_claim_allowed"`
	FailClosedToStage1       bool   `json:"fail_closed_to_stage1"`
}

type Validation struct {
	SchemaVersion     string                 `json:"schema_version,omitempty"`
	PackBaselineCount int                    `json:"pack_baseline_count"`
	Diagnostics       []discovery.Diagnostic `json:"diagnostics"`
}

type Result struct {
	OK                           bool                `json:"ok"`
	Command                      string              `json:"command"`
	Mode                         string              `json:"mode"`
	CLIVersion                   string              `json:"cli_version"`
	DryRun                       bool                `json:"dry_run"`
	TargetProfile                string              `json:"target_profile"`
	ProjectID                    string              `json:"project_id"`
	YAMLStatePath                string              `json:"yaml_state_path"`
	LegacyMarkerPath             string              `json:"legacy_marker_path"`
	StateSource                  string              `json:"state_source"`
	ReadSurfaces                 ReadSurfaces        `json:"read_surfaces"`
	EffectiveStageClaim          EffectiveStageClaim `json:"effective_stage_claim"`
	WriteTargetAfterApprovedSync string              `json:"write_target_after_approved_sync"`
	Validation                   Validation          `json:"validation"`
	NextAction                   string              `json:"next_action"`
}

type stateFile struct {
	version         string
	project         map[string]string
	stage           map[string]string
	upstream        map[string]string
	packBaselines   []map[string]string
	overlayPolicy   map[string]string
	updatePolicy    map[string][]string
	updateScalars   map[string]string
	evidencePosture map[string]string
}

var (
	shaPattern      = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	checksumPattern = regexp.MustCompile(`^sha256:[0-9a-fA-F]{64}$`)
)

func Build(opts Options) Result {
	legacyPath := opts.LegacyMarkerPath
	if legacyPath == "" && opts.StatePath != "" {
		legacyPath = filepath.Join(filepath.Dir(opts.StatePath), filepath.Base(install.KABAdoptionMarkerRelativePath))
	}
	result := Result{
		OK:                           false,
		Command:                      Command,
		Mode:                         "state_validate",
		CLIVersion:                   CLIVersion,
		DryRun:                       opts.DryRun,
		TargetProfile:                opts.Profile,
		ProjectID:                    opts.Project,
		YAMLStatePath:                opts.StatePath,
		LegacyMarkerPath:             legacyPath,
		StateSource:                  "fail_closed",
		WriteTargetAfterApprovedSync: "yaml_state_path",
		EffectiveStageClaim: EffectiveStageClaim{
			Numeric:                  1,
			Canonical:                install.KABStage1Canonical,
			Source:                   "fail_closed",
			KABExecutionClaimAllowed: false,
			FailClosedToStage1:       true,
		},
		Validation: Validation{Diagnostics: []discovery.Diagnostic{}},
	}

	result.ReadSurfaces.LegacyMarker = readLegacyMarker(legacyPath)
	result.ReadSurfaces.YAML = ReadSurface{State: "missing", Path: opts.StatePath}

	if opts.StatePath == "" {
		addDiag(&result, "error", "state_path_required", "sync-project-kas requires --state <path>.")
		result.NextAction = "Provide the project kas-project-state.yaml path and rerun with --dry-run."
		return result
	}
	data, err := os.ReadFile(opts.StatePath)
	if err != nil {
		result.ReadSurfaces.YAML.State = "missing"
		if !os.IsNotExist(err) {
			result.ReadSurfaces.YAML.State = "unreadable"
		}
		result.ReadSurfaces.YAML.Message = err.Error()
		addDiag(&result, "error", "state_file_missing", "kas-project-state.yaml is missing or unreadable; legacy marker read is reporting-only and cannot validate YAML state.")
		result.StateSource = stateSourceForInvalidYAML(result.ReadSurfaces.LegacyMarker.State)
		result.NextAction = "Create or repair kas-project-state.yaml before any project KAS sync; until then only Stage 1 static claims are allowed."
		return result
	}
	sum := sha256.Sum256(data)
	result.ReadSurfaces.YAML = ReadSurface{State: "read", Path: opts.StatePath, SHA256: hex.EncodeToString(sum[:])}

	parsed, parseDiagnostics := parseYAMLSubset(string(data))
	for _, diagnostic := range parseDiagnostics {
		result.Validation.Diagnostics = append(result.Validation.Diagnostics, diagnostic)
	}
	if len(parseDiagnostics) == 0 {
		validateState(&result, parsed, opts)
	}
	if hasErrors(result.Validation.Diagnostics) {
		result.ReadSurfaces.YAML.State = "invalid"
		result.StateSource = stateSourceForInvalidYAML(result.ReadSurfaces.LegacyMarker.State)
		result.NextAction = "Repair kas-project-state.yaml diagnostics before project KAS sync; legacy marker compatibility cannot upgrade invalid YAML."
		return result
	}

	result.OK = true
	result.ReadSurfaces.YAML.State = "valid"
	result.StateSource = "yaml"
	result.Validation.SchemaVersion = parsed.version
	result.Validation.PackBaselineCount = len(parsed.packBaselines)
	result.EffectiveStageClaim = EffectiveStageClaim{
		Numeric:                  mustAtoi(parsed.stage["numeric"]),
		Canonical:                parsed.stage["canonical"],
		Source:                   "yaml",
		KABExecutionClaimAllowed: false,
		FailClosedToStage1:       false,
	}
	result.NextAction = "State is valid for KASUPD-003 dry-run classification; no files were written."
	return result
}

func RenderHuman(result Result) string {
	status := "검증 실패"
	if result.OK {
		status = "검증 완료"
	}
	lines := []string{
		fmt.Sprintf("상태: %s — project %s / profile %s KAS state dry-run read.", status, result.ProjectID, result.TargetProfile),
		fmt.Sprintf("YAML: %s (%s)", result.ReadSurfaces.YAML.State, result.YAMLStatePath),
		fmt.Sprintf("legacy marker: %s (%s)", result.ReadSurfaces.LegacyMarker.State, result.LegacyMarkerPath),
		fmt.Sprintf("effective stage: %d %s (source %s, KAB execution claim allowed: %t)", result.EffectiveStageClaim.Numeric, result.EffectiveStageClaim.Canonical, result.EffectiveStageClaim.Source, result.EffectiveStageClaim.KABExecutionClaimAllowed),
		fmt.Sprintf("write target after approved sync: %s", result.WriteTargetAfterApprovedSync),
	}
	for _, diagnostic := range result.Validation.Diagnostics {
		lines = append(lines, fmt.Sprintf("%s: %s — %s", diagnostic.Level, diagnostic.Code, diagnostic.Message))
	}
	lines = append(lines, "다음: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func readLegacyMarker(path string) ReadSurface {
	surface := ReadSurface{State: "not_requested", Path: path}
	if path == "" {
		return surface
	}
	data, err := os.ReadFile(path)
	if err != nil {
		surface.State = "missing"
		if !os.IsNotExist(err) {
			surface.State = "unreadable"
		}
		surface.Message = err.Error()
		return surface
	}
	sum := sha256.Sum256(data)
	surface.SHA256 = hex.EncodeToString(sum[:])
	if _, ok := install.ParseKABAdoptionStageMarker(data); ok {
		surface.State = "present_compatible"
	} else {
		surface.State = "present_invalid"
	}
	return surface
}

func stateSourceForInvalidYAML(legacyState string) string {
	if legacyState == "present_compatible" {
		return "legacy_marker_only"
	}
	return "fail_closed"
}

func parseYAMLSubset(data string) (stateFile, []discovery.Diagnostic) {
	state := stateFile{
		project:         map[string]string{},
		stage:           map[string]string{},
		upstream:        map[string]string{},
		overlayPolicy:   map[string]string{},
		updatePolicy:    map[string][]string{},
		updateScalars:   map[string]string{},
		evidencePosture: map[string]string{},
	}
	diagnostics := []discovery.Diagnostic{}
	section := ""
	listKey := ""
	currentPack := -1
	lines := strings.Split(data, "\n")
	for i, raw := range lines {
		lineNo := i + 1
		if strings.TrimSpace(raw) == "" || strings.HasPrefix(strings.TrimSpace(raw), "#") {
			continue
		}
		if strings.Contains(raw, "\t") {
			diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d uses tabs; only two-space indentation is supported", lineNo)))
			continue
		}
		if strings.Contains(raw, "&") || strings.Contains(raw, "*") || strings.Contains(raw, "|") || strings.Contains(raw, ">") {
			diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d uses unsupported YAML features; only the documented scalar/list subset is supported", lineNo)))
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		if indent%2 != 0 {
			diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d has unsupported indentation", lineNo)))
			continue
		}
		line := strings.TrimSpace(raw)
		if indent == 0 {
			section = ""
			listKey = ""
			currentPack = -1
			key, value, ok := splitYAMLKeyValue(line)
			if !ok {
				diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d is not a key/value entry", lineNo)))
				continue
			}
			if value == "" {
				section = key
			} else if key == "version" {
				state.version = unquote(value)
			}
			continue
		}
		switch section {
		case "project":
			parseScalarMapLine(&diagnostics, state.project, line, lineNo)
		case "kab_adoption_stage":
			parseScalarMapLine(&diagnostics, state.stage, line, lineNo)
		case "upstream_kas":
			parseScalarMapLine(&diagnostics, state.upstream, line, lineNo)
		case "overlay_policy":
			parseScalarMapLine(&diagnostics, state.overlayPolicy, line, lineNo)
		case "evidence_posture":
			parseScalarMapLine(&diagnostics, state.evidencePosture, line, lineNo)
		case "pack_baselines":
			if indent == 2 && strings.HasPrefix(line, "- ") {
				key, value, ok := splitYAMLKeyValue(strings.TrimSpace(strings.TrimPrefix(line, "- ")))
				if !ok || value == "" {
					diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d has unsupported pack_baselines list syntax", lineNo)))
					continue
				}
				state.packBaselines = append(state.packBaselines, map[string]string{key: unquote(value)})
				currentPack = len(state.packBaselines) - 1
			} else if indent == 4 && currentPack >= 0 {
				parseScalarMapLine(&diagnostics, state.packBaselines[currentPack], line, lineNo)
			} else {
				diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d has unsupported pack_baselines indentation", lineNo)))
			}
		case "update_policy":
			if indent == 2 {
				key, value, ok := splitYAMLKeyValue(line)
				if !ok {
					diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d is not a key/value entry", lineNo)))
					continue
				}
				listKey = ""
				if value == "" {
					listKey = key
					if state.updatePolicy[listKey] == nil {
						state.updatePolicy[listKey] = []string{}
					}
				} else {
					state.updateScalars[key] = unquote(value)
				}
			} else if indent == 4 && strings.HasPrefix(line, "- ") && listKey != "" {
				state.updatePolicy[listKey] = append(state.updatePolicy[listKey], unquote(strings.TrimSpace(strings.TrimPrefix(line, "- "))))
			} else {
				diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d has unsupported update_policy syntax", lineNo)))
			}
		default:
			diagnostics = append(diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d is under unsupported section %q", lineNo, section)))
		}
	}
	if hasSuspiciousInput(data) {
		diagnostics = append(diagnostics, diag("error", "auth_token_gateway_or_provider_mutation_detected", "state input contains secret/auth/token/gateway/provider/runtime-state-like material"))
	}
	return state, diagnostics
}

func parseScalarMapLine(diagnostics *[]discovery.Diagnostic, target map[string]string, line string, lineNo int) {
	key, value, ok := splitYAMLKeyValue(line)
	if !ok || value == "" {
		*diagnostics = append(*diagnostics, diag("error", "state_schema_invalid", fmt.Sprintf("line %d is not a supported scalar key/value entry", lineNo)))
		return
	}
	target[key] = unquote(value)
}

func splitYAMLKeyValue(line string) (string, string, bool) {
	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" || strings.Contains(key, " ") {
		return "", "", false
	}
	return key, value, true
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func validateState(result *Result, state stateFile, opts Options) {
	if state.version != SchemaVersion {
		addDiag(result, "error", "state_schema_invalid", fmt.Sprintf("version must be %q", SchemaVersion))
	}
	requireMap(result, "project", state.project, []string{"id", "repo", "kas_suite", "profile"})
	requireMap(result, "kab_adoption_stage", state.stage, []string{"numeric", "canonical", "selection_source", "selected_at", "approval_evidence", "stage2_activation"})
	requireMap(result, "upstream_kas", state.upstream, []string{"repo", "remote", "commit", "dirty", "synced_at", "sync_task"})
	requireMap(result, "overlay_policy", state.overlayPolicy, []string{"local_overlay_allowed", "preserve_project_authority", "preserve_project_roadmap_ids", "preserve_project_test_commands", "preserve_role_labels", "overwrite_mode"})
	requireMap(result, "evidence_posture", state.evidencePosture, []string{"not_kab_runtime_evidence", "not_stage2_activation_by_itself", "missing_or_unreadable_fails_to_stage1_claims"})
	if opts.Profile != "" && state.project["profile"] != "" && opts.Profile != state.project["profile"] {
		addDiag(result, "error", "profile_mismatch", fmt.Sprintf("--profile %q does not match state project.profile %q", opts.Profile, state.project["profile"]))
	}
	if opts.Project != "" && state.project["id"] != "" && opts.Project != state.project["id"] {
		addDiag(result, "error", "project_mismatch", fmt.Sprintf("--project %q does not match state project.id %q", opts.Project, state.project["id"]))
	}
	validateStage(result, state.stage)
	validateUpstream(result, state.upstream)
	validatePackBaselines(result, state.packBaselines)
	validateOverlay(result, state.overlayPolicy)
	validateUpdatePolicy(result, state.updatePolicy, state.updateScalars)
	validateEvidencePosture(result, state.evidencePosture)
}

func validateStage(result *Result, stage map[string]string) {
	numeric := stage["numeric"]
	canonical := stage["canonical"]
	switch numeric {
	case "1":
		if canonical != install.KABStage1Canonical {
			addDiag(result, "error", "stage_schema_invalid", "Stage 1 numeric value must use the Stage 1 canonical value")
		}
	case "2":
		if canonical != install.KABStage2Canonical {
			addDiag(result, "error", "stage_schema_invalid", "Stage 2 numeric value must use the Stage 2 canonical value")
		}
	case "3":
		addDiag(result, "error", "stage_unsupported", "Stage 3 is reserved and unsupported for project KAS sync state")
	default:
		addDiag(result, "error", "stage_schema_invalid", "kab_adoption_stage.numeric must be 1 or 2")
	}
	if canonical == stage3Canonical {
		addDiag(result, "error", "stage_unsupported", "Stage 3 canonical value is reserved and unsupported")
	}
	if stage["stage2_activation"] != "false" {
		addDiag(result, "error", "stage2_activation_rejected", "stage2_activation must be false; YAML state is not Stage 2 activation by itself")
	}
}

func validateUpstream(result *Result, upstream map[string]string) {
	if commit := upstream["commit"]; commit == "" || !shaPattern.MatchString(commit) {
		addDiag(result, "error", "upstream_commit_unknown", "upstream_kas.commit must be a full 40-character git SHA")
	}
	if dirty := upstream["dirty"]; dirty != "true" && dirty != "false" {
		addDiag(result, "error", "state_schema_invalid", "upstream_kas.dirty must be a boolean")
	}
}

func validatePackBaselines(result *Result, baselines []map[string]string) {
	if len(baselines) == 0 {
		addDiag(result, "error", "checksum_mismatch_without_baseline", "pack_baselines must include at least one checksum baseline")
		return
	}
	required := []string{"upstream_pack", "project_skill", "source_checksum", "project_checksum", "merge_mode"}
	for i, baseline := range baselines {
		for _, key := range required {
			if strings.TrimSpace(baseline[key]) == "" {
				addDiag(result, "error", "state_schema_invalid", fmt.Sprintf("pack_baselines[%d].%s is required", i, key))
			}
		}
		if baseline["source_checksum"] != "" && !checksumPattern.MatchString(baseline["source_checksum"]) {
			addDiag(result, "error", "checksum_mismatch_without_baseline", fmt.Sprintf("pack_baselines[%d].source_checksum must be sha256:<64 hex>", i))
		}
		if baseline["project_checksum"] != "" && !checksumPattern.MatchString(baseline["project_checksum"]) {
			addDiag(result, "error", "checksum_mismatch_without_baseline", fmt.Sprintf("pack_baselines[%d].project_checksum must be sha256:<64 hex>", i))
		}
		if baseline["merge_mode"] != "" && baseline["merge_mode"] != "semantic_port" {
			addDiag(result, "error", "state_schema_invalid", fmt.Sprintf("pack_baselines[%d].merge_mode must be semantic_port for KASUPD-002", i))
		}
	}
}

func validateOverlay(result *Result, overlay map[string]string) {
	for _, key := range []string{"preserve_project_authority", "preserve_project_roadmap_ids", "preserve_project_test_commands", "preserve_role_labels"} {
		if overlay[key] != "true" {
			addDiag(result, "error", "overlay_policy_invalid", key+" must be true")
		}
	}
	if overlay["local_overlay_allowed"] != "true" && overlay["local_overlay_allowed"] != "false" {
		addDiag(result, "error", "overlay_policy_invalid", "local_overlay_allowed must be a boolean")
	}
	if overlay["overwrite_mode"] != "never_without_review" {
		addDiag(result, "error", "overlay_policy_invalid", "overwrite_mode must be never_without_review")
	}
}

func validateUpdatePolicy(result *Result, lists map[string][]string, scalars map[string]string) {
	if scalars["default_mode"] != "dry_run_then_semantic_merge" {
		addDiag(result, "error", "state_schema_invalid", "update_policy.default_mode must be dry_run_then_semantic_merge")
	}
	requiredFailClosed := []string{
		"state_file_missing",
		"state_schema_invalid",
		"stage_unsupported",
		"upstream_commit_unknown",
		"checksum_mismatch_without_baseline",
		"auth_token_gateway_or_provider_mutation_detected",
	}
	have := map[string]bool{}
	for _, value := range lists["fail_closed_when"] {
		have[value] = true
	}
	for _, value := range requiredFailClosed {
		if !have[value] {
			addDiag(result, "error", "state_schema_invalid", "update_policy.fail_closed_when must include "+value)
		}
	}
	for _, key := range []string{"auto_apply_when", "require_llm_merge_when", "fail_closed_when"} {
		if len(lists[key]) == 0 {
			addDiag(result, "error", "state_schema_invalid", "update_policy."+key+" must be non-empty")
		}
	}
}

func validateEvidencePosture(result *Result, posture map[string]string) {
	for _, key := range []string{"not_kab_runtime_evidence", "not_stage2_activation_by_itself", "missing_or_unreadable_fails_to_stage1_claims"} {
		if posture[key] != "true" {
			addDiag(result, "error", "evidence_posture_invalid", key+" must be true")
		}
	}
}

func requireMap(result *Result, section string, values map[string]string, keys []string) {
	for _, key := range keys {
		if strings.TrimSpace(values[key]) == "" {
			addDiag(result, "error", "state_schema_invalid", section+"."+key+" is required")
		}
	}
}

func hasSuspiciousInput(data string) bool {
	allowed := map[string]bool{
		"auth_token_gateway_or_provider_mutation_detected": true,
	}
	suspicious := []string{"token", "secret", "credential", "provider_key", "api_key", "password", "gateway", "session_id", "session_state", "runtime_state"}
	lines := strings.Split(data, "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.FieldsFunc(strings.ToLower(line), func(r rune) bool {
			return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_'
		})
		for _, part := range parts {
			if allowed[part] {
				continue
			}
			for _, word := range suspicious {
				if part == word || strings.Contains(part, word) {
					return true
				}
			}
		}
	}
	return false
}

func addDiag(result *Result, level string, code string, message string) {
	result.Validation.Diagnostics = append(result.Validation.Diagnostics, diag(level, code, message))
}

func diag(level string, code string, message string) discovery.Diagnostic {
	return discovery.Diagnostic{Level: level, Code: code, Message: message}
}

func hasErrors(diagnostics []discovery.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == "error" {
			return true
		}
	}
	return false
}

func mustAtoi(value string) int {
	n, _ := strconv.Atoi(value)
	return n
}
