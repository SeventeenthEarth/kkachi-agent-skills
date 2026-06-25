package projectinstall

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
)

const (
	modeProjectSuiteDoctor = "project_suite_doctor"
	modeRepairDryRun       = "project_repair_dry_run"
	modeRepairApproved     = "project_repair_approved"
	modeMigrateDryRun      = "project_migration_dry_run"
	modeMigrateApproved    = "project_migration_approved"
)

type ProjectSuiteOptions struct {
	Profile            string
	Project            string
	SuiteRole          string
	PruneExtra         bool
	BackupVaultRoot    string
	SourcePack         string
	SourcePackExplicit bool
	ProfileRoot        string
	FromGeneric        bool
}

type ProjectSuiteDiagnostic struct {
	Project        string `json:"project"`
	InstalledSkill string `json:"installed_skill,omitempty"`
	TargetPath     string `json:"target_path,omitempty"`
	Severity       string `json:"severity"`
	Condition      string `json:"condition"`
	Message        string `json:"message"`
	NextAction     string `json:"next_action"`
}

type ProjectSuiteState struct {
	ManifestState       string `json:"manifest_state"`
	PhysicalState       string `json:"physical_state"`
	InstalledSkillCount int    `json:"installed_skill_count"`
	FilesChecked        int    `json:"files_checked"`
}

type ProjectSuiteDoctorResult struct {
	OK                      bool                     `json:"ok"`
	Command                 string                   `json:"command"`
	Mode                    string                   `json:"mode"`
	TargetProfile           discovery.TargetProfile  `json:"target_profile"`
	Project                 Project                  `json:"project"`
	SourcePack              SourcePack               `json:"source_pack"`
	ManifestPath            string                   `json:"manifest_path"`
	ProjectSuite            ProjectSuiteState        `json:"project_suite"`
	Diagnostics             []discovery.Diagnostic   `json:"diagnostics"`
	ProjectSuiteDiagnostics []ProjectSuiteDiagnostic `json:"project_suite_diagnostics"`
	NextAction              string                   `json:"next_action"`
}

type PlannedAction struct {
	Action         string  `json:"action"`
	Project        string  `json:"project"`
	InstalledSkill string  `json:"installed_skill"`
	SourcePackID   string  `json:"source_pack_id"`
	SourceSkill    string  `json:"source_skill"`
	SourcePath     string  `json:"source_path,omitempty"`
	TargetPath     string  `json:"target_path"`
	Reason         string  `json:"reason"`
	PreviousSHA256 *string `json:"previous_sha256,omitempty"`
	NewSHA256      string  `json:"new_sha256,omitempty"`
	Bytes          int     `json:"bytes,omitempty"`
	BackupPath     string  `json:"backup_path,omitempty"`
}

type ManualSemanticPortTask struct {
	Project    string `json:"project"`
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Reason     string `json:"reason"`
	NextAction string `json:"next_action"`
}

type ProjectActionResult struct {
	OK                      bool                     `json:"ok"`
	Command                 string                   `json:"command"`
	Mode                    string                   `json:"mode"`
	CLIVersion              string                   `json:"cli_version"`
	DryRun                  bool                     `json:"dry_run"`
	NoWrite                 NoWriteEvidence          `json:"no_write"`
	SourceRepo              discovery.SourceRepo     `json:"source_repo"`
	TargetProfile           discovery.TargetProfile  `json:"target_profile"`
	Project                 Project                  `json:"project"`
	SourcePack              SourcePack               `json:"source_pack"`
	SuiteRole               string                   `json:"suite_role,omitempty"`
	SuiteMode               string                   `json:"suite_mode,omitempty"`
	RoleLabel               string                   `json:"role_label,omitempty"`
	RoleRegistry            RoleRegistryEvidence     `json:"role_registry,omitempty"`
	SelectedSkills          []RoleSkillEvidence      `json:"selected_skills,omitempty"`
	ExcludedSkills          []RoleSkillEvidence      `json:"excluded_skills,omitempty"`
	PruneExtra              bool                     `json:"prune_extra"`
	Summary                 Summary                  `json:"summary"`
	ManifestPath            string                   `json:"manifest_path"`
	ProjectSuiteDiagnostics []ProjectSuiteDiagnostic `json:"project_suite_diagnostics"`
	PlannedActions          []PlannedAction          `json:"planned_actions"`
	PlannedSkills           []PlannedSkill           `json:"planned_skills,omitempty"`
	CompositionFiles        []PlannedCompositionFile `json:"composition_files,omitempty"`
	ChangedPaths            []ChangedPath            `json:"changed_paths"`
	BackupPlan              []BackupEntry            `json:"backup_plan"`
	NoSpillover             NoSpilloverEvidence      `json:"no_spillover"`
	Checksums               Checksums                `json:"checksums"`
	PlanHash                string                   `json:"plan_hash"`
	ApprovalRequest         ApprovalRequest          `json:"approval_request"`
	Approval                ApprovalEvidence         `json:"approval,omitempty"`
	RepairID                string                   `json:"repair_id,omitempty"`
	MigrationID             string                   `json:"migration_id,omitempty"`
	BackupPath              string                   `json:"backup_path,omitempty"`
	Recovery                *Recovery                `json:"recovery,omitempty"`
	ManualSemanticPortTasks []ManualSemanticPortTask `json:"manual_semantic_port_tasks"`
	Diagnostics             []discovery.Diagnostic   `json:"diagnostics"`
	NextAction              string                   `json:"next_action"`
}

type NoSpilloverEvidence struct {
	ScopeProfile                    string   `json:"scope_profile"`
	ScopeProject                    string   `json:"scope_project"`
	UnknownPersonalSkillsPreserved  []string `json:"unknown_personal_skills_preserved"`
	UnrelatedProjectSuitesPreserved []string `json:"unrelated_project_suites_preserved"`
	UnrelatedProfilesMutated        int      `json:"unrelated_profiles_mutated"`
	KAHStateWriteCount              int      `json:"kah_state_write_count"`
	KABRuntimeMutationCount         int      `json:"kab_runtime_mutation_count"`
	ManifestWriteLast               bool     `json:"manifest_write_last"`
}

type projectSuiteManifestInfo struct {
	ManifestPath string
	ManifestSHA  *string
	State        string
	Payload      map[string]any
	Suite        map[string]any
	Records      map[string]manifestSkillRecord
	Diagnostics  []ProjectSuiteDiagnostic
}

func BuildProjectSuiteDoctor(repo string, opts ProjectSuiteOptions) (ProjectSuiteDoctorResult, error) {
	sourceRepo, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return ProjectSuiteDoctorResult{}, err
	}
	profileRoot := resolveProfileRoot(opts.Profile, opts.ProfileRoot)
	sourcePackID := opts.SourcePack
	if sourcePackID == "" {
		sourcePackID = VirtualSourcePackID
	}
	result := ProjectSuiteDoctorResult{
		OK:            true,
		Command:       "doctor",
		Mode:          modeProjectSuiteDoctor,
		TargetProfile: discovery.TargetProfile{Name: opts.Profile, Root: profileRoot, ManifestPath: filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"), ManifestState: manifestState(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))},
		Project:       Project{ID: opts.Project, TargetSuitePath: "skills/" + opts.Project},
		SourcePack:    SourcePack{ID: sourcePackID, Source: "default_or_manifest", ResolvedFrom: "default_or_manifest", SourceRepo: discovery.SourceRepoInfo(sourceRepo), PackChecksums: map[string]string{}, FormalRegistry: "skill-pack.yaml"},
		ManifestPath:  filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"),
		NextAction:    "Project suite is healthy; rerun doctor after repair, migration, or source-pack updates.",
	}
	if opts.Profile == "" {
		addSuiteDoctorDiag(&result, suiteDiag(opts.Project, "", "", "error", "profile_required", "doctor --project-suite requires --profile <profile>.", "Rerun with --profile <profile>."))
	}
	if opts.Project == "" || !validProjectID.MatchString(opts.Project) {
		addSuiteDoctorDiag(&result, suiteDiag(opts.Project, "", "", "error", "invalid_project_id", fmt.Sprintf("project id %q is not a safe project suite id", opts.Project), "Use lowercase letters, digits, and hyphens without path separators."))
	}
	if sourcePackID != VirtualSourcePackID {
		addSuiteDoctorDiag(&result, suiteDiag(opts.Project, "", "", "error", "unknown_source_pack", fmt.Sprintf("unsupported project source pack %q", sourcePackID), "Use kas-default-project-suite."))
		finalizeProjectSuiteDoctor(&result)
		return result, nil
	}
	packs, registrySHA, err := discoverDefaultProjectSuite(sourceRepo)
	if err != nil {
		addSuiteDoctorDiag(&result, suiteDiag(opts.Project, "", "", "error", "source_suite_unreadable", err.Error(), "Fix skill-pack.yaml/source suite discovery."))
		finalizeProjectSuiteDoctor(&result)
		return result, nil
	}
	populateSourcePackEvidence(&result.SourcePack, packs, registrySHA)
	info := inspectProjectSuite(profileRoot, opts.Project, sourcePackID, packs, true)
	info.Diagnostics = boundProjectSuiteDoctorInspectDiagnostics(info.Diagnostics)
	info.Diagnostics = append(info.Diagnostics, roleAwareProjectSuiteDoctorDiagnostics(sourceRepo, profileRoot, opts.Project, packs, info)...)
	result.TargetProfile.PreviousManifestSHA256 = info.ManifestSHA
	result.TargetProfile.ManifestState = info.State
	result.ManifestPath = info.ManifestPath
	result.ProjectSuiteDiagnostics = append(result.ProjectSuiteDiagnostics, info.Diagnostics...)
	result.ProjectSuite = suiteState(profileRoot, opts.Project, info)
	finalizeProjectSuiteDoctor(&result)
	return result, nil
}

func boundProjectSuiteDoctorInspectDiagnostics(diags []ProjectSuiteDiagnostic) []ProjectSuiteDiagnostic {
	bounded := make([]ProjectSuiteDiagnostic, len(diags))
	for i, diag := range diags {
		bounded[i] = diag
		bounded[i].NextAction = boundProjectSuiteDoctorInspectNextAction(diag)
	}
	return bounded
}

func boundProjectSuiteDoctorInspectNextAction(diag ProjectSuiteDiagnostic) string {
	switch diag.Condition {
	case "missing_file", "checksum_mismatch", "missing_project_suite", "umbrella_only":
		return "Review diagnostics and use only an approved KASROLE-004 repair/prune plan when applicable."
	}
	if strings.Contains(diag.NextAction, "repair-project-kas") || strings.Contains(diag.NextAction, "migrate-project-kas") {
		return "Review diagnostics and use only an approved KASROLE-004 repair/prune plan when applicable."
	}
	return diag.NextAction
}

