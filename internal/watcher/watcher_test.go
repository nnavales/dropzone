package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestNew_InvalidPaths(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := map[string][]string{
		"relative path":      {"relative/path"},
		"missing path":       {filepath.Join(dir, "nope")},
		"file not directory": {file},
	}
	for name, paths := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := New(time.Second, paths...); err == nil {
				t.Fatal("New() expected error, got nil")
			}
		})
	}

	t.Run("non-positive stableFor", func(t *testing.T) {
		for _, d := range []time.Duration{0, -time.Second} {
			if _, err := New(d, dir); err == nil {
				t.Errorf("New(%v) expected error, got nil", d)
			}
		}
	})
}

func TestHandleEvent(t *testing.T) {
	newWatcher := func() *Watcher {
		return &Watcher{pending: make(map[string]*pendingFile)}
	}

	t.Run("create tracks pending", func(t *testing.T) {
		w := newWatcher()
		w.handleEvent(fsnotify.Event{Name: "/tmp/a.txt", Op: fsnotify.Create})
		if _, ok := w.pending["/tmp/a.txt"]; !ok {
			t.Error("Create should add path to pending")
		}
	})

	t.Run("rename tracks pending", func(t *testing.T) {
		w := newWatcher()
		w.handleEvent(fsnotify.Event{Name: "/tmp/a.txt", Op: fsnotify.Rename})
		if _, ok := w.pending["/tmp/a.txt"]; !ok {
			t.Error("Rename should add path to pending")
		}
	})

	t.Run("write only refreshes tracked path", func(t *testing.T) {
		w := newWatcher()
		w.handleEvent(fsnotify.Event{Name: "/tmp/untracked.txt", Op: fsnotify.Write})
		if _, ok := w.pending["/tmp/untracked.txt"]; ok {
			t.Error("Write on untracked path should not add to pending")
		}

		w.pending["/tmp/tracked.txt"] = &pendingFile{size: 1}
		w.handleEvent(fsnotify.Event{Name: "/tmp/tracked.txt", Op: fsnotify.Write})
		if got := w.pending["/tmp/tracked.txt"]; got == nil || got.size != 0 {
			t.Error("Write on tracked path should reset pending entry")
		}
	})

	t.Run("other ops ignored", func(t *testing.T) {
		w := newWatcher()
		for _, op := range []fsnotify.Op{fsnotify.Remove, fsnotify.Chmod} {
			w.handleEvent(fsnotify.Event{Name: "/tmp/a.txt", Op: op})
		}
		if len(w.pending) != 0 {
			t.Errorf("pending = %v, want empty", w.pending)
		}
	})
}

func TestCheckPending(t *testing.T) {
	t.Run("missing file is dropped", func(t *testing.T) {
		w := &Watcher{
			events:  make(chan string, 1),
			pending: map[string]*pendingFile{"/nonexistent/file.txt": {}},
		}
		w.checkPending(time.Now())
		if len(w.pending) != 0 {
			t.Error("missing file should be removed from pending")
		}
	})

	t.Run("changed file updates snapshot without emitting", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a.txt")
		if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
			t.Fatal(err)
		}
		now := time.Now()
		w := &Watcher{
			events:    make(chan string, 1),
			pending:   map[string]*pendingFile{path: {}},
			stableFor: time.Second,
		}
		w.checkPending(now)
		p := w.pending[path]
		if p == nil {
			t.Fatal("changed file should stay pending")
		}
		if p.size != 5 || !p.stableSince.Equal(now) {
			t.Errorf("pending = %+v, want size 5 and stableSince %v", p, now)
		}
		select {
		case e := <-w.events:
			t.Errorf("unexpected event %q", e)
		default:
		}
	})

	t.Run("stable file emits and is removed", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a.txt")
		if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		w := &Watcher{
			events:  make(chan string, 1),
			pending: map[string]*pendingFile{path: {size: info.Size(), modTime: info.ModTime(), stableSince: time.Now().Add(-time.Hour)}},
			stableFor: time.Second,
		}
		w.checkPending(time.Now())
		select {
		case got := <-w.events:
			if got != path {
				t.Errorf("event = %q, want %q", got, path)
			}
		default:
			t.Fatal("expected event for stable file, got none")
		}
		if len(w.pending) != 0 {
			t.Error("emitted file should be removed from pending")
		}
	})

	t.Run("not yet stable does not emit", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a.txt")
		if err := os.WriteFile(path, []byte("hi"), 0o644); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		now := time.Now()
		w := &Watcher{
			events:  make(chan string, 1),
			pending: map[string]*pendingFile{path: {size: info.Size(), modTime: info.ModTime(), stableSince: now}},
			stableFor: time.Hour,
		}
		w.checkPending(now.Add(time.Second))
		select {
		case e := <-w.events:
			t.Errorf("unexpected event %q", e)
		default:
		}
		if _, ok := w.pending[path]; !ok {
			t.Error("unstable file should stay pending")
		}
	})

	t.Run("full events channel does not block", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a.txt")
		if err := os.WriteFile(path, []byte("hi"), 0o644); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		events := make(chan string, 1)
		events <- "already-queued"
		w := &Watcher{
			events:  events,
			pending: map[string]*pendingFile{path: {size: info.Size(), modTime: info.ModTime(), stableSince: time.Now().Add(-time.Hour)}},
			stableFor: time.Second,
		}
		done := make(chan struct{})
		go func() { w.checkPending(time.Now()); close(done) }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("checkPending blocked on full events channel")
		}
		if _, ok := w.pending[path]; !ok {
			t.Error("file should stay pending when event could not be queued")
		}
	})
}

func TestRun_EmitsStableFile(t *testing.T) {
	dir := t.TempDir()
	w, err := New(100*time.Millisecond, dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	path := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-w.Events():
		if got != path {
			t.Errorf("event = %q, want %q", got, path)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for stable file event")
	}
}
