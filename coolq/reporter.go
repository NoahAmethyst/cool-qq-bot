package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/util/coin"
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	"github.com/rs/zerolog/log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (bot *CQBot) openReporter(id int64, isGroup bool) {
	var msg *message.SendingMessage
	if bot.state.reportState.exist(id, isGroup) {
		msg = &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("定时推送已处于开启状态，无需再次开启")}}
	} else {
		msg = &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("定时推送已开启")}}
		bot.state.reportState.add(id, isGroup)
	}

	if isGroup {
		bot.SendGroupMessage(id, msg)
	} else {
		bot.SendPrivateMessage(id, 0, msg)
	}

}

func (bot *CQBot) closeReporter(id int64, isGroup bool) {
	var msg *message.SendingMessage
	if bot.state.reportState.exist(id, isGroup) {
		msg = &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("定时推送已关闭")}}
		bot.state.reportState.del(id, isGroup)
	} else {
		msg = &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("定时推送已处于关闭状态，无需再次关闭")}}
	}

	if isGroup {
		bot.SendGroupMessage(id, msg)
	} else {
		bot.SendPrivateMessage(id, 0, msg)
	}
}

func (bot *CQBot) ReportCoinPrice(group int64, elements []message.IMessageElement, isGroup bool) {
	var textEle *message.TextElement
	for _, ele := range elements {
		switch ele.Type() {
		case message.Text:
			textEle = ele.(*message.TextElement)
			break
		}
	}

	var symbol string
	if textEle != nil {
		symbol, _ = parseSourceText(textEle)
	}

	priceContents := make([]string, 0, 3)

	symbols := make([]string, 0, 3)

	if len(symbol) > 0 {
		symbols = append(symbols, fmt.Sprintf("%sUSDT", symbol))
	} else {
		symbols = coin.Symbols
	}

	for _, _symbol := range symbols {
		coinInfo, err := coin.Get24HPriceInfo(_symbol)
		if err != nil {
			log.Error().Msgf("get %s error:%s", _symbol, err)
			continue
		}
		priceContents = append(priceContents, fmt.Sprintf("\n%s \n价格：%s\n24小时涨跌幅：%s%% \n最高价：%s \n最低价：%s\n",
			coinInfo.Symbol,
			strings.ReplaceAll(coinInfo.LastPrice, "000", ""),
			coinInfo.PriceChangePercent,
			strings.ReplaceAll(coinInfo.HighPrice, "000", ""),
			strings.ReplaceAll(coinInfo.LowPrice, "000", "")))
	}

	var resp string
	if len(priceContents) == 0 {
		resp = "获取币价实时信息失败，请查看日志"
	} else {
		var coinContent strings.Builder
		coinContent.WriteString(fmt.Sprintf("%s 币价实时信息", time.Now().Format("2006-01-02 15:04")))
		for _, _content := range priceContents {
			coinContent.WriteString(_content)
		}
		resp = coinContent.String()
	}

	if isGroup {
		bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
	} else {
		bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
	}
}

func (bot *CQBot) privateWeiboHot(privateMessage *message.PrivateMessage) {

	content := privateMessage.ToString()

	params := parseParam(content)
	if len(params) == 2 {
		indexList := parseIndexList(params)
		bot.ReportSpecificWeibo(privateMessage.Sender.Uin, indexList, false)
	} else {
		bot.ReportWeiboHot([]int64{privateMessage.Sender.Uin}, false)
	}

}

func (bot *CQBot) groupWeiboHot(groupMessage *message.GroupMessage) {
	content := groupMessage.ToString()

	params := parseParam(content)
	if len(params) == 2 {
		indexList := parseIndexList(params)
		bot.ReportSpecificWeibo(groupMessage.GroupCode, indexList, true)
	} else {
		bot.ReportWeiboHot([]int64{groupMessage.GroupCode}, true)
	}
}

func parseIndexList(params []string) []int64 {
	_indexList := strings.TrimSpace(params[1])
	s1 := strings.Split(_indexList, ",")
	s2 := strings.Split(_indexList, "，")
	indexSet := make(map[int64]struct{})
	for _, _index := range s1 {
		if strings.Contains(_index, "，") {
			continue
		}
		index, err := strconv.ParseInt(_index, 10, 64)
		if err != nil {
			index = 0
		}
		if index-1 < 0 {
			index = 0
		}
		indexSet[index-1] = struct{}{}
	}

	for _, _index := range s2 {
		if strings.Contains(_index, ",") {
			continue
		}
		index, err := strconv.ParseInt(_index, 10, 64)
		if err != nil {
			index = 0
		}
		if index-1 < 0 {
			index = 0
		}
		indexSet[index-1] = struct{}{}
	}

	indexList := make([]int64, 0, len(indexSet))
	for _index := range indexSet {
		indexList = append(indexList, _index)
	}

	sort.SliceIsSorted(indexList, func(i, j int) bool {
		return indexList[i] < indexList[j]
	})

	return indexList
}

