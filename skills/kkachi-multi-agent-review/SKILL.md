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
`mar.py` fixture/mock/read-only script surface. MAR-004 owns provider execution,
attempt-all-first coverage, failure reason codes, retry/alternate/waiver decision
paths, merge-pack evidence, and dogfood evidence after adapter proof. MAR-004
also uses the existing `.kkachi/toolchain.yaml` `mar_provider_tools` section as
the local non-secret provider proof overlay for host-specific executable argv,
version, selected model, and validation state; portable defaults remain in
`registries/mar-provider-lanes.json`. MAREV-002 owns deterministic KAH MAR
artifact, gate, or schema behavior.

## Boundaries

- Do not execute reviewers or providers from this skill until MAR-004 provider-run implementation evidence exists.
- Do not treat provider dispatch success, provider availability, rendered
  prompt creation, or unresolved failed reviewer coverage as completion evidence.
- Do not activate KAB as a default MAR path.
- Do not mutate auth, token, provider, gateway, profile, model, or live runtime
  settings.
- Codex/Claude premium reviewers require explicit approval.
- Default reviewer failure requires same-provider retry, 주군-approved alternate,
  explicit 주군 waiver, or non-clean `DEGRADED`/`BLOCKED`/`FAILED` disposition;
  automatic fallback is forbidden.

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
- `templates/prompts/mar/kimi-k2-7-reviewer-request.md.tmpl`
- `templates/prompts/mar/antigravity-gemini-reviewer-request.md.tmpl`
- `templates/prompts/mar/premium-reviewer-request.md.tmpl`
- `templates/run-artifacts/mar-blue-disposition.md.tmpl`
- `templates/run-artifacts/mar-red-adjudication-handoff.md.tmpl`
