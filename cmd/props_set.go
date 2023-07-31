package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"micli/miservice"
	"micli/pkg/util"
	"strconv"
)

var (
	propsSetCmd = &cobra.Command{
		Use:   "set <siid[-piid]=[#]value>[,...]",
		Short: "MIoT Properties Set",
		Long:  `MIoT Properties Set`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			if did == "" {
				pterm.Warning.Println("default DID not set,please set it first.")
				deviceMap := make(map[string]string)
				var devices []*miservice.DeviceInfo
				devices, err = srv.DeviceList(false, 0)
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
				pterm.Info.Println("You choose: " + choice)
				did = deviceMap[choice]
			}
			if !util.IsDigit(did) {
				var devices []*miservice.DeviceInfo
				devices, err = srv.DeviceList(false, 0) // Implement this method for the IOService
				if err != nil {
					handleResult(res, err)
					return
				}
				if len(devices) == 0 {
					err = fmt.Errorf("no device found")
					handleResult(res, err)
					return
				}
				for _, device := range devices {
					if device.Name == did {
						did = device.Did
						break
					}
				}
			}

			miot := true
			var props [][]interface{}
			for _, item := range args {
				key, value := util.TwinsSplit(item, "=", "1")
				siid, iid := util.TwinsSplit(key, "-", "1")
				var prop []interface{}
				if util.IsDigit(siid) && util.IsDigit(iid) {
					s, _ := strconv.Atoi(siid)
					i, _ := strconv.Atoi(iid)
					prop = []interface{}{s, i}
				} else {
					prop = []interface{}{key}
					miot = false
				}
				prop = append(prop, util.StringOrValue(value))
				props = append(props, prop)
			}

			if miot {
				res, err = srv.MiotSetProps(did, props)
			} else {
				var _props map[string]interface{}
				for _, prop := range props {
					_props[prop[0].(string)] = prop[1]
				}
				res, err = srv.HomeSetProps(did, _props)
			}

			handleResult(res, err)
		},
	}
)

func init() {
	propsSetCmd.Flags().StringVarP(&did, "did", "d", "", "Device ID")
}
