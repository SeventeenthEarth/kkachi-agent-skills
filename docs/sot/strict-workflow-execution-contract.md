# STRICT workflow execution contract

Date: 2026-06-19
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / Blue command with Red, Orange, and project-Gray review evidence accepted for `STRICT-001`
Status: accepted SOT for the shared `STRICT` epic; `STRICT-001` docs/SOT registration is complete and does not implement runtime behavior
Authority level: KAS-side planning authority for classification-selected strict workflow execution policy, dispatch contracts, task slices, and PR-candidate sequencing
Scope: `kkachi-agent-skills` docs, registries, skills, templates, CLI orchestration, dispatch packets, and agent-facing execution guidance. No KAH deterministic implementation claim, KAB runtime activation, profile/provider/gateway/auth/token/model mutation, install, release, push, or automatic rollback is authorized by this document.
Related docs: `docs/sot/task-dag-workflow-contract.md`, `docs/sot/kas-cli-contract.md`, `docs/sot/phase-orchestration-policy.md`, `docs/roadmap.md`, KAH `docs/sot/task-dag-state-machine.md`, KAH `docs/sot/strict-workflow-enforcement.md`
Evidence/source paths:
- 주군 direction in 17번째 지구 Discord `#kas` thread `1517399560626901034` on 2026-06-19: after task classification chooses a KAS/KAH workflow spine, agents should strictly follow that selected order rather than merely receiving realtime warnings; KAS should request next node ids from KAH and execute only those nodes.
- Prior WFLOW/DAGSM baseline: KAS `WFLOW-007..009` provide classification-to-bundle routing, run-local materialization, and promotion; KAH `DAGSM-001..006` provide task-DAG validation, workflow instance state, node FSM, ready-node calculation, catalog diagnostics, final gate integration, and catalog proposal/apply substrate.

## Purpose

`STRICT` hardens the post-WFLOW execution model so a Kkachi agent cannot treat the selected task-class workflow as advisory. Once task classification resolves to a standard or custom workflow, KAS must use that workflow as the execution-order authority and KAH must own the authoritative state transitions.

The target behavior is not realtime notification. The target behavior is strict execution admission:

1. KAS records task classification and routes it to exactly one selected workflow.
2. KAS materializes or resumes the run-local workflow through KAH.
3. KAS asks KAH for the current ready node set and instance revision.
4. KAS renders dispatch packets only for KAH-ready nodes.
5. The agent must successfully claim/start the dispatched node in KAH before doing the node work.
6. The agent may claim completion only by providing required evidence to KAH node completion.
7. Final gates verify the KAH transition ledger, workflow instance state, and phase-plan/checklist projection.

## Core policy

Classification-selected workflow execution is fail-closed:

- If task classification is missing, ambiguous, or unsupported, KAS must not invent a default development spine.
- If `workflow-route` does not produce exactly one selected bundle, KAS must ask for clarification or explicit workflow choice.
- If KAH workflow capability or instance creation/resume fails, KAS must not dispatch implementation work.
- If KAH reports no ready nodes, KAS must not infer the next node from skill order or phase-plan text.
- If a dispatch packet is stale, KAS must refresh KAH ready state rather than continue with the stale node.
- If an agent attempts a node outside the KAH ready set, KAH must reject the transition and the KAS run must stop or repair through an explicit approved path.

`development_full`, `docs_only_light`, `research_evidence_light`, `review_light`, `bootstrap_config`, and `direct_report` remain different selected spines. A docs-only run must not be upgraded into the development spine just because an agent knows the development workflow.

## Ownership boundary

| Layer | Owns in `STRICT` | Must not own |
|---|---|---|
| KAS | task classification consumption, workflow-route selection, dispatch packet shape, node contract policy, agent-facing instruction, stale packet handling policy, strict execution templates | KAH state mutation internals, direct workflow-instance file writes, automatic rollback, backend execution authority |
| KAH | workflow-managed run marker validation, workflow instance state, ready-node calculation, node claim/start/complete/block transitions, transition ledger, DAG order verification, final gate/projection checks | task classification, bundle selection, node owner/prompt/backend selection, review policy, KAS phase applicability policy |
| KAB | backend/session execution evidence for a claimed node when selected | workflow order authority, KAH state replacement, silent fallback backend selection |
| Kanban | durable team-member delegation/review/evidence bus | KAH transition ledger replacement |
| KAO / 황충 | Blue synthesis, approval routing, cross-repo sequencing, review fan-in | bypassing KAS dispatch or KAH transition gates |

## Strict execution model

### 1. Route and materialize

KAS starts from an already-classified task contract. `workflow-route` produces a single route result with task class, classification reason, selected bundle, selected workflow id, and required capabilities. `workflow-trigger` consumes that route result and materializes or resumes the run-local workflow. KAS must preserve route result, materialization evidence, selected workflow id, and KAH capability evidence as run artifacts.

