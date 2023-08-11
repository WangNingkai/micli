package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"micli/conf"
	"micli/pkg/miservice"
	"strconv"
	"time"
)

var (
	minaRecordsCmd = &cobra.Command{
		Use:   "records <?limit>",
		Short: "Get records",
		Long:  `Get records`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				res interface{}
				err error
			)
			res, err = askRecords(minaSrv, args)
			handleResult(res, err)
		},
	}
)

func askRecords(srv *miservice.MinaService, args []string) (res interface{}, err error) {
	var limit int
	if len(args) > 0 {
		limit, _ = strconv.Atoi(args[0])
	} else {
		limit = 10
	}

	deviceId := minaDeviceID
	if deviceId == "" {
		deviceId = conf.Cfg.Section("mina").Key("DID").MustString("")
	}
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

func init() {
	minaRecordsCmd.Example = "  mina records 10"
}
