package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	"github.com/NoahAmethyst/go-cqhttp/util/ai_util"
	go_ernie "github.com/anhao/go-ernie"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

var chatModelHandlers map[ai_util.ChatModel]func(assistant Assistant, recvChan chan struct{})

type Assistant interface {
	Reply(content string)
	GetText() *message.TextElement
	Mention() *message.AtElement
	Check() bool
	Chat() int64
	Sender() int64
	Model() ai_util.ChatModel
	ChangeModel(model ai_util.ChatModel)
	Session() *AiAssistantSession
	Me() int64
	Bot() *CQBot
}

type PrivateAssistant struct {
	bot *CQBot
	m   *message.PrivateMessage
}

func (p *PrivateAssistant) Reply(msg string) {
	p.bot.SendPrivateMessage(p.Chat(), 0, &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(
			msg)}})
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
	var mentionEle *message.AtElement
	for _, _ele := range p.m.Elements {
		switch _ele.Type() {
		case message.At:
			mentionEle = _ele.(*message.AtElement)
		default:
			mentionEle = &message.AtElement{}
		}
	}
	return mentionEle
}

func (p *PrivateAssistant) Me() int64 {
	return p.bot.Client.Uin
}

func (p *PrivateAssistant) Model() ai_util.ChatModel {
	return p.bot.state.assistantModel.getModel(p.m.Sender.Uin)
}

func (p *PrivateAssistant) ChangeModel(model ai_util.ChatModel) {
	p.bot.state.assistantModel.setModel(p.Sender(), model)

}

func (p *PrivateAssistant) Session() *AiAssistantSession {
	return p.bot.state.privateDialogueSession
}

type GroupAssistant struct {
	bot *CQBot
	m   *message.GroupMessage
}

func (p *GroupAssistant) Reply(msg string) {

	p.bot.SendGroupMessage(p.Chat(), &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewReply(p.m),
		message.NewText(
			msg)}})
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
	var mentionEle *message.AtElement
	for _, _ele := range p.m.Elements {
		switch _ele.Type() {
		case message.At:
			mentionEle = _ele.(*message.AtElement)
		default:
			mentionEle = &message.AtElement{}
		}
	}
	return mentionEle
}

func (p *GroupAssistant) Me() int64 {
	return p.bot.Client.Uin
}

func (p *GroupAssistant) Model() ai_util.ChatModel {
	return p.bot.state.assistantModel.getModel(p.m.Sender.Uin)
}

func (p *GroupAssistant) ChangeModel(model ai_util.ChatModel) {
	p.bot.state.assistantModel.setModel(p.Sender(), model)

}

func (p *GroupAssistant) Session() *AiAssistantSession {
	return p.bot.state.groupDialogueSession
}