func RenderHumanProjectSuiteDoctor(result ProjectSuiteDoctorResult) string {
	state := "healthy"
	if !result.OK {
		state = "error"
	} else if hasSuiteSeverity(result.ProjectSuiteDiagnostics, "warning") {
		state = "warning"
	}
	lines := []string{
		fmt.Sprintf("Status: %s - profile %s project-suite doctor %s.", state, result.TargetProfile.Name, result.Project.ID),
		fmt.Sprintf("manifest: %s (%s)", result.ProjectSuite.ManifestState, result.ManifestPath),
		fmt.Sprintf("suite: %s (%s), files checked %d.", result.ProjectSuite.PhysicalState, result.Project.TargetSuitePath, result.ProjectSuite.FilesChecked),
		"source_pack: " + result.SourcePack.ID,
	}
	for _, diagnostic := range result.ProjectSuiteDiagnostics {
		lines = append(lines, humanProjectSuiteDiagnosticLine(diagnostic))
	}
	lines = append(lines, "Next: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func BuildProjectRepairDryRun(repo string, opts ProjectSuiteOptions) (ProjectActionResult, error) {
	return buildProjectActionDryRun(repo, opts, "repair-project-kas", modeRepairDryRun, false)
}

func BuildProjectMigrationDryRun(repo string, opts ProjectSuiteOptions) (ProjectActionResult, error) {
	return buildProjectActionDryRun(repo, opts, "migrate-project-kas", modeMigrateDryRun, true)
}

func ApplyApprovedRepair(repo string, opts ProjectSuiteOptions, evidenceRef string) (ProjectActionResult, error) {
	return applyApprovedProjectAction(repo, opts, evidenceRef, "repair-project-kas", modeRepairApproved, BuildProjectRepairDryRun)
}

func ApplyApprovedMigration(repo string, opts ProjectSuiteOptions, evidenceRef string) (ProjectActionResult, error) {
	return applyApprovedProjectAction(repo, opts, evidenceRef, "migrate-project-kas", modeMigrateApproved, BuildProjectMigrationDryRun)
}

func RenderHumanProjectAction(result ProjectActionResult) string {
	status := "ready"
	if !result.OK {
		status = "blocked"
	}
	if !result.DryRun && result.OK {
		status = "complete"
	}
	lines := []string{
		fmt.Sprintf("Status: %s %s - profile %s / project %s.", result.Command, status, result.TargetProfile.Name, result.Project.ID),
		fmt.Sprintf("Source pack: %s", result.SourcePack.ID),
		fmt.Sprintf("Role: %s (%s); prune_extra:%t.", result.RoleLabel, result.SuiteRole, result.PruneExtra),
		fmt.Sprintf("Plan: actions %d, manual tasks %d, plan_hash %s", len(result.PlannedActions), len(result.ManualSemanticPortTasks), result.PlanHash),
		fmt.Sprintf("Counts: keep %d, create %d, update %d, remove %d, manifest_update %d.", result.Summary.CountsByAction["skip"], result.Summary.CountsByAction["create"], result.Summary.CountsByAction["update"], result.Summary.CountsByAction["remove"], result.Summary.CountsByAction["manifest_update"]),
	}
	if result.DryRun {
		lines = append(lines, "Writes: dry-run only; profile/manifest/KAH/KAB/auth/provider writes 0.")
		lines = append(lines, fmt.Sprintf("Approval required: %t", result.ApprovalRequest.Required))
		lines = append(lines, "Approval evidence: "+result.ApprovalRequest.EvidenceRef)
		lines = append(lines, "Backup: apply requires explicit absolute --backup-vault-root; dry-run writes no backup files.")
	} else {
		lines = append(lines, "Approval evidence: "+result.Approval.EvidenceRef, "Recovery: "+result.BackupPath)
	}
	lines = append(lines, fmt.Sprintf("Recovery: manifest write last:%t; rollback uses backup path for changed/removed files.", result.NoSpillover.ManifestWriteLast))
	lines = append(lines, fmt.Sprintf("No-spillover: preserved unknown personal skills %d; unrelated profiles mutated %d; KAH writes %d; KAB runtime mutations %d.", len(result.NoSpillover.UnknownPersonalSkillsPreserved), result.NoSpillover.UnrelatedProfilesMutated, result.NoSpillover.KAHStateWriteCount, result.NoSpillover.KABRuntimeMutationCount))
	for _, diagnostic := range result.ProjectSuiteDiagnostics {
		lines = append(lines, humanProjectSuiteDiagnosticLine(diagnostic))
	}
	for _, action := range firstPlannedActions(result.PlannedActions, 5) {
		lines = append(lines, humanPlannedActionLine(action))
	}
	if len(result.PlannedActions) > 5 {
		lines = append(lines, fmt.Sprintf("Action: ... %d more planned actions", len(result.PlannedActions)-5))
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, "Diagnostic: "+diagnostic.Message)
	}
	for _, task := range result.ManualSemanticPortTasks {
		lines = append(lines, "Manual semantic-port: "+task.Reason+" - "+task.SourcePath)
	}
	lines = append(lines, "Next: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func humanProjectSuiteDiagnosticLine(diagnostic ProjectSuiteDiagnostic) string {
	parts := []string{fmt.Sprintf("project-suite diagnostic: %s/%s - %s", diagnostic.Severity, diagnostic.Condition, diagnostic.Message)}
	if diagnostic.InstalledSkill != "" {
		parts = append(parts, "skill "+diagnostic.InstalledSkill)
	}
	if diagnostic.TargetPath != "" {
		parts = append(parts, "path "+diagnostic.TargetPath)
	}
	if diagnostic.NextAction != "" {
		parts = append(parts, "next_action: "+diagnostic.NextAction)
	}
	return strings.Join(parts, " | ")
}

func firstPlannedActions(actions []PlannedAction, limit int) []PlannedAction {
	if len(actions) <= limit {
		return actions
	}
	return actions[:limit]
}

func humanPlannedActionLine(action PlannedAction) string {
	line := fmt.Sprintf("Action: %s %s", action.Action, action.TargetPath)
	if action.InstalledSkill != "" {
		line += " (" + action.InstalledSkill + ")"
	}
	if action.Reason != "" {
		line += " - " + action.Reason
	}
	if action.SourcePath != "" {
		line += " from " + action.SourcePath
	}
	return line
}

func buildProjectActionDryRun(repo string, opts ProjectSuiteOptions, command string, mode string, migrate bool) (ProjectActionResult, error) {
	sourceRepo, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return ProjectActionResult{}, err
	}
	profileRoot := resolveProfileRoot(opts.Profile, opts.ProfileRoot)
	sourcePackID, source := resolvedProjectActionSourcePack(opts)
	result := baseProjectActionResult(command, mode, sourceRepo, profileRoot, opts, sourcePackID, source)
	if opts.Profile == "" {
		addProjectActionError(&result, "profile_required", "command requires --profile <profile>.")
	}
	if opts.Project == "" || !validProjectID.MatchString(opts.Project) {
		addProjectActionError(&result, "invalid_project_id", fmt.Sprintf("project id %q is not a safe project suite id", opts.Project))
	}
	if sourcePackID != VirtualSourcePackID {
		addProjectActionError(&result, "unknown_source_pack", fmt.Sprintf("unsupported project source pack %q", sourcePackID))
		finalizeProjectAction(&result)
		return result, nil
	}
	if migrate && !opts.FromGeneric {
		addProjectActionError(&result, "from_generic_required", "migrate-project-kas requires --from-generic so generic-to-project migration is never implicit.")
		finalizeProjectAction(&result)
		return result, nil
	}
	packs, registrySHA, err := discoverDefaultProjectSuite(sourceRepo)
	if err != nil {
		addProjectActionError(&result, "source_suite_unreadable", err.Error())
		finalizeProjectAction(&result)
		return result, nil
	}
	populateSourcePackEvidence(&result.SourcePack, packs, registrySHA)
	selectedPacks := packs
	roleAwareRepair := !migrate && (strings.TrimSpace(opts.SuiteRole) != "" || opts.PruneExtra)
	if roleAwareRepair {
		roleRegistry, role, rolePacks, selected, excluded, roleConflicts, roleDiagnostics := resolveProjectSuiteRole(sourceRepo, opts.SuiteRole, packs, opts.Project)
		result.RoleRegistry = roleRegistry
		result.RoleLabel = role.DisplayLabel
		if role.SelectionMode == "full_source_suite" {
			result.SuiteMode = SuiteModeFull
		} else if role.ID != "" {
			result.SuiteMode = SuiteModeRoleSubset
		}
		result.SelectedSkills = selected
		result.ExcludedSkills = excluded
		for _, c := range roleConflicts {
			result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: c.Severity, Code: c.Condition, Message: c.Message})
		}
		result.Diagnostics = append(result.Diagnostics, roleDiagnostics...)
		if len(roleConflicts) > 0 || len(roleDiagnostics) > 0 {
			finalizeProjectAction(&result)
			return result, nil
		}
		selectedPacks = rolePacks
	}
	info := inspectProjectSuite(profileRoot, opts.Project, sourcePackID, packs, !migrate)
	result.TargetProfile.PreviousManifestSHA256 = info.ManifestSHA
	result.TargetProfile.ManifestState = info.State
	result.ProjectSuiteDiagnostics = append(result.ProjectSuiteDiagnostics, info.Diagnostics...)
	if roleAwareRepair {
		result.ProjectSuiteDiagnostics = append(result.ProjectSuiteDiagnostics, roleAwareProjectSuiteDoctorDiagnostics(sourceRepo, profileRoot, opts.Project, packs, info)...)
		addRequestedRoleManifestDiagnostics(&result, info)
		result.NoSpillover = scanProjectActionNoSpillover(opts.Profile, profileRoot, opts.Project, info)
	}

	if migrate {
		planMigrationActions(&result, sourceRepo, profileRoot, opts.Project, packs)
	} else {
		planRepairActions(&result, sourceRepo, profileRoot, opts.Project, selectedPacks, packs, info)
	}
	finalizeProjectAction(&result)
	return result, nil
}

