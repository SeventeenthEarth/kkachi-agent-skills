# Kkachi Harness Concept

Date: 2026-04-27
Owner: Gongmyeong
Status: concept document for readers who do not know Hermes Agent

## 1. What is Kkachi?

Kkachi is a software development harness built around Hermes Agent team members, reusable Hermes skills, `kkachi-agent-bridge`, deterministic helper tooling, and strong external AI coding CLIs.

Kkachi is not a new single-provider coding agent. It is a meta-harness for coordinating existing coding tools and real AI team-member profiles in a disciplined software delivery workflow.

Core motto:

```text
Don't Reinvent the Wheel
```

Instead of rebuilding what Claude Code, GLM, Codex CLI, OpenCode, and Gemini CLI can do well, Kkachi lets Hermes team members select and command those tools through a bridge, then review, verify, document, and improve the result. The bridge is no longer treated as the main MVP risk: Kkachi's main responsibility is capability-aware backend selection, evidence preservation, gate enforcement, documentation traceability, and self-improvement.

## KHS position

`kkachi-hermes-skills` (KHS) is the seed prompt/process skill pack that Hermes Agent uses to prepare and govern Kkachi runs. KHS does not execute external AI CLIs directly and does not own deterministic project state. Instead, KHS teaches Hermes how to classify work, build AI-neutral task contracts, select a backend lane, render backend-specific prompts, and preserve the artifacts needed for review and self-improvement.

The Kkachi responsibility boundary is:

```text
KHS:
  prompt/process skills, phase contracts, backend prompt profiles, artifact templates, phase_plan workflow SOT, QA and improvement rules

Hermes Agent:
  manager/orchestrator judgment, risk approval routing, KAH/KAB invocation, checklist upkeep, final verification, and Korean reporting

KAH:
  deterministic run IDs, status, events, locks, artifact directories, schema validation, and gates; KAH metadata is helper classification, not workflow authority

KAB:
  backend session creation, plan surface capture, prompt delivery, approval/input/event control, and bridge session evidence

External AI CLI:
  substantive planning, implementation, documentation, feedback, or review work inside the selected lane
```

KHS is intentionally a starting point. Project-local overlays and run evidence improve the local process first. Only repeated, generalized, evidence-backed improvements should be promoted back into shared KHS skills, templates, or registries.

## KHS activation policy

KHS should trigger for a Kkachi-governed software run, not for every Hermes coding interaction.

Trigger KHS when the master explicitly asks for KHS/Kkachi/KAH/KAB, asks to apply KHS to a project, requests a Kkachi run, delegates development to a KHS-using commander such as 조운 or 마초, or asks for bridge-backed implementation with durable artifacts, gates, red-team review, or backend evidence.

Do not trigger KHS by default for simple direct Hermes edits such as typo fixes, small one-file changes, quick config patches, or read-only explanations/reviews. Those can proceed without KAH unless the master explicitly requests KHS/Kkachi or assigns the work to a KHS-using commander.

Once KHS is triggered, KAH is mandatory for deterministic state, artifacts, events, locks, schemas, diagnostics, and gates. KHS must use the installed KAH `@latest` command surface and must not create a parallel state system.

User-confirmed orchestration policy:

- The master selects the target roadmap task id or task item for each KHS run.
- KHS creates `phase-plan.yaml` as the workflow SOT; KAH `work_path`, `work_mode`, and `execution_mode` are helper classification metadata only.
- Hermes is manager, risk approval router, and final verifier; KAB backend roles do substantive plan/code/docs/feedback work.
- KHS code-change runs use KAB. If the master forbids KAB for code changes, treat the request as a normal direct Hermes task rather than a KHS run.
- Docs-only KHS runs use KAB by default; direct docs edits require explicit no-KAB instruction and recorded rationale.
- `ask`, feedback round 1, matching feedback handling, and final verification are mandatory.
- Feedback may run up to three rounds. `optimize` is strongly recommended for code changes and requires a skip reason when omitted.

## Task and prompt model

Kkachi separates the durable task contract from the prompt sent to an AI backend.

```text
Task contract:
  AI-neutral desired state, acceptance criteria, constraints, non-goals, context sources, required capabilities, and verification evidence

Rendered prompt:
  backend-specific instruction packet generated from the task contract, path, phase, project overlay, selected backend, capability snapshot, and prompt profile
```

The prompt lifecycle is:

```text
master request
  -> path and phase classification
  -> task-contract.yaml as KHS supplemental contract
  -> KAH task-brief.md/context-pack.md mirror for gate-relevant contract content
  -> backend capability and policy check
  -> selected-cli.json
  -> backend prompt profile selection
  -> prompt.md
  -> KAB prompt delivery
  -> result capture and verification
  -> prompt/process improvement note
```

Prompt profiles are backend-specific because different AI CLIs respond better to different instruction shapes. Claude-oriented profiles may use more explicit stepwise direction and structured sections. Codex-oriented profiles should emphasize desired state, constraints, acceptance criteria, autonomy boundaries, and verification evidence. Gemini, GLM, and OpenCode profiles must encode their own observed strengths, caveats, control-plane behavior, and bridge-specific limits. These are versioned heuristics, not permanent model truths.

## Phase state model

KAH records every run, artifact baseline, event, gate report, and final run state. KHS phase skills must describe when Hermes should ask KAH to update state, but KHS must not invent state out of band.

