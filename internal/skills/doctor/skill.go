package doctor

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
)

const skillDoctorMode = "skill_plugin_overlay_doctor"

type SkillOptions struct {
	Profile     string
	Project     string
	ProfileRoot string
}

type SkillNoWriteEvidence struct {
	Guaranteed                   bool `json:"guaranteed"`
	ProfileWrapperWriteCount     int  `json:"profile_wrapper_write_count"`
	ProjectOverlayWriteCount     int  `json:"project_overlay_write_count"`
	CopiedLegacySuiteWriteCount  int  `json:"copied_legacy_suite_write_count"`
	KAHStateWriteCount           int  `json:"kah_state_write_count"`
	KABRuntimeMutationCount      int  `json:"kab_runtime_mutation_count"`
	HermesRuntimeMutationCount   int  `json:"hermes_runtime_mutation_count"`
	AuthProviderConfigWriteCount int  `json:"auth_provider_config_write_count"`
	ProfileActivationCount       int  `json:"profile_activation_count"`
}

type SkillSourceRecord struct {
	SourceClass string   `json:"source_class"`
	Name        string   `json:"name"`
	Path        string   `json:"path,omitempty"`
	Status      string   `json:"status"`
	Role        string   `json:"role,omitempty"`
	Project     string   `json:"project,omitempty"`
	OverlayFor  string   `json:"overlay_for,omitempty"`
	BaseVersion string   `json:"base_version,omitempty"`
	ReasonCodes []string `json:"reason_codes,omitempty"`
}

type SkillDoctorSummary struct {
	CountsBySourceClass map[string]int `json:"counts_by_source_class"`
	ErrorCount          int            `json:"error_count"`
	WarningCount        int            `json:"warning_count"`
}

type SkillDoctorResult struct {
	OK                         bool                                       `json:"ok"`
	Command                    string                                     `json:"command"`
	Mode                       string                                     `json:"mode"`
	CLIVersion                 string                                     `json:"cli_version"`
	ProvenanceContractVersion  string                                     `json:"provenance_contract_version"`
	NoWrite                    SkillNoWriteEvidence                       `json:"no_write"`
	SourceRepo                 SourceRepo                                 `json:"source_repo"`
	Plugin                     SkillPluginEvidence                        `json:"plugin"`
	TargetProfile              TargetProfile                              `json:"target_profile,omitempty"`
	Project                    string                                     `json:"project,omitempty"`
	SourceClasses              []SkillSourceRecord                        `json:"source_classes"`
	SourceClassEvidence        []discovery.SourceClassEvidence            `json:"source_class_evidence"`
	DependencyAudit            discovery.DependencyAudit                  `json:"dependency_audit"`
	SkillDependencies          []discovery.SkillDependencyRecord          `json:"skill_dependencies"`
	CommandSurfaceDependencies []discovery.CommandSurfaceDependencyRecord `json:"command_surface_dependencies"`
	DeletedBundleReference     any                                        `json:"deleted_bundle_reference"`
	DeletedBundleDiagnostics   []discovery.Diagnostic                     `json:"deleted_bundle_diagnostics"`
	Summary                    SkillDoctorSummary                         `json:"summary"`
	Diagnostics                []discovery.Diagnostic                     `json:"diagnostics"`
	ReasonCodes                []string                                   `json:"reason_codes"`
	NextAction                 string                                     `json:"next_action"`
}

type SkillPluginEvidence struct {
	Namespace      string   `json:"namespace,omitempty"`
	Version        string   `json:"version,omitempty"`
	ManifestPath   string   `json:"manifest_path,omitempty"`
	LoadPolicy     string   `json:"load_policy,omitempty"`
	RoleManifests  []string `json:"role_manifests"`
	GuideSkills    []string `json:"guide_skills"`
	RegisteredBase []string `json:"registered_base"`
	Status         string   `json:"status"`
}

type skillMetadata struct {
	Name            string
	Kind            string
	Role            string
	RoleManifest    string
	PluginNamespace string
	OverlayRoot     string
	Project         string
	OverlayFor      string
	MergeMode       string
	BaseVersion     string
}

