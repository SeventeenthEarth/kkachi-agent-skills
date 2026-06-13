# Project-specific KAS install layout contract

Date: 2026-06-09
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / PR1 docs-contract evidence
Status: canonical SOT for `KASPROJ-001`; `KASPROJ-002` read-only dry-run planner is implemented/pre-commit-ready; `KASPROJ-003` approved install is implemented/in-review pending gates; `KASPROJ-004` project-suite doctor/repair/migrate is implemented/in-review pending gates; `KASPROJ-005` project-tailored doctor checksum policy is in review
Authority level: source of truth for project-specific KAS install layout, naming, manifest vocabulary, and doctor severity semantics
Scope: project-specific KAS skill suites installed under one Hermes profile; KASPROJ-003 authorizes approval-hash-bound project suite install writes, and KASPROJ-004 authorizes approval-hash-bound project suite repair/migration writes after matching current dry-runs. No auth/token/gateway/provider/model config mutation, KAH state mutation, KAB runtime activation, semantic-port completion claim, generic fallback, generic deletion, or operational rollout is authorized by this document alone
Related docs: `docs/sot/kas-cli-contract.md`, `docs/sot/project-kas-sync-state.md`, `docs/sot/minimum-pilot-cli-lane.md`, `docs/README.md`, `docs/roadmap.md`, `docs/kkachi-docs-map.yaml`

## 1. Decision

When a general installs KAS into multiple concrete projects, KAS must install a
separate project-specific skill suite for each project inside that general's
Hermes profile. A profile must not receive one global generic KAS suite and then
reuse it for all projects.

Canonical installed target layout:

```text
~/.hermes/profiles/<profile>/skills/<project>/<project>-<phase-or-skill>/SKILL.md
```

Relative to a Hermes profile, the canonical target path is:

```text
skills/<project>/<project>-<phase-or-skill>/SKILL.md
```

The `<project>` segment is the concrete project suite id, not the upstream KAS
repository name. The skill directory name must begin with the same project id.
For example, `doksuri-server-plan` is valid for the `doksuri-server` suite;
generic names such as `kkachi-plan`, `kas-plan`, or plain `plan` are invalid as
installed project-suite skill names.

## 2. Required examples

For general/profile `kwanwoo`, installing KAS for four concrete Doksuri projects
must produce separate suites like these:

```text
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-server/doksuri-server-plan/SKILL.md
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-server/doksuri-server-implement/SKILL.md
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-server/doksuri-server-final-verify/SKILL.md

/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-daemon/doksuri-daemon-plan/SKILL.md
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-daemon/doksuri-daemon-implement/SKILL.md
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-daemon/doksuri-daemon-final-verify/SKILL.md

/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-client/doksuri-client-plan/SKILL.md
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-client/doksuri-client-implement/SKILL.md
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-client/doksuri-client-final-verify/SKILL.md

/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-cli/doksuri-cli-plan/SKILL.md
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-cli/doksuri-cli-implement/SKILL.md
/Users/draccoon/.hermes/profiles/kwanwoo/skills/doksuri-cli/doksuri-cli-final-verify/SKILL.md
```

Each suite may contain skills with equivalent phase semantics. For example,
`doksuri-server-plan`, `doksuri-daemon-plan`, `doksuri-client-plan`, and
`doksuri-cli-plan` may all represent the KAS planning phase. They are still four
separate installed skills because their content must be tailored to the
project's language, runtime, repository layout, test commands, docs map,
authority ladder, and risk boundaries.

## 3. Naming and collision rules

Project-prefix skill names are mandatory.

Valid examples:

```text
doksuri-server-plan
doksuri-daemon-plan
doksuri-client-implement
doksuri-cli-final-verify
```

Invalid examples:

```text
kkachi-plan
kas-plan
plan
implement
final-verify
```

