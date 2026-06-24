package migrationclassifier

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/discovery"
	"github.com/SeventeenthEarth/kkachi-agent-skills/internal/skills/version"
)

const Mode = "profile_skill_migration_classifier"

type Options struct {
	Profile     string
	Project     string
	ProfileRoot string
}

type NoWriteEvidence struct {
	Guaranteed                   bool   `json:"guaranteed"`
	ProfileTreeHashBefore        string `json:"profile_tree_hash_before,omitempty"`
	ProfileTreeHashAfter         string `json:"profile_tree_hash_after,omitempty"`
	ProfileTreeUnchanged         bool   `json:"profile_tree_unchanged"`
	ProfileSkillWriteCount       int    `json:"profile_skill_write_count"`
	ProfileSkillDeleteCount      int    `json:"profile_skill_delete_count"`
	ProfileSkillMigrationCount   int    `json:"profile_skill_migration_count"`
	KAHStateWriteCount           int    `json:"kah_state_write_count"`
	KABRuntimeMutationCount      int    `json:"kab_runtime_mutation_count"`
	HermesRuntimeMutationCount   int    `json:"hermes_runtime_mutation_count"`
	AuthProviderConfigWriteCount int    `json:"auth_provider_config_write_count"`
}

type NoSpilloverEvidence struct {
	Guaranteed                bool     `json:"guaranteed"`
	ProfileRoot               string   `json:"profile_root,omitempty"`
	SkillsRoot                string   `json:"skills_root,omitempty"`
	InspectedSkillCount       int      `json:"inspected_skill_count"`
	UnrelatedSkillCount       int      `json:"unrelated_skill_count"`
	UnrelatedSkillsPreserved  []string `json:"unrelated_skills_preserved"`
	ExternalProjectWriteCount int      `json:"external_project_write_count"`
	UnrequestedProfileTouched bool     `json:"unrequested_profile_touched"`
	ProfileTreeHashBefore     string   `json:"profile_tree_hash_before,omitempty"`
	ProfileTreeHashAfter      string   `json:"profile_tree_hash_after,omitempty"`
	ProfileTreeHashMatch      bool     `json:"profile_tree_hash_match"`
}

type HashEvidence struct {
	ProfileSkillSHA256 string `json:"profile_skill_sha256,omitempty"`
	SourceBaseSHA256   string `json:"source_base_sha256,omitempty"`
	ProfileBytes       int    `json:"profile_bytes"`
	SourceBaseBytes    int    `json:"source_base_bytes,omitempty"`
	ContentMatch       bool   `json:"content_match"`
	HashAlgorithm      string `json:"hash_algorithm"`
}

type ProvenanceEvidence struct {
	ProfilePath             string                                     `json:"profile_path"`
	SourceBasePath          string                                     `json:"source_base_path,omitempty"`
	SourcePackID            string                                     `json:"source_pack_id,omitempty"`
	SourceClass             string                                     `json:"source_class"`
	ProvenanceState         string                                     `json:"provenance_state"`
	SourceClassEvidence     []discovery.SourceClassEvidence            `json:"source_class_evidence"`
	SkillDependencies       []discovery.SkillDependencyRecord          `json:"skill_dependencies"`
	CommandDependencies     []discovery.CommandSurfaceDependencyRecord `json:"command_surface_dependencies"`
	DependencyEvidenceState string                                     `json:"dependency_evidence_state"`
}

type SemanticExtractionPacket struct {
	PacketType           string   `json:"packet_type"`
	SkillName            string   `json:"skill_name"`
	CandidateOverlayPath string   `json:"candidate_overlay_path,omitempty"`
	ExtractedSignals     []string `json:"extracted_signals"`
	DeltaPreview         []string `json:"delta_preview"`
	ReviewNotes          []string `json:"review_notes"`
}

type Item struct {
	SkillID                  string                   `json:"skill_id"`
	ProfilePath              string                   `json:"profile_path"`
	Bucket                   string                   `json:"bucket"`
	Owner                    string                   `json:"owner"`
	ReviewRequired           bool                     `json:"review_required"`
	HashEvidence             HashEvidence             `json:"hash_evidence"`
	ProvenanceEvidence       ProvenanceEvidence       `json:"provenance_evidence"`
	SemanticExtractionPacket SemanticExtractionPacket `json:"semantic_extraction_packet"`
	Diagnostics              []discovery.Diagnostic   `json:"diagnostics"`
	ForbiddenActions         []string                 `json:"forbidden_actions"`
	NextAction               string                   `json:"next_action"`
	RecoveryHint             string                   `json:"recovery_hint"`
}

