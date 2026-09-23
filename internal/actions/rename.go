package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Rename renames a file to a new name within the same directory.
type Rename struct {
	Name     string
	Conflict string
}

// Execute renames source to Name, which the engine resolves.
func (r Rename) Execute(ctx context.Context, source string) error {
	dst := filepath.Join(filepath.Dir(source), r.Name)

	conflict, err := hasConflict(dst)
	if err != nil {
		return err
	}

	if conflict {
		switch r.Conflict {
		case "skip":
			return nil

		case "overwrite":

		case "rename":
			dst, err = renameDestination(dst)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown conflict policy: %q", r.Conflict)
		}
	}

	return os.Rename(source, dst)
}
