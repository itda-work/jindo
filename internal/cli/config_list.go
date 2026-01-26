package cli

import (
	"fmt"
	"os"

	"github.com/itda-skills/jindo/pkg/config"
	"github.com/spf13/cobra"
)

var (
	configListRaw bool
)

var configListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "Show all configuration settings",
	Long: `Display the entire configuration file content.

Output is formatted as TOML.
Use --raw to show the raw file content (preserving comments).`,
	RunE: runConfigList,
}

func init() {
	configCmd.AddCommand(configListCmd)
	configListCmd.Flags().BoolVar(&configListRaw, "raw", false, "Output raw file content (preserves comments)")
}

func runConfigList(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	if configListRaw {
		return showRawConfig()
	}

	return showParsedConfig()
}

func showRawConfig() error {
	path, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if !config.ConfigExists() {
		return fmt.Errorf("config file not found: %s\nRun 'jd config init' to create it", path)
	}

	content, err := readFileContent(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	fmt.Print(content)
	return nil
}

func showParsedConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.IsEmpty() {
		path, _ := config.GetConfigPath()
		if !config.ConfigExists() {
			return fmt.Errorf("config file not found: %s\nRun 'jd config init' to create it", path)
		}
		fmt.Println("# Config is empty")
		return nil
	}

	tomlStr, err := cfg.ToTOML()
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}

	fmt.Print(tomlStr)
	return nil
}

func readFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
