package docscontract

import (
	"strings"
	"testing"
)

func TestV02FLOW009PhaseContractDefinesCanonicalAliasesAndFailClosedInputs(t *testing.T) {
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"v02flow009_implementer_phase_dispatch_contract:",
		"contract_version: \"v02flow009.v1\"",
		"canonical_phase_aliases:",
		"kas_graph_operator_ids: [implement, impl]",
		"kah_run_local_phase_id: impl",
		"gjc_prompt_phase_id: impl",
		"kas_graph_operator_ids: [enhance-test, test-enhance]",
		"kah_run_local_phase_id: test-enhance",
		"kas_graph_operator_ids: [ai-slop-cleaner]",
		"artifact_names: [slop-cleanup-log.md, cleanup-plan.md]",
		"kas_graph_operator_ids: [docs, docs-update]",
		"kas_graph_operator_ids: [handle-feedback, handle-feedback-*]",
		"required_phase_brief_fields:",
		"approved_scope",
		"accepted_scope_ref",
		"ralplan_artifact_ref",
		"ralplan_artifact_sha256",
		"blue_plan_lock_or_implementation_approval_ref",
		"selected_ultragoal_session_ref",
		"selected_ultragoal_goal_ref",
		"changed_surface_bounds",
		"preservation_locks",
		"stop_block_conditions",
		"fail_closed_on:",
		"unknown_phase_id_or_alias",
		"ambiguous_alias",
		"unsupported_dispatch_shape",
		"missing_kah_v02flow010_capability_readback",
		"missing_or_stale_ralplan_ref",
		"missing_ralplan_hash",
		"missing_blue_plan_lock_or_implementation_approval_ref",
		"missing_accepted_scope_ref",
		"missing_ultragoal_session_or_goal_ref_at_dispatch_time",
		"unsafe_or_out_of_run_ref",
		"checksum_mismatch_or_unsafe_checksum",
		"stale_or_absent_verification",
		"native_gjc_ai_slop_cleaner_or_remove_ai_slop_request",
		"kab_dispatch_success_as_completion_evidence",
		"blue_or_color_source_patch_fallback_without_recorded_exception",
	})
}

func TestV02FLOW009UltragoalPacketEncodesPhaseBriefAndSlopCleanupContract(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/gjc-ultragoal-packet.yaml.tmpl", []string{
		"phase_brief_contract:",
		"canonical_phase_aliases_ref: registries/phase-contracts.yaml#v02flow009_implementer_phase_dispatch_contract",
		"phase_id_required: true",
		"approved_scope_ref: required",
		"ralplan_artifact_ref: required",
		"ralplan_artifact_sha256: required",
		"blue_plan_lock_or_implementation_approval_ref: required",
		"selected_ultragoal_session_ref: required_at_dispatch_time",
		"selected_ultragoal_goal_ref: required_at_dispatch_time",
		"missing_kah_v02flow010_capability_readback",
		"native_gjc_ai_slop_cleaner_or_remove_ai_slop_request",
		"ai_slop_cleaner_phase_contract:",
		"behavior_lock: required",
		"cleanup_plan: required_before_editing",
		"slop_classification: required_before_editing",
		"deletion_first: true",
		"one_smell_at_a_time: true",
		"verification: required_after_each_changed_surface",
		"closeout_evidence:",
		"slop-cleanup-log.md",
	})
}

func TestV02FLOW009ReviewFixPacketEncodesRemediationFindingBundle(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/gjc-review-fix-turn-packet.yaml.tmpl", []string{
		"remediation_prompt_contract:",
		"finding_bundle:",
		"finding_id: required",
		"source_lane: required",
		"severity_or_blocker_state: required",
		"blue_disposition: required",
		"accepted_scope_ref: required",
		"required_reopened_or_amended_phases: required_when_terminal_phase_touched",
		"audited_reopen_or_amend_evidence_ref: required_when_terminal_phase_touched",
		"refreshed_verification: required",
		"focused_re_review_or_mar_refresh_handoff_ref: required",
	})
}

func TestV02FLOW009ActiveSkillsCarryImplementerLaneBoundary(t *testing.T) {
	for _, rel := range []string{
		"skills/kkachi-implement/SKILL.md",
		"skills/kkachi-enhance-test/SKILL.md",
		"skills/kkachi-optimize/SKILL.md",
		"skills/kkachi-docs-update/SKILL.md",
		"skills/kkachi-handle-feedback/SKILL.md",
		"skills/kkachi-prompt-compose/SKILL.md",
		"skills/kkachi-task-contract/SKILL.md",
		"skills/kkachi-verify/SKILL.md",
	} {
		requireContainsAll(t, rel, []string{
			"V02FLOW-009",
			"selected GJC `ultragoal` executor lane",
			"KAH V02FLOW-010 capability/readback",
			"fail closed",
			"native GJC `ai-slop-cleaner` or `remove-ai-slop`",
		})
	}
}

func TestV02FLOW009RejectsCompletionAndFallbackDrift(t *testing.T) {
	active := strings.Join([]string{
		readRepoFile(t, "registries/phase-contracts.yaml"),
		readRepoFile(t, "templates/run-artifacts/gjc-ultragoal-packet.yaml.tmpl"),
		readRepoFile(t, "templates/run-artifacts/gjc-review-fix-turn-packet.yaml.tmpl"),
		readRepoFile(t, "skills/kkachi-implement/SKILL.md"),
		readRepoFile(t, "skills/kkachi-handle-feedback/SKILL.md"),
	}, "\n")

	for _, required := range []string{
		"KAB dispatch success is dispatch evidence only, not completion evidence",
		"Blue/color source-patch fallback is forbidden unless a recorded exception exists",
		"KAT evidence is mechanical/factual only",
	} {
		if !strings.Contains(active, required) {
			t.Fatalf("active V02FLOW-009 surfaces missing negative boundary %q", required)
		}
	}
}
