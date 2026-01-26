package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/itda-skills/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var (
	skillsRevertGlobal bool
	skillsRevertLocal  bool
)

var skillsRevertCmd = &cobra.Command{
	Use:   "revert <skill-id> [version]",
	Short: "Revert a skill to a previous version",
	Long: `Revert a skill to a previous version from its history.

If no version is specified, shows available versions.
Version can be a number (e.g., 1, 2) or 'latest'.`,
	Example: `  # Show available versions
  jd skills revert my-skill

  # Revert to version 1
  jd skills revert my-skill 1

  # Revert to the latest backed up version
  jd skills revert my-skill latest`,
	Args:              cobra.RangeArgs(1, 2),
	RunE:              runSkillsRevert,
	ValidArgsFunction: skillNameCompletion,
}

func init() {
	skillsCmd.AddCommand(skillsRevertCmd)
	skillsRevertCmd.Flags().BoolVarP(&skillsRevertGlobal, "global", "g", false, "Revert from global ~/.claude/skills/")
	skillsRevertCmd.Flags().BoolVarP(&skillsRevertLocal, "local", "l", false, "Revert from local .claude/skills/")
}

func runSkillsRevert(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	skillID := args[0]

	scope, err := ResolveScope(skillsRevertGlobal, skillsRevertLocal)
	if err != nil {
		return err
	}

	skillsDir := GetPathByScope(scope, "skills")
	store := skill.NewStore(skillsDir)

	// Get skill to verify it exists and get its path
	s, err := store.Get(skillID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found in %s: %s", ScopeDescription(scope), skillID)
		}
		return fmt.Errorf("failed to get skill: %w", err)
	}

	// Create history manager
	skillDir := filepath.Dir(s.Path)
	historyMgr := skill.NewHistoryManager(skillDir)

	// If no version specified, show available versions
	if len(args) < 2 {
		versions, err := historyMgr.ListVersions()
		if err != nil {
			return fmt.Errorf("failed to list versions: %w", err)
		}

		if len(versions) == 0 {
			fmt.Printf("No history found for skill: %s\n", skillID)
			return nil
		}

		// Get current content to find active version
		currentContent, _ := store.GetContent(skillID)

		fmt.Printf("Available versions for skill: %s\n\n", skillID)
		for _, v := range versions {
			marker := "  "
			// Check if this version matches current content
			if vContent, _, err := historyMgr.GetVersion(v.Number); err == nil && vContent == currentContent {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, skill.FormatVersionName(&v))
		}
		fmt.Printf("\nUsage: jd skills revert %s <version>\n", skillID)
		return nil
	}

	// Parse version argument
	versionArg := args[1]
	versionNum, err := skill.ParseVersionArg(versionArg)
	if err != nil {
		return err
	}

	var content string
	var version *skill.Version

	if versionNum == -1 {
		// Get latest version
		version, err = historyMgr.GetLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
		content, _, err = historyMgr.GetVersion(version.Number)
	} else {
		content, version, err = historyMgr.GetVersion(versionNum)
	}

	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// Write the reverted content
	if err := os.WriteFile(s.Path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write reverted content: %w", err)
	}

	// Delete all versions after the reverted version
	deleted, err := historyMgr.DeleteVersionsAfter(version.Number)
	if err != nil {
		return fmt.Errorf("failed to cleanup versions: %w", err)
	}

	fmt.Printf("✅ Reverted skill '%s' to %s\n", skillID, skill.FormatVersionName(version))
	if deleted > 0 {
		fmt.Printf("   Removed %d newer version(s)\n", deleted)
	}

	return nil
}
