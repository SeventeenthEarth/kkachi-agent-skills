package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ProvenanceContractVersion = "kasrel-001.v1"

type SourceClass string

const (
	SourceBundleBuiltin       SourceClass = "bundle_builtin"
	SourceHubInstalled        SourceClass = "hub_installed"
	SourceOpsExternal         SourceClass = "ops_external"
	SourceProfilePersonal     SourceClass = "profile_personal"
	SourceKASManagedProfile   SourceClass = "kas_managed_profile"
	SourceUnknownUnclassified SourceClass = "unknown_or_unclassified"

	ProvenanceStateClassified    = "classified"
	ProvenanceStateAmbiguous     = "ambiguous"
	ProvenanceStateUnclassified  = "unclassified"
	ProvenanceStateNotApplicable = "not_applicable"
)

type SourceClassEvidence struct {
	Kind   string `json:"kind"`
	Path   string `json:"path,omitempty"`
	State  string `json:"state"`
	Detail string `json:"detail,omitempty"`
}

type ShadowingRecord struct {
	Name                 string                `json:"name"`
	EffectivePath        string                `json:"effective_path"`
	EffectiveSourceClass SourceClass           `json:"effective_source_class"`
	ShadowedPath         string                `json:"shadowed_path"`
	ShadowedSourceClass  SourceClass           `json:"shadowed_source_class"`
	Reason               string                `json:"reason"`
	SourceClassEvidence  []SourceClassEvidence `json:"evidence"`
}

type ProvenanceRecord struct {
	SkillID                    string                `json:"skill_id,omitempty"`
	PackID                     string                `json:"pack_id,omitempty"`
	EffectivePath              string                `json:"effective_path,omitempty"`
	SourceClass                SourceClass           `json:"source_class"`
	SourceClassEvidence        []SourceClassEvidence `json:"source_class_evidence"`
	ProvenanceState            string                `json:"provenance_state"`
	ManagedByKAS               bool                  `json:"managed_by_kas"`
	ChecksumState              string                `json:"checksum_state,omitempty"`
	Shadowing                  []ShadowingRecord     `json:"shadowing"`
	DeletedBundleReference     any                   `json:"deleted_bundle_reference"`
	Diagnostics                []Diagnostic          `json:"diagnostics"`
	SkillDependencies          []any                 `json:"skill_dependencies"`
	CommandSurfaceDependencies []any                 `json:"command_surface_dependencies"`
}

type SourceRepoPackRecord struct {
	PackID              string                `json:"pack_id"`
	Category            string                `json:"category"`
	Name                string                `json:"name"`
	SourcePath          string                `json:"source_path"`
	PackChecksum        string                `json:"pack_checksum"`
	SourceClassEvidence []SourceClassEvidence `json:"source_class_evidence"`
}

type SourceInventorySummary struct {
	CountsBySourceClass         map[string]int `json:"counts_by_source_class"`
	AmbiguousCount              int            `json:"ambiguous_count"`
	DeletedBundleReferenceCount int            `json:"deleted_bundle_reference_count"`
	ShadowingConflictCount      int            `json:"shadowing_conflict_count"`
}

type SourceInventorySnapshot struct {
	SourceRepoPacks []SourceRepoPackRecord `json:"source_repo_packs"`
	ProfileSkills   []ProvenanceRecord     `json:"profile_skills"`
	ExternalSkills  []ProvenanceRecord     `json:"external_skills"`
	Diagnostics     []Diagnostic           `json:"diagnostics"`
	Summary         SourceInventorySummary `json:"summary"`
}

type DependencyAudit struct {
	State                      string       `json:"state"`
	SkillDependencies          []any        `json:"skill_dependencies"`
	CommandSurfaceDependencies []any        `json:"command_surface_dependencies"`
	Diagnostics                []Diagnostic `json:"diagnostics"`
}

func EmptyDependencyAudit() DependencyAudit {
	return DependencyAudit{
		State:                      "not_implemented_kasrel_003",
		SkillDependencies:          []any{},
		CommandSurfaceDependencies: []any{},
		Diagnostics:                []Diagnostic{},
	}
}

