# KHS Reality Interface Contract

Date: 2026-05-21
Owner: KHS maintainers
Confirming role: Responsible approver / governance evidence record
Status: post-KAH source of truth for the current usable KAS/KAH/KAB interface; KAH 0.1.4 graph and configurable-feedback support are evidenced when effective-binary capability/help evidence exists; `kah graph` alias remains candidate/unimplemented
Authority level: KAS/KAH/KAB interface SOT
Graph evidence: governance evidence records in kanban tasks `t_2fb00394`, `t_38cfc496`, and `t_2b460665`

## Purpose

This document locks the current practical interface that `kkachi-hermes-skills` (KHS) expects while `kkachi-agent-helper` (KAH) and `kkachi-agent-bridge` (KAB) continue parallel development.

The contract is deliberately reality-first: KHS should drive development through the command surfaces that exist now, not through idealized future surfaces. When KHS needs behavior that KAH or KAB does not yet expose, KHS must record the gap as feedback or a roadmap item instead of pretending the interface exists.

## Layer ownership

```text
KHS:
  Globally/profile-level installed Hermes skill pack; owns Hermes skills, phase contracts,
  task contracts, prompt profiles, backend-selection policy, project overlays, artifact templates,
  semantic guidance packets, and self-improvement rules. It may define profile-scoped
  install/list/doctor/sync/proposal support for the minimum/pilot CLI lane. It does
  not own backend-native inventory truth, KAH deterministic state, or KAB runtime control.

Hermes:
  Manager/orchestrator judgment, KAH/KAB invocation, risk approval routing,
  checklist upkeep, final verification, and Korean user report.

KAH:
  Deterministic project-local `.kkachi/` state, run ids, artifacts, capability cache/evidence
  persistence, events, locks, schemas, diagnostics, and gate checks.

KAB:
  External backend runtime sessions, backend identity, bounded backend-native capability discovery,
  prompt dispatch, wait/read/status, plan lifecycle, approval/input control, retained events,
  raw snapshot/report evidence, and bridge evidence.
```

KAH and KAB validate or execute KHS decisions. They do not choose the KHS workflow, commander, backend lane, phase applicability, or project policy. Capability-checked graph support preserves this boundary: KHS chooses templates/policy/proposals, KAH validates/writes/applies graph state, and KAB remains backend/session/plan evidence rather than graph policy authority.

KAS/KHS provides the early orchestration system: baseline skills, phase
contracts, prompt profiles, artifact templates, project-overlay shape, and
shared improvement-promotion rules. After KAH 0.1.4, graph and
configurable-feedback substrates are capability-evidenced, but active KAS
adoption still matures through KAH/KAB evidence capture and `kkachi-improve`
decisions that route lessons to run artifacts, project overlays, prompt/phase
skill references, scripts, or shared KAS proposals.

## Reality baseline checked

- KAH installed PATH binary checked: `kkachi-agent-helper` from the effective PATH.
- KAH installed PATH version checked on 2026-05-25: `kkachi-agent-helper 0.1.4`.
- KAH installed PATH capability surface checked on 2026-05-25: `capabilities --json` reports supported `project`, `run`, `artifact`, `gate`, `event`, `schema`, `lock`, `diagnostics`, `phase-plan`, `approval`, and `graph` command groups.
- KAH graph surface checked on 2026-05-25: `graph --help` lists `init`, `validate`, `explain`, `diff`, `propose`, `apply`, and `export` and reports status supported.
- KAH compatibility flags checked on 2026-05-25 include `workflow_graph_configurable_feedback_intake=true` and `install_command=false`.
- KAB repo binary checked: `/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-bridge/bin/kkachi-agent-bridge`
- KAB installed PATH binary checked: `/usr/local/bin/kkachi-agent-bridge`
- Important PATH note: the installed `/usr/local/bin/kkachi-agent-bridge` did not expose `plan` in the observed help output, while the repo `bin/kkachi-agent-bridge` did. KHS runs that require plan lifecycle must verify the effective KAB binary before use.

