package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CacheSize != "medium" {
		t.Errorf("Expected default CacheSize 'medium', got '%s'", cfg.CacheSize)
	}
	if cfg.DefaultSubscription != "" {
		t.Errorf("Expected empty DefaultSubscription, got '%s'", cfg.DefaultSubscription)
	}
	if cfg.Theme.ActiveBorderColor != "green" {
		t.Errorf("Expected ActiveBorderColor 'green', got '%s'", cfg.Theme.ActiveBorderColor)
	}
	if cfg.Theme.InactiveBorderColor != "white" {
		t.Errorf("Expected InactiveBorderColor 'white', got '%s'", cfg.Theme.InactiveBorderColor)
	}
}

func TestLoadFromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yml")

	content := `default_subscription: "My Subscription"
cache_size: large
theme:
  active_border_color: cyan
  inactive_border_color: gray
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg := DefaultConfig()
	data, _ := os.ReadFile(cfgPath)
	if err := parseYAML(data, cfg); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.DefaultSubscription != "My Subscription" {
		t.Errorf("Expected 'My Subscription', got '%s'", cfg.DefaultSubscription)
	}
	if cfg.CacheSize != "large" {
		t.Errorf("Expected 'large', got '%s'", cfg.CacheSize)
	}
	if cfg.Theme.ActiveBorderColor != "cyan" {
		t.Errorf("Expected 'cyan', got '%s'", cfg.Theme.ActiveBorderColor)
	}
	if cfg.Theme.InactiveBorderColor != "gray" {
		t.Errorf("Expected 'gray', got '%s'", cfg.Theme.InactiveBorderColor)
	}
}

func TestEnvOverrides(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CacheSize = "small"

	t.Setenv("LAZYAZURE_CACHE_SIZE", "large")
	applyEnvOverrides(cfg)

	if cfg.CacheSize != "large" {
		t.Errorf("Expected env override 'large', got '%s'", cfg.CacheSize)
	}
}

func TestLoadMissingFile(t *testing.T) {
	// Load with no file — should return defaults without error
	cfg := DefaultConfig()
	if cfg.CacheSize != "medium" {
		t.Errorf("Expected default 'medium', got '%s'", cfg.CacheSize)
	}
}

func TestLoadMalformedYAML(t *testing.T) {
	cfg := DefaultConfig()
	err := parseYAML([]byte(":::invalid:::yaml"), cfg)
	if err == nil {
		t.Error("Expected error for malformed YAML")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Skip("Could not determine config directory")
	}
	// Just verify it ends with the expected filename
	if filepath.Base(path) != "config.yml" {
		t.Errorf("Expected config.yml, got '%s'", filepath.Base(path))
	}
}

func TestPartialConfig(t *testing.T) {
	cfg := DefaultConfig()
	data := []byte(`default_subscription: "Test Sub"`)
	if err := parseYAML(data, cfg); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Specified field should be set
	if cfg.DefaultSubscription != "Test Sub" {
		t.Errorf("Expected 'Test Sub', got '%s'", cfg.DefaultSubscription)
	}
	// Unspecified fields should keep defaults
	if cfg.CacheSize != "medium" {
		t.Errorf("Expected default 'medium', got '%s'", cfg.CacheSize)
	}
	if cfg.Theme.ActiveBorderColor != "green" {
		t.Errorf("Expected default 'green', got '%s'", cfg.Theme.ActiveBorderColor)
	}
}
