package graphsync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/doctor"
)

const (
	candidateTemplatePath = "templates/workflow-graphs/kas-default.yaml"
	candidateDir          = ".kkachi/graph/candidates"
)

type Options struct {
	Repo     string
	Project  string
	Reason   string
	Proposal string
	Approval string
	Runner   doctor.CommandRunner
}

type Project struct {
	Path      string `json:"path"`
	GraphPath string `json:"graph_path"`
}

type Candidate struct {
	Path     string `json:"path,omitempty"`
	Checksum string `json:"checksum,omitempty"`
	Source   string `json:"source,omitempty"`
}

type SemanticDiff struct {
	State   string         `json:"state"`
	Command string         `json:"command,omitempty"`
	Summary string         `json:"summary,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

type Proposal struct {
	ID               string         `json:"id,omitempty"`
	Path             string         `json:"path,omitempty"`
	ApprovalRequired bool           `json:"approval_required"`
	Payload          map[string]any `json:"payload,omitempty"`
}

type ApplyEvidence struct {
	ProposalID    string         `json:"proposal_id,omitempty"`
	ApprovalRef   string         `json:"approval_ref,omitempty"`
	AuditEventIDs []string       `json:"audit_event_ids,omitempty"`
	AuditPath     string         `json:"audit_path,omitempty"`
	BackupPath    string         `json:"backup_path,omitempty"`
	RecoveryPath  string         `json:"recovery_path,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
}

type PostApplyEvidence struct {
	ValidationState string         `json:"validation_state"`
	ExplainState    string         `json:"explain_state"`
	GraphChecksum   string         `json:"graph_checksum,omitempty"`
	GraphVersion    string         `json:"graph_version,omitempty"`
	ValidatePayload map[string]any `json:"validate_payload,omitempty"`
	ExplainPayload  map[string]any `json:"explain_payload,omitempty"`
}

type Result struct {
	OK               bool                       `json:"ok"`
	Command          string                     `json:"command"`
	Mode             string                     `json:"mode"`
	Status           string                     `json:"status"`
	Project          Project                    `json:"project"`
	Doctor           doctor.WorkflowGraphResult `json:"doctor"`
	Candidate        Candidate                  `json:"candidate,omitempty"`
	SemanticDiff     SemanticDiff               `json:"semantic_diff,omitempty"`
	Proposal         Proposal                   `json:"proposal,omitempty"`
	Apply            ApplyEvidence              `json:"apply,omitempty"`
	PostApply        PostApplyEvidence          `json:"post_apply,omitempty"`
	RiskFlags        []string                   `json:"risk_flags"`
	ReasonCodes      []string                   `json:"reason_codes"`
	Diagnostics      []discovery.Diagnostic     `json:"diagnostics"`
	NextCommand      string                     `json:"next_command,omitempty"`
	NextAction       string                     `json:"next_action"`
	DirectGraphWrite bool                       `json:"direct_graph_write"`
}

type commandRunner struct{}

func (commandRunner) Run(workDir string, args ...string) doctor.CommandResult {
	cmd := exec.Command("kkachi-agent-helper", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return doctor.CommandResult{Stdout: out, Stderr: exitErr.Stderr, Err: err}
		}
		return doctor.CommandResult{Stdout: out, Err: err}
	}
	return doctor.CommandResult{Stdout: out}
}

