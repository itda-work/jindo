package cli

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Short:   "Manage itda-skills configuration",
	Long: `Manage the unified configuration for itda-skills and all skills.

Configuration is stored in ~/.config/itda-skills/config.toml (TOML format).
Use dot notation to access nested values: common.api_keys.tiingo

Examples:
  jd config init                              # Initialize config file
  jd config set common.api_keys.tiingo KEY    # Set a value
  jd config get common.api_keys.tiingo        # Get a value
  jd config list                              # Show all settings
  jd config edit                              # Open in editor`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
