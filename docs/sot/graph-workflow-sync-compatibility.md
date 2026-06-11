# KAS graph workflow sync compatibility SOT

Date: 2026-06-11
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record pending
Status: accepted SOT for KAS v0.1.2 graph workflow sync support; GRSYNC-001 compatibility metadata is implemented by `registries/graph-workflow-sync-compatibility.yaml`, while doctor/repair/apply behavior remains planned until GRSYNC-002/003 pass evidence and release gates
Authority level: KAS-side planning authority for KAS/KAH version compatibility, workflow graph supportability checks, proposal-first sync orchestration, and periodic hardening posture
Scope: `kkachi-agent-skills` docs, registries, CLI doctor/repair/update guidance, workflow graph support envelope, run evidence guidance; no KAH deterministic implementation, KAB runtime behavior, Hermes profile mutation, auth/token/provider/gateway/model change, or automatic apply without approval
Related docs: `docs/sot/workflow-graph-integration.md`, `docs/sot/graph-template-registry.md`, `registries/graph-workflow-sync-compatibility.yaml`, `docs/sot/kas-cli-contract.md`, `docs/roadmap.md`, KAH `docs/sot/graph-workflow-sync-diagnostics-and-repair.md`, KAH `docs/compatibility.md`
Evidence/source paths:
- Master direction in 17번째 지구 Discord thread on 2026-06-11: KAS changes more often than KAH, KAS must know the KAH versions it supports, KAS/KAH should periodically check `.kkachi-workflow.yaml` supportability, old KAH should trigger update guidance, latest KAH plus old/broken graph should use KAH repair mechanics, and user-custom graphs should be judged by supportability rather than default-template equality.

## Decision summary

KAS v0.1.2 will own graph workflow sync policy and compatibility decisions. KAS must know which KAH version and capabilities are required for graph workflow sync, inspect the effective project environment, classify `.kkachi-workflow.yaml` supportability, recommend KAH/KAS updates when needed, and orchestrate proposal-first graph repair through KAH when the effective KAH is sufficient.

KAS must not directly edit `.kkachi-workflow.yaml` as a fallback. KAS may generate complete candidate graph content and ask KAH to diff/propose/apply it, but KAH remains the deterministic validator/apply/audit owner. User-custom workflow graphs must be accepted when they remain inside the KAS/KAH supported envelope.

## Release target

- Target KAS release: `kkachi-agent-skills v0.1.2`.
- Target KAH dependency for graph workflow sync: `kkachi-agent-helper v0.1.9`.
- KAS compatibility metadata for graph workflow sync must set:

```yaml
kas_version: 0.1.2
kah:
  min_required: 0.1.9
  recommended: 0.1.9
  tested:
    - 0.1.9
workflow_graph_sync:
  min_required: 0.1.9
```

KAS v0.1.2 must not claim graph workflow sync support until KAH v0.1.9 implementation/release evidence exists.

## Compatibility metadata

GRSYNC-001 implements the KAS-side machine-readable compatibility source at `registries/graph-workflow-sync-compatibility.yaml`. The registry is repository-visible metadata and is validated by docs-contract readback against this SOT, `docs/roadmap.md`, `docs/README.md`, `docs/kkachi-docs-map.yaml`, and `registries/graph-template-registry.yaml`.

The machine-readable compatibility registry must contain:

- KAS version;
- KAH `min_required`, `recommended`, and `tested` versions for graph workflow sync;
- required KAH graph command groups and subcommands;
- required KAH compatibility flags;
- supported workflow graph schema versions;
- supported KAS graph template ids/versions;
- status vocabulary and remediation guidance;
- release evidence notes tying KAS v0.1.2 to KAH v0.1.9.

The metadata should separate general KAS CLI usability from graph workflow sync support if future KAS commands can run with older KAH. For this workstream, graph workflow sync requires KAH 0.1.9.

## Required KAS behavior

### 1. Read-only workflow graph doctor

KAS must provide a read-only project check, likely through `doctor --project <path> --workflow-graph --json` or the approved lifecycle equivalent. The check must gather evidence from the same effective environment that will run the project workflow:

1. `kkachi-agent-skills --version` or equivalent KAS version evidence.
2. KAS compatibility metadata readback.
3. `kkachi-agent-helper --version`.
4. `kkachi-agent-helper capabilities --json`.
5. `kkachi-agent-helper graph --help`.
6. Project `.kkachi-workflow.yaml` presence.
7. KAH `graph validate --json`, `graph explain --json`, and diagnostics when available.
8. Optional run-local `phase-plan.yaml` compatibility checks when a run is in scope.

