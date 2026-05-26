# KAS graph template registry SOT

Date: 2026-05-26
Owner: KAS workflow/policy layer
Status: current registry schema SOT for GRAPHMVP-001; template instances remain integration-pending until GRAPHMVP-002+
Authority level: KAS graph template registry schema and validation expectation record
Scope: KAS docs/registry metadata only; no KAH code, KAB runtime, profile config, gateway, or direct `.kkachi-workflow.yaml` mutation
Related docs: `workflow-graph-integration.md`, `phase-orchestration-policy.md`, `external-feedback-intake-policy.md`, `../roadmap.md`

## Purpose

The graph template registry lets KAS name, version, review, and select workflow graph templates before KAH deterministically initializes, validates, explains, proposes, applies, or audits graph state. KAS owns template policy and metadata. KAH owns graph-file validation, write/apply mechanics, checksum/version evidence, semantic diff, and audit events.

GRAPHMVP-001 defines the registry schema only. GRAPHMVP-002 may add the first default workflow graph template. GRAPHMVP-003 may wire capability-checked guidance into KAS orchestration. GRAPHMVP-004 may expand run artifact/report preservation.

## File layout

| Path | Meaning | Status |
|---|---|---|
| `registries/graph-template-registry.yaml` | Registry index and schema rules for graph templates | current schema registry |
| `templates/workflow-graphs/<template-id>.yaml` | KAS-owned workflow graph template files named by registry entries | future template instances |
| `docs/examples/graph-template-registry-valid.yaml` | Positive metadata example for schema/readback checks | example only |
| `docs/examples/graph-template-registry-invalid.yaml` | Negative metadata example documenting expected rejections | example only |
| `.kkachi-workflow.yaml` | Project graph instance after KAH init/validate/apply evidence | not created by GRAPHMVP-001 |

KAS must never silently direct-edit `.kkachi-workflow.yaml` as fallback. Registry entries are allowed as template selection/proposal inputs only.

## Registry object

`registries/graph-template-registry.yaml` has this top-level shape:

```yaml
registry_version: 0.1.0
kind: kas_graph_template_registry
status: schema_only
schema:
  id_pattern: "^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$"
  template_path_pattern: "^templates/workflow-graphs/[a-z][a-z0-9]*(?:-[a-z0-9]+)*\\.yaml$"
  version_pattern: "^v?[0-9]+\\.[0-9]+\\.[0-9]+$"
templates: []
```

`templates` remains empty until a later task adds accepted template instances. Example metadata in `docs/examples/` is not active registry content.

## Template entry schema

Each future `templates[]` entry must provide these fields:

| Field | Required | Rule |
|---|---:|---|
| `id` | yes | kebab-case id matching `schema.id_pattern`; stable across versions |
| `path` | yes | repo-relative `templates/workflow-graphs/<id>.yaml`; must not escape or point into `.kkachi/` |
| `version` | yes | semver string matching `schema.version_pattern` |
| `status` | yes | `candidate`, `implemented`, `deprecated`, or `unsupported` |
| `owner.team` | yes | owning Kkachi/KAS lane, e.g. `blue` |
| `owner.role` | yes | responsible role, e.g. `hwangchung` |
| `owner.reviewers` | yes | review lanes required before activation |
| `summary` | yes | one-line operator-readable purpose |
| `compatibility.kah_min_version` | yes | minimum effective `kkachi-agent-helper` version |
| `compatibility.required_capabilities` | yes | exact KAH capability/help surfaces required by the template entry |
| `phases` | yes | phase ids and applicability expectations |
| `edges` | yes | directed dependencies between phases |
| `gates` | yes | gate ids, blocking behavior, and evidence requirements |
| `approvals` | yes | role approvals required for mutations or risky transitions |
| `feedback_intake` | yes | external feedback bounds and round semantics |
| `validation_expectations` | yes | KAH validation/explain/diff expectations and fail-closed cases |
| `evidence_contract` | yes | report/run artifact fields that must preserve graph evidence |

## Id, path, and version rules

- `id` uses lowercase kebab-case only. No spaces, underscores, slashes, dots, or uppercase letters.
- `path` must equal `templates/workflow-graphs/<id>.yaml` unless a future reviewed schema revision explicitly expands the path rule.
- Template paths are repo-relative and must not symlink-escape the repo.
- `version` is semver. A leading `v` is accepted only for human labels; machine comparisons should normalize it away.
- Registry version changes are schema changes and require review before use.

