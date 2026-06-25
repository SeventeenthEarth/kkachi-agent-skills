---
name: kas-overlay-doctor-guide
description: Interpret KAS wrapper and overlay diagnostic findings without executing SKILL-004 doctor behavior.
---

# KAS Overlay Doctor Guide

Use this guide to understand wrapper and overlay diagnostic categories during
SKILL-003. This guide is non-executing guidance only. SKILL-004 owns the
implemented read-only doctor diagnostics, command behavior, and machine output.

## Non-Executing Diagnostic Examples

Example: missing plugin base.

```text
condition: required plugin base skill missing
severity: error
operator action: stop, report `kkachi-agent-skills:<base>` as missing, do not
fall back to a profile-local copied skill, request review/approval.
```

Example: invalid overlay target.

```text
condition: applies_to contains `kkachi-plan` instead of `kkachi-agent-skills:plan`
severity: error
operator action: stop, report `applies_to`, do not rewrite the overlay
silently, request review/approval.
```

Example: copied base suite present.

```text
condition: profile-local full KAS base copy present
severity: warning during migration, error after cutover
operator action: classify as legacy/shadow risk only; do not delete, adopt, or
use it as fallback without an approved migration task.
```

Example: stale base version.

```text
condition: overlay base_version does not match current KAS plugin base
severity: warning
operator action: request review before using the overlay for current work.
```

Example: reusable local improvement.

```text
condition: overlay improvement appears useful beyond one project
severity: info
operator action: record a promotion candidate; do not promote it into the KAS
base without review.
```

## Boundary

Do not claim SKILL-004 doctor implementation from this guide. Do not mutate live
profiles, install wrappers, clean copied suites, write KAH state, activate KAB,
or change auth/token/gateway/provider/model configuration.
