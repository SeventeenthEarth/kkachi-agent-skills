package projectinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
)

const (
	RoleRegistryPath    = "registries/project-suite-roles.yaml"
	RoleRegistryVersion = "role-aware-project-suite/v1"
	SuiteModeFull       = "full"
	SuiteModeRoleSubset = "role_subset"
)

type RoleRegistryEvidence struct {
	Path     string `json:"path"`
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}

type RoleSkillEvidence struct {
	SourceSkill    string `json:"source_skill"`
	InstalledSkill string `json:"installed_skill"`
	SourcePackID   string `json:"source_pack_id,omitempty"`
	TargetPath     string `json:"target_path"`
	Reason         string `json:"reason,omitempty"`
}

type projectSuiteRole struct {
	ID                    string
	DisplayLabel          string
	SelectionMode         string
	RequiredSourceSkills  []string
	OptionalSourceSkills  []string
	ForbiddenSourceSkills []string
}

type projectSuiteRoleRegistry struct {
	Version  string
	Path     string
	Checksum string
	Roles    map[string]projectSuiteRole
}

func loadProjectSuiteRoleRegistry(sourceRepo string) (projectSuiteRoleRegistry, error) {
	path := filepath.Join(sourceRepo, filepath.FromSlash(RoleRegistryPath))
	data, err := os.ReadFile(path)
	if err != nil {
		return projectSuiteRoleRegistry{}, fmt.Errorf("requires readable %s: %w", RoleRegistryPath, err)
	}
	sum := sha256.Sum256(data)
	registry, err := parseProjectSuiteRoleRegistry(string(data))
	if err != nil {
		return projectSuiteRoleRegistry{}, err
	}
	registry.Path = RoleRegistryPath
	registry.Checksum = "sha256:" + hex.EncodeToString(sum[:])
	return registry, nil
}

func parseProjectSuiteRoleRegistry(text string) (projectSuiteRoleRegistry, error) {
	registry := projectSuiteRoleRegistry{Roles: map[string]projectSuiteRole{}}
	lines := strings.Split(text, "\n")
	currentRole := ""
	currentList := ""
	inRoles := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent == 0 {
			currentRole = ""
			currentList = ""
			inRoles = false
			if key, value, ok := splitYAMLScalar(trimmed); ok && key == "version" {
				registry.Version = value
			}
			if trimmed == "roles:" {
				inRoles = true
			}
			continue
		}
		if !inRoles && currentRole == "" {
			continue
		}
		if indent == 2 && strings.HasSuffix(trimmed, ":") {
			currentRole = strings.TrimSuffix(trimmed, ":")
			currentList = ""
			role := registry.Roles[currentRole]
			role.ID = currentRole
			registry.Roles[currentRole] = role
			continue
		}
		if currentRole == "" {
			continue
		}
		role := registry.Roles[currentRole]
		if indent == 4 {
			currentList = ""
			key, value, ok := splitYAMLScalar(trimmed)
			if !ok {
				continue
			}
			switch key {
			case "display_label":
				role.DisplayLabel = value
			case "selection_mode":
				role.SelectionMode = value
			case "required_source_skills":
				if value == "*" {
					role.RequiredSourceSkills = []string{"*"}
				} else if value == "[]" || value == "" {
					role.RequiredSourceSkills = nil
					currentList = key
				}
			case "optional_source_skills":
				if value != "[]" && value == "" {
					currentList = key
				}
			case "forbidden_source_skills":
				if value != "[]" && value == "" {
					currentList = key
				}
			default:
				// Metadata fields are allowed but not behavior-bearing for KASROLE-002.
			}
			registry.Roles[currentRole] = role
			continue
		}
		if indent == 6 && strings.HasPrefix(trimmed, "- ") && currentList != "" {
			value := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `"'`)
			switch currentList {
			case "required_source_skills":
				role.RequiredSourceSkills = append(role.RequiredSourceSkills, value)
			case "optional_source_skills":
				role.OptionalSourceSkills = append(role.OptionalSourceSkills, value)
			case "forbidden_source_skills":
				role.ForbiddenSourceSkills = append(role.ForbiddenSourceSkills, value)
			}
			registry.Roles[currentRole] = role
		}
	}
	return registry, nil
}

