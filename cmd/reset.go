package cmd

import (
	"micli/internal/conf"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
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
