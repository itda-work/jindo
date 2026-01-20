package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/itda-work/itda-jindo/internal/skill"
	"github.com/spf13/cobra"
)

var skillsEditEditor bool

var skillsEditCmd = &cobra.Command{
	Use:   "edit <skill-name>",
	Short: "Edit an existing skill",
	Long: `Edit an existing skill in ~/.claude/skills/ directory.

By default, uses Claude CLI to interactively edit the skill content.
Use --editor to open the skill file directly in your editor.`,
	Args: cobra.ExactArgs(1),
	RunE: runSkillsEdit,
}

func init() {
	skillsCmd.AddCommand(skillsEditCmd)
	skillsEditCmd.Flags().BoolVarP(&skillsEditEditor, "editor", "e", false, "Open in editor directly (skip AI)")
}

func runSkillsEdit(_ *cobra.Command, args []string) error {
	name := args[0]
	store := skill.NewStore("~/.claude/skills")

	// Get skill to verify it exists and get its path
	s, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found: %s", name)
		}
		return fmt.Errorf("failed to get skill: %w", err)
	}

	// If --editor flag, just open in editor
	if skillsEditEditor {
		return openEditor(s.Path)
	}

	// Get current content for context
	content, err := store.GetContent(name)
	if err != nil {
		return fmt.Errorf("failed to read skill content: %w", err)
	}

	// Use Claude CLI to edit
	newContent, err := editSkillWithClaude(name, content)
	if err != nil {
		return fmt.Errorf("failed to edit skill with Claude: %w", err)
	}

	// Write updated content
	if err := os.WriteFile(s.Path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	fmt.Printf("✅ Updated skill: %s\n", s.Path)
	return nil
}

func editSkillWithClaude(name, currentContent string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are helping edit a Claude Code skill named "%s".

Current skill content:
---
%s
---

Help the user modify this skill. When they describe the changes they want:
1. Understand what they want to change
2. Generate the complete updated SKILL.md content

The output must be a valid SKILL.md file with:
- YAML frontmatter (name, description, allowed-tools)
- Markdown content

Ask the user what changes they want to make to this skill.`, name, currentContent)

	cmd := exec.Command("claude",
		"--print",
		"--system-prompt", systemPrompt,
		fmt.Sprintf("I want to edit the '%s' skill. Here's the current content. What would you like to change?", name),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
