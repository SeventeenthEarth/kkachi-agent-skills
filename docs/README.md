# KHS documentation layers

Date: 2026-05-21
Owner: KHS documentation archive
Confirming role: Responsible approver / governance evidence record
Status: post-KAH docs index / authority ladder; KAH 0.1.4 graph and configurable-feedback capability evidence is recognized, CLIMVP and GRAPHMVP-001..004 KAS guidance surfaces are implemented, and KAB runtime work remains separately gated
Authority level: source of truth for how to read KHS docs after confirmation
Scope: `kkachi-hermes-skills/docs` only
Evidence/source path: governance evidence record in kanban task `t_2fb00394`

This directory separates durable KAS/KHS sources of truth from working discussion notes. It also distinguishes capability-evidenced KAH 0.1.4 `kkachi-agent-helper graph` and configurable-feedback surfaces from remaining KAS adoption work, shorthand alias behavior that still requires separate evidence, and KAB-dependent runtime claims. The minimum/pilot CLI lane is recorded separately so install/profile-injection support is not confused with the full KAS+KAH+KAB execution-runtime path.

## Authority ladder

| Path | Meaning | Authority |
|---|---|---|
| `sot/interface-contract.md` | Current KAS/KAH/KAB interface SOT | Authoritative for reality-first interface guidance; KAH 0.1.4 `kkachi-agent-helper graph` and configurable-feedback support are capability-evidenced, remaining KAS adoption is integration-pending, `kah graph` alias remains candidate, and KAB runtime claims remain later |
| `sot/khs-architecture-and-integration.md` | Broad KHS architecture/usage/KAH-integration SOT | Blue-reviewed candidate development base accepted in `t_f29b6ee9`; distinguishes target architecture from operational support claims and records the Stage 1/2/3 KAB adoption model for KAS/KAH development |
| `sot/stage1-direct-codex-sdk-appserver-runner.md` | Candidate Stage 1 direct Codex SDK/app-server runner SOT | Defines the docs/spec contract for the Stage 1 Python runner template: KAS uses `openai_codex`, the SDK spawns `codex app-server --listen stdio://`, `codex exec` and generic `openai` SDK are rejected, and Stage 2 KAB `native_codex` remains separate |
| `sot/khs-pre-kah-readiness-audit-2026-05.md` | Historical pre-KAH readiness audit | Superseded audit input from `t_2e9d918a`; no longer blocks post-KAH KAS MVP work after INITDOC closure, but still preserves overclaim and lineage warnings |
| `sot/khs-delegation-packet-and-report-contract.md` | Candidate Delegation Packet and operator report contract | Candidate docs-only contract; records packet/report fields and status labels without claiming KAH artifact or KAB runtime support |
| `sot/external-feedback-intake-policy.md` | KAH-evidenced policy SOT for configurable `EXTERNAL_FEEDBACK_INTAKE` bounds | KAH 0.1.4 advertises configurable feedback intake capability; KAS registry/template/skill adoption remains integration-pending; current MVP still uses user-supplied `feedback.md`; automated review-by-different-tool remains `kab_later` |
| `sot/phase-orchestration-policy.md` | Current KHS phase/run orchestration policy | Authoritative for run-local phase behavior |
| `sot/workflow-graph-integration.md` | Current KAS/KAH graph integration SOT | Confirmed KAS/KAH ownership, capability preflight, fail-closed guidance, and evidence rules for implemented KAH 0.1.4 `kkachi-agent-helper graph`; `kas-default` template exists and `kah graph` alias stays separately gated |
| `sot/graph-workflow-sync-compatibility.md` | Accepted SOT for KAS/KAH graph workflow sync compatibility | Defines KAS v0.1.2 compatibility metadata, KAH v0.1.9 `min_required`/`recommended`/`tested`, implemented read-only graph doctor, implemented proposal-first repair orchestration, approval-gated apply, periodic hardening posture, and user-custom graph supportability rules; active GRSYNC-001 registry is `registries/graph-workflow-sync-compatibility.yaml` |
| `sot/task-dag-workflow-contract.md` | Planning SOT for KAS task-DAG workflow and custom trigger contract | Defines WFLOW policy for multiple project task DAGs, deterministic selector behavior, node agent/role contracts, generic trigger skills, custom workflow creator scaffolding, KAS/KAH/KAB/Kanban/KAO boundaries, and the seven-PR WFLOW/DAGSM delivery sequence |
| `sot/graph-template-registry.md` | Current KAS graph template registry/default-template SOT | Defines graph template id/path/version/owner metadata, phase/edge/gate/approval/feedback-intake fields, KAH validation expectations, registry evidence requirements, and active `kas-default` template metadata |
| `sot/minimum-pilot-cli-lane.md` | Blue-confirmed KHS+KAH minimum/pilot CLI lane SOT | Confirmed lane split and safety constraints for profile-scoped install/list/doctor/sync/proposal support; does not replace full KHS+KAH+KAB execution-runtime authority |
| `sot/kas-cli-contract.md` | CLIMVP-001 KAS minimum CLI contract SOT plus KABADOPT-001 stage-selector closure | Accepted command-surface and manifest/checksum contract for profile-scoped `list`, `install --dry-run`, approved copy install, and `doctor`; also owns KAB adoption stage selector details and fallback-audit review posture; CLIMVP-001 through CLIMVP-005 are implemented and closed, while broad `proposal` remains gated future surface |
| `sot/project-specific-kas-install-contract.md` | KASPROJ-001 project-specific KAS install layout SOT | Canonical contract for installing one project-specific KAS suite per concrete project under a Hermes profile; requires `skills/<project>/<project>-<phase-or-skill>/SKILL.md`, project-prefix skill names, umbrella-only invalidity, manifest vocabulary, and doctor severity semantics; KASPROJ-002 read-only dry-run planning, KASPROJ-003 approval-hash-bound install, KASPROJ-004 doctor/repair/migrate, and KASPROJ-005 project-tailored doctor policy are implemented/in-review |
| `sot/role-aware-project-suite-contract.md` | Planning SOT for role-aware project KAS suites | Defines the pre-DAGSM/WFLOW corrective contract for full Blue commander suites versus Red/Orange/Gray role subset suites, the canonical role registry path `registries/project-suite-roles.yaml`, unknown-role fail-closed behavior, subset manifest vocabulary, role-aware doctor/repair/prune semantics, and the gate that over-installed reviewer/scribe profiles must be fixed or removed before normal local DAGSM/WFLOW development uses them |
| `sot/project-kas-sync-state.md` | KASUPD-001 project-specific KAS state/sync SOT | Accepted SOT for project-specific KAS static state: one `kas-project-state.yaml` combines KAB adoption stage, upstream KAS commit/checksum baselines, project skill mapping, overlay policy, and dry-run/semantic-merge update workflow for KAN/KLM-style suites; KASUPD-002 validates state and KASUPD-003 emits dry-run classification plus semantic-port packet evidence while write-capable sync/pilot work remains planned |
| `sot/kasrel-hermes-v016-provenance-contract.md` | KASREL-001 Hermes release-compatibility provenance/dependency SOT | Accepted docs/spec-only contract for provenance/dependency vocabulary and release-compatibility boundaries, plus the shared KASREL-004 guidance evidence gate; KASREL-002 implementation is completed/pre-commit-ready, KASREL-003 dependency-audit implementation is completed by commit `75d0361` / event `evt-001339` / run close `evt-001340`, and KASREL-004 guidance work is in progress |
| `sot/token-economy-and-agent-instruction-contract.md` | Accepted token-economy and agent-instruction SOT | Accepted contract for the selected 10 KAS PR + 2 dependent KAH PR workstream, covering English KAS-generated prompt/CLI/console/artifact-template output, compact console output, artifact-first details, task-class gating, `AGENTS.md` / `CLAUDE.md` management, color-team project KAS lifecycle gates, uninstall/vault-backup policy, skill slimming, project verification profiles, no-agent runner evidence, and KAH mechanical gates; no implementation, unapproved profile mutation, KAB activation, or Hermes runtime fork is authorized by the SOT alone |
| `sot/multi-agent-review-policy.md` | Candidate MAR repository-promotion SOT | Promotes the accepted output SOT for Kkachi Multi-Agent Review into KAS planning authority; records Blue-default MAR, role-first required coverage, degraded/fail-closed semantics, premium-review approval rules, Red adjudication triggers, stale GLM Octo marker targets, the `registries/mar-provider-lanes.json` role-lane readback source, and the split between docs-only `MAR-001`/`MAREV-001` and later implementation tasks |
| `sot/concept.md` | KHS concept SOT | Durable background; not changed by this graph docs pass |
| `sot/architecture-understanding.md` | KHS architecture understanding | Durable background; not changed by this graph docs pass |
| `sot/skill-template.md` | Skill template guidance | Durable background; not changed by this graph docs pass |
| `roadmap.md` | KAS post-KAH MVP roadmap | Active roadmap for INITDOC completion history and BOOTSTRAP / CLIMVP / GRAPHMVP / STALECLEAN PR-candidate tasks |
| `discussions/*` | Temporary or evolving discussion material | Historical/non-authoritative unless promoted into `sot/`, registries, templates, or skills |

