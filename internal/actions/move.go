package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Move moves a file to a destination directory.
type Move struct {
	Destination string
	Conflict    string
}

// Execute moves source into Destination.
func (m Move) Execute(ctx context.Context, source string) error {
	dst := filepath.Join(m.Destination, filepath.Base(source))

	conflict, err := hasConflict(dst)
	if err != nil {
		return err
	}

	if conflict {
		switch m.Conflict {
		case "skip":
			return nil

		case "overwrite":
			// Nothing to do: os.Rename overwrites.

		case "rename":
			dst, err = renameDestination(dst)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown conflict policy: %q", m.Conflict)
		}

	}
	return os.Rename(source, dst)
}
