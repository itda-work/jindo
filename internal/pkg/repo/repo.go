package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/itda-skills/jindo/internal/pkg/git"
)

const (
	reposFileName = "repos.json"
	reposDirName  = "repos"
)

var (
	// ErrNamespaceExists is returned when a namespace already exists.
	ErrNamespaceExists = errors.New("namespace already exists")
	// ErrRepoNotFound is returned when a repository is not found.
	ErrRepoNotFound = errors.New("repository not found")
	// ErrInvalidURL is returned when the URL format is invalid.
	ErrInvalidURL = errors.New("invalid repository URL format")
)

// ghURLRegex matches gh:owner/repo format.
var ghURLRegex = regexp.MustCompile(`^gh:([a-zA-Z0-9_-]+)/([a-zA-Z0-9_.-]+)$`)

// Store manages repository registrations.
type Store struct {
	baseDir string
}

// NewStore creates a new repository store.
func NewStore(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// expandDir expands ~ to home directory.
func (s *Store) expandDir() (string, error) {
	dir := s.baseDir
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, dir[2:])
	}
	return dir, nil
}

// reposDir returns the repos directory path.
func (s *Store) reposDir() (string, error) {
	base, err := s.expandDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, reposDirName), nil
}

// reposFilePath returns the path to repos.json.
func (s *Store) reposFilePath() (string, error) {
	base, err := s.expandDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, reposFileName), nil
}

// RepoLocalPath returns the local path for a repository.
func (s *Store) RepoLocalPath(namespace string) (string, error) {
	reposDir, err := s.reposDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(reposDir, namespace), nil
}

// load loads the repos file.
func (s *Store) load() (*ReposFile, error) {
	path, err := s.reposFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ReposFile{Version: 1, Repos: []RepoConfig{}}, nil
		}
		return nil, err
	}

	var repos ReposFile
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("parse repos.json: %w", err)
	}

	return &repos, nil
}

// save saves the repos file.
func (s *Store) save(repos *ReposFile) error {
	baseDir, err := s.expandDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	path, err := s.reposFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal repos.json: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write repos.json: %w", err)
	}

	return nil
}

// ParseURL parses a gh:owner/repo URL.
func ParseURL(url string) (owner, repo string, err error) {
	matches := ghURLRegex.FindStringSubmatch(url)
	if matches == nil {
		return "", "", ErrInvalidURL
	}
	return matches[1], matches[2], nil
}

// GenerateNamespace generates a namespace from owner and repo.
// Format: first 4 chars of owner + "-" + first 4 chars of repo
func GenerateNamespace(owner, repo string) string {
	ownerPart := owner
	if len(ownerPart) > 4 {
		ownerPart = ownerPart[:4]
	}

	repoPart := repo
	if len(repoPart) > 4 {
		repoPart = repoPart[:4]
	}

	return strings.ToLower(ownerPart + "-" + repoPart)
}

// fetchGitHubDescription fetches the repository description from GitHub API.
func fetchGitHubDescription(owner, repo string) string {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	return result.Description
}

