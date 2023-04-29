package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
)

const (
	CMDHeart      = "心跳"
	CMDWeibo      = "微博"
	CMD36kr       = "36"
	CMDWallStreet = "华尔街"
	CMDCoin       = "比特币"
	CMDTrans      = "翻译"
	CMDExist      = "关闭"
)

var groupCmdList []cmdInfo
var groupCmdHandlers map[string]func(*CQBot, *message.GroupMessage)
var privateCmdList []cmdInfo
var privateCmdHandlers map[string]func(bot *CQBot, privateMessage *message.PrivateMessage)

type cmdInfo struct {
	cmd  string
	desc string
}

func init() {
	groupCmdList = []cmdInfo{
		{CMDHeart, "心跳检查"},
		{CMDWeibo, "拉取微博热搜"},
		{CMD36kr, "拉取36氪热榜"},
		{CMDWallStreet, "拉取华尔街见闻最新资讯"},
		{CMDCoin, "获取BTC,ETH,BNB最新币价（USD）"},
		{CMDTrans, "使用\"#翻译 内容\"来翻译文本，注意：中文默认翻译为英文，非中文默认翻译为中文"},
	}

	groupCmdHandlers = map[string]func(bot *CQBot, groupMessage *message.GroupMessage){
		CMDHeart: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.SendGroupMessage(groupMessage.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("存活")}})
		},
		CMDWeibo: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportWeiboHot(groupMessage.GroupCode, true)
		},
		CMD36kr: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.Report36kr(groupMessage.GroupCode, true)
		},
		CMDWallStreet: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportWallStreetNews(groupMessage.GroupCode, true)
		},
		CMDCoin: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportCoinPrice(groupMessage.GroupCode, true)
		},
		CMDTrans: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.TransTextInGroup(groupMessage)
		},
	}

	privateCmdList = []cmdInfo{
		{CMDWeibo, "拉取微博热搜"},
		{CMD36kr, "拉取36氪热榜"},
		{CMDWallStreet, "拉取华尔街见闻最新资讯"},
		{CMDCoin, "获取BTC,ETH,BNB最新币价（USD）"},
		{CMDTrans, "使用\"#翻译 内容\"来翻译文本，注意：中文默认翻译为英文，非中文默认翻译为中文 "},
	}

	privateCmdHandlers = map[string]func(bot *CQBot, privateMessage *message.PrivateMessage){
		CMDWeibo: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportWeiboHot(privateMessage.Sender.Uin, false)
		},
		CMD36kr: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.Report36kr(privateMessage.Sender.Uin, false)
		},
		CMDWallStreet: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportWallStreetNews(privateMessage.Sender.Uin, false)
		},
		CMDCoin: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportCoinPrice(privateMessage.Sender.Uin, false)
		},
		CMDTrans: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.TransTextInPrivate(privateMessage)
		},
		CMDExist: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.SendPrivateMessage(privateMessage.Sender.Uin, 0, &message.SendingMessage{
				Elements: []message.IMessageElement{
					message.NewText("正在关闭Jarvis...")},
			})
			os.Exit(0)
		},
	}
}

// 群命令 - 描述
func (bot *CQBot) reactGroupCmd(_ *client.QQClient, m *message.GroupMessage) {
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if textEle == nil || !strings.Contains(textEle.Content, "#") {
		return
	}

	if textEle.Content == "#" {
		content := ""
		for _, _cmdInfo := range groupCmdList {
			content += fmt.Sprintf("#%s\t%s\n\n", _cmdInfo.cmd, _cmdInfo.desc)
		}

		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(fmt.Sprintf("你可以使用以下命令：\n\n%s", content))}})
		return
	}

	re := regexp.MustCompile(`^#(\S+)`)

	// 匹配字符串
	match := re.FindStringSubmatch(textEle.Content)

	// 输出匹配结果
	if len(match) < 2 {
		return
	}
	cmd := match[1]

	log.Infof("接收到命令:%s", cmd)

	handler, ok := groupCmdHandlers[cmd]
	if ok {
		handler(bot, m)
	} else {
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText("该命令无效")}})
	}
}

// 私聊命令 - 描述
func (bot *CQBot) reactPrivateCmd(_ *client.QQClient, m *message.PrivateMessage) {
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if textEle == nil || !strings.Contains(textEle.Content, "#") {
		return
	}

	if textEle.Content == "#" {
		content := ""
		for _, _cmdInfo := range privateCmdList {
			content += fmt.Sprintf("#%s\t%s\n\n", _cmdInfo.cmd, _cmdInfo.desc)
		}

		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(fmt.Sprintf("你可以使用以下命令：\n\n%s", content))}})
		return
	}

	re := regexp.MustCompile(`^#(\S+)`)

	// 匹配字符串
	match := re.FindStringSubmatch(textEle.Content)

	// 输出匹配结果
	if len(match) < 2 {
		return
	}
	cmd := match[1]

	log.Infof("接收到命令:%s", cmd)

	handler, ok := privateCmdHandlers[cmd]
	if ok {
		handler(bot, m)
	} else {
		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText("该命令无效")}})
	}
}
