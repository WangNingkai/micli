package cmd

import (
	"errors"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"micli/conf"
	"micli/pkg/miservice"
	"strconv"
)

var (
	minaPlayerCmd = &cobra.Command{
		Use:   "player <play|pause|volume|status> <?arg2>",
		Short: "Player",
		Long:  `Player`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			res, err = operatePlayer(minaSrv, args)
			handleResult(res, err)
		},
	}
)

// operatePlayer 播放器操作
func operatePlayer(srv *miservice.MinaService, args []string) (res interface{}, err error) {
	deviceId := minaDeviceID
	if deviceId == "" {
		deviceId = conf.Cfg.Section("mina").Key("DID").MustString("")
		if deviceId == "" {
			deviceId, err = chooseMinaDevice(srv)
			if err != nil {
				return
			}
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

func init() {
	minaPlayerCmd.Example = "  mina player status \n  mina player play\n  mina player pause\n  mina player volume 50"
}
