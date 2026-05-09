# Kkachi Team Development Skill Template

Date: 2026-04-27
Owner: Gongmyeong
Status: draft
Purpose: reusable SKILL.md template for software development by a Hermes field commander, red-team partner, and kkachi-agent-bridge backend agents.

---

````markdown
---
name: <project>-kkachi-team-development
description: Run <project> software development through SOT-first Kkachi Standard Mode by default, with constrained Light Mode reductions, real Hermes field commanders, red-team partners, kkachi-agent-bridge backend lanes, deterministic gates, integration coherence verification, evidence-based skill improvement, and user-facing Korean reports.
version: 0.1.0
---

# <Project> Kkachi Team Development

Use this skill when the master asks a named field commander such as 조운, 마초, 관우, or another Hermes team member to develop, fix, review, or QA software using KHS/Kkachi, or when the request explicitly says to use KHS, Kkachi, KAH, KAB, a Kkachi run, bridge evidence, or gate-backed artifacts. Do not trigger KHS for simple direct Hermes edits unless the master explicitly asks for KHS/Kkachi or delegates the task to a KHS-using commander.

## Skill structure and Progressive Disclosure

Keep this skill lean. `SKILL.md` should contain the core operating rules, phase sequence, gates, and output requirements. Move conditional details into `references/` and repeated deterministic checks into `scripts/`.

Recommended local structure:

```text
<Project> Kkachi Team Development/
  SKILL.md
  references/
    backend-policy.md
    qa-coherence-checks.md
    light-mode-rules.md
    project-docs-map.md
  scripts/
    check-artifacts.*
    check-selected-cli.*
    check-docs-coherence.*
```

Rules:

- Do not duplicate long backend support tables in `SKILL.md`; use the bridge compatibility matrix and capability registry.
- Do not include general programming knowledge unless the absence of that instruction causes repeated failures.
- If repeated evaluations show agents recreating the same helper script, move that script into `scripts/`.
- If a section becomes long and conditional, move it into `references/` and leave a precise loading instruction.

## Core operating rule

- Kkachi is SOT-first. No implementation may start without a SOT basis.
- There is no Path C or hotfix bypass path. Urgency may reduce artifact depth, but it must not bypass SOT basis, roadmap trace, acceptance criteria, verification, docs-update decision, or final report.
- Standard Mode is the default operating mode. Light Mode is a constrained reduction profile, not a separate path.
- Use the real named Hermes team-member profile. Do not replace it with an ephemeral role-prompted helper.
- The field commander owns delivery and final gate responsibility.
- The red-team partner challenges plan, implementation, tests, QA, feedback handling, and final gate.
- Backend agents produce implementation and analysis through kkachi-agent-bridge lanes selected by capability evidence, not vague preference.
- Deterministic state, phase transition checks, schema validation, and artifact/gate enforcement belong to `kkachi-agent-helper` or the project equivalent, not freeform judgment.
- Kkachi-hermes-skills (KHS) is the Hermes prompt/process skill pack. It builds task contracts, chooses phase contracts, selects prompt profiles, and renders prompts for kkachi-agent-bridge.
- KHS never writes directly to project directories. KHS prepares parameters and invokes `kkachi-agent-helper` for all deterministic project state and artifact operations. KHS does not install Hermes skills; skill installation is handled by the Hermes native skill system (`hermes skills install` or manual placement under `~/.hermes/skills/`).
- Task contracts are AI-neutral. Backend prompts are rendered artifacts. Do not put Claude/Codex/Gemini/GLM/OpenCode prompt style inside acceptance criteria or non-goals.

## Language policy

```text
Direct report to the master: Korean.
Inter-member discussion, red-team review, kkachi-agent-bridge backend-agent instructions, plans, implementation logs, test reports, feedback handling, UI QA reports, and reusable artifacts: English.
```

## Required inputs

Before starting, identify:

