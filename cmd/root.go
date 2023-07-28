package cmd

import (
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
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
		var (
			result interface{}
			err    error
		)
		confPath := "conf.ini"
		if !Exists(confPath) {
			// 创建初始配置文件
			f, err := CreatNestedFile(confPath)
			defer f.Close()
			if err != nil {
				fmt.Printf("Fail to create config file: %v", err)
				return
			}
			// 写入配置文件
			_, err = f.WriteString(defaultConf)
			if err != nil {
				fmt.Printf("Fail to write config file: %v", err)
				return
			}

			fmt.Println("Please config your account first!")
			return
		}

		var cfg *ini.File
		cfg, err = ini.Load(confPath)
		if err != nil {
			fmt.Printf("Fail to read config file: %v", err)
			return
		}
		tokenPath := fmt.Sprintf("%s/.mi.token", os.Getenv("HOME"))
		account := miservice.NewAccount(
			cfg.Section("account").Key("MI_USER").MustString(""),
			cfg.Section("account").Key("MI_PASS").MustString(""),
			cfg.Section("account").Key("REGION").MustString("cn"),
			miservice.NewTokenStore(tokenPath),
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
				result, err = miservice.IOCommand(srv, cfg.Section("account").Key("MI_DID").MustString(""), command, args[0]+" ")
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
		fmt.Printf("Usage: Please set the local [conf.ini] file first :\n")
		fmt.Printf("           MI_USER=<Username>\n")
		fmt.Printf("           MI_PASS=<Password>\n")
		fmt.Printf("           MI_DID=<Device ID|Name>\n\n")
		fmt.Printf(miservice.IOCommandHelp("", "micli"))
		return nil
	})
}
