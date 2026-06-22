# KAS task DAG workflow and custom trigger contract

Date: 2026-06-12
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record pending
Status: current SOT for KAS `WFLOW` task-DAG workflow workstream; WFLOW-001..009 are implemented and make up the KAS v0.1.4 workflow release baseline, with persistent promotion dependent on effective KAH DAGSM-006 / v0.1.10 capability evidence
Authority level: KAS-side planning authority for task-DAG workflow policy, selector rules, node/agent contracts, generic trigger skills, and custom workflow skill scaffolding
Scope: `kkachi-agent-skills` docs, registries, templates, skills, and CLI planning; no KAH deterministic implementation claim, no KAB runtime change, no Hermes profile/provider/gateway/auth/token/model mutation, and no automatic workflow execution without explicit evidence/approval gates
Related docs: `docs/sot/workflow-graph-integration.md`, `docs/sot/graph-workflow-sync-compatibility.md`, `docs/sot/graph-template-registry.md`, `docs/sot/phase-orchestration-policy.md`, `docs/roadmap.md`, KAH `docs/sot/task-dag-state-machine.md`
Evidence/source paths:
- Master direction in 17번째 지구 Discord `#kas` thread `1514986770456903781` on 2026-06-12: support project-local multiple task DAGs, KAH order enforcement, KAS node-level agent/role contracts, generic and thin trigger skills, and a KAS custom workflow creator for user-specific scenarios.
- `WFLOW-001` is the logical cross-repo planning acceptance gate: the KAS SOT/roadmap/docs registration and the paired KAH `DAGSM` planning SOT/docs registration may be committed as `WFLOW-001` evidence. This does not complete or implement KAH `DAGSM-001`; `DAGSM-001` remains the subsequent KAH schema-validation/explain implementation task.
- Release-transition direction in 17번째 지구 Discord `#kah` thread `1515219219002818610` on 2026-06-14: KASROLE is complete at KAS v0.1.3, and WFLOW epic completion should target KAS v0.1.4.
- Master direction in 17번째 지구 Discord `#kas` thread `1516002725689819209` on 2026-06-16: extend the existing WFLOW/DAGSM sequence instead of creating a separate BWFLOW axis; record WFLOW-005 and DAGSM-004 as paired SOT-alignment tasks before implementation; make run-local ephemeral workflows the default and promote to project-local custom workflows only by explicit approval.

## Decision summary

KAS will define the policy and operator-facing contract for task-DAG workflows while KAH remains the deterministic DAG state/evidence/order enforcer. The task-DAG workstream extends existing project workflow graph concepts; it does not replace `.kkachi-workflow.yaml`, `.kkachi/runs/<run_id>/phase-plan.yaml`, KAH graph proposal/apply evidence, KAB backend/session evidence, or Kanban as the long-lived team-member work bus.

A project may support multiple task DAGs. KAS must let an operator either select a workflow explicitly or use deterministic selector rules to choose a single eligible workflow. If zero or multiple workflows match, KAS must fail closed and ask for an explicit workflow choice rather than letting an LLM silently pick the "best" graph.

WFLOW-005 extends the SOT from substrate completion into the adoption layer: task classification should resolve to a standard bundle workflow when possible, and the default runtime shape for task-specific or custom workflows is run-local ephemeral materialization under `.kkachi/runs/<run_id>/...`. Project-local persistent workflow registration is a later explicit promotion path, not an install-time side effect and not a silent `.kkachi-workflow.yaml` or catalog write.

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
| `.kkachi-workflow.yaml` or `.kkachi/workflow-catalog.yaml` | Project-local persistent workflow catalog / active graph registry when explicitly promoted or KAH validates/applies it; not the default path for bundle execution | KAS policy proposed through KAH validation/apply |
| `.kkachi/workflows/<workflow-id>.yaml` | Task-DAG definition for one workflow | KAS template/policy, KAH validation |
| `.kkachi/runs/<run_id>/workflow.yaml` or equivalent run-local task-DAG file | Ephemeral per-run workflow materialized from a bundle or task-specific classification | KAS proposes content; KAH validates/creates the instance |
| `.kkachi/runs/<run_id>/workflow-instance.json` | Run-local task-DAG instance state | KAH |
| `.kkachi/runs/<run_id>/node-state.yaml` or equivalent node projection | Node state, evidence refs, and gate status | KAH |
| `registries/task-dag-workflow-*.yaml` | KAS selector/node-contract metadata | KAS |
| trigger skill `SKILL.md` | Operator entrypoint for generic or thin workflow starts | KAS / Hermes skill layer |

