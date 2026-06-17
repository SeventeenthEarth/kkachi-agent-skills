# MAR Reviewer Role Matrix

MAR-002 defines reviewer policy and prompt scaffolding only. It does not prove
provider availability, and it does not execute reviewer providers.

| Reviewer id | Default role | Premium status |
|---|---|---|
| `zcode_glm_5_2` | Kkachi SOT, fail-closed, approval, fallback-policy, and evidence-risk review. | Default-eligible only after adapter proof. |
| `kimi_k2_6` | Requirement, artifact, and traceability review. | Default-eligible only after adapter proof. |
| `antigravity_gemini` | Architecture, integration, security, and operational-risk review. | Default-eligible only after adapter proof. |
| `premium_approval_required` | Codex or Claude premium review when explicitly approved. | Approval-gated; never automatic fallback. |

Dispatch success and provider availability are not completion evidence. MAR
completion requires preserved reviewer output, parsed findings, coverage
state, Blue disposition, and any required Red adjudication handoff.
