package cli

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/itda-skills/jindo/internal/agent"
	"github.com/itda-skills/jindo/internal/guide"
	"github.com/itda-skills/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	guideAgentsInteractive bool
	guideAgentsGlobal      bool
	guideAgentsLocal       bool
	guideAgentsRefresh     bool
	guideAgentsFormat      string
)

var guideAgentsCmd = &cobra.Command{
	Use:     "agents <agent-id>",
	Aliases: []string{"a", "agent"},
	Short:   "Get AI-powered usage guide for an agent",
	Long: `Get an AI-powered usage guide for a Claude Code agent.

The guide explains:
- What the agent is designed for
- When and how it gets triggered
- How to use it effectively
- Customization suggestions

Guides are cached for future use. Use --refresh to regenerate.
Use -i for interactive mode where AI asks about your context.
Use --format html to generate HTML and open in browser.`,
	Example: `  # Get usage guide for an agent (uses cache if available)
  jd guide agents my-agent

  # Force regenerate the guide
  jd guide agents my-agent --refresh

  # Generate HTML and open in browser
  jd guide agents my-agent --format html

  # Interactive mode (not cached)
  jd guide agents my-agent -i`,
	Args:              cobra.ExactArgs(1),
	RunE:              runGuideAgents,
	ValidArgsFunction: agentNameCompletion,
}

func init() {
	guideCmd.AddCommand(guideAgentsCmd)
	guideAgentsCmd.Flags().BoolVarP(&guideAgentsInteractive, "interactive", "i", false, "Interactive mode - AI asks questions for personalized guidance")
	guideAgentsCmd.Flags().BoolVarP(&guideAgentsGlobal, "global", "g", false, "Guide from global ~/.claude/agents/")
	guideAgentsCmd.Flags().BoolVarP(&guideAgentsLocal, "local", "l", false, "Guide from local .claude/agents/")
	guideAgentsCmd.Flags().BoolVarP(&guideAgentsRefresh, "refresh", "r", false, "Regenerate the guide even if cached")
	guideAgentsCmd.Flags().StringVarP(&guideAgentsFormat, "format", "f", "", "Output format: html (opens in browser)")
}

func runGuideAgents(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if guideAgentsFormat != "" && guideAgentsFormat != "html" {
		return fmt.Errorf("invalid format: %s (use 'html')", guideAgentsFormat)
	}

	agentID := args[0]

	scope, err := ResolveScope(guideAgentsGlobal, guideAgentsLocal)
	if err != nil {
		return err
	}

	agentsDir := GetPathByScope(scope, "agents")
	store := agent.NewStore(agentsDir)

	a, err := store.Get(agentID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found in %s: %s", ScopeDescription(scope), agentID)
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}

	content, err := store.GetContent(agentID)
	if err != nil {
		return fmt.Errorf("failed to read agent content: %w", err)
	}

	// Interactive mode
	if guideAgentsInteractive {
		if guideAgentsFormat == "html" {
			return fmt.Errorf("--format html cannot be used with --interactive")
		}
		systemPrompt, err := buildAgentSystemPrompt(agentID, a.Path, content)
		if err != nil {
			return err
		}
		return guide.RunInteractiveGuide(agentID, systemPrompt)
	}

	guideStore, err := guide.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize guide store: %w", err)
	}

	// Use cache if available
	if !guideAgentsRefresh && guideStore.Exists(guide.TypeAgent, agentID) {
		cached, err := guideStore.Get(guide.TypeAgent, agentID)
		if err == nil {
			if guideAgentsFormat == "html" {
				return guide.OpenHTMLGuide(guide.TypeAgent, agentID, cached.Content, cached.CreatedAt)
			}
			guide.PrintGuide(fmt.Sprintf("Agent Guide: %s", agentID), cached.Content, cached.CreatedAt, true)
			return nil
		}
	}

	// Generate new guide
	systemPrompt, err := buildAgentSystemPrompt(agentID, a.Path, content)
	if err != nil {
		return err
	}

	userPrompt := fmt.Sprintf("'%s' 에이전트에 대한 사용법 가이드를 작성해주세요.", agentID)

	generatedContent, err := guide.RunClaudeWithSpinner(systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate guide: %w", err)
	}

	if generatedContent != "" {
		savedGuide, err := guideStore.Save(guide.TypeAgent, agentID, generatedContent)
		if err != nil {
			fmt.Printf("⚠️  가이드 저장 실패: %v\n", err)
		}

		if guideAgentsFormat == "html" {
			return guide.OpenHTMLGuide(guide.TypeAgent, agentID, generatedContent, savedGuide.CreatedAt)
		}

		guide.PrintGuide(fmt.Sprintf("Agent Guide: %s", agentID), generatedContent, savedGuide.CreatedAt, false)
	}

	return nil
}

func buildAgentSystemPrompt(agentID, agentPath, content string) (string, error) {
	promptTemplate, err := prompt.Load("guide-agent")
	if err != nil {
		return "", fmt.Errorf("failed to load guide prompt: %w", err)
	}

	tmpl, err := template.New("guide-agent").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var systemPrompt bytes.Buffer
	err = tmpl.Execute(&systemPrompt, map[string]string{
		"AgentID":   agentID,
		"AgentPath": agentPath,
		"Content":   content,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render prompt: %w", err)
	}

	return systemPrompt.String(), nil
}
