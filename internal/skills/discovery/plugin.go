package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	PluginPackageManifestPath = "skill-pack.yaml"
	PluginRoleManifestVersion = "kas-plugin-role-manifest/v1"
	PluginLoadPolicy          = "plugin_qualified_fail_closed"
)

type PluginPackage struct {
	Namespace       string
	TopLevelName    string
	Version         string
	ManifestPath    string
	PackageManifest string
	LoadPolicy      string
	Skills          []string
	RoleManifests   map[string]string
	GuideSkills     []string
}

type PluginQualifiedSkillReadback struct {
	Namespace            string `json:"namespace"`
	RequestedName        string `json:"requested_name"`
	CanonicalSkillID     string `json:"canonical_skill_id"`
	PluginQualifiedName  string `json:"plugin_qualified_name"`
	SourcePackagePath    string `json:"source_package_path"`
	ResolvedSkillPath    string `json:"resolved_skill_path"`
	PackageVersion       string `json:"package_version"`
	PackageSource        string `json:"package_source"`
	ProfileWrapperSource bool   `json:"profile_wrapper_source"`
	FallbackUsed         bool   `json:"fallback_used"`
}

type SourceRoleManifestReadback struct {
	Namespace      string                 `json:"namespace"`
	PackageVersion string                 `json:"package_version"`
	PackageSource  string                 `json:"package_source"`
	Roles          []RoleManifestReadback `json:"roles"`
}

type RoleManifestReadback struct {
	Role                 string   `json:"role"`
	SourceControlledPath string   `json:"source_controlled_path"`
	Version              string   `json:"version"`
	PackageSource        string   `json:"package_source"`
	PackageVersion       string   `json:"package_version"`
	SkillIDs             []string `json:"skill_ids"`
}

func LoadPluginPackage(repo string) (PluginPackage, error) {
	sourceRepo, err := FindSourceRepo(repo)
	if err != nil {
		return PluginPackage{}, err
	}
	data, err := os.ReadFile(filepath.Join(sourceRepo, PluginPackageManifestPath))
	if err != nil {
		return PluginPackage{}, fmt.Errorf("official plugin package manifest unreadable: %w", err)
	}
	pkg := parsePluginPackageManifest(string(data))
	pkg.ManifestPath = PluginPackageManifestPath
	if pkg.TopLevelName == "" {
		return PluginPackage{}, fmt.Errorf("official plugin package top-level name missing in %s", PluginPackageManifestPath)
	}
	if pkg.Namespace == "" {
		return PluginPackage{}, fmt.Errorf("official plugin package plugin.namespace missing in %s", PluginPackageManifestPath)
	}
	if pkg.Namespace != pkg.TopLevelName {
		return PluginPackage{}, fmt.Errorf("official plugin package plugin.namespace %q does not match top-level name %q", pkg.Namespace, pkg.TopLevelName)
	}
	if pkg.LoadPolicy == "" {
		return PluginPackage{}, fmt.Errorf("official plugin package plugin.load_policy missing in %s", PluginPackageManifestPath)
	}
	if pkg.LoadPolicy != PluginLoadPolicy {
		return PluginPackage{}, fmt.Errorf("official plugin package plugin.load_policy %q is unsupported; want %s", pkg.LoadPolicy, PluginLoadPolicy)
	}
	if pkg.PackageManifest != "" && !isCurrentPluginPackageManifest(pkg.PackageManifest) {
		return PluginPackage{}, fmt.Errorf("official plugin package plugin.package_manifest %q must reference %s", pkg.PackageManifest, PluginPackageManifestPath)
	}
	if pkg.Version == "" {
		return PluginPackage{}, fmt.Errorf("official plugin package version missing in %s", PluginPackageManifestPath)
	}
	if len(pkg.Skills) == 0 {
		return PluginPackage{}, fmt.Errorf("official plugin package has no registered base skills")
	}
	for _, skillID := range pkg.Skills {
		if invalidPluginSkillID(skillID) {
			return PluginPackage{}, fmt.Errorf("invalid plugin skill id in %s: %s", PluginPackageManifestPath, skillID)
		}
	}
	sort.Strings(pkg.Skills)
	if pkg.RoleManifests == nil {
		pkg.RoleManifests = map[string]string{}
	}
	sort.Strings(pkg.GuideSkills)
	return pkg, nil
}