// Add adds a new repository by cloning it locally.
func (s *Store) Add(url, namespace string) (*RepoConfig, error) {
	// Ensure git is installed
	if err := git.EnsureInstalled(); err != nil {
		return nil, err
	}

	owner, repo, err := ParseURL(url)
	if err != nil {
		return nil, err
	}

	// Generate namespace if not provided
	if namespace == "" {
		namespace = GenerateNamespace(owner, repo)
	}

	// Load existing repos
	repos, err := s.load()
	if err != nil {
		return nil, err
	}

	// Check for namespace conflict
	for _, r := range repos.Repos {
		if r.Namespace == namespace {
			return nil, ErrNamespaceExists
		}
	}

	// Create repos directory
	reposDir, err := s.reposDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return nil, fmt.Errorf("create repos directory: %w", err)
	}

	// Clone repository
	localPath := filepath.Join(reposDir, namespace)
	gitURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	fmt.Printf("Cloning %s...\n", gitURL)
	if err := git.Clone(gitURL, localPath); err != nil {
		return nil, fmt.Errorf("clone repository: %w", err)
	}

	// Get default branch
	defaultBranch, err := git.GetDefaultBranch(localPath)
	if err != nil {
		defaultBranch = "main" // fallback
	}

	// Fetch description from GitHub API
	description := fetchGitHubDescription(owner, repo)

	config := RepoConfig{
		Namespace:     namespace,
		URL:           fmt.Sprintf("https://github.com/%s/%s", owner, repo),
		Owner:         owner,
		Repo:          repo,
		DefaultBranch: defaultBranch,
		Description:   description,
		AddedAt:       time.Now().UTC(),
	}

	repos.Repos = append(repos.Repos, config)

	if err := s.save(repos); err != nil {
		// Clean up cloned repo on save failure
		_ = os.RemoveAll(localPath)
		return nil, err
	}

	return &config, nil
}

// List returns all registered repositories.
func (s *Store) List() ([]RepoConfig, error) {
	repos, err := s.load()
	if err != nil {
		return nil, err
	}
	return repos.Repos, nil
}

// Get returns a repository by namespace.
func (s *Store) Get(namespace string) (*RepoConfig, error) {
	repos, err := s.load()
	if err != nil {
		return nil, err
	}

	for _, r := range repos.Repos {
		if r.Namespace == namespace {
			return &r, nil
		}
	}

	return nil, ErrRepoNotFound
}

// Remove removes a repository by namespace.
func (s *Store) Remove(namespace string) error {
	repos, err := s.load()
	if err != nil {
		return err
	}

	found := false
	newRepos := make([]RepoConfig, 0, len(repos.Repos))
	for _, r := range repos.Repos {
		if r.Namespace == namespace {
			found = true
			continue
		}
		newRepos = append(newRepos, r)
	}

	if !found {
		return ErrRepoNotFound
	}

	// Remove local clone
	localPath, err := s.RepoLocalPath(namespace)
	if err == nil {
		_ = os.RemoveAll(localPath)
	}

	repos.Repos = newRepos
	return s.save(repos)
}

// NamespaceExists checks if a namespace already exists.
func (s *Store) NamespaceExists(namespace string) (bool, error) {
	repos, err := s.load()
	if err != nil {
		return false, err
	}

	for _, r := range repos.Repos {
		if r.Namespace == namespace {
			return true, nil
		}
	}

	return false, nil
}

// refreshDescription updates the description for a repository if missing.
func (s *Store) refreshDescription(namespace string) error {
	repos, err := s.load()
	if err != nil {
		return err
	}

	for i, r := range repos.Repos {
		if r.Namespace == namespace {
			if r.Description == "" {
				desc := fetchGitHubDescription(r.Owner, r.Repo)
				if desc != "" {
					repos.Repos[i].Description = desc
					return s.save(repos)
				}
			}
			return nil
		}
	}

	return ErrRepoNotFound
}

// Update pulls the latest changes for a repository.
func (s *Store) Update(namespace string) error {
	if err := git.EnsureInstalled(); err != nil {
		return err
	}

	localPath, err := s.RepoLocalPath(namespace)
	if err != nil {
		return err
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return ErrRepoNotFound
	}

	if err := git.Pull(localPath); err != nil {
		return err
	}

	// Update description if missing
	return s.refreshDescription(namespace)
}

// UpdateAll pulls the latest changes for all repositories.
func (s *Store) UpdateAll() error {
	if err := git.EnsureInstalled(); err != nil {
		return err
	}

	repos, err := s.List()
	if err != nil {
		return err
	}

	for _, r := range repos {
		localPath, err := s.RepoLocalPath(r.Namespace)
		if err != nil {
			continue
		}
		fmt.Printf("Updating %s...\n", r.Namespace)
		if err := git.PullQuiet(localPath); err != nil {
			fmt.Printf("  Warning: failed to update %s: %v\n", r.Namespace, err)
		}
		// Refresh description if missing
		_ = s.refreshDescription(r.Namespace)
	}

	return nil
}

