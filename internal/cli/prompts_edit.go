package cli

import (
	"fmt"

	"github.com/itda-work/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var promptsEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a prompt",
	Long: `Edit a prompt to customize it.

If no override exists, creates one from the embedded version.
Opens the prompt file in your default editor ($EDITOR or $VISUAL).`,
	Example: `  # Edit the adapt-skill prompt
  jd prompts edit adapt-skill`,
	Args:              cobra.ExactArgs(1),
	RunE:              runPromptsEdit,
	ValidArgsFunction: promptNameCompletion,
}

func init() {
	promptsCmd.AddCommand(promptsEditCmd)
}

func runPromptsEdit(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	name := args[0]

	// Check if prompt exists (in embedded)
	_, err := prompt.GetEmbedded(name)
	if err != nil {
		return fmt.Errorf("prompt not found: %s", name)
	}

	// Get override path
	overridePath, err := prompt.GetOverridePath(name)
	if err != nil {
		return fmt.Errorf("failed to get override path: %w", err)
	}

	// If override doesn't exist, create it from embedded
	if !prompt.HasOverride(name) {
		content, err := prompt.GetEmbedded(name)
		if err != nil {
			return fmt.Errorf("failed to get embedded prompt: %w", err)
		}

		if err := prompt.SaveOverride(name, content); err != nil {
			return fmt.Errorf("failed to create override: %w", err)
		}

		fmt.Printf("Created override from embedded: %s\n", overridePath)
	}

	// Open in editor
	return openEditor(overridePath)
}
