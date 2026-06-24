---
name: kas-project-overlay-guide
description: Author KAS project overlays without copying plugin base skills.
---

# KAS Project Overlay Guide

Use this guide when authoring a project-specific KAS overlay for a profile that
uses the KAS plugin base plus a thin color wrapper. Overlays add project facts,
constraints, evidence rules, and promotion notes. They do not copy or replace
the full plugin base skill body.

## Path

Use the profile-relative layout:

```text
skills/<project>/kas-overlays/<project>-<role>-<base-skill>-overlay/SKILL.md
```

Example:

```text
skills/doksuri/kas-overlays/doksuri-blue-plan-overlay/SKILL.md
```

## Valid Frontmatter Example

```yaml
---
name: doksuri-blue-plan-overlay
description: Doksuri-specific Blue planning overlay for KAS plugin plan guidance.
metadata:
  kas:
    kind: project_overlay
    project: doksuri
    role: blue_commander
    overlay_for: kkachi-agent-skills:plan
    merge_mode: additive_constraints
    base_version: "5656dad"
    authority_sources:
      - /path/to/doksuri/docs/sot.md
    last_reviewed: "2026-06-24"
    promotion_candidate: false
---
```

Allowed content:

- project authority sources, repo paths, docs maps, and review lanes;
- project-specific verification commands and evidence artifact names;
- local risk boundaries, fail-closed rules, approval gates, and non-goals;
- role-specific wording for Blue, Red, Orange, or Gray participation;
- selected KAB adoption-stage notes when project evidence supports them;
- Teal/UI notes only when the project has a Teal lane and the task changes UI;
- promotion notes for repeated local improvements.

Forbidden content:

- copied full base skill bodies;
- hidden fallback when the plugin base is missing;
- unreviewed replacement of plugin base safety constraints;
- secrets, tokens, gateway credentials, provider/model settings, or runtime state;
- broad KAS/KAH/KAB claims not grounded in current SOT evidence;
- role expansion beyond the profile's registered role.

## Invalid Example

This overlay is invalid because `overlay_for` does not name a plugin-qualified
KAS base target:

```yaml
metadata:
  kas:
    kind: project_overlay
    project: doksuri
    role: blue_commander
    overlay_for: kkachi-plan
    merge_mode: additive_constraints
```

Stop and request review instead of repairing this silently. Report the exact
invalid field, do not use a stale profile-local copied skill as fallback, and do
not promote the overlay into the KAS base without review.
