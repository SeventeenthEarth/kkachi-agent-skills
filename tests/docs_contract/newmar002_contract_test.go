package docscontract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestNEWMAR002SchemasTemplatesAndGuidanceExist(t *testing.T) {
	requireContainsAll(t, "schemas/mar/request-bundle.schema.json", []string{
		"\"schema_version\": \"kas.mar.request_bundle.v1\"",
		"\"request_metadata\"",
		"\"role_matrix\"",
		"\"prompt_refs\"",
		"\"input_bundle_ref\"",
		"\"request_bundle_ref\"",
		"\"prompt_bundle_refs\"",
		"\"provider_family\"",
		"\"author_backend_correlation\"",
		"\"provider_preflight\"",
		"\"execution_policy_values\"",
		"\"approval_metadata\"",
		"\"provider_registry\"",
		"\"approval_binding\"",
		"\"retry_waiver_approval_ref\"",
		"\"policy_refs\"",
		"\"execution_safety\"",
		"\"additionalProperties\": false",
	})
	requireContainsAll(t, "schemas/mar/provider-registry-correlation.schema.json", []string{
		"\"schema_version\": \"kas.mar.provider_registry_correlation.v1\"",
		"\"lane_identity\"",
		"\"provider_identity\"",
		"\"adapter_type\"",
		"\"adapter_proof_ref\"",
		"\"provenance_ref\"",
		"\"author_backend_correlation\"",
		"\"provider_preflight_ref\"",
		"\"mar_start_approval_binding\"",
		"\"independence_policy\"",
		"\"execution_safety\"",
		"\"additionalProperties\": false",
	})

	requireContainsAll(t, "templates/run-artifacts/mar-request.yaml.tmpl", []string{
		"schema_version: kas.mar.request_bundle.v1",
		"task_id:",
		"run_id:",
		"request_metadata:",
		"role_matrix:",
		"prompt_refs:",
		"input_bundle_ref:",
		"provider_registry:",
		"request_bundle_ref:",
		"prompt_bundle_refs:",
		"provider_family:",
		"author_backend_correlation:",
		"provider_preflight:",
		"execution_policy_values:",
		"approval_metadata:",
		"approval_binding:",
		"retry_waiver_approval_ref:",
		"policy_refs:",
		"execution_safety:",
		"no inferred defaults",
		"no generated semantic prompt text",
	})
	requireContainsAll(t, "templates/run-artifacts/mar-provider-registry-correlation.yaml.tmpl", []string{
		"schema_version: kas.mar.provider_registry_correlation.v1",
		"lane_identity:",
		"provider_identity:",
		"adapter_type:",
		"adapter_proof_ref:",
		"provenance_ref:",
		"author_backend_correlation:",
		"provider_preflight_ref:",
		"mar_start_approval_binding:",
		"independence_policy:",
		"execution_safety:",
		"no inferred defaults",
		"no generated semantic prompt text",
	})

	requireContainsAll(t, "registries/phase-contracts.yaml", []string{
		"mar_request_bundle_contract:",
		"contract_version: \"newmar002.v1\"",
		"schemas/mar/request-bundle.schema.json",
		"schemas/mar/provider-registry-correlation.schema.json",
		"templates/run-artifacts/mar-request.yaml.tmpl",
		"templates/run-artifacts/mar-provider-registry-correlation.yaml.tmpl",
		"missing",
		"stale",
		"mismatched",
		"unsupported",
		"expired",
		"unsafe-ref",
		"extra-unreviewed",
		"checksum-drifting",
		"reviewed schema/capability evidence candidate only",
		"block unsafe provider execution before KAH consumes it",
	})

	requireContainsAll(t, "docs/sot/mar-execution-realignment.md", []string{
		"NEWMAR-002",
		"reviewed schema/template/fixture evidence",
		"NEWMAR-003+",
		"blocks unsafe provider execution before KAH consumes it",
		"may consume only reviewed NEWMAR-002 schema/capability evidence",
		"fallback/default provider execution is forbidden",
		"reviewed KAS compatibility subset",
	})
	requireContainsAll(t, "docs/roadmap.md", []string{
		"NEWMAR-002",
		"reviewed schema/template/fixture evidence",
		"NEWMAR-003+",
		"blocks unsafe provider execution before KAH consumes it",
		"may consume only reviewed NEWMAR-002 schema/capability evidence",
		"validation-only posture",
		"compatibility subset",
	})
	requireContainsAll(t, "skills/kkachi-multi-agent-review/SKILL.md", []string{
		"reviewed NEWMAR-002 request bundle and provider-registry/correlation evidence",
		"fail closed on missing, stale, mismatched, unsupported, expired, unsafe-ref, extra-unreviewed, or checksum-drifting metadata",
		"top-level author backend correlation is a summary view",
		"never fall back to default provider execution",
	})
}

func TestNEWMAR002PositiveFixturesValidate(t *testing.T) {
	request := mustReadJSONObject(t, "tests/fixtures/mar/newmar002/request-bundle-valid.json")
	if err := validateNEWMARRequestBundle(request); err != nil {
		t.Fatalf("valid request bundle rejected: %v", err)
	}

	correlation := mustReadJSONObject(t, "tests/fixtures/mar/newmar002/provider-registry-correlation-valid.json")
	if err := validateNEWMARProviderCorrelation(correlation); err != nil {
		t.Fatalf("valid provider correlation rejected: %v", err)
	}
}

