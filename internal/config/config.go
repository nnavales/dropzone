// Package config parses dropzone's config file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Config is dropzone's configuration.
type Config struct {
	Settings Settings `yaml:"settings"`
	Zones    []Zone   `yaml:"zones"`
}

// Settings are global options.
type Settings struct {
	WaitSeconds int    `yaml:"wait_seconds"`
	Conflict    string `yaml:"conflict"` // skip | overwrite | rename
}

// Zone is a watched directory.
type Zone struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"`
	Rules []Rule `yaml:"rules"`
}

// Rule maps a match to an action.
type Rule struct {
	Name   string `yaml:"name"`
	Match  Match  `yaml:"match"`
	Action Action `yaml:"action"`
}

// Match filters files.
type Match struct {
	Extensions []string `yaml:"extensions,omitempty"`
	Glob       []string `yaml:"glob,omitempty"`
}

// ActionKind identifies the single operation a rule performs.
type ActionKind string

const (
	ActionMove   ActionKind = "move"
	ActionCopy   ActionKind = "copy"
	ActionRename ActionKind = "rename"
	ActionRun    ActionKind = "run"
	ActionDelete ActionKind = "delete"
)

// Action acts on a file: exactly one Kind with its Target.
// Target is unused for delete. YAML stays k:v, e.g. `move: ~/Pictures`
// or `delete: true`; the single-key shape is enforced on unmarshal.
type Action struct {
	Kind   ActionKind
	Target string
}

// UnmarshalYAML decodes a single-key map into an Action.
func (a *Action) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]any
	if err := value.Decode(&raw); err != nil {
		return err
	}

	if len(raw) != 1 {
		if len(raw) == 0 {
			return fmt.Errorf("action: one of move|copy|rename|run|delete is required")
		}
		return fmt.Errorf("action: only one of move|copy|rename|run|delete allowed, got %d", len(raw))
	}

	for kind, v := range raw {
		switch ActionKind(kind) {
		case ActionMove, ActionCopy, ActionRename, ActionRun:
			s, ok := v.(string)
			if !ok || strings.TrimSpace(s) == "" {
				return fmt.Errorf("action %q requires a non-empty string", kind)
			}
			a.Kind = ActionKind(kind)
			a.Target = s
		case ActionDelete:
			b, ok := v.(bool)
			if !ok || !b {
				return fmt.Errorf("action %q must be true", kind)
			}
			a.Kind = ActionDelete
		default:
			return fmt.Errorf("unknown action %q: must be one of move|copy|rename|run|delete", kind)
		}
	}
	return nil
}

// Load reads the config file.
// It uses $DROPZONE_CONFIG when set, otherwise the user config dir.
func Load() (Config, error) {
	return LoadFrom(ConfigPath())
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

func (c *Config) withDefaults() {
	if c.Settings.WaitSeconds == 0 {
		c.Settings.WaitSeconds = 2
	}
	if c.Settings.Conflict == "" {
		c.Settings.Conflict = "rename"
	}
}

// normalize resolves static path syntax (~, env vars) so downstream
// packages receive ready-to-use values. Dynamic placeholders
// ({date}, {stem}, ...) are per-file and must never expand here.
func (c *Config) normalize() {
	for i := range c.Zones {
		c.Zones[i].Path = expandPath(c.Zones[i].Path)
		for j := range c.Zones[i].Rules {
			a := &c.Zones[i].Rules[j].Action
			switch a.Kind {
			case ActionMove, ActionCopy:
				a.Target = expandPath(a.Target)
			}
		}
	}
}

// ConfigPath returns the active config file path.
func ConfigPath() string {
	if p := os.Getenv("DROPZONE_CONFIG"); p != "" {
		return p
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	// TODO: be able to use yml or yaml on extension.
	return filepath.Join(configDir, "dropzone", "config.yml")
}

// LoadFrom reads the config file at path.
func LoadFrom(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("empty config path")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	return Parse(data)
}

// expandPath expands env vars and a leading ~ in p.
func expandPath(p string) string {
	return resolveTilde(os.ExpandEnv(strings.TrimSpace(p)))
}

func resolveTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// Validate checks the config.
func (c Config) Validate() error {
	if c.Settings.WaitSeconds < 0 {
		return fmt.Errorf("settings.wait_seconds must be >= 0")
	}

	switch c.Settings.Conflict {
	case "skip", "overwrite", "rename", "":
	default:
		return fmt.Errorf("settings.conflict must be skip|overwrite|rename")
	}

	if len(c.Zones) == 0 {
		return fmt.Errorf("zones must not be empty")
	}

	for i, z := range c.Zones {
		if z.Path == "" {
			return fmt.Errorf("zones[%d]: path is required", i)
		}
		for j, r := range z.Rules {
			m := r.Match
			if len(m.Extensions) == 0 && len(m.Glob) == 0 {
				return fmt.Errorf("zones[%d].rules[%d].match: at least one of extensions|glob is required", i, j)
			}
			a := r.Action
			switch a.Kind {
			case ActionMove, ActionCopy, ActionRename, ActionRun:
				if strings.TrimSpace(a.Target) == "" {
					return fmt.Errorf("zones[%d].rules[%d].action %q requires a non-empty target", i, j, a.Kind)
				}
			case ActionDelete:
			default:
				return fmt.Errorf("zones[%d].rules[%d].action: one of move|copy|rename|run|delete is required", i, j)
			}
		}
	}
	return nil
}