Implemented state command shape:

```bash
kkachi-agent-helper run create ... --json
kkachi-agent-helper run activate <run_id> --json
kkachi-agent-helper artifact init <run_id> --json
kkachi-agent-helper event append phase.started --run <run_id> --payload '<json-object>' --json
kkachi-agent-helper event append artifact.updated --run <run_id> --payload '<json-object>' --json
kkachi-agent-helper schema validate <file> --schema <selected-cli|bridge-session-snapshot> --json
kkachi-agent-helper artifact validate <run_id> --gate intake --json
kkachi-agent-helper gate check <run_id> <gate> --json
kkachi-agent-helper gate check <run_id> final --json
kkachi-agent-helper event append phase.completed --run <run_id> --payload '<json-object>' --json
kkachi-agent-helper run close <run_id> --json
```

Each phase should follow the same operating shape:

```text
1. KAH creates or activates the run and initializes canonical artifacts.
2. KHS guides Hermes to gather inputs and render artifacts.
3. KHS guides Hermes to compose the backend prompt when a KAB lane is needed.
4. Hermes sends the prompt through KAB.
5. Hermes observes KAB through either CLI wait/read/status or retained `/api/stream`/`/api/events`.
6. Hermes records compact KAH events for phase/artifact milestones.
7. KAH `gate check` records deterministic pass/fail/blocked gate status.
8. KAH final `gate check <run_id> final` and `run close` record final completion, or `run abort` records abandoned work.
```

KAB `send` success is dispatch evidence only. Completion requires public session evidence: either `wait` plus `read/status`, or retained stream/event observations plus final `read/status`. `bridge-events.md` should record the observation mode as `cli_loop`, `retained_stream`, or `hybrid`.

## 2. Why Kkachi exists

Existing AI coding harnesses are often centered on one tool or one provider. Kkachi is designed for a different goal:

```text
Use the best available CLI coding tools, but manage them with a repeatable team workflow.
```

Kkachi aims to provide:

- tool independence across multiple external AI CLIs
- clear delegation to real Hermes team-member profiles
- sequential project execution with one active write lane
- durable run artifacts instead of only chat memory or git commits
- red-team and review gates
- mandatory documentation updates
- self-improvement of prompts, skills, overlays, and checklists from real run evidence

## 3. Glossary

### Hermes Agent

Hermes Agent is the AI assistant runtime used by the team. It can load reusable skills, run terminal tools, work with files, and operate named team-member profiles.

### Hermes skill

A skill is reusable procedural knowledge for a class of work. In Kkachi, skills define how a team member performs phases such as intake, planning, implementation, review, verification, documentation update, and improvement capture.

### General

A general is a named Hermes team-member profile. A general has a role, model/provider configuration, workspace, memory, and operating style.

### Gongmyeong

Gongmyeong is the orchestrator. Gongmyeong receives the master's request, identifies the target project and commander general, delegates work, monitors gates, and reports to the master. In KHS runs, the planner backend authors the substantive plan surface when KAB is used; Gongmyeong or the assigned commander captures it into `plan.md`, normalizes `checklist.md`, and keeps `phase-plan.yaml`/`checklist.md` current as workflow evidence.

### Master

The master is the human owner and final decision maker.

### Commander general

The commander general is the real team-member profile assigned to own a specific development run. The commander reads roadmap/SOT/docs/code, ensures `phase-plan.yaml` is the workflow SOT, asks the KAB planner backend for the plan surface when KAB is used, captures `plan.md`, normalizes `checklist.md`, chooses the external CLI lane from project needs and the backend capability registry, drives implementation through the bridge, and submits evidence. The commander must preserve the selected backend identity, capability snapshot, caveats, unsupported cases, and selection reason in run artifacts.

### Red-team general

A red-team general independently challenges the plan, diff, verification evidence, release risk, and blind spots.

### SOT document

SOT means source of truth. It is a durable project document that records the canonical decision, requirement, design, or feature definition. Kkachi does not treat chat memory or git commit messages as sufficient long-term records.

### Roadmap

A roadmap is a durable project document containing planned tasks such as `TASK-208`. A roadmap task may link to one or more SOT documents.

### Project overlay

A project overlay is project-specific configuration and knowledge. It defines repository path, stack, docs locations, roadmap path, linked SOT conventions, architecture rules, test commands, forbidden changes, done definition, commander general, red-team partner, project-specific gates, backend policy, and required bridge documentation checks.

### Phase skill

A phase skill is a reusable skill for one stage of the workflow, such as intake, plan, implement, review, verify, docs-update, or improve. Phase skills must be backend-capability-aware; they must not assume that all bridge backends share the same approval, question, retained-event, recovery, or observability behavior.

### Progressive Disclosure

Progressive Disclosure means keeping each `SKILL.md` lean and moving conditional details into `references/`, repeated deterministic checks into `scripts/`, and reusable output assets into `assets/`. Kkachi phase skills should load detailed references only when the current phase, project overlay, backend lane, or QA scope requires them.

### Skill QA

Skill QA is the evaluation process for Kkachi phase skills, orchestration skills, project overlays, prompt templates, and review checklists. It uses trigger evals, with-skill vs baseline comparisons, assertion-based grading, eval workspaces, and real run evidence.

### Trigger eval

