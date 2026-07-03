# MAR execution realignment SOT

Date: 2026-06-29
Owner: KAS policy and skill layer with KAH deterministic helper companion
Status: accepted source-side SOT; NEWMAR-001..011 source-side realignment tasks are completed through reviewed KAS request/facade guidance, KAH deterministic MAR execution/status/evidence support, provider-aware source controls, and audited phase-plan amend/reopen support. This SOT does not by itself authorize install, release, runtime activation, broad live-provider rollout, push, profile mutation, or auth/provider/gateway/model changes.
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
- KAH `docs/sot/multi-agent-review-evidence-gates.md` statements that KAH must not run reviewer CLIs, preserving that boundary only for historical/current MAREV-002 v1-only validation behavior unless the effective KAH binary advertises reviewed NEWMAR MAR execution capabilities for the run.
- KAH `docs/roadmap.md`, `docs/README.md`, `docs/specs.md`, `docs/compatibility.md`, and `docs/kkachi-docs-map.yaml`.
- 17thHermes 2026-06-24 MAR optimization and architecture-index SOT surfaces.

## 4. Accepted task sequence

| Task ID | Owner | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| NEWMAR-001 | KAS+KAH docs/SOT | Completed | Register this SOT and the KAH companion, update roadmaps/docs maps/indexes, and stale-mark or supersede active KAS/KAH/17thHermes surfaces that still conflict with the KAS request-bundle + KAH execution split. | Docs readback, docs-map YAML parse in KAS/KAH, KAS/KAH `git diff --check`, repo-appropriate tests, and Red/Orange/Gray/Blue review. No behavior, runtime, provider execution, install, release, push, or profile/auth/provider/gateway/model mutation. |
| NEWMAR-002 | KAS | Completed | Define reviewed MAR request-bundle schema, provider-registry/correlation schema, templates, and deterministic positive/negative fixtures for `request.yaml`, role matrix, prompt refs, input bundle, provider registry refs/checksums/schema version, lane identity, provenance/checksum metadata, author-backend correlation metadata, adapter-proof refs, approval-binding inputs, and complete `approval_binding.mar_start.bound_tuple` / `mar_start_approval_binding.bound_tuple` coverage for request-bundle refs/checksums, prompt/input refs/checksums, provider family, adapter proof, provenance, author-backend session/thread/turn refs when available, provider-preflight timestamps and checked versions, explicit execution-policy values, approval scope/created/expires/checksum, retry/waiver refs, and null/validation-only posture. reviewed schema/template/fixture evidence from NEWMAR-002 is candidate evidence for NEWMAR-003+ only. | KAS docs/schema/template tests, bundle and registry validation fixtures, docs-contract coverage, implementation color review, MAR, post-MAR second color review, and final gate passed. NEWMAR-002 blocks unsafe provider execution before KAH consumes it and preserves schema-backed fail-closed behavior for safe repo-relative refs, `sha256:<64 lowercase hex>` checksums, and missing, stale, mismatched, unsupported, expired, unsafe-ref, extra-unreviewed, and checksum-drifting metadata. |
| NEWMAR-003 | KAH | Completed | Add KAH `mar start/status/validate/wait/cancel` skeleton, request validation, machine-readable `mar` capability/version/schema advertisement, no-provider blocked receipts, status refs, and cancellation/process-ledger safety. No live provider execution. | Source-side KAH command skeleton/capability handshake and KAS fail-closed companion coverage completed; installed/runtime readback remains a separate release gate. |
| NEWMAR-004 | KAH | Completed | Add fake/no-provider async role-attempt ledger with bounded concurrency state, role coverage/status files, status hash, and watcher silence/actionable-transition proof. No live provider execution. | Source-side no-provider async ledger/status/watcher behavior completed with reviewed KAH evidence; no live provider rollout. |
| NEWMAR-005 | KAH | Completed for final/pre-commit gate | Port real provider adapter execution into KAH only after adapter proof: argv-only execution, environment allowlist, timeouts, raw-output caps, redaction/secret scan, mutation guards, adapter provenance, and provider auth/rate/quota diagnostics. | Source-side adapter safety slice passed implementation review, mandatory MAR, post-MAR second-color review, and final gate. Live provider use still requires exact approval/preflight/no-spillover evidence. |
| NEWMAR-006 | KAH with KAS schema alignment | Completed for final/pre-commit gate | Add `mar-evidence.v2` status/gate/final-gate integration while preserving reviewed `mar-evidence.v1` compatibility and explicit v1/v2 status-vocabulary mapping. | Source-side v2 gate/final-gate support passed implementation review, mandatory MAR, post-MAR second-color review, phase-plan final validation, and final gate. |
| NEWMAR-007 | KAH | Completed for final/pre-commit gate | Add callback/watcher/no-agent wake surfaces: idempotency keys, source status hash, silent unchanged `running`, compact actionable transitions, origin/thread metadata proof, rollback/no-spillover artifact refs, and no same-thread wake claim without proof. | Source-side callback/watcher/no-agent surfaces passed focused remediation, mandatory MAR, post-MAR second-color review, and final gate; same-thread wake/runtime readiness remains separately gated. |
| NEWMAR-009 | KAS+KAH diagnostics | Completed | Detect project-local copied `mar.py` and shell adapters, report stale/migration guidance, and preserve no-deletion-without-approval policy before KAS `scripts/mar.py` shrink/deprecation or in the same reviewed slice. | Source-side KAS `toolchain doctor` and KAH `diagnostics export` now expose report-only MAR migration diagnostics with no deletion/live-provider execution by default; implementation color review, mandatory MAR, post-MAR second-color review, and final gate completed. |
| NEWMAR-008 | KAS | Completed via NEWMAR-010 source slice | Shrink/deprecate KAS `scripts/mar.py` toward request/render/facade duties; KAS script no longer executes providers in the healthy path after KAH `mar` is implemented and after NEWMAR-009 diagnostics have landed or are co-reviewed. | KAS script/docs/tests now prove facade/request rendering and migration guidance; legacy `provider-attempt` fails closed with `kah_mar_execution_required` and does not count as role coverage. |
| NEWMAR-011 | KAH+KAS loop-state safety | Completed for final/pre-commit gate | Add an audited KAH phase-plan amend/reopen path for repeated feedback/MAR loops: explicit from/to status, kind (`reopen`, `correction`, `supersede`, or `rollback`), non-empty reason, evidence/approval refs, previous/current evidence refs, prior-evidence preservation, and deterministic event payloads. KAS/Blue/color authority still decides semantic acceptance. | KAH CLI/unit tests and KAS docs-contract guidance completed for stale `from_status`, missing reason, unsafe evidence refs, workflow/final-gate safeguards, help/readback coverage, and audited event trails. |
| NEWMAR-010 | Controlled source-side provider-aware pilot/evidence slice | Completed source-side; runtime rollout held | First live-provider pilot only after explicit approval, provider preflight, adapter proof, bounded concurrency/duration/cost, baseline token/wall-clock evidence, rollback/no-spillover proof, no-write/no-spillover scan, no auth/provider/profile/gateway/model mutation, and NEWMAR-011 completed or explicitly not in scope for repeated feedback/MAR loops. | Source-side KAS trigger-facade shrink and KAH provider-aware execution controls completed and verified. This is not broad rollout: installed/runtime release, PATH/toolchain refresh, live production provider readiness, tag, push, and auth/provider/profile/gateway/model mutation remain separate approvals. |

