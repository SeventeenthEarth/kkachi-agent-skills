# MAR Premium Escalation Guide

Codex and Claude are premium MAR reviewers. They require explicit 주군 approval
before use unless the active task contract already grants that exact
escalation.

There is no automatic premium fallback. If premium review is suggested but
approval is absent, MAR must fail closed as `DEGRADED`, `BLOCKED`, or
`REQUEST_CHANGES` according to the Blue disposition and Red adjudication
trigger state.

Premium use does not erase default reviewer coverage failures, provider
unavailability, parse failures, or missing evidence.
