---
name: kkachi-backend-select
description: Select a KAB backend lane from task requirements, project policy, capability evidence, compatibility matrix, user preference, and backend prompt profile rules, then produce selected-cli.json and capability-check.md.
version: 0.1.0
---

# Kkachi Backend Select

Use this skill after `task-contract.yaml` exists and before composing any KAB prompt. Do not use it for Stage 1 direct Codex SDK/app-server runner baseline work unless the run explicitly selects KAB or claims bridge evidence.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Capability and project-policy gates run before user preference. Preference ranks eligible KAB backends; it cannot make an ineligible backend safe. If KAB runtime is not implemented/ready for the project, select the explicit Stage 1 no-KAB-Codex direct Codex SDK/app-server runner baseline lane outside this KAB backend-selection skill and record the runner metadata/no-KAB-Codex rationale in task/phase artifacts.

Apply KAS's KAB adoption stage before broad backend selection:

- **Stage 1 — Direct Codex SDK/app-server runner baseline:** do not use this skill for the direct Codex lane unless the run separately selects KAB execution or claims bridge evidence. Record `templates/runners/direct-codex-sdk-appserver-runner.py.tmpl` evidence (`openai_codex` -> SDK-managed `codex app-server --listen stdio://`) and no-KAB-Codex rationale instead; do not use `codex exec`, generic `openai` SDK output, raw app-server transport, or KAB `native_codex` evidence as Stage 1 proof.
- **Stage 2 — KAB Codex-first execution:** select Codex through KAB (`backend_type=codex`, `adapter_type=native_codex`) by default for KAS/KAH plan/implementation/fix/docs lanes. Do not broaden to Claude/GLM implementation backend selection in Stage 2 unless 주군 explicitly assigns a Stage 3 backend-selection experiment. MAR remains the default review lane and does not change implementation backend selection.
- **Stage 3 — KAB backend-selected execution:** select among eligible KAB backends such as Codex, Claude, and GLM according to capability evidence, project policy, task requirements, and user preference after gates.

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
6. Copy KAB capability evidence from `registries/cli-capabilities.yaml`, including adapter type, tested CLI, last verified date, plan lifecycle, retained stream support, reasoning-effort policy when relevant, permission preconditions, and caveats.
7. Render `selected-cli.json` with KAH schema `0.1` required fields: `version`, `status`, `backend_type`, `adapter_type`, `source_ledger_ref`, and `caveats`, plus KHS extension fields from the KAB capability registry. The schema-owned value is `selected-cli.json.status=supported|degraded`; do not write lifecycle values such as `complete` into this JSON artifact.
8. Render `capability-check.md` with `Status: complete` and explicit selected `backend_type`, `adapter_type`, tested CLI, observation modes, plan lifecycle, and accepted caveats.
9. Validate selected CLI shape with `kkachi-agent-helper schema validate .kkachi/runs/<run_id>/selected-cli.json --schema selected-cli --json`.
10. Record a compact event with `kkachi-agent-helper event append artifact.updated --run <run_id> --payload '{"path":"selected-cli.json","phase":"backend-select"}' --json` when useful.
11. Run `kkachi-agent-helper gate check <run_id> backend --json` only after all backend evidence artifacts required by the run manifest exist; rely on the schema and backend gate to verify backend evidence completion.

Backend JSON status safety: KHS owns orchestration guidance and backend selection policy, KAB owns bridge runtime evidence, and KAH owns deterministic artifact/gate behavior. KAH v0.1.2 fails closed on schema-owned backend JSON status updates, but KHS operators must still not run `kkachi-agent-helper artifact set-status <run_id> selected-cli.json --status complete`; that would attempt to overwrite the selected CLI schema status and should be treated as an operator error.

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
