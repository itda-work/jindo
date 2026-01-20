package cli

import (
	"errors"
	"fmt"

	"github.com/itda-work/itda-jindo/internal/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

var pkgUninstallCmd = &cobra.Command{
	Use:     "uninstall <name>",
	Aliases: []string{"rm", "remove"},
	Short:   "Uninstall an installed package",
	Long: `Uninstall a package by its installed name.

Use 'jd pkg list' to see installed package names.

Example:
  jd pkg uninstall affa-ever--web-fetch`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgUninstall,
}

func init() {
	pkgCmd.AddCommand(pkgUninstallCmd)
}

func runPkgUninstall(_ *cobra.Command, args []string) error {
	name := args[0]

	manager := pkgmgr.NewManager("~/.itda-jindo")

	// Get package info first for display
	pkg, err := manager.Get(name)
	if err != nil {
		if errors.Is(err, pkgmgr.ErrPackageNotFound) {
			return fmt.Errorf("package '%s' not found. Use 'jd pkg list' to see installed packages", name)
		}
		return fmt.Errorf("get package: %w", err)
	}

	if err := manager.Uninstall(name); err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}

	fmt.Printf("Uninstalled: %s (%s)\n", pkg.Name, pkg.Type)
	return nil
}
