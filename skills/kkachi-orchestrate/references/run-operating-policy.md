# Run Operating Policy

This reference expands the operating-policy bullets in `../SKILL.md`.

## Lane ownership and mutation policy

- The selected implementer lane drafts and revises the substantive implementation plan and performs code, test, build, and task-bound docs mutations:
  - Stage 1 direct Codex SDK/app-server runner via `templates/runners/direct-codex-sdk-appserver-runner.py.tmpl` (`openai_codex` -> `codex app-server --listen stdio://`), not `codex exec` or generic `openai` SDK
  - Stage 2 KAB Codex-first through `native_codex`
  - Stage 3 selected eligible KAB backend
- Blue/Red/Orange/Gray supervise, review, record evidence, ask/answer, and verify. They must not directly author the plan or patch repository artifacts as a substitute for the selected implementer unless 주군 explicitly asks for direct role editing or the work is outside the roadmap/KAS+KAH path.
- Record any exception and its no-Codex/backend rationale in KAH artifacts and the final report.

## Plan-vet loop

- Blue and Red own the plan vet/approval gate.
- Blue asks the selected planner lane for a plan-only draft; the backend returns the plan; Blue and Red vet it; any `REQUEST_CHANGES` goes back to the same planner lane for revision; only Blue+Red approval unlocks implementation.
- Do not let Blue or Red rewrite the substantive plan as a shortcut.

## Fallback audit policy

- Blue/Red plan vet and later color reviews must include fallback audit.
- The preferred outcome is no fallback and fail-closed behavior for missing capability, evidence, approval, or safe state.
- Accept fallback only when it is unavoidable, bounded, evidence-backed, approval-safe, and very small.
- Broad fallback design or unclear policy must be reported to 주군 for a decision instead of merged into the run silently.

## Review and feedback policy

- First color review is the default review gate for every active KAS/KAH run that creates or changes durable repository artifacts, even when the task class is `docs_only`, `research_evidence`, `bootstrap_config`, or `collaboration_review`.
- Capture Blue self-review plus durable Red/Orange/Gray role-review evidence, or mark the review phase `not_applicable` with a concrete reason only for pure read-only/direct command runs where no durable project artifact changed.
- Logical backend roles are planner (`plan`, `ask`), implementer (`implement`, `enhance-test`, `optimize`, `docs-update`, `handle-feedback`), and feedback (`request-feedback`). They may map to the same or different physical backends.

## MAR review policy

- MAR review is mandatory for `development` / implementation tasks unless 주군 explicitly waives or replaces it before start.
- MAR must use role-first required coverage for `logic`, `security`, `arch`, `cve`, and `test_adequacy` with declared primary/secondary provider lanes.
- Provider preflight/toolchain proof, bounded raw-output artifacts, parsed findings, merge pack, Blue disposition, and any Red adjudication handoff must be preserved as evidence.
- Provider availability, prompt rendering, dispatch success, degraded providers, failed providers, or unresolved required roles never count as clean review completion.
- When MAR feedback changes the work, handling its feedback and a fresh post-change Blue + Red/Orange/Gray re-review are mandatory before final report or commit approval.