func ChangeModel(assistant Assistant) {
	if assistant == nil || !assistant.Check() {
		log.Warnf("invalid image generator")
		return
	}

	textEle := assistant.GetText()
	if textEle == nil {
		return
	}
	v := strings.TrimSpace(strings.ReplaceAll(textEle.Content, "#模式 ", ""))

	var currModel string
	switch assistant.Model() {
	case ai_util.BingCopilot:
		currModel = "Bing Copilot"
	case ai_util.ChatGPT4:
		currModel = "ChatGpt4.0"
	case ai_util.Ernie:
		currModel = "文心千帆"
	default:
		currModel = "ChatGpt3.5"
	}
	switchModelMsg := fmt.Sprintf("如需更换模式请使用:\n"+
		"%d - ChatGpt3.5(默认)\n"+
		"%d - Bing Copilot\n"+
		"%d - ChatGpt4.0\n"+
		"%d - 文心千帆", ai_util.ChatGPT, ai_util.BingCopilot, ai_util.ChatGPT4, ai_util.Ernie)

	if len(v) == 0 {
		msg := fmt.Sprintf("当前模式：%s\n%s", currModel, switchModelMsg)
		assistant.Reply(msg)
		return
	}

	if model, err := strconv.ParseInt(v, 10, 64); err != nil {
		msg := fmt.Sprintf("非法的参数\n当前模式：%s\n%s", currModel, switchModelMsg)
		assistant.Reply(msg)
		return
	} else {
		var msg string
		switch model {
		case int64(ai_util.ChatGPT):
			currModel = "ChatGpt3.5"
			msg = fmt.Sprintf("更换模式为：%s\n%s", currModel, switchModelMsg)
			assistant.ChangeModel(ai_util.ChatGPT)
		case int64(ai_util.BingCopilot):
			currModel = "Bing Copilot"
			msg = fmt.Sprintf("更换模式为：%s\n%s", currModel, switchModelMsg)
			assistant.ChangeModel(ai_util.BingCopilot)
		case int64(ai_util.ChatGPT4):
			currModel = "ChatGpt4.0"
			msg = fmt.Sprintf("更换模式为：%s\n%s", currModel, switchModelMsg)
			assistant.ChangeModel(ai_util.ChatGPT4)
		case int64(ai_util.Ernie):
			currModel = "文心千帆"
			msg = fmt.Sprintf("更换模式为：%s\n%s", currModel, switchModelMsg)
			assistant.ChangeModel(ai_util.Ernie)

		default:
			msg = fmt.Sprintf("非法的参数\n当前模式%s\n%s", currModel, switchModelMsg)
		}
		assistant.Reply(msg)
	}
}

func AskAssistant(assistant Assistant) {
	if assistant == nil || !assistant.Check() {
		log.Warnf("invalid image generator")
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

	askHandler := chatModelHandlers[assistant.Model()]
	if askHandler == nil {
		log.Errorf("no handler set,model:%d", assistant.Model())
		assistant.Reply(fmt.Sprintf("no handler set,model:%d", assistant.Model()))
		return
	}

	recvChan := make(chan struct{}, 1)
	go func(assistant Assistant) {
		select {
		case <-recvChan:
			return
		case <-time.After(time.Second * 10):
			vendor := "OpenAI"
			if assistant.Model() == ai_util.BingCopilot {
				vendor = "BingCopilot"
			}
			assistant.Reply(fmt.Sprintf("%s 正在响应，请稍后...", vendor))
		}
	}(assistant)

	askHandler(assistant, recvChan)

}

func askBingCopilot(assistant Assistant, _ chan struct{}) {
	answer, err := spider_svc.AskBingCopilot(assistant.GetText().Content)
	if err != nil {
		assistant.Reply(fmt.Sprintf("创建 bing copilot 会话失败:%s", err.Error()))
		return
	}

	log.Info("Got Bing Copilot answer:%+v", answer.CopilotResp)

	var strBuilder strings.Builder
	content := strings.ReplaceAll(answer.CopilotResp.Content, "*", "")
	strBuilder.WriteString(content)
	if len(answer.CopilotResp.Suggestions) > 0 {
		strBuilder.WriteString("\n\n你也可以这样提问：")
		for i, _suggestion := range answer.CopilotResp.Suggestions {
			strBuilder.WriteString(fmt.Sprintf("\n%d. %s", i, _suggestion))
		}
	}
	assistant.Reply(strBuilder.String())

}

func askBingChat(assistant Assistant, recvChan chan struct{}) {
	defer close(recvChan)
	var err error
	bingChatCli := assistant.Session().getConversation(assistant.Sender())
	if bingChatCli == nil {
		bingChatCli, err = ai_util.NewBingChat()
	}
	if err != nil {
		assistant.Reply(fmt.Sprintf("创建bingchat会话失败:%s", err.Error()))
		return
	}
	answer, err := ai_util.AskBingChat(bingChatCli, assistant.GetText().Content)
	recvChan <- struct{}{}
	if err != nil {
		assistant.Reply(fmt.Sprintf("询问bingchat失败:%s", err.Error()))
		assistant.Session().closeConversation(assistant.Sender())
		return
	}
	var strBuilder strings.Builder
	strBuilder.WriteString(answer.Answer)

	if len(answer.Reference) > 0 {
		strBuilder.WriteString("\n\n参考资料:")
	}
	for title, link := range answer.Reference {
		strBuilder.WriteString(fmt.Sprintf("\n%s %s", title, link))
	}

	if len(answer.Suggestions) > 0 {
		strBuilder.WriteString("\n\n您也可以这样提问")
	}
	for i, suggest := range answer.Suggestions {
		strBuilder.WriteString(fmt.Sprintf("\n%d: %s", i+1, suggest))
	}

	assistant.Session().putConversation(assistant.Sender(), bingChatCli)

	if len(strBuilder.String()) > 0 {
		assistant.Reply(strBuilder.String())
	} else {
		assistant.Reply("BingCopilot 响应超时")
	}
}

func askOfficialChatGpt(assistant Assistant, recvChan chan struct{}) {
	defer close(recvChan)
	textEle := assistant.GetText()
	ctx := assistant.Session().getOpenaiCtx(assistant.Sender())
	msg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: textEle.Content,
	}
	if len(ctx) == 0 {
		ctx = []openai.ChatCompletionMessage{
			msg,
		}
	} else {
		ctx = append(ctx, msg)
	}

	var answer string
	var resp openai.ChatCompletionResponse
	var err error
	switch assistant.Model() {
	case ai_util.ChatGPT:
		resp, err = ai_util.AskChatGpt(ctx)
	case ai_util.ChatGPT4:
		resp, err = ai_util.AskChatGpt4(ctx)
	default:
		resp, err = ai_util.AskChatGpt(ctx)
	}

	if err != nil {
		answer = fmt.Sprintf("调用openAi 失败：%s", err.Error())
		if strings.Contains(err.Error(), "401") && assistant.Sender() != assistant.Bot().state.owner {
			assistant.Bot().SendPrivateMessage(assistant.Bot().state.owner, 0, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText(
					fmt.Sprintf("用户[%d]调用openai失败：%s", assistant.Sender(), err.Error()))}})
		}
	} else {
		if len(resp.Choices) == 0 || len(resp.Choices[0].Message.Content) == 0 {
			log.Warnf("openai 返回空结构：%v", resp)
			answer = "OpenAI未响应，请重试"
		} else {
			answer = resp.Choices[0].Message.Content
			assistant.Session().putOpenaiCtx(assistant.Sender(), msg.Content, answer)
		}
	}
	recvChan <- struct{}{}
	assistant.Reply(answer)
}

