---
name: kkachi-handle-feedback
description: Triage independent feedback, apply valid items, reject invalid items with evidence, rerun required verification, and preserve feedback handling artifacts.
version: 0.1.0
---

# Kkachi Handle Feedback

Use this skill after feedback is received.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not blindly apply feedback. Handle every requested feedback round. Separate valid issues, invalid suggestions, already-handled items, and out-of-scope requests. If feedback has no actionable items, still write the handling artifact and record `no actionable feedback`.

## Outputs

- `feedback-triage-1.md`
- `handle-feedback-1.md`
- optional `feedback-triage-2.md` / `handle-feedback-2.md` as KHS supplemental artifacts when round 2 runs
- optional `feedback-triage-3.md` / `handle-feedback-3.md` as KHS supplemental artifacts when round 3 runs
- updated diff when valid feedback is applied
- rerun verification evidence when feedback changes code/docs
- KAH phase/gate events, when supported
