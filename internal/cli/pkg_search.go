package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/itda-work/jindo/internal/pkg/repo"
	"github.com/spf13/cobra"
)

var pkgSearchJSON bool

var pkgSearchCmd = &cobra.Command{
	Use:     "search <query>",
	Aliases: []string{"se"},
	Short:   "Search for packages across all registered repositories",
	Long: `Search for packages by name across all registered repositories.

The search is case-insensitive and matches package names containing the query.

Examples:
  jd pkg search web
  jd pkg search commit`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgSearch,
}

func init() {
	pkgCmd.AddCommand(pkgSearchCmd)
	pkgSearchCmd.Flags().BoolVar(&pkgSearchJSON, "json", false, "Output in JSON format")
}

func runPkgSearch(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	query := args[0]

	store := repo.NewStore("~/.itda-jindo")

	results, err := store.Search(query)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No packages found matching '%s'.\n", query)
		return nil
	}

	if pkgSearchJSON {
		output, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	// Sort namespaces for consistent output
	namespaces := make([]string, 0, len(results))
	for ns := range results {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	totalCount := 0
	for _, ns := range namespaces {
		items := results[ns]
		totalCount += len(items)

		fmt.Printf("%s:\n", ns)

		// Calculate column widths
		nameWidth := len("NAME")
		typeWidth := len("TYPE")
		pathWidth := len("PATH")

		for _, item := range items {
			if len(item.Name) > nameWidth {
				nameWidth = len(item.Name)
			}
			typeStr := string(item.Type)
			if len(typeStr) > typeWidth {
				typeWidth = len(typeStr)
			}
			if len(item.Path) > pathWidth {
				pathWidth = len(item.Path)
			}
		}

		// Cap widths
		if nameWidth > 25 {
			nameWidth = 25
		}
		if typeWidth > 10 {
			typeWidth = 10
		}
		if pathWidth > 45 {
			pathWidth = 45
		}

		// Print header
		fmt.Printf("  %-*s  %-*s  %-*s\n",
			nameWidth, "NAME",
			typeWidth, "TYPE",
			pathWidth, "PATH")
		fmt.Printf("  %s  %s  %s\n",
			strings.Repeat("-", nameWidth),
			strings.Repeat("-", typeWidth),
			strings.Repeat("-", pathWidth))

		// Print rows
		for _, item := range items {
			name := item.Name
			if len(name) > nameWidth {
				name = name[:nameWidth-3] + "..."
			}

			typeStr := string(item.Type)
			if len(typeStr) > typeWidth {
				typeStr = typeStr[:typeWidth-3] + "..."
			}

			path := item.Path
			if len(path) > pathWidth {
				path = path[:pathWidth-3] + "..."
			}

			fmt.Printf("  %-*s  %-*s  %-*s\n",
				nameWidth, name,
				typeWidth, typeStr,
				pathWidth, path)
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d packages in %d repositories\n", totalCount, len(results))
	return nil
}
