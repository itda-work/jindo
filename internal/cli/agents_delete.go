package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/itda-work/itda-jindo/internal/agent"
	"github.com/spf13/cobra"
)

var agentsDeleteForce bool

var agentsDeleteCmd = &cobra.Command{
	Use:     "delete <agent-name>",
	Aliases: []string{"rm"},
	Short:   "Delete an agent",
	Long: `Delete an agent from ~/.claude/agents/ directory.

This will delete the agent file.
Use --force to skip the confirmation prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: runAgentsDelete,
}

func init() {
	agentsCmd.AddCommand(agentsDeleteCmd)
	agentsDeleteCmd.Flags().BoolVarP(&agentsDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runAgentsDelete(_ *cobra.Command, args []string) error {
	name := args[0]
	store := agent.NewStore("~/.claude/agents")

	// Get agent to verify it exists
	a, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found: %s", name)
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Confirm deletion unless --force
	if !agentsDeleteForce {
		fmt.Printf("Delete agent '%s'?\n", name)
		fmt.Printf("  Path: %s\n", a.Path)
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

	// Delete the agent file
	if err := os.Remove(a.Path); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	fmt.Printf("Deleted agent: %s\n", name)
	return nil
}
