package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"micli/internal/conf"
	"micli/pkg/miservice"
	"micli/pkg/util"

	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var propsGetCmd = &cobra.Command{
	Use:   "get <siid[-piid]>[,...]",
	Short: "MIoT Properties Get",
	Long:  `MIoT Properties Get`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			res         interface{}
			err         error
			deviceModel string
		)
		if did == "" {
			did = conf.Cfg.Section("account").Key("MI_DID").MustString("")
			if did == "" {
				did, err = chooseDevice()
				if err != nil {
					pterm.Error.Println(err.Error())
					return
				}
			}
		}
		var devices []*miservice.DeviceInfo
		if !util.IsDigit(did) {
			devices, err = getDeviceListFromLocal() // Implement this method for the IOService
			if err != nil {
				pterm.Error.Println(err.Error())
				return
			}
			if len(devices) == 0 {
				err = fmt.Errorf("no device found")
				pterm.Error.Println(err.Error())
				return
			}
			for _, device := range devices {
				if device.Name == did {
					did = device.Did
					deviceModel = device.Model
					break
				}
			}
		} else {
			devices, err = getDeviceListFromLocal() // Implement this method for the IOService
			if err != nil {
				pterm.Error.Println(err.Error())
				return
			}
			device, _ := lo.Find(devices, func(d *miservice.DeviceInfo) bool { return d.Did == did })
			if device != nil {
				deviceModel = device.Model
			}
		}
		var specs *miservice.MiotSpecInstancesData
		specs, err = ioSrv.MiotSpec(deviceModel)
		if err != nil {
			pterm.Error.Println(err.Error())
			return
		}
		if len(specs.Services) == 0 {
			err = fmt.Errorf("no service found")
			pterm.Error.Println(err.Error())
			return
		}

		miot := true
		var props [][]interface{}

		var _args []string
		if len(args) > 0 {
			_args = strings.Split(args[0], ",")
		}
		title := specs.Description
		var descs [][]interface{}
		for _, item := range _args {
			siid, iid := util.TwinsSplit(item, "-", "1")
			var prop []interface{}
			var desc []interface{}
			if util.IsDigit(siid) && util.IsDigit(iid) {
				s, _ := strconv.Atoi(siid)
				_service, _ := lo.Find(specs.Services, func(srv *miservice.MiotSpecService) bool { return srv.Iid == s })
				if _service == nil {
					err = fmt.Errorf("service not found")
					pterm.Error.Println(err.Error())
					return
				}
				i, _ := strconv.Atoi(iid)
				_prop, _ := lo.Find(_service.Properties, func(pr *miservice.MiotSpecProperty) bool { return pr.Iid == i })
				if _prop == nil {
					err = fmt.Errorf("property not found")
					pterm.Error.Println(err.Error())
					return
				}
				prop = []interface{}{s, i}
				desc = []interface{}{_service.Description, _prop.Description}
			} else {
				prop = []interface{}{item}
				miot = false
			}
			props = append(props, prop)
			descs = append(descs, desc)
		}
		var data []interface{}
		if miot {
			data, err = ioSrv.MiotGetProps(did, props)
		} else {
			/*var _props []string
			for _, prop := range props {
				_props = append(_props, prop[0].(string))
			}
			res, err = ioSrv.HomeGetProps(did, _props)*/
			err = errors.New("device not support miot")
		}

		var items []pterm.BulletListItem
		items = append(items, pterm.BulletListItem{
			Level:     0,
			TextStyle: pterm.NewStyle(pterm.FgRed),
			Text:      title,
		})
		for i, item := range data {
			desc := descs[i]
			items = append(items, pterm.BulletListItem{
				Level:     1,
				TextStyle: pterm.NewStyle(pterm.FgCyan),
				Text:      fmt.Sprintf("Service: %s", pterm.Cyan(desc[0])),
			})
			items = append(items, pterm.BulletListItem{
				Level:  2,
				Text:   fmt.Sprintf("Prop: %s", pterm.Green(desc[1])),
				Bullet: "-",
			})
			items = append(items, pterm.BulletListItem{
				Level: 2,

				Text:   fmt.Sprintf("Value: %v", pterm.Green(item)),
				Bullet: "-",
			})
		}
		err = pterm.DefaultBulletList.WithItems(items).Render()
		handleResult(res, err)
	},
}

func init() {
	propsGetCmd.Flags().StringVarP(&did, "did", "d", "", "Device ID")
	propsGetCmd.Example = "  get 1,1-2,1-3,1-4,2-1,2-2,3"
}
