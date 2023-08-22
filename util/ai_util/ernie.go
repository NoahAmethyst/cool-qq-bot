package ai_util

import (
	"context"
	go_ernie "github.com/anhao/go-ernie"
	"github.com/pkg/errors"
)

func AskErnie(ctx []go_ernie.ChatCompletionMessage) (go_ernie.ErnieBotResponse, error) {
	initCli()
	if ernieCli == nil {
		return go_ernie.ErnieBotResponse{}, errors.New("Ernie Client not init")
	}
	completion, err := ernieCli.CreateErnieBotChatCompletion(context.Background(), go_ernie.ErnieBotRequest{
		Messages: ctx,
	})
	return completion, err
}
