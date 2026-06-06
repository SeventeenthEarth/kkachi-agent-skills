# KAS roadmap

Date: 2026-05-25
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record; INITDOC post-KAH reset
Status: post-KAH KAS MVP roadmap; KAH 0.1.4 graph/configurable-feedback substrate evidenced; KAS integration work remains separately gated
Authority level: KAS roadmap; not implementation authorization by itself
Scope: KAS docs/skills planning only; no KAH code, KAB docs, runtime configs, profiles, registries, or gateway changes
Related docs: `README.md`, `sot/khs-architecture-and-integration.md`, `sot/workflow-graph-integration.md`, `sot/minimum-pilot-cli-lane.md`, `sot/kas-cli-contract.md`, `sot/project-kas-sync-state.md`, `sot/external-feedback-intake-policy.md`, `sot/phase-orchestration-policy.md`, `sot/interface-contract.md`
Evidence/source paths:
- Governance evidence record in kanban task `t_2fb00394`
- Blue final synthesis in kanban task `t_3e6d8b89` and Gray docs task `t_1af0dc98`
- INITDOC-002 Red acceptance in kanban task `t_b805ce76`
- INITDOC-003 Red acceptance in kanban task `t_81453049`
- Effective KAH check on 2026-05-25: `kkachi-agent-helper --version` => `kkachi-agent-helper 0.1.4`; `kkachi-agent-helper graph --help` reports supported `init`, `validate`, `explain`, `diff`, `propose`, `apply`, and `export`

## Purpose

This roadmap records the post-KAH KAS MVP launch path. KAH 0.1.4 command/capability evidence exists for `kkachi-agent-helper graph` and configurable feedback intake. KAS integration implementation still starts one PR-candidate task at a time only after the responsible approver authorizes that task and the effective binary capability/help checks pass for the specific run.

The MVP target is a usable KAS skill/process pack that can:

1. bootstrap this repository as a KAH-managed project;
2. list, dry-run install, approved-install, and doctor profile-scoped KAS skill packs;
3. provide KAS-owned workflow graph templates and capability-checked KAH graph guidance;
4. clean stale active KAS surfaces so they no longer reintroduce pre-KAH blockers.

## Task sizing policy

- One task is one PR candidate.
- Do not bundle implementation tasks after this docs update.
- Implementation starts only after SOT/spec confirmation, responsible approver authorization, and required risk review / operator workflow review when risk/operator workflow changes apply.
- Each later implementation task must include acceptance criteria, evidence, verification, and docs updates for only the changed surface.
- Write-capable profile installs, sync behavior, graph apply automation, and shared KAS promotion all require explicit approval evidence before mutation.

## Status values

`Planned`, `In Progress`, `Blocked`, `Completed`, `Deferred`.

## MVP delivery order

```text
INITDOC-003
  -> INITDOC-004
  -> BOOTSTRAP-001..003
  -> CLIMVP-001..005
  -> GRAPHMVP-001..004
  -> STALECLEAN-001..004
  -> KABADOPT-001..004
  -> KASUPD-001..004
```

INITDOC is closed before implementation begins so the temporary transition SOT does not become new legacy. `BOOTSTRAP` should happen first because it gives later KAS work a deterministic KAH project state and doctor evidence. `STALECLEAN` may run in parallel with late CLIMVP/GRAPHMVP tasks only when the touched surfaces do not overlap.

## Active roadmap

### EPIC: INITDOC — post-KAH KAS authority reset and MVP roadmap completion

> Goal: remove stale pre-KAH blockers from active KAS documentation and complete an implementation-ready MVP roadmap so KAS work can proceed immediately after the reset. INITDOC was a temporary transition epic; its transition SOT was deleted after its decisions were absorbed into permanent docs.

