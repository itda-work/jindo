package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-skills/jindo/internal/hook"
	"github.com/spf13/cobra"
)

var hooksListJSON bool

var hooksListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List all hooks",
	Long:    `List all hooks from ~/.claude/settings.json and .claude/settings.json.`,
	RunE:    runHooksList,
}

func init() {
	hooksCmd.AddCommand(hooksListCmd)
	hooksListCmd.Flags().BoolVar(&hooksListJSON, "json", false, "Output in JSON format")
}

// hooksListOutput represents JSON output for hooks list with scope
type hooksListOutput struct {
	Global []*hook.Hook `json:"global"`
	Local  []*hook.Hook `json:"local,omitempty"`
}

func runHooksList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	// Get global hooks
	globalStore := hook.NewStore(GetSettingsPathByScope(ScopeGlobal))
	globalHooks, err := globalStore.List()
	if err != nil {
		globalHooks = nil
	}

	// Get local hooks (if .claude/settings.json exists)
	var localHooks []*hook.Hook
	if localPath := GetLocalSettingsPath(); localPath != "" {
		localStore := hook.NewStore(localPath)
		localHooks, _ = localStore.List()
	}

	if hooksListJSON {
		output := hooksListOutput{
			Global: globalHooks,
			Local:  localHooks,
		}
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonOutput))
		return nil
	}

	// Print global section
	fmt.Println("=== Global (~/.claude/settings.json) ===")
	if len(globalHooks) == 0 {
		fmt.Println("No hooks found.")
	} else {
		printHooksTable(globalHooks)
	}

	// Print local section only if exists and has items
	if len(localHooks) > 0 {
		fmt.Println()
		fmt.Println("=== Local (.claude/settings.json) ===")
		printHooksTable(localHooks)
	}

	return nil
}

func printHooksJSON(hooks []*hook.Hook) error {
	output, err := json.MarshalIndent(hooks, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func printHooksTable(hooks []*hook.Hook) {
	// Calculate column widths
	nameWidth := len("NAME")
	eventWidth := len("EVENT")
	matcherWidth := len("MATCHER")

	for _, h := range hooks {
		if len(h.Name) > nameWidth {
			nameWidth = len(h.Name)
		}
		if len(h.EventType) > eventWidth {
			eventWidth = len(string(h.EventType))
		}
		if len(h.Matcher) > matcherWidth {
			matcherWidth = len(h.Matcher)
		}
	}

	// Cap widths
	if nameWidth > 35 {
		nameWidth = 35
	}
	if eventWidth > 15 {
		eventWidth = 15
	}
	if matcherWidth > 20 {
		matcherWidth = 20
	}
	const cmdWidth = 40

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
		nameWidth, "NAME",
		eventWidth, "EVENT",
		matcherWidth, "MATCHER",
		cmdWidth, "COMMANDS")
	fmt.Printf("%s  %s  %s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", eventWidth),
		strings.Repeat("-", matcherWidth),
		strings.Repeat("-", cmdWidth))

	// Print rows
	for _, h := range hooks {
		name := h.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		event := string(h.EventType)
		if len(event) > eventWidth {
			event = event[:eventWidth-3] + "..."
		}

		matcher := h.Matcher
		if len(matcher) > matcherWidth {
			matcher = matcher[:matcherWidth-3] + "..."
		}

		cmds := strings.Join(h.Commands, "; ")
		if len(cmds) > cmdWidth {
			cmds = cmds[:cmdWidth-3] + "..."
		}

		fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
			nameWidth, name,
			eventWidth, event,
			matcherWidth, matcher,
			cmdWidth, cmds)
	}

	fmt.Printf("\nTotal: %d hooks\n", len(hooks))
}