Duplicate generic names are invalid in one Hermes profile because they make the
effective skill namespace ambiguous. Even when Hermes preserves directory
nesting, operators, manifests, diagnostics, dependency-audit records, and future
repair/migration tools need a stable unambiguous `installed_skill` id. If four
projects in one profile all install a generic `kkachi-plan`, KAS cannot safely
state which project language, authority ladder, test commands, or KAB adoption
stage the skill represents. The project prefix is therefore part of the installed
skill identity, not cosmetic naming.

## 4. Umbrella skill boundary

An umbrella or entrypoint skill such as:

```text
skills/<project>/<project>-kas/SKILL.md
```

may be useful as a project suite guide, but umbrella-only installation is incomplete and invalid unless accompanied by the full project-specific suite for
that project. The full suite is the set of project-prefixed phase/operation
skills required by the selected source pack and manifest. A profile that has
only `skills/doksuri-server/doksuri-server-kas/SKILL.md` but lacks required
project-prefixed phase skills such as `doksuri-server-plan` and
`doksuri-server-final-verify` must be reported as `umbrella_only`, not healthy.

Umbrella-only status must not be treated as a successful KAS install, a healthy
project KAS suite, or permission to fall back to global generic skills.

## 5. Manifest vocabulary

KASPROJ-003 extends the existing KAS profile manifest at `.kas/skill-pack-manifest.json` with `project_suites[]`. The top-level manifest kind remains `kas_profile_skill_manifest` for compatibility with existing profile-scoped install/doctor behavior; each project suite entry uses `kind: kas_project_skill_manifest`. These fields keep dry-run, approval, doctor, repair, and migration work on one contract.

| Field | Meaning |
|---|---|
| `project` | Concrete project suite id, for example `doksuri-server`. It determines the `skills/<project>/` directory and the required prefix for installed skill names. |
| `suite_mode` | Role projection mode for the installed project suite. `full` means the explicit Blue commander role selected the complete source suite; `role_subset` means only registry-declared role skills were selected. |
| `suite_role` | Explicit role registry id used for the project suite install, for example `blue_commander`, `red_reviewer`, `orange_pm_reviewer`, or `gray_scribe`. KAS must not infer this from the profile name. |
| `role_registry` | Hash-bound registry evidence including path `registries/project-suite-roles.yaml`, version `role-aware-project-suite/v1`, and `sha256:<hex>` checksum. |
| `source_pack` | Upstream KAS pack/template input used to generate or update the project-specific suite, including source repo, source path, commit/checksum, language assumptions, and phase list. |
| `selected_skills` | Source/rendered skill evidence selected by the role projection, including source skill id, installed skill id, source pack id, and target path. |
| `excluded_skills` | Source/rendered skill evidence excluded by the role projection, including reason `outside_suite_role` for ordinary role-subset exclusion. |
| `installed_skill` | Project-prefixed skill id installed under the project suite, for example `doksuri-server-plan`; generic ids are invalid for project suites. |
| `target_path` | Profile-relative canonical path for the installed skill file or directory, for example `skills/doksuri-server/doksuri-server-plan/SKILL.md`. |
| `drift_policy` | Per-suite or per-skill policy for source/profile drift, for example `fail_closed`, `manual_review_required`, `semantic_port_required`, or `repair_after_approval`. |

Minimum compatible manifest shape:

```json
{
  "version": "0.1",
  "kind": "kas_profile_skill_manifest",
  "profile": {"name": "kwanwoo", "root": "/abs/profile/root"},
  "installs": [],
  "project_suites": [
    {
      "kind": "kas_project_skill_manifest",
      "install_id": "kas-project-install-<timestamp>-<hash>",
      "approval_evidence_ref": "dry-run:sha256:<hex>",
      "dry_run_plan_hash": "sha256:<hex>",
      "approved_plan_hash": "sha256:<hex>",
      "project": "doksuri-server",
      "suite_mode": "role_subset",
      "suite_role": "red_reviewer",
      "role_registry": {
        "path": "registries/project-suite-roles.yaml",
        "version": "role-aware-project-suite/v1",
        "checksum": "sha256:<hex>"
      },
      "source_pack": {
        "id": "kas-default-project-suite",
        "repo": "kkachi-agent-skills",
        "commit": "<sha-or-null>",
        "checksum": "sha256:<hex>",
        "language_profile": "project-specific-prefix-render-only",
        "formal_registry": "skill-pack.yaml"
      },
      "selected_skills": [
        {
          "source_skill": "kkachi-review",
          "installed_skill": "doksuri-server-review",
          "source_pack_id": "kkachi-review",
          "target_path": "skills/doksuri-server/doksuri-server-review/SKILL.md"
        }
      ],
      "excluded_skills": [
        {
          "source_skill": "kkachi-implement",
          "installed_skill": "doksuri-server-implement",
          "source_pack_id": "kkachi-implement",
          "target_path": "skills/doksuri-server/doksuri-server-implement/SKILL.md",
          "reason": "outside_suite_role"
        }
      ],
      "drift_policy": "role_subset_expected",
      "semantic_adaptation_claimed": false,
      "installed_skills": [
        {
          "installed_skill": "doksuri-server-plan",
          "source_pack_id": "kkachi-plan",
          "target_path": "skills/doksuri-server/doksuri-server-plan/SKILL.md",
          "checksum": "sha256:<hex>",
          "backup_relative_path": null,
          "drift_policy": "manual_review_required",
          "tailoring_mode": "prefix_render_only"
        }
      ]
    }
  ]
}
```

The manifest must stay KAS-owned metadata outside Hermes-loaded skill
directories unless a later SOT explicitly changes that rule. It must not contain
auth tokens, secrets, gateway credentials, provider keys, model credentials, or
runtime session state.

## 6. Doctor severity semantics

Future doctor/repair/migration work must report these conditions with stable
severity. This section specifies diagnostics only; it does not implement doctor,
repair, or migration behavior.

| Condition | Required severity | Required diagnostic meaning |
|---|---|---|
| missing project suite | `error` | No `skills/<project>/` suite exists for a manifest/requested project. KAS must fail closed and must not use a global generic suite as a substitute. |
| umbrella-only | `error` | The project umbrella skill exists without the full required project-prefixed suite. The install is incomplete/invalid. |
| missing file | `error` | A manifest-required `target_path` such as `skills/<project>/<project>-plan/SKILL.md` is absent. |
| checksum mismatch | `error` | Prefix-render-only or otherwise untrusted installed content exists but does not match the manifest/source checksum. Manual review or approved repair is required. |
| project tailoring checksum drift | `warning` | Installed content differs from the initial/template checksum, but the manifest marks the skill as project-local semantic tailoring with `tailoring_mode: profile_local_repo_semantic_tailoring` and `drift_policy: manual_review_required`. This is expected for real project use and must not make doctor fail. |
| unknown profile skill dir | `warning` | A profile skill directory not tracked by the KAS manifest is present. It must not be silently adopted, overwritten, or deleted. Escalate to `error` only if it shadows a required project-prefixed skill or makes identity ambiguous. |
| profile/source language drift | `warning` | Installed project guidance language/runtime/test-command assumptions differ from `source_pack` or project metadata. Escalate to `error` when the drift invalidates required phase behavior or verification commands. |

Diagnostic records must include `project`, `installed_skill` when applicable,
`target_path` when applicable, `severity`, `condition`, and `next_action`.

## 7. Dry-run, approval, repair, and migration contract

KASPROJ-002 originally introduced the read-only dry-run planner surface, but
KASROLE-002 deliberately breaks the copyable no-role path. Operators must now
select an explicit suite role even when using the compatibility alias:

```bash
kkachi-agent-skills install-project-kas --profile <profile> --project <project> --suite-role blue_commander --source-pack kas-default-project-suite --dry-run [--json]
```

KASROLE-002 extends the public lifecycle surface with explicit role-aware project install:

```bash
kkachi-agent-skills install --profile <profile> --project <project> --suite-role <role> --dry-run [--json]
kkachi-agent-skills install --profile <profile> --project <project> --suite-role <role> --apply dry-run:sha256:<hash> [--json]
```

