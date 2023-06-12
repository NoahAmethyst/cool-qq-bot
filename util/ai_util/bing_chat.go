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
	remoteConversationWs  = "wss://sydney.vcanbb.chat/sydney/ChatHub"
)

type BingChatResp struct {
	Reference   map[string]string
	Answer      string
	Suggestions []string
}

func AskBingChat(content string) (*BingChatResp, error) {
	var resp *BingChatResp
	chat, err := bingchat_api.NewBingChat(os.Getenv("COOKIE"), bingchat_api.ConversationBalanceStyle, 2*time.Minute)
	if err != nil {
		return resp, err
	}
	chat.SetRemote(remoteConversationUrl, remoteConversationWs)

	message, err := chat.SendMessage(content)
	defer chat.Close()
	if err != nil {
		return resp, err
	}
	var respBuilder strings.Builder
	for {
		msg, ok := <-message.Notify
		if !ok {
			break
		}
		respBuilder.WriteString(msg)
	}

	resp = &BingChatResp{
		Reference:   make(map[string]string),
		Answer:      strings.ReplaceAll(respBuilder.String(), "**", ""),
		Suggestions: message.Suggest,
	}

	// Parse reference
	linkRegex := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	links := linkRegex.FindAllStringSubmatch(respBuilder.String(), -1)

	for _, link := range links {
		resp.Reference[link[1]] = link[2]
	}

	return resp, err
}
