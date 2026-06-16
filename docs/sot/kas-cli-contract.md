# KAS minimum CLI contract

Date: 2026-05-26
Owner: KAS workflow/policy layer
Confirming role: Kkachi-team review accepted; 주유 harness review accepted; 사마의 final Red review accepted
Status: accepted SOT for `CLIMVP-001` and KABADOPT-001 stage-selector closure; implemented surfaces remain bounded by their roadmap evidence
Authority level: command-surface and manifest/checksum contract for the KHS+KAH minimum/pilot CLI lane
Scope: `kkachi-agent-skills` / KAS profile-scoped skill-pack `version`, `list`, `install --dry-run`, approved copy install, and `doctor`; no KAB runtime, KHC command/control, Doksuri integration, or KAH install-command expansion
Related docs: `docs/sot/minimum-pilot-cli-lane.md`, `docs/sot/interface-contract.md`, `docs/sot/khs-architecture-and-integration.md`, `docs/sot/project-specific-kas-install-contract.md`, `docs/README.md`, `docs/roadmap.md`, repository `README.md`
Evidence/source paths: KAH run `run-20260525T161641Z-4160f06cf1be`; populated artifacts include `intake-classification.md`, `sot-basis.md`, `plan.md`, `context-pack.md`, `docs-update.md`, `sot-update.md`, `roadmap-update.md`, `verification.md`, `review.md`, and `final-report.md`; review tasks are 하후연 `t_2fd07174`, 여몽 `t_e04116b6` plus re-review `t_abc30909`, 진궁 `t_b8db3ead` plus re-review `t_e652a24c`, 주유 `t_6cc40c59`, and 사마의 `t_e479171c`

## 1. Decision summary

`CLIMVP-001` defines a narrow KAS-owned CLI contract for profile-scoped skill-pack operations:

```text
kkachi-agent-skills --version
kkachi-agent-skills version [--json]
kkachi-agent-skills list [--profile <profile>] [--category <name>]
kkachi-agent-skills install --profile <profile> <pack-id>... --dry-run
kkachi-agent-skills install --profile <profile> <pack-id>... --approve <evidence-ref>
kkachi-agent-skills doctor --profile <profile> [--project <path>]
kkachi-agent-skills doctor --project <path> --workflow-graph --json
kkachi-agent-skills repair --project <path> --workflow-graph --propose --reason <reason> --json
kkachi-agent-skills repair --project <path> --workflow-graph --apply-proposal <proposal-id> --approval <approval-ref> --json
```

This CLI is a KAS minimum/pilot harness lane. It helps users inspect CLI release/build evidence, inspect available KAS skill packs, preview profile-scoped copies, perform approved copy installs, and verify installed state. It is not a Kkachi runner and must not control KAB sessions, KHC authority, Doksuri integration, backend execution, or KAH deterministic state beyond reading KAH availability/project status for `doctor` reporting.

GRSYNC-002 extends `doctor` with a project workflow graph supportability mode: `doctor --project <path> --workflow-graph --json`. In that mode `--project` is a filesystem path, `--profile` is optional, `--project-suite` is rejected, and the command is read-only. It reports JSON fields `ok`, `command`, `mode: workflow_graph_doctor`, `no_write:true`, `status`, project graph path/presence, KAS compatibility registry readback, KAH version/capabilities/help evidence, KAH graph validate/explain evidence when applicable, diagnostics, reason codes, remediation, and next action. It may classify `pass`, `custom_supported`, `update_kah_required`, `update_kah_recommended`, `update_kas_recommended`, `graph_missing`, `graph_stale`, `graph_broken`, `graph_conflict`, or `unsupported`; only `pass`, `custom_supported`, and non-blocking future update recommendations are success exits. It must not call KAH graph init/diff/propose/apply/export or write `.kkachi-workflow.yaml`, `.kkachi/`, profile files, KAH state, KAB state, auth, token, provider, gateway, model, or project source files.

GRSYNC-003 extends `repair` with a separate workflow graph mode. `repair --project <path> --workflow-graph --propose --reason <reason> --json` always runs the read-only workflow graph doctor first, refuses safe/no-op or unsupported statuses including `pass`, `custom_supported`, update-required/recommended states, conflicts, and unsupported graphs, generates a complete candidate graph file from reviewed KAS template input, calls KAH `graph diff` when a valid base graph exists, calls KAH `graph propose`, and reports `mode: workflow_graph_repair_propose`, `status: proposal_available`, candidate checksum/path, semantic diff state, proposal id/path, approval requirement, risk flags, reason codes, and a safe next command. Missing or invalid base graphs record `semantic_diff.state: not_applicable_missing_or_invalid_base` instead of fabricating a diff. Proposal mode is not read-only: it may create candidate/proposal evidence, but it must not apply or directly edit `.kkachi-workflow.yaml`.

`repair --project <path> --workflow-graph --apply-proposal <proposal-id> --approval <approval-ref> --json` is the safer approved equivalent for graph apply. It does not use project-suite lifecycle `--apply dry-run:sha256:<hash>` and cannot be combined with project-suite repair flags such as `--profile`, `--dry-run`, `--approve`, `--apply`, `--source-pack`, or `--profile-root`. Apply mode reruns doctor and full KAH graph capability preflight, calls only KAH `graph apply --proposal <proposal-id> --approval <approval-ref> --json`, then reruns KAH validate/explain. JSON reports `mode: workflow_graph_repair_apply`, proposal id, approval ref, audit/backup/recovery fields returned by KAH, post-apply checksum/version, validation/explain states, diagnostics, and next action.

WFLOW-002 extends the CLI with a generic explicit task-DAG trigger:

```text
kkachi-agent-skills workflow-trigger --project <path> --workflow-id <id> --node-contract-source <path> [--node-contract-ref <ref>] [--run <run-id>] [--instance-id <id>] --json
```

This command supports only explicit `workflow_id` plus explicit JSON node-contract source/ref. KAS resolves the workflow id to `.kkachi/workflows/<workflow-id>.yaml`, preflights KAH with `--version`, `capabilities --json`, and `workflow --help`, requires KAH workflow subcommands `validate`, `explain`, `create`, `show`, `ready`, and `node`, validates/explains the workflow through KAH, creates with `workflow create --run <run-id> --file <workflow.yaml>` or resumes with `workflow show --run <instance-id>`, reads ready nodes with `workflow ready --run <id>`, and renders dispatch packets only. The result always includes `direct_kah_state_write:false`; KAS must not directly edit KAH workflow instance files or complete/block/start nodes.

WFLOW-002 node-contract input is JSON-only:

```json
{
  "schema_version": "kas-node-contracts/v1",
  "ref": "optional-source-ref",
  "contracts": [
    {
      "workflow_id": "demo",
      "node_id": "setup",
      "owner_role": "implementer_backend",
      "execution_lane": "direct_kas_skill",
      "required_inputs": ["task-contract.yaml"],
      "expected_artifacts": ["artifacts/setup.md"],
      "prompt_ref": "skills/kkachi-implement/SKILL.md",
      "approval_required": true,
      "fallback_policy": "none_fail_closed",
      "verification_gate": "make test"
    }
  ]
}
```

Every ready node must match one contract by `workflow_id` and `node_id`; missing or invalid source/ref, missing KAH workflow capability, KAH validation failure, and missing ready-node contracts fail closed with JSON and no partial packets. `ok:true/status:no_ready_nodes` is the successful no-packet state. Selector search and the full node-contract registry are WFLOW-003 deferrals. Custom workflow creation, dynamic node generation, thin trigger scaffolding, retry/rollback automation, arbitrary webhook runtime, automatic backend/agent fallback, direct KAH state writes, KAB graph authority, and Hermes profile/provider/gateway/auth/token/model mutation are outside WFLOW-002 and remain deferred.

WFLOW-003 extends `workflow-trigger` with deterministic selector mode:

```text
kkachi-agent-skills workflow-trigger --project <path> --selector-registry <path> --task-class <class> [--labels <csv>] [--changed-surfaces <csv>] [--risk <level>] [--required-agent <csv>] [--required-capability <csv>] [--run <run-id>] [--instance-id <id>] --json
```

