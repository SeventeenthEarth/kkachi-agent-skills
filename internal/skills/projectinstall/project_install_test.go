package projectinstall

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
)

func assertNoHangul(t *testing.T, out string) {
	t.Helper()
	for _, r := range out {
		if r >= 0xAC00 && r <= 0xD7AF {
			t.Fatalf("expected no Korean prose in human output, got %q", out)
		}
	}
}

func TestBuildDryRunRendersProjectPrefixedSuiteAndWritesNothing(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-plan":         "---\nname: kkachi-plan\n---\n# kkachi-plan\n",
		"kkachi-final-verify": "---\nname: kkachi-final-verify\n---\n# kkachi-final-verify\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profiles", "kwanwoo")

	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
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
	if result.SuiteRole != "blue_commander" || result.SuiteMode != SuiteModeFull || result.RoleLabel != "Blue commander / full project suite" {
		t.Fatalf("missing role evidence: %+v", result)
	}
	if len(result.SelectedSkills) != 2 || len(result.ExcludedSkills) != 0 || !result.ApprovalRequest.HashIncludesRoleFields {
		t.Fatalf("role fields were not hash-bound: selected=%+v excluded=%+v approval=%+v", result.SelectedSkills, result.ExcludedSkills, result.ApprovalRequest)
	}
	human := RenderHumanDryRun(result)
	assertNoHangul(t, human)
	for _, want := range []string{"Status:", "Source pack:", "Plan:", "Writes:", "Approval evidence:", "Next:"} {
		if !strings.Contains(human, want) {
			t.Fatalf("human dry-run output missing %q:\n%s", want, human)
		}
	}
	if result.ProjectTailoring.Mode != "prefix_render_only" || result.ProjectTailoring.SemanticPortRequiredBeforeApprovedInstall || result.ProjectTailoring.SemanticAdaptationClaimed {
		t.Fatalf("unexpected tailoring posture: %+v", result.ProjectTailoring)
	}
	if result.ApprovalRequest.EvidenceRef != "dry-run:"+result.PlanHash || !result.ApprovalRequest.HashIncludesProfile || !result.ApprovalRequest.HashIncludesNoWriteEvidence || !result.ApprovalRequest.HashIncludesBackupPlan {
		t.Fatalf("approval request is not hash-bound enough: %+v", result.ApprovalRequest)
	}

	repeated, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if repeated.PlanHash != result.PlanHash {
		t.Fatalf("plan hash was not stable: %s != %s", repeated.PlanHash, result.PlanHash)
	}
}

func TestBuildDryRunProjectsOnlySelectedRoleSkills(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-review":       "---\nname: kkachi-review\n---\n# kkachi-review\n",
		"kkachi-verify":       "---\nname: kkachi-verify\n---\n# kkachi-verify\n",
		"kkachi-implement":    "---\nname: kkachi-implement\n---\n# kkachi-implement\n",
		"kkachi-final-verify": "---\nname: kkachi-final-verify\n---\n# kkachi-final-verify\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profile")

	result, err := BuildDryRun(repo, Options{Profile: "hahuyeon", Project: "doksuri-server", SuiteRole: "red_reviewer", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected ok red reviewer dry-run, got diagnostics=%+v conflicts=%+v", result.Diagnostics, result.Conflicts)
	}
	if result.SuiteMode != SuiteModeRoleSubset || result.RoleLabel != "Red safety/fail-closed reviewer subset" {
		t.Fatalf("unexpected role evidence: %+v", result)
	}
	if len(result.PlannedSkills) != 2 || len(result.SelectedSkills) != 2 || len(result.ExcludedSkills) != 2 {
		t.Fatalf("unexpected role projection counts: planned=%+v selected=%+v excluded=%+v", result.PlannedSkills, result.SelectedSkills, result.ExcludedSkills)
	}
	assertPlannedSkill(t, result, "kkachi-review", "doksuri-server-review", "skills/doksuri-server/doksuri-server-review/SKILL.md")
	assertPlannedSkill(t, result, "kkachi-verify", "doksuri-server-verify", "skills/doksuri-server/doksuri-server-verify/SKILL.md")
	for _, excluded := range result.ExcludedSkills {
		if excluded.Reason != "outside_suite_role" || excluded.InstalledSkill == "" || excluded.SourceSkill == "" {
			t.Fatalf("excluded skill missing reason/rendered identity: %+v", excluded)
		}
		if excluded.SourceSkill == "kkachi-implement" && excluded.InstalledSkill != "doksuri-server-implement" {
			t.Fatalf("forbidden/out-of-role rendered name not recorded: %+v", excluded)
		}
	}
	manifestSuite := result.PlannedManifest["project_suites"].([]map[string]any)[0]
	if manifestSuite["suite_role"] != "red_reviewer" || manifestSuite["suite_mode"] != SuiteModeRoleSubset || manifestSuite["drift_policy"] != "role_subset_expected" {
		t.Fatalf("planned manifest missing role vocabulary: %+v", manifestSuite)
	}
	approvedPreview := buildApprovedResult(result, result.ApprovalRequest.EvidenceRef, result.PlanHash, "preview-install", ".kas/backups/preview", nil, true, nil)
	humanApproved := RenderHumanApproved(approvedPreview)
	if !strings.Contains(humanApproved, "drift_policy: role_subset_expected") {
		t.Fatalf("approved human output should render role-subset drift policy, got:\n%s", humanApproved)
	}
}

