---
name: kkachi-phase-state
description: Define and enforce the current KAH run, artifact, event, gate, schema, diagnostics, lock, phase-plan, and approval command sequence for Kkachi phases.
version: 0.1.0
---

# Kkachi Phase State

Use this skill whenever Hermes Agent needs to start or advance a Kkachi run using the implemented kkachi-agent-helper CLI.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

KAH owns deterministic state. KHS must use the installed KAH command surface instead of inventing parallel KAH state. Install or update KAH with `go install github.com/SeventeenthEarth/kkachi-agent-helper@latest`. KAH has no dedicated `phase start` command, but KAH does provide the managed `phase-plan init/show/set/validate` surface for KHS-declared phase state. Phase evidence is represented through run lifecycle commands, KAH-managed phase-plan state, canonical artifact files, `event append`, approval records when required, and gate checks.

`phase-plan.yaml` is KHS-declared run-local execution state stored at `.kkachi/runs/<run_id>/phase-plan.yaml` through `phase-plan init/show/set/validate`. `.kkachi-workflow.yaml` is project-level graph state only when the same effective KAH binary proves `kkachi-agent-helper graph` support and validates/explains or applies it. KAH validates structure, skipped/not-applicable reasons, final evidence links, feedback bounds, approval records for rows marked `approval_required: true`, and graph state when the graph surface is used; KHS still owns phase applicability, phase order, and workflow policy decisions.

## Implemented KAH commands

```bash
kkachi-agent-helper project init ... [--force] [--json]
kkachi-agent-helper project status [--json]
kkachi-agent-helper project doctor [--json]

kkachi-agent-helper run create --title <title> --work-path <A_development_execution|B_discovery_shaping> --work-mode <standard|light> --urgency <normal|urgent|critical> --sot-policy <existing_sot_basis|minimal_sot_before_code|full_sot_before_code> --execution-mode <production_write|adapter_qa|readiness_hardening|research|verification|docs_only> --commander <profile> [--backend-evidence <auto|required|not_applicable>] [--task-id <id>] [--redteam <profile>] [--json]
kkachi-agent-helper run list [--json]
kkachi-agent-helper run activate <run_id-or-prefix> [--json]
kkachi-agent-helper run close <run_id-or-prefix> [--json]
kkachi-agent-helper run abort <run_id-or-prefix> [--json]
kkachi-agent-helper run show <run_id-or-prefix> [--json]

kkachi-agent-helper artifact init <run_id-or-prefix> [--json]
kkachi-agent-helper artifact list <run_id-or-prefix> [--json]
kkachi-agent-helper artifact validate <run_id-or-prefix> [--gate intake] [--json]
kkachi-agent-helper artifact write <run_id-or-prefix> <artifact_path> --from <repo-relative-file> [--json]
kkachi-agent-helper artifact append <run_id-or-prefix> <artifact_path> --from <repo-relative-file> [--json]
kkachi-agent-helper artifact set-status <run_id-or-prefix> <artifact_path> --status <pending|complete|not_applicable> [--reason <text>] [--json]

kkachi-agent-helper gate check <run_id-or-prefix> <intake|sot|roadmap|plan|backend|implementation|review|verification|docs|final> [--json]
kkachi-agent-helper gate final <run_id-or-prefix> [--json]

kkachi-agent-helper event append <event_type> --run <run_id-or-prefix> --payload '<json-object>' [--json]
kkachi-agent-helper schema validate <file> --schema <config|status|event|run-metadata|selected-cli|bridge-session-snapshot> [--json]
kkachi-agent-helper schema export [--schema <name>|--all] [--dry-run] [--json]
kkachi-agent-helper schema migrate --from <version> --to <version> [--dry-run] [--json]
kkachi-agent-helper lock recover <active-run|project-write|all> --reason <text> [--run <run_id-or-prefix>] [--json]
kkachi-agent-helper diagnostics export [--run <run_id-or-prefix>] [--output <repo-relative-path>] [--json]

kkachi-agent-helper phase-plan init <run_id-or-prefix> [--json]
kkachi-agent-helper phase-plan show <run_id-or-prefix> [--json]
kkachi-agent-helper phase-plan set <run_id-or-prefix> <phase-id> --status <pending|in_progress|complete|skipped|not_applicable|blocked> [--evidence <path>] [--reason <text>] [--approval-required true|false] [--json]
kkachi-agent-helper phase-plan validate <run_id-or-prefix> [--final] [--json]

kkachi-agent-helper graph init --from-template <template-id-or-path> [--output .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph validate [--file .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph explain [--file .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph diff --from <repo-relative-graph> --to <repo-relative-graph> [--semantic] [--json]
kkachi-agent-helper graph propose --candidate-file <repo-relative-candidate-graph> --reason <text> [--json]
kkachi-agent-helper graph apply --proposal <proposal-id> --approval <evidence-ref> [--json]
kkachi-agent-helper graph export --format mermaid|plantuml [--output <path>] [--json]

kkachi-agent-helper approval request <run_id-or-prefix> --phase <phase-id> --reason <reason> [--evidence <ref>] [--json]
kkachi-agent-helper approval record <run_id-or-prefix> --phase <phase-id> --decision <approved|rejected> --by <approver> --evidence <ref> [--reason <reason>] [--json]
kkachi-agent-helper approval show <run_id-or-prefix> [--phase <phase-id>] [--json]
```

