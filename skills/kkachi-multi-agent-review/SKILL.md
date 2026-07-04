---
name: kkachi-multi-agent-review
description: MAR request/render/trigger facade for KAS/Kkachi.
version: 0.2.0
---

# Kkachi Multi-Agent Review

Use this skill only when KAS/Kkachi explicitly requests MAR request-bundle,
prompt, trigger, or Blue disposition support. This is the current Multi-agent review
facade skill for KAS.

MAR-002 does not execute providers; that no-provider boundary remains preserved.

KAS is now the MAR contract and trigger facade. It defines the role matrix,
provider-policy declarations, request-bundle refs/checksums, prompt/input refs,
retry/waiver/escalation policy, and Blue disposition template. It may render a
KAH `mar start` command through `scripts/mar.py kah-trigger`, but it must not
own healthy provider execution. Reviewed provider process control, async role
attempts, primary/secondary queueing, provider HOME injection, raw-output
capture, merge/status evidence, and KAH gate artifacts belong to KAH `mar`.

MAR-002/MAR-003 fixture/mock/read-only surfaces and MAR-004..006 KAS-runner
records are historical/pre-NEWMAR evidence only. `docs/sot/mar-execution-realignment.md`
supersedes the long-term provider-execution ownership line. MAR role coverage
still requires `logic`, `security`, `arch`, `cve`, and `test_adequacy`, with
primary/secondary provider candidates declared in `registries/mar-provider-lanes.json`,
but KAS declares those lanes; KAH executes and records them after reviewed
capability/approval evidence exists.

reviewed NEWMAR-002 request bundle and provider-registry/correlation evidence is candidate evidence for NEWMAR-003+ only. KAH must fail closed on missing, stale, mismatched, unsupported, expired, unsafe-ref, extra-unreviewed, or checksum-drifting metadata before provider execution. The reviewed schema contract uses strict ref objects, safe repo-relative paths only, `sha256:<64 lowercase hex>` checksum fields, and complete `approval_binding.mar_start.bound_tuple` / `mar_start_approval_binding.bound_tuple` records for the minimum approval-binding tuple: request-bundle ref/checksum, prompt/input refs/checksums, provider family, adapter proof, provenance, author-backend correlation refs, provider preflight timestamps/checked versions, execution-policy values, approval scope/checksum/times, retry/waiver refs, and explicit null/validation-only posture.
KAH NEWMAR-003+ may consume only reviewed NEWMAR-002 schema/capability evidence, never arbitrary/unreviewed/generated metadata. The top-level author backend correlation is a summary view and the bound-tuple author backend correlation is the exact approval binding; overlapping backend family/identity values must match. If approval metadata is stale, mismatched, expired, or checksum-drifting, regenerate the affected packet/ref, rerun review, bind a new exact approval tuple, or stay validation-only; never fall back to default provider execution.

## Boundaries

- Use KAS MAR surfaces for request/render/trigger/readback/disposition only.
- Do not execute reviewers or providers from this skill or from KAS `scripts/mar.py`
  as the future healthy path; KAH `mar` owns reviewed provider execution.
- Treat KAS `provider-attempt` as a deprecated compatibility facade that must
  fail closed with KAH execution required, not as role coverage.
- Do not treat provider dispatch success, provider availability, rendered
  prompt creation, or unresolved required role coverage as completion evidence.
- Do not activate KAB as a default MAR path.
- Do not mutate auth, token, provider, gateway, profile, model, or live runtime
  settings.
- Codex/Claude premium reviewers require explicit approval.
- Required role coverage is declared in KAS and attempted by KAH only through
  the declared primary provider and then the declared secondary provider for
  that same role. If both fail, report the unresolved role to 주군/operator and
  fail closed; do not silently try undeclared tertiary, premium, alternate, or
  waiver-as-clean coverage.

## Current KAS Procedure

1. Decide that MAR is required from the task contract and workflow gate.
2. Render or validate request-bundle, prompt/input refs, provider registry refs,
   approval binding, retry/waiver/escalation metadata, and role matrix; refs must
   be safe repo-relative paths, not absolute or parent-traversal paths.
3. Read effective KAH `mar` capability/status evidence before presenting the
   trigger as usable.
4. Render the KAH command with `scripts/mar.py kah-trigger --request <request>`
   or the equivalent KAH wrapper; unsafe request refs fail closed and provider
   CLIs must not run from KAS.
5. Read KAH `mar status` / gate artifacts and write Blue disposition from those
   evidence refs. If KAH evidence is missing, stale, degraded, or inconsistent,
   hold fail-closed and route Red adjudication when required.

## Status Semantics

MAR disposition is fail-closed. Supported terminal statuses are:

- `PASS`
- `PASS_WITH_FINDINGS`
- `REQUEST_CHANGES`
- `BLOCKED`
- `DEGRADED`
- `FAILED`

`DEGRADED`, `FAILED`, and `BLOCKED` are never clean review completion claims.

## Red Adjudication

Red adjudication triggers include blocker findings, high-risk findings,
reviewer disagreement on high or blocker issues, unresolved required role
coverage, premium escalation suggestions, low Blue confidence, and KAS/KAH/KAB
boundary or security-sensitive findings.

## References and Templates

- `references/reviewer-role-matrix.md`
- `references/premium-escalation-guide.md`
- `templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl`
- `templates/prompts/mar/kimi-default-reviewer-request.md.tmpl`
- `templates/prompts/mar/antigravity-gemini-reviewer-request.md.tmpl`
- `templates/prompts/mar/premium-reviewer-request.md.tmpl`
- `templates/run-artifacts/mar-blue-disposition.md.tmpl`
- `templates/run-artifacts/mar-red-adjudication-handoff.md.tmpl`
