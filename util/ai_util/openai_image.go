package ai_util

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"github.com/tristan-club/kit/log"
)

func GenerateImage(prompt string) (*openai.ImageResponse, error) {
	initCli()
	resp, err := openaiCli.CreateImage(
		context.Background(),
		openai.ImageRequest{
			Prompt:         prompt,
			Size:           openai.CreateImageSize512x512,
			ResponseFormat: openai.CreateImageResponseFormatURL,
			N:              1,
		},
	)
	log.Info().Msgf("收到openai响应：%+v", resp)
	return &resp, err
}
