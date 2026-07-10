# Kkachi Agent Skills

Kkachi Agent Skills (KAS; repository now `kkachi-agent-skills`, historically `kkachi-hermes-skills` / KHS) is the Hermes Agent skill/process pack for running
Kkachi software-development workflows.

KHS does not own project state and does not replace KAH or KAB runtime control.
It teaches Hermes Agent how to classify a task, prepare an AI-neutral task
contract, call KAH for deterministic run state, and preserve evidence for
review, verification, documentation, and self-improvement. In the v0.2 active
path, KAS owns the workflow/prompt/contract layer, KAH owns deterministic
run/gate/evidence state, and GJC may produce candidate deep-interview,
`ralplan`, or `ultragoal` artifacts through KAH wrappers. KAT evidence is
mechanical/factual only. KAB runtime/session control is used only when a later
approved task explicitly selects a KAB-backed lane and preserves bridge evidence.

Maturity note: KAS v0.2.5 is the current TWAKE source-readiness closeout package
after the v0.2.1 release-readiness baseline; it is paired with KAH v0.2.3
return-path evidence source readiness. Older KAS v0.2.4/v0.2.3/v0.2.1/v0.2.0/v0.1.x snapshots remain
historical release context only; they are not active
operator, migration, compatibility, fallback, warning-only, or profile-suite
paths. Real Hermes/Kkachi runs should mature KAS through captured evidence,
project overlays, prompt/phase skill references, reusable scripts, and the
existing `kkachi-improve` / improvement-promotion rules.

TWAKE source readiness is source-only. It does not imply tagged release,
install, runtime activation, Discord delivery, provider execution, KAB
activation, or auth/provider/profile/gateway/model mutation. Kanban review cards
remain routing and return-path evidence only; Kkachi MAR provider coverage must
come from KAH `mar` / provider-attempt artifacts using the declared zcode/kimi/agy
lanes or an explicit 주군 waiver. The earlier KAS v0.2.3 stage-report contract
remains active for task classification, plan, implementation, and review-stage
reports.

## Current Lane Split

KHS now keeps two lanes distinct:

1. KHS+KAH minimum/pilot harness lane
   - Scope: profile-scoped KHS skill-pack install/list/doctor/sync/proposal
     support through the thin `kkachi-agent-skills` CLI as each CLIMVP task
     lands.
   - Purpose: let users pilot KHS skill injection and KAH evidence/proposal
     workflows without requiring KAB runtime readiness.
   - Non-goals: no `run` verb, backend session control, bridge control, KHC
     command/control, Doksuri integration, or KAB replacement.
   - Candidate record: `docs/sot/minimum-pilot-cli-lane.md`.

2. Full execution-runtime lane
   - Scope: KHS-governed backend-executed runs, automated review-by-different-tool transport, KAB plan lifecycle, and bridge evidence.
   - Authority: existing KHS+KAH+KAB path remains required when the run is KAB-backed or claims backend runtime evidence. Scoped CLIMVP/GRAPHMVP/KAS docs or CLI work may proceed without KAB only when the lane is explicitly KAS/KAH-local and records that no KAB runtime support is claimed.

## v0.2 KAS/KAH/GJC Development Delegation

KAS/KAH/KAT repository self-development is an explicit exception to the older
dogfood default: do not run KAS/KAH/KAT development through KAS/KAH/KAT
workflows unless 주군 explicitly selects that mode for the run. 황충 performs
main development directly, then routes the result through official color review
and fixes/re-review. This exception does not waive tests, color review,
commit/release approval, or fail-closed blocker handling.

KAS owns the operating policy for how KAH evidence, GJC candidate execution,
KAT factual test evidence, color review, MAR, and final gates are used. KAH
remains the deterministic state/evidence/gate tool. GJC is a candidate
planner/executor only, never acceptance authority:

1. **Deep interview** — use only for explicitly approved epic/design
   clarification.
2. **Ralplan** — produce plan artifacts for KAS/Blue/color review; completion
   does not authorize implementation.
3. **Ultragoal** — produce approved implementation-candidate artifacts after an
   explicit bounded approval boundary; completion still requires KAS/KAH
   verification, official color review, MAR applicability/coverage, and final
   Blue synthesis.

