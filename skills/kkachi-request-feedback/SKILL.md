---
name: kkachi-request-feedback
description: Prepare an independent feedback request for a Kkachi run, separate from red-team review, with clear scope, artifacts, questions, and read-only boundaries.
version: 0.2.0
---

# Kkachi Request Feedback

Use this skill when task risk warrants independent feedback from another AI lane or reviewer.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Feedback/request-review round 1 is required at least once for every KHS/KAS run that changes repository files, durable artifacts, workflow state, docs, release evidence, or commit readiness. It is independent, scoped, and read-only unless explicitly authorized. Round 1 should normally capture first Blue + Red/Orange/Gray color review evidence; non-implementation/docs-only runs are not exempt when they produce durable changes. Pure read-only explanations should usually avoid opening a KHS/KAS run, or mark review feedback `not_applicable` with an explicit reason.

Feedback requests for Codex/KAB-authored plans, implementation diffs, or review rounds must ask reviewers to audit fallback paths explicitly. The requested default is fail-closed behavior and removal of unnecessary fallbacks. Reviewers may accept a fallback only when no safe direct handling exists, it is narrowly bounded/evidenced, and the code/docs delta is genuinely small; broad fallback design or new policy/state machinery must be reported to 주군 as an option decision instead of accepted silently.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

MAR is the default independent review lane for `development` / implementation tasks unless 주군 explicitly waives or replaces it before start and the decision is recorded in KAH/run evidence artifacts. Run MAR after the first Blue + Red/Orange/Gray color review and feedback handling, then treat it as a later feedback round: capture provider preflight/toolchain proof, role attempts, raw bounded outputs, parsed findings, merge pack, Blue disposition, and any Red adjudication trigger. Required role coverage is `logic`, `security`, `arch`, `cve`, and `test_adequacy`; unresolved required role coverage fails closed. Providers must follow the role matrix primary/secondary lanes, preserve raw-output caps, avoid mutation, and must not turn provider availability, prompt rendering, or dispatch success into clean review coverage. Hermes may request other optional continuation rounds 2..5 when earlier feedback exposes unresolved risk, broad changes, or unclear verification; never exceed five request-feedback/handle-feedback pairs.

`delegate_task`, temporary subagents, and ad hoc advisor notes may inform the request but do not count as official color review, MAR role coverage, project-Gray review, or KAH evidence. For asynchronous Red/Orange/Gray review, use the watcher only as a mechanical observer and fall back to the durable Hermes Kanban CLI surface when direct Kanban tools are absent.

V02FLOW-007 review-train/watcher boundary: for substantial development work, preserve `first color review -> mandatory MAR -> second color adoption/review -> Blue disposition`. A color-round aggregate watcher is state-report-only; it must not perform Blue synthesis, fake `진행해`, auto-continue, waive lanes, mutate source, or substitute temporary subagents and delegate_task for official authority. temporary subagents and delegate_task do not count as official color review, MAR role coverage, or Blue synthesis.

## TWAKE-003 return-path policy

Feedback requests that dispatch async plan-vet, color review, MAR, second-color review, GJC long-running work, or blocked-condition probes must require a Blue return path or explicit degraded/no-wake evidence. Check effective KAH capability readback for `async_dispatch_return_path_evidence=true` and `async_dispatch_return_path_final_gate=true` before claiming KAH-backed return-path support. If reviewer fan-out lacks watcher/subscription/callback/origin evidence, record `blocked/degraded/no_wake_claim` with an operator-readable recovery hint and do not present the round as cleanly controllable. Required watcher reports must be terminal-only Blue-action-required output. watcher/notifier output is state-report-only and never review, MAR, waiver, Blue synthesis, or final acceptance authority.

See `references/mar-review-lane.md` for the KAS MAR role coverage, provider-preflight, merge-pack, disposition, and cleanup sequence.

## Outputs

- `feedback-request.md`
- `feedback-1.md`
- optional `feedback-2.md` through `feedback-5.md` as KHS supplemental artifacts when additional rounds run
- KAH phase/gate events, when supported
