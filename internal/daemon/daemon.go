package daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/nnavales/dropzone/internal/actions"
	"github.com/nnavales/dropzone/internal/config"
	"github.com/nnavales/dropzone/internal/watcher"
)

// Daemon holds runtime zones and orchestrates watching + execution.
type Daemon struct {
	zones     []Zone
	stableFor time.Duration
}

// New builds runtime zones from config and returns a Daemon.
func New(cfg config.Config) (*Daemon, error) {
	conflict := actions.ConflictPolicy(cfg.Settings.Conflict)

	zones := make([]Zone, 0, len(cfg.Zones))
	for i := range cfg.Zones {
		z, err := newZone(cfg.Zones[i], conflict)
		if err != nil {
			return nil, fmt.Errorf("zones[%d] (%q): %w", i, cfg.Zones[i].Name, err)
		}
		zones = append(zones, z)
	}

	return &Daemon{
		zones:     zones,
		stableFor: time.Duration(cfg.Settings.WaitSeconds) * time.Second,
	}, nil
}

// Start creates one watcher for all zone paths and processes stable files.
func (d *Daemon) Start(ctx context.Context) error {
	paths := make([]string, 0, len(d.zones))
	for _, z := range d.zones {
		paths = append(paths, z.path)
	}

	w, err := watcher.New(d.stableFor, paths...)
	if err != nil {
		return err
	}

	go w.Run(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case path, ok := <-w.Events():
			if !ok {
				return nil
			}
			d.handleFile(ctx, path)
		}
	}
}

