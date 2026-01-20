package cli

import (
	"github.com/spf13/cobra"
)

var pkgRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage registered package repositories",
	Long:  `Manage GitHub repositories that contain Claude Code packages (skills, commands, agents).`,
}

func init() {
	pkgCmd.AddCommand(pkgRepoCmd)
}