KAT records mechanical test/status/summary/raw-log evidence only. KAB remains a
separate runtime/session-control lane and is not the default KAS/KAH v0.2
development path unless a task explicitly selects it with current capability
and bridge evidence. Legacy Stage 1/Stage 2/Stage 3 Codex/KAB adoption wording
is historical/stale and must not be used as active project guidance.

MAR review is independent of the implementation candidate lane. It is the
default KAS role-first feedback lane when applicable, uses validated provider
and toolchain evidence, and does not change plan/implementation authority.
Required roles are `logic`, `security`, `arch`, `cve`, and `test_adequacy`;
unresolved required role coverage fails closed.

Stage reports are governed by `docs/sot/stage-report-contract.md`: task
classification, plan completion, implementation completion, and review closeout
reports must explain the functional change, old behavior/reason, plan drift,
test/phase coverage, review feedback, and remaining approval boundary instead of
only listing files or saying a stage completed.

## Components

```text
KHS: kkachi-agent-skills
  Hermes skills, prompt profiles, phase contracts, templates, and registries.

KAH: kkachi-agent-helper
  Deterministic project bootstrap, run artifacts, events, locks, schemas, gates,
  diagnostics, and project-local `.kkachi/` state.

KAB: kkachi-agent-bridge
  Runtime bridge for Claude Code, GLM, Codex CLI, Kimi CLI, OpenCode, and other
  backend sessions when supported by current KAB evidence. Gemini CLI is legacy
  context unless 주군 explicitly re-enables it for new KAB work.
```

KHS is the prompt/process layer. KAH is the deterministic state layer. KAB is the
backend runtime/control layer.

## Current CLI Surface

CLIMVP implements the profile-scoped minimum CLI surface:

```bash
kkachi-agent-skills --version
kkachi-agent-skills version [--json]
kkachi-agent-skills list [--repo <path>] [--profile <profile>] [--category <name>] [--json]
kkachi-agent-skills install [--repo <path>] --profile <profile> <pack-id>... --dry-run [--json]
kkachi-agent-skills install [--repo <path>] --profile <profile> <pack-id>... --approve dry-run:<hash> [--json]
kkachi-agent-skills doctor [--repo <path>] --profile <profile> [--project <path>] [--json]
kkachi-agent-skills doctor [--repo <path>] --profile <profile> --project <project> --project-suite [--json]
kkachi-agent-skills doctor [--repo <path>] --project <path> --workflow-graph --json
kkachi-agent-skills install [--repo <path>] --profile <profile> --project <project> --suite-role <role> --dry-run [--json]
kkachi-agent-skills install [--repo <path>] --profile <profile> --project <project> --suite-role <role> --apply dry-run:sha256:<hash> [--json]
kkachi-agent-skills install [--repo <path>] --profile <profile> --project <project> --suite-role <role> --from-generic --dry-run [--json]
kkachi-agent-skills install [--repo <path>] --profile <profile> --project <project> --suite-role <role> --from-generic --apply dry-run:sha256:<hash> [--json]
kkachi-agent-skills repair [--repo <path>] --profile <profile> --project <project> --dry-run [--json]
kkachi-agent-skills repair [--repo <path>] --profile <profile> --project <project> --apply dry-run:sha256:<hash> [--json]
kkachi-agent-skills repair [--repo <path>] --project <path> --workflow-graph --propose --reason <reason> [--json]
kkachi-agent-skills repair [--repo <path>] --project <path> --workflow-graph --apply-proposal <proposal-id> --approval <approval-ref> [--json]
kkachi-agent-skills toolchain init --project-root <path> --json
kkachi-agent-skills toolchain doctor --project-root <path> --json
kkachi-agent-skills toolchain refresh --project-root <path> --json
kkachi-agent-skills toolchain install-launchers [--bin-dir <path>] [--json]
kkachi-agent-skills workflow-create --project <path> --workflow-id <id> --mode dag_only|thin_trigger|full_trigger --request <json-path> --dry-run [--json]
kkachi-agent-skills workflow-create --project <path> --workflow-id <id> --mode dag_only|thin_trigger|full_trigger --request <json-path> --apply dry-run:sha256:<hash> [--json]
kkachi-agent-skills workflow-promote --project <path> --run <run-id> --target-workflow-id <id> --reuse-reason <reason> [--thin-trigger] --dry-run [--json]
kkachi-agent-skills workflow-promote --project <path> --run <run-id> --target-workflow-id <id> --reuse-reason <reason> [--thin-trigger] --apply dry-run:sha256:<hash> [--json]
kkachi-agent-skills uninstall --profile <profile> --project <project> --dry-run [--json]
kkachi-agent-skills uninstall --profile <profile> --project <project> --apply dry-run:sha256:<hash> --backup-vault-root <abs-path> [--json]
```

