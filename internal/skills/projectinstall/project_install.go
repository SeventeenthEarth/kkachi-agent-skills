package projectinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/install"
)

const (
	VirtualSourcePackID = "kas-default-project-suite"
	ManifestVersion     = "0.1"
	ManifestKind        = "kas_project_skill_manifest"
	ProfileManifestKind = install.ManifestKind
)

var validProjectID = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)

type Options struct {
	Profile     string
	Project     string
	SuiteRole   string
	SourcePack  string
	ProfileRoot string
	DryRun      bool
}

type NoWriteEvidence struct {
	Guaranteed                   bool `json:"guaranteed"`
	ProfileWriteCount            int  `json:"profile_write_count"`
	SkillWriteCount              int  `json:"skill_write_count"`
	ManifestWriteCount           int  `json:"manifest_write_count"`
	KASDirectoryWriteCount       int  `json:"kas_directory_write_count"`
	KAHStateWriteCount           int  `json:"kah_state_write_count"`
	KABRuntimeMutationCount      int  `json:"kab_runtime_mutation_count"`
	AuthProviderConfigWriteCount int  `json:"auth_provider_config_write_count"`
}

type Project struct {
	ID              string `json:"id"`
	TargetSuitePath string `json:"target_suite_path"`
}

type SourcePack struct {
	ID             string               `json:"id"`
	Source         string               `json:"source,omitempty"`
	ResolvedFrom   string               `json:"resolved_from"`
	SourceRepo     discovery.SourceRepo `json:"source_repo"`
	PackChecksums  map[string]string    `json:"pack_checksums"`
	SuiteChecksum  string               `json:"suite_checksum"`
	FormalRegistry string               `json:"formal_registry"`
}

type ProjectTailoring struct {
	Mode                                      string   `json:"mode"`
	Preserves                                 []string `json:"preserves"`
	Source                                    string   `json:"source"`
	SemanticPortRequiredBeforeApprovedInstall bool     `json:"semantic_port_required_before_approved_install"`
	SemanticAdaptationClaimed                 bool     `json:"semantic_adaptation_claimed"`
}

type Summary struct {
	TotalSkills     int            `json:"total_skills"`
	TotalFiles      int            `json:"total_files"`
	SelectedSkills  int            `json:"selected_skills"`
	ExcludedSkills  int            `json:"excluded_skills"`
	CountsByAction  map[string]int `json:"counts_by_action"`
	ConflictCount   int            `json:"conflict_count"`
	DiagnosticCount int            `json:"diagnostic_count"`
}

type PlannedSkill struct {
	SourcePackID   string `json:"source_pack_id"`
	SourceSkill    string `json:"source_skill"`
	InstalledSkill string `json:"installed_skill"`
	TargetPath     string `json:"target_path"`
	DriftPolicy    string `json:"drift_policy"`
	Checksum       string `json:"checksum"`
	Action         string `json:"action"`
	Bytes          int    `json:"bytes"`
	TailoringMode  string `json:"tailoring_mode"`
}

type PlannedCompositionFile struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	TargetPath    string `json:"target_path"`
	DriftPolicy   string `json:"drift_policy"`
	Checksum      string `json:"checksum"`
	Action        string `json:"action"`
	Bytes         int    `json:"bytes"`
	TailoringMode string `json:"tailoring_mode"`
}

type ChangedPath struct {
	Path           string  `json:"path"`
	Action         string  `json:"action"`
	InstalledSkill string  `json:"installed_skill"`
	SourcePackID   string  `json:"source_pack_id"`
	PreviousSHA256 *string `json:"previous_sha256"`
	NewSHA256      string  `json:"new_sha256,omitempty"`
	Bytes          int     `json:"bytes,omitempty"`
	ErrorCode      string  `json:"error_code,omitempty"`
	ErrorMessage   string  `json:"error_message,omitempty"`
	BackupPath     string  `json:"backup_path,omitempty"`
}

type BackupEntry struct {
	Path           string `json:"path"`
	BackupPath     string `json:"backup_path"`
	PreviousSHA256 string `json:"previous_sha256"`
	Bytes          int    `json:"bytes"`
}

type ApprovalRequest struct {
	Required                      bool   `json:"required"`
	EvidenceRef                   string `json:"evidence_ref"`
	DryRunPlanHash                string `json:"dry_run_plan_hash"`
	HashIncludesProfile           bool   `json:"hash_includes_profile"`
	HashIncludesManifestState     bool   `json:"hash_includes_manifest_state"`
	HashIncludesSourceSuite       bool   `json:"hash_includes_source_suite"`
	HashIncludesRoleFields        bool   `json:"hash_includes_role_fields"`
	HashIncludesNoWriteEvidence   bool   `json:"hash_includes_no_write_evidence"`
	HashIncludesBackupPlan        bool   `json:"hash_includes_backup_plan"`
	HashIncludesConflictsAndDiags bool   `json:"hash_includes_conflicts_and_diagnostics"`
}

