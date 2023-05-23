package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/coin"
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	"github.com/tristan-club/kit/log"

	"strings"
	"time"
)

func (bot *CQBot) ReportCoinPrice(group int64, isGroup bool) {
	var coinContent strings.Builder
	coinContent.WriteString(fmt.Sprintf("%s 币价实时信息", time.Now().Format("2006-01-02 15:04")))

	var priceContents []string
	for _, symbol := range coin.Symbols {
		coinInfo, err := coin.Get24HPriceInfo(symbol)
		if err != nil {
			log.Error().Msgf("get %s error:%s", symbol, err)
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
			coinContent.WriteString(_content)
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(coinContent.String())}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(coinContent.String())}})
		}

	}

}

func (bot *CQBot) ReportWeiboHot(group int64, isGroup bool) {
	var hotContent strings.Builder
	hotContent.WriteString(fmt.Sprintf("%s 微博实时热搜\n", time.Now().Format("2006-01-02 15:04")))
	if hotList, err := top_list.LoadWeiboHot(); err != nil {
		log.Error().Msgf("get hot list error:%s", err.Error())
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取微博热搜失败：%s", err.Error()))}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取微博热搜失败：%s", err.Error()))}})
		}

	} else {
		for _, hot := range hotList {
			hotContent.WriteString(fmt.Sprintf("%d	%s\n", hot.Rank, hot.Title))
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent.String())}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent.String())}})
		}

	}
}

func (bot *CQBot) Report36kr(group int64, isGroup bool) {
	var hotContent strings.Builder
	hotContent.WriteString(fmt.Sprintf("%s 36氪24H热榜\n", time.Now().Format("2006-01-02 15:04")))
	if hotList, err := top_list.Load36krHot(); err != nil {
		log.Error().Msgf("get hot list error:%s", err.Error())
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
			hotContent.WriteString(fmt.Sprintf("%d	%s\n%s\n\n", _hot.Rank, _hot.Title, _hot.Url))
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent.String())}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent.String())}})
		}

	}
}

func (bot *CQBot) ReportWallStreetNews(group int64, isGroup bool) {
	if hotList, err := top_list.LoadWallStreetNews(); err != nil {
		log.Error().Msgf("爬取华尔街见闻最新资讯失败：%s", err.Error())
		//bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
		//	fmt.Sprintf("爬取华尔街见闻最新资讯失败：%s", err.Error()))}})
	} else {
		readyData := make([]top_list.WallStreetNews, 0, 10)
		for _, _data := range hotList {
			if !bot.state.wallstreetSentNews.checkSent(group, _data.Title) {
				readyData = append(readyData, _data)
			}
		}

		if len(readyData) == 0 {
			log.Warn().Msgf("华尔街见闻：没有最新资讯，爬取资讯数量:%d", len(hotList))
			if !isGroup {
				bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("没有华尔街最新资讯")}})
			}
		} else {
			//倒序输出，因为最新资讯在第一个
			for i := len(readyData) - 1; i >= 0; i-- {
				bot.state.wallstreetSentNews.add(group, readyData[i].Title)
				content := fmt.Sprintf("华尔街资讯：%s\n\n摘要：%s\n\n链接：%s", readyData[i].Title, readyData[i].Content, readyData[i].Url)
				if isGroup {
					bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
				} else {
					bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
				}
				time.Sleep(1 * time.Second)
			}
			bot.state.wallstreetSentNews.SaveCache()
		}
	}
}
