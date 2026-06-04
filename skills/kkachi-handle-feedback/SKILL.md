---
name: kkachi-handle-feedback
description: Triage independent feedback, apply valid items, reject invalid items with evidence, rerun required verification, and preserve feedback handling artifacts.
version: 0.1.0
---

# Kkachi Handle Feedback

Use this skill after feedback is received.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not blindly apply feedback. Handle every requested feedback round. Separate valid issues, invalid suggestions, already-handled items, and out-of-scope requests. If feedback has no actionable items, still write the handling artifact and record `no actionable feedback`. A PASS/NIT/non-blocking review still needs artifacted disposition: apply the smallest valid in-scope improvement or record a concrete deferment reason. If Octo or another later feedback round changes the work after first color review, rerun affected verification and request a fresh Blue + Red/Orange/Gray re-review before final/pre-commit reporting.

## Feedback handling sequence

1. Preserve the feedback file or review card output as evidence; do not edit it unless explicitly asked.
2. Enumerate every finding and assign a disposition before making changes: valid, partially valid, invalid, already handled, deferred, or out of scope.
3. If the master asks only for review/triage, write a proposal artifact with dispositions, evidence checked, and recommended patch set; do not silently implement until requested or until the active KHS phase contract already authorizes handle-feedback changes.
4. For each valid or partially valid item, apply the smallest in-scope change that addresses the durable issue.
5. For each rejected, deferred, or out-of-scope item, record the exact scope/evidence reason in the triage artifact.
6. After applying changes, rerun the verification gates affected by the feedback and capture raw logs.
7. Refresh the run diff/evidence artifacts and record KAH phase events.
8. Leave review-ready work uncommitted until the appropriate review/final gate or explicit 주군 commit approval.

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
