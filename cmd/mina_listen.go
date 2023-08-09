package cmd

import (
	"errors"
	"github.com/imroc/req/v3"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"micli/miservice"
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
	minaListenCmd = &cobra.Command{
		Use:   "listen",
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
						pterm.Success.Println("serve stopped")
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
	minaListenCmd.Flags().BoolVarP(&mute, "mute", "m", true, "mute")
}

type Command struct {
	Keyword string `json:"keyword"`
	Step    []struct {
		Type          string `json:"type"`
		URL           string `json:"url"`
		Data          string `json:"data"`
		Method        string `json:"method"`
		Headers       string `json:"headers"`
		Out           string `json:"out"`
		Silent        bool   `json:"silent"`
		UseTTSCommand bool   `json:"useTTSCmd"`
		Wait          bool   `json:"wait"`
		ActionText    string `json:"actionText"`
	} `json:"step"`
}
type Serve struct {
	LastTimestamp  int64
	InConversation bool

	records  chan *miservice.AskRecordItem
	commands []*Command

	minaSrv *miservice.MinaService
	miioSrv *miservice.IOService

	device *miservice.DeviceData
}

func NewServe() *Serve {
	return &Serve{
		records: make(chan *miservice.AskRecordItem),
		minaSrv: miservice.NewMinaService(miAccount),
		miioSrv: miservice.NewIOService(miAccount),
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
	defer f.Close()
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
		pterm.Info.Println("Poll Latest Ask At " + start.Format("2006-01-02 15:04:05"))
		var resp *miservice.AskRecords
		err = s.minaSrv.LastAskList(s.device.DeviceID, s.device.Hardware, 2, &resp)
		if err != nil {
			pterm.Error.Println(err)
			continue
		}
		if resp == nil {
			continue
		}
		var record *miservice.AskRecord
		err = json.Unmarshal([]byte(resp.Data), &record)
		if err != nil {
			pterm.Error.Println(err)
			continue
		}
		if len(record.Records) > 0 {
			r := record.Records[0]
			if r.Time > s.LastTimestamp {
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
	var (
		yes bool
		err error
	)
	if yes, err = s.isPlaying(); yes {
		_, err = s.minaSrv.PlayerPause(minaDeviceID)
		return err
	}
	return nil

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
		time.Sleep(1 * time.Second)
	}
}

func (s *Serve) Run() error {
	err := s.loadCommands()
	if err != nil {
		pterm.Error.Println(err)
		return err
	}
	s.LastTimestamp = time.Now().UnixMilli()
	devices, err := s.minaSrv.DeviceList(0)
	if err != nil {
		return err
	}
	device, ok := lo.Find(devices, func(d *miservice.DeviceData) bool { return d.DeviceID == minaDeviceID })
	if !ok {
		device = devices[0]
	}
	s.device = device
	pterm.Info.Println("Start Listen Device: " + device.Name)
	go s.pollLatestAsk()
	for {
		record := <-s.records
		pterm.Info.Println(record.Query)
		query := record.Query
		command, match := lo.Find(s.commands, func(c *Command) bool { return strings.Contains(query, c.Keyword) })
		if !match {
			continue
		}
		if mute {
			err = s.stop()
			if err != nil {
				pterm.Error.Println(err)
				continue
			}
		} else {
			s.waitForTTSDone()
		}
		for _, step := range command.Step {
			switch step.Type {
			case "tts":
				pterm.Info.Println("execute tts")
				err = s.tts(step.UseTTSCommand, step.Out, step.Wait)
				if err != nil {
					pterm.Error.Println(err)
				}
			case "request":
				pterm.Info.Println("execute request")
				if step.Method == "" {
					step.Method = "GET"
				}
				if step.URL == "" {
					continue
				}
				client := req.C()
				r := client.R()
				if step.Headers != "" {
					var headers map[string]string
					err = json.Unmarshal([]byte(step.Headers), &headers)
					if err != nil {
						pterm.Error.Println(err)
					}
					req.SetHeaders(headers)
				}
				if step.Data != "" {
					var data map[string]interface{}
					err = json.Unmarshal([]byte(step.Data), &data)
					if err != nil {
						pterm.Error.Println(err)
					}
					req.SetBody(data)
				}
				var resp *req.Response
				pterm.Info.Println(step.Method + " " + step.URL)
				resp, err = r.Send(step.Method, step.URL)
				if err != nil {
					pterm.Error.Println(err)
				} else {
					if step.Out != "" {
						value := gjson.Get(resp.String(), step.Out)
						message := value.String()
						err = s.tts(step.UseTTSCommand, message, step.Wait)
						if err != nil {
							pterm.Error.Println(err)
						}
					}
				}
			case "action":
				pterm.Info.Println("execute action")
				err = s.Call(step.Out, step.Silent)
				if err != nil {
					pterm.Error.Println(err)
				}
			}
		}
	}
}