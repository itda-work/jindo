package skill

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill represents a Claude Code skill
type Skill struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	AllowedTools []string `json:"allowed_tools"`
	Path         string   `json:"path"`
}

// skillFrontmatter represents the YAML frontmatter structure
type skillFrontmatter struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	AllowedTools string `yaml:"allowed-tools"`
}

// extractFrontmatter extracts YAML frontmatter from markdown content
func extractFrontmatter(content string) (string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", false
	}

	var frontmatterLines []string
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(frontmatterLines, "\n"), true
		}
		frontmatterLines = append(frontmatterLines, lines[i])
	}

	return "", false
}

// parseSimpleFrontmatter parses frontmatter using simple line-based approach
// This is a fallback for when YAML parsing fails due to special characters
func parseSimpleFrontmatter(frontmatter string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(frontmatter, "\n")

	for _, line := range lines {
		// Find first colon
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Only capture simple keys we care about
		switch key {
		case "name", "description", "allowed-tools":
			result[key] = value
		}
	}

	return result
}

// ParseSkillFile parses a SKILL.md or skill.md file and returns a Skill
func ParseSkillFile(path string) (*Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	skill := &Skill{
		Path: path,
	}

	frontmatter, found := extractFrontmatter(string(content))
	if !found || frontmatter == "" {
		return skill, nil
	}

	var fm skillFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		// If YAML parsing fails, fall back to simple parsing
		simple := parseSimpleFrontmatter(frontmatter)
		skill.Name = simple["name"]
		skill.Description = simple["description"]
		if allowedTools := simple["allowed-tools"]; allowedTools != "" {
			tools := strings.Split(allowedTools, ",")
			for _, tool := range tools {
				tool = strings.TrimSpace(tool)
				if tool != "" {
					skill.AllowedTools = append(skill.AllowedTools, tool)
				}
			}
		}
		return skill, nil
	}

	skill.Name = fm.Name
	skill.Description = fm.Description

	// Parse comma-separated tools
	if fm.AllowedTools != "" {
		tools := strings.Split(fm.AllowedTools, ",")
		for _, tool := range tools {
			tool = strings.TrimSpace(tool)
			if tool != "" {
				skill.AllowedTools = append(skill.AllowedTools, tool)
			}
		}
	}

	return skill, nil
}

// Store manages skills in a directory
type Store struct {
	baseDir string
}

// NewStore creates a new skill store
func NewStore(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// expandDir expands ~ to home directory
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

// findSkillFile finds the actual skill file in a directory
// Returns the full path with the actual filename (handles case-insensitive filesystems)
func findSkillFile(skillDir string) (string, error) {
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		if lower == "skill.md" {
			return filepath.Join(skillDir, name), nil
		}
	}

	return "", os.ErrNotExist
}

// Get retrieves a specific skill by name
func (s *Store) Get(name string) (*Skill, error) {
	dir, err := s.expandDir()
	if err != nil {
		return nil, err
	}

	skillDir := filepath.Join(dir, name)

	// Find the actual skill file (handles case-insensitive filesystems)
	skillFile, err := findSkillFile(skillDir)
	if err != nil {
		return nil, os.ErrNotExist
	}

	skill, err := ParseSkillFile(skillFile)
	if err != nil {
		return nil, err
	}

	if skill.Name == "" {
		skill.Name = name
	}

	return skill, nil
}

// GetContent retrieves the full content of a skill file
func (s *Store) GetContent(name string) (string, error) {
	dir, err := s.expandDir()
	if err != nil {
		return "", err
	}

	skillDir := filepath.Join(dir, name)

	// Find the actual skill file (handles case-insensitive filesystems)
	skillFile, err := findSkillFile(skillDir)
	if err != nil {
		return "", os.ErrNotExist
	}

	content, err := os.ReadFile(skillFile)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// List returns all skills in the store
func (s *Store) List() ([]*Skill, error) {
	var skills []*Skill

	dir, err := s.expandDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return skills, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(dir, entry.Name())

		// Find the actual skill file (handles case-insensitive filesystems)
		skillFile, err := findSkillFile(skillDir)
		if err != nil {
			continue
		}

		skill, err := ParseSkillFile(skillFile)
		if err != nil {
			continue
		}

		// Use directory name if name is empty
		if skill.Name == "" {
			skill.Name = entry.Name()
		}

		skills = append(skills, skill)
	}

	return skills, nil
}
