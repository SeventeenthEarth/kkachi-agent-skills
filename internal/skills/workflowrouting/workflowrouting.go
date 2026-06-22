package workflowrouting

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowregistry"
)

const (
	Command = "workflow-route"
	Mode    = "classification_bundle_route"
)

type Options struct {
	TaxonomyPath          string
	SelectorRegistryPath  string
	TaskClass             string
	ClassificationReason  string
	SelectedSpine         string
	ProjectHasTealLane    *bool
	UIUXChange            *bool
	TealSkipReason        string
	TealWaiverApproved    bool
	TealWaiverApprovalRef string
	TealWaiverScope       string
	TealWaiverExpiresAt   string
	Labels                []string
	ChangedSurfaces       []string
	Risk                  string
	RequiredAgents        []string
	RequiredCapabilities  []string
}

type Result struct {
	OK                   bool              `json:"ok"`
	Command              string            `json:"command"`
	Mode                 string            `json:"mode"`
	Status               string            `json:"status"`
	TaskClass            string            `json:"task_class,omitempty"`
	InputTaskClass       string            `json:"input_task_class,omitempty"`
	ClassificationReason string            `json:"classification_reason,omitempty"`
	SelectedBundle       string            `json:"selected_bundle,omitempty"`
	WorkflowID           string            `json:"workflow_id,omitempty"`
	WorkflowPath         string            `json:"workflow_path,omitempty"`
	WorkPath             string            `json:"work_path,omitempty"`
	WorkMode             string            `json:"work_mode,omitempty"`
	ExecutionMode        string            `json:"execution_mode,omitempty"`
	SelectedSpine        string            `json:"selected_spine,omitempty"`
	RequiredPhases       []string          `json:"required_phases,omitempty"`
	SkippedPhaseReasons  map[string]string `json:"skipped_phase_reasons,omitempty"`
	CapabilityPosture    CapabilityPosture `json:"capability_posture"`
	RequiredCapabilities []string          `json:"required_capabilities,omitempty"`
	Labels               []string          `json:"labels,omitempty"`
	ChangedSurfaces      []string          `json:"changed_surfaces,omitempty"`
	SelectorMatch        SelectorEvidence  `json:"selector_match,omitempty"`
	TealApplicability    TealApplicability `json:"teal_applicability,omitempty"`
	Taxonomy             SourceEvidence    `json:"taxonomy"`
	SelectorRegistry     SourceEvidence    `json:"selector_registry"`
	Diagnostics          []Diagnostic      `json:"diagnostics"`
	ReasonCodes          []string          `json:"reason_codes"`
	NextAction           string            `json:"next_action"`
	DirectKAHStateWrite  bool              `json:"direct_kah_state_write"`
}

type CapabilityPosture struct {
	KAH       string `json:"kah,omitempty"`
	KAB       string `json:"kab,omitempty"`
	Kanban    string `json:"kanban,omitempty"`
	CodeGraph string `json:"codegraph,omitempty"`
}

type SourceEvidence struct {
	Path     string `json:"path,omitempty"`
	Version  string `json:"version,omitempty"`
	Checksum string `json:"checksum,omitempty"`
}

type SelectorEvidence struct {
	Status       string   `json:"status,omitempty"`
	WorkflowID   string   `json:"workflow_id,omitempty"`
	CandidateIDs []string `json:"candidate_ids,omitempty"`
}

