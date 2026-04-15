package miservice

import (
	"crypto/rc4"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"micli/pkg/util"

	"github.com/pterm/pterm"
)

const (
	MiioSid = "mijia"
)

type IOService struct {
	service *Service
	baseUrl string
}

type DeviceInfo struct {
	Name     string `json:"name"`
	Model    string `json:"model"`
	Did      string `json:"did"`
	Token    string `json:"token"`
	HomeName string `json:"home_name,omitempty"`
	HomeID   string `json:"home_id,omitempty"`
}

// HomeInfo represents a Xiaomi smart home.
type HomeInfo struct {
	HomeID   string `json:"id"`
	HomeName string `json:"name"`
	UID      string `json:"uid"`
}

// SceneInfo represents a smart scene/automation.
type SceneInfo struct {
	SceneID   string `json:"scene_id"`
	SceneName string `json:"name"`
	HomeID    string `json:"home_id"`
	OwnerUID  string `json:"owner_uid"`
	Enabled   bool   `json:"enabled"`
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

// DevProp - property from home.miot-spec.com (semantic names)
type DevProp struct {
	Piid       int            `json:"iid"`
	DescZhCN   string         `json:"desc_zh_cn"`
	DescEn     string         `json:"desc_en"`
	Type       string         `json:"type"`
	Format     string         `json:"format"`
	Access     []string       `json:"access"`
	ValueList  []DevValueItem `json:"value-list"`
	ValueRange *DevValueRange `json:"value-range,omitempty"`
	Unit       string         `json:"unit,omitempty"`
}

// DevAction - action from home.miot-spec.com
type DevAction struct {
	Aiid     int    `json:"iid"`
	DescZhCN string `json:"desc_zh_cn"`
	DescEn   string `json:"desc_en"`
	Type     string `json:"type"`
	In       []int  `json:"in"`
	Out      []int  `json:"out"`
}

// DevService - service from home.miot-spec.com
type DevService struct {
	Siid     int                  `json:"iid"`
	DescZhCN string               `json:"desc_zh_cn"`
	DescEn   string               `json:"desc_en"`
	Type     string               `json:"type"`
	Props    map[string]DevProp   `json:"properties"`
	Actions  map[string]DevAction `json:"actions"`
}

// DevSpec - full device spec from home.miot-spec.com
type DevSpec struct {
	Model    string                `json:"model"`
	Type     string                `json:"type"`
	DescZhCN string                `json:"desc_zh_cn"`
	DescEn   string                `json:"desc_en"`
	Services map[string]DevService `json:"services"`
}

// DevValueItem - enum value item
type DevValueItem struct {
	Value    interface{} `json:"value"`
	DescZhCN string      `json:"desc_zh_cn"`
	DescEn   string      `json:"desc_en"`
}

// DevValueRange - numeric range [min, max, step]
type DevValueRange []float64

// DevParam - action parameter
type DevParam struct {
	Piid     int    `json:"piid"`
	DescZhCN string `json:"desc_zh_cn"`
	DescEn   string `json:"desc_en"`
	Type     string `json:"type"`
}

func NewIOService(service *Service) *IOService {
	host := "api.mijia.tech/app"
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
		"User-Agent":                 []string{"Android-15-11.0.701-Xiaomi-23046RP50C-OS2.0.212.0.VMYCNXM-MI_APP_STORE"},
		"x-xiaomi-protocal-flag-cli": []string{"PROTOCAL-HTTP2"},
	}
	var resp interface{}
	err := s.service.Request(MiioSid, s.baseUrl+uri, nil, prepareData, headers, true, &resp)
	if err != nil {
		return nil, err
	}
	result, ok := resp.(map[string]interface{})["result"]
	if !ok {
		// Extract error details from API response
		resultMap, isMap := resp.(map[string]interface{})
		if isMap {
			if code, hasCode := resultMap["code"]; hasCode {
				if message, hasMsg := resultMap["message"]; hasMsg {
					return nil, fmt.Errorf("api error %s: code=%v, message=%v", uri, code, message)
				}
				return nil, fmt.Errorf("api error %s: code=%v", uri, code)
			}
		}
		return nil, fmt.Errorf("api error %s: no result field, response=%v", uri, resp)
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
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map[string]interface{} for device_list result, got %T", result)
	}
	list, ok := resultMap["list"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected []interface{} for 'list', got %T", resultMap["list"])
	}
	devices = make([]*DeviceInfo, 0, len(list))
	for _, item := range list {
		device, ok := item.(map[string]interface{})
		if !ok {
			pterm.Warning.Printf("device item is not map[string]interface{}, got %T, skipping\n", item)
			continue
		}
		name, ok := device["name"].(string)
		if !ok {
			pterm.Warning.Printf("device 'name' is not string, got %T, skipping\n", device["name"])
			continue
		}
		model, ok := device["model"].(string)
		if !ok {
			pterm.Warning.Printf("device 'model' is not string, got %T, skipping\n", device["model"])
			continue
		}
		did, ok := device["did"].(string)
		if !ok {
			pterm.Warning.Printf("device 'did' is not string, got %T, skipping\n", device["did"])
			continue
		}
		token, ok := device["token"].(string)
		if !ok {
			pterm.Warning.Printf("device 'token' is not string, got %T, skipping\n", device["token"])
			continue
		}
		devices = append(devices, &DeviceInfo{
			Name:  name,
			Model: model,
			Did:   did,
			Token: token,
		})
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
	resultMap, ok := results.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map[string]interface{} for get_prop result, got %T", results)
	}
	return resultMap[prop], nil
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
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return -1, fmt.Errorf("expected map[string]interface{} for set_prop result, got %T", result)
	}
	if resultMap["result"] == "ok" {
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
	resultList, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected []interface{} for prop/get result, got %T", result)
	}
	values := make([]interface{}, len(resultList))
	for i, it := range resultList {
		itm, ok := it.(map[string]interface{})
		if !ok {
			pterm.Warning.Printf("prop/get item %d is not map[string]interface{}, got %T\n", i, it)
			values[i] = nil
			continue
		}
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

	resultList, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected []interface{} for prop/set result, got %T", result)
	}
	codes := make([]float64, len(resultList))
	for i, it := range resultList {
		itm, ok := it.(map[string]interface{})
		if !ok {
			pterm.Warning.Printf("prop/set item %d is not map[string]interface{}, got %T\n", i, it)
			codes[i] = -1
			continue
		}
		code, ok := itm["code"].(float64)
		if !ok {
			pterm.Warning.Printf("prop/set item %d 'code' is not float64, got %T\n", i, itm["code"])
			codes[i] = -1
			continue
		}
		codes[i] = code
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
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return -1, fmt.Errorf("expected map[string]interface{} for action result, got %T", result)
	}
	code, ok := resultMap["code"].(float64)
	if !ok {
		return -1, fmt.Errorf("expected float64 for action 'code', got %T", resultMap["code"])
	}
	return code, nil
}

