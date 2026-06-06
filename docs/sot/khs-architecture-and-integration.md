# KAS architecture, usage, and KAH integration SOT

Date: 2026-05-23
Owner: KAS workflow/policy layer
Confirming role: Blue final synthesis after required Red, Orange, and Gray review; INITDOC post-KAH reset pass
Status: current broad KAS architecture SOT; KAH 0.1.4 graph/configurable-feedback substrate evidenced; not a KAB runtime support claim
Authority level: broad KAS architecture/usage/KAH-integration source of truth after INITDOC reset
Scope: `kkachi-hermes-skills` architecture, repo layout, Hermes usage model, KAS/KAH/KAB/KHC boundaries, workflow policy, self-improvement governance, and development acceptance gates
Related docs:
- `README.md`
- `docs/README.md`
- `docs/sot/interface-contract.md`
- `docs/sot/phase-orchestration-policy.md`
- `docs/sot/workflow-graph-integration.md`
- `docs/sot/minimum-pilot-cli-lane.md`
- `docs/roadmap.md`
Evidence/source paths:
- Kanban parent readiness and scope decision: `t_5d35ee82`
- Current broad SOT writing card: `t_f29b6ee9`
- Prior consult cards for readiness: Red `t_7651a159`, Gray `t_1cc2f77c`, Orange `t_7d8607e2`
- Current draft review cards: Red `t_f94bca58`, Orange `t_8a3459c0`, Gray `t_29a6f341`; Teal skipped because no concrete UI/screen-flow surface was introduced
- Post-revision Gray traceability card: `t_6c9ff793` (`REQUEST_CHANGES` for missing `registries/phase-contracts.yaml` stale-line inventory; addressed in this revision)
- KAH support-plan closure lineage for external feedback intake: `t_ccf7beaf` -> `t_56c9688f` -> `t_9caf23c9` -> `t_dfe4db34`; Red closure verdict recorded as `RED_CLOSURE_ACCEPT`
- Historical KAH support-plan path: `/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-helper/.docs/new-support.md` (superseded by KAH 0.1.4 capability evidence for graph/configurable-feedback substrate; still useful for lineage)
- KAS/KHS direction reference: `/Users/draccoon/.hermes/skills/dogfood/hermes-team-operations/references/hermes-team-khs-direction-delegation-packet-evidence-loop-2026-05.md`
- Kkachi reference records used for source context:
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-khc-khs-kae-repo-extension-productization-2026-05-23.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-khs-kah-minimum-profile-scoped-skill-injection-2026-05-23.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-khs-kah-codex-appserver-mvp-thread-workflow-config-2026-05-23.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-khs-kah-codex-appserver-mvp-thread-handles-2026-05-23.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-khs-intended-workflow-gap-review-2026-05-21.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-kah-khs-docs-first-workflow-graph-sot-roadmap-2026-05-21.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-kah-command-managed-workflow-config-2026-05-21.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-phase1-kah-khs-kab-workflow-consensus-2026-05-21.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-phase1-workflow-v2-review-2026-05-21.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-external-feedback-intake-configurable-bounds-2026-05-23.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-universal-agent-harness-coding-first-vertical-2026-05-23.md`
  - `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-agent-evaluation-kae-design-2026-05-23.md`

Record transition: prior readiness work recommended `docs/sot/external-feedback-intake-policy.md` for a narrower external-feedback policy record. INITDOC keeps this broader architecture file as the parent SOT for KAS usage, KAH connection, and repo structure. The external-feedback file remains a child/detail SOT, not a competing path.

## 1. Decision summary

KAS is the Hermes skill/process layer for Kkachi-governed software development. The repository and older docs may still say KHS, but INITDOC treats KAS as the current canonical name. KAS must not be treated as a broad static `SKILL.md` bundle or a toy self-improvement package.

The target definition is:

```text
KAS = Delegation Packet + Evidence Loop + proposal-gated self-improvement system
```

That means KAS turns a user request into a bounded, reviewable delegation packet for Hermes and backend workers, preserves evidence expectations across KAH and KAB, and routes lessons into run-local, project-overlay, or shared-KAS proposals only after review and approval.

This document is a current architecture SOT for post-KAH KAS development. It defines target architecture and gates, recognizes KAH 0.1.4 graph/configurable-feedback capability evidence, and still refuses unsupported KAS adoption or KAB runtime claims.

## 2. ASIS facts

Current repository facts and evidence checked during this drafting pass:

- `README.md` already defines KHS as the prompt/process layer, KAH as deterministic state, and KAB as backend runtime/control.
- `docs/README.md` defines the documentation authority ladder and marks `docs/sot/` as the durable authority lane.
- `docs/sot/interface-contract.md` records the current KAS/KAH/KAB interface and reality-first rule: KAS must use surfaces that exist now, and record gaps instead of pretending future interfaces exist.
- `docs/sot/workflow-graph-integration.md` records `.kkachi-workflow.yaml` as project-level graph state only when backed by KAH graph evidence, while `phase-plan.yaml` remains run-local execution state/evidence.
- `docs/sot/minimum-pilot-cli-lane.md` records a scoped KHS+KAH minimum/pilot lane for profile-scoped install/list/doctor/sync/proposal support, distinct from the full KHS+KAH+KAB execution-runtime lane.
- `registries/phase-contracts.yaml`, run-artifact templates, and phase skills have been updated to the `min=1, max=5` feedback-loop policy; older `1..3` / `maximum_rounds: 3` mentions in this document are historical inventory rows only.
- KAH 0.1.4 capability evidence now advertises graph/configurable-feedback substrate support, and KAS registry/template/skill adoption exists; operator report/e2e adoption remains integration-pending before active KAS runs can label the full surface implemented.
- The KAH `.docs/new-support.md` external feedback intake plan is now historical lineage, not the current blocker.
- The KHS repo had pre-existing dirty and untracked files before this draft was written. Observed status during review included `MM README.md`, `M docs/README.md`, several modified tracked docs/registries/skills/templates, and untracked `AGENTS.md`, `docs/roadmap.md`, `docs/sot/khs-architecture-and-integration.md`, `docs/sot/minimum-pilot-cli-lane.md`, and `docs/sot/workflow-graph-integration.md`. This draft should be reviewed in that context and must not overwrite unrelated work.

## 3. TOBE target

KHS should provide a practical coding-agent harness layer for Hermes:

1. Translate a user request into an AI-neutral task contract.
2. Select the correct KHS phase guidance and backend prompt profile.
3. Build a Delegation Packet that includes task contract, bounded history citations, capability needs, evidence plan, report contract, and recovery rules.
4. Use KAH for deterministic project/run state, artifact persistence, workflow graph validation/proposal/apply/audit, gates, events, locks, diagnostics, and fail-closed validation.
5. Use KAB for backend session dispatch, plan/question/approval lifecycle, wait/read/status, retained stream/event evidence, and backend execution evidence.
6. Let Hermes/KHC command roles approve risk, route review, synthesize verdicts, and report to the user.
7. Capture lessons into a proposal-gated self-improvement loop rather than automatically mutating shared KHS.

KHS must optimize for operator reconstruction: a reviewer should be able to determine what was requested, what was delegated, what evidence exists, what was accepted/rejected/deferred, and what can be safely claimed without rereading the whole raw conversation.

## 4. Non-goals

KHS must not:

- Become a second deterministic state system beside KAH.
- Own backend runtime/session control beside KAB.
- Claim backend callability from semantic guidance, examples, stale docs, or prompt wording.
- Treat `send` or dispatch success as completion evidence.
- Treat old conversation context as hidden truth without bounded source citations.
- Auto-clear review gates, user approval gates, KAH gates, or KAB approval/question pendings.
- Auto-patch shared KHS skills/docs from a single run.
- Promote run-local observations directly into shared skills without evaluation and approval.
- Use `.kkachi/config.yaml` as workflow graph authority.
- Directly hand-edit `.kkachi-workflow.yaml` and call it authoritative as the normal path.
- Hide backend-specific behavior behind a false common abstraction.
- Turn KHC member names into product requirements; product docs should prefer role/layer terms.

## 5. Layer boundaries

| Layer | Owns | Must not own |
|---|---|---|
| KHC | organization, command roles, collaboration rules, review gates, escalation, final synthesis/acceptance policy | KAH state mechanics, KAB session mechanics, direct code ownership by command roles |
| KHS | skills, phase contracts, task contracts, prompt profiles, project overlays, templates, policy guidance, Delegation Packet schema, evidence expectations, self-improvement proposal rules | deterministic state, graph apply/audit, backend execution, backend-native inventory truth, approval bypass |
| KAH | `.kkachi/` deterministic state, project/run artifacts, schemas, events, locks, gates, diagnostics, graph validate/explain/diff/propose/apply/audit, fail-closed validation | workflow policy choice, phase applicability decision, backend execution, KHS shared-skill promotion |
| KAB | backend sessions, dispatch/read-wait-status, retained stream/event observation, plan lifecycle, questions, approvals, backend execution evidence, raw capability discovery | project workflow graph authority, KHS semantic policy, KAH persistence authority |
| Hermes history and Kanban | bounded source material with explicit ids, comments, summaries, and freshness context | hidden KHS-owned truth or unstated authority |
| Project overlay | project-specific conventions, local prompt guidance, local policy proposals, local evidence references | shared KHS mutation without promotion gate |

Boundary rule: KHS may recommend, explain, render, and propose. KAH validates and records deterministic state. KAB executes and observes backends. KHC/Hermes accepts risk and evidence. No layer may silently substitute for another layer.

## 6. How Hermes uses KHS

### 6.1 Activation

Hermes should activate KHS when the user or operating context asks for Kkachi, KHS, KAH, KAB, a Kkachi run, KAB-backed execution evidence, gate-backed artifacts, or assignment to a KHS-using Kkachi development commander/worker.

Hermes should not activate KHS by default for a simple direct edit, typo fix, quick config tweak, or read-only explanation unless the user explicitly asks for KHS/Kkachi handling.

### 6.2 Profile-scoped injection

KHS skills are profile-scoped in Hermes. The target install model is:

- Baseline KHS pack: reusable shared skills, registries, templates, and docs in this repo.
- Profile-scoped installed copy: selected skills/categories copied or otherwise installed into a Hermes profile with manifest/checksum evidence.
- Project overlay: project-specific guidance, conventions, backend policy, docs map, and local proposals.
- Proposal layer: run-local or project-overlay improvement proposals that may later be promoted into shared KHS only after gates.

The minimum/pilot lane may provide `kkachi-hermes-skills list/install/doctor/sync/proposal` style support when implemented and evidenced. That lane is not a runner and must not claim KAB control, KHC command/control, Doksuri integration, or backend session authority.

### 6.3 What a Hermes commander or worker receives

For a KHS-governed run, KHS should provide or render:

- Task contract: desired state, constraints, non-goals, acceptance criteria, risk class, required capabilities, source refs.
- Phase skill guidance: the phase-specific rules for plan, ask, implement, enhance-test, optimize, docs-update, external feedback intake, team review, final verify, and improve.
- Prompt/command draft: backend-specific prompt and command guidance, preserving backend caveats.
- Evidence plan: expected KAH artifacts, KAB session evidence, tests, review outputs, report fields, and fail-closed triggers.
- Report contract: final status, changed files, accepted/rejected/deferred items, evidence paths, remaining risk, and next action.
- Improvement-note path: where a lesson belongs, and why it is run-local, project-overlay, backend prompt profile, phase skill reference, script, or shared KHS candidate.

### 6.4 Bounded evidence citations

Conversation, Kanban, and history may be used only as bounded evidence:

- cite the card/session/artifact id;
- state what was accepted, deferred, rejected, or still stale;
- name freshness and implementation caveats;
- prefer exact paths, command outputs, artifacts, and review verdicts over informal memory;
- do not treat a vague phrase such as “what we discussed so far” as sufficient authority unless explicit source ids are provided.

## 7. KAH connection model

KHS must connect to KAH through validated command/artifact paths, not by creating a parallel state model.

| Path | Meaning | Authority |
|---|---|---|
| `.kkachi-workflow.yaml` | project-level workflow graph/policy instance | authoritative only after KAH validation/init/propose/apply/audit evidence |
| `.kkachi/config.yaml` | KAH helper runtime configuration | helper config only; never workflow graph authority |
| `.kkachi/runs/<run_id>/phase-plan.yaml` | run-local KHS execution state/evidence | source of truth for one run’s phase execution state |
| `.kkachi/runs/<run_id>/checklist.md` | operator-facing progress tracker | human progress view derived from phase policy, task contract, and evidence |
| `.kkachi/runs/<run_id>/...` artifacts | task contract, selected backend, prompt, plan, bridge events, tests, reviews, final report, improvement notes | run evidence persisted/validated by KAH |

Required behavior:

- Use KAH graph validate/explain/diff/propose/apply/audit paths when project workflow graph state is involved.
- Use KAH artifact/event/gate paths for run-local state and evidence.
- Fail closed when KAH support is absent, stale, unknown, or conflicting.
- Do not use direct authoritative YAML hand-edit fallback as normal guidance.
- Do not store live counters such as `current_round`, `performed_count`, or `remaining_rounds` in `.kkachi-workflow.yaml`.
- Store actual/current run state in `phase-plan.yaml`, `checklist.md`, KAH events, and run artifacts.
- If `.kkachi-workflow.yaml`, `phase-plan.yaml`, templates, and KHS registries conflict, stop for responsible-role confirmation rather than guessing.

## 8. KAB connection model

For full execution-runtime KHS runs, KAB is the backend runtime and evidence layer.

KHS must preserve these distinctions:

- Backend selection is based on required capabilities, project policy, effective compatibility evidence, user preference only after gates, and backend caveats.
- KAB `send` success is dispatch evidence only.
- Completion evidence requires `wait + read/status`, retained stream/event reconciliation plus final snapshot, or an equivalent documented KAB evidence path.
- Plan lifecycle evidence is separate from tool approval/question evidence.
- Backend-specific caveats must remain visible in `selected-cli.json`, `capability-check.md`, `prompt.md`, `bridge-events.md`, and the final report.
- Automated different-tool review is KAB-later unless an effective KAB implementation and evidence path exist.

### 8.1 KAB adoption stages for KAS/KAH development

KAS owns the operating policy for applying KAB to KAS/KAH development runs. KAH and KAB are tools used by that policy: KAH records deterministic run state and gates, while KAB executes and observes backend sessions. The stages below describe KAS maturity, not a change in KAH or KAB ownership.

| Stage | Name | KAS/KAH development execution lane | Backend choice posture | Evidence posture |
|---|---|---|---|---|
| 1 | Direct Codex app-server baseline | Direct Codex app-server calls handle plan, implementation, feedback fixes, task-bound docs, cleanup, and verification support. | Codex is fixed by the direct lane. | Record direct Codex prompt/session/output evidence and a no-KAB-Codex rationale. Do not claim KAB Codex execution evidence. |
| 2 | KAB Codex-first execution | Replace direct Codex app-server calls with KAB-backed Codex execution through `native_codex` while preserving the same KAS/KAH phases and review scenarios. | Codex remains the default planning/implementation backend; do not broaden backend selection yet. | Record KAB Codex session, selected CLI/capability, plan/read/status/wait/watch or retained event evidence, and bridge events. |
| 3 | KAB backend-selected execution | Run planning/implementation/fix lanes through a selected KAB backend. | Select among eligible Codex, Claude, and GLM backends from task requirements, project policy, compatibility evidence, and user preference after capability gates. | Record the selected backend's KAB evidence and rejected-backend reasons; fail closed when capability evidence is missing or stale. |

Stage 2 is a transport migration from direct Codex app-server to KAB Codex. Stage 3 is a backend orchestration policy. Do not collapse these stages: proving KAB Codex-first execution is the prerequisite for safely expanding backend selection.

Official GLM Octo review is an independent feedback/review lane. It remains KAB GLM `/octo:review` with its existing trigger policy, preflight, prompt-confirmation, watcher/readback, feedback artifact, and post-Octo re-review requirements. Selecting GLM as a possible Stage 3 implementation backend does not satisfy or replace official GLM Octo review, and running official GLM Octo review does not imply GLM was the implementation backend.

The active stage is a project/profile operating setting, not just prose in a
single phase skill. KAS install or project application must choose the stage,
and an existing project may change stage only through an explicit KAS
reconfiguration record with preserved evidence. KAH does not need a
stage-specific graph or gate model because the graph state, run state, events,
locks, artifact schemas, and gate mechanics are unchanged by the implementation
backend transport. KAS records the chosen stage in installed/project-specific
KAS guidance, preferably as a small project-stage reference under the installed
umbrella project skill, such as
`skills/<project>/<project>-kas/references/kab-adoption-stage.md`. KAS may also
mirror the stage into project backend policy or overlay/reference material when
project state is initialized through KAH, and must record it in run-local
task/phase evidence. If the stage is missing or ambiguous, KAS must fail closed
to Stage 1 behavior: direct Codex evidence may be recorded, but KAB Codex
execution must not be claimed.

## 9. Repository structure and artifact roles

Current and target KHS repo structure:

```text
skills/
  kkachi-orchestrate/
  kkachi-task-contract/
  kkachi-backend-select/
  kkachi-prompt-compose/
  kkachi-phase-state/
  kkachi-plan/
  kkachi-ask/
  kkachi-implement/
  kkachi-enhance-test/
  kkachi-optimize/
  kkachi-docs-update/
  kkachi-request-feedback/
  kkachi-handle-feedback/
  kkachi-review/
  kkachi-verify/
  kkachi-final-verify/
  kkachi-improve/
  kkachi-install-guide/