func baseProjectActionResult(command string, mode string, sourceRepo string, profileRoot string, opts ProjectSuiteOptions, sourcePackID string, source string) ProjectActionResult {
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	return ProjectActionResult{
		OK:                      true,
		Command:                 command,
		Mode:                    mode,
		CLIVersion:              CLIVersionForKASPROJ004(),
		DryRun:                  true,
		NoWrite:                 NoWriteEvidence{Guaranteed: true},
		SourceRepo:              discovery.SourceRepoInfo(sourceRepo),
		TargetProfile:           discovery.TargetProfile{Name: opts.Profile, Root: profileRoot, ManifestPath: manifestPath, ManifestState: manifestState(manifestPath)},
		Project:                 Project{ID: opts.Project, TargetSuitePath: "skills/" + opts.Project},
		SourcePack:              SourcePack{ID: sourcePackID, Source: source, ResolvedFrom: source, SourceRepo: discovery.SourceRepoInfo(sourceRepo), PackChecksums: map[string]string{}, FormalRegistry: "skill-pack.yaml"},
		SuiteRole:               opts.SuiteRole,
		RoleRegistry:            RoleRegistryEvidence{Path: RoleRegistryPath},
		SelectedSkills:          []RoleSkillEvidence{},
		ExcludedSkills:          []RoleSkillEvidence{},
		PruneExtra:              opts.PruneExtra,
		ManifestPath:            manifestPath,
		ProjectSuiteDiagnostics: []ProjectSuiteDiagnostic{},
		PlannedActions:          []PlannedAction{},
		PlannedSkills:           []PlannedSkill{},
		CompositionFiles:        []PlannedCompositionFile{},
		ChangedPaths:            []ChangedPath{},
		BackupPlan:              []BackupEntry{},
		NoSpillover:             NoSpilloverEvidence{ScopeProfile: opts.Profile, ScopeProject: opts.Project, UnrelatedProfilesMutated: 0, KAHStateWriteCount: 0, KABRuntimeMutationCount: 0, ManifestWriteLast: true},
		ManualSemanticPortTasks: []ManualSemanticPortTask{},
		Diagnostics:             []discovery.Diagnostic{},
		NextAction:              "Review dry-run evidence and approve with " + command + " --approve dry-run:<hash>; rerun doctor --project-suite after approved changes.",
	}
}

func CLIVersionForKASPROJ004() string { return version.CLIVersion }

func resolvedProjectActionSourcePack(opts ProjectSuiteOptions) (string, string) {
	if opts.SourcePack == "" {
		return VirtualSourcePackID, "default"
	}
	if opts.SourcePackExplicit {
		return opts.SourcePack, "explicit"
	}
	return opts.SourcePack, "default"
}

func populateSourcePackEvidence(target *SourcePack, packs []discovery.SourcePack, registrySHA string) {
	packChecksums := map[string]string{}
	for _, pack := range packs {
		packChecksums[pack.PackID] = sha256String(pack.Checksum)
	}
	target.PackChecksums = packChecksums
	target.SuiteChecksum = checksumAny(map[string]any{"registry_sha256": registrySHA, "pack_checksums": packChecksums})
}

func inspectProjectSuite(profileRoot string, project string, sourcePackID string, packs []discovery.SourcePack, requireTrustedManifest bool) projectSuiteManifestInfo {
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	info := projectSuiteManifestInfo{ManifestPath: manifestPath, State: manifestState(manifestPath), Records: map[string]manifestSkillRecord{}, Diagnostics: []ProjectSuiteDiagnostic{}}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if requireTrustedManifest {
				info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", "skills/"+project, "error", "missing_project_suite", "no trusted project_suites[] entry exists for project "+project, "Install or repair the project-specific suite; do not use generic KAS skills as fallback."))
			}
			addPhysicalSuiteDiagnostics(&info, profileRoot, project)
			return info
		}
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", ".kas/skill-pack-manifest.json", "error", "manifest_unreadable", "profile manifest is unreadable: "+err.Error(), "Fix manifest permissions before doctor/repair/migrate."))
		return info
	}
	sha := sha256Bytes(data)
	info.ManifestSHA = &sha
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", ".kas/skill-pack-manifest.json", "error", "manifest_parse_error", "cannot parse KAS manifest: "+err.Error(), "Fix the profile manifest before doctor/repair/migrate."))
		return info
	}
	info.Payload = manifest
	if version, _ := manifest["version"].(string); version != ManifestVersion {
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", ".kas/skill-pack-manifest.json", "error", "unsupported_manifest_version", fmt.Sprintf("unsupported KAS manifest version: %q", manifest["version"]), "Use a compatible KAS profile manifest."))
		return info
	}
	if kind, _ := manifest["kind"].(string); kind != ProfileManifestKind {
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", ".kas/skill-pack-manifest.json", "error", "unsupported_manifest_kind", fmt.Sprintf("unsupported KAS manifest kind: %q", manifest["kind"]), "Use kas_profile_skill_manifest."))
		return info
	}
	rawSuites, ok := manifest["project_suites"].([]any)
	if !ok {
		if requireTrustedManifest {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", "skills/"+project, "error", "missing_project_suite", "manifest has no trusted project_suites[] entry for project "+project, "Install or repair the project-specific suite; do not use generic KAS skills as fallback."))
		}
		addPhysicalSuiteDiagnostics(&info, profileRoot, project)
		return info
	}
	matches := []map[string]any{}
	for _, raw := range rawSuites {
		suite, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if suite["project"] == project && sourcePackIDFromProjectSuite(suite) == sourcePackID {
			matches = append(matches, suite)
		}
	}
	if len(matches) == 0 {
		if requireTrustedManifest {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", "skills/"+project, "error", "missing_project_suite", "no trusted matching project_suites[] entry exists for project/source_pack", "Install or repair the project-specific suite; do not use generic KAS skills as fallback."))
		}
		addPhysicalSuiteDiagnostics(&info, profileRoot, project)
		return info
	}
	if len(matches) > 1 {
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", ".kas/skill-pack-manifest.json", "error", "ambiguous_project_suite_manifest", "multiple matching project_suites[] entries exist", "Manually repair duplicate project suite manifest entries before proceeding."))
		return info
	}
	info.Suite = matches[0]
	if kind, _ := info.Suite["kind"].(string); kind != ManifestKind {
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", ".kas/skill-pack-manifest.json", "error", "unsupported_project_suite_manifest_kind", fmt.Sprintf("unsupported project suite manifest kind: %q", info.Suite["kind"]), "Repair the project suite manifest."))
		return info
	}
	rawSkills, ok := info.Suite["installed_skills"].([]any)
	if !ok {
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", ".kas/skill-pack-manifest.json", "error", "manifest_project_skills_invalid", "project suite installed_skills must be an array", "Repair the project suite manifest."))
		return info
	}
	manifestOnlyUmbrella := len(rawSkills) > 0
	for _, raw := range rawSkills {
		skill, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		installed, _ := skill["installed_skill"].(string)
		target, _ := skill["target_path"].(string)
		checksum, _ := skill["checksum"].(string)
		driftPolicy, _ := skill["drift_policy"].(string)
		tailoringMode, _ := skill["tailoring_mode"].(string)
		if installed != project+"-kas" {
			manifestOnlyUmbrella = false
		}
		if unsafeManifestProjectTarget(project, installed, target) {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, installed, target, "error", "profile_source_language_drift", "manifest project skill identity or target path is not project-prefixed/canonical", "Manually repair required phase identity before automated repair/migration."))
			continue
		}
		if info.Records[target].TargetPath != "" {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, installed, target, "error", "duplicate_manifest_target_path", "project suite manifest contains duplicate target_path", "Manually repair duplicate target paths."))
			continue
		}
		info.Records[target] = manifestSkillRecord{InstalledSkill: installed, TargetPath: target, Checksum: checksum, DriftPolicy: driftPolicy, TailoringMode: tailoringMode}
		if lp := sourceLanguageProfile(info.Suite); lp != "" && lp != "project-specific-prefix-render-only" {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, installed, target, "warning", "profile_source_language_drift", "source_pack language_profile differs from prefix-render-only KASPROJ posture", "Review project language/runtime/test-command assumptions before relying on this suite."))
		}
		if drift, _ := skill["drift_policy"].(string); drift != "" && drift != "manual_review_required" && drift != "repair_after_approval" && drift != "fail_closed" {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, installed, target, "warning", "profile_source_language_drift", "installed skill drift_policy is outside known KASPROJ policies", "Review source/profile drift before relying on this suite."))
		}
	}
	if rawComponents, ok := info.Suite["composition_files"].([]any); ok {
		for _, raw := range rawComponents {
			component, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			name, _ := component["name"].(string)
			target, _ := component["target_path"].(string)
			checksum, _ := component["checksum"].(string)
			driftPolicy, _ := component["drift_policy"].(string)
			tailoringMode, _ := component["tailoring_mode"].(string)
			kind, _ := component["kind"].(string)
			if discovery.IsInvalidRelativePath(target) || !strings.HasPrefix(target, "skills/"+project+"/") {
				info.Diagnostics = append(info.Diagnostics, suiteDiag(project, name, target, "error", "profile_source_language_drift", "manifest composition file target path is not project-scoped/canonical", "Manually repair composition file identity before automated repair/migration."))
				continue
			}
			if info.Records[target].TargetPath != "" {
				info.Diagnostics = append(info.Diagnostics, suiteDiag(project, name, target, "error", "duplicate_manifest_target_path", "project suite manifest contains duplicate target_path", "Manually repair duplicate target paths."))
				continue
			}
			info.Records[target] = manifestSkillRecord{InstalledSkill: name, TargetPath: target, Checksum: checksum, DriftPolicy: driftPolicy, TailoringMode: tailoringMode, Kind: kind}
		}
	}
	if manifestOnlyUmbrella {
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, project+"-kas", filepath.ToSlash(filepath.Join("skills", project, project+"-kas", "SKILL.md")), "error", "umbrella_only", "manifest lists only the project umbrella skill", "Repair or install the full project-prefixed suite."))
	}
	addPhysicalSuiteDiagnostics(&info, profileRoot, project)
	for target, record := range info.Records {
		abs := filepath.Join(profileRoot, filepath.FromSlash(target))
		data, err := os.ReadFile(abs)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				info.Diagnostics = append(info.Diagnostics, suiteDiag(project, record.InstalledSkill, target, "error", "missing_file", "manifest target_path is absent", "Run repair-project-kas --dry-run and approve a matching repair plan."))
			} else {
				info.Diagnostics = append(info.Diagnostics, suiteDiag(project, record.InstalledSkill, target, "error", "target_read_failed", "manifest target_path cannot be read: "+err.Error(), "Fix profile file permissions before repair."))
			}
			continue
		}
		if record.Checksum == "" {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, record.InstalledSkill, target, "error", "checksum_mismatch", "installed file checksum differs from manifest/source evidence", "Run repair-project-kas --dry-run and approve a matching repair plan, or review local edits manually."))
			continue
		}
		if sha256Bytes(data) != record.Checksum {
			if projectTailoringChecksumDriftAllowed(record) {
				info.Diagnostics = append(info.Diagnostics, suiteDiag(project, record.InstalledSkill, target, "warning", "project_tailoring_checksum_drift", "installed file differs from the template/install checksum, but manifest marks this skill as project-local semantic tailoring", "Review the project-local skill content and preserve it; do not run source repair unless the tailoring is intentionally being reset."))
				continue
			}
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, record.InstalledSkill, target, "error", "checksum_mismatch", "installed file checksum differs from manifest/source evidence", "Run repair-project-kas --dry-run and approve a matching repair plan, or review local edits manually."))
		}
	}
	return info
}

