package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nnavales/dropzone/internal/config"
)

func newValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "validate",
		Short:        "Validate the config file without running.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath := resolveConfigPath(dev)
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			rules := 0
			for _, z := range cfg.Zones {
				rules += len(z.Rules)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "config valid: %s (%d zones, %d rules)\n",
				cfgPath, len(cfg.Zones), rules)
			return nil
		},
	}
}
