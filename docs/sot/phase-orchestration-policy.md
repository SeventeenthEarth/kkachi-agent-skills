# KHS Phase Orchestration Policy

Date: 2026-05-21
Owner: KHS workflow/policy layer
Confirming role: Responsible approver / governance evidence record
Status: user-confirmed operating policy; workflow graph additions are planning-confirmed and `kkachi-agent-helper graph` implementation evidence is present
Authority level: KHS phase/run orchestration SOT
Evidence/source path: governance evidence record in kanban task `t_2fb00394`

This document records the current user-confirmed KHS behavior for software-development runs. The executable contract lives in `registries/phase-contracts.yaml`; this file explains the decisions for humans.

## Authority

- `phase-plan.yaml` is the run-local KHS workflow/execution state source of truth for one run. It is not the project-level workflow graph and is not deprecated by capability-checked graph support.
- Project graph: `.kkachi-workflow.yaml` is the project-level workflow graph instance when initialized, validated, or applied through capability-checked `kkachi-agent-helper graph`; see `workflow-graph-integration.md`. The `kah graph` alias remains unproven unless separately advertised.
- KAH `work_path`, `work_mode`, `execution_mode`, `urgency`, and `sot_policy` are deterministic helper metadata. They may seed defaults, but they do not decide which KHS phases execute.
- KAH owns run IDs, artifacts, events, locks, schemas, diagnostics, and gates.
- KAB owns backend runtime sessions and public bridge evidence.
- If project graph, KHS phase policy, and run-local `phase-plan.yaml` conflict, KHS must fail closed and seek responsible role confirmation instead of guessing.

## Task selection

- The master selects the target roadmap task or task item for each KHS run.
- KHS must not auto-pick the next roadmap item by default.
- If project docs differ by repository, KHS reads the project overlay/docs map first.

## Hermes and backend roles

- Hermes is manager, risk approval router, final verifier, and Korean reporter.
- Hermes does not normally author production code inside a KHS run.
- KHS defines three logical backend roles:
  - planner: `plan`, `ask`
  - implementer: `implement`, `enhance-test`, `optimize`, `docs-update`, `handle-feedback`
  - feedback: `request-feedback`
- These logical roles may map to one physical KAB backend/session or to different backends according to project policy, user preference, Hermes judgment, and capability evidence.

## KAB usage

- KAB is required only when KAB-backed execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence is part of the contract. Stage 1 direct Codex app-server baseline work remains a no-KAB-Codex lane until Stage 2 KAB Codex-first execution is explicitly selected and evidenced.
- KAB-backed code-change and development runs must preserve selected backend identity, runtime evidence, and bridge completion evidence.
- Scoped KAS/KAH-local CLIMVP, GRAPHMVP, and docs-only maintenance may proceed without KAB when the master or roadmap authorizes that lane.
- No-KAB KAS/KAH-local work must record the rationale in `phase-plan.yaml`, `docs-update.md`, and the final report, and must not claim KAB runtime support.

## Canonical phase spine

```text
plan -> ask -> implement -> enhance-test -> optimize -> docs-update -> request-feedback -> handle-feedback -> final-verify -> improve
```

Mandatory for every KHS run:

- `plan`
- `ask`
- `request-feedback-1`
- `handle-feedback-1`
- `final-verify`

Conditional phases remain visible in `phase-plan.yaml` and `checklist.md`; skipped or not-applicable phases require explicit reasons.

## Plan and checklist

- When KAB planner mode is used, KHS asks the planner backend to include a clearly delimited `KHS Checklist Seed` section in `plan.plan_text`.
- KHS captures `plan.plan_text` into `plan.md` before implementation approval/start.
- KHS/Hermes parses the seed, applies `registries/phase-contracts.yaml`, and writes the normalized progress tracker to `checklist.md`.
- KAB does not own the normalized checklist.
- `checklist.md` is updated after `ask` and after every later phase.

## Ask phase

- `ask` always runs after `plan`.
- If no question remains, record that explicitly instead of skipping the phase.
- Ask the master only when Hermes cannot decide safely from SOT/project docs or when user intent must be confirmed.

## Implementation approval

Hermes may auto-start low-risk work after notifying the master. Explicit master approval is required for:

- API changes
- DB/schema/migration changes
- security/auth/secret-handling changes
- dependency changes
- architecture changes
- SOT changes
- large diffs or broad file fanout
- low confidence
- unresolved ask-phase ambiguity

## KAS/KAH roadmap policy review governance

For active KAS/KAH roadmap policy, workflow, template, test, or shared skill
mirror work, implementation starts only after the backend-authored plan has
Blue synthesis plus Red and Orange plan-vet acceptance. Red/Orange reviewers
and Gray documentation/integrity review are resolved through the project/team
role registry when applicable, not hard-coded to individuals. Any requested plan
change returns to the same selected planner lane before implementation.

