package coolq

import (
	"fmt"
	"strings"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/NoahAmethyst/go-cqhttp/cluster/jarvis_svc"
	"github.com/NoahAmethyst/go-cqhttp/util/content_util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

type Assistant interface {
	Reply(content string)
	GetText() *message.TextElement
	Mention() *message.AtElement
	Check() bool
	Chat() int64
	Sender() int64
	JarvisUserID() string
	Me() int64
	Bot() *CQBot
}

type PrivateAssistant struct {
	bot *CQBot
	m   *message.PrivateMessage
}

func (p *PrivateAssistant) Reply(msg string) {
	_splittedMsg := content_util.SplitLongMessage(msg, 1000) // 每1000字符分割一次
	for i, _msg := range _splittedMsg {
		pageMsg := _msg
		dormant := false
		if len(_splittedMsg) > 1 {
			pageMsg = fmt.Sprintf("【%d/%d】\n%s", i+1, len(_splittedMsg), _msg)
			dormant = true
		}
		p.bot.SendPrivateMessage(p.Chat(), 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(pageMsg),
		}})
		if dormant {
			time.Sleep(time.Second) // 每条消息间隔1秒
		}
	}
}

func (p *PrivateAssistant) Bot() *CQBot {
	return p.bot
}

func (p *PrivateAssistant) Chat() int64 {
	return p.m.Sender.Uin
}

func (p *PrivateAssistant) Sender() int64 {
	return p.m.Sender.Uin
}

func (p *PrivateAssistant) JarvisUserID() string {
	return fmt.Sprintf("qq:private:%d", p.Sender())
}

func (p *PrivateAssistant) Check() bool {
	return p.bot != nil && p.m != nil
}

func (p *PrivateAssistant) GetText() *message.TextElement {
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

func (p *PrivateAssistant) Mention() *message.AtElement {
	mentionEle := &message.AtElement{}
	for _, _ele := range p.m.Elements {
		if _ele.Type() == message.At {
			mentionEle = _ele.(*message.AtElement)
		}
	}
	return mentionEle
}

func (p *PrivateAssistant) Me() int64 {
	return p.bot.Client.Uin
}

type GroupAssistant struct {
	bot *CQBot
	m   *message.GroupMessage
}

func (p *GroupAssistant) Reply(msg string) {
	p.bot.SendGroupMessage(p.Chat(), &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewReply(p.m),
		message.NewText(msg),
	}})
}

func (p *GroupAssistant) Bot() *CQBot {
	return p.bot
}

func (p *GroupAssistant) Chat() int64 {
	return p.m.GroupCode
}

func (p *GroupAssistant) Sender() int64 {
	return p.m.Sender.Uin
}

func (p *GroupAssistant) JarvisUserID() string {
	return fmt.Sprintf("qq:group:%d:%d", p.Chat(), p.Sender())
}

func (p *GroupAssistant) Check() bool {
	return p.bot != nil && p.m != nil
}

func (p *GroupAssistant) GetText() *message.TextElement {
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

func (p *GroupAssistant) Mention() *message.AtElement {
	mentionEle := &message.AtElement{}
	for _, _ele := range p.m.Elements {
		if _ele.Type() == message.At {
			mentionEle = _ele.(*message.AtElement)
		}
	}
	return mentionEle
}

func (p *GroupAssistant) Me() int64 {
	return p.bot.Client.Uin
}

func ShowAssistantMode(assistant Assistant) {
	if assistant == nil || !assistant.Check() {
		log.Warn("invalid assistant")
		return
	}

	textEle := assistant.GetText()
	if textEle == nil {
		return
	}
	argument := strings.TrimSpace(strings.TrimPrefix(textEle.Content, "#模式"))
	if argument == "" {
		assistant.Reply("当前助手模式：Jarvis")
		return
	}
	assistant.Reply("模型已由 Jarvis 服务统一管理，不支持在机器人端切换")
}

func AskAssistant(assistant Assistant) {
	if assistant == nil || !assistant.Check() {
		log.Warn("invalid assistant")
		return
	}

	textEle := assistant.GetText()
	if textEle == nil {
		return
	}

	if !strings.Contains(textEle.Content, "?") &&
		!strings.Contains(textEle.Content, "？") &&
		assistant.Me() != assistant.Mention().Target {
		return
	}

	timer := time.NewTimer(10 * time.Second)
	cancelProgress := make(chan struct{})
	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		select {
		case <-cancelProgress:
		case <-timer.C:
			assistant.Reply("Jarvis 正在响应，请稍后...")
		}
	}()

	userID := assistant.JarvisUserID()
	response, err := jarvis_svc.Chat(textEle.Content, userID)
	if timer.Stop() {
		close(cancelProgress)
	}
	<-progressDone
	if err != nil {
		log.Errorf("Jarvis chat failed: code=%s", status.Code(err))
		assistant.Reply("调用 Jarvis 失败，请稍后重试")
		return
	}
	if response == nil || strings.TrimSpace(response.Answer) == "" {
		log.Warn("Jarvis chat returned empty answer")
		assistant.Reply("Jarvis 返回空回答，请稍后重试")
		return
	}
	content := toPlainQQText(response.GetAnswer())
	if content == "" {
		log.Warn("Jarvis chat returned empty answer after sanitization")
		assistant.Reply("Jarvis 返回空回答，请稍后重试")
		return
	}
	assistant.Reply(content)
}
