package projectinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
)

type manifestSkillRecord struct {
	InstalledSkill string
	TargetPath     string
	Checksum       string
	DriftPolicy    string
	TailoringMode  string
}

func discoverDefaultProjectSuite(sourceRepo string) ([]discovery.SourcePack, string, error) {
	registryPath := filepath.Join(sourceRepo, "skill-pack.yaml")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, "", fmt.Errorf("kas-default-project-suite requires readable skill-pack.yaml: %w", err)
	}
	listed := parseSkillPackYAMLSkills(string(data))
	if len(listed) == 0 {
		return nil, "", fmt.Errorf("kas-default-project-suite has no skills listed in skill-pack.yaml")
	}
	all, err := discovery.DiscoverSourcePacks(sourceRepo)
	if err != nil {
		return nil, "", err
	}
	byID := map[string]discovery.SourcePack{}
	for _, pack := range all {
		byID[pack.PackID] = pack
	}
	packs := []discovery.SourcePack{}
	for _, id := range listed {
		pack, ok := byID[id]
		if !ok {
			return nil, "", fmt.Errorf("skill-pack.yaml lists missing skill %q", id)
		}
		packs = append(packs, pack)
	}
	sort.Slice(packs, func(i, j int) bool { return packs[i].PackID < packs[j].PackID })
	sum := sha256.Sum256(data)
	return packs, "sha256:" + hex.EncodeToString(sum[:]), nil
}

func parseSkillPackYAMLSkills(text string) []string {
	lines := strings.Split(text, "\n")
	inSkills := false
	skills := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && strings.HasSuffix(trimmed, ":") {
			inSkills = trimmed == "skills:"
			continue
		}
		if !inSkills {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			value := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `"'`)
			if value != "" {
				skills = append(skills, value)
			}
			continue
		}
		if !strings.HasPrefix(line, " ") {
			inSkills = false
		}
	}
	return skills
}

func trustedProjectSuite(profileRoot string, project string, sourcePack string) (map[string]manifestSkillRecord, []Conflict, *string) {
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return map[string]manifestSkillRecord{}, nil, nil
	}
	sum := sha256.Sum256(data)
	previous := "sha256:" + hex.EncodeToString(sum[:])
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		return map[string]manifestSkillRecord{}, []Conflict{conflict("manifest_parse_error", project, "", ".kas/skill-pack-manifest.json", "cannot parse KAS manifest: "+err.Error(), "Fix the profile manifest before project install.")}, &previous
	}
	if version, _ := manifest["version"].(string); version != ManifestVersion {
		return map[string]manifestSkillRecord{}, []Conflict{conflict("unsupported_manifest_version", project, "", ".kas/skill-pack-manifest.json", fmt.Sprintf("unsupported KAS manifest version: %q", manifest["version"]), "Use a compatible KAS profile manifest before project install.")}, &previous
	}
	if kind, _ := manifest["kind"].(string); kind != ProfileManifestKind {
		return map[string]manifestSkillRecord{}, []Conflict{conflict("unsupported_manifest_kind", project, "", ".kas/skill-pack-manifest.json", fmt.Sprintf("unsupported KAS manifest kind: %q", manifest["kind"]), "Use the profile manifest kind kas_profile_skill_manifest.")}, &previous
	}
	if rawInstalls, ok := manifest["installs"].([]any); !ok && manifest["installs"] != nil {
		return map[string]manifestSkillRecord{}, []Conflict{conflict("manifest_installs_invalid", project, "", ".kas/skill-pack-manifest.json", "KAS profile manifest installs must be an array", "Fix the manifest before project install.")}, &previous
	} else if rawInstalls != nil {
		_ = rawInstalls
	}
	rawSuites, ok := manifest["project_suites"].([]any)
	if !ok {
		if manifest["project_suites"] == nil {
			return map[string]manifestSkillRecord{}, nil, &previous
		}
		return map[string]manifestSkillRecord{}, []Conflict{conflict("manifest_project_suites_invalid", project, "", ".kas/skill-pack-manifest.json", "KAS profile manifest project_suites must be an array", "Fix the manifest before project install.")}, &previous
	}
	matches := []map[string]any{}
	for _, raw := range rawSuites {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		entryProject, _ := entry["project"].(string)
		entryPack := sourcePackIDFromProjectSuite(entry)
		if entryProject == project && entryPack == sourcePack {
			matches = append(matches, entry)
		}
	}
	if len(matches) > 1 {
		return map[string]manifestSkillRecord{}, []Conflict{conflict("ambiguous_project_suite_manifest", project, "", ".kas/skill-pack-manifest.json", "multiple matching project_suites entries exist for project/source_pack", "Manually repair duplicate project suite manifest entries before approved install.")}, &previous
	}
	if len(matches) == 0 {
		return map[string]manifestSkillRecord{}, nil, &previous
	}
	suite := matches[0]
	if kind, _ := suite["kind"].(string); kind != ManifestKind {
		return map[string]manifestSkillRecord{}, []Conflict{conflict("unsupported_project_suite_manifest_kind", project, "", ".kas/skill-pack-manifest.json", fmt.Sprintf("unsupported project suite manifest kind: %q", suite["kind"]), "Repair the project suite manifest before approved install.")}, &previous
	}
	records := map[string]manifestSkillRecord{}
	rawSkills, ok := suite["installed_skills"].([]any)
	if !ok {
		return records, []Conflict{conflict("manifest_project_skills_invalid", project, "", ".kas/skill-pack-manifest.json", "project suite installed_skills must be an array", "Repair the project suite manifest before approved install.")}, &previous
	}
	for _, raw := range rawSkills {
		skill, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		target, _ := skill["target_path"].(string)
		if discovery.IsInvalidRelativePath(target) || !strings.HasPrefix(target, "skills/"+project+"/") {
			return records, []Conflict{conflict("unsafe_manifest_target_path", project, "", target, "project suite manifest contains an unsafe target_path", "Repair the project suite manifest before approved install.")}, &previous
		}
		checksum, _ := skill["checksum"].(string)
		installed, _ := skill["installed_skill"].(string)
		driftPolicy, _ := skill["drift_policy"].(string)
		tailoringMode, _ := skill["tailoring_mode"].(string)
		if records[target].TargetPath != "" {
			return records, []Conflict{conflict("duplicate_manifest_target_path", project, installed, target, "project suite manifest contains duplicate target_path", "Repair the project suite manifest before approved install.")}, &previous
		}
		records[target] = manifestSkillRecord{InstalledSkill: installed, TargetPath: target, Checksum: checksum, DriftPolicy: driftPolicy, TailoringMode: tailoringMode}
	}
	return records, nil, &previous
}

