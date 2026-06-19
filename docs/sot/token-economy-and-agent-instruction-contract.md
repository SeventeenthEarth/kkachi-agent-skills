# KAS token economy and agent instruction contract

Date: 2026-06-09
Owner: KAS workflow/policy layer
Confirming role: Red, Orange, and Gray token-economy review accepted after blue fixes; color-team install discussion accepted by Red `t_a77d3f90`, Orange `t_8e09b65d`, and Gray `t_bf292f75`; English/lifecycle UX update accepted by Red `t_f76ff099`, Orange `t_c7d1e291`, and Gray `t_2092ddac`; TOKEN-007 through TOKEN-010 extension review accepted by Red `t_8726167f`, Orange `t_c037afad`, and focused Gray re-review `t_148f5ff0` after Gray `t_32211704` requested traceability fixes.
Status: accepted SOT for the token-economy, English KAS product-output, repo-local agent-instruction, and color-team project KAS lifecycle workstream; no implementation, unapproved profile mutation, KAB activation, KAH code change, auth/token/gateway/provider/model config mutation, install/update/repair/uninstall approval, or operational rollout is authorized by this document alone
Authority level: accepted source of truth for KAS-managed token-economy policy, English compact backend output contracts, project KAS lifecycle UX, and `AGENTS.md` / `CLAUDE.md` management expectations
Scope: `kkachi-hermes-skills` KAS docs, skills, templates, registries, prompt guidance, CLI/human-output guidance, project-overlay templates, and KAH mechanical gate requirements. KAB remains a backend bridge/interface, KAH remains deterministic evidence/gate support, and Hermes runtime remains unmodified stock Hermes.
Related docs: `docs/README.md`, `docs/roadmap.md`, `docs/sot/khs-architecture-and-integration.md`, `docs/sot/interface-contract.md`, `docs/sot/phase-orchestration-policy.md`, `docs/sot/kas-cli-contract.md`, `docs/sot/project-specific-kas-install-contract.md`, `docs/sot/project-kas-sync-state.md`, repository `AGENTS.md`
Evidence/source paths: 주군 Discord direction on 2026-06-09; local branch check `git rev-parse --abbrev-ref HEAD` => `main`; observed Hermes/KAS token analysis showed high Hermes-side fixed prompt/tool/context overhead with cache-read dominating billed-looking token counts; public baseline checked from `forrestchang/andrej-karpathy-skills` `CLAUDE.md` for caution/simplicity/surgical-change/goal-driven coding guardrails; Red review `t_e7eb185b` ACCEPT; Orange review `t_c6074150` REQUEST_CHANGES resolved by focused re-review `t_65daf094` ACCEPT; Gray review `t_0166733a` REQUEST_CHANGES resolved by focused re-review `t_87b1e208` ACCEPT; color-team install discussion Red `t_a77d3f90` ACCEPT, Orange `t_8e09b65d` ACCEPT, Gray `t_bf292f75` ACCEPT; English/lifecycle UX review Red `t_f76ff099` ACCEPT, Orange `t_c7d1e291` ACCEPT, Gray `t_2092ddac` ACCEPT; TOKEN-007 through TOKEN-010 extension review Red `t_8726167f` ACCEPT, Orange `t_c037afad` ACCEPT, Gray `t_32211704` REQUEST_CHANGES on acceptance/evidence traceability resolved by focused Gray re-review `t_148f5ff0` ACCEPT.

## 1. Decision summary

KAS must treat token economy as a cross-cutting policy contract, not as a Hermes runtime patch and not as a KAB behavior change.

The selected work plan is:

1. KAS PR 1: token-economy, English product-output, and lifecycle UX SOT/docs-contract.
2. KAS PR 2: English compact artifact-first backend prompt templates and phase guidance.
3. KAS PR 3: English `AGENTS.md` / `CLAUDE.md` template and management workflow.
4. KAS PR 4: project KAS lifecycle UX and read-only planner surface.
5. KAS PR 5: approved lifecycle writes, update apply, and uninstall with vault backup.
6. KAS PR 6: skill slimming and reference split for high-token KAS guidance surfaces.
7. KAS PR 7: project verification profiles and no-agent command runners.
8. KAS PR 8: reversible evidence summaries and compact phase packets.
9. KAS PR 9: compact review bundles and no-agent fan-in watchers.
10. KAS PR 10: change-aware verification matrix.
11. KAH PR 1: mechanical token-economy / English-output / project KAS lifecycle evidence gates.
12. KAH PR 2: mechanical verification-profile, evidence-summary, review-bundle, watcher, and change-aware verification gates.

The same KAS contract must apply to direct Codex app-server lanes and KAB-mediated backend lanes. KAB must stay a connection interface; KAS owns prompt/policy semantics; KAH validates only mechanically checkable evidence.

## 2. Problem statement

When KAS work is driven through Hermes, Hermes-side token usage can look much larger than Codex CLI usage because Hermes carries the commander system prompt, tool schemas, Discord thread context, loaded skills, tool outputs, state checks, review routing, and verification/evidence work. Codex CLI often performs only the implementation turn.

KAS must reduce unnecessary management and console-output token costs without requiring:

- Hermes runtime patches or forks;
- KAB policy logic;
- KAH summarization or subjective warnings;
- shorter-lived Discord sessions as an operator requirement;
- hidden loss of evidence, review gates, or authority boundaries.

## 3. Layer ownership

| Layer | Owns | Must not own |
|---|---|---|
| KAS | task classes, phase applicability, prompt contracts, backend output policy, skill/reference organization, project-specific agent-instruction templates | backend transport/session state, deterministic state storage, auth/token/provider mutation |
| KAH | deterministic artifacts, schemas, events, gates, pass/fail/not_applicable verdicts, evidence existence checks | policy judgment, natural-language summarization, warning-only advisory behavior |
| KAB | backend bridge/session/control interface, retained backend evidence, approval/event transport | KAS policy decisions, task classification, hidden fallback, token-economy judgment |
| Hermes runtime | stock profile/tool/session substrate | fork-only token optimization requirements for this workstream |

## 4. Token-economy and English product-output contract

KAS guidance must prefer compact orchestration with durable artifacts, and KAS-generated product surfaces must be English by default.

### 4.1 Language contract for KAS-generated surfaces

