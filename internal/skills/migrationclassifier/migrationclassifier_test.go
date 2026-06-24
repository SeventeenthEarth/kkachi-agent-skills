package migrationclassifier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
)

func TestBuildClassifiesAllBucketsAndWritesNothing(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "hwangchung-profile")
	writeClassifierRepo(t, repo)
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "base plan")
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-review"), "kkachi-review", "base review with local delta")
	writeClassifierOverlay(t, filepath.Join(profileRoot, "skills", "doksuri", "kas-overlays", "doksuri-blue-plan-overlay"))
	writeClassifierWrapper(t, filepath.Join(profileRoot, "skills", "kkachi-blue-wrapper"))
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "personal-note"), "personal-note", "personal notes")
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kah-companion"), "kah-companion", "KAH companion surface for kkachi-agent-helper handoff")
	before, err := treeSHA256(profileRoot)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Build(repo, Options{Profile: "hwangchung", ProfileRoot: profileRoot, Project: "doksuri"})
	if err != nil {
		t.Fatal(err)
	}
	after, err := treeSHA256(profileRoot)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("classifier wrote to profile tree: before=%s after=%s", before, after)
	}
	if !result.OK || result.Mode != Mode {
		t.Fatalf("unexpected result header: %+v", result)
	}
	for _, bucket := range []string{"base_identical", "base_with_local_delta", "project_overlay_candidate", "role_wrapper_candidate", "unknown_personal_skill", "kah_companion_surface"} {
		if result.Summary.CountsByBucket[bucket] != 1 {
			t.Fatalf("bucket %s count = %d, want 1; items=%+v", bucket, result.Summary.CountsByBucket[bucket], result.Items)
		}
	}
	if !result.NoWriteEvidence.Guaranteed || !result.NoWriteEvidence.ProfileTreeUnchanged || !result.NoSpilloverEvidence.ProfileTreeHashMatch {
		t.Fatalf("missing no-write/no-spillover evidence: %+v %+v", result.NoWriteEvidence, result.NoSpilloverEvidence)
	}
	if result.NoWriteEvidence.ProfileSkillWriteCount != 0 ||
		result.NoWriteEvidence.ProfileSkillDeleteCount != 0 ||
		result.NoWriteEvidence.ProfileSkillMigrationCount != 0 ||
		result.NoWriteEvidence.KAHStateWriteCount != 0 ||
		result.NoWriteEvidence.KABRuntimeMutationCount != 0 ||
		result.NoWriteEvidence.HermesRuntimeMutationCount != 0 ||
		result.NoWriteEvidence.AuthProviderConfigWriteCount != 0 {
		t.Fatalf("unexpected no-write counters: %+v", result.NoWriteEvidence)
	}
	if result.NoSpilloverEvidence.ExternalProjectWriteCount != 0 || result.NoSpilloverEvidence.UnrequestedProfileTouched {
		t.Fatalf("unexpected no-spillover counters: %+v", result.NoSpilloverEvidence)
	}
	if result.ProvenanceContractVersion == "" || result.SourceClassEvidence == nil || result.DependencyAudit.State == "" || result.SkillDependencies == nil || result.CommandSurfaceDependencies == nil || result.DeletedBundleDiagnostics == nil {
		t.Fatalf("missing KASREL readback fields: %+v", result)
	}
	for _, item := range result.Items {
		if item.Bucket == "" || item.HashEvidence.ProfileSkillSHA256 == "" || item.HashEvidence.HashAlgorithm != "sha256" || item.SemanticExtractionPacket.PacketType == "" {
			t.Fatalf("item missing required evidence: %+v", item)
		}
		if len(item.ForbiddenActions) == 0 || item.NextAction == "" || item.Owner == "" || item.RecoveryHint == "" {
			t.Fatalf("item missing operator safety fields: %+v", item)
		}
	}
	human := RenderHuman(result)
	for _, want := range []string{"dry-run/report-only", "no writes performed", "no deletion, migration", "Next approval gate"} {
		if !strings.Contains(human, want) {
			t.Fatalf("human output missing %q:\n%s", want, human)
		}
	}
}

