---
name: kkachi-review
description: Run a Kkachi review phase against task contract, plan, selected backend evidence, rendered prompt, diff, tests, docs impact, and bridge evidence.
version: 0.2.0
---

# Kkachi Review

Use this skill for read-only review, red-team support, or independent KAB review lanes.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Review findings must be grounded in current artifacts, current diff, or reproducible evidence. Reviews must respect `phase-plan.yaml` as the workflow SOT and call out any missing checklist/evidence row instead of silently accepting incomplete phases.

Review-stage reports to 주군 must follow `docs/sot/stage-report-contract.md`: summarize Red/Orange/Gray/Blue verdicts, the core issue each color found or did not find, MAR status/role coverage/findings or waiver/N/A reason, second-color adoption status, what fixes were made from review feedback, which verification was rerun, and whether the result is release-ready, install-ready, commit-ready, or still held. If any feedback is deferred, briefly state what was deferred, why it was safe or necessary to defer now, and which future gate/task owns it. Do not report only that review cards completed.

For project-Gray documentation/integrity review, explicitly verify roadmap/task status rows and status values against accepted evidence, review gates, and commits. If a completed task still says `Planned`, `In Progress`, or `In Review`, or an incomplete task says `Completed`, record it as a traceability finding and request a status-value update before closeout.

Default review order for KAS/KAH runs:

1. Treat first color review as the normal required review layer for any KAS/KAH run that changes repository files, durable artifacts, workflow state, docs, or release/commit evidence — not only implementation/code tasks. Pure read-only explanations should usually not open a KAS/KAH run; if a run is opened but review is genuinely not applicable, mark the phase `not_applicable` with an explicit reason rather than silently skipping it.
2. Hermes/Blue performs a first review against the fixed plan, diff/artifact changes, tests or no-test rationale, docs, KAH evidence, and unauthorized-surface risks.
3. Required Blue+Red+Orange plan review and every color review must audit fallback behavior for active KAS/KAH roadmap policy work. Resolve Red/Orange plan-vet reviewers from the project/team role registry when applicable. Request removal of unnecessary fallback paths and prefer fail-closed diagnostics/evidence. Accept a fallback only when no safe direct handling exists, it is bounded and evidenced, and the required code/docs delta is genuinely small; if the fallback would require broad code or unclear policy, report to 주군 instead of approving it silently.
4. If changes are required, route them back to the implementer/backend or responsible operator and rerun affected verification before requesting final role reviews.
5. First Kkachi-team color review uses durable Kanban lanes when available: 하후연 for Red risk/fail-closed review, 여몽 for Orange operator/user workflow review, and project-Gray documentation/integrity review resolved through the project/team role registry when applicable.
6. For `development` / implementation tasks, run KAS MAR as the default independent review lane after the first color review and feedback handling unless 주군 explicitly waives or replaces it before start and the decision is recorded in KAH/run evidence artifacts. MAR review completion is role-first: required roles are `logic`, `security`, `arch`, `cve`, and `test_adequacy`; unresolved required role coverage fails closed and must be reported rather than replaced by undeclared providers. MARTL treats `REQUEST_CHANGES`, `BLOCKED`, `DEGRADED`, and `FAILED` as non-terminal: accepted changes must route to the selected implementer lane, affected verification must rerun, refreshed MAR evidence must be captured, and post-MAR Red/Orange/Gray review must close before final/pre-commit reporting. `PASS_WITH_FINDINGS` is terminal-candidate only after Blue disposition records each finding as fixed, non-blocking, rejected with rationale, or explicitly deferred with accepted risk.
7. For DESIGN Teal policy, check `project_has_teal_lane`, `ui_ux_change`, derived `teal_required`, and `teal_skip_reason` against the task contract. When Teal is required, `DESIGN_PLAN_GATE` and `DESIGN_FIDELITY_REVIEW` are separate evidence requirements. Color review, MAR, backend evidence, and helper notes must not substitute for a required Teal verdict.
8. Treat `delegate_task`, temporary subagents, and ad hoc advisor notes as pre-review analysis only; they do not count as official color review, project-Gray documentation/integrity review, MAR role coverage, Teal verdict evidence, or KAH evidence.
9. Synthesize every finding as reflected, rejected, deferred, or blocked with evidence. Do not treat a vague summary as sufficient review evidence.

V02FLOW-007 review-train/watcher boundary: for substantial development work, preserve `first color review -> mandatory MAR -> second color adoption/review -> Blue disposition`. A color-round aggregate watcher is state-report-only; it must not perform Blue synthesis, fake `진행해`, auto-continue, waive lanes, mutate source, or substitute temporary subagents and delegate_task for official authority. temporary subagents and delegate_task do not count as official color review, MAR role coverage, or Blue synthesis.

## TWAKE-003 return-path policy

Official async review fan-outs must be reviewable through a Blue return path before they are treated as controllable. For plan-vet, color review, MAR, second-color review, GJC long-running review/fix turns, and blocked-condition probes, check effective KAH capability readback for `async_dispatch_return_path_evidence=true` and `async_dispatch_return_path_final_gate=true` before accepting KAH-backed wake claims. Missing capability readback, watcher/subscription/callback id, origin/thread metadata for same-thread claims, terminal-only policy, or no-authority boundary must produce `blocked/degraded/no_wake_claim` with an operator-readable recovery hint. Required watcher reports must be terminal-only Blue-action-required output. watcher/notifier output is state-report-only and never review, MAR, waiver, Blue synthesis, or final acceptance authority.

V02FLOW-013/V02FLOW-015 review focus: verify that `implementation_goal_bundle_ready is goal-bundle-only and never sufficient for implementation completion`. For `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, and `docs-update`, reviewers must find `implementation_diff_ready` or `implementation_verified` executor-loop evidence fields before accepting closeout: selected executor lane `gjc_ultragoal_executor_loop_candidate`, phase/canonical phase, argv/process refs, real-user `HOME`, cwd, timestamps, exit code, `changed_source_refs`, non-empty `diff_refs`, `checkpoint_ref`, `checkpoint_status`, `verification_output_refs`, checksums, repeat/termination reason, and `no_authority_boundaries`. Goal-bundle-only wording, stale verification, missing executor-loop refs, unsupported-capability warning-only behavior, GJC/KAT/Blue/MAR/review/final authority overclaim, or Blue/color source-patch fallback without a recorded exception is a request-changes finding.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

## Outputs

- `review.md`
- `redteam/<phase>-review.md` when used as red-team
- role review evidence for 하후연/여몽/project-Gray when required by the active project policy
- verdict: `PASS|FAIL|BLOCKED`
- KAH gate event, when supported
