package miservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"micli/conf"
	"micli/pkg/util"
	"strconv"
	"strings"
)

var template = `Get Props: {prefix} <siid[-piid]>[,...]
           {prefix} 1,1-2,1-3,1-4,2-1,2-2,3
		   
Set Props: {prefix} <siid[-piid]=[#]value>[,...]
           {prefix} 2=#60,2-2=#false,3=test
		   
Do Action: {prefix} <siid[-piid]> <arg1|#NA> [...] 
           {prefix} 2 #NA
           {prefix} 5 Hello
           {prefix} 5-4 Hello #1

Call MIoT: {prefix} <cmd=prop/get|prop/set|action> <params>
           {prefix} action {quote}{{"did":"{did}","siid":5,"aiid":1,"in":["Hello"]}}{quote}
		   
Call MiIO: {prefix} /<uri> <data>
           {prefix} /home/device_list {quote}{{"getVirtualModel":false,"getHuamiDevices":1}}{quote}

Devs List: {prefix} list [name=full|name_keyword] [getVirtualModel=false|true] [getHuamiDevices=0|1]
           {prefix} list Light true 0
		   
MIoT Spec: {prefix} spec [model_keyword|type_urn]
           {prefix} spec
           {prefix} spec speaker
           {prefix} spec xiaomi.wifispeaker.lx06
           {prefix} spec urn:miot-spec-v2:device:speaker:0000A015:xiaomi-lx06:1
		   
MIoT Decode: {prefix} decode <ssecurity> <nonce> <data> [gzip]

Reset Account: {prefix} reset 

`

func IOCommandHelp(did string, prefix string) string {
	var quote string
	if prefix == "" {
		prefix = "?"
		quote = ""
	} else {
		quote = "'"
	}
	tmp := strings.ReplaceAll(template, "{prefix}", prefix)
	tmp = strings.ReplaceAll(tmp, "{quote}", quote)
	if did == "" {
		did = "267090026"
	}
	tmp = strings.ReplaceAll(tmp, "{did}", did)
	return tmp
}

func IOCommand(srv *IOService, did string, command string) (res interface{}, err error) {
	cmd, arg := util.TwinsSplit(command, " ", "")
	if strings.HasPrefix(cmd, "/") {
		if util.IsJSON(arg) {
			var args map[string]interface{}
			if err = json.Unmarshal([]byte(arg), &args); err != nil {
				return
			}
			return srv.Request(cmd, args)
		}
		err = errors.New("unsupported command")
		return
	}

	argv := strings.Split(arg, " ")
	argLen := len(argv)

	if did == "" {
		pterm.Warning.Println("default DID not set,please set it first.")
		deviceMap := make(map[string]string)
		var devices []*DeviceInfo
		devices, err = srv.DeviceList(false, 0)
		choices := make([]string, len(devices))
		for i, device := range devices {
			choice := fmt.Sprintf("%s - %s", device.Name, device.Did)
			deviceMap[choice] = device.Did
			choices[i] = choice
		}

		choice, _ := pterm.DefaultInteractiveSelect.
			WithDefaultText("Please select a device").
			WithOptions(choices).
			Show()
		pterm.Info.Println("You choose: " + choice)
		did = deviceMap[choice]
		conf.Cfg.Section("account").Key("MI_DID").SetValue(did)
		err = conf.Cfg.SaveTo(conf.ConfPath)
		if err != nil {
			return nil, errors.New("DID not set")
		}
		if did == "" {
			return nil, errors.New("DID not set")
		}
	}
	if !util.IsDigit(did) {
		var devices []*DeviceInfo
		devices, err = srv.DeviceList(false, 0) // Implement this method for the IOService
		if err != nil {
			return nil, err
		}
		if len(devices) == 0 {
			return nil, errors.New("Device not found: " + did)
		}
		for _, device := range devices {
			if device.Name == did {
				did = device.Did
				break
			}
		}
	}

	if strings.HasPrefix(cmd, "prop") || cmd == "action" {
		if util.IsJSON(arg) {
			var params []map[string]interface{}
			if err = json.Unmarshal([]byte(arg), &params); err != nil {
				return
			}
			return srv.MiotRequest(cmd, params)
		}
	}

	var props [][]interface{}
	setp := true
	miot := true
	for _, item := range strings.Split(cmd, ",") {
		key, value := util.TwinsSplit(item, "=", "")
		siid, iid := util.TwinsSplit(key, "-", "1")
		var prop []interface{}
		if util.IsDigit(siid) && util.IsDigit(iid) {
			s, _ := strconv.Atoi(siid)
			i, _ := strconv.Atoi(iid)
			prop = []interface{}{s, i}
		} else {
			prop = []interface{}{key}
			miot = false
		}
		if value == "" {
			setp = false
		} else if setp {
			prop = append(prop, util.StringOrValue(value))
		}
		props = append(props, prop)
	}

	if miot && argLen > 0 && arg != "" {
		var args []interface{}
		if arg != "#NA" {
			for _, a := range argv {
				args = append(args, util.StringOrValue(a))
			}
		}
		var ids []int
		for _, id := range props[0] {
			if v, ok := id.(int); ok {
				ids = append(ids, v)
			} else if v, ok := id.(string); ok {
				if v2, err := strconv.Atoi(v); err == nil {
					ids = append(ids, v2)
				}
			}
		}
		return srv.MiotAction(did, ids, args)
	}

	if setp {
		if miot {
			return srv.MiotSetProps(did, props)
		} else {
			var _props map[string]interface{}
			for _, prop := range props {
				_props[prop[0].(string)] = prop[1]
			}
			return srv.HomeSetProps(did, _props)
		}
	} else {
		if miot {
			return srv.MiotGetProps(did, props)
		} else {
			var _props []string
			for _, prop := range props {
				_props = append(_props, prop[0].(string))
			}
			return srv.HomeGetProps(did, _props)
		}
	}
}
