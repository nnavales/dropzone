package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nnavales/dropzone/internal/config"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "init",
		Short:        "Scaffold the default config file.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(resolveConfigPath(dev))
		},
	}
}

// runInit scaffolds the default config file. Never overwrites.
func runInit(cfgPath string) error {
	if cfgPath == "" {
		return fmt.Errorf("empty config path")
	}

	if _, err := os.Stat(cfgPath); err == nil {
		return fmt.Errorf("config already exists, not overwriting: %s", cfgPath)
	}

	data, err := config.Marshal(config.Default())
	if err != nil {
		return err
	}

	header := "# dropzone config\n# Full reference: https://github.com/nnavales/dropzone/blob/master/docs/configuration.md\n"
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(cfgPath, append([]byte(header), data...), 0644); err != nil {
		return err
	}

	slog.Info("config created", "path", cfgPath)
	return nil
}
