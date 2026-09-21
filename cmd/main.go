package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nnavales/dropzone/internal/config"
	"github.com/nnavales/dropzone/internal/daemon"
	"github.com/nnavales/dropzone/internal/logger"
)

func main() {
	log := logger.New(os.Stdout, logger.FormatText)
	slog.SetDefault(log)

	log.Info("--->>> dropzone <<<---")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Debug("loading config", "path", config.ConfigPath())
	cfg, err := config.Load()
	if err != nil {
		log.Error("load config failed", "err", err)
		os.Exit(1)
	}
	log.Info("config loaded",
		"wait_seconds", cfg.Settings.WaitSeconds,
		"conflict", cfg.Settings.Conflict,
		"zones", len(cfg.Zones),
	)
	log.Debug("config dump", "config", cfg)
	for zi, z := range cfg.Zones {
		log.Info("zone configured",
			"index", zi,
			"name", z.Name,
			"path", z.Path,
			"rules", len(z.Rules),
		)
		for ri, r := range z.Rules {
			log.Debug("rule configured",
				"zone", z.Name,
				"index", ri,
				"name", r.Name,
				"match", r.Match,
				"action", r.Action,
			)
		}
	}

	log.Debug("building daemon")
	d, err := daemon.New(cfg)
	if err != nil {
		log.Error("build daemon failed", "err", err)
		os.Exit(1)
	}
	log.Info("daemon built", "daemon", d)

	log.Info("starting daemon, watching for stable files")
	if err := d.Start(ctx); err != nil && err != context.Canceled {
		log.Error("daemon stopped with error", "err", err)
		os.Exit(1)
	}
	log.Info("daemon stopped, bye")
}
