package pkgmgr

import (
	"time"

	"github.com/itda-work/itda-jindo/internal/pkg/repo"
)

// InstalledFile represents an installed file.
type InstalledFile struct {
	Source string `json:"source"` // Source path in repository
	Target string `json:"target"` // Target path on filesystem
	SHA    string `json:"sha"`    // Content SHA for change detection
}

// VersionInfo represents version information for an installed package.
type VersionInfo struct {
	Type string `json:"type"` // "commit" or "tag"
	SHA  string `json:"sha"`
	Ref  string `json:"ref"` // branch name or tag name
}

// InstalledPackage represents an installed package.
type InstalledPackage struct {
	Name         string          `json:"name"`          // Full name with namespace (e.g., affa-ever--web-fetch)
	OriginalName string          `json:"original_name"` // Original name without namespace
	Type         repo.PackageType `json:"type"`         // skill, command, agent
	Namespace    string          `json:"namespace"`     // Repository namespace
	SourcePath   string          `json:"source_path"`   // Path in source repository
	Version      VersionInfo     `json:"version"`
	Files        []InstalledFile `json:"files"`
	InstalledAt  time.Time       `json:"installed_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// InstalledFile represents the installed.json file structure.
type InstalledFile2 struct {
	Version  int                `json:"version"`
	Packages []InstalledPackage `json:"packages"`
}

// InstallSpec represents a package installation specification.
type InstallSpec struct {
	Namespace string // Repository namespace
	Path      string // Path in repository (e.g., skills/web-fetch)
	Version   string // Optional: tag or commit SHA
}

// UpdateInfo represents update information for a package.
type UpdateInfo struct {
	Package     *InstalledPackage
	CurrentSHA  string
	LatestSHA   string
	HasUpdate   bool
	ChangedFiles []string
}