Selector mode reads `kas-task-dag-workflow-registry/v1` YAML from the explicit
`--selector-registry` path. The shipped registry lives at
`registries/task-dag-workflow-registry.yaml`. It
evaluates only deterministic predicates and returns `selector_no_match` for
zero candidates, `selector_ambiguous` for multiple candidates, and
`selector_matched` only for exactly one candidate. Zero and multiple candidate
states must not call KAH workflow instance commands. Combining selector inputs
with explicit `--workflow-id` or `--node-contract-source` fails closed with
`selector_explicit_mode_conflict`; there is no choose-first, scoring, ranking,
LLM tie-break, backend fallback, or agent fallback behavior.

Selector dispatch packets include selector registry path/checksum, selector
match readback, task class, labels, changed surfaces, required capabilities,
`completion_authority:kah_only`, `stage1_direct_codex_is_kab_native_codex:false`,
and `direct_kah_state_write:false`. KAS may render dispatch packets only; KAH
remains the authority for node state transitions and completion.

WFLOW-004 extends the CLI with a dry-run-first custom workflow creator:

```text
kkachi-agent-skills workflow-create --project <path> --workflow-id <id> --mode dag_only|thin_trigger|full_trigger --request <json-path> --dry-run --json
kkachi-agent-skills workflow-create --project <path> --workflow-id <id> --mode dag_only|thin_trigger|full_trigger --request <json-path> --apply dry-run:sha256:<hash> --json
```

The request is JSON-only with schema `kas-workflow-create-request/v1`.
`dag_only` emits candidate DAG, catalog, and node-contract registry content.
`thin_trigger` adds a bounded trigger `SKILL.md` wrapper that delegates to
`workflow-trigger`. `full_trigger` is exceptional and requires an explicit
reason. Dry-run emits compact operator output plus a full machine packet with
candidate paths, `target_paths`, `generated_content`, `selector_metadata`,
KAS/KAH `capability_evidence`, base checksums, `changed_paths`,
conflicts/diagnostics, no-write evidence, and an approval hash.

The approval hash is canonical `sha256` over deterministic UTF-8 JSON with
sorted keys and normalized project-relative paths. The hash scope is represented
directly by command, mode, `target_paths`, `changed_paths`,
`generated_content`, selector metadata, capability evidence, base checksums,
conflicts/diagnostics, and no-write evidence. Apply recomputes that packet and
fails closed on malformed, stale, mismatched, blocked, or non-approvable
evidence before any write or KAH delegation.
Missing, null, or empty `selector_metadata` fails closed with
`selector_metadata_invalid` because workflow creation must not proceed without a
deterministic selector contract.

KAS generates candidate content and approval packets only. KAH remains
authoritative for workflow/catalog validation, proposal, apply, audit, final
gate integration, and node-contract registry evidence. Installed KAH `0.1.9` lacks the workflow command group. Source-built DAGSM-003 KAH currently
advertises `workflow catalog validate/explain` and node-contract registry
diagnostics, but does not advertise a reviewed workflow catalog proposal/apply command mapping. Therefore `workflow-create --apply` remains fail-closed with
`blocked_missing_kah_workflow_catalog_capability` until an effective KAH binary
advertises a reviewed apply surface. KAS must not direct-write `.kkachi`
workflow state as a fallback.

KAH remains the deterministic project-local state/evidence layer and currently advertises `install_command=false`. Therefore KAS owns this profile-scoped list/install/doctor surface; KAH must not be described as the skill installer.

## 2. SOT basis

This contract is derived from these current sources of truth:

- `minimum-pilot-cli-lane.md`: keeps the KHS+KAH minimum/pilot harness lane distinct from the full KHS+KAH+KAB execution-runtime lane; permits profile-scoped release evidence plus `list`, `install`, `doctor`, `sync`, and `proposal`, while this CLIMVP task scopes only the implemented MVP surfaces from `roadmap.md`.
- `interface-contract.md`: records KAH 0.1.4 `install_command=false`, KAS/KAH/KAB layer ownership, reality-first command evidence, and the rule that unsupported future surfaces must be recorded as gaps rather than claimed.
- `khs-architecture-and-integration.md`: defines KAS as the skill/process layer, KAH as deterministic state/evidence, KAB as backend runtime, and the profile-scoped installed-copy model.
- `docs/README.md`: defines the documentation authority ladder, status vocabulary, and promotion rule for candidate SOTs.
- Repository `README.md`: states that the future KAS install path must default to copy mode with dry-run, manifest/checksum, changed-path report, and recovery instructions.

## 3. Non-goals and hard boundaries

The KAS minimum CLI must not:

- provide a `run` verb;
- start, stop, read, approve, or otherwise control KAB/backend sessions;
- mutate KAH project state, `.kkachi-workflow.yaml`, project overlays, gateway config, auth, tokens, or Hermes model/provider config;
- use `skills.external_dirs` as the default install path;
- use symlinks as the default install mode;
- write into a profile without dry-run evidence and explicit approval evidence;
- claim that installing the repository root through native Hermes installs every KAS skill pack unless separately evidenced;
- silently overwrite profile-local edits or unknown installed content;
- treat KAH `project doctor` as proof that profile skills are installed correctly;
- treat KAS semantic docs as proof of backend/KAB capability.

## 4. Terms

- **Source repo:** the default embedded `kkachi-agent-skills` source bundled into the Go binary, or an explicit local checkout selected with `--repo <path>`, containing KAS `skills/`, `registries/`, `templates/`, and docs.
- **Source pack:** one installable KAS skill directory under `skills/<category>/<skill>/` or the approved equivalent pack layout if this repo later adopts a different physical layout.
- **Pack id:** stable logical id for a source pack, normally `<category>/<skill>`.
- **Target profile:** a Hermes profile directory under `~/.hermes/profiles/<profile>/`.
- **Target install path:** copy destination under `~/.hermes/profiles/<profile>/skills/<category>/<skill>/`.
- **Install manifest:** KAS-owned profile metadata recording what was copied, where it came from, and checksums.
- **Approval evidence:** explicit user/approver reference that authorizes writes after reviewing dry-run changed paths.

Project-specific KAS install vocabulary is defined canonically by `docs/sot/project-specific-kas-install-contract.md` for KASPROJ-001 and later tasks:

- **Project:** concrete project suite id, for example `doksuri-server`, used in `skills/<project>/`.
- **Source pack:** upstream KAS pack/template input used to render a tailored project suite.
- **Installed skill:** project-prefixed skill id such as `doksuri-server-plan`; generic ids such as `kkachi-plan` are invalid for project suites.
- **Target path:** profile-relative project-suite destination such as `skills/doksuri-server/doksuri-server-plan/SKILL.md`.
- **Drift policy:** fail-closed/manual-review/semantic-port/repair-after-approval policy for source/profile language and checksum drift.

For project-specific suite installs, the target path must use:

```text
skills/<project>/<project>-<phase-or-skill>/SKILL.md
```

This extends the profile-scoped manifest vocabulary. The current CLI implements KASPROJ-002 read-only project-suite dry-run planning, KASPROJ-003 approval-hash-bound project-suite install, and KASPROJ-004 project-suite doctor/repair/migrate. It still must not be described as performing semantic-port completion, operational rollout, generic fallback, or KAB runtime activation.

## 5. First-run operator path

The recommended first-run operator path is:

```bash
kkachi-agent-skills --version
kkachi-agent-skills list --profile <profile>
kkachi-agent-skills install --profile <profile> <pack-id>... --dry-run
# operator reviews changed_paths, counts, conflicts, and dry_run_plan_hash
kkachi-agent-skills install --profile <profile> <pack-id>... --approve dry-run:<dry_run_plan_hash>
kkachi-agent-skills doctor --profile <profile> [--project <path>]
```

Operator UX decision:

