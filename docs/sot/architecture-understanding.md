# Kkachi Multi-Skill Architecture Understanding

Date: 2026-04-27
Owner: KHS maintainers
Status: understanding memo, not an implementation plan

## Purpose

Record the master's intended direction before designing or creating any skills, so the responsible coordinator does not drift into making a single monolithic skill or acting before confirming the plan.

## Big picture

17th Earth has multiple active software projects:

- `/Users/draccoon/Workspace/SeventeenthEarth/sudal/sudal-server`
- `/Users/draccoon/Workspace/SeventeenthEarth/sudal/sudal-app`
- `/Users/draccoon/Workspace/SeventeenthEarth/sudal/sudal-web`
- `/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-bridge`
- `/Users/draccoon/Workspace/SeventeenthEarth/doksuri/doksuri`

They share a development operating model, but differ by stack and project conventions. Common stacks include Go, Flutter, Rust, TypeScript, Python, server development, frontend development, and native or desktop app work.

## Operating taxonomy

Kkachi uses two work paths:

```text
Path A: Development Execution
  Existing SOT basis exists.
  The commander executes implementation, review, verification, docs update, and final reporting.

Path B: Discovery / Shaping / SOT / Roadmap
  SOT basis does not yet exist or is insufficient.
  The commander creates or updates SOT, roadmap trace, acceptance criteria, and implementation handoff before development.
```

Each path defaults to Standard Mode.

Light Mode is a constrained reduction profile inside each path. It is not a separate path and not a bypass.

Kkachi is SOT-first. Kkachi does not provide a Path C or hotfix bypass path. Urgent work still requires SOT basis before implementation. If no SOT basis exists, the commander runs Path B-Light or Path B-Standard before implementation.

Metadata vocabulary:

```yaml
work_path:
  - A_development_execution
  - B_discovery_shaping
work_mode:
  - standard
  - light
urgency:
  - normal
  - urgent
  - critical
sot_policy:
  - existing_sot_basis
  - minimal_sot_before_code
  - full_sot_before_code
execution_mode:
  - production_write
  - adapter_qa
  - readiness_hardening
  - research
  - verification
  - docs_only
```

Urgency is metadata, not a separate path. Path A normally uses `existing_sot_basis`; Path B normally creates `minimal_sot_before_code` or `full_sot_before_code` before implementation.

## Source inspiration

The starting reference is:

`/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-helper-oc`

But the target is not OpenCode-only. `kkachi-agent-bridge` supports backend lanes for Claude Code, GLM, Codex CLI, Gemini CLI, and OpenCode, subject to the current public compatibility matrix and project policy.

The architectural focus for Kkachi is therefore not basic backend connectivity. The focus is capability-aware orchestration, artifact preservation, review gates, documentation traceability, and self-improvement across those backend lanes.

## External lesson from revfactory/harness

`revfactory/harness` is useful as a reference for skill craftsmanship, skill testing, trigger evaluation, QA methodology, and evolution loops. It should not be copied as Kkachi's runtime model.

Applicable lessons:

- keep `SKILL.md` lean through Progressive Disclosure;
- separate conditional details into `references/`;
- bundle repeated deterministic checks into `scripts/`;
- evaluate skill descriptions with `should_trigger` and near-miss `should_not_trigger` cases;
- compare with-skill outputs against no-skill or previous-version baselines;
- grade objective outputs with assertions;
- preserve iteration workspaces instead of overwriting eval results;
- use Integration coherence verification checks to catch boundary mismatches;
- feed real run deltas into skill/template/project-overlay improvements.

Non-adopted parts:

- Claude Code-only runtime assumption;
- experimental Agent Teams API dependency;
- default parallel multi-agent write execution;
- marketing quality-improvement claims as Kkachi evidence.

Kkachi remains a multi-backend, evidence-first software delivery harness with one active write lane per project by default.

## Intended skill architecture

The master does not want one giant skill that tells every general to do all development work. KHS also must not become the default path for every small Hermes coding edit.

KHS activation should be explicit or context-bound: use it when the master names KHS/Kkachi/KAH/KAB, asks to apply KHS to a project, requests a Kkachi run, assigns software development to a KHS-using execution owner, or needs bridge evidence, gates, red-team review, or durable KAH run artifacts. Simple direct Hermes edits, small one-file fixes, typo/config patches, and read-only explanations should stay outside KHS unless explicitly delegated into it.

Once KHS is activated, KAH is mandatory for deterministic state and gates; KHS should use KAH `@latest` rather than pinning a single helper patch version.

The intended structure is:

```text
phase skills + project overlay skills + orchestration skill
```

### 1. Phase skills

Split the reusable workflow by phase. Candidate phases:

- intake / task understanding
- SOT basis check
- discovery / shaping
- roadmap trace update
- planning
- ask / clarify blockers
- implementation handoff
- implementation
- test enhancement
- test execution
- optimization
- documentation update
- feedback request
- feedback handling
- red-team review
- UI or screenshot QA when applicable
- final QA / final gate
- user-facing Korean report
- self-improvement capture

The MVP does not need one skill per micro-phase. `kkachi-discovery-shaping` may cover Path B behavior initially. The orchestrator should select path and mode first, then enforce the artifact and gate requirements for that path/mode.

Each phase skill should be reusable across projects and backend agents. Each phase skill must also be backend-capability-aware: it should not assume that all bridge backends support the same event model, approval path, question path, retained event stream, restart recovery, or observability level.

Each phase skill must carry its own quality boundary:

- concise `SKILL.md` with core behavior only;
- `references/` for conditional details;
- `scripts/` for repeated deterministic checks;
- trigger evals for description boundaries;
- execution evals for material behavior;
- assertion-based grading where objective artifacts exist.

A phase skill is not ready for broad use until its trigger boundary is tested against near-miss cases and its critical outputs have at least minimal assertions.

### 2. Project overlay skills

Start from a common project-neutral template, then tune it per project.

Each project can eventually have its own overlay skill that defines:

- repo path
- stack and build commands
- test commands
- architecture rules
- artifact expectations
- backend policy and forbidden backend states
- commander general
- red-team partner
- UI or server runtime needs
- project-specific gates
- project-specific integration coherence checks
- project-specific near-miss trigger risks
- project-specific QA boundary pairs, such as API ↔ client hook, DB schema ↔ API response, route file ↔ link target, compatibility ledger ↔ capability registry
- project-specific script candidates for repeated deterministic checks

Expected overlays include:

- sudal-server
- sudal-app
- sudal-web
- kkachi-agent-bridge
- doksuri

### 3. Orchestration skill

A higher-level integration skill should coordinate the phase skills and project overlays.

Its job is to:

- identify the target project
- choose the project overlay
- choose the real commander general
- choose the red-team partner
- classify the request as Path A or Path B
- classify the mode as Standard or Light
- record urgency as metadata, not as a separate path
- enforce SOT basis gate before implementation
- enforce roadmap trace gate for new work
- reject any urgent-work bypass that starts implementation before SOT and roadmap gates
- choose the correct artifact requirements for path + mode
- decide which phase skill to run next
- choose Claude / GLM / Codex / Gemini / OpenCode lanes through kkachi-agent-bridge using capability evidence
- require the commander to match task requirements against backend capabilities before lane selection
- enforce artifact and gate rules
- enforce skill QA gate when a run proposes or patches shared skills/templates
- require trigger eval updates when a skill description changes
- require with-skill vs baseline comparison for material skill behavior changes
- require assertion-based grading when objective artifacts can be checked
- route integration coherence verification to the correct phase skill or red-team lane
- prevent skill/template patches that weaken SOT-first, roadmap trace, backend evidence, docs-update, or final-gate requirements
- stop and ask the master when scope or decisions are unclear

Minimum backend-selection inputs:

- required write capability
- required approval control
- required question/answer flow
- required retained event stream
- required tool observability level
- restart/recovery sensitivity
- execution mode: production write vs adapter-QA/readiness/research/verification/docs-only
- project-specific forbidden backend states

Example backend policy:

```yaml
backend_policy:
  production_supported:
    - claude
    - glm
    - codex
    - opencode

  adapter_qa_only:
    - gemini

  disallow_when:
    gemini:
      - production_write_lane
      - question_flow_required
      - retained_events_required

  require_when:
    opencode_supervised_approval:
      - opencode_permission_mode_ask
      - real_needs_approval_evidence
```

## General operating rules

- Use real named Hermes team-member profiles when a general is assigned.
- Do not replace a named general with an ephemeral simulated subagent.
- User-facing reports to the master are Korean.
- Inter-general work, red-team review, kkachi-agent-bridge prompts, plans, logs, test reports, feedback, QA artifacts, and reusable artifacts are English.
- The commander owns delivery and gate responsibility.
- The red-team partner challenges each critical phase.
- `request-feedback` is separate from red-team review and should use an independent feedback lane.

## Important correction

The responsible coordinator must not design or create the skills alone before confirming the master's intended structure.

Before creating, editing, or installing actual skills, the responsible coordinator must ask the master and wait for confirmation.

## Current agreed direction

The master confirmed:

1. Split lower-level skills by phase.
2. Create a project-neutral base skill template first.
3. Tune the base template per project so each project eventually has its own skill.
4. Treat the result as a phase skill plus project overlay skill structure.
5. Name the whole workflow **Kkachi**.
6. Treat `kkachi-agent-bridge` as the bridge layer, not the whole product.
7. Treat Kkachi as a Hermes Agents software development harness composed of Hermes skills, `kkachi-agent-bridge`, deterministic helper CLI, external AI CLIs, real generals, red-team gates, and self-improvement loops.
8. Distribute Kkachi as multiple independently versioned repositories: bridge, helper, and Hermes skill templates.
9. Treat `kkachi-agent-bridge/docs/public/compatibility-matrix.md` as the human-readable source ledger for backend tested support and caveats.
10. Maintain `kkachi-hermes-skills/registries/cli-capabilities.yaml` as the machine-readable registry derived from or synchronized with that compatibility matrix.
11. Require each run to preserve `selected-cli.json`, `capability-check.md`, and bridge session identity evidence.
12. Treat Gemini according to the current KAB public compatibility matrix; Gemini is a supported backend and KHS should use retained stream observation for first-class consumption.
13. Kkachi uses two work paths: Path A for SOT-based development execution, and Path B for discovery/shaping/SOT/roadmap work.
14. Standard Mode is the default mode for both paths.
15. Light Mode is a constrained reduction profile, not a separate path and not a bypass.
16. Kkachi does not provide Path C or a hotfix bypass path.
17. Urgent work still requires SOT basis before implementation.
18. If no SOT basis exists, the commander must run Path B-Light or Path B-Standard before implementation.
19. Documentation must describe Standard Mode first, then Light Mode reductions.
20. KHS is a Hermes prompt/process skill pack, not the deterministic state ledger and not the bridge control plane.
21. Task contracts are AI-neutral. Backend prompts are rendered artifacts derived from task contracts, phase contracts, project overlays, selected backend capability snapshots, and backend prompt profiles.
22. The current v0.2 delivery spine is `plan -> ralplan -> implement -> enhance test -> ai slop cleaner -> optimize -> update docs -> request feedback -> handle feedback`, with explicit approval/question evidence recorded at real boundaries instead of a mandatory normal `ask` phase.
23. User backend preference is a ranking input after capability and project-policy gates pass; it must not override a required capability mismatch.
24. Self-improvement lands project-local first, then promotes to shared KHS only after repeated evidence and skill QA.

## Kkachi product definition

Kkachi is not another single-provider coding agent. It is a meta-harness that lets real Hermes team members command strong existing CLI coding tools through `kkachi-agent-bridge`.

Core motto:

```text
Don't Reinvent the Wheel
```

The current backend lane model is:

- supported lanes tracked by the bridge compatibility matrix: Claude Code, GLM, Codex CLI, Gemini CLI, and OpenCode
- adapter-QA/readiness lane: only for a backend or feature explicitly marked non-production by the compatibility matrix or project policy
- other future CLI lanes if they fit the bridge contract and are represented in the capability registry

Kkachi's intended responsibility split:

```text
Responsible coordinator / orchestrator:
  receive the master's command, identify the target project and assigned execution owner, delegate with source references, monitor gates, preserve process integrity, and report to the master in Korean; do not author the execution owner's substantive plan/checklist unless explicitly assigned as that execution owner

Hermes commander general and skills:
  read the roadmap/SOT/docs/code, understand task, match task requirements against backend capabilities, select CLI lane, record selected-cli/capability/bridge evidence, generate precise English prompts, author plan.md and checklist, enforce gates, review, run e2e, manage project docs, and submit evidence

kkachi-agent-bridge:
  expose external AI coding TUIs as programmable backend sessions with explicit backend identity (`backend_type` and `adapter_type`) and capability/caveat evidence sourced from the compatibility matrix

external AI CLIs:
  perform the actual repository editing, code generation, and implementation work

red-team generals:
  challenge design, diff, verification, release risk, and blind spots

self-improvement / skill QA loop:
  use real run evidence to improve prompts, phase skills, review gates, e2e checklists, and project overlays over time
```

## Repository distribution model

Kkachi should be distributed as three independently versioned GitHub repositories.

Recommended repositories:

```text
kkachi-agent-bridge
  Runtime integration layer between Hermes generals and external AI CLI tools.

kkachi-agent-helper
  Deterministic CLI for status, event logs, run artifacts, locks, schema checks, diagnostics, and project initialization.

kkachi-hermes-skills
  Hermes Agent prompt/process skill pack: phase skills, task contracts, backend prompt profiles, prompt templates, project overlay templates, schemas, and registries.
```

Rationale:

- bridge, helper, and skills have different responsibilities and release cadence
- bridge changes when external CLI protocols change
- helper changes when deterministic state/artifact management changes
- skills change when workflow, prompts, review gates, or project overlay conventions improve
- separate repositories reduce coupling and make version compatibility explicit

