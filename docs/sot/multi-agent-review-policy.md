# Kkachi Multi-Agent Review policy

Date: 2026-06-17
Owner: KAS policy and skill layer
Status: candidate repository promotion SOT; pre-v0.2.1 legacy Python MAR is OFF/HOLD in current v0.2.1 source; MAR-006 KAH CLI binary selection implemented for KAS doctor/workflow surfaces
Authority level: KAS planning authority for MAR promotion from accepted Obsidian output SOT; not installed skill behavior or runtime activation
Source SOT: `/Users/draccoon/Workspace/Hermes/17thHermes/40_outputs/projects/kkachi/2026-06-16-kkachi-multi-agent-review-mar-sot.md`
Paired KAH planning SOT: KAH `docs/sot/multi-agent-review-evidence-gates.md`
Scope: KAS MAR policy, skill, prompt templates, reviewer matrix, premium escalation rules, request-bundle/provider-policy declarations, and KAH execution boundary
NEWMAR marker: `docs/sot/mar-execution-realignment.md` supersedes this document for provider-execution ownership. Current v0.2.1 KAS source keeps provider policy and request/disposition guidance only; pre-v0.2.1 legacy Python MAR is OFF/HOLD until reviewed Go/KAH MAR lands. Historical `MAR-004..006` evidence remains lineage, not an active runtime surface.

## Purpose

Kkachi Multi-Agent Review (MAR) is the planned lightweight KAS-local review lane for routine Kkachi development review. It collects specialized AI reviewer outputs, preserves full evidence through KAH artifacts, and returns a compact merge pack plus Blue disposition path. MAR is advisory evidence; Blue disposition and conditional Red adjudication remain the authority.

This repository document promotes the accepted output SOT into KAS planning authority. Current source evidence covers the `kkachi-multi-agent-review` scaffold, provider-policy registry, prompt templates, request/disposition guidance, and KAH execution boundary. KAS no longer ships the pre-v0.2.1 legacy Python MAR script or shell adapters; KAH `mar` owns reviewed provider execution and evidence once the Go/KAH MAR surface lands with evidence.

## Canonical operating rule

```text
Blue runs MAR by default.
KAS defines the skill, request contract, and provider policy.
KAH preserves evidence.
Red adjudicates only when risk, conflict, or workflow gates require it.
Codex and Claude require explicit 주군 approval unless pre-authorized.
Required roles are logic, security, arch, cve, and test_adequacy; providers are primary/secondary candidates for those roles.
Model consensus is advisory; Blue/Red disposition is authority.
```

## POLPR-003 official evidence boundary

For active KAS/KAH source policy, workflow, template, test, and shared skill
mirror work, MAR is the only independent implementation review lane unless
주군 explicitly waives or replaces it before start and the decision is recorded
in KAH/run evidence artifacts. KAS/KAH artifacts must not promote prior
Octo-style review wording as a default, optional, fallback, or legacy
independent-review path.

`delegate_task`, temporary subagents, and other ad hoc advisor surfaces may be
used only as pre-review analysis. They do not count as official Red/Orange/Gray
color review, project-Gray documentation/integrity review, MAR role coverage,
or KAH review evidence. Official evidence requires review cards/artifacts,
bounded raw outputs where applicable, parsed findings, Blue disposition, and
the recorded verdict path.

## Task split and promotion boundary

MAR promotion is intentionally split so planning documents do not overclaim implementation.

