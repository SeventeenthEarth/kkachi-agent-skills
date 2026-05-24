# KHS roadmap

Date: 2026-05-21
Owner: KHS workflow/policy layer
Confirming role: Responsible approver / governance evidence record
Status: planning-confirmed KHS roadmap; KAH `kkachi-agent-helper graph` implementation evidence present; KHS integration work remains separately gated
Authority level: KHS workflow-graph integration roadmap; not implementation authorization by itself
Scope: KHS docs/skills planning only; no KAH code, KAB docs, runtime configs, profiles, registries, or gateway changes
Related docs: `README.md`, `sot/workflow-graph-integration.md`, `sot/minimum-pilot-cli-lane.md`, `sot/phase-orchestration-policy.md`, `sot/interface-contract.md`
Evidence/source paths:
- Governance evidence record in kanban task `t_2fb00394`
- Blue final synthesis in kanban task `t_3e6d8b89` and Gray docs task `t_1af0dc98`

## Purpose

This roadmap records KHS integration work for a KAH-managed project workflow graph. KAH graph command support is evidenced for `kkachi-agent-helper graph`; KHS integration implementation remains pending until responsible approver authorization and effective-binary capability/help checks exist for the specific run.

## Task sizing policy

- One task is one PR candidate.
- Do not bundle implementation tasks after this docs update.
- Implementation starts only after SOT/spec confirmation, responsible approver authorization, and required risk review / operator workflow review when risk/operator workflow changes apply.
- Each later implementation task must include tests/evidence/verification and update docs only for the changed surface.

## Status values

`Planned`, `In Progress`, `Blocked`, `Completed`, `Deferred`.

## Active roadmap

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
