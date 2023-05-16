package openai_util

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

func AskChatGpt(content string) (*openai.ChatCompletionResponse, error) {
	initCli()
	resp, err := openaiCli.CreateChatCompletion(
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
