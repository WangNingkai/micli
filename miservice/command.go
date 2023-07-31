package miservice

import (
	"encoding/json"
	"errors"
	"fmt"
	inf "github.com/fzdwx/infinite"
	"github.com/fzdwx/infinite/components/selection/singleselect"
	"github.com/fzdwx/infinite/theme"
	"github.com/gosuri/uitable"
	"micli/conf"
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

Call MIoT: {prefix} <cmd=prop/get|/prop/set|action> <params>
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

func IOCommand(srv *IOService, did string, command string, prefix string) (res interface{}, err error) {
	cmd, arg := twinsSplit(command, " ", "")
	if cmd == "" || cmd == "help" || cmd == "-h" || cmd == "--help" {
		return IOCommandHelp(did, prefix), nil
	}
	if strings.HasPrefix(cmd, "/") {
		if isJSON(arg) {
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
	var arg0, arg1, arg2, arg3 string
	if argLen > 0 {
		arg0 = argv[0]
	}
	if argLen > 1 {
		arg1 = argv[1]
	}
	if argLen > 2 {
		arg2 = argv[2]
	}
	if argLen > 3 {
		arg3 = argv[3]
	}
	switch cmd {
	case "list":
		a1 := false
		if arg1 != "" {
			a1, _ = strconv.ParseBool(arg1)
		}
		a2 := 0
		if arg2 != "" {
			a2, _ = strconv.Atoi(arg2)
		}
		var devices []*DeviceInfo
		devices, err = srv.DeviceList(a1, a2)
		if err != nil {
			return
		}
		table := uitable.New()
		table.MaxColWidth = 80
		table.Wrap = true // wrap columns
		for _, device := range devices {
			table.AddRow("") // blank
			table.AddRow("Name:", device.Name)
			table.AddRow("Did:", device.Did)
			table.AddRow("Model:", device.Model)
			table.AddRow("Token:", device.Token)
			table.AddRow("") // blank
		}
		return table, nil
	case "spec":
		return srv.MiotSpec(arg0)
	case "decode":
		if arg3 == "gzip" {
			return srv.MiotDecode(arg0, arg1, arg2, true)
		}
		return srv.MiotDecode(arg0, arg1, arg2, false)
	}

	if did == "" {
		fmt.Println("default DID not set,please set it first.")
		deviceMap := make(map[string]string)
		var devices []*DeviceInfo
		devices, err = srv.DeviceList(false, 0)
		choices := make([]string, len(devices))
		for i, device := range devices {
			choice := fmt.Sprintf("%s - %s", device.Name, device.Did)
			deviceMap[choice] = device.Did
			choices[i] = choice
		}
		didIndex, _ := inf.NewSingleSelect(
			choices,
			singleselect.WithPrompt("choose your default did"),
			singleselect.WithPromptStyle(theme.DefaultTheme.PromptStyle),
			singleselect.WithRowRender(func(c string, h string, choice string) string {
				return fmt.Sprintf("%s [%s] %s", c, h, choice)
			}),
		).Display()
		choice := choices[didIndex]
		did = deviceMap[choice]
		conf.Cfg.Section("account").Key("MI_DID").SetValue(did)
		err = conf.Cfg.SaveTo(conf.ConfPath)
		if err != nil {
			return nil, errors.New("DID not set")
		}
	}
	if !isDigit(did) {
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
		if isJSON(arg) {
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
		key, value := twinsSplit(item, "=", "")
		siid, iid := twinsSplit(key, "-", "1")
		var prop []interface{}
		if isDigit(siid) && isDigit(iid) {
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
			prop = append(prop, stringOrValue(value))
		}
		props = append(props, prop)
	}

	if miot && argLen > 0 && arg != "" {
		var args []interface{}
		if arg != "#NA" {
			for _, a := range argv {
				args = append(args, stringOrValue(a))
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
