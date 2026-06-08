package cli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/doctor"
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/install"
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/kasstate"
	"github.com/SeventeenthEarth/kkachi-hermes-skills/internal/skills/projectinstall"
)

var installPromptInput io.Reader = os.Stdin
var installPromptInteractive = func() bool {
	file, ok := installPromptInput.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func Main(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	if len(argv) == 0 {
		return emitError(stderr, "command_required", "expected list, install, doctor, sync-project-kas, or install-project-kas command", "", false, "")
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
	case "doctor":
		return runDoctor(argv[1:], stdout, stderr, env)
	case "sync-project-kas":
		return runSyncProjectKAS(argv[1:], stdout, stderr, env)
	case "install-project-kas":
		return runInstallProjectKAS(argv[1:], stdout, stderr, env)
	default:
		return emitError(stderr, "unknown_command", "only the list, install, doctor, sync-project-kas, and install-project-kas commands are implemented", argv[0], false, "")
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
	kabStage := fs.String("kab-stage", "", "KAB adoption stage numeric selector (1 or 2)")
	kabAdoptionStage := fs.String("kab-adoption-stage", "", "KAB adoption stage canonical selector")
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
	if !*dryRun {
		if *approve == "" {
			return emitError(stderr, "install_requires_dry_run_or_approve", "install requires --dry-run or --approve dry-run:<hash>.", "install", *jsonOutput, "Rerun with install --profile <profile> <pack-id>... --dry-run.")
		}
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "install requires --profile <profile>.", "install", *jsonOutput, "")
	}
	packIDs := fs.Args()
	if len(packIDs) == 0 {
		return emitError(stderr, "pack_id_required", "install requires at least one pack id.", "install", *jsonOutput, "")
	}
	stageInput, err := resolveInstallStageInput(*kabStage, *kabAdoptionStage, *jsonOutput, stdout, env)
	if err != nil {
		return emitError(stderr, "kab_adoption_stage_invalid", err.Error(), "install", *jsonOutput, "")
	}

	var result install.Result
	if *approve != "" {
		result, err = install.ApplyApprovedInstall(*repo, install.Options{Profile: *profile, PackIDs: packIDs, ProfileRoot: *profileRoot, KABStageSelection: stageInput}, *approve)
	} else {
		result, err = install.BuildDryRun(*repo, install.Options{Profile: *profile, PackIDs: packIDs, ProfileRoot: *profileRoot, KABStageSelection: stageInput})
	}
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
	} else if result.Mode == "approved_copy" {
		fmt.Fprintln(out, install.RenderHumanApproved(result))
	} else {
		fmt.Fprintln(out, install.RenderHumanDryRun(result))
	}
	return code
}

func runDoctor(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", "", "source KAS repo path")
	profile := fs.String("profile", "", "Hermes target profile name")
	project := fs.String("project", "", "KAH project path to inspect")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
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
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "doctor", *jsonOutput, "")
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "doctor requires --profile <profile>.", "doctor", *jsonOutput, "")
	}
	result, err := doctor.Build(*repo, doctor.Options{Profile: *profile, Project: *project, ProfileRoot: *profileRoot})
	if err != nil {
		return emitError(stderr, "doctor_failed", err.Error(), "doctor", *jsonOutput, "")
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
		fmt.Fprintln(out, doctor.RenderHuman(result))
	}
	return code
}

func runSyncProjectKAS(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("sync-project-kas", flag.ContinueOnError)
	fs.SetOutput(stderr)
	profile := fs.String("profile", "", "Hermes target profile name")
	project := fs.String("project", "", "project-specific KAS id")
	statePath := fs.String("state", "", "project kas-project-state.yaml path")
	legacyMarkerPath := fs.String("legacy-marker", "", "optional legacy kab-adoption-stage.md path")
	repoPath := fs.String("repo", "", "current upstream KAS source repo path")
	projectRoot := fs.String("project-root", "", "project-specific KAS root path")
	dryRun := fs.Bool("dry-run", false, "validate state without writing")
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
	if !*dryRun {
		return emitError(stderr, "sync_project_kas_requires_dry_run", "sync-project-kas is read-only for KASUPD-003 and requires --dry-run.", "sync-project-kas", *jsonOutput, "Rerun with sync-project-kas --profile <profile> --project <project-id> --state <path> --dry-run.")
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "sync-project-kas requires --profile <profile>.", "sync-project-kas", *jsonOutput, "")
	}
	if *project == "" {
		return emitError(stderr, "project_required", "sync-project-kas requires --project <project-id>.", "sync-project-kas", *jsonOutput, "")
	}
	result := kasstate.Build(kasstate.Options{Profile: *profile, Project: *project, StatePath: *statePath, LegacyMarkerPath: *legacyMarkerPath, DryRun: *dryRun, RepoPath: *repoPath, ProjectRoot: *projectRoot})
	out := stdout
	code := 0
	if !result.OK {
		out = stderr
		code = 2
	}
	if *jsonOutput {
		_ = writeJSON(out, result)
	} else {
		fmt.Fprintln(out, kasstate.RenderHuman(result))
	}
	return code
}

