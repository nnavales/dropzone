package actions

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ConflictPolicy defines how to handle file name conflicts.
type ConflictPolicy string

const (
	// ConflictSkip ignores the operation when a conflict exists.
	ConflictSkip ConflictPolicy = "skip"
	// ConflictOverwrite replaces the existing file.
	ConflictOverwrite ConflictPolicy = "overwrite"
	// ConflictRename creates a new file with a numeric suffix.
	ConflictRename ConflictPolicy = "rename"
)

func hasConflict(dst string) (bool, error) {
	_, err := os.Stat(dst)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func renameDestination(dst string) (string, error) {
	ext := filepath.Ext(dst)
	base := dst[:len(dst)-len(ext)]

	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d%s", base, i, ext)

		conflict, err := hasConflict(candidate)
		if err != nil {
			return "", err
		}

		if !conflict {
			return candidate, nil
		}
	}
}
