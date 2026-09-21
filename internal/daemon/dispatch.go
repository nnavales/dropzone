package daemon

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/nnavales/dropzone/internal/rules"
)

// handleFile routes a stable path to its zone and first matching rule.
func (d *Daemon) handleFile(ctx context.Context, path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		return
	}

	zone := d.zoneFor(path)
	if zone == nil {
		return
	}

	file := rules.NewFile(path)

	for _, r := range zone.rules {
		if !r.Matches(file) {
			continue
		}
		if err := r.Action.Execute(ctx, path); err != nil {
			slog.Error("execute action", "zone", zone.name, "rule", r.Name, "path", path, "err", err)
		} else {
			slog.Info("rule applied", "zone", zone.name, "rule", r.Name, "path", path)
		}
		return
	}
}

func (d *Daemon) zoneFor(path string) *Zone {
	for i := range d.zones {
		z := &d.zones[i]
		if path == z.path || strings.HasPrefix(path, z.path+string(os.PathSeparator)) {
			return z
		}
	}
	return nil
}
