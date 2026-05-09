# Feedback for KAH Development Team

Date: 2026-05-10
From: KHS user review and phase orchestration design
Target: `kkachi-agent-helper` (KAH)
Related projects: `kkachi-hermes-skills` (KHS), `kkachi-agent-bridge` (KAB)
Status: discussion note to send to the KAH team; not KHS SOT until promoted

## Summary

KHS has been reshaped around a clearer KAH/KAB/KHS boundary:

- **KHS** is the Hermes skill/process layer. It decides whether a request is a
  Kkachi-governed run, builds the task contract, owns the phase workflow, selects
  KAB backend lanes from evidence, asks KAB backends to do substantive work, and
  final-verifies the result.
- **KAH** is the deterministic state, artifact, event, schema, diagnostics, lock,
  and gate layer. It should not decide task intent, backend preference, phase
  applicability, or implementation quality.
- **KAB** is the backend runtime/control layer for external AI CLIs and exposes
  session evidence.

KHS `main` should not pin itself to one KAH or KAB patch release. It should
install or verify the latest compatible helper command surface:

```bash
go install github.com/SeventeenthEarth/kkachi-agent-helper@latest
```

KHS release tags should still publish tested/recommended KAH and KAB versions for
reproducible installs.

During KHS workflow design, several additional helper needs became visible:

- `phase-plan.yaml` is now the KHS workflow source of truth for a run.
- `plan.md` and `checklist.md` are mandatory plan artifacts. KHS captures KAB
  planner output into `plan.md` and normalizes a KHS-owned progress checklist
  into `checklist.md`.
- KHS code-change runs use KAB by default. Therefore KAH needs a way to require
  backend evidence independently of `execution_mode`.
- KHS may auto-start low-risk work but must record explicit master approval for
  high-risk or ambiguous work.
- KHS request-feedback/handle-feedback runs at least once and may run up to three
  rounds; skipped or not-applicable phases require explicit reasons.

This document requests KAH command-surface and documentation changes that let KHS
use KAH `@latest` safely while keeping KAH deterministic.

## Current KHS activation policy

KHS should trigger only for a Kkachi-governed software run.

KHS should be used when any of these are true:

- the user explicitly asks to use KHS, Kkachi, KAH, KAB, or a Kkachi run;
- the user asks to apply KHS to a project directory;
- development work is delegated to a KHS-using commander such as 조운 or 마초;
- the work needs KAB backend execution, backend lane selection, bridge evidence,
  red-team review, gate reports, or durable run artifacts;
- the work is a substantial feature, risky refactor, multi-phase QA,
  production write, readiness hardening, or discovery/shaping run.

KHS should not trigger by default for ordinary direct Hermes edits such as typo
fixes, small one-file patches, quick config tweaks, or read-only explanations and
reviews, unless the user explicitly asks for KHS/Kkachi or delegates the work to
a KHS-using commander.

KAH should not decide whether KHS is triggered. KAH should provide deterministic
surfaces after KHS or the user has chosen to apply the Kkachi workflow.

## Live KAH v0.1.1 observations

Rechecked against the PATH binary on 2026-05-10:

```text
/Users/draccoon/go/bin/kkachi-agent-helper
kkachi-agent-helper 0.1.1
```

Observed gaps:

```text
kkachi-agent-helper capabilities --json     -> unknown_command, exit 2
kkachi-agent-helper help                    -> unknown_command, exit 2
kkachi-agent-helper --help                  -> unknown_command, exit 2
kkachi-agent-helper project init --help     -> unknown_option, exit 2
kkachi-agent-helper run create --help       -> missing_option_value, exit 2
```

Source/docs review also indicates:

- backend gate behavior is manifest-driven;
- `adapter_qa` requires backend artifacts;
- `production_write` can pass backend gate as not applicable when backend
  artifacts are not required by metadata;
- KAH already documents that `project init` replaces the old KAH install model;
- KAH does not currently have a first-class `phase-plan` API;
- KAH artifact commands initialize, list, and validate artifacts, but do not yet
  provide a deterministic artifact write/status update command surface.

## Requested changes

### P1-1. Add command-surface capability output

#### Problem

