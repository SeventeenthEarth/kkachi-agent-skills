package cli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/doctor"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/graphsync"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/install"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/kasstate"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/projectinstall"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/toolchain"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowcreator"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowpromoter"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowrouting"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/workflowtrigger"
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
		return emitError(stderr, "command_required", "expected list, install, doctor, repair, toolchain, workflow-create, workflow-promote, workflow-route, workflow-trigger, uninstall, or version command", "", false, "")
	}
	if isVersionArg(argv[0]) {
		return runVersion(argv[1:], stdout, stderr)
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
	case "repair":
		return runRepair(argv[1:], stdout, stderr, env)
	case "toolchain":
		return runToolchain(argv[1:], stdout, stderr, env)
	case "workflow-create":
		return runWorkflowCreate(argv[1:], stdout, stderr, env)
	case "workflow-promote":
		return runWorkflowPromote(argv[1:], stdout, stderr, env)
	case "workflow-route":
		return runWorkflowRoute(argv[1:], stdout, stderr, env)
	case "workflow-trigger":
		return runWorkflowTrigger(argv[1:], stdout, stderr, env)
	case "uninstall":
		return runUninstall(argv[1:], stdout, stderr, env)
	case "version":
		return runVersion(argv[1:], stdout, stderr)
	case "sync-project-kas":
		return runSyncProjectKAS(argv[1:], stdout, stderr, env)
	case "install-project-kas":
		return runInstallProjectKAS(argv[1:], stdout, stderr, env)
	case "repair-project-kas":
		return runRepairProjectKAS(argv[1:], stdout, stderr, env)
	default:
		return emitError(stderr, "unknown_command", "only the list, install, doctor, repair, toolchain, workflow-create, workflow-promote, workflow-route, workflow-trigger, uninstall, and version commands are routine public lifecycle commands", argv[0], false, "")
	}
}

func runVersion(argv []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(stderr)
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "version does not accept positional arguments", "version", *jsonOutput, "Rerun as kkachi-agent-skills version or kkachi-agent-skills version --json.")
	}
	if *jsonOutput {
		_ = writeJSON(stdout, version.Current())
	} else {
		fmt.Fprintln(stdout, version.Human())
	}
	return 0
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
	project := fs.String("project", "", "project-specific KAS id")
	sourcePack := fs.String("source-pack", projectinstall.VirtualSourcePackID, "project source suite id")
	suiteRole := fs.String("suite-role", "", "explicit role-aware project suite role")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
	fromGeneric := fs.Bool("from-generic", false, "plan explicit generic-to-project migration")
	dryRun := fs.Bool("dry-run", false, "report planned changes without writing")
	approve := fs.String("approve", "", "approval evidence ref for future approved copy install")
	apply := fs.String("apply", "", "approval evidence ref dry-run:sha256:<hash> for TOKEN-005 approved lifecycle writes")
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
	if *project != "" || *fromGeneric {
		return runPublicProjectInstall(*repo, *profile, *project, *sourcePack, *suiteRole, *profileRoot, *dryRun, *fromGeneric, *approve, *apply, *jsonOutput, stdout, stderr, env, argv)
	}
	if *apply != "" {
		if *approve != "" || *dryRun {
			return emitError(stderr, "install_mode_ambiguous", "install accepts only one of --dry-run, --approve, or --apply.", "install", *jsonOutput, "Run dry-run first, then rerun with one approval token.")
		}
		*approve = *apply
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

func runPublicProjectInstall(repo string, profile string, project string, sourcePack string, suiteRole string, profileRoot string, dryRun bool, fromGeneric bool, approve string, apply string, jsonOutput bool, stdout io.Writer, stderr io.Writer, env map[string]string, argv []string) int {
	if profileRoot != "" && envValue(env, "KAS_ALLOW_PROFILE_ROOT_OVERRIDE") != "1" {
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "install", jsonOutput, "")
	}
	if apply != "" && approve != "" {
		return emitError(stderr, "project_install_mode_ambiguous", "install --project accepts only one approval flag.", "install", jsonOutput, "Run dry-run first, then rerun with --apply dry-run:sha256:<hash>.")
	}
	approval := apply
	if approval == "" {
		approval = approve
	}
	if dryRun && approval != "" {
		return emitError(stderr, "project_install_mode_ambiguous", "install --project accepts either --dry-run or --apply, not both.", "install", jsonOutput, "Run dry-run first, then rerun with only --apply dry-run:sha256:<hash>.")
	}
	if !dryRun && approval == "" {
		return emitError(stderr, "project_install_requires_dry_run_or_apply", "install --project requires --dry-run or --apply dry-run:sha256:<hash>.", "install", jsonOutput, "Rerun with install --profile <profile> --project <project> --suite-role <role> --dry-run.")
	}
	if profile == "" {
		return emitError(stderr, "profile_required", "install --project requires --profile <profile>.", "install", jsonOutput, "")
	}
	if project == "" {
		return emitError(stderr, "project_required", "install --project or install --from-generic requires --project <project>.", "install", jsonOutput, "")
	}
	if sourcePack == "" {
		sourcePack = projectinstall.VirtualSourcePackID
	}
	opts := projectinstall.ProjectSuiteOptions{Profile: profile, Project: project, SourcePack: sourcePack, SourcePackExplicit: hasFlag(argv, "--source-pack"), ProfileRoot: profileRoot, FromGeneric: fromGeneric}
	if fromGeneric {
		var result projectinstall.ProjectActionResult
		var err error
		if approval != "" {
			result, err = projectinstall.ApplyApprovedMigration(repo, opts, approval)
		} else {
			result, err = projectinstall.BuildProjectMigrationDryRun(repo, opts)
		}
		if err != nil {
			return emitError(stderr, "project_install_planner_failed", err.Error(), "install", jsonOutput, "")
		}
		result.Command = "install"
		return emitResult(stdout, stderr, result.OK, jsonOutput, result, func() string {
			return projectinstall.RenderHumanProjectAction(result)
		})
	}
	var result projectinstall.Result
	var err error
	if approval != "" {
		result, err = projectinstall.ApplyApprovedInstall(repo, projectinstall.Options{Profile: profile, Project: project, SuiteRole: suiteRole, SourcePack: sourcePack, ProfileRoot: profileRoot, DryRun: true}, approval)
	} else {
		result, err = projectinstall.BuildDryRun(repo, projectinstall.Options{Profile: profile, Project: project, SuiteRole: suiteRole, SourcePack: sourcePack, ProfileRoot: profileRoot, DryRun: true})
	}
	if err != nil {
		return emitError(stderr, "project_install_planner_failed", err.Error(), "install", jsonOutput, "")
	}
	result.Command = "install"
	if result.DryRun && result.OK {
		result.NextAction = "Review install --project dry-run evidence; no files were written. Apply with install --profile " + profile + " --project " + project + " --suite-role " + suiteRole + " --apply " + result.ApprovalRequest.EvidenceRef + "; use doctor --project-suite after approved writes."
	}
	return emitResult(stdout, stderr, result.OK, jsonOutput, result, func() string {
		if result.DryRun {
			return projectinstall.RenderHumanDryRun(result)
		}
		return projectinstall.RenderHumanApproved(result)
	})
}

