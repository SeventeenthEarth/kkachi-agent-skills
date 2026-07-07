package docscontract

import "testing"

func TestV02FLOW011DeferredFeedbackLedgerContractIsDocumented(t *testing.T) {
	requireContainsAll(t, "docs/deferred-feedback.md", []string{
		"## Compact entry checklist",
		"Blue disposition ref: required; N/A is invalid",
		"Blocking current task: false is mandatory for every deferred entry",
		"Blocker finding handling: fix_now_or_hold_current_task",
		"Lifecycle status is one of: open, converted_to_task, resolved, rejected, stale",
		"Waiver is authority evidence, not lifecycle status",
		"Logical/readback status hash:",
		"Byte-level ledger file SHA:",
		"Final-report reciprocal ref:",
		"Source finding ref:",
		"Converted task ref:",
		"Blue disposition ref: <non-empty Blue synthesis/card/event/final-disposition ref; N/A is invalid>",
		"Logical/readback status hash: sha256:<semantic-status-hash or N/A before KAH V02FLOW-012>",
		"Byte-level ledger file SHA: sha256:<file-sha or N/A before readback>",
		"Source finding ref: <exact source review/MAR/Gray/Blue/user card, artifact, line, or event ref>",
		"Final-report reciprocal ref: <final report or Blue synthesis anchor that links this deferred id>",
	})
}

func TestV02FLOW011DocsMapAndRoadmapRegisterImplementationState(t *testing.T) {
	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"deferred_feedback: \"docs/deferred-feedback.md\"",
	})
	requireRoadmapTaskStatus(t, "V02FLOW-011", "Completed")
	requireContainsAll(t, "docs/roadmap.md", []string{
		"sectioned compact ledger template",
		"non-empty Blue disposition ref; N/A is invalid",
		"blockers cannot be deferred",
		"KAH V02FLOW-012 remains companion readback/fail-closed validation",
	})
}

func TestV02FLOW011SOTCarriesNoBlockerDeferAndRefSemantics(t *testing.T) {
	requireContainsAll(t, "docs/sot/v020-gjc-workflow-train-corrections.md", []string{
		"Blocker finding handling: fix_now_or_hold_current_task",
		"Blue disposition ref is required and N/A is invalid",
		"waiver is authority evidence, not lifecycle status",
		"logical/readback status hash distinct from byte-level ledger file SHA",
		"final-report reciprocal ref",
		"source finding ref",
	})
}
