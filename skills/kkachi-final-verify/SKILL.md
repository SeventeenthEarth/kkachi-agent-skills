---
name: kkachi-final-verify
description: Check that all Kkachi gates, artifacts, verification evidence, docs decisions, feedback handling, and final report requirements are complete before reporting to the master.
version: 0.1.0
---

# Kkachi Final Verify

Use this skill before the final Korean report.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Completion is artifact-backed and Hermes-owned. Do not claim done when required gates, tests, docs decisions, KAH state, KAB evidence, `phase-plan.yaml`, or `checklist.md` are incomplete. Final verification must prove every required phase is done, every skipped/not-applicable phase has a reason, feedback rounds are between one and three, and code-change runs include optimize evidence or an explicit skip reason.

For KAB-backed phases, final evidence must include the selected bridge observation path:

- `cli_loop`: `send`, `wait`, `read/status`, pending resolutions, and `stop` evidence.
- `retained_stream`: `/api/stream/<session_id>` or `/api/events/<session_id>` evidence, pending resolutions, final `read/status`, and `stop` evidence.
- `hybrid`: stream evidence plus fallback `wait/read/status` evidence.

Final verification must reject any run where the only bridge proof is `send` success.

Final verification must also confirm selected backend caveats were handled:

- Gemini/OpenCode plan approvals that require explicit start have matching start evidence.
- OpenCode question-flow claims are backed by real upstream API/SSE question events.
- Codex evidence is wrapper/API-derived and does not expose raw Codex app-server payloads as public contract.
- Any GLM reject turn with a response-fidelity warning is called out in the final report.

## Outputs

- `final-report.md`
- `phase-plan.yaml` final state check
- `checklist.md` final state check
- final gate verdict
- Korean report source summary
- `kkachi-agent-helper gate check <run_id> final --json` result
- `kkachi-agent-helper run close <run_id> --json` for successful completion, or `run abort` for abandoned work
