package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-work/itda-jindo/internal/pkg/pkgmgr"
	"github.com/itda-work/itda-jindo/internal/pkg/repo"
	"github.com/itda-work/itda-jindo/internal/tui"
	"github.com/spf13/cobra"
)

var (
	pkgBrowseType string
	pkgBrowseJSON bool
)

var pkgBrowseCmd = &cobra.Command{
	Use:   "browse [namespace]",
	Short: "Browse packages in repositories",
	Long: `Browse available packages (skills, commands, agents, hooks) in registered repositories.

Without arguments, opens an interactive TUI to browse all registered repositories.
With a namespace argument, lists packages in that specific repository.

Use --type to filter by package type.

Examples:
  jd pkg browse                     # Interactive TUI
  jd pkg browse affa-ever           # List packages in affa-ever
  jd pkg browse affa-ever --type skills`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPkgBrowse,
}

func init() {
	pkgCmd.AddCommand(pkgBrowseCmd)
	pkgBrowseCmd.Flags().StringVarP(&pkgBrowseType, "type", "t", "", "Filter by type (skills, commands, agents, hooks)")
	pkgBrowseCmd.Flags().BoolVar(&pkgBrowseJSON, "json", false, "Output in JSON format")
}

func runPkgBrowse(_ *cobra.Command, args []string) error {
	// If no namespace provided, launch TUI
	if len(args) == 0 {
		manager := pkgmgr.NewManager("~/.itda-jindo")
		return tui.Run(manager)
	}

	namespace := args[0]
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

	// Get repository info
	config, err := store.Get(namespace)
	if err != nil {
		return fmt.Errorf("repository '%s' not found", namespace)
	}

	fmt.Printf("Browsing %s (%s)...\n\n", namespace, config.URL)

	items, err := store.Browse(namespace, typeFilter)
	if err != nil {
		return fmt.Errorf("browse repository: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("No packages found.")
		return nil
	}

	if pkgBrowseJSON {
		output, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	// Group by type
	skills := make([]repo.BrowseItem, 0)
	commands := make([]repo.BrowseItem, 0)
	agents := make([]repo.BrowseItem, 0)
	hooks := make([]repo.BrowseItem, 0)

	for _, item := range items {
		switch item.Type {
		case repo.TypeSkill:
			skills = append(skills, item)
		case repo.TypeCommand:
			commands = append(commands, item)
		case repo.TypeAgent:
			agents = append(agents, item)
		case repo.TypeHook:
			hooks = append(hooks, item)
		}
	}

	if len(skills) > 0 {
		fmt.Println("Skills:")
		printBrowseItems(skills, namespace)
		fmt.Println()
	}

	if len(commands) > 0 {
		fmt.Println("Commands:")
		printBrowseItems(commands, namespace)
		fmt.Println()
	}

	if len(agents) > 0 {
		fmt.Println("Agents:")
		printBrowseItems(agents, namespace)
		fmt.Println()
	}

	if len(hooks) > 0 {
		fmt.Println("Hooks:")
		printBrowseItems(hooks, namespace)
		fmt.Println()
	}

	fmt.Printf("Total: %d packages\n", len(items))
	fmt.Println()
	fmt.Println("Install with: jd pkg install <namespace>:<path>")
	fmt.Printf("Example: jd pkg install %s:%s\n", namespace, items[0].Path)

	return nil
}

func printBrowseItems(items []repo.BrowseItem, namespace string) {
	// Calculate column widths
	nameWidth := len("NAME")
	pathWidth := len("PATH")

	for _, item := range items {
		if len(item.Name) > nameWidth {
			nameWidth = len(item.Name)
		}
		if len(item.Path) > pathWidth {
			pathWidth = len(item.Path)
		}
	}

	// Cap widths
	if nameWidth > 30 {
		nameWidth = 30
	}
	if pathWidth > 50 {
		pathWidth = 50
	}

	// Print header
	fmt.Printf("  %-*s  %-*s\n",
		nameWidth, "NAME",
		pathWidth, "PATH")
	fmt.Printf("  %s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", pathWidth))

	// Print rows
	for _, item := range items {
		name := item.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		path := item.Path
		if len(path) > pathWidth {
			path = path[:pathWidth-3] + "..."
		}

		fmt.Printf("  %-*s  %-*s\n",
			nameWidth, name,
			pathWidth, path)
	}
}
