package cmd

import (
	"errors"
	"fmt"
	"github.com/gosuri/uitable"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"micli/miservice"
	"strconv"
	"strings"
)

var (
	minaDeviceID string
	minaCmd      = &cobra.Command{
		Use:   "mina <list|tts|player> <keyword|message|play|pause|set_volume>",
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
				}
			} else {
				res, err = list(srv, "")
			}

			handleResult(res, err)
		},
	}
)

func init() {
	minaCmd.Example = "  mina list 小爱 \n  mina tts message \n  mina player play\n  mina player pause\n  mina player volume 50\n  mina player volume +10\n  mina player volume -1"
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
	table := uitable.New()
	table.MaxColWidth = 80
	table.Wrap = true // wrap columns
	for i, device := range devices {
		table.AddRow("")
		table.AddRow("Index:", i+1)
		table.AddRow("Name:", device.Name)
		table.AddRow("DeviceID:", device.DeviceID)
		table.AddRow("Presence:", device.Presence)
		table.AddRow("MiotDID:", device.MiotDID)
		table.AddRow("")
	}
	res = table
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