func TestNEWMAR002ProviderCorrelationFixturesUseCanonicalGJCGoalBundleStatus(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join(repoRoot(t), "tests/fixtures/mar/newmar002/provider-registry-correlation-*.json"))
	if err != nil {
		t.Fatalf("glob provider correlation fixtures: %v", err)
	}
	sort.Strings(fixtures)
	if len(fixtures) == 0 {
		t.Fatal("expected provider correlation fixtures")
	}
	for _, path := range fixtures {
		rel, err := filepath.Rel(repoRoot(t), path)
		if err != nil {
			t.Fatalf("rel fixture path: %v", err)
		}
		doc := mustReadJSONObject(t, rel)
		authorBackendCorrelation, ok := getMap(doc, "author_backend_correlation")
		if !ok {
			t.Fatalf("%s: missing author_backend_correlation", rel)
		}
		if got := getString(authorBackendCorrelation, "source_candidate_status"); got != "implementation_goal_bundle_ready" {
			t.Fatalf("%s: source_candidate_status = %q, want implementation_goal_bundle_ready", rel, got)
		}
	}
}

func TestNEWMAR002NegativeFixturesFailClosed(t *testing.T) {
	requestCases := map[string]string{
		"tests/fixtures/mar/newmar002/request-bundle-missing-role-matrix.json":                         "missing",
		"tests/fixtures/mar/newmar002/request-bundle-stale-approval-binding.json":                      "stale",
		"tests/fixtures/mar/newmar002/request-bundle-stale-review-git-head.json":                       "stale",
		"tests/fixtures/mar/newmar002/request-bundle-stale-requested-at.json":                          "stale",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-requested-at.json":                        "missing",
		"tests/fixtures/mar/newmar002/request-bundle-mismatched-registry-checksum.json":                "mismatched",
		"tests/fixtures/mar/newmar002/request-bundle-checksum-drifting.json":                           "checksum-drifting",
		"tests/fixtures/mar/newmar002/request-bundle-checksum-drifting-provider-registry-sha.json":     "checksum-drifting",
		"tests/fixtures/mar/newmar002/request-bundle-unsafe-ref.json":                                  "unsafe-ref",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-nested-extra-request-ref.json":            "extra-unreviewed",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-nested-extra-policy-ref.json":             "extra-unreviewed",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-absolute-ref-path.json":                   "unsafe-ref",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-parent-traversal-ref-path.json":           "unsafe-ref",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-drive-letter-ref-path.json":               "unsafe-ref",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-trailing-parent-ref-path.json":            "unsafe-ref",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-bad-checksum.json":                        "missing",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-missing-mar-start-bound-tuple.json":       "missing",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-missing-provider-family-bound-tuple.json": "missing",
		"tests/fixtures/mar/newmar002/request-bundle-invalid-semantic-prompt-text-flag.json":           "missing",
		"tests/fixtures/mar/newmar002/request-bundle-unsupported-schema-version.json":                  "unsupported",
		"tests/fixtures/mar/newmar002/request-bundle-unsupported-approval-posture.json":                "unsupported",
		"tests/fixtures/mar/newmar002/request-bundle-unsupported-execution-safety-policy.json":         "unsupported",
	}
	for rel, wantCode := range requestCases {
		t.Run(filepath.Base(rel), func(t *testing.T) {
			err := validateNEWMARRequestBundle(mustReadJSONObject(t, rel))
			if err == nil || !strings.Contains(err.Error(), wantCode) {
				t.Fatalf("%s error = %v, want code %q", rel, err, wantCode)
			}
		})
	}

	correlationCases := map[string]string{
		"tests/fixtures/mar/newmar002/provider-registry-correlation-unsupported-adapter.json":                          "unsupported",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-expired-preflight.json":                            "expired",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-extra-unreviewed.json":                             "extra-unreviewed",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-invalid-nested-extra-provider-ref.json":            "extra-unreviewed",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-invalid-nested-extra-provider-preflight-ref.json":  "extra-unreviewed",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-invalid-bad-checksum.json":                         "missing",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-invalid-missing-preflight-capability-version.json": "missing",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-stale-lane-provider.json":                          "stale",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-mismatched-provider-registry-sha.json":             "mismatched",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-invalid-semantic-prompt-generation-flag.json":      "missing",
		"tests/fixtures/mar/newmar002/provider-registry-correlation-mismatched-author-backend-family.json":             "mismatched",
	}
	for rel, wantCode := range correlationCases {
		t.Run(filepath.Base(rel), func(t *testing.T) {
			err := validateNEWMARProviderCorrelation(mustReadJSONObject(t, rel))
			if err == nil || !strings.Contains(err.Error(), wantCode) {
				t.Fatalf("%s error = %v, want code %q", rel, err, wantCode)
			}
		})
	}
}

func TestNEWMAR002JSONSchemaBackedFailClosedRefContracts(t *testing.T) {
	requestSchema := mustReadJSONObject(t, "schemas/mar/request-bundle.schema.json")
	correlationSchema := mustReadJSONObject(t, "schemas/mar/provider-registry-correlation.schema.json")

	requireStrictRefDef(t, requestSchema)
	requireStrictRefDef(t, correlationSchema)
	requireStrictProviderPreflightRefDef(t, requestSchema)
	requireStrictProviderPreflightRefDef(t, correlationSchema)
	requireBoundTupleDef(t, requestSchema)
	requireBoundTupleDef(t, correlationSchema)
}

func mustReadJSONObject(t *testing.T, rel string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("parse %s: %v", rel, err)
	}
	return value
}

