package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itda-work/jindo/internal/agent"
	"github.com/itda-work/jindo/internal/command"
	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

// Known valid Claude Code tools
var validTools = map[string]bool{
	"Bash":         true,
	"Read":         true,
	"Write":        true,
	"Edit":         true,
	"Glob":         true,
	"Grep":         true,
	"LS":           true,
	"WebFetch":     true,
	"WebSearch":    true,
	"Task":         true,
	"TodoRead":     true,
	"TodoWrite":    true,
	"NotebookEdit": true,
	"NotebookRead": true,
}

var (
	validateSkillsOnly   bool
	validateCommandsOnly bool
	validateAgentsOnly   bool
	validateVerbose      bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate skills, commands, and agents",
	Long: `Validate the format and content of all skills, commands, and agents.

Checks:
- YAML frontmatter parsing
- Required fields (name, description)
- Skill allowed-tools validity`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVarP(&validateSkillsOnly, "skills", "s", false, "Validate only skills")
	validateCmd.Flags().BoolVarP(&validateCommandsOnly, "commands", "c", false, "Validate only commands")
	validateCmd.Flags().BoolVarP(&validateAgentsOnly, "agents", "a", false, "Validate only agents")
	validateCmd.Flags().BoolVarP(&validateVerbose, "verbose", "v", false, "Show all files, not just errors")
}

// ValidationError represents a single validation error
type ValidationError struct {
	Type    string // "skill", "command", "agent"
	Name    string
	Path    string
	Message string
}

// ValidationResult holds all validation results
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
	Checked  int
}

func runValidate(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	result := &ValidationResult{}

	// Determine which resources to validate
	validateAll := !validateSkillsOnly && !validateCommandsOnly && !validateAgentsOnly

	// Validate skills
	if validateAll || validateSkillsOnly {
		if err := validateSkills(result); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to validate skills: %v\n", err)
		}
	}

	// Validate commands
	if validateAll || validateCommandsOnly {
		if err := validateCommands(result); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to validate commands: %v\n", err)
		}
	}

	// Validate agents
	if validateAll || validateAgentsOnly {
		if err := validateAgents(result); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to validate agents: %v\n", err)
		}
	}

	// Print results
	printValidationResults(result)

	// Return error if there are validation errors
	if len(result.Errors) > 0 {
		return fmt.Errorf("validation failed with %d error(s)", len(result.Errors))
	}

	return nil
}

func validateSkills(result *ValidationResult) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	skillsDir := filepath.Join(home, ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	store := skill.NewStore("~/.claude/skills")

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		result.Checked++

		s, err := store.Get(name)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "skill",
				Name:    name,
				Path:    filepath.Join(skillsDir, name),
				Message: fmt.Sprintf("failed to parse: %v", err),
			})
			continue
		}

		// Check required fields
		if s.Name == "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "skill",
				Name:    name,
				Path:    s.Path,
				Message: "missing 'name' in frontmatter (using directory name)",
			})
		}

		if s.Description == "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "skill",
				Name:    name,
				Path:    s.Path,
				Message: "missing 'description' in frontmatter",
			})
		}

		// Check allowed-tools
		for _, tool := range s.AllowedTools {
			tool = strings.TrimSpace(tool)
			if tool != "" && !validTools[tool] {
				result.Warnings = append(result.Warnings, ValidationError{
					Type:    "skill",
					Name:    name,
					Path:    s.Path,
					Message: fmt.Sprintf("unknown tool in allowed-tools: %s", tool),
				})
			}
		}

		if validateVerbose {
			fmt.Printf("  [OK] skill: %s\n", name)
		}
	}

	return nil
}

func validateCommands(result *ValidationResult) error {
	store := command.NewStore("~/.claude/commands")
	commands, err := store.List()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, cmd := range commands {
		result.Checked++

		// Check required fields
		if cmd.Description == "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "command",
				Name:    cmd.Name,
				Path:    cmd.Path,
				Message: "missing 'description' in frontmatter",
			})
		}

		if validateVerbose {
			fmt.Printf("  [OK] command: %s\n", cmd.Name)
		}
	}

	return nil
}

func validateAgents(result *ValidationResult) error {
	store := agent.NewStore("~/.claude/agents")
	agents, err := store.List()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, a := range agents {
		result.Checked++

		// Check required fields
		if a.Name == "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "agent",
				Name:    filepath.Base(a.Path),
				Path:    a.Path,
				Message: "missing 'name' in frontmatter (using filename)",
			})
		}

		if a.Description == "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "agent",
				Name:    a.Name,
				Path:    a.Path,
				Message: "missing 'description' in frontmatter",
			})
		}

		if a.Model == "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:    "agent",
				Name:    a.Name,
				Path:    a.Path,
				Message: "missing 'model' in frontmatter",
			})
		}

		if validateVerbose {
			fmt.Printf("  [OK] agent: %s\n", a.Name)
		}
	}

	return nil
}

func printValidationResults(result *ValidationResult) {
	// Print errors
	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, e := range result.Errors {
			fmt.Printf("  [ERROR] %s '%s': %s\n", e.Type, e.Name, e.Message)
			fmt.Printf("          Path: %s\n", e.Path)
		}
		fmt.Println()
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  [WARN] %s '%s': %s\n", w.Type, w.Name, w.Message)
		}
		fmt.Println()
	}

	// Print summary
	fmt.Printf("Checked %d items: %d error(s), %d warning(s)\n",
		result.Checked, len(result.Errors), len(result.Warnings))

	if len(result.Errors) == 0 && len(result.Warnings) == 0 {
		fmt.Println("All validations passed!")
	}
}
