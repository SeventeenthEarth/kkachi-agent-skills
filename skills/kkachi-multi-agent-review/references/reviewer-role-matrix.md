# MAR Reviewer Role Matrix

MAR-005 makes review roles the default coverage unit. Providers are execution
candidates attached to each role; provider ids must stay separate from role ids.

| Required role | Default primary provider | Default secondary provider | Scope |
|---|---|---|---|
| `logic` | `zcode_glm_5_2` | `kimi_default` | Requirement logic, acceptance-criteria consistency, and fail-closed reasoning review. |
| `security` | `zcode_glm_5_2` | `kimi_default` | Security-sensitive behavior, secrets/auth boundaries, unsafe mutation, and fail-closed risk review. |
| `arch` | `zcode_glm_5_2` | `kimi_default` | Generic architecture, integration, API/schema, modularity, extensibility, and ownership-boundary review. KAS/KAH/KAB checks apply only when the reviewed scope is Kkachi-specific. |
| `cve` | `zcode_glm_5_2` | `kimi_default` | Known-vulnerability, dependency-risk, exploitability, and advisory relevance review from provided evidence. |
| `test_adequacy` | `zcode_glm_5_2` | `kimi_default` | Test and verification adequacy against the task contract and acceptance criteria. |

`zcode_glm_5_2` is primary by default. `kimi_default` is the authenticated
Kimi Code CLI default/latest lane, not an explicit model alias; it becomes
primary only when there is documented role-specific evidence that the Kimi
lane is materially better for that role. No MAR-005 default role currently has
that evidence.

`antigravity_gemini` remains explicit provider metadata but is
non-default/not required for clean coverage until agy selected-model and health
evidence are fixed. It is not a silent fallback.

If primary and secondary both fail for a required role, MAR must emit
`unresolved_required_roles`, include explicit 주군/operator report wording, and
fail closed. Same-provider retry, approved alternate, waiver, premium reviewer,
or undeclared tertiary provider evidence may be recorded as risk context, but it
does not satisfy clean required role coverage for MAR-005.

Role reviewers return role-scoped findings and role-scoped acceptance-criteria
verdicts only. They must not issue global pass/fail, final Blue disposition, or
model-voting authority. Blue owns the final disposition, and Red adjudication is
triggered by unresolved role coverage, conflict, low confidence,
high/blocker/security/fail-closed findings, premium escalation suggestions, or
explicit task policy.

Premium review remains `premium_approval_required`: Codex or Claude review is
approval-gated and never automatic fallback.
