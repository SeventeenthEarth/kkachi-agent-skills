# KASREL Hermes v0.16 skill provenance and dependency audit contract

Date: 2026-06-08
Owner: KAS workflow/policy layer
Confirming role: responsible approver / color-role governance evidence record
Status: accepted SOT for KASREL-001; docs/spec-only contract; not implementation authorization and not installed-profile mutation approval
Authority level: source of truth for KAS release-compatibility provenance and dependency audit fields plus KASREL guidance evidence gates
Scope: `kkachi-agent-skills` KAS provenance/dependency contract for `list --json`, `install --dry-run --json`, and `doctor --json` output plus KASREL-004 guidance evidence gates
Related docs: `docs/roadmap.md`, `docs/sot/kas-cli-contract.md`, `docs/sot/project-kas-sync-state.md`, `docs/README.md`
Evidence handles: KASREL color review loop round 1 `red=t_0c91b4cf`, `orange=t_cedd82fc`, `gray=t_983ce96e`; round 2 final consensus `red=t_8a90c4ef`, `orange=t_44fa2173`, `gray=t_3e1024bf`; blue final synthesis recorded on the review cards

## 1. Authority, status, and non-authorization boundaries

This document promotes the KASREL-001 planning record into the accepted KAS release-compatibility contract for Hermes v0.16 skill provenance and dependency audit work.

It is authoritative for:

- the vocabulary KAS must use when reporting skill provenance;
- the canonical source classes KAS must distinguish;
- the evidence inputs and precedence rules KAS must use for provenance classification;
- deleted-bundle, shadowing, ambiguity, and dependency diagnostics; and
- JSON fields for `list --json`, `install --dry-run --json`, and `doctor --json`.

This document does **not** authorize:

- profile install, profile repair, approved copy install, or other profile mutation;
- binary rebuild/install or installed-binary parity repair;
- auth, token, gateway, provider, model, credential, or profile configuration mutation;
- KAB activation, KAB runtime/session control, or backend execution claims;
- write-capable KASUPD sync or project-specific KAS update behavior;
- new KASREL implementation beyond separately approved/completed task evidence; or
- fallback lookup paths for deleted Hermes bundle skills.

Implementation and guidance tasks must preserve the KAS/KAH/KAB split: KAS owns policy, contracts, skill-pack provenance, and user-facing guidance; KAH owns deterministic project state, artifacts, schemas, events, locks, diagnostics, and gates; KAB owns backend runtime/session control and bridge execution evidence.

## 2. Terminology

| Term | Contract meaning |
|---|---|
| Skill provenance | The path-, manifest-, and metadata-backed explanation of where an effective skill came from and why KAS assigned its source class. |
| Effective skill | The skill instance that Hermes would resolve for a skill name after profile, external directory, hub, and bundle ordering are considered. |
| Source class | One of the canonical KAS classes in this contract that describes the effective skill's source category. |
| Source evidence | Non-secret evidence used to justify a source class, such as a profile manifest entry, configured external directory, hub metadata record, bundle root match, checksum state, or path match. |
| Skill dependency | A named KAS, external, or Hermes skill that another skill declares or that KAS derives from skill guidance as a skill-to-skill requirement. |
| Command-surface dependency | A required command, CLI, runtime, or evidence surface owned by Hermes, KAH, Kanban, KAB, or another tool. Command-surface dependencies are not fake skills. |
| Deleted-bundle reference | A reference to a Hermes bundle skill that is no longer present in the effective Hermes bundle for the target release. |
| Shadowing/conflict | A same-name or ambiguous-path condition where one source can mask another effective skill candidate, or where KAS cannot prove the intended source class safely. |

## 3. Canonical source classes

KASREL JSON output must use only these source classes unless a later accepted SOT revises this list.

