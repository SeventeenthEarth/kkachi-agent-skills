---
name: kkachi-plan
description: Produce or validate a Kkachi plan from SOT basis, roadmap trace, task contract, acceptance criteria, constraints, non-goals, backend lane requirements, and verification strategy.
version: 0.2.0
---

# Kkachi Plan

Use this skill during the `plan` phase of Path A or Path B.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Plan from durable authority, not chat-only instruction. Path A plans prepare implementation; Path B plans prepare shaping and handoff.

## 주군 development-pipeline preference

For KAS/KAH/KAT repository self-development, do not dogfood this KAS/KAH/GJC pipeline by default. 황충 should do main development directly, then route the result through official color review and fixes/re-review. Use the KAS/KAH/GJC pipeline for these repos only when 주군 explicitly selects it for the run.

When 주군 asks to run KAS/Kkachi development work and the task is classified as
`development`, treat the preferred v0.2 default as a GJC delegated plan-first
loop under KAS/KAH control:

- **deep-interview** is explicit-only epic/design clarification, not a normal
  task step.
- **ralplan** produces candidate plan artifacts for KAS/Blue/color review.
  It does not authorize implementation, and `ask` is not a default normal phase; only explicit approval/question evidence gates implementation.
- **ultragoal** produces implementation-candidate artifacts only after
  explicit bounded implementation approval; it is not review, MAR, final acceptance, commit, install,
  or runtime activation.

V02FLOW-013 planning must preserve the implementation completion status hierarchy: `implementation_goal_bundle_ready is goal-bundle-only and never sufficient for implementation completion`; `implementation_diff_ready` requires executor-loop source diff/checkpoint/checksum evidence; and `implementation_verified` requires passed verification output after that executor loop. Plans for `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, and `docs-update` must name required evidence fields `changed_source_refs`, `diff_refs`, `checkpoint_ref`, `checkpoint_status`, `verification_output_refs`, `checksums`, `termination_reason`, `HOME`, and `no_authority_boundaries`.

KAH records deterministic run/gate/evidence state. KAT is factual/mechanical
evidence only. KAB runtime/session control is out of the default v0.2 path and
requires explicit KAB selection with current capability and bridge evidence.
Legacy Stage 1/Stage 2/Stage 3 Codex/KAB adoption wording is historical/stale
and must not be used as active guidance.

MAR review is independent of the candidate planning/implementation lane. It
remains the default KAS review lane when applicable and does not change the
plan/implementation authority.

Plan drafts for active KAS/KAH roadmap policy, workflow, template, test, or
shared skill mirror work must include a fallback audit note before Blue plus
Red/Orange/Gray plan review. Resolve reviewers from the project/team role
registry when applicable. Ask the planner to identify any fallback paths it
proposes, remove unnecessary fallback behavior, and prefer fail-closed handling
when capability, evidence, approval, or safe state is missing. Allow a fallback
only when no safe direct handling exists, the fallback is tightly
bounded/evidenced, and the required code/docs delta is genuinely small. If the
only viable fallback would add broad code, new state machinery, or unclear
policy, stop and report options to 주군 instead of letting the planner quietly
add it.

## TWAKE-003 return-path policy

Plans for Blue-dispatched async work such as plan-vet, color review, MAR, second-color review, GJC long-running dispatches, and blocked-condition probes must include return-path evidence requirements or an explicit no-wake/blocked state. Require effective KAH capability readback for `async_dispatch_return_path_evidence=true` and `async_dispatch_return_path_final_gate=true` before planning any same-thread wake readiness claim. If the effective binary lacks those capabilities, or watcher/subscription/callback/origin evidence is unavailable, the plan must preserve `blocked/degraded/no_wake_claim` with an operator-readable recovery hint. Required watcher reports must be terminal-only Blue-action-required output. watcher/notifier output is state-report-only and never review, MAR, waiver, Blue synthesis, or final acceptance authority.

For DESIGN Teal work, plans must carry `project_has_teal_lane`,
`ui_ux_change`, derived `teal_required`, and `teal_skip_reason` for false
inputs. `DESIGN_PLAN_GATE` is required before implementation authorization
when Teal is required, and `DESIGN_FIDELITY_REVIEW` is required before final
acceptance. A waiver must identify approval ref, scope, and expiry. Ordinary
Red/Orange/Gray/Blue color review, MAR, backend implementation evidence, and
temporary helpers must not substitute for a required Teal verdict.

Before backend planning for a code-changing or process-changing task, refresh CodeGraph evidence for the target repository. If `.codegraph/` already exists, run `codegraph index <repo>` and preserve `codegraph status <repo>` output. If CodeGraph is due for first initialization after the first completed task and `.codegraph/` is missing, run `codegraph init -i <repo>` and preserve status evidence. If CodeGraph is unavailable, record the missing capability as a blocker or degraded-evidence reason instead of silently planning from stale code context.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

Render planner prompts so every command example uses the real user home, for example `HOME=<real-user-home> <command>` in reusable artifacts. This includes Git commands because commit-time global `user.name`, `user.email`, signing, hooks, and credential helpers must come from the user's real home.

Record deviations in `phase-plan.yaml`, `checklist.md`, and the final report instead of silently using a lighter path. For `research_evidence`, `docs_only`, `simple_command_report`, `bootstrap_config`, or `collaboration_review`, use the selected light spine from `task-contract.yaml`; do not manufacture implementation/test/optimize phases unless the classification changes to `development`.

Planner product output for v0.2 GJC candidate artifacts or any explicitly selected KAB lane must be English and compact: `Status`, `Summary`, `Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and `Next action requested`. Write detailed planning content to `plan.md` and, when backend-authored phase detail is needed, `.kkachi/runs/<run_id>/artifacts/plan/backend-plan.md` or the equivalent requested phase artifact. If the artifact cannot be written, report `Status: blocked` with the artifact-write blocker and do not dump full plans into chat.

Plan-stage reports to 주군 must also follow `docs/sot/stage-report-contract.md`: summarize how the documented plan will be realized, whether the original plan was followed or changed, what Red/Orange/Gray/Blue plan-vet feedback changed, which verification strategy will prove the implementation, and whether implementation remains held for approval. Do not report only `plan stage done` or the list of plan artifact files.

See `references/planner-lane-and-capture.md` for retained historical details; for current v0.2 KAS/KAH tasks, use the GJC `ralplan` candidate path plus fallback-audit capture and KAH evidence rules. See `references/checklist-normalization.md` for the mandatory `checklist.md` transform rules.

## Inputs

- `task-contract.yaml`
- SOT basis and roadmap trace
- project overlay
- related code/docs evidence
- CodeGraph status evidence or explicit unavailable/degraded reason when required

## Outputs

- `plan.md`
- `checklist.md`
- `graph-evidence.md` mapping requirement when the plan initializes, validates, explains, diffs, proposes, applies, or otherwise relies on graph state
- KAH phase/gate events, when supported

## Gate

PASS when scope, non-scope, affected surfaces, backend capability needs, verification strategy, docs impact, rollback/blocker conditions, the mandatory progress checklist, and Blue+Red+Orange plan vet verdicts are explicit when required by KAS/KAH roadmap policy. The KAH plan gate requires `acceptance-criteria.md`, `plan.md`, and `checklist.md`; KAS/KAH roadmap code-development must also preserve required plan-vet approval or a recorded 주군-approved exception in KAH/run evidence artifacts before implementation.
