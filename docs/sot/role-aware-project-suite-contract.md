# Role-aware project KAS suite contract

Date: 2026-06-13
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / KASROLE-001 color review evidence recorded in Kanban cards Red `t_bfad026c`, Orange `t_21f593c9`, Gray `t_acc68322`, Blue synthesis `t_a971aa92`; source implementation completion is recorded in roadmap evidence and operational cleanup remains approval-gated
Status: KASROLE-001..004 source behavior completed through commits `cda451c`, `5082663`, `0ba3f82`, and `7324d9c`; real Red/Orange/Gray profile cleanup/apply remains blocked until explicit 주군 operational approval and must not be implied by this SOT
Authority level: KAS-side source of truth for role-aware project suite selection, subset manifests, doctor/repair semantics, and pre-WFLOW operational cleanup
Scope: `kkachi-agent-skills` docs, registries, templates, CLI planning, profile-suite manifest semantics, and KAS-owned repair guidance. No Hermes profile mutation, installed skill deletion, KAH deterministic state mutation, KAB runtime activation, auth/token/gateway/provider/model change, or automatic rollout is authorized by this document alone.
Related docs: `docs/sot/project-specific-kas-install-contract.md`, `docs/sot/project-kas-sync-state.md`, `docs/sot/task-dag-workflow-contract.md`, `docs/sot/token-economy-and-agent-instruction-contract.md`, `docs/sot/kas-cli-contract.md`, `docs/roadmap.md`, KAH `docs/sot/task-dag-state-machine.md`
Evidence/source paths:
- 주군 direction in 17번째 지구 Discord `#kah` thread `1515219219002818610` on 2026-06-13: before local DAGSM/WFLOW development, do not proceed with Red/Orange/Gray profiles that received full 18-skill `kkachi-agent-helper` project suites; create a KAS SOT and fix role-aware install/doctor/repair first.
- Observed current KAS behavior: `kas-default-project-suite` is derived from `skill-pack.yaml` and installs all 18 project-prefixed skills for every target profile. This matches Blue commander needs but over-provisions reviewer/scribe profiles.
- KASROLE-001 color discussion on 2026-06-13: Red `t_bfad026c` accepted the safety/fail-closed direction, Orange `t_21f593c9` conditionally accepted with UX/role-label refinements, Gray `t_acc68322` conditionally accepted with SOT/evidence trace refinements, and Blue synthesis `t_a971aa92` directed minor doc/design fixes before KASROLE-002 planning.
- KASROLE source implementation closure on 2026-06-13: commits `5082663`, `0ba3f82`, and `7324d9c` implemented role-aware install/manifest, doctor diagnostics, and approval-gated repair/prune source behavior. Read-only doctor evidence still reports `hahuyeon`, `yeomong`, and `jingung` KAH development profile suites as fail-closed legacy/full-suite states with missing `suite_role`; no real profile cleanup/apply is authorized by source closure.

## 1. Decision summary

KAS must distinguish **full commander project suites** from **role-aware project subset suites** before KAH `DAGSM` and KAS `WFLOW` implementation work proceeds beyond planning. The existing full-suite install behavior remains valid for a Blue commander profile such as 황충, but it is not a valid default for Red, Orange, or Gray review/scribe profiles.

The immediate policy is:

1. Blue commander profile may hold the full project KAS suite.
2. Red, Orange, and Gray profiles must receive only the skills required for the nodes/roles they perform.
3. Teal is optional and is not part of the mandatory Blue/Red/Orange/Gray project-suite baseline. Teal workflow gates are required only when `project_has_teal_lane && ui_ux_change` is true.
4. Installing a Teal role suite must not make non-UI Kkachi source work Teal-required; non-UI work still records deterministic skip evidence.
5. A profile-local overlay saying "do not use implement" is not sufficient when the unauthorized skill is still installed and discoverable.
6. KAS doctor must treat the selected role subset as the authority for that profile; missing unselected skills are not errors, and extra out-of-role skills are diagnostics.
7. Approved repair/prune must exist before cleaning existing over-installed role profiles.
8. Local KAS/KAH DAGSM/WFLOW development must not rely on over-installed Red/Orange/Gray/Teal profile suites as the normal operating baseline.

