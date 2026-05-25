# External feedback intake policy

Date: 2026-05-23
Owner: KAS documentation archive
Status: KAH-evidenced policy SOT; KAS integration pending; not KAB automated-review runtime support
Authority level: current policy record for min=1/max=5 bounds; KAS registries/templates/skills/report adoption remains pending
Parent SOT: `docs/sot/khs-architecture-and-integration.md`
KAH evidence: KAH 0.1.4 `capabilities --json` advertises `workflow_graph_configurable_feedback_intake=true`; graph help/capabilities advertise init/validate/explain/diff/propose/apply/export
Scope: KAS preparation guidance for configurable `EXTERNAL_FEEDBACK_INTAKE` bounds

This record preserves the intended KAS policy now that KAH 0.1.4 advertises the
required configurable-feedback graph capability. It is still not a KAB automated
review support claim. KAS must keep generated reports, delegation instructions,
registries, templates, and skills honest: KAH support is evidenced, while KAS
adoption remains integration-pending until active KAS surfaces are updated and
verified.

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
triage it. KAS may prepare triage and implementation instructions from that
file, but KAS must keep automated review-by-different-tool outside the current
MVP and outside KAS-controlled support. Automated review-by-different-tool is
`kab_later`.

## Authority boundary

KAS owns the workflow policy language, phase guidance, templates, registries,
and operator report wording once those surfaces are explicitly updated. KAH must
own deterministic graph validation, schema support, proposal-first mutation,
audit events, run artifact checks, and capability/compatibility evidence. KAB
may later own runtime evidence for automated review transport, but KAB runtime
state is not project workflow policy authority.

KAS must not use direct YAML fallback as the normal path for
`.kkachi-workflow.yaml`. If a project requires graph-managed feedback bounds and
the effective KAH binary cannot validate the fields, the correct result is
`failed_closed` or `unsupported_effective_kah`, not a silent local default.

## KAH evidence and KAS adoption dependencies

KAH 0.1.4 now advertises the required graph/configurable-feedback substrate.
KAS adoption is still required before this policy can be reported as implemented
by active KAS runs:

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
- capability/compatibility evidence that KAS can read before enabling support.

KAS must fail closed when any of these negative cases is observed:

- bounds are absent where graph-managed external feedback intake is required;
- only stale `1..3`, `max_rounds: 3`, or `maximum_rounds: 3` semantics are
  available;
- bounds use unknown fields or unsupported schema versions;
- `min_rounds` is below 1, `max_rounds` is below `min_rounds`, or the current
  intended policy would allow round 6 or higher;
- round 1 is missing or optional;
- optional rounds are described as mandatory;
- request/handle feedback rows are unpaired for a realized round;
- `.kkachi/config.yaml`, generated diagrams, stale runtime state, KAS defaults,
  Kkachi v2 workflow config, or KAB runtime state is offered as fallback
  authority.

## Status label guidance

Use these words precisely:

| Label | Use |
|---|---|
| `candidate` | Proposed KAS/KAH/KAB behavior recorded for review; not current support. |
| `planned` | Expected future work with no current operational guarantee. |
| `implemented` | Proven by current repo command/code/docs/test or effective-binary capability evidence for the specific surface claimed. |
| `kah-evidenced, kas-integration-pending` | KAH advertises/proves the deterministic graph/policy surface, but KAS registries/templates/skills/reports are not yet updated end-to-end. |
| `blocked_by_kah` | KAS can describe the desired policy, but KAH lacks required schema, validation, proposal, audit, or compatibility support. |
| `kab_later` | Requires KAB backend/runtime evidence and is outside current MVP. |
| `historical` | Preserved context; not current authority. |
| `unsupported` | The current toolchain cannot safely execute or validate the requested support. |

For this record after KAH 0.1.4, the correct current label is
`kah-evidenced, kas-integration-pending`. Automated review-by-different-tool is
`kab_later`. Existing fixed-three-round surfaces are stale or historical until
separately remediated.

## Operator and delegation report contract

Reports prepared before KAS adoption is complete must be able to state partial
support truthfully. Candidate fields:

