package cli

import (
	"fmt"

	"github.com/itda-skills/jindo/pkg/config"
	"github.com/spf13/cobra"
)

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value using dot notation.

The value type is automatically inferred:
  - "true" / "false" -> boolean
  - numeric strings -> integer or float
  - other strings -> string

Examples:
  jd config set common.default_market kr
  jd config set common.api_keys.tiingo YOUR_API_KEY
  jd config set skills.quant-data.default_format table
  jd config set skills.quant-data.sources.krx.delay 1000`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func init() {
	configCmd.AddCommand(configSetCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	key := args[0]
	rawValue := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse the value to appropriate type
	value := config.ParseValue(rawValue)

	if err := cfg.Set(key, value); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s = %v\n", key, value)
	return nil
}