// Browse browses a repository for packages from local clone.
func (s *Store) Browse(namespace string, typeFilter PackageType) ([]BrowseItem, error) {
	localPath, err := s.RepoLocalPath(namespace)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return nil, ErrRepoNotFound
	}

	var items []BrowseItem

	// Scan skills directory
	if typeFilter == "" || typeFilter == TypeSkill {
		skillItems, _ := s.scanSkills(localPath)
		items = append(items, skillItems...)
	}

	// Scan commands directory
	if typeFilter == "" || typeFilter == TypeCommand {
		cmdItems, _ := s.scanCommands(localPath)
		items = append(items, cmdItems...)
	}

	// Scan agents directory
	if typeFilter == "" || typeFilter == TypeAgent {
		agentItems, _ := s.scanAgents(localPath)
		items = append(items, agentItems...)
	}

	// Scan hooks directory
	if typeFilter == "" || typeFilter == TypeHook {
		hookItems, _ := s.scanHooks(localPath)
		items = append(items, hookItems...)
	}

	return items, nil
}

// scanSkills scans the skills directories for skill packages.
// It checks both root-level skills/ and .claude/skills/ directories.
func (s *Store) scanSkills(repoPath string) ([]BrowseItem, error) {
	var items []BrowseItem

	// Directories to scan: root-level and .claude/ subdirectory
	scanDirs := []struct {
		dir    string
		prefix string
	}{
		{filepath.Join(repoPath, "skills"), "skills/"},
		{filepath.Join(repoPath, ".claude", "skills"), ".claude/skills/"},
	}

	for _, sd := range scanDirs {
		entries, err := os.ReadDir(sd.dir)
		if err != nil {
			continue // Directory doesn't exist, skip
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			skillDir := filepath.Join(sd.dir, entry.Name())
			// Check for SKILL.md or skill.md
			for _, name := range []string{"SKILL.md", "skill.md"} {
				if _, err := os.Stat(filepath.Join(skillDir, name)); err == nil {
					items = append(items, BrowseItem{
						Name: entry.Name(),
						Path: sd.prefix + entry.Name(),
						Type: TypeSkill,
					})
					break
				}
			}
		}
	}

	return items, nil
}

// scanCommands scans the commands directories for command packages.
// It checks both root-level commands/ and .claude/commands/ directories.
// Supports one level of nesting (e.g., commands/game/init.md -> game:init).
func (s *Store) scanCommands(repoPath string) ([]BrowseItem, error) {
	var items []BrowseItem

	// Directories to scan: root-level and .claude/ subdirectory
	scanDirs := []struct {
		dir    string
		prefix string
	}{
		{filepath.Join(repoPath, "commands"), "commands/"},
		{filepath.Join(repoPath, ".claude", "commands"), ".claude/commands/"},
	}

	for _, sd := range scanDirs {
		entries, err := os.ReadDir(sd.dir)
		if err != nil {
			continue // Directory doesn't exist, skip
		}

		for _, entry := range entries {
			if entry.IsDir() {
				// Scan one level of subdirectory
				subItems := s.scanCommandsSubdir(sd.dir, sd.prefix, entry.Name())
				items = append(items, subItems...)
				continue
			}
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			name := strings.TrimSuffix(entry.Name(), ".md")
			items = append(items, BrowseItem{
				Name: name,
				Path: sd.prefix + entry.Name(),
				Type: TypeCommand,
			})
		}
	}

	return items, nil
}

