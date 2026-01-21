package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	skillsNewEdit   bool
	skillsNewNoAI   bool
	skillsNewDesc   string
	skillsNewTools  string
	skillsNewGlobal bool
	skillsNewLocal  bool
)

var skillsNewCmd = &cobra.Command{
	Use:     "new <skill-name>",
	Aliases: []string{"n", "add", "create"},
	Short:   "Create a new skill",
	Long: `Create a new skill in ~/.claude/skills/ (global) or .claude/skills/ (local) directory.

By default, uses Claude CLI to interactively generate the skill content.
Use --no-ai to create a minimal template without AI assistance.
Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args: cobra.ExactArgs(1),
	RunE: runSkillsNew,
}

func init() {
	skillsCmd.AddCommand(skillsNewCmd)
	skillsNewCmd.Flags().BoolVarP(&skillsNewEdit, "edit", "e", false, "Open editor after creation")
	skillsNewCmd.Flags().BoolVar(&skillsNewNoAI, "no-ai", false, "Create minimal template without AI")
	skillsNewCmd.Flags().StringVarP(&skillsNewDesc, "description", "d", "", "Skill description (for --no-ai mode)")
	skillsNewCmd.Flags().StringVarP(&skillsNewTools, "tools", "t", "", "Allowed tools, comma-separated (for --no-ai mode)")
	skillsNewCmd.Flags().BoolVarP(&skillsNewGlobal, "global", "g", false, "Create in global ~/.claude/skills/")
	skillsNewCmd.Flags().BoolVarP(&skillsNewLocal, "local", "l", false, "Create in local .claude/skills/")
}

func runSkillsNew(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	scope, err := ResolveScope(skillsNewGlobal, skillsNewLocal)
	if err != nil {
		return err
	}

	name := args[0]

	// Get skills directory based on scope
	var skillDir string
	if scope == ScopeLocal {
		localPath, err := GetLocalPathForWrite("skills")
		if err != nil {
			return fmt.Errorf("failed to create local skills directory: %w", err)
		}
		skillDir = filepath.Join(localPath, name)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		skillDir = filepath.Join(home, ".claude", "skills", name)
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")

	// Check if skill already exists
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		return fmt.Errorf("skill already exists: %s", name)
	}

	// Create skill directory
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	var content string
	if skillsNewNoAI {
		content = generateSkillTemplate(name, skillsNewDesc, skillsNewTools)
	} else {
		// Use Claude CLI to generate skill content
		generated, err := generateSkillWithClaude(name)
		if err != nil {
			// Cleanup directory on failure
			_ = os.RemoveAll(skillDir)
			return fmt.Errorf("failed to generate skill with Claude: %w", err)
		}
		content = generated
	}

	// Write skill file
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		_ = os.RemoveAll(skillDir)
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	fmt.Printf("✅ Created skill: %s\n", skillFile)

	// Open editor if requested
	if skillsNewEdit {
		return openEditor(skillFile)
	}

	return nil
}

func generateSkillTemplate(name, description, tools string) string {
	if description == "" {
		description = "Description of " + name
	}
	if tools == "" {
		tools = "Bash, Read, Write, Edit, Glob, Grep"
	}

	return fmt.Sprintf(`---
name: %s
description: %s
allowed-tools: %s
---

# %s

## Overview

Describe what this skill does.

## Usage

Explain how to use this skill.

## Examples

Provide usage examples.
`, name, description, tools, toTitle(name))
}

func generateSkillWithClaude(name string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are helping create a new Claude Code skill named "%s".

Generate a complete SKILL.md file with the following structure:

1. YAML frontmatter with:
   - name: the skill name
   - description: a concise one-line description
   - allowed-tools: comma-separated list of tools this skill needs (e.g., Bash, Read, Write, Edit, Glob, Grep, WebFetch, WebSearch)

2. Markdown content with:
   - A heading with the skill name
   - Overview section explaining what the skill does
   - Usage section with instructions
   - Examples section with practical examples

Ask the user a few questions to understand what the skill should do, then generate the complete SKILL.md content.

Start by asking: "What should the '%s' skill do? Please describe its purpose and main functionality."`, name, name)

	// Run Claude CLI with the system prompt
	cmd := exec.Command("claude",
		"--print",
		"--system-prompt", systemPrompt,
		fmt.Sprintf("I want to create a new skill called '%s'. Help me define it.", name),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// toTitle converts a kebab-case name to Title Case
func toTitle(name string) string {
	words := strings.Split(strings.ReplaceAll(name, "-", " "), " ")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func openEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
