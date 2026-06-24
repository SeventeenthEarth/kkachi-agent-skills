package pluginupdate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildDryRunPopulatesReadbackFields(t *testing.T) {
	repo := t.TempDir()
	writePluginUpdateFixture(t, repo)

	result, err := BuildDryRun(Options{Repo: repo, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}

	if !result.OK || result.Command != "update plugin" || result.Mode != "plugin_update_dry_run" || !result.DryRun {
		t.Fatalf("unexpected dry-run result header: %+v", result)
	}
	if result.Namespace != "kkachi-agent-skills" || result.CurrentVersion != "0.1.0" || result.CurrentSource != "skill-pack.yaml" || result.ProposedVersion != result.CurrentVersion || result.ProposedSource != result.CurrentSource {
		t.Fatalf("unexpected package readback: %+v", result)
	}
	if len(result.PlannedChangedPaths) == 0 || len(result.RoleManifestImpact) != 2 || len(result.GuideSkillImpact) != 1 {
		t.Fatalf("missing impact fields: %+v", result)
	}
	if !result.NoWriteEvidence.Guaranteed ||
		result.NoWriteEvidence.ProfileWrapperWriteCount != 0 ||
		result.NoWriteEvidence.ProjectOverlayWriteCount != 0 ||
		result.NoWriteEvidence.CopiedLegacySuiteWriteCount != 0 ||
		result.NoWriteEvidence.KAHStateWriteCount != 0 ||
		result.NoWriteEvidence.KABRuntimeMutationCount != 0 ||
		result.NoWriteEvidence.HermesRuntimeMutationCount != 0 ||
		result.NoWriteEvidence.AuthProviderConfigWriteCount != 0 ||
		result.NoWriteEvidence.ProfileActivationCount != 0 {
		t.Fatalf("unexpected no-write evidence: %+v", result.NoWriteEvidence)
	}
	if result.SuggestedDoctorCommand == "" || result.NextAction == "" {
		t.Fatalf("missing operator guidance: %+v", result)
	}
	if result.Diagnostics != nil {
		t.Fatalf("SKILL-002 diagnostics should remain empty until SKILL-004, got %+v", result.Diagnostics)
	}
}

func TestBuildDryRunFailsClosedWithoutDryRun(t *testing.T) {
	repo := t.TempDir()
	writePluginUpdateFixture(t, repo)

	if _, err := BuildDryRun(Options{Repo: repo}); err == nil {
		t.Fatal("expected non-dry-run planner call to fail closed")
	}
}

func TestBuildDryRunFailsClosedOnInvalidManifest(t *testing.T) {
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

	if _, err := BuildDryRun(Options{Repo: repo, DryRun: true}); err == nil {
		t.Fatal("expected invalid plugin manifest to fail closed")
	}
}

func writePluginUpdateFixture(t *testing.T, repo string) {
	t.Helper()
	for _, skill := range []string{"kkachi-plan", "kkachi-review"} {
		writePluginUpdateSkill(t, filepath.Join(repo, "skills", skill), skill)
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
  orange: roles/orange.yaml
guides:
  - kkachi-install-guide
skills:
  - kkachi-plan
  - kkachi-review
`
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	writePluginUpdateRole(t, repo, "blue", []string{"kkachi-plan", "kkachi-review"})
	writePluginUpdateRole(t, repo, "orange", []string{"kkachi-review"})
}

func writePluginUpdateSkill(t *testing.T, dir string, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\n---\n# "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writePluginUpdateRole(t *testing.T, repo string, role string, skills []string) {
	t.Helper()
	content := "version: kas-plugin-role-manifest/v1\nrole: " + role + "\nskills:\n"
	for _, skill := range skills {
		content += "  - " + skill + "\n"
	}
	if err := os.WriteFile(filepath.Join(repo, "roles", role+".yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
