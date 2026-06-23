# SKILL plugin-wrapper-overlay contract

Date: 2026-06-24
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / SKILL-001 discussion-origin SOT pending color review and implementation evidence
Status: candidate SOT for `SKILL` epic; records target architecture only. No Hermes plugin package, profile wrapper rollout, project overlay migration, profile cleanup, KAH/KAB runtime behavior, or installed-profile mutation is implemented or authorized by this document alone.
Authority level: KAS-side planning source of truth for plugin-backed KAS base skills, color role wrappers, profile-local project overlays, guide skills, composition policy, doctor diagnostics, and migration from copied profile-local base suites.
Scope: `kkachi-agent-skills` docs, registries, source skill packaging, future Hermes plugin registration, wrapper/overlay guidance, profile-skill manifest semantics, and KAS doctor planning. No auth/token/gateway/provider/model mutation, live Hermes runtime activation, KAH deterministic state mutation, KAB backend/session activation, Kanban routing change, broad profile cleanup, or automatic rollout is authorized by this SOT.
Related docs: `docs/sot/kas-cli-contract.md`, `docs/sot/project-specific-kas-install-contract.md`, `docs/sot/role-aware-project-suite-contract.md`, `docs/sot/project-kas-sync-state.md`, `docs/sot/kasrel-hermes-v016-provenance-contract.md`, `docs/sot/token-economy-and-agent-instruction-contract.md`, `docs/roadmap.md`, Hermes plugin skill documentation.
Evidence/source paths:
- 주군 direction in 17번째 지구 Discord `#kas` thread `1519035413417951372` on 2026-06-24: Hermes plugin skills are managed separately and are not copied into each profile skill directory; repeated KAS profile-copy upgrades force manual rebase/port work; propose plugin skills plus thin wrappers and profile-local project optimization overlays.
- Current local profile observation on 2026-06-24: `hwangchung` has copied KAS/KAH base suites (`kkachi-agent-skills` 23 + `kkachi-agent-helper` 23), `hahuyeon` has review/verify copies for KAS/KAH, `yeomong` has review copies for KAS/KAH, and `jingung` has review/final-verify copies for KAS/KAH. The observed state is migration evidence, not the target model.
- Existing KAS docs already define role-aware project-suite boundaries and project-specific install/update semantics; this SOT changes the target steady state from copied base suites toward plugin base + thin wrapper + project overlay composition.

## 1. Decision summary

The `SKILL` epic changes the intended KAS skill lifecycle from **copy KAS base skills into every Hermes profile** to **Hermes plugin base skills plus profile-local wrappers and project overlays**.

The target model is:

1. **KAS plugin base**: canonical KAS base skills live in the KAS source repo and are registered as Hermes plugin skills. They are loaded with plugin-qualified names such as `kkachi-agent-skills:plan` and are treated as read-only base behavior from the profile point of view.
2. **Color role manifests**: Blue/Red/Orange/Gray role selections live in KAS source metadata. They define which plugin base skills each color role may normally use.
3. **Thin profile wrappers**: each active color profile keeps a small, discoverable profile-local wrapper skill that states its role, points to plugin-qualified base skills, and explains overlay composition. The wrapper is not a copy of the KAS base pack.
4. **Project overlays**: project-specific optimization lives under the owning profile's skill tree, inside a project-named directory, as additive/constraint/evidence overlays that reference a plugin base target. Project overlays do not edit plugin base skills and do not copy full base skill bodies.
5. **Guide skills**: KAS must ship guide skills that teach agents how to author, compose, validate, and promote project overlays without forking the base.
6. **Plugin update lifecycle**: the KAS CLI update surface must update the official KAS plugin package only, including plugin registration metadata, canonical base skills, role manifests, and guide skills. It must not overwrite profile wrappers, project overlays, KAH state, KAB runtime state, or profile-local personal skills.
7. **Doctor and migration**: KAS doctor must detect legacy copied base suites, plugin registration gaps, wrapper/role mismatches, stale/shadowing overlays, and migration candidates. Cleanup remains dry-run-first and approval-gated.

