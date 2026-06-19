---
name: kkachi-workflow-trigger
description: Render dispatch packets for explicit or selector-matched KAH task-DAG workflows.
version: 0.1.0
---

# Kkachi Workflow Trigger

Use this skill when an operator provides any of:

- WFLOW-002 explicit mode: a `workflow_id` and explicit JSON node-contract source/ref; or
- WFLOW-003 selector mode: an explicit selector registry plus deterministic selector inputs; or
- WFLOW-008 run-local mode: a `workflow-route` JSON result or approved
  `workflow-create` dry-run packet to materialize under
  `.kkachi/runs/<run_id>/workflow/`.

The trigger renders dispatch packets only. It is not a custom workflow creator,
thin trigger generator, KAH state editor, KAB graph authority, dynamic node
builder, backend fallback layer, or agent fallback layer.

## CLI

Explicit WFLOW-002 mode:

```bash
kkachi-agent-skills workflow-trigger \
  --project <path> \
  --workflow-id <id> \
  [--workflow-file <repo-relative-yaml>] \
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

Run-local WFLOW-008 mode:

```bash
kkachi-agent-skills workflow-trigger \
  --project <path> \
  --route-result <workflow-route-json> \
  --materialize-run-local \
  --run <run-id> \
  --json
```

Classified workflow-managed KAS/KAH mode adds:

```bash
  --workflow-managed
```

Use `--workflow-managed` for STRICT-003 KAS/KAH runs after task classification
selects a workflow. In this mode dispatch must come from a preserved
`workflow-route` result plus run-local materialization, or from an explicit
resume of an already materialized run-local workflow whose
`.kkachi/runs/<run_id>/workflow/materialization.json` proves route-result
materialization. Direct explicit or selector dispatch is rejected before KAH calls.

Approved one-off WFLOW-008 custom packet mode:

```bash
kkachi-agent-skills workflow-trigger \
  --project <path> \
  --run <run-id> \
  --custom-workflow-packet <workflow-create-dry-run.json> \
  --approval dry-run:sha256:<hash> \
  --materialize-run-local \
  --json
