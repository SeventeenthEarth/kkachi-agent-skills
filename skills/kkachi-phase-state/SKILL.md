---
name: kkachi-phase-state
description: Define and enforce the current KAH run, artifact, event, gate, schema, diagnostics, lock, phase-plan, and approval command sequence for Kkachi phases.
version: 0.2.0
---

# Kkachi Phase State

Use this skill whenever Hermes Agent needs to start or advance a Kkachi run using the implemented kkachi-agent-helper CLI.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

KAH owns deterministic state. KHS must use the installed KAH command surface instead of inventing parallel KAH state. **Approval boundary:** editing KAH or KAS repositories, configs, workflow templates, installed binaries, or persistent skills that govern KAH/KAS operation requires a prior report to 주군 and explicit approval. Do not infer permission to modify KAH/KAS core from a downstream project request such as “make this workflow enforce X”; first present the proposed KAH/KAS changes, scope, side effects such as `go install`, and ask for approval. Project-local workflow files may be changed only when that is the approved target scope.

Install or update KAH with `go install github.com/SeventeenthEarth/kkachi-agent-helper@latest` only after that approval when the task touches the active/global helper binary. KAH has no dedicated `phase start` command, but KAH does provide the managed `phase-plan init/show/set/validate` surface for KHS-declared phase state. Phase evidence is represented through run lifecycle commands, KAH-managed phase-plan state, canonical artifact files, `event append`, approval records when required, and gate checks.

`phase-plan.yaml` is KHS-declared run-local execution state stored at `.kkachi/runs/<run_id>/phase-plan.yaml` through `phase-plan init/show/set/validate`. `.kkachi-workflow.yaml` is project-level graph state only when the same effective KAH binary proves `kkachi-agent-helper graph` support and validates/explains or applies it. Current KAH (with graph-required phase support) initializes and validates phase-plan rows from baseline KHS phases plus graph-declared `required: true` phases, so project workflows can machine-enforce additional rows. KHS policy requires MAR for `development` / implementation tasks unless 주군 explicitly waives or replaces MAR before start and the decision is recorded in KAH/run evidence artifacts, normally represented as later request-feedback/handle-feedback rows after first color review. For non-implementation durable-change runs, MAR is project/request/high-risk opt-in policy; omit the MAR phases/gates from `.kkachi-workflow.yaml` or declare them optional and mark them `not_applicable` with a reason when not applicable. KAH must not require MAR globally for all task classes.

KHS review policy is broader than implementation-only work: for any active KAS/KAH run that changes durable repository artifacts, workflow state, docs, or release/commit evidence, first Blue + Red/Orange/Gray color review should be represented in the phase plan and evidence artifacts even when the task class is `docs_only`, `research_evidence`, `bootstrap_config`, or `collaboration_review`. Pure read-only explanations and bounded command reports should usually stay outside KAS/KAH; if a run exists and no durable artifact changed, review phases may be `not_applicable` only with a concrete reason. If the task class is `development` / implementation, or MAR is explicitly requested/project-opted-in/high-risk opted-in for another task class, post-MAR feedback handling plus a second Blue + Red/Orange/Gray re-review become mandatory before final/pre-commit reporting.

KAH validates structure, skipped/not-applicable reasons, final evidence links, feedback bounds, approval records for rows marked `approval_required: true`, and graph state when the graph surface is used; KHS still owns phase applicability, phase order, and workflow policy decisions.

## Command surface summary

- project lifecycle: `project init`, `project status`, `project doctor`
- run lifecycle: `run create`, `run list`, `run activate`, `run show`, `run close`, `run abort`
- artifact lifecycle: `artifact init`, `artifact list`, `artifact validate`, `artifact write`, `artifact append`, `artifact set-status`
- phase state: `phase-plan init`, `phase-plan show`, `phase-plan set`, `phase-plan validate`
- graph state: `graph init`, `graph validate`, `graph explain`, `graph diff`, `graph propose`, `graph apply`, `graph export`
- approvals, gates, and audit: `approval request/record/show`, `gate check/final`, `event append`, `schema validate/export/migrate`, `lock recover`, `diagnostics export`

See `references/implemented-kah-command-surface.md` for the full command catalog and argument shapes.

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

Use the canonical KAH run order: initialize project state, create and activate the run, create baseline artifacts, initialize `phase-plan.yaml`, maintain phase rows with `phase-plan set/validate`, preserve `graph-evidence.md` when graph state is involved, update canonical artifacts through KAH, run `gate check`/`gate final`, and close or abort the run explicitly.

See `references/run-state-flow.md` for the exact step-by-step command order, graph gap handling, artifact mutation rules, and close/abort sequence.

## Outputs

- exact KAH command sequence
- phase/gate verdict summary
- `graph-evidence.md` when graph state affects the run, or when graph capability is missing for requested graph-managed workflow
- blocker reason when KAH fails closed or a required command exits non-zero

## Gate

PASS when KAH state, phase-plan validation, approval records when declared, gate reports, and run artifacts agree. FAIL when artifacts claim completion but `phase-plan validate --final`, `gate check`, or `gate final` fails. BLOCKED when KAH fails closed due to status/event mismatch, lock conflict, unsafe path, stale schema state, or missing repository root.