- project name and repo path: `<repo>`
- task id or issue id: `<task-id>`
- field commander: `<commander>`
- red-team partner: `<redteam>`
- work path: `<A_development_execution|B_discovery_shaping>`
- work mode: `<standard|light>`
- urgency: `<normal|urgent|critical>`
- SOT policy: `<existing_sot_basis|minimal_sot_before_code|full_sot_before_code>`
- SOT basis path or expected new SOT path
- roadmap path and roadmap item id, if existing
- roadmap update target, if new work
- acceptance criteria source
- target backend lanes: `<codex|claude|glm|opencode|gemini-adapter-qa-only>`
- user backend preference or priority, if provided
- execution mode: `<production_write|adapter_qa|readiness_hardening|research|verification|docs_only>`
- required backend capabilities: write lane, approval control, question/answer flow, retained events, tool observability, restart/recovery sensitivity
- capability registry / source ledger: `registries/cli-capabilities.yaml` and `kkachi-agent-bridge/docs/public/compatibility-matrix.md`
- backend prompt profile: `<claude-stepwise-directive|codex-desired-state|gemini-structured-context|glm-agentic-coding|opencode-control-plane-aware|project-override>`
- artifact root: `<repo>/.kkachi/runs/<run-id>/`
- derived acceptance criteria, if already known
- out-of-scope items
- project overlay version or local overlay path
- skill pack version
- whether this run may propose changes to shared skills/templates
- skill QA requirement, if shared skills/templates are changed
- applicable integration coherence checks for the project stack
- user decisions needed before execution, if any

If repo, task scope, SOT basis, roadmap trace, or acceptance criteria is missing and cannot be established through Path B-Light or Path B-Standard, ask the master before implementation.

Implementation must not start from chat-only instruction.

Path A normally uses `existing_sot_basis`. Path B normally creates `minimal_sot_before_code` or `full_sot_before_code` before implementation.

## KHS / KAH / KAB contract

KHS owns prompt/process guidance. KAH owns deterministic state and artifact recording. KAB owns backend session control and prompt delivery.

KHS never writes directly to project directories. When project-local files are needed (such as `.kkachi/project-overlay.yaml` or `docs/kkachi-docs-map.yaml`), KHS prepares the parameters and invokes `kkachi-agent-helper project init`.

Example invocation from a KHS skill:

```bash
kkachi-agent-helper project init \
  --project-name <name> \
  --stack <go|flutter|rust|typescript|python> \
  --repo-path <path> \
  --commander <commander> \
  --redteam <redteam> \
  --docs-map-roadmap <path> \
  --docs-map-spec <path> \
  --docs-map-architecture <path> \
  --test-commands "<command1>,<command2>" \
  --backend-policy "<lane1,lane2>" \
  --execution-mode <mode> \
  --sot-policy <policy>
```

Kkachi-agent-helper generates the project config, overlay, docs-map, and schema files from the provided parameters. Reconfiguration uses `--force` and preserves existing runs, artifacts, and gate history while rewriting config, overlay, docs-map, and schema copies; a `project.reconfigured` event is recorded. Re-running `project init` without `--force` on an already initialized project fails fast and does not overwrite existing managed state.

KHS does not install Hermes skills. Skill installation is handled by the Hermes native skill system (`hermes skills install` or manual placement under `~/.hermes/skills/`).

Every phase skill must name the KAH command sequence it expects Hermes to run and should assume KAH is installed from `github.com/SeventeenthEarth/kkachi-agent-helper@latest`. KAH does not expose dedicated `phase start` or `phase complete` commands; phase milestones are compact events recorded with `event append`, while gate truth is produced by `gate check`.

```bash
kkachi-agent-helper project init ... [--force] [--json]
kkachi-agent-helper run create ... [--json]
kkachi-agent-helper run activate <run_id> [--json]
kkachi-agent-helper artifact init <run_id> [--json]
kkachi-agent-helper event append <event_type> --run <run_id> --payload '<json-object>' [--json]
kkachi-agent-helper schema validate <file> --schema <config|status|event|run-metadata|selected-cli|bridge-session-snapshot> [--json]
kkachi-agent-helper artifact validate <run_id> [--gate intake] [--json]
kkachi-agent-helper gate check <run_id> <intake|sot|roadmap|plan|backend|implementation|review|verification|docs|final> [--json]
kkachi-agent-helper gate final <run_id> [--json]
kkachi-agent-helper run close <run_id> [--json]
kkachi-agent-helper run abort <run_id> [--json]
```

