package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-skills/jindo/internal/agent"
	"github.com/spf13/cobra"
)

var agentsListJSON bool

var agentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List all agents",
	Long:    `List all agents from ~/.claude/agents/ and .claude/agents/ directories.`,
	RunE:    runAgentsList,
}

func init() {
	agentsCmd.AddCommand(agentsListCmd)
	agentsListCmd.Flags().BoolVar(&agentsListJSON, "json", false, "Output in JSON format")
}

// agentsListOutput represents JSON output for agents list with scope
type agentsListOutput struct {
	Global []*agent.Agent `json:"global"`
	Local  []*agent.Agent `json:"local,omitempty"`
}

func runAgentsList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	// Get global agents
	globalStore := agent.NewStore(GetGlobalPath("agents"))
	globalAgents, err := globalStore.List()
	if err != nil {
		globalAgents = nil
	}

	// Get local agents (if .claude/agents exists)
	var localAgents []*agent.Agent
	if localPath := GetLocalPath("agents"); localPath != "" {
		localStore := agent.NewStore(localPath)
		localAgents, _ = localStore.List()
	}

	if agentsListJSON {
		output := agentsListOutput{
			Global: globalAgents,
			Local:  localAgents,
		}
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonOutput))
		return nil
	}

	// Print global section
	fmt.Println("=== Global (~/.claude/agents/) ===")
	if len(globalAgents) == 0 {
		fmt.Println("No agents found.")
	} else {
		printAgentsTable(globalAgents)
	}

	// Print local section only if exists and has items
	if len(localAgents) > 0 {
		fmt.Println()
		fmt.Println("=== Local (.claude/agents/) ===")
		printAgentsTable(localAgents)
	}

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
