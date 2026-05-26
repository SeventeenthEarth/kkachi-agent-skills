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
