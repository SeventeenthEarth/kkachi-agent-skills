---
name: kkachi-enhance-test
description: Analyze and improve test coverage for a Kkachi task after implementation or during shaping, preserving targeted regression rationale and verification evidence.
version: 0.1.0
---

# Kkachi Enhance Test

Use this skill when the run reaches the test enhancement phase.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Test work scales with risk, but code-changing KHS runs must explicitly assess unit, integration, and e2e coverage for the changed behavior. Add focused regression coverage where practical. If skipped, `phase-plan.yaml` and `checklist.md` must record why no test enhancement is useful or feasible.

Expected test lane meanings:

- `test-prepare`: lint, vet, formatting, guardrails, generation checks, and other pre-test validation.
- `test-unit`: isolated unit tests.
- `test-int` or `test-integration`: component integration with mocks/fakes/stubs and no production external resources.
- `test-e2e`: end-to-end checks in an isolated test environment that cannot affect live Hermes, Discord, gateway, auth, or production user state.

After test enhancement changes, run the relevant target(s) and then `make test`. Do not mark this phase complete without aggregate verification evidence unless the phase is explicitly skipped with a reason.

## Outputs

- `test-plan.md`
- `test-log.md`
- updated diff when tests are added
- `make test` evidence after test enhancement changes
- KAH phase/gate events, when supported