| Task ID | Title | Status | Work guide | Notes |
|---|---|---|---|---|
| INITDOC-001 | Create temporary INITDOC transition SOT and roadmap epic | Completed | Created `docs/sot/initdoc-post-kah-reset.md`, defined INITDOC scope/exit criteria, and registered this epic plus four tasks in `docs/roadmap.md`. | Temporary transition SOT only; deleted by INITDOC-004 after absorption into active docs. |
| INITDOC-002 | Reset active docs/SOTs to post-KAH authority | Completed | Update active docs so KAH 0.1.4-evidenced graph/configurable-feedback surfaces are not blocked by stale pre-KAH wording; keep KAB-later, alias, approval, and mutation gates explicit. | Minimum targets updated; Blue verification passed and Red review accepted in `t_b805ce76`. |
| INITDOC-003 | Complete implementation-ready MVP roadmap | Completed | Rework this roadmap into executable post-INITDOC epics/tasks for BOOTSTRAP, CLIMVP, GRAPHMVP, and STALECLEAN with acceptance criteria, deferrals, and evidence requirements. | This roadmap now leads directly into MVP feature development instead of another planning loop. |
| INITDOC-004 | Absorb INITDOC decisions and remove transition SOT | Completed | Verified INITDOC decisions are absorbed into permanent docs, deleted `docs/sot/initdoc-post-kah-reset.md`, and left absorbed-target/completion history below. | INITDOC is closed; implementation can proceed one PR-candidate task at a time after responsible approver authorization. |

### EPIC: BOOTSTRAP — KAS repo KAH project bootstrap

> Goal: make `kkachi-hermes-skills` a valid KAH-managed project so KAS work can record deterministic project evidence before feature implementation begins.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| BOOTSTRAP-001 | Approve and run KAH `project init` for this repo | Completed | A responsible approver authorizes the exact project init values; `kkachi-agent-helper project init ... --json` creates KAH-managed project bootstrap state; `.kkachi/` is ignored as local operator state, and the generated `docs/kkachi-docs-map.yaml` is preserved as the repo-visible KAH docs map. | Preserved KAH event/run evidence, `project status --json`, `project doctor --json`, changed paths, and the amended BOOTSTRAP-001 commit with the docs map plus `.gitignore` policy update. |
| BOOTSTRAP-002 | Make `project doctor` pass and record evidence | Completed | `kkachi-agent-helper project doctor --json` exits successfully in this repo after bootstrap; docs-map, schema, status, events, paths, and locks all pass without remediation. | Preserved doctor JSON and human summary in KAH run `run-20260525T160007Z-bc45aeb340b3`; doctor result was 12 passed, 0 warnings, 0 failed, with no remediation commands required. |
| BOOTSTRAP-003 | Decide `.kkachi/` repository policy | Completed | The repo records generated `.kkachi/` files as ignored local KAH/operator state, while `docs/kkachi-docs-map.yaml` remains committed repo-visible KAH docs-map metadata. | 주군 chose the `.kkachi/` ignore policy; preserved evidence in KAH run `run-20260525T160512Z-9bd8d42e73b4`; verification includes `git check-ignore -v .kkachi/status.json`, `git status --short --branch`, docs readback, and `git diff --check`. |

### EPIC: CLIMVP — profile-scoped KAS skill-pack list/install/doctor

