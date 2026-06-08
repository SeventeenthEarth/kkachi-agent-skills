package doctor

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

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/install"
)

const kabBoundaryMessage = "KAB is not required for minimum CLI doctor; KAB is required for execution-runtime/code-change KAS runs."

type CommandRunner func(workDir string, args ...string) CommandResult

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type Options struct {
	Profile     string
	Project     string
	ProfileRoot string
	Runner      CommandRunner
}

type SourceRepo struct {
	Path      string  `json:"path"`
	State     string  `json:"state"`
	GitCommit *string `json:"git_commit,omitempty"`
	Dirty     *bool   `json:"dirty,omitempty"`
	PackCount int     `json:"pack_count"`
}

type TargetProfile struct {
	Name         string `json:"name"`
	Root         string `json:"root"`
	State        string `json:"state"`
	ManifestPath string `json:"manifest_path"`
}

type Manifest struct {
	Path         string `json:"path"`
	State        string `json:"state"`
	InstallCount int    `json:"install_count"`
	SHA256       string `json:"sha256,omitempty"`
	Error        string `json:"error,omitempty"`
}

type InstalledPack struct {
	PackID                     string                                     `json:"pack_id"`
	State                      string                                     `json:"state"`
	TargetPath                 string                                     `json:"target_path"`
	KABAdoptionStage           install.KABAdoptionStage                   `json:"kab_adoption_stage"`
	FilesChecked               int                                        `json:"files_checked"`
	Missing                    []string                                   `json:"missing,omitempty"`
	Drifted                    []string                                   `json:"drifted,omitempty"`
	Conflicts                  []string                                   `json:"conflicts,omitempty"`
	ChecksumSummary            string                                     `json:"checksum_summary"`
	SourceClass                discovery.SourceClass                      `json:"source_class"`
	SourceClassEvidence        []discovery.SourceClassEvidence            `json:"source_class_evidence"`
	ProvenanceState            string                                     `json:"provenance_state"`
	ManagedByKAS               bool                                       `json:"managed_by_kas"`
	ChecksumState              string                                     `json:"checksum_state,omitempty"`
	Shadowing                  []discovery.ShadowingRecord                `json:"shadowing"`
	DeletedBundleReference     any                                        `json:"deleted_bundle_reference"`
	SkillDependencies          []discovery.SkillDependencyRecord          `json:"skill_dependencies"`
	CommandSurfaceDependencies []discovery.CommandSurfaceDependencyRecord `json:"command_surface_dependencies"`
}

type KAH struct {
	Available      bool   `json:"available"`
	Version        string `json:"version,omitempty"`
	InstallCommand *bool  `json:"install_command,omitempty"`
	Capabilities   string `json:"capabilities,omitempty"`
	ProjectStatus  string `json:"project_status,omitempty"`
	ProjectDoctor  string `json:"project_doctor,omitempty"`
	ProjectPath    string `json:"project_path,omitempty"`
}

type KAB struct {
	RequiredForMinimumCLI       bool   `json:"required_for_minimum_cli"`
	RequiredForExecutionRuntime bool   `json:"required_for_execution_runtime"`
	Message                     string `json:"message"`
}

type Result struct {
	OK                        bool                              `json:"ok"`
	Command                   string                            `json:"command"`
	ProvenanceContractVersion string                            `json:"provenance_contract_version"`
	SourceRepo                SourceRepo                        `json:"source_repo"`
	TargetProfile             TargetProfile                     `json:"target_profile"`
	Manifest                  Manifest                          `json:"manifest"`
	KABAdoptionStage          install.KABAdoptionStage          `json:"kab_adoption_stage"`
	SourceInventorySummary    discovery.SourceInventorySummary  `json:"source_inventory_summary"`
	ProvenanceAudit           discovery.SourceInventorySnapshot `json:"provenance_audit"`
	DependencyAudit           discovery.DependencyAudit         `json:"dependency_audit"`
	InstalledPacks            []InstalledPack                   `json:"installed_packs"`
	KAH                       KAH                               `json:"kah"`
	KAB                       KAB                               `json:"kab"`
	Diagnostics               []discovery.Diagnostic            `json:"diagnostics"`
	NextAction                string                            `json:"next_action"`
}

