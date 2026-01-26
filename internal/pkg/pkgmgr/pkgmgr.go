package pkgmgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/itda-skills/jindo/internal/pkg/git"
	"github.com/itda-skills/jindo/internal/pkg/repo"
)

const (
	installedFileName = "installed.json"
	namespaceSep      = "--"
)

var (
	// ErrPackageNotFound is returned when a package is not found.
	ErrPackageNotFound = errors.New("package not found")
	// ErrPackageAlreadyInstalled is returned when trying to install an already installed package.
	ErrPackageAlreadyInstalled = errors.New("package already installed")
	// ErrInvalidSpec is returned when the install spec is invalid.
	ErrInvalidSpec = errors.New("invalid package specification")
)

// installSpecRegex matches namespace:path[@version] format.
var installSpecRegex = regexp.MustCompile(`^([a-z0-9-]+):(.+?)(?:@(.+))?$`)

// Manager manages installed packages.
type Manager struct {
	baseDir   string          // ~/.itda-skills (for metadata: installed.json, repos)
	claudeDir string          // ~/.claude (for actual installed files)
	repoStore *repo.Store
}

// NewManager creates a new package manager.
func NewManager(baseDir string) *Manager {
	return &Manager{
		baseDir:   baseDir,
		claudeDir: "~/.claude",
		repoStore: repo.NewStore(baseDir),
	}
}

// expandDir expands ~ to home directory for baseDir.
func (m *Manager) expandDir() (string, error) {
	return expandPath(m.baseDir)
}

// expandClaudeDir expands ~ to home directory for claudeDir.
func (m *Manager) expandClaudeDir() (string, error) {
	return expandPath(m.claudeDir)
}

// expandPath expands ~ to home directory.
func expandPath(dir string) (string, error) {
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, dir[2:])
	}
	return dir, nil
}

// installedFilePath returns the path to installed.json.
func (m *Manager) installedFilePath() (string, error) {
	base, err := m.expandDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, installedFileName), nil
}

// load loads the installed packages file.
func (m *Manager) load() (*InstalledFile2, error) {
	path, err := m.installedFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &InstalledFile2{Version: 1, Packages: []InstalledPackage{}}, nil
		}
		return nil, err
	}

	var installed InstalledFile2
	if err := json.Unmarshal(data, &installed); err != nil {
		return nil, fmt.Errorf("parse installed.json: %w", err)
	}

	return &installed, nil
}

// save saves the installed packages file.
func (m *Manager) save(installed *InstalledFile2) error {
	baseDir, err := m.expandDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	path, err := m.installedFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(installed, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal installed.json: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write installed.json: %w", err)
	}

	return nil
}

// ParseSpec parses an install specification (namespace:path[@version]).
func ParseSpec(spec string) (*InstallSpec, error) {
	matches := installSpecRegex.FindStringSubmatch(spec)
	if matches == nil {
		return nil, ErrInvalidSpec
	}

	return &InstallSpec{
		Namespace: matches[1],
		Path:      matches[2],
		Version:   matches[3], // May be empty
	}, nil
}

// MakeNamespacedName creates a namespaced name.
func MakeNamespacedName(namespace, name string) string {
	return namespace + namespaceSep + name
}

// ParseNamespacedName parses a namespaced name.
func ParseNamespacedName(name string) (namespace, originalName string) {
	parts := strings.SplitN(name, namespaceSep, 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", name
}

// determinePackageType determines the package type from the path.
func determinePackageType(path string) repo.PackageType {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	switch parts[0] {
	case "skills":
		return repo.TypeSkill
	case "commands":
		return repo.TypeCommand
	case "agents":
		return repo.TypeAgent
	case "hooks":
		return repo.TypeHook
	default:
		return ""
	}
}

// extractPackageName extracts the package name from the path.
func extractPackageName(path string, pkgType repo.PackageType) string {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}

	switch pkgType {
	case repo.TypeSkill:
		// skills/<name>/...
		return parts[1]
	case repo.TypeCommand, repo.TypeAgent:
		// commands/<name>.md or agents/<name>.md
		name := parts[1]
		return strings.TrimSuffix(name, ".md")
	case repo.TypeHook:
		// hooks/<name>
		return parts[1]
	default:
		return ""
	}
}

