package workflowregistry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	RegistryVersion      = "kas-task-dag-workflow-registry/v1"
	NodeContractsVersion = "kas-node-contracts/v1"
	NoFallbackPolicy     = "none_fail_closed"
)

type Registry struct {
	Version       string
	Path          string
	Checksum      string
	Workflows     []Workflow
	NodeContracts []NodeContract
}

type Workflow struct {
	WorkflowID     string
	WorkflowPath   string
	Selector       Selector
	FallbackPolicy string
}

type Selector struct {
	TaskClasses             []string
	LabelsAny               []string
	LabelsAll               []string
	ChangedSurfacesAny      []string
	RiskLevels              []string
	RequiredAgentsAll       []string
	RequiredCapabilitiesAll []string
}

type Query struct {
	TaskClass            string
	Labels               []string
	ChangedSurfaces      []string
	Risk                 string
	RequiredAgents       []string
	RequiredCapabilities []string
}

type MatchResult struct {
	Status     string
	Candidates []Workflow
	Selected   Workflow
	Query      Query
}

type NodeContract struct {
	WorkflowID        string   `json:"workflow_id"`
	NodeID            string   `json:"node_id"`
	TaskClass         string   `json:"task_class,omitempty"`
	OwnerRole         string   `json:"owner_role"`
	ExecutionLane     string   `json:"execution_lane"`
	RequiredInputs    []string `json:"required_inputs"`
	ExpectedArtifacts []string `json:"expected_artifacts"`
	PromptRef         string   `json:"prompt_ref"`
	ApprovalRequired  bool     `json:"approval_required"`
	FallbackPolicy    string   `json:"fallback_policy"`
	VerificationGate  string   `json:"verification_gate"`
}

type nodeContractBundle struct {
	SchemaVersion string         `json:"schema_version"`
	Ref           string         `json:"ref"`
	Contracts     []NodeContract `json:"contracts"`
}

func Load(path string) (Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, fmt.Errorf("selector registry is unreadable: %w", err)
	}
	registry, err := Parse(string(data))
	if err != nil {
		return Registry{}, err
	}
	sum := sha256.Sum256(data)
	registry.Path = filepath.ToSlash(path)
	registry.Checksum = "sha256:" + hex.EncodeToString(sum[:])
	return registry, nil
}

func LoadJSONContracts(path string, expectedRef string) ([]NodeContract, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("node-contract source is unreadable: %w", err)
	}
	var bundle nodeContractBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, "", fmt.Errorf("node-contract source must be JSON for explicit mode")
	}
	if bundle.SchemaVersion != NodeContractsVersion {
		return nil, "", fmt.Errorf("node-contract source must use schema_version %s", NodeContractsVersion)
	}
	if expectedRef != "" && bundle.Ref != expectedRef {
		return nil, "", fmt.Errorf("node-contract source ref does not match --node-contract-ref")
	}
	if err := ValidateNodeContracts(bundle.Contracts); err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	return bundle.Contracts, "sha256:" + hex.EncodeToString(sum[:]), nil
}

func Parse(text string) (Registry, error) {
	parser := registryParser{}
	registry, err := parser.parse(text)
	if err != nil {
		return Registry{}, err
	}
	if registry.Version != RegistryVersion {
		return Registry{}, fmt.Errorf("unsupported selector registry version %q", registry.Version)
	}
	if len(registry.Workflows) == 0 {
		return Registry{}, fmt.Errorf("selector registry must include at least one workflow")
	}
	if err := ValidateNodeContracts(registry.NodeContracts); err != nil {
		return Registry{}, err
	}
	for _, contract := range registry.NodeContracts {
		if contract.TaskClass == "" {
			return Registry{}, fmt.Errorf("registry node contracts require task_class")
		}
	}
	seenWorkflows := map[string]bool{}
	for _, workflow := range registry.Workflows {
		if workflow.WorkflowID == "" || workflow.WorkflowPath == "" {
			return Registry{}, fmt.Errorf("workflow registry entries require workflow_id and workflow_path")
		}
		if workflow.FallbackPolicy != NoFallbackPolicy {
			return Registry{}, fmt.Errorf("workflow %s uses unsupported fallback_policy %q", workflow.WorkflowID, workflow.FallbackPolicy)
		}
		if len(workflow.Selector.TaskClasses) == 0 {
			return Registry{}, fmt.Errorf("workflow %s selector requires task_classes", workflow.WorkflowID)
		}
		if seenWorkflows[workflow.WorkflowID] {
			return Registry{}, fmt.Errorf("duplicate workflow_id %q", workflow.WorkflowID)
		}
		seenWorkflows[workflow.WorkflowID] = true
	}
	return registry, nil
}

