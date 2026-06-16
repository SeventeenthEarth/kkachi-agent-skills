---
name: kkachi-workflow-route
description: Route already-classified Kkachi tasks to one standard bundle without KAH calls.
version: 0.1.0
---

# Kkachi Workflow Route

Use this skill after KHS/Kkachi task classification already exists and the
operator needs the KAS-owned standard bundle route for WFLOW-007.

This skill is route-only. It does not infer task class from raw prose, call KAH
workflow create/show/ready/node APIs, render dispatch packets, materialize
`.kkachi/runs/<run_id>/workflow.yaml`, promote project workflow catalogs, mutate
profiles, or write KAH/KAB/Hermes runtime state.

## CLI

```bash
kkachi-agent-skills workflow-route \
  --taxonomy registries/task-taxonomy.yaml \
  --selector-registry registries/task-dag-workflow-registry.yaml \
  --task-class <class> \
  --classification-reason <reason> \
  [--selected-spine <bundle>] \
  [--labels <csv>] \
  [--changed-surfaces <csv>] \
  [--risk <level>] \
  [--required-agent <csv>] \
  [--required-capability <csv>] \
  --json
```

The input is already-classified metadata. KAS validates the class or accepted
taxonomy alias, reads the standard bundle registry, and may return
`bundle_route_matched` only when exactly one standard bundle matches and any
explicit `--selected-spine` agrees with that deterministic result.

## Fail-Closed Statuses

- `classification_required_input_missing`
- `classification_reason_missing`
- `classification_class_unsupported`
- `taxonomy_required`
- `taxonomy_unreadable`
- `taxonomy_schema_unsupported`
- `bundle_registry_required`
- `bundle_registry_unreadable`
- `bundle_registry_schema_unsupported`
- `bundle_default_spine_missing`
- `bundle_no_match`
- `bundle_ambiguous`
- `bundle_selected_mismatch`

Every failure returns `ok:false`, `direct_kah_state_write:false`, and no
dispatch packets.