func roleAwareProjectSuiteDoctorDiagnostics(sourceRepo string, profileRoot string, project string, packs []discovery.SourcePack, info projectSuiteManifestInfo) []ProjectSuiteDiagnostic {
	if info.Suite == nil {
		return nil
	}
	suiteRole, _ := info.Suite["suite_role"].(string)
	if strings.TrimSpace(suiteRole) == "" {
		return []ProjectSuiteDiagnostic{suiteDiag(project, "", ".kas/skill-pack-manifest.json", "error", "missing_suite_role", "project suite manifest is missing suite_role; doctor cannot infer role from profile name or installed skills", "Reinstall with an explicit registered suite_role, or use only an approved KASROLE-004 workflow when applicable.")}
	}
	_, role, _, selected, _, conflicts, diagnostics := resolveProjectSuiteRole(sourceRepo, suiteRole, packs, project)
	roleDiags := []ProjectSuiteDiagnostic{}
	if len(conflicts) > 0 || len(diagnostics) > 0 {
		for _, diag := range diagnostics {
			roleDiags = append(roleDiags, suiteDiag(project, "", RoleRegistryPath, diag.Level, diag.Code, diag.Message, "Use a registered suite_role from registries/project-suite-roles.yaml; do not infer a fallback role."))
		}
		return roleDiags
	}

	selectedTargets := map[string]RoleSkillEvidence{}
	selectedInstalled := map[string]bool{}
	for _, skill := range selected {
		selectedTargets[skill.TargetPath] = skill
		selectedInstalled[skill.InstalledSkill] = true
	}

	managedInstalled := map[string]RoleSkillEvidence{}
	for _, pack := range packs {
		sourceSkill := sourceSkillID(pack.PackID)
		installed := renderInstalledSkill(project, sourceSkill)
		target := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
		managedInstalled[installed] = RoleSkillEvidence{SourceSkill: sourceSkill, InstalledSkill: installed, SourcePackID: pack.PackID, TargetPath: target}
	}
	seenOutOfRole := map[string]bool{}
	outOfRoleSeverity := "error"
	if role.ID == "blue_commander" {
		outOfRoleSeverity = "warning"
	}
	for target, record := range info.Records {
		if record.Kind != "" {
			continue
		}
		if selectedTargets[target].TargetPath != "" {
			continue
		}
		if managed, ok := managedInstalled[record.InstalledSkill]; ok {
			roleDiags = append(roleDiags, suiteDiag(project, record.InstalledSkill, target, outOfRoleSeverity, "out_of_role_kas_managed_skill", fmt.Sprintf("suite_role %s does not select KAS-managed skill %s", suiteRole, record.InstalledSkill), "Prune only through a later approved repair/prune workflow; doctor does not mutate profiles."))
			seenOutOfRole[managed.InstalledSkill] = true
		}
	}

	suiteRoot := filepath.Join(profileRoot, "skills", project)
	entries, err := os.ReadDir(suiteRoot)
	if err != nil {
		return roleDiags
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		target := filepath.ToSlash(filepath.Join("skills", project, name, "SKILL.md"))
		if _, err := os.Stat(filepath.Join(profileRoot, filepath.FromSlash(target))); err != nil {
			continue
		}
		if info.Records[target].TargetPath != "" {
			continue
		}
		if managed, ok := managedInstalled[name]; ok && !selectedInstalled[name] && !seenOutOfRole[name] {
			roleDiags = append(roleDiags, suiteDiag(project, managed.InstalledSkill, target, outOfRoleSeverity, "out_of_role_kas_managed_skill", fmt.Sprintf("suite_role %s does not select KAS-managed physical skill %s", suiteRole, managed.InstalledSkill), "Prune only through a later approved repair/prune workflow; doctor does not mutate profiles."))
			seenOutOfRole[name] = true
			continue
		}
		if _, managed := managedInstalled[name]; managed {
			continue
		}
		if shadowsManagedSkill(name, managedInstalled) {
			roleDiags = append(roleDiags, suiteDiag(project, name, target, "error", "ambiguous_profile_skill_dir", "unknown project skill shadows or ambiguously overlaps a KAS-managed skill identity", "Rename or manually review the personal skill before relying on project-suite doctor health."))
		}
	}
	return roleDiags
}

func shadowsManagedSkill(name string, managed map[string]RoleSkillEvidence) bool {
	for managedName := range managed {
		if name == managedName {
			return true
		}
		if strings.HasPrefix(name, managedName+"-") || strings.HasPrefix(managedName, name+"-") {
			return true
		}
	}
	return false
}

func sourceLanguageProfile(suite map[string]any) string {
	sourcePack, _ := suite["source_pack"].(map[string]any)
	value, _ := sourcePack["language_profile"].(string)
	return value
}

func projectTailoringChecksumDriftAllowed(record manifestSkillRecord) bool {
	return record.TailoringMode == "profile_local_repo_semantic_tailoring" && record.DriftPolicy == "manual_review_required"
}

func addPhysicalSuiteDiagnostics(info *projectSuiteManifestInfo, profileRoot string, project string) {
	suiteRoot := filepath.Join(profileRoot, "skills", project)
	entries, err := os.ReadDir(suiteRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", "skills/"+project, "error", "missing_project_suite", "no skills/<project>/ suite exists", "Install or repair the project-specific suite; never substitute global generic skills."))
		} else {
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, "", "skills/"+project, "error", "project_suite_unreadable", "skills/<project>/ suite is unreadable: "+err.Error(), "Fix profile suite permissions."))
		}
		return
	}
	umbrellaOnly := false
	nonUmbrella := false
	required := map[string]bool{}
	for _, record := range info.Records {
		required[record.InstalledSkill] = true
	}
	manifestedTargets := map[string]bool{}
	for target := range info.Records {
		manifestedTargets[target] = true
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		target := filepath.ToSlash(filepath.Join("skills", project, name, "SKILL.md"))
		if name == project+"-kas" {
			umbrellaOnly = true
		} else if strings.HasPrefix(name, project+"-") {
			nonUmbrella = true
		}
		if _, err := os.Stat(filepath.Join(profileRoot, filepath.FromSlash(target))); err != nil {
			continue
		}
		if !manifestedTargets[target] {
			severity := "warning"
			if !strings.HasPrefix(name, project+"-") || required[name] {
				severity = "error"
			}
			info.Diagnostics = append(info.Diagnostics, suiteDiag(project, name, target, severity, "unknown_profile_skill_dir", "profile skill directory is not tracked by the KAS project manifest", "Do not adopt, overwrite, or delete it automatically; migrate only through explicit approved migration."))
		}
	}
	if umbrellaOnly && !nonUmbrella {
		info.Diagnostics = append(info.Diagnostics, suiteDiag(project, project+"-kas", filepath.ToSlash(filepath.Join("skills", project, project+"-kas", "SKILL.md")), "error", "umbrella_only", "physical suite contains only the project umbrella skill", "Repair or install the full project-prefixed suite."))
	}
}

func suiteState(profileRoot string, project string, info projectSuiteManifestInfo) ProjectSuiteState {
	physical := "present"
	if st, err := os.Stat(filepath.Join(profileRoot, "skills", project)); err != nil || !st.IsDir() {
		physical = "missing"
	}
	filesChecked := 0
	installedCount := 0
	for target, record := range info.Records {
		if record.Kind == "" {
			installedCount++
		}
		if _, err := os.Stat(filepath.Join(profileRoot, filepath.FromSlash(target))); err == nil {
			filesChecked++
		}
	}
	return ProjectSuiteState{ManifestState: info.State, PhysicalState: physical, InstalledSkillCount: installedCount, FilesChecked: filesChecked}
}

func addRequestedRoleManifestDiagnostics(result *ProjectActionResult, info projectSuiteManifestInfo) {
	if info.Suite == nil {
		return
	}
	manifestRole, _ := info.Suite["suite_role"].(string)
	if strings.TrimSpace(manifestRole) == "" {
		result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: "missing_suite_role", Message: "project suite manifest is missing suite_role; repair/prune requires explicit manifest role evidence."})
		return
	}
	if manifestRole != result.SuiteRole {
		result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: "suite_role_mismatch", Message: fmt.Sprintf("requested suite_role %q does not match manifest suite_role %q.", result.SuiteRole, manifestRole)})
	}
}

func scanProjectActionNoSpillover(profile string, profileRoot string, project string, info projectSuiteManifestInfo) NoSpilloverEvidence {
	evidence := NoSpilloverEvidence{ScopeProfile: profile, ScopeProject: project, UnknownPersonalSkillsPreserved: []string{}, UnrelatedProjectSuitesPreserved: []string{}, UnrelatedProfilesMutated: 0, KAHStateWriteCount: 0, KABRuntimeMutationCount: 0, ManifestWriteLast: true}
	manifested := map[string]bool{}
	for target := range info.Records {
		manifested[target] = true
	}
	root := filepath.Join(profileRoot, "skills", project)
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "SKILL.md" {
			return nil
		}
		rel, err := filepath.Rel(profileRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if !manifested[rel] {
			evidence.UnknownPersonalSkillsPreserved = append(evidence.UnknownPersonalSkillsPreserved, rel)
		}
		return nil
	})
	if info.Payload != nil {
		if rawSuites, ok := info.Payload["project_suites"].([]any); ok {
			for _, raw := range rawSuites {
				suite, ok := raw.(map[string]any)
				if !ok || suite["project"] == project {
					continue
				}
				evidence.UnrelatedProjectSuitesPreserved = append(evidence.UnrelatedProjectSuitesPreserved, fmt.Sprint(suite["project"]))
			}
		}
	}
	sort.Strings(evidence.UnknownPersonalSkillsPreserved)
	sort.Strings(evidence.UnrelatedProjectSuitesPreserved)
	return evidence
}