func ValidateNodeContracts(contracts []NodeContract) error {
	if len(contracts) == 0 {
		return fmt.Errorf("node-contract source must include at least one contract")
	}
	seen := map[string]bool{}
	for _, contract := range contracts {
		if contract.WorkflowID == "" || contract.NodeID == "" || contract.OwnerRole == "" || contract.ExecutionLane == "" || contract.PromptRef == "" || contract.FallbackPolicy == "" || contract.VerificationGate == "" {
			return fmt.Errorf("node contracts require workflow_id, node_id, owner_role, execution_lane, prompt_ref, fallback_policy, and verification_gate")
		}
		if len(contract.RequiredInputs) == 0 || len(contract.ExpectedArtifacts) == 0 {
			return fmt.Errorf("node contracts require required_inputs and expected_artifacts")
		}
		if contract.FallbackPolicy != NoFallbackPolicy {
			return fmt.Errorf("node contract %s/%s uses unsupported fallback_policy %q", contract.WorkflowID, contract.NodeID, contract.FallbackPolicy)
		}
		key := contract.WorkflowID + "\x00" + contract.NodeID
		if seen[key] {
			return fmt.Errorf("duplicate node contract for workflow_id %q node_id %q", contract.WorkflowID, contract.NodeID)
		}
		seen[key] = true
	}
	return nil
}

func Select(registry Registry, query Query) (MatchResult, error) {
	query = normalizeQuery(query)
	if query.TaskClass == "" {
		return MatchResult{Status: "selector_required_input_missing", Query: query}, fmt.Errorf("selector mode requires --task-class")
	}
	candidates := []Workflow{}
	for _, workflow := range registry.Workflows {
		if selectorMatches(workflow.Selector, query) {
			candidates = append(candidates, workflow)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].WorkflowID < candidates[j].WorkflowID })
	result := MatchResult{Candidates: candidates, Query: query}
	switch len(candidates) {
	case 0:
		result.Status = "selector_no_match"
	case 1:
		result.Status = "selector_matched"
		result.Selected = candidates[0]
	default:
		result.Status = "selector_ambiguous"
	}
	return result, nil
}

func ContractsForWorkflow(contracts []NodeContract, workflowID string) []NodeContract {
	selected := []NodeContract{}
	for _, contract := range contracts {
		if contract.WorkflowID == workflowID {
			selected = append(selected, contract)
		}
	}
	return selected
}

func ValidateContractsCoverNodeIDs(contracts []NodeContract, workflowID string, nodeIDs []string) error {
	if len(nodeIDs) == 0 {
		return nil
	}
	contractByNode := map[string]bool{}
	for _, contract := range contracts {
		if contract.WorkflowID == workflowID {
			contractByNode[contract.NodeID] = true
		}
	}
	for _, nodeID := range nodeIDs {
		if !contractByNode[nodeID] {
			return fmt.Errorf("workflow node %s has no registry node contract", nodeID)
		}
	}
	nodeSet := map[string]bool{}
	for _, nodeID := range nodeIDs {
		nodeSet[nodeID] = true
	}
	for _, contract := range contracts {
		if contract.WorkflowID == workflowID && !nodeSet[contract.NodeID] {
			return fmt.Errorf("registry node contract %s is not present in KAH workflow explain output", contract.NodeID)
		}
	}
	return nil
}

