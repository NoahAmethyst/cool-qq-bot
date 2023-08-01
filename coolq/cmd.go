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
	CMDHeart                = "心跳"
	CMDWeibo                = "微博"
	CMD36kr                 = "36"
	CMDWallStreet           = "华尔街"
	CMDZhihu                = "知乎"
	CMDCoin                 = "比特币"
	CMDTrans                = "翻译"
	CMDImage                = "图片"
	CMDOpenReporter         = "开启推送"
	CMDCloseReporter        = "关闭推送"
	CMDSwitchAssistantModel = "模式"
	CMDExist                = "关闭"
	CMDEVN                  = "ENV"
)

var groupCmdList []string
var groupCmdHandlers map[string]func(*CQBot, *message.GroupMessage)

var privateCmdList []string
var privateCmdHandlers map[string]func(bot *CQBot, privateMessage *message.PrivateMessage)

var cmdInfo map[string]string

func init() {
	cmdInfo = map[string]string{
		CMDWeibo:                "拉取微博热搜，可以在命令后添加排名获取指定热搜链接，如：\"#微博 1,2,3\"",
		CMD36kr:                 "拉取36氪热榜",
		CMDZhihu:                "拉取知乎热榜，默认最新10条",
		CMDWallStreet:           "拉取华尔街见闻最新资讯",
		CMDCoin:                 "获取BTC,ETH,BNB最新币价（USD），可以指定币种，如 \"#比特币 ETH\"",
		CMDTrans:                "使用\"#翻译 内容\"来翻译文本，注意：中文默认翻译为英文，非中文默认翻译为中文",
		CMDImage:                "AI作图，使用DELL.2生成图片，你需要提供提示词",
		CMDOpenReporter:         "开启微博、华尔街最新资讯、36氪定时推送",
		CMDCloseReporter:        "关闭微博、华尔街最新资讯、36氪定时推送",
		CMDSwitchAssistantModel: "更换助手模式，0:Chatgpt(默认),1:Bing Chat",
		CMDEVN:                  "设置新的环境变量",
	}
	groupCmdList = []string{
		CMDWeibo,
		CMD36kr,
		CMDZhihu,
		CMDWallStreet,
		CMDCoin,
		CMDTrans,
		CMDImage,
		CMDSwitchAssistantModel,
		CMDOpenReporter,
		CMDCloseReporter,
	}

	groupCmdHandlers = map[string]func(bot *CQBot, groupMessage *message.GroupMessage){
		CMDHeart: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.SendGroupMessage(groupMessage.Chat(), &message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("存活")}})
		},
		CMDWeibo: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.groupWeiboHot(groupMessage)
		},
		CMD36kr: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.Report36kr([]int64{groupMessage.Chat()}, true)
		},
		CMDWallStreet: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportWallStreetNews([]int64{groupMessage.Chat()}, true)
		},
		CMDZhihu: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportZhihuHot(groupMessage.Chat(), true)
		},
		CMDCoin: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportCoinPrice(groupMessage.Chat(), groupMessage.GetElements(), true)
		},
		CMDTrans: func(bot *CQBot, groupMessage *message.GroupMessage) {
			TransText(&GroupTranslator{
				bot: bot,
				m:   groupMessage,
			})
		},
		CMDImage: func(bot *CQBot, groupMessage *message.GroupMessage) {
			GenerateImage(&GroupImgGenerator{
				bot: bot,
				m:   groupMessage,
			})
		},

		CMDSwitchAssistantModel: func(bot *CQBot, groupMessage *message.GroupMessage) {
			ChangeModel(&GroupAssistant{
				bot: bot,
				m:   groupMessage,
			})
		},
		CMDOpenReporter: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.openReporter(groupMessage.Chat(), true)
		},
		CMDCloseReporter: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.closeReporter(groupMessage.Chat(), true)
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
		CMDSwitchAssistantModel,
		CMDOpenReporter,
		CMDCloseReporter,
	}

	privateCmdHandlers = map[string]func(bot *CQBot, privateMessage *message.PrivateMessage){
		CMDWeibo: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.privateWeiboHot(privateMessage)
		},
		CMD36kr: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.Report36kr([]int64{privateMessage.Chat()}, false)
		},
		CMDWallStreet: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			if !bot.ReportWallStreetNews([]int64{privateMessage.Chat()}, false) {
				bot.SendPrivateMessage(privateMessage.Chat(), 0,
					&message.SendingMessage{Elements: []message.IMessageElement{message.NewText("没有华尔街最新资讯")}})
			}
		},
		CMDZhihu: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportZhihuHot(privateMessage.Chat(), false)
		},
		CMDCoin: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportCoinPrice(privateMessage.Chat(), privateMessage.GetElements(), false)
		},
		CMDTrans: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			TransText(&PrivateTranslator{
				bot: bot,
				m:   privateMessage,
			})
		},
		CMDImage: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			GenerateImage(&PrivateImgGenerator{
				bot: bot,
				m:   privateMessage,
			})
		},

		CMDSwitchAssistantModel: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			ChangeModel(&PrivateAssistant{
				bot: bot,
				m:   privateMessage,
			})
		},

		CMDOpenReporter: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.openReporter(privateMessage.Chat(), false)
		},

		CMDCloseReporter: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.closeReporter(privateMessage.Chat(), false)
		},
		CMDExist: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			if bot.state.owner != privateMessage.Chat() {
				bot.SendPrivateMessage(privateMessage.Chat(), 0, &message.SendingMessage{
					Elements: []message.IMessageElement{
						message.NewText(fmt.Sprintf("不是 %s 所有者，无法执行此命令", bot.Client.Nickname))},
				})
			}
			log.Warnf("收到关闭命令，来源[%s][%d]", privateMessage.Sender.Nickname, privateMessage.Chat())
			bot.SendPrivateMessage(privateMessage.Chat(), 0, &message.SendingMessage{
				Elements: []message.IMessageElement{
					message.NewText(fmt.Sprintf("%s 正在关闭...", bot.Client.Nickname))},
			})
			os.Exit(0)
		},

		CMDEVN: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			if bot.state.owner != privateMessage.Chat() {
				bot.SendPrivateMessage(privateMessage.Chat(), 0, &message.SendingMessage{
					Elements: []message.IMessageElement{
						message.NewText(fmt.Sprintf("不是 %s 所有者，无法执行此命令", bot.Client.Nickname))},
				})
			}
			bot.SetENV(privateMessage)
		},
	}
}

