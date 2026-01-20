package command

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Command represents a Claude Code command
type Command struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

// ParseCommandFile parses a command .md file and returns a Command
func ParseCommandFile(path string) (result *Command, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	cmd := &Command{
		Path: path,
	}

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	frontmatterFound := false
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Check for frontmatter delimiter
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter && lineCount == 1 {
				inFrontmatter = true
				frontmatterFound = true
				continue
			} else if inFrontmatter {
				// End of frontmatter
				inFrontmatter = false
				continue
			}
		}

		if inFrontmatter {
			// Parse YAML-like frontmatter (simple key: value)
			if idx := strings.Index(line, ":"); idx > 0 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])

				switch key {
				case "description":
					cmd.Description = value
				}
			}
			continue
		}

		// If no frontmatter, try to get description from first heading
		if !frontmatterFound && strings.HasPrefix(line, "# ") {
			// Use first heading as description if no frontmatter description
			if cmd.Description == "" {
				cmd.Description = strings.TrimPrefix(line, "# ")
			}
			break
		}

		// If we have frontmatter but passed it, stop reading
		if frontmatterFound && !inFrontmatter {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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

// List returns all commands in the store
func (s *Store) List() ([]*Command, error) {
	var commands []*Command

	// Expand ~ to home directory
	dir := s.baseDir
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(home, dir[2:])
	}

	err := s.walkDir(dir, "", &commands)
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
