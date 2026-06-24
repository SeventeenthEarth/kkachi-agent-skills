package pluginupdate

import (
	"fmt"
	"sort"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/install"
)

const (
	readbackOnlyAction     = "readback_only"
	metadataReadbackOnly   = "metadata_readback_only"
	suggestedDoctorCommand = "kkachi-agent-skills doctor --plugin --json"
)

type Options struct {
	Repo   string
	DryRun bool
}

type NoWriteEvidence struct {
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

type PathPlan struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	Source string `json:"source"`
}

type RoleManifestImpact struct {
	Role   string   `json:"role"`
	Path   string   `json:"path"`
	Skills []string `json:"skills"`
}

type GuideSkillImpact struct {
	SkillID     string `json:"skill_id"`
	Path        string `json:"path"`
	SourceClass string `json:"source_class"`
	Impact      string `json:"impact"`
}

type Result struct {
	OK                     bool                   `json:"ok"`
	Command                string                 `json:"command"`
	Mode                   string                 `json:"mode"`
	CLIVersion             string                 `json:"cli_version"`
	DryRun                 bool                   `json:"dry_run"`
	Namespace              string                 `json:"namespace"`
	CurrentVersion         string                 `json:"current_version"`
	CurrentSource          string                 `json:"current_source"`
	ProposedVersion        string                 `json:"proposed_version"`
	ProposedSource         string                 `json:"proposed_source"`
	PlannedChangedPaths    []PathPlan             `json:"planned_changed_paths"`
	RoleManifestImpact     []RoleManifestImpact   `json:"role_manifest_impact"`
	GuideSkillImpact       []GuideSkillImpact     `json:"guide_skill_impact"`
	NoWriteEvidence        NoWriteEvidence        `json:"no_write_evidence"`
	SuggestedDoctorCommand string                 `json:"suggested_doctor_command"`
	Diagnostics            []discovery.Diagnostic `json:"diagnostics"`
	NextAction             string                 `json:"next_action"`
}

func BuildDryRun(opts Options) (Result, error) {
	if !opts.DryRun {
		return Result{}, fmt.Errorf("plugin update currently supports --dry-run only; apply/write behavior is outside the implemented SKILL scope")
	}
	pkg, err := discovery.LoadPluginPackage(opts.Repo)
	if err != nil {
		return Result{}, err
	}
	roles, err := discovery.BuildSourceRoleManifestReadback(opts.Repo)
	if err != nil {
		return Result{}, err
	}
	guides, err := discovery.BuildSourceGuideSkillReadback(opts.Repo)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		OK:                     true,
		Command:                "update plugin",
		Mode:                   "plugin_update_dry_run",
		CLIVersion:             install.CLIVersion,
		DryRun:                 true,
		Namespace:              pkg.Namespace,
		CurrentVersion:         pkg.Version,
		CurrentSource:          pkg.ManifestPath,
		ProposedVersion:        pkg.Version,
		ProposedSource:         pkg.ManifestPath,
		NoWriteEvidence:        NoWriteEvidence{Guaranteed: true},
		SuggestedDoctorCommand: suggestedDoctorCommand,
		NextAction:             "Review plugin package readback and run the suggested doctor command only after SKILL-004 doctor behavior exists.",
	}
	result.PlannedChangedPaths = append(result.PlannedChangedPaths, PathPlan{Path: pkg.ManifestPath, Action: readbackOnlyAction, Source: "official_plugin_package"})
	for _, skill := range pkg.Skills {
		result.PlannedChangedPaths = append(result.PlannedChangedPaths, PathPlan{Path: "skills/" + skill + "/SKILL.md", Action: readbackOnlyAction, Source: "official_plugin_skill"})
	}
	for _, role := range roles.Roles {
		result.RoleManifestImpact = append(result.RoleManifestImpact, RoleManifestImpact{Role: role.Role, Path: role.SourceControlledPath, Skills: append([]string(nil), role.SkillIDs...)})
		result.PlannedChangedPaths = append(result.PlannedChangedPaths, PathPlan{Path: role.SourceControlledPath, Action: readbackOnlyAction, Source: "source_role_manifest"})
	}
	for _, guide := range guides.Guides {
		result.GuideSkillImpact = append(result.GuideSkillImpact, GuideSkillImpact{
			SkillID:     guide.SkillID,
			Path:        guide.SourceControlledPath,
			SourceClass: guide.SourceClass,
			Impact:      metadataReadbackOnly,
		})
		result.PlannedChangedPaths = append(result.PlannedChangedPaths, PathPlan{Path: guide.SourceControlledPath, Action: readbackOnlyAction, Source: guide.SourceClass})
	}
	sort.Slice(result.PlannedChangedPaths, func(i, j int) bool {
		if result.PlannedChangedPaths[i].Path == result.PlannedChangedPaths[j].Path {
			return result.PlannedChangedPaths[i].Source < result.PlannedChangedPaths[j].Source
		}
		return result.PlannedChangedPaths[i].Path < result.PlannedChangedPaths[j].Path
	})
	sort.Slice(result.GuideSkillImpact, func(i, j int) bool { return result.GuideSkillImpact[i].SkillID < result.GuideSkillImpact[j].SkillID })
	return result, nil
}
