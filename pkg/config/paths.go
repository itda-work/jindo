package config

import (
	"os"
	"path/filepath"
)

const (
	// AppName is the application name used in config paths
	AppName = "itda-skills"
	// ConfigFileName is the name of the configuration file
	ConfigFileName = "config.toml"
)

// GetConfigDir returns the config directory path.
// Respects XDG_CONFIG_HOME environment variable, defaults to ~/.config/itda-skills/
func GetConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, AppName), nil
}

// GetConfigPath returns the full config file path (~/.config/itda-skills/config.toml)
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// ConfigExists checks if config file exists
func ConfigExists() bool {
	path, err := GetConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
