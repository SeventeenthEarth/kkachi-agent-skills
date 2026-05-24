# External feedback intake policy

Date: 2026-05-23
Owner: KHS documentation archive
Status: candidate/planned child SOT; not operational runtime support
Authority level: candidate policy record pending KAH implementation evidence
Parent SOT: `docs/sot/khs-architecture-and-integration.md`
KAH dependency record: `/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-helper/.docs/new-support.md`
Scope: KHS preparation guidance for configurable `EXTERNAL_FEEDBACK_INTAKE` bounds

This record preserves the intended KHS policy before KAH support exists. It is
not a support claim. KHS must keep generated reports, delegation instructions,
and docs honest: configurable external feedback intake is candidate/planned
until KAH can validate, explain, propose, apply, audit, and advertise the
needed graph/policy support with command, code, docs, and test evidence.

## Policy target

`EXTERNAL_FEEDBACK_INTAKE` has the following target bounds:

- minimum rounds: `1`;
- maximum rounds: `5`;
- round 1 is required;
- rounds 2 through 5 are optional continuation rounds only when useful external
  feedback exists;
- five rounds are never mandatory;
- if no actionable feedback exists after the required intake, optional rounds
  must be skipped with reasons rather than forced.

The current MVP intake source is user-supplied `feedback.md`. The user may run
an outside review, save the result as `feedback.md`, and ask Blue/Hermes to
triage it. KHS may prepare triage and implementation instructions from that
file, but KHS must keep automated review-by-different-tool outside the current
MVP and outside KHS-controlled support. Automated review-by-different-tool is
`kab_later`.

## Authority boundary

KHS owns the workflow policy language, phase guidance, templates, registries,
and operator report wording once those surfaces are explicitly updated. KAH must
own deterministic graph validation, schema support, proposal-first mutation,
audit events, run artifact checks, and capability/compatibility evidence. KAB
may later own runtime evidence for automated review transport, but KAB runtime
state is not project workflow policy authority.

Until KAH support lands, KHS must not use direct YAML fallback as the normal
path for `.kkachi-workflow.yaml`. If the project requires graph-managed
feedback bounds and KAH cannot validate the fields, the correct result is
`blocked_by_kah` or `failed_closed`, not a silent local default.

## KAH dependencies

KAH support is required before this policy can be reported as implemented:

- schema-owned policy identity for `EXTERNAL_FEEDBACK_INTAKE` or an accepted
  equivalent;
- required `min_rounds=1`, `max_rounds=5`, round-1-required semantics, and
  optional continuation range 2..5;
- deterministic validation for missing, stale, duplicate, unknown,
  unsupported, conflicting, or out-of-range bounds;
- proposal-first graph mutation and audit evidence;
- graph-vs-run-artifact conflict detection across `.kkachi-workflow.yaml`,
  run-local `phase-plan.yaml`, `checklist.md`, KAH events, reports, and feedback
  artifacts;
- capability/compatibility evidence that KHS can read before enabling support.

KHS must fail closed when any of these negative cases is observed:

- bounds are absent where graph-managed external feedback intake is required;
- only stale `1..3`, `max_rounds: 3`, or `maximum_rounds: 3` semantics are
  available;
- bounds use unknown fields or unsupported schema versions;
- `min_rounds` is below 1, `max_rounds` is below `min_rounds`, or the current
  intended policy would allow round 6 or higher;
- round 1 is missing or optional;
- optional rounds are described as mandatory;
- request/handle feedback rows are unpaired for a realized round;
- `.kkachi/config.yaml`, generated diagrams, stale runtime state, KHS defaults,
  Kkachi v2 workflow config, or KAB runtime state is offered as fallback
  authority.

## Status label guidance

Use these words precisely:

| Label | Use |
|---|---|
| `candidate` | Proposed KHS/KAH/KAB behavior recorded for review; not current support. |
| `planned` | Expected future work with no current operational guarantee. |
| `implemented` | Proven by current repo command/code/docs/test or effective-binary capability evidence for the specific surface claimed. |
| `blocked_by_kah` | KHS can describe the desired policy, but KAH lacks required schema, validation, proposal, audit, or compatibility support. |
| `kab_later` | Requires KAB backend/runtime evidence and is outside current MVP. |
| `historical` | Preserved context; not current authority. |
| `unsupported` | The current toolchain cannot safely execute or validate the requested support. |

For this record, the correct current labels are `candidate`, `planned`, and
`blocked_by_kah`. Automated review-by-different-tool is `kab_later`. Existing
fixed-three-round surfaces are stale or historical until separately remediated.

