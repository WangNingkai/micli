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
	actionCmd = &cobra.Command{
		Use:   "action <siid[-piid]> <arg1|#NA> [...] ",
		Short: "MIoT Action",
		Long:  `MIoT Action`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			if did == "" {
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
			siid, iid := util.TwinsSplit(args[0], "-", "1")
			var prop []interface{}
			if util.IsDigit(siid) && util.IsDigit(iid) {
				s, _ := strconv.Atoi(siid)
				i, _ := strconv.Atoi(iid)
				prop = []interface{}{s, i}
			} else {
				miot = false

			}
			if miot {
				var _args []interface{}
				if args[1] != "#NA" {
					for _, a := range args[1:] {
						_args = append(_args, util.StringOrValue(a))
					}
				}
				var ids []int
				for _, id := range prop {
					if v, ok := id.(int); ok {
						ids = append(ids, v)
					} else if v, ok := id.(string); ok {
						if v2, err := strconv.Atoi(v); err == nil {
							ids = append(ids, v2)
						}
					}
				}
				res, err = srv.MiotAction(did, ids, _args)
			}
			handleResult(res, err)
		},
	}
)

func init() {
	actionCmd.Flags().StringVarP(&did, "did", "d", "", "Device ID")
}