func Build(repo string, opts Options) (Result, error) {
	if opts.Runner == nil {
		opts.Runner = defaultRunner
	}
	sourceRepoPath, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return Result{}, err
	}
	sourcePacks, err := discovery.DiscoverSourcePacks(sourceRepoPath)
	if err != nil {
		return Result{}, err
	}
	sourceInfo := discovery.SourceRepoInfo(sourceRepoPath)
	source := SourceRepo{
		Path:      sourceInfo.Path,
		State:     "ok",
		GitCommit: sourceInfo.GitCommit,
		Dirty:     sourceInfo.Dirty,
		PackCount: len(sourcePacks),
	}
	packsByID := map[string]discovery.SourcePack{}
	for _, pack := range sourcePacks {
		packsByID[pack.PackID] = pack
	}

	profileRoot := resolveProfileRoot(opts.Profile, opts.ProfileRoot)
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	result := Result{
		OK:                        true,
		Command:                   "doctor",
		ProvenanceContractVersion: discovery.ProvenanceContractVersion,
		SourceRepo:                source,
		TargetProfile: TargetProfile{
			Name:         opts.Profile,
			Root:         profileRoot,
			State:        "ok",
			ManifestPath: manifestPath,
		},
		Manifest:         Manifest{Path: manifestPath, State: "ok"},
		KABAdoptionStage: install.KABAdoptionStage{Applicable: false, State: "not_applicable"},
		DependencyAudit:  discovery.EmptyDependencyAudit(),
		KAB: KAB{
			RequiredForMinimumCLI:       false,
			RequiredForExecutionRuntime: true,
			Message:                     kabBoundaryMessage,
		},
	}

	if opts.Profile == "" {
		addDiag(&result, "error", "profile_required", "doctor requires --profile <profile>.")
	}
	if st, err := os.Stat(profileRoot); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			result.TargetProfile.State = "missing"
			addDiag(&result, "error", "profile_missing", "profile root does not exist: "+profileRoot)
		} else {
			result.TargetProfile.State = "unreadable"
			addDiag(&result, "error", "profile_unreadable", "profile root is not readable: "+err.Error())
		}
	} else if !st.IsDir() {
		result.TargetProfile.State = "invalid"
		addDiag(&result, "error", "profile_not_directory", "profile root is not a directory: "+profileRoot)
	}

	manifest, manifestBytes, manifestOK := readManifest(&result, manifestPath)
	manifestEntries := map[string]map[string]any{}
	if manifestOK {
		manifestEntries = manifestInstallsForProvenance(manifest)
		installed := validateInstalls(&result, manifest, profileRoot, packsByID)
		result.InstalledPacks = installed
		warnUnknownProfileSkillDirs(&result, profileRoot, installed)
	}
	discoveryProfile := &discovery.TargetProfile{Name: opts.Profile, Root: profileRoot, ManifestPath: manifestPath, ManifestState: manifestStateForInventory(result.Manifest.State)}
	inventory := discovery.BuildSourceInventory(sourcePacks, discoveryProfile, manifestEntries)
	result.SourceInventorySummary = inventory.Summary
	result.ProvenanceAudit = inventory
	result.DependencyAudit = discovery.BuildDependencyAudit(sourceRepoPath, sourcePacks, inventory)

	result.KAH = probeKAH(opts.Runner, opts.Project)
	discovery.MarkCommandSurfaceEvidence(&result.DependencyAudit, "KAH", "kkachi-agent-helper", result.KAH.Available)
	result.Diagnostics = append(result.Diagnostics, kahDiagnostics(result.KAH, opts.Project)...)
	if manifestBytes != nil {
		result.Manifest.SHA256 = shaBytes(manifestBytes)
	}
	finalize(&result)
	return result, nil
}

