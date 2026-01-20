package skill

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a Claude Code skill
type Skill struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	AllowedTools []string `json:"allowed_tools"`
	Path         string   `json:"path"`
}

// ParseSkillFile parses a SKILL.md or skill.md file and returns a Skill
func ParseSkillFile(path string) (result *Skill, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	skill := &Skill{
		Path: path,
	}

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Check for frontmatter delimiter
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				// End of frontmatter
				break
			}
		}

		if !inFrontmatter {
			continue
		}

		// Parse YAML-like frontmatter (simple key: value)
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			switch key {
			case "name":
				skill.Name = value
			case "description":
				skill.Description = value
			case "allowed-tools":
				// Parse comma-separated tools
				tools := strings.Split(value, ",")
				for _, tool := range tools {
					tool = strings.TrimSpace(tool)
					if tool != "" {
						skill.AllowedTools = append(skill.AllowedTools, tool)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