The compatibility alias `install-project-kas` accepts the same required
`--suite-role`. Missing or unknown roles fail closed with `ok:false` / exit 2
before writes, and role inference from profile names is forbidden. The
role-aware plan hash binds
`suite_role`, `suite_mode`, role display label, `role_registry`,
`selected_skills`, `excluded_skills`, source suite checksum, changed paths,
conflicts/diagnostics, backup plan, manifest state, and no-write evidence.
Human output must include the role label plus selected/excluded counts and a
copy/paste apply command that preserves `--suite-role`.

The planner resolves the formal source suite `kas-default-project-suite` from repository `skill-pack.yaml` plus current source skills under repo `skills/`, renders project-prefixed installed names and target paths, reports all `target_path` changes, computes checksum and `plan_hash` evidence, and performs no profile writes. Source skill `kkachi-plan` renders to installed skill `<project>-plan` and target path `skills/<project>/<project>-plan/SKILL.md`; other source skills strip a leading `kkachi-` when present before applying the project prefix.

KASPROJ-003 tailoring remains prefix-render only. JSON/manifest evidence must record `semantic_adaptation_claimed:false`, preserve `drift_policy: manual_review_required`, and must not claim semantic language/runtime/test-command adaptation, semantic-port completion, operational rollout, or KAB activation. Public checksums and `plan_hash` use `sha256:<hex>` strings; the canonical plan hash binds CLI version, target profile/root, manifest path/state/previous manifest sha, source repo commit/dirty, formal source suite checksum, planned manifest/skills, changed paths, conflicts/diagnostics, backup plan, and dry-run no-write evidence.

Generic installed skill names, missing project prefix, unsafe/escaping target paths, unknown source pack, and umbrella-only suite examples are fail-closed conflicts: JSON output is `ok:false` and the CLI exits 2, not `ok:true` with warnings. Example conflict shape:

```json
{
  "ok": false,
  "command": "install-project-kas",
  "mode": "project_dry_run",
  "dry_run": true,
  "conflicts": [
    {
      "severity": "error",
      "condition": "umbrella_only",
      "project": "doksuri-server",
      "installed_skill": "doksuri-server-kas",
      "target_path": "skills/doksuri-server/doksuri-server-kas/SKILL.md",
      "next_action": "Install the full project-specific suite after KASPROJ-003 approval evidence."
    }
  ],
  "plan_hash": "sha256:<hex>"
}
```

KASPROJ-003 approved install surface:

```bash
kkachi-agent-skills install-project-kas --profile <profile> --project <project> --suite-role blue_commander --source-pack kas-default-project-suite --approve dry-run:<plan_hash> [--json]
```

Approved install recomputes the current dry-run, compares approval evidence to the recomputed `plan_hash`, and fails closed before any write unless `--suite-role` is present, `--approve dry-run:<plan_hash>` exactly matches the recomputed plan hash, and the plan is `ok:true`. Missing both `--dry-run`/`--approve`, using both together, malformed approval evidence, unsupported write/force/repair/migrate/from-generic flags, unguarded `--profile-root`, unknown profiles, missing/unknown suite roles, unsafe/path-traversing targets, symlink escapes, ambiguous duplicate manifest entries, duplicate installed skills/target paths, unmanifested existing targets, local modifications, and checksum mismatches must return exit 2 with `ok:false` JSON.

Safe write ordering is: preflight all source/target checksums and paths, create backups for trusted replacements, atomically write project skill files, verify post-write checksums, and write/update the compatible profile manifest last. Existing top-level `installs` and unrelated `project_suites` are preserved; only the matching `(project, source_pack.id)` suite is replaced. Human and JSON approved output must include approval evidence, `install_id`, `manifest_path`, backup/recovery evidence, changed-path counts/actions, and `next_action` stating that project-suite doctor should be rerun.

KASPROJ-004 implemented doctor/repair/migrate forms:

