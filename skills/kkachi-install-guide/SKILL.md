---
name: kkachi-install-guide
description: Install and verify Kkachi ecosystem components (KAS, KAH, KAB) on the local system, then explain how to initialize a target project through KAH project init.
version: 0.2.0
---

# Kkachi Install Guide

Use this skill when the master asks to install, set up, or verify the Kkachi development harness (kkachi-agent-skills, kkachi-agent-helper, kkachi-agent-bridge) on the current system.

## Skill structure and Progressive Disclosure

Keep this skill lean. `SKILL.md` contains core behavior only. Move detailed install instructions into `references/` and repeated deterministic checks into `scripts/`.

Recommended local structure:

```text
kkachi-install-guide/
  SKILL.md
  references/
    kah-install.md
    kab-install.md
    kab-build-matrix.md
  scripts/
    check-kah.sh
    check-kab.sh
```

## Core operating rule

- KAS (`kkachi-agent-skills`) can be installed as a self-contained Go CLI with `go install github.com/SeventeenthEarth/kkachi-agent-skills@latest`; the binary embeds canonical KAS skills/templates/registries and normally does not require a source checkout. Use `--repo <path>` only for local development overrides.
- KAS skills can also be installed through Hermes native skill placement when the operator explicitly wants profile-local skill files.
- KAH (kkachi-agent-helper @latest) is installed via `go install`. This skill attempts automatic installation.
- KAB (kkachi-agent-bridge) must be installed from a local git clone by building the Go bridge plus plugin/wrapper subprojects. Do not treat `go install` as a complete KAB install because it omits OpenCode TypeScript and Codex Rust artifacts.
- All install/build attempts are reported to the master with clear pass/fail/status.
- If KAH automatic installation or KAB local build fails, provide the exact failing command plus manual recovery instructions.

## Language policy

```text
Direct report to the master: Korean.
Install logs, command output, and technical details: English.
```

## Required inputs

Before starting, identify:

- target system: macOS (Darwin) only for now
- Go version: 1.26+ required for KAH
- tmux version: 3.0+ required for KAB
- Bun required for the KAB OpenCode TypeScript plugin build
- Rust/Cargo required for the KAB Codex wrapper build
- backend CLI availability: `claude`, `glm`, `codex`, `opencode`, `gemini` as needed for selected KAB lanes
- active v0.2 posture for the target project/profile: KAS/KAH/GJC/KAT by default, with KAB selected only by explicit task approval and current capability evidence; legacy Stage markers are historical and must not be reported as active install guidance
- network access: GitHub, Go module proxy, Bun/TypeScript package resolution when dependencies are missing
- existing installations: check `kkachi-agent-helper` and `kkachi-agent-bridge` binaries

## Installation flow

### Step 1: Verify KAS

KAS CLI is installed if `kkachi-agent-skills` is on `PATH`:

```bash
which kkachi-agent-skills
kkachi-agent-skills --version
kkachi-agent-skills list --json
```

If the CLI is not installed, install the self-contained binary via:

```bash
go install github.com/SeventeenthEarth/kkachi-agent-skills@latest
```

The installed binary embeds `skills/`, `templates/`, `registries/`, and `skill-pack.yaml`; do not require `git clone` for normal installation. If developing from a local checkout, pass `--repo <path>` to override the embedded source.

If profile-local Hermes skills are required separately, install the chosen skill path via Hermes native skill placement after confirming the target profile and approval scope.

## KASREL provenance/dependency evidence gate

Apply the shared KASREL-004 evidence gate in `docs/sot/kasrel-hermes-v016-provenance-contract.md` before this skill claims install health, readiness, release compatibility, orchestration safety, review PASS, verification PASS, or final completion for KAS skills. The local claim must directly cite current non-secret KASREL evidence fields as applicable: `provenance_contract_version`, `source_class_evidence`, `dependency_audit`, `skill_dependencies`, `command_surface_dependencies`, `deleted_bundle_reference`, and `deleted_bundle_diagnostics`. Missing, ambiguous, or stale provenance/dependency evidence fails closed; deleted-bundle references remain cleanup/blocking diagnostics, not fallback lookup or substitution authority.

### Step 2: Install KAH

Check if KAH is already installed:

```bash
which kkachi-agent-helper
kkachi-agent-helper --version
```

If not found, attempt automatic installation:

```bash
go install github.com/SeventeenthEarth/kkachi-agent-helper@latest
```

Verify:

```bash
kkachi-agent-helper --version
```

If automatic installation fails:
- Report error output to master
- Provide manual instructions from `references/kah-install.md`
- Ask master before retrying

### Step 3: Install KAB

Check if KAB is already installed:

```bash
which kkachi-agent-bridge
kkachi-agent-bridge help
```

If not found, install from a KAB repository clone:

```bash
git clone https://github.com/SeventeenthEarth/kkachi-agent-bridge.git
cd kkachi-agent-bridge
make build
make build-opencode-plugin
cargo install --path plugins/codex --locked
mkdir -p "$HOME/.local/bin"
cp bin/kkachi-agent-bridge "$HOME/.local/bin/kkachi-agent-bridge"
kkachi-agent-bridge help
kkachi-codex-wrapper --help
```

