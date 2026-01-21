package cli

import (
	"fmt"

	"github.com/itda-work/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var promptsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available prompts",
	Long: `List all available prompts.

Prompts marked with [override] have custom versions in ~/.claude/jindo/prompts/.`,
	RunE: runPromptsList,
}

func init() {
	promptsCmd.AddCommand(promptsListCmd)
}

func runPromptsList(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	prompts, err := prompt.List()
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	if len(prompts) == 0 {
		fmt.Println("No prompts found.")
		return nil
	}

	fmt.Println("Available prompts:")
	fmt.Println()

	for _, p := range prompts {
		status := ""
		if p.IsOverride {
			status = " [override]"
		}
		fmt.Printf("  %s%s\n", p.Name, status)
	}

	fmt.Println()
	fmt.Println("Use 'jd prompts show <name>' to view a prompt.")
	fmt.Println("Use 'jd prompts edit <name>' to customize a prompt.")

	return nil
}
