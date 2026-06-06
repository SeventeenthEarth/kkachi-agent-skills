package docscontract

import (
	"strings"
	"testing"
)

func TestKASUPD003CLIContractDocumentsDryRunClassification(t *testing.T) {
	requireContainsAll(t, "docs/sot/kas-cli-contract.md", []string{
		"mode\": \"dry_run_classification\"",
		"--repo <current-upstream-kas-repo>",
		"--project-root <project-specific-kas-root>",
		"auto_copy_candidate",
		"local_only",
		"semantic_merge_required",
		"new_upstream_candidate",
		"removed_or_renamed_upstream",
		"fail_closed_conflict",
		"unchanged_mappings",
		"summary.no_action_count",
		"semantic_port_packets",
		"content_sha256",
		"Dirty current upstream source fails closed",
		"must not write packet files",
		"must not mutate checkout state",
	})
}

func TestKASUPD003ProjectStateContractNoLongerMarksClassificationPlanned(t *testing.T) {
	requireContainsAll(t, "docs/sot/project-kas-sync-state.md", []string{
		"KASUPD-003 extends that same dry-run command with read-only three-way classification",
		"represent unchanged mapped packs outside the classification vocabulary",
		"without writing packet files",
		"KASUPD-003 still does not implement approved sync writes",
		"`KASUPD-003` implements dry-run classification plus semantic-port packet evidence",
		"approved sync writes and pilot updates remain planned",
	})
	content := readRepoFile(t, "docs/sot/project-kas-sync-state.md")
	forbidden := []string{
		"while dry-run classification, semantic-port packets, and pilot updates remain planned",
		"KASUPD-002 does not implement dry-run three-way classification",
		"semantic-port packet generation, approved sync writes, or a project pilot. Those remain under KASUPD-003",
	}
	for _, phrase := range forbidden {
		if strings.Contains(content, phrase) {
			t.Fatalf("project KAS sync SOT retains stale planned KASUPD-003 wording %q", phrase)
		}
	}
}

func TestKASUPD003DocsReadmeNoLongerMarksClassificationPlanned(t *testing.T) {
	requireContainsAll(t, "docs/README.md", []string{
		"KASUPD-003 dry-run classification/semantic-port evidence are completed",
		"KASUPD-004 write-capable sync/pilot behavior remains planned",
		"dry-run classification/semantic-port evidence is completed for KASUPD-003",
		"KASUPD-004 pilot/write-capable sync remains planned",
	})
	content := readRepoFile(t, "docs/README.md")
	forbidden := []string{
		"sync classification and pilot updates remain planned",
		"dry-run classification, semantic-port packets, and pilot updates remain planned",
		"semantic-port packet generation, approved sync writes, or a project pilot. Those remain under KASUPD-003",
	}
	for _, phrase := range forbidden {
		if strings.Contains(content, phrase) {
			t.Fatalf("docs README retains stale planned KASUPD-003 wording %q", phrase)
		}
	}
}
