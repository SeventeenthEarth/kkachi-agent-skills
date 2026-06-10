# KHS+KAH minimum/pilot CLI lane

Date: 2026-05-23
Owner: KHS documentation archive
Confirming role: Blue-confirmed project authority in kanban task `t_caee9433` for the stated scope; candidate record from `t_1af0dc98` is confirmed for that scope
Status: Blue-confirmed SOT / roadmap record for the scoped KHS+KAH minimum/pilot harness lane
Authority level: confirmed lane split and safety constraints; does not replace the current KHS+KAH+KAB execution-runtime SOT
Scope: KHS docs and future `kkachi-agent-skills` CLI planning only; no KAH code, KAB docs, runtime configs, profiles, registries, gateway settings, KHC CLI, or Doksuri integration changes
Related docs: `interface-contract.md`, `kas-cli-contract.md`, `../roadmap.md`, `../README.md`, repository `README.md`
Evidence/source paths: kanban task `t_caee9433` Blue confirmation, kanban task `t_3e6d8b89` Blue final synthesis, and child task `t_1af0dc98`

## Current state

Post-KAH note: KAH 0.1.4 now advertises graph and configurable-feedback
capabilities, while `capabilities --json` still reports `install_command=false`.
Therefore KAS owns the CLIMVP install/list/doctor surface; KAH may provide
project doctor, graph proposal/evidence, and deterministic project substrate,
but KAH is not the skill-pack installer.

KHS documentation already records the full KHS execution path as a KHS+KAH+KAB run path for code-changing KHS work. That remains valid for full Kkachi-governed execution-runtime use.

The Blue-confirmed design direction in `t_caee9433`, following `t_3e6d8b89` and `t_1af0dc98`, adds a narrower lane: a KHS+KAH minimum/pilot harness where a future `kkachi-agent-skills` CLI helps users install/profile-inject KHS skills, inspect installed state, compare/sync with explicit approval, and create/validate proposal/evidence records. This lane is not a runner or bridge controller.

## Accepted decision

Two lanes must be kept distinct:

1. KHS+KAH minimum/pilot harness lane
   - Purpose: let users pilot KHS skill-pack installation and KAH-backed proposal/evidence workflows without requiring KAB runtime readiness.
   - Allowed CLI surface: `list`, `install`, `doctor`, `sync`, and `proposal`. `docs/sot/kas-cli-contract.md` is the accepted CLIMVP-001 command-surface contract for the first MVP slice: `list`, `install --dry-run`, approved copy install, and `doctor`; `sync` and broad `proposal` remain gated future surfaces.
   - Allowed authority: profile-scoped KHS skill-pack support, install/sync reporting, KAH availability checks, project-overlay/proposal evidence preparation, and self-improvement proposal validation.
   - Explicitly not allowed: `run` verbs, backend session control, bridge control, KHC command/control, Doksuri integration, or KAB replacement.

2. Full execution-runtime lane
   - Purpose: run KHS-governed work that includes code-changing or backend-executed phases.
   - Authority retained: KHS+KAH+KAB remains the current authoritative path for KAB-backed code-change KHS runs.
   - This minimum/pilot lane does not downgrade, replace, or remove the full execution-runtime path.

## Ownership boundaries

- KHS owns profile/global skill-pack content, phase and process guidance, project overlay semantics, semantic self-improvement policy, and future `kkachi-agent-skills` CLI wording for KHS skill-pack operations.
- KAH owns deterministic project-local state, graph proposals, run artifacts, schemas, events, locks, gates, diagnostics, and evidence persistence. KAH is not a Hermes skill installer and must not become workflow-policy authority.
- KAB owns backend runtime/session control, prompt dispatch, plan/question/approval/input handling, retained events, status/read evidence, and bridge execution evidence for the full execution-runtime lane.

## CLI safety constraints

`kkachi-agent-skills install`:

- Default MVP mode is copy mode into `~/.hermes/profiles/<profile>/skills/<category>/<skill>/`.
- The canonical CLIMVP-001 forms are `kkachi-agent-skills install --profile <profile> <pack-id>... --dry-run` and `kkachi-agent-skills install --profile <profile> <pack-id>... --approve <evidence-ref>`, as defined in `docs/sot/kas-cli-contract.md`.
- The older positional form `kkachi-agent-skills install <profile> <skill-or-category> --dry-run` may remain a human convenience alias only if it resolves to the same internal contract and JSON shape.
- A write-capable install must support dry-run before mutation.
- Mutation reports must name target profile, source pack/version, planned files, actual changed paths, manifest/checksum results, and recovery or rollback instructions.
- Native `hermes -p <profile> skills install ...` may be referenced only for the single-skill identifier or direct `SKILL.md` URL behavior that has been verified; repo-root multi-skill-pack install must not be claimed until separately evidenced.

