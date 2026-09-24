package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/nnavales/dropzone/internal/config"
	"github.com/nnavales/dropzone/internal/daemon"
	"github.com/nnavales/dropzone/internal/logger"
)

func newRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "run",
		Short:        "Start the daemon.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context())
		},
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	level := slog.LevelInfo
	if dev {
		level = slog.LevelDebug
	}
	log := logger.New(os.Stdout, level, dev)
	slog.SetDefault(log)

	cfgPath := resolveConfigPath(dev)
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return err
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return fmt.Errorf("config not found: run `dropzone init` to create one: %s", cfgPath)
	}
	slog.Debug("config loaded", "config", cfg)

	lockPath := resolveLockPath(dev)

	d := daemon.New(cfgPath, cfg, lockPath)

	if err := d.Run(ctx); err != nil && err != context.Canceled {
		return err
	}
	log.Info("daemon stopped, bye")
	return nil
}

// resolveConfigPath returns the dev config path with --dev,
// or the production path otherwise (config.yml preferred, config.yaml accepted).
func resolveConfigPath(dev bool) string {
	if dev {
		return filepath.Join(".local", "cfg", "config.yml")
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	dir := filepath.Join(configDir, "dropzone")
	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); err == nil {
		return filepath.Join(dir, "config.yaml")
	}
	return filepath.Join(dir, "config.yml")
}

func resolveLockPath(dev bool) string {
	if dev {
		return filepath.Join(".local", "dropzone.lock")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".local", "state", "dropzone", "dropzone.lock")

}
