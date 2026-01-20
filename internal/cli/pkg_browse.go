package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-work/itda-jindo/internal/pkg/repo"
	"github.com/spf13/cobra"
)

var (
	pkgBrowseType string
	pkgBrowseJSON bool
)

var pkgBrowseCmd = &cobra.Command{
	Use:   "browse <namespace>",
	Short: "Browse packages in a repository",
	Long: `Browse available packages (skills, commands, agents) in a registered repository.

Use --type to filter by package type.

Examples:
  jd pkg browse affa-ever
  jd pkg browse affa-ever --type skills
  jd pkg browse affa-ever --type commands`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgBrowse,
}

func init() {
	pkgCmd.AddCommand(pkgBrowseCmd)
	pkgBrowseCmd.Flags().StringVarP(&pkgBrowseType, "type", "t", "", "Filter by type (skills, commands, agents)")
	pkgBrowseCmd.Flags().BoolVar(&pkgBrowseJSON, "json", false, "Output in JSON format")
}

func runPkgBrowse(_ *cobra.Command, args []string) error {
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
	default:
		return fmt.Errorf("invalid type: %s (use: skills, commands, agents)", pkgBrowseType)
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

	for _, item := range items {
		switch item.Type {
		case repo.TypeSkill:
			skills = append(skills, item)
		case repo.TypeCommand:
			commands = append(commands, item)
		case repo.TypeAgent:
			agents = append(agents, item)
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
