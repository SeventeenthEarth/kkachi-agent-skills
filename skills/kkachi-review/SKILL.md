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
3. Blue/Red plan review and every color review must audit fallback behavior. Request removal of unnecessary fallback paths and prefer fail-closed diagnostics/evidence. Accept a fallback only when no safe direct handling exists, it is bounded and evidenced, and the required code/docs delta is genuinely small; if the fallback would require broad code or unclear policy, report to 주군 instead of approving it silently.
4. If changes are required, route them back to the implementer/backend or responsible operator and rerun affected verification before requesting final role reviews.
5. First Kkachi-team color review uses durable Kanban lanes when available: 하후연 for Red risk/fail-closed review, 여몽 for Orange operator/user workflow review, and 진궁 for Gray SOT/audit/evidence review.
6. For `development` / implementation tasks, run official GLM Octo after the first color review and feedback handling unless 주군 explicitly waives Octo before start. For non-implementation durable-change runs, run Octo only when explicitly requested by 주군, required by a project-local workflow, or opted in by a recorded high-risk policy gate. Official GLM Octo review scope is requirements plus implemented code only: prompts and permission handling must forbid tests, linters, builds, installs, package managers, network probes, service starts, and runtime verification commands while allowing read-only inspection of requirement artifacts, task contracts, plans/checklists, diffs, implemented source, docs, existing test files as implemented code evidence. After Octo feedback is triaged/applied/rejected, perform a second Blue + Red/Orange/Gray re-review before final/pre-commit reporting.
7. Synthesize every finding as reflected, rejected, deferred, or blocked with evidence. Do not treat a vague summary as sufficient review evidence.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

## Outputs

- `review.md`
- `redteam/<phase>-review.md` when used as red-team
- role review evidence for 하후연/여몽/진궁 when required by the active project policy
- verdict: `PASS|FAIL|BLOCKED`
- KAH gate event, when supported
