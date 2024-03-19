package cmd

import (
	"github.com/spf13/cobra"
)

var decodeCmd = &cobra.Command{
	Use:   "decode <ssecurity> <nonce> <data> [gzip]",
	Short: "MIoT Decode",
	Long:  `MIoT Decode`,
	Run: func(cmd *cobra.Command, args []string) {
		argLen := len(args)
		var (
			arg0, arg1, arg2, arg3 string
			err                    error
			res                    interface{}
		)
		if argLen > 0 {
			arg0 = args[0]
		}
		if argLen > 1 {
			arg1 = args[1]
		}
		if argLen > 2 {
			arg2 = args[2]
		}
		if argLen > 3 {
			arg3 = args[3]
		}
		if arg3 == "gzip" {
			res, err = ioSrv.MiotDecode(arg0, arg1, arg2, true)
		}
		res, err = ioSrv.MiotDecode(arg0, arg1, arg2, false)

		handleResult(res, err)
	},
}
