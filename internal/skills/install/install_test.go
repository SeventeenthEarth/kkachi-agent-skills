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
	marker := changedPathByPath(result.ChangedPaths, "skills/alpha/references/kab-adoption-stage.md")
	if marker == nil || marker.Action != "create" || marker.NewSHA256 == "" {
		t.Fatalf("missing generated marker changed path: %+v", result.ChangedPaths)
	}
	if result.KABAdoptionStage.Numeric != 1 || result.KABAdoptionStage.Canonical != KABStage1Canonical || result.KABAdoptionStage.Source != "default_stage1" || !result.KABAdoptionStage.HashBound {
		t.Fatalf("unexpected KAB adoption stage: %+v", result.KABAdoptionStage)
	}
	if len(result.BackupPlan) != 0 || !result.ApprovalRequest.Required || result.ApprovalRequest.DryRunPlanHash != result.DryRunPlanHash {
		t.Fatalf("unexpected approval/backup: %+v %+v", result.BackupPlan, result.ApprovalRequest)
	}
	if !result.ApprovalRequest.HashIncludesProvenance {
		t.Fatalf("approval request must declare provenance hash binding: %+v", result.ApprovalRequest)
	}
	if result.ProvenanceContractVersion == "" || result.SourceInventorySnapshot.Summary.CountsBySourceClass == nil || result.TargetProfileInventory.Summary.CountsBySourceClass == nil {
		t.Fatalf("missing provenance inventory fields: %+v %+v", result.SourceInventorySnapshot, result.TargetProfileInventory)
	}
	if result.CanonicalPlan["command_mode"] != "install:dry_run" || result.CanonicalPlan["cli_version"] != "0.1.0" {
		t.Fatalf("unexpected canonical plan: %+v", result.CanonicalPlan)
	}
	if result.CanonicalPlan["source_inventory_snapshot"] == nil || result.CanonicalPlan["target_profile_inventory"] == nil {
		t.Fatalf("canonical plan missing provenance hash inputs: %+v", result.CanonicalPlan)
	}
	if _, ok := result.CanonicalPlan["source_repo"].(map[string]any)["git_commit"]; !ok {
		t.Fatalf("canonical plan missing git_commit: %+v", result.CanonicalPlan)
	}
	expected := "sha256:" + hashCanonical(t, result.CanonicalPlan)
	if result.DryRunPlanHash != expected {
		t.Fatalf("hash = %s, want %s", result.DryRunPlanHash, expected)
	}
	repeated, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if repeated.DryRunPlanHash != result.DryRunPlanHash {
		t.Fatalf("dry-run hash changed across identical dry-runs: %s != %s", repeated.DryRunPlanHash, result.DryRunPlanHash)
	}
	writeSkill(t, filepath.Join(profileRoot, "skills", "personal"), "profile personal skill")
	withProfileSkill, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if withProfileSkill.DryRunPlanHash == result.DryRunPlanHash {
		t.Fatal("dry-run hash did not change after provenance-relevant profile inventory changed")
	}
	if !contains(result.NextAction, "KAB is not required") || !contains(result.NextAction, "KAB-gated") {
		t.Fatalf("missing KAB boundary: %s", result.NextAction)
	}
	if !contains(result.NextAction, "approve the current dry_run_plan_hash") || contains(result.NextAction, "CLIMVP-004") {
		t.Fatalf("unexpected approval guidance: %s", result.NextAction)
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
	if jsonString(got) != jsonString(want) || jsonString(got) != jsonString([]string{"conflict skills/beta/SKILL.md", "create skills/alpha/SKILL.md", "create skills/alpha/references/kab-adoption-stage.md", "create skills/beta/references/kab-adoption-stage.md"}) {
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
		skill := changedPathByPath(result.ChangedPaths, "skills/alpha/SKILL.md")
		if !result.OK || result.Packs[0]["installed_state"] != "installed_drifted" || skill == nil || skill.Action != "skip" || changedPathByPath(result.ChangedPaths, "skills/alpha/references/kab-adoption-stage.md") == nil {
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
		skill := changedPathByPath(result.ChangedPaths, "skills/alpha/SKILL.md")
		if !result.OK || result.Packs[0]["installed_state"] != "installed_drifted" || skill == nil || skill.Action != "update" || len(result.BackupPlan) != 1 {
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

func TestKABAdoptionStageParsingHashBindingAndSourceMarkerConflict(t *testing.T) {
	for _, tc := range []struct {
		name      string
		input     StageSelectionInput
		numeric   int
		canonical string
		source    string
		wantErr   bool
	}{
		{name: "default", input: StageSelectionInput{}, numeric: 1, canonical: KABStage1Canonical, source: "default_stage1"},
		{name: "numeric2", input: StageSelectionInput{Numeric: "2"}, numeric: 2, canonical: KABStage2Canonical, source: "explicit_numeric"},
		{name: "canonical1", input: StageSelectionInput{Canonical: KABStage1Canonical}, numeric: 1, canonical: KABStage1Canonical, source: "explicit_canonical"},
		{name: "matching aliases", input: StageSelectionInput{Numeric: "2", Canonical: KABStage2Canonical}, numeric: 2, canonical: KABStage2Canonical, source: "explicit_numeric"},
		{name: "conflict", input: StageSelectionInput{Numeric: "1", Canonical: KABStage2Canonical}, wantErr: true},
		{name: "stage3 numeric", input: StageSelectionInput{Numeric: "3"}, wantErr: true},
		{name: "stage3 canonical", input: StageSelectionInput{Canonical: kabStage3Canonical}, wantErr: true},
		{name: "malformed", input: StageSelectionInput{Numeric: "stage2"}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stage, err := ResolveKABAdoptionStage(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", stage)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if stage.Numeric != tc.numeric || stage.Canonical != tc.canonical || stage.Source != tc.source || !stage.HashBound {
				t.Fatalf("unexpected stage: %+v", stage)
			}
		})
	}

	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	profileRoot := filepath.Join(base, "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "")
	stage1, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot, KABStageSelection: StageSelectionInput{Numeric: "1"}})
	if err != nil {
		t.Fatal(err)
	}
	stage2, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot, KABStageSelection: StageSelectionInput{Canonical: KABStage2Canonical}})
	if err != nil {
		t.Fatal(err)
	}
	if stage1.DryRunPlanHash == stage2.DryRunPlanHash {
		t.Fatalf("stage hash did not change: %s", stage1.DryRunPlanHash)
	}
	if !contains(KABAdoptionStageMarkerContent(stage2.KABAdoptionStage), "operating-policy guidance only") || !contains(KABAdoptionStageMarkerContent(stage2.KABAdoptionStage), "not KAB execution evidence") {
		t.Fatalf("marker text is missing boundary language")
	}

	conflictRepo := filepath.Join(base, "conflict-repo")
	writeSkill(t, filepath.Join(conflictRepo, "skills", "alpha"), "")
	if err := os.MkdirAll(filepath.Join(conflictRepo, "skills", "alpha", "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(conflictRepo, "skills", "alpha", "references", "kab-adoption-stage.md"), []byte("source-owned\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	conflict, err := BuildDryRun(conflictRepo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: filepath.Join(base, "conflict-profile")})
	if err != nil {
		t.Fatal(err)
	}
	if conflict.OK || conflict.Diagnostics[0].Code != "source_kab_adoption_marker_conflict" {
		t.Fatalf("expected source marker conflict, got %+v", conflict)
	}
}

func TestKABAdoptionStageMarkerRunbookEvidencePostureAndReadback(t *testing.T) {
	stage1, err := ResolveKABAdoptionStage(StageSelectionInput{Numeric: "1"})
	if err != nil {
		t.Fatal(err)
	}
	stage1Marker := KABAdoptionStageMarkerContent(stage1)
	for _, needle := range []string{
		"Numeric stage: 1",
		"Canonical stage: " + KABStage1Canonical,
		"Selection source: explicit_stage_selection",
		"Runbook references",
		"Canonical runbook: skills/kkachi-install-guide/references/kas-kab-adoption-stage-runbook.md",
		"Generated marker path: references/kab-adoption-stage.md",
		"Selected-stage evidence posture",
		"direct Codex app-server prompt/session/output evidence",
		"no-KAB-Codex rationale",
		"Do not claim KAB Codex execution evidence for Stage 1 work.",
		"operating-policy guidance only",
		"not Stage 2 activation by itself",
		"not KAB execution evidence",
	} {
		if !contains(stage1Marker, needle) {
			t.Fatalf("Stage 1 marker missing %q:\n%s", needle, stage1Marker)
		}
	}
	parsed1, ok := ParseKABAdoptionStageMarker([]byte(stage1Marker))
	if !ok || parsed1.Numeric != 1 || parsed1.Canonical != KABStage1Canonical || parsed1.Source != "explicit_stage_selection" {
		t.Fatalf("Stage 1 marker readback failed: ok=%v stage=%+v marker=\n%s", ok, parsed1, stage1Marker)
	}

	stage2, err := ResolveKABAdoptionStage(StageSelectionInput{Canonical: KABStage2Canonical})
	if err != nil {
		t.Fatal(err)
	}
	stage2Marker := KABAdoptionStageMarkerContent(stage2)
	for _, needle := range []string{
		"Numeric stage: 2",
		"Canonical stage: " + KABStage2Canonical,
		"Selection source: explicit_stage_selection",
		"Runbook references",
		"Canonical runbook: skills/kkachi-install-guide/references/kas-kab-adoption-stage-runbook.md",
		"Generated marker path: references/kab-adoption-stage.md",
		"KAB native_codex after required preflight/session evidence",
		"break-glass approval and rationale",
		"KAB `native_codex` selected CLI/capability preflight",
		"KAB session/read/status/wait or retained stream evidence plus bridge events",
		"never as silent fallback",
		"not Stage 2 activation by itself",
		"not KAB execution evidence",
	} {
		if !contains(stage2Marker, needle) {
			t.Fatalf("Stage 2 marker missing %q:\n%s", needle, stage2Marker)
		}
	}
	parsed2, ok := ParseKABAdoptionStageMarker([]byte(stage2Marker))
	if !ok || parsed2.Numeric != 2 || parsed2.Canonical != KABStage2Canonical || parsed2.Source != "explicit_stage_selection" {
		t.Fatalf("Stage 2 marker readback failed: ok=%v stage=%+v marker=\n%s", ok, parsed2, stage2Marker)
	}
}

func TestApprovedInstallWritesKABMarkerManifestAndRejectsDifferentStage(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	profileRoot := filepath.Join(base, "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "# Skill\n\nstage\n")
	dryRun, err := BuildDryRun(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot, KABStageSelection: StageSelectionInput{Numeric: "2"}})
	if err != nil {
		t.Fatal(err)
	}
	mismatch, err := ApplyApprovedInstall(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot, KABStageSelection: StageSelectionInput{Numeric: "1"}}, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if mismatch.OK || mismatch.Diagnostics[0].Code != "approval_plan_hash_mismatch" {
		t.Fatalf("expected stage mismatch rejection, got %+v", mismatch)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("stage mismatch wrote profile root: %v", err)
	}

	result, err := ApplyApprovedInstall(repo, Options{Profile: "demo", PackIDs: []string{"alpha"}, ProfileRoot: profileRoot, KABStageSelection: StageSelectionInput{Canonical: KABStage2Canonical}}, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("approved install failed: %+v", result)
	}
	markerPath := filepath.Join(profileRoot, "skills", "alpha", "references", "kab-adoption-stage.md")
	marker := string(mustRead(t, markerPath))
	if !contains(marker, "Numeric stage: 2") || !contains(marker, KABStage2Canonical) || !contains(marker, "not Stage 2 activation") || !contains(marker, "not KAB execution evidence") {
		t.Fatalf("unexpected marker content:\n%s", marker)
	}
	if !contains(marker, "Selected at: "+result.InstallID) || !contains(marker, "Approval evidence: "+dryRun.ApprovalRequest.EvidenceRef) {
		t.Fatalf("marker is missing actual approved-install evidence:\n%s", marker)
	}
	if contains(marker, "approved install records install_id") || contains(marker, "approved install records approval_evidence_ref") {
		t.Fatalf("approved marker kept dry-run placeholder evidence:\n%s", marker)
	}
	markerSHA := shaFile(t, markerPath)
	markerChange := changedPathByPath(result.ChangedPaths, "skills/alpha/references/kab-adoption-stage.md")
	if markerChange == nil || markerChange.NewSHA256 != markerSHA || result.KABAdoptionStage.MarkerSHA256 != markerSHA {
		t.Fatalf("approved marker checksum not reflected in result: change=%+v stage=%+v actual=%s", markerChange, result.KABAdoptionStage, markerSHA)
	}
	var manifest map[string]any
	readJSONFile(t, result.ManifestPath, &manifest)
	entry := manifest["installs"].([]any)[0].(map[string]any)
	stage := entry["kab_adoption_stage"].(map[string]any)
	if stage["canonical"] != KABStage2Canonical || stage["marker_path"] != "skills/alpha/references/kab-adoption-stage.md" || stage["marker_sha256"] != markerSHA {
		t.Fatalf("manifest missing marker metadata: %+v", entry)
	}
	for _, raw := range entry["files"].([]any) {
		file := raw.(map[string]any)
		if file["relative_path"] == KABAdoptionMarkerRelativePath && file["sha256"] != markerSHA {
			t.Fatalf("manifest marker file checksum did not use approved marker content: %+v", file)
		}
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

func changedPathByPath(paths []ChangedPath, path string) *ChangedPath {
	for i := range paths {
		if paths[i].Path == path {
			return &paths[i]
		}
	}
	return nil
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
