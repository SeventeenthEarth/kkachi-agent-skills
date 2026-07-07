# Deferred feedback ledger — kkachi-agent-skills

Status: planning ledger / no current open deferred findings
Owner: KAS contract and policy layer
Docs-map key: `deferred_feedback`
Scope: review, MAR, second-color, Gray integrity, and Blue synthesis findings that are real but deliberately not fixed in the current task because they are non-blocking and better handled by a later bounded task.

## Policy

This file is not a general wishlist. It is the canonical KAS-side ledger for non-blocking review or MAR feedback that Blue explicitly defers to a follow-up task.

- Blockers must not be deferred. A finding that blocks the current task's acceptance, final gate, MAR passability, authority boundary, safety posture, or reviewed acceptance criteria must be fixed in the current loop or the task must remain held.
- Defer is exceptional. The default disposition is `fix_now`; `defer_followup` requires a non-empty Blue disposition ref. Additional 주군/reviewer approval refs are separate and cannot replace Blue disposition evidence.
- Deferred entries must preserve provenance: originating task, source lane, source ref, why the issue matters, why it was not fixed now, and how to fix it later.
- Final reports that defer feedback must link the exact ledger entry, and each task-local deferred entry must record the reciprocal final-report ref so later refactor/cleanup work can recover the rationale without searching old review transcripts.
- KAH may validate/read back this ledger after V02FLOW-012, but KAS owns the contract and disposition vocabulary.

## Status lifecycle

- `open`: accepted non-blocking finding is intentionally deferred and not yet converted to a follow-up task.
- `converted_to_task`: a concrete follow-up task/card exists and carries the deferred id plus required gates.
- `resolved`: the item was fixed and verified by the required resumed gates.
- `rejected`: Blue later rejects the item as no longer valid, with rationale/ref.
- `stale`: the item is superseded by another task/SOT or no longer applies, with rationale/ref.

Lifecycle status is one of: open, converted_to_task, resolved, rejected, stale. Waiver is authority evidence, not lifecycle status.

Every entry must record created date, owner, last-reviewed date, status-change evidence, and terminal evidence/rationale before leaving `open` or `converted_to_task`.

## Compact entry checklist

Use this checklist before pasting or accepting an entry. The full template below remains canonical, but soldiers should keep each field short and point to artifacts instead of writing mini-reports.

1. **Eligibility:** Blocking current task: false is mandatory for every deferred entry. Blocker finding handling: fix_now_or_hold_current_task.
2. **Decision:** Blue disposition must be `defer_followup`; Blue disposition ref: required; N/A is invalid. Additional approval/ref can supplement but cannot replace the Blue ref.
3. **Lifecycle:** Status uses only `open`, `converted_to_task`, `resolved`, `rejected`, or `stale`; waiver is authority evidence, not lifecycle status.
4. **Hashes:** Logical/readback status hash: the KAS/KAH semantic readback hash for the ledger status; Byte-level ledger file SHA: the file checksum of `docs/deferred-feedback.md`. Record both separately when available.
5. **Refs:** Source finding ref: exact review/MAR/Gray/Blue/user source. Final-report reciprocal ref: final report or Blue synthesis that links this deferred id. Converted task ref: concrete follow-up card/roadmap row once converted, otherwise `N/A while open`.

### Compact example

```md
## DEFER-V02FLOW-011-001 — compact ledger examples

- Status: open
- Created: 2026-07-07
- Owner lane: KAS
- Last reviewed: 2026-07-07
- Status changed by/ref: N/A for new open entry
- Originating task: V02FLOW-011 / run-...
- Source phase: first-color-review
- Source lane/reviewer: Orange
- Source ref: t_example / artifact path
- Finding summary:
  - Template wording is operator-usable but could be shorter.
- Why it matters:
  - Long entries slow soldier closeout.
- Blocking current task: false
- Why not fixed now:
  - Better with next task after current accepted scope closes.
- Defer reason: better_with_next_task
- Proposed fix:
  - Add shorter examples in the next docs cleanup task.
- Suggested follow-up task: TBD
- Converted task ref: N/A while open
- Conversion date: N/A
- Converted by/ref: N/A
- Deferred ID copied into task: N/A
- Target repo/files:
  - kkachi-hermes-skills/docs/deferred-feedback.md
- Required gate when resumed:
  - docs-contract, color review, final gate as applicable
- Carried-forward gates: same as required gate unless Blue narrows them
- Final-report ref(s): run final report or Blue synthesis ref that links DEFER-V02FLOW-011-001
- Blue disposition: defer_followup
- Blue disposition ref: t_blue_synthesis_or_event_ref
- Additional approval/ref: N/A when not applicable
- Logical/readback status hash: sha256:<semantic-status-hash or N/A before KAH V02FLOW-012>
- Byte-level ledger file SHA: sha256:<file-sha or N/A before readback>
- Source finding ref: t_example / artifact path
- Final-report reciprocal ref: final-report.md#deferred-feedback
- Terminal evidence/rationale: required when status is resolved/rejected/stale
```

## Entry template

```md
## DEFER-<TASK-ID>-NNN — <short title>

- Status: open | converted_to_task | resolved | rejected | stale
- Created: <YYYY-MM-DD>
- Owner lane: KAS | KAH | KAB | cross-repo | other
- Last reviewed: <YYYY-MM-DD or N/A>
- Status changed by/ref: <Blue/KAS/KAH/card/event ref or N/A for new open entry>
- Originating task: <task id / run id>
- Source phase: plan-vet | first-color-review | MAR | second-color-review | Gray integrity | Blue synthesis | user
- Source lane/reviewer: Red | Orange | Gray | MAR:<role> | Blue | 주군
- Source ref: <card/artifact/path/line/ref>
- Finding summary:
  - <what was found>
- Why it matters:
  - <future risk or operator impact>
- Blocking current task: false
- Why not fixed now:
  - <scope too large / better with next task / requires separate design / opportunistic refactor / other>
- Defer reason: scope_too_large | better_with_next_task | requires_separate_design | opportunistic_refactor | external_dependency | other
- Proposed fix:
  - <bounded fix plan>
- Suggested follow-up task: <task id or TBD>
- Converted task ref: <Kanban/card/roadmap row/ref or N/A while open>
- Conversion date: <YYYY-MM-DD or N/A>
- Converted by/ref: <operator/card/event ref or N/A>
- Deferred ID copied into task: yes | no | N/A
- Target repo/files:
  - <repo/path>
- Required gate when resumed:
  - <tests, stale scans, review lanes, MAR refresh, final gate>
- Carried-forward gates: <same as required gate, or explicit narrowed gate set with Blue ref>
- Final-report ref(s): <run/task final report, Blue synthesis, or final closeout ref that links this entry>
- Blue disposition: defer_followup
- Blue disposition ref: <non-empty Blue synthesis/card/event/final-disposition ref; N/A is invalid>
- Additional approval/ref: <주군/reviewer gate ref, or N/A only when not applicable; never replaces Blue disposition ref>
- Logical/readback status hash: sha256:<semantic-status-hash or N/A before KAH V02FLOW-012>
- Byte-level ledger file SHA: sha256:<file-sha or N/A before readback>
- Source finding ref: <exact source review/MAR/Gray/Blue/user card, artifact, line, or event ref>
- Final-report reciprocal ref: <final report or Blue synthesis anchor that links this deferred id>
- Terminal evidence/rationale: <required when status is resolved/rejected/stale>
```

## Open deferred findings

None recorded.

## Converted or resolved findings

None recorded.
