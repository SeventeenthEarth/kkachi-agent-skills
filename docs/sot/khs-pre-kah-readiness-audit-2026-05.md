# KHS pre-KAH readiness audit

Date: 2026-05-23
Owner: KHS documentation archive
Confirming role: Kanban `t_2e9d918a` Codex app-server soldier draft; Blue-inspected candidate with same-card Red/Gray/Orange ACCEPT gates before final synthesis
Status: historical/superseded pre-KAH readiness audit; preserved for lineage and overclaim prevention
Authority level: historical pre-KAH status record; superseded for post-KAH KAS MVP planning by `docs/sot/initdoc-post-kah-reset.md` and active SOT updates
Scope: `kkachi-hermes-skills` docs-only readiness synthesis before KAH 0.1.4 completion; no longer a blocker for post-KAH KAS MVP work
Runtime note: this draft was requested with `openai_runtime=codex_app_server`; it is not KAB runtime evidence

Related docs:
- `README.md`
- `docs/README.md`
- `docs/roadmap.md`
- `docs/sot/khs-architecture-and-integration.md`
- `docs/sot/external-feedback-intake-policy.md`
- `docs/sot/workflow-graph-integration.md`
- `docs/sot/minimum-pilot-cli-lane.md`
- `docs/sot/khs-delegation-packet-and-report-contract.md`

Evidence/source paths:
- Current corrective task: Kanban `t_2e9d918a`
- Prior scoped child-policy card: Kanban `t_fa8f17a3`
- Prior referenced card: Kanban `t_2a151490`
- External-feedback child policy: `docs/sot/external-feedback-intake-policy.md`
- Broad architecture and parent status: `docs/sot/khs-architecture-and-integration.md`
- Delegation packet and report target: `docs/sot/khs-delegation-packet-and-report-contract.md`
- Graph readiness boundary: `docs/sot/workflow-graph-integration.md`
- Minimum/pilot CLI lane: `docs/sot/minimum-pilot-cli-lane.md`
- Docs index and status vocabulary: `docs/README.md`
- Roadmap sequencing: `docs/roadmap.md`
- KAH candidate external-feedback support plan: `/Users/draccoon/Workspace/SeventeenthEarth/kkachi/kkachi-agent-helper/.docs/new-support.md` (candidate support plan only; not implemented KAH evidence)
- Hermes reference note: `/Users/draccoon/.hermes/skills/dogfood/kkachi-bridge-qa/references/kkachi-khs-pre-kah-readiness-completeness-audit-2026-05-23.md`

## 1. Decision summary

Post-KAH supersession note: KAH 0.1.4 now advertises graph support and
`workflow_graph_configurable_feedback_intake=true`. This audit remains useful to
prevent overclaims, but its `blocked_by_kah` language must be read as historical
unless the effective KAH binary for a run lacks the required capability. Current
KAS work should use `kah-evidenced, kas-integration-pending` when KAH support is
present but KAS registries/templates/skills/reports have not yet adopted it.

KHS pre-KAH readiness must be reported in three separate lanes:

1. External-feedback child completion.
2. Whole KHS pre-KAH readiness.
3. Operational/runtime support readiness.

The scoped external-feedback child policy can be complete as a candidate policy
record while whole KHS pre-KAH readiness is still incomplete. Operational or
runtime support readiness is a separate claim that requires KAH and/or KAB
command, code, test, capability, audit, and evidence paths. Docs alone cannot
promote any behavior to implemented runtime support.

Use these labels precisely:

| Label | Meaning in this audit |
|---|---|
| `candidate` | Proposed KHS/KAH/KAB behavior or record exists, but it is not current runtime support. |
| `planned` | Future work or sequencing exists, with no operational guarantee. |
| `blocked_by_kah` | KHS can describe the target, but KAH lacks required schema, validation, proposal, audit, diagnostics, gate, or compatibility evidence. |
| `kab_later` | The target requires KAB backend/session/runtime evidence and is outside the current MVP. |
| `not_operational_support` | The record is useful for design, audit, or planning, but it must not be reported as implemented support. |

## 2. Card lineage