func (s *IOService) MiotSpec(keyword string) (data *MiotSpecInstancesData, err error) {
	if keyword == "" || !strings.HasPrefix(keyword, "urn") {
		// p := path.Join(os.TempDir(), "miot-spec.json")
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

// GetDeviceInfo fetches device specification from home.miot-spec.com
// Returns semantic property/action names with Chinese descriptions
func (s *IOService) GetDeviceInfo(model string) (*DevSpec, error) {
	if model == "" {
		return nil, errors.New("model is required")
	}

	url := fmt.Sprintf("https://home.miot-spec.com/spec/%s", model)
	resp, err := s.service.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch spec for %s: %w", model, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spec not found for model %s (status %d)", model, resp.StatusCode)
	}

	htmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Find JSON in <div id="app" data-page="...">
	htmlStr := string(htmlData)
	marker := `<div id="app" data-page="`
	jsonStart := strings.Index(htmlStr, marker)
	if jsonStart == -1 {
		return nil, errors.New("no embedded JSON found in spec page")
	}
	jsonStart += len(marker)

	jsonEnd := strings.Index(htmlStr[jsonStart:], `">`)
	if jsonEnd == -1 {
		return nil, errors.New("malformed spec page structure")
	}
	jsonStr := htmlStr[jsonStart : jsonStart+jsonEnd]

	// Decode HTML entities (&quot; -> ", etc.)
	jsonStr = strings.ReplaceAll(jsonStr, `&quot;`, "\"")
	jsonStr = strings.ReplaceAll(jsonStr, `&amp;`, "&")
	jsonStr = strings.ReplaceAll(jsonStr, `&#39;`, "'")

	var pageData struct {
		Props struct {
			Spec DevSpec `json:"spec"`
		} `json:"props"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &pageData); err != nil {
		return nil, fmt.Errorf("parse spec JSON: %w", err)
	}

	return &pageData.Props.Spec, nil
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
	ret := make(map[string]string)
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

// HomeList retrieves the list of smart homes for the current user.
func (s *IOService) HomeList() (homes []*HomeInfo, err error) {
	data := map[string]interface{}{
		"fg":              true,
		"fetch_share":     true,
		"fetch_share_dev": true,
		"fetch_cariot":    true,
		"limit":           300,
		"app_ver":         7,
		"plat_form":       0,
	}
	var result interface{}
	result, err = s.Request("/v2/homeroom/gethome_merged", data)
	if err != nil {
		return nil, fmt.Errorf("failed to get home list: %w", err)
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type for gethome_merged: %T", result)
	}
	homelist, ok := resultMap["homelist"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected homelist type: %T", resultMap["homelist"])
	}
	for _, item := range homelist {
		home, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var idStr string
		switch v := home["id"].(type) {
		case string:
			idStr = v
		case float64:
			idStr = fmt.Sprintf("%.0f", v)
		}
		name, _ := home["name"].(string)
		var uidStr string
		switch v := home["uid"].(type) {
		case string:
			uidStr = v
		case float64:
			uidStr = fmt.Sprintf("%.0f", v)
		}
		homes = append(homes, &HomeInfo{
			HomeID:   idStr,
			HomeName: name,
			UID:      uidStr,
		})
	}
	return
}

// SceneList retrieves all manual scenes for all homes.
func (s *IOService) SceneList() (scenes []*SceneInfo, err error) {
	homes, err := s.HomeList()
	if err != nil {
		return nil, err
	}
	for _, home := range homes {
		data := map[string]interface{}{
			"app_version": 12,
			"get_type":    2,
			"home_id":     home.HomeID,
			"owner_uid":   home.UID,
		}
		var result interface{}
		result, err = s.Request("/appgateway/miot/appsceneservice/AppSceneService/GetSimpleSceneList", data)
		if err != nil {
			pterm.Warning.Printf("Failed to get scenes for home %s: %v\n", home.HomeName, err)
			continue
		}
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}
		list, ok := resultMap["manual_scene_info_list"].([]interface{})
		if !ok {
			continue
		}
		for _, item := range list {
			scene, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			sceneID, _ := scene["scene_id"].(string)
			name, _ := scene["name"].(string)
			enabled := true
			if en, ok := scene["enabled"].(bool); ok {
				enabled = en
			}
			scenes = append(scenes, &SceneInfo{
				SceneID:   sceneID,
				SceneName: name,
				HomeID:    home.HomeID,
				OwnerUID:  home.UID,
				Enabled:   enabled,
			})
		}
	}
	return
}

// RunScene executes a manual scene by ID and homeID.
// It fetches the home list to resolve the owner_uid.
func (s *IOService) RunScene(sceneID, homeID string) (interface{}, error) {
	scenes, err := s.SceneList()
	if err != nil {
		return nil, err
	}
	for _, sc := range scenes {
		if sc.SceneID == sceneID && sc.HomeID == homeID {
			return s.runSceneInternal(sc.SceneID, sc.HomeID, sc.OwnerUID)
		}
	}
	// Scene not found in SceneList, try direct home lookup
	return s.runSceneByHomeID(sceneID, homeID)
}

// runSceneInternal executes a scene when we already have the owner_uid.
func (s *IOService) runSceneInternal(sceneID, homeID, ownerUID string) (interface{}, error) {
	data := map[string]interface{}{
		"scene_id":   sceneID,
		"scene_type": 2,
		"phone_id":   "null",
		"home_id":    homeID,
		"owner_uid":  ownerUID,
	}
	return s.Request("/appgateway/miot/appsceneservice/AppSceneService/NewRunScene", data)
}

// runSceneByHomeID fetches home list to resolve owner_uid for a given homeID.
func (s *IOService) runSceneByHomeID(sceneID, homeID string) (interface{}, error) {
	homes, err := s.HomeList()
	if err != nil {
		return nil, err
	}
	var uid string
	for _, h := range homes {
		if h.HomeID == homeID {
			uid = h.UID
			break
		}
	}
	if uid == "" {
		return nil, ErrHomeNotFound
	}
	return s.runSceneInternal(sceneID, homeID, uid)
}

// ResolveScene finds a scene by ID or name (fuzzy match).
// Returns (sceneID, homeID, error).
func (s *IOService) ResolveScene(input string) (sceneID, homeID string, err error) {
	scenes, err := s.SceneList()
	if err != nil {
		return "", "", err
	}
	// Exact ID match
	for _, sc := range scenes {
		if sc.SceneID == input {
			return sc.SceneID, sc.HomeID, nil
		}
	}
	// Fuzzy name match
	var matched []*SceneInfo
	for _, sc := range scenes {
		if strings.Contains(sc.SceneName, input) {
			matched = append(matched, sc)
		}
	}
	if len(matched) == 1 {
		return matched[0].SceneID, matched[0].HomeID, nil
	}
	if len(matched) > 1 {
		names := make([]string, len(matched))
		for i, sc := range matched {
			names[i] = sc.SceneName
		}
		return "", "", fmt.Errorf("multiple scenes matches '%s': %s", input, strings.Join(names, ", "))
	}
	return "", "", ErrSceneNotFound
}

// DeviceListWithHome retrieves all devices and populates HomeName/HomeID.
func (s *IOService) DeviceListWithHome() ([]*DeviceInfo, error) {
	devices, err := s.DeviceList()
	if err != nil {
		return nil, err
	}
	homes, err := s.HomeList()
	if err != nil {
		pterm.Warning.Printf("Failed to get home list, devices will not include home information: %v\n", err)
		return devices, nil // Return devices even if home lookup fails
	}
	homeMap := make(map[string]*HomeInfo)
	for _, h := range homes {
		homeMap[h.HomeID] = h
	}
	// The device list API does not return home_id directly.
	// We need to fetch devices per home to associate them.
	for _, home := range homes {
		homeDevices, err := s.HomeDeviceList(home.HomeID, home.UID)
		if err != nil {
			continue
		}
		homeDeviceDids := make(map[string]bool)
		for _, d := range homeDevices {
			homeDeviceDids[d] = true
		}
		for _, d := range devices {
			if homeDeviceDids[d.Did] {
				d.HomeName = home.HomeName
				d.HomeID = home.HomeID
			}
		}
	}
	return devices, nil
}

// HomeDeviceList retrieves devices for a specific home.
func (s *IOService) HomeDeviceList(homeID, ownerUID string) (dids []string, err error) {
	data := map[string]interface{}{
		"home_owner":         ownerUID,
		"home_id":            homeID,
		"limit":              200,
		"start_did":          "",
		"get_split_device":   true,
		"support_smart_home": true,
		"get_cariot_device":  true,
		"get_third_device":   true,
	}
	startDid := ""
	hasMore := true
	for hasMore {
		data["start_did"] = startDid
		var result interface{}
		result, err = s.Request("/home/home_device_list", data)
		if err != nil {
			return nil, fmt.Errorf("failed to get home devices: %w", err)
		}
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			break
		}
		list, ok := resultMap["device_info"].([]interface{})
		if !ok {
			break
		}
		for _, item := range list {
			device, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if did, ok := device["did"].(string); ok {
				dids = append(dids, did)
			}
		}
		startDid, _ = resultMap["max_did"].(string)
		hasMore, _ = resultMap["has_more"].(bool)
		hasMore = hasMore && startDid != ""
	}
	return
}

// GetStatistics retrieves device statistics (e.g., power consumption).
func (s *IOService) GetStatistics(did, key, dataType string, timeStart, timeEnd int64) (interface{}, error) {
	data := map[string]interface{}{
		"did":        did,
		"key":        key,
		"data_type":  dataType,
		"limit":      100,
		"time_start": timeStart,
		"time_end":   timeEnd,
	}
	return s.Request("/v2/user/statistics", data)
}

// GetConsumables retrieves consumable items (e.g., filter life, battery).
// If homeFilter is non-empty, only consumables for that home are returned.
func (s *IOService) GetConsumables(homeFilter string) (interface{}, error) {
	homes, err := s.HomeList()
	if err != nil {
		return nil, err
	}
	var allItems []interface{}
	for _, home := range homes {
		if homeFilter != "" && home.HomeName != homeFilter {
			continue
		}
		data := map[string]interface{}{
			"home_id":       home.HomeID,
			"owner_id":      home.UID,
			"filter_ignore": true,
		}
		var result interface{}
		result, err = s.Request("/v2/home/standard_consumable_items", data)
		if err != nil {
			pterm.Warning.Printf("Failed to get consumables for home %s: %v\n", home.HomeName, err)
			continue
		}
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}
		if items, ok := resultMap["items"].([]interface{}); ok {
			for _, item := range items {
				if group, ok := item.(map[string]interface{}); ok {
					if consumes, ok := group["consumes_data"].([]interface{}); ok {
						allItems = append(allItems, consumes...)
					}
				}
			}
		}
	}
	return allItems, nil
}
