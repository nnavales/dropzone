package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nnavales/dropzone/internal/daemon"
	"github.com/nnavales/dropzone/internal/service"
)

func newServiceCommand() *cobra.Command {
	svc := &cobra.Command{
		Use:   "service",
		Short: "Manage dropzone as a systemd user service.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if dev {
				return fmt.Errorf("service commands don't support --dev: run without it")
			}
			return nil
		},
	}

	svc.AddCommand(
		&cobra.Command{
			Use:   "install",
			Short: "Install and start the systemd user service.",
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := checkInstallPrereqs(); err != nil {
					return err
				}
				path, err := service.Install()
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "service installed: %s\n", path)
				return nil
			},
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "Stop and remove the systemd user service.",
			RunE: func(cmd *cobra.Command, args []string) error {
				path, err := service.Uninstall()
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "service uninstalled: %s\n", path)
				return nil
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show the systemd user service status.",
			RunE: func(cmd *cobra.Command, args []string) error {
				st, err := service.Inspect()
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				fmt.Fprintf(out, "unit:      %s\n", st.UnitPath)
				fmt.Fprintf(out, "binary:    %s\n", st.Binary)
				fmt.Fprintf(out, "installed: %t\n", st.Installed)
				fmt.Fprintf(out, "enabled:   %s\n", st.Enabled)
				fmt.Fprintf(out, "active:    %s\n", st.Active)
				return nil
			},
		},
		newServiceLogsCommand(),
	)

	return svc
}

func checkInstallPrereqs() error {
	cfgPath := resolveConfigPath(false)
	if _, err := os.Stat(cfgPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config not found: run `dropzone init` first: %s", cfgPath)
		}
		return err
	}
	st, err := service.Inspect()
	if err != nil {
		return err
	}

	// We check that there is no foreground instance running before installing
	if st.Active != "active" {
		if err := daemon.LockFree(resolveLockPath(false)); err != nil {
			return fmt.Errorf("%w: stop the foreground instance before installing", err)
		}
	}
	return nil
}

func newServiceLogsCommand() *cobra.Command {
	var lines int
	var follow bool

	logs := &cobra.Command{
		Use:   "logs",
		Short: "Show service logs from the journal.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return service.Logs(lines, follow)
		},
	}
	logs.Flags().IntVarP(&lines, "lines", "n", 50, "number of lines to show")
	logs.Flags().BoolVarP(&follow, "follow", "f", false, "follow new log output")

	return logs
}
