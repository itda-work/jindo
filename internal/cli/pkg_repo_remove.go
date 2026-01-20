package cli

import (
	"errors"
	"fmt"

	"github.com/itda-work/itda-jindo/internal/pkg/repo"
	"github.com/spf13/cobra"
)

var pkgRepoRemoveCmd = &cobra.Command{
	Use:     "remove <namespace>",
	Aliases: []string{"rm"},
	Short:   "Remove a registered repository",
	Long: `Remove a registered repository by its namespace.

Note: This only removes the repository registration.
Installed packages from this repository will remain installed.

Example:
  jd pkg repo remove affa-ever`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgRepoRemove,
}

func init() {
	pkgRepoCmd.AddCommand(pkgRepoRemoveCmd)
}

func runPkgRepoRemove(_ *cobra.Command, args []string) error {
	namespace := args[0]

	store := repo.NewStore("~/.itda-jindo")

	// Check if exists
	config, err := store.Get(namespace)
	if err != nil {
		if errors.Is(err, repo.ErrRepoNotFound) {
			return fmt.Errorf("repository '%s' not found", namespace)
		}
		return fmt.Errorf("get repository: %w", err)
	}

	if err := store.Remove(namespace); err != nil {
		return fmt.Errorf("remove repository: %w", err)
	}

	fmt.Printf("Removed repository: %s (%s)\n", namespace, config.URL)
	return nil
}
