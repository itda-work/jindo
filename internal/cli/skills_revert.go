package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/itda-work/jindo/internal/skill"
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
	skillsRevertCmd.Flags().BoolVarP(&skillsRevertGlobal, "global", "g", false, "Revert from global ~/.claude/skills/ (default)")
	skillsRevertCmd.Flags().BoolVarP(&skillsRevertLocal, "local", "l", false, "Revert from local .claude/skills/")
}

func runSkillsRevert(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Validate mutually exclusive flags
	if err := ValidateScopeFlags(skillsRevertGlobal, skillsRevertLocal); err != nil {
		return err
	}

	skillID := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if skillsRevertLocal {
		scope = ScopeLocal
	}

	skillsDir := GetPathByScope(scope, "skills")
	store := skill.NewStore(skillsDir)

	// Get skill to verify it exists and get its path
	s, err := store.Get(skillID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found: %s", skillID)
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

		fmt.Printf("Available versions for skill: %s\n\n", skillID)
		for _, v := range versions {
			fmt.Printf("  %s\n", skill.FormatVersionName(&v))
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

	// Backup current content before reverting
	currentContent, err := store.GetContent(skillID)
	if err != nil {
		return fmt.Errorf("failed to read current content: %w", err)
	}

	_, err = historyMgr.SaveVersion(currentContent)
	if err != nil {
		return fmt.Errorf("failed to backup current version: %w", err)
	}

	// Write the reverted content
	if err := os.WriteFile(s.Path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write reverted content: %w", err)
	}

	fmt.Printf("✅ Reverted skill '%s' to %s\n", skillID, skill.FormatVersionName(version))
	fmt.Println("\nCurrent content has been backed up to history.")

	return nil
}
