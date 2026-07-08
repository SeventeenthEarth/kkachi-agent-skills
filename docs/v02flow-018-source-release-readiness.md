# V02FLOW-018 v0.2.1 source release-readiness package

Status: source-side accepted with carries / publication approval pending
Owner: Hwangchung / 황충, Kkachi Blue commander
Scope: `kkachi-agent-skills` + paired `kkachi-agent-helper` source readiness only

## Boundary

This package assembles source-side readiness evidence for a future v0.2.1 publication request. It does not authorize push, tag/release publication, install, runtime activation, provider MAR execution, KAB activation, or provider/auth/token/profile/gateway/model mutation. Publication must be separately approved after V02FLOW-018 review/Blue synthesis.

## Train evidence accepted before this package

- V02FLOW-013: executor-loop completion contract correction accepted source-side.
- V02FLOW-014: KAH ultragoal executor-loop driver/fail-closed evidence accepted source-side.
- V02FLOW-015: mutation subphase fixture/e2e proof accepted source-side.
- V02FLOW-016: capability, HOME, deferred-feedback, legacy-MAR rejection, and standing-waiver proof accepted source-side.
- V02FLOW-017: official Red `t_ffcab6ab`, Orange `t_b211a0bf`, Gray `t_56f41a79`, and Blue `t_1009f3ec` train review accepted with findings only and `blocking_findings=[]`.

## Source-default version readiness

- KAS root CLI must read back `kkachi-agent-skills 0.2.1`.
- KAS compatibility entrypoint `cmd/kkachi-hermes-skills` must read back `kkachi-agent-skills 0.2.1`.
- KAH root/helper CLI must read back `kkachi-agent-helper 0.2.1` through `--version` and `version --json`.
- KAH release notes for v0.2.1 remain draft/readiness notes until publication approval.

## Required evidence for review

- Fresh KAS `HOME=/Users/draccoon git diff --check && HOME=/Users/draccoon make test`.
- Fresh KAH `HOME=/Users/draccoon git diff --check && HOME=/Users/draccoon make test`.
- KAH `go run . capabilities --json` with:
  - `gjc_executor_loop_evidence=true`
  - `diagnostics_deferred_feedback=true`
  - `mar_legacy_rejection_diagnostics=true`
  - `mar_provider_adapter_safety=true`
  - `mar_migration_diagnostics=false`
- KAH `diagnostics deferred-feedback --json` readback. If the live ledger has zero entries, report it as zero-entry PASS only and cite source tests/SOT for richer schema coverage.
- KAS/KAH roadmap and SOT rows aligned to V02FLOW-017 accepted and V02FLOW-018 source package status.
- V02FLOW-017 KAH SOT packet-extraction gap handled by regenerating the closeout packet excerpt or explicitly marking the prior blank file as nonblocking.

## MAR and release boundary

Provider MAR was not executed by V02FLOW-018. Until official v0.2.1 Go/KAH MAR readiness is approved, the active closeout path is standing waiver evidence plus waiver-only post-MAR second-color adoption marked N/A. No legacy Python MAR script or shell adapter is restored by this package.

## Next gate

Official Red/Orange/Gray review plus dependent Blue synthesis completed: initial Red `t_50f5cbc7` and Orange `t_9f216fa4` accepted with findings, initial Gray `t_9c924e1f` requested changes for the KAH deferred-feedback command, Blue `t_d08e1d60` held, focused Gray `t_aac14d40` accepted the remediation with findings, and focused Blue `t_6bace7c2` accepted V02FLOW-018 as source-ready with carries. The next gate is a separate 주군 publication-approval ask.