func runDoctor(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", "", "source KAS repo path")
	profile := fs.String("profile", "", "Hermes target profile name")
	project := fs.String("project", "", "KAH project path to inspect; project suite id when --project-suite is present")
	projectSuite := fs.Bool("project-suite", false, "interpret --project as a project-specific KAS suite id")
	workflowGraph := fs.Bool("workflow-graph", false, "inspect project .kkachi-workflow.yaml supportability without writing")
	pluginDoctor := fs.Bool("plugin", false, "inspect SKILL plugin/wrapper/overlay diagnostics without writing")
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
	if *pluginDoctor {
		if *workflowGraph || *projectSuite {
			return emitError(stderr, "doctor_mode_ambiguous", "doctor --plugin cannot be combined with --workflow-graph or --project-suite.", "doctor", *jsonOutput, "Rerun with exactly one doctor mode: --plugin, --workflow-graph, --project-suite, or default profile doctor.")
		}
		if *repo == "" {
			return emitError(stderr, "repo_required_for_plugin_doctor", "doctor --plugin requires --repo <kas-repo> so read-only source evidence is explicit and no embedded source cache materialization is needed.", "doctor", *jsonOutput, "Rerun with doctor --plugin --repo <kas-repo> --json.")
		}
		result, err := doctor.BuildSkill(*repo, doctor.SkillOptions{Profile: *profile, Project: *project, ProfileRoot: *profileRoot})
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
			fmt.Fprintln(out, doctor.RenderHumanSkill(result))
		}
		return code
	}
	if *workflowGraph {
		if *projectSuite {
			return emitError(stderr, "doctor_mode_ambiguous", "doctor --workflow-graph cannot be combined with --project-suite.", "doctor", *jsonOutput, "Rerun with either doctor --project <path> --workflow-graph --json or doctor --project <project> --project-suite.")
		}
		if *project == "" {
			return emitError(stderr, "project_required", "doctor --workflow-graph requires --project <path>.", "doctor", *jsonOutput, "")
		}
		result, err := doctor.BuildWorkflowGraph(*repo, doctor.WorkflowGraphOptions{Project: *project})
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
			fmt.Fprintln(out, doctor.RenderHumanWorkflowGraph(result))
		}
		return code
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "doctor requires --profile <profile>.", "doctor", *jsonOutput, "")
	}
	if *projectSuite {
		if *project == "" {
			return emitError(stderr, "project_required", "doctor --project-suite requires --project <project>.", "doctor", *jsonOutput, "")
		}
		result, err := projectinstall.BuildProjectSuiteDoctor(*repo, projectinstall.ProjectSuiteOptions{Profile: *profile, Project: *project, ProfileRoot: *profileRoot})
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
			fmt.Fprintln(out, projectinstall.RenderHumanProjectSuiteDoctor(result))
		}
		return code
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

func runRepair(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("repair", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", "", "source KAS repo path")
	profile := fs.String("profile", "", "Hermes target profile name")
	project := fs.String("project", "", "project-specific KAS id")
	sourcePack := fs.String("source-pack", projectinstall.VirtualSourcePackID, "project source suite id")
	suiteRole := fs.String("suite-role", "", "explicit role-aware project suite role")
	pruneExtra := fs.Bool("prune-extra", false, "prune manifest-tracked KAS-managed skills outside --suite-role")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
	workflowGraph := fs.Bool("workflow-graph", false, "orchestrate workflow graph proposal/apply through KAH")
	propose := fs.Bool("propose", false, "create a KAH workflow graph proposal without applying it")
	reason := fs.String("reason", "", "proposal reason for workflow graph repair")
	applyProposal := fs.String("apply-proposal", "", "KAH workflow graph proposal id to apply")
	graphApproval := fs.String("approval", "", "approval evidence ref for workflow graph apply")
	dryRun := fs.Bool("dry-run", false, "report planned repairs without writing")
	approve := fs.String("approve", "", "compatibility alias for --apply dry-run:sha256:<hash>")
	apply := fs.String("apply", "", "approval evidence ref dry-run:sha256:<hash> for approved lifecycle writes")
	backupVaultRoot := fs.String("backup-vault-root", "", "required absolute backup vault root for repair --apply")
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
	if *workflowGraph {
		return runWorkflowGraphRepair(*repo, *project, *propose, *reason, *applyProposal, *graphApproval, *dryRun, *approve, *apply, *profile, *sourcePack, *profileRoot, hasFlag(argv, "--source-pack"), hasFlag(argv, "--profile-root"), *jsonOutput, stdout, stderr)
	}
	if *profileRoot != "" && envValue(env, "KAS_ALLOW_PROFILE_ROOT_OVERRIDE") != "1" {
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "repair", *jsonOutput, "")
	}
	if *apply != "" && *approve != "" {
		return emitError(stderr, "project_repair_mode_ambiguous", "repair accepts only one approval flag.", "repair", *jsonOutput, "Run dry-run first, then rerun with --apply dry-run:sha256:<hash>.")
	}
	approval := *apply
	if approval == "" {
		approval = *approve
	}
	if *dryRun && approval != "" {
		return emitError(stderr, "project_repair_mode_ambiguous", "repair accepts either --dry-run or --apply, not both.", "repair", *jsonOutput, "Run dry-run first, then rerun with only --apply dry-run:sha256:<hash>.")
	}
	if !*dryRun && approval == "" {
		return emitError(stderr, "project_repair_requires_dry_run_or_apply", "repair requires --dry-run or --apply dry-run:sha256:<hash>.", "repair", *jsonOutput, "Rerun with repair --profile <profile> --project <project> --dry-run.")
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "repair requires --profile <profile>.", "repair", *jsonOutput, "")
	}
	if *project == "" {
		return emitError(stderr, "project_required", "repair requires --project <project>.", "repair", *jsonOutput, "")
	}
	opts := projectinstall.ProjectSuiteOptions{Profile: *profile, Project: *project, SuiteRole: *suiteRole, PruneExtra: *pruneExtra, BackupVaultRoot: *backupVaultRoot, SourcePack: *sourcePack, SourcePackExplicit: hasFlag(argv, "--source-pack"), ProfileRoot: *profileRoot}
	var result projectinstall.ProjectActionResult
	var err error
	if approval != "" {
		result, err = projectinstall.ApplyApprovedRepair(*repo, opts, approval)
	} else {
		result, err = projectinstall.BuildProjectRepairDryRun(*repo, opts)
	}
	if err != nil {
		return emitError(stderr, "project_repair_failed", err.Error(), "repair", *jsonOutput, "")
	}
	result.Command = "repair"
	result.NextAction = publicRepairNextAction(result, approval != "")
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return projectinstall.RenderHumanProjectAction(result)
	})
}

