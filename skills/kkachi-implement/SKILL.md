---
name: kkachi-implement
description: Drive the implementation phase through the selected GJC ultragoal executor lane or an explicitly approved KAB lane, then capture prompt, diff, verification, and deterministic evidence.
version: 0.2.0
---

# Kkachi Implement

Use this skill for Path A implementation after SOT, roadmap, task contract, backend selection, capability check, and prompt composition gates pass.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not implement from chat-only instruction. Do not claim KAB support unless the
work is actually KAB-backed. The active v0.2 KAS/KAH development path is:
KAS owns task contracts/prompt policy, KAH owns deterministic run/gate/evidence
state, and GJC `ultragoal` may produce implementation-candidate artifacts only
after the plan/vet gate and explicit approval evidence pass; `ask` is not a default normal phase. KAT evidence is
mechanical/factual only. KAB runtime/session control is explicit-only and must
carry current capability plus bridge evidence. Legacy Stage 1/Stage 2/Stage 3
Codex/KAB adoption wording and direct Codex app-server lanes are historical
context, not active operator guidance.

KAS/KAH/KAT repository self-development is a standing exception to this dogfood path unless 주군 explicitly reselects it. For KAS, KAH, or KAT source/skill improvements, 황충 performs the main implementation directly, then routes the result through official color review and fixes/re-review. Preserve tests, evidence, color review, and release/commit approval boundaries.

For KAS/KAH roadmap-task implementation, the selected GJC `ultragoal`/approved
implementer lane performs substantive code, test, build, and task-bound docs
edits only after the required plan-vet and explicit bounded approval. Red/Orange plan-vet
reviewers and project-Gray documentation/integrity review are resolved by the
project/team role registry when applicable, not hard-coded to individuals.
Blue/Red/Orange/Gray inspect, review, approve/reject, route feedback, record
KAH evidence, and verify. Direct role editing is allowed only when 주군
explicitly asks for it, when the work is outside the roadmap/KAS+KAH path, or
when an approved run-local exception records the no-GJC/Blue-direct-patch rationale. Implementation evidence must cite GJC `ultragoal`/executor-loop refs (`create-goals -> complete-goals -> execute-goal -> checkpoint -> verify -> repeat-or-terminate`) or that approved exception; missing both fails closed.

V02FLOW-009 routes `impl` mutation through the selected GJC `ultragoal` executor lane by default after bounded approval. Phase briefs for `impl` must include task/run id, approved scope, accepted-scope ref, ralplan artifact ref/hash, Blue plan-lock or implementation approval ref, selected ultragoal session/goal refs when available and required at dispatch time, changed-surface bounds, preservation locks, non-goals, expected evidence, verification commands, and stop/block conditions. Unknown or ambiguous aliases, unsupported dispatch shape, missing KAH V02FLOW-010 capability/readback, stale or missing ralplan/approval/scope refs, unsafe/out-of-run refs or checksums, stale/absent verification, KAB dispatch success as completion evidence, and native GJC `ai-slop-cleaner` or `remove-ai-slop` requests fail closed. KAB dispatch success is dispatch evidence only, not completion evidence. Blue/color source-patch fallback is forbidden unless a recorded exception exists. KAT evidence is mechanical/factual only.

V02FLOW-013/V02FLOW-015 tighten implementation closeout: `implementation_goal_bundle_ready is goal-bundle-only and never sufficient for implementation completion`. `implementation_diff_ready` requires executor-loop source diff, checkpoint, and checksum evidence; `implementation_verified` requires passed verification output after the executor loop. Implementation evidence for `impl` and known aliases (`implement`) must include selected executor lane `gjc_ultragoal_executor_loop_candidate`, phase/canonical phase, argv/process refs, real-user `HOME`, cwd, timestamps, exit code, `changed_source_refs`, non-empty `diff_refs`, `checkpoint_ref`, `checkpoint_status`, `verification_output_refs`, checksums, repeat/termination reason, and `no_authority_boundaries`; missing fields fail closed rather than being filled by `ultragoal create-goals`, goal-bundle-only evidence, unsupported-capability warnings, KAT/Blue/MAR/review/final authority overclaims, or a Blue/color patch fallback.

If Blue/Red/Orange/Gray color review, project-Gray documentation/integrity
review, MAR role review, or any later feedback round finds a required change,
route the exact finding and evidence back to the selected implementer lane for
the fix. Blue synthesizes and approves the requested fix, but repository
mutation remains lane-owned unless a recorded exception applies.

