---
name: kkachi-task-contract
description: Build an AI-neutral Kkachi task contract from the master's request, project overlay, Path A/B classification, phase contract, SOT basis, constraints, non-goals, required capabilities, and verification evidence.
version: 0.1.0
---

# Kkachi Task Contract

Use this skill when Hermes Agent needs to turn a master request into `task-contract.yaml` before selecting an AI backend or rendering a KAB prompt.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

The task contract is backend-neutral. Do not include Claude, Codex, Gemini, GLM, or OpenCode prompt style in acceptance criteria, constraints, or non-goals.

## Inputs

- master request
- project overlay and docs map
- `registries/task-taxonomy.yaml`
- `registries/phase-contracts.yaml`
- SOT basis or Path B shaping output
- roadmap trace or not-applicable reason
- acceptance criteria source
- user backend preference, if provided

## Flow

1. Confirm the master-selected roadmap task id or task item. Do not auto-select the next roadmap task by default.
2. Classify task class first (`development`, `research_evidence`, `docs_only`, `simple_command_report`, `bootstrap_config`, or `collaboration_review`), then task type, work path, phase spine, mode, urgency, SOT policy, and execution mode as KAH helper metadata.
3. Record the classification reason, selected spine, KAB/KAH/CodeGraph defaults, and explicit skip reasons for development-only phases when the task is not a development task.
4. Record desired state, acceptance criteria, constraints, non-goals, context sources, required capabilities, graph capability requirements when graph-managed workflow is selected, and verification evidence.
4. Record graph workflow authority explicitly: `.kkachi-workflow.yaml` is project-level graph state only after capability-checked KAH graph evidence, `phase-plan.yaml` is run-local execution state/evidence, and KAH metadata is helper classification only.
6. Preserve user backend preference only as selection metadata; do not let it bypass capability gates.
7. Render from `templates/run-artifacts/task-contract.yaml.tmpl` as a KHS supplemental run artifact.
8. Also summarize the same contract in KAH canonical `task-brief.md` or `context-pack.md` so current KAH artifact gates can see the run authority.
9. Record a compact event with `kkachi-agent-helper event append artifact.updated --run <run_id> --payload '{"path":"task-contract.yaml","phase":"task-contract"}' --json` when useful.

## Output

- `task-contract.yaml`
- `phase-plan.yaml` reference and selected task fields
- canonical task summary in `task-brief.md` or `context-pack.md`
- optional KAH `event append artifact.updated` evidence

## Gate

PASS only when the contract is explicit enough for backend selection and prompt composition, and when the task class justifies the selected phase spine. BLOCKED when desired state, authority, acceptance criteria, classification, or non-goals require a master decision.