- The stable harness contract is the explicit `--profile <profile>` form because it is unambiguous and keeps profile identity in a named field.
- The older SOT/README form `kkachi-agent-skills install <profile> <skill-or-category> --dry-run` may be supported as a human convenience alias, but it must resolve to the same internal contract and JSON shape.
- `skill-or-category` is operator-facing wording. It resolves to one or more canonical `pack_id` values. A concrete skill resolves to one pack; a category resolves to all approved packs under that category.
- The CLI must print the canonical command equivalent in dry-run output so the operator can copy/paste the approved next command safely.

## 6. Command surface

### 6.1 Global flags

All commands should support:

```text
--json                              emit machine-readable JSON
--repo <path>                       explicit local KAS checkout override; default is embedded source bundled in the installed binary
--profile <profile>                Hermes target profile name where relevant
--profile-root <path>               explicit profile root override for tests/harness only
--kab-stage <1|2>                   select KAS/KAH development KAB adoption stage by numeric operator choice
--kab-adoption-stage <canonical>    select the same stage by canonical stage name
--no-color                          disable color in human output
```

Rules:

- `--version` and `version` must not require source-repo discovery, profile access, KAH state, or KAB runtime availability.
- `version --json` must include stable `ok`, `command`, `cli_version`, `module_path`, `module_version`, `git_commit`, and `dirty` fields. Build metadata may be empty or `(devel)` when Go does not provide VCS/module data for a local development binary.
- `--profile-root` is for tests/harness and must be rejected in normal production use unless paired with an explicit test/harness mode or documented environment guard.
- Human output must be Korean-friendly in summaries, but JSON field names stay stable English.
- JSON output must include `ok`, `command`, `source_repo`, `target_profile` when applicable, `changed_paths` when applicable, `diagnostics`, and `next_action`.
- When a command accepts or resolves a KAB adoption stage, JSON output must include `kab_adoption_stage` with the numeric value, canonical value, source (`explicit_numeric`, `explicit_canonical`, `interactive`, or `default_stage1`), and whether the value is hash-bound.

### 6.2 Minimum human output examples

Human output is not the stable harness contract, but each command must include enough Korean-friendly operator context to avoid unsafe copy/paste.

`list --profile <profile>` example:

```text
상태: 조회 완료 — profile hwangchung 기준 KAS pack 12개 발견.
설치 상태: current 4, missing 8, drifted 0, conflict 0.
소스: /repo/kkachi-agent-skills @ <git-sha>
다음: 설치 전 `install --profile hwangchung <pack-id> --dry-run`으로 변경 경로를 확인하세요.
```

`install --dry-run` example:

```text
상태: 승인 필요 — profile hwangchung에 1개 pack / 3개 file 복사 예정.
변경: create 3, update 0, skip 0, conflict 0, error 0.
증거: dry-run:<dry_run_plan_hash>
주의: 아직 파일을 쓰지 않았습니다. changed_paths를 확인한 뒤 승인하면 아래 명령을 실행하세요.
다음: kkachi-agent-skills install --profile hwangchung software-development/kas-codex-roadmap-development --approve dry-run:<dry_run_plan_hash>
```

Approved install example:

```text
상태: 설치 완료 — profile hwangchung에 1개 pack / 3개 file 반영.
변경: create 3, update 0, backup 0, manifest_update 1.
manifest: ~/.hermes/profiles/hwangchung/.kas/skill-pack-manifest.json
복구: backup_path 없음(create-only). update가 있으면 backup_path를 함께 출력합니다.
다음: kkachi-agent-skills doctor --profile hwangchung
```

`doctor` example:

```text
상태: 정상 — profile hwangchung KAS 설치 상태가 manifest/checksum과 일치합니다.
KAH: 사용 가능, install_command=false 확인.
KAB: minimum CLI에는 필요 없음. 코드 변경/백엔드 실행 KAS run에는 KAB가 필요합니다.
다음: 실행-runtime 작업은 KAS+KAH+KAB 경로로 시작하세요.
```


### 6.2.1 KAB adoption stage selector

Scope:

- The stage selector applies to KAS/KAH development project packs or project-specific KAS overlays that need to record whether future KAS/KAH implementation work uses Stage 1 direct Codex app-server or Stage 2 KAB Codex-first execution.
- The selector does not make the minimum CLI a KAB runner and does not authorize Stage 2 migration by itself. It records the operating policy that later KAS runs must obey.
- Stage 3 backend-selected execution is intentionally not exposed by this CLI selector until a separate SOT/roadmap task authorizes it.

Canonical stage values:

| Numeric choice | Canonical value | Meaning | Current exposure |
|---|---|---|---|
| `1` | `stage1_direct_codex_app_server_baseline` | KAS/KAH development continues through the direct Codex app-server lane and records no-KAB-Codex rationale. | default and selectable |
| `2` | `stage2_kab_codex_first` | KAS/KAH development replaces direct Codex calls with KAB `native_codex` while preserving the same KAS/KAH phase and review structure. | selectable only after operator chooses it |
| `3` | `stage3_kab_backend_selected` | KAS/KAH development selects among eligible KAB backends after capability and policy gates. | reserved; not shown or accepted by this selector |

Flags:

```bash
kkachi-agent-skills install --profile <profile> <project-kas-pack-id> --dry-run --kab-stage 1
kkachi-agent-skills install --profile <profile> <project-kas-pack-id> --dry-run --kab-stage 2
kkachi-agent-skills install --profile <profile> <project-kas-pack-id> --dry-run --kab-adoption-stage stage1_direct_codex_app_server_baseline
kkachi-agent-skills install --profile <profile> <project-kas-pack-id> --dry-run --kab-adoption-stage stage2_kab_codex_first
```

Rules:

- `--kab-stage` and `--kab-adoption-stage` are aliases for the same selection and must not conflict. Supplying both with different meanings fails closed.
- Explicit `--kab-stage 3` or `--kab-adoption-stage stage3_kab_backend_selected` fails closed until Stage 3 is separately authorized.
- Explicit unknown, malformed, or ambiguous stage values fail closed with operator-readable diagnostics.
- When the selector is relevant and no explicit stage is supplied in an interactive TTY, human output asks:

```text
KAB adoption stage for this KAS/KAH project pack:
  [1] Stage 1 — direct Codex app-server baseline (default)
  [2] Stage 2 — KAB Codex-first via native_codex
Choice [1]:
```

- Blank interactive input selects Stage 1.
- Non-interactive mode, `--json`, CI, or unavailable stdin must not prompt. Missing stage in those modes resolves to Stage 1 with source `default_stage1`.
- Human output must clearly state when Stage 1 was defaulted so the operator does not mistake a silent default for Stage 2 migration.

Marker output:

- The selected/defaulted stage is recorded in installed project-specific KAS guidance. `references/kab-adoption-stage.md` is the legacy stage-only marker; new project-specific KAS sync work should record the stage inside `references/kas-project-state.yaml` as defined by `docs/sot/project-kas-sync-state.md`.
- The marker must include the numeric choice, canonical value, selection source, selected-at/install id when available, responsible approval evidence when available, and the required evidence posture for future runs.
- The marker is an operating-policy reference, not KAH state and not KAB execution evidence.
- Stage 2 markers must state that future implementation/fix/docs-bound execution uses KAB `native_codex` and that falling back to direct Codex requires a recorded break-glass approval/rationale.
- If the marker is absent or unreadable for a project-specific KAS pack, KAS must fail closed to Stage 1 claims: it may record direct Codex evidence, but it must not claim KAB Codex execution.

Plan/review fallback audit:

- Any Codex or KAB-authored plan for KAS/KAH development must include a fallback audit note before Blue/Red plan review. The note must identify proposed fallback paths, explain whether each is necessary, and prefer fail-closed behavior when the system lacks capability, evidence, approval, or safe state.
- Blue and Red plan review, first color review, GLM/Octo review when run, and post-review closure must explicitly check for unnecessary fallback paths. The normal requested change is to remove the fallback and fail closed with operator-readable diagnostics and KAH/KAB evidence.
- A fallback may be accepted only when there is no safe direct handling, the fallback is explicitly bounded and evidenced, and the implementation/docs delta is genuinely small. The fallback must not bypass KAH gates, approval checks, KAB evidence requirements, auth/token safety, or the selected KAB adoption stage.
- If handling the condition without a fallback is impossible and the fallback would require broad code, new state machinery, or unclear policy, the run must stop and report options to 주군 instead of quietly adding the fallback.

