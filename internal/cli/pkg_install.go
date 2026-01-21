package cli

import (
	"errors"
	"fmt"

	"github.com/itda-work/jindo/internal/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

var pkgInstallCmd = &cobra.Command{
	Use:     "install <namespace:path[@version]>",
	Aliases: []string{"i"},
	Short:   "Install a package from a registered repository",
	Long: `Install a package from a registered repository.

The specification format is: namespace:path[@version]
- namespace: The repository namespace (from 'jd pkg repo list')
- path: The package path in the repository
- version: Optional tag or commit SHA

Examples:
  jd pkg install affa-ever:skills/web-fetch
  jd pkg install affa-ever:commands/commit.md
  jd pkg install affa-ever:skills/web-fetch@v1.2.0

Installed packages are placed in ~/.itda-jindo/ with namespace prefixes:
  ~/.itda-jindo/skills/affa-ever--web-fetch/
  ~/.itda-jindo/commands/affa-ever--commit.md`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgInstall,
}

func init() {
	pkgCmd.AddCommand(pkgInstallCmd)
}

func runPkgInstall(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	spec := args[0]

	manager := pkgmgr.NewManager("~/.itda-jindo")

	// Validate spec format
	parsedSpec, err := pkgmgr.ParseSpec(spec)
	if err != nil {
		return fmt.Errorf("invalid specification. Format: namespace:path[@version]")
	}

	// Check if repository exists
	_, err = manager.RepoStore().Get(parsedSpec.Namespace)
	if err != nil {
		return fmt.Errorf("repository '%s' not found. Register with: jd pkg repo add gh:owner/repo", parsedSpec.Namespace)
	}

	fmt.Printf("Installing %s...\n", spec)

	pkg, err := manager.Install(spec)
	if err != nil {
		if errors.Is(err, pkgmgr.ErrPackageAlreadyInstalled) {
			return fmt.Errorf("package already installed. Use 'jd pkg update %s' to update", spec)
		}
		return fmt.Errorf("install: %w", err)
	}

	fmt.Printf("Installed successfully!\n")
	fmt.Printf("  Name:      %s\n", pkg.Name)
	fmt.Printf("  Type:      %s\n", pkg.Type)
	fmt.Printf("  Version:   %s (%s)\n", pkg.Version.Ref, pkg.Version.SHA[:8])
	fmt.Printf("  Files:     %d\n", len(pkg.Files))

	if len(pkg.Files) > 0 {
		fmt.Println("\nInstalled files:")
		for _, f := range pkg.Files {
			fmt.Printf("  %s\n", f.Target)
		}
	}

	return nil
}