func BuildSkill(repo string, opts SkillOptions) (SkillDoctorResult, error) {
	sourceRepoPath, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return SkillDoctorResult{}, err
	}
	sourceInfo := discovery.SourceRepoInfo(sourceRepoPath)
	result := SkillDoctorResult{
		OK:                        true,
		Command:                   "doctor",
		Mode:                      skillDoctorMode,
		CLIVersion:                version.CLIVersion,
		ProvenanceContractVersion: discovery.ProvenanceContractVersion,
		NoWrite:                   SkillNoWriteEvidence{Guaranteed: true},
		SourceRepo: SourceRepo{
			Path:      sourceInfo.Path,
			State:     "ok",
			GitCommit: sourceInfo.GitCommit,
			Dirty:     sourceInfo.Dirty,
		},
		Plugin:                     SkillPluginEvidence{Status: "unknown"},
		Project:                    opts.Project,
		SourceClasses:              []SkillSourceRecord{},
		SourceClassEvidence:        []discovery.SourceClassEvidence{},
		DependencyAudit:            discovery.EmptyDependencyAudit(),
		SkillDependencies:          []discovery.SkillDependencyRecord{},
		CommandSurfaceDependencies: []discovery.CommandSurfaceDependencyRecord{},
		DeletedBundleReference:     nil,
		DeletedBundleDiagnostics:   []discovery.Diagnostic{},
		Diagnostics:                []discovery.Diagnostic{},
		ReasonCodes:                []string{},
		Summary: SkillDoctorSummary{CountsBySourceClass: map[string]int{
			"plugin_base":              0,
			"color_wrapper":            0,
			"project_overlay":          0,
			"legacy_copied_base_suite": 0,
			"personal_skill":           0,
			"unknown_source":           0,
		}},
		NextAction: "Review SKILL doctor diagnostics before any migration, cleanup, repair, apply, or profile mutation.",
	}

	sourcePacks, err := discovery.DiscoverSourcePacks(sourceRepoPath)
	if err != nil {
		addSkillDiag(&result, "error", "source_inventory_unavailable", "source inventory evidence is unavailable: "+err.Error())
	} else {
		inventory := discovery.BuildSourceInventory(sourcePacks, nil, nil)
		result.DependencyAudit = discovery.BuildDependencyAudit(sourceRepoPath, sourcePacks, inventory)
		result.SkillDependencies = append([]discovery.SkillDependencyRecord{}, result.DependencyAudit.SkillDependencies...)
		result.CommandSurfaceDependencies = append([]discovery.CommandSurfaceDependencyRecord{}, result.DependencyAudit.CommandSurfaceDependencies...)
		result.DeletedBundleDiagnostics = append([]discovery.Diagnostic{}, result.DependencyAudit.DeletedBundleDiagnostics...)
		for _, pack := range sourcePacks {
			record := discovery.SourceOnlyProvenance(pack)
			result.SourceClassEvidence = append(result.SourceClassEvidence, record.SourceClassEvidence...)
		}
		sort.Slice(result.SourceClassEvidence, func(i, j int) bool {
			if result.SourceClassEvidence[i].Path == result.SourceClassEvidence[j].Path {
				return result.SourceClassEvidence[i].Kind < result.SourceClassEvidence[j].Kind
			}
			return result.SourceClassEvidence[i].Path < result.SourceClassEvidence[j].Path
		})
	}

	pkg, err := discovery.LoadPluginPackage(sourceRepoPath)
	if err != nil {
		result.Plugin.Status = "missing_or_invalid"
		addSkillDiag(&result, "error", "missing_plugin_evidence", "official KAS plugin package evidence is missing or invalid: "+err.Error())
	} else {
		result.Plugin = SkillPluginEvidence{
			Namespace:      pkg.Namespace,
			Version:        pkg.Version,
			ManifestPath:   pkg.ManifestPath,
			LoadPolicy:     pkg.LoadPolicy,
			RoleManifests:  mapValuesSorted(pkg.RoleManifests),
			GuideSkills:    append([]string(nil), pkg.GuideSkills...),
			RegisteredBase: append([]string(nil), pkg.Skills...),
			Status:         "ok",
		}
		for _, skillID := range pkg.Skills {
			relPath := filepath.ToSlash(filepath.Join("skills", skillID, "SKILL.md"))
			record := SkillSourceRecord{SourceClass: "plugin_base", Name: skillID, Path: relPath, Status: "ok"}
			if st, err := os.Stat(filepath.Join(sourceRepoPath, filepath.FromSlash(relPath))); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					record.Status = "missing"
					record.ReasonCodes = append(record.ReasonCodes, "missing_plugin_base_skill")
					addSkillDiag(&result, "error", "missing_plugin_base_skill", "registered plugin base skill file is missing: "+relPath)
				} else {
					record.Status = "unreadable"
					record.ReasonCodes = append(record.ReasonCodes, "missing_plugin_base_skill")
					addSkillDiag(&result, "error", "missing_plugin_base_skill", "registered plugin base skill file is unreadable: "+relPath)
				}
			} else if st.IsDir() {
				record.Status = "invalid"
				record.ReasonCodes = append(record.ReasonCodes, "missing_plugin_base_skill")
				addSkillDiag(&result, "error", "missing_plugin_base_skill", "registered plugin base skill path is not a file: "+relPath)
			}
			result.SourceClasses = append(result.SourceClasses, record)
		}
		addSkillDiag(&result, "info", "plugin_update_readiness_readback_only", "plugin package readiness is based on source readback only; run update plugin --dry-run for update planning before any approved apply.")
		addSkillDiag(&result, "warning", "update_surface_legacy_alias_present", "bare update remains a legacy project-suite alias; use update plugin for plugin package planning and update project-suite for project-suite lifecycle.")
	}

	if _, err := discovery.BuildSourceRoleManifestReadback(sourceRepoPath); err != nil {
		addSkillDiag(&result, "error", "missing_role_evidence", "official plugin role manifest evidence is missing or invalid: "+err.Error())
	}

	if opts.Profile == "" && opts.ProfileRoot == "" {
		addSkillDiag(&result, "warning", "missing_wrapper_evidence", "no profile was provided, so profile-local wrapper, overlay, copied-suite, personal, and unknown source evidence was not inspected.")
		finalizeSkillDoctor(&result)
		return result, nil
	}

	profileRoot := resolveProfileRoot(opts.Profile, opts.ProfileRoot)
	result.TargetProfile = TargetProfile{
		Name:  opts.Profile,
		Root:  profileRoot,
		State: "ok",
	}
	if st, err := os.Stat(profileRoot); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			result.TargetProfile.State = "missing"
			addSkillDiag(&result, "error", "profile_missing", "profile root does not exist: "+profileRoot)
		} else {
			result.TargetProfile.State = "unreadable"
			addSkillDiag(&result, "error", "profile_unreadable", "profile root is not readable: "+err.Error())
		}
		finalizeSkillDoctor(&result)
		return result, nil
	} else if !st.IsDir() {
		result.TargetProfile.State = "invalid"
		addSkillDiag(&result, "error", "profile_not_directory", "profile root is not a directory: "+profileRoot)
		finalizeSkillDoctor(&result)
		return result, nil
	}

	classifyProfileSkills(&result, sourceRepoPath, profileRoot, pkg)
	finalizeSkillDoctor(&result)
	return result, nil
}