> Goal: let a user safely discover, install, and verify KAS skills in a Hermes profile without requiring KAB runtime readiness or turning KAS into KAH/KAB.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| CLIMVP-001 | Specify command surface and manifest/checksum contract | Completed | `docs/sot/kas-cli-contract.md` defines `list`, `install --dry-run`, approved copy install, and `doctor`; it defines manifest fields, SHA-256 checksum and dry-run plan-hash rules, source pack identity, target profile identity, changed-path report, backup/recovery instructions, status vocabulary, and failure modes. | Docs/spec-only task completed with KAH run `run-20260525T161641Z-4160f06cf1be`. Reviews resolved: 하후연 `t_2fd07174` ACCEPT, 여몽 `t_e04116b6` then `t_abc30909` ACCEPT, 진궁 `t_b8db3ead` then `t_e652a24c` ACCEPT, 주유 `t_6cc40c59` HARNESS_ACCEPT, 사마의 `t_e479171c` STRATEGIC_RED_ACCEPT. KAH `install_command=false` and KAS ownership are stated. |
| CLIMVP-002 | Implement skill-pack discovery and `list` | Completed | `kkachi-hermes-skills list [--profile <profile>] [--category <name>] [--json]` or the approved equivalent lists available KAS skill packs/categories without mutating files; output distinguishes source repo packs from installed profile copies. | Reimplemented as stdlib-only Go module `github.com/SeventeenthEarth/kkachi-hermes-skills` with entrypoint `cmd/kkachi-hermes-skills`. Verification: Go unit discovery/profile-state tests, CLI JSON tests, real-repo e2e no-profile-write coverage, `make test`, `git diff --check`, `go install ./cmd/kkachi-hermes-skills`, and JSON/human CLI smoke commands. |
| CLIMVP-003 | Implement dry-run install with changed-path report | Completed | `install --dry-run` resolves source packs and target profile paths, reports creates/updates/conflicts/skips, validates manifest/checksum input, and performs no writes. | Verification: focused RED first for missing Go packages/binary, then GREEN for `go test ./...` / `make test` covering discovery, install dry-run, CLI, and actual-binary e2e. Approved copy install remains fail-closed until CLIMVP-004. |
| CLIMVP-004 | Implement approved copy install with manifest/checksum and recovery | Completed | Approved install copies selected KAS skill packs into `~/.hermes/profiles/<profile>/skills/...`, records manifest/checksum evidence, reports actual changed paths, writes `.kas/skill-pack-manifest.json`, and prints recovery/rollback instructions. | Implemented from KAH run `run-20260526T053031Z-523b017121f1`. Verification: strict RED first for missing approved API / old CLI stub / e2e stub, then GREEN for focused approved-install unit, CLI, and actual-binary e2e tests plus `make test`, `git diff --check`, and `go install ./cmd/kkachi-hermes-skills` with temp Go cache/bin. No `skills.external_dirs`, symlink, gateway, auth, model, KAH, or KAB config mutation. |
| CLIMVP-005 | Implement profile/project `doctor` | Completed | `doctor` verifies source pack integrity, target profile install state, manifest/checksum consistency, KAH availability/version/capabilities, optional project bootstrap/doctor status, and whether KAB is required for the requested lane. | Implemented from KAH run `run-20260526T074524Z-b1dd9c07d7ce`. Verification: strict RED first for missing doctor package/CLI command, then GREEN for focused doctor unit tests, CLI tests, actual-binary e2e, `make test`, `git diff --check`, `go install ./cmd/kkachi-hermes-skills`, and representative JSON/human doctor smoke with a temp profile root. |

Deferred from CLIMVP unless separately approved: `sync`, broad `proposal`, symlink mode, shared `skills.external_dirs` default, KHC/Doksuri integration, KAB run/control verbs, and repo-root Hermes multi-skill-pack install claims.

### EPIC: GRAPHMVP — KAS graph template registry and KAH graph adoption

> Goal: let KAS provide workflow graph templates/policy and use KAH graph surfaces without manual `.kkachi-workflow.yaml` repair, while preserving KAH as deterministic validator/apply/audit owner.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| GRAPHMVP-001 | Define KAS graph template registry schema | Completed | A registry/spec defines template id rules, template file paths, owner metadata, versioning, required phases, edges, gates, approvals, feedback-intake bounds, compatibility requirements, and KAH validation expectations. | Completed by `docs/sot/graph-template-registry.md`, `registries/graph-template-registry.yaml`, and valid/invalid examples; team review via 하후연/여몽/진궁 Kanban completed. |
| GRAPHMVP-002 | Add default KAS workflow graph template | Completed | A default template can initialize `.kkachi-workflow.yaml` through `kkachi-agent-helper graph init --from-template`; it encodes the current KAS MVP phase path without manual `.kkachi-workflow.yaml` repair or `init --profile`. | Completed by `templates/workflow-graphs/kas-default.yaml` and registry entry; verification includes KAH capability/help check, `graph init` in a temp repo, `graph validate`, `graph explain`, and preservation of generated checksum/evidence. |
| GRAPHMVP-003 | Add capability-checked graph guidance to KAS orchestration | Completed | KAS guidance checks effective `kkachi-agent-helper graph` support before graph use; if missing, stale, or unsupported, it records a gap and fails closed instead of pretending commands exist. | Completed by `docs/sot/workflow-graph-integration.md`, KAS orchestration skills, run-artifact templates, and `docs/examples/graph-capability-preflight.md`; verification includes docs/skill readback, stale `kah graph` alias guard, missing-capability examples, no manual `.kkachi-workflow.yaml` repair instructions, and team review follow-up. |
| GRAPHMVP-004 | Define graph evidence preservation in run artifacts/reports | Completed | KAS report/artifact guidance names template id/path/version, proposal id/path, semantic diff output, validation/explain reports, approval/audit evidence, graph checksum/version, KAH graph audit event ids, and capability-check evidence when graph changes affect a run. | Implemented by `templates/run-artifacts/graph-evidence.md.tmpl`, task/phase/capability templates, SOT/registry/report guidance, and docs-contract tests that make `make test` verify required fields. |

