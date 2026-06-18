# KAH Install and Project Bootstrap Reference

Use this reference when `kkachi-agent-helper` is missing, out of date, or when KAS must explain how KAH is applied to a target project. This file is deliberately KAH/KAS-boundary focused: KAH owns deterministic helper state, while KAS/KHS owns Hermes skill behavior and phase policy.

## Authority boundary

```text
KAS / KHS (`kkachi-agent-skills`, historically `kkachi-hermes-skills`)
  - Hermes skill packs, phase guidance, task contracts, backend-selection guidance,
    prompt profiles, evidence expectations, and project-specific operational suites.

KAH (kkachi-agent-helper)
  - Deterministic project bootstrap, project-local `.kkachi/` state, docs map,
    schemas, run lifecycle, artifacts, events, locks, gates, diagnostics,
    phase-plan and approval records, and workflow graph validation/proposal/apply.
```

Do not invent a KAH install command. Current KAH capabilities explicitly omit an `install` surface; KAH is installed as a Go binary. KAS CLI installation is handled by `go install github.com/SeventeenthEarth/kkachi-agent-skills@latest`, and KAS profile skill placement remains a separate KAS/Hermes concern.

## Prerequisites

- macOS/Darwin for the currently supported local operator lane.
- Go on `PATH`.
- Network access to GitHub and the Go module proxy for remote `go install`.
- `$(go env GOPATH)/bin` or the selected Go install bin directory on `PATH`.
- A clear target Hermes profile when verifying KAS/KHS profile installation.
- Optional local KAS checkout only when using `kkachi-agent-skills --repo <path>` for development overrides; normal CLI installation uses the embedded source.

## Install KAH from remote module

Preferred automatic install:

```bash
go install github.com/SeventeenthEarth/kkachi-agent-helper@latest
```

If the binary is not found after install, add Go's bin directory to `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Verify the binary and command surface:

```bash
which kkachi-agent-helper
kkachi-agent-helper --version
kkachi-agent-helper capabilities --json
kkachi-agent-helper --help
```

Minimum expected capability groups for KAS/KAH operation:

```text
project: init, status, doctor
run: create, activate, close, abort, list, show
artifact: init, list, validate, write, append, set-status
gate: check, final
event: append
schema: validate, export, migrate
lock: recover
diagnostics: export
phase-plan: init, show, set, validate
approval: request, record, show
graph: init, validate, explain, diff, propose, apply, export
```

If `capabilities --json` lacks a surface a KAS skill intends to use, treat that surface as unavailable. Do not fall back to manually writing `.kkachi/` state unless the owning KAS skill explicitly documents a safe compatibility representation.

## Install or develop KAH from a local checkout

Use this only when 주군 explicitly wants local KAH development or when remote `go install` fails and a local checkout is available.

```bash
git clone https://github.com/SeventeenthEarth/kkachi-agent-helper.git
cd kkachi-agent-helper
go test ./...
go install ./cmd/kkachi-agent-helper
kkachi-agent-helper --version
kkachi-agent-helper capabilities --json
```

If the repository layout changes, inspect the checkout before running `go install`; do not assume a different command path.

## Verify KAS/KHS profile side

KAS is not installed by KAH. Verify the KAS/KHS profile independently:

```bash
hermes --profile <profile> skills list | grep kkachi
```

If the `kkachi-agent-skills` CLI is installed, use its doctor lane for profile-scoped install health. Do not pass `--repo` for normal installed-binary verification; pass `--repo <local-checkout>` only when intentionally testing local development source.

```bash
kkachi-agent-skills doctor \
  --profile <profile> \
  --json
```

Interpretation:

- `kah.available=true` and `kah.capabilities=ok` proves KAH binary availability for that doctor run.
- Missing KAS manifest/checksum evidence is a KAS profile-install issue, not a KAH project-bootstrap issue.
- KAB is not required for the minimum KAS/KAH profile lane, but is required before claiming backend runtime execution, automated review-by-different-tool transport, KAB plan lifecycle, or bridge evidence.

## Apply KAH to a project repository

Run from the target repository root after KAS/KHS has selected project-specific values. Do not bootstrap from a sibling repo and do not reuse sibling `.kkachi/` state.

```bash
kkachi-agent-helper project init \
  --project-name <project-name> \
  --stack <stack-name> \
  --repo-path "$PWD" \
  --commander <Hermes-commander-profile> \
  --redteam <Hermes-redteam-profile> \
  --docs-map-roadmap <path-to-roadmap> \
  --docs-map-spec <path-to-spec> \
  --docs-map-architecture <path-to-architecture> \
  --docs-map-adr-dir <path-to-adr-dir> \
  --docs-map-todo-dir <path-to-todo-dir> \
  --docs-map-spec-dir <path-to-spec-dir> \
  --test-commands "<command1>,<command2>" \
  --backend-policy "stage=<stage1_direct_codex_app_server_baseline|stage2_kab_codex_first|stage3_kab_backend_selected>; allowed=<allowed-backends-or-policy>" \
  --execution-mode <mode> \
  --sot-policy <policy> \
  --json
```

KAS KAB adoption stage is selected at project application time. KAH does not need
to interpret the stage semantically because the same graph/run/artifact mechanics
apply across Stage 1, Stage 2, and Stage 3. Record the stage in KAS guidance and,
when project state is initialized through KAH, in the backend policy string and
generated project overlay/reference. To change stage for an initialized project,
rerun `project init ... --force` only when the persisted KAH project overlay or
backend-policy must change; preserve the old and new stage in the report, and
rerun project doctor when KAH project state was rewritten.

