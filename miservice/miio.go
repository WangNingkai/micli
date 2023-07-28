package miservice

import (
	"fmt"
	"net/http"
	"net/url"
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

func NewIOService(account *Account) *IOService {
	server := "https://api.io.mi.com/app"
	if account.region != "" && account.region != "cn" {
		server = fmt.Sprintf("https://%miio.api.io.mi.com/app", account.region)
	}
	return &IOService{account: account, server: server}
}

func (miio *IOService) Request(uri string, args map[string]interface{}) (interface{}, error) {
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
	//fmt.Println(resp)
	result, ok := resp["result"].(interface{})
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
	var result interface{}
	result, err = miio.Request("/home/device_list", data)
	if err != nil {
		return nil, err
	}
	deviceList := result.(map[string]interface{})["list"].([]interface{})
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

func (miio *IOService) HomeRequest(did, method string, params interface{}) (interface{}, error) {
	data := map[string]interface{}{
		"id":        1,
		"method":    method,
		"accessKey": "IOS00026747c5acafc2",
		"params":    params,
	}
	return miio.Request("/home/rpc/"+did, data)
}

func (miio *IOService) HomeGetProps(did string, props []string) (interface{}, error) {
	return miio.HomeRequest(did, "get_prop", props)
}

func (miio *IOService) HomeGetProp(did, prop string) (interface{}, error) {
	results, err := miio.HomeGetProps(did, []string{prop})
	if err != nil {
		return nil, err
	}
	return results.(map[string]interface{})[prop], nil
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

func (miio *IOService) HomeSetProp(did, prop string, value interface{}) (int, error) {
	result, err := miio.HomeRequest(did, "set_"+prop, value)
	if err != nil {
		return 0, err
	}
	if result.(map[string]interface{})["result"] == "ok" {
		return 0, nil
	}
	return -1, nil
}

// ----------------- miot -----------------
func (miio *IOService) MiotRequest(cmd string, params interface{}) (interface{}, error) {
	return miio.Request("/miotspec/"+cmd, map[string]interface{}{"params": params})
}

func (miio *IOService) MiotGetProps(did string, props [][]interface{}) ([]interface{}, error) {
	params := make([]map[string]interface{}, len(props))
	for i, prop := range props {
		params[i] = map[string]interface{}{
			"did":  did,
			"siid": prop[0],
			"piid": prop[1],
		}
	}
	result, err := miio.MiotRequest("prop/get", params)
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, len(result.([]interface{})))
	for i, it := range result.([]interface{}) {
		itm := it.(map[string]interface{})
		if code, ok := itm["code"].(float64); ok && code == 0 {
			values[i] = itm["value"]
		} else {
			values[i] = nil
		}
	}
	return values, nil
}

func (miio *IOService) MiotGetProp(did string, prop []interface{}) (interface{}, error) {
	results, err := miio.MiotGetProps(did, [][]interface{}{prop})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

func (miio *IOService) MiotSetProps(did string, props [][]interface{}) ([]float64, error) {
	params := make([]map[string]interface{}, len(props))
	index := 0
	for _, prop := range props {
		params[index] = map[string]interface{}{
			"did":   did,
			"siid":  prop[0],
			"piid":  prop[1],
			"value": prop[2],
		}
		index++
	}
	result, err := miio.MiotRequest("prop/set", params)
	if err != nil {
		return nil, err
	}

	codes := make([]float64, len(result.([]interface{})))
	for i, it := range result.([]interface{}) {
		itm := it.(map[string]interface{})
		codes[i] = itm["code"].(float64)
	}
	return codes, nil
}

func (miio *IOService) MiotSetProp(did string, prop []interface{}) (float64, error) {
	results, err := miio.MiotSetProps(did, [][]interface{}{prop})
	if err != nil {
		return 0, err
	}
	return results[0], nil
}

func (miio *IOService) MiotAction(did string, iid []int, args []interface{}) (float64, error) {
	result, err := miio.MiotRequest("action", map[string]interface{}{
		"did":  did,
		"siid": iid[0],
		"aiid": iid[1],
		"in":   args,
	})
	if err != nil {
		return -1, err
	}
	return result.(map[string]interface{})["code"].(float64), nil
}
