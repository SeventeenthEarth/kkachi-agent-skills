---
name: kkachi-orchestrate
description: Coordinate an explicit KHS/Kkachi software run by selecting Path A or Path B, choosing phase order, invoking KHS prompt/process skills, using KAH for state, using KAB for backend delivery when needed, and reporting to the master in Korean.
version: 0.1.0
---

# Kkachi Orchestrate

Use this skill when the master asks Hermes Agent to run a Kkachi project task end to end, explicitly says to use KHS/Kkachi/KAB/KAH, applies KHS to a project, or assigns the development work to a KHS-using commander such as 조운 or 마초.

Do not use this skill for ordinary direct Hermes edits, small one-file fixes, typo/config tweaks, or read-only explanations unless the master explicitly requests KHS/Kkachi or delegates the work to a KHS-using commander.

## Core rule

Orchestration chooses phases and gates; it does not bypass phase contracts. KHS has two workflow layers: `.kkachi-workflow.yaml` is project-level graph state only after capability-checked `kkachi-agent-helper graph` evidence, while `.kkachi/runs/<run_id>/phase-plan.yaml` is run-local execution state/evidence. KAH is mandatory once KHS is triggered because KAH owns deterministic state, artifacts, events, locks, schemas, diagnostics, gates, and graph validation/apply mechanics, but KAH `work_path`, `work_mode`, and `execution_mode` are helper classification metadata rather than phase authority. KAB-backed work additionally requires task contract, backend selection, capability check, rendered prompt, and bridge evidence.

Hermes is manager/orchestrator/final verifier, not the default code author. KAB backend roles perform substantive planning, implementation, docs, feedback, and feedback handling during KHS runs. Simple direct Hermes fixes are outside KHS unless the master explicitly keeps the task inside KHS.

## Default phase spine

```text
plan -> ask -> implement -> enhance-test -> optimize -> docs-update -> request-feedback -> handle-feedback -> final-verify -> improve
```

Path B replaces production code implementation with shaping, SOT, roadmap, acceptance, and handoff artifacts until Path A gates pass.

## User-confirmed operating policy

- The master selects the roadmap task id or task item for each KHS run; do not auto-pick the next task by default.
- KHS code-change and development runs use KAB as the rule. If the master asks for code changes without KAB, convert the work to a normal direct Hermes task instead of a KHS run.
- Docs-only KHS runs also use KAB by default. If the master explicitly forbids KAB for docs-only work, Hermes may edit directly and must record the rationale.
- Logical backend roles are planner (`plan`, `ask`), implementer (`implement`, `enhance-test`, `optimize`, `docs-update`, `handle-feedback`), and feedback (`request-feedback`). They may map to the same or different physical backends.
- `ask`, `request-feedback-1`, `handle-feedback-1`, and `final-verify` are mandatory for every KHS run.
- `optimize` is conditional but strongly recommended for code-change runs to remove AI slop, duplication, and small structural waste; skipping requires a reason.
- Feedback runs at least once and at most five rounds. Rounds 2..5 are optional continuation rounds, and each requested feedback round must have a matching handle-feedback round.

## Implementation approval policy

After plan and ask are complete, Hermes may start low-risk implementation automatically after notifying the master of the intended direction. Require explicit master approval before implementation when the run touches API, DB/schema/migration, security/auth/secrets, dependencies, architecture, SOT, large diff/broad fanout, low confidence, or unresolved ask-phase ambiguity.

## Required responsibilities

- classify Path A or Path B
- select Standard or Light mode
- invoke `kkachi-task-contract`
- run the graph capability preflight before using `.kkachi-workflow.yaml`: same effective binary, `kkachi-agent-helper --version`, `capabilities --json`, `graph --help`, then `graph validate/explain` for existing graph state or `graph init --from-template` only when graph creation is intended
- create `phase-plan.yaml` from `templates/run-artifacts/phase-plan.yaml.tmpl` and keep it as run-local execution state/evidence
- fail closed into a gap record when graph capability/help evidence is missing; do not use `kah graph` alias text or direct `.kkachi-workflow.yaml` edit fallback
- invoke `kkachi-backend-select` when a KAB lane is needed
- invoke `kkachi-prompt-compose` before KAB delivery
- use `kkachi-phase-state` and KAH at every phase boundary
- preserve user-facing Korean reports and English run artifacts

## Output

- phase plan
- current run state summary
- final Korean report from artifacts
