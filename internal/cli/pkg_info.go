package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/itda-skills/jindo/internal/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

var pkgInfoJSON bool

var pkgInfoCmd = &cobra.Command{
	Use:     "info <name>",
	Aliases: []string{"in"},
	Short:   "Show detailed information about an installed package",
	Long: `Show detailed information about an installed package.

Use 'jd pkg list' to see installed package names.

Example:
  jd pkg info affa-ever--web-fetch`,
	Args: cobra.ExactArgs(1),
	RunE: runPkgInfo,
}

func init() {
	pkgCmd.AddCommand(pkgInfoCmd)
	pkgInfoCmd.Flags().BoolVar(&pkgInfoJSON, "json", false, "Output in JSON format")
}

func runPkgInfo(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	name := args[0]

	manager := pkgmgr.NewManager("~/.itda-skills")

	pkg, err := manager.Get(name)
	if err != nil {
		if errors.Is(err, pkgmgr.ErrPackageNotFound) {
			return fmt.Errorf("package '%s' not found. Use 'jd pkg list' to see installed packages", name)
		}
		return fmt.Errorf("get package: %w", err)
	}

	if pkgInfoJSON {
		output, err := json.MarshalIndent(pkg, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("Name:          %s\n", pkg.Name)
	fmt.Printf("Original Name: %s\n", pkg.OriginalName)
	fmt.Printf("Type:          %s\n", pkg.Type)
	fmt.Printf("Namespace:     %s\n", pkg.Namespace)
	fmt.Printf("Source Path:   %s\n", pkg.SourcePath)
	fmt.Printf("Version Type:  %s\n", pkg.Version.Type)
	fmt.Printf("Version SHA:   %s\n", pkg.Version.SHA)
	fmt.Printf("Version Ref:   %s\n", pkg.Version.Ref)
	fmt.Printf("Installed At:  %s\n", pkg.InstalledAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated At:    %s\n", pkg.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Files:         %d\n", len(pkg.Files))

	if len(pkg.Files) > 0 {
		fmt.Println("\nInstalled Files:")
		for _, f := range pkg.Files {
			fmt.Printf("  Source: %s\n", f.Source)
			fmt.Printf("  Target: %s\n", f.Target)
			fmt.Printf("  SHA:    %s\n", f.SHA)
			fmt.Println()
		}
	}

	return nil
}
