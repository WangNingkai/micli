package cmd

import (
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"micli/conf"
	"micli/pkg/jarvis"
	"micli/pkg/miservice"
	"micli/pkg/tts"
	"micli/pkg/util"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	commandPath         = "./commands.json"
	mute                bool
	HardwareCommandDict = map[string][3]string{
		//hardware: (tts_command, wakeup_command,execute_command)
		"LX06": {"5-1", "5-3", "5-5"},
		"L05B": {"5-3", "5-1", "5-4"},
		"LX01": {"5-1", "5-2", "5-5"},
		"L06A": {"5-1", "5-2", "5-5"},
		"LX04": {"5-1", "5-2", "5-4"},
		"L05C": {"5-3", "5-1", "5-4"},
		"L17A": {"7-3", "7-1", "7-4"},
		"X08E": {"7-3", "7-1", "7-4"},
		"LX5A": {"5-1", "5-3", "5-5"},
		"L15A": {"7-3", "7-1", "7-4"},
		"X6A":  {"7-3", "7-1", "7-4"},
		"L7A":  {"5-1", "5-2", "5-5"},
		"X08C": {"3-1", "3-2", "3-5"},
		"X08A": {"5-1", "5-2", "5-4"},
		"L09A": {"3-1", "3-2", "3-5"},
		"LX05": {"5-1", "5-3", "5-5"},
		"L04M": {"5-1", "5-2", "5-4"},
		"L09B": {"7-3", "7-1", "7-4"},
		// add more here
	}
	minaServeCmd = &cobra.Command{
		Use:   "serve",
		Short: "Hack xiaoai Project",
		Long:  `Hack xiaoai Project`,
		Run: func(cmd *cobra.Command, args []string) {
			s := NewServe()
			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
			go func() {
				for sig := range c {
					switch sig {
					case syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT:
						pterm.Success.Println("serve stopped.")
						os.Exit(0)
					}
				}
			}()
			if err := s.Run(); err != nil {
				pterm.Fatal.Println(err)
			}
		},
	}
)

func init() {
	minaServeCmd.Flags().BoolVarP(&mute, "mute", "m", true, "mute")
}

type Command struct {
	Keyword string `json:"keyword"`
	Mute    bool   `json:"mute"`
	Type    string `json:"type"`
	Config  struct {
		Stream       bool   `json:"stream"`
		UseEdgeTTS   bool   `json:"useEdgeTTS"`
		EdgeTTSVoice string `json:"edgeTTSVoice"`
		UseCmd       bool   `json:"useCmd"`
		Wait         bool   `json:"wait"`
		EndFlag      string `json:"endFlag"`
	} `json:"config,omitempty"`
	Step []struct {
		Type    string `json:"type"`
		Request struct {
			URL          string `json:"url"`
			Data         string `json:"data"`
			Method       string `json:"method"`
			Headers      string `json:"headers"`
			Out          string `json:"out"`
			UseCmd       bool   `json:"useCmd"`
			Wait         bool   `json:"wait"`
			UseEdgeTTS   bool   `json:"useEdgeTTS"`
			EdgeTTSVoice string `json:"edgeTTSVoice"`
		} `json:"request,omitempty"`
		Action struct {
			Text   string `json:"text"`
			Silent bool   `json:"silent"`
		} `json:"action,omitempty"`
		TTS struct {
			Out          string `json:"out"`
			UseCmd       bool   `json:"useCmd"`
			Wait         bool   `json:"wait"`
			UseEdgeTTS   bool   `json:"useEdgeTTS"`
			EdgeTTSVoice string `json:"edgeTTSVoice"`
		} `json:"tts,omitempty"`
	} `json:"step,omitempty"`
}
type Serve struct {
	LastTimestamp  int64
	InConversation bool

	records  chan *miservice.AskRecordItem
	commands []*Command

	minaSrv *miservice.MinaService
	miioSrv *miservice.IOService

	device *miservice.DeviceData

	inChat      bool // 是否在对话中
	chatOptions *Command
}

func NewServe() *Serve {
	return &Serve{
		records: make(chan *miservice.AskRecordItem),
		minaSrv: minaSrv,
		miioSrv: ioSrv,
	}
}