KAB build surfaces:

- Go bridge binary: `make build` creates `bin/kkachi-agent-bridge`
- OpenCode TypeScript plugin: `make build-opencode-plugin` creates `plugins/opencode/dist/`
- Codex Rust wrapper: `cargo install --path plugins/codex --locked` installs `kkachi-codex-wrapper`

KAB runtime placement rules:

- Keep the KAB checkout available when OpenCode hybrid plugin assist may be used. KAB stages the managed server-plugin from the built checkout's `plugins/opencode` package.
- Ensure `kkachi-codex-wrapper` is on `PATH`, or set KAB `codex_wrapper_command` / `KKACHI_AGENT_BRIDGE_CODEX_WRAPPER_COMMAND` to an absolute wrapper path.

Verify from the KAB checkout:

```bash
kkachi-agent-bridge help
kkachi-codex-wrapper --help
test -d plugins/opencode/dist
```

If KAB build fails:
- Report which build surface failed: bridge, OpenCode plugin, or Codex wrapper
- Display missing prerequisites (Go 1.26+, Bun, Rust/Cargo, tmux 3.0+)
- Provide the exact failing command and error output

### Step 4: Verify integration

Once all components are installed, verify basic integration:

```bash
kkachi-agent-helper project status --json
kkachi-agent-helper project doctor --json
kkachi-agent-bridge help
```

Outside a project directory, KAH may report that no project is initialized. That is acceptable during global install verification.

### Step 5: Explain project application

When the master asks to apply KHS to a project directory, prepare values and run KAH project bootstrap from that project root:

```bash
kkachi-agent-helper project init \
  --project-name <project-name> \
  --stack <stack-name> \
  --repo-path "$PWD" \
  --commander <Hermes-commander> \
  --redteam <Hermes-redteam> \
  --docs-map-roadmap <path> \
  --docs-map-spec <path> \
  --docs-map-architecture <path> \
  --docs-map-adr-dir <path> \
  --docs-map-todo-dir <path> \
  --docs-map-spec-dir <path> \
  --test-commands "<command1>,<command2>" \
  --backend-policy "posture=kas_kah_gjc_kat; kab=explicit_only; allowed=<allowed-backends>" \
  --execution-mode <mode> \
  --sot-policy <policy> \
  --json
```

Use `--force` only for non-destructive reconfiguration of existing KAH bootstrap files.
Changing the active KAS/KAH execution posture is a KAS policy change, not a KAH
graph/state semantic change. Report the current posture, target posture, reason,
and evidence plan first. Use `project init ... --force` only when the persisted
KAH project overlay/backend-policy must be rewritten; otherwise record the
posture in installed/project-specific KAS guidance and the next run's task/phase
evidence. KAH does not need to interpret the posture beyond carrying the
backend-policy text and project overlay/reference state. Missing or ambiguous
posture fails closed to v0.2 KAS/KAH/GJC/KAT baseline behavior; do not claim KAB
Codex execution unless a KAB lane is explicitly selected and evidenced.

After project init, update repository ignores for local runtime/tool state before verification:

```gitignore
.kkachi/
.codegraph/
.omx/
.omc/
.external-review-sidecar/
```

Then verify the project application from the target repo root:

```bash
kkachi-agent-helper graph validate --json
kkachi-agent-helper project doctor --json
```

For split repos, also verify that sibling project `.kkachi/`, `.kkachi-workflow.yaml`, docs maps, overlays, and test policies are not referenced or copied.

## KAS install vs project overlay boundary

### Project-suite role fan-out default

When 주군 asks to install, update, repair, or refresh a KAS/KAH project suite and does not explicitly scope the work to one named role, treat the operation as a turnkey role-suite update:

- apply and verify Blue, Red, Orange, and Gray together;
- include Teal only when the target project/team has a Teal lane and the task has UI/UX scope;
- do not report a Blue/Hwangchung-only install or refresh as complete for the project suite unless 주군 explicitly requested a Blue-only operation;
- run dry-run/plan output first for every included role/profile, then apply only after the whole role set matches intent;
- report the result as a role matrix with per-role profile, wrapper, overlay, doctor/verification result, and any skipped/not-applicable reason.

The per-profile installed layout remains role-local: each profile receives only its own wrapper/overlay and must not expand one overlay beyond that profile's registered role. The turnkey requirement is an orchestration scope rule, not permission to put Red/Orange/Gray authority into the Blue profile.

When 주군 asks to apply or update KAS for a specific project, treat project-specific KAS/KAH as isolated per repo. Do not reuse a sibling repo's KAS installation/overlay/reference, `.kkachi` state, workflow graph, docs map, CodeGraph evidence, or test-command policy when language, layout, roadmap IDs, test commands, or authority boundaries differ. Duplication is acceptable and preferred over cross-repo bleed-through.