func BuildSourceInventory(sourcePacks []SourcePack, targetProfile *TargetProfile, manifestInstalls map[string]map[string]any) SourceInventorySnapshot {
	snapshot := SourceInventorySnapshot{
		SourceRepoPacks: sourceRepoPackRecords(sourcePacks),
		ProfileSkills:   []ProvenanceRecord{},
		ExternalSkills:  []ProvenanceRecord{},
		Diagnostics: []Diagnostic{
			{Level: "info", Code: "bundle_root_unavailable", Message: "effective Hermes bundle root was not safely discoverable; no bundle fallback lookup was attempted"},
			{Level: "info", Code: "hub_metadata_unavailable", Message: "safe Hermes hub metadata was not discoverable; no auth/token/cache metadata was read"},
		},
	}
	if targetProfile == nil {
		for _, pack := range sourcePacks {
			snapshot.ProfileSkills = append(snapshot.ProfileSkills, SourceOnlyProvenance(pack))
		}
		snapshot.Summary = summarizeProvenance(snapshot.ProfileSkills, snapshot.ExternalSkills)
		return snapshot
	}

	manifestByTarget := map[string]map[string]any{}
	for _, entry := range manifestInstalls {
		targetPath, _ := entry["target_path"].(string)
		if targetPath != "" {
			manifestByTarget[targetPath] = entry
		}
	}
	profileRecords, diagnostics := scanProfileSkills(targetProfile.Root, manifestByTarget)
	snapshot.ProfileSkills = profileRecords
	snapshot.Diagnostics = append(snapshot.Diagnostics, diagnostics...)

	externalRecords, diagnostics := scanExternalSkills(targetProfile.Root)
	snapshot.ExternalSkills = externalRecords
	snapshot.Diagnostics = append(snapshot.Diagnostics, diagnostics...)

	applyShadowing(snapshot.ProfileSkills, snapshot.ExternalSkills)
	snapshot.Summary = summarizeProvenance(snapshot.ProfileSkills, snapshot.ExternalSkills)
	return snapshot
}

func SourceOnlyProvenance(pack SourcePack) ProvenanceRecord {
	return ProvenanceRecord{
		SkillID:                    pack.Name,
		PackID:                     pack.PackID,
		EffectivePath:              pack.SourcePath,
		SourceClass:                SourceUnknownUnclassified,
		SourceClassEvidence:        []SourceClassEvidence{{Kind: "source_repo", Path: pack.SourcePath, State: "matched", Detail: "source repo pack is not an effective Hermes profile skill"}},
		ProvenanceState:            ProvenanceStateNotApplicable,
		ManagedByKAS:               false,
		Shadowing:                  []ShadowingRecord{},
		DeletedBundleReference:     nil,
		Diagnostics:                []Diagnostic{},
		SkillDependencies:          []any{},
		CommandSurfaceDependencies: []any{},
	}
}

func ProfileManifestProvenance(targetPath string, manifestEntry map[string]any, checksumState string) ProvenanceRecord {
	packID := skillIDFromTarget(targetPath)
	if manifestEntry != nil {
		if value, _ := manifestEntry["pack_id"].(string); value != "" {
			packID = value
		}
	}
	return ProvenanceRecord{
		SkillID:       skillIDFromTarget(targetPath),
		PackID:        packID,
		EffectivePath: targetPath,
		SourceClass:   SourceKASManagedProfile,
		SourceClassEvidence: []SourceClassEvidence{
			{Kind: "profile_path", Path: targetPath, State: "matched"},
			{Kind: "kas_manifest", Path: ".kas/skill-pack-manifest.json", State: "matched"},
		},
		ProvenanceState:            ProvenanceStateClassified,
		ManagedByKAS:               true,
		ChecksumState:              checksumState,
		Shadowing:                  []ShadowingRecord{},
		DeletedBundleReference:     nil,
		Diagnostics:                []Diagnostic{},
		SkillDependencies:          []any{},
		CommandSurfaceDependencies: []any{},
	}
}

