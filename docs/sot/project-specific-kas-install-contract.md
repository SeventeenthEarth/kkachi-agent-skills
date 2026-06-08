# Project-specific KAS install layout contract

Date: 2026-06-09
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / PR1 docs-contract evidence
Status: canonical SOT for `KASPROJ-001`; `KASPROJ-002` read-only dry-run planner is implemented/pre-commit-ready after accepted plan review, while approved install, doctor/repair, and migration behavior remain later KASPROJ tasks
Authority level: source of truth for project-specific KAS install layout, naming, manifest vocabulary, and doctor severity semantics
Scope: project-specific KAS skill suites installed under one Hermes profile; KASPROJ-002 authorizes only read-only dry-run planning evidence. No approved install, profile mutation, repair, migration, auth/token/gateway/provider/model config mutation, KAH state mutation, or KAB runtime activation is authorized by this document alone
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

## 5. Manifest vocabulary for later PRs

Later KASPROJ implementation tasks must extend the KAS profile manifest or a
compatible KAS-owned manifest with this vocabulary. These fields are specified
now so dry-run, approval, doctor, repair, and migration work use one contract.

| Field | Meaning |
|---|---|
| `project` | Concrete project suite id, for example `doksuri-server`. It determines the `skills/<project>/` directory and the required prefix for installed skill names. |
| `source_pack` | Upstream KAS pack/template input used to generate or update the project-specific suite, including source repo, source path, commit/checksum, language assumptions, and phase list. |
| `installed_skill` | Project-prefixed skill id installed under the project suite, for example `doksuri-server-plan`; generic ids are invalid for project suites. |
| `target_path` | Profile-relative canonical path for the installed skill file or directory, for example `skills/doksuri-server/doksuri-server-plan/SKILL.md`. |
| `drift_policy` | Per-suite or per-skill policy for source/profile drift, for example `fail_closed`, `manual_review_required`, `semantic_port_required`, or `repair_after_approval`. |

Minimum future manifest shape:

```json
{
  "version": "0.1",
  "kind": "kas_project_skill_manifest",
  "profile": "kwanwoo",
  "project_suites": [
    {
      "project": "doksuri-server",
      "source_pack": {
        "id": "kas-default-project-suite",
        "repo": "kkachi-hermes-skills",
        "commit": "<sha>",
        "checksum": "sha256:<hex>",
        "language_profile": "project-specific"
      },
      "drift_policy": "manual_review_required",
      "installed_skills": [
        {
          "installed_skill": "doksuri-server-plan",
          "target_path": "skills/doksuri-server/doksuri-server-plan/SKILL.md",
          "checksum": "sha256:<hex>"
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
| checksum mismatch | `error` | Installed content exists but does not match the manifest/source checksum. Manual review or approved repair is required. |
| unknown profile skill dir | `warning` | A profile skill directory not tracked by the KAS manifest is present. It must not be silently adopted, overwritten, or deleted. Escalate to `error` only if it shadows a required project-prefixed skill or makes identity ambiguous. |
| profile/source language drift | `warning` | Installed project guidance language/runtime/test-command assumptions differ from `source_pack` or project metadata. Escalate to `error` when the drift invalidates required phase behavior or verification commands. |

Diagnostic records must include `project`, `installed_skill` when applicable,
`target_path` when applicable, `severity`, `condition`, and `next_action`.

## 7. Dry-run, approval, repair, and migration contract

KASPROJ-002 implements the read-only dry-run planner surface:

```bash
kkachi-hermes-skills install-project-kas --profile <profile> --project <project> --source-pack kas-default-project-suite --dry-run [--json]
```

The KASPROJ-002 planner discovers current source skills under repo `skills/` through the virtual source suite `kas-default-project-suite`, renders project-prefixed installed names and target paths, reports all `target_path` changes, computes checksum and `plan_hash` evidence, and performs no profile writes. Source skill `kkachi-plan` renders to installed skill `<project>-plan` and target path `skills/<project>/<project>-plan/SKILL.md`; other source skills strip a leading `kkachi-` when present before applying the project prefix. This virtual suite must be formalized before KASPROJ-003 write-capable install.

KASPROJ-002 tailoring is dry-run prefix-render only. JSON evidence must emit preservation posture and `semantic_port_required_before_approved_install`, but must not claim semantic language/runtime/test-command adaptation and must not add language/runtime/test-command write behavior. Public checksums and `plan_hash` use `sha256:<hex>` strings; the canonical plan hash binds no-write evidence plus conflicts and diagnostics.

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

Later PRs may implement the remaining behavior only after their own acceptance criteria and verification gates are approved.
- `KASPROJ-003` approved install: copies the full project-specific suite only
  after approval evidence matches the dry-run plan hash. It must reject
  umbrella-only, generic-name, ambiguous-profile, conflict, or checksum-mismatch
  plans.
- `KASPROJ-004` doctor/repair/migrate: reports the severities above; repair and
  migration must start with dry-run evidence, require explicit approval before profile mutation, create backups for changed files, preserve project-tailored
  language/characteristics, and never mutate auth, tokens, secrets, gateway, provider/model config, KAH state, or KAB runtime state.

Migration from older generic/global installs must be explicit. A future migrate
command may propose moves such as `kkachi-plan` to `doksuri-server-plan`, but it
must not guess project language or authority content silently. If project-specific
content cannot be generated safely, it must produce a manual semantic-port task
rather than installing a generic fallback.

## 8. Relationship to existing contracts

- `docs/sot/kas-cli-contract.md` owns the existing CLIMVP profile-scoped
  list/install/doctor contract. KASPROJ extends that contract for
  project-specific suite identity and target layout. The current CLI implements
  only KASPROJ-002 project-suite dry-run planning; it does not implement
  project-suite approved install/repair/migration.
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
