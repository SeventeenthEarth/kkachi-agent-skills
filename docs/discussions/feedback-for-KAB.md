# Feedback for KAB Development Team

Date: 2026-05-12
From: KHS user review, phase orchestration design, and planner-output contract decision
Target: `kkachi-agent-bridge` (KAB)
Related projects: `kkachi-hermes-skills` (KHS), `kkachi-agent-helper` (KAH)
Status: discussion note to send to the KAB team; not KHS SOT until promoted

## Executive summary

KHS has settled the immediate planner-output policy:

- KAB owns backend session control, backend identity, plan lifecycle surfaces, and plan text/evidence.
- KHS owns workflow policy, phase applicability, and the normalized `checklist.md` artifact.
- KAH owns deterministic artifact/gate/state validation.
- Missing or malformed KAB `KHS Checklist Seed` is **not** by itself a KHS/KAB hard fail.
- KHS should recover by deriving `checklist.md` from `plan.plan_text`, the task contract, acceptance criteria, phase contract, expected evidence, and KAH gate requirements.
- KHS should be able to request backend-appropriate reasoning effort per command, and to retry safe failed work with escalated effort when the run has not crossed an unsafe side-effect boundary.
- Hard fail is reserved for missing/empty plan text, plan ambiguity, explicit KAB ambiguity evidence such as OpenCode `PLAN_AMBIGUOUS`, or inability to preserve required plan evidence before implementation starts.

Therefore the KAB request is not to become the owner of KHS checklist generation or retry policy. The KAB request is to keep plan lifecycle evidence stable, explicit, and safe to consume before any backend starts implementation, while exposing effort controls and effective-effort evidence where the backend supports them.

## Handoff request for the KAB team

Please treat this document as the handoff request from KHS to KAB. No separate interpretation layer is required.

Implement the requested changes in priority order:

1. Start with **P0-1**: stabilize the plan-read response contract for KHS preservation.
2. Then implement **P0-2**: make plan ambiguity and empty-plan failures explicit and fail-closed.
3. Then implement **P0-3**: document/test backend-specific approval/start semantics so KHS knows when approval starts implementation.
4. Continue through **P1** and **P2** items after P0 is complete, unless the KAB maintainer finds a smaller prerequisite change is needed first.

A good first KAB PR scope is:

- document the KHS-facing plan capture contract in KAB docs;
- ensure `plan read` / `/api/plan/{session_id}` reliably exposes `plan_state`, `plan_text`, `plan_ref`, `plan_backend`, `source_evidence`, `backend_type`, and `adapter_type` when a reviewable plan exists;
- state that KAB does not require or own KHS `checklist.md` normalization;
- protect fail-closed behavior for ambiguous or empty plan capture.

## Boundary to preserve

KAB must not:

- decide whether a user request should trigger KHS;
- decide KHS phase applicability;
- generate or validate KHS `checklist.md` as the owner of that artifact;
- decide KAH gate pass/fail;
- install Hermes skills;
- silently guess among ambiguous plan sources;
- silently continue when the reviewable plan text is unavailable.

KAB may:

- expose stable plan lifecycle state and plan evidence;
- surface backend ambiguity, unsupported behavior, and degraded behavior explicitly;
- preserve backend identity in every relevant response;
- provide retained events or snapshot evidence where the backend supports it;
- keep backend-specific plan approval/start semantics behind a common public lifecycle API;
- include optional planner-output hints in prompts, as long as KHS remains responsible for checklist normalization.

## Current KAB context observed from docs/source

Current bridge docs state:

- supported backends include Claude Code, GLM, Codex CLI, Gemini CLI, and OpenCode;
- each session carries explicit `backend_type` and `adapter_type`;
- public plan status includes `plan_state`, `plan_text`, `plan_ref`, `plan_backend`, `decision_options`, `post_approval_mode`, and `source_evidence`;
- OpenCode plan orchestration treats exactly one current `.sisyphus/plans/*.md` as the reviewable plan and fails closed as `PLAN_AMBIGUOUS` when multiple current plan files exist;
- Claude-compatible and Codex plan approval can start implementation immediately;
- Gemini and OpenCode use explicit post-approval start semantics, although Gemini may observe implementation pendings immediately after approval in some live paths;
- retained `/api/events` and `/api/stream` are the primary event surfaces for Codex, Gemini, and OpenCode, while Claude/GLM rely on `wait`, `read`, and `status` snapshots.

KHS can work with this model if KAB keeps the plan capture and failure contract stable.

## Requested changes

### P0-1. Stabilize the KHS-facing plan-read response contract

#### Problem

KHS must preserve KAB plan output into KAH `plan.md` before any approval/start action that may begin implementation. That requires a stable KAB response shape at the reviewable-plan point.

For KHS, the critical state is `plan_pending_approval` or an equivalent reviewable state where `plan_text` is available and implementation has not yet started.

#### Request

Document and test that `plan read` / `/api/plan/{session_id}` returns the following fields when a reviewable plan exists:

```json
{
  "session_id": "...",
  "state": "...",
  "backend_type": "...",
  "adapter_type": "...",
  "plan": {
    "plan_state": "plan_pending_approval",
    "plan_text": "...",
    "plan_ref": "...",
    "plan_backend": "...",
    "decision_options": [],
    "post_approval_mode": "...",
    "source_evidence": [
      {
        "kind": "...",
        "value": "...",
        "digest": "..."
      }
    ]
  }
}
```

Exact optional fields can remain backend-specific, but KHS needs these invariants:

- `backend_type` and `adapter_type` are present.
- `plan.plan_state` is present.
- `plan.plan_text` is non-empty when the plan is reviewable.
- `plan.plan_ref` or equivalent source reference is present when the backend has a plan file/reference.
- `plan.source_evidence` includes enough evidence for KHS to cite or hash the plan source.
- `decision_options` are present when approval requires a backend-specific option.

#### Acceptance criteria

- KHS can copy `plan.plan_text` into KAH `plan.md` without scraping TUI text.
- KHS can record `plan_ref` and `source_evidence` in run evidence.
- Tests protect non-empty `plan_text` at the reviewable-plan state.
- KAB docs identify this as the KHS/KAB plan capture contract.

### P0-2. Make ambiguity and empty-plan failures explicit and fail-closed

#### Problem

KHS can recover from a missing or malformed checklist seed, but it cannot safely recover from missing plan text or ambiguous plan sources.

KHS needs KAB to distinguish:

- usable plan with no checklist seed: recoverable by KHS;
- missing/empty plan text: hard fail;
- ambiguous plan source: hard fail;
- backend unsupported/degraded plan lifecycle: explicit unsupported/degraded status.

#### Request

Keep and document fail-closed error codes for plan capture failure. Existing `PLAN_AMBIGUOUS` behavior is the right model.

Recommended error categories:

- `PLAN_AMBIGUOUS` — multiple or ambiguous candidate plans; do not guess.
- `PLAN_EMPTY` or `PLAN_TEXT_UNAVAILABLE` — reviewable plan state exists but text cannot be captured.
- `PLAN_UNSUPPORTED` — backend/session does not support plan lifecycle.
- `PLAN_INVALID_STATE` — requested plan action is invalid for the current plan state.
- backend-specific drift or readiness errors where applicable.

#### Acceptance criteria

- KAB never returns a successful reviewable plan response with empty `plan_text`.
- Ambiguous source cases fail closed with machine-readable error code and useful evidence.
- OpenCode keeps exact-one-current-plan behavior for `.sisyphus/plans/*.md`.
- Gemini/Codex/Claude-compatible plan capture failures surface explicit errors rather than stale or guessed text.
- KHS can map these errors into hard-fail run evidence.

### P0-3. Document and test approval/start semantics per backend

#### Problem

KHS must preserve `plan.md` and derive `checklist.md` before calling any KAB action that may start implementation.

Backend semantics differ:

