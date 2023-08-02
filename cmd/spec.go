package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"micli/miservice"
	"strings"
)

var (
	specCmd = &cobra.Command{
		Use:   "spec [model_keyword|type_urn]",
		Short: "MIoT Spec",
		Long:  `MIoT Spec`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				err     error
				keyword string
			)
			argLen := len(args)
			if argLen > 0 {
				keyword = args[0]
			} else {
				deviceMap := make(map[string]string)
				var devices []*miservice.DeviceInfo
				devices, err = getDeviceListFromLocal()
				if err != nil {
					handleResult(nil, err)
					return
				}
				choices := make([]string, len(devices))
				for i, device := range devices {
					choice := fmt.Sprintf("%s - %s", device.Name, device.Did)
					deviceMap[choice] = device.Model
					choices[i] = choice
				}
				choice, _ := pterm.DefaultInteractiveSelect.
					WithDefaultText("Please select a device").
					WithOptions(choices).
					Show()
				pterm.Info.Println("You choose: " + choice)
				keyword = deviceMap[choice]
			}
			var data *miservice.MiotSpecInstancesData
			data, err = srv.MiotSpec(keyword)
			if err != nil {
				handleResult(nil, err)
				return
			}
			// https://miot-spec.org/miot-spec-v2/spec/service?type=
			// https://miot-spec.org/miot-spec-v2/spec/action?type=
			// https://miot-spec.org/miot-spec-v2/spec/property?type=
			if len(data.Services) > 0 {
				typeValue := data.Type
				info := fmt.Sprintf("# Generated by https://github.com/wangningkai/micli\n# More Detail: https://home.miot-spec.com/spec?type=%s\n# Json: https://miot-spec.org/miot-spec-v2/instance?type=%s", typeValue, typeValue)
				pterm.Info.Println(info)
				var leveledList pterm.LeveledList
				for _, service := range data.Services {
					piids := make(map[interface{}]string, len(service.Properties))
					leveledList = append(leveledList, pterm.LeveledListItem{Level: 0, Text: fmt.Sprintf("%s (siid:%d)", service.Description, service.Iid)})
					if len(service.Properties) > 0 {
						leveledList = append(leveledList, pterm.LeveledListItem{Level: 1, Text: "Properties"})
						for _, property := range service.Properties {
							if property.Description == "" {
								types := strings.Split(property.Type, ":")
								pterm.Info.Println(types)
								if len(types) > 3 {
									property.Description = strings.Title(strings.ReplaceAll(types[3], "-", " "))
								}
							}
							piids[property.Iid] = property.Description
							detail := fmt.Sprintf("%s (piid:%d,format:%s,access:[%s])", property.Description, property.Iid, property.Format, strings.Join(property.Access, ","))
							leveledList = append(leveledList, pterm.LeveledListItem{Level: 2, Text: detail})
							if property.ValueRange != nil {
								min := property.ValueRange[0]
								max := property.ValueRange[1]
								step := property.ValueRange[2]
								rangeData := fmt.Sprintf("range:[%v,%v],step:%v", min, max, step)
								leveledList = append(leveledList, pterm.LeveledListItem{Level: 3, Text: rangeData})
							}
							if property.ValueList != nil {
								for _, value := range property.ValueList {
									leveledList = append(leveledList, pterm.LeveledListItem{Level: 3, Text: fmt.Sprintf("%d-%s", value.Value, value.Description)})
								}
							}
						}
						if service.Actions != nil {
							leveledList = append(leveledList, pterm.LeveledListItem{Level: 1, Text: "Actions"})
							for _, action := range service.Actions {
								leveledList = append(leveledList, pterm.LeveledListItem{Level: 2, Text: fmt.Sprintf("%s (aiid:%d)", action.Description, action.Iid)})
								if action.In != nil {
									for _, in := range action.In {
										p, _ := lo.Find(service.Properties, func(property *miservice.MiotSpecProperty) bool { return property.Iid == int(in) })
										leveledList = append(leveledList, pterm.LeveledListItem{Level: 3, Text: fmt.Sprintf("%v-%v", in, p.Description)})
									}
								}
							}
						}
					}
				}
				root := putils.TreeFromLeveledList(leveledList)
				root.Text = fmt.Sprintf("Miot Spec [%s]", data.Description)
				err = pterm.DefaultTree.WithRoot(root).Render()
				if err != nil {
					handleResult(nil, err)
					return
				}
			}
		},
	}
)

func init() {
	specCmd.Example = "  spec\n  spec xiaomi.wifispeaker.lx06\n  spec urn:miot-spec-v2:device:speaker:0000A015:xiaomi-lx06:1"
}
