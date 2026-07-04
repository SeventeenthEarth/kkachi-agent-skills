# GAJAE — Gajae delegated execution contract SOT

Date: 2026-06-26
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record
Status: planning SOT / pilot-evidenced / implementation not authorized by this document alone
Authority level: KAS-side SOT for GJC delegated planning/execution packets, Kkachi authority gates, plan lock, review loops, and KAT evidence expectations
Scope: `kkachi-agent-skills` source behavior and guidance; paired KAH wrapper/evidence authority is `kkachi-agent-helper/docs/sot/gajae-gjc-wrapper-evidence.md`
Related docs: `docs/roadmap.md`, `docs/sot/token-economy-and-agent-instruction-contract.md`, `docs/sot/multi-agent-review-policy.md`, `docs/sot/mar-task-loop-contract.md`, `docs/sot/strict-workflow-execution-contract.md`, KAH `docs/sot/gajae-gjc-wrapper-evidence.md`
External evidence: `/Users/draccoon/Workspace/Hermes/17thHermes/40_outputs/team/hwangchung/kkachi/2026-06-23-kas-kah-kat-gjc-execution-sot.md`, `/Users/draccoon/Workspace/Hermes/17thHermes/50_health/team/hwangchung/backups/gjc-delegated-execution-pilot-20260625/report.md`

## 1. Decision

`GAJAE` is the shared KAS/KAH epic for making Gajae Code (`gjc`) the delegated planning and implementation executor for Kkachi work while keeping Kkachi authority in KAS/Blue/color gates and deterministic evidence in KAH/KAT.

The selected mapping is:

```text
Epic/design shaping -> GJC deep-interview
Task planning       -> GJC ralplan
Implementation      -> GJC ultragoal
Verification logs   -> KAT deterministic evidence
Run state/evidence  -> KAH GJC wrapper and ledger
Authority gates     -> KAS/Blue + Red/Orange/Gray + MAR/final gates
```

GJC is an executor, not an authority source. GJC output becomes a candidate artifact until accepted by the relevant Kkachi gate.

## 1.1. V02FLOW correction overlay

The later V02FLOW SOT (`docs/sot/v020-gjc-workflow-train-corrections.md`) narrows active v0.2 interpretation of the mapping above:

- `deep-interview` is explicit-request only, not a normal phase for ordinary tasks.
- Current `ralplan --write` evidence records an artifact unless separate consensus-loop evidence is present; active KAS uses `ralplan_candidate_recorded` as the primary candidate/record status, with `ralplan_recorded` as a technical/readback alias and legacy `ralplan_ready` as candidate compatibility evidence only.
- Current `ultragoal create-goals` evidence is goal-bundle readiness only; active KAS uses `implementation_goal_bundle_ready` as the primary operator-facing status, with `ultragoal_goals_ready` / `gjc_goal_bundle_ready` as technical aliases and legacy `ultragoal_ready` as compatibility evidence only. It does not prove source mutation, implementation diff readiness, verification, review, MAR, or final acceptance.
- The default train is `plan -> ralplan -> impl`; `ask` is not a normal phase, though real approval/question gates still stop for the required actor.
- Substantial implementation review uses `color review -> MAR -> 2nd color adoption`; aggregate watchers report state only and do not synthesize Blue decisions or auto-continue.

Read any older GAJAE wording that implies deeper live GJC behavior through this V02FLOW correction overlay.

## 2. Pilot-verified facts

The pilot evidence listed above plus the 2026-06-27 `/tmp/kkachi-gjc` scratch verification prove the following implementation assumptions and blockers are safe to plan from:

1. GJC `deep-interview`, `ralplan`, and `ultragoal` are present in the installed GJC skill set.
2. Native `gjc ralplan --write` can emit plan artifacts with path, stage, stage number, and SHA-256 receipt, but installed GJC 0.7.3 requires KAH to provide required stage/stage-number/artifact or equivalent inputs; `--packet` alone is not a valid live invocation.
3. Native `gjc ultragoal create-goals` and `gjc ultragoal status --json` expose a durable goal ledger/status surface, but installed GJC 0.7.3 requires KAH to provide `--brief`, `--brief-file`, or `--from-stdin`; `--packet` alone is not a valid live invocation.
4. GJC can call Hermes Kanban CLI to add a comment and complete a task.
5. A Hermes background process running GJC can perform that callback, so Hermes does not need to spend LLM tokens while waiting.
6. KAT v0.1.0 can emit raw logs, summary JSON/Markdown, status JSON, status hash, and Kkachi run-id artifact paths; GAJAE-009 KAH behavior can now normalize these factual status/summary/raw-log refs into KAH-bindable KAT evidence without requiring KAT source changes. KAT `status_hash` remains KAT self-integrity metadata, not the GJC source status hash.
7. KAT `--run-id` is a global flag and must appear before `run`.
8. GJC execution from Hermes needs real-user home normalization such as `HOME=/Users/draccoon`; the KAH wrapper must own this instead of leaving it to packet authors.
9. Non-interactive GJC use must carry an explicit `GJC_SESSION_ID` or equivalent wrapper-managed session id.
10. Same-thread Discord wake-up is not yet fully productized; GAJAE must capture notification metadata or watcher evidence before claiming routine same-thread wake.

## 3. Ownership boundaries

### KAS owns

- Korean source-command preservation and English operational briefs for GJC packets.
- SOT envelope, acceptance criteria, non-goals, forbidden changes, stop/approval/question gates, and fail-closed fallback policy.
- GJC packet templates for `deep-interview`, `ralplan`, `ultragoal`, review/fix turns, and completion callbacks.
- Plan review, plan revision requests, plan lock, and plan-conflict handling.
- Color review, MAR disposition, final gate synthesis, and user-facing completion reports.
- Determining whether a GJC result is acceptable; GJC never self-approves scope, plan, implementation, review, or final completion.

### KAH owns

- Thin deterministic `gjc` wrapper commands and state ledgers.
- Real-user HOME/environment normalization, `GJC_SESSION_ID` persistence, and process status capture.
- GJC artifact references and hashes.
- KAT artifact references and hashes.
- Kanban callback status, idempotency keys, notification metadata, and watcher-compatible status surfaces.
- Deterministic validation of evidence shape only; KAH does not choose policy, reviewers, providers, or plan acceptance.

### KAT owns

- Deterministic command execution, raw-log capture, compact summaries, status JSON, and extractor/rule evidence.
- KAT exit code and command exit code remain factual evidence. KAT does not make Kkachi acceptance decisions.

### KAB role

KAB is not the primary control plane for GAJAE. GAJAE does not force GJC through KAB legacy session-control. Later KAB integration may be added only by a separate approved task if concrete capability gaps require it.

## 4. Candidate command and packet flow

```text
KAS source command / SOT envelope
  -> KAS GJC planning packet
  -> KAH gjc start-deep-interview or start-ralplan
  -> GJC artifact/status
  -> KAH callback/status evidence
  -> KAS Blue + Red/Orange/Gray plan review
  -> KAS plan lock
  -> KAS ultragoal execution packet
  -> KAH gjc start-ultragoal
  -> GJC implementation bundle + KAT evidence
  -> KAS color review / MAR / fix packets as needed
  -> KAS final synthesis
```

KAS packets must be English for agent-facing GJC content. Direct reports/questions to 주군 remain Korean.

## 4.1. GAJAE-003 packet contract

GAJAE-003 adds KAS-owned machine-readable packet templates for `deep_interview`,
`ralplan`, `ultragoal`, `review_fix_turn`, and `callback_contract` under
`templates/run-artifacts/`. Each packet preserves `source_command.korean` as the
source of truth and uses `gjc_operational_brief_english` for GJC-facing
operational instructions.

Required packet fields are `schema_version`, `packet_kind`, `task_id`, `run_id`,
`source_command.korean`, `gjc_operational_brief_english`,
`authority_boundaries`, `stop_ask_gates`, `plan_lock`,
`fallback_policy: none_fail_closed`, `no_gjc_self_approval: true`,
`forbidden_scope`, `expected_outputs`, `artifact_ref_contract`, and
`completion_boundary`.