func askErnie(assistant Assistant, recvChan chan struct{}) {
	defer close(recvChan)
	textEle := assistant.GetText()
	ctx := assistant.Session().getErnieCtx(assistant.Sender())
	msg := go_ernie.ChatCompletionMessage{
		Role:    go_ernie.MessageRoleUser,
		Content: textEle.Content,
	}
	if len(ctx) == 0 {
		ctx = []go_ernie.ChatCompletionMessage{
			msg,
		}
	} else {
		ctx = append(ctx, msg)
	}

	var answer string
	resp, err := ai_util.AskErnie(ctx)

	if err != nil {
		answer = fmt.Sprintf("调用文心千帆 失败：%s", err.Error())
	} else {
		if len(resp.Result) == 0 {
			log.Warnf("文心千帆 返回空结构：%+v", resp)
			answer = "文心千帆未响应，请重试"
		} else {
			answer = resp.Result
			assistant.Session().putErnieCtx(assistant.Sender(), msg.Content, answer)
		}
	}
	recvChan <- struct{}{}
	assistant.Reply(answer)
}

func init() {
	chatModelHandlers = map[ai_util.ChatModel]func(assistant Assistant, recvChan chan struct{}){
		//ai_util.ChatGPT:  askRemoteChatGpt,
		ai_util.ChatGPT:     askOfficialChatGpt,
		ai_util.ChatGPT4:    askOfficialChatGpt,
		ai_util.BingCopilot: askBingCopilot,
		ai_util.Ernie:       askErnie,
	}
}
