package kahrunner

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const helperName = "kkachi-agent-helper"

var semverTokenPattern = regexp.MustCompile(`(^|[^A-Za-z0-9._-])v?([0-9]+\.[0-9]+\.[0-9]+)([^A-Za-z0-9._-]|$)`)

type CommandResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

type Runner struct{}

type toolchainSelection struct {
	repoRoot               string
	toolchainPath          string
	expectedVersion        string
	invalidExpectedVersion string
	declaredPath           string
	readError              error
	explicit               bool
}

type resolution struct {
	source          string
	expectedVersion string
	expectedPath    string
	resolvedPath    string
	actualVersion   string
	refusedFallback bool
	recovery        string
}

type resolutionError struct {
	detail resolution
	reason string
}

func (e resolutionError) Error() string {
	return e.reason
}

func (Runner) Run(workDir string, args ...string) CommandResult {
	bin, err := Resolve(workDir)
	if err != nil {
		message := diagnostic(err)
		return CommandResult{Stderr: []byte(message), Err: errors.New(message)}
	}
	cmd := exec.Command(bin, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return CommandResult{Stdout: out, Stderr: exitErr.Stderr, Err: err}
		}
		return CommandResult{Stdout: out, Err: err}
	}
	return CommandResult{Stdout: out}
}

func Resolve(workDir string) (string, error) {
	selection := readToolchain(workDir)
	if envBin := strings.TrimSpace(os.Getenv("KKACHI_KAH_BIN")); envBin != "" {
		return resolveEnv(envBin, selection)
	}
	if selection.readError != nil {
		return "", resolveInvalidToolchain(selection, "repo KAH toolchain file is unreadable")
	}
	if selection.invalidExpectedVersion != "" {
		return "", resolveInvalidToolchain(selection, "declared kah_cli is not a supported semver token")
	}
	if selection.explicit {
		return resolveExplicit(selection)
	}
	if selection.repoRoot != "" {
		local := filepath.Join(selection.repoRoot, ".kkachi", "bin", helperName)
		if executable(local) {
			return local, nil
		}
	}
	bin, err := exec.LookPath(helperName)
	if err != nil {
		return "", resolutionError{
			detail: resolution{source: "PATH", recovery: "Install kkachi-agent-helper on PATH or declare kah_cli/kah_cli_path in .kkachi/toolchain.yaml for durable repo-local selection."},
			reason: "kkachi-agent-helper not found on PATH",
		}
	}
	return bin, nil
}

func resolveInvalidToolchain(selection toolchainSelection, reason string) error {
	recovery := "Correct kah_cli to an exact semver token such as v0.1.12, or declare kah_cli_path with a matching executable; ambient PATH fallback is intentionally refused for invalid repo KAH selection."
	if selection.readError != nil {
		recovery = "Fix permissions or syntax for .kkachi/toolchain.yaml, or remove the file only when no repo-local KAH selection is intended; ambient PATH fallback is intentionally refused while the repo toolchain is unreadable."
	}
	detail := resolution{
		source:          ".kkachi/toolchain.yaml",
		expectedVersion: selection.invalidExpectedVersion,
		expectedPath:    selection.declaredPath,
		refusedFallback: true,
		recovery:        recovery,
	}
	if detail.expectedVersion == "" {
		detail.expectedVersion = selection.expectedVersion
	}
	return resolutionError{detail: detail, reason: reason}
}

func resolveEnv(envBin string, selection toolchainSelection) (string, error) {
	detail := resolution{
		source:          "KKACHI_KAH_BIN",
		expectedVersion: selection.expectedVersion,
		expectedPath:    selection.declaredPath,
		resolvedPath:    envBin,
		recovery:        "Use KKACHI_KAH_BIN only as a temporary absolute-path debug override; the durable fix is to install or update the KAH binary declared by .kkachi/toolchain.yaml.",
	}
	if !filepath.IsAbs(envBin) {
		return "", resolutionError{detail: detail, reason: "KKACHI_KAH_BIN must be an absolute executable path"}
	}
	if !executable(envBin) {
		return "", resolutionError{detail: detail, reason: "KKACHI_KAH_BIN does not point to an executable file"}
	}
	if selection.expectedVersion != "" {
		actual, err := versionOutput(envBin)
		detail.actualVersion = actual
		if err != nil {
			return "", resolutionError{detail: detail, reason: "KKACHI_KAH_BIN version check failed"}
		}
		if !versionMatches(selection.expectedVersion, actual) {
			return "", resolutionError{detail: detail, reason: "KKACHI_KAH_BIN version does not match repo toolchain expectation"}
		}
	}
	return envBin, nil
}

