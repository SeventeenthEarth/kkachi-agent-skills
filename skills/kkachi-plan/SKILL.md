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

## 주군 development-pipeline preference

When 주군 asks to run KAS/Kkachi development work and the task is classified as `development`, treat the preferred default as a Codex-led plan-first loop, with the transport determined by the current KAS KAB adoption stage:

- **Stage 1:** direct Codex app-server planner, with direct Codex evidence and no-KAB-Codex rationale.
- **Stage 2:** KAB Codex-first planner through `native_codex`; this replaces direct Codex app-server calls without changing the KAS/KAH phase or review scenario.
- **Stage 3:** KAB backend-selected planner after backend selection chooses an eligible backend from task requirements, project policy, capability evidence, and user preference after gates.

Official GLM Octo review is independent of these planning stages. It remains a KAB GLM feedback lane when required or requested; it does not change the plan/implementation backend and is not satisfied by selecting GLM as an implementation backend.

Plan-first loop requirements:

1. refresh CodeGraph evidence before planning;
2. ask the Stage 1 direct Codex app-server planner, Stage 2 KAB Codex planner, or Stage 3 selected KAB planner for a plan-only response; do not allow implementation before plan capture;
3. copy the fixed plan into `.kkachi/runs/<run_id>/plan.md` and normalize `checklist.md` from the plan plus KHS phase contract;
4. Blue vets the backend-produced plan for SOT alignment, acceptance criteria, evidence, phase rows, and verification clarity, but must not directly author or rewrite the substantive implementation plan in place of the Stage 1 direct Codex app-server planner, Stage 2 KAB Codex planner, or Stage 3 selected KAB planner output unless 주군 explicitly requests direct Blue planning or the work is outside the roadmap/KAS+KAH path;
5. require Red plan vet/approval from 하후연 before implementation when the work is KAS/Kkachi code-development or another risk-bearing project run;
6. route Blue or Red `REQUEST_CHANGES` back to the same planner lane for a revised plan, then repeat Blue/Red vetting;
7. only after Blue+Red approval, approve/start the implementer backend.

Render planner prompts so every command example uses the real user home, for example `HOME=<real-user-home> <command>` in reusable artifacts. This includes Git commands because commit-time global `user.name`, `user.email`, signing, hooks, and credential helpers must come from the user's real home.

Record deviations in `phase-plan.yaml`, `checklist.md`, and the final report instead of silently using a lighter path. For `research_evidence`, `docs_only`, `simple_command_report`, `bootstrap_config`, or `collaboration_review`, use the selected light spine from `task-contract.yaml`; do not manufacture implementation/test/optimize phases unless the classification changes to `development`.

Before backend planning for a code-changing or process-changing task, refresh CodeGraph evidence for the target repository. If `.codegraph/` already exists, run `codegraph index <repo>` and preserve `codegraph status <repo>` output. If CodeGraph is due for first initialization after the first completed task and `.codegraph/` is missing, run `codegraph init -i <repo>` and preserve status evidence. If CodeGraph is unavailable, record the missing capability as a blocker or degraded-evidence reason instead of silently planning from stale code context.

## Inputs

- `task-contract.yaml`
- SOT basis and roadmap trace
- project overlay
- related code/docs evidence
- CodeGraph status evidence or explicit unavailable/degraded reason when required

## Outputs

- `plan.md`
- `checklist.md`
- `graph-evidence.md` mapping requirement when the plan initializes, validates, explains, diffs, proposes, applies, or otherwise relies on graph state
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

Copy the response field `plan.plan_text` into `plan.md` when a KAB planner lane is used. Preserve `plan.plan_state`, `plan.plan_ref`, and `plan.source_evidence` in `plan.md` and/or `bridge-events.md`. For explicitly authorized KAS/KAH-local planning, record the non-KAB plan source and do not claim KAB plan lifecycle evidence.

Backend timing rules:

- Claude / GLM / Codex: capture while `plan.plan_state=plan_pending_approval`; `plan approve` starts implementation immediately.
- Gemini / OpenCode: capture while `plan_pending_approval` or `plan_approved_waiting_for_start`; after approval, run `plan start-approved` only after `plan.md` and `checklist.md` are preserved.
- OpenCode: fail closed on `PLAN_AMBIGUOUS`; resolve `.sisyphus/plans/*.md` ambiguity before approving or starting.

Safe order:

1. `plan start`
2. `plan read`
3. save `plan.plan_text` to `plan.md`
4. KHS/Hermes derives `checklist.md` from the saved plan, the `KHS Checklist Seed` section, KHS phase contract, task contract, acceptance criteria, and expected evidence
5. when project policy requires Red plan approval, request and resolve that review before implementation
6. then approve/start according to backend timing

## Checklist rule

`checklist.md` is mandatory, not optional. KAB does not directly provide the normalized KHS checklist. When a KAB planner lane is used, KAB provides `plan.plan_text`, and KHS should ask the planner to include a parse-friendly `KHS Checklist Seed` inside that plan. KHS/Hermes must transform that seed plus the KHS phase contract into a normalized KHS progress checklist and store it as the KAH `checklist.md` artifact during the plan phase, after the plan source is captured. Then update it after `ask` and after every later phase. The checklist is the operator-facing progress tracker for the KHS run.

The checklist must include:

- one row for every canonical KHS phase;
- required/conditional applicability for each phase;
- current state (`pending`, `in_progress`, `done`, `skipped`, or `blocked`);
- owner role (`planner`, `implementer`, `feedback`, or `Hermes`);
- backend/session when a KAB lane is used;
- evidence artifact expected for completion;
- gate/check command or review condition;
- explicit skip reason for any skipped or not-applicable phase;
- micro-task rows derived from the approved plan;
- CodeGraph refresh evidence when required;
- repeated `make test` checkpoints after implementation, test enhancement, AI slop cleanup, and optimization when those stages change files.

For code-change runs, include an `optimize` row by default. It may be skipped only with a reason. For feedback, include round 1 as required and rounds 2..5 as conditional continuation rounds; do not exceed five feedback/handle-feedback pairs.

## Gate

PASS when scope, non-scope, affected surfaces, backend capability needs, verification strategy, docs impact, rollback/blocker conditions, the mandatory progress checklist, and Blue+Red plan vet verdicts are explicit. The KAH plan gate requires `acceptance-criteria.md`, `plan.md`, and `checklist.md`; KAS/KAH roadmap code-development must also preserve Blue+Red approval or a recorded 주군-approved exception before implementation.