Deferred from GRAPHMVP unless separately approved: graph `apply` automation, declarative graph proposal generation beyond the default template path, KAB prompt/context alignment with graph versions, `kah graph` alias assumptions, and Kkachi v2 `.kkachi/config/workflows/` merging.

### EPIC: STALECLEAN — remove stale pre-KAH blockers from active KAS surfaces

> Goal: prevent active skills, templates, registries, and operator reports from reintroducing old pre-KAH blockers after INITDOC resets the docs.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| STALECLEAN-001 | Audit active KAS stale surfaces | Completed | Produce a concrete stale-surface manifest for `skills/*/SKILL.md`, `registries/*`, `templates/*`, repository `README.md`, and docs that still mention fixed `1..3` feedback, `blocked_by_kah` for KAH-supported graph/configurable-feedback surfaces, manual `.kkachi-workflow.yaml` repair, or unproven `kah graph` alias behavior. | Completed by `docs/discussions/staleclean-001-active-surface-manifest.md` and KAH run `run-20260527T105713Z-158d70000c08`; verification includes targeted searches and a scripted active-surface scan mapping 198 candidate markers to active-update, historical/guardrail, or review-needed dispositions. |
| STALECLEAN-002 | Update active feedback-intake registries/templates/skills | Completed | Approved KAS surfaces adopt required round 1 plus optional rounds 2..5 where KAH-evidenced configurable feedback intake applies; reports preserve `kah-evidenced, kas-integration-pending` until end-to-end KAS adoption is verified. | Completed by updates to `registries/phase-contracts.yaml`, run-artifact templates, phase skills, SOT status rows, and docs-contract tests; verification includes stale round-3 searches, YAML/template parsing, affected skill readbacks, `go test ./tests/docs_contract`, and `make test`. |
| STALECLEAN-003 | Preserve KAB-later and runtime boundaries | Completed | Active guidance no longer blocks scoped CLIMVP/GRAPHMVP/KAS docs or CLI work on KAB, but still marks backend execution, automated review-by-different-tool transport, KAB plan lifecycle, and bridge evidence as KAB-dependent when required. | Completed by updates to the README, phase contracts, orchestration/plan/implement guidance, run-artifact templates, phase-orchestration SOT, and docs-contract tests; verification includes forbidden overbroad KAB-prerequisite searches, YAML parse, docs-contract tests, and full `make test`. |
| STALECLEAN-004 | Verify active KAS guidance no longer treats KAH 0.1.4 graph/configurable-feedback as absent | Completed | Final stale-clean verification proves active guidance either uses capability-checked KAH 0.1.4 graph/configurable-feedback support or explicitly marks remaining work as `kas-integration-pending`, `kab_later`, `candidate`, `historical`, or `unsupported`. | Completed by KAH 0.1.4 capability/help evidence, graph validate/explain readback, active-guidance scan, cleanup of stale concept/template/status-index wording, and docs-contract tests. Remaining operator report/e2e adoption stays `kas-integration-pending`; KAB runtime/automated review transport stays `kab_later`. |


### EPIC: KABADOPT — KAS KAB adoption stage selector and Stage 2 migration path

