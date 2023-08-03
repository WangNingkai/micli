package cmd

import (
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"micli/miservice"
	"strconv"
	"strings"
	"time"
)

var (
	minaDeviceID string
	minaCmd      = &cobra.Command{
		Use:   "mina <list|tts|player|records> <arg1|keyword|message|play|pause|volume|status> <?arg2>",
		Short: "Mina Service",
		Long:  `Mina Service`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			srv := miservice.NewMinaService(miAccount)
			if len(args) > 0 {
				command := args[0]
				switch command {
				case "list":
					res, err = list(srv, "")
				case "tts":
					if len(args) > 1 {
						res, err = doTTS(srv, args[1:])
					} else {
						err = errors.New("tts message is empty")
					}
				case "player":
					res, err = operatePlayer(srv, args[1:])
				case "records":
					res, err = askRecords(srv, args[1:])
				default:
					err = fmt.Errorf("unknown command: %s", command)
				}
			} else {
				res, err = list(srv, "")
			}

			handleResult(res, err)
		},
	}
)

func init() {
	minaCmd.Example = "  mina list 小爱 \n  mina tts message \n  mina records \n  mina player status \n  mina player play\n  mina player pause\n  mina player volume 50"
	minaCmd.PersistentFlags().StringVarP(&minaDeviceID, "device", "d", "", "device id")
}

// list 设备列表
func list(srv *miservice.MinaService, keyword string) (res interface{}, err error) {
	var devices []*miservice.DeviceData
	devices, err = srv.DeviceList(0)
	if err != nil {
		return
	}
	if keyword != "" {
		devices = lo.Filter(devices, func(s *miservice.DeviceData, index int) bool { return strings.Contains(s.Name, keyword) })
	}
	var items []pterm.BulletListItem
	for i, device := range devices {
		items = append(items, pterm.BulletListItem{
			Level:     0,
			TextStyle: pterm.NewStyle(pterm.FgGreen),
			Text:      device.Name,
		})
		items = append(items, pterm.BulletListItem{
			Level:  1,
			Text:   fmt.Sprintf("Index: %d", i+1),
			Bullet: ">",
		})
		items = append(items, pterm.BulletListItem{
			Level:  1,
			Text:   fmt.Sprintf("Hardware: %s", device.Hardware),
			Bullet: ">",
		})
		items = append(items, pterm.BulletListItem{
			Level:  1,
			Text:   fmt.Sprintf("DeviceID: %s", device.DeviceID),
			Bullet: ">",
		})
		items = append(items, pterm.BulletListItem{
			Level:  1,
			Text:   fmt.Sprintf("Presence: %s", device.Presence),
			Bullet: ">",
		})
		items = append(items, pterm.BulletListItem{
			Level:  1,
			Text:   fmt.Sprintf("MiotDID: %s", device.MiotDID),
			Bullet: ">",
		})
	}
	err = pterm.DefaultBulletList.WithItems(items).Render()
	return
}

// doTTS 语音合成
func doTTS(srv *miservice.MinaService, args []string) (res interface{}, err error) {
	var message string
	deviceId := minaDeviceID
	if deviceId == "" {
		deviceId, err = chooseMinaDevice(srv)
		if err != nil {
			return
		}
	}
	if len(args) >= 1 {
		message = args[0]
	} else {
		err = errors.New("message is empty")
		return
	}
	res, err = srv.TextToSpeech(deviceId, message)
	return
}

