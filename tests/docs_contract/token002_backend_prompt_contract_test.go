package docscontract

import (
	"strings"
	"testing"
)

var token002CompactSchema = []string{
	"Status",
	"Summary",
	"Files",
	"Verification",
	"Risks/blockers",
	"Detailed artifact",
	"Next action requested",
}

var token002PromptTemplates = []string{
	"templates/prompts/claude/path-a-implement.md.tmpl",
	"templates/prompts/claude/path-b-shaping.md.tmpl",
	"templates/prompts/claude/review.md.tmpl",
	"templates/prompts/codex/path-a-implement.md.tmpl",
	"templates/prompts/codex/path-b-shaping.md.tmpl",
	"templates/prompts/codex/review.md.tmpl",
	"templates/prompts/gemini/path-a-implement.md.tmpl",
	"templates/prompts/glm/path-a-implement.md.tmpl",
	"templates/prompts/opencode/path-a-implement.md.tmpl",
}

var token002PhaseSkills = []string{
	"skills/kkachi-plan/SKILL.md",
	"skills/kkachi-prompt-compose/SKILL.md",
	"skills/kkachi-task-contract/SKILL.md",
	"skills/kkachi-orchestrate/SKILL.md",
	"skills/kkachi-implement/SKILL.md",
	"skills/kkachi-verify/SKILL.md",
	"skills/kkachi-final-verify/SKILL.md",
}

func TestToken002BackendProfilesAndTemplatesRequireCompactArtifactFirstOutput(t *testing.T) {
	requireContainsAll(t, "registries/backend-prompt-profiles.yaml", append(token002CompactSchema,
		"direct_codex_app_server",
		"kab_mediated_backend_lanes",
		".kkachi/runs/<run_id>/artifacts/<phase>/backend-<phase>.md",
		"write or update the requested artifact",
		"Status: blocked",
		"do not dump the full detail into chat",
	))

	for _, rel := range token002PromptTemplates {
		requireContainsAll(t, rel, token002CompactSchema)
		requireContainsAll(t, rel, []string{
			"Product/backend output must be English",
			".kkachi/runs/<run_id>/artifacts/<phase>/backend-<phase>.md",
			"If the detailed artifact cannot be written or updated",
			"dump",
			"chat",
		})
	}
}

func TestToken002RunArtifactsCarryPolicyAndConcretePhasePath(t *testing.T) {
	for _, rel := range []string{
		"templates/run-artifacts/prompt.md.tmpl",
		"templates/run-artifacts/plan.md.tmpl",
		"templates/run-artifacts/checklist.md.tmpl",
		"templates/run-artifacts/phase-plan.yaml.tmpl",
	} {
		requireContainsAll(t, rel, token002CompactSchema)
		requireContainsAll(t, rel, []string{
			".kkachi/runs/<run_id>/artifacts/<phase>/backend-<phase>.md",
			"artifact",
		})
	}

	requireContainsAll(t, "templates/run-artifacts/phase-plan.yaml.tmpl", []string{
		".kkachi/runs/<run_id>/artifacts/implement/backend-implement.md",
		"direct_codex_app_server",
		"kab_mediated_backend_lanes",
	})
}

func TestToken002PhaseSkillsCarryPolicyWithoutRemovingKoreanCommanderReports(t *testing.T) {
	for _, rel := range token002PhaseSkills {
		requireContainsAll(t, rel, token002CompactSchema)
		requireContainsAll(t, rel, []string{
			".kkachi/runs/<run_id>/artifacts",
			"v0.2",
		})
	}

	requireContainsAll(t, "skills/kkachi-orchestrate/SKILL.md", []string{
		"Commander chat reports to 주군 may remain Korean",
		"backend prompt/product output must preserve the compact schema",
	})
	requireContainsAll(t, "skills/kkachi-final-verify/SKILL.md", []string{
		"The separate commander-facing Korean report to 주군 may summarize those English artifacts.",
	})
}

func TestToken002TaskClassAliasesAndGatingStayExplicit(t *testing.T) {
	for _, rel := range []string{
		"registries/task-taxonomy.yaml",
		"registries/phase-contracts.yaml",
		"templates/run-artifacts/phase-plan.yaml.tmpl",
	} {
		requireContainsAll(t, rel, []string{
			"simple_report",
			"simple_command_report",
			"investigation",
			"research_evidence",
			"review",
			"collaboration_review",
			"docs_only",
			"enhance_test",
			"optimize",
			"broad review loops",
		})
	}

	requireContainsAll(t, "skills/kkachi-task-contract/SKILL.md", []string{
		"`simple_report`/`simple_command_report`",
		"`investigation`/`research_evidence`",
		"`review`/`collaboration_review`",
		"`docs_only`/`docs_only`",
		"skipped-phase reasons",
	})
}

func TestToken002AvoidsPermissiveLongOutputFallbackWording(t *testing.T) {
	relPaths := append([]string{
		"registries/backend-prompt-profiles.yaml",
		"registries/phase-contracts.yaml",
		"templates/run-artifacts/prompt.md.tmpl",
		"templates/run-artifacts/plan.md.tmpl",
		"templates/run-artifacts/checklist.md.tmpl",
		"templates/run-artifacts/phase-plan.yaml.tmpl",
	}, token002PromptTemplates...)
	relPaths = append(relPaths, token002PhaseSkills...)

	for _, rel := range relPaths {
		content := readRepoFile(t, rel)
		forbidden := []string{
			"dump the full detail into chat when",
			"dump full details into chat",
			"paste full logs into chat",
			"paste full diffs into chat",
			"paste full file contents into chat",
			"fall back to chat dump",
			"fallback to raw chat",
			"use Korean product output",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Fatalf("%s contains permissive TOKEN-002 fallback wording %q", rel, phrase)
			}
		}
	}
}
