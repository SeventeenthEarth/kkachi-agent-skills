package install

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func writeSkill(t *testing.T, dir string, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if body == "" {
		body = "# Skill\n\nBody\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: Sample\n---\n"+body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeManifest(t *testing.T, profileRoot string, manifest map[string]any) {
	t.Helper()
	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func shaFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestCreatePlanIsNoWriteAndHashIncludesSOTFields(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	profileRoot := filepath.Join(base, "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")

	result, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("dry-run created profile root: %v", err)
	}

	if !result.OK || result.Command != "install" || result.Mode != "dry_run" || result.CLIVersion != "0.1.0" {
		t.Fatalf("unexpected result basics: %+v", result)
	}
	if result.Requested.PackIDs[0] != "alpha" || len(result.Requested.CategoryExpansions) != 0 {
		t.Fatalf("unexpected requested: %+v", result.Requested)
	}
	if result.TargetProfile.Name != "demo" || result.TargetProfile.PreviousManifestSHA256 != nil {
		t.Fatalf("unexpected target profile: %+v", result.TargetProfile)
	}
	if result.Packs[0]["pack_checksum_algorithm"] != "sha256" || result.Packs[0]["source_path"] != "skills/alpha" || result.Packs[0]["target_path"] != "skills/alpha" {
		t.Fatalf("unexpected pack: %+v", result.Packs[0])
	}
	if result.ChangedPaths[0].Action != "create" || result.ChangedPaths[0].Path != "skills/alpha/SKILL.md" || result.ChangedPaths[0].PreviousSHA256 != nil || result.ChangedPaths[0].NewSHA256 == "" {
		t.Fatalf("unexpected changed path: %+v", result.ChangedPaths[0])
	}
	if len(result.BackupPlan) != 0 || !result.ApprovalRequest.Required || result.ApprovalRequest.DryRunPlanHash != result.DryRunPlanHash {
		t.Fatalf("unexpected approval/backup: %+v %+v", result.BackupPlan, result.ApprovalRequest)
	}
	if result.CanonicalPlan["command_mode"] != "install:dry_run" || result.CanonicalPlan["cli_version"] != "0.1.0" {
		t.Fatalf("unexpected canonical plan: %+v", result.CanonicalPlan)
	}
	if _, ok := result.CanonicalPlan["source_repo"].(map[string]any)["git_commit"]; !ok {
		t.Fatalf("canonical plan missing git_commit: %+v", result.CanonicalPlan)
	}
	expected := "sha256:" + hashCanonical(t, result.CanonicalPlan)
	if result.DryRunPlanHash != expected {
		t.Fatalf("hash = %s, want %s", result.DryRunPlanHash, expected)
	}
	if !contains(result.NextAction, "KAB is not required") || !contains(result.NextAction, "KAB-gated") {
		t.Fatalf("missing KAB boundary: %s", result.NextAction)
	}
}

func TestChangedPathsAreSortedByActionThenPath(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	profileRoot := filepath.Join(base, "profile")
	writeSkill(t, filepath.Join(repo, "skills", "beta"), "")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
	target := filepath.Join(profileRoot, "skills", "beta", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("local edit"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"beta", "alpha"}, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK {
		t.Fatal("expected conflict")
	}
	got := []string{}
	for _, entry := range result.ChangedPaths {
		got = append(got, entry.Action+" "+entry.Path)
	}
	want := append([]string{}, got...)
	sort.Strings(want)
	if jsonString(got) != jsonString(want) || jsonString(got) != jsonString([]string{"conflict skills/beta/SKILL.md", "create skills/alpha/SKILL.md"}) {
		t.Fatalf("unexpected order: %v", got)
	}
}

func TestManifestSkipUpdateConflictAndPathValidation(t *testing.T) {
	t.Run("matching manifest skips current file", func(t *testing.T) {
		base := t.TempDir()
		repo := filepath.Join(base, "repo")
		profileRoot := filepath.Join(base, "profile")
		writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
		target := filepath.Join(profileRoot, "skills", "alpha", "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		source, _ := os.ReadFile(filepath.Join(repo, "skills", "alpha", "SKILL.md"))
		if err := os.WriteFile(target, source, 0o644); err != nil {
			t.Fatal(err)
		}
		digest := shaFile(t, target)
		writeManifest(t, profileRoot, map[string]any{
			"version": "0.1",
			"kind":    "kas_profile_skill_manifest",
			"installs": []map[string]any{{
				"pack_id": "alpha", "target_path": "skills/alpha", "pack_checksum": "placeholder",
				"files": []map[string]any{{"relative_path": "SKILL.md", "sha256": digest, "new_sha256": digest}},
			}},
		})

		result, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
		if err != nil {
			t.Fatal(err)
		}
		if !result.OK || result.Packs[0]["installed_state"] != "installed_current" || result.ChangedPaths[0].Action != "skip" {
			t.Fatalf("unexpected skip result: %+v", result)
		}
	})

	t.Run("trusted manifest changed source plans update and backup", func(t *testing.T) {
		base := t.TempDir()
		repo := filepath.Join(base, "repo")
		profileRoot := filepath.Join(base, "profile")
		writeSkill(t, filepath.Join(repo, "skills", "alpha"), "# Skill\n\nnew\n")
		target := filepath.Join(profileRoot, "skills", "alpha", "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("---\nname: Sample\n---\n# Skill\n\nold\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		oldDigest := shaFile(t, target)
		writeManifest(t, profileRoot, map[string]any{
			"version": "0.1",
			"kind":    "kas_profile_skill_manifest",
			"installs": []map[string]any{{
				"pack_id": "alpha", "target_path": "skills/alpha", "pack_checksum": "old-pack",
				"files": []map[string]any{{"relative_path": "SKILL.md", "sha256": oldDigest}},
			}},
		})

		result, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
		if err != nil {
			t.Fatal(err)
		}
		if !result.OK || result.Packs[0]["installed_state"] != "installed_drifted" || result.ChangedPaths[0].Action != "update" || len(result.BackupPlan) != 1 {
			t.Fatalf("unexpected update result: %+v", result)
		}
	})

	t.Run("unsafe manifest paths fail closed", func(t *testing.T) {
		for _, unsafePath := range []string{"/tmp/alpha", "skills/../evil", "skills//alpha"} {
			base := t.TempDir()
			repo := filepath.Join(base, "repo")
			profileRoot := filepath.Join(base, "profile")
			writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
			writeManifest(t, profileRoot, map[string]any{
				"version": "0.1", "kind": "kas_profile_skill_manifest",
				"installs": []map[string]any{{"pack_id": "alpha", "target_path": unsafePath, "files": []map[string]any{}}},
			})

			result, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
			if err != nil {
				t.Fatal(err)
			}
			if result.OK || result.Diagnostics[0].Code != "unsafe_manifest_target_path" || result.ChangedPaths[0].Action != "error" {
				t.Fatalf("unexpected unsafe path result: %+v", result)
			}
		}
	})

	t.Run("mismatched and unsafe manifest file paths fail closed", func(t *testing.T) {
		base := t.TempDir()
		repo := filepath.Join(base, "repo")
		profileRoot := filepath.Join(base, "profile")
		writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
		writeManifest(t, profileRoot, map[string]any{
			"version": "0.1", "kind": "kas_profile_skill_manifest",
			"installs": []map[string]any{{
				"pack_id": "alpha", "target_path": "skills/beta",
				"files": []map[string]any{{"relative_path": "../SKILL.md", "sha256": "placeholder"}},
			}},
		})

		result, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
		if err != nil {
			t.Fatal(err)
		}
		if result.OK || result.ChangedPaths[0].ErrorCode != "manifest_target_path_mismatch" || result.ChangedPaths[1].ErrorCode != "unsafe_manifest_file_path" {
			t.Fatalf("unexpected mismatch result: %+v", result)
		}
	})
}

func TestManifestParseUnknownPackSymlinkAndProfileFailures(t *testing.T) {
	t.Run("malformed and unsupported manifests fail closed", func(t *testing.T) {
		base := t.TempDir()
		repo := filepath.Join(base, "repo")
		writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
		malformedProfile := filepath.Join(base, "malformed")
		manifestPath := filepath.Join(malformedProfile, ".kas", "skill-pack-manifest.json")
		if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(manifestPath, []byte("{not-json"), 0o644); err != nil {
			t.Fatal(err)
		}
		malformed, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: malformedProfile})
		if err != nil {
			t.Fatal(err)
		}
		if malformed.OK || malformed.Diagnostics[0].Code != "manifest_parse_error" {
			t.Fatalf("unexpected malformed result: %+v", malformed)
		}

		unsupportedProfile := filepath.Join(base, "unsupported")
		writeManifest(t, unsupportedProfile, map[string]any{"version": "9.9", "kind": "kas_profile_skill_manifest", "installs": []any{}})
		unsupported, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: unsupportedProfile})
		if err != nil {
			t.Fatal(err)
		}
		if unsupported.OK || unsupported.Diagnostics[0].Code != "unsupported_manifest_version" {
			t.Fatalf("unexpected unsupported result: %+v", unsupported)
		}
	})

	t.Run("unknown pack and source symlink fail closed", func(t *testing.T) {
		base := t.TempDir()
		repo := filepath.Join(base, "repo")
		profileRoot := filepath.Join(base, "profile")
		writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
		unknown, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"missing"}, ProfileRoot: profileRoot})
		if err != nil {
			t.Fatal(err)
		}
		if unknown.OK || unknown.Diagnostics[0].Code != "unknown_pack_id" {
			t.Fatalf("unexpected unknown result: %+v", unknown)
		}

		linkPack := filepath.Join(repo, "skills", "linkpack")
		writeSkill(t, linkPack, "")
		if err := os.Symlink(filepath.Join(repo, "skills", "alpha", "SKILL.md"), filepath.Join(linkPack, "link")); err != nil {
			t.Fatal(err)
		}
		symlink, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"linkpack"}, ProfileRoot: profileRoot})
		if err != nil {
			t.Fatal(err)
		}
		if symlink.OK || symlink.Diagnostics[0].Code != "source_symlink_rejected" || symlink.ChangedPaths[0].Action != "error" {
			t.Fatalf("unexpected symlink result: %+v", symlink)
		}
	})

	t.Run("unknown default profile fails closed without harness root", func(t *testing.T) {
		repo := filepath.Join(t.TempDir(), "repo")
		writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")

		result, err := BuildDryRun(repo, Options{Profile: "profile-that-should-not-exist-climvp-003", PackIDs: []string{"alpha"}})
		if err != nil {
			t.Fatal(err)
		}
		if result.OK || result.Diagnostics[0].Code != "unknown_profile" || result.ChangedPaths[0].Path != "." {
			t.Fatalf("unexpected unknown profile result: %+v", result)
		}
	})
}

