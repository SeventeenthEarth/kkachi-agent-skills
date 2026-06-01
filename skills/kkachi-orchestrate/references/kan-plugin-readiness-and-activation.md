# Kan-plugin Readiness and KAS Activation

Use this reference when preparing a Kkachi-governed project directory such as `kkachi-agent-network-plugin` for its first KAS/KAH development run.

## Activation boundary lesson

Do not treat KAS task taxonomy as the commander's global behavior. Default Hwangchung mode remains direct commander response. Switch to KAS mode only when the master explicitly activates KHS/Kkachi/KAB/KAH, applies KHS to a project directory, asks for durable repo artifact changes under a governed Kkachi project, requires phase/gate/evidence tracking, selects KAB-backed execution/KAB plan lifecycle/bridge evidence, or needs long-lived team collaboration.

Inside active KAS mode, classify task class before selecting phases. Outside KAS mode, do not force `task_class` or create a KAH run for ordinary chat, quick status checks, or simple non-durable commands.

## Investigation/spec/roadmap mapping

- Current-state investigation only: `research_evidence`.
- Investigation followed by spec/SOT/roadmap/acceptance/handoff edits: `docs_only + Path B shaping`.
- Code, tests, build behavior, executable contracts, or future execution-policy changes: `development`.

## First-run readiness checks

Before claiming a project is ready for first development run, verify and report:

1. KAH project health: `kkachi-agent-helper project status --json` and `project doctor --json` show healthy state and no lock/event issues.
2. Active run state: no active run is required before the first task; readiness means `run create` + `artifact init` + `phase-plan init` can now be executed.
3. CodeGraph evidence: `.codegraph/` exists and `codegraph status <repo>` is up to date, or record initialization/degraded reason.
4. Repo tests: project aggregate test command passes, even if current scaffold is docs-only and some layers intentionally skip.
5. Backend/runtime capability evidence when future development is expected: for current P1/kan-plugin, verify direct Codex app-server/Codex CLI availability and record no-KAB rationale; check KAB only as future readiness/gap evidence, not as the active execution lane.
6. Git hygiene: `.kkachi/` and `.codegraph/` are ignored; any prep artifacts such as `.gitignore` or docs-map changes are reported as separate bootstrap changes so they do not get silently mixed into the first development task.

## Profile sync pitfall

When a KAS skill support file exists in the profile but is not recorded in the KAS manifest, `doctor` may show checksum/manifest drift. Safe recovery is: absorb the support file into source, commit source, back up the profile skill directory, remove the target directory, dry-run install that pack, then approved-install it so the manifest records all files. Do not silently delete profile-local content without a backup and source absorption decision.
