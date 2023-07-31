package cmd

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"micli/pkg/util"
)

var (
	miioRawCmd = &cobra.Command{
		Use:   "miio_raw [uri] [data]",
		Short: "Call MIIO Raw Command",
		Long:  `Call MIIO Raw Command`,
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