`skills.external_dirs`:

- Developer/pilot mode only.
- It is not the general-user default because one shared directory edit can propagate across multiple profiles and break role isolation.

Symlink mode:

- Deferred or developer-only.
- It must not be the default for general users because it creates cross-profile mutation, filesystem portability, and silent drift risks.

`kkachi-agent-skills sync <profile>`:

- Must not mutate without dry-run, diff, explicit approval, and recovery path.
- Must compare the installed profile copy against the source KHS pack and report planned changes before applying them.
- Must fail closed when manifest/checksum or backup/recovery evidence is missing.

`kkachi-agent-skills proposal <project> ...`:

- May create or validate proposals and evidence records.
- Must not automatically mutate shared KHS, profile skills, project overlays, or KAH graph state without approval/audit evidence.
- KAH graph state changes remain KAH proposal/apply work with approval/audit evidence; KHS proposal wording must not imply direct fallback mutation.

## First-run operator path

The safe minimum user path is:

1. `kkachi-agent-skills list [--profile <profile>] [--category <name>]`
2. `kkachi-agent-skills install --profile <profile> <pack-id>... --dry-run`
3. operator approval for the reported changed paths
4. `kkachi-agent-skills install --profile <profile> <pack-id>... --approve <evidence-ref>` with manifest/checksum and recovery report
5. `kkachi-agent-skills doctor --profile <profile> [--project <path>]`
6. next-action report that clearly says whether the user is in the minimum/pilot lane or must use the full KHS+KAH+KAB execution-runtime lane

## Self-improvement ledger split

- Profile install/sync ledger: records which KHS skills were copied into which profile, with source version and checksums.
- KAH graph proposal records: record project workflow graph/config proposals and approval/audit evidence.
- Run-local `.kkachi/runs/<run_id>/improvement-note.md`: records improvement candidates discovered during actual runs.
- Shared KHS promotion gate: accepts shared KHS changes only after generalized lessons, evidence artifacts, trigger eval/baseline grading where applicable, and explicit approval.

## Rejected or deferred options

- Rejected for this lane: `kkachi-agent-skills run`, backend session control, bridge control, KHC command/control, Doksuri integration, or treating the CLI as a KAB substitute.
- Deferred: repo-root Hermes multi-skill-pack install claims until current Hermes behavior is verified.
- Deferred or developer-only: symlink mode.
- Developer/pilot only: `skills.external_dirs`.
- Not allowed: automatic mutation of shared KHS, profile skills, project overlays, or KAH graph state without approval/audit evidence.

## Stale/conflict markers

- Existing README/SOT wording that says whole-stack KAB verification is required must be read as full execution-runtime guidance, not as a prerequisite for the scoped minimum/pilot install/profile-injection lane.
- Existing KHS code-change run language remains authoritative for code-changing KHS execution: KHS+KAH+KAB is still required for KAB-backed runs.
- Any future wording that presents the minimum/pilot lane as Kkachi/KHC/Doksuri/KAB run/control scope is conflicting and must be corrected before implementation.

## Resolved decisions and open questions

- CLIMVP-001 resolved the minimum manifest/checksum contract for profile installs in `docs/sot/kas-cli-contract.md`, including SHA-256 file and pack checksums, dry-run plan hashes, approved-plan matching, changed-path categories, and the recommended profile metadata path.
- Approved `install` backup/recovery behavior is implemented under CLIMVP-004 for the profile-scoped minimum lane; exact backup/rollback behavior for future `sync` remains separately gated.
- Exact `proposal` CLI arguments and mapping to KAH proposal/evidence paths remain future implementation design.

## Next record action

CLIMVP-001 through CLIMVP-005 are implemented and closed for the profile-scoped minimum CLI lane. Future implementation tasks should proceed one at a time from the remaining roadmap rows; `sync` and broad `proposal` remain gated until their fail-closed behavior is specified and reviewed.
