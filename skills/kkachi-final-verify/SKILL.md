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

- CodeGraph was refreshed at task start (`codegraph index <repo>` when `.codegraph/` exists, or `codegraph init -i <repo>` when first initialization is due), or an explicit unavailable/degraded reason was recorded;
- implementation, test-enhance, AI-slop-cleaner, optimize, and docs-affecting passes each have `make test` evidence after the final relevant change, or an explicit skip/blocker reason;
- docs under `docs/` and the roadmap were updated, or a no-change/no-roadmap-update artifact explains why;
- 황충 completed a first review and any actionable fixes were routed back to the Codex app-server/KAB implementer;
- 하후연/Red, 여몽/Orange, and 진궁/Gray review evidence is captured for final review when the task is a Kkachi/KAS development run;
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
- `kkachi-agent-helper run close <run_id> --json` for successful completion, or `run abort` for abandoned work
