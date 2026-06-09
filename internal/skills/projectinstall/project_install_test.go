package projectinstall

import (
	"encoding/json"
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
	if result.PlannedManifest["kind"] != ProfileManifestKind {
		t.Fatalf("unexpected manifest preview: %+v", result.PlannedManifest)
	}
	if !strings.HasPrefix(result.PlanHash, "sha256:") || !strings.HasPrefix(result.Checksums.PlannedManifest, "sha256:") || !strings.HasPrefix(result.Checksums.ChangedPaths, "sha256:") {
		t.Fatalf("public hashes must use sha256:<hex>: %+v", result.Checksums)
	}
	if result.SourcePack.FormalRegistry != "skill-pack.yaml" {
		t.Fatalf("source suite was not formalized through skill-pack.yaml: %+v", result.SourcePack)
	}
	if result.ProjectTailoring.Mode != "prefix_render_only" || result.ProjectTailoring.SemanticPortRequiredBeforeApprovedInstall || result.ProjectTailoring.SemanticAdaptationClaimed {
		t.Fatalf("unexpected tailoring posture: %+v", result.ProjectTailoring)
	}
	if result.ApprovalRequest.EvidenceRef != "dry-run:"+result.PlanHash || !result.ApprovalRequest.HashIncludesProfile || !result.ApprovalRequest.HashIncludesNoWriteEvidence || !result.ApprovalRequest.HashIncludesBackupPlan {
		t.Fatalf("approval request is not hash-bound enough: %+v", result.ApprovalRequest)
	}

	repeated, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if repeated.PlanHash != result.PlanHash {
		t.Fatalf("plan hash was not stable: %s != %s", repeated.PlanHash, result.PlanHash)
	}
}

func TestApplyApprovedInstallNoWriteOnMismatchAndSuccessfulCreate(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-plan":         "---\nname: kkachi-plan\n---\n# kkachi-plan\n",
		"kkachi-final-verify": "---\nname: kkachi-final-verify\n---\n# kkachi-final-verify\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profiles", "kwanwoo")
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	wrong, err := ApplyApprovedInstall(repo, opts, "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if wrong.OK || wrong.Approval.MatchedCurrentPlan {
		t.Fatalf("wrong hash should fail closed: %+v", wrong)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("wrong hash wrote profile root: %v", err)
	}

	approved, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if !approved.OK || approved.Mode != "project_approved_copy" || approved.InstallID == "" || approved.ManifestPath == "" || approved.Recovery == nil {
		t.Fatalf("approved install missing evidence: %+v", approved)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); err != nil {
		t.Fatalf("approved install did not write project skill: %v", err)
	}
	manifest := readProjectInstallManifest(t, filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))
	if manifest["kind"] != ProfileManifestKind {
		t.Fatalf("manifest did not preserve profile kind: %+v", manifest)
	}
	if len(manifest["installs"].([]any)) != 0 {
		t.Fatalf("create install should preserve empty generic installs: %+v", manifest)
	}
	suite := manifest["project_suites"].([]any)[0].(map[string]any)
	if suite["kind"] != ManifestKind || suite["project"] != "doksuri-server" || suite["semantic_adaptation_claimed"] != false || suite["drift_policy"] != "manual_review_required" {
		t.Fatalf("unexpected project suite manifest: %+v", suite)
	}
}

