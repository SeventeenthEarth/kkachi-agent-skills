package doctor

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/install"
)

var ErrExit = errors.New("fake command exit")

type FakeRunner map[string]CommandResult

func (runner FakeRunner) Run(workDir string, args ...string) CommandResult {
	key := strings.Join(args, " ")
	if result, ok := runner[key]; ok {
		return result
	}
	return CommandResult{Err: ErrExit}
}

func writeSkill(t *testing.T, dir string, name string, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\n---\n# " + name + "\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func installProfile(t *testing.T, repo string, profileRoot string, packIDs ...string) {
	t.Helper()
	dryRun, err := install.BuildDryRun(repo, install.Options{Profile: "demo", PackIDs: packIDs, ProfileRoot: profileRoot})
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.OK {
		t.Fatalf("dry-run not ok: %+v", dryRun)
	}
	approved, err := install.ApplyApprovedInstall(repo, install.Options{Profile: "demo", PackIDs: packIDs, ProfileRoot: profileRoot}, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if !approved.OK {
		t.Fatalf("approved install not ok: %+v", approved)
	}
}

func diagCodes(result Result) map[string]bool {
	codes := map[string]bool{}
	for _, diagnostic := range result.Diagnostics {
		codes[diagnostic.Code] = true
	}
	return codes
}

func TestBuildHealthyInstalledProfileWithFakeKAH(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha", "healthy")
	installProfile(t, repo, profileRoot, "alpha")
	runner := FakeRunner{
		"--version":           {Stdout: []byte("kkachi-agent-helper 0.9.0\n")},
		"capabilities --json": {Stdout: []byte(`{"install_command":false,"commands":["project status","project doctor"]}`)},
	}

	result, err := Build(repo, Options{Profile: "demo", ProfileRoot: profileRoot, Runner: runner.Run})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatalf("expected healthy doctor, got %+v", result)
	}
	if result.Command != "doctor" || result.TargetProfile.State != "ok" || result.Manifest.State != "ok" {
		t.Fatalf("unexpected shape: %+v", result)
	}
	if len(result.InstalledPacks) != 1 || result.InstalledPacks[0].State != "ok" || result.InstalledPacks[0].FilesChecked == 0 {
		t.Fatalf("unexpected installed packs: %+v", result.InstalledPacks)
	}
	if result.ProvenanceContractVersion == "" || result.ProvenanceAudit.Summary.CountsBySourceClass == nil {
		t.Fatalf("missing provenance audit: %+v", result.ProvenanceAudit)
	}
	if result.InstalledPacks[0].SourceClass != "kas_managed_profile" || result.InstalledPacks[0].ProvenanceState != "classified" || len(result.InstalledPacks[0].SkillDependencies) != 0 || len(result.InstalledPacks[0].CommandSurfaceDependencies) != 0 {
		t.Fatalf("unexpected installed pack provenance: %+v", result.InstalledPacks[0])
	}
	if !result.KAH.Available || result.KAH.InstallCommand == nil || *result.KAH.InstallCommand {
		t.Fatalf("unexpected KAH payload: %+v", result.KAH)
	}
	if result.KAB.RequiredForMinimumCLI || !result.KAB.RequiredForExecutionRuntime || !strings.Contains(result.KAB.Message, "execution-runtime") {
		t.Fatalf("unexpected KAB payload: %+v", result.KAB)
	}
}

