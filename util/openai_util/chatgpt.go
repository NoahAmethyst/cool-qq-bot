package openai_util

import (
	"context"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/pkg/errors"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var cli *openai.Client

func AskChatGpt(content string) (string, error) {
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
	if len(resp.Choices) > 0 && len(resp.Choices[0].Message.Content) > 0 {
		return resp.Choices[0].Message.Content, err
	}
	if err == nil {
		err = errors.New("openai响应为空")
	}

	return "", err

}

func initCli() {
	apiKey := os.Getenv(constant.OPENAI_API_KEY)
	if cli == nil {
		if len(os.Getenv(constant.NOT_MIRROR)) > 0 {
			cli = openai.NewClient(apiKey)
		} else {
			config := openai.DefaultConfig(apiKey)
			config.HTTPClient.Timeout = time.Minute * 60
			config.BaseURL = "https://cold-weasel-95.deno.dev/v1"
			cli = openai.NewClientWithConfig(config)
		}
	}
}
