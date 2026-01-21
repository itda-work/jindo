package cli

import (
	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:     "hooks",
	Aliases: []string{"h"},
	Short:   "Manage Claude Code hooks",
	Long:    `Manage Claude Code hooks in ~/.claude/settings.json.`,
}

func init() {
	rootCmd.AddCommand(hooksCmd)
}
