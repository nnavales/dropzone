package daemon

import (
	"errors"
	"os"
)

// resolvePath checks path points to an existing directory.
// Tilde and env expansion already happened in config.Parse.
func resolvePath(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return path, errors.New("path has to point to a existent directory.")
	}

	if !info.IsDir() {
		return path, errors.New("path has to point to a directory, not a file.")
	}

	return path, nil
}
