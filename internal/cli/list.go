package cli

import (
	"encoding/json"
	"fmt"

	"github.com/itda-work/jindo/internal/agent"
	"github.com/itda-work/jindo/internal/command"
	"github.com/itda-work/jindo/internal/hook"
	"github.com/itda-work/jindo/internal/skill"
	"github.com/spf13/cobra"
)

var listJSON bool

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List all skills, agents, commands, and hooks",
	Long:    `List all configured skills, agents, commands, and hooks from ~/.claude/ and .claude/ directories.`,
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

type scopedListOutput struct {
	Skills   []listItem `json:"skills"`
	Agents   []listItem `json:"agents"`
	Commands []listItem `json:"commands"`
	Hooks    []listItem `json:"hooks"`
}

type listOutput struct {
	Global scopedListOutput `json:"global"`
	Local  scopedListOutput `json:"local,omitempty"`
}

func runList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	// Get global items
	globalSkillStore := skill.NewStore(GetGlobalPath("skills"))
	globalAgentStore := agent.NewStore(GetGlobalPath("agents"))
	globalCommandStore := command.NewStore(GetGlobalPath("commands"))
	globalHookStore := hook.NewStore(GetSettingsPathByScope(ScopeGlobal))

	globalSkills, _ := globalSkillStore.List()
	globalAgents, _ := globalAgentStore.List()
	globalCommands, _ := globalCommandStore.List()
	globalHooks, _ := globalHookStore.List()

	// Get local items (if .claude exists)
	var localSkills []*skill.Skill
	var localAgents []*agent.Agent
	var localCommands []*command.Command
	var localHooks []*hook.Hook

	if localPath := GetLocalPath("skills"); localPath != "" {
		localSkillStore := skill.NewStore(localPath)
		localSkills, _ = localSkillStore.List()
	}
	if localPath := GetLocalPath("agents"); localPath != "" {
		localAgentStore := agent.NewStore(localPath)
		localAgents, _ = localAgentStore.List()
	}
	if localPath := GetLocalPath("commands"); localPath != "" {
		localCommandStore := command.NewStore(localPath)
		localCommands, _ = localCommandStore.List()
	}
	if localSettingsPath := GetLocalSettingsPath(); localSettingsPath != "" {
		localHookStore := hook.NewStore(localSettingsPath)
		localHooks, _ = localHookStore.List()
	}

	hasLocal := len(localSkills) > 0 || len(localAgents) > 0 || len(localCommands) > 0 || len(localHooks) > 0

	if listJSON {
		return printListJSON(globalSkills, globalAgents, globalCommands, globalHooks, localSkills, localAgents, localCommands, localHooks)
	}

	// Print Global section
	fmt.Println("=== Global (~/.claude/) ===")
	fmt.Println()

	fmt.Println("Skills:")
	if len(globalSkills) == 0 {
		fmt.Println("  No skills found.")
	} else {
		printSkillsTable(globalSkills)
	}
	fmt.Println()

	fmt.Println("Agents:")
	if len(globalAgents) == 0 {
		fmt.Println("  No agents found.")
	} else {
		printAgentsTable(globalAgents)
	}
	fmt.Println()

	fmt.Println("Commands:")
	if len(globalCommands) == 0 {
		fmt.Println("  No commands found.")
	} else {
		printCommandsTable(globalCommands)
	}

	fmt.Println()

	fmt.Println("Hooks:")
	if len(globalHooks) == 0 {
		fmt.Println("  No hooks found.")
	} else {
		printHooksTable(globalHooks)
	}

	// Print Local section only if has items
	if hasLocal {
		fmt.Println()
		fmt.Println("=== Local (.claude/) ===")
		fmt.Println()

		if len(localSkills) > 0 {
			fmt.Println("Skills:")
			printSkillsTable(localSkills)
			fmt.Println()
		}

		if len(localAgents) > 0 {
			fmt.Println("Agents:")
			printAgentsTable(localAgents)
			fmt.Println()
		}

		if len(localCommands) > 0 {
			fmt.Println("Commands:")
			printCommandsTable(localCommands)
			fmt.Println()
		}

		if len(localHooks) > 0 {
			fmt.Println("Hooks:")
			printHooksTable(localHooks)
		}
	}

	return nil
}

func printListJSON(globalSkills []*skill.Skill, globalAgents []*agent.Agent, globalCommands []*command.Command, globalHooks []*hook.Hook,
	localSkills []*skill.Skill, localAgents []*agent.Agent, localCommands []*command.Command, localHooks []*hook.Hook) error {

	toListItems := func(skills []*skill.Skill, agents []*agent.Agent, commands []*command.Command, hooks []*hook.Hook) scopedListOutput {
		output := scopedListOutput{
			Skills:   make([]listItem, 0, len(skills)),
			Agents:   make([]listItem, 0, len(agents)),
			Commands: make([]listItem, 0, len(commands)),
			Hooks:    make([]listItem, 0, len(hooks)),
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
		for _, h := range hooks {
			desc := fmt.Sprintf("%s: %s", h.EventType, h.Matcher)
			output.Hooks = append(output.Hooks, listItem{Name: h.Name, Description: desc})
		}
		return output
	}

	output := listOutput{
		Global: toListItems(globalSkills, globalAgents, globalCommands, globalHooks),
		Local:  toListItems(localSkills, localAgents, localCommands, localHooks),
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonOutput))
	return nil
}
