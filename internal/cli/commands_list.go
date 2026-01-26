package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-skills/jindo/internal/command"
	"github.com/spf13/cobra"
)

var commandsListJSON bool

var commandsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List all commands",
	Long:    `List all commands from ~/.claude/commands/ and .claude/commands/ directories.`,
	RunE:    runCommandsList,
}

func init() {
	commandsCmd.AddCommand(commandsListCmd)
	commandsListCmd.Flags().BoolVar(&commandsListJSON, "json", false, "Output in JSON format")
}

// commandsListOutput represents JSON output for commands list with scope
type commandsListOutput struct {
	Global []*command.Command `json:"global"`
	Local  []*command.Command `json:"local,omitempty"`
}

func runCommandsList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	// Get global commands
	globalStore := command.NewStore(GetGlobalPath("commands"))
	globalCommands, err := globalStore.List()
	if err != nil {
		globalCommands = nil
	}

	// Get local commands (if .claude/commands exists)
	var localCommands []*command.Command
	if localPath := GetLocalPath("commands"); localPath != "" {
		localStore := command.NewStore(localPath)
		localCommands, _ = localStore.List()
	}

	if commandsListJSON {
		output := commandsListOutput{
			Global: globalCommands,
			Local:  localCommands,
		}
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonOutput))
		return nil
	}

	// Print global section
	fmt.Println("=== Global (~/.claude/commands/) ===")
	if len(globalCommands) == 0 {
		fmt.Println("No commands found.")
	} else {
		printCommandsTable(globalCommands)
	}

	// Print local section only if exists and has items
	if len(localCommands) > 0 {
		fmt.Println()
		fmt.Println("=== Local (.claude/commands/) ===")
		printCommandsTable(localCommands)
	}

	return nil
}

func printCommandsJSON(commands []*command.Command) error {
	output, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func printCommandsTable(commands []*command.Command) {
	// Calculate column widths
	nameWidth := len("NAME")

	for _, c := range commands {
		if len(c.Name) > nameWidth {
			nameWidth = len(c.Name)
		}
	}

	// Cap widths
	if nameWidth > 30 {
		nameWidth = 30
	}
	const descWidth = 50

	// Print header
	fmt.Printf("%-*s  %-*s\n",
		nameWidth, "NAME",
		descWidth, "DESCRIPTION")
	fmt.Printf("%s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", descWidth))

	// Print rows
	for _, c := range commands {
		name := c.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		desc := c.Description
		if len(desc) > descWidth {
			desc = desc[:descWidth-3] + "..."
		}

		fmt.Printf("%-*s  %-*s\n",
			nameWidth, name,
			descWidth, desc)
	}

	fmt.Printf("\nTotal: %d commands\n", len(commands))
}
