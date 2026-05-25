# KAS roadmap

Date: 2026-05-25
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record; INITDOC post-KAH reset
Status: post-KAH KAS MVP roadmap; KAH 0.1.4 graph/configurable-feedback substrate evidenced; KAS integration work remains separately gated
Authority level: KAS roadmap; not implementation authorization by itself
Scope: KAS docs/skills planning only; no KAH code, KAB docs, runtime configs, profiles, registries, or gateway changes
Related docs: `README.md`, `sot/initdoc-post-kah-reset.md`, `sot/khs-architecture-and-integration.md`, `sot/workflow-graph-integration.md`, `sot/minimum-pilot-cli-lane.md`, `sot/external-feedback-intake-policy.md`, `sot/phase-orchestration-policy.md`, `sot/interface-contract.md`
Evidence/source paths:
- Governance evidence record in kanban task `t_2fb00394`
- Blue final synthesis in kanban task `t_3e6d8b89` and Gray docs task `t_1af0dc98`
- INITDOC-002 Red acceptance in kanban task `t_b805ce76`
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
```

`INITDOC-004` should close the temporary transition SOT before implementation begins so INITDOC does not become new legacy. After that, `BOOTSTRAP` should happen first because it gives later KAS work a deterministic KAH project state and doctor evidence. `STALECLEAN` may run in parallel with late CLIMVP/GRAPHMVP tasks only when the touched surfaces do not overlap.

## Active roadmap

### EPIC: INITDOC — post-KAH KAS authority reset and MVP roadmap completion

> Goal: remove stale pre-KAH blockers from active KAS documentation and complete an implementation-ready MVP roadmap so KAS work can proceed immediately after the reset. INITDOC is a temporary transition epic; its transition SOT must be deleted after its decisions are absorbed into permanent docs.

| Task ID | Title | Status | Work guide | Notes |
|---|---|---|---|---|
| INITDOC-001 | Create temporary INITDOC transition SOT and roadmap epic | Completed | Create `docs/sot/initdoc-post-kah-reset.md`, define INITDOC scope/exit criteria, and register this epic plus four tasks in `docs/roadmap.md`. | Temporary transition SOT only; it must be removed by INITDOC-004 after absorption into active docs. |
| INITDOC-002 | Reset active docs/SOTs to post-KAH authority | Completed | Update active docs so KAH 0.1.4-evidenced graph/configurable-feedback surfaces are not blocked by stale pre-KAH wording; keep KAB-later, alias, approval, and mutation gates explicit. | Minimum targets updated; Blue verification passed and Red review accepted in `t_b805ce76`. |
| INITDOC-003 | Complete implementation-ready MVP roadmap | Completed | Rework this roadmap into executable post-INITDOC epics/tasks for BOOTSTRAP, CLIMVP, GRAPHMVP, and STALECLEAN with acceptance criteria, deferrals, and evidence requirements. | This roadmap now leads directly into MVP feature development instead of another planning loop. |
| INITDOC-004 | Absorb INITDOC decisions and remove transition SOT | Planned | Verify INITDOC decisions are absorbed into permanent docs, delete `docs/sot/initdoc-post-kah-reset.md`, and leave absorbed-target/completion history in this roadmap or PR summary. | Keeping the transition SOT after completion is a failure because it would become new legacy. |

### EPIC: BOOTSTRAP — KAS repo KAH project bootstrap

> Goal: make `kkachi-hermes-skills` a valid KAH-managed project so KAS work can record deterministic project evidence before feature implementation begins.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| BOOTSTRAP-001 | Approve and run KAH `project init` for this repo | Planned | A responsible approver authorizes the exact project init values; `kkachi-agent-helper project init ... --json` creates only KAH-managed `.kkachi/` bootstrap files; no source docs/skills/templates/registries are mutated by the init command. | Preserve the approval reference, command, JSON output, changed paths, and refusal/overwrite behavior if any `.kkachi/` state already exists. |
| BOOTSTRAP-002 | Make `project doctor` pass and record evidence | Planned | `kkachi-agent-helper project doctor --json` exits successfully in this repo after bootstrap; docs-map, schema, status, events, and path checks are all explained or fixed inside KAH-owned state. | Preserve doctor JSON, human summary, and any remediation commands. If doctor fails for a policy decision, block for responsible approver instead of guessing. |
| BOOTSTRAP-003 | Decide `.kkachi/` repository policy | Planned | The repo records whether generated `.kkachi/` files are committed, ignored, or kept as local operator state; `.gitignore` or docs are updated only if the chosen policy requires it. | Requires responsible approver decision because it affects repo hygiene and evidence portability. Verification includes `git status --short`, docs/readback if changed, and `git diff --check`. |

### EPIC: CLIMVP — profile-scoped KAS skill-pack list/install/doctor

> Goal: let a user safely discover, install, and verify KAS skills in a Hermes profile without requiring KAB runtime readiness or turning KAS into KAH/KAB.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| CLIMVP-001 | Specify command surface and manifest/checksum contract | Planned | A CLI spec defines `list`, `install --dry-run`, approved copy install, and `doctor`; it defines manifest fields, checksum algorithm, source pack identity, target profile identity, changed-path report, backup/recovery instructions, and failure modes. | Docs/spec-only task. Requires operator-workflow review before implementation. Must state that KAH has `install_command=false` and that KAS owns this surface. |
| CLIMVP-002 | Implement skill-pack discovery and `list` | Planned | `kkachi-hermes-skills list [--profile <profile>] [--category <name>] [--json]` or the approved equivalent lists available KAS skill packs/categories without mutating files; output distinguishes source repo packs from installed profile copies. | Verification includes unit/CLI tests, JSON/human output examples, and docs update. No profile writes. |
| CLIMVP-003 | Implement dry-run install with changed-path report | Planned | `install --dry-run` resolves source packs and target profile paths, reports creates/updates/conflicts/skips, validates manifest/checksum input, and performs no writes. | Verification includes tests for missing profile, missing skill, existing install, conflict, path safety, and no-write behavior. Red/operator review required before enabling write mode. |
| CLIMVP-004 | Implement approved copy install with manifest/checksum and recovery | Planned | Approved install copies selected KAS skill packs into `~/.hermes/profiles/<profile>/skills/...`, records manifest/checksum evidence, reports actual changed paths, and prints recovery/rollback instructions. | Requires explicit approval flag or approval evidence. Verification includes copy behavior, checksum mismatch, partial failure recovery, path safety, and docs. No `skills.external_dirs` or symlink default. |
| CLIMVP-005 | Implement profile/project `doctor` | Planned | `doctor` verifies source pack integrity, target profile install state, manifest/checksum consistency, KAH availability/version/capabilities, optional project bootstrap/doctor status, and whether KAB is required for the requested lane. | Verification includes healthy/unhealthy fixtures, KAH missing/degraded cases, KAB-later messaging, and Korean-friendly operator summary shape. |

Deferred from CLIMVP unless separately approved: `sync`, broad `proposal`, symlink mode, shared `skills.external_dirs` default, KHC/Doksuri integration, KAB run/control verbs, and repo-root Hermes multi-skill-pack install claims.

### EPIC: GRAPHMVP — KAS graph template registry and KAH graph adoption

> Goal: let KAS provide workflow graph templates/policy and use KAH graph surfaces without direct YAML fallback, while preserving KAH as deterministic validator/apply/audit owner.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| GRAPHMVP-001 | Define KAS graph template registry schema | Planned | A registry/spec defines template id rules, template file paths, owner metadata, versioning, required phases, edges, gates, approvals, feedback-intake bounds, compatibility requirements, and KAH validation expectations. | Docs/registry task. Verification includes schema/readback checks and examples for valid and invalid template metadata. |
| GRAPHMVP-002 | Add default KAS workflow graph template | Planned | A default template can initialize `.kkachi-workflow.yaml` through `kkachi-agent-helper graph init --from-template`; it encodes the current KAS MVP phase path without direct YAML fallback or `init --profile`. | Verification includes KAH capability/help check, `graph init` in a temp repo, `graph validate`, `graph explain`, and preservation of generated checksum/evidence. |
| GRAPHMVP-003 | Add capability-checked graph guidance to KAS orchestration | Planned | KAS guidance checks effective `kkachi-agent-helper graph` support before graph use; if missing, stale, or unsupported, it records a gap and fails closed instead of pretending commands exist. | Verification includes docs/skill readback, stale `kah graph` alias guard, missing-capability examples, and no direct `.kkachi-workflow.yaml` fallback instructions. |
| GRAPHMVP-004 | Define graph evidence preservation in run artifacts/reports | Planned | KAS report/artifact guidance names template id/path, proposal id/path, semantic diff, validation report, approval/audit evidence, graph checksum/version, KAH graph audit event ids, and capability-check evidence when graph changes affect a run. | Verification includes template/report examples and compatibility with `docs/sot/workflow-graph-integration.md`. |

Deferred from GRAPHMVP unless separately approved: graph `apply` automation, declarative graph proposal generation beyond the default template path, KAB prompt/context alignment with graph versions, `kah graph` alias assumptions, and Kkachi v2 `.kkachi/config/workflows/` merging.

### EPIC: STALECLEAN — remove stale pre-KAH blockers from active KAS surfaces

> Goal: prevent active skills, templates, registries, and operator reports from reintroducing old pre-KAH blockers after INITDOC resets the docs.

| Task ID | Title | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| STALECLEAN-001 | Audit active KAS stale surfaces | Planned | Produce a concrete stale-surface manifest for `skills/*/SKILL.md`, `registries/*`, `templates/*`, repository `README.md`, and docs that still mention fixed `1..3` feedback, `blocked_by_kah` for KAH-supported graph/configurable-feedback surfaces, direct YAML fallback, or unproven `kah graph` alias behavior. | Verification includes repository searches and an audit artifact or docs section mapping each stale marker to update/defer/historical disposition. |
| STALECLEAN-002 | Update active feedback-intake registries/templates/skills | Planned | Approved KAS surfaces adopt required round 1 plus optional rounds 2..5 where KAH-evidenced configurable feedback intake applies; reports preserve `kah-evidenced, kas-integration-pending` until end-to-end KAS adoption is verified. | Requires scoped change authorization because it touches active skills/templates/registries. Verification includes searches for stale round-3 assumptions, template/schema checks, and affected skill readbacks. |
| STALECLEAN-003 | Preserve KAB-later and runtime boundaries | Planned | Active guidance no longer blocks CLIMVP/GRAPHMVP on KAB, but still marks backend execution, automated review-by-different-tool, plan lifecycle, and bridge evidence as KAB-dependent when required. | Verification includes searches for overbroad KAB prerequisites and unsupported KAB support claims. Boundary review required if wording affects execution-runtime behavior. |
| STALECLEAN-004 | Verify active KAS guidance no longer treats KAH 0.1.4 graph/configurable-feedback as absent | Planned | Final stale-clean verification proves active guidance either uses capability-checked KAH 0.1.4 graph/configurable-feedback support or explicitly marks remaining work as `kas-integration-pending`, `kab_later`, `candidate`, `historical`, or `unsupported`. | Verification includes targeted search output, changed-file summary, docs/readback, and responsible approver closure before KAS labels these surfaces implemented. |

## Deferred / non-MVP work

- KAS `sync` write behavior after dry-run/diff/approval/recovery spec and Red/operator review.
- Broad `proposal` CLI mapping across profile install ledgers, project overlays, KAH graph proposals, run-local improvement notes, and shared KAS promotion gates.
- KAB context/prompt alignment with applied graph version/checksum.
- Automated review-by-different-tool transport and evidence; this remains `kab_later`.
- KHC/Doksuri integration and command/control surfaces.
- Repo-root Hermes multi-skill-pack install claims until Hermes behavior is verified.
- `kah graph` alias support unless effective alias capability/help evidence exists.

## Stale/conflict markers

- Older wording that treats `phase-plan.yaml` as the whole workflow SOT is narrowed to run-local execution state/evidence for one run.
- Any wording that requires KAB verification for scoped profile install/injection, `list`, `doctor`, `sync`, or `proposal` planning must be read as full execution-runtime guidance only.
- Any planned `kah graph` alias reference is candidate until alias capabilities/help evidence proves implementation. `kkachi-agent-helper graph` references still require effective-binary capability/help checks before use.
- Kkachi v2 `.kkachi/config/workflows/` is out of scope and must not be merged with or used as fallback for `.kkachi-workflow.yaml`.
- Fixed `1..3`, `max_rounds: 3`, or `maximum_rounds: 3` feedback-intake behavior in active KAS surfaces is stale unless the specific text is explicitly historical.

## Open questions

- Exact `.kkachi/` repo policy for this repository remains BOOTSTRAP-003 work and requires responsible approver decision.
- Exact CLIMVP manifest/checksum file format remains CLIMVP-001 work.
- Exact backup/rollback implementation for approved profile install remains CLIMVP-001/004 work.
- Exact KAS graph template registry schema remains GRAPHMVP-001 work.
- Exact KAS artifact mapping for KAH proposal, audit, checksum, and event evidence remains GRAPHMVP-004 work.
- KAB alignment requires a separately assigned KAB docs/update task.

## Next record action

After INITDOC-003, the next work item is INITDOC-004: confirm the transition decisions are absorbed, delete `docs/sot/initdoc-post-kah-reset.md`, and leave an absorbed-target completion record. After INITDOC closes, start `BOOTSTRAP-001` unless 주군 chooses a different first implementation slice.
