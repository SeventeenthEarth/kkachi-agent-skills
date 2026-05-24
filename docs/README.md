# KHS documentation layers

Date: 2026-05-21
Owner: KHS documentation archive
Confirming role: Responsible approver / governance evidence record
Status: docs index / authority ladder; workflow graph entries distinguish implemented `kkachi-agent-helper graph` evidence from remaining KHS integration and `kah graph` alias work; external-feedback entries distinguish candidate/planned policy records from operational support claims
Authority level: source of truth for how to read KHS docs after confirmation
Scope: `kkachi-hermes-skills/docs` only
Evidence/source path: governance evidence record in kanban task `t_2fb00394`

This directory separates durable KHS sources of truth from working discussion notes. It also distinguishes capability-evidenced `kkachi-agent-helper graph` surfaces from KHS integration work, shorthand alias behavior that still requires separate evidence, and candidate/planned external-feedback policy records that must not be confused with operational support. The minimum/pilot CLI lane is recorded separately so install/profile-injection support is not confused with the full KHS+KAH+KAB execution-runtime path.

## Authority ladder

| Path | Meaning | Authority |
|---|---|---|
| `sot/interface-contract.md` | Current KHS/KAH/KAB interface SOT | Authoritative for reality-first interface guidance; `kkachi-agent-helper graph` use remains effective-binary capability-checked, `kah graph` alias remains candidate, and external-feedback bounds remain candidate until separate KAH evidence exists |
| `sot/khs-architecture-and-integration.md` | Broad KHS architecture/usage/KAH-integration SOT | Blue-reviewed candidate development base accepted in `t_f29b6ee9`; distinguishes target architecture from operational support claims |
| `sot/khs-pre-kah-readiness-audit-2026-05.md` | Corrective pre-KAH readiness audit | Candidate docs-only audit from `t_2e9d918a`; separates external-feedback child completion, whole KHS pre-KAH readiness, and operational/runtime support readiness |
| `sot/khs-delegation-packet-and-report-contract.md` | Candidate Delegation Packet and operator report contract | Candidate docs-only contract; records packet/report fields and status labels without claiming KAH artifact or KAB runtime support |
| `sot/external-feedback-intake-policy.md` | Candidate child policy for configurable `EXTERNAL_FEEDBACK_INTAKE` bounds | Candidate/planned pre-KAH record; current MVP is user-supplied `feedback.md`; min=1/max=5 operation is `blocked_by_kah` and automated review-by-different-tool is `kab_later` |
| `sot/phase-orchestration-policy.md` | Current KHS phase/run orchestration policy | Authoritative for run-local phase behavior |
| `sot/workflow-graph-integration.md` | Planning-confirmed KHS/KAH graph integration SOT | Confirmed KHS/KAH ownership and evidence rules for implemented `kkachi-agent-helper graph`; remaining KHS integration tasks and external-feedback bounds stay separately gated |
| `sot/minimum-pilot-cli-lane.md` | Blue-confirmed KHS+KAH minimum/pilot CLI lane SOT | Confirmed lane split and safety constraints for profile-scoped install/list/doctor/sync/proposal support; does not replace full KHS+KAH+KAB execution-runtime authority |
| `sot/concept.md` | KHS concept SOT | Durable background; not changed by this graph docs pass |
| `sot/architecture-understanding.md` | KHS architecture understanding | Durable background; not changed by this graph docs pass |
| `sot/skill-template.md` | Skill template guidance | Durable background; not changed by this graph docs pass |
| `roadmap.md` | KHS graph integration roadmap | Candidate active roadmap for graph integration tasks |
| `discussions/*` | Temporary or evolving discussion material | Historical/non-authoritative unless promoted into `sot/`, registries, templates, or skills |

## `sot/`

Durable source-of-truth documents for KHS behavior, architecture, and skill design. These documents should be updated deliberately because they define the shared operating model.

- `sot/concept.md`
- `sot/architecture-understanding.md`
- `sot/skill-template.md`
- `sot/khs-architecture-and-integration.md` — candidate broad SOT for KHS purpose, Hermes usage, KHS/KAH/KAB/KHC boundaries, repo structure, operator report contract, self-improvement governance, and development acceptance gates; external feedback intake is summarized there and detailed by `sot/external-feedback-intake-policy.md`
- `sot/khs-pre-kah-readiness-audit-2026-05.md` — candidate docs-only readiness audit for `t_2e9d918a`; keeps scoped external-feedback completion separate from whole KHS pre-KAH readiness and operational/runtime support readiness
- `sot/khs-delegation-packet-and-report-contract.md` — candidate docs-only Delegation Packet and report contract; defines packet/report fields, evidence paths, and status labels without upgrading them to KAH/KAB support
- `sot/external-feedback-intake-policy.md` — candidate/planned child policy for configurable `EXTERNAL_FEEDBACK_INTAKE` bounds; records min=1/max=5 target semantics, user-supplied `feedback.md` MVP intake, KAH-blocked fail-closed cases, KAB-later automation, status labels, report fields, and a stale-surface manifest
- `sot/phase-orchestration-policy.md`
- `sot/interface-contract.md`
- `sot/workflow-graph-integration.md` — candidate SOT for `.kkachi-workflow.yaml` / capability-checked `kkachi-agent-helper graph` integration; `kah graph` alias and external-feedback bounds remain planned/candidate until separately evidenced
- `sot/minimum-pilot-cli-lane.md` — Blue-confirmed SOT for the KHS+KAH minimum/pilot CLI lane; keeps profile-scoped install/list/doctor/sync/proposal support separate from full KHS+KAH+KAB execution-runtime authority

