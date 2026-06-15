package docscontract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWFLOW002WorkflowTriggerContractIsRegisteredAndBounded(t *testing.T) {
	root := repoRoot(t)
	skill := readWFLOW002File(t, root, "skills/kkachi-workflow-trigger/SKILL.md")
	cli := readWFLOW002File(t, root, "docs/sot/kas-cli-contract.md")
	pack := readWFLOW002File(t, root, "skill-pack.yaml")
	template := readWFLOW002File(t, root, "templates/run-artifacts/workflow-dispatch-packet.yaml.tmpl")

	for label, content := range map[string]string{
		"skill":    skill,
		"cli":      cli,
		"template": template,
	} {
		for _, want := range []string{
			"direct_kah_state_write:false",
			"none_fail_closed",
		} {
			if !strings.Contains(strings.ReplaceAll(content, " ", ""), strings.ReplaceAll(want, " ", "")) {
				t.Fatalf("%s missing %q", label, want)
			}
		}
	}
	for _, want := range []string{
		"kkachi-agent-skills workflow-trigger",
		"--workflow-id <id>",
		"--node-contract-source <path>",
		"JSON-only",
		"WFLOW-003",
		"WFLOW-004",
		"selector search",
		"direct KAH state",
	} {
		if !strings.Contains(skill, want) && !strings.Contains(cli, want) {
			t.Fatalf("WFLOW-002 docs missing %q", want)
		}
	}
	if !strings.Contains(pack, "kkachi-workflow-trigger") {
		t.Fatalf("skill-pack.yaml does not register kkachi-workflow-trigger")
	}
}

func readWFLOW002File(t *testing.T, root string, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