Implementation product output must be English and compact: `Status`, `Summary`,
`Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and
`Next action requested`. Detailed implementation notes, logs, diffs, findings,
and file excerpts go to `.kkachi/runs/<run_id>/artifacts/implement/backend-implement.md`
or the concrete requested phase artifact. If the artifact cannot be written,
report `Status: blocked` with the artifact-write blocker and do not dump full
plans, logs, diffs, files, reviews, or checklists into chat.

Implementation-stage reports to 주군 must follow `docs/sot/stage-report-contract.md`: describe old behavior vs new behavior, whether the accepted plan was applied as written or meaningfully changed, the changed surfaces grouped by purpose, whether `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, and `docs-update` ran or were not applicable, the KAH/KAT/selected test results, and any remaining risk or approval boundary.

All terminal commands issued or requested during implementation must run with
the real user home. In reusable artifacts use `HOME=<real-user-home>`. Do not
run Git, tests, Codex, KAH/KAB, Hermes, or Kanban commands against a role-profile
home unless the task explicitly tests profile isolation.

Implementation starts only after `plan`, `phase-plan.yaml`, `checklist.md`, and explicit bounded approval evidence are complete when required. Do not require a default `ask` phase. If implementation or feedback handling must reopen
or correct a completed/skipped/not_applicable phase, do not edit `phase-plan.yaml` directly and do not use ordinary `phase-plan set` for the corrective change. Use audited KAH phase-plan commands instead:
`kkachi-agent-helper phase-plan reopen <run_id> <phase-id> --from-status <current> --to-status <target> --reason <text> --evidence-ref .kkachi/runs/<run_id>/<evidence>`
or
`kkachi-agent-helper phase-plan amend <run_id> <phase-id> --kind <correction|supersede|rollback> --from-status <current> --to-status <target> --reason <text> --evidence-ref .kkachi/runs/<run_id>/<evidence>`.
KAH records deterministic state facts only; KAS/Blue/color/MAR/final remain semantic acceptance authority. In workflow-managed STRICT runs, implementation
also starts only after a preserved dispatch packet names a KAH-ready node and
`kkachi-agent-helper workflow node start --expect-revision <expected_start_revision>`
succeeds; stale expected revisions, missing required outputs, or KAH complete failures are blockers and must not be converted into completion claims. For
KAB-backed lanes, backend selection, capability check, and prompt composition
must also be complete.

For code-changing tasks, implementation is not complete until the selected verification profile/gate command succeeds after the implementation changes, or the gate is explicitly `not_applicable` with a reason in the phase plan. Do not assume a global `make test` command. Capture `selected_profile_id`, `selected_gate_id`, command, timeout, applicability, status, exit code, duration, log path, log checksum, bounded failure excerpt, and deterministic failure extractor posture. If the selected gate fails, route the compact evidence plus full-log artifact pointer back to the implementer backend for analysis/fix; then rerun the failing target and any selected aggregate/scoped gate required by the active profile.

Failed-test ownership is split: Blue reproduces or classifies the failure, routes compact evidence to the selected implementer lane, and accepts/rejects the fixed result. The implementer owns detailed RCA, code/docs mutation, and rerunning the affected verification target.

Implementation-task review is not complete after first color review alone. After implementation changes, run first Blue + Red/Orange/Gray color review, handle any findings, then run MAR role review as the default independent review round. MAR must preserve role-first coverage for `logic`, `security`, `arch`, `cve`, and `test_adequacy`, use only validated primary/secondary provider lanes, record degraded or unresolved role coverage fail-closed, and write bounded provider outputs plus Blue disposition evidence. After MAR feedback is handled, rerun affected verification and request any required post-change Blue + Red/Orange/Gray re-review before final/pre-commit reporting.

See `references/bridge-observation-and-start-rules.md` for the valid KAB observation modes, backend-specific post-approval start rules, required pre-start artifacts, and the Path B boundary details.

## Outputs

- `cli-output.md`
- `diff.patch`
- `impl-log.md`
- `bridge-session-snapshot.json`
- `bridge-events.md`
- selected verification profile/gate evidence after implementation changes
- KAH phase/gate events, when supported

`bridge-events.md` must say which KAB observation mode was used: `cli_loop`, `retained_stream`, or `hybrid`. For stream-backed evidence, include stream cursor/epoch information when available and record daemon-restart discontinuities as evidence gaps, not as durable replay.
