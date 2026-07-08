package docscontract

import "testing"

func TestV02FLOW016ReleaseEvidenceProofContract(t *testing.T) {
	requireContainsAll(t, "docs/sot/v020-gjc-workflow-train-corrections.md", []string{
		"V02FLOW-016 release-evidence proof",
		"effective helper `capabilities --json` output",
		"`gjc_executor_loop_evidence=true`",
		"`diagnostics_deferred_feedback=true`",
		"`mar_legacy_rejection_diagnostics=true`",
		"`mar_provider_adapter_safety=true`",
		"`mar_migration_diagnostics=false`",
		"KAH GJC evidence must show `HOME=/Users/draccoon`",
		"`toolchain.operator.real_user_home`",
		"`diagnostics deferred-feedback --json`",
		"logical/readback status hash vs byte-level ledger SHA",
		"legacy Python MAR remains OFF/HOLD",
		"V02FLOW-016 does not run provider MAR",
		"waiver-only closeout makes post-MAR second-color adoption N/A",
	})
}

func TestV02FLOW016RoadmapRegistersAcceptedSourceSideStatus(t *testing.T) {
	requireRoadmapTaskStatus(t, "V02FLOW-016", "Completed / source-side accepted")
	requireContainsAll(t, "docs/roadmap.md", []string{
		"effective `capabilities --json` readback",
		"GJC `HOME=/Users/draccoon`",
		"KAH provider HOME source `toolchain.operator.real_user_home`",
		"legacy Python MAR absence/rejected-input evidence",
		"standing MAR waiver wording",
		"waiver-only second-color N/A",
		"V02FLOW-016 item review and V02FLOW-017 train review both accepted with findings only",
	})
}

func TestV02FLOW018SourceReleaseReadinessPackage(t *testing.T) {
	requireRoadmapTaskStatus(t, "V02FLOW-017", "Completed / train accepted with carries")
	requireRoadmapTaskStatus(t, "V02FLOW-018", "Completed / source-side accepted with carries")
	requireContainsAll(t, "docs/v02flow-018-source-release-readiness.md", []string{
		"source release-readiness package",
		"source-side accepted with carries",
		"kkachi-agent-skills 0.2.1",
		"kkachi-agent-helper 0.2.1",
		"gjc_executor_loop_evidence=true",
		"diagnostics_deferred_feedback=true",
		"mar_legacy_rejection_diagnostics=true",
		"mar_provider_adapter_safety=true",
		"mar_migration_diagnostics=false",
		"zero-entry PASS only",
		"Provider MAR was not executed by V02FLOW-018",
		"focused Blue `t_6bace7c2` accepted V02FLOW-018 as source-ready with carries",
		"separate 주군 publication-approval ask",
	})
	requireContainsAll(t, "docs/sot/v020-gjc-workflow-train-corrections.md", []string{
		"V02FLOW-018 source release-readiness package",
		"source-default versions",
		"KAH `capabilities --json`",
		"zero-entry live ledger",
		"V02FLOW-017 packet-local blank KAH SOT excerpt",
		"Provider MAR was not executed",
	})
}

func TestV02FLOW016SkillGuidanceRequiresCapabilityHomeAndWaiverProof(t *testing.T) {
	requireContainsAll(t, "skills/kkachi-final-verify/SKILL.md", []string{
		"V02FLOW-016",
		"`capabilities --json`",
		"`gjc_executor_loop_evidence=true`",
		"`diagnostics_deferred_feedback=true`",
		"`mar_legacy_rejection_diagnostics=true`",
		"`mar_provider_adapter_safety=true`",
		"`mar_migration_diagnostics=false`",
		"HOME=/Users/draccoon",
		"`toolchain.operator.real_user_home`",
		"waiver-only second-color adoption is N/A",
	})
	requireContainsAll(t, "skills/kkachi-multi-agent-review/SKILL.md", []string{
		"V02FLOW-016",
		"pre-v0.2.1 legacy Python MAR is OFF/HOLD",
		"KAH `mar` provider execution must use `toolchain.operator.real_user_home`",
		"do not claim provider MAR execution",
		"waiver-only closeout makes post-MAR second-color adoption N/A",
	})
}
