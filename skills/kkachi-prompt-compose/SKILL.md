---
name: kkachi-prompt-compose
description: Render the backend-specific prompt that Hermes will send through KAB by combining the AI-neutral task contract, phase contract, selected backend, capability check, project overlay, context pack, and backend prompt profile.
version: 0.1.0
---

# Kkachi Prompt Compose

Use this skill after backend selection passes and before Hermes sends a prompt through kkachi-agent-bridge.

Trigger boundary: use this phase skill only after `kkachi-orchestrate` or an explicit master request has selected KHS/Kkachi for the work. Do not trigger it for ordinary direct Hermes edits, quick one-file fixes, typo/config patches, or read-only explanations unless the master explicitly asks for KHS/Kkachi or delegates the work to a KHS-using commander such as 조운 or 마초.

## Core rule

KHS renders a backend-specific prompt, and KAH stores that rendered prompt in the canonical run artifact `prompt.md`. The prompt may adapt instruction style to the selected backend, but it must preserve the same desired state, acceptance criteria, constraints, non-goals, and verification contract from `task-contract.yaml`.

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
5. For active KAS/KAH roadmap-task work, include the staged actor boundary: Stage 1 direct Codex app-server, Stage 2 KAB Codex-first, or Stage 3 selected KAB backend owns the plan draft/revision and substantive code/test/build/task-doc mutations; Blue/Red vet the returned plan and approve or request changes; Blue/Red/Orange/Gray review and verify rather than directly editing, unless 주군 explicitly requested a direct role edit or the work is outside the roadmap/KAS+KAH path.
6. For planner prompts, explicitly say plan-only: the Stage 1 direct Codex app-server planner, Stage 2 KAB Codex planner, or Stage 3 selected KAB planner drafts the plan and waits; Blue and Red vet it; any requested plan change must come back as a revised plan from the same planner lane before implementation starts.
7. For implementation or feedback-handling prompts, explicitly say color-review and GLM-review fixes must be applied by the Stage 1 direct Codex app-server lane, Stage 2 KAB Codex implementer, or Stage 3 selected KAB implementer, not by Blue/Red/Orange/Gray direct patches.
8. Render all shell-command instructions with the real user home, using `HOME=<real-user-home> <command>` in reusable templates. Include Git commands in this rule so commit-time global config comes from the user's home.
9. For planner prompts, require a clearly delimited `KHS Checklist Seed` section with unchecked markdown tasks containing phase, evidence artifact, and verification/check condition. This seed is parsed by KHS, not treated as final `checklist.md`.
10. Render `templates/run-artifacts/prompt.md.tmpl` into `.kkachi/runs/<run_id>/prompt.md`.
11. Record a compact event with `kkachi-agent-helper event append artifact.updated --run <run_id> --payload '{"path":"prompt.md","phase":"prompt-compose"}' --json` when useful.

## KAB caveats to preserve

- Codex prompts must assume wrapper/API-driven control and stable Kkachi retained events, not raw Codex app-server messages.
- Gemini prompts must not imply that plan approval starts implementation; explicit post-approval start is required.
- OpenCode prompts must respect API/SSE authority and must not treat rendered question-like text as a bridge `needs_input` pending.
- GLM review/verification prompts must check `response_fidelity_warning` after a rejected permission.
- Official GLM Octo prompts must be rendered for KAB GLM delivery only, not direct `glm` CLI review. The rendered prompt's first command text must be `/octo:review`, and the prompt/evidence contract must require KAB session id, backend type `glm`, real-user-HOME KAB/GLM path preflight, `prompt_confirmed: true`, readback/events, bounded watcher, and feedback artifact evidence.

## Outputs

- `prompt.md`
- prompt integrity result
- optional KAH `event append artifact.updated` evidence

## Gate

FAIL if the rendered prompt weakens or changes the task contract. BLOCKED if no compatible prompt profile or template exists for the selected backend/path/phase.
