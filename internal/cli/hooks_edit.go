package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/itda-work/jindo/internal/hook"
	"github.com/spf13/cobra"
)

var (
	hooksEditMatcher string
	hooksEditCommand string
	hooksEditGlobal  bool
	hooksEditLocal   bool
)

var hooksEditCmd = &cobra.Command{
	Use:     "edit <name>",
	Aliases: []string{"e", "update", "modify"},
	Short:   "Edit a hook",
	Long: `Edit an existing hook in ~/.claude/settings.json (global) or .claude/settings.json (local).

If no flags are provided, runs in interactive mode showing current values.
Use --local to edit from the current directory's .claude/settings.json.

Examples:
  jd hooks edit PreToolUse-Bash-0
  jd hooks edit PreToolUse-Bash-0 -m "Bash|Write"
  jd hooks edit PreToolUse-Bash-0 -c "new-command.sh"
  jd hooks edit --local PreToolUse-Bash-0`,
	Args:              cobra.ExactArgs(1),
	RunE:              runHooksEdit,
	ValidArgsFunction: hookNameCompletion,
}

func init() {
	hooksCmd.AddCommand(hooksEditCmd)
	hooksEditCmd.Flags().StringVarP(&hooksEditMatcher, "matcher", "m", "", "New matcher pattern")
	hooksEditCmd.Flags().StringVarP(&hooksEditCommand, "command", "c", "", "New command (replaces all existing commands)")
	hooksEditCmd.Flags().BoolVarP(&hooksEditGlobal, "global", "g", false, "Edit from global ~/.claude/settings.json (default)")
	hooksEditCmd.Flags().BoolVarP(&hooksEditLocal, "local", "l", false, "Edit from local .claude/settings.json")
}

func runHooksEdit(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if hooksEditLocal {
		scope = ScopeLocal
	}

	store := hook.NewStore(GetSettingsPathByScope(scope))
	h, err := store.Get(name)
	if err != nil {
		return fmt.Errorf("hook not found: %s", name)
	}

	reader := bufio.NewReader(os.Stdin)
	newMatcher := hooksEditMatcher
	newCommand := hooksEditCommand

	// Interactive mode if no flags provided
	if newMatcher == "" && newCommand == "" {
		fmt.Printf("Editing hook: %s\n\n", name)

		// Matcher
		fmt.Printf("Current matcher: %s\n", h.Matcher)
		fmt.Print("New matcher (press Enter to keep current): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			newMatcher = input
		} else {
			newMatcher = h.Matcher
		}

		// Command
		fmt.Printf("\nCurrent commands:\n")
		for i, cmd := range h.Commands {
			fmt.Printf("  %d. %s\n", i+1, cmd)
		}
		fmt.Print("New command (press Enter to keep current): ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			newCommand = input
		}
	}

	// Apply defaults if still empty
	if newMatcher == "" {
		newMatcher = h.Matcher
	}

	var commands []string
	if newCommand != "" {
		commands = []string{newCommand}
	} else {
		commands = h.Commands
	}

	// Update the hook
	updated, err := store.Update(name, newMatcher, commands)
	if err != nil {
		return fmt.Errorf("failed to update hook: %w", err)
	}

	fmt.Printf("\n✓ Updated hook: %s\n", updated.Name)
	fmt.Printf("  Matcher: %s\n", updated.Matcher)
	fmt.Printf("  Commands: %s\n", strings.Join(updated.Commands, ", "))

	return nil
}
