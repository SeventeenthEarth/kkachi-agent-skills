# V02FLOW — KAS v0.2 GJC workflow train correction SOT

Date: 2026-07-04
Owner: KAS workflow/policy layer
Confirming role: Hwangchung / 황충, Kkachi Blue commander
Status: planning SOT / implementation not authorized by this document alone
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

Each deferred entry must record created date, owner lane, last-reviewed date, status-change evidence, originating task/run, source phase, source lane/reviewer, source ref, finding summary, why it matters, `Blocking current task: false`, why not fixed now, defer reason, proposed fix, suggested follow-up task, conversion fields, target repos/files, required and carried-forward gates when resumed, final-report refs, Blue disposition, non-empty Blue disposition ref, additional approval/ref when applicable, and terminal evidence/rationale for resolved/rejected/stale states. Current final reports must link deferred entries whenever any accepted feedback is intentionally not fixed now. N/A must never satisfy the Blue disposition ref requirement.

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

Physical repo work should normally be 12 PR candidates: a KAS contract/guidance side and a KAH helper/evidence side for each logical set. Avoid compressing to 3 or fewer PRs.

## 6. Verification expectations

Every implementation task must:

- update KAS docs/registries/templates/skills only for its bounded scope;
- keep KAH behavior claims tied to the KAH companion repo evidence;
- run the touched repo's `docs/kkachi-docs-map.yaml` `test_commands` after any code-modifying step;
- preserve Red/Orange/Gray review, MAR, and post-MAR second color adoption as separate gates unless a documented waiver is approved;
- preserve KAT as factual/mechanical evidence only;
- preserve KAB runtime/session control as explicit-only.

## 7. Deferrals

This SOT does not authorize implementation, commit, push, install, release, live/runtime activation, KAB activation, provider/auth/token/gateway/model/profile mutation, automatic Blue synthesis, auto-continue, watcher-launched source mutation, or review-lane waiver.
