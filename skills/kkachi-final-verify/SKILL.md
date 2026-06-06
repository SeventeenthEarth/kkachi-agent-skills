---
name: kkachi-final-verify
description: Check that all Kkachi gates, artifacts, verification evidence, docs decisions, feedback handling, and final report requirements are complete before reporting to the master.
version: 0.1.0
---

# Kkachi Final Verify

Use this skill before the final Korean report.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Completion is artifact-backed and Hermes-owned. Do not claim done when required gates, tests, docs decisions, KAH state, KAB evidence, graph capability evidence when graph-managed workflow is selected, `phase-plan.yaml`, or `checklist.md` are incomplete. Final verification must prove every required phase is done, every skipped/not-applicable phase has a reason, feedback rounds follow the active KAS policy, and code-change runs include optimize evidence or an explicit skip reason.

For 주군's KAS/Kkachi development pipeline, final verification happens before commit and must also confirm:

Before declaring teammate review unavailable or blocked, check the durable Hermes Kanban CLI surface. Absence of injected `kanban_*` worker tools in the current chat schema is not enough to block review. Use `HOME=<real-user-home> hermes kanban --help`, `hermes kanban boards current`, `hermes profile list`, or `hermes kanban assignees --json`, then create/dispatch review cards through `hermes kanban create` and `hermes kanban dispatch --json` when available. Record Red/Orange/Gray card IDs and verdicts back into KAH evidence.

- CodeGraph was refreshed at task start (`codegraph index <repo>` when `.codegraph/` exists, or `codegraph init -i <repo>` when source code already exists but no index exists), or a no-code bootstrap deferral / explicit unavailable-degraded reason was recorded;
- implementation, test-enhance, AI-slop-cleaner, optimize, and docs-affecting passes each have `make test` evidence after the final relevant change, or an explicit skip/blocker reason;
- docs under `docs/` and the roadmap were updated, or a no-change/no-roadmap-update artifact explains why;
- Blue completed a first review and any actionable fixes were routed back to the selected implementer lane or responsible shaping/docs lane;
- plan authorship and substantive repository mutations followed the selected implementer lane: Stage 1 direct Codex app-server, Stage 2 KAB Codex-first through `native_codex`, or Stage 3 selected KAB backend produced the plan draft/revisions and applied code/test/build/task-doc changes, while Blue/Red/Orange/Gray only supervised/reviewed/verified unless a 주군-approved direct-role exception is recorded;
- Blue and Red plan vet evidence exists before implementation start, including verdicts and any REQUEST_CHANGES loop routed back through the Stage 1 direct Codex app-server planner, Stage 2 KAB Codex planner, or Stage 3 selected KAB planner; Blue/Red did not directly rewrite the substantive plan except under a recorded 주군-approved exception;
- color-review and GLM-review requested changes were routed back through the selected implementer lane for mutation, with Blue synthesis/approval and post-change verification evidence;
- plan review, first color review, and any later feedback/review round recorded a fallback audit when the plan or diff could introduce fallback behavior; unnecessary fallback paths were removed or converted to fail-closed handling, and any retained fallback has a bounded/evidenced/small-delta rationale or an explicit 주군 decision path;
- command evidence shows terminal commands used the real user home (`HOME=<real-user-home>` in reusable artifacts), including Git commands and especially `git commit`, so Git global config and user-level tool config were not accidentally read from a Hermes role/profile home;
- official GLM Octo evidence, when required or run, came from a KAB GLM session only: KAB session id, backend type `glm`, selected CLI/capability evidence, bridge session/readback/events, real-user-HOME KAB/GLM path preflight, prompt first line `/octo:review`, `prompt_confirmed: true`, bounded watcher/completion evidence, and feedback artifact are present. Direct `glm` CLI output may be recorded as preflight only; if it is the review source, the Octo gate fails closed;
- Red/Orange/Gray first color review evidence is captured when the KAS/KAH run changed durable repository artifacts, workflow state, docs, or release/commit evidence, even if the task was not classified as implementation. Pure read-only/no-durable-change runs may mark this not applicable only with a concrete reason;
- the repo remains review-ready and uncommitted until 주군 receives the report and approves commit.

Before reporting to the master for commit consideration, verify the repo is in review-ready uncommitted state unless the master has already approved commit. The final report must separate changed files, test evidence, KAH gate evidence, role-review evidence, risks, and the exact approval needed for install/commit.