type TealApplicability struct {
	ContractVersion             string   `json:"contract_version,omitempty"`
	ProjectHasTealLane          bool     `json:"project_has_teal_lane"`
	UIUXChange                  bool     `json:"ui_ux_change"`
	TealRequired                bool     `json:"teal_required"`
	Derivation                  string   `json:"derivation,omitempty"`
	TealSkipReason              string   `json:"teal_skip_reason,omitempty"`
	TealWaiverApproved          bool     `json:"teal_waiver_approved"`
	TealWaiverApprovalRef       string   `json:"teal_waiver_approval_ref,omitempty"`
	TealWaiverScope             string   `json:"teal_waiver_scope,omitempty"`
	TealWaiverExpiresAt         string   `json:"teal_waiver_expires_at,omitempty"`
	RequiredWhenTealRequired    []string `json:"required_when_teal_required,omitempty"`
	MissingRequiredStatus       string   `json:"missing_required_status,omitempty"`
	OrdinaryReviewIsSubstitute  bool     `json:"ordinary_review_is_substitute"`
	MARReviewIsSubstitute       bool     `json:"mar_review_is_substitute"`
	BackendEvidenceIsSubstitute bool     `json:"backend_evidence_is_substitute"`
	HelperNotesAreSubstitute    bool     `json:"helper_notes_are_substitute"`
}

type Diagnostic struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type Taxonomy struct {
	Version  string
	Path     string
	Checksum string
	Aliases  map[string]string
	Classes  map[string]TaskClass
}

type TaskClass struct {
	Name             string
	WorkPath         string
	WorkMode         string
	ExecutionMode    string
	DefaultSpine     string
	KAHDefault       string
	KABDefault       string
	KanbanDefault    string
	CodeGraphDefault string
	RequiredPhases   []string
	SkippedByDefault []string
}

func Route(opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	result := newResult(opts)
	if opts.TaxonomyPath == "" {
		return fail(result, "taxonomy_required", "workflow-route requires --taxonomy <path>."), nil
	}
	if opts.SelectorRegistryPath == "" {
		return fail(result, "bundle_registry_required", "workflow-route requires --selector-registry <path>."), nil
	}
	if opts.TaskClass == "" {
		return failField(result, "classification_required_input_missing", "task_class is required.", "task_class"), nil
	}
	if opts.ClassificationReason == "" {
		return failField(result, "classification_reason_missing", "classification_reason is required.", "classification_reason"), nil
	}
	teal, tealStatus, tealMessage, tealField := deriveTealApplicability(opts)
	if tealStatus != "" {
		return failField(result, tealStatus, tealMessage, tealField), nil
	}
	result.TealApplicability = teal

	taxonomy, err := LoadTaxonomy(opts.TaxonomyPath)
	if err != nil {
		return fail(result, mapTaxonomyError(err), err.Error()), nil
	}
	result.Taxonomy = SourceEvidence{Path: taxonomy.Path, Version: taxonomy.Version, Checksum: taxonomy.Checksum}

	taskClassName, aliasOK := taxonomy.ResolveClass(opts.TaskClass)
	if !aliasOK {
		return failField(result, "classification_class_unsupported", "task_class is not declared by the taxonomy or vocabulary aliases.", "task_class"), nil
	}
	taskClass := taxonomy.Classes[taskClassName]
	result.TaskClass = taskClassName
	result.WorkPath = taskClass.WorkPath
	result.WorkMode = taskClass.WorkMode
	result.ExecutionMode = taskClass.ExecutionMode
	result.SelectedSpine = taskClass.DefaultSpine
	result.RequiredPhases = append([]string{}, taskClass.RequiredPhases...)
	result.SkippedPhaseReasons = skippedReasons(taskClass)
	result.CapabilityPosture = CapabilityPosture{KAH: taskClass.KAHDefault, KAB: taskClass.KABDefault, Kanban: taskClass.KanbanDefault, CodeGraph: taskClass.CodeGraphDefault}
	if result.SelectedSpine == "" {
		return failField(result, "bundle_default_spine_missing", "taxonomy task class does not declare default_spine.", "default_spine"), nil
	}
	if opts.SelectedSpine != "" {
		result.SelectedSpine = opts.SelectedSpine
	}

	registry, err := workflowregistry.Load(opts.SelectorRegistryPath)
	if err != nil {
		return fail(result, mapRegistryError(err), err.Error()), nil
	}
	result.SelectorRegistry = SourceEvidence{Path: registry.Path, Version: registry.Version, Checksum: registry.Checksum}

	query := workflowregistry.Query{
		TaskClass:            taskClassName,
		Labels:               opts.Labels,
		ChangedSurfaces:      opts.ChangedSurfaces,
		Risk:                 opts.Risk,
		RequiredAgents:       opts.RequiredAgents,
		RequiredCapabilities: opts.RequiredCapabilities,
	}
	match, err := workflowregistry.Select(registry, query)
	result.Labels = match.Query.Labels
	result.ChangedSurfaces = match.Query.ChangedSurfaces
	result.RequiredCapabilities = match.Query.RequiredCapabilities
	result.SelectorMatch = SelectorEvidence{Status: match.Status, CandidateIDs: workflowCandidateIDs(match.Candidates)}
	if err != nil {
		return fail(result, mapSelectorStatus(match.Status), err.Error()), nil
	}
	switch match.Status {
	case "selector_no_match":
		return fail(result, "bundle_no_match", "task classification did not match any standard bundle."), nil
	case "selector_ambiguous":
		return fail(result, "bundle_ambiguous", "task classification matched multiple standard bundles; provide a narrower classification or explicit workflow choice."), nil
	case "selector_matched":
		if match.Selected.WorkflowID != result.SelectedSpine {
			return failField(result, "bundle_selected_mismatch", "selected_spine does not match the deterministic registry match.", "selected_spine"), nil
		}
		result.WorkflowID = match.Selected.WorkflowID
		result.WorkflowPath = match.Selected.WorkflowPath
		result.SelectedBundle = match.Selected.WorkflowID
		result.SelectorMatch.WorkflowID = match.Selected.WorkflowID
		result.RequiredCapabilities = workflowregistry.NormalizeSelectorValues(append(result.RequiredCapabilities, match.Selected.Selector.RequiredCapabilitiesAll...))
	default:
		return fail(result, "bundle_registry_schema_unsupported", "selector returned unsupported status."), nil
	}
	result.OK = true
	result.Status = "bundle_route_matched"
	result.NextAction = "Use the selected bundle as WFLOW-008 input for run-local workflow materialization; workflow-route performed no KAH state writes."
	return result, nil
}

