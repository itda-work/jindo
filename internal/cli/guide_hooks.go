package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/itda-work/jindo/internal/guide"
	"github.com/itda-work/jindo/internal/hook"
	"github.com/itda-work/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	guideHooksInteractive bool
	guideHooksGlobal      bool
	guideHooksLocal       bool
	guideHooksRefresh     bool
	guideHooksFormat      string
)

var guideHooksCmd = &cobra.Command{
	Use:     "hooks <hook-name>",
	Aliases: []string{"h", "hook"},
	Short:   "Get AI-powered usage guide for a hook",
	Long: `Get an AI-powered usage guide for a Claude Code hook.

The guide explains:
- What events trigger this hook
- How it integrates with the workflow
- Practical examples
- Customization suggestions

Guides are cached for future use. Use --refresh to regenerate.
Use -i for interactive mode where AI asks about your context.
Use --format html to generate HTML and open in browser.`,
	Example: `  # Get usage guide for a hook (uses cache if available)
  jd guide hooks PreToolUse-Bash-0

  # Force regenerate the guide
  jd guide hooks PreToolUse-Bash-0 --refresh

  # Generate HTML and open in browser
  jd guide hooks PreToolUse-Bash-0 --format html

  # Interactive mode (not cached)
  jd guide hooks PreToolUse-Bash-0 -i`,
	Args:              cobra.ExactArgs(1),
	RunE:              runGuideHooks,
	ValidArgsFunction: hookNameCompletion,
}

func init() {
	guideCmd.AddCommand(guideHooksCmd)
	guideHooksCmd.Flags().BoolVarP(&guideHooksInteractive, "interactive", "i", false, "Interactive mode - AI asks questions for personalized guidance")
	guideHooksCmd.Flags().BoolVarP(&guideHooksGlobal, "global", "g", false, "Guide from global ~/.claude/settings.json")
	guideHooksCmd.Flags().BoolVarP(&guideHooksLocal, "local", "l", false, "Guide from local .claude/settings.json")
	guideHooksCmd.Flags().BoolVarP(&guideHooksRefresh, "refresh", "r", false, "Regenerate the guide even if cached")
	guideHooksCmd.Flags().StringVarP(&guideHooksFormat, "format", "f", "", "Output format: html (opens in browser)")
}

func runGuideHooks(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if guideHooksFormat != "" && guideHooksFormat != "html" {
		return fmt.Errorf("invalid format: %s (use 'html')", guideHooksFormat)
	}

	hookName := args[0]

	scope, err := ResolveScope(guideHooksGlobal, guideHooksLocal)
	if err != nil {
		return err
	}

	store := hook.NewStore(GetSettingsPathByScope(scope))

	h, err := store.Get(hookName)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("hook not found in %s: %s", ScopeDescription(scope), hookName)
		}
		return fmt.Errorf("failed to get hook: %w", err)
	}

	content, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize hook: %w", err)
	}

	// Interactive mode
	if guideHooksInteractive {
		if guideHooksFormat == "html" {
			return fmt.Errorf("--format html cannot be used with --interactive")
		}
		systemPrompt, err := buildHookSystemPrompt(hookName, GetSettingsPathByScope(scope), string(h.EventType), string(content))
		if err != nil {
			return err
		}
		return guide.RunInteractiveGuide(hookName, systemPrompt)
	}

	guideStore, err := guide.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize guide store: %w", err)
	}

	// Use cache if available
	if !guideHooksRefresh && guideStore.Exists(guide.TypeHook, hookName) {
		cached, err := guideStore.Get(guide.TypeHook, hookName)
		if err == nil {
			if guideHooksFormat == "html" {
				return guide.OpenHTMLGuide(guide.TypeHook, hookName, cached.Content, cached.CreatedAt)
			}
			guide.PrintGuide(fmt.Sprintf("Hook Guide: %s", hookName), cached.Content, cached.CreatedAt, true)
			return nil
		}
	}

	// Generate new guide
	systemPrompt, err := buildHookSystemPrompt(hookName, GetSettingsPathByScope(scope), string(h.EventType), string(content))
	if err != nil {
		return err
	}

	userPrompt := fmt.Sprintf("'%s' 훅에 대한 사용법 가이드를 작성해주세요.", hookName)

	generatedContent, err := guide.RunClaudeWithSpinner(systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate guide: %w", err)
	}

	if generatedContent != "" {
		savedGuide, err := guideStore.Save(guide.TypeHook, hookName, generatedContent)
		if err != nil {
			fmt.Printf("⚠️  가이드 저장 실패: %v\n", err)
		}

		if guideHooksFormat == "html" {
			return guide.OpenHTMLGuide(guide.TypeHook, hookName, generatedContent, savedGuide.CreatedAt)
		}

		guide.PrintGuide(fmt.Sprintf("Hook Guide: %s", hookName), generatedContent, savedGuide.CreatedAt, false)
	}

	return nil
}

func buildHookSystemPrompt(hookName, hookPath, hookType, content string) (string, error) {
	promptTemplate, err := prompt.Load("guide-hook")
	if err != nil {
		return "", fmt.Errorf("failed to load guide prompt: %w", err)
	}

	tmpl, err := template.New("guide-hook").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var systemPrompt bytes.Buffer
	err = tmpl.Execute(&systemPrompt, map[string]string{
		"HookName": hookName,
		"HookPath": hookPath,
		"HookType": hookType,
		"Content":  content,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render prompt: %w", err)
	}

	return systemPrompt.String(), nil
}
