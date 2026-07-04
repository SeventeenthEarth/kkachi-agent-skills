package docscontract

import (
	"strings"
	"testing"
)

func TestStage1DirectCodexRunnerSOTDefinesSDKAppServerContract(t *testing.T) {
	requireContainsAll(t, "docs/sot/stage1-direct-codex-sdk-appserver-runner.md", []string{
		"KAS runner template -> openai_codex SDK -> codex app-server --listen stdio:// -> Codex JSON-RPC thread/run lifecycle",
		"must not call `codex exec`",
		"must not use the generic `openai` Python SDK",
		"templates/runners/direct-codex-sdk-appserver-runner.py.tmpl",
		"Sandbox.read_only",
		"Sandbox.workspace_write",
		"Hermes-tracked background process",
		"completion notification and bounded polling/watch evidence",
		"a single foreground Hermes terminal call whose tool timeout can kill the parent before metadata/output artifacts are flushed",
		"thread_id",
		"one Codex thread per KAS/KAH task",
		"plan-only and plan-revision turns use effort `high` by default",
		"implementation, feedback-fix, cleanup, and verification-support turns use effort `medium` by default",
		"no-KAB-Codex rationale",
		"Stage 2 KAB Codex-first",
		"native_codex",
	})
}

func TestStage1DirectCodexRunnerTemplateDefinesExecutableContract(t *testing.T) {
	requireContainsAll(t, "templates/runners/direct-codex-sdk-appserver-runner.py.tmpl", []string{
		"from openai_codex import ApprovalMode, Codex, CodexConfig, Sandbox",
		"SDK_APP_SERVER_MODE = \"openai_codex -> codex app-server --listen stdio://\"",
		"--no-kab-codex-rationale",
		"--preflight-only",
		"thread_resume",
		"thread_start",
		"thread.run",
		"templates/runners/direct-codex-sdk-appserver-runner.py.tmpl",
		"KAB native_codex session evidence",
		"unsupported stage marker",
	})
}

func TestStage1DirectCodexRunnerGuidanceWiredIntoSkillsAndTemplates(t *testing.T) {
	for _, rel := range []string{
		"skills/kkachi-orchestrate/SKILL.md",
		"skills/kkachi-task-contract/SKILL.md",
		"skills/kkachi-backend-select/SKILL.md",
		"skills/kkachi-prompt-compose/SKILL.md",
		"skills/kkachi-plan/SKILL.md",
		"skills/kkachi-implement/SKILL.md",
		"skills/kkachi-handle-feedback/SKILL.md",
		"skills/kkachi-verify/SKILL.md",
		"skills/kkachi-docs-update/SKILL.md",
		"skills/kkachi-final-verify/SKILL.md",
	} {
		requireContainsAll(t, rel, []string{
			"v0.2",
			"GJC",
			"KAB",
		})
	}
	requireContainsAll(t, "skills/kkachi-plan/SKILL.md", []string{
		"Legacy Stage 1/Stage 2/Stage 3 Codex/KAB adoption wording is historical/stale",
		"GJC `ralplan` candidate path",
	})
	requireContainsAll(t, "skills/kkachi-implement/SKILL.md", []string{
		"Legacy Stage 1/Stage 2/Stage 3",
		"not active operator guidance",
		"GJC `ultragoal` may produce implementation-candidate artifacts",
	})
}

func TestStage1DirectCodexRunnerRegisteredInDocsAndRoadmap(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"sot/stage1-direct-codex-sdk-appserver-runner.md",
		"Deprecated after V01CLEAN",
		"not active v0.2 operator guidance",
	})
	readme := readRepoFile(t, "docs/README.md")
	if strings.Contains(readme, "future Python runner template") {
		t.Fatalf("docs/README.md contains stale future-runner wording")
	}

	requireContainsAll(t, "docs/kkachi-docs-map.yaml", []string{
		"docs/sot/stage1-direct-codex-sdk-appserver-runner.md",
	})

	requireContainsAll(t, "docs/roadmap.md", []string{
		"CODEXSDK-001",
		"CODEXSDK-002",
		"CODEXSDK-003",
		"historical Stage 1 direct Codex SDK/app-server runner support",
		"V01CLEAN status: historical/deprecated",
		"not active v0.2 operator guidance",
	})

	requireContainsAll(t, "docs/sot/khs-architecture-and-integration.md", []string{
		"Direct Codex SDK/app-server baseline",
		"KAS Python runner template -> `openai_codex` SDK -> SDK-managed `codex app-server --listen stdio://`",
		"Stage 1 runner must not be used as a silent fallback",
	})
}

func TestStage1DirectCodexRunnerDocsRejectDriftTerms(t *testing.T) {
	for _, rel := range []string{
		"docs/sot/stage1-direct-codex-sdk-appserver-runner.md",
		"docs/roadmap.md",
		"docs/sot/khs-architecture-and-integration.md",
	} {
		content := readRepoFile(t, rel)
		forbidden := []string{
			"use codex exec for Stage 1",
			"generic openai SDK is acceptable Stage 1 evidence",
			"silently fall back to Stage 1",
			"Stage 1 KAB Codex evidence",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Fatalf("%s contains forbidden Stage 1 Codex runner drift wording %q", rel, phrase)
			}
		}
	}
}