| Task | Repository | Scope | Completion claim allowed |
|---|---|---|---|
| `MAR-001` | KAS | Promote MAR policy/SOT, roadmap/docs-map/docs-index records, stale legacy-review marker targets, and implementation task boundaries. | KAS repository has a planning SOT for MAR. No skill/script/provider behavior claim. |
| `MAREV-001` | KAH | Promote KAH-side MAR artifact/gate planning SOT and roadmap/docs-index records. | KAH has a planning SOT for MAR evidence capture. No helper gate/code behavior claim. |
| `MAR-002` | KAS | Add `kkachi-multi-agent-review` skill scaffold, references, prompt templates, and disposition templates without provider execution. | KAS skill scaffold exists; provider execution still pending unless separately implemented. |
| `MAR-003` | KAS | Implement stdlib `mar.py` doctor/render/validate/merge-pack MVP using fixture and read-only local evidence. | Local MAR script surfaces exist for non-provider or mocked/fixture paths. |
| `MAR-004` | KAS | Implement provider run path after adapter proof, authorization boundaries, provider-failure semantics, mutation guards, toolchain overlay proof, and dogfood evidence are recorded. | Provider execution is available only for validated providers; provider failure cannot become clean coverage by availability, prompt rendering, or dispatch success alone. |
| `MAR-005` | KAS | Convert coverage from provider-first reviewer slots to role-first required coverage with primary/secondary provider candidates. | Required roles (`logic`, `security`, `arch`, `cve`, `test_adequacy`) are covered only by their declared primary/secondary provider path; unresolved required roles fail closed with 주군/operator report wording. |
| `MAR-006` | KAS | Select the `kkachi-agent-helper` binary used by KAS doctor/workflow KAH probes through `KKACHI_KAH_BIN`, `.kkachi/toolchain.yaml`, repo `.kkachi/bin`, then ambient PATH. | KAS owns binary selection for its own CLI surfaces only; explicit repo KAH selection mismatch fails closed and refuses PATH. Resolver success is not KAH capability, evidence, project-state, or gate success. |
| `MAREV-002` | KAH | Implement deterministic KAH artifact/gate/schema support for KAS-declared MAR role coverage evidence if needed by final gates. | KAH helper behavior exists only after code/tests/docs/release evidence; KAH does not choose roles or providers. |

`MAREV-001` is docs/SOT only. It must not be read as implementing KAH behavior. Actual KAH code or gate behavior belongs in `MAREV-002` or later.

## Ownership

KAS owns:

- `kkachi-multi-agent-review` skill definition;
- MAR policy and reviewer selection rules;
- prompt templates and reviewer role matrix;
- premium escalation policy;
- request-bundle, prompt-template, role/provider-policy, and Blue disposition guidance;
- pre-v0.2.1 legacy Python MAR is OFF/HOLD; no bundled script or shell adapter is active in current source;
- Blue disposition template guidance.
- selecting which `kkachi-agent-helper` binary its own doctor,
  workflow-create, workflow-promote, workflow-trigger, and graphsync /
  workflow-graph repair surfaces execute when invoking KAH capabilities.

KAH owns deterministic evidence and process control after its paired tasks implement or document support:

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
- helper command behavior, deterministic project state, capability output, and
  gate validation. A KAS resolver selecting a binary does not prove KAH command
  success unless the effective KAH command returns the required evidence.

KAB remains non-default for routine MAR. KAB is used only when a later task explicitly selects KAB runtime orchestration or premium/heavier review support.

## Role-first reviewer coverage

MAR-005 makes review role coverage the default MAR completion unit. Providers are
execution candidates attached to role ids; KAS must not treat provider ids as
role ids or report provider availability as clean review coverage.

Required MAR roles are:

1. `logic` — requirement logic, acceptance-criteria consistency, and
   fail-closed reasoning review;
2. `security` — security-sensitive behavior, auth/secret boundaries, unsafe
   mutation, and fail-closed risk review;
3. `arch` — generic architecture, integration, API/schema, modularity, and
   ownership-boundary review;
4. `cve` — known-vulnerability, dependency-risk, exploitability, and advisory
   relevance review from supplied evidence;
5. `test_adequacy` — test and verification adequacy against the task contract
   and acceptance criteria.

Each required role has declared provider-candidate policy in
`registries/mar-provider-lanes.json` (`schema_version: mar.role_lanes.v1`). The
registry is a portable KAS policy readback source for required roles,
role-to-provider candidates, provider metadata, validation posture, and
provider-failure reason vocabulary. It is role-first and declares
`execution_owner: KAH`; it must not carry KAS-owned executable, shell-adapter,
or argv fields.

