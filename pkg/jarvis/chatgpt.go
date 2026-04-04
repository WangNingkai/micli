package jarvis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
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

	client     *openai.Client
	clientOnce sync.Once
}

func NewChatGPT() *ChatGPT {
	return &ChatGPT{
		Key:     conf.Cfg.Section("openai").Key("KEY").MustString(""),
		BaseURL: conf.Cfg.Section("openai").Key("BASE_URL").MustString(""),
	}
}

// getClient 获取或创建 OpenAI client（懒加载，复用）
func (c *ChatGPT) getClient() *openai.Client {
	c.clientOnce.Do(func() {
		c.client = openai.NewClientWithConfig(c.buildConfig())
	})
	return c.client
}

// buildConfig 构建客户端配置
func (c *ChatGPT) buildConfig() openai.ClientConfig {
	if strings.Contains(c.BaseURL, "azure") {
		config := openai.DefaultAzureConfig(c.Key, c.BaseURL)
		config.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				"gpt-3.5-turbo-16k": "gpt-35-turbo-16k",
			}
			return azureModelMapping[model]
		}
		return config
	}
	config := openai.DefaultConfig(c.Key)
	if c.BaseURL != "" {
		config.BaseURL = c.BaseURL
	}
	return config
}

// getModel 获取使用的模型
func (c *ChatGPT) getModel() string {
	if c.Model != "" {
		return c.Model
	}
	return openai.GPT3Dot5Turbo16K
}

func (c *ChatGPT) Ask(msg string) (reply string, err error) {
	model := c.getModel()
	client := c.getClient()
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
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		err = fmt.Errorf("ChatCompletion error: %v\n", err)
		return
	}

	reply = resp.Choices[0].Message.Content
	return
}

func (c *ChatGPT) AskStream(msg string) error {
	model := c.getModel()
	client := c.getClient()
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
