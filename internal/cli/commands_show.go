package cli

import (
	"fmt"
	"os"

	"github.com/itda-work/jindo/internal/command"
	"github.com/spf13/cobra"
)

var (
	commandsShowBrief  bool
	commandsShowGlobal bool
	commandsShowLocal  bool
)

var commandsShowCmd = &cobra.Command{
	Use:     "show <command-name>",
	Aliases: []string{"s", "get", "view"},
	Short:   "Show command details",
	Long: `Show the full content of a specific command from ~/.claude/commands/ (global) or .claude/commands/ (local) directory.

Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.`,
	Args: cobra.ExactArgs(1),
	RunE: runCommandsShow,
}

func init() {
	commandsCmd.AddCommand(commandsShowCmd)
	commandsShowCmd.Flags().BoolVar(&commandsShowBrief, "brief", false, "Show only metadata (name, description)")
	commandsShowCmd.Flags().BoolVarP(&commandsShowGlobal, "global", "g", false, "Show from global ~/.claude/commands/")
	commandsShowCmd.Flags().BoolVarP(&commandsShowLocal, "local", "l", false, "Show from local .claude/commands/")
}

func runCommandsShow(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]

	scope, err := ResolveScope(commandsShowGlobal, commandsShowLocal)
	if err != nil {
		return err
	}

	store := command.NewStore(GetPathByScope(scope, "commands"))

	if commandsShowBrief {
		return showCommandBrief(store, name, scope)
	}

	return showCommandFull(store, name, scope)
}

func showCommandBrief(store *command.Store, name string, scope PathScope) error {
	cmd, err := store.Get(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("command not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get command: %w", err)
	}

	fmt.Printf("Name:        %s\n", cmd.Name)
	fmt.Printf("Description: %s\n", cmd.Description)
	fmt.Printf("Path:        %s\n", cmd.Path)

	return nil
}

func showCommandFull(store *command.Store, name string, scope PathScope) error {
	content, err := store.GetContent(name)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("command not found in %s: %s", ScopeDescription(scope), name)
		}
		return fmt.Errorf("failed to get command content: %w", err)
	}

	fmt.Print(content)
	return nil
}
