# V02FLOW — KAS v0.2 GJC workflow train correction SOT

Date: 2026-07-04
Owner: KAS workflow/policy layer
Confirming role: Hwangchung / 황충, Kkachi Blue commander
Status: planning SOT / implementation not authorized by this document alone
Authority level: KAS-side planning authority for v0.2 workflow-stage, GJC status semantics, implementation executor-loop, review train, and watcher policy corrections
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
ralplan_recorded
ralplan_candidate_recorded
ralplan_consensus_seeded
ralplan_consensus_complete
```

`ralplan_ready` is legacy/current shorthand only. Treat it as candidate plan evidence until KAS/Blue/color plan review accepts and locks it.

### 2.3 `ultragoal`

Current KAH `start-ultragoal` invokes:

```text
gjc ultragoal create-goals --brief-file <run-local-brief> --json
```

That produces a goal bundle / ledger evidence. It does not prove source mutation, implementation diff readiness, or review-ready implementation.

Preferred status language:

```text
ultragoal_goals_ready
implementation_goal_bundle_ready
gjc_goal_bundle_ready
```

Reserve `implementation_diff_ready`, `implementation_candidate_ready`, and `implementation_verified` for evidence after an actual executor loop and verification.

## 3. Missing implementation executor loop

A real delegated implementation loop must include at least:

```text
create-goals -> complete-goals -> execute-goal -> checkpoint -> verify -> repeat
```

KAS must not silently fill the missing loop by letting Blue patch source directly. Default boundary:

```text
Blue commands and verifies. Executor implements.
```

Allowed exceptions require explicit recorded evidence, such as 주군-approved Blue direct patch, emergency hotfix, docs-only/trivial correction, or an approved backend-unavailable exception. These exceptions do not authorize live/runtime/auth/provider/gateway/profile mutation, commit, push, release, install, or scope expansion.

## 4. Review train and watcher policy

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

A full train may use one aggregate watcher for first color review and one aggregate watcher for 2nd color adoption. MAR may use its own provider/status watcher if async.

The watcher must report all-done or blocked states only. It must not perform Blue synthesis, inject a fake `진행해`, auto-trigger continuation, waive lanes, mutate source, commit, push, or change runtime/auth/provider/gateway/profile state.

## 5. Planned implementation split

Recommended logical sets: 4.

1. Workflow/stage contract migration: remove `ask` as a normal step, normalize `plan -> ralplan -> impl`, and encode the corrected impl/review train.
2. GJC adapter/status semantics: deep-interview explicit-only/stale `--packet`, ralplan record-only versus consensus, ultragoal goal-bundle readiness.
3. Ultragoal executor loop: add/describe `create-goals -> complete-goals -> execute -> checkpoint -> verify -> repeat`, and prevent Blue direct patching from silently filling implementation gaps.
4. Review train + aggregated watcher: color review -> MAR -> 2nd color adoption, one aggregated watcher per color round, no auto Blue synthesis.

Physical repo work should normally be 8 PR candidates: a KAS contract/guidance side and a KAH helper/evidence side for each logical set. Avoid compressing to 3 or fewer PRs.

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
