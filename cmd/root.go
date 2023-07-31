package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/gosuri/uitable"
	"github.com/pterm/pterm"
	"gopkg.in/ini.v1"
	"micli/conf"
	"micli/miservice"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	miAccount *miservice.Account
	srv       *miservice.IOService
	did       string
	rootCmd   = &cobra.Command{
		Use:   "micli",
		Short: "MiService - XiaoMi Cloud Service",
		Long:  `XiaoMi Cloud Service for mi.com`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				result interface{}
				err    error
			)
			command := strings.Join(args, " ")
			if len(args) == 0 {
				err = cmd.Help()
				return
			} else {
				result, err = miservice.IOCommand(srv, conf.Cfg.Section("account").Key("MI_DID").MustString(""), command)
			}

			handleResult(result, err)
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
	cobra.OnInitialize(initConf)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(specCmd)
	rootCmd.AddCommand(propsGetCmd)
	rootCmd.AddCommand(propsSetCmd)
	rootCmd.AddCommand(actionCmd)
	rootCmd.AddCommand(decodeCmd)
	rootCmd.AddCommand(minaCmd)
	rootCmd.AddCommand(miotRawCmd)
	rootCmd.AddCommand(miioRawCmd)
}

func initConf() {
	var err error
	conf.InitDefault()
	conf.Cfg, err = ini.Load(conf.ConfPath)
	if err != nil {
		pterm.Error.Printf("Fail to read config file: %v", err)
		os.Exit(1)
	}
	err = conf.Complete()
	if err != nil {
		pterm.Error.Println(err.Error())
		os.Exit(1)
	}
	tokenPath := fmt.Sprintf("%s/.mi.token", os.Getenv("HOME"))
	miAccount = miservice.NewAccount(
		conf.Cfg.Section("account").Key("MI_USER").MustString(""),
		conf.Cfg.Section("account").Key("MI_PASS").MustString(""),
		conf.Cfg.Section("account").Key("REGION").MustString("cn"),
		miservice.NewTokenStore(tokenPath),
	)
	srv = miservice.NewIOService(miAccount)
}

func handleResult(res interface{}, err error) {
	if err != nil {
		pterm.Error.Println(err.Error())
	} else {
		if resStr, ok := res.(string); ok {
			pterm.Info.Println(resStr)
		} else if table, ok := res.(*uitable.Table); ok {
			pterm.Println(table)
		} else {
			resBytes, _ := json.MarshalIndent(res, "", "  ")
			pterm.Info.Println(string(resBytes))
		}
	}
}
