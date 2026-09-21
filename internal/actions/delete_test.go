package actions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("removes file", func(t *testing.T) {
		path := writeTempFile(t, t.TempDir(), "a.txt", "hello")

		if err := (Delete{}).Execute(ctx, path); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file still exists after delete")
		}
	})

	t.Run("missing file errors", func(t *testing.T) {
		err := (Delete{}).Execute(ctx, filepath.Join(t.TempDir(), "nope.txt"))
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
	})
}
