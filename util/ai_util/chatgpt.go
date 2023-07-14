package ai_util

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

func AskChatGpt(ctx []openai.ChatCompletionMessage) (openai.ChatCompletionResponse, error) {
	initCli()
	resp, err := openaiCli.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo0613,
			Messages: ctx,
		},
	)
	log.Infof("收到openai响应:%+v", resp)
	return resp, err

}

func AskChatGpt4(ctx []openai.ChatCompletionMessage) (openai.ChatCompletionResponse, error) {
	initCli()
	resp, err := chimeraCli.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4,
			Messages: ctx,
		},
	)
	log.Infof("收到openai响应:%+v", resp)
	return resp, err
}
