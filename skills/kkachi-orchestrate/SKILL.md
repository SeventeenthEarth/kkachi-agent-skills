---
name: kkachi-orchestrate
description: Coordinate an explicit KHS/Kkachi software run by selecting Path A or Path B, choosing phase order, invoking KHS prompt/process skills, using KAH for state, using KAB for backend delivery when needed, and reporting to the master in Korean.
version: 0.1.0
---

# Kkachi Orchestrate

Use this skill when the master asks Hermes Agent to run a Kkachi project task end to end, explicitly says to use KHS/Kkachi/KAB/KAH, applies KHS to a project, or assigns the development work to a KHS-using commander such as 조운 or 마초.

Do not use this skill for ordinary direct Hermes edits, small one-file fixes, typo/config tweaks, or read-only explanations unless the master explicitly requests KHS/Kkachi or delegates the work to a KHS-using commander.

## Activation boundary

Default mode is direct commander response: answer, inspect, or perform bounded non-durable work without creating a KHS task. KHS mode starts only when the master explicitly selects KHS/Kkachi/KAB/KAH, applies KHS to a project directory, requests durable repo artifact changes under a Kkachi-governed project, asks for phase/gate evidence, or the work needs backend execution, KAB plan lifecycle, bridge evidence, or long-lived team collaboration. Task classification is mandatory only after KHS mode is active; it must not turn every chat message into a KAS task.

See `references/kas-activation-scope.md` for the durable activation-scope lesson and investigation/spec/roadmap mapping. See `references/kan-plugin-readiness-and-activation.md` for project-readiness checks, first-run criteria, git hygiene, and profile-sync pitfalls observed while preparing kan-plugin.

For state investigation that leads to spec or roadmap work, use `research_evidence` for the read-only evidence stage and `docs_only + Path B shaping` for durable spec/SOT/roadmap/handoff edits. Escalate to `development` only when the request includes code, tests, build behavior, executable contracts, or future execution-policy changes.

## Core rule

Orchestration chooses phases and gates; it does not bypass phase contracts. KHS has two workflow layers: `.kkachi-workflow.yaml` is project-level graph state only after capability-checked `kkachi-agent-helper graph` evidence, while `.kkachi/runs/<run_id>/phase-plan.yaml` is run-local execution state/evidence. KAH is mandatory once KHS is triggered because KAH owns deterministic state, artifacts, events, locks, schemas, diagnostics, gates, and graph validation/apply mechanics, but KAH `work_path`, `work_mode`, and `execution_mode` are helper classification metadata rather than phase authority. KAB-backed work additionally requires task contract, backend selection, capability check, rendered prompt, and bridge evidence. Direct Codex app-server pilot work is a no-KAB lane until KAB runtime implementation is ready; record it as direct Codex app-server evidence, not bridge evidence.

Hermes is manager/orchestrator/final verifier, not the default code author. KAB backend roles perform substantive planning, implementation, docs, feedback, and feedback handling during KAB-backed KHS phases. Scoped KAS/KAH-local CLIMVP, GRAPHMVP, and docs-only maintenance may proceed without KAB when explicitly authorized and recorded, but must not claim backend execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence. Simple direct Hermes fixes are outside KHS unless the master explicitly keeps the task inside KHS.

## Task classification first

Before selecting phases or creating backend prompts, classify the request using `registries/task-taxonomy.yaml` and record the result in `task-contract.yaml`:

- `development`: code, tests, build, architecture, process contract, or behavior-changing work. Use the full development spine.
- `research_evidence`: read-only investigation, evidence gathering, option comparison, log/source inspection, or current-state report. Do not run implementation/test/optimize phases unless the classification changes.
- `docs_only`: durable docs authoring/sync with no executable behavior or execution-policy change. Use docs validation, not code implementation phases.
- `simple_command_report`: bounded command/status check with no durable project change. Keep it outside KHS by default unless the master explicitly keeps it in KHS.
- `bootstrap_config`: repo/KAH project/profile/manifest/tooling/test-bed setup. Use preflight/configure/verification; require explicit approval for auth/secrets/gateway changes.
- `collaboration_review`: durable review/risk/team feedback routing. Use Kanban for long-lived team-member collaboration; temporary subagents are not team-member delegation.

Record the classification reason and every skipped phase reason. Do not silently apply the development spine to research, docs-only, simple-command, bootstrap, or review tasks.

## Default phase spine

```text
plan -> ask -> implement -> enhance-test -> optimize -> docs-update -> request-feedback -> handle-feedback -> final-verify -> improve
```

This full spine is the default only for `development` tasks. Path B replaces production code implementation with shaping, SOT, roadmap, acceptance, and handoff artifacts until Path A gates pass.

## User-confirmed operating policy

- The master selects the roadmap task id or task item for each KHS run; do not auto-pick the next task by default.
- KAB is required only when KAB-backed execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence is part of the contract. Current kan-plugin P1 uses direct Codex app-server through Hermes/KAS/KAH, not KAB.
- Scoped KAS/KAH-local CLIMVP, GRAPHMVP, and docs-only maintenance may proceed without KAB when the master or roadmap authorizes that lane; record the no-KAB rationale and do not claim KAB runtime support.
- Logical backend roles are planner (`plan`, `ask`), implementer (`implement`, `enhance-test`, `optimize`, `docs-update`, `handle-feedback`), and feedback (`request-feedback`). They may map to the same or different physical backends.
- `ask`, `request-feedback-1`, `handle-feedback-1`, and `final-verify` are mandatory for every KHS run.
- `optimize` is conditional but strongly recommended for code-change runs to remove AI slop, duplication, and small structural waste; skipping requires a reason.
- Feedback runs at least once and at most five rounds. Rounds 2..5 are optional continuation rounds, and each requested feedback round must have a matching handle-feedback round.

## Implementation approval policy

After plan and ask are complete, Hermes may start low-risk implementation automatically after notifying the master of the intended direction. Require explicit master approval before implementation when the run touches API, DB/schema/migration, security/auth/secrets, dependencies, architecture, SOT, large diff/broad fanout, low confidence, or unresolved ask-phase ambiguity.

## Required responsibilities

- classify task class before Path A/Path B (`development`, `research_evidence`, `docs_only`, `simple_command_report`, `bootstrap_config`, or `collaboration_review`)
- classify Path A or Path B
- select Standard or Light mode from the task class
- invoke `kkachi-task-contract`
- run the graph capability preflight before using `.kkachi-workflow.yaml`: same effective binary, `kkachi-agent-helper --version`, `capabilities --json`, `graph --help`, then `graph validate/explain` for existing graph state or `graph init --from-template` only when graph creation is intended
- create `phase-plan.yaml` from `templates/run-artifacts/phase-plan.yaml.tmpl` and keep it as run-local execution state/evidence
- fail closed into a gap record when graph capability/help evidence is missing; do not use `kah graph` alias text or direct `.kkachi-workflow.yaml` edit fallback
- invoke `kkachi-backend-select` only when a KAB-backed lane is selected or bridge evidence is claimed; for the current direct Codex app-server pilot, record direct Codex session evidence instead
- invoke `kkachi-prompt-compose` before KAB delivery
- use `kkachi-phase-state` and KAH at every phase boundary
- preserve user-facing Korean reports and English run artifacts

## Output

- phase plan
- current run state summary
- final Korean report from artifacts
