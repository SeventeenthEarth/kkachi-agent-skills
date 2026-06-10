# Checklist Normalization

This reference expands the checklist rule in `../SKILL.md`.

`checklist.md` is mandatory, not optional. KAB does not directly provide the normalized KHS checklist. When a KAB planner lane is used, KAB provides `plan.plan_text`, and KHS should ask the planner to include a parse-friendly `KHS Checklist Seed` inside that plan. KHS/Hermes must transform that seed plus the KHS phase contract into a normalized KHS progress checklist and store it as the KAH `checklist.md` artifact during the plan phase, after the plan source is captured. Then update it after `ask` and after every later phase. The checklist is the operator-facing progress tracker for the KHS run.

The checklist must include:

- one row for every canonical KHS phase
- required/conditional applicability for each phase
- current state (`pending`, `in_progress`, `done`, `skipped`, or `blocked`)
- owner role (`planner`, `implementer`, `feedback`, or `Hermes`)
- backend/session when a KAB lane is used
- evidence artifact expected for completion
- gate/check command or review condition
- explicit skip reason for any skipped or not-applicable phase
- micro-task rows derived from the approved plan
- CodeGraph refresh evidence when required
- repeated `make test` checkpoints after implementation, test enhancement, AI slop cleanup, and optimization when those stages change files

For code-change runs, include an `optimize` row by default. It may be skipped only with a reason. For feedback, include round 1 as required and rounds 2..5 as conditional continuation rounds; do not exceed five feedback/handle-feedback pairs.
