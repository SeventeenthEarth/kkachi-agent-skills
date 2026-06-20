# MAR task loop contract

Date: 2026-06-20
Owner: KAS policy and skill layer
Confirming role: Responsible approver / Blue command with Red, Orange, and project-Gray review evidence required for `MARTL-001`
Status: candidate KAS SOT for the `MARTL` epic; `MARTL-001` is docs/SOT and roadmap registration only and does not implement new runner, gate, provider, KAH, KAB, profile, or runtime behavior
Authority level: KAS-side planning authority for making the already-defined MAR review/fix/re-review operating loop explicit under STRICT workflow-managed KAS runs
Scope: KAS docs, roadmap, registries, skills, templates, and `scripts/mar.py` adoption guidance. No KAH deterministic behavior change, KAB runtime activation, provider/model/gateway/auth/token mutation, profile install, release, push, or automatic source mutation is authorized by this document.
Related docs: `docs/sot/multi-agent-review-policy.md`, `docs/sot/strict-workflow-execution-contract.md`, `docs/sot/task-dag-workflow-contract.md`, `docs/sot/phase-orchestration-policy.md`, `docs/roadmap.md`, KAH `docs/sot/multi-agent-review-evidence-gates.md`, KAH `docs/sot/strict-workflow-enforcement.md`
Evidence/source paths:
- 주군 direction in 17번째 지구 Discord `#kas` thread `1517487253327446176` on 2026-06-20: after STRICT completed, `MARTL` should proceed as two KAS-side PR-candidate tasks; the existing MAR review -> Blue/Red analysis -> fix -> MAR re-review loop should be treated as existing behavior, not a newly invented KAH cross-repo system.
- Existing KAS MAR baseline: `docs/sot/multi-agent-review-policy.md`, `registries/mar-provider-lanes.json`, `scripts/mar.py`, `skills/kkachi-review/SKILL.md`, `skills/kkachi-handle-feedback/SKILL.md`, and `skills/kkachi-final-verify/SKILL.md` define role-first MAR coverage, provider attempts, Blue disposition, requested-change routing, and post-review verification requirements.
- Existing KAH substrate: KAH `multi-agent-review/status.json` gate/schema support is source-side implemented by MAREV and KAH STRICT source-side support provides workflow transition/order/projection validation. MARTL does not require a KAH PR unless implementation evidence later proves the existing deterministic gate cannot validate the needed KAS evidence.

## Purpose

`MARTL` names the MAR task loop that KAS already describes operationally: run MAR, let Blue or Red analyze the role-first findings, route valid requested changes to the selected implementer lane, verify the changed surface, request refreshed MAR review, and repeat until MAR evidence is acceptable for final Blue disposition.

The goal is alignment and adoption, not invention of a new reviewer authority. MARTL makes that loop explicit in KAS SOT, roadmap, workflow registry/phase guidance, runner/adoption surfaces, and final verification evidence so STRICT workflow-managed runs do not treat MAR as a vague advisory note or a one-shot provider transcript.

## Existing behavior to preserve

MARTL starts from already-accepted MAR and STRICT rules:

1. MAR review is role-first. Required roles are `logic`, `security`, `arch`, `cve`, and `test_adequacy` unless a task-specific reviewed policy explicitly changes the role set.
2. Provider lanes are candidates attached to roles. Provider availability, prompt rendering, dispatch success, or a single provider transcript is not review completion evidence.
3. KAS owns MAR policy, role/provider selection, prompt scope, provider attempts, parsing, merge-pack creation, Blue disposition artifacts, and requested-change routing.
4. Blue owns synthesis and disposition. Red adjudication is required when MAR coverage is unresolved, findings conflict, confidence is low, high/blocker/security/fail-closed findings appear, premium escalation is proposed, or task policy requires Red.
5. KAH owns deterministic evidence and gate validation only. KAH must not choose reviewers, run providers, adjudicate findings, mutate source, approve waivers, or convert degraded coverage to warning-only pass.
6. If MAR or any later feedback round changes code, tests, build behavior, docs, SOT, roadmap, templates, registries, or durable evidence, prior review acceptance is stale for that changed surface until focused verification and required re-review evidence exist.

## Loop model

The MARTL loop is bounded and evidence-driven:

```text
review-surface snapshot
  -> MAR role attempts / role coverage
  -> merge pack
  -> Blue disposition
  -> Red adjudication when triggered
  -> if requested changes are accepted: implementer fix + verification + refreshed review surface
  -> repeat MAR on the refreshed surface
  -> terminal only when required role coverage is resolved and Blue/Red disposition has no commit-blocking MAR findings
```

A MARTL round must preserve enough evidence to reconstruct the reviewed surface and decision path:

- task id and run id;
- repository root and current git head or dirty-worktree snapshot reference;
- input bundle path and checksum;
- diff snapshot path and checksum when a diff is reviewed;
- role matrix and role coverage;
- primary/secondary provider attempts with failure reasons when applicable;
- bounded raw outputs and parsed findings;
- merge pack;
- Blue disposition;
- Red adjudication or explicit not-triggered reason;
- accepted/rejected/deferred findings disposition;
- verification evidence after any accepted fix;
- refreshed MAR evidence after any accepted fix.

