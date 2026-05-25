# KAS roadmap

Date: 2026-05-21
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record; INITDOC post-KAH reset
Status: post-KAH KAS roadmap; KAH 0.1.4 graph/configurable-feedback substrate evidenced; KAS integration work remains separately gated
Authority level: KAS roadmap; not implementation authorization by itself
Scope: KAS docs/skills planning only; no KAH code, KAB docs, runtime configs, profiles, registries, or gateway changes
Related docs: `README.md`, `sot/initdoc-post-kah-reset.md`, `sot/workflow-graph-integration.md`, `sot/minimum-pilot-cli-lane.md`, `sot/phase-orchestration-policy.md`, `sot/interface-contract.md`
Evidence/source paths:
- Governance evidence record in kanban task `t_2fb00394`
- Blue final synthesis in kanban task `t_3e6d8b89` and Gray docs task `t_1af0dc98`

## Purpose

This roadmap records KAS integration work for KAH-managed project workflow graph and configurable-feedback surfaces. KAH 0.1.4 command/capability evidence exists for `kkachi-agent-helper graph`; KAS integration implementation remains pending until responsible approver authorization and effective-binary capability/help checks exist for the specific run.

## Task sizing policy

- One task is one PR candidate.
- Do not bundle implementation tasks after this docs update.
- Implementation starts only after SOT/spec confirmation, responsible approver authorization, and required risk review / operator workflow review when risk/operator workflow changes apply.
- Each later implementation task must include tests/evidence/verification and update docs only for the changed surface.

## Status values

`Planned`, `In Progress`, `Blocked`, `Completed`, `Deferred`.

## Active roadmap

### EPIC: INITDOC — post-KAH KAS authority reset and MVP roadmap completion

> Goal: remove stale pre-KAH blockers from active KAS documentation and complete an implementation-ready MVP roadmap so KAS work can proceed immediately after the reset. INITDOC is a temporary transition epic; its transition SOT must be deleted after its decisions are absorbed into permanent docs.

| Task ID | Title | Status | Work guide | Notes |
|---|---|---|---|---|
| INITDOC-001 | Create temporary INITDOC transition SOT and roadmap epic | Completed | Create `docs/sot/initdoc-post-kah-reset.md`, define INITDOC scope/exit criteria, and register this epic plus four tasks in `docs/roadmap.md`. | Temporary transition SOT only; it must be removed by INITDOC-004 after absorption into active docs. |
| INITDOC-002 | Reset active docs/SOTs to post-KAH authority | Completed | Update active docs so KAH 0.1.4-evidenced graph/configurable-feedback surfaces are not blocked by stale pre-KAH wording; keep KAB-later, alias, approval, and mutation gates explicit. | Minimum targets updated; Blue verification passed and Red review accepted in `t_b805ce76`. |
| INITDOC-003 | Complete implementation-ready MVP roadmap | Planned | Rework this roadmap into executable post-INITDOC epics/tasks for BOOTSTRAP, CLIMVP, GRAPHMVP, and STALECLEAN with acceptance criteria, deferrals, and evidence requirements. | The roadmap must lead directly into MVP feature development, not another planning loop. |
| INITDOC-004 | Absorb INITDOC decisions and remove transition SOT | Planned | Verify INITDOC decisions are absorbed into permanent docs, delete `docs/sot/initdoc-post-kah-reset.md`, and leave absorbed-target/completion history in this roadmap or PR summary. | Keeping the transition SOT after completion is a failure because it would become new legacy. |

### EPIC: cli — KHS+KAH minimum/pilot CLI lane

> Goal: provide a safe profile-scoped KHS skill-pack installer/doctor/proposal wrapper without turning KHS into a KHC, Doksuri, KAB runtime, or bridge-control surface.

