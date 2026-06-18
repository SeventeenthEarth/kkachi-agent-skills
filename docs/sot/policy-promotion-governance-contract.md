# POLPR policy-promotion governance contract

Date: 2026-06-18
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / Blue, Red, Orange, and project-Gray review evidence
Status: planning SOT for POLPR docs/SOT intake; not implementation authorization by itself
Authority level: KAS policy-promotion planning authority
Scope: KAS source policy, workflow graph defaults, CLI contract surfaces, skills/runbooks, docs-update/final-review requirements, and KAS/KAH companion task sequencing
Source evidence: `/Users/draccoon/Workspace/Hermes/17thHermes/40_outputs/projects/kkachi/2026-06-14-kas-policy-promotion-candidates.md`
Paired KAH SOT: `kkachi-agent-helper/docs/sot/policy-promotion-helper-evidence.md`
Epic: `POLPR` — policy-promotion and workflow-governance alignment

## Purpose

POLPR names and scopes the accumulated KAS/KAH policy-promotion batch so later work can refer to one roadmap epic instead of a loose candidate note. The candidate note remains source evidence, but this SOT is the KAS-side planning authority for promoting accepted lessons into repository policy, templates, skills, tests, and KAH companion evidence surfaces.

This SOT does not itself patch workflow behavior, activate new KAH gates, install profile skills, mutate runtime/provider/auth configuration, push branches, or replace required review gates. Implementation still proceeds one PR-candidate task at a time after plan approval, review, verification, and explicit 주군 authorization where required.

## Epic name and task prefix

- Epic slug: `POLPR`
- Epic title: policy-promotion and workflow-governance alignment
- Shared cross-repo PR/task ids: `POLPR-001` through `POLPR-008`
- KAS-owned slices: `POLPR-001`, `POLPR-003`, `POLPR-004`, `POLPR-006`, and `POLPR-008`
- KAH-owned companion slices: `POLPR-002`, `POLPR-005`, and `POLPR-007`

## Single-epic multi-repo task strategy

POLPR deliberately uses one epic id with monotonically increasing task ids shared across both repositories. This strategy is allowed only when the responsible approver explicitly authorizes it for a small, tightly coupled multi-repository project where a single ordered sequence improves review, rollout, and operator traceability.

Do not apply this strategy by default. Large repositories, repositories with independent product/release lifecycles, or broad projects whose components need separate ownership should use independent per-repository epics and task sequences instead. If the repository set is ambiguous, fail closed to separate epics until the responsible approver explicitly chooses shared numbering.

When shared numbering is approved, each roadmap should list only the tasks owned by that repository, while the SOT records the complete cross-repository sequence so reviewers can see the total order from the task id alone. Each task id still belongs to exactly one repository PR candidate unless a later approved plan explicitly splits it.

## Promotion principles

1. **Plan-first implementation gate:** roadmap/source-policy changes that affect KAS/KAH development flow require a plan artifact, impact map, Blue vet, Red technical review, and Orange operator-value review before implementation starts.
2. **MAR-only independent review:** KAS/KAH source policy, workflow phases, templates, tests, and skill mirrors must not promote `GLM Octo` review wording. Independent review is MAR-only unless 주군 explicitly authorizes a separate non-KAS/KAH legacy note.
3. **Official review evidence boundary:** `delegate_task`, temporary subagents, local helper processes, and draft model feedback are not official color-review or MAR evidence. Official long-lived routing uses Kanban/KAH evidence and named review artifacts.
4. **Default, configurable workflow spine:** KAS may define a no-input default development workflow, but project-specific/custom graphs remain supported when they pass the applicable KAS/KAH supportability envelope.
5. **Repo-local agent-instruction lifecycle:** `AGENTS.md` and `CLAUDE.md` lifecycle work is a KAS repo/project command surface with dry-run/apply hash discipline and managed-block preservation; it is distinct from Hermes profile-local skill installation.
6. **Test-layer clarity:** unit, integration, and e2e labels must match resource/process/DB usage, and e2e guidance must require isolated disposable resources.
7. **Failed-test ownership split:** Blue owns reproduction/log triage/routing/acceptance; the selected implementer lane owns detailed root-cause analysis, solution finalization, and code mutation unless 주군 approves a direct Blue patch exception.
8. **Kanban watcher fallback:** async watchers default to a bounded about-five-minute cadence; when direct `kanban_*` tools are absent, use `hermes kanban ...` instead of declaring Kanban unavailable.
9. **Document impact map and project-Gray review:** docs-update/final-review must include an impact map for affected SOT, roadmap/status, README/operator docs, repo-local agent instructions, registries/templates/examples, skills/runbooks, and KAH evidence/gate/final-report templates when relevant. The Gray reviewer is resolved from the active Kkachi-using project's role/registry authority, not hard-coded to one individual.
10. **Explicit shared-numbering approval:** single-epic, multi-repo shared task numbering is an opt-in strategy for small tightly coupled repo sets, not a default project-management rule.
11. **No fallback widening:** new fallback behavior must be approval-gated, bounded, evidenced, and fail-closed when capability, approval, or safe state is missing.

