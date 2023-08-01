package miservice

import (
	"fmt"
	"micli/pkg/util"
	"net/http"
	"net/url"
)

// DEVICES: http://miot-spec.org/miot-spec-v2/instances?status=all
// device types: http://miot-spec.org/miot-spec-v2/spec/devices
// service types: http://miot-spec.org/miot-spec-v2/spec/services

// Miot subcommand response codes
// | 0 | Success |
// | 1 | Request received, but the operation has not been completed yet |
// | -4001 | Unreadable attribute |
// | -4002 | Attribute is not writable |
// | -4003 | Properties, methods, events do not exist |
// | -4004 | Other internal errors |
// | -4005 | Attribute value error |
// | -4006 | Method in parameter error |
// | -4007 | did error

type MinaService struct {
	account *Account
}

type DeviceData struct {
	DeviceID     string `json:"deviceID"`
	SerialNumber string `json:"serialNumber"`
	Name         string `json:"name"`
	Alias        string `json:"alias"`
	Current      bool   `json:"current"`
	Presence     string `json:"presence"`
	Address      string `json:"address"`
	MiotDID      string `json:"miotDID"`
	Hardware     string `json:"hardware"`
	RomVersion   string `json:"romVersion"`
	Capabilities struct {
		ChinaMobileIms      int `json:"china_mobile_ims"`
		SchoolTimetable     int `json:"school_timetable"`
		NightMode           int `json:"night_mode"`
		UserNickName        int `json:"user_nick_name"`
		PlayerPauseTimer    int `json:"player_pause_timer"`
		DialogH5            int `json:"dialog_h5"`
		ChildMode2          int `json:"child_mode_2"`
		ReportTimes         int `json:"report_times"`
		AlarmVolume         int `json:"alarm_volume"`
		AiInstruction       int `json:"ai_instruction"`
		ClassifiedAlarm     int `json:"classified_alarm"`
		AiProtocol30        int `json:"ai_protocol_3_0"`
		NightModeDetail     int `json:"night_mode_detail"`
		ChildMode           int `json:"child_mode"`
		BabySchedule        int `json:"baby_schedule"`
		ToneSetting         int `json:"tone_setting"`
		Earthquake          int `json:"earthquake"`
		AlarmRepeatOptionV2 int `json:"alarm_repeat_option_v2"`
		XiaomiVoip          int `json:"xiaomi_voip"`
		NearbyWakeupCloud   int `json:"nearby_wakeup_cloud"`
		FamilyVoice         int `json:"family_voice"`
		BluetoothOptionV2   int `json:"bluetooth_option_v2"`
		Yunduantts          int `json:"yunduantts"`
		MicoCurrent         int `json:"mico_current"`
		VoipUsedTime        int `json:"voip_used_time"`
	} `json:"capabilities"`
	RemoteCtrlType  string `json:"remoteCtrlType"`
	DeviceSNProfile string `json:"deviceSNProfile"`
	DeviceProfile   string `json:"deviceProfile"`
	BrokerEndpoint  string `json:"brokerEndpoint"`
	BrokerIndex     int    `json:"brokerIndex"`
	Mac             string `json:"mac"`
	Ssid            string `json:"ssid"`
}

type Devices struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    []*DeviceData `json:"data"`
}

type PlayerStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Code int    `json:"code"`
		Info string `json:"info"`
	} `json:"data"`
}

func NewMinaService(account *Account) *MinaService {
	return &MinaService{
		account: account,
	}
}

func (mina *MinaService) Request(uri string, data url.Values, out any) error {
	requestId := "app_ios_" + util.GetRandom(30)
	if data != nil {
		data["requestId"] = []string{requestId}
	} else {
		uri += "&requestId=" + requestId
	}

	headers := http.Header{
		"User-Agent": []string{"MiHome/6.0.103 (com.xiaomi.mihome; build:6.0.103.1; iOS 14.4.0) Alamofire/6.0.103 MICO/iOSApp/appStore/6.0.103"},
	}

	return mina.account.Request("micoapi", fmt.Sprintf("https://api2.mina.mi.com%s", uri), data, nil, headers, true, out)
}

func (mina *MinaService) DeviceList(master int) (devices []*DeviceData, err error) {
	var res *Devices
	err = mina.Request(fmt.Sprintf("/admin/v2/device_list?master=%d", master), nil, &res)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (mina *MinaService) UbusRequest(deviceId, method, path string, message map[string]interface{}, res any) error {
	messageJSON, _ := json.Marshal(message)
	data := url.Values{
		"deviceId": []string{deviceId},
		"message":  []string{string(messageJSON)},
		"method":   []string{method},
		"path":     []string{path},
	}

	err := mina.Request("/remote/ubus", data, res)
	if err != nil {
		return err
	}
	return nil
}

func (mina *MinaService) TextToSpeech(deviceId, text string) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := mina.UbusRequest(deviceId, "text_to_speech", "mibrain", map[string]interface{}{"text": text}, &res)
	return res, err
}

func (mina *MinaService) PlayerSetVolume(deviceId string, volume int) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := mina.UbusRequest(deviceId, "player_set_volume", "mediaplayer", map[string]interface{}{"volume": volume, "media": "app_ios"}, &res)
	return res, err
}

func (mina *MinaService) PlayerPause(deviceId string) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := mina.UbusRequest(deviceId, "player_play_operation", "mediaplayer", map[string]interface{}{"action": "pause", "media": "app_ios"}, &res)
	return res, err
}

func (mina *MinaService) PlayerPlay(deviceId string) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := mina.UbusRequest(deviceId, "player_play_operation", "mediaplayer", map[string]interface{}{"action": "play", "media": "app_ios"}, &res)
	return res, err
}

func (mina *MinaService) PlayerGetStatus(deviceId string) (
	*PlayerStatus, error) {
	var res PlayerStatus
	err := mina.UbusRequest(deviceId, "player_get_play_status", "mediaplayer", map[string]interface{}{"media": "app_ios"}, &res)
	return &res, err
}

func (mina *MinaService) PlayByUrl(deviceId, url string) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := mina.UbusRequest(deviceId, "player_play_url", "mediaplayer", map[string]interface{}{"url": url, "type": 1, "media": "app_ios"}, &res)
	return res, err
}

func (mina *MinaService) SendMessage(devices []*DeviceData, devNo int, message string, volume *int) (bool, error) {
	result := false
	for i, device := range devices {
		if devNo == -1 || devNo != i+1 || device.Capabilities.Yunduantts != 0 {
			deviceId := device.DeviceID
			if volume != nil {
				res, err := mina.PlayerSetVolume(deviceId, *volume)
				result = err == nil && res != nil
			}
			if message != "" {
				res, err := mina.TextToSpeech(deviceId, message)
				result = err == nil && res != nil
			}
			if devNo != -1 || !result {
				break
			}
		}
	}

	return result, nil
}
