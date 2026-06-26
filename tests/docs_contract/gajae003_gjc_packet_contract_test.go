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
			"GAJAE-005",
			"GAJAE-006",
		})
	}
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
		"candidate_status: ralplan_ready",
		"plan lock remains pending until KAS/Blue/color plan review acceptance",
		"GAJAE-005 KAT ultragoal evidence pilot",
		"GAJAE-006 watcher/callback closeout",
	})
	requireNotContains(t, "templates/run-artifacts/gjc-ralplan-packet.yaml.tmpl", []string{"GAJAE-004 async callback pilot"})

	requireContainsAll(t, "templates/run-artifacts/gjc-callback-contract-packet.yaml.tmpl", []string{
		"packet_kind: callback_contract",
		"callback_delivered",
		"no-wake-claim",
		"callback delivery is evidence only until KAS plan review",
		"GAJAE-005 KAT ultragoal evidence pilot",
		"GAJAE-006 watcher/callback closeout",
	})
	requireNotContains(t, "templates/run-artifacts/gjc-callback-contract-packet.yaml.tmpl", []string{"evidence_only_until_GAJAE_004_005_006"})

	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"GAJAE-004 async ralplan/callback pilot is allowed only for ralplan_ready and callback evidence.",
		"GAJAE-005 and GAJAE-006 remain deferred unless separately approved.",
	})
	requireContainsAll(t, "docs/sot/gajae-delegated-execution-contract.md", []string{
		"GAJAE-004 source-side pilot",
		"`ralplan_ready` and `callback_delivered` remain candidate/evidence states only",
		"`plan_locked` requires accepted_plan_hash after KAS/Blue/color review",
	})
}
