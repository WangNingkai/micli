package cmd

import (
	"errors"

	"micli/pkg/miservice"

	"github.com/spf13/cobra"
)

var minaTtsCmd = &cobra.Command{
	Use:   "tts <text>",
	Short: "Text to speech",
	Long:  `Text to speech`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			res interface{}
			err error
		)
		if len(args) > 0 {
			res, err = doTTS(minaSrv, args)
		} else {
			err = errors.New("tts message is empty")
		}
		handleResult(res, err)
	},
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

func init() {
	minaTtsCmd.Example = "  mina tts \"hello world\""
}