## `roadmap.md`

KHS workflow-graph integration roadmap. It uses PR-candidate task sizing and is not implementation authorization by itself. KAH graph support is now evidenced for `kkachi-agent-helper graph`; remaining KHS implementation, `kah graph` alias behavior, and configurable external-feedback bounds must still proceed one task at a time after SOT/spec confirmation, responsible approver authorization, and current KAH command/code/test or effective-binary capability evidence for the specific surface.

## `discussions/`

Temporary or evolving discussion material that should not be treated as final policy by itself. Items here may later be promoted into `sot/`, registries, templates, or skills after review and evidence.

- `discussions/follow-up-discussion.md`
- `discussions/feedback-for-KAH.md`
- `discussions/feedback-for-KAB.md`

## Status vocabulary

| Status | Meaning |
|---|---|
| `source of truth` | Confirmed current authority for its stated scope |
| `candidate SOT` | Proposed normative record pending confirmation and/or implementation evidence |
| `planned/candidate` | Roadmap task, proposed command surface, or shorthand alias not proven implemented |
| `candidate` | Proposed KHS/KAH/KAB behavior recorded for review; not current support |
| `planned` | Expected future work with no current operational guarantee |
| `implemented` | Proven by current command/code/docs/test or effective-binary capability evidence for the exact surface claimed |
| `blocked_by_kah` | KHS can describe the desired policy, but KAH lacks required schema, validation, proposal, audit, or compatibility support |
| `kab_later` | Requires KAB backend/runtime evidence and is outside current MVP |
| `unsupported` | The current toolchain cannot safely execute or validate the requested support |
| `historical` | Preserved context; not current authority by itself |
| `stale` | Known to conflict with newer evidence or decisions; preserve with marker rather than silently delete |
| `superseded` | Replaced by a named newer authority |

## Promotion rule

A discussion note becomes SOT only after the decision is reflected in the relevant `sot/` document, registry, template, or skill file. A candidate SOT becomes project authority only after responsible approver confirmation is recorded; implementation status still requires capability/help evidence.

## Decision summary

- `.kkachi-workflow.yaml` is a project-level workflow graph file when backed by KAH graph evidence.
- `.kkachi/runs/<run_id>/phase-plan.yaml` remains run-local execution state/evidence for one KHS run.
- KHS owns graph templates, policy selection, phase applicability, and proposal content.
- KAH owns deterministic validation/write/apply/diff/audit behavior for evidenced `kkachi-agent-helper graph` surfaces; configurable external-feedback bounds require separate KAH schema/compatibility evidence before KHS may label them implemented.
- The KHS+KAH minimum/pilot CLI lane is limited to profile-scoped install/list/doctor/sync/proposal support and does not claim KHC, Doksuri, KAB run/control, or bridge-runtime authority.
- KAB remains backend/session/plan evidence, not graph policy authority.
- Kkachi v2 `.kkachi/config/workflows/` is out of KAH/KHS graph scope and has no fallback/merge relationship with `.kkachi-workflow.yaml`.

## Stale/conflict markers

Older wording that calls `phase-plan.yaml` the whole workflow SOT must be read narrowly as run-local execution state/evidence. Any `kah graph` reference in these docs is planned/candidate unless alias capability/help evidence proves implementation. `kkachi-agent-helper graph` references may be treated as implemented only after effective KAH capability/help checks. Configurable external-feedback bounds remain candidate/planned until their own KAH evidence exists.

## Open questions

- Exact KHS template registry format for graph generation remains roadmap work.
- Exact KHS artifact mapping for KAH graph proposal paths, checksums, and audit event ids remains roadmap work; KAH graph proposal/apply surfaces are implemented, but external-feedback bounds require separate KAH evidence before any support claim.
- KAB alignment requires separate KAB docs work and is non-authoritative here.

## Next record action

Responsible approver review closure should compare the graph SOT files and roadmap. Future KHS implementation tasks should update only the docs sections affected by the implemented KHS surface and preserve evidence paths.
