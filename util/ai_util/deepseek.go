package ai_util

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

const deepSeekChat = "deepseek-chat"
const deepSeekBaseUrl = "https://api.deepseek.com"

func AskDeepSeek(ctx []openai.ChatCompletionMessage) (openai.ChatCompletionResponse, error) {
	initCli()
	resp, err := deepSeekCli.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    deepSeekChat,
			Messages: ctx,
		},
	)
	log.Infof("收到deepSeek响应:%+v", resp)
	return resp, err
}
