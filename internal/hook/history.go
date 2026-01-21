package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const historySubDir = ".history/hooks"

// Version represents a single version in history
type Version struct {
	Number    int       `json:"number"`
	Timestamp time.Time `json:"timestamp"`
	Filename  string    `json:"filename"`
}

// HookSnapshot represents a saved hook configuration
type HookSnapshot struct {
	Name      string    `json:"name"`
	EventType EventType `json:"event_type"`
	Matcher   string    `json:"matcher"`
	Commands  []string  `json:"commands"`
}

// Manifest represents the history manifest for a hook
type Manifest struct {
	HookName string    `json:"hook_name"`
	Versions []Version `json:"versions"`
}

// HistoryManager manages version history for a hook
type HistoryManager struct {
	claudeDir string
	hookName  string
}

// NewHistoryManager creates a new history manager for a hook
// claudeDir is the .claude directory path (e.g., ~/.claude)
// hookName is the hook identifier (e.g., "PreToolUse-Bash-0")
func NewHistoryManager(claudeDir, hookName string) *HistoryManager {
	return &HistoryManager{
		claudeDir: claudeDir,
		hookName:  hookName,
	}
}

// sanitizeHookName converts hook name to safe filename
func sanitizeHookName(name string) string {
	// Replace characters that are problematic in filenames
	name = strings.ReplaceAll(name, "|", "-")
	name = strings.ReplaceAll(name, "*", "all")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

// getHistoryDir returns the .history/hooks/hook-name directory path
func (h *HistoryManager) getHistoryDir() string {
	return filepath.Join(h.claudeDir, historySubDir, sanitizeHookName(h.hookName))
}

// getManifestPath returns the manifest.json path
func (h *HistoryManager) getManifestPath() string {
	return filepath.Join(h.getHistoryDir(), "manifest.json")
}

// ensureHistoryDir creates the history directory if it doesn't exist
func (h *HistoryManager) ensureHistoryDir() error {
	return os.MkdirAll(h.getHistoryDir(), 0755)
}

// loadManifest loads the manifest file
func (h *HistoryManager) loadManifest() (*Manifest, error) {
	path := h.getManifestPath()
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{
				HookName: h.hookName,
				Versions: []Version{},
			}, nil
		}
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// saveManifest saves the manifest file
func (h *HistoryManager) saveManifest(manifest *Manifest) error {
	if err := h.ensureHistoryDir(); err != nil {
		return err
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(h.getManifestPath(), content, 0644)
}

// SaveVersion saves the current hook configuration as a new version
func (h *HistoryManager) SaveVersion(hook *Hook) (*Version, error) {
	manifest, err := h.loadManifest()
	if err != nil {
		return nil, err
	}

	// Determine next version number
	nextNum := 1
	if len(manifest.Versions) > 0 {
		nextNum = manifest.Versions[len(manifest.Versions)-1].Number + 1
	}

	// Create version filename
	now := time.Now()
	timestamp := now.Format("2006-01-02T15-04-05")
	filename := fmt.Sprintf("v%03d-%s.json", nextNum, timestamp)

	// Save version file
	if err := h.ensureHistoryDir(); err != nil {
		return nil, err
	}

	snapshot := HookSnapshot{
		Name:      hook.Name,
		EventType: hook.EventType,
		Matcher:   hook.Matcher,
		Commands:  hook.Commands,
	}

	content, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, err
	}

	versionPath := filepath.Join(h.getHistoryDir(), filename)
	if err := os.WriteFile(versionPath, content, 0644); err != nil {
		return nil, err
	}

	// Update manifest
	version := Version{
		Number:    nextNum,
		Timestamp: now,
		Filename:  filename,
	}
	manifest.Versions = append(manifest.Versions, version)

	if err := h.saveManifest(manifest); err != nil {
		return nil, err
	}

	return &version, nil
}

// ListVersions returns all versions sorted by number (newest first)
func (h *HistoryManager) ListVersions() ([]Version, error) {
	manifest, err := h.loadManifest()
	if err != nil {
		return nil, err
	}

	// Sort by number descending (newest first)
	versions := make([]Version, len(manifest.Versions))
	copy(versions, manifest.Versions)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Number > versions[j].Number
	})

	return versions, nil
}

// GetVersion retrieves a specific version's snapshot
func (h *HistoryManager) GetVersion(versionNum int) (*HookSnapshot, *Version, error) {
	manifest, err := h.loadManifest()
	if err != nil {
		return nil, nil, err
	}

	for _, v := range manifest.Versions {
		if v.Number == versionNum {
			path := filepath.Join(h.getHistoryDir(), v.Filename)
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, err
			}

			var snapshot HookSnapshot
			if err := json.Unmarshal(content, &snapshot); err != nil {
				return nil, nil, err
			}
			return &snapshot, &v, nil
		}
	}

	return nil, nil, fmt.Errorf("version %d not found", versionNum)
}

// GetLatestVersion returns the most recent version
func (h *HistoryManager) GetLatestVersion() (*Version, error) {
	manifest, err := h.loadManifest()
	if err != nil {
		return nil, err
	}

	if len(manifest.Versions) == 0 {
		return nil, fmt.Errorf("no versions found")
	}

	return &manifest.Versions[len(manifest.Versions)-1], nil
}

// HasHistory checks if any history exists
func (h *HistoryManager) HasHistory() bool {
	manifest, err := h.loadManifest()
	if err != nil {
		return false
	}
	return len(manifest.Versions) > 0
}

// FormatVersionName formats a version for display
func FormatVersionName(v *Version) string {
	return fmt.Sprintf("v%03d (%s)", v.Number, v.Timestamp.Format("2006-01-02 15:04:05"))
}

// ParseVersionArg parses a version argument (number or "latest")
func ParseVersionArg(arg string) (int, error) {
	if arg == "" || strings.ToLower(arg) == "latest" {
		return -1, nil // -1 indicates latest
	}

	// Remove 'v' prefix if present
	arg = strings.TrimPrefix(strings.ToLower(arg), "v")

	var num int
	_, err := fmt.Sscanf(arg, "%d", &num)
	if err != nil {
		return 0, fmt.Errorf("invalid version: %s", arg)
	}
	return num, nil
}
