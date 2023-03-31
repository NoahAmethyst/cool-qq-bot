package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/coin"
	"github.com/Mrs4s/go-cqhttp/util/openai_util"
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func (bot *CQBot) askChatGpt(_ *client.QQClient, m *message.GroupMessage) {
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

	bot.SendGroupMessage(m.GroupCode, &message.SendingMessage{Elements: []message.IMessageElement{message.NewReply(m),
		message.NewText(answer)}})

}

func (bot *CQBot) ReportWeiboHot(group int64) (func(), string) {

	f := func() {
		hotContent := fmt.Sprintf("%s 微博实时热搜\n", time.Now().Format("2006-01-02 15:04:05"))
		if hotList, err := top_list.LoadWeiboHot(); err != nil {
			log.Errorf("get hot list error:%s", err.Error())
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取微博热搜失败：", err.Error()))}})
		} else {
			for _, hot := range hotList {
				hotContent += fmt.Sprintf("%d	%s\n", hot.Rank, hot.Title)
			}
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		}

	}
	return f, top_list.Weibo
}

func (bot *CQBot) Report36krHot(group int64) (func(), string) {
	f := func() {
		hotContent := fmt.Sprintf("%s 36氪24H热榜\n", time.Now().Format("2006-01-02 15:04:05"))
		if hotList, err := top_list.Load36krHot(); err != nil {
			log.Errorf("get hot list error:%s", err.Error())
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取36氪热榜失败：", err.Error()))}})
		} else {
			for _, _hot := range hotList {
				hotContent += fmt.Sprintf("%d	%s\n%s\n", _hot.Rank, _hot.Title, _hot.Url)
			}
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		}
	}
	return f, top_list.D36kr
}

func (bot *CQBot) ReportWallStreetNews(group int64) (func(), string) {
	f := func() {
		if hotList, err := top_list.LoadWallStreetNews(); err != nil {
			log.Errorf("get hot list error:%s", err.Error())
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取华尔街见闻最新资讯失败：", err.Error()))}})
		} else {
			for _, _hot := range hotList {
				bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
					fmt.Sprintf("%s\n摘要：%s\n链接：%s", _hot.Title, _hot.Content, _hot.Url))}})
			}
		}
	}

	return f, top_list.WallStreet
}

func (bot *CQBot) ReportCoinPrice(group int64) {

	coinContent := fmt.Sprintf("%s 币价实时信息", time.Now().Format("2006-01-02 15:04:05"))
	for _, symbol := range coin.Symbols {
		coinInfo, err := coin.Get24HPriceInfo(symbol)
		if err != nil {
			log.Error("get %s error:%s", symbol, err)
			continue
		}
		formatContent := fmt.Sprintf("\n%s \n价格：%s\n24小时涨跌幅：%s%% \n最高价：%s \n最低价：%s\n",
			coinInfo.Symbol,
			strings.ReplaceAll(coinInfo.LastPrice, "000", ""),
			coinInfo.PriceChangePercent,
			strings.ReplaceAll(coinInfo.HighPrice, "000", ""),
			strings.ReplaceAll(coinInfo.LowPrice, "000", ""))

		coinContent += formatContent
	}
	bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(coinContent)}})
}