```yaml
external_feedback_intake:
  policy_key: EXTERNAL_FEEDBACK_INTAKE
  current_status: kah_evidenced_kas_integration_pending # or failed_closed
  support_label: kah-evidenced, kas-integration-pending
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
    status: supported_by_effective_kah_0_1_4
    command_evidence: kkachi-agent-helper graph --help
    capability_evidence: workflow_graph_configurable_feedback_intake=true
    proposal_audit_evidence: available_through_kah_graph_propose_apply
  kab_review_by_different_tool:
    status: kab_later
  stale_surface_status:
    status: kas_integration_pending
    manifest: docs/sot/external-feedback-intake-policy.md#stale-surface-manifest
  next_action: update_kas_registries_templates_skills_and_reports
```

Allowed `current_status` values before KAS adoption completes are
`kah_evidenced_kas_integration_pending` and `failed_closed`. A report may use
`failed_closed` when a run or proposal asked for graph-managed bounds but the
effective KAH binary or active KAS surfaces could not validate them.

## Stale-surface manifest

This manifest is the post-KAH KAS remediation plan for the known stale surfaces
listed in `docs/sot/khs-architecture-and-integration.md` under external
feedback intake. The plan records the target without editing registries,
templates, phase skills, or `.kkachi-workflow.yaml` in this preparation task.

