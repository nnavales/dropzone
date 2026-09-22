// Package config parses dropzone's config file.
package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// Config is dropzone's configuration.
type Config struct {
	Settings Settings `yaml:"settings"`
	Zones    []Zone   `yaml:"zones"`
}

// Settings are global options.
type Settings struct {
	StableForSeconds int    `yaml:"stable_for_seconds"`
	OnConflict       string `yaml:"on_conflict"` // skip | overwrite | rename
}

// Zone is a watched directory.
type Zone struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"`
	Rules []Rule `yaml:"rules"`
}

// Rule maps a match to an action.
type Rule struct {
	Name       string `yaml:"name"`
	Match      Match  `yaml:"match"`
	OnConflict string `yaml:"on_conflict,omitempty"` // on_conflict (optional) overrides global settings.
	Action     Action `yaml:"action"`
}

// Match filters files.
type Match struct {
	Extensions []string `yaml:"extensions,omitempty"`
	Glob       []string `yaml:"glob,omitempty"`
}

// Action acts on a file
type Action struct {
	Kind   string // move | copy | rename | run | delete
	Target string
}

// Load reads the config file at path.
func Load(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("empty config path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	cfg, err := Parse(data)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Parse reads YAML into a Config.
func Parse(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.withDefaults()
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
