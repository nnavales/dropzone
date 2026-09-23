// Package match defines runtime matching for files.
package match

import (
	"path/filepath"
	"strings"
)

// Matches reports whether path satisfies globs and extensions.
// Both apply when set (AND); callers guarantee at least one is non-empty.
func Matches(path string, globs []string, extensions []string) bool {
	basename := filepath.Base(path)
	if len(globs) > 0 && !matchGlob(basename, globs) {
		return false
	}

	if len(extensions) > 0 && !matchExtension(basename, extensions) {
		return false
	}

	return true
}

func matchGlob(basename string, globs []string) bool {
	for _, pattern := range globs {
		matched, err := filepath.Match(pattern, basename)
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
