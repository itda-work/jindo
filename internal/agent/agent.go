package agent

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Agent represents a Claude Code agent
type Agent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"`
	Path        string `json:"path"`
}

// ParseAgentFile parses an agent .md file and returns an Agent
func ParseAgentFile(path string) (result *Agent, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	agent := &Agent{
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
			if !inFrontmatter && lineCount == 1 {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
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
				agent.Name = value
			case "description":
				agent.Description = value
			case "model":
				agent.Model = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
