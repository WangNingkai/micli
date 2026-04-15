package cmd

import (
	"github.com/spf13/cobra"
)

var consumablesCmd = &cobra.Command{
	Use:   "consumables",
	Short: "List consumable items (filters, batteries, etc.)",
	Long:  `List consumable items across your homes. Use --home to filter by home name.`,
	Run: func(cmd *cobra.Command, args []string) {
		res, err := ioSrv.GetConsumables(consumablesHome)
		handleResult(res, err)
	},
}

var consumablesHome string

func init() {
	consumablesCmd.Flags().StringVarP(&consumablesHome, "home", "H", "", "Filter by home name")
	consumablesCmd.Example = "  consumables\n  consumables --home \"My Home\""
}
