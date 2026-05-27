# KAS workflow graph integration SOT

Date: 2026-05-21
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record
Status: current KAS/KAH graph integration SOT; KAH 0.1.4 `kkachi-agent-helper graph` implementation evidence present; KAS default template and orchestration guidance are implemented through GRAPHMVP-003; `kah graph` alias remains candidate/unimplemented
Authority level: current KAS/KAH graph integration SOT for capability-checked `kkachi-agent-helper graph` use
Scope: KAS docs only; no KAH code, KAB docs, runtime configs, profiles, registries, or gateway changes
Related docs: `phase-orchestration-policy.md`, `interface-contract.md`, `../roadmap.md`, KAH `docs/specs.md`, KAH `docs/compatibility.md`
Evidence/source paths:
- Governance evidence record in kanban task `t_2fb00394`
- KAH graph implementation evidence from kanban tasks `t_38cfc496` and `t_2b460665`: `kkachi-agent-helper graph --help` and `capabilities --json` advertise graph init/validate/explain/diff/propose/apply/export plus workflow graph compatibility flags.

## Decision summary

KAS owns workflow policy, graph templates, phase applicability, proposal content, and evidence preservation requirements. KAH owns deterministic validation, write/apply, semantic diff, audit events, source precedence, and fail-closed graph state. KAB remains backend/session/plan evidence and is not graph policy authority.

`.kkachi-workflow.yaml` is the canonical project workflow graph file when initialized, validated, or applied through capability-checked KAH graph evidence. `.kkachi/runs/<run_id>/phase-plan.yaml` remains run-local execution state/evidence for a KAS run, derived from or checked against the project graph and KAS phase policy when graph-managed workflow is selected.

## Layer ownership

| Layer | Owns | Must not own |
|---|---|---|
| KAS | graph templates, policy selection, phase applicability/order, declarative proposal content, run evidence requirements | deterministic graph apply/audit mechanics, manual graph mutation |
| KAH | validation, explanation, semantic diff, proposal record, approved graph apply, checksum/version, audit events, fail-closed source precedence | workflow policy decisions, gate/review policy choice, phase applicability |
| KAB | backend sessions, plan lifecycle, prompt dispatch, retained bridge evidence | graph policy authority or project graph source of truth |

## File authority table

| Path / artifact | Meaning | Owner | Authority |
|---|---|---|---|
| `.kkachi-workflow.yaml` | Project-level workflow graph instance | KAS proposes policy/templates; KAH validates/writes/applies | Project graph SOT when backed by KAH init/validation/apply evidence; never a KAS direct-edit fallback |
| `.kkachi/config.yaml` | KAH helper runtime/configuration | KAH | Helper config only; never workflow graph SOT |
| `.kkachi/` | Runtime state, evidence, events, locks, schemas, run artifacts | KAH | Runtime/evidence substrate |
| `.kkachi/runs/<run_id>/phase-plan.yaml` | Run-local execution state/evidence for a KAS run | KAS content stored/validated by KAH | Run-local workflow/execution state; not project graph replacement |
| `.kkachi/config/workflows/` | Kkachi v2 workflow runtime config if present | Kkachi v2, not KAH/KAS graph docs | Out of KAH/KAS graph scope; no merge/fallback |
| Mermaid/PlantUML exports | Generated visualization | KAH export command | Non-authoritative artifact only |

## Required KAS behavior

- KAS chooses a source template or drafts declarative graph proposal content.
- KAS first checks effective KAH capabilities for graph support. If graph support is missing from the effective binary, KAS records a roadmap/feedback gap instead of pretending the command exists.
- KAS uses `kkachi-agent-helper graph` commands only when available and capability-checked. `kah graph` remains shorthand/candidate alias text until a real alias is separately advertised.
- KAS refuses manual `.kkachi-workflow.yaml` repair. Direct edits are unmanaged input until KAH validates or repairs them through proposal/apply evidence.
- KAS uses `init --from-template`, not `init --profile`.
- KAS must not ask KAH to execute imperative policy mutations such as `gate set`, `review-policy set`, or `graph set-policy`.
- KAS preserves graph evidence in run artifacts when graph changes affect a run.

## Orchestration capability preflight

Before any KAS orchestration step uses a project workflow graph, the operator or KAS skill must capture effective-binary evidence from the same environment that will execute the run:

1. Run `kkachi-agent-helper --version`.
2. Run `kkachi-agent-helper capabilities --json` and verify workflow-graph capability flags or equivalent graph support fields are present.
3. Run `kkachi-agent-helper graph --help` and verify the exact subcommands required for the intended action.
4. For template initialization, also verify the selected registry entry and template path, then run `kkachi-agent-helper graph init --from-template <template-id-or-path> --json` only in the target or isolated temp project where graph creation is intended.
5. For existing graph state, run `kkachi-agent-helper graph validate --file .kkachi-workflow.yaml --json` and `kkachi-agent-helper graph explain --file .kkachi-workflow.yaml --json` before deriving run-local phase state from it.

Fail closed and record a gap when any evidence is missing, stale, or unsupported. The gap record must include the attempted command, exit code or missing field, expected capability, affected roadmap/task id, and the reason KAS refused graph-managed workflow. Safe fallback is run-local `phase-plan.yaml` evidence only; it is not permission to write or repair `.kkachi-workflow.yaml` manually.

### Missing-capability gap example

```yaml
graph_guidance:
  status: blocked_by_missing_kah_graph_capability
  attempted_command: kkachi-agent-helper graph --help
  expected_capabilities:
    - kkachi-agent-helper graph init --from-template
    - kkachi-agent-helper graph validate
    - kkachi-agent-helper graph explain
  fallback_allowed: run_local_phase_plan_only
  forbidden_fallbacks:
    - kah graph
    - direct .kkachi-workflow.yaml edit
  gap_record: docs/roadmap.md#GRAPHMVP-or-follow-up
```

## Capability-checked command use

Status: implemented for the real `kkachi-agent-helper graph` command surface when the effective KAH binary advertises it through capabilities/help and command-exit evidence. `kah graph` is not current implementation authority unless separate alias evidence exists.

```text
kkachi-agent-helper graph init --from-template <template-id-or-path> [--output .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph validate [--file .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph explain [--file .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph diff --from <file-or-ref> --to <file-or-ref> [--semantic] [--json]
kkachi-agent-helper graph propose --candidate-file <repo-relative-candidate-graph> --reason <text> [--json]
kkachi-agent-helper graph propose --patch <repo-relative-candidate-graph> --reason <text> [--json]  # legacy compatibility alias for --candidate-file; not a partial patch DSL
kkachi-agent-helper graph apply --proposal <proposal-id> --approval <evidence-ref> [--json]
kkachi-agent-helper graph export --format mermaid|plantuml [--output <path>] [--json]
```

Forbidden as normal KAS guidance:

```text
kkachi-agent-helper workflow ...
kkachi-agent-helper graph init --profile ...
kkachi-agent-helper gate set ...
kkachi-agent-helper review-policy set ...
kkachi-agent-helper graph set-policy ...
```

## Command classification

| Command | Mutates graph? | Category | Policy mutation? |
|---|---:|---|---:|
| `init --from-template` | yes, only initial graph write or approved replacement | deterministic write from selected template | no |
| `validate` | no | validation | no |
| `explain` | no | operator-readable explanation | no |
| `diff` | no | semantic diff | no |
| `propose` | records proposal, does not apply graph | proposal record | no |
| `apply` | yes, after approval/audit evidence | approval-gated deterministic apply | no |
| `export` | no graph mutation | visualization artifact generation | no |

Policy mutation category is empty. KAS and responsible approvers own policy decisions; KAH validates and records state.

## Source precedence and fail-closed rules

### Graph mutation input precedence

1. Explicit `kkachi-agent-helper graph apply --proposal <id>` with approval/audit evidence.
2. Explicit `kkachi-agent-helper graph init --from-template <template-id-or-path>` only when no graph exists, or when replacing through an approved proposal.
3. Explicit `kkachi-agent-helper graph validate/explain/diff --file <path>` for inspection only; inspection does not make the file authoritative.
4. Current `.kkachi-workflow.yaml` on disk only when schema-valid and not in conflict with last KAH audit/checksum evidence.
5. KAS template defaults are proposal/init inputs only, never silent overrides.
6. KAH built-in examples, if any, are examples only and not operational fallback defaults.

### Effective runtime/evidence precedence for graph-managed runs

1. Applied `.kkachi-workflow.yaml` whose checksum/version matches KAH graph audit evidence.
2. KAH graph proposal/apply/audit records proving how the graph changed.
3. Run-local `phase-plan.yaml` for execution state of a specific run.
4. `.kkachi/config.yaml` for helper config only.
5. Generated Mermaid/PlantUML diagrams for visualization only.

Fail closed when:

- graph-managed workflow is required but `.kkachi-workflow.yaml` is missing;
- `.kkachi-workflow.yaml` is invalid, ambiguous, duplicated, or conflicts with KAS phase policy or run-local `phase-plan.yaml`;
- `.kkachi/config.yaml`, generated diagrams, stale `.kkachi/` state, KAS defaults, or Kkachi v2 `.kkachi/config/workflows/` are used as fallback graph authority;
- direct manual edits lack validation/proposal/apply/audit evidence;
- graph proposal/apply changes gates, approvals, review policy, or dependencies without approval/audit evidence;
- KAS asks for imperative KAH policy-setting commands.

## Evidence preservation requirements

When KAS proposes, initializes, validates, or applies graph state, the run must preserve the mapping in `.kkachi/runs/<run_id>/graph-evidence.md` using `templates/run-artifacts/graph-evidence.md.tmpl`, and the final report must summarize the same evidence under `kah_graph_evidence`.

Canonical mapping:

| Required evidence | Artifact/report field |
|---|---|
| source template id/path/version | `template_id`, `template_path`, `template_version` |
| proposal id/path | `proposal_id`, `proposal_path` |
| semantic diff output | `semantic_diff_output_path` |
| validation report | `validation_report_path` |
| explain report | `explain_report_path` |
| approval evidence | `approval_evidence_ref` |
| audit evidence | `audit_evidence_path` |
| graph checksum/version | `graph_checksum`, `graph_version` |
| KAH graph audit event ids | `kah_graph_audit_event_ids` |
| capability-check evidence | `capability_check_evidence` |

`capability_check_evidence` must point back to `capability-check.md` and the captured effective-binary `capabilities --json` / `graph --help` evidence. If graph support is missing or stale, preserve the gap in `graph-evidence.md` and continue only with run-local `phase-plan.yaml` evidence.

## Relationship to run-local phase state

- `.kkachi-workflow.yaml` is project-level graph state.
- `phase-plan.yaml` is the run-local KAS workflow/execution state SOT for one run.
- KAS may derive or constrain a run's phase plan from the project graph after capability-checked graph support is selected.
- A conflict between project graph, KAS phase policy, and `phase-plan.yaml` must stop the run until the responsible role confirms the intended authority.

## Kkachi v2 namespace collision

KAS must not treat Kkachi v2 `.kkachi/config/workflows/templates/*.json` or `.kkachi/config/workflows/addons/*.json` as KAH/KAS graph policy. They are out of this graph scope and have no fallback/merge relationship with `.kkachi-workflow.yaml`.

## Mermaid / PlantUML scope

Mermaid and PlantUML outputs are generated visualization artifacts only. They do not become graph policy, schema, or source of truth. If later exported by KAH, the JSON response must include `authoritative: false` and the source checksum.

## Risk review closure coverage

| Required review item | Resolution in this SOT |
|---|---|
| MF-1 | `.kkachi-workflow.yaml` is project-level graph state; `phase-plan.yaml` remains run-local execution state/evidence and is not deprecated. |
| MF-2 | Kkachi v2 `.kkachi/config/workflows/` is outside KAH/KAS graph scope; no fallback, merge, or namespace sharing is implied. |
| MF-3 | `kkachi-agent-helper graph` is implemented when effective KAH capabilities/help prove it; `kah graph` must not be treated as implemented without alias evidence. |
| MF-4 | Mutation input precedence, runtime/evidence precedence, and fail-closed rules are explicit above. |
| MF-5 | Command classification contains zero policy-mutation commands; KAS must not ask KAH for imperative policy-setting graph commands. |

## Stale/conflict markers

- Any older KAS wording that calls `phase-plan.yaml` the entire workflow SOT is narrowed to run-local execution state/evidence.
- Any prior `.kkachi-config.yaml` or `.kkachi-config.json` graph language is superseded by `.kkachi-workflow.yaml` for this integration plan.
- Any use of `kah graph` without alias capability evidence is non-authoritative shorthand. Use `kkachi-agent-helper graph` for the evidenced command surface after effective-binary capability checks.

## Open questions

- KAS graph template registry format, default template content, capability-checked orchestration guidance, and artifact/report mapping are defined by `graph-template-registry.md`, `registries/graph-template-registry.yaml`, `templates/workflow-graphs/kas-default.yaml`, `templates/run-artifacts/graph-evidence.md.tmpl`, this SOT, and the KAS orchestration skills.
- Exact KAH output payload names remain owned by KAH, but KAS artifacts must normalize proposal paths, event ids, and checksum/version evidence into the canonical fields above.
- KAB alignment with applied graph version is future/non-authoritative until a separate KAB docs update is assigned.

## Record state

GRAPHMVP-004 records artifact/report evidence fields only. Later graph work must stay scoped to `docs/roadmap.md`; do not add graph apply automation, assume `kah graph` alias support, or make KAB graph-version propagation authoritative without separate evidence.
