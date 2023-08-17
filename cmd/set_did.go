package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"micli/internal/conf"
	"micli/pkg/miservice"
	"micli/pkg/util"
)

var (
	reset     bool
	setDidCmd = &cobra.Command{
		Use:   "set_did",
		Short: "Set the default device id",
		Long:  `Set the default device id`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				err error
				res interface{}
			)
			did = conf.Cfg.Section("account").Key("MI_DID").MustString("")
			if did == "" || reset {
				did, err = chooseDevice()
				if err != nil {
					pterm.Error.Println(err.Error())
					return
				}
			} else {
				pterm.Success.Printf("The default device id is %s\n", did)
				return
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
			if did != "" {
				err = conf.SetDefaultDid(did)
			}
			handleResult(res, err)
		},
	}
)

func init() {
	setDidCmd.Flags().BoolVarP(&reset, "reset", "r", false, "reset the default device id")
}