func RenderHuman(result Result) string {
	state := "건강"
	if !result.OK {
		state = "오류"
	} else if countLevel(result.Diagnostics, "warning") > 0 {
		state = "경고"
	}
	lines := []string{
		fmt.Sprintf("상태: %s — profile %s doctor.", state, result.TargetProfile.Name),
		fmt.Sprintf("요약: 건강 %d, 경고 %d, 오류 %d.", healthyCount(result), countLevel(result.Diagnostics, "warning"), countLevel(result.Diagnostics, "error")),
		"manifest: " + result.Manifest.State + " (" + result.Manifest.Path + ")",
		"KAB adoption marker: " + result.KABAdoptionStage.State,
		fmt.Sprintf("KAH: %s", kahHuman(result.KAH)),
		"KAB: " + result.KAB.Message,
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, fmt.Sprintf("진단[%s]: %s", diagnostic.Level, diagnostic.Message))
	}
	lines = append(lines, result.NextAction)
	return strings.Join(lines, "\n")
}

func defaultRunner(workDir string, args ...string) CommandResult {
	cmd := exec.Command("kkachi-agent-helper", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.CombinedOutput()
	return CommandResult{Stdout: out, Err: err}
}

func readManifest(result *Result, path string) (map[string]any, []byte, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			result.Manifest.State = "missing"
			result.TargetProfile.State = "manifest_missing"
			addDiag(result, "error", "manifest_missing", "KAS install manifest is missing: "+path)
		} else {
			result.Manifest.State = "unreadable"
			result.TargetProfile.State = "manifest_unreadable"
			addDiag(result, "error", "manifest_unreadable", "KAS install manifest is unreadable: "+err.Error())
		}
		return nil, nil, false
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		result.Manifest.State = "malformed"
		result.Manifest.Error = err.Error()
		result.TargetProfile.State = "manifest_error"
		addDiag(result, "error", "manifest_parse_error", "cannot parse KAS manifest: "+err.Error())
		return nil, data, false
	}
	if version, _ := manifest["version"].(string); version != install.ManifestVersion {
		result.Manifest.State = "unsupported"
		result.Manifest.Error = fmt.Sprintf("unsupported version %q", manifest["version"])
		result.TargetProfile.State = "manifest_error"
		addDiag(result, "error", "unsupported_manifest_version", fmt.Sprintf("unsupported KAS manifest version: %q", manifest["version"]))
		return nil, data, false
	}
	if kind, _ := manifest["kind"].(string); kind != install.ManifestKind {
		result.Manifest.State = "unsupported"
		result.Manifest.Error = fmt.Sprintf("unsupported kind %q", manifest["kind"])
		result.TargetProfile.State = "manifest_error"
		addDiag(result, "error", "unsupported_manifest_kind", fmt.Sprintf("unsupported KAS manifest kind: %q", manifest["kind"]))
		return nil, data, false
	}
	installs, ok := manifest["installs"].([]any)
	if !ok {
		result.Manifest.State = "malformed"
		addDiag(result, "error", "manifest_installs_invalid", "KAS manifest installs must be an array.")
		return nil, data, false
	}
	result.Manifest.InstallCount = len(installs)
	return manifest, data, true
}

