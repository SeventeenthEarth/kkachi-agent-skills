# SKILL plugin-wrapper-overlay contract

Date: 2026-06-24
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / SKILL-001 accepted SOT; SKILL-002 implementation evidence is tracked separately
Status: accepted SOT for `SKILL` epic; implementation is phased by roadmap task. SKILL-002 implements the plugin package and source role manifests without any public `update` or `migrate` CLI surface. No project wrapper rollout, project overlay refresh, profile cleanup, KAH/KAB runtime behavior, or installed-profile mutation is implemented or authorized by this document alone.
Authority level: KAS-side planning source of truth for plugin-backed KAS base skills, project wrappers, profile-local project overlays, guide skills, composition policy, doctor diagnostics, and migration from copied profile-local base suites.
Scope: `kkachi-agent-skills` docs, registries, source skill packaging, future Hermes plugin registration, wrapper/overlay guidance, profile-skill manifest semantics, and KAS doctor planning. No auth/token/gateway/provider/model mutation, live Hermes runtime activation, KAH deterministic state mutation, KAB backend/session activation, Kanban routing change, broad profile cleanup, or automatic rollout is authorized by this SOT.
Related docs: `docs/sot/kas-cli-contract.md`, `docs/sot/project-specific-kas-install-contract.md`, `docs/sot/role-aware-project-suite-contract.md`, `docs/sot/project-kas-sync-state.md`, `docs/sot/kasrel-hermes-v016-provenance-contract.md`, `docs/sot/token-economy-and-agent-instruction-contract.md`, `docs/roadmap.md`, Hermes plugin skill documentation.
Evidence/source paths:
- 주군 direction in 17번째 지구 Discord `#kas` thread `1519035413417951372` on 2026-06-24: Hermes plugin skills are managed separately and are not copied into each profile skill directory; repeated KAS profile-copy upgrades force manual rebase/port work; propose plugin skills plus thin wrappers and profile-local project optimization overlays.
- Current local profile observation on 2026-06-24: `hwangchung` has copied KAS/KAH base suites (`kkachi-agent-skills` 23 + `kkachi-agent-helper` 23), `hahuyeon` has review/verify copies for KAS/KAH, `yeomong` has review copies for KAS/KAH, and `jingung` has review/final-verify copies for KAS/KAH. The observed state is migration evidence, not the target model.
- Existing KAS docs already define role-aware project-suite boundaries and project-specific install/update semantics; this SOT changes the target steady state from copied base suites toward plugin base + thin wrapper + project overlay composition.

## 1. Decision summary

The `SKILL` epic changes the intended KAS skill lifecycle from **copy KAS base skills into every Hermes profile** to **Hermes plugin base skills plus profile-local wrappers and project overlays**.

The target model is:

1. **KAS plugin base**: canonical KAS base skills live in the KAS source repo and are registered as Hermes plugin skills. They are loaded with plugin-qualified names such as `kkachi-agent-skills:plan` and are treated as read-only base behavior from the profile point of view. Short names such as `kkachi-agent-skills:plan` are in-plugin canonicalization to registered base IDs such as `kkachi-plan`; they are not profile fallback, copied-suite fallback, or permission to resolve outside the official plugin package.
2. **Color role manifests**: Blue/Red/Orange/Gray role selections live in KAS source metadata. They define which plugin base skills each color role may normally use.
3. **Thin project wrappers**: each approved profile/project target keeps a small, discoverable profile-local wrapper skill under the project directory. The wrapper states the project and role, points to plugin-qualified base skills, names the project overlay, and is not a copy of the KAS base pack.
4. **Project overlays**: project-specific optimization lives under the owning profile's skill tree, inside the same project-named directory, as one default project overlay plus optional `references/` files. Project overlays add or narrow project facts, constraints, evidence rules, and phase notes across named plugin base targets; they do not edit plugin base skills and do not copy full base skill bodies.
5. **Guide skills**: KAS must ship guide skills that teach agents how to author, compose, validate, and promote project overlays without forking the base.
6. **No CLI update/migrate lifecycle**: `kkachi-agent-skills` must not expose public `update` or `migrate` commands for plugin packages, project suites, profile skills, or overlays. These names imply semantic migration/update authority that the deterministic CLI cannot safely provide.
7. **Doctor and overlay refresh**: `kkachi-agent-skills doctor` detects legacy copied base suites, plugin registration gaps, wrapper/role mismatches, stale/shadowing overlays, and refresh candidates. Semantic overlay refresh is handled by the `kkachi-agent-skills-overlay-refresh` skill, not by CLI migration.

