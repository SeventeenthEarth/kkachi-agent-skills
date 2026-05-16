# KHS Reality Interface Contract

Date: 2026-05-12
Owner: Gongmyeong
Status: source of truth for the current usable KHS/KAH/KAB interface

## Purpose

This document locks the current practical interface that `kkachi-hermes-skills` (KHS) expects while `kkachi-agent-helper` (KAH) and `kkachi-agent-bridge` (KAB) continue parallel development.

The contract is deliberately reality-first: KHS should drive development through the command surfaces that exist now, not through idealized future surfaces. When KHS needs behavior that KAH or KAB does not yet expose, KHS must record the gap as feedback or a roadmap item instead of pretending the interface exists.

## Layer ownership

```text
KHS:
  Hermes skills, phase contracts, task contracts, prompt profiles, backend-selection policy,
  project overlays, artifact templates, and self-improvement rules.

Hermes:
  Manager/orchestrator judgment, KAH/KAB invocation, risk approval routing,
  checklist upkeep, final verification, and Korean user report.

KAH:
  Deterministic project-local state, run ids, artifacts, events, locks, schemas,
  diagnostics, and gate checks.

KAB:
  External backend runtime sessions, backend identity, prompt dispatch, wait/read/status,
  plan lifecycle, approval/input control, retained events, and bridge evidence.
```

KAH and KAB validate or execute KHS decisions. They do not choose the KHS workflow, commander, backend lane, phase applicability, or project policy.

## Reality baseline checked

- KAH installed PATH binary checked: `/Users/draccoon/go/bin/kkachi-agent-helper`
- KAH installed PATH version checked: `kkachi-agent-helper 0.1.1`
- KAH current-candidate repo binary checked: `/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-helper/bin/kkachi-agent-helper`
- KAH current-candidate repo version checked: `kkachi-agent-helper 0.1.2 commit be5ddb7` (`git describe`: `v0.1.1-8-gbe5ddb7`)
- Important PATH note: the installed KAH PATH binary still lags the repo current-candidate for help and capability discovery: PATH v0.1.1 rejects top-level `--help`, `help`, `gate --help`, `capabilities --json`, and `--json capabilities` with exit 2, while the repo current-candidate exposes those surfaces. KHS must verify the effective KAH binary before depending on current-candidate help or capability output.
- KAB repo binary checked: `/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-bridge/bin/kkachi-agent-bridge`
- KAB installed PATH binary checked: `/usr/local/bin/kkachi-agent-bridge`
- Important PATH note: the installed `/usr/local/bin/kkachi-agent-bridge` did not expose `plan` in the observed help output, while the repo `bin/kkachi-agent-bridge` did. KHS runs that require plan lifecycle must verify the effective KAB binary before use.

## KHS activation boundary

KHS is active only for explicit or context-bound Kkachi-governed software runs:

- the master names KHS, Kkachi, KAH, or KAB;
- the master asks to apply KHS to a project;
- the master requests a Kkachi run;
- the master assigns development to a KHS-using commander such as 조운 or 마초;
- the work requires bridge-backed implementation, durable KAH artifacts, gates, red-team review, or backend evidence.

Simple direct Hermes edits, typo fixes, small one-file config patches, and read-only explanations stay outside KHS unless explicitly moved into KHS.

Once KHS is activated, KAH is mandatory for deterministic run state. KHS code-change and development runs use KAB. If the master requests code changes without KAB, treat the task as a normal direct Hermes task rather than a KHS run.

## KHS-owned artifacts

KHS owns the content and lifecycle of these run artifacts even when KAH stores them:

- `task-contract.yaml`
- `phase-plan.yaml`
- `acceptance-criteria.md`
- `plan.md`
- `checklist.md`
- `selected-cli.json`
- `capability-check.md`
- `prompt.md`
- `bridge-session-snapshot.json`
- `bridge-events.md`
- phase evidence such as `impl-log.md`, `test-log.md`, `verification.md`, `review.md`, `docs-update.md`, feedback artifacts, `final-report.md`, and `improvement-note.md`

KAB may supply backend-authored plan text, events, status snapshots, and output. KHS/Hermes must preserve those into KAH artifacts before treating them as run evidence.

## KAH interface used by KHS

### Bootstrap and inspection

```bash
kkachi-agent-helper project init \
  --project-name <name> \
  --stack <stack> \
  --repo-path <path> \
  --commander <profile> \
  --redteam <profile> \
  --docs-map-roadmap <path> \
  --docs-map-spec <path> \
  --docs-map-architecture <path> \
  --docs-map-adr-dir <path> \
  --docs-map-todo-dir <path> \
  --docs-map-spec-dir <path> \
  --test-commands <comma-separated> \
  --backend-policy <policy> \
  --execution-mode <mode> \
  --sot-policy <policy> \
  --json

kkachi-agent-helper project status --json
kkachi-agent-helper project doctor --json
```