func runToolchain(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	if len(argv) == 0 || hasHelpArg(argv[:1]) {
		printToolchainHelp(stdout)
		return 0
	}
	switch argv[0] {
	case "init":
		return runToolchainAction("init", argv[1:], stdout, stderr)
	case "doctor":
		return runToolchainAction("doctor", argv[1:], stdout, stderr)
	case "refresh":
		return runToolchainAction("refresh", argv[1:], stdout, stderr)
	case "import-legacy":
		return runToolchainImportLegacy(argv[1:], stdout, stderr)
	case "set-stage":
		return runToolchainSetStage(argv[1:], stdout, stderr)
	case "install-launchers":
		return runToolchainInstallLaunchers(argv[1:], stdout, stderr)
	default:
		return emitError(stderr, "unknown_toolchain_command", "toolchain supports init, doctor, refresh, import-legacy, set-stage, and install-launchers.", "toolchain", wantsJSON(argv), "")
	}
}

func runToolchainAction(action string, argv []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("toolchain "+action, flag.ContinueOnError)
	fs.SetOutput(stderr)
	projectRoot := fs.String("project-root", "", "project root for .kkachi/toolchain.yaml")
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "toolchain "+action+" does not accept positional arguments.", "toolchain "+action, *jsonOutput, "")
	}
	if *projectRoot == "" {
		return emitError(stderr, "project_root_required", "toolchain "+action+" requires --project-root <path>.", "toolchain "+action, *jsonOutput, "")
	}
	opts := toolchain.Options{ProjectRoot: *projectRoot}
	var result toolchain.Result
	switch action {
	case "init":
		result = toolchain.Init(opts)
	case "doctor":
		result = toolchain.Doctor(opts)
	case "refresh":
		result = toolchain.Refresh(opts)
	}
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return toolchain.RenderHuman(result)
	})
}

func runToolchainImportLegacy(argv []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("toolchain import-legacy", flag.ContinueOnError)
	fs.SetOutput(stderr)
	projectRoot := fs.String("project-root", "", "project root for .kkachi/toolchain.yaml")
	profile := fs.String("profile", "", "Hermes target profile name")
	project := fs.String("project", "", "project-specific KAS id")
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "toolchain import-legacy does not accept positional arguments.", "toolchain import-legacy", *jsonOutput, "")
	}
	if *projectRoot == "" {
		return emitError(stderr, "project_root_required", "toolchain import-legacy requires --project-root <path>.", "toolchain import-legacy", *jsonOutput, "")
	}
	result := toolchain.ImportLegacy(toolchain.Options{ProjectRoot: *projectRoot, Profile: *profile, Project: *project})
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return toolchain.RenderHuman(result)
	})
}

func runToolchainSetStage(argv []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("toolchain set-stage", flag.ContinueOnError)
	fs.SetOutput(stderr)
	projectRoot := fs.String("project-root", "", "project root for .kkachi/toolchain.yaml")
	stage := fs.String("stage", "", "KAB adoption stage numeric or canonical selector")
	approvalEvidence := fs.String("approval-evidence", "", "approval evidence reference for the stage selection")
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "toolchain set-stage does not accept positional arguments.", "toolchain set-stage", *jsonOutput, "")
	}
	if *projectRoot == "" {
		return emitError(stderr, "project_root_required", "toolchain set-stage requires --project-root <path>.", "toolchain set-stage", *jsonOutput, "")
	}
	result := toolchain.SetStage(toolchain.Options{ProjectRoot: *projectRoot, Stage: *stage, ApprovalEvidence: *approvalEvidence})
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return toolchain.RenderHuman(result)
	})
}

func runToolchainInstallLaunchers(argv []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("toolchain install-launchers", flag.ContinueOnError)
	fs.SetOutput(stderr)
	binDir := fs.String("bin-dir", defaultLauncherBinDir(), "directory where kkachi-agent-*-toolchain launchers are installed")
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "toolchain install-launchers does not accept positional arguments.", "toolchain install-launchers", *jsonOutput, "")
	}
	result := toolchain.InstallLaunchers(toolchain.Options{LauncherBinDir: *binDir})
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return toolchain.RenderHuman(result)
	})
}

