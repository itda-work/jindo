package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	agentsNewEdit  bool
	agentsNewNoAI  bool
	agentsNewDesc  string
	agentsNewModel string
)

var agentsNewCmd = &cobra.Command{
	Use:     "new <agent-name>",
	Aliases: []string{"n"},
	Short:   "Create a new agent",
	Long: `Create a new agent in ~/.claude/agents/ directory.

By default, uses Claude CLI to interactively generate the agent content.
Use --no-ai to create a minimal template without AI assistance.`,
	Args: cobra.ExactArgs(1),
	RunE: runAgentsNew,
}

func init() {
	agentsCmd.AddCommand(agentsNewCmd)
	agentsNewCmd.Flags().BoolVarP(&agentsNewEdit, "edit", "e", false, "Open editor after creation")
	agentsNewCmd.Flags().BoolVar(&agentsNewNoAI, "no-ai", false, "Create minimal template without AI")
	agentsNewCmd.Flags().StringVarP(&agentsNewDesc, "description", "d", "", "Agent description (for --no-ai mode)")
	agentsNewCmd.Flags().StringVarP(&agentsNewModel, "model", "m", "", "Model to use (for --no-ai mode)")
}

func runAgentsNew(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]

	// Get agents directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	agentsDir := filepath.Join(home, ".claude", "agents")
	agentFile := filepath.Join(agentsDir, name+".md")

	// Check if agent already exists
	if _, err := os.Stat(agentFile); !os.IsNotExist(err) {
		return fmt.Errorf("agent already exists: %s", name)
	}

	// Create directory if needed
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	var content string
	if agentsNewNoAI {
		content = generateAgentTemplate(name, agentsNewDesc, agentsNewModel)
	} else {
		// Use Claude CLI to generate agent content
		generated, err := generateAgentWithClaude(name)
		if err != nil {
			return fmt.Errorf("failed to generate agent with Claude: %w", err)
		}
		content = generated
	}

	// Write agent file
	if err := os.WriteFile(agentFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	fmt.Printf("Created agent: %s\n", agentFile)

	// Open editor if requested
	if agentsNewEdit {
		return openEditor(agentFile)
	}

	return nil
}

func generateAgentTemplate(name, description, model string) string {
	if description == "" {
		description = "Description of " + name
	}
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	return fmt.Sprintf(`---
name: %s
description: %s
model: %s
---

# %s

## Overview

Describe what this agent does.

## Capabilities

List what this agent can do.

## Usage

Explain how to use this agent.
`, name, description, model, toTitle(name))
}

func generateAgentWithClaude(name string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are helping create a new Claude Code agent named "%s".

Generate a complete agent .md file with the following structure:

1. YAML frontmatter with:
   - name: the agent name
   - description: a concise one-line description
   - model: the model to use (e.g., claude-sonnet-4-20250514, claude-opus-4-20250514)

2. Markdown content with:
   - A heading with the agent name
   - Overview section explaining what the agent does
   - Capabilities section listing what it can do
   - Usage section with instructions

Agents are specialized Claude instances with specific system prompts and configurations.

Ask the user a few questions to understand what the agent should do, then generate the complete agent file content.

Start by asking: "What should the '%s' agent specialize in? Please describe its purpose and main capabilities."`, name, name)

	cmd := exec.Command("claude",
		"--print",
		"--system-prompt", systemPrompt,
		fmt.Sprintf("I want to create a new agent called '%s'. Help me define it.", name),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
