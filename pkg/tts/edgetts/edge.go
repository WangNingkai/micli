package edgetts

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/imroc/req/v3"
	"github.com/pterm/pterm"
	uuid "github.com/satori/go.uuid"
)

const (
	edgeWssUrl         = `wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1?TrustedClientToken=6A5AA1D4EAFF4E9FB37E23D68491D6F4&ConnectionId=`
	voiceListUrl       = "https://speech.platform.bing.com/consumer/speech/synthesize/readaloud/voices/list?trustedclienttoken=6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	NormalSsmlTemplate = `
<speak xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" xmlns:emo="http://www.w3.org/2009/10/emotionml" version="1.0" xml:lang="en-US">
    <voice name="{voiceName}">
      <prosody rate="0%" pitch="0%">
          {text}
      </prosody >
    </voice >
</speak >`
	voiceFormat = "audio-24khz-48kbitrate-mono-mp3"
)

var edgeChinaIpList = []string{
	// 北京微软云
	"202.89.233.100",
	"202.89.233.101",
	"202.89.233.102",
	"202.89.233.103",
	"202.89.233.104",

	//"182.61.148.24", 广东百度云
}

type Voice struct {
	Name           string `json:"Name"`
	ShortName      string `json:"ShortName"`
	Gender         string `json:"Gender"`
	Locale         string `json:"Locale"`
	SuggestedCodec string `json:"SuggestedCodec"`
	FriendlyName   string `json:"FriendlyName"`
	Status         string `json:"Status"`
	VoiceTag       struct {
		ContentCategories  []string `json:"ContentCategories"`
		VoicePersonalities []string `json:"VoicePersonalities"`
	} `json:"VoiceTag"`
}

type EdgeTTS struct {
	DnsLookupEnabled bool // 使用DNS解析，而不是北京微软云节点。
	DialTimeout      time.Duration
	WriteTimeout     time.Duration

	dialContextCancel context.CancelFunc

	uuid          string
	conn          *websocket.Conn
	onReadMessage TReadMessage
}

type TReadMessage func(messageType int, p []byte, errMessage error) (finished bool)

func (t *EdgeTTS) NewConn() error {
	pterm.Debug.Println("创建WebSocket连接(Edge)...")
	if t.WriteTimeout == 0 {
		t.WriteTimeout = time.Second * 2
	}
	if t.DialTimeout == 0 {
		t.DialTimeout = time.Second * 3
	}
	dl := websocket.Dialer{
		EnableCompression: true,
	}
	if !t.DnsLookupEnabled {
		dialer := &net.Dialer{}
		dl.NetDial = func(network, addr string) (net.Conn, error) {
			if addr == "speech.platform.bing.com:443" {
				i := rand.Intn(len(edgeChinaIpList))
				addr = fmt.Sprintf("%s:443", edgeChinaIpList[i])
			}
			pterm.Debug.Println("connect to IP: " + addr)
			return dialer.Dial(network, addr)
		}
	}
	header := http.Header{}
	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Origin", "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.66 Safari/537.36 Edg/103.0.1264.44")
	var ctx context.Context
	ctx, t.dialContextCancel = context.WithTimeout(context.Background(), t.DialTimeout)
	defer func() {
		t.dialContextCancel()
		t.dialContextCancel = nil
	}()

	var err error
	var resp *http.Response
	t.conn, resp, err = dl.DialContext(ctx, edgeWssUrl+t.uuid, header)
	if err != nil {
		if resp == nil {
			return err
		}
		return fmt.Errorf("%w: %s", err, resp.Status)
	}

	go func() {
		for {
			if t.conn == nil {
				return
			}
			messageType, p, err := t.conn.ReadMessage()
			closed := t.onReadMessage(messageType, p, err)
			if closed {
				t.conn = nil
				return
			}
		}
	}()

	return nil
}

func (t *EdgeTTS) CloseConn() {
	if t.conn != nil {
		_ = t.conn.Close()
		t.conn = nil
	}
}

func (t *EdgeTTS) GetAudio(ssml, format string) (audioData []byte, err error) {
	t.uuid = getUUID()
	if t.conn == nil {
		err := t.NewConn()
		if err != nil {
			return nil, err
		}
	}
	running := true
	defer func() { running = false }()
	finished := make(chan bool)
	failed := make(chan error)
	t.onReadMessage = func(messageType int, p []byte, errMessage error) bool {
		if messageType == -1 && p == nil && errMessage != nil { // 已经断开链接
			if running {
				failed <- errMessage
			}
			return true
		}

		if messageType == websocket.BinaryMessage {
			index := strings.Index(string(p), "Path:audio")
			data := []byte(string(p)[index+12:])
			audioData = append(audioData, data...)
		} else if messageType == websocket.TextMessage && string(p)[len(string(p))-14:len(string(p))-6] == "turn.end" {
			finished <- true
			return false
		}
		return false
	}
	err = t.sendConfigMessage(format)
	if err != nil {
		return nil, err
	}
	err = t.sendSsmlMessage(ssml)
	if err != nil {
		return nil, err
	}

	select {
	case <-finished:
		return audioData, err
	case errMessage := <-failed:
		return nil, errMessage
	}
}

func (t *EdgeTTS) GetAudioStream(ssml, format string, read func([]byte)) error {
	t.uuid = getUUID()
	if t.conn == nil {
		err := t.NewConn()
		if err != nil {
			return err
		}
	}

	running := true
	defer func() { running = false }()
	finished := make(chan bool)
	failed := make(chan error)
	t.onReadMessage = func(messageType int, p []byte, errMessage error) bool {
		if messageType == -1 && p == nil && errMessage != nil { // 已经断开链接
			if running {
				failed <- errMessage
			}
			return true
		}

		if messageType == websocket.BinaryMessage {
			index := strings.Index(string(p), "Path:audio")
			data := []byte(string(p)[index+12:])
			read(data)
		} else if messageType == websocket.TextMessage && string(p)[len(string(p))-14:len(string(p))-6] == "turn.end" {
			finished <- true
			return false
		}
		return false
	}
	err := t.sendConfigMessage(format)
	if err != nil {
		return err
	}
	err = t.sendSsmlMessage(ssml)
	if err != nil {
		return err
	}

	select {
	case <-finished:
		return nil
	case errMessage := <-failed:
		return errMessage
	}
}

func (t *EdgeTTS) sendConfigMessage(format string) error {
	cfgMsg := "X-Timestamp:" + getISOTime() + "\r\nContent-Type:application/json; charset=utf-8\r\n" + "Path:speech.config\r\n\r\n" +
		`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":"false","wordBoundaryEnabled":"false"},"outputFormat":"` + format + `"}}}}`
	_ = t.conn.SetWriteDeadline(time.Now().Add(t.WriteTimeout))
	err := t.conn.WriteMessage(websocket.TextMessage, []byte(cfgMsg))
	if err != nil {
		return fmt.Errorf("发送Config失败: %s", err)
	}

	return nil
}

