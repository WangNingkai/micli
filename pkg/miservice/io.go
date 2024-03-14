package miservice

import (
	"crypto/rc4"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"micli/pkg/util"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pterm/pterm"
)

const (
	MiioSid = "xiaomiio"
)

type IOService struct {
	service *Service
	baseUrl string
}

type DeviceInfo struct {
	Name  string `json:"name"`
	Model string `json:"model"`
	Did   string `json:"did"`
	Token string `json:"token"`
}

type MiotSpecInstances struct {
	Instances []struct {
		Status  string `json:"status"`
		Model   string `json:"model"`
		Version int    `json:"version"`
		Type    string `json:"type"`
		Ts      int    `json:"ts"`
	} `json:"instances"`
}

type MiotSpecInstancesData struct {
	Type        string             `json:"type"`
	Description string             `json:"description"`
	Services    []*MiotSpecService `json:"services"`
}

type MiotSpecService struct {
	Iid         int                 `json:"iid"`
	Type        string              `json:"type"`
	Description string              `json:"description"`
	Properties  []*MiotSpecProperty `json:"properties,omitempty"`
	Actions     []*MiotSpecAction   `json:"actions,omitempty"`
}

type MiotSpecProperty struct {
	Iid         int      `json:"iid"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Format      string   `json:"format"`
	Access      []string `json:"access"`
	ValueList   []struct {
		Value       int    `json:"value"`
		Description string `json:"description"`
	} `json:"value-list,omitempty"`
	ValueRange []interface{} `json:"value-range,omitempty"`
}

type MiotSpecAction struct {
	Iid         int           `json:"iid"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	In          []float64     `json:"in"`
	Out         []interface{} `json:"out"`
}

func NewIOService(service *Service) *IOService {
	host := "api.io.mi.com/app"
	protocol := "https"
	base := fmt.Sprintf("%s://%s", protocol, host)
	if service.region != "" && service.region != "cn" {
		base = fmt.Sprintf("%s://%s.%s", protocol, service.region, host)
	}
	return &IOService{service: service, baseUrl: base}
}

func (s *IOService) Request(uri string, args map[string]interface{}) (interface{}, error) {
	prepareData := func(token *Tokens, cookies map[string]string) url.Values {
		cookies["PassportDeviceId"] = token.DeviceId
		return util.SignData(uri, args, token.Sids[MiioSid].SSecurity)
	}
	headers := http.Header{
		"User-Agent":                 []string{"iOS-14.4-6.0.103-iPhone12,3--D7744744F7AF32F0544445285880DD63E47D9BE9-8816080-84A3F44E137B71AE-iPhone"},
		"x-xiaomi-protocal-flag-cli": []string{"PROTOCAL-HTTP2"},
	}
	var resp interface{}
	err := s.service.Request(MiioSid, s.baseUrl+uri, nil, prepareData, headers, true, &resp)
	if err != nil {
		return nil, err
	}
	result, ok := resp.(map[string]interface{})["result"].(interface{})
	if !ok {
		return nil, fmt.Errorf("error %s: %v", uri, resp)
	}
	return result, nil
}

