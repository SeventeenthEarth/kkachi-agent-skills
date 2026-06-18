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
`mar.py` fixture/mock/read-only script surface. MAR-004 owns provider execution
safety, failure reason codes, raw-output caps, mutation guards, and adapter-proof
blocking after provider proof. MAR-005 owns role-first required coverage:
`logic`, `security`, `arch`, `cve`, and `test_adequacy` are required roles, and
each role has declared primary and secondary provider candidates in
`registries/mar-provider-lanes.json`. MAREV-002 owns later deterministic KAH MAR
artifact, gate, or schema validation behavior.

## Boundaries

- Do not execute reviewers or providers from this skill until MAR-004 provider-run implementation evidence exists.
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