func validateNEWMARRequestBundle(doc map[string]any) error {
	allowed := []string{
		"schema_version",
		"task_id",
		"run_id",
		"request_metadata",
		"role_matrix",
		"prompt_refs",
		"input_bundle_ref",
		"provider_registry",
		"approval_binding",
		"retry_waiver_approval_ref",
		"policy_refs",
		"execution_safety",
	}
	if extra := extraKeys(doc, allowed); len(extra) > 0 {
		return fmt.Errorf("extra-unreviewed: unexpected top-level keys %v", extra)
	}
	if getString(doc, "schema_version") != "kas.mar.request_bundle.v1" {
		return fmt.Errorf("unsupported: schema_version must be kas.mar.request_bundle.v1")
	}
	if getString(doc, "task_id") == "" || getString(doc, "run_id") == "" {
		return fmt.Errorf("missing: task_id/run_id required")
	}

	requestMetadata, ok := getMap(doc, "request_metadata")
	if !ok {
		return fmt.Errorf("missing: request_metadata required")
	}
	if getString(requestMetadata, "request_id") == "" || getString(requestMetadata, "requested_at") == "" || !getBool(requestMetadata, "candidate_evidence_only") {
		return fmt.Errorf("missing: request_metadata fields required")
	}
	requestedAt, err := parseTimestamp(getString(requestMetadata, "requested_at"))
	if err != nil {
		return fmt.Errorf("missing: invalid requested_at: %w", err)
	}
	reviewSurface, ok := getMap(requestMetadata, "review_surface")
	if !ok {
		return fmt.Errorf("missing: review_surface required")
	}
	reviewSurfaceRef, ok := getMap(reviewSurface, "ref")
	if !ok {
		return fmt.Errorf("missing: review_surface.ref required")
	}
	if err := validateRef(reviewSurfaceRef); err != nil {
		return err
	}
	reviewGitHead := getString(reviewSurface, "git_head")
	if reviewGitHead == "" {
		return fmt.Errorf("missing: review_surface.git_head required")
	}

	roleMatrix, ok := getMap(doc, "role_matrix")
	if !ok {
		return fmt.Errorf("missing: role_matrix required")
	}
	requiredRoles, err := getStringSlice(roleMatrix, "required_roles")
	if err != nil || len(requiredRoles) == 0 {
		return fmt.Errorf("missing: role_matrix.required_roles required")
	}
	bindings, ok := roleMatrix["role_bindings"].([]any)
	if !ok || len(bindings) == 0 {
		return fmt.Errorf("missing: role_matrix.role_bindings required")
	}
	boundRoles := map[string]bool{}
	for _, raw := range bindings {
		binding, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("missing: invalid role binding")
		}
		roleID := getString(binding, "role_id")
		if roleID == "" || getString(binding, "lane_id") == "" || getString(binding, "prompt_ref_id") == "" || getString(binding, "provider_registry_entry_id") == "" {
			return fmt.Errorf("missing: role binding fields required")
		}
		boundRoles[roleID] = true
	}
	for _, roleID := range requiredRoles {
		if !boundRoles[roleID] {
			return fmt.Errorf("missing: role binding for %s", roleID)
		}
	}

	promptRefs, ok := doc["prompt_refs"].([]any)
	if !ok || len(promptRefs) == 0 {
		return fmt.Errorf("missing: prompt_refs required")
	}
	promptIDs := map[string]bool{}
	for _, raw := range promptRefs {
		promptRef, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("missing: invalid prompt ref")
		}
		if getString(promptRef, "id") == "" {
			return fmt.Errorf("missing: prompt ref id required")
		}
		if err := validatePromptRef(promptRef); err != nil {
			return err
		}
		promptIDs[getString(promptRef, "id")] = true
	}
	for _, raw := range bindings {
		binding := raw.(map[string]any)
		if !promptIDs[getString(binding, "prompt_ref_id")] {
			return fmt.Errorf("missing: prompt_ref_id %q not declared", getString(binding, "prompt_ref_id"))
		}
	}

	inputBundleRef, ok := getMap(doc, "input_bundle_ref")
	if !ok {
		return fmt.Errorf("missing: input_bundle_ref required")
	}
	if err := validateRef(inputBundleRef); err != nil {
		return err
	}

	providerRegistry, ok := getMap(doc, "provider_registry")
	if !ok {
		return fmt.Errorf("missing: provider_registry required")
	}
	if getString(providerRegistry, "schema_version") != "kas.mar.provider_registry_correlation.v1" {
		return fmt.Errorf("unsupported: provider_registry.schema_version mismatch")
	}
	registryRef, ok := getMap(providerRegistry, "ref")
	if !ok {
		return fmt.Errorf("missing: provider_registry.ref required")
	}
	if err := validateRef(registryRef); err != nil {
		return err
	}
	correlationRef, ok := getMap(providerRegistry, "correlation_ref")
	if !ok {
		return fmt.Errorf("missing: provider_registry.correlation_ref required")
	}
	if err := validateRef(correlationRef); err != nil {
		return err
	}
	if snapshot := getString(providerRegistry, "snapshot_sha256"); snapshot == "" {
		return fmt.Errorf("missing: provider_registry.snapshot_sha256 required")
	} else if snapshot != getString(registryRef, "sha256") {
		return fmt.Errorf("mismatched: provider_registry.snapshot_sha256 must match provider_registry.ref.sha256")
	}

	approvalBinding, ok := getMap(doc, "approval_binding")
	if !ok {
		return fmt.Errorf("missing: approval_binding required")
	}
	marStart, ok := getMap(approvalBinding, "mar_start")
	if !ok {
		return fmt.Errorf("missing: approval_binding.mar_start required")
	}
	if getString(marStart, "approval_ref") == "" {
		return fmt.Errorf("missing: approval_binding.mar_start.approval_ref required")
	}
	if getString(marStart, "task_id") != getString(doc, "task_id") || getString(marStart, "run_id") != getString(doc, "run_id") {
		return fmt.Errorf("stale: mar_start task/run binding mismatch")
	}
	if getString(marStart, "review_git_head") != reviewGitHead {
		return fmt.Errorf("stale: mar_start review_git_head mismatch")
	}
	if getString(marStart, "input_bundle_sha256") != getString(inputBundleRef, "sha256") {
		return fmt.Errorf("checksum-drifting: input bundle checksum mismatch")
	}
	if getString(marStart, "provider_registry_sha256") != getString(registryRef, "sha256") {
		return fmt.Errorf("checksum-drifting: provider registry checksum mismatch")
	}
	marStartRequestedAt, err := parseTimestamp(getString(marStart, "requested_at"))
	if err != nil {
		return fmt.Errorf("missing: invalid mar_start.requested_at: %w", err)
	}
	if !requestedAt.Equal(marStartRequestedAt) {
		return fmt.Errorf("stale: mar_start requested_at mismatch")
	}
	boundTuple, ok := getMap(marStart, "bound_tuple")
	if !ok {
		return fmt.Errorf("missing: approval_binding.mar_start.bound_tuple required")
	}
	if err := validateBoundTuple(boundTuple, requestedAt); err != nil {
		return err
	}

	if retryWaiverRef, ok := getMap(doc, "retry_waiver_approval_ref"); !ok {
		return fmt.Errorf("missing: retry_waiver_approval_ref required")
	} else if err := validateRef(retryWaiverRef); err != nil {
		return err
	}

	policyRefs, ok := getMap(doc, "policy_refs")
	if !ok {
		return fmt.Errorf("missing: policy_refs required")
	}
	for _, key := range []string{"concurrency_policy_ref", "duration_policy_ref", "cost_policy_ref", "independence_policy_ref", "execution_safety_ref"} {
		ref, ok := getMap(policyRefs, key)
		if !ok {
			return fmt.Errorf("missing: policy_refs.%s required", key)
		}
		if err := validateRef(ref); err != nil {
			return err
		}
	}

	executionSafety, ok := getMap(doc, "execution_safety")
	if !ok {
		return fmt.Errorf("missing: execution_safety required")
	}
	if !getBool(executionSafety, "fail_closed_on_stale_or_mismatch") || !getBool(executionSafety, "block_provider_execution_until_kah_capability_match") || !getBool(executionSafety, "no_generated_semantic_prompt_text") {
		return fmt.Errorf("missing: execution_safety booleans required")
	}
	if getString(executionSafety, "unsafe_ref_policy") != "reject" || getString(executionSafety, "extra_unreviewed_metadata_policy") != "reject" {
		return fmt.Errorf("unsupported: execution_safety policies must be reject")
	}

	return nil
}

