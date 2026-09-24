// Package daemon supervises an engine over config.
package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"syscall"

	"github.com/nnavales/dropzone/internal/config"
	"github.com/nnavales/dropzone/internal/engine"
	"github.com/nnavales/dropzone/internal/watcher"
)

// Daemon supervises one engine instance at a time.
type Daemon struct {
	lockPath string
	cfgPath  string
	cfg      config.Config
}

// New returns a Daemon.
func New(cfgPath string, cfg config.Config, lockPath string) *Daemon {
	return &Daemon{cfgPath: cfgPath, cfg: cfg, lockPath: lockPath}
}

// Run starts the daemon and manages its dependencies.
func (d *Daemon) Run(ctx context.Context) error {
	lock, err := acquireLock(d.lockPath)
	if err != nil {
		return err
	}
	defer lock.Close()

	slog.Info("starting daemon, watching for stable files")

	cfgWatcher, err := watcher.NewConfigWatcher(d.cfgPath)
	if err != nil {
		return err
	}
	go cfgWatcher.Run(ctx)

	current := engine.New(d.cfg)

	for {
		engineCtx, cancel := context.WithCancel(ctx)
		engineDone := make(chan error, 1)

		go func() {
			engineDone <- current.Run(engineCtx)
		}()

		select {
		case <-ctx.Done():
			cancel()
			<-engineDone
			return nil

		case err := <-engineDone:
			cancel()
			return err

		case <-cfgWatcher.Events():
			cancel()
			<-engineDone

			next, err := config.Load(d.cfgPath)
			if err != nil {
				slog.Error("failed to reload config", "err", err)
				continue
			}

			if reflect.DeepEqual(next, d.cfg) {
				slog.Debug("config unchanged")
				continue
			}

			d.cfg = next
			current = engine.New(next)

			slog.Info("config reloaded")
			slog.Debug("config dump", "config", next)
		}
	}
}

func acquireLock(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		f.Close()

		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, fmt.Errorf("dropzone is already running")
		}

		return nil, err
	}

	return f, nil
}
