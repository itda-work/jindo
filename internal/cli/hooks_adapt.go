package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/itda-work/jindo/internal/hook"
	"github.com/itda-work/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	hooksAdaptGlobal bool
	hooksAdaptLocal  bool
)

var hooksAdaptCmd = &cobra.Command{
	Use:   "adapt <hook-name>",
	Short: "Customize a hook using AI conversation",
	Long: `Customize a hook to fit your specific workflow using AI-powered conversation.

This command:
1. Backs up the current configuration to .history/
2. Starts an AI conversation to understand your needs
3. Provides guidance on modifying the hook
4. Helps you update the hook configuration

Use --local to adapt from the current directory's .claude/settings.json.`,
	Example: `  # Adapt a global hook
  jd hooks adapt PreToolUse-Bash-0

  # Adapt a local hook
  jd hooks adapt PreToolUse-Bash-0 --local`,
	Args:              cobra.ExactArgs(1),
	RunE:              runHooksAdapt,
	ValidArgsFunction: hookNameCompletion,
}

func init() {
	hooksCmd.AddCommand(hooksAdaptCmd)
	hooksAdaptCmd.Flags().BoolVarP(&hooksAdaptGlobal, "global", "g", false, "Adapt from global ~/.claude/settings.json (default)")
	hooksAdaptCmd.Flags().BoolVarP(&hooksAdaptLocal, "local", "l", false, "Adapt from local .claude/settings.json")
}

func runHooksAdapt(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Validate mutually exclusive flags
	if err := ValidateScopeFlags(hooksAdaptGlobal, hooksAdaptLocal); err != nil {
		return err
	}

	hookName := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if hooksAdaptLocal {
		scope = ScopeLocal
	}

	settingsPath := GetSettingsPathByScope(scope)
	store := hook.NewStore(settingsPath)

	// Get hook to verify it exists
	h, err := store.Get(hookName)
	if err != nil {
		return fmt.Errorf("hook not found: %s", hookName)
	}

	// Get claude dir for history
	claudeDir := filepath.Dir(settingsPath)
	if strings.HasPrefix(claudeDir, "~/") {
		home, _ := os.UserHomeDir()
		claudeDir = filepath.Join(home, claudeDir[2:])
	}

	// Create history manager and backup current version
	historyMgr := hook.NewHistoryManager(claudeDir, hookName)

	version, err := historyMgr.SaveVersion(h)
	if err != nil {
		return fmt.Errorf("failed to backup current version: %w", err)
	}
	fmt.Printf("📦 Backed up to %s\n", hook.FormatVersionName(version))

	// Load and render the adapt prompt
	promptTemplate, err := prompt.Load("adapt-hook")
	if err != nil {
		return fmt.Errorf("failed to load adapt prompt: %w", err)
	}

	tmpl, err := template.New("adapt-hook").Parse(promptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var systemPrompt bytes.Buffer
	err = tmpl.Execute(&systemPrompt, map[string]interface{}{
		"HookName":  h.Name,
		"EventType": h.EventType,
		"Matcher":   h.Matcher,
		"Commands":  h.Commands,
	})
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	// Create a temporary file with hook info for Claude to read
	hookInfo := map[string]interface{}{
		"name":       h.Name,
		"event_type": h.EventType,
		"matcher":    h.Matcher,
		"commands":   h.Commands,
		"settings_path": settingsPath,
	}
	hookInfoJSON, _ := json.MarshalIndent(hookInfo, "", "  ")

	// Show current hook info
	fmt.Println()
	fmt.Println("Current hook configuration:")
	fmt.Println(string(hookInfoJSON))
	fmt.Println()

	// Show tip about customizing the prompt
	fmt.Printf("💡 Tip: Customize this prompt with: jd prompts edit adapt-hook\n")
	fmt.Println()
	fmt.Println("🤖 Starting AI conversation to customize your hook...")
	fmt.Println("   - Describe what changes you want")
	fmt.Println("   - AI will ask clarifying questions")
	fmt.Println("   - Type 'exit' or Ctrl+C to finish")
	fmt.Println()

	// Initial prompt to make Claude start the conversation (passed as positional argument for interactive mode)
	initialPrompt := fmt.Sprintf("I want to customize the '%s' hook. Please start by asking me about my specific needs and how I'd like to adapt this hook to my workflow.", hookName)

	// Run claude command with the system prompt and initial message
	// Note: positional argument (not -p) keeps interactive mode
	claudeCmd := exec.Command("claude",
		"--system-prompt", systemPrompt.String(),
		"--allowedTools", "Edit,Read,Write,Bash",
		initialPrompt,
	)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	if err := claudeCmd.Run(); err != nil {
		// Check if it's just a user exit
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 { // Ctrl+C
				fmt.Println("\n⚠️  Adaptation cancelled")
				return nil
			}
		}
		return fmt.Errorf("claude command failed: %w", err)
	}

	// Read the potentially updated hook
	newHook, err := store.Get(hookName)
	if err != nil {
		// Hook might have been renamed or deleted
		fmt.Println("\n📝 Hook may have been modified. Use 'jd hooks list' to check.")
		return nil
	}

	// Check if hook changed
	if newHook.Matcher == h.Matcher && equalStringSlices(newHook.Commands, h.Commands) {
		fmt.Println("\n📝 No changes made to the hook")
		return nil
	}

	// Save new version
	newVersion, err := historyMgr.SaveVersion(newHook)
	if err != nil {
		return fmt.Errorf("failed to save new version: %w", err)
	}

	fmt.Printf("\n✅ Hook adapted successfully!\n")
	fmt.Printf("   Previous: %s\n", hook.FormatVersionName(version))
	fmt.Printf("   Current:  %s\n", hook.FormatVersionName(newVersion))
	fmt.Printf("\n   To revert: jd hooks revert %s %d\n", hookName, version.Number)

	return nil
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
