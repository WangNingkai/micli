package cmd

import (
	"encoding/json"
	"fmt"
	"micli/miservice"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "micli",
	Short: "MiService - XiaoMi Cloud Service",
	Long:  `XiaoMi Cloud Service for mi.com`,
	Run: func(cmd *cobra.Command, args []string) {
		tokenPath := fmt.Sprintf("%s/.mi.token", os.Getenv("HOME"))
		account := miservice.NewAccount(
			os.Getenv("MI_USER"),
			os.Getenv("MI_PASS"),
			"cn",
			miservice.NewTokenStore(tokenPath),
		)
		var (
			result interface{}
			err    error
		)
		command := strings.Join(args, " ")
		if len(args) == 0 {
			result = miservice.IOCommandHelp("", "micli")
		} else {
			if args[0] == "mina" {
				srv := miservice.NewMinaService(account)
				deviceList, err := srv.DeviceList(0)
				if err == nil && len(command) > 4 {
					_, _ = srv.SendMessage(deviceList, -1, command[4:], nil)
					result = "Message sent!"
				} else {
					result = deviceList
				}
			} else {
				srv := miservice.NewIOService(account)
				result, err = miservice.IOCommand(srv, os.Getenv("MI_DID"), command, args[0]+" ")
			}
		}

		if err != nil {
			fmt.Println(err)
		} else {
			if resStr, ok := result.(string); ok {
				fmt.Println(resStr)
			} else {
				resBytes, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(resBytes))
			}
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Printf("Usage: The following variables must be set:\n")
		fmt.Printf("           export MI_USER=<Username>\n")
		fmt.Printf("           export MI_PASS=<Password>\n")
		fmt.Printf("           export MI_DID=<Device ID|Name>\n\n")
		fmt.Printf(miservice.IOCommandHelp("", "micli"))
		return nil
	})
}
