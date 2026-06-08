# TODO-KASREL Hermes v0.16 skill provenance and dependency audit plan

Date: 2026-06-08
Owner: KAS workflow/policy layer
Confirming role: responsible approver / color-role governance evidence record
Status: TODO candidate SOT; planning record only; not implementation authorization and not installed-profile mutation approval
Authority level: candidate planning SOT for a future KAS release-compatibility epic
Scope: `kkachi-hermes-skills` KAS installer/list/doctor behavior, KAS skill-pack provenance, KAS dependency audit, and follow-on KAS skill documentation updates after Hermes Agent v0.16.0
Related docs: `docs/roadmap.md`, `docs/sot/kas-cli-contract.md`, `docs/sot/project-kas-sync-state.md`, `docs/README.md`

## 1. Current situation

Hermes Agent was updated to v0.16.0. The update changed the effective skill ecosystem enough that KAS should not rely only on the older profile-local pack manifest model.

Observed release-impact facts from the KAS maintenance investigation:

- KAS source health is not an emergency blocker: repository tests passed with `make test`, and direct references to removed Hermes bundle skills were not found in current KAS source packs.
- The installed `kkachi-hermes-skills` binary is older than the source tree and lacks at least one newer command surface that exists in source, so operational parity must be restored before profile repair or new CLI claims.
- The `hwangchung` profile KAS doctor currently fails because installed profile manifest/checksum state is drifted; profile repair requires the existing KAS dry-run/approval gate before any install mutation.
- Hermes v0.16 skill listing can distinguish broad source classes such as builtin, hub-installed, and local, but `local` is not specific enough for KAS policy because it can mix profile-personal skills and configured external/ops skill directories.
- Current Hermes loading order creates a shadowing risk: profile-local skills can mask same-named external/ops or bundle skills. KAS-managed installs must therefore detect name conflicts before writing profile-local copies.
- KAS pack discovery currently records pack id/category/name/source path/description/checksum, but it does not carry skill provenance, installed source class, declared dependencies, or command-surface dependency audit data.
- Current KAS source packs do not directly depend on deleted Hermes bundle skills. Known KAS skill-to-skill references are KAS-internal, while broader KAS operation depends on command surfaces such as Hermes skills listing, Hermes/Kanban CLI behavior, KAH phase/artifact/event/gate surfaces, and KAB backend/session evidence when a lane requires KAB.

Evidence checked for these facts:

- Repository health: `HOME=/Users/draccoon make test` in the KAS repo passed after the TODO SOT was created; it ran Go vet/build, docs-contract tests, discovery/install/doctor/kasstate tests, CLI tests, and e2e tests.
- Working tree scope: `HOME=/Users/draccoon git status --short --branch` showed `main...origin/main [ahead 18]` with only this TODO SOT as an untracked file.
- Source/install CLI parity: source help via `HOME=/Users/draccoon go run ./cmd/kkachi-hermes-skills --help` listed `sync-project-kas`; installed binary help via `HOME=/Users/draccoon kkachi-hermes-skills --help` did not list `sync-project-kas`, proving installed binary drift from source.
- Profile install health: no-write `HOME=/Users/draccoon kkachi-hermes-skills doctor --profile hwangchung --json` returned exit code 2 / `ok:false`, with manifest checksum drift, installed file checksum mismatches, and one missing manifested file under `skills/kkachi-orchestrate/references/kan-plugin-readiness-and-activation.md`.
- Hermes source-label ambiguity: `HOME=/Users/draccoon hermes skills list --source all` and `--source local` showed KAS packs and other profile/external skills under the same `local` source label, which is insufficient for KAS policy to distinguish ops external skills from profile-personal or KAS-managed skills.
- Current KAS pack metadata limit: `internal/skills/discovery/discovery.go` defines `SourcePack` with `PackID`, `Category`, `Name`, `SourcePath`, `Description`, and `Checksum`; current pack payload does not include provenance source class, declared dependencies, or command-surface dependency audit data.
- Removed bundle skill direct-reference check: KAS source pack search for known removed/moved bundle skill names including `kanban-codex-lane`, `debugging-hermes-tui-commands`, `writing-plans`, `linear`, and `spotify` returned zero direct `skills/**/SKILL.md` hits.

