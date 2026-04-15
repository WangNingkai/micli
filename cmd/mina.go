package cmd

import (
	"fmt"

	"micli/pkg/miservice"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var minaCmd = &cobra.Command{
	Use:   "mina",
	Short: "Mina Service",
	Long:  `Mina Service`,
}

func init() {
	minaCmd.PersistentFlags().StringVarP(&minaDeviceID, "device", "d", "", "mina service device id, not miot did")
	minaCmd.AddCommand(minaListCmd)
	minaCmd.AddCommand(minaTtsCmd)
	minaCmd.AddCommand(minaRunCmd)
	minaCmd.AddCommand(minaPlayerCmd)
	minaCmd.AddCommand(minaRecordsCmd)
	minaCmd.AddCommand(minaServeCmd)
	minaCmd.AddCommand(setMinaDidCmd)
}

// selectMinaDevice 选择小爱设备，返回设备ID或完整设备信息
// 如果 returnDetail 为 true，返回完整设备信息；否则只返回设备ID
// 如果 targetDeviceId 不为空，直接匹配该设备；否则交互式选择
func selectMinaDevice(srv *miservice.MinaService, targetDeviceId string, returnDetail bool) (deviceId string, device *miservice.DeviceData, err error) {
	var devices []*miservice.DeviceData
	devices, err = srv.DeviceList(0)
	if err != nil {
		return
	}

	// 如果指定了设备ID，直接查找
	if targetDeviceId != "" {
		for _, d := range devices {
			if d.DeviceID == targetDeviceId {
				deviceId = d.DeviceID
				device = d
				return
			}
		}
	}

	// 交互式选择
	deviceMap := make(map[string]*miservice.DeviceData)
	choices := make([]string, len(devices))
	for i, d := range devices {
		choice := fmt.Sprintf("%s - %s", d.Name, d.DeviceID)
		deviceMap[choice] = d
		choices[i] = choice
	}
	choice, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText("Please select a device").
		WithOptions(choices).
		Show()
	pterm.Info.Println("Choose Device: " + choice)

	selected := deviceMap[choice]
	deviceId = selected.DeviceID
	device = selected
	return
}

// chooseMinaDevice 选择小爱设备，返回设备ID
func chooseMinaDevice(srv *miservice.MinaService) (string, error) {
	deviceId, _, err := selectMinaDevice(srv, minaDeviceID, false)
	return deviceId, err
}

// chooseMinaDeviceDetail 选择小爱设备，返回完整设备信息
func chooseMinaDeviceDetail(srv *miservice.MinaService, targetDeviceId string) (*miservice.DeviceData, error) {
	_, device, err := selectMinaDevice(srv, targetDeviceId, true)
	return device, err
}
