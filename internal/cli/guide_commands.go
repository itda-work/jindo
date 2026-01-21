package cli

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/itda-work/jindo/internal/command"
	"github.com/itda-work/jindo/internal/guide"
	"github.com/itda-work/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	guideCommandsInteractive bool
	guideCommandsGlobal      bool
	guideCommandsLocal       bool
	guideCommandsRefresh     bool
	guideCommandsFormat      string
)

var guideCommandsCmd = &cobra.Command{
	Use:     "commands <command-name>",
	Aliases: []string{"c", "command", "cmd"},
	Short:   "Get AI-powered usage guide for a command",
	Long: `Get an AI-powered usage guide for a Claude Code command.

The guide explains:
- What the command does
- How to invoke it (slash command format)
- Practical examples
- Customization suggestions

Guides are cached for future use. Use --refresh to regenerate.
Use -i for interactive mode where AI asks about your context.
Use --format html to generate HTML and open in browser.`,
	Example: `  # Get usage guide for a command (uses cache if available)
  jd guide commands my-command

  # Force regenerate the guide
  jd guide commands my-command --refresh

  # Generate HTML and open in browser
  jd guide commands my-command --format html

  # Interactive mode (not cached)
  jd guide commands my-command -i`,
	Args:              cobra.ExactArgs(1),
	RunE:              runGuideCommands,
	ValidArgsFunction: commandNameCompletion,
}

func init() {
	guideCmd.AddCommand(guideCommandsCmd)
	guideCommandsCmd.Flags().BoolVarP(&guideCommandsInteractive, "interactive", "i", false, "Interactive mode - AI asks questions for personalized guidance")
	guideCommandsCmd.Flags().BoolVarP(&guideCommandsGlobal, "global", "g", false, "Guide from global ~/.claude/commands/")
	guideCommandsCmd.Flags().BoolVarP(&guideCommandsLocal, "local", "l", false, "Guide from local .claude/commands/")
	guideCommandsCmd.Flags().BoolVarP(&guideCommandsRefresh, "refresh", "r", false, "Regenerate the guide even if cached")
	guideCommandsCmd.Flags().StringVarP(&guideCommandsFormat, "format", "f", "", "Output format: html (opens in browser)")
}

func runGuideCommands(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if guideCommandsFormat != "" && guideCommandsFormat != "html" {
		return fmt.Errorf("invalid format: %s (use 'html')", guideCommandsFormat)
	}

	commandName := args[0]

	scope, err := ResolveScope(guideCommandsGlobal, guideCommandsLocal)
	if err != nil {
		return err
	}

	commandsDir := GetPathByScope(scope, "commands")
	store := command.NewStore(commandsDir)

	c, err := store.Get(commandName)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("command not found in %s: %s", ScopeDescription(scope), commandName)
		}
		return fmt.Errorf("failed to get command: %w", err)
	}

	content, err := store.GetContent(commandName)
	if err != nil {
		return fmt.Errorf("failed to read command content: %w", err)
	}

	// Interactive mode
	if guideCommandsInteractive {
		if guideCommandsFormat == "html" {
			return fmt.Errorf("--format html cannot be used with --interactive")
		}
		systemPrompt, err := buildCommandSystemPrompt(commandName, c.Path, content)
		if err != nil {
			return err
		}
		return guide.RunInteractiveGuide(commandName, systemPrompt)
	}

	guideStore, err := guide.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize guide store: %w", err)
	}

	// Use cache if available
	if !guideCommandsRefresh && guideStore.Exists(guide.TypeCommand, commandName) {
		cached, err := guideStore.Get(guide.TypeCommand, commandName)
		if err == nil {
			if guideCommandsFormat == "html" {
				return guide.OpenHTMLGuide(guide.TypeCommand, commandName, cached.Content, cached.CreatedAt)
			}
			guide.PrintGuide(fmt.Sprintf("Command Guide: %s", commandName), cached.Content, cached.CreatedAt, true)
			return nil
		}
	}

	// Generate new guide
	systemPrompt, err := buildCommandSystemPrompt(commandName, c.Path, content)
	if err != nil {
		return err
	}

	userPrompt := fmt.Sprintf("'%s' 명령에 대한 사용법 가이드를 작성해주세요.", commandName)

	generatedContent, err := guide.RunClaudeWithSpinner(systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate guide: %w", err)
	}

	if generatedContent != "" {
		savedGuide, err := guideStore.Save(guide.TypeCommand, commandName, generatedContent)
		if err != nil {
			fmt.Printf("⚠️  가이드 저장 실패: %v\n", err)
		}

		if guideCommandsFormat == "html" {
			return guide.OpenHTMLGuide(guide.TypeCommand, commandName, generatedContent, savedGuide.CreatedAt)
		}

		guide.PrintGuide(fmt.Sprintf("Command Guide: %s", commandName), generatedContent, savedGuide.CreatedAt, false)
	}

	return nil
}

func buildCommandSystemPrompt(commandName, commandPath, content string) (string, error) {
	promptTemplate, err := prompt.Load("guide-command")
	if err != nil {
		return "", fmt.Errorf("failed to load guide prompt: %w", err)
	}

	tmpl, err := template.New("guide-command").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var systemPrompt bytes.Buffer
	err = tmpl.Execute(&systemPrompt, map[string]string{
		"CommandName": commandName,
		"CommandPath": commandPath,
		"Content":     content,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render prompt: %w", err)
	}

	return systemPrompt.String(), nil
}

// commandNameCompletion provides completion for command names
func commandNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	global, _ := cmd.Flags().GetBool("global")
	local, _ := cmd.Flags().GetBool("local")
	scope, err := ResolveScope(global, local)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	store := command.NewStore(GetPathByScope(scope, "commands"))
	commands, err := store.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, c := range commands {
		if c.Description != "" {
			names = append(names, fmt.Sprintf("%s\t%s", c.Name, c.Description))
		} else {
			names = append(names, c.Name)
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
