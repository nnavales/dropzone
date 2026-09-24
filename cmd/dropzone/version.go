package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version.",
		Run: func(cmd *cobra.Command, args []string) {
			if dev {
				fmt.Println("dev")
				return
			}
			fmt.Println(version)
		},
	}
}