## `sot/`

Durable source-of-truth documents for KHS behavior, architecture, and skill design. These documents should be updated deliberately because they define the shared operating model.

- `sot/concept.md`
- `sot/architecture-understanding.md`
- `sot/skill-template.md`
- `sot/khs-architecture-and-integration.md` — candidate broad SOT for KHS purpose, Hermes usage, KHS/KAH/KAB/KHC boundaries, KAB Stage 1/2/3 adoption model for KAS/KAH development, repo structure, operator report contract, self-improvement governance, and development acceptance gates; external feedback intake is summarized there and detailed by `sot/external-feedback-intake-policy.md`
- `sot/stage1-direct-codex-sdk-appserver-runner.md` — candidate Stage 1 direct Codex SDK/app-server runner SOT; requires the Python runner template to use `openai_codex` so the SDK spawns `codex app-server --listen stdio://`, rejects `codex exec` and generic `openai` SDK drift, and keeps Stage 2 KAB `native_codex` evidence separate
- `sot/khs-pre-kah-readiness-audit-2026-05.md` — historical/superseded pre-KAH readiness audit for `t_2e9d918a`; preserved for lineage and overclaim prevention, not as a post-KAH blocker
- `sot/khs-delegation-packet-and-report-contract.md` — candidate docs-only Delegation Packet and report contract; defines packet/report fields, evidence paths, and status labels without upgrading them to KAH/KAB support
- `sot/external-feedback-intake-policy.md` — KAH-evidenced policy record for configurable `EXTERNAL_FEEDBACK_INTAKE` bounds; records min=1/max=5 semantics, user-supplied `feedback.md` MVP intake, KAS integration-pending surfaces, KAB-later automation, status labels, report fields, and a stale-surface manifest
- `sot/phase-orchestration-policy.md`
- `sot/interface-contract.md`
- `sot/workflow-graph-integration.md` — current SOT for `.kkachi-workflow.yaml` / capability-checked KAH 0.1.4 `kkachi-agent-helper graph` integration; `kas-default` template and KAS orchestration guidance exist, and `kah graph` alias remains candidate until separately evidenced
- `sot/graph-workflow-sync-compatibility.md` — accepted SOT for KAS v0.1.2 graph workflow sync compatibility; KAH v0.1.9 is the `min_required`, `recommended`, and `tested` dependency for graph workflow sync; GRSYNC-001 machine-readable metadata lives in `registries/graph-workflow-sync-compatibility.yaml`; GRSYNC-002 implements `doctor --project <path> --workflow-graph --json` as a read-only supportability classifier; proposal/apply orchestration and periodic check posture are implemented by GRSYNC-003 after review/final gates
- `sot/task-dag-workflow-contract.md` — planning SOT for WFLOW task-DAG workflow support: project-local multiple DAGs, deterministic workflow selectors, KAS node agent/role contracts, generic/thin/custom trigger posture, KAH DAGSM dependency sequence, and fail-closed custom workflow creator boundaries
- `sot/graph-template-registry.md` — current GRAPHMVP-001/002 SOT for graph template registry metadata, schema expectations, and the `kas-default` template; active registry file is `registries/graph-template-registry.yaml`
- `sot/minimum-pilot-cli-lane.md` — Blue-confirmed SOT for the KHS+KAH minimum/pilot CLI lane; keeps profile-scoped install/list/doctor/sync/proposal support separate from full KHS+KAH+KAB execution-runtime authority
- `sot/kas-cli-contract.md` — CLIMVP-001 accepted SOT for KAS-owned `list`, `install --dry-run`, approved copy install, and `doctor`; also records the KABADOPT stage selector, generated marker shape, and fallback-audit review posture; records manifest/checksum, changed-path, approval, backup/recovery, status vocabulary, and fail-closed boundaries for the now-closed CLIMVP surface
- `sot/project-specific-kas-install-contract.md` — KASPROJ-001 canonical SOT for project-specific KAS install layout; requires profile targets like `skills/<project>/<project>-<phase-or-skill>/SKILL.md`, project-prefixed skill names, full suites instead of umbrella-only installs, manifest vocabulary, and doctor severity semantics; KASPROJ-002 dry-run planning, KASPROJ-003 approval-hash-bound install, KASPROJ-004 doctor/repair/migrate, and KASPROJ-005 project-tailored doctor policy are implemented/in-review, while KASPROJ-006 one operational application step remains planned
- `sot/role-aware-project-suite-contract.md` — planning SOT for role-aware project KAS suites; records that Blue commander profiles may use full suites, Red/Orange/Gray reviewer/scribe profiles require selected role subsets, profile-local overlay text is not enough to justify over-installed skills, explicit suite_role selection and unknown-role fail-closed behavior are required, and KASROLE-001..004 now provide source registry/install/doctor/repair-prune behavior while real over-installed KAH development color-profile cleanup remains approval-gated before those profiles are treated as healthy for DAGSM/WFLOW work
- `sot/project-kas-sync-state.md` — KASUPD-001 SOT for project-specific KAS `kas-project-state.yaml`, upstream KAS baselines, dry-run evidence, three-way classification, and semantic-port update workflow; KASUPD-002 implements read-only validation/legacy-marker compatibility and KASUPD-003 implements dry-run classification/semantic-port evidence without writes
- `sot/kasrel-hermes-v016-provenance-contract.md` — KASREL-001 accepted docs/spec-only contract for Hermes v0.16 skill provenance and dependency audit vocabulary; use the SOT for detailed source-class, diagnostic, JSON-field, and follow-on task boundaries
- `sot/token-economy-and-agent-instruction-contract.md` — accepted SOT for the token-economy, English KAS product-output, repo-local agent-instruction, and project KAS lifecycle workstream; records the 10 KAS PR + 2 dependent KAH PR structure, compact console/artifact-first policy, `AGENTS.md` / `CLAUDE.md` management, color-team project KAS lifecycle gates, uninstall/vault-backup policy, project verification profiles, no-agent runner evidence, and mechanical KAH gate boundaries without authorizing implementation or rollout by itself
- `sot/multi-agent-review-policy.md` — candidate repository-promotion SOT for MAR; records that `MAR-001` is KAS docs/SOT only, paired KAH `MAREV-001` is docs/SOT only, local skill/script surfaces are implemented for fixture/mock/read-only paths, role lane readback is rooted in `registries/mar-provider-lanes.json` (`mar.role_lanes.v1`), and provider execution or KAH gate behavior requires implementation evidence