func publicRepairNextAction(result projectinstall.ProjectActionResult, approved bool) string {
	if approved || !result.ApprovalRequest.Required || result.ApprovalRequest.EvidenceRef == "" {
		return result.NextAction
	}
	rolePart := ""
	if strings.TrimSpace(result.SuiteRole) != "" {
		rolePart = " --suite-role " + result.SuiteRole
	}
	prunePart := ""
	if result.PruneExtra {
		prunePart = " --prune-extra"
	}
	return fmt.Sprintf("Review dry-run evidence and apply with repair --profile %s --project %s%s%s --apply %s --backup-vault-root <approved-abs-vault-root>; rerun doctor --project-suite after approved changes.", result.TargetProfile.Name, result.Project.ID, rolePart, prunePart, result.ApprovalRequest.EvidenceRef)
}

func runWorkflowGraphRepair(repo string, project string, propose bool, reason string, applyProposal string, approval string, dryRun bool, approve string, apply string, profile string, sourcePack string, profileRoot string, sourcePackExplicit bool, profileRootExplicit bool, jsonOutput bool, stdout io.Writer, stderr io.Writer) int {
	if dryRun || approve != "" || apply != "" || profile != "" || sourcePackExplicit || profileRootExplicit || profileRoot != "" {
		return emitError(stderr, "workflow_graph_repair_mode_ambiguous", "repair --workflow-graph cannot be combined with project-suite repair flags.", "repair", jsonOutput, "Rerun with either repair --workflow-graph --propose or the project-suite repair lifecycle.")
	}
	if project == "" {
		return emitError(stderr, "project_required", "repair --workflow-graph requires --project <project-path>.", "repair", jsonOutput, "")
	}
	if propose && applyProposal != "" {
		return emitError(stderr, "workflow_graph_repair_mode_ambiguous", "repair --workflow-graph accepts either --propose or --apply-proposal, not both.", "repair", jsonOutput, "")
	}
	if !propose && applyProposal == "" {
		return emitError(stderr, "workflow_graph_repair_mode_required", "repair --workflow-graph requires --propose or --apply-proposal <proposal-id>.", "repair", jsonOutput, "Run doctor first, then choose proposal or approved apply.")
	}
	if propose && strings.TrimSpace(reason) == "" {
		return emitError(stderr, "reason_required", "repair --workflow-graph --propose requires --reason <reason>.", "repair", jsonOutput, "")
	}
	if applyProposal != "" && strings.TrimSpace(approval) == "" {
		result, err := graphsync.Apply(graphsync.Options{Repo: repo, Project: project, Proposal: applyProposal, Approval: approval})
		if err != nil {
			return emitError(stderr, "workflow_graph_repair_failed", err.Error(), "repair", jsonOutput, "")
		}
		return emitResult(stdout, stderr, false, jsonOutput, result, func() string {
			return graphsync.RenderHuman(result)
		})
	}
	var result graphsync.Result
	var err error
	if propose {
		result, err = graphsync.Propose(graphsync.Options{Repo: repo, Project: project, Reason: reason})
	} else {
		result, err = graphsync.Apply(graphsync.Options{Repo: repo, Project: project, Proposal: applyProposal, Approval: approval})
	}
	if err != nil {
		return emitError(stderr, "workflow_graph_repair_failed", err.Error(), "repair", jsonOutput, "")
	}
	return emitResult(stdout, stderr, result.OK, jsonOutput, result, func() string {
		return graphsync.RenderHuman(result)
	})
}

func runWorkflowTrigger(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("workflow-trigger", flag.ContinueOnError)
	fs.SetOutput(stderr)
	project := fs.String("project", "", "project path")
	workflowID := fs.String("workflow-id", "", "explicit workflow id")
	workflowFile := fs.String("workflow-file", "", "explicit repo-relative workflow YAML file")
	nodeContractSource := fs.String("node-contract-source", "", "explicit JSON node-contract source path")
	nodeContractRef := fs.String("node-contract-ref", "", "optional node-contract source ref")
	selectorRegistry := fs.String("selector-registry", "", "selector/node-contract registry path")
	routeResult := fs.String("route-result", "", "workflow-route JSON result to materialize under the run")
	customWorkflowPacket := fs.String("custom-workflow-packet", "", "approved workflow-create dry-run JSON packet to materialize under the run")
	approval := fs.String("approval", "", "approval evidence dry-run:sha256:<hash> for custom workflow packet materialization")
	materializeRunLocal := fs.Bool("materialize-run-local", false, "materialize selected route result or approved custom packet under .kkachi/runs/<run-id>/workflow")
	workflowManaged := fs.Bool("workflow-managed", false, "require classified workflow-route and run-local materialization/resume evidence before dispatch")
	taskClass := fs.String("task-class", "", "selector task class")
	labels := fs.String("labels", "", "selector labels, comma-separated")
	changedSurfaces := fs.String("changed-surfaces", "", "selector changed surfaces, comma-separated")
	risk := fs.String("risk", "", "selector risk level")
	requiredAgent := fs.String("required-agent", "", "selector required agents, comma-separated")
	requiredCapability := fs.String("required-capability", "", "selector required capabilities, comma-separated")
	runID := fs.String("run", "", "KAH run id for workflow create")
	instanceID := fs.String("instance-id", "", "existing KAH workflow instance id to resume")
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "workflow-trigger does not accept positional arguments", "workflow-trigger", *jsonOutput, "Rerun with workflow-trigger --project <path> --workflow-id <id> --node-contract-source <path> --run <run-id> --json.")
	}
	result, err := workflowtrigger.Trigger(workflowtrigger.Options{
		Project:              *project,
		WorkflowID:           *workflowID,
		WorkflowFile:         *workflowFile,
		NodeContractSource:   *nodeContractSource,
		NodeContractRef:      *nodeContractRef,
		SelectorRegistry:     *selectorRegistry,
		RouteResult:          *routeResult,
		CustomWorkflowPacket: *customWorkflowPacket,
		Approval:             *approval,
		MaterializeRunLocal:  *materializeRunLocal,
		WorkflowManaged:      *workflowManaged,
		TaskClass:            *taskClass,
		Labels:               splitFlagList(*labels),
		ChangedSurfaces:      splitFlagList(*changedSurfaces),
		Risk:                 *risk,
		RequiredAgents:       splitFlagList(*requiredAgent),
		RequiredCapabilities: splitFlagList(*requiredCapability),
		RunID:                *runID,
		InstanceID:           *instanceID,
	})
	if err != nil {
		return emitError(stderr, "workflow_trigger_failed", err.Error(), "workflow-trigger", *jsonOutput, "")
	}
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return workflowtrigger.RenderHuman(result)
	})
}

