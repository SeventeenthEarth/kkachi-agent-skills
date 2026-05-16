# KHS Follow-up Discussion Items

Status: condensed open follow-up list after initial KHS phase orchestration review

This document keeps only the currently actionable follow-up decisions. Earlier broad discussion items were intentionally reduced so KHS, KAH, and KAB can converge on the next concrete interface behavior.

## 1. KAB planner output contract

Question: What should KHS do when KAB returns a plan but the planner output does not contain the requested `KHS Checklist Seed` section or the section is malformed?

Decision:

- Do **not** hard-fail only because the checklist seed is missing or loosely formatted.
- KHS owns the normalized `checklist.md` artifact.
- KAB owns plan lifecycle surfaces and exposes the plan text/evidence.
- KHS should derive `checklist.md` from:
  - KAB `plan.plan_text`
  - task contract
  - acceptance criteria
  - KHS phase contract
  - expected evidence and gates
- If the plan is usable but the checklist seed is missing or malformed, continue with a **recoverable warning** and record the normalization decision in run evidence.

Hard-fail conditions:

- `plan.plan_text` is missing or empty.
- The plan is too ambiguous to establish an implementation start point.
- KAB reports explicit ambiguity evidence such as OpenCode `PLAN_AMBIGUOUS`.
- Required plan evidence cannot be preserved before implementation starts.

Current policy:

- Prefer `KHS normalize + warning` over a strict planner-output parser.
- Preserve KAB plan text into KAH `plan.md` before approval/start.
- Generate KAH `checklist.md` as a KHS-owned normalized artifact.
- Revisit machine-readable checklist seed blocks after real dry-run evidence shows repeated parser friction.

## 2. Immediate validation path

Next work should validate the policy through real runs rather than more abstract discussion.

Representative runs:

- docs-only KHS run
- small code-change KHS run
- higher-risk code-change run requiring explicit master approval

Evidence to collect:

- `phase-plan.yaml`
- `plan.md`
- `checklist.md`
- KAB plan capture evidence
- KAH plan/final gate behavior
- any recoverable checklist-seed warning
- any friction where phase applicability, artifact naming, or gate requirements feel wrong

## 3. Deferred follow-ups

The following topics remain real, but should not block the next implementation/dry-run step:

- KAH first-class `phase-plan.yaml` command support
- stricter KAB machine-readable planner response contract
- default backend role mapping for planner/implementer/feedback
- docs-only KHS ergonomics
- feedback round escalation criteria
- optimize phase boundary examples
- release compatibility metadata format
- shared KHS promotion threshold

Rule for deferred topics:

- Promote only after run evidence shows repeated friction or compatibility risk.
- Shared KHS behavior changes require evidence, generalization, eval/update plan, and master approval.
