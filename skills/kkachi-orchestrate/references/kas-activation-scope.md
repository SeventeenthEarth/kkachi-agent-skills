# KAS Activation Scope and Task Classification

Use this reference when updating or applying KAS/Kkachi task taxonomy.

## Durable lesson

Task classification is not a global Hermes/Hwangchung behavior. Default mode is direct commander response: answer, inspect, or perform bounded non-durable work without creating a KHS task.

KAS mode starts only when at least one activation trigger is present:

- the master explicitly selects KHS/Kkachi/KAB/KAH or a Kkachi run;
- KHS is being applied to a project directory;
- a durable repo artifact under a governed Kkachi project is being changed;
- phase/gate/evidence tracking is requested;
- KAB-backed execution, KAB plan lifecycle, bridge evidence, or backend identity through KAB is claimed; default v0.2 KAS/KAH work records KAS policy, KAH evidence, GJC candidate refs, and KAT factual refs instead of KAB backend evidence;
- long-lived team-member collaboration or durable review routing is required.

Inside active KAS mode, classify the task before selecting phases. Outside KAS mode, do not force `task_class` or create KAH runs for ordinary chat.

## Mapping for investigation/spec/roadmap work

- Current-state investigation only: `research_evidence`.
- Investigation followed by spec/SOT/roadmap/acceptance/handoff edits: `docs_only + Path B shaping`.
- Any code, tests, build behavior, executable contract, or future execution-policy change: `development`.

## Reporting rule

When a user asks whether the taxonomy over-scopes the commander, state explicitly that taxonomy is a project-execution router, not the commander's global personality or all-chat constraint.