> Goal: make the KAS/KAH development KAB adoption stage an explicit installed/project-specific operating policy before any Stage 2 migration. This epic records the selector and marker contract first, then implements it, then pilots Stage 2 only with evidence. The roadmap entry itself is not implementation authorization and does not activate Stage 2 for any project.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| KABADOPT-001 | Specify KAS CLI KAB adoption stage selector | Completed | `docs/sot/kas-cli-contract.md` defines numeric stage choices, canonical values, default Stage 1 behavior, non-interactive no-prompt behavior, invalid/Stage 3 fail-closed handling, JSON output, marker path/content, dry-run plan-hash binding, and the plan/review fallback-audit posture. `docs/sot/khs-architecture-and-integration.md` cross-references the selector without duplicating CLI detail, and `docs/kkachi-docs-map.yaml` registers the CLI SOT. | Completed as docs-only closure by SOT/docs/skill updates, `git diff --check`, `make test`, KAH docs/roadmap evidence, and Red/Orange/Gray color review acceptance. No code implementation or Stage 2 activation is authorized by this task. |
| KABADOPT-002 | Implement numeric stage selection in KAS CLI | Completed | Stage-capable install/dry-run accepts `--kab-stage 1|2` and canonical `--kab-adoption-stage`; interactive TTY defaults blank input to Stage 1; JSON/CI/non-interactive mode never prompts and defaults to Stage 1; explicit invalid/conflicting/Stage 3 values fail closed; selected stage is included in changed paths and approval plan hash. | Completed by Go unit/CLI/e2e tests for defaults, explicit values, conflicts, Stage 3 rejection, marker generation, doctor marker reporting, and approval hash mismatch; `make test` PASS after Gray fixes; Red, Orange, and Gray focused re-review ACCEPT. |
| KABADOPT-003 | Add project-specific KAS stage marker/runbook references | Completed | Implementation pass adds compact runbook references and selected-stage evidence posture to generated KAB adoption markers while preserving parse-compatible `Numeric stage:`, `Canonical stage:`, `Selection source:`, `Selected at:`, and `Approval evidence:` lines. The canonical runbook lives at `skills/kkachi-install-guide/references/kas-kab-adoption-stage-runbook.md`; install guide/boundary references and `phase-plan.yaml.tmpl` point runs to marker readback plus the runbook. KABADOPT-003 does not activate Stage 2 or claim KAB Codex execution evidence. | Completed by focused Stage 1/Stage 2 marker content/readback unit tests, `go test ./internal/skills/install ./internal/skills/cli ./internal/skills/doctor`, `make test`, stale-term sweep, `git diff --check`, official GLM Octo ACCEPT_WITH_NOTES, and post-Octo Red/Orange/Gray color review acceptance. KABADOPT-004 remains the Stage 2 pilot. |
| KABADOPT-004 | Pilot Stage 2 KAB Codex-first on one approved project | Planned | One explicitly selected project changes the marker to `stage2_kab_codex_first`; implementation/fix/docs-bound execution uses KAB `native_codex`; evidence records selected CLI/capability, session id, plan/read/status/wait or retained stream events, bridge events, final snapshot, and any rejected fallback. | 주군 project choice/approval, KAB preflight, KAH run artifacts, first color review, official GLM Octo when required by task class/policy, post-fix re-review, final verification, and commit approval. |

KABADOPT deferrals unless separately approved: Stage 3 backend-selected implementation, Antigravity lane work, non-Codex KAB implementation backends for KAS/KAH development, silent fallback from Stage 2 to direct Codex, and mutation of auth/tokens/gateway/model/provider config.

### EPIC: KASUPD — project-specific KAS upstream sync and semantic-port workflow

