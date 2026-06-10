---
name: kkachi-request-feedback
description: Prepare an independent feedback request for a Kkachi run, separate from red-team review, with clear scope, artifacts, questions, and read-only boundaries.
version: 0.1.0
---

# Kkachi Request Feedback

Use this skill when task risk warrants independent feedback from another AI lane or reviewer.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Feedback/request-review round 1 is required at least once for every KHS/KAS run that changes repository files, durable artifacts, workflow state, docs, release evidence, or commit readiness. It is independent, scoped, and read-only unless explicitly authorized. Round 1 should normally capture first Blue + Red/Orange/Gray color review evidence; non-implementation/docs-only runs are not exempt when they produce durable changes. Pure read-only explanations should usually avoid opening a KHS/KAS run, or mark review feedback `not_applicable` with an explicit reason.

Feedback requests for Codex/KAB-authored plans, implementation diffs, or review rounds must ask reviewers to audit fallback paths explicitly. The requested default is fail-closed behavior and removal of unnecessary fallbacks. Reviewers may accept a fallback only when no safe direct handling exists, it is narrowly bounded/evidenced, and the code/docs delta is genuinely small; broad fallback design or new policy/state machinery must be reported to 주군 as an option decision instead of accepted silently.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

GLM Octo is mandatory for `development` / implementation tasks unless 주군 explicitly waives Octo before start. Run official GLM Octo after the first Blue + Red/Orange/Gray color review and feedback handling, then treat it as a later feedback round: capture the Octo prompt/readback/feedback, triage and apply/reject findings, rerun affected verification, then require a second Blue + Red/Orange/Gray color re-review before final/pre-commit reporting. For non-implementation durable-change runs, run official GLM Octo only when 주군 explicitly requests it, a project-local workflow declares it required, or a recorded high-risk policy gate opts in. Official GLM Octo is a requirements-and-implemented-code-only review lane: the prompt and permission handling must explicitly forbid running tests, linters, builds, installs, package managers, network probes, service starts, or runtime verification commands. GLM/Octo may inspect requirement artifacts, task contracts, plans/checklists, diffs, implemented source, docs, existing test files as implemented code evidence, but it must not create new verification by executing commands. During permission handling, approve only read-only inspection commands that match the prompt scope. Hermes/KAB must reject any Octo permission request outside that read-only inspection scope and instruct GLM/Octo to continue from requirements plus implemented code; if a forbidden command is approved or executed, the official Octo gate fails closed unless 주군 grants an explicit after-the-fact waiver. Hermes may request other optional continuation rounds 2..5 when earlier feedback exposes unresolved risk, broad changes, or unclear verification; never exceed five request-feedback/handle-feedback pairs.

See `references/glm-octo-review-lane.md` for the KAB GLM session setup, preflight, prompt-confirmation, watcher, verdict parsing, and cleanup sequence.

## Outputs

- `feedback-request.md`
- `feedback-1.md`
- optional `feedback-2.md` through `feedback-5.md` as KHS supplemental artifacts when additional rounds run
- KAH phase/gate events, when supported