## `roadmap.md`

KAS post-KAH enablement roadmap. It uses PR-candidate task sizing and is not implementation authorization by itself. KAH 0.1.4 graph and configurable-feedback support are now evidenced through capabilities/help; CLIMVP and GRAPHMVP-001..004 KAS surfaces are implemented, while `kah graph` alias behavior and KAB runtime alignment still proceed one task at a time with evidence and review gates.

## `discussions/`

Temporary or evolving discussion material that should not be treated as final policy by itself. Items here may later be promoted into `sot/`, registries, templates, or skills after review and evidence.

- `discussions/follow-up-discussion.md`
- `discussions/feedback-for-KAH.md`
- `discussions/feedback-for-KAB.md`
- `discussions/staleclean-001-active-surface-manifest.md` — STALECLEAN-001 audit artifact mapping active stale markers to update/defer/historical dispositions; not an implementation change by itself

## Status vocabulary

| Status | Meaning |
|---|---|
| `source of truth` | Confirmed current authority for its stated scope |
| `candidate SOT` | Proposed normative record pending confirmation and/or implementation evidence |
| `planned/candidate` | Roadmap task, proposed command surface, or shorthand alias not proven implemented |
| `candidate` | Proposed KHS/KAH/KAB behavior recorded for review; not current support |
| `planned` | Expected future work with no current operational guarantee |
| `implemented` | Proven by current command/code/docs/test or effective-binary capability evidence for the exact surface claimed |
| `kah-evidenced, kas-integration-pending` | KAH advertises or proves the required deterministic surface, but KAS docs/templates/registries/skills/CLI have not yet adopted it end-to-end |
| `blocked_by_kah` | KHS can describe the desired policy, but the effective KAH binary lacks required schema, validation, proposal, audit, or compatibility support. Do not apply this label to KAH 0.1.4 `kkachi-agent-helper graph` or configurable-feedback surfaces when current capability/help evidence proves them. |
| `kab_later` | Requires KAB backend/runtime evidence and is outside current MVP |
| `unsupported` | The current toolchain cannot safely execute or validate the requested support |
| `historical` | Preserved context; not current authority by itself |
| `stale` | Known to conflict with newer evidence or decisions; preserve with marker rather than silently delete |
| `superseded` | Replaced by a named newer authority |

