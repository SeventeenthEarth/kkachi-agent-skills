# Project-specific KAS sync state SOT

Date: 2026-06-06
Owner: KAS workflow/policy layer
Confirming role: Responsible approver / governance evidence record
Status: accepted SOT for `KASUPD-001`; `KASUPD-002` implements read-only state validation and legacy-marker compatibility read, and `KASUPD-003` implements dry-run classification plus semantic-port packet evidence while approved sync writes and pilot updates remain planned
Authority level: SOT for project-specific KAS static state and upstream sync/update workflow
Scope: installed/profile-local project-specific KAS suites such as KAN, KLM, `kan-plugin`, and `kan-control`; no runtime KAB execution, profile mutation, or project KAS update is authorized by this document alone
Related docs: `docs/roadmap.md`, `docs/sot/kas-cli-contract.md`, `docs/sot/khs-architecture-and-integration.md`, `docs/kkachi-docs-map.yaml`

## 1. Decision

Project-specific KAS suites must use one machine-readable static state file as the durable basis for upstream KAS sync decisions. The file combines the KAB adoption stage and upstream KAS baseline metadata because both describe the same project-specific KAS operating state.

Canonical path inside an installed project-specific KAS umbrella skill:

```text
skills/<project>/<project>-kas/references/kas-project-state.yaml
```

Examples:

```text
skills/kan-plugin/kan-plugin-kas/references/kas-project-state.yaml
skills/kan-control/kan-control-kas/references/kas-project-state.yaml
skills/klm/klm-kas/references/kas-project-state.yaml
```

The previous generated `references/kab-adoption-stage.md` marker remains a compatibility surface until migration. New project-specific KAS sync work should prefer `kas-project-state.yaml`; during transition, tools may read either format but must write the YAML state when performing a new approved install/update.

## 2. Ownership and boundaries

- KAS owns the project-specific KAS state schema, upstream sync policy, semantic porting workflow, and skill guidance.
- KAH may audit or carry project overlay/backend-policy mirrors when explicitly initialized, but this YAML is KAS operating state, not KAH run state.
- KAB remains backend runtime/session evidence. This YAML is not KAB execution evidence and does not activate Stage 2 by itself.
- Missing, unreadable, unsupported, or schema-invalid state fails closed to Stage 1 claims: the run may record direct Codex app-server evidence, but it must not claim KAB Codex execution.
- The state file must never carry auth tokens, secrets, gateway credentials, provider keys, model credentials, or mutable runtime session state.

## 3. Required YAML schema

Minimum schema version: `0.1`.

```yaml
version: "0.1"

project:
  id: "kan-plugin"
  repo: "kkachi-agent-network-plugin"
  kas_suite: "kan-plugin"
  profile: "hwangchung"

kab_adoption_stage:
  numeric: 1
  canonical: "stage1_direct_codex_app_server_baseline"
  selection_source: "approved_project_policy"
  selected_at: "2026-06-06T00:00:00Z"
  approval_evidence: "<approval-ref-or-not_applicable>"
  stage2_activation: false

upstream_kas:
  repo: "kkachi-hermes-skills"
  remote: "github.com/SeventeenthEarth/kkachi-hermes-skills"
  commit: "<full-git-sha>"
  dirty: false
  synced_at: "2026-06-06T00:00:00Z"
  sync_task: "KASUPD-001"

pack_baselines:
  - upstream_pack: "kkachi-plan"
    project_skill: "kan-plugin-plan"
    source_checksum: "sha256:<hex>"
    project_checksum: "sha256:<hex>"
    merge_mode: "semantic_port"

overlay_policy:
  local_overlay_allowed: true
  preserve_project_authority: true
  preserve_project_roadmap_ids: true
  preserve_project_test_commands: true
  preserve_role_labels: true
  overwrite_mode: "never_without_review"

update_policy:
  default_mode: "dry_run_then_semantic_merge"
  auto_apply_when:
    - "target_file_missing"
    - "local_unchanged_since_baseline_and_upstream_changed"
  require_llm_merge_when:
    - "local_changed_since_baseline_and_upstream_changed"
    - "project_skill_mapping_exists"
    - "policy_text_requires_project_specific_translation"
  fail_closed_when:
    - "state_file_missing"
    - "state_schema_invalid"
    - "stage_unsupported"
    - "upstream_commit_unknown"
    - "checksum_mismatch_without_baseline"
    - "auth_token_gateway_or_provider_mutation_detected"

evidence_posture:
  not_kab_runtime_evidence: true
  not_stage2_activation_by_itself: true
  missing_or_unreadable_fails_to_stage1_claims: true
```

Additional fields are allowed only when they are deterministic, non-secret, and do not weaken the required fail-closed behavior.

## 4. Update workflow

Project-specific KAS update work must use the state YAML plus dry-run evidence, not blind overwrite.

