package cmd

import (
	"encoding/json"
	"fmt"
	inf "github.com/fzdwx/infinite"
	"github.com/fzdwx/infinite/components/input/text"
	"github.com/fzdwx/infinite/components/selection/singleselect"
	"github.com/fzdwx/infinite/theme"
	"github.com/gosuri/uitable"
	"github.com/ttacon/chalk"
	"gopkg.in/ini.v1"
	"micli/conf"
	"micli/miservice"
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

			if !Exists(conf.ConfPath) {
				// 创建初始配置文件
				f, err := CreatNestedFile(conf.ConfPath)
				defer f.Close()
				if err != nil {
					fmt.Printf("%sFail to create config file: %v", chalk.Red, err)
					return
				}
				// 写入配置文件
				_, err = f.WriteString(conf.DefaultConf)
				if err != nil {

					fmt.Printf("%sFail to write config file: %v", chalk.Red, err)
					return
				}
				return
			}

			conf.Cfg, err = ini.Load(conf.ConfPath)
			if err != nil {
				fmt.Printf("%sFail to read config file: %v", chalk.Red, err)
				return
			}
			needSave := false
			if conf.Cfg.Section("account").Key("MI_USER").MustString("") == "" {
				needSave = true
				name, _ := inf.NewText(
					text.WithRequired(),
					text.WithPrompt("what's your account username?"),
					text.WithPromptStyle(theme.DefaultTheme.PromptStyle),
					text.WithDefaultValue(""),
				).Display()
				conf.Cfg.Section("account").Key("MI_USER").SetValue(name)
			}
			if conf.Cfg.Section("account").Key("MI_PASS").MustString("") == "" {
				needSave = true
				pass, _ := inf.NewText(
					text.WithRequired(),
					text.WithPrompt("what's your password?"),
					text.WithPromptStyle(theme.DefaultTheme.PromptStyle),
					text.WithDefaultValue(""),
					text.WithDisableOutputResult(),
					text.WithEchoPassword(),
				).Display()
				conf.Cfg.Section("account").Key("MI_PASS").SetValue(pass)
			}
			if conf.Cfg.Section("account").Key("REGION").MustString("") == "" {
				needSave = true
				opts := []string{"cn(中国大陆)", "de(Europe)", "us(United States)", "i2(India)", "ru(Russia)", "sg(Singapore)", "tw(中國台灣)"}
				optValues := []string{"cn", "de", "us", "i2", "ru", "sg", "tw"}
				region, _ := inf.NewSingleSelect(opts,
					singleselect.WithPrompt("choose your region"),
					singleselect.WithPromptStyle(theme.DefaultTheme.PromptStyle),
					singleselect.WithRowRender(func(c string, h string, choice string) string {
						return fmt.Sprintf("%s [%s] %s", c, h, choice)
					}),
				).Display()
				conf.Cfg.Section("account").Key("REGION").SetValue(optValues[region])

			}
			if needSave {
				err = conf.Cfg.SaveTo(conf.ConfPath)
				if err != nil {
					fmt.Printf("%sFail to write config file: %v", chalk.Red, err)
					return
				}
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
				fmt.Println(chalk.Red, err)
			} else {
				if resStr, ok := result.(string); ok {
					fmt.Println(chalk.Green, resStr)
				} else if table, ok := result.(*uitable.Table); ok {
					fmt.Println(chalk.Green, table)
				} else {
					resBytes, _ := json.MarshalIndent(result, "", "  ")
					fmt.Println(chalk.Green, string(resBytes))
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
		fmt.Printf("Usage: Please set the local [conf.ini] file first :\n")
		fmt.Printf("           MI_USER=<Username>\n")
		fmt.Printf("           MI_PASS=<Password>\n")
		fmt.Printf("           MI_DID=<Device ID|Name>\n\n")
		fmt.Printf(miservice.IOCommandHelp("", "micli"))
		return nil
	})
}
