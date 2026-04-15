/*
Copyright © 2025 wangningkai

MIT License
*/

package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var homesCmd = &cobra.Command{
	Use:   "homes",
	Short: "List all homes with details",
	Long:  `List all Xiaomi smart homes with their details including device count.`,
	Run:   runHomesCmd,
}

var homesJsonOutput bool

func init() {
	rootCmd.AddCommand(homesCmd)
	homesCmd.Flags().BoolVar(&homesJsonOutput, "json", false, "Output in JSON format")
}

type HomeDetail struct {
	HomeID      string `json:"home_id"`
	HomeName    string `json:"home_name"`
	UID         string `json:"uid"`
	DeviceCount int    `json:"device_count"`
}

func runHomesCmd(cmd *cobra.Command, args []string) {
	homes, err := ioSrv.HomeList()
	if err != nil {
		handleResult(nil, err)
		return
	}

	details := make([]HomeDetail, 0, len(homes))
	for _, home := range homes {
		dids, devErr := ioSrv.HomeDeviceList(home.HomeID, home.UID)
		deviceCount := 0
		if devErr == nil {
			deviceCount = len(dids)
		}

		details = append(details, HomeDetail{
			HomeID:      home.HomeID,
			HomeName:    home.HomeName,
			UID:         home.UID,
			DeviceCount: deviceCount,
		})
	}

	if homesJsonOutput {
		handleResult(details, nil)
	} else {
		printHomesTable(details)
	}
}

func printHomesTable(details []HomeDetail) {
	if len(details) == 0 {
		pterm.Info.Println("No homes found")
		return
	}

	tableData := pterm.TableData{
		{"Home ID", "Home Name", "UID", "Devices"},
	}

	for _, detail := range details {
		tableData = append(tableData, []string{
			detail.HomeID,
			detail.HomeName,
			detail.UID,
			fmt.Sprintf("%d", detail.DeviceCount),
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	pterm.Info.Printf("Total: %d homes\n", len(details))
}