func validateNEWMARProviderCorrelation(doc map[string]any) error {
	allowed := []string{
		"schema_version",
		"lane_identity",
		"provider_identity",
		"adapter_type",
		"provider_registry_ref",
		"provider_registry_schema_version",
		"adapter_proof_ref",
		"provenance_ref",
		"author_backend_correlation",
		"provider_preflight_ref",
		"retry_waiver_approval_ref",
		"mar_start_approval_binding",
		"policy_refs",
		"independence_policy",
		"execution_safety",
	}
	if extra := extraKeys(doc, allowed); len(extra) > 0 {
		return fmt.Errorf("extra-unreviewed: unexpected top-level keys %v", extra)
	}
	if getString(doc, "schema_version") != "kas.mar.provider_registry_correlation.v1" {
		return fmt.Errorf("unsupported: schema_version must be kas.mar.provider_registry_correlation.v1")
	}

	laneIdentity, ok := getMap(doc, "lane_identity")
	if !ok || getString(laneIdentity, "lane_id") == "" || getString(laneIdentity, "role_id") == "" || getString(laneIdentity, "lane_kind") == "" {
		return fmt.Errorf("missing: lane_identity fields required")
	}
	providerIdentity, ok := getMap(doc, "provider_identity")
	if !ok || getString(providerIdentity, "provider_id") == "" || getString(providerIdentity, "provider_family") == "" {
		return fmt.Errorf("missing: provider_identity fields required")
	}

	adapterType := getString(doc, "adapter_type")
	if adapterType != "kah_mar" && adapterType != "fixture_only" {
		return fmt.Errorf("unsupported: adapter_type %q", adapterType)
	}

	providerRegistryRef, ok := getMap(doc, "provider_registry_ref")
	if !ok {
		return fmt.Errorf("missing: provider_registry_ref required")
	}
	if err := validateRef(providerRegistryRef); err != nil {
		return err
	}
	if getString(doc, "provider_registry_schema_version") != "mar.role_lanes.v1" {
		return fmt.Errorf("unsupported: provider_registry_schema_version must be mar.role_lanes.v1")
	}
	for _, key := range []string{"adapter_proof_ref", "provenance_ref", "retry_waiver_approval_ref"} {
		ref, ok := getMap(doc, key)
		if !ok {
			return fmt.Errorf("missing: %s required", key)
		}
		if err := validateRef(ref); err != nil {
			return err
		}
	}

	authorBackendCorrelation, ok := getMap(doc, "author_backend_correlation")
	if !ok {
		return fmt.Errorf("missing: author_backend_correlation required")
	}
	if getString(authorBackendCorrelation, "backend_family") == "" || getString(authorBackendCorrelation, "backend_identity") == "" || getString(authorBackendCorrelation, "correlation_id") == "" || getString(authorBackendCorrelation, "source_candidate_status") == "" {
		return fmt.Errorf("missing: author_backend_correlation fields required")
	}
	if getString(authorBackendCorrelation, "source_packet_sha256") != getString(doc["provenance_ref"].(map[string]any), "sha256") {
		return fmt.Errorf("checksum-drifting: provenance/source packet checksum mismatch")
	}

	providerPreflightRef, ok := getMap(doc, "provider_preflight_ref")
	if !ok {
		return fmt.Errorf("missing: provider_preflight_ref required")
	}
	if err := validateProviderPreflightRef(providerPreflightRef); err != nil {
		return err
	}
	validThrough, err := parseTimestamp(getString(providerPreflightRef, "valid_through_request_ts"))
	if err != nil {
		return fmt.Errorf("missing: invalid provider_preflight_ref.valid_through_request_ts: %w", err)
	}

	marStart, ok := getMap(doc, "mar_start_approval_binding")
	if !ok {
		return fmt.Errorf("missing: mar_start_approval_binding required")
	}
	requestedAt, err := parseTimestamp(getString(marStart, "requested_at"))
	if err != nil {
		return fmt.Errorf("missing: invalid mar_start_approval_binding.requested_at: %w", err)
	}
	if validThrough.Before(requestedAt) {
		return fmt.Errorf("expired: provider preflight expired before request")
	}
	if getString(marStart, "lane_id") != getString(laneIdentity, "lane_id") || getString(marStart, "provider_id") != getString(providerIdentity, "provider_id") {
		return fmt.Errorf("stale: mar_start lane/provider mismatch")
	}
	if getString(marStart, "provider_registry_sha256") != getString(providerRegistryRef, "sha256") {
		return fmt.Errorf("mismatched: mar_start provider_registry_sha256 mismatch")
	}
	boundTuple, ok := getMap(marStart, "bound_tuple")
	if !ok {
		return fmt.Errorf("missing: mar_start_approval_binding.bound_tuple required")
	}
	if err := validateBoundTuple(boundTuple, requestedAt); err != nil {
		return err
	}
	boundAuthorBackendCorrelation, _ := getMap(boundTuple, "author_backend_correlation")
	if getString(authorBackendCorrelation, "backend_family") != getString(boundAuthorBackendCorrelation, "author_backend_family") || getString(authorBackendCorrelation, "backend_identity") != getString(boundAuthorBackendCorrelation, "author_backend_id") {
		return fmt.Errorf("mismatched: author_backend_correlation summary/bound tuple mismatch")
	}

	policyRefs, ok := getMap(doc, "policy_refs")
	if !ok {
		return fmt.Errorf("missing: policy_refs required")
	}
	for _, key := range []string{"concurrency_policy_ref", "duration_policy_ref", "cost_policy_ref", "independence_policy_ref", "execution_safety_ref"} {
		ref, ok := getMap(policyRefs, key)
		if !ok {
			return fmt.Errorf("missing: policy_refs.%s required", key)
		}
		if err := validateRef(ref); err != nil {
			return err
		}
	}

	independencePolicy, ok := getMap(doc, "independence_policy")
	if !ok || !getBool(independencePolicy, "reviewer_independence_required") || !getBool(independencePolicy, "correlation_review_required") {
		return fmt.Errorf("missing: independence_policy booleans required")
	}
	if getBool(independencePolicy, "shared_author_backend_allowed") {
		return fmt.Errorf("unsupported: shared_author_backend_allowed must be false")
	}

	executionSafety, ok := getMap(doc, "execution_safety")
	if !ok {
		return fmt.Errorf("missing: execution_safety required")
	}
	if !getBool(executionSafety, "candidate_evidence_only") || !getBool(executionSafety, "block_unsafe_provider_execution") || !getBool(executionSafety, "no_semantic_prompt_generation") {
		return fmt.Errorf("missing: execution_safety booleans required")
	}
	if getString(executionSafety, "safe_ref_policy") != "repo_relative_only" {
		return fmt.Errorf("unsupported: safe_ref_policy must be repo_relative_only")
	}

	return nil
}