func validateInstalls(result *Result, manifest map[string]any, profileRoot string, sourcePacks map[string]discovery.SourcePack) []InstalledPack {
	rawInstalls, _ := manifest["installs"].([]any)
	installed := make([]InstalledPack, 0, len(rawInstalls))
	for _, raw := range rawInstalls {
		entry, ok := raw.(map[string]any)
		if !ok {
			addDiag(result, "error", "manifest_install_invalid", "KAS manifest install entry must be an object.")
			continue
		}
		packID, _ := entry["pack_id"].(string)
		targetPath, _ := entry["target_path"].(string)
		pack := InstalledPack{PackID: packID, TargetPath: targetPath, State: "ok", ChecksumSummary: "ok"}
		if packID == "" {
			pack.State = "manifest_error"
			pack.Conflicts = append(pack.Conflicts, "pack_id")
			addDiag(result, "error", "manifest_pack_id_missing", "KAS manifest install entry is missing pack_id.")
		}
		sourcePack, sourceKnown := sourcePacks[packID]
		if packID != "" && !sourceKnown {
			pack.State = "unknown_pack"
			pack.Conflicts = append(pack.Conflicts, packID)
			addDiag(result, "error", "unknown_manifest_pack", "KAS manifest references an unknown source pack: "+packID)
		}
		expectedTarget := filepath.ToSlash(filepath.Join("skills", filepath.FromSlash(packID)))
		if discovery.IsInvalidRelativePath(targetPath) || !strings.HasPrefix(targetPath, "skills/") {
			pack.State = worstState(pack.State, "manifest_error")
			addDiag(result, "error", "unsafe_manifest_target_path", fmt.Sprintf("KAS manifest target_path is unsafe for pack %s: %q", packID, targetPath))
		} else if sourceKnown && targetPath != expectedTarget {
			pack.State = worstState(pack.State, "manifest_error")
			addDiag(result, "error", "manifest_target_path_mismatch", fmt.Sprintf("KAS manifest target_path for pack %s is %q, expected %q", packID, targetPath, expectedTarget))
		}
		if sourceKnown {
			recordedChecksum, _ := entry["pack_checksum"].(string)
			if recordedChecksum != "" && recordedChecksum != sourcePack.Checksum {
				pack.State = worstState(pack.State, "checksum_drift")
				pack.ChecksumSummary = "source_drift"
				addDiag(result, "warning", "manifest_pack_checksum_drift", "installed pack checksum differs from current source pack: "+packID)
			}
		}
		files, ok := entry["files"].([]any)
		if !ok {
			pack.State = worstState(pack.State, "manifest_error")
			addDiag(result, "error", "manifest_files_invalid", "KAS manifest files must be an array for pack: "+packID)
			pack.KABAdoptionStage = inspectKABAdoptionMarker(result, profileRoot, packID, targetPath)
			applyInstalledPackProvenance(&pack, entry, profileRoot)
			installed = append(installed, pack)
			continue
		}
		for _, rawFile := range files {
			fileEntry, ok := rawFile.(map[string]any)
			if !ok {
				pack.State = worstState(pack.State, "manifest_error")
				addDiag(result, "error", "manifest_file_invalid", "KAS manifest file entry must be an object for pack: "+packID)
				continue
			}
			rel, _ := fileEntry["relative_path"].(string)
			if discovery.IsInvalidRelativePath(rel) {
				pack.State = worstState(pack.State, "manifest_error")
				pack.Conflicts = append(pack.Conflicts, rel)
				addDiag(result, "error", "unsafe_manifest_file_path", fmt.Sprintf("KAS manifest file relative_path is unsafe for pack %s: %q", packID, rel))
				continue
			}
			expectedSHA := manifestFileSHA(fileEntry)
			if expectedSHA == "" {
				pack.State = worstState(pack.State, "manifest_error")
				addDiag(result, "error", "manifest_file_checksum_missing", "KAS manifest file checksum is missing: "+packID+"/"+rel)
				continue
			}
			target, safe := safeJoin(profileRoot, targetPath, rel)
			if !safe {
				pack.State = worstState(pack.State, "manifest_error")
				addDiag(result, "error", "unsafe_manifest_file_path", "KAS manifest file escapes profile root: "+packID+"/"+rel)
				continue
			}
			actualSHA, err := fileSHA(target)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					pack.Missing = append(pack.Missing, filepath.ToSlash(filepath.Join(targetPath, rel)))
					pack.State = worstState(pack.State, "missing_file")
					addDiag(result, "error", "installed_file_missing", "manifested installed file is missing: "+filepath.ToSlash(filepath.Join(targetPath, rel)))
				} else {
					pack.State = worstState(pack.State, "manifest_error")
					addDiag(result, "error", "installed_file_read_error", "cannot read manifested installed file: "+err.Error())
				}
				continue
			}
			pack.FilesChecked++
			if actualSHA != expectedSHA {
				pack.Drifted = append(pack.Drifted, filepath.ToSlash(filepath.Join(targetPath, rel)))
				pack.State = worstState(pack.State, "checksum_mismatch")
				pack.ChecksumSummary = "drifted"
				addDiag(result, "error", "installed_file_checksum_mismatch", "installed file checksum differs from KAS manifest: "+filepath.ToSlash(filepath.Join(targetPath, rel)))
			}
		}
		pack.KABAdoptionStage = inspectKABAdoptionMarker(result, profileRoot, packID, targetPath)
		applyInstalledPackProvenance(&pack, entry, profileRoot)
		installed = append(installed, pack)
	}
	sort.Slice(installed, func(i, j int) bool { return installed[i].PackID < installed[j].PackID })
	result.KABAdoptionStage = aggregateKABAdoptionStage(installed)
	return installed
}

