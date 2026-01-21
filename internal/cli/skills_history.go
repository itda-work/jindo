package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var (
	skillsHistoryGlobal bool
	skillsHistoryLocal  bool
)

var skillsHistoryCmd = &cobra.Command{
	Use:     "history <skill-id>",
	Aliases: []string{"hist"},
	Short:   "Show version history of a skill",
	Long: `Show the version history of a skill.

Each time a skill is adapted, a new version is saved to .history/.
Use 'jd skills revert' to restore a previous version.`,
	Example: `  # Show history of a global skill
  jd skills history my-skill

  # Show history of a local skill
  jd skills history my-skill --local`,
	Args:              cobra.ExactArgs(1),
	RunE:              runSkillsHistory,
	ValidArgsFunction: skillNameCompletion,
}

func init() {
	skillsCmd.AddCommand(skillsHistoryCmd)
	skillsHistoryCmd.Flags().BoolVarP(&skillsHistoryGlobal, "global", "g", false, "Show from global ~/.claude/skills/")
	skillsHistoryCmd.Flags().BoolVarP(&skillsHistoryLocal, "local", "l", false, "Show from local .claude/skills/")
}

func runSkillsHistory(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	skillID := args[0]

	scope, err := ResolveScope(skillsHistoryGlobal, skillsHistoryLocal)
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

	versions, err := historyMgr.ListVersions()
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Printf("No history found for skill: %s\n", skillID)
		fmt.Println("\nHistory is created when you use 'jd skills adapt'.")
		return nil
	}

	fmt.Printf("Version history for skill: %s\n", skillID)
	fmt.Printf("Path: %s\n\n", s.Path)

	for i, v := range versions {
		marker := "  "
		if i == 0 {
			marker = "* " // Mark the latest
		}
		fmt.Printf("%s%s\n", marker, skill.FormatVersionName(&v))
	}

	fmt.Printf("\nTotal: %d version(s)\n", len(versions))
	fmt.Printf("\nTo revert: jd skills revert %s <version>\n", skillID)

	return nil
}