All KAS-generated prompt templates, backend prompts, CLI help text, human CLI output, console summaries, report schemas, and artifact templates must be English by default.

Korean may appear only in external evidence copied from an operator conversation, pre-existing project content, or an explicitly named proper noun that the source project requires. KAS must not generate Korean prose in prompts or console output.

This language rule applies to direct Codex app-server lanes and KAB-mediated backend lanes. Chat reports from a Hermes team member to 주군 are outside this product-output contract.

### 4.2 English compact console output contract

Backend and commander-facing product reports should keep console output compact and English. The default English console schema is:

```text
Status: <pass|fail|blocked|in_progress|not_applicable>
Summary: <1-5 bullets>
Files: <changed/inspected paths only>
Verification: <commands/checks and short result>
Risks/blockers: <only actionable remaining issues>
Detailed artifact: <path or not_applicable reason>
Next action requested: <approval/review/none>
```

Do not paste long plans, full diffs, full logs, full file contents, long reviews, or exhaustive checklist text into the console when an artifact/document path is available. If detailed reasoning is needed, write it to the requested artifact path and return only the compact pointer.

### 4.3 Artifact-first detail policy

Detailed phase output belongs under a KAH/KAS artifact or durable docs path, for example:

```text
.kkachi/runs/<run_id>/artifacts/<phase>/backend-<phase>.md
docs/sot/<contract>.md
docs/discussions/<note>.md
templates/run-artifacts/<artifact>.md.tmpl
```

A backend that cannot write or update the requested artifact must report that blocker instead of dumping the full detail into chat.

### 4.4 Bounded state discovery

KAS phase guidance should avoid repeated broad polling when narrower evidence is sufficient. It should prefer:

- one capability/help check per surface per run, then reuse the evidence path;
- targeted searches instead of repeated full-tree scans;
- focused diffs and changed-path lists instead of full raw logs;
- compact JSON excerpts for gate failures instead of full command output dumps;
- phase capsules/state packets for long Discord continuity rather than forcing physical session splits.

### 4.5 Task class gating

KAS must classify work before selecting phases. Initial task classes are:

| Task class | Intended use | Default phase posture |
|---|---|---|
| `simple_report` | read-only answer or short status | no implementation/review loop unless requested |
| `investigation` | evidence gathering / diagnosis | compact findings artifact; no code mutation |
| `docs_only` | SOT, roadmap, contract, docs updates | docs verification and color review where required |
| `development` | source/template/CLI behavior change | plan, implement, verify, review, final gate |
| `review` | independent review or feedback handling | read-only unless feedback is accepted and authorized |
| `epic_closure` | close/update roadmap or run state | evidence synthesis and gate/readiness checks |

Task class selection must reduce unnecessary phases, not bypass required gates.

## 5. Repo-local agent instruction management

KAS must manage project-local agent instruction files as part of the AI instruction surface.

### 5.1 Managed files

KAS should support both:

- `AGENTS.md`: generic repo-local agent instruction for Codex/OpenAI-compatible and general coding agents;
- `CLAUDE.md`: Claude-compatible repo-local instruction for Claude Code and Claude-like coding agents.

The files may share the same policy core, but wording can be adapted for the target backend. KAS must not assume a `CLAUDE.md` exists in every project; it should create, update, or mark not_applicable according to project policy and approval.

### 5.2 Baseline behavior

The `CLAUDE.md` template must preserve the useful baseline guardrails from `forrestchang/andrej-karpathy-skills`:

- think before coding;
- simplicity first;
- surgical changes;
- goal-driven execution and verification.

KAS adds project-specific Kkachi policy on top:

- KAS/KAH/KAB layer boundaries;
- English compact console output and artifact-first details;
- no speculative features, no hidden fallback, no invented KAH/KAB commands;
- evidence-backed completion and review gates;
- explicit no-auth/no-token/no-provider mutation unless approved.

### 5.3 Managed block policy

KAS should preserve local project text. Template/application support must prefer managed blocks over blind overwrite:

```md
<!-- KAS:MANAGED:BEGIN core-behavior -->
...
<!-- KAS:MANAGED:END core-behavior -->

<!-- PROJECT:LOCAL:BEGIN -->
Project-specific instructions preserved by KAS.
<!-- PROJECT:LOCAL:END -->
```

If an existing file lacks markers, KAS should produce a dry-run merge plan and require approval before rewriting or inserting managed blocks.

POLPR-006 implements the repo-local agent-instruction lifecycle through
`kkachi-agent-skills update agent-instructions`. This lifecycle is distinct from profile-local skill installation: it targets repository-root `AGENTS.md`
and `CLAUDE.md`, reports dry-run hash evidence, updates only the KAS managed
block, reports any preserved `PROJECT:LOCAL` block as the preservation action
`preserve_project_local_block`, and fails closed with
`blocked_unmarked_existing_file` when an existing instruction file lacks
markers. Malformed KAS source templates that lack the managed block are
reported as the `error` outcome with diagnostic
`source_template_missing_managed_block`. Existing repo-local instruction files
with missing, reversed, or duplicate KAS managed markers are non-approvable
with diagnostic `existing_managed_block_malformed` unless they are fully
unmarked, which remains `blocked_unmarked_existing_file`. The lifecycle must
not use profile install, profile manifests, runtime activation, KAH state
writes, KAB session control, or auth/token/provider/model configuration as a
fallback.

### 5.4 Project-specific adaptation

Templates must include project-role substitutions so KAS/KAH/KAB repositories receive different boundary wording:

- KAS repository: KAS owns skill/process/prompt policy and must not become KAH state or KAB runtime.
- KAH repository: KAH owns deterministic evidence/gates and must not become policy/summarization authority.
- KAB repository: KAB owns backend bridge/session/control and must not absorb KAS task-policy decisions.
- Project-specific suites: preserve the concrete project name, selected KAB adoption stage, upstream KAS baseline, and local authority notes.


## 6. Color-team project KAS lifecycle contract

Project-scoped KAS lifecycle support must not be blue-only. When an operator asks `blue` to manage project KAS for a project, `blue` may orchestrate a color-team KAS lifecycle plan so `blue`, `red`, `orange`, and `gray` role profiles share the same project SOT plus role-specific guidance.

This is orchestration authority only. It does not authorize blanket cross-profile writes, profile activation, operational rollout, gateway/model/provider/auth mutation, KAB policy changes, KAH subjective judgment, or Hermes runtime changes.

