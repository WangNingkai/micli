package cmd

import (
	"fmt"

	"micli/internal/conf"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var setMinaDidCmd = &cobra.Command{
	Use:   "set_did",
	Short: "Set the default mina device id",
	Long:  `Set the default mina device id`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err error
			res interface{}
		)
		did = conf.Cfg.Section("mina").Key("DID").MustString("")
		if did == "" || reset {
			did, err = chooseMinaDevice(minaSrv)
			if err != nil {
				pterm.Error.Println(err.Error())
				return
			}
		} else {
			res = fmt.Sprintf("The default mina device id already set : %s", did)
			pterm.Success.Println(res)
			return
		}

		if did != "" {
			err = conf.SetDefaultMinaDid(did)
		}
		handleResult(res, err)
	},
}

func init() {
	setMinaDidCmd.Flags().BoolVarP(&reset, "reset", "r", false, "reset the default mina device id")
}
