package actions

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
