---
name: kkachi-enhance-test
description: Analyze and improve test coverage for a Kkachi task after implementation or during shaping, preserving targeted regression rationale and verification evidence.
version: 0.1.0
---

# Kkachi Enhance Test

Use this skill when the run reaches the test enhancement phase.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Test work scales with risk. Add focused regression coverage for changed behavior where practical. If skipped, `phase-plan.yaml` and `checklist.md` must record why no test enhancement is useful or feasible.

## Outputs

- `test-plan.md`
- `test-log.md`
- updated diff when tests are added
- KAH phase/gate events, when supported