KHS `main` prefers KAH `@latest`, so KHS should verify command-surface
capabilities rather than pinning one helper patch version. Probing individual
commands is brittle and produces poor user-facing diagnostics.

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
    "project_init_force": true,
    "run_lifecycle": true,
    "artifact_manifest": true,
    "artifact_validation": true,
    "artifact_write": true,
    "phase_plan": true,
    "phase_plan_validation": true,
    "gates": true,
    "backend_gate": true,
    "backend_evidence_requirement": true,
    "approval_records": true,
    "events": true,
    "schemas": true,
    "schema_migration": true,
    "diagnostics": true,
    "locks": true,
    "install_command": false
  },
  "commands": {
    "project_init": "available",
    "run_create": "available",
    "artifact_init": "available",
    "artifact_write": "available",
    "phase_plan_init": "available",
    "phase_plan_set": "available",
    "gate_check": "available",
    "schema_validate": "available",
    "diagnostics_export": "available"
  }
}
```

Exact field names can change, but the output should let KHS answer:

- Can this KAH initialize a Kkachi project?
- Can this KAH create runs and initialize artifacts?
- Can this KAH validate gates and schemas?
- Can this KAH record backend evidence requirements?
- Can this KAH validate phase-plan and checklist completeness?
- Can this KAH export diagnostics?
- Does this KAH intentionally omit the old install command?

#### Acceptance criteria

- `kkachi-agent-helper capabilities --json` exits `0` on a healthy binary.
- Output is stable enough for KHS install-guide checks.
- Output includes helper version and project schema version.
- Output clearly shows that Hermes skill installation is not a KAH capability.
- Missing, deprecated, or optional capabilities are explicit rather than inferred
  from command failures.

### P1-2. Allow KHS to require backend evidence independently of execution mode

#### Problem

KAH currently treats the backend gate as manifest-driven. In the current model,
backend evidence is required for `adapter_qa`, but a `production_write` run can
return a not-applicable pass for the backend gate when backend artifacts are not
listed in `run-metadata.json.required_artifacts`.

That behavior is correct for direct/non-KAB production work. However, KHS often
uses KAB-backed implementation for real `production_write` work. In that case,
KHS needs KAH to require backend evidence even though the execution mode remains
`production_write`.

Execution mode and KAB usage should be separate concepts.

#### Request

Add an explicit way for KHS to declare that a run requires KAB/backend evidence.

Possible interface options:

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

When backend evidence is required, KAH should include the canonical backend
artifacts in `required_artifacts` and make `gate check <run_id> backend` fail
closed until they are complete.

Canonical backend artifacts:

- `selected-cli.json`
- `capability-check.md`
- `bridge-session-snapshot.json`
- `bridge-events.md`

#### Acceptance criteria

- A `production_write` run can explicitly require backend evidence.
- A direct/non-KAB `production_write` run can still keep backend evidence as not
  applicable.
- `artifact init` records required backend artifacts when the run requires them.
- `gate check backend` fails closed when required backend evidence is missing or
  incomplete.
- Existing `adapter_qa` behavior remains compatible.
- KAH does not choose the backend or judge commander reasoning; it validates
  declared artifact shape and completion only.

### P1-3. Add first-class or deterministic `phase-plan` support

#### Problem

KHS now treats `phase-plan.yaml` as the workflow source of truth for a run. KAH
metadata such as `work_path`, `work_mode`, and `execution_mode` remains helper
classification metadata; it should not decide which KHS phases execute.

Today KHS can write `phase-plan.yaml` as a supplemental artifact, but KAH has no
first-class command surface to initialize, update, validate, or report phase
state. This leaves too much deterministic bookkeeping in direct file writes.

#### Request

Add a KAH command surface for KHS-declared phase plans. KAH should store and
validate the declared plan but must not intelligently decide phase applicability.

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

Validation should be deterministic. Examples:

- required phases must not be missing;
- skipped or not-applicable phases must include a reason;
- feedback rounds must be within the configured range, currently 1 to 3;
- code-change runs should include optimize evidence or an explicit skip reason;
- `ask`, `request-feedback-1`, and `handle-feedback-1` must be represented even
  when they produce no actionable question/feedback;
- final verification should confirm that all required phase rows have terminal
  states and evidence links.

#### Acceptance criteria

- KHS can initialize and update phase state through KAH without direct mutation
  of helper-managed metadata.
- KAH validates declared phase-plan structure and completeness deterministically.
- KAH does not choose phases, reorder phases intelligently, choose backends, or
  infer user intent.
- Phase validation can be used by `gate final` or a dedicated phase gate.
- Diagnostics export includes `phase-plan.yaml` or its KAH-managed equivalent.

### P1-4. Add deterministic artifact write/status update commands

#### Problem

KAH currently initializes and validates artifacts, but KHS still has to write key
artifacts directly as files. Direct file writes work, but they bypass a useful
place for KAH to enforce paths, append events, preserve audit metadata, and keep
status consistent.

This matters for KHS because KHS must preserve KAB plan output before
implementation starts and must update `checklist.md` after each phase.

Key artifacts include:

- `phase-plan.yaml`
- `task-contract.yaml`
- `prompt.md`
- `plan.md`
- `acceptance-criteria.md`
- `checklist.md`
- `ask.md`
- `feedback-1.md`, `feedback-2.md`, `feedback-3.md`
- `handle-feedback-1.md`, `handle-feedback-2.md`, `handle-feedback-3.md`
- `docs-update.md`
- `verification.md`
- `final-report.md`

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

KAH should enforce safe paths and known artifact names. It may support a
controlled supplemental-artifact namespace for KHS-owned artifacts such as
`phase-plan.yaml` and `task-contract.yaml`.

#### Acceptance criteria

- Artifact write commands refuse unsafe paths and unknown unmanaged locations.
- Writes are atomic or fail closed.
- KAH records an event for artifact write/update operations.
- KAH can distinguish canonical artifacts from KHS supplemental artifacts.
- Existing direct file compatibility remains possible during migration.
- `artifact list --json` reflects updated status or metadata where applicable.

### P1-5. Stabilize plan/checklist gate contract

#### Problem

KAH already requires `acceptance-criteria.md`, `plan.md`, and `checklist.md` for
the plan gate. KHS now depends on that contract:

1. KAB planner backend produces plan text and a `KHS Checklist Seed`.
2. KHS preserves plan text into `plan.md` before implementation starts.
3. KHS normalizes the checklist seed into `checklist.md`.
4. KAH validates that these artifacts exist and are complete before the run
   advances.

This is currently workable, but the contract should be explicit and stable so
future KAH changes do not accidentally weaken the KHS plan lifecycle.

#### Request

Document and, if needed, strengthen the plan gate contract:

- plan gate requires `acceptance-criteria.md`, `plan.md`, and `checklist.md`;
- `checklist.md` should contain every KHS canonical phase row, including skipped
  or not-applicable rows with reasons;
- `plan.md` should be preserved before any KAB approval/start action that may
  begin implementation;
- plan gate should fail closed if mandatory plan/checklist artifacts are missing
  or incomplete.

Optional future command:

```bash
kkachi-agent-helper gate check <run_id> plan --strict-checklist --json
```

#### Acceptance criteria

- KAH docs specify the plan gate artifact contract.
- KAH tests protect the `acceptance-criteria.md` + `plan.md` + `checklist.md`
  requirement.
- KAH remains neutral about who authored the plan; it only validates the
  declared artifact completeness and shape.
- KHS can rely on the contract across KAH `@latest` within a compatible release
  line.

### P1-6. Improve standard help UX

#### Problem

Current command discovery is awkward. For example:

```text
kkachi-agent-helper help
# unknown_command