### 6.1 Required role suite posture

A team KAS lifecycle plan must identify every targeted color role:

- `blue`: command, task-contract, plan, implementation-supervision, verification, and final synthesis guidance;
- `red`: risk, fail-closed, fallback, unsafe-mutation, and approval-boundary review guidance;
- `orange`: operator workflow, product fit, approval clarity, lifecycle UX, and burden review guidance;
- `gray`: SOT, audit, evidence, manifest, doctor, backup, and traceability review guidance.

Durable SOT text must use color-role labels such as `blue`, `red`, `orange`, and `gray`. Member/profile names are evidence handles only, for example Kanban assignees, card IDs, or command output.

### 6.2 Public operator-facing lifecycle verbs

The public operator-facing project KAS lifecycle verbs are:

- `install`: create a project-scoped KAS suite for the selected profile or color team.
- `update`: compare the installed suite with the current upstream source and prepare safe updates.
- `doctor`: inspect installed state without writing.
- `repair`: restore missing or damaged KAS-managed project-suite files when doctor reports a repairable condition.
- `uninstall`: remove KAS-managed project-suite files with manifest-bound backup and evidence.

Operators should not need to choose between `sync`, `migrate`, and approved write modes for normal use.

### 6.3 Advanced/internal lifecycle capabilities

`sync` is the read-only planner/classifier behind `update --dry-run`. It may remain as an advanced diagnostic command, but normal help and operator guidance must present it as `update --dry-run`.

`migrate` is a one-time compatibility path for clean KAS-managed generic skills moving into a project-specific suite. It should be exposed as `install --from-generic` or an advanced `migrate-from-generic` command, not as a routine lifecycle verb.

Approved writes are not a separate operator concept. The operator flow is always:

1. run a dry-run command;
2. inspect the exact plan;
3. apply that exact plan with the emitted hash-bound apply command.

### 6.4 Dry-run and hash-bound apply gate

Cross-profile, multi-role, or write-capable project KAS lifecycle commands require a dry-run before any mutation. The dry-run evidence must cover all target profiles and must include:

- project id;
- lifecycle verb;
- target color role;
- target Hermes profile;
- source pack;
- installed skill ids;
- target paths;
- source commit or checksum;
- planned state: `create`, `update`, `remove`, `no_change`, `conflict`, `blocked`, or `error`;
- changed-path set;
- backup/recovery posture where applicable;
- doctor command expected after install, update, repair, or uninstall.

All write-capable lifecycle commands must use the same operator-facing apply pattern:

```text
<command> --dry-run
<command> --apply dry-run:sha256:<hash>
```

`--apply` is the preferred operator-facing spelling. Existing `--approve dry-run:sha256:<hash>` behavior may remain as a compatibility alias or internal implementation detail, but user documentation should not require operators to understand a separate `approved sync/write` concept.

Mutation is allowed only after explicit operator approval for the exact dry-run evidence reference or plan hash. Approval must bind the full per-profile manifest and changed-path set. A `blue` request, candidate SOT, or review card is not approval evidence by itself. Partial approval does not authorize partial writes unless the approved scope explicitly lists the subset and its changed-path set.

### 6.5 Fail-closed atomicity and incomplete profiles

If orchestration cannot complete all targeted profiles, the orchestrator must report which roles/profiles succeeded and which remain incomplete. No profile may receive additional writes after a sibling profile's write has failed until the operator approves a remediation plan. A role profile that is unavailable, lacks approval, has manifest drift, or fails doctor verification must be reported as blocked or degraded; KAS must not silently substitute another role's skill, a generic project suite, a symlinked/shared external directory, or `blue` guidance.

### 6.6 Per-profile content integrity

Each color profile's project-scoped skill content must be generated from the upstream KAS source pack using prefix-render-only tailoring equivalent to the approved project-specific KAS install contract. `blue` must not modify `red`, `orange`, or `gray` skill content to add, remove, or alter role-specific review guidance beyond what the source pack rendering produces. Role-specific customization outside the source pack requires separate approval and evidence.

### 6.7 Doctor and manifest verification gate

After approved writes, each affected profile must run the relevant KAS project-suite doctor check independently. The final lifecycle report must show per-role doctor status, manifest/checksum status, backup/recovery path where applicable, and the next safe action for each role. Any profile with conflict, drift, checksum mismatch, unknown shadowing skill, missing managed marker, missing manifest, or missing doctor evidence remains not ready.

KAH may later verify deterministic fields such as dry-run evidence path, approval reference, per-profile manifest path, role labels, target paths, checksum fields, backup path, lifecycle verb, and doctor verdicts. KAH must not judge review-guidance prose quality or decide whether a role's policy is semantically good.

### 6.8 Uninstall and backup policy

`uninstall` must remove only manifest-tracked KAS-managed project-suite artifacts by default. It must not remove local-only files, unmanifested project instructions, credentials, profile configuration, gateway/model/provider/auth settings, or runtime state.

Before removal, uninstall must write a backup package and evidence report to the approved long-lived Obsidian vault backup area, not only inside the active Hermes profile.

The uninstall dry-run must report:

- target profile and role;
- project id;
- manifest path;
- managed files planned for removal;
- files skipped because they are local-only or unmanifested;
- backup destination;
- checksum/manifest evidence;
- exact apply command.

### 6.9 Color-team lifecycle PR boundary

This SOT update authorizes only consensus wording and docs-contract protection. CLI support for team dry-run/install/update/doctor/repair/uninstall must be implemented in later bounded KAS PRs with read-only dry-run tests, approval-hash-bound write tests, per-profile manifest tests, forbidden-mutation tests, doctor-status tests, uninstall backup tests, and English human-output examples. The CLI PRs must remain separate from profile activation, gateway restart, model/provider/auth changes, KAB activation, and KAH subjective review logic.

## 7. KAS PR boundaries

### 7.1 KAS PR 1: SOT and docs-contract

Acceptance criteria:

- This SOT or accepted successor is present and discoverable from docs indexes.
- The ten KAS PRs plus two dependent KAH PR work plan is recorded without claiming implementation.
- The contract states no Hermes fork, no KAB policy change, and KAH mechanical-only gates.
- A docs-contract test or equivalent verification checks critical terms for token-economy, English KAS product output, compact console, artifact-first detail, project KAS lifecycle UX, `AGENTS.md`, `CLAUDE.md`, and layer boundaries.