type Summary struct {
	CountsByBucket map[string]int `json:"counts_by_bucket"`
	ErrorCount     int            `json:"error_count"`
	WarningCount   int            `json:"warning_count"`
}

type SourceRepo struct {
	Path      string  `json:"path"`
	State     string  `json:"state"`
	GitCommit *string `json:"git_commit,omitempty"`
	Dirty     *bool   `json:"dirty,omitempty"`
}

type TargetProfile struct {
	Name  string `json:"name"`
	Root  string `json:"root"`
	State string `json:"state"`
}

type Result struct {
	OK                         bool                                       `json:"ok"`
	Command                    string                                     `json:"command"`
	Mode                       string                                     `json:"mode"`
	CLIVersion                 string                                     `json:"cli_version"`
	ProvenanceContractVersion  string                                     `json:"provenance_contract_version"`
	SourceRepo                 SourceRepo                                 `json:"source_repo"`
	TargetProfile              TargetProfile                              `json:"target_profile"`
	Project                    string                                     `json:"project,omitempty"`
	Items                      []Item                                     `json:"items"`
	SourceClassEvidence        []discovery.SourceClassEvidence            `json:"source_class_evidence"`
	DependencyAudit            discovery.DependencyAudit                  `json:"dependency_audit"`
	SkillDependencies          []discovery.SkillDependencyRecord          `json:"skill_dependencies"`
	CommandSurfaceDependencies []discovery.CommandSurfaceDependencyRecord `json:"command_surface_dependencies"`
	DeletedBundleReference     any                                        `json:"deleted_bundle_reference"`
	DeletedBundleDiagnostics   []discovery.Diagnostic                     `json:"deleted_bundle_diagnostics"`
	NoWriteEvidence            NoWriteEvidence                            `json:"no_write_evidence"`
	NoSpilloverEvidence        NoSpilloverEvidence                        `json:"no_spillover_evidence"`
	Diagnostics                []discovery.Diagnostic                     `json:"diagnostics"`
	ReasonCodes                []string                                   `json:"reason_codes"`
	ForbiddenActions           []string                                   `json:"forbidden_actions"`
	NextAction                 string                                     `json:"next_action"`
	RecoveryHint               string                                     `json:"recovery_hint"`
	Summary                    Summary                                    `json:"summary"`
}

type skillFile struct {
	ID   string
	Rel  string
	Path string
	Data []byte
	Meta metadata
}

type metadata struct {
	Name            string
	Kind            string
	Role            string
	RoleManifest    string
	PluginNamespace string
	OverlayRoot     string
	Project         string
	OverlayFor      string
	MergeMode       string
	BaseVersion     string
}

