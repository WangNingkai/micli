package cmd

import (
	"micli/pkg/jarvis"
	"micli/pkg/util"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test",
	Long:  "Test",
	Run: func(cmd *cobra.Command, args []string) {
		gpt := jarvis.NewChatGPT()
		gpt.StreamMessage = make(chan string)
		gpt.InStream = true
		go func() {
			err := gpt.AskStream("你是谁")
			if err != nil {
				pterm.Error.Println(err.Error())
			}
		}()
		reply := util.SplitSentences(gpt.StreamMessage)
		for {
			sentence, ok := <-reply
			if !ok {
				gpt.InStream = false
				break
			}
			pterm.Debug.Println(sentence)
		}
		/*reply, err := gpt.Ask("你是谁")
		if err != nil {
			pterm.Error.Println(err.Error())
		}

		pterm.Info.Println(reply)*/
	},
}
