# Kkachi Multi-Agent Review policy

Date: 2026-06-17
Owner: KAS policy and skill layer
Status: candidate repository promotion SOT; local skill/script surfaces implemented for fixture/mock/read-only paths; MAR-004 provider-run implementation contract prepared
Authority level: KAS planning authority for MAR promotion from accepted Obsidian output SOT; not installed skill behavior or runtime activation
Source SOT: `/Users/draccoon/Workspace/Hermes/17thHermes/40_outputs/projects/kkachi/2026-06-16-kkachi-multi-agent-review-mar-sot.md`
Paired KAH planning SOT: KAH `docs/sot/multi-agent-review-evidence-gates.md`
Scope: KAS MAR policy, skill, prompt templates, reviewer matrix, premium escalation rules, and script ownership

## Purpose

Kkachi Multi-Agent Review (MAR) is the planned lightweight KAS-local review lane for routine Kkachi development review. It collects specialized AI reviewer outputs, preserves full evidence through KAH artifacts, and returns a compact merge pack plus Blue disposition path. MAR is advisory evidence; Blue disposition and conditional Red adjudication remain the authority.

This repository document promotes the accepted output SOT into KAS planning authority. Current evidence covers the `kkachi-multi-agent-review` scaffold and local `scripts/mar.py` fixture/mock/read-only surfaces only. It does not claim that provider adapters, KAH gates, or installed runtime behavior exist until the follow-up implementation tasks complete with evidence.

## Canonical operating rule

