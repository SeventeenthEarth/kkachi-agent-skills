# Kkachi Hermes Skills

Kkachi Agent Skills (KAS; repository historically named `kkachi-hermes-skills` / KHS) is the Hermes Agent skill/process pack for running
Kkachi software-development workflows.

KHS does not execute external AI CLIs directly and does not own project state.
It teaches Hermes Agent how to classify a task, prepare an AI-neutral task
contract, select a KAB backend lane, render the backend-specific prompt, call
KAH for deterministic run state, call KAB for backend execution, and preserve
evidence for review, verification, documentation, and self-improvement.

Maturity note: KAS is now in the post-KAH 0.1.4 enablement lane. It is still an
early skill/process pack rather than a final polished product, but KAH graph and
configurable-feedback substrates are capability-evidenced. Real Hermes/Kkachi
runs should mature KAS through captured evidence, project overlays, prompt/phase
skill references, reusable scripts, and the existing `kkachi-improve` /
improvement-promotion rules.

## Current Lane Split

KHS now keeps two lanes distinct:

1. KHS+KAH minimum/pilot harness lane
   - Scope: profile-scoped KHS skill-pack install/list/doctor/sync/proposal
     support through a future thin `kkachi-hermes-skills` CLI.
   - Purpose: let users pilot KHS skill injection and KAH evidence/proposal
     workflows without requiring KAB runtime readiness.
   - Non-goals: no `run` verb, backend session control, bridge control, KHC
     command/control, Doksuri integration, or KAB replacement.
   - Candidate record: `docs/sot/minimum-pilot-cli-lane.md`.

2. Full execution-runtime lane
   - Scope: KHS-governed code-change or backend-executed runs.
   - Authority: existing KHS+KAH+KAB path remains the required execution path
     for KAB-backed KHS runs.

## Components

```text
KHS: kkachi-hermes-skills
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
report that KAB is required before any KAB-backed KHS code-change run.

If the master says "apply KHS to this project directory", Hermes should not
manually create ad hoc state files. Hermes should run KAH `project init` in the
target project, using the project's docs map, backend policy, commander,
red-team partner, SOT policy, execution mode, and test commands.

After project init, Hermes should use KAS/KHS skills for Kkachi-governed runs. `.kkachi-workflow.yaml` is the project workflow graph only when backed by KAH graph validation/proposal/apply evidence; `phase-plan.yaml` remains run-local execution state/evidence. KAH `work_path`, `work_mode`, and `execution_mode` remain deterministic helper metadata only.

```text
orchestrate
  -> task-contract
  -> phase-plan
  -> backend-select
  -> prompt-compose
  -> plan / ask
  -> implement / enhance-test / optimize
  -> docs-update / request-feedback(1..5, round 1 required, rounds 2-5 optional) / handle-feedback(1..5)
  -> verify / final-verify
  -> improve
```

The user selects the target roadmap task for each run. Hermes manages, approves risk, and final-verifies; KAB backend roles do the substantive planning, implementation, docs, feedback, and feedback handling. Code-change KHS runs use KAB; if the user forbids KAB for code changes, treat the work as a normal direct Hermes task instead of a KHS run.

## Install Flow

### 1. Install KHS

Use the Hermes native skill installer only for behavior that has been verified,
or place the skill directories where Hermes expects skills.

Verified constraint: current live Hermes evidence proves a single hub identifier
or direct `SKILL.md` URL install path, not repo-root multi-skill-pack install.
Do not claim that installing this repository root installs every KHS skill until
that behavior is separately evidenced. For the minimum/pilot lane, the future
`kkachi-hermes-skills install <profile> <skill-or-category>` path must default
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

Hermes must select KAB backend lanes from evidence, not vague preference.

Use:

- `registries/cli-capabilities.yaml`
- `registries/backend-selection-policy.yaml`
- `registries/backend-prompt-profiles.yaml`
- the target project's backend policy
- `kkachi-agent-bridge/docs/public/compatibility-matrix.md`

User preference can rank eligible backends, but it must not override missing
required capability, project policy, or compatibility matrix caveats.

## KAB Execution Evidence

KAB prompt dispatch has two supported observation styles:

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
