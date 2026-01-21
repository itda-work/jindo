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
	agentsRevertGlobal bool
	agentsRevertLocal  bool
)

var agentsRevertCmd = &cobra.Command{
	Use:   "revert <agent-id> [version]",
	Short: "Revert an agent to a previous version",
	Long: `Revert an agent to a previous version from its history.

If no version is specified, shows available versions.
Version can be a number (e.g., 1, 2) or 'latest'.`,
	Example: `  # Show available versions
  jd agents revert my-agent

  # Revert to version 1
  jd agents revert my-agent 1

  # Revert to the latest backed up version
  jd agents revert my-agent latest`,
	Args:              cobra.RangeArgs(1, 2),
	RunE:              runAgentsRevert,
	ValidArgsFunction: agentNameCompletion,
}

func init() {
	agentsCmd.AddCommand(agentsRevertCmd)
	agentsRevertCmd.Flags().BoolVarP(&agentsRevertGlobal, "global", "g", false, "Revert from global ~/.claude/agents/")
	agentsRevertCmd.Flags().BoolVarP(&agentsRevertLocal, "local", "l", false, "Revert from local .claude/agents/")
}

func runAgentsRevert(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	agentID := args[0]

	scope, err := ResolveScope(agentsRevertGlobal, agentsRevertLocal)
	if err != nil {
		return err
	}

	agentsDir := GetPathByScope(scope, "agents")
	store := agent.NewStore(agentsDir)

	// Verify agent exists and get its path
	a, err := store.Get(agentID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found in %s: %s", ScopeDescription(scope), agentID)
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

	// If no version specified, show available versions
	if len(args) < 2 {
		versions, err := historyMgr.ListVersions()
		if err != nil {
			return fmt.Errorf("failed to list versions: %w", err)
		}

		if len(versions) == 0 {
			fmt.Printf("No history found for agent: %s\n", agentID)
			return nil
		}

		// Get current content to find active version
		currentContent, _ := store.GetContent(agentID)

		fmt.Printf("Available versions for agent: %s\n\n", agentID)
		for _, v := range versions {
			marker := "  "
			// Check if this version matches current content
			if vContent, _, err := historyMgr.GetVersion(v.Number); err == nil && vContent == currentContent {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, agent.FormatVersionName(&v))
		}
		fmt.Printf("\nUsage: jd agents revert %s <version>\n", agentID)
		return nil
	}

	// Parse version argument
	versionArg := args[1]
	versionNum, err := agent.ParseVersionArg(versionArg)
	if err != nil {
		return err
	}

	var content string
	var version *agent.Version

	if versionNum == -1 {
		// Get latest version
		version, err = historyMgr.GetLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
		content, _, err = historyMgr.GetVersion(version.Number)
	} else {
		content, version, err = historyMgr.GetVersion(versionNum)
	}

	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// Write the reverted content
	if err := os.WriteFile(a.Path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write reverted content: %w", err)
	}

	// Delete all versions after the reverted version
	deleted, err := historyMgr.DeleteVersionsAfter(version.Number)
	if err != nil {
		return fmt.Errorf("failed to cleanup versions: %w", err)
	}

	fmt.Printf("✅ Reverted agent '%s' to %s\n", agentID, agent.FormatVersionName(version))
	if deleted > 0 {
		fmt.Printf("   Removed %d newer version(s)\n", deleted)
	}

	return nil
}
