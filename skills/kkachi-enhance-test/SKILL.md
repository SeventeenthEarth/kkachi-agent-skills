---
name: kkachi-enhance-test
description: Analyze and improve test coverage for a Kkachi task after implementation or during shaping, preserving targeted regression rationale and verification evidence.
version: 0.2.0
---

# Kkachi Enhance Test

Use this skill when the run reaches the test enhancement phase.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Test work scales with risk, but code-changing KHS runs must explicitly assess unit, integration, and e2e coverage for the changed behavior. Add focused regression coverage where practical. If skipped, `phase-plan.yaml` and `checklist.md` must record why no test enhancement is useful or feasible.

V02FLOW-009 routes `test-enhance` mutation through the selected GJC `ultragoal` executor lane by default after bounded approval. The phase brief must carry accepted-scope, ralplan/hash, Blue approval, changed-surface bounds, preservation locks, focused regression expectations, and selected ultragoal session/goal refs required at dispatch time. Missing KAH V02FLOW-010 capability/readback, unsupported aliases, stale evidence, or a request for native GJC `ai-slop-cleaner` or `remove-ai-slop` must fail closed rather than falling back to Blue/color edits.

V02FLOW-013 tightens `test-enhance` closeout: `implementation_goal_bundle_ready is goal-bundle-only and never sufficient for implementation completion`. `implementation_diff_ready` and `implementation_verified` require executor-loop evidence fields before this mutation-capable phase can close: `changed_source_refs`, `diff_refs`, `checkpoint_ref`, `checkpoint_status`, `verification_output_refs`, `checksums`, `termination_reason`, `HOME`, and `no_authority_boundaries`. Goal-bundle-only evidence, stale verification, or missing executor-loop refs must fail closed.

Expected test lane meanings:

- `test-prepare`: lint, vet, formatting, guardrails, generation checks, and other pre-test validation.
- `test-unit`: isolated unit tests.
- `test-int` or `test-integration`: component integration with mocks/fakes/stubs and no production external resources.
- `test-e2e`: end-to-end checks in an isolated test environment that cannot affect live Hermes, Discord, gateway, auth, or production user state.

After test enhancement changes, run the relevant target(s) and then the selected verification profile/gate command. Do not assume a global `make test`. Do not mark this phase complete without selected profile/gate evidence unless the phase or gate is explicitly `not_applicable` with a reason.

## Outputs

- `test-plan.md`
- `test-log.md`
- updated diff when tests are added
- selected verification profile/gate evidence after test enhancement changes
- KAH phase/gate events, when supported
