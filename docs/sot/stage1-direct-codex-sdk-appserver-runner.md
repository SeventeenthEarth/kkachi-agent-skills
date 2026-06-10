# Stage 1 direct Codex SDK app-server runner SOT

Date: 2026-06-10
Owner: KAS workflow/policy layer
Confirming role: Blue draft pending Red, Orange, and Gray review
Status: candidate SOT plus local runner-template draft for Stage 1 direct Codex SDK/app-server runner support; pending review/final gates before completion claim
Authority level: source-of-truth candidate for the Stage 1 direct Codex runner contract after confirmation
Scope: KAS Stage 1 direct Codex runner semantics, template contract, evidence requirements, and KAB boundary for `kkachi-hermes-skills`
Related docs:
- `docs/README.md`
- `docs/roadmap.md`
- `docs/sot/khs-architecture-and-integration.md`
- `docs/sot/kas-cli-contract.md`
- `skills/kkachi-orchestrate/references/run-operating-policy.md`
Evidence/source paths:
- Codex source reference: `/Users/draccoon/Workspace/Tools/ai-cli/codex/sdk/python/src/openai_codex/client.py`
- Codex SDK API reference: `/Users/draccoon/Workspace/Tools/ai-cli/codex/sdk/python/docs/api-reference.md`
- Codex app-server source reference: `/Users/draccoon/Workspace/Tools/ai-cli/codex/codex-rs/app-server/src/main.rs`
- KAB Codex adapter reference for Stage 2 boundary: `/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-bridge/internal/backend/codex_api/adapter.go`
- Local Codex probe: `codex --version` observed `codex-cli 0.138.0` during the Stage 1 runner investigation

## 1. Decision summary

KAS Stage 1 direct Codex work must use a KAS-owned Python runner template built on the `openai_codex` SDK. The runner is the supported Stage 1 way to access Codex app-server. Stage 1 must not call `codex exec`, must not use the generic `openai` Python SDK as if it were Codex app-server, and must not expose raw app-server `ws://`, `unix://`, or REST-like control as the ordinary KAS Stage 1 interface.

The Stage 1 runner contract is:

```text
KAS runner template -> openai_codex SDK -> codex app-server --listen stdio:// -> Codex JSON-RPC thread/run lifecycle
```

The `openai_codex` SDK is acceptable precisely because it starts and controls `codex app-server --listen stdio://` internally. A plain OpenAI chat/completions/responses client is not acceptable Stage 1 evidence.

## 2. ASIS facts

Current KAS docs already distinguish Stage 1 direct Codex app-server from Stage 2 KAB Codex-first execution. The gap is that Stage 1 has not yet named the concrete Python runner contract clearly enough, so operators and implementation backends can drift into `codex exec` or generic OpenAI SDK usage.

Observed Codex facts:

- The Python package namespace for the Codex SDK is `openai_codex`.
- The SDK client starts `codex app-server --listen stdio://` as a child process and communicates through JSON-RPC.
- Public SDK usage is based on `Codex`, `thread_start`, `thread_resume`, and `run`/turn execution APIs, not shelling out to `codex exec`.
- Raw app-server transports such as unix sockets and websockets exist for app-server/daemon/proxy use cases, but they are KAB/wrapper or explicit infrastructure work, not the default KAS Stage 1 runner interface.
- KAB Stage 2 Codex-first execution is separate: KAB uses its `native_codex` adapter and `kkachi-codex-wrapper` supervision/evidence path. KAS must not call that Stage 1.

## 3. TOBE target

KAS has added a shared Python runner template at:

```text
templates/runners/direct-codex-sdk-appserver-runner.py.tmpl
```

The template is copied or rendered into a run-local artifact before use. It provides one standard Stage 1 execution surface for plan, implementation, feedback fixes, task-bound docs, cleanup, and verification-support phases.

Required runner behavior:

1. Use `openai_codex` SDK imports only for Codex app-server control.
2. Launch Codex through the SDK-managed `codex app-server --listen stdio://` path.
3. Run with the real user home, for example `HOME=/Users/draccoon`, not a Hermes role-profile home.
4. Accept explicit run inputs: project directory, phase, prompt/artifact path, sandbox mode, model/reasoning settings when policy allows, and optional prior `thread_id`.
5. Use `Sandbox.read_only` for planning/review-like phases and `Sandbox.workspace_write` only for authorized mutation phases.
6. Preserve `thread_id` and runner metadata in KAH artifacts so later phases can resume the correct Codex thread when appropriate.
7. Write detailed backend output to the configured run artifact such as `.kkachi/runs/<run_id>/artifacts/<phase>/backend-<phase>.md` instead of dumping full detail into chat.
8. Exit non-zero and fail closed on missing Codex CLI, missing `openai_codex`, missing prompt artifact, unsupported sandbox, stale approval evidence, missing writable artifact path, or Codex/app-server startup failure.

