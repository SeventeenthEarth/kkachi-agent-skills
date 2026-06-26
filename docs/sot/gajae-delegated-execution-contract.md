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

## 2. Pilot-verified facts

The pilot evidence listed above proves the following implementation assumptions are safe to plan from:

1. GJC `deep-interview`, `ralplan`, and `ultragoal` are present in the installed GJC skill set.
2. `gjc ralplan --write` can emit plan artifacts with path, stage, stage number, and SHA-256 receipt.
3. `gjc ultragoal create-goals` and `gjc ultragoal status --json` expose a durable goal ledger/status surface.
4. GJC can call Hermes Kanban CLI to add a comment and complete a task.
5. A Hermes background process running GJC can perform that callback, so Hermes does not need to spend LLM tokens while waiting.
6. KAT v0.1.0 can emit raw logs, summary JSON/Markdown, status JSON, status hash, and KAH-compatible run-id artifact paths.
7. KAT `--run-id` is a global flag and must appear before `run`.
8. GJC execution from Hermes needs real-user home normalization such as `HOME=/Users/draccoon`; the KAH wrapper must own this instead of leaving it to packet authors.
9. Non-interactive GJC use must carry an explicit `GJC_SESSION_ID` or equivalent wrapper-managed session id.
10. Same-thread Discord wake-up is not yet fully productized; GAJAE must capture notification metadata or watcher evidence before claiming routine same-thread wake.

## 3. Ownership boundaries

### KAS owns

- Korean source-command preservation and English operational briefs for GJC packets.
- SOT envelope, acceptance criteria, non-goals, forbidden changes, stop/ask gates, and fail-closed fallback policy.
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
and when reading status. `artifact_refs` are GJC candidate output evidence:
run-local artifact paths plus hashes emitted by GJC and preserved by KAH. KAH validates packet and artifact references mechanically only; it does not parse packet policy, choose fallback behavior, approve plans, adjudicate review/MAR, or decide final acceptance.

Packets must encode stop/ask gates, plan-lock expectations where applicable, no
hidden fallback, no warning-only final gates, no GJC self-approval, and
KAS/Blue/color/MAR/final authority boundaries. Missing, stale, cross-run,
unreadable, non-regular, or checksum-mismatched `packet_ref` evidence fails
closed and requires regenerating or repairing the run-local packet evidence
before consuming GJC status.

## 5. Gate vocabulary

- `deep-interview_ready`: design/epic candidate exists; not accepted until KAS/Blue accepts it.
- `ralplan_ready`: task plan candidate exists; not accepted until plan review and plan lock pass.
- `plan_locked`: accepted plan hash is recorded and future drift requires a plan-conflict report.
- `ultragoal_ready`: implementation bundle candidate exists; not accepted until verification and review gates pass.
- `kat_evidence_ready`: KAT command evidence exists; does not imply acceptance by itself.
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

## 7. Async and wake-up policy

GAJAE must minimize Hermes-visible token usage:

1. KAS sends compact packets and exits the active reasoning loop when GJC/KAT work is running.
2. KAH/GJC writes deterministic status and artifact refs.
3. A no-agent watcher, Kanban callback, or process completion callback prints only compact actionable state changes.
4. KAS/Blue re-enters only for questions, plan review, implementation review, MAR disposition, conflict handling, or final gate.

Watchers and callbacks must not approve plans, summarize raw logs with an LLM, decide review results, or mark final completion by themselves.

## 8. Shared task sequence

GAJAE uses shared logical task ids across KAS and KAH. Repo-local commits/PRs and gates remain separate.

| Task ID | Owner | Title | Status |
|---|---|---|---|
| GAJAE-001 | KAS-led docs/SOT | Register GAJAE SOTs and roadmap sequence | Completed |
| GAJAE-002 | KAH | Implement KAH GJC wrapper MVP | Completed |
| GAJAE-003 | KAS+KAH (KAS-led) | Add GJC packet/template and artifact-reference contract | Completed |
| GAJAE-004 | KAS+KAH | Pilot async ralplan with Kanban wake and plan lock | Planned |
| GAJAE-005 | KAS+KAH | Pilot async ultragoal with KAT evidence and review loop | Planned |
| GAJAE-006 | KAS+KAH | Productize watcher/callback closeout and docs/evidence surfaces | Planned |

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
