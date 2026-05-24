# AGENTS.md

Behavioral guidelines for agents working in `kkachi-hermes-skills` (KHS).
These rules merge general LLM coding guardrails with this repository's KHS,
KAH, and KAB ownership boundaries.

KHS is a Hermes prompt/process skill pack. It classifies work, builds
AI-neutral task contracts, selects phase guidance, selects backend prompt
profiles, and prepares evidence expectations. KAH owns deterministic project
state, artifacts, schemas, events, locks, diagnostics, and gates. KAB owns
backend runtime/session control and bridge execution evidence.

Do not turn KHS into a second state system.

## 1. Start From Authority

Before changing shared KHS behavior, read the relevant authority first:

- `README.md` for repository purpose, KHS/KAH/KAB boundaries, install flow, and
  run shape.
- `docs/README.md` for the documentation authority ladder.
- Relevant `docs/sot/*` files for durable source-of-truth behavior.
- Relevant `registries/*` and `templates/*` when changing backend policy,
  phase behavior, prompt profiles, or run artifacts.
- Relevant `skills/*/SKILL.md` files when changing a phase skill.

Treat documentation layers carefully:

- `docs/sot/` is durable authority for its stated scope.
- `docs/discussions/` is historical or evolving context until promoted into
  `docs/sot/`, a registry, template, or skill.
- `docs/roadmap.md` is planning guidance. It is not implementation
  authorization by itself.
- Planned or candidate graph/KAH surfaces remain planned until current
  capability/help evidence proves they exist.

## 2. Think Before Coding

Do not assume. Do not hide confusion. Surface tradeoffs.

Before substantial edits:

- State the working assumptions and success criteria.
- Investigate discoverable repository facts first.
- If multiple product or authority interpretations remain, present them.
- Ask only for user judgment: scope, authority, risky tradeoffs, destructive
  actions, credentials, external-production effects, or materially branching
  choices.
- If a simpler approach solves the request, say so and keep the change small.

For KHS work, "do not assume" means evidence-first investigation, not stopping
to ask about facts that the repository can answer.

## 3. Simplicity First

Write the minimum code or documentation that solves the request.

- No features beyond what was asked.
- No abstractions for single-use behavior.
- No speculative configurability.
- No invented KAH or KAB command surfaces.
- No long duplicated compatibility tables when the bridge compatibility matrix
  or registry is the source of truth.
- If a change can be made in a project overlay or run artifact first, prefer
  that over promoting new reusable behavior into shared KHS.

Ask whether a senior engineer would call the result overcomplicated. If yes,
simplify before finishing.

## 4. Surgical Changes

Touch only what the task requires. Clean up only your own mess.

When editing existing files:

- Match the existing style and terminology.
- Do not reformat, rename, or refactor unrelated content.
- Do not delete unrelated dead code or stale text unless explicitly asked.
- If you notice unrelated drift, mention it instead of silently fixing it.
- Remove imports, variables, functions, or references made unused by your own
  changes.

When changing shared skills, templates, registries, or SOT docs:

- Keep task-contract language AI-neutral.
- Do not put Claude, Codex, Gemini, GLM, or OpenCode prompt style into
  acceptance criteria, constraints, or non-goals.
- Keep `phase-plan.yaml` as the KHS run workflow SOT and `checklist.md` as the
  operator-facing progress tracker where run guidance discusses execution.
- Keep KAH metadata described as deterministic helper state, not phase
  authority.
- Require KAB evidence for KAB-backed work. `send` success is dispatch evidence
  only, not completion evidence.

Every changed line should trace directly to the user request.

## 5. Goal-Driven Execution

Turn work into verifiable goals and loop until checked.

For multi-step tasks, keep a brief plan in this shape:

```text
1. [Step] -> verify: [check]
2. [Step] -> verify: [check]
3. [Step] -> verify: [check]
```

Examples:

- "Add validation" -> write or update invalid-input checks, then make them pass.
- "Fix the bug" -> reproduce or identify the failure path, then verify the fix.
- "Refactor X" -> verify behavior before and after when behavior is not already
  protected.
- "Update KHS guidance" -> verify the affected authority ladder,
  cross-references, and KHS/KAH/KAB ownership boundaries.

For docs and skill edits, verification should include the smallest useful set of
file reads, cross-reference searches, schema/template consistency checks, or
tests needed to support the completion claim.

## 6. KHS Run Guidance

Only trigger KHS behavior when the user or operating context explicitly asks for
KHS, Kkachi, KAH, KAB, a Kkachi run, a KHS-using commander, bridge evidence, or
gate-backed artifacts.

Do not trigger KHS for ordinary direct edits, typo fixes, small one-file
patches, quick config tweaks, or read-only explanations unless explicitly asked.

When KHS is triggered:

- KHS prepares task contracts, phase plans, backend selection, prompt profiles,
  and evidence expectations.
- KAH records project state, run artifacts, events, schema validation, and gate
  verdicts.
- KAB controls backend sessions and preserves execution evidence.
- User backend preference may rank eligible lanes, but must not bypass required
  capabilities, project policy, or compatibility caveats.
- If KAB is forbidden for code-change KHS work, treat the task as a normal
  direct Hermes task instead of pretending it is a KHS run.

## 7. Completion Reports

Final reports should be concise and evidence-backed. Include:

- Changed files.
- Simplifications made or avoided.
- Verification performed.
- Remaining risks or known gaps.

Do not claim completion when required gates, artifacts, tests, docs decisions,
KAH state, KAB evidence, `phase-plan.yaml`, or `checklist.md` are incomplete for
the kind of work being reported.
