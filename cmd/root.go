package cmd

import (
	"fmt"
	"os"

	"micli/internal/conf"
	"micli/pkg/miservice"

	jsoniter "github.com/json-iterator/go"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	ms           *miservice.Service
	ioSrv        *miservice.IOService
	minaSrv      *miservice.MinaService
	did          string
	minaDeviceID string
	rootCmd      = &cobra.Command{
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
		os.Exit(0)
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
	rootCmd.AddCommand(setDidCmd)
	rootCmd.AddCommand(ttsCmd)
	rootCmd.AddCommand(TestCmd)
}

func initConf() {
	var err error
	conf.InitDefault()
	conf.Cfg, err = ini.Load(conf.ConfPath)
	if conf.Cfg.Section("app").Key("DEBUG").MustBool(false) {
		pterm.EnableDebugMessages()
	}
	if err != nil {
		pterm.Error.Printf("Fail to read config file: %v", err)
		os.Exit(0)
	}
	err = conf.Complete()
	if err != nil {
		pterm.Error.Println(err.Error())
		os.Exit(0)
	}
	tokenPath := fmt.Sprintf("%s/.mi.token", os.Getenv("HOME"))
	ms = miservice.New(
		conf.Cfg.Section("account").Key("MI_USER").MustString(""),
		conf.Cfg.Section("account").Key("MI_PASS").MustString(""),
		conf.Cfg.Section("account").Key("REGION").MustString("cn"),
		miservice.NewTokenStore(tokenPath),
	)
	ioSrv = miservice.NewIOService(ms)
	minaSrv = miservice.NewMinaService(ms)
}

func handleResult(res interface{}, err error) {
	if err != nil {
		pterm.Error.Println(err.Error())
	} else {
		if res == nil {
			return
		}
		if resStr, ok := res.(string); ok {
			pterm.NewStyle(pterm.FgGreen).Println(resStr)
		} else {
			resBytes, _ := json.MarshalIndent(res, "", "  ")
			pterm.NewStyle(pterm.FgGreen).Println(string(resBytes))
		}
	}
}