func Propose(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	result, err := baseResult("workflow_graph_repair_propose", opts)
	if err != nil {
		return result, err
	}
	if strings.TrimSpace(opts.Reason) == "" {
		return fail(result, "reason_required", "repair --workflow-graph --propose requires --reason <reason>."), nil
	}
	if !proposalAllowed(result.Doctor.Status) {
		return fail(result, "workflow_graph_proposal_refused", "workflow graph proposal refused for doctor status "+result.Doctor.Status), nil
	}
	if err := requireFullGraphEnvelope(&result, opts.Runner); err != nil {
		return fail(result, "kah_graph_repair_envelope_unavailable", err.Error()), nil
	}

	candidate, err := writeCandidate(opts.Repo, opts.Project)
	if err != nil {
		return fail(result, "candidate_graph_write_failed", err.Error()), nil
	}
	result.Candidate = candidate
	if result.Doctor.Status == "graph_stale" {
		diff := runJSON(opts.Runner, result.Project.Path, "graph", "diff", "--from", ".kkachi-workflow.yaml", "--to", candidate.Path, "--semantic", "--json")
		result.SemanticDiff = normalizeDiff(diff, candidate.Path)
		if diff.Err != nil {
			return fail(result, "kah_graph_diff_failed", "KAH graph diff failed for valid stale base."), nil
		}
	} else {
		result.SemanticDiff = SemanticDiff{
			State:   "not_applicable_missing_or_invalid_base",
			Summary: "Semantic diff was not run because the base graph is missing or invalid.",
		}
	}

	propose := runJSON(opts.Runner, result.Project.Path, "graph", "propose", "--candidate-file", candidate.Path, "--reason", opts.Reason, "--json")
	if propose.Err != nil {
		return fail(result, "kah_graph_propose_failed", "KAH graph propose failed."), nil
	}
	result.Proposal = normalizeProposal(propose.Payload)
	result.RiskFlags = appendUnique(result.RiskFlags, collectStringValues(propose.Payload, "risk_flags")...)
	result.ReasonCodes = appendUnique(result.ReasonCodes, collectStringValues(propose.Payload, "reason_codes")...)
	result.Status = "proposal_available"
	result.OK = true
	if result.Proposal.ID != "" {
		result.NextCommand = fmt.Sprintf("kkachi-agent-skills repair --project %s --workflow-graph --apply-proposal %s --approval <approval-ref> --json", shellQuote(result.Project.Path), result.Proposal.ID)
	} else {
		result.NextCommand = "kkachi-agent-skills repair --project <project-path> --workflow-graph --apply-proposal <proposal-id> --approval <approval-ref> --json"
	}
	result.NextAction = "Review proposal and semantic diff evidence; apply requires explicit approval evidence."
	return result, nil
}

func Apply(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	if strings.TrimSpace(opts.Proposal) == "" {
		return fail(newResult("workflow_graph_repair_apply", opts), "proposal_required", "repair --workflow-graph --apply-proposal requires a proposal id."), nil
	}
	if strings.TrimSpace(opts.Approval) == "" {
		return fail(newResult("workflow_graph_repair_apply", opts), "approval_required", "repair --workflow-graph --apply-proposal requires --approval <approval-ref>."), nil
	}
	result, err := baseResult("workflow_graph_repair_apply", opts)
	if err != nil {
		return result, err
	}
	if !applyAllowed(result.Doctor.Status) {
		return fail(result, "workflow_graph_apply_refused", "workflow graph apply refused for doctor status "+result.Doctor.Status), nil
	}
	if err := requireFullGraphEnvelope(&result, opts.Runner); err != nil {
		return fail(result, "kah_graph_repair_envelope_unavailable", err.Error()), nil
	}

	apply := runJSON(opts.Runner, result.Project.Path, "graph", "apply", "--proposal", opts.Proposal, "--approval", opts.Approval, "--json")
	if apply.Err != nil {
		return fail(result, "kah_graph_apply_failed", "KAH graph apply failed."), nil
	}
	result.Apply = normalizeApply(opts.Proposal, opts.Approval, apply.Payload)

	validate := runJSON(opts.Runner, result.Project.Path, "graph", "validate", "--file", ".kkachi-workflow.yaml", "--json")
	explain := runJSON(opts.Runner, result.Project.Path, "graph", "explain", "--file", ".kkachi-workflow.yaml", "--json")
	result.PostApply = normalizePostApply(validate, explain)
	if validate.Err != nil || explain.Err != nil {
		return fail(result, "kah_graph_post_apply_validation_failed", "KAH graph validate/explain failed after apply."), nil
	}
	result.Status = "applied"
	result.OK = true
	result.NextAction = "Graph apply completed through KAH; preserve apply, validate, explain, checksum, and audit evidence in graph-evidence.md when this affects an active run."
	return result, nil
}

func normalizeOptions(opts Options) Options {
	if opts.Runner == nil {
		opts.Runner = commandRunner{}.Run
	}
	return opts
}

func baseResult(mode string, opts Options) (Result, error) {
	result := newResult(mode, opts)
	doctorResult, err := doctor.BuildWorkflowGraph(opts.Repo, doctor.WorkflowGraphOptions{Project: opts.Project, Runner: opts.Runner})
	result.Doctor = doctorResult
	result.ReasonCodes = appendUnique(result.ReasonCodes, doctorResult.ReasonCodes...)
	result.Diagnostics = append(result.Diagnostics, doctorResult.Diagnostics...)
	if err != nil {
		return result, err
	}
	return result, nil
}

func newResult(mode string, opts Options) Result {
	projectPath, err := filepath.Abs(opts.Project)
	if err != nil {
		projectPath = opts.Project
	}
	return Result{
		Command:     "repair",
		Mode:        mode,
		Status:      "unsupported",
		Project:     Project{Path: projectPath, GraphPath: filepath.Join(projectPath, ".kkachi-workflow.yaml")},
		RiskFlags:   []string{},
		ReasonCodes: []string{},
		Diagnostics: []discovery.Diagnostic{},
	}
}

