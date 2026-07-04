---
name: kkachi-ask
description: Capture unresolved decisions or blockers in a Kkachi run and convert answers into updated task, plan, or shaping artifacts before implementation proceeds.
version: 0.2.0
---

# Kkachi Ask

Use this skill when a phase is blocked by missing master decision, unclear acceptance criteria, missing authority, or materially branching scope.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

Ask always runs after `plan` in KHS. It is not optional just because no blocker is obvious. Ask only for decisions Hermes cannot derive safely from SOT, project docs, or Path B shaping; when no open question remains, record `no unresolved decisions` instead of skipping the phase.

## Flow

1. Ask the planner backend: "Were there ambiguous parts in this design, or anything requiring the user's decision?"
2. Hermes answers directly when the answer is safely derivable from SOT, project docs, task contract, or prior master decision.
3. Hermes asks the master only when intent, risk, or policy cannot be resolved safely.
4. Update `plan.md`, `phase-plan.yaml`, and `checklist.md` when an answer changes scope, approval risk, phase applicability, or evidence.

## Outputs

- `ask.md`
- `answered-decisions.md`
- updated task contract or plan when the answer changes the work
- KAH `event append question.asked` / `question.answered` records, when useful
- rerun the applicable `gate check` after the answer updates required artifacts