A trigger eval checks whether a skill description activates for the right user requests and avoids activating for near-miss requests. Each phase skill should maintain both `should_trigger` and `should_not_trigger` cases.

### Near-miss should-not-trigger

Near-miss should-not-trigger cases are realistic prompts that look adjacent to a phase skill but should route to another skill or no Kkachi skill. They prevent accidental over-triggering and phase boundary overlap.

### With-skill vs baseline evaluation

With-skill vs baseline evaluation compares the output of a Kkachi skill against either no skill or the previous version of the same skill. It is required for material changes to shared phase skills or orchestration skills.

### Eval workspace

An Eval workspace preserves one evaluation iteration without overwriting prior outputs. It stores prompts, with-skill output, baseline output, timing, grading, comparison notes, and metadata for later review.

### Delta-capture improvement note

A Delta-capture improvement note records what changed between expected and actual run behavior, why it generalizes beyond one prompt, which surfaces are affected, and what evidence supports a candidate skill or template change.

### Skill release gate

A Skill release gate is the approval boundary before a shared Kkachi skill, orchestration skill, project overlay template, prompt template, review checklist, or shared documentation-update checklist is broadly used. It combines trigger evals, with-skill vs baseline comparison, assertion-based grading, qualitative review when needed, and master approval before patching shared behavior surfaces.

### Assertion-based grading

Assertion-based grading checks objective pass/fail conditions, such as required artifact presence, required JSON fields, phase ordering, backend policy compliance, documentation-update decisions, and final gate completeness.

### Integration coherence verification

Integration coherence verification checks whether connected surfaces agree with each other. In Kkachi this includes SOT basis ↔ acceptance criteria ↔ plan ↔ diff, compatibility matrix ↔ capability registry ↔ selected-cli.json, bridge events ↔ bridge session snapshot, and API ↔ client/hook/type boundaries in application projects.

### Orchestration skill

The orchestration skill coordinates project overlays and phase skills. It determines which phase should run next and enforces stop conditions and artifact requirements.

### kkachi-agent-bridge

`kkachi-agent-bridge` is the runtime integration layer that exposes supported external AI coding TUIs as programmable backend sessions. Current backend states are tracked in `kkachi-agent-bridge/docs/public/compatibility-matrix.md`. This concept document intentionally describes only the operating model and should not duplicate detailed backend implementation status.

### kkachi-agent-helper

`kkachi-agent-helper` is a deterministic CLI. It manages run IDs, `.kkachi/status.json`, `.kkachi/events.jsonl`, run artifact directories, locks, schema checks, and project initialization. It must not become the intelligence layer.

Kkachi-agent-helper does not install Hermes skills. Hermes skill installation is handled by the Hermes native skill system (`hermes skills install` or manual placement under `~/.hermes/skills/`). Kkachi-agent-helper provides `project init` for creating and resetting project-local bootstrap files (such as `.kkachi/project-overlay.yaml` and `docs/kkachi-docs-map.yaml`). Kkachi-hermes-skills (KHS) never writes directly to project directories; KHS prepares template parameters and invokes kkachi-agent-helper `project init` to perform the actual file operations. Reconfiguration uses `--force` and preserves existing runs, artifacts, and gate history while rewriting config, overlay, docs-map, and schema copies; a `project.reconfigured` event is recorded.

### Backend capability registry

A machine-readable registry, such as `registries/cli-capabilities.yaml`, that records each bridge backend's supported capabilities, caveats, unsupported cases, verified versions, and recommended usage constraints. It should be derived from or kept consistent with `kkachi-agent-bridge/docs/public/compatibility-matrix.md`, the human-readable canonical tested-support ledger.

### Bridge session identity

The backend identity attached to each bridge session. At minimum this includes `backend_type` and `adapter_type`; run evidence may also include bridge version, backend CLI version, config identity, runtime mode, capability snapshot, caveats, and open pendings.

### Selected CLI artifact

`selected-cli.json` is the run-local evidence artifact that records which backend lane the commander selected, why it was selected, what capability snapshot was used, which caveats were accepted, and which unsupported cases were ruled out.

### Work path

Kkachi uses exactly two work paths:

- `A_development_execution`: development execution when a sufficient SOT basis already exists.
- `B_discovery_shaping`: discovery, shaping, SOT update, and roadmap trace work when the SOT basis is missing or insufficient.

### Work mode

`standard` is the default operating mode. `light` is a constrained reduction profile inside Path A or Path B, not a separate path.

### Urgency

Urgency is metadata only: `normal`, `urgent`, or `critical`. Urgency does not create a separate work path and does not bypass SOT-first gates.

### SOT policy

`sot_policy` records how the run satisfies SOT-first operation: `existing_sot_basis`, `minimal_sot_before_code`, or `full_sot_before_code`. Path A normally uses `existing_sot_basis`; Path B normally creates `minimal_sot_before_code` or `full_sot_before_code` before implementation.

### Execution mode

`execution_mode` records the concrete lane of work, such as `production_write`, `adapter_qa`, `readiness_hardening`, `research`, `verification`, or `docs_only`.


## 4. Architecture overview

Kkachi has five main layers:

```text
Human master
  -> Gongmyeong orchestrator
    -> real commander general + Hermes phase skills
      -> project overlay + docs map
      -> CLI capability registry
      -> selected-cli.json + capability-check.md
      -> task-contract.yaml + prompt.md + context-pack.md
      -> bridge-session-snapshot.json + bridge-events.md
      -> kkachi-agent-bridge
        -> backend session identity:
             backend_type + adapter_type + capability/caveat snapshot
        -> external AI coding CLI
```