func classifyProfileSkills(result *SkillDoctorResult, sourceRepo string, profileRoot string, pkg discovery.PluginPackage) {
	pluginSkillSet := map[string]bool{}
	for _, skillID := range pkg.Skills {
		pluginSkillSet[skillID] = true
	}
	roleManifestRoles := roleNamesFromPluginPackage(pkg)
	wrapperCount := 0
	skillsRoot := filepath.Join(profileRoot, "skills")
	if st, err := os.Stat(skillsRoot); err != nil || !st.IsDir() {
		addSkillDiag(result, "error", "missing_wrapper_evidence", "profile skills root is missing or unreadable; wrapper evidence cannot be inspected: "+skillsRoot)
		return
	}
	if err := filepath.WalkDir(skillsRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "SKILL.md" {
			return nil
		}
		rel, relErr := filepath.Rel(profileRoot, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		dirRel := filepath.ToSlash(rel)
		skillID := filepath.Base(filepath.Dir(path))
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			result.SourceClasses = append(result.SourceClasses, SkillSourceRecord{SourceClass: "unknown_source", Name: skillID, Path: filepath.ToSlash(filepath.Join(dirRel, "SKILL.md")), Status: "unreadable", ReasonCodes: []string{"skill_unreadable"}})
			addSkillDiag(result, "error", "unknown_source_unreadable", "skill file is unreadable: "+filepath.ToSlash(filepath.Join(dirRel, "SKILL.md")))
			return nil
		}
		meta := parseSkillMetadata(data)
		if meta.Name == "" {
			meta.Name = skillID
		}
		pathForJSON := filepath.ToSlash(filepath.Join(dirRel, "SKILL.md"))
		if isOverlayPath(dirRel) || meta.Kind == "project_overlay" {
			if result.Project != "" && overlayOutsideRequestedProject(dirRel, meta, result.Project) {
				return filepath.SkipDir
			}
			classifyOverlay(result, sourceRepo, pathForJSON, dirRel, data, meta, pkg)
			return nil
		}
		if meta.Kind == "color_wrapper" {
			wrapperCount++
			record := SkillSourceRecord{SourceClass: "color_wrapper", Name: meta.Name, Path: pathForJSON, Status: "ok", Role: meta.Role}
			validateWrapper(result, &record, meta, pkg, roleManifestRoles)
			result.SourceClasses = append(result.SourceClasses, record)
			return filepath.SkipDir
		}
		if pluginSkillSet[skillID] {
			result.SourceClasses = append(result.SourceClasses, SkillSourceRecord{SourceClass: "legacy_copied_base_suite", Name: skillID, Path: pathForJSON, Status: "legacy_copy", ReasonCodes: []string{"legacy_copied_base_suite_present", "profile_skill_shadows_plugin_base"}})
			addSkillDiag(result, "warning", "legacy_copied_base_suite_present", "profile-local copied KAS base suite is present and must not be used as plugin fallback: "+dirRel)
			addSkillDiag(result, "error", "profile_skill_shadows_plugin_base", "profile-local skill shadows an official plugin base skill: "+skillID)
			return filepath.SkipDir
		}
		if meta.Kind != "" && strings.HasPrefix(meta.Kind, "kas_") {
			result.SourceClasses = append(result.SourceClasses, SkillSourceRecord{SourceClass: "unknown_source", Name: meta.Name, Path: pathForJSON, Status: "unknown", ReasonCodes: []string{"unknown_kas_skill_kind"}})
			addSkillDiag(result, "error", "unknown_source_ambiguous", "KAS-like skill kind is not recognized by SKILL doctor: "+pathForJSON)
			return filepath.SkipDir
		}
		result.SourceClasses = append(result.SourceClasses, SkillSourceRecord{SourceClass: "personal_skill", Name: meta.Name, Path: pathForJSON, Status: "ok"})
		return filepath.SkipDir
	}); err != nil {
		addSkillDiag(result, "error", "profile_unreadable", "profile skill inventory walk failed: "+err.Error())
	}
	if wrapperCount == 0 {
		addSkillDiag(result, "error", "missing_wrapper_evidence", "no profile-local color wrapper was found under profile skills.")
	}
}

