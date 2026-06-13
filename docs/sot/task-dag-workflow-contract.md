# KAS task DAG workflow and custom trigger contract

Date: 2026-06-12
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record pending
Status: planning SOT for KAS `WFLOW` task-DAG workflow workstream; KASROLE v0.1.3 is the prerequisite release baseline and WFLOW completion targets KAS v0.1.4; not implementation behavior until roadmap tasks pass evidence and review gates
Authority level: KAS-side planning authority for task-DAG workflow policy, selector rules, node/agent contracts, generic trigger skills, and custom workflow skill scaffolding
Scope: `kkachi-agent-skills` docs, registries, templates, skills, and CLI planning; no KAH deterministic implementation claim, no KAB runtime change, no Hermes profile/provider/gateway/auth/token/model mutation, and no automatic workflow execution without explicit evidence/approval gates
Related docs: `docs/sot/workflow-graph-integration.md`, `docs/sot/graph-workflow-sync-compatibility.md`, `docs/sot/graph-template-registry.md`, `docs/sot/phase-orchestration-policy.md`, `docs/roadmap.md`, KAH `docs/sot/task-dag-state-machine.md`
Evidence/source paths:
- Master direction in 17번째 지구 Discord `#kas` thread `1514986770456903781` on 2026-06-12: support project-local multiple task DAGs, KAH order enforcement, KAS node-level agent/role contracts, generic and thin trigger skills, and a KAS custom workflow creator for user-specific scenarios.
- `WFLOW-001` is the logical cross-repo planning acceptance gate: the KAS SOT/roadmap/docs registration and the paired KAH `DAGSM` planning SOT/docs registration may be committed as `WFLOW-001` evidence. This does not complete or implement KAH `DAGSM-001`; `DAGSM-001` remains the subsequent KAH schema-validation/explain implementation task.
- Release-transition direction in 17번째 지구 Discord `#kah` thread `1515219219002818610` on 2026-06-14: KASROLE is complete at KAS v0.1.3, and WFLOW epic completion should target KAS v0.1.4.

## Decision summary

KAS will define the policy and operator-facing contract for task-DAG workflows while KAH remains the deterministic DAG state/evidence/order enforcer. The task-DAG workstream extends existing project workflow graph concepts; it does not replace `.kkachi-workflow.yaml`, `.kkachi/runs/<run_id>/phase-plan.yaml`, KAH graph proposal/apply evidence, KAB backend/session evidence, or Kanban as the long-lived team-member work bus.

A project may support multiple task DAGs. KAS must let an operator either select a workflow explicitly or use deterministic selector rules to choose a single eligible workflow. If zero or multiple workflows match, KAS must fail closed and ask for an explicit workflow choice rather than letting an LLM silently pick the "best" graph.

## Ownership boundary

| Layer | Owns | Must not own |
|---|---|---|
| KAS | workflow catalog policy, selector rules, task-class mapping, node agent/role/backend/skill contracts, trigger skill templates, custom workflow scaffold generation, KAH capability requirements | KAH node state mutation mechanics, graph apply/audit, backend runtime execution, automatic fallback agent selection |
| KAH | deterministic task DAG schema validation, dependency/order enforcement, workflow instance state, ready-node calculation, node FSM, required artifact/evidence checks, audit/events/diagnostics/gates | agent suitability, task policy choice, selector policy, prompt content, backend choice, review policy |
| KAB | backend/session/plan/read/status/wait/retained event evidence for executed nodes when selected | workflow policy authority, KAH state authority |
| Kanban | durable team-member assignment, dependency cards, long-lived review/evidence routing | KAH state replacement or hidden backend execution evidence |
| KAO / 황충 | Blue orchestration, approval routing, final synthesis, color-review fan-in | direct deterministic KAH state mutation without KAH commands/evidence |

## File and artifact model

Initial MVP planning uses these conceptual files. Exact paths may be refined by implementation tasks, but the ownership boundary is fixed:

| Artifact | Purpose | Owner |
|---|---|---|
| `.kkachi-workflow.yaml` or `.kkachi/workflow-catalog.yaml` | Project workflow catalog / active graph registry when KAH validates or applies it | KAS policy proposed through KAH validation/apply |
| `.kkachi/workflows/<workflow-id>.yaml` | Task-DAG definition for one workflow | KAS template/policy, KAH validation |
| `.kkachi/runs/<run_id>/workflow-instance.yaml` | Run-local task-DAG instance state | KAH |
| `.kkachi/runs/<run_id>/node-state.yaml` | Node state, evidence refs, and gate status | KAH |
| `registries/task-dag-workflow-*.yaml` | KAS selector/node-contract metadata | KAS |
| trigger skill `SKILL.md` | Operator entrypoint for generic or thin workflow starts | KAS / Hermes skill layer |

KAS must not directly hand-edit KAH authoritative graph state as a fallback. KAS may generate complete candidate workflow/catalog files and ask KAH to validate, diff, propose, or apply them through the supported KAH command surface.

## Workflow selection policy

KAS supports two trigger modes:

1. **Explicit workflow selection**: the operator or thin trigger skill names `workflow_id`.
2. **Selector-based workflow selection**: KAS evaluates deterministic metadata such as task class, labels, changed surfaces, risk level, required agents, required capabilities, and project policy.

Operator examples:

```text
kkachi-workflow-trigger --workflow security-release-flow --task <task-ref>
kkachi-workflow-trigger --task-class development --labels security,release --task <task-ref>
```

Selector result handling is fail-closed:

| Candidate count | Required behavior |
|---:|---|
| 0 | BLOCKED: no matching workflow; ask for explicit workflow or create a candidate workflow through the custom creator |
| 1 | May proceed after KAH capability/catalog/workflow validation passes |
| 2+ | BLOCKED: ambiguous workflow; ask operator to choose or narrow metadata |

LLM judgment may explain options, but it must not override selector ambiguity, KAH validation failure, missing capability evidence, approval requirements, or node evidence checks.

## Node contract requirements

Each KAS node contract must declare at least:

- `workflow_id` and `node_id`;
- task class or applicability condition;
- owner agent or logical role, for example `planner_backend`, `implementer_backend`, `hahuyeon`, `yeomong`, `jingung`, `hwangchung`;
- execution lane expectation: direct KAS skill, KAB backend, Kanban/team-member card, manual approval, or no-agent evidence step;
- required inputs and artifact/evidence outputs;
- prompt/skill template or reference;
- approval requirement when the node mutates code, docs, workflow state, profile skills, release state, or other gated surfaces;
- fallback policy, normally `none_fail_closed`;
- verification and completion gate reference.

KAS must not mark a node complete. Node completion is a KAH state transition backed by the required artifacts/evidence.

## Trigger skill posture

The generic trigger skill should be implemented before thin or custom trigger generation. `WFLOW-002` is explicitly scoped to `workflow_id` execution only: it may accept an explicit workflow id plus an explicit node-contract source/ref and must not implement selector search or a full registry. Selector-driven workflow choice and formal node-contract registry behavior are deferred to `WFLOW-003`. Its MVP responsibilities are:

1. read the project workflow catalog or explicit workflow id;
2. check effective KAH capabilities and workflow validation;
3. choose exactly one workflow or fail closed;
4. create or resume a KAH workflow instance;
5. ask KAH for ready nodes;
6. render KAS node contracts into dispatch packets;
7. dispatch through the selected lane only when the task contract and approval state permit it;
8. record node evidence paths and ask KAH to advance node state;
9. summarize final state from KAH evidence.

Thin trigger skills are allowed only as wrappers around the generic trigger, for example `project-release-trigger` calling the generic trigger with `workflow_id=release`. Full custom trigger skills require a recorded reason that generic trigger plus node contract metadata is insufficient.

## Custom workflow creator posture

KAS should provide a custom workflow creator because KAS understands KAH capability boundaries and Hermes skill structure. The creator must be dry-run first and should support three modes:

| Mode | Purpose | Default posture |
|---|---|---|
| `dag_only` | Generate a workflow DAG and catalog entry for the generic trigger | Preferred default |
| `thin_trigger` | Generate a small trigger skill that pins a workflow id | Allowed after the DAG is validated |
| `full_trigger` | Generate a custom trigger skill with scenario-specific input/dispatch logic | Exceptional; requires explicit reason and review |

