package ai_util

import (
	"context"
	go_ernie "github.com/anhao/go-ernie"
	"github.com/baidubce/bce-qianfan-sdk/go/qianfan"
	"github.com/pkg/errors"
)

func AskErnie(ctx []go_ernie.ChatCompletionMessage) (go_ernie.ErnieBotTurboResponse, error) {
	initCli()
	if ernieCli == nil {
		return go_ernie.ErnieBotTurboResponse{}, errors.New("Ernie Client not init")
	}
	//completion, err := ernieCli.CreateErnieBotTurboChatCompletion(context.Background(), go_ernie.ErnieBotTurboRequest{
	//	Messages: ctx,
	//})

	chat := qianfan.NewChatCompletion(
		qianfan.WithModel("ERNIE-4.0-Turbo-8K"),
	)

	message := make([]qianfan.ChatCompletionMessage, 0, 2)
	for _, _ctx := range ctx {
		message = append(message, qianfan.ChatCompletionMessage{
			Role:    _ctx.Role,
			Content: _ctx.Content,
		})
	}

	completion := go_ernie.ErnieBotTurboResponse{
		ErnieBotResponse: go_ernie.ErnieBotResponse{
			Id:               "",
			Object:           "",
			Created:          0,
			SentenceId:       0,
			IsEnd:            false,
			IsTruncated:      false,
			Result:           "",
			NeedClearHistory: false,
			Usage:            go_ernie.ErnieUsage{},
			APIError:         go_ernie.APIError{},
		},
	}

	resp, err := chat.Do(
		context.TODO(),
		&qianfan.ChatCompletionRequest{
			Messages: message,
		},
	)
	if err != nil {
		return completion, err
	} else {
		completion.ErnieBotResponse.Id = resp.Id
		completion.ErnieBotResponse.Object = resp.Object
		completion.ErnieBotResponse.Created = resp.Created
		completion.ErnieBotResponse.SentenceId = resp.SentenceId
		completion.ErnieBotResponse.IsEnd = resp.IsEnd
		completion.ErnieBotResponse.IsTruncated = resp.IsTruncated
		completion.ErnieBotResponse.Result = resp.Result
		completion.ErnieBotResponse.NeedClearHistory = resp.NeedClearHistory
		completion.ErnieBotResponse.Usage = go_ernie.ErnieUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
		completion.ErnieBotResponse.APIError = go_ernie.APIError{
			ErrorCode: resp.ErrorCode,
			ErrorMsg:  resp.ErrorMsg,
			ID:        resp.Id,
		}
	}

	return completion, err
}