func Build(repo string, opts Options) (Result, error) {
	sourceRepoPath, err := discovery.FindSourceRepo(repo)
	if err != nil {
		return Result{}, err
	}
	sourceInfo := discovery.SourceRepoInfo(sourceRepoPath)
	result := Result{
		OK:                        true,
		Command:                   "migrate-profile-skills",
		Mode:                      Mode,
		CLIVersion:                version.CLIVersion,
		ProvenanceContractVersion: discovery.ProvenanceContractVersion,
		SourceRepo: SourceRepo{
			Path:      sourceInfo.Path,
			State:     "ok",
			GitCommit: sourceInfo.GitCommit,
			Dirty:     sourceInfo.Dirty,
		},
		Project:                    opts.Project,
		Items:                      []Item{},
		SourceClassEvidence:        []discovery.SourceClassEvidence{},
		DependencyAudit:            discovery.EmptyDependencyAudit(),
		SkillDependencies:          []discovery.SkillDependencyRecord{},
		CommandSurfaceDependencies: []discovery.CommandSurfaceDependencyRecord{},
		DeletedBundleReference:     nil,
		DeletedBundleDiagnostics:   []discovery.Diagnostic{},
		Diagnostics:                []discovery.Diagnostic{},
		ReasonCodes:                []string{},
		ForbiddenActions:           forbiddenActions(),
		NextAction:                 "Review this dry-run report with Blue/Red/Orange and obtain explicit SKILL-006-like approval before any deletion, migration, conversion, install, update, apply, or profile mutation.",
		RecoveryHint:               "Rerun after fixing missing or ambiguous inventory, hash, provenance, dependency, ownership, or boundary evidence; do not use copied profile skills as fallback.",
		NoWriteEvidence: NoWriteEvidence{
			Guaranteed: true,
		},
		NoSpilloverEvidence: NoSpilloverEvidence{
			Guaranteed: true,
		},
		Summary: Summary{CountsByBucket: emptyBucketCounts()},
	}

	profileName := strings.TrimSpace(opts.Profile)
	if profileName == "" {
		addDiag(&result, "error", "profile_required", "migrate-profile-skills requires --profile <profile>.")
		finalize(&result)
		return result, nil
	}
	if err := validateProfileName(profileName); err != nil {
		result.TargetProfile = TargetProfile{Name: profileName, State: "invalid"}
		addDiag(&result, "error", "profile_invalid", err.Error())
		finalize(&result)
		return result, nil
	}
	opts.Profile = profileName

	sourcePacks, err := discovery.DiscoverSourcePacks(sourceRepoPath)
	if err != nil {
		addDiag(&result, "error", "source_inventory_unavailable", "source inventory evidence is unavailable: "+err.Error())
	} else {
		inventory := discovery.BuildSourceInventory(sourcePacks, nil, nil)
		result.DependencyAudit = discovery.BuildDependencyAudit(sourceRepoPath, sourcePacks, inventory)
		result.SkillDependencies = append([]discovery.SkillDependencyRecord{}, result.DependencyAudit.SkillDependencies...)
		result.CommandSurfaceDependencies = append([]discovery.CommandSurfaceDependencyRecord{}, result.DependencyAudit.CommandSurfaceDependencies...)
		result.DeletedBundleDiagnostics = append([]discovery.Diagnostic{}, result.DependencyAudit.DeletedBundleDiagnostics...)
		for _, pack := range sourcePacks {
			record := discovery.SourceOnlyProvenance(pack)
			result.SourceClassEvidence = append(result.SourceClassEvidence, record.SourceClassEvidence...)
		}
	}
	if result.DependencyAudit.State == "" {
		addDiag(&result, "error", "dependency_evidence_missing", "dependency audit state is missing; classifier cannot prove KASREL dependency readback.")
	}

	pkg, err := discovery.LoadPluginPackage(sourceRepoPath)
	if err != nil {
		addDiag(&result, "error", "missing_plugin_evidence", "official KAS plugin package evidence is missing or invalid: "+err.Error())
	}

	profileRoot := resolveProfileRoot(opts.Profile, opts.ProfileRoot)
	result.TargetProfile = TargetProfile{Name: opts.Profile, Root: profileRoot, State: "ok"}
	if st, err := os.Stat(profileRoot); err != nil {
		if os.IsNotExist(err) {
			result.TargetProfile.State = "missing"
			addDiag(&result, "error", "profile_missing", "profile root does not exist: "+profileRoot)
		} else {
			result.TargetProfile.State = "unreadable"
			addDiag(&result, "error", "profile_unreadable", "profile root is not readable: "+err.Error())
		}
		finalize(&result)
		return result, nil
	} else if !st.IsDir() {
		result.TargetProfile.State = "invalid"
		addDiag(&result, "error", "profile_not_directory", "profile root is not a directory: "+profileRoot)
		finalize(&result)
		return result, nil
	}

	before, err := treeSHA256(profileRoot)
	if err != nil {
		addDiag(&result, "error", "profile_hash_unavailable", "profile tree hash before classification is unavailable: "+err.Error())
	}
	result.NoWriteEvidence.ProfileTreeHashBefore = before
	result.NoSpilloverEvidence.ProfileTreeHashBefore = before
	result.NoSpilloverEvidence.ProfileRoot = profileRoot
	result.NoSpilloverEvidence.SkillsRoot = filepath.Join(profileRoot, "skills")

	if len(sourcePacks) > 0 && pkg.Namespace != "" {
		classifyProfile(&result, sourceRepoPath, profileRoot, sourcePacks)
	}

	after, err := treeSHA256(profileRoot)
	if err != nil {
		addDiag(&result, "error", "profile_hash_unavailable", "profile tree hash after classification is unavailable: "+err.Error())
	}
	result.NoWriteEvidence.ProfileTreeHashAfter = after
	result.NoWriteEvidence.ProfileTreeUnchanged = before != "" && before == after
	result.NoSpilloverEvidence.ProfileTreeHashAfter = after
	result.NoSpilloverEvidence.ProfileTreeHashMatch = before != "" && before == after
	if before != "" && after != "" && before != after {
		addDiag(&result, "error", "no_write_proof_failed", "profile tree hash changed during dry-run classification.")
	}

	finalize(&result)
	return result, nil
}