func planRepairActions(result *ProjectActionResult, sourceRepo string, profileRoot string, project string, packs []discovery.SourcePack, allPacks []discovery.SourcePack, info projectSuiteManifestInfo) {
	pruneTargets := map[string]bool{}
	if result.PruneExtra {
		for _, removal := range plannedRolePruneActions(result, profileRoot, project, allPacks, info) {
			pruneTargets[removal.TargetPath] = true
			result.PlannedActions = append(result.PlannedActions, removal)
			prev := removal.PreviousSHA256
			result.ChangedPaths = append(result.ChangedPaths, ChangedPath{Path: removal.TargetPath, Action: "remove", InstalledSkill: removal.InstalledSkill, SourcePackID: removal.SourcePackID, PreviousSHA256: prev, Bytes: removal.Bytes})
			if prev != nil {
				result.BackupPlan = append(result.BackupPlan, BackupEntry{Path: removal.TargetPath, BackupPath: filepath.ToSlash(filepath.Join("<backup-vault-root>", "dry-run", "files", filepath.FromSlash(removal.TargetPath))), PreviousSHA256: *prev, Bytes: removal.Bytes})
			}
		}
	}
	for _, diag := range info.Diagnostics {
		if result.PruneExtra && diag.Condition == "out_of_role_kas_managed_skill" && pruneTargets[diag.TargetPath] {
			continue
		}
		if diag.Severity == "error" && !repairableProjectCondition(diag.Condition) {
			result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: diag.Condition, Message: diag.Message})
		}
	}
	for _, diag := range result.ProjectSuiteDiagnostics {
		if result.PruneExtra && diag.Condition == "out_of_role_kas_managed_skill" && pruneTargets[diag.TargetPath] {
			continue
		}
		if diag.Severity == "error" && !repairableProjectCondition(diag.Condition) {
			if !hasDiagnosticCode(result.Diagnostics, diag.Condition) {
				result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: diag.Condition, Message: diag.Message})
			}
		}
	}
	composition, compositionChanged, compositionActions, compositionBackups, compositionManifestNeeds := planRepairCompositionFiles(profileRoot, project, projectSuiteRole{ID: result.SuiteRole, DisplayLabel: result.RoleLabel}, result.SelectedSkills, info)
	result.CompositionFiles = composition
	result.ChangedPaths = append(result.ChangedPaths, compositionChanged...)
	for _, changed := range compositionChanged {
		if changed.Action == "conflict" || changed.Action == "error" {
			code := changed.ErrorCode
			if code == "" {
				code = "project_composition_repair_conflict"
			}
			message := changed.ErrorMessage
			if message == "" {
				message = "project composition repair plan contains an unsafe target action"
			}
			result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: code, Message: message})
		}
	}
	result.PlannedActions = append(result.PlannedActions, compositionActions...)
	result.BackupPlan = append(result.BackupPlan, compositionBackups...)
	if manifestNeedsProjectRepair(result, info) || compositionManifestNeeds || len(pruneTargets) > 0 || (result.SuiteRole != "" && manifestRoleNeedsRepair(result, info)) {
		result.PlannedActions = append(result.PlannedActions, PlannedAction{Action: "manifest_update", Project: project, TargetPath: ".kas/skill-pack-manifest.json", Reason: "refresh_project_suite_manifest"})
	}
	sortProjectActionResult(result)
}

func plannedRolePruneActions(result *ProjectActionResult, profileRoot string, project string, packs []discovery.SourcePack, info projectSuiteManifestInfo) []PlannedAction {
	selectedTargets := map[string]bool{}
	for _, skill := range result.SelectedSkills {
		selectedTargets[skill.TargetPath] = true
	}
	managedByTarget := map[string]RoleSkillEvidence{}
	for _, pack := range packs {
		sourceSkill := sourceSkillID(pack.PackID)
		installed := renderInstalledSkill(project, sourceSkill)
		target := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
		managedByTarget[target] = RoleSkillEvidence{SourceSkill: sourceSkill, InstalledSkill: installed, SourcePackID: pack.PackID, TargetPath: target}
	}
	actions := []PlannedAction{}
	for target, record := range info.Records {
		if selectedTargets[target] {
			continue
		}
		managed := managedByTarget[target]
		if managed.TargetPath == "" || record.Checksum == "" {
			continue
		}
		current, size, err := existingFileChecksumAndSize(filepath.Join(profileRoot, filepath.FromSlash(target)))
		if err != nil || current != record.Checksum {
			continue
		}
		prev := current
		actions = append(actions, PlannedAction{Action: "remove", Project: project, InstalledSkill: record.InstalledSkill, SourcePackID: managed.SourcePackID, SourceSkill: managed.SourceSkill, TargetPath: target, Reason: "prune_manifest_tracked_out_of_role_kas_managed_skill", PreviousSHA256: &prev, Bytes: size, BackupPath: filepath.ToSlash(filepath.Join("<backup-vault-root>", "dry-run", "files", filepath.FromSlash(target)))})
	}
	sort.Slice(actions, func(i, j int) bool { return actions[i].TargetPath < actions[j].TargetPath })
	return actions
}

func hasDiagnosticCode(diagnostics []discovery.Diagnostic, code string) bool {
	for _, diag := range diagnostics {
		if diag.Code == code {
			return true
		}
	}
	return false
}

func manifestRoleNeedsRepair(result *ProjectActionResult, info projectSuiteManifestInfo) bool {
	if info.Suite == nil {
		return false
	}
	return info.Suite["suite_role"] != result.SuiteRole || info.Suite["suite_mode"] != result.SuiteMode
}

func planRepairCompositionFiles(profileRoot string, project string, role projectSuiteRole, selected []RoleSkillEvidence, info projectSuiteManifestInfo) ([]PlannedCompositionFile, []ChangedPath, []PlannedAction, []BackupEntry, bool) {
	plans := []PlannedCompositionFile{}
	changed := []ChangedPath{}
	actions := []PlannedAction{}
	backups := []BackupEntry{}
	manifestNeeds := false
	for _, spec := range projectCompositionSpecs(project) {
		content := []byte(projectCompositionContent(project, spec.Kind, role, selected))
		newSHA := sha256Bytes(content)
		record := info.Records[spec.TargetPath]
		checksum := newSHA
		action := "skip"
		bytes := len(content)
		abs := filepath.Join(profileRoot, filepath.FromSlash(spec.TargetPath))
		current, size, readErr := existingFileChecksumAndSize(abs)
		if readErr != nil {
			if errors.Is(readErr, os.ErrNotExist) {
				action = "create"
				entry := ChangedPath{Path: spec.TargetPath, Action: "create", InstalledSkill: spec.Name, SourcePackID: "project_composition", NewSHA256: newSHA, Bytes: len(content)}
				changed = append(changed, entry)
				actions = append(actions, PlannedAction{Action: "create", Project: project, InstalledSkill: spec.Name, SourcePackID: "project_composition", SourceSkill: spec.Kind, TargetPath: spec.TargetPath, Reason: "restore_missing_project_composition_file", NewSHA256: newSHA, Bytes: len(content)})
			} else {
				action = "error"
				changed = append(changed, ChangedPath{Path: spec.TargetPath, Action: "error", InstalledSkill: spec.Name, SourcePackID: "project_composition", NewSHA256: newSHA, Bytes: len(content), ErrorCode: "target_read_failed", ErrorMessage: readErr.Error()})
			}
		} else if spec.Kind == "project_overlay" || spec.Kind == "project_overlay_reference" {
			checksum = current
			bytes = size
			if record.TargetPath == "" || record.Checksum != current || record.Kind != spec.Kind || record.DriftPolicy != spec.DriftPolicy || record.TailoringMode != spec.TailoringMode {
				action = "adopt_existing"
				manifestNeeds = true
			}
		} else if current != newSHA {
			action = "update"
			prev := current
			entry := ChangedPath{Path: spec.TargetPath, Action: "update", InstalledSkill: spec.Name, SourcePackID: "project_composition", PreviousSHA256: &prev, NewSHA256: newSHA, Bytes: len(content)}
			backup := BackupEntry{Path: spec.TargetPath, BackupPath: filepath.ToSlash(filepath.Join(".kas", "backups", "dry-run", filepath.FromSlash(spec.TargetPath))), PreviousSHA256: prev, Bytes: size}
			entry.BackupPath = backup.BackupPath
			changed = append(changed, entry)
			backups = append(backups, backup)
			actions = append(actions, PlannedAction{Action: "update", Project: project, InstalledSkill: spec.Name, SourcePackID: "project_composition", SourceSkill: spec.Kind, TargetPath: spec.TargetPath, Reason: "repair_project_wrapper_from_canonical_template", PreviousSHA256: &prev, NewSHA256: newSHA, Bytes: len(content), BackupPath: backup.BackupPath})
		} else if record.TargetPath == "" || record.Checksum != newSHA || record.Kind != spec.Kind || record.DriftPolicy != spec.DriftPolicy || record.TailoringMode != spec.TailoringMode {
			action = "adopt_existing"
			manifestNeeds = true
		}
		if record.TargetPath == "" || record.Checksum != checksum || record.Kind != spec.Kind || record.DriftPolicy != spec.DriftPolicy || record.TailoringMode != spec.TailoringMode {
			manifestNeeds = true
		}
		plans = append(plans, PlannedCompositionFile{Kind: spec.Kind, Name: spec.Name, TargetPath: spec.TargetPath, DriftPolicy: spec.DriftPolicy, Checksum: checksum, Action: action, Bytes: bytes, TailoringMode: spec.TailoringMode})
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].TargetPath < plans[j].TargetPath })
	sort.Slice(changed, func(i, j int) bool { return changed[i].Path < changed[j].Path })
	sort.Slice(actions, func(i, j int) bool { return actions[i].TargetPath < actions[j].TargetPath })
	sort.Slice(backups, func(i, j int) bool { return backups[i].Path < backups[j].Path })
	return plans, changed, actions, backups, manifestNeeds
}

