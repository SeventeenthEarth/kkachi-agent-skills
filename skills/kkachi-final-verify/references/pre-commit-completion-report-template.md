# Pre-commit Completion Report Template

Use this template for 주군-facing Kkachi/KAS reports before asking for commit approval. Keep it concise but evidence-backed. If a section is not applicable, write `해당 없음` and give the reason.

## Required Korean report sections

1. `상태 / 커밋 승인 요청`
   - Task id/title, current state (`review-ready uncommitted` or `committed` only when already approved), KAH run id, gate status.

2. `Enhance Test`
   - Count tests added by lane when applicable: `unit +N`, `integration +N`, `e2e +N`.
   - Name key added test files, or say none added with reason.
   - Include final verification command summary.

3. `AI Slop Cleaner / Optimize`
   - Summarize cleanup: comments/prose, over-abstraction, speculative branches, naming, duplicate code, noisy docs, etc.
   - State whether files changed and which verification reran after cleanup, or mark not applicable with reason.

4. `Docs / Roadmap`
   - List durable docs/skills updated.
   - State whether the roadmap/backlog task was marked completed; if not, state why.

5. `1차 Blue/Red/Orange/Gray 리뷰 및 개선`
   - List reviewer/card ids and verdicts.
   - Summarize requested changes/improvement points and how they were applied or deferred.
   - This section is required for KAS/KAH runs with durable repo/artifact changes, even when the task is not implementation.

6. `GLM Octo Review 및 개선`
   - For implementation tasks, include official `/octo:review` evidence or an explicit pre-start 주군 waiver/blocker. For non-implementation tasks, state whether Octo was requested/declared/run, or mark `해당 없음` with the reason when Octo was not part of the active policy.
   - When Octo ran, include session id, watcher/evidence path, and verdict.
   - Summarize GLM findings by severity and disposition: fixed/deferred/rejected.
   - If only plain GLM review ran, say it is not official Octo gate evidence.

7. `재리뷰 및 개선 확인`
   - Required for implementation tasks after official Octo, and for any other task when official Octo ran or another later feedback round changed the work after first color review; otherwise mark `해당 없음` with the reason.
   - List Blue/Red/Orange/Gray re-review card ids and verdicts when required.
   - Summarize remaining risks or confirm no commit-blocking issues.

8. `남은 위험 / 다음 작업`
   - Separate blockers, deferred risks, assumptions, and next epic/task.

9. `추천 커밋 메시지`
   - Provide one-line English commit message, e.g. `[TASK-ID] concise action summary`.

## Reporting rules

- Do not ask for commit approval until all required sections are present or explicitly marked not applicable.
- Prefer evidence handles: KAH artifact path, Kanban card id, KAB session id, command name, or gate report.
- Do not hide deferred review findings. Mark them as `deferred` with the owning future gate/task.
- Keep report source of truth aligned with `final-report.md` and KAH events.