Minimum marker content:

```markdown
# KAB adoption stage

Numeric stage: 1
Canonical stage: stage1_direct_codex_app_server_baseline
Selection source: default_stage1
Selected at: <install-id-or-timestamp>
Approval evidence: <approval-ref-or-not_applicable>

## Runbook references

- Canonical runbook: skills/kkachi-install-guide/references/kas-kab-adoption-stage-runbook.md
- Generated marker path: references/kab-adoption-stage.md

## Execution policy

Future KAS/KAH development runs for this project use the direct Codex app-server baseline unless a later approved reconfiguration changes this marker to Stage 2.

## Selected-stage evidence posture

- Record direct Codex app-server prompt/session/output evidence for plan, implementation, docs/fix, cleanup, and verification support.
- Record the no-KAB-Codex rationale for the selected direct lane.
- Do not claim KAB Codex execution evidence for Stage 1 work.

This marker is operating-policy guidance only. It is not Stage 2 activation by itself and is not KAB execution evidence.
```

JSON shape extension for dry-run, approved install, and doctor where relevant:

```json
"kab_adoption_stage": {
  "applicable": true,
  "numeric": 1,
  "canonical": "stage1_direct_codex_app_server_baseline",
  "source": "default_stage1",
  "hash_bound": true,
  "marker_path": "skills/<project>/<project>-kas/references/kab-adoption-stage.md"
}
```

`doctor --profile <profile> [--project <path>] --json` should report the installed marker state for project-specific KAS packs when discoverable. It must distinguish `not_applicable`, `marker_present`, `marker_missing`, `marker_unreadable`, and `unsupported_stage`.

### 6.2.2 Project KAS dry-run classification read

KASUPD-002 added a bounded read-only validation surface for project-specific KAS state. KASUPD-003 extends the same `sync-project-kas --dry-run` surface: after state validation succeeds, the command performs read-only three-way classification and emits semantic-port packet content as JSON evidence.

```bash
kkachi-agent-skills sync-project-kas \
  --profile <profile> \
  --project <project-id> \
  --state skills/<project>/<project>-kas/references/kas-project-state.yaml \
  [--repo <current-upstream-kas-repo>] \
  [--project-root <project-specific-kas-root>] \
  --dry-run \
  --json
```

Purpose:

- read and validate `kas-project-state.yaml` using the schema in `docs/sot/project-kas-sync-state.md`;
- optionally read the sibling legacy `references/kab-adoption-stage.md` marker for compatibility reporting;
- report the effective static stage claim without claiming KAB runtime execution;
- compare current upstream KAS packs, recorded upstream baseline checksums, and mapped project-local skills without checkout mutation;
- classify changed mapped packs with the documented six-value vocabulary;
- represent unchanged mapped packs separately as `unchanged_mappings` and `summary.no_action_count`;
- generate semantic-port packet JSON content for `semantic_merge_required` and `new_upstream_candidate` classifications;
- perform no writes, no profile install, no KAH state mutation, and no KAB/backend/session control.

Rules:

- `--dry-run` is mandatory. Missing `--dry-run` fails closed.
- `--profile`, `--project`, and `--state` are required. CLI `--profile` and `--project` must match the YAML `project.profile` and `project.id` when the YAML is readable.
- Valid YAML may recognize Stage 1 or Stage 2 static state values. Stage 3 remains reserved and unsupported.
- YAML state is not KAB runtime evidence and is not Stage 2 activation by itself.
- Missing, unreadable, unsupported, or schema-invalid YAML fails closed to Stage 1 claims and must not claim KAB Codex execution.
- The legacy marker can be read for compatibility/reporting, but legacy-only evidence must not upgrade missing or invalid YAML to a valid state.
- The parser intentionally supports the documented scalar/list YAML subset only. Unsupported YAML features fail closed with diagnostics rather than adding an external dependency or accepting ambiguous state.
- State input must not contain auth tokens, secrets, gateway credentials, provider keys, model credentials, or mutable runtime/session state.
- Dirty current upstream source fails closed for KASUPD-003 unless explicit approval exists.
- Baseline pack checksum verification must use git object reads or an equivalent internal read abstraction. It must not mutate checkout state, overwrite the worktree, or write temporary baseline files into the project.
- Project skill path inference must fail closed on ambiguity, path traversal, symlink escape, or missing expected mapped local skill.
- Any `fail_closed_conflict` makes the command return `ok:false` with a non-zero exit while still emitting useful JSON diagnostics.
- The required classification vocabulary is exactly `auto_copy_candidate`, `local_only`, `semantic_merge_required`, `new_upstream_candidate`, `removed_or_renamed_upstream`, and `fail_closed_conflict`. Do not add no-op/unchanged as a seventh classification without a new SOT update.
- Semantic-port packet content is evidence only. Human output summarizes packet counts; the command must not write packet files.

TOKEN-004 public lifecycle wrapper:

```bash
kkachi-agent-skills update \
  --profile <profile> \
  --project <project-id> \
  --state skills/<project>/<project>-kas/references/kas-project-state.yaml \
  [--repo <current-upstream-kas-repo>] \
  [--project-root <project-specific-kas-root>] \
  --dry-run \
  --json
```

`update --dry-run` is the public project KAS sync command. It exposes the same
read-only classifier as `sync-project-kas`. JSON output reports target
roles/profiles, source packs, skill ids, target paths, checksums, planned
states, changed paths, backup/recovery posture, doctor commands, and explicit
zero-write evidence for profile writes, auth/token/gateway/provider or model
mutation, KAB runtime/policy mutation, KAH subjective-judgment mutation, Hermes
runtime mutation, and profile activation. `sync-project-kas` remains available
as a compatibility command.

Minimum JSON shape:

```json
{
  "ok": true,
  "command": "sync-project-kas",
  "mode": "dry_run_classification",
  "dry_run": true,
  "target_profile": "hwangchung",
  "project_id": "kan-plugin",
  "yaml_state_path": "skills/kan-plugin/kan-plugin-kas/references/kas-project-state.yaml",
  "legacy_marker_path": "skills/kan-plugin/kan-plugin-kas/references/kab-adoption-stage.md",
  "state_source": "yaml",
  "read_surfaces": {
    "yaml": {"state": "valid", "path": "skills/kan-plugin/kan-plugin-kas/references/kas-project-state.yaml"},
    "legacy_marker": {"state": "missing", "path": "skills/kan-plugin/kan-plugin-kas/references/kab-adoption-stage.md"}
  },
  "effective_stage_claim": {
    "numeric": 1,
    "canonical": "stage1_direct_codex_app_server_baseline",
    "source": "yaml",
    "kab_execution_claim_allowed": false,
    "fail_closed_to_stage1": false
  },
  "write_target_after_approved_sync": "yaml_state_path",
  "validation": {
    "schema_version": "0.1",
    "pack_baseline_count": 1,
    "diagnostics": []
  },
  "source_repo": {
    "path": "embedded://github.com/SeventeenthEarth/kkachi-agent-skills@<version-or-checksum>",
    "git_commit": "<current-sha-or-null>",
    "dirty": false
  },
  "baseline_repo": {
    "git_commit": "<state.upstream_kas.commit>",
    "dirty_recorded": false,
    "baseline_verified": true
  },
  "project_root": {
    "path": "/project-root",
    "resolution": "explicit",
    "state": "resolved"
  },
  "summary": {
    "total_mappings": 1,
    "counts_by_classification": {
      "auto_copy_candidate": 0,
      "local_only": 0,
      "semantic_merge_required": 1,
      "new_upstream_candidate": 0,
      "removed_or_renamed_upstream": 0,
      "fail_closed_conflict": 0
    },
    "no_action_count": 0,
    "semantic_port_packet_count": 1,
    "write_count": 0
  },
  "classifications": [
    {
      "id": "kas-sync-item-0001",
      "upstream_pack": "kkachi-plan",
      "project_skill": "kan-plugin-plan",
      "classification": "semantic_merge_required",
      "basis": [
        "project_skill_mapping_exists",
        "local_changed_since_baseline",
        "upstream_changed_since_baseline"
      ],
      "paths": {
        "baseline_upstream_path": "skills/kkachi-plan",
        "current_upstream_path": "skills/kkachi-plan",
        "project_skill_path": "skills/kan-plugin/kan-plugin-plan"
      },
      "checksums": {
        "recorded_source_checksum": "sha256:<hex>",
        "computed_baseline_source_checksum": "sha256:<hex>",
        "current_source_checksum": "sha256:<hex>",
        "recorded_project_checksum": "sha256:<hex>",
        "current_project_checksum": "sha256:<hex>"
      },
      "requires_semantic_port": true,
      "diagnostics": []
    }
  ],
  "unchanged_mappings": [],
  "semantic_port_packets": [
    {
      "packet_id": "semantic-port-kas-sync-item-0001",
      "classification_id": "kas-sync-item-0001",
      "recommended_artifact_path": ".kas/dry-runs/semantic-port-kas-sync-item-0001.md",
      "content_sha256": "sha256:<hex>",
      "content": "<semantic-port packet content>"
    }
  ],
  "next_action": "Review dry-run classifications and semantic-port packets; no files were written."
}
```

