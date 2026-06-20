package kahrunner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDeclaredKAHCLIPathWinsOverOldAmbientPATH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixtures are POSIX-only")
	}
	repo := newRepo(t)
	declared := writeHelper(t, filepath.Join(repo, "toolchains", "kah", "kkachi-agent-helper"), "kkachi-agent-helper 0.1.12")
	writeToolchain(t, repo, "kah_cli: v0.1.12\nkah_cli_path: "+declared+"\n")
	pathDir := t.TempDir()
	writeHelper(t, filepath.Join(pathDir, helperName), "kkachi-agent-helper 0.1.9")
	t.Setenv("PATH", pathDir)
	nested := filepath.Join(repo, "nested")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("create nested workdir: %v", err)
	}

	result := Runner{}.Run(nested, "--version")
	if result.Err != nil {
		t.Fatalf("runner failed: %v\nstderr=%s", result.Err, result.Stderr)
	}
	if got := strings.TrimSpace(string(result.Stdout)); got != "kkachi-agent-helper 0.1.12" {
		t.Fatalf("runner used %q, want declared helper", got)
	}
}

func TestV1ToolchainKAHSelectionWinsOverOldAmbientPATH(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixtures are POSIX-only")
	}
	repo := newRepo(t)
	declared := writeHelper(t, filepath.Join(repo, "toolchains", "kah", "kkachi-agent-helper"), "kkachi-agent-helper 0.2.0")
	writeToolchain(t, repo, strings.Join([]string{
		`schema_version: "kkachi.toolchain.v1"`,
		"kah:",
		`  cli_version: "0.2.0"`,
		`  binary_path: "` + declared + `"`,
		"",
	}, "\n"))
	pathDir := t.TempDir()
	writeHelper(t, filepath.Join(pathDir, helperName), "kkachi-agent-helper 0.1.9")
	t.Setenv("PATH", pathDir)

	result := Runner{}.Run(repo, "--version")
	if result.Err != nil {
		t.Fatalf("runner failed: %v\nstderr=%s", result.Err, result.Stderr)
	}
	if got := strings.TrimSpace(string(result.Stdout)); got != "kkachi-agent-helper 0.2.0" {
		t.Fatalf("runner used %q, want v1 toolchain helper", got)
	}

	writeHelper(t, declared, "kkachi-agent-helper 0.1.9")
	result = Runner{}.Run(repo, "--version")
	if result.Err == nil {
		t.Fatalf("runner unexpectedly fell back to PATH after v1 declared helper mismatch")
	}
	diagnostic := string(result.Stderr)
	for _, want := range []string{
		"selected_source: .kkachi/toolchain.yaml",
		"expected_version: 0.2.0",
		"expected_path: " + declared,
		"path_fallback_refused: ambient PATH fallback was refused because repo KAH selection is explicit.",
	} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("diagnostic missing %q\n%s", want, diagnostic)
		}
	}
}

func TestDeclaredPathMismatchRefusesPATHWithDiagnostic(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixtures are POSIX-only")
	}
	repo := newRepo(t)
	declared := writeHelper(t, filepath.Join(repo, "toolchains", "kah", "kkachi-agent-helper"), "kkachi-agent-helper 0.1.10")
	writeToolchain(t, repo, "kah_cli: v0.1.12\nkah_cli_path: "+declared+"\n")
	pathDir := t.TempDir()
	writeHelper(t, filepath.Join(pathDir, helperName), "kkachi-agent-helper 0.1.12")
	t.Setenv("PATH", pathDir)

	result := Runner{}.Run(repo, "--version")
	if result.Err == nil {
		t.Fatalf("runner unexpectedly succeeded with mismatched declared helper")
	}
	diagnostic := string(result.Stderr)
	for _, want := range []string{
		"selected_source: .kkachi/toolchain.yaml",
		"expected_version: 0.1.12",
		"expected_path: " + declared,
		"resolved_binary_path: " + declared,
		"actual_version_output: kkachi-agent-helper 0.1.10",
		"path_fallback_refused: ambient PATH fallback was refused because repo KAH selection is explicit.",
		"preferred_durable_recovery: Install or update the KAH binary declared by .kkachi/toolchain.yaml",
	} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("diagnostic missing %q\n%s", want, diagnostic)
		}
	}
}

func TestInvalidDeclaredKAHCLIRefusesPATHWithDiagnostic(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixtures are POSIX-only")
	}
	repo := newRepo(t)
	writeToolchain(t, repo, "kah_cli: definitely-not-semver\n")
	pathDir := t.TempDir()
	writeHelper(t, filepath.Join(pathDir, helperName), "kkachi-agent-helper 0.1.9")
	t.Setenv("PATH", pathDir)

	result := Runner{}.Run(repo, "--version")
	if result.Err == nil {
		t.Fatalf("runner unexpectedly fell back to PATH after invalid declared kah_cli; stdout=%q", strings.TrimSpace(string(result.Stdout)))
	}
	diagnostic := string(result.Stderr)
	for _, want := range []string{
		"KAH runner resolution failed: declared kah_cli is not a supported semver token",
		"selected_source: .kkachi/toolchain.yaml",
		"expected_version: definitely-not-semver",
		"path_fallback_refused: ambient PATH fallback was refused because repo KAH selection is explicit.",
		"preferred_durable_recovery: Correct kah_cli to an exact semver token such as v0.1.12, or declare kah_cli_path with a matching executable; ambient PATH fallback is intentionally refused for invalid repo KAH selection.",
	} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("diagnostic missing %q\n%s", want, diagnostic)
		}
	}
}