func proposalAllowed(status string) bool {
	return status == "graph_missing" || status == "graph_stale" || status == "graph_broken"
}

func applyAllowed(status string) bool {
	switch status {
	case "update_kah_required", "update_kah_recommended", "update_kas_recommended", "graph_conflict", "unsupported", "custom_supported":
		return false
	default:
		return true
	}
}

func requireFullGraphEnvelope(result *Result, runner doctor.CommandRunner) error {
	capabilities := runner("", "capabilities", "--json")
	if capabilities.Err != nil {
		return errors.New("kkachi-agent-helper capabilities --json is unavailable")
	}
	capabilityText := string(capabilities.Stdout)
	for _, want := range []string{"workflow_graph_no_direct_yaml_fallback", "workflow_graph_apply"} {
		if !strings.Contains(capabilityText, want) {
			return fmt.Errorf("kkachi-agent-helper capabilities --json missing %s", want)
		}
	}
	help := runner("", "graph", "--help")
	if help.Err != nil {
		return errors.New("kkachi-agent-helper graph --help is unavailable")
	}
	helpText := string(help.Stdout) + "\n" + string(help.Stderr)
	for _, want := range []string{"diff", "propose", "apply"} {
		if !strings.Contains(helpText, want) {
			return fmt.Errorf("kkachi-agent-helper graph --help does not advertise %s", want)
		}
	}
	return nil
}

func writeCandidate(repo string, project string) (Candidate, error) {
	sourceRepo, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return Candidate{}, err
	}
	data, err := os.ReadFile(filepath.Join(sourceRepo, candidateTemplatePath))
	if err != nil {
		return Candidate{}, err
	}
	sum := sha256.Sum256(data)
	checksum := "sha256:" + hex.EncodeToString(sum[:])
	rel := filepath.ToSlash(filepath.Join(candidateDir, "kas-default-"+hex.EncodeToString(sum[:8])+".yaml"))
	projectPath, err := filepath.Abs(project)
	if err != nil {
		projectPath = project
	}
	target := filepath.Join(projectPath, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return Candidate{}, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".kas-graph-candidate-*")
	if err != nil {
		return Candidate{}, err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return Candidate{}, err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return Candidate{}, err
	}
	if err := os.Rename(tmpPath, target); err != nil {
		_ = os.Remove(tmpPath)
		return Candidate{}, err
	}
	return Candidate{Path: rel, Checksum: checksum, Source: candidateTemplatePath}, nil
}

type jsonRun struct {
	Payload map[string]any
	Err     error
	Command string
}

func runJSON(runner doctor.CommandRunner, workDir string, args ...string) jsonRun {
	result := runner(workDir, args...)
	out := jsonRun{Err: result.Err, Command: strings.Join(args, " ")}
	if len(result.Stdout) == 0 {
		return out
	}
	var payload map[string]any
	if err := json.Unmarshal(result.Stdout, &payload); err != nil {
		out.Err = err
		return out
	}
	out.Payload = payload
	if ok, hasOK := payload["ok"].(bool); hasOK && !ok {
		out.Err = errors.New("json ok:false")
	}
	return out
}

func normalizeDiff(run jsonRun, candidate string) SemanticDiff {
	diff := SemanticDiff{State: "failed", Command: run.Command, Payload: run.Payload}
	if run.Err == nil {
		diff.State = "completed"
	}
	if summary := firstString(run.Payload, "summary", "semantic_diff_summary", "diff_summary", "status"); summary != "" {
		diff.Summary = summary
	} else if run.Err == nil {
		diff.Summary = "KAH semantic diff completed for " + candidate + "."
	}
	return diff
}

func normalizeProposal(payload map[string]any) Proposal {
	proposal := Proposal{
		ID:               firstString(payload, "proposal_id", "id"),
		Path:             firstString(payload, "proposal_path", "path"),
		ApprovalRequired: true,
		Payload:          payload,
	}
	if value, ok := findBool(payload, "approval_required"); ok {
		proposal.ApprovalRequired = value
	}
	return proposal
}

func normalizeApply(proposalID string, approvalRef string, payload map[string]any) ApplyEvidence {
	return ApplyEvidence{
		ProposalID:    firstNonEmpty(firstString(payload, "proposal_id", "id"), proposalID),
		ApprovalRef:   firstNonEmpty(firstString(payload, "approval_ref", "approval_evidence_ref", "approval"), approvalRef),
		AuditEventIDs: collectStringValues(payload, "audit_event_ids", "kah_graph_audit_event_ids", "event_ids"),
		AuditPath:     firstString(payload, "audit_path", "audit_evidence_path"),
		BackupPath:    firstString(payload, "backup_path", "backup_evidence_path"),
		RecoveryPath:  firstString(payload, "recovery_path", "recovery_evidence_path"),
		Payload:       payload,
	}
}

