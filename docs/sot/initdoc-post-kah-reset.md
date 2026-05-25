# INITDOC post-KAH KAS authority reset

Date: 2026-05-25
Owner: KAS documentation and enablement lane
Confirming role: 주군 approval / Blue execution command; Red, Orange, and Gray review through Kanban when assigned
Status: temporary transition SOT for INITDOC only
Authority level: transition authority for post-KAH document reset and MVP-roadmap completion; not permanent KAS behavior SOT
Lifecycle: delete after INITDOC completion once all decisions are absorbed into active SOTs, roadmap, docs index, skills/templates/registries where needed, and the completion record names those absorbed targets
Scope: `kkachi-hermes-skills` documentation reset and KAS MVP readiness planning only; no KAH code, KAB docs, gateway credentials, profile secrets, or production runtime changes
Related docs: `../roadmap.md`, `../README.md`, `interface-contract.md`, `workflow-graph-integration.md`, `minimum-pilot-cli-lane.md`, `external-feedback-intake-policy.md`, `khs-pre-kah-readiness-audit-2026-05.md`
Evidence/source paths:
- KAH effective binary check on 2026-05-25: `kkachi-agent-helper --version` => `kkachi-agent-helper 0.1.4`
- KAH capability check on 2026-05-25: `kkachi-agent-helper capabilities --json` reports `graph` supported with `init`, `validate`, `explain`, `diff`, `propose`, `apply`, and `export`; `workflow_graph_configurable_feedback_intake=true`; `install_command=false`
- KAH graph help check on 2026-05-25: `kkachi-agent-helper graph --help` reports graph status supported
- KAS repo bootstrap check on 2026-05-25: `kkachi-agent-helper project doctor --json` fails because `.kkachi/config.yaml`, `.kkachi/status.json`, `.kkachi/events.jsonl`, and schema files are missing

## 1. Purpose

INITDOC exists to cut KAS loose from stale pre-KAH planning language now that KAH 0.1.4 is available. Its output must make KAS development immediately executable instead of repeatedly stopping at old `candidate`, `planned`, or `blocked_by_kah` wording.

INITDOC has two equal goals:

1. Reset active documents so current KAH capabilities are reflected accurately.
2. Complete a usable MVP roadmap that leads directly into implementation tasks.

INITDOC is not a new permanent operating model. It is a temporary transition scaffold used to update the real authorities, then removed.

## 2. Naming and authority convention

- Use `KAS` for the current Kkachi Skills layer name.
- Preserve `KHS` only when referring to historical file names, legacy document wording, or repository-local names that have not yet been renamed.
- Treat `kkachi-hermes-skills` as the current repository path/name, not as proof that the behavior layer must keep old KHS-era status labels.
- KAS owns skill packs, phase/process guidance, templates, registries, prompt/profile guidance, task/report contracts, and MVP CLI semantics.
- KAH owns deterministic project-local state, graph validation/write/apply/diff/propose/export, schemas, events, locks, gates, diagnostics, and evidence persistence.
- KAB owns backend runtime/session, plan/question/approval/input handling, retained events, and backend execution evidence for full execution-runtime lanes.

## 3. Post-KAH baseline

The current baseline is:

- KAH graph support is implemented when the effective binary advertises it through capabilities/help and command execution evidence.
- KAH support does not automatically mean KAS has already adopted that support in docs, templates, skills, registries, or CLI behavior.
- KAS repo project state is not bootstrapped yet. `project doctor` currently fails because `.kkachi/` project files are missing.
- Hermes/KAS skill installation is not a KAH command. KAH explicitly omits `install`; KAS must define its own safe profile-scoped installation path or use verified Hermes-native tooling where appropriate.
- KAB-dependent runtime claims remain `kab_later` until KAB evidence exists.

Therefore, the preferred status split after INITDOC is:

| Old pre-KAH wording | Post-KAH replacement |
|---|---|
| `blocked_by_kah` for graph features now advertised by KAH 0.1.4 | `kah-evidenced, kas-integration-pending` |
| `candidate` for confirmed ownership/boundary rules | `current SOT` after confirmation |
| `planned` for the MVP path | `ready-for-implementation` once acceptance criteria and evidence are listed |
| pre-KAH readiness blocker | `historical` or `superseded` audit input |
| runtime/backend evidence without KAB | keep `kab_later` |
| `kah graph` alias | keep `candidate` unless alias evidence exists |

## 4. INITDOC epic tasks

INITDOC must be recorded in `docs/roadmap.md` as an epic with the following tasks. Each task is one PR candidate unless 주군 explicitly approves bundling.

| Task ID | Title | Status target | Required outcome |
|---|---|---|---|
| INITDOC-001 | Create temporary INITDOC transition SOT and roadmap epic | Ready first | This file exists, defines scope/exit criteria, and `docs/roadmap.md` records the INITDOC epic and four tasks. |
| INITDOC-002 | Reset active docs/SOTs to post-KAH authority | Ready after 001 | Active docs no longer use stale pre-KAH wording as blockers where KAH 0.1.4 evidence exists; remaining KAB or alias gaps stay explicitly gated. |
| INITDOC-003 | Complete implementation-ready MVP roadmap | Ready after 002 | `docs/roadmap.md` contains executable post-INITDOC epics/tasks for bootstrap, CLI MVP, graph MVP, and stale-surface cleanup. |
| INITDOC-004 | Absorb INITDOC decisions and remove temporary SOT | Final | All transition decisions are absorbed into permanent docs; this file is deleted; roadmap keeps INITDOC completion history and absorbed-target list. |

## 5. Documents to reset

INITDOC-002 must review and update at least these files:

- `docs/README.md`
- `docs/roadmap.md`
- `docs/sot/interface-contract.md`
- `docs/sot/workflow-graph-integration.md`
- `docs/sot/external-feedback-intake-policy.md`
- `docs/sot/minimum-pilot-cli-lane.md`
- `docs/sot/khs-pre-kah-readiness-audit-2026-05.md`
- `docs/sot/khs-delegation-packet-and-report-contract.md` if its status labels still overstate KAH blockage after KAH 0.1.4 evidence
- repository `README.md` if it sends operators back into superseded pre-KAH interpretation

INITDOC-002 must not silently delete historical context. Historical records may remain if they are clearly marked as historical/superseded and no longer block post-KAH execution.

## 6. MVP roadmap to complete

INITDOC-003 must turn the roadmap into a launch path for usable KAS. The expected post-INITDOC roadmap shape is:

### EPIC: BOOTSTRAP — KAS repo KAH project bootstrap

Goal: make `kkachi-hermes-skills` a valid KAH-managed project so KAS can record deterministic evidence.

Minimum tasks:

- `BOOTSTRAP-001`: Run/record KAH `project init` for this repo after approval.
- `BOOTSTRAP-002`: Make `kkachi-agent-helper project doctor --json` pass and record evidence.
- `BOOTSTRAP-003`: Decide whether generated `.kkachi/` state is committed, ignored, or handled as local operator state according to repo policy.

### EPIC: CLIMVP — profile-scoped KAS skill-pack install/list/doctor

Goal: let a user safely install and verify KAS skills in a Hermes profile without KAB runtime readiness.

Minimum tasks:

- `CLIMVP-001`: Specify command surface and manifest/checksum format for `list`, `install --dry-run`, approved install, and `doctor`.
- `CLIMVP-002`: Implement list/discovery over KAS skill packs/categories.
- `CLIMVP-003`: Implement dry-run install with planned changed-path report.
- `CLIMVP-004`: Implement approved copy install with manifest/checksum and rollback/recovery instructions.
- `CLIMVP-005`: Implement doctor checks for source pack, target profile, installed skills, KAH availability, and project bootstrap state.