// Install installs a package from local repository clone.
func (m *Manager) Install(specStr string) (*InstalledPackage, error) {
	spec, err := ParseSpec(specStr)
	if err != nil {
		return nil, err
	}

	// Get repository info and local path
	repoConfig, err := m.repoStore.Get(spec.Namespace)
	if err != nil {
		return nil, fmt.Errorf("repository not found: %w", err)
	}

	repoLocalPath, err := m.repoStore.RepoLocalPath(spec.Namespace)
	if err != nil {
		return nil, err
	}

	// Determine package type and name
	pkgType := determinePackageType(spec.Path)
	if pkgType == "" {
		return nil, fmt.Errorf("cannot determine package type from path: %s", spec.Path)
	}

	originalName := extractPackageName(spec.Path, pkgType)
	if originalName == "" {
		return nil, fmt.Errorf("cannot extract package name from path: %s", spec.Path)
	}

	namespacedName := MakeNamespacedName(spec.Namespace, originalName)

	// Check if already installed
	installed, err := m.load()
	if err != nil {
		return nil, err
	}

	for _, pkg := range installed.Packages {
		if pkg.Name == namespacedName {
			return nil, ErrPackageAlreadyInstalled
		}
	}

	// Get current commit SHA from local repo
	currentSHA, err := git.GetCurrentCommit(repoLocalPath)
	if err != nil {
		currentSHA = "unknown"
	}

	// Install files to ~/.claude directory
	claudeDir, err := m.expandClaudeDir()
	if err != nil {
		return nil, err
	}

	var files []InstalledFile

	switch pkgType {
	case repo.TypeSkill:
		files, err = m.installSkill(repoLocalPath, spec.Path, namespacedName, claudeDir)
	case repo.TypeCommand:
		files, err = m.installCommand(repoLocalPath, spec.Path, namespacedName, claudeDir)
	case repo.TypeAgent:
		files, err = m.installAgent(repoLocalPath, spec.Path, namespacedName, claudeDir)
	case repo.TypeHook:
		files, err = m.installHook(repoLocalPath, spec.Path, namespacedName, claudeDir)
	}

	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	pkg := InstalledPackage{
		Name:         namespacedName,
		OriginalName: originalName,
		Type:         pkgType,
		Namespace:    spec.Namespace,
		SourcePath:   spec.Path,
		Version: VersionInfo{
			Type: "commit",
			SHA:  currentSHA,
			Ref:  repoConfig.DefaultBranch,
		},
		Files:       files,
		InstalledAt: now,
		UpdatedAt:   now,
	}

	installed.Packages = append(installed.Packages, pkg)

	if err := m.save(installed); err != nil {
		// Try to clean up installed files
		for _, f := range files {
			_ = os.RemoveAll(f.Target)
		}
		return nil, err
	}

	return &pkg, nil
}

// installSkill installs a skill package from local clone.
func (m *Manager) installSkill(repoLocalPath, path, namespacedName, baseDir string) ([]InstalledFile, error) {
	srcDir := filepath.Join(repoLocalPath, path)
	destDir := filepath.Join(baseDir, "skills", namespacedName)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("create skill directory: %w", err)
	}

	var files []InstalledFile

	err := filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Copy file
		if err := copyFile(srcPath, destPath); err != nil {
			return err
		}

		files = append(files, InstalledFile{
			Source: filepath.Join(path, relPath),
			Target: destPath,
			SHA:    "", // Not tracking file SHA for local copies
		})

		return nil
	})

	if err != nil {
		_ = os.RemoveAll(destDir)
		return nil, fmt.Errorf("copy skill files: %w", err)
	}

	if len(files) == 0 {
		_ = os.RemoveAll(destDir)
		return nil, fmt.Errorf("no files found in skill: %s", path)
	}

	return files, nil
}

// installCommand installs a command package from local clone.
func (m *Manager) installCommand(repoLocalPath, path, namespacedName, baseDir string) ([]InstalledFile, error) {
	srcPath := filepath.Join(repoLocalPath, path)
	commandsDir := filepath.Join(baseDir, "commands")

	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return nil, fmt.Errorf("create commands directory: %w", err)
	}

	destPath := filepath.Join(commandsDir, namespacedName+".md")
	if err := copyFile(srcPath, destPath); err != nil {
		return nil, fmt.Errorf("copy command file: %w", err)
	}

	return []InstalledFile{{
		Source: path,
		Target: destPath,
		SHA:    "",
	}}, nil
}

// installAgent installs an agent package from local clone.
func (m *Manager) installAgent(repoLocalPath, path, namespacedName, baseDir string) ([]InstalledFile, error) {
	srcPath := filepath.Join(repoLocalPath, path)
	agentsDir := filepath.Join(baseDir, "agents")

	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return nil, fmt.Errorf("create agents directory: %w", err)
	}

	destPath := filepath.Join(agentsDir, namespacedName+".md")
	if err := copyFile(srcPath, destPath); err != nil {
		return nil, fmt.Errorf("copy agent file: %w", err)
	}

	return []InstalledFile{{
		Source: path,
		Target: destPath,
		SHA:    "",
	}}, nil
}

