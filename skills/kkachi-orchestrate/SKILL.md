---
name: kkachi-orchestrate
description: Coordinate an explicit KHS/Kkachi software run by selecting Path A or Path B, choosing phase order, invoking KHS prompt/process skills, using KAH for state, using KAB for backend delivery when needed, and reporting to the master in Korean.
version: 0.2.0
---

# Kkachi Orchestrate

Use this skill when the master asks Hermes Agent to run a Kkachi project task end to end, explicitly says to use KHS/Kkachi/KAB/KAH, applies KHS to a project, or assigns the development work to a KHS-using commander such as 조운 or 마초.

Do not use this skill for ordinary direct Hermes edits, small one-file fixes, typo/config tweaks, or read-only explanations unless the master explicitly requests KHS/Kkachi or delegates the work to a KHS-using commander.

## Activation boundary

Default mode is direct commander response: answer, inspect, or perform bounded non-durable work without creating a KHS task. KHS mode starts only when the master explicitly selects KHS/Kkachi/KAB/KAH, applies KHS to a project directory, requests durable repo artifact changes under a Kkachi-governed project, asks for phase/gate evidence, or the work needs backend execution, KAB plan lifecycle, bridge evidence, or long-lived team collaboration. Task classification is mandatory only after KHS mode is active; it must not turn every chat message into a KAS task.

See `references/kas-activation-scope.md` for the durable activation-scope lesson and investigation/spec/roadmap mapping. See `references/kas-installed-profile-vs-source-promotion.md` for installed-profile vs KAS-source promotion boundaries, optional `references/` semantics, and cleanup of project-specific source references.

For state investigation that leads to spec or roadmap work, use `research_evidence` for the read-only evidence stage and `docs_only + Path B shaping` for durable spec/SOT/roadmap/handoff edits. Escalate to `development` only when the request includes code, tests, build behavior, executable contracts, or future execution-policy changes. If only epic-level roadmap content exists, say so explicitly and describe any suggested task id/title as a proposed decomposition, not as an existing task.

## Core rule

Orchestration chooses phases and gates; it does not bypass phase contracts. KHS has two workflow layers: `.kkachi-workflow.yaml` is project-level graph state only after capability-checked `kkachi-agent-helper graph` evidence, while `.kkachi/runs/<run_id>/phase-plan.yaml` is run-local execution state/evidence. KAH is mandatory once KHS is triggered because KAH owns deterministic state, artifacts, events, locks, schemas, diagnostics, gates, and graph validation/apply mechanics, but KAH `work_path`, `work_mode`, and `execution_mode` are helper classification metadata rather than phase authority. The active v0.2 default path is KAS policy plus KAH deterministic evidence, approved GJC candidate artifacts, and KAT factual evidence when applicable. KAB-backed work is explicit-only: it requires current task approval, task contract, backend selection, capability check, rendered prompt, bridge evidence, and no reliance on historical Stage/direct-Codex/native_codex adoption markers.

Hermes is manager/orchestrator/final verifier, not the default code author. KAB backend roles perform substantive planning, implementation, docs, feedback, and feedback handling during KAB-backed KHS phases. Scoped KAS/KAH-local CLIMVP, GRAPHMVP, and docs-only maintenance may proceed without KAB when explicitly authorized and recorded, but must not claim backend execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence. Simple direct Hermes fixes are outside KHS unless the master explicitly keeps the task inside KHS.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

## Task classification first

Before selecting phases or creating backend prompts, classify the request using `registries/task-taxonomy.yaml` and record the result in `task-contract.yaml`:

- `development`: code, tests, build, architecture, process contract, or behavior-changing work. Use the full development spine.
- `research_evidence`: read-only investigation, evidence gathering, option comparison, log/source inspection, or current-state report. Do not run implementation/test/optimize phases unless the classification changes.
- `docs_only`: durable docs authoring/sync with no executable behavior or execution-policy change. Use docs validation, not code implementation phases.
- `simple_command_report`: bounded command/status check with no durable project change. Keep it outside KHS by default unless the master explicitly keeps it in KHS.
- `bootstrap_config`: repo/KAH project/profile/manifest/tooling/test-bed setup. Use preflight/configure/verification; require explicit approval for auth/secrets/gateway changes.
- `collaboration_review`: durable review/risk/team feedback routing. Use Kanban for long-lived team-member collaboration; temporary subagents are not team-member delegation.

Record the classification reason and every skipped phase reason. Do not silently apply the development spine to research, docs-only, simple-command, bootstrap, or review tasks.

## Output policy

