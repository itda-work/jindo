package cli

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrMutuallyExclusiveFlags is returned when both --global and --local flags are specified
var ErrMutuallyExclusiveFlags = errors.New("--global and --local flags are mutually exclusive")

// ValidateScopeFlags checks that --global and --local flags are not both specified
func ValidateScopeFlags(global, local bool) error {
	if global && local {
		return ErrMutuallyExclusiveFlags
	}
	return nil
}

// DefaultScope returns the default scope.
// If a local .claude directory exists in the current working directory, local scope is preferred.
func DefaultScope() PathScope {
	if LocalClaudeDirExists() {
		return ScopeLocal
	}
	return ScopeGlobal
}

// ResolveScope determines the effective scope given optional --global/--local flags.
// Default is local if .claude exists, otherwise global.
func ResolveScope(globalFlag, localFlag bool) (PathScope, error) {
	if err := ValidateScopeFlags(globalFlag, localFlag); err != nil {
		return "", err
	}
	if localFlag {
		return ScopeLocal, nil
	}
	if globalFlag {
		return ScopeGlobal, nil
	}
	return DefaultScope(), nil
}

// ScopeDescription returns a user-facing description for a scope.
func ScopeDescription(scope PathScope) string {
	switch scope {
	case ScopeLocal:
		return "local (.claude)"
	default:
		return "global (~/.claude)"
	}
}

const (
	globalClaudeDir = "~/.claude"
	localClaudeDir  = ".claude"
)

// PathScope represents the scope of a path (global or local)
type PathScope string

const (
	ScopeGlobal PathScope = "global"
	ScopeLocal  PathScope = "local"
)

// GetGlobalPath returns the global ~/.claude path
func GetGlobalPath(subdir string) string {
	return filepath.Join(globalClaudeDir, subdir)
}

// GetLocalPath returns the local .claude path (CWD-based)
// Returns empty string if local .claude directory doesn't exist
func GetLocalPath(subdir string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	localDir := filepath.Join(cwd, localClaudeDir)
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		return ""
	}

	return filepath.Join(localDir, subdir)
}

// GetLocalPathForWrite returns the local .claude path for writing
// Creates the directory if it doesn't exist
func GetLocalPathForWrite(subdir string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	localDir := filepath.Join(cwd, localClaudeDir, subdir)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return "", err
	}

	return localDir, nil
}

// LocalClaudeDirExists checks if .claude directory exists in CWD
func LocalClaudeDirExists() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	localDir := filepath.Join(cwd, localClaudeDir)
	info, err := os.Stat(localDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetPathByScope returns the appropriate path based on scope
func GetPathByScope(scope PathScope, subdir string) string {
	switch scope {
	case ScopeLocal:
		cwd, err := os.Getwd()
		if err != nil {
			return GetGlobalPath(subdir) // fallback to global
		}
		return filepath.Join(cwd, localClaudeDir, subdir)
	default:
		return GetGlobalPath(subdir)
	}
}

// GetSettingsPathByScope returns the settings.json path based on scope
func GetSettingsPathByScope(scope PathScope) string {
	switch scope {
	case ScopeLocal:
		cwd, err := os.Getwd()
		if err != nil {
			return filepath.Join(globalClaudeDir, "settings.json") // fallback to global
		}
		return filepath.Join(cwd, localClaudeDir, "settings.json")
	default:
		return filepath.Join(globalClaudeDir, "settings.json")
	}
}

// GetLocalSettingsPath returns the local settings.json path if exists
// Returns empty string if local .claude/settings.json doesn't exist
func GetLocalSettingsPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	settingsPath := filepath.Join(cwd, localClaudeDir, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return ""
	}

	return settingsPath
}
