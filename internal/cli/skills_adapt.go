package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/itda-work/jindo/internal/prompt"
	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var (
	skillsAdaptGlobal bool
	skillsAdaptLocal  bool
)

var skillsAdaptCmd = &cobra.Command{
	Use:   "adapt <skill-id>",
	Short: "Customize a skill using AI conversation",
	Long: `Customize a skill to fit your specific workflow using AI-powered conversation.

This command:
1. Backs up the current version to .history/
2. Starts an AI conversation to understand your needs
3. Modifies the skill based on the conversation
4. Saves changes and updates version history

Use --local to adapt from the current directory's .claude/skills/.`,
	Example: `  # Adapt a global skill
  jd skills adapt my-skill

  # Adapt a local skill
  jd skills adapt my-skill --local`,
	Args:              cobra.ExactArgs(1),
	RunE:              runSkillsAdapt,
	ValidArgsFunction: skillNameCompletion,
}

func init() {
	skillsCmd.AddCommand(skillsAdaptCmd)
	skillsAdaptCmd.Flags().BoolVarP(&skillsAdaptGlobal, "global", "g", false, "Adapt from global ~/.claude/skills/ (default)")
	skillsAdaptCmd.Flags().BoolVarP(&skillsAdaptLocal, "local", "l", false, "Adapt from local .claude/skills/")
}

func runSkillsAdapt(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Validate mutually exclusive flags
	if err := ValidateScopeFlags(skillsAdaptGlobal, skillsAdaptLocal); err != nil {
		return err
	}

	skillID := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if skillsAdaptLocal {
		scope = ScopeLocal
	}

	skillsDir := GetPathByScope(scope, "skills")
	store := skill.NewStore(skillsDir)

	// Get skill to verify it exists
	s, err := store.Get(skillID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found: %s", skillID)
		}
		return fmt.Errorf("failed to get skill: %w", err)
	}

	// Get current content
	content, err := store.GetContent(skillID)
	if err != nil {
		return fmt.Errorf("failed to read skill content: %w", err)
	}

	// Create history manager and backup current version
	skillDir := filepath.Dir(s.Path)
	historyMgr := skill.NewHistoryManager(skillDir)

	version, err := historyMgr.SaveVersion(content)
	if err != nil {
		return fmt.Errorf("failed to backup current version: %w", err)
	}
	fmt.Printf("📦 Backed up to %s\n", skill.FormatVersionName(version))

	// Load and render the adapt prompt
	promptTemplate, err := prompt.Load("adapt-skill")
	if err != nil {
		return fmt.Errorf("failed to load adapt prompt: %w", err)
	}

	tmpl, err := template.New("adapt-skill").Parse(promptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var systemPrompt bytes.Buffer
	err = tmpl.Execute(&systemPrompt, map[string]string{
		"SkillID":   skillID,
		"SkillPath": s.Path,
		"Content":   content,
	})
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	// Show tip about customizing the prompt
	fmt.Println()
	fmt.Printf("💡 Tip: Customize this prompt with: jd prompts edit adapt-skill\n")
	fmt.Println()

	// Run claude command with the system prompt
	claudeCmd := exec.Command("claude",
		"--system-prompt", systemPrompt.String(),
		"--allowedTools", "Edit,Read,Write,Glob,Grep",
	)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	if err := claudeCmd.Run(); err != nil {
		// Check if it's just a user exit
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 { // Ctrl+C
				fmt.Println("\n⚠️  Adaptation cancelled")
				return nil
			}
		}
		return fmt.Errorf("claude command failed: %w", err)
	}

	// Read the potentially updated content
	newContent, err := store.GetContent(skillID)
	if err != nil {
		return fmt.Errorf("failed to read updated skill: %w", err)
	}

	// Check if content changed
	if newContent == content {
		fmt.Println("\n📝 No changes made to the skill")
		return nil
	}

	// Save new version
	newVersion, err := historyMgr.SaveVersion(newContent)
	if err != nil {
		return fmt.Errorf("failed to save new version: %w", err)
	}

	fmt.Printf("\n✅ Skill adapted successfully!\n")
	fmt.Printf("   Previous: %s\n", skill.FormatVersionName(version))
	fmt.Printf("   Current:  %s\n", skill.FormatVersionName(newVersion))
	fmt.Printf("\n   To revert: jd skills revert %s %d\n", skillID, version.Number)

	return nil
}