Every KAB-backed phase must preserve:

```text
task-contract.yaml            # KHS supplemental contract, mirrored into task-brief.md/context-pack.md
selected-cli.json
capability-check.md
prompt.md
bridge-session-snapshot.json
bridge-events.md
cli-output.md
```

KAH canonical artifacts include `prompt.md`; the rendered-prompt content must use that artifact path. `task-contract.yaml` is a KHS supplemental artifact until KAH adds a first-class schema for it; any contract content needed by KAH gates must also be reflected in KAH-managed `task-brief.md` or `context-pack.md`.

KAB prompt dispatch and observation have two valid operating shapes:

```text
cli_loop:
  kkachi-agent-bridge send
  -> kkachi-agent-bridge wait
  -> kkachi-agent-bridge read/status
  -> kkachi-agent-bridge approve|reject|answer as open_pendings require
  -> repeat until terminal or accepted idle state
  -> kkachi-agent-bridge stop

retained_stream:
  kkachi-agent-bridge send
  -> subscribe to GET /api/stream/<session_id> or backfill with GET /api/events/<session_id>
  -> kkachi-agent-bridge approve|reject|answer as stream events expose open pendings
  -> confirm current state with kkachi-agent-bridge read/status
  -> kkachi-agent-bridge stop
```

`send` success means dispatch was accepted; it is not completion evidence. `bridge-events.md` must record whether the run used `cli_loop`, `retained_stream`, or `hybrid`, and stream-backed evidence must record cursor/epoch details when available. Retained event streams are Kkachi-stable public events, not raw backend protocol payloads, and they are not durable replay across daemon restart.

## Task contract and prompt lifecycle

The task contract is the stable, AI-neutral delivery contract:

```text
desired state
acceptance criteria
constraints
non-goals
context sources
required capabilities
verification evidence
```

The rendered prompt is a backend-specific instruction packet derived from the task contract:

```text
task-contract.yaml
  + work_path
  + phase contract
  + project overlay
  + selected-cli.json
  + capability-check.md
  + backend prompt profile
  + current run state
  = prompt.md
```

Backend prompt profile rules:

- Claude-oriented prompts may be more stepwise, explicit, and sectioned.
- Codex-oriented prompts should emphasize desired state, constraints, acceptance criteria, autonomy boundaries, and verification evidence.
- Gemini, GLM, and OpenCode prompts must follow their own project-recorded profile and bridge caveats.
- Prompt profile differences must not change the task contract.
- User backend preference is applied only after capability and project-policy gates pass.

## Role map

```text
Gongmyeong / orchestrator
  - scopes the job
  - delegates to the real field commander
  - classifies path, mode, urgency, SOT policy, and execution mode
  - supervises reports and sends Korean user-facing summary

Field commander: <commander>
  - owns the task
  - creates or confirms SOT basis, roadmap trace, and acceptance criteria
  - creates or confirms plan
  - records the AI-neutral task contract before backend prompting
  - selects backend lanes through kkachi-agent-bridge using capability evidence
  - records selected-cli.json, capability-check.md, prompt.md, bridge-session-snapshot.json, and bridge-events.md
  - sends backend prompts through KAB; does not bypass KAB for external AI CLI work
  - integrates outputs
  - ensures tests, docs, and final gate

Red-team partner: <redteam>
  - reviews every phase
  - records blocking objections
  - does not implement unless explicitly reassigned

Backend agents via kkachi-agent-bridge
  - receive English instructions
  - work in isolated lanes
  - return patches, reasoning, test commands, and risks

External feedback AI
  - independent review, not the same as red-team
  - read-only unless explicitly authorized
```

## Artifact layout

This is the Standard Mode artifact layout. Light Mode uses the same artifact names where possible, but permits shorter content or explicit not-applicable records.

`skill-qa/` is conditional. Create it only when the run proposes, evaluates, or applies changes to shared Kkachi phase skills, orchestration skills, project overlay templates, prompt templates, review checklists, or shared documentation-update checklists.