1. Read `kas-project-state.yaml` and validate schema, KAB stage, upstream commit, pack mappings, and overlay policy.
2. Resolve the current KAS source repo and commit. If the source repo is dirty, record that fact and require explicit approval before using it as an upstream baseline.
3. Run KAS dry-run/list style discovery to generate machine evidence for current upstream packs, checksums, changed paths, and conflicts. Dry-run is an evidence generator, not permission to mutate project-specific skills.
4. Classify each mapped file or pack with three-way state:
   - `auto_copy_candidate`: local file still matches the recorded project checksum and upstream changed.
   - `local_only`: project-specific file changed but upstream baseline did not.
   - `semantic_merge_required`: local project file changed and upstream changed, or project-specific translation is required.
   - `new_upstream_candidate`: upstream pack/file is new and may or may not apply to the project-specific suite.
   - `removed_or_renamed_upstream`: upstream pack/file disappeared or moved; report before deleting anything.
   - `fail_closed_conflict`: state/checksum/schema/stage evidence is insufficient.
5. For `semantic_merge_required` and `new_upstream_candidate`, compose an LLM semantic-port prompt with the state YAML, dry-run evidence, upstream diff, local project skill, and preservation constraints.
6. Preserve project-specific authority labels, roadmap IDs, test commands, role names, KAN/KLM boundaries, and selected KAB adoption stage unless 주군 explicitly approves changing them.
7. Run verification and Red/Orange/Gray review for durable project-specific KAS changes before reporting commit/install readiness.
8. Update `kas-project-state.yaml` only after the accepted sync result is verified. The updated state must record the new upstream commit, pack baselines, sync task, and evidence posture.

## 5. Automation contract

KASUPD-002 implemented the first bounded read-only validation surface for this workflow. KASUPD-003 extends that same dry-run command with read-only three-way classification and semantic-port packet evidence:

```bash
kkachi-hermes-skills sync-project-kas \
  --profile <profile> \
  --project <project-id> \
  --state skills/<project>/<project>-kas/references/kas-project-state.yaml \
  [--repo <current-upstream-kas-repo>] \
  [--project-root <project-specific-kas-root>] \
  --dry-run \
  --json
```

This command must:

- require `--dry-run` and fail closed when it is omitted;
- read `kas-project-state.yaml` from `--state`;
- read the sibling legacy `kab-adoption-stage.md` marker for compatibility/reporting when present;
- keep legacy-only evidence from upgrading missing or invalid YAML to valid state;
- emit `yaml_state_path`, `legacy_marker_path`, read-surface states, effective static stage claim, write target after future approved sync, validation diagnostics, and next action;
- recognize valid Stage 1 and Stage 2 static YAML values but keep `kab_execution_claim_allowed=false`;
- keep Stage 3 unsupported/reserved;
- reject unsupported YAML features, invalid schema, bad commits/checksums, weakened overlay/evidence posture, and secret/auth/token/gateway/provider/runtime-state-like input;
- when state is valid, compare current upstream KAS packs, the recorded upstream baseline at `upstream_kas.commit`, and mapped project-local skills;
- verify baseline pack checksums without checkout mutation or worktree overwrite;
- emit only the six documented classifications: `auto_copy_candidate`, `local_only`, `semantic_merge_required`, `new_upstream_candidate`, `removed_or_renamed_upstream`, and `fail_closed_conflict`;
- represent unchanged mapped packs outside the classification vocabulary as `unchanged_mappings` and `summary.no_action_count`;
- generate semantic-port packet JSON content for `semantic_merge_required` and `new_upstream_candidate` classifications without writing packet files;
- fail closed with `ok:false` and non-zero exit on dirty current upstream source, baseline checksum mismatch, ambiguous or escaping project skill paths, missing expected mapped local skills, unreadable baseline commits, or any `fail_closed_conflict`;
- perform no writes.

KASUPD-003 still does not implement approved sync writes, write-capable apply mode, or a project pilot. Those remain under KASUPD-004 or a later explicitly approved task.

Write-capable project-specific KAS sync must require a later approved task and:

- valid state YAML;
- current dry-run evidence;
- approved plan hash or explicit 주군 approval evidence;
- backups for overwritten files;
- no conflicts or an accepted semantic merge artifact;
- KAH/KAS run evidence and review gates.

Machine-readable output must distinguish the preferred YAML state path from any legacy stage-only marker path, for example `yaml_state_path` and `legacy_marker_path`. Operators must be able to tell whether `kas-project-state.yaml`, `kab-adoption-stage.md`, or both were read, and which surface will be written after an approved sync.

## 6. Relationship to KAB adoption stage

The YAML state file replaces the stage-only marker as the preferred static state surface, but the stage semantics remain unchanged:

- Stage 1: direct Codex app-server baseline; do not claim KAB Codex execution.
- Stage 2: KAB Codex-first via `native_codex`; direct Codex fallback requires recorded break-glass approval/rationale.
- Stage 3: reserved until separately authorized.

Fallback audit policy from `docs/sot/kas-cli-contract.md` applies regardless of stage.

## 7. Acceptance criteria for KASUPD implementation

KASUPD-002 may only close after it proves:

- state YAML schema parsing and validation;
- compatibility read of the old stage-only marker where needed;
- machine output that distinguishes YAML state paths from legacy marker paths;
- source commit/checksum capture;
- fail-closed behavior for schema, checksum, unsupported stage, auth/token/gateway/provider mutation, and ambiguous mapping cases;
- no-write behavior for the validation command.

Later KASUPD tasks may only close after they additionally prove:

- dry-run evidence integration;
- three-way classification;
- semantic-port prompt/evidence artifact generation;
- project-specific pilot evidence on one approved KAS suite such as KAN, KLM, `kan-plugin`, or `kan-control`.