## 2. Problem statement

`docs/sot/project-specific-kas-install-contract.md` correctly requires project-specific installed skill identities such as `skills/<project>/<project>-plan/SKILL.md`. However, the current default project source suite is the complete `skill-pack.yaml` phase/operation set. That set contains 18 skills and was designed as a full operational KAS pack, not as a reviewer-only projection.

The result is safe enough for a commander who orchestrates all phases, but wrong for color reviewers:

- 하후연 / Red reviewer should not have implementation, backend selection, prompt composition, optimization, or orchestration authority just because the project suite installer copied those skills.
- 여몽 / Orange PM/operator-value reviewer should not inherit code mutation or backend routing phases.
- 진궁 / Gray scribe should not inherit implementation authority or broad command phases; evidence/docs integrity is his review lane.

Role overlays can reduce behavioral risk but do not fix the source-of-truth problem. The installed skill inventory and manifest must encode the role boundary directly.

## 3. Ownership boundary

| Layer | Owns | Must not own |
|---|---|---|
| KAS | role-aware project suite registry, selected skill projection, subset manifest vocabulary, doctor/repair/prune semantics, dry-run/apply hash binding, role-to-skill policy, node contract required-skill policy | KAH node state mutation, hidden fallback role assignment, live Hermes profile mutation without approval, auth/token/gateway/provider/model changes |
| KAH | deterministic project/run/workflow state, DAG validation, node FSM, ready-node calculation, evidence gates, diagnostics/events once DAGSM exists | KAS role suitability, skill selection policy, subjective reviewer authority, profile skill inventory policy |
| KAB | backend/session/runtime evidence when a node dispatches through a selected backend lane | project suite selection, color-role policy, KAH state authority |
| Kanban | long-lived team-member assignment, review cards, dependencies, evidence routing | replacement for KAS profile inventory or KAH DAG state |
| KAO / 황충 | Blue synthesis, approval routing, role-bound rollout decision, final reporting | silent role over-provisioning or unapproved profile pruning |

## 4. Role suite model

KAS must support at least these suite roles for project-specific installed suites.

| `suite_role` | Intended profile | Required selected skills | Notes |
|---|---|---|---|
| `blue_commander` | 황충 / Blue commander | full project suite from `skill-pack.yaml` | This is the current full-source-suite behavior, but it must be explicit rather than assumed for every profile. Human output should label it as Blue commander / full project suite. |
| `red_reviewer` | 하후연 / Red reviewer | review and verification-focused project skills | Initial minimal set: `<project>-review`, `<project>-verify`. No implement/backend-select/prompt-compose/optimize/orchestrate. Human output should label it as Red safety/fail-closed reviewer subset. |
| `orange_pm_reviewer` | 여몽 / Orange PM/operator-value reviewer | operator-value review and assigned final acceptance support | Initial minimal set: `<project>-review`; optionally `<project>-final-verify` only when Orange is assigned an operator-value final gate. Human output should label it as Orange operator-value reviewer subset. |
| `gray_scribe` | 진궁 / Gray scribe | evidence/docs integrity review and assigned final-gate evidence support | Initial minimal set: `<project>-review`, `<project>-final-verify`; `final-verify` here means evidence/final-gate support when assigned, not broad final approval authority. A future Gray-specific evidence-integrity skill is preferred over reusing docs-update as a write-capable scribe surface. |
| `teal_design_reviewer` | Goong / Teal lead or project-registered Teal designer | optional UI/UX design review when the project has a Teal lane and the task has UI/UX scope | Initial minimal set: `<project>-design-review`, `<project>-review`. This role supplies `DESIGN_PLAN_GATE` and `DESIGN_FIDELITY_REVIEW` evidence only when `teal_required=true`; it is not part of the mandatory non-UI Kkachi source-work baseline. |

The selected skill list is a policy projection, not a recommendation. A role subset install must not install non-selected phase skills unless the operator explicitly selects a different role or an approved custom role. Stable JSON values stay English (`blue_commander`, `red_reviewer`, `orange_pm_reviewer`, `gray_scribe`, `teal_design_reviewer`), while human output must include an operator-readable role summary so 주군 can inspect dry-runs without decoding enum names.

## 5. Registry requirements

