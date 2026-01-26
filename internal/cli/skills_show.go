package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itda-skills/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var (
	skillsShowBrief  bool
	skillsShowGlobal bool
	skillsShowLocal  bool
)

var skillsShowCmd = &cobra.Command{
	Use:     "show <skill-name>",
	Aliases: []string{"s", "get", "view"},
	Short:   "Show skill details",
	Long: `Show the full content of a specific skill from ~/.claude/skills/ (global) or .claude/skills/ (local) directory.

Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args:              cobra.ExactArgs(1),
	RunE:              runSkillsShow,
	ValidArgsFunction: skillNameCompletion,
}

func init() {
	skillsCmd.AddCommand(skillsShowCmd)
	skillsShowCmd.Flags().BoolVar(&skillsShowBrief, "brief", false, "Show only frontmatter (name, description, allowed-tools)")
	skillsShowCmd.Flags().BoolVarP(&skillsShowGlobal, "global", "g", false, "Show from global ~/.claude/skills/")
	skillsShowCmd.Flags().BoolVarP(&skillsShowLocal, "local", "l", false, "Show from local .claude/skills/")
}

func runSkillsShow(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]

	scope, err := ResolveScope(skillsShowGlobal, skillsShowLocal)
	if err != nil {
		return err
	}

	store := skill.NewStore(GetPathByScope(scope, "skills"))

	if skillsShowBrief {
		return showSkillBrief(store, name, scope)
	}

	return showSkillFull(store, name, scope)
}

func showSkillBrief(store *skill.Store, name string, scope PathScope) error {
	s, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get skill: %w", err)
	}

	fmt.Printf("Name:          %s\n", s.Name)
	fmt.Printf("Description:   %s\n", s.Description)
	fmt.Printf("Allowed Tools: %s\n", strings.Join(s.AllowedTools, ", "))
	fmt.Printf("Path:          %s\n", s.Path)

	return nil
}

func showSkillFull(store *skill.Store, name string, scope PathScope) error {
	content, err := store.GetContent(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get skill content: %w", err)
	}

	fmt.Print(content)
	return nil
}

// skillNameCompletion provides completion for skill names
func skillNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	global, _ := cmd.Flags().GetBool("global")
	local, _ := cmd.Flags().GetBool("local")
	scope, err := ResolveScope(global, local)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	store := skill.NewStore(GetPathByScope(scope, "skills"))
	skills, err := store.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, s := range skills {
		// Use directory name (the actual ID used for lookup), not frontmatter name
		dirName := filepath.Base(filepath.Dir(s.Path))
		if s.Description != "" {
			names = append(names, fmt.Sprintf("%s\t%s", dirName, s.Description))
		} else {
			names = append(names, dirName)
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
