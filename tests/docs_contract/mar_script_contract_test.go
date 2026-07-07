package docscontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestKASV021LegacyPythonMARSourcesAreRemoved(t *testing.T) {
	for _, rel := range []string{
		"scripts/mar.py",
		"scripts/mar_adapters/mar-zcode.sh",
		"scripts/mar_adapters/mar-kimi.sh",
		"scripts/mar_adapters/mar-agy.sh",
	} {
		if _, err := os.Stat(filepath.Join(repoRoot(t), rel)); !os.IsNotExist(err) {
			t.Fatalf("legacy Python MAR source surface %s must be absent in KAS v0.2.1 OFF/HOLD posture; stat err=%v", rel, err)
		}
	}

	mainGo := readRepoFile(t, "main.go")
	if strings.Contains(mainGo, " registries scripts") || strings.Contains(mainGo, "scripts\nvar embeddedSource") {
		t.Fatalf("main.go must not embed legacy scripts in KAS v0.2.1:\n%s", mainGo)
	}
	if !strings.Contains(mainGo, "//go:embed skills skill-pack.yaml templates registries") {
		t.Fatalf("main.go must keep the scripts-free embedded source declaration:\n%s", mainGo)
	}
}

func TestKASV021MARRegistryIsProviderPolicyOnly(t *testing.T) {
	var registry struct {
		SchemaVersion         string                    `json:"schema_version"`
		RuntimePosture        string                    `json:"runtime_posture"`
		LegacyPythonMARStatus string                    `json:"legacy_python_mar_status"`
		RequiredRoles         []string                  `json:"required_roles"`
		Roles                 map[string]map[string]any `json:"roles"`
		Providers             map[string]map[string]any `json:"providers"`
	}
	if err := json.Unmarshal([]byte(readRepoFile(t, "registries/mar-provider-lanes.json")), &registry); err != nil {
		t.Fatalf("parse MAR provider lane registry: %v", err)
	}
	if registry.SchemaVersion != "mar.role_lanes.v1" {
		t.Fatalf("schema_version = %q", registry.SchemaVersion)
	}
	if registry.RuntimePosture != "kas_policy_only_kah_execution_required" || registry.LegacyPythonMARStatus != "off_hold_until_go_kah_mar_v0_2_1" {
		t.Fatalf("registry must record KAS policy-only / legacy Python MAR OFF posture: %+v", registry)
	}
	for _, role := range []string{"logic", "security", "arch", "cve", "test_adequacy"} {
		if _, ok := registry.Roles[role]; !ok || !hasString(registry.RequiredRoles, role) {
			t.Fatalf("registry missing required MAR role %q", role)
		}
	}
	for providerID, provider := range registry.Providers {
		for _, forbidden := range []string{"executable", "adapter_script", "command_args"} {
			if _, ok := provider[forbidden]; ok {
				t.Fatalf("provider %s must not carry KAS-owned executable field %q in v0.2.1: %+v", providerID, forbidden, provider)
			}
		}
		encoded, _ := json.Marshal(provider)
		if strings.Contains(string(encoded), "scripts/mar") || strings.Contains(string(encoded), "mar_adapters") {
			t.Fatalf("provider %s must not point at removed KAS script/adapter paths: %s", providerID, encoded)
		}
		if provider["execution_owner"] != "KAH" || provider["provider_binding_source"] != "kah_toolchain_provider_registry" || provider["no_kas_provider_execution"] != true {
			t.Fatalf("provider %s must declare KAH-owned execution boundary: %+v", providerID, provider)
		}
		if provider["prompt_template"] == "" || provider["timeout_seconds"] == nil {
			t.Fatalf("provider %s must retain portable KAS policy/template metadata: %+v", providerID, provider)
		}
	}
}

func TestKASV021MARGuidanceRoutesRuntimeToKAH(t *testing.T) {
	surfaces := map[string]string{
		"docs/sot/multi-agent-review-policy.md":     readRepoFile(t, "docs/sot/multi-agent-review-policy.md"),
		"skills/kkachi-multi-agent-review/SKILL.md": readRepoFile(t, "skills/kkachi-multi-agent-review/SKILL.md"),
		"docs/README.md": readRepoFile(t, "docs/README.md"),
	}
	for path, content := range surfaces {
		for _, want := range []string{
			"pre-v0.2.1 legacy Python MAR is OFF/HOLD",
			"KAH `mar` owns reviewed provider execution",
			"registries/mar-provider-lanes.json",
		} {
			if !strings.Contains(content, want) {
				t.Fatalf("%s missing %q", path, want)
			}
		}
		for _, forbidden := range []string{
			"scripts/mar.py kah-trigger",
			"scripts/mar.py role-lanes",
			"scripts/mar.py provider-attempt",
			"bundled `scripts/mar.py`",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s still advertises removed legacy MAR surface %q", path, forbidden)
			}
		}
	}
}

func TestKASV021MARFixturesRemainDataOnly(t *testing.T) {
	for _, name := range []string{
		"pass-review.json",
		"request-changes-review.json",
		"provider-unavailable.json",
		"all-failed.json",
		"insufficient-coverage.json",
		"mutation-detected.json",
	} {
		data, err := os.ReadFile(filepath.Join(repoRoot(t), "tests", "fixtures", "mar", name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		var fixture struct {
			Case   string `json:"case"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &fixture); err != nil {
			t.Fatalf("parse fixture %s: %v", name, err)
		}
		if fixture.Case == "" || fixture.Status == "" {
			t.Fatalf("fixture %s must remain data-only evidence with case/status fields: %+v", name, fixture)
		}
	}
}