func (s *IOService) DeviceList() (devices []*DeviceInfo, err error) {
	data := map[string]interface{}{
		"getVirtualModel":    true,
		"getHuamiDevices":    1,
		"get_split_device":   false,
		"support_smart_home": true,
	}
	var result interface{}
	result, err = s.Request("/home/device_list", data)
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

func (s *IOService) HomeRequest(did, method string, params interface{}) (interface{}, error) {
	data := map[string]interface{}{
		"id":        1,
		"method":    method,
		"accessKey": "IOS00026747c5acafc2",
		"params":    params,
	}
	return s.Request(fmt.Sprintf("/home/rpc/%s", did), data)
}

func (s *IOService) HomeGetProps(did string, props []string) (interface{}, error) {
	return s.HomeRequest(did, "get_prop", props)
}

func (s *IOService) HomeGetProp(did, prop string) (interface{}, error) {
	results, err := s.HomeGetProps(did, []string{prop})
	if err != nil {
		return nil, err
	}
	return results.(map[string]interface{})[prop], nil
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

func (s *IOService) HomeSetProp(did, prop string, value interface{}) (int, error) {
	result, err := s.HomeRequest(did, fmt.Sprintf("set_%s", prop), value)
	if err != nil {
		return 0, err
	}
	if result.(map[string]interface{})["result"] == "ok" {
		return 0, nil
	}
	return -1, nil
}

func (s *IOService) MiotRequest(cmd string, params interface{}) (interface{}, error) {
	return s.Request(fmt.Sprintf("/miotspec/%s", cmd), map[string]interface{}{"params": params})
}

func (s *IOService) MiotGetProps(did string, props [][]interface{}) ([]interface{}, error) {
	params := make([]map[string]interface{}, len(props))
	for i, prop := range props {
		params[i] = map[string]interface{}{
			"did":  did,
			"siid": prop[0],
			"piid": prop[1],
		}
	}
	result, err := s.MiotRequest("prop/get", params)
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

func (s *IOService) MiotGetProp(did string, prop []interface{}) (interface{}, error) {
	results, err := s.MiotGetProps(did, [][]interface{}{prop})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

func (s *IOService) MiotSetProps(did string, props [][]interface{}) ([]float64, error) {
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
	result, err := s.MiotRequest("prop/set", params)
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

func (s *IOService) MiotSetProp(did string, prop []interface{}) (float64, error) {
	results, err := s.MiotSetProps(did, [][]interface{}{prop})
	if err != nil {
		return 0, err
	}
	return results[0], nil
}

func (s *IOService) MiotAction(did string, iid []int, args []interface{}) (float64, error) {
	pterm.Info.Println(map[string]interface{}{
		"did":  did,
		"siid": iid[0],
		"aiid": iid[1],
		"in":   args,
	})
	result, err := s.MiotRequest("action", map[string]interface{}{
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

func (s *IOService) MiotSpec(keyword string) (data *MiotSpecInstancesData, err error) {
	if keyword == "" || !strings.HasPrefix(keyword, "urn") {
		//p := path.Join(os.TempDir(), "miot-spec.json")
		p := "./data/miot-spec.json"
		var specs map[string]string
		specs, err = s.loadSpec(p)
		if err != nil {
			var rr *http.Response
			rr, err = s.service.client.Get("https://miot-spec.org/miot-spec-v2/instances?status=all")
			if err != nil {
				return
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(rr.Body)
			var instanceSpec *MiotSpecInstances
			err = json.NewDecoder(rr.Body).Decode(&instanceSpec)
			if err != nil {
				return
			}
			specs = make(map[string]string)
			for _, v := range instanceSpec.Instances {
				specs[v.Model] = v.Type
			}
			var f *os.File
			// 创建初始配置文件
			f, err = util.CreatNestedFile(p)
			defer func(f *os.File) {
				_ = f.Close()
			}(f)
			if err != nil {
				return
			}
			_ = json.NewEncoder(f).Encode(specs)
		}
		specs = s.getSpec(keyword, specs)
		if len(specs) != 1 {
			instances := make([]string, 0, len(specs))
			for _, v := range specs {
				instances = append(instances, v)
			}
			err = fmt.Errorf("found %d instances: %s", len(specs), strings.Join(instances, ", "))
			return
		}
		for _, v := range specs {
			keyword = v
			break
		}
	}
	u := fmt.Sprintf("https://miot-spec.org/miot-spec-v2/instance?type=%s", keyword)
	var rs *http.Response
	rs, err = s.service.client.Get(u)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(rs.Body)

	err = json.NewDecoder(rs.Body).Decode(&data)
	if err != nil {
		return
	}

	return
}

func (s *IOService) MiotDecode(ssecurity string, nonce string, data string, gzip bool) (interface{}, error) {
	signNonceStr, err := util.SignNonce(ssecurity, nonce)
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(signNonceStr)
	if err != nil {
		return nil, err
	}
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipher.XORKeyStream(key[:1024], key[:1024])
	encryptedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(encryptedData))
	cipher.XORKeyStream(decrypted, encryptedData)

	if gzip {
		decrypted, err = util.Unzip(decrypted)
		if err != nil {
			return nil, err
		}
	}

	var result interface{}
	err = json.Unmarshal(decrypted, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *IOService) getSpec(keyword string, specs map[string]string) map[string]string {
	if keyword == "" {
		return specs
	}
	var ret = make(map[string]string)
	for k, v := range specs {
		if k == keyword {
			return map[string]string{k: v}
		} else if strings.Contains(k, keyword) {
			ret[k] = v
		}
	}
	return ret
}

func (s *IOService) loadSpec(p string) (map[string]string, error) {
	if !util.Exists(p) {
		return nil, errors.New("spec file not found")
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	j := json.NewDecoder(f)
	var specs map[string]string
	err = j.Decode(&specs)
	if err != nil {
		return nil, err
	}
	return specs, nil
}
