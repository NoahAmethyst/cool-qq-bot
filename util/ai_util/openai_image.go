package ai_util

import (
	"context"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

func GenerateImage(prompt, size string) (*openai.ImageResponse, error) {
	initCli()
	resp, err := openaiCli.CreateImage(
		context.Background(),
		openai.ImageRequest{
			Prompt:         prompt,
			Model:          openai.CreateImageModelDallE3,
			N:              1,
			Quality:        openai.CreateImageQualityHD,
			Size:           size,
			Style:          openai.CreateImageStyleVivid,
			ResponseFormat: openai.CreateImageResponseFormatURL,
			User:           "",
		},
	)

	log.Infof("receive openai response:%+v", resp)

	return &resp, err
}
