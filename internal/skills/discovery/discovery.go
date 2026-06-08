package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const NextListAction = "Run install --dry-run before any profile writes."

type SourcePack struct {
	PackID                     string
	Category                   string
	Name                       string
	SourcePath                 string
	Description                string
	Checksum                   string
	SkillDependencies          []SkillDependencyRecord
	CommandSurfaceDependencies []CommandSurfaceDependencyRecord
	DependencyDiagnostics      []Diagnostic
}

type SourceRepo struct {
	Path      string  `json:"path"`
	GitCommit *string `json:"git_commit"`
	Dirty     *bool   `json:"dirty"`
}

type TargetProfile struct {
	Name                   string  `json:"name"`
	Root                   string  `json:"root"`
	ManifestPath           string  `json:"manifest_path"`
	ManifestState          string  `json:"manifest_state"`
	PreviousManifestSHA256 *string `json:"previous_manifest_sha256,omitempty"`
}

type Diagnostic struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ListOptions struct {
	Category    string
	Profile     string
	ProfileRoot string
}

type ListResult struct {
	OK                        bool                    `json:"ok"`
	Command                   string                  `json:"command"`
	ProvenanceContractVersion string                  `json:"provenance_contract_version"`
	SourceRepo                SourceRepo              `json:"source_repo"`
	TargetProfile             *TargetProfile          `json:"target_profile,omitempty"`
	SourceInventorySummary    SourceInventorySummary  `json:"source_inventory_summary"`
	SourceInventorySnapshot   SourceInventorySnapshot `json:"source_inventory_snapshot"`
	Packs                     []map[string]any        `json:"packs"`
	Diagnostics               []Diagnostic            `json:"diagnostics"`
	NextAction                string                  `json:"next_action"`
}

type manifestInstall struct {
	PackID       string           `json:"pack_id"`
	TargetPath   string           `json:"target_path"`
	PackChecksum string           `json:"pack_checksum"`
	Files        []map[string]any `json:"files"`
	Extra        map[string]any   `json:"-"`
}

func FindSourceRepo(start string) (string, error) {
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = cwd
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("source repo path does not exist: %s", abs)
	}
	current := abs
	if !info.IsDir() {
		current = filepath.Dir(abs)
	}
	for {
		if st, err := os.Stat(filepath.Join(current, "skills")); err == nil && st.IsDir() {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("source repo not found from %s", abs)
}

func DiscoverSourcePacks(repo string) ([]SourcePack, error) {
	sourceRepo, err := FindSourceRepo(repo)
	if err != nil {
		return nil, err
	}
	skillsDir := filepath.Join(sourceRepo, "skills")
	info, err := os.Stat(skillsDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("source repo has no skills directory: %s", sourceRepo)
	}

	children, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}
	sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })

	packs := []SourcePack{}
	for _, child := range children {
		if !child.IsDir() {
			continue
		}
		childPath := filepath.Join(skillsDir, child.Name())
		if isFile(filepath.Join(childPath, "SKILL.md")) {
			pack, err := readPack(sourceRepo, childPath, "core", child.Name())
			if err != nil {
				return nil, err
			}
			packs = append(packs, pack)
			continue
		}
		grandchildren, err := os.ReadDir(childPath)
		if err != nil {
			return nil, err
		}
		sort.Slice(grandchildren, func(i, j int) bool { return grandchildren[i].Name() < grandchildren[j].Name() })
		for _, grandchild := range grandchildren {
			if !grandchild.IsDir() {
				continue
			}
			packPath := filepath.Join(childPath, grandchild.Name())
			if !isFile(filepath.Join(packPath, "SKILL.md")) {
				continue
			}
			pack, err := readPack(sourceRepo, packPath, child.Name(), child.Name()+"/"+grandchild.Name())
			if err != nil {
				return nil, err
			}
			packs = append(packs, pack)
		}
	}
	if len(packs) == 0 {
		return nil, fmt.Errorf("no readable KAS skill packs found under %s", skillsDir)
	}
	return packs, nil
}