kkachi-agent-helper project init --help
# unknown_option, then a useful hint
```

The hints are helpful, but the command still looks like a failure to users and
automation.

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
kkachi-agent-helper phase-plan --help
kkachi-agent-helper approval --help
```

#### Acceptance criteria

- Help commands exit `0`.
- Help output lists required arguments and options.
- `project init --help` includes the full bootstrap option list.
- JSON mode either returns structured help or fails with clear documented
  behavior; structured help is preferred for automation.

### P2-1. Add approval record support for high-risk phases

#### Problem

KHS may auto-start low-risk implementation after notifying the master, but
explicit master approval is required for high-risk or ambiguous cases:

- API, DB, security, dependency, architecture, or SOT changes;
- large diffs;
- low confidence;
- unresolved ask-phase ambiguity.

KHS can record this manually in artifacts or events, but a KAH-native approval
record would make final verification and diagnostics cleaner.

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

This can also be implemented as a stricter wrapper around `event append` if KAH
prefers not to add a new top-level command.

#### Acceptance criteria

- KHS can record approval requests and decisions with phase, reason, decision,
  approver, timestamp, and evidence reference.
- Approval records are included in diagnostics export.
- KAH does not decide whether approval is needed; it records the declaration and
  decision.
- Final verification can check that required approvals are present when the
  phase plan says they were required.

### P2-2. Document the KHS main-vs-release compatibility policy

#### Problem