// 群命令 - 描述
func (bot *CQBot) reactGroupCmd(_ *client.QQClient, m *message.GroupMessage) {
	texts := m.Texts()

	if len(texts) == 0 || !strings.Contains(texts[0], "#") {
		return
	}

	if texts[0] == "#" {
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
	match := re.FindStringSubmatch(texts[0])

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
		bot.SendGroupMessage(m.Chat(), &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText("该命令无效")}})
	}
}

// 私聊命令 - 描述
func (bot *CQBot) reactPrivateCmd(_ *client.QQClient, m *message.PrivateMessage) {
	texts := m.Texts()

	if len(texts) == 0 || !strings.Contains(texts[0], "#") {
		return
	}

	if texts[0] == "#" {
		content := ""
		for _, _cmd := range privateCmdList {
			content += fmt.Sprintf("#%s\t%s\n\n", _cmd, cmdInfo[_cmd])
		}

		bot.SendPrivateMessage(m.Chat(), 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(fmt.Sprintf("你可以使用以下命令：\n\n%s", content))}})
		return
	}

	re := regexp.MustCompile(`^#(\S+)`)

	// 匹配字符串
	match := re.FindStringSubmatch(texts[0])

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
		bot.SendPrivateMessage(m.Chat(), 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText("该命令无效")}})
	}
}