Reality note: the KAH current-candidate repo binary exposes conventional top-level `--help` and `help` command UX, but the installed PATH v0.1.1 baseline still rejects those commands. KHS must use capability discovery and observed command exits from the effective binary as the machine activation source; help output is supplemental documentation, not a version-guess substitute.

### Run lifecycle

```bash
kkachi-agent-helper run create \
  --title <title> \
  --work-path <A_development_execution|B_discovery_shaping> \
  --work-mode <standard|light> \
  --urgency <normal|urgent|critical> \
  --sot-policy <existing_sot_basis|minimal_sot_before_code|full_sot_before_code> \
  --execution-mode <production_write|adapter_qa|readiness_hardening|research|verification|docs_only> \
  --commander <profile> \
  [--redteam <profile>] \
  [--task-id <id>] \
  --json

kkachi-agent-helper run activate <run_id-or-unique-prefix> --json
kkachi-agent-helper run show <run_id-or-unique-prefix> --json
kkachi-agent-helper run list --json
kkachi-agent-helper run close <run_id-or-unique-prefix> --json
kkachi-agent-helper run abort <run_id-or-unique-prefix> --json
```

KAH run metadata is helper classification. It seeds deterministic requirements but does not decide which KHS phases execute. `phase-plan.yaml` remains the workflow SOT.

### Artifact, event, schema, diagnostics, and gates

```bash
kkachi-agent-helper artifact init <run_id-or-unique-prefix> --json
kkachi-agent-helper artifact list <run_id-or-unique-prefix> --json
kkachi-agent-helper artifact validate <run_id-or-unique-prefix> [--gate <gate>] --json

kkachi-agent-helper event append <type> --run <run_id> --payload '<json-object>' --json

kkachi-agent-helper schema validate <file> --schema <config|status|event|run-metadata|selected-cli|bridge-session-snapshot> --json
kkachi-agent-helper schema export [--schema <name>|--all] [--dry-run] --json
kkachi-agent-helper schema migrate --from <version> --to <version> [--dry-run] --json

kkachi-agent-helper gate check <run_id-or-unique-prefix> <gate> --json
kkachi-agent-helper gate final <run_id-or-unique-prefix> --json
kkachi-agent-helper diagnostics export [--run <run_id-or-unique-prefix>] [--output <repo-relative-path>] --json
```

Current KAH gate names include:

```text
intake, sot, roadmap, plan, backend, implementation, review, verification, docs, final
```

Reality correction: the KAH current-candidate repo binary supports both `gate check <run_id> final --json` and the dedicated final gate shortcut `gate final <run_id> --json`. KHS should use the explicit canonical final gate invocation when the observed effective capability surface includes `gate.final` or an equivalent `gate` command group with `final`; compatibility fallbacks must be based on capabilities and command-exit evidence, not guessed from a version string.

### KAH schema boundary

KAH v0.1.1 validates core fields only for the bridge-selection artifacts:

`selected-cli.json` required core:

```json
["version", "status", "backend_type", "adapter_type", "source_ledger_ref", "caveats"]
```

`bridge-session-snapshot.json` required core:

```json
["version", "session_id", "backend_type", "adapter_type", "state", "lifecycle_class", "open_pendings"]
```

KHS templates may include richer fields such as tested CLI, capability snapshot, prompt profile, observation mode, plan lifecycle, and accepted caveats. Those fields are KHS-level evidence and must remain compatible with KAH's core schema.

## KAB interface used by KHS

### Required binary capability check

Before a KHS run that uses KAB plan lifecycle, verify the effective KAB binary exposes `plan`:

```bash
kkachi-agent-bridge --help
kkachi-agent-bridge plan --help
```

If the PATH binary lacks `plan`, use the project-built binary or rebuild/install KAB before starting the KHS run.

### Session lifecycle and observation

KHS selects a backend by selecting or generating the KAB config whose `backend_type` matches `selected-cli.json`; the observed CLI does not provide a `send --backend` flag.

```bash
kkachi-agent-bridge send --new --cwd <project-root> <prompt>
kkachi-agent-bridge send --session <session_id> <prompt>
kkachi-agent-bridge wait --session <session_id> --timeout <seconds>
kkachi-agent-bridge read --session <session_id>
kkachi-agent-bridge read --session <session_id> --tui
kkachi-agent-bridge status --session <session_id>
kkachi-agent-bridge list
kkachi-agent-bridge stop --session <session_id>
kkachi-agent-bridge stop --session <session_id> --force
```

