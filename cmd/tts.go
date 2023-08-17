package cmd

import (
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"micli/internal/conf"
	"micli/pkg/tts"
	"micli/pkg/tts/edgetts"
)

var (
	text   string
	voice  string
	ttsCmd = &cobra.Command{
		Use:   "tts",
		Short: "Text To Speech",
		Long:  `Use Microsoft Edge's online text-to-speech service WITHOUT needing Microsoft Edge or Windows or an API key`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				fp  string
				err error
			)
			pterm.Debug.Println("text:", text)

			if voice == "" {
				voiceMap := make(map[string]string)
				var voiceList []*edgetts.Voice
				voiceList, err = tts.LoadVoiceList()
				if err != nil {
					return
				}
				choices := make([]string, len(voiceList))
				for i, _voice := range voiceList {
					choice := fmt.Sprintf("%s(%s) - %s", _voice.ShortName, _voice.Gender, _voice.Locale)
					voiceMap[choice] = _voice.ShortName
					choices[i] = choice
				}
				choice, _ := pterm.DefaultInteractiveSelect.
					WithDefaultText("Please select a voice").
					WithOptions(choices).
					Show()
				pterm.Info.Println("Choose voice: " + choice)
				voice = voiceMap[choice]

			}

			fp, err = tts.TextToMp3(text, voice)
			if err != nil {
				return
			}
			client := req.C()
			r := client.R()
			r.SetFile("file", fp)
			var resp *req.Response
			resp, err = r.Put(fmt.Sprintf("%s/edge_tts.mp3", conf.Cfg.Section("file").Key("TRANSFER_SH").MustString("https://transfer.sh")))
			textUrl := resp.String()
			pterm.Success.Println("tts url:", textUrl)
		},
	}
)

func init() {
	ttsCmd.Flags().StringVarP(&text, "text", "t", "", "text(required)")
	ttsCmd.Flags().StringVarP(&voice, "voice", "v", "", "voice")
	_ = ttsCmd.MarkFlagRequired("text")
	ttsCmd.Example = `  micli tts -t "你好，世界"`
}