This SOT intentionally preserves project-specific optimization as a first-class requirement. The change is not “no local skill customization”; it is “local customization is a delta overlay, not a forked base copy.”

## 2. Problem statement

The existing profile-copy model makes every installed KAS suite a potential fork. When the KAS source repo changes, each profile-local copy may contain a mixture of old base text, local edits, project-specific adaptations, stale KAH/KAB policy, and role-specific constraints. Upgrading then becomes a manual rebase: start from the new base, recover the useful old differences, and avoid carrying stale or unsafe behavior forward.

This is costly for Blue profiles with broad KAS coverage and risky for Red/Orange/Gray profiles because copied skills remain discoverable even when role prompts say not to use them. `docs/sot/role-aware-project-suite-contract.md` already fixes role over-installation at the copied-suite layer. The `SKILL` epic goes one layer further: the target base should not be copied into profiles by default.

The target architecture must still support the practical benefit of prior project-specific installs: after a project adopts KAS, the profile may need project-specific authority sources, repo paths, test commands, evidence formats, Teal conditions, KAB stage notes, and local workflow expectations. Those belong in project overlays.

## 3. Ownership boundary

| Layer | Owns | Must not own |
|---|---|---|
| KAS plugin base | canonical KAS phase/operation skills, color role manifests, guide skills, plugin registration metadata, base versioning, and base dependency guidance | profile-local project authority, profile secrets, gateway/provider/model runtime, project-specific semantic edits that only apply to one profile/project |
| Project wrapper | discoverability for the approved profile/project target, project declaration, color role declaration, plugin-qualified base references, paired overlay lookup rules, and role-specific safety reminders | full base skill bodies, broad KAS fork content, hidden fallback to missing plugin skills, or authority outside the profile's registered role |
| Project overlay | project-specific SOT pointers, repo paths, verification commands, local risk boundaries, evidence format additions, role-local constraints, and promotion candidates | editing plugin base skills, copying entire base bodies, removing fail-closed base constraints without review, changing auth/tokens/gateway/provider/model state |
| `kkachi-agent-skills` CLI lifecycle | deterministic install, doctor, repair, toolchain, workflow, uninstall, and version surfaces only; no public update or migrate commands | semantic overlay refresh, plugin/package update, project-suite migration, profile-skill migration, hidden cleanup, personal skill mutation, KAH state writes, KAB runtime activation, auth/token/gateway/provider/model mutation |
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
    guides/kkachi-agent-skills-overlay-refresh/SKILL.md