Kkachi does not select a backend by vague preference. The commander must select a lane using project needs, the backend capability registry, the bridge compatibility matrix, and the project overlay's backend policy. The selected backend identity and caveats must be preserved in run artifacts so later review, red-team, verification, docs-update, and improvement steps can reconstruct why that lane was used.

Supporting layers:

```text
kkachi-agent-helper
  -> deterministic state, artifact, schema, lock, diagnostics, and project bootstrap management

red-team general
  -> independent review and release-risk challenge

project docs
  -> roadmap, SOT, architecture, ADR, todo/task, spec, and other durable project records
```

The external CLI produces code. The commander general owns strategy, prompting, review, documentation, and final evidence. The helper records state. The bridge connects tools. The red team challenges quality.

## 5. Repository distribution model

Kkachi should be distributed as three independently versioned GitHub repositories:

```text
kkachi-agent-bridge
  Runtime integration layer between Hermes generals and external AI CLI tools.

kkachi-agent-helper
  Deterministic CLI for status, events, run artifacts, locks, schema checks, diagnostics, and project initialization.

kkachi-hermes-skills
  Hermes skill templates, phase skills, orchestration skill, project overlay templates, prompt templates, schemas, registries, and examples.
```

The separation matters because each repository changes for different reasons:

- bridge changes when external CLI protocols change
- helper changes when state and artifact management changes
- skills change when workflow, prompts, reviews, or project conventions improve

## 6. Recommended kkachi-hermes-skills layout

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
        path-a-implement.md.tmpl
        path-b-shaping.md.tmpl
        review.md.tmpl
      codex/
        path-a-implement.md.tmpl
        path-b-shaping.md.tmpl
        review.md.tmpl
      gemini/
      glm/
      opencode/

  evals/
    trigger/
      kkachi-orchestrate.trigger-evals.yaml
      kkachi-task-contract.trigger-evals.yaml
      kkachi-backend-select.trigger-evals.yaml
      kkachi-prompt-compose.trigger-evals.yaml
      kkachi-phase-state.trigger-evals.yaml
      kkachi-plan.trigger-evals.yaml
      kkachi-ask.trigger-evals.yaml
      kkachi-implement.trigger-evals.yaml
      kkachi-enhance-test.trigger-evals.yaml
      kkachi-optimize.trigger-evals.yaml
      kkachi-review.trigger-evals.yaml
      kkachi-verify.trigger-evals.yaml
      kkachi-docs-update.trigger-evals.yaml
      kkachi-request-feedback.trigger-evals.yaml
      kkachi-handle-feedback.trigger-evals.yaml
      kkachi-final-verify.trigger-evals.yaml
      kkachi-improve.trigger-evals.yaml
    execution/
      kkachi-task-contract.execution-evals.yaml
      kkachi-backend-select.execution-evals.yaml
      kkachi-prompt-compose.execution-evals.yaml
      kkachi-plan.execution-evals.yaml
      kkachi-verify.execution-evals.yaml
      kkachi-docs-update.execution-evals.yaml
      kkachi-final-verify.execution-evals.yaml
      kkachi-improve.execution-evals.yaml

  benchmarks/
    kkachi-plan/
      iteration-001/
        eval-<descriptive-name>/
          eval_metadata.json
          with_skill/
            outputs/
            timing.json
            grading.json
          without_skill/
            outputs/
            timing.json
            grading.json
          comparison.md
        benchmark.json

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

Kkachi skills should follow Progressive Disclosure. `SKILL.md` contains the core phase behavior and routing rules. `references/` contains conditional details such as backend-specific caveats, project-stack-specific QA rules, or long checklists. `scripts/` contains deterministic or repeated checks that agents would otherwise recreate during evaluation or verification.

Evaluation artifacts live in the skills repository, not in ordinary project run artifacts, unless a specific development run modifies shared skills or templates. Ordinary project runs continue to use `.kkachi/runs/<run_id>/`.

`skill-pack.yaml` declares the skill pack version, included skills, and compatible helper/bridge versions.

Example:

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
  - kkachi-final-verify
  - kkachi-improve
```

## 7. Project docs map

Each project should expose a docs map so the commander does not guess where canonical documents live.

Example:

```yaml
project: kkachi-agent-bridge
repo_path: /Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-bridge

canonical_docs:
  readme: README.md
  compatibility_matrix: docs/public/compatibility-matrix.md
  requirements: docs/dev/spec/requirements.md

backend_support_docs:
  roadmap: docs/dev/roadmap.md
  adr: docs/dev/adr.md
  e2e_version_baselines: docs/dev/e2e/version-baselines.md
  failure_triage: docs/dev/e2e/failure-triage.md

test_commands:
  unit:
    - go test ./...
  integration:
    - go test -tags=integration ./test/integration

forbidden_changes:
  - mutate user global Gemini config
  - use terminal scraping as shipped control plane
  - mark Gemini answer supported without ask_user hook-injection evidence
  - mark retained events supported without reducer-owned cursor/backfill tests
