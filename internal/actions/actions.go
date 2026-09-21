// Package actions defines file operations.
package actions

import (
	"context"
	"fmt"

	"github.com/nnavales/dropzone/internal/config"
)

// Action represents a file operation to execute.
// source is the triggering file path, supplied per event.
type Action interface {
	Execute(ctx context.Context, source string) error
}

// New translates a config action into a runtime Action.
func New(action config.Action, conflict ConflictPolicy) (Action, error) {
	switch conflict {
	case ConflictSkip, ConflictOverwrite, ConflictRename:
	default:
		return nil, fmt.Errorf("unknown conflict policy: %q", conflict)
	}

	switch action.Kind {
	case config.ActionMove:
		if action.Target == "" {
			return nil, fmt.Errorf("move requires a non-empty target")
		}
		return Move{Destination: action.Target, Conflict: conflict}, nil
	case config.ActionCopy:
		if action.Target == "" {
			return nil, fmt.Errorf("copy requires a non-empty target")
		}
		return Copy{Destination: action.Target, Conflict: conflict}, nil
	case config.ActionRename:
		if action.Target == "" {
			return nil, fmt.Errorf("rename requires a non-empty target")
		}
		return Rename{Name: action.Target, Conflict: conflict}, nil
	case config.ActionRun:
		if action.Target == "" {
			return nil, fmt.Errorf("run requires a non-empty target")
		}
		return Run{Command: action.Target}, nil
	case config.ActionDelete:
		return Delete{}, nil
	default:
		return nil, fmt.Errorf("one of move|copy|rename|run|delete is required")
	}
}
