---
name: kkachi-implement
description: Drive the implementation phase through a selected KAB backend lane using prompt.md, then capture CLI output, diff evidence, implementation notes, and bridge events.
version: 0.1.0
---

# Kkachi Implement

Use this skill for Path A implementation after SOT, roadmap, task contract, backend selection, capability check, and prompt composition gates pass.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not implement from chat-only instruction. Do not claim KAB support unless the work is actually KAB-backed. KAS/KAH development uses a staged KAB adoption model: Stage 1 records authorized direct Codex SDK/app-server evidence through `templates/runners/direct-codex-sdk-appserver-runner.py.tmpl` (`openai_codex` -> `codex app-server --listen stdio://`) without claiming KAB Codex execution; Stage 2 replaces those direct Codex calls with KAB Codex-first execution through `native_codex`; Stage 3 may select among eligible KAB backends such as Codex, Claude, and GLM after capability and policy gates. Stage 1 must not use `codex exec`, generic `openai` SDK evidence, or silent fallback from a Stage 2 marker. When KAB runtime implementation is selected, preserve KAB session/read-status-wait or retained event evidence instead of direct-runner evidence. When KAB runtime implementation is not ready, the task explicitly selects Stage 1, or the work is explicitly KAS/KAH-local, record the no-KAB-Codex or local-lane rationale in task/phase artifacts. Scoped KAS/KAH-local CLIMVP, GRAPHMVP, docs-only maintenance, and explicitly authorized Stage 1 direct Codex SDK/app-server baseline work may proceed without KAB Codex when recorded; such work must not claim KAB runtime support.

For KAS/KAH roadmap-task implementation, the selected Codex/KAB implementer performs substantive code, test, build, and task-bound docs edits after the Codex or selected-backend plan has the required Blue synthesis and Red/Orange plan-vet approval. Red/Orange plan-vet reviewers and project-Gray documentation/integrity review are resolved by the project/team role registry when applicable, not hard-coded to individuals. In Stage 1 this implementer is the direct Codex SDK/app-server runner lane; in Stage 2 it is KAB Codex-first through `native_codex`; in Stage 3 it is the selected eligible KAB backend. Blue/Red/Orange/Gray do not directly patch repository implementation artifacts; they inspect, review, approve/reject, route feedback, record KAH evidence, and verify. Direct role editing is allowed only when 주군 explicitly asks for it or when the work is not a roadmap/KAS+KAH-path task; record the exception and no-Codex/backend rationale before reporting completion.

If Blue/Red/Orange/Gray color review, project-Gray documentation/integrity review, MAR role review, or any later feedback round finds a required change, route the exact finding and evidence back to the Stage 1 direct Codex SDK/app-server runner, Stage 2 KAB Codex implementer, or Stage 3 selected KAB implementer for the fix. Blue synthesizes and approves the requested fix, but the repository mutation remains backend-owned unless a recorded exception applies.

Implementation backend product output must be English and compact for both Stage 1 direct Codex SDK/app-server runner and KAB-mediated lanes. The console report uses `Status`, `Summary`, `Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and `Next action requested`; detailed implementation notes, logs, diffs, findings, and file excerpts go to `.kkachi/runs/<run_id>/artifacts/implement/backend-implement.md` or the concrete requested phase artifact. If the artifact cannot be written, report `Status: blocked` with the artifact-write blocker and do not dump full plans, logs, diffs, files, reviews, or checklists into chat.

All terminal commands issued or requested during implementation must run with the real user home. In reusable artifacts use `HOME=<real-user-home>`. Do not run Git, tests, Codex, KAH/KAB, Hermes, or Kanban commands against a role-profile home unless the task explicitly tests profile isolation.

Implementation starts only after `plan`, `ask`, `phase-plan.yaml`, and `checklist.md` are complete. In workflow-managed STRICT runs, implementation also starts only after a preserved dispatch packet names a KAH-ready node and `kkachi-agent-helper workflow node start --expect-revision <expected_start_revision>` succeeds; stale expected revisions, start failures, missing required outputs, or KAH complete failures are blockers and must not be converted into completion claims. For KAB-backed lanes, backend selection, capability check, and prompt composition must also be complete. For Stage 1 direct Codex SDK/app-server baseline work, render/copy `templates/runners/direct-codex-sdk-appserver-runner.py.tmpl`, record runner metadata, direct Codex session/prompt/output evidence, `thread_id`, and no-KAB-Codex rationale instead. Resume the same task's recorded Stage 1 Codex `thread_id` for implementation, feedback-fix, cleanup, and verification-support turns unless the thread is unavailable or unsafe; record any new-thread exception. Do not bind Codex continuity to the Discord/Hermes chat session, and do not reuse a prior task's Codex thread for a new KAS/KAH task. Do not pass an explicit `--model` by default; use the configured Codex account default model. Use reasoning effort `medium` for non-plan Stage 1 turns by default; plan-only turns remain `high` under the plan skill. Override model or effort only with an explicit task reason and artifacted rationale. Long implementation, feedback-fix, cleanup, and verification-support turns must use a Hermes-tracked background runner with completion notification and bounded polling/watch evidence; do not run them as a single foreground terminal call whose timeout can kill the wrapper before output/metadata artifacts flush. Foreground calls are acceptable for short preflight and bounded plan-only/review turns only. Hermes may auto-start only low-risk work after notifying the master. Require explicit master approval for API, DB/schema/migration, security/auth/secrets, dependency, architecture, SOT, large diff/broad fanout, low confidence, or unresolved ask ambiguity.

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
