---
name: kkachi-request-feedback
description: Prepare an independent feedback request for a Kkachi run, separate from red-team review, with clear scope, artifacts, questions, and read-only boundaries.
version: 0.1.0
---

# Kkachi Request Feedback

Use this skill when task risk warrants independent feedback from another AI lane or reviewer.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Feedback request is required at least once for every KHS run. It is independent, scoped, and read-only unless explicitly authorized. Hermes may request optional continuation rounds 2..5 when earlier feedback exposes unresolved risk, broad changes, or unclear verification; never exceed five request-feedback/handle-feedback pairs.

## Outputs

- `feedback-request.md`
- `feedback-1.md`
- optional `feedback-2.md` through `feedback-5.md` as KHS supplemental artifacts when additional rounds run
- KAH phase/gate events, when supported