```bash
kkachi-agent-skills doctor --profile <profile> --project <project> --project-suite [--json]
kkachi-agent-skills repair-project-kas --profile <profile> --project <project> --dry-run [--json]
kkachi-agent-skills repair-project-kas --profile <profile> --project <project> --approve dry-run:<hash> [--json]
kkachi-agent-skills migrate-project-kas --profile <profile> --project <project> --from-generic --dry-run [--json]
kkachi-agent-skills migrate-project-kas --profile <profile> --project <project> --from-generic --approve dry-run:<hash> [--json]
```

When `doctor --project-suite` is absent, `doctor --project <path>` keeps its existing KAH project-path meaning. When `--project-suite` is present, `--project` is the project suite id. Repair and migration default `source_pack` to `kas-default-project-suite`; optional explicit `--source-pack` values fail closed when unknown, and the resolved source pack is included in JSON, diagnostics, and plan-hash evidence.

`KASPROJ-004` doctor/repair/migrate reports the severities above; repair and migration start with dry-run evidence, require explicit approval before profile mutation, create backups for changed files and previous manifests, preserve project-tailored language/characteristics, and never mutate auth, tokens, secrets, gateway, provider/model config, KAH state, or KAB runtime state.

`KASPROJ-005` fixes the project-tailored doctor policy: source/template checksum equality is not required after a project suite has been adapted for a concrete repository. `doctor --project-suite` must report `project_tailoring_checksum_drift` as a warning, not an error, when the manifest skill record uses `tailoring_mode: profile_local_repo_semantic_tailoring` with `drift_policy: manual_review_required`. Prefix-render-only records, missing files, missing checksums, unsafe targets, and unmanifested edits remain fail-closed errors. KASPROJ-005 does not approve repair writes, manifest resealing, broad rollout, or any auth/token/gateway/provider/model/KAH/KAB mutation.

Migration from older generic/global installs is explicit through `--from-generic`. It may propose non-deleting creation from clean KAS-managed generic candidates such as `kkachi-plan` to `doksuri-server-plan`, but it must not guess project language or authority content silently. If project-specific content cannot be generated safely, it produces `manual_semantic_port_tasks[]` rather than installing a generic fallback. KASPROJ-004 retains generic files; deletion or de-manifest policy is deferred unless separately approved.

## 8. Relationship to existing contracts

- `docs/sot/kas-cli-contract.md` owns the existing CLIMVP profile-scoped
  list/install/doctor contract. KASPROJ extends that contract for
  project-specific suite identity and target layout. The current CLI implements
  KASPROJ-002 project-suite dry-run planning, KASPROJ-003 approval-hash-bound
  project-suite install, KASPROJ-004 project-suite doctor/repair/migration,
  and KASPROJ-005 project-tailored doctor checksum policy.
- `docs/sot/project-kas-sync-state.md` owns upstream sync state after a
  project-specific suite exists. This document owns initial install layout,
  naming, manifest vocabulary, and doctor severity for suite presence/identity.
- KAH remains deterministic project state/evidence tooling. KAS project-suite
  manifests are KAS operating metadata, not KAH run state.
- KAB remains backend runtime/session evidence. Project-specific skill install
  status is not KAB evidence and does not activate any KAB adoption stage.

## 9. Acceptance and completion boundary

`KASPROJ-001` may be considered in-review/commit-ready when:

1. this SOT exists under `docs/sot/` and preserves the canonical layout,
   project-prefix, umbrella-only, manifest vocabulary, doctor severity, and
   no-implementation boundaries;
2. `docs/README.md`, `docs/kkachi-docs-map.yaml`, `docs/roadmap.md`, and
   `docs/sot/kas-cli-contract.md` reference the project-specific install
   contract without overclaiming implementation;
3. docs-contract tests fail if the required layout, vocabulary, severity, docs
   registration, or approval boundaries disappear;
4. `gofmt`, `go test ./tests/docs_contract`, and `git diff --check` pass.

It must not be marked `Completed` until the responsible commit/review gate for
this docs/spec PR is satisfied. Approved install, doctor, repair, and migration
work must remain under KASPROJ-003 through KASPROJ-004 or later explicit tasks.