func overlayOutsideRequestedProject(dirRel string, meta skillMetadata, requestedProject string) bool {
	if requestedProject == "" {
		return false
	}
	pathProject := ""
	parts := strings.Split(filepath.ToSlash(dirRel), "/")
	if len(parts) >= 2 && parts[0] == "skills" {
		pathProject = parts[1]
	}
	if pathProject != "" && pathProject != requestedProject {
		return true
	}
	return meta.Project != "" && meta.Project != requestedProject
}

func classifyOverlay(result *SkillDoctorResult, sourceRepo string, pathForJSON string, dirRel string, data []byte, meta skillMetadata, pkg discovery.PluginPackage) {
	record := SkillSourceRecord{SourceClass: "project_overlay", Name: meta.Name, Path: pathForJSON, Status: "ok", Role: meta.Role, Project: meta.Project, OverlayFor: meta.OverlayFor, BaseVersion: meta.BaseVersion}
	parts := strings.Split(dirRel, "/")
	if len(parts) < 4 || parts[0] != "skills" || parts[2] != "kas-overlays" {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "invalid_overlay_layout")
		addSkillDiag(result, "error", "invalid_overlay_layout", "project overlay must live under skills/<project>/kas-overlays/<overlay>/SKILL.md: "+pathForJSON)
	}
	if meta.Kind != "project_overlay" {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "invalid_overlay_frontmatter")
		addSkillDiag(result, "error", "invalid_overlay_frontmatter", "project overlay is missing metadata.kas.kind: project_overlay: "+pathForJSON)
	}
	if meta.Project == "" || (len(parts) >= 2 && parts[1] != "" && meta.Project != "" && meta.Project != parts[1]) {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "invalid_overlay_frontmatter")
		addSkillDiag(result, "error", "invalid_overlay_frontmatter", "project overlay metadata.kas.project is missing or does not match its path: "+pathForJSON)
	}
	if result.Project != "" && meta.Project != "" && meta.Project != result.Project {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "project_overlay_out_of_scope")
		addSkillDiag(result, "error", "project_overlay_out_of_scope", "project overlay does not match requested --project: "+pathForJSON)
	}
	if meta.MergeMode == "" {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "invalid_overlay_frontmatter")
		addSkillDiag(result, "error", "invalid_overlay_frontmatter", "project overlay metadata.kas.merge_mode is missing: "+pathForJSON)
	}
	if !pluginQualifiedOverlayTarget(meta.OverlayFor, pkg) {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "invalid_overlay_frontmatter")
		addSkillDiag(result, "error", "invalid_overlay_frontmatter", "project overlay overlay_for must name a registered plugin-qualified KAS base skill: "+pathForJSON)
	}
	if shadowsPluginBase(pkg, meta.Name, filepath.Base(filepath.Dir(pathForJSON))) {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "project_overlay_shadows_plugin_base")
		addSkillDiag(result, "error", "project_overlay_shadows_plugin_base", "project overlay name or directory shadows a registered plugin base skill: "+pathForJSON)
	}
	if meta.BaseVersion != "" && baseVersionStale(meta.BaseVersion, pkg.Version, result.SourceRepo.GitCommit) {
		record.Status = worstSkillStatus(record.Status, "stale")
		record.ReasonCodes = append(record.ReasonCodes, "stale_base_version")
		addSkillDiag(result, "warning", "stale_base_version", "project overlay base_version does not match current plugin package version or source commit evidence: "+pathForJSON)
	}
	if containsRuntimeConfig(data) {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "overlay_runtime_config_boundary_violation")
		addSkillDiag(result, "error", "overlay_runtime_config_boundary_violation", "project overlay appears to contain auth/token/gateway/provider/model/runtime configuration text: "+pathForJSON)
	}
	if containsKAHOwnershipViolation(data) {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "kah_boundary_violation")
		addSkillDiag(result, "error", "kah_boundary_violation", "project overlay appears to assign KAS plugin/wrapper/update ownership to KAH: "+pathForJSON)
	}
	if looksLikeCopiedBase(data) {
		record.Status = worstSkillStatus(record.Status, "suspicious")
		record.ReasonCodes = append(record.ReasonCodes, "overlay_copies_base_body")
		addSkillDiag(result, "warning", "overlay_copies_base_body", "project overlay appears large enough to be a copied base body; review before use: "+pathForJSON)
	}
	result.SourceClasses = append(result.SourceClasses, record)
}