## KHS activation boundary

KHS is active only for explicit or context-bound Kkachi-governed software runs:

- the master names KHS, Kkachi, KAH, or KAB;
- the master asks to apply KHS to a project;
- the master requests a Kkachi run;
- the master assigns development to a KHS-using execution owner;
- the work requires bridge-backed implementation, durable KAH artifacts, gates, red-team review, or backend evidence.

Simple direct Hermes edits, typo fixes, small one-file config patches, and read-only explanations stay outside KHS unless explicitly moved into KHS.

Once KHS is activated for a full execution-runtime run, KAH is mandatory for deterministic run state. KHS code-change and development runs use KAB. If the master requests code changes without KAB, treat the task as a normal direct Hermes task rather than a KHS run.

The scoped KHS+KAH minimum/pilot harness lane is not a full execution-runtime run. It may cover profile-scoped KHS skill install/list/doctor/sync/proposal support through a future `kkachi-hermes-skills` CLI without requiring KAB verification, but it must not claim `run`, backend session, bridge control, KHC command, Doksuri integration, or KAB substitute authority.

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

KAB may supply backend-authored plan text, events, status snapshots, raw capability snapshots, refresh reports, and output. KHS/Hermes must preserve those into KAH artifacts before treating them as run evidence.

## Capability storage and semantic guidance boundary

Capability records follow this locked authority split:

- Backend-native sources are the inventory source of truth for refresh. Examples include backend project/user command or skill directories, native agent/app/plugin catalogs, slash-command evidence, backend help/source fingerprints, and installed Hermes/KHS locations when the Hermes backend itself is being inspected.
- KAB owns raw bounded discovery/verification and produces `capability_snapshot` / `capability_refresh_report` records with source fingerprints, scan boundaries, layered cache keys, raw callability states, and `snapshot_required` / `user_next_step` fail-closed output when needed.
- KAH owns project-local `.kkachi/` persistence/evidence/audit after project init. Expected paths are `.kkachi/capabilities/current.json`, `snapshots/<id>.json`, `reports/<id>.json`, `fingerprints/<id>.json`, `drift/<id>.json`, and run-local `capability-snapshot.json` / `capability-check.md`.
- KHS owns semantic guidance and prompt composition only: purpose text, usage examples, caveats, fallback wording, trust labels, and review-gated enrichment records. KHS semantic guidance cannot create, upgrade, or imply backend callability.
- The responsible approver owns final active prompt selection from raw KAB evidence plus labeled KHS semantic guidance. Promotion from `semantic_draft` to active guidance requires review-gated evidence refs.

`capability list` may normally read `.kkachi/capabilities/current.json`, but `capability refresh` must not trust `.kkachi/` alone. Refresh scans or inspects bounded backend-native sources for the effective workspace/profile/backend/project scope, then KAB generates raw snapshot/report records and KAH persists them. Snapshot/cache keys stay layered by workspace, profile, backend, project overlay, and session/runtime fields where applicable.

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

Reality note: the KAH 0.1.4 PATH baseline exposes conventional top-level `--help`, `help`, command-specific help, and `capabilities --json` surfaces. KAS must still use capability discovery and observed command exits from the effective binary as the machine activation source; help output is supplemental documentation, not a version-guess substitute.

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

KAH run metadata is helper classification. It seeds deterministic requirements but does not decide which KHS phases execute. `phase-plan.yaml` remains the run-local KHS workflow/execution state for one run, not project-level graph authority.

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

Reality correction: the KAH 0.1.4 PATH baseline supports both `gate check <run_id> final --json` and the dedicated final gate shortcut `gate final <run_id> --json`. KAS operator-facing guidance should use `gate final <run_id> --json` as the canonical final gate. `gate check <run_id> final --json` is a compatibility fallback for older effective helpers that lack the shortcut; fallbacks must be based on capabilities and command-exit evidence, not guessed from a version string.

### KAH schema boundary

KAH 0.1.4 validates core fields for the bridge-selection artifacts:

`selected-cli.json` required core:

```json
["version", "status", "backend_type", "adapter_type", "source_ledger_ref", "caveats"]
```

`bridge-session-snapshot.json` required core:

```json
["version", "session_id", "backend_type", "adapter_type", "state", "lifecycle_class", "open_pendings"]
```

KHS templates may include richer fields such as tested CLI, capability snapshot, prompt profile, observation mode, plan lifecycle, and accepted caveats. Those fields are KHS-level evidence and must remain compatible with KAH's core schema.

### KAH graph interface and `kah` alias boundary

Status: implemented for the real `kkachi-agent-helper graph` command surface when the effective KAH binary advertises it through capabilities/help and command-exit evidence. If the real command remains `kkachi-agent-helper`, `kah graph` is shorthand and does not prove a `kah` alias exists.

Capability-checked command surface:

```text
kkachi-agent-helper graph init --from-template <template-id-or-path> [--output .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph validate [--file .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph explain [--file .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph diff --from <file-or-ref> --to <file-or-ref> [--semantic] [--json]
kkachi-agent-helper graph propose --candidate-file <repo-relative-candidate-graph> --reason <text> [--json]
kkachi-agent-helper graph propose --patch <patch-file> --reason <text> [--json]  # legacy compatibility alias
kkachi-agent-helper graph apply --proposal <proposal-id> --approval <evidence-ref> [--json]
kkachi-agent-helper graph export --format mermaid|plantuml [--output <path>] [--json]
```

Boundary rules:

- `.kkachi-workflow.yaml` is project-level graph state when backed by KAH graph init/validation/apply evidence.
- `.kkachi/runs/<run_id>/phase-plan.yaml` remains run-local execution state/evidence.
- `.kkachi/config.yaml` remains KAH helper config only and is never workflow graph authority.
- Kkachi v2 `.kkachi/config/workflows/` is out of scope and must not be used as fallback/merge input.
- KHS chooses templates/policy/proposals; KAH validates, explains, diffs, proposes, applies approved graph state, and records audit evidence.
- KHS must refuse silent direct YAML fallback when KAH graph support is missing.
- KAB plan/session evidence remains separate and is not graph policy authority.

Graph evidence KHS must preserve when a graph mutation affects a run:

- source template id/path;
- proposal id/path;
- semantic diff;
- validation report;
- approval/audit evidence;
- applied graph checksum/version;
- KAH graph audit event ids;
- capability check proving the graph surface existed.

Command classification: `init --from-template`, `validate`, `explain`, `diff`, `propose`, `apply`, and `export` contain zero imperative policy-mutation commands. Do not document `kkachi-agent-helper workflow ...`, `kkachi-agent-helper graph init --profile ...`, `kkachi-agent-helper gate set ...`, `kkachi-agent-helper review-policy set ...`, or `kkachi-agent-helper graph set-policy ...` as normal KHS guidance.

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