This SOT intentionally preserves project-specific optimization as a first-class requirement. The change is not “no local skill customization”; it is “local customization is a delta overlay, not a forked base copy.”

## 2. Problem statement

The existing profile-copy model makes every installed KAS suite a potential fork. When the KAS source repo changes, each profile-local copy may contain a mixture of old base text, local edits, project-specific adaptations, stale KAH/KAB policy, and role-specific constraints. Upgrading then becomes a manual rebase: start from the new base, recover the useful old differences, and avoid carrying stale or unsafe behavior forward.

This is costly for Blue profiles with broad KAS coverage and risky for Red/Orange/Gray profiles because copied skills remain discoverable even when role prompts say not to use them. `docs/sot/role-aware-project-suite-contract.md` already fixes role over-installation at the copied-suite layer. The `SKILL` epic goes one layer further: the target base should not be copied into profiles by default.

The target architecture must still support the practical benefit of prior project-specific installs: after a project adopts KAS, the profile may need project-specific authority sources, repo paths, test commands, evidence formats, Teal conditions, KAB stage notes, and local workflow expectations. Those belong in project overlays.

## 3. Ownership boundary

| Layer | Owns | Must not own |
|---|---|---|
| KAS plugin base | canonical KAS phase/operation skills, color role manifests, guide skills, plugin registration metadata, base versioning, and base dependency guidance | profile-local project authority, profile secrets, gateway/provider/model runtime, project-specific semantic edits that only apply to one profile/project |
| Profile wrapper | discoverability for the active profile, color role declaration, plugin-qualified base references, overlay lookup rules, and role-specific safety reminders | full base skill bodies, broad KAS fork content, hidden fallback to missing plugin skills, or authority outside the profile's registered role |
| Project overlay | project-specific SOT pointers, repo paths, verification commands, local risk boundaries, evidence format additions, role-local constraints, and promotion candidates | editing plugin base skills, copying entire base bodies, removing fail-closed base constraints without review, changing auth/tokens/gateway/provider/model state |
| KAS plugin update command | official KAS plugin package update planning/apply for plugin metadata, plugin-registered base skills, role manifests, guide skills, and plugin version/readback evidence | profile-local overlay merge, wrapper overwrite, legacy copied-suite cleanup, personal skill mutation, KAH state writes, KAB runtime activation, auth/token/gateway/provider/model mutation |
| KAS doctor | read-only diagnosis of plugin/wrapper/overlay/base-copy drift and approval-gated repair planning | unapproved deletion, silent adoption of unknown profile skills, live runtime mutation, or KAH/KAB state changes |
| KAH | deterministic project/run/workflow state and evidence gates when a KAS run uses KAH | KAS plugin packaging, KAS plugin update, profile skill inventory policy, subjective role selection |
| KAB | backend/session/runtime evidence when an approved lane dispatches work | KAS skill distribution or overlay composition policy |

## 4. Target architecture

The `SKILL` epic defines four active surfaces.

```text
KAS source repo
  Hermes plugin package
    plugin.yaml / plugin registration code
    skills/<base-skill>/SKILL.md
    roles/blue.yaml
    roles/red.yaml
    roles/orange.yaml
    roles/gray.yaml
    guides/kas-project-overlay-guide/SKILL.md
    guides/kas-overlay-compose-guide/SKILL.md
    guides/kas-overlay-doctor-guide/SKILL.md

Hermes profile skill tree
  skills/<color-wrapper>/SKILL.md
  skills/<project>/kas-overlays/<project>-<role>-<base-skill>-overlay/SKILL.md
```

The exact in-repo plugin package path may be chosen during SKILL-002 implementation, but the implementation must preserve these public contracts:

- KAS base skills are registered as Hermes plugin skills and are loadable through plugin-qualified names.
- Plugin skill names do not require copying into `~/.hermes/profiles/<profile>/skills/`.
- Color role manifests are source-controlled and deterministic.
- Profile-local wrappers and overlays are intentionally small and may reference plugin-qualified base names.
- Any different packaging path or namespace requires a SOT update before implementation is marked complete.

## 5. Color role model

The initial role projection follows the current Kkachi color boundaries.

