package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-work/jindo/internal/agent"
	"github.com/spf13/cobra"
)

var agentsListJSON bool

var agentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List all agents",
	Long:    `List all agents from ~/.claude/agents/ directory.`,
	RunE:    runAgentsList,
}

func init() {
	agentsCmd.AddCommand(agentsListCmd)
	agentsListCmd.Flags().BoolVar(&agentsListJSON, "json", false, "Output in JSON format")
}

func runAgentsList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	store := agent.NewStore("~/.claude/agents")
	agents, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	if len(agents) == 0 {
		fmt.Println("No agents found.")
		return nil
	}

	if agentsListJSON {
		return printAgentsJSON(agents)
	}

	printAgentsTable(agents)
	return nil
}

func printAgentsJSON(agents []*agent.Agent) error {
	output, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func printAgentsTable(agents []*agent.Agent) {
	// Calculate column widths
	nameWidth := len("NAME")
	modelWidth := len("MODEL")

	for _, a := range agents {
		if len(a.Name) > nameWidth {
			nameWidth = len(a.Name)
		}
		if len(a.Model) > modelWidth {
			modelWidth = len(a.Model)
		}
	}

	// Cap widths
	if nameWidth > 25 {
		nameWidth = 25
	}
	if modelWidth > 10 {
		modelWidth = 10
	}
	const descWidth = 50

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s\n",
		nameWidth, "NAME",
		modelWidth, "MODEL",
		descWidth, "DESCRIPTION")
	fmt.Printf("%s  %s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", modelWidth),
		strings.Repeat("-", descWidth))

	// Print rows
	for _, a := range agents {
		name := a.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		model := a.Model
		if len(model) > modelWidth {
			model = model[:modelWidth-3] + "..."
		}

		desc := a.Description
		if len(desc) > descWidth {
			desc = desc[:descWidth-3] + "..."
		}

		fmt.Printf("%-*s  %-*s  %-*s\n",
			nameWidth, name,
			modelWidth, model,
			descWidth, desc)
	}

	fmt.Printf("\nTotal: %d agents\n", len(agents))
}