type ApprovalEvidence struct {
	EvidenceRef        string `json:"evidence_ref"`
	DryRunPlanHash     string `json:"dry_run_plan_hash"`
	ApprovedPlanHash   string `json:"approved_plan_hash"`
	MatchedCurrentPlan bool   `json:"matched_current_plan"`
}

type Recovery struct {
	RollbackSupported      bool     `json:"rollback_supported"`
	BackupPath             string   `json:"backup_path"`
	PreviousManifestSHA256 *string  `json:"previous_manifest_sha256,omitempty"`
	Instructions           []string `json:"instructions"`
}

type Checksums struct {
	SourcePack      string `json:"source_pack"`
	PlannedManifest string `json:"planned_manifest"`
	PlannedSkills   string `json:"planned_skills"`
	ChangedPaths    string `json:"changed_paths"`
}

type Conflict struct {
	Severity       string `json:"severity"`
	Condition      string `json:"condition"`
	Project        string `json:"project"`
	InstalledSkill string `json:"installed_skill,omitempty"`
	TargetPath     string `json:"target_path,omitempty"`
	Message        string `json:"message"`
	NextAction     string `json:"next_action"`
}

type Result struct {
	OK               bool                     `json:"ok"`
	Command          string                   `json:"command"`
	Mode             string                   `json:"mode"`
	CLIVersion       string                   `json:"cli_version"`
	DryRun           bool                     `json:"dry_run"`
	NoWrite          NoWriteEvidence          `json:"no_write"`
	SourceRepo       discovery.SourceRepo     `json:"source_repo"`
	TargetProfile    discovery.TargetProfile  `json:"target_profile"`
	Project          Project                  `json:"project"`
	SourcePack       SourcePack               `json:"source_pack"`
	SuiteRole        string                   `json:"suite_role"`
	SuiteMode        string                   `json:"suite_mode"`
	RoleLabel        string                   `json:"role_label"`
	RoleRegistry     RoleRegistryEvidence     `json:"role_registry"`
	SelectedSkills   []RoleSkillEvidence      `json:"selected_skills"`
	ExcludedSkills   []RoleSkillEvidence      `json:"excluded_skills"`
	ProjectTailoring ProjectTailoring         `json:"project_tailoring"`
	Summary          Summary                  `json:"summary"`
	PlannedManifest  map[string]any           `json:"planned_manifest"`
	PlannedSkills    []PlannedSkill           `json:"planned_skills"`
	CompositionFiles []PlannedCompositionFile `json:"composition_files"`
	ChangedPaths     []ChangedPath            `json:"changed_paths"`
	BackupPlan       []BackupEntry            `json:"backup_plan"`
	Checksums        Checksums                `json:"checksums"`
	PlanHash         string                   `json:"plan_hash"`
	ApprovalRequest  ApprovalRequest          `json:"approval_request"`
	Approval         ApprovalEvidence         `json:"approval,omitempty"`
	InstallID        string                   `json:"install_id,omitempty"`
	ManifestPath     string                   `json:"manifest_path,omitempty"`
	BackupPath       string                   `json:"backup_path,omitempty"`
	Recovery         *Recovery                `json:"recovery,omitempty"`
	Conflicts        []Conflict               `json:"conflicts"`
	Diagnostics      []discovery.Diagnostic   `json:"diagnostics"`
	NextAction       string                   `json:"next_action"`
}

