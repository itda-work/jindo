package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/itda-skills/jindo/internal/agent"
	"github.com/spf13/cobra"
)

var (
	agentsEditEditor bool
	agentsEditGlobal bool
	agentsEditLocal  bool
)

var agentsEditCmd = &cobra.Command{
	Use:     "edit <agent-name>",
	Aliases: []string{"e", "update", "modify"},
	Short:   "Edit an existing agent",
	Long: `Edit an existing agent in ~/.claude/agents/ (global) or .claude/agents/ (local) directory.

By default, uses Claude CLI to interactively edit the agent content.
Use --editor to open the agent file directly in your editor.
Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args:              cobra.ExactArgs(1),
	RunE:              runAgentsEdit,
	ValidArgsFunction: agentNameCompletion,
}

func init() {
	agentsCmd.AddCommand(agentsEditCmd)
	agentsEditCmd.Flags().BoolVarP(&agentsEditEditor, "editor", "e", false, "Open in editor directly (skip AI)")
	agentsEditCmd.Flags().BoolVarP(&agentsEditGlobal, "global", "g", false, "Edit from global ~/.claude/agents/")
	agentsEditCmd.Flags().BoolVarP(&agentsEditLocal, "local", "l", false, "Edit from local .claude/agents/")
}

func runAgentsEdit(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	scope, err := ResolveScope(agentsEditGlobal, agentsEditLocal)
	if err != nil {
		return err
	}

	name := args[0]

	store := agent.NewStore(GetPathByScope(scope, "agents"))

	// Get agent to verify it exists and get its path
	a, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// If --editor flag, just open in editor
	if agentsEditEditor {
		return openEditor(a.Path)
	}

	// Get current content for context
	content, err := store.GetContent(name)
	if err != nil {
		return fmt.Errorf("failed to read agent content: %w", err)
	}

	// Use Claude CLI to edit
	newContent, err := editAgentWithClaude(name, content)
	if err != nil {
		return fmt.Errorf("failed to edit agent with Claude: %w", err)
	}

	// Write updated content
	if err := os.WriteFile(a.Path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	fmt.Printf("Updated agent: %s\n", a.Path)
	return nil
}

func editAgentWithClaude(name, currentContent string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are helping edit a Claude Code agent named "%s".

Current agent content:
---
%s
---

Help the user modify this agent. When they describe the changes they want:
1. Understand what they want to change
2. Generate the complete updated agent file content

The output must be a valid agent .md file with:
- YAML frontmatter (name, description, model)
- Markdown content

Ask the user what changes they want to make to this agent.`, name, currentContent)

	cmd := exec.Command("claude",
		"--print",
		"--system-prompt", systemPrompt,
		fmt.Sprintf("I want to edit the '%s' agent. Here's the current content. What would you like to change?", name),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
