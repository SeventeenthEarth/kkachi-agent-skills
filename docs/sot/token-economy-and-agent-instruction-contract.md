# KAS token economy and agent instruction contract

Date: 2026-06-09
Owner: KAS workflow/policy layer
Confirming role: Red, Orange, and Gray token-economy review accepted after blue fixes; color-team install discussion accepted by Red `t_a77d3f90`, Orange `t_8e09b65d`, and Gray `t_bf292f75`; English/lifecycle UX update accepted by Red `t_f76ff099`, Orange `t_c7d1e291`, and Gray `t_2092ddac`.
Status: accepted SOT for the token-economy, English KAS product-output, repo-local agent-instruction, and color-team project KAS lifecycle workstream; no implementation, unapproved profile mutation, KAB activation, KAH code change, auth/token/gateway/provider/model config mutation, install/update/repair/uninstall approval, or operational rollout is authorized by this document alone
Authority level: accepted source of truth for KAS-managed token-economy policy, English compact backend output contracts, project KAS lifecycle UX, and `AGENTS.md` / `CLAUDE.md` management expectations
Scope: `kkachi-hermes-skills` KAS docs, skills, templates, registries, prompt guidance, CLI/human-output guidance, project-overlay templates, and KAH mechanical gate requirements. KAB remains a backend bridge/interface, KAH remains deterministic evidence/gate support, and Hermes runtime remains unmodified stock Hermes.
Related docs: `docs/README.md`, `docs/roadmap.md`, `docs/sot/khs-architecture-and-integration.md`, `docs/sot/interface-contract.md`, `docs/sot/phase-orchestration-policy.md`, `docs/sot/kas-cli-contract.md`, `docs/sot/project-specific-kas-install-contract.md`, `docs/sot/project-kas-sync-state.md`, repository `AGENTS.md`
Evidence/source paths: 주군 Discord direction on 2026-06-09; local branch check `git rev-parse --abbrev-ref HEAD` => `main`; observed Hermes/KAS token analysis showed high Hermes-side fixed prompt/tool/context overhead with cache-read dominating billed-looking token counts; public baseline checked from `forrestchang/andrej-karpathy-skills` `CLAUDE.md` for caution/simplicity/surgical-change/goal-driven coding guardrails; Red review `t_e7eb185b` ACCEPT; Orange review `t_c6074150` REQUEST_CHANGES resolved by focused re-review `t_65daf094` ACCEPT; Gray review `t_0166733a` REQUEST_CHANGES resolved by focused re-review `t_87b1e208` ACCEPT; color-team install discussion Red `t_a77d3f90` ACCEPT, Orange `t_8e09b65d` ACCEPT, Gray `t_bf292f75` ACCEPT; English/lifecycle UX review Red `t_f76ff099` ACCEPT, Orange `t_c7d1e291` ACCEPT, Gray `t_2092ddac` ACCEPT.

## 1. Decision summary

KAS must treat token economy as a cross-cutting policy contract, not as a Hermes runtime patch and not as a KAB behavior change.

The selected work plan is:

1. KAS PR 1: token-economy, English product-output, and lifecycle UX SOT/docs-contract.
2. KAS PR 2: English compact artifact-first backend prompt templates and phase guidance.
3. KAS PR 3: English `AGENTS.md` / `CLAUDE.md` template and management workflow.
4. KAS PR 4: project KAS lifecycle UX and read-only planner surface.
5. KAS PR 5: approved lifecycle writes, update apply, and uninstall with vault backup.
6. KAS PR 6: skill slimming and reference split for high-token KAS guidance surfaces.
7. KAH PR 1: mechanical token-economy / English-output / project KAS lifecycle evidence gates.

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
- The six KAS PRs plus one dependent KAH PR work plan is recorded without claiming implementation.
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

## 8. KAH PR boundary

### 8.1 KAH PR 1: mechanical gates

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

## 9. Non-goals and hard boundaries

This workstream does not authorize:

- Hermes runtime fork/patch work;
- KAB policy, fallback, or selection-judgment changes;
- auth, token, gateway, provider, or model configuration mutation;
- unapproved profile skill installation or operational rollout;
- removal of existing project-local agent instructions, local-only files, or unmanifested project-suite files without approval;
- replacing color review, KAH gates, or 주군 approvals with token-saving shortcuts;
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
- color review / KAH gate evidence when required by task class.

## 11. Acceptance record

This SOT is accepted. Acceptance evidence includes:

1. responsible review accepted the six KAS PRs plus one dependent KAH PR structure and SOT wording;
2. docs indexes include this SOT as the active token-economy / agent-instruction authority;
3. docs-contract verification protects the core terms;
4. required color review accepted the docs/spec boundary;
5. later TOKEN tasks remain separately gated and must not claim implementation, rollout, profile mutation, KAB activation, Hermes runtime changes, or auth/token/provider/gateway/model mutation before their own evidence exists.
