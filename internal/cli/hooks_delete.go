package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/itda-work/jindo/internal/hook"
	"github.com/spf13/cobra"
)

var hooksDeleteForce bool

var hooksDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"d", "rm", "remove"},
	Short:   "Delete a hook",
	Long: `Delete a hook from ~/.claude/settings.json.

Examples:
  jd hooks delete PreToolUse-Bash-0
  jd hooks delete PreToolUse-Bash-0 -f`,
	Args: cobra.ExactArgs(1),
	RunE: runHooksDelete,
}

func init() {
	hooksCmd.AddCommand(hooksDeleteCmd)
	hooksDeleteCmd.Flags().BoolVarP(&hooksDeleteForce, "force", "f", false, "Skip confirmation")
}

func runHooksDelete(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]

	store := hook.NewStore("~/.claude/settings.json")
	h, err := store.Get(name)
	if err != nil {
		return fmt.Errorf("hook not found: %s", name)
	}

	// Confirm deletion
	if !hooksDeleteForce {
		fmt.Printf("Hook to delete:\n")
		fmt.Printf("  Name:    %s\n", h.Name)
		fmt.Printf("  Event:   %s\n", h.EventType)
		fmt.Printf("  Matcher: %s\n", h.Matcher)
		fmt.Printf("  Commands: %s\n", strings.Join(h.Commands, ", "))
		fmt.Print("\nAre you sure you want to delete this hook? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := store.Delete(name); err != nil {
		return fmt.Errorf("failed to delete hook: %w", err)
	}

	fmt.Printf("✓ Deleted hook: %s\n", name)
	return nil
}
