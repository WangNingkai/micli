package cmd

import "github.com/spf13/cobra"

var (
	testCmd = &cobra.Command{
		Use:   "test",
		Short: "Test",
		Long:  "Test",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				err error
				res interface{}
			)
			props := []string{"microphone_MicrophoneMute"}
			res, err = srv.HomeGetProps(did, props)
			handleResult(res, err)
		},
	}
)