func validateWrapper(result *SkillDoctorResult, record *SkillSourceRecord, meta skillMetadata, pkg discovery.PluginPackage, roles map[string]bool) {
	if meta.PluginNamespace == "" || meta.PluginNamespace != pkg.Namespace {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "missing_plugin_evidence")
		addSkillDiag(result, "error", "missing_plugin_evidence", "color wrapper plugin_namespace is missing or does not match official plugin namespace: "+record.Path)
	}
	roleFromManifest := roleNameFromQualifiedRoleManifest(meta.RoleManifest, pkg.Namespace)
	if roleFromManifest == "" || !roles[roleFromManifest] {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "missing_role_evidence")
		addSkillDiag(result, "error", "missing_role_evidence", "color wrapper role_manifest is missing or does not name a known source role manifest: "+record.Path)
	}
	if meta.Role != "" && roleFromManifest != "" && !wrapperRoleMatchesManifest(meta.Role, roleFromManifest) {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "wrapper_role_manifest_mismatch")
		addSkillDiag(result, "error", "wrapper_role_manifest_mismatch", "color wrapper metadata.kas.role does not match its role_manifest role: "+record.Path)
	}
	if meta.OverlayRoot == "" {
		record.Status = "invalid"
		record.ReasonCodes = append(record.ReasonCodes, "missing_wrapper_evidence")
		addSkillDiag(result, "error", "missing_wrapper_evidence", "color wrapper overlay_root evidence is missing: "+record.Path)
	}
}