KAS must not directly hand-edit KAH authoritative graph state as a fallback. KAS may generate complete candidate workflow/catalog files and ask KAH to validate, diff, propose, or apply them through the supported KAH command surface.

## Bundle workflow and run-local ephemeral policy

WFLOW-005 records the adoption-layer target before implementation. Standard bundle workflows are KAS-owned policy templates for common task classes; they are not KAH-selected behavior. The initial bundle vocabulary is:

| Bundle workflow | Primary task class | Default posture |
|---|---|---|
| `development_full` | `development` | Full KAS/KAH development spine with plan, implementation, verification, docs, review, feedback handling, and final verification. |
| `docs_only_light` | `docs_only` | Durable docs/SOT/roadmap shaping with docs validation and explicit skipped-phase reasons. |
| `research_evidence_light` | `research_evidence` | Read-only evidence collection, source citation, and final report without implementation/test/optimize phases. |
| `review_light` | `collaboration_review` | Read-only review or risk triage with preserved feedback evidence and no direct mutation. |
| `bootstrap_config` | `bootstrap_config` | Repository/helper/profile/tooling setup with approval gates for unsafe or external mutations. |
| `direct_report` | `simple_command_report` | Bounded command/status evidence and final report; normally outside KHS unless the master explicitly keeps it inside a run. |

Classification-to-bundle routing must be deterministic and fail closed. If task classification is missing, ambiguous, or maps to more than one eligible bundle, KAS must ask for a narrowed task contract or explicit workflow choice. KAH must not classify the task, rank bundles, or choose the workflow.

Run-local ephemeral workflow materialization is the default for bundle execution and one-off custom task flows. The generated DAG and node contracts live under `.kkachi/runs/<run_id>/...`, are validated by the effective KAH workflow command surface, and become authoritative for that run only after KAH creates the workflow instance. Promotion from a run-local workflow to a project-local persistent workflow requires a separate dry-run, hash-bound approval, KAH capability evidence, and later WFLOW/DAGSM promotion tasks.

## Workflow selection policy

Through WFLOW-004, KAS supports two workflow selection modes:

1. **Explicit workflow selection**: the operator or thin trigger skill names `workflow_id`.
2. **Selector-based workflow selection**: KAS evaluates deterministic metadata such as task class, labels, changed surfaces, risk level, required agents, required capabilities, and project policy.

WFLOW-005+ adds the planned adoption-layer route: deterministic task-classification-to-bundle routing. That route must resolve to exactly one standard bundle before run-local materialization; missing, no-match, or ambiguous classification fails closed, and KAH must not classify, rank, or choose bundles.

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

WFLOW-003 implements the selector registry as KAS-owned metadata at
`registries/task-dag-workflow-registry.yaml` with schema version
`kas-task-dag-workflow-registry/v1`. Selector mode is requested by
`--selector-registry` or selector fields, requires an explicit registry path,
and accepts task class, labels, changed surfaces, risk, required agents, and
required capabilities. Explicit
`--workflow-id` / `--node-contract-source` inputs and selector inputs are
mutually exclusive; mixed mode fails closed with
`selector_explicit_mode_conflict`.

Selector diagnostics are deterministic and non-ranking:

| Diagnostic | Meaning |
|---|---|
| `selector_registry_required` | Selector fields were supplied without a registry path |
| `selector_registry_unreadable` | Registry path could not be read |
| `selector_registry_schema_unsupported` | Registry version or YAML shape is unsupported |
| `selector_required_input_missing` | Required selector input such as task class is missing |
| `selector_no_match` | Zero workflows matched; KAS must not call KAH workflow instance commands |
| `selector_ambiguous` | Multiple workflows matched; KAS must not choose first or rank candidates |
| `selector_explicit_mode_conflict` | Explicit and selector inputs were combined |

WFLOW-007 implements classification-to-standard-bundle routing through the
separate route-only `workflow-route` CLI. The input is already-classified
metadata, not raw natural-language task inference. KAS validates the task class
or declared taxonomy alias from `registries/task-taxonomy.yaml`, preserves the
classification reason, then uses the standard bundle registry to resolve
exactly one bundle. Route output records the selected bundle/workflow id,
workflow path, work path/mode, execution mode, skipped-phase reasons, required
capabilities, taxonomy checksum, registry checksum, diagnostics, and
`direct_kah_state_write:false`.

DESIGN-003 extends WFLOW-007 route evidence with explicit
`teal_applicability`. `workflow-route` records `teal_applicability` by deriving
`teal_required` from `project_has_teal_lane && ui_ux_change`. Missing explicit
Teal facts fail closed with `teal_applicability_required`; a false derivation
requires a concrete `teal_skip_reason`. Route output may record required Teal
verdict expectations, but it remains route-only and must not create design
workflow nodes.

Route diagnostics are deterministic and non-ranking:

| Diagnostic | Meaning |
|---|---|
| `classification_required_input_missing` | No task class was supplied |
| `classification_reason_missing` | Classification reason was not supplied |
| `classification_class_unsupported` | Task class or alias is not declared by the taxonomy |
| `teal_applicability_required` | Explicit project_has_teal_lane and ui_ux_change facts were not supplied |
| `teal_skip_reason_required` | Teal is not required but the concrete skip reason is absent |
| `taxonomy_required` / `taxonomy_unreadable` / `taxonomy_schema_unsupported` | Taxonomy path, read, or schema validation failed |
| `bundle_registry_required` / `bundle_registry_unreadable` / `bundle_registry_schema_unsupported` | Bundle registry path, read, or schema validation failed |
| `bundle_default_spine_missing` | Taxonomy class does not declare a default bundle spine |
| `bundle_no_match` | Classification matched zero standard bundles |
| `bundle_ambiguous` | Classification matched multiple standard bundles |
| `bundle_selected_mismatch` | Explicit selected spine conflicts with the deterministic registry match |

`workflow-route` must not call KAH workflow create/show/ready/node APIs, render
dispatch packets, materialize `.kkachi/runs/<run_id>/workflow.yaml`, promote or
apply project workflow catalogs, choose first, score, rank, or use an LLM
tie-break. WFLOW-008 owns run-local workflow materialization, and WFLOW-009 plus
optional DAGSM-006 own promotion/apply.

WFLOW-008 implements run-local materialization through either
`workflow-trigger --route-result <workflow-route-json> --materialize-run-local
--run <run-id>` for selected standard bundles, or `workflow-trigger
--custom-workflow-packet <workflow-create-dry-run.json> --approval
dry-run:sha256:<hash> --materialize-run-local --run <run-id>` for approved
one-off custom DAGs. The route-result path consumes the existing WFLOW-007
route result from a file; it does not reroute raw task metadata, infer a class,
rank bundles, or choose a workflow inside the trigger. The custom packet path
consumes the existing WFLOW-004 workflow-create dry-run machine packet shape,
requires approval-hash-bound evidence, and recomputes the hash with
WFLOW-004 canonical approval semantics before any write; this is not a second
hash algorithm. KAS preflights effective KAH workflow capability before non-dry-run materialization.
Installed KAH `0.1.9` lacks the workflow command group. If the installed helper
lacks the `workflow` command group, KAS fails closed with
`blocked_missing_kah_workflow_capability` before writing run-local artifacts.