func TestApprovedInstallRejectsWrongHashAndWritesNothing(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	profileRoot := filepath.Join(base, "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")

	result, err := ApplyApprovedInstall(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot}, "dry-run:sha256:wrong")
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Diagnostics[0].Code != "approval_plan_hash_mismatch" {
		t.Fatalf("unexpected approval result: %+v", result)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("wrong approval wrote profile root: %v", err)
	}
}

func TestApprovedInstallCreateCopiesFilesAndWritesManifest(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	profileRoot := filepath.Join(base, "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "# Skill\n\ncreated\n")
	dryRun, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}

	result, err := ApplyApprovedInstall(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot}, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Mode != "approved_copy" || result.Approval.DryRunPlanHash != dryRun.DryRunPlanHash || !result.Approval.MatchedCurrentPlan {
		t.Fatalf("unexpected approved result: %+v", result)
	}
	target := filepath.Join(profileRoot, "skills", "alpha", "SKILL.md")
	if string(mustRead(t, target)) != string(mustRead(t, filepath.Join(repo, "skills", "alpha", "SKILL.md"))) {
		t.Fatalf("target was not copied from source")
	}
	if result.ManifestPath != filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json") || result.BackupPath == "" || result.Recovery.BackupPath != result.BackupPath {
		t.Fatalf("missing manifest/backup/recovery: %+v", result)
	}
	if result.Summary.CountsByAction["manifest_update"] != 1 {
		t.Fatalf("missing manifest_update count: %+v", result.Summary)
	}
	var manifest map[string]any
	readJSONFile(t, result.ManifestPath, &manifest)
	installs := manifest["installs"].([]any)
	if len(installs) != 1 {
		t.Fatalf("unexpected installs: %+v", installs)
	}
	entry := installs[0].(map[string]any)
	if entry["pack_id"] != "alpha" || entry["install_id"] != result.InstallID || entry["dry_run_plan_hash"] != dryRun.DryRunPlanHash {
		t.Fatalf("unexpected manifest entry: %+v", entry)
	}
	if entry["approved_plan_hash"] != dryRun.DryRunPlanHash || entry["approval_evidence_ref"] != dryRun.ApprovalRequest.EvidenceRef {
		t.Fatalf("unexpected approval manifest entry: %+v", entry)
	}
	files := entry["files"].([]any)
	if files[0].(map[string]any)["action"] != "create" || files[0].(map[string]any)["backup_relative_path"] != nil {
		t.Fatalf("unexpected manifest files: %+v", files)
	}
}

