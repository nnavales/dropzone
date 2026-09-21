package actions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRename(t *testing.T) {
	ctx := context.Background()

	t.Run("renames in same directory", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "hello")

		if err := (Rename{Name: "b.txt"}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Errorf("source still exists after rename")
		}
		if got := readFile(t, filepath.Join(dir, "b.txt")); got != "hello" {
			t.Errorf("dst content = %q, want %q", got, "hello")
		}
	})

	t.Run("skip keeps both files", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		writeTempFile(t, dir, "b.txt", "old")

		if err := (Rename{Name: "b.txt", Conflict: ConflictSkip}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, src); got != "new" {
			t.Errorf("src content = %q, want %q", got, "new")
		}
		if got := readFile(t, filepath.Join(dir, "b.txt")); got != "old" {
			t.Errorf("dst content = %q, want %q", got, "old")
		}
	})

	t.Run("overwrite replaces destination", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		writeTempFile(t, dir, "b.txt", "old")

		if err := (Rename{Name: "b.txt", Conflict: ConflictOverwrite}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, filepath.Join(dir, "b.txt")); got != "new" {
			t.Errorf("dst content = %q, want %q", got, "new")
		}
	})

	t.Run("rename picks suffixed name", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		writeTempFile(t, dir, "b.txt", "old")

		if err := (Rename{Name: "b.txt", Conflict: ConflictRename}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, filepath.Join(dir, "b.txt")); got != "old" {
			t.Errorf("dst content = %q, want %q", got, "old")
		}
		if got := readFile(t, filepath.Join(dir, "b-1.txt")); got != "new" {
			t.Errorf("renamed content = %q, want %q", got, "new")
		}
	})

	t.Run("unknown policy errors", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		writeTempFile(t, dir, "b.txt", "old")

		if err := (Rename{Name: "b.txt", Conflict: "bogus"}).Execute(ctx, src); err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
	})
}
