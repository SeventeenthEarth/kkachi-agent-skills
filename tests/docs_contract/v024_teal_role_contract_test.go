package docscontract

import (
	"strings"
	"testing"
)

func TestV024OptionalTealRoleSupportIsDocumentedAndPackaged(t *testing.T) {
	requireContainsAll(t, "internal/skills/version/version.go", []string{
		`CLIVersion  = "0.2.5"`,
	})
	requireContainsAll(t, "skill-pack.yaml", []string{
		"teal: roles/teal.yaml",
		"- kkachi-design-review",
	})
	requireContainsAll(t, "roles/teal.yaml", []string{
		"role: teal",
		"Optional Teal UI/UX design reviewer KAS base subset.",
		"kkachi-design-review",
		"kkachi-review",
	})
	requireContainsAll(t, "skills/kkachi-design-review/SKILL.md", []string{
		"name: kkachi-design-review",
		"DESIGN_PLAN_GATE",
		"DESIGN_FIDELITY_REVIEW",
		"Teal is optional unless `project_has_teal_lane && ui_ux_change` is true",
		"Ordinary Red/Orange/Gray/Blue review, MAR, backend evidence, helper notes, and temporary subagents are not substitutes",
	})
}

func TestV024OptionalTealRoleRegistryAndSOTBoundaries(t *testing.T) {
	requireContainsAll(t, "registries/project-suite-roles.yaml", []string{
		"teal_design_reviewer:",
		`display_label: "Teal optional UI/UX design reviewer subset"`,
		"- kkachi-design-review",
		"- kkachi-review",
		"- kkachi-implement",
		"- kkachi-backend-select",
		"- kkachi-prompt-compose",
		"- kkachi-optimize",
		"- kkachi-orchestrate",
		"- kkachi-docs-update",
		"- kkachi-final-verify",
	})
	registry := readRepoFile(t, "registries/project-suite-roles.yaml")
	tealStart := strings.Index(registry, "  teal_design_reviewer:")
	if tealStart < 0 {
		t.Fatal("missing teal_design_reviewer registry stanza")
	}
	tealStanza := registry[tealStart:]
	for _, required := range []string{
		`display_label: "Teal optional UI/UX design reviewer subset"`,
		"      - kkachi-design-review",
		"      - kkachi-review",
		"      - kkachi-implement",
		"      - kkachi-backend-select",
		"      - kkachi-prompt-compose",
		"      - kkachi-optimize",
		"      - kkachi-orchestrate",
		"      - kkachi-docs-update",
		"      - kkachi-final-verify",
	} {
		if !strings.Contains(tealStanza, required) {
			t.Fatalf("teal_design_reviewer stanza missing %q:\n%s", required, tealStanza)
		}
	}
	requireContainsAll(t, "docs/sot/role-aware-project-suite-contract.md", []string{
		"`teal_design_reviewer`",
		"Teal optional UI/UX design reviewer subset",
		"Teal is optional and is not part of the mandatory Blue/Red/Orange/Gray project-suite baseline",
		"Teal workflow gates are required only when `project_has_teal_lane && ui_ux_change` is true",
		"Installing a Teal role suite must not make non-UI Kkachi source work Teal-required",
	})
	requireContainsAll(t, "docs/sot/teal-ui-workflow-policy.md", []string{
		"Teal role installation support is KAS-owned",
		"KAS v0.2.4",
		"KAH design-evidence validation does not need to change for Teal role-suite installation",
	})
	requireContainsAll(t, "skills/kkachi-design-review/SKILL.md", []string{
		"Blue commander full-suite possession of `kkachi-design-review` is operational context only",
		"it does not make Blue, MAR, or temporary helper output an official Teal verdict when `teal_required=true`",
	})
}
