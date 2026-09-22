package config

import (
	"fmt"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Marshal writes a Config as YAML.
func Marshal(cfg Config) ([]byte, error) {
	return yaml.Marshal(cfg)
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
		switch kind {
		case "move", "copy", "rename", "run":
			s, ok := v.(string)
			if !ok || strings.TrimSpace(s) == "" {
				return fmt.Errorf("action %q requires a non-empty string", kind)
			}
			a.Kind = kind
			a.Target = s
		case "delete":
			b, ok := v.(bool)
			if !ok || !b {
				return fmt.Errorf("action %q must be true", kind)
			}
			a.Kind = "delete"
		default:
			return fmt.Errorf("unknown action %q: must be one of move|copy|rename|run|delete", kind)
		}
	}
	return nil
}
