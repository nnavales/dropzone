package actions

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Copy copies a file to a destination directory.
type Copy struct {
	Destination string
	Conflict    ConflictPolicy
}

// Execute performs the copy operation.
func (c Copy) Execute(ctx context.Context, source string) error {
	dst := filepath.Join(c.Destination, filepath.Base(source))

	conflict, err := hasConflict(dst)
	if err != nil {
		return err
	}

	if conflict {
		switch c.Conflict {
		case ConflictSkip:
			return nil

		case ConflictOverwrite:

		case ConflictRename:
			dst, err = renameDestination(dst)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown conflict policy: %q", c.Conflict)
		}
	}

	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
