# DESIGN Teal UI workflow policy for KAS

Status: planning SOT plus DESIGN-003 KAS selector/materializer source contract for the `DESIGN` shared KAS/KAH epic; KAH gates remain deferred
Owner: KAS workflow/policy layer
Source evidence: `/Users/draccoon/Workspace/Hermes/17thHermes/40_outputs/projects/kkachi/2026-06-21-kas-kah-teal-ui-workflow-sot.md`
Paired KAH SOT: `kkachi-agent-helper/docs/sot/teal-ui-evidence-gates.md`

## Purpose

This document registers the KAS-owned side of the `DESIGN` epic. The epic promotes the accepted candidate Teal UI workflow direction into KAS/KAH source planning and records the implemented DESIGN-003 KAS selector/materializer contract without claiming KAH DESIGN-004/005 behavior.

KAS owns workflow policy, trigger semantics, role contracts, node contracts, selector/materializer behavior, skill guidance, and agent-facing expectations. KAS does not own KAH schema/gate validation, and KAS must not substitute Blue, Red, Orange, Gray, MAR, backend agents, or temporary helpers for official Teal verdicts when Teal is required.

DESIGN-003 implements the KAS selector/materializer portion. `workflow-route` derives and records `teal_required` from explicit `project_has_teal_lane` and `ui_ux_change` facts, and `workflow-trigger` route-backed materialization inserts design gates only for `teal_required=true`. KAH schema and gate enforcement remain DESIGN-004 and DESIGN-005.

DESIGN-006 records cross-repo compatibility examples in `docs/examples/design006-teal-compatibility-scenarios.json`. The golden cases are `kkachi_non_ui_skip`, `kkachi_teal_lane_non_ui_skip`, `sudal_ui_required`, and `doksuri_ui_required`; KAS owns the declarations and expected materialized Teal nodes, while KAH readback proves the same declarations satisfy or fail the deterministic `design-evidence` gate.

## Applicability rule

KAS derives `teal_required` from declared project/task facts:

```yaml
project_has_teal_lane: true|false
ui_ux_change: true|false
teal_required: project_has_teal_lane && ui_ux_change
```

When either input is false, KAS must preserve a concrete `teal_skip_reason` instead of injecting design workflow steps.

Teal waiver evidence is not a skip reason. A skip records why Teal did not
apply from project/task facts. A waiver records that Teal did apply but a
responsible approval intentionally waived one or more required Teal evidence
items for a bounded scope. KAS records waiver metadata only as policy/evidence
expectations; KAH schema and gate validation for those fields belongs to
`DESIGN-004` and `DESIGN-005`.

For Kkachi KAS/KAH/KAB source work, the default is:

```yaml
project_has_teal_lane: false
ui_ux_change: false
teal_required: false
teal_skip_reason: "No UI/UX surface in this project/task."
teal_waiver_approved: false
teal_waiver_approval_ref: ""
teal_waiver_scope: ""
teal_waiver_expires_at: ""
```

For UI-bearing Sudal or Doksuri work, the project Teal designer or Goong as Teal Team Lead is routed according to project registration and explicit task scope.

The DESIGN-006 golden examples use one mixed Kkachi Teal-lane/non-UI skip to prove AND-not-OR derivation, and use Sudal and Doksuri as UI-required fixture contexts only. They do not implement downstream Sudal/Doksuri UI changes, assign Teal owners, or let ordinary color review, MAR, backend evidence, helper notes, or temporary subagents substitute for required Teal verdicts.

When `teal_required=true`, KAS must require the design gate records before the
ordinary implementation and final-acceptance claims they protect:

```yaml
design_plan_gate: DESIGN_PLAN_GATE
design_fidelity_review: DESIGN_FIDELITY_REVIEW
missing_required_status: required_teal_verdict_missing
```

`DESIGN_PLAN_GATE` is the Teal planning verdict before implementation
authorization. `DESIGN_FIDELITY_REVIEW` is the Teal fidelity verdict before
final acceptance. Ordinary Red/Orange/Gray/Blue color review, MAR role review,
backend implementation evidence, temporary subagents, and helper notes remain
separate evidence lanes; they must not substitute Blue, Red, Orange, Gray, MAR,
backend agents, or temporary helpers for official Teal verdicts when Teal is
required.

If `teal_required=false`, KAS must not inject Teal workflow nodes into Kkachi
source-work runs. It records the false inputs and concrete `teal_skip_reason`
instead.

## KAS policy surfaces

KAS implementation tasks must add or update support for:

- project workflow metadata for Teal availability;
- task-level `ui_ux_change` classification;
- derived `teal_required`, explicit skip reasons, and waiver policy;
- `DESIGN_PLAN_GATE` before implementation authorization;
- `DESIGN_FIDELITY_REVIEW` before final acceptance;
- optional Teal participation in general color review when UI changed;
- node contracts and skill guidance that keep design planning, design fidelity review, and ordinary color review separate;
- fail-closed no-substitution language when required Teal verdict/evidence is missing.

## Sequential task order

주군 selected a sequential seven-task roadmap. Do not parallelize these tasks. Execute by task id order:

1. `DESIGN-001` — KAS docs/SOT adoption and shared roadmap registration.
2. `DESIGN-002` — KAS Teal applicability and node contract semantics.
3. `DESIGN-003` — KAS workflow selector/materializer and skill guidance.
4. `DESIGN-004` — KAH design evidence schema and artifact bootstrap.
5. `DESIGN-005` — KAH fail-closed gate and diagnostics support.
6. `DESIGN-006` — Cross-repo compatibility examples and proof fixtures.
7. `DESIGN-007` — Full verification, docs map, Red/Orange/Gray/Teal review, and Blue closeout.

KAS/KAH remain separated by ownership. A KAS task may make a small KAH compatibility note or fixture-name alignment only when needed, and a KAH task may make a small KAS reference update only when needed. Any new command, schema, gate, selector, materializer, or independent acceptance criteria in the other repository must become its own planned work rather than a hidden cross-edit.

## Boundaries

This planning SOT does not authorize code behavior, installed profile skill updates, runtime/provider/auth/token/gateway/model mutation, KAB activation, release, push, or applying Teal to non-UI Kkachi repository work. Implementation status requires task-specific source changes, tests, KAH evidence where applicable, and official review gates.
