package ai_util

import (
	"context"
	go_ernie "github.com/anhao/go-ernie"
	"github.com/pkg/errors"
)

func AskErnie(ctx []go_ernie.ChatCompletionMessage) (go_ernie.ErnieBotTurboResponse, error) {
	initCli()
	if ernieCli == nil {
		return go_ernie.ErnieBotTurboResponse{}, errors.New("Ernie Client not init")
	}
	completion, err := ernieCli.CreateErnieBotTurboChatCompletion(context.Background(), go_ernie.ErnieBotTurboRequest{
		Messages: ctx,
	})
	return completion, err
}