| Source class | Meaning |
|---|---|
| `bundle_builtin` | Skill is supplied by the effective Hermes bundle root for the target Hermes release. |
| `hub_installed` | Skill is installed through Hermes hub metadata or lock/manifest evidence. |
| `ops_external` | Skill is found in configured operator/external directories such as `skills.external_dirs`, not in the target profile's local skill install path. |
| `profile_personal` | Skill is in the target profile-local `skills/` tree and is not owned by the KAS profile manifest. |
| `kas_managed_profile` | Skill is in the target profile-local `skills/` tree and matches KAS profile manifest ownership evidence. |
| `unknown_or_unclassified` | KAS cannot classify safely because evidence is missing, conflicting, ambiguous, unreadable, or path resolution fails closed. |

`unknown_or_unclassified` is a fail-closed class. It must not be treated as permission to install, repair, substitute, or assume a safer source.

## 4. Resolution inputs

KAS provenance classification must use source evidence from these inputs when available:

1. target profile configuration, including configured profile-local skill paths and relevant skill-loading settings;
2. effective Hermes bundle paths for the target Hermes release;
3. Hermes hub lock, manifest, or installed-skill metadata;
4. KAS profile manifest data, including managed pack identity, install paths, and checksum state;
5. configured `skills.external_dirs` and equivalent external/ops skill directory settings; and
6. profile-local `skills/` entries.

Hermes CLI source labels are advisory only. In particular, a Hermes CLI `local` label is not sufficient to distinguish `profile_personal`, `kas_managed_profile`, and `ops_external`; KAS must use path, manifest, and configuration evidence before assigning those classes.

## 5. Classification precedence

When implementation classifies an effective skill, it must apply these precedence and diagnostic rules:

1. KAS profile manifest ownership plus profile-local path evidence identifies `kas_managed_profile`, subject to checksum and manifest-state diagnostics.
2. Profile-local skills without KAS manifest ownership are `profile_personal`.
3. Configured external/ops directories identify `ops_external` when the resolved path is outside the target profile-local skill tree and matches configured external directory evidence.
4. Hub metadata identifies `hub_installed` when the resolved path/identity is backed by hub lock or manifest data.
5. Effective Hermes bundle roots identify `bundle_builtin` when the path is inside the release bundle and no higher-precedence profile/external/hub source is the effective skill.
6. Missing, conflicting, symlink-escaping, unreadable, or insufficient evidence yields `unknown_or_unclassified` plus a diagnostic.

Precedence explains classification; it does not authorize mutation. A write-capable install or repair task still requires the existing dry-run and approval gate from `docs/sot/kas-cli-contract.md`.

## 6. Deleted-bundle references

Deleted-bundle references are cleanup/fail-closed diagnostics only.

KAS must report a deleted-bundle reference when current KAS guidance, a manifest, or a dependency declaration still points to a Hermes bundle skill that is absent from the effective bundle for the target release. The diagnostic should identify the stale reference and the evidence source that created the claim.

KAS must not:

- look up stale bundle paths for deleted skills;
- substitute a different bundle, hub, external, profile, or KAS-managed skill;
- invent fallback candidates;
- downgrade the diagnostic to a warning when the missing deleted-bundle skill is required; or
- use deleted-bundle handling to bypass the source-class or dependency audit.

## 7. Shadowing and ambiguity diagnostics

KAS provenance output must report shadowing or ambiguity when evidence shows any of these conditions:

- a profile-local skill masks a same-named external/ops, hub, or bundle skill;
- a KAS-managed profile copy masks a non-KAS same-named skill;
- multiple configured external directories contain the same skill name and KAS cannot prove the effective path safely;
- a path resolves outside the expected profile, external, hub, or bundle root; or
- symlink, case-sensitivity, unreadable path, or manifest mismatch evidence makes the effective source ambiguous.

Ambiguity must fail closed as `unknown_or_unclassified` unless implementation has enough non-secret evidence to prove a canonical source class. Diagnostics should identify the ambiguous source evidence without leaking secrets.

## 8. Dependency taxonomy

KASREL separates skill dependencies from command-surface dependencies.

