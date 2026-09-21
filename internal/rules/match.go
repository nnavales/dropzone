package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nnavales/dropzone/internal/config"
)

// Match is a runtime match.
type Match struct {
	Glob       []string
	Extensions []string
}

// NewMatch translates a config match into a runtime Match.
func NewMatch(match config.Match) (Match, error) {
	if len(match.Glob) == 0 && len(match.Extensions) == 0 {
		return Match{}, fmt.Errorf("match must have at least one of extensions|glob")
	}
	for _, g := range match.Glob {
		if err := validateGlob(g); err != nil {
			return Match{}, err
		}
	}
	for _, e := range match.Extensions {
		if err := validateExtension(e); err != nil {
			return Match{}, err
		}
	}

	return Match{
		Glob:       match.Glob,
		Extensions: match.Extensions,
	}, nil
}

func (m Match) match(file File) bool {
	if len(m.Glob) > 0 && !matchGlob(file.Name, m.Glob) {
		return false
	}

	if len(m.Extensions) > 0 && !matchExtension(file.Name, m.Extensions) {
		return false
	}

	return true
}

func matchGlob(name string, globs []string) bool {
	for _, pattern := range globs {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			continue
		}

		if matched {
			return true
		}
	}

	return false
}

func matchExtension(name string, extensions []string) bool {
	ext := filepath.Ext(name)
	ext = strings.ToLower(ext)

	for _, targetExt := range extensions {
		target := strings.ToLower(targetExt)
		if !strings.HasPrefix(target, ".") {
			target = "." + target
		}

		if ext == target {
			return true
		}
	}

	return false

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