```

Rules:

- In explicit mode, `--workflow-id` is required. Without `--workflow-file`, it
  preserves the legacy `.kkachi/workflows/<workflow-id>.yaml` resolution.
- When `--workflow-file` is supplied, the trigger passes that explicit
  repository-relative YAML path to KAH `workflow validate`, `workflow explain`,
  and `workflow create`; it does not derive `.kkachi/workflows/<workflow-id>.yaml`.
- In explicit mode, `--node-contract-source` is required and must be JSON.
- Explicit mode does not perform selector search or registry matching.
- In selector mode, `--selector-registry` is required. The bundled registry is `registries/task-dag-workflow-registry.yaml`, but the CLI does not auto-load it; callers must pass the path explicitly.
- Selector mode uses deterministic registry predicates. Exactly one workflow may proceed; zero matches return `selector_no_match`, multiple matches return `selector_ambiguous`, and mixed explicit/selector inputs return `selector_explicit_mode_conflict`.
- In run-local mode, the route result must already be `ok:true` with
  `status: bundle_route_matched`; `workflow-trigger` does not reroute raw task
  metadata or infer a task class. It produces route-backed run-local materialization evidence before workflow-managed dispatch.
- In approved one-off custom packet mode, the packet must be the existing
  `workflow-create` dry-run machine packet shape, `--approval` must be exactly
  `dry-run:sha256:<hash>`, and the hash is recomputed with WFLOW-004 canonical
  approval semantics before any run-local write.
- `--route-result` and `--custom-workflow-packet` are mutually exclusive.
- Run-local materialization writes only `.kkachi/runs/<run_id>/workflow/`
  artifacts: `materialization.json`, `workflow.yaml`, `node-contracts.json`,
  either `route-result.json` or `custom-workflow-packet.json`, and
  `checksums.json`.
- Run-local materialization is preflighted: if the effective KAH binary lacks
  workflow support, the command fails closed before writing run-local files.
- In workflow-managed mode, missing route-result evidence, selector bypass,
  explicit dispatch bypass, custom one-off packet materialization, missing
  materialization evidence, or mismatched resume inputs fail closed with no
  dispatch packets and no fallback to the legacy development spine.
- Use `--run <run-id>` when KAH should create the workflow instance.
- Use `--instance-id <id>` when KAH should resume by `workflow show`; the instance id is passed to KAH as the run id.
- JSON output is the stable contract. Human output is compact status only.

## KAH Preflight

The trigger fails closed unless the effective `kkachi-agent-helper` supports all required workflow surfaces:

- `kkachi-agent-helper --version`
- `kkachi-agent-helper capabilities --json`
- `kkachi-agent-helper workflow --help`
- workflow subcommands: `validate`, `explain`, `create`, `show`, `ready`, and `node`
- capability flags: `task_dag_schema_validation` and `workflow_instance_state`; in `--workflow-managed` mode also require `workflow_strict_transition_ledger` and `workflow_transition_order_verification`

Missing workflow capability returns `ok:false`,
`status: blocked_missing_kah_workflow_capability` in ordinary mode or
`status: strict_workflow_missing_kah_capability` in workflow-managed mode,
`direct_kah_state_write:false`, and no dispatch packets. Run-local
materialization also performs this preflight before writes, so installed KAH
without the `workflow` command group must not create `.kkachi/runs/.../workflow`
artifacts.

Recovery for `strict_workflow_missing_kah_capability`: verify the effective
repo-selected KAH binary exposes `task_dag_schema_validation`,
`workflow_instance_state`, `workflow_strict_transition_ledger`, and
`workflow_transition_order_verification` before rerunning the same route-backed
workflow-managed trigger. Recovery for
`strict_workflow_expected_start_revision_missing`: inspect KAH `workflow show`
and `workflow ready` JSON for a positive `instance.revision`; repair/recreate
the route-backed workflow instance if revision evidence is absent.

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

1. resolves exactly one workflow from explicit inputs, selector-registry inputs,
   a previously matched route-result file, or an approved custom workflow
   packet;
2. validates/explains the selected workflow through KAH;
3. creates a KAH workflow instance with `workflow create --run <run-id> --file <path>` or resumes with `workflow show --run <instance-id>`;
4. reads ready nodes with `workflow ready --run <id>`;
5. renders one dispatch packet per ready node with `strict_order:true`, run id, instance revision, KAH ready-node reasons, and `expected_start_revision` from the observed instance revision; workflow-managed mode fails closed with `strict_workflow_expected_start_revision_missing` if that revision is unavailable;
6. returns `ok:true/status:no_ready_nodes` when KAH has no ready nodes;
7. always reports `direct_kah_state_write:false`, `completion_authority:kah_only`, `fallback_policy:none_fail_closed`, and `stage1_direct_codex_is_kab_native_codex:false`.

Dispatch packets are instructions for the declared lane only. This skill does not start, complete, block, retry, or roll back KAH nodes.

For classified workflow-managed KAS/KAH runs, step 1 is narrowed by STRICT-003:
the trigger must consume a successful `workflow-route` result
(`ok:true/status:bundle_route_matched`) through `--route-result
--materialize-run-local --run <run-id> --workflow-managed`, or resume an
existing KAH instance only after reading matching route-backed run-local
materialization evidence. It must not substitute selector mode, explicit
`.kkachi/workflows/<workflow-id>.yaml` mode, custom one-off packet mode, skill
order, phase-plan text, or the old default development spine.

Run-local materialization records the selected bundle, task class,
classification reason or approved one-off packet evidence, source
registry/taxonomy checksums or approval hash/packet checksum, run-local
posture, `persistent_promotion:false`, and `no_promotion:true`. It must not write `.kkachi-workflow.yaml`,
`.kkachi/workflow-catalog.yaml`, `.kkachi/workflows/*`, profiles, providers,
gateways, auth, tokens, models, or KAB runtime state.

## Deferrals

WFLOW-004 owns custom workflow creation and thin trigger scaffolding. WFLOW-009
owns persistent workflow promotion/apply. This skill may consume an approved
WFLOW-004 dry-run machine packet only to materialize a one-off run-local
workflow; it must not implement project-local custom creator/thin-trigger
apply, dynamic node generation, arbitrary webhooks, retry/rollback automation,
hidden backend or agent fallback, KAB graph authority, direct KAH state
mutation, KAH completion authority, registry auto-loading without
`--selector-registry`, project-local catalog promotion, or Hermes
profile/provider/gateway/auth/token/model mutation.