func (s *Serve) loadCommands() (err error) {
	var (
		f *os.File
	)
	if !util.Exists(commandPath) {
		err = errors.New("not found commands.json")
		return
	}
	f, err = os.Open(commandPath)
	if err != nil {
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	j := json.NewDecoder(f)
	var commands []*Command
	err = j.Decode(&commands)
	if err != nil {
		return
	}
	s.commands = commands
	return
}

func (s *Serve) pollLatestAsk() {
	var err error
	for {
		start := time.Now()
		pterm.Debug.Println("Poll Latest Ask At " + start.Format("2006-01-02 15:04:05"))
		var resp *miservice.AskRecords
		err = s.minaSrv.LastAskList(s.device.DeviceID, s.device.Hardware, 2, &resp)
		if err != nil {
			pterm.Error.Println(err.Error())
			continue
		}
		if resp == nil {
			continue
		}
		var record *miservice.AskRecord
		err = json.Unmarshal([]byte(resp.Data), &record)
		if err != nil {
			pterm.Error.Println(err.Error())
			continue
		}
		if len(record.Records) > 0 {
			r := record.Records[0]
			if time.UnixMilli(record.Records[0].Time).After(time.UnixMilli(s.LastTimestamp)) {
				s.LastTimestamp = r.Time
				s.records <- r
			}
		}
		elapsed := time.Since(start)
		if elapsed < time.Second {
			time.Sleep(time.Second - elapsed)
		}
	}
}

func (s *Serve) wakeup() error {
	v, ok := HardwareCommandDict[s.device.Hardware]
	if !ok {
		return errors.New("not found hardware command")
	}
	_cmd := v[1]
	siid, aiid := util.TwinsSplit(_cmd, "-", "1")
	sid, _ := strconv.Atoi(siid)
	aid, _ := strconv.Atoi(aiid)
	_, err := s.miioSrv.MiotAction(s.device.MiotDID, []int{sid, aid}, nil)
	return err
}

func (s *Serve) stop() error {
	/*var (
		yes bool
		err error
	)
	if yes, err = s.isPlaying(); yes {
		_, err = s.minaSrv.PlayerPause(minaDeviceID)
		return err
	}*/
	_, err := s.minaSrv.PlayerPause(minaDeviceID)
	return err

}

func (s *Serve) isPlaying() (bool, error) {
	res, err := s.minaSrv.PlayerGetStatus(minaDeviceID)
	if err != nil {
		return false, err
	}
	type info struct {
		Status int `json:"status"`
		Volume int `json:"volume"`
	}
	var dataInfo *info
	err = json.Unmarshal([]byte(res.Data.Info), &dataInfo)
	if err != nil {
		return false, err
	}
	return dataInfo.Status == 1, nil
}

func (s *Serve) tts(useCommand bool, message string, wait bool) (err error) {
	if useCommand {
		v, ok := HardwareCommandDict[s.device.Hardware]
		if !ok {
			err = errors.New("not found hardware command")
			return
		}
		_cmd := v[0]
		siid, aiid := util.TwinsSplit(_cmd, "-", "1")
		sid, _ := strconv.Atoi(siid)
		aid, _ := strconv.Atoi(aiid)
		_, err = s.miioSrv.MiotAction(s.device.MiotDID, []int{sid, aid}, []interface{}{message})
	} else {
		_, err = s.minaSrv.TextToSpeech(s.device.DeviceID, message)
	}
	if err != nil {
		return
	}
	if wait {
		elapse := util.CalculateTTSElapse(message)
		time.Sleep(elapse)
		s.waitForTTSDone()
	}
	return
}

func (s *Serve) edgeTTS(voice string, message string, wait bool) (err error) {
	var fp string
	if voice == "" {
		voice = "zh-CN-XiaoxiaoNeural"
	}
	fp, err = tts.TextToMp3(message, voice)
	if err != nil {
		return
	}
	client := req.C()
	r := client.R()
	r.SetFile("file", fp)
	resp, err := r.Put(fmt.Sprintf("%s/edge_tts.mp3", conf.Cfg.Section("file").Key("TRANSFER_SH").MustString("https://transfer.sh")))
	textUrl := resp.String()
	pterm.Debug.Println("Play Audio URL: ", textUrl)
	_, err = s.minaSrv.PlayByUrl(s.device.DeviceID, textUrl)
	if err != nil {
		return
	}
	if wait {
		elapse := util.CalculateTTSElapse(message)
		time.Sleep(elapse)
		s.waitForTTSDone()
	}

	return
}

func (s *Serve) Call(text string, silent bool) (err error) {
	v, ok := HardwareCommandDict[s.device.Hardware]
	if !ok {
		err = errors.New("not found hardware command")
		return
	}
	_cmd := v[2]
	siid, aiid := util.TwinsSplit(_cmd, "-", "1")
	sid, _ := strconv.Atoi(siid)
	aid, _ := strconv.Atoi(aiid)
	var silentInt = 0
	if silent {
		silentInt = 1
	}
	_, err = s.miioSrv.MiotAction(s.device.MiotDID, []int{sid, aid}, []interface{}{text, silentInt})
	return
}

func (s *Serve) waitForTTSDone() {
	for {
		isPlaying, _ := s.isPlaying()
		if !isPlaying {
			break
		}
		time.Sleep(200 * time.Microsecond)
	}
}

func (s *Serve) Run() error {
	err := s.loadCommands() // 加载训练计划
	if err != nil {
		pterm.Error.Println(err.Error())
		return err
	}
	deviceId := minaDeviceID
	if deviceId == "" {
		deviceId = conf.Cfg.Section("mina").Key("DID").MustString("")
	}
	var device *miservice.DeviceData
	device, err = chooseMinaDeviceDetail(s.minaSrv, deviceId)
	s.device = device
	s.inChat = false
	pterm.Info.Println("Start Listen Device: " + device.Name)
	s.LastTimestamp = time.Now().UnixMilli()
	go s.pollLatestAsk()
	for {
		record := <-s.records
		pterm.Debug.Println("Latest Query: ", record.Query)
		query := record.Query
		if s.inChat {
			if s.chatOptions.Mute {
				_ = s.stop()
			} else {
				s.waitForTTSDone()
			}
			if strings.Contains(query, s.chatOptions.Config.EndFlag) {
				pterm.Debug.Println("End Chat")
				s.inChat = false
				waitMsg := "再见～"
				if s.chatOptions.Config.UseEdgeTTS {
					_ = s.edgeTTS(s.chatOptions.Config.EdgeTTSVoice, waitMsg, s.chatOptions.Config.Wait)
				} else {
					_ = s.tts(s.chatOptions.Config.UseCmd, waitMsg, s.chatOptions.Config.Wait)
				}
				if err != nil {
					pterm.Error.Println(err.Error())
				}
				_ = s.stop()
				continue
			}
			waitMsg := "让我想想，请耐心等待哦～"
			if s.chatOptions.Config.UseEdgeTTS {
				_ = s.edgeTTS(s.chatOptions.Config.EdgeTTSVoice, waitMsg, s.chatOptions.Config.Wait)
			} else {
				_ = s.tts(s.chatOptions.Config.UseCmd, waitMsg, s.chatOptions.Config.Wait)
			}
			if len(record.Answers) > 0 {
				var str string
				for _, a := range record.Answers {
					str += a.Tts.Text
				}
				pterm.Debug.Println("xiaoai's answer: ", str)
			} else {
				pterm.Debug.Println("xiaoai's answer: ", "No Answer")
			}
			gpt := jarvis.NewChatGPT()
			if s.chatOptions.Config.Stream {
				gpt.StreamMessage = make(chan string)
				gpt.InStream = true
				go func() {
					err = gpt.AskStream(query)
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
					pterm.Debug.Println("jarvis's answer: ", sentence)
					if s.chatOptions.Config.UseEdgeTTS {
						_ = s.edgeTTS(s.chatOptions.Config.EdgeTTSVoice, sentence, s.chatOptions.Config.Wait)
					} else {
						_ = s.tts(s.chatOptions.Config.UseCmd, sentence, s.chatOptions.Config.Wait)
					}
				}
			} else {
				var sentence string
				sentence, err = gpt.Ask(query)
				if err != nil {
					pterm.Error.Println(err.Error())
				}
				if s.chatOptions.Config.UseEdgeTTS {
					_ = s.edgeTTS(s.chatOptions.Config.EdgeTTSVoice, sentence, s.chatOptions.Config.Wait)
				} else {
					_ = s.tts(s.chatOptions.Config.UseCmd, sentence, s.chatOptions.Config.Wait)
				}
				pterm.Debug.Println("jarvis's answer: ", sentence)
			}
			pterm.Debug.Println("continue chat")
			_ = s.wakeup()
			continue
		} else {
			command, match := lo.Find(s.commands, func(c *Command) bool { return strings.Contains(query, c.Keyword) })
			if !match {
				continue
			}
			if command.Mute {
				_ = s.stop()
			} else {
				s.waitForTTSDone()
			}
			if command.Type == "chat" {
				pterm.Debug.Println("Start Chat")
				s.inChat = true
				s.chatOptions = command
				waitMsg := "有什么可以帮您的吗？"
				if s.chatOptions.Config.UseEdgeTTS {
					_ = s.edgeTTS(s.chatOptions.Config.EdgeTTSVoice, waitMsg, s.chatOptions.Config.Wait)
				} else {
					_ = s.tts(s.chatOptions.Config.UseCmd, waitMsg, s.chatOptions.Config.Wait)
				}
				_ = s.wakeup()
				continue
			} else {
				for _, step := range command.Step {
					switch step.Type {
					case "tts":
						pterm.Debug.Println("Start execute tts")
						if step.TTS.UseEdgeTTS {
							err = s.edgeTTS(step.TTS.EdgeTTSVoice, step.TTS.Out, step.TTS.Wait)
						} else {
							err = s.tts(step.TTS.UseCmd, step.TTS.Out, step.TTS.Wait)
						}
						if err != nil {
							pterm.Error.Println(err.Error())
						}
					case "request":
						pterm.Debug.Println("Start execute request")
						if step.Request.Method == "" {
							step.Request.Method = "GET"
						}
						if step.Request.URL == "" {
							continue
						}
						client := req.C()
						r := client.R()
						if step.Request.Headers != "" {
							var headers map[string]string
							err = json.Unmarshal([]byte(step.Request.Headers), &headers)
							if err != nil {
								pterm.Error.Println(err.Error())
							}
							req.SetHeaders(headers)
						}
						if step.Request.Data != "" {
							var data map[string]interface{}
							err = json.Unmarshal([]byte(step.Request.Data), &data)
							if err != nil {
								pterm.Error.Println(err.Error())
							}
							req.SetBody(data)
						}
						var resp *req.Response
						pterm.Info.Println(step.Request.Method + " " + step.Request.URL)
						resp, err = r.Send(step.Request.Method, step.Request.URL)
						if err != nil {
							pterm.Error.Println(err.Error())
						} else {
							if step.Request.Out != "" {
								if strings.Contains(step.Request.Out, ".") {
									value := gjson.Get(resp.String(), step.Request.Out)
									message := value.String()
									if step.TTS.UseEdgeTTS {
										err = s.edgeTTS(step.Request.EdgeTTSVoice, step.Request.Out, step.Request.Wait)
									} else {
										err = s.tts(step.Request.UseCmd, message, step.Request.Wait)
									}
								} else {
									if step.Request.UseEdgeTTS {
										err = s.edgeTTS(step.Request.EdgeTTSVoice, step.Request.Out, step.Request.Wait)
									} else {
										err = s.tts(step.Request.UseCmd, step.Request.Out, step.Request.Wait)
									}
								}
								if err != nil {
									pterm.Error.Println(err.Error())
								}
							}
						}
					case "action":
						pterm.Debug.Println("Start execute action")
						err = s.Call(step.Action.Text, step.Action.Silent)
						if err != nil {
							pterm.Error.Println(err.Error())
						}
					}
				}
			}
		}

	}
}