func runWorkflowRoute(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("workflow-route", flag.ContinueOnError)
	fs.SetOutput(stderr)
	taxonomy := fs.String("taxonomy", "", "task taxonomy path")
	selectorRegistry := fs.String("selector-registry", "", "standard bundle selector registry path")
	taskClass := fs.String("task-class", "", "already-classified task class")
	classificationReason := fs.String("classification-reason", "", "classification reason from the task contract or classifier artifact")
	selectedSpine := fs.String("selected-spine", "", "optional expected selected bundle/spine")
	projectHasTealLane := fs.String("project-has-teal-lane", "", "explicit Teal lane fact: true or false")
	uiUXChange := fs.String("ui-ux-change", "", "explicit UI/UX change fact: true or false")
	tealSkipReason := fs.String("teal-skip-reason", "", "concrete skip reason when Teal is not required")
	tealWaiverApproved := fs.Bool("teal-waiver-approved", false, "record bounded Teal waiver approval expectation")
	tealWaiverApprovalRef := fs.String("teal-waiver-approval-ref", "", "bounded Teal waiver approval reference")
	tealWaiverScope := fs.String("teal-waiver-scope", "", "bounded Teal waiver scope")
	tealWaiverExpiresAt := fs.String("teal-waiver-expires-at", "", "bounded Teal waiver expiry")
	labels := fs.String("labels", "", "selector labels, comma-separated")
	changedSurfaces := fs.String("changed-surfaces", "", "selector changed surfaces, comma-separated")
	risk := fs.String("risk", "", "selector risk level")
	requiredAgent := fs.String("required-agent", "", "selector required agents, comma-separated")
	requiredCapability := fs.String("required-capability", "", "selector required capabilities, comma-separated")
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "workflow-route does not accept positional arguments", "workflow-route", *jsonOutput, "Rerun with workflow-route --taxonomy <path> --selector-registry <path> --task-class <class> --classification-reason <reason> --json.")
	}
	projectHasTealLaneValue, ok := parseOptionalBoolFlag(*projectHasTealLane)
	if !ok {
		return emitError(stderr, "teal_applicability_invalid", "--project-has-teal-lane must be true or false when supplied.", "workflow-route", *jsonOutput, "")
	}
	uiUXChangeValue, ok := parseOptionalBoolFlag(*uiUXChange)
	if !ok {
		return emitError(stderr, "teal_applicability_invalid", "--ui-ux-change must be true or false when supplied.", "workflow-route", *jsonOutput, "")
	}
	result, err := workflowrouting.Route(workflowrouting.Options{
		TaxonomyPath:          *taxonomy,
		SelectorRegistryPath:  *selectorRegistry,
		TaskClass:             *taskClass,
		ClassificationReason:  *classificationReason,
		SelectedSpine:         *selectedSpine,
		ProjectHasTealLane:    projectHasTealLaneValue,
		UIUXChange:            uiUXChangeValue,
		TealSkipReason:        *tealSkipReason,
		TealWaiverApproved:    *tealWaiverApproved,
		TealWaiverApprovalRef: *tealWaiverApprovalRef,
		TealWaiverScope:       *tealWaiverScope,
		TealWaiverExpiresAt:   *tealWaiverExpiresAt,
		Labels:                splitFlagList(*labels),
		ChangedSurfaces:       splitFlagList(*changedSurfaces),
		Risk:                  *risk,
		RequiredAgents:        splitFlagList(*requiredAgent),
		RequiredCapabilities:  splitFlagList(*requiredCapability),
	})
	if err != nil {
		return emitError(stderr, "workflow_route_failed", err.Error(), "workflow-route", *jsonOutput, "")
	}
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return workflowrouting.RenderHuman(result)
	})
}

func parseOptionalBoolFlag(value string) (*bool, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, true
	}
	switch value {
	case "true":
		v := true
		return &v, true
	case "false":
		v := false
		return &v, true
	default:
		return nil, false
	}
}

func runWorkflowCreate(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("workflow-create", flag.ContinueOnError)
	fs.SetOutput(stderr)
	project := fs.String("project", "", "project path")
	workflowID := fs.String("workflow-id", "", "workflow id")
	mode := fs.String("mode", "", "creator mode: dag_only, thin_trigger, or full_trigger")
	request := fs.String("request", "", "workflow-create request JSON path")
	fullTriggerReason := fs.String("full-trigger-reason", "", "required reason for full_trigger mode")
	dryRun := fs.Bool("dry-run", false, "emit workflow-create packet without writing")
	apply := fs.String("apply", "", "approval evidence ref dry-run:sha256:<hash>")
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "workflow-create does not accept positional arguments", "workflow-create", *jsonOutput, "Rerun with workflow-create --project <path> --workflow-id <id> --mode dag_only --request <json-path> --dry-run --json.")
	}
	if *dryRun && *apply != "" {
		return emitError(stderr, "workflow_create_mode_ambiguous", "workflow-create accepts either --dry-run or --apply, not both.", "workflow-create", *jsonOutput, "Run dry-run first, then rerun with only --apply dry-run:sha256:<hash>.")
	}
	if !*dryRun && *apply == "" {
		return emitError(stderr, "workflow_create_requires_dry_run_or_apply", "workflow-create requires --dry-run or --apply dry-run:sha256:<hash>.", "workflow-create", *jsonOutput, "Rerun with workflow-create --project <path> --workflow-id <id> --mode dag_only --request <json-path> --dry-run.")
	}
	opts := workflowcreator.Options{Project: *project, WorkflowID: *workflowID, Mode: *mode, RequestPath: *request, FullTriggerReason: *fullTriggerReason, Approval: *apply}
	var result workflowcreator.Result
	var err error
	if *apply != "" {
		result, err = workflowcreator.Apply(opts)
	} else {
		result, err = workflowcreator.BuildDryRun(opts)
	}
	if err != nil {
		return emitError(stderr, "workflow_create_failed", err.Error(), "workflow-create", *jsonOutput, "")
	}
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return workflowcreator.RenderHuman(result)
	})
}

