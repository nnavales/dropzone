package actions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestCopy(t *testing.T) {
	ctx := context.Background()

	t.Run("copies content and keeps source", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "hello")
		dstDir := t.TempDir()

		if err := (Copy{Destination: dstDir}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, filepath.Join(dstDir, "a.txt")); got != "hello" {
			t.Errorf("dst content = %q, want %q", got, "hello")
		}
		if got := readFile(t, src); got != "hello" {
			t.Errorf("src content = %q, want %q (source must be preserved)", got, "hello")
		}
	})

	t.Run("skip keeps existing destination", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		dstDir := t.TempDir()
		writeTempFile(t, dstDir, "a.txt", "old")

		if err := (Copy{Destination: dstDir, Conflict: "skip"}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, filepath.Join(dstDir, "a.txt")); got != "old" {
			t.Errorf("dst content = %q, want %q", got, "old")
		}
	})

	t.Run("overwrite replaces destination and keeps source", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		dstDir := t.TempDir()
		writeTempFile(t, dstDir, "a.txt", "old")

		if err := (Copy{Destination: dstDir, Conflict: "overwrite"}).Execute(ctx, src); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, filepath.Join(dstDir, "a.txt")); got != "new" {
			t.Errorf("dst content = %q, want %q", got, "new")
		}
		if got := readFile(t, src); got != "new" {
			t.Errorf("src content = %q, want %q (source must be preserved)", got, "new")
		}
	})

	t.Run("rename writes to suffixed file", func(t *testing.T) {
		dir := t.TempDir()
		src := writeTempFile(t, dir, "a.txt", "new")
		dstDir := t.TempDir()
		writeTempFile(t, dstDir, "a.txt", "old")

		if err := (Copy{Destination: dstDir, Conflict: "rename"}).Execute(ctx, src); err != nil {
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

		err := (Copy{Destination: dstDir, Conflict: "bogus"}).Execute(ctx, src)
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
	})

	t.Run("missing source errors", func(t *testing.T) {
		err := (Copy{Destination: t.TempDir()}).Execute(ctx, filepath.Join(t.TempDir(), "nope.txt"))
		if err == nil {
			t.Fatal("Execute() expected error, got nil")
		}
	})
}