func BuildDryRun(repo string, opts Options) (Result, error) {
	sourceRepo, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return Result{}, err
	}
	profileRoot := resolveProfileRoot(opts.Profile, opts.ProfileRoot)
	result := baseResult(sourceRepo, profileRoot, opts)

	if !opts.DryRun {
		addDiagnostic(&result, "error", "dry_run_or_approve_required", "install-project-kas requires --dry-run or --approve dry-run:<hash>.")
		finalize(&result)
		return result, nil
	}
	if opts.ProfileRoot == "" {
		if st, err := os.Stat(profileRoot); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				addConflict(&result, "unknown_profile", "", ".", "unknown Hermes profile: "+opts.Profile, "Create or select a Hermes profile before project install.")
				finalize(&result)
				return result, nil
			}
			addConflict(&result, "profile_unreadable", "", ".", "Hermes profile is not readable: "+err.Error(), "Fix profile permissions and rerun dry-run.")
			finalize(&result)
			return result, nil
		} else if !st.IsDir() {
			addConflict(&result, "profile_not_directory", "", ".", "Hermes profile root is not a directory: "+profileRoot, "Use a valid Hermes profile root.")
			finalize(&result)
			return result, nil
		}
	}
	if !validProjectID.MatchString(opts.Project) {
		addConflict(&result, "invalid_project_id", "", "", fmt.Sprintf("project id %q is not a safe project suite id", opts.Project), "Use lowercase letters, digits, and hyphens without path separators.")
		finalize(&result)
		return result, nil
	}
	if opts.SourcePack != VirtualSourcePackID {
		addConflict(&result, "unknown_source_pack", "", "", fmt.Sprintf("unsupported project source pack %q", opts.SourcePack), "Use kas-default-project-suite for KASPROJ-002 dry-run planning.")
		finalize(&result)
		return result, nil
	}

	packs, registryChecksum, err := discoverDefaultProjectSuite(sourceRepo)
	if err != nil {
		addConflict(&result, "source_suite_unreadable", "", "", err.Error(), "Fix skill-pack.yaml/source suite discovery before approved install.")
		finalize(&result)
		return result, nil
	}
	roleRegistry, role, _, selected, excluded, roleConflicts, roleDiagnostics := resolveProjectSuiteRole(sourceRepo, opts.SuiteRole, packs, opts.Project)
	result.RoleRegistry = roleRegistry
	result.RoleLabel = role.DisplayLabel
	if role.SelectionMode == "full_source_suite" {
		result.SuiteMode = SuiteModeFull
	} else if role.ID != "" {
		result.SuiteMode = SuiteModeRoleSubset
	}
	result.SelectedSkills = selected
	result.ExcludedSkills = excluded
	result.Conflicts = append(result.Conflicts, roleConflicts...)
	result.Diagnostics = append(result.Diagnostics, roleDiagnostics...)
	if len(roleConflicts) > 0 || len(roleDiagnostics) > 0 {
		finalize(&result)
		return result, nil
	}
	trust, manifestConflicts, manifestChecksum := trustedProjectSuite(profileRoot, opts.Project, opts.SourcePack)
	for _, c := range manifestConflicts {
		result.Conflicts = append(result.Conflicts, c)
		result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: c.Condition, Message: c.Message})
	}
	composition, componentChanged, componentConflicts, componentBackups := planProjectCompositionFiles(profileRoot, opts.Project, role, selected, trust)
	result.PlannedSkills = []PlannedSkill{}
	result.CompositionFiles = composition
	result.ChangedPaths = append(result.ChangedPaths, componentChanged...)
	result.BackupPlan = append(result.BackupPlan, componentBackups...)
	sort.Slice(result.ChangedPaths, func(i, j int) bool { return result.ChangedPaths[i].Path < result.ChangedPaths[j].Path })
	sort.Slice(result.BackupPlan, func(i, j int) bool { return result.BackupPlan[i].Path < result.BackupPlan[j].Path })
	for _, conflict := range componentConflicts {
		result.Conflicts = append(result.Conflicts, conflict)
		result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: conflict.Condition, Message: conflict.Message})
	}
	packChecksums := map[string]string{}
	for _, pack := range packs {
		packChecksums[pack.PackID] = sha256String(pack.Checksum)
	}
	result.SourcePack.PackChecksums = packChecksums
	result.SourcePack.SuiteChecksum = checksumAny(map[string]any{"registry_sha256": registryChecksum, "pack_checksums": packChecksums})

	scanProfileConflicts(&result, profileRoot, opts.Project)
	if manifestChecksum != nil {
		result.TargetProfile.PreviousManifestSHA256 = manifestChecksum
	}
	result.PlannedManifest = plannedManifest(opts, result.SourcePack.SuiteChecksum, result.PlannedSkills, result)
	finalize(&result)
	return result, nil
}