func TestBuildMissingAndMalformedManifest(t *testing.T) {
	repo := t.TempDir()
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha", "healthy")
	missingRoot := filepath.Join(t.TempDir(), "missing-profile")

	missing, err := Build(repo, Options{Profile: "demo", ProfileRoot: missingRoot, Runner: FakeRunner{}.Run})
	if err != nil {
		t.Fatal(err)
	}
	if missing.OK || !diagCodes(missing)["profile_missing"] || !diagCodes(missing)["manifest_missing"] {
		t.Fatalf("expected missing profile/manifest diagnostics, got %+v", missing)
	}
	if _, err := os.Stat(missingRoot); !os.IsNotExist(err) {
		t.Fatalf("doctor created missing profile root: %v", err)
	}

	profileRoot := filepath.Join(t.TempDir(), "profile")
	if err := os.MkdirAll(filepath.Join(profileRoot, ".kas"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"), []byte(`{"version":"9","kind":"kas_profile_skill_manifest"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	malformed, err := Build(repo, Options{Profile: "demo", ProfileRoot: profileRoot, Runner: FakeRunner{}.Run})
	if err != nil {
		t.Fatal(err)
	}
	if malformed.OK || !diagCodes(malformed)["unsupported_manifest_version"] {
		t.Fatalf("expected unsupported manifest, got %+v", malformed)
	}
}

func TestBuildDetectsMissingFileChecksumDriftUnknownPackAndUnsafePaths(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha", "healthy")
	installProfile(t, repo, profileRoot, "alpha")

	if err := os.Remove(filepath.Join(profileRoot, "skills", "alpha", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	missingFile, err := Build(repo, Options{Profile: "demo", ProfileRoot: profileRoot, Runner: FakeRunner{}.Run})
	if err != nil {
		t.Fatal(err)
	}
	if missingFile.OK || !diagCodes(missingFile)["installed_file_missing"] {
		t.Fatalf("expected missing file diagnostic, got %+v", missingFile)
	}

	installProfile(t, repo, profileRoot, "alpha")
	if err := os.WriteFile(filepath.Join(profileRoot, "skills", "alpha", "SKILL.md"), []byte("local edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	drift, err := Build(repo, Options{Profile: "demo", ProfileRoot: profileRoot, Runner: FakeRunner{}.Run})
	if err != nil {
		t.Fatal(err)
	}
	if drift.OK || !diagCodes(drift)["installed_file_checksum_mismatch"] {
		t.Fatalf("expected checksum mismatch, got %+v", drift)
	}

	manifestPath := filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	installs := manifest["installs"].([]any)
	unknown := map[string]any{
		"pack_id":       "ghost",
		"target_path":   "skills/ghost",
		"pack_checksum": "abc",
		"files":         []any{map[string]any{"relative_path": "SKILL.md", "sha256": "abc"}},
	}
	unsafe := map[string]any{
		"pack_id":       "alpha",
		"target_path":   "../escape",
		"pack_checksum": "abc",
		"files":         []any{map[string]any{"relative_path": "../SKILL.md", "sha256": "abc"}},
	}
	manifest["installs"] = append(installs, unknown, unsafe)
	rewritten, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, rewritten, 0o644); err != nil {
		t.Fatal(err)
	}
	badManifest, err := Build(repo, Options{Profile: "demo", ProfileRoot: profileRoot, Runner: FakeRunner{}.Run})
	if err != nil {
		t.Fatal(err)
	}
	codes := diagCodes(badManifest)
	if badManifest.OK || !codes["unknown_manifest_pack"] || !codes["unsafe_manifest_target_path"] || !codes["unsafe_manifest_file_path"] {
		t.Fatalf("expected manifest diagnostics, got %+v", badManifest)
	}
}

func TestBuildKAHMissingDegradedAndProjectProbes(t *testing.T) {
	repo := t.TempDir()
	project := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha", "healthy")
	installProfile(t, repo, profileRoot, "alpha")

	missing, err := Build(repo, Options{Profile: "demo", ProfileRoot: profileRoot, Project: project, Runner: FakeRunner{}.Run})
	if err != nil {
		t.Fatal(err)
	}
	if missing.OK || missing.KAH.Available || !diagCodes(missing)["kah_missing"] || !diagCodes(missing)["kah_project_status_skipped"] {
		t.Fatalf("expected KAH missing project failure, got %+v", missing)
	}

	runner := FakeRunner{
		"--version":             {Stdout: []byte("kkachi-agent-helper 0.9.0\n")},
		"capabilities --json":   {Stdout: []byte(`not-json`)},
		"project status --json": {Stdout: []byte(`{"ok":true}`)},
		"project doctor --json": {Stdout: []byte(`{"ok":false}`), Err: ErrExit},
	}
	degraded, err := Build(repo, Options{Profile: "demo", ProfileRoot: profileRoot, Project: project, Runner: runner.Run})
	if err != nil {
		t.Fatal(err)
	}
	codes := diagCodes(degraded)
	if degraded.OK || !degraded.KAH.Available || degraded.KAH.ProjectStatus != "ok" || degraded.KAH.ProjectDoctor != "failed" || !codes["kah_capabilities_degraded"] || !codes["kah_project_doctor_failed"] {
		t.Fatalf("expected degraded KAH project report, got %+v", degraded)
	}
}

func TestRenderHumanKoreanFriendlySummary(t *testing.T) {
	repo := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profile")
	writeSkill(t, filepath.Join(repo, "skills", "alpha"), "Alpha", "healthy")
	installProfile(t, repo, profileRoot, "alpha")

	result, err := Build(repo, Options{Profile: "demo", ProfileRoot: profileRoot, Runner: FakeRunner{"--version": {Stdout: []byte("kkachi-agent-helper 0.9.0\n")}}.Run})
	if err != nil {
		t.Fatal(err)
	}
	human := RenderHuman(result)
	for _, want := range []string{"상태:", "건강", "경고", "오류", "KAB"} {
		if !strings.Contains(human, want) {
			t.Fatalf("human output missing %q: %s", want, human)
		}
	}
}