- Claude-compatible and Codex approval may start implementation immediately.
- Gemini and OpenCode normally require explicit post-approval start.
- Gemini has documented live behavior where implementation pendings may appear immediately after approval in some cases.

KHS needs a stable way to know whether approval starts execution or whether a separate start call is required.

#### Request

Document and test backend-specific plan approval/start semantics in the public KAB plan contract.

KHS-facing rules should be explicit:

- For backends where `plan approve` can start implementation, KHS must call `plan read` and preserve plan evidence before approval.
- For explicit-start backends, `plan approve` should report either `plan_approved_waiting_for_start` or a clearly documented `execution_started` branch when live runtime behavior already opened implementation.
- `plan start-approved` should be required only when the current `plan_state` is `plan_approved_waiting_for_start`.
- Plan approval remains separate from tool/file/command/input pending approval.

#### Acceptance criteria

- Compatibility docs identify approval/start behavior per backend.
- Tests cover immediate-start and explicit-start branches.
- KHS can decide whether `plan start-approved` is needed solely from KAB response state.
- KAB does not auto-approve implementation tool/file/command/input pendings as part of plan approval.

### P1-1. Keep backend identity and session snapshot evidence stable for KHS artifacts

#### Problem

KHS run evidence includes `selected-cli.json`, `capability-check.md`, `bridge-session-snapshot.json`, and `bridge-events.md`. KAB is the source of truth for session identity and backend runtime state.

#### Request

Document the recommended KHS evidence extraction path from existing KAB outputs:

- selected backend identity from compatibility matrix plus session response;
- session snapshot from `status`, `read`, or `list`;
- backend `backend_type` and `adapter_type` in every relevant response;
- `open_pendings`, state, lifecycle class, and backend caveats where available;
- retained events via `/api/events` or `/api/stream` when supported.

If current commands already expose this sufficiently, this can be a docs/test task rather than a new feature.

#### Acceptance criteria

- KHS can populate `bridge-session-snapshot.json` from documented KAB commands.
- KHS can cite the compatibility matrix as the source ledger for backend capability selection.
- KHS does not need to scrape tmux/TUI output for normal session evidence.
- TUI capture remains diagnostic only where KAB docs already define it that way.

### P1-2. Preserve plan lifecycle evidence in retained events where supported

#### Problem

For Codex, Gemini, and OpenCode, retained event surfaces are the preferred first-class evidence stream. KHS benefits if plan lifecycle transitions are visible in retained events or clearly recoverable through snapshots.

#### Request

For retained-event-capable backends, ensure plan lifecycle transitions can be audited through `/api/events` or `/api/stream`, or document exactly which snapshot calls are authoritative when plan events are not retained.

Useful event evidence includes:

- plan started;
- plan text/source captured;
- plan revised;
- plan approved;
- approved plan waiting for start;
- execution started;
- plan rejected;
- plan failed/ambiguous.

#### Acceptance criteria

- KHS can preserve `bridge-events.md` without relying on TUI scraping.
- Retained events and `plan read` snapshots reconcile on `session_id`, `backend_type`, `adapter_type`, and plan state.
- Restart/event-buffer discontinuity remains explicit, not silently hidden.

### P1-3. Add an optional KHS planner-output hint only as guidance

#### Problem

KHS may ask planner backends to include a `KHS Checklist Seed` section, but this is not a strict KAB contract. KHS will normalize `checklist.md` itself.

#### Request

If KAB owns reusable plan prompts or plan-start templates, it may include optional guidance such as:

```text
If useful, include a concise KHS Checklist Seed with proposed phase/evidence rows.
This section is advisory; the orchestrator will normalize the final checklist.
```

Do not make this a KAB parser requirement or backend compatibility gate yet.

#### Acceptance criteria

- Missing checklist seed does not make KAB plan lifecycle fail.
- KAB docs state that KHS owns final checklist normalization.
- Any future machine-readable checklist seed block is deferred until KHS dry-runs show repeated parser friction.

