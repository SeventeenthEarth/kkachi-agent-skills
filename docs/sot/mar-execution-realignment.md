# MAR execution realignment SOT

Date: 2026-06-29
Owner: KAS policy and skill layer with KAH deterministic helper companion
Status: accepted source-side planning SOT; `NEWMAR-001` is docs/SOT and roadmap registration only and does not implement KAH `mar` behavior, live provider execution, install, release, runtime activation, profile mutation, or auth/provider/gateway/model changes
Authority level: KAS-side planning authority for moving MAR provider execution from the older KAS-owned runner direction to KAS request bundles plus KAH deterministic execution/evidence
Scope: KAS docs, request-bundle contracts, role/provider policy declarations, prompt refs, Blue disposition templates, KAS `scripts/mar.py` migration guidance, and cross-repo roadmap sequencing
Canonical external SOT: `/Users/draccoon/Workspace/Hermes/17thHermes/40_outputs/team/hwangchung/kkachi/2026-06-29-kas-kah-mar-execution-realignment-sot.md`
Paired KAH companion SOT: KAH `docs/sot/mar-execution-realignment.md`
Related docs: `docs/sot/multi-agent-review-policy.md`, `docs/sot/mar-task-loop-contract.md`, `docs/roadmap.md`, KAH `docs/sot/multi-agent-review-evidence-gates.md`, KAH `docs/sot/mar-execution-realignment.md`
Evidence/source paths:
- First-round review: Red `t_7298c234` REQUEST_CHANGES, Orange `t_1cd9c7e5` APPROVE_WITH_CHANGES, Gray `t_4341d8e9` APPROVE_WITH_CHANGES, Blue `t_96a5d6cf` incorporated conditions.
- Second-round confirmation: Red `t_920316a4` APPROVE, Orange `t_b04da6bd` APPROVE, Gray `t_0b0bd466` APPROVE, Blue `t_7ac47555` FINAL_FOUR_COLOR_SOURCE_SIDE_AGREEMENT.
- Adoption-boundary consultation: Red `t_2ba94fa9`, Orange `t_e692671a`, Gray `t_35fbde6d`, Blue synthesis `t_e8a88baa` agreed on `ADOPT_WITH_TASK_BOUNDARY_CHANGES`.
- GJC advisory functional-plan review: `ADOPT_WITH_CHANGES`; accepted improvements are explicit provider-registry/correlation schema in NEWMAR-002, KAH capability handshake in NEWMAR-003, watcher rollback/no-spillover criteria in NEWMAR-007, and NEWMAR-009 diagnostics before NEWMAR-008 shrink/deprecation.
- Final adoption color review: Red `t_a89535ff` APPROVE, Orange `t_00c79343` APPROVE_WITH_CHANGES, Gray `t_f3b3de3f` APPROVE_WITH_CHANGES, and Blue synthesis `t_753b4445` `ADOPT_WITH_REQUIRED_TRACEABILITY_CHANGES`; incorporated changes include NEWMAR-011 summary coverage and the NEWMAR-011 pre-live-pilot gate.

## 1. Decision

NEWMAR adopts the following KAS/KAH split for the next MAR architecture:

```text
KAS owns MAR semantics, role matrix, provider-policy declarations, request bundles, prompt refs, retry/waiver/escalation policy, and Blue disposition templates.
KAH owns deterministic MAR request validation, authorized execution, async role attempts, process/status evidence, role coverage facts, merge/status artifacts, and watcher/callback surfaces.
```

This amends the older 2026-06-24 Hermes-native MAR optimization direction only for provider-execution ownership. The earlier token-economy, async, role-first, artifact-first, provider-independence, and fail-closed goals are retained. The older statement that the healthy long-term provider runner is KAS-plugin-owned is superseded after `NEWMAR-001` updates or stale-marks active repository docs.

## 2. Authority boundaries

KAS must not treat KAH status, provider wording, GJC output, KAT summaries, or model consensus as final review acceptance. Blue disposition and triggered Red adjudication remain authority.

KAH must not choose roles, providers, alternates, retries, waivers, Red escalation, or final pass/fail acceptance. KAH may execute only a KAS-declared request after the exact approval/preflight/safety evidence is present.

KAT remains factual test evidence only. No KAT source change is part of NEWMAR unless a later task proves KAS/KAH cannot consume existing KAT status/summary/raw-log artifacts for `test_adequacy` evidence.

## 3. Supersession and stale-marker targets

`NEWMAR-001` must update or stale-mark active surfaces before later workers treat KAH provider execution as source authority:

- KAS `docs/sot/multi-agent-review-policy.md` statements that KAS owns long-term provider execution, parsing, merge-pack creation, status aggregation, and role coverage as the healthy path.
- KAS `docs/sot/mar-task-loop-contract.md` statements that MARTL is KAS-only/no-new-KAH by default, preserving that statement only for the pre-NEWMAR task-loop alignment scope.
- KAS `docs/roadmap.md`, `docs/README.md`, and `docs/kkachi-docs-map.yaml`.
- KAH `docs/sot/multi-agent-review-evidence-gates.md` statements that KAH must not run reviewer CLIs, preserving that boundary for current `mar-evidence.v1` behavior until NEWMAR implementation lands.
- KAH `docs/roadmap.md`, `docs/README.md`, `docs/specs.md`, `docs/compatibility.md`, and `docs/kkachi-docs-map.yaml`.
- 17thHermes 2026-06-24 MAR optimization and architecture-index SOT surfaces.

## 4. Accepted task sequence

