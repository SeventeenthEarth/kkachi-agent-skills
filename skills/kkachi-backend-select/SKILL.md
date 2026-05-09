---
name: kkachi-backend-select
description: Select a KAB backend lane from task requirements, project policy, capability evidence, compatibility matrix, user preference, and backend prompt profile rules, then produce selected-cli.json and capability-check.md.
version: 0.1.0
---

# Kkachi Backend Select

Use this skill after `task-contract.yaml` exists and before composing any KAB prompt.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Capability and project-policy gates run before user preference. Preference ranks eligible backends; it cannot make an ineligible backend safe.

## Inputs

- `task-contract.yaml`
- project overlay backend policy
- `registries/backend-selection-policy.yaml`
- `registries/backend-prompt-profiles.yaml`
- `registries/cli-capabilities.yaml`
- `kkachi-agent-bridge/docs/public/compatibility-matrix.md`
- `kkachi-agent-bridge/README.md` summary view
- user backend preference or priority, if provided

## Flow

1. Eliminate backends that lack required task capabilities.
2. Apply project backend policy and execution-mode restrictions.
3. Apply user preference among remaining eligible backends.
4. Select the matching backend prompt profile.
5. Record rejected backends and reasons.
6. Copy KAB capability evidence from `registries/cli-capabilities.yaml`, including adapter type, tested CLI, last verified date, plan lifecycle, retained stream support, permission preconditions, and caveats.
7. Render `selected-cli.json` with KAH schema `0.1` required fields: `version`, `status`, `backend_type`, `adapter_type`, `source_ledger_ref`, and `caveats`, plus KHS extension fields from the KAB capability registry.
8. Render `capability-check.md` with `Status: complete` and explicit selected `backend_type`, `adapter_type`, tested CLI, observation modes, plan lifecycle, and accepted caveats.
9. Validate selected CLI shape with `kkachi-agent-helper schema validate .kkachi/runs/<run_id>/selected-cli.json --schema selected-cli --json`.
10. Record a compact event with `kkachi-agent-helper event append artifact.updated --run <run_id> --payload '{"path":"selected-cli.json","phase":"backend-select"}' --json` when useful.
11. Run `kkachi-agent-helper gate check <run_id> backend --json` only after all backend evidence artifacts required by the run manifest exist.

## KAB-specific guards

- Gemini is a supported KAB backend when project policy permits it. For first-class KHS consumption, prefer retained `/api/stream` observation with `/api/events` plus `read`/`status`/`wait` reconciliation.
- OpenCode is supported, but question flow is only API/SSE-routed for real upstream `question.asked` events. Do not select it when verified live CLI question production is a hard task requirement.
- Codex is wrapper/API-driven through `native_codex`; selected evidence must not depend on raw Codex app-server events.
- For Gemini and OpenCode plan lanes, record that explicit post-approval start is required.
- For supervised approval tasks, record backend-specific permission preconditions such as Claude default mode, GLM no `--yolo`, Codex managed args, OpenCode `opencode_permission_mode: ask`, and Gemini direct shell disabled.

## Outputs

- `selected-cli.json`
- `capability-check.md`
- optional KAH artifact update event
- KAH backend gate report when backend evidence is required by `run-metadata.json.required_artifacts`

## Gate

FAIL when the selected backend contradicts the compatibility matrix, project policy, or required capabilities. BLOCKED when the compatibility source ledger is missing or user authority is required.