func classifyProfile(result *Result, sourceRepo string, profileRoot string, sourcePacks []discovery.SourcePack) {
	skillsRoot := filepath.Join(profileRoot, "skills")
	if st, err := os.Stat(skillsRoot); err != nil {
		addDiag(result, "error", "profile_inventory_missing", "profile skills root is missing or unreadable: "+skillsRoot)
		return
	} else if !st.IsDir() {
		addDiag(result, "error", "profile_inventory_invalid", "profile skills root is not a directory: "+skillsRoot)
		return
	}
	baseByID := map[string]discovery.SourcePack{}
	for _, pack := range sourcePacks {
		baseByID[pack.Name] = pack
		baseByID[pack.PackID] = pack
	}
	err := filepath.WalkDir(skillsRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			addDiag(result, "error", "profile_inventory_unreadable", "profile inventory walk failed: "+err.Error())
			return nil
		}
		if entry.IsDir() || entry.Name() != "SKILL.md" {
			return nil
		}
		rel, relErr := filepath.Rel(profileRoot, path)
		if relErr != nil {
			addDiag(result, "error", "profile_inventory_ambiguous", "profile skill path could not be relativized: "+relErr.Error())
			return filepath.SkipDir
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			addUnreadableItem(result, filepath.ToSlash(rel), readErr)
			return filepath.SkipDir
		}
		dirRel := filepath.ToSlash(filepath.Dir(rel))
		skillID := filepath.Base(filepath.Dir(path))
		meta := parseMetadata(data)
		if meta.Name == "" {
			meta.Name = skillID
		}
		item := classifySkill(sourceRepo, skillFile{ID: skillID, Rel: filepath.ToSlash(rel), Path: path, Data: data, Meta: meta}, baseByID)
		result.NoSpilloverEvidence.InspectedSkillCount++
		if item.Bucket == "unknown_personal_skill" {
			result.NoSpilloverEvidence.UnrelatedSkillCount++
			result.NoSpilloverEvidence.UnrelatedSkillsPreserved = append(result.NoSpilloverEvidence.UnrelatedSkillsPreserved, dirRel)
		}
		result.Items = append(result.Items, item)
		return filepath.SkipDir
	})
	if err != nil {
		addDiag(result, "error", "profile_inventory_unreadable", "profile inventory walk failed: "+err.Error())
	}
	sort.Slice(result.Items, func(i, j int) bool { return result.Items[i].ProfilePath < result.Items[j].ProfilePath })
	sort.Strings(result.NoSpilloverEvidence.UnrelatedSkillsPreserved)
}

