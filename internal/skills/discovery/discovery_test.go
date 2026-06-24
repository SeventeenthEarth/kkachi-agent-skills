package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
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

func assertNoHangul(t *testing.T, out string) {
	t.Helper()
	for _, r := range out {
		if r >= 0xAC00 && r <= 0xD7AF {
			t.Fatalf("expected no Korean prose in human output, got %q", out)
		}
	}
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

func TestEmbeddedSourceInvalidPathRejectsUnsafeNames(t *testing.T) {
	cases := map[string]bool{
		"":                true,
		".":               true,
		"/absolute":       true,
		"../escape":       true,
		"skills/../bad":   true,
		"skills\\bad":     true,
		"skills/good":     false,
		"skill-pack.yaml": false,
	}
	for path, wantInvalid := range cases {
		if got := discoveryEmbeddedInvalidPath(path); got != wantInvalid {
			t.Fatalf("discoveryEmbeddedInvalidPath(%q) = %v, want %v", path, got, wantInvalid)
		}
	}
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
	missingHuman := RenderHumanList(missing)
	assertNoHangul(t, missingHuman)
	for _, want := range []string{"Status:", "Source:", "Diagnostic:", "Next:"} {
		if !strings.Contains(missingHuman, want) {
			t.Fatalf("missing human label %q: %s", want, missingHuman)
		}
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

func TestPluginQualifiedLoadSmoke(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "Plan work")
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(`name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: skill-pack.yaml
  load_policy: plugin_qualified_fail_closed
skills:
  - kkachi-plan
`), 0o644); err != nil {
		t.Fatal(err)
	}

	readback, err := BuildPluginQualifiedSkillReadback(repo, "kkachi-agent-skills:plan")
	if err != nil {
		t.Fatal(err)
	}

	if readback.Namespace != "kkachi-agent-skills" || readback.RequestedName != "plan" || readback.CanonicalSkillID != "kkachi-plan" || readback.PluginQualifiedName != "kkachi-agent-skills:kkachi-plan" || readback.SourcePackagePath != "skill-pack.yaml" {
		t.Fatalf("unexpected plugin readback: %+v", readback)
	}
	if readback.ResolvedSkillPath != "skills/kkachi-plan/SKILL.md" || readback.PackageVersion != "0.1.0" {
		t.Fatalf("unexpected resolved path/version: %+v", readback)
	}
	if readback.ProfileWrapperSource || readback.FallbackUsed {
		t.Fatalf("plugin load must not use profile wrappers or fallback: %+v", readback)
	}
}

func TestPluginQualifiedLoadFailsClosedWithoutPluginBase(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "Plan work")
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(`name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: skill-pack.yaml
  load_policy: plugin_qualified_fail_closed
skills:
  - kkachi-review
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := BuildPluginQualifiedSkillReadback(repo, "kkachi-agent-skills:kkachi-plan"); err == nil || !strings.Contains(err.Error(), "official plugin base skill not registered") {
		t.Fatalf("expected fail-closed missing base error, got %v", err)
	}
}

func TestPluginPackageManifestValidationFailsClosed(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "missing plugin block",
			content: `name: kkachi-agent-skills
version: 0.1.0
skills:
  - kkachi-plan
`,
			want: "plugin.namespace missing",
		},
		{
			name: "invalid load policy",
			content: `name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: skill-pack.yaml
  load_policy: profile_fallback
skills:
  - kkachi-plan
`,
			want: "plugin.load_policy",
		},
		{
			name: "namespace mismatch",
			content: `name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: other-plugin
  package_manifest: skill-pack.yaml
  load_policy: plugin_qualified_fail_closed
skills:
  - kkachi-plan
`,
			want: "does not match top-level name",
		},
		{
			name: "package manifest mismatch",
			content: `name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: other.yaml
  load_policy: plugin_qualified_fail_closed
skills:
  - kkachi-plan
`,
			want: "plugin.package_manifest",
		},
		{
			name: "unsafe skill id",
			content: `name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: skill-pack.yaml
  load_policy: plugin_qualified_fail_closed
skills:
  - ../escape
`,
			want: "invalid plugin skill id",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			if err := os.MkdirAll(filepath.Join(repo, "skills"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadPluginPackage(repo); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("LoadPluginPackage error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestPluginQualifiedMalformedNamesFailClosed(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "Plan work")
	writePluginManifest(t, repo, []string{"kkachi-plan"})

	for _, qualified := range []string{
		"kkachi-agent-skills",
		"kkachi-agent-skills:",
		":kkachi-plan",
		"kkachi-agent-skills:../kkachi-plan",
		"kkachi-agent-skills:kkachi-plan:extra",
		"kkachi-agent-skills: kkachi-plan",
	} {
		if _, err := BuildPluginQualifiedSkillReadback(repo, qualified); err == nil {
			t.Fatalf("expected malformed qualified name %q to fail closed", qualified)
		}
	}
}

func TestSourceRoleManifestReadback(t *testing.T) {
	repo := t.TempDir()
	writePluginRoleFixture(t, repo)

	readback, err := BuildSourceRoleManifestReadback(repo)
	if err != nil {
		t.Fatal(err)
	}
	if readback.Namespace != "kkachi-agent-skills" || readback.PackageVersion != "0.1.0" {
		t.Fatalf("unexpected package readback: %+v", readback)
	}
	if got := readback.Roles[0].Role; got != "blue" {
		t.Fatalf("roles not deterministic, first role = %s", got)
	}
	for _, role := range readback.Roles {
		if role.SourceControlledPath == "" || role.PackageSource != "skill-pack.yaml" {
			t.Fatalf("role missing source readback: %+v", role)
		}
		if !sort.StringsAreSorted(role.SkillIDs) {
			t.Fatalf("role skills not sorted: %+v", role)
		}
	}
}

func TestSourceGuideSkillReadback(t *testing.T) {
	repo := t.TempDir()
	writePluginRoleFixture(t, repo)
	writePluginGuide(t, repo, "kas-project-overlay-guide")
	writePluginManifestWithRolesAndGuides(t, repo, []string{"kkachi-final-verify", "kkachi-plan", "kkachi-review", "kkachi-verify"}, map[string]string{
		"blue":   "roles/blue.yaml",
		"red":    "roles/red.yaml",
		"orange": "roles/orange.yaml",
		"gray":   "roles/gray.yaml",
	}, []string{"kas-project-overlay-guide"})

	readback, err := BuildSourceGuideSkillReadback(repo)
	if err != nil {
		t.Fatal(err)
	}
	if readback.Namespace != "kkachi-agent-skills" || readback.PackageVersion != "0.1.0" {
		t.Fatalf("unexpected guide package readback: %+v", readback)
	}
	assertGuideReadback(t, readback.Guides, map[string]string{
		"kas-project-overlay-guide": "skills/kas-project-overlay-guide/SKILL.md",
	})
	if readback.Guides[0].SourceClass != "official_plugin_guide" {
		t.Fatalf("unexpected guide source class: %+v", readback.Guides[0])
	}
}

func TestSourceGuideSkillReadbackFailsClosedOnMissingGuideFile(t *testing.T) {
	repo := t.TempDir()
	writePluginRoleFixture(t, repo)
	writePluginManifestWithRolesAndGuides(t, repo, []string{"kkachi-final-verify", "kkachi-plan", "kkachi-review", "kkachi-verify"}, map[string]string{
		"blue":   "roles/blue.yaml",
		"red":    "roles/red.yaml",
		"orange": "roles/orange.yaml",
		"gray":   "roles/gray.yaml",
	}, []string{"kas-project-overlay-guide"})

	if _, err := BuildSourceGuideSkillReadback(repo); err == nil || !strings.Contains(err.Error(), "official plugin guide skill file missing") {
		t.Fatalf("expected missing guide file to fail closed, got %v", err)
	}
}

func TestKASRoleManifestMappings(t *testing.T) {
	repo := t.TempDir()
	writePluginRoleFixture(t, repo)

	readback, err := BuildSourceRoleManifestReadback(repo)
	if err != nil {
		t.Fatal(err)
	}
	roles := map[string][]string{}
	for _, role := range readback.Roles {
		roles[role.Role] = role.SkillIDs
	}
	assertStringSlice(t, roles["blue"], []string{"kkachi-final-verify", "kkachi-plan", "kkachi-review", "kkachi-verify"})
	assertStringSlice(t, roles["red"], []string{"kkachi-review", "kkachi-verify"})
	assertStringSlice(t, roles["orange"], []string{"kkachi-review"})
	assertStringSlice(t, roles["gray"], []string{"kkachi-final-verify", "kkachi-review"})
}

func TestCommittedPluginPackageGuideSkillsConsistent(t *testing.T) {
	root := discoveryRepoRoot(t)
	readback, err := BuildSourceGuideSkillReadback(root)
	if err != nil {
		t.Fatal(err)
	}
	assertGuideReadback(t, readback.Guides, map[string]string{
		"kas-overlay-compose-guide": "skills/kas-overlay-compose-guide/SKILL.md",
		"kas-overlay-doctor-guide":  "skills/kas-overlay-doctor-guide/SKILL.md",
		"kas-project-overlay-guide": "skills/kas-project-overlay-guide/SKILL.md",
		"kkachi-install-guide":      "skills/kkachi-install-guide/SKILL.md",
	})
	for _, guide := range readback.Guides {
		if guide.SourceClass != "official_plugin_guide" || guide.PackageSource != PluginPackageManifestPath {
			t.Fatalf("unexpected committed guide metadata: %+v", guide)
		}
	}
}

func TestNoProfileFallbackForPluginReadback(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "kkachi-review"), "kkachi-review", "Review work")
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(`name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: skill-pack.yaml
  load_policy: plugin_qualified_fail_closed
skills:
  - kkachi-review
`), 0o644); err != nil {
		t.Fatal(err)
	}
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "stale copied profile skill")
	t.Setenv("HOME", filepath.Dir(profileRoot))

	if _, err := BuildPluginQualifiedSkillReadback(repo, "kkachi-agent-skills:kkachi-plan"); err == nil {
		t.Fatal("expected official plugin base lookup to fail instead of using profile copy")
	}
}

func TestRoleManifestValidationFailsClosed(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		role       string
		roleSkills []string
		want       string
	}{
		{name: "invalid version", version: "kas-plugin-role-manifest/v0", role: "blue", roleSkills: []string{"kkachi-plan"}, want: "unsupported role manifest version"},
		{name: "unregistered skill", version: PluginRoleManifestVersion, role: "blue", roleSkills: []string{"kkachi-unknown"}, want: "unregistered plugin skill"},
		{name: "role mismatch", version: PluginRoleManifestVersion, role: "red", roleSkills: []string{"kkachi-plan"}, want: "declares role"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			writeSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "Plan work")
			writePluginManifestWithRoles(t, repo, []string{"kkachi-plan"}, map[string]string{"blue": "roles/blue.yaml"})
			if err := os.MkdirAll(filepath.Join(repo, "roles"), 0o755); err != nil {
				t.Fatal(err)
			}
			writeRoleManifestContent(t, repo, "blue", tc.version, tc.role, tc.roleSkills)
			if _, err := BuildSourceRoleManifestReadback(repo); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("BuildSourceRoleManifestReadback error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestCommittedPluginPackageAndRoleManifestsConsistent(t *testing.T) {
	root := discoveryRepoRoot(t)
	pkg, err := LoadPluginPackage(root)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Namespace != "kkachi-agent-skills" || pkg.TopLevelName != pkg.Namespace || pkg.LoadPolicy != PluginLoadPolicy || pkg.PackageManifest != PluginPackageManifestPath {
		t.Fatalf("unexpected committed package metadata: %+v", pkg)
	}
	readback, err := BuildPluginQualifiedSkillReadback(root, "kkachi-agent-skills:plan")
	if err != nil {
		t.Fatal(err)
	}
	if readback.RequestedName != "plan" || readback.CanonicalSkillID != "kkachi-plan" || readback.PluginQualifiedName != "kkachi-agent-skills:kkachi-plan" {
		t.Fatalf("unexpected committed normalized readback: %+v", readback)
	}
	roles, err := BuildSourceRoleManifestReadback(root)
	if err != nil {
		t.Fatal(err)
	}
	roleSkills := map[string][]string{}
	for _, role := range roles.Roles {
		roleSkills[role.Role] = role.SkillIDs
	}
	assertStringSlice(t, roleSkills["blue"], pkg.Skills)
	assertStringSlice(t, roleSkills["red"], []string{"kkachi-review", "kkachi-verify"})
	assertStringSlice(t, roleSkills["orange"], []string{"kkachi-review"})
	assertStringSlice(t, roleSkills["gray"], []string{"kkachi-final-verify", "kkachi-review"})
}

func writePluginRoleFixture(t *testing.T, repo string) {
	t.Helper()
	skills := []string{"kkachi-final-verify", "kkachi-plan", "kkachi-review", "kkachi-verify"}
	for _, skill := range skills {
		writeSkill(t, filepath.Join(repo, "skills", skill), skill, skill)
	}
	writePluginManifestWithRoles(t, repo, skills, map[string]string{
		"blue":   "roles/blue.yaml",
		"red":    "roles/red.yaml",
		"orange": "roles/orange.yaml",
		"gray":   "roles/gray.yaml",
	})
	writeRoleManifest(t, repo, "blue", skills)
	writeRoleManifest(t, repo, "red", []string{"kkachi-review", "kkachi-verify"})
	writeRoleManifest(t, repo, "orange", []string{"kkachi-review"})
	writeRoleManifest(t, repo, "gray", []string{"kkachi-final-verify", "kkachi-review"})
}

func writePluginManifest(t *testing.T, repo string, skills []string) {
	t.Helper()
	writePluginManifestWithRoles(t, repo, skills, nil)
}

func writePluginManifestWithRoles(t *testing.T, repo string, skills []string, roles map[string]string) {
	t.Helper()
	writePluginManifestWithRolesAndGuides(t, repo, skills, roles, nil)
}

func writePluginManifestWithRolesAndGuides(t *testing.T, repo string, skills []string, roles map[string]string, guides []string) {
	t.Helper()
	content := "name: kkachi-agent-skills\nversion: 0.1.0\nplugin:\n  namespace: kkachi-agent-skills\n  package_manifest: skill-pack.yaml\n  load_policy: plugin_qualified_fail_closed\n"
	if len(roles) > 0 {
		content += "roles:\n"
		names := make([]string, 0, len(roles))
		for role := range roles {
			names = append(names, role)
		}
		sort.Strings(names)
		for _, role := range names {
			content += "  " + role + ": " + roles[role] + "\n"
		}
	}
	if len(guides) > 0 {
		content += "guides:\n"
		for _, guide := range guides {
			content += "  - " + guide + "\n"
		}
	}
	content += "skills:\n"
	for _, skill := range skills {
		content += "  - " + skill + "\n"
	}
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writePluginGuide(t *testing.T, repo string, guide string) {
	t.Helper()
	writeSkill(t, filepath.Join(repo, "skills", guide), guide, guide)
}

func writeRoleManifest(t *testing.T, repo string, role string, skills []string) {
	t.Helper()
	writeRoleManifestContent(t, repo, role, PluginRoleManifestVersion, role, skills)
}

func writeRoleManifestContent(t *testing.T, repo string, fileRole string, version string, role string, skills []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repo, "roles"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "version: " + version + "\nrole: " + role + "\nskills:\n"
	for _, skill := range skills {
		content += "  - " + skill + "\n"
	}
	if err := os.WriteFile(filepath.Join(repo, "roles", fileRole+".yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func discoveryRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func assertStringSlice(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func assertGuideReadback(t *testing.T, got []GuideSkillReadback, want map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got guides %+v, want %v", got, want)
	}
	for _, guide := range got {
		if wantPath, ok := want[guide.SkillID]; !ok || guide.SourceControlledPath != wantPath {
			t.Fatalf("unexpected guide readback: %+v want %v", guide, want)
		}
	}
}
