package conf

import (
	"github.com/pterm/pterm"
	"gopkg.in/ini.v1"
	"micli/pkg/util"
)

const DefaultConf = `# MiService Config
[app]
DEBUG = false
[account]
MI_USER = ""
MI_PASS = ""
MI_DID = ""
REGION = ""


`

var (
	Cfg      *ini.File
	ConfPath = "conf.ini"
)

func Reset() {
	if !util.Exists(ConfPath) {
		// 创建初始配置文件
		f, err := util.CreatNestedFile(ConfPath)
		defer f.Close()
		if err != nil {
			pterm.Error.Printf("Fail to create config file: %v", err)
			return
		}
		// 写入配置文件
		_, err = f.WriteString(DefaultConf)
		if err != nil {

			pterm.Error.Printf("Fail to write config file: %v", err)
			return
		}
		return
	}
}