func TestBuildDryRunRoleRegistryFailClosedCases(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-review": "---\nname: kkachi-review\n---\n# kkachi-review\n",
		"kkachi-verify": "---\nname: kkachi-verify\n---\n# kkachi-verify\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profile")

	missingRole, err := BuildDryRun(repo, Options{Profile: "hahuyeon", Project: "doksuri-server", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, missingRole, "suite_role_required")

	unknownRole, err := BuildDryRun(repo, Options{Profile: "hahuyeon", Project: "doksuri-server", SuiteRole: "purple_reviewer", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, unknownRole, "unknown_suite_role")

	if err := os.Remove(filepath.Join(repo, filepath.FromSlash(RoleRegistryPath))); err != nil {
		t.Fatal(err)
	}
	missingRegistry, err := BuildDryRun(repo, Options{Profile: "hahuyeon", Project: "doksuri-server", SuiteRole: "red_reviewer", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, missingRegistry, "role_registry_unreadable")
}

func TestApplyApprovedInstallNoWriteOnMismatchAndSuccessfulCreate(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-plan":         "---\nname: kkachi-plan\n---\n# kkachi-plan\n",
		"kkachi-final-verify": "---\nname: kkachi-final-verify\n---\n# kkachi-final-verify\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profiles", "kwanwoo")
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
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

func TestApplyApprovedInstallUsesMaterializedSourceWhenPublicSourceRepoIsEmbedded(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-review": "---\nname: kkachi-review\n---\n# kkachi-review\n",
	})
	if err := os.WriteFile(filepath.Join(repo, ".kas-embedded-source-sha256"), []byte("abc123\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	profileRoot := filepath.Join(t.TempDir(), "profile")
	opts := Options{Profile: "hahuyeon", Project: "doksuri-server", SuiteRole: "orange_pm_reviewer", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK {
		t.Fatalf("dry-run should be ok with materialized source repo: %+v", dryRun)
	}
	if !strings.HasPrefix(dryRun.SourceRepo.Path, "embedded://") {
		t.Fatalf("test must exercise public embedded source identity, got %q", dryRun.SourceRepo.Path)
	}

	approved, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if !approved.OK {
		t.Fatalf("approved install should render from the internal materialized source path, got diagnostics=%+v", approved.Diagnostics)
	}
	written, err := os.ReadFile(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-review", "SKILL.md"))
	if err != nil {
		t.Fatalf("approved install did not write projected skill: %v", err)
	}
	if !strings.Contains(string(written), "doksuri-server-review") {
		t.Fatalf("project-prefixed skill content was not rendered: %s", string(written))
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
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
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
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
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
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
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
	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, result, "duplicate_installed_skill")
	assertConflict(t, result, "duplicate_target_path")

	unknown, err := BuildDryRun(repo, Options{Profile: "missing", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, unknown, "unknown_profile")

	repo2 := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	rootA := filepath.Join(t.TempDir(), "profile-a")
	rootB := filepath.Join(t.TempDir(), "profile-b")
	a, err := BuildDryRun(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: rootA, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	b, err := BuildDryRun(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: rootB, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if a.PlanHash == b.PlanHash {
		t.Fatal("plan hash did not change when target profile root changed")
	}
	writeProjectInstallFile(t, filepath.Join(repo2, "skills", "kkachi-plan", "SKILL.md"), "---\nname: kkachi-plan\n---\n# kkachi-plan\nchanged\n")
	writeSkillPackYAML(t, repo2, []string{"kkachi-plan"})
	c, err := BuildDryRun(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: rootA, DryRun: true})
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
	symlink, err := BuildDryRun(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: symlinkRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !symlink.OK {
		t.Fatalf("dry-run may plan create before write preflight: %+v", symlink)
	}
	approved, err := ApplyApprovedInstall(repo2, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: symlinkRoot, DryRun: true}, symlink.ApprovalRequest.EvidenceRef)
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

	invalidProject, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "../escape", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, invalidProject, "invalid_project_id")

	unknownPack, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: "missing-suite", ProfileRoot: profileRoot, DryRun: true})
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

	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
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

	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, result, "umbrella_only")
}

func TestBuildDryRunRejectsUmbrellaOnlySourceSuite(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-kas": "---\nname: kkachi-kas\n---\n# kkachi-kas\n"})
	result, err := BuildDryRun(repo, Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: filepath.Join(t.TempDir(), "profile"), DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	assertConflict(t, result, "umbrella_only")
}

func TestPlannerSourceContainsNoWriteAPIs(t *testing.T) {
	assertSourceContainsNoWriteAPIs(t, "project_install.go", "")
	assertSourceContainsNoWriteAPIs(t, "kasproj004.go", sourceSection(t, "kasproj004.go", "func buildProjectActionDryRun", "func applyApprovedProjectAction"))
	assertSourceContainsNoWriteAPIs(t, "uninstall.go", sourceSection(t, "uninstall.go", "func BuildProjectUninstallDryRun", "func RenderHumanProjectUninstall"))
}

func assertSourceContainsNoWriteAPIs(t *testing.T, path string, source string) {
	t.Helper()
	if source == "" {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source = string(data)
	}
	for _, forbidden := range []string{"os.WriteFile", "os.MkdirAll", "os.Remove", "os.Rename", "writeFileAtomic", "writeJSONFile", "copyProfileFile"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("projectinstall planner source %s must stay read-only; found %s", path, forbidden)
		}
	}
}

func sourceSection(t *testing.T, path string, startMarker string, endMarker string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	start := strings.Index(source, startMarker)
	end := strings.Index(source, endMarker)
	if start < 0 || end < 0 || end <= start {
		t.Fatalf("could not find guarded source section %s..%s in %s", startMarker, endMarker, path)
	}
	return source[start:end]
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
	writeProjectSuiteRoleRegistry(t, repo)
	return repo
}

func writeProjectSuiteRoleRegistry(t *testing.T, repo string) {
	t.Helper()
	content := `version: role-aware-project-suite/v1
roles:
  blue_commander:
    display_label: "Blue commander / full project suite"
    selection_mode: full_source_suite
    required_source_skills: "*"
    optional_source_skills: []
    forbidden_source_skills: []
  red_reviewer:
    display_label: "Red safety/fail-closed reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
      - kkachi-verify
    optional_source_skills: []
    forbidden_source_skills:
      - kkachi-implement
  orange_pm_reviewer:
    display_label: "Orange operator-value reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
    optional_source_skills:
      - kkachi-final-verify
    optional_selection_default: unselected
    forbidden_source_skills:
      - kkachi-implement
  gray_scribe:
    display_label: "Gray evidence/scribe reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
      - kkachi-final-verify
    optional_source_skills: []
    forbidden_source_skills:
      - kkachi-implement
      - kkachi-docs-update
`
	if err := os.MkdirAll(filepath.Join(repo, "registries"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, filepath.FromSlash(RoleRegistryPath)), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
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

func TestProjectSuiteDoctorDetectsHealthyMissingUmbrellaChecksumAndUnknownDirs(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-plan":         "---\nname: kkachi-plan\n---\n# kkachi-plan\n",
		"kkachi-final-verify": "---\nname: kkachi-final-verify\n---\n# kkachi-final-verify\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	approved, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !approved.OK {
		t.Fatalf("approved install failed: result=%+v err=%v", approved, err)
	}

	healthy, err := BuildProjectSuiteDoctor(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !healthy.OK || healthy.Mode != modeProjectSuiteDoctor || healthy.SourcePack.ID != VirtualSourcePackID || healthy.ProjectSuite.InstalledSkillCount != 2 {
		t.Fatalf("expected healthy project-suite doctor, got %+v", healthy)
	}

	writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-extra", "SKILL.md"), "extra")
	warning, err := BuildProjectSuiteDoctor(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !warning.OK || !hasProjectSuiteDiag(warning.ProjectSuiteDiagnostics, "unknown_profile_skill_dir", "warning") {
		t.Fatalf("expected unknown project-prefixed dir warning, got %+v", warning.ProjectSuiteDiagnostics)
	}

	if err := os.WriteFile(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md"), []byte("local edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mismatch, err := BuildProjectSuiteDoctor(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if mismatch.OK || !hasProjectSuiteDiag(mismatch.ProjectSuiteDiagnostics, "checksum_mismatch", "error") {
		t.Fatalf("expected checksum mismatch error, got %+v", mismatch.ProjectSuiteDiagnostics)
	}

	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	manifest := readProjectInstallManifest(t, manifestPath)
	for _, rawSuite := range manifest["project_suites"].([]any) {
		suite := rawSuite.(map[string]any)
		for _, rawSkill := range suite["installed_skills"].([]any) {
			skill := rawSkill.(map[string]any)
			if skill["installed_skill"] == "doksuri-server-plan" {
				skill["tailoring_mode"] = "profile_local_repo_semantic_tailoring"
				skill["drift_policy"] = "manual_review_required"
			}
		}
	}
	writeJSON(t, manifestPath, manifest)
	tailored, err := BuildProjectSuiteDoctor(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !tailored.OK || !hasProjectSuiteDiag(tailored.ProjectSuiteDiagnostics, "project_tailoring_checksum_drift", "warning") || hasProjectSuiteDiag(tailored.ProjectSuiteDiagnostics, "checksum_mismatch", "error") {
		t.Fatalf("expected project-local tailoring drift warning, got %+v", tailored.ProjectSuiteDiagnostics)
	}

	missingRoot := filepath.Join(t.TempDir(), "profile")
	missing, err := BuildProjectSuiteDoctor(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: missingRoot})
	if err != nil {
		t.Fatal(err)
	}
	if missing.OK || !hasProjectSuiteDiag(missing.ProjectSuiteDiagnostics, "missing_project_suite", "error") {
		t.Fatalf("expected missing suite error, got %+v", missing.ProjectSuiteDiagnostics)
	}

	umbrellaRoot := filepath.Join(t.TempDir(), "profile")
	writeProjectInstallFile(t, filepath.Join(umbrellaRoot, "skills", "doksuri-server", "doksuri-server-kas", "SKILL.md"), "umbrella")
	umbrella, err := BuildProjectSuiteDoctor(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: umbrellaRoot})
	if err != nil {
		t.Fatal(err)
	}
	if umbrella.OK || !hasProjectSuiteDiag(umbrella.ProjectSuiteDiagnostics, "umbrella_only", "error") {
		t.Fatalf("expected umbrella-only error, got %+v", umbrella.ProjectSuiteDiagnostics)
	}
}

func TestProjectSuiteDoctorRoleAwareSemantics(t *testing.T) {
	skills := map[string]string{
		"kkachi-review":       "---\nname: kkachi-review\n---\n# kkachi-review\n",
		"kkachi-verify":       "---\nname: kkachi-verify\n---\n# kkachi-verify\n",
		"kkachi-implement":    "---\nname: kkachi-implement\n---\n# kkachi-implement\n",
		"kkachi-final-verify": "---\nname: kkachi-final-verify\n---\n# kkachi-final-verify\n",
		"kkachi-docs-update":  "---\nname: kkachi-docs-update\n---\n# kkachi-docs-update\n",
	}
	repo := makeProjectInstallRepo(t, skills)
	project := "doksuri-server"

	t.Run("blue full suite healthy", func(t *testing.T) {
		doctor := installProjectSuiteAndDoctor(t, repo, project, "blue_commander")
		if !doctor.OK || doctor.ProjectSuite.InstalledSkillCount != len(skills) {
			t.Fatalf("expected blue full suite healthy, got ok=%t state=%+v diagnostics=%+v", doctor.OK, doctor.ProjectSuite, doctor.ProjectSuiteDiagnostics)
		}
	})

	for _, tc := range []struct {
		role  string
		count int
	}{
		{role: "red_reviewer", count: 2},
		{role: "orange_pm_reviewer", count: 1},
		{role: "gray_scribe", count: 2},
	} {
		t.Run(tc.role+" subset healthy", func(t *testing.T) {
			doctor := installProjectSuiteAndDoctor(t, repo, project, tc.role)
			if !doctor.OK || doctor.ProjectSuite.InstalledSkillCount != tc.count {
				t.Fatalf("expected role subset healthy with only selected skills installed, got ok=%t state=%+v diagnostics=%+v", doctor.OK, doctor.ProjectSuite, doctor.ProjectSuiteDiagnostics)
			}
			if hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "missing_file", "error") || hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "missing_selected_role_skill", "error") {
				t.Fatalf("unselected full-suite skills must not be missing errors: %+v", doctor.ProjectSuiteDiagnostics)
			}
		})
	}

	t.Run("missing selected role skill is an error", func(t *testing.T) {
		profileRoot := installProjectSuite(t, repo, project, "red_reviewer")
		removeManifestInstalledSkill(t, profileRoot, project, "doksuri-server-verify")
		doctor := buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
		if doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "missing_selected_role_skill", "error") {
			t.Fatalf("expected missing selected role skill error, got ok=%t diagnostics=%+v", doctor.OK, doctor.ProjectSuiteDiagnostics)
		}
	})

	t.Run("selected checksum mismatch is an error", func(t *testing.T) {
		profileRoot := installProjectSuite(t, repo, project, "orange_pm_reviewer")
		if err := os.WriteFile(filepath.Join(profileRoot, "skills", project, "doksuri-server-review", "SKILL.md"), []byte("local drift\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		doctor := buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
		if doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "checksum_mismatch", "error") {
			t.Fatalf("expected selected checksum mismatch error, got ok=%t diagnostics=%+v", doctor.OK, doctor.ProjectSuiteDiagnostics)
		}
	})

	t.Run("out of role KAS-managed physical extra is an error", func(t *testing.T) {
		profileRoot := installProjectSuite(t, repo, project, "gray_scribe")
		writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", project, "doksuri-server-implement", "SKILL.md"), skills["kkachi-implement"])
		doctor := buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
		if doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "out_of_role_kas_managed_skill", "error") {
			t.Fatalf("expected out-of-role KAS-managed skill error, got ok=%t diagnostics=%+v", doctor.OK, doctor.ProjectSuiteDiagnostics)
		}
	})

	t.Run("benign unknown personal skill is warning only", func(t *testing.T) {
		profileRoot := installProjectSuite(t, repo, project, "red_reviewer")
		writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", project, "doksuri-server-local-note", "SKILL.md"), "personal\n")
		doctor := buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
		if !doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "unknown_profile_skill_dir", "warning") {
			t.Fatalf("expected benign unknown personal skill warning, got ok=%t diagnostics=%+v", doctor.OK, doctor.ProjectSuiteDiagnostics)
		}
	})

	t.Run("shadowing unknown personal skill is an error", func(t *testing.T) {
		profileRoot := installProjectSuite(t, repo, project, "red_reviewer")
		writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", project, "doksuri-server-review-local", "SKILL.md"), "shadow\n")
		doctor := buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
		if doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "ambiguous_profile_skill_dir", "error") {
			t.Fatalf("expected shadowing unknown personal skill error, got ok=%t diagnostics=%+v", doctor.OK, doctor.ProjectSuiteDiagnostics)
		}
	})

	t.Run("missing suite_role fails closed", func(t *testing.T) {
		profileRoot := installProjectSuite(t, repo, project, "blue_commander")
		setManifestSuiteRole(t, profileRoot, project, nil)
		doctor := buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
		if doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "missing_suite_role", "error") {
			t.Fatalf("expected missing suite_role fail-closed error, got ok=%t diagnostics=%+v", doctor.OK, doctor.ProjectSuiteDiagnostics)
		}
	})

	t.Run("unknown suite_role fails closed", func(t *testing.T) {
		profileRoot := installProjectSuite(t, repo, project, "blue_commander")
		role := "purple_reviewer"
		setManifestSuiteRole(t, profileRoot, project, &role)
		doctor := buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
		if doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "unknown_suite_role", "error") {
			t.Fatalf("expected unknown suite_role fail-closed error, got ok=%t diagnostics=%+v", doctor.OK, doctor.ProjectSuiteDiagnostics)
		}
	})

	for _, role := range []string{"red_reviewer", "orange_pm_reviewer", "gray_scribe"} {
		t.Run("legacy full suite unhealthy for "+role, func(t *testing.T) {
			profileRoot := installProjectSuite(t, repo, project, "blue_commander")
			setManifestSuiteRole(t, profileRoot, project, &role)
			doctor := buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
			if doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "out_of_role_kas_managed_skill", "error") {
				t.Fatalf("expected legacy full suite to be unhealthy for %s, got ok=%t diagnostics=%+v", role, doctor.OK, doctor.ProjectSuiteDiagnostics)
			}
		})
	}
}