DESIGN-003 route-result materialization consumes and validates
`teal_applicability` from `workflow-route`. Missing or invalid route
applicability fails closed with `route_result_teal_applicability_required` or a
specific invalid-applicability status before run-local writes. `workflow-trigger` route-backed materialization injects Teal nodes only when `teal_required=true`:
`DESIGN_PLAN_GATE` is inserted before implementation authorization and
`DESIGN_FIDELITY_REVIEW` before final acceptance. Non-UI Kkachi source work
with `teal_required=false` and a concrete `teal_skip_reason` receives no design
nodes. These conditional run-local materialized nodes are not universal nodes in `registries/task-dag-workflow-registry.yaml`.

Ordinary Red/Orange/Gray/Blue review, MAR, backend evidence, helper notes, and
temporary subagents remain separate evidence lanes and are not substitutes for
required Teal verdicts.

The WFLOW-008 run-local layout is:

```text
.kkachi/runs/<run_id>/workflow/materialization.json
.kkachi/runs/<run_id>/workflow/workflow.yaml
.kkachi/runs/<run_id>/workflow/node-contracts.json
.kkachi/runs/<run_id>/workflow/route-result.json              # route-result source only
.kkachi/runs/<run_id>/workflow/custom-workflow-packet.json    # approved one-off source only
.kkachi/runs/<run_id>/workflow/checksums.json
```

`materialization.json` records source registry/taxonomy checksums, selected
bundle, task class, classification reason, run-local posture, and
`no_promotion:true` for route-result sources. For approved custom packet
sources it records the packet path/checksum, approval evidence, dry-run plan
hash, approved plan hash, run-local posture, `persistent_promotion:false`, and
`no_promotion:true`. `workflow.yaml` is the run-local task-DAG file passed
explicitly to KAH. `node-contracts.json` is a `kas-node-contracts/v1` bundle
rendered from KAS node contracts or copied from the approved generated node
contracts in the WFLOW-004 packet. KAS must pass the explicit run-local
`--workflow-file` to KAH `workflow validate`, `workflow explain`, and `workflow
create`; it must not assume `.kkachi/workflows/<workflow-id>.yaml` for
run-local execution.

WFLOW-008 rejects unsafe paths, symlink/path escapes, project-local
`.kkachi/workflows`, `.kkachi/workflow-catalog.yaml`, `.kkachi-workflow.yaml`,
profile/provider/gateway/auth/token/model paths, KAB runtime paths, direct KAH
workflow-instance writes, fallback agent/backend selection, retry/rollback
automation, dynamic node generation, arbitrary webhook runtime, and WFLOW-009
promotion/apply.

WFLOW-009 implements explicit promotion proposal through a separate
`workflow-promote` command, not by overloading `workflow-trigger`:

```text
workflow-promote --project <path> --run <run-id> --target-workflow-id <id> --reuse-reason <reason> [--thin-trigger] --dry-run --json
workflow-promote --project <path> --run <run-id> --target-workflow-id <id> --reuse-reason <reason> [--thin-trigger] --apply dry-run:sha256:<hash> --json
```

The source is an existing WFLOW-008 run-local bundle containing
`materialization.json`, `workflow.yaml`, `node-contracts.json`,
`checksums.json`, and source/checksum evidence. `--target-workflow-id` and
`--reuse-reason` are required. `--dry-run` and
`--apply dry-run:sha256:<hash>` are mutually exclusive. Dry-run emits
`kas-workflow-promote-packet/v1` with source run/materialization provenance,
target project-local paths, generated workflow/catalog/node-contract registry
content, optional thin trigger content, base checksums, source checksums,
changed paths, diagnostics/conflicts, no-write evidence, KAS/KAH capability
evidence, and a deterministic approval hash.

