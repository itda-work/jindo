package repo

import "time"

// RepoConfig represents a registered repository.
type RepoConfig struct {
	Namespace     string    `json:"namespace"`
	URL           string    `json:"url"`
	Owner         string    `json:"owner"`
	Repo          string    `json:"repo"`
	DefaultBranch string    `json:"default_branch"`
	AddedAt       time.Time `json:"added_at"`
}

// ReposFile represents the repos.json file structure.
type ReposFile struct {
	Version int          `json:"version"`
	Repos   []RepoConfig `json:"repos"`
}

// PackageType represents the type of Claude Code package.
type PackageType string

const (
	TypeSkill   PackageType = "skill"
	TypeCommand PackageType = "command"
	TypeAgent   PackageType = "agent"
)

// BrowseItem represents an item found during browsing.
type BrowseItem struct {
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Type        PackageType `json:"type"`
	Description string      `json:"description,omitempty"`
}