func runWorkflowPromote(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("workflow-promote", flag.ContinueOnError)
	fs.SetOutput(stderr)
	project := fs.String("project", "", "project path")
	runID := fs.String("run", "", "source WFLOW-008 run id")
	materialization := fs.String("materialization", "", "optional explicit materialization.json path")
	targetWorkflowID := fs.String("target-workflow-id", "", "target project-local workflow id")
	reuseReason := fs.String("reuse-reason", "", "required operator reason for promoting the run-local workflow")
	thinTrigger := fs.Bool("thin-trigger", false, "include an optional thin trigger candidate")
	dryRun := fs.Bool("dry-run", false, "emit workflow-promote proposal packet without writing")
	apply := fs.String("apply", "", "approval evidence ref dry-run:sha256:<hash>")
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
	if fs.NArg() != 0 {
		return emitError(stderr, "unexpected_argument", "workflow-promote does not accept positional arguments", "workflow-promote", *jsonOutput, "Rerun with workflow-promote --project <path> --run <run-id> --target-workflow-id <id> --reuse-reason <reason> --dry-run --json.")
	}
	if *dryRun && *apply != "" {
		return emitError(stderr, "workflow_promote_mode_ambiguous", "workflow-promote accepts either --dry-run or --apply, not both.", "workflow-promote", *jsonOutput, "Run dry-run first, then rerun with only --apply dry-run:sha256:<hash>.")
	}
	if !*dryRun && *apply == "" {
		return emitError(stderr, "workflow_promote_requires_dry_run_or_apply", "workflow-promote requires --dry-run or --apply dry-run:sha256:<hash>.", "workflow-promote", *jsonOutput, "Rerun with workflow-promote --project <path> --run <run-id> --target-workflow-id <id> --reuse-reason <reason> --dry-run.")
	}
	opts := workflowpromoter.Options{Project: *project, RunID: *runID, Materialization: *materialization, TargetWorkflowID: *targetWorkflowID, ReuseReason: *reuseReason, ThinTrigger: *thinTrigger, Approval: *apply}
	var result workflowpromoter.Result
	var err error
	if *apply != "" {
		result, err = workflowpromoter.Apply(opts)
	} else {
		result, err = workflowpromoter.BuildDryRun(opts)
	}
	if err != nil {
		return emitError(stderr, "workflow_promote_failed", err.Error(), "workflow-promote", *jsonOutput, "")
	}
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return workflowpromoter.RenderHuman(result)
	})
}

func runUninstall(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(stderr)
	profile := fs.String("profile", "", "Hermes target profile name")
	project := fs.String("project", "", "project-specific KAS id")
	sourcePack := fs.String("source-pack", projectinstall.VirtualSourcePackID, "project source suite id")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
	dryRun := fs.Bool("dry-run", false, "report planned removals without writing")
	apply := fs.String("apply", "", "approval evidence ref dry-run:sha256:<hash> for approved lifecycle writes")
	backupVaultRoot := fs.String("backup-vault-root", "", "required absolute Obsidian vault backup root for uninstall --apply")
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
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "uninstall", *jsonOutput, "")
	}
	if *dryRun && *apply != "" {
		return emitError(stderr, "project_uninstall_mode_ambiguous", "uninstall accepts either --dry-run or --apply, not both.", "uninstall", *jsonOutput, "Run dry-run first, then rerun with only --apply dry-run:sha256:<hash> --backup-vault-root <abs-path>.")
	}
	if !*dryRun && *apply == "" {
		return emitError(stderr, "project_uninstall_requires_dry_run_or_apply", "uninstall requires --dry-run or --apply dry-run:sha256:<hash>.", "uninstall", *jsonOutput, "Rerun with uninstall --profile <profile> --project <project> --dry-run.")
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "uninstall requires --profile <profile>.", "uninstall", *jsonOutput, "")
	}
	if *project == "" {
		return emitError(stderr, "project_required", "uninstall requires --project <project>.", "uninstall", *jsonOutput, "")
	}
	opts := projectinstall.ProjectSuiteOptions{Profile: *profile, Project: *project, SourcePack: *sourcePack, SourcePackExplicit: hasFlag(argv, "--source-pack"), ProfileRoot: *profileRoot}
	var result projectinstall.ProjectUninstallResult
	if *apply != "" {
		result, _ = projectinstall.ApplyProjectUninstall(opts, *apply, *backupVaultRoot)
	} else {
		result = projectinstall.BuildProjectUninstallDryRun(opts)
	}
	return emitResult(stdout, stderr, result.OK, *jsonOutput, result, func() string {
		return projectinstall.RenderHumanProjectUninstall(result)
	})
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
	suiteRole := fs.String("suite-role", "", "explicit role-aware project suite role")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
	dryRun := fs.Bool("dry-run", false, "render project-specific install plan without writing")
	approve := fs.String("approve", "", "approval evidence ref dry-run:<plan_hash>")
	jsonOutput := fs.Bool("json", false, "emit machine-readable JSON")
	fs.Bool("no-color", false, "accepted for stable CLI shape; output is uncolored")
	if hasHelpArg(argv) {
		fs.SetOutput(stdout)
		fs.Usage()
		return 0
	}
	if writeFlag := unsupportedProjectInstallWriteFlag(argv); writeFlag != "" {
		return emitError(stderr, "project_install_write_form_unsupported", "install-project-kas supports only --dry-run or --approve dry-run:<hash>; unsupported flag: "+writeFlag, "install-project-kas", wantsJSON(argv), "Rerun with install-project-kas --profile <profile> --project <project> --suite-role <role> --source-pack kas-default-project-suite --dry-run.")
	}
	if projectInstallApproveMissingValue(argv) {
		return emitError(stderr, "approval_evidence_malformed", "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>.", "install-project-kas", wantsJSON(argv), "Rerun with the dry-run JSON approval_request.evidence_ref value.")
	}
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	if *profileRoot != "" && envValue(env, "KAS_ALLOW_PROFILE_ROOT_OVERRIDE") != "1" {
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "install-project-kas", *jsonOutput, "")
	}
	if *dryRun && *approve != "" {
		return emitError(stderr, "project_install_mode_ambiguous", "install-project-kas accepts either --dry-run or --approve, not both.", "install-project-kas", *jsonOutput, "Run dry-run first, then rerun with only --approve dry-run:<hash>.")
	}
	if !*dryRun && *approve == "" {
		return emitError(stderr, "project_install_requires_dry_run_or_approve", "install-project-kas requires --dry-run or --approve dry-run:<hash>.", "install-project-kas", *jsonOutput, "Rerun with install-project-kas --profile <profile> --project <project> --suite-role <role> --source-pack kas-default-project-suite --dry-run.")
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
	var result projectinstall.Result
	var err error
	opts := projectinstall.Options{Profile: *profile, Project: *project, SuiteRole: *suiteRole, SourcePack: *sourcePack, ProfileRoot: *profileRoot, DryRun: true}
	if *approve != "" {
		result, err = projectinstall.ApplyApprovedInstall(*repo, opts, *approve)
	} else {
		result, err = projectinstall.BuildDryRun(*repo, opts)
	}
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
	} else if result.Mode == "project_approved_copy" {
		fmt.Fprintln(out, projectinstall.RenderHumanApproved(result))
	} else {
		fmt.Fprintln(out, projectinstall.RenderHumanDryRun(result))
	}
	return code
}

