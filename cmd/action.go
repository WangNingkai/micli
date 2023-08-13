package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"micli/conf"
	"micli/pkg/miservice"
	"micli/pkg/util"
	"strconv"
)

var (
	actionCmd = &cobra.Command{
		Use:   "action <siid[-piid]> <arg1> [...] ",
		Short: "MIoT Action",
		Long:  `MIoT Action`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			if did == "" {
				did = conf.Cfg.Section("account").Key("MI_DID").MustString("")
				if did == "" {
					did, err = chooseDevice()
					if err != nil {
						pterm.Error.Println(err.Error())
						return
					}
				}
			}
			if !util.IsDigit(did) {
				var devices []*miservice.DeviceInfo
				devices, err = getDeviceListFromLocal() // Implement this method for the IOService
				if err != nil {
					pterm.Error.Println(err.Error())
					return
				}
				if len(devices) == 0 {
					err = fmt.Errorf("no device found")
					pterm.Error.Println(err.Error())
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
				if len(args) > 1 {
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
				res, err = ioSrv.MiotAction(did, ids, _args)
				if res.(float64) == 0 {
					res = "success."
				}
			}
			handleResult(res, err)
		},
	}
)

func init() {
	actionCmd.Flags().StringVarP(&did, "did", "d", "", "Device ID")
	actionCmd.Example = "  action 2 #NA\n  action 5 Hello #1\n  action 5-4 Hello 0"
}
