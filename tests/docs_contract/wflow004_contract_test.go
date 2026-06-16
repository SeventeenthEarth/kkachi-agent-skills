package docscontract

import "testing"

func TestWFLOW004WorkflowCreateContractIsDocumentedAndBounded(t *testing.T) {
	cases := map[string][]string{
		"skills/kkachi-workflow-create/SKILL.md": {
			"kkachi-agent-skills workflow-create",
			"dag_only|thin_trigger|full_trigger",
			"kas-workflow-create-request/v1",
			"kas-workflow-create-approval/v1",
			"target_paths",
			"generated_content",
			"selector_metadata",
			"capability_evidence",
			"base_checksums",
			"changed_paths",
			"no_write",
			"approval_hash",
			"selector_metadata_invalid",
			"KAH v0.1.10 is the first release line",
			"workflow_catalog_proposal_apply=true",
			"workflow catalog proposal/apply",
			"direct KAH state",
			"fallback agent/backend",
		},
		"docs/sot/kas-cli-contract.md": {
			"WFLOW-004",
			"workflow-create",
			"dag_only",
			"thin_trigger",
			"full_trigger",
			"approval hash",
			"target_paths",
			"changed_paths",
			"generated_content",
			"selector_metadata_invalid",
			"Installed KAH `0.1.9` lacks the workflow command group",
			"does not advertise a reviewed workflow catalog proposal/apply command mapping",
		},
		"docs/sot/task-dag-workflow-contract.md": {
			"WFLOW-004",
			"workflow-create",
			"canonical `sha256`",
			"selector_metadata_invalid",
			"blocked_missing_kah_workflow_catalog_capability",
			"generated_skill_validation_failed",
			"no automatic fallback",
			"Installed KAH `0.1.9`",
		},
		"templates/run-artifacts/workflow-create-packet.yaml.tmpl": {
			"schema_version: kas-workflow-create-packet/v1",
			"approval_schema: kas-workflow-create-approval/v1",
			"canonicalization: utf8-json-sorted-keys-normalized-relative-paths/v1",
			"candidate_paths:",
			"generated_content:",
			"selector_metadata:",
			"capability_evidence:",
			"base_checksums:",
			"changed_paths:",
			"no_write:",
			"approval_hash:",
		},
		"templates/workflow-triggers/thin-trigger/SKILL.md.tmpl": {
			"kkachi-agent-skills workflow-trigger",
			"fallback workflows",
			"fallback agents",
			"fallback backends",
			"direct-write KAH state",
		},
		"skill-pack.yaml": {
			"kkachi-workflow-create",
		},
	}
	for rel, needles := range cases {
		requireContainsAll(t, rel, needles)
	}
}
