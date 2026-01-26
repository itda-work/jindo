package cli

import (
	"fmt"

	"github.com/itda-skills/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var promptsShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show a prompt's content",
	Long: `Show the content of a prompt.

If an override exists, shows the override version.
Use --embedded to show the original embedded version.`,
	Example: `  # Show adapt-skill prompt
  jd prompts show adapt-skill

  # Show embedded version (ignoring override)
  jd prompts show adapt-skill --embedded`,
	Args:              cobra.ExactArgs(1),
	RunE:              runPromptsShow,
	ValidArgsFunction: promptNameCompletion,
}

var promptsShowEmbedded bool

func init() {
	promptsCmd.AddCommand(promptsShowCmd)
	promptsShowCmd.Flags().BoolVar(&promptsShowEmbedded, "embedded", false, "Show the embedded version (ignore override)")
}

func runPromptsShow(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	name := args[0]

	var content string
	var err error

	if promptsShowEmbedded {
		content, err = prompt.GetEmbedded(name)
		if err != nil {
			return fmt.Errorf("embedded prompt not found: %s", name)
		}
		fmt.Printf("# Embedded prompt: %s\n\n", name)
	} else {
		var info *prompt.PromptInfo
		content, info, err = prompt.LoadInfo(name)
		if err != nil {
			return fmt.Errorf("prompt not found: %s", name)
		}

		if info.IsOverride {
			fmt.Printf("# Override prompt: %s\n", name)
			fmt.Printf("# Path: %s\n\n", info.Path)
		} else {
			fmt.Printf("# Embedded prompt: %s\n\n", name)
		}
	}

	fmt.Print(content)
	return nil
}

// promptNameCompletion provides completion for prompt names
func promptNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	prompts, err := prompt.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, p := range prompts {
		desc := "embedded"
		if p.IsOverride {
			desc = "override"
		}
		names = append(names, fmt.Sprintf("%s\t%s", p.Name, desc))
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
