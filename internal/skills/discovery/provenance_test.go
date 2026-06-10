package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestProvenanceSourceClassVocabularyAndSourceOnlyListDefaults(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "", "")

	result, err := BuildListResult(repo, ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if result.ProvenanceContractVersion != ProvenanceContractVersion {
		t.Fatalf("provenance contract version = %q", result.ProvenanceContractVersion)
	}
	for _, sourceClass := range []SourceClass{
		SourceBundleBuiltin,
		SourceHubInstalled,
		SourceOpsExternal,
		SourceProfilePersonal,
		SourceKASManagedProfile,
		SourceUnknownUnclassified,
	} {
		if _, ok := result.SourceInventorySummary.CountsBySourceClass[string(sourceClass)]; !ok {
			t.Fatalf("missing source class count for %s: %+v", sourceClass, result.SourceInventorySummary)
		}
	}
	if len(result.Packs) != 1 {
		t.Fatalf("packs len = %d", len(result.Packs))
	}
	pack := result.Packs[0]
	if pack["source_class"] != string(SourceUnknownUnclassified) || pack["provenance_state"] != ProvenanceStateNotApplicable {
		t.Fatalf("source-only pack should be unknown/not_applicable: %+v", pack)
	}
	if deps, ok := pack["skill_dependencies"].([]SkillDependencyRecord); !ok || len(deps) != 0 {
		t.Fatalf("KASREL-003 must expose typed empty skill_dependencies: %#v", pack["skill_dependencies"])
	}
	if deps, ok := pack["command_surface_dependencies"].([]CommandSurfaceDependencyRecord); !ok || len(deps) != 0 {
		t.Fatalf("KASREL-003 must expose typed empty command_surface_dependencies: %#v", pack["command_surface_dependencies"])
	}
}