func TestNonAbsoluteKKACHI_KAH_BINFailsClosed(t *testing.T) {
	repo := newRepo(t)
	writeToolchain(t, repo, "kah_cli: v0.1.12\n")
	t.Setenv("KKACHI_KAH_BIN", helperName)

	result := Runner{}.Run(repo, "--version")
	if result.Err == nil {
		t.Fatalf("runner unexpectedly accepted non-absolute KKACHI_KAH_BIN")
	}
	diagnostic := string(result.Stderr)
	for _, want := range []string{
		"selected_source: KKACHI_KAH_BIN",
		"resolved_binary_path: kkachi-agent-helper",
		"KKACHI_KAH_BIN must be an absolute executable path",
	} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("diagnostic missing %q\n%s", want, diagnostic)
		}
	}
}

func TestExactSemverTokenMatchingRejectsFalseSubstrings(t *testing.T) {
	for _, output := range []string{
		"kkachi-agent-helper 0.1.120",
		"kkachi-agent-helper 10.1.12",
		"wrapper output has no parsed helper version",
		"kkachi-agent-helper 0.1.12-source",
	} {
		if versionMatches("v0.1.12", output) {
			t.Fatalf("versionMatches accepted false output %q", output)
		}
	}
	for _, output := range []string{
		"kkachi-agent-helper 0.1.12",
		"kkachi-agent-helper v0.1.12\n",
		"(0.1.12)",
	} {
		if !versionMatches("v0.1.12", output) {
			t.Fatalf("versionMatches rejected valid output %q", output)
		}
	}
}

func TestDeclaredKAHCLIWithoutPathUsesRepoBinAndRefusesPATHOnMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixtures are POSIX-only")
	}
	repo := newRepo(t)
	writeToolchain(t, repo, "kah_cli: v0.1.12\n")
	repoBin := writeHelper(t, filepath.Join(repo, ".kkachi", "bin", helperName), "kkachi-agent-helper 0.1.12")
	pathDir := t.TempDir()
	writeHelper(t, filepath.Join(pathDir, helperName), "kkachi-agent-helper 0.1.9")
	t.Setenv("PATH", pathDir)

	result := Runner{}.Run(repo, "--version")
	if result.Err != nil {
		t.Fatalf("runner failed: %v\nstderr=%s", result.Err, result.Stderr)
	}
	if got := strings.TrimSpace(string(result.Stdout)); got != "kkachi-agent-helper 0.1.12" {
		t.Fatalf("runner used %q, want repo bin %s", got, repoBin)
	}

	writeHelper(t, repoBin, "kkachi-agent-helper 0.1.10")
	result = Runner{}.Run(repo, "--version")
	if result.Err == nil {
		t.Fatalf("runner unexpectedly fell back to PATH after repo-bin mismatch")
	}
	diagnostic := string(result.Stderr)
	for _, want := range []string{
		"selected_source: .kkachi/toolchain.yaml kah_cli via .kkachi/bin",
		"expected_version: 0.1.12",
		"resolved_binary_path: " + repoBin,
		"actual_version_output: kkachi-agent-helper 0.1.10",
		"path_fallback_refused: ambient PATH fallback was refused because repo KAH selection is explicit.",
	} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("diagnostic missing %q\n%s", want, diagnostic)
		}
	}
}

func TestDeclaredKAHCLIWithoutPathRefusesPATHWhenRepoBinMissing(t *testing.T) {
	repo := newRepo(t)
	writeToolchain(t, repo, "kah_cli: v0.1.12\n")
	pathDir := t.TempDir()
	writeHelper(t, filepath.Join(pathDir, helperName), "kkachi-agent-helper 0.1.12")
	t.Setenv("PATH", pathDir)

	result := Runner{}.Run(repo, "--version")
	if result.Err == nil {
		t.Fatalf("runner unexpectedly fell back to PATH when explicit repo bin was missing")
	}
	diagnostic := string(result.Stderr)
	if !strings.Contains(diagnostic, "declared KAH binary is missing or not executable") ||
		!strings.Contains(diagnostic, "path_fallback_refused: ambient PATH fallback was refused because repo KAH selection is explicit.") {
		t.Fatalf("diagnostic did not report missing explicit repo bin/no fallback\n%s", diagnostic)
	}
}

func TestAmbientPATHFallbackWorksWithoutRepoSelection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper fixtures are POSIX-only")
	}
	project := t.TempDir()
	pathDir := t.TempDir()
	writeHelper(t, filepath.Join(pathDir, helperName), "kkachi-agent-helper 0.1.9")
	t.Setenv("PATH", pathDir)

	result := Runner{}.Run(project, "--version")
	if result.Err != nil {
		t.Fatalf("runner failed: %v\nstderr=%s", result.Err, result.Stderr)
	}
	if got := strings.TrimSpace(string(result.Stdout)); got != "kkachi-agent-helper 0.1.9" {
		t.Fatalf("runner used %q, want ambient PATH helper", got)
	}
}

func newRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".kkachi"), 0755); err != nil {
		t.Fatalf("create repo .kkachi: %v", err)
	}
	return repo
}

func writeToolchain(t *testing.T, repo, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repo, ".kkachi", "toolchain.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("write toolchain: %v", err)
	}
}

func writeHelper(t *testing.T, path, version string) string {
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