- KAS must verify the effective KAH binary before each controlled run; the current PATH baseline is KAH 0.1.4 with top-level help, capability discovery, `gate final`, `phase-plan`, `approval`, `graph`, and configurable-feedback support.
- KAS still treats `capabilities --json` plus observed command exits from the effective binary as the activation source; help output remains supplemental documentation.
- Final gate guidance is canonicalized on `gate final <run_id> --json`; `gate check <run_id> final --json` remains only a compatibility fallback for older helpers that lack the dedicated shortcut.
- KAH does not decide KHS phase applicability; KHS must keep `phase-plan.yaml` and `checklist.md` current.
- Installed KAB on PATH may lag the repo binary; KHS must verify that required commands such as `plan` exist before use.
- Current KAB backend selection is config-driven (`backend_type` in `--config` or default config), not a `send --backend` flag in the observed repo binary.
- KAH validates only core selected-cli and bridge-session fields; KHS owns richer run evidence fields, but not raw backend-native inventory truth.
- Capability snapshot/list/refresh support is docs/design-locked but not accepted as implemented until KAB/KAH capabilities, schemas, help, and command exits prove it.
- KHS semantic catalogs/guidance do not prove backend callability; they must reference raw KAB snapshot evidence and remain review-gated before active prompt use.
- KAH graph support remains effective-binary gated: use `kkachi-agent-helper graph` only when capabilities/help prove it; when absent, KHS records a gap instead of directly editing `.kkachi-workflow.yaml` as fallback. `kah graph` alias behavior remains candidate until separately evidenced.
- The KHS+KAH minimum/pilot CLI lane is Blue-confirmed for its stated scope and limited to install/list/doctor/sync/proposal support; any wording that turns it into KHC, Doksuri, KAB runtime, bridge-control, or code-change run authority is stale/conflicting.
- KAB graph alignment is future/non-authoritative; KAB evidence must remain separate from graph policy authority.
- Kkachi v2 `.kkachi/config/workflows/` is outside KAH/KHS graph scope and must not be used as fallback graph policy.
- OpenCode question-flow support is API/SSE-routed only for real upstream `question.asked` events; rendered `<tool_call>question` text is not evidence.

## Acceptance rule for future changes

A KHS/KAH/KAB interface change is not accepted until all relevant surfaces agree:

1. this document;
2. `registries/phase-contracts.yaml` when phase behavior changes;
3. `registries/cli-capabilities.yaml` when backend capability changes;
4. KHS artifact templates when required evidence changes;
5. KAH docs/tests when deterministic helper commands or schemas change;
6. KAB README/compatibility matrix/tests when bridge commands, evidence, or backend capability changes.
7. KAH `.kkachi/` layout/schema docs when project-local persistence or evidence paths change.
8. KHS semantic-guidance docs/templates when prompt guidance consumes or enriches capability snapshots.

## Capability storage decision appendix

Date: 2026-05-21
Owner: KHS/KAH/KAB interface archive
Confirming role: Responsible approver / governance evidence record
Status: docs/design lock; implementation remains separately gated
Scope: KHS/KAH/KAB capability storage and authority split
Decision summary: KHS is global/profile-level semantic/workflow guidance, KAB is raw bounded backend-native discovery, KAH persists `.kkachi/` project-local cache/evidence, and the responsible approver selects final active prompt guidance.
Stale/conflict markers: `.kkachi/` is cache/evidence only, not native inventory SOT; KHS semantic catalog entries cannot prove callability; first-run/stale state must fail closed with `snapshot_required` / `user_next_step`.
Next record action: risk review, operator workflow review, and record review of fail-open risk, operator UX, and SOT/evidence consistency before implementation authorization.

## Graph record appendix

Date: 2026-05-21
Owner: KHS/KAH/KAB interface archive
Confirming role: Responsible approver / governance evidence record
Status: graph interface addition to current reality-first interface SOT; `kkachi-agent-helper graph` implemented, `kah graph` alias candidate
Authority level: interface SOT for existing surfaces; graph section is active only after effective KAH capabilities/help evidence
Scope: KHS/KAH/KAB interface docs only; no KAB docs update in this task
Related docs: `workflow-graph-integration.md`, `phase-orchestration-policy.md`, `../roadmap.md`, KAH `docs/specs.md`, KAH `docs/compatibility.md`
Decision summary: KHS must capability-check `kkachi-agent-helper graph` support, preserve graph mutation evidence, refuse silent direct YAML fallback, and keep KAB plan/session evidence separate from graph policy authority. `kah graph` alias references remain candidate until alias evidence exists.
Evidence/source paths: governance evidence records in kanban tasks `t_2fb00394`, `t_38cfc496`, and `t_2b460665`
Stale/conflict markers: `kah graph` alias references are planned/candidate; Kkachi v2 `.kkachi/config/workflows/` and KAB evidence are not graph authority.
Open questions: KHS artifact mapping for applied graph version/checksum and KAB graph-version propagation remain future tasks.
Next record action: update KHS guidance/registries/templates only through separately assigned implementation tasks.
