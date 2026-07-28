package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config represents the agent configuration loaded from JSON file.
type Config struct {
	BackendURL      string        `json:"backend_url"`
	ServerKey       string        `json:"server_key"`
	IntervalSeconds int           `json:"interval_seconds"`
	Collect         CollectConfig `json:"collect"`
	LogFile         string        `json:"log_file,omitempty"`
}

// CollectConfig specifies which metrics to collect.
type CollectConfig struct {
	System   bool     `json:"system"`
	Docker   bool     `json:"docker"`
	PM2      bool     `json:"pm2"`
	Services []string `json:"services"`
}

// MaskedKey returns the server key with only the last 6 characters visible.
func (c *Config) MaskedKey() string {
	if len(c.ServerKey) <= 6 {
		return strings.Repeat("*", len(c.ServerKey))
	}
	return strings.Repeat("*", len(c.ServerKey)-6) + c.ServerKey[len(c.ServerKey)-6:]
}

// Load loads configuration from the given path, or searches default locations.
// Search order: /etc/p-mon-agent.json, ~/.p-mon-agent.json
func Load(path string) (*Config, error) {
	if path != "" {
		return loadFromFile(path)
	}

	paths := defaultConfigPaths()
	for _, p := range paths {
		cfg, err := loadFromFile(p)
		if err == nil {
			return cfg, nil
		}
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error reading %s: %w", p, err)
		}
	}

	return nil, fmt.Errorf("config file not found in: %s", strings.Join(paths, ", "))
}

func defaultConfigPaths() []string {
	paths := []string{"/etc/p-mon-agent.json"}

	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	if home != "" {
		paths = append(paths, filepath.Join(home, ".p-mon-agent.json"))
	}

	return paths
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(data)
}

func parse(data []byte) (*Config, error) {
	cfg := &Config{
		IntervalSeconds: 30,
		Collect: CollectConfig{
			System: true,
			Docker: true,
			PM2:    true,
		},
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.BackendURL == "" {
		return nil, fmt.Errorf("backend_url is required")
	}
	if cfg.ServerKey == "" {
		return nil, fmt.Errorf("server_key is required")
	}
	if cfg.IntervalSeconds <= 0 {
		cfg.IntervalSeconds = 30
	}

	// Normalize: strip trailing slash from backend URL
	cfg.BackendURL = strings.TrimRight(cfg.BackendURL, "/")

	return cfg, nil
}
