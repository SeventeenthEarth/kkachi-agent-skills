package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
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
}