## Promotion rule

A discussion note becomes SOT only after the decision is reflected in the relevant `sot/` document, registry, template, or skill file. A candidate SOT becomes project authority only after responsible approver confirmation is recorded; implementation status still requires capability/help evidence.

## Decision summary

- `.kkachi-workflow.yaml` is a project-level workflow graph file when backed by KAH graph evidence.
- Graph workflow sync compatibility is owned by `sot/graph-workflow-sync-compatibility.md` and `registries/graph-workflow-sync-compatibility.yaml`: KAS v0.1.2 requires/recommends/tests KAH v0.1.9 for graph workflow sync, `doctor --project <path> --workflow-graph --json` classifies supportability without writes, first recommends KAH update when KAH is old, GRSYNC-003 implemented KAH proposal/apply orchestration for stale or broken graph repair through explicit `repair --workflow-graph --propose` and approval-gated `--apply-proposal`, and user-custom graphs remain accepted when they stay inside the supported KAS/KAH envelope.
- Task-DAG workflow policy is planned by `sot/task-dag-workflow-contract.md`: KAS owns selector rules, node agent/role/backend contracts, trigger skills, and custom workflow scaffolding; KAH owns deterministic DAG validation, node state/order enforcement, evidence, diagnostics, and gates under the paired KAH `DAGSM` SOT.
- WFLOW-004 custom workflow creation is exposed through `workflow-create` as
  dry-run-first candidate packet generation for `dag_only`, `thin_trigger`, and
  exceptional `full_trigger` modes; apply recomputes the canonical hash and
  remains fail-closed unless effective KAH capability/help evidence advertises a
  reviewed workflow catalog proposal/apply surface.