## Terminal states

MARTL uses MAR status values from the existing MAR contract but interprets loop closure through Blue/Red disposition:

| MAR evidence status | MARTL posture |
|---|---|
| `PASS` | Terminal candidate when required role coverage is resolved, Blue disposition accepts the result, and no Red trigger remains unresolved. |
| `PASS_WITH_FINDINGS` | Terminal candidate only when Blue disposition records every finding as non-blocking, fixed, rejected with rationale, or explicitly deferred with accepted risk; Red adjudication is present when triggered. |
| `REQUEST_CHANGES` | Non-terminal. Blue/Red disposition must route accepted changes to the selected implementer lane, then verification and refreshed MAR are required. |
| `BLOCKED`, `DEGRADED`, or `FAILED` | Non-terminal unless 주군 explicitly records a waiver/stop decision with accepted residual risk; otherwise the run is blocked or failed closed. |

MARTL must not treat a model phrase like "approved" as authority by itself. The loop closes on artifacted role coverage plus Blue/Red disposition, not on provider wording alone.

## Strict workflow relationship

MARTL is a KAS adoption layer over existing STRICT workflow-managed execution:

- KAS route/materialize/dispatch remains governed by `STRICT`.
- KAH node start is required before running any workflow-managed MARTL node work.
- KAH node complete is allowed only after required MARTL artifacts exist for that node.
- If a MARTL round changes the reviewed surface, the previous MAR evidence is stale for final completion and a refreshed MAR round or explicit reviewed exception is required.
- The first MARTL implementation task should align KAS workflow registry, phase contracts, skills, templates, and docs so the existing MAR loop is visible to STRICT runs.

## Task split

MARTL is intentionally KAS-only unless implementation evidence proves a KAH deterministic gap.

| Task ID | Repo | Title | Status | Completion claim allowed |
|---|---|---|---|---|
| `MARTL-001` | KAS | Register MAR task-loop SOT and roadmap sequence | Completed | KAS has the MARTL SOT, roadmap rows, docs index/map registration, and explicit KAS-only scope. No runner, registry behavior, provider execution, KAH gate, install, release, or runtime activation claim. |
| `MARTL-002` | KAS | Align MARTL workflow/runner/adoption surfaces | Completed | KAS aligns workflow registry/phase guidance/skills/templates and, if needed, `scripts/mar.py` orchestration so MAR requested-change rounds are artifacted, routed, verified, and re-reviewed before final completion. KAH code remains unchanged unless a separately approved evidence-backed follow-up is opened. |

## Acceptance criteria for MARTL-001

- `docs/sot/mar-task-loop-contract.md` exists and defines the MAR task loop, boundaries, existing behavior, terminal states, and KAS-only two-task sequence.
- `docs/roadmap.md` registers the `MARTL` epic and two PR-candidate tasks.
- `docs/README.md` and `docs/kkachi-docs-map.yaml` register the SOT.
- The SOT explicitly says existing MAR review/fix/re-review behavior exists and MARTL aligns/adopts it rather than inventing a new KAH reviewer system.
- Verification includes docs readback, docs-map YAML parse, `git diff --check`, repository docs-contract/aggregate tests, and Red/Orange/Gray color review before completion/commit readiness.

## MARTL-002 implementation target

MARTL-002 should be limited to KAS alignment and adoption:

- align `registries/task-dag-workflow-registry.yaml`, `registries/phase-contracts.yaml`, and affected workflow/template guidance so the MAR loop is represented consistently with STRICT execution;
- improve active skills/templates/final verification guidance so `REQUEST_CHANGES` from MAR cannot be treated as terminal and fixed surfaces require refreshed evidence;
- add or polish `scripts/mar.py` wrapper behavior only if the existing subcommands cannot produce the required round artifacts cleanly;
- add focused docs-contract, registry, script, or e2e fixtures proving non-terminal statuses fail closed and terminal candidate states require Blue/Red disposition;
- use existing KAH `multi-agent-review/status.json` gate/schema support rather than modifying KAH unless an actual deterministic validation gap is demonstrated.

## Deferrals and non-goals

- No new KAH PR is part of MARTL by default.
- No automatic source-code mutation by MAR reviewers.
- No automatic retry until success without Blue/Red disposition and verification.
- No undeclared tertiary provider, premium provider, or waiver-as-clean coverage path.
- No KAB runtime activation, provider/model/gateway/auth/token mutation, profile skill installation, release tagging, push, or automatic rollback.
- No replacement of Blue/Red authority with model voting or provider wording.
- No broad claim that installed/runtime KAS profiles have MARTL behavior until install/effective-runtime evidence is separately provided.

## Next action

MARTL-002 is source-side complete in KAS after implementation evidence, MAR coverage, focused Red/Orange/Gray review, KAH final gate, and local source commit closure. Install, release, push, downstream runtime adoption, and effective-runtime claims remain separate approvals.
