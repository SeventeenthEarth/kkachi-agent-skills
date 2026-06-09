---
name: kkachi-implement
description: Drive the implementation phase through a selected KAB backend lane using prompt.md, then capture CLI output, diff evidence, implementation notes, and bridge events.
version: 0.1.0
---

# Kkachi Implement

Use this skill for Path A implementation after SOT, roadmap, task contract, backend selection, capability check, and prompt composition gates pass.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not implement from chat-only instruction. Do not claim KAB support unless the work is actually KAB-backed. KAS/KAH development uses a staged KAB adoption model: Stage 1 records authorized direct Codex app-server evidence without claiming KAB Codex execution; Stage 2 replaces those direct Codex calls with KAB Codex-first execution through `native_codex`; Stage 3 may select among eligible KAB backends such as Codex, Claude, and GLM after capability and policy gates. When KAB runtime implementation is not ready, the task explicitly selects Stage 1, or the work is explicitly KAS/KAH-local, record the no-KAB-Codex or local-lane rationale in task/phase artifacts. Scoped KAS/KAH-local CLIMVP, GRAPHMVP, docs-only maintenance, and explicitly authorized Stage 1 direct Codex app-server baseline work may proceed without KAB Codex when recorded; such work must not claim KAB runtime support.

For KAS/KAH roadmap-task implementation, the selected Codex/KAB implementer performs substantive code, test, build, and task-bound docs edits after the Codex or selected-backend plan has been vetted/approved by Blue and Red. In Stage 1 this implementer is the direct Codex app-server lane; in Stage 2 it is KAB Codex-first through `native_codex`; in Stage 3 it is the selected eligible KAB backend. Blue/Red/Orange/Gray do not directly patch repository implementation artifacts; they inspect, review, approve/reject, route feedback, record KAH evidence, and verify. Direct role editing is allowed only when 주군 explicitly asks for it or when the work is not a roadmap/KAS+KAH-path task; record the exception and no-Codex/backend rationale before reporting completion.

If Blue/Red color review, Orange/Gray review, official GLM Octo review, or post-Octo re-review finds a required change, route the exact finding and evidence back to the Stage 1 direct Codex app-server implementer, Stage 2 KAB Codex implementer, or Stage 3 selected KAB implementer for the fix. Blue synthesizes and approves the requested fix, but the repository mutation remains backend-owned unless a recorded exception applies.

Implementation backend product output must be English and compact for both Stage 1 direct Codex app-server and KAB-mediated lanes. The console report uses `Status`, `Summary`, `Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and `Next action requested`; detailed implementation notes, logs, diffs, findings, and file excerpts go to `.kkachi/runs/<run_id>/artifacts/implement/backend-implement.md` or the concrete requested phase artifact. If the artifact cannot be written, report `Status: blocked` with the artifact-write blocker and do not dump full plans, logs, diffs, files, reviews, or checklists into chat.

All terminal commands issued or requested during implementation must run with the real user home. In reusable artifacts use `HOME=<real-user-home>`. Do not run Git, tests, Codex, KAH/KAB, Hermes, or Kanban commands against a role-profile home unless the task explicitly tests profile isolation.

Implementation starts only after `plan`, `ask`, `phase-plan.yaml`, and `checklist.md` are complete. For KAB-backed lanes, backend selection, capability check, and prompt composition must also be complete. For Stage 1 direct Codex app-server baseline work, record the direct Codex session/prompt evidence and no-KAB-Codex rationale instead. Hermes may auto-start only low-risk work after notifying the master. Require explicit master approval for API, DB/schema/migration, security/auth/secrets, dependency, architecture, SOT, large diff/broad fanout, low confidence, or unresolved ask ambiguity.

For code-changing tasks, implementation is not complete until the repository aggregate verification command, normally `make test`, succeeds after the implementation changes. If `make test` fails, capture the failing command, exit status, and relevant output; route the evidence back to the implementer backend for analysis/fix; then rerun the failing target and the aggregate command.

Implementation-task review is not complete after first color review alone. After implementation changes, run first Blue + Red/Orange/Gray color review, handle any findings, then run official GLM Octo review as the next feedback round unless 주군 explicitly waives it before start. Official GLM Octo must be KAB-backed even if implementation used the direct Codex app-server/no-KAB lane: use a KAB GLM session, submit `/octo:review` as the first prompt command, constrain the prompt to requirements plus implemented code only, forbid tests/linters/builds/installs/package managers/network probes/service starts/runtime verification commands, capture `prompt_confirmed: true`, and record KAB session/readback/event evidence plus permission rejections for any out-of-scope command request. Direct `glm` CLI review output is not official Octo evidence, and an Octo run that executes a forbidden command fails closed unless 주군 explicitly waives the boundary. After Octo feedback is handled, rerun affected verification and request post-Octo Blue + Red/Orange/Gray re-review before final/pre-commit reporting.

KAB dispatch has two valid observation modes:

- CLI loop: `send` -> `wait` -> `read` or `status` -> resolve `approve` / `reject` / `answer` pendings -> repeat until terminal or accepted idle state -> `stop`.
- Retained stream loop: `send` -> subscribe to KAB HTTP `GET /api/stream/<session_id>` and/or backfill with `GET /api/events/<session_id>` -> resolve `approve` / `reject` / `answer` pendings as stream events expose them -> confirm completion with `read` or `status` -> `stop`.

`send` success only proves dispatch acceptance. Completion evidence requires either `wait` plus `read/status`, or retained stream evidence plus `read/status`.

Plan mode and pending control are backend-sensitive:

- Claude, GLM, and Codex plan approval can start execution immediately on supported idle-only plan lanes.
- Gemini and OpenCode plan approval records approval but requires explicit post-approval start before implementation.
- Plan approval never resolves tool/file/command/input pendings. Continue to handle `approve`, `reject`, and `answer` separately.
- OpenCode question handling is valid only for real upstream API/SSE question events, not rendered `<tool_call>question` text.

## Required artifacts before start

- `task-contract.yaml`
- `phase-plan.yaml`
- `plan.md`
- `checklist.md`
- `ask.md` or explicit no-unresolved-decisions record
- `selected-cli.json`
- `capability-check.md`
- `prompt.md`
- `graph-evidence.md` when graph state affects implementation scope or graph-managed workflow was requested

## Outputs

- `cli-output.md`
- `diff.patch`
- `impl-log.md`
- `bridge-session-snapshot.json`
- `bridge-events.md`
- `make test` evidence after implementation changes
- KAH phase/gate events, when supported

`bridge-events.md` must say which KAB observation mode was used: `cli_loop`, `retained_stream`, or `hybrid`. For stream-backed evidence, include stream cursor/epoch information when available and record daemon-restart discontinuities as evidence gaps, not as durable replay.

## Path B boundary

In Path B, this skill may only produce docs/SOT/handoff artifacts unless the run has passed into Path A.
