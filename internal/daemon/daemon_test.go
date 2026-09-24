package daemon

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestLockFree(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dropzone.lock")

	if err := LockFree(path); err != nil {
		t.Fatalf("free lock should succeed: %v", err)
	}

	held, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	if err := syscall.Flock(int(held.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}

	if err := LockFree(path); err == nil {
		t.Fatal("held lock should fail")
	}
}