func resolveExplicit(selection toolchainSelection) (string, error) {
	candidate := strings.TrimSpace(selection.declaredPath)
	source := ".kkachi/toolchain.yaml"
	if candidate == "" && selection.repoRoot != "" {
		candidate = filepath.Join(selection.repoRoot, ".kkachi", "bin", helperName)
		source = ".kkachi/toolchain.yaml kah_cli via .kkachi/bin"
	}
	detail := resolution{
		source:          source,
		expectedVersion: selection.expectedVersion,
		expectedPath:    selection.declaredPath,
		resolvedPath:    candidate,
		refusedFallback: true,
		recovery:        "Install or update the KAH binary declared by .kkachi/toolchain.yaml, or correct kah_cli/kah_cli_path; ambient PATH fallback is intentionally refused for explicit repo KAH selection.",
	}
	if candidate == "" {
		return "", resolutionError{detail: detail, reason: "explicit repo KAH selection did not provide a candidate path"}
	}
	if !filepath.IsAbs(candidate) {
		return "", resolutionError{detail: detail, reason: "declared kah_cli_path must be an absolute executable path"}
	}
	if !executable(candidate) {
		return "", resolutionError{detail: detail, reason: "declared KAH binary is missing or not executable"}
	}
	if selection.expectedVersion != "" {
		actual, err := versionOutput(candidate)
		detail.actualVersion = actual
		if err != nil {
			return "", resolutionError{detail: detail, reason: "declared KAH binary version check failed"}
		}
		if !versionMatches(selection.expectedVersion, actual) {
			return "", resolutionError{detail: detail, reason: "declared KAH binary version does not match repo toolchain expectation"}
		}
	}
	return candidate, nil
}

func diagnostic(err error) string {
	var rerr resolutionError
	if !errors.As(err, &rerr) {
		return err.Error()
	}
	d := rerr.detail
	lines := []string{
		"KAH runner resolution failed: " + rerr.reason,
		"selected_source: " + valueOrUnknown(d.source),
		"expected_version: " + valueOrNone(d.expectedVersion),
		"expected_path: " + valueOrNone(d.expectedPath),
		"resolved_binary_path: " + valueOrNone(d.resolvedPath),
		"actual_version_output: " + valueOrNone(strings.TrimSpace(d.actualVersion)),
	}
	if d.refusedFallback {
		lines = append(lines, "path_fallback_refused: ambient PATH fallback was refused because repo KAH selection is explicit.")
	} else {
		lines = append(lines, "path_fallback_refused: no explicit repo KAH selection required PATH refusal.")
	}
	lines = append(lines, "preferred_durable_recovery: "+valueOrUnknown(d.recovery))
	return strings.Join(lines, "\n")
}

func readToolchain(workDir string) toolchainSelection {
	repoRoot, toolchainPath := findToolchain(workDir)
	if toolchainPath == "" {
		return toolchainSelection{repoRoot: repoRoot}
	}
	data, err := os.ReadFile(toolchainPath)
	if err != nil {
		return toolchainSelection{repoRoot: repoRoot, toolchainPath: toolchainPath, readError: err, explicit: true}
	}
	values := parseTopLevelYAML(data)
	rawVersion, hasVersion := values["kah_cli"]
	rawVersion = strings.TrimSpace(rawVersion)
	version := normalizeExpectedVersion(rawVersion)
	declaredPath := strings.TrimSpace(values["kah_cli_path"])
	invalidVersion := ""
	if hasVersion && version == "" {
		invalidVersion = rawVersion
		if invalidVersion == "" {
			invalidVersion = "<empty>"
		}
	}
	return toolchainSelection{
		repoRoot:               repoRoot,
		toolchainPath:          toolchainPath,
		expectedVersion:        version,
		invalidExpectedVersion: invalidVersion,
		declaredPath:           declaredPath,
		explicit:               hasVersion || declaredPath != "",
	}
}

func findToolchain(workDir string) (string, string) {
	start := strings.TrimSpace(workDir)
	if start == "" {
		if cwd, err := os.Getwd(); err == nil {
			start = cwd
		}
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", ""
	}
	if st, err := os.Stat(abs); err == nil && !st.IsDir() {
		abs = filepath.Dir(abs)
	}
	for {
		candidate := filepath.Join(abs, ".kkachi", "toolchain.yaml")
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return abs, candidate
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", ""
		}
		abs = parent
	}
}

func parseTopLevelYAML(data []byte) map[string]string {
	values := map[string]string{}
	lines := bytes.Split(data, []byte("\n"))
	for _, raw := range lines {
		line := strings.TrimSpace(string(raw))
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		if len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t') {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(stripYAMLComment(value))
		value = strings.Trim(value, `"'`)
		if key != "" {
			values[key] = value
		}
	}
	return values
}

func stripYAMLComment(value string) string {
	inSingle := false
	inDouble := false
	for i, r := range value {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return strings.TrimSpace(value[:i])
			}
		}
	}
	return value
}

func executable(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir() && st.Mode().Perm()&0111 != 0
}

func versionOutput(bin string) (string, error) {
	cmd := exec.Command(bin, "--version")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func versionMatches(expected, output string) bool {
	expected = normalizeExpectedVersion(expected)
	if expected == "" {
		return true
	}
	for _, match := range semverTokenPattern.FindAllStringSubmatch(output, -1) {
		if len(match) >= 3 && match[2] == expected {
			return true
		}
	}
	return false
}

func normalizeExpectedVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	if semverTokenPattern.MatchString(version) {
		matches := semverTokenPattern.FindStringSubmatch(version)
		if len(matches) >= 3 {
			return matches[2]
		}
	}
	if regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(version) {
		return version
	}
	return ""
}

func valueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "<none>"
	}
	return value
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "<unknown>"
	}
	return value
}