### 8.1 Skill dependencies

`skill_dependencies` describe named skills. They may include KAS pack names, external skill names, or Hermes skill names when a skill truly depends on another skill as a skill. A missing required skill dependency is a provenance/dependency diagnostic and may fail readiness checks closed.

Skill dependency records should identify declaration evidence such as the declaring file, manifest field, or derived guidance source when available.

### 8.2 Command-surface dependencies

`command_surface_dependencies` describe commands, CLIs, runtimes, or evidence surfaces. They may include Hermes skill-listing behavior, KAH phase/artifact/event/gate commands, Kanban CLI behavior, KAB backend/session evidence when a lane actually requires KAB, and other non-skill runtime surfaces.

Command surfaces must not be represented as fake skill dependencies. A missing command surface is reported as a command-surface diagnostic owned by its actual layer. KAB evidence is a command/runtime dependency only when the selected lane requires KAB or claims backend runtime evidence; KASREL does not activate KAB.

## 9. KASREL-004 guidance evidence gate

Before KAS guidance claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills, it must cite current non-secret KASREL provenance/dependency evidence from `kkachi-agent-skills list --json`, `kkachi-agent-skills install --dry-run --json`, or `kkachi-agent-skills doctor --json` as appropriate. Required readback fields are `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, and `deleted_bundle_reference` / `deleted_bundle_diagnostics` when present. Missing, ambiguous, or stale evidence fails closed to `not ready`, `blocked`, or `needs audit`; do not upgrade it to confidence by assumption.

Deleted-bundle references are cleanup/fail-closed diagnostics only. Do not look up stale bundle paths, substitute another bundle/hub/external/profile/KAS-managed skill, invent fallback candidates, or downgrade a required missing deleted-bundle skill to a warning. Report the `deleted_bundle_reference` or deleted-bundle diagnostic with evidence and treat it as blocking readiness when required.

Keep dependency taxonomy explicit: `skill_dependencies` are named skills only, while Hermes/KAH/Kanban/KAB commands, CLIs, runtimes, and evidence surfaces are `command_surface_dependencies`, not fake skill dependencies. KAS owns policy, skill/process guidance, skill-pack provenance/dependency evidence, source-class interpretation, and user-facing readiness language. KAH owns deterministic project state, run artifacts, schemas, events, locks, diagnostics, gates, and project doctor/gate evidence. KAB owns backend runtime/session control and bridge execution evidence. KASREL evidence by itself must not start KAB, require KAB preflight, or claim bridge/backend runtime completion.

Active KAS guidance surfaces may reference this shared gate instead of duplicating the full text, but they must make the reference mandatory before any confidence/readiness/review/verification/final-completion claim.

## 10. KASREL JSON fields

The fields in this section are the KASREL contract fields implemented by KASREL-002 and KASREL-003. They remain compatibility-bound by this SOT and must not authorize profile mutation, KAB activation, or fallback behavior by themselves.

Concrete JSON values and examples below are illustrative only. They describe field shape and vocabulary, not a required current value, profile, path, pack id, command result, or mutation.

### 10.1 Common top-level fields

`list --json`, `install --dry-run --json`, and `doctor --json` outputs include a provenance contract version and inventory summary when provenance auditing is in scope:

```json
{
  "provenance_contract_version": "kasrel-001.v1",
  "source_inventory_summary": {
    "counts_by_source_class": {
      "bundle_builtin": 0,
      "hub_installed": 0,
      "ops_external": 0,
      "profile_personal": 0,
      "kas_managed_profile": 0,
      "unknown_or_unclassified": 0
    },
    "ambiguous_count": 0,
    "deleted_bundle_reference_count": 0,
    "shadowing_conflict_count": 0
  }
}
```

### 10.2 Per-skill or per-pack fields

Per-skill or per-pack records include provenance, diagnostics, shadowing, deleted-bundle, and dependency fields where relevant:

```json
{
  "skill_id": "illustrative-skill",
  "pack_id": "illustrative-pack",
  "effective_path": "skills/illustrative-skill",
  "source_class": "kas_managed_profile",
  "source_class_evidence": [
    {
      "kind": "kas_manifest",
      "path": ".kas/skill-pack-manifest.json",
      "state": "matched"
    }
  ],
  "provenance_state": "classified",
  "managed_by_kas": true,
  "checksum_state": "matched",
  "shadowing": [],
  "deleted_bundle_reference": null,
  "diagnostics": [],
  "skill_dependencies": [],
  "command_surface_dependencies": []
}
```

### 10.3 Dependency fields

Dependency records preserve the taxonomy split:

```json
{
  "skill_dependencies": [
    {
      "name": "illustrative-skill-dependency",
      "kind": "kas_pack",
      "required": true,
      "declared_by": "SKILL.md",
      "resolution_state": "resolved",
      "resolved_source_class": "kas_managed_profile"
    }
  ],
  "command_surface_dependencies": [
    {
      "surface": "kkachi-agent-helper gate check",
      "owner": "KAH",
      "required_when": "KHS gate verification is claimed",
      "evidence_state": "required_later",
      "not_a_skill_dependency": true
    }
  ]
}
```

### 10.4 `list --json` additions

`list --json` records include these fields without removing the existing CLIMVP output contract until a separately accepted compatibility decision says otherwise:

- `packs[].source_class`
- `packs[].source_class_evidence`
- `packs[].provenance_state`
- `packs[].shadowing`
- `packs[].skill_dependencies`
- `packs[].command_surface_dependencies`
- top-level `source_inventory_summary`

### 10.5 `install --dry-run --json` additions

`install --dry-run --json` records include no-write provenance evidence that can be bound into the existing approval hash when later implementation authorizes it:

- `source_inventory_snapshot`
- `target_profile_inventory`
- `provenance_conflicts`
- `shadowing_conflicts`
- `dependency_audit`
- `deleted_bundle_diagnostics`
- `approval_request.hash_includes_provenance: true`

These fields are dry-run evidence only. They do not authorize approved install execution and do not mutate profiles.

### 10.6 `doctor --json` additions

`doctor --json` records include audit summaries, per-installed-pack provenance fields, and diagnostics:

- `provenance_audit`
- `dependency_audit`
- `installed_packs[].source_class`
- `installed_packs[].source_class_evidence`
- `installed_packs[].provenance_state`
- `installed_packs[].shadowing`
- `installed_packs[].deleted_bundle_reference`
- diagnostics for `deleted_bundle_reference`, `source_class_ambiguous`, `command_surface_missing`, and `skill_dependency_missing`

`doctor --json` may require evidence for a missing KAB command-surface only when the selected lane requires KAB or claims backend runtime evidence. KASREL doctor output must not activate KAB or claim KAB completion.

## 11. Verification and follow-on task boundaries

KASREL-001 is complete when this accepted SOT, the docs index, roadmap references/status, and narrow docs contract assertions are verified. It remains docs/spec-only.

KASREL-002 is the completed implementation task for skill inventory and provenance classification.

KASREL-003 is the completed implementation task for dependency audit behavior, with completion evidence recorded by commit `75d0361`, KAH commit event `evt-001339`, and run close `evt-001340`.

KASREL-004 is the in-progress guidance-update task for install/readiness/orchestration/review/final-verification behavior; it must not be marked completed until its own guidance, docs-contract, search, review, final verification, and commit-approval gates pass.

Follow-on tasks must preserve these boundaries:

- no profile mutation without the existing dry-run and explicit approval-hash gate;
- no production CLI changes under `internal/skills/**` unless the specific later implementation task authorizes them;
- no fallback lookup path for deleted Hermes bundle skills;
- no command-surface dependency represented as a skill dependency;
- no KAB activation or backend runtime evidence claim from KASREL alone; and
- no KASUPD write-capable sync behavior unless a separate KASUPD task authorizes it.