`packet_ref` is KAS input packet evidence: a repository-relative run-local packet
path plus `sha256:<hash>` that KAH may validate mechanically before starting GJC
and when reading status. `native_input_ref` is KAH-materialized run-local
native input evidence derived from KAS packet fields such as `native_ralplan_input`
or `native_ultragoal_input`; KAH validates its path and hash mechanically and
uses it only as GJC invocation input, not as approval evidence. `artifact_refs` are GJC candidate output evidence:
run-local artifact paths plus hashes emitted by GJC and preserved by KAH. KAH validates packet and artifact references mechanically only; it does not parse packet policy, choose fallback behavior, approve plans, adjudicate review/MAR, or decide final acceptance.

Packets must encode stop/approval/question gates, plan-lock expectations where
applicable, no hidden fallback, no warning-only final gates, no GJC
self-approval, and
KAS/Blue/color/MAR/final authority boundaries. Missing, stale, cross-run,
unreadable, non-regular, or checksum-mismatched `packet_ref` evidence fails
closed and requires regenerating or repairing the run-local packet evidence
before consuming GJC status.

## 5. Gate vocabulary

Current preferred GJC delegated-execution status vocabulary:

- `deep-interview_ready`: legacy design/epic candidate compatibility status; deep-interview remains explicit-request only and is not accepted until KAS/Blue accepts it.
- `ralplan_candidate_recorded`: primary current status for a recorded ralplan candidate artifact; not accepted until plan review and plan lock pass.
- `ralplan_recorded`: technical/readback alias for recorded ralplan candidate evidence.
- `ralplan_ready`: legacy/current-helper compatibility alias for ralplan candidate evidence; not consensus, plan acceptance, implementation approval, or final readiness.
- `plan_locked`: accepted plan hash is recorded and future drift requires a plan-conflict report.
- `implementation_goal_bundle_ready`: primary current operator-facing status for an ultragoal create-goals bundle; not accepted until executor-loop, verification, and review gates pass.
- `ultragoal_goals_ready`: technical/backend alias for ultragoal goal-bundle evidence.
- `gjc_goal_bundle_ready`: technical/backend alias for GJC goal-bundle evidence.
- `ultragoal_ready`: legacy/current-helper compatibility alias for goal-bundle evidence; not source mutation, implementation diff readiness, verification, review, MAR, waiver, or final acceptance.
- `kat_evidence_ready`: KAT command evidence exists; does not imply acceptance by itself.
- `review_fix_candidate_ready`: bounded review/fix candidate evidence exists for accepted findings; it is distinct from ordinary goal-bundle readiness and does not close findings or approve MAR/color/final gates.
- `callback_delivered`: Kanban/watcher callback evidence exists; does not imply review or completion.
- `final_accepted`: KAS/Blue final synthesis after required evidence/review gates.

## 6. KAT invocation contract

KAS templates and guidance must show KAT run-id usage as:

```bash
kkachi-agent-tester --run-id <run_id> run --lane <lane> -- <command...>
```

The following form is invalid and must not be used in generated packets:

```bash
kkachi-agent-tester run --run-id <run_id> ...
```

Extractor status handling:

- `precise` or `partial` can support compact failure triage.
- `degraded` or `no_match` is a rule-mining signal, not a pass/fail override.
- Command exit code remains authoritative for command success/failure.

KAH KAT attachment evidence is factual only and should include run id,
status ref/hash, summary JSON ref/hash, summary Markdown ref/hash, raw-log
ref/hash, extractor status, command exit code, and attachment status. For
GAJAE-009, KAH may derive summary/raw refs and hashes from KAT v0.1.0
`summary_path`, `summary_sha256`, `raw_log_path`, and `raw_log_sha256` fields
when the caller supplies the run-local KAT status ref/hash; the derived Markdown
summary ref is the sibling `<summary>.md` file. Missing, unsafe, cross-run,
malformed, hashless, checksum-mismatched, run-id-mismatched, symlinked,
non-regular, approval-claiming, waiver-claiming, or final-acceptance-claiming
KAT evidence fails closed. KAH records the evidence shape; KAS/Blue/color/MAR/final gates
interpret acceptance.