```

The docs map is used by project overlays and by the commander when preparing `plan.md`, checklists, documentation updates, and completion evidence.

## 8. Work Path A: Development Execution

Path A is used when the work already has a sufficient SOT basis.

A SOT basis may be:

- an existing SOT document;
- a roadmap task linked to SOT documents;
- a specification;
- an ADR;
- an architecture document;
- an existing test contract;
- a canonical project rule;
- a clearly documented bug expectation.

Path A is not limited to large feature work. It also covers small bug fixes, small refactors, and small feature implementation when the expected behavior is already anchored to durable project evidence.

Default mode: Standard Mode.
Optional constrained reduction: Light Mode.

The master may say:

```text
Work on TASK-208 from the roadmap file.
```

### Standard Mode flow

Standard Mode is the canonical Kkachi operating mode. The following flow defines the default behavior for Path A.

Expected flow:

1. Gongmyeong identifies the target project, roadmap file, and real commander general.
2. Gongmyeong delegates the task reference to the commander.
3. The commander reads the roadmap entry.
4. The commander reads SOT documents linked from the task.
5. The commander reads architecture, ADR, spec, todo/task docs, coding rules, related code, and tests.
6. The commander records the SOT basis in `sot-basis.md`.
7. The commander confirms roadmap trace, or records why the existing roadmap task already provides traceability.
8. The commander derives acceptance criteria from the SOT basis before implementation.
9. The commander creates `plan.md`.
10. The commander creates the execution checklist.
11. The commander chooses an external CLI lane using the CLI capability registry, the bridge compatibility matrix, and the project overlay's backend policy.
12. The commander records the selected lane in `selected-cli.json`, including `backend_type`, `adapter_type`, capability snapshot, caveats, unsupported cases, and selection reason.
13. The commander records `capability-check.md` before implementation starts when a bridge backend lane is used.
14. The commander generates a precise English prompt and sends it through `kkachi-agent-bridge`.
15. The external CLI edits the repository.
16. The commander records bridge session identity and relevant bridge events.
17. The commander reviews the diff and evidence.
18. Red-team review challenges risk and blind spots.
19. The commander runs tests/e2e verification.
20. The commander updates required docs or records why no docs update is needed.
21. The commander records improvement candidates from real evidence.
22. Gongmyeong reports the result to the master in Korean.

If the task requires capabilities that the selected backend does not support, the commander must either choose another backend or record why the task is safe despite that unsupported capability. Backends or features marked planned, unsupported, degraded, or API/SSE-only must not be used for incompatible production work unless the run is explicitly scoped to adapter QA, readiness, research, verification, or docs.

For `research`, `verification`, or `docs_only` execution modes, roadmap trace may be marked not applicable only with an explicit reason recorded in the run artifacts.

Important boundary:

```text
The KAB planner backend authors the plan surface when a KAB planner lane is used. KHS/Hermes captures that surface into plan.md, normalizes the KHS Checklist Seed into checklist.md, and keeps phase-plan.yaml/checklist.md current as workflow evidence.
```

### Path A Light Mode reductions

Light Mode is not a separate path and not a bypass. It is a constrained reduction profile for small, low-risk Path A work.

Path A Light Mode may be used only when:

- the work already has a SOT basis;
- the scope is small and bounded;
- no architecture boundary changes are required;
- no data model, schema, permission, billing, migration, or API contract decision remains open;
- acceptance criteria are clear;
- targeted verification is sufficient.

Allowed reductions:

- `plan.md` may be concise.
- red-team plan review may be merged into final red-team review.
- `context-pack.md` may be minimal.
- feedback round 1 is mandatory for KHS; rounds 2-3 may be skipped with an explicit reason.
- test scope may be targeted rather than broad.
- UI QA may be marked not applicable when the change has no UI impact.

Not skippable:

- SOT basis;
- roadmap trace or explicit existing trace;
- acceptance criteria;
- selected backend evidence, if a bridge backend lane is used;
- capability check, if a bridge backend lane is used;
- diff capture;
- verification evidence;
- docs-update decision;
- final report.

## 9. Work Path B: Discovery / Shaping / SOT / Roadmap

Path B is used when the work does not yet have a sufficient SOT basis.

Typical Path B inputs include:

- new feature ideas;
- new improvement requests;
- unclear bug reports;
- behavior changes;
- architecture decisions;
- roadmap item creation;
- requests that start from an idea rather than a defined task.

Path B does not primarily implement code. Its purpose is to turn an unclear or undocumented request into a traceable development item by creating or updating SOT documents, roadmap entries, acceptance criteria, and implementation handoff artifacts.

Default mode: Standard Mode.
Optional constrained reduction: Light Mode.

The master may report a new bug, improvement, or feature idea that does not yet exist in the roadmap.

### Standard Mode flow

Expected flow:

1. Gongmyeong identifies the target project and commander general.
2. The commander records request classification in `intake-classification.md`.
3. The commander reads existing docs, roadmap, related SOT documents, and relevant code.
4. The commander records existing documentation review in `discovery/existing-docs-review.md`.
5. The commander records problem framing in `discovery/problem-framing.md`.
6. The commander classifies the request as bug, improvement, feature, architecture decision, or roadmap item.
7. The commander researches external or technical context if needed.
8. The commander records strategy options and selected strategy.
9. The commander creates or updates the required SOT document before any implementation starts.
10. The commander updates the roadmap so the new work is traceable.
11. The commander defines acceptance criteria and non-scope.
12. The commander records task breakdown and implementation readiness.
13. The commander records `discovery/handoff-to-development.md` if implementation will follow.
14. If implementation is approved or included, the run must pass the SOT basis and roadmap trace gates before entering Path A.
15. Gongmyeong reports the shaping result to the master in Korean.

Path B may hand off to Path A either in the same run or in a follow-up run, but implementation must not start until the Path B artifacts have passed the SOT basis and roadmap trace gates.

### Path B Light Mode reductions

Path B Light Mode is used when the request lacks a SOT basis but the required shaping is small and low-risk.

B-Light still happens before implementation.

Allowed reductions:

- `discovery/existing-docs-review.md` may be concise.
- external research may be omitted when no external or technical research is needed.
- `discovery/strategy-options.md` may be reduced to one selected strategy plus obvious rejected alternatives.
- the SOT may be a minimal SOT stub.
- roadmap update may create a single task entry.
- task breakdown may contain a single implementation task.
- red-team shaping review may be merged into final gate review.

Not skippable:

- request classification;
- minimal SOT creation or SOT update;
- roadmap trace;
- acceptance criteria;
- non-scope;
- handoff to Path A when implementation will follow;
- final report.

## 10. No bypass path for urgent work

Kkachi does not provide a Path C or hotfix bypass path.

Urgency may reduce artifact depth, review depth, and research depth, but it must not bypass the SOT gate.

For urgent work:

```text
existing SOT basis exists
  -> Path A Light or Path A Standard

