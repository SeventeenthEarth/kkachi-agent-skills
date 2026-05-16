# Feedback for KAH Development Team

Date: 2026-05-12
From: KHS user review, phase orchestration design, and planner-output contract decision
Target: `kkachi-agent-helper` (KAH)
Related projects: `kkachi-hermes-skills` (KHS), `kkachi-agent-bridge` (KAB)
Status: discussion note to send to the KAH team; not KHS SOT until promoted

## Executive summary

KHS has settled the immediate planner-output policy:

- KAB owns backend session control and plan lifecycle evidence.
- KHS owns workflow policy and the normalized `checklist.md` artifact.
- KAH owns deterministic state, artifacts, schemas, gates, events, locks, and diagnostics.
- Missing or malformed KAB `KHS Checklist Seed` is **not** by itself a KAH/KHS hard fail.
- KHS should recover by deriving `checklist.md` from `plan.plan_text`, the task contract, acceptance criteria, phase contract, expected evidence, and gate requirements.
- Hard fail is reserved for missing/empty plan text, plan ambiguity, explicit KAB ambiguity evidence such as OpenCode `PLAN_AMBIGUOUS`, or inability to preserve required plan evidence before implementation starts.

Therefore the KAH request is not to parse KAB plan text or enforce a KAB-specific checklist seed. The KAH request is to provide deterministic surfaces that let KHS preserve, validate, and audit the artifacts it owns.

## Handoff request for the KAH team

Please treat this document as the handoff request from KHS to KAH. No separate interpretation layer is required.

Implement the requested changes in priority order:

1. Start with **P0-1**: stabilize and document the plan/checklist gate contract around KHS-owned checklist normalization.
2. Then implement **P0-2**: allow KHS to declare backend evidence as required independently of `execution_mode`.
3. Continue through **P1** and **P2** items after P0 is complete, unless the KAH maintainer finds a smaller prerequisite change is needed first.

The first KAH PR should be small and reviewable. A good first PR scope is:

- update KAH docs/specs to state that KHS owns `checklist.md` normalization;
- protect the plan gate contract requiring `acceptance-criteria.md`, `plan.md`, and `checklist.md`;
- state that KAH does not parse or require KAB `KHS Checklist Seed` sections.

## Boundary to preserve

KAH must not:

- decide whether a user request should trigger KHS;
- choose backend lanes;
- run external AI CLIs;
- parse KAB planner intent semantically;
- decide phase applicability from task semantics;
- plan or implement software changes;
- install Hermes skills;
- judge commander reasoning beyond deterministic artifact validation.

KAH may:

- store and validate KHS-declared state;
- enforce artifact and gate completeness;
- fail closed when declared requirements are missing;
- validate artifact shape, required fields, status markers, and evidence links;
- expose command-surface capabilities for KHS `@latest` compatibility checks;
- include phase-plan, approvals, backend evidence, and artifact status in diagnostics.

## Current live observations

Rechecked against the local helper binary on 2026-05-12:

```text
/Users/draccoon/go/bin/kkachi-agent-helper --version
kkachi-agent-helper 0.1.1
```

Observed command-surface gaps:

```text
kkachi-agent-helper --help                  -> unknown_command, exit 2
kkachi-agent-helper help                    -> unknown_command, exit 2
kkachi-agent-helper project init --help     -> unknown_option, exit 2
kkachi-agent-helper run create --help       -> missing_option_value, exit 2
kkachi-agent-helper capabilities --json     -> unknown_command, exit 2
kkachi-agent-helper phase-plan --help       -> unknown_command, exit 2
```

Source/docs review indicates the helper already has strong foundations:

- deterministic project init, status, events, locks, run lifecycle, artifact init/list/validate, gates, schemas, diagnostics;
- plan gate requiring `acceptance-criteria.md`, `plan.md`, and `checklist.md`;
- manifest-driven backend gate for backend artifacts;
- diagnostics export and gate reports;
- explicit boundary that KAH no longer installs Hermes skills.

The remaining feedback is therefore focused on KHS integration surfaces, not on replacing KAH's core architecture.

## Requested changes

### P0-1. Stabilize the plan/checklist gate contract around KHS-owned checklist normalization

#### Problem

KHS now treats `plan.md` and `checklist.md` as mandatory pre-implementation artifacts:

1. KAB planner backend exposes plan text/evidence.
2. KHS preserves KAB plan text into `plan.md` before approval/start.
3. KHS derives the normalized progress ledger into `checklist.md`.
4. KAH validates that the required plan artifacts exist and are complete before implementation proceeds.

The important policy change is that KHS no longer requires a perfect KAB `KHS Checklist Seed`. Missing or malformed checklist seed is recoverable when the plan itself is usable.

#### Request

Document and test the plan gate as a stable KHS/KAH contract:

- `gate check <run_id> plan` requires completed `acceptance-criteria.md`, `plan.md`, and `checklist.md`.
- KAH should not require or parse the KAB `KHS Checklist Seed` section.
- KAH should validate the artifacts KHS writes, not the original KAB planner format.
- `checklist.md` should be allowed to be a KHS-normalized artifact derived from multiple inputs.
- Plan gate should fail closed if mandatory artifacts are missing, empty, or not marked complete.
- If KAH later validates checklist rows strictly, skipped/not-applicable phase rows must include explicit reasons.

Optional future interface:

```bash
kkachi-agent-helper gate check <run_id> plan --strict-checklist --json
```

#### Acceptance criteria

- KAH docs state that KHS owns checklist normalization.
- KAH tests protect the `acceptance-criteria.md` + `plan.md` + `checklist.md` requirement.
- KAH does not depend on KAB-specific planner text sections.
- KHS can rely on this contract across compatible KAH `@latest` updates.

### P0-2. Allow KHS to require backend evidence independently of execution mode

#### Problem

KAH's current backend gate is manifest-driven. This is correct, but KHS often uses KAB-backed implementation for real `production_write` work. In that case, backend evidence must be required even though the execution mode is still `production_write`.

Execution mode and KAB usage should be separate concepts.

#### Request

Add an explicit way for KHS to declare backend evidence requirements.

Possible interface:

```bash
kkachi-agent-helper run create ... --backend-evidence required
```

or:

```bash
kkachi-agent-helper run update <run_id> --backend-evidence required --json
```

or a KAH-managed metadata field:

```json
{
  "backend_evidence": "required"
}
```

When backend evidence is required, KAH should include the canonical backend artifacts in `required_artifacts` and make the backend gate fail closed until they are complete.

Canonical backend artifacts:

- `selected-cli.json`
- `capability-check.md`
- `bridge-session-snapshot.json`
- `bridge-events.md`

#### Acceptance criteria

- A KAB-backed `production_write` run can explicitly require backend evidence.
- A direct/non-KAB `production_write` run can still keep backend evidence not applicable.
- `artifact init` records required backend artifacts when the run requires them.
- `gate check <run_id> backend` fails closed when required backend evidence is missing or incomplete.
- KAH does not choose or override the backend; it validates declared artifact shape and completion only.

### P1-1. Add command-surface capability output

#### Problem

KHS `main` should use KAH `@latest` where possible, but it needs to verify command-surface compatibility by capability rather than by a fragile patch-version pin.

#### Request

Add a machine-readable command such as:

```bash
kkachi-agent-helper capabilities --json
```

Suggested output shape:

```json
{
  "name": "kkachi-agent-helper",
  "version": "0.1.1",
  "schema_version": "0.1",
  "capabilities": {
    "project_init": true,
    "run_lifecycle": true,
    "artifact_manifest": true,
    "artifact_validation": true,
    "gates": true,
    "backend_gate": true,
    "backend_evidence_requirement": true,
    "phase_plan": true,
    "approval_records": true,
    "diagnostics": true,
    "install_command": false
  },
  "commands": {
    "project_init": "available",
    "run_create": "available",
    "artifact_init": "available",
    "gate_check": "available",
    "diagnostics_export": "available"
  }
}
```

Exact field names can change, but the output should let KHS answer:

- Can this KAH initialize a project?
- Can this KAH create runs and initialize artifacts?
- Can this KAH validate gates and schemas?
- Can this KAH require backend evidence when KHS declares it?
- Can this KAH validate or report phase-plan state?
- Does this KAH intentionally omit the old install command?

#### Acceptance criteria

- `capabilities --json` exits `0` on a healthy binary.
- Output includes helper version and project schema version.
- Missing, deprecated, or optional capabilities are explicit.
- Output is stable enough for KHS install/activation checks.

### P1-2. Add first-class or deterministic `phase-plan` support

#### Problem

KHS treats `phase-plan.yaml` as the workflow source of truth for a run. KAH metadata such as `work_path`, `work_mode`, and `execution_mode` remains helper classification metadata; it should not decide which KHS phases execute.

Today KHS can write `phase-plan.yaml` directly, but KAH has no first-class command surface to initialize, update, validate, or report phase state.

#### Request

Add a deterministic command surface for KHS-declared phase plans. KAH should store and validate the declared plan but must not intelligently decide phase applicability.

Possible interface:

```bash
kkachi-agent-helper phase-plan init <run_id> --from phase-plan.yaml --json
kkachi-agent-helper phase-plan show <run_id> --json
kkachi-agent-helper phase-plan set <run_id> <phase> \
  --state <pending|running|passed|failed|skipped|not_applicable> \
  --reason <reason> \
  --evidence <artifact-path> \
  --json
kkachi-agent-helper phase-plan validate <run_id> --json
```

Deterministic validations can include:

- required phases are present;
- skipped or not-applicable phases include reasons;
- feedback rounds stay within the configured range, currently 1 to 3;
- code-change runs include optimize evidence or an explicit skip reason;
- `ask`, `request-feedback-1`, and `handle-feedback-1` are represented even when they produce no actionable question or feedback;
- final verification confirms required phase rows have terminal states and evidence links.

