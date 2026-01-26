package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/itda-skills/jindo/internal/prompt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var (
	claudemdTidyGlobal bool
	claudemdTidyLocal  bool
	claudemdTidyDryRun bool
	claudemdTidyStyle  string
)

var claudemdTidyCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Analyze and optimize CLAUDE.md file",
	Long: `Analyze and optimize CLAUDE.md file using AI.

Uses Claude CLI to analyze the CLAUDE.md file and:
- Remove duplicate instructions
- Improve structure and organization
- Ensure consistency
- Apply style preferences

The original file is backed up before any changes.
Default scope is local (.claude/CLAUDE.md) if present, otherwise global (~/.claude/CLAUDE.md).

Requires Claude CLI: npm install -g @anthropic-ai/claude-cli`,
	Example: `  # Tidy local CLAUDE.md (if exists) or global
  jd claudemd tidy

  # Tidy with minimal style
  jd claudemd tidy --style minimal

  # Preview changes without applying
  jd claudemd tidy --dry-run

  # Tidy global CLAUDE.md explicitly
  jd claudemd tidy --global`,
	RunE: runClaudemdTidy,
}

func init() {
	claudemdCmd.AddCommand(claudemdTidyCmd)

	claudemdTidyCmd.Flags().BoolVarP(&claudemdTidyGlobal, "global", "g", false, "Tidy global ~/.claude/CLAUDE.md")
	claudemdTidyCmd.Flags().BoolVarP(&claudemdTidyLocal, "local", "l", false, "Tidy local .claude/CLAUDE.md")
	claudemdTidyCmd.Flags().BoolVar(&claudemdTidyDryRun, "dry-run", false, "Preview changes without applying")
	claudemdTidyCmd.Flags().StringVar(&claudemdTidyStyle, "style", "structured", "Style: minimal, detailed, structured")
}

func runClaudemdTidy(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	// Validate style
	if err := validateStyle(claudemdTidyStyle); err != nil {
		return err
	}

	// Validate and resolve scope
	scope, err := ResolveScope(claudemdTidyGlobal, claudemdTidyLocal)
	if err != nil {
		return err
	}

	// Check Claude CLI installed
	if err := checkClaudeInstalled(); err != nil {
		return err
	}

	// Get CLAUDE.md path
	claudemdPath := getCLAUDEmdPath(scope)

	// Check if CLAUDE.md exists
	if _, err := os.Stat(claudemdPath); os.IsNotExist(err) {
		return fmt.Errorf("CLAUDE.md not found at %s", claudemdPath)
	}

	// Read current content
	originalContent, err := os.ReadFile(claudemdPath)
	if err != nil {
		return fmt.Errorf("failed to read CLAUDE.md: %w", err)
	}

	// Create backup (unless dry-run)
	var backupPath string
	if !claudemdTidyDryRun {
		backupPath, err = backupCLAUDEmd(claudemdPath)
		if err != nil {
			return fmt.Errorf("failed to backup CLAUDE.md: %w", err)
		}
	}

	// Run Claude to tidy the content
	fmt.Printf("🔍 Analyzing CLAUDE.md with Claude CLI (style: %s)...\n", claudemdTidyStyle)
	tidiedContent, err := runClaudeTidy(string(originalContent), claudemdTidyStyle)
	if err != nil {
		if backupPath != "" {
			return fmt.Errorf("%w\n\nBackup preserved at: %s", err, backupPath)
		}
		return err
	}

	// Validate output
	if len(strings.TrimSpace(tidiedContent)) == 0 {
		if backupPath != "" {
			return fmt.Errorf("empty output from claude\n\nBackup preserved at: %s", backupPath)
		}
		return fmt.Errorf("empty output from claude")
	}

	// If dry-run, show diff and exit
	if claudemdTidyDryRun {
		showDiff(string(originalContent), tidiedContent)
		fmt.Println("\n💡 To apply changes, run without --dry-run")
		return nil
	}

	// Write tidied content to file
	if err := os.WriteFile(claudemdPath, []byte(tidiedContent), 0644); err != nil {
		return fmt.Errorf("failed to write tidied CLAUDE.md: %w\n\nBackup preserved at: %s", err, backupPath)
	}

	// Show success message
	fmt.Println("\n✅ CLAUDE.md tidied successfully!")
	fmt.Printf("\n📍 Location: %s\n", claudemdPath)
	fmt.Printf("💾 Backup: %s\n", backupPath)
	fmt.Printf("🎨 Style: %s\n", claudemdTidyStyle)

	return nil
}

// checkClaudeInstalled checks if Claude CLI is installed
func checkClaudeInstalled() error {
	_, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found. Install: npm install -g @anthropic-ai/claude-cli")
	}
	return nil
}

// getCLAUDEmdPath returns the path to CLAUDE.md based on scope
func getCLAUDEmdPath(scope PathScope) string {
	switch scope {
	case ScopeGlobal:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".claude", "CLAUDE.md")
	case ScopeLocal:
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, ".claude", "CLAUDE.md")
	default:
		// This shouldn't happen as ResolveScope handles it, but handle anyway
		if LocalClaudeDirExists() {
			cwd, _ := os.Getwd()
			return filepath.Join(cwd, ".claude", "CLAUDE.md")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".claude", "CLAUDE.md")
	}
}

// backupCLAUDEmd creates a timestamped backup of CLAUDE.md
func backupCLAUDEmd(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Create .claude/backups/ directory
	backupDir := filepath.Join(filepath.Dir(path), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	// Timestamped backup: CLAUDE.md.20260122-153045.bak
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("CLAUDE.md.%s.bak", timestamp))

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}

// runClaudeTidy executes Claude CLI to tidy the CLAUDE.md content
func runClaudeTidy(content, style string) (string, error) {
	// Load prompt template
	promptTemplate, err := prompt.Load("tidy-claudemd")
	if err != nil {
		return "", fmt.Errorf("failed to load tidy prompt: %w", err)
	}

	// Render with template
	tmpl, err := template.New("tidy").Parse(promptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"Content": content,
		"Style":   style,
	})
	if err != nil {
		return "", err
	}

	// Execute claude CLI
	cmd := exec.Command("claude",
		"-p", buf.String(),
		"--output-format", "text",
	)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// showDiff displays the diff between original and tidied content
func showDiff(oldContent, newContent string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldContent, newContent, false)
	fmt.Println("\n📋 Preview of changes:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(dmp.DiffPrettyText(diffs))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// validateStyle validates the style option
func validateStyle(style string) error {
	validStyles := []string{"minimal", "detailed", "structured"}
	for _, valid := range validStyles {
		if style == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid style: %s (valid options: %s)",
		style, strings.Join(validStyles, ", "))
}