### 2. Ready-node admission

KAS asks KAH for the current ready node set and workflow instance revision. The ready set is the allowed execution set. In a linear spine the set usually has one node; in a DAG it may contain multiple independent nodes. Strict order means respecting the selected DAG topological constraints, not imposing an artificial single list when the DAG allows fan-out.

### 3. Dispatch packet as execution permit

A strict dispatch packet must include at least:

```json
{
  "strict_order": true,
  "workflow_id": "development_full",
  "run_id": "run-...",
  "node_id": "plan",
  "instance_revision": 3,
  "expected_start_revision": 3,
  "completion_authority": "kah_only",
  "direct_kah_state_write": false,
  "fallback_policy": "none_fail_closed"
}
```

The packet is not a completion authority. It is an execution permit for one KAH-ready node at one observed instance revision.

### 4. Start before work

An agent must call KAH node start/claim before performing node work. If KAH rejects the start because the node is not ready, the revision is stale, dependencies are incomplete, or the node state is not pending, the agent must not execute the node.

### 5. Complete after evidence

An agent must call KAH node complete only after the node artifacts/evidence exist. Completion must fail closed when required outputs are missing, evidence paths are unsafe, the node is not running, or the revision is stale.

### 6. Reject instead of rollback by default

Unexpected node ids should not be appended and then rolled back. The default safe behavior is to reject the transition before it mutates KAH state. If files were already modified before a valid KAH start, the run should enter blocked/violation repair handling; automatic rollback is deferred because it can destroy unrelated work and requires stronger worktree/checkpoint policy.

## Shared STRICT task sequence

`STRICT` uses one shared cross-repo task id sequence so KAS and KAH numbers do not collide.

| Task ID | Repo | Title | Status | Outcome |
|---|---|---|---|---|
| `STRICT-001` | KAS | Strict workflow execution SOT and roadmap registration | Completed | Registered the shared epic, KAS policy contract, KAH companion SOT link, and cross-repo PR-candidate sequence. |
| `STRICT-002` | KAH | Workflow-managed run marker and strict final-gate mode | In Progress | KAH source-side implementation adds workflow-managed run markers and final-gate missing-marker/absence/mismatch failures; review/final closeout remains pending before the shared sequence advances to `STRICT-003`. |
| `STRICT-003` | KAS | Classification route/trigger mandatory orchestration | Planned | Classified KAS/KAH runs must route, materialize/resume, and record selected workflow evidence before dispatch. |
| `STRICT-004` | KAH | Node claim ledger and transition-order verification | Planned | KAH records/verifies append-only node transition order against the selected DAG. |
| `STRICT-005` | KAS | Dispatch packet expected-revision and node execution guard | Planned | Dispatch packets include current revision/ready-node evidence and require KAH start success before backend/agent work. |
| `STRICT-006` | KAH | Phase-plan projection and workflow consistency gate | Planned | Workflow-managed phase-plan/checklist evidence must project from the workflow instance rather than contradict it. |
| `STRICT-007` | KAS | Strict orchestration skill/templates/e2e adoption | Planned | Active KAS skills, runner prompts, templates, and fixtures enforce strict dispatch/start/complete flow. |

## Acceptance criteria for STRICT-001

- This KAS SOT exists and records the strict execution policy, ownership boundary, task sequence, and deferrals.
- KAS `docs/roadmap.md` registers the `STRICT` epic and KAS-owned rows `STRICT-001`, `STRICT-003`, `STRICT-005`, and `STRICT-007`.
- KAS `docs/README.md` and `docs/kkachi-docs-map.yaml` reference this SOT.
- The paired KAH SOT `docs/sot/strict-workflow-enforcement.md` and KAH roadmap/docs index/map references exist.
- Verification includes docs readback, docs-map YAML parse for both repos, `git diff --check`, and repository test commands or explicit degraded-evidence blockers.

## Deferrals and non-goals

- No realtime alerting/watchers are introduced by this epic.
- No automatic rollback/revert is introduced without a later explicit worktree/checkpoint policy.
- No dynamic node generation, retry automation, fallback agent/backend selection, or broad custom workflow promotion is authorized.
- No KAB runtime activation, provider/model/gateway/auth/token mutation, profile skill installation, release tagging, push, or commit is authorized by this SOT alone.
- No task may claim strict runtime behavior until its repo implementation, tests, docs, review gates, final verification, and effective-binary/skill evidence are complete.

## Next action

Complete review/final closeout for KAH `STRICT-002`; after acceptance, advance to KAS `STRICT-003` so classification route/trigger becomes mandatory before dispatch.