| Color role | Example profile | Initial KAS plugin base selection | Boundary |
|---|---|---|---|
| Blue commander | `hwangchung` | full KAS phase/operation set, currently 23 KAS base skills | May plan, orchestrate, authorize implementation, request review, verify, and synthesize evidence within KAS/KAO authority. |
| Red reviewer | `hahuyeon` | `review`, `verify` | Safety, fail-closed, risk, evidence sufficiency, deterministic/runtime risk review. No implementation or backend-routing authority by skill presence. |
| Orange PM reviewer | `yeomong` | `review` | Operator value, scope, workflow clarity, acceptance criteria, report readability, approval/recovery clarity. No implementation or backend-routing authority by skill presence. |
| Gray scribe | `jingung` | `review`, `final-verify` | Decision trace, evidence paths, stale/candidate/accepted/runtime-approved distinctions, final-gate evidence support. `final-verify` does not replace Blue final authority. |

KAH companion skills may have analogous color subsets, but this SOT owns KAS first. KAH mirror behavior must be updated through the KAH repo or a recorded KAS/KAH companion task rather than silently assuming this KAS SOT mutates KAH distribution.

Unknown, misspelled, or unregistered color roles fail closed. A profile wrapper must not infer elevated capability from profile name alone; it must reference a registered role manifest and the registry/profile authority surface.

## 6. Profile wrapper contract

A profile wrapper is a discoverability and composition entrypoint. It should be visible in ordinary profile-local skill discovery so the model can learn the local operating rule even when plugin skills are not listed in the same flat skill inventory.

Minimum wrapper fields:

```yaml
---
name: kkachi-blue-wrapper
description: Thin profile-local wrapper for Blue KAS plugin skills and project overlays.
metadata:
  kas:
    kind: color_wrapper
    role: blue_commander
    role_manifest: kkachi-agent-skills:roles/blue.yaml
    plugin_namespace: kkachi-agent-skills
    overlay_root: skills/<project>/kas-overlays
    base_copy_policy: forbidden_by_default
---
```

The wrapper must:

1. instruct the agent to load plugin-qualified KAS base skills such as `skill_view("kkachi-agent-skills:plan")` when a matching phase skill is needed;
2. name the role manifest and allowed role subset;
3. explain how project overlays are selected and composed;
4. fail closed when the plugin base skill or role manifest is unavailable;
5. avoid embedding or duplicating full base skill bodies;
6. preserve Kkachi authority boundaries and explicit approval gates.

Wrappers may contain local phrasing for the profile's Korean name, role, and channels, but they must not become project-specific dumping grounds. Project-specific material belongs in project overlays.

## 7. Project overlay contract

Project overlays are the approved place for project-specific KAS optimization under the plugin model.

Canonical profile-relative layout:

```text
skills/<project>/kas-overlays/<project>-<role>-<base-skill>-overlay/SKILL.md
```

Examples:

```text
skills/doksuri/kas-overlays/doksuri-blue-plan-overlay/SKILL.md
skills/doksuri/kas-overlays/doksuri-blue-implement-overlay/SKILL.md
skills/sudal/kas-overlays/sudal-red-review-overlay/SKILL.md
skills/space-compiler/kas-overlays/space-compiler-orange-review-overlay/SKILL.md
```

Minimum overlay frontmatter:

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
    base_version: "<kas-version-or-commit>"
    authority_sources:
      - /path/to/project/sot.md
    last_reviewed: "YYYY-MM-DD"
    promotion_candidate: false