func normalizeQuery(query Query) Query {
	query.TaskClass = normalizeAtom(query.TaskClass)
	query.Risk = normalizeAtom(query.Risk)
	query.Labels = NormalizeSelectorValues(query.Labels)
	query.ChangedSurfaces = NormalizeSelectorValues(query.ChangedSurfaces)
	query.RequiredAgents = NormalizeSelectorValues(query.RequiredAgents)
	query.RequiredCapabilities = NormalizeSelectorValues(query.RequiredCapabilities)
	return query
}

func selectorMatches(selector Selector, query Query) bool {
	if !contains(NormalizeSelectorValues(selector.TaskClasses), query.TaskClass) {
		return false
	}
	if len(selector.RiskLevels) > 0 && (query.Risk == "" || !contains(NormalizeSelectorValues(selector.RiskLevels), query.Risk)) {
		return false
	}
	if len(selector.LabelsAny) > 0 && !intersects(NormalizeSelectorValues(selector.LabelsAny), query.Labels) {
		return false
	}
	if len(selector.LabelsAll) > 0 && !containsAll(query.Labels, NormalizeSelectorValues(selector.LabelsAll)) {
		return false
	}
	if len(selector.ChangedSurfacesAny) > 0 && !intersects(NormalizeSelectorValues(selector.ChangedSurfacesAny), query.ChangedSurfaces) {
		return false
	}
	if len(selector.RequiredAgentsAll) > 0 && !containsAll(query.RequiredAgents, NormalizeSelectorValues(selector.RequiredAgentsAll)) {
		return false
	}
	if len(selector.RequiredCapabilitiesAll) > 0 && !containsAll(query.RequiredCapabilities, NormalizeSelectorValues(selector.RequiredCapabilitiesAll)) {
		return false
	}
	return true
}

type registryParser struct {
	registry       Registry
	section        string
	workflowIndex  int
	contractIndex  int
	inSelector     bool
	currentListKey string
}

func (p *registryParser) parse(text string) (Registry, error) {
	p.workflowIndex = -1
	p.contractIndex = -1
	for lineNo, raw := range strings.Split(text, "\n") {
		if strings.Contains(raw, "\t") {
			return Registry{}, fmt.Errorf("line %d: tabs are not supported in selector registry YAML", lineNo+1)
		}
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		if indent%2 != 0 {
			return Registry{}, fmt.Errorf("line %d: indentation must use two-space levels", lineNo+1)
		}
		if err := p.parseLine(lineNo+1, indent, trimmed); err != nil {
			return Registry{}, err
		}
	}
	return p.registry, nil
}

