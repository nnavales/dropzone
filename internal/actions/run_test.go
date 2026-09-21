package actions

import (
	"context"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Run("succeeds on exit zero", func(t *testing.T) {
		if err := (Run{Command: "true"}).Execute(context.Background(), ""); err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	})

	t.Run("fails on nonzero exit", func(t *testing.T) {
		if err := (Run{Command: "false"}).Execute(context.Background(), ""); err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		if err := (Run{Command: "sleep 5"}).Execute(ctx, ""); err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
	})
}
