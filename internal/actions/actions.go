// Package actions defines file operations.
package actions

import (
	"context"
	"fmt"
)

// Action is a file operation run against a triggering source path.
type Action interface {
	Execute(ctx context.Context, source string) error
}

// New returns the Action for kind, target and conflict policy.
func New(kind, target string, conflict string) (Action, error) {
	switch kind {
	case "move":
		return Move{Destination: target, Conflict: conflict}, nil
	case "copy":
		return Copy{Destination: target, Conflict: conflict}, nil
	case "rename":
		return Rename{Name: target, Conflict: conflict}, nil
	case "run":
		return Run{Command: target}, nil
	case "delete":
		return Delete{}, nil
	default:
		return nil, fmt.Errorf("one of move|copy|rename|run|delete is required")
	}
}
