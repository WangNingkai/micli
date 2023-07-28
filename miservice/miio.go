package miservice

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type IOService struct {
	account *Account
	server  string
}

type DeviceInfo struct {
	Name  string `json:"name"`
	Model string `json:"model"`
	Did   string `json:"did"`
	Token string `json:"token"`
}

type Iid struct {
	Siid int `json:"siid"`
	Piid int `json:"piid"`
}

func NewIOService(account *Account) *IOService {
	server := "https://api.io.mi.com/app"
	if account.region != "" && account.region != "cn" {
		server = fmt.Sprintf("https://%miio.api.io.mi.com/app", account.region)
	}
	return &IOService{account: account, server: server}
}

func (miio *IOService) Request(uri string, args map[string]interface{}) (map[string]interface{}, error) {
	prepareData := func(token *Tokens, cookies map[string]string) url.Values {
		cookies["PassportDeviceId"] = token.DeviceId
		return signData(uri, args, token.Sids[MiioSid].Ssecurity)
	}
	headers := http.Header{
		"User-Agent":                 []string{"iOS-14.4-6.0.103-iPhone12,3--D7744744F7AF32F0544445285880DD63E47D9BE9-8816080-84A3F44E137B71AE-iPhone"},
		"x-xiaomi-protocal-flag-cli": []string{"PROTOCAL-HTTP2"},
	}
	var resp map[string]interface{}
	err := miio.account.Request(MiioSid, miio.server+uri, nil, prepareData, headers, true, &resp)
	if err != nil {
		return nil, err
	}
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error %s: %v", uri, resp)
	}
	return result, nil
}

func (miio *IOService) DeviceList(getVirtualModel bool, getHuamiDevices int) (devices []*DeviceInfo, err error) {
	data := map[string]interface{}{
		"getVirtualModel": getVirtualModel,
		"getHuamiDevices": getHuamiDevices,
	}
	var result map[string]interface{}
	result, err = miio.Request("/home/device_list", data)
	if err != nil {
		return nil, err
	}
	deviceList := result["list"].([]interface{})
	devices = make([]*DeviceInfo, len(deviceList))
	for i, item := range deviceList {
		device := item.(map[string]interface{})
		devices[i] = &DeviceInfo{
			Name:  device["name"].(string),
			Model: device["model"].(string),
			Did:   device["did"].(string),
			Token: device["token"].(string),
		}
	}
	return
}

func (miio *IOService) HomeRequest(did, method string, params interface{}) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"id":        1,
		"method":    method,
		"accessKey": "IOS00026747c5acafc2",
		"params":    params,
	}
	return miio.Request("/home/rpc/"+did, data)
}

func (miio *IOService) HomeGetProps(did string, props []string) (map[string]interface{}, error) {
	return miio.HomeRequest(did, "get_prop", props)
}

func (miio *IOService) HomeSetProps(did string, props map[string]interface{}) (map[string]int, error) {
	results := make(map[string]int, len(props))
	for prop, value := range props {
		result, err := miio.HomeSetProp(did, prop, value)
		if err != nil {
			return nil, err
		}
		results[prop] = result
	}
	return results, nil
}

func (miio *IOService) HomeGetProp(did, prop string) (interface{}, error) {
	results, err := miio.HomeGetProps(did, []string{prop})
	if err != nil {
		return nil, err
	}
	return results[prop], nil
}

func (miio *IOService) HomeSetProp(did, prop string, value interface{}) (int, error) {
	result, err := miio.HomeRequest(did, "set_"+prop, value)
	if err != nil {
		return 0, err
	}
	if result["result"] == "ok" {
		return 0, nil
	}
	return -1, nil
}

//----------------- miot -----------------

func (miio *IOService) MiotRequest(cmd string, params interface{}) (map[string]interface{}, error) {
	return miio.Request("/miotspec/"+cmd, map[string]interface{}{"params": params})
}

func (miio *IOService) MiotGetProps(did string, iids []Iid) ([]interface{}, error) {
	params := make([]map[string]interface{}, len(iids))
	for i, iid := range iids {
		params[i] = map[string]interface{}{
			"did":  did,
			"siid": iid.Siid,
			"piid": iid.Piid,
		}
	}
	result, err := miio.MiotRequest("prop/get", params)
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(result))
	for i, it := range result {
		index, _ := strconv.Atoi(i)
		itm := it.(map[string]interface{})
		if code, ok := itm["code"].(int); ok && code == 0 {
			values[index] = itm["value"]
		} else {
			values[index] = nil
		}
	}
	return values, nil
}

func (miio *IOService) MiotSetProps(did string, props map[Iid]interface{}) ([]int, error) {
	params := make([]map[string]interface{}, len(props))
	index := 0
	for i, prop := range props {
		params[index] = map[string]interface{}{
			"did":   did,
			"siid":  i.Siid,
			"piid":  i.Piid,
			"value": prop,
		}
		index++
	}
	result, err := miio.MiotRequest("prop/set", params)
	if err != nil {
		return nil, err
	}

	codes := make([]int, len(result))
	for i, it := range result {
		ii, _ := strconv.Atoi(i)
		itm := it.(map[string]interface{})
		codes[ii] = itm["code"].(int)
	}
	return codes, nil
}

func (miio *IOService) MiotGetProp(did string, iid Iid) (interface{}, error) {
	results, err := miio.MiotGetProps(did, []Iid{iid})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

func (miio *IOService) MiotSetProp(did string, iid Iid, value interface{}) (int, error) {
	results, err := miio.MiotSetProps(did, map[Iid]interface{}{iid: value})
	if err != nil {
		return 0, err
	}
	return results[0], nil
}

func (miio *IOService) MiotAction(did string, iid []int, args []interface{}) (int, error) {
	result, err := miio.MiotRequest("action", map[string]interface{}{
		"did":  did,
		"siid": iid[0],
		"aiid": iid[1],
		"in":   args,
	})
	if err != nil {
		return -1, err
	}
	return result["code"].(int), nil
}