### 7.2 KAS PR 2: English compact backend prompts and phase guidance

Acceptance criteria:

- Backend prompt/profile templates include the English compact console schema.
- Phase skills reference artifact-first detail policy and phase artifact paths.
- Task class gating prevents unnecessary phase loops for simple reports/investigations/docs-only work.
- Verification proves active guidance no longer requires broad raw-output dumping or non-English product output for normal backend reports.

### 7.3 KAS PR 3: English agent instruction templates and management workflow

Acceptance criteria:

- `templates/agent-instructions/AGENTS.md.tmpl` and `templates/agent-instructions/CLAUDE.md.tmpl` or accepted equivalent exist.
- Templates include English managed block content, managed block markers, and project-local preservation guidance.
- CLI or documented dry-run workflow reports create/update/no-change/not_applicable posture without blind overwrite.
- CLI dry-run workflow reports `create`, `update_managed_block`, `no_change`, `not_applicable`, `blocked_unmarked_existing_file`, and `error` outcomes without blind overwrite.
- CLI dry-run workflow reports `preserve_project_local_block` as a preservation action for existing `PROJECT:LOCAL` blocks, not as a top-level file-plan outcome.
- Malformed KAS source templates missing managed markers report `source_template_missing_managed_block` and remain non-approvable.
- Existing repo-local instruction files with malformed managed markers report `existing_managed_block_malformed` and remain non-approvable.
- Approved repo-local writes use an exact `--apply dry-run:sha256:<hash>` token and fail closed on stale, malformed, mismatched, or blocked dry-run hash evidence.
- Project-specific adaptation fields preserve KAS/KAH/KAB boundaries and English compact output rules.

### 7.4 KAS PR 4: project KAS lifecycle UX and read-only planner surface

Acceptance criteria:

- Operator-facing help presents `install`, `update`, `doctor`, `repair`, and `uninstall` as the normal lifecycle verbs.
- `update --dry-run` exposes sync classification in English without requiring routine operators to choose a separate `sync` command.
- Generic-to-project migration is presented as `install --from-generic` or advanced `migrate-from-generic`, not a routine lifecycle verb.
- Read-only dry-run output reports all target roles/profiles, source packs, skill ids, target paths, checksums, planned states, changed paths, backup/recovery posture, and doctor commands.
- Tests prove the planner introduces no profile writes, auth/token/gateway/provider/model mutation, KAB policy mutation, KAH subjective-judgment mutation, Hermes runtime mutation, or profile activation.

### 7.5 KAS PR 5: approved lifecycle writes, update apply, and uninstall with vault backup

Acceptance criteria:

- Write-capable lifecycle commands use `--apply dry-run:sha256:<hash>` as the preferred operator-facing spelling and may retain `--approve dry-run:sha256:<hash>` only as a compatibility alias/internal detail.
- Approved writes are bound to an exact dry-run evidence reference or plan hash and fail closed on mismatched manifests, partial approval, sibling-profile write failure, or unapproved content drift.
- Each affected profile receives a per-profile manifest and independent project-suite doctor verification.
- Uninstall removes only manifest-tracked KAS-managed project-suite artifacts by default and writes backup/evidence to the approved long-lived Obsidian vault backup area. Apply requires explicit `--backup-vault-root <abs-path>` and fails closed when that root is missing, relative, inside the target profile, symlink-unsafe, unwritable, or unverifiable.
- English human output distinguishes `create`, `update`, `remove`, `no_change`, `conflict`, `error`, `blocked`, and `degraded` per role/profile.
- Tests prove no auth, token, gateway, provider, model, KAB policy, KAH subjective-judgment, Hermes runtime, profile-activation, local-only file removal, or unmanifested instruction removal is introduced.

### 7.6 KAS PR 6: skill slimming and reference split

Acceptance criteria:

- High-frequency `SKILL.md` files keep trigger/decision/checklist content compact.
- Long procedures, examples, and rare troubleshooting move to `references/` files.
- Guidance still remains discoverable through linked files and related skills.
- Verification compares before/after high-frequency guidance size or at least records changed skill surfaces and expected token-impact rationale.

### 7.7 KAS PR 7: project verification profiles and no-agent command runners

Acceptance criteria:

- KAS defines a project verification profile contract that avoids a global `make test` assumption and lets each project declare aggregate, prepare, unit, integration, e2e, docs, or task-specific verification commands. A profile is policy selected by KAS from project SOT, task contract, or approved run guidance; it is not inferred by KAH and it is not a Hermes runtime feature.
- Each selected gate record must preserve the selected profile/gate id as `selected_profile_id` plus `selected_gate_id`, the gate kind, command, timeout, applicability, and `not_applicable` reason when a gate is out of scope.
- Applicability vocabulary is `required`, `conditional`, `optional`, or `not_applicable`. Result vocabulary is exactly `pass`, `fail`, or `not_applicable` for mechanical verification evidence; warning-only states are not introduced.
- No-agent command runners are preferred for mechanical test/build/check execution. The runner executes the selected command directly, without asking a coding agent or review backend to run the command, and must preserve full stdout/stderr as an artifact while returning only a compact console/model-visible summary. This is the full-log artifact preservation and compact console/model-visible summary contract.
- Every no-agent runner result must include command, timeout, applicability, exit code, duration, log path, log checksum, bounded failure excerpt, deterministic failure extractor posture, and status. When the result is `not_applicable`, the record must keep selected ids plus the reason and must not fake exit code, duration, log path, or checksum evidence.
- Success output includes command, exit code, duration, log path, log checksum, and status. Failure output includes command, exit code, duration, bounded failure excerpt, likely failing target when deterministically extractable, full log path, log checksum, extractor posture, and status.
- The bounded failure excerpt is a compact excerpt for console/model visibility only. The full log artifact remains authoritative and must be recoverable by log path plus checksum.
- Deterministic failure extraction covers common Go, Python/pytest, JavaScript/Vitest/Jest, Playwright, and generic traceback/error patterns before LLM review. Extractor posture must say which deterministic extractor ran, whether it found a likely failing target, and whether extraction was `extracted`, `not_found`, or `not_applicable`; KAH later validates those fields mechanically and must not summarize log meaning.
- TOKEN-007 does not define the TOKEN-010 changed-path skip matrix. It may record selected gates and `not_applicable` reasons, but changed-path classification, skipped-gate policy, and final aggregate preservation rules remain TOKEN-010 scope.
- KAS owns policy/profile selection. KAH may later validate only mechanical evidence fields, artifact existence, checksums, status vocabulary, and required reasons. KAB/Hermes runtime behavior, profile/auth/token/provider/gateway/model configuration, headroom/proxy dependencies, and KAH gate implementation are out of scope.
- Verification proves the runner contract does not require Hermes runtime patches, KAB policy changes, auth/token/provider/gateway/model mutation, profile mutation, a Hermes proxy/headroom dependency, or KAH subjective summarization.