func applyInstalledPackProvenance(pack *InstalledPack, manifestEntry map[string]any, profileRoot string) {
	record := discovery.ProfileManifestProvenance(pack.TargetPath, manifestEntry, discovery.ManifestPackChecksumState(manifestEntry, profileRoot))
	if pack.State == "manifest_error" || pack.State == "unknown_pack" {
		record.SourceClass = discovery.SourceUnknownUnclassified
		record.ProvenanceState = discovery.ProvenanceStateAmbiguous
		record.ManagedByKAS = false
	}
	pack.SourceClass = record.SourceClass
	pack.SourceClassEvidence = record.SourceClassEvidence
	pack.ProvenanceState = record.ProvenanceState
	pack.ManagedByKAS = record.ManagedByKAS
	pack.ChecksumState = record.ChecksumState
	pack.Shadowing = record.Shadowing
	pack.DeletedBundleReference = record.DeletedBundleReference
	pack.SkillDependencies = record.SkillDependencies
	pack.CommandSurfaceDependencies = record.CommandSurfaceDependencies
}

func manifestInstallsForProvenance(manifest map[string]any) map[string]map[string]any {
	installs := map[string]map[string]any{}
	rawInstalls, _ := manifest["installs"].([]any)
	for _, raw := range rawInstalls {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		packID, _ := entry["pack_id"].(string)
		if packID != "" {
			installs[packID] = entry
		}
	}
	return installs
}

func manifestStateForInventory(state string) string {
	switch state {
	case "ok":
		return "manifest_present"
	case "missing":
		return "manifest_missing"
	default:
		return "manifest_unreadable"
	}
}

func inspectKABAdoptionMarker(result *Result, profileRoot string, packID string, targetPath string) install.KABAdoptionStage {
	markerRel := filepath.ToSlash(filepath.Join(targetPath, install.KABAdoptionMarkerRelativePath))
	stage := install.KABAdoptionStage{Applicable: true, HashBound: true, MarkerPath: markerRel, MarkerPaths: []string{markerRel}, State: "marker_missing"}
	if packID != "" {
		stage.Markers = []install.KABAdoptionStageMarker{{PackID: packID, Path: markerRel}}
	}
	markerPath, safe := safeJoin(profileRoot, markerRel)
	if !safe {
		stage.State = "unsupported_stage"
		addDiag(result, "error", "kab_adoption_marker_unsafe_path", "KAB adoption marker path escapes profile root: "+markerRel)
		return stage
	}
	data, err := os.ReadFile(markerPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			addDiag(result, "error", "kab_adoption_marker_missing", "KAB adoption marker is missing: "+markerRel)
			return stage
		}
		stage.State = "marker_unreadable"
		addDiag(result, "error", "kab_adoption_marker_unreadable", "KAB adoption marker is unreadable: "+err.Error())
		return stage
	}
	stage.MarkerSHA256 = shaBytes(data)
	stage.Markers = []install.KABAdoptionStageMarker{{PackID: packID, Path: markerRel, SHA256: stage.MarkerSHA256}}
	parsed, ok := install.ParseKABAdoptionStageMarker(data)
	if !ok {
		stage.State = "unsupported_stage"
		addDiag(result, "error", "kab_adoption_marker_unsupported_stage", "KAB adoption marker has an unsupported or malformed stage: "+markerRel)
		return stage
	}
	parsed.State = "marker_present"
	parsed.MarkerPath = markerRel
	parsed.MarkerPaths = []string{markerRel}
	parsed.MarkerSHA256 = stage.MarkerSHA256
	parsed.Markers = stage.Markers
	return parsed
}

