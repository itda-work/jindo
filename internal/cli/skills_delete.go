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

var skillsDeleteForce bool

var skillsDeleteCmd = &cobra.Command{
	Use:     "delete <skill-name>",
	Aliases: []string{"d", "rm"},
	Short:   "Delete a skill",
	Long: `Delete a skill from ~/.claude/skills/ directory.

This will delete the entire skill folder including all files.
Use --force to skip the confirmation prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: runSkillsDelete,
}

func init() {
	skillsCmd.AddCommand(skillsDeleteCmd)
	skillsDeleteCmd.Flags().BoolVarP(&skillsDeleteForce, "force", "f", false, "Skip confirmation prompt")
}

func runSkillsDelete(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]
	store := skill.NewStore("~/.claude/skills")

	// Get skill to verify it exists
	s, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found: %s", name)
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
