package projectinstall

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildDryRunRendersProjectPrefixedSuiteAndWritesNothing(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-plan":         "---\nname: kkachi-plan\n---\n# kkachi-plan\n",
		"kkachi-final-verify": "---\nname: kkachi-final-verify\n---\n# kkachi-final-verify\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profiles", "kwanwoo")

	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected ok dry-run, got diagnostics=%+v conflicts=%+v", result.Diagnostics, result.Conflicts)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("dry-run created profile root or unexpected stat error: %v", err)
	}
	if result.Command != "install-project-kas" || result.Mode != "project_dry_run" || !result.DryRun || !result.NoWrite.Guaranteed {
		t.Fatalf("unexpected command/no-write shape: %+v", result)
	}
	assertPlannedSkill(t, result, "kkachi-plan", "doksuri-server-plan", "skills/doksuri-server/doksuri-server-plan/SKILL.md")
	assertPlannedSkill(t, result, "kkachi-final-verify", "doksuri-server-final-verify", "skills/doksuri-server/doksuri-server-final-verify/SKILL.md")
	if result.PlannedManifest["kind"] != ManifestKind || result.PlannedManifest["profile"] != "kwanwoo" {
		t.Fatalf("unexpected manifest preview: %+v", result.PlannedManifest)
	}
	if !strings.HasPrefix(result.PlanHash, "sha256:") || !strings.HasPrefix(result.Checksums.PlannedManifest, "sha256:") || !strings.HasPrefix(result.Checksums.ChangedPaths, "sha256:") {
		t.Fatalf("public hashes must use sha256:<hex>: %+v", result.Checksums)
	}
	if result.ProjectTailoring.Mode != "dry_run_prefix_render_only" || !result.ProjectTailoring.SemanticPortRequiredBeforeApprovedInstall || result.ProjectTailoring.SemanticAdaptationClaimed {
		t.Fatalf("unexpected tailoring posture: %+v", result.ProjectTailoring)
	}

	repeated, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if repeated.PlanHash != result.PlanHash {
		t.Fatalf("plan hash was not stable: %s != %s", repeated.PlanHash, result.PlanHash)
	}
}

func TestBuildDryRunRejectsInvalidProjectAndUnknownSourcePack(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")

	invalidProject, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "../escape", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, invalidProject, "invalid_project_id")

	unknownPack, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: "missing-suite", ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, unknownPack, "unknown_source_pack")
}

func TestBuildDryRunFailClosesGenericUmbrellaOnlyAndExistingTargets(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", "doksuri-server", "kkachi-plan", "SKILL.md"), "generic")
	writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md"), "local edit")

	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, result, "generic_installed_skill_name")
	assertConflict(t, result, "existing_target_not_manifested")
	if result.OK {
		t.Fatalf("conflicted dry-run must be ok:false: %+v", result)
	}
	if result.Summary.ConflictCount < 2 || result.PlanHash == "" {
		t.Fatalf("conflicts must be summarized and hash-bound: %+v", result.Summary)
	}
}

func TestBuildDryRunFailClosesUmbrellaOnlyProfile(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-kas", "SKILL.md"), "umbrella")

	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, result, "umbrella_only")
}

func TestBuildDryRunRejectsUmbrellaOnlySourceSuite(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-kas": "---\nname: kkachi-kas\n---\n# kkachi-kas\n"})
	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: filepath.Join(t.TempDir(), "profile"), DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, result, "umbrella_only")
}

func TestPlannerSourceContainsNoWriteAPIs(t *testing.T) {
	data, err := os.ReadFile("project_install.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"os.WriteFile", "os.MkdirAll", "os.Remove", "os.Rename"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("projectinstall planner must stay read-only; found %s", forbidden)
		}
	}
}

func makeProjectInstallRepo(t *testing.T, skills map[string]string) string {
	t.Helper()
	repo := t.TempDir()
	for name, content := range skills {
		writeProjectInstallFile(t, filepath.Join(repo, "skills", name, "SKILL.md"), content)
	}
	return repo
}

func writeProjectInstallFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertPlannedSkill(t *testing.T, result Result, source string, installed string, target string) {
	t.Helper()
	for _, skill := range result.PlannedSkills {
		if skill.SourcePackID == source && skill.InstalledSkill == installed && skill.TargetPath == target {
			return
		}
	}
	t.Fatalf("missing planned skill source=%s installed=%s target=%s in %+v", source, installed, target, result.PlannedSkills)
}

func assertConflict(t *testing.T, result Result, condition string) {
	t.Helper()
	if result.OK {
		t.Fatalf("expected ok:false for condition %s", condition)
	}
	for _, conflict := range result.Conflicts {
		if conflict.Condition == condition {
			return
		}
	}
	t.Fatalf("missing conflict %s in %+v", condition, result.Conflicts)
}