KAH NEWMAR-003+ may consume only reviewed NEWMAR-002 schema/capability evidence and must reject arbitrary, generated, or otherwise unreviewed metadata. When approval metadata is stale, mismatched, expired, or checksum-drifting, operators must regenerate the affected packet or referenced artifact, rerun review, bind a new exact approval tuple, or remain in validation-only posture; fallback/default provider execution is forbidden. Full Draft 2020-12 schema consumption remains outside the reviewed compatibility subset unless separately evidenced; NEWMAR-002 proves the reviewed KAS compatibility subset and fails closed within that subset.

## 4.1. Post-NEWMAR V01CLEAN follow-up

V01CLEAN is the planned cleanup after source-side NEWMAR: remove v0.1.x, KAS `scripts/mar.py`, copied `.kkachi/scripts/mar.py`, shell MAR adapters, and interpreter-wrapper MAR paths from active operator-facing KAS/KAH skills, docs, help, install examples, and profile-suite guidance. The only allowed remaining uses of those surfaces are fail-closed diagnostics, rejected fixtures, and negative tests. V01CLEAN must not introduce compatibility, fallback, warning-only, or migration paths for v0.1.x or any legacy MAR surface, and it does not authorize profile-local cleanup/deletion/mutation, installed skill mutation, commit/release/install/runtime/live-provider rollout, or auth/provider/gateway/model changes.

No dedicated V01CLEAN cleanup SOT has been approved yet. Until one is approved, this follow-up section and the KAS/KAH roadmap V01CLEAN entries are interim planning authority only; the accepted source-side completion claims in this SOT remain limited to NEWMAR-001..011.

Execution order: V01CLEAN-001 (KAS stale-surface removal) -> V01CLEAN-002 (KAH rejection/capability hardening) -> V01CLEAN-003 (cross-repo closeout). Older NEWMAR wording about "migration guidance" or "migration risk" is historical stale-risk reporting only; V01CLEAN must remove or stale-mark operator-facing migration instructions instead of preserving them as a supported path.

| Task ID | Owner | Status | Acceptance criteria | Evidence and review gates |
|---|---|---|---|---|
| V01CLEAN-001 | KAS | Planned | Remove v0.1.x active skills/operator guidance and remove KAS Python/shell/interpreter-wrapper MAR surfaces from supported paths; keep only fail-closed detection and negative tests. | KAS stale-surface audit, skill/docs/readback diff, focused tests including negative fail-closed coverage for v0.1.x/Python/shell/interpreter-wrapper rejection, `git diff --check`, Red/Orange/Gray review, and Blue synthesis. |
| V01CLEAN-002 | KAH | Planned | Make KAH Go-native `go_native_argv` the only supported MAR execution target and reject Python/shell/legacy adapter/interpreter-wrapper paths before spawn. | KAH rejection tests for Python/shell/legacy/interpreter-wrapper paths, capability/help/readback audit, diagnostics evidence, `git diff --check`, repo tests, Red/Orange/Gray review, and Blue synthesis. |
| V01CLEAN-003 | KAS+KAH | Planned | Cross-repo docs/skills closeout so v0.2 is the only active path and all remaining legacy references are fail-closed/negative-test only. | Cross-repo stale-term scan, docs readback, KAS/KAH tests, `git diff --check`, four-color cleanup review, and final Blue report. |

## 5. Non-goals and holds

This SOT does not authorize further implementation, live provider execution, installed runtime readiness, same-thread wake readiness, profile/auth/provider/gateway/model mutation, KAB activation, version bump, release, tag, push, or deletion of local copied MAR scripts/adapters by itself. Completion of NEWMAR may feed a later KAS/KAH v0.2.1 readiness pass only after separate approval.

## 6. KAT relationship

KAT remains a deterministic test/log evidence producer. NEWMAR uses KAT evidence for `test_adequacy` only through KAS/KAH-declared refs. No KAT roadmap task is opened unless a later KAS/KAH implementation or pilot proves that current KAT status/summary/raw-log artifacts are insufficient.
