package openai_util

import (
	"context"
	"github.com/Mrs4s/go-cqhttp/constant"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var cli *openai.Client

func AskChatGpt(content string) (*openai.ChatCompletionResponse, error) {
	initCli()
	resp, err := cli.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
	log.Infof("收到openai响应:%+v", resp)
	return &resp, err

}

func initCli() {
	apiKey := os.Getenv(constant.OPENAI_API_KEY)
	if cli == nil {
		if len(os.Getenv(constant.NOT_MIRROR)) > 0 {
			cli = openai.NewClient(apiKey)
		} else {
			config := openai.DefaultConfig(apiKey)
			config.HTTPClient.Timeout = time.Minute * 60
			config.BaseURL = "https://open.aiproxy.xyz/v1"
			cli = openai.NewClientWithConfig(config)
		}
	}
}
