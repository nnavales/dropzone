package actions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMove(t *testing.T) {
	ctx := context.Background()

	t.Run("moves file", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "hello")
		dstDir := t.TempDir()

		if err := (Move{Destination: dstDir}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Errorf("source still exists after move")
		}
		if got := readFile(t, filepath.Join(dstDir, "a.txt")); got != "hello" {
			t.Errorf("dst content = %q, want %q", got, "hello")
		}
	})

	t.Run("skip keeps both files", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		dstDir := t.TempDir()
		writeTempFile(t, dstDir, "a.txt", "old")

		if err := (Move{Destination: dstDir, Conflict: "skip"}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, src); got != "new" {
			t.Errorf("src content = %q, want %q", got, "new")
		}
		if got := readFile(t, filepath.Join(dstDir, "a.txt")); got != "old" {
			t.Errorf("dst content = %q, want %q", got, "old")
		}
	})

	t.Run("overwrite replaces destination", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		dstDir := t.TempDir()
		writeTempFile(t, dstDir, "a.txt", "old")

		if err := (Move{Destination: dstDir, Conflict: "overwrite"}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, filepath.Join(dstDir, "a.txt")); got != "new" {
			t.Errorf("dst content = %q, want %q", got, "new")
		}
	})

	t.Run("rename moves to suffixed file", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		dstDir := t.TempDir()
		writeTempFile(t, dstDir, "a.txt", "old")

		if err := (Move{Destination: dstDir, Conflict: "rename"}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, filepath.Join(dstDir, "a.txt")); got != "old" {
			t.Errorf("dst content = %q, want %q", got, "old")
		}
		if got := readFile(t, filepath.Join(dstDir, "a-1.txt")); got != "new" {
			t.Errorf("renamed content = %q, want %q", got, "new")
		}
	})

	t.Run("unknown policy errors", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		dstDir := t.TempDir()
		writeTempFile(t, dstDir, "a.txt", "old")

		if err := (Move{Destination: dstDir, Conflict: "bogus"}).Execute(ctx, src); err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
	})
}