func RenderHumanSkill(result SkillDoctorResult) string {
	state := "pass"
	if !result.OK {
		state = "error"
	} else if result.Summary.WarningCount > 0 {
		state = "warning"
	}
	lines := []string{
		fmt.Sprintf("Status: %s; SKILL plugin/wrapper/overlay doctor.", state),
		fmt.Sprintf("Summary: source classes %d, warnings %d, errors %d.", len(result.SourceClasses), result.Summary.WarningCount, result.Summary.ErrorCount),
		"Mode: " + result.Mode,
		fmt.Sprintf("Writes: no-write guaranteed=%t; profile/wrapper/overlay/KAH/KAB/auth/provider/model writes 0.", result.NoWrite.Guaranteed),
		"Plugin: " + result.Plugin.Status,
	}
	if result.TargetProfile.Root != "" {
		lines = append(lines, "Profile: "+result.TargetProfile.State+" ("+result.TargetProfile.Root+")")
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, fmt.Sprintf("Diagnostic[%s:%s]: %s", diagnostic.Level, diagnostic.Code, diagnostic.Message))
	}
	lines = append(lines, "Next: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func parseSkillMetadata(data []byte) skillMetadata {
	text := string(data)
	lines := strings.Split(text, "\n")
	meta := skillMetadata{}
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return meta
	}
	for i := 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "---" {
			break
		}
		key, value, ok := splitSkillYAMLLine(trimmed)
		if !ok {
			continue
		}
		switch key {
		case "name":
			meta.Name = value
		case "kind":
			meta.Kind = value
		case "role":
			meta.Role = value
		case "role_manifest":
			meta.RoleManifest = value
		case "plugin_namespace":
			meta.PluginNamespace = value
		case "overlay_root":
			meta.OverlayRoot = value
		case "project":
			meta.Project = value
		case "overlay_for":
			meta.OverlayFor = value
		case "merge_mode":
			meta.MergeMode = value
		case "base_version":
			meta.BaseVersion = value
		}
	}
	return meta
}

func splitSkillYAMLLine(trimmed string) (string, string, bool) {
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "- ") {
		return "", "", false
	}
	key, value, ok := strings.Cut(trimmed, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), `"'`), true
}

func isOverlayPath(dirRel string) bool {
	parts := strings.Split(filepath.ToSlash(dirRel), "/")
	return len(parts) >= 4 && parts[0] == "skills" && parts[1] != "" && parts[2] == "kas-overlays"
}

func pluginQualifiedOverlayTarget(value string, pkg discovery.PluginPackage) bool {
	namespace, requested, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok || namespace != pkg.Namespace || requested == "" {
		return false
	}
	if strings.Contains(requested, "/") || strings.Contains(requested, "\\") || strings.Contains(requested, ":") {
		return false
	}
	for _, skillID := range pkg.Skills {
		if requested == skillID || (!strings.HasPrefix(requested, "kkachi-") && "kkachi-"+requested == skillID) {
			return true
		}
	}
	return false
}

func shadowsPluginBase(pkg discovery.PluginPackage, values ...string) bool {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		for _, skillID := range pkg.Skills {
			if value == skillID {
				return true
			}
		}
	}
	return false
}

func wrapperRoleMatchesManifest(wrapperRole string, manifestRole string) bool {
	wrapperRole = strings.TrimSpace(wrapperRole)
	manifestRole = strings.TrimSpace(manifestRole)
	return wrapperRole == manifestRole || strings.HasPrefix(wrapperRole, manifestRole+"_")
}