func (p *registryParser) parseLine(lineNo int, indent int, trimmed string) error {
	switch indent {
	case 0:
		p.currentListKey = ""
		p.inSelector = false
		if trimmed == "workflows:" {
			p.section = "workflows"
			return nil
		}
		if trimmed == "node_contracts:" {
			p.section = "node_contracts"
			return nil
		}
		key, value, ok := splitScalar(trimmed)
		if !ok || key != "version" {
			return fmt.Errorf("line %d: unsupported top-level registry field %q", lineNo, trimmed)
		}
		p.registry.Version = value
		return nil
	case 2:
		p.currentListKey = ""
		p.inSelector = false
		if !strings.HasPrefix(trimmed, "- ") {
			return fmt.Errorf("line %d: expected list item", lineNo)
		}
		item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if item == "" {
			return fmt.Errorf("line %d: empty list item is unsupported", lineNo)
		}
		key, value, ok := splitScalar(item)
		if !ok {
			return fmt.Errorf("line %d: list item must start with a scalar field", lineNo)
		}
		switch p.section {
		case "workflows":
			p.registry.Workflows = append(p.registry.Workflows, Workflow{})
			p.workflowIndex = len(p.registry.Workflows) - 1
			p.contractIndex = -1
			return p.setWorkflowScalar(lineNo, key, value)
		case "node_contracts":
			p.registry.NodeContracts = append(p.registry.NodeContracts, NodeContract{})
			p.contractIndex = len(p.registry.NodeContracts) - 1
			p.workflowIndex = -1
			return p.setContractScalar(lineNo, key, value)
		default:
			return fmt.Errorf("line %d: list item outside supported registry section", lineNo)
		}
	case 4:
		p.currentListKey = ""
		key, value, ok := splitScalar(trimmed)
		if !ok {
			return fmt.Errorf("line %d: expected scalar field", lineNo)
		}
		if p.section == "workflows" && p.workflowIndex >= 0 {
			if key == "selector" && value == "" {
				p.inSelector = true
				return nil
			}
			p.inSelector = false
			return p.setWorkflowScalar(lineNo, key, value)
		}
		if p.section == "node_contracts" && p.contractIndex >= 0 {
			return p.setContractScalar(lineNo, key, value)
		}
		return fmt.Errorf("line %d: field outside supported list item", lineNo)
	case 6:
		if p.section == "node_contracts" && p.contractIndex >= 0 && p.currentListKey != "" && strings.HasPrefix(trimmed, "- ") {
			value, err := parseBlockListValue(lineNo, trimmed)
			if err != nil {
				return err
			}
			return p.appendContractListValue(lineNo, p.currentListKey, value)
		}
		if p.section != "workflows" || p.workflowIndex < 0 || !p.inSelector {
			return fmt.Errorf("line %d: nesting is supported only inside workflow selector", lineNo)
		}
		key, value, ok := splitScalar(trimmed)
		if !ok {
			return fmt.Errorf("line %d: expected selector scalar", lineNo)
		}
		p.currentListKey = ""
		if value == "" {
			p.currentListKey = key
			return p.setSelectorList(lineNo, key, nil)
		}
		values, err := parseInlineList(value)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
		return p.setSelectorList(lineNo, key, values)
	case 8:
		if p.currentListKey == "" || !strings.HasPrefix(trimmed, "- ") {
			return fmt.Errorf("line %d: unsupported nested shape", lineNo)
		}
		value, err := parseBlockListValue(lineNo, trimmed)
		if err != nil {
			return err
		}
		if p.section == "workflows" && p.workflowIndex >= 0 && p.inSelector {
			return p.appendSelectorListValue(lineNo, p.currentListKey, value)
		}
		return fmt.Errorf("line %d: unsupported nested list", lineNo)
	default:
		return fmt.Errorf("line %d: unsupported nesting depth", lineNo)
	}
}

func (p *registryParser) setWorkflowScalar(lineNo int, key string, value string) error {
	if p.workflowIndex < 0 {
		return fmt.Errorf("line %d: workflow field without workflow item", lineNo)
	}
	workflow := p.registry.Workflows[p.workflowIndex]
	switch key {
	case "workflow_id":
		workflow.WorkflowID = value
	case "workflow_path":
		workflow.WorkflowPath = filepath.ToSlash(value)
	case "fallback_policy":
		workflow.FallbackPolicy = value
	default:
		return fmt.Errorf("line %d: unsupported workflow field %q", lineNo, key)
	}
	p.registry.Workflows[p.workflowIndex] = workflow
	return nil
}

func (p *registryParser) setContractScalar(lineNo int, key string, value string) error {
	if p.contractIndex < 0 {
		return fmt.Errorf("line %d: node contract field without contract item", lineNo)
	}
	contract := p.registry.NodeContracts[p.contractIndex]
	switch key {
	case "workflow_id":
		contract.WorkflowID = value
	case "node_id":
		contract.NodeID = value
	case "task_class":
		contract.TaskClass = value
	case "owner_role":
		contract.OwnerRole = value
	case "execution_lane":
		contract.ExecutionLane = value
	case "required_inputs":
		if value == "" {
			p.currentListKey = key
			contract.RequiredInputs = nil
			p.registry.NodeContracts[p.contractIndex] = contract
			return nil
		}
		values, err := parseInlineList(value)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
		contract.RequiredInputs = values
	case "expected_artifacts":
		if value == "" {
			p.currentListKey = key
			contract.ExpectedArtifacts = nil
			p.registry.NodeContracts[p.contractIndex] = contract
			return nil
		}
		values, err := parseInlineList(value)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
		contract.ExpectedArtifacts = values
	case "prompt_ref":
		contract.PromptRef = value
	case "approval_required":
		switch value {
		case "true":
			contract.ApprovalRequired = true
		case "false":
			contract.ApprovalRequired = false
		default:
			return fmt.Errorf("line %d: approval_required must be true or false", lineNo)
		}
	case "fallback_policy":
		contract.FallbackPolicy = value
	case "verification_gate":
		contract.VerificationGate = value
	default:
		return fmt.Errorf("line %d: unsupported node_contract field %q", lineNo, key)
	}
	p.registry.NodeContracts[p.contractIndex] = contract
	return nil
}

