package ai_util

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

func GenerateImage(prompt, size string) (*openai.ImageResponse, error) {
	initCli()
	resp, err := openaiCli.CreateImage(
		context.Background(),
		openai.ImageRequest{
			Prompt:         prompt,
			Size:           size,
			ResponseFormat: openai.CreateImageResponseFormatURL,
			N:              1,
		},
	)
	log.Info().Msgf("收到openai响应：%+v", resp)
	return &resp, err
}
