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

## Outputs

- `review.md`
- `redteam/<phase>-review.md` when used as red-team
- verdict: `PASS|FAIL|BLOCKED`
- KAH gate event, when supported