func BuildPluginQualifiedSkillReadback(repo string, qualified string) (PluginQualifiedSkillReadback, error) {
	sourceRepo, err := FindSourceRepo(repo)
	if err != nil {
		return PluginQualifiedSkillReadback{}, err
	}
	pkg, err := LoadPluginPackage(sourceRepo)
	if err != nil {
		return PluginQualifiedSkillReadback{}, err
	}
	namespace, requested, ok := strings.Cut(strings.TrimSpace(qualified), ":")
	if !ok || namespace == "" || requested == "" {
		return PluginQualifiedSkillReadback{}, fmt.Errorf("plugin-qualified skill must use namespace:skill")
	}
	if namespace != pkg.Namespace {
		return PluginQualifiedSkillReadback{}, fmt.Errorf("plugin namespace %q does not match official namespace %q", namespace, pkg.Namespace)
	}
	if malformedPluginQualifiedComponent(namespace) || malformedPluginQualifiedComponent(requested) || strings.Contains(requested, ":") {
		return PluginQualifiedSkillReadback{}, fmt.Errorf("plugin-qualified skill must use namespace:skill")
	}
	canonical, ok := resolvePluginSkillID(pkg.Skills, requested)
	if !ok {
		return PluginQualifiedSkillReadback{}, fmt.Errorf("official plugin base skill not registered: %s", qualified)
	}
	if invalidPluginSkillID(canonical) {
		return PluginQualifiedSkillReadback{}, fmt.Errorf("invalid plugin skill id resolved from %s: %s", qualified, canonical)
	}
	skillPath := filepath.ToSlash(filepath.Join("skills", canonical, "SKILL.md"))
	if !isFile(filepath.Join(sourceRepo, filepath.FromSlash(skillPath))) {
		return PluginQualifiedSkillReadback{}, fmt.Errorf("official plugin base skill file missing: %s", skillPath)
	}
	return PluginQualifiedSkillReadback{
		Namespace:            pkg.Namespace,
		RequestedName:        requested,
		CanonicalSkillID:     canonical,
		PluginQualifiedName:  pkg.Namespace + ":" + canonical,
		SourcePackagePath:    pkg.ManifestPath,
		ResolvedSkillPath:    skillPath,
		PackageVersion:       pkg.Version,
		PackageSource:        pkg.ManifestPath,
		ProfileWrapperSource: false,
		FallbackUsed:         false,
	}, nil
}

func BuildSourceRoleManifestReadback(repo string) (SourceRoleManifestReadback, error) {
	sourceRepo, err := FindSourceRepo(repo)
	if err != nil {
		return SourceRoleManifestReadback{}, err
	}
	pkg, err := LoadPluginPackage(sourceRepo)
	if err != nil {
		return SourceRoleManifestReadback{}, err
	}
	if len(pkg.RoleManifests) == 0 {
		return SourceRoleManifestReadback{}, fmt.Errorf("official plugin package has no role manifests")
	}
	registered := stringSet(pkg.Skills)
	roles := make([]string, 0, len(pkg.RoleManifests))
	for role := range pkg.RoleManifests {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	readback := SourceRoleManifestReadback{Namespace: pkg.Namespace, PackageVersion: pkg.Version, PackageSource: pkg.ManifestPath}
	for _, role := range roles {
		manifestPath := pkg.RoleManifests[role]
		manifest, err := readPluginRoleManifest(sourceRepo, manifestPath)
		if err != nil {
			return SourceRoleManifestReadback{}, err
		}
		if manifest.Role != role {
			return SourceRoleManifestReadback{}, fmt.Errorf("role manifest %s declares role %q, want %q", manifestPath, manifest.Role, role)
		}
		for _, skillID := range manifest.SkillIDs {
			if !registered[skillID] {
				return SourceRoleManifestReadback{}, fmt.Errorf("role manifest %s selects unregistered plugin skill %s", manifestPath, skillID)
			}
		}
		readback.Roles = append(readback.Roles, RoleManifestReadback{
			Role:                 role,
			SourceControlledPath: manifestPath,
			Version:              manifest.Version,
			PackageSource:        pkg.ManifestPath,
			PackageVersion:       pkg.Version,
			SkillIDs:             manifest.SkillIDs,
		})
	}
	return readback, nil
}

type pluginRoleManifest struct {
	Version  string
	Role     string
	SkillIDs []string
}

func parsePluginPackageManifest(text string) PluginPackage {
	pkg := PluginPackage{RoleManifests: map[string]string{}}
	currentList := ""
	inRoles := false
	inPlugin := false
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent == 0 {
			currentList = ""
			inRoles = false
			inPlugin = false
			key, value, ok := splitYAMLLine(trimmed)
			if ok {
				switch key {
				case "name", "namespace":
					pkg.TopLevelName = value
				case "version":
					pkg.Version = value
				case "plugin":
					inPlugin = true
				case "skills", "guides":
					currentList = key
				}
			}
			if trimmed == "roles:" {
				inRoles = true
			}
			continue
		}
		if inPlugin && indent == 2 {
			key, value, ok := splitYAMLLine(trimmed)
			if ok {
				switch key {
				case "namespace":
					pkg.Namespace = value
				case "package_manifest":
					pkg.PackageManifest = value
				case "load_policy":
					pkg.LoadPolicy = value
				}
			}
			continue
		}
		if currentList != "" && indent == 2 && strings.HasPrefix(trimmed, "- ") {
			value := cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
			if currentList == "skills" {
				pkg.Skills = append(pkg.Skills, value)
			} else if currentList == "guides" {
				pkg.GuideSkills = append(pkg.GuideSkills, value)
			}
			continue
		}
		if inRoles && indent == 2 {
			key, value, ok := splitYAMLLine(trimmed)
			if ok && key != "" && value != "" {
				pkg.RoleManifests[key] = value
			}
		}
	}
	return pkg
}

