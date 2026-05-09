---
name: kkachi-phase-state
description: Define and enforce the current KAH run, artifact, event, gate, schema, diagnostics, and lock command sequence for Kkachi phases.
version: 0.1.0
---

# Kkachi Phase State

Use this skill whenever Hermes Agent needs to start or advance a Kkachi run using the implemented kkachi-agent-helper CLI.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

KAH owns deterministic state. KHS must use the installed KAH command surface instead of inventing parallel KAH state. Install or update KAH with `go install github.com/SeventeenthEarth/kkachi-agent-helper@latest`. KAH has no dedicated `phase start` command and currently no first-class `phase_plan` command surface; phase evidence is represented through run lifecycle commands, canonical artifact files, KHS supplemental artifacts such as `phase-plan.yaml`, `event append`, and `gate check`.

`phase-plan.yaml` is a KHS workflow artifact, not KAH metadata. KHS may store it under `.kkachi/runs/<run_id>/` and update it directly until KAH grows managed phase-plan support.

## Implemented KAH commands

```bash
kkachi-agent-helper project init ... [--force] [--json]
kkachi-agent-helper project status [--json]
kkachi-agent-helper project doctor [--json]

kkachi-agent-helper run create --title <title> --work-path <A_development_execution|B_discovery_shaping> --work-mode <standard|light> --urgency <normal|urgent|critical> --sot-policy <existing_sot_basis|minimal_sot_before_code|full_sot_before_code> --execution-mode <production_write|adapter_qa|readiness_hardening|research|verification|docs_only> --commander <profile> [--task-id <id>] [--redteam <profile>] [--json]
kkachi-agent-helper run activate <run_id-or-prefix> [--json]
kkachi-agent-helper run close <run_id-or-prefix> [--json]
kkachi-agent-helper run abort <run_id-or-prefix> [--json]
kkachi-agent-helper run show <run_id-or-prefix> [--json]

kkachi-agent-helper artifact init <run_id-or-prefix> [--json]
kkachi-agent-helper artifact list <run_id-or-prefix> [--json]
kkachi-agent-helper artifact validate <run_id-or-prefix> [--gate intake] [--json]

kkachi-agent-helper gate check <run_id-or-prefix> <intake|sot|roadmap|plan|backend|implementation|review|verification|docs|final> [--json]
kkachi-agent-helper gate final <run_id-or-prefix> [--json]

kkachi-agent-helper event append <event_type> --run <run_id-or-prefix> --payload '<json-object>' [--json]
kkachi-agent-helper schema validate <file> --schema <config|status|event|run-metadata|selected-cli|bridge-session-snapshot> [--json]
kkachi-agent-helper diagnostics export [--run <run_id-or-prefix>] [--output <repo-relative-path>] [--json]
```

KAH mutating commands fail closed when `.kkachi/status.json.last_event_id` disagrees with the tail of `.kkachi/events.jsonl`.

## Inputs

- run id
- project overlay
- current phase
- current artifact root
- phase gate verdict
- artifacts created or updated
- blocker or failure reason, when present

## Flow

1. Initialize or reconfigure the project with `project init`; use `--force` only for non-destructive reconfiguration.
2. Create the run with `run create`, then activate it with `run activate`.
3. Create canonical baseline artifacts with `artifact init <run_id>`.
4. Create or update KHS supplemental `phase-plan.yaml` from `templates/run-artifacts/phase-plan.yaml.tmpl`; treat it as workflow SOT.
5. Populate the canonical artifact files in `.kkachi/runs/<run_id>/`.
6. Use `event append <type> --run <run_id> --payload '<json-object>'` for compact phase milestones such as `phase.started`, `phase.completed`, `artifact.updated`, or `kab.prompt.sent`.
7. Use `schema validate` for `selected-cli.json` and `bridge-session-snapshot.json` when those artifacts are present.
8. Use `artifact validate <run_id> --gate intake` for intake validation.
9. Use `gate check <run_id> <gate>` after the evidence for each implemented gate is complete.
10. Use `gate final <run_id>` before the final report.
11. Close successful runs with `run close`; abort failed or abandoned runs with `run abort`.

## Outputs

- exact KAH command sequence
- phase/gate verdict summary
- blocker reason when KAH fails closed or a required command exits non-zero

## Gate

PASS when KAH state, gate reports, and run artifacts agree. FAIL when artifacts claim completion but `gate check` or `gate final` fails. BLOCKED when KAH fails closed due to status/event mismatch, lock conflict, unsafe path, or missing repository root.