func classifySkill(sourceRepo string, skill skillFile, baseByID map[string]discovery.SourcePack) Item {
	profileHash := sha256Hex(skill.Data)
	item := Item{
		SkillID:        skill.ID,
		ProfilePath:    skill.Rel,
		Bucket:         "unknown_personal_skill",
		Owner:          "profile_owner",
		ReviewRequired: true,
		HashEvidence: HashEvidence{
			ProfileSkillSHA256: profileHash,
			ProfileBytes:       len(skill.Data),
			HashAlgorithm:      "sha256",
		},
		ProvenanceEvidence: ProvenanceEvidence{
			ProfilePath:             skill.Rel,
			SourceClass:             "profile_local",
			ProvenanceState:         "review_required",
			SourceClassEvidence:     []discovery.SourceClassEvidence{{Kind: "profile_skill", Path: skill.Rel, State: "matched", Detail: "profile-local skill inspected by dry-run classifier"}},
			SkillDependencies:       []discovery.SkillDependencyRecord{},
			CommandDependencies:     []discovery.CommandSurfaceDependencyRecord{},
			DependencyEvidenceState: "readback_required",
		},
		SemanticExtractionPacket: SemanticExtractionPacket{
			PacketType:       "preserve_review",
			SkillName:        skill.Meta.Name,
			ExtractedSignals: []string{},
			DeltaPreview:     previewLines(skill.Data, 4),
			ReviewNotes:      []string{"Preserve profile-local material until reviewed; SKILL-005 does not migrate or delete it."},
		},
		Diagnostics:      []discovery.Diagnostic{},
		ForbiddenActions: forbiddenActions(),
		NextAction:       "Preserve and review; do not migrate, delete, convert, install, update, apply, or use as fallback.",
		RecoveryHint:     "Provide explicit source/provenance/dependency evidence and later approval before any migration action.",
	}

	if pack, ok := baseByID[skill.ID]; ok {
		applyBaseEvidence(sourceRepo, &item, pack)
		if item.HashEvidence.ContentMatch {
			item.Bucket = "base_identical"
			item.Owner = "kas_plugin_base"
			item.SemanticExtractionPacket.PacketType = "base_identical_readback"
			item.SemanticExtractionPacket.ReviewNotes = []string{"Content matches source base. Removal is allowed only in a later approved SKILL-006-like flow, never in this classifier."}
			item.NextAction = "Record as removable candidate only for later approved SKILL-006-like migration with backup, no-spillover scan, recovery evidence, and explicit approval."
		} else {
			item.Bucket = "base_with_local_delta"
			item.Owner = "kas_profile_migration_review"
			item.SemanticExtractionPacket.PacketType = "base_delta_extraction"
			item.SemanticExtractionPacket.ExtractedSignals = append(item.SemanticExtractionPacket.ExtractedSignals, "profile-local copy differs from source base")
			item.SemanticExtractionPacket.ReviewNotes = []string{"Extract semantic delta for overlay/common-guide review; do not overwrite source base or profile skill."}
			item.NextAction = "Review local delta and decide whether it belongs in a project overlay or common guide candidate; no migration is authorized here."
		}
	}

	switch {
	case isProjectOverlayCandidate(skill):
		item.Bucket = "project_overlay_candidate"
		item.Owner = "kas_overlay_review"
		item.ReviewRequired = true
		item.SemanticExtractionPacket.PacketType = "project_overlay_candidate"
		item.SemanticExtractionPacket.CandidateOverlayPath = candidateOverlayPath(skill)
		item.SemanticExtractionPacket.ExtractedSignals = append(item.SemanticExtractionPacket.ExtractedSignals, "project overlay metadata or layout detected")
		item.SemanticExtractionPacket.ReviewNotes = []string{"Review as project overlay candidate; SKILL-005 does not create or move overlay files."}
		item.NextAction = "Review for project overlay extraction and conflict handling; create overlays only in a later approved flow."
	case isRoleWrapperCandidate(skill):
		item.Bucket = "role_wrapper_candidate"
		item.Owner = "kas_wrapper_review"
		item.ReviewRequired = true
		item.SemanticExtractionPacket.PacketType = "role_wrapper_candidate"
		item.SemanticExtractionPacket.ExtractedSignals = append(item.SemanticExtractionPacket.ExtractedSignals, "role wrapper metadata detected")
		item.SemanticExtractionPacket.ReviewNotes = []string{"Review as thin wrapper candidate; SKILL-005 does not write wrapper files."}
		item.NextAction = "Review wrapper role, manifest, plugin namespace, and overlay root before any wrapper adoption."
	case isKAHCompanionSurface(skill):
		item.Bucket = "kah_companion_surface"
		item.Owner = "kah_companion_review"
		item.ReviewRequired = true
		item.SemanticExtractionPacket.PacketType = "kah_companion_surface"
		item.SemanticExtractionPacket.ExtractedSignals = append(item.SemanticExtractionPacket.ExtractedSignals, "KAH companion surface detected")
		item.SemanticExtractionPacket.ReviewNotes = []string{"Route to KAH or paired KAS/KAH companion handling; do not perform KAS-only cleanup."}
		item.NextAction = "Route to KAH/companion review; KAS-only deletion or conversion is forbidden."
	}

	if containsRuntimeConfig(skill.Data) {
		item.Diagnostics = append(item.Diagnostics, discovery.Diagnostic{Level: "error", Code: "runtime_config_boundary_violation", Message: "profile skill appears to contain auth/token/gateway/provider/model/runtime configuration text: " + skill.Rel})
		item.ReviewRequired = true
	}
	if containsOwnershipConflict(skill.Data) {
		item.Diagnostics = append(item.Diagnostics, discovery.Diagnostic{Level: "error", Code: "ownership_boundary_conflict", Message: "profile skill appears to assign KAS plugin/wrapper/update ownership to KAH or another runtime layer: " + skill.Rel})
		item.ReviewRequired = true
	}
	return item
}