registries/
  task-taxonomy.yaml
  phase-contracts.yaml
  backend-selection-policy.yaml
  backend-prompt-profiles.yaml
  cli-capabilities.yaml
  improvement-promotion-policy.yaml

templates/
  run-artifacts/
  prompts/
  project-overlay/

docs/
  README.md
  sot/
  roadmap.md
  discussions/
```

### 9.1 `skills/`

`skills/` contains phase and support guidance loaded into Hermes profiles. Each skill must be narrow enough to guide a phase without pretending to be an executable state system.

Target phase mapping:

- intake/orchestrate: classify work, choose KHS activation, choose path A/B, assemble initial packet.
- task-contract: create AI-neutral desired state, constraints, non-goals, acceptance criteria, and context refs.
- backend-select: select KAB backend from capability evidence and policy.
- prompt-compose: render backend-specific prompt while preserving caveats.
- phase-state: keep run-local phase status aligned with KAH artifacts.
- plan and ask: produce plan, identify decisions, and preserve plan evidence before implementation.
- implement, enhance-test, optimize, docs-update: execute the work through appropriate backend/session and preserve evidence.
- request-feedback and handle-feedback: external feedback intake layer, bounded by current policy and capability evidence.
- review and verify: read-only checks, risk review, and evidence review.
- final-verify: verify artifacts/gates/reports before completion.
- improve: classify lessons and propose improvements without auto-promotion.

### 9.2 `registries/`

`registries/` is the machine-readable policy/catalog layer.

- `task-taxonomy.yaml`: AI-neutral task types and contract fields.
- `phase-contracts.yaml`: canonical phase spine, applicability, required evidence, and gates.
- `backend-selection-policy.yaml`: how to rank eligible backends after capability/policy gates.
- `backend-prompt-profiles.yaml`: backend prompt style/caveat guidance.
- `cli-capabilities.yaml`: derived support labels from KAB public compatibility evidence; semantic guidance only, not raw callability truth.
- `improvement-promotion-policy.yaml`: destination and gate rules for self-improvement proposals.

Registries must use support labels such as `planned`, `candidate`, `implemented`, `verified`, `stale`, and `superseded` where applicable. A registry entry must not imply current implementation unless effective evidence exists.

### 9.3 `templates/`

`templates/` contains reusable shapes for KAH run artifacts and backend prompts. Templates are not state authority until instantiated and validated/persisted through KAH.

Important run artifact templates include:

- `task-contract.yaml.tmpl`
- `phase-plan.yaml.tmpl`
- `checklist.md.tmpl`
- `plan.md.tmpl`
- `prompt.md.tmpl`
- `selected-cli.json.tmpl`
- `capability-check.md.tmpl`
- `bridge-session-snapshot.json.tmpl`
- `bridge-events.md.tmpl`
- skill QA / evaluation templates

### 9.4 `docs/sot/`

`docs/sot/` stores durable source-of-truth records. New SOTs should state scope, authority level, evidence, ASIS facts, TOBE target, known gaps, accepted/deferred/rejected points, and verification gates.

### 9.5 Project overlays and proposals

Project overlays should capture project-specific docs maps, backend policy, test commands, conventions, and local prompt adjustments. They are the preferred first destination for project-specific lessons.

Shared KHS promotion should happen only after generalized evidence, evaluation, review, and approval.

### 9.6 Thin CLI/binary surface

A thin `kkachi-hermes-skills` surface is acceptable only for scoped KHS pack operations when implemented and evidenced. Candidate verbs include:

- `list`: inspect available skills/categories and installed profile state.
- `install`: profile-scoped copy install with dry-run, manifest/checksum, changed-path report, and recovery path.
- `doctor`: inspect profile/project/KAH/KAB readiness without mutating.
- `sync`: compare installed copy to source pack with dry-run, diff, explicit approval, and recovery path.
- `validate-pack`: validate KHS pack structure, manifests, links, and stale labels.
- `propose-improvement`: create proposal/evidence records without auto-mutating shared KHS.

This CLI must not claim `run`, KAB backend session control, KHC command/control, or Doksuri integration unless separately authorized and evidenced.

## 10. Development workflow KHS should support

KHS target workflow should support these phases or roadmap states, with current status labels preserved:

| Workflow item | Target meaning | Current stance |
|---|---|---|
| planning | create task contract, phase plan, evidence plan, and plan.md | active seed behavior |
| vetting | review plan/risk/acceptance before implementation | known gap / roadmap state |
| implementation | apply accepted plan through backend/session evidence | full execution-runtime requires KAB |
| enhance-test | add or improve targeted tests where useful | active seed behavior |
| remove-ai-slop / optimize | bounded cleanup after behavior is protected | active seed behavior under `optimize` |
| update-docs | update durable docs or record no-docs decision | active seed behavior |
| external-feedback-intake | ingest external/different-tool or user-supplied feedback, triage, handle, record accepted/rejected/deferred | MVP is user-supplied `feedback.md`; automation is KAB-later |
| team-review | KHC role review after implementation/feedback evidence | required gate for non-trivial conclusions; not collapsed with external feedback |
| ready-for-commit | technical readiness, final report, PR title/summary, user approval status | target / roadmap state |
| github-gate | PR/open/merge/release handling | manual or separately assigned; not automatic KHS authority |

Two review layers must remain distinct:

1. External feedback layer: different tool, external AI, or user-supplied `feedback.md` review. KHS records source/ref/checksum where possible, triages feedback, accepts/rejects/deferred items, and sends valid fixes through implementation/verification.
2. KHC team review layer: Blue synthesis plus required Red/Orange/Gray/Teal-if-UI consultation according to KHC rules.

KHS must not merge these into one indistinct “feedback loop.”

## 11. External feedback intake child section

External feedback intake is one child policy under this broader SOT, not the entire KHS architecture.

Current intended policy:

- `EXTERNAL_FEEDBACK_INTAKE` bounds are `min=1`, `max=5`.
- Round 1 is required.
- Rounds 2 through 5 are optional continuation only when useful and when external feedback exists.
- Five rounds are never mandatory.
- If no actionable feedback exists, record the required intake/handling outcome and skip optional rounds with reasons.
- Current MVP path is user-supplied `feedback.md`: the user may run an outside review, save it as `feedback.md`, and ask Blue/Hermes to triage and route accepted fixes.
- Automated `review-by-different-tool` is KAB-later and must not be claimed as current KHS-controlled support.

Authority placement:

- Project-level policy/bounds belong in KAH-validated `.kkachi-workflow.yaml` when supported.
- Actual/current round state belongs in run-local `phase-plan.yaml`, `checklist.md`, KAH events/artifacts, and feedback triage/handling artifacts.
- `.kkachi/config.yaml` is not graph policy authority.

Known stale surfaces before support claim, confirmed by repository search during this review pass:

| Surface | Confirmed marker | Required status before support claim |
|---|---|---|
| `README.md:108` | Historical `request-feedback(1..3) / handle-feedback(1..3)` marker | updated by INITDOC-002 to min=1/max=5 semantics |
| `registries/phase-contracts.yaml:72-73` | `minimum_rounds: 1`, `maximum_rounds: 3` | updated by STALECLEAN-002 to `maximum_rounds: 5`; support label remains `kah-evidenced, kas-integration-pending` until end-to-end adoption is verified |
| `registries/phase-contracts.yaml:140` | `request_feedback_2/3 and handle_feedback_2/3 ... must not exceed round 3` | updated by STALECLEAN-002 to conditional continuation rounds 2..5 |
| `registries/phase-contracts.yaml:202-203` | `conditional_feedback_rounds_2_to_3` | updated by STALECLEAN-002 to `conditional_feedback_rounds_2_to_5` |
| `registries/phase-contracts.yaml:222` | `Feedback must run at least once and at most three times` | updated by STALECLEAN-002 to min 1 / max 5 with optional rounds 2..5 |
| `registries/phase-contracts.yaml:286-293` | request-feedback intent says optional extra rounds up to three; applicability says `rounds_2_to_3`; outputs stop at `feedback-3.md` | updated by STALECLEAN-002 with optional request-feedback outputs through `feedback-5.md` |
| `registries/phase-contracts.yaml:304-305` | handle-feedback outputs stop at `feedback-triage-3.md` / `handle-feedback-3.md` | updated by STALECLEAN-002 with handle-feedback outputs through round 5 |
| `templates/run-artifacts/task-contract.yaml.tmpl:59-61` | `feedback_rounds: min: 1 / max: 3` | updated by STALECLEAN-002 to min 1 / max 5 |
| `templates/run-artifacts/phase-plan.yaml.tmpl:64-65` | `min_rounds: 1`, `max_rounds: 3` | updated by STALECLEAN-002 to `max_rounds: 5` |
| `templates/run-artifacts/phase-plan.yaml.tmpl:170-188` | explicit `request_feedback_3` / `handle_feedback_3` maximum-round shape | updated by STALECLEAN-002 with optional request/handle rows for rounds 4 and 5 |
| `templates/run-artifacts/checklist.md.tmpl:41-42` | `request_feedback_3` / `handle_feedback_3` as maximum feedback round | updated by STALECLEAN-002 with optional request/handle rows for rounds 4 and 5 |
| `docs/sot/phase-orchestration-policy.md` feedback loop | historical max-three wording | updated by STALECLEAN-003 while preserving `kah-evidenced, kas-integration-pending` until end-to-end adoption is verified |
| `docs/sot/phase-orchestration-policy.md` final verification | historical unconditional KAB evidence requirement | updated by STALECLEAN-003 to require KAB evidence when the run is KAB-backed or claims backend runtime evidence |
| `docs/sot/skill-template.md:356` and `581` | historical rounds 2-3 only wording | updated by STALECLEAN-004 to optional continuation rounds 2..5 |
| `docs/sot/concept.md:64` | historical feedback may run up to three rounds wording | updated by STALECLEAN-004 to max 5 optional-continuation policy |
| `docs/sot/concept.md:690` | historical rounds 2-3 only wording | updated by STALECLEAN-004 to optional continuation rounds 2..5 |
| `skills/kkachi-final-verify/SKILL.md:15` | active-policy check could be misread as fixed round count | already active-policy wording; no STALECLEAN-004 edit required |
| `skills/kkachi-orchestrate/SKILL.md:35` | at most three rounds | updated by STALECLEAN-002 to at most five rounds |
| `skills/kkachi-plan/SKILL.md:80` | rounds 2-3 conditional; do not exceed three pairs | updated by STALECLEAN-002 to optional continuation rounds 2..5 and max five pairs |
| `skills/kkachi-request-feedback/SKILL.md:15` and `21` | never exceed three pairs; optional `feedback-2.md` / `feedback-3.md` only | updated by STALECLEAN-002 to optional `feedback-2.md` through `feedback-5.md` and max five pairs |
| `skills/kkachi-handle-feedback/SKILL.md:21-22` | optional round 2/3 handling artifacts only | updated by STALECLEAN-002 to optional handling artifacts through round 5 |

Operator impact: until these stale KAS surfaces are fixed or explicitly marked historical, generated reports, checklists, and support labels must describe `min=1, max=5` external feedback intake as `kah-evidenced, kas-integration-pending` and must not present it as implemented KAS runtime support.

Child record: `docs/sot/external-feedback-intake-policy.md` is the current detail SOT for configurable external feedback intake. It locks the min=1/max=5 policy, negative tests, stale-surface remediation manifest, and KAH/KAS/KAB evidence boundary. That child SOT references this architecture SOT rather than replacing it.

Preconditions not yet met for an active KAS support claim:

- stale surface inventory above is remediated or marked historical with evidence;
- KAS registry/template/skill adoption exists for the KAH-evidenced graph/configurable-feedback substrate; operator report/e2e adoption remains integration-pending until verified;
- KAB evidence exists before automated different-tool review is claimed;
- operator report fields and allowed values in this SOT are implemented or clearly labeled integration-pending;
- verification results are recorded in the child SOT or a registry/manifest reviewed by the responsible roles.

## 12. Self-improvement governance

KHS self-improvement follows this loop:

```text
capture -> classify -> evaluate -> propose -> review -> approve -> version/deploy -> measure
```

### 12.1 Capture

Every material lesson must cite evidence:

- Kanban task/comment id;
- KAH run artifact path;
- KAB session/read/status/event evidence;
- test/verification output;
- review verdict;
- operator report or user correction.

### 12.2 Classify

Route each lesson to the narrowest safe destination:

- run artifact: one-off or insufficient evidence;
- project overlay: project-specific convention or local backend policy;
- backend prompt profile: repeated backend-specific prompt adaptation;
- phase skill reference: reusable phase guidance too long for `SKILL.md`;
- script/check: deterministic repeated validation;
- shared KHS: generalized, reviewed, evaluated, approved behavior.

### 12.3 Evaluate

Shared KHS candidates require evaluation appropriate to risk:

- trigger eval when activation/description changes;
- with-skill vs baseline comparison for material behavior changes;
- assertion grading when objective outputs exist;
- qualitative review when evidence is non-deterministic;
- stale/superseded labels for rejected or deferred proposals.

### 12.4 Propose, review, approve

A single run may create a run-local improvement note or a project-overlay proposal. Shared KHS promotion requires evidence, evaluation, Blue fit, Red risk review, Orange operator value review, Gray traceability review, Teal when UI/UX surfaces are affected, and responsible-approver/user approval for shared behavior changes.

### 12.5 Version/deploy/measure

A promoted shared KHS change must record version, changed files, install/sync impact, rollback/recovery path, and measurement plan. It must not silently mutate installed profiles or shared skills.

## 13. Toy-package rejection criteria

Reject or redesign KHS work if it:

- is only a broad `SKILL.md` bundle without Delegation Packets, evidence refs, approval clarity, or metrics;
- auto-patches shared KHS from one run;
- uses old chats as hidden truth without citations and freshness checks;
- duplicates KAH state or bypasses KAH gates;
- treats KHS semantic guidance as backend callability;
- treats KAB send/dispatch success as completion;
- selects backends by preference before capability and policy gates;
- collapses final verification, docs update, feedback handling, and review into one prose summary;
- claims current support from planned/candidate docs;
- hides stale surfaces or conflicts instead of failing closed;
- turns product docs into member-name requirements rather than role/layer requirements.

Minimum non-toy structures:

- semantic catalog with trust labels;
- task-pattern library;
- Delegation Packet schema;
- phase/evidence/review contract;
- backend-selection and capability caveat model;
- KAH/KAB evidence/report schema;
- improvement proposal and evaluation ledger;
- stale/superseded markers;
- fail-closed recovery taxonomy.

## 14. ASIS to TOBE migration plan

### Step 1: Confirm this SOT

- Red review: risk, fail-closed, premature support claim prevention.
- Orange review: operator usability and report clarity.
- Gray review: traceability, authority ladder, stale/superseded markers.
- Teal review: only if UI/screen-flow/visual hierarchy surfaces are added.
- Blue final synthesis: accepted/deferred/rejected points and remaining risks.

### Step 2: Separate child policy SOTs where needed

Recommended child records:

- `docs/sot/external-feedback-intake-policy.md` for min=1/max=5 external feedback intake details.
- Future delegation packet/evidence loop SOT if the schema grows beyond this architecture document.
- Future evaluation/KAE boundary SOT if KAE integration becomes implementation scope.

### Step 3: Update stale KHS surfaces in small tasks

Do not bundle broad implementation. Create one PR-candidate task per surface group:

1. docs/README and README support/status wording, mapped to the stale inventory row for `README.md:108` and this docs index;
2. `registries/phase-contracts.yaml` and task taxonomy alignment, mapped to stale inventory rows for `phase-contracts.yaml:72-73` and `202-203`;
3. run artifact templates, mapped to stale inventory rows for `task-contract.yaml.tmpl`, `phase-plan.yaml.tmpl`, and `checklist.md.tmpl`;
4. phase skills for plan/orchestrate/request-feedback/handle-feedback/final-verify/phase-state/improve, mapped to the skill rows in the stale inventory;
5. legacy/background SOT wording in `phase-orchestration-policy.md`, `skill-template.md`, and `concept.md`, either updated, superseded, or marked historical;
6. project overlay and CLI/pilot lane docs if support labels surface there;
7. KAH dependency and capability checks only after KAH implementation evidence exists;
8. KAB-later automation only after KAB surfaces exist.

### Step 4: Verify support claims

Before KHS claims support for any target behavior, run the acceptance matrix in this SOT and the relevant child SOT.

## 15. Acceptance and verification matrix

| Area | Required evidence before support claim | Verification owner / evidence record |
|---|---|---|
| Authority | SOT reviewed by Blue, Red, Orange, Gray; Teal if UI/UX applies | Blue closure on the parent card with Red/Orange/Gray verdicts named |
| ASIS/stale state | search results show stale/conflicting references updated or explicitly marked historical | Gray traceability review; record search commands/results in child SOT or registry manifest |
| Delegation Packet | task contract, context refs, capability needs, prompt draft, evidence plan, report contract, improvement destination are represented in templates or guidance | Blue+Orange review of operator contract; KHS template/skill diff evidence |
| KHS/KAH boundary | KHS uses KAH artifact/graph/gate paths and fails closed when missing/unknown/stale | Red risk review plus KAH command/help/capability evidence |
| KHS/KAB boundary | KHS requires KAB evidence for backend execution; `send` success alone is not completion | Red review plus KAB session/read/status/event evidence when runtime support is claimed |
| `.kkachi-workflow.yaml` | KAH validation/proposal/apply/audit evidence exists for project graph changes | KAH audit path and command output; no direct authoritative YAML hand-edit fallback |
| `.kkachi/config.yaml` | no workflow graph authority wording remains | Gray search verification against docs/templates/skills |
| `phase-plan.yaml` | run-local state/evidence is preserved and checked against project graph when graph-managed | KAH run artifact path and phase-plan validation output |
| `checklist.md` | operator tracker shows phase, state, evidence artifact, gate/check, skip reason | KAH run artifact path and Orange operator review |
| External feedback intake | min=1/max=5 policy, round 1 required, rounds 2..5 optional, no five-mandatory wording | Child external-feedback SOT or registry manifest with search evidence |
| External feedback stale surfaces | `1..3`, `max_rounds: 3`, `maximum_rounds: 3`, `round 3`, and `three` in feedback context are removed, updated, or marked historical | Gray traceability review after each surface-group task; evidence recorded in child SOT or stale-surface manifest |
| MVP feedback path | user-supplied `feedback.md` source/ref/checksum and triage/handle artifacts exist when used | KAH run artifacts or operator report with file path/ref/checksum |
| KAH dependency | KAH 0.1.4 capability/help evidence advertises graph and configurable-feedback support; KHS still fails closed when the effective binary, graph policy, or run-local state is missing, unknown, stale, or conflicting | Red+Gray review of KAH implementation evidence and audit output |
| KAB-later review | automated different-tool review remains deferred unless KAB implementation/evidence is present | Red review of KAB evidence and Blue deferral/acceptance record |
| Team review | external feedback and KHC team review are separate gates with separate evidence | Parent/child Kanban review cards, final report, and evidence paths |
| Self-improvement | improvement note/proposal/evaluation/review/approval/version/measure path exists before shared KHS mutation | responsible-approver/user approval record plus proposal/evaluation ledger |
| Operator report | final report exposes current status, evidence paths, accepted/rejected/deferred items, boundary labels, user action required, next valid command, and remaining risk | Orange operator review against section 16 field contract |

## 16. Operator report contract

A KHS-supported run should report exact fields and allowed values clearly enough that an operator can distinguish graph/config evidence, run-local phase/checklist state, feedback ingestion/handling, backend evidence, and final team review.

Minimum field contract:

| Field | Meaning / allowed values |
|---|---|
| `parsed_intent` | concise task interpretation and source ref |
| `current_status` | `ready`, `needs_plan`, `awaiting_user`, `awaiting_review`, `blocked`, `failed_closed`, `done`, `candidate_only` |
| `selected_work_path` | e.g. direct Hermes, KHS+KAH minimum/pilot lane, KHS+KAH+KAB runtime lane, docs-only SOT lane |
| `execution_mode` | `read_only`, `docs_only`, `code_change`, `review_only`, `proposal_only`, `runtime_execution` |
| `workflow_config.status` | `found`, `created_with_confirmation`, `missing_blocked`, `invalid_blocked`, `conflict_blocked` |
| `workflow_config.path` | root `.kkachi-workflow.yaml` when graph-managed; otherwise `not_applicable` with reason |
| `workflow_config.source` | existing file, template/policy version, proposal id, or `missing` |
| `kah_graph_evidence.status` | `validated`, `explain_generated`, `not_available`, `unsupported`, `conflict_blocked` |
| `kah_graph_evidence.artifact` | `graph-evidence.md` when graph state affects the run or graph-managed workflow was requested |
| `kah_graph_evidence.template_id` / `template_path` / `template_version` | source graph template identity when initialized or selected |
| `kah_graph_evidence.proposal_id` / `proposal_path` | KAH graph proposal identity when a graph change is proposed |
| `kah_graph_evidence.semantic_diff_output_path` | KAH semantic diff output path when diff was used |
| `kah_graph_evidence.validation_report_path` / `explain_report_path` | KAH graph validation and explain report paths |
| `kah_graph_evidence.approval_evidence_ref` / `audit_evidence_path` | approval and audit evidence for applied graph changes |
| `kah_graph_evidence.graph_checksum` / `graph_version` | graph identity after init/validate/apply, as reported by KAH evidence |
| `kah_graph_evidence.kah_graph_audit_event_ids` | KAH audit event ids for graph init/propose/apply/audit events |
| `kah_graph_evidence.capability_check_evidence` | reference to `capability-check.md` plus effective-binary `capabilities --json` and `graph --help` captures |
| `phase_plan.status` | `created`, `loaded`, `validated`, `updated`, `missing_blocked`, `not_applicable`; include `phase_plan.path` |
| `checklist.status` | `created`, `loaded`, `updated`, `validated`, `missing_blocked`, `not_applicable`; include `checklist.path` |
| `feedback_rounds.bounds` | `min=1, max=5` for the target policy; label `kah-evidenced, kas-integration-pending` until stale KAS surfaces are closed and active KAS adoption is verified |
| `feedback_rounds.required` | `round_1` |
| `feedback_rounds.optional` | `rounds_2_to_5_optional_continuation_only` |
| `feedback_rounds.current` | current handled/requested round number, or `not_applicable` |
| `feedback_rounds.next_allowed_round` | next valid round number, `none`, or `blocked` |
| `next_stop_reason` | `no_actionable_feedback`, `max_reached`, `awaiting_external_feedback`, `awaiting_user`, `blocked_by_stale_support`, `not_applicable`, or concise custom reason |
| `feedback_source` | MVP user-supplied `feedback.md` path/ref/checksum when present; otherwise source kind and reason |
| `per_round_status` | for each applicable round: `requested`, `ingested`, `triaged`, `handled`, `skipped`, `not_applicable`, or `blocked` |
| `evidence_or_artifact` | path/id for KAH artifact, KAB session/read/status/event evidence, test output, review card, or user-supplied source |
| `boundary_label` | `KHS policy/template`, `KAH graph/run evidence`, `KAB backend evidence`, `user-supplied feedback`, `KHC review`, `project overlay`, `shared KHS proposal` |
| `accepted_items` / `rejected_items` / `deferred_items` | explicit feedback/review/task decisions with reasons |
| `tests_or_verification` | command/artifact path, `not_run` with reason, or `not_applicable` |
| `user_action_required` | `none`, `confirm_graph_create`, `supply_feedback_md`, `answer_question`, `approve_promotion`, `review_required`, or concise custom action |
| `next_valid_command` | next safe command/card/action, or `none` |
| `fail_closed_reason` | required when status is blocked/failed: missing KAH support, invalid graph, conflict, stale support, missing evidence, or other concrete reason |
| `repair_or_confirmation_action` | exact repair path or confirmation needed before continuing |
| `remaining_risk` | explicit risk statement; use `none_known` only when evidence supports it |

Minimal report shape:

```yaml
parsed_intent: "triage user-supplied feedback.md for KHS run <run_id>"
current_status: failed_closed
workflow_config:
  status: conflict_blocked
  path: .kkachi-workflow.yaml