## 7. Async and wake-up policy

GAJAE must minimize Hermes-visible token usage:

1. KAS sends compact packets and exits the active reasoning loop when GJC/KAT work is running.
2. KAH/GJC writes deterministic status and artifact refs.
3. A no-agent watcher, Kanban callback, or process completion callback prints only compact actionable state changes.
4. KAS/Blue re-enters only for questions, plan review, implementation review, MAR disposition, conflict handling, or final gate.

Watchers and callbacks must not approve plans, summarize raw logs with an LLM, decide review results, or mark final completion by themselves. Productized callback/watcher closeout evidence must report only factual routing state: run id, task id, callback idempotency key, source status hash, callback result, notification metadata availability, wake-evidence status, current required actor, wait reason, recovery hint, and artifact/status refs. Missing origin/thread/watcher evidence remains `no-wake-claim` rather than a degraded wake-readiness claim.

### GAJAE-004 source-side pilot

GAJAE-004 may update source-side KAS templates and KAH helper behavior for the
async `ralplan` / callback / plan-lock pilot after explicit approval/apply evidence.
This authorization is bounded to source, tests, and docs. It does not authorize
commit, push, install, release, live/default runtime activation, KAB Stage 2/3
activation, real GJC/KAT execution outside tests/mocks, or
profile/provider/auth/token/gateway/model mutation.

`ralplan_candidate_recorded` is the primary active KAS status for a recorded
ralplan candidate. `ralplan_recorded` may appear as a technical/readback alias.
Legacy `ralplan_ready` and `callback_delivered` remain candidate/evidence states only.
Neither state approves the plan, starts implementation, satisfies color/MAR
review, or marks final completion. `plan_locked` requires accepted_plan_hash after KAS/Blue/color review
and must bind that hash to the reviewed candidate plan artifact. Post-lock drift
requires a plan-conflict report before continuation.

### GAJAE-005 source-side pilot

GAJAE-005 may update source-side KAS templates and KAH helper behavior for the
async `ultragoal` / KAT attachment / review-fix pilot after explicit
approval/apply evidence. This authorization is bounded to source, tests, and docs. It does not
authorize commit, push, install, release, live/default runtime activation, KAB
Stage 2/3 activation, real GJC `ultragoal` invocation, live KAT execution, or
profile/provider/auth/token/gateway/model mutation.

`implementation_goal_bundle_ready` is the primary active KAS operator-facing
status for `ultragoal create-goals` output. `ultragoal_goals_ready` and
`gjc_goal_bundle_ready` may appear as technical/backend aliases. Legacy
`ultragoal_ready`, `kat_evidence_ready`, and `review_fix_candidate_ready` remain
candidate/factual evidence states only. They do not approve implementation,
close findings, satisfy color review, satisfy MAR, approve waivers, or mark
final completion. Review/fix packets must use `review_fix_candidate_ready` or an
equally distinct factual status so fix-turn evidence cannot be confused with
ordinary implementation readiness.

Review/fix-turn KAT attachment is not required in GAJAE-005 until KAH implements
a dedicated review-fix command/status path. The `review_fix_turn` packet may
preserve deferred KAT attachment metadata, but it must not claim that KAH can
attach KAT evidence to `review_fix_candidate_ready` through the current
`attach-kat-evidence` path. Current helper compatibility may still attach KAT
evidence to legacy `ultragoal_ready` until a later approved KAH task expands the
status path; KAS guidance treats the primary goal-bundle status as
`implementation_goal_bundle_ready`.

## 8. Shared task sequence

GAJAE uses shared logical task ids across KAS and KAH. Repo-local commits/PRs and gates remain separate.

