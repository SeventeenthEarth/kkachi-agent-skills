# KHS Delegation Packet and report contract

Date: 2026-05-23
Owner: KHS documentation archive
Confirming role: Kanban `t_2e9d918a` Codex app-server soldier draft; Blue-inspected candidate with same-card Red/Gray/Orange ACCEPT gates before final synthesis
Status: candidate contract; not_operational_support
Authority level: candidate SOT for packet/report shape; does not implement KAH artifacts or KAB execution
Scope: KHS docs-only contract for delegation packets, evidence expectations, and operator reports
Runtime note: this draft was requested with `openai_runtime=codex_app_server`; it is not KAB runtime evidence

Related docs:
- `docs/sot/khs-architecture-and-integration.md`
- `docs/sot/khs-pre-kah-readiness-audit-2026-05.md`
- `docs/sot/external-feedback-intake-policy.md`
- `docs/sot/workflow-graph-integration.md`
- `docs/sot/minimum-pilot-cli-lane.md`
- `docs/README.md`

Evidence/source paths:
- Current corrective task: Kanban `t_2e9d918a`
- External-feedback child policy lineage: Kanban `t_fa8f17a3`
- Prior referenced card: Kanban `t_2a151490`
- Parent architecture target: `docs/sot/khs-architecture-and-integration.md`
- Readiness audit: `docs/sot/khs-pre-kah-readiness-audit-2026-05.md`
- External-feedback child policy: `docs/sot/external-feedback-intake-policy.md`
- Graph boundary: `docs/sot/workflow-graph-integration.md`
- Pilot CLI boundary: `docs/sot/minimum-pilot-cli-lane.md`

## 1. Decision summary

KHS should describe a Delegation Packet and report contract before KAH and KAB
runtime support exists, but this document is not itself runtime support.

The contract is safe pre-KAH because it records the desired packet/report shape,
status labels, evidence fields, and fail-closed rules without changing
registries, templates, phase skills, KAH state, KAB runtime behavior, or
`.kkachi-workflow.yaml`.

Current labels:

| Area | Label |
|---|---|
| This document | `candidate`, `not_operational_support` |
| KAH artifact persistence/validation | `blocked_by_kah` |
| KAB backend/session evidence | `kab_later` |
| Future packet/report generation | `planned` |

## 2. Layer boundaries

| Layer | Packet/report role | Must not be inferred |
|---|---|---|
| KHS | Defines AI-neutral task contract, packet fields, prompt guidance, evidence expectations, report fields, and support labels. | KHS does not own deterministic state, graph apply/audit, backend runtime, or proof of completion. |
| KAH | Future owner for deterministic run artifacts, schema validation, events, locks, gates, graph validation/proposal/apply/audit, diagnostics, and persisted report evidence. | KAH does not decide workflow policy or backend prompt strategy. |
| KAB | Future owner for backend dispatch, wait/read/status, plan/question/approval lifecycle, retained stream/event evidence, and backend completion evidence. | KAB does not own KHS policy or KAH project graph authority. |
| Hermes/KHC | Selects lane, approves risk, routes reviews, synthesizes accepted/rejected/deferred items, and reports to the user. | Human-facing synthesis does not substitute for missing KAH/KAB evidence. |

## 3. Delegation Packet fields

A KHS Delegation Packet should be compact enough to execute and complete enough
to reconstruct decisions without hidden conversation state.

```yaml
delegation_packet:
  packet_id: "<stable id or candidate id>"
  status: "candidate|planned|blocked_by_kah|kab_later|not_operational_support"
  source_refs:
    kanban_cards: []
    docs:
      - "README.md"
      - "docs/README.md"
      - "docs/sot/khs-architecture-and-integration.md"
  parsed_intent: "<what the user asked for>"
  selected_lane: "direct_hermes|khs_kah_minimum_pilot|khs_kah_kab_runtime|docs_only_sot|proposal_only"
  execution_mode: "read_only|docs_only|code_change|review_only|proposal_only|runtime_execution"
  constraints:
    must_do: []
    must_not_do: []
    destructive_or_external_actions: []
  non_goals: []
  acceptance_criteria: []
  evidence_plan:
    kah_artifacts: []
    kab_evidence: []
    docs_evidence: []
    tests_or_checks: []
    review_cards: []
  capability_needs:
    kah: []
    kab: []
    local_repo: []
  boundary_labels:
    - "KHS policy/template"
    - "KAH graph/run evidence"
    - "KAB backend evidence"
    - "user-supplied feedback"
    - "KHC review"
    - "project overlay"
    - "shared KHS proposal"
  fail_closed_triggers: []
  prompt_profile:
    backend: "<candidate backend or not_applicable>"
    caveats: []
  report_contract_ref: "docs/sot/khs-delegation-packet-and-report-contract.md#4-report-fields"
```

Before KAH support exists, any packet field that implies persisted artifact
validation must be labeled `planned` or `blocked_by_kah`. Before KAB support
exists, backend execution evidence fields must be labeled `kab_later` or
`not_applicable`.

## 4. Report fields

A KHS report should let an operator distinguish child-policy completion, whole
KHS readiness, and operational/runtime support readiness.

