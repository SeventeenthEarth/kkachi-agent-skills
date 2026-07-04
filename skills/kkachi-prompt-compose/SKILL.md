---
name: kkachi-prompt-compose
description: Render the backend-specific prompt that Hermes will send through KAB by combining the AI-neutral task contract, phase contract, selected backend, capability check, project overlay, context pack, and backend prompt profile.
version: 0.2.0
---

# Kkachi Prompt Compose

Use this skill after backend selection passes and before Hermes sends a prompt through kkachi-agent-bridge.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

KHS renders a backend-specific prompt, and KAH stores that rendered prompt in the canonical run artifact `prompt.md`. The prompt may adapt instruction style to the selected backend, but it must preserve the same desired state, acceptance criteria, constraints, non-goals, and verification contract from `task-contract.yaml`.

Rendered prompts must preserve the TOKEN-002 output policy for both v0.2 GJC candidate artifacts and any explicitly selected KAB lane: backend product output is English; console summaries use `Status`, `Summary`, `Files`, `Verification`, `Risks/blockers`, `Detailed artifact`, and `Next action requested`; detailed backend reasoning goes to `.kkachi/runs/<run_id>/artifacts/<phase>/backend-<phase>.md` or the requested phase artifact. If the backend cannot write that artifact, the prompt must require a compact `Status: blocked` artifact-write blocker instead of a full chat dump.

## Inputs

- `task-contract.yaml`
- `selected-cli.json`
- `capability-check.md`
- current phase contract from `registries/phase-contracts.yaml`
- backend profile from `registries/backend-prompt-profiles.yaml`
- backend capability entry from `registries/cli-capabilities.yaml`
- project overlay and context pack
- phase-specific prompt template from `templates/prompts/<backend>/`

## Flow

1. Choose the prompt template for backend, path, and phase.
2. Render the prompt body in English.
3. Run an integrity check against `task-contract.yaml`.
4. Include KAB caveats, capability limits, plan lifecycle differences, permission/input preconditions, supported observation modes, and required output artifacts.
5. For active KAS/KAH roadmap-task work, include the v0.2 actor boundary: KAS owns contracts/prompt policy; KAH owns deterministic run/gate/evidence state; approved GJC `ralplan` or `ultragoal` artifacts are candidates only; KAT evidence is factual/mechanical only; and KAB runtime/session control is explicit-only. Required plan reviewers vet candidate plans and approve or request changes, including Blue synthesis plus Red and Orange plan-vet acceptance where policy requires it, with Red/Orange reviewers resolved from the project/team role registry when applicable. Blue/Red/Orange/Gray review and verify rather than directly editing, unless 주군 explicitly requested a direct role edit or the work is outside the roadmap/KAS+KAH path.
6. For planner prompts, explicitly say plan-only: the selected v0.2 candidate lane drafts the plan and waits. Required Blue+Red+Orange(+Gray when documentation/integrity applies) plan reviewers vet it for active KAS/KAH roadmap policy work; any requested plan change must come back as a revised plan before implementation starts.
7. For implementation or feedback-handling prompts, explicitly say color-review, project-Gray, and MAR fixes must be applied by the selected candidate/implementation lane, not by Blue/Red/Orange/Gray direct patches unless an exception is recorded.
8. Render all shell-command instructions with the real user home, using `HOME=<real-user-home> <command>` in reusable templates. Include Git commands in this rule so commit-time global config comes from the user's home.
9. For planner prompts, require a clearly delimited `KHS Checklist Seed` section with unchecked markdown tasks containing phase, evidence artifact, and verification/check condition. This seed is parsed by KHS, not treated as final `checklist.md`.
10. Render `templates/run-artifacts/prompt.md.tmpl` into `.kkachi/runs/<run_id>/prompt.md`.
11. Record a compact event with `kkachi-agent-helper event append artifact.updated --run <run_id> --payload '{"path":"prompt.md","phase":"prompt-compose"}' --json` when useful.

## KAB caveats to preserve

- Codex prompts apply only to explicitly selected KAB Codex work and must assume wrapper/API-driven control plus stable Kkachi retained events, not raw Codex app-server messages.
- Gemini prompts must not imply that plan approval starts implementation; explicit post-approval start is required.
- OpenCode prompts must respect API/SSE authority and must not treat rendered question-like text as a bridge `needs_input` pending.
- GLM review/verification prompts must check `response_fidelity_warning` after a rejected permission.
- MAR prompts must be rendered from the role-first reviewer matrix and selected provider lane. Required roles are `logic`, `security`, `arch`, `cve`, and `test_adequacy`. The rendered request must include the task contract, acceptance criteria, diff/artifact bundle, required role id, provider lane id, read-only review boundary, raw-output cap, and explicit instruction that provider dispatch success is not review completion. The prompt/evidence contract must require provider preflight/toolchain proof, bounded raw-output path, parse status, role coverage status, and Blue disposition evidence.

## Outputs

- `prompt.md`
- prompt integrity result
- optional KAH `event append artifact.updated` evidence

## Gate

FAIL if the rendered prompt weakens or changes the task contract. BLOCKED if no compatible prompt profile or template exists for the selected backend/path/phase.


## V01CLEAN active-baseline note

Any legacy Stage 1/Stage 2/Stage 3, direct Codex app-server, or KAB `native_codex` wording retained in this file is historical context only unless a later approved task explicitly selects KAB with current capability evidence. The active KAS/KAH v0.2 path is KAS policy + KAH deterministic evidence + approved GJC candidate artifacts, with KAT factual evidence only.