The WFLOW-009 approval hash binds source materialization provenance, source
checksums, target paths, generated content, trigger plan, capability evidence,
base checksums, changed paths, diagnostics/conflicts, and no-write evidence.
Apply recomputes the hash before any apply decision. Persistent project-local
apply requires effective KAH DAGSM-006 / v0.1.10 `workflow catalog propose/apply`
and `workflow_catalog_proposal_apply=true` evidence; without that evidence, apply remains
non-authoritative and fail-closed, and it must not direct-write
`.kkachi/workflows/*`, `.kkachi/workflow-catalog.yaml`, `.kkachi-workflow.yaml`,
profile files, KAH state, KAB state, auth/token/provider/gateway/model config,
or fallback backend selection. Generated evidence preserves
`direct_kah_state_write:false`, `completion_authority:kah_only`, and
`fallback_policy:none_fail_closed`.

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

DESIGN-002 adds KAS-owned Teal node-contract semantics without adding universal
Teal nodes to the standard Kkachi source-work bundle. The applicability fields
are `project_has_teal_lane`, `ui_ux_change`, derived
`teal_required = project_has_teal_lane && ui_ux_change`, `teal_skip_reason`
when either input is false, and bounded waiver evidence fields when a required
Teal gate is approved as waived. A skip reason is not waiver evidence; waiver
evidence is not clean completion when the task still lacks the recorded
approval fields.

When `teal_required=true`, `DESIGN_PLAN_GATE` is the Teal design planning gate
before implementation authorization, and `DESIGN_FIDELITY_REVIEW` is the Teal
fidelity review gate before final acceptance. These are distinct node-contract
semantics from normal implementation, verification, and review nodes. The
ordinary Red/Orange/Gray/Blue color review remains separate, MAR remains separate,
and backend implementation evidence remains separate. They must not
be represented as substitutes for required Teal verdicts. Missing required Teal
evidence uses `required_teal_verdict_missing` and the existing
`none_fail_closed` posture.

KAS must not mark a node complete. Node completion is a KAH state transition backed by the required artifacts/evidence.

WFLOW-003 registry node contracts require the WFLOW-002 core fields plus
`task_class`. JSON node-contract bundles used by explicit WFLOW-002 mode and
YAML registry node contracts share the same core validation path; the registry
adds the selector-specific `task_class` invariant. If KAH `workflow explain`
returns reliable node ids, selector-mode registry contracts must exactly cover
those ids before KAS creates or resumes an instance. If KAH does not expose
reliable node ids, KAS records an informational diagnostic and still fails
closed later for any KAH-ready node without a matching contract.

Dispatch packets must preserve `direct_kah_state_write:false` and include
`completion_authority: kah_only`. Stage 1 direct Codex SDK/app-server evidence
must be reported separately from KAB `native_codex` evidence with
`stage1_direct_codex_is_kab_native_codex:false`.

## Trigger skill posture

The generic trigger skill should be implemented before thin or custom trigger generation. `WFLOW-002` is explicitly scoped to `workflow_id` execution only: it may accept an explicit workflow id plus an explicit node-contract source/ref and must not implement selector search or a full registry. `WFLOW-003` adds deterministic selector-driven workflow choice and a formal node-contract registry without adding custom workflow creation, thin trigger scaffolding, scoring, LLM tie-breaks, backend fallback, or direct KAH state writes. Its MVP responsibilities are:

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

WFLOW-004 implements this posture through
`kkachi-agent-skills workflow-create --mode dag_only|thin_trigger|full_trigger`.
The machine packet uses `kas-workflow-create-packet/v1`; approval uses
`kas-workflow-create-approval/v1`; the canonical `sha256` binds command, mode,
target paths, generated content, selector metadata, KAH/KAS capability evidence,
base checksums, changed paths, conflicts/diagnostics, and no-write evidence.
Fail-closed diagnostics include `approval_plan_hash_mismatch`,
`unsafe_target_path`, `selector_metadata_invalid`,
`workflow_creator_mode_unsupported`,
`blocked_missing_kah_workflow_capability`,
`blocked_missing_kah_workflow_catalog_capability`,
`base_catalog_unreadable`, `base_checksum_mismatch`, and
`generated_skill_validation_failed`.