`--version` / `version` prints the CLI release before any profile or source-repo
discovery. `--version --json` / `version --json` emits only `name` and `version`.

`toolchain install-launchers` installs KAS-owned embedded local wrappers
`kkachi-agent-skills-toolchain`, `kkachi-agent-helper-toolchain`, and
`kkachi-agent-tester-toolchain` into `--bin-dir` or the default user
`~/.local/bin`. The wrappers read only `.kkachi/toolchain.yaml` schema
`kkachi.toolchain.v1`, resolve effective KAS/KAH/KAT binaries, export the
selected KAH/KAT paths for downstream commands, print paths and versions with
`--toolchain-status`, and fail closed on missing, malformed, unsupported,
non-executable, or version-mismatched metadata. KAT launcher evidence remains
mechanical/factual only and does not grant review, MAR, or final acceptance
authority.

`list` discovers source KAS packs from `skills/`, reports direct-layout packs
as category `core`, supports future `skills/<category>/<skill>/SKILL.md`
layouts, and can compare against a profile manifest at
`~/.hermes/profiles/<profile>/.kas/skill-pack-manifest.json` without creating
profile directories or files.

`install --dry-run` resolves source packs and target profile paths, validates
manifest/checksum inputs, reports create/update/skip/conflict/error paths, emits
a deterministic `dry_run_plan_hash`, and performs no profile writes. Approved
copy install requires `--approve dry-run:<hash>`, recomputes the plan, fails
closed on hash mismatch, copies selected packs into the target profile, records
manifest/checksum evidence, and prints recovery guidance.

For project-specific `kkachi-agent-skills` suites, the public lifecycle commands are
`install`, `doctor`, `repair`, and `uninstall`. Public `update` and `migrate`
commands are intentionally not exposed: semantic overlay refresh belongs in the
`kkachi-agent-skills-overlay-refresh` skill, not in the deterministic CLI.
Dry-run forms preserve the TOKEN-004 no-write behavior: `install --project
--suite-role <role> --dry-run` uses the project-suite install planner and plans
the selected project-prefixed phase skills plus the canonical
`<project>-wrapper`, `<project>-overlay`, and
`<project>-overlay/references/legacy-delta-extract.md` composition files;
`install --from-generic --dry-run` plans explicit generic-to-project setup,
`repair --dry-run` uses the project repair planner, and `uninstall --dry-run`
reads the manifest and filesystem to plan manifest-tracked removals while
skipping local-only or unmanifested files. TOKEN-005 write forms require
`--apply dry-run:sha256:<hash>` from the matching current dry-run plan and fail
closed before mutation on malformed, mismatched, or stale evidence. `uninstall
--apply` additionally requires `--backup-vault-root <abs-path>` and rejects
missing, relative, profile-contained, symlink-unsafe, or unwritable backup roots.

Compatibility commands remain available for existing automation:
`sync-project-kas`, `install-project-kas`, and `repair-project-kas`. Public
compatibility `migrate-project-kas` is removed to avoid implying that the CLI
can perform semantic migration.

