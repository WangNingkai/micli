package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"micli/conf"
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
				did = conf.Cfg.Section("account").Key("MI_DID").MustString("")
				if did == "" {
					did, err = chooseDevice()
					if err != nil {
						return
					}
				}
			}
			if !util.IsDigit(did) {
				var devices []*miservice.DeviceInfo
				devices, err = getDeviceListFromLocal() // Implement this method for the IOService
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
				/*var _props map[string]interface{}
				for _, prop := range props {
					_props[prop[0].(string)] = prop[1]
				}
				res, err = srv.HomeSetProps(did, _props)*/
				err = errors.New("device not support miot")
			}

			handleResult(res, err)
		},
	}
)

func init() {
	propsSetCmd.Flags().StringVarP(&did, "did", "d", "", "Device ID")
	propsSetCmd.Example = "  set 2=#60,2-2=#false,3=test"
}
