package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-work/jindo/internal/pkg/repo"
	"github.com/spf13/cobra"
)

var pkgRepoListJSON bool

var pkgRepoListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List registered repositories",
	Long:    `List all registered GitHub repositories.`,
	RunE:    runPkgRepoList,
}

func init() {
	pkgRepoCmd.AddCommand(pkgRepoListCmd)
	pkgRepoListCmd.Flags().BoolVar(&pkgRepoListJSON, "json", false, "Output in JSON format")
}

func runPkgRepoList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	store := repo.NewStore("~/.itda-jindo")

	repos, err := store.List()
	if err != nil {
		return fmt.Errorf("list repositories: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("No repositories registered.")
		fmt.Println()
		fmt.Println("Add a repository with:")
		fmt.Println("  jd pkg repo add gh:owner/repo")
		return nil
	}

	if pkgRepoListJSON {
		output, err := json.MarshalIndent(repos, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	// Calculate column widths
	nsWidth := len("NAMESPACE")
	urlWidth := len("URL")
	branchWidth := len("BRANCH")

	for _, r := range repos {
		if len(r.Namespace) > nsWidth {
			nsWidth = len(r.Namespace)
		}
		if len(r.URL) > urlWidth {
			urlWidth = len(r.URL)
		}
		if len(r.DefaultBranch) > branchWidth {
			branchWidth = len(r.DefaultBranch)
		}
	}

	// Cap widths
	if nsWidth > 20 {
		nsWidth = 20
	}
	if urlWidth > 50 {
		urlWidth = 50
	}
	if branchWidth > 15 {
		branchWidth = 15
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s\n",
		nsWidth, "NAMESPACE",
		urlWidth, "URL",
		branchWidth, "BRANCH")
	fmt.Printf("%s  %s  %s\n",
		strings.Repeat("-", nsWidth),
		strings.Repeat("-", urlWidth),
		strings.Repeat("-", branchWidth))

	// Print rows
	for _, r := range repos {
		ns := r.Namespace
		if len(ns) > nsWidth {
			ns = ns[:nsWidth-3] + "..."
		}

		url := r.URL
		if len(url) > urlWidth {
			url = url[:urlWidth-3] + "..."
		}

		branch := r.DefaultBranch
		if len(branch) > branchWidth {
			branch = branch[:branchWidth-3] + "..."
		}

		fmt.Printf("%-*s  %-*s  %-*s\n",
			nsWidth, ns,
			urlWidth, url,
			branchWidth, branch)
	}

	fmt.Printf("\nTotal: %d repositories\n", len(repos))
	return nil
}