func applyBaseEvidence(sourceRepo string, item *Item, pack discovery.SourcePack) {
	basePath := filepath.Join(sourceRepo, filepath.FromSlash(pack.SourcePath), "SKILL.md")
	data, err := os.ReadFile(basePath)
	item.ProvenanceEvidence.SourceBasePath = filepath.ToSlash(filepath.Join(pack.SourcePath, "SKILL.md"))
	item.ProvenanceEvidence.SourcePackID = pack.PackID
	item.ProvenanceEvidence.SourceClass = "plugin_base"
	item.ProvenanceEvidence.ProvenanceState = "classified"
	item.ProvenanceEvidence.SourceClassEvidence = append(item.ProvenanceEvidence.SourceClassEvidence, discovery.SourceClassEvidence{Kind: "source_repo", Path: item.ProvenanceEvidence.SourceBasePath, State: "matched", Detail: "profile skill name matches source plugin base"})
	item.ProvenanceEvidence.SkillDependencies = append([]discovery.SkillDependencyRecord{}, pack.SkillDependencies...)
	item.ProvenanceEvidence.CommandDependencies = append([]discovery.CommandSurfaceDependencyRecord{}, pack.CommandSurfaceDependencies...)
	item.ProvenanceEvidence.DependencyEvidenceState = "readback"
	if err != nil {
		item.Diagnostics = append(item.Diagnostics, discovery.Diagnostic{Level: "error", Code: "source_base_hash_unavailable", Message: "source base skill hash is unavailable: " + err.Error()})
		return
	}
	item.HashEvidence.SourceBaseSHA256 = sha256Hex(data)
	item.HashEvidence.SourceBaseBytes = len(data)
	item.HashEvidence.ContentMatch = item.HashEvidence.ProfileSkillSHA256 == item.HashEvidence.SourceBaseSHA256
}

func addUnreadableItem(result *Result, rel string, err error) {
	item := Item{
		SkillID:        filepath.Base(filepath.Dir(rel)),
		ProfilePath:    rel,
		Bucket:         "unknown_personal_skill",
		Owner:          "profile_owner",
		ReviewRequired: true,
		HashEvidence:   HashEvidence{HashAlgorithm: "sha256"},
		ProvenanceEvidence: ProvenanceEvidence{
			ProfilePath:             rel,
			SourceClass:             "unknown",
			ProvenanceState:         "unreadable",
			SourceClassEvidence:     []discovery.SourceClassEvidence{{Kind: "profile_skill", Path: rel, State: "unreadable", Detail: "profile skill could not be read"}},
			SkillDependencies:       []discovery.SkillDependencyRecord{},
			CommandDependencies:     []discovery.CommandSurfaceDependencyRecord{},
			DependencyEvidenceState: "missing",
		},
		SemanticExtractionPacket: SemanticExtractionPacket{
			PacketType:       "unreadable_review",
			SkillName:        filepath.Base(filepath.Dir(rel)),
			ExtractedSignals: []string{},
			DeltaPreview:     []string{},
			ReviewNotes:      []string{"Unreadable profile skill must be reviewed before classification."},
		},
		Diagnostics:      []discovery.Diagnostic{{Level: "error", Code: "profile_skill_unreadable", Message: "profile skill is unreadable: " + rel + ": " + err.Error()}},
		ForbiddenActions: forbiddenActions(),
		NextAction:       "Fix readability/provenance evidence and rerun dry-run classifier; do not use fallback classification.",
		RecoveryHint:     "Restore readable profile skill evidence from backup or permissions before migration planning.",
	}
	result.Items = append(result.Items, item)
}

