// Package daemon supervises an engine over config.
package daemon

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/nnavales/dropzone/internal/config"
	"github.com/nnavales/dropzone/internal/engine"
	"github.com/nnavales/dropzone/internal/watcher"
)

// Daemon supervises one engine instance at a time.
type Daemon struct {
	cfgPath string
	cfg     config.Config
}

// New returns a Daemon.
func New(cfgPath string, cfg config.Config) *Daemon {
	return &Daemon{cfgPath: cfgPath, cfg: cfg}
}

// Run starts the daemon and manages its dependencies.
func (d *Daemon) Run(ctx context.Context) error {
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
		}
	}
}
