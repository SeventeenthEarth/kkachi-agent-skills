---
name: kkachi-workflow-create
description: Plan custom task-DAG workflow candidates with approval-hash-bound dry-run packets.
version: 0.1.0
---

# Kkachi Workflow Create

Use this skill when an operator asks KAS to create custom task-DAG workflow
candidates for a project. This skill plans candidate content only. It is not a
workflow executor, KAH state writer, KAB graph authority, profile installer,
webhook runtime, retry/rollback engine, or fallback selector.

## CLI

Dry-run first:

```bash
kkachi-agent-skills workflow-create \
  --project <path> \
  --workflow-id <id> \
  --mode dag_only|thin_trigger|full_trigger \
  --request <json-path> \
  --dry-run \
  --json
```

Approved apply attempt:

```bash
kkachi-agent-skills workflow-create \
  --project <path> \
  --workflow-id <id> \
  --mode dag_only|thin_trigger|full_trigger \
  --request <json-path> \
  --apply dry-run:sha256:<hash> \
  --json
```

`--request` is JSON-only and must use
`schema_version: kas-workflow-create-request/v1`. It declares selector metadata,
nodes, KAS node-contract fields, required outputs, and optional trigger
metadata. `full_trigger` is exceptional and requires a non-empty
`full_trigger_reason` in the request or `--full-trigger-reason`.

## Modes

`dag_only` emits candidate DAG, catalog, and node-contract registry content for
the generic workflow trigger.

`thin_trigger` adds a small generated `SKILL.md` wrapper that pins the workflow
id and delegates to `kkachi-agent-skills workflow-trigger`.

`full_trigger` adds exceptional custom trigger scaffolding. It must carry the
explicit reason in the packet and still must not add backend fallback, arbitrary
webhooks, retry/rollback automation, dynamic node generation, or direct KAH
state writes.

## Packet And Approval Hash

Dry-run emits compact human output plus a full JSON `machine_packet` containing:

- `candidate_paths`
- `target_paths`
- `generated_content`
- `selector_metadata`
- `node_contracts`
- `trigger_plan`
- KAS/KAH `capability_evidence`
- `base_checksums`
- `changed_paths`
- `conflicts` and `diagnostics`
- `no_write`
- `approval_hash`

The approval schema is `kas-workflow-create-approval/v1`. The approval hash is
`sha256` over deterministic UTF-8 JSON with sorted keys and normalized
project-relative paths. It binds command, mode, target paths, generated content,
selector metadata, KAH/KAS capability-version evidence, base checksums, changed
paths, conflicts/diagnostics, and no-write evidence.

Apply recomputes the packet and refuses malformed, stale, mismatched, blocked,
or non-approvable evidence before any write or KAH call.

## KAH Boundary

KAS generates candidate content and approval packets only. KAH remains
authoritative for workflow/catalog validation, proposal, apply, audit, final
gate integration, and node-contract registry evidence.

Current source-built KAH DAGSM-003 evidence advertises `workflow catalog
validate/explain` and node-contract registry diagnostics. It does not advertise
a reviewed workflow catalog proposal/apply command mapping. KAH v0.1.10 is the first release line for reviewed workflow catalog proposal/apply. Therefore runtime apply
stays fail-closed until the effective helper advertises `workflow_catalog_proposal_apply=true` and the catalog
proposal/apply surface. KAS must not use direct KAH state writes as a fallback.

## Fail-Closed Codes

Stable blockers include `workflow_create_mode_ambiguous`,
`workflow_creator_mode_unsupported`, `selector_metadata_invalid`,
`unsafe_target_path`,
`blocked_missing_kah_workflow_capability`,
`blocked_missing_kah_workflow_catalog_capability`,
`approval_evidence_malformed`, `approval_plan_hash_mismatch`,
`base_catalog_unreadable`, `base_checksum_mismatch`,
`generated_skill_validation_failed`, and `workflow_create_apply_refused`.

Do not resolve these by selecting a fallback workflow, selecting a fallback
agent/backend, auto-loading a default selector registry, direct-writing `.kkachi`
state, or mutating Hermes profiles, providers, auth, tokens, gateways, models,
or KAB runtime state.

The combined fallback agent/backend path is intentionally unsupported.