KHS's main output during a Kkachi run is not code. Its main output is a disciplined set of artifacts and prompts that Hermes uses to command KAB safely:

```text
task-contract.yaml          # KHS supplemental contract
task-brief.md/context-pack.md # KAH gate-visible contract mirror
selected-cli.json
capability-check.md
prompt.md
bridge-session-snapshot.json
bridge-events.md
verification.md
improvement-note.md
```

KHS must keep the task contract and rendered prompt separate. The task contract says what must be true when the work is done. The rendered prompt says how the selected backend should pursue that contract in the current phase and is stored in KAH's canonical `prompt.md` artifact. Until KAH has a first-class task-contract schema, `task-contract.yaml` remains supplemental and gate-relevant contract content must be mirrored into `task-brief.md` or `context-pack.md`.

Recommended `kkachi-hermes-skills` repository layout:

```text
kkachi-hermes-skills/
  README.md
  VERSION
  skill-pack.yaml

  docs/
    skill-writing-standard.md
    skill-testing-standard.md
    qa-coherence-standard.md

  skills/
    kkachi-orchestrate/
      SKILL.md
      references/
      scripts/
    kkachi-task-contract/
      SKILL.md
      references/
      scripts/
    kkachi-backend-select/
      SKILL.md
      references/
      scripts/
    kkachi-prompt-compose/
      SKILL.md
      references/
      scripts/
    kkachi-phase-state/
      SKILL.md
      references/
      scripts/
    kkachi-plan/
      SKILL.md
      references/
      scripts/
    kkachi-ask/
      SKILL.md
      references/
      scripts/
    kkachi-implement/
      SKILL.md
      references/
      scripts/
    kkachi-enhance-test/
      SKILL.md
      references/
      scripts/
    kkachi-optimize/
      SKILL.md
      references/
      scripts/
    kkachi-review/
      SKILL.md
      references/
      scripts/
    kkachi-verify/
      SKILL.md
      references/
      scripts/
    kkachi-docs-update/
      SKILL.md
      references/
      scripts/
    kkachi-request-feedback/
      SKILL.md
      references/
      scripts/
    kkachi-handle-feedback/
      SKILL.md
      references/
      scripts/
    kkachi-improve/
      SKILL.md
      references/
      scripts/

  templates/
    project-overlay/
      SKILL.md.tmpl
      context-pack.md.tmpl
      docs-map.yaml.tmpl
      references/
        backend-policy.md.tmpl
    run-artifacts/
      task-contract.yaml.tmpl
      plan.md.tmpl
      checklist.md.tmpl
      selected-cli.json.tmpl
      prompt.md.tmpl
      bridge-session-snapshot.json.tmpl
      bridge-events.md.tmpl
      capability-check.md.tmpl
      impl-log.md.tmpl
      test-log.md.tmpl
      verification.md.tmpl
      docs-update.md.tmpl
      redteam/
        plan-review.md.tmpl
        impl-review.md.tmpl
        test-review.md.tmpl
        qa-review.md.tmpl
        shaping-review.md.tmpl
        final-gate-review.md.tmpl
      improvement-note.md.tmpl
      skill-qa/
        skill-change-request.md.tmpl
        trigger-eval-results.json.tmpl
        with-skill-vs-baseline.md.tmpl
        grading.json.tmpl
        comparison.md.tmpl
    prompts/
      claude/
      codex/
      gemini/
      glm/
      opencode/

  evals/
    trigger/
    execution/

  benchmarks/
    <skill-name>/
      iteration-001/
      iteration-002/

  schemas/
    eval-metadata.schema.json
    grading.schema.json
    timing.schema.json
    trigger-eval.schema.json
    benchmark.schema.json

  registries/
    phase-contracts.yaml
    task-taxonomy.yaml
    cli-capabilities.yaml
    backend-prompt-profiles.yaml
    backend-selection-policy.yaml
    improvement-promotion-policy.yaml
    bridge-session-schema.json
    selected-cli-schema.json
    event-schema.yaml
    status-schema.json

  examples/
    kkachi-agent-bridge-overlay/
    sudal-app-overlay/
```

The repository should treat skills as versioned, testable assets. Evaluation files and benchmark workspaces belong in `kkachi-hermes-skills`, while project-local `.kkachi/runs/<run_id>/` stores evidence from real development runs.

Recommended `skill-pack.yaml` role:

```yaml
name: kkachi-hermes-skills
version: 0.1.0
requires:
  kkachi-agent-helper: "latest"
  kkachi-agent-bridge: "latest-or-project-policy"
skills:
  - kkachi-orchestrate
  - kkachi-task-contract
  - kkachi-backend-select
  - kkachi-prompt-compose
  - kkachi-phase-state
  - kkachi-plan
  - kkachi-ask
  - kkachi-implement
  - kkachi-enhance-test
  - kkachi-optimize
  - kkachi-review
  - kkachi-verify
  - kkachi-docs-update
  - kkachi-request-feedback
  - kkachi-handle-feedback
  - kkachi-improve
```

Kkachi-hermes-skills (KHS) never writes directly to project directories. KHS prepares template parameters and invokes kkachi-agent-helper `project init` to perform file operations. KHS does not install Hermes skills; skill installation is handled by the Hermes native skill system.

Recommended project overlay initialization via kkachi-agent-helper:

```bash
kkachi-agent-helper project init \
  --project-name kkachi-agent-bridge \
  --stack go \
  --repo-path /path/to/project \
  --commander responsible-approver \
  --redteam required-reviewer \
  --docs-map-roadmap docs/ROADMAP.md \
  --docs-map-spec docs/SPEC.md \
  --docs-map-architecture docs/ARCHITECTURE.md \
  --test-commands "go test,make test" \
  --backend-policy "claude,codex,opencode" \
  --execution-mode production_write \
  --sot-policy existing_sot_basis
```

Kkachi-agent-helper generates the project config, overlay, docs-map, and schema files from the provided parameters. Reconfiguration uses `--force` and preserves existing runs, artifacts, and gate history while rewriting config, overlay, docs-map, and schema copies; a `project.reconfigured` event is recorded. Re-running `project init` without `--force` on an already initialized project fails fast and does not overwrite existing managed state.

The skills repository provides upstream templates. Local installed skills and project overlays may be customized, but the customization boundary must be explicit:

```text
upstream template:
  kkachi-hermes-skills repository

Hermes skill installation:
  hermes skills install or manual placement under ~/.hermes/skills/

project overlay:
  target project docs/config, created by kkachi-agent-helper project init
```

Each project should have an explicit docs map so commanders do not guess documentation locations:

```yaml
project: kkachi-agent-bridge
roadmap: docs/ROADMAP.md
sot_docs:
  - docs/SPEC.md
  - docs/ARCHITECTURE.md
adr_dir: docs/adr
todo_dir: docs/todo
spec_dir: docs/specs
test_commands:
  - bun test
  - cargo test
```

## Kkachi work paths

Kkachi supports exactly two work paths.

### Path A: Development Execution

The master's command may look like:

```text
Work on TASK-208 from the roadmap file.
```

Path A is used when existing SOT basis exists. Default mode: Standard. Light Mode may be used only as a constrained reduction profile and only when the eligibility rules are satisfied.

Expected flow:

1. The responsible coordinator identifies the target project, roadmap file, and assigned execution owner.
2. The responsible coordinator delegates the work to that execution owner with the roadmap task reference, not with a pre-authored implementation plan.
3. The commander reads:
   - the roadmap entry for the task
   - the SOT documents linked from that roadmap entry
   - the project's standard docs such as architecture, ADR, spec, todo/task docs, and coding rules
   - the relevant source code and tests
4. The commander records `sot-basis.md`, confirms roadmap trace, and derives acceptance criteria from the SOT basis.
5. The commander asks the KAB planner backend for the plan surface when KAB is used, then captures `plan.plan_text` into run-local `plan.md` before approval/start.
6. KHS/Hermes normalizes the planner's `KHS Checklist Seed` plus phase contracts into `checklist.md`, and keeps `phase-plan.yaml`/`checklist.md` current.
7. The commander then drives capability-aware CLI selection, `selected-cli.json` capture, `capability-check.md`, prompt generation, bridge execution, bridge session evidence capture, review, e2e, docs update, feedback loop, and final evidence.

Important boundary:

```text
The KAB planner backend authors the plan surface when a planner lane is used. KHS/Hermes, through the responsible coordinator or assigned execution owner, captures that surface into `plan.md`, normalizes `checklist.md`, and maintains `phase-plan.yaml`/`checklist.md` as workflow evidence.
```

### Path B: Discovery / Shaping / SOT / Roadmap

The master's command may start from an external bug report, an observed improvement need, or a new feature idea rather than an existing roadmap task.

Path B is used when the SOT basis does not yet exist or is insufficient. Default mode: Standard. Light Mode may be used only as a constrained reduction profile and only when the eligibility rules are satisfied.

Expected flow:

1. The responsible coordinator identifies the target project and assigned execution owner.
2. The commander records request classification in `intake-classification.md`.
3. The commander reads existing project docs, roadmap, related SOT files, and relevant code.
4. The commander records existing-docs review, problem framing, research notes if needed, strategy options, and selected strategy.
5. The commander creates or updates the required SOT document before implementation.
6. The commander updates the roadmap file so the new work is traceable.
7. The commander defines acceptance criteria, non-scope, task breakdown, and implementation readiness.
8. If implementation is approved or included in the same run, the commander records `handoff-to-development.md` and enters Path A only after SOT basis and roadmap trace gates pass.

For both paths, roadmap and SOT documents are first-class sources of truth. Kkachi must not treat the code diff, git commit message, chat-only instruction, or agent memory as a sufficient long-term record.

For `research`, `verification`, or `docs_only` execution modes, roadmap trace may be marked not applicable only with an explicit reason recorded in the run artifacts.

## Deterministic helper CLI

A separate helper CLI such as `kkachi-agent-helper` is a good fit, but it must stay deterministic and must not become the intelligence layer.

Recommended responsibilities:

- create run IDs
- create and maintain `.kkachi/status.json`
- append `.kkachi/events.jsonl`
- store `work_path`, `work_mode`, `urgency`, and `sot_policy` in status/run metadata
- create `.kkachi/runs/<run_id>/` directories
- create empty or template-backed artifact paths when needed, but not author the substantive plan or checklist
- store the assigned general's run-local `plan.md`
- store the assigned general's execution checklists in `status.json` and/or checklist artifacts
- store generated prompts, CLI outputs, diffs, test logs, review notes, and verdicts
- create `selected-cli.json` from a validated template
- record selected CLI lane and the reason supplied by the general
- record bridge session identity snapshots
- append bridge lifecycle and pending events into run evidence
- validate that planned, unsupported, or degraded backend capabilities are not used for incompatible production write lanes unless explicitly scoped as adapter QA/readiness
- validate that implementation cannot start without a SOT basis artifact
- validate that new work has roadmap trace before implementation
- validate Light Mode eligibility when Light Mode is selected
- validate that required artifacts exist before phase transition
- fail closed if a run attempts to use a hotfix bypass or code-first path
- fail closed if selected backend capability does not satisfy the run's declared requirements
- record verification commands and results
- record documentation-update requirements and completion evidence
- manage active run and project write locks
- support resume by run ID

Do not put these responsibilities in the helper:

- deciding architecture
- deciding whether the SOT content is product-correct
- authoring the substantive SOT decision by itself
- choosing the CLI lane by itself; it may validate the commander's selected backend against the registry, but the substantive selection reason belongs to the commander general
- writing prompts by itself
- authoring the substantive `plan.md` or checklist for a development task
- silently downgrading Standard Mode to Light Mode
- issuing review verdicts by itself
- silently masking failed commands

The helper should fail closed and expose state clearly.

The helper layout below is a minimal deterministic state layout. The full run artifact model is defined in `concept.md` and project skill templates. When they differ, helper schemas should be synchronized with the full run artifact model before implementation.

Recommended state layout:

```text
.kkachi/
  status.json
  events.jsonl
  active_run.lock
  project_write.lock
  runs/
    <run_id>/
      run-metadata.json
      intake-classification.md
      sot-basis.md
      task-brief.md
      task-contract.yaml        # KHS supplemental
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
      redteam/
        plan-review.md
        impl-review.md
        test-review.md
        qa-review.md
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
```

## Required standards before broad use

Before Kkachi grows into many skills, define these standards first:

1. **Status and artifact schema**
   - `status.json` for current state
   - `events.jsonl` for append-only history
   - run artifact directory for reproducibility and self-improvement
   - run-local `plan.md` and checklist are authored by the commander general after reading roadmap, SOT docs, related docs, and code

2. **CLI capability registry**
   - record each CLI lane's strengths, weaknesses, best use cases, risk areas, caveats, unsupported cases, verified versions, and recommended usage constraints
   - derive from or keep synchronized with `kkachi-agent-bridge/docs/public/compatibility-matrix.md`
   - let generals choose based on this registry, not by vague preference

3. **Prompt and backend evidence artifact standard**
   - preserve task brief, task contract, selected CLI, selection reason, capability snapshot, caveats accepted, unsupported capabilities ruled out, backend prompt profile, rendered prompt, context pack, bridge session snapshot, bridge events, expected output, actual output, changed files, verification result, review verdict, and improvement note
   - treat `task-contract.yaml` as AI-neutral and `prompt.md` as backend-specific
   - never let backend-specific prompt style mutate the task contract's acceptance criteria, constraints, or non-goals

