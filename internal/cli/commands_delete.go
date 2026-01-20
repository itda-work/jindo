package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/itda-work/itda-jindo/internal/command"
	"github.com/spf13/cobra"
)

var commandsDeleteForce bool

var commandsDeleteCmd = &cobra.Command{
	Use:     "delete <command-name>",
	Aliases: []string{"rm"},
	Short:   "Delete a command",
	Long: `Delete a command from ~/.claude/commands/ directory.

This will delete the command file.
Use --force to skip the confirmation prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: runCommandsDelete,
}

func init() {
	commandsCmd.AddCommand(commandsDeleteCmd)
	commandsDeleteCmd.Flags().BoolVarP(&commandsDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runCommandsDelete(_ *cobra.Command, args []string) error {
	name := args[0]
	store := command.NewStore("~/.claude/commands")

	// Get command to verify it exists
	cmd, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("command not found: %s", name)
		}
		return fmt.Errorf("failed to get command: %w", err)
	}

	// Confirm deletion unless --force
	if !commandsDeleteForce {
		fmt.Printf("Delete command '%s'?\n", name)
		fmt.Printf("  Path: %s\n", cmd.Path)
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
	if err := os.Remove(cmd.Path); err != nil {
		return fmt.Errorf("failed to delete command: %w", err)
	}

	fmt.Printf("Deleted command: %s\n", name)
	return nil
}
