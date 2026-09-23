package engine

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/nnavales/dropzone/internal/actions"
	"github.com/nnavales/dropzone/internal/config"
	"github.com/nnavales/dropzone/internal/match"
)

// handleFile routes a stable path to its zone and first matching rule.
func (e *Engine) handleFile(ctx context.Context, path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		return
	}

	zone := e.zoneFor(path)
	if zone == nil {
		return
	}

	for _, r := range zone.Rules {
		if !match.Matches(path, r.Match.Glob, r.Match.Extensions) {
			continue
		}
		// Engine resolves placeholders per event; actions receive clean values.
		target := r.Action.Target
		if target != "" {
			target = resolveAction(target, path, time.Now())
		}
		conflict := r.OnConflict
		act, err := actions.New(r.Action.Kind, target, conflict)
		if err != nil {
			slog.Error("build action", "zone", zone.Name, "rule", r.Name, "path", path, "err", err)
			return
		}
		if err := act.Execute(ctx, path); err != nil {
			slog.Error("execute action", "zone", zone.Name, "rule", r.Name, "path", path, "err", err)
		} else {
			slog.Info("rule applied", "zone", zone.Name, "rule", r.Name, "path", path)
		}
		return
	}
}

func (e *Engine) zoneFor(path string) *config.Zone {
	for i := range e.cfg.Zones {
		z := &e.cfg.Zones[i]
		if path == z.Path || strings.HasPrefix(path, z.Path+string(os.PathSeparator)) {
			return z
		}
	}
	return nil
}
