# Review Readiness And Final Gate

This reference expands the late verification runbook in `../SKILL.md`.

## Teammate review readiness

Before declaring teammate review unavailable or blocked, check the durable Hermes Kanban CLI surface:

- `HOME=<real-user-home> hermes kanban --help`
- `hermes kanban boards current`
- `hermes profile list`
- `hermes kanban assignees --json`

When available, create and dispatch review cards through `hermes kanban create` and `hermes kanban dispatch --json`, then record Red/Orange/Gray card IDs and verdicts back into KAH evidence.

## Selected bridge observation path

For KAB-backed phases, final evidence must include the selected bridge observation path:

- `cli_loop`: `send`, `wait`, `read/status`, pending resolutions, and `stop` evidence
- `retained_stream`: `/api/stream/<session_id>` or `/api/events/<session_id>` evidence, pending resolutions, final `read/status`, and `stop` evidence
- `hybrid`: stream evidence plus fallback `wait/read/status` evidence

Reject any run where the only bridge proof is `send` success.

## Final artifact and gate checklist

- Produce `final-report.md`.
- Check `graph-evidence.md` when graph state affected the run or graph-managed workflow was requested.
- Confirm final `phase-plan.yaml` and `checklist.md` state.
- Preserve the final gate verdict, Korean report source summary, CodeGraph refresh evidence or explicit unavailable/degraded reason, final selected verification profile/gate evidence after the last relevant change, and the review-ready pre-commit repo state summary.
- Run `kkachi-agent-helper gate final <run_id> --json`; use `gate check <run_id> final --json` only as an older-helper compatibility fallback.

## Gate freshness and commit-approval sequence

- Final gate must run after the last artifact or evidence change. If `final-report.md`, `checklist.md`, or any referenced artifact changes after a final gate pass, rerun the affected prior gate and then rerun the final gate.
- Before asking for commit approval, use `pre-commit-completion-report-template.md`.
- When 주군 authorizes commit conditional on final Blue/Red/Orange/Gray approval, create a local Blue artifact, fan out evidence-pinned Kanban cards for Red/Orange/Gray, require explicit `ACCEPT` from all lanes, rerun final verification/gate after report updates, stage only intended files, commit with `HOME=<real-user-home> git commit ...`, then verify clean git state and record the commit handle.
- Close successful runs with `kkachi-agent-helper run close <run_id> --json`; abort abandoned work with `run abort`.