func runRepairProjectKAS(argv []string, stdout io.Writer, stderr io.Writer, env map[string]string) int {
	fs := flag.NewFlagSet("repair-project-kas", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", "", "source KAS repo path")
	profile := fs.String("profile", "", "Hermes target profile name")
	project := fs.String("project", "", "project-specific KAS id")
	sourcePack := fs.String("source-pack", projectinstall.VirtualSourcePackID, "project source suite id; defaults to kas-default-project-suite")
	suiteRole := fs.String("suite-role", "", "explicit role-aware project suite role")
	pruneExtra := fs.Bool("prune-extra", false, "prune manifest-tracked KAS-managed skills outside --suite-role")
	profileRoot := fs.String("profile-root", "", "test/harness-only explicit profile root")
	dryRun := fs.Bool("dry-run", false, "report planned repairs without writing")
	approve := fs.String("approve", "", "approval evidence ref dry-run:<plan_hash>")
	apply := fs.String("apply", "", "approval evidence ref dry-run:sha256:<hash> for approved lifecycle writes")
	backupVaultRoot := fs.String("backup-vault-root", "", "required absolute backup vault root for approved repair")
	jsonOutput := fs.Bool("json", false, "emit machine-readable JSON")
	fs.Bool("no-color", false, "accepted for stable CLI shape; output is uncolored")
	if hasHelpArg(argv) {
		fs.SetOutput(stdout)
		fs.Usage()
		return 0
	}
	if projectInstallApproveMissingValue(argv) {
		return emitError(stderr, "approval_evidence_malformed", "approval evidence must be exactly dry-run:sha256:<64 lowercase hex>.", "repair-project-kas", wantsJSON(argv), "Rerun with the dry-run JSON approval_request.evidence_ref value.")
	}
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	if *profileRoot != "" && envValue(env, "KAS_ALLOW_PROFILE_ROOT_OVERRIDE") != "1" {
		return emitError(stderr, "profile_root_override_rejected", "--profile-root is only allowed under an explicit test/harness guard.", "repair-project-kas", *jsonOutput, "")
	}
	if *apply != "" && *approve != "" {
		return emitError(stderr, "project_repair_mode_ambiguous", "repair-project-kas accepts only one approval flag.", "repair-project-kas", *jsonOutput, "Run dry-run first, then rerun with --approve or --apply dry-run:sha256:<hash>.")
	}
	approval := *apply
	if approval == "" {
		approval = *approve
	}
	if *dryRun && approval != "" {
		return emitError(stderr, "project_repair_mode_ambiguous", "repair-project-kas accepts either --dry-run or approval, not both.", "repair-project-kas", *jsonOutput, "Run dry-run first, then rerun with only --approve dry-run:<hash>.")
	}
	if !*dryRun && approval == "" {
		return emitError(stderr, "project_repair_requires_dry_run_or_approve", "repair-project-kas requires --dry-run or --approve dry-run:<hash>.", "repair-project-kas", *jsonOutput, "Rerun with repair-project-kas --profile <profile> --project <project> --dry-run.")
	}
	if *profile == "" {
		return emitError(stderr, "profile_required", "repair-project-kas requires --profile <profile>.", "repair-project-kas", *jsonOutput, "")
	}
	if *project == "" {
		return emitError(stderr, "project_required", "repair-project-kas requires --project <project>.", "repair-project-kas", *jsonOutput, "")
	}
	opts := projectinstall.ProjectSuiteOptions{Profile: *profile, Project: *project, SuiteRole: *suiteRole, PruneExtra: *pruneExtra, BackupVaultRoot: *backupVaultRoot, SourcePack: *sourcePack, SourcePackExplicit: hasFlag(argv, "--source-pack"), ProfileRoot: *profileRoot}
	var result projectinstall.ProjectActionResult
	var err error
	if approval != "" {
		result, err = projectinstall.ApplyApprovedRepair(*repo, opts, approval)
	} else {
		result, err = projectinstall.BuildProjectRepairDryRun(*repo, opts)
	}
	if err != nil {
		return emitError(stderr, "project_repair_failed", err.Error(), "repair-project-kas", *jsonOutput, "")
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
		fmt.Fprintln(out, projectinstall.RenderHumanProjectAction(result))
	}
	return code
}

func normalizeInstallArgs(argv []string) []string {
	rewritten := []string{}
	positionals := []string{}
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch arg {
		case "--repo", "--profile", "--project", "--source-pack", "--suite-role", "--profile-root", "--approve", "--apply", "--kab-stage", "--kab-adoption-stage":
			rewritten = append(rewritten, arg)
			if i+1 < len(argv) {
				i++
				rewritten = append(rewritten, argv[i])
			}
		case "--from-generic", "--dry-run", "--json", "--no-color":
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
	for _, name := range []string{"--repo=", "--profile=", "--project=", "--source-pack=", "--suite-role=", "--profile-root=", "--approve=", "--apply=", "--kab-stage=", "--kab-adoption-stage=", "--from-generic=", "--dry-run=", "--json=", "--no-color="} {
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

func isVersionArg(arg string) bool {
	return arg == "--version" || arg == "-version"
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: kkachi-agent-skills <command> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Available commands:")
	fmt.Fprintln(w, "  list     List available KAS skill packs")
	fmt.Fprintln(w, "  install  Plan KAS install or project-suite setup")
	fmt.Fprintln(w, "  doctor   Verify a profile-scoped KAS install")
	fmt.Fprintln(w, "  repair   Plan project-suite repair without writing")
	fmt.Fprintln(w, "  toolchain  Generate and validate ignored .kkachi/toolchain.yaml")
	fmt.Fprintln(w, "  workflow-create   Plan custom task-DAG workflow candidates without writing")
	fmt.Fprintln(w, "  workflow-promote  Propose run-local workflow promotion without writing")
	fmt.Fprintln(w, "  workflow-route    Route classified tasks to one standard bundle without KAH calls")
	fmt.Fprintln(w, "  workflow-trigger  Render dispatch packets for an explicit or selector-matched KAH workflow")
	fmt.Fprintln(w, "  uninstall  Plan project-suite removal without writing")
	fmt.Fprintln(w, "  version  Print CLI version information")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Compatibility commands:")
	fmt.Fprintln(w, "  sync-project-kas, install-project-kas, repair-project-kas")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Use \"kkachi-agent-skills <command> --help\" for command options.")
}

func printToolchainHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage of toolchain:")
	fmt.Fprintln(w, "  kkachi-agent-skills toolchain init --project-root <path> --json")
	fmt.Fprintln(w, "  kkachi-agent-skills toolchain doctor --project-root <path> --json")
	fmt.Fprintln(w, "  kkachi-agent-skills toolchain refresh --project-root <path> --json")
	fmt.Fprintln(w, "  kkachi-agent-skills toolchain import-legacy --project-root <path> --profile <profile> --project <id> --json")
	fmt.Fprintln(w, "  kkachi-agent-skills toolchain set-stage --project-root <path> --stage <stage> --approval-evidence <ref> --json")
	fmt.Fprintln(w, "  kkachi-agent-skills toolchain install-launchers --bin-dir <path> --json")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  init               Create .kkachi/toolchain.yaml from KAS facts and KAH probe facts")
	fmt.Fprintln(w, "  doctor             Validate existing .kkachi/toolchain.yaml without writing")
	fmt.Fprintln(w, "  refresh            Update observed facts while preserving stage approval and policy fields")
	fmt.Fprintln(w, "  import-legacy      Import explicit legacy profile/project state without overwriting conflicts")
	fmt.Fprintln(w, "  set-stage          Record approved Stage 1 metadata; fail closed for unauthorized stages")
	fmt.Fprintln(w, "  install-launchers  Install embedded kkachi-agent-*-toolchain wrappers")
}

func defaultLauncherBinDir() string {
	home := defaultUserHome()
	if strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".local", "bin")
}

func defaultUserHome() string {
	home, err := os.UserHomeDir()
	if err == nil && strings.TrimSpace(home) != "" && !strings.Contains(home, string(filepath.Separator)+".hermes"+string(filepath.Separator)+"profiles"+string(filepath.Separator)) {
		return home
	}
	user := strings.TrimSpace(os.Getenv("USER"))
	if user == "" {
		user = strings.TrimSpace(os.Getenv("LOGNAME"))
	}
	if user != "" {
		for _, candidate := range []string{filepath.Join("/Users", user), filepath.Join("/home", user)} {
			if st, err := os.Stat(candidate); err == nil && st.IsDir() {
				return candidate
			}
		}
	}
	if err == nil {
		return home
	}
	return ""
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
		fmt.Fprintln(w, "Error: ["+code+"] "+message)
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

func emitResult(stdout io.Writer, stderr io.Writer, ok bool, jsonOutput bool, value any, renderHuman func() string) int {
	out := stdout
	code := 0
	if !ok {
		out = stderr
		code = 2
	}
	if jsonOutput {
		_ = writeJSON(out, value)
	} else {
		fmt.Fprintln(out, renderHuman())
	}
	return code
}

func envValue(env map[string]string, key string) string {
	if env != nil {
		return env[key]
	}
	return os.Getenv(key)
}

func splitFlagList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	items := []string{}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
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

func projectInstallApproveMissingValue(argv []string) bool {
	for i, arg := range argv {
		if strings.HasPrefix(arg, "--approve=") {
			return strings.TrimPrefix(arg, "--approve=") == ""
		}
		if arg == "--approve" {
			return i+1 >= len(argv) || strings.HasPrefix(argv[i+1], "--")
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

func hasFlag(argv []string, name string) bool {
	for _, arg := range argv {
		if arg == name || strings.HasPrefix(arg, name+"=") {
			return true
		}
	}
	return false
}
