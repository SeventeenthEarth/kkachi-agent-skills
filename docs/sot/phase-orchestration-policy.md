# KHS Phase Orchestration Policy

Status: user-confirmed operating policy

This document records the current user-confirmed KHS behavior for software-development runs. The executable contract lives in `registries/phase-contracts.yaml`; this file explains the decisions for humans.

## Authority

- `phase-plan.yaml` is the KHS workflow source of truth for a run.
- KAH `work_path`, `work_mode`, `execution_mode`, `urgency`, and `sot_policy` are deterministic helper metadata. They may seed defaults, but they do not decide which KHS phases execute.
- KAH owns run IDs, artifacts, events, locks, schemas, diagnostics, and gates.
- KAB owns backend runtime sessions and public bridge evidence.

## Task selection

- The master selects the target roadmap task or task item for each KHS run.
- KHS must not auto-pick the next roadmap item by default.
- If project docs differ by repository, KHS reads the project overlay/docs map first.

## Hermes and backend roles

- Hermes is manager, risk approval router, final verifier, and Korean reporter.
- Hermes does not normally author production code inside a KHS run.
- KHS defines three logical backend roles:
  - planner: `plan`, `ask`
  - implementer: `implement`, `enhance-test`, `optimize`, `docs-update`, `handle-feedback`
  - feedback: `request-feedback`
- These logical roles may map to one physical KAB backend/session or to different backends according to project policy, user preference, Hermes judgment, and capability evidence.

## KAB usage

- KHS code-change and development runs use KAB.
- If the master wants code changes without KAB, the task becomes a normal direct Hermes task, not a KHS run.
- Docs-only KHS runs also use KAB by default.
- Direct docs editing is allowed only when the master explicitly forbids KAB; KHS records the no-KAB rationale in `phase-plan.yaml`, `docs-update.md`, and the final report.

## Canonical phase spine

```text
plan -> ask -> implement -> enhance-test -> optimize -> docs-update -> request-feedback -> handle-feedback -> final-verify -> improve
```

Mandatory for every KHS run:

- `plan`
- `ask`
- `request-feedback-1`
- `handle-feedback-1`
- `final-verify`

Conditional phases remain visible in `phase-plan.yaml` and `checklist.md`; skipped or not-applicable phases require explicit reasons.

## Plan and checklist

- When KAB planner mode is used, KHS asks the planner backend to include a clearly delimited `KHS Checklist Seed` section in `plan.plan_text`.
- KHS captures `plan.plan_text` into `plan.md` before implementation approval/start.
- KHS/Hermes parses the seed, applies `registries/phase-contracts.yaml`, and writes the normalized progress tracker to `checklist.md`.
- KAB does not own the normalized checklist.
- `checklist.md` is updated after `ask` and after every later phase.

## Ask phase

- `ask` always runs after `plan`.
- If no question remains, record that explicitly instead of skipping the phase.
- Ask the master only when Hermes cannot decide safely from SOT/project docs or when user intent must be confirmed.

## Implementation approval

Hermes may auto-start low-risk work after notifying the master. Explicit master approval is required for:

- API changes
- DB/schema/migration changes
- security/auth/secret-handling changes
- dependency changes
- architecture changes
- SOT changes
- large diffs or broad file fanout
- low confidence
- unresolved ask-phase ambiguity

## Enhance-test and optimize

- `enhance-test` is conditional; skipping requires a reason when no test enhancement is useful or feasible.
- For code-change runs, `optimize` is conditional but strongly recommended.
- Optimize focuses on AI slop removal, duplicated logic, dead/verbose code, and small structural waste.
- Skipping optimize in a code-change run requires a reason in both `phase-plan.yaml` and `checklist.md`.

## Feedback loop

- Request feedback at least once for every KHS run.
- Hermes may request up to two additional rounds when more feedback is useful.
- Maximum: three request-feedback/handle-feedback pairs.
- Each requested feedback round must have matching handling.
- If feedback has no actionable items, record `no actionable feedback`.

## Final verification

Hermes owns final verification. Before reporting completion, Hermes checks:

- `phase-plan.yaml` final state
- `checklist.md` final state
- KAH gate/final gate results
- KAB evidence, not just dispatch success
- required artifacts and skip reasons
- test/verification evidence
- docs-update decision
- feedback rounds between one and three
- optimize evidence or skip reason for code-change runs
