package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/itda-work/jindo/internal/agent"
	"github.com/itda-work/jindo/internal/prompt"
	"github.com/spf13/cobra"
)

var (
	agentsAdaptGlobal bool
	agentsAdaptLocal  bool
)

var agentsAdaptCmd = &cobra.Command{
	Use:   "adapt <agent-id>",
	Short: "Customize an agent using AI conversation",
	Long: `Customize an agent to fit your specific workflow using AI-powered conversation.

This command:
1. Backs up the current version to .history/
2. Starts an AI conversation to understand your needs
3. Modifies the agent based on the conversation
4. Saves changes and updates version history

Use --local to adapt from the current directory's .claude/agents/.`,
	Example: `  # Adapt a global agent
  jd agents adapt my-agent

  # Adapt a local agent
  jd agents adapt my-agent --local`,
	Args:              cobra.ExactArgs(1),
	RunE:              runAgentsAdapt,
	ValidArgsFunction: agentNameCompletion,
}

func init() {
	agentsCmd.AddCommand(agentsAdaptCmd)
	agentsAdaptCmd.Flags().BoolVarP(&agentsAdaptGlobal, "global", "g", false, "Adapt from global ~/.claude/agents/ (default)")
	agentsAdaptCmd.Flags().BoolVarP(&agentsAdaptLocal, "local", "l", false, "Adapt from local .claude/agents/")
}

func runAgentsAdapt(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	// Validate mutually exclusive flags
	if err := ValidateScopeFlags(agentsAdaptGlobal, agentsAdaptLocal); err != nil {
		return err
	}

	agentID := args[0]

	// Determine scope (default: global)
	scope := ScopeGlobal
	if agentsAdaptLocal {
		scope = ScopeLocal
	}

	agentsDir := GetPathByScope(scope, "agents")
	store := agent.NewStore(agentsDir)

	// Get agent to verify it exists
	a, err := store.Get(agentID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent not found: %s", agentID)
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}

	// Get current content
	content, err := store.GetContent(agentID)
	if err != nil {
		return fmt.Errorf("failed to read agent content: %w", err)
	}

	// Expand agentsDir for history manager
	expandedAgentsDir := agentsDir
	if strings.HasPrefix(expandedAgentsDir, "~/") {
		home, _ := os.UserHomeDir()
		expandedAgentsDir = filepath.Join(home, expandedAgentsDir[2:])
	}

	// Create history manager and backup current version
	historyMgr := agent.NewHistoryManager(expandedAgentsDir, agentID)

	version, err := historyMgr.SaveVersion(content)
	if err != nil {
		return fmt.Errorf("failed to backup current version: %w", err)
	}
	fmt.Printf("📦 Backed up to %s\n", agent.FormatVersionName(version))

	// Load and render the adapt prompt
	promptTemplate, err := prompt.Load("adapt-agent")
	if err != nil {
		return fmt.Errorf("failed to load adapt prompt: %w", err)
	}

	tmpl, err := template.New("adapt-agent").Parse(promptTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var systemPrompt bytes.Buffer
	err = tmpl.Execute(&systemPrompt, map[string]string{
		"AgentID":   agentID,
		"AgentPath": a.Path,
		"Content":   content,
	})
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	// Show tip about customizing the prompt
	fmt.Println()
	fmt.Printf("💡 Tip: Customize this prompt with: jd prompts edit adapt-agent\n")
	fmt.Println()
	fmt.Println("🤖 Starting AI conversation to customize your agent...")
	fmt.Println("   - Describe what changes you want")
	fmt.Println("   - AI will ask clarifying questions")
	fmt.Println("   - Type 'exit' or Ctrl+C to finish")
	fmt.Println()

	// Initial prompt to make Claude start the conversation (passed as positional argument for interactive mode)
	initialPrompt := fmt.Sprintf("I want to customize the '%s' agent. Please start by asking me about my specific needs and how I'd like to adapt this agent to my workflow.", agentID)

	// Run claude command with the system prompt and initial message
	// Note: positional argument (not -p) keeps interactive mode
	claudeCmd := exec.Command("claude",
		"--system-prompt", systemPrompt.String(),
		"--allowedTools", "Edit,Read,Write,Glob,Grep",
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

	// Read the potentially updated content
	newContent, err := store.GetContent(agentID)
	if err != nil {
		return fmt.Errorf("failed to read updated agent: %w", err)
	}

	// Check if content changed
	if newContent == content {
		fmt.Println("\n📝 No changes made to the agent")
		return nil
	}

	// Save new version
	newVersion, err := historyMgr.SaveVersion(newContent)
	if err != nil {
		return fmt.Errorf("failed to save new version: %w", err)
	}

	fmt.Printf("\n✅ Agent adapted successfully!\n")
	fmt.Printf("   Previous: %s\n", agent.FormatVersionName(version))
	fmt.Printf("   Current:  %s\n", agent.FormatVersionName(newVersion))
	fmt.Printf("\n   To revert: jd agents revert %s %d\n", agentID, version.Number)

	return nil
}
