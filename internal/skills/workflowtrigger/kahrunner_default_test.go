package workflowtrigger

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCommandRunnerUsesRepoToolchainKAH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixture is POSIX-only")
	}
	project := t.TempDir()
	helper := writeWorkflowTriggerHelper(t, filepath.Join(project, "toolchain", "kkachi-agent-helper"), "kkachi-agent-helper 0.1.12")
	writeWorkflowTriggerToolchain(t, project, "kah_cli: v0.1.12\nkah_cli_path: "+helper+"\n")
	pathDir := t.TempDir()
	writeWorkflowTriggerHelper(t, filepath.Join(pathDir, "kkachi-agent-helper"), "kkachi-agent-helper 0.1.9")
	t.Setenv("PATH", pathDir)

	result := commandRunner{}.Run(project, "--version")
	if result.Err != nil {
		t.Fatalf("commandRunner failed: %v\nstderr=%s", result.Err, result.Stderr)
	}
	if got := strings.TrimSpace(string(result.Stdout)); got != "kkachi-agent-helper 0.1.12" {
		t.Fatalf("commandRunner used %q, want repo toolchain helper", got)
	}
}

func writeWorkflowTriggerToolchain(t *testing.T, project, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(project, ".kkachi"), 0755); err != nil {
		t.Fatalf("create .kkachi: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project, ".kkachi", "toolchain.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("write toolchain: %v", err)
	}
}

func writeWorkflowTriggerHelper(t *testing.T, path, version string) string {
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
