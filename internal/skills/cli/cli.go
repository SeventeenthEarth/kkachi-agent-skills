package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/install"
)

func Main(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	if len(argv) == 0 {
		return emitError(stderr, "command_required", "expected list or install command", "", false, "")
	}
	if isHelpArg(argv[0]) {
		printRootHelp(stdout)
		return 0
	}
	switch argv[0] {
	case "list":
		return runList(argv[1:], stdout, stderr, env)
	case "install":
		return runInstall(argv[1:], stdout, stderr, env)
	default:
		return emitError(stderr, "unknown_command", "only the list and install commands are implemented", argv[0], false, "")
	}
}

func runList(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", "", "source KAS repo path")
	profile := fs.String("profile", "", "Hermes target profile name")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
	category := fs.String("category", "", "filter packs by category")
	jsonOutput := fs.Bool("json", false, "emit machine-readable JSON")
	fs.Bool("no-color", false, "accepted for stable CLI shape; output is uncolored")
	if hasHelpArg(argv) {
		fs.SetOutput(stdout)
		fs.Usage()
		return 0
	}
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	if *profileRoot != "" && envValue(env, "KAS_ALLOW_PROFILE_ROOT_OVERRIDE") != "1" {
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "list", *jsonOutput, "")
	}
	result, err := discovery.BuildListResult(*repo, discovery.ListOptions{Category: *category, Profile: *profile, ProfileRoot: *profileRoot})
	if err != nil {
		return emitError(stderr, "discovery_failed", err.Error(), "list", *jsonOutput, "")
	}
	if *jsonOutput {
		_ = writeJSON(stdout, result)
	} else {
		fmt.Fprintln(stdout, discovery.RenderHumanList(result))
	}
	return 0
}

func runInstall(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", "", "source KAS repo path")
	profile := fs.String("profile", "", "Hermes target profile name")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
	dryRun := fs.Bool("dry-run", false, "report planned changes without writing")
	approve := fs.String("approve", "", "approval evidence ref for future approved copy install")
	jsonOutput := fs.Bool("json", false, "emit machine-readable JSON")
	fs.Bool("no-color", false, "accepted for stable CLI shape; output is uncolored")
	if hasHelpArg(argv) {
		fs.SetOutput(stdout)
		fs.Usage()
		return 0
	}
	if err := fs.Parse(normalizeInstallArgs(argv)); err != nil {
		return 2
	}
	if *profileRoot != "" && envValue(env, "KAS_ALLOW_PROFILE_ROOT_OVERRIDE") != "1" {
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "install", *jsonOutput, "")
	}
	if *approve != "" {
		return emitError(stderr, "approved_install_not_implemented", "approved copy install is not implemented until CLIMVP-004; rerun with --dry-run only.", "install", *jsonOutput, "Use install --dry-run for CLIMVP-003. Approved writes remain closed until CLIMVP-004.")
	}
	if !*dryRun {
		return emitError(stderr, "install_requires_dry_run_or_approve", "install requires --dry-run; approved writes are not implemented until CLIMVP-004.", "install", *jsonOutput, "Rerun with install --profile <profile> <pack-id>... --dry-run.")
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "install requires --profile <profile>.", "install", *jsonOutput, "")
	}
	packIDs := fs.Args()
	if len(packIDs) == 0 {
		return emitError(stderr, "pack_id_required", "install requires at least one pack id.", "install", *jsonOutput, "")
	}

	result, err := install.BuildDryRun(*repo, install.Options{Profile: *profile, PackIDs: packIDs, ProfileRoot: *profileRoot})
	if err != nil {
		return emitError(stderr, "discovery_failed", err.Error(), "install", *jsonOutput, "")
	}
	out := stdout
	code := 0
	if !result.OK {
		out = stderr
		code = 2
	}
	if *jsonOutput {
		_ = writeJSON(out, result)
	} else {
		fmt.Fprintln(out, install.RenderHumanDryRun(result))
	}
	return code
}

func normalizeInstallArgs(argv []string) []string {
	rewritten := []string{}
	positionals := []string{}
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch arg {
		case "--repo", "--profile", "--profile-root", "--approve":
			rewritten = append(rewritten, arg)
			if i+1 < len(argv) {
				i++
				rewritten = append(rewritten, argv[i])
			}
		case "--dry-run", "--json", "--no-color":
			rewritten = append(rewritten, arg)
		default:
			if hasInstallFlagValue(arg) || arg == "--dry-run=true" || arg == "--json=true" || arg == "--no-color=true" {
				rewritten = append(rewritten, arg)
			} else {
				positionals = append(positionals, arg)
			}
		}
	}
	return append(rewritten, positionals...)
}

func hasInstallFlagValue(arg string) bool {
	for _, name := range []string{"--repo=", "--profile=", "--profile-root=", "--approve=", "--dry-run=", "--json=", "--no-color="} {
		if strings.HasPrefix(arg, name) {
			return true
		}
	}
	return false
}

func hasHelpArg(argv []string) bool {
	for _, arg := range argv {
		if isHelpArg(arg) {
			return true
		}
	}
	return false
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "-help"
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: kkachi-hermes-skills <command> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Available commands:")
	fmt.Fprintln(w, "  list     List available KAS skill packs")
	fmt.Fprintln(w, "  install  Plan a profile-scoped KAS skill-pack install")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Use \"kkachi-hermes-skills <command> --help\" for command options.")
}

func emitError(w io.Writer, code string, message string, command string, jsonOutput bool, nextAction string) int {
	if nextAction == "" {
		nextAction = "Fix the reported issue and rerun " + command + "."
	}
	payload := map[string]any{
		"ok":          false,
		"command":     command,
		"source_repo": nil,
		"packs":       []any{},
		"diagnostics": []discovery.Diagnostic{{Level: "error", Code: code, Message: message}},
		"next_action": nextAction,
	}
	if jsonOutput {
		_ = writeJSON(w, payload)
	} else {
		fmt.Fprintln(w, "오류: "+message)
	}
	return 2
}

func writeJSON(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func envValue(env map[string]string, key string) string {
	if env != nil {
		return env[key]
	}
	return os.Getenv(key)
}