func parseParam(content string) []string {
	re := regexp.MustCompile(`#(\S+)\s+(?s)(.*)`)

	match := re.FindStringSubmatch(content)

	params := make([]string, 0, 2)

	if len(match) > 1 {
		params = match[1:]
	}

	return params
}

func (bot *CQBot) ReportSpecificWeibo(group int64, indexList []int64, isGroup bool) {
	layout := "2006-01-02 15:04"
	data := top_list.WeiboHotDailyRecord.GetData()
	var lastestT time.Time
	for t := range data {
		reportTime, _ := time.Parse(layout, t)
		if lastestT.Before(reportTime) {
			lastestT = reportTime
		}
	}

	k := lastestT.Format(layout)
	if _data, ok := data[k]; !ok {
		log.Warn().Msgf("can't get latest weibo daily report")
		bot.ReportWeiboHot([]int64{group}, isGroup)
	} else {
		for _index := range indexList {
			content := fmt.Sprintf("微博热搜#%d\n%s\n链接:%s", _index+1, _data[_index].Title, _data[_index].Url)
			if isGroup {
				bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			} else {
				bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			}
		}

	}

}

func (bot *CQBot) ReportWeiboHot(groups []int64, isGroup bool) {
	var resp string
	if hotList, err := top_list.LoadWeiboHot(); err != nil {
		log.Error().Msgf("get hot list error:%s", err.Error())
		resp = fmt.Sprintf("爬取微博热搜失败：%s", err.Error())

	} else {
		var hotContent strings.Builder
		hotContent.WriteString(fmt.Sprintf("%s 微博实时热搜\n", time.Now().Format("2006-01-02 15:04")))
		for _, hot := range hotList {
			hotContent.WriteString(fmt.Sprintf("%d\t%s\n", hot.Rank, hot.Title))
		}
		resp = hotContent.String()
	}
	for _, _group := range groups {
		if isGroup {
			bot.SendGroupMessage(_group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
		} else {
			bot.SendPrivateMessage(_group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
		}
	}
}

func (bot *CQBot) Report36kr(groups []int64, isGroup bool) {

	var resp string
	if hotList, err := top_list.Load36krHot(); err != nil {
		log.Error().Msgf("get hot list error:%s", err.Error())
		resp = fmt.Sprintf("爬取36氪热榜失败：%s", err.Error())
	} else {
		var hotContent strings.Builder
		hotContent.WriteString(fmt.Sprintf("%s 36氪24H热榜\n", time.Now().Format("2006-01-02 15:04")))
		for _i, _hot := range hotList {
			if _i > 10 {
				break
			}
			hotContent.WriteString(fmt.Sprintf("%d	%s\n%s\n\n", _hot.Rank, _hot.Title, _hot.Url))
		}
		resp = hotContent.String()
	}
	for _, _group := range groups {
		if isGroup {
			bot.SendGroupMessage(_group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
		} else {
			bot.SendPrivateMessage(_group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
		}
	}

}

func (bot *CQBot) ReportWallStreetNews(groups []int64, isGroup bool) bool {
	hasNews := true
	if hotList, err := top_list.LoadWallStreetNews(); err != nil {
		log.Error().Msgf("爬取华尔街见闻最新资讯失败：%s", err.Error())
	} else {
		var wait sync.WaitGroup
		for _, _group := range groups {
			wait.Add(1)
			go func(group int64) {
				defer wait.Done()
				readyData := make([]top_list.WallStreetNews, 0, 10)
				for _, _data := range hotList {
					if !bot.state.wallstreetSentNews.checkSent(group, _data.Title) {
						readyData = append(readyData, _data)
					}
				}

				if len(readyData) == 0 {
					log.Warn().Msgf("华尔街见闻：没有最新资讯，爬取资讯数量:%d", len(hotList))
					hasNews = false
				} else {
					//倒序输出，因为最新资讯在第一个
					for i := len(readyData) - 1; i >= 0; i-- {
						bot.state.wallstreetSentNews.add(group, readyData[i].Title)
						content := fmt.Sprintf("%s\n\n摘要：%s\n\n链接：%s", readyData[i].Title, readyData[i].Content, readyData[i].Url)
						if isGroup {
							bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
						} else {
							bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
						}
						time.Sleep(1 * time.Second)
					}
				}
			}(_group)
		}
		wait.Wait()
		bot.state.wallstreetSentNews.SaveCache()

	}
	return hasNews
}

func (bot *CQBot) ReportZhihuHot(group int64, isGroup bool) {
	if hotList, err := top_list.LoadZhihuHot(); err != nil {
		log.Error().Msgf("拉取知乎热榜失败：%s", err.Error())
	} else {
		for i := 0; i < len(hotList); i++ {
			content := fmt.Sprintf("知乎热榜#%d %s\n\n摘要：%s\n\n链接：%s", hotList[i].Rank, hotList[i].Title, hotList[i].Excerpt, hotList[i].Url)
			if isGroup {
				bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			} else {
				bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			}
			time.Sleep(1 * time.Second)
		}
	}
}