func (p *registryParser) appendContractListValue(lineNo int, key string, value string) error {
	contract := p.registry.NodeContracts[p.contractIndex]
	switch key {
	case "required_inputs":
		contract.RequiredInputs = append(contract.RequiredInputs, value)
	case "expected_artifacts":
		contract.ExpectedArtifacts = append(contract.ExpectedArtifacts, value)
	default:
		return fmt.Errorf("line %d: unsupported node_contract list %q", lineNo, key)
	}
	p.registry.NodeContracts[p.contractIndex] = contract
	return nil
}

func (p *registryParser) setSelectorList(lineNo int, key string, values []string) error {
	workflow := p.registry.Workflows[p.workflowIndex]
	switch key {
	case "task_classes":
		workflow.Selector.TaskClasses = values
	case "labels_any":
		workflow.Selector.LabelsAny = values
	case "labels_all":
		workflow.Selector.LabelsAll = values
	case "changed_surfaces_any":
		workflow.Selector.ChangedSurfacesAny = values
	case "risk_levels":
		workflow.Selector.RiskLevels = values
	case "required_agents_all":
		workflow.Selector.RequiredAgentsAll = values
	case "required_capabilities_all":
		workflow.Selector.RequiredCapabilitiesAll = values
	default:
		return fmt.Errorf("line %d: unsupported selector field %q", lineNo, key)
	}
	p.registry.Workflows[p.workflowIndex] = workflow
	return nil
}

func (p *registryParser) appendSelectorListValue(lineNo int, key string, value string) error {
	workflow := p.registry.Workflows[p.workflowIndex]
	switch key {
	case "task_classes":
		workflow.Selector.TaskClasses = append(workflow.Selector.TaskClasses, value)
	case "labels_any":
		workflow.Selector.LabelsAny = append(workflow.Selector.LabelsAny, value)
	case "labels_all":
		workflow.Selector.LabelsAll = append(workflow.Selector.LabelsAll, value)
	case "changed_surfaces_any":
		workflow.Selector.ChangedSurfacesAny = append(workflow.Selector.ChangedSurfacesAny, value)
	case "risk_levels":
		workflow.Selector.RiskLevels = append(workflow.Selector.RiskLevels, value)
	case "required_agents_all":
		workflow.Selector.RequiredAgentsAll = append(workflow.Selector.RequiredAgentsAll, value)
	case "required_capabilities_all":
		workflow.Selector.RequiredCapabilitiesAll = append(workflow.Selector.RequiredCapabilitiesAll, value)
	default:
		return fmt.Errorf("line %d: unsupported selector list %q", lineNo, key)
	}
	p.registry.Workflows[p.workflowIndex] = workflow
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

func parseBlockListValue(lineNo int, trimmed string) (string, error) {
	value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
	if _, _, ok := splitScalar(value); ok {
		return "", fmt.Errorf("line %d: unsupported mapping list item", lineNo)
	}
	value = cleanScalar(value)
	if value == "" {
		return "", fmt.Errorf("line %d: empty list value is unsupported", lineNo)
	}
	return value, nil
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

func NormalizeSelectorValues(values []string) []string {
	seen := map[string]bool{}
	normalized := []string{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = normalizeAtom(part)
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			normalized = append(normalized, part)
		}
	}
	sort.Strings(normalized)
	return normalized
}

func normalizeAtom(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func containsAll(values []string, required []string) bool {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	for _, requiredValue := range required {
		if !set[requiredValue] {
			return false
		}
	}
	return true
}

func intersects(left []string, right []string) bool {
	set := map[string]bool{}
	for _, value := range left {
		set[value] = true
	}
	for _, value := range right {
		if set[value] {
			return true
		}
	}
	return false
}