## PR-candidate slices

| Task ID | Title | Primary scope | KAH companion |
|---|---|---|---|
| `POLPR-001` | KAS docs/SOT and roadmap registration | Create this SOT, register the KAS side of POLPR in roadmap/docs map/index, and preserve candidate-source evidence. | Paired with KAH `POLPR-002`. |
| `POLPR-002` | KAH helper docs/SOT and roadmap registration | KAH records the helper-side planning SOT and roadmap entry using the same `POLPR` prefix. | KAH repo PR; no helper behavior claim. |
| `POLPR-003` | KAS review governance and policy cleanup | Promote plan/vet, color-review, MAR-only, official evidence boundary, project-Gray, failed-test split, and watcher fallback rules into KAS policy/skills. | KAH guidance only where helper evidence/final reports need deterministic fields or wording. |
| `POLPR-004` | KAS default workflow graph policy/template alignment | Update KAS phase policy, graph registry/template, tests, and default configurable spine; remove `octo-review` phase naming from active KAS surfaces. | Precedes KAH default phase-plan companion `POLPR-005`. |
| `POLPR-005` | KAH default phase-plan and MAR naming support | Update KAH default phase-plan support/tests from `octo-review` to `mar-review` without making KAH the policy owner. | KAH repo PR; custom workflows remain supportability-based. |
| `POLPR-006` | KAS agent-instruction lifecycle and test-layer contract | Add/align `AGENTS.md` / `CLAUDE.md` lifecycle contract, help/test expectations, test taxonomy, e2e isolation, and failed-test repair rules. | KAH evidence fields remain separate in `POLPR-007` if needed. |
| `POLPR-007` | KAH deterministic docs/test/review evidence support | Add deterministic evidence labels or docs wording only if KAH surfaces need to record impact-map, project-Gray, test-layer, or failed-test ownership fields. | Evidence presence/shape only; KAS owns policy and reviewer meaning. |
| `POLPR-008` | KAS skill mirror, stale scan, and rollout closure | Mirror accepted source policy into KAS skills/profile guidance, perform stale `GLM Octo`/compressed workflow/docs-impact scans, and collect final review evidence. | Validate KAH companion docs/tests are aligned and no stale KAH `GLM Octo` review wording remains. |

## Impact map baseline

POLPR implementation planning must inspect at least these surfaces before source mutation:

- KAS SOT docs: `docs/sot/phase-orchestration-policy.md`, `docs/sot/kas-cli-contract.md`, `docs/sot/multi-agent-review-policy.md`, this SOT, and any SOT that names review, graph, test, docs-update, or agent-instruction behavior.
- KAS roadmap/docs index/docs map: `docs/roadmap.md`, `docs/README.md`, `docs/kkachi-docs-map.yaml`.
- KAS registries/templates/examples: `registries/graph-template-registry.yaml`, `templates/workflow-graphs/kas-default.yaml`, run-artifact templates, and tested contract fixtures.
- KAS skills/runbooks: plan, review, request-feedback, handle-feedback, docs-update, final-verify, orchestrate, workflow, backend, and prompt-composition skills when their instructions overlap POLPR policy.
- KAS CLI/help/tests: command surfaces and tests affected by agent-instruction lifecycle, graph defaults, doctor/repair, and test-layer policy.
- KAH companion surfaces: `docs/roadmap.md`, `docs/README.md`, `docs/kkachi-docs-map.yaml`, `docs/sot/policy-promotion-helper-evidence.md`, `internal/project/phase_plan.go`, relevant gate/evidence tests, specs/compatibility docs only when helper behavior changes.

## Acceptance criteria for POLPR-001

- This SOT exists and records the epic name, source evidence, shared-numbering strategy, principles, PR-candidate slices, impact map baseline, boundaries, and deferrals.
- KAS `docs/roadmap.md` registers `POLPR` in delivery order and active roadmap with the KAS-owned shared-number slices.
- KAH `docs/roadmap.md` registers `POLPR` in delivery order and active roadmap with the KAH-owned shared-number slices.
- KAS/KAH docs indexes and docs maps reference the new SOTs.
- Verification includes docs readback, YAML parse for docs maps, `git diff --check` in both repos, and repository test command or an explicit blocker/degraded reason.

## Deferrals and non-goals

- No source policy or code behavior beyond docs/SOT/roadmap registration is completed by `POLPR-001`.
- No profile skill install, runtime activation, provider/gateway/auth/token/model mutation, KAB activation, push, or release tagging is authorized.
- No automatic review comments, automatic code mutation by reviewers, model voting as authority, or warning-only MAR gate state is introduced.
- No universal forced graph shape is introduced; default workflow behavior remains configurable and supportability-based.

## Next action

After `POLPR-001` and companion `POLPR-002` are reviewed and accepted, start `POLPR-003` with a fresh impact map and plan gate. Implementation tasks must preserve source evidence links to the candidate note and this SOT, then update only the touched policy/code/docs/test surfaces for that slice.