## Operator and delegation report contract

Reports prepared before KAH support exists must be able to state unsupported
truthfully. Candidate fields:

```yaml
external_feedback_intake:
  policy_key: EXTERNAL_FEEDBACK_INTAKE
  current_status: candidate_only # or failed_closed
  support_label: blocked_by_kah
  configured_bounds:
    min_rounds: 1
    max_rounds: 5
    round_1_required: true
    optional_continuation_rounds: [2, 3, 4, 5]
    five_rounds_mandatory: false
  mvp_intake_source:
    status: supported_by_process
    path: feedback.md
    description: user-supplied external feedback file
  kah_graph_evidence:
    status: unsupported
    command_evidence: missing
    schema_evidence: missing
    proposal_audit_evidence: missing
    compatibility_evidence: missing
  kab_review_by_different_tool:
    status: kab_later
  stale_surface_status:
    status: blocked_by_kah
    manifest: docs/sot/external-feedback-intake-policy.md#stale-surface-manifest
  next_action: wait_for_kah_support_or_keep_process_candidate_only
```

Allowed `current_status` values before KAH support are `candidate_only` and
`failed_closed`. A report may use `failed_closed` when a run or proposal asked
for graph-managed bounds but KAH could not validate them.

## Stale-surface manifest

This manifest is the safe pre-KAH remediation plan for the known stale surfaces
listed in `docs/sot/khs-architecture-and-integration.md` under external
feedback intake. The plan records the target without editing registries,
templates, phase skills, or `.kkachi-workflow.yaml` in this preparation task.