Deferred from MVP unless separately approved: `sync`, broad `proposal`, symlink mode, shared `skills.external_dirs` default, KHC/Doksuri integration, and KAB run/control verbs.

### EPIC: GRAPHMVP — KAS graph template registry and KAH graph adoption

Goal: let KAS provide workflow graph templates/policy and use KAH graph surfaces without direct YAML fallback.

Minimum tasks:

- `GRAPHMVP-001`: Define KAS graph template registry schema/format and template id rules.
- `GRAPHMVP-002`: Add a default KAS workflow graph template for `.kkachi-workflow.yaml` generation.
- `GRAPHMVP-003`: Add capability-checked guidance for `kkachi-agent-helper graph init --from-template`, `validate`, and `explain`.
- `GRAPHMVP-004`: Define evidence preservation for template id, graph checksum/version, validation report, proposal id/path, approval/audit evidence, and semantic diff when graph changes affect a run.

Deferred from MVP unless separately approved: graph `apply` automation, declarative graph proposal generation, KAB prompt/context alignment, and `kah graph` alias assumptions.

### EPIC: STALECLEAN — remove stale pre-KAH blockers from active KAS surfaces

Goal: prevent active skills/templates/registries from reintroducing old pre-KAH stops after docs are reset.

Minimum tasks:

- `STALECLEAN-001`: Audit `skills/*/SKILL.md`, `registries/*`, `templates/*`, and repository `README.md` for stale pre-KAH wording.
- `STALECLEAN-002`: Update active skill guidance to check effective KAH capabilities and proceed when supported.
- `STALECLEAN-003`: Keep unsupported KAB-dependent behavior explicitly `kab_later` without blocking CLIMVP/GRAPHMVP.
- `STALECLEAN-004`: Verify no active guidance still treats KAH graph support or configurable feedback intake as absent when KAH 0.1.4 evidence is present.

## 7. Non-goals and safety gates

INITDOC must not:

- implement CLI code while the transition SOT is still the only authority;
- mutate shared Hermes profiles, gateway config, credentials, or auth tokens;
- claim KAB runtime support from KAH evidence;
- treat KAH `send`, graph help, or project doctor output as backend execution evidence;
- assume a `kah graph` alias without capability/help evidence;
- use direct `.kkachi-workflow.yaml` edits as the normal fallback path;
- erase history without replacing it with a named current authority.

Risk gates that remain after post-KAH reset:

- write-capable install requires dry-run, changed-path report, explicit approval, manifest/checksum, and recovery/rollback instruction;
- `sync` remains high-risk and needs Red/operator review before implementation;
- graph mutation/apply remains proposal/approval/audit gated;
- KAB runtime alignment remains separate and later unless 주군 explicitly assigns it.

## 8. Completion and deletion criteria

INITDOC is complete only when:

1. `docs/roadmap.md` records INITDOC and post-INITDOC MVP epics/tasks.
2. Active SOTs and docs index reflect KAH 0.1.4 as current evidence where appropriate.
3. Pre-KAH audit documents are marked historical/superseded or narrowed so they no longer block KAS MVP implementation.
4. The MVP roadmap includes BOOTSTRAP, CLIMVP, GRAPHMVP, and STALECLEAN with clear task boundaries and deferrals.
5. A verification pass confirms active docs do not contain stale `blocked_by_kah` claims for KAH-supported graph/configurable-feedback surfaces.
6. The decisions in this transition SOT are absorbed into the permanent docs listed in INITDOC-004.

After those conditions are met, INITDOC-004 must delete this file and leave an absorbed-target record in `docs/roadmap.md` or the PR summary. Keeping this file after completion is a failure because INITDOC is meant to remove transition scaffolding, not create new legacy.

## 9. Immediate next action

After this file is created, update `docs/roadmap.md` to add the INITDOC epic and tasks, then proceed with INITDOC-002 document reset before any CLIMVP or GRAPHMVP implementation begins.