KAS must introduce a machine-readable role suite registry or equivalent formal source. The initial canonical path for KASROLE-002 is `registries/project-suite-roles.yaml`, owned by the KAS workflow/policy layer; a different path requires a recorded SOT update before implementation. The registry must be deterministic and hash-bound by install/repair plans. Unknown, misspelled, or unregistered `suite_role` values must fail closed (`ok:false`, error diagnostic, no writes) rather than falling back to a full suite or guessed role. Registry version `role-aware-project-suite/v1` must be bumped or explicitly compatibility-reviewed when role semantics, selected/forbidden skill policy, or manifest interpretation changes.

Minimum fields:

```yaml
version: role-aware-project-suite/v1
roles:
  blue_commander:
    source: skill-pack.yaml
    display_label: "Blue commander / full project suite"
    selection_mode: full_source_suite
    required_source_skills: "*"
  red_reviewer:
    display_label: "Red safety/fail-closed reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
      - kkachi-verify
    forbidden_source_skills:
      - kkachi-implement
      - kkachi-backend-select
      - kkachi-prompt-compose
      - kkachi-optimize
      - kkachi-orchestrate
  orange_pm_reviewer:
    display_label: "Orange operator-value reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
    optional_source_skills:
      - kkachi-final-verify
    forbidden_source_skills:
      - kkachi-implement
      - kkachi-backend-select
      - kkachi-prompt-compose
      - kkachi-optimize
      - kkachi-orchestrate
  gray_scribe:
    display_label: "Gray evidence/scribe reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-review
      - kkachi-final-verify
    forbidden_source_skills:
      - kkachi-implement
      - kkachi-backend-select
      - kkachi-prompt-compose
      - kkachi-optimize
      - kkachi-orchestrate
      - kkachi-docs-update
    future_preferred_skills:
      - kkachi-evidence-review
      - kkachi-scribe-review
  teal_design_reviewer:
    display_label: "Teal optional UI/UX design reviewer subset"
    selection_mode: explicit_source_subset
    required_source_skills:
      - kkachi-design-review
      - kkachi-review
    forbidden_source_skills:
      - kkachi-implement
      - kkachi-backend-select
      - kkachi-prompt-compose
      - kkachi-optimize
      - kkachi-orchestrate
      - kkachi-docs-update
      - kkachi-final-verify
```

Skill source ids must resolve against the same source pack inventory used by `kas-default-project-suite`. Explicit `forbidden_source_skills` lists document high-risk out-of-role skills; every source skill not selected by the role remains excluded even if it is not repeated in the forbidden list.

## 6. CLI contract

KAS must add a role-aware project install surface before using KAS/KAH local development profiles for DAGSM/WFLOW execution.

Required behavior: `--suite-role` is mandatory for role-aware project suite install/repair. KAS must not silently default Red/Orange/Gray profiles to a full source suite. If Blue compatibility keeps an older full-suite command, it must either require `--suite-role blue_commander` or print a fail-closed diagnostic instructing the operator to rerun with the explicit role.

Required public shape:

```bash
kkachi-agent-skills install --profile <profile> --project <project> --suite-role <role> --dry-run [--json]
kkachi-agent-skills install --profile <profile> --project <project> --suite-role <role> --apply dry-run:sha256:<hash> [--json]
```

Equivalent existing aliases such as `install-project-kas` may be kept, but normal operator guidance should use the public `install` lifecycle verb if that is the current KAS UX convention.

Dry-run output must include:

- target profile and project;
- selected `suite_role` plus operator-readable `display_label`;
- selected source skills and rendered installed skills;
- selected/excluded skill counts in the compact human summary;
- excluded source skills and reason, including `outside_suite_role`;
- role registry checksum;
- source suite checksum;
- changed paths;
- conflicts/diagnostics;
- no-write evidence;
- plan hash binding all of the above;
- a copy/paste next command that preserves the explicit `--suite-role`.

Apply must recompute the current dry-run and fail closed before writes unless the approval hash exactly matches. Malformed approval evidence, missing or unknown `suite_role`, role/source mismatch, or any attempt to infer a reviewer role from profile name alone must stop before writes.

## 7. Manifest vocabulary

`project_suites[]` entries must distinguish full suites from role subset suites.