| Surface | ASIS | TOBE target | Safe pre-KAH status | KAH-blocked status | Verification evidence |
|---|---|---|---|---|---|
| `README.md:108` | `request-feedback(1..3)` and `handle-feedback(1..3)` run shape | Document min=1/max=5 as candidate or mark the example historical until runtime support exists | Leave unchanged in this task because `README.md` has unrelated dirty state and the docs index is the requested pointer surface | Update only after KAH/KHS support evidence or with an explicit historical marker | `rg` found the stale marker; no README edit in this task |
| `registries/phase-contracts.yaml:72-73` | `minimum_rounds: 1`, `maximum_rounds: 3` | Schema/registry declares min 1 and max 5 after KAH support exists | Do not edit registry in this task | Requires KAH schema/compatibility evidence and KHS registry change authorization | `rg` found `maximum_rounds: 3`; registry edit forbidden |
| `registries/phase-contracts.yaml:140` | Conditional feedback must not exceed round 3 | Optional continuation rounds are 2..5 and not mandatory | Do not edit registry in this task | Requires KAH validation and phase-contract update | `rg` found `round 3`; registry edit forbidden |
| `registries/phase-contracts.yaml:202-203` | `conditional_feedback_rounds_2_to_3` | Conditional feedback range is 2..5 | Do not edit registry in this task | Requires KAH support and registry/schema alignment | Parent SOT inventory records the exact rows; registry edit forbidden |
| `registries/phase-contracts.yaml:222` | Final checklist rule says feedback runs at most three times | Final checks allow required round 1 and optional rounds 2..5 when realized | Do not edit registry in this task | Requires KAH/run-artifact support and final-check contract update | Parent SOT inventory records the row; registry edit forbidden |
| `registries/phase-contracts.yaml:286-293` | Request-feedback outputs stop at `feedback-3.md` and applicability says rounds 2..3 | Request-feedback supports realized rounds through `feedback-5.md` once implemented | Do not edit registry in this task | Requires KAH/KHS evidence and phase contract update | Parent SOT inventory records the rows; registry edit forbidden |
| `registries/phase-contracts.yaml:304-305` | Handle-feedback outputs stop at round 3 artifacts | Handle-feedback supports realized rounds through 5 once implemented | Do not edit registry in this task | Requires KAH/KHS evidence and phase contract update | Parent SOT inventory records the rows; registry edit forbidden |
| `templates/run-artifacts/task-contract.yaml.tmpl:59-61` | `feedback_rounds` max is 3 | Task contracts carry candidate min 1/max 5 semantics once support evidence exists | Do not edit run templates in this task | Requires KAH schema/template support and generated-artifact checks | Template edit forbidden; parent SOT records the row |
| `templates/run-artifacts/phase-plan.yaml.tmpl:64-65` | `max_rounds: 3` | Phase plans are generated from validated policy with max 5 | Do not edit run templates in this task | Requires KAH phase-plan generation/validation support | `rg` found `max_rounds: 3`; template edit forbidden |
| `templates/run-artifacts/phase-plan.yaml.tmpl:170-188` | Explicit round-3 maximum shape | Optional realized rows can include rounds 4 and 5 when useful feedback exists | Do not edit run templates in this task | Requires KAH/KHS template and artifact evidence | Parent SOT inventory records the rows; template edit forbidden |
| `templates/run-artifacts/checklist.md.tmpl:41-42` | Round 3 is the maximum feedback round | Checklist reflects configured bounds and realized optional rounds up to 5 | Do not edit run templates in this task | Requires KAH/KHS checklist/gate evidence | `rg` found round-3 checklist markers; template edit forbidden |
| `docs/sot/phase-orchestration-policy.md:98-99` | Up to two additional rounds, maximum three pairs | Supersede or amend to required round 1 and optional 2..5 | Leave unchanged; this child policy records candidate target | Requires KAH support or explicit docs remediation task | Parent SOT inventory records the rows |
| `docs/sot/phase-orchestration-policy.md:121-122` | Final verification expects one to three rounds | Final verification distinguishes candidate policy from implemented support | Leave unchanged; this child policy records candidate target | Requires KAH/run-artifact support and final-verifier update | Parent SOT inventory records the rows |
| `docs/sot/skill-template.md:356` and `581` | Rounds 2-3 only | Rounds 2..5 optional continuation when support exists | Leave unchanged; phase skill/template edits are out of scope | Requires KAH support and skill-template update | Parent SOT inventory records the rows |
| `docs/sot/concept.md:64` | Feedback may run up to three rounds | Mark historical or update after support evidence | Leave unchanged; this child policy provides candidate pointer | Requires KAH support or explicit concept-doc remediation | `rg` found stale concept marker |
| `docs/sot/concept.md:690` | Rounds 2-3 only | Rounds 2..5 optional continuation when support exists | Leave unchanged; this child policy provides candidate pointer | Requires KAH support or explicit concept-doc remediation | Parent SOT inventory records the row |
| `skills/kkachi-final-verify/SKILL.md:15` | Final verify checks one to three rounds | Final verify accepts required round 1 and optional realized rounds 2..5 | Do not edit phase skills in this task | Requires skill-change approval and KAH support evidence | Skill edit forbidden; parent SOT records the row |
| `skills/kkachi-orchestrate/SKILL.md:35` | Feedback runs at most three rounds | Orchestration can plan optional rounds 2..5 after support evidence | Do not edit phase skills in this task | Requires skill-change approval and KAH support evidence | `rg` found stale skill marker; skill edit forbidden |
| `skills/kkachi-plan/SKILL.md:80` | Rounds 2-3 conditional; do not exceed three pairs | Plan phase can express optional continuation 2..5 after support evidence | Do not edit phase skills in this task | Requires skill-change approval and KAH support evidence | Parent SOT inventory records the row |
| `skills/kkachi-request-feedback/SKILL.md:15` and `21` | Never exceed three pairs; optional files only through `feedback-3.md` | Request-feedback can request realized rounds through 5 after support evidence | Do not edit phase skills in this task | Requires skill-change approval, KAB-later automation distinction, and KAH support evidence | Parent SOT inventory records the rows |
| `skills/kkachi-handle-feedback/SKILL.md:21-22` | Optional handling artifacts only through round 3 | Handle-feedback can triage/handle realized rounds through 5 after support evidence | Do not edit phase skills in this task | Requires skill-change approval and KAH support evidence | `rg` found stale skill marker; skill edit forbidden |

## Verification requirements before support claim

Before KHS may label configurable external feedback intake as implemented, the
following evidence must exist:

- KAH docs/specs/compatibility state the accepted graph policy contract;
- KAH command/code/test evidence proves validation, proposal-first mutation,
  audit, diagnostics, and fail-closed cases for the policy;
- KHS registries/templates/phase skills are updated through approved KHS change
  tasks;
- stale `1..3` and `max3` surfaces are removed, updated, or explicitly marked
  historical with evidence;
- KAB evidence exists before any automated review-by-different-tool claim;
- operator/delegation reports can show source, bounds, current realized rounds,
  skipped optional rounds, missing evidence, and `kah_graph_evidence.status`.

Until then, this policy remains `candidate/planned`, `blocked_by_kah`, and
process-only for user-supplied `feedback.md`.
