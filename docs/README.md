# KHS documentation layers

Date: 2026-05-21
Owner: KHS documentation archive
Confirming role: Responsible approver / governance evidence record
Status: post-KAH docs index / authority ladder; KAH 0.1.4 graph and configurable-feedback capability evidence is recognized, CLIMVP and GRAPHMVP-001..003 KAS guidance surfaces are implemented, and KAB runtime work remains separately gated
Authority level: source of truth for how to read KHS docs after confirmation
Scope: `kkachi-hermes-skills/docs` only
Evidence/source path: governance evidence record in kanban task `t_2fb00394`

This directory separates durable KAS/KHS sources of truth from working discussion notes. It also distinguishes capability-evidenced KAH 0.1.4 `kkachi-agent-helper graph` and configurable-feedback surfaces from remaining KAS adoption work, shorthand alias behavior that still requires separate evidence, and KAB-dependent runtime claims. The minimum/pilot CLI lane is recorded separately so install/profile-injection support is not confused with the full KAS+KAH+KAB execution-runtime path.

## Authority ladder

| Path | Meaning | Authority |
|---|---|---|
| `sot/interface-contract.md` | Current KAS/KAH/KAB interface SOT | Authoritative for reality-first interface guidance; KAH 0.1.4 `kkachi-agent-helper graph` and configurable-feedback support are capability-evidenced, remaining KAS adoption is integration-pending, `kah graph` alias remains candidate, and KAB runtime claims remain later |
| `sot/khs-architecture-and-integration.md` | Broad KHS architecture/usage/KAH-integration SOT | Blue-reviewed candidate development base accepted in `t_f29b6ee9`; distinguishes target architecture from operational support claims |
| `sot/khs-pre-kah-readiness-audit-2026-05.md` | Historical pre-KAH readiness audit | Superseded audit input from `t_2e9d918a`; no longer blocks post-KAH KAS MVP work after INITDOC closure, but still preserves overclaim and lineage warnings |
| `sot/khs-delegation-packet-and-report-contract.md` | Candidate Delegation Packet and operator report contract | Candidate docs-only contract; records packet/report fields and status labels without claiming KAH artifact or KAB runtime support |
| `sot/external-feedback-intake-policy.md` | KAH-evidenced policy SOT for configurable `EXTERNAL_FEEDBACK_INTAKE` bounds | KAH 0.1.4 advertises configurable feedback intake capability; KAS registry/template/skill adoption remains integration-pending; current MVP still uses user-supplied `feedback.md`; automated review-by-different-tool remains `kab_later` |
| `sot/phase-orchestration-policy.md` | Current KHS phase/run orchestration policy | Authoritative for run-local phase behavior |
| `sot/workflow-graph-integration.md` | Current KAS/KAH graph integration SOT | Confirmed KAS/KAH ownership, capability preflight, fail-closed guidance, and evidence rules for implemented KAH 0.1.4 `kkachi-agent-helper graph`; `kas-default` template exists and `kah graph` alias stays separately gated |
| `sot/graph-template-registry.md` | Current KAS graph template registry/default-template SOT | Defines graph template id/path/version/owner metadata, phase/edge/gate/approval/feedback-intake fields, KAH validation expectations, registry evidence requirements, and active `kas-default` template metadata |
| `sot/minimum-pilot-cli-lane.md` | Blue-confirmed KHS+KAH minimum/pilot CLI lane SOT | Confirmed lane split and safety constraints for profile-scoped install/list/doctor/sync/proposal support; does not replace full KHS+KAH+KAB execution-runtime authority |
| `sot/kas-cli-contract.md` | CLIMVP-001 KAS minimum CLI contract SOT | Accepted command-surface and manifest/checksum contract for profile-scoped `list`, `install --dry-run`, approved copy install, and `doctor`; CLIMVP-001 through CLIMVP-005 are implemented and closed, while `sync` and broad `proposal` remain gated future surfaces |
| `sot/concept.md` | KHS concept SOT | Durable background; not changed by this graph docs pass |
| `sot/architecture-understanding.md` | KHS architecture understanding | Durable background; not changed by this graph docs pass |
| `sot/skill-template.md` | Skill template guidance | Durable background; not changed by this graph docs pass |
| `roadmap.md` | KAS post-KAH MVP roadmap | Active roadmap for INITDOC completion history and BOOTSTRAP / CLIMVP / GRAPHMVP / STALECLEAN PR-candidate tasks |
| `discussions/*` | Temporary or evolving discussion material | Historical/non-authoritative unless promoted into `sot/`, registries, templates, or skills |

## `sot/`

Durable source-of-truth documents for KHS behavior, architecture, and skill design. These documents should be updated deliberately because they define the shared operating model.