// installHook installs a hook package from local clone.
func (m *Manager) installHook(repoLocalPath, path, namespacedName, baseDir string) ([]InstalledFile, error) {
	srcPath := filepath.Join(repoLocalPath, path)
	hooksDir := filepath.Join(baseDir, "hooks")

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return nil, fmt.Errorf("create hooks directory: %w", err)
	}

	// Get original filename (including extension)
	originalName := filepath.Base(path)
	// Prepend namespace
	destName := namespacedName
	if ext := filepath.Ext(originalName); ext != "" {
		destName += ext
	}

	destPath := filepath.Join(hooksDir, destName)
	if err := copyFile(srcPath, destPath); err != nil {
		return nil, fmt.Errorf("copy hook file: %w", err)
	}

	// Make hook executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return nil, fmt.Errorf("make hook executable: %w", err)
	}

	return []InstalledFile{{
		Source: path,
		Target: destPath,
		SHA:    "",
	}}, nil
}

// copyFile copies a file from src to dest.
func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// Uninstall removes an installed package.
func (m *Manager) Uninstall(name string) error {
	installed, err := m.load()
	if err != nil {
		return err
	}

	var pkg *InstalledPackage
	var idx int
	for i, p := range installed.Packages {
		if p.Name == name {
			pkg = &installed.Packages[i]
			idx = i
			break
		}
	}

	if pkg == nil {
		return ErrPackageNotFound
	}

	// Remove files
	for _, f := range pkg.Files {
		_ = os.Remove(f.Target)
	}

	// For skills, remove the directory
	if pkg.Type == repo.TypeSkill {
		baseDir, err := m.expandDir()
		if err == nil {
			skillDir := filepath.Join(baseDir, "skills", pkg.Name)
			_ = os.RemoveAll(skillDir)
		}
	}

	// Remove from installed list
	installed.Packages = append(installed.Packages[:idx], installed.Packages[idx+1:]...)

	return m.save(installed)
}

// List returns all installed packages.
func (m *Manager) List() ([]InstalledPackage, error) {
	installed, err := m.load()
	if err != nil {
		return nil, err
	}
	return installed.Packages, nil
}

// Get returns an installed package by name.
func (m *Manager) Get(name string) (*InstalledPackage, error) {
	installed, err := m.load()
	if err != nil {
		return nil, err
	}

	for _, pkg := range installed.Packages {
		if pkg.Name == name {
			return &pkg, nil
		}
	}

	return nil, ErrPackageNotFound
}

// CheckUpdates checks for updates for installed packages.
func (m *Manager) CheckUpdates(names ...string) ([]UpdateInfo, error) {
	installed, err := m.load()
	if err != nil {
		return nil, err
	}

	var results []UpdateInfo

	for _, pkg := range installed.Packages {
		// Filter by names if provided
		if len(names) > 0 {
			found := false
			for _, n := range names {
				if pkg.Name == n {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		info, err := m.checkPackageUpdate(&pkg)
		if err != nil {
			// Skip packages that fail to check
			continue
		}

		results = append(results, *info)
	}

	return results, nil
}

// checkPackageUpdate checks for updates for a single package.
func (m *Manager) checkPackageUpdate(pkg *InstalledPackage) (*UpdateInfo, error) {
	repoLocalPath, err := m.repoStore.RepoLocalPath(pkg.Namespace)
	if err != nil {
		return nil, err
	}

	repoConfig, err := m.repoStore.Get(pkg.Namespace)
	if err != nil {
		return nil, err
	}

	// Fetch latest changes
	if err := git.Fetch(repoLocalPath); err != nil {
		return nil, err
	}

	// Get remote commit
	latestSHA, err := git.GetRemoteCommit(repoLocalPath, repoConfig.DefaultBranch)
	if err != nil {
		return nil, err
	}

	info := &UpdateInfo{
		Package:    pkg,
		CurrentSHA: pkg.Version.SHA,
		LatestSHA:  latestSHA,
		HasUpdate:  pkg.Version.SHA != latestSHA,
	}

	if info.HasUpdate {
		// Get changed files
		changedFiles, err := git.ListChangedFiles(repoLocalPath, pkg.Version.SHA, "origin/"+repoConfig.DefaultBranch)
		if err == nil {
			for _, f := range changedFiles {
				if strings.HasPrefix(f, pkg.SourcePath) {
					info.ChangedFiles = append(info.ChangedFiles, f)
				}
			}
		}
	}

	return info, nil
}

// Update updates a package to the latest version.
func (m *Manager) Update(name string) (*InstalledPackage, error) {
	pkg, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	// Pull latest changes in the repo first
	repoLocalPath, err := m.repoStore.RepoLocalPath(pkg.Namespace)
	if err != nil {
		return nil, err
	}

	if err := git.Pull(repoLocalPath); err != nil {
		return nil, fmt.Errorf("pull latest changes: %w", err)
	}

	// Uninstall old version
	if err := m.Uninstall(name); err != nil {
		return nil, fmt.Errorf("uninstall old version: %w", err)
	}

	// Reinstall
	spec := fmt.Sprintf("%s:%s", pkg.Namespace, pkg.SourcePath)
	return m.Install(spec)
}

// RepoStore returns the repository store.
func (m *Manager) RepoStore() *repo.Store {
	return m.repoStore
}