KAH mutating commands fail closed when `.kkachi/status.json.last_event_id` disagrees with the tail of `.kkachi/events.jsonl`.

`artifact set-status` is only for lifecycle/status artifacts whose status field uses KAH artifact lifecycle values (`pending`, `complete`, `not_applicable`), such as markdown checklist-style artifacts. Do not apply `artifact set-status complete` blindly across canonical artifacts. Schema-owned backend JSON artifacts keep their own status vocabularies; for example, `selected-cli.json.status=supported|degraded` and is validated by the `selected-cli` schema plus the backend gate. KAH v0.1.2 guards schema-owned backend JSON status updates fail-closed; KHS operators must still avoid `artifact set-status` on `selected-cli.json` and similar schema-owned backend JSON evidence.

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
2. Create the run with `run create`; set `--backend-evidence required` only when KHS has selected backend evidence as required, then activate it with `run activate`.
3. Create canonical baseline artifacts with `artifact init <run_id>`.
4. Initialize or inspect KHS-declared phase state with `phase-plan init <run_id>` or `phase-plan show <run_id>`. Project bootstrap alone is not a state-machine injection; a project with only `project init` is KAH-ready, but the KAS phase state machine is not active until a run exists and `phase-plan init` has created the run-local phase plan.
5. Update declared phase rows with `phase-plan set <run_id> <phase-id> ...`; use `phase-plan validate <run_id>` during the run and `phase-plan validate <run_id> --final` before final reporting.
5a. Use `.kkachi-workflow.yaml` only after graph capability preflight passes for the effective KAH binary. If `capabilities --json`, `graph --help`, or the required `graph validate/explain/init` command is missing or stale, record a gap and continue only with run-local `phase-plan.yaml` evidence; never repair graph state through manual `.kkachi-workflow.yaml` edits.
5b. When graph state affects the run through init, validate, explain, diff, propose, or apply, write `graph-evidence.md` from `templates/run-artifacts/graph-evidence.md.tmpl`. Preserve template id/path/version, proposal id/path, semantic diff output path, validation/explain report paths, approval/audit evidence, graph checksum/version, KAH graph audit event ids, and capability-check evidence. Missing graph support is also recorded there as a gap.
6. Populate canonical artifact files in `.kkachi/runs/<run_id>/`; prefer `artifact write` and `artifact append` when available so KAH records path-safety checks, atomic mutation, and audit events. Use `artifact set-status` only for artifacts whose status field is KAH lifecycle-owned; never use it as a blanket completion step for schema-owned backend JSON artifacts such as `selected-cli.json`.
7. Use `event append <type> --run <run_id> --payload '<json-object>'` for compact phase milestones such as `phase.started`, `phase.completed`, `artifact.updated`, or `kab.prompt.sent`.
8. Use `approval request`, `approval record`, and `approval show` when KHS has declared that a high-risk phase needs approval; KAH records the approval state but KHS decides when approval is required.
9. Use `schema validate` for `selected-cli.json` and `bridge-session-snapshot.json` when those artifacts are present; use `schema export` or `schema migrate` only for explicit schema maintenance work.
10. Use `artifact validate <run_id> --gate intake` for intake validation.
11. Use `gate check <run_id> <gate>` after the evidence for each implemented gate is complete.
12. Use `gate final <run_id>` before the final report.
13. Use `lock recover` only for explicit stale-lock recovery with a durable reason.
14. Close successful runs with `run close`; abort failed or abandoned runs with `run abort`.

## Outputs

- exact KAH command sequence
- phase/gate verdict summary
- `graph-evidence.md` when graph state affects the run, or when graph capability is missing for requested graph-managed workflow
- blocker reason when KAH fails closed or a required command exits non-zero

## Gate

PASS when KAH state, phase-plan validation, approval records when declared, gate reports, and run artifacts agree. FAIL when artifacts claim completion but `phase-plan validate --final`, `gate check`, or `gate final` fails. BLOCKED when KAH fails closed due to status/event mismatch, lock conflict, unsafe path, stale schema state, or missing repository root.
