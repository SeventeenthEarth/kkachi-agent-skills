---
name: kkachi-workflow-trigger
description: Render dispatch packets for an explicit KAH task-DAG workflow id using an explicit JSON node-contract source.
version: 0.1.0
---

# Kkachi Workflow Trigger

Use this skill only when an operator explicitly provides a `workflow_id` and an explicit node-contract source/ref for a Kkachi task-DAG workflow. This is the WFLOW-002 generic trigger surface. It is not a selector, registry, custom workflow creator, KAH state editor, KAB graph authority, or backend fallback layer.

## CLI

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

Rules:

- `--workflow-id` is required and resolves deterministically to `.kkachi/workflows/<workflow-id>.yaml`.
- `--node-contract-source` is required and must be JSON for the WFLOW-002 MVP.
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

WFLOW-002 accepts only this JSON bundle shape:

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

## Behavior

The trigger:

1. validates/explains `.kkachi/workflows/<workflow-id>.yaml` through KAH;
2. creates a KAH workflow instance with `workflow create --run <run-id> --file <path>` or resumes with `workflow show --run <instance-id>`;
3. reads ready nodes with `workflow ready --run <id>`;
4. renders one dispatch packet per ready node;
5. returns `ok:true/status:no_ready_nodes` when KAH has no ready nodes;
6. always reports `direct_kah_state_write:false`.

Dispatch packets are instructions for the declared lane only. This skill does not start, complete, block, retry, or roll back KAH nodes.

## Deferrals

WFLOW-003 owns selector search and the full node-contract registry. WFLOW-004 owns custom workflow creation and thin trigger scaffolding. This skill must not implement selector search, registry discovery, dynamic node generation, arbitrary webhooks, retry/rollback automation, hidden backend or agent fallback, KAB graph authority, direct KAH state mutation, or Hermes profile/provider/gateway/auth/token/model mutation.
