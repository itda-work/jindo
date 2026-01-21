package cli

import (
	"fmt"

	"github.com/itda-work/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var promptsResetCmd = &cobra.Command{
	Use:   "reset <name>",
	Short: "Reset a prompt to embedded default",
	Long: `Reset a prompt by removing the override file.

After reset, the embedded default version will be used.`,
	Example: `  # Reset adapt-skill to default
  jd prompts reset adapt-skill`,
	Args:              cobra.ExactArgs(1),
	RunE:              runPromptsReset,
	ValidArgsFunction: promptNameCompletion,
}

func init() {
	promptsCmd.AddCommand(promptsResetCmd)
}

func runPromptsReset(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	name := args[0]

	// Check if prompt exists (in embedded)
	_, err := prompt.GetEmbedded(name)
	if err != nil {
		return fmt.Errorf("prompt not found: %s", name)
	}

	if !prompt.HasOverride(name) {
		fmt.Printf("No override exists for: %s\n", name)
		fmt.Println("Already using embedded default.")
		return nil
	}

	if err := prompt.DeleteOverride(name); err != nil {
		return fmt.Errorf("failed to reset prompt: %w", err)
	}

	fmt.Printf("✅ Reset prompt '%s' to embedded default\n", name)
	return nil
}