## 2. Policy decisions to preserve

1. **No fallback for deleted bundle skills.** If a skill was removed from the Hermes bundle and a live KAS surface still references it, KAS should report cleanup/doctor failure or policy drift. It must not silently fall back to stale bundle paths or substitute another skill.
2. **KAS install scope remains profile-local.** KAS may inspect bundle, hub, ops external, and profile-local skills to produce evidence and conflict diagnostics, but approved KAS install still writes only the target profile-local KAS pack paths unless a separate task explicitly changes that policy.
3. **Source class must be path/manifest aware.** KAS should not treat Hermes CLI `--source local` as enough to distinguish personal profile skills from external/ops skills.
4. **Dry-run before mutation remains mandatory.** Any profile repair, install, or update must use `--dry-run --json`, preserve the dry-run hash, and require explicit `--approve dry-run:<hash>` before writing.
5. **Release compatibility is not KAB runtime authority.** This plan can improve KAS CLI diagnostics and skill guidance, but it does not activate KAB execution, mutate auth/tokens/gateway/provider settings, or claim backend runtime support.
6. **One task is one PR candidate.** The follow-on work should be split so each task has its own acceptance criteria, verification, evidence, and review gate.

## 3. Recommended roadmap shape

Open a new epic instead of adding one broad task to a completed or unrelated epic.

Recommended epic id:

```text
KASREL — KAS release-compatibility and skill provenance audit
```

Rationale:

- `CLIMVP` is the nearest surface because it owns `list`, `install`, and `doctor`, but its tasks are already completed. Reopening it as one omnibus implementation task would blur completion evidence.
- `KASUPD` owns project-specific KAS sync/semantic-port workflow, not general Hermes release compatibility or profile skill provenance.
- The new work spans CLI inventory, dependency audit, dry-run conflict behavior, manifest/output schema, and KAS skill guidance. Those surfaces should not be bundled into a single PR candidate.

## 4. Proposed task breakdown

### KASREL-001 — Specify Hermes v0.16 skill provenance and dependency audit contract

Type: docs/spec-only.

Acceptance criteria:

- Define canonical KAS source classes for skill audit output, at minimum:
  - `bundle_builtin`
  - `hub_installed`
  - `ops_external`
  - `profile_personal`
  - `kas_managed_profile`
  - `unknown_or_unclassified`
- Define how KAS resolves those classes from profile config, bundle paths, hub lock/manifest data, KAS manifest data, and configured `skills.external_dirs`.
- Define deleted-bundle references as cleanup/fail-closed diagnostics, not fallback candidates.
- Define the distinction between skill dependencies and command-surface dependencies.
- Define required JSON fields for future `list`, `install --dry-run`, and `doctor --json` outputs without forcing immediate implementation.
- State that the SOT itself does not authorize profile mutation or KAB runtime activation.

Verification/evidence:

- Readback of the new/updated SOT.
- Targeted stale/fallback wording search.
- `git diff --check`.
- Roadmap entry added as `Planned` only after responsible-approver confirmation of the task shape.

### KASREL-002 — Implement skill inventory and provenance classification

Type: CLI implementation.

Acceptance criteria:

- KAS can build an inventory of installed/effective skills for a profile using profile-local skills, KAS manifest, hub-installed metadata, bundle paths, and configured external/ops skill directories.
- `list --json`, `install --dry-run --json`, and/or `doctor --json` expose provenance evidence without leaking secrets.
- Name shadowing is reported deterministically, including profile-local over ops/bundle and KAS-managed over non-KAS skill cases.
- Existing install behavior remains profile-local and approval-gated.

Verification/evidence:

- Unit tests for path classification and shadowing cases.
- CLI smoke tests on isolated temporary profiles.
- Representative real-profile no-write `--dry-run --json` evidence.
- `make test` after final change.

