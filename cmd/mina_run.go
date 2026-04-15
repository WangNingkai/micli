package cmd

import (
	"errors"
	"strconv"

	"micli/pkg/miservice"
	"micli/pkg/util"

	"github.com/spf13/cobra"
)

var minaRunSilent bool

var minaRunCmd = &cobra.Command{
	Use:   "run <text>",
	Short: "Execute natural language command via XiaoAi speaker",
	Long:  `Execute natural language command (like "打开卧室台灯") via XiaoAi speaker's execute-text-directive action`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			res interface{}
			err error
		)
		if len(args) > 0 {
			res, err = doRun(minaSrv, ioSrv, args[0], minaRunSilent)
		} else {
			err = errors.New("command text is empty")
		}
		handleResult(res, err)
	},
}

func doRun(minaSrv *miservice.MinaService, ioSrv *miservice.IOService, text string, silent bool) (res interface{}, err error) {
	device, err := chooseMinaDeviceDetail(minaSrv, minaDeviceID)
	if err != nil {
		return
	}

	v, ok := HardwareCommandDict[device.Hardware]
	if !ok {
		err = errors.New("unsupported hardware: " + device.Hardware)
		return
	}

	executeCmd := v[2]
	siid, aiid := util.TwinsSplit(executeCmd, "-", "1")
	sid, _ := strconv.Atoi(siid)
	aid, _ := strconv.Atoi(aiid)

	silentInt := 0
	if silent {
		silentInt = 1
	}

	res, err = ioSrv.MiotAction(device.MiotDID, []int{sid, aid}, []interface{}{text, silentInt})
	return
}

func init() {
	minaRunCmd.Flags().BoolVarP(&minaRunSilent, "silent", "s", false, "silent execution (no voice response from speaker)")
	minaRunCmd.Example = `  micli mina run "打开卧室台灯"
  micli mina run "把亮度调到50%" -d <device_id>
  micli mina run "关闭所有灯" --silent`
}
