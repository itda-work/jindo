package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itda-skills/jindo/internal/agent"
	"github.com/spf13/cobra"
)

var (
	agentsShowBrief  bool
	agentsShowGlobal bool
	agentsShowLocal  bool
)

var agentsShowCmd = &cobra.Command{
	Use:     "show <agent-name>",
	Aliases: []string{"s", "get", "view"},
	Short:   "Show agent details",
	Long: `Show the full content of a specific agent from ~/.claude/agents/ (global) or .claude/agents/ (local) directory.

Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args:              cobra.ExactArgs(1),
	RunE:              runAgentsShow,
	ValidArgsFunction: agentNameCompletion,
}

func init() {
	agentsCmd.AddCommand(agentsShowCmd)
	agentsShowCmd.Flags().BoolVar(&agentsShowBrief, "brief", false, "Show only metadata (name, description, model)")
	agentsShowCmd.Flags().BoolVarP(&agentsShowGlobal, "global", "g", false, "Show from global ~/.claude/agents/")
	agentsShowCmd.Flags().BoolVarP(&agentsShowLocal, "local", "l", false, "Show from local .claude/agents/")
}

func runAgentsShow(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]

	scope, err := ResolveScope(agentsShowGlobal, agentsShowLocal)
	if err != nil {
		return err
	}

	store := agent.NewStore(GetPathByScope(scope, "agents"))

	if agentsShowBrief {
		return showAgentBrief(store, name, scope)
	}

	return showAgentFull(store, name, scope)
}

func showAgentBrief(store *agent.Store, name string, scope PathScope) error {
	a, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}

	fmt.Printf("Name:        %s\n", a.Name)
	fmt.Printf("Description: %s\n", a.Description)
	fmt.Printf("Model:       %s\n", a.Model)
	fmt.Printf("Path:        %s\n", a.Path)

	return nil
}

func showAgentFull(store *agent.Store, name string, scope PathScope) error {
	content, err := store.GetContent(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get agent content: %w", err)
	}

	fmt.Print(content)
	return nil
}

// agentNameCompletion provides completion for agent names
func agentNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	global, _ := cmd.Flags().GetBool("global")
	local, _ := cmd.Flags().GetBool("local")
	scope, err := ResolveScope(global, local)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	store := agent.NewStore(GetPathByScope(scope, "agents"))
	agents, err := store.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, a := range agents {
		// Use filename without .md extension (the actual ID used for lookup)
		fileName := strings.TrimSuffix(filepath.Base(a.Path), ".md")
		if a.Description != "" {
			names = append(names, fmt.Sprintf("%s\t%s", fileName, a.Description))
		} else {
			names = append(names, fileName)
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
