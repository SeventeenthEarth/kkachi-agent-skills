# KHS Follow-up Discussion Items

Status: open follow-up list for post-initial PR discussion

This document records topics intentionally left open after the initial KHS seed implementation. The initial PR should establish the phase orchestration baseline; these items should be discussed and refined through later PRs or real KHS dry-runs.

## 1. Real project dry-run validation

Question: Does the current phase spine work cleanly in a real project run?

Validate with at least one representative project task:

- docs-only KHS run
- small code-change KHS run
- higher-risk code-change run that requires explicit master approval

Evidence to collect:

- generated `phase-plan.yaml`
- generated `plan.md`
- generated `checklist.md`
- KAB plan capture evidence
- KAH gate/final gate behavior
- friction points where phase applicability or artifact naming feels wrong

## 2. KAH first-class phase-plan support

Question: Should KAH eventually manage `phase-plan.yaml` directly?

Current decision:

- KHS owns `phase-plan.yaml` as workflow SOT.
- KAH stores deterministic run artifacts and gates, but currently has no first-class phase-plan command surface.

Possible future KAH support:

- `phase-plan init <run_id>`
- `phase-plan set <run_id> <phase> --state ... --reason ...`
- validation that skipped/not-applicable phases include reasons
- validation that feedback rounds are between 1 and 3
- validation that code-change runs include optimize evidence or skip reason

Open concern:

- KAH must remain deterministic and must not decide phase applicability intelligently.

## 3. KAB planner output contract hardening

Question: How strict should KAB be about planner output shape?

Current decision:

- KHS asks KAB planner backends to include a `KHS Checklist Seed` section inside `plan.plan_text`.
- KHS parses and normalizes that seed into `checklist.md`.

Possible follow-up:

- define a stricter planner response convention
- add a machine-readable checklist seed block
- add KAB adapter warnings when plan text lacks the requested section
- decide whether missing checklist seed is a hard failure or a KHS-recoverable warning

Open concern:

- Too much strictness may reduce backend compatibility; too little strictness increases KHS parsing burden.

## 4. Logical backend role mapping policy

Question: When should planner, implementer, and feedback roles use the same backend/session versus separate backends?

Current decision:

- KHS defines logical roles:
  - planner
  - implementer
  - feedback
- Physical backend mapping is decided by project policy, user preference, Hermes judgment, and capability evidence.

Possible follow-up:

- default to one backend for low-risk runs
- use separate read-only feedback backend for medium/high-risk runs
- use red-team/review backend when API, DB, security, dependency, architecture, SOT, large diff, low confidence, or unresolved ambiguity is present

## 5. Docs-only KHS ergonomics

Question: Is KAB-by-default too heavy for docs-only runs?

Current decision:

- Docs-only KHS runs use KAB by default.
- Direct docs editing is allowed only when the master explicitly forbids KAB and the rationale is recorded.

Possible follow-up:

- test a docs-only run in practice
- decide whether there should be a lightweight docs-only profile that still uses KAB but reduces prompts/artifacts
- ensure direct-Hermes exception remains outside KHS rather than becoming a silent bypass

## 6. Feedback loop practical limits

Question: When should Hermes request rounds 2 and 3?

Current decision:

- feedback round 1 is mandatory
- rounds 2 and 3 are conditional
- maximum is three request-feedback/handle-feedback pairs

Possible follow-up criteria for extra rounds:

- feedback changed code or docs materially
- feedback exposed architecture or SOT ambiguity
- backend output quality is uncertain
- verification passed but risk remains high
- the first feedback was shallow or missed known risk surfaces

## 7. Optimize phase boundaries

Question: How much cleanup is appropriate in `optimize`?

Current decision:

- code-change runs should run optimize when practical
- optimize removes AI slop, duplication, dead/verbose code, and small structural waste
- broad rewrites are out of scope unless separately approved

Possible follow-up:

- define examples of acceptable vs unacceptable optimize changes
- decide when optimize should be performed by the same implementer backend or a separate review/red-team backend
- define when optimize requires another verification pass

## 8. Release compatibility metadata

Question: What exact dependency metadata should KHS release tags publish?

Current decision:

- KHS main tracks latest-compatible KAH/KAB command surfaces.
- KHS release tags should publish tested/recommended KAH and KAB versions.

Possible follow-up:

- define release metadata format in `skill-pack.yaml` or a dedicated compatibility file
- record tested KAH version, KAB commit/tag, backend CLI versions, and required capabilities
- distinguish `minimum`, `tested`, and `recommended` versions

## 9. Artifact naming and KAH gate alignment

Question: Are all KHS supplemental artifacts aligned with KAH canonical artifacts and gate expectations?

Current known alignment:

- KAH plan gate requires `acceptance-criteria.md`, `plan.md`, and `checklist.md`.
- KHS now creates plan/checklist templates and phase-plan supplemental artifact.

Possible follow-up:

- run KAH artifact/gate checks against a sample KHS run directory
- identify artifacts that should become canonical KAH artifacts
- identify KHS-only artifacts that should remain supplemental

## 10. Shared KHS promotion policy

Question: When should project-local lessons be promoted back into shared KHS?

Current decision:

- project-local overlays and run evidence should absorb customization first
- shared KHS changes require repeated evidence and review

Possible follow-up:

- define a threshold for promotion, such as two or more independent runs showing the same failure mode
- require improvement notes to include evidence, generalized lesson, affected files, proposed patch, and regression/eval plan
- decide whether a user approval checkpoint is required for all shared KHS behavior changes
