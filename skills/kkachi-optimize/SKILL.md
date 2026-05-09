---
name: kkachi-optimize
description: Perform bounded cleanup or optimization after behavior is understood and protected, preferring deletion, reuse, and project conventions over broad rewrites.
version: 0.1.0
---

# Kkachi Optimize

Use this skill after implementation and test coverage analysis identify low-risk cleanup or optimization work.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not optimize before behavior is anchored. For code-change KHS runs, optimize is conditional but strongly recommended: remove AI slop, duplicated logic, dead/verbose code, and small structural waste while keeping cleanup narrow, reversible, and tied to the approved task. Skipping optimize requires an explicit reason in `phase-plan.yaml` and `checklist.md`.

## Outputs

- `optimize-log.md`
- updated diff when cleanup is applied
- verification evidence after cleanup
- KAH phase/gate events, when supported
