---
name: kkachi-workflow-trigger
description: Render dispatch packets for explicit or selector-matched KAH task-DAG workflows.
version: 0.1.0
---

# Kkachi Workflow Trigger

Use this skill when an operator provides either:

- WFLOW-002 explicit mode: a `workflow_id` and explicit JSON node-contract source/ref; or
- WFLOW-003 selector mode: an explicit selector registry plus deterministic selector inputs.

The trigger renders dispatch packets only. It is not a custom workflow creator,
thin trigger generator, KAH state editor, KAB graph authority, dynamic node
builder, backend fallback layer, or agent fallback layer.

## CLI

Explicit WFLOW-002 mode:

```bash
kkachi-agent-skills workflow-trigger \
  --project <path> \
  --workflow-id <id> \
  --node-contract-source <path> \
  [--node-contract-ref <ref>] \
  [--run <run-id>] \
  [--instance-id <id>] \
  --json
```

Selector-registry WFLOW-003 mode:

```bash
kkachi-agent-skills workflow-trigger \
  --project <path> \
  --selector-registry <path> \
  --task-class <class> \
  [--labels <csv>] \
  [--changed-surfaces <csv>] \
  [--risk <level>] \
  [--required-agent <csv>] \
  [--required-capability <csv>] \
  [--run <run-id>] \
  [--instance-id <id>] \
  --json
```

Rules:

- In explicit mode, `--workflow-id` is required and resolves deterministically to `.kkachi/workflows/<workflow-id>.yaml`.
- In explicit mode, `--node-contract-source` is required and must be JSON.
- Explicit mode does not perform selector search or registry matching.
- In selector mode, `--selector-registry` is required. The bundled registry is `registries/task-dag-workflow-registry.yaml`, but the CLI does not auto-load it; callers must pass the path explicitly.
- Selector mode uses deterministic registry predicates. Exactly one workflow may proceed; zero matches return `selector_no_match`, multiple matches return `selector_ambiguous`, and mixed explicit/selector inputs return `selector_explicit_mode_conflict`.
- Use `--run <run-id>` when KAH should create the workflow instance.
- Use `--instance-id <id>` when KAH should resume by `workflow show`; the instance id is passed to KAH as the run id.
- JSON output is the stable contract. Human output is compact status only.

## KAH Preflight

The trigger fails closed unless the effective `kkachi-agent-helper` supports all required workflow surfaces:

- `kkachi-agent-helper --version`
- `kkachi-agent-helper capabilities --json`
- `kkachi-agent-helper workflow --help`
- workflow subcommands: `validate`, `explain`, `create`, `show`, `ready`, and `node`
- capability flags: `task_dag_schema_validation` and `workflow_instance_state`

Missing workflow capability returns `ok:false`, `status: blocked_missing_kah_workflow_capability`, `direct_kah_state_write:false`, and no dispatch packets. The trigger must not call workflow instance commands after a failed capability preflight.

## Node-Contract JSON

Explicit WFLOW-002 mode accepts only this JSON bundle shape:

```json
{
  "schema_version": "kas-node-contracts/v1",
  "ref": "optional-source-ref",
  "contracts": [
    {
      "workflow_id": "demo",
      "node_id": "setup",
      "owner_role": "implementer_backend",
      "execution_lane": "direct_kas_skill",
      "required_inputs": ["task-contract.yaml"],
      "expected_artifacts": ["artifacts/setup.md"],
      "prompt_ref": "skills/kkachi-implement/SKILL.md",
      "approval_required": true,
      "fallback_policy": "none_fail_closed",
      "verification_gate": "make test"
    }
  ]
}
```

Every KAH-ready node must have exactly one matching contract by `workflow_id` and `node_id`. A missing matching contract fails closed with `status: blocked_missing_ready_node_contract` and no partial packets.

## Selector Registry

WFLOW-003 selector mode reads `kas-task-dag-workflow-registry/v1` YAML from the
explicit `--selector-registry` path. Registry entries carry selector metadata
such as task class, labels, changed surfaces, risk, required agents, required
capabilities, workflow id/path, and node contracts. Registry node contracts
share the WFLOW-002 core fields and add the selector-specific `task_class`
invariant.

Selector mode must not rank, score, choose first, use LLM tie-breaks, infer a
backend/agent fallback, or call KAH workflow instance commands for zero or
ambiguous matches.

## Behavior

The trigger:

1. resolves exactly one workflow from explicit inputs or selector-registry inputs;
2. validates/explains the selected workflow through KAH;
3. creates a KAH workflow instance with `workflow create --run <run-id> --file <path>` or resumes with `workflow show --run <instance-id>`;
4. reads ready nodes with `workflow ready --run <id>`;
5. renders one dispatch packet per ready node;
6. returns `ok:true/status:no_ready_nodes` when KAH has no ready nodes;
7. always reports `direct_kah_state_write:false`, `completion_authority:kah_only`, and `stage1_direct_codex_is_kab_native_codex:false`.

Dispatch packets are instructions for the declared lane only. This skill does not start, complete, block, retry, or roll back KAH nodes.

## Deferrals

WFLOW-004 owns custom workflow creation and thin trigger scaffolding. This skill
must not implement custom creator/thin-trigger behavior, dynamic node
generation, arbitrary webhooks, retry/rollback automation, hidden backend or
agent fallback, KAB graph authority, direct KAH state mutation, KAH completion authority, registry auto-loading without `--selector-registry`, or Hermes
profile/provider/gateway/auth/token/model mutation.