4. **Prompt composition standard**
   - prompt rendering inputs are task contract, work path, phase contract, project overlay, selected backend, capability snapshot, prompt profile, current run state, and required artifacts
   - backend prompt profiles are versioned heuristics and must record source evidence or observed run evidence when changed
   - Claude, Codex, Gemini, GLM, and OpenCode prompt templates may differ in structure, but must all preserve the same acceptance criteria, constraints, non-goals, verification contract, and language policy
   - prompt profiles belong in `registries/backend-prompt-profiles.yaml`; repeated prompt bodies belong in `templates/prompts/<backend>/`

5. **KAH phase state standard**
   - phase skills call KAH through the implemented `run`, `artifact`, `event append`, `schema validate`, `gate check`, canonical `gate final <run_id>`, and `diagnostics export` commands
   - phase milestones are KHS-defined event types such as `phase.started`, `artifact.updated`, and `phase.completed`; gate truth comes from `gate check`
   - KAH owns `status.json`, `events.jsonl`, locks, artifact directories, gate reports, and schema validation
   - KHS may describe the required KAH command sequence but must not write parallel state files or silently bypass helper state transitions

6. **Project context packs and documentation targets**
   - each project overlay should expose architecture, coding rules, test commands, forbidden changes, done definition, and e2e checklist
   - each project overlay should explicitly identify canonical documentation directories and files, including architecture docs, ADR docs, todo or task docs, specification docs, and any other required `docs/` locations
   - if these directories are not known, the skill must ask the master instead of guessing
   - completion should require docs updates when the code or behavior changes; git commit messages and agent memory are not sufficient durable records

7. **Skill QA and evaluation standard**
   - each phase skill has trigger evals;
   - near-miss should-not-trigger cases are mandatory;
   - critical phase skills have execution evals;
   - material skill changes use with-skill vs baseline comparison;
   - objective outputs use assertion-based grading;
   - Eval workspace records preserve iteration history;
   - non-discriminating assertions are removed or strengthened;
   - repeated helper scripts become `scripts/` candidates;
   - long conditional instructions move to `references/`;
   - shared skill, orchestration skill, project overlay template, prompt template, review checklist, or shared documentation-update checklist patches require master approval;
   - the Skill release gate must pass before broad use of material shared behavior-surface changes.

8. **Phase skill templates**
   - start with a small reusable set: `kkachi-task-contract`, `kkachi-backend-select`, `kkachi-prompt-compose`, `kkachi-phase-state`, `kkachi-plan`, `kkachi-implement`, `kkachi-verify`, `kkachi-improve`
   - expand to the full current spine: `plan`, `ralplan`, `implement`, `enhance-test`, `ai-slop-cleaner`, `optimize`, `update-docs`, `request-feedback`, and `handle-feedback`

9. **Red-team gates**
   - risk review before design finalization
   - diff review after implementation
   - failure-mode review before e2e
   - release/block verdict before completion

10. **Self-improvement discipline**
   - improvements must come from run evidence, not speculation;
   - record candidates first;
   - generalize feedback instead of overfitting to one prompt;
   - identify affected surfaces precisely;
   - decide whether the change belongs in `SKILL.md`, `references/`, `scripts/`, templates, registries, project overlays, prompt profiles, or helper checks;
   - apply project-local overlay/process improvements first unless the lesson is already general across projects;
   - patch shared skill templates only after the master's approval;
   - run skill QA when a shared skill/template change is material.

11. **Bridge backend evidence standard**
   - selected backend
   - adapter type
   - source compatibility ledger
   - tested version or required minimum
   - capability snapshot
   - caveats accepted
   - unsupported capabilities ruled out
   - bridge session snapshot
   - pending approval/input evidence when relevant

12. **Compatibility synchronization standard**
   - README summary support view
   - `docs/public/compatibility-matrix.md` canonical tested-support ledger
   - `kkachi-hermes-skills/registries/cli-capabilities.yaml` machine-readable registry
   - project overlay backend policy

## Sequential development policy

The master prefers sequential development as Kkachi's default mode.

Rationale:

- real generals build useful memory as the development story unfolds
- requirement decisions, failed attempts, review standards, and project constraints accumulate in order
- root-cause analysis is cleaner than with multiple parallel worktrees
- self-improvement evidence is easier to trust when prompt, output, review, verification, and improvement notes form one timeline
- red-team gates remain clear and enforceable
- role boundaries are less likely to blur

Default policy:

```text
One active write lane per project.
```

Allowed parallelism by default:

- read-only research
- codebase inspection
- documentation review
- risk review
- test-log analysis
- independent red-team review
- CLI capability comparison

Not allowed by default:

- two simultaneous write runs in the same project
- two generals editing the same feature in different worktrees
- concurrent updates to the same `.kkachi/status.json`
- concurrent bridge sessions that mutate the same repository state
- one real general profile handling conflicting active project sessions

Git worktree mode is not rejected, but it is an advanced explicit mode only. It requires:

- separate run IDs
- separate worktrees and artifact directories
- one named merge owner
- clear isolation from the main write lane
- final sequential integration gate
- explicit approval before use

The helper should enforce this with project-level locks and fail closed when a second write lane is attempted without explicit advanced-mode approval.

## Recommended MVP

Start small rather than building the full system at once.

Recommended MVP:

- pilot project: `kkachi-agent-bridge`
- pilot execution mode: backend implementation, adapter validation, or readiness hardening as selected by current project policy
- available KAB lanes: Claude Code, GLM, Codex CLI, Gemini CLI, OpenCode
- experimental lane: only for backend features explicitly marked experimental or degraded by the compatibility matrix
- helper scope: `status.json`, `events.jsonl`, run directories, locks, selected-cli capture, bridge-session snapshots, capability checks, and artifact existence checks
- skill scope: intake -> plan -> implement -> review -> red-team -> verify -> docs-update -> improve
- registry scope: `cli-capabilities.yaml` synchronized with the bridge compatibility matrix
- project overlay: one minimal overlay for the pilot project, including README, compatibility matrix, requirements, backend SOT/todo docs, test commands, forbidden changes, and backend policy
- red-team: one required review gate before completion
- success condition: one real sequential Gemini-related bridge task completed with plan, checklist, selected backend evidence, bridge session evidence, prompt, output, diff, tests, review, red-team, verification, docs update evidence, and improvement note preserved

## Next discussion needed

Before writing real skills, decide:

- exact `status.json` schema
- exact `events.jsonl` event names
- run artifact directory schema
- plan and checklist artifact schema
- `cli-capabilities.yaml` schema derived from `docs/public/compatibility-matrix.md`
- KHS extension fields for `selected-cli.json`
- KHS extension fields for `bridge-session-snapshot.json`
- KHS prose standard for `capability-check.md` beyond KAH's backend gate requirements
- project context pack schema
- required project documentation paths: README, compatibility matrix, linked SOT documents, requirements, backend todo/SOT docs, architecture, ADR, todo/task, spec, and other docs targets
- commander-vs-orchestrator boundary for who authors `plan.md` and checklists
- backend lane policy for production-supported, degraded, planned, and unsupported states
- backend support-promotion/demotion policy
- exact phase skill list for MVP
- red-team gate minimum for MVP
- KHS version-gating policy for future helper CLI command changes
- `kkachi-hermes-skills` repository structure and `skill-pack.yaml` schema
- skill install/update/version compatibility policy
- orchestration skill responsibilities and stop conditions
- exact Path A Standard artifact set
- exact Path A Light permitted reductions
- exact Path B Standard artifact set
- exact Path B Light permitted reductions
- SOT basis gate schema
- roadmap trace gate schema
- Light Mode eligibility schema
- no-hotfix-bypass enforcement rule
- exact Skill QA repository layout
- trigger eval schema and threshold
- execution eval schema and threshold
- with-skill vs baseline comparison policy
- assertion grading schema
- benchmark workspace retention policy
- non-discriminating assertion handling policy
- skill release gate policy
- skill versioning and compatibility with helper/bridge versions
- when skill changes require master approval
- which skill QA checks are performed by helper, commander, red-team, or external feedback lane
- integration coherence checklist per project overlay


## Gemini lane policy

Gemini is currently supported in the KAB public compatibility matrix. KHS must preserve Gemini-specific caveats while treating Gemini as an eligible backend when project policy permits it.

Rules:

- Gemini production write delivery is allowed when project policy permits it and required capabilities match.
- Gemini plan approval does not start implementation; explicit post-approval start evidence is required.
- Gemini answer/question flow uses the hook-native `BeforeTool ask_user` path.
- Gemini retained events are supported through KAB public `/api/events` and `/api/stream`.
- First-class Gemini clients should open `/api/stream` early and use `/api/events` plus `read`/`status`/`wait`/`read?tui=true` for reconciliation and failure evidence.
- Terminal capture is diagnostic only and must not become the shipped control plane.
- Direct TUI shell bypass must remain disabled; readiness fails closed if it cannot be disabled.

## Support-ledger synchronization rule

Kkachi documents should not become stale copies of backend implementation details. The synchronization rule is:

1. `kkachi-agent-bridge/docs/public/compatibility-matrix.md` owns tested support, tested versions, capability details, and caveats.
2. `kkachi-agent-bridge/README.md` provides the summary support view.
3. `kkachi-hermes-skills/registries/cli-capabilities.yaml` is the machine-readable registry derived from or kept consistent with the compatibility matrix.
4. Project overlays record project-specific backend policy and forbidden states.
5. When a backend status changes, review all four surfaces together.