existing SOT basis does not exist
  -> Path B Light
  -> minimal SOT + roadmap trace + acceptance criteria
  -> Path A Light or Path A Standard
```

Implementation must not start from chat-only instruction, agent memory, or code diff alone.

## 11. Sequential development model

Kkachi defaults to sequential development.

Default rule:

```text
One active write lane per project.
```

Reasons:

- commander memory accumulates in the same development story
- roadmap, SOT, plan, checklist, diff, review, verification, docs, and improvement notes form one timeline
- `.kkachi/status.json` can represent one clear active state
- root-cause analysis is easier
- red-team gates are clearer
- project docs are less likely to diverge

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
- concurrent bridge sessions that mutate the same repository

Git worktree mode may exist later as an explicit advanced mode, but it requires separate run IDs, separate artifacts, one merge owner, clear isolation, and a final sequential integration gate.

## 12. Run artifact model

Recommended project-local state layout:

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

Standard Mode defines the canonical artifact set. Light Mode may use the same artifact names with shorter content or explicit not-applicable records. Kkachi should avoid creating a separate incompatible Light artifact schema unless the helper schema explicitly supports it.

`skill-qa/` is created only when the run proposes, tests, or applies changes to shared Kkachi skills, project overlay templates, prompt templates, or review checklists. Ordinary development runs do not need this directory unless they produce a skill/template improvement candidate that is evaluated during the same run.

`status.json` is the current state snapshot.

`events.jsonl` is append-only history.

`runs/<run_id>/` preserves reproducible evidence for the current work.

`run-metadata.json` records `work_path`, `work_mode`, `urgency`, `sot_policy`, `execution_mode`, assigned generals, and gate state.

`intake-classification.md` records Path A / Path B classification and why the selected mode is eligible.

`sot-basis.md` records the durable SOT basis that permits implementation, or the Path B-created minimal/full SOT basis.

`acceptance-criteria.md` records criteria derived from the SOT basis.

`selected-cli.json` records the commander's backend lane decision.

`capability-check.md` records which task requirements were matched against which backend capabilities before implementation started.

`bridge-session-snapshot.json` records the bridge session identity and current public session fields, including `session_id`, `backend_type`, `adapter_type`, state, lifecycle class, open pendings, bridge version/config identity when available, and backend-specific hook data when available.

`bridge-events.md` summarizes relevant bridge events, pending flows, approvals, rejections, answers, stream observations, and recovery behavior.

`impl-log.md` records implementation notes and scope-sensitive rationale.

`test-log.md` preserves raw command output when tests or verification commands are run.

`verification.md` records summarized verification results, exact commands, failure classification, and verdict.

## 13. Documentation update policy

A Kkachi run is not complete merely because code changed or a commit exists.

Completion requires checking whether durable docs need updates:

- roadmap
- linked SOT documents
- architecture docs
- ADRs
- specs
- todo/task docs
- operation docs
- test/e2e docs

For `kkachi-agent-bridge` runs, documentation update checks must also include:

- `README.md` summary support table
- `docs/public/compatibility-matrix.md` tested-support ledger
- `docs/dev/spec/requirements.md` when bridge contract changes
- backend-specific SOT or todo documents
- project overlay and CLI capability registry when supported/caveat states change

If code behavior changes, relevant docs must be updated or the commander must explicitly record why no docs update was needed.

`docs-update.md` should record:

- docs inspected
- docs changed
- docs intentionally not changed and why
- roadmap/SOT traceability
- remaining documentation follow-up if any

Documentation update checks must include integration coherence verification when the change crosses document, API, registry, or runtime boundaries.

Examples:

- `kkachi-agent-bridge/docs/public/compatibility-matrix.md` ↔ `kkachi-hermes-skills/registries/cli-capabilities.yaml` ↔ project overlay backend policy;
- SOT document ↔ roadmap task ↔ acceptance criteria ↔ plan.md;
- selected-cli.json ↔ capability-check.md ↔ bridge-session-snapshot.json;
- API response shape ↔ frontend hook type ↔ UI state handling;
- DB schema field ↔ API response field ↔ client type definition.

## 14. Self-improvement loop

Kkachi should improve from real evidence, not speculation.

Each run can produce improvement candidates for:

- prompt templates
- CLI selection rules
- phase skills
- project overlays
- docs map
- review checklists
- e2e procedures
- red-team gates

A candidate should be recorded in `improvement-note.md`. Shared skill, orchestration skill, project overlay template, prompt template, review checklist, or shared documentation-update checklist changes should be reviewed and approved before being patched into shared repositories.

Improvement candidates should use this structure when possible:

```yaml
improvement_candidate:
  source_run_id: "<run-id>"
  source_project: "<project>"
  observed_failure: "<what failed or degraded>"
  generalized_lesson: "<principle-level lesson, not a one-off patch>"
  affected_surface:
    - "skills/kkachi-plan/SKILL.md"
    - "templates/run-artifacts/capability-check.md.tmpl"
    - "registries/cli-capabilities.yaml"
  proposed_patch: "<concrete change>"
  evidence:
    - ".kkachi/runs/<run-id>/plan.md"
    - ".kkachi/runs/<run-id>/selected-cli.json"
    - ".kkachi/runs/<run-id>/redteam/final-gate-review.md"
  eval_required:
    trigger_eval: true
    with_skill_baseline: true
    assertion_grading: true
  approval_required: true
  status: "candidate|approved|rejected|patched"
