# V02FLOW — KAS v0.2 GJC workflow train correction SOT

Date: 2026-07-04
Owner: KAS workflow/policy layer
Confirming role: Hwangchung / 황충, Kkachi Blue commander
Status: source-side SOT; V02FLOW-013 executor-loop contract correction accepted source-side; V02FLOW-014 KAH executor-loop driver/fail-closed evidence accepted source-side with final gate PASS for `run-20260708T014514Z-832321c87e08`; V02FLOW-015 cross-repo fixture/e2e proof is source-side aligned; V02FLOW-016 capability/HOME/deferred-feedback/MAR-waiver proof is source-side aligned; V02FLOW-017..018 v0.2.1 review/release train remains separately gated by task evidence
Authority level: KAS-side planning authority for v0.2 workflow-stage, GJC status semantics, implementation executor-loop, implementer-owned mutation phases, review-feedback remediation, review train, and watcher policy corrections
Scope: `kkachi-agent-skills` source docs, registries, templates, and skills; paired KAH companion SOT is `kkachi-agent-helper/docs/sot/v020-gjc-workflow-train-corrections.md`
Related docs: `docs/roadmap.md`, `docs/sot/gajae-delegated-execution-contract.md`, `docs/sot/task-dag-workflow-contract.md`, `docs/sot/strict-workflow-execution-contract.md`, `docs/sot/mar-task-loop-contract.md`, KAH `docs/sot/gajae-gjc-wrapper-evidence.md`, KAH `docs/sot/v020-gjc-workflow-train-corrections.md`
External evidence: `/Users/draccoon/Workspace/Hermes/17thHermes/40_outputs/team/hwangchung/kkachi/2026-07-04-kas-kah-v020-deepinterview-ralplan-ultragoal-aggregated-review-watcher-notes.md`

## 1. Decision

KAS/KAH v0.2 keeps the broad GJC delegated-execution direction, but the active workflow language must stop overstating what current KAH/GJC calls prove.

Canonical corrected train:

```text
plan stage:
  plan -> ralplan
  after ralplan completes, route directly into impl stage
  ask is not a normal phase or normal step

impl stage:
  impl -> test enhance -> ai slop cleaner -> optimize -> docs update
  any code-modifying step runs the touched repo/scope test command from docs/kkachi-docs-map.yaml test_commands

review stage:
  color review -> MAR -> 2nd color adoption
```

Approval or decision evidence still exists at real approval boundaries, but KAS must not model an always-present `ask` phase between planning and implementation.

## 2. GJC status semantics

### 2.1 `deep-interview`

`deep-interview` is explicit-request only. It is not a normal task step.

Use it for epic/design shaping, major architecture ambiguity, emergency/confusing repair analysis, or high-ambiguity requirements. Otherwise record it as not applicable with reason `explicit_request_not_present`.

Current KAH source evidence shows a stale native adapter shape can call:

```text
gjc deep-interview --packet <packet> --json
```

Current native GJC help expects mode/write-style inputs such as `--quick|--standard|--deep` or `--write --stage final --slug ... --spec ...`. KAS must not claim the current KAH wrapper proves end-to-end native deep-interview readiness until KAH has matching adapter evidence.

### 2.2 `ralplan`

Current KAH `start-ralplan` maps packet data to native:

```text
gjc ralplan --write --stage <stage> --stage_n <stage_n> --artifact <artifact> --json
```

This records/persists an existing artifact. It does not prove a full Planner / Architect / Critic consensus loop generated that plan during the KAH call.

KAS status language should distinguish:

```text
ralplan_candidate_recorded   # primary KAS candidate/record status
ralplan_recorded             # technical/readback alias
ralplan_consensus_seeded
ralplan_consensus_complete
```

`ralplan_ready` is legacy/current-helper shorthand only. Treat it as candidate plan evidence until KAS/Blue/color plan review accepts and locks it.

### 2.3 `ultragoal`

Current KAH `start-ultragoal` invokes:

```text
gjc ultragoal create-goals --brief-file <run-local-brief> --json
```

That produces a goal bundle / ledger evidence. It does not prove source mutation, implementation diff readiness, or review-ready implementation.

Preferred status language:

```text
implementation_goal_bundle_ready  # primary operator-facing goal-bundle status
ultragoal_goals_ready             # technical/backend alias
gjc_goal_bundle_ready             # technical/backend alias
```

`ultragoal_ready` is legacy/current-helper shorthand only. Reserve `implementation_diff_ready`, `implementation_candidate_ready`, and `implementation_verified` for evidence after an actual executor loop and verification.

## 3. Missing implementation executor loop

A real delegated implementation loop must include the V02FLOW-005 executor-loop policy:

```text
create-goals -> complete-goals -> execute-goal -> checkpoint -> verify -> repeat-or-terminate
```

`complete-goals` freezes the selected goal bundle only; it is not implementation completion, implementation acceptance, verification, MAR, second-color, final, or commit readiness.

`goal_bundle.status`, `goal_bundle.goals[].status`, and `executor_candidate.candidate_status` must reject `accepted`.

A top-level packet `status: accepted`, if retained, is a derived/read-only post-final-gate alias of the dedicated gated `acceptance` object only.

KAS must not silently fill the missing loop by letting Blue patch source directly. Default boundary:

```text
Blue commands and verifies. Executor implements.
```

Allowed exceptions require explicit recorded evidence, such as 주군-approved Blue direct patch, emergency hotfix, docs-only/trivial correction, or an approved backend-unavailable exception. Blue direct patch exception records must include reason, scope, why executor loop was not used, changed-surface refs, verification required, review required, MAR required, final gate required, and waiver ref. These exceptions do not authorize live/runtime/auth/provider/gateway/profile mutation, commit, push, release, install, or scope expansion.

### 3.1 GJC-owned implementer phases and feedback remediation

V02FLOW-009 extends the executor-loop policy from a generic `impl` step to every mutation-capable implementer phase:

```text
impl -> test-enhance -> ai-slop-cleaner -> optimize -> docs-update
```

The selected implementer lane, normally GJC `ultragoal` plus the executor loop, owns repository mutation for all five phases. Blue commands, scopes, approves, and verifies; Red/Orange/Gray review; KAH records deterministic evidence. Direct Blue or color-lane editing remains an explicit recorded exception, not the default Kkachi strategy.

Current GJC bundled workflow skills are class-level lanes such as `deep-interview`, `ralplan`, `team`, and `ultragoal`. There is no separate native GJC `ai-slop-cleaner` or `remove-ai-slop` command/skill. KAS must therefore compose phase-specific prompt/packet contracts and KAH must materialize them as run-local briefs for the GJC executor lane.

KAS phase prompt contracts must include at least:

- phase id and approved run/task scope;
- canonical phase-id and alias mapping across KAS graph ids, KAH run-local ids, GJC prompt ids, and expected artifact names;
- ralplan artifact ref/hash, Blue plan-lock or approval ref, accepted-scope ref, and selected ultragoal session/goal id when available;
- changed-file or approved surface bounds;
- behavior, API, test, acceptance-criteria, and reviewed-intent preservation requirements;
- non-goals: no new feature work, no fallback expansion, no broad redesign, no dependency addition unless explicitly approved;
- expected evidence outputs and verification commands;
- stop/blocked conditions that require Blue or 주군 decision.

Unknown phase ids or aliases must fail closed. Until V02FLOW-010 supplies KAH capability/readback for a phase dispatch shape, KAS guidance must treat executor dispatch for that shape as unsupported rather than asking Blue/color lanes to patch directly. The intended mapping is:

| Intent | KAS graph/operator id | KAH run-local phase id | GJC prompt phase id | Primary artifact expectation |
|---|---|---|---|---|
| implementation | `implement` / `impl` | `impl` | `impl` | `impl-log.md`, diff refs, verification refs |
| test enhancement | `enhance-test` / `test-enhance` | `test-enhance` | `test-enhance` | `test-log.md`, focused regression refs |
| AI slop cleanup | `ai-slop-cleaner` | `ai-slop-cleaner` | `ai-slop-cleaner` | `slop-cleanup-log.md`, cleanup plan, verification refs |
| optimization | `optimize` | `optimize` | `optimize` | `optimize-log.md`, behavior-preservation refs |
| docs update | `docs` / `docs-update` | `docs-update` | `docs-update` | `docs-update.md`, changed-doc refs |
| feedback remediation | `handle-feedback` / `handle-feedback-*` | `handle-feedback-*` | `handle-feedback-*` | finding bundle, remediation log, focused re-review or MAR-refresh refs |

`ai-slop-cleaner` prompt contracts should use a regression-safe, deletion-first cleanup pattern: protect behavior first, write a cleanup plan before editing, classify slop before editing, run one smell-focused pass at a time, prefer deletion/reuse over additions, and close with `slop-cleanup-log.md`, changed files, simplifications, behavior-lock/verification evidence, and remaining risks. Smell classes include duplicate logic, dead code, needless wrappers or abstractions, boundary leaks, weak regression coverage, over-explaining comments, speculative branches, ungrounded/generated prose, and UI/design defaults only when the task has an applicable UI/UX surface.

`test-enhance` should add or strengthen focused regression/fail-closed coverage through the same implementer lane, then run the selected touched-scope verification. `optimize` should perform bounded behavior-preserving simplification only after tests protect behavior. `docs-update` may let GJC draft or edit task-scoped docs, but SOT/roadmap lifecycle, authority decisions, final disposition, and evidence acceptance remain Blue/Gray/color-gated.

Review and feedback loops use the same mutation boundary. If first color review, project-Gray integrity review, MAR, second color adoption, or a later feedback round finds required changes, KAS must package the exact findings, evidence refs, accepted/rejected scope, and focused verification expectations into a remediation prompt for the selected implementer lane. A remediation finding bundle must carry finding id, source lane, severity/blocker state, Blue disposition, accepted scope, required reopened/amended phases, required verification, required focused re-review or MAR-refresh path, and close evidence refs. `handle-feedback-*` closes only after executor-loop evidence, independent verification, and the required focused re-review or MAR refresh path exist.

### 3.2 Deferred review and MAR feedback ledger

V02FLOW-011 adds a first-class deferred-feedback ledger for non-blocking findings that should not expand the current task. KAS docs-map must expose `deferred_feedback: "docs/deferred-feedback.md"`, and `docs/deferred-feedback.md` must preserve review/MAR/Blue provenance for intentionally deferred work.

Blockers must not be deferred. A finding that blocks current acceptance, final gate, MAR passability, authority boundaries, safety posture, or task acceptance criteria must stay in the current remediation loop or leave the task held. Defer is only allowed for real but non-blocking issues where the Blue disposition records that the change is too large for the current slice, belongs with a near follow-up task, requires separate design, or is an opportunistic refactor/cleanup.

Each deferred entry must record created date, owner lane, last-reviewed date, status-change evidence, originating task/run, source phase, source lane/reviewer, source ref, finding summary, why it matters, `Blocking current task: false`, why not fixed now, defer reason, proposed fix, suggested follow-up task, conversion fields, target repos/files, required and carried-forward gates when resumed, final-report refs, Blue disposition, non-empty Blue disposition ref, additional approval/ref when applicable, and terminal evidence/rationale for resolved/rejected/stale states. Current final reports must link deferred entries whenever any accepted feedback is intentionally not fixed now. Blue disposition ref is required and N/A is invalid. Blocker finding handling: fix_now_or_hold_current_task. The lifecycle vocabulary is exactly `open`, `converted_to_task`, `resolved`, `rejected`, and `stale`; waiver is authority evidence, not lifecycle status. The ledger must distinguish logical/readback status hash distinct from byte-level ledger file SHA, and every entry must preserve deterministic converted-task ref, final-report reciprocal ref, and source finding ref semantics.