func installProjectSuiteAndDoctor(t *testing.T, repo string, project string, suiteRole string) ProjectSuiteDoctorResult {
	t.Helper()
	profileRoot := installProjectSuite(t, repo, project, suiteRole)
	return buildProjectSuiteDoctorForTest(t, repo, profileRoot, project)
}

func installProjectSuite(t *testing.T, repo string, project string, suiteRole string) string {
	t.Helper()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	opts := Options{Profile: "kwanwoo", Project: project, SuiteRole: suiteRole, SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK {
		t.Fatalf("dry-run failed: diagnostics=%+v conflicts=%+v", dryRun.Diagnostics, dryRun.Conflicts)
	}
	approved, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !approved.OK {
		t.Fatalf("approved install failed: result=%+v err=%v", approved, err)
	}
	return profileRoot
}

func buildProjectSuiteDoctorForTest(t *testing.T, repo string, profileRoot string, project string) ProjectSuiteDoctorResult {
	t.Helper()
	doctor, err := BuildProjectSuiteDoctor(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	return doctor
}

func setManifestSuiteRole(t *testing.T, profileRoot string, project string, suiteRole *string) {
	t.Helper()
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	manifest := readProjectInstallManifest(t, manifestPath)
	for _, rawSuite := range manifest["project_suites"].([]any) {
		suite := rawSuite.(map[string]any)
		if suite["project"] != project {
			continue
		}
		if suiteRole == nil {
			delete(suite, "suite_role")
			delete(suite, "suite_mode")
			delete(suite, "selected_skills")
			delete(suite, "excluded_skills")
			delete(suite, "role_registry")
		} else {
			suite["suite_role"] = *suiteRole
			suite["suite_mode"] = SuiteModeRoleSubset
		}
	}
	writeJSON(t, manifestPath, manifest)
}

func removeManifestInstalledSkill(t *testing.T, profileRoot string, project string, installedSkill string) {
	t.Helper()
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	manifest := readProjectInstallManifest(t, manifestPath)
	for _, rawSuite := range manifest["project_suites"].([]any) {
		suite := rawSuite.(map[string]any)
		if suite["project"] != project {
			continue
		}
		rawSkills := suite["installed_skills"].([]any)
		filtered := make([]any, 0, len(rawSkills))
		for _, rawSkill := range rawSkills {
			skill := rawSkill.(map[string]any)
			if skill["installed_skill"] == installedSkill {
				continue
			}
			filtered = append(filtered, skill)
		}
		suite["installed_skills"] = filtered
	}
	writeJSON(t, manifestPath, manifest)
}

func TestApplyProjectUninstallFailsClosedOnDriftSymlinkAndBacksUp(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	installed, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !installed.OK {
		t.Fatalf("approved install failed: result=%+v err=%v", installed, err)
	}
	target := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")
	originalTarget, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	localOnly := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-local", "SKILL.md")
	writeProjectInstallFile(t, localOnly, "local only\n")
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	manifest := readProjectInstallManifest(t, manifestPath)
	manifest["installs"] = append(manifest["installs"].([]any), map[string]any{
		"pack_id":     "generic-kas",
		"target_path": "skills/generic-kas",
		"files":       []any{},
	})
	manifest["project_suites"] = append(manifest["project_suites"].([]any), map[string]any{
		"kind":        ManifestKind,
		"project":     "other-project",
		"source_pack": map[string]any{"id": VirtualSourcePackID},
		"target_path": "skills/other-project",
		"files":       []any{},
	})
	if err := writeJSONFile(manifestPath, manifest); err != nil {
		t.Fatal(err)
	}
	uninstallOpts := ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot}
	uninstallDryRun := BuildProjectUninstallDryRun(uninstallOpts)
	if !uninstallDryRun.OK || uninstallDryRun.ApprovalRequest.EvidenceRef != "dry-run:"+uninstallDryRun.PlanHash {
		t.Fatalf("unexpected uninstall dry-run: %+v", uninstallDryRun)
	}
	if !strings.Contains(uninstallDryRun.FutureApplyCommand, "--apply dry-run:sha256:") ||
		!strings.Contains(uninstallDryRun.FutureApplyCommand, "--backup-vault-root <abs-path>") {
		t.Fatalf("future apply command must include approval evidence and backup vault placeholder, got %q", uninstallDryRun.FutureApplyCommand)
	}
	assertNoHangul(t, uninstallDryRun.FutureApplyCommand)

	if err := os.WriteFile(target, []byte("local drift\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	drift := BuildProjectUninstallDryRun(uninstallOpts)
	if drift.OK || firstUninstallDiagnosticCode(drift) != "checksum_mismatch" {
		t.Fatalf("checksum drift must block uninstall dry-run: %+v", drift)
	}
	wrong, err := ApplyProjectUninstall(uninstallOpts, uninstallDryRun.ApprovalRequest.EvidenceRef, filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if wrong.OK || firstUninstallDiagnosticCode(wrong) != "approval_plan_hash_mismatch" {
		t.Fatalf("stale uninstall hash should fail closed after drift: %+v", wrong)
	}

	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "escape"), target); err != nil {
		t.Fatal(err)
	}
	symlink := BuildProjectUninstallDryRun(uninstallOpts)
	if symlink.OK || firstUninstallDiagnosticCode(symlink) != "target_symlink_rejected" {
		t.Fatalf("symlink target must block uninstall dry-run: %+v", symlink)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	writeProjectInstallFile(t, target, string(originalTarget))
	fresh := BuildProjectUninstallDryRun(uninstallOpts)
	unsafeVault, err := ApplyProjectUninstall(uninstallOpts, fresh.ApprovalRequest.EvidenceRef, profileRoot)
	if err != nil {
		t.Fatal(err)
	}
	if unsafeVault.OK || firstUninstallDiagnosticCode(unsafeVault) != "backup_vault_root_rejected" {
		t.Fatalf("profile-local backup vault root should fail closed: %+v", unsafeVault)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("unsafe backup vault root removed target: %v", err)
	}
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	applied, err := ApplyProjectUninstall(uninstallOpts, fresh.ApprovalRequest.EvidenceRef, vault)
	if err != nil || !applied.OK || !applied.BackupRecovery.BackupVerified {
		t.Fatalf("approved uninstall failed: result=%+v err=%v", applied, err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("approved uninstall did not remove manifest target: %v", err)
	}
	if _, err := os.Stat(localOnly); err != nil {
		t.Fatalf("approved uninstall touched local-only file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(applied.BackupPath, "files", "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); err != nil {
		t.Fatalf("approved uninstall did not write backup file: %v", err)
	}
	updatedManifest := readProjectInstallManifest(t, manifestPath)
	if len(updatedManifest["installs"].([]any)) != 1 {
		t.Fatalf("approved uninstall did not preserve unrelated generic manifest entries: %+v", updatedManifest)
	}
	projectSuites := updatedManifest["project_suites"].([]any)
	if len(projectSuites) != 1 || projectSuites[0].(map[string]any)["project"] != "other-project" {
		t.Fatalf("approved uninstall did not preserve unrelated project suite entries: %+v", updatedManifest)
	}
}

func firstUninstallDiagnosticCode(result ProjectUninstallResult) string {
	if len(result.Diagnostics) == 0 {
		return ""
	}
	return result.Diagnostics[0].Code
}

func TestProjectRepairHumanOutputShowsSuiteDiagnosticsAndActions(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	repairDryRun, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !repairDryRun.OK || !repairDryRun.ApprovalRequest.Required {
		t.Fatalf("missing suite repair should remain approvable, got %+v", repairDryRun)
	}
	human := RenderHumanProjectAction(repairDryRun)
	assertNoHangul(t, human)
	for _, want := range []string{
		"project-suite diagnostic: error/missing_project_suite",
		"no trusted project_suites[] entry exists",
		"next_action: Install or repair the project-specific suite",
		"path skills/doksuri-server",
		"Action: create skills/doksuri-server/doksuri-server-plan/SKILL.md",
		"restore_missing_project_suite_file",
		"Approval required: true",
	} {
		if !strings.Contains(human, want) {
			t.Fatalf("human repair output missing %q:\n%s", want, human)
		}
	}
}

func TestProjectSuiteDoctorHumanOutputShowsDiagnosticDetailsAndBoundedNext(t *testing.T) {
	result := ProjectSuiteDoctorResult{
		OK:            false,
		Command:       "doctor",
		Mode:          modeProjectSuiteDoctor,
		TargetProfile: discovery.TargetProfile{Name: "kwanwoo"},
		Project:       Project{ID: "doksuri-server", TargetSuitePath: "skills/doksuri-server"},
		SourcePack:    SourcePack{ID: VirtualSourcePackID},
		ManifestPath:  ".kas/skill-pack-manifest.json",
		ProjectSuite:  ProjectSuiteState{ManifestState: "present", PhysicalState: "drifted", FilesChecked: 1},
		ProjectSuiteDiagnostics: []ProjectSuiteDiagnostic{
			suiteDiag("doksuri-server", "doksuri-server-review", "skills/doksuri-server/doksuri-server-review/SKILL.md", "error", "checksum_mismatch", "installed file checksum differs from manifest/source evidence", "Review diagnostics and use only an approved KASROLE-004 repair/prune plan when applicable."),
		},
	}
	finalizeProjectSuiteDoctor(&result)

	human := RenderHumanProjectSuiteDoctor(result)
	assertNoHangul(t, human)
	for _, want := range []string{
		"project-suite diagnostic: error/checksum_mismatch - installed file checksum differs from manifest/source evidence",
		"skill doksuri-server-review",
		"path skills/doksuri-server/doksuri-server-review/SKILL.md",
		"next_action: Review diagnostics and use only an approved KASROLE-004 repair/prune plan when applicable.",
		"Next: Review project-suite diagnostics; use approved dry-run planning only, and reserve repair/prune/profile cleanup for approved KASROLE-004 workflows when applicable.",
	} {
		if !strings.Contains(human, want) {
			t.Fatalf("human project-suite doctor output missing %q:\n%s", want, human)
		}
	}
	if strings.Contains(human, "Run repair-project-kas --dry-run for deterministic repairs, or migrate-project-kas --from-generic --dry-run") {
		t.Fatalf("human project-suite doctor output contains misleading generic repair/migration next action:\n%s", human)
	}
}

func TestProjectSuiteDoctorRealPathBoundsInspectNextAction(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n",
	})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	opts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	approved, err := ApplyApprovedInstall(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !approved.OK {
		t.Fatalf("approved install failed: result=%+v err=%v", approved, err)
	}

	target := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")
	if err := os.WriteFile(target, []byte("local edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doctor, err := BuildProjectSuiteDoctor(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if doctor.OK || !hasProjectSuiteDiag(doctor.ProjectSuiteDiagnostics, "checksum_mismatch", "error") {
		t.Fatalf("expected checksum mismatch error, got ok=%t diagnostics=%+v", doctor.OK, doctor.ProjectSuiteDiagnostics)
	}

	human := RenderHumanProjectSuiteDoctor(doctor)
	assertNoHangul(t, human)
	for _, want := range []string{
		"project-suite diagnostic: error/checksum_mismatch - installed file checksum differs from manifest/source evidence",
		"skill doksuri-server-plan",
		"path skills/doksuri-server/doksuri-server-plan/SKILL.md",
		"next_action: Review diagnostics and use only an approved KASROLE-004 repair/prune plan when applicable.",
	} {
		if !strings.Contains(human, want) {
			t.Fatalf("real-path human project-suite doctor output missing %q:\n%s", want, human)
		}
	}
	for _, forbidden := range []string{
		"Run repair-project-kas --dry-run",
		"migrate-project-kas --from-generic",
	} {
		if strings.Contains(human, forbidden) {
			t.Fatalf("real-path human project-suite doctor output contains unbounded next action %q:\n%s", forbidden, human)
		}
	}
}

func TestProjectRepairBlocksErrorUnknownProfileSkillDirButAllowsWarning(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	installOpts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, installOpts)
	if err != nil {
		t.Fatal(err)
	}
	approved, err := ApplyApprovedInstall(repo, installOpts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !approved.OK {
		t.Fatalf("approved install failed: result=%+v err=%v", approved, err)
	}

	writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-extra", "SKILL.md"), "extra")
	warning, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !warning.OK || warning.ApprovalRequest.Required || !hasProjectSuiteDiag(warning.ProjectSuiteDiagnostics, "unknown_profile_skill_dir", "warning") {
		t.Fatalf("project-prefixed unknown dir warning should remain non-blocking, got %+v", warning)
	}

	writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", "doksuri-server", "kkachi-plan", "SKILL.md"), "rogue")
	blocked, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if blocked.OK || blocked.ApprovalRequest.Required || !hasProjectSuiteDiag(blocked.ProjectSuiteDiagnostics, "unknown_profile_skill_dir", "error") || !hasProjectActionDiagnostic(blocked, "unknown_profile_skill_dir") {
		t.Fatalf("rogue unknown dir must block repair approval, got %+v", blocked)
	}
}

func TestProjectRepairDefaultsSourcePackDryRunApprovalAndUnknownExplicit(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	installOpts := Options{Profile: "kwanwoo", Project: "doksuri-server", SuiteRole: "blue_commander", SourcePack: VirtualSourcePackID, ProfileRoot: profileRoot, DryRun: true}
	dryRun, err := BuildDryRun(repo, installOpts)
	if err != nil {
		t.Fatal(err)
	}
	approved, err := ApplyApprovedInstall(repo, installOpts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !approved.OK {
		t.Fatalf("approved install failed: result=%+v err=%v", approved, err)
	}
	target := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}

	repairDryRun, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !repairDryRun.OK || repairDryRun.SourcePack.ID != VirtualSourcePackID || repairDryRun.SourcePack.Source != "default" || !repairDryRun.NoWrite.Guaranteed || len(repairDryRun.PlannedActions) == 0 {
		t.Fatalf("unexpected repair dry-run: %+v", repairDryRun)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("repair dry-run wrote missing skill: %v", err)
	}
	wrong, err := ApplyApprovedRepair(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot}, "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if wrong.OK || firstProjectActionDiagnosticCode(wrong) != "approval_plan_hash_mismatch" {
		t.Fatalf("wrong hash should fail closed: %+v", wrong)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("wrong repair approval wrote skill: %v", err)
	}
	vault := filepath.Join(t.TempDir(), "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	repaired, err := ApplyApprovedRepair(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot, BackupVaultRoot: vault}, repairDryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !repaired.OK || repaired.RepairID == "" || repaired.Recovery == nil {
		t.Fatalf("approved repair failed: result=%+v err=%v", repaired, err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("approved repair did not restore skill: %v", err)
	}

	unknown, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", SourcePack: "missing-suite", SourcePackExplicit: true, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if unknown.OK || firstProjectActionDiagnosticCode(unknown) != "unknown_source_pack" {
		t.Fatalf("unknown explicit source-pack must fail closed: %+v", unknown)
	}
}

func TestProjectRepairRoleAwarePrune(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-review":       "---\nname: kkachi-review\n---\n# kkachi-review\n",
		"kkachi-verify":       "---\nname: kkachi-verify\n---\n# kkachi-verify\n",
		"kkachi-implement":    "---\nname: kkachi-implement\n---\n# kkachi-implement\n",
		"kkachi-final-verify": "---\nname: kkachi-final-verify\n---\n# kkachi-final-verify\n",
	})
	project := "doksuri-server"
	profileRoot := installProjectSuite(t, repo, project, "blue_commander")
	role := "red_reviewer"
	setManifestSuiteRole(t, profileRoot, project, &role)
	personal := filepath.Join(profileRoot, "skills", project, "doksuri-server-local-note", "SKILL.md")
	writeProjectInstallFile(t, personal, "personal\n")

	missingRole, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, ProfileRoot: profileRoot, PruneExtra: true})
	if err != nil {
		t.Fatal(err)
	}
	if missingRole.OK || firstProjectActionDiagnosticCode(missingRole) != "suite_role_required" {
		t.Fatalf("missing suite_role must fail closed: %+v", missingRole)
	}

	unknownRole, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: "purple_reviewer", ProfileRoot: profileRoot, PruneExtra: true})
	if err != nil {
		t.Fatal(err)
	}
	if unknownRole.OK || firstProjectActionDiagnosticCode(unknownRole) != "unknown_suite_role" {
		t.Fatalf("unknown suite_role must fail closed: %+v", unknownRole)
	}

	mismatch, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: "gray_scribe", ProfileRoot: profileRoot, PruneExtra: true})
	if err != nil {
		t.Fatal(err)
	}
	if mismatch.OK || !hasProjectActionDiagnostic(mismatch, "suite_role_mismatch") {
		t.Fatalf("manifest/request role mismatch must fail closed: %+v", mismatch)
	}

	noPrune, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: role, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if noPrune.OK || !hasProjectActionDiagnostic(noPrune, "out_of_role_kas_managed_skill") {
		t.Fatalf("out-of-role extras without --prune-extra must remain blocked: %+v", noPrune)
	}
	for _, action := range noPrune.PlannedActions {
		if action.Action == "remove" {
			t.Fatalf("repair without --prune-extra planned removal: %+v", noPrune.PlannedActions)
		}
	}

	prune, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: role, ProfileRoot: profileRoot, PruneExtra: true})
	if err != nil {
		t.Fatal(err)
	}
	if !prune.OK || !prune.ApprovalRequest.Required {
		t.Fatalf("role-aware prune should be approvable: %+v", prune)
	}
	if prune.SuiteRole != role || prune.SuiteMode != SuiteModeRoleSubset || prune.RoleLabel == "" || !prune.ApprovalRequest.HashIncludesRoleFields {
		t.Fatalf("missing role-bound evidence: %+v", prune)
	}
	if prune.Summary.CountsByAction["remove"] != 2 || prune.Summary.CountsByAction["skip"] != 2 {
		t.Fatalf("unexpected compact counts: %+v", prune.Summary)
	}
	for _, removed := range []string{"doksuri-server-final-verify", "doksuri-server-implement"} {
		assertProjectAction(t, prune, "remove", filepath.ToSlash(filepath.Join("skills", project, removed, "SKILL.md")))
	}
	if !containsString(prune.NoSpillover.UnknownPersonalSkillsPreserved, filepath.ToSlash(filepath.Join("skills", project, "doksuri-server-local-note", "SKILL.md"))) {
		t.Fatalf("unknown personal skill was not preserved in no-spillover evidence: %+v", prune.NoSpillover)
	}
	human := RenderHumanProjectAction(prune)
	assertNoHangul(t, human)
	countsLine := "Counts: keep 2, create 0, update 0, remove 2"
	if !strings.Contains(human, countsLine) || strings.Index(human, countsLine) > strings.Index(human, "Action:") {
		t.Fatalf("human output must show compact counts before detailed actions:\n%s", human)
	}
	if !strings.Contains(human, "Backup: apply requires explicit absolute --backup-vault-root") ||
		!strings.Contains(human, "Recovery: manifest write last") ||
		!strings.Contains(human, "No-spillover: preserved unknown personal skills 1") {
		t.Fatalf("human output missing backup/recovery/no-spillover evidence:\n%s", human)
	}

	repeated, err := BuildProjectRepairDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: role, ProfileRoot: profileRoot, PruneExtra: true})
	if err != nil {
		t.Fatal(err)
	}
	if repeated.PlanHash != prune.PlanHash {
		t.Fatalf("plan hash must be deterministic: %s != %s", repeated.PlanHash, prune.PlanHash)
	}
	if _, err := os.Stat(personal); err != nil {
		t.Fatalf("dry-run touched personal skill: %v", err)
	}
}