For KAB-backed phases, final evidence must include the selected bridge observation path:

- `cli_loop`: `send`, `wait`, `read/status`, pending resolutions, and `stop` evidence.
- `retained_stream`: `/api/stream/<session_id>` or `/api/events/<session_id>` evidence, pending resolutions, final `read/status`, and `stop` evidence.
- `hybrid`: stream evidence plus fallback `wait/read/status` evidence.

Final verification must reject any run where the only bridge proof is `send` success.

Final verification must also reject graph-managed workflow claims when the run lacks effective-binary evidence for `kkachi-agent-helper graph`, lacks `graph validate/explain` evidence for the graph file used, uses `kah graph` as if it were implemented, or describes manual `.kkachi-workflow.yaml` edits as graph repair. Missing graph capability is acceptable only when `graph-evidence.md` and the final report record a gap and state that the run used run-local phase evidence only.

When graph state affected the run, final verification must check `graph-evidence.md` and the final report `kah_graph_evidence` section for the canonical GRAPHMVP-004 fields: `template_id`, `template_path`, `template_version`, `proposal_id`, `proposal_path`, `semantic_diff_output_path`, `validation_report_path`, `explain_report_path`, `approval_evidence_ref`, `audit_evidence_path`, `graph_checksum`, `graph_version`, `kah_graph_audit_event_ids`, and `capability_check_evidence`.

Final verification must also confirm selected backend caveats were handled:

- Gemini/OpenCode plan approvals that require explicit start have matching start evidence.
- OpenCode question-flow claims are backed by real upstream API/SSE question events.
- Codex evidence is wrapper/API-derived and does not expose raw Codex app-server payloads as public contract.
- Official GLM Octo review is KAB-only. Reject final/pre-commit reports that substitute a direct `glm` CLI review, omit KAB GLM session evidence, omit `/octo:review` as the first submitted command, or lack `prompt_confirmed: true`/completion evidence.
- Any GLM reject turn with a response-fidelity warning is called out in the final report.

## Outputs

- `final-report.md`
- `graph-evidence.md` check when graph state affected the run or graph-managed workflow was requested
- `phase-plan.yaml` final state check
- `checklist.md` final state check
- final gate verdict
- Korean report source summary
- CodeGraph refresh evidence or explicit unavailable/degraded reason
- final `make test` evidence after the last relevant change
- review-ready pre-commit repo state summary
- `kkachi-agent-helper gate final <run_id> --json` result; use `gate check <run_id> final --json` only as an older-helper compatibility fallback
- KAH gate freshness behavior: final gate must be run after the last artifact/evidence change. If updating `final-report.md`, `checklist.md`, or any artifact referenced by a prior gate after a final gate pass, immediately rerun the affected prior gate and then rerun the final gate.
- before asking for commit approval, use the standardized pre-commit completion report format in `references/pre-commit-completion-report-template.md`; for implementation tasks include official GLM Octo evidence and post-Octo re-review evidence, and for non-implementation tasks include not-applicable reasons when Octo was not requested/declared/run.
- official GLM Octo evidence and the post-Octo Blue/Red/Orange/Gray re-review are required for implementation tasks and for any other workflow where Octo was requested/declared/run. Do not report “Octo review complete” or ask for commit approval on an Octo-required/run workflow until KAB GLM `/octo:review` actually ran or was explicitly waived before start by 주군, accepted findings were applied or deferred with reasons, affected verification reran, and Blue plus Red/Orange/Gray re-review happened after Octo fixes. For non-implementation tasks where Octo was not requested/declared/run, record it as not applicable. Direct `glm` CLI review is not a waiver and not official Octo evidence.
- when 주군 authorizes commit conditional on final Blue/Red/Orange/Gray approval, create a local Blue artifact, fan out evidence-pinned Kanban cards for Red/Orange/Gray, require explicit ACCEPT from all lanes, rerun final verification/gate after report updates, stage only intended files, commit with `HOME=<real-user-home> git commit ...` and the task-prefixed one-line message so global Git config is loaded from the user home, then verify clean git state and record the commit handle.
- `kkachi-agent-helper run close <run_id> --json` for successful completion, or `run abort` for abandoned work
