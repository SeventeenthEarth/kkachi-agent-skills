package docscontract

import "testing"

func TestTWAKE003SOTAndRoadmapDeclarePolicyConsumerScope(t *testing.T) {
	requireContainsAll(t, "docs/sot/thread-wake-return-path-contract.md", []string{
		"TWAKE-003 | KAS | KAS `v0.2.5`",
		"plan-vet and plan-review cards",
		"first color review Red/Orange/Gray fan-out",
		"MAR start/status/wait/callback/watcher paths",
		"post-MAR second-color adoption/review fan-out",
		"bounded blocked-condition probes",
		"effective KAH binary lacks the reviewed return-path evidence capability",
		"normalize the return-path vocabulary as `blocked`, `degraded`, or `no_wake_claim`, each with an operator-readable recovery hint",
	})

	requireContainsAll(t, "docs/roadmap.md", []string{
		"TWAKE-003 | KAS | Blue dispatch return-path policy and phase-skill guidance",
		"effective KAH capability-check wording for `async_dispatch_return_path_evidence=true`",
		"async_dispatch_return_path_final_gate=true",
		"blocked/degraded/no_wake_claim",
	})
}

func TestTWAKE003PhaseSkillsRequireReturnPathForAsyncDispatches(t *testing.T) {
	for _, rel := range []string{
		"skills/kkachi-orchestrate/SKILL.md",
		"skills/kkachi-plan/SKILL.md",
		"skills/kkachi-review/SKILL.md",
		"skills/kkachi-request-feedback/SKILL.md",
		"skills/kkachi-multi-agent-review/SKILL.md",
		"skills/kkachi-final-verify/SKILL.md",
	} {
		requireContainsAll(t, rel, []string{
			"TWAKE-003 return-path policy",
			"effective KAH capability readback",
			"async_dispatch_return_path_evidence=true",
			"async_dispatch_return_path_final_gate=true",
			"blocked/degraded/no_wake_claim",
			"operator-readable recovery hint",
			"terminal-only Blue-action-required output",
			"watcher/notifier output is state-report-only and never review, MAR, waiver, Blue synthesis, or final acceptance authority",
		})
	}
}

func TestTWAKE003GJCPacketsCarryReturnPathRequirementOrNoWakeState(t *testing.T) {
	for _, rel := range []string{
		"templates/run-artifacts/gjc-ralplan-packet.yaml.tmpl",
		"templates/run-artifacts/gjc-ultragoal-packet.yaml.tmpl",
		"templates/run-artifacts/gjc-review-fix-turn-packet.yaml.tmpl",
		"templates/run-artifacts/gjc-callback-contract-packet.yaml.tmpl",
	} {
		requireContainsAll(t, rel, []string{
			"async_return_path:",
			"schema_version: twake.return_path.v1",
			"async_dispatch_return_path_evidence",
			"async_dispatch_return_path_final_gate",
			"acceptable_states: [proven, blocked, degraded, no_wake_claim]",
			"fail_closed_on_missing_watcher_or_origin: true",
			"terminal_only: true",
			"blue_action_required_output: true",
			"no_authority: true",
			"recovery_hint:",
			"recovery_hint_required_for_no_wake: true",
		})
	}
}

func TestTWAKE003PhaseContractsAndFinalVerifyFailClosedOnMissingReturnPath(t *testing.T) {
	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"TWAKE-003 return-path policy",
		"plan_vet, color_review, mar_review, second_color_review, gjc_long_running, blocked_condition_probe",
		"missing return-path evidence blocks clean closeout",
		"blocked/degraded/no_wake_claim requires an operator-readable recovery hint",
		"watcher/notifier output is state-report-only and never review, MAR, waiver, Blue synthesis, or final acceptance authority",
	})

	requireContainsAll(t, "skills/kkachi-final-verify/SKILL.md", []string{
		"TWAKE-003 return-path policy",
		"Required async dispatches must have proven return-path evidence or explicit blocked/degraded/no_wake_claim evidence",
		"Missing return-path evidence blocks clean final/pre-commit/closeout claims",
		"operator-readable recovery hint",
	})
}
