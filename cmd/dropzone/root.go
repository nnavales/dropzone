package main

import (
	"github.com/spf13/cobra"
)

var version = "0.1.0"
var dev bool

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:          "dropzone",
		Short:        "Watch directories and act on stable files.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context())
		},
	}
	root.PersistentFlags().BoolVar(&dev, "dev", false, "run with local dev config and source logging")
	_ = root.PersistentFlags().MarkHidden("dev")

	root.AddCommand(newRunCommand(), newInitCommand(), newValidateCommand(), newVersionCommand(), newServiceCommand())

	return root
}