KHS `main` should remain easy to use with the current KAH and KAB latest
versions, but KHS release tags need reproducible dependency guidance. Users who
install a released KHS version should know which KAH and KAB versions were tested
and recommended at release time.

#### Request

Define a compatibility metadata shape that KHS releases can publish and KAH/KAB
can reference from their own compatibility docs.

Suggested KHS release metadata shape:

```yaml
name: kkachi-hermes-skills
version: 0.2.0
compatibility:
  policy:
    main_branch: latest-compatible
    release_tags: tested-and-recommended
  kah:
    install_ref_for_main: latest
    tested_version: v0.1.x
    recommended_version: v0.1.x
    required_capabilities:
      - project_init
      - run_lifecycle
      - artifact_manifest
      - artifact_write
      - phase_plan
      - gates
      - backend_gate
      - backend_evidence_requirement
      - schemas
      - diagnostics
  kab:
    install_ref_for_main: latest
    tested_version: v0.x.y
    recommended_version: v0.x.y
    required_capabilities:
      - backend_sessions
      - selected_cli_evidence
      - bridge_session_snapshot
      - retained_events_or_cli_loop
      - plan_capture
```

Exact file name can be decided by the KHS team. Candidate names:

- `compatibility.yaml`
- `release-compatibility.yaml`
- `skill-pack.yaml` compatibility block

KAH does not need to consume KHS release metadata directly, but KAH docs should
acknowledge the policy:

- KHS `main` may verify KAH by `@latest` plus capabilities.
- KHS release tags may recommend a specific tested KAH version.
- KAH should expose enough capability information for both workflows.

#### Acceptance criteria

- KAH compatibility docs describe the difference between KHS `main` and KHS
  release tags.
- KAH docs do not require KHS `main` to pin a KAH patch version.
- KAH docs allow KHS releases to publish tested/recommended KAH versions for
  reproducibility.
- The policy does not make KAH responsible for choosing or installing KHS.

### P2-3. Document the KHS/KAH `@latest` compatibility contract

#### Problem

KHS install docs now prefer `@latest`. KAH docs should state what KHS can rely on
across the current release line.

#### Request

Add a short compatibility section to KAH docs that states:

- KAH does not install Hermes skills.
- KAH does not decide whether KHS should trigger.
- KAH project bootstrap is through `project init` / `project init --force`.
- KHS should validate KAH by command-surface capabilities rather than pinning a
  single patch release when possible.
- KAH owns deterministic state only after KHS or the user chooses to apply the
  Kkachi workflow.
- KAH may validate KHS-declared `phase-plan.yaml` but must not decide phase
  applicability intelligently.

#### Acceptance criteria

- README and compatibility docs reflect the same boundary.
- No stale references imply that KAH `install` installs KHS/Hermes skills.
- The docs make it clear that simple direct Hermes edits do not require KAH.
- The docs distinguish helper metadata (`work_path`, `work_mode`,
  `execution_mode`) from KHS phase workflow authority (`phase-plan.yaml`).

## Non-goals

Do not move these responsibilities into KAH:

- deciding whether a user request should trigger KHS;
- choosing the best backend lane;
- planning or implementing software changes;
- running external AI CLIs;
- installing Hermes skills;
- deciding which phases are applicable from task semantics;
- judging commander reasoning or output quality beyond deterministic artifact
  validation.

## Suggested implementation order

1. Add `capabilities --json`.
2. Add backend evidence declaration support for production runs.
3. Add `phase-plan` initialization/update/validation support.
4. Add deterministic artifact write/status update commands.
5. Stabilize and document the plan/checklist gate contract.
6. Add standard help command handling.
7. Add approval record support or a documented approval event schema.
8. Document the KHS main-vs-release compatibility policy.
9. Update README / compatibility docs for the KHS/KAH `@latest` contract.

## Why this matters

These changes let KHS use KAH `@latest` safely on `main`, while allowing KHS
release tags to publish tested/recommended KAH and KAB versions for reproducible
installs. The boundary remains clean:

- KHS decides when the Kkachi workflow applies.
- KHS owns phase workflow policy through `phase-plan.yaml`.
- KAH deterministically records and validates the selected workflow.
- KAB provides backend runtime evidence when the selected workflow uses backend
  sessions.

The result is less version pinning, fewer false KHS activations for small edits,
stronger fail-closed behavior for KAB-backed production work, and clearer audit
evidence for plan preservation, checklist progress, feedback rounds, approvals,
and final verification.
