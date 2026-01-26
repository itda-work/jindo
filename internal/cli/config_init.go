package cli

import (
	"fmt"
	"os"

	"github.com/itda-skills/jindo/pkg/config"
	"github.com/spf13/cobra"
)

var (
	configInitForce bool
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long: `Create a new configuration file with default template.

The config file is created at ~/.config/itda-skills/config.toml.
If the file already exists, use --force to overwrite it.`,
	RunE: runConfigInit,
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configInitCmd.Flags().BoolVarP(&configInitForce, "force", "f", false, "Overwrite existing config without confirmation")
}

func runConfigInit(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	path, err := config.InitConfig(configInitForce)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("config file already exists: %s\nUse --force to overwrite", path)
		}
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Printf("Created config file: %s\n", path)
	return nil
}
