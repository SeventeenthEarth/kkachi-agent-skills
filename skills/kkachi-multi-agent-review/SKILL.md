---
name: kkachi-multi-agent-review
description: Multi-agent review orchestration/scaffold for KAS/Kkachi.
version: 0.1.0
---

# Kkachi Multi-Agent Review

Use this skill only when KAS/Kkachi explicitly requests Multi-agent review
scaffold, prompt, or disposition support for MAR policy work.

MAR-002 does not execute providers. This skill provides only scaffold, reviewer
prompt, reference, and disposition-template support. MAR-003 adds the local
`mar.py` fixture/mock/read-only script surface. Historical MAR-004..006 evidence
remains valid for the pre-NEWMAR KAS-runner path, but `docs/sot/mar-execution-realignment.md`
supersedes the long-term provider-execution ownership line: future healthy MAR
provider execution moves to KAH `mar` after KAS request bundles, KAH
validation/status/wait/cancel, fake/no-provider async proof, real adapter proof,
and reviewed live-pilot approval. MAR-005 owns role-first required coverage:
`logic`, `security`, `arch`, `cve`, and `test_adequacy` are required roles, and
each role has declared primary and secondary provider candidates in
`registries/mar-provider-lanes.json`. Current MAREV-002/KAH behavior is
`mar-evidence.v1` artifact/gate/schema validation only until NEWMAR
implementation evidence exists.

reviewed NEWMAR-002 request bundle and provider-registry/correlation evidence is candidate evidence for NEWMAR-003+ only. KAH must fail closed on missing, stale, mismatched, unsupported, expired, unsafe-ref, extra-unreviewed, or checksum-drifting metadata before provider execution. The reviewed schema contract uses strict ref objects, safe repo-relative paths only, `sha256:<64 lowercase hex>` checksum fields, and complete `approval_binding.mar_start.bound_tuple` / `mar_start_approval_binding.bound_tuple` records for the minimum approval-binding tuple: request-bundle ref/checksum, prompt/input refs/checksums, provider family, adapter proof, provenance, author-backend correlation refs, provider preflight timestamps/checked versions, execution-policy values, approval scope/checksum/times, retry/waiver refs, and explicit null/validation-only posture.
KAH NEWMAR-003+ may consume only reviewed NEWMAR-002 schema/capability evidence, never arbitrary/unreviewed/generated metadata. The top-level author backend correlation is a summary view and the bound-tuple author backend correlation is the exact approval binding; overlapping backend family/identity values must match. If approval metadata is stale, mismatched, expired, or checksum-drifting, regenerate the affected packet/ref, rerun review, bind a new exact approval tuple, or stay validation-only; never fall back to default provider execution.

## Boundaries

- Do not execute reviewers or providers from this skill as the future healthy path; after NEWMAR, KAS should render/request/facade while KAH `mar` owns reviewed provider execution. Historical MAR-004 provider-run evidence remains pre-NEWMAR only.
- Do not treat provider dispatch success, provider availability, rendered
  prompt creation, or unresolved required role coverage as completion evidence.
- Do not activate KAB as a default MAR path.
- Do not mutate auth, token, provider, gateway, profile, model, or live runtime
  settings.
- Codex/Claude premium reviewers require explicit approval.
- Required role coverage tries only the declared primary provider and then the
  declared secondary provider for that same role. If both fail, report the
  unresolved role to 주군/operator and fail closed; do not silently try
  undeclared tertiary, premium, alternate, or waiver-as-clean coverage.

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