For invalid or missing YAML, JSON output must keep `ok: false`, exit non-zero, include both `yaml_state_path` and `legacy_marker_path`, set `effective_stage_claim.fail_closed_to_stage1: true`, and keep `write_target_after_approved_sync: "yaml_state_path"`.

### 6.3 `list`

Canonical form:

```bash
kkachi-agent-skills list [--profile <profile>] [--category <name>] [--json]
```

Purpose:

- list available source KAS packs without mutating any files;
- optionally compare against a target profile manifest/install state;
- distinguish source repo packs from installed profile copies.

Required behavior:

- Read source packs from repo `skills/`.
- Fail closed if the source repo cannot be identified or has no readable pack metadata.
- With `--profile`, read installed manifest if present and report `not_installed`, `installed_current`, `installed_drifted`, `installed_unknown`, or `manifest_missing`.
- With `--category`, filter source and installed records by category.
- Do not require KAB.
- Do not require KAH unless `--project` is added in a future extension; KAH availability may be reported as informational only.

Minimum JSON shape:

```json
{
  "ok": true,
  "command": "list",
  "source_repo": {
    "path": "embedded://github.com/SeventeenthEarth/kkachi-agent-skills@<version-or-checksum>",
    "git_commit": "<sha-or-null>",
    "dirty": false
  },
  "target_profile": {
    "name": "hwangchung",
    "root": "/Users/name/.hermes/profiles/hwangchung",
    "manifest_path": "/Users/name/.hermes/profiles/hwangchung/.kas/skill-pack-manifest.json",
    "manifest_state": "manifest_present"
  },
  "packs": [
    {
      "pack_id": "software-development/kas-codex-roadmap-development",
      "category": "software-development",
      "name": "kas-codex-roadmap-development",
      "source_path": "skills/software-development/kas-codex-roadmap-development",
      "installed_state": "installed_current",
      "installed_path": "skills/software-development/kas-codex-roadmap-development"
    }
  ],
  "diagnostics": [],
  "next_action": "Run install --dry-run before any profile writes."
}
```

### 6.4 `install --dry-run`

Canonical form:

```bash
kkachi-agent-skills install --profile <profile> <pack-id>... --dry-run [--json]
```

Optional convenience forms may be added later only if they resolve to the same internal contract:

```bash
kkachi-agent-skills install <profile> <pack-id> --dry-run
kkachi-agent-skills install --profile <profile> --category <category> --dry-run
```

Purpose:

- resolve source packs and target profile paths;
- validate manifest/checksum inputs;
- report creates, updates, conflicts, skips, and failures;
- perform no writes.

Required behavior:

- Exit non-zero if `--dry-run` is missing and no approval flag/evidence is supplied.
- Refuse path traversal or symlink escape from source repo or target profile.
- Refuse unknown profile unless `--profile-root` is test/harness-authorized.
- Refuse unknown pack ids.
- Compute checksums for all source files that would be copied.
- Compare target files when they already exist.
- Report planned backup paths for any update/delete risk, but do not create backups in dry-run.
- Treat profile-local modifications as `conflict` unless an explicit future resolution flag is designed and reviewed.
- Never mutate `skills.external_dirs`, gateway config, profile model/provider config, auth, tokens, or KAH/KAB config.

Dry-run changed-path categories:

| Category | Meaning | Write allowed later? |
|---|---|---|
| `create` | target file/dir absent and source is valid | yes after approval |
| `update` | target installed by KAS and source checksum changed | yes after approval and backup plan |
| `skip` | target already matches source checksum | no write needed |
| `conflict` | target exists but manifest absent/mismatched or local edit detected | no, fail closed |
| `error` | invalid source, path, profile, checksum, or manifest | no |

Minimum JSON shape:

```json
{
  "ok": true,
  "command": "install",
  "mode": "dry_run",
  "source_repo": {"path": "/repo", "git_commit": "<sha>", "dirty": false},
  "target_profile": {"name": "hwangchung", "root": "/profile"},
  "kab_adoption_stage": {"applicable": true, "numeric": 1, "canonical": "stage1_direct_codex_app_server_baseline", "source": "default_stage1", "hash_bound": true, "marker_path": "skills/<project>/<project>-kas/references/kab-adoption-stage.md"},
  "summary": {
    "total_packs": 1,
    "total_files": 1,
    "counts_by_action": {"create": 1, "update": 0, "skip": 0, "conflict": 0, "error": 0},
    "conflict_count": 0
  },
  "dry_run_plan_hash": "sha256:<canonical-plan-json-hash>",
  "packs": [
    {
      "pack_id": "software-development/kas-codex-roadmap-development",
      "source_path": "skills/software-development/kas-codex-roadmap-development",
      "target_path": "skills/software-development/kas-codex-roadmap-development",
      "installed_state": "not_installed",
      "files": [
        {
          "relative_path": "SKILL.md",
          "action": "create",
          "sha256": "<hex>",
          "bytes": 1234
        }
      ]
    }
  ],
  "changed_paths": [
    {"path": "skills/software-development/kas-codex-roadmap-development/SKILL.md", "action": "create"}
  ],
  "approval_request": {
    "required": true,
    "summary": "Approve copying 1 pack / 1 file into profile hwangchung.",
    "evidence_ref": "dry-run:<dry_run_plan_hash>",
    "dry_run_plan_hash": "sha256:<canonical-plan-json-hash>"
  },
  "diagnostics": [],
  "next_action": "Review changed_paths, then rerun with --approve <evidence-ref>."
}
```

### 6.5 Approved copy install

Canonical form:

```bash
kkachi-agent-skills install --profile <profile> <pack-id>... --approve <evidence-ref> [--json]
```

Purpose:

- copy source packs into the target profile only after operator approval;
- record manifest/checksum evidence;
- report actual changed paths and recovery instructions.

Approval rules:

- `--approve <evidence-ref>` must reference the prior dry-run report or a user/approver decision that includes the changed-path summary.
- If the current computed dry-run plan differs from the approved evidence, fail closed and require a new dry-run.
- Approval does not allow conflicts, path escapes, profile-root ambiguity, symlink mode, `skills.external_dirs`, or auth/gateway/model mutations.

Write order:

1. Recompute and validate the dry-run plan.
2. Create backup files for paths that will be updated, under a KAS-owned backup directory outside `skills/`.
3. Copy files atomically where possible.
4. Write/update the install manifest after copied files are in place.
5. Emit actual changed paths, manifest path, backup path, and rollback instructions.

Recommended profile metadata paths:

```text
~/.hermes/profiles/<profile>/.kas/skill-pack-manifest.json
~/.hermes/profiles/<profile>/.kas/backups/<timestamp-or-install-id>/
```

Rationale: KAS metadata should not appear as a Hermes skill directory under `skills/`.

Minimum JSON shape:

```json
{
  "ok": true,
  "command": "install",
  "mode": "approved_copy",
  "source_repo": {"path": "/repo", "git_commit": "<sha>", "dirty": false},
  "target_profile": {"name": "hwangchung", "root": "/profile"},
  "kab_adoption_stage": {"applicable": true, "numeric": 1, "canonical": "stage1_direct_codex_app_server_baseline", "source": "default_stage1", "hash_bound": true, "marker_path": "skills/<project>/<project>-kas/references/kab-adoption-stage.md"},
  "approval": {
    "evidence_ref": "dry-run:<hash>",
    "dry_run_plan_hash": "sha256:<canonical-plan-json-hash>",
    "approved_plan_hash": "sha256:<canonical-plan-json-hash>",
    "matched_current_plan": true
  },
  "summary": {
    "total_packs": 1,
    "total_files": 1,
    "counts_by_action": {"create": 1, "update": 0, "backup": 0, "manifest_update": 1},
    "conflict_count": 0
  },
  "install_id": "kas-install-20260526T000000Z-abcdef",
  "manifest_path": "/profile/.kas/skill-pack-manifest.json",
  "backup_path": "/profile/.kas/backups/kas-install-20260526T000000Z-abcdef",
  "packs": [
    {
      "pack_id": "software-development/kas-codex-roadmap-development",
      "source_path": "skills/software-development/kas-codex-roadmap-development",
      "target_path": "skills/software-development/kas-codex-roadmap-development",
      "files": [
        {"relative_path": "SKILL.md", "action": "create", "sha256": "<hex>", "bytes": 1234}
      ]
    }
  ],
  "changed_paths": [
    {"path": "skills/software-development/kas-codex-roadmap-development/SKILL.md", "action": "create", "sha256": "<hex>"}
  ],
  "recovery": {
    "rollback_supported": true,
    "backup_path": "/profile/.kas/backups/kas-install-20260526T000000Z-abcdef",
    "previous_manifest_sha256": "<sha256-or-null>",
    "instructions": ["Restore files from backup_path", "Restore previous manifest snapshot"]
  },
  "diagnostics": [],
  "next_action": "Run kkachi-agent-skills doctor --profile hwangchung."
}
```

### 6.6 `doctor`

Canonical form:

```bash
kkachi-agent-skills doctor --profile <profile> [--project <path>] [--json]
```

Purpose:

- verify source pack integrity;
- verify target profile install state;
- verify manifest/checksum consistency;
- report KAH availability/version/capabilities and optional project bootstrap/doctor status;
- clearly say when KAB is required for full execution-runtime work.

Required checks:

| Check | Healthy condition |
|---|---|
| source repo | repo path readable; expected `skills/` layout present |
| source checksums | manifest or computed source checksums are internally consistent |
| target profile | profile root exists and is not ambiguous |
| installed files | manifest entries exist at target paths and match SHA-256 |
| unknown profile files | reported as warnings or conflicts, not silently adopted |
| KAH availability | `kkachi-agent-helper --version` and `capabilities --json` run when available |
| KAH install boundary | capabilities report or local evidence preserves `install_command=false` when checked |
| optional project | `kkachi-agent-helper project status/doctor --json` pass for `--project <path>` when requested |
| KAB boundary | reports `kab_required_for_execution_runtime` when user intent needs backend/session/code-change authority |

Minimum JSON shape:

```json
{
  "ok": true,
  "command": "doctor",
  "source_repo": {"path": "/repo", "state": "ok"},
  "target_profile": {"name": "hwangchung", "root": "/profile", "state": "ok"},
  "manifest": {"path": "/profile/.kas/skill-pack-manifest.json", "state": "ok"},
  "installed_packs": [
    {"pack_id": "software-development/kas-codex-roadmap-development", "state": "ok", "files_checked": 12}
  ],
  "kah": {
    "available": true,
    "version": "kkachi-agent-helper 0.1.4",
    "install_command": false,
    "project_status": "ok"
  },
  "kab": {
    "required_for_minimum_cli": false,
    "required_for_execution_runtime": true,
    "message": "KAB is required before KAS-governed backend/code-change runs, not for profile-scoped list/install/doctor."
  },
  "kab_adoption_stage": {
    "applicable": true,
    "marker_state": "marker_present",
    "numeric": 1,
    "canonical": "stage1_direct_codex_app_server_baseline",
    "marker_path": "skills/<project>/<project>-kas/references/kab-adoption-stage.md"
  },
  "diagnostics": [],
  "next_action": "Profile install state is healthy. Use full KAS+KAH+KAB path for execution-runtime work."
}
```


### 6.7 Project-specific dry-run and approved install

`docs/sot/project-specific-kas-install-contract.md` is the detailed SOT for the
KASPROJ project-specific install layout. KASROLE-002 requires explicit
`--suite-role` for both the public lifecycle form and the compatibility alias.
KASPROJ-002 implements the dry-run planner and KASPROJ-003 implements
approval-hash-bound install:

```bash
kkachi-agent-skills install-project-kas --profile <profile> --project <project> --suite-role <role> --source-pack <source_pack> --dry-run [--json]
```

The implemented source pack is `kas-default-project-suite`, resolved from repository `skill-pack.yaml` plus current discovered source skills under repo `skills/`. The command also accepts `--repo` and harness-only `--profile-root` when `KAS_ALLOW_PROFILE_ROOT_OVERRIDE=1`. Missing both `--dry-run`/`--approve`, using both together, malformed `--approve`, missing required flags, unsupported write/force/repair/migrate/from-generic flags, and unguarded profile-root overrides fail closed with exit 2 and `ok:false` JSON when `--json` is requested.

Dry-run JSON evidence includes `ok`, `command`, `mode`, `cli_version`, `dry_run`, `no_write`, `project`, `source_pack`, `project_tailoring`, `summary`, `planned_manifest`, `planned_skills`, `changed_paths`, `backup_plan`, `checksums`, `plan_hash`, `approval_request`, `conflicts`, `diagnostics`, and `next_action`. The `plan_hash` binds CLI version, target profile/root, manifest path/state/previous manifest sha, source repo commit/dirty, formal source suite checksum, planned manifest/skills, changed paths, conflicts/diagnostics, backup plan, and no-write evidence. The planner performs no profile writes and does not create profile roots, skill directories, manifest files, `.kas` directories, KAH state, KAB runtime state, auth paths, token files, gateway config, provider config, or model config.

KASPROJ-002 conflict example:

```json
{
  "ok": false,
  "command": "install-project-kas",
  "mode": "project_dry_run",
  "dry_run": true,
  "conflicts": [
    {
      "severity": "error",
      "condition": "generic_installed_skill_name",
      "project": "doksuri-server",
      "installed_skill": "kkachi-plan",
      "target_path": "skills/doksuri-server/kkachi-plan/SKILL.md"
    }
  ],
  "plan_hash": "sha256:<hex>"
}
```

Approved install form:

```bash
kkachi-agent-skills install-project-kas --profile <profile> --project <project> --suite-role <role> --source-pack kas-default-project-suite --approve dry-run:<hash> [--json]
```

Approved install recomputes the current dry-run, compares the approval evidence to the recomputed `plan_hash`, and writes nothing unless the required `--suite-role` is present, the hash matches, and the plan has no conflicts/errors. It preflights all targets and source checksums, rejects missing/unknown roles, unsafe/path-traversing/symlink-escaping/ambiguous/duplicate/local-modified targets, backs up trusted replacements under `.kas/backups/<install_id>/`, atomically writes project skill files, verifies checksums, and writes `.kas/skill-pack-manifest.json` last. Output includes approval evidence, `install_id`, `manifest_path`, backup/recovery evidence, changed path counts/actions, and a next action to run project-suite doctor.

