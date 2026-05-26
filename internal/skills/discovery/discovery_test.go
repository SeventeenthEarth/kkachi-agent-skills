package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func writeSkill(t *testing.T, dir string, title string, description string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	lines := []string{"---"}
	if title != "" {
		lines = append(lines, "name: "+title)
	}
	if description != "" {
		lines = append(lines, "description: "+description)
	}
	lines = append(lines, "---", "# Skill", "", "Body")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(joinLines(lines)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func joinLines(lines []string) string {
	out := ""
	for i, line := range lines {
		if i > 0 {
			out += "\n"
		}
		out += line
	}
	return out
}

func expectedPackChecksum(t *testing.T, packDir string) string {
	t.Helper()
	type entry struct {
		Path   string `json:"path"`
		Bytes  int    `json:"bytes"`
		Mode   string `json:"mode"`
		SHA256 string `json:"sha256"`
	}
	entries := []entry{}
	err := filepath.WalkDir(packDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(packDir, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		entries = append(entries, entry{
			Path:   filepath.ToSlash(rel),
			Bytes:  len(data),
			Mode:   modeString(info.Mode().Perm()),
			SHA256: hex.EncodeToString(sum[:]),
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	payload, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func TestDiscoversDirectSkillLayoutAsCorePack(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha Pack", "Alpha description")

	packs, err := DiscoverSourcePacks(repo)
	if err != nil {
		t.Fatal(err)
	}

	if len(packs) != 1 {
		t.Fatalf("packs len = %d", len(packs))
	}
	pack := packs[0]
	if pack.PackID != "alpha" || pack.Category != "core" || pack.Name != "Alpha Pack" || pack.Description != "Alpha description" || pack.SourcePath != "skills/alpha" {
		t.Fatalf("unexpected pack: %+v", pack)
	}
	if pack.Checksum != expectedPackChecksum(t, filepath.Join(repo, "skills", "alpha")) {
		t.Fatalf("unexpected checksum: %s", pack.Checksum)
	}
}

func TestDiscoversCategorySkillLayout(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "software-development", "roadmap"), "", "Roadmap work")

	packs, err := DiscoverSourcePacks(repo)
	if err != nil {
		t.Fatal(err)
	}

	if len(packs) != 1 {
		t.Fatalf("packs len = %d", len(packs))
	}
	pack := packs[0]
	if pack.PackID != "software-development/roadmap" || pack.Category != "software-development" || pack.Name != "roadmap" || pack.Description != "Roadmap work" || pack.SourcePath != "skills/software-development/roadmap" {
		t.Fatalf("unexpected pack: %+v", pack)
	}
}

func TestFailsClosedWhenNoReadablePackMetadataExists(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "skills", "empty"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := DiscoverSourcePacks(repo); err == nil {
		t.Fatal("expected discovery error")
	}
}

func TestListCategoryAndProfileStates(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "", "")
	writeSkill(t, filepath.Join(repo, "skills", "docs", "beta"), "", "Docs work")

	docs, err := BuildListResult(repo, ListOptions{Category: "docs"})
	if err != nil {
		t.Fatal(err)
	}
	if len(docs.Packs) != 1 || docs.Packs[0]["pack_id"] != "docs/beta" {
		t.Fatalf("unexpected docs result: %+v", docs.Packs)
	}

	missing, err := BuildListResult(repo, ListOptions{Category: "missing"})
	if err != nil {
		t.Fatal(err)
	}
	if len(missing.Packs) != 0 || len(missing.Diagnostics) != 1 || missing.Diagnostics[0].Code != "unknown_category" {
		t.Fatalf("unexpected missing category result: %+v", missing)
	}

	profileRoot := filepath.Join(t.TempDir(), "profile")
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"version": "0.1",
		"kind":    "kas_profile_skill_manifest",
		"installs": []map[string]any{
			{"pack_id": "alpha", "target_path": "skills/alpha", "pack_checksum": expectedPackChecksum(t, filepath.Join(repo, "skills", "alpha"))},
			{"pack_id": "docs/beta", "target_path": "skills/docs/beta", "pack_checksum": "not-current"},
		},
	}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	present, err := BuildListResult(repo, ListOptions{Profile: "demo", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	states := map[string]string{}
	for _, pack := range present.Packs {
		states[pack["pack_id"].(string)] = pack["installed_state"].(string)
	}
	if present.TargetProfile.ManifestState != "manifest_present" || states["alpha"] != "installed_current" || states["docs/beta"] != "installed_drifted" {
		t.Fatalf("unexpected profile states: %+v", present)
	}
}

func TestListRejectsNonNormalizedManifestTargetPaths(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "", "")
	writeSkill(t, filepath.Join(repo, "skills", "beta"), "", "")
	profileRoot := filepath.Join(t.TempDir(), "profile")
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(map[string]any{
		"version": "0.1",
		"kind":    "kas_profile_skill_manifest",
		"installs": []map[string]any{
			{"pack_id": "alpha", "target_path": "", "pack_checksum": "checksum"},
			{"pack_id": "beta", "target_path": "skills//beta", "pack_checksum": "checksum"},
		},
	})
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildListResult(repo, ListOptions{Profile: "demo", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}

	states := map[string]string{}
	for _, pack := range result.Packs {
		states[pack["pack_id"].(string)] = pack["installed_state"].(string)
	}
	if states["alpha"] != "conflict" || states["beta"] != "conflict" {
		t.Fatalf("unexpected states: %+v", states)
	}
}
