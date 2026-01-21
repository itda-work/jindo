package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var skillsListJSON bool

var skillsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List all skills",
	Long:    `List all skills from ~/.claude/skills/ directory.`,
	RunE:    runSkillsList,
}

func init() {
	skillsCmd.AddCommand(skillsListCmd)
	skillsListCmd.Flags().BoolVar(&skillsListJSON, "json", false, "Output in JSON format")
}

func runSkillsList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	store := skill.NewStore("~/.claude/skills")
	skills, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list skills: %w", err)
	}

	if len(skills) == 0 {
		fmt.Println("No skills found.")
		return nil
	}

	if skillsListJSON {
		return printSkillsJSON(skills)
	}

	printSkillsTable(skills)
	return nil
}

func printSkillsJSON(skills []*skill.Skill) error {
	output, err := json.MarshalIndent(skills, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func printSkillsTable(skills []*skill.Skill) {
	// Calculate column widths
	nameWidth := len("NAME")
	descWidth := len("DESCRIPTION")
	toolsWidth := len("ALLOWED-TOOLS")

	for _, s := range skills {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
		// Truncate description for display
		desc := s.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		if len(desc) > descWidth {
			descWidth = len(desc)
		}
		tools := strings.Join(s.AllowedTools, ", ")
		if len(tools) > toolsWidth {
			toolsWidth = len(tools)
		}
	}

	// Cap widths
	if nameWidth > 25 {
		nameWidth = 25
	}
	if descWidth > 50 {
		descWidth = 50
	}
	if toolsWidth > 30 {
		toolsWidth = 30
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s\n",
		nameWidth, "NAME",
		descWidth, "DESCRIPTION",
		toolsWidth, "ALLOWED-TOOLS")
	fmt.Printf("%s  %s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", descWidth),
		strings.Repeat("-", toolsWidth))

	// Print rows
	for _, s := range skills {
		name := s.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		desc := s.Description
		if len(desc) > descWidth {
			desc = desc[:descWidth-3] + "..."
		}

		tools := strings.Join(s.AllowedTools, ", ")
		if len(tools) > toolsWidth {
			tools = tools[:toolsWidth-3] + "..."
		}

		fmt.Printf("%-*s  %-*s  %-*s\n",
			nameWidth, name,
			descWidth, desc,
			toolsWidth, tools)
	}

	fmt.Printf("\nTotal: %d skills\n", len(skills))
}