KASPROJ-004 implemented project-suite doctor/repair/migrate forms:

```bash
kkachi-agent-skills doctor --profile <profile> --project <project> --project-suite [--json]
kkachi-agent-skills repair-project-kas --profile <profile> --project <project> --dry-run [--json]
kkachi-agent-skills repair-project-kas --profile <profile> --project <project> --approve dry-run:<hash> [--json]
kkachi-agent-skills migrate-project-kas --profile <profile> --project <project> --from-generic --dry-run [--json]
kkachi-agent-skills migrate-project-kas --profile <profile> --project <project> --from-generic --approve dry-run:<hash> [--json]
```

TOKEN-004 public lifecycle forms wrap the same dry-run planners:

```bash
kkachi-agent-skills install --profile <profile> --project <project> --dry-run [--json]
kkachi-agent-skills install --profile <profile> --project <project> --from-generic --dry-run [--json]
kkachi-agent-skills repair --profile <profile> --project <project> --dry-run [--json]
kkachi-agent-skills uninstall --profile <profile> --project <project> --dry-run [--json]
```

In TOKEN-004, these public project forms were read-only. `install --project`
used the project-suite dry-run planner, `install --from-generic` used the
generic-to-project migration dry-run planner, and `repair` used the project
repair dry-run planner. Public write/apply forms failed closed before TOKEN-005.
The old `install-project-kas`, `repair-project-kas`, and `migrate-project-kas`
commands remain compatibility surfaces for existing tests and scripts.

`uninstall --dry-run` is planner-only in TOKEN-004. It reads the profile
manifest and filesystem, then reports manifest-tracked planned removals,
skipped local-only or unmanifested files, checksums, changed paths,
backup/recovery posture, and the future apply command. It must not remove
files, write profile manifests, create backup files, mutate KAH/KAB/Hermes
runtime state, change auth/token/gateway/provider/model configuration, or
activate profiles. Removal and backup/evidence writing are TOKEN-005 behavior.

TOKEN-005 public lifecycle write forms use `--apply dry-run:sha256:<hash>`.
`install --project --apply`, `install --from-generic --apply`,
`repair --apply`, `update --apply`, and `uninstall --apply` recompute the
current dry-run evidence and fail closed before mutation when the token is
malformed, mismatched, stale, or paired with `--dry-run`. `update --apply`
writes only hash-bound auto-copy candidates and keeps semantic merge, removed
upstream, new upstream, and conflict classifications blocked. `uninstall
--apply` removes only manifest-tracked project-suite files and requires
`--backup-vault-root <abs-path>`; the backup root must be absolute, outside the
target profile, symlink-safe, and writable/verifiable. Backup files, manifest
snapshot, skipped local-only files, checksums, approval evidence, and uninstall
evidence are written under the explicit vault root before any removal, and the
profile manifest is updated last while preserving unrelated `installs` and
`project_suites`.

KASROLE-002 makes role-aware project install explicit on the public lifecycle
surface:

```bash
kkachi-agent-skills install --profile <profile> --project <project> --suite-role <role> --dry-run [--json]
kkachi-agent-skills install --profile <profile> --project <project> --suite-role <role> --apply dry-run:sha256:<hash> [--json]
```

The compatibility alias `install-project-kas` accepts the same
`--suite-role`. Missing, unknown, or unregistered roles fail closed with
`ok:false` and exit 2 before writes; KAS must not infer a role from the profile
name. Dry-run JSON includes `suite_role`, `suite_mode`, `role_registry`,
`selected_skills`, `excluded_skills`, selected/excluded counts, and
`approval_request.hash_includes_role_fields:true`. Human output includes the
role display label and preserves `--suite-role` in the copy/paste apply command.
Apply recomputes the same role-aware dry-run and refuses stale role registry
checksums, source-suite changes, selected/excluded skill changes, or approval
hash mismatches before writing profile files.

When `doctor --project-suite` is absent, `doctor --project <path>` keeps its existing KAH project-path meaning. When `--project-suite` is present, `--project` is interpreted as the project suite id. Repair and migration default `source_pack` to `kas-default-project-suite`; optional explicit `--source-pack` values fail closed when unknown, and the resolved source pack is included in JSON, diagnostics, and plan-hash evidence.

Contractual rules for project commands:

- dry-run planner behavior belongs to KASPROJ-002 and performs no profile writes;
- approved project-specific install belongs to KASPROJ-003 and requires approval evidence matching the recomputed dry-run plan hash;
- doctor/repair/migrate behavior belongs to KASPROJ-004 and uses the severity semantics from the project-specific install SOT;
- KASPROJ-005 refines doctor semantics so project-local semantic tailoring is a warning condition, not a failed source/template checksum comparison;
- repair and migration are dry-run-first, approval-gated, backup-aware, and must not silently convert generic skills into project-tailored skills;
- all project-suite writes must use `target_path` values under `skills/<project>/<project>-<phase-or-skill>/SKILL.md`;
- umbrella-only installs are incomplete/invalid and must not be treated as healthy;
- duplicate generic installed skill names such as `kkachi-plan` are invalid for one-profile multi-project suites;
- no project command may mutate auth, tokens, secrets, gateway, provider/model config, KAH project state, or KAB runtime state;
- KASPROJ-003 is prefix-render-only: it records `semantic_adaptation_claimed:false`, preserves `drift_policy: manual_review_required`, and must not claim semantic-port completion or operational rollout.

Doctor severity names for KASPROJ are `error` and `warning` with these
required conditions: missing project suite, umbrella-only, missing file, and
prefix-render-only or otherwise untrusted checksum mismatch are `error`;
`project_tailoring_checksum_drift` is `warning` when a manifest skill record is
marked `tailoring_mode: profile_local_repo_semantic_tailoring` with
`drift_policy: manual_review_required`; unknown profile skill dir is `warning`
unless it shadows or ambiguates a required project skill; profile/source language
drift is `warning` unless it invalidates required behavior or verification, in
which case it escalates to `error`.


## 7. Manifest/checksum contract

### 7.1 Manifest path

The approved install writes KAS profile metadata to:

```text
~/.hermes/profiles/<profile>/.kas/skill-pack-manifest.json
```

This path is candidate until implementation review, but the contract requires the manifest to stay outside `skills/` so Hermes does not attempt to load KAS metadata as a skill.

### 7.2 Manifest schema

Minimum manifest fields:

```json
{
  "version": "0.1",
  "kind": "kas_profile_skill_manifest",
  "profile": {
    "name": "hwangchung",
    "root": "/Users/name/.hermes/profiles/hwangchung"
  },
  "source_repo": {
    "path": "embedded://github.com/SeventeenthEarth/kkachi-agent-skills@<version-or-checksum>",
    "git_remote": "https://github.com/SeventeenthEarth/kkachi-agent-skills.git",
    "git_commit": "<sha-or-null>",
    "dirty": false
  },
  "installs": [
    {
      "install_id": "kas-install-20260526T000000Z-abcdef",
      "installed_at": "2026-05-26T00:00:00Z",
      "approval_evidence_ref": "dry-run:<hash-or-path>",
      "dry_run_plan_hash": "sha256:<canonical-plan-json-hash>",
      "approved_plan_hash": "sha256:<canonical-plan-json-hash>",
      "pack_id": "software-development/kas-codex-roadmap-development",
      "category": "software-development",
      "name": "kas-codex-roadmap-development",
      "source_path": "skills/software-development/kas-codex-roadmap-development",
      "target_path": "skills/software-development/kas-codex-roadmap-development",
      "kab_adoption_stage": {
        "applicable": true,
        "numeric": 1,
        "canonical": "stage1_direct_codex_app_server_baseline",
        "source": "default_stage1",
        "marker_path": "skills/<project>/<project>-kas/references/kab-adoption-stage.md"
      },
      "checksum_algorithm": "sha256",
      "pack_checksum": "<sha256-hex>",
      "backup": {
        "required": false,
        "path": "/Users/name/.hermes/profiles/hwangchung/.kas/backups/kas-install-20260526T000000Z-abcdef",
        "created": false
      },
      "previous_manifest": {
        "path": "/Users/name/.hermes/profiles/hwangchung/.kas/skill-pack-manifest.json.previous",
        "sha256": "<sha256-or-null>"
      },
      "files": [
        {
          "relative_path": "SKILL.md",
          "action": "create",
          "bytes": 1234,
          "previous_sha256": null,
          "new_sha256": "<sha256-hex>",
          "sha256": "<sha256-hex>",
          "backup_relative_path": null,
          "mode": "0644"
        }
      ]
    }
  ]
}
```

