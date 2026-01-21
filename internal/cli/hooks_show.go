package cli

import (
	"encoding/json"
	"fmt"

	"github.com/itda-work/jindo/internal/hook"
	"github.com/spf13/cobra"
)

var (
	hooksShowJSON   bool
	hooksShowGlobal bool
	hooksShowLocal  bool
)

var hooksShowCmd = &cobra.Command{
	Use:     "show <name>",
	Aliases: []string{"s", "get", "view"},
	Short:   "Show hook details",
	Long: `Show details of a specific hook from ~/.claude/settings.json (global) or .claude/settings.json (local).

Use --local to show from the current directory's .claude/settings.json.`,
	Args:              cobra.ExactArgs(1),
	RunE:              runHooksShow,
	ValidArgsFunction: hookNameCompletion,
}

func init() {
	hooksCmd.AddCommand(hooksShowCmd)
	hooksShowCmd.Flags().BoolVar(&hooksShowJSON, "json", false, "Output in JSON format")
	hooksShowCmd.Flags().BoolVarP(&hooksShowGlobal, "global", "g", false, "Show from global ~/.claude/settings.json (default)")
	hooksShowCmd.Flags().BoolVarP(&hooksShowLocal, "local", "l", false, "Show from local .claude/settings.json")
}

func runHooksShow(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if hooksShowLocal {
		scope = ScopeLocal
	}

	store := hook.NewStore(GetSettingsPathByScope(scope))
	h, err := store.Get(name)
	if err != nil {
		return fmt.Errorf("hook not found: %s", name)
	}

	if hooksShowJSON {
		output, err := json.MarshalIndent(h, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	// Pretty print
	fmt.Printf("Name:      %s\n", h.Name)
	fmt.Printf("Event:     %s\n", h.EventType)
	fmt.Printf("Matcher:   %s\n", h.Matcher)
	fmt.Printf("Commands:\n")
	for i, cmd := range h.Commands {
		fmt.Printf("  %d. %s\n", i+1, cmd)
	}

	// Show event type description
	fmt.Printf("\nEvent Description:\n")
	switch h.EventType {
	case hook.PreToolUse:
		fmt.Println("  Runs before a tool is executed.")
		fmt.Println("  Available vars: $TOOL_NAME, $TOOL_INPUT")
	case hook.PostToolUse:
		fmt.Println("  Runs after a tool is executed.")
		fmt.Println("  Available vars: $TOOL_NAME, $TOOL_INPUT, $TOOL_OUTPUT")
	case hook.Notification:
		fmt.Println("  Runs on notifications.")
	case hook.Stop:
		fmt.Println("  Runs when Claude stops.")
	case hook.SubagentStop:
		fmt.Println("  Runs when a subagent stops.")
	}

	return nil
}

// hookNameCompletion provides completion for hook names
func hookNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Check if --local flag is set
	scope := ScopeGlobal
	if local, _ := cmd.Flags().GetBool("local"); local {
		scope = ScopeLocal
	}

	store := hook.NewStore(GetSettingsPathByScope(scope))
	hooks, err := store.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, h := range hooks {
		desc := fmt.Sprintf("%s: %s", h.EventType, h.Matcher)
		names = append(names, fmt.Sprintf("%s\t%s", h.Name, desc))
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
