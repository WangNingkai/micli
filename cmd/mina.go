package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"micli/miservice"
)

var (
	minaDeviceID string
	minaCmd      = &cobra.Command{
		Use:   "mina",
		Short: "Mina Service",
		Long:  `Mina Service`,
	}
)

func init() {
	minaCmd.PersistentFlags().StringVarP(&minaDeviceID, "device", "d", "", "mina service DeviceId not did")
	minaCmd.AddCommand(minaListCmd)
	minaCmd.AddCommand(minaTtsCmd)
	minaCmd.AddCommand(minaPlayerCmd)
	minaCmd.AddCommand(minaRecordsCmd)
	minaCmd.AddCommand(minaListenCmd)
}

func chooseMinaDevice(srv *miservice.MinaService) (deviceId string, err error) {
	if deviceId == "" {
		var devices []*miservice.DeviceData
		devices, err = srv.DeviceList(0)
		if err != nil {
			return
		}
		deviceMap := make(map[string]string)
		choices := make([]string, len(devices))
		for i, device := range devices {
			choice := fmt.Sprintf("%s - %s", device.Name, device.DeviceID)
			deviceMap[choice] = device.DeviceID
			choices[i] = choice
		}
		choice, _ := pterm.DefaultInteractiveSelect.
			WithDefaultText("Please select a device").
			WithOptions(choices).
			Show()
		pterm.Info.Println("Choose Device: " + choice)
		deviceId = deviceMap[choice]
	}
	return
}

func chooseMinaDeviceDetail(srv *miservice.MinaService, deviceId string) (device *miservice.DeviceData, err error) {
	var devices []*miservice.DeviceData
	devices, err = srv.DeviceList(0)
	if err != nil {
		return
	}
	deviceMap := make(map[string]*miservice.DeviceData)
	choices := make([]string, len(devices))
	for i, _device := range devices {
		if deviceId != "" {
			if _device.DeviceID == deviceId {
				device = _device
				return
			}
		}
		choice := fmt.Sprintf("%s - %s", _device.Name, _device.DeviceID)
		deviceMap[choice] = _device
		choices[i] = choice
	}
	choice, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText("Please select a device").
		WithOptions(choices).
		Show()
	pterm.Info.Println("Choose Device: " + choice)
	device = deviceMap[choice]
	return
}
