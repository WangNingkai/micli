package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"micli/pkg/util"
)

var (
	miotRawCmd = &cobra.Command{
		Use:   "miot_raw <cmd> <params>",
		Short: "Call MIoT Raw Request",
		Long:  `Call MIoT Raw Request`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			uri := args[0]
			if util.IsJSON(args[1]) {
				var params []map[string]interface{}
				if err = json.Unmarshal([]byte(args[1]), &params); err != nil {
					return
				}
				res, err = srv.MiotRequest(uri, params)
			}
			handleResult(res, err)
		},
	}
)

func init() {
	miotRawCmd.Example = "  miot_raw prop/get '[{\"did\":\"636889807\",\"siid\":2,\"piid\":1}]'\n  miot_raw <prop/get|prop/set|action> <params>"
}
