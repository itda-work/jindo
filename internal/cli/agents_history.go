package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itda-work/jindo/internal/agent"
	"github.com/spf13/cobra"
)

var (
	agentsHistoryGlobal bool
	agentsHistoryLocal  bool
)

var agentsHistoryCmd = &cobra.Command{
	Use:     "history <agent-id>",
	Aliases: []string{"hist"},
	Short:   "Show version history of an agent",
	Long: `Show the version history of an agent.

Each time an agent is adapted, a new version is saved to .history/.
Use 'jd agents revert' to restore a previous version.`,
	Example: `  # Show history of a global agent
  jd agents history my-agent

  # Show history of a local agent
  jd agents history my-agent --local`,
	Args:              cobra.ExactArgs(1),
	RunE:              runAgentsHistory,
	ValidArgsFunction: agentNameCompletion,
}

func init() {
	agentsCmd.AddCommand(agentsHistoryCmd)
	agentsHistoryCmd.Flags().BoolVarP(&agentsHistoryGlobal, "global", "g", false, "Show from global ~/.claude/agents/ (default)")
	agentsHistoryCmd.Flags().BoolVarP(&agentsHistoryLocal, "local", "l", false, "Show from local .claude/agents/")
}

func runAgentsHistory(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Validate mutually exclusive flags
	if err := ValidateScopeFlags(agentsHistoryGlobal, agentsHistoryLocal); err != nil {
		return err
	}

	agentID := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if agentsHistoryLocal {
		scope = ScopeLocal
	}

	agentsDir := GetPathByScope(scope, "agents")
	store := agent.NewStore(agentsDir)

	// Verify agent exists
	a, err := store.Get(agentID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found: %s", agentID)
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Expand agentsDir for history manager
	expandedAgentsDir := agentsDir
	if strings.HasPrefix(expandedAgentsDir, "~/") {
		home, _ := os.UserHomeDir()
		expandedAgentsDir = filepath.Join(home, expandedAgentsDir[2:])
	}

	// Create history manager
	historyMgr := agent.NewHistoryManager(expandedAgentsDir, agentID)

	versions, err := historyMgr.ListVersions()
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Printf("No history found for agent: %s\n", agentID)
		fmt.Println("\nHistory is created when you use 'jd agents adapt'.")
		return nil
	}

	fmt.Printf("Version history for agent: %s\n", agentID)
	fmt.Printf("Path: %s\n\n", a.Path)

	for i, v := range versions {
		marker := "  "
		if i == 0 {
			marker = "* " // Mark the latest
		}
		fmt.Printf("%s%s\n", marker, agent.FormatVersionName(&v))
	}

	fmt.Printf("\nTotal: %d version(s)\n", len(versions))
	fmt.Printf("\nTo revert: jd agents revert %s <version>\n", agentID)

	return nil
}
