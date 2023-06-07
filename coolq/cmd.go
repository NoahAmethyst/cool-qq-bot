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
	CMDHeart         = "心跳"
	CMDWeibo         = "微博"
	CMD36kr          = "36"
	CMDWallStreet    = "华尔街"
	CMDZhihu         = "知乎"
	CMDCoin          = "比特币"
	CMDTrans         = "翻译"
	CMDImage         = "图片"
	CMDOpenReporter  = "开启推送"
	CMDCloseReporter = "关闭推送"
	CMDExist         = "关闭"
)

var groupCmdList []string
var groupCmdHandlers map[string]func(*CQBot, *message.GroupMessage)

var privateCmdList []string
var privateCmdHandlers map[string]func(bot *CQBot, privateMessage *message.PrivateMessage)

var cmdInfo map[string]string

func init() {
	cmdInfo = map[string]string{
		CMDWeibo:         "拉取微博热搜，可以在命令后添加排名获取指定热搜链接，如：\"#微博 1,2,3\"",
		CMD36kr:          "拉取36氪热榜",
		CMDZhihu:         "拉取知乎热榜，默认最新10条",
		CMDWallStreet:    "拉取华尔街见闻最新资讯",
		CMDCoin:          "获取BTC,ETH,BNB最新币价（USD）",
		CMDTrans:         "使用\"#翻译 内容\"来翻译文本，注意：中文默认翻译为英文，非中文默认翻译为中文",
		CMDImage:         "AI作图，使用DELL.2生成图片，你需要提供提示词",
		CMDOpenReporter:  "开启微博、华尔街最新资讯、36氪定时推送",
		CMDCloseReporter: "关闭微博、华尔街最新资讯、36氪定时推送",
	}
	groupCmdList = []string{
		CMDWeibo,
		CMD36kr,
		CMDZhihu,
		CMDWallStreet,
		CMDCoin,
		CMDTrans,
		CMDImage,
		CMDOpenReporter,
		CMDCloseReporter,
	}

	groupCmdHandlers = map[string]func(bot *CQBot, groupMessage *message.GroupMessage){
		CMDHeart: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.SendGroupMessage(groupMessage.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("存活")}})
		},
		CMDWeibo: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.groupWeiboHot(groupMessage)
		},
		CMD36kr: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.Report36kr(groupMessage.GroupCode, true)
		},
		CMDWallStreet: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportWallStreetNews(groupMessage.GroupCode, true)
		},
		CMDZhihu: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportZhihuHot(groupMessage.GroupCode, true)
		},
		CMDCoin: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportCoinPrice(groupMessage.GroupCode, true)
		},
		CMDTrans: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.transTextInGroup(groupMessage)
		},
		CMDImage: func(bot *CQBot, groupMessage *message.GroupMessage) {
			generateImg(&groupImgGenerator{
				bot: bot,
				m:   groupMessage,
			})
		},
		CMDOpenReporter: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.openReporter(groupMessage.GroupCode, true)
		},
		CMDCloseReporter: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.closeReporter(groupMessage.GroupCode, true)
		},
	}

	privateCmdList = []string{
		CMDWeibo,
		CMD36kr,
		CMDWallStreet,
		CMDZhihu,
		CMDCoin,
		CMDTrans,
		CMDImage,
		CMDOpenReporter,
		CMDCloseReporter,
	}

	privateCmdHandlers = map[string]func(bot *CQBot, privateMessage *message.PrivateMessage){
		CMDWeibo: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.privateWeiboHot(privateMessage)
		},
		CMD36kr: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.Report36kr(privateMessage.Sender.Uin, false)
		},
		CMDWallStreet: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			if !bot.ReportWallStreetNews(privateMessage.Sender.Uin, false) {
				bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
					&message.SendingMessage{Elements: []message.IMessageElement{message.NewText("没有华尔街最新资讯")}})
			}
		},
		CMDZhihu: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportZhihuHot(privateMessage.Sender.Uin, false)
		},
		CMDCoin: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportCoinPrice(privateMessage.Sender.Uin, false)
		},
		CMDTrans: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.transTextInPrivate(privateMessage)
		},
		CMDImage: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			generateImg(&privateImgGenerator{
				bot: bot,
				m:   privateMessage,
			})
		},

		CMDOpenReporter: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.openReporter(privateMessage.Sender.Uin, false)
		},

		CMDCloseReporter: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.closeReporter(privateMessage.Sender.Uin, false)
		},
		CMDExist: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			if bot.state.owner != privateMessage.Sender.Uin {
				bot.SendPrivateMessage(privateMessage.Sender.Uin, 0, &message.SendingMessage{
					Elements: []message.IMessageElement{
						message.NewText(fmt.Sprintf("不是 %s 所有者，无法执行此命令", bot.Client.Nickname))},
				})
			}
			log.Warnf("收到关闭命令，来源[%s][%d]", privateMessage.Sender.Nickname, privateMessage.Sender.Uin)
			bot.SendPrivateMessage(privateMessage.Sender.Uin, 0, &message.SendingMessage{
				Elements: []message.IMessageElement{
					message.NewText(fmt.Sprintf("%s 正在关闭...", bot.Client.Nickname))},
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
		for _, _cmd := range groupCmdList {
			content += fmt.Sprintf("#%s\t%s\n\n", _cmd, cmdInfo[_cmd])
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
		for _, _cmd := range privateCmdList {
			content += fmt.Sprintf("#%s\t%s\n\n", _cmd, cmdInfo[_cmd])
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
