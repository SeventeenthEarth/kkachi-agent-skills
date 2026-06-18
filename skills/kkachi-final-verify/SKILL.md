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

- CodeGraph was refreshed at task start (`codegraph index <repo>` when `.codegraph/` exists, or `codegraph init -i <repo>` when source code already exists but no index exists), or a no-code bootstrap deferral / explicit unavailable-degraded reason was recorded.
- Implementation, test-enhance, AI-slop-cleaner, optimize, and docs-affecting passes each have evidence for the selected verification profile/gate after the final relevant change, or an explicit `not_applicable`/blocker reason. Do not assume a global `make test`; final verification must preserve the selected profile/gate id, command, timeout, applicability, status, exit code, duration, log path, log checksum, bounded failure excerpt, and deterministic failure extractor posture.
- Docs under `docs/` and the roadmap were updated, or a no-change/no-roadmap-update artifact explains why.
- Blue completed a first review and any actionable fixes were routed back to the selected implementer lane or responsible shaping/docs lane.
- Plan authorship and substantive repository mutations followed the selected implementer lane: Stage 1 direct Codex SDK/app-server runner through `templates/runners/direct-codex-sdk-appserver-runner.py.tmpl` (`openai_codex` -> `codex app-server --listen stdio://`), Stage 2 KAB Codex-first through `native_codex`, or Stage 3 selected KAB backend produced the plan draft/revisions and applied code/test/build/task-doc changes, while Blue/Red/Orange/Gray only supervised/reviewed/verified unless a 주군-approved direct-role exception is recorded. Stage 1 final evidence must not rely on `codex exec`, generic `openai` SDK output, raw app-server transport, or KAB `native_codex` evidence.
- Blue and Red plan vet evidence exists before implementation start, including verdicts and any `REQUEST_CHANGES` loop routed back through the selected planner lane. Color-review and GLM-review requested changes were routed back through the selected implementer lane for mutation, with Blue synthesis/approval and post-change verification evidence.
- Plan review, first color review, and any later feedback/review round recorded fallback audit when the plan or diff could introduce fallback behavior. Unnecessary fallback paths were removed or converted to fail-closed handling, and any retained fallback has a bounded/evidenced/small-delta rationale or an explicit 주군 decision path.
- Command evidence shows terminal commands used the real user home (`HOME=<real-user-home>` in reusable artifacts), including Git commands and especially `git commit`, so Git global config and user-level tool config were not accidentally read from a Hermes role/profile home.
- MAR evidence, when required or run, records role-first coverage for `logic`, `security`, `arch`, `cve`, and `test_adequacy`; selected primary/secondary provider lanes; provider preflight/toolchain proof; bounded raw-output artifacts; parse/disposition status; Blue disposition; and fail-closed handling for degraded, failed, or unresolved required roles. Provider availability, prompt rendering, or dispatch success alone is not review completion evidence.
- Red/Orange/Gray first color review evidence is captured when the KAS/KAH run changed durable repository artifacts, workflow state, docs, or release/commit evidence, even if the task was not classified as implementation. Pure read-only/no-durable-change runs may mark this not applicable only with a concrete reason.
- The repo remains review-ready and uncommitted until 주군 receives the report and approves commit.

Before declaring teammate review unavailable or blocked, check the durable Hermes Kanban CLI surface. See `references/review-readiness-and-final-gate.md` for the Kanban readiness commands, bridge observation-path evidence, final gate freshness rules, and commit-approval sequence.

Before reporting to the master for commit consideration, verify the repo is in review-ready uncommitted state unless the master has already approved commit. The final report must separate changed files, test evidence, KAH gate evidence, role-review evidence, risks, and the exact approval needed for install/commit.

For KAB-backed phases, final evidence must include the selected bridge observation path:

- `cli_loop`: `send`, `wait`, `read/status`, pending resolutions, and `stop` evidence.
- `retained_stream`: `/api/stream/<session_id>` or `/api/events/<session_id>` evidence, pending resolutions, final `read/status`, and `stop` evidence.
- `hybrid`: stream evidence plus fallback `wait/read/status` evidence.

Final verification must reject any run where the only bridge proof is `send` success.

Final product output from generated backend/report artifacts must be English and compact for Stage 1 direct Codex SDK/app-server runner and KAB-mediated lanes: `Status`, `Summary`, `Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and `Next action requested`. Detailed final evidence belongs in `final-report.md` or `.kkachi/runs/<run_id>/artifacts/final-verify/backend-final-verify.md`; if the detailed artifact cannot be written, report `Status: blocked` with the artifact-write blocker instead of dumping full logs, diffs, plans, files, reviews, or checklists into chat. The separate commander-facing Korean report to 주군 may summarize those English artifacts.

Final verification must also reject graph-managed workflow claims when the run lacks effective-binary evidence for `kkachi-agent-helper graph`, lacks `graph validate/explain` evidence for the graph file used, uses `kah graph` as if it were implemented, or describes manual `.kkachi-workflow.yaml` edits as graph repair. Missing graph capability is acceptable only when `graph-evidence.md` and the final report record a gap and state that the run used run-local phase evidence only.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

When graph state affected the run, final verification must check `graph-evidence.md` and the final report `kah_graph_evidence` section for the canonical GRAPHMVP-004 fields: `template_id`, `template_path`, `template_version`, `proposal_id`, `proposal_path`, `semantic_diff_output_path`, `validation_report_path`, `explain_report_path`, `approval_evidence_ref`, `audit_evidence_path`, `graph_checksum`, `graph_version`, `kah_graph_audit_event_ids`, and `capability_check_evidence`.

Final verification must also confirm selected backend caveats were handled:

- Gemini/OpenCode plan approvals that require explicit start have matching start evidence.
- OpenCode question-flow claims are backed by real upstream API/SSE question events.
- Codex evidence is wrapper/API-derived and does not expose raw Codex app-server payloads as public contract.
- MAR review is role-first and fail-closed. Reject final/pre-commit reports that substitute provider availability for required role coverage, omit provider preflight/toolchain evidence, omit bounded raw-output artifacts, omit Blue disposition, or report degraded/failed/unresolved required roles as clean PASS.
- Any provider response-fidelity or parse warning is called out in the final report.

## Outputs

- `final-report.md`
- `graph-evidence.md` check when graph state affected the run or graph-managed workflow was requested
- `phase-plan.yaml` final state check
- `checklist.md` final state check
- final gate verdict, Korean report source summary, CodeGraph refresh evidence or explicit unavailable/degraded reason, final selected verification profile/gate evidence after the last relevant change, and the review-ready pre-commit repo state summary
- `kkachi-agent-helper gate final <run_id> --json` result; use `gate check <run_id> final --json` only as an older-helper compatibility fallback
- before asking for commit approval, use the standardized pre-commit completion report format in `references/pre-commit-completion-report-template.md`; for implementation tasks include MAR role-coverage evidence and post-change re-review evidence when MAR feedback changed the work
- MAR evidence and any required post-change Blue/Red/Orange/Gray re-review are required for implementation tasks and for any other workflow where MAR or another later feedback round changed the work

See `references/review-readiness-and-final-gate.md` for the late-gate artifact checklist, final gate freshness rule, and close/abort sequence.