func readPluginRoleManifest(sourceRepo string, relPath string) (pluginRoleManifest, error) {
	if IsInvalidRelativePath(relPath) {
		return pluginRoleManifest{}, fmt.Errorf("invalid role manifest path: %s", relPath)
	}
	data, err := os.ReadFile(filepath.Join(sourceRepo, filepath.FromSlash(relPath)))
	if err != nil {
		return pluginRoleManifest{}, fmt.Errorf("role manifest unreadable %s: %w", relPath, err)
	}
	manifest := pluginRoleManifest{}
	currentList := ""
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent == 0 {
			currentList = ""
			key, value, ok := splitYAMLLine(trimmed)
			if !ok {
				continue
			}
			switch key {
			case "version":
				manifest.Version = value
			case "role":
				manifest.Role = value
			case "skills":
				currentList = key
			}
			continue
		}
		if currentList == "skills" && indent == 2 && strings.HasPrefix(trimmed, "- ") {
			manifest.SkillIDs = append(manifest.SkillIDs, cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))))
		}
	}
	if manifest.Version != PluginRoleManifestVersion {
		return pluginRoleManifest{}, fmt.Errorf("unsupported role manifest version in %s: %q", relPath, manifest.Version)
	}
	if manifest.Role == "" {
		return pluginRoleManifest{}, fmt.Errorf("role manifest %s is missing role", relPath)
	}
	if len(manifest.SkillIDs) == 0 {
		return pluginRoleManifest{}, fmt.Errorf("role manifest %s has no skills", relPath)
	}
	sort.Strings(manifest.SkillIDs)
	return manifest, nil
}

func resolvePluginSkillID(skills []string, requested string) (string, bool) {
	for _, skill := range skills {
		if skill == requested {
			return skill, true
		}
	}
	if !strings.HasPrefix(requested, "kkachi-") {
		candidate := "kkachi-" + requested
		for _, skill := range skills {
			if skill == candidate {
				return skill, true
			}
		}
	}
	return "", false
}

func isCurrentPluginPackageManifest(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || filepath.IsAbs(value) || strings.Contains(value, "\\") {
		return false
	}
	return filepath.ToSlash(filepath.Clean(value)) == PluginPackageManifestPath
}

func invalidPluginSkillID(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || trimmed == "." || value != trimmed || strings.Contains(trimmed, "/") || IsInvalidRelativePath(trimmed)
}

func malformedPluginQualifiedComponent(value string) bool {
	return value == "" || value != strings.TrimSpace(value) || strings.Contains(value, "/") || strings.Contains(value, "\\")
}

func splitYAMLLine(trimmed string) (string, string, bool) {
	key, value, ok := strings.Cut(trimmed, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), cleanYAMLScalar(value), true
}

func cleanYAMLScalar(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

func stringSet(values []string) map[string]bool {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	return set
}