### P1-4. Expose reasoning-effort controls and retry-escalation evidence

#### Problem

KHS can choose different command strategies depending on task difficulty. Some tasks should run with low or default reasoning effort for speed, while ambiguous design/debugging work may need higher reasoning effort from the beginning. If a run repeatedly fails before an unsafe side-effect boundary, KHS should be able to reissue the task with increased reasoning effort instead of changing the task semantics.

KAB does not need to decide KHS retry policy, but KHS needs a stable way to request and audit backend effort settings.

#### Request

Expose a KHS-facing effort control where each backend supports it, and preserve what actually happened in session/run evidence.

Recommended request fields or CLI flags:

```json
{
  "reasoning_effort": "low|medium|high|backend_default",
  "retry_of": "optional previous KAB session/run id",
  "retry_reason": "optional short reason from KHS",
  "retry_policy_owner": "KHS"
}
```

Recommended response/evidence fields:

```json
{
  "requested_reasoning_effort": "high",
  "effective_reasoning_effort": "high|backend_default|unsupported",
  "reasoning_effort_supported": true,
  "reasoning_effort_source": "request|backend_default|adapter_default",
  "retry_of": "...",
  "retry_reason": "..."
}
```

Backend-specific mapping can remain adapter-owned. If a backend cannot support effort control, KAB should report `unsupported` or documented default behavior explicitly rather than silently pretending the request was honored.

KHS retry/escalation constraints:

- KHS may retry with higher effort for plan creation, review, debugging, or implementation attempts that have not crossed an unsafe side-effect boundary.
- KHS must not silently retry after partial implementation, external side effects, irreversible commands, or unclear worktree/session state.
- KHS should preserve the failed attempt evidence and link the new attempt through `retry_of`.
- KAB should not decide whether escalation is warranted; it only executes the requested effort setting and reports the effective setting.

#### Acceptance criteria

- KHS can request per-session or per-command reasoning effort without backend-specific prompt hacks.
- KHS can tell from KAB evidence whether the requested effort was honored, defaulted, or unsupported.
- Retry escalation can be audited through session snapshots/events and does not erase failed-attempt evidence.
- Unsupported effort control is explicit and non-deceptive.

### P2-1. Publish KHS/KAB compatibility metadata for release tags

#### Problem

KHS `main` can track latest-compatible KAB behavior, but KHS release tags need reproducible tested/recommended KAB versions and required capabilities.

#### Request

Coordinate with KHS release metadata so KAB docs can answer:

- which KAB version/commit was tested with a KHS release;
- which backend capabilities are required for KHS plan/implement/feedback runs;
- which backends have retained events;
- which backends require explicit post-approval start;
- which caveats affect KHS run evidence.

#### Acceptance criteria

- KAB compatibility matrix remains the canonical tested-support ledger.
- KHS release metadata can cite a KAB tested/recommended version or commit.
- KAB does not become responsible for installing KHS.

## Suggested implementation order for KAB

1. Stabilize/document the KHS-facing plan-read response contract.
2. Add or strengthen fail-closed ambiguity/empty-plan errors and tests.
3. Document/test approval/start semantics per backend.
4. Document the KHS evidence extraction path for session snapshots and retained events.
5. Add retained plan lifecycle event evidence where supported or document authoritative snapshot fallback.
6. Add optional planner-output hints only as non-binding guidance.
7. Expose reasoning-effort controls and retry-escalation evidence where supported.
8. Coordinate compatibility metadata for KHS release tags.

## Why this matters

These changes let KHS use KAB as a reliable backend runtime layer without making KAB responsible for KHS workflow artifacts:

- KAB remains the source of truth for backend runtime and plan lifecycle evidence.
- KHS preserves `plan.md` and owns `checklist.md` normalization.
- KAH deterministically validates the resulting artifacts and gates.

The result is safer plan preservation, fewer false hard-fails from planner formatting variance, clear hard-fails for real ambiguity, and cleaner run evidence for KHS dry-runs and future release compatibility.
