package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/math_engine"
)

type Calculator interface {
	Reply(content string)
	Calculate(expression string)
	GetExpression() string
	Check() bool
	Chat() int64
	Sender() int64
	Bot() *CQBot
}

type GroupCalculator struct {
	bot *CQBot
	m   *message.GroupMessage
}

func (c *GroupCalculator) Reply(content string) {
	c.bot.SendGroupMessage(c.Chat(), &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewReply(c.m),
		message.NewText(
			content)}})
}

func (c *GroupCalculator) Calculate(expression string) {
	if result, err := math_engine.Calculate(expression); err != nil {
		c.Reply(fmt.Sprintf("计算表达式出错：%s", err.Error()))
	} else {
		c.Reply(fmt.Sprintf("结果：%+v", result))
	}

}

func (c *GroupCalculator) GetExpression() string {
	var expression string
	var textEle *message.TextElement
	for _, _ele := range c.m.Elements {
		if _ele.Type() == message.Text {
			textEle = _ele.(*message.TextElement)
			break
		}
	}

	if textEle != nil {
		expression = textEle.Content
	}
	return expression
}

func (c *GroupCalculator) Check() bool {
	if c.bot == nil || c.m == nil {
		return false
	}
	return true
}

func (c *GroupCalculator) Chat() int64 {
	return c.m.GroupCode
}

func (c *GroupCalculator) Sender() int64 {
	return c.m.Sender.Uin
}

func (c *GroupCalculator) Bot() *CQBot {
	return c.bot
}

type PrivateCalculator struct {
	bot *CQBot
	m   *message.PrivateMessage
}

func (c *PrivateCalculator) Reply(content string) {
	c.bot.SendPrivateMessage(c.Chat(), 0, &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(
			content)}})
}

func (c *PrivateCalculator) Calculate(expression string) {
	if result, err := math_engine.Calculate(expression); err != nil {
		c.Reply(fmt.Sprintf("计算表达式出错：%s", err.Error()))
	} else {
		c.Reply(fmt.Sprintf("结果：%+v", result))
	}
}

func (c *PrivateCalculator) GetExpression() string {
	var expression string
	var textEle *message.TextElement
	for _, _ele := range c.m.Elements {
		if _ele.Type() == message.Text {
			textEle = _ele.(*message.TextElement)
			break
		}
	}

	if textEle != nil {
		expression = textEle.Content
	}
	return expression
}

func (c *PrivateCalculator) Check() bool {
	if c.bot == nil || c.m == nil {
		return false
	}
	return true
}

func (c *PrivateCalculator) Chat() int64 {
	return c.m.Sender.Uin
}

func (c *PrivateCalculator) Sender() int64 {
	return c.m.Sender.Uin
}

func (c *PrivateCalculator) Bot() *CQBot {
	return c.bot
}

func Calculate(calculator Calculator) {
	if !calculator.Check() {
		return
	}
	expression := calculator.GetExpression()
	if !math_engine.IsMathExpression(expression) {
		return
	}

	calculator.Calculate(expression)
}
