---
name: kas-project-overlay-guide
description: Author KAS project overlays without copying plugin base skills.
---

# KAS Project Overlay Guide

Use this guide when authoring a project-specific KAS overlay for a profile that
uses the KAS plugin base plus a thin project wrapper. Overlays add project facts,
constraints, evidence rules, phase notes, and promotion notes. They do not copy
or replace the full plugin base skill body.

## Default Path

Use the profile-relative canonical layout:

```text
skills/<project>/<project>-overlay/SKILL.md
skills/<project>/<project>-overlay/references/*.md
```

Example:

```text
skills/kkachi-agent-helper/kkachi-agent-helper-overlay/SKILL.md
skills/kkachi-agent-helper/kkachi-agent-helper-overlay/references/verification.md
```

Use optional advanced phase/base overlays only when one project overlay is too
coarse and the reason is reviewed:

```text
skills/<project>/kas-overlays/<project>-<role>-<phase-or-base>-overlay/SKILL.md
```

## Valid Frontmatter Example

```yaml
---
name: kkachi-agent-helper-overlay
description: Project-specific KAS overlay for kkachi-agent-helper.
metadata:
  kas:
    kind: project_overlay
    project: kkachi-agent-helper
    role: blue_commander
    plugin_namespace: kkachi-agent-skills
    applies_to:
      - kkachi-agent-skills:orchestrate
      - kkachi-agent-skills:plan
      - kkachi-agent-skills:implement
      - kkachi-agent-skills:verify
      - kkachi-agent-skills:final-verify
    merge_mode: additive_constraints
    base_version: "5656dad"
    authority_sources:
      - /path/to/kkachi-agent-helper/docs/sot.md
    references:
      - references/project-context.md
      - references/architecture.md
      - references/verification.md
      - references/authority-sources.md
    last_reviewed: "2026-06-24"
    promotion_candidate: false
---
```

Allowed content:

- project authority sources, repo paths, docs maps, and review lanes;
- programming language, architecture, backend, and verification notes;
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

This overlay is invalid because `applies_to` does not name plugin-qualified KAS
base targets:

```yaml
metadata:
  kas:
    kind: project_overlay
    project: kkachi-agent-helper
    role: blue_commander
    applies_to:
      - kkachi-plan
    merge_mode: additive_constraints
```

Stop and request review instead of repairing this silently. Report the exact
invalid field, do not use a stale profile-local copied skill as fallback, and do
not promote the overlay into the KAS base without review.