func RenderHuman(result Result) string {
	state := "pass"
	if !result.OK {
		state = "blocked"
	}
	lines := []string{
		fmt.Sprintf("Status: %s; SKILL-005 migration classifier dry-run/report-only.", state),
		fmt.Sprintf("Summary: items %d, errors %d, warnings %d.", len(result.Items), result.Summary.ErrorCount, result.Summary.WarningCount),
		"Mode: " + result.Mode,
		fmt.Sprintf("Writes: no writes performed=%t; profile deletions/migrations/conversions 0; KAH/KAB/runtime/auth/provider/model mutations 0.", result.NoWriteEvidence.Guaranteed && result.NoWriteEvidence.ProfileTreeUnchanged),
		"Boundary: no deletion, migration, install, update, apply, repair, conversion, profile mutation, or fallback is authorized.",
		"Next approval gate: Blue/Red/Orange review plus explicit SKILL-006-like approval before any action beyond reporting.",
	}
	for _, item := range result.Items {
		lines = append(lines, fmt.Sprintf("Item[%s]: bucket=%s review_required=%t owner=%s next=%s", item.ProfilePath, item.Bucket, item.ReviewRequired, item.Owner, item.NextAction))
	}
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, fmt.Sprintf("Diagnostic[%s:%s]: %s", diagnostic.Level, diagnostic.Code, diagnostic.Message))
	}
	lines = append(lines, "Next: "+result.NextAction)
	return strings.Join(lines, "\n")
}

