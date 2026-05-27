# STALECLEAN-001 stale-surface audit manifest

Status: complete
Run: `run-20260527T105713Z-158d70000c08`
Task: `STALECLEAN-001 — Audit active KAS stale surfaces`

## Scope

Audit only. No active skill/template/registry/doc cleanup is performed in this task unless 주군 separately authorizes STALECLEAN-002+ implementation.

Searched active surfaces requested by roadmap:

- `README.md`
- `docs/**/*.md`
- `skills/*/SKILL.md`
- `registries/*`
- `templates/**/*`

Marker classes:

1. fixed `1..3` / `max_rounds: 3` / `maximum_rounds: 3` / round-3 feedback assumptions;
2. `blocked_by_kah` wording for surfaces where KAH 0.1.4 graph/configurable-feedback substrate is now evidenced;
3. manual/direct/fallback `.kkachi-workflow.yaml` repair wording;
4. `kah graph` alias or `kah_graph_*` references that could be mistaken for implemented alias support or canonical GRAPHMVP-004 evidence.

## Search summary

Scripted active-surface scan found 198 candidate marker hits:

| Marker class | Count |
|---|---:|
| `fixed_1_3_feedback` | 51 |
| `blocked_by_kah` | 9 |
| `manual_workflow_yaml` | 30 |
| `kah_graph_alias` / `kah_graph_*` | 108 |

Preliminary disposition counts:

| Disposition | Count | Meaning |
|---|---:|---|
| `active_update_candidate` | 43 | Active skills/templates/registries likely need STALECLEAN-002+ normalization or explicit candidate/historical labeling. |
| `historical_or_guardrail` | 57 | Negative assertions, explicit stale markers, historical notes, or guardrails; preserve unless wording is confusing. |
| `review_needed` | 98 | Docs examples/SOT/README references need line-level judgment before edit/defer. |

## Active update candidates

These are the highest-value follow-up inputs for STALECLEAN-002 and related cleanup tasks.

### Feedback-intake fixed round-3 assumptions

| Path | Lines | Current marker | Proposed disposition |
|---|---:|---|---|
| `registries/phase-contracts.yaml` | 73, 203, 304, 305 | `maximum_rounds: 3`, `conditional_feedback_rounds_2_to_3`, handle-feedback artifacts stop at round 3 | Update in STALECLEAN-002 to required round 1 + optional rounds 2..5 where KAH-evidenced configurable feedback applies. Preserve `kas-integration-pending` until end-to-end KAS adoption is verified. |
| `skills/kkachi-handle-feedback/SKILL.md` | 22 | handle-feedback round bound still fixed to 3 | Update in STALECLEAN-002 with optional 2..5 guidance and fail-closed bounds. |
| `skills/kkachi-plan/SKILL.md` | 81 | include rounds 2-3 only; do not exceed three pairs | Update in STALECLEAN-002 to configured bounds / optional 2..5. |
| `templates/run-artifacts/checklist.md.tmpl` | 41, 42 | request/handle feedback 3 as maximum | Update in STALECLEAN-002 to generated/configured optional rounds. |
| `templates/run-artifacts/phase-plan.yaml.tmpl` | 72, 177, 186, 192, 193, 194 | max 3 and explicit round-3 finality | Update in STALECLEAN-002; likely needs matching docs-contract/template validation. |

### Graph evidence / alias field candidates

| Path | Lines | Current marker | Proposed disposition |
|---|---:|---|---|
| `registries/graph-template-registry.yaml` | 200-202 | `kah_graph_init_json`, `kah_graph_validate_json`, `kah_graph_explain_json` remain in `kas-default` gate evidence | Non-blocking from GRAPHMVP-004 Red review; normalize or document as raw KAH JSON gate evidence in later cleanup. |
| `skills/kkachi-task-contract/SKILL.md` | 33 | `kah_graph_evidence` / graph evidence naming | Review for alignment with GRAPHMVP-004 `graph-evidence.md` canonical fields. |
| `templates/run-artifacts/capability-check.md.tmpl` | 36, 50 | graph evidence / direct YAML fallback guard | Likely guardrail, but review for wording clarity. |
| `templates/run-artifacts/phase-plan.yaml.tmpl` | 8, 14-15 | graph preflight/evidence and no direct fallback | Likely guardrail; verify no alias overclaim. |
| `templates/run-artifacts/task-contract.yaml.tmpl` | 13, 19, 33 | graph evidence fields | Likely aligned after GRAPHMVP-004; verify field naming. |
| `templates/run-artifacts/graph-evidence.md.tmpl` | 20, 53, 54, 67, 68 | `kah_graph_audit_event_ids` and `kah graph` not assumed guard | Canonical GRAPHMVP-004 content; preserve. |
| `skills/kkachi-final-verify/SKILL.md` | 25, 27 | rejects `kah graph` alias as implemented; checks canonical fields | Guardrail; preserve. |
| `skills/kkachi-orchestrate/SKILL.md` | 48 | no `kah graph` alias/direct fallback | Guardrail; preserve. |
| `skills/kkachi-phase-state/SKILL.md` | 89-90 | no manual repair; `kah_graph_audit_event_ids` | Guardrail/canonical; preserve unless wording needs consistency. |

