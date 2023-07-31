package cmd

import (
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"micli/miservice"
	"strconv"
)

var (
	listCmd = &cobra.Command{
		Use:   "list [name=full|name_keyword] [getVirtualModel=false|true] [getHuamiDevices=0|1]",
		Short: "Devs List",
		Long:  `Devs List`,
		Run: func(cmd *cobra.Command, args []string) {
			argLen := len(args)
			var (
				arg1, arg2 string
				err        error
			)
			if argLen > 1 {
				arg1 = args[1]
			}
			if argLen > 2 {
				arg2 = args[2]
			}
			a1 := false
			if arg1 != "" {
				a1, _ = strconv.ParseBool(arg1)
			}
			a2 := 0
			if arg2 != "" {
				a2, _ = strconv.Atoi(arg2)
			}
			var devices []*miservice.DeviceInfo
			devices, err = srv.DeviceList(a1, a2)
			if err != nil {
				return
			}
			table := uitable.New()
			table.MaxColWidth = 80
			table.Wrap = true // wrap columns
			for _, device := range devices {
				table.AddRow("")
				table.AddRow("Name:", device.Name)
				table.AddRow("Did:", device.Did)
				table.AddRow("Model:", device.Model)
				table.AddRow("Token:", device.Token)
				table.AddRow("")
			}
			handleResult(table, err)
		},
	}
)

func init() {
	listCmd.Example = "  list Light true 0"
}
