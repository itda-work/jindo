package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const historyDir = ".history"

// Version represents a single version in history
type Version struct {
	Number    int       `json:"number"`
	Timestamp time.Time `json:"timestamp"`
	Filename  string    `json:"filename"`
}

// Manifest represents the history manifest for an agent
type Manifest struct {
	AgentID  string    `json:"agent_id"`
	Versions []Version `json:"versions"`
}

// HistoryManager manages version history for an agent
type HistoryManager struct {
	agentsDir string
	agentID   string
}

// NewHistoryManager creates a new history manager for an agent
// agentsDir is the agents directory (e.g., ~/.claude/agents)
// agentID is the agent name without .md extension
func NewHistoryManager(agentsDir, agentID string) *HistoryManager {
	return &HistoryManager{
		agentsDir: agentsDir,
		agentID:   agentID,
	}
}

// getHistoryDir returns the .history directory path for this agent
func (h *HistoryManager) getHistoryDir() string {
	return filepath.Join(h.agentsDir, historyDir, h.agentID)
}

// getManifestPath returns the manifest.json path
func (h *HistoryManager) getManifestPath() string {
	return filepath.Join(h.getHistoryDir(), "manifest.json")
}

// ensureHistoryDir creates the .history/agent-id directory if it doesn't exist
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
				AgentID:  h.agentID,
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

// SaveVersion saves the current agent content as a new version
func (h *HistoryManager) SaveVersion(content string) (*Version, error) {
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
	filename := fmt.Sprintf("v%03d-%s.md", nextNum, timestamp)

	// Save version file
	if err := h.ensureHistoryDir(); err != nil {
		return nil, err
	}

	versionPath := filepath.Join(h.getHistoryDir(), filename)
	if err := os.WriteFile(versionPath, []byte(content), 0644); err != nil {
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

// GetVersion retrieves a specific version's content
func (h *HistoryManager) GetVersion(versionNum int) (string, *Version, error) {
	manifest, err := h.loadManifest()
	if err != nil {
		return "", nil, err
	}

	for _, v := range manifest.Versions {
		if v.Number == versionNum {
			path := filepath.Join(h.getHistoryDir(), v.Filename)
			content, err := os.ReadFile(path)
			if err != nil {
				return "", nil, err
			}
			return string(content), &v, nil
		}
	}

	return "", nil, fmt.Errorf("version %d not found", versionNum)
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
