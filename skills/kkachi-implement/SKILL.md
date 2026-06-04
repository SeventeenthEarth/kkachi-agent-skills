---
name: kkachi-implement
description: Drive the implementation phase through a selected KAB backend lane using prompt.md, then capture CLI output, diff evidence, implementation notes, and bridge events.
version: 0.1.0
---

# Kkachi Implement

Use this skill for Path A implementation after SOT, roadmap, task contract, backend selection, capability check, and prompt composition gates pass.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not implement from chat-only instruction. Do not claim KAB support unless the work is actually KAB-backed. When KAB runtime implementation is not ready or the task explicitly selects a no-KAB lane, the authorized direct Codex app-server lane through Hermes/KAS/KAH records direct Codex evidence instead of bridge evidence. Scoped KAS/KAH-local CLIMVP, GRAPHMVP, docs-only maintenance, and direct Codex app-server pilot work may proceed without KAB when explicitly authorized and recorded; such work must not claim KAB runtime support.

For KAS/KAH roadmap-task implementation on the current Codex app-server lane, Codex app-server/KAB implementer performs substantive code, test, build, and task-bound docs edits after the Codex plan has been vetted/approved by Blue and Red. Blue/Red/Orange/Gray do not directly patch repository implementation artifacts; they inspect, review, approve/reject, route feedback, record KAH evidence, and verify. Direct role editing is allowed only when 주군 explicitly asks for it or when the work is not a roadmap/KAS+KAH-path task; record the exception and no-Codex rationale before reporting completion.

If Blue/Red color review, Orange/Gray review, official GLM Octo review, or post-Octo re-review finds a required change, route the exact finding and evidence back to Codex app-server/KAB implementer for the fix. Blue synthesizes and approves the requested fix, but the repository mutation remains Codex-owned unless a recorded exception applies.

All terminal commands issued or requested during implementation must run with the real user home. In reusable artifacts use `HOME=<real-user-home>`. Do not run Git, tests, Codex, KAH/KAB, Hermes, or Kanban commands against a role-profile home unless the task explicitly tests profile isolation.

Implementation starts only after `plan`, `ask`, `phase-plan.yaml`, and `checklist.md` are complete. For KAB-backed lanes, backend selection, capability check, and prompt composition must also be complete. For the direct Codex app-server pilot lane, record the direct Codex session/prompt evidence and no-KAB rationale instead. Hermes may auto-start only low-risk work after notifying the master. Require explicit master approval for API, DB/schema/migration, security/auth/secrets, dependency, architecture, SOT, large diff/broad fanout, low confidence, or unresolved ask ambiguity.

For code-changing tasks, implementation is not complete until the repository aggregate verification command, normally `make test`, succeeds after the implementation changes. If `make test` fails, capture the failing command, exit status, and relevant output; route the evidence back to the implementer backend for analysis/fix; then rerun the failing target and the aggregate command.

Implementation-task review is not complete after first color review alone. After implementation changes, run first Blue + Red/Orange/Gray color review, handle any findings, then run official GLM Octo review as the next feedback round unless 주군 explicitly waives it before start. Official GLM Octo must be KAB-backed even if implementation used the direct Codex app-server/no-KAB lane: use a KAB GLM session, submit `/octo:review` as the first prompt command, capture `prompt_confirmed: true`, and record KAB session/readback/event evidence. Direct `glm` CLI review output is not official Octo evidence. After Octo feedback is handled, rerun affected verification and request post-Octo Blue + Red/Orange/Gray re-review before final/pre-commit reporting.

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
