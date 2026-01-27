package config

import (
	"bytes"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the hierarchical configuration structure
type Config struct {
	data map[string]any
}

// New creates an empty Config
func New() *Config {
	return &Config{
		data: make(map[string]any),
	}
}

// Load reads config from the default path
// Returns an empty config if file doesn't exist
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadFromPath(path)
}

// LoadFromPath reads config from a specific path
// Returns an empty config if file doesn't exist
func LoadFromPath(path string) (*Config, error) {
	c := New()

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}

	if err := toml.Unmarshal(content, &c.data); err != nil {
		return nil, err
	}

	return c, nil
}

// Save writes config to the default path
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	return c.SaveToPath(path)
}

// SaveToPath writes config to a specific path
func (c *Config) SaveToPath(path string) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	encoder.SetIndentTables(true)
	if err := encoder.Encode(c.data); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// Get retrieves a value using dot notation (e.g., "common.api_keys.tiingo")
func (c *Config) Get(key string) (any, error) {
	keys, err := parseDotKey(key)
	if err != nil {
		return nil, err
	}
	return getNestedValue(c.data, keys)
}

// Set sets a value using dot notation, creating intermediate maps as needed
func (c *Config) Set(key string, value any) error {
	keys, err := parseDotKey(key)
	if err != nil {
		return err
	}
	return setNestedValue(c.data, keys, value)
}

// Delete removes a key using dot notation
func (c *Config) Delete(key string) error {
	keys, err := parseDotKey(key)
	if err != nil {
		return err
	}
	return deleteNestedValue(c.data, keys)
}

// GetWithEnv retrieves a value, checking environment variable first
// Env var format: ITDA_<SECTION>_<KEY> (uppercase, underscores)
// Returns the value and a boolean indicating if found
func (c *Config) GetWithEnv(key string) (any, bool) {
	// Check environment variable first
	envKey := toEnvKey(key)
	if envVal := os.Getenv(envKey); envVal != "" {
		return ParseValue(envVal), true
	}

	// Fall back to config file
	val, err := c.Get(key)
	if err != nil {
		return nil, false
	}
	return val, true
}

// ToMap returns the full config as a nested map
func (c *Config) ToMap() map[string]any {
	return c.data
}

// IsEmpty returns true if config has no values
func (c *Config) IsEmpty() bool {
	return len(c.data) == 0
}

// ToTOML returns the config as a TOML string
func (c *Config) ToTOML() (string, error) {
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	encoder.SetIndentTables(true)
	if err := encoder.Encode(c.data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// DefaultTemplate returns the default config file template
const DefaultTemplate = `# itda-skills Configuration
# https://github.com/itda-skills/jindo

[common]
# default_market = "kr"

[common.api_keys]
# tiingo = "your-api-key"
# polygon = "your-api-key"

# [skills.quant-data]
# default_format = "json"

# [skills.igm]
# Add igm-specific settings here

# [skills.hangul]
# Add hangul-specific settings here

# [skills.web-auto]
# Add web-auto-specific settings here
`

// InitConfig creates a new config file with the default template
// Returns error if file already exists and force is false
func InitConfig(force bool) (string, error) {
	path, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	if !force && ConfigExists() {
		return path, os.ErrExist
	}

	if err := EnsureConfigDir(); err != nil {
		return "", err
	}

	if err := os.WriteFile(path, []byte(DefaultTemplate), 0644); err != nil {
		return "", err
	}

	return path, nil
}
