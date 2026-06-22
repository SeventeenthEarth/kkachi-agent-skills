package docscontract

import (
	"strings"
	"testing"
)

func TestDesign002TealApplicabilityContractIsDocumented(t *testing.T) {
	requireContainsAll(t, "docs/sot/teal-ui-workflow-policy.md", []string{
		"project_has_teal_lane: true|false",
		"ui_ux_change: true|false",
		"teal_required: project_has_teal_lane && ui_ux_change",
		"teal_skip_reason",
		"Teal waiver evidence is not a skip reason.",
		"teal_waiver_approved",
		"teal_waiver_approval_ref",
		"teal_waiver_scope",
		"teal_waiver_expires_at",
		"DESIGN_PLAN_GATE",
		"DESIGN_FIDELITY_REVIEW",
		"must not substitute Blue, Red, Orange, Gray, MAR, backend agents, or temporary helpers for official Teal verdicts",
	})
}

func TestDesign002TaskContractAndTaxonomyCarryTealFields(t *testing.T) {
	requireContainsAll(t, "registries/task-taxonomy.yaml", []string{
		"teal_applicability_fields:",
		"project_has_teal_lane",
		"ui_ux_change",
		"teal_required",
		"teal_skip_reason",
		"teal_waiver_approved",
		"teal_waiver_approval_ref",
		"teal_waiver_scope",
		"teal_waiver_expires_at",
		"DESIGN_PLAN_GATE",
		"DESIGN_FIDELITY_REVIEW",
		"required_teal_verdict_missing",
	})
	requireContainsAll(t, "templates/run-artifacts/task-contract.yaml.tmpl", []string{
		"teal_applicability:",
		"project_has_teal_lane: {{ project_has_teal_lane }}",
		"ui_ux_change: {{ ui_ux_change }}",
		"teal_required: {{ derived_teal_required }}",
		"derivation: \"project_has_teal_lane && ui_ux_change\"",
		"teal_skip_reason: \"{{ teal_skip_reason }}\"",
		"teal_waiver_approved: {{ teal_waiver_approved }}",
		"teal_waiver_approval_ref: \"{{ teal_waiver_approval_ref }}\"",
		"teal_waiver_scope: \"{{ teal_waiver_scope }}\"",
		"teal_waiver_expires_at: \"{{ teal_waiver_expires_at }}\"",
	})

	template := readRepoFile(t, "templates/run-artifacts/task-contract.yaml.tmpl")
	for _, forbidden := range []string{
		"teal_required: \"{{ derived_teal_required }}\"",
		"teal_waiver_approved: \"{{ teal_waiver_approved }}\"",
	} {
		if strings.Contains(template, forbidden) {
			t.Fatalf("task-contract template must preserve typed-safe Teal booleans, found quoted boolean placeholder %q", forbidden)
		}
	}
}

func TestDesign002NodeContractSemanticsStaySeparated(t *testing.T) {
	requireContainsAll(t, "docs/sot/task-dag-workflow-contract.md", []string{
		"DESIGN-002",
		"DESIGN_PLAN_GATE",
		"before implementation authorization",
		"DESIGN_FIDELITY_REVIEW",
		"before final acceptance",
		"ordinary Red/Orange/Gray/Blue color review remains separate",
		"MAR remains separate",
		"backend implementation evidence remains separate",
		"required_teal_verdict_missing",
		"none_fail_closed",
	})
	requireContainsAll(t, "templates/run-artifacts/phase-plan.yaml.tmpl", []string{
		"design_teal_policy:",
		"contract_version: \"design002.v1\"",
		"design_plan_gate: \"DESIGN_PLAN_GATE\"",
		"design_fidelity_review: \"DESIGN_FIDELITY_REVIEW\"",
		"required_teal_verdict_missing",
		"ordinary_color_review_is_substitute: false",
		"mar_review_is_substitute: false",
		"backend_evidence_is_substitute: false",
	})
}

func TestDesign002GuidanceRejectsSubstitutionAndUniversalInjection(t *testing.T) {
	for _, rel := range []string{
		"skills/kkachi-task-contract/SKILL.md",
		"skills/kkachi-plan/SKILL.md",
		"skills/kkachi-review/SKILL.md",
		"skills/kkachi-final-verify/SKILL.md",
	} {
		requireContainsAll(t, rel, []string{
			"project_has_teal_lane",
			"ui_ux_change",
			"teal_required",
			"teal_skip_reason",
			"DESIGN_PLAN_GATE",
			"DESIGN_FIDELITY_REVIEW",
			"required Teal verdict",
			"must not substitute",
		})
	}

	registry := readRepoFile(t, "registries/task-dag-workflow-registry.yaml")
	for _, forbidden := range []string{
		"DESIGN_PLAN_GATE",
		"DESIGN_FIDELITY_REVIEW",
		"node_id: design_plan_gate",
		"node_id: design_fidelity_review",
		"teal_required: true",
	} {
		if strings.Contains(registry, forbidden) {
			t.Fatalf("standard workflow registry must not universally inject Teal gates into non-UI Kkachi source work; found %q", forbidden)
		}
	}
}