### 7.8 KAS PR 8: reversible evidence summaries and compact phase packets

Acceptance criteria:

- KAS defines compact phase packets or equivalent run summaries for each phase, including run id, task id, phase, status, changed paths, verification summaries, blocker list, artifact paths, checksums where applicable, and next phase/action.
- Full detail remains recoverable by artifact path and checksum; KAS must not discard evidence to save tokens.
- The contract names compression-forbidden fields, including acceptance criteria, explicit operator approvals/denials, forbidden scope, auth/token/provider/gateway/model boundaries, exact failing assertions and command failures, blocking review findings, and KAH final-gate failures.
- The contract names compression-safe fields, including successful repetitive logs, dependency install/build noise, broad unchanged inventories, repeated status readbacks, non-critical progress chatter, large passing command output with log path/checksum preserved, and already-preserved artifact bodies.
- KAS guidance uses artifact-first retrieval: read compact packets first, then targeted artifact ranges only when the decision requires them.
- Verification proves the summary/packet contract remains reversible and does not introduce a `headroom` dependency, Hermes proxy, runtime context-pruning requirement, KAB compression-policy decision, KAH subjective summarization, or auth/token/provider/gateway/model mutation.

#### 7.8.1 Compact phase packet schema

A compact phase packet is a per-phase reversible index, not a replacement for the source artifact. The canonical template is `templates/run-artifacts/phase-packet.yaml.tmpl`. Each packet must preserve these fields:

```yaml
packet_version: "token008.v1"
summary_id: "phase-summary:<run_id>:<phase_id>:<source_checksum_prefix>"
run_id: "<run_id>"
task_id: "<task_id>"
phase_id: "<plan|ask|implement|docs_validation|final_verify|...>"
status: "<pass|fail|blocked|in_progress|not_applicable>"
packet_validity: "<current|stale|invalid|superseded>"
source_artifact:
  path: ".kkachi/runs/<run_id>/<artifact>"
  checksum: "sha256:<hash>"
  ranges:
    - "<line or section range when applicable>"
evidence_class: "<acceptance_criteria|approval|forbidden_scope|failure|review_finding|gate_failure|verification_log|status_readback|...>"
compression_policy: "<no_compression|direct_reference_only|summary_with_pointer|summary_safe>"
changed_paths:
  - "<changed or inspected path>"
verification_summary: "<compact command/check result or not_applicable reason>"
blocker_list:
  - "<blocking issue or empty>"
summary: "<compact English summary>"
critical_references:
  - class: "<evidence class>"
    path: "<artifact path>"
    checksum: "sha256:<hash>"
    retrieval_instruction: "Read exact artifact/range before deciding."
retrieval_instructions:
  default: "Read this packet first."
  required_when:
    - "checksum mismatch"
    - "status is fail or blocked"
    - "approval, denial, forbidden scope, or non-goal is needed"
    - "exact failing assertion or command failure is needed"
    - "review or final gate has blocking findings"
invalidation_behavior:
  invalid_if:
    - "source artifact checksum changes"
    - "referenced artifact is missing"
    - "phase status changes after packet generation"
    - "acceptance criteria, forbidden scope, approval, review finding, or gate verdict changes"
  on_invalid: "Do not rely on packet summary; retrieve source artifact and regenerate packet."
next_action: "<next phase/action or none>"
```

Field semantics:

- `summary_id` must bind the run, phase, and source checksum prefix so a stale packet can be detected.
- `source_artifact.path` and `source_artifact.checksum` are mandatory for reversibility. If a source artifact has no checksum yet, the packet status must stay `blocked` or `not_applicable` with a reason; it must not pretend full detail is recoverable.
- `critical_references` must carry direct artifact path, checksum, evidence class, and retrieval instruction for every compression-forbidden item that is not repeated exactly in the packet.
- `compression_policy` records only the KAS-declared evidence handling policy for the packet. KAB does not decide this policy, and KAH may later validate the recorded field mechanically only.

#### 7.8.2 Compact run summary schema

A compact run summary is a run-level index over phase packets and final critical evidence pointers. It is not a substitute for packets or source artifacts. The canonical template is `templates/run-artifacts/run-summary.yaml.tmpl`. It must preserve:

```yaml
summary_version: "token008.v1"
summary_id: "run-summary:<run_id>:<source_checksum_prefix>"
run_id: "<run_id>"
task_id: "<task_id>"
status: "<pass|fail|blocked|in_progress|not_applicable>"
packet_validity: "<current|stale|invalid|superseded>"
phase_packets:
  - phase_id: "<phase>"
    status: "<pass|fail|blocked|in_progress|not_applicable>"
    packet_path: ".kkachi/runs/<run_id>/phase-packets/<phase>.yaml"
    packet_checksum: "sha256:<hash>"
changed_paths:
  - "<changed or inspected path>"
verification_summary: "<compact run-level verification result or not_applicable reason>"
blocker_list:
  - "<blocking issue or empty>"
critical_references:
  - class: "<evidence class>"
    path: "<artifact path>"
    checksum: "sha256:<hash>"
    retrieval_instruction: "Read exact artifact/range before deciding."
retrieval_instructions:
  default: "Read run-summary.yaml first, then the relevant phase packet."
  required_when:
    - "run summary checksum or packet checksum mismatch"
    - "run status or phase status is fail or blocked"
    - "a decision needs compression-forbidden evidence"
invalidation_behavior:
  invalid_if:
    - "any referenced phase packet checksum changes"
    - "any referenced source artifact checksum changes"
    - "referenced packet or artifact is missing"
    - "run status, phase status, acceptance criteria, forbidden scope, approval, review finding, or gate verdict changes"
  on_invalid: "Do not rely on run summary; retrieve packet/source artifacts and regenerate summary."
next_action: "<next phase/action or none>"
```

