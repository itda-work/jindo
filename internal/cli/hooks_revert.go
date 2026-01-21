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
	hooksRevertGlobal bool
	hooksRevertLocal  bool
)

var hooksRevertCmd = &cobra.Command{
	Use:   "revert <hook-name> [version]",
	Short: "Revert a hook to a previous version",
	Long: `Revert a hook to a previous version from its history.

If no version is specified, shows available versions.
Version can be a number (e.g., 1, 2) or 'latest'.`,
	Example: `  # Show available versions
  jd hooks revert PreToolUse-Bash-0

  # Revert to version 1
  jd hooks revert PreToolUse-Bash-0 1

  # Revert to the latest backed up version
  jd hooks revert PreToolUse-Bash-0 latest`,
	Args:              cobra.RangeArgs(1, 2),
	RunE:              runHooksRevert,
	ValidArgsFunction: hookNameCompletion,
}

func init() {
	hooksCmd.AddCommand(hooksRevertCmd)
	hooksRevertCmd.Flags().BoolVarP(&hooksRevertGlobal, "global", "g", false, "Revert from global ~/.claude/ (default)")
	hooksRevertCmd.Flags().BoolVarP(&hooksRevertLocal, "local", "l", false, "Revert from local .claude/")
}

func runHooksRevert(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Validate mutually exclusive flags
	if err := ValidateScopeFlags(hooksRevertGlobal, hooksRevertLocal); err != nil {
		return err
	}

	hookName := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if hooksRevertLocal {
		scope = ScopeLocal
	}

	settingsPath := GetSettingsPathByScope(scope)
	store := hook.NewStore(settingsPath)

	// Verify hook exists
	currentHook, err := store.Get(hookName)
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

	// If no version specified, show available versions
	if len(args) < 2 {
		versions, err := historyMgr.ListVersions()
		if err != nil {
			return fmt.Errorf("failed to list versions: %w", err)
		}

		if len(versions) == 0 {
			fmt.Printf("No history found for hook: %s\n", hookName)
			return nil
		}

		fmt.Printf("Available versions for hook: %s\n\n", hookName)
		for _, v := range versions {
			fmt.Printf("  %s\n", hook.FormatVersionName(&v))
		}
		fmt.Printf("\nUsage: jd hooks revert %s <version>\n", hookName)
		return nil
	}

	// Parse version argument
	versionArg := args[1]
	versionNum, err := hook.ParseVersionArg(versionArg)
	if err != nil {
		return err
	}

	var snapshot *hook.HookSnapshot
	var version *hook.Version

	if versionNum == -1 {
		// Get latest version
		version, err = historyMgr.GetLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
		snapshot, _, err = historyMgr.GetVersion(version.Number)
	} else {
		snapshot, version, err = historyMgr.GetVersion(versionNum)
	}

	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// Backup current hook before reverting
	_, err = historyMgr.SaveVersion(currentHook)
	if err != nil {
		return fmt.Errorf("failed to backup current version: %w", err)
	}

	// Update the hook with the reverted configuration
	_, err = store.Update(hookName, snapshot.Matcher, snapshot.Commands)
	if err != nil {
		return fmt.Errorf("failed to update hook: %w", err)
	}

	fmt.Printf("✅ Reverted hook '%s' to %s\n", hookName, hook.FormatVersionName(version))
	fmt.Println("\nCurrent configuration has been backed up to history.")

	return nil
}