---
```

Allowed overlay content:

- project authority sources, repo paths, docs maps, SOT freshness rules, and reviewer lanes;
- project-specific test/verification commands and evidence artifacts;
- local risk boundaries, fail-closed rules, approval gates, and non-goals;
- role-specific wording for Blue/Red/Orange/Gray participation in the project;
- KAB adoption stage notes, only as selected/evidenced project-local guidance;
- Teal/UI policy application notes when the project has a Teal lane and the task changes UI/UX;
- promotion notes when a repeated local improvement should become common KAS guidance.

Forbidden overlay content:

- copied full base skill bodies;
- unreviewed replacement of plugin base safety constraints;
- hidden fallback when the plugin base is missing;
- auth tokens, secrets, gateway credentials, provider/model settings, or live runtime state;
- broad KAS/KAH/KAB architecture claims not grounded in current SOT/evidence;
- role expansion beyond the profile's registry authority.

## 8. Composition semantics

When a task uses KAS under the `SKILL` target model, the agent composes guidance in this order:

1. **Current command/card/channel context**: the user request, Kanban card, current channel prompt, and explicit approvals are the immediate task context.
2. **Project overlay**: profile-local project-specific constraints and evidence rules add to or narrow the base behavior for the selected project.
3. **Color wrapper**: the profile's role wrapper supplies role boundary, base skill selection, and overlay lookup policy.
4. **KAS plugin base**: the canonical plugin skill supplies default phase/operation behavior.

This order is not permission to silently override safety. Merge policy is fail-closed:

| Merge mode | Meaning | Conflict behavior |
|---|---|---|
| `additive_constraints` | Overlay adds project constraints, paths, or checks without changing base semantics. | Safe default when target base exists. |
| `evidence_policy` | Overlay tightens or specifies project evidence/reporting requirements. | Missing evidence becomes a task blocker or request-changes condition. |
| `authority_overlay` | Overlay clarifies project authority lanes, reviewers, or SOT precedence. | Conflicts with registry/SOT require Blue synthesis and review before use. |
| `replacement_candidate` | Overlay proposes a base behavior change for future promotion. | Must not be treated as current replacement until reviewed and promoted. |
| `replacement_approved` | Overlay replaces a named base subsection after explicit review/approval. | Requires approval reference, base version, and doctor support; absent evidence is an error. |

If an overlay conflicts with plugin base safety, role authority, or a current SOT, the agent must stop and request clarification/review rather than selecting the more convenient rule.

## 9. Guide skill requirements

The plugin must provide guide skills so future agents do not recreate copied forks.

Required guide skills:

1. `kas-project-overlay-guide`: authoring guide for project overlays, including path/frontmatter rules, allowed content, forbidden content, and examples.
2. `kas-overlay-compose-guide`: runtime composition guide for selecting plugin base, wrapper, and project overlay; includes merge modes and conflict handling.
3. `kas-overlay-doctor-guide`: diagnostic guide for identifying copied base suites, invalid overlays, stale base versions, missing plugin registration, shadowing, and promotion candidates.

These guide skills are KAS source skills, not profile-local per-project overlays. They may be exposed through the plugin and referenced by wrapper skills. If Hermes plugin skill discoverability remains weaker than flat profile-local discovery, the wrapper must explicitly instruct the agent to load these guide skills by plugin-qualified name.

## 10. Update command semantics

The routine update path under the `SKILL` target model is **plugin-package update first**, not profile-skill-copy refresh.

Canonical command shape:

```text
kkachi-agent-skills update plugin --dry-run [--json]
kkachi-agent-skills update plugin --apply dry-run:sha256:<hash> [--json]
```

The exact flag set may be refined during implementation, but the command semantics are fixed:

| Surface | Update command may change | Update command must not change |
|---|---|---|
| KAS plugin package | plugin metadata, registered base skill files, role manifests, guide skills, plugin version/readback evidence, and local plugin package lock/manifest when approved | profile-local wrappers, project overlays, legacy copied base suites, personal skills, KAH `.kkachi/` state, KAB runtime/session state, auth/token/gateway/provider/model config |
| KAS profile wrappers | no automatic write by plugin update; at most report compatibility/readback diagnostics and suggested wrapper template changes | silent wrapper overwrite or hidden fallback from stale wrapper content |
| Project overlays | no automatic write by plugin update; at most report stale `base_version`, invalid `overlay_for`, or conflict diagnostics | automatic merge, semantic port, deletion, or promotion into plugin base |
| Legacy copied skills | no automatic write by plugin update; at most classify as legacy/shadowing risk | deletion, adoption, or use as fallback base |

`update plugin` must be dry-run-first and approval-hash-bound for writes. A dry-run must include the target plugin namespace, current version/source, proposed version/source, planned changed paths, role manifest impact, guide skill impact, no-write evidence, and the doctor command that should be run after apply. Apply must recompute the current plan and fail closed if the approved hash no longer matches.

Existing project-suite lifecycle update behavior must not remain ambiguous with plugin update. If the current `update` command continues to serve project-specific KAS state/classification, its public surface must be scoped explicitly, for example `update project-suite ...`, while `update agent-instructions` remains a separate repo-instruction surface. None of these update surfaces may copy the official KAS base pack into profile skill directories as the steady-state path.

KAH has no implementation responsibility for `update plugin`. KAH may later record a KAS update evidence reference in run artifacts or gate reports, but only as evidence preservation; KAH must not install, update, vendor, merge, or clean Hermes/KAS skill content.

## 11. Doctor, manifest, and diagnostic semantics

KAS doctor must gain a read-only SKILL mode before any cleanup/apply behavior.

Minimum diagnostics:

| Condition | Severity | Meaning |
|---|---|---|
| KAS plugin missing or disabled | error | Wrapper/overlay cannot safely resolve base skills. |
| Required plugin base skill missing | error | The selected KAS phase cannot run from canonical base. |
| Role manifest missing or unknown | error | The profile's color role cannot be safely projected. |
| Profile wrapper missing for an active color profile | warning or error by rollout stage | Discoverability/composition is incomplete. |
| Wrapper role mismatches registry/profile authority | error | Role boundary is unsafe. |
| Profile-local full KAS base copy present | warning during migration, error after cutover | Legacy fork/shadow risk. |
| Overlay lacks `overlay_for` or references non-plugin target | error | Cannot safely compose overlay. |
| Overlay copies too much base content | error or warning by threshold | Likely fork instead of delta. |
| Overlay shadows base skill name | error | Ambiguous skill identity. |
| Overlay base version is stale | warning | Needs review against current plugin base. |
| Overlay contains secrets/runtime settings | error | Must be removed before use. |
| Overlay improvement appears reusable | info | Promotion candidate for KAS guide/base review. |
| Plugin update available or partially applied | warning or error by rollout stage | Run `update plugin --dry-run` or complete approved apply before claiming current plugin-base readiness. |
| `update` command surface remains ambiguous | warning before cutover, error after cutover | Operators must be able to distinguish plugin update from project-suite update and agent-instruction update. |
| KAH appears to own KAS plugin update/install | error | Boundary violation; KAH may preserve evidence refs only. |

Doctor output must distinguish `plugin_base`, `wrapper`, `project_overlay`, `legacy_copy`, `personal_skill`, and `unknown` source classes. It must not delete, rewrite, or adopt profile-local skills without explicit dry-run/apply approval.

## 12. Migration policy

Existing profile-local KAS/KAH copied suites are migration input, not the steady-state target.

Migration must be dry-run-first and classify each profile-local KAS-like skill as:

- `base_identical`: content matches plugin/source base and can be removed after approval once wrapper/plugin readiness is proven;
- `base_with_local_delta`: content differs and needs semantic extraction into a project overlay or common guide candidate;
- `project_overlay_candidate`: local content is project-specific and should be moved into `skills/<project>/kas-overlays/...`;
- `role_wrapper_candidate`: local content belongs in a thin profile wrapper;
- `unknown_personal_skill`: preserve and request review;
- `kah_companion_surface`: belongs to KAH or a paired KAS/KAH migration task, not silent KAS-only cleanup.

Apply must require explicit 주군 approval for exact profiles/projects, backup/vault evidence, no-spillover scan, recovery instructions, and post-apply doctor evidence. This SOT alone does not authorize deleting the currently observed copied profile skills.

## 13. Relationship to existing SOTs

- `docs/sot/project-specific-kas-install-contract.md` remains authority for the legacy/project-suite copied layout until SKILL implementation supersedes specific install paths through reviewed tasks. The SKILL target model treats copied project suites as a migration source and shifts future steady state to overlays.
- `docs/sot/role-aware-project-suite-contract.md` remains authority for role-aware subset safety. SKILL reuses those color boundaries for plugin role manifests and wrappers.
- `docs/sot/project-kas-sync-state.md` remains authority for project-specific state/sync until SKILL replaces or extends state files for overlay/version tracking.
- `docs/sot/kasrel-hermes-v016-provenance-contract.md` remains authority for release/provenance diagnostics. SKILL adds plugin-base/wrapper/overlay source classes that must integrate with provenance reporting.
- KAH and KAB behavior must be updated through their own docs/tasks when companion changes are needed. The expected default for this SKILL change is no KAH implementation change: KAH may preserve KAS lifecycle evidence references in artifacts/gates, but KAS must not imply KAH owns plugin packaging, update, wrapper/overlay composition, cleanup, or KAB runtime support from plugin/wrapper docs alone.

## 14. Required roadmap sequence

`SKILL` is a planning and migration epic. It should start with SOT acceptance and then proceed in small PR-candidate tasks.

1. `SKILL-001` — accept this plugin-wrapper-overlay SOT, register the docs/roadmap references, and record color review / responsible approval.
2. `SKILL-002` — implement KAS Hermes plugin package registration for canonical base skills, role manifests, guide skill exposure, and the plugin-package update lifecycle surface; prove plugin-qualified skill loading and update dry-run/readback in non-mutating smoke tests.
3. `SKILL-003` — add thin color wrapper templates and guide skills (`kas-project-overlay-guide`, `kas-overlay-compose-guide`, `kas-overlay-doctor-guide`); verify wrapper instructions do not copy base bodies.
4. `SKILL-004` — implement read-only SKILL doctor diagnostics for plugin/wrapper/overlay/legacy-copy provenance and invalid overlay detection.
5. `SKILL-005` — implement dry-run migration classifier for existing copied KAS/KAH-like profile skills, including semantic extraction packets and no-spillover evidence; no deletion.
6. `SKILL-006` — run one approved pilot migration for exact profile/project targets after explicit 주군 approval, backup, no-spillover scan, and recovery plan.

Later tasks may merge SKILL behavior into `install`, `doctor`, `sync-project-kas`, `update project-suite`, or replacement commands only after the earlier SKILL tasks provide evidence. The public update surface must remain unambiguous: plugin-package update is not project-suite update and neither is KAH-owned.

## 15. Non-goals and deferrals

- No immediate deletion or rewrite of `hwangchung`, `hahuyeon`, `yeomong`, or `jingung` profile-local skills.
- No automatic broad rollout to all Kkachi profiles.
- No auth/token/gateway/provider/model mutation.
- No KAH deterministic state mutation.
- No KAH ownership of KAS plugin install/update, wrapper/overlay composition, or cleanup.
- No KAB Stage 2/3 activation.
- No hidden fallback from missing plugin base to stale local copies.
- No profile-local full KAS base copy as a healthy final state.
- No plugin base edits for one-project-only optimization.
- No promotion of project overlay content into KAS base without review.
- No claim that Hermes plugin skill discoverability is identical to flat profile-local skill discovery; wrappers must compensate explicitly.

## 16. Acceptance gates

Before KAS claims SKILL support as implemented:

- Plugin-qualified KAS base skills load from the Hermes plugin path without copying into profile skill directories.
- `kkachi-agent-skills update plugin --dry-run --json` reports only official plugin-package changes and no profile wrapper/project overlay/KAH/KAB/auth/provider/runtime mutation.
- Blue/Red/Orange/Gray role manifests are deterministic, source-controlled, and tested.
- Thin wrapper templates exist and are small enough to be discoverability/composition surfaces rather than base forks.
- Project overlay frontmatter and layout are documented and tested with valid/invalid fixtures.
- Doctor reports plugin base, wrapper, overlay, legacy copy, personal skill, and unknown source classes distinctly.
- Doctor reports plugin update readiness, ambiguous update command surfaces, and KAH-boundary violations distinctly.
- Migration dry-run classifies existing copied profile-local KAS/KAH-like skills without writes.
- Cleanup/apply requires explicit 주군 approval, backup, no-spillover scan, and recovery evidence.
- `docs/README.md`, `docs/kkachi-docs-map.yaml`, and `docs/roadmap.md` register this SOT and preserve the no-implementation boundary.
- Red/Orange/Gray review accepts the safety, operator clarity, and evidence trace before broad rollout.