func PackProvenance(pack SourcePack, inventory SourceInventorySnapshot) ProvenanceRecord {
	targetPath := filepath.ToSlash(filepath.Join("skills", filepath.FromSlash(pack.PackID)))
	for _, record := range inventory.ProfileSkills {
		if record.PackID == pack.PackID || record.EffectivePath == targetPath {
			return record
		}
	}
	return SourceOnlyProvenance(pack)
}

func ApplyProvenance(payload map[string]any, record ProvenanceRecord) {
	payload["source_class"] = string(record.SourceClass)
	payload["source_class_evidence"] = record.SourceClassEvidence
	payload["provenance_state"] = record.ProvenanceState
	payload["managed_by_kas"] = record.ManagedByKAS
	if record.ChecksumState != "" {
		payload["checksum_state"] = record.ChecksumState
	}
	payload["shadowing"] = record.Shadowing
	payload["deleted_bundle_reference"] = record.DeletedBundleReference
	payload["diagnostics"] = record.Diagnostics
	payload["skill_dependencies"] = record.SkillDependencies
	payload["command_surface_dependencies"] = record.CommandSurfaceDependencies
}

func ProvenanceConflictRecords(inventory SourceInventorySnapshot) []ProvenanceRecord {
	records := []ProvenanceRecord{}
	for _, record := range append(append([]ProvenanceRecord{}, inventory.ProfileSkills...), inventory.ExternalSkills...) {
		if len(record.Shadowing) > 0 || record.ProvenanceState == ProvenanceStateAmbiguous || len(record.Diagnostics) > 0 {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].EffectivePath < records[j].EffectivePath })
	return records
}

func ShadowingConflicts(inventory SourceInventorySnapshot) []ShadowingRecord {
	conflicts := []ShadowingRecord{}
	for _, record := range append(append([]ProvenanceRecord{}, inventory.ProfileSkills...), inventory.ExternalSkills...) {
		conflicts = append(conflicts, record.Shadowing...)
	}
	sort.Slice(conflicts, func(i, j int) bool {
		if conflicts[i].Name == conflicts[j].Name {
			if conflicts[i].EffectivePath == conflicts[j].EffectivePath {
				return conflicts[i].ShadowedPath < conflicts[j].ShadowedPath
			}
			return conflicts[i].EffectivePath < conflicts[j].EffectivePath
		}
		return conflicts[i].Name < conflicts[j].Name
	})
	return conflicts
}

func sourceRepoPackRecords(sourcePacks []SourcePack) []SourceRepoPackRecord {
	records := make([]SourceRepoPackRecord, 0, len(sourcePacks))
	for _, pack := range sourcePacks {
		records = append(records, SourceRepoPackRecord{
			PackID:       pack.PackID,
			Category:     pack.Category,
			Name:         pack.Name,
			SourcePath:   pack.SourcePath,
			PackChecksum: pack.Checksum,
			SourceClassEvidence: []SourceClassEvidence{{
				Kind:  "source_repo",
				Path:  pack.SourcePath,
				State: "matched",
			}},
		})
	}
	return records
}

