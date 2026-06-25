---
name: kas-overlay-compose-guide
description: Compose KAS plugin base, thin wrappers, and project overlays safely.
---

# KAS Overlay Compose Guide

Use this guide when deciding how to combine a current task, project overlay,
project wrapper, and KAS plugin base. Composition is additive and fail-closed.

## Composition Order

Apply guidance in this order:

1. Current command, card, channel, and explicit approvals.
2. Project overlay: project-specific constraints and evidence rules.
3. Project wrapper: project/role boundary, base skill selection, and overlay lookup.
4. KAS plugin base: canonical phase or operation behavior.

This order does not permit overriding safety constraints for convenience.

## Example

For a Doksuri Blue planning task:

```text
current context:
  task: plan a Doksuri implementation
  approval: planning only
project overlay:
  skills/doksuri/doksuri-overlay/SKILL.md
  references: skills/doksuri/doksuri-overlay/references/verification.md
wrapper:
  doksuri-wrapper
plugin base:
  skill_view("kkachi-agent-skills:plan")
```

The overlay may add Doksuri-specific docs, commands, architecture, and evidence
expectations across the plugin base targets named in `applies_to`. It must not
remove KAS fail-closed rules or replace a base skill body without approved
replacement evidence.

## Merge Modes

- `additive_constraints`: add project constraints, paths, or checks.
- `evidence_policy`: tighten or specify project evidence requirements.
- `authority_overlay`: clarify project SOT/reviewer lanes.
- `replacement_candidate`: propose a reusable base change for later review.
- `replacement_approved`: replace a named base subsection only with approval
  reference, base version, and doctor support.

## Optional Advanced Overlays

Use `skills/<project>/kas-overlays/<project>-<role>-<phase-or-base>-overlay/`
only when the single project overlay is too coarse. Advanced overlays must still
reference plugin-qualified base targets, must not copy base bodies, and must not
be treated as the default project setup.

## Conflict Handling

If the overlay conflicts with plugin base safety, role authority, or a current
SOT, stop. Report the exact conflict, do not select the more convenient rule,
do not use profile-local copied skills as fallback, and request Blue synthesis
or explicit review.

Missing plugin base, role manifest, project wrapper, or project overlay is also
fail-closed: report the missing surface and wait for review/approval.
