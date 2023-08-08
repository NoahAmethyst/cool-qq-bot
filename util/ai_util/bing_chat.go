package ai_util

import (
	bingchat_api "github.com/NoahAmethyst/bingchat-api"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	remoteConversationUrl = "https://bing.vcanbb.top/turing/conversation/create"
	remoteConversationWs  = "wss://bing.vcanbb.top/sydney/ChatHub"
)

type BingChatResp struct {
	Reference   map[string]string
	Answer      string
	Suggestions []string
}

func NewBingChat() (bingchat_api.IBingChat, error) {
	chat, err := bingchat_api.NewBingChat(os.Getenv("COOKIE"), bingchat_api.ConversationBalanceStyle, 2*time.Minute)
	if err != nil {
		return chat, err
	}
	return chat, err
}

func AskBingChat(chat bingchat_api.IBingChat, content string) (*BingChatResp, error) {
	var resp *BingChatResp
	message, err := chat.SendMessage(content)
	if err != nil {
		return resp, err
	}
	var respBuilder strings.Builder
	done := false
	for {
		select {
		case msg, ok := <-message.Notify:
			if !ok {
				done = true
				break
			}
			respBuilder.WriteString(msg)
		case <-time.After(2 * time.Minute):
			done = true
			break
		}
		if done {
			break
		}
	}

	answer := strings.ReplaceAll(respBuilder.String(), "**", "")
	//Remove all reference symbols
	var re = regexp.MustCompile(`\[\^(\d+)\^]`)
	answer = re.ReplaceAllString(answer, "")

	resp = &BingChatResp{
		Reference:   message.References,
		Answer:      answer,
		Suggestions: message.Suggest,
	}

	return resp, err
}
