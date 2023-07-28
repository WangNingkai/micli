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

const MiioSid = "xiaomiio"

func NewIOService(account *Account, region *string) *IOService {
	server := "https://"
	if region != nil && *region != "cn" {
		server += *region + "."
	}
	server += "api.io.mi.com/app"
	return &IOService{account: account, server: server}
}

func (s *IOService) Request(uri string, data map[string]interface{}) (map[string]interface{}, error) {
	prepareData := func(token *Tokens, cookies map[string]string) url.Values {
		cookies["PassportDeviceId"] = token.DeviceId
		return signData(uri, data, token.Sids[MiioSid].Ssecurity)
	}

	headers := http.Header{
		"User-Agent":                 []string{"iOS-14.4-6.0.103-iPhone12,3--D7744744F7AF32F0544445285880DD63E47D9BE9-8816080-84A3F44E137B71AE-iPhone"},
		"x-xiaomi-protocal-flag-cli": []string{"PROTOCAL-HTTP2"},
	}
	var resp map[string]interface{}
	err := s.account.Request(MiioSid, s.server+uri, nil, prepareData, headers, true, &resp)
	if err != nil {
		return nil, err
	}

	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error %s: %v", uri, resp)
	}

	return result, nil
}

type DeviceInfo struct {
	Name  string `json:"name"`
	Model string `json:"model"`
	Did   string `json:"did"`
	Token string `json:"token"`
}

func (s *IOService) DeviceList(getVirtualModel bool, getHuamiDevices int) (devices []DeviceInfo, err error) {
	data := map[string]interface{}{
		"getVirtualModel": getVirtualModel,
		"getHuamiDevices": getHuamiDevices,
	}
	result, err := s.Request("/home/device_list", data)
	if err != nil {
		return nil, err
	}
	deviceList := result["list"].([]interface{})

	devices = make([]DeviceInfo, len(deviceList))
	for i, item := range deviceList {
		device := item.(map[string]interface{})
		devices[i] = DeviceInfo{
			Name:  device["name"].(string),
			Model: device["model"].(string),
			Did:   device["did"].(string),
			Token: device["token"].(string),
		}
	}
	return
}

func (s *IOService) HomeRequest(did, method string, params interface{}) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"id":        1,
		"method":    method,
		"accessKey": "IOS00026747c5acafc2",
		"params":    params,
	}
	return s.Request("/home/rpc/"+did, data)
}

func (s *IOService) HomeGetProps(did string, props []string) (map[string]interface{}, error) {
	return s.HomeRequest(did, "get_prop", props)
}

func (s *IOService) HomeSetProps(did string, props map[string]interface{}) (map[string]int, error) {
	results := make(map[string]int, len(props))
	for prop, value := range props {
		result, err := s.HomeSetProp(did, prop, value)
		if err != nil {
			return nil, err
		}
		results[prop] = result
	}
	return results, nil
}

func (s *IOService) HomeGetProp(did, prop string) (interface{}, error) {
	results, err := s.HomeGetProps(did, []string{prop})
	if err != nil {
		return nil, err
	}
	return results[prop], nil
}

func (s *IOService) HomeSetProp(did, prop string, value interface{}) (int, error) {
	result, err := s.HomeRequest(did, "set_"+prop, value)
	if err != nil {
		return 0, err
	}
	if result["result"] == "ok" {
		return 0, nil
	}
	return -1, nil
}

//----------------- miot -----------------

func (s *IOService) MiotRequest(cmd string, params interface{}) (map[string]interface{}, error) {
	return s.Request("/miotspec/"+cmd, map[string]interface{}{"params": params})
}

type Iid struct {
	Siid int `json:"siid"`
	Piid int `json:"piid"`
}

func (s *IOService) MiotGetProps(did string, iids []Iid) ([]interface{}, error) {
	params := make([]map[string]interface{}, len(iids))
	for i, iid := range iids {
		params[i] = map[string]interface{}{
			"did":  did,
			"siid": iid.Siid,
			"piid": iid.Piid,
		}
	}
	result, err := s.MiotRequest("prop/get", params)
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

func (s *IOService) MiotSetProps(did string, props map[Iid]interface{}) ([]int, error) {
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
	result, err := s.MiotRequest("prop/set", params)
	if err != nil {
		return nil, err
	}

	codes := make([]int, len(result))
	for i, it := range result {
		index, _ := strconv.Atoi(i)
		itm := it.(map[string]interface{})
		codes[index] = itm["code"].(int)
	}
	return codes, nil
}

func (s *IOService) MiotGetProp(did string, iid Iid) (interface{}, error) {
	results, err := s.MiotGetProps(did, []Iid{iid})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

func (s *IOService) MiotSetProp(did string, iid Iid, value interface{}) (int, error) {
	results, err := s.MiotSetProps(did, map[Iid]interface{}{iid: value})
	if err != nil {
		return 0, err
	}
	return results[0], nil
}

func (s *IOService) MiotAction(did string, iid []int, args []interface{}) (int, error) {
	result, err := s.MiotRequest("action", map[string]interface{}{
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
