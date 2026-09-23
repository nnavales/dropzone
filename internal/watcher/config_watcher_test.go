package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigWatcherCheckStable(t *testing.T) {
	newWatcher := func(t *testing.T, content string) (*ConfigWatcher, string) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		w, err := NewConfigWatcher(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(w.Close)
		return w, path
	}

	t.Run("missing file keeps last values", func(t *testing.T) {
		w, err := NewConfigWatcher(filepath.Join(t.TempDir(), "gone.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		defer w.Close()
		changed, size, _ := w.checkStable(42, time.Now())
		if changed || size != 42 {
			t.Errorf("checkStable() = (%v, %d, _), want (false, 42, _)", changed, size)
		}
	})

	t.Run("changed file reports change", func(t *testing.T) {
		w, path := newWatcher(t, "a: 1")
		changed, size, mod := w.checkStable(0, time.Time{})
		if !changed {
			t.Error("checkStable() changed = false, want true")
		}
		if size == 0 || mod.IsZero() {
			t.Error("checkStable() should return current size and modTime")
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		changed, _, _ = w.checkStable(info.Size(), info.ModTime())
		if changed {
			t.Error("checkStable() changed = true for identical snapshot, want false")
		}
	})
}
