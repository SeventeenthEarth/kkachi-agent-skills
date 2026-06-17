# MAR Reviewer Role Matrix

MAR-002 defines reviewer policy and prompt scaffolding only. MAR-003 adds local
fixture/mock/read-only script surfaces. MAR-004 must prove provider availability
and attempt evidence before any default reviewer coverage claim.

| Reviewer id | Default role | Premium status |
|---|---|---|
| `zcode_glm_5_2` | Kkachi SOT, fail-closed, approval, fallback-policy, and evidence-risk review. | Default-eligible only after adapter proof. |
| `kimi_k2_7` | Requirement, artifact, and traceability review. | Default-eligible only after adapter proof; unavailable Kimi is failed coverage unless retry, approved alternate, or explicit waiver evidence resolves it. |
| `antigravity_gemini` | Architecture, integration, security, and operational-risk review. | Default-eligible only after adapter proof. |
| `premium_approval_required` | Codex or Claude premium review when explicitly approved. | Approval-gated; never automatic fallback. |

Dispatch success and provider availability are not completion evidence. MAR
completion requires preserved reviewer output, parsed findings, coverage
state, provider-failure resolution where applicable, Blue disposition, and any
required Red adjudication handoff. Automatic alternate-provider fallback is
forbidden.
