package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/openai_util"
)

func (bot *CQBot) askChatGpt(source message.Source, m *message.GroupMessage) {
	if len(m.Elements) <= 1 {
		return
	}

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
		return
	}
	go func(b *CQBot, m *message.GroupMessage) {
		answer, err := openai_util.AskChatGpt(textEle.Content)
		if err != nil {
			answer = fmt.Sprintf("调用openAi 失败：%s", err.Error())
		}
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(answer)}})
	}(bot, m)
}
