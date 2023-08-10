package tts

import (
	"fmt"
	"github.com/pterm/pterm"
	"os"
	"path/filepath"
	"time"

	"github.com/longbai/edgetts"
)

const voiceFormat = "audio-24khz-48kbitrate-mono-mp3"

func TextToMp3(text string, ttsLang string) (string, error) {
	ttsEdge := edgetts.EdgeTTS{}
	ssml := edgetts.CreateSSML(text, ttsLang)
	//pterm.Debug.Println(ssml)
	b, err := ttsEdge.GetAudio(ssml, voiceFormat)
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