func scanProfileSkills(profileRoot string, manifestByTarget map[string]map[string]any) ([]ProvenanceRecord, []Diagnostic) {
	records := []ProvenanceRecord{}
	diagnostics := []Diagnostic{}
	skillsRoot := filepath.Join(profileRoot, "skills")
	if info, err := os.Lstat(skillsRoot); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "profile_skills_unreadable", Message: err.Error()})
		}
		return records, diagnostics
	} else if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "profile_skills_unclassified", Message: "profile skills root is not a regular directory"})
		return records, diagnostics
	}
	err := filepath.WalkDir(skillsRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "profile_skill_unreadable", Message: err.Error()})
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		if path == skillsRoot {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "profile_skill_unreadable", Message: err.Error()})
			return filepath.SkipDir
		}
		if info.Mode()&os.ModeSymlink != 0 {
			rel := safeRel(profileRoot, path)
			records = append(records, unknownRecord(skillIDFromTarget(rel), rel, "profile_path", "escaped", "profile skill directory is a symlink"))
			return filepath.SkipDir
		}
		skillPath := filepath.Join(path, "SKILL.md")
		if st, err := os.Lstat(skillPath); err != nil || st.Mode()&os.ModeSymlink != 0 || !st.Mode().IsRegular() {
			return nil
		}
		rel := safeRel(profileRoot, path)
		if rel == "" || IsInvalidRelativePath(rel) || !strings.HasPrefix(rel, "skills/") {
			records = append(records, unknownRecord(filepath.Base(path), rel, "profile_path", "escaped", "profile skill path escapes profile root"))
			return filepath.SkipDir
		}
		record := profileRecord(rel, manifestByTarget[rel])
		records = append(records, record)
		return filepath.SkipDir
	})
	if err != nil {
		diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "profile_skills_unreadable", Message: err.Error()})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].EffectivePath < records[j].EffectivePath })
	return records, diagnostics
}

func profileRecord(targetPath string, manifestEntry map[string]any) ProvenanceRecord {
	skillID := skillIDFromTarget(targetPath)
	record := ProvenanceRecord{
		SkillID:                    skillID,
		PackID:                     skillID,
		EffectivePath:              targetPath,
		SourceClass:                SourceProfilePersonal,
		SourceClassEvidence:        []SourceClassEvidence{{Kind: "profile_path", Path: targetPath, State: "matched"}},
		ProvenanceState:            ProvenanceStateClassified,
		ManagedByKAS:               false,
		ChecksumState:              "unknown",
		Shadowing:                  []ShadowingRecord{},
		DeletedBundleReference:     nil,
		Diagnostics:                []Diagnostic{},
		SkillDependencies:          []any{},
		CommandSurfaceDependencies: []any{},
	}
	if manifestEntry != nil {
		return ProfileManifestProvenance(targetPath, manifestEntry, checksumStateFromManifest(manifestEntry))
	}
	return record
}

func scanExternalSkills(profileRoot string) ([]ProvenanceRecord, []Diagnostic) {
	records := []ProvenanceRecord{}
	diagnostics := []Diagnostic{}
	externalDirs, configDiagnostics := readExternalDirs(profileRoot)
	diagnostics = append(diagnostics, configDiagnostics...)
	profileSkillsRoot := filepath.Join(profileRoot, "skills")
	for _, dir := range externalDirs {
		abs, err := filepath.Abs(dir)
		if err != nil || !filepath.IsAbs(dir) {
			diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "external_dir_unclassified", Message: "external skill directory must resolve to an absolute path"})
			continue
		}
		if pathInside(abs, profileSkillsRoot) {
			diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "external_dir_inside_profile_skills", Message: "external skill directory is inside profile-local skills and was not classified: " + abs})
			continue
		}
		if info, err := os.Lstat(abs); err != nil {
			diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "external_dir_unreadable", Message: err.Error()})
			continue
		} else if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "external_dir_unclassified", Message: "external skill directory is not a regular directory: " + abs})
			continue
		}
		found, walkDiagnostics := walkExternalDir(abs)
		records = append(records, found...)
		diagnostics = append(diagnostics, walkDiagnostics...)
	}
	markAmbiguousExternal(records)
	sort.Slice(records, func(i, j int) bool { return records[i].EffectivePath < records[j].EffectivePath })
	return records, diagnostics
}

