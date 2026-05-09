---
name: kkachi-install-guide
description: Install and verify Kkachi ecosystem components (KHS, KAH, KAB) on the local system, then explain how to initialize a target project through KAH project init.
version: 0.1.0
---

# Kkachi Install Guide

Use this skill when the master asks to install, set up, or verify the Kkachi development harness (kkachi-hermes-skills, kkachi-agent-helper, kkachi-agent-bridge) on the current system.

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

- KHS (kkachi-hermes-skills) is installed via Hermes native skill system (`hermes skills install`) or equivalent local skill placement.
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
- network access: GitHub, Go module proxy, Bun/TypeScript package resolution when dependencies are missing
- existing installations: check `kkachi-agent-helper` and `kkachi-agent-bridge` binaries

## Installation flow

### Step 1: Verify KHS

KHS is already installed if this skill is running. Confirm with:

```bash
hermes skills list | grep kkachi
```

If KHS skills are not listed, install via:

```bash
hermes skills install SeventeenthEarth/kkachi-hermes-skills/skills/kkachi-install-guide --category kkachi --yes
```

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
  --backend-policy "<allowed-backends>" \
  --execution-mode <mode> \
  --sot-policy <policy> \
  --json
```

Use `--force` only for non-destructive reconfiguration of existing KAH bootstrap files.

## Report format

Report to master in Korean:

```text
까치 설치 상태 보고:
- KHS (스킬): ✅ 설치됨 / ❌ 미설치
- KAH (helper): ✅ v0.x.x / ❌ 미설치 (자동 설치 시도: 성공/실패)
- KAB (bridge+plugins): ✅ 설치됨 / ❌ 미설치 (bridge/opencode/codex 빌드 상태 포함)
- 프로젝트 적용: KAH project init 실행 준비됨 / 추가 정보 필요

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