```text
Blue runs MAR by default.
KAS defines the skill and script.
KAH preserves evidence.
Red adjudicates only when risk, conflict, or workflow gates require it.
Codex and Claude require explicit 주군 approval unless pre-authorized.
zcode/glm-5.2, Kimi K2.7, and Antigravity/Gemini are the default reviewer set once adapter proof exists.
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
| `MAR-004` | KAS | Implement provider run path after adapter proof, authorization boundaries, default-attempt coverage, provider-failure semantics, retry/alternate/waiver decision paths, and dogfood evidence are recorded. | Default reviewer execution is available only for validated providers; failed coverage cannot become clean PASS without same-provider retry success, approved alternate success, or explicit 주군 waiver evidence. |
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
| `kimi_k2_7` | `kimi` to `k2.7` | Requirement, artifact, and traceability review | Unavailable Kimi creates degraded coverage; it must not become clean PASS. |
| `antigravity_gemini` | `agy` to selected Gemini model | Architecture, integration, security, and operational-risk review | Flash-class model is only sufficient for quick low-risk smoke; high-risk final MAR needs Pro-class or recorded policy. |

Codex and Claude are premium reviewers. They require explicit 주군 approval unless the active task contract already grants that escalation.

## MAR-004 provider-run contract

MAR-004 must make provider execution evidence-driven and fail-closed. It may not
claim default MAR coverage until all required default reviewers have an attempt
record or a pre-scoped narrower reviewer set is approved before execution.

Required default reviewers for MAR-004 are:

1. `zcode_glm_5_2` using `zcode` with `glm-5.2`;
2. `kimi_k2_7` using `kimi` with `k2.7`;
3. `antigravity_gemini` using Antigravity/Gemini with the selected model
   recorded in evidence.

For each default reviewer, KAS must preserve:

- provider id, command lane, selected model, started/ended timestamps, timeout,
  exit code, and mutation-check result;
- redacted command/preflight evidence sufficient to prove the selected lane and
  model without exposing secrets;
- raw output path, parsed finding path, parser status, and capped-output note;
- terminal status from the MAR vocabulary;
- provider failure reason when status is `DEGRADED`, `BLOCKED`, or `FAILED`.

The KAS provider lane source is `registries/mar-provider-lanes.json`
(`schema_version: mar.provider_lanes.v1`). `scripts/mar.py provider-lanes`
is the stdlib readback surface for default reviewer ids, selected models,
prompt templates, validation posture, and the provider-failure reason
vocabulary. Registry entries may exist with `validated: false`; such entries
are not successful live coverage until preflight/adapter proof and successful
provider-attempt evidence exist.

Host-specific executable resolution belongs in the existing project toolchain
state, not in the portable registry. MAR provider proof may extend
`.kkachi/toolchain.yaml` with `mar_provider_tools` (`schema_version:
mar.provider_tools.v1`) containing non-secret `resolved_argv`, selected model,
version, validation status, and proof-evidence references for reviewer ids. KAS
merges the portable registry with this toolchain overlay at provider readback,
preflight, and attempt time. The overlay may resolve user-interactive aliases or
PATH-only commands to explicit argv arrays, but it must not store auth tokens,
session files, provider cookies, or gateway credentials. Per-run artifacts must
still snapshot the resolved proof used for that run.

Supported provider-failure reason codes are:

```text
auth_failed
token_exhausted
quota_exhausted
rate_limited
cli_missing
model_unavailable
timeout
nonzero_exit
parse_failure
mutation_detected
unknown_provider_failure
```

Default policy is attempt-all-first: zcode/glm-5.2, kimi/k2.7, and
Antigravity/Gemini must all be attempted before Blue disposition unless the task
contract pre-scopes a narrower set. When one or more default reviewers fail,
KAS must not silently substitute another provider and must not report clean
`PASS`. Blue must record one of these explicit paths for each failed default
reviewer:

- same-provider retry succeeds, with linkage to the failed attempt;
- 주군 approves a named alternate provider for that reviewer slot and the
  approved alternate provider attempt succeeds;
- 주군 grants explicit waiver for the missing reviewer coverage;
- external condition remains unresolved, so the MAR result is `BLOCKED`,
  `DEGRADED`, or `FAILED` as appropriate.

KAS owns provider execution, parsing, merge-pack creation, status aggregation,
and Blue disposition. KAH owns deterministic evidence/gate validation after the
paired MAREV implementation exists. KAH must not choose retries, alternates, or
waivers, and KAS must not claim that KAH has validated provider attempts until
MAREV code/test evidence exists.

MAR-004 dogfood evidence must run beside the still-active review workflow before
broad wording claims MAR has replaced legacy GLM Octo or team color review
requirements. Dogfood evidence must include at least one representative KAS/KAH
diff, provider-attempt artifacts, compact merge pack, Blue disposition, and Red
adjudication when any trigger fires.

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

1. `PASS` requires complete default coverage or a pre-scoped narrower reviewer set recorded before execution, resolved failed-coverage paths when applicable, and no actionable findings.
2. `DEGRADED` is not a soft pass. It requires a Blue reason and may require Red adjudication.
3. All-provider failure is `FAILED` or `BLOCKED`, never clean review. Partial default failure is `DEGRADED`, `BLOCKED`, or `FAILED` until retry, approved alternate, or explicit waiver evidence resolves the missing coverage.
4. One successful default reviewer on nontrivial development is insufficient coverage and requires Red adjudication before final Blue disposition.
5. Provider disagreement creates a disposition obligation, not a vote.
6. Premium review does not erase default coverage failures.

## Red adjudication triggers

Red adjudication is required when any of the following are true:

- blocker finding exists;
- two or more high-severity findings exist;
- providers disagree on a high/blocker issue;
- findings involve SOT, approval, fallback, runtime, auth, secret, security, architecture boundary, or KAS/KAH/KAB responsibility;
- all default reviewers fail or any default reviewer failure remains unresolved;
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

Deferred unless separately approved: automatic PR inline comments, automatic code mutation, automatic test/build execution by reviewer models, model voting as authority, silent premium-provider fallback, automatic alternate-provider substitution, profile/provider/gateway/auth/token/model mutation, KAB activation as default MAR path, and replacing required team color review gates.