func normalizePostApply(validate jsonRun, explain jsonRun) PostApplyEvidence {
	post := PostApplyEvidence{
		ValidationState: commandState(validate),
		ExplainState:    commandState(explain),
		ValidatePayload: validate.Payload,
		ExplainPayload:  explain.Payload,
	}
	for _, payload := range []map[string]any{validate.Payload, explain.Payload} {
		if post.GraphChecksum == "" {
			post.GraphChecksum = firstString(payload, "graph_checksum", "checksum")
		}
		if post.GraphVersion == "" {
			post.GraphVersion = firstString(payload, "graph_version", "schema_version", "version")
		}
	}
	return post
}

func commandState(run jsonRun) string {
	if run.Err != nil {
		return "failed"
	}
	return "ok"
}

func fail(result Result, code string, message string) Result {
	result.OK = false
	if result.Status == "" || result.Status == "unsupported" {
		result.Status = "blocked_for_approval"
	}
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: code, Message: message})
	result.ReasonCodes = appendUnique(result.ReasonCodes, code)
	result.NextAction = message
	return result
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := findString(payload, key); ok {
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func findString(value any, key string) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		if raw, ok := typed[key]; ok {
			if text, ok := raw.(string); ok {
				return text, true
			}
		}
		keys := make([]string, 0, len(typed))
		for childKey := range typed {
			keys = append(keys, childKey)
		}
		sort.Strings(keys)
		for _, childKey := range keys {
			if text, ok := findString(typed[childKey], key); ok {
				return text, true
			}
		}
	case []any:
		for _, item := range typed {
			if text, ok := findString(item, key); ok {
				return text, true
			}
		}
	}
	return "", false
}

func findBool(value any, key string) (bool, bool) {
	switch typed := value.(type) {
	case map[string]any:
		if raw, ok := typed[key]; ok {
			if flag, ok := raw.(bool); ok {
				return flag, true
			}
		}
		keys := make([]string, 0, len(typed))
		for childKey := range typed {
			keys = append(keys, childKey)
		}
		sort.Strings(keys)
		for _, childKey := range keys {
			if flag, ok := findBool(typed[childKey], key); ok {
				return flag, true
			}
		}
	case []any:
		for _, item := range typed {
			if flag, ok := findBool(item, key); ok {
				return flag, true
			}
		}
	}
	return false, false
}

func collectStringValues(payload map[string]any, keys ...string) []string {
	wanted := map[string]bool{}
	for _, key := range keys {
		wanted[key] = true
	}
	values := []string{}
	var walk func(any, string)
	walk = func(value any, parent string) {
		switch typed := value.(type) {
		case string:
			if wanted[parent] {
				values = append(values, typed)
			}
		case []any:
			for _, item := range typed {
				walk(item, parent)
			}
		case map[string]any:
			for key, item := range typed {
				walk(item, key)
			}
		}
	}
	walk(payload, "")
	return appendUnique(nil, values...)
}

func appendUnique(base []string, values ...string) []string {
	seen := map[string]bool{}
	for _, value := range base {
		seen[value] = true
	}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		base = append(base, value)
		seen[value] = true
	}
	return base
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if strings.ContainsAny(value, " \t\n'\"\\$`") {
		return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
	}
	return value
}

func RenderHuman(result Result) string {
	state := "blocked"
	if result.OK {
		state = "ok"
	}
	lines := []string{
		fmt.Sprintf("Status: %s; workflow graph %s status %s.", state, strings.TrimPrefix(result.Mode, "workflow_graph_repair_"), result.Status),
		"Project: " + result.Project.Path,
		"Doctor status: " + result.Doctor.Status,
	}
	if result.Candidate.Path != "" {
		lines = append(lines, "Candidate: "+result.Candidate.Path)
	}
	if result.Proposal.ID != "" {
		lines = append(lines, "Proposal: "+result.Proposal.ID)
	}
	if result.Apply.ProposalID != "" {
		lines = append(lines, "Applied proposal: "+result.Apply.ProposalID)
	}
	if result.NextCommand != "" {
		lines = append(lines, "Next command: "+result.NextCommand)
	}
	lines = append(lines, "Next: "+result.NextAction)
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, fmt.Sprintf("Diagnostic[%s]: %s", diagnostic.Level, diagnostic.Message))
	}
	return strings.Join(lines, "\n")
}
