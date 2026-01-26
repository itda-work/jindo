package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/itda-skills/jindo/internal/command"
	"github.com/spf13/cobra"
)

var (
	commandsDeleteForce  bool
	commandsDeleteGlobal bool
	commandsDeleteLocal  bool
)

var commandsDeleteCmd = &cobra.Command{
	Use:     "delete <command-name>",
	Aliases: []string{"d", "rm", "remove"},
	Short:   "Delete a command",
	Long: `Delete a command from ~/.claude/commands/ (global) or .claude/commands/ (local) directory.

This will delete the command file.
Use --force to skip the confirmation prompt.
Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args: cobra.ExactArgs(1),
	RunE: runCommandsDelete,
}

func init() {
	commandsCmd.AddCommand(commandsDeleteCmd)
	commandsDeleteCmd.Flags().BoolVarP(&commandsDeleteForce, "force", "f", false, "Skip confirmation prompt")
	commandsDeleteCmd.Flags().BoolVarP(&commandsDeleteGlobal, "global", "g", false, "Delete from global ~/.claude/commands/")
	commandsDeleteCmd.Flags().BoolVarP(&commandsDeleteLocal, "local", "l", false, "Delete from local .claude/commands/")
}

func runCommandsDelete(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	scope, err := ResolveScope(commandsDeleteGlobal, commandsDeleteLocal)
	if err != nil {
		return err
	}

	name := args[0]

	store := command.NewStore(GetPathByScope(scope, "commands"))

	// Get command to verify it exists
	c, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("command not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get command: %w", err)
	}

	// Confirm deletion unless --force
	if !commandsDeleteForce {
		fmt.Printf("Delete command '%s'?\n", name)
		fmt.Printf("  Path: %s\n", c.Path)
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete the command file
	if err := os.Remove(c.Path); err != nil {
		return fmt.Errorf("failed to delete command: %w", err)
	}

	fmt.Printf("Deleted command: %s\n", name)
	return nil
}
