# Kkachi Multi-Agent Review policy

Date: 2026-06-17
Owner: KAS policy and skill layer
Status: candidate repository promotion SOT; repository implementation pending
Authority level: KAS planning authority for MAR promotion from accepted Obsidian output SOT; not installed skill behavior or runtime activation
Source SOT: `/Users/draccoon/Workspace/Hermes/17thHermes/40_outputs/projects/kkachi/2026-06-16-kkachi-multi-agent-review-mar-sot.md`
Paired KAH planning SOT: KAH `docs/sot/multi-agent-review-evidence-gates.md`
Scope: KAS MAR policy, skill, prompt templates, reviewer matrix, premium escalation rules, and script ownership

## Purpose

Kkachi Multi-Agent Review (MAR) is the planned lightweight KAS-local review lane for routine Kkachi development review. It collects specialized AI reviewer outputs, preserves full evidence through KAH artifacts, and returns a compact merge pack plus Blue disposition path. MAR is advisory evidence; Blue disposition and conditional Red adjudication remain the authority.

This repository document promotes the accepted output SOT into KAS planning authority. It does not claim that the `kkachi-multi-agent-review` skill, `mar.py`, provider adapters, KAH gates, or installed runtime behavior exist until the follow-up implementation tasks complete with evidence.

## Canonical operating rule

```text
Blue runs MAR by default.
KAS defines the skill and script.
KAH preserves evidence.
Red adjudicates only when risk, conflict, or workflow gates require it.
Codex and Claude require explicit 주군 approval unless pre-authorized.
zcode/glm-5.2, Kimi K2.6, and Antigravity are the default reviewer set once adapter proof exists.
Model consensus is advisory; Blue/Red disposition is authority.
```

## Task split and promotion boundary

MAR promotion is intentionally split so planning documents do not overclaim implementation.

| Task | Repository | Scope | Completion claim allowed |
|---|---|---|---|
| `MAR-001` | KAS | Promote MAR policy/SOT, roadmap/docs-map/docs-index records, stale GLM Octo marker targets, and implementation task boundaries. | KAS repository has a planning SOT for MAR. No skill/script/provider behavior claim. |
| `MAREV-001` | KAH | Promote KAH-side MAR artifact/gate planning SOT and roadmap/docs-index records. | KAH has a planning SOT for MAR evidence capture. No helper gate/code behavior claim. |
| `MAR-002` | KAS | Add `kkachi-multi-agent-review` skill scaffold, references, prompt templates, and disposition templates without provider execution. | KAS skill scaffold exists; provider execution still pending unless separately implemented. |
| `MAR-003` | KAS | Implement stdlib `mar.py` doctor/render/validate/merge-pack MVP using fixture and read-only local evidence. | Local MAR script surfaces exist for non-provider or mocked/fixture paths. |
| `MAR-004` | KAS | Implement provider run path after adapter proof and authorization boundaries are recorded. | Default reviewer execution is available only for validated providers and recorded degraded semantics. |
| `MAREV-002` | KAH | Implement deterministic KAH artifact/gate/schema support for MAR evidence if needed by final gates. | KAH helper behavior exists only after code/tests/docs/release evidence. |

`MAREV-001` is docs/SOT only. It must not be read as implementing KAH behavior. Actual KAH code or gate behavior belongs in `MAREV-002` or later.

## Ownership

KAS owns:

- `kkachi-multi-agent-review` skill definition;
- MAR policy and reviewer selection rules;
- prompt templates and reviewer role matrix;
- premium escalation policy;
- bundled `scripts/mar.py`;
- output schema and parsing rules;
- Blue disposition template guidance.

KAH owns deterministic evidence only after its paired tasks implement or document support:

- MAR request metadata;
- provider doctor/preflight results;
- input bundle and diff snapshot;
- rendered prompts;
- raw reviewer outputs;
- parsed findings;
- compact merge pack;
- Blue disposition record;
- optional Red adjudication record;
- final gate evidence when implemented.

KAB remains non-default for routine MAR. KAB is used only when a later task explicitly selects KAB runtime orchestration or premium/heavier review support.