func plannedRepairSkill(sourceRepo string, profileRoot string, project string, pack discovery.SourcePack, info projectSuiteManifestInfo) (PlannedSkill, ChangedPath) {
	sourceSkill := sourceSkillID(pack.PackID)
	installed := renderInstalledSkill(project, sourceSkill)
	target := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
	content, err := plannedSkillContent(sourceRepo, pack, sourceSkill, installed)
	if err != nil {
		content = []byte{}
	}
	newSHA := sha256Bytes(content)
	abs := filepath.Join(profileRoot, filepath.FromSlash(target))
	prev, readErr := existingFileSHA(abs)
	action := "skip"
	errCode := ""
	errMessage := ""
	if readErr != nil {
		action = "error"
		errCode = readErr.Code
		errMessage = readErr.Message
	} else if prev == nil {
		action = "create"
	} else if *prev != newSHA {
		if info.Records[target].TargetPath == "" && !hasRepairableSuiteAbsence(info) {
			action = "conflict"
			errCode = "unmanifested_target_not_repairable"
			errMessage = "existing unmanifested target differs from trusted source"
		} else {
			action = "update"
		}
	}
	return PlannedSkill{SourcePackID: pack.PackID, SourceSkill: sourceSkill, InstalledSkill: installed, TargetPath: target, DriftPolicy: "manual_review_required", Checksum: newSHA, Action: action, Bytes: len(content), TailoringMode: "prefix_render_only"}, ChangedPath{Path: target, Action: action, InstalledSkill: installed, SourcePackID: pack.PackID, PreviousSHA256: prev, NewSHA256: newSHA, Bytes: len(content), ErrorCode: errCode, ErrorMessage: errMessage}
}

func planMigrationActions(result *ProjectActionResult, sourceRepo string, profileRoot string, project string, packs []discovery.SourcePack) {
	genericCandidates := explicitGenericCandidates(profileRoot, packs, result)
	if len(genericCandidates) == 0 && len(result.ManualSemanticPortTasks) == 0 {
		result.ManualSemanticPortTasks = append(result.ManualSemanticPortTasks, ManualSemanticPortTask{Project: project, SourcePath: "skills/kkachi-*", TargetPath: "skills/" + project, Reason: "no clean KAS-managed generic candidate found", NextAction: "Install project suite directly or manually port project-specific guidance."})
	}
	candidateByPack := map[string]string{}
	for _, candidate := range genericCandidates {
		candidateByPack[candidate.packID] = candidate.path
	}
	for _, pack := range packs {
		sourceSkill := sourceSkillID(pack.PackID)
		genericRel, ok := candidateByPack[pack.PackID]
		if !ok {
			continue
		}
		installed := renderInstalledSkill(project, sourceSkill)
		target := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
		content, err := plannedSkillContent(sourceRepo, pack, sourceSkill, installed)
		if err != nil {
			result.ManualSemanticPortTasks = append(result.ManualSemanticPortTasks, ManualSemanticPortTask{Project: project, SourcePath: genericRel, TargetPath: target, Reason: "trusted source skill cannot be rendered deterministically", NextAction: "Repair source pack before migration."})
			continue
		}
		newSHA := sha256Bytes(content)
		abs := filepath.Join(profileRoot, filepath.FromSlash(target))
		prev, readErr := existingFileSHA(abs)
		if readErr != nil {
			result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: readErr.Code, Message: readErr.Message})
			continue
		}
		if prev != nil {
			result.ManualSemanticPortTasks = append(result.ManualSemanticPortTasks, ManualSemanticPortTask{Project: project, SourcePath: genericRel, TargetPath: target, Reason: "project target already exists", NextAction: "Run project-suite doctor/repair or review existing target manually."})
			continue
		}
		skill := PlannedSkill{SourcePackID: pack.PackID, SourceSkill: sourceSkill, InstalledSkill: installed, TargetPath: target, DriftPolicy: "manual_review_required", Checksum: newSHA, Action: "create", Bytes: len(content), TailoringMode: "prefix_render_only"}
		changed := ChangedPath{Path: target, Action: "create", InstalledSkill: installed, SourcePackID: pack.PackID, NewSHA256: newSHA, Bytes: len(content)}
		result.PlannedSkills = append(result.PlannedSkills, skill)
		result.ChangedPaths = append(result.ChangedPaths, changed)
		result.PlannedActions = append(result.PlannedActions, PlannedAction{Action: "create", Project: project, InstalledSkill: installed, SourcePackID: pack.PackID, SourceSkill: sourceSkill, TargetPath: target, SourcePath: genericRel, Reason: "clean_generic_candidate", NewSHA256: newSHA, Bytes: len(content)})
	}
	if len(result.PlannedActions) > 0 {
		result.PlannedActions = append(result.PlannedActions, PlannedAction{Action: "manifest_update", Project: project, TargetPath: ".kas/skill-pack-manifest.json", Reason: "record_project_suite_without_deleting_generic_skills"})
	}
	sortProjectActionResult(result)
}

type genericCandidate struct{ packID, path string }

func explicitGenericCandidates(profileRoot string, packs []discovery.SourcePack, result *ProjectActionResult) []genericCandidate {
	packIDs := map[string]bool{}
	for _, pack := range packs {
		packIDs[pack.PackID] = true
	}
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		addUnmanifestedGenericTasks(profileRoot, result)
		return nil
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		result.ManualSemanticPortTasks = append(result.ManualSemanticPortTasks, ManualSemanticPortTask{Project: result.Project.ID, SourcePath: ".kas/skill-pack-manifest.json", TargetPath: "skills/" + result.Project.ID, Reason: "generic migration manifest is unreadable", NextAction: "Repair the profile manifest before migration."})
		return nil
	}
	manifested := map[string]bool{}
	candidates := []genericCandidate{}
	installs, _ := manifest["installs"].([]any)
	for _, raw := range installs {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		packID, _ := entry["pack_id"].(string)
		if !packIDs[packID] || !strings.HasPrefix(packID, "kkachi-") {
			continue
		}
		targetPath, _ := entry["target_path"].(string)
		if targetPath == "" {
			targetPath = filepath.ToSlash(filepath.Join("skills", packID))
		}
		manifested[targetPath] = true
		if discovery.IsInvalidRelativePath(targetPath) || targetPath != filepath.ToSlash(filepath.Join("skills", packID)) {
			result.ManualSemanticPortTasks = append(result.ManualSemanticPortTasks, ManualSemanticPortTask{Project: result.Project.ID, SourcePath: targetPath, TargetPath: "skills/" + result.Project.ID, Reason: "generic manifest target is ambiguous or unsafe", NextAction: "Repair manifest identity before migration."})
			continue
		}
		fileRel := filepath.ToSlash(filepath.Join(targetPath, "SKILL.md"))
		actual, readErr := existingFileSHA(filepath.Join(profileRoot, filepath.FromSlash(fileRel)))
		if readErr != nil || actual == nil {
			reason := "generic manifest target is missing"
			if readErr != nil {
				reason = readErr.Message
			}
			result.ManualSemanticPortTasks = append(result.ManualSemanticPortTasks, ManualSemanticPortTask{Project: result.Project.ID, SourcePath: fileRel, TargetPath: "skills/" + result.Project.ID, Reason: reason, NextAction: "Repair generic source or install project suite directly."})
			continue
		}
		expected := genericManifestSkillSHA(entry)
		if expected == "" || *actual != expected {
			result.ManualSemanticPortTasks = append(result.ManualSemanticPortTasks, ManualSemanticPortTask{Project: result.Project.ID, SourcePath: fileRel, TargetPath: "skills/" + result.Project.ID, Reason: "generic candidate is locally modified or lacks checksum evidence", NextAction: "Manual semantic port required; KASPROJ-004 will not guess or overwrite."})
			continue
		}
		candidates = append(candidates, genericCandidate{packID: packID, path: fileRel})
	}
	addUnmanifestedGenericTasksExcept(profileRoot, manifested, result)
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].path < candidates[j].path })
	return candidates
}

func genericManifestSkillSHA(entry map[string]any) string {
	files, _ := entry["files"].([]any)
	for _, raw := range files {
		file, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		rel, _ := file["relative_path"].(string)
		if rel != "SKILL.md" {
			continue
		}
		for _, key := range []string{"sha256", "new_sha256"} {
			value, _ := file[key].(string)
			if value != "" {
				return value
			}
		}
	}
	return ""
}

func addUnmanifestedGenericTasks(profileRoot string, result *ProjectActionResult) {
	addUnmanifestedGenericTasksExcept(profileRoot, map[string]bool{}, result)
}

func addUnmanifestedGenericTasksExcept(profileRoot string, manifested map[string]bool, result *ProjectActionResult) {
	root := filepath.Join(profileRoot, "skills")
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "kkachi-") {
			continue
		}
		targetPath := filepath.ToSlash(filepath.Join("skills", entry.Name()))
		if manifested[targetPath] {
			continue
		}
		fileRel := filepath.ToSlash(filepath.Join(targetPath, "SKILL.md"))
		if _, err := os.Stat(filepath.Join(profileRoot, filepath.FromSlash(fileRel))); err == nil {
			result.ManualSemanticPortTasks = append(result.ManualSemanticPortTasks, ManualSemanticPortTask{Project: result.Project.ID, SourcePath: fileRel, TargetPath: "skills/" + result.Project.ID, Reason: "unmanifested generic profile skill", NextAction: "Manual semantic port required; KASPROJ-004 does not adopt unmanifested generic skills."})
		}
	}
}

func actionFromChanged(project string, skill PlannedSkill, changed ChangedPath, reason string) PlannedAction {
	return PlannedAction{Action: changed.Action, Project: project, InstalledSkill: skill.InstalledSkill, SourcePackID: skill.SourcePackID, SourceSkill: skill.SourceSkill, TargetPath: skill.TargetPath, Reason: reason, PreviousSHA256: changed.PreviousSHA256, NewSHA256: changed.NewSHA256, Bytes: changed.Bytes, BackupPath: changed.BackupPath}
}

func repairReason(action string) string {
	switch action {
	case "create":
		return "restore_missing_project_suite_file"
	case "update":
		return "repair_checksum_mismatch_from_trusted_source"
	default:
		return "refresh_project_suite_manifest"
	}
}

func manifestNeedsProjectRepair(result *ProjectActionResult, info projectSuiteManifestInfo) bool {
	if info.Suite == nil || info.ManifestSHA == nil {
		return true
	}
	for _, skill := range result.PlannedSkills {
		record := info.Records[skill.TargetPath]
		if record.TargetPath == "" || record.Checksum != skill.Checksum || record.InstalledSkill != skill.InstalledSkill {
			return true
		}
	}
	return false
}

func repairableProjectCondition(condition string) bool {
	switch condition {
	case "missing_project_suite", "umbrella_only", "missing_file", "checksum_mismatch":
		return true
	default:
		return false
	}
}

func hasRepairableSuiteAbsence(info projectSuiteManifestInfo) bool {
	for _, diag := range info.Diagnostics {
		if diag.Condition == "missing_project_suite" || diag.Condition == "umbrella_only" {
			return true
		}
	}
	return false
}