func baseVersionStale(baseVersion string, packageVersion string, gitCommit *string) bool {
	if baseVersion == "" {
		return false
	}
	if packageVersion != "" && baseVersion == packageVersion {
		return false
	}
	if gitCommit != nil && *gitCommit != "" && strings.HasPrefix(*gitCommit, baseVersion) {
		return false
	}
	return true
}

func containsRuntimeConfig(data []byte) bool {
	text := strings.ToLower(string(data))
	for _, token := range []string{"auth_token", "api_key", "gateway_url", "provider_model", "model_config", "kab runtime", "hermes runtime activation"} {
		if strings.Contains(text, token) {
			return true
		}
	}
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "- ") {
			continue
		}
		key, _, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "provider", "model", "auth_token", "api_key", "gateway_url", "provider_model", "model_config":
			return true
		}
	}
	return false
}

func containsKAHOwnershipViolation(data []byte) bool {
	text := strings.ToLower(string(data))
	for _, token := range []string{"kah owns plugin", "kah owns kas plugin", "kah installs plugin", "kah updates plugin", "kah mutates plugin"} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func looksLikeCopiedBase(data []byte) bool {
	text := string(data)
	return len(data) > 5000 || strings.Count(text, "\n## ") > 8
}

func roleNamesFromPluginPackage(pkg discovery.PluginPackage) map[string]bool {
	roles := map[string]bool{}
	for role := range pkg.RoleManifests {
		roles[role] = true
	}
	return roles
}

func roleNameFromQualifiedRoleManifest(value string, namespace string) string {
	prefix := namespace + ":roles/"
	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, ".yaml") {
		return ""
	}
	role := strings.TrimSuffix(strings.TrimPrefix(value, prefix), ".yaml")
	if role == "" || strings.Contains(role, "/") || strings.Contains(role, "\\") {
		return ""
	}
	return role
}

func mapValuesSorted(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func finalizeSkillDoctor(result *SkillDoctorResult) {
	sort.Slice(result.SourceClasses, func(i, j int) bool {
		if result.SourceClasses[i].SourceClass == result.SourceClasses[j].SourceClass {
			return result.SourceClasses[i].Path < result.SourceClasses[j].Path
		}
		return result.SourceClasses[i].SourceClass < result.SourceClasses[j].SourceClass
	})
	if result.Summary.CountsBySourceClass == nil {
		result.Summary.CountsBySourceClass = map[string]int{}
	}
	for _, record := range result.SourceClasses {
		result.Summary.CountsBySourceClass[record.SourceClass]++
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Level == "error" {
			result.Summary.ErrorCount++
		}
		if diagnostic.Level == "warning" {
			result.Summary.WarningCount++
		}
	}
	result.OK = result.Summary.ErrorCount == 0
	if result.OK {
		result.NextAction = "SKILL doctor found no blocking plugin/wrapper/overlay diagnostics. Keep migration, cleanup, repair, apply, and profile mutation separately approved."
	} else {
		result.NextAction = "Fix the reported SKILL plugin/wrapper/overlay diagnostics, then rerun doctor --plugin. Do not fall back to copied profile-local KAS suites."
	}
	result.ReasonCodes = uniqueSkillReasonCodes(result.Diagnostics)
}

func addSkillDiag(result *SkillDoctorResult, level string, code string, message string) {
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: level, Code: code, Message: message})
}

func uniqueSkillReasonCodes(diagnostics []discovery.Diagnostic) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == "" || seen[diagnostic.Code] {
			continue
		}
		seen[diagnostic.Code] = true
		out = append(out, diagnostic.Code)
	}
	sort.Strings(out)
	return out
}

func worstSkillStatus(current string, next string) string {
	if current == "invalid" || next == "invalid" {
		return "invalid"
	}
	if current == "stale" || next == "stale" {
		return "stale"
	}
	if current == "suspicious" || next == "suspicious" {
		return "suspicious"
	}
	if current == "" {
		return next
	}
	return current
}

func treeSHA256(root string) (string, error) {
	entries := []string{}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		entries = append(entries, filepath.ToSlash(rel)+"\x00"+hex.EncodeToString(sum[:]))
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(entries)
	sum := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(sum[:]), nil
}
