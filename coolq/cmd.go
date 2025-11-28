package coolq

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	log "github.com/sirupsen/logrus"
)

const (
	CMDHeart                = "心跳"
	CMDWeibo                = "微博"
	CMD36kr                 = "36"
	CMDWallStreet           = "华尔街"
	CMDCaixin               = "财新"
	CMDZhihu                = "知乎"
	CMDCoin                 = "比特币"
	CMDTrans                = "翻译"
	CMDImage                = "图片"
	CMDOpenReporter         = "开启推送"
	CMDCloseReporter        = "关闭推送"
	CMDSwitchAssistantModel = "模式"
	CMDExist                = "关闭"
	CMDEVN                  = "ENV"
	CMDKelly                = "凯利"
	CMDGold                 = "黄金"
	CMDExchange             = "汇率"
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
		CMDCaixin:               "拉取财新网最新新闻",
		CMDCoin:                 "获取BTC,ETH,BNB最新币价（USD），可以指定币种，如 \"#比特币 ETH\"",
		CMDTrans:                "使用\"#翻译 内容\"来翻译文本，注意：中文默认翻译为英文，非中文默认翻译为中文",
		CMDImage:                "AI作图，使用DELL.2生成图片，你需要提供提示词",
		CMDOpenReporter:         "开启微博、华尔街最新资讯、36氪定时推送",
		CMDCloseReporter:        "关闭微博、华尔街最新资讯、36氪定时推送",
		CMDSwitchAssistantModel: "更换助手模式，\n0:Chatgpt(默认) \n1:Bing Chat \n2:ChatGpt(4) \n3:百度千帆 \n4:DeepSeek",
		CMDEVN:                  "设置新的环境变量",
		CMDKelly:                "使用凯利公式(Kelly strategy)计算投资资金比例，依次输入【潜在正收益率】、【潜在损失率】、【获胜概率/收益概率】，输入数值为概率x100，以空格分隔",
		CMDGold:                 "获取黄金最新价格",
		CMDExchange:             "计算汇率，实用\"#汇率 源币 目标币\" 的格式获取汇率，如不输入则获取支持的汇率币种列表",
	}
	groupCmdList = []string{
		CMDWeibo,
		CMD36kr,
		CMDZhihu,
		CMDWallStreet,
		CMDCaixin,
		CMDCoin,
		CMDTrans,
		CMDImage,
		CMDSwitchAssistantModel,
		CMDOpenReporter,
		CMDCloseReporter,
		CMDKelly,
		CMDGold,
		CMDExchange,
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
			bot.Report36kr([]int64{groupMessage.GroupCode}, true)
		},
		CMDWallStreet: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportWallStreetNews([]int64{groupMessage.GroupCode}, true)
		},
		CMDCaixin: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportCaiXinNews([]int64{groupMessage.GroupCode}, true)
		},
		CMDZhihu: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportZhihuHot(groupMessage.GroupCode, true)
		},
		CMDCoin: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.ReportCoinPrice(groupMessage.GroupCode, groupMessage.Elements, true)
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
			bot.openReporter(groupMessage.GroupCode, true)
		},
		CMDCloseReporter: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.closeReporter(groupMessage.GroupCode, true)
		},
		CMDKelly: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.kellyStrategyForGroup(groupMessage)
		},
		CMDGold: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.goldPriceForGroup(groupMessage)
		},
		CMDExchange: func(bot *CQBot, groupMessage *message.GroupMessage) {
			bot.exChangeRateForGroup(groupMessage)
		},
	}

	privateCmdList = []string{
		CMDWeibo,
		CMD36kr,
		CMDWallStreet,
		CMDCaixin,
		CMDZhihu,
		CMDCoin,
		CMDTrans,
		CMDImage,
		CMDSwitchAssistantModel,
		CMDOpenReporter,
		CMDCloseReporter,
		CMDKelly,
	}

	privateCmdHandlers = map[string]func(bot *CQBot, privateMessage *message.PrivateMessage){
		CMDWeibo: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.privateWeiboHot(privateMessage)
		},
		CMD36kr: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.Report36kr([]int64{privateMessage.Sender.Uin}, false)
		},
		CMDWallStreet: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			if !bot.ReportWallStreetNews([]int64{privateMessage.Sender.Uin}, false) {
				bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
					&message.SendingMessage{Elements: []message.IMessageElement{message.NewText("没有华尔街最新资讯")}})
			}
		},
		CMDCaixin: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			if !bot.ReportCaiXinNews([]int64{privateMessage.Sender.Uin}, false) {
				bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
					&message.SendingMessage{Elements: []message.IMessageElement{message.NewText("没有财新网最新新闻")}})
			}
		},
		CMDZhihu: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportZhihuHot(privateMessage.Sender.Uin, false)
		},
		CMDCoin: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.ReportCoinPrice(privateMessage.Sender.Uin, privateMessage.Elements, false)
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
			bot.openReporter(privateMessage.Sender.Uin, false)
		},

		CMDCloseReporter: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.closeReporter(privateMessage.Sender.Uin, false)
		},
		CMDKelly: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.kellyStrategyForPrivate(privateMessage)
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

		CMDEVN: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			if bot.state.owner != privateMessage.Sender.Uin {
				bot.SendPrivateMessage(privateMessage.Sender.Uin, 0, &message.SendingMessage{
					Elements: []message.IMessageElement{
						message.NewText(fmt.Sprintf("不是 %s 所有者，无法执行此命令", bot.Client.Nickname))},
				})
			}
			bot.SetENV(privateMessage)
		},
		CMDGold: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.goldPriceForPrivate(privateMessage)
		},
		CMDExchange: func(bot *CQBot, privateMessage *message.PrivateMessage) {
			bot.exChangeRateForPrivate(privateMessage)
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
		var content strings.Builder
		for _, _cmd := range groupCmdList {
			content.WriteString(fmt.Sprintf("#%s\t%s\n\n", _cmd, cmdInfo[_cmd]))
		}

		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(fmt.Sprintf("你可以使用以下命令：\n\n%s", content.String()))}})
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

	log.Infof("接收到命令 [%s]", cmd)

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
		var content strings.Builder
		for _, _cmd := range privateCmdList {
			content.WriteString(fmt.Sprintf("#%s\t%s\n\n", _cmd, cmdInfo[_cmd]))
		}

		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(fmt.Sprintf("你可以使用以下命令：\n\n%s", content.String()))}})
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

	log.Infof("接收到命令 [%s]", cmd)

	handler, ok := privateCmdHandlers[cmd]
	if ok {
		handler(bot, m)
	} else {
		bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText("该命令无效")}})
	}
}