```text
<repo>/.kkachi/
  status.json
  events.jsonl
  active_run.lock
  project_write.lock
  runs/
    <run-id>/
      run-metadata.json
      intake-classification.md
      sot-basis.md
      task-brief.md
      task-contract.yaml        # KHS supplemental; mirror gate-relevant content into task-brief.md/context-pack.md
      acceptance-criteria.md
      plan.md
      checklist.md
      selected-cli.json
      capability-check.md
      bridge-session-snapshot.json
      bridge-events.md
      prompt.md
      context-pack.md
      cli-output.md
      diff.patch
      impl-log.md
      test-log.md
      verification.md
      review.md
      docs-update.md
      sot-update.md
      roadmap-update.md
      improvement-note.md
      skill-qa/
        skill-change-request.md
        trigger-eval-results.json
        with-skill-vs-baseline.md
        grading.json
        comparison.md
      feedback-request.md
      feedback-1.md
      feedback-triage-1.md
      handle-feedback-1.md
      redteam/
        plan-review.md
        impl-review.md
        test-review.md
        qa-review.md
        shaping-review.md
        final-gate-review.md
      discovery/
        existing-docs-review.md
        problem-framing.md
        research-notes.md
        strategy-options.md
        selected-strategy.md
        task-breakdown.md
        implementation-readiness.md
        handoff-to-development.md
      ui/
        scenarios/
        runs/<ui-run-id>/
          manifest.json
          app-log.txt
          server-log.txt
          daemon-log.txt
          screenshots/
          visual-qa-request.md
          visual-qa-report.md
          redteam-ui-review.md
      final-report.md
```

## Phase sequence

The phase sequence is Standard Mode by default. The orchestrator first classifies the run as Path A or Path B, then applies the Standard flow unless Light Mode is explicitly selected and eligible.

### Path A Standard sequence

```text
0. intake
1. classify path/mode/urgency/sot_policy
2. SOT basis gate
3. roadmap trace gate
4. plan
5. redteam-plan-review
6. ask-if-blocked
7. backend selection
8. prompt-compose
9. implement
10. enhance-tests
11. run-tests
12. optimize
13. update-docs
14. request-feedback round 1 always, rounds 2-3 if Hermes judges more feedback useful
15. handle-feedback
16. rerun-tests-if-feedback-changed
17. ui-qa-if-needed
18. redteam-final-review
19. skill-qa-if-shared-skill-or-template-changed
20. final-gate
21. Korean user report
```

### Path B Standard sequence

```text
0. intake
1. classify path/mode/urgency/sot_policy
2. existing docs review
3. request classification
4. problem framing
5. research if needed
6. strategy options
7. selected strategy
8. create or update SOT
9. update roadmap trace
10. define acceptance criteria and non-scope
11. task breakdown
12. implementation readiness review
13. redteam-shaping-review
14. handoff to Path A if implementation follows
15. skill-qa-if-shared-skill-or-template-changed
16. redteam-final-review
17. final-gate
18. Korean user report
```

Path B may hand off to Path A in the same run or in a follow-up run, but implementation must not start until the SOT basis and roadmap trace gates pass. Path B red-team review focuses on problem framing, SOT correctness risk, roadmap traceability, non-scope, implementation readiness, and whether the handoff can safely enter Path A.

## Phase gates

### intake gate

PASS when:
- repo exists
- task scope is clear enough to classify Path A or Path B
- commander and red-team partner are assigned
- artifact session exists
- `work_path`, `work_mode`, `urgency`, `sot_policy`, and `execution_mode` are recorded

BLOCKED when:
- scope or acceptance criteria are ambiguous and cannot be shaped through Path B
- repo path is missing
- task depends on user decision

### SOT basis gate

PASS when:
- existing SOT / roadmap / spec / ADR / architecture doc / test contract / canonical evidence is identified; or
- Path B has created or updated a minimal/full SOT before implementation;
- acceptance criteria are derived from that SOT basis;
- the SOT basis is recorded in `sot-basis.md`.

FAIL when:
- implementation starts from chat-only instruction;
- expected behavior is not anchored to durable evidence;
- code diff, git commit message, or agent memory is treated as sufficient long-term record;
- a hotfix bypass is attempted.

