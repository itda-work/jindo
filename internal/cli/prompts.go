package cli

import (
	"github.com/spf13/cobra"
)

var promptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "Manage jindo prompts",
	Long: `Manage prompts used by jindo commands like 'adapt'.

Prompts are stored embedded in the binary by default.
You can override them by creating files in ~/.claude/jindo/prompts/.`,
}

func init() {
	rootCmd.AddCommand(promptsCmd)
}
