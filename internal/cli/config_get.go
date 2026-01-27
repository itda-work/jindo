package cli

import (
	"bytes"
	"fmt"

	"github.com/itda-skills/jindo/pkg/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

var (
	configGetEnv bool
)

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value using dot notation.

For leaf values, prints the value directly.
For nested sections, prints the section in TOML format.

Use --env to also check environment variable override.
Environment variables use the format: ITDA_<KEY> (uppercase, dots replaced with underscores)

Examples:
  jd config get common.api_keys.tiingo
  jd config get skills.quant-data
  jd config get common.api_keys.tiingo --env`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configGetCmd.Flags().BoolVarP(&configGetEnv, "env", "e", false, "Check environment variable override")
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var value any
	var found bool

	if configGetEnv {
		value, found = cfg.GetWithEnv(key)
		if !found {
			return fmt.Errorf("key not found: %s", key)
		}
	} else {
		value, err = cfg.Get(key)
		if err != nil {
			if err == config.ErrKeyNotFound {
				return fmt.Errorf("key not found: %s", key)
			}
			return fmt.Errorf("failed to get value: %w", err)
		}
	}

	// If value is a map, format as TOML
	if m, ok := value.(map[string]any); ok {
		var buf bytes.Buffer
		encoder := toml.NewEncoder(&buf)
		encoder.SetIndentTables(true)
		if err := encoder.Encode(m); err != nil {
			return fmt.Errorf("failed to format value: %w", err)
		}
		fmt.Print(buf.String())
	} else {
		fmt.Println(value)
	}

	return nil
}
