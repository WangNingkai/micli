package conf

import (
	"github.com/pterm/pterm"
	"gopkg.in/ini.v1"
	"micli/pkg/util"
	"os"
)

const DefaultConf = `# MiService Config
[app]
DEBUG = false
[account]
MI_USER = ""
MI_PASS = ""
MI_DID = ""
REGION = ""
[openai]
KEY = ""
PROXY = ""
[mina]
DID = ""

`

var (
	Cfg      *ini.File
	ConfPath = "conf.ini"
)

func InitDefault() {
	if !util.Exists(ConfPath) {
		var (
			f   *os.File
			err error
		)
		// 创建初始配置文件
		f, err = util.CreatNestedFile(ConfPath)
		defer f.Close()
		if err != nil {
			pterm.Error.Printf("Fail to create config file: %v", err)
			os.Exit(0)
		}
		// 写入配置文件
		_, err = f.WriteString(DefaultConf)
		if err != nil {
			pterm.Error.Printf("Fail to write config file: %v", err)
			os.Exit(0)
		}
	}
}

func Reset() {
	var (
		f   *os.File
		err error
	)
	if !util.Exists(ConfPath) {
		// 创建初始配置文件
		f, err = util.CreatNestedFile(ConfPath)

		if err != nil {
			pterm.Error.Printf("Fail to create config file: %v", err)
			os.Exit(0)
		}

	} else {
		f, err = os.OpenFile(ConfPath, os.O_WRONLY|os.O_TRUNC, 0600)
	}
	defer f.Close()

	// 写入配置文件
	_, err = f.WriteString(DefaultConf)
	if err != nil {
		pterm.Error.Printf("Fail to write config file: %v", err)
		os.Exit(0)
	}
}

func Complete() (err error) {
	var name, pass, region string
	needSave := false
	name = Cfg.Section("account").Key("MI_USER").MustString("")
	pass = Cfg.Section("account").Key("MI_PASS").MustString("")
	region = Cfg.Section("account").Key("REGION").MustString("")
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
		Cfg.Section("account").Key("MI_USER").SetValue(name)
	}
	if pass == "" {
		needSave = true
		pass, err = pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter your password")
		if err != nil {
			pterm.Error.Printf("Fail to get your account password: %v", err)
			return
		}
		Cfg.Section("account").Key("MI_PASS").SetValue(pass)
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
		Cfg.Section("account").Key("REGION").SetValue(region)

	}
	if needSave {
		err = Cfg.SaveTo(ConfPath)
		if err != nil {
			pterm.Error.Printf("Fail to write config file: %v", err)
			return
		}
		pterm.Success.Println("Config saved!Please rerun the command.")
		err = Cfg.Reload()
		if err != nil {
			pterm.Error.Printf("Fail to reload config file: %v", err)
			return
		}
	}
	return
}

func SetDefaultDid(did string) (err error) {
	Cfg.Section("account").Key("MI_DID").SetValue(did)
	err = Cfg.SaveTo(ConfPath)
	if err != nil {
		pterm.Error.Printf("Fail to write config file: %v", err)
		return
	}
	pterm.Success.Println("Config saved! Please rerun the command.")
	err = Cfg.Reload()
	if err != nil {
		pterm.Error.Printf("Fail to reload config file: %v", err)
		return
	}
	return
}

func SetDefaultMinaDid(did string) (err error) {
	Cfg.Section("mina").Key("DID").SetValue(did)
	err = Cfg.SaveTo(ConfPath)
	if err != nil {
		pterm.Error.Printf("Fail to write config file: %v", err)
		return
	}
	pterm.Success.Println("Config saved! Please rerun the command.")
	err = Cfg.Reload()
	if err != nil {
		pterm.Error.Printf("Fail to reload config file: %v", err)
		return
	}
	return
}
