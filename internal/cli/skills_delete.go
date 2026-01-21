package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var (
	skillsDeleteForce  bool
	skillsDeleteGlobal bool
	skillsDeleteLocal  bool
)

var skillsDeleteCmd = &cobra.Command{
	Use:     "delete <skill-name>",
	Aliases: []string{"d", "rm", "remove"},
	Short:   "Delete a skill",
	Long: `Delete a skill from ~/.claude/skills/ (global) or .claude/skills/ (local) directory.

This will delete the entire skill folder including all files.
Use --force to skip the confirmation prompt.
Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args:              cobra.ExactArgs(1),
	RunE:              runSkillsDelete,
	ValidArgsFunction: skillNameCompletion,
}

func init() {
	skillsCmd.AddCommand(skillsDeleteCmd)
	skillsDeleteCmd.Flags().BoolVarP(&skillsDeleteForce, "force", "f", false, "Skip confirmation prompt")
	skillsDeleteCmd.Flags().BoolVarP(&skillsDeleteGlobal, "global", "g", false, "Delete from global ~/.claude/skills/")
	skillsDeleteCmd.Flags().BoolVarP(&skillsDeleteLocal, "local", "l", false, "Delete from local .claude/skills/")
}

func runSkillsDelete(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	scope, err := ResolveScope(skillsDeleteGlobal, skillsDeleteLocal)
	if err != nil {
		return err
	}

	name := args[0]

	store := skill.NewStore(GetPathByScope(scope, "skills"))

	// Get skill to verify it exists
	s, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get skill: %w", err)
	}

	// Get the skill directory (parent of SKILL.md)
	skillDir := filepath.Dir(s.Path)

	// Confirm deletion unless --force
	if !skillsDeleteForce {
		fmt.Printf("Delete skill '%s'?\n", name)
		fmt.Printf("  Path: %s\n", skillDir)
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete the skill directory
	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to delete skill: %w", err)
	}

	fmt.Printf("Deleted skill: %s\n", name)
	return nil
}
