# Bridge Observation And Start Rules

This reference expands the backend observation and startup details in `../SKILL.md`.

## KAB dispatch observation modes

- CLI loop: `send` -> `wait` -> `read` or `status` -> resolve `approve` / `reject` / `answer` pendings -> repeat until terminal or accepted idle state -> `stop`
- Retained stream loop: `send` -> subscribe to KAB HTTP `GET /api/stream/<session_id>` and/or backfill with `GET /api/events/<session_id>` -> resolve `approve` / `reject` / `answer` pendings as stream events expose them -> confirm completion with `read` or `status` -> `stop`

`send` success only proves dispatch acceptance. Completion evidence requires either `wait` plus `read/status`, or retained stream evidence plus `read/status`.

## Plan mode and pending control

- Claude, GLM, and Codex plan approval can start execution immediately on supported idle-only plan lanes.
- Gemini and OpenCode plan approval records approval but requires explicit post-approval start before implementation.
- Plan approval never resolves tool/file/command/input pendings. Continue to handle `approve`, `reject`, and `answer` separately.
- OpenCode question handling is valid only for real upstream API/SSE question events, not rendered `<tool_call>question` text.

## Required artifacts before start

- `task-contract.yaml`
- `phase-plan.yaml`
- `plan.md`
- `checklist.md`
- explicit bounded approval/question evidence when required (`approval.md`, `answered-decisions.md`, or `plan.md#Decision Clarifications`)
- `selected-cli.json`
- `capability-check.md`
- `prompt.md`
- `graph-evidence.md` when graph state affects implementation scope or graph-managed workflow was requested

## Path B boundary

In Path B, this skill may only produce docs/SOT/handoff artifacts unless the run has passed into Path A.
