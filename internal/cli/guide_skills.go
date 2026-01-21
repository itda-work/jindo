package cli

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/itda-work/jindo/internal/guide"
	"github.com/itda-work/jindo/internal/prompt"
	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var (
	guideSkillsInteractive bool
	guideSkillsGlobal      bool
	guideSkillsLocal       bool
	guideSkillsRefresh     bool
	guideSkillsFormat      string
)

var guideSkillsCmd = &cobra.Command{
	Use:     "skills <skill-id>",
	Aliases: []string{"s", "skill"},
	Short:   "Get AI-powered usage guide for a skill",
	Long: `Get an AI-powered usage guide for a Claude Code skill.

The guide explains:
- When the skill gets triggered
- How to use it effectively
- Practical examples
- Customization suggestions

Guides are cached for future use. Use --refresh to regenerate.
Use -i for interactive mode where AI asks about your context.
Use --format html to generate HTML and open in browser.`,
	Example: `  # Get usage guide for a skill (uses cache if available)
  jd guide skills my-skill

  # Force regenerate the guide
  jd guide skills my-skill --refresh

  # Generate HTML and open in browser
  jd guide skills my-skill --format html

  # Interactive mode (not cached)
  jd guide skills my-skill -i`,
	Args:              cobra.ExactArgs(1),
	RunE:              runGuideSkills,
	ValidArgsFunction: skillNameCompletion,
}

func init() {
	guideCmd.AddCommand(guideSkillsCmd)
	guideSkillsCmd.Flags().BoolVarP(&guideSkillsInteractive, "interactive", "i", false, "Interactive mode - AI asks questions for personalized guidance")
	guideSkillsCmd.Flags().BoolVarP(&guideSkillsGlobal, "global", "g", false, "Guide from global ~/.claude/skills/")
	guideSkillsCmd.Flags().BoolVarP(&guideSkillsLocal, "local", "l", false, "Guide from local .claude/skills/")
	guideSkillsCmd.Flags().BoolVarP(&guideSkillsRefresh, "refresh", "r", false, "Regenerate the guide even if cached")
	guideSkillsCmd.Flags().StringVarP(&guideSkillsFormat, "format", "f", "", "Output format: html (opens in browser)")
}

func runGuideSkills(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if guideSkillsFormat != "" && guideSkillsFormat != "html" {
		return fmt.Errorf("invalid format: %s (use 'html')", guideSkillsFormat)
	}

	skillID := args[0]

	scope, err := ResolveScope(guideSkillsGlobal, guideSkillsLocal)
	if err != nil {
		return err
	}

	skillsDir := GetPathByScope(scope, "skills")
	store := skill.NewStore(skillsDir)

	s, err := store.Get(skillID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found in %s: %s", ScopeDescription(scope), skillID)
		}
		return fmt.Errorf("failed to get skill: %w", err)
	}

	content, err := store.GetContent(skillID)
	if err != nil {
		return fmt.Errorf("failed to read skill content: %w", err)
	}

	// Interactive mode
	if guideSkillsInteractive {
		if guideSkillsFormat == "html" {
			return fmt.Errorf("--format html cannot be used with --interactive")
		}
		systemPrompt, err := buildSkillSystemPrompt(skillID, s.Path, content)
		if err != nil {
			return err
		}
		return guide.RunInteractiveGuide(skillID, systemPrompt)
	}

	guideStore, err := guide.NewStore()
	if err != nil {
		return fmt.Errorf("failed to initialize guide store: %w", err)
	}

	// Use cache if available
	if !guideSkillsRefresh && guideStore.Exists(guide.TypeSkill, skillID) {
		cached, err := guideStore.Get(guide.TypeSkill, skillID)
		if err == nil {
			if guideSkillsFormat == "html" {
				return guide.OpenHTMLGuide(guide.TypeSkill, skillID, cached.Content, cached.CreatedAt)
			}
			guide.PrintGuide(fmt.Sprintf("Skill Guide: %s", skillID), cached.Content, cached.CreatedAt, true)
			return nil
		}
	}

	// Generate new guide
	systemPrompt, err := buildSkillSystemPrompt(skillID, s.Path, content)
	if err != nil {
		return err
	}

	userPrompt := fmt.Sprintf("'%s' 스킬에 대한 사용법 가이드를 작성해주세요.", skillID)

	generatedContent, err := guide.RunClaudeWithSpinner(systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate guide: %w", err)
	}

	if generatedContent != "" {
		savedGuide, err := guideStore.Save(guide.TypeSkill, skillID, generatedContent)
		if err != nil {
			fmt.Printf("⚠️  가이드 저장 실패: %v\n", err)
		}

		if guideSkillsFormat == "html" {
			return guide.OpenHTMLGuide(guide.TypeSkill, skillID, generatedContent, savedGuide.CreatedAt)
		}

		guide.PrintGuide(fmt.Sprintf("Skill Guide: %s", skillID), generatedContent, savedGuide.CreatedAt, false)
	}

	return nil
}

func buildSkillSystemPrompt(skillID, skillPath, content string) (string, error) {
	promptTemplate, err := prompt.Load("guide-skill")
	if err != nil {
		return "", fmt.Errorf("failed to load guide prompt: %w", err)
	}

	tmpl, err := template.New("guide-skill").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var systemPrompt bytes.Buffer
	err = tmpl.Execute(&systemPrompt, map[string]string{
		"SkillID":   skillID,
		"SkillPath": skillPath,
		"Content":   content,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render prompt: %w", err)
	}

	return systemPrompt.String(), nil
}