`doctor` verifies source pack integrity, installed profile state,
manifest/checksum consistency, KAH availability/version/capabilities, optional
project bootstrap/doctor status, and the KAB boundary for the requested lane.
`toolchain init`, `refresh`, `import-legacy`, and `set-stage` do not
materialize project-local KAS `mar.py`, shell adapters, provider-lane registries,
or adapter-proof files. `toolchain doctor --project-root <path> --json` treats
copied project-local KAS `mar.py` or shell adapter surfaces as V01CLEAN-001
fail-closed diagnostics (`mar_legacy_surface_present`) with
`no_deletion_without_approval: true` and `live_provider_execution: false`; non-empty
local `mar.provider_tools.providers.*` proof fields also fail closed with
`toolchain_mar_provider_tools_legacy_surface`. It must not delete local files,
refresh legacy wrappers, or execute providers.
`doctor --project <path> --workflow-graph --json` is a separate read-only
workflow graph supportability check: it reads KAS compatibility metadata,
effective KAH version/capabilities/help, and KAH graph validate/explain evidence
for `.kkachi-workflow.yaml`; it classifies pass/custom-supported/update/graph
states without calling graph init/diff/propose/apply/export or writing project,
profile, KAH, KAB, auth, provider, gateway, token, or model state.
`repair --project <path> --workflow-graph --propose --reason <reason> --json`
is the explicit opt-in proposal path: it runs the read-only doctor first,
requires the KAH graph diff/propose/apply capability envelope, writes a complete
candidate graph from the reviewed KAS template input, calls KAH graph
diff/propose, and reports proposal evidence without applying the graph.
`repair --project <path> --workflow-graph --apply-proposal <proposal-id>
--approval <approval-ref> --json` is the approval-gated apply path: it reruns
preflight, calls KAH graph apply, then reruns KAH validate/explain. Periodic
checks must default to doctor/report only; proposal is opt-in and apply is never
automatic from cron or CI.
`workflow-create --dry-run` plans WFLOW-004 custom task-DAG workflow candidates
for `dag_only`, `thin_trigger`, and exceptional `full_trigger` modes. It emits
compact operator output plus a full machine packet with candidate DAG, catalog,
node-contract, trigger paths, generated content, selector metadata, KAH/KAS
capability evidence, base checksums, changed paths, diagnostics, no-write
evidence, and a canonical `sha256` approval hash. `workflow-create --apply`
recomputes the packet hash and fails closed before any write or KAH delegation
on mismatches or missing KAH workflow/catalog capability. KAS does not
direct-write `.kkachi` workflow state or install generated trigger skills into a
Hermes profile; KAH remains authoritative for workflow/catalog validation,
proposal, apply, audit, and final gate evidence.
`workflow-promote --dry-run` implements WFLOW-009 as an explicit promotion
proposal from an existing WFLOW-008 run-local materialization bundle to
project-local workflow/catalog/node-contract candidates and an optional thin
trigger. It requires `--target-workflow-id` and `--reuse-reason`, verifies
`materialization.json`, `workflow.yaml`, `node-contracts.json`, and checksum
evidence, emits `kas-workflow-promote-packet/v1`, and binds the approval hash to
source provenance/checksums, target paths, generated content, trigger plan,
KAS/KAH capability evidence, base checksums, changed paths,
diagnostics/conflicts, and no-write evidence. `workflow-promote --apply`
recomputes the hash and requires effective KAH DAGSM-006 / v0.1.10 catalog
proposal/apply support before persistent writes; KAS does not direct-write `.kkachi/workflows/*`,
`.kkachi/workflow-catalog.yaml`, `.kkachi-workflow.yaml`, profile files, KAH
state, KAB state, auth/token/provider/gateway/model config, or fallback backend
selection; no fallback backend selection is introduced.
KAB is not required for this profile-scoped minimum CLI lane. KAB remains
required for backend execution, automated review-by-different-tool transport,
KAB plan lifecycle, and bridge evidence when those surfaces are in scope.

`--profile-root <path>` is accepted only under the explicit test/harness guard
`KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1`.

Verification for this surface is:

```bash
go install .
make test
```

The Go module path is `github.com/SeventeenthEarth/kkachi-agent-skills`.
Local development installs can use `go install .`; the remote install path is
`go install github.com/SeventeenthEarth/kkachi-agent-skills@latest`. The root
binary embeds `skills/`, `templates/`, `registries/`, and `skill-pack.yaml`, so
normal installs do not require a local source checkout. Use `--repo <path>` only
when intentionally overriding the embedded source with a local development checkout.

## When KHS Should Trigger

KHS is not the default path for every small Hermes coding edit. Trigger KHS only
when the user or operating context asks for a Kkachi-governed software run.

Use KHS when any of these are true:

- the master explicitly says to use KHS, Kkachi, KAB, KAH, or a Kkachi run;
- the master asks to apply KHS to a project directory;
- the task is assigned to a Kkachi development commander such as 조운 or 마초, or
  says that named generals should develop through KHS;