func LoadTaxonomy(path string) (Taxonomy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Taxonomy{}, fmt.Errorf("taxonomy is unreadable: %w", err)
	}
	taxonomy, err := ParseTaxonomy(string(data))
	if err != nil {
		return Taxonomy{}, err
	}
	sum := sha256.Sum256(data)
	taxonomy.Path = filepath.ToSlash(path)
	taxonomy.Checksum = "sha256:" + hex.EncodeToString(sum[:])
	return taxonomy, nil
}

func ParseTaxonomy(text string) (Taxonomy, error) {
	parser := taxonomyParser{taxonomy: Taxonomy{Aliases: map[string]string{}, Classes: map[string]TaskClass{}}}
	return parser.parse(text)
}

func (t Taxonomy) ResolveClass(value string) (string, bool) {
	value = normalizeAtom(value)
	if _, ok := t.Classes[value]; ok {
		return value, true
	}
	resolved := normalizeAtom(t.Aliases[value])
	if resolved == "" {
		return "", false
	}
	if _, ok := t.Classes[resolved]; !ok {
		return "", false
	}
	return resolved, true
}

func RenderHuman(result Result) string {
	if result.OK {
		return fmt.Sprintf("Status: %s\nBundle: %s\nTask class: %s\nDirect KAH state write: false\nNext: %s", result.Status, result.SelectedBundle, result.TaskClass, result.NextAction)
	}
	return fmt.Sprintf("Status: %s\nTask class: %s\nDirect KAH state write: false\nNext: %s", result.Status, result.TaskClass, result.NextAction)
}

