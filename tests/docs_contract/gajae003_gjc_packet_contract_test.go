package docscontract

import "testing"

func TestGAJAE003GJCPacketTemplatesExistAndPreserveBoundaries(t *testing.T) {
	templates := map[string]string{
		"templates/run-artifacts/gjc-deep-interview-packet.yaml.tmpl":    "deep_interview",
		"templates/run-artifacts/gjc-ralplan-packet.yaml.tmpl":           "ralplan",
		"templates/run-artifacts/gjc-ultragoal-packet.yaml.tmpl":         "ultragoal",
		"templates/run-artifacts/gjc-review-fix-turn-packet.yaml.tmpl":   "review_fix_turn",
		"templates/run-artifacts/gjc-callback-contract-packet.yaml.tmpl": "callback_contract",
	}

	for rel, packetKind := range templates {
		requireContainsAll(t, rel, []string{
			"schema_version:",
			"packet_kind: " + packetKind,
			"task_id:",
			"run_id:",
			"source_command:",
			"korean:",
			"gjc_operational_brief_english:",
			"authority_boundaries:",
			"stop_ask_gates:",
			"plan_lock:",
			"fallback_policy: none_fail_closed",
			"no_gjc_self_approval: true",
			"forbidden_scope:",
			"expected_outputs:",
			"artifact_ref_contract:",
			"completion_boundary:",
			"packet_ref",
			"artifact_refs",
			"KAS/Blue/color/MAR/final",
			"GAJAE-004",
			"GAJAE-006",
		})
	}
}

func TestGAJAE005UltragoalKATEvidenceAndReviewFixBoundaries(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/gjc-ultragoal-packet.yaml.tmpl", []string{
		"candidate_status: implementation_goal_bundle_ready",
		"technical_aliases:",
		"ultragoal_goals_ready",
		"gjc_goal_bundle_ready",
		"legacy_status: ultragoal_ready",
		"reserved_after_executor_loop:",
		"implementation_diff_ready",
		"implementation_candidate_ready",
		"implementation_verified",
		"native_ultragoal_input:",
		"mode: brief_file",
		"{{ gjc_ultragoal_brief }}",
		"native_input_ref",
		"ultragoal create-goals --brief-file <native_input_ref.path> --json",
		"must not pass the packet path as `--packet`",
		"kat_status: kat_evidence_ready",
		"kat_refs",
		"status_ref",
		"summary_ref",
		"raw_log_ref",
		"status_hash",
		"raw_log_hash",
		"extractor_status",
		"command_exit_code",
		"attachment_status",
		"real async GJC ultragoal invocation except the explicitly approved bounded GAJAE-008 native GJC smoke",
		"live KAT invocation unless separately approved",
		"Does not approve implementation, tests, review findings, MAR, color review, waiver, or final completion.",
	})
	requireNotContains(t, "templates/run-artifacts/gjc-ultragoal-packet.yaml.tmpl", []string{
		"GAJAE-005 KAT ultragoal evidence pilot",
	})

	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"native_ultragoal_input.mode",
		"native_ultragoal_input.brief",
		"KAH materializes `native_ultragoal_input.brief` as run-local `native_input_ref` evidence",
		"ultragoal create-goals --brief-file <path> --json",
	})

	requireContainsAll(t, "templates/run-artifacts/gjc-review-fix-turn-packet.yaml.tmpl", []string{
		"candidate_status: review_fix_candidate_ready",
		"kat_status: deferred_until_KAH_review_fix_status_support",
		"support_status: deferred_until_KAH_review_fix_status_support",
		"artifact_refs_required: false",
		"kat_refs",
		"real async GJC ultragoal invocation except the explicitly approved bounded GAJAE-008 native GJC smoke",
		"live KAT invocation unless separately approved",
		"Does not close findings, approve MAR, approve color review, approve waiver, or mark final completion.",
	})
	requireNotContains(t, "templates/run-artifacts/gjc-review-fix-turn-packet.yaml.tmpl", []string{
		"candidate_status: ultragoal_ready",
		"GAJAE-005 KAT ultragoal evidence pilot",
	})

	requireContainsAll(t, "docs/sot/gajae-delegated-execution-contract.md", []string{
		"`review_fix_candidate_ready`",
		"`implementation_goal_bundle_ready` is the primary active KAS operator-facing",
		"`ultragoal_ready`, `kat_evidence_ready`, and `review_fix_candidate_ready` remain",
		"Review/fix-turn KAT attachment is not required in GAJAE-005 until KAH implements",
		"Current helper compatibility may still attach KAT",
		"summary ref is the sibling `<summary>.md` file",
		"malformed, hashless, checksum-mismatched, run-id-mismatched, symlinked",
		"real GJC `ultragoal` invocation, live KAT execution",
	})
}

