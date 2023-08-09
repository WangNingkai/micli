package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"micli/miservice"
	"strings"
)

var (
	minaListCmd = &cobra.Command{
		Use:   "list <?keyword>",
		Short: "List devices",
		Long:  `List devices`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				keyword string
				res     interface{}
				err     error
			)
			srv := miservice.NewMinaService(miAccount)
			if len(args) > 0 {
				keyword = args[0]
			}
			res, err = list(srv, keyword)
			handleResult(res, err)
		},
	}
)

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

func init() {
	minaListCmd.Example = "  mina list 小爱"
}