func aggregateKABAdoptionStage(installed []InstalledPack) install.KABAdoptionStage {
	if len(installed) == 0 {
		return install.KABAdoptionStage{Applicable: false, State: "not_applicable"}
	}
	aggregate := install.KABAdoptionStage{Applicable: true, HashBound: true, State: "marker_present"}
	for _, pack := range installed {
		stage := pack.KABAdoptionStage
		if aggregate.Numeric == 0 && stage.Numeric != 0 {
			aggregate.Numeric = stage.Numeric
			aggregate.Canonical = stage.Canonical
			aggregate.Source = stage.Source
		}
		if stage.State != "" && stage.State != "marker_present" && aggregate.State == "marker_present" {
			aggregate.State = stage.State
		}
		for _, marker := range stage.Markers {
			aggregate.Markers = append(aggregate.Markers, marker)
			aggregate.MarkerPaths = append(aggregate.MarkerPaths, marker.Path)
		}
	}
	sort.Slice(aggregate.Markers, func(i, j int) bool { return aggregate.Markers[i].Path < aggregate.Markers[j].Path })
	sort.Strings(aggregate.MarkerPaths)
	if len(aggregate.Markers) > 0 {
		aggregate.MarkerPath = aggregate.Markers[0].Path
		aggregate.MarkerSHA256 = aggregate.Markers[0].SHA256
	}
	return aggregate
}

func warnUnknownProfileSkillDirs(result *Result, profileRoot string, installed []InstalledPack) {
	known := map[string]bool{}
	for _, pack := range installed {
		if pack.TargetPath != "" {
			known[pack.TargetPath] = true
		}
	}
	skillsRoot := filepath.Join(profileRoot, "skills")
	_ = filepath.WalkDir(skillsRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return nil
		}
		if path == skillsRoot {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, "SKILL.md")); err != nil {
			return nil
		}
		rel, err := filepath.Rel(profileRoot, path)
		if err != nil {
			return nil
		}
		target := filepath.ToSlash(rel)
		if !known[target] {
			addDiag(result, "warning", "unknown_profile_skill_dir", "profile skill directory is not recorded in KAS manifest: "+target)
		}
		return filepath.SkipDir
	})
}

func probeKAH(runner CommandRunner, project string) KAH {
	kah := KAH{}
	version := runner("", "--version")
	if version.Err != nil {
		return kah
	}
	kah.Available = true
	kah.Version = strings.TrimSpace(string(version.Stdout))
	capabilities := runner("", "capabilities", "--json")
	if capabilities.Err != nil {
		kah.Capabilities = "unavailable"
	} else {
		var payload map[string]any
		if err := json.Unmarshal(capabilities.Stdout, &payload); err != nil {
			kah.Capabilities = "degraded"
		} else {
			kah.Capabilities = "ok"
			if raw, ok := payload["install_command"].(bool); ok {
				kah.InstallCommand = &raw
			}
		}
	}
	if project != "" {
		kah.ProjectPath = project
		status := runner(project, "project", "status", "--json")
		kah.ProjectStatus = commandState(status)
		doctor := runner(project, "project", "doctor", "--json")
		kah.ProjectDoctor = commandState(doctor)
	}
	return kah
}