func BuildListResult(repo string, opts ListOptions) (ListResult, error) {
	sourceRepo, err := FindSourceRepo(repo)
	if err != nil {
		return ListResult{}, err
	}
	packs, err := DiscoverSourcePacks(sourceRepo)
	if err != nil {
		return ListResult{}, err
	}

	diagnostics := []Diagnostic{}
	if opts.Category != "" {
		filtered := []SourcePack{}
		for _, pack := range packs {
			if pack.Category == opts.Category {
				filtered = append(filtered, pack)
			}
		}
		if len(filtered) == 0 {
			diagnostics = append(diagnostics, Diagnostic{
				Level:   "info",
				Code:    "unknown_category",
				Message: fmt.Sprintf("카테고리 '%s'에 해당하는 KAS pack이 없습니다.", opts.Category),
			})
		}
		packs = filtered
	}

	targetProfile, installs := loadProfile(opts.Profile, opts.ProfileRoot)
	inventory := BuildSourceInventory(packs, targetProfile, installs)
	payloads := make([]map[string]any, 0, len(packs))
	for _, pack := range packs {
		payloads = append(payloads, packPayload(pack, installs, targetProfile, inventory))
	}
	return ListResult{
		OK:                        true,
		Command:                   "list",
		ProvenanceContractVersion: ProvenanceContractVersion,
		SourceRepo:                SourceRepoInfo(sourceRepo),
		TargetProfile:             targetProfile,
		SourceInventorySummary:    inventory.Summary,
		SourceInventorySnapshot:   inventory,
		Packs:                     payloads,
		Diagnostics:               diagnostics,
		NextAction:                NextListAction,
	}, nil
}

func RenderHumanList(result ListResult) string {
	lines := []string{fmt.Sprintf("상태: 조회 완료 — KAS pack %d개 발견.", len(result.Packs))}
	if result.TargetProfile != nil {
		counts := map[string]int{}
		for _, pack := range result.Packs {
			if state, ok := pack["installed_state"].(string); ok {
				counts[state]++
			}
		}
		lines = append(lines, fmt.Sprintf(
			"설치 상태: current %d, missing %d, drifted %d, unknown %d, conflict %d, error %d.",
			counts["installed_current"],
			counts["not_installed"]+counts["manifest_missing"],
			counts["installed_drifted"],
			counts["installed_unknown"],
			counts["conflict"],
			counts["error"],
		))
	}
	commit := "unknown"
	if result.SourceRepo.GitCommit != nil && *result.SourceRepo.GitCommit != "" {
		commit = *result.SourceRepo.GitCommit
	}
	lines = append(lines, fmt.Sprintf("소스: %s @ %s", result.SourceRepo.Path, commit))
	for _, diagnostic := range result.Diagnostics {
		lines = append(lines, "진단: "+diagnostic.Message)
	}
	lines = append(lines, "다음: 설치 전 `install --dry-run`으로 변경 경로를 확인하세요.")
	return strings.Join(lines, "\n")
}

func PackPayload(pack SourcePack) map[string]any {
	payload := map[string]any{
		"pack_id":     pack.PackID,
		"category":    pack.Category,
		"name":        pack.Name,
		"source_path": pack.SourcePath,
	}
	if pack.Description != "" {
		payload["description"] = pack.Description
	}
	return payload
}

func SourceRepoInfo(repo string) SourceRepo {
	commit := gitValue(repo, "rev-parse", "HEAD")
	dirty := gitDirty(repo)
	return SourceRepo{Path: repo, GitCommit: commit, Dirty: dirty}
}

func IsInvalidRelativePath(value string) bool {
	if value == "" || filepath.IsAbs(value) {
		return true
	}
	if strings.Contains(value, "\\") {
		return true
	}
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "" || part == ".." {
			return true
		}
	}
	return false
}

func ModeString(mode os.FileMode) string {
	return modeString(mode)
}

func readPack(sourceRepo string, packDir string, category string, packID string) (SourcePack, error) {
	skillPath := filepath.Join(packDir, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return SourcePack{}, fmt.Errorf("cannot read pack metadata: %s", skillPath)
	}
	frontmatter := parseFrontmatterValues(string(data), knownFrontmatterKeys())
	metadata := frontmatterMetadata(frontmatter)
	rel, err := filepath.Rel(sourceRepo, packDir)
	if err != nil {
		return SourcePack{}, err
	}
	name := metadata["name"]
	if name == "" {
		name = filepath.Base(packDir)
	}
	checksum, err := ComputePackChecksum(packDir)
	if err != nil {
		return SourcePack{}, err
	}
	skillRelPath := filepath.ToSlash(filepath.Join(rel, "SKILL.md"))
	return SourcePack{
		PackID:                     packID,
		Category:                   category,
		Name:                       name,
		Description:                metadata["description"],
		SourcePath:                 filepath.ToSlash(rel),
		Checksum:                   checksum,
		SkillDependencies:          frontmatterSkillDependencies(frontmatter, skillRelPath),
		CommandSurfaceDependencies: frontmatterCommandSurfaceDependencies(frontmatter, skillRelPath),
		DependencyDiagnostics:      []Diagnostic{},
	}, nil
}

