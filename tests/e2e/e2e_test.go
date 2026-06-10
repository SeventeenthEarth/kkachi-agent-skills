package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func buildBinary(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	binary := filepath.Join(t.TempDir(), "kkachi-hermes-skills")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/kkachi-hermes-skills")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}
	return binary
}

func buildRootBinary(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	binary := filepath.Join(t.TempDir(), "kkachi-agent-skills")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("root go build failed: %v\n%s", err, output)
	}
	return binary
}

func assertNoHangul(t *testing.T, out string) {
	t.Helper()
	for _, r := range out {
		if r >= 0xAC00 && r <= 0xD7AF {
			t.Fatalf("expected no Korean prose in human output, got %q", out)
		}
	}
}

func TestRootInstalledBinaryVersion(t *testing.T) {
	binary := buildRootBinary(t)
	cmd := exec.Command(binary, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("root binary --version failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "kkachi-agent-skills 0.1.1" {
		t.Fatalf("unexpected --version output: %q", out)
	}
}

func TestRootInstalledBinaryUsesEmbeddedSourceOutsideRepo(t *testing.T) {
	binary := buildRootBinary(t)
	workdir := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	list := exec.Command(binary, "list", "--profile", "e2e", "--profile-root", profileRoot, "--json")
	list.Dir = workdir
	list.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	listOut, err := list.CombinedOutput()
	if err != nil {
		t.Fatalf("embedded-source list failed outside repo: %v\n%s", err, listOut)
	}
	var listPayload map[string]any
	if err := json.Unmarshal(listOut, &listPayload); err != nil {
		t.Fatal(err)
	}
	if listPayload["ok"] != true || len(listPayload["packs"].([]any)) == 0 {
		t.Fatalf("unexpected embedded-source list payload: %+v", listPayload)
	}
	if sourceRepo := listPayload["source_repo"].(map[string]any); !strings.Contains(sourceRepo["path"].(string), "embedded") {
		t.Fatalf("expected embedded source path outside repo, got %+v", sourceRepo)
	}

	install := exec.Command(binary, "install", "--profile", "e2e", "kkachi-plan", "--dry-run", "--profile-root", profileRoot, "--json")
	install.Dir = workdir
	install.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	installOut, err := install.CombinedOutput()
	if err != nil {
		t.Fatalf("embedded-source install dry-run failed outside repo: %v\n%s", err, installOut)
	}
	var installPayload map[string]any
	if err := json.Unmarshal(installOut, &installPayload); err != nil {
		t.Fatal(err)
	}
	if installPayload["ok"] != true || installPayload["mode"] != "dry_run" {
		t.Fatalf("unexpected embedded-source install payload: %+v", installPayload)
	}
}

func TestRootInstalledBinaryPrefersEmbeddedSourceOverArbitraryCwdSkills(t *testing.T) {
	binary := buildRootBinary(t)
	workdir := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")
	fakeSkillDir := filepath.Join(workdir, "skills", "fake-local-pack")
	if err := os.MkdirAll(fakeSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fakeSkillDir, "SKILL.md"), []byte("---\nname: fake-local-pack\ndescription: should not be discovered by embedded root binary\n---\n# Fake\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binary, "list", "--profile", "e2e", "--profile-root", profileRoot, "--json")
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("embedded-source list with cwd skills failed: %v\n%s", err, out)
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatal(err)
	}
	sourceRepo := payload["source_repo"].(map[string]any)
	if !strings.Contains(sourceRepo["path"].(string), "embedded://github.com/SeventeenthEarth/kkachi-agent-skills") {
		t.Fatalf("expected embedded source to override cwd skills, got %+v", sourceRepo)
	}
	for _, pack := range payload["packs"].([]any) {
		if pack.(map[string]any)["pack_id"] == "fake-local-pack" {
			t.Fatalf("root binary discovered arbitrary cwd skills instead of embedded source: %+v", payload)
		}
	}
}

func TestRootInstalledBinaryUsesEmbeddedSkillPackForProjectDryRunOutsideRepo(t *testing.T) {
	binary := buildRootBinary(t)
	workdir := t.TempDir()
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	cmd := exec.Command(binary, "install-project-kas", "--profile", "e2e", "--project", "doksuri-server", "--source-pack", "kas-default-project-suite", "--dry-run", "--profile-root", profileRoot, "--json")
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("embedded-source project dry-run failed outside repo: %v\n%s", err, out)
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "install-project-kas" || len(payload["planned_skills"].([]any)) == 0 {
		t.Fatalf("unexpected embedded-source project dry-run payload: %+v", payload)
	}
}

func TestRealRepoListAndInstallDryRunDoNotWriteProfile(t *testing.T) {
	root := repoRoot(t)
	binary := buildBinary(t)
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	list := exec.Command(binary, "list", "--repo", root, "--profile", "e2e", "--profile-root", profileRoot, "--json")
	list.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	listOut, err := list.CombinedOutput()
	if err != nil {
		t.Fatalf("list failed: %v\n%s", err, listOut)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("list created profile root: %v", err)
	}
	var listPayload map[string]any
	if err := json.Unmarshal(listOut, &listPayload); err != nil {
		t.Fatal(err)
	}
	if listPayload["ok"] != true || len(listPayload["packs"].([]any)) == 0 {
		t.Fatalf("unexpected list payload: %+v", listPayload)
	}
	if listPayload["provenance_contract_version"] != discovery.ProvenanceContractVersion || listPayload["source_inventory_summary"] == nil {
		t.Fatalf("list payload missing provenance fields: %+v", listPayload)
	}

	install := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--dry-run", "--profile-root", profileRoot, "--json")
	install.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	installOut, err := install.CombinedOutput()
	if err != nil {
		t.Fatalf("install failed: %v\n%s", err, installOut)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("install dry-run created profile root: %v", err)
	}
	var installPayload map[string]any
	if err := json.Unmarshal(installOut, &installPayload); err != nil {
		t.Fatal(err)
	}
	summary := installPayload["summary"].(map[string]any)
	counts := summary["counts_by_action"].(map[string]any)
	if installPayload["ok"] != true || installPayload["mode"] != "dry_run" || counts["create"].(float64) != summary["total_files"].(float64) {
		t.Fatalf("unexpected install payload: %+v", installPayload)
	}
	if installPayload["provenance_contract_version"] != discovery.ProvenanceContractVersion || installPayload["source_inventory_snapshot"] == nil || installPayload["approval_request"].(map[string]any)["hash_includes_provenance"] != true {
		t.Fatalf("install payload missing hash-bound provenance fields: %+v", installPayload)
	}
}

func TestRealRepoInstallProjectKASDryRunWritesNothing(t *testing.T) {
	root := repoRoot(t)
	binary := buildBinary(t)
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	cmd := exec.Command(binary, "install-project-kas", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--source-pack", "kas-default-project-suite", "--dry-run", "--profile-root", profileRoot, "--json")
	cmd.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install-project-kas dry-run failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("install-project-kas dry-run created profile root: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "install-project-kas" || payload["mode"] != "project_dry_run" || payload["dry_run"] != true {
		t.Fatalf("unexpected install-project-kas payload: %+v", payload)
	}
	noWrite := payload["no_write"].(map[string]any)
	if noWrite["guaranteed"] != true || noWrite["profile_write_count"] != float64(0) || noWrite["manifest_write_count"] != float64(0) || noWrite["kah_state_write_count"] != float64(0) || noWrite["kab_runtime_mutation_count"] != float64(0) {
		t.Fatalf("unexpected no-write evidence: %+v", noWrite)
	}
	planned := payload["planned_skills"].([]any)
	if len(planned) == 0 {
		t.Fatalf("expected planned project skills: %+v", payload)
	}
	foundPlan := false
	for _, raw := range planned {
		skill := raw.(map[string]any)
		if skill["source_pack_id"] == "kkachi-plan" && skill["installed_skill"] == "doksuri-server-plan" && skill["target_path"] == "skills/doksuri-server/doksuri-server-plan/SKILL.md" {
			foundPlan = true
		}
	}
	if !foundPlan || !strings.HasPrefix(payload["plan_hash"].(string), "sha256:") {
		t.Fatalf("missing project plan/hash evidence: %+v", payload)
	}
}

func TestRealRepoApprovedProjectKASInstallAndWrongHashNoWrite(t *testing.T) {
	root := repoRoot(t)
	binary := buildBinary(t)
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	wrong := exec.Command(binary, "install-project-kas", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--source-pack", "kas-default-project-suite", "--approve", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--profile-root", profileRoot, "--json")
	wrong.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	wrongOut, err := wrong.CombinedOutput()
	if err == nil {
		t.Fatalf("wrong hash approved project install unexpectedly succeeded: %s", wrongOut)
	}
	var wrongPayload map[string]any
	if err := json.Unmarshal(wrongOut, &wrongPayload); err != nil {
		t.Fatal(err)
	}
	if wrongPayload["ok"] != false || wrongPayload["diagnostics"].([]any)[0].(map[string]any)["code"] != "approval_plan_hash_mismatch" {
		t.Fatalf("unexpected wrong-hash payload: %+v", wrongPayload)
	}
	if _, err := os.Stat(profileRoot); !os.IsNotExist(err) {
		t.Fatalf("wrong hash created profile root: %v", err)
	}

	dryRun := exec.Command(binary, "install-project-kas", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--source-pack", "kas-default-project-suite", "--dry-run", "--profile-root", profileRoot, "--json")
	dryRun.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	dryRunOut, err := dryRun.CombinedOutput()
	if err != nil {
		t.Fatalf("project dry-run failed: %v\n%s", err, dryRunOut)
	}
	var dryRunPayload map[string]any
	if err := json.Unmarshal(dryRunOut, &dryRunPayload); err != nil {
		t.Fatal(err)
	}
	evidence := dryRunPayload["approval_request"].(map[string]any)["evidence_ref"].(string)
	approve := exec.Command(binary, "install-project-kas", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--source-pack", "kas-default-project-suite", "--approve", evidence, "--profile-root", profileRoot, "--json")
	approve.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	approveOut, err := approve.CombinedOutput()
	if err != nil {
		t.Fatalf("approved project install failed: %v\n%s", err, approveOut)
	}
	var approvePayload map[string]any
	if err := json.Unmarshal(approveOut, &approvePayload); err != nil {
		t.Fatal(err)
	}
	if approvePayload["ok"] != true || approvePayload["mode"] != "project_approved_copy" || approvePayload["install_id"] == "" {
		t.Fatalf("unexpected approved project payload: %+v", approvePayload)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")); err != nil {
		t.Fatalf("approved project install did not write expected project skill: %v", err)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))
	if err != nil {
		t.Fatalf("approved project install did not write manifest: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest["kind"] != "kas_profile_skill_manifest" || len(manifest["project_suites"].([]any)) != 1 {
		t.Fatalf("unexpected project manifest: %+v", manifest)
	}

	doctor := exec.Command(binary, "doctor", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--project-suite", "--profile-root", profileRoot, "--json")
	doctor.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	doctorOut, err := doctor.CombinedOutput()
	if err != nil {
		t.Fatalf("project-suite doctor after install failed: %v\n%s", err, doctorOut)
	}
	var doctorPayload map[string]any
	if err := json.Unmarshal(doctorOut, &doctorPayload); err != nil {
		t.Fatal(err)
	}
	if doctorPayload["ok"] != true || doctorPayload["mode"] != "project_suite_doctor" {
		t.Fatalf("unexpected project-suite doctor payload: %+v", doctorPayload)
	}

	target := filepath.Join(profileRoot, "skills", "doksuri-server", "doksuri-server-plan", "SKILL.md")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	brokenDoctor := exec.Command(binary, "doctor", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--project-suite", "--profile-root", profileRoot, "--json")
	brokenDoctor.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	brokenOut, err := brokenDoctor.CombinedOutput()
	if err == nil {
		t.Fatalf("broken project-suite doctor unexpectedly succeeded: %s", brokenOut)
	}
	var brokenPayload map[string]any
	if err := json.Unmarshal(brokenOut, &brokenPayload); err != nil {
		t.Fatal(err)
	}
	if brokenPayload["ok"] != false || brokenPayload["project_suite_diagnostics"] == nil {
		t.Fatalf("unexpected broken doctor payload: %+v", brokenPayload)
	}

	repairDryRun := exec.Command(binary, "repair-project-kas", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--dry-run", "--profile-root", profileRoot, "--json")
	repairDryRun.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	repairDryRunOut, err := repairDryRun.CombinedOutput()
	if err != nil {
		t.Fatalf("repair dry-run failed: %v\n%s", err, repairDryRunOut)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("repair dry-run wrote target: %v", err)
	}
	var repairPayload map[string]any
	if err := json.Unmarshal(repairDryRunOut, &repairPayload); err != nil {
		t.Fatal(err)
	}
	if repairPayload["ok"] != true || repairPayload["source_pack"].(map[string]any)["id"] != "kas-default-project-suite" {
		t.Fatalf("unexpected repair dry-run payload: %+v", repairPayload)
	}
	repairWrong := exec.Command(binary, "repair-project-kas", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--approve", "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000", "--profile-root", profileRoot, "--json")
	repairWrong.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	repairWrongOut, err := repairWrong.CombinedOutput()
	if err == nil {
		t.Fatalf("wrong repair hash unexpectedly succeeded: %s", repairWrongOut)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("wrong repair hash wrote target: %v", err)
	}
	evidence = repairPayload["approval_request"].(map[string]any)["evidence_ref"].(string)
	repairApprove := exec.Command(binary, "repair-project-kas", "--repo", root, "--profile", "e2e", "--project", "doksuri-server", "--approve", evidence, "--profile-root", profileRoot, "--json")
	repairApprove.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	repairApproveOut, err := repairApprove.CombinedOutput()
	if err != nil {
		t.Fatalf("approved repair failed: %v\n%s", err, repairApproveOut)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("approved repair did not restore target: %v", err)
	}
}

func TestRealRepoApprovedCopyWritesTempProfileOnly(t *testing.T) {
	root := repoRoot(t)
	binary := buildBinary(t)
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	dryRun := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--dry-run", "--profile-root", profileRoot, "--json")
	dryRun.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	dryRunOut, err := dryRun.CombinedOutput()
	if err != nil {
		t.Fatalf("dry-run failed: %v\n%s", err, dryRunOut)
	}
	var dryRunPayload map[string]any
	if err := json.Unmarshal(dryRunOut, &dryRunPayload); err != nil {
		t.Fatal(err)
	}
	evidence := dryRunPayload["approval_request"].(map[string]any)["evidence_ref"].(string)

	approve := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--approve", evidence, "--profile-root", profileRoot, "--json")
	approve.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	approveOut, err := approve.CombinedOutput()
	if err != nil {
		t.Fatalf("approve failed: %v\n%s", err, approveOut)
	}
	var approvePayload map[string]any
	if err := json.Unmarshal(approveOut, &approvePayload); err != nil {
		t.Fatal(err)
	}
	if approvePayload["ok"] != true || approvePayload["mode"] != "approved_copy" {
		t.Fatalf("unexpected approve payload: %+v", approvePayload)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, "skills", "kkachi-plan", "SKILL.md")); err != nil {
		t.Fatalf("approved copy did not write expected skill: %v", err)
	}
	if _, err := os.Stat(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json")); err != nil {
		t.Fatalf("approved copy did not write manifest: %v", err)
	}
}

func TestRealRepoDoctorReportsApprovedTempProfileReadOnly(t *testing.T) {
	root := repoRoot(t)
	binary := buildBinary(t)
	profileRoot := filepath.Join(t.TempDir(), "profiles", "e2e")

	dryRun := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--dry-run", "--profile-root", profileRoot, "--json")
	dryRun.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	dryRunOut, err := dryRun.CombinedOutput()
	if err != nil {
		t.Fatalf("dry-run failed: %v\n%s", err, dryRunOut)
	}
	var dryRunPayload map[string]any
	if err := json.Unmarshal(dryRunOut, &dryRunPayload); err != nil {
		t.Fatal(err)
	}
	evidence := dryRunPayload["approval_request"].(map[string]any)["evidence_ref"].(string)

	approve := exec.Command(binary, "install", "--repo", root, "--profile", "e2e", "kkachi-plan", "--approve", evidence, "--profile-root", profileRoot, "--json")
	approve.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	approveOut, err := approve.CombinedOutput()
	if err != nil {
		t.Fatalf("approve failed: %v\n%s", err, approveOut)
	}
	before, err := os.ReadFile(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	doctorJSON := exec.Command(binary, "doctor", "--repo", root, "--profile", "e2e", "--profile-root", profileRoot, "--json")
	doctorJSON.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	doctorJSONOut, err := doctorJSON.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor json failed: %v\n%s", err, doctorJSONOut)
	}
	var payload map[string]any
	if err := json.Unmarshal(doctorJSONOut, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["command"] != "doctor" || payload["manifest"].(map[string]any)["state"] != "ok" {
		t.Fatalf("unexpected doctor payload: %+v", payload)
	}
	if payload["provenance_contract_version"] != discovery.ProvenanceContractVersion || payload["provenance_audit"] == nil {
		t.Fatalf("doctor payload missing provenance audit: %+v", payload)
	}
	kab := payload["kab"].(map[string]any)
	if kab["required_for_minimum_cli"] != false || kab["required_for_execution_runtime"] != true {
		t.Fatalf("unexpected KAB payload: %+v", payload)
	}

	doctorHuman := exec.Command(binary, "doctor", "--repo", root, "--profile", "e2e", "--profile-root", profileRoot, "--no-color")
	doctorHuman.Env = append(os.Environ(), "KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1")
	doctorHumanOut, err := doctorHuman.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor human failed: %v\n%s", err, doctorHumanOut)
	}
	doctorHumanText := string(doctorHumanOut)
	assertNoHangul(t, doctorHumanText)
	if !strings.Contains(doctorHumanText, "Status:") || !strings.Contains(doctorHumanText, "healthy") {
		t.Fatalf("unexpected doctor human output: %s", doctorHumanText)
	}
	after, err := os.ReadFile(filepath.Join(profileRoot, ".kas", "skill-pack-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("doctor rewrote the install manifest")
	}
}

func TestSyncProjectKASValidateStateReadOnly(t *testing.T) {
	binary := buildBinary(t)
	repo, projectRoot, statePath := setupE2ESyncFixture(t)
	before, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(binary, "sync-project-kas", "--profile", "hwangchung", "--project", "kan-plugin", "--state", statePath, "--dry-run", "--json", "--repo", repo, "--project-root", projectRoot)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync-project-kas failed: %v\n%s", err, out)
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["mode"] != "dry_run_classification" || payload["yaml_state_path"] != statePath || payload["legacy_marker_path"] != filepath.Join(filepath.Dir(statePath), "kab-adoption-stage.md") {
		t.Fatalf("unexpected sync-project-kas payload: %+v", payload)
	}
	summary := payload["summary"].(map[string]any)
	counts := summary["counts_by_classification"].(map[string]any)
	if counts["semantic_merge_required"] != float64(1) || summary["semantic_port_packet_count"] != float64(1) || summary["write_count"] != float64(0) {
		t.Fatalf("unexpected classification summary: %+v", summary)
	}
	packets := payload["semantic_port_packets"].([]any)
	if len(packets) != 1 || !strings.Contains(packets[0].(map[string]any)["content"].(string), "no_write_statement") {
		t.Fatalf("unexpected semantic packet payload: %+v", packets)
	}
	after, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("sync-project-kas rewrote state file")
	}

	missingDryRun := exec.Command(binary, "sync-project-kas", "--profile", "hwangchung", "--project", "kan-plugin", "--state", statePath, "--json")
	missingOut, err := missingDryRun.CombinedOutput()
	if err == nil {
		t.Fatalf("expected missing --dry-run failure, got success: %s", missingOut)
	}
	var missingPayload map[string]any
	if err := json.Unmarshal(missingOut, &missingPayload); err != nil {
		t.Fatal(err)
	}
	if missingPayload["ok"] != false || missingPayload["diagnostics"].([]any)[0].(map[string]any)["code"] != "sync_project_kas_requires_dry_run" {
		t.Fatalf("unexpected missing dry-run payload: %+v", missingPayload)
	}
}

func setupE2ESyncFixture(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	repo := filepath.Join(dir, "upstream")
	projectRoot := filepath.Join(dir, "project")
	e2eGitInit(t, repo)
	writeE2ESkillPack(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "baseline")
	writeE2ESkillPack(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-plan"), "kkachi-plan", "baseline")
	e2eGitCommitAll(t, repo, "baseline")
	commit := strings.TrimSpace(e2eRunGit(t, repo, "rev-parse", "HEAD"))
	sourceChecksum := e2eChecksumFor(t, filepath.Join(repo, "skills", "kkachi-plan"))
	projectChecksum := e2eChecksumFor(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-plan"))
	writeE2ESkillPack(t, filepath.Join(repo, "skills", "kkachi-plan"), "kkachi-plan", "upstream changed")
	e2eGitCommitAll(t, repo, "current")
	writeE2ESkillPack(t, filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-plan"), "kkachi-plan", "local changed")

	state := strings.Replace(e2eValidKASState(), "commit: \"0123456789abcdef0123456789abcdef01234567\"", "commit: \""+commit+"\"", 1)
	state = strings.Replace(state, "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", sourceChecksum, 1)
	state = strings.Replace(state, "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", projectChecksum, 1)
	statePath := filepath.Join(projectRoot, "skills", "kan-plugin", "kan-plugin-kas", "references", "kas-project-state.yaml")
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}
	return repo, projectRoot, statePath
}

func writeE2ESkillPack(t *testing.T, dir string, name string, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\n---\n# " + name + "\n\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func e2eChecksumFor(t *testing.T, dir string) string {
	t.Helper()
	checksum, err := discovery.ComputePackChecksum(dir)
	if err != nil {
		t.Fatal(err)
	}
	return "sha256:" + checksum
}

func e2eGitInit(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	e2eRunGit(t, dir, "init")
	e2eRunGit(t, dir, "config", "user.email", "test@example.com")
	e2eRunGit(t, dir, "config", "user.name", "Test User")
}

func e2eGitCommitAll(t *testing.T, dir string, message string) {
	t.Helper()
	e2eRunGit(t, dir, "add", "-A")
	e2eRunGit(t, dir, "commit", "-m", message)
}

func e2eRunGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func e2eValidKASState() string {
	checksum := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	return `version: "0.1"
project:
  id: "kan-plugin"
  repo: "kkachi-agent-network-plugin"
  kas_suite: "kan-plugin"
  profile: "hwangchung"
kab_adoption_stage:
  numeric: 1
  canonical: "stage1_direct_codex_app_server_baseline"
  selection_source: "approved_project_policy"
  selected_at: "2026-06-06T00:00:00Z"
  approval_evidence: "not_applicable"
  stage2_activation: false
upstream_kas:
  repo: "kkachi-agent-skills"
  remote: "github.com/SeventeenthEarth/kkachi-agent-skills"
  commit: "0123456789abcdef0123456789abcdef01234567"
  dirty: false
  synced_at: "2026-06-06T00:00:00Z"
  sync_task: "KASUPD-001"
pack_baselines:
  - upstream_pack: "kkachi-plan"
    project_skill: "kan-plugin-plan"
    source_checksum: "` + checksum + `"
    project_checksum: "` + checksum + `"
    merge_mode: "semantic_port"
overlay_policy:
  local_overlay_allowed: true
  preserve_project_authority: true
  preserve_project_roadmap_ids: true
  preserve_project_test_commands: true
  preserve_role_labels: true
  overwrite_mode: "never_without_review"
update_policy:
  default_mode: "dry_run_then_semantic_merge"
  auto_apply_when:
    - "target_file_missing"
  require_llm_merge_when:
    - "project_skill_mapping_exists"
  fail_closed_when:
    - "state_file_missing"
    - "state_schema_invalid"
    - "stage_unsupported"
    - "upstream_commit_unknown"
    - "checksum_mismatch_without_baseline"
    - "auth_token_gateway_or_provider_mutation_detected"
evidence_posture:
  not_kab_runtime_evidence: true
  not_stage2_activation_by_itself: true
  missing_or_unreadable_fails_to_stage1_claims: true
`
}
