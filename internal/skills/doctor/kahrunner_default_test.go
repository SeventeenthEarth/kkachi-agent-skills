package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultRunnerUsesRepoToolchainKAH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixture is POSIX-only")
	}
	project := t.TempDir()
	helper := writeDoctorHelper(t, filepath.Join(project, "toolchain", "kkachi-agent-helper"), "kkachi-agent-helper 0.1.11")
	writeDoctorToolchain(t, project, "kah_cli: v0.1.11\nkah_cli_path: "+helper+"\n")
	pathDir := t.TempDir()
	writeDoctorHelper(t, filepath.Join(pathDir, "kkachi-agent-helper"), "kkachi-agent-helper 0.1.9")
	t.Setenv("PATH", pathDir)

	result := defaultRunner(project, "--version")
	if result.Err != nil {
		t.Fatalf("defaultRunner failed: %v\nstderr=%s", result.Err, result.Stderr)
	}
	if got := strings.TrimSpace(string(result.Stdout)); got != "kkachi-agent-helper 0.1.11" {
		t.Fatalf("defaultRunner used %q, want repo toolchain helper", got)
	}
}

func TestProbeKAHUsesProjectToolchainForGlobalProbe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixture is POSIX-only")
	}
	project := t.TempDir()
	helper := writeDoctorHelper(t, filepath.Join(project, "toolchain", "kkachi-agent-helper"), "kkachi-agent-helper 0.1.11")
	writeDoctorToolchain(t, project, "kah_cli: v0.1.11\nkah_cli_path: "+helper+"\n")
	pathDir := t.TempDir()
	writeDoctorHelper(t, filepath.Join(pathDir, "kkachi-agent-helper"), "kkachi-agent-helper 0.1.9")
	t.Setenv("KKACHI_KAH_BIN", "")
	t.Setenv("PATH", pathDir)
	oldCWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	isolatedCWD := t.TempDir()
	if err := os.Chdir(isolatedCWD); err != nil {
		t.Fatalf("chdir isolated cwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldCWD) })

	kah := probeKAH(defaultRunner, project)
	if !kah.Available {
		t.Fatalf("probeKAH did not find repo toolchain KAH")
	}
	if kah.Version != "kkachi-agent-helper 0.1.11" {
		t.Fatalf("probeKAH version = %q, want repo toolchain helper", kah.Version)
	}
}

func writeDoctorToolchain(t *testing.T, project, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(project, ".kkachi"), 0755); err != nil {
		t.Fatalf("create .kkachi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project, ".kkachi", "toolchain.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("write toolchain: %v", err)
	}
}

func writeDoctorHelper(t *testing.T, path, version string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create helper dir: %v", err)
	}
	content := "#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then\n  echo '" + version + "'\nelse\n  echo \"$@\"\nfi\n"
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("write helper: %v", err)
	}
	return path
}
