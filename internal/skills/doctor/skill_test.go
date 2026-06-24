package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSkillClassifiesPluginWrapperOverlayPersonalAndNoWrite(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	writeSkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"), "blue")
	writeSkillDoctorOverlay(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "doksuri-blue-plan-overlay"), "doksuri", "kkachi-agent-skills:plan", "0.1.0", "")
	writeSkill(t, filepath.Join(profileRoot, "skills", "personal-note"), "personal-note", "personal")
	before, err := treeSHA256(profileRoot)
	if err != nil {
		t.Fatal(err)
	}

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot, Project: "doksuri"})
	if err != nil {
		t.Fatal(err)
	}
	after, err := treeSHA256(profileRoot)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("SKILL doctor wrote to the profile tree")
	}
	if !result.OK || result.Mode != skillDoctorMode || !result.NoWrite.Guaranteed {
		t.Fatalf("unexpected result header: %+v", result)
	}
	if result.ProvenanceContractVersion == "" || result.SourceClassEvidence == nil || result.DependencyAudit.State == "" || result.SkillDependencies == nil || result.CommandSurfaceDependencies == nil || result.DeletedBundleDiagnostics == nil {
		t.Fatalf("missing KASREL evidence fields: %+v", result)
	}
	if result.DeletedBundleReference != nil {
		t.Fatalf("deleted bundle reference should be explicit nil when absent: %+v", result.DeletedBundleReference)
	}
	counts := result.Summary.CountsBySourceClass
	for _, sourceClass := range []string{"plugin_base", "color_wrapper", "project_overlay", "personal_skill"} {
		if counts[sourceClass] == 0 {
			t.Fatalf("missing source class %s in %+v", sourceClass, counts)
		}
	}
	if !hasSkillDiag(result, "plugin_update_readiness_readback_only") || !hasSkillDiag(result, "update_surface_legacy_alias_present") {
		t.Fatalf("missing update-surface diagnostics: %+v", result.Diagnostics)
	}
	human := RenderHumanSkill(result)
	for _, want := range []string{"Status:", "SKILL plugin/wrapper/overlay doctor", "Writes:", "Next:"} {
		if !strings.Contains(human, want) {
			t.Fatalf("human output missing %q: %s", want, human)
		}
	}
}

func TestBuildSkillFailsClosedOnLegacyCopyAndMissingWrapper(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	writeSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "copied base")

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK {
		t.Fatalf("expected copied base and missing wrapper diagnostics to fail closed: %+v", result)
	}
	if !hasSkillDiag(result, "legacy_copied_base_suite_present") || !hasSkillDiag(result, "profile_skill_shadows_plugin_base") || !hasSkillDiag(result, "missing_wrapper_evidence") {
		t.Fatalf("missing fail-closed diagnostics: %+v", result.Diagnostics)
	}
	if result.Summary.CountsBySourceClass["legacy_copied_base_suite"] != 1 {
		t.Fatalf("legacy copied suite not classified: %+v", result.SourceClasses)
	}
	if !strings.Contains(result.NextAction, "Do not fall back") {
		t.Fatalf("next_action should forbid copied-suite fallback: %q", result.NextAction)
	}
}