- the work needs KAB backend execution, backend lane selection, bridge evidence,
  red-team review, gate reports, or durable run artifacts;
- the task is a substantial feature, risky refactor, multi-phase QA, production
  write, readiness hardening, or discovery/shaping run.

Do not trigger KHS by default for simple direct Hermes edits such as typo fixes,
small one-file patches, quick config tweaks, or read-only explanation/review
requests unless the master explicitly asks for KHS/Kkachi or delegates the work
to a KHS-using commander.

## When Hermes Agent Reads This README

If the master gives Hermes Agent this repository URL and says "install this"
for the full execution-runtime path, Hermes should install and verify the whole
Kkachi stack:

1. Install KHS through the Hermes native skill system or the KHS-supported
   profile install path once implemented.
2. Install or verify the latest KAH (`kkachi-agent-helper@latest`).
3. Clone/build or verify KAB (`kkachi-agent-bridge`) and its plugin/wrapper artifacts.
4. Verify core commands are available.
5. Report the installed status in Korean.

If the master asks only for the scoped KHS+KAH minimum/pilot harness, KAB
verification is not mandatory for profile-scoped skill install/injection,
`list`, `doctor`, `sync`, or `proposal` planning. That scoped lane must still
report that KAB is required before KAB-backed execution, automated
review-by-different-tool transport, KAB plan lifecycle, or bridge evidence are
claimed. Current KAS/KAH-local lanes use KAS contracts, KAH evidence, and approved GJC candidate artifacts when needed, with no KAB runtime claim unless KAB is explicitly selected and evidenced.

If the master says "apply KHS to this project directory", Hermes should not
manually create ad hoc state files. Hermes should run KAH `project init` in the
target project, using the project's docs map, backend policy, commander,
red-team partner, SOT policy, execution mode, and test commands.

After project init, Hermes should use KAS/KHS skills for Kkachi-governed runs. `.kkachi-workflow.yaml` is the project workflow graph only when backed by capability-checked `kkachi-agent-helper graph` validation/proposal/apply evidence; `phase-plan.yaml` remains run-local execution state/evidence. If the effective KAH binary lacks required graph support, record a gap and continue only with run-local phase evidence rather than writing `.kkachi-workflow.yaml` manually. KAH `work_path`, `work_mode`, and `execution_mode` remain deterministic helper metadata only.

KHS task classification is not a global Hermes personality or all-chat rule. Default Hermes mode remains direct commander response. Classification starts only after KHS/Kkachi project-execution mode is active: explicit KHS/Kkachi/KAB/KAH instruction, applying KHS to a project directory, durable repo artifact changes under a governed Kkachi project, phase/gate evidence requirements, backend execution/KAB plan lifecycle/bridge evidence claims, or long-lived team collaboration.

Every active KHS run then starts by classifying the task class: `development`, `research_evidence`, `docs_only`, `simple_command_report`, `bootstrap_config`, or `collaboration_review`. The full development loop below applies only to `development`; lighter classes record explicit skipped-phase reasons and use evidence/docs/review/config verification instead of implementation/test/optimize phases. State investigation that only reports facts maps to `research_evidence`; changing spec/SOT/roadmap/handoff from that evidence maps to `docs_only + Path B shaping` unless executable behavior or execution policy changes.

```text
orchestrate
  -> task-classification
  -> task-contract
  -> phase-plan
  -> codegraph-refresh(index, or init -i when first initialization is due)
  -> backend-select
  -> prompt-compose
  -> plan / ralplan candidate evidence / explicit approval boundary when project policy requires it
  -> implement / selected verification profile or gate
  -> enhance-test(unit, integration, e2e) / selected verification profile or gate
  -> AI slop cleanup / selected verification profile or gate
  -> optimize(duplication, abstraction, algorithm/structure) / selected verification profile or gate
  -> docs-update / roadmap-update
  -> Blue first review / request-feedback(하후연, 여몽, 진궁 when required) / handle-feedback
  -> verify / final graph phase (final-verify skill alias)
  -> improve
```

