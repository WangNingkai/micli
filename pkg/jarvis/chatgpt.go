package jarvis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"micli/internal/conf"

	"github.com/pterm/pterm"
	"github.com/sashabaranov/go-openai"
)

type ChatGPT struct {
	Key     string
	Proxy   string
	BaseURL string
	Model   string

	InStream      bool
	StreamMessage chan string
}

func NewChatGPT() *ChatGPT {
	return &ChatGPT{
		Key:     conf.Cfg.Section("openai").Key("KEY").MustString(""),
		BaseURL: conf.Cfg.Section("openai").Key("BASE_URL").MustString(""),
	}
}

func (c *ChatGPT) Ask(msg string) (reply string, err error) {
	model := openai.GPT3Dot5Turbo16K
	var config openai.ClientConfig
	var resp openai.ChatCompletionResponse
	if strings.Contains(c.BaseURL, "azure") {
		config = openai.DefaultAzureConfig(c.Key, c.BaseURL)
		config.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				"gpt-3.5-turbo-16k": "gpt-35-turbo-16k",
			}
			return azureModelMapping[model]
		}
	} else {
		if c.Model != "" {
			model = c.Model
		}
		config := openai.DefaultConfig(c.Key)
		if c.BaseURL != "" {
			config.BaseURL = c.BaseURL
		}
	}
	config = openai.DefaultConfig(c.Key)
	if c.BaseURL != "" {
		config.BaseURL = c.BaseURL
	}

	client := openai.NewClientWithConfig(config)
	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("\nYou are ChatGPT, a large language model trained by OpenAI.\nKnowledge cutoff: 2021-09\nCurrent model: %s\nCurrent time: %s\n", model, time.Now().Format("2006-01-02 15:04:05")),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: msg,
			},
		},
		Temperature:      0.5,
		TopP:             1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
	}
	resp, err = client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		err = fmt.Errorf("ChatCompletion error: %v\n", err)
		return
	}

	reply = resp.Choices[0].Message.Content
	return
}

func (c *ChatGPT) AskStream(msg string) error {
	var config openai.ClientConfig
	model := openai.GPT3Dot5Turbo16K
	if strings.Contains(c.BaseURL, "azure") {
		config = openai.DefaultAzureConfig(c.Key, c.BaseURL)
		config.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				"gpt-3.5-turbo-16k": "gpt-35-turbo-16k",
			}
			return azureModelMapping[model]
		}
	} else {
		if c.Model != "" {
			model = c.Model
		}
		config := openai.DefaultConfig(c.Key)
		if c.BaseURL != "" {
			config.BaseURL = c.BaseURL
		}
	}

	client := openai.NewClientWithConfig(config)
	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("\nYou are ChatGPT, a large language model trained by OpenAI.\nKnowledge cutoff: 2021-09\nCurrent model: %s\nCurrent time: %s\n", model, time.Now().Format("2006-01-02 15:04:05")),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: msg,
			},
		},
		Temperature:      0.5,
		TopP:             1,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		Stream:           true,
	}
	var (
		err    error
		stream *openai.ChatCompletionStream
	)
	stream, err = client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		err = fmt.Errorf("ChatCompletionStream error: %v\n", err)
		return err
	}
	defer stream.Close()
	for {
		var response openai.ChatCompletionStreamResponse
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			close(c.StreamMessage)
			break
		}
		if err != nil {
			pterm.Debug.Printf("\nStream error: %v\n", err)
			return err
		}
		if response.Choices[0].FinishReason != "stop" {
			reply := response.Choices[0].Delta.Content
			c.StreamMessage <- reply
		}
	}
	return nil
}
