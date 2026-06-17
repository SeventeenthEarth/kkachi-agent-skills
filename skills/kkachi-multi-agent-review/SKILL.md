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
`mar.py` fixture/mock/read-only script surface, MAR-004 owns provider execution
after adapter proof, and MAREV-002 owns deterministic KAH MAR artifact, gate,
or schema behavior.

## Boundaries

- Do not execute reviewers or providers from this skill.
- Do not treat provider dispatch success, provider availability, or rendered
  prompt creation as completion evidence.
- Do not activate KAB as a default MAR path.
- Do not mutate auth, token, provider, gateway, profile, model, or live runtime
  settings.
- Codex/Claude premium reviewers require explicit approval.

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
reviewer disagreement on high or blocker issues, degraded reviewer coverage,
premium escalation suggestions, low Blue confidence, and KAS/KAH/KAB boundary
or security-sensitive findings.

## References and Templates

- `references/reviewer-role-matrix.md`
- `references/premium-escalation-guide.md`
- `templates/prompts/mar/zcode-glm-5-2-reviewer-request.md.tmpl`
- `templates/prompts/mar/kimi-k2-6-reviewer-request.md.tmpl`
- `templates/prompts/mar/antigravity-gemini-reviewer-request.md.tmpl`
- `templates/prompts/mar/premium-reviewer-request.md.tmpl`
- `templates/run-artifacts/mar-blue-disposition.md.tmpl`
- `templates/run-artifacts/mar-red-adjudication-handoff.md.tmpl`
