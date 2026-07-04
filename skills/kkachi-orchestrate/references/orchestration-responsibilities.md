# Orchestration Responsibilities

This reference expands the responsibility summary in `../SKILL.md`.

## Required execution sequence

1. Classify task class before Path A/Path B:
   - `development`
   - `research_evidence`
   - `docs_only`
   - `simple_command_report`
   - `bootstrap_config`
   - `collaboration_review`
2. Classify Path A or Path B.
3. Select Standard or Light mode from the task class.
4. Invoke `kkachi-task-contract`.
5. Run graph capability preflight before using `.kkachi-workflow.yaml`:
   - same effective binary
   - `kkachi-agent-helper --version`
   - `capabilities --json`
   - `graph --help`
   - `graph validate/explain` for existing graph state, or `graph init --from-template` only when graph creation is intended
6. Create `phase-plan.yaml` from `templates/run-artifacts/phase-plan.yaml.tmpl` and keep it as run-local execution state/evidence.
7. Fail closed into a gap record when graph capability/help evidence is missing; do not use `kah graph` alias text or direct `.kkachi-workflow.yaml` edit fallback.
8. Invoke `kkachi-backend-select` only when a KAB-backed lane is selected or bridge evidence is claimed:
   - default v0.2 KAS/KAH work records KAS policy, KAH evidence, GJC candidate refs, and KAT factual refs instead of KAB backend evidence;
   - explicitly selected KAB work records selected backend, current capability evidence, and bridge observation evidence.
9. Invoke `kkachi-prompt-compose` before KAB delivery.
10. Use `kkachi-phase-state` and KAH at every phase boundary.
11. Preserve Korean user-facing reports and English run artifacts.
