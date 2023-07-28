package miservice

import (
	"encoding/json"
	"errors"
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
		   
MIoT Spec: {prefix} spec [model_keyword|type_urn] [format=text|python|json]
           {prefix} spec
           {prefix} spec speaker
           {prefix} spec xiaomi.wifispeaker.lx04
           {prefix} spec urn:miot-spec-v2:device:speaker:0000A015:xiaomi-lx04:1
		   
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

func IOCommand(service *IOService, did string, text string, prefix string) (interface{}, error) {
	cmd, arg := twinsSplit(text, " ", "")
	//if strings.HasPrefix(cmd, "/") {
	//    return service.Request(cmd, arg)
	//}
	//
	if strings.HasPrefix(cmd, "prop") || cmd == "action" {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(arg), &args); err != nil {
			return nil, err
		}
		return service.Request(cmd, args)
	}

	argv := strings.Split(arg, " ")
	argc := len(argv)
	var arg0 string
	if argc > 0 {
		arg0 = argv[0]
	}
	var arg1 string
	if argc > 1 {
		arg1 = argv[1]
	}
	var arg2 string
	if argc > 2 {
		arg2 = argv[2]
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
		return service.DeviceList(a1, a2) // Implement this method for the IOService
	case "spec":
		return service.IotSpec(arg0) // Implement this method for the IOService
	case "decode":
		if argc > 3 && argv[3] == "gzip" {
			return service.IotDecode(argv[0], argv[1], argv[2], true) // Implement this method for the IOService
		}
		return service.IotDecode(argv[0], argv[1], argv[2], false) // Implement this method for the IOService
	}

	if did == "" || cmd == "" || cmd == "help" || cmd == "-h" || cmd == "--help" {
		return IOCommandHelp(did, prefix), nil
	}

	if !isDigit(did) {
		devices, err := service.DeviceList(false, 0) // Implement this method for the IOService
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

	if miot && argc > 0 {
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
		return service.MiotAction(did, ids, args)
	}

	return nil, errors.New("Unknown command: " + cmd)
}
