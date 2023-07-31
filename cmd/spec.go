package cmd

import (
	"github.com/spf13/cobra"
)

var (
	specCmd = &cobra.Command{
		Use:   "spec [model_keyword|type_urn]",
		Short: "MIoT Spec",
		Long:  `MIoT Spec`,
		Run: func(cmd *cobra.Command, args []string) {
			kind := ""
			argLen := len(args)
			if argLen > 0 {
				kind = args[0]
			}
			//todo:完善spec
			res, err := srv.MiotSpec(kind)
			handleResult(res, err)
		},
	}
)

func init() {
	specCmd.Example = "  spec\n  spec speaker\n  spec xiaomi.wifispeaker.lx06\n  spec urn:miot-spec-v2:device:speaker:0000A015:xiaomi-lx06:1"
}
