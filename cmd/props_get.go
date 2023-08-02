package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"micli/miservice"
	"micli/pkg/util"
	"strconv"
)

var (
	propsGetCmd = &cobra.Command{
		Use:   "get <siid[-piid]>[,...]",
		Short: "MIoT Properties Get",
		Long:  `MIoT Properties Get`,
		Run: func(cmd *cobra.Command, args []string) {

			var (
				res interface{}
				err error
			)
			if did == "" {
				did, err = chooseDevice()
				if err != nil {
					return
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
				siid, iid := util.TwinsSplit(item, "-", "1")
				var prop []interface{}
				if util.IsDigit(siid) && util.IsDigit(iid) {
					s, _ := strconv.Atoi(siid)
					i, _ := strconv.Atoi(iid)
					prop = []interface{}{s, i}
				} else {
					prop = []interface{}{item}
					miot = false
				}
				props = append(props, prop)
			}

			if miot {
				res, err = srv.MiotGetProps(did, props)
			} else {
				var _props []string
				for _, prop := range props {
					_props = append(_props, prop[0].(string))
				}
				res, err = srv.HomeGetProps(did, _props)
			}
			handleResult(res, err)
		},
	}
)

func init() {
	propsGetCmd.Flags().StringVarP(&did, "did", "d", "", "Device ID")
	propsGetCmd.Example = "  get 1,1-2,1-3,1-4,2-1,2-2,3"
}
