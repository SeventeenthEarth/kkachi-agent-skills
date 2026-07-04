---
name: kkachi-docs-update
description: Check and update durable project documentation after a Kkachi task changes behavior, tests, architecture, backend support, or process expectations.
version: 0.2.0
---

# Kkachi Docs Update

Use this skill during the docs-update phase.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Git commits, chat memory, and implementation notes are not durable project docs. Update canonical docs or record why no docs update is needed. Docs-only KHS runs use KAS/KAH deterministic evidence by default; GJC candidate docs work may be used only after the applicable approval gate, and KAB is used only when explicitly selected with bridge evidence. Record no-backend/no-KAB rationale in `phase-plan.yaml`, `docs-update.md`, and the final report. For active KAS/KAH roadmap-task work, task-bound docs updates are routed through the selected v0.2 lane: approved GJC `ultragoal` candidate output, explicitly selected KAB backend, or a recorded 주군-approved direct role-editing exception. Blue/Red/Orange/Gray may edit docs directly only when 주군 explicitly requests direct role editing or the work is outside the roadmap/KAS+KAH path, and the exception must be recorded. Legacy Stage/Codex app-server wording is historical/stale and must not be used as active guidance.

When a roadmap task is completed, update the roadmap only after implementation scope, verification, KAH evidence, and required review gates support completion. If completion confidence is insufficient, record `Blocked` or `In Progress` with the missing evidence instead of marking `Completed`. After docs changes that affect generated outputs, examples, or tested contracts, run the selected verification profile/gate command or the repository's docs validation target.

For active KAS/KAH policy, workflow, template, test, or shared skill mirror docs work, preserve a docs impact map and require project-Gray documentation/integrity review through the project role/registry. The review checks authority ladder, source trace, stale terminology, and cross-doc consistency; do not hard-code the role to an individual.

## Outputs

- `docs-update.md`
- `roadmap-update.md` when roadmap status changes
- updated docs diff, when needed
- KAH phase/gate events, when supported
