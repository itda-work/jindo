package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itda-work/jindo/internal/hook"
	"github.com/spf13/cobra"
)

var (
	hooksHistoryGlobal bool
	hooksHistoryLocal  bool
)

var hooksHistoryCmd = &cobra.Command{
	Use:     "history <hook-name>",
	Aliases: []string{"hist"},
	Short:   "Show version history of a hook",
	Long: `Show the version history of a hook.

Each time a hook is adapted, a new version is saved.
Use 'jd hooks revert' to restore a previous version.`,
	Example: `  # Show history of a global hook
  jd hooks history PreToolUse-Bash-0

  # Show history of a local hook
  jd hooks history PreToolUse-Bash-0 --local`,
	Args:              cobra.ExactArgs(1),
	RunE:              runHooksHistory,
	ValidArgsFunction: hookNameCompletion,
}

func init() {
	hooksCmd.AddCommand(hooksHistoryCmd)
	hooksHistoryCmd.Flags().BoolVarP(&hooksHistoryGlobal, "global", "g", false, "Show from global ~/.claude/ (default)")
	hooksHistoryCmd.Flags().BoolVarP(&hooksHistoryLocal, "local", "l", false, "Show from local .claude/")
}

func runHooksHistory(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Validate mutually exclusive flags
	if err := ValidateScopeFlags(hooksHistoryGlobal, hooksHistoryLocal); err != nil {
		return err
	}

	hookName := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if hooksHistoryLocal {
		scope = ScopeLocal
	}

	settingsPath := GetSettingsPathByScope(scope)
	store := hook.NewStore(settingsPath)

	// Verify hook exists
	_, err := store.Get(hookName)
	if err != nil {
		return fmt.Errorf("hook not found: %s", hookName)
	}

	// Get claude dir for history
	claudeDir := filepath.Dir(settingsPath)
	if strings.HasPrefix(claudeDir, "~/") {
		home, _ := os.UserHomeDir()
		claudeDir = filepath.Join(home, claudeDir[2:])
	}

	// Create history manager
	historyMgr := hook.NewHistoryManager(claudeDir, hookName)

	versions, err := historyMgr.ListVersions()
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Printf("No history found for hook: %s\n", hookName)
		fmt.Println("\nHistory is created when you use 'jd hooks adapt'.")
		return nil
	}

	fmt.Printf("Version history for hook: %s\n\n", hookName)

	for i, v := range versions {
		marker := "  "
		if i == 0 {
			marker = "* " // Mark the latest
		}
		fmt.Printf("%s%s\n", marker, hook.FormatVersionName(&v))
	}

	fmt.Printf("\nTotal: %d version(s)\n", len(versions))
	fmt.Printf("\nTo revert: jd hooks revert %s <version>\n", hookName)

	return nil
}
