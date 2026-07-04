---
name: kkachi-optimize
description: Perform bounded cleanup or optimization after behavior is understood and protected, preferring deletion, reuse, and project conventions over broad rewrites.
version: 0.2.0
---

# Kkachi Optimize

Use this skill after implementation and test coverage analysis identify low-risk cleanup or optimization work.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not optimize before behavior is anchored. For code-change KHS runs, optimize is conditional but strongly recommended and has two micro-stages when a backend/code actor is available:

1. AI slop cleanup: ask the selected v0.2 implementer lane to remove AI-ish artifacts, overexplaining comments, speculative branches, noisy abstractions, ungrounded text, and awkward generated prose. The default lane is approved GJC candidate/fix work; KAB is explicit-only with current capability and bridge evidence. Run the selected verification profile/gate command after this stage if files changed; do not assume a global `make test`.
2. Structural optimization: ask for bounded duplicated-logic removal, unused-code removal, unnecessary abstraction-layer removal, naming cleanup, avoidable-complexity reduction, and small internal algorithm or data-flow simplification. Run the selected verification profile/gate command again after this stage if files changed.

Do not merge these into an unverified broad rewrite. Keep cleanup narrow, reversible, and tied to the approved task. Skipping optimize or either micro-stage requires an explicit reason in `phase-plan.yaml` and `checklist.md`.

## Outputs

- `slop-cleanup-log.md`
- `optimize-log.md`
- updated diff when cleanup is applied
- verification evidence after cleanup
- KAH phase/gate events, when supported


## V01CLEAN active-baseline note

Any legacy Stage 1/Stage 2/Stage 3, direct Codex app-server, or KAB `native_codex` wording retained in this file is historical context only unless a later approved task explicitly selects KAB with current capability evidence. The active KAS/KAH v0.2 path is KAS policy + KAH deterministic evidence + approved GJC candidate artifacts, with KAT factual evidence only.
