// Package engine routes stable file events to matching rules.
package engine

import (
	"context"
	"time"

	"github.com/nnavales/dropzone/internal/config"
	"github.com/nnavales/dropzone/internal/watcher"
)

// Engine orchestrates watching + execution directly over config.
type Engine struct {
	cfg       config.Config
	stableFor time.Duration
}

// New returns an Engine for cfg.
func New(cfg config.Config) *Engine {
	return &Engine{
		cfg:       cfg,
		stableFor: time.Duration(cfg.Settings.StableForSeconds) * time.Second,
	}
}

// Run watches configured paths and processes stable files.
func (e *Engine) Run(ctx context.Context) error {
	paths := make([]string, 0, len(e.cfg.Zones))
	for _, z := range e.cfg.Zones {
		paths = append(paths, z.Path)
	}

	w, err := watcher.New(e.stableFor, paths...)
	if err != nil {
		return err
	}

	go w.Run(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil

		case path, ok := <-w.Events():
			if !ok {
				return nil
			}
			e.handleFile(ctx, path)
		}
	}
}