func ComputePackChecksum(packDir string) (string, error) {
	type entry struct {
		Path   string `json:"path"`
		Bytes  int    `json:"bytes"`
		Mode   string `json:"mode"`
		SHA256 string `json:"sha256"`
	}
	entries := []entry{}
	err := filepath.WalkDir(packDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(packDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if excluded(rel) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		entries = append(entries, entry{
			Path:   rel,
			Bytes:  len(data),
			Mode:   modeString(info.Mode().Perm()),
			SHA256: hex.EncodeToString(sum[:]),
		})
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	payload, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func packPayload(pack SourcePack, installs map[string]map[string]any, profile *TargetProfile, inventory SourceInventorySnapshot) map[string]any {
	payload := PackPayload(pack)
	ApplyProvenance(payload, PackProvenance(pack, inventory))
	if profile == nil {
		return payload
	}
	switch profile.ManifestState {
	case "manifest_missing":
		payload["installed_state"] = "manifest_missing"
		return payload
	case "manifest_unreadable":
		payload["installed_state"] = "error"
		return payload
	}
	install := installs[pack.PackID]
	if install == nil {
		payload["installed_state"] = "not_installed"
		return payload
	}
	if targetPath, ok := install["target_path"].(string); ok {
		payload["installed_path"] = targetPath
		if IsInvalidRelativePath(targetPath) {
			payload["installed_state"] = "conflict"
			return payload
		}
	}
	installedChecksum, _ := install["pack_checksum"].(string)
	if installedChecksum == "" {
		payload["installed_state"] = "installed_unknown"
	} else if installedChecksum == pack.Checksum {
		payload["installed_state"] = "installed_current"
	} else {
		payload["installed_state"] = "installed_drifted"
	}
	return payload
}

func loadProfile(profile string, profileRoot string) (*TargetProfile, map[string]map[string]any) {
	if profile == "" {
		return nil, nil
	}
	root := profileRoot
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		root = filepath.Join(home, ".hermes", "profiles", profile)
	}
	rootAbs, _ := filepath.Abs(root)
	manifestPath := filepath.Join(rootAbs, ".kas", "skill-pack-manifest.json")
	target := &TargetProfile{
		Name:          profile,
		Root:          rootAbs,
		ManifestPath:  manifestPath,
		ManifestState: "manifest_missing",
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return target, nil
		}
		target.ManifestState = "manifest_unreadable"
		return target, nil
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		target.ManifestState = "manifest_unreadable"
		return target, nil
	}
	target.ManifestState = "manifest_present"
	installs := map[string]map[string]any{}
	rawInstalls, _ := manifest["installs"].([]any)
	for _, raw := range rawInstalls {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		packID, ok := entry["pack_id"].(string)
		if ok && packID != "" {
			installs[packID] = entry
		}
	}
	return target, installs
}

func parseFrontmatter(text string) map[string]string {
	values := parseFrontmatterValues(text, map[string]bool{"name": true, "description": true})
	return frontmatterMetadata(values)
}

func frontmatterMetadata(values map[string][]string) map[string]string {
	metadata := map[string]string{}
	for key, entries := range values {
		if len(entries) > 0 {
			metadata[key] = entries[0]
		}
	}
	return metadata
}

func knownFrontmatterKeys() map[string]bool {
	return map[string]bool{
		"name":              true,
		"description":       true,
		"related_skills":    true,
		"required_skills":   true,
		"required_commands": true,
		"required_env":      true,
	}
}

func parseFrontmatterValues(text string, allowed map[string]bool) map[string][]string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return map[string][]string{}
	}
	metadata := map[string][]string{}
	currentListKey := ""
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if currentListKey != "" && strings.HasPrefix(trimmed, "- ") {
			value := cleanScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
			if value != "" {
				metadata[currentListKey] = append(metadata[currentListKey], value)
			}
			continue
		}
		currentListKey = ""
		before, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key := strings.TrimSpace(before)
		if !allowed[key] {
			continue
		}
		value := cleanScalar(strings.TrimSpace(after))
		if value == "" {
			currentListKey = key
			if metadata[key] == nil {
				metadata[key] = []string{}
			}
			continue
		}
		metadata[key] = []string{value}
	}
	return metadata
}

func cleanScalar(value string) string {
	if len(value) >= 2 {
		first := value[0]
		if (first == '\'' || first == '"') && value[len(value)-1] == first {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func excluded(relativePath string) bool {
	parts := strings.Split(relativePath, "/")
	for _, part := range parts {
		if part == ".git" || part == ".kkachi" || part == "__pycache__" {
			return true
		}
		if part == ".DS_Store" || strings.HasSuffix(part, ".swp") || strings.HasSuffix(part, ".swo") {
			return true
		}
	}
	return false
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func modeString(mode os.FileMode) string {
	return fmt.Sprintf("%04o", mode.Perm())
}

func gitValue(repo string, args ...string) *string {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	value := strings.TrimSpace(string(out))
	if value == "" {
		return nil
	}
	return &value
}

func gitDirty(repo string) *bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	dirty := strings.TrimSpace(string(out)) != ""
	return &dirty
}