func normalizeOptions(opts Options) Options {
	opts.TaxonomyPath = strings.TrimSpace(opts.TaxonomyPath)
	opts.SelectorRegistryPath = strings.TrimSpace(opts.SelectorRegistryPath)
	opts.TaskClass = normalizeAtom(opts.TaskClass)
	opts.ClassificationReason = strings.TrimSpace(opts.ClassificationReason)
	opts.SelectedSpine = normalizeAtom(opts.SelectedSpine)
	opts.TealSkipReason = strings.TrimSpace(opts.TealSkipReason)
	opts.TealWaiverApprovalRef = strings.TrimSpace(opts.TealWaiverApprovalRef)
	opts.TealWaiverScope = strings.TrimSpace(opts.TealWaiverScope)
	opts.TealWaiverExpiresAt = strings.TrimSpace(opts.TealWaiverExpiresAt)
	opts.Risk = normalizeAtom(opts.Risk)
	opts.Labels = workflowregistry.NormalizeSelectorValues(opts.Labels)
	opts.ChangedSurfaces = workflowregistry.NormalizeSelectorValues(opts.ChangedSurfaces)
	opts.RequiredAgents = workflowregistry.NormalizeSelectorValues(opts.RequiredAgents)
	opts.RequiredCapabilities = workflowregistry.NormalizeSelectorValues(opts.RequiredCapabilities)
	if len(opts.RequiredCapabilities) == 0 {
		opts.RequiredCapabilities = standardBundleRequiredCapabilities()
	}
	return opts
}

func standardBundleRequiredCapabilities() []string {
	return []string{"task_dag_schema_validation", "workflow_instance_state"}
}

func deriveTealApplicability(opts Options) (TealApplicability, string, string, string) {
	if opts.ProjectHasTealLane == nil || opts.UIUXChange == nil {
		return TealApplicability{}, "teal_applicability_required", "workflow-route requires explicit --project-has-teal-lane and --ui-ux-change facts.", "teal_applicability"
	}
	projectHasTealLane := *opts.ProjectHasTealLane
	uiUXChange := *opts.UIUXChange
	tealRequired := projectHasTealLane && uiUXChange
	if !tealRequired && strings.TrimSpace(opts.TealSkipReason) == "" {
		return TealApplicability{}, "teal_skip_reason_required", "workflow-route requires --teal-skip-reason when Teal is not required.", "teal_skip_reason"
	}
	return TealApplicability{
		ContractVersion:             "design003.v1",
		ProjectHasTealLane:          projectHasTealLane,
		UIUXChange:                  uiUXChange,
		TealRequired:                tealRequired,
		Derivation:                  "project_has_teal_lane && ui_ux_change",
		TealSkipReason:              opts.TealSkipReason,
		TealWaiverApproved:          opts.TealWaiverApproved,
		TealWaiverApprovalRef:       opts.TealWaiverApprovalRef,
		TealWaiverScope:             opts.TealWaiverScope,
		TealWaiverExpiresAt:         opts.TealWaiverExpiresAt,
		RequiredWhenTealRequired:    []string{"DESIGN_PLAN_GATE", "DESIGN_FIDELITY_REVIEW"},
		MissingRequiredStatus:       "required_teal_verdict_missing",
		OrdinaryReviewIsSubstitute:  false,
		MARReviewIsSubstitute:       false,
		BackendEvidenceIsSubstitute: false,
		HelperNotesAreSubstitute:    false,
	}, "", "", ""
}

func newResult(opts Options) Result {
	return Result{
		Command:              Command,
		Mode:                 Mode,
		Status:               "blocked",
		InputTaskClass:       opts.TaskClass,
		TaskClass:            opts.TaskClass,
		ClassificationReason: opts.ClassificationReason,
		SelectedSpine:        opts.SelectedSpine,
		Labels:               opts.Labels,
		ChangedSurfaces:      opts.ChangedSurfaces,
		RequiredCapabilities: opts.RequiredCapabilities,
		SkippedPhaseReasons:  map[string]string{},
		Diagnostics:          []Diagnostic{},
		ReasonCodes:          []string{},
		NextAction:           "Fix the reported workflow route blocker and rerun with one supported task classification.",
		DirectKAHStateWrite:  false,
	}
}

