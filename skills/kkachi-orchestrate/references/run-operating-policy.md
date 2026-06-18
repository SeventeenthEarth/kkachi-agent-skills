# Run Operating Policy

This reference expands the operating-policy bullets in `../SKILL.md`.

## Lane ownership and mutation policy

- The selected implementer lane drafts and revises the substantive implementation plan and performs code, test, build, and task-bound docs mutations:
  - Stage 1 direct Codex SDK/app-server runner via `templates/runners/direct-codex-sdk-appserver-runner.py.tmpl` (`openai_codex` -> `codex app-server --listen stdio://`), not `codex exec` or generic `openai` SDK
  - Stage 2 KAB Codex-first through `native_codex`
  - Stage 3 selected eligible KAB backend
- Stage 1 Codex continuity is task-scoped: one task uses one recorded Codex `thread_id` across plan, implementation, feedback, cleanup, and verification-support turns when safe; the next task starts a new thread. Plan-only turns use effort `high`; non-plan turns use effort `medium`; deviations require artifacted rationale. This is not bound to the Discord/Hermes chat session, and the app-server subprocess may remain invocation-scoped.
- Blue/Red/Orange/Gray supervise, review, record evidence, ask/answer, and verify. They must not directly author the plan or patch repository artifacts as a substitute for the selected implementer unless 주군 explicitly asks for direct role editing or the work is outside the roadmap/KAS+KAH path.
- Record any exception and its no-Codex/backend rationale in KAH artifacts and the final report.

## Plan-vet loop

- Blue owns synthesis, with Red and Orange plan vet required for active KAS/KAH roadmap source policy, workflow, template, test, or shared skill mirror work. Red/Orange plan-vet reviewers and project-Gray documentation/integrity review are resolved through the project/team role registry when applicable, not hard-coded to individuals.
- Blue asks the selected planner lane for a plan-only draft; the backend returns the plan; required reviewers vet it; any `REQUEST_CHANGES` goes back to the same planner lane for revision; only required approval unlocks implementation.
- Do not let Blue, Red, Orange, or Gray rewrite the substantive plan as a shortcut.

## Fallback audit policy

- Blue+Red+Orange plan vet and later color reviews must include fallback audit when active KAS/KAH roadmap policy requires those plan reviewers.
- The preferred outcome is no fallback and fail-closed behavior for missing capability, evidence, approval, or safe state.
- Accept fallback only when it is unavoidable, bounded, evidence-backed, approval-safe, and very small.
- Broad fallback design or unclear policy must be reported to 주군 for a decision instead of merged into the run silently.

## Review and feedback policy

- First color review is the default review gate for every active KAS/KAH run that creates or changes durable repository artifacts, even when the task class is `docs_only`, `research_evidence`, `bootstrap_config`, or `collaboration_review`.
- Capture Blue self-review plus durable Red/Orange/Gray role-review evidence, or mark the review phase `not_applicable` with a concrete reason only for pure read-only/direct command runs where no durable project artifact changed.
- `delegate_task`, temporary subagents, and ad hoc advisor notes are pre-review analysis only. They do not substitute for official color review, MAR role coverage, project-Gray documentation/integrity review, or KAH evidence.
- For async review fan-in, attach the watcher as a mechanical observer only. If direct Kanban tools are absent, use the durable Hermes Kanban CLI surface before declaring review unavailable.
- Logical backend roles are planner (`plan`, `ask`), implementer (`implement`, `enhance-test`, `ai-slop-cleaner`, `optimize`, `docs-update`, `handle-feedback`), and feedback (`request-feedback`). They may map to the same or different physical backends.

## MAR review policy

- MAR review is mandatory for `development` / implementation tasks unless 주군 explicitly waives or replaces it before start and the decision is recorded in KAH/run evidence artifacts.
- For active KAS/KAH source policy, workflow, template, test, and shared skill mirror work, MAR is the only independent implementation review lane. Do not promote prior Octo-style review wording as a default, optional, fallback, or legacy review path in active artifacts.
- MAR must use role-first required coverage for `logic`, `security`, `arch`, `cve`, and `test_adequacy` with declared primary/secondary provider lanes.
- Provider preflight/toolchain proof, bounded raw-output artifacts, parsed findings, merge pack, Blue disposition, and any Red adjudication handoff must be preserved as evidence.
- Provider availability, prompt rendering, dispatch success, degraded providers, failed providers, or unresolved required roles never count as clean review completion.
- When MAR feedback changes the work, handling its feedback and a fresh post-change Blue + Red/Orange/Gray re-review are mandatory before final report or commit approval.