func RenderHumanDryRun(result Result) string {
	status := "ready"
	if !result.OK {
		status = "blocked"
	}
	lines := []string{
		fmt.Sprintf("Status: project KAS dry-run %s - %s", status, result.Project.ID),
		fmt.Sprintf("Source pack: %s", result.SourcePack.ID),
		fmt.Sprintf("Role: %s (%s); selected %d, excluded %d.", result.RoleLabel, result.SuiteRole, len(result.SelectedSkills), len(result.ExcludedSkills)),
		fmt.Sprintf("Plan: skills %d, conflicts %d, plan_hash %s", result.Summary.TotalSkills, result.Summary.ConflictCount, result.PlanHash),
		"Writes: dry-run only; profile/manifest/KAH/KAB writes 0.",
		"Approval evidence: " + result.ApprovalRequest.EvidenceRef,
	}
	for _, conflict := range result.Conflicts {
		lines = append(lines, "Conflict: "+conflict.Condition+" - "+conflict.Message)
	}
	lines = append(lines, "Next: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func baseResult(sourceRepo string, profileRoot string, opts Options) Result {
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	return Result{
		OK:               true,
		Command:          "install-project-kas",
		Mode:             "project_dry_run",
		CLIVersion:       install.CLIVersion,
		DryRun:           true,
		NoWrite:          NoWriteEvidence{Guaranteed: true},
		SourceRepo:       discovery.SourceRepoInfo(sourceRepo),
		TargetProfile:    discovery.TargetProfile{Name: opts.Profile, Root: profileRoot, ManifestPath: manifestPath, ManifestState: manifestState(manifestPath)},
		Project:          Project{ID: opts.Project, TargetSuitePath: "skills/" + opts.Project},
		SourcePack:       SourcePack{ID: opts.SourcePack, ResolvedFrom: "skill-pack.yaml", SourceRepo: discovery.SourceRepoInfo(sourceRepo), PackChecksums: map[string]string{}, FormalRegistry: "skill-pack.yaml"},
		SuiteRole:        opts.SuiteRole,
		RoleRegistry:     RoleRegistryEvidence{Path: RoleRegistryPath},
		SelectedSkills:   []RoleSkillEvidence{},
		ExcludedSkills:   []RoleSkillEvidence{},
		ProjectTailoring: ProjectTailoring{Mode: "plugin_base_with_project_overlay", Preserves: []string{"project_language", "project_runtime", "project_test_commands", "project_authority_ladder"}, Source: "plugin_namespace_skill_pack_yaml_and_project_overlay", SemanticPortRequiredBeforeApprovedInstall: false, SemanticAdaptationClaimed: false},
		PlannedSkills:    []PlannedSkill{},
		CompositionFiles: []PlannedCompositionFile{},
		ChangedPaths:     []ChangedPath{},
		BackupPlan:       []BackupEntry{},
		Conflicts:        []Conflict{},
		Diagnostics:      []discovery.Diagnostic{},
		NextAction:       "Review changed_paths and approve with install --profile " + opts.Profile + " --project " + opts.Project + " --suite-role " + opts.SuiteRole + " --apply dry-run:<plan_hash>; verify with doctor --project-suite after approved install.",
	}
}

func planProjectCompositionFiles(profileRoot string, project string, role projectSuiteRole, selected []RoleSkillEvidence, trust map[string]manifestSkillRecord) ([]PlannedCompositionFile, []ChangedPath, []Conflict, []BackupEntry) {
	plans := []PlannedCompositionFile{}
	changed := []ChangedPath{}
	conflicts := []Conflict{}
	backups := []BackupEntry{}
	for _, spec := range projectCompositionSpecs(project) {
		content := projectCompositionContent(project, spec.Kind, role, selected)
		newSHA := sha256Bytes([]byte(content))
		trusted := trust[spec.TargetPath]
		action, prev, errCode, errMessage := targetAction(profileRoot, spec.TargetPath, newSHA, trusted)
		if (spec.Kind == "project_overlay" || spec.Kind == "project_overlay_reference") && action == "update" && prev != nil && trusted.TargetPath != "" && trusted.Checksum == *prev {
			action = "skip"
			newSHA = *prev
		}
		if errCode != "" {
			conflicts = append(conflicts, conflict(errCode, project, spec.Name, spec.TargetPath, errMessage, "Resolve the existing profile target before approved install."))
		}
		plans = append(plans, PlannedCompositionFile{Kind: spec.Kind, Name: spec.Name, TargetPath: spec.TargetPath, DriftPolicy: spec.DriftPolicy, Checksum: newSHA, Action: action, Bytes: len(content), TailoringMode: spec.TailoringMode})
		entry := ChangedPath{Path: spec.TargetPath, Action: action, InstalledSkill: spec.Name, SourcePackID: "project_composition", PreviousSHA256: prev, NewSHA256: newSHA, Bytes: len(content), ErrorCode: errCode, ErrorMessage: errMessage}
		if action == "update" && prev != nil {
			backup := BackupEntry{Path: spec.TargetPath, BackupPath: filepath.ToSlash(filepath.Join(".kas", "backups", "dry-run", filepath.FromSlash(spec.TargetPath))), PreviousSHA256: *prev, Bytes: len(content)}
			entry.BackupPath = backup.BackupPath
			backups = append(backups, backup)
		}
		changed = append(changed, entry)
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].TargetPath < plans[j].TargetPath })
	sort.Slice(changed, func(i, j int) bool { return changed[i].Path < changed[j].Path })
	sort.Slice(backups, func(i, j int) bool { return backups[i].Path < backups[j].Path })
	return plans, changed, conflicts, backups
}

type projectCompositionSpec struct {
	Kind          string
	Name          string
	TargetPath    string
	DriftPolicy   string
	TailoringMode string
}

func projectCompositionSpecs(project string) []projectCompositionSpec {
	return []projectCompositionSpec{
		{Kind: "project_wrapper", Name: project + "-wrapper", TargetPath: filepath.ToSlash(filepath.Join("skills", project, project+"-wrapper", "SKILL.md")), DriftPolicy: "manual_review_required", TailoringMode: "generated_project_wrapper"},
		{Kind: "project_overlay", Name: project + "-overlay", TargetPath: filepath.ToSlash(filepath.Join("skills", project, project+"-overlay", "SKILL.md")), DriftPolicy: "manual_review_required", TailoringMode: "profile_local_repo_semantic_tailoring"},
		{Kind: "project_overlay_reference", Name: project + "-overlay-legacy-delta-extract", TargetPath: filepath.ToSlash(filepath.Join("skills", project, project+"-overlay", "references", "legacy-delta-extract.md")), DriftPolicy: "manual_review_required", TailoringMode: "profile_local_repo_semantic_tailoring"},
	}
}