Current default policy keeps `zcode_glm_5_2` as primary and `kimi_default` as
secondary for every required role, but KAS does not define the runtime argv for
those providers. Kimi default/latest selection, ZCode invocation, Antigravity
health/model selection, provider preflight, raw-output capture, parsing,
status vocabulary enforcement, and mutation guards belong to reviewed KAH `mar`
execution and its evidence artifacts. `antigravity_gemini` remains explicit
non-default provider metadata and is not required for clean coverage until
KAH-owned selected-model and health evidence are fixed.

Pre-v0.2.1 legacy Python MAR is OFF/HOLD and must not be used as role coverage
or as the healthy execution path.

## Historical MAR-004 provider-run contract

MAR-004 made provider execution evidence-driven and fail-closed for the
pre-NEWMAR KAS-runner path. NEWMAR-008 keeps these fields as compatibility and
migration evidence only; healthy provider execution/control now belongs to KAH
`mar`. Historical provider attempt
records must preserve:

- role id, provider id, provider candidate (`primary` or `secondary`), command
  lane, selected model or explicit default/latest model-selection mode,
  started/ended timestamps, timeout, exit code, and mutation-check result;
- redacted command/preflight evidence sufficient to prove the selected lane and
  explicit model or default/latest model-selection posture without exposing
  secrets;
- raw output path, parsed finding path, parser status, and capped-output note;
- terminal status from the MAR vocabulary;
- provider failure reason when status is `DEGRADED`, `BLOCKED`, or `FAILED`.

Host-specific executable resolution belongs in the existing project toolchain
state, not in the portable registry. MAR provider proof may extend
`.kkachi/toolchain.yaml` with `mar_provider_tools` (`schema_version:
mar.provider_tools.v1`) containing non-secret `resolved_argv`, selected model,
version, validation status, and proof-evidence references for provider ids. KAS
merges the portable registry with this toolchain overlay at readback, preflight,
and attempt time. The overlay may resolve user-interactive aliases or PATH-only
commands to explicit argv arrays, but it must not store auth tokens, session
files, provider cookies, or gateway credentials. Per-run artifacts must still
snapshot the resolved proof used for that run.