func walkExternalDir(root string) ([]ProvenanceRecord, []Diagnostic) {
	records := []ProvenanceRecord{}
	diagnostics := []Diagnostic{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "external_skill_unreadable", Message: err.Error()})
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		if path == root {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			rel := safeRel(root, path)
			records = append(records, unknownRecord(filepath.Base(path), rel, "external_dir", "escaped", "external skill directory is unreadable or a symlink"))
			return filepath.SkipDir
		}
		skillPath := filepath.Join(path, "SKILL.md")
		if st, err := os.Lstat(skillPath); err != nil || st.Mode()&os.ModeSymlink != 0 || !st.Mode().IsRegular() {
			return nil
		}
		rel := safeRel(root, path)
		record := ProvenanceRecord{
			SkillID:                    filepath.ToSlash(rel),
			EffectivePath:              path,
			SourceClass:                SourceOpsExternal,
			SourceClassEvidence:        []SourceClassEvidence{{Kind: "external_dir", Path: root, State: "matched"}},
			ProvenanceState:            ProvenanceStateClassified,
			ManagedByKAS:               false,
			ChecksumState:              "unknown",
			Shadowing:                  []ShadowingRecord{},
			DeletedBundleReference:     nil,
			Diagnostics:                []Diagnostic{},
			SkillDependencies:          []any{},
			CommandSurfaceDependencies: []any{},
		}
		records = append(records, record)
		return filepath.SkipDir
	})
	if err != nil {
		diagnostics = append(diagnostics, Diagnostic{Level: "warning", Code: "external_dir_unreadable", Message: err.Error()})
	}
	return records, diagnostics
}

func readExternalDirs(profileRoot string) ([]string, []Diagnostic) {
	for _, name := range []string{"config.json", "config.yaml", "config.yml"} {
		path := filepath.Join(profileRoot, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, []Diagnostic{externalDirsConfigDiagnostic("warning", "external_dirs_config_unreadable", name)}
		}
		if strings.HasSuffix(name, ".json") {
			return externalDirsFromJSON(data, name)
		}
		return externalDirsFromYAMLSubset(string(data), name)
	}
	return nil, []Diagnostic{{Level: "info", Code: "external_dirs_config_unavailable", Message: "target profile external skill directories were not safely discoverable"}}
}

func externalDirsFromJSON(data []byte, name string) ([]string, []Diagnostic) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, []Diagnostic{externalDirsConfigDiagnostic("warning", "external_dirs_config_invalid", name)}
	}
	skills, ok := payload["skills"].(map[string]any)
	if !ok {
		return nil, []Diagnostic{externalDirsConfigDiagnostic("info", "external_dirs_schema_unavailable", name)}
	}
	raw, ok := skills["external_dirs"].([]any)
	if !ok {
		return nil, []Diagnostic{externalDirsConfigDiagnostic("info", "external_dirs_schema_unavailable", name)}
	}
	dirs := []string{}
	for _, item := range raw {
		dir, ok := item.(string)
		if !ok {
			return nil, []Diagnostic{externalDirsConfigDiagnostic("warning", "external_dirs_schema_unavailable", name)}
		}
		if dir != "" {
			dirs = append(dirs, expandHome(dir))
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func externalDirsFromYAMLSubset(text string, name string) ([]string, []Diagnostic) {
	dirs := []string{}
	inSkills := false
	inExternalDirs := false
	externalDirsIndent := -1
	foundExternalDirs := false
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(strings.Split(line, "#")[0])
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(line, "\t") {
			return nil, []Diagnostic{externalDirsConfigDiagnostic("warning", "external_dirs_config_invalid", name)}
		}
		if !strings.HasPrefix(line, " ") {
			inSkills = trimmed == "skills:"
			inExternalDirs = false
			if strings.HasPrefix(trimmed, "skills.external_dirs:") {
				foundExternalDirs = true
				inExternalDirs = true
				if strings.TrimSpace(strings.TrimPrefix(trimmed, "skills.external_dirs:")) != "" {
					return nil, []Diagnostic{externalDirsConfigDiagnostic("info", "external_dirs_schema_unavailable", name)}
				}
			}
			continue
		}
		if inSkills && strings.HasPrefix(trimmed, "external_dirs:") {
			foundExternalDirs = true
			inExternalDirs = true
			externalDirsIndent = leadingSpaceCount(line)
			if strings.TrimSpace(strings.TrimPrefix(trimmed, "external_dirs:")) != "" {
				return nil, []Diagnostic{externalDirsConfigDiagnostic("info", "external_dirs_schema_unavailable", name)}
			}
			continue
		}
		if inExternalDirs {
			if !strings.HasPrefix(trimmed, "- ") {
				if leadingSpaceCount(line) <= externalDirsIndent {
					inExternalDirs = false
					continue
				}
				return nil, []Diagnostic{externalDirsConfigDiagnostic("warning", "external_dirs_config_invalid", name)}
			}
			value := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `"'`)
			if value == "" {
				return nil, []Diagnostic{externalDirsConfigDiagnostic("warning", "external_dirs_config_invalid", name)}
			}
			dirs = append(dirs, expandHome(value))
		}
	}
	if !foundExternalDirs {
		return nil, []Diagnostic{externalDirsConfigDiagnostic("info", "external_dirs_schema_unavailable", name)}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func externalDirsConfigDiagnostic(level string, code string, name string) Diagnostic {
	return Diagnostic{Level: level, Code: code, Message: "target profile " + name + " did not provide safely parseable skills.external_dirs; external skill directories were not inferred"}
}

func leadingSpaceCount(line string) int {
	count := 0
	for _, ch := range line {
		if ch != ' ' {
			break
		}
		count++
	}
	return count
}

func expandHome(value string) string {
	if strings.HasPrefix(value, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(value, "~/"))
		}
	}
	return value
}