## 4. Review train and watcher policy

V02FLOW-007 review-train and aggregate-watcher policy: substantial development work must preserve `first color review -> mandatory MAR -> second color adoption/review -> Blue disposition` as distinct authority gates. In fully expanded operator terms this means first official Red/Orange/Gray color review plus dependent Blue synthesis, mandatory MAR for development work unless an explicit task-specific waiver is recorded before the MAR gate, and second official Red/Orange/Gray color adoption/review plus dependent Blue final disposition.

For substantial KAS/KAH implementation work, review proceeds:

```text
implementation cleanup/verification
  -> color review
  -> MAR
  -> 2nd color adoption
  -> Blue final disposition / next gate
```

This train is feasible under 주군 approval structure if new approval boundaries stop for approval.

Official color review remains Kanban-based Red / Orange / Gray. `delegate_task`, temporary subagents, local helper critique, informal Discord prose, or unpersisted model critique may support pre-review QA but do not replace official review evidence.

Watchers should be aggregated per color-review round:

```text
one aggregate watcher for a color round:
  watches Red card
  watches Orange card
  watches Gray card
  optionally watches KAH review/run status
```

A full train may use one aggregate watcher for first color review and one aggregate watcher for 2nd color adoption. MAR may use its own provider/status watcher if async. Each one aggregate watcher per color round is state-report-only.

The watcher must report all-done or blocked states only. It must not perform Blue synthesis, inject or fake `진행해`, auto-trigger continuation, waive lanes, mutate source, substitute temporary subagents or self-approval for official review, commit, push, install, release, or change runtime/auth/provider/gateway/profile/model state.

## 5. Planned implementation split

Recommended logical sets: 6.

1. Workflow/stage contract migration: remove `ask` as a normal step, normalize `plan -> ralplan -> impl`, and encode the corrected impl/review train.
2. GJC adapter/status semantics: deep-interview explicit-only/stale `--packet`, ralplan record-only versus consensus, ultragoal goal-bundle readiness.
3. Ultragoal executor loop: add/describe `create-goals -> complete-goals -> execute-goal -> checkpoint -> verify -> repeat-or-terminate`, and prevent Blue direct patching from silently filling implementation gaps.
4. Review train + aggregated watcher: color review -> MAR -> 2nd color adoption, one aggregated watcher per color round, no auto Blue synthesis.
5. GJC-owned implementer and remediation phases: KAS V02FLOW-009 defines phase prompt/packet contracts for `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, `docs-update`, and `handle-feedback-*`; KAH V02FLOW-010 materializes and records execution/remediation evidence through the GJC executor loop.
6. Deferred review/MAR feedback ledger: KAS V02FLOW-011 defines the `deferred_feedback` docs-map field, `docs/deferred-feedback.md` template, no-blocker-defer policy, provenance fields, and final-report linkage; KAH V02FLOW-012 validates/readbacks the ledger and prevents hidden blocker deferral.

V02FLOW-001..012 used paired KAS/KAH physical source slices for the six logical corrections. For the v0.2.1 continuation, use the task ids `V02FLOW-013` through `V02FLOW-018` as the operational labels rather than PR1/PR2-style labels.

### 5.1 V02FLOW-013..018 v0.2.1 remediation train

The V02FLOW completion audit on 2026-07-08 confirmed that V02FLOW-001..012 established the correction contracts and readback surfaces, but also found one release-critical semantic gap: current `start-ultragoal`/`ultragoal create-goals` proves goal-bundle readiness only. It must not be treated as source mutation, diff readiness, verification, review, MAR, final, commit, install, or release readiness.

V02FLOW therefore continues through six focused v0.2.1 tasks:

| Task ID | Primary owner | Purpose | Required boundary |
|---|---|---|---|
| V02FLOW-013 | KAS | Correct KAS contract/guidance/tests so goal-bundle readiness cannot close implementation and all mutation-capable phases require executor-loop evidence. | MAR remains standing-waived until official v0.2.1 Go/KAH MAR readiness; no provider-MAR claim. |
| V02FLOW-014 | KAH | Implement or fail-closed the real ultragoal executor-loop driver/evidence path that produces diff/checkpoint/verification evidence instead of stopping at `create-goals`. | Large full-development slice; real-user HOME and JSON-only output must be proved. |
| V02FLOW-015 | KAS+KAH | Prove `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, and `docs-update` through e2e/fixture evidence and negative goal-bundle-only tests. | Cross-repo verification; no phase can close from goal-bundle evidence alone. |
| V02FLOW-016 | KAH+KAS | Capture capability, HOME, deferred-feedback, legacy-MAR rejection, standing-waiver, and second-color-N/A proof. | Do not execute or claim provider MAR before the official v0.2.1 MAR path. |
| V02FLOW-017 | KAS-led review | Run official Red/Orange/Gray review plus Blue synthesis for V02FLOW-013..016 correction evidence. | Temporary agents or `delegate_task` are not official color review. |
| V02FLOW-018 | KAS+KAH | Assemble the v0.2.1 source release-readiness package after 013..017: fresh tests, roadmap/SOT alignment, capability/HOME/deferred evidence, waiver statement, and release notes/readiness. | V02FLOW-018 is readiness packaging only; release/tag/push/install/runtime/provider/auth/profile/gateway/model changes still require explicit separate approval. |