## Phase and edge rules

`phases[]` entries must include:

```yaml
- id: plan
  applicability: required
  owner_role: planner_backend
  evidence_artifact: plan.md
```

Allowed `applicability` values are `required`, `conditional`, `optional`, and `not_applicable`. Required phases must have a gate or evidence artifact. Conditional phases must name the condition. Phase ids should align with `registries/phase-contracts.yaml` unless the registry entry explicitly marks a compatibility exception.

`edges[]` entries must include `from`, `to`, and `type`. Allowed edge types are `sequence`, `conditional`, and `feedback_loop`. Every edge endpoint must reference a declared phase id. Cycles are forbidden except explicit feedback loops with a bounded maximum round count.

## Gate and approval rules

`gates[]` entries must identify:

- `id`;
- `phase` or `scope`;
- `blocking: true|false`;
- `evidence_required` list;
- responsible role or authority.

`approvals[]` entries must name role/authority, trigger condition, and required evidence. KAS must not encode imperative KAH policy commands such as `gate set`, `review-policy set`, or `graph set-policy` in a template entry.

## Feedback-intake bounds

Every template entry must include an `EXTERNAL_FEEDBACK_INTAKE` policy block compatible with `external-feedback-intake-policy.md`:

```yaml
feedback_intake:
  policy_key: EXTERNAL_FEEDBACK_INTAKE
  min_rounds: 1
  max_rounds: 5
  round_1_required: true
  optional_continuation_rounds: [2, 3, 4, 5]
  five_rounds_mandatory: false
```

Bounds below round 1, max below min, fixed stale `1..3` assumptions, unknown fields, missing round-1-required semantics, or mandatory rounds 2..5 must fail closed.

## KAH validation expectations

A registry entry is eligible for graph use only when the effective KAH binary proves the required graph surface through capability/help evidence. Registry-wide command expectations define the complete surface KAS plans around; a template entry's `compatibility.required_capabilities` must either list the matching full capability set or explicitly justify any narrower, read-only subset in a future reviewed schema revision. GRAPHMVP-001 examples use the full set so GRAPHMVP-002/003 implementers do not confuse action-specific command use with registry eligibility.

Minimum expected command surfaces:

```text
kkachi-agent-helper graph init --from-template
kkachi-agent-helper graph validate
kkachi-agent-helper graph explain
kkachi-agent-helper graph diff
kkachi-agent-helper graph propose
kkachi-agent-helper graph apply
kkachi-agent-helper graph export
```

KAH validates graph instances, not this KAS registry by default. Until KAH or KAS adds a deterministic registry validator, GRAPHMVP-001 verification is docs/schema/readback plus YAML parse and example inspection. KAS guidance must fail closed rather than use a registry entry when the effective KAH binary lacks required capabilities, the template file is missing, the template graph fails KAH validation, or graph/runtime/run-local evidence conflicts.

## Evidence contract

When a graph template affects a run, reports and artifacts must preserve at least:

- registry version;
- template id;
- template path;
- template version;
- owner/reviewer metadata;
- compatibility/capability evidence;
- KAH validation report path;
- KAH explain or diff output path when used;
- proposal id/path when a graph change is proposed;
- approval evidence ref when a graph change is applied;
- applied graph checksum/version and KAH audit event ids;
- conflict/fail-closed reason if graph use is refused.

GRAPHMVP-004 owns the final artifact/report mapping. This document defines the required fields so later tasks do not invent incompatible evidence names.

## Negative cases

KAS must reject or mark unsupported:

- duplicate template ids;
- invalid ids or path escapes;
- path/id mismatch;
- missing owner or reviewer metadata;
- unsupported or stale KAH capability requirements;
- required phases not declared in `phases[]`;
- edges pointing at undeclared phases;
- unbounded feedback loops;
- stale fixed-three feedback semantics;
- direct YAML fallback instructions;
- claims that `kah graph` alias is implemented without separate alias evidence;
- claims that KAB runtime state is graph policy authority.

## Next record action

GRAPHMVP-002 may add a default KAS workflow graph template using this registry schema and verify it with capability-checked `kkachi-agent-helper graph init --from-template`, `graph validate`, and `graph explain` in a temporary repo.
