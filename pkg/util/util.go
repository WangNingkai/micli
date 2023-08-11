package util

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	mathRand "math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CreatNestedFile 给定path创建文件，如果目录不存在就递归创建
func CreatNestedFile(path string) (*os.File, error) {
	basePath := filepath.Dir(path)
	if !Exists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			return nil, err
		}
	}

	return os.Create(path)
}

func GetRandom(length int) string {
	var r = mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randomStr := make([]byte, length)
	for i := range randomStr {
		randomStr[i] = charset[r.Intn(len(charset))]
	}
	return string(randomStr)
}

func SignNonce(sSecurity string, nonce string) (string, error) {
	decodedSSecurity, err := base64.StdEncoding.DecodeString(sSecurity)
	if err != nil {
		return "", err
	}

	decodedNonce, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(decodedSSecurity)
	hash.Write(decodedNonce)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

func GenNonce() string {
	nonce := make([]byte, 12)
	_, err := rand.Read(nonce[:8])
	if err != nil {
		return ""
	}
	binary.BigEndian.PutUint32(nonce[8:], uint32(time.Now().Unix()/60))
	return base64.StdEncoding.EncodeToString(nonce)
}

func SignData(uri string, data any, sSecurity string) url.Values {
	var dataStr []byte
	if s, ok := data.(string); ok {
		dataStr = []byte(s)
	} else {
		var err error
		dataStr, err = json.Marshal(data)
		if err != nil {
			return nil
		}
	}

	encodedNonce := GenNonce()
	sNonce, err := SignNonce(sSecurity, encodedNonce)
	if err != nil {
		return nil
	}
	msg := fmt.Sprintf("%s&%s&%s&data=%s", uri, sNonce, encodedNonce, dataStr)
	sb, _ := base64.StdEncoding.DecodeString(sNonce)
	sign := hmac.New(sha256.New, sb)
	sign.Write([]byte(msg))
	signature := base64.StdEncoding.EncodeToString(sign.Sum(nil))
	return url.Values{
		"_nonce":    {encodedNonce},
		"data":      {string(dataStr)},
		"signature": {signature},
	}
}

func TwinsSplit(str, sep string, def string) (string, string) {
	pos := strings.Index(str, sep)
	if pos == -1 {
		return str, def
	}
	return str[0:pos], str[pos+1:]
}

// StringToValue converts a string to a value.
func StringToValue(str string) interface{} {
	switch str {
	case "null", "none":
		return nil
	case "false":
		return false
	case "true":
		return true
	default:
		if intValue, err := strconv.Atoi(str); err == nil {
			return intValue
		}
		return str
	}
}

func StringOrValue(str string) interface{} {
	if IsDigit(str) {
		return StringToValue(str)
	}
	return str
}

func Unzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func IsDigit(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func IsJSON(str string) bool {
	var js interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

var (
	regex1             = regexp.MustCompile(`[「」『』《》“”'\"()（）]`)
	regex2             = regexp.MustCompile(`(^|[^-])-($|[^-])`)
	endingPunctuations = []string{"。", "？", "！", "；", ".", "?", "!", ";"}
)

// CalculateTTSElapse returns the elapsed time for TTS
func CalculateTTSElapse(text string) time.Duration {
	speed := 4.5

	// Replace the first part of the regex
	result := regex1.ReplaceAllString(text, "")

	// Replace the second part of the regex
	result = regex2.ReplaceAllString(result, "")

	v := float64(len(result)) / speed
	return time.Duration(v+1) * time.Second
}

func SplitSentences(textStream <-chan string) <-chan string {
	result := make(chan string)
	go func() {
		cur := ""
		for text := range textStream {
			cur += text
			for _, p := range endingPunctuations {
				if strings.HasSuffix(cur, p) {
					result <- cur
					cur = ""
					break
				}
			}
		}
		if cur != "" {
			result <- cur
		}
		close(result)
	}()
	return result
}

func GetHostname() string {
	if hostname, exists := os.LookupEnv("MICLI_HOSTNAME"); exists {
		return hostname
	}

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func FindKeyByPartialString(dictionary map[string]string, partialKey string) string {
	for key, value := range dictionary {
		if strings.Contains(partialKey, key) {
			return value
		}
	}
	return ""
}