After V02FLOW-018, the intended operator path is to submit the v0.2.1 release-readiness package for separate 주군 publication approval and move additional polish/hardening into post-release tasks rather than expanding this remediation train indefinitely.

### 5.2 V02FLOW-013 executor-loop completion status hierarchy

V02FLOW-013 defines the KAS status hierarchy that separates goal-bundle readiness from implementation completion:

- `implementation_goal_bundle_ready: goal bundle only; never sufficient for implementation completion`. This status may be emitted when `ultragoal create-goals` / `start-ultragoal` has produced a candidate goal bundle, but it is not source mutation, diff readiness, verification, review, MAR, final, commit, install, or release readiness.
- `implementation_diff_ready: executor-loop source diff, checkpoint, and checksum evidence is present`. This status may be used only after the selected executor loop has produced changed source refs or an explicit no-change rationale, diff refs, checkpoint ref/status, and checksum evidence; verification may still be pending.
- `implementation_verified: executor-loop diff, checkpoint, HOME, checksum, termination, and verification-output evidence passed`. This status requires post-change verification output refs and a terminal executor-loop result, and it still does not bypass color review, MAR applicability/waiver, final gate, commit, install, runtime, or publication approval.

Required executor-loop evidence fields: changed_source_refs, diff_refs, checkpoint_ref, checkpoint_status, verification_output_refs, checksums, termination_reason, HOME, no_authority_boundaries.

The implementer-owned mutation phases `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, and `docs-update` must not complete from goal-bundle-only evidence. Their KAS prompt/packet/registry surfaces must require `implementation_diff_ready` or `implementation_verified` evidence, or fail closed with an explicit unsupported-capability or approved direct-edit exception. `implementation_goal_bundle_ready is goal-bundle-only and never sufficient for implementation completion` is the shared wording reviewers should use when auditing these phases.

### 5.3 V02FLOW-015 mutation-phase fixture/e2e proof

V02FLOW-015 extends the status hierarchy into explicit per-phase proof for `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, and `docs-update`. For these mutation-capable phases, completion evidence must name the selected executor lane `gjc_ultragoal_executor_loop_candidate`, the requested phase id, the canonical phase id, argv/process refs, real-user `HOME`, cwd, timestamps, exit code, changed-file refs, non-empty diff refs, checkpoint ref/status/timestamp, verification command/status/output refs, checksums, repeat/termination reason, and candidate-only authority boundaries.

