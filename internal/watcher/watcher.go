// Package watcher monitors filesystem changes and emits stable events.
package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches directories and produces events.
type Watcher struct {
	watcher   *fsnotify.Watcher
	events    chan string
	pending   map[string]*pendingFile
	stableFor time.Duration
}

type pendingFile struct {
	path        string
	size        int64
	modTime     time.Time
	stableSince time.Time
}

// New creates a Watcher over the given directories.
func New(stableFor time.Duration, paths ...string) (*Watcher, error) {
	if stableFor <= 0 {
		return nil, fmt.Errorf("stableFor must be greater than 0")
	}

	fw := &Watcher{
		events:    make(chan string, 1),
		pending:   make(map[string]*pendingFile),
		stableFor: stableFor,
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw.watcher = w

	if err := validatePaths(paths...); err != nil {
		return nil, err
	}

	for _, path := range paths {
		if err := w.Add(path); err != nil {
			w.Close()
			return nil, err
		}
	}

	return fw, nil
}

// Events returns a channel that emits stable file paths.
func (w *Watcher) Events() <-chan string {
	return w.events
}

// Close cleans-up the watcher.
func (w *Watcher) Close() {
	w.watcher.Close()
	close(w.events)

}

// Run starts the Watcher process over the directories.
func (w *Watcher) Run(ctx context.Context) {
	defer w.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case event := <-w.watcher.Events:
			w.handleEvent(event)

		case <-ticker.C:
			w.checkPending(time.Now())
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Create:
		w.pending[event.Name] = &pendingFile{}
		return
	case fsnotify.Rename:
		w.pending[event.Name] = &pendingFile{}
		return
	case fsnotify.Write:
		if _, ok := w.pending[event.Name]; ok {
			w.pending[event.Name] = &pendingFile{}
		}
		return
	default:
		return
	}
}

func (w *Watcher) checkPending(now time.Time) {
	for path, pending := range w.pending {
		info, err := os.Stat(path)
		if err != nil {
			delete(w.pending, path)
			continue
		}

		size := info.Size()
		modTime := info.ModTime()

		if size != pending.size || !modTime.Equal(pending.modTime) {
			pending.size = size
			pending.modTime = modTime
			pending.stableSince = now
			continue
		}

		if now.Sub(pending.stableSince) >= w.stableFor {
			select {
			case w.events <- path:
				delete(w.pending, path)
			default:
			}
		}
	}
}

func validatePaths(paths ...string) error {
	var errs []string

	for _, p := range paths {
		if !filepath.IsAbs(p) {
			errs = append(errs, fmt.Sprintf("watch path must be absolute: %s", p))
			continue
		}

		info, err := os.Stat(p)
		if err != nil {
			errs = append(errs, fmt.Sprintf("unable to get stats from %s, %s", p, err))
			continue
		}

		if !info.IsDir() {
			errs = append(errs, fmt.Sprintf("watch path must be a directory: %s", p))
			continue
		}
	}

	if len(errs) > 0 {
		errs := strings.Join(errs, "\n")
		return errors.New(errs)
	}

	return nil
}
