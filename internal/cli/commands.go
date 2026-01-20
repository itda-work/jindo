package cli

import (
	"github.com/spf13/cobra"
)

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Manage Claude Code commands",
	Long:  `Manage Claude Code commands in ~/.claude/commands/ directory.`,
}

func init() {
	rootCmd.AddCommand(commandsCmd)
}