BLOCKED when:
- the SOT decision requires master's product or architecture decision.

### roadmap trace gate

PASS when:
- existing roadmap task links to the SOT basis; or
- Path B has created or updated the roadmap entry before implementation; or
- commander records why roadmap trace is not applicable for this `research`, `verification`, or `docs_only` run.

FAIL when:
- new feature, improvement, behavior change, or undocumented bug proceeds to implementation without roadmap trace.

BLOCKED when:
- roadmap ownership or target file is unknown.

### plan gate

Plan must include:
- problem statement
- scope and non-scope
- SOT basis and roadmap trace
- acceptance criteria
- affected files/modules
- architecture boundaries
- backend lane requirements and selected lane policy
- test strategy
- rollback condition
- risk list

PASS requires red-team plan review with no blocking objection, unless eligible Path A Light Mode records that plan review is merged into final red-team review.

### backend selection gate

PASS when:
- `selected-cli.json` records `backend_type`, `adapter_type`, execution mode, selection reason, capability snapshot, caveats, unsupported capabilities, and source ledger
- `capability-check.md` maps task requirements to the selected backend capabilities
- planned, unsupported, or degraded backend capabilities are not used for incompatible production write lanes
- Gemini is treated as eligible when project policy permits it and required capabilities match; selected Gemini lanes record explicit post-approval start and stream/reconcile evidence requirements.

FAIL when:
- selected backend lacks a required capability and no safe exception is recorded
- selected backend status contradicts the compatibility matrix
- helper or script chooses the backend without commander reasoning

### implementation gate

PASS when:
- implementation diff matches plan
- no unrelated broad changes
- architecture boundary is preserved
- commander wrote `impl-log.md` or equivalent implementation notes
- red-team implementation review is pass or pass-with-nonblocking-notes

FAIL when:
- implementation starts before SOT basis gate and roadmap trace gate pass

### test gate

PASS when:
- relevant unit/integration/e2e tests ran or targeted verification is justified for Light Mode
- failures are fixed or explicitly classified as unrelated
- `verification.md` records exact commands, results, failures, classification, and verdict;
- `test-log.md` preserves raw command output when tests or verification commands are run;
- applicable integration coherence checks are complete, such as API response ↔ frontend hook type, route path ↔ link target, state transition map ↔ status updates, SOT ↔ plan ↔ diff, or compatibility matrix ↔ cli-capabilities.yaml ↔ selected-cli.json
- red-team test review has no blocking objection, unless an eligible Light Mode reduction merges it into final gate review

### feedback gate

PASS when:
- independent feedback requested when task risk warrants it
- feedback triaged
- accepted feedback handled
- rejected feedback has reason
- skipped feedback has an explicit reason

### skill QA gate

This is the run-local Skill release gate for shared skill/template changes.

This gate applies only when the run proposes, evaluates, or applies changes to shared Kkachi skills, orchestration skills, project overlay templates, prompt templates, review checklists, or documentation-update checklists.

PASS when:
- the improvement is grounded in run evidence, not speculation;
- `improvement-note.md` records the observed failure, generalized lesson, affected surface, proposed patch, evidence, eval requirement, and approval requirement;
- description changes include Trigger eval updates, including Near-miss should-not-trigger cases;
- material behavior changes include With-skill vs baseline comparison;
- objective outputs include Assertion-based grading;
- repeated helper logic is moved to `scripts/` or recorded as a bundling candidate;
- long conditional guidance is moved to `references/`;
- shared skill, orchestration skill, project overlay template, prompt template, review checklist, or shared documentation-update checklist changes have master approval before patching shared repositories.

FAIL when:
- a shared skill is patched from a one-off anecdote without generalization;
- the change only overfits a single eval prompt;
- `SKILL.md` gains bulky conditional content that belongs in `references/`;
- the change weakens SOT-first, roadmap trace, backend evidence, docs-update, or final-gate requirements;
- trigger changes cause obvious overlap with another Kkachi phase skill.

BLOCKED when:
- master approval is required but not yet granted;
- Eval workspace fixtures or baseline snapshots are missing;
- the proposed change affects helper/bridge compatibility but version requirements are not known.

