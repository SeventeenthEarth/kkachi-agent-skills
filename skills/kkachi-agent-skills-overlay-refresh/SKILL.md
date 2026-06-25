---
name: kkachi-agent-skills-overlay-refresh
description: Use when refreshing a project overlay after kkachi-agent-skills changes; preserve legacy overlay temporarily, rebuild from the current template, semantically port only durable project-specific rules, then remove the legacy archive after verification.
version: 0.1.0
---

# kkachi-agent-skills Overlay Refresh

## Overview

Use this skill when a project already has a `kkachi-agent-skills` project overlay and the project needs to refresh that overlay after a new `kkachi-agent-skills` version, wrapper template, guide, or policy change.

This workflow is intentionally LLM-assisted and **not a CLI migration**. The `kkachi-agent-skills` CLI must not perform semantic overlay merges because it cannot judge which legacy lines are durable project-specific constraints versus stale copied base behavior. The agent performs that judgment, preserves evidence, asks when uncertain, and leaves a reviewable diff.

## Base Skill Non-Mutation Rule

Overlay refresh is a profile-local/project-local operation. During this workflow, modify the active project overlay first:

```text
~/.hermes/profiles/<profile>/skills/<project>/<project>-overlay/SKILL.md
~/.hermes/profiles/<profile>/skills/<project>/<project>-overlay/references/*.md
```

Do **not** edit the official `kkachi-agent-skills` plugin/base skills as part of an overlay refresh. If a finding looks generally reusable for the plugin, record it as a separate promotion candidate and route it through the normal `kkachi-agent-skills` review/implementation gates; do not silently patch plugin skills while refreshing a project overlay.

## When to Use

Use this skill when:

- `skills/<project>/<project>-overlay/SKILL.md` already exists and must be refreshed against the current `kkachi-agent-skills` template or SOT.
- A previous overlay contains project-specific rules that may need to survive a new plugin/wrapper baseline.
- The old overlay must be compared against a clean new overlay before active content changes are accepted.

Do not use this skill for:

- initial project setup with no existing overlay; use the install/project-overlay authoring guides instead.
- automated profile cleanup or copied base-suite deletion.
- runtime, auth, token, provider, gateway, model, or kkachi-agent-bridge / kkachi-agent-helper state changes.
- changing official plugin base skills; reusable improvements become review/promotion candidates, not silent overlay rewrites.

## Canonical Paths

For project `<project>` under profile `<profile>`:

```text
~/.hermes/profiles/<profile>/skills/<project>/<project>-overlay/SKILL.md
~/.hermes/profiles/<profile>/skills/<project>/<project>-overlay/references/*.md
```

Temporary legacy archive during refresh:

```text
~/.hermes/profiles/<profile>/skills/<project>/<project>-overlay-legacy/SKILL.md
~/.hermes/profiles/<profile>/skills/<project>/<project>-overlay-legacy/references/*
```

The legacy archive is a short-lived comparison aid. After the refreshed overlay is verified and accepted, remove `<project>-overlay-legacy` from the active skills tree.

## Legacy Archive Frontmatter

When copying the active overlay to the temporary legacy archive, mark it clearly inactive:

```yaml
---
name: <project>-overlay-legacy
description: Legacy archive for temporary kkachi-agent-skills overlay refresh comparison. Do not load for active work.
version: 0.1.0
metadata:
  kas:
    kind: project_overlay_legacy
    project: <project>
    active: false
    superseded_by: <project>-overlay
    archive_purpose: temporary_refresh_comparison
---
```

If the original overlay frontmatter contains project metadata, preserve it below the archive metadata or in `references/legacy-frontmatter.md`; do not let the archive claim to be an active `project_overlay`.

## Refresh Workflow

1. **Preflight and scope**
   - Confirm exact `<profile>` and `<project>`.
   - Read the current active overlay and references.
   - Read current wrapper, plugin guide, and SOT basis.
   - Record that this is an LLM-assisted overlay refresh, not a CLI update or migration.