func fail(result Result, code string, message string) Result {
	result.OK = false
	if result.Status == "" || result.Status == "blocked" {
		result.Status = code
	}
	result.ReasonCodes = appendUnique(append(result.ReasonCodes, code)...)
	result.Diagnostics = append(result.Diagnostics, Diagnostic{Level: "error", Code: code, Message: message})
	result.NextAction = "Fix the reported workflow route blocker and rerun with one supported task classification."
	result.WorkflowID = ""
	result.WorkflowPath = ""
	result.SelectedBundle = ""
	return result
}

func failField(result Result, code string, message string, field string) Result {
	result = fail(result, code, message)
	result.Diagnostics[len(result.Diagnostics)-1].Field = field
	return result
}

func mapTaxonomyError(err error) string {
	message := err.Error()
	switch {
	case strings.Contains(message, "unreadable"):
		return "taxonomy_unreadable"
	case strings.Contains(message, "unsupported taxonomy version"):
		return "taxonomy_schema_unsupported"
	default:
		return "taxonomy_schema_unsupported"
	}
}

func mapRegistryError(err error) string {
	message := err.Error()
	switch {
	case strings.Contains(message, "unreadable"):
		return "bundle_registry_unreadable"
	default:
		return "bundle_registry_schema_unsupported"
	}
}

func mapSelectorStatus(status string) string {
	switch status {
	case "selector_required_input_missing":
		return "classification_required_input_missing"
	default:
		return status
	}
}

func workflowCandidateIDs(candidates []workflowregistry.Workflow) []string {
	ids := []string{}
	for _, candidate := range candidates {
		ids = append(ids, candidate.WorkflowID)
	}
	sort.Strings(ids)
	return ids
}

func skippedReasons(taskClass TaskClass) map[string]string {
	reasons := map[string]string{}
	for _, phase := range taskClass.SkippedByDefault {
		reasons[phase] = "skipped_by_default_for_" + taskClass.Name
	}
	return reasons
}