> Goal: let KAN, KLM, `kan-plugin`, `kan-control`, and other project-specific KAS suites adopt new shared KAS policy without overwriting project-local authority, roadmap IDs, test commands, or role boundaries. This epic records the static state YAML first, then implements dry-run evidence, three-way classification, semantic-port prompting, and one approved pilot. The roadmap entry itself is not write authorization for any installed profile or project-specific KAS suite.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| KASUPD-001 | Specify project-specific KAS state and sync workflow SOT | Completed | `docs/sot/project-kas-sync-state.md` defines `kas-project-state.yaml` as the canonical project-specific KAS static state file, combining KAB adoption stage, upstream KAS commit/checksum baselines, project skill mapping, overlay policy, dry-run evidence use, three-way classification, semantic-port merge rules, and fail-closed conditions. `docs/README.md` and `docs/kkachi-docs-map.yaml` register the SOT. | Completed as docs-only SOT freeze with `git diff --check`, docs-map YAML parse, `make test`, KAH docs/roadmap evidence, and Red/Orange/Gray ACCEPT. No implementation, profile mutation, or project KAS update is authorized by this task. |
| KASUPD-002 | Implement state YAML validation and compatibility read | Completed | `sync-project-kas --profile <profile> --project <project-id> --state <path> --dry-run [--json]` reads and validates `kas-project-state.yaml`, validates schema version, stage, upstream commit/checksum fields, pack mappings, overlay policy, and evidence posture; it reads old `kab-adoption-stage.md` markers for compatibility reporting only. Machine output distinguishes `yaml_state_path` from `legacy_marker_path`. Missing/unreadable/invalid YAML fails closed to Stage 1 claims. | Completed with implementation/docs updates, unit/CLI/e2e validation coverage, invalid YAML/schema and compatibility-marker fixtures, Blue verification (`gofmt`, `git diff --check`, `HOME=/Users/draccoon make test`, coverage, CLI JSON smoke, docs-map parse), and Red/Orange/Gray ACCEPT reviews (`t_8a1f37d3`, `t_44a4334a`, `t_f4be7e82`). No auth/token/gateway/provider mutation; KASUPD-003/004 remain planned. |
| KASUPD-003 | Implement dry-run three-way classification and semantic-port packet | Completed | `sync-project-kas --dry-run` now validates state, compares current upstream KAS, recorded baseline checksums, and mapped project-local skills, emits the six documented classifications, represents unchanged mappings separately, and includes semantic-port packet JSON content without mutating project KAS files, profiles, manifests, KAH state, KAB state, or packet files. Dirty source repos and fail-closed conflicts return `ok:false` with diagnostics and non-zero exit. | Completed with Go implementation/docs updates and verification: `gofmt`, `git diff --check`, `go test ./internal/skills/kasstate ./internal/skills/cli ./tests/e2e ./tests/docs_contract`, `make test`, e2e binary JSON semantic-packet/no-write coverage, checksum mismatch/dirty-source/missing-or-ambiguous project-skill fail-closed fixtures, and docs-contract checks. KASUPD-004 write-capable sync/pilot remains planned. |
| KASUPD-004 | Pilot project-specific KAS update on one approved suite | Planned | One 주군-approved project-specific suite such as KAN, KLM, `kan-plugin`, or `kan-control` uses the state YAML plus dry-run/semantic-port workflow to adopt selected upstream KAS changes while preserving project-local authority and selected KAB stage. | 주군 project choice/approval, preflight state readback, dry-run evidence, semantic-port diff, verification, Blue/Red/Orange/Gray review, official GLM Octo only if the task class/policy requires it, final KAH gate, and commit approval. |

KASUPD deferrals unless separately approved: broad write-capable sync across multiple projects, automatic overwrite of project-specific local changes, shared `skills.external_dirs` behavior, auth/token/gateway/provider config mutation, Stage 3 backend-selected policy updates, and any hidden fallback from selected project KAS stage.

## Deferred / non-MVP work

- KAS write-capable sync behavior after KASUPD dry-run/state/semantic-port evidence, approval, recovery spec, and Red/operator review.
- Broad `proposal` CLI mapping across profile install ledgers, project overlays, KAH graph proposals, run-local improvement notes, and shared KAS promotion gates.
- KAB context/prompt alignment with applied graph version/checksum.
- KAS/KAH development Stage 2 KAB Codex-first adoption remains planned under KABADOPT until selector implementation, marker/runbook references, and one approved pilot provide evidence.
- Project-specific KAS upstream sync remains planned under KASUPD until state YAML, dry-run classification, semantic-port packet, and one approved pilot provide evidence.
- Automated review-by-different-tool transport and evidence; this remains `kab_later`.
- KHC/Doksuri integration and command/control surfaces.
- Repo-root Hermes multi-skill-pack install claims until Hermes behavior is verified.
- `kah graph` alias support unless effective alias capability/help evidence exists.

## INITDOC completion record

INITDOC is closed. The temporary transition SOT `docs/sot/initdoc-post-kah-reset.md` was deleted after its decisions were absorbed into the active records below:

- `docs/roadmap.md` — post-KAH MVP delivery order, INITDOC completion history, BOOTSTRAP / CLIMVP / GRAPHMVP / STALECLEAN task boundaries, deferrals, evidence requirements, and next action.
- `README.md` — current KAS/KAH/KAB lane split, KAH 0.1.4 maturity note, profile-scoped minimum/pilot lane, graph/configurable-feedback posture, and KAB-later runtime boundary.
- `docs/README.md` — authority ladder, status vocabulary, and post-INITDOC next-record guidance.
- `docs/sot/interface-contract.md` — KAS/KAH/KAB reality-first interface boundary, KAH 0.1.4 capability evidence, graph alias boundary, and minimum/pilot lane boundary.
- `docs/sot/khs-architecture-and-integration.md` — broad current KAS architecture SOT, self-improvement governance, layer boundaries, and active stale-surface inventory.
- `docs/sot/workflow-graph-integration.md` — KAS/KAH graph integration SOT, capability-checked `kkachi-agent-helper graph` use, no manual `.kkachi-workflow.yaml` repair, and graph evidence preservation requirements.
- `docs/sot/minimum-pilot-cli-lane.md` — profile-scoped KAS skill-pack list/install/doctor lane and safety constraints.
- `docs/sot/external-feedback-intake-policy.md` — KAH-evidenced configurable feedback-intake policy, KAS adoption dependencies, and stale-surface manifest.
- `docs/sot/khs-pre-kah-readiness-audit-2026-05.md` — historical/superseded pre-KAH audit narrowed so it no longer blocks post-KAH KAS MVP work.
- `docs/sot/khs-delegation-packet-and-report-contract.md` — packet/report status labels, fail-closed rules, and readiness example updated to reference active SOTs instead of the deleted transition file.

Implementation remains gated by each later task's responsible approver authorization, capability/help evidence, and review requirements.

## Stale/conflict markers

- Older wording that treats `phase-plan.yaml` as the whole workflow SOT is narrowed to run-local execution state/evidence for one run.
- Any wording that requires KAB verification for scoped profile install/injection, `list`, `doctor`, `sync`, or `proposal` planning must be read as full execution-runtime guidance only.
- Any planned `kah graph` alias reference is candidate until alias capabilities/help evidence proves implementation. `kkachi-agent-helper graph` references still require effective-binary capability/help checks before use.
- Kkachi v2 `.kkachi/config/workflows/` is out of scope and must not be merged with or used as fallback for `.kkachi-workflow.yaml`.
- Fixed `1..3`, `max_rounds: 3`, or `maximum_rounds: 3` feedback-intake behavior in active KAS surfaces is stale unless the specific text is explicitly historical.

## Open questions

- `.kkachi/` repo policy is closed by BOOTSTRAP-003: generated `.kkachi/` files are ignored local KAH/operator state; `docs/kkachi-docs-map.yaml` is committed repo-visible KAH metadata.
- CLIMVP manifest/checksum file format is specified by `docs/sot/kas-cli-contract.md`; `CLIMVP-001` through `CLIMVP-005` are implemented and closed.
- Approved profile install backup/recovery behavior is implemented for CLIMVP; future `sync` recovery behavior remains separately gated.
- KAS graph template registry schema and default `kas-default` template content are defined by `docs/sot/graph-template-registry.md`, `registries/graph-template-registry.yaml`, and `templates/workflow-graphs/kas-default.yaml`.
- Exact KAS artifact mapping for KAH proposal, semantic diff, validation/explain, approval/audit, checksum/version, event-id, and capability-check evidence is defined by GRAPHMVP-004 in `templates/run-artifacts/graph-evidence.md.tmpl` and the graph SOTs.
- Project-specific KAS sync/update is now tracked by KASUPD; implementation still requires a separately approved KASUPD task.
- KAB alignment requires a separately assigned KAB docs/update task.

## Next record action

INITDOC, BOOTSTRAP, `CLIMVP-001` through `CLIMVP-005`, `GRAPHMVP-001` through `GRAPHMVP-004`, `STALECLEAN-001` through `STALECLEAN-004`, `KABADOPT-001` through `KABADOPT-003`, and KASUPD-001 through KASUPD-003 are completed with the Go module path `github.com/SeventeenthEarth/kkachi-hermes-skills`. KABADOPT-004 remains `Planned` until its own Stage 2 pilot evidence/review gates pass. KASUPD-004 remains `Planned` until its own write-capable sync/pilot evidence gates pass. No further STALECLEAN task is listed; any epic-level closure needs a dedicated review/update task.
