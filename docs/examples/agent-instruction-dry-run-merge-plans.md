# Agent Instruction Dry-Run Merge Plans

These examples document the TOKEN-003 no-write planning posture for repo-local
`AGENTS.md` and `CLAUDE.md` instruction files. They are examples only: every
case reports a plan without writing target files.

## Existing Managed File

Input:

- target_path: `AGENTS.md`
- file_state: `exists`
- marker_state: `managed_block_valid`
- local_block_state: `project_local_block_valid`
- template: `templates/agent-instructions/AGENTS.md.tmpl`

Dry-run plan:

```yaml
ok: true
mode: dry_run_only
target_path: AGENTS.md
planned_state: update
managed_block: core-behavior
local_content_preservation: preserve_project_local_block
target_file_writes: []
diagnostics:
  - code: managed_block_detected
    severity: info
    message: Existing KAS managed block can be replaced by a separate approved task.
next_action: review_plan
```

Boundary: TOKEN-003 records the planned update only. It does not rewrite
`AGENTS.md`.

## Existing Unmarked AGENTS.md

Input:

- target_path: `AGENTS.md`
- file_state: `exists`
- marker_state: `no_kas_markers`
- template: `templates/agent-instructions/AGENTS.md.tmpl`

Dry-run plan:

```yaml
ok: false
mode: dry_run_only
target_path: AGENTS.md
planned_state: blocked
managed_block: core-behavior
local_content_preservation: preserve_entire_existing_file
target_file_writes: []
diagnostics:
  - code: unmarked_existing_instruction_file
    severity: blocked
    message: Existing unmarked AGENTS.md requires separate approval before managed blocks are inserted.
next_action: request_explicit_instruction_file_approval
```

Boundary: TOKEN-003 must not blind-overwrite or insert managed blocks into an
unmarked root `AGENTS.md`.

## Missing CLAUDE.md

Input:

- target_path: `CLAUDE.md`
- file_state: `missing`
- project_policy: `claude_instruction_optional`
- template: `templates/agent-instructions/CLAUDE.md.tmpl`

Dry-run plan:

```yaml
ok: true
mode: dry_run_only
target_path: CLAUDE.md
planned_state: not_applicable
managed_block: core-behavior
local_content_preservation: not_applicable
target_file_writes: []
diagnostics:
  - code: missing_optional_claude_instruction
    severity: info
    message: Project policy does not require CLAUDE.md creation in TOKEN-003.
next_action: none
```

If project policy requires `CLAUDE.md`, TOKEN-003 still records only the
dry-run plan and must not create root `CLAUDE.md`.

## Malformed Or Conflicting Markers

Input:

- target_path: `AGENTS.md`
- file_state: `exists`
- marker_state: `malformed_or_conflicting`
- examples:
  - `KAS:MANAGED:BEGIN` without matching `KAS:MANAGED:END`
  - duplicate `KAS:MANAGED:BEGIN core-behavior` blocks
  - nested `PROJECT:LOCAL:BEGIN` blocks

Dry-run plan:

```yaml
ok: false
mode: dry_run_only
target_path: AGENTS.md
planned_state: conflict
managed_block: core-behavior
local_content_preservation: preserve_entire_existing_file
target_file_writes: []
diagnostics:
  - code: malformed_or_conflicting_markers
    severity: blocked
    message: Marker conflict requires manual review; TOKEN-003 does not repair target files.
next_action: manual_review
```

Boundary: marker repair and target-file mutation are outside TOKEN-003. These
examples remain no-write plans.