- `sot/concept.md`
- `sot/architecture-understanding.md`
- `sot/skill-template.md`
- `sot/khs-architecture-and-integration.md` — candidate broad SOT for KHS purpose, Hermes usage, KHS/KAH/KAB/KHC boundaries, repo structure, operator report contract, self-improvement governance, and development acceptance gates; external feedback intake is summarized there and detailed by `sot/external-feedback-intake-policy.md`
- `sot/khs-pre-kah-readiness-audit-2026-05.md` — historical/superseded pre-KAH readiness audit for `t_2e9d918a`; preserved for lineage and overclaim prevention, not as a post-KAH blocker
- `sot/khs-delegation-packet-and-report-contract.md` — candidate docs-only Delegation Packet and report contract; defines packet/report fields, evidence paths, and status labels without upgrading them to KAH/KAB support
- `sot/external-feedback-intake-policy.md` — KAH-evidenced policy record for configurable `EXTERNAL_FEEDBACK_INTAKE` bounds; records min=1/max=5 semantics, user-supplied `feedback.md` MVP intake, KAS integration-pending surfaces, KAB-later automation, status labels, report fields, and a stale-surface manifest
- `sot/phase-orchestration-policy.md`
- `sot/interface-contract.md`
- `sot/workflow-graph-integration.md` — current SOT for `.kkachi-workflow.yaml` / capability-checked KAH 0.1.4 `kkachi-agent-helper graph` integration; `kas-default` template and KAS orchestration guidance exist, and `kah graph` alias remains candidate until separately evidenced
- `sot/graph-template-registry.md` — current GRAPHMVP-001/002 SOT for graph template registry metadata, schema expectations, and the `kas-default` template; active registry file is `registries/graph-template-registry.yaml`
- `sot/minimum-pilot-cli-lane.md` — Blue-confirmed SOT for the KHS+KAH minimum/pilot CLI lane; keeps profile-scoped install/list/doctor/sync/proposal support separate from full KHS+KAH+KAB execution-runtime authority
- `sot/kas-cli-contract.md` — CLIMVP-001 accepted SOT for KAS-owned `list`, `install --dry-run`, approved copy install, and `doctor`; records manifest/checksum, changed-path, approval, backup/recovery, status vocabulary, and fail-closed boundaries for the now-closed CLIMVP surface

## `roadmap.md`

KAS post-KAH enablement roadmap. It uses PR-candidate task sizing and is not implementation authorization by itself. KAH 0.1.4 graph and configurable-feedback support are now evidenced through capabilities/help; CLIMVP and GRAPHMVP-001..003 KAS surfaces are implemented, while `kah graph` alias behavior, GRAPHMVP artifact/report mapping, and KAB runtime alignment still proceed one task at a time with evidence and review gates.

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
| `kah-evidenced, kas-integration-pending` | KAH advertises or proves the required deterministic surface, but KAS docs/templates/registries/skills/CLI have not yet adopted it end-to-end |
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
- KAS owns graph templates, policy selection, phase applicability, proposal content, and skill/CLI adoption work.
- KAH owns deterministic validation/write/apply/diff/audit behavior for evidenced `kkachi-agent-helper graph` surfaces; KAH 0.1.4 also advertises configurable external-feedback intake capability, while KAS adoption remains integration-pending.
- The KHS+KAH minimum/pilot CLI lane is limited to profile-scoped install/list/doctor/sync/proposal support and does not claim KHC, Doksuri, KAB run/control, or bridge-runtime authority.
- `sot/kas-cli-contract.md` is the accepted CLIMVP-001 contract for KAS-owned profile-scoped `list`, `install --dry-run`, approved copy install, and `doctor`; implementation evidence now covers the full CLIMVP surface, while `sync` and broad `proposal` remain future gated work.
- KAB remains backend/session/plan evidence, not graph policy authority.
- Kkachi v2 `.kkachi/config/workflows/` is out of KAH/KHS graph scope and has no fallback/merge relationship with `.kkachi-workflow.yaml`.

## Stale/conflict markers

Older wording that calls `phase-plan.yaml` the whole workflow SOT must be read narrowly as run-local execution state/evidence. Any `kah graph` reference in these docs is planned/candidate unless alias capability/help evidence proves implementation. `kkachi-agent-helper graph` references may be treated as implemented after effective KAH 0.1.4 capability/help checks. Configurable external-feedback bounds are KAH-evidenced in 0.1.4 but remain KAS integration-pending until registries/templates/skills and reports are updated.

## Open questions

- KAS graph template registry format, default `kas-default` template content, and capability-checked orchestration adoption are defined by `sot/graph-template-registry.md`, `registries/graph-template-registry.yaml`, `templates/workflow-graphs/kas-default.yaml`, `sot/workflow-graph-integration.md`, and the KAS orchestration skills.
- Exact KAS artifact mapping for KAH graph proposal paths, semantic diff, validation/explain reports, approvals/audit evidence, checksums, graph versions, audit event ids, and capability-check evidence is defined by `templates/run-artifacts/graph-evidence.md.tmpl` and summarized in final reports under `kah_graph_evidence`.
- KAB alignment requires separate KAB docs work and is non-authoritative here.

## Next record action

INITDOC is closed and its temporary transition SOT has been deleted after absorption into active records. Future KAS implementation tasks should follow `docs/roadmap.md`, update only the docs/skills/templates/registries affected by the implemented surface, and preserve evidence paths.
