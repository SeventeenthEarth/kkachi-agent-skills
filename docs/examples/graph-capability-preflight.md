# Graph capability preflight examples

These examples define how KAS orchestration guidance should behave before using project-level workflow graph state. They are examples only; the authority is `docs/sot/workflow-graph-integration.md`.

## Passing preflight

```yaml
graph_guidance:
  status: ready
  effective_binary:
    version_command: kkachi-agent-helper --version
    capabilities_command: kkachi-agent-helper capabilities --json
    help_command: kkachi-agent-helper graph --help
  required_commands_present:
    - kkachi-agent-helper graph init --from-template
    - kkachi-agent-helper graph validate
    - kkachi-agent-helper graph explain
  selected_template:
    id: kas-default
    path: templates/workflow-graphs/kas-default.yaml
  allowed_next_steps:
    - kkachi-agent-helper graph init --from-template templates/workflow-graphs/kas-default.yaml --json
    - kkachi-agent-helper graph validate --file .kkachi-workflow.yaml --json
    - kkachi-agent-helper graph explain --file .kkachi-workflow.yaml --json
```

## Missing capability

```yaml
graph_guidance:
  status: blocked_by_missing_kah_graph_capability
  attempted_command: kkachi-agent-helper graph --help
  expected_capability: kkachi-agent-helper graph validate
  action: record_gap
  fallback_allowed: run_local_phase_plan_only
  forbidden_fallbacks:
    - kah graph
    - direct .kkachi-workflow.yaml edit
```

## Stale alias claim

```yaml
graph_guidance:
  status: rejected_stale_alias_claim
  bad_instruction: kah graph validate
  reason: kah graph alias has no separate capability/help evidence
  replacement: kkachi-agent-helper graph validate --file .kkachi-workflow.yaml --json
```

## Direct YAML fallback claim

```yaml
graph_guidance:
  status: rejected_direct_yaml_fallback
  bad_instruction: edit .kkachi-workflow.yaml manually when graph init is unavailable
  reason: direct edits are unmanaged input until KAH validates, proposes, or applies them
  safe_action: record_gap_and_continue_with_run_local_phase_plan_only
```
