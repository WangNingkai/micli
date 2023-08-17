package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"micli/internal/conf"
)

var (
	resetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Config Reset",
		Long:  `Config Reset`,
		Run: func(cmd *cobra.Command, args []string) {
			confirm, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure to reset config file?")
			if confirm {
				conf.Reset()
				pterm.Info.Println("Config file has been reset.")
				return
			}
		},
	}
)
