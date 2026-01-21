package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itda-work/jindo/internal/hook"
	"github.com/spf13/cobra"
)

var (
	hooksNewEventType    string
	hooksNewMatcher      string
	hooksNewCommand      string
	hooksNewCreateScript bool
	hooksNewGlobal       bool
	hooksNewLocal        bool
)

var hooksNewCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n", "add", "create"},
	Short:   "Create a new hook",
	Long: `Create a new hook in ~/.claude/settings.json (global) or .claude/settings.json (local).

This command runs in wizard mode if no flags are provided.
You can also specify all options via flags for non-interactive use.
Default scope is local if a .claude directory exists in the current working directory, otherwise global.
Use --global or --local to override.

Event types (with aliases):
  - PreToolUse (pre): Runs before a tool is executed
  - PostToolUse (post): Runs after a tool is executed
  - Notification (notify): Runs on notifications
  - Stop: Runs when Claude stops
  - SubagentStop (sub): Runs when a subagent stops

Matcher patterns:
  - Single tool: "Bash", "Write", "Edit"
  - Multiple tools: "Bash|Write|Edit" (regex OR)
  - All tools: "*"

Examples:
  jd hooks new
  jd hooks new -e pre -m "Bash" -c "echo 'Running bash'"
  jd hooks new -e post -m "Bash|Write" -c "~/.claude/hooks/log.sh"
  jd hooks new -e post -m "Bash" --script
  jd hooks new --local -e pre -m "Bash" -c "echo 'local hook'"`,
	RunE:              runHooksNew,
	ValidArgsFunction: hooksNewCompletion,
}

func init() {
	hooksCmd.AddCommand(hooksNewCmd)
	hooksNewCmd.Flags().StringVarP(&hooksNewEventType, "event", "e", "", "Event type: pre, post, notify, stop, sub")
	hooksNewCmd.Flags().StringVarP(&hooksNewMatcher, "matcher", "m", "", "Tool matcher pattern (e.g., Bash, \"Bash|Write\", *)")
	hooksNewCmd.Flags().StringVarP(&hooksNewCommand, "command", "c", "", "Command to execute")
	hooksNewCmd.Flags().BoolVar(&hooksNewCreateScript, "script", false, "Create a script file in ~/.claude/hooks/")
	hooksNewCmd.Flags().BoolVarP(&hooksNewGlobal, "global", "g", false, "Create in global ~/.claude/settings.json")
	hooksNewCmd.Flags().BoolVarP(&hooksNewLocal, "local", "l", false, "Create in local .claude/settings.json")

	// Register completion for --event flag
	_ = hooksNewCmd.RegisterFlagCompletionFunc("event", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"PreToolUse\tRuns before a tool is executed",
			"PostToolUse\tRuns after a tool is executed",
			"Notification\tRuns on notifications",
			"Stop\tRuns when Claude stops",
			"SubagentStop\tRuns when a subagent stops",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	// Register completion for --matcher flag
	_ = hooksNewCmd.RegisterFlagCompletionFunc("matcher", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"Bash\tExecute shell commands",
			"Read\tRead files",
			"Write\tWrite files",
			"Edit\tEdit files",
			"Glob\tFind files by pattern",
			"Grep\tSearch file contents",
			"Task\tRun subagent tasks",
			"*\tAll tools",
		}, cobra.ShellCompDirectiveNoFileComp
	})
}

func hooksNewCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func runHooksNew(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	scope, err := ResolveScope(hooksNewGlobal, hooksNewLocal)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	// Get event type
	eventTypeStr := hooksNewEventType
	if eventTypeStr == "" {
		fmt.Println("Select event type:")
		eventTypes := hook.AllEventTypes()
		aliases := []string{"pre", "post", "notify", "", "sub"}
		for i, et := range eventTypes {
			if aliases[i] != "" {
				fmt.Printf("  %d. %s (%s)\n", i+1, et, aliases[i])
			} else {
				fmt.Printf("  %d. %s\n", i+1, et)
			}
		}
		fmt.Print("Enter number (1-5) or alias: ")
		input, _ := reader.ReadString('\n')
		eventTypeStr = strings.TrimSpace(input)

		// Check if it's a number
		var idx int
		if _, err := fmt.Sscanf(eventTypeStr, "%d", &idx); err == nil && idx >= 1 && idx <= 5 {
			eventTypeStr = string(eventTypes[idx-1])
		}
	}

	// Parse and validate event type using ParseEventType
	validEventType, err := hook.ParseEventType(eventTypeStr)
	if err != nil {
		return err
	}

	// Get matcher
	matcher := hooksNewMatcher
	if matcher == "" {
		fmt.Println("\nEnter matcher pattern:")
		fmt.Println("  Examples: Bash, \"Bash|Write\", * (all tools)")
		fmt.Print("Matcher: ")
		matcher, _ = reader.ReadString('\n')
		matcher = strings.TrimSpace(matcher)
	}
	if matcher == "" {
		return fmt.Errorf("matcher is required (use * for all tools)")
	}

	// Get command
	command := hooksNewCommand
	if command == "" {
		fmt.Println("\nEnter command to execute:")
		fmt.Println("  Examples: echo 'hello', ~/.claude/hooks/myscript.sh")
		fmt.Print("Command: ")
		command, _ = reader.ReadString('\n')
		command = strings.TrimSpace(command)
	}
	if command == "" {
		return fmt.Errorf("command is required")
	}

	// Optionally create script file
	if hooksNewCreateScript || (!hooksNewCreateScript && hooksNewCommand == "") {
		if hooksNewCommand == "" {
			fmt.Print("\nCreate a script file? (y/N): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input == "y" || input == "yes" {
				hooksNewCreateScript = true
			}
		}
	}

	if hooksNewCreateScript {
		scriptName := fmt.Sprintf("%s-%s.sh", strings.ToLower(string(validEventType)), sanitizeMatcherForFilename(matcher))
		fmt.Printf("\nScript filename [%s]: ", scriptName)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			scriptName = input
		}

		// Create script with template
		template := fmt.Sprintf(`#!/usr/bin/env sh
# Hook: %s
# Matcher: %s
# Created by jd hooks new

# Available environment variables:
# $TOOL_NAME - Name of the tool being called
# $TOOL_INPUT - JSON input to the tool
# $TOOL_OUTPUT - JSON output from the tool (PostToolUse only)

echo "Hook triggered: %s for $TOOL_NAME"
`, validEventType, matcher, validEventType)

		scriptPath, err := hook.CreateScript(scriptName, template)
		if err != nil {
			return fmt.Errorf("failed to create script: %w", err)
		}

		fmt.Printf("Created script: %s\n", scriptPath)
		command = scriptPath
	}

	// Add hook to settings.json
	store := hook.NewStore(GetSettingsPathByScope(scope))
	newHook, err := store.Add(validEventType, matcher, []string{command})
	if err != nil {
		return fmt.Errorf("failed to add hook: %w", err)
	}

	fmt.Printf("\n✓ Created hook: %s\n", newHook.Name)
	fmt.Printf("  Event: %s\n", newHook.EventType)
	fmt.Printf("  Matcher: %s\n", newHook.Matcher)
	fmt.Printf("  Command: %s\n", strings.Join(newHook.Commands, ", "))

	return nil
}

func sanitizeMatcherForFilename(matcher string) string {
	result := matcher
	if result == "*" {
		result = "all"
	}
	result = strings.ReplaceAll(result, "|", "-")
	result = strings.ReplaceAll(result, " ", "_")
	// Keep only alphanumeric, dash, underscore
	var clean strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			clean.WriteRune(r)
		}
	}
	cleaned := clean.String()
	if cleaned == "" {
		return "hook"
	}
	return filepath.Clean(cleaned)
}