func finalize(result *Result) {
	result.Summary.CountsByBucket = emptyBucketCounts()
	result.Summary.ErrorCount = 0
	result.Summary.WarningCount = 0
	result.ReasonCodes = []string{}
	result.OK = true
	diagnostics := append([]discovery.Diagnostic{}, result.Diagnostics...)
	if len(result.Items) == 0 && result.TargetProfile.State == "ok" {
		diagnostics = append(diagnostics, discovery.Diagnostic{Level: "error", Code: "profile_inventory_empty", Message: "profile skills inventory is empty or no readable SKILL.md files were found."})
	}
	for _, item := range result.Items {
		result.Summary.CountsByBucket[item.Bucket]++
		for _, diagnostic := range item.Diagnostics {
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	for _, diagnostic := range result.DependencyAudit.Diagnostics {
		if diagnostic.Level == "error" {
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	result.Diagnostics = dedupeDiagnostics(diagnostics)
	sort.Slice(result.Diagnostics, func(i, j int) bool {
		if result.Diagnostics[i].Code == result.Diagnostics[j].Code {
			return result.Diagnostics[i].Message < result.Diagnostics[j].Message
		}
		return result.Diagnostics[i].Code < result.Diagnostics[j].Code
	})
	seen := map[string]bool{}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Level == "error" {
			result.Summary.ErrorCount++
			result.OK = false
		}
		if diagnostic.Level == "warning" {
			result.Summary.WarningCount++
		}
		if !seen[diagnostic.Code] {
			result.ReasonCodes = append(result.ReasonCodes, diagnostic.Code)
			seen[diagnostic.Code] = true
		}
	}
	sort.Strings(result.ReasonCodes)
}

func dedupeDiagnostics(diagnostics []discovery.Diagnostic) []discovery.Diagnostic {
	seen := map[string]bool{}
	out := make([]discovery.Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		key := diagnostic.Level + "\x00" + diagnostic.Code + "\x00" + diagnostic.Message
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, diagnostic)
	}
	return out
}

func addDiag(result *Result, level string, code string, message string) {
	result.Diagnostics = append(result.Diagnostics, discovery.Diagnostic{Level: level, Code: code, Message: message})
}

func validateProfileName(profile string) error {
	if profile == "" {
		return fmt.Errorf("profile name must not be empty")
	}
	if profile == "." || profile == ".." {
		return fmt.Errorf("profile name %q is not allowed", profile)
	}
	if strings.Contains(profile, "/") || strings.Contains(profile, `\`) || strings.Contains(profile, string(os.PathSeparator)) {
		return fmt.Errorf("profile name %q must not contain path separators", profile)
	}
	return nil
}

func emptyBucketCounts() map[string]int {
	return map[string]int{
		"base_identical":            0,
		"base_with_local_delta":     0,
		"project_overlay_candidate": 0,
		"role_wrapper_candidate":    0,
		"unknown_personal_skill":    0,
		"kah_companion_surface":     0,
	}
}

func forbiddenActions() []string {
	return []string{"delete", "migrate", "convert", "install", "update", "apply", "repair", "profile_mutation", "kah_state_mutation", "kab_runtime_mutation", "auth_token_gateway_provider_model_runtime_mutation", "fallback_to_copied_skill"}
}

func parseMetadata(data []byte) metadata {
	meta := metadata{}
	lines := strings.Split(string(data), "\n")
	inFrontmatter := false
	pathByIndent := map[int][]string{}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 && trimmed == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && trimmed == "---" {
			break
		}
		if !inFrontmatter || trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "- ") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if value == "" {
			parent := pathByIndent[indent-2]
			pathByIndent[indent] = append(append([]string{}, parent...), splitMetadataKey(key)...)
			continue
		}
		parent := pathByIndent[indent-2]
		path := append(append([]string{}, parent...), splitMetadataKey(key)...)
		if len(path) == 3 && path[0] == "metadata" && path[1] == "kas" {
			assignMetadata(&meta, path[2], value)
			continue
		}
		if indent == 0 {
			assignMetadata(&meta, key, value)
			continue
		}
	}
	return meta
}

func splitMetadataKey(key string) []string {
	parts := strings.Split(key, ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func assignMetadata(meta *metadata, key string, value string) {
	switch key {
	case "name":
		meta.Name = value
	case "kind":
		meta.Kind = value
	case "role":
		meta.Role = value
	case "role_manifest":
		meta.RoleManifest = value
	case "plugin_namespace":
		meta.PluginNamespace = value
	case "overlay_root":
		meta.OverlayRoot = value
	case "project":
		meta.Project = value
	case "overlay_for":
		meta.OverlayFor = value
	case "merge_mode":
		meta.MergeMode = value
	case "base_version":
		meta.BaseVersion = value
	}
}

func isProjectOverlayCandidate(skill skillFile) bool {
	return skill.Meta.Kind == "project_overlay" || strings.Contains(filepath.ToSlash(skill.Rel), "/kas-overlays/") || skill.Meta.OverlayFor != "" || skill.Meta.MergeMode != ""
}

func isRoleWrapperCandidate(skill skillFile) bool {
	return skill.Meta.Kind == "color_wrapper" || skill.Meta.RoleManifest != "" || skill.Meta.PluginNamespace != "" || skill.Meta.OverlayRoot != ""
}

func isKAHCompanionSurface(skill skillFile) bool {
	text := strings.ToLower(string(skill.Data))
	return strings.Contains(text, "kkachi-agent-helper") || strings.Contains(text, "kah companion") || strings.Contains(text, "kah-owned") || strings.Contains(text, "kah owns")
}

func containsRuntimeConfig(data []byte) bool {
	text := strings.ToLower(string(data))
	for _, needle := range []string{"auth_token", "api_key", "gateway:", "provider:", "model:", "runtime:", "token:"} {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func containsOwnershipConflict(data []byte) bool {
	text := strings.ToLower(string(data))
	return strings.Contains(text, "kah owns plugin") || strings.Contains(text, "kah owns kas") || strings.Contains(text, "kah owns wrapper") || strings.Contains(text, "kah owns update")
}

func candidateOverlayPath(skill skillFile) string {
	project := skill.Meta.Project
	if project == "" {
		parts := strings.Split(filepath.ToSlash(skill.Rel), "/")
		if len(parts) >= 4 && parts[0] == "skills" && parts[2] == "kas-overlays" {
			project = parts[1]
		}
	}
	if project == "" {
		project = "<project>"
	}
	name := strings.TrimPrefix(skill.ID, "kkachi-")
	return "skills/" + project + "/kas-overlays/" + project + "-" + name + "-overlay/SKILL.md"
}

func previewLines(data []byte, limit int) []string {
	lines := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" {
			continue
		}
		lines = append(lines, trimmed)
		if len(lines) >= limit {
			break
		}
	}
	return lines
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func treeSHA256(root string) (string, error) {
	entries := []string{}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		entries = append(entries, filepath.ToSlash(rel)+"\x00"+hex.EncodeToString(sum[:]))
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(entries)
	sum := sha256.Sum256([]byte(strings.Join(entries, "\n")))
	return hex.EncodeToString(sum[:]), nil
}

func resolveProfileRoot(profile string, override string) string {
	root := strings.TrimSpace(override)
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			home = "."
		}
		root = filepath.Join(home, ".hermes", "profiles", profile)
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}
