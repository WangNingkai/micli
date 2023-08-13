package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"micli/pkg/miservice"
	"micli/pkg/util"
	"os"
	"strings"
)

var (
	devicesPath = "./data/devices.json"
	reload      bool
	listCmd     = &cobra.Command{
		Use:   "list [?name=full|name_keyword]",
		Short: "Devs List",
		Long:  `Devs List`,
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Debug.Println("listCmd called")
			argLen := len(args)
			var (
				arg0    string
				err     error
				devices []*miservice.DeviceInfo
			)
			if argLen > 0 {
				arg0 = args[0]
			}

			if reload {
				devices, err = getDeviceListFromRemote()
				if err != nil {
					pterm.Error.Println(err.Error())
					return
				}
				err = writeIntoLocal(devices)
				if err != nil {
					pterm.Error.Println(err.Error())
					return
				}
			} else {
				devices, err = getDeviceListFromLocal()
				if err != nil {
					pterm.Error.Println(err.Error())
					return
				}
			}
			if arg0 != "" {
				devices = lo.Filter(devices, func(s *miservice.DeviceInfo, index int) bool { return strings.Contains(s.Name, arg0) })
			}

			var items []pterm.BulletListItem
			for _, device := range devices {
				items = append(items, pterm.BulletListItem{
					Level:     0,
					TextStyle: pterm.NewStyle(pterm.FgGreen),
					Text:      device.Name,
				})
				items = append(items, pterm.BulletListItem{
					Level:  1,
					Text:   fmt.Sprintf("DID: %s", device.Did),
					Bullet: ">",
				})
				items = append(items, pterm.BulletListItem{
					Level:  1,
					Text:   fmt.Sprintf("Model: %s", device.Model),
					Bullet: ">",
				})
				items = append(items, pterm.BulletListItem{
					Level:  1,
					Text:   fmt.Sprintf("Token: %s", device.Token),
					Bullet: ">",
				})
			}
			err = pterm.DefaultBulletList.WithItems(items).Render()
			if err != nil {
				pterm.Error.Println(err.Error())
			}

		},
	}
)

func init() {
	listCmd.Example = "  list Light"
	listCmd.Flags().BoolVarP(&reload, "reload", "r", false, "reload device list")
}

func getDeviceListFromRemote() (res []*miservice.DeviceInfo, err error) {
	res, err = ioSrv.DeviceList()
	return
}

func getDeviceListFromLocal() (list []*miservice.DeviceInfo, err error) {
	if !util.Exists(devicesPath) {
		list, err = getDeviceListFromRemote()
		if err != nil {
			return
		}
		err = writeIntoLocal(list)
		if err != nil {
			return
		}
		return
	}
	var f *os.File
	f, err = os.Open(devicesPath)
	if err != nil {
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	err = json.NewDecoder(f).Decode(&list)
	if err != nil {
		return
	}
	return
}

func writeIntoLocal(list []*miservice.DeviceInfo) (err error) {
	var f *os.File
	f, err = util.CreatNestedFile(devicesPath)
	if err != nil {
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	err = json.NewEncoder(f).Encode(list)
	if err != nil {
		return
	}
	return
}

func chooseDevice() (did string, err error) {
	deviceMap := make(map[string]string)
	var devices []*miservice.DeviceInfo
	devices, err = getDeviceListFromLocal()
	if err != nil {
		return
	}
	choices := make([]string, len(devices))
	for i, device := range devices {
		choice := fmt.Sprintf("%s - %s", device.Name, device.Did)
		deviceMap[choice] = device.Did
		choices[i] = choice
	}
	choice, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText("Please select a device").
		WithOptions(choices).
		Show()
	pterm.Info.Println("Choose Device: " + choice)
	did = deviceMap[choice]
	return
}