### 7.3 Checksum algorithm

- File checksum: lowercase hex SHA-256 over exact file bytes.
- Pack checksum: SHA-256 over a deterministic manifest payload containing sorted file entries by normalized POSIX relative path, each with path, byte length, mode, and file SHA-256.
- Exclusions: `.git/`, `.DS_Store`, editor swap files, KAH `.kkachi/` state, and generated caches must not be part of pack checksums.
- Symlink handling: default install refuses symlinks. If a source pack contains a symlink, dry-run reports `error` unless a future reviewed developer mode explicitly supports it.
- Path normalization: use POSIX relative paths inside manifests; reject absolute paths, `..`, empty segments, and symlink escapes.

### 7.4 Approval plan hash

Dry-run approval is matched by a canonical plan hash:

```text
dry_run_plan_hash = sha256(canonical_json(plan))
```

The canonical JSON plan must include, with deterministic key ordering and deterministic `changed_paths` ordering by `action` then `path`:

- command mode and CLI version;
- requested pack ids and any category expansion results;
- source repo path, git commit, dirty flag, and source pack checksums;
- target profile name and resolved profile root;
- selected KAB adoption stage, selection source, and marker path when the stage selector is applicable;
- normalized pack source/target paths;
- changed-path entries including action, path, previous checksum when present, new checksum when present, and bytes;
- conflict/error states;
- backup plan for update/delete-risk paths;
- manifest path and previous manifest checksum when present.

Approved install output, manifest entries, and approval records must repeat both `dry_run_plan_hash` and `approved_plan_hash`. An approved install must fail closed if the recomputed plan hash differs from the approved hash. CLIMVP-004 implementation must also review evidence-ref replay/reuse prevention before writes are enabled.

## 8. Changed-path report contract

Every dry-run and approved install must report:

- target profile name and root;
- source repo path and source commit/dirty state;
- pack ids requested and resolved;
- target paths for each pack;
- per-path action: `create`, `update`, `skip`, `conflict`, `error`, `backup`, `manifest_update`;
- checksum before/after when applicable;
- approval requirement and evidence ref;
- KAB adoption stage and marker path when applicable;
- recovery/rollback path for approved writes.

Human output must be concise and operator-oriented. JSON output is the stable harness contract.

## 9. Status vocabulary

Separate status fields must use distinct vocabularies so harness checks do not confuse profile/install state with roadmap state.

| Field | Allowed values | Meaning |
|---|---|---|
| `target_profile.manifest_state` | `manifest_present`, `manifest_missing`, `manifest_unreadable`, `not_applicable` | Whether KAS profile metadata is present/readable. |
| `pack.installed_state` | `not_installed`, `installed_current`, `installed_drifted`, `installed_unknown`, `conflict`, `error` | Relation between a source pack and target profile install. |
| `file.action` / `changed_paths[].action` | `create`, `update`, `skip`, `conflict`, `error`, `backup`, `manifest_update` | Planned or actual per-path action. |
| `doctor.state` | `ok`, `warning`, `failed`, `skipped` | Health-check result. |
| `doctor.project_suite_diagnostic.severity` | `error`, `warning` | KASPROJ project-specific suite diagnostic severity for missing project suite, umbrella-only, missing file, checksum mismatch, unknown profile skill dir, and profile/source language drift. |
| `kab_adoption_stage.canonical` | `stage1_direct_codex_app_server_baseline`, `stage2_kab_codex_first`, `stage3_kab_backend_selected` | KAS/KAH development stage name; Stage 3 is reserved and unsupported by this selector until separately authorized. |
| `kab_adoption_stage.marker_state` | `not_applicable`, `marker_present`, `marker_missing`, `marker_unreadable`, `unsupported_stage` | Installed project-specific KAS marker status. |
| roadmap task status | `Planned`, `In Progress`, `In Review`, `Blocked`, `Completed`, `Deferred` | Roadmap state only; never use it as file/install state. |

## 10. Failure modes

The CLI must fail closed for:

- missing or ambiguous target profile;
- source repo not readable;
- unknown pack id/category;
- path traversal, absolute target paths, or symlink escape;
- missing `--dry-run` and missing `--approve` for install writes;
- approval evidence that does not match the current changed-path plan;
- invalid, conflicting, unsupported, or Stage 3 KAB adoption selector values;
- checksum mismatch after copy;
- manifest parse error or version unsupported;
- existing target files not recorded in a trusted manifest;
- project-specific suite plans that use generic installed skill names instead of `<project>-<phase-or-skill>`;
- umbrella-only project-specific installs that omit required suite skills;
- missing backup/recovery plan for an update;
- attempt to mutate `skills.external_dirs`, symlink default, gateway config, auth/tokens, model/provider config, KAH project state, or KAB runtime config;
- KAH/KAB capability uncertainty when a command tries to claim KAH/KAB-backed status.

Warnings are acceptable for:

- KAH unavailable during `list` or profile-only `doctor`, as long as the report clearly says KAH-backed project checks were skipped;
- KAB unavailable during minimum CLI checks, as long as the report clearly says KAB is required before execution-runtime work;
- unknown profile skill directories outside the KAS manifest, as long as they are not silently adopted or overwritten.

## 11. Harness/testability expectations

Later implementation tasks must provide tests or harness fixtures for:

- list without profile;
- list with profile and manifest present/missing;
- dry-run create/update/skip/conflict/error cases;
- dry-run no-write guarantee;
- approved install rejects changed plan after dry-run;
- approved install writes manifest and recovery instructions;
- checksum mismatch after copy fails closed;
- path traversal and symlink escape rejection;
- doctor healthy/unhealthy profile states;
- KAH unavailable/degraded reporting;
- KAB boundary messaging for minimum vs execution-runtime lanes;
- JSON schema stability for `list`, `install --dry-run`, approved install, and `doctor`;
- `--profile-root` production rejection fixture, per 하후연 Red condition C1;
- evidence-ref replay/reuse prevention design review before CLIMVP-004 write mode, per 하후연 Red condition C2;
- KAB adoption stage selector defaults, explicit numeric/canonical forms, conflicting flags, Stage 3 fail-closed behavior, non-interactive no-prompt behavior, marker generation, doctor marker reporting, and dry-run hash mismatch after stage changes.

## 12. Review questions for CLIMVP-001

Reviewers should focus on these decisions before this candidate SOT becomes active:

1. Is `~/.hermes/profiles/<profile>/.kas/skill-pack-manifest.json` the right manifest location, or should KAS use a different profile metadata path?
2. Is the canonical `install --profile <profile> <pack-id>... --dry-run` command shape acceptable, or should the positional profile form be primary for operator UX?
3. Are the changed-path categories sufficient for harness assertions?
4. Are approval evidence and dry-run matching rules strict enough for profile writes?
5. Does the `doctor` boundary clearly distinguish KAS profile install health from KAH project health and KAB execution-runtime readiness?
6. Are 주유 harness requirements and 사마의 Red-review failure modes adequately represented before implementation?

## 13. Completion rule

`CLIMVP-001` may be marked `Completed` only after:

1. Kkachi-team review by 하후연, 여몽, and 진궁 is resolved;
2. 주유 harness review and 사마의 Red review are resolved;
3. this SOT or its accepted successor is updated with final decisions;
4. `docs/README.md` and `docs/roadmap.md` are updated if needed;
5. `git diff --check` passes;
6. KAH run evidence records the review ids, accepted changes, deferred items, and final commit.