| Task ID | Owner | Title | Status |
|---|---|---|---|
| GAJAE-001 | KAS-led docs/SOT | Register GAJAE SOTs and roadmap sequence | Completed |
| GAJAE-002 | KAH | Implement KAH GJC wrapper MVP | Completed |
| GAJAE-003 | KAS+KAH (KAS-led) | Add GJC packet/template and artifact-reference contract | Completed |
| GAJAE-004 | KAS+KAH | Pilot async ralplan with Kanban wake and plan lock | Completed / source-side pilot |
| GAJAE-005 | KAS+KAH | Pilot async ultragoal with KAT evidence and review loop | Source-side pilot |
| GAJAE-006 | KAS+KAH | Productize watcher/callback closeout and docs/evidence surfaces | Completed |
| GAJAE-007 | KAH+KAS | Real GJC `ralplan` adapter | Completed |
| GAJAE-008 | KAH+KAS | Real GJC `ultragoal` adapter | Completed |
| GAJAE-009 | KAH+KAT (KAH-led) | KAT evidence normalization / KAH attach adapter | Completed |
| GAJAE-010 | KAS+KAH+KAT | Contract docs and skill guidance update | Completed |

## 8.1. GAJAE-007..010 pilot-unblock task scope

- `GAJAE-007` changes code/contract behavior so KAS packets carry `native_ralplan_input.stage`, `.stage_n`, and `.artifact`, KAH derives native GJC 0.7.3 `ralplan --write` flags from those fields, and KAH records only candidate plan evidence.
- `GAJAE-008` changes code/contract behavior so KAH materializes `native_ultragoal_input.brief` into run-local `native_input_ref` evidence, invokes GJC 0.7.3 as `ultragoal create-goals --brief-file <path> --json`, adapts native goals/ledger output into run-local `artifact_refs`, and records only implementation-candidate evidence.
- `GAJAE-009` changes code/contract behavior so KAH can attach factual KAT v0.1.0 evidence through KAH-side normalization of status/summary/raw-log refs without requiring KAT source changes. KAT remains factual and never authoritative.
- `GAJAE-010` updates KAS/KAH/KAT repo docs and Hermes skill guidance after the adapters settle. The closeout preserves `native_ralplan_input`, `native_ultragoal_input`, KAH-derived `native_input_ref`, KAH-side normalization of KAT factual status/summary/raw-log refs, and the rule that GJC/KAT/KAH evidence is candidate, factual, or mechanical until KAS/Blue/color/MAR/final gates accept it. Verification evidence remains a done criterion inside those tasks, not a separate roadmap task.

## 9. GAJAE-001 acceptance criteria

GAJAE-001 is complete only when:

1. KAS and KAH SOT docs exist and cross-link each other.
2. KAS and KAH roadmaps register the GAJAE epic and task sequence.
3. Docs indexes and docs maps include the new SOTs.
4. Docs verification passes in both repositories, or blockers are recorded explicitly.
5. Red/Orange/Gray review accepts the planning SOTs, or 주군 explicitly waives the review for docs-only registration. Completed evidence: Red `t_4cbf4624`, Orange `t_18dccb4c`, Gray initial `t_bbb1af05`, Gray focused re-review `t_c6ba0567`, and Blue synthesis `t_6be5b0e5` accepted the docs-only planning registration after traceability fixes; synthesis artifact: `/Users/draccoon/Workspace/Hermes/17thHermes/50_health/team/hwangchung/backups/gajae-001-color-review-20260626/blue-synthesis.md`.
6. No implementation, runtime activation, profile mutation, provider/auth/gateway/model mutation, push, release, or install is claimed.

## 10. Deferrals

GAJAE does not authorize:

- KAB Stage 2/3 activation or backend selection changes;
- provider/auth/token/gateway/model mutation;
- automatic fallback from GJC failure to another backend;
- automatic plan re-write after plan lock;
- GJC self-approval of scope, plan, implementation, or final completion;
- warning-only behavior for missing GJC/KAT/KAH evidence;
- broad profile mutation, install, release, push, or runtime activation;
- same-thread Discord wake claims until metadata/watcher evidence is implemented and verified.