2. **Create temporary legacy archive**
   - Copy `<project>-overlay/` to `<project>-overlay-legacy/`.
   - Rewrite the legacy archive `SKILL.md` frontmatter to `kind: project_overlay_legacy`, `active: false`, and `superseded_by: <project>-overlay`.
   - Do not delete the active overlay until the fresh overlay is ready to write.

3. **Create fresh overlay skeleton**
   - Build a new `<project>-overlay/` from the current canonical project overlay template/SOT.
   - Include only current metadata fields such as `kind: project_overlay`, `project`, `plugin_namespace: kkachi-agent-skills`, `applies_to`, `merge_mode: additive_constraints`, `base_version`, and `references`.
   - Create `references/` files only for durable project context, architecture, verification, authority, or environment facts.

4. **Semantic port from legacy**
   - Compare legacy archive against the fresh skeleton.
   - Port only project-specific durable rules, paths, architecture notes, verification requirements, authority constraints, and recurring lessons.
   - Do not port copied plugin base instructions, stale phase text, runtime secrets, provider/model settings, fallback behavior, or rules that conflict with current SOT/plugin base.
   - If a legacy rule appears generally reusable, record it as a promotion candidate instead of embedding it silently in the project overlay.
   - If confidence is low, stop and ask 주군 or route review; do not guess.

5. **Verify and review**
   - Run `kkachi-agent-skills doctor --plugin --repo <kkachi-agent-skills-repo> --profile <profile> --project <project> --json` when available.
   - Inspect the diff: active overlay should be project-specific and compact; legacy archive should be inactive.
   - Confirm no copied base suite was used as fallback and no unrelated profile/project was touched.

6. **Remove temporary legacy archive**
   - After the refreshed overlay passes review/verification, delete `<project>-overlay-legacy/`.
   - Record the diff/evidence that the active overlay retained required project-specific content.

## Porting Decision Rules

Port to active overlay:

- project paths, repository layout, language/runtime architecture, test commands, evidence paths, and durable team authority constraints.
- project-specific safety rules that narrow plugin base behavior.
- recurring lessons from real project runs that remain true after the current `kkachi-agent-skills` version.

Do not port:

- copied phase-skill bodies from the plugin base.
- stale workaround text that the new plugin base already covers.
- commands that mutate auth, tokens, provider settings, gateway config, model defaults, or live kkachi-agent-bridge / kkachi-agent-helper runtime state.
- fallback behavior that would hide missing plugin/wrapper/overlay evidence.
- one-off task progress, commit IDs, PR numbers, or evidence that will go stale.

## Required Report Shape

Report in Korean to 주군 with:

- `Status`: refreshed, blocked, or review-required.
- `Changed`: active overlay files changed and whether legacy archive was removed.
- `Ported`: concise list of legacy rules kept.
- `Rejected`: copied/stale/unsafe legacy content not ported.
- `Verification`: doctor/tests/diff commands and results.
- `Risks`: uncertainty, manual review needs, or approval gates.

## Common Pitfalls

1. **Leaving legacy active.** `overlay-legacy` must be marked `project_overlay_legacy` and `active: false`, then removed after refresh completes.
2. **Treating CLI as semantic migrator.** The CLI can verify and inspect; it must not decide which legacy prose to port.
3. **Copying base text forward.** The refreshed overlay should contain project-specific constraints, not a fork of plugin base skills.
4. **Using short aliases in durable docs.** Use `kkachi-agent-skills` in documentation and skill names; reserve short forms for casual conversation only.
5. **Deleting evidence too early.** Keep the temporary legacy archive until diff review confirms required project knowledge was ported.

## Verification Checklist

- [ ] Exact profile and project were confirmed.
- [ ] Legacy archive was created with `kind: project_overlay_legacy` and `active: false`.
- [ ] New active overlay uses canonical `kind: project_overlay` metadata and `applies_to` plugin-qualified targets.
- [ ] Only durable project-specific content was ported.
- [ ] Copied base text, stale workaround text, and unsafe runtime configuration were rejected.
- [ ] `doctor --plugin` or documented equivalent verification was run.
- [ ] `<project>-overlay-legacy/` was removed after successful refresh and review.