func projectCompositionContent(project string, kind string, role projectSuiteRole, selected []RoleSkillEvidence) string {
	switch kind {
	case "project_wrapper":
		return projectWrapperContent(project, role)
	case "project_overlay":
		return projectOverlayContent(project, role, selected)
	case "project_overlay_reference":
		return projectOverlayReferenceContent(project)
	default:
		return ""
	}
}

func normalizeProjectSuiteRoleID(roleID string) string {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return "blue_commander"
	}
	return roleID
}

func projectSuiteRoleManifest(roleID string) string {
	switch normalizeProjectSuiteRoleID(roleID) {
	case "blue_commander":
		return "kkachi-agent-skills:roles/blue.yaml"
	case "red_reviewer":
		return "kkachi-agent-skills:roles/red.yaml"
	case "orange_pm_reviewer":
		return "kkachi-agent-skills:roles/orange.yaml"
	case "gray_scribe":
		return "kkachi-agent-skills:roles/gray.yaml"
	default:
		color, _, _ := strings.Cut(normalizeProjectSuiteRoleID(roleID), "_")
		if color == "" {
			color = "blue"
		}
		return "kkachi-agent-skills:roles/" + color + ".yaml"
	}
}

func renderAppliesTo(selected []RoleSkillEvidence) string {
	if len(selected) == 0 {
		return "      - kkachi-agent-skills:kkachi-plan"
	}
	lines := make([]string, 0, len(selected))
	for _, skill := range selected {
		if strings.TrimSpace(skill.SourceSkill) == "" {
			continue
		}
		lines = append(lines, "      - kkachi-agent-skills:"+skill.SourceSkill)
	}
	if len(lines) == 0 {
		return "      - kkachi-agent-skills:kkachi-plan"
	}
	return strings.Join(lines, "\n")
}

func projectWrapperContent(project string, role projectSuiteRole) string {
	roleID := normalizeProjectSuiteRoleID(role.ID)
	roleManifest := projectSuiteRoleManifest(roleID)
	return fmt.Sprintf(`---
name: %[1]s-wrapper
description: Thin project wrapper for kkachi-agent-skills plugin skills while working on %[1]s.
version: 0.1.0
metadata:
  kas:
    kind: project_wrapper
    project: %[1]s
    role: %[2]s
    role_manifest: %[3]s
    plugin_namespace: kkachi-agent-skills
    overlay_skill: %[1]s-overlay
    refresh_skill: kkachi-agent-skills:kkachi-agent-skills-overlay-refresh
    base_copy_policy: forbidden_by_default
---

# %[1]s wrapper

Use this wrapper as the profile-local entry point for %[1]s work.

## Composition

1. Load the relevant plugin-qualified base skill from kkachi-agent-skills for the active phase.
2. Load %[1]s-overlay for project-specific constraints and retained local lessons.
3. Treat copied project-suite skills as stale legacy; the profile-local project suite must contain only this wrapper and the project overlay.

## Source-built overlay refresh

When refreshing this project overlay from a source-built kkachi-agent-skills checkout, also load:

`+"```text"+`
skill_view("kkachi-agent-skills:kkachi-agent-skills-overlay-refresh")
`+"```"+`

Treat refresh as LLM-assisted semantic porting, not CLI migration.

## Boundary

- Scope is %[1]s project guidance for this Hermes profile.
- This wrapper does not authorize credential/service configuration, backend activation, broad profile cleanup, commit, push, or release.
`, project, roleID, roleManifest)
}

func projectOverlayContent(project string, role projectSuiteRole, selected []RoleSkillEvidence) string {
	roleID := normalizeProjectSuiteRoleID(role.ID)
	appliesTo := renderAppliesTo(selected)
	return fmt.Sprintf(`---
name: %[1]s-overlay
description: Project overlay for kkachi-agent-skills plugin skills while working on %[1]s.
version: 0.1.0
metadata:
  kas:
    kind: project_overlay
    project: %[1]s
    role: %[2]s
    applies_to:
%[3]s
    merge_mode: additive_constraints
    base_version: "0.1.0"
    refresh_skill: kkachi-agent-skills:kkachi-agent-skills-overlay-refresh
    references:
      - references/legacy-delta-extract.md
---

# %[1]s overlay

Use this overlay when applying KAS phase skills to the %[1]s project from this Hermes profile.

## Source-built overlay refresh

When refreshing this project overlay from a source-built kkachi-agent-skills checkout, load the refresh guide before proposing CLI repair/install steps:

`+"```text"+`
skill_view("kkachi-agent-skills:kkachi-agent-skills-overlay-refresh")
`+"```"+`

Treat refresh as LLM-assisted semantic porting, not CLI migration.

## Project boundaries

- Keep project-specific rules, runtime notes, architecture hints, verification posture, and retained local lessons here or under references/.
- Keep shared KAS semantics in the plugin base skills; do not copy full base skill bodies into this overlay.
- Keep credential, service, backend activation, push, release, and public distribution changes out of scope unless 주군 approves them explicitly.

## Retained local lessons

- Legacy delta extract: references/legacy-delta-extract.md
- If a legacy suite was archived before install, semantically port only durable project-specific lessons into this overlay/reference tree.

## Operating rules

1. Load the relevant plugin KAS base skill first.
2. Apply this overlay as additive project constraints.
3. Delete temporary legacy comparison archives after refresh/extraction/verification; keep only non-secret backup evidence when needed.
`, project, roleID, appliesTo)
}