func appendUnique(values ...string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func normalizeAtom(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

type taxonomyParser struct {
	taxonomy    Taxonomy
	section     string
	currentKey  string
	currentCls  string
	currentList string
}

func (p *taxonomyParser) parse(text string) (Taxonomy, error) {
	for lineNo, raw := range strings.Split(text, "\n") {
		if strings.Contains(raw, "\t") {
			return Taxonomy{}, fmt.Errorf("line %d: tabs are not supported in task taxonomy YAML", lineNo+1)
		}
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		if indent%2 != 0 {
			return Taxonomy{}, fmt.Errorf("line %d: indentation must use two-space levels", lineNo+1)
		}
		if err := p.parseLine(lineNo+1, indent, trimmed); err != nil {
			return Taxonomy{}, err
		}
	}
	if p.taxonomy.Version != "0.1.0" {
		return Taxonomy{}, fmt.Errorf("unsupported taxonomy version %q", p.taxonomy.Version)
	}
	if len(p.taxonomy.Classes) == 0 {
		return Taxonomy{}, fmt.Errorf("taxonomy must declare task_classes")
	}
	for name, class := range p.taxonomy.Classes {
		if class.Name == "" {
			class.Name = name
		}
		if class.WorkPath == "" || class.WorkMode == "" || class.ExecutionMode == "" || class.DefaultSpine == "" {
			return Taxonomy{}, fmt.Errorf("task class %s requires default_work_path, default_work_mode, default_execution_mode, and default_spine", name)
		}
		if len(class.RequiredPhases) == 0 {
			return Taxonomy{}, fmt.Errorf("task class %s requires required_phases", name)
		}
		p.taxonomy.Classes[name] = class
	}
	for alias, target := range p.taxonomy.Aliases {
		if _, ok := p.taxonomy.Classes[target]; !ok {
			return Taxonomy{}, fmt.Errorf("taxonomy alias %s targets unsupported class %s", alias, target)
		}
	}
	return p.taxonomy, nil
}

func (p *taxonomyParser) parseLine(lineNo int, indent int, trimmed string) error {
	switch indent {
	case 0:
		p.currentKey = ""
		p.currentCls = ""
		p.currentList = ""
		if trimmed == "vocabulary_aliases:" {
			p.section = "aliases"
			return nil
		}
		if trimmed == "task_classes:" {
			p.section = "classes"
			return nil
		}
		key, value, ok := splitScalar(trimmed)
		if ok && key == "version" {
			p.taxonomy.Version = value
			return nil
		}
		p.section = ""
		return nil
	case 2:
		p.currentList = ""
		key, value, ok := splitScalar(trimmed)
		if !ok {
			return nil
		}
		switch p.section {
		case "aliases":
			if value == "" {
				p.currentKey = normalizeAtom(key)
				return nil
			}
		case "classes":
			if value == "" {
				name := normalizeAtom(key)
				p.currentCls = name
				p.taxonomy.Classes[name] = TaskClass{Name: name}
				return nil
			}
		}
	case 4:
		key, value, ok := splitScalar(trimmed)
		if !ok {
			return nil
		}
		if p.section == "aliases" && p.currentKey != "" && key == "registry_class" {
			p.taxonomy.Aliases[p.currentKey] = normalizeAtom(value)
			return nil
		}
		if p.section != "classes" || p.currentCls == "" {
			return nil
		}
		class := p.taxonomy.Classes[p.currentCls]
		p.currentList = ""
		switch key {
		case "default_work_path":
			class.WorkPath = value
		case "default_work_mode":
			class.WorkMode = value
		case "default_execution_mode":
			class.ExecutionMode = value
		case "default_spine":
			class.DefaultSpine = normalizeAtom(value)
		case "kah_default":
			class.KAHDefault = value
		case "kab_default":
			class.KABDefault = value
		case "kanban_default":
			class.KanbanDefault = value
		case "codegraph_default":
			class.CodeGraphDefault = value
		case "required_phases", "skipped_by_default":
			p.currentList = key
			if value != "" {
				values, err := parseInlineList(value)
				if err != nil {
					return fmt.Errorf("line %d: %w", lineNo, err)
				}
				if key == "required_phases" {
					class.RequiredPhases = values
				} else {
					class.SkippedByDefault = values
				}
			}
		}
		p.taxonomy.Classes[p.currentCls] = class
	case 6:
		if p.section != "classes" || p.currentCls == "" || p.currentList == "" || !strings.HasPrefix(trimmed, "- ") {
			return nil
		}
		value := cleanScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
		if value == "" {
			return fmt.Errorf("line %d: empty list value is unsupported", lineNo)
		}
		class := p.taxonomy.Classes[p.currentCls]
		switch p.currentList {
		case "required_phases":
			class.RequiredPhases = append(class.RequiredPhases, value)
		case "skipped_by_default":
			class.SkippedByDefault = append(class.SkippedByDefault, value)
		}
		p.taxonomy.Classes[p.currentCls] = class
	}
	return nil
}

func splitScalar(trimmed string) (string, string, bool) {
	key, value, ok := strings.Cut(trimmed, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), cleanScalar(value), true
}

func cleanScalar(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

func parseInlineList(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return []string{}, nil
	}
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil, fmt.Errorf("expected inline list")
	}
	body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
	if body == "" {
		return []string{}, nil
	}
	values := []string{}
	for _, raw := range strings.Split(body, ",") {
		item := cleanScalar(raw)
		if item == "" {
			return nil, fmt.Errorf("empty list value is unsupported")
		}
		values = append(values, item)
	}
	return values, nil
}