func extraKeys(doc map[string]any, allowed []string) []string {
	allowedSet := map[string]bool{}
	for _, key := range allowed {
		allowedSet[key] = true
	}
	var extras []string
	for key := range doc {
		if !allowedSet[key] {
			extras = append(extras, key)
		}
	}
	sort.Strings(extras)
	return extras
}

func validateRef(ref map[string]any) error {
	if extra := extraKeys(ref, []string{"path", "sha256"}); len(extra) > 0 {
		return fmt.Errorf("extra-unreviewed: unexpected ref keys %v", extra)
	}
	return validateRefShape(ref)
}

func validatePromptRef(ref map[string]any) error {
	if extra := extraKeys(ref, []string{"id", "path", "sha256"}); len(extra) > 0 {
		return fmt.Errorf("extra-unreviewed: unexpected prompt ref keys %v", extra)
	}
	return validateRefShape(ref)
}

func validateProviderPreflightRef(ref map[string]any) error {
	if extra := extraKeys(ref, []string{"path", "sha256", "valid_through_request_ts"}); len(extra) > 0 {
		return fmt.Errorf("extra-unreviewed: unexpected provider preflight ref keys %v", extra)
	}
	return validateRefShape(ref)
}

func validateRefShape(ref map[string]any) error {
	pathValue := getString(ref, "path")
	checksum := getString(ref, "sha256")
	if pathValue == "" || checksum == "" {
		return fmt.Errorf("missing: path and sha256 required")
	}
	normalizedPath := strings.ReplaceAll(pathValue, `\`, "/")
	if filepath.IsAbs(pathValue) || driveLetterPattern.MatchString(normalizedPath) {
		return fmt.Errorf("unsafe-ref: %s", pathValue)
	}
	for _, segment := range strings.Split(normalizedPath, "/") {
		if segment == ".." {
			return fmt.Errorf("unsafe-ref: %s", pathValue)
		}
	}
	if !sha256Pattern.MatchString(checksum) {
		return fmt.Errorf("missing: invalid sha256 %q", checksum)
	}
	return nil
}

func validateBoundTuple(tuple map[string]any, requestedAt time.Time) error {
	allowed := []string{"task_id", "run_id", "request_bundle_schema_version", "request_bundle_ref", "request_schema_ref", "role_matrix_ref", "prompt_refs_sha256", "prompt_bundle_refs", "input_bundle_ref", "required_roles", "provider_lane_id", "provider_id", "provider_family", "adapter_type", "adapter_proof_ref", "provider_registry_schema_version", "provider_registry_ref", "provider_correlation_ref", "provider_provenance_ref", "provider_preflight_ref", "provider_preflight", "required_capabilities", "author_backend_correlation", "execution_policy_refs", "execution_policy_values", "approval_metadata_ref", "approval_metadata", "approved_tuple_checksum", "approval_expires_at", "retry_waiver_approval_ref"}
	if extra := extraKeys(tuple, allowed); len(extra) > 0 {
		return fmt.Errorf("extra-unreviewed: unexpected bound tuple keys %v", extra)
	}
	for _, key := range []string{"request_bundle_ref", "request_schema_ref", "role_matrix_ref", "input_bundle_ref", "adapter_proof_ref", "provider_registry_ref", "provider_correlation_ref", "provider_provenance_ref", "approval_metadata_ref", "retry_waiver_approval_ref"} {
		ref, ok := getMap(tuple, key)
		if !ok {
			return fmt.Errorf("missing: bound_tuple.%s required", key)
		}
		if err := validateRef(ref); err != nil {
			return err
		}
	}
	providerPreflightRef, ok := getMap(tuple, "provider_preflight_ref")
	if !ok {
		return fmt.Errorf("missing: bound_tuple.provider_preflight_ref required")
	}
	if err := validateProviderPreflightRef(providerPreflightRef); err != nil {
		return err
	}
	providerPreflightValidThrough, err := parseTimestamp(getString(providerPreflightRef, "valid_through_request_ts"))
	if err != nil {
		return fmt.Errorf("missing: invalid bound_tuple.provider_preflight_ref.valid_through_request_ts: %w", err)
	}
	if providerPreflightValidThrough.Before(requestedAt) {
		return fmt.Errorf("expired: bound_tuple provider preflight expired before request")
	}
	promptBundleRefs, ok := tuple["prompt_bundle_refs"].([]any)
	if !ok || len(promptBundleRefs) == 0 {
		return fmt.Errorf("missing: bound_tuple.prompt_bundle_refs required")
	}
	for _, raw := range promptBundleRefs {
		promptRef, ok := raw.(map[string]any)
		if !ok || getString(promptRef, "id") == "" {
			return fmt.Errorf("missing: bound_tuple.prompt_bundle_refs entries required")
		}
		if err := validatePromptRef(promptRef); err != nil {
			return err
		}
	}
	if _, err := getStringSlice(tuple, "required_roles"); err != nil {
		return fmt.Errorf("missing: bound_tuple.required_roles required")
	}
	if _, err := getStringSlice(tuple, "required_capabilities"); err != nil {
		return fmt.Errorf("missing: bound_tuple.required_capabilities required")
	}
	for _, key := range []string{"prompt_refs_sha256", "approved_tuple_checksum"} {
		if !sha256Pattern.MatchString(getString(tuple, key)) {
			return fmt.Errorf("missing: bound_tuple.%s must be sha256", key)
		}
	}
	for _, key := range []string{"task_id", "run_id", "request_bundle_schema_version", "provider_lane_id", "provider_id", "provider_family", "adapter_type", "provider_registry_schema_version"} {
		if getString(tuple, key) == "" {
			return fmt.Errorf("missing: bound_tuple.%s required", key)
		}
	}
	if adapterType := getString(tuple, "adapter_type"); adapterType != "kah_mar" && adapterType != "fixture_only" {
		return fmt.Errorf("unsupported: bound_tuple.adapter_type %q", adapterType)
	}
	if getString(tuple, "provider_registry_schema_version") != "mar.role_lanes.v1" {
		return fmt.Errorf("unsupported: bound_tuple.provider_registry_schema_version must be mar.role_lanes.v1")
	}
	authorCorrelation, ok := getMap(tuple, "author_backend_correlation")
	if !ok || getString(authorCorrelation, "author_backend_id") == "" || getString(authorCorrelation, "author_backend_family") == "" || getString(authorCorrelation, "correlation_status") == "" {
		return fmt.Errorf("missing: bound_tuple.author_backend_correlation required")
	}
	if ref, ok := getMap(authorCorrelation, "independence_policy_ref"); !ok {
		return fmt.Errorf("missing: bound_tuple.author_backend_correlation.independence_policy_ref required")
	} else if err := validateRef(ref); err != nil {
		return err
	}
	for _, key := range []string{"author_session_ref", "author_thread_ref", "author_turn_ref"} {
		if raw, exists := authorCorrelation[key]; exists && raw != nil {
			ref, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("missing: bound_tuple.author_backend_correlation.%s must be ref or null", key)
			}
			if err := validateRef(ref); err != nil {
				return err
			}
		}
	}
	providerPreflight, ok := getMap(tuple, "provider_preflight")
	if !ok {
		return fmt.Errorf("missing: bound_tuple.provider_preflight required")
	}
	preflightRef, ok := getMap(providerPreflight, "ref")
	if !ok {
		return fmt.Errorf("missing: bound_tuple.provider_preflight.ref required")
	}
	if err := validateRef(preflightRef); err != nil {
		return err
	}
	if getString(preflightRef, "path") != getString(providerPreflightRef, "path") || getString(preflightRef, "sha256") != getString(providerPreflightRef, "sha256") {
		return fmt.Errorf("mismatched: bound_tuple.provider_preflight ref mismatch")
	}
	for _, key := range []string{"generated_at", "expires_at"} {
		if _, err := parseTimestamp(getString(providerPreflight, key)); err != nil {
			return fmt.Errorf("missing: invalid bound_tuple.provider_preflight.%s: %w", key, err)
		}
	}
	if getString(providerPreflight, "expires_at") != getString(providerPreflightRef, "valid_through_request_ts") {
		return fmt.Errorf("mismatched: bound_tuple.provider_preflight expiry mismatch")
	}
	for _, key := range []string{"checked_binary_version", "checked_capability_version"} {
		if getString(providerPreflight, key) == "" {
			return fmt.Errorf("missing: bound_tuple.provider_preflight.%s required", key)
		}
	}
	if _, err := parseTimestamp(getString(tuple, "approval_expires_at")); err != nil {
		return fmt.Errorf("missing: invalid bound_tuple.approval_expires_at: %w", err)
	}
	policyRefs, ok := getMap(tuple, "execution_policy_refs")
	if !ok {
		return fmt.Errorf("missing: bound_tuple.execution_policy_refs required")
	}
	for _, key := range []string{"concurrency_policy_ref", "duration_policy_ref", "cost_policy_ref", "independence_policy_ref", "execution_safety_ref"} {
		ref, ok := getMap(policyRefs, key)
		if !ok {
			return fmt.Errorf("missing: bound_tuple.execution_policy_refs.%s required", key)
		}
		if err := validateRef(ref); err != nil {
			return err
		}
	}
	executionPolicy, ok := getMap(tuple, "execution_policy_values")
	if !ok {
		return fmt.Errorf("missing: bound_tuple.execution_policy_values required")
	}
	if concurrency, ok := executionPolicy["max_concurrency"].(float64); !ok || concurrency < 1 {
		return fmt.Errorf("missing: bound_tuple.execution_policy_values.max_concurrency required")
	}
	for _, key := range []string{"max_duration", "max_cost_or_token_budget", "raw_output_cap"} {
		if getString(executionPolicy, key) == "" {
			return fmt.Errorf("missing: bound_tuple.execution_policy_values.%s required", key)
		}
	}
	if !getBool(executionPolicy, "secret_scan_redaction_required") || !getBool(executionPolicy, "mutation_guard_required") {
		return fmt.Errorf("missing: bound_tuple.execution_policy_values safety booleans required")
	}
	if safetyPolicyRef, ok := getMap(executionPolicy, "safety_policy_ref"); !ok {
		return fmt.Errorf("missing: bound_tuple.execution_policy_values.safety_policy_ref required")
	} else if err := validateRef(safetyPolicyRef); err != nil {
		return err
	}
	approvalMetadata, ok := getMap(tuple, "approval_metadata")
	if !ok {
		return fmt.Errorf("missing: bound_tuple.approval_metadata required")
	}
	for _, key := range []string{"approval_posture", "approver_id", "approval_ref_id", "approval_scope", "null_validation_posture"} {
		if getString(approvalMetadata, key) == "" {
			return fmt.Errorf("missing: bound_tuple.approval_metadata.%s required", key)
		}
	}
	if posture := getString(approvalMetadata, "approval_posture"); posture != "approved" && posture != "validation_only" {
		return fmt.Errorf("unsupported: bound_tuple.approval_metadata.approval_posture %q", posture)
	}
	if !sha256Pattern.MatchString(getString(approvalMetadata, "approved_tuple_checksum")) {
		return fmt.Errorf("missing: bound_tuple.approval_metadata.approved_tuple_checksum must be sha256")
	}
	for _, key := range []string{"approval_created_at", "approval_expires_at"} {
		if _, err := parseTimestamp(getString(approvalMetadata, key)); err != nil {
			return fmt.Errorf("missing: invalid bound_tuple.approval_metadata.%s: %w", key, err)
		}
	}
	if retryWaiverRef, ok := getMap(approvalMetadata, "retry_waiver_approval_ref"); !ok {
		return fmt.Errorf("missing: bound_tuple.approval_metadata.retry_waiver_approval_ref required")
	} else if err := validateRef(retryWaiverRef); err != nil {
		return err
	}
	return nil
}

var sha256Pattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var driveLetterPattern = regexp.MustCompile(`^[A-Za-z]:`)

func parseTimestamp(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	return time.Parse(time.RFC3339, raw)
}

func getMap(doc map[string]any, key string) (map[string]any, bool) {
	value, ok := doc[key].(map[string]any)
	return value, ok
}

func getString(doc map[string]any, key string) string {
	value, _ := doc[key].(string)
	return value
}

func getBool(doc map[string]any, key string) bool {
	value, _ := doc[key].(bool)
	return value
}

func getStringSlice(doc map[string]any, key string) ([]string, error) {
	raw, ok := doc[key].([]any)
	if !ok {
		return nil, fmt.Errorf("not a string slice")
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		value, ok := item.(string)
		if !ok || value == "" {
			return nil, fmt.Errorf("invalid string slice item")
		}
		out = append(out, value)
	}
	return out, nil
}

func requireStrictRefDef(t *testing.T, schema map[string]any) {
	t.Helper()
	refDef := schemaDef(t, schema, "ref")
	if !getBool(refDef, "additionalProperties") && refDef["additionalProperties"] != false {
		t.Fatalf("$defs.ref.additionalProperties must be false")
	}
	required, err := getStringSlice(refDef, "required")
	if err != nil || !containsAllStrings(required, []string{"path", "sha256"}) {
		t.Fatalf("$defs.ref.required = %v, want path+sha256", required)
	}
	properties := mustMap(t, refDef, "properties")
	pathProp := mustMap(t, properties, "path")
	if pathProp["$ref"] != "#/$defs/safe_relative_path" {
		t.Fatalf("$defs.ref.path must use safe_relative_path schema, got %v", pathProp)
	}
	shaProp := mustMap(t, properties, "sha256")
	if shaProp["$ref"] != "#/$defs/sha256" {
		t.Fatalf("$defs.ref.sha256 must use sha256 schema, got %v", shaProp)
	}
	sha := schemaDef(t, schema, "sha256")
	if getString(sha, "pattern") != `^sha256:[0-9a-f]{64}$` {
		t.Fatalf("$defs.sha256.pattern = %q", getString(sha, "pattern"))
	}
	path := schemaDef(t, schema, "safe_relative_path")
	pattern := getString(path, "pattern")
	if !strings.Contains(pattern, "?!/") || !strings.Contains(pattern, "\\.\\.") {
		t.Fatalf("$defs.safe_relative_path.pattern does not reject absolute and parent traversal paths: %q", pattern)
	}
}

func requireStrictProviderPreflightRefDef(t *testing.T, schema map[string]any) {
	t.Helper()
	preflightDef := schemaDef(t, schema, "provider_preflight_ref")
	if preflightDef["additionalProperties"] != false {
		t.Fatalf("$defs.provider_preflight_ref.additionalProperties must be false")
	}
	required, err := getStringSlice(preflightDef, "required")
	if err != nil || !containsAllStrings(required, []string{"path", "sha256", "valid_through_request_ts"}) {
		t.Fatalf("provider_preflight_ref.required = %v", required)
	}
}

func requireBoundTupleDef(t *testing.T, schema map[string]any) {
	t.Helper()
	bound := schemaDef(t, schema, "mar_start_bound_tuple")
	if bound["additionalProperties"] != false {
		t.Fatalf("$defs.mar_start_bound_tuple.additionalProperties must be false")
	}
	required, err := getStringSlice(bound, "required")
	if err != nil {
		t.Fatalf("mar_start_bound_tuple.required invalid: %v", err)
	}
	want := []string{"task_id", "run_id", "request_bundle_schema_version", "request_bundle_ref", "request_schema_ref", "role_matrix_ref", "prompt_refs_sha256", "prompt_bundle_refs", "input_bundle_ref", "required_roles", "provider_lane_id", "provider_id", "provider_family", "adapter_type", "adapter_proof_ref", "provider_registry_schema_version", "provider_registry_ref", "provider_provenance_ref", "provider_preflight_ref", "provider_preflight", "required_capabilities", "author_backend_correlation", "execution_policy_refs", "execution_policy_values", "approval_metadata_ref", "approval_metadata", "approved_tuple_checksum", "approval_expires_at", "retry_waiver_approval_ref"}
	if !containsAllStrings(required, want) {
		t.Fatalf("mar_start_bound_tuple.required = %v, missing one of %v", required, want)
	}
	properties := mustMap(t, bound, "properties")
	adapterType := mustMap(t, properties, "adapter_type")
	adapterEnum, err := getStringSlice(adapterType, "enum")
	if err != nil || !containsAllStrings(adapterEnum, []string{"kah_mar", "fixture_only"}) {
		t.Fatalf("mar_start_bound_tuple.adapter_type.enum = %v", adapterEnum)
	}
	providerPreflightRef := mustMap(t, properties, "provider_preflight_ref")
	if providerPreflightRef["$ref"] != "#/$defs/provider_preflight_ref" {
		t.Fatalf("mar_start_bound_tuple.provider_preflight_ref must use strict provider_preflight_ref schema, got %v", providerPreflightRef)
	}
	providerRegistrySchemaVersion := mustMap(t, properties, "provider_registry_schema_version")
	if getString(providerRegistrySchemaVersion, "const") != "mar.role_lanes.v1" {
		t.Fatalf("mar_start_bound_tuple.provider_registry_schema_version.const = %q", getString(providerRegistrySchemaVersion, "const"))
	}
}

func schemaDef(t *testing.T, schema map[string]any, name string) map[string]any {
	t.Helper()
	defs := mustMap(t, schema, "$defs")
	return mustMap(t, defs, name)
}

func mustMap(t *testing.T, doc map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := doc[key].(map[string]any)
	if !ok {
		t.Fatalf("%s missing or not object", key)
	}
	return value
}

func containsAllStrings(have []string, want []string) bool {
	set := map[string]bool{}
	for _, item := range have {
		set[item] = true
	}
	for _, item := range want {
		if !set[item] {
			return false
		}
	}
	return true
}
