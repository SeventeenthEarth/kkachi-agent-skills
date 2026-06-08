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
	if deps, ok := pack["skill_dependencies"].([]any); !ok || len(deps) != 0 {
		t.Fatalf("KASREL-002 must leave skill_dependencies empty: %#v", pack["skill_dependencies"])
	}
	if deps, ok := pack["command_surface_dependencies"].([]any); !ok || len(deps) != 0 {
		t.Fatalf("KASREL-002 must leave command_surface_dependencies empty: %#v", pack["command_surface_dependencies"])
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
