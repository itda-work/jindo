package cli

import (
	"fmt"
	"os"

	"github.com/itda-work/itda-jindo/internal/agent"
	"github.com/spf13/cobra"
)

var agentsShowBrief bool

var agentsShowCmd = &cobra.Command{
	Use:   "show <agent-name>",
	Short: "Show agent details",
	Long:  `Show the full content of a specific agent from ~/.claude/agents/ directory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAgentsShow,
}

func init() {
	agentsCmd.AddCommand(agentsShowCmd)
	agentsShowCmd.Flags().BoolVar(&agentsShowBrief, "brief", false, "Show only metadata (name, description, model)")
}

func runAgentsShow(_ *cobra.Command, args []string) error {
	name := args[0]
	store := agent.NewStore("~/.claude/agents")

	if agentsShowBrief {
		return showAgentBrief(store, name)
	}

	return showAgentFull(store, name)
}

func showAgentBrief(store *agent.Store, name string) error {
	a, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found: %s", name)
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}

	fmt.Printf("Name:        %s\n", a.Name)
	fmt.Printf("Description: %s\n", a.Description)
	fmt.Printf("Model:       %s\n", a.Model)
	fmt.Printf("Path:        %s\n", a.Path)

	return nil
}

func showAgentFull(store *agent.Store, name string) error {
	content, err := store.GetContent(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found: %s", name)
		}
		return fmt.Errorf("failed to get agent content: %w", err)
	}

	fmt.Print(content)
	return nil
}