Use `--force` only for intentional non-destructive reconfiguration of bootstrap files. KAH is expected to preserve runs, status, artifacts, events, and gate history during force reconfiguration; still report that `--force` was used.

After project init, verify:

```bash
kkachi-agent-helper project status --json
kkachi-agent-helper project doctor --json
```

Expected project files/directories:

```text
.kkachi/                    # local helper state, usually ignored
.kkachi/config.yaml
.kkachi/project-overlay.yaml
.kkachi/events.jsonl
.kkachi/status.json
.kkachi/schemas/*.schema.json
docs/kkachi-docs-map.yaml   # durable docs map created/managed by project init
```

Recommended repository ignore entries for local helper/tool state:

```gitignore
.kkachi/
.codegraph/
.omx/
.omc/
.external-review-sidecar/
```

Whether `.kkachi-workflow.yaml` and `docs/kkachi-docs-map.yaml` are tracked is a project policy decision; `.kkachi/` is local helper state and should not be committed unless 주군 explicitly directs otherwise.

## Initialize and validate the workflow graph

Check the actual KAH graph surface first:

```bash
kkachi-agent-helper help graph
kkachi-agent-helper capabilities --json
```

Initialize a default graph when the project policy requires project-level workflow graph state:

```bash
kkachi-agent-helper graph init --from-template khs-default --json
```

Validate and inspect:

```bash
kkachi-agent-helper graph validate --json
kkachi-agent-helper graph explain --json
```

For graph changes after initial creation, prefer proposal/apply evidence rather than direct YAML edits:

```bash
kkachi-agent-helper graph validate --file <candidate-graph.yaml> --json
kkachi-agent-helper graph diff --from .kkachi-workflow.yaml --to <candidate-graph.yaml> --semantic --json
kkachi-agent-helper graph propose --candidate-file <candidate-graph.yaml> --reason "<reason>" --json
kkachi-agent-helper graph apply --proposal <proposal-id> --approval <evidence-ref> --json
```

Compatibility note: if the installed KAH rejects a desired gate check type during `graph validate`, preserve the semantics with KAH-supported `artifact.exists`, `markdown.field`, `text.contains`, `text.contains_all`, or `phase.status` checks, then document the compatibility representation. Do not claim support based only on source-code inspection.

## Start KAH-backed run evidence

KAS owns whether a task should start a full Kkachi run. Once selected, KAH records deterministic run evidence:

```bash
kkachi-agent-helper run create --json
kkachi-agent-helper run activate <run_id> --json
kkachi-agent-helper artifact init <run_id> --json
kkachi-agent-helper phase-plan init <run_id> --json
kkachi-agent-helper gate check <run_id> <gate-name> --json
kkachi-agent-helper gate final <run_id> --json
```

Use `kkachi-agent-helper help <command-group>` before relying on subcommand-specific flags. Some help topics are group-level only; for example, `kkachi-agent-helper help graph` is supported even when `kkachi-agent-helper help graph init` is not.

## Split-repo KAS/KAH isolation checklist

For independently operated split repos such as `kan-plugin` and `kan-control`:

- Create or verify a complete project-specific KAS suite under the active Hermes profile, e.g. `skills/kan-plugin/<skill-name>/SKILL.md` or `skills/kan-control/<skill-name>/SKILL.md`.
- Keep skill names globally unique, e.g. `kan-plugin-plan`, `kan-control-plan`.
- Run KAH `project init` separately inside each target repo.
- Do not copy `.kkachi/`, `.kkachi-workflow.yaml`, docs maps, overlays, test commands, CodeGraph evidence, or approval policy from a sibling repo unless 주군 explicitly approves and the same-card evidence explains why it is safe.
- Verify each repo with `project doctor` and, if graph is used, `graph validate`.

## Failure handling

| Failure | Response |
|---|---|
| `go install github.com/SeventeenthEarth/kkachi-agent-helper@latest` fails | Report the exact command and error. Check network, module path, Go version, and whether a local checkout should be used. |
| `kkachi-agent-helper` installed but not found | Check `go env GOPATH`; add `$(go env GOPATH)/bin` to `PATH`; rerun `which kkachi-agent-helper`. |
| `capabilities --json` missing required KAH surfaces | Hold for KAH update or adjust KAS guidance to a validated compatibility path; do not invent manual state writes. |
| `project init` fails on required docs paths | Verify paths relative to repo root, create intended docs directories/files when in scope, or ask 주군 when SOT location is unknown. |
| `project doctor` fails outside a KAH project | Treat as acceptable during global install verification if KAH binary/capabilities pass; for project application, rerun inside target repo root. |
| `graph validate` rejects a gate type | Use supported gate checks only and document the compatibility mapping. |
| KAS profile doctor reports missing manifest | Treat as KAS profile-install health issue, separate from KAH binary/project health. |

## Report evidence template

Report KAH/KAS install or bootstrap status to 주군 in Korean with explicit evidence:

```text
KAH/KAS 확인 보고:
- KAH binary: ✅/❌ <which path>, <version>
- KAH capabilities: ✅/❌ <capability evidence command>
- KAS profile skills: ✅/❌ <profile>, <skills list or doctor result>
- Project bootstrap: ✅/❌ <repo>, <project status/doctor result>
- Workflow graph: ✅/❌/N/A <graph validate/explain result>
- Isolation check: ✅/❌ <sibling repo state not reused>
- Remaining blockers: <none or exact blocker>
```