### final gate

PASS when:
- deterministic gate returns PASS
- SOT basis, roadmap trace, acceptance criteria, verification, docs-update decision, and final report are complete
- `docs-update.md` was checked after the final code/diff state, not before later feedback or UI QA changes
- final red-team review has no blocker
- integration coherence verification has no unresolved blocker for changed boundaries
- final-report is complete
- Korean user report can be produced from artifacts

FAIL when:
- required tests fail
- red-team has unresolved blocker
- artifact set is incomplete
- implementation started before required SOT and roadmap gates

BLOCKED when:
- user decision is needed
- external dependency is unavailable
- task scope changed materially

## Light Mode policy

Light Mode is not a separate path. Light Mode is a constrained reduction profile applied inside Path A or Path B.

Standard Mode remains the default. Documentation and templates must describe Standard Mode first.

### General Light Mode eligibility

Light Mode may be selected only when:

- the scope is small and bounded;
- risk is low;
- SOT basis can be established before implementation;
- acceptance criteria are clear;
- no major architecture, schema, API, permission, security, billing, or migration decision remains open;
- targeted verification is adequate.

### General Light Mode non-skippable gates

Light Mode must not skip:

- SOT basis gate;
- roadmap trace gate for new work;
- acceptance criteria;
- backend capability gate when bridge lanes are used;
- verification;
- docs-update decision;
- final report.

### Path A Light permitted reductions

- `plan.md` may be concise.
- red-team plan review may be merged into final red-team review.
- feedback round 1 is mandatory; additional rounds 2-3 may be skipped with explicit reason.
- `context-pack.md` may be minimal.
- test scope may be targeted.
- UI QA may be marked not applicable.

### Path B Light permitted reductions

- existing docs review may be concise.
- external research may be omitted when unnecessary.
- strategy options may be reduced.
- SOT may be a minimal SOT stub.
- roadmap update may create a single task entry.
- task breakdown may be a single implementation task.
- red-team shaping review may be merged into final gate.

## Integration coherence verification

QA must prioritize connection correctness over isolated existence checks.

General rule:

```text
Read both sides of every changed boundary.
```

Kkachi coherence pairs:

```text
SOT basis ↔ acceptance criteria
acceptance criteria ↔ plan.md
task-contract.yaml ↔ selected-cli.json
plan.md ↔ prompt.md
prompt.md ↔ actual diff
selected-cli.json ↔ capability-check.md
capability-check.md ↔ bridge compatibility matrix
bridge-session-snapshot.json ↔ bridge-events.md
docs-update.md ↔ actual behavior change
```

Project-specific examples:

```text
API route response ↔ frontend hook type
DB schema field ↔ API response field ↔ client type
route file path ↔ href/router.push/redirect target
state transition map ↔ actual status update code
backend support ledger ↔ README summary support view
```

A check is weak if it only confirms that each side exists. A check is strong if it confirms that connected sides agree.

## Backend lane selection guide

Use `registries/cli-capabilities.yaml`, `kkachi-agent-bridge/docs/public/compatibility-matrix.md`, and the project overlay's backend policy. Do not select lanes by vague preference.

Minimum selection inputs:

- required write capability
- required approval control
- required question/answer flow
- required retained event stream
- required tool observability level
- restart/recovery sensitivity
- execution mode: production write vs adapter-QA/readiness/research/verification/docs-only
- project-specific forbidden backend states
- user backend preference or priority, applied only after the required capability and project policy gates pass

Current backend policy summary belongs in `references/backend-policy.md`, not in the main `SKILL.md`. This keeps backend maturity details out of the executable template and reduces stale-copy risk.

`references/backend-policy.md` must state:

```markdown
This backend summary is advisory. If it conflicts with `kkachi-agent-bridge/docs/public/compatibility-matrix.md` or `registries/cli-capabilities.yaml`, the compatibility matrix and synchronized registry win.
```

Keep only stable selection principles here:

- choose lanes from declared task requirements plus capability evidence;
- do not choose lanes by vague model preference;
- do respect explicit user preference as a ranking input among capable, policy-allowed lanes;
- planned, unsupported, or caveated backend features must not be used for incompatible production writes unless the compatibility matrix and project policy allow that exact use;
- Gemini is eligible when the compatibility matrix and project policy allow it; prefer retained `/api/stream` observation with `/api/events` plus snapshot reconciliation;
- for high-risk changes, use one active implementation/write lane plus one read-only critique/review lane;
- the critique/review lane must not mutate the repository unless explicit advanced-mode approval is granted.

## Standard commander prompts

### Backend implementation prompt

```text
You are a backend implementation lane controlled by <commander> through kkachi-agent-bridge.
Use English only.
Repo: <repo>
Task: <task-id>
Artifact run: <artifact-path>
Work path: <A_development_execution|B_discovery_shaping>
Work mode: <standard|light>
Urgency: <normal|urgent|critical>
SOT policy: <existing_sot_basis|minimal_sot_before_code|full_sot_before_code>
Selected backend: <backend_type>/<adapter_type>
Execution mode: <production_write|adapter_qa|readiness_hardening|research|verification|docs_only>
Capability source ledger: <compatibility-matrix path>

Read:
- sot-basis.md
- acceptance-criteria.md
- task-brief.md
- plan.md
- redteam/plan-review.md when present

Implement only the approved scope.
Preserve Clean Architecture, SRP, existing conventions, and testability.
Write a concise result with:
- files changed
- rationale
- tests run or proposed
- risks
- blockers
Do not produce user-facing Korean.
```

### Red-team review prompt

```text
You are the red-team partner for this phase.
Use English only.
Review the artifact and current diff for blockers.
Focus on SOT alignment, roadmap trace, correctness, architecture boundary, SRP, tests, edge cases, rollback risk, and hidden scope expansion.
Return:
- verdict: PASS / FAIL / BLOCKED
- must-fix
- should-check
- evidence
- reproduction or verification steps
```

### Skill QA grader prompt

```text
You are the Skill QA grader for a Kkachi skill/template change.
Use English only.
Evaluate the proposed skill/template change against the provided assertions.

Inputs:
- skill-change-request.md
- improvement-note.md
- changed skill/template files
- eval metadata
- with_skill outputs
- baseline outputs
- timing data, if available
- grading assertions

Return:
- verdict: PASS / FAIL / BLOCKED
- assertion results with evidence
- non-discriminating assertions, if any
- overfitting risks
- missing trigger evals
- missing baseline comparison
- recommended patch or rejection reason
```

### Skill QA comparator prompt

```text
You are the blind comparator for Kkachi skill evaluation.
Use English only.
Compare Output A and Output B without knowing which one used the skill.

Judge:
- SOT alignment
- artifact completeness
- backend evidence quality
- phase-order correctness
- verification usefulness
- documentation-update decision quality
- user-facing report readiness

Return:
- preferred output: A / B / tie
- reason
- critical deficiencies
- whether the improvement is material enough to justify the skill change
```

## Commander final report format

The final artifact `final-report.md` is English and must include:

- run id
- repo path
- task id
- work path, work mode, urgency, SOT policy, and execution mode
- SOT basis and roadmap trace
- acceptance criteria
- phase history
- backend lanes used
- selected backend identity and selection reason
- capability snapshot and caveats accepted
- bridge session snapshot summary
- files changed
- verification commands, raw test-log reference, and result verdict
- docs-update decision and post-final-diff check
- red-team verdicts
- feedback rounds
- UI QA result, if any
- integration coherence verification result
- skill/template improvement candidates
- skill QA result, if shared skills/templates were changed
- final verdict: PASS / FAIL / BLOCKED
- remaining risks

The user-facing report to the master is Korean and must include only:

```text
결론: PASS / FAIL / BLOCKED
작업 범위:
변경 요약:
검증:
레드팀 판정:
스킬/템플릿 개선:
남은 결정 사항:
산출물 경로:
```

If there is no skill/template improvement candidate or Skill QA result, write:

```text
스킬/템플릿 개선: 해당 없음
```

## Failure handling

