# KAS Stage Report Contract

Date: 2026-07-09
Owner: KAS workflow/policy layer
Confirming role: Hwangchung / 황충, Kkachi Blue commander
Status: source-side v0.2.3 reporting contract
Authority level: KAS stage-report SOT for operator-facing summaries
Scope: KAS skills, docs, and report guidance for task classification, plan, implementation, and review closeout reports
Non-scope: KAH state schema, KAT factual test authority, KAB runtime control, live install, tag publication, commit/push approval

## Purpose

KAS reports must let 주군 understand the functional change without reading every
code diff or artifact. A stage report is not a file list. It is a compact
operator-facing summary of what feature/change is being pursued, why it matters,
how the plan or implementation changed behavior, what review feedback changed,
and which evidence supports the claim.

Detailed logs, diffs, prompts, test output, and long review bodies belong in
artifacts. The chat or final stage report must stay short, but it must include
enough semantic content for Project Leader judgment.

## Repository self-development exception

KAS/KAH/KAT repository development does not dogfood the KAS/KAH/KAT workflow by
default. 황충 performs main development directly, then routes the result through
official color review and fixes/re-review. KAS/KAH/KAT dogfood workflows may be
used only after explicit 주군 selection for that run.

This exception does not waive official color review, evidence-backed testing,
commit approval, release/tag approval, or fail-closed handling of blockers.

## Common report shape

Every stage report should use compact Korean for 주군 unless another output
language is explicitly requested. It must include:

- **Functional change:** what functional behavior, operator workflow, or skill guidance is being added or changed.
- **Reason:** what old behavior, missing information, or operator pain made the change necessary.
- **Evidence basis:** authoritative docs, SOT, task request, code/test evidence, or review IDs checked.
- **Verification:** tests, static checks, review gates, or explicit not-applicable reason.
- **Remaining decision:** approval, risk, release, install, commit, or follow-up decision still required.

Do not report only `completed`, `files changed`, or `stage done` for these
stages.

## Task classification report requirements

A classification report must state:

1. **Feature/change identity:** the function, skill behavior, report format, or workflow surface being added/changed.
2. **Need reason:** the current/old behavior that makes the feature necessary.
3. **Document direction:** which docs/SOT/roadmap/request direction says how the work should proceed.
4. **Task class and route:** selected class (`development`, `docs_only`, `research_evidence`, etc.) and why that route fits.
5. **Non-scope and gates:** what is not authorized yet, especially implementation, release, install, commit, push, runtime, provider, auth, or gateway mutation.

## Plan-stage report requirements

A plan-stage closeout report must state:

1. **Realization strategy:** how the documented plan will be implemented in practical terms.
2. **Plan drift:** whether the original document plan was followed, narrowed, or changed; if changed, what changed and why.
3. **Color-vet feedback:** what Red, Orange, Gray, and Blue/creator synthesis found and how the plan was revised.
4. **Verification plan:** which tests/checks/review evidence will prove the implementation.
5. **Approval boundary:** whether implementation is held for 주군 approval or already inside a bounded approved scope.

## Implementation-stage report requirements

An implementation-stage closeout report must state:

1. **Behavior before/after:** what used to happen and what now happens.
2. **Plan conformance:** whether the accepted plan was applied as written; if not, list the meaningful deviations and reasons.
3. **Changed surfaces:** concise files/components list grouped by purpose, not raw path spam.
4. **Phase-step coverage:** whether `impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, and `docs-update` ran or why a step was not applicable.
5. **KAH/KAT/test coverage:** whether registered KAH tests, selected unit/integration/e2e/docs-contract checks, and any KAT factual evidence passed; include failures or not-applicable reasons.
6. **Remaining risk:** unresolved behavior risk, review risk, release/install/commit boundary, or evidence gap.

## Review-stage report requirements

A review-stage closeout report must state:

1. **Color review result:** Red/Orange/Gray/Blue verdicts and the core issue each color did or did not find.
2. **MAR result:** MAR status, role coverage, findings, waiver/N/A reason, and Blue disposition when applicable.
3. **Second-color adoption:** whether post-MAR/second-color review occurred, what feedback remained, and how it was adopted, fixed, rejected, deferred, or waived.
4. **Defer reason:** for every deferred review/MAR/second-color item, what was deferred, why it was safe or necessary to defer now, and which future gate/task owns it.
5. **Fix evidence:** what changed after review feedback and which verification was rerun.
6. **Final boundary:** whether the result is release-ready, install-ready, commit-ready, or still held.

## Fail-closed cases

Report `blocked` or `request changes` instead of completion when:

- the report cannot explain the functional change or reason for the work;
- plan drift occurred but is not documented;
- required color/MAR/second-color evidence is missing or only represented by temporary subagents;
- a review report says an item was deferred but does not briefly state what was deferred, why, and where it will be handled;
- tests are missing without a concrete not-applicable reason;
- the work claims release/install/commit/readiness without current binary, diff, or review evidence;
- a KAS/KAH/KAT self-development run accidentally starts dogfooding after 주군 selected direct development.
