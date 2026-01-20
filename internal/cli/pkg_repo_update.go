package cli

import (
	"fmt"

	"github.com/itda-work/itda-jindo/internal/pkg/repo"
	"github.com/spf13/cobra"
)

var pkgRepoUpdateCmd = &cobra.Command{
	Use:   "update [namespace...]",
	Short: "Update registered repositories",
	Long: `Pull the latest changes for registered repositories.

Without arguments, updates all registered repositories.
With arguments, updates only the specified repositories.

Examples:
  jd pkg repo update              # Update all
  jd pkg repo update affa-ever    # Update specific repo`,
	RunE: runPkgRepoUpdate,
}

func init() {
	pkgRepoCmd.AddCommand(pkgRepoUpdateCmd)
}

func runPkgRepoUpdate(_ *cobra.Command, args []string) error {
	store := repo.NewStore("~/.itda-jindo")

	if len(args) == 0 {
		// Update all
		fmt.Println("Updating all repositories...")
		return store.UpdateAll()
	}

	// Update specific repos
	for _, namespace := range args {
		fmt.Printf("Updating %s...\n", namespace)
		if err := store.Update(namespace); err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		fmt.Println("  Done")
	}

	return nil
}
