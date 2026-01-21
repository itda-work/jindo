package prompt

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed prompts/*.md
var embeddedPrompts embed.FS

// PromptInfo contains information about a prompt
type PromptInfo struct {
	Name       string // e.g., "adapt-skill"
	IsOverride bool   // true if loaded from override file
	Path       string // path to the prompt file (empty for embedded)
}

// GetOverrideDir returns the directory for override prompts
func GetOverrideDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "jindo", "prompts"), nil
}

// EnsureOverrideDir creates the override directory if it doesn't exist
func EnsureOverrideDir() (string, error) {
	dir, err := GetOverrideDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// GetOverridePath returns the override path for a prompt name
func GetOverridePath(name string) (string, error) {
	dir, err := GetOverrideDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".md"), nil
}

// Load loads a prompt by name, preferring override over embedded
// name should be without extension, e.g., "adapt-skill"
func Load(name string) (string, error) {
	// First, try to load from override directory
	overridePath, err := GetOverridePath(name)
	if err == nil {
		if content, err := os.ReadFile(overridePath); err == nil {
			return string(content), nil
		}
	}

	// Fall back to embedded prompt
	content, err := embeddedPrompts.ReadFile("prompts/" + name + ".md")
	if err != nil {
		return "", fmt.Errorf("prompt not found: %s", name)
	}

	return string(content), nil
}

// LoadInfo loads a prompt and returns its info
func LoadInfo(name string) (string, *PromptInfo, error) {
	info := &PromptInfo{Name: name}

	// First, try to load from override directory
	overridePath, err := GetOverridePath(name)
	if err == nil {
		if content, err := os.ReadFile(overridePath); err == nil {
			info.IsOverride = true
			info.Path = overridePath
			return string(content), info, nil
		}
	}

	// Fall back to embedded prompt
	content, err := embeddedPrompts.ReadFile("prompts/" + name + ".md")
	if err != nil {
		return "", nil, fmt.Errorf("prompt not found: %s", name)
	}

	return string(content), info, nil
}

// GetEmbedded returns the embedded version of a prompt
func GetEmbedded(name string) (string, error) {
	content, err := embeddedPrompts.ReadFile("prompts/" + name + ".md")
	if err != nil {
		return "", fmt.Errorf("embedded prompt not found: %s", name)
	}
	return string(content), nil
}

// SaveOverride saves an override prompt
func SaveOverride(name, content string) error {
	dir, err := EnsureOverrideDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, name+".md")
	return os.WriteFile(path, []byte(content), 0644)
}

// DeleteOverride removes an override prompt
func DeleteOverride(name string) error {
	path, err := GetOverridePath(name)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no override exists for: %s", name)
		}
		return err
	}
	return nil
}

// HasOverride checks if an override exists for a prompt
func HasOverride(name string) bool {
	path, err := GetOverridePath(name)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// List returns all available prompt names
func List() ([]PromptInfo, error) {
	entries, err := embeddedPrompts.ReadDir("prompts")
	if err != nil {
		return nil, err
	}

	var prompts []PromptInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		info := PromptInfo{
			Name:       name,
			IsOverride: HasOverride(name),
		}
		if info.IsOverride {
			info.Path, _ = GetOverridePath(name)
		}
		prompts = append(prompts, info)
	}

	return prompts, nil
}
