---
name: kkachi-handle-feedback
description: Triage independent feedback, apply valid items, reject invalid items with evidence, rerun required verification, and preserve feedback handling artifacts.
version: 0.2.0
---

# Kkachi Handle Feedback

Use this skill after feedback is received.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not blindly apply feedback. Handle every requested feedback round. Separate valid issues, invalid suggestions, already-handled items, and out-of-scope requests. If feedback has no actionable items, still write the handling artifact and record `no actionable feedback`. A PASS/NIT/non-blocking review still needs artifacted disposition: apply the smallest valid in-scope improvement or record a concrete deferment reason. If MAR or another later feedback round changes the work after first color review, a fresh post-change Blue + Red/Orange/Gray re-review is required before final/pre-commit reporting.

For KAS/KAH roadmap-task work, feedback-driven code, test, build, or task-bound docs changes from Blue/Red/Orange/Gray color review, project-Gray documentation/integrity review, MAR role review, or post-change re-review must be routed back to the selected v0.2 implementer lane: approved GJC candidate/fix lane by default, or explicitly selected KAB lane when the task has current KAB capability and bridge evidence. Blue/Red/Orange/Gray synthesize and approve/reject findings, but do not directly patch repository implementation artifacts unless 주군 explicitly asks for direct role editing or the work is outside the roadmap/KAS+KAH path. Record any exception and no-GJC/KAB rationale in the handling artifact.


V02FLOW-009 routes `handle-feedback-*` remediation mutation through the selected GJC `ultragoal` executor lane by default after Blue accepts the finding scope. The remediation prompt must carry a finding bundle with finding id, source lane, severity/blocker state, Blue disposition, accepted scope, required reopened/amended phases, audited reopen/amend evidence when terminal phases are touched, refreshed verification, and focused re-review or MAR-refresh handoff refs. Missing KAH V02FLOW-010 capability/readback, missing ralplan/hash or approval refs, stale/absent verification, unsafe refs/checksums, KAB dispatch success as completion evidence, Blue/color source-patch fallback, and native GJC `ai-slop-cleaner` or `remove-ai-slop` requests fail closed. KAB dispatch success is dispatch evidence only, not completion evidence. Blue/color source-patch fallback is forbidden unless a recorded exception exists. KAT evidence is mechanical/factual only.

V02FLOW-013 tightens review-feedback remediation closeout: `implementation_goal_bundle_ready is goal-bundle-only and never sufficient for implementation completion`. Feedback remediation or `handle-feedback-*` mutation is not complete from implementation goal-bundle readiness, `start-ultragoal`, or goal-bundle-only evidence. `implementation_diff_ready` and `implementation_verified` require executor-loop evidence fields before review-feedback remediation can close: `changed_source_refs`, `diff_refs`, `checkpoint_ref`, `checkpoint_status`, `verification_output_refs`, `checksums`, `termination_reason`, `HOME`, and `no_authority_boundaries`. Missing executor-loop evidence fails closed instead of being filled by Blue/color source-patch fallback or dispatch-only proof.
## Feedback handling sequence

1. Preserve the feedback file or review card output as evidence; do not edit it unless explicitly asked.
2. Enumerate every finding and assign a disposition before making changes: valid, partially valid, invalid, already handled, deferred, or out of scope.
3. If the master asks only for review/triage, write a proposal artifact with dispositions, evidence checked, and recommended patch set; do not silently implement until requested or until the active KHS phase contract already authorizes handle-feedback changes.
4. For each valid or partially valid item, route the smallest in-scope change that addresses the durable issue to the selected v0.2 implementer lane. GJC applies candidate/fix work only when approved; explicitly selected KAB lanes apply fixes only with current capability and bridge evidence. Blue records the disposition/evidence.
5. For fallback-related feedback, prefer removal plus fail-closed diagnostics/evidence. Retain or add a fallback only when no safe direct handling exists, the fallback is narrowly bounded/evidenced, and the required code/docs delta is genuinely small. If the fallback requires broad code, new state machinery, or unclear policy, stop and report options to 주군.
6. For each rejected, deferred, or out-of-scope item, record the exact scope/evidence reason in the triage artifact.
7. After applying changes, rerun the verification gates affected by the feedback and capture raw logs.
8. For failed-test feedback, Blue owns reproduction/classification/routing/acceptance while the selected implementer owns detailed RCA, code/docs mutation, and rerunning affected verification.
9. Refresh the run diff/evidence artifacts and record KAH phase events.
10. Leave review-ready work uncommitted until the appropriate review/final gate or explicit 주군 commit approval.

## Outputs

- `feedback-triage-1.md`
- `handle-feedback-1.md`
- optional `feedback-triage-2.md` / `handle-feedback-2.md` as KHS supplemental artifacts when round 2 runs
- optional `feedback-triage-3.md` / `handle-feedback-3.md` as KHS supplemental artifacts when round 3 runs
- optional `feedback-triage-4.md` / `handle-feedback-4.md` as KHS supplemental artifacts when round 4 runs
- optional `feedback-triage-5.md` / `handle-feedback-5.md` as KHS supplemental artifacts when round 5 runs
- updated diff when valid feedback is applied
- rerun verification evidence when feedback changes code/docs
- KAH phase/gate events, when supported


## V01CLEAN active-baseline note

Any legacy Stage 1/Stage 2/Stage 3, direct Codex app-server, or KAB `native_codex` wording retained in this file is historical context only unless a later approved task explicitly selects KAB with current capability evidence. The active KAS/KAH v0.2 path is KAS policy + KAH deterministic evidence + approved GJC candidate artifacts, with KAT factual evidence only.