Hermes and backend prompts may use the run summary as the first model-visible continuity surface, then retrieve the specific phase packet and source artifact/range needed for the current decision. This is a read-summary-first retrieval policy, not a runtime context-pruning, Hermes proxy, or `headroom` requirement.

#### 7.8.3 Evidence classes and compression policy

Compression-forbidden evidence classes must remain exact in the packet or be directly referenced with artifact path, checksum, and retrieval instructions:

- acceptance criteria;
- explicit operator approvals and denials;
- forbidden scope and non-goals;
- auth/token/provider/gateway/model boundaries;
- exact failing assertions and command failures;
- blocking review findings;
- KAH final-gate failures and gate report paths;
- any artifact checksum or evidence path required to prove reversibility.

Compression-safe evidence classes may be summarized only when the full artifact remains preserved and directly retrievable by path and checksum:

- successful repetitive logs;
- dependency install/build noise;
- broad unchanged inventories;
- repeated status readbacks;
- already-preserved artifact bodies;
- non-critical progress chatter;
- large passing command output with log path/checksum preserved.

`summary_safe` means safe to keep out of model-visible context by default. It never means safe to delete, overwrite, or omit the underlying evidence artifact.

#### 7.8.4 Retrieval and invalidation behavior

KAS guidance must prefer this retrieval order for continuity and review:

1. Read `run-summary.yaml` first when it exists and `packet_validity` is `current`.
2. Read the relevant `phase-packet.yaml` next when a phase decision or phase evidence is needed.
3. Retrieve targeted source artifact ranges when checksums mismatch, status is `fail` or `blocked`, compression-forbidden evidence is needed, review/final-gate findings are blocking, or the packet/run summary is stale, invalid, or superseded.

A packet or run summary is invalid if any referenced artifact is missing, any referenced checksum changes, a phase/run status changes after generation, or acceptance criteria, forbidden scope, approval/denial, review finding, or gate verdict changes. Invalid packets and summaries must not be used as decision evidence until the source artifact is retrieved and the packet/summary is regenerated.

#### 7.8.5 Layer boundaries and adjacent TOKEN contracts

KAS owns the packet/run-summary schema, evidence class vocabulary, compression policy wording, and prompt/template guidance. KAH may later validate only mechanical fields: required keys, status vocabulary, artifact existence, checksum shape/match, validity vocabulary, and required references. KAH must not summarize evidence meaning, judge prose quality, decide compression policy, choose verification/profile/skip policy, or implement subjective warnings.

KAB remains backend/session/control evidence only. TOKEN-008 does not change backend selection, KAB activation, KAB policy, prompt-profile routing, provider/gateway/model settings, auth/token handling, or Hermes runtime behavior. Hermes may read compact packets first and then targeted source artifacts when exact evidence is needed, but TOKEN-008 does not require a Hermes runtime fork, proxy, headroom dependency, runtime context-pruning feature, or evidence discard.

TOKEN-009 review bundles and no-agent fan-in watchers, and TOKEN-010 changed-path verification matrices, remain separate contracts. TOKEN-008 may mention them only to preserve boundaries and must not define their full schemas here.

### 7.9 KAS PR 9: compact review bundle and role fan-in watcher

TOKEN-009 defines a compact review bundle for Red/Orange/Gray review
requests and a terminal-only role fan-in watcher. The bundle is a compact
pointer index for Blue synthesis, not a replacement for review artifacts,
Kanban comments, KAH evidence, review cards, or final synthesis.

Full review evidence remains recoverable by artifact path and checksum; KAS must not discard review evidence to save tokens.

#### 7.9.1 Compact review bundle schema

KAS run guidance must write the compact review bundle as
`review-bundle.yaml` when review fan-in is requested. The schema is YAML-ish
and must preserve the fields Blue needs without pasting full raw review prose
into model context by default:

```yaml
review_bundle_version: "token009.v1"
task_id: "<task id>"
run_id: "<KAH run id>"
acceptance_reference:
  path: "<task contract, SOT, roadmap, issue, or review acceptance path>"
  checksum: "sha256:<checksum>"
diff_artifact:
  path: "<diff or patch artifact path>"
  checksum: "sha256:<checksum>"
diff_checksum: "sha256:<diff checksum>"
changed_paths:
  - "<changed path>"
verification_summaries:
  - phase: "<phase or command>"
    status: "<pass|fail|blocked|not_applicable>"
    artifact_path: "<verification artifact path>"
    artifact_checksum: "sha256:<checksum>"
forbidden_scope:
  - "<forbidden scope or non-goal>"
requested_verdict_vocabulary:
  - accepted
  - accepted_with_required_rework
  - rejected
  - blocked
role_verdicts:
  - role: "<Red|Orange|Gray|other requested role>"
    verdict: "<requested verdict vocabulary value>"
    artifact_path: "<role review artifact path>"
    artifact_checksum: "sha256:<checksum>"
finding_dispositions:
  - finding_id: "<stable finding id>"
    disposition: "<accepted|rejected|blocked|deferred_with_owner>"
    artifact_path: "<disposition artifact path>"
    artifact_checksum: "sha256:<checksum>"
artifact_pointers:
  - artifact_path: "<review evidence path>"
    artifact_checksum: "sha256:<checksum>"
blue_synthesis_inputs:
  role_verdicts: "<compact lane verdict table>"
  finding_dispositions: "<compact disposition table>"
  artifact_pointers: "<recoverable evidence pointers>"
retrieval_instructions:
  default: "Read review-bundle.yaml first, then retrieve targeted source review artifacts by path and checksum when exact evidence is needed."
  required_when:
    - "artifact existence cannot be confirmed"
    - "artifact checksum mismatches"
    - "role verdict presence is missing"
    - "finding disposition presence is missing"
    - "a blocking finding or forbidden scope decision is needed"
invalidation_behavior:
  invalid_if:
    - "referenced artifact is missing"
    - "referenced artifact checksum changes"
    - "requested verdict vocabulary changes"
    - "role verdict presence changes"
    - "finding disposition presence changes"
  on_invalid: "Do not rely on review-bundle.yaml; retrieve source artifacts and regenerate the bundle."
```

