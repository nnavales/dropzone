package config

import (
	"os"
	"path/filepath"
	"strings"
)

func (c *Config) normalize() {
	for i := range c.Zones {
		c.Zones[i].Path = cleanPath(c.Zones[i].Path)
		for j := range c.Zones[i].Rules {
			// Canonicalize extensions: ".jpg" and "jpg" are the same.
			for k, e := range c.Zones[i].Rules[j].Match.Extensions {
				c.Zones[i].Rules[j].Match.Extensions[k] = strings.TrimPrefix(strings.TrimSpace(e), ".")
			}
			a := &c.Zones[i].Rules[j].Action
			switch a.Kind {
			case "move", "copy":
				a.Target = cleanPath(a.Target)
			}
		}
	}
}

// cleanPath expands and cleans path.
func cleanPath(p string) string {
	if strings.TrimSpace(p) == "" {
		return ""
	}
	return filepath.Clean(expandPath(p))
}

// expandPath expands env vars and a leading ~
func expandPath(p string) string {
	p = os.ExpandEnv(strings.TrimSpace(p))
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
