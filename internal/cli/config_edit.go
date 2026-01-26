package cli

import (
	"fmt"

	"github.com/itda-skills/jindo/pkg/config"
	"github.com/spf13/cobra"
)

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration in editor",
	Long: `Open the configuration file in your default editor.

Uses $EDITOR or $VISUAL environment variable.
Falls back to 'vi' if neither is set.

If the config file doesn't exist, it will be created first with the default template.`,
	RunE: runConfigEdit,
}

func init() {
	configCmd.AddCommand(configEditCmd)
}

func runConfigEdit(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	// Ensure config exists
	if !config.ConfigExists() {
		path, err := config.InitConfig(false)
		if err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		fmt.Printf("Created config file: %s\n", path)
	}

	path, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	return openEditor(path)
}