`task_id`, `run_id`, `acceptance_reference`, `diff_artifact`,
`diff_checksum`, `changed_paths`, `verification_summaries`,
`forbidden_scope`, `requested_verdict_vocabulary`, `role_verdicts`,
`finding_dispositions`, `artifact_pointers`, `blue_synthesis_inputs`,
`retrieval_instructions`, and `invalidation_behavior` are required review
bundle fields.

#### 7.9.2 Role fan-in watcher contract

The role fan-in watcher is terminal-only and mechanical. pending emits no output; terminal emits compact report. The watcher is collection/notification only and does not replace Kanban comments, KAH evidence, review cards, review
artifacts, or Blue synthesis.

allowed scopes are mechanically checkable:

- artifact existence;
- artifact checksum;
- role verdict presence;
- finding disposition presence.

The watcher may cover review fan-in, long-running verification completion,
Codex/KAB completion signals, or blocked-condition probes only when the
condition is mechanically checkable. It must not create warning-only gate
states, subjective KAH review logic, hidden fallback, Discord/channel delivery
completion claims, evidence discard, Hermes proxy/headroom requirements, KAB
policy/runtime mutation, or auth/token/provider/gateway/model mutation. No KAH
subjective review judgment or review-quality decision is authorized.

### 7.10 KAS PR 10: change-aware verification matrix

TOKEN-010 defines the KAS-owned change-aware verification matrix. The matrix
records why a run selected particular verification commands for the actual
changed path set, what scoped verification ran, what gates were skipped, why
they were skipped, which deterministic evidence proves that decision, and how
final aggregate verification was preserved.

KAS owns verification-selection policy. KAH validates recorded deterministic
evidence only. KAH must not decide skip policy or choose tests to skip.

KAS run guidance must write the matrix as `change-verification-matrix.yaml`
when changed-path classification affects verification selection. The canonical
matrix contract is `token010.v1`:

```yaml
matrix_version: "token010.v1"
task_id: "<task id>"
run_id: "<KAH run id>"
policy_owner: "KAS"
verification_selection_policy_owner: "KAS"
kah_validation_role: "mechanical_recorded_evidence_only"
kah_forbidden_decisions:
  - "decide skip policy"
  - "choose tests to skip"
  - "infer skips from file extensions"
changed_path_source:
  source_type: "<git diff|git status|artifact manifest|review artifact|not_applicable>"
  source_ref: "<command/artifact path>"
  source_checksum: "sha256:<hash>"
changed_paths:
  - path: "<path or not_applicable>"
    change_class: "<no-change|docs-only|source-code|template|schema|test-only|artifact-only|review-comment-only>"
    deterministic_evidence_refs:
      - path: "<artifact path>"
        checksum: "sha256:<hash>"
rules:
  - selected_rule_id: "<token010.rule.id>"
    change_class: "<single change class or not_applicable when the rule covers a changed path set>"
    changed_path_set_classes:
      - "<one changed_paths.change_class value covered by this rule>"
    selected_verification_commands:
      - selected_profile_id: "<profile id>"
        selected_gate_id: "<gate id>"
        command: "<command or not_applicable>"
        timeout: "<timeout or not_applicable>"
        applicability: "<required|conditional|optional|not_applicable>"
        status: "<pass|fail|not_applicable>"
        evidence_ref:
          path: "<log/artifact path or empty when not run>"
          checksum: "sha256:<hash or empty when not run>"
    scoped_verification:
      - command: "<scoped command or not_applicable>"
        scope_reason: "<why this scoped check covers the changed path set>"
        status: "<pass|fail|not_applicable>"
        evidence_ref:
          path: "<artifact path>"
          checksum: "sha256:<hash>"
    skipped_gates:
      - selected_profile_id: "<profile id>"
        selected_gate_id: "<gate id>"
        skip_reason: "<explicit deterministic reason>"
        deterministic_evidence_refs:
          - path: "<prior pass, unchanged-path evidence, task contract, or approval artifact>"
            checksum: "sha256:<hash>"
    no_skipped_gates_reason: "<required deterministic reason when skipped_gates is []>"
    final_aggregate_preservation:
      status: "<preserved|required|not_applicable>"
      not_applicable_reason: "<only populated when status is not_applicable; empty otherwise>"
      deterministic_evidence_refs:
        - path: "<aggregate log, task contract, approval, or equivalent artifact>"
          checksum: "sha256:<hash>"
boundary_notes:
  - "KAS owns verification-selection policy."
  - "KAH validates recorded deterministic evidence only."
  - "KAH must not decide skip policy or choose tests to skip."
  - "Development tasks preserve final aggregate verification unless the active task contract or responsible approval explicitly marks it `not_applicable` with deterministic evidence."
```

Required matrix fields are `selected_rule_id`, `changed_paths`,
`selected_verification_commands`, `scoped_verification`, `skipped_gates`,
`skip_reason`, deterministic evidence refs with checksums,
`changed_path_set_classes`, `no_skipped_gates_reason` when
`skipped_gates: []`, and `final_aggregate_preservation.status`.

#### 7.10.1 Change classes and required evidence

The matrix supports exactly these `changed_paths.change_class` values:

- `no-change`: empty or unchanged path evidence, prior pass evidence,
  unchanged-path evidence, explicit skip reason, and final aggregate
  preservation status. This class can avoid redundant reruns only when all
  deterministic evidence refs are recorded.
- `docs-only`: changed docs path set, docs/scoped verification command
  evidence, selected verification rule, skipped gate reasons, and explicit
  final aggregate status. File extension alone must not imply a skip.
- `source-code`: changed source path set, focused unit/integration/e2e or task-specific verification evidence, selected verification command evidence, skipped gate reasons, and final aggregate preservation for development tasks.
- `template`: changed template path set, template consumer or docs-contract verification evidence, selected command evidence, skipped gate reasons, and final aggregate preservation status.
- `schema`: changed schema/registry path set, schema or registry validation evidence, dependent docs-contract evidence, selected command evidence, skipped gate reasons, and final aggregate preservation status.
- `test-only`: changed test path set, focused changed-test command evidence, selected command evidence, skipped gate reasons, and explicit aggregate policy. Test-only changes do not automatically make aggregate verification not applicable.
- `artifact-only`: changed run/review/evidence artifact path set, artifact path/checksum/schema evidence, selected command evidence or not_applicable reason, skipped gate reasons, and a boundary note that artifact-only evidence does not claim product-source completion.
- `review-comment-only`: changed review comment or review disposition artifact evidence, role/finding disposition pointers, selected command evidence or not_applicable reason, skipped gate reasons, and a boundary note that review comments do not replace implementation, KAH evidence, or final verification.