#### Acceptance criteria

- KHS can initialize and update phase state through KAH without direct mutation of helper-managed metadata.
- KAH validates declared phase-plan structure and completeness deterministically.
- KAH does not choose phases, reorder phases intelligently, choose backends, or infer user intent.
- Diagnostics export includes `phase-plan.yaml` or its KAH-managed equivalent.

### P1-3. Improve standard help UX

#### Problem

Current command discovery is awkward for humans and automation. Help attempts return usage errors instead of stable help output.

#### Request

Support standard help commands:

```bash
kkachi-agent-helper help
kkachi-agent-helper --help
kkachi-agent-helper project --help
kkachi-agent-helper project init --help
kkachi-agent-helper run --help
kkachi-agent-helper run create --help
kkachi-agent-helper artifact --help
kkachi-agent-helper gate --help
kkachi-agent-helper diagnostics --help
kkachi-agent-helper phase-plan --help
```

#### Acceptance criteria

- Help commands exit `0`.
- Help output lists required arguments and options.
- JSON mode either returns structured help or fails with clear documented behavior; structured help is preferred.

### P2-1. Add deterministic artifact write/status update commands

#### Problem

KHS currently writes major artifacts directly. Direct file writes work, but they bypass a useful KAH point for safe path enforcement, audit events, atomic writes, and artifact status updates.

This matters most for pre-start preservation of `plan.md` and repeated updates to `checklist.md` and `phase-plan.yaml`.

#### Request

Add deterministic artifact mutation commands such as:

```bash
kkachi-agent-helper artifact write <run_id> plan.md --from <file> --json
kkachi-agent-helper artifact write <run_id> checklist.md --from <file> --json
kkachi-agent-helper artifact append <run_id> bridge-events.md --from <file> --json
kkachi-agent-helper artifact set-status <run_id> checklist.md \
  --status complete \
  --reason <reason> \
  --json
```

#### Acceptance criteria

- Artifact write commands refuse unsafe paths and unknown unmanaged locations.
- Writes are atomic or fail closed.
- KAH records an event for artifact write/update operations.
- KAH can distinguish canonical artifacts from KHS supplemental artifacts.
- Existing direct file compatibility remains possible during migration.

### P2-2. Add approval records for high-risk phases

#### Problem

KHS may auto-start low-risk work, but high-risk or ambiguous cases require explicit master approval. KHS can record this manually today, but a KAH-native approval record would make diagnostics and final verification cleaner.

#### Request

Add deterministic approval commands or a documented event schema:

```bash
kkachi-agent-helper approval request <run_id> \
  --phase implement \
  --reason <reason> \
  --json

kkachi-agent-helper approval record <run_id> \
  --phase implement \
  --decision approved \
  --by master \
  --evidence <artifact-or-message-ref> \
  --json

kkachi-agent-helper approval show <run_id> --json
```

This can also be implemented as a strict wrapper around `event append` if KAH prefers not to add a top-level command.

#### Acceptance criteria

- KHS can record approval requests and decisions with phase, reason, decision, approver, timestamp, and evidence reference.
- Approval records are included in diagnostics export.
- KAH does not decide whether approval is needed; it records the declaration and decision.
- Final verification can check approvals when the phase plan says they were required.

### P2-3. Document the KHS/KAH compatibility contract

#### Request

Add a short KAH README/docs compatibility section stating:

- KAH does not install Hermes skills.
- KAH does not decide whether KHS should trigger.
- KAH project bootstrap is through `project init` / `project init --force`.
- KHS `main` may verify KAH by `@latest` plus command-surface capabilities.
- KHS release tags may publish tested/recommended KAH versions for reproducibility.
- KAH owns deterministic state only after KHS or the user chooses to apply the Kkachi workflow.
- KAH may validate KHS-declared `phase-plan.yaml` but must not decide phase applicability intelligently.

## Suggested implementation order for KAH

1. Stabilize/document the plan/checklist gate contract around KHS-owned checklist normalization.
2. Add backend evidence declaration support for KAB-backed production runs.
3. Add `capabilities --json`.
4. Add `phase-plan` initialization/update/validation/reporting support.
5. Improve standard help command handling.
6. Add deterministic artifact write/status update commands.
7. Add approval records or a strict approval event schema.
8. Update README/compatibility docs for the KHS/KAH `@latest` contract.

## Why this matters

These changes let KHS use KAH `@latest` safely while keeping the boundary clean:

- KHS decides when the Kkachi workflow applies.
- KHS owns phase workflow policy and checklist normalization.
- KAH deterministically records and validates declared workflow evidence.
- KAB provides backend runtime evidence when the selected workflow uses backend sessions.

The result is less version pinning, stronger fail-closed behavior for KAB-backed production work, cleaner plan preservation, clearer checklist progress, and more useful diagnostics for feedback rounds, approvals, and final verification.
