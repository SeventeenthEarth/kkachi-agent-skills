package docscontract

import "testing"

func TestNEWMAR011PhasePlanAuditGuidance(t *testing.T) {
	requireContainsAll(t, "skills/kkachi-implement/SKILL.md", []string{
		"do not edit `phase-plan.yaml` directly",
		"do not use ordinary `phase-plan set` for the corrective change",
		"kkachi-agent-helper phase-plan reopen <run_id> <phase-id> --from-status <current> --to-status <target> --reason <text> --evidence-ref .kkachi/runs/<run_id>/<evidence>",
		"kkachi-agent-helper phase-plan amend <run_id> <phase-id> --kind <correction|supersede|rollback> --from-status <current> --to-status <target> --reason <text> --evidence-ref .kkachi/runs/<run_id>/<evidence>",
		"KAH records deterministic state facts only; KAS/Blue/color/MAR/final remain semantic acceptance authority.",
	})

	requireContainsAll(t, "skills/kkachi-final-verify/SKILL.md", []string{
		"phase-plan-audit.jsonl",
		"reopen/amend records resolve to terminal phase states",
		"kkachi-agent-helper phase-plan validate <run_id> --final",
	})
}