`mar_provider_tools.resolved_argv` is authority only for MAR provider command
execution. It must not be used as KAH CLI selection authority. KAS KAH invoker
surfaces select `kkachi-agent-helper` through `KKACHI_KAH_BIN`, top-level
`kah_cli_path`/`kah_cli` in `.kkachi/toolchain.yaml`, repo
`.kkachi/bin/kkachi-agent-helper`, and finally ambient PATH only when no
explicit repo KAH selection exists.

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
adapter_proof_required
unknown_provider_failure
```

## MAR-005 role-first run contract

MAR-005 attempts required roles, not default reviewer slots. For each required
role, KAS attempts the declared primary provider first. If the primary provider
does not produce a covered status (`PASS` or `PASS_WITH_FINDINGS`), KAS may
attempt only that role's declared secondary provider. If both primary and
secondary fail for a required role, KAS must emit `unresolved_required_roles`,
include explicit 주군/operator report wording, and fail closed. KAS must not
silently try undeclared tertiary, premium, alternate, or waiver-as-clean coverage
for that role.

Role reviewers return role-scoped findings and role-scoped
acceptance-criteria verdicts only. Reviewers must not issue global pass/fail,
final Blue disposition, model-voting authority, or KAH gate claims. Blue owns the
final disposition and matrix synthesis; Red adjudication is triggered by
unresolved required role coverage, conflict, low confidence,
high/blocker/security/fail-closed findings, premium escalation suggestions, or
explicit task policy.

Pre-NEWMAR, KAS owns provider execution, parsing, merge-pack creation,
status aggregation, role coverage, Blue matrix inputs, and Blue disposition.
For future healthy MAR execution, `docs/sot/mar-execution-realignment.md`
supersedes only the provider-execution ownership line: KAS will own request
bundles/policy/disposition, while KAH will own deterministic MAR execution and
evidence after reviewed NEWMAR implementation. KAH must not choose roles,
providers, retries, alternates, or waivers, and KAS must not claim that KAH has
validated role coverage or executed providers until corresponding KAH code/test
evidence exists.

MAR dogfood evidence must run beside the still-active review workflow before
broad wording claims MAR has replaced legacy review or team color review
requirements. Dogfood evidence must include at least one representative KAS/KAH
diff, KAH MAR role-attempt artifacts, compact merge pack, Blue disposition, and
Red adjudication when any trigger fires.

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

1. `PASS` requires complete required-role coverage or a pre-scoped narrower role set recorded before execution, resolved failed-role paths when applicable, and no actionable findings.
2. `DEGRADED` is not a soft pass. It requires a Blue reason and may require Red adjudication.
3. All required roles unresolved is `FAILED` or `BLOCKED`, never clean review. Any unresolved required role is `DEGRADED`, `BLOCKED`, or `FAILED` until the declared primary/secondary path produces covered evidence or the task is explicitly re-scoped before execution.
4. One successful provider on nontrivial development is insufficient unless all required roles are covered; unresolved role coverage requires Red adjudication before final Blue disposition.
5. Provider disagreement creates a disposition obligation, not a vote.
6. Premium review does not erase required role coverage failures.

## Red adjudication triggers

Red adjudication is required when any of the following are true:

- blocker finding exists;
- two or more high-severity findings exist;
- role reviewers or providers disagree on a high/blocker issue;
- findings involve SOT, approval, fallback, runtime, auth, secret, security, architecture boundary, or KAS/KAH/KAB responsibility;
- all required roles remain unresolved or any required role remains unresolved;
- Blue confidence is below 0.75;
- premium escalation is suggested;
- task class is release readiness, architecture SOT change, or workflow policy change.

## Safety rules

MAR is read-only. Reviewer prompts and scripts must prohibit reviewer models from editing files, running tests, running builds, installing packages, starting services, probing networks, changing auth, changing secrets, changing provider settings, or mutating runtime state.

The KAS script must use `subprocess.run([...], shell=False)`, enforce provider timeouts, cap raw output size, preserve parse failures, compare git status before/after execution, fail closed on mutation detection, record unavailable providers as degraded coverage, and never interpret provider failure as clean review.

## Legacy review handling

Existing KAS surfaces that require a pre-MAR default routine review lane are stale-marker targets during MAR promotion. They must not be silently deleted. Before replacement, each update must record:

- exact file path and section;
- whether the old wording is active, historical, alternate, or still required for a specific workflow;
- whether MAR dogfood evidence exists;
- whether another separate active workflow gate still requires external review.

Dogfood evidence is required before broad wording changes claim that MAR replaces legacy default routine review behavior.

## Verification before implementation claim

KAS must not claim MAR implementation until evidence exists for the specific claimed surface:

- docs-map/index/roadmap registration for the policy SOT;
- skill scaffold readback for `kkachi-multi-agent-review`;
- prompt template readback and reviewer-role matrix;
- `mar.py` help/doctor/render/validate tests;
- provider adapter proof for any provider counted as successful role coverage;
- fixture tests for degraded, insufficient, failed, and request-changes semantics;
- KAH artifact/gate evidence once MAREV implementation is claimed;
- Red/Orange/Gray review or a recorded 주군-approved exception when required by workflow risk.

## Deferrals

Deferred unless separately approved: automatic PR inline comments, automatic code mutation, automatic test/build execution by reviewer models, model voting as authority, silent premium-provider fallback, automatic alternate-provider substitution, profile/provider/gateway/auth/token/model mutation, KAB activation as default MAR path, and replacing required team color review gates.