The creator must emit a dry-run packet containing generated paths, graph/catalog candidate, node contracts, trigger skill plan, selector metadata, KAH validation expectations, approval requirements, and an apply hash. Apply must recompute and match a canonical dry-run hash before writing. The canonical apply hash is `sha256` over a deterministic UTF-8 JSON payload with sorted keys and normalized relative paths; implementation may add a versioned canonicalization field, but must not use ambiguous prose or timestamp-dependent content as the hash input. The hash must bind target paths, candidate DAG/catalog content, node contracts, trigger plan, selector metadata, approval scope, effective KAH/KAS capability-version evidence, base graph/catalog checksums when present, changed-path set, conflicts/diagnostics, and no-write evidence. Any mismatch fails closed. The dry-run output must include both a compact operator summary and the full machine packet so 주군 can approve from stable deltas without reading raw hash inputs. It must not mutate installed profile skills, project `.kkachi` state, source KAS files, or KAH state without explicit approval evidence.

Mode output expectations:

| Mode | Typical dry-run generated paths | Apply posture |
|---|---|---|
| `dag_only` | `.kkachi/workflows/<workflow-id>.yaml`, catalog/proposal packet, node-contract packet | KAH validate/propose/apply when authoritative graph state changes |
| `thin_trigger` | `dag_only` outputs plus one `<project>-<workflow-id>-trigger/SKILL.md` wrapper plan | Requires skill-scope approval and regenerated hash |
| `full_trigger` | `thin_trigger` outputs plus custom input/dispatch sections | Exceptional; requires explicit reason and color review |

## Required roadmap sequence

The post-KASROLE workstream is intentionally linear until the substrate is proven. KASROLE v0.1.3 is the prerequisite release baseline, and completing this WFLOW epic should produce the KAS v0.1.4 target release:

1. `WFLOW-001` — accept this KAS SOT plus the paired KAH `DAGSM` planning SOT/docs registration as the cross-repo planning contract.
2. `DAGSM-001` — KAH task-DAG schema and validation/explain substrate, dependent on `WFLOW-001` policy acceptance.
3. `DAGSM-002` — KAH workflow instance, node FSM, ready-node calculation, and evidence-gated completion, dependent on `DAGSM-001`.
4. `WFLOW-002` — KAS generic workflow trigger for explicit `workflow_id` only, dependent on `DAGSM-002`; selector search and full node-contract registry are deferred.
5. `WFLOW-003` — KAS selector and node contract registry, dependent on `WFLOW-002` and `DAGSM-002`.
6. `DAGSM-003` — KAH multi-DAG catalog, diagnostics, and gate integration, dependent on `DAGSM-002` and `WFLOW-003` metadata shape.
7. `WFLOW-004` — KAS custom workflow creator and optional thin trigger generator, dependent on `WFLOW-003` and `DAGSM-003`.

Implementation must not skip ahead from KAS trigger work to custom workflow generation before KAH node-state enforcement exists. `WFLOW-002` must stay explicit-workflow-only until `WFLOW-003` provides selector and registry contracts.

## Non-goals and deferrals

- No automatic KAH/KAS binary update.
- No automatic graph apply from cron/CI.
- No arbitrary webhook runtime in the MVP.
- No hidden fallback from one agent/backend to another.
- No dynamic node generation during execution until a later approved design.
- No retry/rollback automation in the MVP beyond deterministic fail/block state recording.
- No KAB graph policy authority.
- No merge/fallback with Kkachi v2 `.kkachi/config/workflows/`.
- No Hermes profile/provider/gateway/auth/token/model mutation.

## Acceptance gates for WFLOW workstream

Before KAS claims task-DAG workflow support:

- KAH advertises the required DAG/FSM/catalog capabilities through current effective-binary evidence.
- KAS trigger and selector behavior has deterministic tests for explicit, unique, zero-match, and ambiguous workflow selection.
- KAS node contracts preserve agent/role/backend boundaries and never complete nodes without KAH evidence.
- Custom workflow creator defaults to dry-run, emits approval-hash-bound plans, and refuses direct `.kkachi-workflow.yaml` edit fallback.
- Red/Orange/Gray review gates have no unresolved blocking findings for policy, operator UX, and evidence trace.
