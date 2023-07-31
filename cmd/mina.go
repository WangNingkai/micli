package cmd

import (
	"github.com/spf13/cobra"
	"micli/miservice"
	"strings"
)

var (
	minaCmd = &cobra.Command{
		Use:   "mina [command]",
		Short: "Mina Service",
		Long:  `Mina Service`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			command := strings.Join(args, " ")
			srv := miservice.NewMinaService(miAccount)
			var devices []*miservice.DeviceData
			devices, err = srv.DeviceList(0)
			if err == nil && len(args) > 3 {
				_, _ = srv.SendMessage(devices, -1, command, nil)
				res = "Message sent!"
			} else {
				res = devices
			}
			handleResult(res, err)
		},
	}
)
