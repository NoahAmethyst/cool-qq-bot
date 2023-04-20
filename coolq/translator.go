package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/trans"
	translator_engine "github.com/NoahAmethyst/translator-engine"
	"regexp"
	"strings"
)

func (bot *CQBot) TransTextInPrivate(m *message.PrivateMessage) {
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if textEle == nil {
		return
	}

	text, done := parseSourceText(textEle)
	if done {
		return
	}

	from := translator_engine.AUTO
	var to string
	if isChinese(text) {
		to = translator_engine.EN
	} else {
		to = translator_engine.ZH
	}

	if r, err := trans.BalanceTranText(text, from, to); err != nil {
		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(
				fmt.Sprintf("翻译失败：%s", err.Error()))}})
	} else {
		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(
				r.Dst)}})
	}
}

func (bot *CQBot) TransTextInGroup(m *message.GroupMessage) {
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if textEle == nil {
		return
	}

	text, done := parseSourceText(textEle)
	if done {
		return
	}

	from := translator_engine.AUTO
	var to string
	if isChinese(text) {
		to = translator_engine.EN
	} else {
		to = translator_engine.ZH
	}

	if r, err := trans.BalanceTranText(text, from, to); err != nil {
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{message.NewReply(m),
			message.NewText(
				fmt.Sprintf("翻译失败：%s", err.Error()))}})
	} else {
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{message.NewReply(m),
			message.NewText(
				r.Dst)}})
	}

}

func parseSourceText(textEle *message.TextElement) (string, bool) {
	//re := regexp.MustCompile(`^#(\S+)\s(.*)$`)
	//
	//match := re.FindStringSubmatch(textEle.Content)
	//
	//if len(match) != 3 {
	//	return "", true
	//}

	//text := strings.TrimSpace(match[2])

	match := strings.ReplaceAll(textEle.Content, "#翻译 ", "")
	text := strings.TrimSpace(match)
	return text, false
}

func isChinese(str string) bool {
	re := regexp.MustCompile("[\u4e00-\u9fa5]")
	return re.MatchString(str)
}