func (t *EdgeTTS) sendSsmlMessage(ssml string) error {
	msg := "Path: ssml\r\nX-RequestId: " + t.uuid + "\r\nX-Timestamp: " + getISOTime() + "\r\nContent-Type: application/ssml+xml\r\n\r\n" + ssml
	_ = t.conn.SetWriteDeadline(time.Now().Add(t.WriteTimeout))
	err := t.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func (t *EdgeTTS) TextToMp3(text string, voice string, filePath string) error {
	ssml := CreateSSML(text, voice)
	pterm.Debug.Println(ssml)
	b, err := t.GetAudio(ssml, voiceFormat)
	if err != nil {
		pterm.Debug.Printf("Error: %v\n", err)
		return err
	}
	err = os.WriteFile(filePath, b, 0o644)
	if err != nil {
		pterm.Debug.Printf("Error: %v\n", err)
		return err
	}
	return nil
}

func (t *EdgeTTS) GetVoiceList() (voiceList []*Voice, err error) {
	client := req.C()
	r := client.R()
	_, err = r.SetSuccessResult(&voiceList).Get(voiceListUrl)
	return
}

func CreateSSML(text, voiceName string) string {
	r := strings.ReplaceAll(NormalSsmlTemplate, "{text}", text)
	r = strings.ReplaceAll(r, "{voiceName}", voiceName)
	return r
}

func getUUID() string {
	return uuid.NewV4().String()
}

func getISOTime() string {
	T := time.Now().String()
	return T[:23][:10] + "T" + T[:23][11:] + "Z"
}
