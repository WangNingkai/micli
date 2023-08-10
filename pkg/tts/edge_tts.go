package tts

import (
	"fmt"
	"github.com/pterm/pterm"
	"log"
	"math/rand"
	"micli/pkg/util"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/longbai/edgetts"
)

const voiceFormat = "audio-24khz-48kbitrate-mono-mp3"

type edgeTTS struct {
	hostname string
	port     int
	tempDir  string
}

func (s *edgeTTS) startServer() {
	s.tempDir, _ = os.MkdirTemp("", "micli-tts-")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(s.tempDir, filepath.Base(r.URL.Path)))
	})
	s.hostname = util.GetHostname()
	rd := rand.Uint32() % 50
	s.port = int(8000 + rd)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.hostname, s.port), nil)
	if err != nil {
		pterm.Error.Printf("Error starting HTTP server: %v\n", err)
	}
	pterm.Success.Printf("Serving on %s:%d\n", s.hostname, s.port)
}

func (s *edgeTTS) textToMp3(text string, ttsLang string) (string, error) {
	ttsEdge := edgetts.EdgeTTS{}
	ssml := edgetts.CreateSSML(text, ttsLang)
	log.Println(ssml)
	b, err := ttsEdge.GetAudio(ssml, voiceFormat)
	if err != nil {
		pterm.Error.Printf("Error: %v\n", err)
		return "", err
	}
	filename := fmt.Sprintf("%s.mp3", time.Now().Format("20060102150405"))
	fp := filepath.Join(s.tempDir, filename)
	err = os.WriteFile(fp, b, 0644)
	if err != nil {
		pterm.Error.Printf("Error: %v\n", err)
		return "", err
	}
	return filename, nil
}
