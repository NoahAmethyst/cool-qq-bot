package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/trans"
	translator_engine "github.com/NoahAmethyst/translator-engine"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

type Translator interface {
	Reply(content string)
	GetText() *message.TextElement
	Check() bool
	Target() int64
}

type PrivateTranslator struct {
	bot *CQBot
	m   *message.PrivateMessage
}

func (p *PrivateTranslator) Reply(msg string) {

	p.bot.SendPrivateMessage(p.Target(), 0, &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(
			msg)}})

}

func (p *PrivateTranslator) Target() int64 {
	return p.m.Sender.Uin
}

func (p *PrivateTranslator) Check() bool {
	return p.bot != nil && p.m != nil
}

func (p *PrivateTranslator) GetText() *message.TextElement {
	var textEle *message.TextElement
	for _, _ele := range p.m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}
	return textEle
}

type GroupTranslator struct {
	bot *CQBot
	m   *message.GroupMessage
}

func (p *GroupTranslator) Reply(msg string) {

	p.bot.SendGroupMessage(p.Target(), &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewReply(p.m),
		message.NewText(
			msg)}})

}

func (p *GroupTranslator) Target() int64 {
	return p.m.GroupCode
}

func (p *GroupTranslator) Check() bool {
	return p.bot != nil && p.m != nil
}

func (p *GroupTranslator) GetText() *message.TextElement {
	var textEle *message.TextElement
	for _, _ele := range p.m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}
	return textEle
}

func TransText(t Translator) {

	if t == nil || !t.Check() {
		log.Warn("invalid translator")
		return
	}
	textEle := t.GetText()

	if textEle == nil {
		return
	}
	text, done := parseSourceText(textEle)
	if done {
		t.Reply("缺少待翻译文本")
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
		t.Reply(fmt.Sprintf("翻译失败：%s", err.Error()))
	} else {
		t.Reply(r.Dst)
	}

}

func parseSourceText(textEle *message.TextElement) (string, bool) {
	var text string
	re := regexp.MustCompile(`#(\S+)\s+(?s)(.*)`)

	match := re.FindStringSubmatch(textEle.Content)

	if len(match) != 3 {
		return text, true
	}

	text = strings.TrimSpace(match[2])

	if len(text) == 0 {
		return text, true
	}

	//match := strings.ReplaceAll(textEle.Content, "#翻译 ", "")
	//text := strings.TrimSpace(match)
	return text, false
}

func isChinese(str string) bool {
	re := regexp.MustCompile("[\u4e00-\u9fa5]")
	return re.MatchString(str)
}