```

Do not update shared skills from speculation alone. Repeated run evidence, a clear failure mode, or a high-risk single incident may justify a candidate.

## 15. Skill QA and Evaluation Model

Kkachi skills must improve through evidence, not speculation. Skill QA applies to phase skills, orchestration skills, project overlay templates, prompt templates, review checklists, and documentation-update checklists.

### Evaluation types

Kkachi uses four evaluation types:

1. Trigger evals
   - Confirm that each skill activates for the right user requests.
   - Include `should_trigger` and `should_not_trigger` cases.
   - `should_not_trigger` must include near-miss cases, not only obviously unrelated requests.

2. With-skill vs baseline execution evals
   - For a new skill, compare against no skill.
   - For a changed skill, compare against the previous skill version.
   - Preserve both outputs, timing, grading, and comparison notes.

3. Assertion-based grading
   - Use deterministic checks when possible.
   - Check artifact existence, schema fields, phase ordering, backend policy compliance, docs-update decisions, and final gate completeness.
   - Avoid assertions that both with-skill and baseline always pass.

4. Qualitative review
   - Use commander, red-team, and external feedback lanes when output quality is partly judgment-based.
   - Record evidence and reviewer rationale.

### Minimum trigger eval standard

Each phase skill should maintain:

```yaml
should_trigger: 8-10 realistic user prompts
should_not_trigger: 8-10 realistic near-miss prompts
```

Trigger evals must cover:

- formal and casual phrasing;
- explicit and implicit intent;
- project-specific task references;
- ambiguous but common user phrasing;
- near-miss requests that should route to another skill or no Kkachi skill.

### Minimum execution eval standard

For MVP, execution evals are required for:

- `kkachi-plan`
- `kkachi-task-contract`
- `kkachi-backend-select`
- `kkachi-prompt-compose`
- `kkachi-verify`
- `kkachi-docs-update`
- `kkachi-final-verify`
- `kkachi-improve`

Each execution eval should include:

- realistic user prompt;
- expected artifacts;
- assertions;
- with-skill output;
- baseline output;
- timing data when available;
- grading result;
- comparison summary.

### Skill change gate

A shared skill or template must not be patched directly from a single anecdote. A proposed change should be recorded first in `improvement-note.md` and then promoted only after review.

A material skill change requires:

- source run evidence;
- generalized failure description;
- affected surface list;
- proposed patch;
- trigger eval update when description changes;
- with-skill vs baseline comparison when behavior changes;
- assertion grading when the output is objectively checkable;
- master approval before patching shared skills, orchestration skills, project overlay templates, prompt templates, review checklists, or shared documentation-update checklists.

## 16. MVP scope

Recommended MVP:

- pilot project: `kkachi-agent-bridge`
- pilot task: backend implementation, adapter validation, or readiness hardening selected from the current roadmap
- available KAB lanes for Kkachi pilot selection: Claude Code, GLM, Codex CLI, Gemini CLI, OpenCode
- experimental lane: only for backend features explicitly marked experimental or degraded by the compatibility matrix
- helper scope: `status.json`, `events.jsonl`, run directories, locks, selected-cli capture, capability checks, bridge session snapshots, and artifact existence checks
- phases: intake -> plan -> implement -> review -> red-team -> verify -> docs-update -> improve
- project overlay: one minimal overlay with README, compatibility matrix, requirements, backend SOT/todo docs, test commands, forbidden changes, and backend policy
- registry scope: `cli-capabilities.yaml` synchronized with the bridge compatibility matrix
- red-team: one required review gate before completion
- success condition: one real sequential Gemini-related bridge task completed with plan, checklist, selected backend evidence, bridge session evidence, prompt, output, diff, tests, review, red-team, verification, docs update evidence, and improvement note preserved
- skill QA scope:
  - trigger eval files for all MVP phase skills
  - execution evals for `kkachi-plan`, `kkachi-verify`, `kkachi-docs-update`, and `kkachi-improve`
  - assertion schemas for artifact completeness, backend selection evidence, docs-update decision, and final gate
  - one with-skill vs baseline comparison for at least one MVP phase skill before broad use
  - improvement-note schema with delta-capture fields

The MVP no longer needs to prove that Kkachi can use more than one bridge backend. The bridge already provides multi-backend support. The MVP should prove that Kkachi can select the correct backend based on capability, preserve bridge evidence, enforce gates, update docs, and capture improvement notes.

## 17. Open decisions

Before implementing the full system, decide KHS extensions around the current KAH `@latest` base contract:

- KHS expectations for `status.json` fields beyond the KAH schema
- canonical KHS event names to record through `event append`
- exact `cli-capabilities.yaml` schema derived from `docs/public/compatibility-matrix.md`
- KHS extension fields for `selected-cli.json`
- KHS extension fields for `bridge-session-snapshot.json`
- KHS prose standard for `capability-check.md` beyond KAH's backend gate requirements
- project context pack schema
- docs map schema
- backend lane policy for production-supported, degraded, planned, and unsupported states
- policy for when a planned/degraded/backend-specific feature may be used for adapter QA or readiness runs
- backend support-promotion/demotion policy
- policy for synchronizing README summary support view, compatibility matrix, Kkachi capability registry, and project overlay
- exact phase skill list for MVP
- red-team gate minimum for MVP
- KHS version-gating policy for future helper CLI command changes
- skill install/update/version compatibility policy
- orchestration skill responsibilities and stop conditions
- exact `work_path` schema
- exact `work_mode` schema
- exact `urgency` schema
- exact `sot_policy` schema
- exact `execution_mode` schema
- exact Light Mode eligibility rules
- exact Light Mode permitted reductions by path
- exact SOT basis gate standard
- exact roadmap trace gate standard
- whether Light Mode uses identical artifact names with reduced content or separate artifact templates
- exact trigger eval schema
- exact execution eval schema
- exact grading.json schema
- exact benchmark workspace layout
- minimum trigger accuracy threshold before skill release
- minimum assertion pass threshold before skill release
- when with-skill vs baseline evaluation is mandatory
- how to version skill changes independently from helper and bridge changes
- how to promote improvement candidates from run-local `improvement-note.md` to shared skill/template patches
- which skill QA checks are deterministic helper responsibilities
- which skill QA checks require commander, red-team, or external feedback review


## 18. Bridge support SOT synchronization

Kkachi must not duplicate detailed backend support truth in every concept document. Backend tested support belongs in `kkachi-agent-bridge/docs/public/compatibility-matrix.md`.

Document roles:

| Document | Role |
|---|---|
| `kkachi-agent-bridge/README.md` | User-facing summary support view and usage overview. |
| `kkachi-agent-bridge/docs/public/compatibility-matrix.md` | Canonical tested-support ledger, tested versions, capability details, and caveats. |
| `kkachi-hermes-skills/registries/cli-capabilities.yaml` | Machine-readable backend capability registry for orchestration and commander use. |
| `kkachi-hermes-skills/docs/sot/concept.md` | Kkachi product definition and operating philosophy. |
| `kkachi-hermes-skills/docs/sot/architecture-understanding.md` | Phase skill, project overlay, and orchestration structure agreement. |
| `kkachi-hermes-skills/docs/sot/skill-template.md` | Reusable KHS skill authoring and execution template guidance. |
| `kkachi-hermes-skills/docs/sot/phase-orchestration-policy.md` | User-confirmed KHS phase orchestration policy. |
| `kkachi-hermes-skills/docs/discussions/` | Temporary discussion notes; not policy until promoted into SOT, registries, templates, or skills. |

Synchronization rules:

1. `docs/public/compatibility-matrix.md` is the canonical source for backend tested support.
2. `README.md` must remain a summary view of `docs/public/compatibility-matrix.md`.
3. `cli-capabilities.yaml` must be derived from or manually synchronized with `docs/public/compatibility-matrix.md`.
4. `docs/sot/concept.md` and `docs/sot/architecture-understanding.md` should describe capability-aware operation and link to the support ledger rather than copying every backend detail.
5. Discussion notes under `docs/discussions/` should be promoted into SOT/registry/template/skill surfaces before they are treated as final policy.
6. When a backend status changes, check README, compatibility matrix, capability registry, and the relevant project overlay together.

## 19. Gemini lane policy

Gemini is currently supported in the KAB public compatibility matrix. KHS must preserve Gemini-specific caveats while treating Gemini as an eligible backend when project policy permits it.

Rules:

- Gemini production write delivery is allowed when project policy permits it and required capabilities match.
- Gemini answer/question flow uses the hook-native `BeforeTool ask_user` path.
- Gemini retained events are supported through KAB public `/api/events` and `/api/stream`.
- First-class Gemini clients should open `/api/stream` early and use `/api/events` plus `read`/`status`/`wait`/`read?tui=true` for reconciliation and failure evidence.
- Terminal capture is diagnostic only and must not become the shipped control plane.
- Direct TUI shell bypass must remain disabled; readiness fails closed if it cannot be disabled.
- Gemini plan approval requires explicit post-approval start before implementation.