Minimum compatible extension:

```json
{
  "kind": "kas_project_skill_manifest",
  "project": "kkachi-agent-helper",
  "suite_mode": "role_subset",
  "suite_role": "red_reviewer",
  "role_registry": {
    "path": "registries/project-suite-roles.yaml",
    "version": "role-aware-project-suite/v1",
    "checksum": "sha256:<hex>"
  },
  "source_pack": {
    "id": "kas-default-project-suite",
    "checksum": "sha256:<hex>",
    "formal_registry": "skill-pack.yaml"
  },
  "selected_skills": [
    "kkachi-agent-helper-review",
    "kkachi-agent-helper-verify"
  ],
  "excluded_skills": [
    {
      "installed_skill": "kkachi-agent-helper-implement",
      "reason": "outside_suite_role"
    }
  ],
  "installed_skills": [],
  "drift_policy": "role_subset_expected",
  "semantic_adaptation_claimed": false
}
```

For `suite_mode: full`, the selected skill list may be all source skills and `suite_role` should be `blue_commander` unless a later SOT defines another full-suite role.

## 8. Doctor semantics

`doctor --project-suite` must become role-aware.

| Condition | Required severity | Meaning |
|---|---|---|
| selected skill present and checksum matches manifest/trusted tailoring | pass | Required role skill is installed and trusted. |
| selected skill missing | error | The profile cannot perform its declared role. |
| selected skill checksum mismatch without trusted tailoring | error | Manual review or approved repair required. |
| unselected out-of-role skill present and KAS-managed by the same project suite | error for Red/Orange/Gray; warning for Blue unless custom policy says otherwise | Role boundary is over-provisioned and must be pruned before role use. |
| unselected unknown personal skill exists under the project directory | warning or error depending on shadowing | Must not be silently adopted or deleted. Escalate to error if it shadows selected role skills or creates identity ambiguity. |
| unknown or unregistered `suite_role` requested or recorded | error | KAS cannot safely determine the selected skill projection. Fail closed before install, doctor, repair, or prune writes. |
| manifest lacks `suite_role` for a project suite installed after this contract is implemented | error | KAS cannot determine whether missing full-suite skills are real defects. |
| legacy full suite in Red/Orange/Gray profile | error or blocked diagnostic | Existing over-install must be repaired/pruned before being treated as a healthy role profile. |

Doctor must not treat missing non-selected full-suite skills as errors for a role subset.

## 9. Repair and prune contract

KAS must provide approval-gated repair/prune before cleaning existing over-installed role profiles.

Required behavior:

```bash
kkachi-agent-skills repair --profile <profile> --project <project> --suite-role <role> --prune-extra --dry-run [--json]
kkachi-agent-skills repair --profile <profile> --project <project> --suite-role <role> --prune-extra --apply dry-run:sha256:<hash> [--json]
```

Dry-run must list:

- compact keep/create/update/remove counts before detailed paths;
- skills to keep;
- skills to create/update;
- skills to remove because they are outside the selected role;
- whether each removal is manifest-tracked and KAS-managed;
- backup/vault path plan;
- recovery steps and rollback limits;
- manifest rewrite plan, with manifest write last;
- no-spillover scan evidence covering unrelated profiles/projects and unknown personal skills;
- no-write evidence;
- approval hash.

Apply requires explicit 주군 operational approval for the exact target profiles/projects before any profile mutation or prune. The implementation plan for KASROLE-004 must name the approval owner, affected profiles, affected projects, backup/vault root, no-spillover scan command/evidence, and recovery instructions. Apply must:

1. recompute and match the dry-run hash;
2. back up every removed/changed skill and the previous manifest;
3. remove only manifest-tracked KAS-managed project suite files or explicitly approved role-suite files;
4. preserve unrelated personal/profile-local skills;
5. write the manifest last;
6. report recovery instructions;
7. never mutate profile config, SOUL, gateway, auth, tokens, providers, models, KAH state, or KAB runtime state.

## 10. Immediate KAH/DAGSM development gate

Before beginning local KAH `DAGSM-001` or KAS `WFLOW-002` implementation work with color-profile review support, the KAS/KAH development environment must satisfy one of these states:

1. Preferred: KASROLE implementation exists, Red/Orange/Gray profiles are repaired to role subsets, and doctor reports their role suites healthy.
2. Temporary fail-closed alternative: Red/Orange/Gray project KAS suites are removed entirely, and reviews are routed only by Kanban/profile role prompt without installed KAH project suite skills. This makes the profile role prompt the temporary behavior authority and must be recorded as a bounded transition state, not as installed-suite health.
3. Not allowed as normal baseline: Red/Orange/Gray retain full 18-skill `kkachi-agent-helper` project suites with only overlay text limiting behavior.

The current over-installed state may be kept only as a backed-up transition state while KASROLE is being implemented. It must not be reported as the target healthy state.

## 11. Relationship to DAGSM and WFLOW

This contract is intentionally before WFLOW/DAGSM implementation, but it does not replace the WFLOW SOT.

- `docs/sot/task-dag-workflow-contract.md` already defines future node-level owner role and required-skill contracts.
- KASROLE provides the prerequisite installed-profile inventory semantics so reviewer/scribe nodes do not start from over-provisioned profiles.
- Later WFLOW work may derive profile subsets from DAG node contracts, but the first implementation should use explicit `suite_role` registry selection to unblock safe local development.
- KAH DAGSM must not decide role suitability or skill projection; KAH may later validate that node evidence references a declared role/profile and required artifacts, but KAS owns which skills belong to that role.

## 12. Required roadmap sequence

KASROLE is a pre-WFLOW/DAGSM corrective epic. It gates local development use of color profiles for `DAGSM` and WFLOW implementation.

1. `KASROLE-001` — accept this SOT and add role suite registry planning entries.
2. `KASROLE-002` — implement role-aware project install dry-run/apply and manifest vocabulary, including explicit `--suite-role`, unknown-role fail-closed diagnostics, operator-readable role labels, and role registry checksum binding.
3. `KASROLE-003` — implement role-aware doctor diagnostics for full vs subset suites, including fixtures for Blue full suite, Red/Orange/Gray subsets, missing selected skills, out-of-role KAS-managed extras, unknown personal skills, legacy full-suite profiles, and manifests missing or carrying unknown `suite_role`.
4. `KASROLE-004` — implement approval-gated repair/prune and use it to clean the approved KAH development color profiles only after explicit 주군 operational approval, backup/vault path evidence, no-spillover scan evidence, and recovery instructions are recorded.
5. Resume `DAGSM-001` / WFLOW implementation only after the preferred or temporary fail-closed development gate in section 10 is satisfied.

## 13. Non-goals and deferrals

- No automatic cleanup of existing Red/Orange/Gray profiles from this SOT alone.
- No profile config, SOUL, gateway, auth, token, provider, or model mutation.
- No KAH deterministic state mutation.
- No KAB Stage 2/3 activation.
- No dynamic derivation from DAGSM node contracts in the first implementation pass.
- No custom role language model selection or provider routing.
- No broad rollout across all Kkachi profiles.
- No deletion of personal, unknown, or unmanifested skills without a separate explicit approval path.

## 14. Acceptance gates

Before KAS claims role-aware project suite support:

- The role registry is machine-readable at `registries/project-suite-roles.yaml`, documented, checksum-bound, and covered by tests.
- Install dry-run/apply hash-binds role selection, selected/excluded skills, role registry checksum, source suite checksum, changed paths, conflicts/diagnostics, no-write evidence, and fails closed for unknown/unregistered roles.
- Manifest entries distinguish `suite_mode: full` and `suite_mode: role_subset`, preserve stable `suite_role` JSON values, and expose operator-readable role labels in human output.
- Doctor treats missing unselected skills as healthy for role subsets and flags out-of-role extras for Red/Orange/Gray and optional `teal_design_reviewer` suites.
- Repair/prune is dry-run-first, approval-hash-bound, backup-backed, manifest-tracked, no-spillover-scanned, and reports recovery instructions before any apply.
- Existing KAH development color profiles can be moved from the over-installed transition state to a healthy role-subset or explicit no-suite state with evidence.
- Red/Orange/Gray review gates, plus Teal review when a UI/UX Teal lane is in scope, have no unresolved blockers for role safety, operator clarity, and evidence trace.
