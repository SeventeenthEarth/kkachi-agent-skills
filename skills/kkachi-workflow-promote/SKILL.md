---
name: kkachi-workflow-promote
description: Propose explicit promotion from a WFLOW-008 run-local workflow bundle to project-local workflow artifacts.
version: 0.1.0
---

# kkachi-workflow-promote

Use this skill only when an operator explicitly asks to promote an existing
WFLOW-008 run-local workflow for repeated project-local reuse.

Required source bundle:

- `.kkachi/runs/<run_id>/workflow/materialization.json`
- `.kkachi/runs/<run_id>/workflow/workflow.yaml`
- `.kkachi/runs/<run_id>/workflow/node-contracts.json`
- `.kkachi/runs/<run_id>/workflow/checksums.json`
- source/checksum evidence recorded by WFLOW-008 materialization

Dry-run proposal:

```bash
kkachi-agent-skills workflow-promote \
  --project <path> \
  --run <run-id> \
  --target-workflow-id <id> \
  --reuse-reason <reason> \
  [--thin-trigger] \
  --dry-run \
  --json
```

Approval-gated apply check:

```bash
kkachi-agent-skills workflow-promote \
  --project <path> \
  --run <run-id> \
  --target-workflow-id <id> \
  --reuse-reason <reason> \
  [--thin-trigger] \
  --apply dry-run:sha256:<hash> \
  --json
```

`--dry-run` and `--apply dry-run:sha256:<hash>` are mutually exclusive.
`--target-workflow-id` and `--reuse-reason` are always required.

The dry-run emits a stable `kas-workflow-promote-packet/v1` machine packet with:

- source run/materialization provenance and source checksums;
- target project-local candidate paths;
- generated workflow, catalog, node-contract registry, and optional thin trigger
  content;
- base checksums, changed paths, diagnostics/conflicts, no-write evidence, and
  KAS/KAH capability evidence;
- deterministic approval hash bound to the source bundle, target paths,
  generated content, trigger plan, capability evidence, base checksums,
  diagnostics/conflicts, and no-write evidence.

Apply recomputes the approval hash before any apply decision. While reviewed
DAGSM-006 workflow catalog proposal/apply support is absent, correct-hash apply
must fail closed with no writes. KAS must not direct-write
`.kkachi/workflows/*`, `.kkachi/workflow-catalog.yaml`,
`.kkachi-workflow.yaml`, profile files, KAH state, KAB state, auth/token,
provider/gateway/model config, or fallback backend selection.

Generated evidence must preserve:

- `direct_kah_state_write:false`
- `completion_authority:kah_only`
- `fallback_policy:none_fail_closed`

`workflow-promote` is explicit and must not be folded into `workflow-trigger`.
