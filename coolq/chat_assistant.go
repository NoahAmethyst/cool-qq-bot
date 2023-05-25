package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/ai_util"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func (bot *CQBot) askAIAssistantInPrivate(_ *client.QQClient, m *message.PrivateMessage) {
	textEle, done := parsePMAsk(m)
	if done {
		return
	}

	v, ok := bot.state.privateDialogueSession.getParentMsgId(m.Sender.Uin)

	var answer *ai_util.AIAssistantResp
	var err error
	recvChan := make(chan struct{}, 1)
	defer close(recvChan)
	go func(uid int64) {
		select {
		case <-recvChan:
			return
		case <-time.After(time.Second * 10):
			bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("OPENAI正在响应，请稍等...")}})
		}
	}(m.Sender.Uin)
	if !ok {
		answer, err = ai_util.AskAIAssistant(textEle.Content)
	} else {
		answer, err = ai_util.AskAIAssistant(textEle.Content, v)
	}
	recvChan <- struct{}{}

	if err != nil {
		log.Errorf("ask ai assistent error:%s", err.Error())
		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(err.Error())}})
	} else {
		bot.state.privateDialogueSession.putParentMsgId(m.Sender.Uin, answer.ID)
		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(answer.Text)}})
	}

}

func (bot *CQBot) askAIAssistantInGroup(_ *client.QQClient, m *message.GroupMessage) {
	textEle, done := parseGMAsk(m, bot)
	if done {
		return
	}

	v, ok := bot.state.groupDialogueSession.getParentMsgId(m.Sender.Uin)

	var answer *ai_util.AIAssistantResp
	var err error
	recvChan := make(chan struct{}, 1)
	defer close(recvChan)
	go func(group int64) {
		select {
		case <-recvChan:
			return
		case <-time.After(time.Second * 10):
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("OPENAI正在响应，请稍等...")}})
		}
	}(m.GroupCode)

	if !ok {
		answer, err = ai_util.AskAIAssistant(textEle.Content)
	} else {
		answer, err = ai_util.AskAIAssistant(textEle.Content, v)
	}
	recvChan <- struct{}{}

	if err != nil {
		log.Errorf("ask ai assistent error:%s", err.Error())
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{message.NewReply(m),
			message.NewText(err.Error())}})
	} else {
		bot.state.groupDialogueSession.putParentMsgId(m.Sender.Uin, answer.ID)
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{message.NewReply(m),
			message.NewText(answer.Text)}})
	}

}

func (bot *CQBot) askChatGptInPrivate(_ *client.QQClient, m *message.PrivateMessage) {
	textEle, done := parsePMAsk(m)
	if done {
		return
	}

	recvChan := make(chan struct{}, 1)
	defer close(recvChan)
	go func(uid int64) {
		select {
		case <-recvChan:
			return
		case <-time.After(time.Second * 10):
			bot.SendPrivateMessage(uid, 0, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("OPENAI正在响应，请稍等...")}})
		}
	}(m.Sender.Uin)

	answer := askChatGpt(textEle)

	recvChan <- struct{}{}

	bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(answer)}})
}

func parsePMAsk(m *message.PrivateMessage) (*message.TextElement, bool) {
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if textEle == nil ||
		(!strings.Contains(textEle.Content, "？") &&
			!strings.Contains(textEle.Content, "?")) {
		return nil, true
	}
	return textEle, false
}

func (bot *CQBot) askChatGptInGroup(_ *client.QQClient, m *message.GroupMessage) {
	textEle, done := parseGMAsk(m, bot)
	if done {
		return
	}

	recvChan := make(chan struct{}, 1)
	defer close(recvChan)
	go func(group int64) {
		select {
		case <-recvChan:
			return
		case <-time.After(time.Second * 10):
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("OPENAI正在响应，请稍等...")}})
		}
	}(m.GroupCode)

	recvChan <- struct{}{}

	answer := askChatGpt(textEle)

	bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{message.NewReply(m),
		message.NewText(answer)}})

}

func parseGMAsk(m *message.GroupMessage, bot *CQBot) (*message.TextElement, bool) {
	var atEle *message.AtElement
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.At:
			atEle = _ele.(*message.AtElement)
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if atEle == nil || textEle == nil {
		return nil, true
	}

	if atEle.Target != bot.Client.Uin {
		log.Warnf("mention target is not bot")
		return nil, true
	}
	return textEle, false
}

func askChatGpt(textEle *message.TextElement) string {
	var answer string
	resp, err := ai_util.AskChatGpt(textEle.Content)
	//重试机制
	if err != nil {
		maxRetry := 6
		for i := 0; i < maxRetry; i++ {
			time.Sleep(500 * time.Millisecond)
			log.Warnf("call openai failed cause:%s,retry:%d", err.Error(), i+1)
			resp, err = ai_util.AskChatGpt(textEle.Content)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		answer = fmt.Sprintf("调用openAi 失败：%s", err.Error())
	} else {
		if resp == nil || len(resp.Choices) == 0 {
			log.Warnf("openai 返回空结构：%v", resp)
			answer = fmt.Sprintf("openai返回空结构")
		} else {
			answer = resp.Choices[0].Message.Content
		}
	}
	return answer
}