func applyApprovedProjectAction(repo string, opts ProjectSuiteOptions, evidenceRef string, command string, approvedMode string, dryRunFn func(string, ProjectSuiteOptions) (ProjectActionResult, error)) (ProjectActionResult, error) {
	dryRun, err := dryRunFn(repo, opts)
	if err != nil {
		return ProjectActionResult{}, err
	}
	approvedHash, evidenceOK := approvedHashFromEvidence(evidenceRef)
	if !evidenceOK {
		return projectActionApprovalFailure(dryRun, evidenceRef, approvedHash, approvedMode, "approval_evidence_malformed", "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>."), nil
	}
	if approvedHash != dryRun.PlanHash {
		return projectActionApprovalFailure(dryRun, evidenceRef, approvedHash, approvedMode, "approval_plan_hash_mismatch", "approval evidence does not match the current dry-run plan hash; no files were written."), nil
	}
	if !dryRun.OK || !dryRun.ApprovalRequest.Required {
		return projectActionApprovalFailure(dryRun, evidenceRef, approvedHash, approvedMode, "project_action_plan_not_approvable", "current dry-run plan is not approvable; no files were written."), nil
	}
	operationID := makeProjectActionID(command, dryRun.PlanHash, time.Now().UTC())
	backupRoot := filepath.Join(dryRun.TargetProfile.Root, ".kas", "backups", operationID)
	if command == "repair-project-kas" {
		validated, err := validateBackupVaultRoot(opts.BackupVaultRoot, dryRun.TargetProfile.Root, dryRun.TargetProfile.Name, dryRun.Project.ID, operationID)
		if err != nil {
			return projectActionApprovalFailure(dryRun, evidenceRef, approvedHash, approvedMode, "backup_vault_root_rejected", err.Error()), nil
		}
		backupRoot = validated
	}
	if err := preflightProjectAction(dryRun, ""); err != nil {
		return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, nil, false, []discovery.Diagnostic{{Level: "error", Code: "project_action_preflight_failed", Message: err.Error()}}), nil
	}
	actualChanged := []ChangedPath{}
	previousManifestPath := ""
	if dryRun.TargetProfile.PreviousManifestSHA256 != nil {
		previousManifestPath = filepath.Join(backupRoot, "skill-pack-manifest.json.before")
		if err := copyAbsoluteFile(dryRun.TargetProfile.ManifestPath, previousManifestPath); err != nil {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "previous_manifest_backup_failed", Message: err.Error()}}), nil
		}
	}
	backupByPath := map[string]string{}
	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "update" && entry.Action != "remove" {
			continue
		}
		source, err := safeProfilePath(dryRun.TargetProfile.Root, entry.Path)
		if err != nil {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "backup_failed", Message: err.Error()}}), nil
		}
		backupAbs := filepath.Join(backupRoot, "files", filepath.FromSlash(entry.Path))
		if err := copyAbsoluteFile(source, backupAbs); err != nil {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "backup_failed", Message: err.Error()}}), nil
		}
		sum, err := checksumFile(backupAbs)
		if err != nil || entry.PreviousSHA256 == nil || sum != *entry.PreviousSHA256 {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "backup_failed", Message: "backup checksum verification failed for " + entry.Path}}), nil
		}
		backup := entry
		backup.Action = "backup"
		backup.NewSHA256 = ""
		backup.BackupPath = backupAbs
		actualChanged = append(actualChanged, backup)
		backupByPath[entry.Path] = backupAbs
	}
	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "create" && entry.Action != "update" {
			continue
		}
		content, err := renderedContentForProjectAction(dryRun, entry)
		if err != nil {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "source_render_failed", Message: err.Error()}}), nil
		}
		target, err := safeProfileWritePath(dryRun.TargetProfile.Root, entry.Path)
		if err != nil {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "unsafe_target_path", Message: err.Error()}}), nil
		}
		if err := writeFileAtomic(target, content, 0o644); err != nil {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "skill_write_failed", Message: err.Error()}}), nil
		}
		sum, err := checksumFile(target)
		if err != nil || sum != entry.NewSHA256 {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "checksum_verify_failed", Message: fmt.Sprintf("%s checksum verify failed", entry.Path)}}), nil
		}
		if backupRel := backupByPath[entry.Path]; backupRel != "" {
			entry.BackupPath = backupRel
		}
		actualChanged = append(actualChanged, entry)
	}
	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "remove" {
			continue
		}
		target, err := safeProfileWritePath(dryRun.TargetProfile.Root, entry.Path)
		if err != nil {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "unsafe_target_path", Message: err.Error()}}), nil
		}
		if err := os.Remove(target); err != nil {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "remove_failed", Message: err.Error()}}), nil
		}
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "remove_verify_failed", Message: "removed file is still present: " + entry.Path}}), nil
		}
		if backupRel := backupByPath[entry.Path]; backupRel != "" {
			entry.BackupPath = backupRel
		}
		actualChanged = append(actualChanged, entry)
	}
	manifest, err := buildUpdatedProjectManifest(projectActionAsInstallResult(dryRun), evidenceRef, approvedHash, operationID, backupRoot, previousManifestPath, actualChanged)
	if err != nil {
		return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "manifest_build_failed", Message: err.Error()}}), nil
	}
	manifestTarget, err := safeProfileWritePath(dryRun.TargetProfile.Root, ".kas/skill-pack-manifest.json")
	if err != nil {
		return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "manifest_write_failed", Message: err.Error()}}), nil
	}
	if err := writeJSONFile(manifestTarget, manifest); err != nil {
		return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, false, []discovery.Diagnostic{{Level: "error", Code: "manifest_write_failed", Message: err.Error()}}), nil
	}
	actualChanged = append(actualChanged, ChangedPath{Path: ".kas/skill-pack-manifest.json", Action: "manifest_update", PreviousSHA256: dryRun.TargetProfile.PreviousManifestSHA256})
	return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, operationID, backupRoot, actualChanged, true, nil), nil
}

func projectActionApprovalFailure(dryRun ProjectActionResult, evidenceRef string, approvedHash string, approvedMode string, code string, message string) ProjectActionResult {
	return approvedProjectActionResult(dryRun, evidenceRef, approvedHash, approvedMode, "", "", nil, false, []discovery.Diagnostic{{Level: "error", Code: code, Message: message}})
}

func approvedProjectActionResult(dryRun ProjectActionResult, evidenceRef string, approvedHash string, approvedMode string, operationID string, backupRoot string, changed []ChangedPath, ok bool, diagnostics []discovery.Diagnostic) ProjectActionResult {
	if diagnostics == nil {
		diagnostics = dryRun.Diagnostics
	} else {
		diagnostics = append(diagnostics, dryRun.Diagnostics...)
	}
	sort.Slice(changed, func(i, j int) bool {
		if changed[i].Action == changed[j].Action {
			return changed[i].Path < changed[j].Path
		}
		return changed[i].Action < changed[j].Action
	})
	counts := map[string]int{"create": 0, "update": 0, "remove": 0, "skip": 0, "conflict": 0, "error": 0, "backup": 0, "manifest_update": 0}
	for _, entry := range changed {
		counts[entry.Action]++
	}
	result := dryRun
	result.OK = ok
	result.Mode = approvedMode
	result.DryRun = false
	result.NoWrite = NoWriteEvidence{Guaranteed: false, ProfileWriteCount: counts["create"] + counts["update"] + counts["remove"], SkillWriteCount: counts["create"] + counts["update"] + counts["remove"], ManifestWriteCount: counts["manifest_update"], KASDirectoryWriteCount: counts["manifest_update"]}
	result.Summary.CountsByAction = counts
	result.Summary.TotalFiles = len(changed)
	result.Summary.DiagnosticCount = len(diagnostics)
	result.ChangedPaths = changed
	result.Approval = ApprovalEvidence{EvidenceRef: evidenceRef, DryRunPlanHash: dryRun.PlanHash, ApprovedPlanHash: approvedHash, MatchedCurrentPlan: approvedHash == dryRun.PlanHash}
	result.BackupPath = backupRoot
	result.Recovery = &Recovery{RollbackSupported: true, BackupPath: backupRoot, PreviousManifestSHA256: dryRun.TargetProfile.PreviousManifestSHA256, Instructions: []string{"Restore updated or removed files from backup_path when present.", "Restore previous manifest snapshot when present.", "Created project files may be removed manually if rollback is needed; unknown personal skills are not deleted by KASROLE-004 repair/prune."}}
	result.Diagnostics = diagnostics
	if strings.Contains(approvedMode, "repair") {
		result.RepairID = operationID
	} else {
		result.MigrationID = operationID
	}
	if ok {
		result.NextAction = "Approved changes written. Rerun doctor --profile <profile> --project <project> --project-suite to verify project suite health."
	} else {
		result.NextAction = "No approved project-suite changes were completed. Resolve diagnostics and rerun a fresh dry-run."
	}
	return result
}

func preflightProjectAction(dryRun ProjectActionResult, _ string) error {
	seenPaths := map[string]bool{}
	for _, entry := range dryRun.ChangedPaths {
		if entry.Action != "create" && entry.Action != "update" && entry.Action != "remove" && entry.Action != "skip" {
			return fmt.Errorf("plan contains non-writable action: %s %s", entry.Action, entry.Path)
		}
		if seenPaths[entry.Path] {
			return fmt.Errorf("duplicate target path: %s", entry.Path)
		}
		seenPaths[entry.Path] = true
		if _, err := safeProfileWritePath(dryRun.TargetProfile.Root, entry.Path); err != nil {
			return err
		}
		actual, readErr := existingFileSHA(filepath.Join(dryRun.TargetProfile.Root, filepath.FromSlash(entry.Path)))
		if readErr != nil {
			return fmt.Errorf("%s: %s", readErr.Code, readErr.Message)
		}
		switch entry.Action {
		case "create":
			if actual != nil {
				return fmt.Errorf("target appeared after dry-run: %s", entry.Path)
			}
		case "update", "remove", "skip":
			if actual == nil {
				return fmt.Errorf("target disappeared after dry-run: %s", entry.Path)
			}
			if entry.PreviousSHA256 == nil || *actual != *entry.PreviousSHA256 {
				return fmt.Errorf("target checksum changed after dry-run: %s", entry.Path)
			}
		}
		if entry.Action == "create" || entry.Action == "update" {
			content, err := renderedContentForProjectAction(dryRun, entry)
			if err != nil {
				return err
			}
			if sha256Bytes(content) != entry.NewSHA256 {
				return fmt.Errorf("source checksum changed after dry-run: %s", entry.Path)
			}
		}
	}
	if _, err := safeProfileWritePath(dryRun.TargetProfile.Root, ".kas/skill-pack-manifest.json"); err != nil {
		return err
	}
	return nil
}