The user selects the target roadmap task for each run. Hermes manages, approves risk, and final-verifies. KAB backend roles perform substantive planning, implementation, docs, feedback, and feedback handling only for KAB-backed phases. Scoped KAS/KAH-local CLIMVP, GRAPHMVP, and docs-only maintenance may proceed without KAB when explicitly authorized and recorded, but must not claim backend execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence. For development-path runs, the default commander workflow is plan-first, fixed-plan, backend implementation, selected verification profile/gate checkpoints, test enhancement, AI slop cleanup, bounded optimization, docs/roadmap update, Blue review, required role reviews, and review-ready pre-commit reporting. The selected aggregate gate may be `make test` for this repository when the task contract/profile requires it, but KHS must not assume that globally. For research/evidence, docs-only, simple command/report, bootstrap/config, and collaboration/review tasks, the selected light spine must fit the class and the final report must say which development phases were skipped and why.

## Install Flow

### 1. Install KHS

Use the Hermes native skill installer only for behavior that has been verified,
or place the skill directories where Hermes expects skills.

Verified constraint: current live Hermes evidence proves a single hub identifier
or direct `SKILL.md` URL install path, not repo-root multi-skill-pack install.
Do not claim that installing this repository root installs every KHS skill until
that behavior is separately evidenced. For the minimum/pilot lane, the future
`kkachi-agent-skills install <profile> <skill-or-category>` path must default
to copy mode with dry-run, manifest/checksum, changed-path report, and recovery
instruction.

If Hermes installs skills one directory at a time, install every directory under
`skills/`.

### 2. Install KAH

```bash
go install github.com/SeventeenthEarth/kkachi-agent-helper@latest
kkachi-agent-helper --version
```

KHS uses the current KAH project-bootstrap command surface. Do not pin KHS to a
single helper patch version unless a future KAH release breaks `project init`,
run artifacts, gates, schemas, or diagnostics.

### 3. Build And Install KAB

```bash
git clone https://github.com/SeventeenthEarth/kkachi-agent-bridge.git
cd kkachi-agent-bridge
make build
make build-opencode-plugin
cargo install --path plugins/codex --locked
mkdir -p "$HOME/.local/bin"
cp bin/kkachi-agent-bridge "$HOME/.local/bin/kkachi-agent-bridge"
kkachi-agent-bridge help
kkachi-codex-wrapper --help
```

KAB is not a single `go install` package. The Go bridge, OpenCode TypeScript
plugin, and Codex Rust wrapper are separate build surfaces:

- Bridge binary: `make build` -> `bin/kkachi-agent-bridge`
- OpenCode plugin: `make build-opencode-plugin` -> `plugins/opencode/dist/`
- Codex wrapper: `cargo install --path plugins/codex --locked` -> `kkachi-codex-wrapper`

Keep the KAB checkout available after copying the bridge binary if OpenCode
hybrid plugin assist may be used. The bridge stages the managed OpenCode plugin
from the built checkout's `plugins/opencode` package. For Codex, either install
`kkachi-codex-wrapper` on `PATH` or configure KAB `codex_wrapper_command` with
an absolute wrapper path.

Backend CLIs must still be installed separately for the lanes the project will
use: `claude`, `glm`, `codex`, `gemini`, and/or `opencode`.

### 4. Verify The Stack

```bash
kkachi-agent-helper --version
kkachi-agent-helper project doctor --json
kkachi-agent-bridge help
kkachi-codex-wrapper --help
test -d <kab-repo>/plugins/opencode/dist
```

`project doctor` may report that no Kkachi project has been initialized when run
outside a target repository. That is acceptable during global install
verification.

## Apply KHS To A Project

From the target project directory, Hermes should prepare project-specific values
and run:

```bash
kkachi-agent-helper project init \
  --project-name <project-name> \
  --stack <stack-name> \
  --repo-path "$PWD" \
  --commander <Hermes-commander> \
  --redteam <Hermes-redteam> \
  --docs-map-roadmap <path-to-roadmap> \
  --docs-map-spec <path-to-spec> \
  --docs-map-architecture <path-to-architecture> \
  --docs-map-adr-dir <path-to-adr-dir> \
  --docs-map-todo-dir <path-to-todo-dir> \
  --docs-map-spec-dir <path-to-spec-dir> \
  --test-commands "<command1>,<command2>" \
  --backend-policy "<allowed-backends>" \
  --execution-mode <production_write|adapter_qa|readiness_hardening|research|verification|docs_only> \
  --sot-policy <existing_sot_basis|minimal_sot_before_code|full_sot_before_code> \
  --json
```