kah_graph_evidence:
  status: conflict_blocked
  artifact: graph-evidence.md
  capability_check_evidence: capability-check.md
feedback_rounds:
  bounds: "min=1, max=5 (kah-evidenced, kas-integration-pending until stale KAS surfaces close)"
  current: 1
  next_allowed_round: blocked
next_stop_reason: blocked_by_stale_support
boundary_label: ["KHS policy/template", "KAH graph/run evidence", "user-supplied feedback"]
user_action_required: review_required
next_valid_command: "repair workflow graph through KAH propose/apply path, then rerun validation"
```

Product pass condition: a Blue/operator should be able to decide `start / ask / block / downgrade` within about two minutes from the packet/report.

## 17. Accepted, deferred, and rejected points

### Accepted

- KHS target definition is Delegation Packet + Evidence Loop + proposal-gated self-improvement.
- KHS/KAH/KAB/KHC boundaries must remain explicit.
- `.kkachi-workflow.yaml` is project-level workflow graph/policy only after KAH validation/audit.
- `phase-plan.yaml` and `checklist.md` carry run-local execution state and operator tracking.
- External feedback intake is a child section under the broader KHS architecture.
- Current MVP external feedback path is user-supplied `feedback.md`.
- The minimum/pilot CLI lane is scoped and must not become KAB/KHC/Doksuri runtime authority.

### Deferred

- Automated different-tool review.
- KHS-controlled backend switching for independent review.
- Backend capability-based automatic review selection.
- KAB-managed multi-session review evidence.
- KAE-backed scoring/routing beyond evaluation references.
- UI/UX-specific review unless a concrete UI surface appears.

### Rejected

- Current operational support claims without stale-surface cleanup and KAH/KAB evidence.
- Five mandatory external feedback loops.
- Direct authoritative `.kkachi-workflow.yaml` hand-edit fallback.
- Live counters in project graph policy.
- `.kkachi/config.yaml` as graph authority.
- KHS as a second state system or runtime bridge.
- Shared KHS auto-promotion from a single run.

## 18. Review gate for this SOT

This draft was routed through the required review gate before Blue closure:

- Red `t_f94bca58`: `REQUEST_CHANGES`; required exhaustive stale-surface inventory and verification owner/evidence path.
- Orange `t_8a3459c0`: `REQUEST_CHANGES`; required explicit operator status/report field contract and support-claim clarity.
- Gray `t_29a6f341`: `REQUEST_CHANGES`; required exact evidence paths, path-transition record, `docs/README.md` index update, confirmed stale marker wording, and dirty-repo context.
- Teal: skipped because no UI/screen-flow/visual hierarchy surface was introduced.
- Blue: final synthesis on `t_f29b6ee9` must name consulted roles, verdicts, accepted/deferred/rejected points, remaining risks, and next action.

Until Blue closure exists on `t_f29b6ee9`, this document remains a candidate SOT draft and development base, not final acceptance evidence. After that closure, implementation status still requires the acceptance matrix and child-policy evidence; the document itself is not proof of operational support.