func renderedContentForProjectAction(result ProjectActionResult, entry ChangedPath) ([]byte, error) {
	for _, component := range result.CompositionFiles {
		if component.TargetPath == entry.Path {
			return []byte(projectCompositionContent(result.Project.ID, component.Kind, projectSuiteRole{ID: result.SuiteRole, DisplayLabel: result.RoleLabel}, result.SelectedSkills)), nil
		}
	}
	for _, skill := range result.PlannedSkills {
		if skill.TargetPath != entry.Path {
			continue
		}
		pack, ok := sourcePackForPlannedSkill(result.SourceRepo.Path, skill.SourcePackID)
		if !ok {
			return nil, fmt.Errorf("missing source pack for %s", skill.SourcePackID)
		}
		return plannedSkillContent(result.SourceRepo.Path, pack, skill.SourceSkill, skill.InstalledSkill)
	}
	return nil, fmt.Errorf("no planned skill maps target path %s", entry.Path)
}

func projectActionAsInstallResult(action ProjectActionResult) Result {
	return Result{OK: action.OK, Command: action.Command, Mode: action.Mode, CLIVersion: action.CLIVersion, DryRun: action.DryRun, NoWrite: action.NoWrite, SourceRepo: action.SourceRepo, TargetProfile: action.TargetProfile, Project: action.Project, SourcePack: action.SourcePack, SuiteRole: action.SuiteRole, SuiteMode: action.SuiteMode, RoleLabel: action.RoleLabel, RoleRegistry: action.RoleRegistry, SelectedSkills: action.SelectedSkills, ExcludedSkills: action.ExcludedSkills, PlannedSkills: action.PlannedSkills, CompositionFiles: action.CompositionFiles, ChangedPaths: action.ChangedPaths, BackupPlan: action.BackupPlan, PlanHash: action.PlanHash, Diagnostics: action.Diagnostics}
}

func makeProjectActionID(command string, planHash string, now time.Time) string {
	short := strings.TrimPrefix(planHash, "sha256:")
	if len(short) > 12 {
		short = short[:12]
	}
	prefix := "kas-project-repair"
	if command == "migrate-project-kas" {
		prefix = "kas-project-migration"
	}
	return prefix + "-" + now.UTC().Format("20060102T150405Z") + "-" + short
}

func finalizeProjectAction(result *ProjectActionResult) {
	sortProjectActionResult(result)
	sortProjectActionDiagnostics(result)
	counts := map[string]int{"create": 0, "update": 0, "skip": 0, "conflict": 0, "error": 0, "backup": 0, "manifest_update": 0, "remove": 0}
	for _, changed := range result.ChangedPaths {
		counts[changed.Action]++
	}
	result.Summary = Summary{TotalSkills: len(result.PlannedSkills), TotalFiles: len(result.ChangedPaths), SelectedSkills: len(result.SelectedSkills), ExcludedSkills: len(result.ExcludedSkills), CountsByAction: counts, ConflictCount: 0, DiagnosticCount: len(result.Diagnostics)}
	result.Checksums = Checksums{SourcePack: result.SourcePack.SuiteChecksum, PlannedManifest: checksumAny(projectActionPlannedManifest(*result)), PlannedSkills: checksumAny(map[string]any{"planned_skills": result.PlannedSkills, "composition_files": result.CompositionFiles}), ChangedPaths: checksumAny(result.ChangedPaths)}
	canonical := map[string]any{"command": result.Command, "mode": result.Mode, "cli_version": result.CLIVersion, "dry_run": result.DryRun, "no_write": result.NoWrite, "source_repo": result.SourceRepo, "target_profile": result.TargetProfile, "manifest_path": result.ManifestPath, "previous_manifest_sha256": result.TargetProfile.PreviousManifestSHA256, "project": result.Project, "source_pack": result.SourcePack, "suite_role": result.SuiteRole, "suite_mode": result.SuiteMode, "role_label": result.RoleLabel, "role_registry": result.RoleRegistry, "selected_skills": result.SelectedSkills, "excluded_skills": result.ExcludedSkills, "composition_files": result.CompositionFiles, "prune_extra": result.PruneExtra, "summary": result.Summary, "project_suite_diagnostics": result.ProjectSuiteDiagnostics, "planned_actions": result.PlannedActions, "planned_skills": result.PlannedSkills, "changed_paths": result.ChangedPaths, "backup_plan": result.BackupPlan, "no_spillover": result.NoSpillover, "checksums": result.Checksums, "manual_semantic_port_tasks": result.ManualSemanticPortTasks, "diagnostics": result.Diagnostics}
	result.PlanHash = checksumAny(canonical)
	blocking := noErrorDiagnostics(result.Diagnostics)
	result.OK = blocking
	requiresApproval := result.OK && hasWritableProjectAction(result.PlannedActions)
	result.ApprovalRequest = ApprovalRequest{Required: requiresApproval, EvidenceRef: "dry-run:" + result.PlanHash, DryRunPlanHash: result.PlanHash, HashIncludesProfile: true, HashIncludesManifestState: true, HashIncludesSourceSuite: true, HashIncludesRoleFields: result.SuiteRole != "", HashIncludesNoWriteEvidence: true, HashIncludesBackupPlan: true, HashIncludesConflictsAndDiags: true}
	if !result.OK {
		result.ApprovalRequest.Required = false
		result.NextAction = "Resolve diagnostics and rerun dry-run; approved project-suite writes fail closed until the current plan is ok:true."
	} else if !requiresApproval && len(result.ManualSemanticPortTasks) > 0 {
		result.NextAction = "No deterministic approved write is available; complete manual_semantic_port_tasks without generic fallback."
	} else if !requiresApproval {
		result.NextAction = "No project-suite writes are required. Rerun doctor --project-suite when desired."
	}
}

func projectActionPlannedManifest(result ProjectActionResult) map[string]any {
	return plannedManifest(Options{Profile: result.TargetProfile.Name, Project: result.Project.ID, SuiteRole: result.SuiteRole, SourcePack: result.SourcePack.ID}, result.SourcePack.SuiteChecksum, result.PlannedSkills, projectActionRoleManifestResult(result))
}

func projectActionRoleManifestResult(action ProjectActionResult) Result {
	return Result{SuiteRole: action.SuiteRole, SuiteMode: action.SuiteMode, RoleLabel: action.RoleLabel, RoleRegistry: action.RoleRegistry, SelectedSkills: action.SelectedSkills, ExcludedSkills: action.ExcludedSkills, CompositionFiles: action.CompositionFiles}
}

func hasWritableProjectAction(actions []PlannedAction) bool {
	for _, action := range actions {
		if action.Action == "create" || action.Action == "update" || action.Action == "remove" || action.Action == "manifest_update" {
			return true
		}
	}
	return false
}

func sortProjectActionResult(result *ProjectActionResult) {
	sort.Slice(result.PlannedSkills, func(i, j int) bool { return result.PlannedSkills[i].TargetPath < result.PlannedSkills[j].TargetPath })
	sort.Slice(result.ChangedPaths, func(i, j int) bool { return result.ChangedPaths[i].Path < result.ChangedPaths[j].Path })
	sort.Slice(result.BackupPlan, func(i, j int) bool { return result.BackupPlan[i].Path < result.BackupPlan[j].Path })
	sort.Slice(result.PlannedActions, func(i, j int) bool {
		if result.PlannedActions[i].TargetPath == result.PlannedActions[j].TargetPath {
			return result.PlannedActions[i].Action < result.PlannedActions[j].Action
		}
		return result.PlannedActions[i].TargetPath < result.PlannedActions[j].TargetPath
	})
}

func sortProjectActionDiagnostics(result *ProjectActionResult) {
	sort.Slice(result.ProjectSuiteDiagnostics, func(i, j int) bool {
		left := result.ProjectSuiteDiagnostics[i]
		right := result.ProjectSuiteDiagnostics[j]
		if left.TargetPath != right.TargetPath {
			return left.TargetPath < right.TargetPath
		}
		if left.Condition != right.Condition {
			return left.Condition < right.Condition
		}
		if left.InstalledSkill != right.InstalledSkill {
			return left.InstalledSkill < right.InstalledSkill
		}
		return left.Message < right.Message
	})
	sort.Slice(result.Diagnostics, func(i, j int) bool {
		left := result.Diagnostics[i]
		right := result.Diagnostics[j]
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		if left.Level != right.Level {
			return left.Level < right.Level
		}
		return left.Message < right.Message
	})
}

func addProjectActionError(result *ProjectActionResult, code string, message string) {
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: "error", Code: code, Message: message})
}

func finalizeProjectSuiteDoctor(result *ProjectSuiteDoctorResult) {
	for _, diag := range result.ProjectSuiteDiagnostics {
		result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: diag.Severity, Code: diag.Condition, Message: diag.Message})
	}
	result.OK = !hasSuiteSeverity(result.ProjectSuiteDiagnostics, "error")
	if !result.OK {
		result.NextAction = "Review project-suite diagnostics; use approved dry-run planning only, and reserve repair/prune/profile cleanup for approved KASROLE-004 workflows when applicable."
	} else if hasSuiteSeverity(result.ProjectSuiteDiagnostics, "warning") {
		result.NextAction = "Review warnings; rerun project-suite doctor after any approved repair or migration."
	}
}

func addSuiteDoctorDiag(result *ProjectSuiteDoctorResult, diag ProjectSuiteDiagnostic) {
	result.ProjectSuiteDiagnostics = append(result.ProjectSuiteDiagnostics, diag)
}

func suiteDiag(project, installed, target, severity, condition, message, nextAction string) ProjectSuiteDiagnostic {
	return ProjectSuiteDiagnostic{Project: project, InstalledSkill: installed, TargetPath: target, Severity: severity, Condition: condition, Message: message, NextAction: nextAction}
}

func hasSuiteSeverity(diags []ProjectSuiteDiagnostic, severity string) bool {
	for _, diag := range diags {
		if diag.Severity == severity {
			return true
		}
	}
	return false
}

func unsafeManifestProjectTarget(project string, installed string, target string) bool {
	if installed == "" || !strings.HasPrefix(installed, project+"-") || discovery.IsInvalidRelativePath(target) {
		return true
	}
	want := filepath.ToSlash(filepath.Join("skills", project, installed, "SKILL.md"))
	return target != want
}
