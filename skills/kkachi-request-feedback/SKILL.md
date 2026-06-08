---
name: kkachi-request-feedback
description: Prepare an independent feedback request for a Kkachi run, separate from red-team review, with clear scope, artifacts, questions, and read-only boundaries.
version: 0.1.0
---

# Kkachi Request Feedback

Use this skill when task risk warrants independent feedback from another AI lane or reviewer.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Feedback/request-review round 1 is required at least once for every KHS/KAS run that changes repository files, durable artifacts, workflow state, docs, release evidence, or commit readiness. It is independent, scoped, and read-only unless explicitly authorized. Round 1 should normally capture first Blue + Red/Orange/Gray color review evidence; non-implementation/docs-only runs are not exempt when they produce durable changes. Pure read-only explanations should usually avoid opening a KHS/KAS run, or mark review feedback `not_applicable` with an explicit reason.

Feedback requests for Codex/KAB-authored plans, implementation diffs, or review rounds must ask reviewers to audit fallback paths explicitly. The requested default is fail-closed behavior and removal of unnecessary fallbacks. Reviewers may accept a fallback only when no safe direct handling exists, it is narrowly bounded/evidenced, and the code/docs delta is genuinely small; broad fallback design or new policy/state machinery must be reported to 주군 as an option decision instead of accepted silently.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

GLM Octo is mandatory for `development` / implementation tasks unless 주군 explicitly waives Octo before start. Run official GLM Octo after the first Blue + Red/Orange/Gray color review and feedback handling, then treat it as a later feedback round: capture the Octo prompt/readback/feedback, triage and apply/reject findings, rerun affected verification, then require a second Blue + Red/Orange/Gray color re-review before final/pre-commit reporting. For non-implementation durable-change runs, run official GLM Octo only when 주군 explicitly requests it, a project-local workflow declares it required, or a recorded high-risk policy gate opts in. Hermes may request other optional continuation rounds 2..5 when earlier feedback exposes unresolved risk, broad changes, or unclear verification; never exceed five request-feedback/handle-feedback pairs.

## Outputs

- `feedback-request.md`
- `feedback-1.md`
- optional `feedback-2.md` through `feedback-5.md` as KHS supplemental artifacts when additional rounds run
- KAH phase/gate events, when supported

## KAB GLM Octo feedback lane notes (mandatory for official Octo)

Official GLM Octo feedback is valid only when it is run through `kkachi-agent-bridge` as a GLM backend session. Direct `glm` CLI invocation is allowed only for real-user-HOME path/auth/version preflight; a direct `glm` review transcript is not official Octo evidence and must fail closed.

When running official GLM `/octo:review` feedback:

1. Use this after the normal KAS/KAH path has reached docs/verification, first Blue/Red/Orange/Gray color review, and any required improvement pass, unless 주군 explicitly asks for earlier feedback. This is mandatory for `development` / implementation tasks unless 주군 explicitly waives Octo before start. For non-implementation durable-change runs, Octo remains requested/project-local/high-risk opt-in.
2. Use an explicit KAB config/evidence artifact for the run, including `backend_type: glm`, adapter identity, GLM model, resolved `glm_command` from the real user home/PATH, approved args/caveats, and the KAB session controller path. Do not copy one operator's home path into generic KAS source guidance; concrete run artifacts may record the actual resolved path such as `/Users/draccoon/.local/bin/glm`.
3. Before start/send, run KAB and GLM HOME/auth preflight from the real user home that owns the GLM token store, e.g. `HOME=<real-user-home> zsh -lc 'command -v kkachi-agent-bridge && command -v glm && glm --version'`. This preflight may prove availability only; it must not perform the review outside KAB.
4. Start/send through KAB only. Evidence must include KAB session id, backend type `glm`, bridge/tmux/session controller evidence, `selected-cli.json`, `capability-check.md`, `bridge-session-snapshot.json`, and `bridge-events.md` or equivalent KAH artifacts.
5. The actual prompt sent to the KAB GLM backend must begin with `/octo:review` as the first command text. Evidence must include the exact prompt artifact and TUI/readback snippet showing the command was submitted and activated.
6. If the prompt appears pasted but not submitted, send Enter to the bridge tmux/session and re-read status/TUI until `prompt_confirmed: true` or a bounded failure is recorded.
7. Use a bounded watcher after activation is confirmed; Octo may take up to 30 minutes. On completion, copy feedback into the KAH run directory, append the KAH feedback event, create/dispatch Blue/Red/Orange/Gray triage/re-review cards as required, then remove any watcher.
8. Parse verdicts only from the actual verdict heading/field, not examples or requested-output text.
9. If GLM/Octo feedback requires repository changes, do not patch them directly in Blue/Red/Orange/Gray review lanes. Record the finding and route the fix through the selected implementer lane via `handle-feedback`; after the implementer changes anything, rerun affected verification and the required post-change color review.
10. Clean generated local sidecars such as `.claude/`, `.claude-octopus/`, or GLM/Octo scratch state before final status unless they are intentionally in scope and ignored/recorded.
