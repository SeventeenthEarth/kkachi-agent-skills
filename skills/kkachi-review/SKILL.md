---
name: kkachi-review
description: Run a Kkachi review phase against task contract, plan, selected backend evidence, rendered prompt, diff, tests, docs impact, and bridge evidence.
version: 0.1.0
---

# Kkachi Review

Use this skill for read-only review, red-team support, or independent KAB review lanes.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Review findings must be grounded in current artifacts, current diff, or reproducible evidence. Reviews must respect `phase-plan.yaml` as the workflow SOT and call out any missing checklist/evidence row instead of silently accepting incomplete phases.

Default review order for KAS/KAH runs:

1. Treat first color review as the normal required review layer for any KAS/KAH run that changes repository files, durable artifacts, workflow state, docs, or release/commit evidence — not only implementation/code tasks. Pure read-only explanations should usually not open a KAS/KAH run; if a run is opened but review is genuinely not applicable, mark the phase `not_applicable` with an explicit reason rather than silently skipping it.
2. Hermes/Blue performs a first review against the fixed plan, diff/artifact changes, tests or no-test rationale, docs, KAH evidence, and unauthorized-surface risks.
3. If changes are required, route them back to the implementer/backend or responsible operator and rerun affected verification before requesting final role reviews.
4. First Kkachi-team color review uses durable Kanban lanes when available: 하후연 for Red risk/fail-closed review, 여몽 for Orange operator/user workflow review, and 진궁 for Gray SOT/audit/evidence review.
5. For `development` / implementation tasks, run official GLM Octo after the first color review and feedback handling unless 주군 explicitly waives Octo before start. For non-implementation durable-change runs, run Octo only when explicitly requested by 주군, required by a project-local workflow, or opted in by a recorded high-risk policy gate. After Octo feedback is triaged/applied/rejected, perform a second Blue + Red/Orange/Gray re-review before final/pre-commit reporting.
6. Synthesize every finding as reflected, rejected, deferred, or blocked with evidence. Do not treat a vague summary as sufficient review evidence.

## Outputs

- `review.md`
- `redteam/<phase>-review.md` when used as red-team
- role review evidence for 하후연/여몽/진궁 when required by the active project policy
- verdict: `PASS|FAIL|BLOCKED`
- KAH gate event, when supported