## Docs and SOT review-needed groups

These are not immediate active behavior changes without line-level review.

| Path | Marker class | Lines / note | Proposed disposition |
|---|---|---|---|
| `docs/README.md` | `blocked_by_kah`, `kah graph`, `.kkachi-workflow.yaml` | status vocabulary and graph boundary lines | Mostly authority labels and guardrails. Keep if clearly historical/candidate; adjust only if it implies current KAH support is absent. |
| `docs/roadmap.md` | all marker classes | STALECLEAN task definitions, deferrals, stale marker section | Source roadmap; preserve as task authority until later task updates statuses. |
| `docs/sot/external-feedback-intake-policy.md` | fixed round 3, `blocked_by_kah`, graph | existing stale manifest and policy status labels | Use as prior inventory; STALECLEAN-002 should update active surfaces, not erase useful audit history. |
| `docs/sot/khs-architecture-and-integration.md` | fixed round 3, manual YAML, graph | active stale-surface inventory and acceptance matrix | Use as prior inventory; update after cleanup tasks to reflect resolved surfaces. |
| `docs/sot/khs-pre-kah-readiness-audit-2026-05.md` | `blocked_by_kah`, fixed round 3 | historical pre-KAH audit | Preserve historical status; ensure reports never treat it as current blocker. |
| `docs/sot/workflow-graph-integration.md` | manual YAML, `kah graph` | current graph integration SOT | Most matches are guardrails; preserve unless wording implies alias implementation. |
| `docs/sot/interface-contract.md` | graph/manual YAML | interface boundary and alias candidate notes | Mostly guardrail; preserve with current status labels. |
| `docs/sot/concept.md` and `docs/sot/skill-template.md` | fixed round 3 / graph | older conceptual/template guidance | Review whether active or historical; likely STALECLEAN-004 verification surface. |
| `docs/examples/graph-capability-preflight.md` | manual YAML / graph | negative examples | Preserve as guardrail if clear. |

## Suggested follow-up mapping

| Follow-up task | Inputs from this manifest | Recommended action |
|---|---|---|
| `STALECLEAN-002` | `registries/phase-contracts.yaml`, `skills/kkachi-handle-feedback`, `skills/kkachi-plan`, `templates/run-artifacts/checklist.md.tmpl`, `templates/run-artifacts/phase-plan.yaml.tmpl` | Normalize feedback-intake bounds to required round 1 + optional rounds 2..5 where KAH-evidenced configurable feedback applies; add tests/search guards. |
| `STALECLEAN-003` | docs/SOT groups with KAB/runtime boundary wording | Verify KAB-later is preserved only for runtime/backend evidence, not CLIMVP/GRAPHMVP blockers. |
| `STALECLEAN-004` | all remaining `review_needed` docs and preserved guardrails | Final search/readback proof that active KAS guidance no longer treats KAH 0.1.4 graph/configurable-feedback as absent. |

## Verification evidence

Commands/tools used:

- `read_file docs/roadmap.md` lines 103-112 and 158-160 to confirm STALECLEAN-001 is next.
- `kkachi-agent-helper project status --json` before run creation: health `ok`, no active run.
- `kkachi-agent-helper run create ... --task-id STALECLEAN-001 ... --execution-mode docs_only --backend-evidence not_applicable` -> run `run-20260527T105713Z-158d70000c08`.
- `kkachi-agent-helper run activate run-20260527T105713Z-158d70000c08`.
- `kkachi-agent-helper artifact init run-20260527T105713Z-158d70000c08`.
- Four targeted `search_files` searches over README/docs/skills/registries/templates.
- Scripted active-surface scan over the same file families using four regex marker classes.
- Repo-visible audit artifact written at `docs/discussions/staleclean-001-active-surface-manifest.md`.

## Current decision

STALECLEAN-001 is complete as an audit/manifest task without modifying active behavior surfaces. The current manifest identifies concrete update candidates for STALECLEAN-002 and separates guardrail/historical wording from active stale behavior.