func TestDependencyAuditUsesSkillPackRequiresAsCommandSurfaces(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "Plan", "")
	writeSkill(t, filepath.Join(repo, "skills", "kkachi-implement"), "Implement", "")
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(`name: kkachi-agent-skills
requires:
  kkachi-agent-helper: "latest"
  kkachi-agent-bridge: "latest"
skills:
  - kkachi-plan
  - kkachi-implement
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "registries"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "registries", "phase-contracts.yaml"), []byte("canonical_spine:\n  - plan\n  - implement\n  - ambiguous_phase\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	packs, err := DiscoverSourcePacks(repo)
	if err != nil {
		t.Fatal(err)
	}
	inventory := BuildSourceInventory(packs, nil, nil)
	audit := BuildDependencyAudit(repo, packs, inventory)

	if audit.State != "complete" {
		t.Fatalf("audit state = %q", audit.State)
	}
	for _, dep := range audit.SkillDependencies {
		if dep.Name == "kkachi-agent-helper" || dep.Name == "kkachi-agent-bridge" {
			t.Fatalf("command surface appeared as fake skill dependency: %+v", dep)
		}
	}
	skills := map[string]SkillDependencyRecord{}
	for _, dep := range audit.SkillDependencies {
		skills[dep.Name] = dep
	}
	if skills["kkachi-plan"].ResolutionState != "resolved" || skills["kkachi-implement"].ResolutionState != "resolved" {
		t.Fatalf("phase skill dependencies not resolved: %+v", audit.SkillDependencies)
	}
	commands := map[string]CommandSurfaceDependencyRecord{}
	for _, dep := range audit.CommandSurfaceDependencies {
		commands[dep.Surface] = dep
	}
	if commands["kkachi-agent-helper"].Owner != "KAH" || !commands["kkachi-agent-helper"].NotASkillDependency {
		t.Fatalf("KAH require not reported as command surface: %+v", commands["kkachi-agent-helper"])
	}
	if commands["kkachi-agent-bridge"].Owner != "KAB" || commands["kkachi-agent-bridge"].EvidenceState != "required_later" {
		t.Fatalf("KAB require not safely deferred: %+v", commands["kkachi-agent-bridge"])
	}
}

func TestResolveKASPackDependencyPrefersManagedProfileInventoryEvidence(t *testing.T) {
	sourcePacks := []SourcePack{{
		PackID:     "kkachi-plan",
		Name:       "Plan",
		SourcePath: "skills/kkachi-plan",
	}}
	dep := SkillDependencyRecord{
		Name:            "kkachi-plan",
		Kind:            "kas_pack",
		Required:        true,
		ResolutionState: "not_checked",
		Diagnostics:     []Diagnostic{},
	}
	inventory := SourceInventorySnapshot{
		ProfileSkills: []ProvenanceRecord{{
			SkillID:         "kkachi-plan",
			PackID:          "kkachi-plan",
			EffectivePath:   "skills/kkachi-plan",
			SourceClass:     SourceKASManagedProfile,
			ProvenanceState: ProvenanceStateClassified,
			ManagedByKAS:    true,
		}},
	}

	resolved := resolveSkillDependency(dep, sourcePacks, inventory)

	if resolved.ResolutionState != "resolved" {
		t.Fatalf("resolution state = %q", resolved.ResolutionState)
	}
	if resolved.ResolvedSourceClass != SourceKASManagedProfile || !resolved.ManagedByKAS {
		t.Fatalf("managed profile evidence not reflected: %+v", resolved)
	}
	if resolved.ResolvedPath != "skills/kkachi-plan" {
		t.Fatalf("effective profile path not preferred: %+v", resolved)
	}
}

func TestResolveKASPackDependencyKeepsSourceAvailabilityWhenProfileCopyAbsent(t *testing.T) {
	sourcePacks := []SourcePack{{
		PackID:     "kkachi-plan",
		Name:       "Plan",
		SourcePath: "skills/kkachi-plan",
	}}
	dep := SkillDependencyRecord{
		Name:            "kkachi-plan",
		Kind:            "kas_pack",
		Required:        true,
		ResolutionState: "not_checked",
		Diagnostics:     []Diagnostic{},
	}

	resolved := resolveSkillDependency(dep, sourcePacks, SourceInventorySnapshot{})

	if resolved.ResolutionState != "resolved" || resolved.ResolvedPath != "skills/kkachi-plan" {
		t.Fatalf("source repo availability should remain enough for planning: %+v", resolved)
	}
	if len(resolved.Diagnostics) != 0 {
		t.Fatalf("absent profile inventory should not create diagnostics: %+v", resolved.Diagnostics)
	}
}

func TestDeletedBundleHandlingStaysDiagnosticsOnlyWithoutFallback(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha", "")

	packs, err := DiscoverSourcePacks(repo)
	if err != nil {
		t.Fatal(err)
	}
	inventory := BuildSourceInventory(packs, nil, nil)
	audit := BuildDependencyAudit(repo, packs, inventory)

	if !hasInventoryDiagnostic(inventory, "info", "bundle_root_unavailable") {
		t.Fatalf("missing bundle-root diagnostic: %+v", inventory.Diagnostics)
	}
	if inventory.Summary.DeletedBundleReferenceCount != 0 {
		t.Fatalf("deleted bundle references must not be synthesized: %+v", inventory.Summary)
	}
	for _, record := range inventory.ProfileSkills {
		if record.DeletedBundleReference != nil {
			t.Fatalf("source-only record gained deleted-bundle fallback reference: %+v", record)
		}
		if record.SourceClass == SourceBundleBuiltin || record.SourceClass == SourceHubInstalled {
			t.Fatalf("bundle/hub fallback classification must not be inferred: %+v", record)
		}
	}
	if len(audit.DeletedBundleDiagnostics) != 0 {
		t.Fatalf("dependency audit must not synthesize deleted-bundle fallback diagnostics: %+v", audit.DeletedBundleDiagnostics)
	}
}

func TestFrontmatterDependencyListsParseOnlyKnownSimpleLists(t *testing.T) {
	repo := t.TempDir()
	packDir := filepath.Join(repo, "skills", "alpha")
	if err := os.MkdirAll(packDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: Alpha
required_skills:
  - beta
related_skills:
  - gamma
required_commands:
  - kkachi-agent-helper graph validate
required_env:
  - KAS_HOME
unknown_dependencies:
  - should-not-parse
---
# Alpha
`
	if err := os.WriteFile(filepath.Join(packDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	packs, err := DiscoverSourcePacks(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 1 {
		t.Fatalf("packs len = %d", len(packs))
	}
	if len(packs[0].SkillDependencies) != 2 {
		t.Fatalf("unexpected skill deps: %+v", packs[0].SkillDependencies)
	}
	if packs[0].SkillDependencies[0].Name != "beta" || !packs[0].SkillDependencies[0].Required {
		t.Fatalf("required skill list not parsed narrowly: %+v", packs[0].SkillDependencies)
	}
	if len(packs[0].CommandSurfaceDependencies) != 2 {
		t.Fatalf("unexpected command deps: %+v", packs[0].CommandSurfaceDependencies)
	}
	for _, dep := range packs[0].CommandSurfaceDependencies {
		if !dep.NotASkillDependency {
			t.Fatalf("frontmatter command/env dependency must not be a skill: %+v", dep)
		}
		if dep.Surface == "KAS_HOME" && dep.Owner != "environment" {
			t.Fatalf("environment dependency owner should name the actual layer: %+v", dep)
		}
	}
}

func TestProfileInventoryClassifiesManifestPersonalExternalAndShadowing(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "", "")
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkill(t, filepath.Join(profileRoot, "skills", "alpha"), "Alpha", "")
	writeSkill(t, filepath.Join(profileRoot, "skills", "personal"), "Personal", "")
	externalRoot := filepath.Join(t.TempDir(), "ops-skills")
	writeSkill(t, filepath.Join(externalRoot, "personal"), "External Personal", "")
	if err := os.WriteFile(filepath.Join(profileRoot, "config.yaml"), []byte("skills:\n  external_dirs:\n    - "+externalRoot+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(map[string]any{
		"version": "0.1",
		"kind":    "kas_profile_skill_manifest",
		"installs": []map[string]any{{
			"pack_id":       "alpha",
			"target_path":   "skills/alpha",
			"pack_checksum": "checksum",
			"files":         []map[string]any{{"relative_path": "SKILL.md", "sha256": "checksum"}},
		}},
	})
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildListResult(repo, ListOptions{Profile: "demo", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}

	alpha := map[string]any{}
	for _, pack := range result.Packs {
		if pack["pack_id"] == "alpha" {
			alpha = pack
		}
	}
	if alpha["source_class"] != string(SourceKASManagedProfile) || alpha["managed_by_kas"] != true {
		t.Fatalf("manifested profile pack not classified as KAS managed: %+v", alpha)
	}
	if result.SourceInventorySummary.CountsBySourceClass[string(SourceKASManagedProfile)] != 1 ||
		result.SourceInventorySummary.CountsBySourceClass[string(SourceProfilePersonal)] != 1 ||
		result.SourceInventorySummary.CountsBySourceClass[string(SourceOpsExternal)] != 1 ||
		result.SourceInventorySummary.ShadowingConflictCount != 1 {
		t.Fatalf("unexpected inventory summary: %+v", result.SourceInventorySummary)
	}
	if len(result.SourceInventorySnapshot.ProfileSkills) != 2 || len(result.SourceInventorySnapshot.ExternalSkills) != 1 {
		t.Fatalf("unexpected inventory snapshot: %+v", result.SourceInventorySnapshot)
	}
	var personal *ProvenanceRecord
	for i := range result.SourceInventorySnapshot.ProfileSkills {
		if result.SourceInventorySnapshot.ProfileSkills[i].SkillID == "personal" {
			personal = &result.SourceInventorySnapshot.ProfileSkills[i]
		}
	}
	if personal == nil || personal.SourceClass != SourceProfilePersonal || len(personal.Shadowing) != 1 {
		t.Fatalf("profile personal shadowing not recorded: %+v", result.SourceInventorySnapshot.ProfileSkills)
	}
}

func TestExternalDuplicateAmbiguityFailsClosedAsUnknown(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "", "")
	profileRoot := filepath.Join(t.TempDir(), "profile")
	externalOne := filepath.Join(t.TempDir(), "ops-one")
	externalTwo := filepath.Join(t.TempDir(), "ops-two")
	writeSkill(t, filepath.Join(externalOne, "shared"), "Shared", "")
	writeSkill(t, filepath.Join(externalTwo, "shared"), "Shared", "")
	if err := os.MkdirAll(profileRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	config := "skills:\n  external_dirs:\n    - " + externalOne + "\n    - " + externalTwo + "\n"
	if err := os.WriteFile(filepath.Join(profileRoot, "config.yaml"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildListResult(repo, ListOptions{Profile: "demo", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.SourceInventorySnapshot.ExternalSkills) != 2 {
		t.Fatalf("expected both ambiguous external candidates, got %+v", result.SourceInventorySnapshot.ExternalSkills)
	}
	for _, record := range result.SourceInventorySnapshot.ExternalSkills {
		if record.SkillID != "shared" || record.SourceClass != SourceUnknownUnclassified || record.ProvenanceState != ProvenanceStateAmbiguous {
			t.Fatalf("ambiguous external candidate did not fail closed: %+v", record)
		}
		if len(record.Diagnostics) != 1 || record.Diagnostics[0].Code != "source_class_ambiguous" {
			t.Fatalf("missing ambiguity diagnostic: %+v", record.Diagnostics)
		}
	}
	if result.SourceInventorySummary.CountsBySourceClass[string(SourceUnknownUnclassified)] != 2 ||
		result.SourceInventorySummary.AmbiguousCount != 2 {
		t.Fatalf("summary did not count ambiguous unknown records: %+v", result.SourceInventorySummary)
	}
	conflicts := ProvenanceConflictRecords(result.SourceInventorySnapshot)
	if len(conflicts) != 2 {
		t.Fatalf("ambiguous records not exposed as provenance conflicts: %+v", conflicts)
	}
}

func TestExternalDirsConfigDiagnosticsFailClosedForPresentInvalidConfig(t *testing.T) {
	for _, tc := range []struct {
		name      string
		file      string
		content   string
		wantCode  string
		wantLevel string
	}{
		{
			name:      "invalid_json",
			file:      "config.json",
			content:   `{"skills":{"external_dirs":[`,
			wantCode:  "external_dirs_config_invalid",
			wantLevel: "warning",
		},
		{
			name:      "invalid_yaml",
			file:      "config.yaml",
			content:   "skills:\n  external_dirs:\n    -\n",
			wantCode:  "external_dirs_config_invalid",
			wantLevel: "warning",
		},
		{
			name:      "unsupported_yaml_schema",
			file:      "config.yaml",
			content:   "skills:\n  external_dirs: [/tmp/ops-skills]\n",
			wantCode:  "external_dirs_schema_unavailable",
			wantLevel: "info",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			writeSkill(t, filepath.Join(repo, "skills", "alpha"), "", "")
			profileRoot := filepath.Join(t.TempDir(), "profile")
			if err := os.MkdirAll(profileRoot, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(profileRoot, tc.file), []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}

			result, err := BuildListResult(repo, ListOptions{Profile: "demo", ProfileRoot: profileRoot})
			if err != nil {
				t.Fatal(err)
			}

			if len(result.SourceInventorySnapshot.ExternalSkills) != 0 {
				t.Fatalf("invalid external dirs config must not infer external skills: %+v", result.SourceInventorySnapshot.ExternalSkills)
			}
			if !hasInventoryDiagnostic(result.SourceInventorySnapshot, tc.wantLevel, tc.wantCode) {
				t.Fatalf("missing diagnostic %s/%s: %+v", tc.wantLevel, tc.wantCode, result.SourceInventorySnapshot.Diagnostics)
			}
		})
	}
}

func hasInventoryDiagnostic(snapshot SourceInventorySnapshot, level string, code string) bool {
	for _, diagnostic := range snapshot.Diagnostics {
		if diagnostic.Level == level && diagnostic.Code == code {
			return true
		}
	}
	return false
}