`send` success is dispatch evidence only. Completion requires bounded observation through `wait` plus final `read` or `status`; retained stream/event observations still require final snapshot reconciliation.

### Pending approval and input

```bash
kkachi-agent-bridge wait --session <session_id> --timeout <seconds>
kkachi-agent-bridge approve --pending <pending_id>
kkachi-agent-bridge reject --pending <pending_id> --reason <reason>
kkachi-agent-bridge answer --pending <pending_id> '<json-answer>'
```

Plan approval is separate from tool/file/command approval. Do not treat `plan approve` as resolving `needs_approval` or `needs_input` pendings.

### Plan lifecycle

```bash
kkachi-agent-bridge plan start --session <session_id> <planning-prompt>
kkachi-agent-bridge plan read --session <session_id>
kkachi-agent-bridge plan approve --session <session_id> --option <option>
kkachi-agent-bridge plan revise --session <session_id> <feedback>
kkachi-agent-bridge plan reject --session <session_id> --reason <reason>
kkachi-agent-bridge plan start-approved --session <session_id>
```

KHS must capture `plan.plan_text`, `plan.plan_state`, `plan.plan_ref`, and `plan.source_evidence` into `plan.md` and/or `bridge-events.md` before implementation starts.

Backend timing rules:

- Claude / GLM / Codex: capture plan while `plan_pending_approval`; approval starts execution.
- Gemini / OpenCode: capture plan while `plan_pending_approval` or `plan_approved_waiting_for_start`; implementation requires explicit `plan start-approved`.
- OpenCode: fail closed on ambiguous plan files or `PLAN_AMBIGUOUS` evidence.

### KAB evidence that KHS must preserve

At minimum, a KAB-backed KHS run records:

- selected backend: `backend_type`, `adapter_type`, tested version, caveats, selection reason;
- bridge session id;
- session `state`, `lifecycle_class`, and `open_pendings` snapshot;
- plan lifecycle state and plan source evidence when plan mode is used;
- observation mode: `cli_loop`, `retained_stream`, or `hybrid`;
- final `read` or `status` evidence after any retained stream/event observation;
- approval, rejection, and answer decisions when pendings occurred.

## Backend capability source of truth

Human-readable SOT:

```text
kkachi-agent-bridge/docs/public/compatibility-matrix.md
```

KHS machine-readable derived view:

```text
kkachi-hermes-skills/registries/cli-capabilities.yaml
```

KHS selection policy:

```text
kkachi-hermes-skills/registries/backend-selection-policy.yaml
```

Selection order is:

1. eliminate backends missing required capabilities;
2. apply project backend policy;
3. apply execution-mode restrictions;
4. apply user preference only after capability and policy gates;
5. apply project-local performance history;
6. select backend prompt profile;
7. record `selected-cli.json` and `capability-check.md`.

## Known reality gaps to keep visible

- Installed KAH on PATH may lag the repo current-candidate; PATH v0.1.1 rejects top-level help and capability discovery, so KHS must verify the effective KAH binary before relying on current-candidate surfaces.
- KAH current-candidate top-level `--help` and `help` are supported; KHS still treats `capabilities --json` plus observed command exits from the effective binary as the activation source.
- KAH current-candidate final gate supports both `gate final <run_id>` and `gate check <run_id> final`; prefer the explicit `gate final <run_id>` form only when effective capabilities include `gate.final` or the equivalent `gate` subcommand `final`, with command-exit fallback for helpers that lack capability discovery.
- KAH does not decide KHS phase applicability; KHS must keep `phase-plan.yaml` and `checklist.md` current.
- Installed KAB on PATH may lag the repo binary; KHS must verify that required commands such as `plan` exist before use.
- Current KAB backend selection is config-driven (`backend_type` in `--config` or default config), not a `send --backend` flag in the observed repo binary.
- KAH validates only core selected-cli and bridge-session fields; KHS owns richer evidence fields.
- OpenCode question-flow support is API/SSE-routed only for real upstream `question.asked` events; rendered `<tool_call>question` text is not evidence.

## Acceptance rule for future changes

A KHS/KAH/KAB interface change is not accepted until all relevant surfaces agree:

1. this document;
2. `registries/phase-contracts.yaml` when phase behavior changes;
3. `registries/cli-capabilities.yaml` when backend capability changes;
4. KHS artifact templates when required evidence changes;
5. KAH docs/tests when deterministic helper commands or schemas change;
6. KAB README/compatibility matrix/tests when bridge commands, evidence, or backend capability changes.