Hermes profile skill tree
  skills/<project>/<project>-wrapper/SKILL.md
  skills/<project>/<project>-overlay/SKILL.md
  skills/<project>/<project>-overlay/references/*.md

Optional advanced extension when a single project overlay is too coarse
  skills/<project>/kas-overlays/<project>-<role>-<phase-or-base>-overlay/SKILL.md
```

The exact in-repo plugin package path may be chosen during SKILL-002 implementation, but the implementation must preserve these public contracts:

SKILL-003 source package convention stores guide bodies at `skills/<guide-id>/SKILL.md`.
`guides:` manifest metadata maps those source skill directories into the plugin
guide readback surface. The target diagram above names the logical plugin guide
surface, not a second source loader root.

- KAS base skills are registered as Hermes plugin skills and are loadable through plugin-qualified names.
- Plugin skill names do not require copying into `~/.hermes/profiles/<profile>/skills/`.
- Color role manifests are source-controlled and deterministic.
- Profile-local project wrappers and overlays are intentionally small and may reference plugin-qualified base names.
- `kkachi-agent-skills install --project ... --suite-role ... --apply dry-run:<hash>` must materialize the default wrapper, default overlay, and overlay `references/legacy-delta-extract.md` alongside the selected project phase skills. The manifest records those deterministic composition files under `composition_files[]`, separate from role-selected `installed_skills[]`. After a manifest entry exists, install/repair must preserve overlay/reference content as project-local semantic tailoring; repair may adopt their current checksums into `composition_files[]`, while the generated wrapper may be repaired from the canonical template.
- Any different packaging path or namespace requires a SOT update before implementation is marked complete.

## 5. Color role model

The initial role projection follows the current Kkachi color boundaries.

| Color role | Example profile | Initial KAS plugin base selection | Boundary |
|---|---|---|---|
| Blue commander | `hwangchung` | full KAS phase/operation set, currently 24 KAS base skills | May plan, orchestrate, authorize implementation, request review, verify, and synthesize evidence within KAS/KAO authority. |
| Red reviewer | `hahuyeon` | `review`, `verify` | Safety, fail-closed, risk, evidence sufficiency, deterministic/runtime risk review. No implementation or backend-routing authority by skill presence. |
| Orange PM reviewer | `yeomong` | `review` | Operator value, scope, workflow clarity, acceptance criteria, report readability, approval/recovery clarity. No implementation or backend-routing authority by skill presence. |
| Gray scribe | `jingung` | `review`, `final-verify` | Decision trace, evidence paths, stale/candidate/accepted/runtime-approved distinctions, final-gate evidence support. `final-verify` does not replace Blue final authority. |
| Teal design reviewer | `goong` or project-registered Teal designer | `design-review`, `review` | Optional UI/UX design review only when project/task facts make Teal required. Not part of the mandatory Blue/Red/Orange/Gray non-UI baseline and no implementation/backend-routing authority by skill presence. |

KAH companion skills may have analogous color subsets, but this SOT owns KAS first. KAH mirror behavior must be updated through the KAH repo or a recorded KAS/KAH companion task rather than silently assuming this KAS SOT mutates KAH distribution.

Unknown, misspelled, or unregistered color roles fail closed. A project wrapper must not infer elevated capability from profile name alone; it must reference a registered role manifest and the registry/profile authority surface.

## 6. Project wrapper contract

A project wrapper is a profile-local discoverability and composition entrypoint for one approved profile/project target. It should be visible in ordinary profile-local skill discovery under `skills/<project>/<project>-wrapper/` so the model can learn the local operating rule even when plugin skills are not listed in the same flat skill inventory.

Minimum wrapper fields:

```yaml
---
name: kkachi-agent-helper-wrapper
description: Thin project-local wrapper for KAS plugin skills and the kkachi-agent-helper project overlay.
metadata:
  kas:
    kind: project_wrapper
    project: kkachi-agent-helper
    role: blue_commander
    role_manifest: kkachi-agent-skills:roles/blue.yaml
    plugin_namespace: kkachi-agent-skills
    overlay_skill: kkachi-agent-helper-overlay
    refresh_skill: kkachi-agent-skills:kkachi-agent-skills-overlay-refresh
    base_copy_policy: forbidden_by_default
---
```

The wrapper must:

1. instruct the agent to load plugin-qualified KAS base skills such as `skill_view("kkachi-agent-skills:plan")` when a matching phase skill is needed;
2. name the project, the explicit suite role, role manifest, and allowed role subset; generated wrappers must not hard-code Blue metadata when `--suite-role red_reviewer`, `orange_pm_reviewer`, or `gray_scribe` was selected;
3. point to the single default project overlay and explain that optional `kas-overlays/` entries are advanced extensions only;
4. record `refresh_skill: kkachi-agent-skills:kkachi-agent-skills-overlay-refresh` and a body reminder to load that guide before source-built overlay refresh or CLI repair/install proposals;
5. fail closed when the plugin base skill, role manifest, or project overlay is unavailable;
6. avoid embedding or duplicating full base skill bodies;
7. route source-built project overlay refresh work to `skill_view("kkachi-agent-skills:kkachi-agent-skills-overlay-refresh")` and identify it as LLM-assisted semantic porting, not CLI migration;
8. preserve Kkachi authority boundaries and explicit approval gates.

Wrappers may contain local phrasing for the profile's Korean name, role, channel, and project, but they must not become base-skill forks. Project-specific detail belongs in the paired project overlay and its `references/` files.

## 7. Project overlay contract

Project overlays are the approved place for project-specific KAS optimization under the plugin model. The default is one project overlay per profile/project target, not one copied overlay per base skill.

Canonical profile-relative layout:

```text
skills/<project>/<project>-overlay/SKILL.md
skills/<project>/<project>-overlay/references/*.md
```

Examples:

```text
skills/kkachi-agent-helper/kkachi-agent-helper-overlay/SKILL.md
skills/kkachi-agent-helper/kkachi-agent-helper-overlay/references/project-context.md
skills/doksuri/doksuri-overlay/SKILL.md
skills/doksuri/doksuri-overlay/references/verification.md
```

Optional advanced extension when a single project overlay is too coarse:

```text
skills/<project>/kas-overlays/<project>-<role>-<phase-or-base>-overlay/SKILL.md
```

Minimum overlay frontmatter:

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
    base_version: "<kas-version-or-commit>"
    refresh_skill: kkachi-agent-skills:kkachi-agent-skills-overlay-refresh
    authority_sources:
      - /path/to/project/sot.md
    references:
      - references/project-context.md
      - references/architecture.md
      - references/verification.md
      - references/authority-sources.md
    last_reviewed: "YYYY-MM-DD"
    promotion_candidate: false
---
```

Allowed overlay content:

- role-aware `applies_to` lists that match the selected suite role subset; Red, Orange, and Gray overlays must not advertise Blue-only implementation, backend-selection, orchestration, or optimization bases unless explicitly selected by their role registry;
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
3. **Project wrapper**: the profile-local project wrapper supplies project/role boundary, base skill selection, and overlay lookup policy.
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
4. `kkachi-agent-skills-overlay-refresh`: LLM-assisted workflow guide for refreshing a project overlay by temporarily archiving the current overlay as inactive legacy, rebuilding a clean current overlay, semantically porting only durable project-specific rules, verifying, and then removing the legacy archive.

These guide skills are KAS source skills, not profile-local per-project overlays. They may be exposed through the plugin and referenced by wrapper skills. If Hermes plugin skill discoverability remains weaker than flat profile-local discovery, the wrapper must explicitly instruct the agent to load these guide skills by plugin-qualified name.

## 10. No CLI update/migrate semantics and overlay refresh skill

The `kkachi-agent-skills` CLI must not expose public `update` or `migrate` commands. This includes plugin update, project-suite update/migration, profile-skill migration, and overlay migration. Those names imply semantic authority and safe merge behavior that a deterministic CLI cannot provide; semantic overlay refresh belongs in the `kkachi-agent-skills-overlay-refresh` skill.

Canonical command surface is limited to deterministic or diagnostic operations such as:

```text
kkachi-agent-skills list
kkachi-agent-skills install ...
kkachi-agent-skills doctor ...
kkachi-agent-skills repair ...
kkachi-agent-skills toolchain ...
kkachi-agent-skills workflow-create|workflow-promote|workflow-route|workflow-trigger ...
kkachi-agent-skills uninstall ...
kkachi-agent-skills version
```

Overlay refresh is intentionally handled by the `kkachi-agent-skills-overlay-refresh` skill rather than a CLI command. The refresh workflow is:

1. copy the active `skills/<project>/<project>-overlay/` to a temporary `skills/<project>/<project>-overlay-legacy/`;
2. mark the temporary legacy archive with `metadata.kas.kind: project_overlay_legacy`, `active: false`, and `superseded_by: <project>-overlay`;
3. create a fresh canonical `skills/<project>/<project>-overlay/` from the current `kkachi-agent-skills` template/SOT;
4. use LLM-assisted review to port only durable project-specific constraints, paths, architecture notes, verification rules, authority constraints, and recurring lessons from the legacy archive;
5. reject copied base text, stale workaround text, runtime/provider/auth/model settings, and fallback behavior;
6. run doctor/diff verification and request review when confidence is low;
7. after successful refresh and review, remove `skills/<project>/<project>-overlay-legacy/` because it is only a temporary comparison aid.

The CLI may provide read-only doctor evidence that helps this skill, but it must not decide which legacy prose to port and must not perform semantic overlay merges.

KAH has no implementation responsibility for overlay refresh. KAH may preserve `kkachi-agent-skills` overlay refresh evidence references in run artifacts or gate reports, but only as evidence preservation; KAH must not install, update, vendor, merge, or clean Hermes/`kkachi-agent-skills` skill content.

## 11. Doctor, manifest, and diagnostic semantics

KAS doctor has a read-only SKILL mode exposed as `kkachi-agent-skills doctor --plugin --repo <kas-repo> --json`. This mode is diagnostic-only: it reads explicit source plugin package evidence, role manifests, project wrappers, project overlays, copied base-suite evidence, personal skills, and unknown profile skills, then reports reason-coded JSON/human diagnostics without writing. `--repo` is required for plugin doctor mode so the source evidence root is explicit and embedded source cache materialization is not part of the no-write claim. It is incompatible with `--workflow-graph` and `--project-suite`; mixed doctor modes fail closed with `doctor_mode_ambiguous` and bounded next-action guidance.

The `doctor --plugin --json` packet includes KASREL readback fields for review readiness checks: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. These fields are non-secret readback evidence only; absent evidence is emitted as explicit empty arrays or `null` and does not authorize profile mutation, KAB activation, provider/model configuration, deleted-bundle fallback lookup, or install readiness by assumption.

Minimum diagnostics:

| Condition | Severity | Meaning |
|---|---|---|
| KAS plugin missing or disabled | error | Wrapper/overlay cannot safely resolve base skills. |
| Required plugin base skill missing | error | The selected KAS phase cannot run from canonical base. |
| Role manifest missing or unknown | error | The profile's color role cannot be safely projected. |
| Project wrapper missing for an active color profile | warning or error by rollout stage | Discoverability/composition is incomplete. |
| Wrapper role mismatches registry/profile authority | error | Role boundary is unsafe. |
| Profile-local full KAS base copy present | warning during migration, error after cutover | Legacy fork/shadow risk. |
| Overlay lacks plugin-qualified `applies_to` targets, or optional advanced overlay lacks `overlay_for`, or either references a non-plugin target | error | Cannot safely compose overlay. |
| Overlay copies too much base content | error or warning by threshold | Likely fork instead of delta. |
| Overlay shadows base skill name | error | Ambiguous skill identity. |
| Overlay base version is stale | warning | Needs review against current plugin base. |
| Overlay contains secrets/runtime settings | error | Must be removed before use. |
| Overlay improvement appears reusable | info | Promotion candidate for KAS guide/base review. |
| Public update/migrate CLI command present | error | `kkachi-agent-skills` must not expose public `update` or `migrate` commands; use `kkachi-agent-skills-overlay-refresh` for semantic overlay refresh. |
| Temporary overlay legacy archive still active or left after refresh | error | `overlay-legacy` is a comparison aid only; mark inactive and remove after successful refresh/review. |
| KAH appears to own `kkachi-agent-skills` plugin update/install/refresh | error | Boundary violation; KAH may preserve evidence refs only. |

Doctor output must distinguish `plugin_base`, `project_wrapper`, `project_overlay`, `legacy_copied_base_suite`, `personal_skill`, and `unknown_source` source classes. It must not delete, rewrite, or adopt profile-local skills without explicit dry-run/apply approval. Missing plugin, role, wrapper, or overlay evidence must produce diagnostics; profile-local copied suites are never a fallback for missing plugin/wrapper/overlay evidence.

## 12. Legacy profile material and overlay refresh policy

Existing profile-local copied `kkachi-agent-skills` / `kkachi-agent-helper` suites are legacy input, not the steady-state target. The public CLI must not provide a profile-skill migration command. Legacy material is handled through review, doctor evidence, and the `kkachi-agent-skills-overlay-refresh` skill when an active project overlay needs semantic refresh.

Legacy material may be classified during review as:

- `base_identical`: content matches plugin/source base and should not be copied forward into an overlay;
- `base_with_local_delta`: content differs and needs human/LLM-assisted semantic review before any durable project-specific part is ported;
- `project_overlay_candidate`: local content appears project-specific and may belong in `skills/<project>/<project>-overlay/` or, only when justified, an optional advanced `skills/<project>/kas-overlays/...` extension;
- `role_wrapper_candidate`: local content belongs in a thin project wrapper;
- `unknown_personal_skill`: preserve and request review;
- `kah_companion_surface`: belongs to `kkachi-agent-helper` or a paired `kkachi-agent-skills`/`kkachi-agent-helper` task, not silent `kkachi-agent-skills`-only cleanup.

The `kkachi-agent-skills-overlay-refresh` skill owns the safe semantic path for refreshing overlays. It may temporarily create `skills/<project>/<project>-overlay-legacy/` with `kind: project_overlay_legacy` and `active: false`; after the refreshed overlay is verified and accepted, the legacy archive must be removed. This SOT does not authorize broad copied-suite cleanup, automatic deletion, CLI migration, or use of copied profile skills as fallback base behavior.

Any real profile mutation requires explicit 주군 approval for exact profiles/projects, backup/vault evidence when destructive changes are possible, no-spillover scan, recovery instructions, and post-change doctor evidence. This SOT alone does not authorize deleting the currently observed copied profile skills.

## 13. Relationship to existing SOTs

- `docs/sot/project-specific-kas-install-contract.md` remains authority for the legacy/project-suite copied layout until SKILL implementation supersedes specific install paths through reviewed tasks. The SKILL target model treats copied project suites as a migration source and shifts future steady state to overlays.
- `docs/sot/role-aware-project-suite-contract.md` remains authority for role-aware subset safety. SKILL reuses those color boundaries for plugin role manifests and wrappers.
- `docs/sot/project-kas-sync-state.md` remains authority for project-specific state/sync until SKILL replaces or extends state files for overlay/version tracking.
- `docs/sot/kasrel-hermes-v016-provenance-contract.md` remains authority for release/provenance diagnostics. SKILL adds plugin-base/wrapper/overlay source classes that must integrate with provenance reporting.
- KAH and KAB behavior must be updated through their own docs/tasks when companion changes are needed. The expected default for this SKILL change is no KAH implementation change: KAH may preserve KAS lifecycle evidence references in artifacts/gates, but KAS must not imply KAH owns plugin packaging, update, wrapper/overlay composition, cleanup, or KAB runtime support from plugin/wrapper docs alone.

## 14. Required roadmap sequence

This amendment supersedes the earlier SKILL-006 pilot assumption that `kas-overlays/` was the default canonical project overlay path. `kas-overlays/` is retained only as an optional advanced extension after review.

`SKILL` is a planning and migration epic. It should start with SOT acceptance and then proceed in small PR-candidate tasks.

1. `SKILL-001` — accept this plugin-wrapper-overlay SOT, register the docs/roadmap references, and record color review / responsible approval.
2. `SKILL-002` — implement `kkachi-agent-skills` Hermes plugin package registration for canonical base skills, deterministic source role manifests, and bounded guide metadata/readback exposure; prove plugin-qualified skill loading in non-mutating smoke tests. SKILL-002 must not expose public `update` or `migrate` CLI surfaces and does not author thin wrapper templates, project overlay authoring guides, composition-guide bodies, overlay-refresh workflow bodies, or doctor-guide bodies.
3. `SKILL-003` — add thin project wrapper templates and guide skills (`kas-project-overlay-guide`, `kas-overlay-compose-guide`, `kas-overlay-doctor-guide`, `kkachi-agent-skills-overlay-refresh`); verify wrapper instructions do not copy base bodies and overlay refresh is skill-mediated rather than CLI-mediated.
4. `SKILL-004` — implement read-only SKILL doctor diagnostics for plugin/wrapper/overlay/legacy-copy provenance and invalid overlay detection.
5. `SKILL-005` — remove public update/migrate CLI surfaces and keep legacy copied-suite handling as doctor/review evidence only; no deletion.
6. `SKILL-006` — run one approved pilot overlay refresh for exact profile/project targets through `kkachi-agent-skills-overlay-refresh` after explicit 주군 approval, backup/no-spillover evidence when needed, and recovery plan.

Later tasks may merge SKILL behavior into `install`, `doctor`, `sync-project-kas`, or replacement commands only after the earlier SKILL tasks provide evidence. Public `update` and `migrate` surfaces remain removed unless a future accepted SOT explicitly reinstates a non-semantic deterministic surface under a non-ambiguous name.

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
- The public CLI exposes no `update` or `migrate` commands; semantic overlay refresh is handled by `kkachi-agent-skills-overlay-refresh`, not deterministic CLI migration.
- Blue/Red/Orange/Gray role manifests are deterministic, source-controlled, and tested.
- Thin project wrapper templates exist and are small enough to be discoverability/composition surfaces rather than base forks.
- Project overlay frontmatter and layout are documented and tested with valid/invalid fixtures.
- Doctor reports plugin base, wrapper, overlay, legacy copy, personal skill, and unknown source classes distinctly.
- Doctor reports public update/migrate command exposure and KAH-boundary violations distinctly.
- Legacy copied profile-local `kkachi-agent-skills` / `kkachi-agent-helper` material is review input only and is not a CLI migration target.
- Cleanup/apply requires explicit 주군 approval, backup, no-spillover scan, and recovery evidence.
- `docs/README.md`, `docs/kkachi-docs-map.yaml`, and `docs/roadmap.md` register this SOT and preserve the no-implementation boundary.
- Red/Orange/Gray review accepts the safety, operator clarity, and evidence trace before broad rollout.