- `.kkachi/runs/<run_id>/phase-plan.yaml` remains run-local execution state/evidence for one KHS run.
- KAS owns graph templates, policy selection, phase applicability, proposal content, and skill/CLI adoption work.
- KAH owns deterministic validation/write/apply/diff/audit behavior for evidenced `kkachi-agent-helper graph` surfaces; KAH 0.1.4 also advertises configurable external-feedback intake capability, while KAS adoption remains integration-pending.
- The KHS+KAH minimum/pilot CLI lane is limited to profile-scoped install/list/doctor/sync/proposal support and does not claim KHC, Doksuri, KAB run/control, or bridge-runtime authority.
- `sot/kas-cli-contract.md` is the accepted CLIMVP-001 contract for KAS-owned profile-scoped `list`, `install --dry-run`, approved copy install, and `doctor`, and the accepted KABADOPT-001 contract for the stage selector/fallback-audit posture; implementation evidence now covers the full CLIMVP surface, while broad `proposal` remains future gated work.
- `sot/project-specific-kas-install-contract.md` is the KASPROJ-001 contract for installing separate project-specific KAS suites in one Hermes profile; generic duplicate skill names such as `kkachi-plan` are invalid for project suites, umbrella-only installs are incomplete, KASPROJ-002 dry-run planning, KASPROJ-003 approval-hash-bound install, KASPROJ-004 doctor/repair/migrate, and KASPROJ-005 project-tailored doctor policy are implemented/in-review pending gates, and KASPROJ-006 one operational application step remains planned.
- `sot/role-aware-project-suite-contract.md` is the KASROLE planning contract for role-aware project suite subsets: `blue_commander` may use the full source suite, reviewer/scribe roles must install only selected role skills, unknown `suite_role` values fail closed, source role subset manifests/doctor/repair-prune behavior is implemented through KASROLE-004, and WFLOW/DAGSM implementation should not normalize the still-unapproved Red/Orange/Gray over-installed transition state.
- `sot/project-kas-sync-state.md` is the accepted KASUPD-001 contract for project-specific KAS sync state and update workflow; KASUPD-002 validation and KASUPD-003 dry-run classification/semantic-port evidence are completed, while KASUPD-004 write-capable sync/pilot behavior remains planned.
- `sot/kasrel-hermes-v016-provenance-contract.md` is the accepted KASREL-001 contract for release-compatibility skill provenance and dependency audit vocabulary and the shared KASREL-004 guidance evidence gate; KASREL-002 implementation is completed/pre-commit-ready after the in-tree provenance inventory pass and required review gates, KASREL-003 dependency-audit implementation is completed by commit `75d0361` / event `evt-001339` / run close `evt-001340`, and KASREL-004 guidance work is in progress; this contract does not authorize profile mutation, production CLI implementation, or KAB activation; guidance changes are owned by KASREL-004.
- `sot/multi-agent-review-policy.md` records the accepted MAR design as KAS repository-planning authority while preserving implementation separation: local `kkachi-multi-agent-review`, `mar.py` fixture/mock/read-only surfaces, `registries/mar-provider-lanes.json` role-first lane readback (`mar.role_lanes.v1`), and toolchain-overlay provider proof exist, but MAR replaces no installed behavior until role coverage, adapter proof, and any KAH MAREV gates are implemented and verified.
- KAB remains backend/session/plan evidence, not graph policy authority.
- Stage 1 direct Codex runner support is defined by `sot/stage1-direct-codex-sdk-appserver-runner.md`: the KAS Python runner template must use `openai_codex` so the SDK manages `codex app-server --listen stdio://`; `codex exec`, generic `openai` SDK usage, and raw app-server transport control are not valid ordinary Stage 1 evidence.
- Kkachi v2 `.kkachi/config/workflows/` is out of KAH/KHS graph scope and has no fallback/merge relationship with `.kkachi-workflow.yaml`.