func runInstallProjectKAS(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("install-project-kas", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", "", "source KAS repo path")
	profile := fs.String("profile", "", "Hermes target profile name")
	project := fs.String("project", "", "project-specific KAS id")
	sourcePack := fs.String("source-pack", "", "project source suite id")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
	dryRun := fs.Bool("dry-run", false, "render project-specific install plan without writing")
	fs.String("approve", "", "unsupported until KASPROJ-003")
	jsonOutput := fs.Bool("json", false, "emit machine-readable JSON")
	fs.Bool("no-color", false, "accepted for stable CLI shape; output is uncolored")
	if hasHelpArg(argv) {
		fs.SetOutput(stdout)
		fs.Usage()
		return 0
	}
	if hasProjectInstallApproveFlag(argv) {
		return emitError(stderr, "project_install_approve_unsupported", "install-project-kas --approve belongs to KASPROJ-003 and is not supported by this dry-run planner.", "install-project-kas", wantsJSON(argv), "Rerun with --dry-run and review the plan hash; approved install remains KASPROJ-003.")
	}
	if writeFlag := unsupportedProjectInstallWriteFlag(argv); writeFlag != "" {
		return emitError(stderr, "project_install_write_form_unsupported", "install-project-kas is dry-run only for KASPROJ-002; unsupported write/approval flag: "+writeFlag, "install-project-kas", wantsJSON(argv), "Rerun with install-project-kas --profile <profile> --project <project> --source-pack kas-default-project-suite --dry-run.")
	}
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	if *profileRoot != "" && envValue(env, "KAS_ALLOW_PROFILE_ROOT_OVERRIDE") != "1" {
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "install-project-kas", *jsonOutput, "")
	}
	if !*dryRun {
		return emitError(stderr, "project_install_requires_dry_run", "install-project-kas requires --dry-run for KASPROJ-002.", "install-project-kas", *jsonOutput, "Rerun with install-project-kas --profile <profile> --project <project> --source-pack kas-default-project-suite --dry-run.")
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "install-project-kas requires --profile <profile>.", "install-project-kas", *jsonOutput, "")
	}
	if *project == "" {
		return emitError(stderr, "project_required", "install-project-kas requires --project <project>.", "install-project-kas", *jsonOutput, "")
	}
	if *sourcePack == "" {
		return emitError(stderr, "source_pack_required", "install-project-kas requires --source-pack <source_pack>.", "install-project-kas", *jsonOutput, "")
	}
	result, err := projectinstall.BuildDryRun(*repo, projectinstall.Options{Profile: *profile, Project: *project, SourcePack: *sourcePack, ProfileRoot: *profileRoot, DryRun: *dryRun})
	if err != nil {
		return emitError(stderr, "project_install_planner_failed", err.Error(), "install-project-kas", *jsonOutput, "")
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
		fmt.Fprintln(out, projectinstall.RenderHumanDryRun(result))
	}
	return code
}

func normalizeInstallArgs(argv []string) []string {
	rewritten := []string{}
	positionals := []string{}
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch arg {
		case "--repo", "--profile", "--profile-root", "--approve", "--kab-stage", "--kab-adoption-stage":
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
	for _, name := range []string{"--repo=", "--profile=", "--profile-root=", "--approve=", "--kab-stage=", "--kab-adoption-stage=", "--dry-run=", "--json=", "--no-color="} {
		if strings.HasPrefix(arg, name) {
			return true
		}
	}
	return false
}

func resolveInstallStageInput(numeric string, canonical string, jsonOutput bool, stdout io.Writer, env map[string]string) (install.StageSelectionInput, error) {
	input := install.StageSelectionInput{Numeric: numeric, Canonical: canonical}
	if numeric != "" || canonical != "" {
		if _, err := install.ResolveKABAdoptionStage(input); err != nil {
			return input, err
		}
		return input, nil
	}
	if jsonOutput || envValue(env, "CI") != "" || !installPromptInteractive() {
		input.Numeric = "1"
		input.Source = "default_stage1"
		return input, nil
	}
	fmt.Fprint(stdout, "KAB adoption stage for this KAS/KAH project pack:\n  [1] Stage 1 — direct Codex app-server baseline (default)\n  [2] Stage 2 — KAB Codex-first via native_codex\nChoice [1]: ")
	line, err := bufio.NewReader(installPromptInput).ReadString('\n')
	if err != nil && len(line) == 0 {
		input.Numeric = "1"
		input.Source = "default_stage1"
		return input, nil
	}
	choice := strings.TrimSpace(line)
	if choice == "" {
		choice = "1"
	}
	input.Numeric = choice
	input.Source = "interactive"
	if _, err := install.ResolveKABAdoptionStage(input); err != nil {
		return input, err
	}
	return input, nil
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
	fmt.Fprintln(w, "  doctor   Verify a profile-scoped KAS skill-pack install")
	fmt.Fprintln(w, "  sync-project-kas  Validate project-specific KAS state without writing")
	fmt.Fprintln(w, "  install-project-kas  Plan a project-specific KAS suite install without writing")
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

func unsupportedProjectInstallWriteFlag(argv []string) string {
	for _, arg := range argv {
		name := arg
		if before, _, ok := strings.Cut(arg, "="); ok {
			name = before
		}
		switch name {
		case "--write", "--force", "--yes", "--repair", "--migrate", "--from-generic":
			return name
		}
	}
	return ""
}

func hasProjectInstallApproveFlag(argv []string) bool {
	for _, arg := range argv {
		name := arg
		if before, _, ok := strings.Cut(arg, "="); ok {
			name = before
		}
		if name == "--approve" {
			return true
		}
	}
	return false
}

func wantsJSON(argv []string) bool {
	for _, arg := range argv {
		if arg == "--json" || arg == "--json=true" {
			return true
		}
	}
	return false
}
