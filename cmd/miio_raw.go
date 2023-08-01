package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"micli/pkg/util"
)

var (
	miioRawCmd = &cobra.Command{
		Use:   "miio_raw /<uri> <data>",
		Short: "Call MiIO Raw Request",
		Long:  `Call MiIO Raw Request`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			uri := args[0]
			if util.IsJSON(args[1]) {
				var params map[string]interface{}
				if err = json.Unmarshal([]byte(args[1]), &params); err != nil {
					return
				}
				res, err = srv.Request(uri, params)
			}
			handleResult(res, err)
		},
	}
)

func init() {
	miioRawCmd.Example = "  miio_raw /home/device_list '{\"getVirtualModel\":false,\"getHuamiDevices\":1}'\n  miio_raw /<uri> <data>"
}