func projectOverlayReferenceContent(project string) string {
	return fmt.Sprintf(`# Legacy delta extract for %[1]s overlay

No legacy archive was present or read by this deterministic install step.

Use this reference to preserve manually reviewed project-specific lessons from an archived profile-local suite. Do not paste full copied base skill bodies here. Record only durable deltas that improve work on %[1]s, such as programming language/runtime notes, repository architecture, verification commands, authority boundaries, recurring pitfalls, and approved workflow exceptions.

## Semantic port checklist

- [ ] Compare archived %[1]s-* copied skills against current plugin base behavior.
- [ ] Extract only project-specific constraints or proven local lessons.
- [ ] Remove stale provider/runtime/auth details unless re-verified.
- [ ] Keep secrets and credentials out of this file.
- [ ] Cite evidence paths for retained lessons when available.
`, project)
}

func planVirtualSuite(sourceRepo string, profileRoot string, project string, packs []discovery.SourcePack, trust map[string]manifestSkillRecord) ([]PlannedSkill, []ChangedPath, []Conflict, []BackupEntry) {
	planned := []PlannedSkill{}
	changed := []ChangedPath{}
	conflicts := []Conflict{}
	backups := []BackupEntry{}
	seenSkills := map[string]bool{}
	seenTargets := map[string]bool{}
	for _, pack := range packs {
		sourceSkill := sourceSkillID(pack.PackID)
		installed := renderInstalledSkill(project, sourceSkill)
		target := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
		if seenSkills[installed] {
			conflicts = append(conflicts, conflict("duplicate_installed_skill", project, installed, target, "source suite maps multiple skills to the same installed skill", "Fix kas-default-project-suite mapping before approved install."))
		}
		if seenTargets[target] {
			conflicts = append(conflicts, conflict("duplicate_target_path", project, installed, target, "source suite maps multiple skills to the same target path", "Fix kas-default-project-suite mapping before approved install."))
		}
		seenSkills[installed] = true
		seenTargets[target] = true
		if isGenericInstalledSkill(installed, project) {
			conflicts = append(conflicts, conflict("generic_installed_skill_name", project, installed, target, "planned installed skill is generic or lacks the project prefix", "Fix the source suite mapping so every installed skill starts with "+project+"-."))
		}
		if unsafeTargetPath(project, installed, target) {
			conflicts = append(conflicts, conflict("unsafe_target_path", project, installed, target, "planned target path escapes the project suite layout", "Use the canonical skills/<project>/<project>-<skill>/SKILL.md layout."))
		}
		content, err := plannedSkillContent(sourceRepo, pack, sourceSkill, installed)
		if err != nil {
			conflicts = append(conflicts, conflict("source_skill_unreadable", project, installed, target, err.Error(), "Fix the source skill before planning project install."))
			content = []byte{}
		}
		newSHA := sha256Bytes(content)
		action, prev, errCode, errMessage := targetAction(profileRoot, target, newSHA, trust[target])
		if errCode != "" {
			conflicts = append(conflicts, conflict(errCode, project, installed, target, errMessage, "Resolve the existing profile target before approved install."))
		}
		tailoringMode := "prefix_render_only"
		planned = append(planned, PlannedSkill{SourcePackID: pack.PackID, SourceSkill: sourceSkill, InstalledSkill: installed, TargetPath: target, DriftPolicy: "manual_review_required", Checksum: newSHA, Action: action, Bytes: len(content), TailoringMode: tailoringMode})
		entry := ChangedPath{Path: target, Action: action, InstalledSkill: installed, SourcePackID: pack.PackID, PreviousSHA256: prev, NewSHA256: newSHA, Bytes: len(content), ErrorCode: errCode, ErrorMessage: errMessage}
		if action == "update" && prev != nil {
			backup := BackupEntry{Path: target, BackupPath: filepath.ToSlash(filepath.Join(".kas", "backups", "dry-run", filepath.FromSlash(target))), PreviousSHA256: *prev, Bytes: len(content)}
			entry.BackupPath = backup.BackupPath
			backups = append(backups, backup)
		}
		changed = append(changed, entry)
	}
	sort.Slice(planned, func(i, j int) bool { return planned[i].TargetPath < planned[j].TargetPath })
	sort.Slice(changed, func(i, j int) bool { return changed[i].Path < changed[j].Path })
	sort.Slice(backups, func(i, j int) bool { return backups[i].Path < backups[j].Path })
	return planned, changed, conflicts, backups
}