KHS product output for generated prompts, executor reports, console schemas, and run artifacts is English for the active v0.2 KAS/KAH/GJC/KAT path and for any explicitly approved KAB-mediated lane. Commander chat reports to 주군 may remain Korean, but backend prompt/product output must preserve the compact schema for explicit KAB lanes, and executor prompt/product output must preserve the same compact schema for v0.2 GJC/KAH/KAT work: `Status`, `Summary`, `Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and `Next action requested`.

Detailed phase content is artifact-first. Use `.kkachi/runs/<run_id>/artifacts/<phase>/backend-<phase>.md` or the concrete requested phase artifact for plans, logs, diffs, reviews, findings, and file excerpts. If that artifact cannot be written, report a compact `Status: blocked` artifact-write blocker and do not dump the full detail into chat.

## Default phase spine

```text
intake -> sot -> roadmap -> task-classification -> plan -> ralplan -> implement -> enhance-test -> ai-slop-cleaner -> optimize -> docs -> verify -> review -> request-feedback-1 -> handle-feedback-1 -> mar-review -> second-color-review -> final; explicit approval/question evidence is recorded at real boundaries, not as a default ask phase, and plan-vet remains an acceptance/review gate term rather than a normal phase node
```

This full spine is the default/base graph only for no-input `development` tasks. It must not be forced onto custom project workflows or non-development task classes. Its graph phase ids are `docs`, `verify`, and `final`; older activity aliases such as `update_docs` and `final_verify` are translation-only compatibility names. Path B replaces production code implementation with shaping, SOT, roadmap, acceptance, and handoff artifacts until Path A gates pass.

## Operating policy

- The master selects the roadmap task id or task item for each KHS run; do not auto-pick the next task by default.
- All terminal commands for an active KAS/KAH run must execute with the real user home, not a Hermes role/profile home. Use `HOME=<real-user-home> ...` in reusable prompts/artifacts, including Git, tests, KAH/KAB/Hermes/Kanban commands, and Codex probes.
- KAB is required only when KAB-backed execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence is part of the contract. The default v0.2 KAS/KAH path uses KAS policy, KAH evidence, approved GJC candidate artifacts, and KAT factual evidence only. Do not substitute historical Stage/Codex app-server evidence for active KAB, implementation evidence, or capability-gated backend selection among eligible KAB backends.
- Scoped KAS/KAH-local CLIMVP, GRAPHMVP, and docs-only maintenance may proceed without KAB when that lane is explicitly authorized and recorded; do not claim KAB runtime support in those cases.
- `request-feedback-1`, `handle-feedback-1`, `mar-review`, and the `final` graph phase are mandatory in the default/base KHS development graph unless a recorded policy decision says otherwise; explicit approval/question evidence is a boundary, not a default `ask` phase. `final-verify` is only the skill/activity alias where older phase-contract wording still requires it. `ai-slop-cleaner` and `optimize` are conditional but strongly recommended for code-change runs, and skipping either requires a reason.
- Feedback runs at least once and at most five rounds. Round 1 is the normal first color review/feedback round. Rounds 2..5 are optional continuation rounds, and each requested feedback round must have a matching handle-feedback round.
- Classified KAS/KAH development runs must enter `workflow-trigger` through `--workflow-managed` with preserved `workflow-route` evidence before backend dispatch; selector, explicit, or custom-packet bypasses are not valid substitutes for the strict route-backed path.
- STRICT workflow-managed runs must follow route -> trigger -> ready -> start -> work -> required outputs -> complete. Treat KAH node start failure, stale `expected_start_revision`, missing required outputs, KAH node complete failure, missing `workflow_phase_projection_validation`, or phase-plan/checklist projection failure as blockers, not warnings; do not continue by skill order, phase-plan text, manual direct KAH state writes, fallback workflow, fallback backend, or fallback agent.
- MAR review is required for `development` / implementation tasks after first color review and feedback handling, unless the master explicitly waives or replaces MAR before start and the decision is recorded in KAH/run evidence artifacts. MAR is a role-first independent review lane for `logic`, `security`, `arch`, `cve`, and `test_adequacy`; unresolved required role coverage fails closed. For non-implementation durable-change runs, run MAR when the master explicitly requests independent review, a project-local approved workflow declares it required, or a recorded high-risk policy gate opts in.

See `references/run-operating-policy.md` for the active v0.2 lane ownership rules, required plan-vet loop, fallback-audit expectations, explicit-only KAB boundary, and MAR review details.

## Implementation approval policy

After plan/`ralplan` candidate evidence and required plan-vet acceptance are complete, Hermes may start low-risk implementation automatically only when policy and approval-boundary evidence allow it. Require explicit master approval before implementation when the run touches API, DB/schema/migration, security/auth/secrets, dependencies, architecture, SOT, large diff/broad fanout, low confidence, or unresolved approval-boundary ambiguity.

## Required responsibilities

- classify task class before Path A/Path B
- classify Path A or Path B and select Standard or Light mode
- invoke `kkachi-task-contract`
- run graph capability preflight before using `.kkachi-workflow.yaml`
- create and maintain `.kkachi/runs/<run_id>/phase-plan.yaml`
- invoke `kkachi-backend-select` and `kkachi-prompt-compose` only when the selected lane requires them
- use `kkachi-phase-state` and KAH at every phase boundary
- preserve Korean commander reports and English run artifacts

See `references/orchestration-responsibilities.md` for the exact preflight sequence, fail-closed graph boundary, and per-lane backend invocation details.

## Output

- phase plan
- current run state summary
- final Korean report from artifacts