| Card | Current classification | Evidence and caveat |
|---|---|---|
| `t_fa8f17a3` | External-feedback child completion only | The Hermes reference says this card completed `docs/sot/external-feedback-intake-policy.md` and the docs index update. That is scoped policy completion, not whole KHS pre-KAH readiness. |
| `t_2a151490` | Prior read-only broad readiness audit | The task handoff records that this audit concluded broad KHS pre-KAH readiness was not complete beyond the external-feedback child policy. This audit preserves that bounded card lineage rather than upgrading it into implementation evidence. |
| `t_2e9d918a` | Corrective readiness audit | This card asks for the broad split between child completion, whole KHS pre-KAH readiness, and operational/runtime support readiness. |

## 3. Broad status matrix

| Area | Current status | Safe pre-KAH work | KAH/KAB-blocked work | Evidence paths |
|---|---|---|---|---|
| External-feedback child policy | `candidate`, scoped complete, `not_operational_support` | Keep child SOT and docs index clear that min=1/max=5 is target policy only. | Operational bounds validation, graph proposal/apply/audit, generated artifacts, and fail-closed runtime behavior are `blocked_by_kah`; automated different-tool review is `kab_later`. | `docs/sot/external-feedback-intake-policy.md`; `docs/sot/khs-architecture-and-integration.md` section 11 |
| Whole KHS architecture | `candidate`, development base | Keep KHS as Delegation Packet + Evidence Loop + proposal-gated self-improvement; maintain boundaries and acceptance matrix. | Any implemented support claim from the architecture SOT alone is `not_operational_support`. | `docs/sot/khs-architecture-and-integration.md` |
| Delegation Packet and report contract | `candidate`, split into child record by this task | Define packet/report fields, status labels, evidence paths, and fail-closed rules. | Emission, validation, KAH artifact persistence, and KAB runtime evidence are `blocked_by_kah` or `kab_later` until implemented. | `docs/sot/khs-delegation-packet-and-report-contract.md`; `docs/sot/khs-architecture-and-integration.md` sections 15-16 |
| Workflow graph integration | KAH-evidenced; KAS integration-pending | Keep `.kkachi-workflow.yaml` rules and KAS/KAH ownership clear; do not create graph files in this docs task. | KAS template registry, guidance, and evidence mapping remain roadmap work; KAH owns applied graph changes, checksum/audit evidence, graph-vs-run conflict handling, and diagnostics. | `docs/sot/workflow-graph-integration.md`; `docs/roadmap.md` graph epic |
| Minimum/pilot CLI lane | Blue-confirmed lane split; implementation pending | Continue spec work for profile-scoped `list`, `install`, `doctor`, `sync`, and `proposal` support without runtime claims. | Actual CLI behavior, manifests, checksums, backup/recovery, and install/sync writes require implementation evidence. | `docs/sot/minimum-pilot-cli-lane.md`; `docs/roadmap.md` cli epic |
| KHS activation and phase guidance | partial seed behavior with stale surfaces | Preserve activation boundaries and mark stale/conflicting guidance before support claims. | Registry/template/skill behavior changes are outside this task and may require KAH-compatible artifact validation. | `README.md`; `registries/phase-contracts.yaml`; `templates/run-artifacts/*.tmpl`; `skills/*/SKILL.md` |
| Operator report clarity | `candidate` | Keep exact report fields, evidence paths, and boundary labels in SOT docs. | Machine emission and KAH/KAB-backed verification are blocked until implemented. | `docs/sot/khs-architecture-and-integration.md` section 16; `docs/sot/khs-delegation-packet-and-report-contract.md` |
| Proposal-gated self-improvement | `planned/candidate` | Keep destination rules: run artifact, project overlay, backend prompt profile, phase reference, script, or shared KHS proposal. | Automated promotion, profile mutation, install sync, and shared KHS mutation require approval/evidence mechanisms. | `docs/sot/khs-architecture-and-integration.md` section 12; `skills/kkachi-improve/SKILL.md` |
| Operational/runtime support | `not_operational_support` from KHS docs alone | Report current docs as readiness inputs only. | Requires KAH/KAB command/code/test/capability/runtime evidence. | KAH/KAB repos and command outputs are intentionally not edited or created by this task. |

## 4. Safe pre-KAH versus blocked separation

### Safe pre-KAH

These are safe as docs/candidate work when explicitly scoped:

- Maintain broad readiness and stale/overclaim audit records.
- Maintain candidate Delegation Packet and operator report contract docs.
- Maintain status-label maps that prevent support overclaims.
- Maintain workflow graph KHS/KAH ownership and capability-check wording.
- Maintain minimum/pilot CLI lane specs without implementing CLI behavior.
- Record self-improvement proposal/evaluation destinations.
- Update `docs/README.md` only to help readers find the new records.

### Blocked by KAH

In this historical audit, these were not claimable before KAH evidence. Post-KAH KAS work must now read them as integration-pending unless a current effective KAH check lacks the needed capability:

- `EXTERNAL_FEEDBACK_INTAKE` schema, validation, explain, proposal, apply, audit, diagnostics, and compatibility evidence.
- Generated and validated `phase-plan.yaml`, `checklist.md`, task-contract, report, and graph-vs-run conflict behavior.
- Deterministic stale-surface rejection or migration across registries/templates/skills.
- `.kkachi-workflow.yaml` graph mutation, checksum, audit event, and source-precedence enforcement.
- Any implemented label for runtime behavior that currently exists only as candidate/planned prose.

### KAB later

These require KAB runtime evidence and are outside current MVP:

- Automated review-by-different-tool.
- Backend session dispatch, wait/read/status, retained stream/event evidence, and final backend execution evidence.
- Multi-session review evidence and capability-based backend review selection.
- Treating KAB `send` success as anything more than dispatch evidence remains forbidden.

## 5. Broad stale and overclaim coverage

The stale/overclaim problem is broader than the external-feedback child policy.
Current coverage should be read as follows:

| Surface family | Coverage now | Remaining risk |
|---|---|---|
| External-feedback stale `1..3` / max-3 wording | Detailed manifest exists in `docs/sot/external-feedback-intake-policy.md`; parent inventory exists in `docs/sot/khs-architecture-and-integration.md`. | Registries, templates, skills, and older SOT files still contain stale markers by design; this task does not edit behavior files. |
| Architecture support overclaims | Parent SOT says it is a candidate development base and not operational support. | Any report that says "KHS pre-KAH ready" without qualifiers is overbroad. |
| Delegation Packet/report support | Candidate child record is created by this task. | No implementation evidence proves emitted packets/reports are generated, validated, or persisted. |
| Graph integration | Graph SOT requires KAH capability checks and forbids direct YAML fallback. | KHS template registry and artifact mapping remain roadmap work; this task does not create `.kkachi-workflow.yaml`. |
| Minimum/pilot CLI | Lane SOT keeps profile install support separate from full runtime. | CLI implementation, manifest/checksum, backup/recovery, and sync/proposal behavior remain future work. |
| KHS/KAH/KAB boundary | README, docs index, and parent SOT define KHS guidance, KAH state, KAB runtime. | Dirty existing edits and candidate docs can still be misread unless reports cite exact status labels. |

## 6. Reporting rule

Reports about this readiness work must use this shape:

```yaml
external_feedback_child_completion:
  status: candidate
  card: t_fa8f17a3
  support_label: not_operational_support
  evidence: docs/sot/external-feedback-intake-policy.md
whole_khs_pre_kah_readiness:
  status: planned
  current_gap: delegation_packet_report_contract_and_broad_stale_overclaim_coverage_needed
  evidence:
    - docs/sot/khs-architecture-and-integration.md
    - docs/sot/khs-pre-kah-readiness-audit-2026-05.md
    - docs/sot/khs-delegation-packet-and-report-contract.md
operational_runtime_support_readiness:
  status: blocked
  support_label: historical_not_operational_support
  kah_status: historical_blocked_by_kah_or_current_unsupported_effective_kah
  kab_status: kab_later
```

Do not collapse these into an unqualified readiness claim unless the report immediately names
which lane is ready and which lane remains blocked.

## 7. Verification for this docs task

This audit is complete only when:

- `docs/sot/khs-pre-kah-readiness-audit-2026-05.md` exists.
- `docs/sot/khs-delegation-packet-and-report-contract.md` exists or is explicitly rejected as unsafe.
- `docs/README.md` links both records for discoverability.
- No KAH or KAB repo is edited.
- No `registries/`, `templates/`, or `skills/` behavior file is edited.
- No `.kkachi-workflow.yaml` is created.
- No git staging, commit, or push occurs.

