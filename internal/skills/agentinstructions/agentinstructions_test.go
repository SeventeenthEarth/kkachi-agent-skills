package agentinstructions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemplates(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "templates", "agent-instructions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		content := "# " + name + "\n\n" +
			ManagedBegin + "\n" +
			"Project {{project_name}} role {{repository_role}} suite {{project_suite_id}} stage {{kab_adoption_stage}} baseline {{upstream_kas_baseline}} notes {{local_authority_notes}}\n" +
			ManagedEnd + "\n\n" +
			LocalBegin + "\n" +
			"Template local block\n" +
			LocalEnd + "\n"
		if err := os.WriteFile(filepath.Join(dir, name+".tmpl"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeMalformedTemplate(t *testing.T, root string, name string) {
	t.Helper()
	writeTemplateContent(t, root, name, "# "+name+"\n\nmissing managed markers\n")
}

func writeTemplateContent(t *testing.T, root string, name string, content string) {
	t.Helper()
	dir := filepath.Join(root, "templates", "agent-instructions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".tmpl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func baseOptions(repo string, source string) Options {
	return Options{
		RepoPath:            repo,
		SourceRepoPath:      source,
		Project:             "kan-control",
		RepositoryRole:      "KAS repository",
		ProjectSuiteID:      "kan-control-kas",
		KABAdoptionStage:    "stage1_direct_codex_app_server_baseline",
		UpstreamKASBaseline: "kas-main",
		LocalAuthorityNotes: "local SOT first",
	}
}

func hasDiagnostic(result Result, code string) bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func readFileBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func planByPath(t *testing.T, result Result, path string) FilePlan {
	t.Helper()
	for _, plan := range result.FilePlans {
		if plan.Path == path {
			return plan
		}
	}
	t.Fatalf("missing plan for %s: %+v", path, result.FilePlans)
	return FilePlan{}
}

func TestDryRunCoversCreateNoWriteAndStableHash(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)

	result, err := BuildDryRun(baseOptions(repo, source))
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || !result.NoWrite.Guaranteed || result.NoWrite.RepoWriteCount != 0 {
		t.Fatalf("unexpected dry-run state: %+v", result)
	}
	if result.Summary.CountsByOutcome["create"] != 2 || result.Summary.WritableCount != 2 {
		t.Fatalf("unexpected summary: %+v", result.Summary)
	}
	if _, err := os.Stat(filepath.Join(repo, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote AGENTS.md: %v", err)
	}
	repeated, err := BuildDryRun(baseOptions(repo, source))
	if err != nil {
		t.Fatal(err)
	}
	if repeated.PlanHash != result.PlanHash {
		t.Fatalf("hash changed across identical dry-runs: %s != %s", repeated.PlanHash, result.PlanHash)
	}
}

func TestManagedUpdatePreservesProjectLocalBlockAndApplyHash(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	existing := "# AGENTS.md\n\n" + ManagedBegin + "\nold managed\n" + ManagedEnd + "\n\n" + LocalBegin + "\nkeep this local text\n" + LocalEnd + "\n"
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildDryRun(Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}})
	if err != nil {
		t.Fatal(err)
	}
	agents := planByPath(t, result, "AGENTS.md")
	if agents.Outcome != "update_managed_block" || len(agents.PreservationActions) != 1 || agents.PreservationActions[0].Action != "preserve_project_local_block" {
		t.Fatalf("expected managed update with local preservation: %+v", agents)
	}
	human := RenderHuman(result)
	if !strings.Contains(human, "Preserved project-local blocks: 1.") {
		t.Fatalf("human output did not surface preserved project-local block count: %s", human)
	}
	bad, err := Apply(Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}, "dry-run:sha256:0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if bad.OK || bad.Approval == nil || bad.Approval.MatchedCurrentPlan {
		t.Fatalf("expected mismatched approval to fail closed: %+v", bad)
	}
	applied, err := Apply(Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}, result.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if !applied.OK || applied.NoWrite.Guaranteed || applied.NoWrite.RepoWriteCount != 1 {
		t.Fatalf("unexpected apply result: %+v", applied)
	}
	data, err := os.ReadFile(filepath.Join(repo, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "Project kan-control") || !strings.Contains(content, "keep this local text") || strings.Contains(content, "old managed") {
		t.Fatalf("apply did not update managed block while preserving local text: %s", content)
	}
}

func TestDryRunBlocksUnmarkedExistingFileAndSupportsNotApplicable(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), []byte("# local only\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := BuildDryRun(Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}, NotApplicableReason: "Claude is not used by this project"})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Summary.CountsByOutcome["blocked_unmarked_existing_file"] != 1 || result.Summary.CountsByOutcome["not_applicable"] != 1 {
		t.Fatalf("expected blocked + not_applicable: %+v", result)
	}
	if planByPath(t, result, "AGENTS.md").Outcome != "blocked_unmarked_existing_file" {
		t.Fatalf("expected AGENTS.md to block: %+v", result.FilePlans)
	}
	if planByPath(t, result, "CLAUDE.md").Reason != "Claude is not used by this project" {
		t.Fatalf("missing not_applicable reason: %+v", result.FilePlans)
	}
}

func TestApplyBlockedUnmarkedFileRejectsAndPreservesBytes(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	target := filepath.Join(repo, "AGENTS.md")
	original := []byte("# local only\nkeep exactly\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	opts := Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}
	dryRun := mustDryRun(t, opts)
	if dryRun.OK || dryRun.ApprovalRequest.EvidenceRef == "" {
		t.Fatalf("expected blocked dry-run with syntactically valid evidence ref: %+v", dryRun)
	}

	applied, err := Apply(opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if applied.OK || !hasDiagnostic(applied, "agent_instructions_plan_not_approvable") {
		t.Fatalf("expected non-approvable blocked apply: %+v", applied)
	}
	if got := string(readFileBytes(t, target)); got != string(original) {
		t.Fatalf("blocked apply changed file bytes: got %q want %q", got, original)
	}
}

func TestApplyMalformedApprovalEvidenceRejectsAndPreservesBytes(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	target := filepath.Join(repo, "AGENTS.md")
	original := []byte("# AGENTS.md\n\n" + ManagedBegin + "\nold managed\n" + ManagedEnd + "\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	opts := Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}

	applied, err := Apply(opts, "not-a-dry-run-hash")
	if err != nil {
		t.Fatal(err)
	}
	if applied.OK || !hasDiagnostic(applied, "approval_evidence_malformed") {
		t.Fatalf("expected malformed approval rejection: %+v", applied)
	}
	if got := string(readFileBytes(t, target)); got != string(original) {
		t.Fatalf("malformed approval changed file bytes: got %q want %q", got, original)
	}
}

func TestApplyStaleTemplateHashRejectsAndPreservesBytes(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	target := filepath.Join(repo, "AGENTS.md")
	original := []byte("# AGENTS.md\n\n" + ManagedBegin + "\nold managed\n" + ManagedEnd + "\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	opts := Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}
	dryRun := mustDryRun(t, opts)
	if !dryRun.OK {
		t.Fatalf("expected approvable dry-run: %+v", dryRun)
	}
	changedTemplate := "# AGENTS.md\n\n" + ManagedBegin + "\nchanged managed {{project_name}}\n" + ManagedEnd + "\n"
	if err := os.WriteFile(filepath.Join(source, "templates", "agent-instructions", "AGENTS.md.tmpl"), []byte(changedTemplate), 0o644); err != nil {
		t.Fatal(err)
	}

	applied, err := Apply(opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if applied.OK || !hasDiagnostic(applied, "approval_hash_mismatch") {
		t.Fatalf("expected stale/tampered template hash mismatch: %+v", applied)
	}
	if got := string(readFileBytes(t, target)); got != string(original) {
		t.Fatalf("stale template apply changed file bytes: got %q want %q", got, original)
	}
}

func TestMalformedSourceTemplateFailsClosedWithoutDeletingManagedBlock(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	writeMalformedTemplate(t, source, "AGENTS.md")
	target := filepath.Join(repo, "AGENTS.md")
	original := []byte("# AGENTS.md\n\n" + ManagedBegin + "\nold managed\n" + ManagedEnd + "\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	opts := Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}

	dryRun, err := BuildDryRun(opts)
	if err != nil {
		t.Fatal(err)
	}
	if dryRun.OK || !hasDiagnostic(dryRun, "source_template_missing_managed_block") {
		t.Fatalf("expected malformed source template to fail closed: %+v", dryRun)
	}
	human := RenderHuman(dryRun)
	if !strings.Contains(human, "Errors: 1.") || !strings.Contains(human, "source_template_missing_managed_block") {
		t.Fatalf("human output did not surface source-template error: %s", human)
	}
	applied, err := Apply(opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if applied.OK || !hasDiagnostic(applied, "agent_instructions_plan_not_approvable") {
		t.Fatalf("expected malformed source template apply to be non-approvable: %+v", applied)
	}
	if got := string(readFileBytes(t, target)); got != string(original) {
		t.Fatalf("malformed template path changed managed block: got %q want %q", got, original)
	}
}

func TestReversedSourceTemplateMarkersFailClosedWithoutDeletingManagedBlock(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	writeTemplateContent(t, source, "AGENTS.md", "# AGENTS.md\n\n"+ManagedEnd+"\nreversed\n"+ManagedBegin+"\n")
	target := filepath.Join(repo, "AGENTS.md")
	original := []byte("# AGENTS.md\n\n" + ManagedBegin + "\nold managed\n" + ManagedEnd + "\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	opts := Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}

	dryRun := mustDryRun(t, opts)
	if dryRun.OK || !hasDiagnostic(dryRun, "source_template_missing_managed_block") {
		t.Fatalf("expected reversed source markers to fail closed: %+v", dryRun)
	}
	applied, err := Apply(opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if applied.OK || !hasDiagnostic(applied, "agent_instructions_plan_not_approvable") {
		t.Fatalf("expected reversed source marker apply to be non-approvable: %+v", applied)
	}
	if got := string(readFileBytes(t, target)); got != string(original) {
		t.Fatalf("reversed template path changed managed block: got %q want %q", got, original)
	}
}

func TestDuplicateSourceTemplateMarkersFailClosed(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	writeTemplateContent(t, source, "AGENTS.md", "# AGENTS.md\n\n"+ManagedBegin+"\nfirst\n"+ManagedBegin+"\nsecond\n"+ManagedEnd+"\n")
	opts := Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}

	dryRun := mustDryRun(t, opts)
	if dryRun.OK || !hasDiagnostic(dryRun, "source_template_missing_managed_block") {
		t.Fatalf("expected duplicate source markers to fail closed: %+v", dryRun)
	}
	if _, err := os.Stat(filepath.Join(repo, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run with duplicate source markers wrote AGENTS.md: %v", err)
	}
}

func TestMalformedExistingManagedMarkersFailClosedWithoutRewrite(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	target := filepath.Join(repo, "AGENTS.md")
	original := []byte("# AGENTS.md\n\n" + ManagedEnd + "\nreversed existing\n" + ManagedBegin + "\n")
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	opts := Options{RepoPath: repo, SourceRepoPath: source, Project: "kan-control", NotApplicable: []string{"CLAUDE.md"}}

	dryRun := mustDryRun(t, opts)
	if dryRun.OK || !hasDiagnostic(dryRun, "existing_managed_block_malformed") {
		t.Fatalf("expected malformed existing markers to fail closed: %+v", dryRun)
	}
	applied, err := Apply(opts, dryRun.ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if applied.OK || !hasDiagnostic(applied, "agent_instructions_plan_not_approvable") {
		t.Fatalf("expected malformed existing marker apply to be non-approvable: %+v", applied)
	}
	if got := string(readFileBytes(t, target)); got != string(original) {
		t.Fatalf("malformed existing marker path changed bytes: got %q want %q", got, original)
	}
}

func TestApplyFailureResultReportsPriorWrites(t *testing.T) {
	base := Result{OK: true, NoWrite: noWriteEvidence(true)}
	partial := applyFailureResult(base, 1, Diagnostic{Level: "error", Code: "agent_instruction_write_verification_failed", Message: "verify failed"}, "inspect file state")
	if partial.OK || partial.NoWrite.Guaranteed || partial.NoWrite.RepoWriteCount != 1 || !hasDiagnostic(partial, "agent_instruction_write_verification_failed") {
		t.Fatalf("partial apply failure must report prior writes: %+v", partial)
	}
	beforeWrite := applyFailureResult(base, 0, Diagnostic{Level: "error", Code: "planned_content_unavailable", Message: "missing"}, "rerun dry-run")
	if beforeWrite.OK || !beforeWrite.NoWrite.Guaranteed || beforeWrite.NoWrite.RepoWriteCount != 0 || !hasDiagnostic(beforeWrite, "planned_content_unavailable") {
		t.Fatalf("pre-write apply failure should preserve no-write guarantee: %+v", beforeWrite)
	}
}

func TestDryRunReportsNoChangeForCurrentManagedFile(t *testing.T) {
	source := t.TempDir()
	repo := t.TempDir()
	writeTemplates(t, source)
	opts := baseOptions(repo, source)
	created, err := Apply(opts, mustDryRun(t, opts).ApprovalRequest.EvidenceRef)
	if err != nil {
		t.Fatal(err)
	}
	if !created.OK {
		t.Fatalf("create apply failed: %+v", created)
	}
	result, err := BuildDryRun(opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.CountsByOutcome["no_change"] != 2 || result.Summary.WritableCount != 0 {
		t.Fatalf("expected no_change after apply: %+v", result.Summary)
	}
}

func mustDryRun(t *testing.T, opts Options) Result {
	t.Helper()
	result, err := BuildDryRun(opts)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
