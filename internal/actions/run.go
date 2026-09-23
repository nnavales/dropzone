package actions

import (
	"context"
	"os/exec"
)

// Run executes a shell command.
type Run struct {
	Command string
}

// Execute runs Command, which the engine resolves.
func (r Run) Execute(ctx context.Context, source string) error {
	return exec.CommandContext(ctx, "sh", "-c", r.Command).Run()
}
