# KAS Installed Profile vs Source Promotion

Use this note when 주군 asks to update KAS guidance in an active Hermes profile first, then later promote the generalized result into the KAS source repository.

## Correct layers

- Installed active KAS lives under the active Hermes profile, for example:
  - `$HERMES_HOME/skills/...`
  - `~/.hermes/profiles/<profile>/skills/...`
- KAS source lives in the source repository, for example:
  - `<workspace>/kkachi-hermes-skills/skills/...`
- Project KAH state/config lives in the target project, for example:
  - `<project>/.kkachi/`
  - `<project>/.kkachi-workflow.yaml`

## Operating rule

When 주군 asks to update an installed KAS skill or project-specific KAS guidance first, patch the installed Hermes profile skill/reference. Do not edit KAS source as the first move unless 주군 explicitly asks for KAS source promotion, official KAS update, or a commit-ready KAS repo change.

When 주군 later approves source promotion, treat it as a separate KAS repository task. Compare installed profile guidance against source, port only generalized policy/procedure, and avoid copying profile-local, project-local, or operator-specific details verbatim.

## Multi-skill policy correction workflow

When 주군 corrects KAS/KAH operating policy and asks to update installed profile first:

1. Patch the active profile's class-level KAS/KAH skills that govern the policy, not a project-local `skills/` directory and not KAS source.
2. Prefer the loaded/in-play skills first, then existing `references/` support files. For review/feedback/final-gate policy, this usually means `kkachi-orchestrate`, `kkachi-phase-state`, `kkachi-review`, `kkachi-request-feedback`, `kkachi-handle-feedback`, and `kkachi-final-verify` as applicable.
3. Encode future behavior in `SKILL.md`; use `references/` for examples, promotion notes, or longer rationale.
4. Verify the installed profile files structurally after patching and report whether KAS source was changed.
5. For source promotion, port generalized policy into source skills/refs, run skill validation, run the repository verification target, request color review, and leave the repo uncommitted until 주군 approves.

## References directory rule

`references/` is an optional support directory for a skill. It is not required for skill promotion. Promotion can update `SKILL.md` directly, add or update a generic support reference, add a template/script, or decide not to promote project-specific guidance at all.

Do not infer that every project using KAS needs a project-named reference in the KAS repo. Project-specific references in source KAS are a design smell unless the project is intentionally a canonical exemplar or the content has been generalized.

## Cleanup rule

If a project-specific reference was added to KAS source by mistake:

1. Check whether `SKILL.md` or other files link to it.
2. Remove those links or replace them with generic guidance.
3. Delete the project-specific source reference.
4. Preserve installed profile guidance if it is still needed for the active operator.
5. Verify no dangling references remain before reporting.

## Common pitfall

Do not conflate three meanings of `reference`:

- a Hermes skill support file under `references/`;
- an installed profile copy used by the active agent;
- a source-repo file that would be shipped to all users after promotion.

The same path shape can exist in both installed profile and source repo, but their authority differs.
