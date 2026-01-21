package cli

import (
	"encoding/json"
	"fmt"

	"github.com/itda-work/jindo/internal/agent"
	"github.com/itda-work/jindo/internal/command"
	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var listJSON bool

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List all skills, agents, and commands",
	Long:    `List all configured skills, agents, and commands from ~/.claude/ directory.`,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
}

type listItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type listOutput struct {
	Skills   []listItem `json:"skills"`
	Agents   []listItem `json:"agents"`
	Commands []listItem `json:"commands"`
}

func runList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	skillStore := skill.NewStore("~/.claude/skills")
	agentStore := agent.NewStore("~/.claude/agents")
	commandStore := command.NewStore("~/.claude/commands")

	skills, err := skillStore.List()
	if err != nil {
		skills = nil
	}

	agents, err := agentStore.List()
	if err != nil {
		agents = nil
	}

	commands, err := commandStore.List()
	if err != nil {
		commands = nil
	}

	if listJSON {
		return printListJSON(skills, agents, commands)
	}

	// Use each type's table output format
	fmt.Println("=== Skills ===")
	if len(skills) == 0 {
		fmt.Println("No skills found.")
	} else {
		printSkillsTable(skills)
	}
	fmt.Println()

	fmt.Println("=== Agents ===")
	if len(agents) == 0 {
		fmt.Println("No agents found.")
	} else {
		printAgentsTable(agents)
	}
	fmt.Println()

	fmt.Println("=== Commands ===")
	if len(commands) == 0 {
		fmt.Println("No commands found.")
	} else {
		printCommandsTable(commands)
	}

	return nil
}

func printListJSON(skills []*skill.Skill, agents []*agent.Agent, commands []*command.Command) error {
	output := listOutput{
		Skills:   make([]listItem, 0, len(skills)),
		Agents:   make([]listItem, 0, len(agents)),
		Commands: make([]listItem, 0, len(commands)),
	}

	for _, s := range skills {
		output.Skills = append(output.Skills, listItem{Name: s.Name, Description: s.Description})
	}
	for _, a := range agents {
		output.Agents = append(output.Agents, listItem{Name: a.Name, Description: a.Description})
	}
	for _, c := range commands {
		output.Commands = append(output.Commands, listItem{Name: c.Name, Description: c.Description})
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonOutput))
	return nil
}
