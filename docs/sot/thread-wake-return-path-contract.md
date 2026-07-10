# TWAKE — thread wake return-path contract SOT

Date: 2026-07-10
Owner: KAS workflow/policy layer
Confirming role: Hwangchung / 황충, Kkachi Blue commander
Status: planning SOT for TWAKE-001..004; no implementation, helper behavior, release, install, runtime activation, Discord delivery, provider execution, profile mutation, or auth/provider/gateway/model change is authorized by this document alone
Authority level: KAS-side planning authority for Blue async-dispatch return-path policy, watcher/notifier requirements, and final verification expectations
Scope: `kkachi-agent-skills` source docs, skills, templates, registries, final-verification guidance, and roadmap records for target KAS `v0.2.5`
Paired KAH companion SOT: KAH `docs/sot/thread-wake-return-path-evidence.md`
Related docs: `docs/roadmap.md`, `docs/sot/v020-gjc-workflow-train-corrections.md`, `docs/sot/gajae-delegated-execution-contract.md`, `docs/sot/token-economy-and-agent-instruction-contract.md`, KAH `docs/sot/thread-wake-return-path-evidence.md`

## 1. Decision

Blue-dispatched asynchronous Kkachi work must have a return path back to the active operator context before Blue can rely on the dispatch as controllable. For Discord-origin work, the preferred return path is the current Discord thread. If the current thread cannot be proven, the run must preserve `no_wake_claim` or stop for a new return-path decision instead of implying Blue will be notified.

This contract applies when Blue dispatches or coordinates any long-running or fan-in work whose result requires Blue response, including:

- plan-vet and plan-review cards;
- first color review Red/Orange/Gray fan-out;
- MAR start/status/wait/callback/watcher paths;
- post-MAR second-color adoption/review fan-out;
- GJC `ralplan`, `ultragoal`, implementer-phase, and remediation dispatches that may complete after the current reasoning turn;
- bounded blocked-condition probes or long-running verification watchers when they are part of an active KAS/KAH run.

The return path is not acceptance authority. It exists only to wake Blue or the operator with a compact actionable state such as all required lanes returned, lane blocked/failed/cancelled, MAR coverage unresolved, GJC candidate ready, or Blue disposition required.

## 2. KAS/KAH ownership split

```text
KAS owns: when return-path evidence is required, what dispatch classes require it, reviewer/GJC/MAR policy, Blue disposition, and final acceptance rules.
KAH owns: deterministic evidence shape, validation/readback, capability flags, watcher-output compatibility, and fail-closed diagnostics when required evidence is missing.
KAT owns: factual test/status/summary/raw-log evidence only; it does not own notifications, Discord delivery, review, MAR, waiver, or final acceptance.
```

KAS must fail closed or report a blocking/degraded condition when the effective KAH binary lacks the reviewed return-path evidence capability needed for a run. Source docs, intended behavior, or KAH checkout state are not enough; KAS must check the effective binary/capability evidence before claiming same-thread wake readiness.

## 3. Return-path requirement

For a required async dispatch, KAS prompt/skill/gate guidance must require evidence equivalent to:

```yaml
return_path:
  schema_version: twake.return_path.v1
  requires_return_path: true
  origin_platform: discord | local | other
  origin_chat_id: "<redacted-or-stable-chat-ref>"
  origin_thread_id: "<thread/topic/ref when present>"
  notification_method: kanban_notify_subscribe | cron_no_agent_watcher | process_callback | kah_watcher_output | none
  notification_id: "<subscription/watcher/process/callback id or no_wake_claim>"
  delivery_target: origin | local | explicit
  watched_condition: "<mechanically checkable condition>"
  terminal_only: true
  blue_action_required_output: true
  no_authority: true
  same_thread_wake_claim: false
  wake_evidence_status: proven | no_wake_claim | missing_origin_metadata | missing_watcher_evidence
  artifact_refs:
    - path: ".kkachi/runs/<run_id>/..."
      checksum: "sha256:<64 lowercase hex>"
```

Acceptable methods include same-card Kanban `notify-subscribe`, bounded no-agent/cron watcher with `deliver=origin`, tracked process completion callback, or KAH watcher-output/callback evidence that an external notifier can deliver. The method must be mechanically checkable and must produce terminal/actionable-only output.

Missing watcher/subscription id, missing origin/thread metadata for a same-thread claim, missing terminal-only policy, missing no-authority boundary, or unsafe/missing artifact refs must block clean closeout for dispatch classes where KAS declared return-path evidence required.

## 4. Operational rules

- Blue must attach the return path before or at dispatch time, not after the fact during final reporting.
- Same-card Kanban remains the durable work bus for reviewer questions and results; notifier/watcher output does not replace Kanban comments, review artifacts, KAH evidence, or Blue synthesis.
- Watchers must stay silent for unchanged/nonterminal state and emit only compact Blue-action-required terminal/actionable reports.
- Watchers must not fake `진행해`, auto-continue, approve plans, satisfy color review, satisfy MAR, waive lanes, mutate source, commit, push, install, release, or change runtime/auth/provider/gateway/profile/model settings.
- If return-path delivery is unavailable, the correct state is `blocked`, `degraded`, or `no_wake_claim` with an explicit recovery hint; never imply same-thread wake readiness without proof.
- TWAKE-003 KAS consumer guidance must normalize the return-path vocabulary as `blocked`, `degraded`, or `no_wake_claim`, each with an operator-readable recovery hint, so source guidance, templates, and final verification cannot treat a no-wake state as clean same-thread notification.

## 5. TWAKE task sequence

| Task ID | Primary owner | Target | Purpose | Required boundary |
|---|---|---|---|---|
| TWAKE-001 | KAH | KAH `v0.2.3` | Add or document deterministic return-path evidence schema/capability/readback and fail-closed validation for required async dispatches. | KAH validates evidence only; it does not choose notification policy or send Discord messages by itself. |
| TWAKE-002 | KAH | KAH `v0.2.3` | Integrate return-path validation with review watcher-output, MAR watcher/callback, GJC long-running dispatch/status, and final/phase diagnostics. | Missing return-path evidence cannot close cleanly when KAS declared it required. |
| TWAKE-003 | KAS | KAS `v0.2.5` | Update KAS phase skills/templates/final guidance so Blue dispatches plan-vet, color review, MAR, second-color, and GJC long-running work only with return-path evidence or an explicit no-wake/blocked state. | KAS must capability-check effective KAH before relying on the evidence surface. |
| TWAKE-004 | KAS-led closeout | KAS `v0.2.5` + KAH `v0.2.3` | Align roadmaps/docs, run official review, and prepare source-readiness evidence for the TWAKE train. | Readiness only; publication, push/tag/install/runtime/provider/auth/profile/gateway/model changes need separate approval. |

## 6. Verification expectations

TWAKE implementation tasks must prove positive and negative cases:

- valid return-path evidence passes deterministic validation;
- missing watcher/subscription id fails closed;
- same-thread wake claim without origin/thread metadata fails closed;
- missing terminal-only/no-authority policy fails closed;
- no-wake/no-delivery states are explicit and cannot be mistaken for notification success;
- final/pre-commit reporting rejects required async dispatches that lack return-path evidence;
- KAT status JSON remains factual input only and does not become notification authority.

## 7. Non-goals and holds

TWAKE does not authorize implementation beyond its bounded tasks, KAB activation, provider execution, profile/provider/auth/token/gateway/model mutation, live Discord delivery changes, runtime activation, install, release publication, tag, push, or commit by itself. TWAKE also does not replace Blue judgment, Red/Orange/Gray review, MAR role coverage, KAH final gates, or 주군 approval boundaries.
