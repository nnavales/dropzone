package watcher

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

const configStableFor = 2 * time.Second

// ConfigWatcher monitors a single config file for stable changes.
type ConfigWatcher struct {
	watcher *fsnotify.Watcher
	events  chan string
	path    string
}

// NewConfigWatcher creates a ConfigWatcher for the given path.
func NewConfigWatcher(path string) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(filepath.Dir(path)); err != nil {
		watcher.Close()
		return nil, err
	}

	return &ConfigWatcher{
		watcher: watcher,
		path:    path,
		events:  make(chan string, 1),
	}, nil
}

// Run starts the ConfigWatcher process.
func (w *ConfigWatcher) Run(ctx context.Context) {
	defer w.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var (
		pending     bool
		lastSize    int64
		lastMod     time.Time
		stableSince time.Time
	)

	for {
		select {
		case <-ctx.Done():
			return

		case event := <-w.watcher.Events:
			if event.Name == w.path {
				pending = true
			}

		case err := <-w.watcher.Errors:
			slog.Warn("watch config error", "err", err)

		case <-ticker.C:
			if !pending {
				continue
			}

			changed, size, mod := w.checkStable(lastSize, lastMod)
			lastSize = size
			lastMod = mod

			if changed {
				stableSince = time.Now()
				continue
			}

			if time.Since(stableSince) < configStableFor {
				continue
			}

			pending = false

			select {
			case w.events <- w.path:
			default:
			}
		}
	}
}

func (w *ConfigWatcher) checkStable(lastSize int64, lastMod time.Time) (bool, int64, time.Time) {
	info, err := os.Stat(w.path)
	if err != nil {
		return false, lastSize, lastMod
	}

	if info.Size() != lastSize || !info.ModTime().Equal(lastMod) {
		return true, info.Size(), info.ModTime()
	}

	return false, lastSize, lastMod
}

// Events returns a channel that emits the config path when stable.
func (w *ConfigWatcher) Events() <-chan string {
	return w.events
}

// Close stops the ConfigWatcher and releases resources.
func (w *ConfigWatcher) Close() {
	close(w.events)
	w.watcher.Close()
}