func splitYAMLScalar(trimmed string) (string, string, bool) {
	key, value, ok := strings.Cut(trimmed, ":")
	if !ok {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	value = strings.Trim(strings.TrimSpace(value), `"'`)
	return key, value, true
}

func resolveProjectSuiteRole(sourceRepo string, suiteRole string, packs []discovery.SourcePack, project string) (RoleRegistryEvidence, projectSuiteRole, []discovery.SourcePack, []RoleSkillEvidence, []RoleSkillEvidence, []Conflict, []discovery.Diagnostic) {
	registryEvidence := RoleRegistryEvidence{Path: RoleRegistryPath}
	conflicts := []Conflict{}
	diagnostics := []discovery.Diagnostic{}
	if strings.TrimSpace(suiteRole) == "" {
		c := conflict("suite_role_required", project, "", "", "project suite install requires explicit --suite-role; role is never inferred from profile name", "Rerun with --suite-role blue_commander, red_reviewer, orange_pm_reviewer, gray_scribe, or teal_design_reviewer when Teal applies.")
		return registryEvidence, projectSuiteRole{}, nil, nil, nil, []Conflict{c}, []discovery.Diagnostic{{Level: "error", Code: c.Condition, Message: c.Message}}
	}
	registry, err := loadProjectSuiteRoleRegistry(sourceRepo)
	if err != nil {
		c := conflict("role_registry_unreadable", project, "", RoleRegistryPath, err.Error(), "Restore a readable registries/project-suite-roles.yaml and rerun.")
		return registryEvidence, projectSuiteRole{}, nil, nil, nil, []Conflict{c}, []discovery.Diagnostic{{Level: "error", Code: c.Condition, Message: c.Message}}
	}
	registryEvidence = RoleRegistryEvidence{Path: registry.Path, Version: registry.Version, Checksum: registry.Checksum}
	if registry.Version != RoleRegistryVersion {
		c := conflict("unsupported_role_registry_schema", project, "", RoleRegistryPath, fmt.Sprintf("unsupported role registry version: %q", registry.Version), "Use role-aware-project-suite/v1 or update KAS compatibility before install.")
		return registryEvidence, projectSuiteRole{}, nil, nil, nil, []Conflict{c}, []discovery.Diagnostic{{Level: "error", Code: c.Condition, Message: c.Message}}
	}
	role, ok := registry.Roles[suiteRole]
	if !ok {
		c := conflict("unknown_suite_role", project, "", RoleRegistryPath, fmt.Sprintf("unknown suite_role %q", suiteRole), "Use a registered suite role from registries/project-suite-roles.yaml.")
		return registryEvidence, projectSuiteRole{}, nil, nil, nil, []Conflict{c}, []discovery.Diagnostic{{Level: "error", Code: c.Condition, Message: c.Message}}
	}
	if strings.TrimSpace(role.DisplayLabel) == "" {
		c := conflict("suite_role_display_label_missing", project, "", RoleRegistryPath, "suite role "+suiteRole+" is missing display_label", "Fix the role registry before project install.")
		conflicts = append(conflicts, c)
		diagnostics = append(diagnostics, discovery.Diagnostic{Level: "error", Code: c.Condition, Message: c.Message})
	}
	if role.SelectionMode != SuiteModeFull && role.SelectionMode != "full_source_suite" && role.SelectionMode != "explicit_source_subset" {
		c := conflict("unsupported_suite_role_selection_mode", project, "", RoleRegistryPath, fmt.Sprintf("suite role %s uses unsupported selection_mode %q", suiteRole, role.SelectionMode), "Use full_source_suite or explicit_source_subset.")
		conflicts = append(conflicts, c)
		diagnostics = append(diagnostics, discovery.Diagnostic{Level: "error", Code: c.Condition, Message: c.Message})
	}
	bySourceSkill := map[string]discovery.SourcePack{}
	for _, pack := range packs {
		bySourceSkill[sourceSkillID(pack.PackID)] = pack
	}
	selectedIDs := []string{}
	if role.SelectionMode == "full_source_suite" || sameStringSet(role.RequiredSourceSkills, []string{"*"}) {
		for _, pack := range packs {
			selectedIDs = append(selectedIDs, sourceSkillID(pack.PackID))
		}
		role.SelectionMode = "full_source_suite"
	} else {
		selectedIDs = append(selectedIDs, role.RequiredSourceSkills...)
	}
	forbidden := map[string]bool{}
	for _, id := range role.ForbiddenSourceSkills {
		forbidden[id] = true
	}
	selectedMap := map[string]bool{}
	for _, id := range selectedIDs {
		selectedMap[id] = true
		if forbidden[id] {
			c := conflict("forbidden_selected_skill", project, renderInstalledSkill(project, id), "", fmt.Sprintf("suite role %s selects forbidden source skill %s", suiteRole, id), "Fix registries/project-suite-roles.yaml before project install.")
			conflicts = append(conflicts, c)
			diagnostics = append(diagnostics, discovery.Diagnostic{Level: "error", Code: c.Condition, Message: c.Message})
		}
		if _, ok := bySourceSkill[id]; !ok {
			c := conflict("selected_source_skill_missing", project, renderInstalledSkill(project, id), "", fmt.Sprintf("suite role %s selects missing source skill %s", suiteRole, id), "Fix skill-pack.yaml or the role registry before project install.")
			conflicts = append(conflicts, c)
			diagnostics = append(diagnostics, discovery.Diagnostic{Level: "error", Code: c.Condition, Message: c.Message})
		}
	}
	selectedPacks := []discovery.SourcePack{}
	selected := []RoleSkillEvidence{}
	excluded := []RoleSkillEvidence{}
	for _, pack := range packs {
		id := sourceSkillID(pack.PackID)
		installed := renderInstalledSkill(project, id)
		target := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
		evidence := RoleSkillEvidence{SourceSkill: id, InstalledSkill: installed, SourcePackID: pack.PackID, TargetPath: target}
		if selectedMap[id] {
			selectedPacks = append(selectedPacks, pack)
			selected = append(selected, evidence)
		} else {
			evidence.Reason = "outside_suite_role"
			excluded = append(excluded, evidence)
		}
	}
	sort.Slice(selectedPacks, func(i, j int) bool { return selectedPacks[i].PackID < selectedPacks[j].PackID })
	sortRoleSkillEvidence(selected)
	sortRoleSkillEvidence(excluded)
	return registryEvidence, role, selectedPacks, selected, excluded, conflicts, diagnostics
}

func sameStringSet(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sortRoleSkillEvidence(skills []RoleSkillEvidence) {
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].TargetPath == skills[j].TargetPath {
			return skills[i].SourceSkill < skills[j].SourceSkill
		}
		return skills[i].TargetPath < skills[j].TargetPath
	})
}
