package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/go-cqhttp/util/trans"
	translator_engine "github.com/NoahAmethyst/translator-engine"
	"regexp"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/coin"
	"github.com/Mrs4s/go-cqhttp/util/openai_util"
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func (bot *CQBot) askChatGptInPrivate(_ *client.QQClient, m *message.PrivateMessage) {
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if textEle == nil || !strings.Contains(textEle.Content, "?") {
		return
	}
	answer := askChatGpt(textEle)
	bot.SendPrivateMessage(m.Sender.Uin, 0, &message.SendingMessage{Elements: []message.IMessageElement{
		message.NewText(answer)}})
}

func (bot *CQBot) askChatGptInGroup(_ *client.QQClient, m *message.GroupMessage) {
	var atEle *message.AtElement
	var textEle *message.TextElement
	for _, _ele := range m.Elements {
		switch _ele.Type() {
		case message.At:
			atEle = _ele.(*message.AtElement)
		case message.Text:
			textEle = _ele.(*message.TextElement)
		default:

		}
	}

	if atEle == nil || textEle == nil {
		return
	}

	if atEle.Target != bot.Client.Uin {
		log.Warnf("mention target is not bot")
		return
	}

	answer := askChatGpt(textEle)

	bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{message.NewReply(m),
		message.NewText(answer)}})

}

func askChatGpt(textEle *message.TextElement) string {
	answer, err := openai_util.AskChatGpt(textEle.Content)

	//重试机制
	if err != nil {
		maxRetry := 6
		for i := 0; i < maxRetry; i++ {
			time.Sleep(500 * time.Millisecond)
			log.Warnf("call openai failed cause:%s,retry:%d", err.Error(), i+1)
			answer, err = openai_util.AskChatGpt(textEle.Content)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		answer = fmt.Sprintf("调用openAi 失败：%s", err.Error())
	}
	return answer
}

func (bot *CQBot) ReportCoinPrice(group int64, isGroup bool) {

	coinContent := fmt.Sprintf("%s 币价实时信息", time.Now().Format("2006-01-02 15:04"))
	var priceContents []string
	for _, symbol := range coin.Symbols {
		coinInfo, err := coin.Get24HPriceInfo(symbol)
		if err != nil {
			log.Error("get %s error:%s", symbol, err)
			continue
		}
		priceContents = append(priceContents, fmt.Sprintf("\n%s \n价格：%s\n24小时涨跌幅：%s%% \n最高价：%s \n最低价：%s\n",
			coinInfo.Symbol,
			strings.ReplaceAll(coinInfo.LastPrice, "000", ""),
			coinInfo.PriceChangePercent,
			strings.ReplaceAll(coinInfo.HighPrice, "000", ""),
			strings.ReplaceAll(coinInfo.LowPrice, "000", "")))
	}

	if len(priceContents) == 0 {
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("获取币价实时信息失败，请查看日志")}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("获取币价实时信息失败，请查看日志")}})
		}

	} else {
		for _, _content := range priceContents {
			coinContent += _content
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(coinContent)}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(coinContent)}})
		}

	}

}

func (bot *CQBot) ReportWeiboHot(group int64, isGroup bool) {
	hotContent := fmt.Sprintf("%s 微博实时热搜\n", time.Now().Format("2006-01-02 15:04"))
	if hotList, err := top_list.LoadWeiboHot(); err != nil {
		log.Errorf("get hot list error:%s", err.Error())
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取微博热搜失败：%s", err.Error()))}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取微博热搜失败：%s", err.Error()))}})
		}

	} else {
		for _, hot := range hotList {
			hotContent += fmt.Sprintf("%d	%s\n", hot.Rank, hot.Title)
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		}

	}
}

func (bot *CQBot) Report36kr(group int64, isGroup bool) {
	hotContent := fmt.Sprintf("%s 36氪24H热榜\n", time.Now().Format("2006-01-02 15:04"))
	if hotList, err := top_list.Load36krHot(); err != nil {
		log.Errorf("get hot list error:%s", err.Error())
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取36氪热榜失败：%s", err.Error()))}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取36氪热榜失败：%s", err.Error()))}})
		}

	} else {
		for _i, _hot := range hotList {
			if _i > 10 {
				break
			}
			hotContent += fmt.Sprintf("%d	%s\n%s\n\n", _hot.Rank, _hot.Title, _hot.Url)
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		}

	}
}

func (bot *CQBot) ReportWallStreetNews(group int64, isGroup bool) {
	if hotList, err := top_list.LoadWallStreetNews(); err != nil {
		log.Errorf("爬取华尔街见闻最新资讯失败：%s", err.Error())
		//bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
		//	fmt.Sprintf("爬取华尔街见闻最新资讯失败：%s", err.Error()))}})
	} else {
		//倒序输出，因为最新资讯在第一个
		for i := len(hotList) - 1; i >= 0; i-- {
			content := fmt.Sprintf("%s\n\n摘要：%s\n\n链接：%s", hotList[i].Title, hotList[i].Content, hotList[i].Url)
			if isGroup {
				bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			} else {
				bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			}

			time.Sleep(3 * time.Second)
		}
	}
}

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
	re := regexp.MustCompile(`^#(\S+)\s(.*)$`)

	match := re.FindStringSubmatch(textEle.Content)

	if len(match) != 3 {
		return "", true
	}

	text := strings.TrimSpace(match[2])
	return text, false
}

func isChinese(str string) bool {
	re := regexp.MustCompile("[\u4e00-\u9fa5]")
	return re.MatchString(str)
}
