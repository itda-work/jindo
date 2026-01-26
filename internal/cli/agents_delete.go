package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/itda-skills/jindo/internal/agent"
	"github.com/spf13/cobra"
)

var (
	agentsDeleteForce  bool
	agentsDeleteGlobal bool
	agentsDeleteLocal  bool
)

var agentsDeleteCmd = &cobra.Command{
	Use:     "delete <agent-name>",
	Aliases: []string{"d", "rm", "remove"},
	Short:   "Delete an agent",
	Long: `Delete an agent from ~/.claude/agents/ (global) or .claude/agents/ (local) directory.

This will delete the agent file.
Use --force to skip the confirmation prompt.
Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args:              cobra.ExactArgs(1),
	RunE:              runAgentsDelete,
	ValidArgsFunction: agentNameCompletion,
}

func init() {
	agentsCmd.AddCommand(agentsDeleteCmd)
	agentsDeleteCmd.Flags().BoolVarP(&agentsDeleteForce, "force", "f", false, "Skip confirmation prompt")
	agentsDeleteCmd.Flags().BoolVarP(&agentsDeleteGlobal, "global", "g", false, "Delete from global ~/.claude/agents/")
	agentsDeleteCmd.Flags().BoolVarP(&agentsDeleteLocal, "local", "l", false, "Delete from local .claude/agents/")
}

func runAgentsDelete(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	scope, err := ResolveScope(agentsDeleteGlobal, agentsDeleteLocal)
	if err != nil {
		return err
	}

	name := args[0]

	store := agent.NewStore(GetPathByScope(scope, "agents"))

	// Get agent to verify it exists
	a, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found in %s: %s", ScopeDescription(scope), name)
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
