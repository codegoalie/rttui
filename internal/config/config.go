package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config holds application configuration loaded from $XDG_CONFIG_HOME/rttui/config.json.
type Config struct {
	DefaultFilter string `json:"default_filter"`
	AddPreset     string `json:"add_preset"`
}

// Load reads config from disk. Returns empty defaults if the file doesn't exist.
func Load() (Config, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}
	p := filepath.Join(dir, "rttui", "config.json")
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