| Task ID | Owner | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| NEWMAR-001 | KAS+KAH docs/SOT | Completed | Register this SOT and the KAH companion, update roadmaps/docs maps/indexes, and stale-mark or supersede active KAS/KAH/17thHermes surfaces that still conflict with the KAS request-bundle + KAH execution split. | Docs readback, docs-map YAML parse in KAS/KAH, KAS/KAH `git diff --check`, repo-appropriate tests, and Red/Orange/Gray/Blue review. No behavior, runtime, provider execution, install, release, push, or profile/auth/provider/gateway/model mutation. |
| NEWMAR-002 | KAS | Planned | Define reviewed MAR request-bundle schema, provider-registry/correlation schema, and templates: `request.yaml`, role matrix, prompt refs, input bundle, provider registry refs/checksums/schema version, lane identity, provenance/checksum metadata, author-backend correlation metadata, adapter-proof refs, approval-binding inputs, concurrency/duration/cost policy refs, independence policy, and execution-safety refs. | KAS docs/schema/template tests, bundle and registry validation fixtures, docs-contract coverage, and color review before KAH consumes the schema. |
| NEWMAR-003 | KAH | Planned | Add KAH `mar start/status/validate/wait/cancel` skeleton, request validation, machine-readable `mar` capability/version/schema advertisement, no-provider blocked receipts, status refs, and cancellation/process-ledger safety. No live provider execution. | KAH unit/CLI/e2e tests for malformed request, missing registry, absent approval/preflight, cross-run refs, `wait` timeout, `cancel`, no-provider blocked status, and KAS fail-closed behavior when the effective KAH binary lacks matching reviewed `mar` capability evidence. |
| NEWMAR-004 | KAH | Planned | Add fake/no-provider async role-attempt ledger with bounded concurrency state, role coverage/status files, status hash, and watcher silence/actionable-transition proof. No live provider execution. | KAH tests for role-first state, uncovered required roles, hash-idempotent watcher output, and fail-closed diagnostic coverage. |
| NEWMAR-005 | KAH | Planned | Port real provider adapter execution into KAH only after adapter proof: argv-only execution, environment allowlist, timeouts, raw-output caps, redaction/secret scan, mutation guards, adapter provenance, and provider auth/rate/quota diagnostics. | Per-provider adapter tests/proof, mutation/no-secret scans, fail-closed status tests, and Red risk review. |
| NEWMAR-006 | KAH with KAS schema alignment | Planned | Add `mar-evidence.v2` status/gate/final-gate integration while preserving reviewed `mar-evidence.v1` compatibility and explicit v1/v2 status-vocabulary mapping. | KAH schema/gate/final-gate tests, KAS compatibility fixtures, diagnostics for v1 historical/current vs v2 planned/implemented states. |
| NEWMAR-007 | KAH | Planned | Add callback/watcher/no-agent wake surfaces: idempotency keys, source status hash, silent unchanged `running`, compact actionable transitions, origin/thread metadata proof, rollback/no-spillover artifact refs, and no same-thread wake claim without proof. | Watcher/callback tests, no-spillover/rollback refs, origin/thread metadata proof, and operator-readable status examples. |
| NEWMAR-009 | KAS+KAH diagnostics | Planned | Detect project-local copied `mar.py` and shell adapters, report stale/migration guidance, and preserve no-deletion-without-approval policy before KAS `scripts/mar.py` shrink/deprecation or in the same reviewed slice. | Diagnostic fixtures/readback and docs guidance. No local deletion by default. |
| NEWMAR-008 | KAS | Planned | Shrink/deprecate KAS `scripts/mar.py` toward request/render/facade duties; KAS script no longer executes providers in the healthy path after KAH `mar` is implemented and after NEWMAR-009 diagnostics have landed or are co-reviewed. | KAS script/docs/tests proving facade/request rendering and migration guidance; no local copied runner fallback claim. |
| NEWMAR-011 | KAH+KAS loop-state safety | Planned | Add an audited KAH phase-plan amend/reopen path for repeated feedback/MAR loops: explicit from/to status, kind (`reopen`, `correction`, `supersede`, or `rollback`), non-empty reason, evidence/approval refs, previous/current evidence refs, prior-evidence preservation, and deterministic event payloads. KAS/Blue/color authority still decides semantic acceptance. | KAH CLI/unit/e2e tests for completed `request-feedback-*` reopen, stale `from_status` fail-closed, missing reason fail-closed, unsafe evidence ref rejection, workflow-projection/final-gate safeguards, and KAS docs/skill guidance readback. |
| NEWMAR-010 | Controlled pilot | Planned | First live-provider pilot only after explicit approval, provider preflight, adapter proof, bounded concurrency/duration/cost, baseline token/wall-clock evidence, rollback/no-spillover proof, no-write/no-spillover scan, no auth/provider/profile/gateway/model mutation, and NEWMAR-011 completed or explicitly not in scope for repeated feedback/MAR loops. | Red/Orange/Gray/Blue review, current baseline comparison, KAH status/merge refs, Blue disposition, audited reopen-state proof when applicable, and explicit report that source-side acceptance did not authorize broad rollout. |

## 5. Non-goals and holds

This SOT does not authorize implementation by itself. It does not authorize live provider execution, installed runtime readiness, same-thread wake readiness, profile/auth/provider/gateway/model mutation, KAB activation, version bump, release, tag, push, or deletion of local copied MAR scripts/adapters. Completion of NEWMAR may feed a later KAS/KAH v0.2.1 readiness pass only after separate approval.

## 6. KAT relationship

KAT remains a deterministic test/log evidence producer. NEWMAR uses KAT evidence for `test_adequacy` only through KAS/KAH-declared refs. No KAT roadmap task is opened unless a later KAS/KAH implementation or pilot proves that current KAT status/summary/raw-log artifacts are insufficient.
