package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"micli/pkg/util"
)

var (
	miotRawCmd = &cobra.Command{
		Use:   "miot_raw [uri] [data]",
		Short: "Call MIoT Raw Command",
		Long:  `Call MIoT Raw Command`,
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
