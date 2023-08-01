package cmd

import (
	"github.com/gosuri/uitable"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"micli/miservice"
	"micli/pkg/util"
	"os"
	"strings"
)

var (
	devicesPath = "devices.json"
	reload      bool
	listCmd     = &cobra.Command{
		Use:   "list [name=full|name_keyword]",
		Short: "Devs List",
		Long:  `Devs List`,
		Run: func(cmd *cobra.Command, args []string) {
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
					handleResult(nil, err)
					return
				}
				err = writeIntoLocal(devices)
				if err != nil {
					handleResult(nil, err)
					return
				}
			} else {
				devices, err = getDeviceListFromLocal()
				if err != nil {
					handleResult(nil, err)
					return
				}
			}
			if arg0 != "" {
				devices = lo.Filter(devices, func(s *miservice.DeviceInfo, index int) bool { return strings.Contains(s.Name, arg0) })
			}

			table := uitable.New()
			table.MaxColWidth = 80
			table.Wrap = true // wrap columns
			for _, device := range devices {
				table.AddRow("")
				table.AddRow("Name:", device.Name)
				table.AddRow("Did:", device.Did)
				table.AddRow("Model:", device.Model)
				table.AddRow("Token:", device.Token)
				table.AddRow("")
			}
			handleResult(table, err)
		},
	}
)

func init() {
	listCmd.Example = "  list Light"
	listCmd.Flags().BoolVarP(&reload, "reload", "r", false, "reload device list")
}

func getDeviceListFromRemote() (res []*miservice.DeviceInfo, err error) {
	res, err = srv.DeviceList()
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
	defer f.Close()
	j := json.NewDecoder(f)
	err = j.Decode(&list)
	if err != nil {
		return
	}
	return
}

func writeIntoLocal(list []*miservice.DeviceInfo) (err error) {
	var f *os.File
	f, err = os.Create(devicesPath)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(list)
	if err != nil {
		return
	}
	return
}