Each `rules.change_class` value must be a single value from that vocabulary, or
`not_applicable` only when the selected rule applies to the changed path set
rather than one class. Rules that cover multiple path classes must record every
covered class in `changed_path_set_classes`; composite strings such as
`docs-only+schema+template+test-only` are invalid.

`skipped_gates: []` is valid only when no selected or aggregate gate was
skipped, and the rule records `no_skipped_gates_reason` with deterministic
evidence wording. When any gate is skipped, each skipped gate entry must carry
its own `skip_reason` and deterministic evidence refs.

#### 7.10.2 Final aggregate preservation rule

Development tasks preserve final aggregate verification unless the active task
contract or responsible approval explicitly marks it `not_applicable` with
deterministic evidence. A matrix may record scoped verification first, but it
must not convert scoped verification into an implicit aggregate skip. A
`not_applicable` final aggregate status requires a task contract or responsible
approval artifact path and checksum.

#### 7.10.3 Layer boundaries

KAS owns the change classes, selected-rule vocabulary, verification-selection
policy, skip-policy wording, and matrix template. KAH may later validate only
mechanical fields: required keys, allowed class/status vocabulary, artifact
existence, checksum shape/match, and recorded approval/evidence references.
KAH must not decide skip policy, choose tests to skip, infer skips from file
extensions, judge prose quality, or create warning-only gates.

KAB remains backend/session/control evidence only. TOKEN-010 does not change
KAB policy, backend selection, prompt-profile routing, provider/gateway/model
settings, auth/token handling, Hermes runtime behavior, context pruning,
proxy/headroom behavior, or KAH gate implementation. Full verification evidence
remains recoverable by artifact path and checksum; evidence discard is
forbidden.

## 8. KAH PR boundaries

### 8.1 KAH PR 1: token-economy, English-output, and lifecycle mechanical gates

KAH must expose only mechanically checkable results: `pass`, `fail`, or `not_applicable`. No warning-only state is introduced by this workstream.

Gate candidates:

- task class recorded or not_applicable with reason;
- English compact console / detailed artifact policy acknowledged in phase plan or prompt artifact;
- detailed artifact path exists when detail is required;
- `AGENTS.md` and/or `CLAUDE.md` exists, or not_applicable reason is recorded;
- required KAS managed block markers or accepted no-marker migration evidence exists;
- final product report contains English compact summary and artifact pointer;
- project KAS lifecycle dry-run, apply/approval reference, per-profile manifest, role labels, target paths, checksum fields, backup paths, lifecycle verbs, and doctor verdicts exist when team lifecycle work is in scope;
- no broad KAB/Hermes runtime/config mutation claim is made without explicit approval evidence.

KAH must not judge prose quality or rewrite instructions. It validates required fields, files, markers, paths, and approval/evidence records only.

### 8.2 KAH PR 2: verification, evidence-summary, review-bundle, watcher, and change-aware mechanical gates

KAH must expose only mechanically checkable results: `pass`, `fail`, or `not_applicable`. No warning-only state is introduced by this workstream.

Gate candidates:

- selected verification profile id, selected gate id, command, timeout, applicability, exit code, duration, log path, log checksum, and status exist when verification-profile evidence is in scope;
- full log artifacts exist for no-agent runner output, and failure summaries contain bounded excerpts plus full-log pointers without requiring KAH to summarize the failure;
- compact phase packet artifacts match the KAS-declared schema and referenced artifact paths/checksums exist;
- compression-forbidden fields are present as required markers or artifact references when the task contract says they are in scope;
- review bundle artifacts contain required role/review fields, requested verdict vocabulary, diff/evidence pointers, forbidden scope, and finding-disposition fields;
- watcher terminal reports are present only for terminal conditions, include compact status and artifact pointers, and do not claim to replace Kanban/KAH evidence;
- changed-path classification evidence records rule id, changed paths, selected/scoped/skipped verification, skip reason, and final aggregate preservation status;
- no broad KAB/Hermes runtime/config mutation claim is made without explicit approval evidence.

KAS owns verification-selection policy for the change-aware matrix. KAH validates
recorded deterministic evidence only. KAH must not decide skip policy or choose
tests to skip.

KAH must not choose verification profiles, decide skip policy, judge review quality, summarize logs, operate watchers as orchestration policy, or activate KAB/Hermes runtime behavior. It validates required fields, files, paths, checksums, statuses, and approval/evidence records only.

## 9. Non-goals and hard boundaries

This workstream does not authorize:

- Hermes runtime fork/patch work;
- KAB policy, fallback, or selection-judgment changes;
- auth, token, gateway, provider, or model configuration mutation;
- unapproved profile skill installation or operational rollout;
- removal of existing project-local agent instructions, local-only files, or unmanifested project-suite files without approval;
- replacing color review, KAH gates, or 주군 approvals with token-saving shortcuts;
- replacing final aggregate verification for development tasks unless the task contract or responsible approval explicitly marks it `not_applicable` with deterministic evidence;
- treating cache-read-heavy token accounting as proof that all work is wasteful.

## 10. Verification expectations

For each PR in this workstream, reports must include:

- changed file list;
- compact summary of the token-economy behavior affected;
- relevant docs/skill/template readback;
- targeted searches for forbidden fallback or overclaim wording;
- `git diff --check`;
- focused docs-contract tests when present;
- repository test gate when the changed surface requires it;
- change-aware verification matrix evidence when changed-path classification affects selected/scoped/skipped verification or final aggregate preservation;
- color review / KAH gate evidence when required by task class.

## 11. Acceptance record

This SOT is accepted. Acceptance evidence includes:

1. responsible review accepted the ten KAS PRs plus two dependent KAH PR structure and SOT wording;
2. docs indexes include this SOT as the active token-economy / agent-instruction authority;
3. docs-contract verification protects the core terms;
4. required color review accepted the docs/spec boundary;
5. later TOKEN tasks remain separately gated and must not claim implementation, rollout, profile mutation, KAB activation, Hermes runtime changes, or auth/token/provider/gateway/model mutation before their own evidence exists.
