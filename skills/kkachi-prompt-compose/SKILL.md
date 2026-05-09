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
5. For planner prompts, require a clearly delimited `KHS Checklist Seed` section with unchecked markdown tasks containing phase, evidence artifact, and verification/check condition. This seed is parsed by KHS, not treated as final `checklist.md`.
6. Render `templates/run-artifacts/prompt.md.tmpl` into `.kkachi/runs/<run_id>/prompt.md`.
7. Record a compact event with `kkachi-agent-helper event append artifact.updated --run <run_id> --payload '{"path":"prompt.md","phase":"prompt-compose"}' --json` when useful.

## KAB caveats to preserve

- Codex prompts must assume wrapper/API-driven control and stable Kkachi retained events, not raw Codex app-server messages.
- Gemini prompts must not imply that plan approval starts implementation; explicit post-approval start is required.
- OpenCode prompts must respect API/SSE authority and must not treat rendered question-like text as a bridge `needs_input` pending.
- GLM review/verification prompts must check `response_fidelity_warning` after a rejected permission.

## Outputs

- `prompt.md`
- prompt integrity result
- optional KAH `event append artifact.updated` evidence

## Gate

FAIL if the rendered prompt weakens or changes the task contract. BLOCKED if no compatible prompt profile or template exists for the selected backend/path/phase.
