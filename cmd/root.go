package cmd

import (
	"fmt"
	"github.com/gosuri/uitable"
	jsoniter "github.com/json-iterator/go"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	"micli/conf"
	"micli/miservice"
	"os"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	miAccount *miservice.Account
	srv       *miservice.IOService
	did       string
	rootCmd   = &cobra.Command{
		Version: "1.0.0",
		Use:     "micli",
		Short:   "Take XiaoMi Cloud Service to the command line",
		Long: `
MiCLI brings XiaoMi Cloud Service to your terminal. 
Free and open source.`,
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
			pterm.Println(resStr)
		} else if table, ok := res.(*uitable.Table); ok {
			pterm.Println(table)
		} else {
			resBytes, _ := json.MarshalIndent(res, "", "  ")
			pterm.Println(string(resBytes))
		}
	}
}
