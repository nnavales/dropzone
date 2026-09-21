package actions

import (
	"context"
	"os/exec"
)

// Run executes a shell command.
type Run struct {
	Command string
}

// Execute performs the command execution.
func (r Run) Execute(ctx context.Context, source string) error {
	return exec.CommandContext(ctx, "sh", "-c", renderCommand(r.Command, source)).Run()
}
