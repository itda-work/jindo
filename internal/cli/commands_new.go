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
	commandsNewEdit   bool
	commandsNewNoAI   bool
	commandsNewDesc   string
	commandsNewGlobal bool
	commandsNewLocal  bool
)

var commandsNewCmd = &cobra.Command{
	Use:     "new <command-name>",
	Aliases: []string{"n", "add", "create"},
	Short:   "Create a new command",
	Long: `Create a new command in ~/.claude/commands/ (global) or .claude/commands/ (local) directory.

By default, uses Claude CLI to interactively generate the command content.
Use --no-ai to create a minimal template without AI assistance.
Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.

Command names can include subdirectory prefix (e.g., "game:asset" creates game/asset.md).`,
	Args: cobra.ExactArgs(1),
	RunE: runCommandsNew,
}

func init() {
	commandsCmd.AddCommand(commandsNewCmd)
	commandsNewCmd.Flags().BoolVarP(&commandsNewEdit, "edit", "e", false, "Open editor after creation")
	commandsNewCmd.Flags().BoolVar(&commandsNewNoAI, "no-ai", false, "Create minimal template without AI")
	commandsNewCmd.Flags().StringVarP(&commandsNewDesc, "description", "d", "", "Command description (for --no-ai mode)")
	commandsNewCmd.Flags().BoolVarP(&commandsNewGlobal, "global", "g", false, "Create in global ~/.claude/commands/")
	commandsNewCmd.Flags().BoolVarP(&commandsNewLocal, "local", "l", false, "Create in local .claude/commands/")
}

func runCommandsNew(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	scope, err := ResolveScope(commandsNewGlobal, commandsNewLocal)
	if err != nil {
		return err
	}

	name := args[0]

	// Get commands directory based on scope
	var baseDir string
	if scope == ScopeLocal {
		localPath, err := GetLocalPathForWrite("commands")
		if err != nil {
			return fmt.Errorf("failed to create local commands directory: %w", err)
		}
		baseDir = localPath
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, ".claude", "commands")
	}

	// Convert name:subname format to path
	parts := strings.Split(name, ":")
	pathParts := append(parts[:len(parts)-1], parts[len(parts)-1]+".md")
	cmdFile := filepath.Join(baseDir, filepath.Join(pathParts...))
	cmdDir := filepath.Dir(cmdFile)

	// Check if command already exists
	if _, err := os.Stat(cmdFile); !os.IsNotExist(err) {
		return fmt.Errorf("command already exists: %s", name)
	}

	// Create directory if needed
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return fmt.Errorf("failed to create command directory: %w", err)
	}

	var content string
	if commandsNewNoAI {
		content = generateCommandTemplate(name, commandsNewDesc)
	} else {
		// Use Claude CLI to generate command content
		generated, err := generateCommandWithClaude(name)
		if err != nil {
			return fmt.Errorf("failed to generate command with Claude: %w", err)
		}
		content = generated
	}

	// Write command file
	if err := os.WriteFile(cmdFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write command file: %w", err)
	}

	fmt.Printf("Created command: %s\n", cmdFile)

	// Open editor if requested
	if commandsNewEdit {
		return openEditor(cmdFile)
	}

	return nil
}

func generateCommandTemplate(name, description string) string {
	if description == "" {
		description = "Description of " + name
	}

	// Get base name for title
	parts := strings.Split(name, ":")
	baseName := parts[len(parts)-1]

	return fmt.Sprintf(`---
description: %s
---

# %s

## Overview

Describe what this command does.

## Usage

Explain how to use this command.

## Examples

Provide usage examples.
`, description, toTitle(baseName))
}

func generateCommandWithClaude(name string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are helping create a new Claude Code slash command named "%s".

Generate a complete command .md file with the following structure:

1. YAML frontmatter with:
   - description: a concise one-line description of what the command does

2. Markdown content with:
   - A heading with the command name
   - Overview section explaining what the command does
   - Usage section with instructions
   - Examples section with practical examples

Claude Code slash commands are invoked via /<command-name> and provide instructions for Claude to follow.

Ask the user a few questions to understand what the command should do, then generate the complete command file content.

Start by asking: "What should the '/%s' command do? Please describe its purpose and main functionality."`, name, name)

	cmd := exec.Command("claude",
		"--print",
		"--system-prompt", systemPrompt,
		fmt.Sprintf("I want to create a new slash command called '/%s'. Help me define it.", name),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