func plannedSkillContent(sourceRepo string, pack discovery.SourcePack, sourceSkill string, installed string) ([]byte, error) {
	path := filepath.Join(sourceRepo, filepath.FromSlash(pack.SourcePath), "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	if sourceSkill != installed {
		content = strings.ReplaceAll(content, sourceSkill, installed)
	}
	if pack.Name != "" && pack.Name != sourceSkill && pack.Name != installed {
		content = strings.ReplaceAll(content, pack.Name, installed)
	}
	return []byte(content), nil
}

func targetAction(profileRoot string, target string, newSHA string, trusted manifestSkillRecord) (string, *string, string, string) {
	abs := filepath.Join(profileRoot, filepath.FromSlash(target))
	info, err := os.Lstat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "create", nil, "", ""
		}
		return "error", nil, "target_stat_failed", err.Error()
	}
	if !info.Mode().IsRegular() {
		return "conflict", nil, "existing_target_not_regular", "existing target is not a regular file"
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "error", nil, "target_read_failed", err.Error()
	}
	prev := sha256Bytes(data)
	if trusted.TargetPath == "" {
		if prev == newSHA {
			return "conflict", &prev, "existing_target_not_manifested", "existing target exists but no trusted project suite manifest entry owns it"
		}
		return "conflict", &prev, "existing_target_not_manifested", "existing target exists but is not trusted by a KAS project suite manifest"
	}
	if trusted.Checksum == "" {
		return "error", &prev, "manifest_checksum_missing", "trusted project suite manifest entry is missing checksum"
	}
	if prev != trusted.Checksum {
		return "conflict", &prev, "profile_local_modification", "existing target differs from the trusted project suite manifest checksum"
	}
	if prev == newSHA {
		return "skip", &prev, "", ""
	}
	return "update", &prev, "", ""
}

func scanProfileConflicts(result *Result, profileRoot string, project string) {
	suiteRoot := filepath.Join(profileRoot, "skills", project)
	entries, err := os.ReadDir(suiteRoot)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == project+"-wrapper" || name == project+"-overlay" {
			continue
		}
		if strings.HasPrefix(name, project+"-") {
			target := filepath.ToSlash(filepath.Join("skills", project, name, "SKILL.md"))
			addConflict(result, "stale_copied_project_skill", name, target, "profile project suite contains a copied project-prefixed skill directory; plugin base skills must stay in the plugin namespace", "Move any durable local deltas into the project overlay/reference tree, then remove the copied skill directory.")
			continue
		}
		target := filepath.ToSlash(filepath.Join("skills", project, name, "SKILL.md"))
		addConflict(result, "generic_installed_skill_name", name, target, "profile project suite contains a generic or non-project-prefixed skill directory", "Move or migrate generic skills only through a later approved migration workflow.")
	}
}

func plannedManifest(opts Options, sourceChecksum string, skills []PlannedSkill, result Result) map[string]any {
	installed := make([]map[string]any, 0, len(skills))
	for _, skill := range skills {
		installed = append(installed, map[string]any{"installed_skill": skill.InstalledSkill, "source_pack_id": skill.SourcePackID, "target_path": skill.TargetPath, "checksum": skill.Checksum, "drift_policy": skill.DriftPolicy, "tailoring_mode": skill.TailoringMode})
	}
	composition := make([]map[string]any, 0, len(result.CompositionFiles))
	for _, component := range result.CompositionFiles {
		composition = append(composition, map[string]any{"kind": component.Kind, "name": component.Name, "target_path": component.TargetPath, "checksum": component.Checksum, "drift_policy": component.DriftPolicy, "tailoring_mode": component.TailoringMode})
	}
	suite := map[string]any{
		"kind":                        ManifestKind,
		"project":                     opts.Project,
		"source_pack":                 map[string]any{"id": opts.SourcePack, "repo": "kkachi-agent-skills", "checksum": sourceChecksum, "language_profile": "project-specific-prefix-render-only", "formal_registry": "skill-pack.yaml"},
		"drift_policy":                suiteDriftPolicy(result.SuiteMode),
		"semantic_adaptation_claimed": false,
		"installed_skills":            installed,
		"composition_files":           composition,
	}
	if result.SuiteRole != "" {
		suite["suite_mode"] = result.SuiteMode
		suite["suite_role"] = result.SuiteRole
		suite["role_registry"] = result.RoleRegistry
		suite["selected_skills"] = result.SelectedSkills
		suite["excluded_skills"] = result.ExcludedSkills
	}
	return map[string]any{
		"version":        ManifestVersion,
		"kind":           ProfileManifestKind,
		"profile":        map[string]any{"name": opts.Profile},
		"project_suites": []map[string]any{suite},
		"installs":       []any{},
	}
}

func suiteDriftPolicy(suiteMode string) string {
	if suiteMode == SuiteModeRoleSubset {
		return "role_subset_expected"
	}
	return "manual_review_required"
}

