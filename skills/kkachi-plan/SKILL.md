---
name: kkachi-plan
description: Produce or validate a Kkachi plan from SOT basis, roadmap trace, task contract, acceptance criteria, constraints, non-goals, backend lane requirements, and verification strategy.
version: 0.1.0
---

# Kkachi Plan

Use this skill during the `plan` phase of Path A or Path B.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Plan from durable authority, not chat-only instruction. Path A plans prepare implementation; Path B plans prepare shaping and handoff.

## 주군 development-pipeline preference

When 주군 asks to run KAS/Kkachi development work and the task is classified as `development`, treat the preferred default as a Codex-led plan-first loop, with the transport determined by the current KAS KAB adoption stage:

- **Stage 1:** direct Codex SDK/app-server planner through `templates/runners/direct-codex-sdk-appserver-runner.py.tmpl`; the runner imports `openai_codex` and lets the SDK start `codex app-server --listen stdio://`, with direct Codex evidence and no-KAB-Codex rationale. Do not use `codex exec` or generic `openai` SDK evidence. Do not pass an explicit `--model` by default; use the configured Codex account default model. For plan-only turns, pass reasoning effort `high` and record any lower-effort exception in the run artifacts. A KAS/KAH task should use one Codex thread across its plan/revision/implementation/feedback turns by resuming the recorded `thread_id`; start a new thread for the next task. This thread continuity is task-bound and is not Discord-session-bound.
- **Stage 2:** KAB Codex-first planner through `native_codex`; this replaces direct Codex app-server calls without changing the KAS/KAH phase or review scenario.
- **Stage 3:** KAB backend-selected planner after backend selection chooses an eligible backend from task requirements, project policy, capability evidence, and user preference after gates.

MAR review is independent of these planning stages. It remains the default KAS review lane and does not change the plan/implementation backend.

Plan drafts for active KAS/KAH roadmap policy, workflow, template, test, or shared skill mirror work must include a fallback audit note before Blue+Red+Orange plan review. Resolve Red/Orange plan-vet reviewers from the project/team role registry when applicable. Ask the planner to identify any fallback paths it proposes, remove unnecessary fallback behavior, and prefer fail-closed handling when capability, evidence, approval, or safe state is missing. Allow a fallback only when no safe direct handling exists, the fallback is tightly bounded/evidenced, and the required code/docs delta is genuinely small. If the only viable fallback would add broad code, new state machinery, or unclear policy, stop and report options to 주군 instead of letting the planner quietly add it.

Before backend planning for a code-changing or process-changing task, refresh CodeGraph evidence for the target repository. If `.codegraph/` already exists, run `codegraph index <repo>` and preserve `codegraph status <repo>` output. If CodeGraph is due for first initialization after the first completed task and `.codegraph/` is missing, run `codegraph init -i <repo>` and preserve status evidence. If CodeGraph is unavailable, record the missing capability as a blocker or degraded-evidence reason instead of silently planning from stale code context.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

Render planner prompts so every command example uses the real user home, for example `HOME=<real-user-home> <command>` in reusable artifacts. This includes Git commands because commit-time global `user.name`, `user.email`, signing, hooks, and credential helpers must come from the user's real home.

Record deviations in `phase-plan.yaml`, `checklist.md`, and the final report instead of silently using a lighter path. For `research_evidence`, `docs_only`, `simple_command_report`, `bootstrap_config`, or `collaboration_review`, use the selected light spine from `task-contract.yaml`; do not manufacture implementation/test/optimize phases unless the classification changes to `development`.

Planner product output for both Stage 1 direct Codex SDK/app-server runner and KAB-mediated lanes must be English and compact: `Status`, `Summary`, `Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and `Next action requested`. Write detailed planning content to `plan.md` and, when backend-authored phase detail is needed, `.kkachi/runs/<run_id>/artifacts/plan/backend-plan.md` or the equivalent requested phase artifact. If the artifact cannot be written, report `Status: blocked` with the artifact-write blocker and do not dump full plans into chat.

See `references/planner-lane-and-capture.md` for the full Stage 1/2/3 plan-first loop, fallback-audit capture prompt, KAB plan capture rule, and backend timing details. See `references/checklist-normalization.md` for the mandatory `checklist.md` transform rules.

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
