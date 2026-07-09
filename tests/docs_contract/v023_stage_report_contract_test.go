package docscontract

import (
	"strings"
	"testing"
)

func TestV023StageReportContractSurfaces(t *testing.T) {
	requireContainsAll(t, "docs/sot/stage-report-contract.md", []string{
		"KAS Stage Report Contract",
		"functional change",
		"Task classification report requirements",
		"Feature/change identity",
		"Need reason",
		"Document direction",
		"Plan-stage report requirements",
		"Plan drift",
		"Color-vet feedback",
		"Implementation-stage report requirements",
		"Behavior before/after",
		"Phase-step coverage",
		"Review-stage report requirements",
		"Color review result",
		"MAR result",
		"Second-color adoption",
		"Defer reason",
		"what was deferred, why it was safe or necessary to defer now",
		"KAS/KAH/KAT repository development does not dogfood the KAS/KAH/KAT workflow by",
		"default. 황충 performs main development directly",
	})

	requireContainsAll(t, "README.md", []string{
		"KAS v0.2.3",
		"dogfood default: do not run KAS/KAH/KAT development through KAS/KAH/KAT",
		"workflows unless 주군 explicitly selects that mode for the run",
		"docs/sot/stage-report-contract.md",
		"only listing files or saying a stage completed",
	})

	requireContainsAll(t, "docs/README.md", []string{
		"sot/stage-report-contract.md",
		"KAS v0.2.3 stage-report contract",
		"file-list-only stage completion reports",
	})
}

func TestV023StageReportSkillGuidance(t *testing.T) {
	checks := map[string][]string{
		"skills/kkachi-task-contract/SKILL.md": {
			"feature/change identity",
			"old behavior or missing report content",
			"document/SOT/request direction",
			"docs/sot/stage-report-contract.md",
		},
		"skills/kkachi-plan/SKILL.md": {
			"do not dogfood this KAS/KAH/GJC pipeline by default",
			"Plan-stage reports to 주군",
			"original plan was followed or changed",
			"Red/Orange/Gray/Blue plan-vet feedback",
		},
		"skills/kkachi-implement/SKILL.md": {
			"KAS/KAH/KAT repository self-development is a standing exception",
			"Implementation-stage reports to 주군",
			"old behavior vs new behavior",
			"`impl`, `test-enhance`, `ai-slop-cleaner`, `optimize`, and `docs-update`",
		},
		"skills/kkachi-review/SKILL.md": {
			"Review-stage reports to 주군",
			"Red/Orange/Gray/Blue verdicts",
			"MAR status/role coverage/findings",
			"second-color adoption status",
			"what was deferred, why it was safe or necessary to defer now",
		},
		"skills/kkachi-final-verify/SKILL.md": {
			"Final reports to 주군 must enforce `docs/sot/stage-report-contract.md`",
			"functional change, old behavior/reason",
			"color/MAR/second-color feedback",
			"every deferred item with a short what/why/future-owner reason",
			"Reject file-list-only or `stage done` summaries",
		},
		"skills/kkachi-handle-feedback/SKILL.md": {
			"Deferred feedback must name what is deferred",
			"why deferring is safe or necessary now",
			"future gate/task owner",
		},
		"skills/kkachi-final-verify/references/pre-commit-completion-report-template.md": {
			"For every deferred item, state what was deferred",
			"why it was safe or necessary to defer now",
			"owning future gate/task",
		},
		"AGENTS.md": {
			"Functional change: what behavior, skill guidance, or operator workflow changed",
			"Stage-specific substance for task classification, plan, implementation, and",
			"docs/sot/stage-report-contract.md",
		},
	}
	for rel, needles := range checks {
		requireContainsAll(t, rel, needles)
	}
}

func TestV023StageReportDoesNotRetainOldCurrentTarget(t *testing.T) {
	readme := readRepoFile(t, "README.md")
	if strings.Contains(readme, "KAS v0.2.2 is the current source development target") {
		t.Fatal("README still marks v0.2.2 as current source development target")
	}
}