func finalize(result *Result) {
	counts := map[string]int{"create": 0, "update": 0, "skip": 0, "conflict": 0, "error": 0, "backup": 0, "manifest_update": 0}
	for _, changed := range result.ChangedPaths {
		counts[changed.Action]++
	}
	result.Summary = Summary{TotalSkills: len(result.PlannedSkills), TotalFiles: len(result.ChangedPaths), SelectedSkills: len(result.SelectedSkills), ExcludedSkills: len(result.ExcludedSkills), CountsByAction: counts, ConflictCount: len(result.Conflicts), DiagnosticCount: len(result.Diagnostics)}
	result.Checksums = Checksums{SourcePack: result.SourcePack.SuiteChecksum, PlannedManifest: checksumAny(result.PlannedManifest), PlannedSkills: checksumAny(result.PlannedSkills), ChangedPaths: checksumAny(result.ChangedPaths)}
	canonical := map[string]any{"command": result.Command, "mode": result.Mode, "cli_version": result.CLIVersion, "dry_run": result.DryRun, "no_write": result.NoWrite, "source_repo": result.SourceRepo, "target_profile": result.TargetProfile, "manifest_path": result.TargetProfile.ManifestPath, "manifest_state": result.TargetProfile.ManifestState, "previous_manifest_sha256": result.TargetProfile.PreviousManifestSHA256, "project": result.Project, "source_pack": result.SourcePack, "suite_role": result.SuiteRole, "suite_mode": result.SuiteMode, "role_label": result.RoleLabel, "role_registry": result.RoleRegistry, "selected_skills": result.SelectedSkills, "excluded_skills": result.ExcludedSkills, "project_tailoring": result.ProjectTailoring, "summary": result.Summary, "planned_manifest": result.PlannedManifest, "planned_skills": result.PlannedSkills, "changed_paths": result.ChangedPaths, "backup_plan": result.BackupPlan, "checksums": result.Checksums, "conflicts": result.Conflicts, "diagnostics": result.Diagnostics}
	result.PlanHash = checksumAny(canonical)
	result.ApprovalRequest = ApprovalRequest{Required: result.OK, EvidenceRef: "dry-run:" + result.PlanHash, DryRunPlanHash: result.PlanHash, HashIncludesProfile: true, HashIncludesManifestState: true, HashIncludesSourceSuite: true, HashIncludesRoleFields: true, HashIncludesNoWriteEvidence: true, HashIncludesBackupPlan: true, HashIncludesConflictsAndDiags: true}
	result.OK = len(result.Conflicts) == 0 && noErrorDiagnostics(result.Diagnostics)
	if !result.OK {
		result.ApprovalRequest.Required = false
		result.NextAction = "Resolve project-suite conflicts and rerun dry-run; approved install fails closed until the current plan is ok:true."
	}
}

func addDiagnostic(result *Result, level string, code string, message string) {
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: level, Code: code, Message: message})
}

func addConflict(result *Result, condition string, installedSkill string, targetPath string, message string, nextAction string) {
	c := conflict(condition, result.Project.ID, installedSkill, targetPath, message, nextAction)
	result.Conflicts = append(result.Conflicts, c)
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: c.Condition, Message: c.Message})
}

func conflict(condition string, project string, installedSkill string, targetPath string, message string, nextAction string) Conflict {
	return Conflict{Severity: "error", Condition: condition, Project: project, InstalledSkill: installedSkill, TargetPath: targetPath, Message: message, NextAction: nextAction}
}

func noErrorDiagnostics(diagnostics []discovery.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == "error" {
			return false
		}
	}
	return true
}

func resolveProfileRoot(profile string, override string) string {
	root := override
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		root = filepath.Join(home, ".hermes", "profiles", profile)
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}

func manifestState(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "manifest_missing"
		}
		return "manifest_unreadable"
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return "manifest_unreadable"
	}
	return "manifest_present"
}

func sourceSkillID(packID string) string {
	parts := strings.Split(packID, "/")
	return parts[len(parts)-1]
}

func renderInstalledSkill(project string, sourceSkill string) string {
	tail := strings.TrimPrefix(sourceSkill, "kkachi-")
	return project + "-" + tail
}

func isGenericInstalledSkill(installed string, project string) bool {
	if !strings.HasPrefix(installed, project+"-") {
		return true
	}
	generic := map[string]bool{"kkachi-plan": true, "kas-plan": true, "plan": true, "implement": true, "final-verify": true, "kkachi-implement": true, "kkachi-final-verify": true}
	return generic[installed]
}

func unsafeTargetPath(project string, installed string, target string) bool {
	if discovery.IsInvalidRelativePath(target) {
		return true
	}
	want := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
	if target != want {
		return true
	}
	return !strings.HasPrefix(target, "skills/"+project+"/")
}

func checksumAny(value any) string {
	data, _ := json.Marshal(value)
	return sha256Bytes(data)
}

func sha256Bytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func sha256String(hexValue string) string {
	if strings.HasPrefix(hexValue, "sha256:") {
		return hexValue
	}
	return "sha256:" + hexValue
}