func kahDiagnostics(kah KAH, project string) []discovery.Diagnostic {
	diagnostics := []discovery.Diagnostic{}
	if !kah.Available {
		diagnostics = append(diagnostics, discovery.Diagnostic{Level: "warning", Code: "kah_missing", Message: "kkachi-agent-helper is unavailable; KAH project checks were skipped."})
		if project != "" {
			diagnostics = append(diagnostics, discovery.Diagnostic{Level: "error", Code: "kah_project_status_skipped", Message: "cannot run KAH project status/doctor because kkachi-agent-helper is unavailable."})
		}
		return diagnostics
	}
	switch kah.Capabilities {
	case "degraded":
		diagnostics = append(diagnostics, discovery.Diagnostic{Level: "warning", Code: "kah_capabilities_degraded", Message: "kkachi-agent-helper capabilities --json did not return parseable JSON."})
	case "unavailable":
		diagnostics = append(diagnostics, discovery.Diagnostic{Level: "warning", Code: "kah_capabilities_unavailable", Message: "kkachi-agent-helper capabilities --json is unavailable."})
	}
	if kah.InstallCommand != nil && *kah.InstallCommand {
		diagnostics = append(diagnostics, discovery.Diagnostic{Level: "warning", Code: "kah_install_command_true", Message: "KAH reports install_command=true; KAS profile install ownership should be reviewed."})
	}
	if project != "" {
		if kah.ProjectStatus != "ok" {
			diagnostics = append(diagnostics, discovery.Diagnostic{Level: "error", Code: "kah_project_status_failed", Message: "KAH project status failed for: " + project})
		}
		if kah.ProjectDoctor != "ok" {
			diagnostics = append(diagnostics, discovery.Diagnostic{Level: "error", Code: "kah_project_doctor_failed", Message: "KAH project doctor failed for: " + project})
		}
	}
	return diagnostics
}

func finalize(result *Result) {
	if result.Diagnostics == nil {
		result.Diagnostics = []discovery.Diagnostic{}
	}
	if result.InstalledPacks == nil {
		result.InstalledPacks = []InstalledPack{}
	}
	result.OK = countLevel(result.Diagnostics, "error") == 0
	if result.OK {
		result.NextAction = "Profile install state is healthy. Use full KAS+KAH+KAB path for execution-runtime/code-change work."
	} else {
		result.NextAction = "프로필 설치 상태를 수정한 뒤 doctor를 다시 실행하세요. KAH project doctor는 profile install health의 증거가 아닙니다."
	}
}

func addDiag(result *Result, level string, code string, message string) {
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: level, Code: code, Message: message})
}

func countLevel(diagnostics []discovery.Diagnostic, level string) int {
	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == level {
			count++
		}
	}
	return count
}

func healthyCount(result Result) int {
	count := 0
	if result.SourceRepo.State == "ok" {
		count++
	}
	if result.TargetProfile.State == "ok" {
		count++
	}
	if result.Manifest.State == "ok" {
		count++
	}
	for _, pack := range result.InstalledPacks {
		if pack.State == "ok" {
			count++
		}
	}
	if result.KAH.Available {
		count++
	}
	return count
}

func kahHuman(kah KAH) string {
	if !kah.Available {
		return "missing"
	}
	if kah.ProjectPath != "" {
		return fmt.Sprintf("%s, project status %s, project doctor %s", kah.Version, kah.ProjectStatus, kah.ProjectDoctor)
	}
	return kah.Version
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

func manifestFileSHA(file map[string]any) string {
	for _, key := range []string{"new_sha256", "sha256"} {
		value, _ := file[key].(string)
		if value != "" {
			return value
		}
	}
	return ""
}

func safeJoin(root string, parts ...string) (string, bool) {
	all := append([]string{root}, parts...)
	target := filepath.Clean(filepath.Join(all...))
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", false
	}
	return target, true
}

func fileSHA(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("target path is a symlink: %s", path)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("target path is not a regular file: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return shaBytes(data), nil
}

func shaBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func commandState(result CommandResult) string {
	if result.Err != nil {
		return "failed"
	}
	var payload map[string]any
	if err := json.Unmarshal(result.Stdout, &payload); err != nil {
		return "degraded"
	}
	if ok, hasOK := payload["ok"].(bool); hasOK && !ok {
		return "failed"
	}
	return "ok"
}

func worstState(current string, next string) string {
	if current == "manifest_error" || next == "manifest_error" {
		return "manifest_error"
	}
	if current == "missing_file" || next == "missing_file" {
		return "missing_file"
	}
	if current == "checksum_mismatch" || next == "checksum_mismatch" {
		return "checksum_mismatch"
	}
	if current == "unknown_pack" || next == "unknown_pack" {
		return "unknown_pack"
	}
	if current == "checksum_drift" || next == "checksum_drift" {
		return "checksum_drift"
	}
	return next
}
