package command

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Command represents a Claude Code command
type Command struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

// commandFrontmatter represents the YAML frontmatter structure
type commandFrontmatter struct {
	Description string `yaml:"description"`
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
		if key == "description" {
			result[key] = value
		}
	}

	return result
}

// findFirstHeading finds the first H1 heading in markdown content
func findFirstHeading(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

// ParseCommandFile parses a command .md file and returns a Command
func ParseCommandFile(path string) (*Command, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cmd := &Command{
		Path: path,
	}

	frontmatter, found := extractFrontmatter(string(content))
	if found && frontmatter != "" {
		var fm commandFrontmatter
		if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
			// If YAML parsing fails, fall back to simple parsing
			simple := parseSimpleFrontmatter(frontmatter)
			cmd.Description = simple["description"]
			if cmd.Description == "" {
				cmd.Description = findFirstHeading(string(content))
			}
			return cmd, nil
		}
		cmd.Description = fm.Description
	}

	// If no description from frontmatter, try first heading
	if cmd.Description == "" {
		cmd.Description = findFirstHeading(string(content))
	}

	return cmd, nil
}

// Store manages commands in a directory
type Store struct {
	baseDir string
}

// NewStore creates a new command store
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

// Get retrieves a specific command by name (supports subdir:name format)
func (s *Store) Get(name string) (*Command, error) {
	dir, err := s.expandDir()
	if err != nil {
		return nil, err
	}

	// Convert name:subname format to path
	parts := strings.Split(name, ":")
	pathParts := append(parts[:len(parts)-1], parts[len(parts)-1]+".md")
	cmdFile := filepath.Join(dir, filepath.Join(pathParts...))

	if _, err := os.Stat(cmdFile); os.IsNotExist(err) {
		return nil, os.ErrNotExist
	}

	cmd, err := ParseCommandFile(cmdFile)
	if err != nil {
		return nil, err
	}

	cmd.Name = name
	return cmd, nil
}

// GetContent retrieves the full content of a command file
func (s *Store) GetContent(name string) (string, error) {
	dir, err := s.expandDir()
	if err != nil {
		return "", err
	}

	// Convert name:subname format to path
	parts := strings.Split(name, ":")
	pathParts := append(parts[:len(parts)-1], parts[len(parts)-1]+".md")
	cmdFile := filepath.Join(dir, filepath.Join(pathParts...))

	if _, err := os.Stat(cmdFile); os.IsNotExist(err) {
		return "", os.ErrNotExist
	}

	content, err := os.ReadFile(cmdFile)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// List returns all commands in the store
func (s *Store) List() ([]*Command, error) {
	var commands []*Command

	dir, err := s.expandDir()
	if err != nil {
		return nil, err
	}

	err = s.walkDir(dir, "", &commands)
	if err != nil {
		if os.IsNotExist(err) {
			return commands, nil
		}
		return nil, err
	}

	return commands, nil
}

// walkDir recursively walks the directory and collects commands
func (s *Store) walkDir(dir, prefix string, commands *[]*Command) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		if entry.IsDir() {
			// Recurse into subdirectory with prefix
			newPrefix := name
			if prefix != "" {
				newPrefix = prefix + ":" + name
			}
			if err := s.walkDir(fullPath, newPrefix, commands); err != nil {
				continue
			}
		} else if strings.HasSuffix(name, ".md") {
			cmd, err := ParseCommandFile(fullPath)
			if err != nil {
				continue
			}

			// Set command name from filename (without .md extension)
			baseName := strings.TrimSuffix(name, ".md")
			if prefix != "" {
				cmd.Name = prefix + ":" + baseName
			} else {
				cmd.Name = baseName
			}

			*commands = append(*commands, cmd)
		}
	}

	return nil
}