func TestApplyApprovedInstallRejectsKASSymlinkBeforeSkillWrites(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	if err := os.MkdirAll(profileRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(profileRoot, ".kas")); err != nil {
		t.Fatal(err)
	}
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK {
		t.Fatalf("dry-run should be ok before approved-write preflight: %+v", dryRun)
	}

	approved, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if approved.OK || firstDiagnosticCode(approved) != "project_install_preflight_failed" {
		t.Fatalf("approved install should fail closed on .kas symlink: %+v", approved)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("approved install wrote skill before manifest preflight failure or unexpected stat error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("approved install wrote manifest through symlink or unexpected stat error: %v", err)
	}
}

func TestApplyApprovedInstallTrustedUpdateBacksUpAndLocalModificationConflicts(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\nold\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	first, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !first.OK {
		t.Fatalf("initial install failed: result=%+v err=%v", first, err)
	}
	writeProjectInstallFile(t, filepath.Join(repo, "skills", "kkachi-plan", "SKILL.md"), "---\nname: kkachi-plan\n---\n# kkachi-plan\nnew\n")
	writeSkillPackYAML(t, repo, []string{"kkachi-plan"})
	updatePlan, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	if updatePlan.Summary.CountsByAction["update"] != 1 || len(updatePlan.BackupPlan) != 1 {
		t.Fatalf("expected trusted update with backup plan: %+v", updatePlan)
	}
	updated, err := ApplyApprovedInstall(repo, opts, updatePlan.ApprovalRequest.EvidenceRef)
	if err != nil || !updated.OK {
		t.Fatalf("trusted update failed: result=%+v err=%v", updated, err)
	}
	if updated.Summary.CountsByAction["backup"] != 1 {
		t.Fatalf("trusted update did not record backup: %+v", updated.Summary)
	}
	if _, err := os.Stat(filepath.Join(updated.BackupPath, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); err != nil {
		t.Fatalf("backup file missing: %v", err)
	}

	writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md"), "local edit\n")
	conflicted, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, conflicted, "profile_local_modification")
}

func TestManifestCompatibilityPreservesInstallsAndRejectsDuplicateProjectSuites(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	writeJSON(t, manifestPath, map[string]any{
		"version": ManifestVersion,
		"kind":    ProfileManifestKind,
		"profile": map[string]any{"name": "kwanwoo"},
		"installs": []any{map[string]any{
			"pack_id": "kkachi-review", "target_path": "skills/kkachi-review", "pack_checksum": "legacy", "files": []any{},
		}},
	})
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	approved, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !approved.OK {
		t.Fatalf("approved failed: result=%+v err=%v", approved, err)
	}
	manifest := readProjectInstallManifest(t, manifestPath)
	if len(manifest["installs"].([]any)) != 1 || len(manifest["project_suites"].([]any)) != 1 {
		t.Fatalf("manifest did not preserve installs/add suite: %+v", manifest)
	}

	suites := manifest["project_suites"].([]any)
	manifest["project_suites"] = append(suites, suites[0])
	writeJSON(t, manifestPath, manifest)
	duplicate, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, duplicate, "ambiguous_project_suite_manifest")
}

func TestBuildDryRunRejectsDuplicateUnsafeSymlinkUnknownProfileAndHashStateChanges(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n",
		"plan":        "---\nname: plan\n---\n# plan\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, result, "duplicate_installed_skill")
	assertConflict(t, result, "duplicate_target_path")

	unknown, err := BuildDryRun(repo, Options{Profile: "missing", Project: "doksuri-server", SourcePack: VirtualSourcePackID, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, unknown, "unknown_profile")

	repo2 := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	rootA := filepath.Join(t.TempDir(), "profile-a")
	rootB := filepath.Join(t.TempDir(), "profile-b")
	a, err := BuildDryRun(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: rootA, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	b, err := BuildDryRun(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: rootB, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if a.PlanHash == b.PlanHash {
		t.Fatal("plan hash did not change when target profile root changed")
	}
	writeProjectInstallFile(t, filepath.Join(repo2, "skills", "kkachi-plan", "SKILL.md"), "---\nname: kkachi-plan\n---\n# kkachi-plan\nchanged\n")
	writeSkillPackYAML(t, repo2, []string{"kkachi-plan"})
	c, err := BuildDryRun(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: rootA, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if a.PlanHash == c.PlanHash {
		t.Fatal("plan hash did not change when source state changed")
	}

	symlinkRoot := filepath.Join(t.TempDir(), "profile")
	if err := os.MkdirAll(filepath.Join(symlinkRoot, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(symlinkRoot, "skills", "doksuri-server")); err != nil {
		t.Fatal(err)
	}
	symlink, err := BuildDryRun(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: symlinkRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !symlink.OK {
		t.Fatalf("dry-run may plan create before write preflight: %+v", symlink)
	}
	approved, err := ApplyApprovedInstall(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: symlinkRoot, DryRun: true}, symlink.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if approved.OK || firstDiagnosticCode(approved) != "project_install_preflight_failed" {
		t.Fatalf("approved install should reject symlink path: %+v", approved)
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
	names := make([]string, 0, len(skills))
	for name, content := range skills {
		writeProjectInstallFile(t, filepath.Join(repo, "skills", name, "SKILL.md"), content)
		names = append(names, name)
	}
	writeSkillPackYAML(t, repo, names)
	return repo
}

func writeSkillPackYAML(t *testing.T, repo string, skills []string) {
	t.Helper()
	content := "name: fixture\nversion: 0.1.0\nskills:\n"
	for _, skill := range skills {
		content += "  - " + skill + "\n"
	}
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
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

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readProjectInstallManifest(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func firstDiagnosticCode(result Result) string {
	if len(result.Diagnostics) == 0 {
		return ""
	}
	return result.Diagnostics[0].Code
}