Use `--force` only for non-destructive reconfiguration of existing KAH project
bootstrap files. KAH preserves runs, artifacts, events, and gate history during
that reconfiguration.

### v0.2 project execution posture at install/reconfigure time

Project application records the active v0.2 posture before the first
Kkachi-governed run: KAS policy and prompt contracts, KAH deterministic
run/gate/evidence state, GJC candidate `deep-interview`/`ralplan`/`ultragoal`
only when approved, and KAT factual evidence only. KAB is not selected by
default. If a project needs KAB runtime/session control, that task must name the
KAB lane explicitly, pass current capability and compatibility gates, and
preserve bridge evidence in the run artifacts.

KAH does not interpret KAS policy semantically: graph state, run state, gates,
events, and artifact validation remain helper mechanics. KAS owns the active
workflow policy and records it in project-specific KAS guidance first. Missing,
ambiguous, or stale legacy Stage 1/Stage 2/Stage 3 markers fail closed to
"no KAB runtime claim"; they do not authorize direct Codex, KAB `native_codex`,
or backend-selected execution.

After initialization, create and run Kkachi work through KAH artifacts and gates:

```bash
kkachi-agent-helper run create ... --json
kkachi-agent-helper run activate <run_id> --json
kkachi-agent-helper artifact init <run_id> --json
kkachi-agent-helper gate check <run_id> <gate> --json
kkachi-agent-helper gate final <run_id> --json
```

## Phase Orchestration Policy

For the user-confirmed phase behavior, see `docs/sot/phase-orchestration-policy.md`. The executable phase contract is `registries/phase-contracts.yaml`. Documentation layers are described in `docs/README.md`: durable SOT documents live under `docs/sot/`, while temporary discussion notes live under `docs/discussions/`.

## Backend Selection

KAB backend selection is explicit-only in the v0.2 KAS/KAH path. Hermes must select KAB backend lanes from evidence, not vague preference, and must not treat legacy Stage markers or direct Codex wording as active authorization.

Use:

- `registries/cli-capabilities.yaml`
- `registries/backend-selection-policy.yaml`
- `registries/backend-prompt-profiles.yaml`
- the target project's backend policy
- `kkachi-agent-bridge/docs/public/compatibility-matrix.md`

User preference can rank eligible backends, but it must not override missing
required capability, project policy, or compatibility matrix caveats.

## KAB Execution Evidence

This section applies only to explicitly approved KAB-backed work. KAB prompt dispatch has two supported observation styles:

```text
cli_loop:
  send -> wait -> read/status -> approve/reject/answer -> repeat -> stop

retained_stream:
  send -> GET /api/stream/<session_id> or GET /api/events/<session_id>
       -> approve/reject/answer as pendings appear
       -> final read/status -> stop
```

`send` success is dispatch evidence only. Completion requires public session
evidence through `wait + read/status` or retained stream/event observations plus
final `read/status`.

Record the selected observation mode in `bridge-events.md` as `cli_loop`,
`retained_stream`, or `hybrid`.

## Repository Layout

```text
skills/
  kkachi-orchestrate/
  kkachi-task-contract/
  kkachi-backend-select/
  kkachi-prompt-compose/
  kkachi-phase-state/
  kkachi-plan/
  kkachi-ask/
  kkachi-implement/
  kkachi-enhance-test/
  kkachi-optimize/
  kkachi-review/
  kkachi-verify/
  kkachi-docs-update/
  kkachi-request-feedback/
  kkachi-handle-feedback/
  kkachi-final-verify/
  kkachi-improve/
  kkachi-install-guide/

registries/
  cli-capabilities.yaml
  backend-selection-policy.yaml
  backend-prompt-profiles.yaml
  phase-contracts.yaml
  task-taxonomy.yaml
  improvement-promotion-policy.yaml

templates/
  run-artifacts/
    phase-plan.yaml.tmpl
    checklist.md.tmpl
    plan.md.tmpl
  prompts/
  project-overlay/
```

## Operating Rule For Hermes

Do not turn KHS into a second state system. KHS guides Hermes. KAH records state
and gates. KAB runs backend sessions. Project-specific customization should land
in the project overlay or project-local artifacts first; promote a change back
to shared KHS only after repeated evidence and review.