### KASREL-003 — Implement KAS dependency audit

Type: CLI implementation.

Acceptance criteria:

- KAS records or derives KAS pack dependencies separately from external skill names and command-surface dependencies.
- KAS-internal dependencies such as phase-skill handoffs can be checked against KAS-managed/profile install state.
- Command surfaces such as Hermes skills list, KAH phase/artifact/event/gate commands, Kanban CLI behavior, and KAB runtime evidence are reported as command-surface dependencies rather than fake skill dependencies.
- Missing deleted-bundle references fail closed as cleanup diagnostics.

Verification/evidence:

- Unit tests for declared/derived dependency reporting.
- Doctor tests for missing dependency, source-class mismatch, and deleted-bundle cleanup diagnostic cases.
- Real-profile no-write doctor evidence.
- `make test` after final change.

### KASREL-004 — Update KAS skill guidance and final gates

Type: docs/skill maintenance.

Acceptance criteria:

- Update KAS guidance where release/provenance/dependency checks affect operator behavior, likely including:
  - `kkachi-install-guide`
  - `kkachi-final-verify`
  - `kkachi-orchestrate`
  - `kkachi-review`
- Guidance requires provenance/dependency evidence before reporting install/readiness confidence.
- Guidance preserves no-fallback handling for deleted bundle skills.
- Guidance distinguishes KAS/KAH/KAB responsibilities and does not turn release compatibility checks into KAB runtime authorization.

Verification/evidence:

- Readback of changed skills/docs.
- Targeted search for forbidden fallback wording.
- Relevant docs-contract or repository test target, plus `make test` if docs changes affect tested surfaces.
- Review-ready report with changed files and residual risks.

## 5. Proposed P0 before implementation tasks

This section records proposed preparation work only. This TODO SOT does not authorize running P0. Rebuilding or installing the local `kkachi-hermes-skills` binary is an operator-environment mutation and requires explicit responsible-approver authorization before execution, separate from any later approval to repair profile skill-pack state.

After explicit responsible-approver authorization, perform a narrow operational parity/repair preparation step before implementing KASREL code tasks:

1. Rebuild/install the local `kkachi-hermes-skills` binary from the source checkout using real-user HOME and explicit `GOBIN`.
2. Verify `kkachi-hermes-skills --help` includes the current source command surface.
3. Run no-write doctor/list/install dry-run checks against the `hwangchung` profile.
4. If profile repair is needed, present the `--dry-run --json` output and approval hash to the responsible approver; do not run approved install until explicit hash approval is recorded.

This P0 is not a substitute for KASREL implementation. It restores the operator lane so future evidence is trustworthy.

## 6. Risks and blockers

- **Manifest drift:** existing profile KAS installs are drifted; implementation tests must isolate temp profiles so the real profile does not mask bugs.
- **Classification ambiguity:** Hermes `local` source class is too broad; KAS must use configured paths and manifests, not just Hermes CLI labels.
- **Shadowing:** same-name profile-local skills can hide ops/bundle skills; install dry-run must report this before writes.
- **Overreach risk:** dependency audit must not become a broad runtime verifier for KAB or KAH. It should report command-surface assumptions and require existing KAH/KAB evidence where applicable.
- **Stale fallback risk:** any logic that tries to recover deleted bundle skills would violate current cleanup policy.

## 7. Open decisions for responsible approver

1. Confirm the epic id: `KASREL` or a different id.
2. Confirm whether `KASREL-001..004` is the desired task split.
3. Confirm whether the TODO SOT should be promoted into the docs authority ladder and roadmap immediately, or remain a TODO planning note until after review.
4. Confirm when P0 binary parity/profile dry-run repair should be executed.

## 8. Non-goals

- No profile install, repair, or approved copy mutation is authorized by this document.
- No auth, token, gateway, provider, model, or credential mutation is authorized.
- No KAB Stage 2/Stage 3 activation is authorized.
- No write-capable project-specific KAS sync is authorized.
- No fallback path for removed Hermes bundle skills is authorized.
