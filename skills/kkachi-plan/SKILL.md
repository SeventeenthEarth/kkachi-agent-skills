---
name: kkachi-plan
description: Produce or validate a Kkachi plan from SOT basis, roadmap trace, task contract, acceptance criteria, constraints, non-goals, backend lane requirements, and verification strategy.
version: 0.1.0
---

# Kkachi Plan

Use this skill during the `plan` phase of Path A or Path B.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Plan from durable authority, not chat-only instruction. Path A plans prepare implementation; Path B plans prepare shaping and handoff.

## Inputs

- `task-contract.yaml`
- SOT basis and roadmap trace
- project overlay
- related code/docs evidence

## Outputs

- `plan.md`
- `checklist.md`
- KAH phase/gate events, when supported

## KAB plan capture rule

When a KAB planner lane is used, do not manually reconstruct the plan from chat. Ask the backend planner to include a clearly delimited `KHS Checklist Seed` section inside the plan, then capture the backend plan surface before implementation starts:

Required planner prompt clause:

```text
Include a section titled "KHS Checklist Seed" with unchecked markdown tasks. Each task must include phase, expected evidence artifact, and verification/check condition. Do not start implementation.
```

Then read the plan:

```bash
kkachi-agent-bridge plan start --session <session_id> "Plan only: propose the change; do not edit until approved"
kkachi-agent-bridge plan read --session <session_id>
# or: GET /api/plan/<session_id>
```

Copy the response field `plan.plan_text` into `plan.md`. Preserve `plan.plan_state`, `plan.plan_ref`, and `plan.source_evidence` in `plan.md` and/or `bridge-events.md`.

Backend timing rules:

- Claude / GLM / Codex: capture while `plan.plan_state=plan_pending_approval`; `plan approve` starts implementation immediately.
- Gemini / OpenCode: capture while `plan_pending_approval` or `plan_approved_waiting_for_start`; after approval, run `plan start-approved` only after `plan.md` and `checklist.md` are preserved.
- OpenCode: fail closed on `PLAN_AMBIGUOUS`; resolve `.sisyphus/plans/*.md` ambiguity before approving or starting.

Safe order:

1. `plan start`
2. `plan read`
3. save `plan.plan_text` to `plan.md`
4. KHS/Hermes derives `checklist.md` from the saved plan, the `KHS Checklist Seed` section, KHS phase contract, task contract, acceptance criteria, and expected evidence
5. then approve/start according to backend timing

## Checklist rule

`checklist.md` is mandatory, not optional. KAB does not directly provide the normalized KHS checklist. KAB provides `plan.plan_text`, and KHS should ask the planner to include a parse-friendly `KHS Checklist Seed` inside that plan. KHS/Hermes must transform that seed plus the KHS phase contract into a normalized KHS progress checklist and store it as the KAH `checklist.md` artifact during the plan phase, after `plan.plan_text` is captured. Then update it after `ask` and after every later phase. The checklist is the operator-facing progress tracker for the KHS run.

The checklist must include:

- one row for every canonical KHS phase;
- required/conditional applicability for each phase;
- current state (`pending`, `in_progress`, `done`, `skipped`, or `blocked`);
- owner role (`planner`, `implementer`, `feedback`, or `Hermes`);
- backend/session when a KAB lane is used;
- evidence artifact expected for completion;
- gate/check command or review condition;
- explicit skip reason for any skipped or not-applicable phase;
- micro-task rows derived from the approved plan.

For code-change runs, include an `optimize` row by default. It may be skipped only with a reason. For feedback, include round 1 as required and rounds 2-3 as conditional; do not exceed three feedback/handle-feedback pairs.

## Gate

PASS when scope, non-scope, affected surfaces, backend capability needs, verification strategy, docs impact, rollback/blocker conditions, and the mandatory progress checklist are explicit. The KAH plan gate requires `acceptance-criteria.md`, `plan.md`, and `checklist.md`.