For split repositories that are operated independently, isolation is not satisfied by only creating one thin umbrella skill. Create or verify a complete project-specific operational KAS suite that covers the recurring phase classes for that repo (task contract, backend select, phase state, prompt compose, plan, implement, ask/blockers, review, feedback, verify, docs update, final verify, improvement capture, and any repo-specific bootstrap/conformance lane). If only the umbrella exists, report it as an incomplete bootstrap, not as KAS installation complete.

Canonical installed layout for independently operated split repos is project-name category grouping under the active Hermes profile:

```text
~/.hermes/profiles/<profile>/skills/<project-name>/<skill-name>/SKILL.md
```

Examples:

```text
~/.hermes/profiles/hwangchung/skills/kan-plugin/<skill-name>/SKILL.md
~/.hermes/profiles/hwangchung/skills/kan-control/<skill-name>/SKILL.md
```

Keep the frontmatter `name` globally unique even when the directory is project-scoped, for example `kan-control-plan` or `kan-plugin-plan`. This avoids collisions with generic skills such as `kkachi-plan` or `software-development/plan` while still making filesystem management project-local.

When 주군 asks to apply or update KAS for a specific project, distinguish three locations before writing files:

- **Installed operational KAS**: `~/.hermes/profiles/<profile>/skills/...` (or `$HERMES_HOME/skills/...`). This is the first target when the user wants the currently running Hermes/KAS behavior updated for one active profile.
- **KAS source repository**: `kkachi-agent-skills/skills/...` for local development, or the embedded `kkachi-agent-skills` binary source when installed via `go install`. Update the repository source when promoting the installed/profile-local learning back into canonical KAS.
- **Project repository state**: `<project>/.kkachi/`, `<project>/.kkachi-workflow.yaml`, project docs/config. Do not create `<project>/skills/` as a stand-in for KAS unless the user explicitly asks for a project-local Hermes skill package.

The preferred place for the active v0.2 execution posture is the installed
project-specific KAS suite, for example
`~/.hermes/profiles/hwangchung/skills/kan-plugin/kan-plugin-kas/references/kab-adoption-stage.md`
or
`~/.hermes/profiles/hwangchung/skills/kan-control/kan-control-kas/references/kab-adoption-stage.md`.
The umbrella project skill should point to that reference and load it before
selecting planner/implementer lanes. Mirror the stage into KAH project
overlay/backend-policy only when project-local persistence or multi-operator
visibility is needed.

Pitfall: phrases like “apply to the kan-plugin KAS first, promote to KAS later” usually mean “patch the installed Hermes-profile KAS guidance for kan-plugin now, then later promote that change to the KAS repo.” They do **not** by themselves authorize creating a new `skills/` directory inside `kkachi-agent-network-plugin`.

Reference: `references/kas-kab-adoption-stage-boundary.md` captures the Stage
1/2/3 marker pattern, preferred installed-profile location, and KAH boundary.
Reference: `references/kas-kab-adoption-stage-runbook.md` is historical after V01CLEAN. Use it only as provenance when auditing stale Stage markers; do not duplicate it in active phase skills. Point
operators to the marker plus this reference.

## Report format

Report to master in Korean:

```text
까치 설치 상태 보고:
- KHS (스킬): ✅ 설치됨 / ❌ 미설치
- KAH (helper): ✅ v0.x.x / ❌ 미설치 (자동 설치 시도: 성공/실패)
- KAB (bridge+plugins): ✅ 설치됨 / ❌ 미설치 (bridge/opencode/codex 빌드 상태 포함)
- 프로젝트 적용: KAH project init 실행 준비됨 / 추가 정보 필요
- KAS/KAH v0.2 execution posture: GJC/KAT baseline 확인 또는 explicit KAB selection evidence 필요
- Project-suite role matrix: Blue / Red / Orange / Gray 적용·검증 상태, Teal 포함/제외 사유

다음 단계:
- [구체적인 다음 작업]
```

## Error handling

| Error | Action |
|---|---|
| KAH `go install` fails | Show manual build instructions, ask master |
| `kkachi-agent-helper` not on PATH | Suggest `export PATH="$(go env GOPATH)/bin:$PATH"` |
| KAB bridge build fails | Report `make build` output and Go environment |
| KAB OpenCode plugin build fails | Report `make build-opencode-plugin` output and Bun/TypeScript environment |
| KAB Codex wrapper build fails | Report Cargo output and Rust/Cargo environment |
| Network issues | Suggest retry, report specific error |

## Improvement capture

After each install run, record in `improvement-note.md`:
- Which install method worked/failed
- Environment-specific issues
- Missing prerequisites not caught early
- Suggested script/check improvements


## V01CLEAN active-baseline note

Any legacy Stage 1/Stage 2/Stage 3, direct Codex app-server, or KAB `native_codex` wording retained in this file is historical context only unless a later approved task explicitly selects KAB with current capability evidence. The active KAS/KAH v0.2 path is KAS policy + KAH deterministic evidence + approved GJC candidate artifacts, with KAT factual evidence only.