| Field | Required meaning |
|---|---|
| `parsed_intent` | Concise interpretation and source refs. |
| `current_status` | One of `candidate`, `planned`, `blocked`, `failed_closed`, `done`, `candidate_only`, or `not_operational_support` as applicable. |
| `selected_lane` | Direct Hermes, KHS+KAH minimum/pilot, KHS+KAH+KAB runtime, docs-only SOT, or proposal-only. |
| `execution_mode` | Read-only, docs-only, code-change, review-only, proposal-only, or runtime execution. |
| `changed_files` | Exact repo-relative paths changed by the run. |
| `forbidden_files_touched` | `none` or exact paths; must be `none` for KAH/KAB repos and behavior files in this task. |
| `external_feedback_child_completion` | Status and evidence path for `docs/sot/external-feedback-intake-policy.md` when relevant. |
| `whole_khs_pre_kah_readiness` | Status, gaps, and evidence paths for broad readiness. |
| `operational_runtime_support_readiness` | Explicit KAH/KAB status labels; docs-only records should say `not_operational_support`. |
| `safe_pre_kah_actions` | Actions that can proceed without KAH/KAB implementation. |
| `blocked_by_kah` | Exact missing KAH support or evidence. |
| `kab_later` | Exact KAB/runtime evidence deferred. |
| `accepted_items` | Accepted feedback/review/task items with reasons. |
| `rejected_items` | Rejected items with reasons. |
| `deferred_items` | Deferred items with owner or blocking evidence. |
| `evidence_paths` | Exact file paths, command outputs, Kanban ids, review cards, or artifact ids. |
| `tests_or_verification` | Commands/checks run, or `not_run` with reason. |
| `remaining_risk` | Concrete residual risk; use `none_known` only with evidence. |
| `next_valid_action` | Exact next safe command, card, review, or `none`. |

## 5. Status matrix

| Claim type | Allowed current label | Required evidence before stronger claim |
|---|---|---|
| Candidate packet/report shape exists | `candidate`, `not_operational_support` | This document and docs index entry. |
| KAH can persist and validate packet/report artifacts | `blocked_by_kah` | KAH schema, command, code, docs, tests, compatibility/capability output, and artifact examples. |
| KAB can produce backend execution evidence for packet phases | `kab_later` | KAB send/wait/read/status/event evidence and backend-specific caveat handling. |
| External-feedback child policy is documented | `candidate`, scoped complete | `docs/sot/external-feedback-intake-policy.md` and docs index. |
| Whole KHS pre-KAH readiness is complete | `planned` until all broad readiness areas close | Readiness audit, packet/report contract, stale/overclaim coverage, graph/CLI/self-improvement status, and review closure. |
| Operational runtime support is ready | `not_operational_support` from docs alone | KAH/KAB implementation evidence plus KHS registry/template/skill updates and verification. |

## 6. Fail-closed rules

Fail closed rather than reporting success when:

- a packet references KAH artifact validation but no KAH evidence exists;
- a report claims backend completion from KAB dispatch/send success alone;
- status labels are missing or collapse `candidate`, `planned`, `blocked_by_kah`, and `kab_later`;
- child external-feedback completion is reported as whole KHS pre-KAH readiness;
- stale `1..3` or max-3 feedback surfaces are treated as support for min=1/max=5;
- `.kkachi-workflow.yaml` is manually edited or created without KAH validation/proposal/apply/audit evidence;
- `.kkachi/config.yaml`, generated diagrams, stale runtime state, or KAB state is offered as workflow policy authority;
- accepted/rejected/deferred feedback items are not separated;
- evidence paths are vague, missing, or conversation-only.

## 7. Safe pre-KAH usage

This contract may be used before KAH/KAB implementation for:

- docs-only SOT drafts;
- readiness audits;
- review triage reports;
- planned packet/report schema discussion;
- project-overlay or shared-KHS proposal wording;
- stale/overclaim prevention.

It must not be used before implementation evidence to claim:

- validated KAH artifact creation;
- KAH graph mutation or audit;
- KAB backend execution completion;
- automated different-tool review;
- operational support for `EXTERNAL_FEEDBACK_INTAKE` min=1/max=5;
- production-ready KHS runtime behavior.

## 8. Minimal example

```yaml
parsed_intent: "create KHS pre-KAH readiness docs for t_2e9d918a"
current_status: candidate
selected_lane: docs_only_sot
execution_mode: docs_only
external_feedback_child_completion:
  status: candidate
  evidence: docs/sot/external-feedback-intake-policy.md
whole_khs_pre_kah_readiness:
  status: planned
  evidence:
    - docs/sot/khs-pre-kah-readiness-audit-2026-05.md
    - docs/sot/khs-delegation-packet-and-report-contract.md
operational_runtime_support_readiness:
  status: not_operational_support
  kah_status: blocked_by_kah
  kab_status: kab_later
blocked_by_kah:
  - "artifact schema/validation/proposal/apply/audit evidence"
kab_later:
  - "backend review transport and execution evidence"
```

## 9. Verification for this contract

This contract is complete only when:

- it is discoverable from `docs/README.md`;
- it names its candidate/non-operational status;
- it separates KHS, KAH, KAB, and Hermes/KHC responsibilities;
- it preserves `t_fa8f17a3`, `t_2a151490`, and `t_2e9d918a` as bounded source ids;
- it does not edit registries, templates, phase skills, KAH/KAB repos, or `.kkachi-workflow.yaml`;
- final reporting includes changed files, diff summary, verification, and remaining risks.

