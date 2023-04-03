package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

const (
	CMDHeart      = "心跳"
	CMDWeibo      = "微博"
	CMD36氪        = "36"
	CMDWallStreet = "华尔街"
	CMDCoin       = "比特币"
)

var cmdList map[string]string

func init() {
	cmdList = map[string]string{
		CMDHeart:      "心跳检查",
		CMDWeibo:      "拉取微博热搜",
		CMD36氪:        "拉取36氪热榜",
		CMDWallStreet: "拉取华尔街见闻最新资讯",
		CMDCoin:       "获取BTC,ETH,BNB最新币价（USD）",
	}
}

// 命令 - 描述
func (bot *CQBot) reactCmd(_ *client.QQClient, m *message.GroupMessage) {
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
		for _cmd, _desc := range cmdList {
			content += fmt.Sprintf("#%s	%s\n", _cmd, _desc)
		}

		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(fmt.Sprintf("你可以使用以下命令：\n%s", content))}})
		return
	}

	re := regexp.MustCompile(`#(.*)`)

	// 匹配字符串
	match := re.FindStringSubmatch(textEle.Content)

	// 输出匹配结果
	if len(match) == 0 {
		return
	}
	cmd := match[1]

	log.Infof("接收到命令:%s", cmd)

	switch cmd {
	case CMDHeart:
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText("存活")}})
	case CMDWeibo:
		bot.ReportWeiboHot(m.GroupCode)
	case CMD36氪:
		bot.Report36kr(m.GroupCode)
	case CMDWallStreet:
		bot.ReportWallStreetNews(m.GroupCode)
	default:
		bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText("该命令无效")}})
	}

}
