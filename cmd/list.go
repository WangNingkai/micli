package cmd

import (
	"fmt"
	"os"
	"strings"

	"micli/pkg/miservice"
	"micli/pkg/util"

	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var (
	devicesPath = "./data/devices.json"
	reload      bool
	homeFilter  string
	outputJSON  bool
	listCmd     = &cobra.Command{
		Use:   "list [?name=full|name_keyword]",
		Short: "Devs List",
		Long:  `List all MiHome devices. Use --home to filter by home name.`,
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
			if homeFilter != "" {
				devices = lo.Filter(devices, func(s *miservice.DeviceInfo, index int) bool { return strings.Contains(s.HomeName, homeFilter) })
			}

			if outputJSON {
				var jsonBytes []byte
				jsonBytes, err = json.MarshalIndent(devices, "", "  ")
				if err != nil {
					pterm.Error.Println(err.Error())
					return
				}
				fmt.Println(string(jsonBytes))
				return
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
	listCmd.Flags().StringVarP(&homeFilter, "home", "H", "", "filter devices by home name")
	listCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "output in JSON format")
}

func getDeviceListFromRemote() (res []*miservice.DeviceInfo, err error) {
	res, err = ioSrv.DeviceListWithHome()
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

func resolveDevice(input string) (did string, device *miservice.DeviceInfo, err error) {
	devices, err := getDeviceListFromLocal()
	if err != nil {
		return
	}

	// Priority 1: exact DID match (all digits)
	for _, d := range devices {
		if d.Did == input {
			return d.Did, d, nil
		}
	}

	// Priority 2: delegate to AliasStore for alias/fuzzy resolution
	if aliasStore != nil {
		did, err = aliasStore.Resolve(input, devices)
		if err == nil {
			// Find the corresponding DeviceInfo
			for _, d := range devices {
				if d.Did == did {
					return did, d, nil
				}
			}
			// Alias resolved to a DID not in local cache, return anyway
			return did, nil, nil
		}
	}

	// If aliasStore.Resolve failed, return its error (e.g. ErrAmbiguousDevice or ErrDeviceNotFound)
	if aliasStore != nil && err != nil {
		return
	}

	err = miservice.ErrDeviceNotFound
	return
}
