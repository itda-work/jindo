package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/itda-work/jindo/internal/command"
	"github.com/spf13/cobra"
)

var (
	commandsEditEditor bool
	commandsEditGlobal bool
	commandsEditLocal  bool
)

var commandsEditCmd = &cobra.Command{
	Use:     "edit <command-name>",
	Aliases: []string{"e", "update", "modify"},
	Short:   "Edit an existing command",
	Long: `Edit an existing command in ~/.claude/commands/ (global) or .claude/commands/ (local) directory.

By default, uses Claude CLI to interactively edit the command content.
Use --editor to open the command file directly in your editor.
Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args: cobra.ExactArgs(1),
	RunE: runCommandsEdit,
}

func init() {
	commandsCmd.AddCommand(commandsEditCmd)
	commandsEditCmd.Flags().BoolVarP(&commandsEditEditor, "editor", "e", false, "Open in editor directly (skip AI)")
	commandsEditCmd.Flags().BoolVarP(&commandsEditGlobal, "global", "g", false, "Edit from global ~/.claude/commands/")
	commandsEditCmd.Flags().BoolVarP(&commandsEditLocal, "local", "l", false, "Edit from local .claude/commands/")
}

func runCommandsEdit(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	scope, err := ResolveScope(commandsEditGlobal, commandsEditLocal)
	if err != nil {
		return err
	}

	name := args[0]

	store := command.NewStore(GetPathByScope(scope, "commands"))

	// Get command to verify it exists and get its path
	c, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("command not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get command: %w", err)
	}

	// If --editor flag, just open in editor
	if commandsEditEditor {
		return openEditor(c.Path)
	}

	// Get current content for context
	content, err := store.GetContent(name)
	if err != nil {
		return fmt.Errorf("failed to read command content: %w", err)
	}

	// Use Claude CLI to edit
	newContent, err := editCommandWithClaude(name, content)
	if err != nil {
		return fmt.Errorf("failed to edit command with Claude: %w", err)
	}

	// Write updated content
	if err := os.WriteFile(c.Path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write command file: %w", err)
	}

	fmt.Printf("Updated command: %s\n", c.Path)
	return nil
}

func editCommandWithClaude(name, currentContent string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are helping edit a Claude Code slash command named "%s".

Current command content:
---
%s
---

Help the user modify this command. When they describe the changes they want:
1. Understand what they want to change
2. Generate the complete updated command file content

The output must be a valid command .md file with:
- YAML frontmatter (description)
- Markdown content

Ask the user what changes they want to make to this command.`, name, currentContent)

	cmd := exec.Command("claude",
		"--print",
		"--system-prompt", systemPrompt,
		fmt.Sprintf("I want to edit the '/%s' command. Here's the current content. What would you like to change?", name),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
