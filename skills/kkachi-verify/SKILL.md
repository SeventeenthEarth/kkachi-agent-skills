---
name: kkachi-verify
description: Run or record verification for a Kkachi task, including exact commands, results, failure classification, integration coherence checks, and final test verdict.
version: 0.2.0
---

# Kkachi Verify

Use this skill when tests, static checks, manual QA, or integration coherence evidence must be recorded.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Verify before claiming completion. If verification fails, classify the failure and continue recovery when safe. Verification must update `phase-plan.yaml` and `checklist.md` with the current phase state and evidence links before final verification.

For KAB-backed work, do not treat `send` success as completion. Completion evidence must include one of:

- CLI evidence: `wait` returned a real state or pending change, followed by `read` or `status` showing the expected assistant result, terminal state, or actionable pending state.
- Stream evidence: retained `/api/stream/<session_id>` or `/api/events/<session_id>` observations show the relevant session/turn/pending/completion events, followed by `read` or `status` confirming the current public session state.

If stream evidence is used, record that retained events are bridge-owned public events and may not be durable across daemon restart.

Verification product output must be English and compact for v0.2 GJC candidate artifacts and any explicitly selected KAB lane: `Status`, `Summary`, `Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and `Next action requested`. Do not use historical Stage/Codex app-server evidence as an active verification path unless a later approved task explicitly selects and evidences that lane. Detailed verification logs and findings belong in `verification.md`, `test-log.md`, or `.kkachi/runs/<run_id>/artifacts/verify/backend-verify.md`. For active v0.2 KAS/KAH development, `ralplan_candidate_recorded` is the primary ralplan candidate status and `implementation_goal_bundle_ready` is the primary ultragoal goal-bundle status; legacy/current-helper aliases `ralplan_ready` and `ultragoal_ready` must not be used as consensus, acceptance, implementation diff readiness, or approval evidence. If the detailed artifact cannot be written, report `Status: blocked` with the artifact-write blocker; do not paste full logs, diffs, files, reviews, or exhaustive checklists into chat.

V02FLOW-009 verification for implementer phases must bind evidence to the selected GJC `ultragoal` executor lane or an explicit recorded exception. For `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, `docs-update`, and `handle-feedback-*`, fail closed on missing KAH V02FLOW-010 capability/readback, stale/absent post-change verification, missing accepted-scope/ralplan/hash/approval refs, unsafe refs/checksums, KAB dispatch success as completion evidence, Blue/color source-patch fallback without exception evidence, or native GJC `ai-slop-cleaner` or `remove-ai-slop` requests.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

Backend-specific verification requirements:

- Explicitly selected KAB Codex stream evidence must use Kkachi-stable wrapper-derived public event kinds, not raw Codex app-server messages.
- Gemini plan-mode verification must prove explicit approved-plan start when implementation followed plan approval.
- OpenCode readiness must not count rendered `<tool_call>question` assistant text as question-flow evidence; only real upstream API/SSE question events count.
- GLM verification after reject must inspect `response_fidelity_warning` before trusting the assistant text.

## Outputs

- `verification.md`
- `test-log.md`
- updated `phase-plan.yaml` and `checklist.md` rows
- integration coherence notes
- KAH gate event, when supported


## V01CLEAN active-baseline note

Any legacy Stage 1/Stage 2/Stage 3, direct Codex app-server, or KAB `native_codex` wording retained in this file is historical context only unless a later approved task explicitly selects KAB with current capability evidence. The active KAS/KAH v0.2 path is KAS policy + KAH deterministic evidence + approved GJC candidate artifacts, with KAT factual evidence only.
