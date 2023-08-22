package ai_util

import (
	"context"
	go_ernie "github.com/anhao/go-ernie"
)

func AskErnie(ctx []go_ernie.ChatCompletionMessage) (go_ernie.ErnieBotResponse, error) {
	completion, err := ernieCli.CreateErnieBotChatCompletion(context.Background(), go_ernie.ErnieBotRequest{
		Messages: ctx,
	})
	return completion, err
}