## Stale/conflict markers

Older wording that calls `phase-plan.yaml` the whole workflow SOT must be read narrowly as run-local execution state/evidence. Any `kah graph` reference in these docs is planned/candidate unless alias capability/help evidence proves implementation. `kkachi-agent-helper graph` references may be treated as implemented after effective KAH 0.1.4 capability/help checks. Configurable external-feedback bounds are KAH-evidenced in 0.1.4; KAS registries/templates/skills are updated, while operator report/e2e adoption remains `kas-integration-pending` until verified.

## Open questions

- Project-specific KAS install layout and naming are defined by `sot/project-specific-kas-install-contract.md`; KASPROJ-001 is in review/commit-ready as a docs/spec contract after local docs-contract evidence, KASPROJ-002 dry-run planning, KASPROJ-003 approval-hash-bound install, KASPROJ-004 doctor/repair/migrate, and KASPROJ-005 project-tailored doctor policy are implemented/in-review pending gates, and KASPROJ-006 one operational application step remains planned.
- Role-aware project KAS suite subset policy is defined by `sot/role-aware-project-suite-contract.md`; KASROLE-001..004 source work is implemented and committed through `7324d9c`, while real cleanup or removal of over-installed KAH development reviewer/scribe profile suites remains blocked on exact 주군 operational approval and evidence.
- Project-specific KAS static state and sync/update workflow are defined by `sot/project-kas-sync-state.md`; read-only validation is completed for KASUPD-002 and dry-run classification/semantic-port evidence is completed for KASUPD-003, while KASUPD-004 pilot/write-capable sync remains planned under the KASUPD roadmap epic.
- Hermes release compatibility, skill provenance, shadowing, and dependency audit vocabulary are defined by `sot/kasrel-hermes-v016-provenance-contract.md`; KASREL-002 implementation is completed/pre-commit-ready, KASREL-003 dependency-audit implementation is completed by commit `75d0361` / event `evt-001339` / run close `evt-001340`, and KASREL-004 is in progress as follow-on guidance work.
- KAS graph template registry format, default `kas-default` template content, and capability-checked orchestration adoption are defined by `sot/graph-template-registry.md`, `registries/graph-template-registry.yaml`, `templates/workflow-graphs/kas-default.yaml`, `sot/workflow-graph-integration.md`, and the KAS orchestration skills. Graph workflow sync compatibility, KAH v0.1.9 dependency metadata, doctor/repair/apply posture, and periodic hardening rules are defined by `sot/graph-workflow-sync-compatibility.md` and `registries/graph-workflow-sync-compatibility.yaml`.
- Stage 1 direct Codex SDK/app-server runner support is defined by `sot/stage1-direct-codex-sdk-appserver-runner.md`; CODEXSDK-001 defines/accepts the SOT, CODEXSDK-002 tracks the Python runner template, and CODEXSDK-003 wires phase-skill/template guidance without using `codex exec` or generic `openai` SDK as Stage 1 evidence.
- Exact KAS artifact mapping for KAH graph proposal paths, semantic diff, validation/explain reports, approvals/audit evidence, checksums, graph versions, audit event ids, and capability-check evidence is defined by `templates/run-artifacts/graph-evidence.md.tmpl` and summarized in final reports under `kah_graph_evidence`.
- KAB alignment requires separate KAB docs work and is non-authoritative here.

## Next record action

INITDOC is closed and its temporary transition SOT has been deleted after absorption into active records. Future KAS implementation tasks should follow `docs/roadmap.md`, update only the docs/skills/templates/registries affected by the implemented surface, preserve evidence paths, and avoid final-completion claims until required review gates pass.