func TestBuildSkillFailsClosedOnMissingRegisteredPluginBaseSkill(t *testing.T) {
	repo := t.TempDir()
	writeSkillDoctorPluginFixture(t, repo)
	if err := os.Remove(filepath.Join(repo, "skills", "kkachi-plan", "SKILL.md")); err != nil {
		t.Fatal(err)
	}

	result, err := BuildSkill(repo, SkillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasSkillDiag(result, "missing_plugin_base_skill") {
		t.Fatalf("expected missing plugin base diagnostic: %+v", result)
	}
	found := false
	for _, record := range result.SourceClasses {
		if record.SourceClass == "plugin_base" && record.Name == "kkachi-plan" && record.Status == "missing" {
			found = hasReason(record.ReasonCodes, "missing_plugin_base_skill")
		}
	}
	if !found {
		t.Fatalf("missing plugin base source record not marked missing: %+v", result.SourceClasses)
	}
}

func TestBuildSkillFailsClosedOnWrapperRoleManifestMismatch(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	wrapperDir := filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper")
	if err := os.MkdirAll(wrapperDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: kkachi-blue-wrapper
metadata:
  kas:
    kind: color_wrapper
    role: red_reviewer
    role_manifest: kkachi-agent-skills:roles/blue.yaml
    plugin_namespace: kkachi-agent-skills
    overlay_root: skills/<project>/kas-overlays
---
# Wrapper
`
	if err := os.WriteFile(filepath.Join(wrapperDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasSkillDiag(result, "wrapper_role_manifest_mismatch") {
		t.Fatalf("expected wrapper role mismatch diagnostic: %+v", result)
	}
}

func TestBuildSkillReportsInvalidOverlayStaleAndBoundaryDiagnostics(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	writeSkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"), "blue")
	writeSkillDoctorOverlay(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "bad-overlay"), "doksuri", "kkachi-plan", "old-version", "KAH owns plugin updates. auth_token: secret")

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot, Project: "doksuri"})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK {
		t.Fatalf("expected invalid overlay diagnostics to fail closed: %+v", result)
	}
	for _, code := range []string{"invalid_overlay_frontmatter", "stale_base_version", "overlay_runtime_config_boundary_violation", "kah_boundary_violation"} {
		if !hasSkillDiag(result, code) {
			t.Fatalf("missing diagnostic %s in %+v", code, result.Diagnostics)
		}
	}
	if result.Summary.CountsBySourceClass["project_overlay"] != 1 {
		t.Fatalf("overlay not classified: %+v", result.SourceClasses)
	}
}

func TestBuildSkillDetectsSeparateProviderAndModelOverlayKeys(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	writeSkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"), "blue")
	writeSkillDoctorOverlay(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "provider-model-overlay"), "doksuri", "kkachi-agent-skills:plan", "0.1.0", "provider: openai\nmodel: gpt-5\n")

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot, Project: "doksuri"})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasSkillDiag(result, "overlay_runtime_config_boundary_violation") {
		t.Fatalf("expected separate provider/model keys to fail closed: %+v", result)
	}
}

func TestBuildSkillProjectScopeSkipsOtherProjectOverlays(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	writeSkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"), "blue")
	writeSkillDoctorOverlay(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "doksuri-blue-plan-overlay"), "doksuri", "kkachi-agent-skills:plan", "0.1.0", "")
	writeSkillDoctorOverlay(t, filepath.Join(profileRoot, "skills", "sudal", "kas-overlays", "sudal-blue-plan-overlay"), "sudal", "kkachi-agent-skills:plan", "0.1.0", "")

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot, Project: "doksuri"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || hasSkillDiag(result, "project_overlay_out_of_scope") {
		t.Fatalf("other-project overlay should not fail requested-project doctor: %+v", result)
	}
	overlays := 0
	for _, record := range result.SourceClasses {
		if record.SourceClass == "project_overlay" {
			overlays++
			if record.Project != "doksuri" {
				t.Fatalf("unexpected non-requested overlay record: %+v", record)
			}
		}
	}
	if overlays != 1 {
		t.Fatalf("expected only requested project overlay, got %d records: %+v", overlays, result.SourceClasses)
	}
}

func TestBuildSkillFailsClosedOnProfileWalkError(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	writeSkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"), "blue")
	unreadable := filepath.Join(profileRoot, "skills", "unreadable")
	if err := os.MkdirAll(unreadable, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(unreadable, 0); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(unreadable, 0o755)
	}()

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasSkillDiag(result, "profile_unreadable") {
		t.Fatalf("expected profile walk error diagnostic: %+v", result)
	}
}

func TestBuildSkillFailsClosedOnOverlayShadowingPluginBase(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	writeSkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"), "blue")
	writeSkillDoctorOverlay(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "kkachi-plan"), "doksuri", "kkachi-agent-skills:plan", "0.1.0", "")

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot, Project: "doksuri"})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasSkillDiag(result, "project_overlay_shadows_plugin_base") {
		t.Fatalf("expected overlay shadowing diagnostic: %+v", result)
	}
}

func TestBuildSkillClassifiesUnknownKASLikeSource(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkillDoctorPluginFixture(t, repo)
	writeSkillDoctorWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"), "blue")
	if err := os.MkdirAll(filepath.Join(profileRoot, "skills", "mystery-kas"), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: mystery-kas\nmetadata:\n  kas:\n    kind: kas_unreviewed_surface\n---\n# Mystery\n"
	if err := os.WriteFile(filepath.Join(profileRoot, "skills", "mystery-kas", "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildSkill(repo, SkillOptions{Profile: "demo", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasSkillDiag(result, "unknown_source_ambiguous") {
		t.Fatalf("expected unknown source diagnostic: %+v", result)
	}
	if result.Summary.CountsBySourceClass["unknown_source"] != 1 {
		t.Fatalf("unknown source not classified: %+v", result.SourceClasses)
	}
}

func TestBuildSkillFailsClosedOnInvalidPluginEvidence(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(`name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: skill-pack.yaml
  load_policy: profile_fallback
skills:
  - kkachi-plan
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "skills", "kkachi-plan"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "skills", "kkachi-plan", "SKILL.md"), []byte("---\nname: kkachi-plan\n---\n# Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildSkill(repo, SkillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasSkillDiag(result, "missing_plugin_evidence") || !hasSkillDiag(result, "missing_role_evidence") {
		t.Fatalf("expected invalid plugin evidence diagnostics: %+v", result)
	}
}

func writeSkillDoctorPluginFixture(t *testing.T, repo string) {
	t.Helper()
	for _, skillID := range []string{"kkachi-plan", "kkachi-review"} {
		writeSkill(t, filepath.Join(repo, "skills", skillID), skillID, "base")
	}
	if err := os.MkdirAll(filepath.Join(repo, "roles"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: skill-pack.yaml
  load_policy: plugin_qualified_fail_closed
roles:
  blue: roles/blue.yaml
guides:
  - kas-overlay-doctor-guide
skills:
  - kkachi-plan
  - kkachi-review
`
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "roles", "blue.yaml"), []byte("version: kas-plugin-role-manifest/v1\nrole: blue\nskills:\n  - kkachi-plan\n  - kkachi-review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSkill(t, filepath.Join(repo, "skills", "kas-overlay-doctor-guide"), "kas-overlay-doctor-guide", "guide")
}

func writeSkillDoctorWrapper(t *testing.T, dir string, role string) {
	t.Helper()
	content := `---
name: kkachi-blue-wrapper
metadata:
  kas:
    kind: color_wrapper
    role: blue_commander
    role_manifest: kkachi-agent-skills:roles/` + role + `.yaml
    plugin_namespace: kkachi-agent-skills
    overlay_root: skills/<project>/kas-overlays
---
# Wrapper
`
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeSkillDoctorOverlay(t *testing.T, dir string, project string, overlayFor string, baseVersion string, body string) {
	t.Helper()
	content := `---
name: ` + filepath.Base(dir) + `
metadata:
  kas:
    kind: project_overlay
    project: ` + project + `
    role: blue_commander
    overlay_for: ` + overlayFor + `
    merge_mode: additive_constraints
    base_version: "` + baseVersion + `"
---
# Overlay
` + body + "\n"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasSkillDiag(result SkillDoctorResult, code string) bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func hasReason(codes []string, code string) bool {
	for _, candidate := range codes {
		if candidate == code {
			return true
		}
	}
	return false
}