Known workflow-projection aliases must map to the same executor-loop requirement instead of bypassing it: `impl`/`implement`, `test-enhance`/`test_enhance`/`enhance-test`/`enhance_test`, `ai-slop-cleaner`/`ai_slop_cleaner`/`slop_cleanup`, and `docs-update`/`docs_update`/`update_docs`/`docs`.

Goal-bundle-only evidence, existence-only workflow projection, missing selected lane, missing HOME/cwd/process/timestamps/exit evidence, missing changed refs, empty diff refs, missing checkpoint, missing verification output, missing SHA256/checksum integrity, stale run/ref/checksum/timestamp/artifact refs, unsupported-capability warning-only behavior, or GJC/KAT/Blue/MAR/review/final authority overclaim must fail closed. `docs-update` remains mutation-capable in this train; a no-change rationale does not substitute for required executor-loop evidence and non-empty diff refs for these five phases.

### 5.4 V02FLOW-016 release-evidence proof

V02FLOW-016 turns the release-evidence slice into explicit source-side proof requirements. Before V02FLOW-018 release-readiness packaging, KAS/KAH evidence must capture the effective helper `capabilities --json` output and verify at least `gjc_executor_loop_evidence=true`, `diagnostics_deferred_feedback=true`, `mar_legacy_rejection_diagnostics=true`, `mar_provider_adapter_safety=true`, and `mar_migration_diagnostics=false` from the binary actually selected for the run.

Real-user HOME proof is mandatory for both terminal/GJC execution and any future provider process execution. KAH GJC evidence must show `HOME=/Users/draccoon`; KAH MAR provider execution evidence must show `toolchain.operator.real_user_home` as the HOME source and reject Hermes profile homes such as `.hermes/profiles/...`. Missing HOME proof, ambient Hermes-profile HOME, warning-only HOME diagnostics, or source-checkout-only capability claims without effective-binary readback fail closed.

Deferred-feedback proof must include `diagnostics deferred-feedback --json` readback for `docs/kkachi-docs-map.yaml` `deferred_feedback: "docs/deferred-feedback.md"`, ledger path/hash, entry counts, blocker-defer rejection, non-empty Blue disposition refs, distinct logical/readback status hash vs byte-level ledger SHA, and final-report hidden-open checks.

Until the official v0.2.1 Go/KAH MAR path is ready, legacy Python MAR remains OFF/HOLD and must be proven only as absent/rejected input: KAS source must not restore `scripts/mar.py` or `scripts/mar_adapters/*.sh`; KAH diagnostics may report copied local MAR surfaces only as `rejected_input` with `live_provider_execution=false`. V02FLOW-016 does not run provider MAR. Development-task MAR is closed only through the standing MAR waiver evidence for this interim window, and waiver-only closeout makes post-MAR second-color adoption N/A because there is no provider MAR result to adopt.

## 6. Verification expectations

Every implementation task must:

- update KAS docs/registries/templates/skills only for its bounded scope;
- keep KAH behavior claims tied to the KAH companion repo evidence;
- run the touched repo's `docs/kkachi-docs-map.yaml` `test_commands` after any code-modifying step;
- preserve Red/Orange/Gray review, MAR, and post-MAR second color adoption as separate gates unless a documented waiver is approved;
- until official v0.2.1 Go/KAH MAR readiness, record MAR as standing-waived per run, do not claim provider execution, and mark post-MAR second-color adoption not applicable when the closeout is waiver-only;
- preserve KAT as factual/mechanical evidence only;
- preserve KAB runtime/session control as explicit-only.

## 7. Deferrals

This SOT does not authorize implementation, commit, push, install, tag/release publication before V02FLOW-018 closeout or at any time without separate 주군 approval, live/runtime activation, KAB activation, provider/auth/token/gateway/model/profile mutation, automatic Blue synthesis, auto-continue, watcher-launched source mutation, review-lane waiver, or provider-MAR execution claims before the official v0.2.1 Go/KAH MAR path is ready. V02FLOW-018 is release closeout/readiness packaging only; actual release publication still requires separate approval.
