package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var skillsShowBrief bool

var skillsShowCmd = &cobra.Command{
	Use:     "show <skill-name>",
	Aliases: []string{"s"},
	Short:   "Show skill details",
	Long:  `Show the full content of a specific skill from ~/.claude/skills/ directory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillsShow,
}

func init() {
	skillsCmd.AddCommand(skillsShowCmd)
	skillsShowCmd.Flags().BoolVar(&skillsShowBrief, "brief", false, "Show only frontmatter (name, description, allowed-tools)")
}

func runSkillsShow(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]
	store := skill.NewStore("~/.claude/skills")

	if skillsShowBrief {
		return showSkillBrief(store, name)
	}

	return showSkillFull(store, name)
}

func showSkillBrief(store *skill.Store, name string) error {
	s, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found: %s", name)
		}
		return fmt.Errorf("failed to get skill: %w", err)
	}

	fmt.Printf("Name:          %s\n", s.Name)
	fmt.Printf("Description:   %s\n", s.Description)
	fmt.Printf("Allowed Tools: %s\n", strings.Join(s.AllowedTools, ", "))
	fmt.Printf("Path:          %s\n", s.Path)

	return nil
}

func showSkillFull(store *skill.Store, name string) error {
	content, err := store.GetContent(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found: %s", name)
		}
		return fmt.Errorf("failed to get skill content: %w", err)
	}

	fmt.Print(content)
	return nil
}