Official review evidence is limited to recorded color review, MAR role review,
project-Gray review, KAH artifacts, and accepted run evidence. `delegate_task`
calls, temporary subagents, and ad hoc advisor notes may support pre-review
analysis, but they are not official Red/Orange/Gray color review or MAR
evidence.

MAR is the only independent implementation review lane for active KAS/KAH
source policy, workflow, template, test, and shared skill mirror changes unless
주군 explicitly waives or replaces it before the run starts and the decision is
recorded in KAH/run evidence artifacts. Do not promote GLM Octo as a default,
optional, fallback, or legacy independent-review path in active KAS/KAH
artifacts.

When verification fails, Blue owns compact triage: reproduce or classify the
failure, route it to the selected implementer lane, and accept or reject the
resulting fix. The selected implementer owns detailed RCA, code/docs mutation,
and rerunning the affected verification.

For asynchronous review fan-in, attach the watcher only as an observer. When
direct Kanban tools are absent, check and use the durable Hermes Kanban CLI
surface before declaring review unavailable. Do not substitute `delegate_task`
or temporary subagents for official review cards or verdicts.

Fallback behavior fails closed by default. Do not add fallback behavior unless
it is necessary, bounded, evidenced, small, and accepted through the plan/review
path; otherwise stop and report the policy choice instead of silently widening
the system.

## Enhance-test and optimize

- `enhance-test` is conditional; skipping requires a reason when no test enhancement is useful or feasible.
- For code-change runs, `optimize` is conditional but strongly recommended.
- Optimize focuses on AI slop removal, duplicated logic, dead/verbose code, and small structural waste.
- Skipping optimize in a code-change run requires a reason in both `phase-plan.yaml` and `checklist.md`.

## Feedback loop

- Request feedback at least once for every KHS run.
- Hermes may request optional continuation rounds 2..5 when more feedback is useful.
- Maximum: five request-feedback/handle-feedback pairs.
- Each requested feedback round must have matching handling.
- If feedback has no actionable items, record `no actionable feedback`.

## Project graph relationship

Status: planning-confirmed KHS relationship; KAH `kkachi-agent-helper graph` support is implemented when effective-binary capabilities/help prove it. `.kkachi-workflow.yaml` is the project-level workflow graph instance for graph-managed runs. It may constrain or seed run-local `phase-plan.yaml`, but it does not replace the run artifact.

Fail-closed rule: if `.kkachi-workflow.yaml`, `registries/phase-contracts.yaml`, and `.kkachi/runs/<run_id>/phase-plan.yaml` disagree about phase applicability, dependencies, gates, or approvals, KHS must stop and ask the responsible role to confirm the intended authority. Do not silently prefer KHS defaults, `.kkachi/config.yaml`, generated diagrams, Kkachi v2 `.kkachi/config/workflows/`, or stale run state.

Stale marker: older wording that says `phase-plan.yaml` is the KHS workflow source of truth must be read as run-local source of truth for a run, not project graph authority.

## Final verification

Hermes owns final verification. Before reporting completion, Hermes checks:

- `phase-plan.yaml` final state
- `checklist.md` final state
- KAH gate/final gate results
- KAB evidence, not just dispatch success, when the run is KAB-backed or claims backend runtime evidence
- required artifacts and skip reasons
- test/verification evidence
- docs-update decision
- feedback rounds between one and three
- optimize evidence or skip reason for code-change runs

## Candidate graph record appendix

Date: 2026-05-21
Owner: KHS workflow/policy layer
Confirming role: Responsible approver / governance evidence record
Status: graph relationship addition to current phase orchestration policy; `kkachi-agent-helper graph` implemented, `kah graph` alias candidate
Authority level: KHS phase/run orchestration SOT; graph relationship remains capability-checked and KHS-policy-owned
Scope: KHS phase orchestration docs only
Related docs: `workflow-graph-integration.md`, `interface-contract.md`, `../roadmap.md`
Decision summary: `phase-plan.yaml` is run-local workflow/execution state for one run; `.kkachi-workflow.yaml` is project-level graph state when backed by KAH graph evidence.
Evidence/source paths: governance evidence records in kanban tasks `t_2fb00394`, `t_38cfc496`, and `t_2b460665`
Stale/conflict markers: older broad wording that `phase-plan.yaml` is the workflow SOT is narrowed to run-local scope.
Open questions: exact derivation/check behavior between project graph and run phase state remains future implementation work.
Next record action: implement remaining KHS graph relationship behavior through separately assigned capability-checked guidance/template tasks.