func TestProjectRepairRoleAwareApply(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{
		"kkachi-review":    "---\nname: kkachi-review\n---\n# kkachi-review\n",
		"kkachi-verify":    "---\nname: kkachi-verify\n---\n# kkachi-verify\n",
		"kkachi-implement": "---\nname: kkachi-implement\n---\n# kkachi-implement\n",
	})
	project := "doksuri-server"
	profileRoot := installProjectSuite(t, repo, project, "blue_commander")
	role := "red_reviewer"
	setManifestSuiteRole(t, profileRoot, project, &role)
	personal := filepath.Join(profileRoot, "skills", project, "doksuri-server-local-note", "SKILL.md")
	writeProjectInstallFile(t, personal, "personal\n")
	opts := ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: role, ProfileRoot: profileRoot, PruneExtra: true}
	dryRun, err := BuildProjectRepairDryRun(repo, opts)
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK || dryRun.Summary.CountsByAction["remove"] != 1 {
		t.Fatalf("unexpected dry-run: %+v", dryRun)
	}

	malformed, err := ApplyApprovedRepair(repo, opts, "not-a-hash")
	if err != nil {
		t.Fatal(err)
	}
	if malformed.OK || firstProjectActionDiagnosticCode(malformed) != "approval_evidence_malformed" {
		t.Fatalf("malformed approval must fail closed: %+v", malformed)
	}
	missingVault, err := ApplyApprovedRepair(repo, opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if missingVault.OK || firstProjectActionDiagnosticCode(missingVault) != "backup_vault_root_rejected" {
		t.Fatalf("apply without explicit backup vault root must fail closed: %+v", missingVault)
	}
	relativeVault := opts
	relativeVault.BackupVaultRoot = "relative"
	unsafeVault, err := ApplyApprovedRepair(repo, relativeVault, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if unsafeVault.OK || firstProjectActionDiagnosticCode(unsafeVault) != "backup_vault_root_rejected" {
		t.Fatalf("relative backup vault root must fail closed: %+v", unsafeVault)
	}
	wrong, err := ApplyApprovedRepair(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: role, ProfileRoot: profileRoot, PruneExtra: true, BackupVaultRoot: t.TempDir()}, "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if wrong.OK || firstProjectActionDiagnosticCode(wrong) != "approval_plan_hash_mismatch" {
		t.Fatalf("wrong hash must fail closed: %+v", wrong)
	}
	removeTarget := filepath.Join(profileRoot, "skills", project, "doksuri-server-implement", "SKILL.md")
	if _, err := os.Stat(removeTarget); err != nil {
		t.Fatalf("failed approvals removed target: %v", err)
	}

	driftRoot := installProjectSuite(t, repo, project, "blue_commander")
	setManifestSuiteRole(t, driftRoot, project, &role)
	driftOpts := ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: role, ProfileRoot: driftRoot, PruneExtra: true}
	driftDryRun, err := BuildProjectRepairDryRun(repo, driftOpts)
	if err != nil || !driftDryRun.OK {
		t.Fatalf("drift dry-run failed: result=%+v err=%v", driftDryRun, err)
	}
	if err := os.WriteFile(filepath.Join(driftRoot, "skills", project, "doksuri-server-implement", "SKILL.md"), []byte("local drift\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	driftVault := t.TempDir()
	driftApply, err := ApplyApprovedRepair(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: role, ProfileRoot: driftRoot, PruneExtra: true, BackupVaultRoot: driftVault}, driftDryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if driftApply.OK || firstProjectActionDiagnosticCode(driftApply) != "approval_plan_hash_mismatch" {
		t.Fatalf("checksum drift must change current plan hash before writes: %+v", driftApply)
	}

	badVault := t.TempDir()
	if err := os.WriteFile(filepath.Join(badVault, "kas-backups"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestBackupFail, err := ApplyApprovedRepair(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: project, SuiteRole: role, ProfileRoot: profileRoot, PruneExtra: true, BackupVaultRoot: badVault}, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if manifestBackupFail.OK || firstProjectActionDiagnosticCode(manifestBackupFail) != "previous_manifest_backup_failed" {
		t.Fatalf("manifest backup failure must fail before profile mutation: %+v", manifestBackupFail)
	}
	if _, err := os.Stat(removeTarget); err != nil {
		t.Fatalf("manifest backup failure mutated target before failing: %v", err)
	}

	vault := t.TempDir()
	applyOpts := opts
	applyOpts.BackupVaultRoot = vault
	applied, err := ApplyApprovedRepair(repo, applyOpts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil || !applied.OK || applied.RepairID == "" || applied.Recovery == nil {
		t.Fatalf("approved repair prune failed: result=%+v err=%v", applied, err)
	}
	if _, err := os.Stat(removeTarget); !os.IsNotExist(err) {
		t.Fatalf("approved repair did not remove out-of-role target or unexpected stat error: %v", err)
	}
	if _, err := os.Stat(personal); err != nil {
		t.Fatalf("approved repair touched personal skill: %v", err)
	}
	if _, err := os.Stat(filepath.Join(applied.BackupPath, "files", "skills", project, "doksuri-server-implement", "SKILL.md")); err != nil {
		t.Fatalf("approved repair did not back up removed skill: %v", err)
	}
	if _, err := os.Stat(filepath.Join(applied.BackupPath, "skill-pack-manifest.json.before")); err != nil {
		t.Fatalf("approved repair did not back up manifest before mutation: %v", err)
	}
	manifest := readProjectInstallManifest(t, filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))
	suite := manifest["project_suites"].([]any)[0].(map[string]any)
	if suite["suite_role"] != role || len(suite["installed_skills"].([]any)) != 2 || manifest["last_repair"] == nil {
		t.Fatalf("manifest was not written last with selected role subset evidence: %+v", manifest)
	}
	if !containsString(applied.NoSpillover.UnknownPersonalSkillsPreserved, filepath.ToSlash(filepath.Join("skills", project, "doksuri-server-local-note", "SKILL.md"))) {
		t.Fatalf("apply missing no-spillover evidence: %+v", applied.NoSpillover)
	}
}

func TestProjectMigrationCleanGenericAndManualTaskForUnmanifested(t *testing.T) {
	repo := makeProjectInstallRepo(t, map[string]string{"kkachi-plan": "---\nname: kkachi-plan\n---\n# kkachi-plan\n"})
	profileRoot := filepath.Join(t.TempDir(), "profile")
	generic := "---\nname: kkachi-plan\n---\n# kkachi-plan\n"
	writeProjectInstallFile(t, filepath.Join(profileRoot, "skills", "kkachi-plan", "SKILL.md"), generic)
	writeJSON(t, filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"), map[string]any{
		"version": ManifestVersion,
		"kind":    ProfileManifestKind,
		"profile": map[string]any{"name": "kwanwoo", "root": profileRoot},
		"installs": []any{map[string]any{
			"pack_id": "kkachi-plan", "target_path": "skills/kkachi-plan", "files": []any{map[string]any{"relative_path": "SKILL.md", "sha256": sha256Bytes([]byte(generic))}},
		}},
	})
	migration, err := BuildProjectMigrationDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot, FromGeneric: true})
	if err != nil {
		t.Fatal(err)
	}
	if !migration.OK || migration.SourcePack.Source != "default" || len(migration.PlannedActions) == 0 || len(migration.ManualSemanticPortTasks) != 0 {
		t.Fatalf("expected clean generic migration plan, got %+v", migration)
	}
	approved, err := ApplyApprovedMigration(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: profileRoot, FromGeneric: true}, migration.ApprovalRequest.EvidenceRef)
	if err != nil || !approved.OK || approved.MigrationID == "" {
		t.Fatalf("approved migration failed: result=%+v err=%v", approved, err)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "kkachi-plan", "SKILL.md")); err != nil {
		t.Fatalf("migration must retain generic source skill, got stat error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); err != nil {
		t.Fatalf("migration did not create project skill: %v", err)
	}

	manualRoot := filepath.Join(t.TempDir(), "profile")
	writeProjectInstallFile(t, filepath.Join(manualRoot, "skills", "kkachi-plan", "SKILL.md"), "locally modified\n")
	manual, err := BuildProjectMigrationDryRun(repo, ProjectSuiteOptions{Profile: "kwanwoo", Project: "doksuri-server", ProfileRoot: manualRoot, FromGeneric: true})
	if err != nil {
		t.Fatal(err)
	}
	if !manual.OK || len(manual.ManualSemanticPortTasks) == 0 || manual.ApprovalRequest.Required {
		t.Fatalf("expected manual task without approvable write: %+v", manual)
	}
}

func hasProjectSuiteDiag(diags []ProjectSuiteDiagnostic, condition string, severity string) bool {
	for _, diag := range diags {
		if diag.Condition == condition && diag.Severity == severity {
			return true
		}
	}
	return false
}

func firstProjectActionDiagnosticCode(result ProjectActionResult) string {
	if len(result.Diagnostics) == 0 {
		return ""
	}
	return result.Diagnostics[0].Code
}

func hasProjectActionDiagnostic(result ProjectActionResult, code string) bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func assertProjectAction(t *testing.T, result ProjectActionResult, action string, targetPath string) {
	t.Helper()
	for _, planned := range result.PlannedActions {
		if planned.Action == action && planned.TargetPath == targetPath {
			return
		}
	}
	t.Fatalf("missing action=%s target=%s in %+v", action, targetPath, result.PlannedActions)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
