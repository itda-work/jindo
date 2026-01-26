package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itda-skills/jindo/internal/hook"
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
	hooksRevertCmd.Flags().BoolVarP(&hooksRevertGlobal, "global", "g", false, "Revert from global ~/.claude/")
	hooksRevertCmd.Flags().BoolVarP(&hooksRevertLocal, "local", "l", false, "Revert from local .claude/")
}

func runHooksRevert(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	hookName := args[0]

	scope, err := ResolveScope(hooksRevertGlobal, hooksRevertLocal)
	if err != nil {
		return err
	}

	settingsPath := GetSettingsPathByScope(scope)
	store := hook.NewStore(settingsPath)

	// Verify hook exists
	_, err = store.Get(hookName)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("hook not found in %s: %s", ScopeDescription(scope), hookName)
		}
		return fmt.Errorf("failed to get hook: %w", err)
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

		// Get current hook to find active version
		currentHook, _ := store.Get(hookName)

		fmt.Printf("Available versions for hook: %s\n\n", hookName)
		for _, v := range versions {
			marker := "  "
			// Check if this version matches current hook
			if snapshot, _, err := historyMgr.GetVersion(v.Number); err == nil && currentHook != nil {
				if snapshot.Matcher == currentHook.Matcher && equalStringSlices(snapshot.Commands, currentHook.Commands) {
					marker = "* "
				}
			}
			fmt.Printf("%s%s\n", marker, hook.FormatVersionName(&v))
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

	// Update the hook with the reverted configuration
	_, err = store.Update(hookName, snapshot.Matcher, snapshot.Commands)
	if err != nil {
		return fmt.Errorf("failed to update hook: %w", err)
	}

	// Delete all versions after the reverted version
	deleted, err := historyMgr.DeleteVersionsAfter(version.Number)
	if err != nil {
		return fmt.Errorf("failed to cleanup versions: %w", err)
	}

	fmt.Printf("✅ Reverted hook '%s' to %s\n", hookName, hook.FormatVersionName(version))
	if deleted > 0 {
		fmt.Printf("   Removed %d newer version(s)\n", deleted)
	}

	return nil
}
