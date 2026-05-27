---
name: kkachi-implement
description: Drive the implementation phase through a selected KAB backend lane using prompt.md, then capture CLI output, diff evidence, implementation notes, and bridge events.
version: 0.1.0
---

# Kkachi Implement

Use this skill for Path A implementation after SOT, roadmap, task contract, backend selection, capability check, and prompt composition gates pass.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Do not implement from chat-only instruction. Do not bypass KAB for KHS code-change or development runs. If the master requires code change without KAB, convert to a normal direct Hermes task rather than continuing as KHS.

Implementation starts only after `plan`, `ask`, `phase-plan.yaml`, `checklist.md`, backend selection, capability check, and prompt composition are complete. Hermes may auto-start only low-risk work after notifying the master. Require explicit master approval for API, DB/schema/migration, security/auth/secrets, dependency, architecture, SOT, large diff/broad fanout, low confidence, or unresolved ask ambiguity.

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
- KAH phase/gate events, when supported

`bridge-events.md` must say which KAB observation mode was used: `cli_loop`, `retained_stream`, or `hybrid`. For stream-backed evidence, include stream cursor/epoch information when available and record daemon-restart discontinuities as evidence gaps, not as durable replay.

## Path B boundary

In Path B, this skill may only produce docs/SOT/handoff artifacts unless the run has passed into Path A.
