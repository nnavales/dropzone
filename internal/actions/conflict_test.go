package actions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasConflict(t *testing.T) {
	dir := t.TempDir()

	existing := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	conflict, err := hasConflict(existing)
	if err != nil || !conflict {
		t.Errorf("hasConflict(existing) = %v, %v; want true, nil", conflict, err)
	}

	conflict, err = hasConflict(filepath.Join(dir, "missing.txt"))
	if err != nil || conflict {
		t.Errorf("hasConflict(missing) = %v, %v; want false, nil", conflict, err)
	}
}

func TestRenameDestination(t *testing.T) {
	dir := t.TempDir()

	dst := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(dst, []byte("orig"), 0o644); err != nil {
		t.Fatal(err)
	}

	first, err := renameDestination(dst)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "file-1.txt"); first != want {
		t.Errorf("renameDestination() = %q, want %q", first, want)
	}

	// Occupy -1, next candidate must skip to -2.
	if err := os.WriteFile(first, []byte("taken"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := renameDestination(dst)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "file-2.txt"); second != want {
		t.Errorf("renameDestination() = %q, want %q", second, want)
	}
}
