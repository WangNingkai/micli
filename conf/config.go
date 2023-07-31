package conf

import "gopkg.in/ini.v1"

const DefaultConf = `# MiService Config
[app]
DEBUG = false
[account]
MI_USER = ""
MI_PASS = ""
MI_DID = ""
REGION = "cn"


`

var (
	Cfg      *ini.File
	ConfPath = "conf.ini"
)