func applyShadowing(profileRecords []ProvenanceRecord, externalRecords []ProvenanceRecord) {
	externalByName := map[string][]ProvenanceRecord{}
	for _, record := range externalRecords {
		externalByName[skillNameKey(record.SkillID)] = append(externalByName[skillNameKey(record.SkillID)], record)
	}
	for i := range profileRecords {
		for _, external := range externalByName[skillNameKey(profileRecords[i].SkillID)] {
			profileRecords[i].Shadowing = append(profileRecords[i].Shadowing, ShadowingRecord{
				Name:                 skillNameKey(profileRecords[i].SkillID),
				EffectivePath:        profileRecords[i].EffectivePath,
				EffectiveSourceClass: profileRecords[i].SourceClass,
				ShadowedPath:         external.EffectivePath,
				ShadowedSourceClass:  external.SourceClass,
				Reason:               "profile_local_masks_external",
				SourceClassEvidence:  []SourceClassEvidence{{Kind: "profile_path", Path: profileRecords[i].EffectivePath, State: "matched"}, {Kind: "external_dir", Path: external.EffectivePath, State: "matched"}},
			})
		}
	}
	caseGroups := map[string][]int{}
	for i := range profileRecords {
		caseGroups[strings.ToLower(profileRecords[i].SkillID)] = append(caseGroups[strings.ToLower(profileRecords[i].SkillID)], i)
	}
	for _, indexes := range caseGroups {
		if len(indexes) < 2 {
			continue
		}
		for _, idx := range indexes {
			profileRecords[idx].ProvenanceState = ProvenanceStateAmbiguous
			profileRecords[idx].SourceClass = SourceUnknownUnclassified
			profileRecords[idx].Diagnostics = append(profileRecords[idx].Diagnostics, Diagnostic{Level: "warning", Code: "source_class_ambiguous", Message: "case-insensitive profile skill name collision"})
		}
	}
}

func markAmbiguousExternal(records []ProvenanceRecord) {
	groups := map[string][]int{}
	for i := range records {
		groups[skillNameKey(records[i].SkillID)] = append(groups[skillNameKey(records[i].SkillID)], i)
	}
	for _, indexes := range groups {
		if len(indexes) < 2 {
			continue
		}
		for _, idx := range indexes {
			records[idx].SourceClass = SourceUnknownUnclassified
			records[idx].ProvenanceState = ProvenanceStateAmbiguous
			records[idx].Diagnostics = append(records[idx].Diagnostics, Diagnostic{Level: "warning", Code: "source_class_ambiguous", Message: "multiple configured external directories contain the same skill name"})
		}
	}
}

