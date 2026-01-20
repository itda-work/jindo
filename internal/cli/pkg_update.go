package cli

import (
	"fmt"
	"strings"

	"github.com/itda-work/itda-jindo/internal/pkg/pkgmgr"
	"github.com/spf13/cobra"
)

var pkgUpdateApply bool

var pkgUpdateCmd = &cobra.Command{
	Use:   "update [name...]",
	Short: "Check for and apply package updates",
	Long: `Check for updates to installed packages.

Without --apply, shows available updates.
With --apply, downloads and installs updates.

Examples:
  jd pkg update                    # Check all packages
  jd pkg update affa-ever--web-fetch  # Check specific package
  jd pkg update --apply            # Apply all updates`,
	RunE: runPkgUpdate,
}

func init() {
	pkgCmd.AddCommand(pkgUpdateCmd)
	pkgUpdateCmd.Flags().BoolVar(&pkgUpdateApply, "apply", false, "Apply available updates")
}

func runPkgUpdate(_ *cobra.Command, args []string) error {
	manager := pkgmgr.NewManager("~/.itda-jindo")

	fmt.Println("Checking for updates...")

	updates, err := manager.CheckUpdates(args...)
	if err != nil {
		return fmt.Errorf("check updates: %w", err)
	}

	if len(updates) == 0 {
		fmt.Println("No packages to check.")
		return nil
	}

	// Count packages with updates
	updateCount := 0
	for _, u := range updates {
		if u.HasUpdate {
			updateCount++
		}
	}

	if updateCount == 0 {
		fmt.Println("All packages are up to date.")
		return nil
	}

	// Calculate column widths
	nameWidth := len("NAME")
	currentWidth := len("CURRENT")
	latestWidth := len("LATEST")
	changesWidth := len("CHANGES")

	for _, u := range updates {
		if !u.HasUpdate {
			continue
		}
		if len(u.Package.Name) > nameWidth {
			nameWidth = len(u.Package.Name)
		}
		current := u.CurrentSHA
		if len(current) > 8 {
			current = current[:8]
		}
		if len(current) > currentWidth {
			currentWidth = len(current)
		}
		latest := u.LatestSHA
		if len(latest) > 8 {
			latest = latest[:8]
		}
		if len(latest) > latestWidth {
			latestWidth = len(latest)
		}
		changes := fmt.Sprintf("%d files", len(u.ChangedFiles))
		if len(changes) > changesWidth {
			changesWidth = len(changes)
		}
	}

	// Cap widths
	if nameWidth > 35 {
		nameWidth = 35
	}
	if currentWidth > 12 {
		currentWidth = 12
	}
	if latestWidth > 12 {
		latestWidth = 12
	}
	if changesWidth > 15 {
		changesWidth = 15
	}

	fmt.Printf("\n%d package(s) have updates available:\n\n", updateCount)

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
		nameWidth, "NAME",
		currentWidth, "CURRENT",
		latestWidth, "LATEST",
		changesWidth, "CHANGES")
	fmt.Printf("%s  %s  %s  %s\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", currentWidth),
		strings.Repeat("-", latestWidth),
		strings.Repeat("-", changesWidth))

	// Print rows
	for _, u := range updates {
		if !u.HasUpdate {
			continue
		}

		name := u.Package.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}

		current := u.CurrentSHA
		if len(current) > 8 {
			current = current[:8]
		}

		latest := u.LatestSHA
		if len(latest) > 8 {
			latest = latest[:8]
		}

		changes := fmt.Sprintf("%d files", len(u.ChangedFiles))

		fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
			nameWidth, name,
			currentWidth, current,
			latestWidth, latest,
			changesWidth, changes)
	}

	if !pkgUpdateApply {
		fmt.Println()
		fmt.Println("Run with --apply to install updates:")
		fmt.Println("  jd pkg update --apply")
		return nil
	}

	// Apply updates
	fmt.Println()
	fmt.Println("Applying updates...")

	successCount := 0
	for _, u := range updates {
		if !u.HasUpdate {
			continue
		}

		fmt.Printf("  Updating %s... ", u.Package.Name)
		_, err := manager.Update(u.Package.Name)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}
		fmt.Println("OK")
		successCount++
	}

	fmt.Printf("\nUpdated %d of %d packages.\n", successCount, updateCount)
	return nil
}
