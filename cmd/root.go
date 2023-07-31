package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/gosuri/uitable"
	"github.com/pterm/pterm"
	"gopkg.in/ini.v1"
	"micli/conf"
	"micli/miservice"
	"micli/pkg/util"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "micli",
		Short: "MiService - XiaoMi Cloud Service",
		Long:  `XiaoMi Cloud Service for mi.com`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				result interface{}
				err    error
			)

			if len(args) >= 1 && args[0] == "reset" {
				confirm, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure to reset config file?")
				if confirm {
					conf.Reset()
					pterm.Info.Println("Config file has been reset.")
					return
				}
			}

			if !util.Exists(conf.ConfPath) {
				// 创建初始配置文件
				f, err := util.CreatNestedFile(conf.ConfPath)
				defer f.Close()
				if err != nil {
					pterm.Error.Printf("Fail to create config file: %v", err)
					return
				}
				// 写入配置文件
				_, err = f.WriteString(conf.DefaultConf)
				if err != nil {

					pterm.Error.Printf("Fail to write config file: %v", err)
					return
				}
				return
			}
			conf.Cfg, err = ini.Load(conf.ConfPath)
			if err != nil {
				pterm.Error.Printf("Fail to read config file: %v", err)
				return
			}
			var name, pass, region string
			needSave := false
			name = conf.Cfg.Section("account").Key("MI_USER").MustString("")
			pass = conf.Cfg.Section("account").Key("MI_PASS").MustString("")
			region = conf.Cfg.Section("account").Key("REGION").MustString("")
			if name == "" || pass == "" || region == "" {
				pterm.Warning.Println("Please complete your account information")
			}
			if name == "" {
				needSave = true
				name, err = pterm.DefaultInteractiveTextInput.Show("Enter your account username")
				if err != nil {
					pterm.Error.Printf("Fail to get your account username: %v", err)
					return
				}
				conf.Cfg.Section("account").Key("MI_USER").SetValue(name)
			}
			if pass == "" {
				needSave = true
				pass, err = pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter your password")
				if err != nil {
					pterm.Error.Printf("Fail to get your account password: %v", err)
					return
				}
				conf.Cfg.Section("account").Key("MI_PASS").SetValue(pass)
			}
			if region == "" {
				needSave = true
				opts := []string{"中国大陆", "Europe", "United States", "India", "Russia", "Singapore", "中國台灣"}
				regionMap := map[string]string{
					"中国大陆":          "cn",
					"Europe":        "de",
					"United States": "us",
					"India":         "i2",
					"Russia":        "ru",
					"Singapore":     "sg",
					"中國台灣":          "tw",
				}
				var regionI string
				regionI, err = pterm.DefaultInteractiveSelect.
					WithOptions(opts).
					WithDefaultOption("中国大陆").
					Show("Choose your account region")
				if err != nil {
					pterm.Error.Printf("Fail to get your account region: %v", err)
					return
				}
				region = regionMap[regionI]
				conf.Cfg.Section("account").Key("REGION").SetValue(region)

			}
			if needSave {
				err = conf.Cfg.SaveTo(conf.ConfPath)
				if err != nil {
					pterm.Error.Printf("Fail to write config file: %v", err)
					return
				}
				pterm.Success.Println("Config saved!Please rerun the command.")
				return
			}
			tokenPath := fmt.Sprintf("%s/.mi.token", os.Getenv("HOME"))
			account := miservice.NewAccount(
				conf.Cfg.Section("account").Key("MI_USER").MustString(""),
				conf.Cfg.Section("account").Key("MI_PASS").MustString(""),
				conf.Cfg.Section("account").Key("REGION").MustString("cn"),
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
					result, err = miservice.IOCommand(srv, conf.Cfg.Section("account").Key("MI_DID").MustString(""), command, args[0]+" ")
				}
			}

			if err != nil {
				pterm.Error.Println(err.Error())
			} else {
				if resStr, ok := result.(string); ok {
					pterm.Println(resStr)
				} else if table, ok := result.(*uitable.Table); ok {
					pterm.Println(table)
				} else {
					resBytes, _ := json.MarshalIndent(result, "", "  ")
					pterm.Println(string(resBytes))
				}
			}
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Printf(miservice.IOCommandHelp("", "micli"))
		return nil
	})
}