| Surface | ASIS | TOBE target | Current post-KAH status | Remaining KAS work | Verification evidence |
|---|---|---|---|---|---|
| `README.md:108` | `request-feedback(1..3)` and `handle-feedback(1..3)` run shape | Document min=1/max=5 with required round 1 and optional rounds 2..5 | Updated by INITDOC-002 in repository README | Remaining registries/templates/skills/reports still require KAS adoption evidence | README read-back and stale-marker search |
| `registries/phase-contracts.yaml:72-73` | `minimum_rounds: 1`, `maximum_rounds: 3` | Schema/registry declares min 1 and max 5 after KAS adoption task | Do not edit registry in this task | Requires KAS registry change authorization using KAH-evidenced schema/compatibility substrate | `rg` found `maximum_rounds: 3`; registry edit forbidden |
| `registries/phase-contracts.yaml:140` | Conditional feedback must not exceed round 3 | Optional continuation rounds are 2..5 and not mandatory | Do not edit registry in this task | Requires KAH validation and phase-contract update | `rg` found `round 3`; registry edit forbidden |
| `registries/phase-contracts.yaml:202-203` | `conditional_feedback_rounds_2_to_3` | Conditional feedback range is 2..5 | Do not edit registry in this task | Requires KAS registry/schema alignment using KAH-evidenced substrate | Parent SOT inventory records the exact rows; registry edit forbidden |
| `registries/phase-contracts.yaml:222` | Final checklist rule says feedback runs at most three times | Final checks allow required round 1 and optional rounds 2..5 when realized | Do not edit registry in this task | Requires KAS run-artifact/final-check contract update using KAH-evidenced substrate | Parent SOT inventory records the row; registry edit forbidden |
| `registries/phase-contracts.yaml:286-293` | Request-feedback outputs stop at `feedback-3.md` and applicability says rounds 2..3 | Request-feedback supports realized rounds through `feedback-5.md` once implemented | Do not edit registry in this task | Requires KAH/KAS evidence and phase contract update | Parent SOT inventory records the rows; registry edit forbidden |
| `registries/phase-contracts.yaml:304-305` | Handle-feedback outputs stop at round 3 artifacts | Handle-feedback supports realized rounds through 5 once implemented | Do not edit registry in this task | Requires KAH/KAS evidence and phase contract update | Parent SOT inventory records the rows; registry edit forbidden |
| `templates/run-artifacts/task-contract.yaml.tmpl:59-61` | `feedback_rounds` max is 3 | Task contracts carry candidate min 1/max 5 semantics once support evidence exists | Do not edit run templates in this task | Requires KAH schema/template support and generated-artifact checks | Template edit forbidden; parent SOT records the row |
| `templates/run-artifacts/phase-plan.yaml.tmpl:64-65` | `max_rounds: 3` | Phase plans are generated from validated policy with max 5 | Do not edit run templates in this task | Requires KAH phase-plan generation/validation support | `rg` found `max_rounds: 3`; template edit forbidden |
| `templates/run-artifacts/phase-plan.yaml.tmpl:170-188` | Explicit round-3 maximum shape | Optional realized rows can include rounds 4 and 5 when useful feedback exists | Do not edit run templates in this task | Requires KAH/KAS template and artifact evidence | Parent SOT inventory records the rows; template edit forbidden |
| `templates/run-artifacts/checklist.md.tmpl:41-42` | Round 3 is the maximum feedback round | Checklist reflects configured bounds and realized optional rounds up to 5 | Do not edit run templates in this task | Requires KAH/KAS checklist/gate evidence | `rg` found round-3 checklist markers; template edit forbidden |
| `docs/sot/phase-orchestration-policy.md:98-99` | Up to two additional rounds, maximum three pairs | Supersede or amend to required round 1 and optional 2..5 | Leave unchanged; this child policy records candidate target | Requires KAH support or explicit docs remediation task | Parent SOT inventory records the rows |
| `docs/sot/phase-orchestration-policy.md:121-122` | Final verification expects one to three rounds | Final verification distinguishes candidate policy from implemented support | Leave unchanged; this child policy records candidate target | Requires KAH/run-artifact support and final-verifier update | Parent SOT inventory records the rows |
| `docs/sot/skill-template.md:356` and `581` | Rounds 2-3 only | Rounds 2..5 optional continuation when support exists | Leave unchanged; phase skill/template edits are out of scope | Requires KAS skill-template update using KAH-evidenced substrate | Parent SOT inventory records the rows |
| `docs/sot/concept.md:64` | Feedback may run up to three rounds | Mark historical or update after support evidence | Leave unchanged; this child policy provides candidate pointer | Requires KAH support or explicit concept-doc remediation | `rg` found stale concept marker |
| `docs/sot/concept.md:690` | Rounds 2-3 only | Rounds 2..5 optional continuation when support exists | Leave unchanged; this child policy provides candidate pointer | Requires KAH support or explicit concept-doc remediation | Parent SOT inventory records the row |
| `skills/kkachi-final-verify/SKILL.md:15` | Final verify checks one to three rounds | Final verify accepts required round 1 and optional realized rounds 2..5 | Do not edit phase skills in this task | Requires skill-change approval and KAH support evidence | Skill edit forbidden; parent SOT records the row |
| `skills/kkachi-orchestrate/SKILL.md:35` | Feedback runs at most three rounds | Orchestration can plan optional rounds 2..5 after support evidence | Do not edit phase skills in this task | Requires skill-change approval and KAH support evidence | `rg` found stale skill marker; skill edit forbidden |
| `skills/kkachi-plan/SKILL.md:80` | Rounds 2-3 conditional; do not exceed three pairs | Plan phase can express optional continuation 2..5 after support evidence | Do not edit phase skills in this task | Requires skill-change approval and KAH support evidence | Parent SOT inventory records the row |
| `skills/kkachi-request-feedback/SKILL.md:15` and `21` | Never exceed three pairs; optional files only through `feedback-3.md` | Request-feedback can request realized rounds through 5 after support evidence | Do not edit phase skills in this task | Requires skill-change approval, KAB-later automation distinction, and KAH support evidence | Parent SOT inventory records the rows |
| `skills/kkachi-handle-feedback/SKILL.md:21-22` | Optional handling artifacts only through round 3 | Handle-feedback can triage/handle realized rounds through 5 after support evidence | Do not edit phase skills in this task | Requires skill-change approval and KAH support evidence | `rg` found stale skill marker; skill edit forbidden |

## Verification requirements before support claim

Before KAS may label configurable external feedback intake as implemented, the
following evidence must exist:

- KAH docs/specs/compatibility state the accepted graph policy contract;
- KAH command/code/test evidence proves validation, proposal-first mutation,
  audit, diagnostics, and fail-closed cases for the policy;
- KAS registries/templates/phase skills are updated through approved KAS change
  tasks;
- stale `1..3` and `max3` surfaces are removed, updated, or explicitly marked
  historical with evidence;
- KAB evidence exists before any automated review-by-different-tool claim;
- operator/delegation reports can show source, bounds, current realized rounds,
  skipped optional rounds, missing evidence, and `kah_graph_evidence.status`.

Until then, this policy remains `kah-evidenced, kas-integration-pending` and
process-only for user-supplied `feedback.md` in the KAS MVP lane.
