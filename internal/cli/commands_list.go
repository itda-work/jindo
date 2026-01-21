package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-work/jindo/internal/command"
	"github.com/spf13/cobra"
)

var commandsListJSON bool

var commandsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List all commands",
	Long:    `List all commands from ~/.claude/commands/ directory.`,
	RunE:    runCommandsList,
}

func init() {
	commandsCmd.AddCommand(commandsListCmd)
	commandsListCmd.Flags().BoolVar(&commandsListJSON, "json", false, "Output in JSON format")
}

func runCommandsList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	store := command.NewStore("~/.claude/commands")
	commands, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to list commands: %w", err)
	}

	if len(commands) == 0 {
		fmt.Println("No commands found.")
		return nil
	}

	if commandsListJSON {
		return printCommandsJSON(commands)
	}

	printCommandsTable(commands)
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
