package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Validate checks the config.
func (c Config) Validate() error {
	if c.Settings.StableForSeconds < 0 {
		return fmt.Errorf("settings.stable_for_seconds must be >= 0")
	}

	switch c.Settings.OnConflict {
	case "skip", "overwrite", "rename", "":
	default:
		return fmt.Errorf("settings.on_conflict must be skip|overwrite|rename")
	}

	seen := map[string]int{}
	for i, z := range c.Zones {
		if z.Path == "" {
			return fmt.Errorf("zones[%d]: path is required", i)
		}
		if strings.IndexByte(z.Path, 0) >= 0 {
			return fmt.Errorf("zones[%d]: path must not contain null bytes", i)
		}
		if !filepath.IsAbs(z.Path) {
			return fmt.Errorf("zones[%d]: path must be absolute", i)
		}
		if first, ok := seen[z.Path]; ok {
			return fmt.Errorf("zones[%d]: duplicate path %q (also zones[%d])", i, z.Path, first)
		}
		seen[z.Path] = i
		for j, r := range z.Rules {
			switch r.OnConflict {
			case "skip", "overwrite", "rename":
			default:
				return fmt.Errorf("zones[%d].rules[%d].on_conflict must be skip|overwrite|rename", i, j)
			}
			m := r.Match
			if len(m.Extensions) == 0 && len(m.Glob) == 0 {
				return fmt.Errorf("zones[%d].rules[%d].match: at least one of extensions|glob is required", i, j)
			}
			for _, g := range m.Glob {
				if err := validateGlob(g); err != nil {
					return fmt.Errorf("zones[%d].rules[%d].match: %w", i, j, err)
				}
			}
			for _, e := range m.Extensions {
				if err := validateExtension(e); err != nil {
					return fmt.Errorf("zones[%d].rules[%d].match: %w", i, j, err)
				}
			}
			a := r.Action
			switch a.Kind {
			case "move", "copy", "rename", "run":
				if strings.TrimSpace(a.Target) == "" {
					return fmt.Errorf("zones[%d].rules[%d].action %q requires a non-empty target", i, j, a.Kind)
				}
			case "delete":
			default:
				return fmt.Errorf("zones[%d].rules[%d].action: one of move|copy|rename|run|delete is required", i, j)
			}
		}
	}
	return nil
}

func validateGlob(pattern string) error {
	if strings.TrimSpace(pattern) == "" {
		return fmt.Errorf("glob must not be empty")
	}
	if _, err := filepath.Match(pattern, ""); err != nil {
		return fmt.Errorf("invalid glob %q: %w", pattern, err)
	}
	return nil
}

func validateExtension(ext string) error {
	e := strings.TrimSpace(ext)
	e = strings.TrimPrefix(e, ".")
	if e == "" {
		return fmt.Errorf("extension must not be empty")
	}
	if strings.ContainsAny(e, `/\.*+?[]{}!`) {
		return fmt.Errorf("invalid extension %q", ext)
	}
	return nil
}
