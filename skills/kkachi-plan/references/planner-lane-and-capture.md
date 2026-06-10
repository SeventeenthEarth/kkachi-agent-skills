# Planner Lane And Capture

This reference expands the planner-lane procedure in `../SKILL.md`.

## Plan-first loop requirements

1. Refresh CodeGraph evidence before planning.
2. Ask the Stage 1 direct Codex app-server planner, Stage 2 KAB Codex planner, or Stage 3 selected KAB planner for a plan-only response; do not allow implementation before plan capture.
3. Copy the fixed plan into `.kkachi/runs/<run_id>/plan.md` and normalize `checklist.md` from the plan plus the KHS phase contract.
4. Blue vets the backend-produced plan for SOT alignment, acceptance criteria, evidence, phase rows, and verification clarity, but must not directly author or rewrite the substantive implementation plan unless 주군 explicitly requests direct Blue planning or the work is outside the roadmap/KAS+KAH path.
5. Require Red plan vet/approval from 하후연 before implementation when the work is KAS/Kkachi code-development or another risk-bearing project run.
6. Route Blue or Red `REQUEST_CHANGES` back to the same planner lane for a revised plan, then repeat Blue/Red vetting.
7. Only after Blue+Red approval, approve or start the implementer backend.

## KAB plan capture rule

When a KAB planner lane is used, do not manually reconstruct the plan from chat. Ask the backend planner to include a clearly delimited `KHS Checklist Seed` section inside the plan, then capture the backend plan surface before implementation starts.

Required planner prompt clause:

```text
Include a section titled "KHS Checklist Seed" with unchecked markdown tasks. Each task must include phase, expected evidence artifact, and verification/check condition. Do not start implementation.
Include a section titled "Fallback Audit" that lists every proposed fallback path, says whether it should be removed or retained, and explains any retained fallback using the KAS fail-closed/default-no-fallback policy.
```

Then read and preserve the plan:

```bash
kkachi-agent-bridge plan start --session <session_id> "Plan only: propose the change; do not edit until approved"
kkachi-agent-bridge plan read --session <session_id>
```

Copy `plan.plan_text` into `plan.md` when a KAB planner lane is used. Preserve `plan.plan_state`, `plan.plan_ref`, and `plan.source_evidence` in `plan.md` and/or `bridge-events.md`. For explicitly authorized KAS/KAH-local planning, record the non-KAB plan source and do not claim KAB plan lifecycle evidence.

## Backend timing rules

- Claude / GLM / Codex: capture while `plan.plan_state=plan_pending_approval`; `plan approve` starts implementation immediately.
- Gemini / OpenCode: capture while `plan_pending_approval` or `plan_approved_waiting_for_start`; after approval, run `plan start-approved` only after `plan.md` and `checklist.md` are preserved.
- OpenCode: fail closed on `PLAN_AMBIGUOUS`; resolve `.sisyphus/plans/*.md` ambiguity before approving or starting.