| Task ID | Title | Status | Work guide | Notes |
|---|---|---|---|---|
| cli-001 | Record Blue-confirmed minimum/pilot CLI lane SOT | Completed | Create `docs/sot/minimum-pilot-cli-lane.md`, index it, and mark whole-stack/KAB wording as full execution-runtime guidance only. | Docs-only Gray record from `t_1af0dc98`; Blue confirmation in `t_caee9433` controls final authority for the stated scope. |
| cli-002 | Specify `list`, `install --dry-run`, approved copy install, and `doctor` MVP | Planned | Define manifest/checksum format, changed-path report, target-profile verification, and recovery/rollback instruction before implementation. | Default install mode is copy into `~/.hermes/profiles/<profile>/skills/...`; repo-root Hermes multi-skill-pack install remains unverified. |
| cli-003 | Specify fail-closed `sync` behavior | Planned | Require dry-run, diff, explicit approval, manifest/checksum comparison, backup/recovery path, and no automatic profile/shared-KHS mutation. | High-risk command; do not implement before Red/operator review. |
| cli-004 | Specify `proposal` evidence-only behavior | Planned | Map CLI proposal output to KAH graph proposal records, project overlay proposals, run-local improvement notes, and shared KHS promotion gates. | No automatic mutation of shared KHS, profile skills, project overlays, or KAH graph state without approval/audit evidence. |

### EPIC: graph — KAH-managed workflow graph integration

> Goal: let KHS select workflow graph templates/policy and use KAH's deterministic graph surface without making KHS silently mutate `.kkachi-workflow.yaml` or making KAH decide policy.

| Task ID | Title | Status | Work guide | Notes |
|---|---|---|---|---|
| graph-001 | Create/confirm workflow graph integration SOT and docs index updates | Planned | Confirm `docs/sot/workflow-graph-integration.md`, update docs index/authority ladder, and align phase/interface SOT wording around project graph vs run-local phase state. | KAH `kkachi-agent-helper graph` implementation evidence is recorded; responsible approver confirmation controls any roadmap status promotion; remaining KHS integration work continues in later tasks. |
| graph-002 | Add KHS graph template registry/format for `.kkachi-workflow.yaml` generation | Planned | Define template id/path format, owner metadata, phase/gate/approval declarations, and validation expectations for KAH input. | KHS owns template content and policy; KAH validates and writes/applies. |
| graph-003 | Update orchestration guidance for capability-checked KAH graph use | Planned | Teach KHS guidance to check KAH capabilities/help, then call `kkachi-agent-helper graph init --from-template`, `validate`, and `explain` only when available. | If effective graph support is missing, record a roadmap/feedback gap instead of pretending commands exist. Do not assume a `kah graph` alias. |
| graph-004 | Add declarative graph patch/proposal generation | Planned | Generate declarative graph patches/proposals for KAH `diff`/`propose`; forbid silent direct YAML fallback. | No `init --profile`; no imperative KAH policy-setting commands. |
| graph-005 | Preserve graph evidence in run artifacts and reports | Planned | Store source template, proposal id/path, semantic diff, validation report, approval evidence, graph checksum/version, and audit event ids. | Required when graph changes affect a run. |
| graph-006 | Align KAB context/prompt evidence with applied graph version | Planned | After separate KAB docs update is assigned, record applied graph version/checksum in KAB context/prompt evidence where relevant. | Future/non-authoritative here; KAB is not graph policy authority. |

## Stale/conflict markers

- Older wording that treats `phase-plan.yaml` as the whole workflow SOT is narrowed to run-local execution state/evidence for one run.
- Any wording that requires KAB verification for scoped profile install/injection, `list`, `doctor`, `sync`, or `proposal` planning must be read as full execution-runtime guidance only.
- Any planned `kah graph` alias reference is candidate until alias capabilities/help evidence proves implementation. `kkachi-agent-helper graph` references still require effective-binary capability/help checks before use.
- Kkachi v2 `.kkachi/config/workflows/` is out of scope and must not be merged with or used as fallback for `.kkachi-workflow.yaml`.

## Open questions

- Exact template registry format and compatibility checks remain `graph-002` work.
- Exact KHS artifact mapping for KAH proposal, audit, and checksum evidence remains roadmap work.
- KAB alignment requires a separately assigned KAB docs update.

## Next record action

Responsible approver review closure should compare this roadmap with KAH `docs/roadmap.md` and the KHS/KAH graph SOT files. After closure, create or assign one implementation task at a time.