// scanCommandsSubdir scans a subdirectory within commands for .md files.
// Files are named as "subdir:filename" (e.g., game:init for game/init.md).
func (s *Store) scanCommandsSubdir(parentDir, prefix, subdir string) []BrowseItem {
	var items []BrowseItem

	subdirPath := filepath.Join(parentDir, subdir)
	entries, err := os.ReadDir(subdirPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Only one level of nesting supported
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		baseName := strings.TrimSuffix(entry.Name(), ".md")
		items = append(items, BrowseItem{
			Name: subdir + ":" + baseName,
			Path: prefix + subdir + "/" + entry.Name(),
			Type: TypeCommand,
		})
	}

	return items
}

// scanAgents scans the agents directories for agent packages.
// It checks both root-level agents/ and .claude/agents/ directories.
// Supports one level of nesting (e.g., agents/dev/tester.md -> dev:tester).
func (s *Store) scanAgents(repoPath string) ([]BrowseItem, error) {
	var items []BrowseItem

	// Directories to scan: root-level and .claude/ subdirectory
	scanDirs := []struct {
		dir    string
		prefix string
	}{
		{filepath.Join(repoPath, "agents"), "agents/"},
		{filepath.Join(repoPath, ".claude", "agents"), ".claude/agents/"},
	}

	for _, sd := range scanDirs {
		entries, err := os.ReadDir(sd.dir)
		if err != nil {
			continue // Directory doesn't exist, skip
		}

		for _, entry := range entries {
			if entry.IsDir() {
				// Scan one level of subdirectory
				subItems := s.scanAgentsSubdir(sd.dir, sd.prefix, entry.Name())
				items = append(items, subItems...)
				continue
			}
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			name := strings.TrimSuffix(entry.Name(), ".md")
			items = append(items, BrowseItem{
				Name: name,
				Path: sd.prefix + entry.Name(),
				Type: TypeAgent,
			})
		}
	}

	return items, nil
}

// scanAgentsSubdir scans a subdirectory within agents for .md files.
// Files are named as "subdir:filename" (e.g., dev:tester for dev/tester.md).
func (s *Store) scanAgentsSubdir(parentDir, prefix, subdir string) []BrowseItem {
	var items []BrowseItem

	subdirPath := filepath.Join(parentDir, subdir)
	entries, err := os.ReadDir(subdirPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Only one level of nesting supported
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		baseName := strings.TrimSuffix(entry.Name(), ".md")
		items = append(items, BrowseItem{
			Name: subdir + ":" + baseName,
			Path: prefix + subdir + "/" + entry.Name(),
			Type: TypeAgent,
		})
	}

	return items
}

// scanHooks scans the hooks directories for hook packages.
// It checks both root-level hooks/ and .claude/hooks/ directories.
func (s *Store) scanHooks(repoPath string) ([]BrowseItem, error) {
	var items []BrowseItem

	// Directories to scan: root-level and .claude/ subdirectory
	scanDirs := []struct {
		dir    string
		prefix string
	}{
		{filepath.Join(repoPath, "hooks"), "hooks/"},
		{filepath.Join(repoPath, ".claude", "hooks"), ".claude/hooks/"},
	}

	for _, sd := range scanDirs {
		entries, err := os.ReadDir(sd.dir)
		if err != nil {
			continue // Directory doesn't exist, skip
		}

		for _, entry := range entries {
			// Hooks can be shell scripts or other executable files
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			items = append(items, BrowseItem{
				Name: name,
				Path: sd.prefix + name,
				Type: TypeHook,
			})
		}
	}

	return items, nil
}

// Search searches for packages across all registered repositories.
func (s *Store) Search(query string) (map[string][]BrowseItem, error) {
	repos, err := s.List()
	if err != nil {
		return nil, err
	}

	results := make(map[string][]BrowseItem)
	query = strings.ToLower(query)

	for _, r := range repos {
		items, err := s.Browse(r.Namespace, "")
		if err != nil {
			continue // Skip repos that fail
		}

		var matches []BrowseItem
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Name), query) {
				matches = append(matches, item)
			}
		}

		if len(matches) > 0 {
			results[r.Namespace] = matches
		}
	}

	return results, nil
}