Older installed KAH `0.1.9` lacks the workflow command group, so WFLOW-004 apply
remains non-approvable/fail-closed under that effective runtime. KAH `0.1.10`
adds the workflow command group and DAGSM-006 workflow catalog proposal/apply
capability, but KAS must still capture effective binary capability/help evidence
before any run relies on it. Missing, stale, or insufficient capability evidence
keeps apply fail-closed; KAS must not invent a direct write or apply fallback.
This preserves no automatic fallback for ambiguous selectors, unsupported modes,
missing capability, unsafe paths, hash mismatch, unreadable or mismatched base
catalog evidence, generated-skill validation failure, profile/provider mutation,
KAB graph authority, arbitrary webhook runtime, retry/rollback automation, and
dynamic node generation during execution.

Mode output expectations:

| Mode | Typical dry-run generated paths | Apply posture |
|---|---|---|
| `dag_only` | For WFLOW-004 project-local proposal mode: `.kkachi/workflows/<workflow-id>.yaml`, catalog/proposal packet, node-contract packet. For WFLOW-005+ bundle or one-off defaults: run-local `.kkachi/runs/<run_id>/...` materialization until explicit WFLOW-009 promotion is approved. | KAH validate/propose/apply only when authoritative graph/catalog state changes and reviewed capability exists; otherwise proposal-only/fail-closed |
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
8. `WFLOW-005` plus KAH `DAGSM-004` — paired SOT/roadmap alignment for bundle workflows, deterministic classification routing, run-local ephemeral defaults, effective KAH capability evidence, and explicit promotion boundaries.
9. `WFLOW-006` — implement standard bundle workflow registry/templates and KAH-compatible DAG rendering.
10. `DAGSM-005` — harden effective workflow capability, installed-binary alignment, and KAS-generated DAG compatibility evidence where KAH support is required.
11. `WFLOW-007` — implement task classification to bundle routing with no-match and ambiguity fail-closed behavior.
12. `WFLOW-008` — implement run-local ephemeral workflow materialization and trigger support for run-local workflow files.
13. `WFLOW-009` plus optional KAH `DAGSM-006` — implement explicit promotion from run-local workflows to project-local persistent workflows when proposal/apply support is approved.

Implementation must not skip ahead from KAS trigger work to custom workflow generation before KAH node-state enforcement exists. WFLOW-005 and DAGSM-004 are documentation/SOT alignment tasks only; they do not claim installed-runtime support, binary update, bundle implementation, classification automation, or project-local persistence.

## Non-goals and deferrals

- No automatic KAH/KAS binary update.
- No install-time bundle persistence or silent `.kkachi-workflow.yaml` / `.kkachi/workflow-catalog.yaml` write.
- No automatic graph/catalog apply from cron/CI.
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
- For bundle-routed/run-local task-DAG workflow support, KAS records the selected bundle, task class, classification reason, skipped-phase reasons, and run-local/persistent posture before workflow execution; for explicit workflow mode, KAS records the explicit workflow id/source and no-bundle reason instead.
- KAS trigger and selector behavior has deterministic tests for explicit, unique, zero-match, and ambiguous workflow selection.
- KAS node contracts preserve agent/role/backend boundaries and never complete nodes without KAH evidence.
- Custom workflow creator defaults to dry-run, emits approval-hash-bound plans, and refuses direct `.kkachi-workflow.yaml` edit fallback.
- Red/Orange/Gray review gates have no unresolved blocking findings for policy, operator UX, and evidence trace.
