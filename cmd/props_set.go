package cmd

import (
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"micli/conf"
	"micli/pkg/miservice"
	"micli/pkg/util"
	"strconv"
	"strings"
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
				devices, err = getDeviceListFromLocal()
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
			var props [][]interface{}
			var _args []string
			if len(args) > 0 {
				_args = strings.Split(args[0], ",")
			}
			for _, item := range _args {
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
				var data []float64
				data, err = ioSrv.MiotSetProps(did, props)
				arr := lo.Filter(data, func(item float64, index int) bool { return item != 0 })
				if len(arr) > 0 {
					err = errors.New("set failed")
				}
				res = "success."
			} else {
				/*var _props map[string]interface{}
				for _, prop := range props {
					_props[prop[0].(string)] = prop[1]
				}
				res, err = ioSrv.HomeSetProps(did, _props)*/
				err = errors.New("device not support miot")
			}

			handleResult(res, err)
		},
	}
)

func init() {
	propsSetCmd.Flags().StringVarP(&did, "did", "d", "", "Device ID")
	propsSetCmd.Example = "  set 2=60,2-2=false,3=test"
}
