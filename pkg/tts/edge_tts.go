package tts

import (
	"fmt"
	"github.com/pterm/pterm"
	"micli/pkg/tts/edgetts"
	"os"
	"path/filepath"
	"time"
)

const voiceFormat = "audio-24khz-48kbitrate-mono-mp3"

func TextToMp3(text string, voice string) (string, error) {
	tts := edgetts.EdgeTTS{
		DnsLookupEnabled: true,
	}
	ssml := edgetts.CreateSSML(text, voice)
	//pterm.Debug.Println(ssml)
	b, err := tts.GetAudio(ssml, voiceFormat)
	if err != nil {
		pterm.Error.Printf("Error: %v\n", err)
		return "", err
	}
	filename := fmt.Sprintf("%s.mp3", time.Now().Format("20060102150405"))
	tempDir, _ := os.MkdirTemp("", "micli-tts-")
	fp := filepath.Join(tempDir, filename)
	err = os.WriteFile(fp, b, 0644)
	if err != nil {
		pterm.Error.Printf("Error: %v\n", err)
		return "", err
	}
	return fp, nil
}

func LoadVoiceList() ([]*edgetts.Voice, error) {
	tts := edgetts.EdgeTTS{}
	var err error
	var list []*edgetts.Voice
	list, err = tts.GetVoiceList()
	return list, err
}