func TestBuildFailsClosedOnRuntimeConfigOwnershipConflictAndMissingInventory(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "hahuyeon-profile")
	writeClassifierRepo(t, repo)
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "base plan\nprovider: openai\nKAH owns plugin updates\n")

	result, err := Build(repo, Options{Profile: "hahuyeon", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK {
		t.Fatalf("expected runtime/ownership conflict to fail closed: %+v", result)
	}
	for _, code := range []string{"runtime_config_boundary_violation", "ownership_boundary_conflict"} {
		if !hasReason(result, code) {
			t.Fatalf("missing reason %s in %+v", code, result.ReasonCodes)
		}
	}

	missingResult, err := Build(repo, Options{Profile: "yeomong", ProfileRoot: filepath.Join(t.TempDir(), "missing-profile")})
	if err != nil {
		t.Fatal(err)
	}
	if missingResult.OK || !hasReason(missingResult, "profile_missing") {
		t.Fatalf("expected missing profile to fail closed: %+v", missingResult)
	}
}

func TestBuildFailsClosedOnEmptyJingungInventory(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "jingung-profile")
	writeClassifierRepo(t, repo)
	if err := os.MkdirAll(filepath.Join(profileRoot, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := Build(repo, Options{Profile: "jingung", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasReason(result, "profile_inventory_empty") {
		t.Fatalf("expected empty inventory to fail closed: %+v", result)
	}
}

func TestBuildDefaultProfileRootUsesHermesProfilesUnderHome(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	profileRoot := filepath.Join(home, ".hermes", "profiles", "hwangchung")
	t.Setenv("HOME", home)
	writeClassifierRepo(t, repo)
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "base plan")

	result, err := Build(repo, Options{Profile: "hwangchung"})
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(profileRoot)
	if err != nil {
		t.Fatal(err)
	}
	if result.TargetProfile.Root != want {
		t.Fatalf("target profile root = %q, want %q", result.TargetProfile.Root, want)
	}
	if !result.OK || result.Summary.CountsByBucket["base_identical"] != 1 {
		t.Fatalf("unexpected default-root result: %+v", result)
	}
}

func TestBuildPreservesNoWriteEvidenceWithDirectorySymlinkInProfileRoot(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "hwangchung-profile")
	targetDir := filepath.Join(t.TempDir(), "uv-archive")
	writeClassifierRepo(t, repo)
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "base plan")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "pyvenv.cfg"), []byte("home = /tmp\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(profileRoot, "home", ".cache", "uv", "environments-v2", "a")
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetDir, filepath.Join(linkDir, "symlink-dir")); err != nil {
		t.Fatal(err)
	}

	result, err := Build(repo, Options{Profile: "hwangchung", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || hasReason(result, "profile_hash_unavailable") {
		t.Fatalf("directory symlink should not break profile hash evidence: %+v", result)
	}
	if !result.NoWriteEvidence.ProfileTreeUnchanged || !result.NoSpilloverEvidence.ProfileTreeHashMatch {
		t.Fatalf("missing no-write evidence with directory symlink: %+v %+v", result.NoWriteEvidence, result.NoSpilloverEvidence)
	}
}

func TestBuildPreservesNoWriteEvidenceWithDirectorySymlinkInSkillsTree(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "hwangchung-profile")
	targetDir := filepath.Join(t.TempDir(), "shared-skill-assets")
	writeClassifierRepo(t, repo)
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "base plan")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "README.md"), []byte("shared asset\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(profileRoot, "skills", "kkachi-plan", "references")
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(targetDir, filepath.Join(linkDir, "shared-assets")); err != nil {
		t.Fatal(err)
	}

	result, err := Build(repo, Options{Profile: "hwangchung", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || hasReason(result, "profile_hash_unavailable") {
		t.Fatalf("directory symlink inside skills tree should not break profile hash evidence: %+v", result)
	}
	if !result.NoWriteEvidence.ProfileTreeUnchanged || !result.NoSpilloverEvidence.ProfileTreeHashMatch {
		t.Fatalf("missing no-write evidence with skills-tree directory symlink: %+v %+v", result.NoWriteEvidence, result.NoSpilloverEvidence)
	}
}

func TestRuntimeConfigDetectionIgnoresProseModelLabel(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "hwangchung-profile")
	writeClassifierRepo(t, repo)
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "Do not implement from chat-only instruction. KAS/KAH development uses a staged KAB adoption model: Stage 1 records evidence.\n")

	result, err := Build(repo, Options{Profile: "hwangchung", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || hasReason(result, "runtime_config_boundary_violation") {
		t.Fatalf("prose model label should not trigger runtime config boundary violation: %+v", result)
	}

	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-review"), "kkachi-review", "provider: openai\n")
	result, err = Build(repo, Options{Profile: "hwangchung", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasReason(result, "runtime_config_boundary_violation") {
		t.Fatalf("actual provider config key should still fail closed: %+v", result)
	}
}

func TestNoWriteEvidenceHashesStableSkillsTreeNotVolatileProfileHome(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "hwangchung-profile")
	writeClassifierRepo(t, repo)
	writeClassifierSkill(t, filepath.Join(profileRoot, "skills", "kkachi-plan"), "kkachi-plan", "base plan")
	volatileDir := filepath.Join(profileRoot, "home", ".codex")
	if err := os.MkdirAll(volatileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(volatileDir, "logs.sqlite-wal"), []byte("volatile runtime state\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	skillsHash, err := treeSHA256(filepath.Join(profileRoot, "skills"))
	if err != nil {
		t.Fatal(err)
	}

	result, err := Build(repo, Options{Profile: "hwangchung", ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("volatile profile home should not block skill migration classifier: %+v", result)
	}
	if result.NoWriteEvidence.ProfileTreeHashBefore != skillsHash || result.NoWriteEvidence.ProfileTreeHashAfter != skillsHash {
		t.Fatalf("no-write hashes should cover stable skills tree: before=%s after=%s skills=%s", result.NoWriteEvidence.ProfileTreeHashBefore, result.NoWriteEvidence.ProfileTreeHashAfter, skillsHash)
	}
}

func TestBuildDefaultProfileRootFailsClosedWhenMissing(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	profile := "missing-default"
	t.Setenv("HOME", home)
	writeClassifierRepo(t, repo)

	result, err := Build(repo, Options{Profile: profile})
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(filepath.Join(home, ".hermes", "profiles", profile))
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasReason(result, "profile_missing") {
		t.Fatalf("expected missing default profile to fail closed: %+v", result)
	}
	if result.TargetProfile.Root != want || !strings.Contains(result.TargetProfile.Root, filepath.Join(".hermes", "profiles")) {
		t.Fatalf("target profile root = %q, want under %q", result.TargetProfile.Root, want)
	}
}

func TestBuildRejectsUnsafeProfileNamesBeforeDefaultRootResolution(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeClassifierRepo(t, repo)

	for _, profile := range []string{"..", ".", " ../escape ", "team/blue", `team\blue`, ""} {
		result, err := Build(repo, Options{Profile: profile})
		if err != nil {
			t.Fatalf("Build(%q) returned error: %v", profile, err)
		}
		if result.OK || !hasReason(result, "profile_invalid") && strings.TrimSpace(profile) != "" {
			t.Fatalf("expected invalid profile %q to fail closed: %+v", profile, result)
		}
		if strings.TrimSpace(profile) != "" && result.TargetProfile.Root != "" {
			t.Fatalf("unsafe profile %q resolved target root %q", profile, result.TargetProfile.Root)
		}
		if strings.TrimSpace(profile) == "" && !hasReason(result, "profile_required") {
			t.Fatalf("expected empty profile to require profile: %+v", result)
		}
	}
}

func TestFinalizeEmptyInventoryDoesNotDuplicateDiagnostics(t *testing.T) {
	result := Result{
		OK:            true,
		TargetProfile: TargetProfile{Name: "hwangchung", State: "ok"},
		Items:         []Item{},
		Diagnostics:   []discovery.Diagnostic{{Level: "error", Code: "global_failure", Message: "global failure"}},
		Summary:       Summary{CountsByBucket: emptyBucketCounts()},
	}

	finalize(&result)
	finalize(&result)

	if got := countReason(result, "profile_inventory_empty"); got != 1 {
		t.Fatalf("profile_inventory_empty reason count = %d, want 1; diagnostics=%+v reasons=%+v", got, result.Diagnostics, result.ReasonCodes)
	}
	if got := countDiagnostics(result, "profile_inventory_empty"); got != 1 {
		t.Fatalf("profile_inventory_empty diagnostic count = %d, want 1; diagnostics=%+v", got, result.Diagnostics)
	}
	if got := countDiagnostics(result, "global_failure"); got != 1 {
		t.Fatalf("global_failure diagnostic count = %d, want 1; diagnostics=%+v", got, result.Diagnostics)
	}
	if result.Summary.ErrorCount != 2 {
		t.Fatalf("error count = %d, want 2; diagnostics=%+v", result.Summary.ErrorCount, result.Diagnostics)
	}
}

func TestParseMetadataReadsOnlyTopLevelAndMetadataKASKeys(t *testing.T) {
	meta := parseMetadata([]byte(`---
name: top-name
kind: top-kind
metadata:
  unrelated:
    role: ignored-role
    overlay_for: ignored-overlay
  kas:
    kind: project_overlay
    role: blue_commander
    role_manifest: kkachi-agent-skills:roles/blue.yaml
    plugin_namespace: kkachi-agent-skills
    overlay_root: skills/<project>/kas-overlays
    project: doksuri
    overlay_for: kkachi-agent-skills:plan
    merge_mode: additive_constraints
    base_version: v1
---
# skill
`))
	if meta.Name != "top-name" ||
		meta.Kind != "project_overlay" ||
		meta.Role != "blue_commander" ||
		meta.RoleManifest != "kkachi-agent-skills:roles/blue.yaml" ||
		meta.PluginNamespace != "kkachi-agent-skills" ||
		meta.OverlayRoot != "skills/<project>/kas-overlays" ||
		meta.Project != "doksuri" ||
		meta.OverlayFor != "kkachi-agent-skills:plan" ||
		meta.MergeMode != "additive_constraints" ||
		meta.BaseVersion != "v1" {
		t.Fatalf("unexpected metadata parse: %+v", meta)
	}

	dotted := parseMetadata([]byte(`---
metadata.kas.kind: color_wrapper
metadata.other.role: ignored-role
---
`))
	if dotted.Kind != "color_wrapper" || dotted.Role != "" {
		t.Fatalf("unexpected dotted metadata parse: %+v", dotted)
	}
}

func writeClassifierRepo(t *testing.T, repo string) {
	t.Helper()
	writeClassifierSkill(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "base plan")
	writeClassifierSkill(t, filepath.Join(repo, "skills", "kkachi-review"), "kkachi-review", "base review")
	manifest := `name: kkachi-agent-skills
version: 0.1.0
plugin:
  namespace: kkachi-agent-skills
  package_manifest: skill-pack.yaml
  load_policy: plugin_qualified_fail_closed
roles:
  blue: roles/blue.yaml
skills:
  - kkachi-plan
  - kkachi-review
`
	if err := os.MkdirAll(filepath.Join(repo, "roles"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "skill-pack.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "roles", "blue.yaml"), []byte("version: kas-plugin-role-manifest/v1\nrole: blue\nskills:\n  - kkachi-plan\n  - kkachi-review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeClassifierSkill(t *testing.T, dir string, name string, body string) {
	t.Helper()
	content := "---\nname: " + name + "\n---\n# " + name + "\n" + body + "\n"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeClassifierOverlay(t *testing.T, dir string) {
	t.Helper()
	content := `---
name: doksuri-blue-plan-overlay
metadata:
  kas:
    kind: project_overlay
    project: doksuri
    overlay_for: kkachi-agent-skills:plan
    merge_mode: additive_constraints
---
# Overlay
project-specific plan guidance
`
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeClassifierWrapper(t *testing.T, dir string) {
	t.Helper()
	content := `---
name: kkachi-blue-wrapper
metadata:
  kas:
    kind: color_wrapper
    role: blue_commander
    role_manifest: kkachi-agent-skills:roles/blue.yaml
    plugin_namespace: kkachi-agent-skills
    overlay_root: skills/<project>/kas-overlays
---
# Wrapper
thin wrapper
`
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasReason(result Result, code string) bool {
	for _, reason := range result.ReasonCodes {
		if reason == code {
			return true
		}
	}
	return false
}

func countReason(result Result, code string) int {
	count := 0
	for _, reason := range result.ReasonCodes {
		if reason == code {
			count++
		}
	}
	return count
}

func countDiagnostics(result Result, code string) int {
	count := 0
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == code {
			count++
		}
	}
	return count
}
