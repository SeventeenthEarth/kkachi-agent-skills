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

Every entry must record created date, owner, last-reviewed date, status-change evidence, and terminal evidence/rationale before leaving `open` or `converted_to_task`.

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
- Blue disposition ref: <non-empty Blue synthesis/card/event/final-disposition ref>
- Additional approval/ref: <주군/reviewer gate ref, or N/A only when not applicable; never replaces Blue disposition ref>
- Terminal evidence/rationale: <required when status is resolved/rejected/stale>
```

## Open deferred findings

None recorded.

## Converted or resolved findings

None recorded.