The doctor must not mutate `.kkachi-workflow.yaml`, `.kkachi/`, Hermes profiles, KAH binaries, KAB state, or project source files.

### 2. Classification vocabulary

KAS should emit compact machine-readable status labels. Initial vocabulary:

- `pass`
- `custom_supported`
- `update_kah_required`
- `update_kah_recommended`
- `update_kas_recommended`
- `graph_missing`
- `graph_stale`
- `graph_broken`
- `graph_conflict`
- `proposal_available`
- `blocked_for_approval`
- `unsupported`

Classification rules:

- If effective KAH is below KAS graph workflow sync `min_required`, KAS recommends KAH update before graph repair.
- If KAH is current enough and graph is old or broken, KAS may offer proposal-first repair.
- If KAH is current enough and the graph is valid/custom but inside the supported envelope, KAS returns `custom_supported` or `pass` without forcing the default template.
- If the graph uses a newer schema/policy than KAS understands, KAS recommends KAS update or fails closed as `unsupported`.
- If KAS/KAH are both current but diagnostics conflict, KAS fails closed and records a bug/spec gap rather than silently repairing.

### 3. Proposal-first sync orchestration

KAS may provide `repair --workflow-graph --propose` or an approved lifecycle equivalent. Proposal generation must:

- run the read-only doctor first;
- refuse to propose when KAH is below the required graph workflow sync version;
- generate only complete candidate graph files, never partial patches;
- preserve user-custom graph intent when the existing graph is supported;
- call KAH `graph diff` and `graph propose` for stale/broken supported repair cases;
- output proposal id/path, semantic diff, risk flags, approval requirement, and next command;
- avoid applying changes by default.

### 4. Approval-gated apply orchestration

KAS may provide `repair --workflow-graph --apply --approval <ref>` or an approved lifecycle equivalent. Apply orchestration must:

- require explicit approval evidence;
- call KAH apply rather than directly writing `.kkachi-workflow.yaml`;
- rerun validate/explain/diagnostics after apply;
- preserve graph evidence in run artifacts when a KAS run is active;
- report backup/recovery/audit references returned by KAH;
- fail closed on KAH drift, proposal mismatch, missing approval, or unsupported graph state.

### 5. Periodic hardening posture

KAS should support safe periodic graph checks through CI, cron, or no-agent runners. The periodic default is read-only doctor/report. Proposal generation is opt-in. Apply is never automatic without explicit approval evidence.

Periodic checks should help detect:

- KAH binary drift below required version;
- KAS compatibility metadata drift;
- `.kkachi-workflow.yaml` schema/source-precedence drift;
- stale feedback-intake bounds;
- unsupported user edits;
- phase-plan/project-graph conflicts.

## User-custom graph rule

A user-custom `.kkachi-workflow.yaml` is acceptable when KAH validates/explains it and KAS can classify it inside the supported envelope. KAS must not force replacement with `kas-default` solely because the graph differs from the default template. KAS may recommend hardening or optimization separately, through proposal-first evidence.

## Non-goals

- No direct `.kkachi-workflow.yaml` manual repair fallback.
- No automatic KAH binary update.
- No automatic graph apply from doctor or periodic checks.
- No KAB graph policy authority.
- No Hermes profile/provider/model/gateway/auth/token mutation.
- No Kkachi v2 `.kkachi/config/workflows/` merge/fallback.
- No claim that KAH decides KAS policy.

## Roadmap slice mapping

This SOT maps to three KAS PR-candidate tasks:

1. `GRSYNC-001` — compatibility registry and graph workflow sync SOT/metadata.
2. `GRSYNC-002` — read-only workflow graph doctor.
3. `GRSYNC-003` — proposal/apply orchestration and periodic check guidance.

These tasks depend on KAH v0.1.9 graph diagnostics and repair substrate for complete support.

## Acceptance gates before KAS v0.1.2 release

- KAS compatibility metadata states KAH 0.1.9 as `min_required`, `recommended`, and `tested` for graph workflow sync.
- Read-only doctor distinguishes KAH outdated, graph missing, graph stale, graph broken, graph conflict, custom-supported, and pass states.
- KAS refuses graph repair before KAH update when effective KAH is below 0.1.9.
- KAS proposal-first repair uses KAH graph diff/propose and does not directly edit `.kkachi-workflow.yaml`.
- KAS apply orchestration is approval-gated and uses KAH graph apply.
- User-custom supported graphs are not overwritten or treated as errors solely because they differ from KAS defaults.
- Periodic check docs default to read-only report; proposal/apply remain opt-in/approval-gated.
- `make test` and relevant CLI/docs-contract/e2e tests pass.
- KAS docs, roadmap, compatibility registry, CLI guidance, and release notes match the implemented behavior.