## Default reviewer set

Default reviewer policy after adapter proof:

| Reviewer | Command lane | Primary role | Notes |
|---|---|---|---|
| `zcode_glm_5_2` | `zcode` to `glm-5.2` | Kkachi SOT, fail-closed, approval, fallback, and evidence-risk review | Legacy `glm` CLI is not a MAR lane. Non-interactive command template must be proven before success coverage. |
| `kimi_k2_6` | `kimi` to `k2.6` | Requirement, artifact, and traceability review | Unavailable Kimi creates degraded coverage; it must not become clean PASS. |
| `antigravity_gemini` | `agy` to selected Gemini model | Architecture, integration, security, and operational-risk review | Flash-class model is only sufficient for quick low-risk smoke; high-risk final MAR needs Pro-class or recorded policy. |

Codex and Claude are premium reviewers. They require explicit 주군 approval unless the active task contract already grants that escalation.

## Coverage and status semantics

MAR status values are:

```text
PASS
PASS_WITH_FINDINGS
REQUEST_CHANGES
BLOCKED
DEGRADED
FAILED
```

Rules:

1. `PASS` requires complete default coverage or a pre-scoped narrower reviewer set recorded before execution, and no actionable findings.
2. `DEGRADED` is not a soft pass. It requires a Blue reason and may require Red adjudication.
3. All-provider failure is `FAILED` or `BLOCKED`, never clean review.
4. One successful default reviewer on nontrivial development is insufficient coverage and requires Red adjudication before final Blue disposition.
5. Provider disagreement creates a disposition obligation, not a vote.
6. Premium review does not erase default coverage failures.

## Red adjudication triggers

Red adjudication is required when any of the following are true:

- blocker finding exists;
- two or more high-severity findings exist;
- providers disagree on a high/blocker issue;
- findings involve SOT, approval, fallback, runtime, auth, secret, security, architecture boundary, or KAS/KAH/KAB responsibility;
- all default reviewers fail;
- Blue confidence is below 0.75;
- premium escalation is suggested;
- task class is release readiness, architecture SOT change, or workflow policy change.

## Safety rules

MAR is read-only. Reviewer prompts and scripts must prohibit reviewer models from editing files, running tests, running builds, installing packages, starting services, probing networks, changing auth, changing secrets, changing provider settings, or mutating runtime state.

The KAS script must use `subprocess.run([...], shell=False)`, enforce provider timeouts, cap raw output size, preserve parse failures, compare git status before/after execution, fail closed on mutation detection, record unavailable providers as degraded coverage, and never interpret provider failure as clean review.

## Legacy GLM Octo handling

Existing KAS surfaces that require KAB-mediated GLM Octo as the default routine review lane are stale-marker targets during MAR promotion. They must not be silently deleted. Before replacement, each update must record:

- exact file path and section;
- whether the old wording is active, historical, alternate, or still required for a specific workflow;
- whether MAR dogfood evidence exists;
- whether KAB GLM Octo remains required by a separate active workflow gate.

Dogfood evidence is required before broad wording changes claim that MAR replaces GLM Octo as default routine review behavior.

## Verification before implementation claim

KAS must not claim MAR implementation until evidence exists for the specific claimed surface:

- docs-map/index/roadmap registration for the policy SOT;
- skill scaffold readback for `kkachi-multi-agent-review`;
- prompt template readback and reviewer-role matrix;
- `mar.py` help/doctor/render/validate tests;
- provider adapter proof for any reviewer counted as successful coverage;
- fixture tests for degraded, insufficient, failed, and request-changes semantics;
- KAH artifact/gate evidence once MAREV implementation is claimed;
- Red/Orange/Gray review or a recorded 주군-approved exception when required by workflow risk.

## Deferrals

Deferred unless separately approved: automatic PR inline comments, automatic code mutation, automatic test/build execution by reviewer models, model voting as authority, silent premium-provider fallback, profile/provider/gateway/auth/token/model mutation, KAB activation as default MAR path, and replacing required team color review gates.
