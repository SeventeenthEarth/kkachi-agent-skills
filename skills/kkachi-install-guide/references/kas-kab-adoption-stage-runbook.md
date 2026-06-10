# KAS KAB adoption stage runbook

Use this runbook after reading the generated project/profile marker at
`references/kab-adoption-stage.md`. The marker records the selected operating
stage; this reference records what a KAS/KAH run must preserve as evidence.

This runbook does not activate Stage 2 by itself and is not KAB execution
evidence. KABADOPT-004 remains the first Stage 2 pilot unless a later approved
run explicitly changes that plan.

## Shared readback for every run

Before selecting planner or implementer lanes, record:

- marker path and readback result;
- numeric stage and canonical stage;
- selection source, selected-at value, and approval evidence;
- whether the marker was missing, unreadable, unsupported, or ambiguous.

If the marker is missing or ambiguous, fail closed to Stage 1 claims: direct
Codex evidence may be recorded, but KAB Codex execution must not be claimed.

## Stage 1 — direct Codex SDK/app-server runner baseline

Stage 1 is the default KAS/KAH development lane.

Required evidence posture:

- direct Codex SDK/app-server runner prompt, thread, output, and metadata evidence for plan,
  implementation, docs/fix, cleanup, and verification support through
  `templates/runners/direct-codex-sdk-appserver-runner.py.tmpl`
  (`openai_codex` -> SDK-managed `codex app-server --listen stdio://`);
- explicit rejection of `codex exec`, generic `openai` SDK output, raw
  app-server transport, or KAB `native_codex` evidence as Stage 1 proof;
- a no-KAB-Codex rationale explaining why KAB `native_codex` was not used;
- KAH task/phase artifacts, checklist, gates, tests, and final verification as
  required by the task class;
- an explicit statement that the run does not claim KAB Codex execution
  evidence.

Allowed KAB use in Stage 1 is limited to independently required review lanes,
such as official GLM Octo review, with their own KAB evidence. That review lane
does not turn the implementation lane into Stage 2.

## Stage 2 — KAB Codex-first execution

Stage 2 replaces direct Codex SDK/app-server runner execution for KAS/KAH implementation,
fix, and docs-bound work with KAB-backed Codex execution through `native_codex`.
Codex remains the selected implementation/planning backend; Stage 2 does not
broaden backend selection to Claude, GLM, or other KAB backends.

Required evidence posture:

- explicit project/profile approval that selected Stage 2;
- marker readback showing `stage2_kab_codex_first` before execution;
- KAB `native_codex` selected CLI and capability preflight evidence;
- KAB session id and the command/prompt handoff evidence;
- KAB plan/read/status/wait/watch evidence, or retained stream evidence with a
  final read/status snapshot;
- bridge events showing dispatch, observation, approvals/questions when present,
  and completion/reconciliation;
- KAH task/phase artifacts, checklist, gates, tests, and final verification as
  required by the task class.

Direct Codex SDK/app-server runner use during a Stage 2 run is break-glass only. It
requires recorded approval, rationale, scope, and evidence explaining why KAB
`native_codex` could not safely complete that slice. It must not be silent
fallback and must not be reported as KAB Codex execution evidence.

## Reserved Stage 3

`stage3_kab_backend_selected` is reserved. Do not select it, generate it, or
claim it under the current selector. Stage 3 requires a separate SOT/roadmap task
and capability evidence before KAS may choose among multiple KAB implementation
backends.
