package agent

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Agent represents a Claude Code agent
type Agent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"`
	Path        string `json:"path"`
}

// agentFrontmatter represents the YAML frontmatter structure
type agentFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Model       string `yaml:"model"`
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
		case "name", "description", "model":
			result[key] = value
		}
	}

	return result
}

// ParseAgentFile parses an agent .md file and returns an Agent
func ParseAgentFile(path string) (*Agent, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		Path: path,
	}

	frontmatter, found := extractFrontmatter(string(content))
	if found && frontmatter != "" {
		var fm agentFrontmatter
		if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
			// If YAML parsing fails, fall back to simple parsing
			simple := parseSimpleFrontmatter(frontmatter)
			agent.Name = simple["name"]
			agent.Description = simple["description"]
			agent.Model = simple["model"]
			return agent, nil
		}
		agent.Name = fm.Name
		agent.Description = fm.Description
		agent.Model = fm.Model
	}

	return agent, nil
}

// Store manages agents in a directory
type Store struct {
	baseDir string
}

// NewStore creates a new agent store
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

// Get retrieves a specific agent by name
func (s *Store) Get(name string) (*Agent, error) {
	dir, err := s.expandDir()
	if err != nil {
		return nil, err
	}

	agentFile := filepath.Join(dir, name+".md")

	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		return nil, os.ErrNotExist
	}

	agent, err := ParseAgentFile(agentFile)
	if err != nil {
		return nil, err
	}

	if agent.Name == "" {
		agent.Name = name
	}

	return agent, nil
}

// GetContent retrieves the full content of an agent file
func (s *Store) GetContent(name string) (string, error) {
	dir, err := s.expandDir()
	if err != nil {
		return "", err
	}

	agentFile := filepath.Join(dir, name+".md")

	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		return "", os.ErrNotExist
	}

	content, err := os.ReadFile(agentFile)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// List returns all agents in the store
func (s *Store) List() ([]*Agent, error) {
	var agents []*Agent

	dir, err := s.expandDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return agents, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		fullPath := filepath.Join(dir, name)
		agent, err := ParseAgentFile(fullPath)
		if err != nil {
			continue
		}

		// Use filename if name is empty
		if agent.Name == "" {
			agent.Name = strings.TrimSuffix(name, ".md")
		}

		agents = append(agents, agent)
	}

	return agents, nil
}
