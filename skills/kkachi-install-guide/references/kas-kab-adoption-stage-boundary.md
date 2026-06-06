# KAS KAB adoption stage boundary

Use this reference when installing or changing a project-specific KAS suite and the project must declare whether it runs as Stage 1, Stage 2, or Stage 3.

## Ownership

KAS owns the adoption-stage policy. KAH does not need to interpret the stage semantically: graph state, run state, events, locks, artifacts, schemas, and gates use the same helper mechanics across all stages.

KAB owns backend execution and observation evidence when Stage 2 or Stage 3 uses bridge-backed execution. Official GLM Octo review remains an independent KAB GLM review lane and is not replaced by the implementation-stage setting.

## Preferred installed-profile record

For project-specific KAS installations, record the active stage inside the installed Hermes profile skill suite, not in the target project repository.

Preferred layout:

```text
~/.hermes/profiles/<profile>/skills/<project-name>/<project-name>-kas/SKILL.md
~/.hermes/profiles/<profile>/skills/<project-name>/<project-name>-kas/references/kab-adoption-stage.md
```

For current split projects this means examples such as:

```text
~/.hermes/profiles/hwangchung/skills/kan-plugin/kan-plugin-kas/references/kab-adoption-stage.md
~/.hermes/profiles/hwangchung/skills/kan-control/kan-control-kas/references/kab-adoption-stage.md
```

The umbrella project KAS skill should explicitly say to read `references/kab-adoption-stage.md` before selecting planner/implementer lanes.

## Stage record shape

Keep the file human-readable and grep-friendly. Recommended contents:

```yaml
kas_kab_adoption_stage: stage2_kab_codex_first
allowed_implementation_backends:
  - codex
stage_owner: KAS
kah_semantics: unchanged_graph_run_artifact_gates
glm_octo_review: independent_review_lane
changed_from: stage1_direct_codex_app_server_baseline
changed_reason: "KAB native_codex execution is ready for this project"
evidence_required:
  - selected_cli_json_when_kab_backed
  - capability_check_md_when_kab_backed
  - bridge_events_md_when_kab_backed
  - direct_codex_no_kab_rationale_when_stage1
```

Use one of these stage values:

- `stage1_direct_codex_app_server_baseline`
- `stage2_kab_codex_first`
- `stage3_kab_backend_selected`

## When to touch KAH project state

Do not run `kkachi-agent-helper project init --force` merely because the KAS stage changed in the installed profile skill suite. Use KAH reconfiguration only when project-local KAH overlay/backend-policy state must also be rewritten for durability, multi-operator visibility, or project-local policy mirroring.

If KAH project state is rewritten, preserve:

- previous stage
- new stage
- reason for change
- exact KAH command and `--force` use
- `project doctor --json` result
- next-run task/phase evidence proving the selected lane

## Fail-closed rule

If the stage marker is missing, invalid, or ambiguous, KAS must fail closed to Stage 1 behavior. Direct Codex evidence may be recorded, but KAB Codex execution must not be claimed unless Stage 2 or Stage 3 is explicitly selected and evidenced.
