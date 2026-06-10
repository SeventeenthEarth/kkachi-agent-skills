# Run State Flow

This reference expands the operational flow in `../SKILL.md`.

1. Initialize or reconfigure the project with `project init`; use `--force` only for non-destructive reconfiguration.
2. Create the run with `run create`; set `--backend-evidence required` only when KHS has selected backend evidence as required, then activate it with `run activate`.
3. Create canonical baseline artifacts with `artifact init <run_id>`.
4. Initialize or inspect KHS-declared phase state with `phase-plan init <run_id>` or `phase-plan show <run_id>`. Project bootstrap alone is not a state-machine injection; a project with only `project init` is KAH-ready, but the KAS phase state machine is not active until a run exists and `phase-plan init` has created the run-local phase plan.
5. Update declared phase rows with `phase-plan set <run_id> <phase-id> ...`; use `phase-plan validate <run_id>` during the run and `phase-plan validate <run_id> --final` before final reporting.
6. Use `.kkachi-workflow.yaml` only after graph capability preflight passes for the effective KAH binary. If `capabilities --json`, `graph --help`, or the required `graph validate/explain/init` command is missing or stale, record a gap and continue only with run-local `phase-plan.yaml` evidence; never repair graph state through manual `.kkachi-workflow.yaml` edits.
7. When graph state affects the run through init, validate, explain, diff, propose, or apply, write `graph-evidence.md` from `templates/run-artifacts/graph-evidence.md.tmpl`. Preserve template id/path/version, proposal id/path, semantic diff output path, validation/explain report paths, approval/audit evidence, graph checksum/version, KAH graph audit event ids, and capability-check evidence. Missing graph support is also recorded there as a gap.
8. Populate canonical artifact files in `.kkachi/runs/<run_id>/`; prefer `artifact write` and `artifact append` when available so KAH records path-safety checks, atomic mutation, and audit events. Use `artifact set-status` only for artifacts whose status field is KAH lifecycle-owned; never use it as a blanket completion step for schema-owned backend JSON artifacts such as `selected-cli.json`.
9. Use `event append <type> --run <run_id> --payload '<json-object>'` for compact phase milestones such as `phase.started`, `phase.completed`, `artifact.updated`, or `kab.prompt.sent`.
10. Use `approval request`, `approval record`, and `approval show` when KHS has declared that a high-risk phase needs approval; KAH records the approval state but KHS decides when approval is required.
11. Use `schema validate` for `selected-cli.json` and `bridge-session-snapshot.json` when those artifacts are present; use `schema export` or `schema migrate` only for explicit schema maintenance work.
12. Use `artifact validate <run_id> --gate intake` for intake validation.
13. Use `gate check <run_id> <gate>` after the evidence for each implemented gate is complete.
14. Use `gate final <run_id>` before the final report.
15. Use `lock recover` only for explicit stale-lock recovery with a durable reason.
16. Close successful runs with `run close`; abort failed or abandoned runs with `run abort`.
