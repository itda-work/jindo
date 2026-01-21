package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/itda-work/jindo/internal/pkg/pkgmgr"
	"github.com/itda-work/jindo/internal/pkg/repo"
	"github.com/itda-work/jindo/internal/tui"
	"github.com/spf13/cobra"
)

var (
	pkgBrowseType string
	pkgBrowseJSON bool
)

var pkgBrowseCmd = &cobra.Command{
	Use:     "browse [namespace]",
	Aliases: []string{"b"},
	Short:   "Browse packages in repositories",
	Long: `Browse available packages (skills, commands, agents, hooks) in registered repositories.

Without arguments, opens an interactive TUI to browse all registered repositories.
With a namespace argument, opens TUI filtered to that specific repository.

Use --type to select the initial tab (TUI) or filter output (--json).
Use --json for machine-readable output.

Examples:
  jd pkg browse                     # Interactive TUI
  jd pkg browse affa-ever           # TUI filtered to affa-ever
  jd pkg browse --json              # JSON output of all packages
  jd pkg browse affa-ever --json    # JSON output of affa-ever packages`,
	Args:              cobra.MaximumNArgs(1),
	RunE:              runPkgBrowse,
	ValidArgsFunction: pkgBrowseCompletion,
}

func init() {
	pkgCmd.AddCommand(pkgBrowseCmd)
	pkgBrowseCmd.Flags().StringVarP(&pkgBrowseType, "type", "t", "", "Filter by type (skills, commands, agents, hooks)")
	pkgBrowseCmd.Flags().BoolVar(&pkgBrowseJSON, "json", false, "Output in JSON format")
}

func runPkgBrowse(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	namespace := ""
	if len(args) > 0 {
		namespace = args[0]
	}

	// If JSON output is requested, use CLI mode
	if pkgBrowseJSON {
		return runPkgBrowseCLI(namespace)
	}

	// Validate/parse type for TUI starting tab
	var startTab tui.Tab
	switch pkgBrowseType {
	case "", "skills", "skill":
		startTab = tui.TabSkills
	case "commands", "command":
		startTab = tui.TabCommands
	case "agents", "agent":
		startTab = tui.TabAgents
	case "hooks", "hook":
		startTab = tui.TabHooks
	default:
		return fmt.Errorf("invalid type: %s (use: skills, commands, agents, hooks)", pkgBrowseType)
	}

	// Launch TUI (with optional namespace filter)
	manager := pkgmgr.NewManager("~/.itda-jindo")

	// Validate namespace exists if provided
	if namespace != "" {
		store := repo.NewStore("~/.itda-jindo")
		if _, err := store.Get(namespace); err != nil {
			return fmt.Errorf("repository '%s' not found", namespace)
		}
	}

	return tui.Run(manager, namespace, startTab)
}

func runPkgBrowseCLI(namespace string) error {
	store := repo.NewStore("~/.itda-jindo")

	// Validate type filter
	var typeFilter repo.PackageType
	switch pkgBrowseType {
	case "":
		// No filter
	case "skills", "skill":
		typeFilter = repo.TypeSkill
	case "commands", "command":
		typeFilter = repo.TypeCommand
	case "agents", "agent":
		typeFilter = repo.TypeAgent
	case "hooks", "hook":
		typeFilter = repo.TypeHook
	default:
		return fmt.Errorf("invalid type: %s (use: skills, commands, agents, hooks)", pkgBrowseType)
	}

	// If no namespace, browse all repositories
	if namespace == "" {
		repos, err := store.List()
		if err != nil {
			return fmt.Errorf("list repositories: %w", err)
		}

		var allItems []repo.BrowseItem
		for _, r := range repos {
			items, err := store.Browse(r.Namespace, typeFilter)
			if err != nil {
				continue
			}
			allItems = append(allItems, items...)
		}

		output, err := json.MarshalIndent(allItems, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	// Get repository info
	config, err := store.Get(namespace)
	if err != nil {
		return fmt.Errorf("repository '%s' not found", namespace)
	}

	fmt.Fprintf(os.Stderr, "Browsing %s (%s)...\n\n", namespace, config.URL)

	items, err := store.Browse(namespace, typeFilter)
	if err != nil {
		return fmt.Errorf("browse repository: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("[]")
		return nil
	}

	output, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

// pkgBrowseCompletion provides tab completion for repository namespaces
func pkgBrowseCompletion(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// Only complete first argument
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	store := repo.NewStore("~/.itda-jindo")
	repos, err := store.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, r := range repos {
		// Format: "namespace\tdescription" for shell completion with description
		desc := r.Description
		if desc == "" {
			desc = r.URL // fallback to URL if no description
		}
		completions = append(completions, fmt.Sprintf("%s\t%s", r.Namespace, desc))
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