func sourcePackIDFromProjectSuite(entry map[string]any) string {
	sourcePack, _ := entry["source_pack"].(map[string]any)
	id, _ := sourcePack["id"].(string)
	return id
}

func buildUpdatedProjectManifest(dryRun Result, evidenceRef string, approvedHash string, installID string, backupRoot string, previousManifestPath string, changed []ChangedPath) (map[string]any, error) {
	existing := map[string]any{}
	data, err := os.ReadFile(dryRun.TargetProfile.ManifestPath)
	if err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return nil, err
		}
	}
	installs := []any{}
	if rawInstalls, ok := existing["installs"].([]any); ok {
		installs = rawInstalls
	}
	projectSuites := []any{}
	if rawSuites, ok := existing["project_suites"].([]any); ok {
		for _, raw := range rawSuites {
			entry, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if entry["project"] == dryRun.Project.ID && sourcePackIDFromProjectSuite(entry) == dryRun.SourcePack.ID {
				continue
			}
			projectSuites = append(projectSuites, entry)
		}
	}
	projectSuites = append(projectSuites, projectSuiteManifestEntry(dryRun, evidenceRef, approvedHash, installID, backupRoot, previousManifestPath, changed))
	sort.Slice(projectSuites, func(i, j int) bool {
		left := projectSuites[i].(map[string]any)
		right := projectSuites[j].(map[string]any)
		return fmt.Sprint(left["project"], sourcePackIDFromProjectSuite(left)) < fmt.Sprint(right["project"], sourcePackIDFromProjectSuite(right))
	})
	return map[string]any{
		"version": ManifestVersion,
		"kind":    ProfileManifestKind,
		"profile": map[string]any{
			"name": dryRun.TargetProfile.Name,
			"root": dryRun.TargetProfile.Root,
		},
		"source_repo": map[string]any{
			"path":       dryRun.SourceRepo.Path,
			"git_remote": nil,
			"git_commit": dryRun.SourceRepo.GitCommit,
			"dirty":      dryRun.SourceRepo.Dirty,
		},
		"installs":       installs,
		"project_suites": projectSuites,
	}, nil
}

func projectSuiteManifestEntry(dryRun Result, evidenceRef string, approvedHash string, installID string, backupRoot string, previousManifestPath string, changed []ChangedPath) map[string]any {
	changedByPath := map[string]ChangedPath{}
	for _, entry := range changed {
		if entry.Path != "" {
			changedByPath[entry.Path] = entry
		}
	}
	installed := []any{}
	for _, skill := range dryRun.PlannedSkills {
		changedEntry := changedByPath[skill.TargetPath]
		backupRel := any(nil)
		if changedEntry.BackupPath != "" {
			backupRel = changedEntry.BackupPath
		}
		installed = append(installed, map[string]any{
			"installed_skill":      skill.InstalledSkill,
			"source_pack_id":       skill.SourcePackID,
			"source_skill":         skill.SourceSkill,
			"target_path":          skill.TargetPath,
			"checksum":             skill.Checksum,
			"previous_sha256":      changedEntry.PreviousSHA256,
			"backup_relative_path": backupRel,
			"bytes":                skill.Bytes,
			"drift_policy":         "manual_review_required",
			"tailoring_mode":       "prefix_render_only",
		})
	}
	backupRequired := false
	for _, entry := range changed {
		if entry.Action == "backup" {
			backupRequired = true
			break
		}
	}
	return map[string]any{
		"kind":                        ManifestKind,
		"install_id":                  installID,
		"installed_at":                time.Now().UTC().Format(time.RFC3339),
		"approval_evidence_ref":       evidenceRef,
		"dry_run_plan_hash":           dryRun.PlanHash,
		"approved_plan_hash":          approvedHash,
		"project":                     dryRun.Project.ID,
		"source_pack":                 map[string]any{"id": dryRun.SourcePack.ID, "repo": "kkachi-hermes-skills", "commit": dryRun.SourceRepo.GitCommit, "checksum": dryRun.SourcePack.SuiteChecksum, "language_profile": "project-specific-prefix-render-only", "formal_registry": "skill-pack.yaml"},
		"drift_policy":                "manual_review_required",
		"semantic_adaptation_claimed": false,
		"tailoring_mode":              "prefix_render_only",
		"backup":                      map[string]any{"required": backupRequired, "path": backupRoot, "created": backupRequired},
		"previous_manifest":           map[string]any{"path": previousManifestPath, "sha256": dryRun.TargetProfile.PreviousManifestSHA256},
		"installed_skills":            installed,
	}
}