func TestGAJAE003GJCPacketContractIsRegistered(t *testing.T) {
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"gjc_packet_contract:",
		"templates/run-artifacts/gjc-deep-interview-packet.yaml.tmpl",
		"templates/run-artifacts/gjc-ralplan-packet.yaml.tmpl",
		"templates/run-artifacts/gjc-ultragoal-packet.yaml.tmpl",
		"templates/run-artifacts/gjc-review-fix-turn-packet.yaml.tmpl",
		"templates/run-artifacts/gjc-callback-contract-packet.yaml.tmpl",
		"packet_ref",
		"artifact_refs",
		"none_fail_closed",
		"no_gjc_self_approval",
		"mechanical_validation_only",
	})
	requireContainsAll(t, "docs/sot/gajae-delegated-execution-contract.md", []string{
		"GAJAE-003 packet contract",
		"`packet_ref` is KAS input packet evidence",
		"`artifact_refs` are GJC candidate output evidence",
		"fallback_policy: none_fail_closed",
		"no_gjc_self_approval: true",
		"KAH validates packet and artifact references mechanically only",
	})
}

func TestGAJAE004AsyncRalplanPilotIsAllowedWithoutBroadeningDeferredScope(t *testing.T) {
	requireContainsAll(t, "templates/run-artifacts/gjc-ralplan-packet.yaml.tmpl", []string{
		"GAJAE-004 async ralplan/callback pilot",
		"candidate_status: ralplan_candidate_recorded",
		"technical_aliases:",
		"ralplan_recorded",
		"legacy_status: ralplan_ready",
		"plan lock remains pending until KAS/Blue/color plan review acceptance",
		"GAJAE-005 KAT ultragoal evidence pilot",
		"GAJAE-006 watcher/callback closeout",
	})
	requireNotContains(t, "templates/run-artifacts/gjc-ralplan-packet.yaml.tmpl", []string{"GAJAE-004 async callback pilot"})

	requireContainsAll(t, "templates/run-artifacts/gjc-ralplan-packet.yaml.tmpl", []string{
		"native_ralplan_input:",
		"stage: \"{{ native_ralplan_stage }}\"",
		"stage_n: {{ native_ralplan_stage_n }}",
		"artifact: \"{{ native_ralplan_artifact }}\"",
		"KAH must pass these fields to GJC 0.7.3 native `ralplan --write` flags and must not pass the packet path as `--packet`.",
	})
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"native_ralplan_input.stage",
		"native_ralplan_input.stage_n",
		"native_ralplan_input.artifact",
		"KAH projects `native_ralplan_input` into GJC 0.7.3 `ralplan --write` flags and must fail closed instead of falling back to `--packet`.",
	})

	requireContainsAll(t, "templates/run-artifacts/gjc-callback-contract-packet.yaml.tmpl", []string{
		"packet_kind: callback_contract",
		"callback_delivered",
		"no-wake-claim",
		"GAJAE-004 callback delivery and GAJAE-006",
		"source_status_hash",
		"idempotency_key",
		"callback_result",
		"notification_status",
		"wake_evidence_status",
		"missing_watcher_evidence",
		"metadata_recorded_no_wake_claim",
		"same_thread_wake_default: no-wake-claim",
		"GAJAE-005 KAT/ultragoal evidence as approval authority",
		"GAJAE-006 watcher/callback closeout as final completion or wake readiness without evidence",
	})
	requireNotContains(t, "templates/run-artifacts/gjc-callback-contract-packet.yaml.tmpl", []string{"evidence_only_until_GAJAE_004_005_006"})

	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"GAJAE-004 async ralplan/callback pilot uses `ralplan_candidate_recorded` as the primary recorded-candidate status; `ralplan_recorded` is a technical/readback alias, while legacy `ralplan_ready`/callback evidence is candidate/current-compatibility state only and is not consensus, plan acceptance, implementation approval, or final readiness.",
		"GAJAE-005 ultragoal goal-bundle evidence uses `implementation_goal_bundle_ready` as the primary operator-facing status; `ultragoal_goals_ready`/`gjc_goal_bundle_ready` are technical aliases, legacy `ultragoal_ready` is current-helper compatibility only, and GAJAE-006 watcher/callback closeout remains factual source-side evidence unless separately approved for final completion, live runtime, or same-thread wake readiness.",
	})
	requireContainsAll(t, "docs/sot/gajae-delegated-execution-contract.md", []string{
		"GAJAE-004 source-side pilot",
		"`ralplan_candidate_recorded` is the primary active KAS status for a recorded",
		"Legacy `ralplan_ready` and `callback_delivered` remain candidate/evidence states only",
		"`plan_locked` requires accepted_plan_hash after KAS/Blue/color review",
	})
}
