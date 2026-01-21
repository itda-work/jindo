package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itda-work/jindo/internal/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

var pkgListJSON bool

var pkgListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List installed packages",
	Long:    `List all installed packages from registered repositories.`,
	RunE:    runPkgList,
}

func init() {
	pkgCmd.AddCommand(pkgListCmd)
	pkgListCmd.Flags().BoolVar(&pkgListJSON, "json", false, "Output in JSON format")
}

func runPkgList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	manager := pkgmgr.NewManager("~/.itda-jindo")

	packages, err := manager.List()
	if err != nil {
		return fmt.Errorf("list packages: %w", err)
	}

	if len(packages) == 0 {
		fmt.Println("No packages installed.")
		fmt.Println()
		fmt.Println("Install a package with:")
		fmt.Println("  jd pkg install <namespace>:<path>")
		return nil
	}

	if pkgListJSON {
		output, err := json.MarshalIndent(packages, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	}

	// Calculate column widths
	nameWidth := len("NAME")
	typeWidth := len("TYPE")
	nsWidth := len("NAMESPACE")
	versionWidth := len("VERSION")

	for _, pkg := range packages {
		if len(pkg.Name) > nameWidth {
			nameWidth = len(pkg.Name)
		}
		typeStr := string(pkg.Type)
		if len(typeStr) > typeWidth {
			typeWidth = len(typeStr)
		}
		if len(pkg.Namespace) > nsWidth {
			nsWidth = len(pkg.Namespace)
		}
		version := pkg.Version.SHA
		if len(version) > 8 {
			version = version[:8]
		}
		if len(version) > versionWidth {
			versionWidth = len(version)
		}
	}

	// Cap widths
	if nameWidth > 35 {
		nameWidth = 35
	}
	if typeWidth > 10 {
		typeWidth = 10
	}
	if nsWidth > 15 {
		nsWidth = 15
	}
	if versionWidth > 12 {
		versionWidth = 12
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
		nameWidth, "NAME",
		typeWidth, "TYPE",
		nsWidth, "NAMESPACE",
		versionWidth, "VERSION")
	fmt.Printf("%s  %s  %s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", typeWidth),
		strings.Repeat("-", nsWidth),
		strings.Repeat("-", versionWidth))

	// Print rows
	for _, pkg := range packages {
		name := pkg.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		typeStr := string(pkg.Type)
		if len(typeStr) > typeWidth {
			typeStr = typeStr[:typeWidth-3] + "..."
		}

		ns := pkg.Namespace
		if len(ns) > nsWidth {
			ns = ns[:nsWidth-3] + "..."
		}

		version := pkg.Version.SHA
		if len(version) > 8 {
			version = version[:8]
		}

		fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
			nameWidth, name,
			typeWidth, typeStr,
			nsWidth, ns,
			versionWidth, version)
	}

	fmt.Printf("\nTotal: %d packages\n", len(packages))
	return nil
}
