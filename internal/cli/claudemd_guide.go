package cli

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/itda-skills/jindo/internal/guide"
	"github.com/itda-skills/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	claudemdGuideInteractive bool
	claudemdGuideGlobal      bool
	claudemdGuideLocal       bool
	claudemdGuideRefresh     bool
	claudemdGuideFormat      string
	claudemdGuideAnalyze     bool
	claudemdGuideTemplate    bool
)

var claudemdGuideCmd = &cobra.Command{
	Use:     "guide",
	Aliases: []string{"g"},
	Short:   "AI-powered CLAUDE.md best practices guide",
	Long: `Get an AI-powered guide for writing effective CLAUDE.md files.

The guide provides:
- Best practices for structuring CLAUDE.md
- Writing effective instructions
- Common patterns and anti-patterns
- Templates for different project types

Use --analyze to get improvement suggestions for your current CLAUDE.md.
Use --template to get ready-to-use templates.
Use -i for interactive mode where AI asks about your context.`,
	Example: `  # Get general CLAUDE.md best practices guide
  jd claudemd guide

  # Analyze current CLAUDE.md and get improvement suggestions
  jd claudemd guide --analyze

  # Analyze local CLAUDE.md specifically
  jd claudemd guide --analyze --local

  # Get ready-to-use templates
  jd claudemd guide --template

  # Interactive mode for personalized guidance
  jd claudemd guide -i

  # Generate HTML and open in browser
  jd claudemd guide --format html

  # Force regenerate the guide
  jd claudemd guide --refresh`,
	RunE: runClaudemdGuide,
}

func init() {
	claudemdCmd.AddCommand(claudemdGuideCmd)

	claudemdGuideCmd.Flags().BoolVarP(&claudemdGuideInteractive, "interactive", "i", false, "Interactive mode - AI asks questions for personalized guidance")
	claudemdGuideCmd.Flags().BoolVarP(&claudemdGuideGlobal, "global", "g", false, "Analyze global ~/.claude/CLAUDE.md")
	claudemdGuideCmd.Flags().BoolVarP(&claudemdGuideLocal, "local", "l", false, "Analyze local .claude/CLAUDE.md")
	claudemdGuideCmd.Flags().BoolVarP(&claudemdGuideRefresh, "refresh", "r", false, "Regenerate the guide even if cached")
	claudemdGuideCmd.Flags().StringVarP(&claudemdGuideFormat, "format", "f", "", "Output format: html (opens in browser)")
	claudemdGuideCmd.Flags().BoolVarP(&claudemdGuideAnalyze, "analyze", "a", false, "Analyze current CLAUDE.md and suggest improvements")
	claudemdGuideCmd.Flags().BoolVarP(&claudemdGuideTemplate, "template", "t", false, "Show ready-to-use CLAUDE.md templates")
}

func runClaudemdGuide(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	if claudemdGuideFormat != "" && claudemdGuideFormat != "html" {
		return fmt.Errorf("invalid format: %s (use 'html')", claudemdGuideFormat)
	}

	// Validate mutually exclusive flags
	if claudemdGuideAnalyze && claudemdGuideTemplate {
		return fmt.Errorf("--analyze and --template cannot be used together")
	}

	// Check Claude CLI installed
	if err := checkClaudeInstalled(); err != nil {
		return err
	}

	// Determine mode
	mode := "general"
	if claudemdGuideAnalyze {
		mode = "analyze"
	} else if claudemdGuideTemplate {
		mode = "template"
	}

	// For analyze mode, read current CLAUDE.md
	var claudemdContent string
	if claudemdGuideAnalyze {
		scope, err := ResolveScope(claudemdGuideGlobal, claudemdGuideLocal)
		if err != nil {
			return err
		}

		claudemdPath := getCLAUDEmdPath(scope)
		content, err := os.ReadFile(claudemdPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("CLAUDE.md not found at %s\nCreate one first or use a different scope (--global/--local)", claudemdPath)
			}
			return fmt.Errorf("failed to read CLAUDE.md: %w", err)
		}
		claudemdContent = string(content)
		fmt.Printf("📄 분석 대상: %s\n\n", claudemdPath)
	}

	// Interactive mode
	if claudemdGuideInteractive {
		if claudemdGuideFormat == "html" {
			return fmt.Errorf("--format html cannot be used with --interactive")
		}
		systemPrompt, err := buildClaudemdGuideSystemPrompt(mode, claudemdContent)
		if err != nil {
			return err
		}
		return guide.RunInteractiveGuide("CLAUDE.md", systemPrompt)
	}

	guideStore, err := guide.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize guide store: %w", err)
	}

	// Cache key includes mode
	cacheKey := fmt.Sprintf("claudemd-%s", mode)

	// Use cache if available (only for general and template modes, not analyze)
	if !claudemdGuideRefresh && mode != "analyze" && guideStore.Exists(guide.TypeClaudemd, cacheKey) {
		cached, err := guideStore.Get(guide.TypeClaudemd, cacheKey)
		if err == nil {
			if claudemdGuideFormat == "html" {
				return guide.OpenHTMLGuide(guide.TypeClaudemd, cacheKey, cached.Content, cached.CreatedAt)
			}
			guide.PrintGuide(getGuideTitle(mode), cached.Content, cached.CreatedAt, true)
			return nil
		}
	}

	// Generate new guide
	systemPrompt, err := buildClaudemdGuideSystemPrompt(mode, claudemdContent)
	if err != nil {
		return err
	}

	userPrompt := getGuideUserPrompt(mode)

	generatedContent, err := guide.RunClaudeWithSpinner(systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate guide: %w", err)
	}

	if generatedContent != "" {
		// Save to cache (skip for analyze mode as content is context-specific)
		var createdAt = guide.Guide{}.CreatedAt
		if mode != "analyze" {
			savedGuide, err := guideStore.Save(guide.TypeClaudemd, cacheKey, generatedContent)
			if err != nil {
				fmt.Printf("⚠️  가이드 저장 실패: %v\n", err)
			} else {
				createdAt = savedGuide.CreatedAt
			}
		}

		if claudemdGuideFormat == "html" {
			return guide.OpenHTMLGuide(guide.TypeClaudemd, cacheKey, generatedContent, createdAt)
		}

		guide.PrintGuide(getGuideTitle(mode), generatedContent, createdAt, mode != "analyze" && !claudemdGuideRefresh)
	}

	return nil
}

func buildClaudemdGuideSystemPrompt(mode, content string) (string, error) {
	promptTemplate, err := prompt.Load("guide-claudemd")
	if err != nil {
		return "", fmt.Errorf("failed to load guide prompt: %w", err)
	}

	tmpl, err := template.New("guide-claudemd").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var systemPrompt bytes.Buffer
	err = tmpl.Execute(&systemPrompt, map[string]string{
		"Mode":    mode,
		"Content": content,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render prompt: %w", err)
	}

	return systemPrompt.String(), nil
}

func getGuideTitle(mode string) string {
	switch mode {
	case "analyze":
		return "CLAUDE.md 분석 결과"
	case "template":
		return "CLAUDE.md 템플릿"
	default:
		return "CLAUDE.md 베스트 프랙티스 가이드"
	}
}

func getGuideUserPrompt(mode string) string {
	switch mode {
	case "analyze":
		return "현재 CLAUDE.md를 분석하고 구체적인 개선점을 제안해주세요."
	case "template":
		return "다양한 프로젝트 유형에 맞는 CLAUDE.md 템플릿을 제공해주세요."
	default:
		return "CLAUDE.md 작성에 대한 베스트 프랙티스 가이드를 작성해주세요."
	}
}