- If SOT basis is missing: run Path B-Light or Path B-Standard before implementation, or ask the master if the SOT decision requires ownership input.
- If roadmap trace is missing for new work: update the roadmap before implementation, or mark not applicable only for research/verification/docs-only with explicit reason.
- If backend lane stalls: stop that lane, preserve logs, start a new lane with the latest artifact context.
- If red-team blocks: commander must fix, narrow scope, or escalate to the master.
- If tests cannot run: mark BLOCKED unless there is an explicit accepted substitute verification.
- If artifact and code diverge: artifact is stale; update artifact before final gate.
- If selected backend capability does not satisfy declared run requirements: fail closed, choose another supported lane, or escalate before implementation.
- If Gemini is selected for production write delivery before support promotion: fail closed and choose a production-supported lane instead.
- If a named member is requested but wrapper/profile is unavailable: stop and report the blocker. Do not simulate.
- If an urgent run attempts implementation before gates: fail closed and return to Path B or Path A gates.
- If a shared skill/template change lacks run evidence: record it as a rejected or deferred improvement candidate, not as a patch.
- If a skill description change causes trigger overlap with another phase skill: fail the skill QA gate and refine the boundary.
- If with-skill and baseline both pass all assertions: mark the assertions as non-discriminating and strengthen the eval before claiming improvement.
- If a repeated helper script appears across evals but is not bundled: record a scripts/ bundling candidate.
- If QA checks only verify existence and not connected-surface agreement: fail the integration coherence portion of the test or final gate.

## Completion checklist

- [ ] work_path recorded
- [ ] work_mode recorded
- [ ] urgency recorded
- [ ] sot_policy recorded
- [ ] execution_mode recorded
- [ ] SOT basis identified or created before implementation
- [ ] roadmap trace confirmed or updated before implementation, or explicit not-applicable reason recorded for research/verification/docs-only
- [ ] acceptance criteria derived from SOT basis
- [ ] no hotfix bypass used
- [ ] Light Mode eligibility recorded, if Light Mode used
- [ ] real field commander profile used
- [ ] real red-team partner used or explicit exception recorded
- [ ] run artifact directory created
- [ ] selected-cli.json written in English/JSON with backend_type and adapter_type when a bridge backend lane is used
- [ ] capability-check.md written before implementation when a bridge backend lane is used
- [ ] bridge-session-snapshot.json captured when bridge session starts
- [ ] bridge-events.md captured when relevant
- [ ] plan.md written in English
- [ ] redteam/plan-review.md written in English or merged into final review under eligible Light Mode
- [ ] backend lane prompts sent in English through a capability-approved lane
- [ ] impl-log.md or equivalent implementation notes written in English
- [ ] verification.md written in English
- [ ] test-log.md captured when test or verification commands ran
- [ ] docs-update decision recorded after final code/diff state
- [ ] feedback handled or explicitly skipped with reason
- [ ] UI QA completed or explicitly not applicable
- [ ] final red-team review complete
- [ ] integration coherence checks completed for changed boundaries
- [ ] skill/template improvement candidate recorded when real evidence suggests reusable improvement
- [ ] skill-qa/ artifacts created when shared skills/templates or shared behavior-surface checklists were changed
- [ ] trigger eval updated when a skill description changed
- [ ] with-skill vs baseline comparison completed for material skill behavior changes
- [ ] assertion-based grading completed when objective checks are possible
- [ ] master approval recorded before shared skill, orchestration skill, project overlay template, prompt template, review checklist, or shared documentation-update checklist patch
- [ ] non-discriminating assertions removed or strengthened
- [ ] repeated helper scripts identified for `scripts/` bundling when applicable
- [ ] long conditional guidance moved to `references/` when applicable
- [ ] final gate PASS / FAIL / BLOCKED recorded
- [ ] Korean user report delivered
````

---

## Notes for adapting this template

- For Doksuri, default commander is Kwanwoo and default red-team partner is Hahuyeon.
- For sudal-app mobile work, default commander is Macho and default red-team partner is Seohwang.
- For strategic cross-domain review, include Samaui as an additional red-team advisor.
- For kkachi-agent-bridge QA itself, Gongmyeong handles QA directly except OpenCode/Codex exclusions previously set by the master.