func summarizeProvenance(recordGroups ...[]ProvenanceRecord) SourceInventorySummary {
	summary := SourceInventorySummary{CountsBySourceClass: emptySourceClassCounts()}
	for _, records := range recordGroups {
		for _, record := range records {
			summary.CountsBySourceClass[string(record.SourceClass)]++
			if record.ProvenanceState == ProvenanceStateAmbiguous {
				summary.AmbiguousCount++
			}
			if record.DeletedBundleReference != nil {
				summary.DeletedBundleReferenceCount++
			}
			summary.ShadowingConflictCount += len(record.Shadowing)
		}
	}
	return summary
}

func emptySourceClassCounts() map[string]int {
	return map[string]int{
		string(SourceBundleBuiltin):       0,
		string(SourceHubInstalled):        0,
		string(SourceOpsExternal):         0,
		string(SourceProfilePersonal):     0,
		string(SourceKASManagedProfile):   0,
		string(SourceUnknownUnclassified): 0,
	}
}

func unknownRecord(skillID string, effectivePath string, kind string, state string, message string) ProvenanceRecord {
	return ProvenanceRecord{
		SkillID:                    skillID,
		EffectivePath:              effectivePath,
		SourceClass:                SourceUnknownUnclassified,
		SourceClassEvidence:        []SourceClassEvidence{{Kind: kind, Path: effectivePath, State: state}},
		ProvenanceState:            ProvenanceStateAmbiguous,
		ManagedByKAS:               false,
		ChecksumState:              "unknown",
		Shadowing:                  []ShadowingRecord{},
		DeletedBundleReference:     nil,
		Diagnostics:                []Diagnostic{{Level: "warning", Code: "source_class_ambiguous", Message: message}},
		SkillDependencies:          []any{},
		CommandSurfaceDependencies: []any{},
	}
}

func checksumStateFromManifest(entry map[string]any) string {
	files, ok := entry["files"].([]any)
	if !ok || len(files) == 0 {
		if checksum, _ := entry["pack_checksum"].(string); checksum != "" {
			return "unknown"
		}
		return "missing"
	}
	for _, raw := range files {
		file, ok := raw.(map[string]any)
		if !ok || manifestFileChecksum(file) == "" {
			return "missing"
		}
	}
	return "unknown"
}

func ManifestPackChecksumState(entry map[string]any, profileRoot string) string {
	targetPath, _ := entry["target_path"].(string)
	files, ok := entry["files"].([]any)
	if !ok || len(files) == 0 || IsInvalidRelativePath(targetPath) {
		return checksumStateFromManifest(entry)
	}
	checked := 0
	for _, raw := range files {
		file, ok := raw.(map[string]any)
		if !ok {
			return "unknown"
		}
		rel, _ := file["relative_path"].(string)
		expected := manifestFileChecksum(file)
		if IsInvalidRelativePath(rel) || expected == "" {
			return "missing"
		}
		path := filepath.Join(profileRoot, filepath.FromSlash(targetPath), filepath.FromSlash(rel))
		actual, err := regularFileSHA(path)
		if err != nil {
			return "missing"
		}
		checked++
		if actual != expected {
			return "drifted"
		}
	}
	if checked == 0 {
		return "unknown"
	}
	return "matched"
}

func manifestFileChecksum(file map[string]any) string {
	for _, key := range []string{"new_sha256", "sha256"} {
		value, _ := file[key].(string)
		if value != "" {
			return value
		}
	}
	return ""
}

func regularFileSHA(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func safeRel(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	if IsInvalidRelativePath(rel) {
		return ""
	}
	return rel
}

func skillIDFromTarget(targetPath string) string {
	return strings.TrimPrefix(targetPath, "skills/")
}

func skillNameKey(skillID string) string {
	return strings.ToLower(filepath.Base(filepath.FromSlash(skillID)))
}

func pathInside(path string, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." && !filepath.IsAbs(rel)
}