// operatePlayer 播放器操作
func operatePlayer(srv *miservice.MinaService, args []string) (res interface{}, err error) {
	deviceId := minaDeviceID
	if deviceId == "" {
		deviceId, err = chooseMinaDevice(srv)
		if err != nil {
			return
		}
	}
	if len(args) > 0 {
		command := args[0]
		switch command {
		case "status":
			var statusData *miservice.PlayerStatus
			statusData, err = srv.PlayerGetStatus(deviceId)
			type info struct {
				Status int `json:"status"`
				Volume int `json:"volume"`
			}
			var dataInfo *info
			err = json.Unmarshal([]byte(statusData.Data.Info), &dataInfo)
			if err != nil {
				return
			}
			var items []pterm.BulletListItem
			items = append(items, pterm.BulletListItem{
				Level:     0,
				TextStyle: pterm.NewStyle(pterm.FgGreen),
				Text:      deviceId,
			})
			items = append(items, pterm.BulletListItem{
				Level:  1,
				Text:   fmt.Sprintf("Status: %d", dataInfo.Status),
				Bullet: ">",
			})
			items = append(items, pterm.BulletListItem{
				Level:  1,
				Text:   fmt.Sprintf("Volume: %d", dataInfo.Volume),
				Bullet: ">",
			})
			err = pterm.DefaultBulletList.WithItems(items).Render()
		case "play":
			res, err = srv.PlayerPlay(deviceId)
		case "pause":
			res, err = srv.PlayerPause(deviceId)
		case "volume":
			if len(args) > 1 {
				volume, _ := strconv.Atoi(args[1])
				res, err = srv.PlayerSetVolume(deviceId, volume)
			} else {
				err = errors.New("volume is empty")
			}
		}
	} else {
		err = errors.New("player command is empty")
	}
	return
}

func askRecords(srv *miservice.MinaService, args []string) (res interface{}, err error) {
	var limit int
	if len(args) > 0 {
		limit, _ = strconv.Atoi(args[0])
	} else {
		limit = 10
	}

	deviceId := minaDeviceID
	var device *miservice.DeviceData
	device, err = chooseMinaDeviceDetail(srv, deviceId)
	if err != nil {
		return
	}
	var resp *miservice.AskRecords
	err = srv.LastAskList(device.DeviceID, device.Hardware, limit, &resp)

	var record *miservice.AskRecord
	err = json.Unmarshal([]byte(resp.Data), &record)
	if err != nil {
		return
	}
	var items []pterm.BulletListItem
	items = append(items, pterm.BulletListItem{
		Level:     0,
		TextStyle: pterm.NewStyle(pterm.FgGreen),
		Text:      device.Name,
	})

	for _, _record := range record.Records {
		items = append(items, pterm.BulletListItem{
			Level:     1,
			Text:      fmt.Sprintf("Time: %s", time.UnixMilli(_record.Time).Format("2006-01-02 15:04:05")),
			Bullet:    "-",
			TextStyle: pterm.NewStyle(pterm.FgCyan),
		})
		items = append(items, pterm.BulletListItem{
			Level:  2,
			Text:   fmt.Sprintf("Q: %s", _record.Query),
			Bullet: ">",
		})
		var a string
		if len(_record.Answers) > 0 {
			a = _record.Answers[0].Tts.Text
		}

		items = append(items, pterm.BulletListItem{
			Level:  2,
			Text:   fmt.Sprintf("A: %s", a),
			Bullet: ">",
		})
	}
	err = pterm.DefaultBulletList.WithItems(items).Render()

	return

}

func chooseMinaDevice(srv *miservice.MinaService) (deviceId string, err error) {
	if deviceId == "" {
		var devices []*miservice.DeviceData
		devices, err = srv.DeviceList(0)
		if err != nil {
			return
		}
		deviceMap := make(map[string]string)
		choices := make([]string, len(devices))
		for i, device := range devices {
			choice := fmt.Sprintf("%s - %s", device.Name, device.DeviceID)
			deviceMap[choice] = device.DeviceID
			choices[i] = choice
		}
		choice, _ := pterm.DefaultInteractiveSelect.
			WithDefaultText("Please select a device").
			WithOptions(choices).
			Show()
		pterm.Info.Println("Choose Device: " + choice)
		deviceId = deviceMap[choice]
	}
	return
}

func chooseMinaDeviceDetail(srv *miservice.MinaService, deviceId string) (device *miservice.DeviceData, err error) {
	var devices []*miservice.DeviceData
	devices, err = srv.DeviceList(0)
	if err != nil {
		return
	}
	deviceMap := make(map[string]*miservice.DeviceData)
	choices := make([]string, len(devices))
	for i, _device := range devices {
		if deviceId != "" {
			if _device.DeviceID == deviceId {
				device = _device
				return
			}
		}
		choice := fmt.Sprintf("%s - %s", _device.Name, _device.DeviceID)
		deviceMap[choice] = _device
		choices[i] = choice
	}
	choice, _ := pterm.DefaultInteractiveSelect.
		WithDefaultText("Please select a device").
		WithOptions(choices).
		Show()
	pterm.Info.Println("Choose Device: " + choice)
	device = deviceMap[choice]
	return
}
