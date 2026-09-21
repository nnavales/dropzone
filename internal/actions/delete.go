package actions

import (
	"context"
	"os"
)

// Delete removes a file from disk.
type Delete struct{}

// Execute performs the delete operation.
func (d Delete) Execute(ctx context.Context, source string) error {
	return os.Remove(source)
}
