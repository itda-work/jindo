package cli

import (
	"github.com/spf13/cobra"
)

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Manage Claude Code packages from GitHub repositories",
	Long: `Manage Claude Code packages (skills, commands, agents) from GitHub repositories.

This command allows you to:
- Register GitHub repositories containing Claude Code configurations
- Browse and search available packages
- Install, update, and uninstall packages with namespace isolation`,
}

func init() {
	rootCmd.AddCommand(pkgCmd)
}
