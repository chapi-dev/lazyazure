package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Config holds all user configuration
type Config struct {
	// DefaultSubscription is the name or ID of the subscription to auto-select on startup
	DefaultSubscription string `yaml:"default_subscription"`

	// CacheSize controls the cache size: small, medium (default), large
	CacheSize string `yaml:"cache_size"`

	// Theme customizes UI colors
	Theme ThemeConfig `yaml:"theme"`
}

// ThemeConfig holds theme/color customization
type ThemeConfig struct {
	// ActiveBorderColor is the color for focused panel borders (default: green)
	ActiveBorderColor string `yaml:"active_border_color"`
	// InactiveBorderColor is the color for unfocused panel borders (default: white)
	InactiveBorderColor string `yaml:"inactive_border_color"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		CacheSize: "medium",
		Theme: ThemeConfig{
			ActiveBorderColor:   "green",
			InactiveBorderColor: "white",
		},
	}
}

// configDir returns the config directory path
func configDir() string {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "lazyazure")
		}
	}

	// XDG_CONFIG_HOME or ~/.config
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return filepath.Join(xdgConfig, "lazyazure")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "lazyazure")
}

// ConfigPath returns the full path to the config file
func ConfigPath() string {
	dir := configDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "config.yml")
}

// Load reads the config from the default config file, falling back to defaults.
// Environment variables override file values.
func Load() *Config {
	cfg := DefaultConfig()

	path := ConfigPath()
	if path == "" {
		applyEnvOverrides(cfg)
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist or can't be read — use defaults
		applyEnvOverrides(cfg)
		return cfg
	}

	if err := parseYAML(data, cfg); err != nil {
		// Malformed YAML — use defaults
		cfg = DefaultConfig()
		applyEnvOverrides(cfg)
		return cfg
	}

	applyEnvOverrides(cfg)
	return cfg
}

// parseYAML unmarshals YAML data into a Config struct
func parseYAML(data []byte, cfg *Config) error {
	return yaml.Unmarshal(data, cfg)
}

// applyEnvOverrides lets environment variables take precedence over file values
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("LAZYAZURE_CACHE_SIZE"); v != "" {
		cfg.CacheSize = v
	}
}
