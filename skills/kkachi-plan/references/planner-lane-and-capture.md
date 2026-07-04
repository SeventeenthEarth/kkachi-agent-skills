# Planner Lane And Capture

This reference expands the planner-lane procedure in `../SKILL.md`.

## Plan-first loop requirements

1. Refresh CodeGraph evidence before planning.
2. Ask the selected v0.2 planner lane for a plan-only response; do not allow implementation before plan capture. The default KAS/KAH lane is GJC `ralplan` candidate output under KAH evidence. KAB planners are explicit-only and require current capability plus bridge evidence. Historical Stage/Codex planner evidence must not be used as active proof unless a later approved task explicitly selects and evidences that lane.
3. Copy the fixed plan into `.kkachi/runs/<run_id>/plan.md` and normalize `checklist.md` from the plan plus the KHS phase contract.
4. Blue vets the backend-produced plan for SOT alignment, acceptance criteria, evidence, phase rows, and verification clarity, but must not directly author or rewrite the substantive implementation plan unless 주군 explicitly requests direct Blue planning or the work is outside the roadmap/KAS+KAH path.
5. Require Red and Orange plan vet/approval before implementation when active KAS/KAH roadmap policy requires it for code-development, source policy, workflow, template, test, or shared skill mirror work. Resolve Red/Orange reviewers and Gray documentation/integrity review through the project/team role registry when applicable rather than hard-coded individuals.
6. Route Blue, Red, Orange, or project-Gray `REQUEST_CHANGES` back to the same planner lane for a revised plan, then repeat the required vetting.
7. Only after the required Blue+Red+Orange plan approval and Blue synthesis, approve or start the implementer backend.

## KAB plan capture rule

When a KAB planner lane is explicitly used, do not manually reconstruct the plan from chat. Ask the backend planner to include a clearly delimited `KHS Checklist Seed` section inside the plan, then capture the backend plan surface before implementation starts.

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


## V01CLEAN active-baseline note

Any legacy Stage 1/Stage 2/Stage 3, direct Codex app-server, or KAB `native_codex` wording retained in this file is historical context only unless a later approved task explicitly selects KAB with current capability evidence. The active KAS/KAH v0.2 path is KAS policy + KAH deterministic evidence + approved GJC candidate artifacts, with KAT factual evidence only.