## 4. Non-goals

This SOT does not authorize:

- treating the runner draft as accepted/completed before CODEXSDK-002 review and final gates pass;
- using `codex exec` as Stage 1 evidence;
- using the generic `openai` SDK as Stage 1 Codex app-server evidence;
- exposing raw app-server websocket, unix-socket, daemon, or REST-like transport as normal Stage 1 KAS guidance;
- replacing KAB `native_codex` or calling the KAB Stage 2 lane Stage 1;
- changing auth, tokens, model/provider/gateway config, live Codex config, KAB runtime config, or Hermes runtime settings;
- keeping a cross-project long-lived app-server daemon as default KAS behavior;
- bypassing plan, ask, approval, color review, GLM Octo, KAH gate, or final verification requirements.

## 5. Stage 1 lifecycle

Stage 1 should be task/phase scoped, not a hidden persistent service.

Default lifecycle:

1. KAS/KAH create or select the run artifact paths.
2. KAS renders the prompt and runner invocation record.
3. The runner starts an SDK-managed Codex app-server subprocess for the invocation.
4. The runner starts or resumes a Codex thread and executes the turn.
5. The runner records output, metadata, and evidence paths.
6. The runner exits and lets the SDK clean up the app-server child process.

Thread continuity comes from `thread_id` and KAH artifacts, not from assuming a long-lived app-server remains running across phases. A persistent daemon may be explored only as a separate KAB/wrapper or explicit infrastructure task.

## 6. Evidence contract

Every Stage 1 direct Codex runner invocation must preserve enough evidence to distinguish it from `codex exec`, generic OpenAI SDK usage, and KAB Codex execution.

Minimum evidence fields:

- runner template path and rendered runner path;
- `codex --version` or equivalent preflight output;
- `openai_codex` import/preflight evidence;
- real `HOME` used for the command;
- working directory/project path;
- KAS phase and run id;
- prompt artifact path and checksum when available;
- output artifact path;
- SDK/app-server mode: `openai_codex -> codex app-server --listen stdio://`;
- `thread_id` and whether it was started or resumed;
- sandbox mode;
- model/reasoning settings if supplied;
- exit code and bounded failure excerpt on failure;
- explicit no-KAB-Codex rationale for Stage 1 runs.

KAB evidence fields such as KAB session id, retained stream events, bridge events, `native_codex`, wrapper readiness, and KAB read/status/wait evidence must not be claimed for Stage 1 direct runner invocations.

## 7. Stage boundary with KAB

The runner is a Stage 1 bridge substitute only for KAS/KAH development while Stage 2 is not active. It is not a KAB adapter and does not prove KAB Codex support.

Stage boundary rules:

| Lane | Correct control path | Evidence claim |
|---|---|---|
| Stage 1 direct Codex SDK/app-server | KAS Python runner template using `openai_codex`; SDK spawns `codex app-server --listen stdio://` | direct Codex runner evidence plus no-KAB-Codex rationale |
| Stage 2 KAB Codex-first | KAB `native_codex` adapter and `kkachi-codex-wrapper` | KAB Codex session/bridge/read-status-wait/event evidence |
| Stage 3 KAB backend-selected | KAB selected eligible backend | selected backend KAB evidence and rejected-backend reasons |

If the installed/project KAS marker says Stage 2, the Stage 1 runner must not be used silently as fallback. The run must fail closed or request 주군 approval for an explicit stage exception.

## 8. Roadmap implementation gates

This SOT is a docs/spec foundation. Follow-on implementation must be split into separately reviewable tasks:

1. CODEXSDK-001: accept this SOT and roadmap/docs registration.
2. CODEXSDK-002: implement the Python runner template and minimal smoke docs/tests that prove it uses `openai_codex` and records required evidence.
3. CODEXSDK-003: update KAS phase skills/templates to consume the runner contract consistently and reject `codex exec`/generic OpenAI SDK drift.

CODEXSDK-002 and CODEXSDK-003 must include focused verification, docs-contract coverage, Red/Orange/Gray review, and final KAH evidence before completion claims.

## 9. Open questions

- Real Codex-turn smoke passed with `uv run --with 'pydantic>=2.12'` and the default Codex model/account settings; source-tree Python without that dependency still fails closed with metadata/error evidence.
- Model override policy remains cautious: a trial `--model gpt-5.1-codex-mini` returned a ChatGPT-account unsupported-model error, so runner guidance should avoid hard-coding model overrides unless a task explicitly verifies account/model eligibility.
- Whether the runner should be invoked only as a copied run artifact or also as a source-tree helper script for local operator use.

Until these are resolved in follow-up review/pilot work, the accepted Stage 1 facts are the control path, fail-closed preflight behavior, and evidence contract in this SOT.