func TestApprovedInstallRejectsSymlinkedTargetParentAndWritesNothing(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	profileRoot := filepath.Join(base, "profile")
	outside := filepath.Join(base, "outside")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "# Skill\n\ncreated\n")
	if err := os.MkdirAll(filepath.Join(profileRoot, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(profileRoot, "skills", "alpha")); err != nil {
		t.Fatal(err)
	}
	dryRun, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK || dryRun.ChangedPaths[0].Action != "create" {
		t.Fatalf("test setup expected an approvable create plan: %+v", dryRun)
	}

	result, err := ApplyApprovedInstall(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot}, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Diagnostics[0].Code != "unsafe_target_path" {
		t.Fatalf("expected symlink parent rejection, got: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(outside, "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("approved install wrote through symlinked parent: %v", err)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("approved install wrote manifest after symlink rejection: %v", err)
	}
}

func TestApprovedInstallUpdateBacksUpBeforeReplacingAndReplacesManifestEntry(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	profileRoot := filepath.Join(base, "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "# Skill\n\nnew\n")
	target := filepath.Join(profileRoot, "skills", "alpha", "SKILL.md")
	oldBody := []byte("---\nname: Sample\n---\n# Skill\n\nold\n")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, oldBody, 0o644); err != nil {
		t.Fatal(err)
	}
	oldDigest := shaFile(t, target)
	writeManifest(t, profileRoot, map[string]any{
		"version": "0.1",
		"kind":    "kas_profile_skill_manifest",
		"installs": []map[string]any{
			{"pack_id": "alpha", "target_path": "skills/alpha", "pack_checksum": "old-pack", "files": []map[string]any{{"relative_path": "SKILL.md", "sha256": oldDigest}}},
			{"pack_id": "alpha", "target_path": "skills/alpha", "pack_checksum": "duplicate", "files": []map[string]any{{"relative_path": "SKILL.md", "sha256": oldDigest}}},
			{"pack_id": "beta", "target_path": "skills/beta", "files": []map[string]any{}},
		},
	})
	dryRun, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	result, err := ApplyApprovedInstall(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot}, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Summary.CountsByAction["backup"] != 1 || result.Summary.CountsByAction["update"] != 1 {
		t.Fatalf("unexpected update result: %+v", result)
	}
	backupPath := filepath.Join(profileRoot, result.ChangedPaths[0].BackupPath)
	if string(mustRead(t, backupPath)) != string(oldBody) {
		t.Fatalf("backup did not preserve old file")
	}
	if string(mustRead(t, target)) != string(mustRead(t, filepath.Join(repo, "skills", "alpha", "SKILL.md"))) {
		t.Fatalf("target was not replaced")
	}
	var manifest map[string]any
	readJSONFile(t, result.ManifestPath, &manifest)
	installs := manifest["installs"].([]any)
	if len(installs) != 2 {
		t.Fatalf("expected duplicate alpha entries to be replaced by one current entry plus beta: %+v", installs)
	}
	alphaCount := 0
	for _, raw := range installs {
		entry := raw.(map[string]any)
		if entry["pack_id"] == "alpha" {
			alphaCount++
			if entry["install_id"] != result.InstallID {
				t.Fatalf("alpha manifest entry was not replaced: %+v", entry)
			}
			files := entry["files"].([]any)
			backupRel := files[0].(map[string]any)["backup_relative_path"].(string)
			if !contains(backupRel, result.InstallID) || contains(backupRel, "dry-run") {
				t.Fatalf("manifest used wrong backup path: %+v", files[0])
			}
		}
	}
	if alphaCount != 1 {
		t.Fatalf("expected one alpha entry, got %d in %+v", alphaCount, installs)
	}
}

func TestApprovedInstallRejectsConflictOrErrorPlanAndWritesNothing(t *testing.T) {
	t.Run("conflict", func(t *testing.T) {
		base := t.TempDir()
		repo := filepath.Join(base, "repo")
		profileRoot := filepath.Join(base, "profile")
		writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
		target := filepath.Join(profileRoot, "skills", "alpha", "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("local"), 0o644); err != nil {
			t.Fatal(err)
		}
		before := shaFile(t, target)
		dryRun, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
		if err != nil {
			t.Fatal(err)
		}
		result, err := ApplyApprovedInstall(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot}, dryRun.ApprovalRequest.EvidenceRef)
		if err != nil {
			t.Fatal(err)
		}
		if result.OK || result.Diagnostics[0].Code != "install_plan_not_approvable" || shaFile(t, target) != before {
			t.Fatalf("unexpected conflict approval result: %+v", result)
		}
		if _, err := os.Stat(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")); !os.IsNotExist(err) {
			t.Fatalf("conflict wrote manifest: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		base := t.TempDir()
		repo := filepath.Join(base, "repo")
		profileRoot := filepath.Join(base, "profile")
		writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
		dryRun, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"missing"}, ProfileRoot: profileRoot})
		if err != nil {
			t.Fatal(err)
		}
		result, err := ApplyApprovedInstall(repo, Options{Profile: "demo", PackIDs: []string{"missing"}, ProfileRoot: profileRoot}, dryRun.ApprovalRequest.EvidenceRef)
		if err != nil {
			t.Fatal(err)
		}
		if result.OK || result.Diagnostics[0].Code != "install_plan_not_approvable" {
			t.Fatalf("unexpected error approval result: %+v", result)
		}
		if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
			t.Fatalf("error approval wrote profile root: %v", err)
		}
	})
}

func hashCanonical(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func jsonString(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func contains(value, needle string) bool {
	return len(needle) == 0 || (len(value) >= len(needle) && stringContains(value, needle))
}

func stringContains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	data := mustRead(t, path)
	if err := json.Unmarshal(data, value); err != nil {
		t.Fatal(err)
	}
}
